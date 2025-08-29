#!/bin/bash

set -euo pipefail

# Set default DEBUG value if not provided
DEBUG=${DEBUG:-false}

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

# Validate required environment variables
validate_env() {
  local missing_vars=()
  
  [[ -z "${SECRET_NAME:-}" ]] && missing_vars+=("SECRET_NAME")
  [[ -z "${SECRET_KEY:-}" ]] && missing_vars+=("SECRET_KEY")
  [[ -z "${AES_KEY:-}" ]] && missing_vars+=("AES_KEY")
  
  if [[ ${#missing_vars[@]} -gt 0 ]]; then
    log_error "Missing required environment variables: ${missing_vars[*]}"
    exit 1
  fi
  
  # Validate AES key length
  if [[ ${#AES_KEY} -ne 32 ]]; then
    log_error "AES key must be exactly 32 characters long. Current length: ${#AES_KEY}"
    exit 1
  fi
}

# Check if secret exists
check_secret_exists() {
  local secret_name="$1"
  local namespace="${NAMESPACE:-default}"
  
  if ! kubectl get secret "$secret_name" -n "$namespace" &>/dev/null; then
    log_error "Secret '$secret_name' not found in namespace '$namespace'"
    exit 1
  fi
  
  log_info "Secret '$secret_name' found in namespace '$namespace'"
}

# Check if key exists in secret
check_key_exists() {
  local secret_name="$1"
  local secret_key="$2"
  local namespace="${NAMESPACE:-default}"
  
  if ! kubectl get secret "$secret_name" -n "$namespace" -o jsonpath="{.data.$secret_key}" &>/dev/null; then
    log_error "Key '$secret_key' not found in secret '$secret_name'"
    log_info "Available keys:"
    # Use yq to list keys to avoid jq dependency
    if kubectl get secret "$secret_name" -n "$namespace" -o yaml | yq '.data | keys | .[]' 2>/dev/null; then
      :
    else
      echo "Unable to list keys"
    fi
    exit 1
  fi
  
  log_info "Key '$secret_key' found in secret '$secret_name'"
}

# Decrypt value from secret
decrypt_secret_value() {
  local secret_name="$1"
  local secret_key="$2"
  local namespace="${NAMESPACE:-default}"
  
  # Get the base64 encoded encrypted value
  log_info "Retrieving encrypted value for key '$secret_key'..."
  local encrypted_base64
  encrypted_base64=$(kubectl get secret "$secret_name" -n "$namespace" -o jsonpath="{.data.$secret_key}")
  
  if [[ -z "$encrypted_base64" ]]; then
    log_error "Empty value for key '$secret_key' in secret '$secret_name'"
    exit 1
  fi
  
  # Decode base64 and save to temporary file
  local temp_file
  temp_file=$(mktemp)
  trap "rm -f $temp_file" EXIT
  
  echo "$encrypted_base64" | base64 -d > "$temp_file"
  
  # Decrypt the value
  log_info "Decrypting value..."
  local decrypted_value
  if decrypted_value=$(/usr/local/bin/aes-tool decrypt -k "$AES_KEY" -f "$temp_file" | sed -n 's/^Decrypted: //p' | tr -d '\r'); then
    log_info "Successfully decrypted value for key '$secret_key'"
    echo "$decrypted_value"
  else
    log_error "Failed to decrypt value for key '$secret_key'"
    exit 1
  fi
}

# Main execution
main() {
  log_info "Starting Secret decryption process..."
  
  # Validate environment variables
  validate_env
  
  local secret_name="$SECRET_NAME"
  local secret_key="$SECRET_KEY"
  local namespace="${NAMESPACE:-default}"
  
  log_info "Decrypting key '$secret_key' from Secret '$secret_name' in namespace '$namespace'"
  
  # Check if secret and key exist
  check_secret_exists "$secret_name"
  check_key_exists "$secret_name" "$secret_key"
  
  # Decrypt and output the value
  decrypt_secret_value "$secret_name" "$secret_key"
  
  log_info "Secret decryption completed successfully"
}

# Execute main function
main "$@"
