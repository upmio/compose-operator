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

package redisutil

import (
	"fmt"
	"time"

	"github.com/go-logr/logr"
)

// IReplicationAdmin redis cluster admin interface
type IReplicationAdmin interface {
	// Connections returns the connection map of all clients
	Connections() IAdminConnections
	// Close the admin connections
	Close()
	// GetReplicationStatus get redis replication information
	GetReplicationStatus() *ReplicationInfo

	// ReplicaOfNoOne redis replicaof no one
	ReplicaOfNoOne(addr string) error

	// ReplicaOfSource redis replicaof source node
	ReplicaOfSource(addr, sourceHost, sourcePort string) error

	// WaitReplicaCatchUp waits until the replica at addr catches up to within the provided offset threshold.
	WaitReplicaCatchUp(addr string, maxOffsetLag int64, timeout time.Duration) error
}

// ReplicationAdmin wraps redis cluster admin logic
type ReplicationAdmin struct {
	cnx IAdminConnections
	log logr.Logger
}

// NewReplicationAdmin returns new AdminInterface instance
// at the same time it connects to all Redis ReplicationNodes thanks to the address list
func NewReplicationAdmin(addrs []string, options *AdminOptions, log logr.Logger) IReplicationAdmin {
	a := &ReplicationAdmin{
		log: log.WithName("redis_util"),
	}

	// perform initial connections
	a.cnx = NewAdminConnections(addrs, options, log)

	return a
}

// Connections returns the connection map of all clients
func (a *ReplicationAdmin) Connections() IAdminConnections {
	return a.cnx
}

// Close used to close all possible resources instanciate by the ReplicationAdmin
func (a *ReplicationAdmin) Close() {
	a.Connections().Reset()
}

// GetReplicationStatus return the ReplicationNodes infos for all nodes
func (a *ReplicationAdmin) GetReplicationStatus() *ReplicationInfo {
	infos := NewReplicationInfo()

	for addr := range a.Connections().GetAll() {

		node, err := a.getInfos(addr)
		if err != nil {
			a.log.Error(err, "failed to gather redis replication information")
			continue
		}
		if node != nil && node.HostPort() == addr {
			infos.Nodes[addr] = node
		} else {
			a.log.Info("bad node info retrieved from", "addr", addr)
		}
	}

	return infos
}

func (a *ReplicationAdmin) getInfos(addr string) (*ReplicationNode, error) {
	c, err := a.cnx.Get(addr)
	if err != nil {
		return nil, err
	}

	resp := c.Cmd("info", "replication")

	input, err := resp.Str()
	if err != nil {
		return nil, fmt.Errorf("failed to execute 'INFO REPLICATION' on [%s]: %v", addr, err)
	}

	node := DecodeNode(&input, addr, a.log)

	return node, nil
}

func (a *ReplicationAdmin) ReplicaOfNoOne(addr string) error {
	c, err := a.cnx.Get(addr)
	if err != nil {
		return err
	}

	resp := c.Cmd("REPLICAOF", "NO", "ONE")
	if err := resp.Err; err != nil {
		return fmt.Errorf("failed to execute 'REPLICAOF NO ONE' on [%s]: %v", addr, err)
	}

	resp = c.Cmd("CONFIG", "REWRITE")
	if err := resp.Err; err != nil {
		return fmt.Errorf("failed to execute 'CONFIG REWRITE' on [%s]: %v", addr, err)
	}

	return nil
}

func (a *ReplicationAdmin) ReplicaOfSource(addr, sourceHost, sourcePort string) error {
	c, err := a.cnx.Get(addr)
	if err != nil {
		return err
	}

	resp := c.Cmd("REPLICAOF", sourceHost, sourcePort)
	if err := resp.Err; err != nil {
		return fmt.Errorf("failed to execute 'REPLICAOF %s %s' on [%s]: %v", sourceHost, sourcePort, addr, err)
	}

	resp = c.Cmd("CONFIG", "REWRITE")
	if err := resp.Err; err != nil {
		return fmt.Errorf("failed to execute 'CONFIG REWRITE' on [%s]: %v", addr, err)
	}

	return nil
}

func (a *ReplicationAdmin) WaitReplicaCatchUp(addr string, maxOffsetLag int64, timeout time.Duration) error {
	deadline := time.Now().Add(timeout)

	for {
		node, err := a.getInfos(addr)
		if err != nil {
			return fmt.Errorf("failed to check replication offset on [%s]: %v", addr, err)
		}
		if node == nil {
			return fmt.Errorf("failed to check replication offset on [%s]: empty node info", addr)
		}

		if node.MasterSyncInProgress {
			if time.Now().After(deadline) {
				return fmt.Errorf("replica [%s] master sync still in progress after %s", addr, timeout)
			}
			time.Sleep(time.Second)
			continue
		}

		lag, err := node.OffsetLag()
		if err != nil {
			return fmt.Errorf("failed to check replication offset on [%s]: %v", addr, err)
		}

		if lag <= maxOffsetLag {
			return nil
		}

		if time.Now().After(deadline) {
			return fmt.Errorf("replica [%s] offset lag '%d' exceeds threshold '%d'", addr, lag, maxOffsetLag)
		}

		time.Sleep(time.Second)
	}
}
