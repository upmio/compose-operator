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
	"fmt"
	"net"
	"regexp"
	"strconv"

	"github.com/go-logr/logr"
)

const (
	// DefaultHashMaxSlots higher value of slot
	// as slots start at 0, total number of slots is defaultHashMaxSlots+1
	DefaultHashMaxSlots = 16383

	// ResetHard HARD mode for RESET command
	ResetHard = "HARD"
	// ResetSoft SOFT mode for RESET command
	ResetSoft = "SOFT"
)

const (
	clusterKnownNodesREString = "cluster_known_nodes:([0-9]+)"
)

var (
	clusterKnownNodesRE = regexp.MustCompile(clusterKnownNodesREString)
)

// IClusterAdmin redis cluster admin interface
type IClusterAdmin interface {
	// Connections returns the connection map of all clients
	Connections() IAdminConnections
	// Close the admin connections
	Close()
	// GetClusterInfos get node infos for all nodes
	GetClusterInfos() (*ClusterInfos, error)
	// ClusterManagerNodeIsEmpty Checks whether the node is empty. ClusterNode is considered not-empty if it has
	// some key or if it already knows other nodes
	ClusterManagerNodeIsEmpty() (bool, error)
	// SetConfigEpoch Assign a different config epoch to each node
	SetConfigEpoch() error
	// GetAllConfig get redis config by CONFIG GET *
	GetAllConfig(c IClient, addr string) (map[string]string, error)
	// AttachNodeToCluster command use to connect a ClusterNode to the cluster
	// the connection will be done on a random node part of the connection pool
	AttachNodeToCluster(addr string) error
	// AttachSlaveToMaster attach a slave to a master node
	AttachSlaveToMaster(slave *ClusterNode, masterID string) error
	// DetachSlave dettach a slave to its master
	DetachSlave(slave *ClusterNode) error
	// ForgetNode execute the Redis command to force the cluster to forget the ClusterNode
	ForgetNode(id string) error
	// SetSlots exec the redis command to set slots in a pipeline, provide
	// and empty nodeID if the set slots commands doesn't take a nodeID in parameter
	SetSlots(addr string, action string, slots []Slot, nodeID string) error
	// AddSlots exec the redis command to add slots in a pipeline
	AddSlots(addr string, slots []Slot) error
	// SetSlot use to set SETSLOT command on a slot
	SetSlot(addr, action string, slot Slot, nodeID string) error
	// MigrateKeys from addr to destination node. returns number of slot migrated. If replace is true, replace key on busy error
	MigrateKeys(addr string, dest *ClusterNode, slots []Slot, batch, timeout int, replace bool) (int, error)
	// MigrateKeysInSlot use to migrate keys from slot to other slot. if replace is true, replace key on busy error
	// timeout is in milliseconds
	MigrateKeysInSlot(addr string, dest *ClusterNode, slot Slot, batch int, timeout int, replace bool) (int, error)
	// GetHashMaxSlot get the max slot value
	GetHashMaxSlot() Slot
}

// ClusterAdmin wraps redis cluster admin logic
type ClusterAdmin struct {
	hashMaxSlots Slot
	cnx          IAdminConnections
	log          logr.Logger
}

// NewClusterAdmin returns new AdminInterface instance
// at the same time it connects to all Redis ReplicationNodes thanks to the addrs list
func NewClusterAdmin(addrs []string, options *AdminOptions, log logr.Logger) IClusterAdmin {
	a := &ClusterAdmin{
		hashMaxSlots: DefaultHashMaxSlots,
		log:          log.WithName("redis_util"),
	}

	// perform initial connections
	a.cnx = NewAdminConnections(addrs, options, log)

	return a
}

// Connections returns the connection map of all clients
func (a *ClusterAdmin) Connections() IAdminConnections {
	return a.cnx
}

// Close used to close all possible resources instanciate by the ClusterAdmin
func (a *ClusterAdmin) Close() {
	a.Connections().Reset()
}

