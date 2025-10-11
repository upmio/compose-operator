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

// RedisClusterSecret defines the secret information of RedisCluster
type RedisClusterSecret struct {
	// Name is the name of the secret resource which store authentication information for Redis.
	Name string `json:"name"`

	//  Redis is the key of the secret, which contains the value used to connect to Redis.
	Redis string `json:"redis"`
}

// RedisClusterSpec defines the desired state of RedisCluster
type RedisClusterSpec struct {
	// Secret is the reference to the secret resource containing authentication information, it must be in the same namespace as the RedisReplication object.
	Secret RedisClusterSecret `json:"secret"`

	// AESSecret is the reference to the secret resource containing aes key, it must be in the same namespace as the RedisReplication Object.
	AESSecret *AESSecret `json:"aesSecret,omitempty"`

	// Members is a list of nodes in the Redis Cluster topology.
	Members map[string]CommonNodes `json:"members"`
}

// RedisClusterNode represent a RedisCluster Node
type RedisClusterNode struct {
	// ID represents the id of the Redis node.
	ID string `json:"id"`

	// Role represents the role of the node in the redis cluster topology (e.g., source, replica).
	Role RedisClusterNodeRole `json:"role"`

	// Status indicates the current status of the node (e.g., Healthy, Failed).
	Status NodeStatus `json:"status"`

	// Host indicates the host of the Redis node.
	Host string `json:"host"`

	// Port indicates the port of the Redis node.
	Port int `json:"port"`

	// Slots indicates the slots assigned to this Redis node.
	Slots []string `json:"slots,omitempty"`

	// MasterRef indicates which source node this replica node reference.
	MasterRef string `json:"masterRef,omitempty"`

	// Shard indicates which shard this node belong with.
	Shard string `json:"shard"`

	// Ready indicates whether the node is ready for reads and writes.
	Ready bool `json:"ready"`
}

// RedisClusterTopology defines the RedisCluster topology
type RedisClusterTopology map[string]*RedisClusterNode

// RedisClusterStatus defines the observed state of RedisCluster
type RedisClusterStatus struct {
	// Topology indicates the current Redis replication topology.
	Topology RedisClusterTopology `json:"topology"`

	// NumberOfShard indicates the number of Redis Cluster shard.
	NumberOfShard int `json:"numberOfShard"`

	// ClusterJoined indicates whether all node have joined the cluster.
	ClusterJoined bool `json:"clusterJoined"`

	// ClusterJoined indicates whether all node in this cluster have synced.
	ClusterSynced bool `json:"clusterSynced"`

	// Ready indicates whether this RedisCluster object is read or not.
	Ready bool `json:"ready"`

	// Represents a list of detailed status of the RedisCluster object.
	// Each condition in the list provides real-time information about certain aspect of the RedisCluster object.
	//
	// This field is crucial for administrators and developers to monitor and respond to changes within the RedisCluster.
	// It provides a history of state transitions and a snapshot of the current state that can be used for
	// automated logic or direct inspection.
	//
	// +optional
	Conditions []metav1.Condition `json:"conditions,omitempty"`

	// ObservedGeneration The generation observed by controller.
	ObservedGeneration int64 `json:"observedGeneration"`
}

// RedisCluster is the Schema for the Redis Cluster API
// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
// +kubebuilder:resource:shortName=rcc
// +kubebuilder:printcolumn:name="SHARD",type=integer,JSONPath=`.status.numberOfShard`
// +kubebuilder:printcolumn:name="READY",type=boolean,JSONPath=`.status.ready`
// +kubebuilder:printcolumn:name="AGE",type="date",JSONPath=".metadata.creationTimestamp"
type RedisCluster struct {
	// The metadata for the API version and kind of the RedisCluster.
	metav1.TypeMeta `json:",inline"`

	// The metadata for the RedisCluster object, including name, namespace, labels, and annotations.
	metav1.ObjectMeta `json:"metadata,omitempty"`

	// Defines the desired state of the RedisCluster.
	Spec RedisClusterSpec `json:"spec,omitempty"`

	// Populated by the system, it represents the current information about the RedisCluster.
	Status RedisClusterStatus `json:"status,omitempty"`
}

// RedisClusterList contains a list of RedisCluster
// +kubebuilder:object:root=true
type RedisClusterList struct {
	// Contains the metadata for the API objects, including the Kind and Version of the object.
	metav1.TypeMeta `json:",inline"`

	// Contains the metadata for the list objects, including the continue and remainingItemCount for the list.
	metav1.ListMeta `json:"metadata,omitempty"`

	// Contains the list of RedisCluster.
	Items []RedisCluster `json:"items"`
}

func init() {
	SchemeBuilder.Register(&RedisCluster{}, &RedisClusterList{})
}
