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

package mysqlutil

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"github.com/go-logr/logr"
	"strconv"
)

// IReplicationAdmin mysql admin interface
type IReplicationAdmin interface {
	// Connections returns the connection map of all clients
	Connections() IAdminConnections

	// Close the admin connections
	Close()

	// GetReplicationStatus return the ReplicationNodes infos for all nodes
	GetReplicationStatus(ctx context.Context) *ReplicationInfo

	// SetSemiSyncSourceON set mysql variables semi_sync_source_enabled to ON
	SetSemiSyncSourceON(ctx context.Context, addr string) error

	// SetSemiSyncSourceOFF set mysql variables semi_sync_source_enabled to OFF
	SetSemiSyncSourceOFF(ctx context.Context, addr string) error

	// SetSemiSyncReplicaON set mysql variables semi_sync_replica_enabled to ON
	SetSemiSyncReplicaON(ctx context.Context, addr string) error

	// SetSemiSyncReplicaOFF set mysql variables semi_sync_replica_enabled to OFF
	SetSemiSyncReplicaOFF(ctx context.Context, addr string) error

	// SetAllNodeReadOnly set all mysql node read_only super_read_only to true
	SetAllNodeReadOnly(ctx context.Context) error

	// SetReadOnly set mysql variables read_only to flag
	SetReadOnly(ctx context.Context, addr string, flag bool) error

	// SetSuperReadOnly set mysql variables super_read_only to flag
	SetSuperReadOnly(ctx context.Context, addr string, flag bool) error

	// ResetReplica set mysql reset replica all
	ResetReplica(ctx context.Context, addr string) error

	// WaitSourcePos wait mysql source position
	WaitSourcePos(ctx context.Context, addr, SourceLogFile, ReadSourceLogPos string) error

	// SetReplica set mysql replica source and start replica
	SetReplica(ctx context.Context, addr, sourceHost, sourcePort, replicationUser, replicationPwd string) error

	// StopReplica stop mysql replica process
	StopReplica(ctx context.Context, addr string) error

	// StartReplica start mysql replica process
	StartReplica(ctx context.Context, addr string) error

	// GetUser get mysql users list
	GetUser(ctx context.Context, addr, host string, filter []string) (Users, error)

	// GetVersion return mysql version
	GetVersion(ctx context.Context, addr string) (string, error)
}

// ReplicationAdmin wraps redis cluster admin logic
type ReplicationAdmin struct {
	cnx IAdminConnections
	log logr.Logger
}

// NewReplicationAdmin returns new AdminInterface instance
// at the same time it connects to all Redis ReplicationNodes thanks to the addrs list
func NewReplicationAdmin(addrs []string, options *AdminOptions, log logr.Logger) IReplicationAdmin {
	a := &ReplicationAdmin{
		log: log.WithName("mysql_util"),
	}

	// perform initial connections
	a.cnx = NewAdminConnections(addrs, options, log)

	return a
}

// Connections returns the connection map of all clients
func (a *ReplicationAdmin) Connections() IAdminConnections {
	return a.cnx
}

// Close used to close all possible resources instanciate by the ReplicationAdmin
func (a *ReplicationAdmin) Close() {
	a.Connections().Reset()
}

// GetReplicationStatus return the ReplicationNodes infos for all nodes
func (a *ReplicationAdmin) GetReplicationStatus(ctx context.Context) *ReplicationInfo {
	infos := NewReplicationInfo()

	for addr := range a.Connections().GetAll() {

		node, err := a.getInfos(ctx, addr)
		if err != nil {
			a.log.Error(err, "failed to gather mysql replication information")
			continue
		}
		if node != nil && node.HostPort() == addr {
			infos.Nodes[addr] = node
		} else {
			a.log.Info("bad node info retrieved from", "addr", addr)
		}
	}

	return infos
}

func (a *ReplicationAdmin) getInfos(ctx context.Context, addr string) (*ReplicationNode, error) {
	c, err := a.cnx.Get(addr)
	if err != nil {
		return nil, err
	}

	rows, err := c.Query(ctx, showReplicaStatusSQL)
	if err != nil {
		return nil, fmt.Errorf("failed to execute 'SHOW REPLICA STATUS' on [%s]: %v", addr, err)
	}

	defer rows.Close()

	result, err := ScanMap(rows)
	if err != nil {
		return nil, fmt.Errorf("failed to scan replica status map on [%s]: %v", addr, err)
	}

	node := DecodeNode(result, addr, a.log)

	var flag string

	if err := c.QueryRow(ctx, &flag, showReadOnlySQL); err != nil {
		return nil, fmt.Errorf("failed to get 'read_only' variables on [%s]: %v", addr, err)
	} else {
		switch flag {
		case "1":
			node.ReadOnly = true
		case "0":
			node.ReadOnly = false
		}
	}

	if err := c.QueryRow(ctx, &flag, showSuperReadOnlySQL); err != nil {
		return nil, fmt.Errorf("failed to get 'super_read_only' variables on [%s]: %v", addr, err)
	} else {
		switch flag {
		case "1":
			node.SuperReadOnly = true
		case "0":
			node.SuperReadOnly = false
		}
	}

	if err := c.QueryRow(ctx, &flag, showSemiSyncSourceEnabled); err != nil {
		return nil, fmt.Errorf("failed to get 'rpl_semi_sync_source_enabled' variables on [%s]: %v", addr, err)
	} else {
		switch flag {
		case "1":
			node.SemiSyncSourceEnabled = true
		case "0":
			node.SemiSyncSourceEnabled = false
		}
	}

	if err := c.QueryRow(ctx, &flag, showSemiSyncReplicaEnabled); err != nil {
		return nil, fmt.Errorf("failed to get 'rpl_semi_sync_replica_enabled' variables on [%s]: %v", addr, err)
	} else {
		switch flag {
		case "1":
			node.SemiSyncReplicaEnabled = true
		case "0":
			node.SemiSyncReplicaEnabled = false
		}
	}

	return node, nil
}

