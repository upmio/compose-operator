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

package mysqlgroupreplication

import (
	"context"
	"fmt"
	"github.com/upmio/compose-operator/api/v1alpha1"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var _ = Describe("MysqlGroupReplication Webhook", func() {

	var mgra = mysqlGroupReplicationAdmission{}

	Context("When creating MysqlGroupReplication under Defaulting Webhook", func() {
		It("Should fill in the default values for secret keys", func() {
			mysqlGR := &v1alpha1.MysqlGroupReplication{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test-mysql-gr",
					Namespace: "default",
				},
				Spec: v1alpha1.MysqlGroupReplicationSpec{
					Secret: v1alpha1.MysqlGroupReplicationSecret{
						Name: "test-secret",
						// mysql and replication keys are empty - should get defaults
					},
					Member: []*v1alpha1.CommonNode{
						{
							Name: "mysql-0",
							Host: "mysql-0.mysql.default.svc.cluster.local",
							Port: 3306,
						},
						{
							Name: "mysql-1",
							Host: "mysql-1.mysql.default.svc.cluster.local",
							Port: 3306,
						},
						{
							Name: "mysql-2",
							Host: "mysql-2.mysql.default.svc.cluster.local",
							Port: 3306,
						},
					},
				},
			}

			err := mgra.Default(context.TODO(), mysqlGR)
			Expect(err).NotTo(HaveOccurred())

			// Check defaults were set
			Expect(mysqlGR.Spec.Secret.Mysql).To(Equal("mysql"))
			Expect(mysqlGR.Spec.Secret.Replication).To(Equal("replication"))
		})

		It("Should not override existing values", func() {
			mysqlGR := &v1alpha1.MysqlGroupReplication{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test-mysql-gr",
					Namespace: "default",
				},
				Spec: v1alpha1.MysqlGroupReplicationSpec{
					Secret: v1alpha1.MysqlGroupReplicationSecret{
						Name:        "test-secret",
						Mysql:       "custom-mysql",
						Replication: "custom-repl",
					},
					Member: []*v1alpha1.CommonNode{
						{
							Name: "mysql-0",
							Host: "mysql-0.mysql.default.svc.cluster.local",
							Port: 3306,
						},
					},
				},
			}

			err := mgra.Default(context.TODO(), mysqlGR)
			Expect(err).NotTo(HaveOccurred())

			// Check existing values were preserved
			Expect(mysqlGR.Spec.Secret.Mysql).To(Equal("custom-mysql"))
			Expect(mysqlGR.Spec.Secret.Replication).To(Equal("custom-repl"))
		})
	})

	Context("When creating MysqlGroupReplication under Validating Webhook", func() {
		It("Should deny if no members are specified", func() {
			mysqlGR := &v1alpha1.MysqlGroupReplication{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test-mysql-gr",
					Namespace: "default",
				},
				Spec: v1alpha1.MysqlGroupReplicationSpec{
					Secret: v1alpha1.MysqlGroupReplicationSecret{
						Name:        "test-secret",
						Mysql:       "mysql",
						Replication: "replication",
					},
					// No members
					Member: []*v1alpha1.CommonNode{},
				},
			}

			_, err := mgra.ValidateCreate(context.TODO(), mysqlGR)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("at least one group member is required"))
		})

		It("Should warn if less than 3 members are specified", func() {
			mysqlGR := &v1alpha1.MysqlGroupReplication{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test-mysql-gr",
					Namespace: "default",
				},
				Spec: v1alpha1.MysqlGroupReplicationSpec{
					Secret: v1alpha1.MysqlGroupReplicationSecret{
						Name:        "test-secret",
						Mysql:       "mysql",
						Replication: "replication",
					},
					Member: []*v1alpha1.CommonNode{
						{
							Name: "mysql-0",
							Host: "mysql-0.mysql.default.svc.cluster.local",
							Port: 3306,
						},
					},
				},
			}

			warnings, err := mgra.ValidateCreate(context.TODO(), mysqlGR)
			Expect(err).NotTo(HaveOccurred())
			Expect(warnings).To(HaveLen(1))
			Expect(warnings[0]).To(ContainSubstring("at least 3 members for proper fault tolerance"))
		})

		It("Should deny if more than 9 members are specified", func() {
			members := make([]*v1alpha1.CommonNode, 10)
			for i := 0; i < 10; i++ {
				members[i] = &v1alpha1.CommonNode{
					Name: fmt.Sprintf("mysql-%d", i),
					Host: fmt.Sprintf("mysql-%d.mysql.default.svc.cluster.local", i),
					Port: 3306,
				}
			}

			mysqlGR := &v1alpha1.MysqlGroupReplication{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test-mysql-gr",
					Namespace: "default",
				},
				Spec: v1alpha1.MysqlGroupReplicationSpec{
					Secret: v1alpha1.MysqlGroupReplicationSecret{
						Name:        "test-secret",
						Mysql:       "mysql",
						Replication: "replication",
					},
					Member: members,
				},
			}

			_, err := mgra.ValidateCreate(context.TODO(), mysqlGR)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("supports a maximum of 9 members"))
		})

		It("Should deny if secret name is empty", func() {
			mysqlGR := &v1alpha1.MysqlGroupReplication{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test-mysql-gr",
					Namespace: "default",
				},
				Spec: v1alpha1.MysqlGroupReplicationSpec{
					Secret: v1alpha1.MysqlGroupReplicationSecret{
						Name: "", // Empty secret name
					},
					Member: []*v1alpha1.CommonNode{
						{
							Name: "mysql-0",
							Host: "mysql-0.mysql.default.svc.cluster.local",
							Port: 3306,
						},
					},
				},
			}

			_, err := mgra.ValidateCreate(context.TODO(), mysqlGR)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("secret name is required"))
		})

		It("Should deny if member names are duplicate", func() {
			mysqlGR := &v1alpha1.MysqlGroupReplication{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test-mysql-gr",
					Namespace: "default",
				},
				Spec: v1alpha1.MysqlGroupReplicationSpec{
					Secret: v1alpha1.MysqlGroupReplicationSecret{
						Name:        "test-secret",
						Mysql:       "mysql",
						Replication: "replication",
					},
					Member: []*v1alpha1.CommonNode{
						{
							Name: "mysql-duplicate",
							Host: "mysql-0.mysql.default.svc.cluster.local",
							Port: 3306,
						},
						{
							Name: "mysql-duplicate", // Duplicate name
							Host: "mysql-1.mysql.default.svc.cluster.local",
							Port: 3306,
						},
					},
				},
			}

			_, err := mgra.ValidateCreate(context.TODO(), mysqlGR)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("group member names must be unique"))
		})

		It("Should deny if member has invalid port", func() {
			mysqlGR := &v1alpha1.MysqlGroupReplication{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test-mysql-gr",
					Namespace: "default",
				},
				Spec: v1alpha1.MysqlGroupReplicationSpec{
					Secret: v1alpha1.MysqlGroupReplicationSecret{
						Name:        "test-secret",
						Mysql:       "mysql",
						Replication: "replication",
					},
					Member: []*v1alpha1.CommonNode{
						{
							Name: "mysql-0",
							Host: "mysql-0.mysql.default.svc.cluster.local",
							Port: 0, // Invalid port
						},
					},
				},
			}

			_, err := mgra.ValidateCreate(context.TODO(), mysqlGR)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("port must be between 1 and 65535"))
		})

		It("Should deny if member has empty hostname", func() {
			mysqlGR := &v1alpha1.MysqlGroupReplication{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test-mysql-gr",
					Namespace: "default",
				},
				Spec: v1alpha1.MysqlGroupReplicationSpec{
					Secret: v1alpha1.MysqlGroupReplicationSecret{
						Name:        "test-secret",
						Mysql:       "mysql",
						Replication: "replication",
					},
					Member: []*v1alpha1.CommonNode{
						{
							Name: "mysql-0",
							Host: "", // Empty host
							Port: 3306,
						},
					},
				},
			}

			_, err := mgra.ValidateCreate(context.TODO(), mysqlGR)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("member node host is required"))
		})

		It("Should accept valid MysqlGroupReplication with 3 members", func() {
			mysqlGR := &v1alpha1.MysqlGroupReplication{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test-mysql-gr",
					Namespace: "default",
				},
				Spec: v1alpha1.MysqlGroupReplicationSpec{
					Secret: v1alpha1.MysqlGroupReplicationSecret{
						Name:        "test-secret",
						Mysql:       "mysql",
						Replication: "replication",
					},
					Member: []*v1alpha1.CommonNode{
						{
							Name: "mysql-0",
							Host: "mysql-0.mysql.default.svc.cluster.local",
							Port: 3306,
						},
						{
							Name: "mysql-1",
							Host: "mysql-1.mysql.default.svc.cluster.local",
							Port: 3306,
						},
						{
							Name: "mysql-2",
							Host: "mysql-2.mysql.default.svc.cluster.local",
							Port: 3306,
						},
					},
				},
			}

			warnings, err := mgra.ValidateCreate(context.TODO(), mysqlGR)
			Expect(err).NotTo(HaveOccurred())
			Expect(warnings).To(HaveLen(0))
		})
	})

	Context("When updating MysqlGroupReplication under Validating Webhook", func() {
		It("Should deny if secret name is changed", func() {
			oldMysqlGR := &v1alpha1.MysqlGroupReplication{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test-mysql-gr",
					Namespace: "default",
				},
				Spec: v1alpha1.MysqlGroupReplicationSpec{
					Secret: v1alpha1.MysqlGroupReplicationSecret{
						Name:        "old-secret",
						Mysql:       "mysql",
						Replication: "replication",
					},
					Member: []*v1alpha1.CommonNode{
						{
							Name: "mysql-0",
							Host: "mysql-0.mysql.default.svc.cluster.local",
							Port: 3306,
						},
					},
				},
			}

			newMysqlGR := &v1alpha1.MysqlGroupReplication{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test-mysql-gr",
					Namespace: "default",
				},
				Spec: v1alpha1.MysqlGroupReplicationSpec{
					Secret: v1alpha1.MysqlGroupReplicationSecret{
						Name:        "new-secret", // Changed secret name
						Mysql:       "mysql",
						Replication: "replication",
					},
					Member: []*v1alpha1.CommonNode{
						{
							Name: "mysql-0",
							Host: "mysql-0.mysql.default.svc.cluster.local",
							Port: 3306,
						},
					},
				},
			}

			_, err := mgra.ValidateUpdate(context.TODO(), oldMysqlGR, newMysqlGR)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("secret reference cannot be changed after creation"))
		})

		It("Should warn when changing group member count", func() {
			oldMysqlGR := &v1alpha1.MysqlGroupReplication{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test-mysql-gr",
					Namespace: "default",
				},
				Spec: v1alpha1.MysqlGroupReplicationSpec{
					Secret: v1alpha1.MysqlGroupReplicationSecret{
						Name:        "test-secret",
						Mysql:       "mysql",
						Replication: "replication",
					},
					Member: []*v1alpha1.CommonNode{
						{
							Name: "mysql-0",
							Host: "mysql-0.mysql.default.svc.cluster.local",
							Port: 3306,
						},
					},
				},
			}

			newMysqlGR := &v1alpha1.MysqlGroupReplication{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test-mysql-gr",
					Namespace: "default",
				},
				Spec: v1alpha1.MysqlGroupReplicationSpec{
					Secret: v1alpha1.MysqlGroupReplicationSecret{
						Name:        "test-secret",
						Mysql:       "mysql",
						Replication: "replication",
					},
					Member: []*v1alpha1.CommonNode{
						{
							Name: "mysql-0",
							Host: "mysql-0.mysql.default.svc.cluster.local",
							Port: 3306,
						},
						{
							Name: "mysql-1",
							Host: "mysql-1.mysql.default.svc.cluster.local",
							Port: 3306,
						},
					},
				},
			}

			warnings, err := mgra.ValidateUpdate(context.TODO(), oldMysqlGR, newMysqlGR)
			Expect(err).NotTo(HaveOccurred())
			Expect(warnings).To(HaveLen(2))
			Expect(warnings[0]).To(ContainSubstring("at least 3 members for proper fault tolerance"))
			Expect(warnings[1]).To(ContainSubstring("Changing group members may cause group reconfiguration"))
		})
	})

	Context("When deleting MysqlGroupReplication under Validating Webhook", func() {
		It("Should provide warning about cleanup", func() {
			mysqlGR := &v1alpha1.MysqlGroupReplication{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test-mysql-gr",
					Namespace: "default",
				},
				Spec: v1alpha1.MysqlGroupReplicationSpec{
					Secret: v1alpha1.MysqlGroupReplicationSecret{
						Name:        "test-secret",
						Mysql:       "mysql",
						Replication: "replication",
					},
					Member: []*v1alpha1.CommonNode{
						{
							Name: "mysql-0",
							Host: "mysql-0.mysql.default.svc.cluster.local",
							Port: 3306,
						},
					},
				},
			}

			warnings, err := mgra.ValidateDelete(context.TODO(), mysqlGR)
			Expect(err).NotTo(HaveOccurred())
			Expect(warnings).To(HaveLen(1))
			Expect(warnings[0]).To(ContainSubstring("Deleting MysqlGroupReplication will stop managing"))
		})
	})

	Context("Hostname validation", func() {
		It("Should accept valid IP addresses", func() {
			mysqlGR := &v1alpha1.MysqlGroupReplication{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test-mysql-gr",
					Namespace: "default",
				},
				Spec: v1alpha1.MysqlGroupReplicationSpec{
					Secret: v1alpha1.MysqlGroupReplicationSecret{
						Name:        "test-secret",
						Mysql:       "mysql",
						Replication: "replication",
					},
					Member: []*v1alpha1.CommonNode{
						{
							Name: "mysql-0",
							Host: "192.168.1.100",
							Port: 3306,
						},
						{
							Name: "mysql-1",
							Host: "10.0.0.1",
							Port: 3306,
						},
						{
							Name: "mysql-2",
							Host: "172.16.0.1",
							Port: 3306,
						},
					},
				},
			}

			_, err := mgra.ValidateCreate(context.TODO(), mysqlGR)
			Expect(err).NotTo(HaveOccurred())
		})

		It("Should accept valid hostnames", func() {
			mysqlGR := &v1alpha1.MysqlGroupReplication{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test-mysql-gr",
					Namespace: "default",
				},
				Spec: v1alpha1.MysqlGroupReplicationSpec{
					Secret: v1alpha1.MysqlGroupReplicationSecret{
						Name:        "test-secret",
						Mysql:       "mysql",
						Replication: "replication",
					},
					Member: []*v1alpha1.CommonNode{
						{
							Name: "mysql-0",
							Host: "mysql-0.mysql.default.svc.cluster.local",
							Port: 3306,
						},
						{
							Name: "mysql-1",
							Host: "example.com",
							Port: 3306,
						},
						{
							Name: "mysql-2",
							Host: "mysql-server",
							Port: 3306,
						},
					},
				},
			}

			_, err := mgra.ValidateCreate(context.TODO(), mysqlGR)
			Expect(err).NotTo(HaveOccurred())
		})
	})
})
