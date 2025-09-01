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
  [[ -z "${SECRET_KEYS:-}" ]] && missing_vars+=("SECRET_KEYS")
  [[ -z "${SECRET_VALUES:-}" ]] && missing_vars+=("SECRET_VALUES")
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

# Parse comma-separated values
parse_values() {
  local input="$1"
  echo "$input" | tr ',' '\n'
}

# Encrypt a single value
encrypt_value() {
  local value="$1"
  local temp_prefix="/tmp/temp$$"
  local temp_file="${temp_prefix}.bin"
  
  # Encrypt the value using aes-tool
  /usr/local/bin/aes-tool encrypt -k "$AES_KEY" -p "$value" -u "$temp_prefix"
  
  # Check if the encrypted file exists and encode to base64
  if [[ -f "$temp_file" ]]; then
    # Encode base64 without line wrapping; prefer -w 0 when available, fallback otherwise
    local base64_output
    if base64 --help 2>&1 | grep -q -- "-w"; then
      base64_output=$(base64 -w 0 < "$temp_file" | tr -d '\n\r')
    else
      base64_output=$(base64 < "$temp_file" | tr -d '\n\r')
    fi
    
    # Validate base64 output
    if [[ -n "$base64_output" ]] && [[ "$base64_output" =~ ^[A-Za-z0-9+/]*={0,2}$ ]]; then
      log_debug "Base64 encoded value length: ${#base64_output}"
      echo "$base64_output"
    else
      log_error "Invalid base64 output generated for value"
      rm -f "$temp_file"
      return 1
    fi
    
    rm -f "$temp_file"
  else
    log_error "Failed to create encrypted file: $temp_file"
    return 1
  fi
}

# Generate Secret YAML using yq
generate_secret_yaml() {
  local secret_name="$1"
  local namespace="${NAMESPACE:-default}"
  local labels="${SECRET_LABELS:-}"
  
  # Create base YAML structure using yq
  local yaml_content
  yaml_content=$(yq eval -n '
    .apiVersion = "v1" |
    .kind = "Secret" |
    .metadata.name = "'"$secret_name"'" |
    .metadata.namespace = "'"$namespace"'" |
    .type = "Opaque" |
    .data = {}
  ')
  
  # Add labels if provided
  if [[ -n "$labels" ]]; then
    IFS=',' read -ra LABEL_PAIRS <<< "$labels"
    for pair in "${LABEL_PAIRS[@]}"; do
      IFS='=' read -ra LABEL <<< "$pair"
      if [[ ${#LABEL[@]} -eq 2 ]]; then
        yaml_content=$(echo "$yaml_content" | yq eval '.metadata.labels."'"${LABEL[0]}"'" = "'"${LABEL[1]}"'"' -)
      fi
    done
  fi
  
  echo "$yaml_content"
}

# Validate YAML format using yq
validate_yaml() {
  local yaml_file="$1"
  
  if ! yq eval '.' "$yaml_file" > /dev/null 2>&1; then
    log_error "YAML validation failed: $yaml_file"
    log_error "Please check the YAML file's syntax and formatting"
    return 1
  fi
  
  log_debug "YAML validation passed: $yaml_file"
  return 0
}

# Main execution
main() {
  log_info "Starting Secret generation process..."
  
  # Validate environment variables
  validate_env
  
  # Parse keys and values
  readarray -t keys < <(parse_values "$SECRET_KEYS")
  readarray -t values < <(parse_values "$SECRET_VALUES")
  
  # Validate key-value pairs count
  if [[ ${#keys[@]} -ne ${#values[@]} ]]; then
    log_error "Number of keys (${#keys[@]}) does not match number of values (${#values[@]})"
    exit 1
  fi
  
  log_info "Processing ${#keys[@]} key-value pairs for Secret '$SECRET_NAME'"
  
  # Prepare temporary directory for encrypted files
  temp_dir="/tmp/secret-${SECRET_NAME}-$$"
  mkdir -p "$temp_dir"

  # Build --from-file args by encrypting each value into its own binary file
  declare -a from_file_args
  for i in "${!keys[@]}"; do
    key="${keys[i]}"
    value="${values[i]}"

    log_debug "Processing key: $key"

    # Encrypt value to file using aes-tool (output file: ${file_prefix}.bin)
    file_prefix="${temp_dir}/${key}"
    file_path="${file_prefix}.bin"
    /usr/local/bin/aes-tool encrypt -k "$AES_KEY" -p "$value" -u "$file_prefix"

    if [[ ! -f "$file_path" ]]; then
      log_error "Failed to create encrypted file for key: $key"
      rm -rf "$temp_dir"
      exit 1
    fi

    from_file_args+=("--from-file=${key}=${file_path}")
    log_info "Encrypted file prepared for key: $key"
  done

  # Create or update Secret using kubectl create --dry-run=client piped to apply
  log_info "Applying Secret to Kubernetes cluster..."
  namespace="${NAMESPACE:-default}"
  if kubectl create secret generic "$SECRET_NAME" -n "$namespace" "${from_file_args[@]}" --dry-run=client -o yaml | kubectl apply -f -; then
    log_info "Secret '$SECRET_NAME' created/updated successfully in namespace '${namespace}'"
  else
    log_error "Failed to create/update Secret '$SECRET_NAME'"
    rm -rf "$temp_dir"
    exit 1
  fi

  # Apply labels if provided
  if [[ -n "${SECRET_LABELS:-}" ]]; then
    IFS=',' read -ra LABEL_PAIRS <<< "$SECRET_LABELS"
    for pair in "${LABEL_PAIRS[@]}"; do
      if [[ -n "$pair" ]]; then
        kubectl label secret "$SECRET_NAME" -n "$namespace" "$pair" --overwrite 1>/dev/null || true
      fi
    done
  fi

  # Cleanup
  rm -rf "$temp_dir"

  log_info "Secret generation completed successfully"
}

# Execute main function
main "$@"
