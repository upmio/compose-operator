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

func TestPostgresReplication(t *testing.T) {
	original := &PostgresReplication{
		TypeMeta: metav1.TypeMeta{
			Kind:       "PostgresReplication",
			APIVersion: "v1alpha1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      "postgresreplication-sample",
			Namespace: "test",
		},
		Spec: PostgresReplicationSpec{
			Mode: PostgresRplAsync,
			Secret: PostgresReplicationSecret{
				Name:        "postgresreplication-secret",
				Postgresql:  "postgres",
				Replication: "replication",
			},
			Primary: &CommonNode{
				Name: "node01",
				Host: "10.0.0.1",
				Port: 5432,
			},
			Service: &Service{Type: ServiceTypeClusterIP},
			Standby: ReplicaNodes{
				&ReplicaNode{
					CommonNode: CommonNode{
						Name: "node01",
						Host: "10.0.0.1",
						Port: 5432,
					},
					Isolated: false,
				},
			},
		},
		Status: PostgresReplicationStatus{
			Topology: PostgresReplicationTopology{
				"node01": &PostgresReplicationNode{
					Host:   "10.0.0.1",
					Port:   5432,
					Role:   PostgresReplicationRolePrimary,
					Status: NodeStatusOK,
					Ready:  true,
				},
				"node02": &PostgresReplicationNode{
					Host:   "10.0.0.2",
					Port:   5432,
					Role:   PostgresReplicationRoleStandby,
					Status: NodeStatusOK,
					Ready:  true,
				},
			},
			ReadWriteService: "postgresreplication-sample-readwrite",
			ReadOnlyService:  "postgresreplication-sample-readonly",
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
		t.Errorf("PostgresReplication.DeepCopy() mismatch (-original +copy):\n%s", diff)
	}
}

func TestPostgresReplicationList(t *testing.T) {
	original := &PostgresReplicationList{
		TypeMeta: metav1.TypeMeta{
			Kind:       "PostgresReplicationList",
			APIVersion: "v1alpha1",
		},
		Items: []PostgresReplication{
			{
				TypeMeta: metav1.TypeMeta{
					Kind:       "PostgresReplication",
					APIVersion: "v1alpha1",
				},
				ObjectMeta: metav1.ObjectMeta{
					Name:      "postgresreplication-sample",
					Namespace: "test",
				},
				Spec: PostgresReplicationSpec{
					Mode: PostgresRplAsync,
					Secret: PostgresReplicationSecret{
						Name:        "postgresreplication-secret",
						Postgresql:  "postgres",
						Replication: "replication",
					},
					Primary: &CommonNode{
						Name: "node01",
						Host: "10.0.0.1",
						Port: 5432,
					},
					Service: &Service{Type: ServiceTypeClusterIP},
					Standby: ReplicaNodes{
						&ReplicaNode{
							CommonNode: CommonNode{
								Name: "node02",
								Host: "10.0.0.2",
								Port: 5432,
							},
							Isolated: false,
						},
					},
				},
				Status: PostgresReplicationStatus{
					Topology: PostgresReplicationTopology{
						"node01": &PostgresReplicationNode{
							Host:   "10.0.0.1",
							Port:   5432,
							Role:   PostgresReplicationRolePrimary,
							Status: NodeStatusOK,
							Ready:  true,
						},
						"node02": &PostgresReplicationNode{
							Host:   "10.0.0.2",
							Port:   5432,
							Role:   PostgresReplicationRoleStandby,
							Status: NodeStatusOK,
							Ready:  true,
						},
					},
					ReadWriteService: "postgresreplication-sample-readwrite",
					ReadOnlyService:  "postgresreplication-sample-readonly",
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
		t.Errorf("PostgresReplicationList.DeepCopy() mismatch (-original +copy):\n%s", diff)
	}
}

func TestPostgresReplicationNode(t *testing.T) {
	original := &PostgresReplicationNode{
		Host:   "",
		Port:   0,
		Role:   "",
		Status: "",
		Ready:  false,
	}

	deepCopy := original.DeepCopy()

	if diff := cmp.Diff(original, deepCopy); diff != "" {
		t.Errorf("PostgresReplicationNode.DeepCopy() mismatch (-original +copy):\n%s", diff)
	}
}

func TestPostgresReplicationSecret(t *testing.T) {
	original := &PostgresReplicationSecret{
		Name:        "",
		Postgresql:  "",
		Replication: "",
	}

	deepCopy := original.DeepCopy()

	if diff := cmp.Diff(original, deepCopy); diff != "" {
		t.Errorf("PostgresReplicationSecret.DeepCopy() mismatch (-original +copy):\n%s", diff)
	}
}

func TestPostgresReplicationSpec(t *testing.T) {
	original := &PostgresReplicationSpec{
		Mode: PostgresRplAsync,
		Secret: PostgresReplicationSecret{
			Name:        "postgresreplication-secret",
			Postgresql:  "postgres",
			Replication: "replication",
		},
		Primary: &CommonNode{
			Name: "node01",
			Host: "10.0.0.1",
			Port: 5432,
		},
		Service: &Service{Type: ServiceTypeClusterIP},
		Standby: ReplicaNodes{
			&ReplicaNode{
				CommonNode: CommonNode{
					Name: "node02",
					Host: "10.0.0.2",
					Port: 5432,
				},
				Isolated: false,
			},
		},
	}

	deepCopy := original.DeepCopy()

	if diff := cmp.Diff(original, deepCopy); diff != "" {
		t.Errorf("PostgresReplicationSpec.DeepCopy() mismatch (-original +copy):\n%s", diff)
	}
}

func TestPostgresReplicationStatus(t *testing.T) {
	original := &PostgresReplicationStatus{
		Topology: PostgresReplicationTopology{
			"node01": &PostgresReplicationNode{
				Host:   "10.0.0.1",
				Port:   5432,
				Role:   PostgresReplicationRolePrimary,
				Status: NodeStatusOK,
				Ready:  true,
			},
			"node02": &PostgresReplicationNode{
				Host:   "10.0.0.2",
				Port:   5432,
				Role:   PostgresReplicationRoleStandby,
				Status: NodeStatusOK,
				Ready:  true,
			},
		},
		ReadWriteService: "postgresreplication-sample-readwrite",
		ReadOnlyService:  "postgresreplication-sample-readonly",
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
		t.Errorf("PostgresReplicationStatus.DeepCopy() mismatch (-original +copy):\n%s", diff)
	}
}

func TestPostgresReplicationTopology(t *testing.T) {
	original := PostgresReplicationTopology{
		"node01": &PostgresReplicationNode{
			Host:   "10.0.0.1",
			Port:   5432,
			Role:   PostgresReplicationRolePrimary,
			Status: NodeStatusOK,
			Ready:  true,
		},
		"node02": &PostgresReplicationNode{
			Host:   "10.0.0.2",
			Port:   5432,
			Role:   PostgresReplicationRoleStandby,
			Status: NodeStatusOK,
			Ready:  true,
		},
	}

	deepCopy := original.DeepCopy()

	if diff := cmp.Diff(original, deepCopy); diff != "" {
		t.Errorf("PostgresReplicationTopology.DeepCopy() mismatch (-original +copy):\n%s", diff)
	}
}