func (a *ReplicationAdmin) SetSemiSyncSourceON(ctx context.Context, addr string) error {
	c, err := a.cnx.Get(addr)
	if err != nil {
		return err
	}

	if _, err := c.Exec(ctx, setSemiSyncSourceEnabled, 1); err != nil {
		return fmt.Errorf("failed to set 'rpl_semi_sync_source_enabled=ON' on [%s]: %v", addr, err)
	}

	if _, err := c.Exec(ctx, setSemiSyncTimeout); err != nil {
		return fmt.Errorf("failed to set 'rpl_semi_sync_source_timeout=10000000000000' on [%s]: %v", addr, err)
	}

	if _, err := c.Exec(ctx, setSemiSyncWaitPoint); err != nil {
		return fmt.Errorf("failed to set 'rpl_semi_sync_source_wait_point=AFTER_SYNC' on [%s]: %v", addr, err)
	}

	return nil
}

func (a *ReplicationAdmin) SetSemiSyncSourceOFF(ctx context.Context, addr string) error {
	c, err := a.cnx.Get(addr)
	if err != nil {
		return err
	}

	if _, err := c.Exec(ctx, setSemiSyncSourceEnabled, 0); err != nil {
		return fmt.Errorf("failed to set 'rpl_semi_sync_source_enabled=OFF' on [%s]: %v", addr, err)
	}

	return nil
}

func (a *ReplicationAdmin) SetSemiSyncReplicaON(ctx context.Context, addr string) error {
	c, err := a.cnx.Get(addr)
	if err != nil {
		return err
	}

	if _, err := c.Exec(ctx, setSemiSyncReplicaEnabled, 1); err != nil {
		return fmt.Errorf("failed to set 'rpl_semi_sync_replica_enabled=ON' on [%s]: %v", addr, err)
	}

	if _, err := c.Exec(ctx, setSemiSyncTimeout); err != nil {
		return fmt.Errorf("failed to set 'rpl_semi_sync_source_timeout=10000000000000' on [%s]: %v", addr, err)
	}

	if _, err := c.Exec(ctx, setSemiSyncWaitPoint); err != nil {
		return fmt.Errorf("failed to set 'rpl_semi_sync_source_wait_point=AFTER_SYNC' on [%s]: %v", addr, err)
	}

	return nil
}

func (a *ReplicationAdmin) SetSemiSyncReplicaOFF(ctx context.Context, addr string) error {
	c, err := a.cnx.Get(addr)
	if err != nil {
		return err
	}

	if _, err := c.Exec(ctx, setSemiSyncReplicaEnabled, 0); err != nil {
		return fmt.Errorf("failed to set 'rpl_semi_sync_replica_enabled=OFF' on [%s]: %v", addr, err)
	}

	return nil
}

func (a *ReplicationAdmin) SetReadOnly(ctx context.Context, addr string, flag bool) error {
	c, err := a.cnx.Get(addr)
	if err != nil {
		return err
	}

	if flag {
		if _, err := c.Exec(ctx, setReadOnlySQL, "ON"); err != nil {
			return fmt.Errorf("failed to set 'READ_ONLY=ON' on [%s]: %v", addr, err)

		}
	} else {
		if _, err := c.Exec(ctx, setReadOnlySQL, "OFF"); err != nil {
			return fmt.Errorf("failed to set 'READ_ONLY=OFF' on [%s]: %v", addr, err)
		}
	}

	return nil
}

func (a *ReplicationAdmin) SetSuperReadOnly(ctx context.Context, addr string, flag bool) error {
	c, err := a.cnx.Get(addr)
	if err != nil {
		return err
	}

	if flag {
		if _, err := c.Exec(ctx, setSuperReadOnlySQL, "ON"); err != nil {
			return fmt.Errorf("failed to set 'SUPER_READ_ONLY=ON' on [%s]: %v", addr, err)
		}
	} else {
		if _, err := c.Exec(ctx, setSuperReadOnlySQL, "OFF"); err != nil {
			return fmt.Errorf("failed to set 'SUPER_READ_ONLY=OFF' on [%s]: %v", addr, err)
		}
	}

	return nil
}

