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

package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// PostgresReplicationSecret defines the secret information of PostgresqlReplication
type PostgresReplicationSecret struct {
	// Name is the name of the secret resource which store authentication information for Postgres.
	Name string `json:"name"`
	//  Mysql is the key of the secret, which contains the value used to connect to Postgres.
	// +kubebuilder:default:=postgres
	Postgresql string `json:"postgres"`
	// Replication is the key of the secret, which contains the value used to set up Postgres replication.
	// +kubebuilder:default:=replication
	Replication string `json:"replication"`
}

// PostgresReplicationSpec defines the desired state of PostgresReplication
type PostgresReplicationSpec struct {
	// Mode specifies the postgres replication sync mode.
	// Valid values are:
	// - "rpl_sync": sync;
	// - "rpl_async": async;
	// +optional
	// +kubebuilder:default:=rpl_async
	Mode PostgresReplicationMode `json:"mode"`

	// Secret is the reference to the secret resource containing authentication information, it must be in the same namespace as the PostgresReplication object.
	Secret PostgresReplicationSecret `json:"secret"`

	// AESSecret is the reference to the secret resource containing aes key, it must be in the same namespace as the PostgresReplication Object.
	AESSecret *AESSecret `json:"aesSecret,omitempty"`

	// Primary references the primary Postgres node.
	Primary *CommonNode `json:"primary"`

	// Service references the service providing the Postgres replication endpoint.
	Service *Service `json:"service"`

	// Standby is a list of standby nodes in the Postgres replication topology.
	Standby CommonNodes `json:"standby,omitempty"`
}

// PostgresReplicationNode represents a node in the Postgres replication topology.
type PostgresReplicationNode struct {
	// Host indicates the host of the MySQL node.
	Host string `json:"host"`

	// Port indicates the port of the MySQL node.
	Port int `json:"port"`

	// Role represents the role of the node in the replication topology (e.g., primary, standby).
	Role PostgresReplicationRole `json:"role"`

	// Status indicates the current status of the node (e.g., Healthy, Failed).
	Status NodeStatus `json:"status"`

	// Ready indicates whether the node is ready for reads and writes.
	Ready bool `json:"ready"`

	// WalDiff indicates the standby node pg_wal_lsn_diff(pg_last_wal_replay_lsn(), pg_last_wal_receive_lsn()).
	WalDiff *int `json:"walDiff,omitempty"`
}

// PostgresReplicationTopology defines the PostgresReplication topology
type PostgresReplicationTopology map[string]*PostgresReplicationNode

// PostgresReplicationStatus defines the observed state of PostgresReplication
type PostgresReplicationStatus struct {
	// Topology indicates the current Postgres replication topology.
	Topology PostgresReplicationTopology `json:"topology"`

	// ReadWriteService specify the service name provides read-write access to database.
	ReadWriteService string `json:"readwriteService"`

	// ReadOnlyService specify the service name provides read-only access to database.
	ReadOnlyService string `json:"readonlyService,omitempty"`

	// Ready indicates whether this PostgresReplication object is read or not.
	Ready bool `json:"ready"`

	// Represents a list of detailed status of the PostgresReplication object.
	// Each condition in the list provides real-time information about certain aspect of the PostgresReplication object.
	//
	// This field is crucial for administrators and developers to monitor and respond to changes within the PostgresReplication.
	// It provides a history of state transitions and a snapshot of the current state that can be used for
	// automated logic or direct inspection.
	//
	// +optional
	Conditions []metav1.Condition `json:"conditions,omitempty"`
}

// PostgresReplication is the Schema for the Postgres Replications API
// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
// +kubebuilder:resource:shortName=pr
// +kubebuilder:printcolumn:name="READY",type=boolean,JSONPath=`.status.ready`
// +kubebuilder:printcolumn:name="AGE",type="date",JSONPath=".metadata.creationTimestamp"
type PostgresReplication struct {
	// The metadata for the API version and kind of the PostgresReplication.
	metav1.TypeMeta `json:",inline"`

	// The metadata for the PostgresReplication object, including name, namespace, labels, and annotations.
	metav1.ObjectMeta `json:"metadata,omitempty"`

	// Defines the desired state of the PostgresReplication.
	Spec PostgresReplicationSpec `json:"spec,omitempty"`

	// Populated by the system, it represents the current information about the PostgresReplication.
	Status PostgresReplicationStatus `json:"status,omitempty"`
}

// PostgresReplicationList contains a list of PostgresReplication
// +kubebuilder:object:root=true
type PostgresReplicationList struct {
	// Contains the metadata for the API objects, including the Kind and Version of the object.
	metav1.TypeMeta `json:",inline"`

	// Contains the metadata for the list objects, including the continue and remainingItemCount for the list.
	metav1.ListMeta `json:"metadata,omitempty"`

	// Contains the list of PostgresReplication.
	Items []PostgresReplication `json:"items"`
}

func init() {
	SchemeBuilder.Register(&PostgresReplication{}, &PostgresReplicationList{})
}
