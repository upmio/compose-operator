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

package mongodbreplicaset

import (
	"context"
	"testing"
	"time"

	"github.com/go-logr/logr"
	"github.com/stretchr/testify/assert"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/tools/record"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	composev1alpha1 "github.com/upmio/compose-operator/api/v1alpha1"
)

func TestReconcileMongoDBReplicaset_ValidateInstance(t *testing.T) {
	// Create a fake client
	scheme := runtime.NewScheme()
	err := composev1alpha1.AddToScheme(scheme)
	assert.NoError(t, err)

	instance := &composev1alpha1.MongoDBReplicaset{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-mongodb-replicaset",
			Namespace: "default",
		},
		Spec: composev1alpha1.MongoDBReplicasetSpec{
			Secret: composev1alpha1.MongoDBReplicasetSecret{
				Name:   "mongodb-secret",
				Mongod: "mongod",
			},
			AESSecret: &composev1alpha1.AESSecret{
				Name: "AESSecret",
				Key:  "AESKey",
			},
			Member: composev1alpha1.CommonNodes{
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

	client := fake.NewClientBuilder().WithScheme(scheme).WithObjects(instance).Build()

	r := &ReconcileMongoDBReplicaset{
		client:   client,
		scheme:   scheme,
		recorder: &record.FakeRecorder{},
		logger:   logr.Discard(),
	}

	req := reconcile.Request{
		NamespacedName: types.NamespacedName{
			Name:      "test-mongodb-replicaset",
			Namespace: "default",
		},
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	result, err := r.Reconcile(ctx, req)

	// Should not return error for valid instance, but will requeue due to missing secret
	assert.NoError(t, err)
	assert.Equal(t, requeueAfter, result.RequeueAfter)

	// Verify the instance status was updated with failed condition due to missing secret
	updatedInstance := &composev1alpha1.MongoDBReplicaset{}
	err = client.Get(ctx, req.NamespacedName, updatedInstance)
	assert.NoError(t, err)
	assert.False(t, updatedInstance.Status.Ready)
}

func TestBuildDefaultTopologyStatus(t *testing.T) {
	instance := &composev1alpha1.MongoDBReplicaset{
		Spec: composev1alpha1.MongoDBReplicasetSpec{
			Member: composev1alpha1.CommonNodes{
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
			},
		},
	}

	status := buildDefaultTopologyStatus(instance)

	assert.Len(t, status.Topology, 2)
	assert.False(t, status.Ready)

	node0 := status.Topology["mongodb-0"]
	assert.Equal(t, "mongodb-0.mongodb.default.svc.cluster.local", node0.Host)
	assert.Equal(t, 27017, node0.Port)
	assert.Equal(t, composev1alpha1.MongoDBReplicasetNodeRoleNone, node0.Role)
	assert.Equal(t, composev1alpha1.NodeStatusKO, node0.Status)
}

func TestCompareNodes(t *testing.T) {
	nodeA := &composev1alpha1.MongoDBReplicasetNode{
		Host:   "mongodb-0.mongodb.default.svc.cluster.local",
		Port:   27017,
		Role:   composev1alpha1.MongoDBReplicasetNodeRolePrimary,
		Status: composev1alpha1.NodeStatusOK,
		State:  "PRIMARY",
	}

	nodeB := &composev1alpha1.MongoDBReplicasetNode{
		Host:   "mongodb-0.mongodb.default.svc.cluster.local",
		Port:   27017,
		Role:   composev1alpha1.MongoDBReplicasetNodeRolePrimary,
		Status: composev1alpha1.NodeStatusOK,
		State:  "PRIMARY",
	}

	// Same nodes should not show changes
	logger := logr.Discard()
	changed := compareNodes(nodeA, nodeB, logger)
	assert.False(t, changed)

	// Different role should show changes
	nodeB.Role = composev1alpha1.MongoDBReplicasetNodeRoleSecondary
	changed = compareNodes(nodeA, nodeB, logger)
	assert.True(t, changed)
}
