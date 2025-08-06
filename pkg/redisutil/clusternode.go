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
	"fmt"
	"net"
	"sort"
	"strconv"
	"strings"
	"time"

	composev1alpha1 "github.com/upmio/compose-operator/api/v1alpha1"
	"github.com/upmio/compose-operator/pkg/utils"
)

// ClusterNode Represent a Redis ClusterNode
type ClusterNode struct {
	ID              string
	IP              string
	Port            string
	Role            string
	LinkState       string
	MasterReferent  string
	FailStatus      []string
	PingSent        int64
	PongRecv        int64
	ConfigEpoch     int64
	Slots           []Slot
	balance         int
	MigratingSlots  map[Slot]string
	ImportingSlots  map[Slot]string
	ServerStartTime time.Time

	NodeName  string
	ShardName string
}

// ClusterNodes represent a ClusterNode slice
type ClusterNodes []*ClusterNode

func (n ClusterNodes) String() string {
	stringer := []utils.Stringer{}
	for _, node := range n {
		stringer = append(stringer, node)
	}

	return utils.SliceJoin(stringer, ",")
}

// NewDefaultClusterNode builds and returns new default cluster node instance
func NewDefaultClusterNode() *ClusterNode {
	return &ClusterNode{
		Port:           DefaultRedisPort,
		Slots:          []Slot{},
		MigratingSlots: map[Slot]string{},
		ImportingSlots: map[Slot]string{},
	}
}

// NewClusterNode builds and returns new ClusterNode instance
func NewClusterNode(name, id, ip string) *ClusterNode {
	node := NewDefaultClusterNode()
	node.ID = id
	node.IP = ip
	node.NodeName = name

	return node
}

// SetRole from a flags string list set the ClusterNode's role
func (n *ClusterNode) SetRole(flags string) error {
	n.Role = "" // reset value before setting the new one
	vals := strings.Split(flags, ",")
	for _, val := range vals {
		switch val {
		case RedisMasterRole:
			n.Role = RedisMasterRole
		case RedisSlaveRole:
			n.Role = RedisSlaveRole
		}
	}

	if n.Role == "" {
		return errors.New("node setRole failed")
	}

	return nil
}

// GetRole return the Redis Cluster ClusterNode GetRole
func (n *ClusterNode) GetRole() composev1alpha1.RedisClusterNodeRole {
	switch n.Role {
	case RedisMasterRole:
		return composev1alpha1.RedisClusterNodeRoleSource
	case RedisSlaveRole:
		return composev1alpha1.RedisClusterNodeRoleReplica
	default:
		if n.MasterReferent != "" {
			return composev1alpha1.RedisClusterNodeRoleReplica
		}
		if len(n.Slots) > 0 {
			return composev1alpha1.RedisClusterNodeRoleSource
		}
	}

	return composev1alpha1.RedisClusterNodeRoleNone
}

// String representation of an Instance
func (n *ClusterNode) String() string {
	if n.ServerStartTime.IsZero() {
		return fmt.Sprintf("{Redis ID: %s, role: %s, master: %s, link: %s, status: %s, addr: %s, slots: %s, len(migratingSlots): %d, len(importingSlots): %d}", n.ID, n.GetRole(), n.MasterReferent, n.LinkState, n.FailStatus, n.IPPort(), SlotSlice(n.Slots), len(n.MigratingSlots), len(n.ImportingSlots))
	}
	return fmt.Sprintf("{Redis ID: %s, role: %s, master: %s, link: %s, status: %s, addr: %s, slots: %s, len(migratingSlots): %d, len(importingSlots): %d, ServerStartTime: %s}", n.ID, n.GetRole(), n.MasterReferent, n.LinkState, n.FailStatus, n.IPPort(), SlotSlice(n.Slots), len(n.MigratingSlots), len(n.ImportingSlots), n.ServerStartTime.Format("2006-01-02 15:04:05"))
}

// IPPort returns join Ip Port string
func (n *ClusterNode) IPPort() string {
	return net.JoinHostPort(n.IP, n.Port)
}

// GetNodesByFunc returns first node found by the FindNodeFunc
func (n ClusterNodes) GetNodesByFunc(f FindNodeFunc) (ClusterNodes, error) {
	nodes := ClusterNodes{}
	for _, node := range n {
		if f(node) {
			nodes = append(nodes, node)
		}
	}
	if len(nodes) == 0 {
		return nodes, nodeNotFoundedError
	}
	return nodes, nil
}

// ToAPINode used to convert the current ClusterNode to an API redisv1alpha1.RedisClusterNode
func (n *ClusterNode) ToAPINode() composev1alpha1.RedisClusterNode {
	port, _ := strconv.Atoi(n.Port)

	apiNode := composev1alpha1.RedisClusterNode{
		ID:    n.ID,
		Host:  n.IP,
		Role:  n.GetRole(),
		Port:  port,
		Slots: []string{},
	}

	return apiNode
}

// Clear used to clear possible ressources attach to the current ClusterNode
func (n *ClusterNode) Clear() {

}

func (n *ClusterNode) Balance() int {
	return n.balance
}

func (n *ClusterNode) SetBalance(balance int) {
	n.balance = balance
}

// SetLinkStatus set the ClusterNode link status
func (n *ClusterNode) SetLinkStatus(status string) error {
	n.LinkState = "" // reset value before setting the new one
	switch status {
	case RedisLinkStateConnected:
		n.LinkState = RedisLinkStateConnected
	case RedisLinkStateDisconnected:
		n.LinkState = RedisLinkStateDisconnected
	}

	if n.LinkState == "" {
		return errors.New("ClusterNode SetLinkStatus failed")
	}

	return nil
}

