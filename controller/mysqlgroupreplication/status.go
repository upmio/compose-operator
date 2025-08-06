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
	"context"
	"fmt"
	"net"
	"reflect"
	"strconv"

	"github.com/go-logr/logr"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	composev1alpha1 "github.com/upmio/compose-operator/api/v1alpha1"
	"github.com/upmio/compose-operator/pkg/mysqlutil"
	"github.com/upmio/compose-operator/pkg/utils"
)

func (r *ReconcileMysqlGroupReplication) updateInstanceIfNeed(instance *composev1alpha1.MysqlGroupReplication,
	oldStatus *composev1alpha1.MysqlGroupReplicationStatus,
	reqLogger logr.Logger) {

	if compareStatus(&instance.Status, oldStatus, reqLogger) {
		if err := r.client.Status().Update(context.TODO(), instance); err != nil {
			reqLogger.Error(err, "failed to update mysql group replication status")
		}
	}
}

func compareStatus(new, old *composev1alpha1.MysqlGroupReplicationStatus, reqLogger logr.Logger) bool {

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

func compareNodes(nodeA, nodeB *composev1alpha1.MysqlGroupReplicationNode, reqLogger logr.Logger) bool {
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

	if utils.CompareStringValue("Node.MemberState", nodeA.MemberState, nodeB.MemberState, reqLogger) {
		reqLogger.Info(fmt.Sprintf("found status.Topology[Node].MemberState changed: the old one is %s, new one is %s", nodeB.MemberState, nodeA.MemberState))
		return true
	}

	if nodeA.ReadOnly != nodeB.ReadOnly {
		reqLogger.Info(fmt.Sprintf("found status.Topology[Node].ReadOnly changed: the old one is %v, new one is %v", nodeB.ReadOnly, nodeA.ReadOnly))
		return true
	}

	if nodeA.SuperReadOnly != nodeB.SuperReadOnly {
		reqLogger.Info(fmt.Sprintf("found status.Topology[Node].SuperReadOnly changed: the old one is %v, new one is %v", nodeB.SuperReadOnly, nodeA.SuperReadOnly))
		return true
	}

	if utils.CompareStringValue("Node.GtidExecuted", nodeA.GtidExecuted, nodeB.GtidExecuted, reqLogger) {
		reqLogger.Info(fmt.Sprintf("found status.Topology[Node].GtidExecuted changed: the old one is %v, new one is %v", nodeB.GtidExecuted, nodeA.GtidExecuted))
		return true
	}

	return false
}

func buildDefaultTopologyStatus(instance *composev1alpha1.MysqlGroupReplication) composev1alpha1.MysqlGroupReplicationStatus {
	status := composev1alpha1.MysqlGroupReplicationStatus{}
	status.Topology = make(composev1alpha1.MysqlGroupReplicationTopology)

	status.Conditions = instance.Status.Conditions

	for _, node := range instance.Spec.Member {
		status.Topology[node.Name] = &composev1alpha1.MysqlGroupReplicationNode{
			Host:          node.Host,
			Port:          node.Port,
			Role:          composev1alpha1.MysqlGroupReplicationNodeRoleNone,
			Status:        composev1alpha1.NodeStatusKO,
			GtidExecuted:  "",
			MemberState:   mysqlutil.MysqlUnreachableState,
			ReadOnly:      false,
			SuperReadOnly: false,
		}
	}

	return status
}

func generateTopologyStatusByReplicationInfo(info *mysqlutil.GroupInfos, instance *composev1alpha1.MysqlGroupReplication) {
	allNodeReady := true
	for _, node := range instance.Spec.Member {
		addr := net.JoinHostPort(node.Host, strconv.Itoa(node.Port))
		if nodeInfo, ok := info.Infos[addr]; ok {
			instance.Status.Topology[node.Name].Role = nodeInfo.GetRole()
			instance.Status.Topology[node.Name].Status = composev1alpha1.NodeStatusOK
			instance.Status.Topology[node.Name].GtidExecuted = nodeInfo.GtidExecuted
			instance.Status.Topology[node.Name].MemberState = nodeInfo.State

			instance.Status.Topology[node.Name].ReadOnly = nodeInfo.ReadOnly
			instance.Status.Topology[node.Name].SuperReadOnly = nodeInfo.SuperReadOnly
		} else {
			allNodeReady = false
		}
	}

	if info.Status == mysqlutil.GroupInfoConsistent && allNodeReady {
		instance.Status.Ready = true
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
