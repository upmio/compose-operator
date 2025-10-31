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

package redisreplication

import (
	"errors"
	"fmt"
	"net"
	"strconv"
	"strings"
	"time"

	corev1 "k8s.io/api/core/v1"
	kubeErrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/intstr"

	composev1alpha1 "github.com/upmio/compose-operator/api/v1alpha1"
	"github.com/upmio/compose-operator/pkg/redisutil"
)

const (
	redisSwitchoverCatchUpTimeout = 30 * time.Second
)

func (r *ReconcileRedisReplication) handleRedisReplicationInstance(syncCtx *syncContext) error {
	instance := syncCtx.instance
	admin := syncCtx.admin

	replicationInfo := admin.GetReplicationStatus()

	if err := r.ensureSourceNode(syncCtx, replicationInfo); err != nil {
		return err
	}

	var errs []error
	for _, replica := range instance.Spec.Replica {
		if err := r.ensureReplicaNode(syncCtx,
			replicationInfo,
			replica); err != nil {
			errs = append(errs, err)
		}
	}

	return errors.Join(errs...)
}

func (r *ReconcileRedisReplication) handleResources(syncCtx *syncContext) error {
	instance := syncCtx.instance
	readWriteService := fmt.Sprintf("%s-readwrite", instance.Name)
	instance.Status.ReadWriteService = readWriteService

	var errs []error
	//ensure source pod labels
	if err := r.ensurePodLabels(syncCtx, instance.Spec.Source.Name, "false", false); err != nil {
		errs = append(errs, err)
	}

	//ensure readwrite service
	if err := r.ensureService(syncCtx, readWriteService, "false", true); err != nil {
		errs = append(errs, err)
	}

	readOnlyService := fmt.Sprintf("%s-readonly", instance.Name)
	if len(instance.Spec.Replica) != 0 {
		instance.Status.ReadOnlyService = readOnlyService

		//ensure replica pod labels
		for _, replica := range instance.Spec.Replica {
			if err := r.ensurePodLabels(syncCtx, replica.Name, "true", replica.Isolated); err != nil {
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

	//ensure sentinel pod labels
	if len(instance.Spec.Sentinel) != 0 {
		if err := r.ensureSentinelPodLabels(syncCtx); err != nil {
			errs = append(errs, err)
		}
	}

	return errors.Join(errs...)
}

func (r *ReconcileRedisReplication) ensureSourceNode(syncCtx *syncContext, replicationInfo *redisutil.ReplicationInfo) error {
	instance := syncCtx.instance
	admin := syncCtx.admin

	address := net.JoinHostPort(instance.Spec.Source.Host, strconv.Itoa(instance.Spec.Source.Port))
	nodeInfo, ok := replicationInfo.Nodes[address]

	if !ok {
		return fmt.Errorf("cannot retrieve status from source Redis instance [%s]", address)
	}

	if nodeInfo.Role == redisutil.RedisReplicaRole {
		//switchover logic
		r.recorder.Eventf(instance, corev1.EventTypeNormal, Synced, "switchover triggered: promoting replica node [%s] to source", address)

		if err := r.ensureAllPodReadOnly(syncCtx); err != nil {
			return err
		}
		r.recorder.Event(instance, corev1.EventTypeNormal, Synced, "switchover triggered: set read-only label to true on all Redis pods")

		if nodeInfo.MasterLinkStatus != "up" {
			r.recorder.Eventf(instance, corev1.EventTypeNormal, Synced, "switchover triggered: replica node [%s] link status is '%s'", address, nodeInfo.MasterLinkStatus)
		}

		if nodeInfo.MasterSyncInProgress {
			r.recorder.Eventf(instance, corev1.EventTypeNormal, Synced, "switchover triggered: replica node [%s] master sync still in progress, waiting until completion", address)
		}

		r.recorder.Eventf(instance, corev1.EventTypeNormal, Synced, "switchover triggered: waiting for replica node [%s] to reach offset gap '0'", address)

		if err := admin.WaitReplicaCatchUp(address, 0, redisSwitchoverCatchUpTimeout); err != nil {
			return fmt.Errorf("switchover aborted: %v", err)
		}
		r.recorder.Eventf(instance, corev1.EventTypeNormal, Synced, "switchover triggered: replica node [%s] offset lag reached 0", address)

		if err := admin.ReplicaOfNoOne(address); err != nil {
			return err
		}

		r.recorder.Eventf(instance, corev1.EventTypeNormal, Synced, "execute REPLICAOF NO ONE on [%s] successfully", address)

	}

	return nil
}

func (r *ReconcileRedisReplication) ensureReplicaNode(syncCtx *syncContext,
	replicationInfo *redisutil.ReplicationInfo,
	replica *composev1alpha1.RedisReplicaNode) error {

	instance := syncCtx.instance
	admin := syncCtx.admin

	address := net.JoinHostPort(replica.Host, strconv.Itoa(replica.Port))
	nodeInfo, ok := replicationInfo.Nodes[address]
	if !ok {
		return fmt.Errorf("cannot retrieve status from replica Redis instance [%s]", address)
	}

	sourceNodeHost := instance.Spec.Source.Host
	sourceNodePort := strconv.Itoa(instance.Spec.Source.Port)

	if replica.Isolated {
		// isolate the replica node
		if nodeInfo.Role == redisutil.RedisReplicaRole {
			if err := admin.ReplicaOfNoOne(address); err != nil {
				return err
			}

			r.recorder.Eventf(instance, corev1.EventTypeNormal, Synced, "isolation triggered: replica [%s] marked as isolated, set replica of none completed", address)
		}
	} else {
		if nodeInfo.Role == redisutil.RedisSourceRole ||
			(nodeInfo.Role == redisutil.RedisReplicaRole &&
				(nodeInfo.SourceHost != sourceNodeHost || nodeInfo.SourcePort != sourceNodePort)) || nodeInfo.MasterLinkStatus != "up" {
			if err := admin.ReplicaOfSource(address, sourceNodeHost, sourceNodePort); err != nil {
				return err
			}
			r.recorder.Eventf(instance, corev1.EventTypeNormal, Synced, "configure replication on [%s] successfully", address)
		}
	}
	return nil
}

func (r *ReconcileRedisReplication) ensureAllPodReadOnly(syncCtx *syncContext) error {
	instance := syncCtx.instance
	var errs []error

	//ensure source pod labels
	if err := r.ensurePodLabels(syncCtx, instance.Spec.Source.Name, "true", false); err != nil {
		errs = append(errs, err)
	}

	//ensure replica pod labels
	for _, replica := range instance.Spec.Replica {
		if err := r.ensurePodLabels(syncCtx, replica.Name, "true", false); err != nil {
			errs = append(errs, err)
		}
	}

	return errors.Join(errs...)
}

func (r *ReconcileRedisReplication) ensureService(syncCtx *syncContext, serviceName, isReadOnly string, isExisted bool) error {
	instance := syncCtx.instance
	ctx := syncCtx.ctx

	ownerReference := composev1alpha1.DefaultRedisReplicationOwnerReferences(instance)

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
					Port:       int32(instance.Spec.Source.Port),
					TargetPort: intstr.IntOrString{IntVal: int32(instance.Spec.Source.Port)},
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

func (r *ReconcileRedisReplication) ensurePodLabels(syncCtx *syncContext, podName, isReadOnly string, preserveLabels bool) error {
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
func (r *ReconcileRedisReplication) setLabelsOnPod(pod *corev1.Pod, instanceName, isReadOnly string) (bool, string) {
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
func (r *ReconcileRedisReplication) removeLabelsFromPod(pod *corev1.Pod, instanceName string) (bool, string) {
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

func (r *ReconcileRedisReplication) ensureSentinelPodLabels(syncCtx *syncContext) error {
	ctx := syncCtx.ctx
	instance := syncCtx.instance

	// Determine the label value based on source node count
	hostLabelValue := ""
	portLabelValue := ""
	sourceNode := instance.Spec.Source

	if sourceNodeStatus := instance.Status.Topology[sourceNode.Name]; sourceNodeStatus.Role == composev1alpha1.RedisReplicationNodeRoleSource && sourceNodeStatus.Status == composev1alpha1.NodeStatusOK {
		hostLabelValue = sourceNodeStatus.AnnounceHost
		portLabelValue = strconv.Itoa(sourceNodeStatus.AnnouncePort)
	}

	var errs []error

	// Iterate through sentinel pod names
	for _, sentinelPodName := range instance.Spec.Sentinel {
		foundPod := &corev1.Pod{}
		if err := r.client.Get(ctx, types.NamespacedName{
			Name:      sentinelPodName,
			Namespace: instance.Namespace,
		}, foundPod); err != nil {
			errs = append(errs, fmt.Errorf("failed to fetch sentinel pod [%s]: %v", sentinelPodName, err))
			continue
		}

		// Check if labels map exists, create if not
		if foundPod.Labels == nil {
			foundPod.Labels = make(map[string]string)
		}

		var needUpdate bool
		// Check if the sentinel source label already has the correct value
		if currentHostLabelValue, ok := foundPod.Labels[sourceHostKey]; !ok || currentHostLabelValue != hostLabelValue {
			foundPod.Labels[sourceHostKey] = hostLabelValue
			needUpdate = true
		}

		if currentPortLabelValue, ok := foundPod.Labels[sourcePortKey]; !ok || currentPortLabelValue != portLabelValue {
			foundPod.Labels[sourcePortKey] = portLabelValue
			needUpdate = true
		}

		// Update pod
		if needUpdate {
			if err := r.client.Update(ctx, foundPod); err != nil {
				errs = append(errs, fmt.Errorf("failed to update sentinel pod [%s]: %v", sentinelPodName, err))
				continue
			}
			r.recorder.Eventf(instance, corev1.EventTypeNormal, Synced, "sentinel pod [%s] update source label successfully", sentinelPodName)
		}

	}

	return errors.Join(errs...)
}
