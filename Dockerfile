# Multi-stage Dockerfile for Michishirube
# Supports multiarch builds (amd64, arm64) with optimized SQLite drivers

# Build stage
FROM --platform=$BUILDPLATFORM golang:1.24.6-alpine AS builder

# Install build dependencies
RUN apk add --no-cache git ca-certificates tzdata gcc musl-dev sqlite-dev make

# Set working directory
WORKDIR /app

# Copy go mod files
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy source code
COPY . .

# Build arguments for cross-compilation
ARG TARGETOS TARGETARCH

# Install build tools
RUN go install github.com/golang/mock/mockgen@latest
RUN go install github.com/swaggo/swag/cmd/swag@latest

# Generate code and documentation
RUN make generate
RUN make docs

# Build the application with architecture-specific optimizations
RUN CGO_ENABLED=1 GOOS=$TARGETOS GOARCH=$TARGETARCH \
    go build -ldflags="-w -s \
        -X main.version=${VERSION} \
        -X main.commit=$(git rev-parse --short HEAD 2>/dev/null || echo ${COMMIT}) \
        -X main.date=$(date -u +%Y-%m-%dT%H:%M:%SZ) \
        -X main.builtBy=${BUILT_BY}" \
    -tags sqlite_omit_load_extension \
    -o michishirube \
    ./cmd/server

# Final runtime stage
FROM alpine:3.21

# Install runtime dependencies
RUN apk add --no-cache ca-certificates tzdata sqlite wget

# Create non-root user
RUN addgroup -g 1001 -S michishirube && \
    adduser -u 1001 -S michishirube -G michishirube

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