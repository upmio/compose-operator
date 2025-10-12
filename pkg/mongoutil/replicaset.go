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
	composev1alpha1 "github.com/upmio/compose-operator/api/v1alpha1"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// ReplicaSetInfoStatus describes the various status of replica set
type ReplicaSetInfoStatus string

const (
	// ReplicaSetInfoConsistent all the nodes have the expected role and configuration
	ReplicaSetInfoConsistent ReplicaSetInfoStatus = "Consistent"
	// ReplicaSetInfoInconsistent some nodes have unexpected configuration
	ReplicaSetInfoInconsistent ReplicaSetInfoStatus = "Inconsistent"
	// ReplicaSetInfoUnset the replica set is not configured
	ReplicaSetInfoUnset ReplicaSetInfoStatus = "Unset"
	// ReplicaSetInfoUnavailable no nodes are reachable
	ReplicaSetInfoUnavailable ReplicaSetInfoStatus = "Unavailable"
	// ReplicaSetInfoPartial status of the replicaset info: data is not complete (some nodes didn't respond) but group is avaiable
	ReplicaSetInfoPartial = "Partial"
)

// MongoDB node role types
const (
	MongoPrimaryRole   = "PRIMARY"
	MongoSecondaryRole = "SECONDARY"
	MongoArbiterRole   = "ARBITER"
	MongoUnknownRole   = "UNKNOWN"
)

// MongoDB node state types
const (
	MongoPrimaryState   = "PRIMARY"
	MongoSecondaryState = "SECONDARY"
	MongoStartupState   = "STARTUP"
	MongoStartup2State  = "STARTUP2"

	MongoRecoveringState = "RECOVERING"
	MongoArbiterState    = "ARBITER"
	MongoDownState       = "DOWN"
	MongoUnknownState    = "UNKNOWN"
	MongoRemovedState    = "REMOVED"
)

// ReplicaSetNode represents a MongoDB replica set member
type ReplicaSetNode struct {
	Host        string
	Port        int
	ID          int
	State       string
	Role        string
	Health      int
	Priority    float64
	Hidden      bool
	ArbiterOnly bool
	Votes       int
}

// GetRole returns the MongoDBReplicasetRole based on the node's state
func (n *ReplicaSetNode) GetRole() composev1alpha1.MongoDBReplicaSetRole {
	switch n.State {
	case MongoPrimaryState:
		return composev1alpha1.MongoDBReplicaSetNodeRolePrimary
	case MongoSecondaryState:
		return composev1alpha1.MongoDBReplicaSetNodeRoleSecondary
	case MongoArbiterState:
		return composev1alpha1.MongoDBReplicaSetNodeRoleArbiter
	default:
		return composev1alpha1.MongoDBReplicaSetNodeRoleNone
	}
}

// NewDefaultReplicaSetNode creates a new ReplicaSetNode with default values
func NewDefaultReplicaSetNode() *ReplicaSetNode {
	return &ReplicaSetNode{
		State:       MongoUnknownState,
		Role:        MongoUnknownRole,
		Health:      0,
		Priority:    1.0,
		Hidden:      false,
		ArbiterOnly: false,
		Votes:       1,
	}
}

// ReplicaSetInfos contains information about all nodes in the replica set
type ReplicaSetInfos struct {
	Infos  map[string]*ReplicaSetNode
	Status ReplicaSetInfoStatus
}

// NewReplicaSetInfos creates a new ReplicaSetInfos instance
func NewReplicaSetInfos() *ReplicaSetInfos {
	return &ReplicaSetInfos{
		Infos:  make(map[string]*ReplicaSetNode),
		Status: ReplicaSetInfoUnset,
	}
}

// ElectPrimary selects a suitable node to become the primary
// In Kubernetes, we rely on automatic mechanisms rather than manual priority configuration
func (i *ReplicaSetInfos) ElectPrimary() string {
	// Simply return the first healthy, non-arbiter node
	// Let MongoDB's native election mechanism handle the actual primary selection
	for addr, node := range i.Infos {
		// Skip arbiter nodes - they can't be primary
		if node.ArbiterOnly {
			continue
		}

		// Skip unhealthy nodes
		if node.Health != 1 {
			continue
		}

		if node.State == MongoPrimaryState {
			return addr
		}
	}

	return ""
}

// ReplicaSetConfig represents the MongoDB replica set configuration
type ReplicaSetConfig struct {
	Id       string                    `bson:"_id"`
	Version  int                       `bson:"version"`
	Members  []ReplicaSetConfigMember  `bson:"members"`
	Settings *ReplicaSetConfigSettings `bson:"settings,omitempty"`
}

// ReplicaSetConfigMember represents a member in the replica set configuration
type ReplicaSetConfigMember struct {
	ID                 int     `bson:"_id"`
	Host               string  `bson:"host"`
	Priority           float64 `bson:"priority"`
	Hidden             bool    `bson:"hidden,omitempty"`
	ArbiterOnly        bool    `bson:"arbiterOnly,omitempty"`
	Votes              int     `bson:"votes,omitempty"`
	SecondaryDelaySecs int     `bson:"secondaryDelaySecs,omitempty"`
	BuildIndexes       bool    `bson:"buildIndexes,omitempty"`
}

// ReplicaSetConfigSettings represents the settings section of replica set configuration
type ReplicaSetConfigSettings struct {
	ChainingAllowed         bool                      `bson:"chainingAllowed,omitempty"`
	HeartbeatIntervalMillis int                       `bson:"heartbeatIntervalMillis,omitempty"`
	HeartbeatTimeoutSecs    int                       `bson:"heartbeatTimeoutSecs,omitempty"`
	ElectionTimeoutMillis   int                       `bson:"electionTimeoutMillis,omitempty"`
	CatchUpTimeoutMillis    int                       `bson:"catchUpTimeoutMillis,omitempty"`
	GetLastErrorModes       map[string]map[string]int `bson:"getLastErrorModes,omitempty"`
	GetLastErrorDefaults    map[string]interface{}    `bson:"getLastErrorDefaults,omitempty"`
	ReplicaSetId            primitive.ObjectID        `bson:"replicaSetId,omitempty"`
}

// ReplicaSetStatus represents the result of rs.status() command
type ReplicaSetStatus struct {
	Set     string                   `bson:"set"`
	Date    interface{}              `bson:"date"`
	MyState int                      `bson:"myState"`
	Term    int                      `bson:"term,omitempty"`
	Members []ReplicaSetStatusMember `bson:"members"`
}

// ReplicaSetStatusMember represents a member in the replica set status
type ReplicaSetStatusMember struct {
	ID                int         `bson:"_id"`
	Name              string      `bson:"name"`
	Health            int         `bson:"health"`
	State             int         `bson:"state"`
	StateStr          string      `bson:"stateStr"`
	Uptime            int         `bson:"uptime,omitempty"`
	OptimeDate        interface{} `bson:"optimeDate,omitempty"`
	ConfigVersion     int         `bson:"configVersion,omitempty"`
	LastHeartbeat     interface{} `bson:"lastHeartbeat,omitempty"`
	LastHeartbeatRecv interface{} `bson:"lastHeartbeatRecv,omitempty"`
	PingMs            int         `bson:"pingMs,omitempty"`
}
