# API Reference

## Packages
- [upm.syntropycloud.io/v1alpha1](#upmsyntropolycloudiov1alpha1)


## upm.syntropycloud.io/v1alpha1

Package v1alpha1 contains API Schema definitions for the compose-operator v1alpha1 API group

### Resource Types
- [MysqlGroupReplication](#mysqlgroupreplication)
- [MysqlGroupReplicationList](#mysqlgroupreplicationlist)
- [MysqlReplication](#mysqlreplication)
- [MysqlReplicationList](#mysqlreplicationlist)
- [PostgresReplication](#postgresreplication)
- [PostgresReplicationList](#postgresreplicationlist)
- [ProxysqlSync](#proxysqlsync)
- [ProxysqlSyncList](#proxysqlsynclist)
- [RedisCluster](#rediscluster)
- [RedisClusterList](#redisclusterlist)
- [RedisReplication](#redisreplication)
- [RedisReplicationList](#redisreplicationlist)



#### CommonNode



CommonNode information for node to connect



_Appears in:_
- [CommonNodes](#commonnodes)
- [MysqlReplicationSpec](#mysqlreplicationspec)
- [PostgresReplicationSpec](#postgresreplicationspec)
- [RedisReplicationSpec](#redisreplicationspec)

| Field | Description | Default | Validation |
| --- | --- | --- | --- |
| `name` _string_ | Name specifies the identifier of node |  |  |
| `host` _string_ | Host specifies the ip or hostname of node |  |  |
| `port` _integer_ | Port specifies the port of node |  |  |


#### CommonNodes

_Underlying type:_ _[CommonNode](#commonnode)_

CommonNodes array Node



_Appears in:_
- [MysqlGroupReplicationSpec](#mysqlgroupreplicationspec)
- [MysqlReplicationSpec](#mysqlreplicationspec)
- [PostgresReplicationSpec](#postgresreplicationspec)
- [ProxysqlSyncSpec](#proxysqlsyncspec)
- [RedisClusterSpec](#redisclusterspec)
- [RedisReplicationSpec](#redisreplicationspec)

| Field | Description | Default | Validation |
| --- | --- | --- | --- |
| `name` _string_ | Name specifies the identifier of node |  |  |
| `host` _string_ | Host specifies the ip or hostname of node |  |  |
| `port` _integer_ | Port specifies the port of node |  |  |




#### MysqlGroupReplication



MysqlGroupReplication is the Schema for the Mysql Group Replication API


_Appears in:_
- [MysqlGroupReplicationList](#mysqlgroupreplicationlist)

| Field | Description | Default | Validation |
| --- | --- | --- | --- |
| `apiVersion` _string_ | `upm.syntropycloud.io/v1alpha1` | | |
| `kind` _string_ | `MysqlGroupReplication` | | |
| `metadata` _[ObjectMeta](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.22/#objectmeta-v1-meta)_ | Refer to Kubernetes API documentation for fields of `metadata`. |  |  |
| `spec` _[MysqlGroupReplicationSpec](#mysqlgroupreplicationspec)_ | Defines the desired state of the MysqlGroupReplication. |  |  |
| `status` _[MysqlGroupReplicationStatus](#mysqlgroupreplicationstatus)_ | Populated by the system, it represents the current information about the MysqlGroupReplication. |  |  |




#### MysqlGroupReplicationList



MysqlGroupReplicationList contains a list of MysqlGroupReplication



| Field | Description | Default | Validation |
| --- | --- | --- | --- |
| `apiVersion` _string_ | `upm.syntropycloud.io/v1alpha1` | | |
| `kind` _string_ | `MysqlGroupReplicationList` | | |
| `metadata` _[ListMeta](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.22/#listmeta-v1-meta)_ | Refer to Kubernetes API documentation for fields of `metadata`. |  |  |
| `items` _[MysqlGroupReplication](#mysqlgroupreplication) array_ | Contains the list of MysqlGroupReplication. |  |  |




#### MysqlGroupReplicationNode



MysqlGroupReplicationNode represents a node in the MySQL Group Replication topology.

_Appears in:_
- [MysqlGroupReplicationTopology](#mysqlgroupreplicationtopology)

| Field | Description | Default | Validation |
| --- | --- | --- | --- |
| `host` _string_ | Host indicates the host of the MySQL node. |  |  |
| `port` _integer_ | Port indicates the port of the MySQL node. |  |  |
| `role` _[MysqlGroupReplicationRole](#mysqlgroupreplicationrole)_ | Role represents the role of the node in the group replication topology (e.g., primary, secondary). |  |  |
| `status` _[NodeStatus](#nodestatus)_ | Ready indicates whether the node is ready for reads and writes. |  |  |
| `gtidExecuted` _string_ | GtidExecuted indicates the gtid_executed of the MySQL node. |  |  |
| `memberState` _string_ | MemberState indicates the member_state of the MySQL node. |  |  |
| `readonly` _boolean_ | ReadOnly specifies whether the node is read-only. |  |  |
| `superReadonly` _boolean_ | SuperReadOnly specifies whether the node is super-read-only (i.e., cannot even write to its own database). |  |  |




#### MysqlGroupReplicationRole

_Underlying type:_ _string_

MysqlGroupReplicationRole defines the mysql group replication role

_Validation:_
- Enum: [Primary Secondary None]

_Appears in:_
- [MysqlGroupReplicationNode](#mysqlgroupreplicationnode)



#### MysqlGroupReplicationSecret



MysqlGroupReplicationSecret defines the secret information of MysqlGroupReplication

_Appears in:_
- [MysqlGroupReplicationSpec](#mysqlgroupreplicationspec)

| Field | Description | Default | Validation |
| --- | --- | --- | --- |
| `name` _string_ | Name is the name of the secret resource which store authentication information for MySQL. |  |  |
| `mysql` _string_ | Mysql is the key of the secret, which contains the value used to connect to MySQL. | mysql |  |
| `replication` _string_ | Replication is the key of the secret, which contains the value used to set up MySQL Group Replication. | replication |  |




#### MysqlGroupReplicationSpec



MysqlGroupReplicationSpec defines the desired state of MysqlGroupReplication

_Appears in:_
- [MysqlGroupReplication](#mysqlgroupreplication)

| Field | Description | Default | Validation |
| --- | --- | --- | --- |
| `secret` _[MysqlGroupReplicationSecret](#mysqlgroupreplicationsecret)_ | Secret is the reference to the secret resource containing authentication information, it must be in the same namespace as the MysqlGroupReplication object. |  |  |
| `member` _[CommonNodes](#commonnodes)_ | Member is a list of nodes in the MySQL Group Replication topology. |  |  |




#### MysqlGroupReplicationStatus



MysqlGroupReplicationStatus defines the observed state of MysqlGroupReplication

_Appears in:_
- [MysqlGroupReplication](#mysqlgroupreplication)

| Field | Description | Default | Validation |
| --- | --- | --- | --- |
| `topology` _[MysqlGroupReplicationTopology](#mysqlgroupreplicationtopology)_ | Topology indicates the current MySQL Group Replication topology. |  |  |
| `ready` _boolean_ | Ready indicates whether this MysqlGroupReplication object is ready or not. |  |  |
| `conditions` _[Condition](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.22/#condition-v1-meta) array_ | Represents a list of detailed status of the MysqlGroupReplication object. Each condition in the list provides real-time information about certain aspect of the MysqlGroupReplication object.<br /><br />This field is crucial for administrators and developers to monitor and respond to changes within the MysqlGroupReplication. It provides a history of state transitions and a snapshot of the current state that can be used for automated logic or direct inspection. |  |  |




#### MysqlGroupReplicationTopology

_Underlying type:_ _[map[string]*MysqlGroupReplicationNode](#map[string]*mysqlgroupreplicationnode)_



_Appears in:_
- [MysqlGroupReplicationStatus](#mysqlgroupreplicationstatus)



#### MysqlReplication



MysqlReplication is the Schema for the Mysql Replication API



_Appears in:_
- [MysqlReplicationList](#mysqlreplicationlist)

| Field | Description | Default | Validation |
| --- | --- | --- | --- |
| `apiVersion` _string_ | `upm.syntropycloud.io/v1alpha1` | | |
| `kind` _string_ | `MysqlReplication` | | |
| `metadata` _[ObjectMeta](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.22/#objectmeta-v1-meta)_ | Refer to Kubernetes API documentation for fields of `metadata`. |  |  |
| `spec` _[MysqlReplicationSpec](#mysqlreplicationspec)_ | Defines the desired state of the MysqlReplication. |  |  |


#### MysqlReplicationList



MysqlReplicationList contains a list of MysqlReplication





| Field | Description | Default | Validation |
| --- | --- | --- | --- |
| `apiVersion` _string_ | `upm.syntropycloud.io/v1alpha1` | | |
| `kind` _string_ | `MysqlReplicationList` | | |
| `metadata` _[ListMeta](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.22/#listmeta-v1-meta)_ | Refer to Kubernetes API documentation for fields of `metadata`. |  |  |
| `items` _[MysqlReplication](#mysqlreplication) array_ | Contains the list of MysqlReplication. |  |  |


#### MysqlReplicationMode

_Underlying type:_ _string_

MysqlReplicationMode describes how the mysql replication will be handled.
Only one of the following sync mode must be specified.

_Validation:_
- Enum: [rpl_async rpl_semi_sync]

_Appears in:_
- [MysqlReplicationSpec](#mysqlreplicationspec)



#### MysqlReplicationNode



MysqlReplicationNode represents a node in the MySQL replication topology.



_Appears in:_
- [MysqlReplicationTopology](#mysqlreplicationtopology)

| Field | Description | Default | Validation |
| --- | --- | --- | --- |
| `host` _string_ | Host indicates the host of the MySQL node. |  |  |
| `port` _integer_ | Port indicates the port of the MySQL node. |  |  |
| `role` _[MysqlReplicationRole](#mysqlreplicationrole)_ | Role represents the role of the node in the replication topology (e.g., source, replica). |  |  |
| `ready` _boolean_ | Ready indicates whether the node is ready for reads and writes. |  |  |
| `readonly` _boolean_ | ReadOnly specifies whether the node is read-only. |  |  |
| `superReadonly` _boolean_ | SuperReadOnly specifies whether the node is super-read-only (i.e., cannot even write to its own database). |  |  |
| `sourceHost` _string_ | SourceHost indicates the hostname or IP address of the source node that this replica node is replicating from. |  |  |
| `sourcePort` _integer_ | SourcePort indicates the port of the source node that this replica node is replicating from. |  |  |
| `replicaIO` _string_ | ReplicaIO indicates the status of I/O thread of the replica node. |  |  |
| `replicaSQL` _string_ | ReplicaSQL indicates the status of SQL thread of the replica node. |  |  |
| `readSourceLogPos` _integer_ | ReadSourceLogPos the position in the source node's binary log file where the replica node should start reading from. |  |  |
| `sourceLogFile` _string_ | SourceLogFile indicates the name of the binary log file on the source node that the replica node should read from. |  |  |
| `secondsBehindSource` _integer_ | SecondsBehindSource indicates the metric that shows how far behind the source node the replica node is, measured in seconds. |  |  |


#### MysqlReplicationRole

_Underlying type:_ _string_

MysqlReplicationRole defines the mysql replication role



_Appears in:_
- [MysqlReplicationNode](#mysqlreplicationnode)



#### MysqlReplicationSecret



MysqlReplicationSecret defines the secret information of MysqlReplication



_Appears in:_
- [MysqlReplicationSpec](#mysqlreplicationspec)

| Field | Description | Default | Validation |
| --- | --- | --- | --- |
| `name` _string_ | Name is the name of the secret resource which store authentication information for MySQL. |  |  |
| `mysql` _string_ |  Mysql is the key of the secret, which contains the value used to connect to MySQL. | mysql |  |
| `replication` _string_ | Replication is the key of the secret, which contains the value used to set up MySQL replication. | replication |  |


#### MysqlReplicationSpec



MysqlReplicationSpec defines the desired state of MysqlReplication



_Appears in:_
- [MysqlReplication](#mysqlreplication)

| Field | Description | Default | Validation |
| --- | --- | --- | --- |
| `mode` _[MysqlReplicationMode](#mysqlreplicationmode)_ | Mode specifies the mysql replication sync mode.<br />Valid values are:<br />- "rpl_semi_sync": semi_sync;<br />- "rpl_async": async; | rpl_async | Enum: [rpl_async rpl_semi_sync] <br /> |
| `secret` _[MysqlReplicationSecret](#mysqlreplicationsecret)_ | Secret is the reference to the secret resource containing authentication information, it must be in the same namespace as the MysqlReplication object. |  |  |
| `source` _[CommonNode](#commonnode)_ | Source references the source MySQL node. |  |  |
| `service` _[Service](#service)_ | Service references the service providing the MySQL replication endpoint. |  |  |
| `replica` _[CommonNodes](#commonnodes)_ | Replica is a list of replica nodes in the MySQL replication topology. |  |  |




#### MysqlReplicationTopology

_Underlying type:_ _[map[string]*MysqlReplicationNode](#map[string]*mysqlreplicationnode)_

MysqlReplicationTopology defines the MysqlReplication topology



_Appears in:_
- [MysqlReplicationStatus](#mysqlreplicationstatus)





#### PostgresReplication



PostgresReplication is the Schema for the Postgres Replications API



_Appears in:_
- [PostgresReplicationList](#postgresreplicationlist)

| Field | Description | Default | Validation |
| --- | --- | --- | --- |
| `apiVersion` _string_ | `upm.syntropycloud.io/v1alpha1` | | |
| `kind` _string_ | `PostgresReplication` | | |
| `metadata` _[ObjectMeta](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.22/#objectmeta-v1-meta)_ | Refer to Kubernetes API documentation for fields of `metadata`. |  |  |
| `spec` _[PostgresReplicationSpec](#postgresreplicationspec)_ | Defines the desired state of the PostgresReplication. |  |  |


#### PostgresReplicationList



PostgresReplicationList contains a list of PostgresReplication





| Field | Description | Default | Validation |
| --- | --- | --- | --- |
| `apiVersion` _string_ | `upm.syntropycloud.io/v1alpha1` | | |
| `kind` _string_ | `PostgresReplicationList` | | |
| `metadata` _[ListMeta](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.22/#listmeta-v1-meta)_ | Refer to Kubernetes API documentation for fields of `metadata`. |  |  |
| `items` _[PostgresReplication](#postgresreplication) array_ | Contains the list of PostgresReplication. |  |  |


#### PostgresReplicationMode

_Underlying type:_ _string_

PostgresReplicationMode describes how the postgres stream replication will be handled.
Only one of the following sync mode must be specified.

_Validation:_
- Enum: [rpl_async rpl_sync]

_Appears in:_
- [PostgresReplicationSpec](#postgresreplicationspec)



#### PostgresReplicationNode



PostgresReplicationNode represents a node in the Postgres replication topology.



_Appears in:_
- [PostgresReplicationTopology](#postgresreplicationtopology)

| Field | Description | Default | Validation |
| --- | --- | --- | --- |
| `host` _string_ | Host indicates the host of the MySQL node. |  |  |
| `port` _integer_ | Port indicates the port of the MySQL node. |  |  |
| `role` _[PostgresReplicationRole](#postgresreplicationrole)_ | Role represents the role of the node in the replication topology (e.g., primary, standby). |  |  |
| `ready` _boolean_ | Ready indicates whether the node is ready for reads and writes. |  |  |
| `walDiff` _integer_ | WalDiff indicates the standby node pg_wal_lsn_diff(pg_last_wal_replay_lsn(), pg_last_wal_receive_lsn()). |  |  |


#### PostgresReplicationRole

_Underlying type:_ _string_

PostgresReplicationRole defines the postgres replication role



_Appears in:_
- [PostgresReplicationNode](#postgresreplicationnode)



#### PostgresReplicationSecret



PostgresReplicationSecret defines the secret information of PostgresqlReplication



_Appears in:_
- [PostgresReplicationSpec](#postgresreplicationspec)

| Field | Description | Default | Validation |
| --- | --- | --- | --- |
| `name` _string_ | Name is the name of the secret resource which store authentication information for Postgres. |  |  |
| `postgres` _string_ |  Mysql is the key of the secret, which contains the value used to connect to Postgres. | postgres |  |
| `replication` _string_ | Replication is the key of the secret, which contains the value used to set up Postgres replication. | replication |  |


#### PostgresReplicationSpec



PostgresReplicationSpec defines the desired state of PostgresReplication



_Appears in:_
- [PostgresReplication](#postgresreplication)

| Field | Description | Default | Validation |
| --- | --- | --- | --- |
| `mode` _[PostgresReplicationMode](#postgresreplicationmode)_ | Mode specifies the postgres replication sync mode.<br />Valid values are:<br />- "rpl_sync": sync;<br />- "rpl_async": async; | rpl_async | Enum: [rpl_async rpl_sync] <br /> |
| `secret` _[PostgresReplicationSecret](#postgresreplicationsecret)_ | Secret is the reference to the secret resource containing authentication information, it must be in the same namespace as the PostgresReplication object. |  |  |
| `primary` _[CommonNode](#commonnode)_ | Primary references the primary Postgres node. |  |  |
| `service` _[Service](#service)_ | Service references the service providing the Postgres replication endpoint. |  |  |
| `standby` _[CommonNodes](#commonnodes)_ | Standby is a list of standby nodes in the Postgres replication topology. |  |  |




#### PostgresReplicationTopology

_Underlying type:_ _[map[string]*PostgresReplicationNode](#map[string]*postgresreplicationnode)_

PostgresReplicationTopology defines the PostgresReplication topology



_Appears in:_
- [PostgresReplicationStatus](#postgresreplicationstatus)



#### ProxysqlSync



ProxysqlSync is the Schema for the proxysqlsyncs API



_Appears in:_
- [ProxysqlSyncList](#proxysqlsynclist)

| Field | Description | Default | Validation |
| --- | --- | --- | --- |
| `apiVersion` _string_ | `upm.syntropycloud.io/v1alpha1` | | |
| `kind` _string_ | `ProxysqlSync` | | |
| `metadata` _[ObjectMeta](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.22/#objectmeta-v1-meta)_ | Refer to Kubernetes API documentation for fields of `metadata`. |  |  |
| `spec` _[ProxysqlSyncSpec](#proxysqlsyncspec)_ | Defines the desired state of the ProxysqlSync. |  |  |


#### ProxysqlSyncList



ProxysqlSyncList contains a list of ProxysqlSync





| Field | Description | Default | Validation |
| --- | --- | --- | --- |
| `apiVersion` _string_ | `upm.syntropycloud.io/v1alpha1` | | |
| `kind` _string_ | `ProxysqlSyncList` | | |
| `metadata` _[ListMeta](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.22/#listmeta-v1-meta)_ | Refer to Kubernetes API documentation for fields of `metadata`. |  |  |
| `items` _[ProxysqlSync](#proxysqlsync) array_ | Contains the list of ProxysqlSync. |  |  |


#### ProxysqlSyncNode







_Appears in:_
- [ProxysqlSyncTopology](#proxysqlsynctopology)

| Field | Description | Default | Validation |
| --- | --- | --- | --- |
| `users` _[Users](#users)_ | Users references the user list synced from MySQL to ProxySQL |  |  |
| `ready` _boolean_ | Ready references whether ProxySQL server synced from MysqlReplication is correct. |  |  |


#### ProxysqlSyncSecret



ProxysqlSyncSecret defines the secret information of ProxysqlSync



_Appears in:_
- [ProxysqlSyncSpec](#proxysqlsyncspec)

| Field | Description | Default | Validation |
| --- | --- | --- | --- |
| `name` _string_ | Name is the name of the secret resource which store authentication information for MySQL and ProxySQL. |  |  |
| `proxysql` _string_ |  Proxysql is the key of the secret, which contains the value used to connect to ProxySQL. |  |  |
| `mysql` _string_ |  Mysql is the key of the secret, which contains the value used to connect to MySQL. |  |  |


#### ProxysqlSyncSpec



ProxysqlSyncSpec defines the desired state of ProxysqlSync



_Appears in:_
- [ProxysqlSync](#proxysqlsync)

| Field | Description | Default | Validation |
| --- | --- | --- | --- |
| `proxysql` _[CommonNodes](#commonnodes)_ | Proxysql references the list of proxysql nodes. |  |  |
| `secret` _[ProxysqlSyncSecret](#proxysqlsyncsecret)_ | Secret is the reference to the secret resource containing authentication information, it must be in the same namespace as the ProxysqlSync object. |  |  |
| `mysqlReplication` _string_ | Source references the source MySQL node. |  |  |
| `rule` _[Rule](#rule)_ | Rule references the rule of sync users from MySQL to ProxySQL. |  |  |




#### ProxysqlSyncTopology

_Underlying type:_ _[map[string]*ProxysqlSyncNode](#map[string]*proxysqlsyncnode)_

ProxysqlSyncTopology defines the MysqlReplication topology



_Appears in:_
- [ProxysqlSyncStatus](#proxysqlsyncstatus)



#### RedisCluster



RedisCluster is the Schema for the Redis Cluster API



_Appears in:_
- [RedisClusterList](#redisclusterlist)

| Field | Description | Default | Validation |
| --- | --- | --- | --- |
| `apiVersion` _string_ | `upm.syntropycloud.io/v1alpha1` | | |
| `kind` _string_ | `RedisCluster` | | |
| `metadata` _[ObjectMeta](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.22/#objectmeta-v1-meta)_ | Refer to Kubernetes API documentation for fields of `metadata`. |  |  |
| `spec` _[RedisClusterSpec](#redisclusterspec)_ | Defines the desired state of the RedisCluster. |  |  |


#### RedisClusterList



RedisClusterList contains a list of RedisCluster





| Field | Description | Default | Validation |
| --- | --- | --- | --- |
| `apiVersion` _string_ | `upm.syntropycloud.io/v1alpha1` | | |
| `kind` _string_ | `RedisClusterList` | | |
| `metadata` _[ListMeta](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.22/#listmeta-v1-meta)_ | Refer to Kubernetes API documentation for fields of `metadata`. |  |  |
| `items` _[RedisCluster](#rediscluster) array_ | Contains the list of RedisCluster. |  |  |


#### RedisClusterNode



RedisClusterNode represent a RedisCluster Node



_Appears in:_
- [RedisClusterTopology](#redisclustertopology)

| Field | Description | Default | Validation |
| --- | --- | --- | --- |
| `id` _string_ | ID represents the id of the Redis node. |  |  |
| `role` _[RedisClusterNodeRole](#redisclusternoderole)_ | Role represents the role of the node in the redis cluster topology (e.g., source, replica). |  |  |
| `host` _string_ | Host indicates the host of the Redis node. |  |  |
| `port` _integer_ | Port indicates the port of the Redis node. |  |  |
| `slots` _string array_ | Slots indicates the slots assigned to this Redis node. |  |  |
| `masterRef` _string_ | MasterRef indicates which source node this replica node reference. |  |  |
| `shard` _string_ | Shard indicates which shard this node belong with. |  |  |
| `ready` _boolean_ | Ready indicates whether the node is ready for reads and writes. |  |  |


#### RedisClusterNodeRole

_Underlying type:_ _string_

RedisClusterNodeRole defines the redis cluster role



_Appears in:_
- [RedisClusterNode](#redisclusternode)



#### RedisClusterSecret



RedisClusterSecret defines the secret information of RedisCluster



_Appears in:_
- [RedisClusterSpec](#redisclusterspec)

| Field | Description | Default | Validation |
| --- | --- | --- | --- |
| `name` _string_ | Name is the name of the secret resource which store authentication information for Redis. |  |  |
| `redis` _string_ |  Redis is the key of the secret, which contains the value used to connect to Redis. |  |  |


#### RedisClusterSpec



RedisClusterSpec defines the desired state of RedisCluster



_Appears in:_
- [RedisCluster](#rediscluster)

| Field | Description | Default | Validation |
| --- | --- | --- | --- |
| `secret` _[RedisClusterSecret](#redisclustersecret)_ | Secret is the reference to the secret resource containing authentication information, it must be in the same namespace as the RedisReplication object. |  |  |
| `members` _object (keys:string, values:[CommonNodes](#commonnodes))_ | Members is a list of nodes in the Redis Cluster topology. |  |  |




#### RedisClusterTopology

_Underlying type:_ _[map[string]*RedisClusterNode](#map[string]*redisclusternode)_

RedisClusterTopology defines the RedisCluster topology



_Appears in:_
- [RedisClusterStatus](#redisclusterstatus)



#### RedisReplication



RedisReplication is the Schema for the Redis Replication API



_Appears in:_
- [RedisReplicationList](#redisreplicationlist)

| Field | Description | Default | Validation |
| --- | --- | --- | --- |
| `apiVersion` _string_ | `upm.syntropycloud.io/v1alpha1` | | |
| `kind` _string_ | `RedisReplication` | | |
| `metadata` _[ObjectMeta](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.22/#objectmeta-v1-meta)_ | Refer to Kubernetes API documentation for fields of `metadata`. |  |  |
| `spec` _[RedisReplicationSpec](#redisreplicationspec)_ | Defines the desired state of the RedisReplication. |  |  |


#### RedisReplicationList



RedisReplicationList contains a list of RedisReplication





| Field | Description | Default | Validation |
| --- | --- | --- | --- |
| `apiVersion` _string_ | `upm.syntropycloud.io/v1alpha1` | | |
| `kind` _string_ | `RedisReplicationList` | | |
| `metadata` _[ListMeta](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.22/#listmeta-v1-meta)_ | Refer to Kubernetes API documentation for fields of `metadata`. |  |  |
| `items` _[RedisReplication](#redisreplication) array_ | Contains the list of RedisReplication. |  |  |


#### RedisReplicationNode



RedisReplicationNode represents a node in the Redis replication topology.



_Appears in:_
- [RedisReplicationTopology](#redisreplicationtopology)

| Field | Description | Default | Validation |
| --- | --- | --- | --- |
| `host` _string_ | Host indicates the host of the Redis node. |  |  |
| `port` _integer_ | Port indicates the port of the Redis node. |  |  |
| `role` _[RedisReplicationRole](#redisreplicationrole)_ | Role represents the role of the node in the replication topology (e.g., source, replica). |  |  |
| `ready` _boolean_ | Ready indicates whether the node is ready for reads and writes. |  |  |
| `sourceHost` _string_ | SourceHost indicates the hostname or IP address of the source node that this replica node is replicating from. |  |  |
| `sourcePort` _integer_ | SourcePort indicates the port of the source node that this replica node is replicating from. |  |  |


#### RedisReplicationRole

_Underlying type:_ _string_

RedisReplicationRole defines the redis replication role



_Appears in:_
- [RedisReplicationNode](#redisreplicationnode)



#### RedisReplicationSecret



RedisReplicationSecret defines the secret information of RedisReplication



_Appears in:_
- [RedisReplicationSpec](#redisreplicationspec)

| Field | Description | Default | Validation |
| --- | --- | --- | --- |
| `name` _string_ | Name is the name of the secret resource which store authentication information for Redis. |  |  |
| `redis` _string_ |  Redis is the key of the secret, which contains the value used to connect to Redis. |  |  |


#### RedisReplicationSpec



RedisReplicationSpec defines the desired state of RedisReplication



_Appears in:_
- [RedisReplication](#redisreplication)

| Field | Description | Default | Validation |
| --- | --- | --- | --- |
| `secret` _[RedisReplicationSecret](#redisreplicationsecret)_ | Secret is the reference to the secret resource containing authentication information, it must be in the same namespace as the RedisReplication object. |  |  |
| `source` _[CommonNode](#commonnode)_ | Source references the source Redis node. |  |  |
| `replica` _[CommonNodes](#commonnodes)_ | Replica is a list of replica nodes in the Redis replication topology. |  |  |
| `service` _[Service](#service)_ | Service references the service providing the Redis replication endpoint. |  |  |
| `sentinel` _string array_ | List of Sentinel pod names to label with the current Redis source pod name. The operator writes label `compose-operator/redis-replication.source` on each listed Sentinel pod to the detected source pod (or `unknown` when not determinable). This enables the Sentinel container to inject the active master into its configuration after restarts. Optional; when empty, Sentinel labeling is skipped. |  |  |



#### RedisReplicationTopology

_Underlying type:_ _[map[string]*RedisReplicationNode](#map[string]*redisreplicationnode)_

RedisReplicationTopology defines the RedisReplication topology



_Appears in:_
- [RedisReplicationStatus](#redisreplicationstatus)



#### Rule



Rule defines the proxysqlsync rule



_Appears in:_
- [ProxysqlSyncSpec](#proxysqlsyncspec)

| Field | Description | Default | Validation |
| --- | --- | --- | --- |
| `filter` _string array_ |  |  |  |
| `pattern` _string_ |  |  |  |


#### Service



Service reflection of kubernetes service



_Appears in:_
- [MysqlReplicationSpec](#mysqlreplicationspec)
- [PostgresReplicationSpec](#postgresreplicationspec)
- [RedisReplicationSpec](#redisreplicationspec)

| Field | Description | Default | Validation |
| --- | --- | --- | --- |
| `type` _[ServiceType](#servicetype)_ | Type string describes ingress methods for a service | ClusterIP | Enum: [ClusterIP NodePort LoadBalancer ExternalName] <br /> |


#### ServiceType

_Underlying type:_ _string_

ServiceType reflection of kubernetes service type

_Validation:_
- Enum: [ClusterIP NodePort LoadBalancer ExternalName]

_Appears in:_
- [Service](#service)



#### Users

_Underlying type:_ _string array_

Users indicates the list of syncd users from mysql to proxysql.



_Appears in:_
- [ProxysqlSyncNode](#proxysqlsyncnode)



