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
	"strings"

	composev1alpha1 "github.com/upmio/compose-operator/api/v1alpha1"
)

const (
	// GroupInfoUnset status of the group info: no data set
	GroupInfoUnset = "Unset"

	// GroupInfoInconsistent status of the group info: nodesinfos is not consistent between nodes
	GroupInfoInconsistent = "Inconsistent"

	// GroupInfoConsistent status of the group info: nodeinfos is complete and consistent between nodes
	GroupInfoConsistent = "Consistent"

	// GroupInfoPartial status of the group info: data is not complete (some nodes didn't respond) but group is avaiable
	GroupInfoPartial = "Partial"

	// GroupInfoUnavailable status of the group info: data is not complete (some nodes didn't respond) but group is unavailable
	GroupInfoUnavailable = "Unavailable"
)

const (
	// MysqlPrimaryRole mysql role primary
	MysqlPrimaryRole = "PRIMARY"
	// MysqlSecondaryRole mysql role secondary
	MysqlSecondaryRole = "SECONDARY"
	// MysqlUnknownRole mysql role unknown
	MysqlUnknownRole = "UNKNOWN"
)

const (
	MysqlOnlineState      = "ONLINE"
	MysqlOfflineState     = "OFFLINE"
	MysqlUnreachableState = "UNREACHABLE"
	MysqlRecoveringState  = "RECOVERING"
	MysqlErrorState       = "ERROR"
	MysqlMissingState     = "MISSING"
)

// GroupInfos represents the node infos for all nodes of the cluster
type GroupInfos struct {
	Infos  map[string]*GroupNode
	Status string
}

type GroupNodeInfo struct {
	ID    string
	Host  string
	Port  int
	Role  string
	State string
}

type GroupNode struct {
	*GroupNodeInfo
	ReadOnly      bool
	SuperReadOnly bool
	GtidExecuted  string
}

// NewGroupInfos returns an instance of GroupInfos
func NewGroupInfos() *GroupInfos {
	return &GroupInfos{
		Infos:  make(map[string]*GroupNode),
		Status: GroupInfoUnset,
	}
}

// NewDefaultGroupNodeInfo builds and returns new default GroupNodeInfo instance
func NewDefaultGroupNodeInfo() *GroupNodeInfo {
	return &GroupNodeInfo{
		Port:  DefaultMysqlPort,
		Role:  MysqlUnknownRole,
		State: MysqlOfflineState,
	}
}

// NewDefaultGroupNode builds and returns new defaultNode instance
func NewDefaultGroupNode() *GroupNode {
	return &GroupNode{
		GroupNodeInfo: NewDefaultGroupNodeInfo(),
	}
}

// compareGTID compares two GTID strings
func compareGTID(gtid1, gtid2 string) bool {
	gtid1Parts := strings.Split(gtid1, "-")
	gtid2Parts := strings.Split(gtid2, "-")

	for i := 0; i < len(gtid1Parts) && i < len(gtid2Parts); i++ {
		if gtid1Parts[i] > gtid2Parts[i] {
			return true
		} else if gtid1Parts[i] < gtid2Parts[i] {
			return false
		}
	}

	// If the preceding parts are identical, the longer GTID is considered greater
	if len(gtid1Parts) > len(gtid2Parts) {
		return true
	} else if len(gtid1Parts) < len(gtid2Parts) {
		return false
	}

	return false
}

// GetRole return the Mysql Replication Node GetRole
func (n *GroupNode) GetRole() composev1alpha1.MysqlGroupReplicationRole {
	switch n.Role {
	case MysqlPrimaryRole:
		return composev1alpha1.MysqlGroupReplicationNodeRolePrimary
	case MysqlSecondaryRole:
		return composev1alpha1.MysqlGroupReplicationNodeRoleSecondary
	}

	return composev1alpha1.MysqlGroupReplicationNodeRoleNone
}

func (g *GroupInfos) ElectPrimary() string {
	var primaryAddr string
	var maxGTID string

	for addr, node := range g.Infos {
		// Skip nodes without GTID
		if node.GtidExecuted == "" {
			continue
		}

		// Update the candidate node (primaryAddr) with the largest GTID among all nodes
		if maxGTID == "" || compareGTID(node.GtidExecuted, maxGTID) {
			maxGTID = node.GtidExecuted
			primaryAddr = addr
		}

	}
	return primaryAddr
}
