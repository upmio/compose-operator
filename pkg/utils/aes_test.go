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
	"os"
	"testing"
)

// Test AES_CTR_Encrypt and AES_CTR_Decrypt methods
func TestAES_CTR_EncryptDecrypt(t *testing.T) {
	aesKey = "bec62eddcb834ece8488c88263a5f248"
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
			encrypted, err := AES_CTR_Encrypt([]byte(tt.input))
			if err != nil {
				t.Errorf("%s: encryption failed: %v", tt.name, err)
				return
			}

			// Decrypt
			decrypted, err := AES_CTR_Decrypt(encrypted)
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

// TestValidateAndSetAESKey tests the ValidateAndSetAESKey function
func TestValidateAndSetAESKey(t *testing.T) {
	// Save original environment to restore later
	originalKey := os.Getenv(AESKeyEnvVar)
	defer func() {
		if originalKey != "" {
			_ = os.Setenv(AESKeyEnvVar, originalKey)
		} else {
			_ = os.Unsetenv(AESKeyEnvVar)
		}
		// Reset global aesKey
		aesKey = ""
	}()

	tests := []struct {
		name          string
		envKeyValue   string
		shouldSetEnv  bool
		expectError   bool
		expectedError string
	}{
		{
			name:         "Valid 32-character key",
			envKeyValue:  "bec62eddcb834ece8488c88263a5f248",
			shouldSetEnv: true,
			expectError:  false,
		},
		{
			name:          "Missing environment variable",
			shouldSetEnv:  false,
			expectError:   true,
			expectedError: "AES encryption key not found",
		},
		{
			name:          "Invalid key length - too short",
			envKeyValue:   "shortkey",
			shouldSetEnv:  true,
			expectError:   true,
			expectedError: "invalid AES key length",
		},
		{
			name:          "Invalid key length - too long",
			envKeyValue:   "verylongkeythatismorethan32characters",
			shouldSetEnv:  true,
			expectError:   true,
			expectedError: "invalid AES key length",
		},
		{
			name:          "Empty string key",
			envKeyValue:   "",
			shouldSetEnv:  true,
			expectError:   true,
			expectedError: "AES encryption key not found",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Reset global aesKey for each test
			aesKey = ""

			// Setup environment
			if tt.shouldSetEnv {
				_ = os.Setenv(AESKeyEnvVar, tt.envKeyValue)
			} else {
				_ = os.Unsetenv(AESKeyEnvVar)
			}

			// Test the function
			err := ValidateAndSetAESKey()

			// Check error expectation
			if tt.expectError {
				if err == nil {
					t.Errorf("Expected error but got none")
				} else if tt.expectedError != "" && !containsString(err.Error(), tt.expectedError) {
					t.Errorf("Expected error containing '%s', got '%s'", tt.expectedError, err.Error())
				}
			} else {
				if err != nil {
					t.Errorf("Expected no error but got: %v", err)
				}
				// For successful case, verify the key was set
				if aesKey != tt.envKeyValue {
					t.Errorf("Expected aesKey to be '%s', got '%s'", tt.envKeyValue, aesKey)
				}
			}
		})
	}
}

// TestGetAESKey tests the getAESKey function
func TestGetAESKey(t *testing.T) {
	tests := []struct {
		name        string
		setupKey    string
		expectError bool
		expectedKey string
	}{
		{
			name:        "Valid key set",
			setupKey:    "bec62eddcb834ece8488c88263a5f248",
			expectError: false,
			expectedKey: "bec62eddcb834ece8488c88263a5f248",
		},
		{
			name:        "No key set",
			setupKey:    "",
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup
			aesKey = tt.setupKey

			// Test
			key, err := getAESKey()

			// Verify
			if tt.expectError {
				if err == nil {
					t.Errorf("Expected error but got none")
				}
			} else {
				if err != nil {
					t.Errorf("Expected no error but got: %v", err)
				}
				if key != tt.expectedKey {
					t.Errorf("Expected key '%s', got '%s'", tt.expectedKey, key)
				}
			}
		})
	}
}

// Helper function to check if a string contains a substring
func containsString(s, substr string) bool {
	return len(s) >= len(substr) &&
		(len(substr) == 0 ||
			len(s) > 0 && (s[:len(substr)] == substr ||
				(len(s) > len(substr) && containsString(s[1:], substr))))
}
