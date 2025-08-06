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
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"os"
)

const (
	// Environment variable name for AES key
	AESKeyEnvVar = "AES_SECRET_KEY"
)

var (
	// Global AES key that should be set during application startup
	aesKey string
)

// ValidateAndSetAESKey validates the AES key from environment variable and sets it for use
// This function should be called during application startup (e.g., in main.go)
// Returns error if key is missing or invalid
func ValidateAndSetAESKey() error {
	key := os.Getenv(AESKeyEnvVar)
	if key == "" {
		return fmt.Errorf("AES encryption key not found in environment variable %s", AESKeyEnvVar)
	}

	// Validate key length (should be 32 characters for AES-256)
	if len(key) != 32 {
		return fmt.Errorf("invalid AES key length: expected 32 characters, got %d. Key: %s", len(key), key)
	}

	aesKey = key
	return nil
}

func getAESKey() (string, error) {
	if aesKey == "" {
		return "", fmt.Errorf("cannot get AES key")
	}
	return aesKey, nil
}


// AES_CTR_EncryptToBytes encrypts plaintext and returns raw bytes (for file storage)
func AES_CTR_EncryptToBytes(plainText []byte) ([]byte, error) {
	keyStr, err := getAESKey()
	if err != nil {
		return nil, err
	}

	// Convert key to OpenSSL compatible format
	opensslKey := hex.EncodeToString([]byte(keyStr))
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

// AES_CTR_DecryptFromBytes decrypts raw bytes and returns plaintext (for file decryption)
func AES_CTR_DecryptFromBytes(encryptedData []byte) ([]byte, error) {
	keyStr, err := getAESKey()
	if err != nil {
		return nil, err
	}

	// Convert key to OpenSSL compatible format
	opensslKey := hex.EncodeToString([]byte(keyStr))
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

// AES_CTR_Encrypt encrypts plaintext and returns base64 encoded string (for backward compatibility)
func AES_CTR_Encrypt(plainText []byte) (string, error) {
	encryptedData, err := AES_CTR_EncryptToBytes(plainText)
	if err != nil {
		return "", err
	}
	return base64.StdEncoding.EncodeToString(encryptedData), nil
}

// AES_CTR_Decrypt decrypts base64 encoded string and returns plaintext (for backward compatibility)
func AES_CTR_Decrypt(base64EncryptedData string) ([]byte, error) {
	encryptedData, err := base64.StdEncoding.DecodeString(base64EncryptedData)
	if err != nil {
		return nil, err
	}
	return AES_CTR_DecryptFromBytes(encryptedData)
}
