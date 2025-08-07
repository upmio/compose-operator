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
	"errors"
	"fmt"
	"net"
	"strconv"
	"strings"

	"k8s.io/apimachinery/pkg/types"

	corev1 "k8s.io/api/core/v1"

	"github.com/upmio/compose-operator/pkg/mysqlutil"
)

func (r *ReconcileMysqlGroupReplication) handleMysqlGroupReplicationInstance(syncCtx *syncContext, replicationPassword string) error {
	instance := syncCtx.instance
	ctx := syncCtx.ctx
	admin := syncCtx.admin

	replicationUser := instance.Spec.Secret.Replication
	groupInfo := admin.GetGroupInfos(ctx, len(instance.Spec.Member))

	if err := r.ensureMGRConfig(syncCtx, groupInfo); err != nil {
		return err
	}

	switch groupInfo.Status {
	case mysqlutil.GroupInfoConsistent:
		syncCtx.reqLogger.Info("group status is consistent")
		return nil
	case mysqlutil.GroupInfoPartial:
		syncCtx.reqLogger.Info("group status is partial")
	case mysqlutil.GroupInfoUnset:
		primaryAddr := groupInfo.ElectPrimary()
		if primaryAddr == "" {
			primaryAddr = net.JoinHostPort(instance.Spec.Member[0].Host, strconv.Itoa(instance.Spec.Member[0].Port))
		}

		syncCtx.reqLogger.Info(fmt.Sprintf("group status is unset, try setup group process on [%s]", primaryAddr))

		if err := admin.SetupGroup(ctx, primaryAddr, replicationUser, replicationPassword); err != nil {
			return err
		}

		r.recorder.Eventf(instance, corev1.EventTypeNormal, Synced, "setup group on [%s] successfully", primaryAddr)

		delete(groupInfo.Infos, primaryAddr)

	case mysqlutil.GroupInfoInconsistent:
		return fmt.Errorf("group status is inconsistent, please verify manually")
	case mysqlutil.GroupInfoUnavailable:
		var errs []error
		syncCtx.reqLogger.Info("group status is unavailable")

		for addr, node := range groupInfo.Infos {
			if node.State == mysqlutil.MysqlOfflineState {
				continue
			}
			if err := admin.StopGroupReplication(ctx, addr); err != nil {
				errs = append(errs, err)
			}
			r.recorder.Eventf(instance, corev1.EventTypeNormal, Synced, "group status is unavailable, stop group replication on [%s] successfully", addr)

		}

		return errors.Join(errs...)
	}

	var errs []error

	for addr, node := range groupInfo.Infos {
		if node.State == mysqlutil.MysqlOfflineState || node.State == mysqlutil.MysqlMissingState {
			syncCtx.reqLogger.Info(fmt.Sprintf("found [%s] in unexpected %s state, attempting to join group", addr, node.State))

			if err := admin.JoinGroup(ctx, addr, replicationUser, replicationPassword); err != nil {
				errs = append(errs, err)
			} else {
				r.recorder.Eventf(instance, corev1.EventTypeNormal, Synced, "[%s] joined group successfully", addr)

			}
		}
	}

	return errors.Join(errs...)
}
func (r *ReconcileMysqlGroupReplication) ensureMGRConfig(syncCtx *syncContext, groupInfos *mysqlutil.GroupInfos) error {
	instance := syncCtx.instance
	ctx := syncCtx.ctx
	admin := syncCtx.admin

	var (
		errs                      []error
		MGRSeedList, MGRAllowList []string
	)

	// ensure group_replication_ip_allowlist group_replication_group_seeds
	for _, node := range instance.Spec.Member {
		MGRSeedList = append(MGRSeedList, fmt.Sprintf("%s:%d", node.Host, 33061))
		MGRAllowList = append(MGRAllowList, node.Host)
	}
	seeds := strings.Join(MGRSeedList, ",")
	allowList := strings.Join(MGRAllowList, ",")

	for addr := range groupInfos.Infos {
		if err := admin.EnsureGroupSeeds(ctx, addr, seeds, allowList); err != nil {
			errs = append(errs, err)
		}
	}

	return errors.Join(errs...)
}

func (r *ReconcileMysqlGroupReplication) handleResources(syncCtx *syncContext) error {
	instance := syncCtx.instance

	var errs []error
	// set pod label
	for _, node := range instance.Spec.Member {
		if err := r.ensurePodLabels(syncCtx, node.Name); err != nil {
			errs = append(errs, err)
		}
	}

	return errors.Join(errs...)
}

func (r *ReconcileMysqlGroupReplication) ensurePodLabels(syncCtx *syncContext, podName string) error {
	ctx := syncCtx.ctx
	instance := syncCtx.instance
	namespace := instance.Namespace
	instanceName := instance.Name

	foundPod := &corev1.Pod{}
	if err := r.client.Get(ctx, types.NamespacedName{
		Name:      podName,
		Namespace: namespace,
	}, foundPod); err != nil {
		return fmt.Errorf("failed to fetch pod [%s]: %v", podName, err)
	}

	// Check and update default label
	if instanceValue, ok := foundPod.Labels[defaultKey]; !ok || instanceValue != instanceName {
		foundPod.Labels[defaultKey] = instanceName

		if err := r.client.Update(ctx, foundPod); err != nil {
			return fmt.Errorf("failed to update pod [%s]: %v", podName, err)
		}
		r.recorder.Eventf(instance, corev1.EventTypeNormal, Synced, "pod [%s] update labels '%s' successfully", podName, defaultKey)
	}

	return nil
}
