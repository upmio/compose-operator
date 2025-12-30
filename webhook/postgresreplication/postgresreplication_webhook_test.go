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
	"context"

	"github.com/upmio/compose-operator/api/v1alpha1"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var _ = Describe("PostgresReplication Webhook", func() {

	var pra = postgresReplicationAdmission{}

	Context("When creating PostgresReplication under Defaulting Webhook", func() {
		It("Should fill in the default values for mode, secret keys, and service", func() {
			postgresRepl := &v1alpha1.PostgresReplication{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test-postgres-repl",
					Namespace: "default",
				},
				Spec: v1alpha1.PostgresReplicationSpec{
					Secret: v1alpha1.PostgresReplicationSecret{
						Name: "test-secret",
						// postgresql and replication keys are empty - should get defaults
					},
					Primary: &v1alpha1.CommonNode{
						Name: "primary-node",
						Host: "192.168.1.10",
						Port: 5432,
					},
					// mode and service are empty - should get defaults
				},
			}

			err := pra.Default(context.TODO(), postgresRepl)
			Expect(err).NotTo(HaveOccurred())

			// Check defaults were set
			Expect(postgresRepl.Spec.Mode).To(Equal(v1alpha1.PostgresRplAsync))
			Expect(postgresRepl.Spec.Secret.Postgresql).To(Equal("postgres"))
			Expect(postgresRepl.Spec.Secret.Replication).To(Equal("replication"))
			Expect(postgresRepl.Spec.Service).NotTo(BeNil())
			Expect(postgresRepl.Spec.Service.Type).To(Equal(v1alpha1.ServiceTypeClusterIP))
		})

		It("Should not override existing values", func() {
			postgresRepl := &v1alpha1.PostgresReplication{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test-postgres-repl",
					Namespace: "default",
				},
				Spec: v1alpha1.PostgresReplicationSpec{
					Mode: v1alpha1.PostgresRplSync,
					Secret: v1alpha1.PostgresReplicationSecret{
						Name:        "test-secret",
						Postgresql:  "custom-postgres",
						Replication: "custom-repl",
					},
					Primary: &v1alpha1.CommonNode{
						Name: "primary-node",
						Host: "192.168.1.10",
						Port: 5432,
					},
					Service: &v1alpha1.Service{Type: v1alpha1.ServiceTypeNodePort},
				},
			}

			err := pra.Default(context.TODO(), postgresRepl)
			Expect(err).NotTo(HaveOccurred())

			// Check existing values were preserved
			Expect(postgresRepl.Spec.Mode).To(Equal(v1alpha1.PostgresRplSync))
			Expect(postgresRepl.Spec.Secret.Postgresql).To(Equal("custom-postgres"))
			Expect(postgresRepl.Spec.Secret.Replication).To(Equal("custom-repl"))
			Expect(postgresRepl.Spec.Service.Type).To(Equal(v1alpha1.ServiceTypeNodePort))
		})
	})

	Context("When creating PostgresReplication under Validating Webhook", func() {
		It("Should deny if primary node is missing", func() {
			postgresRepl := &v1alpha1.PostgresReplication{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test-postgres-repl",
					Namespace: "default",
				},
				Spec: v1alpha1.PostgresReplicationSpec{
					Mode: v1alpha1.PostgresRplAsync,
					Secret: v1alpha1.PostgresReplicationSecret{
						Name:        "test-secret",
						Postgresql:  "postgresql",
						Replication: "replication",
					},
					// Primary is missing
				},
			}

			_, err := pra.ValidateCreate(context.TODO(), postgresRepl)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("primary node is required"))
		})

		It("Should deny if secret name is empty", func() {
			postgresRepl := &v1alpha1.PostgresReplication{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test-postgres-repl",
					Namespace: "default",
				},
				Spec: v1alpha1.PostgresReplicationSpec{
					Mode: v1alpha1.PostgresRplAsync,
					Secret: v1alpha1.PostgresReplicationSecret{
						Name: "", // Empty secret name
					},
					Primary: &v1alpha1.CommonNode{
						Name: "primary-node",
						Host: "192.168.1.10",
						Port: 5432,
					},
				},
			}

			_, err := pra.ValidateCreate(context.TODO(), postgresRepl)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("secret name is required"))
		})

		It("Should deny if node names are duplicate", func() {
			postgresRepl := &v1alpha1.PostgresReplication{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test-postgres-repl",
					Namespace: "default",
				},
				Spec: v1alpha1.PostgresReplicationSpec{
					Mode: v1alpha1.PostgresRplAsync,
					Secret: v1alpha1.PostgresReplicationSecret{
						Name:        "test-secret",
						Postgresql:  "postgresql",
						Replication: "replication",
					},
					Primary: &v1alpha1.CommonNode{
						Name: "postgres-duplicate",
						Host: "192.168.1.10",
						Port: 5432,
					},
					Standby: []*v1alpha1.ReplicaNode{
						{
							CommonNode: v1alpha1.CommonNode{
								Name: "postgres-duplicate", // Duplicate name
								Host: "192.168.1.11",
								Port: 5432,
							},
							Isolated: false,
						},
					},
				},
			}

			_, err := pra.ValidateCreate(context.TODO(), postgresRepl)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("node names must be unique"))
		})

		It("Should deny if node has invalid port", func() {
			postgresRepl := &v1alpha1.PostgresReplication{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test-postgres-repl",
					Namespace: "default",
				},
				Spec: v1alpha1.PostgresReplicationSpec{
					Mode: v1alpha1.PostgresRplAsync,
					Secret: v1alpha1.PostgresReplicationSecret{
						Name:        "test-secret",
						Postgresql:  "postgresql",
						Replication: "replication",
					},
					Primary: &v1alpha1.CommonNode{
						Name: "primary-node",
						Host: "192.168.1.10",
						Port: 0, // Invalid port
					},
				},
			}

			_, err := pra.ValidateCreate(context.TODO(), postgresRepl)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("port must be between 1 and 65535"))
		})

		It("Should deny if node has empty hostname", func() {
			postgresRepl := &v1alpha1.PostgresReplication{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test-postgres-repl",
					Namespace: "default",
				},
				Spec: v1alpha1.PostgresReplicationSpec{
					Mode: v1alpha1.PostgresRplAsync,
					Secret: v1alpha1.PostgresReplicationSecret{
						Name:        "test-secret",
						Postgresql:  "postgresql",
						Replication: "replication",
					},
					Primary: &v1alpha1.CommonNode{
						Name: "primary-node",
						Host: "", // Empty host
						Port: 5432,
					},
				},
			}

			_, err := pra.ValidateCreate(context.TODO(), postgresRepl)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("primary node host is required"))
		})

		It("Should accept valid PostgresReplication", func() {
			postgresRepl := &v1alpha1.PostgresReplication{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test-postgres-repl",
					Namespace: "default",
				},
				Spec: v1alpha1.PostgresReplicationSpec{
					Mode: v1alpha1.PostgresRplAsync,
					Secret: v1alpha1.PostgresReplicationSecret{
						Name:        "test-secret",
						Postgresql:  "postgresql",
						Replication: "replication",
					},
					Primary: &v1alpha1.CommonNode{
						Name: "primary-node",
						Host: "postgres-primary.default.svc.cluster.local",
						Port: 5432,
					},
					Standby: []*v1alpha1.ReplicaNode{
						{
							CommonNode: v1alpha1.CommonNode{
								Name: "standby-1", // Duplicate name
								Host: "postgres-standby-1.default.svc.cluster.local",
								Port: 5432,
							},
							Isolated: false,
						},
						{
							CommonNode: v1alpha1.CommonNode{
								Name: "standby-2", // Duplicate name
								Host: "postgres-standby-2.default.svc.cluster.local",
								Port: 5432,
							},
							Isolated: false,
						},
					},
				},
			}

			warnings, err := pra.ValidateCreate(context.TODO(), postgresRepl)
			Expect(err).NotTo(HaveOccurred())
			Expect(warnings).To(HaveLen(0))
		})
	})

	Context("When updating PostgresReplication under Validating Webhook", func() {
		It("Should deny if secret name is changed", func() {
			oldPostgresRepl := &v1alpha1.PostgresReplication{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test-postgres-repl",
					Namespace: "default",
				},
				Spec: v1alpha1.PostgresReplicationSpec{
					Mode: v1alpha1.PostgresRplAsync,
					Secret: v1alpha1.PostgresReplicationSecret{
						Name:        "old-secret",
						Postgresql:  "postgresql",
						Replication: "replication",
					},
					Primary: &v1alpha1.CommonNode{
						Name: "primary-node",
						Host: "postgres-primary.default.svc.cluster.local",
						Port: 5432,
					},
				},
			}

			newPostgresRepl := &v1alpha1.PostgresReplication{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test-postgres-repl",
					Namespace: "default",
				},
				Spec: v1alpha1.PostgresReplicationSpec{
					Mode: v1alpha1.PostgresRplAsync,
					Secret: v1alpha1.PostgresReplicationSecret{
						Name:        "new-secret", // Changed secret name
						Postgresql:  "postgresql",
						Replication: "replication",
					},
					Primary: &v1alpha1.CommonNode{
						Name: "primary-node",
						Host: "postgres-primary.default.svc.cluster.local",
						Port: 5432,
					},
				},
			}

			_, err := pra.ValidateUpdate(context.TODO(), oldPostgresRepl, newPostgresRepl)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("secret reference cannot be changed after creation"))
		})

		It("Should warn when changing primary node", func() {
			oldPostgresRepl := &v1alpha1.PostgresReplication{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test-postgres-repl",
					Namespace: "default",
				},
				Spec: v1alpha1.PostgresReplicationSpec{
					Mode: v1alpha1.PostgresRplAsync,
					Secret: v1alpha1.PostgresReplicationSecret{
						Name:        "test-secret",
						Postgresql:  "postgresql",
						Replication: "replication",
					},
					Primary: &v1alpha1.CommonNode{
						Name: "old-primary",
						Host: "postgres-primary-old.default.svc.cluster.local",
						Port: 5432,
					},
				},
			}

			newPostgresRepl := &v1alpha1.PostgresReplication{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test-postgres-repl",
					Namespace: "default",
				},
				Spec: v1alpha1.PostgresReplicationSpec{
					Mode: v1alpha1.PostgresRplAsync,
					Secret: v1alpha1.PostgresReplicationSecret{
						Name:        "test-secret",
						Postgresql:  "postgresql",
						Replication: "replication",
					},
					Primary: &v1alpha1.CommonNode{
						Name: "new-primary", // Changed primary node
						Host: "postgres-primary-new.default.svc.cluster.local",
						Port: 5432,
					},
				},
			}

			warnings, err := pra.ValidateUpdate(context.TODO(), oldPostgresRepl, newPostgresRepl)
			Expect(err).NotTo(HaveOccurred())
			Expect(warnings).To(HaveLen(2))
			Expect(warnings[0]).To(ContainSubstring("No standby nodes specified"))
			Expect(warnings[1]).To(ContainSubstring("Changing primary node may cause data inconsistency"))
		})
	})

	Context("When deleting PostgresReplication under Validating Webhook", func() {
		It("Should provide warning about cleanup", func() {
			postgresRepl := &v1alpha1.PostgresReplication{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test-postgres-repl",
					Namespace: "default",
				},
				Spec: v1alpha1.PostgresReplicationSpec{
					Mode: v1alpha1.PostgresRplAsync,
					Secret: v1alpha1.PostgresReplicationSecret{
						Name:        "test-secret",
						Postgresql:  "postgresql",
						Replication: "replication",
					},
					Primary: &v1alpha1.CommonNode{
						Name: "primary-node",
						Host: "postgres-primary.default.svc.cluster.local",
						Port: 5432,
					},
				},
			}

			warnings, err := pra.ValidateDelete(context.TODO(), postgresRepl)
			Expect(err).NotTo(HaveOccurred())
			Expect(warnings).To(HaveLen(1))
			Expect(warnings[0]).To(ContainSubstring("Deleting PostgresReplication will stop managing"))
		})
	})

	Context("Hostname validation", func() {
		It("Should accept valid IP addresses", func() {
			postgresRepl := &v1alpha1.PostgresReplication{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test-postgres-repl",
					Namespace: "default",
				},
				Spec: v1alpha1.PostgresReplicationSpec{
					Mode: v1alpha1.PostgresRplAsync,
					Secret: v1alpha1.PostgresReplicationSecret{
						Name:        "test-secret",
						Postgresql:  "postgresql",
						Replication: "replication",
					},
					Primary: &v1alpha1.CommonNode{
						Name: "primary",
						Host: "192.168.1.100",
						Port: 5432,
					},
					Standby: []*v1alpha1.ReplicaNode{
						{
							CommonNode: v1alpha1.CommonNode{
								Name: "standby-1", // Duplicate name
								Host: "10.0.0.1",
								Port: 5432,
							},
							Isolated: false,
						},
						{
							CommonNode: v1alpha1.CommonNode{
								Name: "standby-2", // Duplicate name
								Host: "10.0.0.1",
								Port: 5432,
							},
							Isolated: false,
						},
					},
				},
			}

			_, err := pra.ValidateCreate(context.TODO(), postgresRepl)
			Expect(err).NotTo(HaveOccurred())
		})

		It("Should accept valid hostnames", func() {
			postgresRepl := &v1alpha1.PostgresReplication{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test-postgres-repl",
					Namespace: "default",
				},
				Spec: v1alpha1.PostgresReplicationSpec{
					Mode: v1alpha1.PostgresRplAsync,
					Secret: v1alpha1.PostgresReplicationSecret{
						Name:        "test-secret",
						Postgresql:  "postgresql",
						Replication: "replication",
					},
					Primary: &v1alpha1.CommonNode{
						Name: "primary",
						Host: "postgres-primary.default.svc.cluster.local",
						Port: 5432,
					},
					//Standby: []*v1alpha1.CommonNode{
					//	{
					//		Name: "standby-1",
					//		Host: "example.com",
					//		Port: 5432,
					//	},
					//	{
					//		Name: "standby-2",
					//		Host: "postgres-server",
					//		Port: 5432,
					//	},
					//},
					Standby: []*v1alpha1.ReplicaNode{
						{
							CommonNode: v1alpha1.CommonNode{
								Name: "standby-1",
								Host: "example.com",
								Port: 5432,
							},
							Isolated: false,
						},
						{
							CommonNode: v1alpha1.CommonNode{
								Name: "standby-2",
								Host: "postgres-server",
								Port: 5432,
							},
							Isolated: false,
						},
					},
				},
			}

			_, err := pra.ValidateCreate(context.TODO(), postgresRepl)
			Expect(err).NotTo(HaveOccurred())
		})
	})
})
