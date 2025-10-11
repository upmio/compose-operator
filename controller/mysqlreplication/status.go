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

package mysqlreplication

import (
	"context"
	"fmt"
	"net"
	"reflect"
	"strconv"

	"github.com/go-logr/logr"
	composev1alpha1 "github.com/upmio/compose-operator/api/v1alpha1"
	"github.com/upmio/compose-operator/pkg/mysqlutil"
	"github.com/upmio/compose-operator/pkg/utils"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func (r *ReconcileMysqlReplication) updateInstanceIfNeed(ctx context.Context, instance *composev1alpha1.MysqlReplication,
	oldStatus *composev1alpha1.MysqlReplicationStatus,
	reqLogger logr.Logger) {

	if compareStatus(&instance.Status, oldStatus, reqLogger) {
		if err := r.client.Status().Update(ctx, instance); err != nil {
			reqLogger.Error(err, "failed to update mysql replication status")
		}
	}
}

func compareStatus(new, old *composev1alpha1.MysqlReplicationStatus, reqLogger logr.Logger) bool {
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

	if old.ObservedGeneration != new.ObservedGeneration {
		reqLogger.Info(fmt.Sprintf("found status.ObservedGeneration changed: the old one is %d, new one is %d", old.ObservedGeneration, new.ObservedGeneration))
		return true
	}

	return false
}

func compareNodes(nodeA, nodeB *composev1alpha1.MysqlReplicationNode, reqLogger logr.Logger) bool {
	if utils.CompareStringValue("ReplicationNode.Host", nodeA.Host, nodeB.Host, reqLogger) {
		reqLogger.Info(fmt.Sprintf("found status.Topology[ReplicationNode].Host changed: the old one is %v, new one is %v", nodeB.Host, nodeA.Host))
		return true
	}

	if utils.CompareInt32("ReplicationNode.Port", int32(nodeA.Port), int32(nodeB.Port), reqLogger) {
		reqLogger.Info(fmt.Sprintf("found status.Topology[ReplicationNode].Port changed: the old one is %d, new one is %d", nodeB.Port, nodeA.Port))
		return true
	}

	if utils.CompareStringValue("ReplicationNode.Role", string(nodeA.Role), string(nodeB.Role), reqLogger) {
		reqLogger.Info(fmt.Sprintf("found status.Topology[ReplicationNode].Role changed: the old one is %s, new one is %s", nodeB.Role, nodeA.Role))
		return true
	}

	if utils.CompareStringValue("ReplicationNode.Status", string(nodeA.Status), string(nodeB.Status), reqLogger) {
		reqLogger.Info(fmt.Sprintf("found status.Topology[ReplicationNode].Status changed: the old one is %s, new one is %s", nodeB.Status, nodeA.Status))
		return true
	}

	if nodeA.Ready != nodeB.Ready {
		reqLogger.Info(fmt.Sprintf("found status.Topology[ReplicationNode].Ready changed: the old one is %v, new one is %v", nodeB.Ready, nodeA.Ready))
		return true
	}

	if nodeA.ReadOnly != nodeB.ReadOnly {
		reqLogger.Info(fmt.Sprintf("found status.Topology[ReplicationNode].ReadOnly changed: the old one is %v, new one is %v", nodeB.ReadOnly, nodeA.ReadOnly))
		return true
	}

	if nodeA.SuperReadOnly != nodeB.SuperReadOnly {
		reqLogger.Info(fmt.Sprintf("found status.Topology[ReplicationNode].SuperReadOnly changed: the old one is %v, new one is %v", nodeB.SuperReadOnly, nodeA.SuperReadOnly))
		return true
	}

	if utils.CompareStringValue("ReplicationNode.SourceHost", nodeA.SourceHost, nodeB.SourceHost, reqLogger) {
		reqLogger.Info(fmt.Sprintf("found status.Topology[ReplicationNode].SourceHost changed: the old one is %v, new one is %v", nodeB.SourceHost, nodeA.SourceHost))
		return true
	}

	if utils.CompareInt32("ReplicationNode.SourcePort", int32(nodeA.SourcePort), int32(nodeB.SourcePort), reqLogger) {
		reqLogger.Info(fmt.Sprintf("found status.Topology[ReplicationNode].SourcePort changed: the old one is %d, new one is %d", nodeB.SourcePort, nodeA.SourcePort))
		return true
	}

	if utils.CompareStringValue("ReplicationNode.ReplicaIO", nodeA.ReplicaIO, nodeB.ReplicaIO, reqLogger) {
		reqLogger.Info(fmt.Sprintf("found status.Topology[ReplicationNode].ReplicaIO changed: the old one is %s, new one is %s", nodeB.ReplicaIO, nodeA.ReplicaIO))
		return true
	}

	if utils.CompareStringValue("ReplicationNode.ReplicaSQL", nodeA.ReplicaSQL, nodeB.ReplicaSQL, reqLogger) {
		reqLogger.Info(fmt.Sprintf("found status.Topology[ReplicationNode].ReplicaSQL changed: the old one is %s, new one is %s", nodeB.ReplicaSQL, nodeA.ReplicaSQL))
		return true
	}

	if utils.CompareInt32("ReplicationNode.ReadSourceLogPos", int32(nodeA.ReadSourceLogPos), int32(nodeB.ReadSourceLogPos), reqLogger) {
		reqLogger.Info(fmt.Sprintf("found status.Topology[ReplicationNode].ReadSourceLogPos changed: the old one is %d, new one is %d", nodeB.ReadSourceLogPos, nodeA.ReadSourceLogPos))
		return true
	}

	if utils.CompareInt32("ReplicationNode.ExecSourceLogPos", int32(nodeA.ExecSourceLogPos), int32(nodeB.ExecSourceLogPos), reqLogger) {
		reqLogger.Info(fmt.Sprintf("found status.Topology[ReplicationNode].ExecSourceLogPos changed: the old one is %d, new one is %d", nodeB.ExecSourceLogPos, nodeA.ExecSourceLogPos))
		return true
	}

	if utils.CompareStringValue("ReplicationNode.SourceLogFile", nodeA.SourceLogFile, nodeB.SourceLogFile, reqLogger) {
		reqLogger.Info(fmt.Sprintf("found status.Topology[ReplicationNode].SourceLogFile changed: the old one is %s, new one is %s", nodeB.SourceLogFile, nodeA.SourceLogFile))
		return true
	}

	if utils.CompareIntPoint("ReplicationNode.SecondsBehindSource", nodeA.SecondsBehindSource, nodeB.SecondsBehindSource, reqLogger) {
		reqLogger.Info(fmt.Sprintf("found status.Topology[ReplicationNode].SecondsBehindSource changed: the old one is %d, new one is %d", nodeB.SecondsBehindSource, nodeA.SecondsBehindSource))
		return true
	}

	return false
}

func buildDefaultTopologyStatus(instance *composev1alpha1.MysqlReplication) composev1alpha1.MysqlReplicationStatus {
	status := composev1alpha1.MysqlReplicationStatus{}
	status.Topology = make(composev1alpha1.MysqlReplicationTopology)
	status.Conditions = instance.Status.Conditions
	status.Topology[instance.Spec.Source.Name] = &composev1alpha1.MysqlReplicationNode{
		Host:          instance.Spec.Source.Host,
		Port:          instance.Spec.Source.Port,
		Role:          composev1alpha1.MysqlReplicationNodeRoleNone,
		Status:        composev1alpha1.NodeStatusKO,
		Ready:         false,
		ReadOnly:      false,
		SuperReadOnly: false,
	}

	for _, replica := range instance.Spec.Replica {
		status.Topology[replica.Name] = &composev1alpha1.MysqlReplicationNode{
			Host:          replica.Host,
			Port:          replica.Port,
			Role:          composev1alpha1.MysqlReplicationNodeRoleNone,
			Status:        composev1alpha1.NodeStatusKO,
			Ready:         false,
			ReadOnly:      false,
			SuperReadOnly: false,
		}
	}

	status.ObservedGeneration = instance.Generation

	return status
}

func generateTopologyStatusByReplicationInfo(info *mysqlutil.ReplicationInfo, instance *composev1alpha1.MysqlReplication) {
	if info == nil || instance == nil {
		return
	}

	// Ensure Topology is initialized
	if instance.Status.Topology == nil {
		instance.Status.Topology = make(composev1alpha1.MysqlReplicationTopology)
	}

	isInstanceReady := true

	// Process source node
	sourceAddr := net.JoinHostPort(instance.Spec.Source.Host, strconv.Itoa(instance.Spec.Source.Port))

	// Process source node normally
	if node, ok := info.Nodes[sourceAddr]; ok {
		instance.Status.Topology[instance.Spec.Source.Name].Role = node.GetRole()
		instance.Status.Topology[instance.Spec.Source.Name].Status = composev1alpha1.NodeStatusOK

		if node.GetRole() == mysqlutil.MysqlSourceRole && !node.ReadOnly && !node.SuperReadOnly {
			instance.Status.Topology[instance.Spec.Source.Name].Ready = true
		} else {
			instance.Status.Topology[instance.Spec.Source.Name].Ready = false
			isInstanceReady = false
		}

		instance.Status.Topology[instance.Spec.Source.Name].ReadOnly = node.ReadOnly
		instance.Status.Topology[instance.Spec.Source.Name].SuperReadOnly = node.SuperReadOnly

		if node.GetRole() == composev1alpha1.MysqlReplicationNodeRoleReplica {
			instance.Status.Topology[instance.Spec.Source.Name].SourceHost = node.SourceHost
			instance.Status.Topology[instance.Spec.Source.Name].SourcePort = node.GetSourcePort()
			instance.Status.Topology[instance.Spec.Source.Name].ReplicaIO = node.ReplicaIO
			instance.Status.Topology[instance.Spec.Source.Name].ReplicaSQL = node.ReplicaSQL
			instance.Status.Topology[instance.Spec.Source.Name].ReadSourceLogPos = node.GetReadSourceLogPos()
			instance.Status.Topology[instance.Spec.Source.Name].ExecSourceLogPos = node.GetExecSourceLogPos()
			instance.Status.Topology[instance.Spec.Source.Name].SourceLogFile = node.SourceLogFile
			instance.Status.Topology[instance.Spec.Source.Name].SecondsBehindSource = &node.SecondsBehindSource
		}
	} else {
		isInstanceReady = false
	}

	// Process replica nodes
	for _, replica := range instance.Spec.Replica {
		// Process replica node normally
		addr := net.JoinHostPort(replica.Host, strconv.Itoa(replica.Port))
		if node, ok := info.Nodes[addr]; ok {

			// Check if replica node is isolated
			if replica.Isolated && node.GetRole() == mysqlutil.MysqlSourceRole && node.ReadOnly && node.SuperReadOnly {
				// Remove replica node from topology if it's isolated
				delete(instance.Status.Topology, replica.Name)
			} else {

				instance.Status.Topology[replica.Name].Role = node.GetRole()
				instance.Status.Topology[replica.Name].Status = composev1alpha1.NodeStatusOK

				if node.GetRole() == mysqlutil.MysqlReplicaRole && node.ReadOnly && node.SuperReadOnly && node.ReplicaIO == "Yes" && node.ReplicaSQL == "Yes" {
					instance.Status.Topology[replica.Name].Ready = true
				} else {
					instance.Status.Topology[replica.Name].Ready = false
					isInstanceReady = false
				}

				instance.Status.Topology[replica.Name].ReadOnly = node.ReadOnly
				instance.Status.Topology[replica.Name].SuperReadOnly = node.SuperReadOnly
				instance.Status.Topology[replica.Name].SourceHost = node.SourceHost
				instance.Status.Topology[replica.Name].SourcePort = node.GetSourcePort()
				instance.Status.Topology[replica.Name].ReplicaIO = node.ReplicaIO
				instance.Status.Topology[replica.Name].ReplicaSQL = node.ReplicaSQL
				instance.Status.Topology[replica.Name].ReadSourceLogPos = node.GetReadSourceLogPos()
				instance.Status.Topology[replica.Name].ExecSourceLogPos = node.GetExecSourceLogPos()
				instance.Status.Topology[replica.Name].SourceLogFile = node.SourceLogFile
				instance.Status.Topology[replica.Name].SecondsBehindSource = &node.SecondsBehindSource
			}
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
