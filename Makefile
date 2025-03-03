# Michishirube Makefile

# Variables
BINARY_NAME=michishirube
BUILD_DIR=bin
MAIN_PATH=./cmd
GO=go
GOFLAGS=-v
LDFLAGS=-ldflags="-s -w"
DOCKER_IMAGE=michishirube
DOCKER_TAG=latest
PLATFORMS=linux/amd64,linux/arm64,darwin/amd64,darwin/arm64

# Default target
.PHONY: all
all: clean build

# Build the application
.PHONY: build
build:
	@echo "Building $(BINARY_NAME)..."
	@mkdir -p $(BUILD_DIR)
	$(GO) build $(GOFLAGS) $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME) $(MAIN_PATH)
	@echo "Build complete: $(BUILD_DIR)/$(BINARY_NAME)"

# Build for development with debug info
.PHONY: build-dev
build-dev:
	@echo "Building $(BINARY_NAME) (development)..."
	@mkdir -p $(BUILD_DIR)
	$(GO) build $(GOFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME) $(MAIN_PATH)
	@echo "Development build complete: $(BUILD_DIR)/$(BINARY_NAME)"

# Run the application
.PHONY: run
run: build
	@echo "Running $(BINARY_NAME)..."
	$(BUILD_DIR)/$(BINARY_NAME)

# Run the application with fixtures
.PHONY: run-with-fixtures
run-with-fixtures: build
	@echo "Running $(BINARY_NAME) with fixtures..."
	LOAD_FIXTURES=true $(BUILD_DIR)/$(BINARY_NAME)

# Clean build artifacts
.PHONY: clean
clean:
	@echo "Cleaning..."
	@rm -f $(BINARY_NAME)
	@rm -rf $(BUILD_DIR)
	@echo "Clean complete"

# Run tests
.PHONY: test
test:
	@echo "Running tests..."
	$(GO) test -v ./...

# Run tests with coverage
.PHONY: test-coverage
test-coverage:
	@echo "Running tests with coverage..."
	$(GO) test -v -coverprofile=coverage.out ./...
	$(GO) tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report generated: coverage.html"

# Format code
.PHONY: fmt
fmt:
	@echo "Formatting code..."
	$(GO) fmt ./...

# Lint code
.PHONY: lint
lint:
	@echo "Linting code..."
	@if command -v golangci-lint > /dev/null; then \
		golangci-lint run ./...; \
	else \
		echo "golangci-lint not installed. Run: go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest"; \
		exit 1; \
	fi

# Install dependencies
.PHONY: deps
deps:
	@echo "Installing dependencies..."
	$(GO) mod download
	$(GO) mod tidy

# Reset database
.PHONY: reset-db
reset-db:
	@echo "Resetting database..."
	@rm -f *.db
	@rm -f *.db-journal
	@echo "Database reset complete"

# Cross-compile for multiple platforms
.PHONY: cross-build
cross-build:
	@echo "Cross-compiling for multiple platforms..."
	@mkdir -p $(BUILD_DIR)
	GOOS=linux GOARCH=amd64 $(GO) build $(GOFLAGS) $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)-linux-amd64 $(MAIN_PATH)
	GOOS=linux GOARCH=arm64 $(GO) build $(GOFLAGS) $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)-linux-arm64 $(MAIN_PATH)
	GOOS=darwin GOARCH=amd64 $(GO) build $(GOFLAGS) $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)-darwin-amd64 $(MAIN_PATH)
	GOOS=darwin GOARCH=arm64 $(GO) build $(GOFLAGS) $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)-darwin-arm64 $(MAIN_PATH)
	@echo "Cross-compilation complete"

# Docker targets
# Setup Docker buildx
.PHONY: docker-setup
docker-setup:
	@echo "Setting up Docker buildx..."
	@if ! docker buildx ls | grep -q michishirube-builder; then \
		docker buildx create --name michishirube-builder --use; \
	else \
		docker buildx use michishirube-builder; \
	fi
	@docker buildx inspect --bootstrap

# Build multi-platform Docker image
.PHONY: docker-build
docker-build: docker-setup
	@echo "Building multi-platform Docker image $(DOCKER_IMAGE):$(DOCKER_TAG)..."
	docker buildx build --platform $(PLATFORMS) \
		-t $(DOCKER_IMAGE):$(DOCKER_TAG) \
		--build-arg LOAD_FIXTURES=true \
		--push .

# Build and load Docker image for local use
.PHONY: docker-build-local
docker-build-local:
	@echo "Building Docker image for local use $(DOCKER_IMAGE):$(DOCKER_TAG)..."
	docker build -t $(DOCKER_IMAGE):$(DOCKER_TAG) \
		--build-arg LOAD_FIXTURES=true \
		.

