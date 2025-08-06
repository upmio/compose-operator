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

// TestRound test to round.
func TestRound(t *testing.T) {
	tests := []struct {
		name     string
		input    float64
		expected int
	}{
		{name: "Case 1", input: 1.5, expected: 2},
		{name: "Case 2", input: 2.5, expected: 3},
		{name: "Case 3", input: -1.5, expected: -2},
		{name: "Case 4", input: -2.5, expected: -3},
		{name: "Case 5", input: 0.5, expected: 1},
		{name: "Case 6", input: -0.5, expected: -1},
		{name: "Case 7", input: 0.49, expected: 0},
		{name: "Case 8", input: -0.49, expected: 0},
		{name: "Case 9", input: 1.49, expected: 1},
		{name: "Case 10", input: -1.49, expected: -1},
		{name: "Case 11", input: 2.51, expected: 3},
		{name: "Case 12", input: -2.51, expected: -3},
		{name: "Case 13", input: 0.0, expected: 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := Round(tt.input)
			if result != tt.expected {
				t.Errorf("Round(%v) = %v; want %v", tt.input, result, tt.expected)
			}
		})
	}
}
