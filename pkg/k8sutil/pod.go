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

package k8sutil

import (
	"context"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// IPodControl defines the interface that uses to create, update, and delete Pods.
type IPodControl interface {
	// CreatePod creates a Pod in a DistributedRedisCluster.
	CreatePod(*corev1.Pod) error
	// UpdatePod updates a Pod in a DistributedRedisCluster.
	UpdatePod(*corev1.Pod) error
	// DeletePod deletes a Pod in a DistributedRedisCluster.
	DeletePod(*corev1.Pod) error
	DeletePodByName(namespace, name string) error
	// GetPod get Pod in a DistributedRedisCluster.
	GetPod(namespace, name string) (*corev1.Pod, error)
}

type PodController struct {
	client client.Client
}

// NewPodController creates a concrete implementation of the
// IPodControl.
func NewPodController(client client.Client) IPodControl {
	return &PodController{client: client}
}

// CreatePod implement the IPodControl.Interface.
func (p *PodController) CreatePod(pod *corev1.Pod) error {
	return p.client.Create(context.TODO(), pod)
}

// UpdatePod implement the IPodControl.Interface.
func (p *PodController) UpdatePod(pod *corev1.Pod) error {
	return p.client.Update(context.TODO(), pod)
}

// DeletePod implement the IPodControl.Interface.
func (p *PodController) DeletePod(pod *corev1.Pod) error {
	return p.client.Delete(context.TODO(), pod)
}

// DeletePod implement the IPodControl.Interface.
func (p *PodController) DeletePodByName(namespace, name string) error {
	pod, err := p.GetPod(namespace, name)
	if err != nil {
		return err
	}
	return p.client.Delete(context.TODO(), pod)
}

// GetPod implement the IPodControl.Interface.
func (p *PodController) GetPod(namespace, name string) (*corev1.Pod, error) {
	pod := &corev1.Pod{}
	err := p.client.Get(context.TODO(), types.NamespacedName{
		Name:      name,
		Namespace: namespace,
	}, pod)
	return pod, err
}

// ---------------------------------------------------
// copy from  k8s.io/kubernetes/pkg/api/v1/pod

// IsPodReady returns true if a pod is ready; false otherwise.
func IsPodReady(pod *corev1.Pod) bool {
	return IsPodReadyConditionTrue(pod.Status)
}

// IsPodReadyConditionTrue returns true if a pod is ready; false otherwise.
func IsPodReadyConditionTrue(status corev1.PodStatus) bool {
	condition := GetPodReadyCondition(status)
	return condition != nil && condition.Status == corev1.ConditionTrue
}

// GetPodReadyCondition extracts the pod ready condition from the given status and returns that.
// Returns nil if the condition is not present.
func GetPodReadyCondition(status corev1.PodStatus) *corev1.PodCondition {
	_, condition := GetPodCondition(&status, corev1.PodReady)
	return condition
}

// GetPodCondition extracts the provided condition from the given status and returns that.
// Returns nil and -1 if the condition is not present, and the index of the located condition.
func GetPodCondition(status *corev1.PodStatus, conditionType corev1.PodConditionType) (int, *corev1.PodCondition) {
	if status == nil {
		return -1, nil
	}
	return GetPodConditionFromList(status.Conditions, conditionType)
}

// GetPodConditionFromList extracts the provided condition from the given list of condition and
// returns the index of the condition and the condition. Returns -1 and nil if the condition is not present.
func GetPodConditionFromList(conditions []corev1.PodCondition, conditionType corev1.PodConditionType) (int, *corev1.PodCondition) {
	if conditions == nil {
		return -1, nil
	}
	for i := range conditions {
		if conditions[i].Type == conditionType {
			return i, &conditions[i]
		}
	}
	return -1, nil
}
