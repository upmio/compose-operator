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

package redisutil

import (
	"errors"
	"testing"
)

func TestErrorString(t *testing.T) {
	err := Error("test error")
	expected := "test error"
	if err.Error() != expected {
		t.Errorf("expected %s, got %s", expected, err.Error())
	}
}

func TestIsNodeNotFoundedError(t *testing.T) {
	err := nodeNotFoundedError
	if !IsNodeNotFoundedError(err) {
		t.Errorf("expected true, got false")
	}

	otherErr := errors.New("other error")
	if IsNodeNotFoundedError(otherErr) {
		t.Errorf("expected false, got true")
	}
}

func TestNewClusterInfosError(t *testing.T) {
	err := NewClusterInfosError()
	if err.errs == nil || err.partial || err.inconsistent {
		t.Errorf("expected initialized ClusterInfosError, got %v", err)
	}
}

func TestClusterInfosError_Error(t *testing.T) {
	err := NewClusterInfosError()
	if err.Error() != "" {
		t.Errorf("expected empty string, got %s", err.Error())
	}

	err.partial = true
	err.errs["127.0.0.1"] = errors.New("some error")
	expected := "Cluster infos partial: 127.0.0.1: 'some error'"
	if err.Error() != expected {
		t.Errorf("expected %s, got %s", expected, err.Error())
	}

	err = NewClusterInfosError()
	err.inconsistent = true
	expected = "Cluster view is inconsistent"
	if err.Error() != expected {
		t.Errorf("expected %s, got %s", expected, err.Error())
	}
}

func TestClusterInfosError_Partial(t *testing.T) {
	err := NewClusterInfosError()
	if err.Partial() {
		t.Errorf("expected false, got true")
	}

	err.partial = true
	if !err.Partial() {
		t.Errorf("expected true, got false")
	}
}

func TestClusterInfosError_Inconsistent(t *testing.T) {
	err := NewClusterInfosError()
	if err.Inconsistent() {
		t.Errorf("expected false, got true")
	}

	err.inconsistent = true
	if !err.Inconsistent() {
		t.Errorf("expected true, got false")
	}
}

func TestIsPartialError(t *testing.T) {
	err := NewClusterInfosError()
	if IsPartialError(err) {
		t.Errorf("expected false, got true")
	}

	err.partial = true
	if !IsPartialError(err) {
		t.Errorf("expected true, got false")
	}

	otherErr := errors.New("other error")
	if IsPartialError(otherErr) {
		t.Errorf("expected false, got true")
	}
}

func TestIsInconsistentError(t *testing.T) {
	err := NewClusterInfosError()
	if IsInconsistentError(err) {
		t.Errorf("expected false, got true")
	}

	err.inconsistent = true
	if !IsInconsistentError(err) {
		t.Errorf("expected true, got false")
	}

	otherErr := errors.New("other error")
	if IsInconsistentError(otherErr) {
		t.Errorf("expected false, got true")
	}
}
