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

const (
	showReplicaStatusSQL       = `SHOW REPLICA STATUS`
	showReadOnlySQL            = `SELECT @@READ_ONLY`
	showSuperReadOnlySQL       = `SELECT @@SUPER_READ_ONLY`
	showSemiSyncReplicaEnabled = `SELECT @@rpl_semi_sync_replica_enabled;`
	showSemiSyncSourceEnabled  = `SELECT @@rpl_semi_sync_source_enabled`

	setSemiSyncTimeout        = `SET GLOBAL rpl_semi_sync_source_timeout = 10000000000000`
	setSemiSyncWaitPoint      = `SET GLOBAL rpl_semi_sync_source_wait_point = AFTER_SYNC`
	setSemiSyncSourceEnabled  = `SET GLOBAL rpl_semi_sync_source_enabled = ?`
	setSemiSyncReplicaEnabled = `SET GLOBAL rpl_semi_sync_replica_enabled = ?`

	setReadOnlySQL      = `SET GLOBAL READ_ONLY = ?`
	setSuperReadOnlySQL = `SET GLOBAL SUPER_READ_ONLY = ?`

	stopReplicaSQL = `STOP REPLICA`

	resetReplicaSQL = `RESET REPLICA ALL`

	waitSourcePosSql = `select SOURCE_POS_WAIT(?,?,120)`

	setReplicaSQL = `CHANGE REPLICATION SOURCE TO 
    SOURCE_HOST = ?,
    SOURCE_PORT = ?,
    SOURCE_USER = ?,
    SOURCE_PASSWORD = ?,
    SOURCE_AUTO_POSITION = 1`
	startReplicaSQL = `START REPLICA`

	getUserSql    = `SELECT user,authentication_string FROM mysql.user WHERE host=?`
	getVersionSql = `SELECT VERSION();`

	getGroupReplicationMembersSQL = `SELECT MEMBER_ID, MEMBER_ROLE, MEMBER_STATE FROM performance_schema.replication_group_members WHERE MEMBER_HOST=? and MEMBER_PORT=?`
	getReadOnlySQL                = `SELECT @@READ_ONLY`
	getSuperReadOnlySQL           = `SELECT @@SUPER_READ_ONLY`
	getGtidExecutedSQL            = `SELECT @@gtid_executed`
	getGroupReplicationSeeds      = `SELECT @@group_replication_group_seeds`
	setGroupReplicationSeeds      = `SET GLOBAL group_replication_group_seeds=?`
	getGroupReplicationAllowList  = `SELECT @@group_replication_ip_allowlist`
	setGroupReplicationAllowList  = `SET GLOBAL group_replication_ip_allowlist=?`

	setGroupReplicationBootStrapON  = `SET GLOBAL group_replication_bootstrap_group=ON`
	setGroupReplicationBootStrapOFF = `SET GLOBAL group_replication_bootstrap_group=OFF`

	startGroupReplicationSQL = `START GROUP_REPLICATION`
	stopGroupReplicationSQL  = `STOP GROUP_REPLICATION`
	setGroupReplicationSQL   = `CHANGE REPLICATION SOURCE TO 
    SOURCE_USER = ?, 
    SOURCE_PASSWORD = ? FOR CHANNEL 'group_replication_recovery'`
)
