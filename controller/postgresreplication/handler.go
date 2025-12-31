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

package postgresreplication

import (
	"context"
	"errors"
	"fmt"
	"net"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/upmio/compose-operator/pkg/utils"
	corev1 "k8s.io/api/core/v1"
	kubeErrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/intstr"

	composev1alpha1 "github.com/upmio/compose-operator/api/v1alpha1"
	"github.com/upmio/compose-operator/pkg/k8sutil"
	"github.com/upmio/compose-operator/pkg/postgresutil"
)

func (r *ReconcilePostgresReplication) handlePostgresReplicationInstance(syncCtx *syncContext, replicationPassword, postgresPassword string) error {
	instance := syncCtx.instance
	ctx := syncCtx.ctx
	admin := syncCtx.admin

	replicationInfo := admin.GetReplicationStatus(ctx)

	if err := r.ensurePrimaryNode(syncCtx, replicationInfo, replicationPassword, postgresPassword); err != nil {
		return err
	}

	var errs []error
	for _, standby := range instance.Spec.Standby {
		if err := r.configureStandbyNode(syncCtx,
			replicationInfo,
			standby,
			replicationPassword); err != nil {
			errs = append(errs, err)
		}
	}

	return errors.Join(errs...)
}

func (r *ReconcilePostgresReplication) handleResources(syncCtx *syncContext) error {
	instance := syncCtx.instance
	readWriteService := fmt.Sprintf("%s-readwrite", instance.Name)
	instance.Status.ReadWriteService = readWriteService

	var errs []error
	//ensure primary pod labels
	if err := r.ensurePodLabels(syncCtx, instance.Spec.Primary.Name, "false", false); err != nil {
		errs = append(errs, err)
	}

	//ensure readwrite service
	if err := r.ensureService(syncCtx, readWriteService, "false", true); err != nil {
		errs = append(errs, err)
	}

	readOnlyService := fmt.Sprintf("%s-readonly", instance.Name)
	if len(instance.Spec.Standby) != 0 {
		instance.Status.ReadOnlyService = readOnlyService

		//ensure standby pod labels
		for _, standby := range instance.Spec.Standby {
			if err := r.ensurePodLabels(syncCtx, standby.Name, "true", standby.Isolated); err != nil {
				errs = append(errs, err)
			}
		}

		//ensure readonly service
		if err := r.ensureService(syncCtx, readOnlyService, "true", true); err != nil {
			errs = append(errs, err)

		}
	} else {
		//ensure readonly service
		if err := r.ensureService(syncCtx, readOnlyService, "true", false); err != nil {
			errs = append(errs, err)

		}
	}

	return errors.Join(errs...)
}

func (r *ReconcilePostgresReplication) ensureSyncMode(syncCtx *syncContext, nodeInfo *postgresutil.Node, address string) (err error) {

	instance := syncCtx.instance
	admin := syncCtx.admin
	ctx := syncCtx.ctx

	// 1. ensure sync mode
	standbyApplications := make([]string, 0)
	for _, standby := range instance.Spec.Standby {
		if !standby.Isolated {
			standbyApplications = append(standbyApplications, strings.ReplaceAll(standby.Name, "-", "_"))
		}
	}

	if instance.Spec.Mode == composev1alpha1.PostgresRplSync && nodeInfo.ReplicationMode != "quorum" && utils.EqualSlicesIgnoreOrder(nodeInfo.ReplicationStat, standbyApplications) {
		if len(standbyApplications) == 0 {
			return fmt.Errorf("PostgreSQL synchronous replication is enabled, but no available standby nodes were detected (standby count = 0)")
		}
		if err := admin.ConfigureSyncMode(ctx, address, strings.Join(standbyApplications, ",")); err != nil {
			return err
		}
		r.recorder.Eventf(instance, corev1.EventTypeNormal, Synced, "set mode to rpl_sync on [%s] successfully", address)
	} else if instance.Spec.Mode == composev1alpha1.PostgresRplAsync && nodeInfo.ReplicationMode != "async" {
		if err := admin.ConfigureAsyncMode(ctx, address); err != nil {
			return err
		}
		r.recorder.Eventf(instance, corev1.EventTypeNormal, Synced, "set mode to rpl_async on [%s] successfully", address)
	}

	return nil
}

