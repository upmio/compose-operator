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

// MongoDBReplicasetSecret defines the secret information of MongoDBReplicaset
type MongoDBReplicasetSecret struct {
	// Name is the name of the secret resource which store authentication information for MongoDB.
	Name string `json:"name"`
	// Mongod is the key of the secret, which contains the value used to connect to MongoDB.
	// +kubebuilder:default:=mongod
	Mongod string `json:"mongod"`
}

// MongoDBReplicasetSpec defines the desired state of MongoDBReplicaset
type MongoDBReplicasetSpec struct {

	// Secret is the reference to the secret resource containing authentication information, it must be in the same namespace as the MongoDBReplicaset object.
	Secret MongoDBReplicasetSecret `json:"secret"`

	// AESSecret is the reference to the secret resource containing aes key, it must be in the same namespace as the MongoDBReplicaset Object.
	AESSecret *AESSecret `json:"aesSecret,omitempty"`

	// Member is a list of nodes in the MongoDB Replica Set topology.
	Member CommonNodes `json:"member"`

	// ReplicaSetName is the name of the MongoDB replica set.
	ReplicaSetName string `json:"replicaSetName"`
}

type MongoDBReplicasetNode struct {
	// Host indicates the host of the MongoDB node.
	Host string `json:"host"`

	// Port indicates the port of the MongoDB node.
	Port int `json:"port"`

	// Role represents the role of the node in the replica set topology (e.g., primary, secondary, arbiter).
	Role MongoDBReplicasetRole `json:"role"`

	// Status indicates whether the node is ready for reads and writes.
	Status NodeStatus `json:"status"`

	// State indicates the replica set member state of the MongoDB node.
	State string `json:"state"`
}

type MongoDBReplicasetTopology map[string]*MongoDBReplicasetNode

// MongoDBReplicasetStatus defines the observed state of MongoDBReplicaset
type MongoDBReplicasetStatus struct {
	// Topology indicates the current MongoDB Replica Set topology.
	Topology MongoDBReplicasetTopology `json:"topology"`

	// Ready indicates whether this MongoDBReplicaset object is ready or not.
	Ready bool `json:"ready"`

	// Represents a list of detailed status of the MongoDBReplicaset object.
	// Each condition in the list provides real-time information about certain aspect of the MongoDBReplicaset object.
	//
	// This field is crucial for administrators and developers to monitor and respond to changes within the MongoDBReplicaset.
	// It provides a history of state transitions and a snapshot of the current state that can be used for
	// automated logic or direct inspection.
	//
	// +optional
	Conditions []metav1.Condition `json:"conditions,omitempty"`

	// ObservedGeneration The generation observed by controller.
	ObservedGeneration int64 `json:"observedGeneration"`
}

// MongoDBReplicaset is the Schema for the MongoDB Replica Set API
// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
// +kubebuilder:resource:shortName=mrs
// +kubebuilder:printcolumn:name="READY",type=boolean,JSONPath=`.status.ready`
// +kubebuilder:printcolumn:name="REPLICA_SET",type=string,JSONPath=`.spec.replicaSetName`
// +kubebuilder:printcolumn:name="AGE",type="date",JSONPath=".metadata.creationTimestamp"
type MongoDBReplicaset struct {
	// The metadata for the API version and kind of the MongoDBReplicaset.
	metav1.TypeMeta `json:",inline"`

	// The metadata for the MongoDBReplicaset object, including name, namespace, labels, and annotations.
	metav1.ObjectMeta `json:"metadata,omitempty"`

	// Defines the desired state of the MongoDBReplicaset.
	Spec MongoDBReplicasetSpec `json:"spec,omitempty"`

	// Populated by the system, it represents the current information about the MongoDBReplicaset.
	Status MongoDBReplicasetStatus `json:"status,omitempty"`
}

// MongoDBReplicasetList contains a list of MongoDBReplicaset
// +kubebuilder:object:root=true
type MongoDBReplicasetList struct {
	// Contains the metadata for the API objects, including the Kind and Version of the object.
	metav1.TypeMeta `json:",inline"`

	// Contains the metadata for the list objects, including the continue and remainingItemCount for the list.
	metav1.ListMeta `json:"metadata,omitempty"`

	// Contains the list of MongoDBReplicaset.
	Items []MongoDBReplicaset `json:"items"`
}

func init() {
	SchemeBuilder.Register(&MongoDBReplicaset{}, &MongoDBReplicasetList{})
}
