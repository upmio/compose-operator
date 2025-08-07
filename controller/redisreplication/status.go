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

func (r *ReconcileRedisReplication) updateInstanceIfNeed(instance *composev1alpha1.RedisReplication,
	oldStatus *composev1alpha1.RedisReplicationStatus,
	reqLogger logr.Logger) {

	if compareStatus(&instance.Status, oldStatus, reqLogger) {
		if err := r.client.Status().Update(context.TODO(), instance); err != nil {
			reqLogger.Error(err, "failed to update redis replication status")
		}
	}
}

func compareStatus(new, old *composev1alpha1.RedisReplicationStatus, reqLogger logr.Logger) bool {
	if utils.CompareStringValue("ReadOnlyService", old.ReadOnlyService, new.ReadOnlyService, reqLogger) {
		reqLogger.Info(fmt.Sprintf("found status.ReadOnlyService changed: the old one is %s, new one is %s", old.ReadOnlyService, new.ReadOnlyService))
		return true
	}

	if utils.CompareStringValue("ReadWriteService", old.ReadWriteService, new.ReadWriteService, reqLogger) {
		reqLogger.Info(fmt.Sprintf("found status.ReadWriteService changed: the old one is %s, new one is %s", old.ReadWriteService, new.ReadWriteService))
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

func compareNodes(nodeA, nodeB *composev1alpha1.RedisReplicationNode, reqLogger logr.Logger) bool {
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

	if nodeA.Ready != nodeB.Ready {
		reqLogger.Info(fmt.Sprintf("found status.Topology[Node].Ready changed: the old one is %v, new one is %v", nodeB.Ready, nodeA.Ready))
		return true
	}

	if utils.CompareStringValue("Node.SourceHost", nodeA.SourceHost, nodeB.SourceHost, reqLogger) {
		reqLogger.Info(fmt.Sprintf("found status.Topology[Node].SourceHost changed: the old one is %v, new one is %v", nodeB.SourceHost, nodeA.SourceHost))
		return true
	}

	if utils.CompareInt32("Node.SourcePort", int32(nodeA.SourcePort), int32(nodeB.SourcePort), reqLogger) {
		reqLogger.Info(fmt.Sprintf("found status.Topology[Node].SourcePort changed: the old one is %d, new one is %d", nodeB.SourcePort, nodeA.SourcePort))
		return true
	}

	return false
}

func buildDefaultTopologyStatus(instance *composev1alpha1.RedisReplication) composev1alpha1.RedisReplicationStatus {
	status := composev1alpha1.RedisReplicationStatus{}
	status.Topology = make(composev1alpha1.RedisReplicationTopology)
	status.Conditions = instance.Status.Conditions
	status.Topology[instance.Spec.Source.Name] = &composev1alpha1.RedisReplicationNode{
		Host:   instance.Spec.Source.Host,
		Port:   instance.Spec.Source.Port,
		Role:   composev1alpha1.RedisReplicationNodeRoleNone,
		Status: composev1alpha1.NodeStatusKO,
		Ready:  false,
	}

	for _, replica := range instance.Spec.Replica {
		status.Topology[replica.Name] = &composev1alpha1.RedisReplicationNode{
			Host:   replica.Host,
			Port:   replica.Port,
			Role:   composev1alpha1.RedisReplicationNodeRoleNone,
			Status: composev1alpha1.NodeStatusKO,
			Ready:  false,
		}
	}

	return status
}

func generateTopologyStatusByReplicationInfo(info *redisutil.ReplicationInfo, instance *composev1alpha1.RedisReplication) {
	isInstanceReady := true

	sourceAddr := net.JoinHostPort(instance.Spec.Source.Host, strconv.Itoa(instance.Spec.Source.Port))
	if node, ok := info.Nodes[sourceAddr]; ok {
		instance.Status.Topology[instance.Spec.Source.Name].Role = node.GetRole()
		instance.Status.Topology[instance.Spec.Source.Name].Status = composev1alpha1.NodeStatusOK
		if node.GetRole() == redisutil.RedisSourceRole {
			instance.Status.Topology[instance.Spec.Source.Name].Ready = true
		} else {
			instance.Status.Topology[instance.Spec.Source.Name].Ready = false
			isInstanceReady = false
		}
		instance.Status.Topology[instance.Spec.Source.Name].SourceHost = node.SourceHost
		instance.Status.Topology[instance.Spec.Source.Name].SourcePort = node.GetSourcePort()
	} else {
		isInstanceReady = false
	}

	for _, replica := range instance.Spec.Replica {
		addr := net.JoinHostPort(replica.Host, strconv.Itoa(replica.Port))
		if node, ok := info.Nodes[addr]; ok {
			instance.Status.Topology[replica.Name].Role = node.GetRole()
			instance.Status.Topology[replica.Name].Status = composev1alpha1.NodeStatusOK
			if node.GetRole() == redisutil.RedisReplicaRole {
				instance.Status.Topology[replica.Name].Ready = true
			} else {
				instance.Status.Topology[replica.Name].Ready = false
				isInstanceReady = false
			}
			instance.Status.Topology[replica.Name].SourceHost = node.SourceHost
			instance.Status.Topology[replica.Name].SourcePort = node.GetSourcePort()
		} else {
			isInstanceReady = false
		}
	}

	instance.Status.Ready = isInstanceReady
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

// newSucceedSyncResourceCondition creates a condition when sync resource succeed.
func newSucceedSyncResourceCondition() metav1.Condition {
	return metav1.Condition{
		Type:    composev1alpha1.ConditionTypeResourceReady,
		Status:  metav1.ConditionTrue,
		Message: "Successfully sync resource",
		Reason:  SyncResourceSucceed,
	}
}

// newFailedSyncResourceCondition creates a condition when sync resource failed.
func newFailedSyncResourceCondition(err error) metav1.Condition {
	return metav1.Condition{
		Type:    composev1alpha1.ConditionTypeResourceReady,
		Status:  metav1.ConditionFalse,
		Message: err.Error(),
		Reason:  SyncResourceFailed,
	}
}
