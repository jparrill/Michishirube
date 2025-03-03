# Running Michishirube in Docker

This document explains how to run Michishirube in Docker, which can be useful for consistent environments across different platforms.

## Prerequisites

- [Docker](https://docs.docker.com/get-docker/) installed on your system
- [Docker Compose](https://docs.docker.com/compose/install/) installed on your system
- For GUI applications:
  - On Linux: An X11 server (usually pre-installed)
  - On macOS: [XQuartz](https://www.xquartz.org/) installed

## Quick Start

The easiest way to run Michishirube in Docker is to use the Makefile commands:

```bash
# Build and run the application in Docker (auto-detects platform)
make docker-run
```

This command will:
1. Detect your platform (Linux or macOS)
2. Set up the necessary X11 forwarding
3. Build the Docker image if needed
4. Run the application in a Docker container

## Multi-Platform Support

Michishirube Docker images can be built for multiple platforms using Docker Buildx:

```bash
# Set up Docker Buildx
make docker-setup

# Build multi-platform images (requires Docker Hub or registry access)
make docker-build

# Build for local use only (current platform)
make docker-build-local
```

## Manual Setup

If you prefer to run commands manually:

### On Linux

```bash
# Allow X server connections
xhost +local:

# Build and run with docker-compose
docker-compose up --build

# When done
xhost -local:
```

### On macOS

```bash
# Start XQuartz if not running
open -a XQuartz

# In XQuartz preferences, enable "Allow connections from network clients"
# Restart XQuartz after changing settings

# Get your IP address
IP=$(ifconfig en0 | grep inet | awk '$1=="inet" {print $2}')

# Allow X server connections
xhost +$IP

# Run with docker
docker run -e DISPLAY=$IP:0 -v $(pwd)/data:/app/data michishirube:latest

# When done
xhost -$IP
```

## Data Persistence

Application data is stored in the `./data` directory, which is mounted as a volume in the container. This ensures that your data persists between container runs.

## Cross-Compilation

You can build native binaries for multiple platforms without Docker:

```bash
# Build binaries for Linux and macOS (both amd64 and arm64)
make cross-build
```

## Troubleshooting

### Display Issues

If you encounter display issues:

- On Linux, ensure the X11 socket is properly shared and that you've run `xhost +local:`
- On macOS, ensure XQuartz is running and configured to allow network connections

### Permission Issues

If you encounter permission issues with the mounted volumes:

```bash
# Fix permissions on the data directory
sudo chown -R $(id -u):$(id -g) ./data
```

### Accessing the Container Shell

To access a shell in a running container:

```bash
docker exec -it $(docker ps -q -f name=michishirube) /bin/bash
```