func (r *ReconcilePostgresReplication) switchOver(syncCtx *syncContext, nodeInfo *postgresutil.Node, address string) (err error) {
	instance := syncCtx.instance

	if nodeInfo.Role == postgresutil.PostgresStandbyRole {
		r.recorder.Eventf(instance, corev1.EventTypeNormal, Synced, "switchover triggered: promoting standby node [%s] to primary", address)

		//1. check this node wal_diff equal 0 or will do nothing.
		if nodeInfo.WalDiff != 0 {
			return fmt.Errorf("switchover aborted: replica [%s] wal_diff '%d' exceeds threshold '0'", address, nodeInfo.WalDiff)
		}

		//2. set all node pod label readonly to true
		if err = r.ensureAllPodReadOnly(syncCtx); err != nil {
			return err
		}
		r.recorder.Event(instance, corev1.EventTypeNormal, Synced, "switchover triggered: set read-only label to true on all PostgreSQL pods")

		//3. get primary pod
		pod := &corev1.Pod{}
		if err := r.client.Get(syncCtx.ctx, types.NamespacedName{
			Name:      instance.Spec.Primary.Name,
			Namespace: instance.Namespace,
		}, pod); err != nil {
			return fmt.Errorf("failed to fetch primary pod [%s]: %v", instance.Spec.Primary.Name, err)
		}

		//4.  check primary pod is ready or not
		if !k8sutil.IsPodReady(pod) {
			return fmt.Errorf("primary pod [%s] is not ready", instance.Spec.Primary.Name)
		}

		//5. execute pg_ctl remote -d
		cmd := fmt.Sprintf("pg_ctl promote -D %s",
			nodeInfo.DataDir,
		)

		if _, _, err = r.execer.ExecCommandInContainer(pod, containerName, cmdPrefix, "-c", cmd); err != nil {
			return fmt.Errorf("failed to execute pg_ctl promote: %v", err)
		}
		r.recorder.Eventf(instance, corev1.EventTypeNormal, Synced, "switchover triggered: execute pg_ctl promote on [%s] completed", address)
	}

	return nil
}

func (r *ReconcilePostgresReplication) ensurePrimaryNode(syncCtx *syncContext, replicationInfo *postgresutil.ReplicationInfo, replicationPassword, postgresPassword string) (err error) {
	instance := syncCtx.instance

	address := net.JoinHostPort(instance.Spec.Primary.Host, strconv.Itoa(instance.Spec.Primary.Port))
	nodeInfo, ok := replicationInfo.Nodes[address]

	if !ok {
		return fmt.Errorf("cannot retrieve status from primary PostgreSQL instance [%s]", address)
	}

	// 1. make all node to be standby first
	for _, standby := range instance.Spec.Standby {
		if err := r.ensureStandbyNode(syncCtx,
			replicationInfo,
			standby,
			replicationPassword,
			postgresPassword); err != nil {
			r.recorder.Event(instance, corev1.EventTypeWarning, ErrSynced, err.Error())
		}
	}

	// 2. switchover
	if err := r.switchOver(syncCtx, nodeInfo, address); err != nil {
		return err
	}

	// 3. ensure sync mode
	return r.ensureSyncMode(syncCtx, nodeInfo, address)
}

func (r *ReconcilePostgresReplication) configureStandbyNode(syncCtx *syncContext,
	replicationInfo *postgresutil.ReplicationInfo,
	standby *composev1alpha1.ReplicaNode,
	replicationPassword string) (err error) {

	instance := syncCtx.instance
	admin := syncCtx.admin
	ctx := syncCtx.ctx

	address := net.JoinHostPort(standby.Host, strconv.Itoa(standby.Port))
	nodeInfo, ok := replicationInfo.Nodes[address]
	if !ok {
		return fmt.Errorf("cannot retrieve status from standby PostgreSQL instance [%s]", address)
	}

	replicationUser := instance.Spec.Secret.Replication
	primaryNodeHost := instance.Spec.Primary.Host
	primaryNodePort := strconv.Itoa(instance.Spec.Primary.Port)

	if nodeInfo.Role == postgresutil.PostgresStandbyRole {
		for _, standbyName := range replicationInfo.Nodes[net.JoinHostPort(primaryNodeHost, primaryNodePort)].ReplicationStat {
			if strings.ReplaceAll(standby.Name, "-", "_") == standbyName {
				if standby.Isolated {
					syncCtx.reqLogger.Info(fmt.Sprintf("standby node [%s] found in pg_stat_replication of primary node, need to isolate", address))
					if err = admin.ConfigureIsolated(ctx, address); err != nil {
						return err
					}
				} else {
					syncCtx.reqLogger.Info(fmt.Sprintf("standby node [%s] found in pg_stat_replication of primary node, skipping reconcile", address))
				}

				return nil
			}
		}
	}

	if err = admin.ConfigureStandby(ctx, address, standby.Name, primaryNodeHost, primaryNodePort, replicationUser, replicationPassword); err != nil {
		return err
	}
	r.recorder.Eventf(instance, corev1.EventTypeNormal, Synced, "configure replication on [%s] successfully", address)

	return nil
}

