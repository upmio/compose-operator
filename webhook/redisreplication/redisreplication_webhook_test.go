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

package redisreplication

import (
	"context"
	"github.com/upmio/compose-operator/api/v1alpha1"
	"strings"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var _ = Describe("RedisReplication Webhook", func() {

	var rra = redisReplicationAdmission{}

	Context("When creating RedisReplication under Defaulting Webhook", func() {
		It("Should fill in the default values for redis secret key and service type", func() {
			redisRepl := &v1alpha1.RedisReplication{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test-redis-repl",
					Namespace: "default",
				},
				Spec: v1alpha1.RedisReplicationSpec{
					Secret: v1alpha1.RedisReplicationSecret{
						Name: "test-secret",
						// redis key is empty - should get default
					},
					Source: &v1alpha1.RedisNode{
						CommonNode: v1alpha1.CommonNode{
							Name: "source-node",
							Host: "192.168.1.10",
							Port: 6379,
						},
					},
					// service is nil - should get default
				},
			}

			err := rra.Default(context.TODO(), redisRepl)
			Expect(err).NotTo(HaveOccurred())

			// Check defaults were set
			Expect(redisRepl.Spec.Secret.Redis).To(Equal("redis"))
			Expect(redisRepl.Spec.Service).NotTo(BeNil())
			Expect(redisRepl.Spec.Service.Type).To(Equal(v1alpha1.ServiceTypeClusterIP))
		})

		It("Should set default service type when service is provided but type is empty", func() {
			redisRepl := &v1alpha1.RedisReplication{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test-redis-repl",
					Namespace: "default",
				},
				Spec: v1alpha1.RedisReplicationSpec{
					Secret: v1alpha1.RedisReplicationSecret{
						Name: "test-secret",
					},
					Source: &v1alpha1.RedisNode{
						CommonNode: v1alpha1.CommonNode{
							Name: "source-node",
							Host: "192.168.1.10",
							Port: 6379,
						},
					},
					Service: &v1alpha1.Service{
						// Type is empty - should get default
					},
				},
			}

			err := rra.Default(context.TODO(), redisRepl)
			Expect(err).NotTo(HaveOccurred())

			// Check defaults were set
			Expect(redisRepl.Spec.Service.Type).To(Equal(v1alpha1.ServiceTypeClusterIP))
		})

		It("Should not override existing values", func() {
			redisRepl := &v1alpha1.RedisReplication{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test-redis-repl",
					Namespace: "default",
				},
				Spec: v1alpha1.RedisReplicationSpec{
					Secret: v1alpha1.RedisReplicationSecret{
						Name:  "test-secret",
						Redis: "custom-redis",
					},
					Source: &v1alpha1.RedisNode{
						CommonNode: v1alpha1.CommonNode{
							Name: "source-node",
							Host: "192.168.1.10",
							Port: 6379,
						},
					},
					Service: &v1alpha1.Service{Type: v1alpha1.ServiceTypeNodePort},
				},
			}

			err := rra.Default(context.TODO(), redisRepl)
			Expect(err).NotTo(HaveOccurred())

			// Check existing values were preserved
			Expect(redisRepl.Spec.Secret.Redis).To(Equal("custom-redis"))
			Expect(redisRepl.Spec.Service.Type).To(Equal(v1alpha1.ServiceTypeNodePort))
		})
	})

	Context("When creating RedisReplication under Validating Webhook", func() {
		It("Should deny if source node is missing", func() {
			redisRepl := &v1alpha1.RedisReplication{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test-redis-repl",
					Namespace: "default",
				},
				Spec: v1alpha1.RedisReplicationSpec{
					Secret: v1alpha1.RedisReplicationSecret{
						Name:  "test-secret",
						Redis: "redis",
					},
					// Source is missing
				},
			}

			_, err := rra.ValidateCreate(context.TODO(), redisRepl)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("source node is required"))
		})

		It("Should deny if secret name is empty", func() {
			redisRepl := &v1alpha1.RedisReplication{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test-redis-repl",
					Namespace: "default",
				},
				Spec: v1alpha1.RedisReplicationSpec{
					Secret: v1alpha1.RedisReplicationSecret{
						Name: "", // Empty secret name
					},
					Source: &v1alpha1.RedisNode{
						CommonNode: v1alpha1.CommonNode{
							Name: "source-node",
							Host: "192.168.1.10",
							Port: 6379,
						},
					},
				},
			}

			_, err := rra.ValidateCreate(context.TODO(), redisRepl)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("secret name is required"))
		})

		It("Should deny if redis secret key is empty", func() {
			redisRepl := &v1alpha1.RedisReplication{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test-redis-repl",
					Namespace: "default",
				},
				Spec: v1alpha1.RedisReplicationSpec{
					Secret: v1alpha1.RedisReplicationSecret{
						Name:  "test-secret",
						Redis: "", // Empty redis key
					},
					Source: &v1alpha1.RedisNode{
						CommonNode: v1alpha1.CommonNode{
							Name: "source-node",
							Host: "192.168.1.10",
							Port: 6379,
						},
					},
				},
			}

			_, err := rra.ValidateCreate(context.TODO(), redisRepl)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("redis secret key is required"))
		})

		It("Should deny if source node name is empty", func() {
			redisRepl := &v1alpha1.RedisReplication{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test-redis-repl",
					Namespace: "default",
				},
				Spec: v1alpha1.RedisReplicationSpec{
					Secret: v1alpha1.RedisReplicationSecret{
						Name:  "test-secret",
						Redis: "redis",
					},
					Source: &v1alpha1.RedisNode{
						CommonNode: v1alpha1.CommonNode{
							Host: "192.168.1.10",
							Port: 6379,
						},
					},
				},
			}

			_, err := rra.ValidateCreate(context.TODO(), redisRepl)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("source node name is required"))
		})

		It("Should deny if source node host is empty", func() {
			redisRepl := &v1alpha1.RedisReplication{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test-redis-repl",
					Namespace: "default",
				},
				Spec: v1alpha1.RedisReplicationSpec{
					Secret: v1alpha1.RedisReplicationSecret{
						Name:  "test-secret",
						Redis: "redis",
					},
					Source: &v1alpha1.RedisNode{
						CommonNode: v1alpha1.CommonNode{
							Name: "source-node",
							Port: 6379,
						},
					},
				},
			}

			_, err := rra.ValidateCreate(context.TODO(), redisRepl)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("source node host is required"))
		})

		It("Should deny if source node port is invalid", func() {
			redisRepl := &v1alpha1.RedisReplication{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test-redis-repl",
					Namespace: "default",
				},
				Spec: v1alpha1.RedisReplicationSpec{
					Secret: v1alpha1.RedisReplicationSecret{
						Name:  "test-secret",
						Redis: "redis",
					},
					Source: &v1alpha1.RedisNode{
						CommonNode: v1alpha1.CommonNode{
							Name: "source-node",
							Host: "192.168.1.10",
							Port: 99999,
						},
					},
				},
			}

			_, err := rra.ValidateCreate(context.TODO(), redisRepl)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("port must be between 1 and 65535"))
		})

		It("Should deny if source node host is invalid", func() {
			redisRepl := &v1alpha1.RedisReplication{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test-redis-repl",
					Namespace: "default",
				},
				Spec: v1alpha1.RedisReplicationSpec{
					Secret: v1alpha1.RedisReplicationSecret{
						Name:  "test-secret",
						Redis: "redis",
					},
					Source: &v1alpha1.RedisNode{
						CommonNode: v1alpha1.CommonNode{
							Name: "source-node",
							Host: "invalid..hostname",
							Port: 6379,
						},
					},
				},
			}

			_, err := rra.ValidateCreate(context.TODO(), redisRepl)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("host must be a valid IP address or hostname"))
		})

		It("Should deny if replica node names are duplicate", func() {
			redisRepl := &v1alpha1.RedisReplication{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test-redis-repl",
					Namespace: "default",
				},
				Spec: v1alpha1.RedisReplicationSpec{
					Secret: v1alpha1.RedisReplicationSecret{
						Name:  "test-secret",
						Redis: "redis",
					},
					Source: &v1alpha1.RedisNode{
						CommonNode: v1alpha1.CommonNode{
							Name: "source-node",
							Host: "192.168.1.10",
							Port: 6379,
						},
					},
					Replica: v1alpha1.RedisReplicaNodes{
						&v1alpha1.RedisReplicaNode{
							RedisNode: v1alpha1.RedisNode{
								CommonNode: v1alpha1.CommonNode{
									Name: "duplicate-replica",
									Host: "192.168.1.11",
									Port: 6379,
								},
								AnnounceHost: "",
								AnnouncePort: 0,
							},
							Isolated: false,
						},
						&v1alpha1.RedisReplicaNode{
							RedisNode: v1alpha1.RedisNode{
								CommonNode: v1alpha1.CommonNode{
									Name: "duplicate-replica",
									Host: "192.168.1.12",
									Port: 6379,
								},
								AnnounceHost: "",
								AnnouncePort: 0,
							},
							Isolated: false,
						},
					},
				},
			}

			_, err := rra.ValidateCreate(context.TODO(), redisRepl)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("node names must be unique across source and replicas"))
		})

		It("Should deny if source and replica names are duplicate", func() {
			redisRepl := &v1alpha1.RedisReplication{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test-redis-repl",
					Namespace: "default",
				},
				Spec: v1alpha1.RedisReplicationSpec{
					Secret: v1alpha1.RedisReplicationSecret{
						Name:  "test-secret",
						Redis: "redis",
					},
					Source: &v1alpha1.RedisNode{
						CommonNode: v1alpha1.CommonNode{
							Name: "duplicate-node", // Same name as replica
							Host: "192.168.1.10",
							Port: 6379,
						},
					},
					Replica: v1alpha1.RedisReplicaNodes{
						&v1alpha1.RedisReplicaNode{
							RedisNode: v1alpha1.RedisNode{
								CommonNode: v1alpha1.CommonNode{
									Name: "duplicate-node",
									Host: "192.168.1.11",
									Port: 6379,
								},
								AnnounceHost: "",
								AnnouncePort: 0,
							},
							Isolated: false,
						},
					},
				},
			}

			_, err := rra.ValidateCreate(context.TODO(), redisRepl)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("node names must be unique across source and replicas"))
		})

		It("Should deny if replica node has invalid fields", func() {
			redisRepl := &v1alpha1.RedisReplication{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test-redis-repl",
					Namespace: "default",
				},
				Spec: v1alpha1.RedisReplicationSpec{
					Secret: v1alpha1.RedisReplicationSecret{
						Name:  "test-secret",
						Redis: "redis",
					},
					Source: &v1alpha1.RedisNode{
						CommonNode: v1alpha1.CommonNode{
							Name: "source-node",
							Host: "192.168.1.10",
							Port: 6379,
						},
					},
					Replica: v1alpha1.RedisReplicaNodes{
						&v1alpha1.RedisReplicaNode{
							RedisNode: v1alpha1.RedisNode{
								CommonNode: v1alpha1.CommonNode{
									Name: "replica-node",
									Host: "invalid..hostname",
									Port: 6379,
								},
								AnnounceHost: "",
								AnnouncePort: 0,
							},
							Isolated: false,
						},
					},
				},
			}

			_, err := rra.ValidateCreate(context.TODO(), redisRepl)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("host must be a valid IP address or hostname"))
		})

		It("Should deny if service type is invalid", func() {
			redisRepl := &v1alpha1.RedisReplication{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test-redis-repl",
					Namespace: "default",
				},
				Spec: v1alpha1.RedisReplicationSpec{
					Secret: v1alpha1.RedisReplicationSecret{
						Name:  "test-secret",
						Redis: "redis",
					},
					Source: &v1alpha1.RedisNode{
						CommonNode: v1alpha1.CommonNode{
							Name: "source-node",
							Host: "192.168.1.10",
							Port: 6379,
						},
					},
					Service: &v1alpha1.Service{
						Type: v1alpha1.ServiceType("InvalidServiceType"), // Invalid service type
					},
				},
			}

			_, err := rra.ValidateCreate(context.TODO(), redisRepl)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("service type must be ClusterIP, NodePort, LoadBalancer, or ExternalName"))
		})

		It("Should return warning for single-node setup", func() {
			redisRepl := &v1alpha1.RedisReplication{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test-redis-repl",
					Namespace: "default",
				},
				Spec: v1alpha1.RedisReplicationSpec{
					Secret: v1alpha1.RedisReplicationSecret{
						Name:  "test-secret",
						Redis: "redis",
					},
					Source: &v1alpha1.RedisNode{
						CommonNode: v1alpha1.CommonNode{
							Name: "source-node",
							Host: "192.168.1.10",
							Port: 6379,
						},
					},
					// No replicas - should generate warning
				},
			}

			warnings, err := rra.ValidateCreate(context.TODO(), redisRepl)
			Expect(err).NotTo(HaveOccurred())
			Expect(warnings).To(ContainElement(ContainSubstring("No replica nodes specified")))
		})

		It("Should admit if all required fields are provided correctly", func() {
			redisRepl := &v1alpha1.RedisReplication{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test-redis-repl",
					Namespace: "default",
				},
				Spec: v1alpha1.RedisReplicationSpec{
					Secret: v1alpha1.RedisReplicationSecret{
						Name:  "test-secret",
						Redis: "redis",
					},
					Source: &v1alpha1.RedisNode{
						CommonNode: v1alpha1.CommonNode{
							Name: "source-node",
							Host: "192.168.1.10",
							Port: 6379,
						},
					},
					Replica: v1alpha1.RedisReplicaNodes{
						&v1alpha1.RedisReplicaNode{
							RedisNode: v1alpha1.RedisNode{
								CommonNode: v1alpha1.CommonNode{
									Name: "replica-1",
									Host: "192.168.1.11",
									Port: 6379,
								},
								AnnounceHost: "",
								AnnouncePort: 0,
							},
							Isolated: false,
						},
						&v1alpha1.RedisReplicaNode{
							RedisNode: v1alpha1.RedisNode{
								CommonNode: v1alpha1.CommonNode{
									Name: "replica-2",
									Host: "replica-2.example.com",
									Port: 6379,
								},
								AnnounceHost: "",
								AnnouncePort: 0,
							},
							Isolated: false,
						},
					},

					Service: &v1alpha1.Service{Type: v1alpha1.ServiceTypeClusterIP},
				},
			}

			warnings, err := rra.ValidateCreate(context.TODO(), redisRepl)
			Expect(err).NotTo(HaveOccurred())
			Expect(warnings).To(BeEmpty()) // No warnings expected for valid config
		})
	})

	Context("When updating RedisReplication under Validating Webhook", func() {
		It("Should deny if secret name is changed", func() {
			oldRedisRepl := &v1alpha1.RedisReplication{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test-redis-repl",
					Namespace: "default",
				},
				Spec: v1alpha1.RedisReplicationSpec{
					Secret: v1alpha1.RedisReplicationSecret{
						Name: "old-secret",
					},
				},
			}

			newRedisRepl := &v1alpha1.RedisReplication{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test-redis-repl",
					Namespace: "default",
				},
				Spec: v1alpha1.RedisReplicationSpec{
					Secret: v1alpha1.RedisReplicationSecret{
						Name:  "new-secret", // Changed secret name
						Redis: "redis",
					},
					Source: &v1alpha1.RedisNode{
						CommonNode: v1alpha1.CommonNode{
							Name: "source-node",
							Host: "192.168.1.10",
							Port: 6379,
						},
					},
				},
			}

			_, err := rra.ValidateUpdate(context.TODO(), oldRedisRepl, newRedisRepl)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("secret reference cannot be changed"))
		})

		It("Should return warning when source node is changed", func() {
			oldRedisRepl := &v1alpha1.RedisReplication{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test-redis-repl",
					Namespace: "default",
				},
				Spec: v1alpha1.RedisReplicationSpec{
					Secret: v1alpha1.RedisReplicationSecret{Name: "test-secret"},
					Source: &v1alpha1.RedisNode{
						CommonNode: v1alpha1.CommonNode{
							Name: "old-source",
							Host: "192.168.1.10",
							Port: 6379,
						},
					},
				},
			}

			newRedisRepl := &v1alpha1.RedisReplication{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test-redis-repl",
					Namespace: "default",
				},
				Spec: v1alpha1.RedisReplicationSpec{
					Secret: v1alpha1.RedisReplicationSecret{
						Name:  "test-secret",
						Redis: "redis",
					},
					Source: &v1alpha1.RedisNode{
						CommonNode: v1alpha1.CommonNode{
							Name: "new-source", // Changed source name
							Host: "192.168.1.20",
							Port: 6379,
						},
					},
				},
			}

			warnings, err := rra.ValidateUpdate(context.TODO(), oldRedisRepl, newRedisRepl)
			Expect(err).NotTo(HaveOccurred())
			Expect(warnings).To(ContainElement(ContainSubstring("Changing source node may cause data inconsistency")))
		})

		It("Should return warning when source host is changed", func() {
			oldRedisRepl := &v1alpha1.RedisReplication{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test-redis-repl",
					Namespace: "default",
				},
				Spec: v1alpha1.RedisReplicationSpec{
					Secret: v1alpha1.RedisReplicationSecret{Name: "test-secret"},
					Source: &v1alpha1.RedisNode{
						CommonNode: v1alpha1.CommonNode{
							Name: "source-node",
							Host: "192.168.1.10",
							Port: 6379,
						},
					},
				},
			}

			newRedisRepl := &v1alpha1.RedisReplication{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test-redis-repl",
					Namespace: "default",
				},
				Spec: v1alpha1.RedisReplicationSpec{
					Secret: v1alpha1.RedisReplicationSecret{
						Name:  "test-secret",
						Redis: "redis",
					},
					Source: &v1alpha1.RedisNode{
						CommonNode: v1alpha1.CommonNode{
							Name: "source-node",
							Host: "192.168.1.20",
							Port: 6379,
						},
					},
				},
			}

			warnings, err := rra.ValidateUpdate(context.TODO(), oldRedisRepl, newRedisRepl)
			Expect(err).NotTo(HaveOccurred())
			Expect(warnings).To(ContainElement(ContainSubstring("Changing source node may cause data inconsistency")))
		})

		It("Should return warning when source port is changed", func() {
			oldRedisRepl := &v1alpha1.RedisReplication{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test-redis-repl",
					Namespace: "default",
				},
				Spec: v1alpha1.RedisReplicationSpec{
					Secret: v1alpha1.RedisReplicationSecret{Name: "test-secret"},
					Source: &v1alpha1.RedisNode{
						CommonNode: v1alpha1.CommonNode{
							Name: "source-node",
							Host: "192.168.1.10",
							Port: 6379,
						},
					},
				},
			}

			newRedisRepl := &v1alpha1.RedisReplication{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test-redis-repl",
					Namespace: "default",
				},
				Spec: v1alpha1.RedisReplicationSpec{
					Secret: v1alpha1.RedisReplicationSecret{
						Name:  "test-secret",
						Redis: "redis",
					},
					Source: &v1alpha1.RedisNode{
						CommonNode: v1alpha1.CommonNode{
							Name: "source-node",
							Host: "192.168.1.10",
							Port: 6380,
						},
					},
				},
			}

			warnings, err := rra.ValidateUpdate(context.TODO(), oldRedisRepl, newRedisRepl)
			Expect(err).NotTo(HaveOccurred())
			Expect(warnings).To(ContainElement(ContainSubstring("Changing source node may cause data inconsistency")))
		})

		It("Should allow update when only non-critical fields change", func() {
			oldRedisRepl := &v1alpha1.RedisReplication{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test-redis-repl",
					Namespace: "default",
				},
				Spec: v1alpha1.RedisReplicationSpec{
					Secret: v1alpha1.RedisReplicationSecret{Name: "test-secret", Redis: "redis"},
					Source: &v1alpha1.RedisNode{
						CommonNode: v1alpha1.CommonNode{
							Name: "source-node",
							Host: "192.168.1.10",
							Port: 6379,
						},
					},
				},
			}

			newRedisRepl := &v1alpha1.RedisReplication{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test-redis-repl",
					Namespace: "default",
					Labels:    map[string]string{"updated": "true"}, // Only metadata change
				},
				Spec: v1alpha1.RedisReplicationSpec{
					Secret: v1alpha1.RedisReplicationSecret{Name: "test-secret", Redis: "redis"}, // Same secret
					Source: &v1alpha1.RedisNode{
						CommonNode: v1alpha1.CommonNode{
							Name: "source-node",
							Host: "192.168.1.10",
							Port: 6379,
						},
					},
					Replica: v1alpha1.RedisReplicaNodes{
						&v1alpha1.RedisReplicaNode{
							RedisNode: v1alpha1.RedisNode{
								CommonNode: v1alpha1.CommonNode{
									Name: "duplicate-replica", // Same name as next replica
									Host: "192.168.1.11",
									Port: 6379,
								},
								AnnounceHost: "",
								AnnouncePort: 0,
							},
							Isolated: false,
						},
					},
				},
			}

			warnings, err := rra.ValidateUpdate(context.TODO(), oldRedisRepl, newRedisRepl)
			Expect(err).NotTo(HaveOccurred())
			Expect(warnings).To(BeEmpty()) // No warnings for safe changes
		})
	})

	Context("When deleting RedisReplication under Validating Webhook", func() {
		It("Should return warning about cleanup", func() {
			redisRepl := &v1alpha1.RedisReplication{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test-redis-repl",
					Namespace: "default",
				},
			}

			warnings, err := rra.ValidateDelete(context.TODO(), redisRepl)
			Expect(err).NotTo(HaveOccurred())
			Expect(warnings).To(ContainElement(ContainSubstring("Deleting RedisReplication will stop managing replication")))
		})
	})

	Context("Hostname validation", func() {
		DescribeTable("Valid hostnames",
			func(hostname string) {
				Expect(isValidHostname(hostname)).To(BeTrue())
			},
			Entry("simple hostname", "localhost"),
			Entry("FQDN", "example.com"),
			Entry("subdomain", "redis.example.com"),
			Entry("hyphenated", "redis-server.example.com"),
			Entry("numbers", "redis01.example.com"),
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
