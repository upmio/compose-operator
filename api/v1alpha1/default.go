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
	"k8s.io/apimachinery/pkg/runtime/schema"
)

// NodeStatus defines node status
type NodeStatus string

const (
	// NodeStatusOK Status OK
	NodeStatusOK NodeStatus = "Healthy"
	// NodeStatusKO Status KO
	NodeStatusKO NodeStatus = "Failed"
)

const SkipReconcileKey = "compose-operator.skip.reconcile"

// CommonNode information for node to connect
type CommonNode struct {
	// Name specifies the identifier of node
	Name string `json:"name"`

	// Host specifies the ip or hostname of node
	Host string `json:"host"`

	// Port specifies the port of node
	Port int `json:"port"`
}

// CommonNodes array Node
type CommonNodes []*CommonNode

// ReplicaNode is a CommonNode with additional replication-specific info
type ReplicaNode struct {
	CommonNode `json:",inline"` // 嵌入基础字段

	// Isolated indicates whether this node will be isolated.
	// If not specified, it defaults to false.
	// +kubebuilder:default=false
	// +optional
	Isolated bool `json:"isolated,omitempty"`
}

type ReplicaNodes []*ReplicaNode

// ServiceType reflection of kubernetes service type
// +kubebuilder:validation:Enum=ClusterIP;NodePort;LoadBalancer;ExternalName
type ServiceType string

const (
	// ServiceTypeClusterIP means a service will only be accessible inside the
	// cluster, via the cluster IP.
	ServiceTypeClusterIP ServiceType = "ClusterIP"

	// ServiceTypeNodePort means a service will be exposed on one port of
	// every node, in addition to 'ClusterIP' type.
	ServiceTypeNodePort ServiceType = "NodePort"

	// ServiceTypeLoadBalancer means a service will be exposed via an
	// external load balancer (if the cloud provider supports it), in addition
	// to 'NodePort' type.
	ServiceTypeLoadBalancer ServiceType = "LoadBalancer"

	// ServiceTypeExternalName means a service consists of only a reference to
	// an external name that kubedns or equivalent will return as a CNAME
	// record, with no exposing or proxying of any pods involved.
	ServiceTypeExternalName ServiceType = "ExternalName"
)

// Service reflection of kubernetes service
type Service struct {
	//Type string describes ingress methods for a service
	// +kubebuilder:default:=ClusterIP
	Type ServiceType `json:"type"`
}

// MysqlReplicationRole defines the mysql replication role
type MysqlReplicationRole string

const (
	// MysqlReplicationNodeRoleSource MysqlReplication Source node role
	MysqlReplicationNodeRoleSource MysqlReplicationRole = "Source"
	// MysqlReplicationNodeRoleReplica MysqlReplication Replica node role
	MysqlReplicationNodeRoleReplica MysqlReplicationRole = "Replica"
	// MysqlReplicationNodeRoleNone MysqlReplication None node role
	MysqlReplicationNodeRoleNone MysqlReplicationRole = "None"
)

// PostgresReplicationRole defines the postgres replication role
type PostgresReplicationRole string

const (
	// PostgresReplicationRolePrimary PostgresReplicationRole Primary node role
	PostgresReplicationRolePrimary PostgresReplicationRole = "Primary"
	// PostgresReplicationRoleStandby PostgresReplicationRole Standby node role
	PostgresReplicationRoleStandby PostgresReplicationRole = "Standby"
	// PostgresReplicationRoleNone PostgresReplicationRole None node role
	PostgresReplicationRoleNone PostgresReplicationRole = "None"
)

// MysqlGroupReplicationRole defines the mysql group replication role
type MysqlGroupReplicationRole string

const (
	// MysqlGroupReplicationNodeRolePrimary MysqlGroupReplication Primary node role
	MysqlGroupReplicationNodeRolePrimary MysqlGroupReplicationRole = "Primary"
	// MysqlGroupReplicationNodeRoleSecondary MysqlGroupReplication Secondary node role
	MysqlGroupReplicationNodeRoleSecondary MysqlGroupReplicationRole = "Secondary"
	// MysqlGroupReplicationNodeRoleNone MysqlGroupReplication None node role
	MysqlGroupReplicationNodeRoleNone MysqlGroupReplicationRole = "None"
)

// RedisReplicationRole defines the redis replication role
type RedisReplicationRole string

const (
	// RedisReplicationNodeRoleSource RedisReplication Source node role
	RedisReplicationNodeRoleSource RedisReplicationRole = "Source"
	// RedisReplicationNodeRoleReplica RedisReplication Replica node role
	RedisReplicationNodeRoleReplica RedisReplicationRole = "Replica"
	// RedisReplicationNodeRoleNone RedisReplication None node role
	RedisReplicationNodeRoleNone RedisReplicationRole = "None"
)

// RedisClusterNodeRole defines the redis cluster role
type RedisClusterNodeRole string

const (
	// RedisClusterNodeRoleSource RedisCluster Source node role
	RedisClusterNodeRoleSource RedisClusterNodeRole = "Source"
	// RedisClusterNodeRoleReplica RedisCluster Replica node role
	RedisClusterNodeRoleReplica RedisClusterNodeRole = "Replica"
	// RedisClusterNodeRoleNone RedisCluster None node role
	RedisClusterNodeRoleNone RedisClusterNodeRole = "None"
)

// MysqlReplicationMode describes how the mysql replication will be handled.
// Only one of the following sync mode must be specified.
// +kubebuilder:validation:Enum=rpl_async;rpl_semi_sync
type MysqlReplicationMode string

const (
	MysqlRplASync    MysqlReplicationMode = "rpl_async"
	MysqlRplSemiSync MysqlReplicationMode = "rpl_semi_sync"
)

// PostgresReplicationMode describes how the postgres stream replication will be handled.
// Only one of the following sync mode must be specified.
// +kubebuilder:validation:Enum=rpl_async;rpl_sync
type PostgresReplicationMode string

const (
	PostgresRplAsync PostgresReplicationMode = "rpl_async"
	PostgresRplSync  PostgresReplicationMode = "rpl_sync"
)

// Rule defines the proxysqlsync rule
type Rule struct {
	Filter  []string `json:"filter"`
	Pattern string   `json:"pattern"`
}

const (
	ConditionTypeTopologyReady = "TopologyReady"
	ConditionTypeResourceReady = "ResourceReady"
	ConditionTypeUserReady     = "UserReady"
	ConditionTypeServerReady   = "ServerReady"
)

func DefaultMysqlReplicationOwnerReferences(instance *MysqlReplication) []metav1.OwnerReference {
	return []metav1.OwnerReference{
		*metav1.NewControllerRef(instance, schema.GroupVersionKind{
			Group:   GroupVersion.Group,
			Version: GroupVersion.Version,
			Kind:    "MysqlReplication",
		}),
	}
}

func DefaultRedisReplicationOwnerReferences(instance *RedisReplication) []metav1.OwnerReference {
	return []metav1.OwnerReference{
		*metav1.NewControllerRef(instance, schema.GroupVersionKind{
			Group:   GroupVersion.Group,
			Version: GroupVersion.Version,
			Kind:    "RedisReplication",
		}),
	}
}

func DefaultPostgresReplicationOwnerReferences(instance *PostgresReplication) []metav1.OwnerReference {
	return []metav1.OwnerReference{
		*metav1.NewControllerRef(instance, schema.GroupVersionKind{
			Group:   GroupVersion.Group,
			Version: GroupVersion.Version,
			Kind:    "PostgresReplication",
		}),
	}
}
