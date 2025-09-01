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
	"fmt"
	"time"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/upmio/compose-operator/pkg/utils"
)

// DecryptSecretPasswords decrypts multiple passwords from a Kubernetes Secret, returning a map of key->password
func DecryptSecretPasswords(client client.Client, secretName, namespace, aesSecretName, aesSecretKey string, keys []string) (map[string]string, error) {

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	aesKey, err := getAesKey(ctx, client, aesSecretName, namespace, aesSecretKey)
	if err != nil {
		return nil, err
	}

	secret := &corev1.Secret{}

	err = client.Get(ctx, types.NamespacedName{
		Name:      secretName,
		Namespace: namespace,
	}, secret)

	if err != nil {
		return nil, fmt.Errorf("failed to fetch secret [%s]: %v", secretName, err)
	}

	passwords := make(map[string]string)
	for _, key := range keys {

		decrypted, err := utils.AES_CTR_Decrypt(secret.Data[key], aesKey)
		if err != nil {
			return nil, fmt.Errorf("failed to decrypt secret [%s] key '%s': %v", secretName, key, err)
		}
		passwords[key] = string(decrypted)
	}

	return passwords, nil
}

func getAesKey(ctx context.Context, c client.Client, aesSecretName, namespace, aesSecretKey string) (string, error) {
	secret := &corev1.Secret{}

	err := c.Get(ctx, types.NamespacedName{
		Name:      aesSecretName,
		Namespace: namespace,
	}, secret)
	if err != nil {
		return "", fmt.Errorf("failed to fetch aes secret [%s]: %v", aesSecretName, err)
	}

	aesKey, ok := secret.Data[aesSecretKey]
	if !ok {
		return "", fmt.Errorf("failed to found key %s from aes secret [%s]", aesSecretKey, aesSecretName)
	}

	if len(aesKey) != 32 {
		return "", fmt.Errorf("aes secret key length does not match expected block size")
	}

	return string(aesKey), nil
}
