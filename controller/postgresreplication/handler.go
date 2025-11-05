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
	"errors"
	"fmt"
	"net"
	"path/filepath"
	"strconv"
	"strings"
	"time"

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
	if err := r.ensurePodLabels(syncCtx, instance.Spec.Primary.Name, "false"); err != nil {
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
			if err := r.ensurePodLabels(syncCtx, standby.Name, "true"); err != nil {
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

func (r *ReconcilePostgresReplication) ensurePrimaryNode(syncCtx *syncContext, replicationInfo *postgresutil.ReplicationInfo, replicationPassword, postgresPassword string) (err error) {
	instance := syncCtx.instance
	admin := syncCtx.admin
	ctx := syncCtx.ctx

	address := net.JoinHostPort(instance.Spec.Primary.Host, strconv.Itoa(instance.Spec.Primary.Port))
	nodeInfo, ok := replicationInfo.Nodes[address]

	if !ok {
		return fmt.Errorf("cannot retrieve status from primary PostgreSQL instance [%s]", address)
	}

	// 1. ensure sync mode
	standbyApplications := make([]string, 0)
	for _, standby := range instance.Spec.Standby {
		standbyApplications = append(standbyApplications, strings.ReplaceAll(standby.Name, "-", "_"))
	}

	if instance.Spec.Mode == composev1alpha1.PostgresRplSync && nodeInfo.ReplicationMode != "quorum" {
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

	// 2. make old primary node to be standby
	for _, standby := range instance.Spec.Standby {
		if err := r.ensureStandbyNode(syncCtx,
			replicationInfo,
			standby,
			replicationPassword,
			postgresPassword); err != nil {
			r.recorder.Event(instance, corev1.EventTypeWarning, ErrSynced, err.Error())
		}
	}

	// 3. switch
	if nodeInfo.Role == postgresutil.PostgresStandbyRole {
		r.recorder.Eventf(instance, corev1.EventTypeNormal, Synced, "switchover triggered: promoting standby node [%s] to priamry", address)

		//1. check this node wal_diff equal 0 or will do nothing.
		if nodeInfo.WalDiff != 0 {
			return fmt.Errorf("switchover aborted: replica [%s] wal_diff '%d' exceeds threshol '0'", address, nodeInfo.WalDiff)
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

		if _, _, err = r.execer.ExecCommandInContainer(pod, containerName, "/bin/sh", "-c", cmd); err != nil {
			return fmt.Errorf("failed to execute pg_ctl promote: %v", err)
		}
		r.recorder.Eventf(instance, corev1.EventTypeNormal, Synced, "switchover triggered: execute pg_ctl promote on [%] completed", address)
	}

	return nil
}
func (r *ReconcilePostgresReplication) configureStandbyNode(syncCtx *syncContext,
	replicationInfo *postgresutil.ReplicationInfo,
	standby *composev1alpha1.CommonNode,
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
				syncCtx.reqLogger.Info(fmt.Sprintf("standby node [%s] found in pg_stat_replication of primary node, skipping reconcile", address))
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

func (r *ReconcilePostgresReplication) ensureStandbyNode(syncCtx *syncContext,
	replicationInfo *postgresutil.ReplicationInfo,
	standby *composev1alpha1.CommonNode,
	replicationPassword,
	postgresPassword string) (err error) {

	instance := syncCtx.instance
	admin := syncCtx.admin
	ctx := syncCtx.ctx

	address := net.JoinHostPort(standby.Host, strconv.Itoa(standby.Port))
	nodeName := standby.Name
	nodeInfo, ok := replicationInfo.Nodes[address]
	if !ok {
		return fmt.Errorf("cannot retrieve status from replica MySQL instance [%s]", address)
	}

	primaryNodeHost := instance.Spec.Primary.Host
	primaryNodePort := strconv.Itoa(instance.Spec.Primary.Port)

	if nodeInfo.Role == postgresutil.PostgresPrimaryRole {

		client, err := admin.Connections().Get(address)
		if err != nil {
			syncCtx.reqLogger.Error(err, fmt.Sprintf("failed to get client for [%s] from admin connection pool", address))
		} else {
			if err := client.Close(); err != nil {
				syncCtx.reqLogger.Error(err, fmt.Sprintf("failed to close connection to [%s]", address))
			} else {
				syncCtx.reqLogger.Info(fmt.Sprintf("closed connection to [%s] successfully", address))
			}
		}

		defer func() {
			if err := admin.Connections().Add(address); err != nil {
				syncCtx.reqLogger.Error(err, fmt.Sprintf("failed to add connection to [%s]", address))
			} else {
				syncCtx.reqLogger.Info(fmt.Sprintf("successfully added connection to [%s]", address))
			}
		}()

		pod := &corev1.Pod{}
		if err := r.client.Get(ctx, types.NamespacedName{
			Name:      nodeName,
			Namespace: instance.Namespace,
		}, pod); err != nil {
			return fmt.Errorf("failed to fetch pod [%s]: %v", nodeName, err)
		}

		if !k8sutil.IsPodReady(pod) {
			return fmt.Errorf("standby pod [%s] is not ready", nodeName)
		}

		if unit, err := r.unitController.GetUnit(instance.Namespace, nodeName); err != nil {
			return fmt.Errorf("failed to fetch unit [%s]: %v", nodeName, err)
		} else if unit.Spec.Startup {
			unit.Spec.Startup = false
			if err := r.unitController.UpdateUnit(unit); err != nil {
				return fmt.Errorf("failed to set unit [%s] stop: %v", nodeName, err)
			}
			r.recorder.Eventf(instance, corev1.EventTypeNormal, Synced, "set unit [%s] stop successfully", nodeName)
		} else {
			return fmt.Errorf("unit [%s] is already marked as stop, skipping reconcile", nodeName)
		}

		defer func() {
			if unit, err := r.unitController.GetUnit(instance.Namespace, standby.Name); err != nil {
				syncCtx.reqLogger.Error(err, fmt.Sprintf("failed to fetch unit [%s]", nodeName))
			} else if !unit.Spec.Startup {
				unit.Spec.Startup = true
				if err := r.unitController.UpdateUnit(unit); err != nil {
					syncCtx.reqLogger.Error(err, fmt.Sprintf("failed to set unit [%s] started", nodeName))
				}
			}
		}()

		//Start the timer and produce a Timer object
		timeout := time.Minute
		interval := 2 * time.Second

		timeoutCh1 := time.After(timeout)
		ticker1 := time.NewTicker(interval)
		defer ticker1.Stop()

	LOOP1:
		for {
			select {
			case <-ticker1.C:
				pod := &corev1.Pod{}
				if err := r.client.Get(ctx, types.NamespacedName{
					Name:      nodeName,
					Namespace: instance.Namespace,
				}, pod); err != nil {
					return fmt.Errorf("failed to fetch pod [%s]: %v", nodeName, err)
				}

				if !k8sutil.IsPodReady(pod) {
					r.recorder.Eventf(instance, corev1.EventTypeNormal, Synced, "finished waiting for pod [%s] to stop", nodeName)
					break LOOP1
				}

				syncCtx.reqLogger.Info(fmt.Sprintf("pod [%s] is stopping, will recheck in 5 seconds...", nodeName))
			case <-timeoutCh1:
				return fmt.Errorf("waiting for pod [%s] to stop timed out", nodeName)
			}
		}

		//execute pg_rewind
		var cmd string
		cmd = fmt.Sprintf("[[ ! -f %s ]] || rm -rf %s",
			filepath.Join(nodeInfo.DataDir, "standby.signal"),
			filepath.Join(nodeInfo.DataDir, "standby.signal"),
		)

		if _, _, err = r.execer.ExecCommandInContainer(pod, containerName, "/bin/sh", "-c", cmd); err != nil {
			return fmt.Errorf("failed to confirm standby.signal file does not exist in pod [%s]: %v", nodeName, err)
		}
		r.recorder.Eventf(instance, corev1.EventTypeNormal, Synced, "confirmed standby.signal file does not exist in pod [%s]", nodeName)

		cmd = fmt.Sprintf("pg_rewind --target-pgdata=%s --source-server='host=%s user=%s password=%s port=%s'",
			nodeInfo.DataDir,
			primaryNodeHost,
			instance.Spec.Secret.Postgresql,
			postgresPassword,
			primaryNodePort)

		if _, stderr, err := r.execer.ExecCommandInContainer(pod, containerName, "/bin/sh", "-c", cmd); err != nil {
			if strings.Contains(stderr, "pg_rewind: error: source and target clusters are from different systems") {
				cmd = fmt.Sprintf("rm -rf %s",
					nodeInfo.DataDir,
				)

				if _, _, err = r.execer.ExecCommandInContainer(pod, containerName, "/bin/sh", "-c", cmd); err != nil {
					return fmt.Errorf("failed to remove old data directory on pod [%s]: %v", nodeName, err)
				}
				r.recorder.Eventf(instance, corev1.EventTypeNormal, Synced, "remove old data directory on pod [%s] successfully", nodeName)

				// su postgres -c 'PGPASSWORD="D&a6!40Lv82Qq49O" pg_basebackup -h tddotpnq-postgresql-n39-0.tddotpnq-postgresql-n39-headless-svc.demo -D /DATA_MOUNT/backup -U replication -P -v -X stream'
				cmd = fmt.Sprintf("PGPASSWORD=\"%s\" pg_basebackup -h %s -D %s -U %s -P -v -X stream",
					replicationPassword,
					primaryNodeHost,
					nodeInfo.DataDir,
					instance.Spec.Secret.Replication,
				)

				if _, _, err = r.execer.ExecCommandInContainer(pod, containerName, "/bin/sh", "-c", cmd); err != nil {
					return fmt.Errorf("failed to pg_basebackup from primary node: %v", err)
				}
				r.recorder.Eventf(instance, corev1.EventTypeNormal, Synced, "pg_basebackup from primary node on pod [%s] successfully", nodeName)

			} else {
				return fmt.Errorf("failed to execute pg_rewind: %v", err)
			}
		}

		cmd = fmt.Sprintf("[[ -f %s ]] || touch %s",
			filepath.Join(nodeInfo.DataDir, "standby.signal"),
			filepath.Join(nodeInfo.DataDir, "standby.signal"),
		)

		if _, _, err = r.execer.ExecCommandInContainer(pod, containerName, "/bin/sh", "-c", cmd); err != nil {
			return fmt.Errorf("failed to confirm standby.signal file exists in pod [%s]: %v", nodeName, err)
		}
		r.recorder.Eventf(instance, corev1.EventTypeNormal, Synced, "confirmed standby.signal file exists in pod [%s]", nodeName)

		if unit, err := r.unitController.GetUnit(instance.Namespace, nodeName); err != nil {
			return fmt.Errorf("failed to fetch unit [%s]: %v", nodeName, err)
		} else if !unit.Spec.Startup {
			unit.Spec.Startup = true
			if err := r.unitController.UpdateUnit(unit); err != nil {
				return fmt.Errorf("failed to set unit [%s] start: %v", nodeName, err)
			}
			r.recorder.Eventf(instance, corev1.EventTypeNormal, Synced, "set unit [%s] start successfully", nodeName)
		}

		//Start the timer and produce a Timer object
		timeoutCh2 := time.After(timeout)
		ticker2 := time.NewTicker(interval)
		defer ticker2.Stop()

	LOOP2:
		for {
			select {
			case <-ticker2.C:
				pod := &corev1.Pod{}
				if err := r.client.Get(ctx, types.NamespacedName{
					Name:      nodeName,
					Namespace: instance.Namespace,
				}, pod); err != nil {
					return fmt.Errorf("failed to fetch pod [%s]: %v", nodeName, err)
				}

				if k8sutil.IsPodReady(pod) {
					r.recorder.Eventf(instance, corev1.EventTypeNormal, Synced, "finished waiting for pod [%s] to start", nodeName)
					break LOOP2
				}

				syncCtx.reqLogger.Info(fmt.Sprintf("pod [%s] is starting, will recheck in 5 seconds...", nodeName))
			case <-timeoutCh2:
				return fmt.Errorf("[%s] waiting for pod start timeout", nodeName)
			}
		}
	}

	return nil
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
	if err := r.ensurePodLabels(syncCtx, instance.Spec.Primary.Name, "true"); err != nil {
		errs = append(errs, err)
	}

	//ensure standby pod labels
	for _, standby := range instance.Spec.Standby {
		if err := r.ensurePodLabels(syncCtx, standby.Name, "true"); err != nil {
			errs = append(errs, err)
		}
	}

	return errors.Join(errs...)
}

func (r *ReconcilePostgresReplication) ensurePodLabels(syncCtx *syncContext, podName, isReadOnly string) error {
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

	// update the value of pod readonly label
	if readOnlyValue, ok := foundPod.Labels[readOnlyKey]; !ok || readOnlyValue != isReadOnly {
		foundPod.Labels[readOnlyKey] = isReadOnly
		// update the value of pod default label
		if instanceValue, ok := foundPod.Labels[defaultKey]; !ok || instanceValue != instance.Name {
			foundPod.Labels[defaultKey] = instance.Name
		}
		// update pod
		if err := r.client.Update(ctx, foundPod); err != nil {
			return err
		}
		r.recorder.Eventf(instance, corev1.EventTypeNormal, Synced, "pod [%s] update labels successfully", podName)
	}

	return nil
}
