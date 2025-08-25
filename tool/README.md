# AES Encryption Tool

This tool provides AES-256-CTR encryption and decryption functionality compatible with the Compose Operator.

## Building

Build the AES encryption tool from the project root directory:

```bash
# Build for current platform
go build -o aes-tool ./tool/

# Verify the build
aes-tool -help
```

This creates an `aes-tool` executable in the project root directory.

## Usage

### Encrypt password and save to binary file (Recommended for Kubernetes Secrets)

```bash
# Encrypt password for mysql user and save to mysql.bin
aes-tool -key "your-32-character-aes-key-here" -plaintext "mysql_password" -username "mysql"

# Encrypt password for replication user and save to replication.bin  
aes-tool -key "your-32-character-aes-key-here" -plaintext "repl_password" -username "replication"
```

Output:
```
Plaintext: mysql_password
Encrypted and saved to: mysql.bin
```

### Decrypt from binary file

```bash
# Decrypt password from binary file
aes-tool -key "your-32-character-aes-key-here" -file "mysql.bin"
```

Output:
```
File: mysql.bin
Decrypted: mysql_password
```

### Legacy base64 encryption/decryption

```bash
# Encrypt to base64 (for backward compatibility)
aes-tool -key "your-32-character-aes-key-here" -plaintext "mypassword"

# Decrypt from base64
aes-tool -key "your-32-character-aes-key-here" -decrypt "base64-encrypted-string"
```

### Show help

```bash
aes-tool -help
```

## Examples

### Complete workflow example with binary files (Recommended)

Here's a complete example of how to encrypt passwords for the compose-operator using binary files:

```bash
# Step 1: Build the AES tool
go build -o aes-tool ./tool/

# Step 2: Get the AES key from operator (replace with your values)
RELEASE_NAME="compose-operator"
NAMESPACE="upm-system"
AES_KEY=$(kubectl get secret ${RELEASE_NAME}-aes-secret -n ${NAMESPACE} -o jsonpath='{.data.AES_SECRET_KEY}' | base64 -d)

# Step 3: Verify key length (should be 32)
echo "AES key length: ${#AES_KEY}"

# Step 4: Encrypt passwords and save to binary files
aes-tool -key "$AES_KEY" -plaintext "mysql_root_password" -username "mysql"
aes-tool -key "$AES_KEY" -plaintext "replication_user_password" -username "replication"

# Step 5: Create Kubernetes secret from binary files
kubectl create secret generic mysql-credentials \
  -n ${NAMESPACE} \
  --from-file=mysql=mysql.bin \
  --from-file=replication=replication.bin

# Step 6: Verify (optional)
aes-tool -key "$AES_KEY" -file "mysql.bin"
aes-tool -key "$AES_KEY" -file "replication.bin"

# Step 7: Clean up binary files
rm mysql.bin replication.bin
```

### Step-by-step password encryption with binary files

```bash
# 1. Build the tool
go build -o aes-tool ./tool/

# 2. Get the AES key from operator secret
AES_KEY=$(kubectl get secret aes-secret-key -n upm-system -o jsonpath='{.data.AES_SECRET_KEY}' | base64 -d)

# 3. Encrypt individual passwords to binary files
aes-tool -key "$AES_KEY" -plaintext "mysql_root_password" -username "mysql"
aes-tool -key "$AES_KEY" -plaintext "replication_password" -username "replication"

# 4. Create Kubernetes secret from binary files
kubectl create secret generic mysql-credentials \
  --from-file=mysql=mysql.bin \
  --from-file=replication=replication.bin

# 5. Clean up
rm mysql.bin replication.bin
```

### Creating secrets manually (for users without the AES tool)

If you need to manually create a secret with encrypted passwords, follow these steps:

```bash
# Method 1: Using the AES tool to generate binary files
AES_KEY=$(kubectl get secret aes-secret-key -n upm-system -o jsonpath='{.data.AES_SECRET_KEY}' | base64 -d)
go run tool/main.go -key "$AES_KEY" -plaintext "mysql_password" -username "mysql"
go run tool/main.go -key "$AES_KEY" -plaintext "repl_password" -username "replication"

# Create secret from binary files
kubectl create secret generic ${SECRET_NAME} \
  -n ${NAMESPACE} \
  --from-file=mysql=mysql.bin \
  --from-file=replication=replication.bin

# Method 2: Direct kubectl creation with base64 encoded binary content
kubectl create secret generic ${SECRET_NAME} \
  -n ${NAMESPACE} \
  --from-literal=mysql="$(base64 -i mysql.bin)" \
  --from-literal=replication="$(base64 -i replication.bin)"
```

