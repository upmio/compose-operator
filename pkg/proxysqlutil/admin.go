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

package proxysqlutil

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"reflect"

	"github.com/go-logr/logr"

	"github.com/upmio/compose-operator/pkg/mysqlutil"
)

// IAdmin proxysql admin interface
type IAdmin interface {
	// Connections returns the connection map of all clients
	Connections() mysqlutil.IAdminConnections

	// Close the admin connections
	Close()

	// SyncMysqlUsers sync users to proxysql
	SyncMysqlUsers(ctx context.Context, addr string, defaultHostGroup, maxConnections int, expectedUsers, foundUsers mysqlutil.Users) error

	// SyncMysqlReplicationHostGroup sync server to proxysql
	SyncMysqlReplicationHostGroup(ctx context.Context, addr string, hostgroup *ReplicationHostgroup) error

	// SyncMysqlReplicationServers sync hostGroup to proxysql
	SyncMysqlReplicationServers(ctx context.Context, addr string, expectedServers, foundServers Servers) error

	// GetRuntimeMysqlUsers get runtime user from proxysql
	GetRuntimeMysqlUsers(ctx context.Context, addr string) (mysqlutil.Users, error)

	// GetRuntimeMysqlReplicationHostGroup get runtime hostGroup from proxysql
	GetRuntimeMysqlReplicationHostGroup(ctx context.Context, addr string, writerHostGroup int) (*ReplicationHostgroup, error)

	// GetRuntimeMysqlReplicationServers get runtime server from proxysql
	GetRuntimeMysqlReplicationServers(ctx context.Context, addr string) (Servers, error)

	// SyncMysqlVersion sync the mysql-server_version variable
	SyncMysqlVersion(ctx context.Context, addr, version string) error
}

// Admin wraps redis cluster admin logic
type Admin struct {
	cnx mysqlutil.IAdminConnections
	log logr.Logger
}

// NewAdmin returns new AdminInterface instance
// at the same time it connects to all Redis ReplicationNodes thanks to the addrs list
func NewAdmin(addrs []string, options *mysqlutil.AdminOptions, log logr.Logger) IAdmin {
	a := &Admin{
		log: log.WithName("proxysql_util"),
	}

	// perform initial connections
	a.cnx = mysqlutil.NewAdminConnections(addrs, options, log)

	return a
}

// Connections returns the connection map of all clients
func (a *Admin) Connections() mysqlutil.IAdminConnections {
	return a.cnx
}

// Close used to close all possible resources instanciate by the Admin
func (a *Admin) Close() {
	a.Connections().Reset()
}

func (a *Admin) GetRuntimeMysqlUsers(ctx context.Context, addr string) (mysqlutil.Users, error) {
	c, err := a.Connections().Get(addr)
	if err != nil {
		return nil, err
	}

	users := mysqlutil.NewDefaultUsers()

	if rows, err := c.Query(ctx, getRuntimeUsernameSql); err != nil {
		return users, fmt.Errorf("failed to query rows from runtime_mysql_users on [%s]: %v", addr, err)
	} else {
		defer func() { _ = rows.Close() }()

		for rows.Next() {
			user := &mysqlutil.User{}
			if err := rows.Scan(&user.Username); err != nil {
				return users, fmt.Errorf("failed to scan runtime_mysql_users on [%s]: %v", addr, err)
			}

			querySql := fmt.Sprintf(getRuntimePasswordSql, user.Username)
			if err := c.QueryRow(ctx, &user.Authentication, querySql); err != nil {
				return users, fmt.Errorf("failed to get password from runtime_mysql_users on [%s]: %v", addr, err)
			}

			users[user.Username] = user
		}
	}

	return users, nil
}

