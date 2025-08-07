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
	"context"
	"fmt"
	"github.com/upmio/compose-operator/pkg/proxysqlutil"
	"net"
	"reflect"
	"strconv"

	"github.com/go-logr/logr"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	composev1alpha1 "github.com/upmio/compose-operator/api/v1alpha1"
)

func (r *ReconcileProxysqlSync) updateInstanceIfNeed(instance *composev1alpha1.ProxysqlSync,
	oldStatus *composev1alpha1.ProxysqlSyncStatus,
	reqLogger logr.Logger) {

	if compareStatus(&instance.Status, oldStatus, reqLogger) {
		if err := r.client.Status().Update(context.TODO(), instance); err != nil {
			reqLogger.Error(err, "failed to update proxysql sync status")
		}
	}
}

func compareStatus(new, old *composev1alpha1.ProxysqlSyncStatus, reqLogger logr.Logger) bool {

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

func compareNodes(nodeA, nodeB *composev1alpha1.ProxysqlSyncNode, reqLogger logr.Logger) bool {
	if nodeA.Synced != nodeB.Synced {
		reqLogger.Info(fmt.Sprintf("found status.Ready changed: the old one is %v, new one is %v", nodeB.Synced, nodeA.Synced))
		return true
	}

	if compareUsers(nodeA.Users, nodeB.Users, reqLogger) {
		return true
	}

	return false
}

func compareUsers(new, old []string, reqLogger logr.Logger) bool {
	oldUser := make(map[string]string)
	newUser := make(map[string]string)

	for _, username := range new {
		newUser[username] = username
	}

	for _, username := range old {
		oldUser[username] = username
	}

	if !reflect.DeepEqual(oldUser, newUser) {
		reqLogger.Info("found status.Topology[Node].Users changed")
		return true
	}

	return false
}

func buildDefaultTopologyStatus(instance *composev1alpha1.ProxysqlSync) composev1alpha1.ProxysqlSyncStatus {
	status := composev1alpha1.ProxysqlSyncStatus{}
	status.Topology = make(composev1alpha1.ProxysqlSyncTopology)
	status.Conditions = instance.Status.Conditions

	for _, proxysql := range instance.Spec.Proxysql {
		status.Topology[proxysql.Name] = &composev1alpha1.ProxysqlSyncNode{
			Users: make([]string, 0),
		}
	}

	return status
}

func generateTopologyStatusByReplicationInfo(admin proxysqlutil.IAdmin, instance *composev1alpha1.ProxysqlSync) {
	for _, node := range instance.Spec.Proxysql {
		address := net.JoinHostPort(node.Host, strconv.Itoa(node.Port))

		users, err := admin.GetRuntimeMysqlUsers(context.TODO(), address)
		if err != nil {
			instance.Status.Ready = false
			instance.Status.Topology[node.Name].Synced = false
		}

		instance.Status.Topology[node.Name].Users = users.TransferToAPI()
	}
}

// newSucceedSyncServerCondition creates a condition when sync server succeed.
func newSucceedSyncServerCondition() metav1.Condition {
	return metav1.Condition{
		Type:    composev1alpha1.ConditionTypeServerReady,
		Status:  metav1.ConditionTrue,
		Message: "Successfully sync server",
		Reason:  SyncServerSucceed,
	}
}

// newFailedSyncServerCondition creates a condition when sync server failed.
func newFailedSyncServerCondition(err error) metav1.Condition {
	return metav1.Condition{
		Type:    composev1alpha1.ConditionTypeServerReady,
		Status:  metav1.ConditionFalse,
		Message: err.Error(),
		Reason:  SyncServerFailed,
	}
}

// newSucceedSyncUserCondition creates a condition when sync user succeed.
func newSucceedSyncUserCondition() metav1.Condition {
	return metav1.Condition{
		Type:    composev1alpha1.ConditionTypeUserReady,
		Status:  metav1.ConditionTrue,
		Message: "Successfully sync user",
		Reason:  SyncUserSucceed,
	}
}

// newFailedSyncUserCondition creates a condition when sync user failed.
func newFailedSyncUserCondition(err error) metav1.Condition {
	return metav1.Condition{
		Type:    composev1alpha1.ConditionTypeUserReady,
		Status:  metav1.ConditionFalse,
		Message: err.Error(),
		Reason:  SyncUserFailed,
	}
}
