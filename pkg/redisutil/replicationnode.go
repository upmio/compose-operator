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
	"net"
	"strconv"
	"strings"

	"github.com/go-logr/logr"

	composev1alpha1 "github.com/upmio/compose-operator/api/v1alpha1"
)

type ReplicationNode struct {
	Host                 string
	Port                 string
	Role                 string
	MasterLinkStatus     string
	SourceHost           string
	SourcePort           string
	SourceOffset         int64
	ReplicaOffset        int64
	MasterSyncInProgress bool
}

// ReplicationNodes represent a ReplicationNode slice
type ReplicationNodes []*ReplicationNode

// NewDefaultReplicationNode builds and returns new defaultNode instance
func NewDefaultReplicationNode() *ReplicationNode {
	return &ReplicationNode{
		Port:                 DefaultRedisPort,
		MasterLinkStatus:     "",
		SourceOffset:         0,
		ReplicaOffset:        0,
		MasterSyncInProgress: false,
	}
}

// DecodeNode decode from the cmd output the MySQK nodes info. Second argument is the node on which we are connected to request info
func DecodeNode(input *string, addr string, log logr.Logger) *ReplicationNode {
	node := NewDefaultReplicationNode()

	//remove trailing port for cluster internal protocol
	if host, port, err := net.SplitHostPort(addr); err == nil {
		node.Host = host
		node.Port = port
	} else {
		log.Error(err, fmt.Sprintf("failed to parse host:port from node address '%s'", addr))
	}

	lines := strings.Split(*input, "\r\n")

	scanMap := make(map[string]string, 0)
	for _, line := range lines {
		values := strings.Split(line, ":")
		if len(values) != 2 {
			continue
		}

		key, value := values[0], strings.TrimSpace(values[1])
		scanMap[key] = value
	}

	switch scanMap["role"] {
	case "master", "source":
		node.Role = RedisSourceRole
	case "slave", "replica":
		node.Role = RedisReplicaRole
		node.SourceHost = scanMap["master_host"]
		node.SourcePort = scanMap["master_port"]
	}

	if value, ok := scanMap["master_link_status"]; ok {
		node.MasterLinkStatus = value
	}

	if value, ok := scanMap["master_repl_offset"]; ok {
		if offset, err := strconv.ParseInt(value, 10, 64); err == nil {
			node.SourceOffset = offset
		} else {
			log.Error(err, fmt.Sprintf("failed to parse master_repl_offset from node address '%s'", addr))
		}
	}

	if value, ok := scanMap["slave_repl_offset"]; ok {
		if offset, err := strconv.ParseInt(value, 10, 64); err == nil {
			node.ReplicaOffset = offset
		} else {
			log.Error(err, fmt.Sprintf("failed to parse slave_repl_offset from node address '%s'", addr))
		}
	}

	if value, ok := scanMap["master_sync_in_progress"]; ok {
		if flag, err := strconv.Atoi(value); err == nil {
			node.MasterSyncInProgress = flag == 1
		} else {
			log.Error(err, fmt.Sprintf("failed to parse master_sync_in_progress from node address '%s'", addr))
		}
	}

	return node
}

// HostPort returns join Host Port string
func (n *ReplicationNode) HostPort() string {
	return net.JoinHostPort(n.Host, n.Port)
}

// GetRole return the Mysql Replication ReplicationNode GetRole
func (n *ReplicationNode) GetRole() composev1alpha1.RedisReplicationRole {
	switch n.Role {
	case RedisSourceRole:
		return composev1alpha1.RedisReplicationNodeRoleSource
	case RedisReplicaRole:
		return composev1alpha1.RedisReplicationNodeRoleReplica
	}

	return composev1alpha1.RedisReplicationNodeRoleNone
}

func (n *ReplicationNode) GetSourcePort() int {
	port, _ := strconv.Atoi(n.SourcePort)
	return port
}

// OffsetLag returns the replication offset lag when the node acts as a replica.
func (n *ReplicationNode) OffsetLag() (int64, error) {
	if n.SourceOffset < 0 || n.ReplicaOffset < 0 {
		return 0, fmt.Errorf("replication offsets are not available")
	}

	lag := n.SourceOffset - n.ReplicaOffset
	if lag < 0 {
		return 0, nil
	}

	return lag, nil
}