func (a *Admin) SyncMysqlUsers(ctx context.Context, addr string, defaultHostGroup, maxConnections int, expectedUsers, foundUsers mysqlutil.Users) error {
	c, err := a.Connections().Get(addr)
	if err != nil {
		return err
	}

	var execSql string
	for username, userA := range expectedUsers {
		if userB, ok := foundUsers[username]; !ok {
			// insert
			execSql = fmt.Sprintf(insertUserSql, userA.Username, userA.Authentication, defaultHostGroup, maxConnections)
			if _, err := c.Exec(ctx, execSql); err != nil {
				return fmt.Errorf("failed to insert user %s into mysql_users on [%s]: %v ", userA.Username, addr, err)
			}

		} else if !reflect.DeepEqual(userA, userB) {
			// delete
			execSql = fmt.Sprintf(deleteUserSql, userA.Username)
			if _, err := c.Exec(ctx, execSql); err != nil {
				return fmt.Errorf("failed to delete user %s from mysql_users on [%s]: %v ", userA.Username, addr, err)
			}

			// insert
			execSql = fmt.Sprintf(insertUserSql, userA.Username, userA.Authentication, defaultHostGroup, maxConnections)
			if _, err := c.Exec(ctx, execSql); err != nil {
				return fmt.Errorf("failed to insert user %s to mysql_users on [%s]: %v ", userA.Username, addr, err)
			}
		}

		delete(foundUsers, username)
	}

	for _, user := range foundUsers {
		// delete
		execSql = fmt.Sprintf(deleteUserSql, user.Username)
		if _, err := c.Exec(ctx, execSql); err != nil {
			return fmt.Errorf("failed to delete user %s from mysql_users on [%s]: %v ", user.Username, addr, err)
		}

	}

	if _, err := c.Exec(ctx, saveUserSql); err != nil {
		return fmt.Errorf("failed to execute save mysql users to disk on [%s]: %v", addr, err)
	}

	if _, err := c.Exec(ctx, loadUserSql); err != nil {
		return fmt.Errorf("failed to execute load users to runtime on [%s]: %v", addr, err)
	}

	return nil
}

func (a *Admin) GetRuntimeMysqlReplicationHostGroup(ctx context.Context, addr string, writerHostGroupId int) (*ReplicationHostgroup, error) {
	c, err := a.Connections().Get(addr)
	if err != nil {
		return nil, err
	}

	var (
		count    int
		hg       = &ReplicationHostgroup{}
		querySql string
	)

	if err := c.QueryRow(ctx, &count, getRuntimeReplicationHostgroupCountSql); err != nil {
		return hg, fmt.Errorf("failed to get count from runtime_mysql_replication_hostgroups on [%s]: %v", addr, err)
	} else if count != 1 {
		return hg, nil
	}

	querySql = fmt.Sprintf(getReaderHostgroupSql, writerHostGroupId)
	if err := c.QueryRow(ctx, &hg.ReaderHostGroupId, querySql); err != nil && !errors.Is(err, sql.ErrNoRows) {
		return hg, fmt.Errorf("failed to get reader_hostgroup from runtime_mysql_replication_hostgroups on [%s]: %v", addr, err)
	}

	querySql = fmt.Sprintf(getCheckTypeSql, writerHostGroupId)
	if err := c.QueryRow(ctx, &hg.CheckType, querySql); err != nil && !errors.Is(err, sql.ErrNoRows) {
		return hg, fmt.Errorf("failed to get check_type from runtime_mysql_replication_hostgroups on [%s]:  %v", addr, err)
	}

	querySql = fmt.Sprintf(getCommentSql, writerHostGroupId)
	if err := c.QueryRow(ctx, &hg.Comment, querySql); err != nil && !errors.Is(err, sql.ErrNoRows) {
		return hg, fmt.Errorf("failed to get comment from runtime_mysql_replication_hostgroups on [%s]: %v", addr, err)
	}

	return hg, nil
}

func (a *Admin) GetRuntimeMysqlReplicationServers(ctx context.Context, addr string) (Servers, error) {
	c, err := a.Connections().Get(addr)
	if err != nil {
		return nil, err
	}

	servers := make(Servers)

	if rows, err := c.Query(ctx, getMysqlServerSql); err != nil {
		return servers, fmt.Errorf("failed to query rows from runtime_mysql_servers: %v", err)

	} else {
		defer func() { _ = rows.Close() }()

		for rows.Next() {
			server := &Server{}
			if err := rows.Scan(&server.HostGroupId, &server.Hostname, &server.Port, &server.MaxConnections); err != nil {
				return servers, fmt.Errorf("failed to scan runtime_mysql_servers on [%s]: %v", addr, err)
			}

			servers[fmt.Sprintf("%s:%d", server.Hostname, server.Port)] = server
		}
	}

	return servers, nil
}

