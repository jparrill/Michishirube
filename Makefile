# Michishirube Makefile

.PHONY: build run test lint clean docker-build docker-up fixtures-update fixtures-validate generate

# Build the application
build:
	go build -o michishirube ./cmd/server

# Run in development mode
run:
	go run ./cmd/server

# Generate code (mocks, etc.)
generate:
	@echo "Installing mockgen if needed..."
	@which mockgen > /dev/null || go install github.com/golang/mock/mockgen@latest
	@echo "Generating code..."
	go generate ./...
	@echo "Code generation completed"

# Run tests (generates code first)
test: generate
	go test ./...

# Run tests with coverage
test-coverage:
	go test -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out -o coverage.html

# Run linting (requires golangci-lint)
lint:
	golangci-lint run

# Clean build artifacts
clean:
	rm -f michishirube
	rm -f *.db
	rm -f coverage.out coverage.html

# Docker build
docker-build:
	docker build -t michishirube .

# Run with docker-compose
docker-up:
	docker-compose up --build

# Update fixtures with current test data
fixtures-update:
	@echo "Updating fixtures with current test data..."
	UPDATE=true go test ./internal/storage/sqlite/ -run "TestSQLiteStorage_WithFixtures"
	UPDATE=true go test ./internal/storage/sqlite/ -run "TestSQLiteStorage_WorkflowScenario"
	@echo "Fixtures updated successfully"

# Validate fixtures for reserved words
fixtures-validate:
	@echo "Validating fixtures for reserved words..."
	go test ./internal/storage/sqlite/ -run "TestSQLiteStorage_WithFixtures" -v
	go test ./internal/storage/sqlite/ -run "TestSQLiteStorage_SearchScenario" -v
	go test ./internal/storage/sqlite/ -run "TestSQLiteStorage_WorkflowScenario" -v
	@echo "Fixture validation completed"

# Run all storage tests
test-storage:
	go test ./internal/storage/sqlite/ -v

# Run search-specific tests
test-search:
	go test ./internal/storage/sqlite/ -run "Search" -v

# Development helper: run tests with fixtures update
dev-test:
	make fixtures-update && make test