func (a *ReplicationAdmin) ResetReplica(ctx context.Context, addr string) error {
	c, err := a.cnx.Get(addr)
	if err != nil {
		return err
	}

	if _, err := c.Exec(ctx, stopReplicaSQL); err != nil {
		return fmt.Errorf("failed to execute 'STOP REPLICA' on [%s]: %v", addr, err)
	}

	if _, err := c.Exec(ctx, resetReplicaSQL); err != nil {
		return fmt.Errorf("failed to execute 'RESET REPLICA ALL' on [%s]: %v", addr, err)
	}

	return nil
}

func (a *ReplicationAdmin) SetAllNodeReadOnly(ctx context.Context) error {
	var errs []error

	for addr, _ := range a.Connections().GetAll() {
		if err := a.SetReadOnly(ctx, addr, true); err != nil {
			errs = append(errs, err)
		}

		if err := a.SetSuperReadOnly(ctx, addr, true); err != nil {
			errs = append(errs, err)
		}
	}

	return errors.Join(errs...)
}

func (a *ReplicationAdmin) WaitSourcePos(ctx context.Context, addr, SourceLogFile, ReadSourceLogPos string) error {
	c, err := a.cnx.Get(addr)
	if err != nil {
		return err
	}

	if _, err := c.Exec(ctx, waitSourcePosSql, SourceLogFile, ReadSourceLogPos); err != nil {
		return fmt.Errorf("failed to execute 'select SOURCE_POS_WAIT(?,?,120)' on [%s]: %v", addr, err)
	}

	return nil
}

func (a *ReplicationAdmin) SetReplica(ctx context.Context, addr, sourceHost, sourcePort, replicationUser, replicationPwd string) error {
	c, err := a.cnx.Get(addr)
	if err != nil {
		return err
	}

	port, err := strconv.Atoi(sourcePort)
	if err != nil {
		return err
	}
	if _, err := c.Exec(ctx, setReplicaSQL, sourceHost, port, replicationUser, replicationPwd); err != nil {
		return fmt.Errorf("failed to execute 'CHANGE REPLICATION SOURCE TO...' on [%s]: %v", addr, err)
	}

	if err := a.StartReplica(ctx, addr); err != nil {
		return err
	}

	return nil
}

func (a *ReplicationAdmin) StartReplica(ctx context.Context, addr string) error {
	c, err := a.cnx.Get(addr)
	if err != nil {
		return err
	}

	if _, err := c.Exec(ctx, startReplicaSQL); err != nil {
		return fmt.Errorf("failed to execute 'START REPLICA' on [%s]: %v", addr, err)
	}

	return nil
}

func (a *ReplicationAdmin) StopReplica(ctx context.Context, addr string) error {
	c, err := a.cnx.Get(addr)
	if err != nil {
		return err
	}

	if _, err := c.Exec(ctx, stopReplicaSQL); err != nil {
		return fmt.Errorf("failed to execute 'STOP REPLICA' on [%s]: %v", addr, err)
	}

	return nil
}

func ScanMap(rows *sql.Rows) (map[string]sql.NullString, error) {

	columns, err := rows.Columns()

	if err != nil {
		return nil, err
	}

	if !rows.Next() {
		err = rows.Err()
		if err != nil {
			return nil, err
		} else {
			return nil, nil
		}
	}

	values := make([]interface{}, len(columns))

	for index := range values {
		values[index] = new(sql.NullString)
	}

	err = rows.Scan(values...)

	if err != nil {
		return nil, err
	}

	result := make(map[string]sql.NullString)

	for index, columnName := range columns {
		result[columnName] = *values[index].(*sql.NullString)
	}

	return result, nil
}

func (a *ReplicationAdmin) GetUser(ctx context.Context, addr, host string, filters []string) (Users, error) {
	users := NewDefaultUsers()

	c, err := a.cnx.Get(addr)
	if err != nil {
		return users, err
	}

	if rows, err := c.Query(ctx, getUserSql, host); err != nil {
		return nil, fmt.Errorf("failed to gather user list with host '%s' on [%s]: %v", host, addr, err)
	} else {
		defer rows.Close()

	queryUserLoop:
		for rows.Next() {
			user := &User{}
			if err := rows.Scan(&user.Username, &user.Authentication); err != nil {
				return users, fmt.Errorf("failed to scan user map on [%s]: %v", addr, err)

			}

			for _, filter := range filters {
				if user.Username == filter {
					continue queryUserLoop
				}
			}
			users[user.Username] = user
		}
	}

	return users, nil
}

func (a *ReplicationAdmin) GetVersion(ctx context.Context, addr string) (string, error) {

	c, err := a.cnx.Get(addr)
	if err != nil {
		return "", err
	}

	var version string
	if err := c.QueryRow(ctx, &version, getVersionSql); err != nil {
		return version, fmt.Errorf("failed to execute 'SELECT VERSION();' on [%s]: %v", addr, err)
	}

	return version, nil
}
