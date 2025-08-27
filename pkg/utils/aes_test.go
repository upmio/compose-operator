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
	"testing"
)

// Test AES_CTR_Encrypt and AES_CTR_Decrypt methods
func TestAES_CTR_EncryptDecrypt(t *testing.T) {
	aesKey := "bec62eddcb834ece8488c88263a5f248"
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "Encrypt and decrypt a simple string",
			input:    "Hello, World!",
			expected: "Hello, World!",
		},
		{
			name:     "Encrypt and decrypt empty string",
			input:    "",
			expected: "",
		},
		{
			name:     "Encrypt and decrypt long string",
			input:    "This is a very long test string used to verify that AES-CTR encryption and decryption work correctly!",
			expected: "This is a very long test string used to verify that AES-CTR encryption and decryption work correctly!",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Encrypt
			encrypted, err := AES_CTR_Encrypt([]byte(tt.input), aesKey)
			if err != nil {
				t.Errorf("%s: encryption failed: %v", tt.name, err)
				return
			}

			// Decrypt
			decrypted, err := AES_CTR_Decrypt(encrypted, aesKey)
			if err != nil {
				t.Errorf("%s: decryption failed: %v", tt.name, err)
				return
			}

			// Verify result
			if string(decrypted) != tt.expected {
				t.Errorf("%s: decryption result mismatch\nExpected: %s\nActual: %s", tt.name, tt.expected, string(decrypted))
			}
		})
	}
}
