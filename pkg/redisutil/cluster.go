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

// ClusterStatus Redis Cluster status
type ClusterStatus string

const (
	// ClusterStatusOK ClusterStatus OK
	ClusterStatusOK ClusterStatus = "Healthy"
	// ClusterStatusKO ClusterStatus KO
	ClusterStatusKO ClusterStatus = "Failed"
)

// Cluster represents a Redis Cluster
type Cluster struct {
	Name        string
	Namespace   string
	Nodes       map[string]*ClusterNode
	Status      ClusterStatus
	ActionsInfo ClusterActionsInfo
}

// ClusterActionsInfo use to store information about current action on the Cluster
type ClusterActionsInfo struct {
	NbslotsToMigrate int32
}

// NewCluster builds and returns new Cluster instance
func NewCluster(name, namespace string) *Cluster {
	c := &Cluster{
		Name:      name,
		Namespace: namespace,
		Nodes:     make(map[string]*ClusterNode),
	}

	return c
}

// AddNode used to add new ClusterNode in the cluster
// if node with the same ID is already present in the cluster
// the previous ClusterNode is replaced
func (c *Cluster) AddNode(node *ClusterNode) {
	if n, ok := c.Nodes[node.ID]; ok {
		n.Clear()
	}

	c.Nodes[node.ID] = node
}

// GetNodeByID returns a Cluster ClusterNode by its ID
// if not present in the cluster return an error
func (c *Cluster) GetNodeByID(id string) (*ClusterNode, error) {
	if n, ok := c.Nodes[id]; ok {
		return n, nil
	}
	return nil, nodeNotFoundedError
}

// GetNodeByIP returns a Cluster ClusterNode by its ID
// if not present in the cluster return an error
func (c *Cluster) GetNodeByIP(ip string) (*ClusterNode, error) {
	findFunc := func(node *ClusterNode) bool {
		return node.IP == ip
	}

	return c.GetNodeByFunc(findFunc)
}

// GetNodeByFunc returns first node found by the FindNodeFunc
func (c *Cluster) GetNodeByFunc(f FindNodeFunc) (*ClusterNode, error) {
	for _, n := range c.Nodes {
		if f(n) {
			return n, nil
		}
	}
	return nil, nodeNotFoundedError
}

// FindNodeFunc function for finding a ClusterNode
// it is use as input for GetNodeByFunc and GetNodesByFunc
type FindNodeFunc func(node *ClusterNode) bool
