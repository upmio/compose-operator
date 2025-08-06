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

package postgresutil

import (
	"net"
	"strconv"

	composev1alpha1 "github.com/upmio/compose-operator/api/v1alpha1"
)

const (
	// DefaultPostgresPort define the default Postgres Port
	DefaultPostgresPort = 5432
	// PostgresPrimaryRole Postgres role primary
	PostgresPrimaryRole = "Primary"
	// PostgresStandbyRole Postgres role standby
	PostgresStandbyRole = "Standby"
)

type Node struct {
	Host            string
	Port            int
	Role            string
	ReplicationStat []string
	DataDir         string
	ReplicationMode string
	WalDiff         int
}

// Nodes represent a Node slice
type Nodes []*Node

// NewDefaultNode builds and returns new defaultNode instance
func NewDefaultNode() *Node {
	return &Node{
		Port:            DefaultPostgresPort,
		ReplicationStat: make([]string, 0),
	}
}

// HostPort returns join Host Port string
func (n *Node) HostPort() string {
	return net.JoinHostPort(n.Host, strconv.Itoa(n.Port))
}

// GetRole return the Postgres Replication Node GetRole
func (n *Node) GetRole() composev1alpha1.PostgresReplicationRole {
	switch n.Role {
	case PostgresPrimaryRole:
		return composev1alpha1.PostgresReplicationRolePrimary
	case PostgresStandbyRole:
		return composev1alpha1.PostgresReplicationRoleStandby
	}

	return composev1alpha1.PostgresReplicationRoleNone
}
