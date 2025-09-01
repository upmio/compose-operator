#!/bin/bash

# Logging functions
log_info() {
  echo "[INFO] $(date '+%Y-%m-%d %H:%M:%S') - $1"
}

log_error() {
  echo "[ERROR] $(date '+%Y-%m-%d %H:%M:%S') - $1" >&2
}

log_debug() {
  if [[ "$DEBUG" == "true" ]]; then
    echo "[DEBUG] $(date '+%Y-%m-%d %H:%M:%S') - $1"
  fi
}

# Function to validate and load AES key
load_aes_key() {
  if [[ -z "$AES_KEY_PATH" ]]; then
    log_error "AES_KEY_PATH environment variable is not set"
    exit 1
  fi

  if [[ ! -f "$AES_KEY_PATH" ]]; then
    log_error "AES key file not found at: $AES_KEY_PATH"
    exit 1
  fi

  AES_KEY=$(cat "$AES_KEY_PATH")
  if [[ ${#AES_KEY} -ne 32 ]]; then
    log_error "AES key must be exactly 32 characters long. Current length: ${#AES_KEY}"
    exit 1
  fi

  log_info "AES key loaded successfully from $AES_KEY_PATH"
  export AES_KEY
}

# Function to check required permissions against Secrets only
check_kubectl() {
  local namespace
  namespace="${NAMESPACE:-default}"

  # Determine required verbs based on COMMAND
  local required_verbs=()
  case "${COMMAND:-}" in
    "generate")
      # create + patch are sufficient for apply/label flows
      required_verbs=(create patch)
      ;;
    "decrypt")
      required_verbs=(get)
      ;;
    *)
      required_verbs=(get)
      ;;
  esac

  # Check RBAC for each required verb on secrets in the target namespace
  for verb in "${required_verbs[@]}"; do
    local can
    can=$(kubectl auth can-i "$verb" secrets -n "$namespace" 2>/dev/null || true)
    if [[ "$can" != "yes" ]]; then
      log_error "Insufficient RBAC: cannot $verb secrets in namespace '$namespace' for current ServiceAccount"
      exit 1
    fi
  done

  log_info "RBAC check passed for secrets in namespace '$namespace'"
}

# Function to show help
show_help() {
  cat << EOF
Kubernetes AES Encryption Tool

Usage: $0 <command> [options]

Commands:
  generate    Generate encrypted Kubernetes Secret
  decrypt     Decrypt value from Kubernetes Secret
  encrypt     Encrypt a value using AES (standalone)
  decrypt-file Decrypt from file using AES (standalone)
  help        Show this help message

Environment Variables:
  AES_KEY_PATH    Path to AES key file (required)
  DEBUG          Enable debug logging (optional)

For Kubernetes operations:
  SECRET_NAME     Name of the Secret (required for generate/decrypt)
  SECRET_KEYS     Comma-separated keys (required for generate)
  SECRET_VALUES   Comma-separated values (required for generate)
  SECRET_KEY      Single key to decrypt (required for decrypt)
  NAMESPACE       Kubernetes namespace (default: default)
  SECRET_LABELS   Labels for Secret (optional, format: key1=value1,key2=value2)

Examples:
  # Generate encrypted Secret
  export SECRET_NAME="my-secret"
  export SECRET_KEYS="username,password"
  export SECRET_VALUES="admin,secret123"
  $0 generate

  # Decrypt Secret value
  export SECRET_NAME="my-secret"
  export SECRET_KEY="password"
  $0 decrypt

  # Standalone encryption
  echo "mysecret" | $0 encrypt

  # Standalone decryption from file
  $0 decrypt-file /path/to/encrypted/file
EOF
}

# Main execution
COMMAND="${1:-}"

case "$COMMAND" in
  "generate")
    log_info "Starting Secret generation process..."
    load_aes_key
    check_kubectl
    exec /app/generate-k8s-secret.sh
    ;;
  "decrypt")
    log_info "Starting Secret decryption process..."
    load_aes_key
    check_kubectl
    exec /app/decrypt-k8s-secret.sh
    ;;
  "encrypt")
    log_info "Starting standalone encryption..."
    load_aes_key
    # Allow plaintext via env or stdin; username is required for output file naming unless stdout
    if [[ -n "${PLAINTEXT:-}" ]]; then
      exec /usr/local/bin/aes-tool encrypt -k "$AES_KEY" -p "$PLAINTEXT" -u "${USERNAME:-stdout}" ${STDOUT:+--stdout}
    else
      # Read from stdin to PLAINTEXT variable and encrypt
      if [[ -t 0 ]]; then
        log_error "No PLAINTEXT provided and no stdin. Set PLAINTEXT or pipe input."
        exit 1
      fi
      read -r -d '' PLAINTEXT_STDIN || true
      if [[ -z "$PLAINTEXT_STDIN" ]]; then
        PLAINTEXT_STDIN=$(cat -)
      fi
      # Output base64 to keep output printable
      encrypted_bin=$(/usr/local/bin/aes-tool encrypt -k "$AES_KEY" -p "$PLAINTEXT_STDIN" --stdout | base64 -w 0)
      echo "$encrypted_bin"
      exit 0
    fi
    ;;
  "decrypt-file")
    log_info "Starting standalone decryption from file..."
    load_aes_key
    # File path can come from arg2 or ENCRYPTED_FILE
    target_file="${2:-${ENCRYPTED_FILE:-}}"
    if [[ -z "$target_file" ]]; then
      log_error "Missing encrypted file path. Provide as arg or set ENCRYPTED_FILE."
      exit 1
    fi
    exec /usr/local/bin/aes-tool decrypt -k "$AES_KEY" -f "$target_file"
    ;;
  "help"|"--help"|"-h")
    show_help
    ;;
  "")
    log_error "No command specified"
    show_help
    exit 1
    ;;
  *)
    log_error "Unknown command: $COMMAND"
    show_help
    exit 1
    ;;
esac
