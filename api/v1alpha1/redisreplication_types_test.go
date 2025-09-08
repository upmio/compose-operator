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
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestRedisReplication(t *testing.T) {
	original := &RedisReplication{
		TypeMeta: metav1.TypeMeta{
			Kind:       "RedisReplication",
			APIVersion: "v1alpha1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      "redisreplication-sample",
			Namespace: "test",
		},
		Spec: RedisReplicationSpec{
			Secret: RedisReplicationSecret{
				Name:  "redisreplication-secret",
				Redis: "root",
			},
			Source: &RedisNode{
				CommonNode: CommonNode{
					Name: "node01",
					Host: "10.0.0.1",
					Port: 6379,
				},
			},
			Service: &Service{Type: ServiceTypeClusterIP},
			Replica: RedisNodes{
				&RedisNode{
					CommonNode: CommonNode{
						Name: "node02",
						Host: "10.0.0.2",
						Port: 6379,
					},
				},
			},
		},
		Status: RedisReplicationStatus{
			Topology: RedisReplicationTopology{
				"node01": &RedisReplicationNode{
					Host:   "10.0.0.1",
					Port:   6379,
					Role:   RedisReplicationNodeRoleSource,
					Status: NodeStatusOK,
					Ready:  true,
				},
				"node02": &RedisReplicationNode{
					Host:       "10.0.0.2",
					Port:       6379,
					Role:       RedisReplicationNodeRoleReplica,
					Status:     NodeStatusOK,
					Ready:      true,
					SourceHost: "10.0.0.1",
					SourcePort: 6379,
				},
			},
			ReadWriteService: "redisreplication-sample-readwrite",
			ReadOnlyService:  "redisreplication-sample-readonly",
			Ready:            true,
			Conditions: []metav1.Condition{
				{
					Type:    ConditionTypeTopologyReady,
					Status:  metav1.ConditionTrue,
					Message: "Successfully sync topology",
					Reason:  "SyncTopologySucceed",
				},
				{
					Type:    ConditionTypeResourceReady,
					Status:  metav1.ConditionTrue,
					Message: "Successfully sync resource",
					Reason:  "SyncResourceSucceed",
				},
			},
		},
	}

	deepCopy := original.DeepCopyObject()

	if diff := cmp.Diff(original, deepCopy); diff != "" {
		t.Errorf("RedisReplication.DeepCopy() mismatch (-original +copy):\n%s", diff)
	}
}

func TestRedisReplicationList(t *testing.T) {
	original := &RedisReplicationList{
		TypeMeta: metav1.TypeMeta{
			Kind:       "RedisReplicationList",
			APIVersion: "v1alpha1",
		},
		Items: []RedisReplication{
			{
				TypeMeta: metav1.TypeMeta{
					Kind:       "RedisReplication",
					APIVersion: "v1alpha1",
				},
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test",
					Namespace: "default",
				},
				Spec: RedisReplicationSpec{
					Secret: RedisReplicationSecret{
						Name:  "redisreplication-secret",
						Redis: "root",
					},
					Source: &RedisNode{
						CommonNode: CommonNode{
							Name: "node01",
							Host: "10.0.0.1",
							Port: 6379,
						},
					},
					Service: &Service{Type: ServiceTypeClusterIP},
					Replica: RedisNodes{
						&RedisNode{
							CommonNode: CommonNode{
								Name: "node02",
								Host: "10.0.0.2",
								Port: 6379,
							},
						},
					},
				},
				Status: RedisReplicationStatus{
					Topology: RedisReplicationTopology{
						"node01": &RedisReplicationNode{
							Host:   "10.0.0.1",
							Port:   6379,
							Role:   RedisReplicationNodeRoleSource,
							Status: NodeStatusOK,
							Ready:  true,
						},
						"node02": &RedisReplicationNode{
							Host:       "10.0.0.2",
							Port:       3306,
							Role:       RedisReplicationNodeRoleReplica,
							Status:     NodeStatusOK,
							Ready:      true,
							SourceHost: "10.0.0.1",
							SourcePort: 6379,
						},
					},
					ReadWriteService: "redisreplication-sample-readwrite",
					ReadOnlyService:  "redisreplication-sample-readonly",
					Ready:            true,
					Conditions: []metav1.Condition{
						{
							Type:    ConditionTypeTopologyReady,
							Status:  metav1.ConditionTrue,
							Message: "Successfully sync topology",
							Reason:  "SyncTopologySucceed",
						},
						{
							Type:    ConditionTypeResourceReady,
							Status:  metav1.ConditionTrue,
							Message: "Successfully sync resource",
							Reason:  "SyncResourceSucceed",
						},
					},
				},
			},
		},
	}
	deepCopy := original.DeepCopyObject()

	if diff := cmp.Diff(original, deepCopy); diff != "" {
		t.Errorf("RedisReplicationList.DeepCopy() mismatch (-original +copy):\n%s", diff)
	}
}

