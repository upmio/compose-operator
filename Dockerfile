# Build the compose-operator binary
FROM golang:1.25.1 AS builder

ARG TARGETOS
ARG TARGETARCH

ENV GO111MODULE on

WORKDIR /workspace
# Copy the Go Modules manifests
COPY go.mod go.mod
COPY go.sum go.sum
# cache deps before building and copying source so that we don't need to re-download as much
# and so that source changes don't invalidate our downloaded layer
RUN go mod download

# Copy the go source
COPY cmd/main.go cmd/main.go
COPY api/ api/
COPY pkg/  pkg/
COPY controller/ controller/
COPY webhook/ webhook/
COPY .git .git

ENV GOPROXY 'https://goproxy.cn,direct'

# Build
# the GOARCH has not a default value to allow the binary be built according to the host where the command
# was called. For example, if we call make docker-build in a local env which has the Apple Silicon M1 SO
# the docker BUILDPLATFORM arg will be linux/arm64 when for Apple x86 it will be linux/amd64. Therefore,
# by leaving it empty we can ensure that the container and binary shipped on it will have the same platform.
RUN CGO_ENABLED=0 GOOS=${TARGETOS:-linux} GOARCH=${TARGETARCH} go build -a -o dist/compose-operator \
    -ldflags "-s -w"  \
    -ldflags "-X  'github.com/upmio/compose-operator/pkg/version.GIT_BRANCH=$(git rev-parse --abbrev-ref HEAD)'  -X 'github.com/upmio/compose-operator/pkg/version.GIT_COMMIT=$(git rev-parse HEAD)'  -X 'github.com/upmio/compose-operator/pkg/version.BUILD_TIME=$(date '+%Y-%m-%d %H:%M:%S')'  -X 'github.com/upmio/compose-operator/pkg/version.GIT_BRANCH=$(git rev-parse --abbrev-ref HEAD)'  -X 'github.com/upmio/compose-operator/pkg/version.GO_VERSION=$(go version | grep -o  'go[0-9].[0-9].*')'" \
    cmd/main.go

# Use distroless as minimal base image to package the compose-operator binary
# Refer to https://github.com/GoogleContainerTools/distroless for more details
FROM rockylinux:9.3.20231119

# set timezone
ENV TZ=Asia/Shanghai
RUN ln -snf /usr/share/zoneinfo/${TZ} /etc/localtime \
  && echo "$TZ" > /etc/timezone \
  && dnf install -y \
         procps-ng \
         net-tools \
         telnet \
         epel-release

WORKDIR /
COPY --from=builder /workspace/dist/compose-operator /usr/local/bin/compose-operator
USER 65532:65532

ENTRYPOINT ["/usr/local/bin/compose-operator","--leader-elect","true"]
