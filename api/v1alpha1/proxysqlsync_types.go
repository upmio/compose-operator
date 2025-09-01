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

// ProxysqlSyncSecret defines the secret information of ProxysqlSync
type ProxysqlSyncSecret struct {
	// Name is the name of the secret resource which store authentication information for MySQL and ProxySQL.
	Name string `json:"name"`

	//  Proxysql is the key of the secret, which contains the value used to connect to ProxySQL.
	Proxysql string `json:"proxysql"`

	//  Mysql is the key of the secret, which contains the value used to connect to MySQL.
	Mysql string `json:"mysql"`
}

// ProxysqlSyncSpec defines the desired state of ProxysqlSync
type ProxysqlSyncSpec struct {

	// Proxysql references the list of proxysql nodes.
	Proxysql CommonNodes `json:"proxysql"`

	// Secret is the reference to the secret resource containing authentication information, it must be in the same namespace as the ProxysqlSync object.
	Secret ProxysqlSyncSecret `json:"secret"`

	// AESSecret is the reference to the secret resource containing aes key, it must be in the same namespace as the ProxysqlSync Object.
	AESSecret *AESSecret `json:"aesSecret,omitempty"`

	// MysqlReplication references the name of MysqlReplication.
	MysqlReplication string `json:"mysqlReplication"`

	// Rule references the rule of sync users from MySQL to ProxySQL.
	Rule *Rule `json:"rule"`
}

type ProxysqlSyncNode struct {
	// Users references the user list synced from MySQL to ProxySQL
	Users []string `json:"users,omitempty"`

	// Ready references whether ProxySQL server synced from MysqlReplication is correct.
	Synced bool `json:"synced"`
}

// ProxysqlSyncTopology defines the ProxysqlSync topology
type ProxysqlSyncTopology map[string]*ProxysqlSyncNode

// ProxysqlSyncStatus defines the observed state of ProxysqlSync
type ProxysqlSyncStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file
	Topology ProxysqlSyncTopology `json:"topology"`

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
}

// ProxysqlSync is the Schema for the proxysqlsyncs API
// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
// +kubebuilder:resource:shortName=ps
// +kubebuilder:printcolumn:name="READY",type=boolean,JSONPath=`.status.ready`
// +kubebuilder:printcolumn:name="AGE",type="date",JSONPath=".metadata.creationTimestamp"
type ProxysqlSync struct {
	// The metadata for the API version and kind of the ProxysqlSync.
	metav1.TypeMeta `json:",inline"`

	// The metadata for the ProxysqlSync object, including name, namespace, labels, and annotations.
	metav1.ObjectMeta `json:"metadata,omitempty"`

	// Defines the desired state of the ProxysqlSync.
	Spec ProxysqlSyncSpec `json:"spec,omitempty"`

	// Populated by the system, it represents the current information about the ProxysqlSync.
	Status ProxysqlSyncStatus `json:"status,omitempty"`
}

// ProxysqlSyncList contains a list of ProxysqlSync
// +kubebuilder:object:root=true
type ProxysqlSyncList struct {
	// Contains the metadata for the API objects, including the Kind and Version of the object.
	metav1.TypeMeta `json:",inline"`

	// Contains the metadata for the list objects, including the continue and remainingItemCount for the list.
	metav1.ListMeta `json:"metadata,omitempty"`

	// Contains the list of ProxysqlSync.
	Items []ProxysqlSync `json:"items"`
}

func init() {
	SchemeBuilder.Register(&ProxysqlSync{}, &ProxysqlSyncList{})
}
