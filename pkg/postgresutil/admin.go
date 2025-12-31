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

package postgresutil

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"net"
	"strconv"
	"strings"

	"github.com/go-logr/logr"
)

// IAdmin mysql admin interface
type IAdmin interface {
	// Connections returns the connection map of all clients
	Connections() IAdminConnections

	// Close the admin connections
	Close()

	// GetReplicationStatus return the Nodes infos for all nodes
	GetReplicationStatus(ctx context.Context) *ReplicationInfo

	// ConfigureStandby set postgres standby primary connect info
	ConfigureStandby(ctx context.Context, addr, nodeName, sourceHost, sourcePort, replicationUser, replicationPwd string) error

	// ConfigureSyncMode set postgres replication to synchronous
	ConfigureSyncMode(ctx context.Context, addr, standbyList string) error

	// ConfigureAsyncMode set postgres replication to asynchronous
	ConfigureAsyncMode(ctx context.Context, addr string) error

	// ConfigureIsolated clean postgres replication config
	ConfigureIsolated(ctx context.Context, addr string) error
}

// AdminOptions optional options for redis admin
type AdminOptions struct {
	ConnectionTimeout int
	Username          string
	Password          string
}

// Admin wraps redis cluster admin logic
type Admin struct {
	cnx IAdminConnections
	log logr.Logger
}

// NewAdmin returns new AdminInterface instance
// at the same time it connects to all Redis Nodes thanks to the addrs list
func NewAdmin(addrs []string, options *AdminOptions, log logr.Logger) IAdmin {
	a := &Admin{
		log: log.WithName("postgres_util"),
	}

	// perform initial connections
	a.cnx = NewAdminConnections(addrs, options, log)

	return a
}

// Connections returns the connection map of all clients
func (a *Admin) Connections() IAdminConnections {
	return a.cnx
}

// Close used to close all possible resources instanciate by the Admin
func (a *Admin) Close() {
	a.Connections().Reset()
}