### Examples for Different CRD Types

#### MySQL Replication & Group Replication
```bash
AES_KEY=$(kubectl get secret aes-secret-key -n upm-system -o jsonpath='{.data.AES_SECRET_KEY}' | base64 -d)
aes-tool -key "$AES_KEY" -plaintext "mysql_root_password" -username "mysql"
aes-tool -key "$AES_KEY" -plaintext "replication_password" -username "replication"
kubectl create secret generic mysql-credentials \
  --from-file=mysql=mysql.bin \
  --from-file=replication=replication.bin
```

#### Redis Replication & Cluster
```bash
AES_KEY=$(kubectl get secret aes-secret-key -n upm-system -o jsonpath='{.data.AES_SECRET_KEY}' | base64 -d)
aes-tool -key "$AES_KEY" -plaintext "redis_auth_password" -username "redis"
kubectl create secret generic redis-credentials \
  --from-file=redis=redis.bin
```

#### PostgreSQL Replication
```bash
AES_KEY=$(kubectl get secret aes-secret-key -n upm-system -o jsonpath='{.data.AES_SECRET_KEY}' | base64 -d)
aes-tool -key "$AES_KEY" -plaintext "postgresql_admin_password" -username "postgresql"
aes-tool -key "$AES_KEY" -plaintext "replication_password" -username "replication"
kubectl create secret generic postgres-credentials \
  --from-file=postgresql=postgresql.bin \
  --from-file=replication=replication.bin
```

#### ProxySQL Sync
```bash
AES_KEY=$(kubectl get secret aes-secret-key -n upm-system -o jsonpath='{.data.AES_SECRET_KEY}' | base64 -d)
aes-tool -key "$AES_KEY" -plaintext "proxysql_admin_password" -username "proxysql"
aes-tool -key "$AES_KEY" -plaintext "mysql_backend_password" -username "mysql"
kubectl create secret generic proxysql-credentials \
  --from-file=proxysql=proxysql.bin \
  --from-file=mysql=mysql.bin
```

### Decrypting Passwords from Kubernetes Secrets

If you need to decrypt passwords stored in Kubernetes Secrets that were created using this tool:

```bash
# Get the AES key used by the compose-operator
AES_KEY=$(kubectl get secret aes-secret-key -n upm-system -o jsonpath='{.data.AES_SECRET_KEY}' | base64 -d)

# Method 1: Extract and decrypt a specific key from a secret
kubectl get secret mysql-credentials -o jsonpath='{.data.mysql}' | base64 -d > mysql.bin
aes-tool -key "$AES_KEY" -file "mysql.bin"
rm mysql.bin

# Method 2: One-liner for quick password retrieval
kubectl get secret mysql-credentials -o jsonpath='{.data.mysql}' | base64 -d > temp.bin && aes-tool -key "$AES_KEY" -file "temp.bin" && rm temp.bin

# Method 3: Using JSON output for multiple keys
kubectl get secret mysql-credentials -o json | jq '.data.mysql' -r | base64 -d > mysql.bin
kubectl get secret mysql-credentials -o json | jq '.data.replication' -r | base64 -d > replication.bin
aes-tool -key "$AES_KEY" -file "mysql.bin"
aes-tool -key "$AES_KEY" -file "replication.bin"
rm mysql.bin replication.bin
```

### Examples for Different Secret Types

#### Decrypt MySQL credentials
```bash
AES_KEY=$(kubectl get secret aes-secret-key -n upm-system -o jsonpath='{.data.AES_SECRET_KEY}' | base64 -d)
kubectl get secret mysql-credentials -o jsonpath='{.data.mysql}' | base64 -d > mysql.bin
kubectl get secret mysql-credentials -o jsonpath='{.data.replication}' | base64 -d > replication.bin
echo "MySQL password:"
aes-tool -key "$AES_KEY" -file "mysql.bin"
echo "Replication password:"
aes-tool -key "$AES_KEY" -file "replication.bin"
rm mysql.bin replication.bin
```

#### Decrypt Redis credentials
```bash
AES_KEY=$(kubectl get secret aes-secret-key -n upm-system -o jsonpath='{.data.AES_SECRET_KEY}' | base64 -d)
kubectl get secret redis-credentials -o jsonpath='{.data.redis}' | base64 -d > redis.bin
echo "Redis password:"
aes-tool -key "$AES_KEY" -file "redis.bin"
rm redis.bin
```

