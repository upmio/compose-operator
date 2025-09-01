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
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/testcontainers/testcontainers-go/modules/redis"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	logf "sigs.k8s.io/controller-runtime/pkg/log"

	composev1alpha1 "github.com/upmio/compose-operator/api/v1alpha1"
	"github.com/upmio/compose-operator/pkg/testutil"
)

var _ = Describe("RedisCluster Controller", func() {
	When("with default settings", func() {
		var (
			firstHost, secondHost, thirdHost                string
			firstPort, secondPort, thirdPort                int
			firstContainer, secondContainer, thirdContainer *redis.RedisContainer
			err                                             error
		)

		BeforeEach(func() {
			By("creating first redis container")
			firstContainer, err = testutil.CreateRedisContainer(ctx)
			Expect(err).ShouldNot(HaveOccurred())

			firstHost, err = firstContainer.Host(ctx)
			Expect(err).ShouldNot(HaveOccurred())

			firstMappedPort, err := firstContainer.MappedPort(ctx, "6379")
			Expect(err).ShouldNot(HaveOccurred())
			firstPort = firstMappedPort.Int()

			By("creating second redis container")
			secondContainer, err = testutil.CreateRedisContainer(ctx)
			Expect(err).ShouldNot(HaveOccurred())

			secondHost, err = secondContainer.Host(ctx)
			Expect(err).ShouldNot(HaveOccurred())

			secondMappedPort, err := secondContainer.MappedPort(ctx, "6379")
			Expect(err).ShouldNot(HaveOccurred())
			secondPort = secondMappedPort.Int()

			By("creating third redis container")
			thirdContainer, err = testutil.CreateRedisContainer(ctx)
			Expect(err).ShouldNot(HaveOccurred())

			thirdHost, err = thirdContainer.Host(ctx)
			Expect(err).ShouldNot(HaveOccurred())

			thirdMappedPort, err := thirdContainer.MappedPort(ctx, "6379")
			Expect(err).ShouldNot(HaveOccurred())
			thirdPort = thirdMappedPort.Int()
		})

		AfterEach(func() {
			By("clearing first redis container")
			Expect(firstContainer.Terminate(ctx)).To(Succeed())

			By("clearing second redis container")
			Expect(secondContainer.Terminate(ctx)).To(Succeed())

			By("clearing third redis container")
			Expect(thirdContainer.Terminate(ctx)).To(Succeed())
		})

		Context("creates a redis cluster sample", func() {
			const (
				resourceName = "redis-cluster-sample"
				namespace    = "default"
			)

			AfterEach(func() {
				secret := &corev1.Secret{}
				By("clearing secret")
				if err := k8sClient.Get(ctx, types.NamespacedName{
					Name:      resourceName,
					Namespace: namespace,
				}, secret); err == nil {
					logf.Log.Info("clear dependent secret success")
					Expect(k8sClient.Delete(ctx, secret)).To(Succeed())
				}

				pod := &corev1.Pod{}
				By("clearing first redis pod")
				if err := k8sClient.Get(ctx, types.NamespacedName{
					Name:      fmt.Sprintf("%s-0", resourceName),
					Namespace: namespace,
				}, pod); err == nil {
					logf.Log.Info("clear dependent first redis pod success")
					Expect(k8sClient.Delete(ctx, pod)).To(Succeed())
				}

				By("clearing second redis pod")
				if err := k8sClient.Get(ctx, types.NamespacedName{
					Name:      fmt.Sprintf("%s-1", resourceName),
					Namespace: namespace,
				}, pod); err == nil {
					logf.Log.Info("clear dependent second redis pod success")
					Expect(k8sClient.Delete(ctx, pod)).To(Succeed())
				}

				By("clearing third redis pod")
				if err := k8sClient.Get(ctx, types.NamespacedName{
					Name:      fmt.Sprintf("%s-0", resourceName),
					Namespace: namespace,
				}, pod); err == nil {
					logf.Log.Info("clear dependent third redis pod success")
					Expect(k8sClient.Delete(ctx, pod)).To(Succeed())
				}
			})

			BeforeEach(func() {
				By("creating secret")
				Expect(err).ShouldNot(HaveOccurred())
				secret := &corev1.Secret{
					ObjectMeta: metav1.ObjectMeta{
						Name:      resourceName,
						Namespace: namespace,
					},
					Data: map[string][]byte{
						"redis": []byte(""),
					},
				}
				if err := k8sClient.Get(ctx, types.NamespacedName{
					Name:      resourceName,
					Namespace: namespace,
				}, secret); err != nil && errors.IsNotFound(err) {
					logf.Log.Info("create dependent secret success")
					Expect(k8sClient.Create(ctx, secret)).To(Succeed())
				} else {
					Expect(err).To(Not(nil))
				}

				By("creating first redis pod")
				pod := &corev1.Pod{
					ObjectMeta: metav1.ObjectMeta{
						Name:      fmt.Sprintf("%s-0", resourceName),
						Namespace: namespace,
						Labels:    make(map[string]string),
					},
					Spec: corev1.PodSpec{
						Containers: []corev1.Container{
							{
								Name:  "redis",
								Image: "redis:7",
								Ports: []corev1.ContainerPort{
									{
										Name:          "tcp",
										ContainerPort: 6379,
										Protocol:      corev1.ProtocolTCP,
									},
								},
							},
						},
					},
				}
				if err := k8sClient.Get(ctx, types.NamespacedName{
					Name:      fmt.Sprintf("%s-0", resourceName),
					Namespace: namespace,
				}, pod); err != nil && errors.IsNotFound(err) {
					logf.Log.Info("create dependent first redis pod success")
					Expect(k8sClient.Create(ctx, pod)).To(Succeed())
				} else {
					Expect(err).To(Not(nil))
				}

				By("creating second mysql pod")
				pod = &corev1.Pod{
					ObjectMeta: metav1.ObjectMeta{
						Name:      fmt.Sprintf("%s-1", resourceName),
						Namespace: namespace,
						Labels:    make(map[string]string),
					},
					Spec: corev1.PodSpec{
						Containers: []corev1.Container{
							{
								Name:  "redis",
								Image: "redis:7",
								Ports: []corev1.ContainerPort{
									{
										Name:          "tcp",
										ContainerPort: 6379,
										Protocol:      corev1.ProtocolTCP,
									},
								},
							},
						},
					},
				}
				if err := k8sClient.Get(ctx, types.NamespacedName{
					Name:      fmt.Sprintf("%s-1", resourceName),
					Namespace: namespace,
				}, pod); err != nil && errors.IsNotFound(err) {
					logf.Log.Info("create dependent second redis pod success")
					Expect(k8sClient.Create(ctx, pod)).To(Succeed())
				} else {
					Expect(err).To(Not(nil))
				}

				By("creating third mysql pod")
				pod = &corev1.Pod{
					ObjectMeta: metav1.ObjectMeta{
						Name:      fmt.Sprintf("%s-2", resourceName),
						Namespace: namespace,
						Labels:    make(map[string]string),
					},
					Spec: corev1.PodSpec{
						Containers: []corev1.Container{
							{
								Name:  "redis",
								Image: "redis:7",
								Ports: []corev1.ContainerPort{
									{
										Name:          "tcp",
										ContainerPort: 6379,
										Protocol:      corev1.ProtocolTCP,
									},
								},
							},
						},
					},
				}
				if err := k8sClient.Get(ctx, types.NamespacedName{
					Name:      fmt.Sprintf("%s-2", resourceName),
					Namespace: namespace,
				}, pod); err != nil && errors.IsNotFound(err) {
					logf.Log.Info("create dependent third redis pod success")
					Expect(k8sClient.Create(ctx, pod)).To(Succeed())
				} else {
					Expect(err).To(Not(nil))
				}
			})

			It("should successfully reconcile the resource", func() {
				By("creating the custom resource redis cluster sample")
				instance := &composev1alpha1.RedisCluster{
					ObjectMeta: metav1.ObjectMeta{
						Name:      resourceName,
						Namespace: namespace,
					},
					Spec: composev1alpha1.RedisClusterSpec{
						Secret: composev1alpha1.RedisClusterSecret{
							Name:  resourceName,
							Redis: "redis",
						},
						AESSecret: &composev1alpha1.AESSecret{
							Name: "AESSecret",
							Key:  "AESKey",
						},
						Members: map[string]composev1alpha1.CommonNodes{
							"shard01": {
								&composev1alpha1.CommonNode{
									Name: fmt.Sprintf("%s-0", resourceName),
									Host: firstHost,
									Port: firstPort,
								},
							},
							"shard02": {
								&composev1alpha1.CommonNode{
									Name: fmt.Sprintf("%s-1", resourceName),
									Host: secondHost,
									Port: secondPort,
								},
							},
							"shard03": {
								&composev1alpha1.CommonNode{
									Name: fmt.Sprintf("%s-2", resourceName),
									Host: thirdHost,
									Port: thirdPort,
								},
							},
						},
					},
				}

				if err := k8sClient.Get(ctx, types.NamespacedName{
					Name:      resourceName,
					Namespace: namespace,
				}, instance); err != nil && errors.IsNotFound(err) {
					Expect(k8sClient.Create(ctx, instance)).To(Succeed())
				} else {
					Expect(err).To(Not(nil))
				}

				for i := 0; i < 10; i++ {
					err := k8sClient.Get(ctx, types.NamespacedName{
						Name:      resourceName,
						Namespace: namespace,
					}, instance)
					Expect(err).NotTo(HaveOccurred())

					//logf.Log.Info("wait redis cluster sample meet the check condition")
					//if node1, ok := instance.Status.Topology[fmt.Sprintf("%s-0", resourceName)]; ok {
					//	if node2, ok := instance.Status.Topology[fmt.Sprintf("%s-1", resourceName)]; ok {
					//		if node1.Ready && !node2.Ready {
					//			break
					//		}
					//	}
					//}
					//
					time.Sleep(time.Second * 5)
				}

				By("clearing the custom resource redis cluster sample")
				Expect(k8sClient.Delete(ctx, instance)).To(Succeed())
			})
		})

	})
})
