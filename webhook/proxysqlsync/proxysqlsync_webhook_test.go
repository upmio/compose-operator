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
	"github.com/upmio/compose-operator/api/v1alpha1"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var _ = Describe("ProxysqlSync Webhook", func() {

	var psa = proxysqlSyncAdmission{}

	Context("When creating ProxysqlSync under Defaulting Webhook", func() {
		It("Should fill in the default values for secret keys", func() {
			proxysqlSync := &v1alpha1.ProxysqlSync{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test-proxysql-sync",
					Namespace: "default",
				},
				Spec: v1alpha1.ProxysqlSyncSpec{
					Secret: v1alpha1.ProxysqlSyncSecret{
						Name: "test-secret",
						// proxysql and mysql keys are empty - should get defaults
					},
					MysqlReplication: "test-mysql-replication",
					Proxysql: []*v1alpha1.CommonNode{
						{
							Name: "proxysql-0",
							Host: "proxysql-0.proxysql.default.svc.cluster.local",
							Port: 6032,
						},
					},
				},
			}

			err := psa.Default(context.TODO(), proxysqlSync)
			Expect(err).NotTo(HaveOccurred())

			// Check defaults were set
			Expect(proxysqlSync.Spec.Secret.Proxysql).To(Equal("proxysql"))
			Expect(proxysqlSync.Spec.Secret.Mysql).To(Equal("mysql"))
		})

		It("Should not override existing values", func() {
			proxysqlSync := &v1alpha1.ProxysqlSync{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test-proxysql-sync",
					Namespace: "default",
				},
				Spec: v1alpha1.ProxysqlSyncSpec{
					Secret: v1alpha1.ProxysqlSyncSecret{
						Name:     "test-secret",
						Proxysql: "custom-proxysql",
						Mysql:    "custom-mysql",
					},
					MysqlReplication: "test-mysql-replication",
					Proxysql: []*v1alpha1.CommonNode{
						{
							Name: "proxysql-0",
							Host: "proxysql-0.proxysql.default.svc.cluster.local",
							Port: 6032,
						},
					},
				},
			}

			err := psa.Default(context.TODO(), proxysqlSync)
			Expect(err).NotTo(HaveOccurred())

			// Check existing values were preserved
			Expect(proxysqlSync.Spec.Secret.Proxysql).To(Equal("custom-proxysql"))
			Expect(proxysqlSync.Spec.Secret.Mysql).To(Equal("custom-mysql"))
		})
	})

	Context("When creating ProxysqlSync under Validating Webhook", func() {
		It("Should deny if no ProxySQL nodes are specified", func() {
			proxysqlSync := &v1alpha1.ProxysqlSync{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test-proxysql-sync",
					Namespace: "default",
				},
				Spec: v1alpha1.ProxysqlSyncSpec{
					Secret: v1alpha1.ProxysqlSyncSecret{
						Name:     "test-secret",
						Proxysql: "proxysql",
						Mysql:    "mysql",
					},
					MysqlReplication: "test-mysql-replication",
					// No ProxySQL nodes
					Proxysql: []*v1alpha1.CommonNode{},
				},
			}

			_, err := psa.ValidateCreate(context.TODO(), proxysqlSync)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("at least one ProxySQL node is required"))
		})

		It("Should deny if MysqlReplication reference is empty", func() {
			proxysqlSync := &v1alpha1.ProxysqlSync{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test-proxysql-sync",
					Namespace: "default",
				},
				Spec: v1alpha1.ProxysqlSyncSpec{
					Secret: v1alpha1.ProxysqlSyncSecret{
						Name:     "test-secret",
						Proxysql: "proxysql",
						Mysql:    "mysql",
					},
					MysqlReplication: "", // Empty MysqlReplication reference
					Proxysql: []*v1alpha1.CommonNode{
						{
							Name: "proxysql-0",
							Host: "proxysql-0.proxysql.default.svc.cluster.local",
							Port: 6032,
						},
					},
				},
			}

			_, err := psa.ValidateCreate(context.TODO(), proxysqlSync)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("MysqlReplication reference is required"))
		})

		It("Should deny if secret name is empty", func() {
			proxysqlSync := &v1alpha1.ProxysqlSync{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test-proxysql-sync",
					Namespace: "default",
				},
				Spec: v1alpha1.ProxysqlSyncSpec{
					Secret: v1alpha1.ProxysqlSyncSecret{
						Name: "", // Empty secret name
					},
					MysqlReplication: "test-mysql-replication",
					Proxysql: []*v1alpha1.CommonNode{
						{
							Name: "proxysql-0",
							Host: "proxysql-0.proxysql.default.svc.cluster.local",
							Port: 6032,
						},
					},
				},
			}

			_, err := psa.ValidateCreate(context.TODO(), proxysqlSync)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("secret name is required"))
		})

		It("Should deny if ProxySQL node names are duplicate", func() {
			proxysqlSync := &v1alpha1.ProxysqlSync{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test-proxysql-sync",
					Namespace: "default",
				},
				Spec: v1alpha1.ProxysqlSyncSpec{
					Secret: v1alpha1.ProxysqlSyncSecret{
						Name:     "test-secret",
						Proxysql: "proxysql",
						Mysql:    "mysql",
					},
					MysqlReplication: "test-mysql-replication",
					Proxysql: []*v1alpha1.CommonNode{
						{
							Name: "proxysql-duplicate",
							Host: "proxysql-0.proxysql.default.svc.cluster.local",
							Port: 6032,
						},
						{
							Name: "proxysql-duplicate", // Duplicate name
							Host: "proxysql-1.proxysql.default.svc.cluster.local",
							Port: 6032,
						},
					},
				},
			}

			_, err := psa.ValidateCreate(context.TODO(), proxysqlSync)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("ProxySQL node names must be unique"))
		})

		It("Should deny if ProxySQL node has invalid port", func() {
			proxysqlSync := &v1alpha1.ProxysqlSync{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test-proxysql-sync",
					Namespace: "default",
				},
				Spec: v1alpha1.ProxysqlSyncSpec{
					Secret: v1alpha1.ProxysqlSyncSecret{
						Name:     "test-secret",
						Proxysql: "proxysql",
						Mysql:    "mysql",
					},
					MysqlReplication: "test-mysql-replication",
					Proxysql: []*v1alpha1.CommonNode{
						{
							Name: "proxysql-0",
							Host: "proxysql-0.proxysql.default.svc.cluster.local",
							Port: 0, // Invalid port
						},
					},
				},
			}

			_, err := psa.ValidateCreate(context.TODO(), proxysqlSync)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("port must be between 1 and 65535"))
		})

		It("Should deny if ProxySQL node has empty hostname", func() {
			proxysqlSync := &v1alpha1.ProxysqlSync{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test-proxysql-sync",
					Namespace: "default",
				},
				Spec: v1alpha1.ProxysqlSyncSpec{
					Secret: v1alpha1.ProxysqlSyncSecret{
						Name:     "test-secret",
						Proxysql: "proxysql",
						Mysql:    "mysql",
					},
					MysqlReplication: "test-mysql-replication",
					Proxysql: []*v1alpha1.CommonNode{
						{
							Name: "proxysql-0",
							Host: "", // Empty host
							Port: 6032,
						},
					},
				},
			}

			_, err := psa.ValidateCreate(context.TODO(), proxysqlSync)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("proxysql node host is required"))
		})

		It("Should accept valid ProxysqlSync", func() {
			proxysqlSync := &v1alpha1.ProxysqlSync{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test-proxysql-sync",
					Namespace: "default",
				},
				Spec: v1alpha1.ProxysqlSyncSpec{
					Secret: v1alpha1.ProxysqlSyncSecret{
						Name:     "test-secret",
						Proxysql: "proxysql",
						Mysql:    "mysql",
					},
					MysqlReplication: "test-mysql-replication",
					Proxysql: []*v1alpha1.CommonNode{
						{
							Name: "proxysql-0",
							Host: "proxysql-0.proxysql.default.svc.cluster.local",
							Port: 6032,
						},
						{
							Name: "proxysql-1",
							Host: "proxysql-1.proxysql.default.svc.cluster.local",
							Port: 6032,
						},
					},
				},
			}

			warnings, err := psa.ValidateCreate(context.TODO(), proxysqlSync)
			Expect(err).NotTo(HaveOccurred())
			Expect(warnings).To(HaveLen(0))
		})
	})

	Context("When updating ProxysqlSync under Validating Webhook", func() {
		It("Should deny if secret name is changed", func() {
			oldProxysqlSync := &v1alpha1.ProxysqlSync{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test-proxysql-sync",
					Namespace: "default",
				},
				Spec: v1alpha1.ProxysqlSyncSpec{
					Secret: v1alpha1.ProxysqlSyncSecret{
						Name:     "old-secret",
						Proxysql: "proxysql",
						Mysql:    "mysql",
					},
					MysqlReplication: "test-mysql-replication",
					Proxysql: []*v1alpha1.CommonNode{
						{
							Name: "proxysql-0",
							Host: "proxysql-0.proxysql.default.svc.cluster.local",
							Port: 6032,
						},
					},
				},
			}

			newProxysqlSync := &v1alpha1.ProxysqlSync{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test-proxysql-sync",
					Namespace: "default",
				},
				Spec: v1alpha1.ProxysqlSyncSpec{
					Secret: v1alpha1.ProxysqlSyncSecret{
						Name:     "new-secret", // Changed secret name
						Proxysql: "proxysql",
						Mysql:    "mysql",
					},
					MysqlReplication: "test-mysql-replication",
					Proxysql: []*v1alpha1.CommonNode{
						{
							Name: "proxysql-0",
							Host: "proxysql-0.proxysql.default.svc.cluster.local",
							Port: 6032,
						},
					},
				},
			}

			_, err := psa.ValidateUpdate(context.TODO(), oldProxysqlSync, newProxysqlSync)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("secret reference cannot be changed after creation"))
		})

		It("Should warn when changing MysqlReplication reference", func() {
			oldProxysqlSync := &v1alpha1.ProxysqlSync{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test-proxysql-sync",
					Namespace: "default",
				},
				Spec: v1alpha1.ProxysqlSyncSpec{
					Secret: v1alpha1.ProxysqlSyncSecret{
						Name:     "test-secret",
						Proxysql: "proxysql",
						Mysql:    "mysql",
					},
					MysqlReplication: "old-mysql-replication",
					Proxysql: []*v1alpha1.CommonNode{
						{
							Name: "proxysql-0",
							Host: "proxysql-0.proxysql.default.svc.cluster.local",
							Port: 6032,
						},
					},
				},
			}

			newProxysqlSync := &v1alpha1.ProxysqlSync{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test-proxysql-sync",
					Namespace: "default",
				},
				Spec: v1alpha1.ProxysqlSyncSpec{
					Secret: v1alpha1.ProxysqlSyncSecret{
						Name:     "test-secret",
						Proxysql: "proxysql",
						Mysql:    "mysql",
					},
					MysqlReplication: "new-mysql-replication", // Changed MysqlReplication
					Proxysql: []*v1alpha1.CommonNode{
						{
							Name: "proxysql-0",
							Host: "proxysql-0.proxysql.default.svc.cluster.local",
							Port: 6032,
						},
					},
				},
			}

			warnings, err := psa.ValidateUpdate(context.TODO(), oldProxysqlSync, newProxysqlSync)
			Expect(err).NotTo(HaveOccurred())
			Expect(warnings).To(HaveLen(1))
			Expect(warnings[0]).To(ContainSubstring("Changing MysqlReplication reference may cause synchronization issues"))
		})
	})

	Context("When deleting ProxysqlSync under Validating Webhook", func() {
		It("Should provide warning about cleanup", func() {
			proxysqlSync := &v1alpha1.ProxysqlSync{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test-proxysql-sync",
					Namespace: "default",
				},
				Spec: v1alpha1.ProxysqlSyncSpec{
					Secret: v1alpha1.ProxysqlSyncSecret{
						Name:     "test-secret",
						Proxysql: "proxysql",
						Mysql:    "mysql",
					},
					MysqlReplication: "test-mysql-replication",
					Proxysql: []*v1alpha1.CommonNode{
						{
							Name: "proxysql-0",
							Host: "proxysql-0.proxysql.default.svc.cluster.local",
							Port: 6032,
						},
					},
				},
			}

			warnings, err := psa.ValidateDelete(context.TODO(), proxysqlSync)
			Expect(err).NotTo(HaveOccurred())
			Expect(warnings).To(HaveLen(1))
			Expect(warnings[0]).To(ContainSubstring("Deleting ProxysqlSync will stop managing"))
		})
	})

	Context("Hostname validation", func() {
		It("Should accept valid IP addresses", func() {
			proxysqlSync := &v1alpha1.ProxysqlSync{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test-proxysql-sync",
					Namespace: "default",
				},
				Spec: v1alpha1.ProxysqlSyncSpec{
					Secret: v1alpha1.ProxysqlSyncSecret{
						Name:     "test-secret",
						Proxysql: "proxysql",
						Mysql:    "mysql",
					},
					MysqlReplication: "test-mysql-replication",
					Proxysql: []*v1alpha1.CommonNode{
						{
							Name: "proxysql-0",
							Host: "192.168.1.100",
							Port: 6032,
						},
						{
							Name: "proxysql-1",
							Host: "10.0.0.1",
							Port: 6032,
						},
					},
				},
			}

			_, err := psa.ValidateCreate(context.TODO(), proxysqlSync)
			Expect(err).NotTo(HaveOccurred())
		})

		It("Should accept valid hostnames", func() {
			proxysqlSync := &v1alpha1.ProxysqlSync{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test-proxysql-sync",
					Namespace: "default",
				},
				Spec: v1alpha1.ProxysqlSyncSpec{
					Secret: v1alpha1.ProxysqlSyncSecret{
						Name:     "test-secret",
						Proxysql: "proxysql",
						Mysql:    "mysql",
					},
					MysqlReplication: "test-mysql-replication",
					Proxysql: []*v1alpha1.CommonNode{
						{
							Name: "proxysql-0",
							Host: "proxysql-0.proxysql.default.svc.cluster.local",
							Port: 6032,
						},
						{
							Name: "proxysql-1",
							Host: "example.com",
							Port: 6032,
						},
						{
							Name: "proxysql-2",
							Host: "proxysql-server",
							Port: 6032,
						},
					},
				},
			}

			_, err := psa.ValidateCreate(context.TODO(), proxysqlSync)
			Expect(err).NotTo(HaveOccurred())
		})
	})
})
