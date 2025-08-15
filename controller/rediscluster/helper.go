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
	"fmt"
	"net"
	"time"

	"github.com/go-logr/logr"
	"sigs.k8s.io/controller-runtime/pkg/client"

	composev1alpha1 "github.com/upmio/compose-operator/api/v1alpha1"
	"github.com/upmio/compose-operator/pkg/k8sutil"
	"github.com/upmio/compose-operator/pkg/redisutil"
)

func newRedisAdmin(cluster *composev1alpha1.RedisCluster, password string, reqLogger logr.Logger) redisutil.IClusterAdmin {
	nodesAddrs := []string{}

	for _, shard := range cluster.Spec.Members {
		for _, node := range shard {
			redisPort := fmt.Sprintf("%d", node.Port)
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

// decryptSecret returns the current redis password.
func decryptSecret(client client.Client, instance *composev1alpha1.RedisCluster, reqLogger logr.Logger) (string, error) {
	passwords, err := k8sutil.DecryptSecretPasswords(
		client,
		instance.Spec.Secret.Name,
		instance.Namespace,
		[]string{instance.Spec.Secret.Redis},
	)
	if err != nil {
		return "", err
	}
	return passwords[instance.Spec.Secret.Redis], nil
}
