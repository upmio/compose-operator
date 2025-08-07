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

package mysqlutil

import (
	"context"
	"fmt"
	"github.com/go-logr/logr"
	"net"
	"strconv"
)

// IGroupAdmin mysql group replication admin interface
type IGroupAdmin interface {
	// Connections returns the connection map of all clients
	Connections() IAdminConnections

	// Close the admin connections
	Close()

	// GetGroupInfos return the GroupInfos for all nodes
	GetGroupInfos(ctx context.Context, memberCount int) *GroupInfos

	SetupGroup(ctx context.Context, addr, replicationUser, replicationPassword string) error

	JoinGroup(ctx context.Context, addr, replicationUser, replicationPassword string) error

	EnsureGroupSeeds(ctx context.Context, addr, seeds, allowList string) error

	StopGroupReplication(ctx context.Context, addr string) error
}

// GroupAdmin wraps redis cluster admin logic
type GroupAdmin struct {
	cnx IAdminConnections
	log logr.Logger
}

// NewGroupAdmin returns new AdminInterface instance
// at the same time it connects to all Redis ReplicationNodes thanks to the address list
func NewGroupAdmin(addrs []string, options *AdminOptions, log logr.Logger) IGroupAdmin {
	a := &GroupAdmin{
		log: log.WithName("mysql_util"),
	}

	// perform initial connections
	a.cnx = NewAdminConnections(addrs, options, log)

	return a
}

// Connections returns the connection map of all clients
func (a *GroupAdmin) Connections() IAdminConnections {
	return a.cnx
}

// Close used to close all possible resources instance by the ReplicationAdmin
func (a *GroupAdmin) Close() {
	a.Connections().Reset()
}

// GetGroupInfos return the ReplicationNodes infos for all nodes
func (a *GroupAdmin) GetGroupInfos(ctx context.Context, memberCount int) *GroupInfos {
	infos := NewGroupInfos()

	var (
		primaryCount int
		primaryAddr  string
	)

	for addr := range a.Connections().GetAll() {

		node, err := a.getGroupNode(ctx, addr)

		if err != nil {
			a.log.Error(err, "failed to gather mysql group replication information")
			continue
		}

		// issue if there is two or more primary node, how to handel the group info
		if node.Role == MysqlPrimaryRole && node.State == MysqlOnlineState {
			primaryCount++
			primaryAddr = addr
		}

		infos.Infos[addr] = node
	}

	switch {
	case primaryCount == 1:
		for addr, node := range infos.Infos {

			if node.Role == MysqlPrimaryRole {
				continue
			}

			nodeInfo, err := a.getGroupNodeInfos(ctx, primaryAddr, node.Host, node.Port)
			if err != nil {
				a.log.Error(err, "failed to gather mysql group replication information")
				continue
			}

			if nodeInfo == nil {
				infos.Infos[addr].State = MysqlMissingState
				infos.Infos[addr].Role = MysqlUnknownRole
			}
		}
	case primaryCount > 1:
		infos.Status = GroupInfoInconsistent
		return infos
	}

	onlineNodeCount := 0
	offlineNodeCount := 0
	recoveringNodeCount := 0
	for _, node := range infos.Infos {
		switch node.State {
		case MysqlOnlineState:
			onlineNodeCount++
		case MysqlRecoveringState:
			recoveringNodeCount++
		case MysqlOfflineState:
			offlineNodeCount++
		}
	}

	quorum := (memberCount / 2) + 1

	switch {
	case onlineNodeCount == 0 && recoveringNodeCount == 0 && offlineNodeCount == memberCount:
		infos.Status = GroupInfoUnset
	case onlineNodeCount == memberCount && recoveringNodeCount == 0 && offlineNodeCount == 0:
		infos.Status = GroupInfoConsistent
	case (onlineNodeCount+recoveringNodeCount) >= 0 && (onlineNodeCount+recoveringNodeCount) < quorum:
		infos.Status = GroupInfoUnavailable
	case (onlineNodeCount+recoveringNodeCount) <= memberCount && (onlineNodeCount+recoveringNodeCount) >= quorum:
		infos.Status = GroupInfoPartial
	}

	return infos
}

func (a *GroupAdmin) getGroupNodeInfos(ctx context.Context, addr, host string, port int) (*GroupNodeInfo, error) {
	c, err := a.cnx.Get(addr)
	if err != nil {
		return nil, err
	}

	nodeInfo := NewDefaultGroupNodeInfo()
	nodeInfo.Host = host
	nodeInfo.Port = port

	if rows, err := c.Query(ctx, getGroupReplicationMembersSQL, host, port); err != nil {
		return nil, fmt.Errorf("failed to query replication group member (host=%s, port=%d) on [%s]: %v", host, port, addr, err)
	} else {
		defer func() { _ = rows.Close() }()

		if !rows.Next() {
			a.log.Info(fmt.Sprintf("no row (host=%s, port=%d) found in performance_schema.replication_group_members on [%s]", host, port, addr))

			return nil, nil
		} else {

			if err = rows.Scan(&nodeInfo.ID, &nodeInfo.Role, &nodeInfo.State); err != nil {
				return nil, fmt.Errorf("failed to scan group replication members on [%s]: %v", addr, err)
			}

		}
	}

	return nodeInfo, nil
}

