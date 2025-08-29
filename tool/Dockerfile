# Build stage
FROM golang:1.24-alpine AS builder

# Install build dependencies
RUN apk add --no-cache git ca-certificates

# Set working directory
WORKDIR /workspace

# Copy repository root (we build from monorepo root module)
COPY . .

# Build the tool binary from root module
WORKDIR /workspace/tool
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o /workspace/bin/aes-tool .

# Runtime stage
FROM alpine:3.18

# Install runtime dependencies
RUN apk add --no-cache \
    ca-certificates \
    curl \
    bash

# Install kubectl
ARG TARGETARCH
RUN curl -LO "https://dl.k8s.io/release/$(curl -L -s https://dl.k8s.io/release/stable.txt)/bin/linux/${TARGETARCH}/kubectl" && \
    chmod +x kubectl && \
    mv kubectl /usr/local/bin/

# Install yq
RUN YQ_VERSION=$(curl -s https://api.github.com/repos/mikefarah/yq/releases/latest | grep '"tag_name":' | sed -E 's/.*"([^"]+)".*/\1/') && \
    curl -LO "https://github.com/mikefarah/yq/releases/download/${YQ_VERSION}/yq_linux_${TARGETARCH}" && \
    chmod +x yq_linux_${TARGETARCH} && \
    mv yq_linux_${TARGETARCH} /usr/local/bin/yq

# Create non-root user
RUN addgroup -g 1001 -S aes && \
    adduser -u 1001 -S aes -G aes

# Set working directory
WORKDIR /app

# Copy binary from builder stage
COPY --from=builder /workspace/bin/aes-tool /usr/local/bin/aes-tool

# Copy scripts (path relative to repository root)
COPY tool/k8s-scripts/ /app/

# Make scripts executable
RUN chmod +x /app/*.sh

# Switch to non-root user
USER aes

# Set entrypoint
ENTRYPOINT ["/app/entrypoint.sh"]