func (r *ReconcilePostgresReplication) closeAdminConn(syncCtx *syncContext, address string) error {
	client, err := syncCtx.admin.Connections().Get(address)
	if err != nil {
		return err
	}

	if err := client.Close(); err != nil {
		return err
	}

	syncCtx.reqLogger.Info(fmt.Sprintf("closed [%s] connection successfully", address))

	return nil
}

func (r *ReconcilePostgresReplication) reopenAdminConn(syncCtx *syncContext, address string) {
	if err := syncCtx.admin.Connections().Add(address); err != nil {
		syncCtx.reqLogger.Error(err, fmt.Sprintf("failed to reopen [%s] connection", address))
		return
	}
	syncCtx.reqLogger.Info(fmt.Sprintf("reopen [%s] connection successfully", address))
}

func (r *ReconcilePostgresReplication) getPod(ctx context.Context, name, ns string) (*corev1.Pod, error) {
	pod := &corev1.Pod{}
	if err := r.client.Get(ctx, types.NamespacedName{Name: name, Namespace: ns}, pod); err != nil {
		return nil, fmt.Errorf("failed to fetch pod [%s]: %v", name, err)
	}
	return pod, nil
}

func (r *ReconcilePostgresReplication) ensurePodReady(pod *corev1.Pod, node string) error {
	if !k8sutil.IsPodReady(pod) {
		return fmt.Errorf("standby pod [%s] is not ready", node)
	}
	return nil
}

func (r *ReconcilePostgresReplication) ensureUnitStartup(
	syncCtx *syncContext,
	node string,
	desired string,
) error {

	unit, err := r.unitController.GetUnit(syncCtx.instance.Namespace, node)
	if err != nil {
		return fmt.Errorf("failed to fetch unit [%s]: %v", node, err)
	}

	var desiredState bool
	switch desired {
	case desiredUnitStart:
		desiredState = true
	case desiredUnitStop:
		desiredState = false
	default:
		return fmt.Errorf("invalid unit [%s] desired", desired)
	}

	// already desired state
	if unit.Spec.Startup == desiredState {
		return fmt.Errorf("unit [%s] is already marked %s, skipping reconcile", node, desired)
	}

	unit.Spec.Startup = desiredState
	if err := r.unitController.UpdateUnit(unit); err != nil {
		return fmt.Errorf("failed to %s unit [%s]: %v", desired, node, err)
	}

	r.recorder.Eventf(
		syncCtx.instance,
		corev1.EventTypeNormal,
		Synced,
		"%s unit [%s] successfully",
		desired,
		node,
	)

	return nil
}

func (r *ReconcilePostgresReplication) restoreUnit(syncCtx *syncContext, node string) {
	unit, err := r.unitController.GetUnit(syncCtx.instance.Namespace, node)
	if err != nil {
		syncCtx.reqLogger.Error(err, fmt.Sprintf("failed to fetch unit [%s] to restore", node))
		return
	}

	if unit.Spec.Startup {
		return
	}

	unit.Spec.Startup = true
	if err := r.unitController.UpdateUnit(unit); err != nil {
		syncCtx.reqLogger.Error(err, fmt.Sprintf("failed to restore unit [%s]", node))
	}
}

func (r *ReconcilePostgresReplication) waitPodStopped(syncCtx *syncContext, name string) error {
	return r.waitPod(syncCtx.ctx, name, syncCtx, false, desiredUnitStop)
}

func (r *ReconcilePostgresReplication) waitPodStarted(syncCtx *syncContext, name string) error {
	return r.waitPod(syncCtx.ctx, name, syncCtx, true, desiredUnitStart)
}

