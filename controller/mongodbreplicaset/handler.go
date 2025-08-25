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

package mongodbreplicaset

import (
	"errors"
	"fmt"
	"k8s.io/apimachinery/pkg/types"
	"net"
	"strconv"

	corev1 "k8s.io/api/core/v1"

	"github.com/upmio/compose-operator/pkg/mongoutil"
)

func (r *ReconcileMongoDBReplicaset) handleMongoDBReplicasetInstance(syncCtx *syncContext) error {
	instance := syncCtx.instance
	ctx := syncCtx.ctx
	admin := syncCtx.admin

	// Get replica set information from available nodes
	replicaSetInfo := admin.GetReplicaSetInfos(ctx, len(instance.Spec.Member))

	// Step 2: Handle different cluster states
	switch replicaSetInfo.Status {
	case mongoutil.ReplicaSetInfoUnset:
		return r.handleUninitializedCluster(syncCtx)
	case mongoutil.ReplicaSetInfoPartial:
		return r.handlePartialCluster(syncCtx, replicaSetInfo)
	}

	return nil
}

// handleUninitializedCluster handles cluster initialization with intelligent node selection
func (r *ReconcileMongoDBReplicaset) handleUninitializedCluster(syncCtx *syncContext) error {
	instance := syncCtx.instance
	ctx := syncCtx.ctx
	admin := syncCtx.admin

	// Build replica set configuration
	var members []mongoutil.ReplicaSetConfigMember
	for i, node := range instance.Spec.Member {
		addr := net.JoinHostPort(node.Host, strconv.Itoa(node.Port))
		member := mongoutil.ReplicaSetConfigMember{
			ID:       i,
			Host:     addr,
			Priority: 1.0, // Use default priority for all nodes
			Votes:    1,
		}
		members = append(members, member)
	}

	initNode := net.JoinHostPort(instance.Spec.Member[0].Host, strconv.Itoa(instance.Spec.Member[0].Port))
	if err := admin.InitiateReplicaSet(ctx, initNode, instance.Spec.ReplicaSetName, members); err != nil {
		return fmt.Errorf("failed to initialize replica set on %s: %v", initNode, err)
	}

	r.recorder.Eventf(instance, corev1.EventTypeNormal, Synced,
		"initialized replica set on [%s] successfully", initNode)

	return nil
}

// handlePartialCluster handles adding missing members to the replica set
func (r *ReconcileMongoDBReplicaset) handlePartialCluster(syncCtx *syncContext, replicaSetInfo *mongoutil.ReplicaSetInfos) error {

	var errs []error

	//TODO addnode removenode logica

	return errors.Join(errs...)
}

func (r *ReconcileMongoDBReplicaset) handleResources(syncCtx *syncContext) error {
	instance := syncCtx.instance

	var errs []error
	// set pod label
	for _, node := range instance.Spec.Member {
		if err := r.ensurePodLabels(syncCtx, node.Name); err != nil {
			errs = append(errs, err)
		}
	}

	return errors.Join(errs...)
}

func (r *ReconcileMongoDBReplicaset) ensurePodLabels(syncCtx *syncContext, podName string) error {
	ctx := syncCtx.ctx
	instance := syncCtx.instance
	namespace := instance.Namespace
	instanceName := instance.Name

	foundPod := &corev1.Pod{}
	if err := r.client.Get(ctx, types.NamespacedName{
		Name:      podName,
		Namespace: namespace,
	}, foundPod); err != nil {
		return fmt.Errorf("failed to fetch pod [%s]: %v", podName, err)
	}

	// Check and update default label
	if instanceValue, ok := foundPod.Labels[defaultKey]; !ok || instanceValue != instanceName {
		foundPod.Labels[defaultKey] = instanceName

		if err := r.client.Update(ctx, foundPod); err != nil {
			return fmt.Errorf("failed to update pod [%s]: %v", podName, err)
		}
		r.recorder.Eventf(instance, corev1.EventTypeNormal, Synced, "pod [%s] update labels '%s' successfully", podName, defaultKey)
	}

	return nil
}
