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
	"os"
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"

	"github.com/upmio/compose-operator/pkg/utils"
)

func TestSecret(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "K8sutil Secret Suite")
}

var _ = Describe("Secret Decryption Functions", func() {
	var (
		k8sClient client.Client
		scheme    *runtime.Scheme
		ctx       context.Context
		testKey   = "bec62eddcb834ece8488c88263a5f248"
	)

	BeforeEach(func() {
		ctx = context.Background()
		scheme = runtime.NewScheme()
		Expect(corev1.AddToScheme(scheme)).To(Succeed())
		k8sClient = fake.NewClientBuilder().WithScheme(scheme).Build()

		// Setup AES key for testing
		err := os.Setenv(utils.AESKeyEnvVar, testKey)
		Expect(err).ShouldNot(HaveOccurred())
		err = utils.ValidateAndSetAESKey()
		Expect(err).ShouldNot(HaveOccurred())
	})

	Describe("DecryptSecretPasswords", func() {
		It("should decrypt multiple passwords from a secret", func() {
			// Create test secret with encrypted passwords
			plaintext1 := "password1"
			plaintext2 := "password2"

			encrypted1, err := utils.AES_CTR_Encrypt([]byte(plaintext1))
			Expect(err).ShouldNot(HaveOccurred())

			encrypted2, err := utils.AES_CTR_Encrypt([]byte(plaintext2))
			Expect(err).ShouldNot(HaveOccurred())

			secret := &corev1.Secret{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test-secret",
					Namespace: "default",
				},
				Data: map[string][]byte{
					"key1": encrypted1,
					"key2": encrypted2,
				},
			}

			err = k8sClient.Create(ctx, secret)
			Expect(err).ShouldNot(HaveOccurred())

			// Test decryption
			passwords, err := DecryptSecretPasswords(k8sClient, "test-secret", "default", []string{"key1", "key2"})
			Expect(err).ShouldNot(HaveOccurred())
			Expect(passwords).To(HaveLen(2))
			Expect(passwords["key1"]).To(Equal(plaintext1))
			Expect(passwords["key2"]).To(Equal(plaintext2))
		})

		It("should return error when secret does not exist", func() {
			passwords, err := DecryptSecretPasswords(k8sClient, "nonexistent-secret", "default", []string{"key1"})
			Expect(err).Should(HaveOccurred())
			Expect(passwords).To(BeNil())
			Expect(err.Error()).To(ContainSubstring("failed to fetch secret"))
		})

		It("should return error when key does not exist in secret", func() {
			secret := &corev1.Secret{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test-secret",
					Namespace: "default",
				},
				Data: map[string][]byte{
					"existing-key": []byte("encrypted-data"),
				},
			}

			err := k8sClient.Create(ctx, secret)
			Expect(err).ShouldNot(HaveOccurred())

			passwords, err := DecryptSecretPasswords(k8sClient, "test-secret", "default", []string{"nonexistent-key"})
			Expect(err).Should(HaveOccurred())
			Expect(passwords).To(BeNil())
			Expect(err.Error()).To(ContainSubstring("failed to decrypt secret"))
		})
	})

	Describe("Real-world usage examples", func() {
		It("should decrypt single redis password", func() {
			plaintext := "redis-password"
			encrypted, err := utils.AES_CTR_Encrypt([]byte(plaintext))
			Expect(err).ShouldNot(HaveOccurred())

			secret := &corev1.Secret{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "redis-secret",
					Namespace: "default",
				},
				Data: map[string][]byte{
					"redis": encrypted,
				},
			}

			err = k8sClient.Create(ctx, secret)
			Expect(err).ShouldNot(HaveOccurred())

			// Use the generic function as controllers do
			passwords, err := DecryptSecretPasswords(k8sClient, "redis-secret", "default", []string{"redis"})
			Expect(err).ShouldNot(HaveOccurred())
			Expect(passwords["redis"]).To(Equal(plaintext))
		})

		It("should decrypt mysql and replication passwords", func() {
			mysqlPassword := "mysql-password"
			replPassword := "replication-password"

			encryptedMysql, err := utils.AES_CTR_Encrypt([]byte(mysqlPassword))
			Expect(err).ShouldNot(HaveOccurred())

			encryptedRepl, err := utils.AES_CTR_Encrypt([]byte(replPassword))
			Expect(err).ShouldNot(HaveOccurred())

			secret := &corev1.Secret{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "mysql-secret",
					Namespace: "default",
				},
				Data: map[string][]byte{
					"mysql":       encryptedMysql,
					"replication": encryptedRepl,
				},
			}

			err = k8sClient.Create(ctx, secret)
			Expect(err).ShouldNot(HaveOccurred())

			// Use the generic function as controllers do
			passwords, err := DecryptSecretPasswords(k8sClient, "mysql-secret", "default", []string{"mysql", "replication"})
			Expect(err).ShouldNot(HaveOccurred())
			Expect(passwords["mysql"]).To(Equal(mysqlPassword))
			Expect(passwords["replication"]).To(Equal(replPassword))
		})
	})
})
