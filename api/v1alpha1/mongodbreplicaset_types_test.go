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

func TestMongoDBReplicaset(t *testing.T) {
	instance := &MongoDBReplicaset{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-mongodb-replicaset",
			Namespace: "default",
		},
		Spec: MongoDBReplicasetSpec{
			Secret: MongoDBReplicasetSecret{
				Name:   "mongodb-secret",
				Mongod: "mongod",
			},
			Member: CommonNodes{
				{
					Name: "mongodb-0",
					Host: "mongodb-0.mongodb.default.svc.cluster.local",
					Port: 27017,
				},
				{
					Name: "mongodb-1",
					Host: "mongodb-1.mongodb.default.svc.cluster.local",
					Port: 27017,
				},
				{
					Name: "mongodb-2",
					Host: "mongodb-2.mongodb.default.svc.cluster.local",
					Port: 27017,
				},
			},
			ReplicaSetName: "rs0",
		},
	}

	if instance.Spec.Secret.Name != "mongodb-secret" {
		t.Errorf("Expected secret name to be 'mongodb-secret', got '%s'", instance.Spec.Secret.Name)
	}

	if instance.Spec.ReplicaSetName != "rs0" {
		t.Errorf("Expected replica set name to be 'rs0', got '%s'", instance.Spec.ReplicaSetName)
	}

	if len(instance.Spec.Member) != 3 {
		t.Errorf("Expected 3 members, got %d", len(instance.Spec.Member))
	}
}

func TestMongoDBReplicasetStatus(t *testing.T) {
	status := MongoDBReplicasetStatus{
		Ready: true,
		Topology: MongoDBReplicasetTopology{
			"mongodb-0": &MongoDBReplicasetNode{
				Host:   "mongodb-0.mongodb.default.svc.cluster.local",
				Port:   27017,
				Role:   MongoDBReplicasetNodeRolePrimary,
				Status: NodeStatusOK,
				State:  "PRIMARY",
			},
			"mongodb-1": &MongoDBReplicasetNode{
				Host:   "mongodb-1.mongodb.default.svc.cluster.local",
				Port:   27017,
				Role:   MongoDBReplicasetNodeRoleSecondary,
				Status: NodeStatusOK,
				State:  "SECONDARY",
			},
		},
	}

	if !status.Ready {
		t.Error("Expected status to be ready")
	}

	if len(status.Topology) != 2 {
		t.Errorf("Expected 2 nodes in topology, got %d", len(status.Topology))
	}

	primaryNode := status.Topology["mongodb-0"]
	if primaryNode.Role != MongoDBReplicasetNodeRolePrimary {
		t.Errorf("Expected node role to be 'Primary', got '%s'", primaryNode.Role)
	}

	if primaryNode.State != "PRIMARY" {
		t.Errorf("Expected node state to be 'PRIMARY', got '%s'", primaryNode.State)
	}
}
