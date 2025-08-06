SET @@SESSION.SQL_LOG_BIN=0;
SET @@GLOBAL.GTID_MODE = ON;
INSTALL PLUGIN rpl_semi_sync_source SONAME 'semisync_source.so';
INSTALL PLUGIN rpl_semi_sync_replica SONAME 'semisync_replica.so';
INSTALL PLUGIN group_replication SONAME 'group_replication.so';
INSTALL PLUGIN clone SONAME 'mysql_clone.so';
GRANT REPLICATION CLIENT, REPLICATION SLAVE, SYSTEM_VARIABLES_ADMIN, REPLICATION_SLAVE_ADMIN, GROUP_REPLICATION_ADMIN, RELOAD, BACKUP_ADMIN, CLONE_ADMIN ON *.* TO 'replication'@'%';
GRANT SELECT ON performance_schema.* TO 'replication'@'%';