func (a *Admin) SyncMysqlReplicationHostGroup(ctx context.Context, addr string, hostgroup *ReplicationHostgroup) error {
	c, err := a.Connections().Get(addr)
	if err != nil {
		return err
	}

	// clean table mysql_replication_hostgroup to ensure that table no rows.
	if _, err := c.Exec(ctx, cleanReplicationHostgroupSql); err != nil {
		return fmt.Errorf("failed to clean table mysql_replication_hostgroups on [%s]: %v", addr, err)
	}

	// insert
	execSql := fmt.Sprintf(insertReplicationHostgroupSql, hostgroup.WriterHostGroupId, hostgroup.ReaderHostGroupId, hostgroup.CheckType, hostgroup.Comment)
	if _, err := c.Exec(ctx, execSql); err != nil {
		return fmt.Errorf("failed to insert into mysql_replication_hostgroups on [%s]: %v", addr, err)
	}

	if _, err := c.Exec(ctx, saveServerSql); err != nil {
		return fmt.Errorf("failed to execute save mysql servers to disk on [%s]: %v", addr, err)
	}

	if _, err := c.Exec(ctx, loadServerSql); err != nil {
		return fmt.Errorf("failed to execute load mysql servers to runtime on [%s]: %v", addr, err)
	}

	return nil
}

func (a *Admin) SyncMysqlReplicationServers(ctx context.Context, addr string, expectedServers, foundServers Servers) error {
	c, err := a.Connections().Get(addr)
	if err != nil {
		return err
	}

	var execSql string
	for serverAddr, serverA := range expectedServers {
		if serverB, ok := foundServers[serverAddr]; !ok {
			// insert
			execSql = fmt.Sprintf(insertServersSql, serverA.HostGroupId, serverA.Hostname, serverA.Port, serverA.MaxConnections)
			if _, err := c.Exec(ctx, execSql); err != nil {
				return fmt.Errorf("failed to insert %s into mysql_servers on [%s]: %v", serverA.Hostname, addr, err)
			}

		} else if !reflect.DeepEqual(serverA, serverB) {
			// delete
			execSql = fmt.Sprintf(deleteServerSql, serverA.Hostname, serverA.Port)
			if _, err := c.Exec(ctx, execSql); err != nil {
				return fmt.Errorf("failed to delete %s from mysql_servers on [%s]: %v ", serverA.Hostname, addr, err)
			}

			// insert
			execSql = fmt.Sprintf(insertServersSql, serverA.HostGroupId, serverA.Hostname, serverA.Port, serverA.MaxConnections)
			if _, err := c.Exec(ctx, execSql); err != nil {
				return fmt.Errorf("failed to insert %s into mysql_servers on [%s]: %v ", serverA.Hostname, addr, err)
			}
		}

		delete(foundServers, serverAddr)
	}

	for _, server := range foundServers {
		// delete
		execSql = fmt.Sprintf(deleteServerSql, server.Hostname, server.Port)
		if _, err := c.Exec(ctx, execSql); err != nil {
			return fmt.Errorf("failed to delete server %s from mysql_servers on [%s]: %v ", server.Hostname, addr, err)
		}
	}

	if _, err := c.Exec(ctx, saveServerSql); err != nil {
		return fmt.Errorf("failed to execute save servers to disk on [%s]: %v", addr, err)
	}

	if _, err := c.Exec(ctx, loadServerSql); err != nil {
		return fmt.Errorf("failed to execute load servers to runtime on [%s]: %v", addr, err)
	}

	return nil
}

func (a *Admin) SyncMysqlVersion(ctx context.Context, addr, version string) error {
	c, err := a.Connections().Get(addr)
	if err != nil {
		return err
	}

	var v string
	if err := c.QueryRow(ctx, &v, getMysqlServerVersion); err != nil {
		return fmt.Errorf("failed to get variable mysql-server_version on [%s]: %v", addr, err)
	}

	if version == v {
		return nil
	}

	execSql := fmt.Sprintf(setMysqlServerVersion, version)
	if _, err = c.Exec(ctx, execSql); err != nil {
		return fmt.Errorf("failed to set variable mysql-server_version on [%s]: %v", addr, err)
	}

	return nil
}
