# Examples

This directory contains comprehensive examples for using the Compose Operator with different database types and configurations.

## Directory Structure

```
examples/
├── README.md                           # This file
├── mysql/                              # MySQL examples
│   ├── basic-replication.yaml          # Simple master-slave setup
│   ├── semi-sync-replication.yaml      # Semi-synchronous replication
│   ├── group-replication.yaml          # MySQL Group Replication
├── redis/                              # Redis examples
│   ├── basic-replication.yaml          # Simple source-replica setup
│   ├── cluster.yaml                    # Redis Cluster mode
├── postgres/                           # PostgreSQL examples
│   ├── streaming-replication.yaml      # Basic streaming replication
└── proxysql/                           # ProxySQL examples
    ├── sync-mysql-backend.yaml          # ProxySQL with MySQL backend
```

## Prerequisites

Before running any examples, ensure you have:

1. **Kubernetes cluster** with the Compose Operator installed
2. **Database instances** running and accessible from the cluster
3. **Database users** with appropriate replication privileges
4. **Encrypted passwords** stored in Kubernetes Secrets

### Password Encryption

All examples use encrypted passwords. Generate them using the AES encryption tool:

```bash
# Build the encryption tool (run from project root)
cd .. && go build -o aes-tool ./tool/ && cd examples

# Get the AES key used by the compose-operator
# Replace 'compose-operator' with your actual Helm release name and namespace
AES_KEY=$(kubectl get secret compose-operator-aes-secret -n upm-system -o jsonpath='{.data.AES_SECRET_KEY}' | base64 -d)

# Examples for different CRD types:

# For MySQL Replication & Group Replication
../aes-tool -key "$AES_KEY" -plaintext "mysql_root_password" -username "mysql"
../aes-tool -key "$AES_KEY" -plaintext "replication_password" -username "replication"
kubectl create secret generic mysql-credentials \
  --from-file=mysql=mysql.bin \
  --from-file=replication=replication.bin

# For Redis Replication & Cluster
../aes-tool -key "$AES_KEY" -plaintext "redis_auth_password" -username "redis"
kubectl create secret generic redis-credentials \
  --from-file=redis=redis.bin

# For PostgreSQL Replication
../aes-tool -key "$AES_KEY" -plaintext "postgresql_admin_password" -username "postgresql"
../aes-tool -key "$AES_KEY" -plaintext "replication_password" -username "replication"
kubectl create secret generic postgres-credentials \
  --from-file=postgresql=postgresql.bin \
  --from-file=replication=replication.bin

# For ProxySQL Sync
../aes-tool -key "$AES_KEY" -plaintext "proxysql_admin_password" -username "proxysql"
../aes-tool -key "$AES_KEY" -plaintext "mysql_backend_password" -username "mysql"
kubectl create secret generic proxysql-credentials \
  --from-file=proxysql=proxysql.bin \
  --from-file=mysql=mysql.bin

# Clean up binary files
rm *.bin
```

Note: The AES key used here must match the one configured during operator installation.

### Decrypting Passwords from Existing Secrets

If you need to decrypt passwords from secrets that have already been created:

```bash
# Get the AES key used by the compose-operator
AES_KEY=$(kubectl get secret compose-operator-aes-secret -n upm-system -o jsonpath='{.data.AES_SECRET_KEY}' | base64 -d)

# Extract and decrypt passwords from different secret types:

# MySQL credentials
kubectl get secret mysql-credentials -o jsonpath='{.data.mysql}' | base64 -d > mysql.bin
../aes-tool -key "$AES_KEY" -file "mysql.bin"

# Redis credentials  
kubectl get secret redis-credentials -o jsonpath='{.data.redis}' | base64 -d > redis.bin
../aes-tool -key "$AES_KEY" -file "redis.bin"

# PostgreSQL credentials
kubectl get secret postgres-credentials -o jsonpath='{.data.postgresql}' | base64 -d > postgresql.bin
../aes-tool -key "$AES_KEY" -file "postgresql.bin"

# ProxySQL credentials
kubectl get secret proxysql-credentials -o jsonpath='{.data.proxysql}' | base64 -d > proxysql.bin
../aes-tool -key "$AES_KEY" -file "proxysql.bin"

# Clean up
rm *.bin
```

## Quick Start

1. **Choose an example** that matches your setup
2. **Update the configuration** with your database hosts and credentials
3. **Create the secret** with encrypted passwords
4. **Apply the configuration** to your cluster

Example:
```bash
# Create secret
kubectl create secret generic mysql-credentials \
  --from-literal=mysql=<encrypted_password> \
  --from-literal=replication=<encrypted_replication_password>

# Apply configuration
kubectl apply -f examples/mysql/basic-replication.yaml

# Check status
kubectl get mysqlreplication -o wide
```

## Testing Your Setup

After applying any example, verify it's working:

```bash
# Check the custom resource status
kubectl describe <resource-type> <resource-name>

# Verify services were created
kubectl get services | grep <resource-name>

# Test database connectivity through services
kubectl run mysql-client --image=mysql:8.0 --rm -it -- mysql -h <service-name> -u root -p
```

## Production Considerations

When adapting these examples for production:

1. **Configure TLS/SSL** for database connections
2. **Set appropriate resource limits** on services
3. **Use separate namespaces** for different environments
4. **Implement monitoring** and alerting
5. **Plan for backup and disaster recovery**

## Contributing Examples

We welcome additional examples! When contributing:

1. Follow the existing naming convention
2. Include comprehensive comments
3. Test the example in a real environment  
4. Update this README with a description
5. Include any special prerequisites or setup steps