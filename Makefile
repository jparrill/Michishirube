# Michishirube Makefile

.PHONY: build run test test-unit test-integration test-coverage test-bench test-search test-help lint clean docker-build docker-up fixtures-update fixtures-validate generate docs dev-test

# Build the application
build:
	@mkdir -p build
	go build -o build/michishirube ./cmd/server

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

# Generate OpenAPI documentation
docs:
	@echo "Installing swag if needed..."
	@which swag > /dev/null || go install github.com/swaggo/swag/cmd/swag@latest
	@echo "Generating OpenAPI documentation..."
	@mkdir -p docs
	swag init -g cmd/server/main.go --output docs
	@echo "OpenAPI documentation generated in docs/ directory"
	@echo "Available endpoints:"
	@echo "  http://localhost:8080/docs - Custom Swagger UI"
	@echo "  http://localhost:8080/openapi.yaml - YAML specification"
	@echo "  http://localhost:8080/swagger/doc.json - JSON specification"

# Run complete test suite
test: generate fixtures-validate
	@echo "Running complete test suite..."
	@echo "================================"
	@echo "1. Unit Tests"
	@echo "================================"
	go test ./internal/models/ ./internal/config/ ./internal/logger/ -v
	@echo ""
	@echo "================================"
	@echo "2. Storage Tests (with fixtures)"
	@echo "================================"
	go test ./internal/storage/sqlite/ -v
	@echo ""
	@echo "================================"
	@echo "3. Handler Tests (with mocks)"
	@echo "================================"
	go test ./internal/handlers/ -v
	@echo ""
	@echo "================================"
	@echo "4. Integration Tests"
	@echo "================================"
	go test ./internal/ -v -run="TestIntegration"
	@echo ""
	@echo "================================"
	@echo "5. All Tests Summary"
	@echo "================================"
	go test ./... -v
	@echo ""
	@echo "✅ Complete test suite finished successfully!"

# Run tests with coverage
test-coverage: generate
	@echo "Running tests with coverage analysis..."
	@mkdir -p build
	go test -coverprofile=build/coverage.out ./...
	go tool cover -html=build/coverage.out -o build/coverage.html
	@echo "Coverage report generated: build/coverage.html"

# Run only unit tests (fast)
test-unit: generate
	@echo "Running unit tests only..."
	go test ./internal/models/ ./internal/config/ ./internal/logger/ ./internal/handlers/ -v

# Run only integration tests
test-integration: generate fixtures-validate
	@echo "Running integration tests only..."
	go test ./internal/ -v -run="TestIntegration"

# Run performance benchmarks
test-bench: generate
	@echo "Running performance benchmarks..."
	go test ./... -bench=. -run="^$$" -benchmem

# Run linting (requires golangci-lint)
lint:
	golangci-lint run

# Clean build artifacts
clean:
	rm -rf build/
	rm -rf docs/
	rm -f *.db
	rm -f config.yaml
	rm -f michishirube

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

# Show test help
test-help:
	@echo "Available test targets:"
	@echo "  make test              - Run complete test suite (recommended)"
	@echo "  make test-unit         - Run only unit tests (fast)"
	@echo "  make test-integration  - Run only integration tests"
	@echo "  make test-coverage     - Run tests with coverage report"
	@echo "  make test-bench        - Run performance benchmarks"
	@echo "  make fixtures-validate - Validate fixture data for issues"
	@echo "  make fixtures-update   - Update fixtures with current data"
	@echo "  make generate          - Generate mocks and code"
	@echo "  make docs              - Generate OpenAPI documentation"
	@echo "  make dev-test          - Update fixtures then run full suite"