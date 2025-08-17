# Multi-stage Dockerfile for Michishirube
# Supports multiarch builds (amd64, arm64) with optimized SQLite drivers

# Build stage
FROM --platform=$BUILDPLATFORM debian:bookworm-slim AS builder

# Build arguments for cross-compilation
ARG TARGETOS TARGETARCH

# Install build dependencies and Go 1.24.6
RUN apt-get update && apt-get install -y \
    git ca-certificates tzdata make wget tar \
    gcc libc6-dev libsqlite3-dev \
    gcc-aarch64-linux-gnu libc6-dev-arm64-cross \
    gcc-x86-64-linux-gnu libc6-dev-amd64-cross && \
    rm -rf /var/lib/apt/lists/*

# Install Go 1.24.6 manually
RUN GOARCH_MAP="amd64=amd64 arm64=arm64" && \
    GOARCH_TARGET=$(echo $GOARCH_MAP | grep -o "$TARGETARCH=[^[:space:]]*" | cut -d= -f2) && \
    wget -O go.tar.gz "https://go.dev/dl/go1.24.6.linux-${GOARCH_TARGET}.tar.gz" && \
    tar -C /usr/local -xzf go.tar.gz && \
    rm go.tar.gz

ENV PATH="/usr/local/go/bin:$PATH"
ENV GOPATH="/go"
ENV PATH="$GOPATH/bin:$PATH"

# Set working directory
WORKDIR /app

# Copy go mod files
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy source code
COPY . .

# Install build tools
RUN CGO_ENABLED=0 go install github.com/golang/mock/mockgen@latest
RUN CGO_ENABLED=0 go install github.com/swaggo/swag/cmd/swag@latest

# Generate code and documentation
RUN make generate
RUN make docs

# Build the application with architecture-specific optimizations
RUN if [ "$TARGETARCH" = "arm64" ]; then \
        export CC=aarch64-linux-gnu-gcc; \
    elif [ "$TARGETARCH" = "amd64" ]; then \
        export CC=x86_64-linux-gnu-gcc; \
    else \
        export CC=gcc; \
    fi && \
    CGO_ENABLED=1 GOOS=$TARGETOS GOARCH=$TARGETARCH \
    go build -ldflags="-w -s \
        -X main.version=${VERSION} \
        -X main.commit=$(git rev-parse --short HEAD 2>/dev/null || echo ${COMMIT}) \
        -X main.date=$(date -u +%Y-%m-%dT%H:%M:%SZ) \
        -X main.builtBy=${BUILT_BY}" \
    -tags sqlite_omit_load_extension \
    -o michishirube \
    ./cmd/server

# Final runtime stage
FROM debian:bookworm-slim

# Install runtime dependencies
RUN apt-get update && apt-get install -y ca-certificates tzdata sqlite3 wget && rm -rf /var/lib/apt/lists/*

# Create non-root user
RUN groupadd -g 1001 michishirube && \
    useradd -u 1001 -g michishirube -s /bin/bash -m michishirube

# Set working directory
WORKDIR /app

# Copy binary from builder
COPY --from=builder /app/michishirube .

# Copy web assets
COPY --from=builder /app/web ./web
COPY --from=builder /app/docs ./docs

# Create data directory with proper permissions
RUN mkdir -p /data && chown -R michishirube:michishirube /data /app

# Switch to non-root user
USER michishirube

# Expose port
EXPOSE 8080

# Health check
HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
    CMD wget --no-verbose --tries=1 --spider http://localhost:8080/health || exit 1

# Set default environment variables
# Application configuration
ENV PORT=8080
ENV DB_PATH=/data/michishirube.db
ENV LOG_LEVEL=info
ENV CONFIG_PATH=""

# Build-time variables that can be overridden
ARG VERSION=docker
ARG COMMIT=unknown
ARG BUILD_DATE=unknown
ARG BUILT_BY=docker

# Runtime environment info
ENV APP_VERSION=${VERSION}
ENV APP_COMMIT=${COMMIT}
ENV APP_BUILD_DATE=${BUILD_DATE}
ENV APP_BUILT_BY=${BUILT_BY}

# Run the application
CMD ["./michishirube"]