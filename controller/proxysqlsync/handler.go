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
	"errors"
	"fmt"
	"github.com/upmio/compose-operator/pkg/proxysqlutil"
	corev1 "k8s.io/api/core/v1"
	"net"
	"reflect"
	"strconv"
)

func (r *ReconcileProxysqlSync) syncMysqlReplicationServer(syncCtx *syncContext) error {
	mrInstance := syncCtx.mrInstance
	instance := syncCtx.instance

	if !mrInstance.Status.Ready {
		return fmt.Errorf("MySQL replication [%s] is not ready; please check the replication status", mrInstance.Name)
	}

	version, err := syncCtx.mysqlAdmin.GetVersion(syncCtx.ctx, net.JoinHostPort(mrInstance.Spec.Source.Host, strconv.Itoa(mrInstance.Spec.Source.Port)))
	if err != nil {
		r.recorder.Eventf(instance, corev1.EventTypeWarning, ErrSynced, "failed to get mysql version: %v", err)
	}

	expectedHostgroup := &proxysqlutil.ReplicationHostgroup{
		WriterHostGroupId: WriterHostGroupId,
		ReaderHostGroupId: ReaderHostGroupId,
		CheckType:         "read_only",
		Comment:           string(mrInstance.Spec.Mode),
	}

	expectedServers := proxysqlutil.Servers{
		net.JoinHostPort(mrInstance.Spec.Source.Host, strconv.Itoa(mrInstance.Spec.Source.Port)): &proxysqlutil.Server{
			HostGroupId:    WriterHostGroupId,
			Hostname:       mrInstance.Spec.Source.Host,
			Port:           mrInstance.Spec.Source.Port,
			MaxConnections: MaxConnections,
		},
	}

	for _, replica := range mrInstance.Spec.Replica {
		expectedServers[net.JoinHostPort(replica.Host, strconv.Itoa(replica.Port))] = &proxysqlutil.Server{
			HostGroupId:    ReaderHostGroupId,
			Hostname:       replica.Host,
			Port:           replica.Port,
			MaxConnections: MaxConnections,
		}
	}

	var errs []error
	for _, node := range instance.Spec.Proxysql {

		instance.Status.Topology[node.Name].Synced = true
		address := net.JoinHostPort(node.Host, strconv.Itoa(node.Port))

		if err := syncCtx.proxysqlAdmin.SyncMysqlVersion(syncCtx.ctx, address, version); err != nil {
			errs = append(errs, err)
		}

		if foundHostgroup, err := syncCtx.proxysqlAdmin.GetRuntimeMysqlReplicationHostGroup(syncCtx.ctx, address, WriterHostGroupId); err != nil {
			syncCtx.instance.Status.Topology[node.Name].Synced = false
			errs = append(errs, err)
		} else if !reflect.DeepEqual(foundHostgroup, expectedHostgroup) {
			if err := syncCtx.proxysqlAdmin.SyncMysqlReplicationHostGroup(syncCtx.ctx, address, expectedHostgroup); err != nil {
				syncCtx.instance.Status.Topology[node.Name].Synced = false
				errs = append(errs, err)
			} else {
				r.recorder.Eventf(instance, corev1.EventTypeNormal, Synced, "sync MySQL replication hostgroup on [%s] successfully", address)
			}
		}

		if foundServers, err := syncCtx.proxysqlAdmin.GetRuntimeMysqlReplicationServers(syncCtx.ctx, address); err != nil {
			syncCtx.instance.Status.Topology[node.Name].Synced = false
			errs = append(errs, err)
		} else if !reflect.DeepEqual(foundServers, expectedServers) {
			if err := syncCtx.proxysqlAdmin.SyncMysqlReplicationServers(syncCtx.ctx, address, expectedServers, foundServers); err != nil {
				syncCtx.instance.Status.Topology[node.Name].Synced = false
				errs = append(errs, err)
			} else {
				r.recorder.Eventf(instance, corev1.EventTypeNormal, Synced, "sync MySQL replication servers on [%s] successfully", address)
			}
		}

	}

	return errors.Join(errs...)
}

func (r *ReconcileProxysqlSync) syncMysqlUser(syncCtx *syncContext) error {
	mrInstance := syncCtx.mrInstance
	instance := syncCtx.instance

	mysqlAddress := net.JoinHostPort(mrInstance.Spec.Source.Host, strconv.Itoa(mrInstance.Spec.Source.Port))
	expectedUsers, err := syncCtx.mysqlAdmin.GetUser(syncCtx.ctx, mysqlAddress, instance.Spec.Rule.Pattern, instance.Spec.Rule.Filter)
	if err != nil {
		return err
	}

	var errs []error

	for _, node := range instance.Spec.Proxysql {
		address := net.JoinHostPort(node.Host, strconv.Itoa(node.Port))

		if foundUsers, err := syncCtx.proxysqlAdmin.GetRuntimeMysqlUsers(syncCtx.ctx, address); err != nil {
			syncCtx.instance.Status.Topology[node.Name].Synced = false
			errs = append(errs, err)
		} else if !reflect.DeepEqual(foundUsers, expectedUsers) {
			if err := syncCtx.proxysqlAdmin.SyncMysqlUsers(syncCtx.ctx, address, WriterHostGroupId, MaxConnections, expectedUsers, foundUsers); err != nil {
				syncCtx.instance.Status.Topology[node.Name].Synced = false
				errs = append(errs, err)
			} else {
				r.recorder.Eventf(instance, corev1.EventTypeNormal, Synced, "sync MySQL users on [%s] successfully", address)
			}
		}
	}

	return errors.Join(errs...)
}