// GetClusterInfos return the ReplicationNodes infos for all nodes
func (a *ClusterAdmin) GetClusterInfos() (*ClusterInfos, error) {
	infos := NewClusterInfos()
	clusterErr := NewClusterInfosError()

	for addr, c := range a.Connections().GetAll() {
		nodeinfos, err := a.getInfos(c, addr)
		if err != nil {
			a.log.WithValues("err", err).Info("get redis info failed")
			infos.Status = ClusterInfosPartial
			clusterErr.partial = true
			clusterErr.errs[addr] = err
			continue
		}
		if nodeinfos.Node != nil && nodeinfos.Node.IPPort() == addr {
			infos.Infos[addr] = nodeinfos
		} else {
			a.log.Info("bad node info retrieved from", "addr", addr)
		}
	}

	if len(clusterErr.errs) == 0 {
		clusterErr.inconsistent = !infos.ComputeStatus(a.log)
	}
	if infos.Status == ClusterInfosConsistent {
		return infos, nil
	}
	return infos, clusterErr
}

func (a *ClusterAdmin) getInfos(c IClient, addr string) (*NodeInfos, error) {
	resp := c.Cmd("CLUSTER", "NODES")
	if err := a.Connections().ValidateResp(resp, addr, "unable to retrieve node info"); err != nil {
		return nil, err
	}

	var raw string
	var err error
	raw, err = resp.Str()

	if err != nil {
		return nil, fmt.Errorf("wrong format from CLUSTER NODES: %v", err)
	}

	nodeInfos := DecodeNodeInfos(&raw, addr, a.log)

	return nodeInfos, nil
}

// ClusterManagerNodeIsEmpty Checks whether the node is empty. ClusterNode is considered not-empty if it has
// some key or if it already knows other nodes
func (a *ClusterAdmin) ClusterManagerNodeIsEmpty() (bool, error) {
	for addr, c := range a.Connections().GetAll() {
		knowNodes, err := a.clusterKnowNodes(c, addr)
		if err != nil {
			return false, err
		}
		if knowNodes != 1 {
			return false, nil
		}
	}
	return true, nil
}

func (a *ClusterAdmin) clusterKnowNodes(c IClient, addr string) (int, error) {
	resp := c.Cmd("CLUSTER", "INFO")
	if err := a.Connections().ValidateResp(resp, addr, "unable to retrieve cluster info"); err != nil {
		return 0, err
	}

	var raw string
	var err error
	raw, err = resp.Str()

	if err != nil {
		return 0, fmt.Errorf("wrong format from CLUSTER INFO: %v", err)
	}

	match := clusterKnownNodesRE.FindStringSubmatch(raw)
	if len(match) == 0 {
		return 0, fmt.Errorf("cluster_known_nodes regex not found")
	}
	return strconv.Atoi(match[1])
}

// AttachSlaveToMaster attach a slave to a master node
func (a *ClusterAdmin) AttachSlaveToMaster(slave *ClusterNode, masterID string) error {
	c, err := a.Connections().Get(slave.IPPort())
	if err != nil {
		return err
	}

	resp := c.Cmd("CLUSTER", "REPLICATE", masterID)
	if err := a.Connections().ValidateResp(resp, slave.IPPort(), "unable to run command REPLICATE"); err != nil {
		return err
	}

	slave.SetReferentMaster(masterID)
	slave.SetRole(RedisSlaveRole)

	return nil
}

// AddSlots use to ADDSLOT commands on several slots
func (a *ClusterAdmin) AddSlots(addr string, slots []Slot) error {
	if len(slots) == 0 {
		return nil
	}
	c, err := a.Connections().Get(addr)
	if err != nil {
		return err
	}

	resp := c.Cmd("CLUSTER", "ADDSLOTS", slots)

	return a.Connections().ValidateResp(resp, addr, "unable to run CLUSTER ADDSLOTS")
}

// SetSlots use to set SETSLOT command on several slots
func (a *ClusterAdmin) SetSlots(addr, action string, slots []Slot, nodeID string) error {
	if len(slots) == 0 {
		return nil
	}
	c, err := a.Connections().Get(addr)
	if err != nil {
		return err
	}
	for _, slot := range slots {
		if nodeID == "" {
			c.PipeAppend("CLUSTER", "SETSLOT", slot, action)
		} else {
			c.PipeAppend("CLUSTER", "SETSLOT", slot, action, nodeID)
		}
	}
	if !a.Connections().ValidatePipeResp(c, addr, "Cannot SETSLOT") {
		return fmt.Errorf("Error occured during CLUSTER SETSLOT %s", action)
	}
	c.PipeClear()

	return nil
}

