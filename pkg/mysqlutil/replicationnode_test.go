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

package mysqlutil

import (
	"testing"

	composev1alpha1 "github.com/upmio/compose-operator/api/v1alpha1"
)

func TestGetRole(t *testing.T) {
	node := NewDefaultReplicationNode()
	if node.GetRole() != composev1alpha1.MysqlReplicationNodeRoleNone {
		t.Fatal("node.Role None test failed")
	}

	node.Role = "Source"
	if node.GetRole() != composev1alpha1.MysqlReplicationNodeRoleSource {
		t.Fatal("node.Role Source test failed")
	}

	node.Role = "Replica"
	if node.GetRole() != composev1alpha1.MysqlReplicationNodeRoleReplica {
		t.Fatal("node.Role replica test failed")
	}
}

func TestGetSourcePort(t *testing.T) {
	node := NewDefaultReplicationNode()

	node.SourcePort = "3306"
	if node.GetSourcePort() != 3306 {
		t.Fatal("node.SourcePort GetSourcePort test failed")
	}
}

func TestGetReadSourceLogPos(t *testing.T) {
	node := NewDefaultReplicationNode()

	node.ReadSourceLogPos = "5413"
	if node.GetReadSourceLogPos() != 5413 {
		t.Fatal("node.ReadSourceLogPos GetReadSourceLogPos test failed")
	}
}
