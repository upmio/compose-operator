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

package mysqlreplication

import (
	"context"
	"github.com/upmio/compose-operator/api/v1alpha1"
	"strings"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var _ = Describe("MysqlReplication Webhook", func() {

	var mra = mysqlReplicationAdmission{}

	Context("When creating MysqlReplication under Defaulting Webhook", func() {
		It("Should fill in the default values for mode, secret keys, and service", func() {
			mysqlRepl := &v1alpha1.MysqlReplication{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test-mysql-repl",
					Namespace: "default",
				},
				Spec: v1alpha1.MysqlReplicationSpec{
					Secret: v1alpha1.MysqlReplicationSecret{
						Name: "test-secret",
						// mysql and replication keys are empty - should get defaults
					},
					Source: &v1alpha1.CommonNode{
						Name: "source-node",
						Host: "192.168.1.10",
						Port: 3306,
					},
					// mode and service are empty - should get defaults
				},
			}

			err := mra.Default(context.TODO(), mysqlRepl)
			Expect(err).NotTo(HaveOccurred())

			// Check defaults were set
			Expect(mysqlRepl.Spec.Mode).To(Equal(v1alpha1.MysqlRplASync))
			Expect(mysqlRepl.Spec.Secret.Mysql).To(Equal("mysql"))
			Expect(mysqlRepl.Spec.Secret.Replication).To(Equal("replication"))
			Expect(mysqlRepl.Spec.Service).NotTo(BeNil())
			Expect(mysqlRepl.Spec.Service.Type).To(Equal(v1alpha1.ServiceTypeClusterIP))
		})

		It("Should not override existing values", func() {
			mysqlRepl := &v1alpha1.MysqlReplication{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test-mysql-repl",
					Namespace: "default",
				},
				Spec: v1alpha1.MysqlReplicationSpec{
					Mode: v1alpha1.MysqlRplSemiSync,
					Secret: v1alpha1.MysqlReplicationSecret{
						Name:        "test-secret",
						Mysql:       "custom-mysql",
						Replication: "custom-repl",
					},
					Source: &v1alpha1.CommonNode{
						Name: "source-node",
						Host: "192.168.1.10",
						Port: 3306,
					},
					Service: &v1alpha1.Service{Type: v1alpha1.ServiceTypeNodePort},
				},
			}

			err := mra.Default(context.TODO(), mysqlRepl)
			Expect(err).NotTo(HaveOccurred())

			// Check existing values were preserved
			Expect(mysqlRepl.Spec.Mode).To(Equal(v1alpha1.MysqlRplASync))
			Expect(mysqlRepl.Spec.Secret.Mysql).To(Equal("custom-mysql"))
			Expect(mysqlRepl.Spec.Secret.Replication).To(Equal("custom-repl"))
			Expect(mysqlRepl.Spec.Service.Type).To(Equal(v1alpha1.ServiceTypeNodePort))
		})
	})

	Context("When creating MysqlReplication under Validating Webhook", func() {
		It("Should deny if source node is missing", func() {
			mysqlRepl := &v1alpha1.MysqlReplication{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test-mysql-repl",
					Namespace: "default",
				},
				Spec: v1alpha1.MysqlReplicationSpec{
					Mode: v1alpha1.MysqlRplASync,
					Secret: v1alpha1.MysqlReplicationSecret{
						Name:        "test-secret",
						Mysql:       "mysql",
						Replication: "replication",
					},
					// Source is missing
				},
			}

			_, err := mra.ValidateCreate(context.TODO(), mysqlRepl)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("source node is required"))
		})

		It("Should deny if secret name is empty", func() {
			mysqlRepl := &v1alpha1.MysqlReplication{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test-mysql-repl",
					Namespace: "default",
				},
				Spec: v1alpha1.MysqlReplicationSpec{
					Mode: v1alpha1.MysqlRplASync,
					Secret: v1alpha1.MysqlReplicationSecret{
						Name: "", // Empty secret name
					},
					Source: &v1alpha1.CommonNode{
						Name: "source-node",
						Host: "192.168.1.10",
						Port: 3306,
					},
				},
			}

			_, err := mra.ValidateCreate(context.TODO(), mysqlRepl)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("secret name is required"))
		})

		It("Should deny if node names are duplicate", func() {
			mysqlRepl := &v1alpha1.MysqlReplication{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test-mysql-repl",
					Namespace: "default",
				},
				Spec: v1alpha1.MysqlReplicationSpec{
					Mode: v1alpha1.MysqlRplASync,
					Secret: v1alpha1.MysqlReplicationSecret{
						Name:        "test-secret",
						Mysql:       "mysql",
						Replication: "replication",
					},
					Source: &v1alpha1.CommonNode{
						Name: "duplicate-node", // Same name as replica
						Host: "192.168.1.10",
						Port: 3306,
					},
					Replica: v1alpha1.ReplicaNodes{
						{
							CommonNode: v1alpha1.CommonNode{
								Name: "duplicate-node", // Duplicate name
								Host: "192.168.1.11",
								Port: 3306,
							},
						},
					},
				},
			}

			_, err := mra.ValidateCreate(context.TODO(), mysqlRepl)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("node names must be unique"))
		})

		It("Should deny if port is invalid", func() {
			mysqlRepl := &v1alpha1.MysqlReplication{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test-mysql-repl",
					Namespace: "default",
				},
				Spec: v1alpha1.MysqlReplicationSpec{
					Mode: v1alpha1.MysqlRplASync,
					Secret: v1alpha1.MysqlReplicationSecret{
						Name:        "test-secret",
						Mysql:       "mysql",
						Replication: "replication",
					},
					Source: &v1alpha1.CommonNode{
						Name: "source-node",
						Host: "192.168.1.10",
						Port: 99999, // Invalid port
					},
				},
			}

			_, err := mra.ValidateCreate(context.TODO(), mysqlRepl)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("port must be between 1 and 65535"))
		})

		It("Should deny if host is invalid", func() {
			mysqlRepl := &v1alpha1.MysqlReplication{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test-mysql-repl",
					Namespace: "default",
				},
				Spec: v1alpha1.MysqlReplicationSpec{
					Mode: v1alpha1.MysqlRplASync,
					Secret: v1alpha1.MysqlReplicationSecret{
						Name:        "test-secret",
						Mysql:       "mysql",
						Replication: "replication",
					},
					Source: &v1alpha1.CommonNode{
						Name: "source-node",
						Host: "invalid..host", // Invalid hostname
						Port: 3306,
					},
				},
			}

			_, err := mra.ValidateCreate(context.TODO(), mysqlRepl)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("host must be a valid IP address or hostname"))
		})

		It("Should deny if replication mode is invalid", func() {
			mysqlRepl := &v1alpha1.MysqlReplication{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test-mysql-repl",
					Namespace: "default",
				},
				Spec: v1alpha1.MysqlReplicationSpec{
					Mode: v1alpha1.MysqlReplicationMode("invalid_mode"), // Invalid mode
					Secret: v1alpha1.MysqlReplicationSecret{
						Name:        "test-secret",
						Mysql:       "mysql",
						Replication: "replication",
					},
					Source: &v1alpha1.CommonNode{
						Name: "source-node",
						Host: "192.168.1.10",
						Port: 3306,
					},
				},
			}

			_, err := mra.ValidateCreate(context.TODO(), mysqlRepl)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("mode must be either 'rpl_async' or 'rpl_semi_sync'"))
		})

		It("Should admit if all required fields are provided correctly", func() {
			mysqlRepl := &v1alpha1.MysqlReplication{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test-mysql-repl",
					Namespace: "default",
				},
				Spec: v1alpha1.MysqlReplicationSpec{
					Mode: v1alpha1.MysqlRplSemiSync,
					Secret: v1alpha1.MysqlReplicationSecret{
						Name:        "test-secret",
						Mysql:       "mysql",
						Replication: "replication",
					},
					Source: &v1alpha1.CommonNode{
						Name: "source-node",
						Host: "192.168.1.10",
						Port: 3306,
					},
					Replica: v1alpha1.ReplicaNodes{
						{
							CommonNode: v1alpha1.CommonNode{
								Name: "replica-1",
								Host: "192.168.1.11",
								Port: 3306,
							},
						},
						{
							CommonNode: v1alpha1.CommonNode{
								Name: "replica-2",
								Host: "replica-2.example.com",
								Port: 3306,
							},
						},
					},
					Service: &v1alpha1.Service{Type: v1alpha1.ServiceTypeClusterIP},
				},
			}

			warnings, err := mra.ValidateCreate(context.TODO(), mysqlRepl)
			Expect(err).NotTo(HaveOccurred())
			Expect(warnings).To(BeEmpty()) // No warnings expected for valid config
		})

		It("Should return warning for single-node setup", func() {
			mysqlRepl := &v1alpha1.MysqlReplication{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test-mysql-repl",
					Namespace: "default",
				},
				Spec: v1alpha1.MysqlReplicationSpec{
					Mode: v1alpha1.MysqlRplASync,
					Secret: v1alpha1.MysqlReplicationSecret{
						Name:        "test-secret",
						Mysql:       "mysql",
						Replication: "replication",
					},
					Source: &v1alpha1.CommonNode{
						Name: "source-node",
						Host: "192.168.1.10",
						Port: 3306,
					},
					// No replicas - should generate warning
				},
			}

			warnings, err := mra.ValidateCreate(context.TODO(), mysqlRepl)
			Expect(err).NotTo(HaveOccurred())
			Expect(warnings).To(ContainElement(ContainSubstring("No replica nodes specified")))
		})
	})

	Context("When updating MysqlReplication under Validating Webhook", func() {
		It("Should deny if secret name is changed", func() {
			oldMysqlRepl := &v1alpha1.MysqlReplication{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test-mysql-repl",
					Namespace: "default",
				},
				Spec: v1alpha1.MysqlReplicationSpec{
					Secret: v1alpha1.MysqlReplicationSecret{
						Name: "old-secret",
					},
				},
			}

			newMysqlRepl := &v1alpha1.MysqlReplication{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test-mysql-repl",
					Namespace: "default",
				},
				Spec: v1alpha1.MysqlReplicationSpec{
					Mode: v1alpha1.MysqlRplASync,
					Secret: v1alpha1.MysqlReplicationSecret{
						Name:        "new-secret", // Changed secret name
						Mysql:       "mysql",
						Replication: "replication",
					},
					Source: &v1alpha1.CommonNode{
						Name: "source-node",
						Host: "192.168.1.10",
						Port: 3306,
					},
				},
			}

			_, err := mra.ValidateUpdate(context.TODO(), oldMysqlRepl, newMysqlRepl)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("secret reference cannot be changed"))
		})

		It("Should return warning when source node is changed", func() {
			oldMysqlRepl := &v1alpha1.MysqlReplication{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test-mysql-repl",
					Namespace: "default",
				},
				Spec: v1alpha1.MysqlReplicationSpec{
					Secret: v1alpha1.MysqlReplicationSecret{Name: "test-secret"},
					Source: &v1alpha1.CommonNode{
						Name: "old-source",
						Host: "192.168.1.10",
						Port: 3306,
					},
				},
			}

			newMysqlRepl := &v1alpha1.MysqlReplication{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test-mysql-repl",
					Namespace: "default",
				},
				Spec: v1alpha1.MysqlReplicationSpec{
					Mode: v1alpha1.MysqlRplASync,
					Secret: v1alpha1.MysqlReplicationSecret{
						Name:        "test-secret",
						Mysql:       "mysql",
						Replication: "replication",
					},
					Source: &v1alpha1.CommonNode{
						Name: "new-source", // Changed source
						Host: "192.168.1.20",
						Port: 3306,
					},
				},
			}

			warnings, err := mra.ValidateUpdate(context.TODO(), oldMysqlRepl, newMysqlRepl)
			Expect(err).NotTo(HaveOccurred())
			Expect(warnings).To(ContainElement(ContainSubstring("Changing source node may cause data inconsistency")))
		})
	})

	Context("When deleting MysqlReplication under Validating Webhook", func() {
		It("Should return warning about cleanup", func() {
			mysqlRepl := &v1alpha1.MysqlReplication{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test-mysql-repl",
					Namespace: "default",
				},
			}

			warnings, err := mra.ValidateDelete(context.TODO(), mysqlRepl)
			Expect(err).NotTo(HaveOccurred())
			Expect(warnings).To(ContainElement(ContainSubstring("Deleting MysqlReplication will stop managing replication")))
		})
	})

	Context("Hostname validation", func() {
		DescribeTable("Valid hostnames",
			func(hostname string) {
				Expect(isValidHostname(hostname)).To(BeTrue())
			},
			Entry("simple hostname", "localhost"),
			Entry("FQDN", "example.com"),
			Entry("subdomain", "mysql.example.com"),
			Entry("hyphenated", "mysql-server.example.com"),
			Entry("numbers", "mysql01.example.com"),
		)

		DescribeTable("Invalid hostnames",
			func(hostname string) {
				Expect(isValidHostname(hostname)).To(BeFalse())
			},
			Entry("empty string", ""),
			Entry("double dots", "example..com"),
			Entry("starting hyphen", "-example.com"),
			Entry("ending hyphen", "example.com-"),
			Entry("special characters", "example.com!"),
			Entry("too long label", strings.Repeat("a", 64)+".com"),
		)
	})
})
