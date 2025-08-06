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
	"reflect"
	"testing"
	"time"
)

func TestNodes_SortByFunc(t *testing.T) {
	n1 := ClusterNode{
		ID:      "n1",
		IP:      "10.1.1.1",
		Port:    "",
		Role:    "master",
		balance: 1365,
	}
	n2 := ClusterNode{
		ID:      "n2",
		IP:      "10.1.1.2",
		Port:    "",
		Role:    "master",
		balance: 1366,
	}
	n3 := ClusterNode{
		ID:      "n3",
		IP:      "10.1.1.3",
		Port:    "",
		Role:    "master",
		balance: 1365,
	}
	n4 := ClusterNode{
		ID:      "n4",
		IP:      "10.1.1.4",
		Port:    "",
		Role:    "master",
		balance: -4096,
	}
	type args struct {
		less func(*ClusterNode, *ClusterNode) bool
	}
	tests := []struct {
		name string
		n    ClusterNodes
		args args
		want ClusterNodes
	}{
		{
			name: "asc by balance",
			n:    ClusterNodes{&n1, &n2, &n3, &n4},
			args: args{less: func(a, b *ClusterNode) bool { return a.Balance() < b.Balance() }},
			want: ClusterNodes{&n4, &n1, &n3, &n2},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.n.SortByFunc(tt.args.less); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("SortByFunc() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestNewDefaultNode(t *testing.T) {
	node := NewDefaultClusterNode()

	if node.Port != DefaultRedisPort {
		t.Errorf("expected %s, got %s", DefaultRedisPort, node.Port)
	}

	if len(node.Slots) != 0 {
		t.Errorf("expected 0 slots, got %d", len(node.Slots))
	}
}

func TestNewNode(t *testing.T) {
	name := "node1"
	id := "12345"
	ip := "127.0.0.1"

	node := NewClusterNode(name, id, ip)

	if node.ID != id {
		t.Errorf("expected %s, got %s", id, node.ID)
	}

	if node.IP != ip {
		t.Errorf("expected %s, got %s", ip, node.IP)
	}

	if node.NodeName != name {
		t.Errorf("expected %s, got %s", name, node.NodeName)
	}
}

func TestSetRole(t *testing.T) {
	node := NewDefaultClusterNode()

	err := node.SetRole(RedisMasterRole)
	if err != nil {
		t.Errorf("expected nil, got %v", err)
	}
	if node.Role != RedisMasterRole {
		t.Errorf("expected %s, got %s", RedisMasterRole, node.Role)
	}

	err = node.SetRole("unknown")
	if err == nil {
		t.Errorf("expected error, got nil")
	}
}

func TestSetLinkStatus(t *testing.T) {
	node := NewDefaultClusterNode()

	err := node.SetLinkStatus(RedisLinkStateConnected)
	if err != nil {
		t.Errorf("expected nil, got %v", err)
	}
	if node.LinkState != RedisLinkStateConnected {
		t.Errorf("expected %s, got %s", RedisLinkStateConnected, node.LinkState)
	}

	err = node.SetLinkStatus("unknown")
	if err == nil {
		t.Errorf("expected error, got nil")
	}
}

func TestSetFailureStatus(t *testing.T) {
	node := NewDefaultClusterNode()

	node.SetFailureStatus(NodeStatusFail)
	if len(node.FailStatus) != 1 {
		t.Errorf("expected 1, got %d", len(node.FailStatus))
	}
	if node.FailStatus[0] != NodeStatusFail {
		t.Errorf("expected %s, got %s", NodeStatusFail, node.FailStatus[0])
	}
}

func TestSetReferentMaster(t *testing.T) {
	node := NewDefaultClusterNode()

	node.SetReferentMaster("master1")
	if node.MasterReferent != "master1" {
		t.Errorf("expected master1, got %s", node.MasterReferent)
	}

	node.SetReferentMaster("-")
	if node.MasterReferent != "" {
		t.Errorf("expected empty string, got %s", node.MasterReferent)
	}
}

func TestNode_String(t *testing.T) {
	node := NewDefaultClusterNode()
	node.ID = "12345"
	node.Role = RedisMasterRole
	node.MasterReferent = "master1"
	node.LinkState = RedisLinkStateConnected
	node.FailStatus = []string{NodeStatusFail}
	node.IP = "127.0.0.1"
	node.Port = "6379"
	node.Slots = []Slot{}
	node.MigratingSlots = map[Slot]string{}
	node.ImportingSlots = map[Slot]string{}
	node.ServerStartTime = time.Time{}

	expected := "{Redis ID: 12345, role: Source, master: master1, link: connected, status: [fail], addr: 127.0.0.1:6379, slots: [], len(migratingSlots): 0, len(importingSlots): 0}"
	if node.String() != expected {
		t.Errorf("expected %s, got %s", expected, node.String())
	}
}

func TestNode_IPPort(t *testing.T) {
	node := NewDefaultClusterNode()
	node.IP = "127.0.0.1"
	node.Port = "6379"

	expected := "127.0.0.1:6379"
	if node.IPPort() != expected {
		t.Errorf("expected %s, got %s", expected, node.IPPort())
	}
}

func TestNodes_SortNodes(t *testing.T) {
	node1 := NewClusterNode("node1", "1", "127.0.0.1")
	node2 := NewClusterNode("node2", "2", "127.0.0.2")

	nodes := ClusterNodes{node2, node1}
	sortedNodes := nodes.SortNodes()

	if sortedNodes[0].ID != "1" {
		t.Errorf("expected %s, got %s", "1", sortedNodes[0].ID)
	}
	if sortedNodes[1].ID != "2" {
		t.Errorf("expected %s, got %s", "2", sortedNodes[1].ID)
	}
}

func TestNodes_GetNodeByID(t *testing.T) {
	node1 := NewClusterNode("node1", "1", "127.0.0.1")
	node2 := NewClusterNode("node2", "2", "127.0.0.2")

	nodes := ClusterNodes{node1, node2}

	node, err := nodes.GetNodeByID("1")
	if err != nil {
		t.Errorf("expected nil, got %v", err)
	}
	if node.ID != "1" {
		t.Errorf("expected %s, got %s", "1", node.ID)
	}

	node, err = nodes.GetNodeByID("3")
	if err == nil {
		t.Errorf("expected error, got nil")
	}
}
