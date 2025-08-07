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
	"fmt"
	"net"
	"strconv"
	"time"

	corev1 "k8s.io/api/core/v1"

	"github.com/upmio/compose-operator/pkg/redisutil"
	"github.com/upmio/compose-operator/pkg/utils"
)

func (r *ReconcileRedisCluster) handleRedisClusterInstance(syncCtx *syncContext) error {
	admin := syncCtx.admin
	reqLogger := syncCtx.reqLogger
	instance := syncCtx.instance

	clusterInfos, err := admin.GetClusterInfos()
	reqLogger.Info(fmt.Sprintf("current cluster status is %s", clusterInfos.Status))
	if err != nil {
		if clusterInfos.Status == redisutil.ClusterInfosPartial {
			return fmt.Errorf("failed to get cluster infos, found cluster status partial, %v", err)
		}
	}

	if !instance.Status.ClusterJoined {
		err = r.waitForClusterJoin(syncCtx, clusterInfos)
		if err != nil {
			return err
		}
		r.recorder.Event(instance, corev1.EventTypeNormal, Synced, "wait for cluster join successfully")
		setSuccessClusterJoined(&instance.Status)
	}

	if !instance.Status.ClusterSynced {
		if err := r.attachingSlavesToMaster(syncCtx, clusterInfos); err != nil {
			return fmt.Errorf("failed to attach slave to master: %v", err)
		}
		r.recorder.Event(instance, corev1.EventTypeNormal, Synced, "attach slave to master successfully")

		if err := r.allocSlots(syncCtx); err != nil {
			return fmt.Errorf("failed to allocate slots: %v", err)
		}
		r.recorder.Event(instance, corev1.EventTypeNormal, Synced, "allocate slots successfully")

		setSuccessClusterSynced(&instance.Status)
	}

	return nil
}

func (r *ReconcileRedisCluster) waitForClusterJoin(syncCtx *syncContext, clusterInfos *redisutil.ClusterInfos) error {
	if _, err := syncCtx.admin.GetClusterInfos(); err == nil {
		return nil
	}

	var firstNode *redisutil.ClusterNode
	for _, nodeInfo := range clusterInfos.Infos {
		firstNode = nodeInfo.Node
		break
	}
	syncCtx.reqLogger.Info(">>> sending CLUSTER MEET messages to join the cluster")
	err := syncCtx.admin.AttachNodeToCluster(firstNode.IPPort())
	if err != nil {
		return err
	}
	// Give two second for the join to start, in order to avoid that
	// waiting for cluster join will find all the nodes agree about
	// the config as they are still empty with unassigned slots.
	time.Sleep(2 * time.Second)

	_, err = syncCtx.admin.GetClusterInfos()
	if err != nil {
		return fmt.Errorf("failed to wait for cluster join: %v", err)
	}

	return nil
}

// allocSlots used to allocSlots to their masters
func (r *ReconcileRedisCluster) allocSlots(ctx *syncContext) error {
	mastersNum := len(ctx.instance.Spec.Members)
	clusterHashSlots := int(ctx.admin.GetHashMaxSlot() + 1)
	//ctx.reqLogger.Info(fmt.Sprintf("cluster hash max slot %d", clusterHashSlots))
	slotsPerNode := float64(clusterHashSlots) / float64(mastersNum)
	first := 0
	cursor := 0.0
	index := 0
	for _, shard := range ctx.instance.Spec.Members {
		node := shard[0]
		last := utils.Round(cursor + slotsPerNode - 1)
		if last > clusterHashSlots || index == mastersNum-1 {
			last = clusterHashSlots - 1
		}

		if last < first {
			last = first
		}

		slots := redisutil.BuildSlotSlice(redisutil.Slot(first), redisutil.Slot(last))
		ctx.reqLogger.Info(fmt.Sprintf("node %s slot slice range %v-%v", node.Name, slots[0], slots[len(slots)-1]))

		first = last + 1
		cursor += slotsPerNode

		ipaddr, err := net.LookupHost(node.Host)
		if err != nil {
			return err
		}

		addr := net.JoinHostPort(ipaddr[0], strconv.Itoa(node.Port))
		if err := ctx.admin.AddSlots(addr, slots); err != nil {
			return err
		}
	}

	return nil
}

// attachingSlavesToMaster used to attach slaves to their masters
func (r *ReconcileRedisCluster) attachingSlavesToMaster(ctx *syncContext, clusterInfos *redisutil.ClusterInfos) error {
	for _, shard := range ctx.instance.Spec.Members {
		var masterNode *redisutil.ClusterNode
	ATTACH:
		for index, node := range shard {
			ipaddr, err := net.LookupHost(node.Host)
			if err != nil {
				return err
			}

			addr := net.JoinHostPort(ipaddr[0], strconv.Itoa(node.Port))
			nodeInfo, ok := clusterInfos.Infos[addr]
			if !ok {
				ctx.reqLogger.Error(err, fmt.Sprintf("unable fo found the Cluster.ClusterNode:%s", addr))
				return err
			}

			if index == 0 {
				masterNode = nodeInfo.Node
				continue ATTACH
			}

			slaveNode := clusterInfos.Infos[addr].Node
			ctx.reqLogger.Info(fmt.Sprintf("attaching node %s to master %s", slaveNode.ID, masterNode.ID))
			err = ctx.admin.AttachSlaveToMaster(slaveNode, masterNode.ID)

			if err != nil {
				ctx.reqLogger.Error(err, fmt.Sprintf("attaching node %s to master %s", slaveNode.ID, masterNode.ID))
				return err
			}
		}
	}

	return nil
}
