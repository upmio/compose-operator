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
	"reflect"
	"strconv"

	"github.com/go-logr/logr"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	composev1alpha1 "github.com/upmio/compose-operator/api/v1alpha1"
	"github.com/upmio/compose-operator/pkg/redisutil"
	"github.com/upmio/compose-operator/pkg/utils"
)

func (r *ReconcileRedisCluster) updateInstanceIfNeed(ctx context.Context, instance *composev1alpha1.RedisCluster,
	oldStatus *composev1alpha1.RedisClusterStatus,
	reqLogger logr.Logger) {

	if compareStatus(&instance.Status, oldStatus, reqLogger) {
		if err := r.client.Status().Update(ctx, instance); err != nil {
			reqLogger.Error(err, "failed to update redis cluster status")
		}
	}
}

func compareStatus(new, old *composev1alpha1.RedisClusterStatus, reqLogger logr.Logger) bool {
	if old.NumberOfShard != new.NumberOfShard {
		reqLogger.Info(fmt.Sprintf("found status.NumberOfShard changed: the old one is %d, new one is %d", old.NumberOfShard, new.NumberOfShard))
		return true
	}

	if old.Ready != new.Ready {
		reqLogger.Info(fmt.Sprintf("found status.Ready changed: the old one is %v, new one is %v", old.Ready, new.Ready))
		return true
	}

	if len(old.Topology) != len(new.Topology) {
		reqLogger.Info(fmt.Sprintf("found the length of status.Topology changed: the old one is %d, new one is %d", len(old.Topology), len(new.Topology)))
		return true
	}

	for nodeName, nodeA := range new.Topology {
		if nodeB, ok := old.Topology[nodeName]; !ok {
			reqLogger.Info(fmt.Sprintf("found node %s in new status.Topology but not in old status.Topology", nodeName))
			return true
		} else if compareNodes(nodeA, nodeB, reqLogger) {
			return true
		}
	}

	if !reflect.DeepEqual(new.Conditions, old.Conditions) {
		reqLogger.Info(fmt.Sprintf("found status.Conditions changed: the old one is %#v, new one is %#v", old.Conditions, new.Conditions))
		return true
	}

	return false
}

func compareNodes(nodeA, nodeB *composev1alpha1.RedisClusterNode, reqLogger logr.Logger) bool {
	if utils.CompareStringValue("Node.Host", nodeA.Host, nodeB.Host, reqLogger) {
		reqLogger.Info(fmt.Sprintf("found status.Topology[Node].Host changed: the old one is %v, new one is %v", nodeB.Host, nodeA.Host))
		return true
	}

	if utils.CompareInt32("Node.Port", int32(nodeA.Port), int32(nodeB.Port), reqLogger) {
		reqLogger.Info(fmt.Sprintf("found status.Topology[Node].Port changed: the old one is %d, new one is %d", nodeB.Port, nodeA.Port))
		return true
	}

	if utils.CompareStringValue("Node.Role", string(nodeA.Role), string(nodeB.Role), reqLogger) {
		reqLogger.Info(fmt.Sprintf("found status.Topology[Node].Role changed: the old one is %s, new one is %s", nodeB.Role, nodeA.Role))
		return true
	}

	if utils.CompareStringValue("Node.Status", string(nodeA.Status), string(nodeB.Status), reqLogger) {
		reqLogger.Info(fmt.Sprintf("found status.Topology[Node].Status changed: the old one is %s, new one is %s", nodeB.Status, nodeA.Status))
		return true
	}

	if utils.CompareStringValue("Node.ID", nodeA.ID, nodeB.ID, reqLogger) {
		reqLogger.Info(fmt.Sprintf("found status.Topology[Node].ID changed: the old one is %s, new one is %s", nodeB.ID, nodeA.ID))
		return true
	}

	if utils.CompareStringValue("Node.MasterRef", nodeA.MasterRef, nodeB.MasterRef, reqLogger) {
		reqLogger.Info(fmt.Sprintf("found status.Topology[Node].MasterRef changed: the old one is %s, new one is %s", nodeB.MasterRef, nodeA.MasterRef))
		return true
	}

	if utils.CompareStringValue("Node.Shard", nodeA.Shard, nodeB.Shard, reqLogger) {
		reqLogger.Info(fmt.Sprintf("found status.Topology[Node].Shard changed: the old one is %s, new one is %s", nodeB.Shard, nodeA.Shard))
		return true
	}

	sizeSlotsA := 0
	sizeSlotsB := 0
	if nodeA.Slots != nil {
		sizeSlotsA = len(nodeA.Slots)
	}
	if nodeB.Slots != nil {
		sizeSlotsB = len(nodeB.Slots)
	}
	if sizeSlotsA != sizeSlotsB {
		reqLogger.Info(fmt.Sprintf("found the length of status.Topology[Node].Slots changed: the old one is %d, new one is %d", sizeSlotsB, sizeSlotsA))
		return true
	}

	if !reflect.DeepEqual(nodeA.Slots, nodeB.Slots) {
		reqLogger.Info("found status.Topology[Node].Slots changed")
		return true
	}

	return false
}