func (r *ReconcilePostgresReplication) waitPod(
	ctx context.Context,
	name string,
	syncCtx *syncContext,
	expectReady bool,
	phase string) error {

	timeout := time.Minute
	ticker := time.NewTicker(2 * time.Second)
	defer ticker.Stop()

	timeoutCh := time.After(timeout)

	for {
		select {
		case <-ticker.C:
			pod, err := r.getPod(ctx, name, syncCtx.instance.Namespace)
			if err != nil {
				return err
			}

			ready := k8sutil.IsPodReady(pod)
			if ready == expectReady {
				r.recorder.Eventf(syncCtx.instance, corev1.EventTypeNormal, Synced,
					"pod [%s] %s completed", name, phase)
				return nil
			}

		case <-timeoutCh:
			return fmt.Errorf("waiting pod [%s] %s timeout", name, phase)
		}
	}
}

func (r *ReconcilePostgresReplication) syncStandbyData(
	syncCtx *syncContext,
	pod *corev1.Pod,
	nodeInfo *postgresutil.Node,
	nodeName, replicationPassword, postgresPassword string) error {

	instance := syncCtx.instance
	signalFile := filepath.Join(nodeInfo.DataDir, signalFileName)

	var cmd string
	cmd = fmt.Sprintf("[[ ! -f %s ]] || rm -rf %s",
		signalFile,
		signalFile,
	)

	if _, _, err := r.execer.ExecCommandInContainer(pod, containerName, cmdPrefix, "-c", cmd); err != nil {
		return fmt.Errorf("failed to confirm %s file does not exist in pod [%s]: %v", signalFileName, nodeName, err)
	}
	r.recorder.Eventf(instance, corev1.EventTypeNormal, Synced, "confirmed %s file does not exist in pod [%s]", signalFileName, nodeName)

	primaryNodeHost := instance.Spec.Primary.Host
	primaryNodePort := strconv.Itoa(instance.Spec.Primary.Port)
	cmd = fmt.Sprintf("pg_rewind --target-pgdata=%s --source-server='host=%s user=%s password=%s port=%s'",
		nodeInfo.DataDir,
		primaryNodeHost,
		instance.Spec.Secret.Postgresql,
		postgresPassword,
		primaryNodePort)

	if _, stderr, err := r.execer.ExecCommandInContainer(pod, containerName, cmdPrefix, "-c", cmd); err != nil {
		if strings.Contains(stderr, "pg_rewind: error: source and target clusters are from different systems") {
			cmd = fmt.Sprintf("rm -rf %s",
				nodeInfo.DataDir,
			)

			if _, _, err = r.execer.ExecCommandInContainer(pod, containerName, cmdPrefix, "-c", cmd); err != nil {
				return fmt.Errorf("failed to remove old data directory on pod [%s]: %v", nodeName, err)
			}
			r.recorder.Eventf(instance, corev1.EventTypeNormal, Synced, "remove old data directory on pod [%s] successfully", nodeName)

			cmd = fmt.Sprintf("PGPASSWORD=\"%s\" pg_basebackup -h %s -D %s -U %s -P -v -X stream",
				replicationPassword,
				primaryNodeHost,
				nodeInfo.DataDir,
				instance.Spec.Secret.Replication,
			)

			if _, _, err = r.execer.ExecCommandInContainer(pod, containerName, cmdPrefix, "-c", cmd); err != nil {
				return fmt.Errorf("failed to pg_basebackup from primary node: %v", err)
			}
			r.recorder.Eventf(instance, corev1.EventTypeNormal, Synced, "pg_basebackup from primary node on pod [%s] successfully", nodeName)

		} else {
			return fmt.Errorf("failed to execute pg_rewind: %v", err)
		}
	}

	cmd = fmt.Sprintf("[[ -f %s ]] || touch %s",
		signalFile,
		signalFile,
	)

	if _, _, err := r.execer.ExecCommandInContainer(pod, containerName, cmdPrefix, "-c", cmd); err != nil {
		return fmt.Errorf("failed to confirm %s file exists in pod [%s]: %v", signalFileName, nodeName, err)
	}
	r.recorder.Eventf(instance, corev1.EventTypeNormal, Synced, "confirmed %s file exists in pod [%s]", signalFileName, nodeName)

	return nil
}