// GetReplicationStatus return the Nodes infos for all nodes
func (a *Admin) GetReplicationStatus(ctx context.Context) *ReplicationInfo {
	infos := NewReplicationInfo()

	for addr := range a.Connections().GetAll() {
		node, err := a.getInfos(ctx, addr)
		if err != nil {
			a.log.Error(err, "failed to gather postgres replication information")
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

func (a *Admin) getInfos(ctx context.Context, addr string) (*Node, error) {
	c, err := a.cnx.Get(addr)
	if err != nil {
		return nil, err
	}

	// Parse host and port from address
	host, portStr, err := net.SplitHostPort(addr)
	if err != nil {
		return nil, fmt.Errorf("unable to split host and port from address: %s", addr)
	}

	port, err := strconv.Atoi(portStr)
	if err != nil {
		return nil, fmt.Errorf("failed to convert port from string to int")
	}

	node := NewDefaultNode()
	node.Host = host
	node.Port = port

	// Determine node role
	var roleStatus string
	if err := c.QueryRow(ctx, &roleStatus, getReplicationRoleSql); err != nil {
		return nil, fmt.Errorf("failed to query pg_is_in_recovery() on [%s]: %v", addr, err)
	}

	// Handle role-specific logic
	switch roleStatus {
	case "false":
		if err := a.handlePrimaryNode(ctx, addr, node); err != nil {
			return nil, err
		}
	case "true":
		if err := a.handleStandbyNode(ctx, addr, node); err != nil {
			return nil, err
		}
	default:
		return nil, fmt.Errorf("unexpected replication role status: %q", roleStatus)
	}

	// Get common data directory
	if err := c.QueryRow(ctx, &node.DataDir, getDataDirSql); err != nil {
		return nil, fmt.Errorf("failed to get data directory: %v", err)
	}

	return node, nil
}

func (a *Admin) handlePrimaryNode(ctx context.Context, addr string, node *Node) error {
	c, err := a.cnx.Get(addr)
	if err != nil {
		return err
	}

	node.Role = PostgresPrimaryRole

	// Get replication statistics
	rows, err := c.Query(ctx, getReplicationStatSql)
	if err != nil {
		return fmt.Errorf("failed to query application_name from pg_stat_replication on [%s]: %v", addr, err)
	}
	defer func() { _ = rows.Close() }()

	for rows.Next() {
		var name string
		if err := rows.Scan(&name); err != nil {
			return fmt.Errorf("failed to scan pg_stat_replication on [%s]: %v", addr, err)
		}
		node.ReplicationStat = append(node.ReplicationStat, name)
	}
	if err := rows.Err(); err != nil {
		return fmt.Errorf("failed to query pg_stat_replication on [%s]: %v", addr, err)
	}

	// Get replication mode
	if err := c.QueryRow(ctx, &node.ReplicationMode, getReplicationModeSql); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			node.ReplicationMode = "async"
			return nil
		}
		return fmt.Errorf("failed to query replication sync state on [%s]: %v", addr, err)
	}
	return nil
}

func (a *Admin) handleStandbyNode(ctx context.Context, addr string, node *Node) error {
	c, err := a.cnx.Get(addr)
	if err != nil {
		return err
	}

	node.Role = PostgresStandbyRole

	var walDiff sql.NullInt64
	if err := c.QueryRow(ctx, &walDiff, getWalDiffSql); err != nil {
		return fmt.Errorf("failed to get wal diff on [%s]: %v", addr, err)
	}

	if walDiff.Valid {
		node.WalDiff = int(walDiff.Int64)
	} else {
		node.WalDiff = 0
	}
	return nil
}

func (a *Admin) ConfigureSyncMode(ctx context.Context, addr, standbyList string) error {
	c, err := a.cnx.Get(addr)
	if err != nil {
		return err
	}

	sqlStr := fmt.Sprintf(setSyncSql, standbyList)
	if _, err := c.Exec(ctx, sqlStr); err != nil {
		return fmt.Errorf("failed to set synchronous_standby_names='ANY 1 (%s)' on [%s]: %v", standbyList, addr, err)
	}

	if _, err := c.Exec(ctx, reloadConfSql); err != nil {
		return fmt.Errorf("failed to reload config on %s: %v", addr, err)
	}

	return nil
}

func (a *Admin) ConfigureAsyncMode(ctx context.Context, addr string) error {
	c, err := a.cnx.Get(addr)
	if err != nil {
		return err
	}

	if _, err := c.Exec(ctx, setAsyncSql); err != nil {
		return fmt.Errorf("failed to set synchronous_standby_names='' on [%s]: %v", addr, err)
	}

	if _, err := c.Exec(ctx, reloadConfSql); err != nil {
		return fmt.Errorf("failed to reload config on %s: %v", addr, err)
	}

	return nil
}

func (a *Admin) ConfigureStandby(ctx context.Context, addr, nodeName, sourceHost, sourcePort, replicationUser, replicationPwd string) error {
	c, err := a.cnx.Get(addr)
	if err != nil {
		return err
	}

	port, err := strconv.Atoi(sourcePort)
	if err != nil {
		return err
	}

	var version, sqlStr string

	if err := c.QueryRow(ctx, &version, "SELECT VERSION()"); err != nil {
		return fmt.Errorf("failed to get PostgreSQL version: %v", err)
	}

	appName := strings.ReplaceAll(nodeName, "-", "_")

	if strings.HasPrefix(version, "PostgreSQL 15.") {
		sqlStr = fmt.Sprintf(setPG15PrimaryConnInfoSql, appName, replicationUser, replicationPwd, sourceHost, port)
	} else if strings.HasPrefix(version, "PostgreSQL 16.") {
		sqlStr = fmt.Sprintf(setPG16PrimaryConnInfoSql, appName, replicationUser, replicationPwd, sourceHost, port)
	} else {
		return fmt.Errorf("version %s is unsupported", version)
	}

	if _, err := c.Exec(ctx, sqlStr); err != nil {
		return fmt.Errorf("failed to configure replication on %s: %v", addr, err)
	}

	if _, err := c.Exec(ctx, reloadConfSql); err != nil {
		return fmt.Errorf("failed to reload config on %s: %v", addr, err)
	}

	return nil
}

func (a *Admin) ConfigureIsolated(ctx context.Context, addr string) error {
	c, err := a.cnx.Get(addr)
	if err != nil {
		return err
	}

	if _, err := c.Exec(ctx, cleanPrimaryConnInfoSql); err != nil {
		return fmt.Errorf("failed to clean replication on %s: %v", addr, err)
	}

	if _, err := c.Exec(ctx, reloadConfSql); err != nil {
		return fmt.Errorf("failed to reload config on %s: %v", addr, err)
	}

	return nil
}
