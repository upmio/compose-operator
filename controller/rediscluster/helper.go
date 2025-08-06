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

package rediscluster

import (
	"context"
	"fmt"
	"net"
	"time"

	"github.com/go-logr/logr"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"

	composev1alpha1 "github.com/upmio/compose-operator/api/v1alpha1"
	"github.com/upmio/compose-operator/pkg/redisutil"
	"github.com/upmio/compose-operator/pkg/utils"
)

func newRedisAdmin(cluster *composev1alpha1.RedisCluster, password string, reqLogger logr.Logger) redisutil.IClusterAdmin {
	nodesAddrs := []string{}

	for _, shard := range cluster.Spec.Members {
		redisPort := redisutil.DefaultRedisPort
		for _, node := range shard {
			redisPort = fmt.Sprintf("%d", node.Port)
			reqLogger.V(4).Info("append redis admin addr", "host", node.Host, "port", node.Port)
			ipaddr, err := net.LookupHost(node.Host)
			if err != nil {
				reqLogger.Info(fmt.Sprintf("failed to lookup host [%s]: %v", node.Host, err.Error()))
				continue
			}
			if len(ipaddr) == 0 || ipaddr == nil {
				reqLogger.Info(fmt.Sprintf("can't resolv host [%s] ip address", node.Host))
				continue
			}
			nodesAddrs = append(nodesAddrs, net.JoinHostPort(ipaddr[0], redisPort))
		}
	}

	adminConfig := redisutil.AdminOptions{
		ConnectionTimeout: 5 * time.Second,
		Password:          password,
	}

	return redisutil.NewClusterAdmin(nodesAddrs, &adminConfig, reqLogger)
}

// decryptSecret return current redis password.
func decryptSecret(client client.Client, instance *composev1alpha1.RedisCluster, reqLogger logr.Logger) (string, error) {
	secret := &corev1.Secret{}
	err := client.Get(context.TODO(), types.NamespacedName{
		Name:      instance.Spec.Secret.Name,
		Namespace: instance.Namespace,
	}, secret)

	if err != nil {
		return "", fmt.Errorf("failed to fetch secret [%s]: %v", instance.Spec.Secret, err)
	}

	password, err := utils.AES_CTR_Decrypt(secret.Data[instance.Spec.Secret.Redis])
	if err != nil {
		return "", fmt.Errorf("failed to decrypt secret [%s] key '%s': %v", instance.Spec.Secret, instance.Spec.Secret.Redis, err)
	}

	return string(password), nil
}
