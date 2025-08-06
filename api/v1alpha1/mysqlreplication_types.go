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

// MysqlReplicationSecret defines the secret information of MysqlReplication
type MysqlReplicationSecret struct {
	// Name is the name of the secret resource which store authentication information for MySQL.
	Name string `json:"name"`
	//  Mysql is the key of the secret, which contains the value used to connect to MySQL.
	// +kubebuilder:default:=mysql
	Mysql string `json:"mysql"`
	// Replication is the key of the secret, which contains the value used to set up MySQL replication.
	// +kubebuilder:default:=replication
	Replication string `json:"replication"`
}

// MysqlReplicationSpec defines the desired state of MysqlReplication
type MysqlReplicationSpec struct {
	// Mode specifies the mysql replication sync mode.
	// Valid values are:
	// - "rpl_semi_sync": semi_sync;
	// - "rpl_async": async;
	// +optional
	// +kubebuilder:default:=rpl_async
	Mode MysqlReplicationMode `json:"mode"`

	// Secret is the reference to the secret resource containing authentication information, it must be in the same namespace as the MysqlReplication object.
	Secret MysqlReplicationSecret `json:"secret"`

	// Source references the source MySQL node.
	Source *CommonNode `json:"source"`

	// Service references the service providing the MySQL replication endpoint.
	Service *Service `json:"service"`

	// Replica is a list of replica nodes in the MySQL replication topology.
	Replica ReplicaNodes `json:"replica,omitempty"`
}

// MysqlReplicationNode represents a node in the MySQL replication topology.
type MysqlReplicationNode struct {
	// Host indicates the host of the MySQL node.
	Host string `json:"host"`

	// Port indicates the port of the MySQL node.
	Port int `json:"port"`

	// Role represents the role of the node in the replication topology (e.g., source, replica).
	Role MysqlReplicationRole `json:"role"`

	// Status indicates the current status of the node (e.g., Healthy, Failed).
	Status NodeStatus `json:"status"`

	// Ready indicates whether the node is ready for reads and writes.
	Ready bool `json:"ready"`

	// ReadOnly specifies whether the node is read-only.
	ReadOnly bool `json:"readonly"`

	// SuperReadOnly specifies whether the node is super-read-only (i.e., cannot even write to its own database).
	SuperReadOnly bool `json:"superReadonly"`

	// SourceHost indicates the hostname or IP address of the source node that this replica node is replicating from.
	SourceHost string `json:"sourceHost,omitempty"`

	// SourcePort indicates the port of the source node that this replica node is replicating from.
	SourcePort int `json:"sourcePort,omitempty"`

	// ReplicaIO indicates the status of I/O thread of the replica node.
	ReplicaIO string `json:"replicaIO,omitempty"`

	// ReplicaSQL indicates the status of SQL thread of the replica node.
	ReplicaSQL string `json:"replicaSQL,omitempty"`

	// ReadSourceLogPos the position in the source node's binary log file where the replica node should start reading from.
	ReadSourceLogPos int `json:"readSourceLogPos,omitempty"`

	// SourceLogFile indicates the name of the binary log file on the source node that the replica node should read from.
	SourceLogFile string `json:"sourceLogFile,omitempty"`

	// SecondsBehindSource indicates the metric that shows how far behind the source node the replica node is, measured in seconds.
	SecondsBehindSource *int `json:"secondsBehindSource,omitempty"`

	// ExecSourceLogPos the position in the source node's binary log file where the replica node should execute from.
	ExecSourceLogPos int `json:"execSourceLogPos,omitempty"`
}

// MysqlReplicationTopology defines the MysqlReplication topology
type MysqlReplicationTopology map[string]*MysqlReplicationNode

// MysqlReplicationStatus defines the observed state of MysqlReplication
type MysqlReplicationStatus struct {
	// Topology indicates the current MySQL replication topology.
	Topology MysqlReplicationTopology `json:"topology"`

	// ReadWriteService specify the service name provides read-write access to database.
	ReadWriteService string `json:"readwriteService"`

	// ReadOnlyService specify the service name provides read-only access to database.
	ReadOnlyService string `json:"readonlyService,omitempty"`

	// Ready indicates whether this MysqlReplication object is ready or not.
	Ready bool `json:"ready"`

	// Represents a list of detailed status of the MysqlReplication object.
	// Each condition in the list provides real-time information about certain aspect of the MysqlReplication object.
	//
	// This field is crucial for administrators and developers to monitor and respond to changes within the MysqlReplication.
	// It provides a history of state transitions and a snapshot of the current state that can be used for
	// automated logic or direct inspection.
	//
	// +optional
	Conditions []metav1.Condition `json:"conditions,omitempty"`
}

// MysqlReplication is the Schema for the Mysql Replication API
// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
// +kubebuilder:resource:shortName=mr
// +kubebuilder:printcolumn:name="READY",type=boolean,JSONPath=`.status.ready`
// +kubebuilder:printcolumn:name="AGE",type="date",JSONPath=".metadata.creationTimestamp"
type MysqlReplication struct {
	// The metadata for the API version and kind of the MysqlReplication.
	metav1.TypeMeta `json:",inline"`

	// The metadata for the MysqlReplication object, including name, namespace, labels, and annotations.
	metav1.ObjectMeta `json:"metadata,omitempty"`

	// Defines the desired state of the MysqlReplication.
	Spec MysqlReplicationSpec `json:"spec,omitempty"`

	// Populated by the system, it represents the current information about the MysqlReplication.
	Status MysqlReplicationStatus `json:"status,omitempty"`
}

// MysqlReplicationList contains a list of MysqlReplication
// +kubebuilder:object:root=true
type MysqlReplicationList struct {
	// Contains the metadata for the API objects, including the Kind and Version of the object.
	metav1.TypeMeta `json:",inline"`

	// Contains the metadata for the list objects, including the continue and remainingItemCount for the list.
	metav1.ListMeta `json:"metadata,omitempty"`

	// Contains the list of MysqlReplication.
	Items []MysqlReplication `json:"items"`
}

func init() {
	SchemeBuilder.Register(&MysqlReplication{}, &MysqlReplicationList{})
}
