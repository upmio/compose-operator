# Compose Operator

[![Go Report Card](https://goreportcard.com/badge/github.com/upmio/compose-operator)](https://goreportcard.com/report/github.com/upmio/compose-operator)
[![License](https://img.shields.io/badge/License-Apache%202.0-blue.svg)](https://github.com/upmio/compose-operator/blob/main/LICENSE)
[![Release](https://img.shields.io/github/v/release/upmio/compose-operator)](https://github.com/upmio/compose-operator/releases)
[![Stars](https://img.shields.io/github/stars/upmio/compose-operator)](https://github.com/upmio/compose-operator)

<img src="./docs/compose-operator-icon.png" alt="Compose Operator" height="128">

A Kubernetes operator for managing database replication topologies across MySQL, Redis, PostgreSQL, and ProxySQL instances.

## Overview

The Compose Operator is a cloud-native solution that manages traditional database applications through Kubernetes Custom Resource Definitions (CRDs). Unlike other database operators, it focuses solely on managing replication relationships between existing database pods within Kubernetes clusters, without creating the underlying Pod resources itself.

### Key Features

- **🔄 Multi-Database Support**: MySQL, Redis, PostgreSQL, and ProxySQL
- **☸️ Kubernetes Native**: Manage database instances running as pods within Kubernetes clusters
- **⚡ Topology Management**: Real-time replication status monitoring and manual switchover support
- **📊 Monitoring**: Real-time status tracking and metrics collection
- **🛡️ Security**: Encrypted credential management and secure connections
- **🔧 Read/Write Separation**: Automatic read/write service creation supporting application read/write splitting

### Supported Database Types

| Database | Replication Types | Service Creation |
|----------|-------------------|------------------|
| **MySQL** | Source-Replica, Group Replication | ✅ |
| **Redis** | Source-Replica, Cluster | ✅ |
| **PostgreSQL** | Primary-Standby | ✅ |
| **ProxySQL** | User/Server Sync | ❌ |

## Table of Contents

- [Installation](#installation)
- [Quick Start](#quick-start)
- [Monitoring](#monitoring-and-switchover)
- [Examples](#examples)
- [Architecture](#architecture)
- [Configuration](#configuration)
- [Development](#development)
- [Contributing](#contributing)
- [FAQ](#faq)
- [Security](#security)
- [License](#license)

## Installation

### Prerequisites

- Kubernetes 1.29+
- Helm 3.0+
- Database instances with appropriate user permissions

### Helm Installation

```bash
# Add the Helm repository
helm repo add compose-operator https://upmio.github.io/compose-operator
helm repo update

# Install the operator
helm install compose-operator compose-operator/compose-operator \
  --namespace upm-system \
  --create-namespace \
```

### Manual Installation

```bash
# Install CRDs
kubectl apply -f https://github.com/upmio/compose-operator/releases/latest/download/crds.yaml

# Create namespace
kubectl create namespace upm-system

# Install the operator  
kubectl apply -f https://github.com/upmio/compose-operator/releases/latest/download/operator.yaml -n upm-system
```

### Verify Installation

```bash
kubectl get pods -n upm-system
kubectl get crd | grep upm.syntropycloud.io
```

## Quick Start

### 1. Prepare Database Credentials

Create a secret with encrypted database passwords using binary files (recommended):

```bash
# Build the AES encryption tool
go build -o aes-tool ./tool/

# Get the AES key used by the compose-operator
# Replace 'compose-operator' with your actual Helm release name
AES_KEY=$(kubectl get secret aes-secret-key -n upm-system -o jsonpath='{.data.AES_SECRET_KEY}' | base64 -d)

# Encrypt passwords and save to binary files
aes-tool -key "$AES_KEY" -plaintext "mysql_root_password" -username "mysql"
aes-tool -key "$AES_KEY" -plaintext "replication_password" -username "replication"

# Create secret from binary files
kubectl create secret generic mysql-credentials \
  -n upm-system \
  --from-file=mysql=mysql.bin \
  --from-file=replication=replication.bin

# Clean up binary files
rm mysql.bin replication.bin
```

### 2. Create MySQL Replication

```yaml
apiVersion: upm.syntropycloud.io/v1alpha1
kind: MysqlReplication
metadata:
  name: mysql-replication-semi-sync-sample
spec:
  mode: rpl_semi_sync
  secret:
    name: mysql-credentials
    mysql: mysql
    replication: replication
  source:
    host: mysql-semi-sync-node-0.default.svc.cluster.local
    name: mysql-semi-sync-node-0
    port: 3306
  replica:
    - host: mysql-semi-sync-node-1.default.svc.cluster.local
      name: mysql-semi-sync-node-1
      port: 3306
    - host: mysql-semi-sync-node-2.default.svc.cluster.local
      name: mysql-semi-sync-node-2
      port: 3306
  service:
    type: ClusterIP
```

### 3. Apply and Verify

```bash
kubectl apply -f mysql-replication.yaml

# Check status
kubectl get mysqlreplication mysql-replication-semi-sync-sample -o yaml

# Monitor replication status
kubectl describe mysqlreplication mysql-replication-semi-sync-sample

# Verify services created
kubectl get services | grep mysql-replication-semi-sync-sample

# Watch for status changes
kubectl get mysqlreplication mysql-replication-semi-sync-sample -w
```

## Monitoring and Switchover

### Status Monitoring

The operator continuously monitors replication health and reports detailed status:

```bash
# Check overall status
kubectl get mysqlreplication,redisreplication,postgresreplication -A

# View detailed topology status
kubectl describe mysqlreplication mysql-replication-semi-sync-sample

# Monitor real-time status changes
kubectl get mysqlreplication mysql-replication-semi-sync-sample -w -o jsonpath='{.status}'
```

### Manual Switchover/Failover

When issues are detected through status monitoring, perform manual intervention:

```bash
# 1. Observe the current status
kubectl get mysqlreplication mysql-replication-semi-sync-sample -o jsonpath='{.status.topology}' | jq

# 2. Identify the issue (e.g., source node failure)
kubectl describe mysqlreplication mysql-replication-semi-sync-sample

# 3. Update spec to promote a replica to source
kubectl patch mysqlreplication mysql-replication-semi-sync-sample --type='merge' -p='
spec:
  source:
    name: mysql-semi-sync-node-1  # Promote replica to source
    host: mysql-semi-sync-node-1.default.svc.cluster.local
    port: 3306
  replica:
    - name: mysql-semi-sync-node-0
      host: mysql-semi-sync-node-0.default.svc.cluster.local
      port: 3306
    - name: mysql-semi-sync-node-2
      host: mysql-semi-sync-node-2.default.svc.cluster.local
      port: 3306
'

# 4. Monitor the reconfiguration
kubectl get mysqlreplication mysql-replication-semi-sync-sample -w
```

## Examples

Comprehensive examples are available in the [examples/](examples/) directory:

- [MySQL Source-Replica Replication](examples/mysql/basic-replication.yaml)
- [MySQL Group Replication](examples/mysql/group-replication.yaml)
- [Redis Source-Replica Replication](examples/redis/basic-replication.yaml)
- [Redis Cluster](examples/redis/cluster.yaml)
- [PostgreSQL Streaming Replication](examples/postgres/streaming-replication.yaml)
- [ProxySQL with MySQL Integration](examples/proxysql/mysql-integration.yaml)

## Architecture

### Overview

The Compose Operator follows the Kubernetes operator pattern, extending the Kubernetes API with Custom Resource Definitions (CRDs) to manage database replication topologies.

```text
┌─────────────────┐    ┌──────────────────┐    ┌─────────────────┐
│   User/GitOps   │───▶│  Kubernetes API  │───▶│ Compose Operator│
└─────────────────┘    └──────────────────┘    └─────────────────┘
                                                         │
                        ┌────────────────────────────────┼──────────────────────────────────┐
                        │                                │                                  │
                        ▼                                ▼                                  ▼
                ┌───────────────┐                ┌───────────────┐                ┌───────────────┐
                │ MySQL Cluster │                │ Redis Cluster │                │ PostgreSQL    │
                │               │                │               │                │ Cluster       │
                │ ┌───────────┐ │                │ ┌────────────┐│                │               │
                │ │   Source  │ │                │ │   Source   ││                │ ┌───────────┐ │
                │ └───────────┘ │                │ └────────────┘│                │ │  Primary  │ │
                │ ┌───────────┐ │                │ ┌────────────┐│                │ └───────────┘ │
                │ │ Replica 1 │ │                │ │ Replica 1  ││                │ ┌───────────┐ │
                │ └───────────┘ │                │ └────────────┘│                │ │ Standby 1 │ │
                │ ┌───────────┐ │                │ ┌────────────┐│                │ └───────────┘ │
                │ │ Replica 2 │ │                │ │ Replica 2  ││                └───────────────┘
                │ └───────────┘ │                │ └────────────┘│
                └───────────────┘                └───────────────┘
```

### Components

#### Custom Resource Definitions (CRDs)

- **MysqlReplication**: Manages MySQL source-replica replication
- **MysqlGroupReplication**: Manages MySQL Group Replication
- **RedisReplication**: Manages Redis source-replica replication
  > **Note**: The Compose Operator focuses on managing Redis replication relationships directly. However, when Redis Sentinel is required for high availability, you can use the `skip-reconcile` design pattern:
  >
  > 1. **Controller Behavior**: When the `compose-operator.skip.reconcile: "true"` annotation is present, the controller only:
  >
  >       \- Updates the `status` field based on actual topology
  >
  >       \- Compares `spec` with `status` to determine `ready` state
  >
  >       \- Does NOT manage the replication configuration
  >
  > 2. **Sentinel Integration Challenge**: When Sentinel performs failover, it creates a mismatch between `spec` and `status`, causing the resource to appear not ready.
  >
  > 3. **Solution via Unit-Operator**: To address this design challenge, we implement a complementary solution in [github.com/upmio/unit-operator](https://github.com/upmio/unit-operator):
  >
  >       \- Sentinel calls unit-agent API after failover
  >
  >       \- Unit-agent updates the RedisReplication spec to match current topology
  >
  >       \- This ensures spec-status consistency and maintains ready state
  >
  >    This design pattern allows Redis Sentinel's high availability features while preserving the declarative nature of Kubernetes resources.
  >
  >    Additionally, when `spec.sentinel` is provided, the controller writes the label `compose-operator.redisreplication.source=<current-source-pod>` onto each listed Sentinel pod (or `unknown` if not determinable). Your Sentinel container can read this label at startup to inject the active master into its configuration.
- **RedisCluster**: Manages Redis cluster topology
- **PostgresReplication**: Manages PostgreSQL primary-standby streaming replication
  > **Note**: Currently only compatible with PostgreSQL instances created using Unit CRDs from [github.com/upmio/unit-operator](https://github.com/upmio/unit-operator), as it requires access to Unit CRD start and stop service capabilities.
- **ProxysqlSync**: Synchronizes ProxySQL with backend MySQL topology
  
#### Controllers

Each CRD has a dedicated controller that:

- Monitors the custom resource state
- Connects to database instances using provided credentials
- Configures replication relationships
- Creates Kubernetes services for read/write splitting  
- Updates resource status with topology information
- Monitors replication health and provides status feedback for manual intervention

#### Database Utilities

Located in `pkg/`, these provide database-specific functionality:

- **mysqlutil**: MySQL client, replication management, group replication
- **redisutil**: Redis client, cluster operations, replication setup
- **postgresutil**: PostgreSQL client, streaming replication
- **proxysqlutil**: ProxySQL administration and configuration
- **k8sutil**: Kubernetes API operations

## Configuration

### Database User Requirements

#### MySQL

Create a user with necessary privileges:

```sql
-- Create replication user
CREATE USER 'replication_user'@'%' IDENTIFIED WITH mysql_native_password BY 'password';

-- Grant required privileges
GRANT REPLICATION CLIENT, REPLICATION SLAVE, SYSTEM_VARIABLES_ADMIN, 
      REPLICATION_SLAVE_ADMIN, GROUP_REPLICATION_ADMIN, RELOAD, 
      BACKUP_ADMIN, CLONE_ADMIN ON *.* TO 'replication_user'@'%';

GRANT SELECT ON performance_schema.* TO 'replication_user'@'%';

FLUSH PRIVILEGES;
```

#### Redis

For Redis with AUTH enabled:

```bash
# Set password in redis.conf
requirepass your_redis_password

# For Redis Cluster
masterauth your_redis_password
```

#### PostgreSQL

```sql
-- Create replication user
CREATE USER replication_user REPLICATION LOGIN ENCRYPTED PASSWORD 'password';

-- Configure pg_hba.conf
# TYPE  DATABASE        USER            ADDRESS                 METHOD
host    replication     replication_user    0.0.0.0/0             md5
```

### Advanced Configuration

#### Custom Resource Validation

The operator includes admission webhooks for validating custom resources:

```yaml
# Enable webhooks in values.yaml
webhooks:
  enabled: true
  certManager:
    enabled: true
```

#### Resource Limits

Configure resource limits for the operator:

```yaml
resources:
  limits:
    cpu: 200m
    memory: 256Mi
  requests:
    cpu: 100m
    memory: 128Mi
```

## Development

### Building from Source

```bash
# Clone the repository
git clone https://github.com/upmio/compose-operator.git
cd compose-operator

# Build the binary
make build

# Run tests
make test

# Build container image
make docker-build IMG=compose-operator:dev
```

### Local Development

```bash
# Install CRDs
make install

# Run locally (requires kubeconfig)
make run

# Generate code and manifests
make generate manifests
```

## Contributing

We welcome contributions! Please see [CONTRIBUTING.md](CONTRIBUTING.md) for details on:

- How to submit bug reports and feature requests
- Development workflow and coding standards
- Code review process
- Community guidelines

### Quick Contributing Guide

1. Fork the repository
2. Create a feature branch (`git checkout -b feature/amazing-feature`)
3. Commit your changes (`git commit -m 'Add amazing feature'`)
4. Push to the branch (`git push origin feature/amazing-feature`)
5. Open a Pull Request

## FAQ

### General Questions

**Q: Can the Compose Operator manage databases outside of Kubernetes?**
A: No, currently the operator only manages database instances running as pods within the same Kubernetes cluster. It labels database pods to create services with proper selectors. Support for external databases could be added in the future but is not currently implemented.

**Q: What happens if the operator is unavailable?**
A: Database operations continue normally. The operator only manages replication topology and creates services - it doesn't impact database availability.

**Q: Can I use this operator with existing database deployments?**
A: Yes, as long as the databases are running as pods within the same Kubernetes cluster. The operator works with existing database pods by establishing replication relationships and labeling them for service creation.

### Technical Questions

**Q: How are database passwords secured?**
A: Passwords are AES-256-CTR encrypted and stored in Kubernetes Secrets. The operator decrypts them at runtime using a configurable AES key to establish database connections.

**Q: What replication modes are supported for MySQL?**
A: The operator supports:

- `rpl_async`: Asynchronous replication
- `rpl_semi_sync`: Semi-synchronous replication
- `group_replication`: MySQL Group Replication

**Q: Can I have multiple replicas for high availability?**
A: Yes. You can configure multiple read replicas, and the operator will create separate services for read and write operations.

**Q: How does failover work?**
A: The operator monitors replication health and reports status in the custom resource. When issues are detected, administrators can observe the status and manually perform switchover/failover by updating the spec configuration. The operator then reconfigures replication and updates Kubernetes services accordingly.

**Q: Does the operator perform automatic failover?**
A: No, the operator does not perform automatic failover. It provides real-time monitoring and status reporting. Administrators must manually initiate switchover or failover by modifying the spec based on the observed status information.

### Troubleshooting

**Q: My replication setup failed. How do I debug?**
A: Check the operator logs and the status of your custom resource:

```bash
kubectl logs -n upm-system deployment/compose-operator-controller-manager
kubectl describe mysqlreplication your-resource-name
```

**Q: Database connections are failing. What should I check?**
A: Verify:

- Database credentials in the secret are correct and encrypted properly
- Network connectivity between operator and database instances
- Database user has necessary privileges
- Firewall rules allow connections on specified ports

## Security

### Password Encryption

All database passwords must be AES-256-CTR encrypted before storing in Kubernetes Secrets. Use binary files for better security:

```bash
# Build the encryption tool
go build -o aes-tool ./tool/

# Get the AES key used by the compose-operator
# Replace 'compose-operator' with your actual Helm release name
AES_KEY=$(kubectl get secret aes-secret-key -n upm-system -o jsonpath='{.data.AES_SECRET_KEY}' | base64 -d)

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
```

### Decrypting Passwords from Kubernetes Secrets

If you need to decrypt passwords that are stored in existing Kubernetes Secrets:

```bash
# Get the AES key used by the compose-operator
AES_KEY=$(kubectl get secret aes-secret-key -n upm-system -o jsonpath='{.data.AES_SECRET_KEY}' | base64 -d)

# Extract and decrypt a password from a secret
kubectl get secret mysql-credentials -o jsonpath='{.data.mysql}' | base64 -d > mysql.bin
aes-tool -key "$AES_KEY" -file "mysql.bin"
rm mysql.bin

# One-liner for quick password retrieval
kubectl get secret redis-credentials -o jsonpath='{.data.redis}' | base64 -d > temp.bin && aes-tool -key "$AES_KEY" -file "temp.bin" && rm temp.bin
```

### Network Security

- Use TLS connections to databases when possible
- Restrict network access using Kubernetes Network Policies
- Ensure database instances are not exposed to public networks unnecessarily

### RBAC

The operator uses minimal RBAC permissions:

- Manages only CRDs it owns
- Creates/updates services in the same namespace as custom resources
- Reads secrets containing database credentials

### Secret Management

- Store sensitive data only in Kubernetes Secrets
- Use separate secrets for different database instances

## License

This project is licensed under the Apache License 2.0 - see the [LICENSE](LICENSE) file for details.

## Support

- **Documentation**: [docs/](docs/)
- **Issues**: [GitHub Issues](https://github.com/upmio/compose-operator/issues)

## Roadmap

- [ ] Cross-cluster database management (manage databases across different Kubernetes clusters)
- [ ] External database support (manage databases outside Kubernetes clusters)

## Project Status

This project is actively maintained. We follow semantic versioning and maintain backward compatibility.