func (a *GroupAdmin) getGroupNode(ctx context.Context, addr string) (*GroupNode, error) {

	c, err := a.cnx.Get(addr)
	if err != nil {
		return nil, err
	}

	host, port, err := net.SplitHostPort(addr)
	if err != nil {
		return nil, fmt.Errorf("unable to split host and port from address: %s", addr)
	}

	portInt, err := strconv.Atoi(port)
	if err != nil {
		return nil, fmt.Errorf("failed to convert port from string to int")
	}

	node := NewDefaultGroupNode()
	node.Host = host
	node.Port = portInt

	nodeInfo, err := a.getGroupNodeInfos(ctx, addr, host, portInt)
	if err != nil {
		return node, err
	}

	if nodeInfo != nil {
		node.GroupNodeInfo = nodeInfo
	}

	var flag string

	if err = c.QueryRow(ctx, &flag, getReadOnlySQL); err != nil {
		return nil, fmt.Errorf("failed to get 'read_only' variables on [%s]: %v", addr, err)
	} else {
		switch flag {
		case "1":
			node.ReadOnly = true
		case "0":
			node.ReadOnly = false
		}
	}

	if err = c.QueryRow(ctx, &flag, getSuperReadOnlySQL); err != nil {
		return nil, fmt.Errorf("failed to get 'super_read_only' variables on [%s]: %v", addr, err)
	} else {
		switch flag {
		case "1":
			node.SuperReadOnly = true
		case "0":
			node.SuperReadOnly = false
		}
	}

	if err = c.QueryRow(ctx, &node.GtidExecuted, getGtidExecutedSQL); err != nil {
		return node, fmt.Errorf("failed to execute 'SELECT @@gtid_executed' on [%s]: %v ", addr, err)
	}

	return node, nil
}

func (a *GroupAdmin) EnsureGroupSeeds(ctx context.Context, addr, seeds, allowList string) error {

	c, err := a.cnx.Get(addr)
	if err != nil {
		return err
	}

	var currentSeeds string
	if err = c.QueryRow(ctx, &currentSeeds, getGroupReplicationSeeds); err != nil {
		return fmt.Errorf("failed to execute 'SELECT @@group_replication_group_seeds' on '%s', %v", addr, err)
	}

	if currentSeeds != seeds {
		if _, err = c.Exec(ctx, setGroupReplicationSeeds, seeds); err != nil {
			return fmt.Errorf("failed to set 'group_replication_group_seeds=%s' on '%s', %v", seeds, addr, err)
		}
	}

	var currentAllowList string
	if err = c.QueryRow(ctx, &currentAllowList, getGroupReplicationAllowList); err != nil {
		return fmt.Errorf("failed to execute 'SELECT @@group_replication_ip_allowlist' on '%s', %v", addr, err)
	}

	if currentAllowList != allowList {
		if _, err = c.Exec(ctx, setGroupReplicationAllowList, allowList); err != nil {
			return fmt.Errorf("failed to set 'group_replication_ip_allowlist=%s' on '%s', %v", allowList, addr, err)
		}
	}

	return nil
}

func (a *GroupAdmin) SetupGroup(ctx context.Context, addr, replicationUser, replicationPassword string) error {
	c, err := a.cnx.Get(addr)
	if err != nil {
		return err
	}

	if err := a.StopGroupReplication(ctx, addr); err != nil {
		return err
	}

	if _, err := c.Exec(ctx, setGroupReplicationSQL, replicationUser, replicationPassword); err != nil {
		return fmt.Errorf("failed to execute 'CHANGE REPLICATION SOURCE TO...' on [%s]: %v", addr, err)
	}

	if _, err := c.Exec(ctx, setGroupReplicationBootStrapON); err != nil {
		return fmt.Errorf("failed to set 'group_replication_bootstrap_group=ON' on [%s]: %v", addr, err)
	}

	if err := a.StartGroupReplication(ctx, addr); err != nil {
		return err
	}

	if _, err := c.Exec(ctx, setGroupReplicationBootStrapOFF); err != nil {
		return fmt.Errorf("failed to set 'group_replication_bootstrap_group=OFF' on [%s]: %v", addr, err)
	}

	return nil
}

func (a *GroupAdmin) JoinGroup(ctx context.Context, addr, replicationUser, replicationPassword string) error {
	c, err := a.cnx.Get(addr)
	if err != nil {
		return err
	}

	//if _, err := c.Exec(ctx, stopGroupReplicationSQL); err != nil {
	//	return err
	//}

	if _, err := c.Exec(ctx, setGroupReplicationSQL, replicationUser, replicationPassword); err != nil {
		return fmt.Errorf("failed to execute 'CHANGE REPLICATION SOURCE TO...' on [%s]: %v", addr, err)
	}

	if err := a.StartGroupReplication(ctx, addr); err != nil {
		return err
	}

	return nil
}

// StartGroupReplication stop group replication
func (a *GroupAdmin) StartGroupReplication(ctx context.Context, addr string) error {
	c, err := a.cnx.Get(addr)
	if err != nil {
		return err
	}

	if _, err := c.Exec(ctx, startGroupReplicationSQL); err != nil {
		return fmt.Errorf("failed to execute 'START GROUP REPLICA' on [%s]: %v", addr, err)
	}

	return nil
}

// StopGroupReplication stop group replication
func (a *GroupAdmin) StopGroupReplication(ctx context.Context, addr string) error {
	c, err := a.cnx.Get(addr)
	if err != nil {
		return err
	}

	if _, err := c.Exec(ctx, stopGroupReplicationSQL); err != nil {
		return fmt.Errorf("failed to execute 'STOP GROUP REPLICA' on [%s]: %v", addr, err)
	}

	return nil
}
