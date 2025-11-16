/*
Copyright 2025 The Compose Operator Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.

SPDX-License-Identifier: Apache-2.0
*/

package mongoutil

import (
	"context"
	"fmt"
	"net"
	"strconv"
	"strings"

	"github.com/go-logr/logr"
)

// IReplicaSetAdmin MongoDB replica set admin interface
type IReplicaSetAdmin interface {
	// Connections returns the connection map of all clients
	Connections() IAdminConnections

	// Close the admin connections
	Close()

	// GetReplicaSetInfos return the ReplicaSetInfos for all nodes
	GetReplicaSetInfos(ctx context.Context, memberCount int) *ReplicaSetInfos

	// InitiateReplicaSet initializes a new replica set
	InitiateReplicaSet(ctx context.Context, addr, replicaSetName string, members []ReplicaSetConfigMember) error

	// AddMember adds a new member to the replica set
	AddMember(ctx context.Context, primaryAddr, memberHost string, memberPort int) error

	// RemoveMember removes a member from the replica set
	RemoveMember(ctx context.Context, primaryAddr, memberHost string, memberPort int) error

	// StepDown forces the primary to step down
	StepDown(ctx context.Context, primaryAddr string, stepDownSecs int) error
}

// ReplicaSetAdmin wraps MongoDB replica set admin logic
type ReplicaSetAdmin struct {
	cnx IAdminConnections
	log logr.Logger
}

// NewReplicaSetAdmin returns new IReplicaSetAdmin instance
func NewReplicaSetAdmin(addrs []string, options *AdminOptions, log logr.Logger) IReplicaSetAdmin {
	a := &ReplicaSetAdmin{
		log: log.WithName("mongo_util"),
	}

	// perform initial connections
	a.cnx = NewAdminConnections(addrs, options, log)

	return a
}

// Connections returns the connection map of all clients
func (a *ReplicaSetAdmin) Connections() IAdminConnections {
	return a.cnx
}

// Close used to close all possible resources instance by the ReplicaSetAdmin
func (a *ReplicaSetAdmin) Close() {
	a.Connections().Reset()
}

// GetReplicaSetInfos return the ReplicaSetInfos for all nodes
func (a *ReplicaSetAdmin) GetReplicaSetInfos(ctx context.Context, memberCount int) *ReplicaSetInfos {
	infos := NewReplicaSetInfos()

	var (
		primaryCount int
		primaryAddr  string
	)

	for addr := range a.Connections().GetAll() {
		node, err := a.getReplicaSetNode(ctx, addr)
		if err != nil {
			a.log.Error(err, "failed to gather MongoDB replica set information", "address", addr)
			continue
		}

		// Check for multiple primaries
		if node.State == MongoPrimaryState {
			primaryCount++
			primaryAddr = addr
		}

		infos.Infos[addr] = node
	}

	switch {
	case primaryCount == 1:
		hasRecovering := false

		// Check for missing members from the perspective of the primary
		for addr, node := range infos.Infos {
			if node.State == MongoPrimaryState {
				continue
			}

			// Verify the member is in the replica set configuration
			if err := a.verifyMemberInConfig(ctx, primaryAddr, node.Host, node.Port); err != nil {
				a.log.Info(fmt.Sprintf("failed to verify member in config, %v", err), "address", addr)
				infos.Infos[addr].State = MongoRemovedState
			}

			if node.State == MongoRecoveringState || node.State == MongoStartup2State || node.State == MongoRemovedState {
				hasRecovering = true
			}
		}

		if hasRecovering || len(infos.Infos) != memberCount {
			infos.Status = ReplicaSetInfoPartial
			return infos
		} else {
			infos.Status = ReplicaSetInfoConsistent
		}
	case primaryCount > 1:
		infos.Status = ReplicaSetInfoInconsistent
		return infos
	case primaryCount == 0:
		// No primary found, check if cluster is uninitialized or unavailable
		allUnknown := true

		for _, node := range infos.Infos {
			if node.State != MongoUnknownState {
				allUnknown = false
			}
		}

		if allUnknown && len(infos.Infos) == memberCount {
			infos.Status = ReplicaSetInfoUnset
			return infos
		} else {
			infos.Status = ReplicaSetInfoUnavailable
			return infos
		}

	}

	return infos
}

