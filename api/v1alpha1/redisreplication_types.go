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

// RedisReplicationSecret defines the secret information of RedisReplication
type RedisReplicationSecret struct {
	// Name is the name of the secret resource which store authentication information for Redis.
	Name string `json:"name"`

	//  Redis is the key of the secret, which contains the value used to connect to Redis.
	Redis string `json:"redis"`
}

// RedisReplicationSpec defines the desired state of RedisReplication
type RedisReplicationSpec struct {
	// Secret is the reference to the secret resource containing authentication information, it must be in the same namespace as the RedisReplication object.
	Secret RedisReplicationSecret `json:"secret"`

	// Source references the source Redis node.
	Source *CommonNode `json:"source"`

	// Replica is a list of replica nodes in the Redis replication topology.
	Replica CommonNodes `json:"replica,omitempty"`

	// Service references the service providing the Redis replication endpoint.
	Service *Service `json:"service"`
}

// RedisReplicationNode represents a node in the Redis replication topology.
type RedisReplicationNode struct {
	// Host indicates the host of the Redis node.
	Host string `json:"host"`

	// Port indicates the port of the Redis node.
	Port int `json:"port"`

	// Role represents the role of the node in the replication topology (e.g., source, replica).
	Role RedisReplicationRole `json:"role"`

	// Status indicates the current status of the node (e.g., Healthy, Failed).
	Status NodeStatus `json:"status"`

	// Ready indicates whether the node is ready for reads and writes.
	Ready bool `json:"ready"`

	// SourceHost indicates the hostname or IP address of the source node that this replica node is replicating from.
	SourceHost string `json:"sourceHost,omitempty"`

	// SourcePort indicates the port of the source node that this replica node is replicating from.
	SourcePort int `json:"sourcePort,omitempty"`
}

// RedisReplicationTopology defines the RedisReplication topology
type RedisReplicationTopology map[string]*RedisReplicationNode

// RedisReplicationStatus defines the observed state of RedisReplication
type RedisReplicationStatus struct {
	// Topology indicates the current Redis replication topology.
	Topology RedisReplicationTopology `json:"topology"`

	// ReadWriteService specify the service name provides read-write access to database.
	ReadWriteService string `json:"readwriteService"`

	// ReadOnlyService specify the service name provides read-only access to database.
	ReadOnlyService string `json:"readonlyService,omitempty"`

	// Ready indicates whether this RedisReplication object is read or not.
	Ready bool `json:"ready"`

	// Represents a list of detailed status of the RedisReplication object.
	// Each condition in the list provides real-time information about certain aspect of the RedisReplication object.
	//
	// This field is crucial for administrators and developers to monitor and respond to changes within the RedisReplication.
	// It provides a history of state transitions and a snapshot of the current state that can be used for
	// automated logic or direct inspection.
	//
	// +optional
	Conditions []metav1.Condition `json:"conditions,omitempty"`
}

// RedisReplication is the Schema for the Redis Replication API
// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
// +kubebuilder:resource:shortName=rr
// +kubebuilder:printcolumn:name="READY",type=boolean,JSONPath=`.status.ready`
// +kubebuilder:printcolumn:name="AGE",type="date",JSONPath=".metadata.creationTimestamp"
type RedisReplication struct {
	// The metadata for the API version and kind of the RedisReplication.
	metav1.TypeMeta `json:",inline"`

	// The metadata for the RedisReplication object, including name, namespace, labels, and annotations.
	metav1.ObjectMeta `json:"metadata,omitempty"`

	// Defines the desired state of the RedisReplication.
	Spec RedisReplicationSpec `json:"spec,omitempty"`

	// Populated by the system, it represents the current information about the RedisReplication.
	Status RedisReplicationStatus `json:"status,omitempty"`
}

// RedisReplicationList contains a list of RedisReplication
// +kubebuilder:object:root=true
type RedisReplicationList struct {
	// Contains the metadata for the API objects, including the Kind and Version of the object.
	metav1.TypeMeta `json:",inline"`

	// Contains the metadata for the list objects, including the continue and remainingItemCount for the list.
	metav1.ListMeta `json:"metadata,omitempty"`

	// Contains the list of RedisReplication.
	Items []RedisReplication `json:"items"`
}

func init() {
	SchemeBuilder.Register(&RedisReplication{}, &RedisReplicationList{})
}
