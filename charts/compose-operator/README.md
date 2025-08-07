# Compose Operator Helm Chart

This Helm chart installs the [Compose Operator](https://github.com/upmio/compose-operator) into your Kubernetes (v1.27+) or Openshift cluster (v4.6+).

## TL;DR


```sh
# Add the repo to helm:
helm repo add compose-operator https://upmio.github.io/compose-operator
helm repo update

# Install the operator with auto-generated AES key (recommended for testing)
helm install compose-operator --namespace upm-system --create-namespace \
  --set crds.enabled=true \
  compose-operator/compose-operator

# Or install with custom AES key (recommended for production)
helm install compose-operator --namespace upm-system --create-namespace \
  --set crds.enabled=true \
  --set global.aesKey="your-32-character-aes-key-here" \
  compose-operator/compose-operator
```

## Introduction

The Compose operator is installed into the `upm-system` namespace for Kubernetes clusters.

The Compose and ComposeSet Custom Resource Definitions (CRDs) can either be installed manually (the recommended approach, part of the Helm chart (`crds.enabled=true`).
Installing the CRDs as part of the Helm chart is not recommended for production setups, since uninstalling the Helm chart will also uninstall the CRDs and subsequently delete any remaining CRs.
The CRDs allow you to configure individual parts of your Compose setup:

* [`MysqlGroupReplication`](https://github.com/upmio/compose-operator/blob/main/docs/compose-operator-api.md#mysqlgroupreplication)
* [`MysqlReplication`](https://github.com/upmio/compose-operator/blob/main/docs/compose-operator-api.md#mysqlreplication)
* [`PostgresReplication`](https://github.com/upmio/compose-operator/blob/main/docs/compose-operator-api.md#postgresreplication)
* [`ProxysqlSync`](https://github.com/upmio/compose-operator/blob/main/docs/compose-operator-api.md#proxysqlsync)
* [`RedisCluster`](https://github.com/upmio/compose-operator/blob/main/docs/compose-operator-api.md#rediscluster)
* [`RedisReplication`](https://github.com/upmio/compose-operator/blob/main/docs/compose-operator-api.md#redisreplication)

After the installation of the Compose-operator chart, you can start inject the Custom Resources (CRs) into your cluster.
The Compose operator will then automatically start installing the components.
Please see the documentation of each CR for details.

## Configuration

### AES Key Configuration

The Compose operator uses AES-256-CTR encryption for database password security. You can configure the AES key in two ways:

#### Auto-generated AES Key (Default)
If no `global.aesKey` is specified, a 32-character AES key will be automatically generated using UUID:

```yaml
global:
  aesKey: ""  # Empty value triggers auto-generation
```

#### Custom AES Key
For production environments, specify a custom 32-character AES key:

```yaml
global:
  aesKey: "your-32-character-aes-key-here"
```

### Password Encryption

All database passwords must be AES-256-CTR encrypted before storing in Kubernetes Secrets:

```bash
# Build the AES encryption tool (run from project root)
go build -o aes-tool ./tool/

# Get the AES key used by the compose-operator
# Replace 'compose-operator' with your actual Helm release name and namespace
AES_KEY=$(kubectl get secret compose-operator-aes-secret -n upm-system -o jsonpath='{.data.AES_SECRET_KEY}' | base64 -d)

# Examples for different CRD types:

# For MySQL Replication & Group Replication
aes-tool -key "$AES_KEY" -plaintext "mysql_root_password" -username "mysql"
aes-tool -key "$AES_KEY" -plaintext "replication_password" -username "replication"
kubectl create secret generic mysql-credentials \
  --from-file=mysql=mysql.bin \
  --from-file=replication=replication.bin

# For Redis Replication & Cluster  
aes-tool -key "$AES_KEY" -plaintext "redis_auth_password" -username "redis"
kubectl create secret generic redis-credentials \
  --from-file=redis=redis.bin

# For PostgreSQL Replication
aes-tool -key "$AES_KEY" -plaintext "postgresql_admin_password" -username "postgresql"
aes-tool -key "$AES_KEY" -plaintext "replication_password" -username "replication"
kubectl create secret generic postgres-credentials \
  --from-file=postgresql=postgresql.bin \
  --from-file=replication=replication.bin

# For ProxySQL Sync
aes-tool -key "$AES_KEY" -plaintext "proxysql_admin_password" -username "proxysql"
aes-tool -key "$AES_KEY" -plaintext "mysql_backend_password" -username "mysql"
kubectl create secret generic proxysql-credentials \
  --from-file=proxysql=proxysql.bin \
  --from-file=mysql=mysql.bin

# Method 2 (Legacy): Get base64 encoded values directly
aes-tool -key "$AES_KEY" -plaintext "your_password"
```

### Decrypting Passwords from Kubernetes Secrets

To decrypt passwords stored in existing Kubernetes Secrets:

```bash
# Get the AES key used by the compose-operator
AES_KEY=$(kubectl get secret compose-operator-aes-secret -n upm-system -o jsonpath='{.data.AES_SECRET_KEY}' | base64 -d)

# Extract and decrypt a password from a secret
kubectl get secret mysql-credentials -o jsonpath='{.data.mysql}' | base64 -d > mysql.bin
aes-tool -key "$AES_KEY" -file "mysql.bin"
rm mysql.bin

# One-liner for quick password retrieval
kubectl get secret redis-credentials -o jsonpath='{.data.redis}' | base64 -d > temp.bin && aes-tool -key "$AES_KEY" -file "temp.bin" && rm temp.bin
```

## Uninstalling

Before removing the Compose operator from your cluster, you should first make sure that there are no instances of resources managed by the operator left:

```sh
kubectl get mysqlgroupreplications,mysqlreplications,postgresreplications,proxysqlsyncs,redisreplications,redisclusters --all-namespaces
```

Now you can use Helm to uninstall the Compose operator:

```sh
# for Kubernetes:
helm uninstall --namespace upm-system compose-operator --wait

# optionally remove repository from helm:
helm repo remove compose-operator
```

**Important:** if you installed the CRDs with the Helm chart (by setting `crds.enabled=true`), the CRDs will be removed as well: this means any remaining Compose resources (e.g. Compose Pipelines) in the cluster will be deleted!

If you installed the CRDs manually, you can use the following command to remove them (*this will remove all Compose resources from your cluster*):
```
kubectl delete crd mysqlgroupreplications.upm.syntropycloud.io,mysqlreplications.upm.syntropycloud.io,postgresreplications.upm.syntropycloud.io,proxysqlsyncs.upm.syntropycloud.io,redisreplications.upm.syntropycloud.io,redisclusters.upm.syntropycloud.io --ignore-not-found
```
