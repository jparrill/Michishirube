FROM golang:1.21-alpine AS builder

# Install build dependencies
RUN apk add --no-cache gcc musl-dev git

# Set working directory
WORKDIR /app

# Copy go mod and sum files
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy source code
COPY . .

# Build the application
RUN mkdir -p bin && \
    go build -ldflags="-s -w" -o bin/michishirube ./cmd

# Use a Debian-based image for better GUI support
FROM debian:bullseye-slim

# Build argument for loading fixtures
ARG LOAD_FIXTURES=false

# Install runtime dependencies for GUI applications
RUN apt-get update && apt-get install -y --no-install-recommends \
    ca-certificates \
    libx11-6 \
    libxcursor1 \
    libxrandr2 \
    libxinerama1 \
    libxi6 \
    libgl1 \
    libgtk-3-0 \
    xdg-utils \
    socat \
    && rm -rf /var/lib/apt/lists/*

# Set working directory
WORKDIR /app

# Copy the binary from the builder stage
COPY --from=builder /app/bin/michishirube /app/michishirube

# Create a directory for the database
RUN mkdir -p /app/data

# Set environment variables
ENV LOAD_FIXTURES=${LOAD_FIXTURES}

# Create platform detection and startup script
RUN echo '#!/bin/bash\n\
\n\
# Detect platform and set up X11 forwarding if needed\n\
if [ -n "$DISPLAY" ] && [[ "$DISPLAY" == *":"* ]]; then\n\
    # Standard X11 forwarding (Linux)\n\
    echo "Using standard X11 forwarding"\n\
    exec /app/michishirube\n\
elif [ -n "$DISPLAY" ] && [[ "$DISPLAY" == *"."* ]]; then\n\
    # macOS with XQuartz\n\
    echo "Using macOS X11 forwarding with XQuartz"\n\
    # Forward X11 via socat\n\
    socat TCP-LISTEN:6000,reuseaddr,fork UNIX-CLIENT:"$DISPLAY" &\n\
    export DISPLAY=:0\n\
    exec /app/michishirube\n\
else\n\
    # No X11 forwarding detected\n\
    echo "No X11 forwarding detected, running directly"\n\
    exec /app/michishirube\n\
fi\n\
' > /app/entrypoint.sh && chmod +x /app/entrypoint.sh

# Run the entrypoint script
ENTRYPOINT ["/app/entrypoint.sh"]