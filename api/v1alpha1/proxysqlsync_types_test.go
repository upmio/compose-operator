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

func TestProxysqlSync(t *testing.T) {
	original := &ProxysqlSync{
		TypeMeta: metav1.TypeMeta{
			Kind:       "ProxysqlSync",
			APIVersion: "v1alpha1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      "proxysqlsync-sample",
			Namespace: "test",
		},
		Spec: ProxysqlSyncSpec{
			Proxysql: CommonNodes{
				{
					Name: "node01",
					Host: "10.0.0.1",
					Port: 6033,
				},
			},
			Secret: ProxysqlSyncSecret{
				Name:     "proxysqlsync-secret",
				Proxysql: "proxysql",
				Mysql:    "root",
			},
			MysqlReplication: "mysqlreplication-sample",
			Rule: &Rule{
				Filter: []string{
					"%",
				},
				Pattern: "10.0.0.1",
			},
		},
		Status: ProxysqlSyncStatus{
			Topology: ProxysqlSyncTopology{
				"node01": &ProxysqlSyncNode{
					Users: []string{
						"user01", "user02",
					},
					Synced: true,
				},
			},
			Ready: true,
			Conditions: []metav1.Condition{
				{
					Type:    ConditionTypeUserReady,
					Status:  metav1.ConditionTrue,
					Message: "Successfully sync user",
					Reason:  "SyncUserSucceed",
				},
				{
					Type:    ConditionTypeServerReady,
					Status:  metav1.ConditionTrue,
					Message: "Successfully sync server",
					Reason:  "SyncServerSucceed",
				},
			},
		},
	}

	deepCopy := original.DeepCopyObject()

	if diff := cmp.Diff(original, deepCopy); diff != "" {
		t.Errorf("ProxysqlSync.DeepCopy() mismatch (-original +copy):\n%s", diff)
	}
}

func TestProxysqlSyncList(t *testing.T) {
	original := &ProxysqlSyncList{
		TypeMeta: metav1.TypeMeta{
			Kind:       "ProxysqlSyncList",
			APIVersion: "v1alpha1",
		},
		ListMeta: metav1.ListMeta{
			// Initialize fields with test data
		},
		Items: []ProxysqlSync{
			{
				TypeMeta: metav1.TypeMeta{
					Kind:       "ProxysqlSync",
					APIVersion: "v1alpha1",
				},
				ObjectMeta: metav1.ObjectMeta{
					Name:      "proxysqlsync-sample",
					Namespace: "test",
				},
				Spec: ProxysqlSyncSpec{
					Proxysql: CommonNodes{
						{
							Name: "node01",
							Host: "10.0.0.1",
							Port: 6033,
						},
					},
					Secret: ProxysqlSyncSecret{
						Name:     "proxysqlsync-secret",
						Proxysql: "proxysql",
						Mysql:    "root",
					},
					MysqlReplication: "mysqlreplication-sample",
					Rule: &Rule{
						Filter: []string{
							"%",
						},
						Pattern: "10.0.0.1",
					},
				},
				Status: ProxysqlSyncStatus{
					Topology: ProxysqlSyncTopology{
						"node01": &ProxysqlSyncNode{
							Users: []string{
								"user01", "user02",
							},
							Synced: true,
						},
					},
					Ready: true,
					Conditions: []metav1.Condition{
						{
							Type:    ConditionTypeUserReady,
							Status:  metav1.ConditionTrue,
							Message: "Successfully sync user",
							Reason:  "SyncUserSucceed",
						},
						{
							Type:    ConditionTypeServerReady,
							Status:  metav1.ConditionTrue,
							Message: "Successfully sync server",
							Reason:  "SyncServerSucceed",
						},
					},
				},
			},
		},
	}
	deepCopy := original.DeepCopyObject()

	if diff := cmp.Diff(original, deepCopy); diff != "" {
		t.Errorf("ProxysqlSyncList.DeepCopy() mismatch (-original +copy):\n%s", diff)
	}
}

func TestProxysqlSyncNode(t *testing.T) {
	original := &ProxysqlSyncNode{
		Users: []string{
			"user01", "user02",
		},
		Synced: true,
	}

	deepCopy := original.DeepCopy()

	if diff := cmp.Diff(original, deepCopy); diff != "" {
		t.Errorf("ProxysqlSyncNode.DeepCopy() mismatch (-original +copy):\n%s", diff)
	}
}

func TestProxysqlSyncSecret(t *testing.T) {
	original := &ProxysqlSyncSecret{
		Name:     "proxysqlsync-secret",
		Proxysql: "proxysql",
		Mysql:    "root",
	}

	deepCopy := original.DeepCopy()

	if diff := cmp.Diff(original, deepCopy); diff != "" {
		t.Errorf("ProxysqlSyncSecret.DeepCopy() mismatch (-original +copy):\n%s", diff)
	}
}

func TestProxysqlSyncSpec(t *testing.T) {
	original := &ProxysqlSyncSpec{
		Proxysql: CommonNodes{
			{
				Name: "node01",
				Host: "10.0.0.1",
				Port: 6033,
			},
		},
		Secret: ProxysqlSyncSecret{
			Name:     "proxysqlsync-secret",
			Proxysql: "proxysql",
			Mysql:    "root",
		},
		MysqlReplication: "mysqlreplication-sample",
		Rule: &Rule{
			Filter: []string{
				"%",
			},
			Pattern: "10.0.0.1",
		},
	}

	deepCopy := original.DeepCopy()

	if diff := cmp.Diff(original, deepCopy); diff != "" {
		t.Errorf("ProxysqlSyncSpec.DeepCopy() mismatch (-original +copy):\n%s", diff)
	}
}

func TestProxysqlSyncStatus(t *testing.T) {
	original := &ProxysqlSyncStatus{
		Topology: ProxysqlSyncTopology{
			"node01": &ProxysqlSyncNode{
				Users: []string{
					"user01", "user02",
				},
				Synced: true,
			},
		},
		Ready: true,
		Conditions: []metav1.Condition{
			{
				Type:    ConditionTypeUserReady,
				Status:  metav1.ConditionTrue,
				Message: "Successfully sync user",
				Reason:  "SyncUserSucceed",
			},
			{
				Type:    ConditionTypeServerReady,
				Status:  metav1.ConditionTrue,
				Message: "Successfully sync server",
				Reason:  "SyncServerSucceed",
			},
		},
	}

	deepCopy := original.DeepCopy()

	if diff := cmp.Diff(original, deepCopy); diff != "" {
		t.Errorf("ProxysqlSyncStatus.DeepCopy() mismatch (-original +copy):\n%s", diff)
	}
}

func TestProxysqlSyncTopology(t *testing.T) {
	original := ProxysqlSyncTopology{
		"node01": &ProxysqlSyncNode{
			Users: []string{
				"user01", "user02",
			},
			Synced: true,
		},
	}

	deepCopy := original.DeepCopy()

	if diff := cmp.Diff(original, deepCopy); diff != "" {
		t.Errorf("ProxysqlSyncTopology.DeepCopy() mismatch (-original +copy):\n%s", diff)
	}
}
