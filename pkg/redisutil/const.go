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

const (
	// DefaultRedisPort define the default Redis Port
	DefaultRedisPort = "6379"
	// RedisMasterRole redis role master
	RedisMasterRole = "master"
	// RedisSlaveRole redis role slave
	RedisSlaveRole = "slave"

	// RedisSourceRole redis role source
	RedisSourceRole = "Source"
	// RedisReplicaRole redis role replica
	RedisReplicaRole = "Replica"

	// RedisLinkStateConnected redis connection status connected
	RedisLinkStateConnected = "connected"
	// RedisLinkStateDisconnected redis connection status disconnected
	RedisLinkStateDisconnected = "disconnected"

	// NodeStatusPFail ClusterNode is in PFAIL state. Not reachable for the node you are contacting, but still logically reachable
	NodeStatusPFail = "fail?"
	// NodeStatusFail ClusterNode is in FAIL state. It was not reachable for multiple nodes that promoted the PFAIL state to FAIL
	NodeStatusFail = "fail"
	// NodeStatusHandshake Untrusted node, we are handshaking.
	NodeStatusHandshake = "handshake"
	// NodeStatusNoAddr No address known for this node
	NodeStatusNoAddr = "noaddr"
	// NodeStatusNoFlags no flags at all
	NodeStatusNoFlags = "noflags"
)