#### Decrypt PostgreSQL credentials
```bash
AES_KEY=$(kubectl get secret aes-secret-key -n upm-system -o jsonpath='{.data.AES_SECRET_KEY}' | base64 -d)
kubectl get secret postgres-credentials -o jsonpath='{.data.postgresql}' | base64 -d > postgresql.bin
kubectl get secret postgres-credentials -o jsonpath='{.data.replication}' | base64 -d > replication.bin
echo "PostgreSQL password:"
aes-tool -key "$AES_KEY" -file "postgresql.bin"
echo "Replication password:"
aes-tool -key "$AES_KEY" -file "replication.bin"
rm postgresql.bin replication.bin
```

#### Decrypt ProxySQL credentials
```bash
AES_KEY=$(kubectl get secret aes-secret-key -n upm-system -o jsonpath='{.data.AES_SECRET_KEY}' | base64 -d)
kubectl get secret proxysql-credentials -o jsonpath='{.data.proxysql}' | base64 -d > proxysql.bin
kubectl get secret proxysql-credentials -o jsonpath='{.data.mysql}' | base64 -d > mysql.bin
echo "ProxySQL admin password:"
aes-tool -key "$AES_KEY" -file "proxysql.bin"
echo "MySQL backend password:"
aes-tool -key "$AES_KEY" -file "mysql.bin"
rm proxysql.bin mysql.bin
```

### Verify encryption/decryption

```bash
# Get the AES key used by the compose-operator
AES_KEY=$(kubectl get secret aes-secret-key -n upm-system -o jsonpath='{.data.AES_SECRET_KEY}' | base64 -d)

# Test complete workflow: encrypt -> create secret -> extract -> decrypt
aes-tool -key "$AES_KEY" -plaintext "test_password" -username "test"
kubectl create secret generic test-secret --from-file=test=test.bin --dry-run=client -o json | jq '.data.test' -r | base64 -d > extracted-test.bin
aes-tool -key "$AES_KEY" -file "extracted-test.bin"
rm test.bin extracted-test.bin
```

## Key Requirements

- The AES key must be exactly 32 characters long
- The same key used for encryption must be used for decryption
- **IMPORTANT**: Always use the actual AES key from the compose-operator deployment, not a custom key

### Getting the Operator's AES Key

The compose-operator uses an AES key stored in a Kubernetes secret. To encrypt passwords that the operator can decrypt, you must use the same key:

#### Getting the AES key with kubectl

```bash
# Get the AES key from the operator's secret
# Replace 'compose-operator' with your Helm release name
# Replace 'upm-system' with your installation namespace
AES_KEY=$(kubectl get secret aes-secret-key -n upm-system -o jsonpath='{.data.AES_SECRET_KEY}' | base64 -d)

# Verify the key length (should be 32 characters)
echo "AES key length: ${#AES_KEY}"
```

### Troubleshooting

#### Common Issues

**1. Secret not found**
```bash
# List secrets to find the correct name
kubectl get secrets -n upm-system | grep aes

# Check if the operator is running
kubectl get pods -n upm-system
```

**2. Wrong secret name or namespace**
```bash
# The secret name format is: {release-name}-aes-secret
# For Helm release "my-operator" in namespace "my-ns":
kubectl get secret my-operator-aes-secret -n my-ns -o jsonpath='{.data.AES_SECRET_KEY}' | base64 -d
```

**3. Decryption errors in operator logs**
- Verify you're using the correct secret name: `{release-name}-aes-secret`
- Verify you're using the correct namespace where the operator is installed
- Ensure the key length is exactly 32 characters
- Make sure you're using the `AES_SECRET_KEY` field from the secret

**4. Key validation**
```bash
# Verify the key length (should be 32)
AES_KEY=$(kubectl get secret aes-secret-key -n upm-system -o jsonpath='{.data.AES_SECRET_KEY}' | base64 -d)
echo "Key length: ${#AES_KEY}"
if [ ${#AES_KEY} -eq 32 ]; then echo "✅ Key length is correct"; else echo "❌ Key length is wrong"; fi
```

## Compatibility

This tool uses the same AES-CTR encryption implementation as the Compose Operator's `pkg/utils` package, ensuring full compatibility for password encryption and decryption.