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

	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"
)

// TestCompareInt32Value test to compare int32 value.
func TestCompareInt32Value(t *testing.T) {
	ctrl.SetLogger(zap.New(zap.UseFlagOptions(&zap.Options{
		Development: true,
	})))

	tests := []struct {
		// name of the test (for helpful errors)
		name string

		// old integer input for the function
		oldInt int32

		// new integer input for the function
		newInt int32

		// expected result for the function
		expected bool
	}{
		{
			name:     "Compare different integer",
			oldInt:   5,
			newInt:   10,
			expected: true,
		},
		{
			name:     "Compare same integer",
			oldInt:   5,
			newInt:   5,
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := CompareInt32("integer", tt.oldInt, tt.newInt, ctrl.Log.WithName("CompareInt32Value"))
			if result != tt.expected {
				t.Errorf("Compare different integer test failed.\nExpected: %t\nActual: %t\n", tt.expected, result)
			}
		})
	}
}

// TestCompareStringValue test to compare string value.
func TestCompareStringValue(t *testing.T) {
	ctrl.SetLogger(zap.New(zap.UseFlagOptions(&zap.Options{
		Development: true,
	})))

	tests := []struct {
		// name of the test (for helpful errors)
		name string

		// old string input for the function
		oldString string

		// new string input for the function
		newString string

		// expected result for the function
		expected bool
	}{
		{
			name:      "Compare different integer",
			oldString: "Compare",
			newString: "compare",
			expected:  true,
		},
		{
			name:      "Compare same integer",
			oldString: "Compare",
			newString: "Compare",
			expected:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := CompareStringValue("integer", tt.oldString, tt.newString, ctrl.Log.WithName("CompareStringValue"))
			if result != tt.expected {
				t.Errorf("Compare different string test failed.\nExpected: %t\nActual: %t\n", tt.expected, result)
			}
		})
	}
}
