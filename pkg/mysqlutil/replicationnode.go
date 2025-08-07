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
	"database/sql"
	"fmt"
	"net"
	"strconv"

	"github.com/go-logr/logr"

	composev1alpha1 "github.com/upmio/compose-operator/api/v1alpha1"
)

const (
	// DefaultMysqlPort define the default MySQL Port
	DefaultMysqlPort = 3306
	// MysqlSourceRole mysql role source
	MysqlSourceRole = "Source"
	// MysqlReplicaRole mysql role replica
	MysqlReplicaRole = "Replica"
)

type ReplicationNode struct {
	Host                   string
	Port                   int
	Role                   string
	Ready                  bool
	ReadOnly               bool
	SuperReadOnly          bool
	SourceHost             string
	SourcePort             string
	ReplicaIO              string
	ReplicaSQL             string
	ReadSourceLogPos       string
	ExecSourceLogPos       string
	SourceLogFile          string
	SecondsBehindSource    int
	SemiSyncSourceEnabled  bool
	SemiSyncReplicaEnabled bool
}

// ReplicationNodes represent a ReplicationNode slice
type ReplicationNodes []*ReplicationNode

// NewDefaultReplicationNode builds and returns new defaultNode instance
func NewDefaultReplicationNode() *ReplicationNode {
	return &ReplicationNode{
		Port:  DefaultMysqlPort,
		Ready: true,
	}
}

// DecodeNode decode from the cmd output the MySQL nodes info. Second argument is the node on which we are connected to request info
func DecodeNode(result map[string]sql.NullString, addr string, log logr.Logger) *ReplicationNode {
	node := NewDefaultReplicationNode()

	//remove trailing port for cluster internal protocol
	if host, port, err := net.SplitHostPort(addr); err == nil {
		node.Host = host
		node.Port, _ = strconv.Atoi(port)
	} else {
		log.Error(err, fmt.Sprintf("failed to parse host:port from node address '%s'", addr))
	}

	if result == nil {
		node.Role = MysqlSourceRole
	} else {
		node.Role = MysqlReplicaRole
		node.SourceHost = result["Source_Host"].String
		node.SourcePort = result["Source_Port"].String
		node.ReplicaIO = result["Replica_IO_Running"].String
		node.ReplicaSQL = result["Replica_SQL_Running"].String
		node.ReadSourceLogPos = result["Read_Source_Log_Pos"].String
		node.SourceLogFile = result["Source_Log_File"].String
		node.ExecSourceLogPos = result["Exec_Source_Log_Pos"].String
		behindSource, _ := strconv.Atoi(result["Seconds_Behind_Source"].String)
		node.SecondsBehindSource = behindSource
	}

	return node
}

// HostPort returns join Host Port string
func (n *ReplicationNode) HostPort() string {
	return net.JoinHostPort(n.Host, strconv.Itoa(n.Port))
}

// GetRole return the Mysql Replication ReplicationNode GetRole
func (n *ReplicationNode) GetRole() composev1alpha1.MysqlReplicationRole {
	switch n.Role {
	case MysqlSourceRole:
		return composev1alpha1.MysqlReplicationNodeRoleSource
	case MysqlReplicaRole:
		return composev1alpha1.MysqlReplicationNodeRoleReplica
	}

	return composev1alpha1.MysqlReplicationNodeRoleNone
}

func (n *ReplicationNode) GetSourcePort() int {
	port, _ := strconv.Atoi(n.SourcePort)
	return port
}

func (n *ReplicationNode) GetReadSourceLogPos() int {
	position, _ := strconv.Atoi(n.ReadSourceLogPos)
	return position
}

func (n *ReplicationNode) GetExecSourceLogPos() int {
	position, _ := strconv.Atoi(n.ExecSourceLogPos)
	return position
}
