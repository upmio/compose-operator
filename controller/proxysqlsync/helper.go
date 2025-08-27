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

package proxysqlsync

import (
	"context"
	"github.com/go-logr/logr"
	composev1alpha1 "github.com/upmio/compose-operator/api/v1alpha1"
	"github.com/upmio/compose-operator/pkg/k8sutil"
	"github.com/upmio/compose-operator/pkg/mysqlutil"
	"github.com/upmio/compose-operator/pkg/proxysqlutil"
	"k8s.io/apimachinery/pkg/fields"
	"k8s.io/apimachinery/pkg/types"
	"net"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	"strconv"
)

// newMysqlAdmin builds and returns new mysql.ReplicationAdmin from the list of mysql nodes
func newMysqlAdmin(mrInstance *composev1alpha1.MysqlReplication, instance *composev1alpha1.ProxysqlSync, password string, reqLogger logr.Logger) mysqlutil.IReplicationAdmin {

	nodesAddrs := []string{
		net.JoinHostPort(mrInstance.Spec.Source.Host, strconv.Itoa(mrInstance.Spec.Source.Port)),
	}
	reqLogger.V(4).Info("append mysql admin source addr", "host", mrInstance.Spec.Source.Host, "port", mrInstance.Spec.Source.Port)

	adminConfig := mysqlutil.AdminOptions{
		ConnectionTimeout: 2,
		Username:          instance.Spec.Secret.Mysql,
		Password:          password,
	}

	return mysqlutil.NewReplicationAdmin(nodesAddrs, &adminConfig, reqLogger)
}

// newProxysqlAdmin builds and returns new proxysql.ReplicationAdmin from the list of proxysql nodes
func newProxysqlAdmin(instance *composev1alpha1.ProxysqlSync, password string, reqLogger logr.Logger) proxysqlutil.IAdmin {
	nodesAddrs := make([]string, 0)
	for _, nodeInfo := range instance.Spec.Proxysql {
		addr := net.JoinHostPort(nodeInfo.Host, strconv.Itoa(nodeInfo.Port))
		nodesAddrs = append(nodesAddrs, addr)
		reqLogger.V(4).Info("append proxysql admin addr", "host", nodeInfo.Host, "port", nodeInfo.Port)
	}

	adminConfig := mysqlutil.AdminOptions{
		ConnectionTimeout: 2,
		Username:          instance.Spec.Secret.Proxysql,
		Password:          password,
	}

	return proxysqlutil.NewAdmin(nodesAddrs, &adminConfig, reqLogger)
}

// decryptSecret returns the current proxysql password and mysql password.
func decryptSecret(client client.Client, reqLogger logr.Logger, instance *composev1alpha1.ProxysqlSync) (string, string, error) {
	passwords, err := k8sutil.DecryptSecretPasswords(
		client,
		reqLogger,
		instance.Spec.Secret.Name,
		instance.Namespace,
		[]string{instance.Spec.Secret.Mysql, instance.Spec.Secret.Proxysql},
	)
	if err != nil {
		return "", "", err
	}
	return passwords[instance.Spec.Secret.Mysql], passwords[instance.Spec.Secret.Proxysql], nil
}

func (r *ReconcileProxysqlSync) triggerReconcileBecauseMysqlReplicationHasChanged(ctx context.Context, mysqlReplication client.Object) []reconcile.Request {
	attachedProxysqlSyncs := &composev1alpha1.ProxysqlSyncList{}
	listOps := &client.ListOptions{
		FieldSelector: fields.OneTermEqualSelector(".spec.mysqlReplication", mysqlReplication.GetName()),
		Namespace:     mysqlReplication.GetNamespace(),
	}
	err := r.client.List(ctx, attachedProxysqlSyncs, listOps)
	if err != nil {
		r.logger.Info(err.Error())
		return []reconcile.Request{}
	}

	requests := make([]reconcile.Request, len(attachedProxysqlSyncs.Items))
	for i, item := range attachedProxysqlSyncs.Items {
		requests[i] = reconcile.Request{
			NamespacedName: types.NamespacedName{
				Name:      item.GetName(),
				Namespace: item.GetNamespace(),
			},
		}
	}
	return requests
}
