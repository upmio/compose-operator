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

func TestMysqlReplication(t *testing.T) {
	original := &MysqlReplication{
		TypeMeta: metav1.TypeMeta{
			Kind:       "MysqlReplication",
			APIVersion: "v1alpha1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      "mysqlreplication-sample",
			Namespace: "test",
		},
		Spec: MysqlReplicationSpec{
			Mode: MysqlRplASync,
			Secret: MysqlReplicationSecret{
				Name:        "mysqlreplication-secret",
				Mysql:       "root",
				Replication: "replication",
			},
			Source: &CommonNode{
				Name: "node01",
				Host: "10.0.0.1",
				Port: 3306,
			},
			Service: &Service{Type: ServiceTypeClusterIP},
			Replica: ReplicaNodes{
				&ReplicaNode{
					CommonNode: CommonNode{
						Name: "node02",
						Host: "10.0.0.2",
						Port: 3306,
					},
				},
			},
		},
		Status: MysqlReplicationStatus{
			Topology: MysqlReplicationTopology{
				"node01": &MysqlReplicationNode{
					Host:          "10.0.0.1",
					Port:          3306,
					Role:          MysqlReplicationNodeRoleSource,
					Status:        NodeStatusOK,
					Ready:         true,
					ReadOnly:      false,
					SuperReadOnly: false,
				},
				"node02": &MysqlReplicationNode{
					Host:             "10.0.0.2",
					Port:             3306,
					Role:             MysqlReplicationNodeRoleReplica,
					Status:           NodeStatusOK,
					Ready:            true,
					ReadOnly:         true,
					SuperReadOnly:    true,
					SourceHost:       "10.0.0.1",
					SourcePort:       3306,
					ReplicaIO:        "yes",
					ReplicaSQL:       "yes",
					ReadSourceLogPos: 0,
					SourceLogFile:    "node01-log",
				},
			},
			ReadWriteService: "mysqlreplication-sample-readwrite",
			ReadOnlyService:  "mysqlreplication-sample-readonly",
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
		t.Errorf("MysqlReplication.DeepCopy() mismatch (-original +copy):\n%s", diff)
	}
}

func TestMysqlReplicationList(t *testing.T) {
	original := &MysqlReplicationList{
		TypeMeta: metav1.TypeMeta{
			Kind:       "MysqlReplicationList",
			APIVersion: "v1alpha1",
		},
		Items: []MysqlReplication{
			{
				TypeMeta: metav1.TypeMeta{
					Kind:       "MysqlReplication",
					APIVersion: "v1alpha1",
				},
				ObjectMeta: metav1.ObjectMeta{
					Name:      "mysqlreplication-sample",
					Namespace: "test",
				},
				Spec: MysqlReplicationSpec{
					Mode: MysqlRplASync,
					Secret: MysqlReplicationSecret{
						Name:        "mysqlreplication-secret",
						Mysql:       "root",
						Replication: "replication",
					},
					Source: &CommonNode{
						Name: "node01",
						Host: "10.0.0.1",
						Port: 3306,
					},
					Service: &Service{Type: ServiceTypeClusterIP},
					Replica: ReplicaNodes{
						&ReplicaNode{
							CommonNode: CommonNode{
								Name: "node02",
								Host: "10.0.0.2",
								Port: 3306,
							},
						},
					},
				},
				Status: MysqlReplicationStatus{
					Topology: MysqlReplicationTopology{
						"node01": &MysqlReplicationNode{
							Host:          "10.0.0.1",
							Port:          3306,
							Role:          MysqlReplicationNodeRoleSource,
							Status:        NodeStatusOK,
							Ready:         true,
							ReadOnly:      false,
							SuperReadOnly: false,
						},
						"node02": &MysqlReplicationNode{
							Host:             "10.0.0.2",
							Port:             3306,
							Role:             MysqlReplicationNodeRoleReplica,
							Status:           NodeStatusOK,
							Ready:            true,
							ReadOnly:         true,
							SuperReadOnly:    true,
							SourceHost:       "10.0.0.1",
							SourcePort:       3306,
							ReplicaIO:        "yes",
							ReplicaSQL:       "yes",
							ReadSourceLogPos: 0,
							SourceLogFile:    "node01-log",
						},
					},
					ReadWriteService: "mysqlreplication-sample-readwrite",
					ReadOnlyService:  "mysqlreplication-sample-readonly",
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
		t.Errorf("MysqlReplicationList.DeepCopy() mismatch (-original +copy):\n%s", diff)
	}
}

func TestMysqlReplicationNode(t *testing.T) {
	original := &MysqlReplicationNode{
		Host:             "127.0.0.1",
		Port:             3306,
		Role:             MysqlReplicationNodeRoleSource,
		Status:           NodeStatusOK,
		Ready:            true,
		ReadOnly:         false,
		SuperReadOnly:    false,
		SourceHost:       "",
		SourcePort:       0,
		ReplicaIO:        "",
		ReplicaSQL:       "",
		ReadSourceLogPos: 0,
		SourceLogFile:    "",
	}

	deepCopy := original.DeepCopy()

	if diff := cmp.Diff(original, deepCopy); diff != "" {
		t.Errorf("MysqlReplicationNode.DeepCopy() mismatch (-original +copy):\n%s", diff)
	}
}

func TestMysqlReplicationSecret(t *testing.T) {
	original := &MysqlReplicationSecret{
		Name:        "mysqlreplication-secret",
		Mysql:       "root",
		Replication: "replication",
	}

	deepCopy := original.DeepCopy()

	if diff := cmp.Diff(original, deepCopy); diff != "" {
		t.Errorf("MysqlReplicationSecret.DeepCopy() mismatch (-original +copy):\n%s", diff)
	}
}

func TestMysqlReplicationSpec(t *testing.T) {
	original := &MysqlReplicationSpec{
		Mode: MysqlRplASync,
		Secret: MysqlReplicationSecret{
			Name:        "mysqlreplication-secret",
			Mysql:       "root",
			Replication: "replication",
		},
		Source: &CommonNode{
			Name: "node01",
			Host: "10.0.0.1",
			Port: 3306,
		},
		Service: &Service{Type: ServiceTypeClusterIP},
		Replica: ReplicaNodes{
			&ReplicaNode{
				CommonNode: CommonNode{
					Name: "node02",
					Host: "10.0.0.2",
					Port: 3306,
				},
			},
		},
	}

	deepCopy := original.DeepCopy()

	if diff := cmp.Diff(original, deepCopy); diff != "" {
		t.Errorf("MysqlReplicationSpec.DeepCopy() mismatch (-original +copy):\n%s", diff)
	}
}

func TestMysqlReplicationStatus(t *testing.T) {
	original := &MysqlReplicationStatus{
		Topology: MysqlReplicationTopology{
			"node01": &MysqlReplicationNode{
				Host:          "10.0.0.1",
				Port:          3306,
				Role:          MysqlReplicationNodeRoleSource,
				Status:        NodeStatusOK,
				Ready:         true,
				ReadOnly:      false,
				SuperReadOnly: false,
			},
			"node02": &MysqlReplicationNode{
				Host:             "10.0.0.2",
				Port:             3306,
				Role:             MysqlReplicationNodeRoleReplica,
				Status:           NodeStatusOK,
				Ready:            true,
				ReadOnly:         true,
				SuperReadOnly:    true,
				SourceHost:       "10.0.0.1",
				SourcePort:       3306,
				ReplicaIO:        "yes",
				ReplicaSQL:       "yes",
				ReadSourceLogPos: 0,
				SourceLogFile:    "node01-log",
			},
		},
		ReadWriteService: "mysqlreplication-sample-readwrite",
		ReadOnlyService:  "mysqlreplication-sample-readonly",
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
		t.Errorf("MysqlReplicationStatus.DeepCopy() mismatch (-original +copy):\n%s", diff)
	}
}

func TestMysqlReplicationTopology(t *testing.T) {
	original := MysqlReplicationTopology{
		"node01": &MysqlReplicationNode{
			Host:          "10.0.0.1",
			Port:          3306,
			Role:          MysqlReplicationNodeRoleSource,
			Status:        NodeStatusOK,
			Ready:         true,
			ReadOnly:      false,
			SuperReadOnly: false,
		},
		"node02": &MysqlReplicationNode{
			Host:             "10.0.0.2",
			Port:             3306,
			Role:             MysqlReplicationNodeRoleReplica,
			Status:           NodeStatusOK,
			Ready:            true,
			ReadOnly:         true,
			SuperReadOnly:    true,
			SourceHost:       "10.0.0.1",
			SourcePort:       3306,
			ReplicaIO:        "yes",
			ReplicaSQL:       "yes",
			ReadSourceLogPos: 0,
			SourceLogFile:    "node01-log",
		},
	}

	deepCopy := original.DeepCopy()

	if diff := cmp.Diff(original, deepCopy); diff != "" {
		t.Errorf("MysqlReplicationTopology.DeepCopy() mismatch (-original +copy):\n%s", diff)
	}
}