func buildDefaultTopologyStatus(instance *composev1alpha1.RedisCluster) composev1alpha1.RedisClusterStatus {
	status := composev1alpha1.RedisClusterStatus{}
	status.Topology = make(composev1alpha1.RedisClusterTopology)
	status.Conditions = instance.Status.Conditions
	status.ClusterSynced = instance.Status.ClusterSynced
	status.ClusterJoined = instance.Status.ClusterJoined

	for shardName, shard := range instance.Spec.Members {
		for _, node := range shard {
			status.Topology[node.Name] = &composev1alpha1.RedisClusterNode{
				ID:        "",
				Role:      composev1alpha1.RedisClusterNodeRoleNone,
				Status:    composev1alpha1.NodeStatusKO,
				Host:      node.Host,
				Port:      node.Port,
				Slots:     nil,
				MasterRef: "",
				Shard:     shardName,
				Ready:     false,
			}
		}
	}

	return status
}

func generateTopologyStatusByReplicationInfo(info *redisutil.ClusterInfos, instance *composev1alpha1.RedisCluster) {
	if info.Status == redisutil.ClusterInfosConsistent {
		instance.Status.Ready = true
	}

	instance.Status.NumberOfShard = len(instance.Spec.Members)

	for _, shard := range instance.Spec.Members {
		for _, node := range shard {
			ipaddr, err := net.LookupHost(node.Host)
			if len(ipaddr) == 0 || err != nil {
				instance.Status.Ready = false
				continue
			}
			addr := net.JoinHostPort(ipaddr[0], strconv.Itoa(node.Port))
			nodeInfo, ok := info.Infos[addr]
			if !ok {
				instance.Status.Ready = false
				continue
			}
			instance.Status.Topology[node.Name].Status = composev1alpha1.NodeStatusOK
			instance.Status.Topology[node.Name].Ready = true
			instance.Status.Topology[node.Name].ID = nodeInfo.Node.ID
			instance.Status.Topology[node.Name].Role = nodeInfo.Node.GetRole()
			instance.Status.Topology[node.Name].MasterRef = nodeInfo.Node.MasterReferent
			if len(nodeInfo.Node.Slots) > 0 {
				slots := redisutil.SlotRangesFromSlots(nodeInfo.Node.Slots)
				for _, slot := range slots {
					instance.Status.Topology[node.Name].Slots = append(instance.Status.Topology[node.Name].Slots, slot.String())
				}
			}

		}

	}

}

// newSucceedSyncTopologyCondition creates a condition when sync topology succeed.
func newSucceedSyncTopologyCondition() metav1.Condition {
	return metav1.Condition{
		Type:    composev1alpha1.ConditionTypeTopologyReady,
		Status:  metav1.ConditionTrue,
		Message: "Successfully sync topology",
		Reason:  SyncTopologySucceed,
	}
}

// newFailedSyncTopologyCondition creates a condition when sync topology failed.
func newFailedSyncTopologyCondition(err error) metav1.Condition {
	return metav1.Condition{
		Type:    composev1alpha1.ConditionTypeTopologyReady,
		Status:  metav1.ConditionFalse,
		Message: err.Error(),
		Reason:  SyncTopologyFailed,
	}
}

func setSuccessClusterJoined(status *composev1alpha1.RedisClusterStatus) {
	status.ClusterJoined = true
}

func setSuccessClusterSynced(status *composev1alpha1.RedisClusterStatus) {
	status.ClusterSynced = true
}