func (r *ReconcilePostgresReplication) ensureStandbyNode(syncCtx *syncContext,
	replicationInfo *postgresutil.ReplicationInfo,
	standby *composev1alpha1.ReplicaNode,
	replicationPassword,
	postgresPassword string) (err error) {

	ctx := syncCtx.ctx

	address := net.JoinHostPort(standby.Host, strconv.Itoa(standby.Port))
	nodeName := standby.Name
	nodeInfo, ok := replicationInfo.Nodes[address]
	if !ok {
		return fmt.Errorf("cannot retrieve status from replica PostgreSQL instance [%s]", address)
	}

	// close postgres connection first, then reopen the connection.
	if nodeInfo.Role != postgresutil.PostgresPrimaryRole {
		return nil
	}

	if err := r.closeAdminConn(syncCtx, address); err != nil {
		syncCtx.reqLogger.Error(err, fmt.Sprintf("failed to close [%s] connection", address))
	}
	defer r.reopenAdminConn(syncCtx, address)

	// check pod is ready or not
	pod, err := r.getPod(ctx, standby.Name, syncCtx.instance.Namespace)
	if err != nil {
		return err
	}
	if err := r.ensurePodReady(pod, standby.Name); err != nil {
		return err
	}

	// stop unit process
	if err := r.ensureUnitStartup(syncCtx, standby.Name, "stop"); err != nil {
		return err
	}
	defer r.restoreUnit(syncCtx, standby.Name)

	// wait pod stopped
	if err := r.waitPodStopped(syncCtx, standby.Name); err != nil {
		return err
	}

	if err := r.syncStandbyData(syncCtx, pod, nodeInfo, nodeName, replicationPassword, postgresPassword); err != nil {
		return err
	}

	if err := r.ensureUnitStartup(syncCtx, standby.Name, "start"); err != nil {
		return err
	}

	return r.waitPodStarted(syncCtx, standby.Name)
}

func (r *ReconcilePostgresReplication) ensureService(syncCtx *syncContext, serviceName, isReadOnly string, isExisted bool) error {
	instance := syncCtx.instance
	ctx := syncCtx.ctx

	ownerReference := composev1alpha1.DefaultPostgresReplicationOwnerReferences(instance)

	desiredService := &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:            serviceName,
			Namespace:       instance.Namespace,
			Labels:          instance.Labels,
			OwnerReferences: ownerReference,
		},
		Spec: corev1.ServiceSpec{
			Ports: []corev1.ServicePort{
				{
					Name:       portName,
					Protocol:   corev1.ProtocolTCP,
					Port:       int32(instance.Spec.Primary.Port),
					TargetPort: intstr.IntOrString{IntVal: int32(instance.Spec.Primary.Port)},
				},
			},
			Selector: map[string]string{
				readOnlyKey: isReadOnly,
				defaultKey:  instance.Name,
			},
		},
	}

	//transfer service type
	switch instance.Spec.Service.Type {
	case composev1alpha1.ServiceTypeClusterIP:
		desiredService.Spec.Type = corev1.ServiceTypeClusterIP
	case composev1alpha1.ServiceTypeNodePort:
		desiredService.Spec.Type = corev1.ServiceTypeNodePort
	case composev1alpha1.ServiceTypeLoadBalancer:
		desiredService.Spec.Type = corev1.ServiceTypeLoadBalancer
	case composev1alpha1.ServiceTypeExternalName:
		desiredService.Spec.Type = corev1.ServiceTypeExternalName
	}

	//ensure service object
	foundService := &corev1.Service{}
	err := r.client.Get(ctx, types.NamespacedName{
		Name:      serviceName,
		Namespace: instance.Namespace,
	}, foundService)

	if isExisted {
		if kubeErrors.IsNotFound(err) {
			//create service
			if err := r.client.Create(ctx, desiredService); err != nil {
				return fmt.Errorf("failed to create service [%s]: %v", serviceName, err)
			}
			r.recorder.Eventf(instance, corev1.EventTypeNormal, Synced, "create service [%s] successfully", serviceName)

		} else if err != nil {
			return fmt.Errorf("failed to fetch service [%s]: %v", serviceName, err)
		} else if compareService(desiredService, foundService) {
			//update service
			if err := r.client.Update(ctx, desiredService); err != nil {
				return fmt.Errorf("failed to update service [%s]: %v", serviceName, err)
			}
			r.recorder.Eventf(instance, corev1.EventTypeNormal, Synced, "update service [%s] successfully", serviceName)
		}
	} else {
		if err != nil && !kubeErrors.IsNotFound(err) {
			return fmt.Errorf("failed to fetch service [%s]: %v", serviceName, err)
		} else if err == nil {
			if err := r.client.Delete(ctx, foundService); err != nil {
				return fmt.Errorf("failed to remove service [%s]: %v", serviceName, err)
			}
			r.recorder.Eventf(instance, corev1.EventTypeNormal, Synced, "remove service [%s] successfully", serviceName)
		}
	}

	return nil
}

