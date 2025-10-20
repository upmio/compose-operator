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
	"fmt"

	"github.com/go-logr/logr"
)

func CompareInt32(name string, old, new int32, reqLogger logr.Logger) bool {
	if old != new {
		reqLogger.V(4).Info(fmt.Sprintf("compare status.%s: %d - %d", name, old, new))
		return true
	}

	return false
}

func CompareInt64(name string, old, new int64, reqLogger logr.Logger) bool {
	if old != new {
		reqLogger.V(4).Info(fmt.Sprintf("compare status.%s: %d - %d", name, old, new))
		return true
	}

	return false
}

func CompareIntPoint(name string, old, new *int, reqLogger logr.Logger) bool {
	if old == nil || new == nil {
		return false
	}

	if *old != *new {
		reqLogger.V(4).Info(fmt.Sprintf("compare status.%s: %d - %d", name, old, new))
		return true
	}

	return false
}

func CompareBoolPoint(name string, old, new *bool, reqLogger logr.Logger) bool {
	if old == nil || new == nil {
		return false
	}

	if *old != *new {
		reqLogger.V(4).Info(fmt.Sprintf("compare status.%s: %d - %d", name, old, new))
		return true
	}

	return false
}

func CompareStringValue(name string, old, new string, reqLogger logr.Logger) bool {
	if old != new {
		reqLogger.V(4).Info(fmt.Sprintf("compare %s: %s - %s", name, old, new))
		return true
	}

	return false
}
