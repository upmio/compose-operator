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
	"fmt"
	"net"
	"reflect"
	"strconv"
	"strings"

	"github.com/go-logr/logr"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	composev1alpha1 "github.com/upmio/compose-operator/api/v1alpha1"
	"github.com/upmio/compose-operator/pkg/postgresutil"
	"github.com/upmio/compose-operator/pkg/utils"
)

func (r *ReconcilePostgresReplication) updateInstanceIfNeed(ctx context.Context, instance *composev1alpha1.PostgresReplication,
	oldStatus *composev1alpha1.PostgresReplicationStatus,
	reqLogger logr.Logger) {

	if compareStatus(&instance.Status, oldStatus, reqLogger) {
		if err := r.client.Status().Update(ctx, instance); err != nil {
			reqLogger.Error(err, "failed to update postgres replication status")
		}
	}
}

func compareStatus(new, old *composev1alpha1.PostgresReplicationStatus, reqLogger logr.Logger) bool {
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

func compareNodes(nodeA, nodeB *composev1alpha1.PostgresReplicationNode, reqLogger logr.Logger) bool {
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

	if utils.CompareIntPoint("Node.WalDiff", nodeA.WalDiff, nodeB.WalDiff, reqLogger) {
		reqLogger.Info("found status.Topology[Node].WalDiff changed")
		return true
	}

	if nodeA.Ready != nodeB.Ready {
		reqLogger.Info(fmt.Sprintf("found status.Topology[Node].Ready changed: the old one is %v, new one is %v", nodeB.Ready, nodeA.Ready))
		return true
	}

	return false
}

func buildDefaultTopologyStatus(instance *composev1alpha1.PostgresReplication) composev1alpha1.PostgresReplicationStatus {
	status := composev1alpha1.PostgresReplicationStatus{}
	status.Topology = make(composev1alpha1.PostgresReplicationTopology)
	status.Conditions = instance.Status.Conditions
	status.Topology[instance.Spec.Primary.Name] = &composev1alpha1.PostgresReplicationNode{
		Host:   instance.Spec.Primary.Host,
		Port:   instance.Spec.Primary.Port,
		Role:   composev1alpha1.PostgresReplicationRoleNone,
		Status: composev1alpha1.NodeStatusKO,
		Ready:  false,
	}

	for _, stanby := range instance.Spec.Standby {
		status.Topology[stanby.Name] = &composev1alpha1.PostgresReplicationNode{
			Host:   stanby.Host,
			Port:   stanby.Port,
			Role:   composev1alpha1.PostgresReplicationRoleNone,
			Status: composev1alpha1.NodeStatusKO,
			Ready:  false,
		}
	}

	status.ObservedGeneration = instance.Generation

	return status
}

func generateTopologyStatusByReplicationInfo(info *postgresutil.ReplicationInfo, instance *composev1alpha1.PostgresReplication) {
	if info == nil || instance == nil {
		return
	}

	// Ensure Topology is initialized
	if instance.Status.Topology == nil {
		instance.Status.Topology = make(composev1alpha1.PostgresReplicationTopology)
	}

	// Process primary node
	primaryAddr := net.JoinHostPort(instance.Spec.Primary.Host, strconv.Itoa(instance.Spec.Primary.Port))

	// Process primary node normally
	if node, ok := info.Nodes[primaryAddr]; ok {
		instance.Status.Topology[instance.Spec.Primary.Name].Role = node.GetRole()
		instance.Status.Topology[instance.Spec.Primary.Name].Status = composev1alpha1.NodeStatusOK

		if node.GetRole() == postgresutil.PostgresPrimaryRole {
			switch instance.Spec.Mode {
			case composev1alpha1.PostgresRplAsync:
				if node.ReplicationMode == "async" {
					instance.Status.Topology[instance.Spec.Primary.Name].Ready = true
				}
			case composev1alpha1.PostgresRplSync:
				if node.ReplicationMode == "quorum" {
					instance.Status.Topology[instance.Spec.Primary.Name].Ready = true
				}
			}
		}
	}

	for _, standby := range instance.Spec.Standby {
		addr := net.JoinHostPort(standby.Host, strconv.Itoa(standby.Port))
		if standbyNode, ok := info.Nodes[addr]; ok {

			instance.Status.Topology[standby.Name].Role = standbyNode.GetRole()
			instance.Status.Topology[standby.Name].Status = composev1alpha1.NodeStatusOK
			instance.Status.Topology[standby.Name].WalDiff = &standbyNode.WalDiff

			if node, ok := info.Nodes[primaryAddr]; standbyNode.GetRole() == postgresutil.PostgresStandbyRole && ok {
				instance.Status.Topology[standby.Name].Ready = false
				for _, standbyName := range node.ReplicationStat {
					if strings.ReplaceAll(standby.Name, "-", "_") == standbyName {
						instance.Status.Topology[standby.Name].Ready = true
					}
				}
				if !instance.Status.Topology[standby.Name].Ready && standby.Isolated {
					delete(instance.Status.Topology, standby.Name)
				}
			}
		}
	}

	isInstanceReady := true

	for _, node := range instance.Status.Topology {
		isInstanceReady = isInstanceReady && node.Ready
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
