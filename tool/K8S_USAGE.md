# Kubernetes AES Encryption Tool

A containerized tool for encrypting and decrypting Kubernetes Secrets using AES-256-CTR encryption, designed to work seamlessly with the Compose Operator.

## Features

- **Secure AES-256-CTR Encryption**: Industry-standard encryption for sensitive data
- **Kubernetes Native**: Designed specifically for Kubernetes Secret management
- **Multi-Architecture Support**: Available for AMD64 and ARM64 platforms
- **RBAC Compliant**: Follows Kubernetes security best practices
- **Compose Operator Compatible**: Uses the same encryption standards as the Compose Operator
- **Flexible Operations**: Support for both generation and decryption of encrypted Secrets
- **Container Security**: Runs as non-root user with minimal privileges

## Quick Start

### 1. Build Multi-Architecture Image

```bash
docker buildx build --platform linux/amd64,linux/arm64 \
  -t quay.io/upmio/k8s-aes-tool:latest \
  --push .
```

### 2. Create AES Key Secret

```bash
# Generate a 32-character AES key
AES_KEY=$(openssl rand -base64 32 | head -c 32)

# Create the secret (use field name AES_SECRET_KEY)
kubectl create secret generic aes-secret-key \
  --from-literal=AES_SECRET_KEY="$AES_KEY" \
  -n default
```

### 3. Apply RBAC Configuration

```bash
kubectl apply -f k8s-manifests/rbac.yaml
```

### 4. Generate Encrypted Secret

```bash
kubectl apply -f k8s-manifests/generate-secret-job.yaml
```

### 5. Decrypt Secret Value

```bash
kubectl apply -f k8s-manifests/decrypt-secret-job.yaml
```

## Usage Examples

### Generate InnoDB Cluster Credentials

Use the Job manifest to generate an encrypted Secret that contains multiple database user passwords:

```yaml
apiVersion: batch/v1
kind: Job
metadata:
  name: generate-innodb-cluster-secret-job
  namespace: default
spec:
  template:
    spec:
      serviceAccountName: aes-tool-sa
      restartPolicy: Never
      containers:
      - name: aes-tool
        image: quay.io/upmio/k8s-aes-tool:latest
        imagePullPolicy: Always
        args: ["generate"]
        env:
        - name: SECRET_NAME
          value: "innodb-cluster-sg-demo-secret"
        - name: SECRET_KEYS
          value: "helix,monitor,radminuser,replication,root"
        - name: SECRET_VALUES
          value: "helix_password,monitor_password,radmin_password,replication_password,root_password"
        - name: NAMESPACE
          value: "default"
        - name: AES_KEY_PATH
          value: "/etc/aes-key/key"
        - name: SECRET_LABELS
          value: "upm.api/service-group.name=demo,upm.api/service-group.type=innodb-cluster-sg,upm.io/owner=upm"
        volumeMounts:
        - name: aes-key
          mountPath: /etc/aes-key
          readOnly: true
      volumes:
      - name: aes-key
        secret:
          secretName: aes-secret-key
          items:
          - key: AES_SECRET_KEY
            path: key
  backoffLimit: 3
```

### Decrypt Specific Password Fields

```yaml
apiVersion: batch/v1
kind: Job
metadata:
  name: decrypt-innodb-cluster-secret-job
  namespace: default
spec:
  template:
    spec:
      serviceAccountName: aes-tool-sa
      restartPolicy: Never
      containers:
      - name: aes-tool
        image: quay.io/upmio/k8s-aes-tool:latest
        imagePullPolicy: Always
        args: ["decrypt"]
        env:
        - name: SECRET_NAME
          value: "innodb-cluster-sg-demo-secret"
        - name: SECRET_KEY
          value: "root"  # Decrypt root user password
        - name: NAMESPACE
          value: "default"
        - name: AES_KEY_PATH
          value: "/etc/aes-key/key"
        volumeMounts:
        - name: aes-key
          mountPath: /etc/aes-key
          readOnly: true
      volumes:
      - name: aes-key
        secret:
          secretName: aes-secret-key
          items:
          - key: AES_SECRET_KEY
            path: key
  backoffLimit: 3
```

## Environment Variables Reference

### Common Variables

| Variable | Description | Default | Required |
|----------|-------------|---------|----------|
| `AES_KEY_PATH` | Path to AES key file | `/etc/aes-key/key` | No |
| `NAMESPACE` | Kubernetes namespace | `default` | No |
| `DEBUG` | Enable debug logging | `false` | No |

### Generate Mode Variables