# Run Docker container with X11 forwarding (Linux)
.PHONY: docker-run-linux
docker-run-linux: docker-build-local
	@echo "Running Docker container with X11 forwarding (Linux)..."
	@# Check if Docker is installed
	@if ! command -v docker > /dev/null; then \
		echo "Docker is not installed. Please install Docker first."; \
		exit 1; \
	fi
	@# Allow X server connections from local users
	@xhost +local: || { echo "Failed to set xhost permissions. Is X11 running?"; exit 1; }
	@# Create data directory if it doesn't exist
	@mkdir -p data
	@# Run the container
	docker-compose up --build
	@# Disallow X server connections when done
	@xhost -local:

# Run Docker container with X11 forwarding (macOS)
.PHONY: docker-run-mac
docker-run-mac: docker-build-local
	@echo "Running Docker container with X11 forwarding (macOS)..."
	@# Check if Docker is installed
	@if ! command -v docker > /dev/null; then \
		echo "Docker is not installed. Please install Docker first."; \
		exit 1; \
	fi
	@# Check if XQuartz is installed
	@if ! ls /Applications/Utilities/XQuartz.app > /dev/null 2>&1; then \
		echo "XQuartz is not installed. Please install XQuartz first."; \
		echo "You can install it with: brew install --cask xquartz"; \
		exit 1; \
	fi
	@# Start XQuartz if not running
	@if ! ps -e | grep -q XQuartz; then \
		echo "Starting XQuartz..."; \
		open -a XQuartz; \
		sleep 5; \
	fi
	@# Configure XQuartz to allow connections from network clients
	@defaults write org.macosforge.xquartz.X11 nolisten_tcp 0
	@# Restart XQuartz to apply settings
	@killall Xquartz 2>/dev/null || true
	@open -a XQuartz
	@# Wait for XQuartz to start
	@sleep 5
	@# Get IP address
	@IP=$$(ifconfig en0 | grep inet | awk '$$1=="inet" {print $$2}'); \
	if [ -z "$$IP" ]; then \
		IP=$$(ifconfig en1 | grep inet | awk '$$1=="inet" {print $$2}'); \
	fi; \
	if [ -z "$$IP" ]; then \
		echo "Could not determine IP address. Please check your network connection."; \
		exit 1; \
	fi; \
	echo "Using IP: $$IP"; \
	# Allow connections from your IP \
	xhost + $$IP; \
	# Create data directory if it doesn't exist \
	mkdir -p data; \
	# Run the container \
	docker run -it --rm \
		-e DISPLAY=$$IP:0 \
		-v ./data:/app/data \
		$(DOCKER_IMAGE):$(DOCKER_TAG); \
	# Disallow connections when done \
	xhost - $$IP

# Run Docker container (auto-detect platform)
.PHONY: docker-run
docker-run:
	@echo "Detecting platform..."
	@if [ "$$(uname)" = "Darwin" ]; then \
		echo "macOS detected, running with XQuartz..."; \
		$(MAKE) docker-run-mac; \
	else \
		echo "Linux detected, running with X11..."; \
		$(MAKE) docker-run-linux; \
	fi

# Docker Compose targets
.PHONY: docker-compose-up
docker-compose-up:
	@echo "Starting services with docker-compose..."
	docker-compose up --build

.PHONY: docker-compose-down
docker-compose-down:
	@echo "Stopping services with docker-compose..."
	docker-compose down

# Help
.PHONY: help
help:
	@echo "Michishirube Makefile targets:"
	@echo "  all             - Clean and build the application"
	@echo "  build           - Build the application"
	@echo "  build-dev       - Build the application with debug info"
	@echo "  run             - Build and run the application"
	@echo "  run-with-fixtures - Build and run with sample data"
	@echo "  clean           - Remove build artifacts"
	@echo "  test            - Run tests"
	@echo "  test-coverage   - Run tests with coverage report"
	@echo "  fmt             - Format code"
	@echo "  lint            - Lint code"
	@echo "  deps            - Install dependencies"
	@echo "  reset-db        - Reset the database"
	@echo "  cross-build     - Cross-compile for multiple platforms"
	@echo "  docker-setup    - Set up Docker buildx for multi-platform builds"
	@echo "  docker-build    - Build multi-platform Docker image"
	@echo "  docker-build-local - Build Docker image for local use"
	@echo "  docker-run      - Run application in Docker (auto-detect platform)"
	@echo "  docker-run-linux - Run application in Docker with X11 forwarding (Linux)"
	@echo "  docker-run-mac  - Run application in Docker with XQuartz (macOS)"
	@echo "  docker-compose-up - Start services with docker-compose"
	@echo "  docker-compose-down - Stop services with docker-compose"
	@echo "  help            - Show this help message"