func TestRedisReplicationNode(t *testing.T) {
	original := &RedisReplicationNode{
		Host:   "10.0.0.1",
		Port:   6379,
		Role:   RedisReplicationNodeRoleSource,
		Status: NodeStatusOK,
		Ready:  true,
	}

	deepCopy := original.DeepCopy()

	if diff := cmp.Diff(original, deepCopy); diff != "" {
		t.Errorf("RedisReplicationNode.DeepCopy() mismatch (-original +copy):\n%s", diff)
	}
}

func TestRedisReplicationSecret(t *testing.T) {
	original := &RedisReplicationSecret{
		Name:  "redisreplication-secret",
		Redis: "root",
	}

	deepCopy := original.DeepCopy()

	if diff := cmp.Diff(original, deepCopy); diff != "" {
		t.Errorf("RedisReplicationSecret.DeepCopy() mismatch (-original +copy):\n%s", diff)
	}
}

func TestRedisReplicationSpec(t *testing.T) {
	original := &RedisReplicationSpec{
		Secret: RedisReplicationSecret{
			Name:  "redisreplication-secret",
			Redis: "root",
		},
		Source: &RedisNode{
			CommonNode: CommonNode{
				Name: "node01",
				Host: "10.0.0.1",
				Port: 6379,
			},
		},
		Service: &Service{Type: ServiceTypeClusterIP},
		Replica: RedisNodes{
			&RedisNode{
				CommonNode: CommonNode{
					Name: "node02",
					Host: "10.0.0.2",
					Port: 6379,
				},
			},
		},
	}

	deepCopy := original.DeepCopy()

	if diff := cmp.Diff(original, deepCopy); diff != "" {
		t.Errorf("RedisReplicationSpec.DeepCopy() mismatch (-original +copy):\n%s", diff)
	}
}

func TestRedisReplicationStatus(t *testing.T) {
	original := &RedisReplicationStatus{
		Topology: RedisReplicationTopology{
			"node01": &RedisReplicationNode{
				Host:   "10.0.0.1",
				Port:   6379,
				Role:   RedisReplicationNodeRoleSource,
				Status: NodeStatusOK,
				Ready:  true,
			},
			"node02": &RedisReplicationNode{
				Host:       "10.0.0.2",
				Port:       3306,
				Role:       RedisReplicationNodeRoleReplica,
				Status:     NodeStatusOK,
				Ready:      true,
				SourceHost: "10.0.0.1",
				SourcePort: 6379,
			},
		},
		ReadWriteService: "redisreplication-sample-readwrite",
		ReadOnlyService:  "redisreplication-sample-readonly",
		Ready:            true,
		Conditions: []metav1.Condition{
			{
				Type:    ConditionTypeTopologyReady,
				Status:  metav1.ConditionTrue,
				Message: "Successfully sync topology",
				Reason:  "SyncTopologySucceed",
			},
			{
				Type:    ConditionTypeResourceReady,
				Status:  metav1.ConditionTrue,
				Message: "Successfully sync resource",
				Reason:  "SyncResourceSucceed",
			},
		},
	}

	deepCopy := original.DeepCopy()

	if diff := cmp.Diff(original, deepCopy); diff != "" {
		t.Errorf("RedisReplicationStatus.DeepCopy() mismatch (-original +copy):\n%s", diff)
	}
}

func TestRedisReplicationTopology(t *testing.T) {
	original := RedisReplicationTopology{
		"node01": &RedisReplicationNode{
			Host:   "10.0.0.1",
			Port:   6379,
			Role:   RedisReplicationNodeRoleSource,
			Status: NodeStatusOK,
			Ready:  true,
		},
		"node02": &RedisReplicationNode{
			Host:       "10.0.0.2",
			Port:       3306,
			Role:       RedisReplicationNodeRoleReplica,
			Status:     NodeStatusOK,
			Ready:      true,
			SourceHost: "10.0.0.1",
			SourcePort: 6379,
		},
	}

	deepCopy := original.DeepCopy()

	if diff := cmp.Diff(original, deepCopy); diff != "" {
		t.Errorf("RedisReplicationTopology.DeepCopy() mismatch (-original +copy):\n%s", diff)
	}
}
