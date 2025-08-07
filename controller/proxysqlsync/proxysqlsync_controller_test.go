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
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"

	//corev1 "k8s.io/api/core/v1"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"

	composev1alpha1 "github.com/upmio/compose-operator/api/v1alpha1"
)

var _ = Describe("ProxysqlSyncController", func() {
	Context("Testing proxysql sync controller", func() {
		var (
			instance *composev1alpha1.ProxysqlSync
			secret   *corev1.Secret
		)

		const (
			resourceName = "test-proxysql-sync"
			namespace    = "default"
		)

		ctx := context.Background()

		typeNamespacedName := types.NamespacedName{
			Name:      resourceName,
			Namespace: namespace, // TODO(user):Modify as needed
		}

		//secret := &corev1.Secret{}
		BeforeEach(func() {
			secret = &corev1.Secret{
				ObjectMeta: metav1.ObjectMeta{
					Name:      resourceName,
					Namespace: "default",
				},
				Data: map[string][]byte{
					"mysql":    []byte("/ru+KsOJgjj+JZS11HRh1IDFsQILgnyoqn16XqyoKoo="),
					"proxysql": []byte("/ru+KsOJgjj+JZS11HRh1IDFsQILgnyoqn16XqyoKoo="),
				},
			}

			instance = &composev1alpha1.ProxysqlSync{
				ObjectMeta: metav1.ObjectMeta{
					Name:      resourceName,
					Namespace: namespace,
				},
				Spec: composev1alpha1.ProxysqlSyncSpec{
					MysqlReplication: "test-mysql-replication",
					Secret: composev1alpha1.ProxysqlSyncSecret{
						Name:     resourceName,
						Mysql:    "mysql",
						Proxysql: "proxysql",
					},
					Proxysql: composev1alpha1.CommonNodes{
						{
							Name: "node02",
							Host: "node02",
							Port: 3306,
						},
						{
							Name: "node03",
							Host: "node03",
							Port: 3306,
						},
					},
					Rule: &composev1alpha1.Rule{
						Filter:  []string{"mysql"},
						Pattern: "",
					},
				},
			}
		})

		AfterEach(func() {
			err := k8sClient.Get(ctx, typeNamespacedName, instance)
			Expect(err).NotTo(HaveOccurred())

			By("Cleanup the specific resource instance ProxysqlSync")
			Expect(k8sClient.Delete(ctx, instance)).To(Succeed())
		})

		It("should successfully reconcile the resource", func() {
			By("creating the secret for the Kind ProxysqlSync")

			if err := k8sClient.Get(ctx, typeNamespacedName, secret); err != nil && errors.IsNotFound(err) {
				Expect(k8sClient.Create(ctx, secret)).To(Succeed())
			} else {
				Expect(err).To(Not(nil))
			}

			By("creating the custom resource for the Kind ProxysqlSync")

			if err := k8sClient.Get(ctx, typeNamespacedName, instance); err != nil && errors.IsNotFound(err) {
				Expect(k8sClient.Create(ctx, instance)).To(Succeed())
			} else {
				Expect(err).To(Not(nil))
			}
		})
	})
})