// getReplicaSetNode gets information about a single replica set node
func (a *ReplicaSetAdmin) getReplicaSetNode(ctx context.Context, addr string) (*ReplicaSetNode, error) {
	client, err := a.cnx.Get(addr)
	if err != nil {
		return nil, err
	}

	host, port, err := net.SplitHostPort(addr)
	if err != nil {
		return nil, fmt.Errorf("unable to split host and port from address: %s", addr)
	}

	portInt, err := strconv.Atoi(port)
	if err != nil {
		return nil, fmt.Errorf("failed to convert port from string to int: %s", port)
	}

	node := NewDefaultReplicaSetNode()
	node.Host = host
	node.Port = portInt

	// Get isMaster information
	isMasterResult, err := client.IsMaster(ctx)
	if err != nil {
		return node, fmt.Errorf("failed to get isMaster info from %s: %v", addr, err)
	}

	// Determine state and role based on isMaster result
	if isMasterResult.IsMaster {
		node.Role = MongoPrimaryRole
	} else if isMasterResult.IsSecondary {
		node.Role = MongoSecondaryRole
	}

	// Try to get more detailed status if possible
	status, err := client.GetReplicaSetStatus(ctx)
	if err != nil {
		// If we can't get status, that's okay, we have basic info from isMaster
		if strings.Contains(err.Error(), "NotYetInitialized") {
			a.log.V(4).Info("not yet initialized", "address", addr, "error", err)
		} else {
			return nil, err
		}
	} else {
		// Find this node in the status
		for _, member := range status.Members {
			if member.Name == addr || member.Name == fmt.Sprintf("%s:%d", host, portInt) {
				node.Health = member.Health
				node.State = member.StateStr
				node.ID = member.ID
				break
			}
		}
	}

	return node, nil
}

// verifyMemberInConfig verifies that a member is present in the replica set configuration
func (a *ReplicaSetAdmin) verifyMemberInConfig(ctx context.Context, primaryAddr, memberHost string, memberPort int) error {
	client, err := a.cnx.Get(primaryAddr)
	if err != nil {
		return err
	}

	config, err := client.GetReplicaSetConfig(ctx)
	if err != nil {
		return fmt.Errorf("failed to get replica set config: %v", err)
	}

	memberAddr := fmt.Sprintf("%s:%d", memberHost, memberPort)
	for _, member := range config.Members {
		if member.Host == memberAddr {
			return nil // Member found in config
		}
	}

	return fmt.Errorf("member %s not found in replica set configuration", memberAddr)
}

// InitiateReplicaSet initializes a new replica set
func (a *ReplicaSetAdmin) InitiateReplicaSet(ctx context.Context, addr, replicaSetName string, members []ReplicaSetConfigMember) error {
	client, err := a.cnx.Get(addr)
	if err != nil {
		return err
	}

	config := &ReplicaSetConfig{
		Id:      replicaSetName,
		Version: 1,
		Members: members,
	}

	return client.InitiateReplicaSet(ctx, config)
}

// AddMember adds a new member to the replica set
func (a *ReplicaSetAdmin) AddMember(ctx context.Context, primaryAddr, memberHost string, memberPort int) error {
	client, err := a.cnx.Get(primaryAddr)
	if err != nil {
		return err
	}

	return client.AddMemberToReplicaSet(ctx, memberHost, memberPort)
}

// RemoveMember removes a member from the replica set
func (a *ReplicaSetAdmin) RemoveMember(ctx context.Context, primaryAddr, memberHost string, memberPort int) error {
	client, err := a.cnx.Get(primaryAddr)
	if err != nil {
		return err
	}

	return client.RemoveMemberFromReplicaSet(ctx, memberHost, memberPort)
}

// StepDown forces the primary to step down
func (a *ReplicaSetAdmin) StepDown(ctx context.Context, primaryAddr string, stepDownSecs int) error {
	client, err := a.cnx.Get(primaryAddr)
	if err != nil {
		return err
	}

	return client.StepDown(ctx, stepDownSecs)
}
