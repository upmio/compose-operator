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

// TestStringer is a simple implementation of the Stringer interface for testing.
type TestStringer struct {
	value string
}

func (ts TestStringer) String() string {
	return ts.value
}

func TestSliceJoin(t *testing.T) {
	tests := []struct {
		name     string
		input    []Stringer
		sep      string
		expected string
	}{
		{
			name:     "Empty slice",
			input:    []Stringer{},
			sep:      ", ",
			expected: "",
		},
		{
			name:     "Single element",
			input:    []Stringer{TestStringer{"a"}},
			sep:      ", ",
			expected: "a",
		},
		{
			name:     "Two elements",
			input:    []Stringer{TestStringer{"a"}, TestStringer{"b"}},
			sep:      ", ",
			expected: "a, b",
		},
		{
			name:     "Three elements",
			input:    []Stringer{TestStringer{"a"}, TestStringer{"b"}, TestStringer{"c"}},
			sep:      ", ",
			expected: "a, b, c",
		},
		{
			name:     "Multiple elements",
			input:    []Stringer{TestStringer{"a"}, TestStringer{"b"}, TestStringer{"c"}, TestStringer{"d"}},
			sep:      ", ",
			expected: "a, b, c, d",
		},
		{
			name:     "Custom separator",
			input:    []Stringer{TestStringer{"a"}, TestStringer{"b"}, TestStringer{"c"}},
			sep:      " - ",
			expected: "a - b - c",
		},
		{
			name:     "Empty string elements",
			input:    []Stringer{TestStringer{""}, TestStringer{"b"}, TestStringer{""}},
			sep:      ", ",
			expected: ", b, ",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := SliceJoin(tt.input, tt.sep)
			if result != tt.expected {
				t.Errorf("SliceJoin() = %v; want %v", result, tt.expected)
			}
		})
	}
}