// SetFailureStatus set from inputs flags the possible failure status
func (n *ClusterNode) SetFailureStatus(flags string) {
	n.FailStatus = []string{} // reset value before setting the new one
	vals := strings.Split(flags, ",")
	for _, val := range vals {
		switch val {
		case NodeStatusFail:
			n.FailStatus = append(n.FailStatus, NodeStatusFail)
		case NodeStatusPFail:
			n.FailStatus = append(n.FailStatus, NodeStatusPFail)
		case NodeStatusHandshake:
			n.FailStatus = append(n.FailStatus, NodeStatusHandshake)
		case NodeStatusNoAddr:
			n.FailStatus = append(n.FailStatus, NodeStatusNoAddr)
		case NodeStatusNoFlags:
			n.FailStatus = append(n.FailStatus, NodeStatusNoFlags)
		}
	}
}

// SetReferentMaster set the redis node parent referent
func (n *ClusterNode) SetReferentMaster(ref string) {
	n.MasterReferent = ""
	if ref == "-" {
		return
	}
	n.MasterReferent = ref
}

// TotalSlots return the total number of slot
func (n *ClusterNode) TotalSlots() int {
	return len(n.Slots)
}

// HasStatus returns true if the node has the provided fail status flag
func (n *ClusterNode) HasStatus(flag string) bool {
	for _, status := range n.FailStatus {
		if status == flag {
			return true
		}
	}
	return false
}

// IsMasterWithNoSlot anonymous function for searching Master ClusterNode with no slot
var IsMasterWithNoSlot = func(n *ClusterNode) bool {
	if (n.GetRole() == composev1alpha1.RedisClusterNodeRoleSource) && (n.TotalSlots() == 0) {
		return true
	}
	return false
}

// IsMasterWithSlot anonymous function for searching Master ClusterNode withslot
var IsMasterWithSlot = func(n *ClusterNode) bool {
	if (n.GetRole() == composev1alpha1.RedisClusterNodeRoleSource) && (n.TotalSlots() > 0) {
		return true
	}
	return false
}

// IsSlave anonymous function for searching Slave ClusterNode
var IsSlave = func(n *ClusterNode) bool {
	return n.GetRole() == composev1alpha1.RedisClusterNodeRoleReplica
}

// SortNodes sort ReplicationNodes and return the sorted ReplicationNodes
func (n ClusterNodes) SortNodes() ClusterNodes {
	sort.Sort(n)
	return n
}

// GetNodeByID returns a Redis ClusterNode by its ID
// if not present in the ReplicationNodes slice return an error
func (n ClusterNodes) GetNodeByID(id string) (*ClusterNode, error) {
	for _, node := range n {
		if node.ID == id {
			return node, nil
		}
	}

	return nil, nodeNotFoundedError
}

// CountByFunc gives the number elements of NodeSlice that return true for the passed func.
func (n ClusterNodes) CountByFunc(fn func(*ClusterNode) bool) (result int) {
	for _, v := range n {
		if fn(v) {
			result++
		}
	}
	return
}

// FilterByFunc remove a node from a slice by node ID and returns the slice. If not found, fail silently. Value must be unique
func (n ClusterNodes) FilterByFunc(fn func(*ClusterNode) bool) ClusterNodes {
	newSlice := ClusterNodes{}
	for _, node := range n {
		if fn(node) {
			newSlice = append(newSlice, node)
		}
	}
	return newSlice
}

// SortByFunc returns a new ordered NodeSlice, determined by a func defining ‘less’.
func (n ClusterNodes) SortByFunc(less func(*ClusterNode, *ClusterNode) bool) ClusterNodes {
	//result := make(ReplicationNodes, len(n))
	//copy(result, n)
	by(less).Sort(n)
	return n
}

// Len is the number of elements in the collection.
func (n ClusterNodes) Len() int {
	return len(n)
}

// Less few reports whether the element with
// index i should sort before the element with index j.
func (n ClusterNodes) Less(i, j int) bool {
	return n[i].ID < n[j].ID
}

// Swap swaps the elements with indexes i and j.
func (n ClusterNodes) Swap(i, j int) {
	n[i], n[j] = n[j], n[i]
}

// By is the type of "less" function that defines the ordering of its ClusterNode arguments.
type by func(p1, p2 *ClusterNode) bool

// Sort is a method on the function type, By, that sorts the argument slice according to the function.
func (b by) Sort(nodes ClusterNodes) {
	ps := &nodeSorter{
		nodes: nodes,
		by:    b, // The Sort method's receiver is the function (closure) that defines the sort order.
	}
	sort.Sort(ps)
}

// nodeSorter joins a By function and a slice of ReplicationNodes to be sorted.
type nodeSorter struct {
	nodes ClusterNodes
	by    func(p1, p2 *ClusterNode) bool // Closure used in the Less method.
}

// Len is part of sort.Interface.
func (s *nodeSorter) Len() int {
	return len(s.nodes)
}

// Swap is part of sort.Interface.
func (s *nodeSorter) Swap(i, j int) {
	s.nodes[i], s.nodes[j] = s.nodes[j], s.nodes[i]
}

// Less is part of sort.Interface. It is implemented by calling the "by" closure in the sorter.
func (s *nodeSorter) Less(i, j int) bool {
	return s.by(s.nodes[i], s.nodes[j])
}

// LessByID compare 2 ReplicationNodes with their ID
func LessByID(n1, n2 *ClusterNode) bool {
	return n1.ID < n2.ID
}

// MoreByID compare 2 ReplicationNodes with their ID
func MoreByID(n1, n2 *ClusterNode) bool {
	return n1.ID > n2.ID
}
