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

package mysqlgroupreplication

import (
	"context"
	"net"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	"strconv"

	"github.com/go-logr/logr"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"

	composev1alpha1 "github.com/upmio/compose-operator/api/v1alpha1"
	"github.com/upmio/compose-operator/pkg/k8sutil"
	"github.com/upmio/compose-operator/pkg/mysqlutil"
)

// newMysqlAdmin builds and returns new mysql.ReplicationAdmin from the list of mysql address
func newMysqlAdmin(instance *composev1alpha1.MysqlGroupReplication, password string, reqLogger logr.Logger) mysqlutil.IGroupAdmin {
	nodesAddrs := make([]string, 0)

	for _, node := range instance.Spec.Member {
		nodesAddrs = append(nodesAddrs, net.JoinHostPort(node.Host, strconv.Itoa(node.Port)))
		reqLogger.V(4).Info("append mysql admin node addr", "host", node.Host, "port", node.Port)
	}

	adminConfig := mysqlutil.AdminOptions{
		ConnectionTimeout: 2,
		Username:          instance.Spec.Secret.Mysql,
		Password:          password,
	}

	return mysqlutil.NewGroupAdmin(nodesAddrs, &adminConfig, reqLogger)
}

// decryptSecret returns the current mysql password and replication password.
func decryptSecret(client client.Client, reqLogger logr.Logger, instance *composev1alpha1.MysqlGroupReplication) (string, string, error) {
	passwords, err := k8sutil.DecryptSecretPasswords(
		client,
		instance.Spec.Secret.Name,
		instance.Namespace,
		instance.Spec.AESSecret.Name,
		instance.Spec.AESSecret.Key,
		[]string{instance.Spec.Secret.Mysql, instance.Spec.Secret.Replication},
	)
	if err != nil {
		return "", "", err
	}
	return passwords[instance.Spec.Secret.Mysql], passwords[instance.Spec.Secret.Replication], nil
}

func podMapFunc(_ context.Context, o client.Object) []reconcile.Request {
	pod := o.(*corev1.Pod)

	// Get MysqlGroupReplication's name from pod labels
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