// SetSlot use to set SETSLOT command on a slot
func (a *ClusterAdmin) SetSlot(addr, action string, slot Slot, nodeID string) error {
	c, err := a.Connections().Get(addr)
	if err != nil {
		return err
	}
	if nodeID == "" {
		c.PipeAppend("CLUSTER", "SETSLOT", slot, action)
	} else {
		c.PipeAppend("CLUSTER", "SETSLOT", slot, action, nodeID)
	}
	if !a.Connections().ValidatePipeResp(c, addr, "Cannot SETSLOT") {
		return fmt.Errorf("Error occured during CLUSTER SETSLOT %s", action)
	}
	c.PipeClear()

	return nil
}

func (a *ClusterAdmin) SetConfigEpoch() error {
	configEpoch := 1
	for addr, c := range a.Connections().GetAll() {
		resp := c.Cmd("CLUSTER", "SET-CONFIG-EPOCH", configEpoch)
		if err := a.Connections().ValidateResp(resp, addr, "unable to run command SET-CONFIG-EPOCH"); err != nil {
			return err
		}
		configEpoch++
	}
	return nil
}

// AttachNodeToCluster command use to connect a ClusterNode to the cluster
func (a *ClusterAdmin) AttachNodeToCluster(addr string) error {
	ip, port, err := net.SplitHostPort(addr)
	if err != nil {
		return err
	}

	all := a.Connections().GetAll()
	if len(all) == 0 {
		return fmt.Errorf("no connection for other redis-node found")
	}
	for cAddr, c := range a.Connections().GetAll() {
		if cAddr == addr {
			continue
		}
		a.log.V(3).Info("CLUSTER MEET", "from addr", cAddr, "to", addr)
		resp := c.Cmd("CLUSTER", "MEET", ip, port)
		if err = a.Connections().ValidateResp(resp, addr, "cannot attach node to cluster"); err != nil {
			return err
		}
	}

	a.Connections().Add(addr)

	a.log.Info(fmt.Sprintf("node %s attached properly", addr))
	return nil
}

// GetAllConfig get redis config by CONFIG GET *
func (a *ClusterAdmin) GetAllConfig(c IClient, addr string) (map[string]string, error) {
	resp := c.Cmd("CONFIG", "GET", "*")
	if err := a.Connections().ValidateResp(resp, addr, "unable to retrieve config"); err != nil {
		return nil, err
	}

	var raw map[string]string
	var err error
	raw, err = resp.Map()

	if err != nil {
		return nil, fmt.Errorf("wrong format from CONFIG GET *: %v", err)
	}

	return raw, nil
}

// GetHashMaxSlot get the max slot value
func (a *ClusterAdmin) GetHashMaxSlot() Slot {
	return a.hashMaxSlots
}

// MigrateKeys use to migrate keys from slots to other slots. if replace is true, replace key on busy error
// timeout is in milliseconds
func (a *ClusterAdmin) MigrateKeys(addr string, dest *ClusterNode, slots []Slot, batch int, timeout int, replace bool) (int, error) {
	if len(slots) == 0 {
		return 0, nil
	}
	keyCount := 0
	c, err := a.Connections().Get(addr)
	if err != nil {
		return keyCount, err
	}
	timeoutStr := strconv.Itoa(timeout)
	batchStr := strconv.Itoa(batch)

	for _, slot := range slots {
		for {
			resp := c.Cmd("CLUSTER", "GETKEYSINSLOT", slot, batchStr)
			if err := a.Connections().ValidateResp(resp, addr, "Unable to run command GETKEYSINSLOT"); err != nil {
				return keyCount, err
			}
			keys, err := resp.List()
			if err != nil {
				a.log.Error(err, "wrong returned format for CLUSTER GETKEYSINSLOT")
				return keyCount, err
			}

			keyCount += len(keys)
			if len(keys) == 0 {
				break
			}

			args := a.migrateCmdArgs(dest, timeoutStr, replace, keys)

			resp = c.Cmd("MIGRATE", args)
			if err := a.Connections().ValidateResp(resp, addr, "Unable to run command MIGRATE"); err != nil {
				return keyCount, err
			}
		}
	}

	return keyCount, nil
}

