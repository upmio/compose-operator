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

// MysqlGroupReplicationSecret defines the secret information of MysqlGroupReplication
type MysqlGroupReplicationSecret struct {
	// Name is the name of the secret resource which store authentication information for MySQL.
	Name string `json:"name"`
	//  Mysql is the key of the secret, which contains the value used to connect to MySQL.
	// +kubebuilder:default:=mysql
	Mysql string `json:"mysql"`
	// Replication is the key of the secret, which contains the value used to set up MySQL Group Replication.
	// +kubebuilder:default:=replication
	Replication string `json:"replication"`
}

// MysqlGroupReplicationSpec defines the desired state of MysqlGroupReplication
type MysqlGroupReplicationSpec struct {

	// Secret is the reference to the secret resource containing authentication information, it must be in the same namespace as the MysqlGroupReplication object.
	Secret MysqlGroupReplicationSecret `json:"secret"`

	// AESSecret is the reference to the secret resource containing aes key, it must be in the same namespace as the MysqlGroupReplication Object.
	AESSecret *AESSecret `json:"aesSecret,omitempty"`

	// Member is a list of nodes in the MySQL Group Replication topology.
	Member CommonNodes `json:"member"`
}

type MysqlGroupReplicationNode struct {
	// Host indicates the host of the MySQL node.
	Host string `json:"host"`

	// Port indicates the port of the MySQL node.
	Port int `json:"port"`

	// Role represents the role of the node in the group replication topology (e.g., primary, secondary).
	Role MysqlGroupReplicationRole `json:"role"`

	// Ready indicates whether the node is ready for reads and writes.
	Status NodeStatus `json:"status"`

	// GtidExecuted indicates the gtid_executed of the MySQL node.
	GtidExecuted string `json:"gtidExecuted"`

	// MemberState indicates the member_state of the MySQL node.
	MemberState string `json:"memberState"`

	// ReadOnly specifies whether the node is read-only.
	ReadOnly bool `json:"readonly"`

	// SuperReadOnly specifies whether the node is super-read-only (i.e., cannot even write to its own database).
	SuperReadOnly bool `json:"superReadonly"`
}

type MysqlGroupReplicationTopology map[string]*MysqlGroupReplicationNode

// MysqlGroupReplicationStatus defines the observed state of MysqlGroupReplication
type MysqlGroupReplicationStatus struct {
	// Topology indicates the current MySQL Group Replication topology.
	Topology MysqlGroupReplicationTopology `json:"topology"`

	// Ready indicates whether this MysqlGroupReplication object is ready or not.
	Ready bool `json:"ready"`

	// Represents a list of detailed status of the MysqlGroupReplication object.
	// Each condition in the list provides real-time information about certain aspect of the MysqlGroupReplication object.
	//
	// This field is crucial for administrators and developers to monitor and respond to changes within the MysqlGroupReplication.
	// It provides a history of state transitions and a snapshot of the current state that can be used for
	// automated logic or direct inspection.
	//
	// +optional
	Conditions []metav1.Condition `json:"conditions,omitempty"`
}

// MysqlGroupReplication is the Schema for the Mysql Group Replication API
// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
// +kubebuilder:resource:shortName=mgr
// +kubebuilder:printcolumn:name="READY",type=boolean,JSONPath=`.status.ready`
// +kubebuilder:printcolumn:name="AGE",type="date",JSONPath=".metadata.creationTimestamp"
type MysqlGroupReplication struct {
	// The metadata for the API version and kind of the MysqlReplication.
	metav1.TypeMeta `json:",inline"`

	// The metadata for the MysqlReplication object, including name, namespace, labels, and annotations.
	metav1.ObjectMeta `json:"metadata,omitempty"`

	// Defines the desired state of the MysqlGroupReplication.
	Spec MysqlGroupReplicationSpec `json:"spec,omitempty"`

	// Populated by the system, it represents the current information about the MysqlGroupReplication.
	Status MysqlGroupReplicationStatus `json:"status,omitempty"`
}

// MysqlGroupReplicationList contains a list of MysqlGroupReplication
// +kubebuilder:object:root=true
type MysqlGroupReplicationList struct {
	// Contains the metadata for the API objects, including the Kind and Version of the object.
	metav1.TypeMeta `json:",inline"`

	// Contains the metadata for the list objects, including the continue and remainingItemCount for the list.
	metav1.ListMeta `json:"metadata,omitempty"`

	// Contains the list of MysqlGroupReplication.
	Items []MysqlGroupReplication `json:"items"`
}

func init() {
	SchemeBuilder.Register(&MysqlGroupReplication{}, &MysqlGroupReplicationList{})
}
