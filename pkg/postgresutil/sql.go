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

const (
	getReplicationRoleSql     = `SELECT pg_is_in_recovery()`
	getReplicationStatSql     = `SELECT application_name FROM pg_stat_replication`
	getReplicationModeSql     = `SELECT DISTINCT(sync_state) FROM pg_stat_replication`
	getDataDirSql             = `SHOW data_directory;`
	setPG16PrimaryConnInfoSql = `ALTER SYSTEM SET primary_conninfo='application_name=%s user=%s password=%s host=%s port=%d channel_binding=prefer sslmode=prefer sslcompression=0 sslcertmode=allow sslsni=1 ssl_min_protocol_version=TLSv1.2 gssencmode=prefer krbsrvname=postgres gssdelegation=0 target_session_attrs=any load_balance_hosts=disable'`
	setPG15PrimaryConnInfoSql = `ALTER SYSTEM SET primary_conninfo='application_name=%s user=%s password=%s host=%s port=%d channel_binding=prefer sslmode=prefer sslcompression=0 sslsni=1 ssl_min_protocol_version=TLSv1.2 gssencmode=prefer krbsrvname=postgres target_session_attrs=any'`
	cleanPrimaryConnInfoSql   = `ALTER SYSTEM SET primary_conninfo = ''`
	reloadConfSql             = `SELECT pg_reload_conf()`
	setSyncSql                = `ALTER SYSTEM SET synchronous_standby_names='ANY 1 (%s)'`
	setAsyncSql               = `ALTER SYSTEM SET synchronous_standby_names=''`
	getWalDiffSql             = `SELECT pg_wal_lsn_diff(pg_last_wal_replay_lsn(),pg_last_wal_receive_lsn()) AS wal_diff`
)
