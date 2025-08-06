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

// Stringer implement the string interface
type Stringer interface {
	String() string
}

// SliceJoin concatenates the elements of a to create a single string. The separator string
// sep is placed between elements in the resulting string.
func SliceJoin(a []Stringer, sep string) string {
	switch len(a) {
	case 0:
		return ""
	case 1:
		return a[0].String()
	case 2:
		// Special case for common small values.
		// Remove if golang.org/issue/6714 is fixed
		return a[0].String() + sep + a[1].String()
	case 3:
		// Special case for common small values.
		// Remove if golang.org/issue/6714 is fixed
		return a[0].String() + sep + a[1].String() + sep + a[2].String()
	}
	n := len(sep) * (len(a) - 1)
	for i := 0; i < len(a); i++ {
		n += len(a[i].String())
	}

	b := make([]byte, n)
	bp := copy(b, a[0].String())
	for _, s := range a[1:] {
		bp += copy(b[bp:], sep)
		bp += copy(b[bp:], s.String())
	}
	return string(b)
}