| Variable | Description | Example | Required |
|----------|-------------|---------|----------|
| `SECRET_NAME` | Name of the secret to create | `innodb-cluster-sg-demo-secret` | Yes |
| `SECRET_KEYS` | Comma-separated key names | `helix,monitor,radminuser,replication,root` | Yes |
| `SECRET_VALUES` | Comma-separated values | `password1,password2,password3,password4,password5` | Yes |
| `SECRET_LABELS` | Secret labels (optional) | `key1=value1,key2=value2` | No |

### Decrypt Mode Variables

| Variable | Description | Example | Required |
|----------|-------------|---------|----------|
| `SECRET_NAME` | Name of the secret to decrypt | `innodb-cluster-sg-demo-secret` | Yes |
| `SECRET_KEY` | Specific key to decrypt | `root` | Yes |

### Direct Mode Variables

| Variable | Description | Example | Required |
|----------|-------------|---------|----------|
| `PLAINTEXT` | Text to encrypt (encrypt mode) | `my secret` | Yes* |
| `ENCRYPTED_FILE` | File to decrypt (decrypt-file mode) | `/path/to/file.bin` | Yes* |

*Required only for respective modes

## Security Features

### Container Security

- **Non-root execution**: Runs as user ID 1001
- **Minimal attack surface**: Based on minimal Alpine Linux
- **Read-only root filesystem**: Prevents runtime modifications
- **Resource limits**: Configurable CPU and memory constraints

### Kubernetes Security

- **RBAC enforcement**: Minimal required permissions
- **Secret mounting**: AES keys never exposed in environment
- **Namespace isolation**: Operates within specified namespace
- **Audit logging**: All operations logged for compliance

### Encryption Security

- **AES-256-CTR**: Industry-standard encryption algorithm
- **Key validation**: Ensures proper key length and format
- **Secure key storage**: Keys stored in Kubernetes Secrets
- **No key exposure**: Keys never logged or exposed

## Troubleshooting

### Common Issues

#### Job fails with "AES key file not found"

```bash
kubectl get secret aes-secret-key -n default
kubectl get secret aes-secret-key -o yaml
```

#### Job fails with "Cannot connect to Kubernetes cluster"

```bash
kubectl auth can-i create secrets --as=system:serviceaccount:default:aes-tool-sa
kubectl get serviceaccount aes-tool-sa
```

#### Job fails with "Secret not found"

```bash
kubectl get secrets -n <namespace>
kubectl get secrets -l encryption.upmio.com/encrypted=true
```

### Debug Mode

Enable debug logging:

```yaml
env:
- name: DEBUG
  value: "true"
```

### Logs Analysis

```bash
kubectl logs job/<job-name>
kubectl logs -l job-name=<job-name>
kubectl describe job <job-name>
```

## Advanced Usage

### Batch Secret Generation

```bash
#!/bin/bash
SECRETS=(
  "db-credentials:username,password:admin,secret123"
  "api-keys:token,refresh:abc123,def456"
  "cache-config:host,port:redis.local,6379"
)

for secret_config in "${SECRETS[@]}"; do
  IFS=':' read -r name keys values <<< "$secret_config"
  
  kubectl create job "generate-$name" --image=quay.io/upmio/k8s-aes-tool:latest \
    --env="MODE=generate" \
    --env="SECRET_NAME=$name" \
    --env="SECRET_KEYS=$keys" \
    --env="SECRET_VALUES=$values"
done
```

### Cross-Namespace Operations

```yaml
env:
- name: NAMESPACE
  value: "production"
```

### Custom Resource Limits

```yaml
resources:
  requests:
    memory: "64Mi"
    cpu: "100m"
  limits:
    memory: "128Mi"
    cpu: "200m"
```

## Best Practices

1. **Key rotation**: Rotate AES encryption keys regularly
2. **Namespace isolation**: Use separate keys per namespace
3. **Monitoring**: Monitor job execution and failures
4. **Backup**: Back up encrypted secrets and keys securely
5. **Access control**: Limit RBAC permissions to the minimum required
6. **Audit**: Enable audit logging for secret operations

## Configuration Files

### RBAC Configuration (rbac.yaml)

This tool ships with a complete RBAC configuration including:

- ServiceAccount: `aes-tool-sa`
- Role: Permissions for get, list, create, update, patch on Secrets
- RoleBinding: Binds the role to the service account

### Job Manifests

- `generate-secret-job.yaml`: Job template for generating encrypted Secrets
- `decrypt-secret-job.yaml`: Job template for decrypting existing Secrets

## Support

If you run into issues:

1. Check the troubleshooting section
2. Review job logs and events
3. Verify RBAC and network connectivity
4. Ensure the AES key is properly configured

## License

This tool is part of the compose-operator project and follows the same license terms.
