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
	"context"
	"github.com/upmio/compose-operator/api/v1alpha1"
	"strings"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var _ = Describe("RedisCluster Webhook", func() {

	var rca = redisClusterAdmission{}

	Context("When creating RedisCluster under Defaulting Webhook", func() {
		It("Should fill in the default value for redis secret key", func() {
			redisCluster := &v1alpha1.RedisCluster{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test-redis-cluster",
					Namespace: "default",
				},
				Spec: v1alpha1.RedisClusterSpec{
					Secret: v1alpha1.RedisClusterSecret{
						Name: "test-secret",
						// redis key is empty - should get default
					},
					Members: map[string]v1alpha1.CommonNodes{
						"shard1": {
							{
								Name: "redis-shard1-1",
								Host: "192.168.1.10",
								Port: 6379,
							},
						},
					},
				},
			}

			err := rca.Default(context.TODO(), redisCluster)
			Expect(err).NotTo(HaveOccurred())

			// Check default was set
			Expect(redisCluster.Spec.Secret.Redis).To(Equal("redis"))
		})

		It("Should not override existing redis secret key value", func() {
			redisCluster := &v1alpha1.RedisCluster{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test-redis-cluster",
					Namespace: "default",
				},
				Spec: v1alpha1.RedisClusterSpec{
					Secret: v1alpha1.RedisClusterSecret{
						Name:  "test-secret",
						Redis: "custom-redis",
					},
					Members: map[string]v1alpha1.CommonNodes{
						"shard1": {
							{
								Name: "redis-shard1-1",
								Host: "192.168.1.10",
								Port: 6379,
							},
						},
					},
				},
			}

			err := rca.Default(context.TODO(), redisCluster)
			Expect(err).NotTo(HaveOccurred())

			// Check existing value was preserved
			Expect(redisCluster.Spec.Secret.Redis).To(Equal("custom-redis"))
		})
	})

	Context("When creating RedisCluster under Validating Webhook", func() {
		It("Should deny if members are empty", func() {
			redisCluster := &v1alpha1.RedisCluster{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test-redis-cluster",
					Namespace: "default",
				},
				Spec: v1alpha1.RedisClusterSpec{
					Secret: v1alpha1.RedisClusterSecret{
						Name:  "test-secret",
						Redis: "redis",
					},
					Members: map[string]v1alpha1.CommonNodes{},
				},
			}

			_, err := rca.ValidateCreate(context.TODO(), redisCluster)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("at least one shard with members is required"))
		})

		It("Should deny if shard has no members", func() {
			redisCluster := &v1alpha1.RedisCluster{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test-redis-cluster",
					Namespace: "default",
				},
				Spec: v1alpha1.RedisClusterSpec{
					Secret: v1alpha1.RedisClusterSecret{
						Name:  "test-secret",
						Redis: "redis",
					},
					Members: map[string]v1alpha1.CommonNodes{
						"empty-shard": {},
					},
				},
			}

			_, err := rca.ValidateCreate(context.TODO(), redisCluster)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("shard 'empty-shard' must have at least one member"))
		})

		It("Should deny if secret name is empty", func() {
			redisCluster := &v1alpha1.RedisCluster{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test-redis-cluster",
					Namespace: "default",
				},
				Spec: v1alpha1.RedisClusterSpec{
					Secret: v1alpha1.RedisClusterSecret{
						Name:  "", // Empty secret name
						Redis: "redis",
					},
					Members: map[string]v1alpha1.CommonNodes{
						"shard1": {
							{
								Name: "redis-shard1-1",
								Host: "192.168.1.10",
								Port: 6379,
							},
						},
					},
				},
			}

			_, err := rca.ValidateCreate(context.TODO(), redisCluster)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("secret name is required"))
		})

		It("Should deny if redis secret key is empty", func() {
			redisCluster := &v1alpha1.RedisCluster{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test-redis-cluster",
					Namespace: "default",
				},
				Spec: v1alpha1.RedisClusterSpec{
					Secret: v1alpha1.RedisClusterSecret{
						Name:  "test-secret",
						Redis: "", // Empty redis key
					},
					Members: map[string]v1alpha1.CommonNodes{
						"shard1": {
							{
								Name: "redis-shard1-1",
								Host: "192.168.1.10",
								Port: 6379,
							},
						},
					},
				},
			}

			_, err := rca.ValidateCreate(context.TODO(), redisCluster)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("redis secret key is required"))
		})

		It("Should deny if node name is empty", func() {
			redisCluster := &v1alpha1.RedisCluster{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test-redis-cluster",
					Namespace: "default",
				},
				Spec: v1alpha1.RedisClusterSpec{
					Secret: v1alpha1.RedisClusterSecret{
						Name:  "test-secret",
						Redis: "redis",
					},
					Members: map[string]v1alpha1.CommonNodes{
						"shard1": {
							{
								Name: "", // Empty node name
								Host: "192.168.1.10",
								Port: 6379,
							},
						},
					},
				},
			}

			_, err := rca.ValidateCreate(context.TODO(), redisCluster)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("redis node name is required"))
		})

		It("Should deny if node host is empty", func() {
			redisCluster := &v1alpha1.RedisCluster{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test-redis-cluster",
					Namespace: "default",
				},
				Spec: v1alpha1.RedisClusterSpec{
					Secret: v1alpha1.RedisClusterSecret{
						Name:  "test-secret",
						Redis: "redis",
					},
					Members: map[string]v1alpha1.CommonNodes{
						"shard1": {
							{
								Name: "redis-shard1-1",
								Host: "", // Empty host
								Port: 6379,
							},
						},
					},
				},
			}

			_, err := rca.ValidateCreate(context.TODO(), redisCluster)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("redis node host is required"))
		})

		It("Should deny if node port is invalid", func() {
			redisCluster := &v1alpha1.RedisCluster{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test-redis-cluster",
					Namespace: "default",
				},
				Spec: v1alpha1.RedisClusterSpec{
					Secret: v1alpha1.RedisClusterSecret{
						Name:  "test-secret",
						Redis: "redis",
					},
					Members: map[string]v1alpha1.CommonNodes{
						"shard1": {
							{
								Name: "redis-shard1-1",
								Host: "192.168.1.10",
								Port: 99999, // Invalid port
							},
						},
					},
				},
			}

			_, err := rca.ValidateCreate(context.TODO(), redisCluster)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("port must be between 1 and 65535"))
		})

		It("Should deny if node host is invalid", func() {
			redisCluster := &v1alpha1.RedisCluster{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test-redis-cluster",
					Namespace: "default",
				},
				Spec: v1alpha1.RedisClusterSpec{
					Secret: v1alpha1.RedisClusterSecret{
						Name:  "test-secret",
						Redis: "redis",
					},
					Members: map[string]v1alpha1.CommonNodes{
						"shard1": {
							{
								Name: "redis-shard1-1",
								Host: "invalid..host", // Invalid hostname
								Port: 6379,
							},
						},
					},
				},
			}

			_, err := rca.ValidateCreate(context.TODO(), redisCluster)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("host must be a valid IP address or hostname"))
		})

		It("Should deny if node names are duplicate across shards", func() {
			redisCluster := &v1alpha1.RedisCluster{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test-redis-cluster",
					Namespace: "default",
				},
				Spec: v1alpha1.RedisClusterSpec{
					Secret: v1alpha1.RedisClusterSecret{
						Name:  "test-secret",
						Redis: "redis",
					},
					Members: map[string]v1alpha1.CommonNodes{
						"shard1": {
							{
								Name: "duplicate-node",
								Host: "192.168.1.10",
								Port: 6379,
							},
						},
						"shard2": {
							{
								Name: "duplicate-node", // Duplicate name across shards
								Host: "192.168.1.11",
								Port: 6379,
							},
						},
					},
				},
			}

			_, err := rca.ValidateCreate(context.TODO(), redisCluster)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("node name 'duplicate-node' is already used"))
		})

		It("Should return warning for less than 3 shards", func() {
			redisCluster := &v1alpha1.RedisCluster{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test-redis-cluster",
					Namespace: "default",
				},
				Spec: v1alpha1.RedisClusterSpec{
					Secret: v1alpha1.RedisClusterSecret{
						Name:  "test-secret",
						Redis: "redis",
					},
					Members: map[string]v1alpha1.CommonNodes{
						"shard1": {
							{
								Name: "redis-shard1-1",
								Host: "192.168.1.10",
								Port: 6379,
							},
						},
						"shard2": {
							{
								Name: "redis-shard2-1",
								Host: "192.168.1.11",
								Port: 6379,
							},
						},
					},
				},
			}

			warnings, err := rca.ValidateCreate(context.TODO(), redisCluster)
			Expect(err).NotTo(HaveOccurred())
			Expect(warnings).To(ContainElement(ContainSubstring("Redis Cluster requires at least 3 shards")))
		})

		It("Should admit if all required fields are provided correctly", func() {
			redisCluster := &v1alpha1.RedisCluster{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test-redis-cluster",
					Namespace: "default",
				},
				Spec: v1alpha1.RedisClusterSpec{
					Secret: v1alpha1.RedisClusterSecret{
						Name:  "test-secret",
						Redis: "redis",
					},
					Members: map[string]v1alpha1.CommonNodes{
						"shard1": {
							{
								Name: "redis-shard1-1",
								Host: "192.168.1.10",
								Port: 6379,
							},
							{
								Name: "redis-shard1-2",
								Host: "192.168.1.11",
								Port: 6379,
							},
						},
						"shard2": {
							{
								Name: "redis-shard2-1",
								Host: "192.168.1.12",
								Port: 6379,
							},
							{
								Name: "redis-shard2-2",
								Host: "redis-shard2-2.example.com",
								Port: 6379,
							},
						},
						"shard3": {
							{
								Name: "redis-shard3-1",
								Host: "192.168.1.14",
								Port: 6379,
							},
							{
								Name: "redis-shard3-2",
								Host: "redis-shard3-2.example.com",
								Port: 6379,
							},
						},
					},
				},
			}

			warnings, err := rca.ValidateCreate(context.TODO(), redisCluster)
			Expect(err).NotTo(HaveOccurred())
			Expect(warnings).To(BeEmpty()) // No warnings expected for valid config with 3+ shards
		})
	})

	Context("When updating RedisCluster under Validating Webhook", func() {
		It("Should deny if secret name is changed", func() {
			oldRedisCluster := &v1alpha1.RedisCluster{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test-redis-cluster",
					Namespace: "default",
				},
				Spec: v1alpha1.RedisClusterSpec{
					Secret: v1alpha1.RedisClusterSecret{
						Name: "old-secret",
					},
				},
			}

			newRedisCluster := &v1alpha1.RedisCluster{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test-redis-cluster",
					Namespace: "default",
				},
				Spec: v1alpha1.RedisClusterSpec{
					Secret: v1alpha1.RedisClusterSecret{
						Name:  "new-secret", // Changed secret name
						Redis: "redis",
					},
					Members: map[string]v1alpha1.CommonNodes{
						"shard1": {
							{
								Name: "redis-shard1-1",
								Host: "192.168.1.10",
								Port: 6379,
							},
						},
					},
				},
			}

			_, err := rca.ValidateUpdate(context.TODO(), oldRedisCluster, newRedisCluster)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("secret reference cannot be changed"))
		})

		It("Should return warning when cluster members are changed", func() {
			oldRedisCluster := &v1alpha1.RedisCluster{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test-redis-cluster",
					Namespace: "default",
				},
				Spec: v1alpha1.RedisClusterSpec{
					Secret: v1alpha1.RedisClusterSecret{Name: "test-secret"},
					Members: map[string]v1alpha1.CommonNodes{
						"shard1": {
							{
								Name: "redis-shard1-1",
								Host: "192.168.1.10",
								Port: 6379,
							},
						},
					},
				},
			}

			newRedisCluster := &v1alpha1.RedisCluster{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test-redis-cluster",
					Namespace: "default",
				},
				Spec: v1alpha1.RedisClusterSpec{
					Secret: v1alpha1.RedisClusterSecret{
						Name:  "test-secret",
						Redis: "redis",
					},
					Members: map[string]v1alpha1.CommonNodes{
						"shard1": {
							{
								Name: "redis-shard1-1",
								Host: "192.168.1.10",
								Port: 6379,
							},
						},
						"shard2": {
							{
								Name: "redis-shard2-1",
								Host: "192.168.1.11",
								Port: 6379,
							},
						},
					},
				},
			}

			warnings, err := rca.ValidateUpdate(context.TODO(), oldRedisCluster, newRedisCluster)
			Expect(err).NotTo(HaveOccurred())
			Expect(warnings).To(ContainElement(ContainSubstring("Changing cluster members may cause data redistribution")))
		})

		It("Should allow update when only non-critical fields change", func() {
			oldRedisCluster := &v1alpha1.RedisCluster{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test-redis-cluster",
					Namespace: "default",
				},
				Spec: v1alpha1.RedisClusterSpec{
					Secret: v1alpha1.RedisClusterSecret{Name: "test-secret", Redis: "redis"},
					Members: map[string]v1alpha1.CommonNodes{
						"shard1": {
							{
								Name: "redis-shard1-1",
								Host: "192.168.1.10",
								Port: 6379,
							},
						},
						"shard2": {
							{
								Name: "redis-shard1-2",
								Host: "192.168.1.11",
								Port: 6379,
							},
						},
						"shard3": {
							{
								Name: "redis-shard1-3",
								Host: "192.168.1.12",
								Port: 6379,
							},
						},
					},
				},
			}

			newRedisCluster := &v1alpha1.RedisCluster{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test-redis-cluster",
					Namespace: "default",
					Labels:    map[string]string{"updated": "true"}, // Only metadata change
				},
				Spec: v1alpha1.RedisClusterSpec{
					Secret: v1alpha1.RedisClusterSecret{Name: "test-secret", Redis: "redis"}, // Same secret
					Members: map[string]v1alpha1.CommonNodes{
						"shard1": {
							{
								Name: "redis-shard1-1",
								Host: "192.168.1.10",
								Port: 6379,
							},
						},
						"shard2": {
							{
								Name: "redis-shard1-2",
								Host: "192.168.1.11",
								Port: 6379,
							},
						},
						"shard3": {
							{
								Name: "redis-shard1-3",
								Host: "192.168.1.12",
								Port: 6379,
							},
						},
					},
				},
			}

			warnings, err := rca.ValidateUpdate(context.TODO(), oldRedisCluster, newRedisCluster)
			Expect(err).NotTo(HaveOccurred())
			Expect(warnings).To(BeEmpty()) // No warnings for safe changes
		})
	})

	Context("When deleting RedisCluster under Validating Webhook", func() {
		It("Should return warning about cleanup", func() {
			redisCluster := &v1alpha1.RedisCluster{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test-redis-cluster",
					Namespace: "default",
				},
			}

			warnings, err := rca.ValidateDelete(context.TODO(), redisCluster)
			Expect(err).NotTo(HaveOccurred())
			Expect(warnings).To(ContainElement(ContainSubstring("Deleting RedisCluster will stop managing the cluster")))
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
