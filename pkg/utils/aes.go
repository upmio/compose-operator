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
	"encoding/hex"
	"fmt"
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
