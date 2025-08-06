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

package mysqlreplication

import (
	"context"
	"fmt"
	"net"
	"reflect"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	"strconv"

	"github.com/go-logr/logr"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"

	composev1alpha1 "github.com/upmio/compose-operator/api/v1alpha1"
	"github.com/upmio/compose-operator/pkg/mysqlutil"
	"github.com/upmio/compose-operator/pkg/utils"
)

// newMysqlAdmin builds and returns new mysql.ReplicationAdmin from the list of mysql address
func newMysqlAdmin(instance *composev1alpha1.MysqlReplication, password string, reqLogger logr.Logger) mysqlutil.IReplicationAdmin {
	nodesAddrs := []string{
		net.JoinHostPort(instance.Spec.Source.Host, strconv.Itoa(instance.Spec.Source.Port)),
	}
	reqLogger.V(4).Info("append mysql admin source node addr", "host", instance.Spec.Source.Host, "port", instance.Spec.Source.Port)

	for _, replica := range instance.Spec.Replica {
		nodesAddrs = append(nodesAddrs, net.JoinHostPort(replica.Host, strconv.Itoa(replica.Port)))
		reqLogger.V(4).Info("append mysql admin replica node addr", "host", replica.Host, "port", replica.Port)
	}

	adminConfig := mysqlutil.AdminOptions{
		ConnectionTimeout: 2,
		Username:          instance.Spec.Secret.Mysql,
		Password:          password,
	}

	return mysqlutil.NewReplicationAdmin(nodesAddrs, &adminConfig, reqLogger)
}

// decryptSecret return current mysql password & replication password.
func decryptSecret(client client.Client, instance *composev1alpha1.MysqlReplication) (string, string, error) {
	secret := &corev1.Secret{}
	err := client.Get(context.TODO(), types.NamespacedName{
		Name:      instance.Spec.Secret.Name,
		Namespace: instance.Namespace,
	}, secret)

	if err != nil {
		return "", "", fmt.Errorf("failed to fetch secret [%s]: %v", instance.Spec.Secret, err)
	}

	mysqlPassword, err := utils.AES_CTR_Decrypt(secret.Data[instance.Spec.Secret.Mysql])
	if err != nil {
		return "", "", fmt.Errorf("failed to decrypt secret [%s] key '%s': %v", instance.Spec.Secret, instance.Spec.Secret.Mysql, err)
	}

	replicationPassword, err := utils.AES_CTR_Decrypt(secret.Data[instance.Spec.Secret.Replication])
	if err != nil {
		return "", "", fmt.Errorf("failed to decrypt secret [%s] key '%s': %v", instance.Spec.Secret, instance.Spec.Secret.Replication, err)
	}

	return string(mysqlPassword), string(replicationPassword), nil
}

func compareService(serviceA, serviceB *corev1.Service) bool {
	if !reflect.DeepEqual(serviceA.Spec.Selector, serviceB.Spec.Selector) {
		return true
	} else if !reflect.DeepEqual(serviceA.Spec.Ports[0].Port, serviceB.Spec.Ports[0].Port) {
		return true
	} else if !reflect.DeepEqual(serviceA.Spec.Ports[0].TargetPort, serviceB.Spec.Ports[0].TargetPort) {
		return true
	} else if !reflect.DeepEqual(serviceA.Spec.Type, serviceB.Spec.Type) {
		return true
	}

	return false
}

func podMapFunc(_ context.Context, o client.Object) []reconcile.Request {
	pod := o.(*corev1.Pod)

	// Get MysqlReplication's name from pod labels
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
