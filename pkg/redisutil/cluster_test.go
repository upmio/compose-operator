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
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewCluster(t *testing.T) {
	cluster := NewCluster("test-cluster", "default")
	assert.NotNil(t, cluster)
	assert.Equal(t, "test-cluster", cluster.Name)
	assert.Equal(t, "default", cluster.Namespace)
	assert.NotNil(t, cluster.Nodes)
}

func TestAddNode(t *testing.T) {
	cluster := NewCluster("test-cluster", "default")
	node := &ClusterNode{ID: "node1", IP: "192.168.1.1"}

	cluster.AddNode(node)
	assert.Equal(t, 1, len(cluster.Nodes))
	assert.Equal(t, node, cluster.Nodes["node1"])

	node2 := &ClusterNode{ID: "node1", IP: "192.168.1.2"}
	cluster.AddNode(node2)
	assert.Equal(t, 1, len(cluster.Nodes))
	assert.Equal(t, node2, cluster.Nodes["node1"])
}

func TestGetNodeByID(t *testing.T) {
	cluster := NewCluster("test-cluster", "default")
	node := &ClusterNode{ID: "node1", IP: "192.168.1.1"}

	cluster.AddNode(node)

	foundNode, err := cluster.GetNodeByID("node1")
	assert.NoError(t, err)
	assert.Equal(t, node, foundNode)

	_, err = cluster.GetNodeByID("node2")
	assert.Error(t, err)
	assert.True(t, IsNodeNotFoundedError(err))
}

func TestGetNodeByIP(t *testing.T) {
	cluster := NewCluster("test-cluster", "default")
	node := &ClusterNode{ID: "node1", IP: "192.168.1.1"}

	cluster.AddNode(node)

	foundNode, err := cluster.GetNodeByIP("192.168.1.1")
	assert.NoError(t, err)
	assert.Equal(t, node, foundNode)

	_, err = cluster.GetNodeByIP("192.168.1.2")
	assert.Error(t, err)
	assert.True(t, IsNodeNotFoundedError(err))
}

func TestGetNodeByFunc(t *testing.T) {
	cluster := NewCluster("test-cluster", "default")
	node := &ClusterNode{ID: "node1", IP: "192.168.1.1"}
	cluster.AddNode(node)

	findFunc := func(n *ClusterNode) bool {
		return n.IP == "192.168.1.1"
	}

	foundNode, err := cluster.GetNodeByFunc(findFunc)
	assert.NoError(t, err)
	assert.Equal(t, node, foundNode)

	findFunc = func(n *ClusterNode) bool {
		return n.IP == "192.168.1.2"
	}

	_, err = cluster.GetNodeByFunc(findFunc)
	assert.Error(t, err)
	assert.True(t, IsNodeNotFoundedError(err))
}