func (r *ReconcilePostgresReplication) ensureAllPodReadOnly(syncCtx *syncContext) error {
	instance := syncCtx.instance
	var errs []error

	//ensure primary pod labels
	if err := r.ensurePodLabels(syncCtx, instance.Spec.Primary.Name, "true", false); err != nil {
		errs = append(errs, err)
	}

	//ensure standby pod labels
	for _, standby := range instance.Spec.Standby {
		if err := r.ensurePodLabels(syncCtx, standby.Name, "true", false); err != nil {
			errs = append(errs, err)
		}
	}

	return errors.Join(errs...)
}

func (r *ReconcilePostgresReplication) ensurePodLabels(syncCtx *syncContext, podName, isReadOnly string, preserveLabels bool) error {
	ctx := syncCtx.ctx
	instance := syncCtx.instance

	foundPod := &corev1.Pod{}
	if err := r.client.Get(ctx, types.NamespacedName{
		Name:      podName,
		Namespace: instance.Namespace,
	}, foundPod); err != nil {
		return fmt.Errorf("failed to fetch pod [%s]: %v", podName, err)
	}

	if foundPod.Labels == nil {
		foundPod.Labels = make(map[string]string)
	}

	var needsUpdate bool
	var eventMessage string

	if preserveLabels {
		// Remove managed labels if they exist
		needsUpdate, eventMessage = r.removeLabelsFromPod(foundPod, instance.Name)
	} else {
		// Ensure labels are set correctly
		needsUpdate, eventMessage = r.setLabelsOnPod(foundPod, instance.Name, isReadOnly)
	}

	// Update pod only if changes are needed
	if needsUpdate {
		if err := r.client.Update(ctx, foundPod); err != nil {
			return fmt.Errorf("failed to update pod [%s]: %v", podName, err)
		}
		r.recorder.Eventf(instance, corev1.EventTypeNormal, Synced, "pod [%s] %s", podName, eventMessage)
	}

	return nil
}

// setLabelsOnPod sets the required labels on the pod and returns whether update is needed
func (r *ReconcilePostgresReplication) setLabelsOnPod(pod *corev1.Pod, instanceName, isReadOnly string) (bool, string) {
	var needsUpdate bool
	var updatedLabels []string

	// Check and update readonly label
	if readOnlyValue, ok := pod.Labels[readOnlyKey]; !ok || readOnlyValue != isReadOnly {
		pod.Labels[readOnlyKey] = isReadOnly
		needsUpdate = true
		updatedLabels = append(updatedLabels, readOnlyKey)
	}

	// Check and update default label
	if instanceValue, ok := pod.Labels[defaultKey]; !ok || instanceValue != instanceName {
		pod.Labels[defaultKey] = instanceName
		needsUpdate = true
		updatedLabels = append(updatedLabels, defaultKey)
	}

	var eventMessage string
	if needsUpdate {
		eventMessage = fmt.Sprintf("update labels '%s' successfully", strings.Join(updatedLabels, ", "))
	}

	return needsUpdate, eventMessage
}

// removeLabelsFromPod removes the managed labels from the pod and returns whether update is needed
func (r *ReconcilePostgresReplication) removeLabelsFromPod(pod *corev1.Pod, instanceName string) (bool, string) {
	var needsUpdate bool
	var removedLabels []string

	// Remove readonly label if it exists
	if _, ok := pod.Labels[readOnlyKey]; ok {
		delete(pod.Labels, readOnlyKey)
		needsUpdate = true
		removedLabels = append(removedLabels, readOnlyKey)
	}

	// Remove default label if it exists and matches the instance
	if instanceValue, ok := pod.Labels[defaultKey]; ok && instanceValue == instanceName {
		delete(pod.Labels, defaultKey)
		needsUpdate = true
		removedLabels = append(removedLabels, defaultKey)
	}

	var eventMessage string
	if needsUpdate {
		eventMessage = fmt.Sprintf("remove labels '%s' successfully", strings.Join(removedLabels, ", "))
	}

	return needsUpdate, eventMessage
}
