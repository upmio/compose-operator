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

const (
	getRuntimePasswordSql = `SELECT DISTINCT(password) FROM runtime_mysql_users WHERE username='%s'`
	getRuntimeUsernameSql = `SELECT DISTINCT(username) FROM runtime_mysql_users`

	insertUserSql = `INSERT INTO mysql_users(username,password,default_hostgroup,max_connections) VALUES ('%s','%s',%d,%d)`
	deleteUserSql = `DELETE FROM mysql_users WHERE username='%s'`

	loadUserSql = `LOAD MYSQL USERS TO RUNTIME`
	saveUserSql = `SAVE MYSQL USERS TO DISK`

	getRuntimeReplicationHostgroupCountSql = `SELECT COUNT(*) FROM runtime_mysql_replication_hostgroups`
	getReaderHostgroupSql                  = `SELECT reader_hostgroup FROM runtime_mysql_replication_hostgroups WHERE writer_hostgroup=%d`
	getCheckTypeSql                        = `SELECT check_type FROM runtime_mysql_replication_hostgroups WHERE writer_hostgroup=%d`
	getCommentSql                          = `SELECT comment FROM runtime_mysql_replication_hostgroups WHERE writer_hostgroup=%d`
	cleanReplicationHostgroupSql           = `DELETE FROM mysql_replication_hostgroups`
	insertReplicationHostgroupSql          = `INSERT INTO mysql_replication_hostgroups(writer_hostgroup,reader_hostgroup,check_type,comment) VALUES (%d,%d,'%s','%s')`

	getMysqlServerSql = `SELECT hostgroup_id,hostname,port,max_connections FROM runtime_mysql_servers;`
	insertServersSql  = `INSERT INTO mysql_servers(hostgroup_id,hostname,port,max_connections) VALUES (%d,'%s',%d,%d)`
	deleteServerSql   = `DELETE FROM mysql_servers WHERE hostname='%s' and port=%d`

	loadServerSql = `LOAD MYSQL SERVERS TO RUNTIME`
	saveServerSql = `SAVE MYSQL SERVERS TO DISK`

	getMysqlServerVersion = "SELECT variable_value FROM global_variables WHERE variable_name='mysql-server_version';"
	setMysqlServerVersion = "SET mysql-server_version='%s';"
)
