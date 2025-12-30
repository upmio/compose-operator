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
	"fmt"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/testcontainers/testcontainers-go/modules/postgres"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	logf "sigs.k8s.io/controller-runtime/pkg/log"

	composev1alpha1 "github.com/upmio/compose-operator/api/v1alpha1"
	"github.com/upmio/compose-operator/pkg/testutil"
	"github.com/upmio/compose-operator/pkg/utils"
)

var _ = Describe("PostgresReplication Controller", func() {
	When("with default settings", func() {
		var (
			sourceHost, replicaHost           string
			sourcePort, replicaPort           int
			username                          = "replication"
			password                          = "18c6!@nkBNK9P!*d8&1Iq2Qt"
			sourceContainer, replicaContainer *postgres.PostgresContainer
			err                               error
		)

		BeforeEach(func() {
			By("creating source postgres container")
			sourceContainer, err = testutil.CreatePostgresqlContainer(ctx, username, password)
			Expect(err).ShouldNot(HaveOccurred())

			sourceHost, err = sourceContainer.Host(ctx)
			Expect(err).ShouldNot(HaveOccurred())

			sourceMappedPort, err := sourceContainer.MappedPort(ctx, "5432")
			Expect(err).ShouldNot(HaveOccurred())
			sourcePort = sourceMappedPort.Int()

			By("creating replica postgres container")
			replicaContainer, err = testutil.CreatePostgresqlContainer(ctx, username, password)
			Expect(err).ShouldNot(HaveOccurred())

			replicaHost, err = replicaContainer.Host(ctx)
			Expect(err).ShouldNot(HaveOccurred())

			replicaMappedPort, err := replicaContainer.MappedPort(ctx, "5432")
			Expect(err).ShouldNot(HaveOccurred())
			replicaPort = replicaMappedPort.Int()

		})

		AfterEach(func() {
			By("clearing source postgres container")
			Expect(sourceContainer.Terminate(ctx)).To(Succeed())

			By("clearing replica postgres container")
			Expect(replicaContainer.Terminate(ctx)).To(Succeed())
		})

		Context("creates a postgres replication sample", func() {
			const (
				resourceName = "postgres-replication-sample"
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
				By("clearing source postgres pod")
				if err := k8sClient.Get(ctx, types.NamespacedName{
					Name:      fmt.Sprintf("%s-0", resourceName),
					Namespace: namespace,
				}, pod); err == nil {
					logf.Log.Info("clear dependent source postgres pod success")
					Expect(k8sClient.Delete(ctx, pod)).To(Succeed())
				}

				By("clearing replica postgres pod")
				if err := k8sClient.Get(ctx, types.NamespacedName{
					Name:      fmt.Sprintf("%s-1", resourceName),
					Namespace: namespace,
				}, pod); err == nil {
					logf.Log.Info("clear dependent replica postgres pod success")
					Expect(k8sClient.Delete(ctx, pod)).To(Succeed())
				}
			})

			BeforeEach(func() {

				By("creating secret")

				encryptPwd, err := utils.AES_CTR_Encrypt([]byte(password), "bec62eddcb834ece8488c88263a5f248")
				Expect(err).ShouldNot(HaveOccurred())
				secret := &corev1.Secret{
					ObjectMeta: metav1.ObjectMeta{
						Name:      resourceName,
						Namespace: namespace,
					},
					Data: map[string][]byte{
						"replication": []byte(encryptPwd),
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

				By("creating source postgres pod")
				pod := &corev1.Pod{
					ObjectMeta: metav1.ObjectMeta{
						Name:      fmt.Sprintf("%s-0", resourceName),
						Namespace: namespace,
						Labels:    make(map[string]string),
					},
					Spec: corev1.PodSpec{
						Containers: []corev1.Container{
							{
								Name:  "postgres",
								Image: "docker.io/postgres:16-alpine",
								Ports: []corev1.ContainerPort{
									{
										Name:          "tcp",
										ContainerPort: 5432,
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
					logf.Log.Info("create dependent source postgres pod success")
					Expect(k8sClient.Create(ctx, pod)).To(Succeed())
				} else {
					Expect(err).To(Not(nil))
				}

				By("creating replica postgres pod")
				pod = &corev1.Pod{
					ObjectMeta: metav1.ObjectMeta{
						Name:      fmt.Sprintf("%s-1", resourceName),
						Namespace: namespace,
						Labels:    make(map[string]string),
					},
					Spec: corev1.PodSpec{
						Containers: []corev1.Container{
							{
								Name:  "postgres",
								Image: "docker.io/postgres:16-alpine",
								Ports: []corev1.ContainerPort{
									{
										Name:          "tcp",
										ContainerPort: 5432,
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
					logf.Log.Info("create dependent replica postgres pod success")
					Expect(k8sClient.Create(ctx, pod)).To(Succeed())
				} else {
					Expect(err).To(Not(nil))
				}
			})

			It("should successfully reconcile the resource", func() {
				By("creating the custom resource postgres replication sample")
				instance := &composev1alpha1.PostgresReplication{
					ObjectMeta: metav1.ObjectMeta{
						Name:      resourceName,
						Namespace: namespace,
					},
					Spec: composev1alpha1.PostgresReplicationSpec{
						Mode: composev1alpha1.PostgresRplSync,
						Secret: composev1alpha1.PostgresReplicationSecret{
							Name:        resourceName,
							Postgresql:  username,
							Replication: username,
						},
						AESSecret: &composev1alpha1.AESSecret{
							Name: "AESSecret",
							Key:  "AESKey",
						},
						Primary: &composev1alpha1.CommonNode{
							Name: fmt.Sprintf("%s-0", resourceName),
							Host: sourceHost,
							Port: sourcePort,
						},
						Service: &composev1alpha1.Service{Type: composev1alpha1.ServiceTypeClusterIP},
						Standby: composev1alpha1.ReplicaNodes{
							{
								CommonNode: composev1alpha1.CommonNode{
									Name: fmt.Sprintf("%s-1", resourceName),
									Host: replicaHost,
									Port: replicaPort,
								},
								Isolated: false,
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

					logf.Log.Info("wait postgres replication sample meet the check condition")
					if node1, ok := instance.Status.Topology[fmt.Sprintf("%s-0", resourceName)]; ok {
						if node2, ok := instance.Status.Topology[fmt.Sprintf("%s-1", resourceName)]; ok {
							if node1.Ready && !node2.Ready {
								break
							}
						}
					}

					time.Sleep(time.Second * 5)
				}

				//test switchover
				instance.Spec.Primary = &composev1alpha1.CommonNode{
					Name: fmt.Sprintf("%s-1", resourceName),
					Host: replicaHost,
					Port: sourcePort,
				}
				instance.Spec.Standby = composev1alpha1.ReplicaNodes{
					{
						CommonNode: composev1alpha1.CommonNode{
							Name: fmt.Sprintf("%s-0", resourceName),
							Host: sourceHost,
							Port: sourcePort,
						},
						Isolated: false,
					},
				}

				Expect(k8sClient.Update(ctx, instance)).To(Succeed())

				for i := 0; i < 10; i++ {
					err := k8sClient.Get(ctx, types.NamespacedName{
						Name:      resourceName,
						Namespace: namespace,
					}, instance)
					Expect(err).NotTo(HaveOccurred())

					logf.Log.Info("wait postgres replication sample meet the check condition")
					if node1, ok := instance.Status.Topology[fmt.Sprintf("%s-0", resourceName)]; ok {
						if node2, ok := instance.Status.Topology[fmt.Sprintf("%s-1", resourceName)]; ok {
							if !node1.Ready && !node2.Ready {
								break
							}
						}
					}

					time.Sleep(time.Second * 5)
				}

				By("clearing the custom resource postgres replication sample")
				Expect(k8sClient.Delete(ctx, instance)).To(Succeed())
			})
		})

	})
})
