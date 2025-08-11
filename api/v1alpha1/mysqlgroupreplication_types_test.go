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

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestMysqlGroupReplicationDeepCopy(t *testing.T) {
	original := &MysqlGroupReplication{
		TypeMeta: metav1.TypeMeta{
			Kind:       "MysqlGroupReplication",
			APIVersion: "upm.syntropycloud.io/v1alpha1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-mgr",
			Namespace: "default",
		},
		Spec: MysqlGroupReplicationSpec{
			Secret: MysqlGroupReplicationSecret{
				Name:        "test-secret",
				Mysql:       "mysql",
				Replication: "replication",
			},
			Member: []*CommonNode{
				{
					Name: "mysql-0",
					Host: "mysql-0.mysql.default.svc.cluster.local",
					Port: 3306,
				},
				{
					Name: "mysql-1",
					Host: "mysql-1.mysql.default.svc.cluster.local",
					Port: 3306,
				},
			},
		},
		Status: MysqlGroupReplicationStatus{
			Ready: true,
		},
	}

	// Test DeepCopy
	copied := original.DeepCopy()
	if copied == nil {
		t.Error("DeepCopy() returned nil")
		return
	}
	if copied == original {
		t.Error("DeepCopy() returned same instance")
	}
	if copied.Name != original.Name {
		t.Errorf("Name not copied correctly: expected %s, got %s", original.Name, copied.Name)
	}

	// Test DeepCopyObject
	obj := original.DeepCopyObject()
	if obj == nil {
		t.Error("DeepCopyObject() returned nil")
	}
	mgr, ok := obj.(*MysqlGroupReplication)
	if !ok {
		t.Error("DeepCopyObject() returned wrong type")
	}
	if mgr.Name != original.Name {
		t.Errorf("DeepCopyObject() name not copied correctly: expected %s, got %s", original.Name, mgr.Name)
	}
}

func TestMysqlGroupReplicationListDeepCopy(t *testing.T) {
	original := &MysqlGroupReplicationList{
		TypeMeta: metav1.TypeMeta{
			Kind:       "MysqlGroupReplicationList",
			APIVersion: "upm.syntropycloud.io/v1alpha1",
		},
		ListMeta: metav1.ListMeta{
			ResourceVersion: "1",
		},
		Items: []MysqlGroupReplication{
			{
				ObjectMeta: metav1.ObjectMeta{
					Name: "test-mgr-1",
				},
			},
			{
				ObjectMeta: metav1.ObjectMeta{
					Name: "test-mgr-2",
				},
			},
		},
	}

	// Test DeepCopy
	copied := original.DeepCopy()

	if len(copied.Items) != len(original.Items) {
		t.Errorf("Items length not copied correctly: expected %d, got %d", len(original.Items), len(copied.Items))
	}

	// Test DeepCopyObject
	obj := original.DeepCopyObject()
	if obj == nil {
		t.Error("DeepCopyObject() returned nil")
	}
	list, ok := obj.(*MysqlGroupReplicationList)
	if !ok {
		t.Error("DeepCopyObject() returned wrong type")
	}
	if len(list.Items) != len(original.Items) {
		t.Errorf("DeepCopyObject() items length not copied correctly: expected %d, got %d", len(original.Items), len(list.Items))
	}
}

func TestMysqlGroupReplicationSecretDeepCopy(t *testing.T) {
	original := &MysqlGroupReplicationSecret{
		Name:        "test-secret",
		Mysql:       "mysql",
		Replication: "replication",
	}

	// Test DeepCopy
	copied := original.DeepCopy()

	if copied == original {
		t.Error("DeepCopy() returned same instance")
	}
	if copied.Name != original.Name {
		t.Errorf("Name not copied correctly: expected %s, got %s", original.Name, copied.Name)
	}
	if copied.Mysql != original.Mysql {
		t.Errorf("Mysql not copied correctly: expected %s, got %s", original.Mysql, copied.Mysql)
	}
}

func TestMysqlGroupReplicationSpecDeepCopy(t *testing.T) {
	original := &MysqlGroupReplicationSpec{
		Secret: MysqlGroupReplicationSecret{
			Name:        "test-secret",
			Mysql:       "mysql",
			Replication: "replication",
		},
		Member: []*CommonNode{
			{
				Name: "mysql-0",
				Host: "mysql-0.mysql.default.svc.cluster.local",
				Port: 3306,
			},
		},
	}

	// Test DeepCopy
	copied := original.DeepCopy()

	if copied == original {
		t.Error("DeepCopy() returned same instance")
	}
	if len(copied.Member) != len(original.Member) {
		t.Errorf("Member length not copied correctly: expected %d, got %d", len(original.Member), len(copied.Member))
	}
}

func TestMysqlGroupReplicationStatusDeepCopy(t *testing.T) {
	original := &MysqlGroupReplicationStatus{
		Ready: true,
		Topology: MysqlGroupReplicationTopology{
			"mysql-0": &MysqlGroupReplicationNode{
				Host: "mysql-0.mysql.default.svc.cluster.local",
				Port: 3306,
			},
		},
	}

	// Test DeepCopy
	copied := original.DeepCopy()

	if copied == original {
		t.Error("DeepCopy() returned same instance")
	}
	if copied.Ready != original.Ready {
		t.Errorf("Ready not copied correctly: expected %t, got %t", original.Ready, copied.Ready)
	}
}
