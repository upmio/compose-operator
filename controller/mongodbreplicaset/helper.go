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
	"net"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	"strconv"
	"time"

	"github.com/go-logr/logr"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"

	composev1alpha1 "github.com/upmio/compose-operator/api/v1alpha1"
	"github.com/upmio/compose-operator/pkg/k8sutil"
	"github.com/upmio/compose-operator/pkg/mongoutil"
)

// newMongoAdmin builds and returns new mongoutil.IReplicaSetAdmin from the list of mongodb address
func newMongoAdmin(instance *composev1alpha1.MongoDBReplicaset, password string, reqLogger logr.Logger) mongoutil.IReplicaSetAdmin {
	nodesAddrs := make([]string, 0)

	for _, node := range instance.Spec.Member {
		nodesAddrs = append(nodesAddrs, net.JoinHostPort(node.Host, strconv.Itoa(node.Port)))
		reqLogger.V(4).Info("append mongodb admin node addr", "host", node.Host, "port", node.Port)
	}

	adminConfig := mongoutil.AdminOptions{
		ConnectionTimeout: 2 * time.Second,
		Username:          instance.Spec.Secret.Mongod,
		Password:          password,
		AuthDatabase:      "admin",
	}

	return mongoutil.NewReplicaSetAdmin(nodesAddrs, &adminConfig, reqLogger)
}

// decryptSecret returns the current mongodb password.
func decryptSecret(client client.Client, reqLogger logr.Logger, instance *composev1alpha1.MongoDBReplicaset) (string, error) {
	password, err := k8sutil.DecryptSecretPasswords(
		client,
		reqLogger,
		instance.Spec.Secret.Name,
		instance.Namespace,
		[]string{instance.Spec.Secret.Mongod},
	)
	if err != nil {
		return "", err
	}
	return password[instance.Spec.Secret.Mongod], nil
}

func podMapFunc(_ context.Context, o client.Object) []reconcile.Request {
	pod := o.(*corev1.Pod)

	// Get MongoDBReplicaset's name from pod labels
	name, exists := pod.Labels[defaultKey]
	if !exists {
		return nil
	}

	// return reconcile request
	return []reconcile.Request{
		{
			NamespacedName: types.NamespacedName{
				Namespace: pod.Namespace,
				Name:      name,
			},
		},
	}
}
