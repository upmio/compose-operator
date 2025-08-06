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
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/stretchr/testify/assert"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestNodeStatus(t *testing.T) {
	assert.Equal(t, NodeStatusOK, NodeStatus("Healthy"))
	assert.Equal(t, NodeStatusKO, NodeStatus("Failed"))
}

func TestServiceType(t *testing.T) {
	assert.Equal(t, ServiceTypeClusterIP, ServiceType("ClusterIP"))
	assert.Equal(t, ServiceTypeNodePort, ServiceType("NodePort"))
	assert.Equal(t, ServiceTypeLoadBalancer, ServiceType("LoadBalancer"))
	assert.Equal(t, ServiceTypeExternalName, ServiceType("ExternalName"))
}

func TestMysqlReplicationRole(t *testing.T) {
	assert.Equal(t, MysqlReplicationNodeRoleSource, MysqlReplicationRole("Source"))
	assert.Equal(t, MysqlReplicationNodeRoleReplica, MysqlReplicationRole("Replica"))
	assert.Equal(t, MysqlReplicationNodeRoleNone, MysqlReplicationRole("None"))
}

func TestPostgresReplicationRole(t *testing.T) {
	assert.Equal(t, PostgresReplicationRolePrimary, PostgresReplicationRole("Primary"))
	assert.Equal(t, PostgresReplicationRoleStandby, PostgresReplicationRole("Standby"))
	assert.Equal(t, PostgresReplicationRoleNone, PostgresReplicationRole("None"))
}

func TestRedisReplicationRole(t *testing.T) {
	assert.Equal(t, RedisReplicationNodeRoleSource, RedisReplicationRole("Source"))
	assert.Equal(t, RedisReplicationNodeRoleReplica, RedisReplicationRole("Replica"))
	assert.Equal(t, RedisReplicationNodeRoleNone, RedisReplicationRole("None"))
}

func TestRedisClusterNodeRole(t *testing.T) {
	assert.Equal(t, RedisClusterNodeRoleSource, RedisClusterNodeRole("Source"))
	assert.Equal(t, RedisClusterNodeRoleReplica, RedisClusterNodeRole("Replica"))
	assert.Equal(t, RedisClusterNodeRoleNone, RedisClusterNodeRole("None"))
}

func TestMysqlReplicationMode(t *testing.T) {
	assert.Equal(t, MysqlRplASync, MysqlReplicationMode("rpl_async"))
	assert.Equal(t, MysqlRplSemiSync, MysqlReplicationMode("rpl_semi_sync"))
}

func TestPostgresReplicationMode(t *testing.T) {
	assert.Equal(t, PostgresRplAsync, PostgresReplicationMode("rpl_async"))
	assert.Equal(t, PostgresRplSync, PostgresReplicationMode("rpl_sync"))
}

func TestRule(t *testing.T) {
	original := &Rule{
		Filter:  []string{"filter1", "filter2"},
		Pattern: "pattern",
	}

	deepCopy := original.DeepCopy()

	if diff := cmp.Diff(original, deepCopy); diff != "" {
		t.Errorf("Rule.DeepCopy() mismatch (-original +copy):\n%s", diff)
	}
}

func TestService(t *testing.T) {
	original := &Service{Type: ServiceTypeClusterIP}
	deepCopy := original.DeepCopy()

	if diff := cmp.Diff(original, deepCopy); diff != "" {
		t.Errorf("Service.DeepCopy() mismatch (-original +copy):\n%s", diff)
	}
}

func TestCommonNode(t *testing.T) {
	original := &CommonNode{
		Name: "TestNode",
		Host: "10.0.0.1",
		Port: 2214,
	}

	deepCopy := original.DeepCopy()

	if diff := cmp.Diff(original, deepCopy); diff != "" {
		t.Errorf("CommonNode.DeepCopy() mismatch (-original +copy):\n%s", diff)
	}
}

func TestCommonNodes(t *testing.T) {
	original := CommonNodes{
		&CommonNode{
			Name: "TestNode",
			Host: "10.0.0.1",
			Port: 2214,
		},
	}
	deepCopy := original.DeepCopy()

	if diff := cmp.Diff(original, deepCopy); diff != "" {
		t.Errorf("CommonNodes.DeepCopy() mismatch (-original +deepCopy):\n%s", diff)
	}
}

func TestDefaultMysqlReplicationOwnerReferences(t *testing.T) {
	instance := &MysqlReplication{
		ObjectMeta: metav1.ObjectMeta{
			Name: "mysqlreplication-instance",
		},
	}
	ownerRefs := DefaultMysqlReplicationOwnerReferences(instance)

	expectedOwnerRef := metav1.OwnerReference{
		APIVersion: "upm.syntropycloud.io/v1alpha1",
		Kind:       "MysqlReplication",
		Name:       "mysqlreplication-instance",
		Controller: new(bool),
	}
	*expectedOwnerRef.Controller = true

	assert.Equal(t, 1, len(ownerRefs))
	assert.Equal(t, expectedOwnerRef.APIVersion, ownerRefs[0].APIVersion)
	assert.Equal(t, expectedOwnerRef.Kind, ownerRefs[0].Kind)
	assert.Equal(t, expectedOwnerRef.Name, ownerRefs[0].Name)
	assert.Equal(t, expectedOwnerRef.Controller, ownerRefs[0].Controller)
}

func TestDefaultRedisReplicationOwnerReferences(t *testing.T) {
	instance := &RedisReplication{
		ObjectMeta: metav1.ObjectMeta{
			Name: "redisreplication-instance",
		},
	}
	ownerRefs := DefaultRedisReplicationOwnerReferences(instance)

	expectedOwnerRef := metav1.OwnerReference{
		APIVersion: "upm.syntropycloud.io/v1alpha1",
		Kind:       "RedisReplication",
		Name:       "redisreplication-instance",
		Controller: new(bool),
	}
	*expectedOwnerRef.Controller = true

	assert.Equal(t, 1, len(ownerRefs))
	assert.Equal(t, expectedOwnerRef.APIVersion, ownerRefs[0].APIVersion)
	assert.Equal(t, expectedOwnerRef.Kind, ownerRefs[0].Kind)
	assert.Equal(t, expectedOwnerRef.Name, ownerRefs[0].Name)
	assert.Equal(t, expectedOwnerRef.Controller, ownerRefs[0].Controller)
}

func TestDefaultPostgresReplicationOwnerReferences(t *testing.T) {
	instance := &PostgresReplication{
		ObjectMeta: metav1.ObjectMeta{
			Name: "postgresreplication-instance",
		},
	}
	ownerRefs := DefaultPostgresReplicationOwnerReferences(instance)

	expectedOwnerRef := metav1.OwnerReference{
		APIVersion: "upm.syntropycloud.io/v1alpha1",
		Kind:       "PostgresReplication",
		Name:       "postgresreplication-instance",
		Controller: new(bool),
	}
	*expectedOwnerRef.Controller = true

	assert.Equal(t, 1, len(ownerRefs))
	assert.Equal(t, expectedOwnerRef.APIVersion, ownerRefs[0].APIVersion)
	assert.Equal(t, expectedOwnerRef.Kind, ownerRefs[0].Kind)
	assert.Equal(t, expectedOwnerRef.Name, ownerRefs[0].Name)
	assert.Equal(t, expectedOwnerRef.Controller, ownerRefs[0].Controller)
}