// MigrateKeysInSlot use to migrate keys from slot to other slot. if replace is true, replace key on busy error
// timeout is in milliseconds
func (a *ClusterAdmin) MigrateKeysInSlot(addr string, dest *ClusterNode, slot Slot, batch int, timeout int, replace bool) (int, error) {
	keyCount := 0
	c, err := a.Connections().Get(addr)
	if err != nil {
		return keyCount, err
	}
	timeoutStr := strconv.Itoa(timeout)
	batchStr := strconv.Itoa(batch)

	for {
		resp := c.Cmd("CLUSTER", "GETKEYSINSLOT", slot, batchStr)
		if err := a.Connections().ValidateResp(resp, addr, "Unable to run command GETKEYSINSLOT"); err != nil {
			return keyCount, err
		}
		keys, err := resp.List()
		if err != nil {
			a.log.Error(err, "wrong returned format for CLUSTER GETKEYSINSLOT")
			return keyCount, err
		}

		keyCount += len(keys)
		if len(keys) == 0 {
			break
		}

		args := a.migrateCmdArgs(dest, timeoutStr, replace, keys)

		resp = c.Cmd("MIGRATE", args)
		if err := a.Connections().ValidateResp(resp, addr, "Unable to run command MIGRATE"); err != nil {
			return keyCount, err
		}
	}

	return keyCount, nil
}

func (a *ClusterAdmin) migrateCmdArgs(dest *ClusterNode, timeoutStr string, replace bool, keys []string) []string {
	args := []string{dest.IP, dest.Port, "", "0", timeoutStr}
	if password, ok := a.Connections().GetAUTH(); ok {
		args = append(args, "AUTH", password)
	}
	if replace {
		args = append(args, "REPLACE", "KEYS")
	} else {
		args = append(args, "KEYS")
	}
	args = append(args, keys...)
	return args
}

// ForgetNode used to force other redis cluster node to forget a specific node
func (a *ClusterAdmin) ForgetNode(id string) error {
	infos, _ := a.GetClusterInfos()
	for nodeAddr, nodeinfos := range infos.Infos {
		if nodeinfos.Node.ID == id {
			continue
		}

		if IsSlave(nodeinfos.Node) && nodeinfos.Node.MasterReferent == id {
			if err := a.DetachSlave(nodeinfos.Node); err != nil {
				a.log.Error(err, "DetachSlave", "node", nodeAddr)
			}
			a.log.Info(fmt.Sprintf("detach slave id: %s of master: %s", nodeinfos.Node.ID, id))
		}

		c, err := a.Connections().Get(nodeAddr)
		if err != nil {
			a.log.Error(err, fmt.Sprintf("failed to get connection to node %s for forgetting node %s", nodeAddr, id))
			continue
		}

		a.log.Info("CLUSTER FORGET", "id", id, "from", nodeAddr)
		resp := c.Cmd("CLUSTER", "FORGET", id)
		a.Connections().ValidateResp(resp, nodeAddr, "Unable to execute FORGET command")
	}

	a.log.Info("Forget node done", "node", id)
	return nil
}

// DetachSlave use to detach a slave to a master
func (a *ClusterAdmin) DetachSlave(slave *ClusterNode) error {
	c, err := a.Connections().Get(slave.IPPort())
	if err != nil {
		a.log.Error(err, fmt.Sprintf("unable to get the connection for slave ID:%s, addr:%s", slave.ID, slave.IPPort()))
		return err
	}

	resp := c.Cmd("CLUSTER", "RESET", "SOFT")
	if err = a.Connections().ValidateResp(resp, slave.IPPort(), "cannot attach node to cluster"); err != nil {
		return err
	}

	if err = a.AttachNodeToCluster(slave.IPPort()); err != nil {
		a.log.Error(err, fmt.Sprintf("[DetachSlave] unable to AttachNodeToCluster the Slave id: %s addr:%s", slave.ID, slave.IPPort()))
		return err
	}

	slave.SetReferentMaster("")
	slave.SetRole(RedisMasterRole)

	return nil
}

// FlushAndReset flush the cluster and reset the cluster configuration of the node. Commands are piped, to ensure no items arrived between flush and reset
func (a *ClusterAdmin) FlushAndReset(addr string, mode string) error {
	c, err := a.Connections().Get(addr)
	if err != nil {
		return err
	}
	c.PipeAppend("FLUSHALL")
	c.PipeAppend("CLUSTER", "RESET", mode)

	if !a.Connections().ValidatePipeResp(c, addr, "Cannot reset node") {
		return fmt.Errorf("cannot reset node %s", addr)
	}

	return nil
}
