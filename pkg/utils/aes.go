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

package utils

import (
	"context"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"github.com/go-logr/logr"
	"io"
	corev1 "k8s.io/api/core/v1"
	kubeErrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"os"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"time"
)

const (
	AESSecretNameENVKey      = "AES_SECRET_NAME"
	AESSecretNamespaceENVKey = "NAMESPACE"

	defaultAESSecretNamespace = "upm-system"
	defaultAESSecretName      = "aes-secret-key"
	defaultAESSecretKey       = "AES_SECRET_KEY"
)

// AES_CTR_Encrypt encrypts plaintext and returns base64 encoded string (for backward compatibility)
func AES_CTR_Encrypt(plainText []byte, aesKey string) ([]byte, error) {

	// Convert key to OpenSSL compatible format
	opensslKey := hex.EncodeToString([]byte(aesKey))
	key, err := hex.DecodeString(opensslKey)
	if err != nil {
		return nil, err
	}

	// Create AES cipher block
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}

	// Generate random IV
	iv := make([]byte, aes.BlockSize)
	if _, err := rand.Read(iv); err != nil {
		return nil, err
	}

	// Create CTR mode stream cipher
	stream := cipher.NewCTR(block, iv)

	// Encrypt data
	ciphertext := make([]byte, len(plainText))
	stream.XORKeyStream(ciphertext, plainText)

	// Combine IV and ciphertext
	encryptedData := append(iv, ciphertext...)

	return encryptedData, nil
}

// AES_CTR_Decrypt decrypts base64 encoded string and returns plaintext (for backward compatibility)
func AES_CTR_Decrypt(encryptedData []byte, aesKey string) ([]byte, error) {

	// Convert key to OpenSSL compatible format
	opensslKey := hex.EncodeToString([]byte(aesKey))
	key, err := hex.DecodeString(opensslKey)
	if err != nil {
		return nil, err
	}

	// Check minimum length (at least 16 bytes for IV)
	if len(encryptedData) < aes.BlockSize {
		return nil, fmt.Errorf("encrypted data too short")
	}

	// Extract IV (first 16 bytes)
	iv := encryptedData[:aes.BlockSize]

	// Extract ciphertext (remaining part)
	ciphertext := encryptedData[aes.BlockSize:]

	// Create AES cipher block
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}

	// Create CTR mode stream cipher
	stream := cipher.NewCTR(block, iv)

	// Decrypt data
	plaintext := make([]byte, len(ciphertext))
	stream.XORKeyStream(plaintext, ciphertext)

	return plaintext, nil
}

type Decryptor interface {
	Decrypt(context.Context, []byte) ([]byte, error)
}
type SecretDecryptor struct {
	aesSecretNamespace string
	aesSecretName      string
	c                  client.Client
	reqlogger          logr.Logger
}

func NewSecretDecyptor(client client.Client, reqlooger logr.Logger) Decryptor {
	aesSecretName := os.Getenv(AESSecretNameENVKey)
	if aesSecretName == "" {
		reqlooger.Info("No AES secret name specified, using default")
		aesSecretName = defaultAESSecretName
	}

	aesSecretNamespace := os.Getenv(AESSecretNamespaceENVKey)
	if aesSecretNamespace == "" {
		reqlooger.Info("No AES secret namespace specified, using default")
		aesSecretNamespace = defaultAESSecretNamespace
	}

	return &SecretDecryptor{
		reqlogger:          reqlooger,
		c:                  client,
		aesSecretName:      aesSecretName,
		aesSecretNamespace: aesSecretNamespace,
	}
}

func (d *SecretDecryptor) Decrypt(ctx context.Context, encryptedData []byte) ([]byte, error) {
	aesKey, err := d.getAesSecret(ctx)
	if err != nil {
		return nil, err
	}
	return AES_CTR_Decrypt(encryptedData, aesKey)
}

func (d *SecretDecryptor) getAesSecret(ctx context.Context) (string, error) {
	secret := &corev1.Secret{}
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	err := d.c.Get(ctx, types.NamespacedName{
		Name:      d.aesSecretName,
		Namespace: d.aesSecretNamespace,
	}, secret)
	if err != nil {
		if kubeErrors.IsNotFound(err) {
			return d.createAesSecret(ctx)
		}
		return "", fmt.Errorf("failed to fetch aes secret [%s]: %v", d.aesSecretName, err)
	}

	aeskey, ok := secret.Data[defaultAESSecretKey]
	if !ok {
		return d.updateAesSecret(ctx, secret)
	}

	if len(aeskey) != 32 {
		return "", fmt.Errorf("aes secret key length does not match expected block size")
	}

	return string(aeskey), nil
}

func (d *SecretDecryptor) createAesSecret(ctx context.Context) (string, error) {
	data, err := d.generateAES256Key()
	if err != nil {
		return "", err
	}
	secret := &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      d.aesSecretName,
			Namespace: d.aesSecretNamespace,
		},
		Data: map[string][]byte{
			defaultAESSecretKey: data,
		},
		Type: corev1.SecretTypeOpaque,
	}

	err = d.c.Create(ctx, secret)
	return string(data), err
}

func (d *SecretDecryptor) updateAesSecret(ctx context.Context, secret *corev1.Secret) (string, error) {
	data, err := d.generateAES256Key()
	if err != nil {
		return "", err
	}

	secret.Data[defaultAESSecretKey] = data

	err = d.c.Update(ctx, secret)
	return string(data), err
}

// generateAES256KeyAndIV generate AES-256 key
func (d *SecretDecryptor) generateAES256Key() (key []byte, err error) {
	key = make([]byte, 32)
	if _, err = io.ReadFull(rand.Reader, key); err != nil {
		return nil, err
	}

	return key, nil
}
