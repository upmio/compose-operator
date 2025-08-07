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

func TestRedisCluster(t *testing.T) {
	original := &RedisCluster{
		TypeMeta: metav1.TypeMeta{
			Kind:       "RedisCluster",
			APIVersion: "v1alpha1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test",
			Namespace: "default",
		},
		Spec: RedisClusterSpec{
			Secret: RedisClusterSecret{
				Name:  "rediscluster-secret",
				Redis: "root",
			},
			Members: map[string]CommonNodes{
				"shard01": {
					&CommonNode{
						Name: "node01",
						Host: "10.0.0.1",
						Port: 6379,
					},
				},
				"shard02": {
					&CommonNode{
						Name: "node02",
						Host: "10.0.0.2",
						Port: 6379,
					},
				},
				"shard03": {
					&CommonNode{
						Name: "node03",
						Host: "10.0.0.3",
						Port: 6379,
					},
				},
			},
		},
		Status: RedisClusterStatus{
			Topology: RedisClusterTopology{
				"node01": &RedisClusterNode{
					ID:        "node01",
					Role:      RedisClusterNodeRoleSource,
					Status:    NodeStatusOK,
					Host:      "10.0.0.1",
					Port:      6379,
					Slots:     []string{"0", "1"},
					MasterRef: "node01",
					Shard:     "shard01",
					Ready:     true,
				},
				"node02": &RedisClusterNode{
					ID:        "node02",
					Role:      RedisClusterNodeRoleSource,
					Status:    NodeStatusOK,
					Host:      "10.0.0.2",
					Port:      6379,
					Slots:     []string{"2", "3"},
					MasterRef: "node02",
					Shard:     "shard02",
					Ready:     true,
				},
				"node03": &RedisClusterNode{
					ID:        "node03",
					Role:      RedisClusterNodeRoleSource,
					Status:    NodeStatusOK,
					Host:      "10.0.0.3",
					Port:      6379,
					Slots:     []string{"4", "5"},
					MasterRef: "node03",
					Shard:     "shard03",
					Ready:     true,
				},
			},
			NumberOfShard: 3,
			ClusterJoined: true,
			ClusterSynced: true,
			Ready:         true,
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
		t.Errorf("RedisCluster.DeepCopy() mismatch (-original +copy):\n%s", diff)
	}
}

func TestRedisClusterList(t *testing.T) {
	original := &RedisClusterList{
		TypeMeta: metav1.TypeMeta{
			Kind:       "RedisClusterList",
			APIVersion: "v1alpha1",
		},
		Items: []RedisCluster{
			{
				TypeMeta: metav1.TypeMeta{
					Kind:       "RedisCluster",
					APIVersion: "v1alpha1",
				},
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test",
					Namespace: "default",
				},
				Spec: RedisClusterSpec{
					Secret: RedisClusterSecret{
						Name:  "rediscluster-secret",
						Redis: "root",
					},
					Members: map[string]CommonNodes{
						"shard01": {
							&CommonNode{
								Name: "node01",
								Host: "10.0.0.1",
								Port: 6379,
							},
						},
						"shard02": {
							&CommonNode{
								Name: "node02",
								Host: "10.0.0.2",
								Port: 6379,
							},
						},
						"shard03": {
							&CommonNode{
								Name: "node03",
								Host: "10.0.0.3",
								Port: 6379,
							},
						},
					},
				},
				Status: RedisClusterStatus{
					Topology: RedisClusterTopology{
						"node01": &RedisClusterNode{
							ID:        "node01",
							Role:      RedisClusterNodeRoleSource,
							Status:    NodeStatusOK,
							Host:      "10.0.0.1",
							Port:      6379,
							Slots:     []string{"0", "1"},
							MasterRef: "node01",
							Shard:     "shard01",
							Ready:     true,
						},
						"node02": &RedisClusterNode{
							ID:        "node02",
							Role:      RedisClusterNodeRoleSource,
							Status:    NodeStatusOK,
							Host:      "10.0.0.2",
							Port:      6379,
							Slots:     []string{"2", "3"},
							MasterRef: "node02",
							Shard:     "shard02",
							Ready:     true,
						},
						"node03": &RedisClusterNode{
							ID:        "node03",
							Role:      RedisClusterNodeRoleSource,
							Status:    NodeStatusOK,
							Host:      "10.0.0.3",
							Port:      6379,
							Slots:     []string{"4", "5"},
							MasterRef: "node03",
							Shard:     "shard03",
							Ready:     true,
						},
					},
					NumberOfShard: 3,
					ClusterJoined: true,
					ClusterSynced: true,
					Ready:         true,
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
		t.Errorf("RedisClusterList.DeepCopy() mismatch (-original +copy):\n%s", diff)
	}
}

func TestRedisClusterNode(t *testing.T) {
	original := &RedisClusterNode{
		ID:        "node01",
		Role:      RedisClusterNodeRoleSource,
		Status:    NodeStatusOK,
		Host:      "10.0.0.1",
		Port:      6379,
		Slots:     []string{"0", "1"},
		MasterRef: "node01",
		Shard:     "shard01",
		Ready:     true,
	}

	deepCopy := original.DeepCopy()

	if diff := cmp.Diff(original, deepCopy); diff != "" {
		t.Errorf("RedisClusterNode.DeepCopy() mismatch (-original +copy):\n%s", diff)
	}
}

func TestRedisClusterSecret(t *testing.T) {
	original := &RedisClusterSecret{
		Name:  "rediscluster-secret",
		Redis: "root",
	}

	deepCopy := original.DeepCopy()

	if diff := cmp.Diff(original, deepCopy); diff != "" {
		t.Errorf("RedisClusterSecret.DeepCopy() mismatch (-original +copy):\n%s", diff)
	}
}

func TestRedisClusterSpec(t *testing.T) {
	original := &RedisClusterSpec{
		Secret: RedisClusterSecret{
			Name:  "rediscluster-secret",
			Redis: "root",
		},
		Members: map[string]CommonNodes{
			"shard01": {
				&CommonNode{
					Name: "node01",
					Host: "10.0.0.1",
					Port: 6379,
				},
			},
			"shard02": {
				&CommonNode{
					Name: "node02",
					Host: "10.0.0.2",
					Port: 6379,
				},
			},
			"shard03": {
				&CommonNode{
					Name: "node03",
					Host: "10.0.0.3",
					Port: 6379,
				},
			},
		},
	}

	deepCopy := original.DeepCopy()

	if diff := cmp.Diff(original, deepCopy); diff != "" {
		t.Errorf("RedisClusterSpec.DeepCopy() mismatch (-original +copy):\n%s", diff)
	}
}

func TestRedisClusterStatus(t *testing.T) {
	original := &RedisClusterStatus{
		Topology: RedisClusterTopology{
			"node01": &RedisClusterNode{
				ID:        "node01",
				Role:      RedisClusterNodeRoleSource,
				Status:    NodeStatusOK,
				Host:      "10.0.0.1",
				Port:      6379,
				Slots:     []string{"0", "1"},
				MasterRef: "node01",
				Shard:     "shard01",
				Ready:     true,
			},
			"node02": &RedisClusterNode{
				ID:        "node02",
				Role:      RedisClusterNodeRoleSource,
				Status:    NodeStatusOK,
				Host:      "10.0.0.2",
				Port:      6379,
				Slots:     []string{"2", "3"},
				MasterRef: "node02",
				Shard:     "shard02",
				Ready:     true,
			},
			"node03": &RedisClusterNode{
				ID:        "node03",
				Role:      RedisClusterNodeRoleSource,
				Status:    NodeStatusOK,
				Host:      "10.0.0.3",
				Port:      6379,
				Slots:     []string{"4", "5"},
				MasterRef: "node03",
				Shard:     "shard03",
				Ready:     true,
			},
		},
		NumberOfShard: 3,
		ClusterJoined: true,
		ClusterSynced: true,
		Ready:         true,
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
		t.Errorf("RedisClusterStatus.DeepCopy() mismatch (-original +copy):\n%s", diff)
	}
}

func TestRedisClusterTopology(t *testing.T) {
	original := RedisClusterTopology{
		"node01": &RedisClusterNode{
			ID:        "node01",
			Role:      RedisClusterNodeRoleSource,
			Status:    NodeStatusOK,
			Host:      "10.0.0.1",
			Port:      6379,
			Slots:     []string{"0", "1"},
			MasterRef: "node01",
			Shard:     "shard01",
			Ready:     true,
		},
		"node02": &RedisClusterNode{
			ID:        "node02",
			Role:      RedisClusterNodeRoleSource,
			Status:    NodeStatusOK,
			Host:      "10.0.0.2",
			Port:      6379,
			Slots:     []string{"2", "3"},
			MasterRef: "node02",
			Shard:     "shard02",
			Ready:     true,
		},
		"node03": &RedisClusterNode{
			ID:        "node03",
			Role:      RedisClusterNodeRoleSource,
			Status:    NodeStatusOK,
			Host:      "10.0.0.3",
			Port:      6379,
			Slots:     []string{"4", "5"},
			MasterRef: "node03",
			Shard:     "shard03",
			Ready:     true,
		},
	}

	deepCopy := original.DeepCopy()

	if diff := cmp.Diff(original, deepCopy); diff != "" {
		t.Errorf("RedisClusterTopology.DeepCopy() mismatch (-original +copy):\n%s", diff)
	}
}
