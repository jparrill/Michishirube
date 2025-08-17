# Michishirube Makefile

.PHONY: build run test test-unit test-integration test-coverage test-bench test-search test-help lint clean docker-build docker-up docker-multiarch docker-dev docker-down docker-logs docker-help fixtures-update fixtures-validate generate docs dev-test release release-check release-snapshot ci-local ci-help deps deps-update deps-clean deps-verify deps-help security security-gosec security-govulncheck security-install security-help security-ci security-strict

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

# Docker build (single arch)
docker-build:
	@VERSION=$$(git describe --tags --exact-match 2>/dev/null || echo "latest"); \
	docker build \
		--build-arg VERSION=$$VERSION \
		--build-arg COMMIT=$$(git rev-parse --short HEAD 2>/dev/null || echo "unknown") \
		--build-arg BUILD_DATE=$$(date -u +%Y-%m-%dT%H:%M:%SZ) \
		--build-arg BUILT_BY=make-local \
		-t quay.io/jparrill/michishirube:$$VERSION \
		-t quay.io/jparrill/michishirube:latest \
		-t michishirube:latest \
		.

# Docker multiarch build with buildx
docker-multiarch:
	@echo "Setting up Docker buildx for multiarch builds..."
	docker buildx create --name michishirube-builder --use || docker buildx use michishirube-builder
	@echo "Building for multiple architectures..."
	@VERSION=$$(git describe --tags --exact-match 2>/dev/null || echo "latest"); \
	docker buildx build \
		--platform linux/amd64,linux/arm64 \
		--build-arg VERSION=$$VERSION \
		--build-arg COMMIT=$$(git rev-parse --short HEAD 2>/dev/null || echo "unknown") \
		--build-arg BUILD_DATE=$$(date -u +%Y-%m-%dT%H:%M:%SZ) \
		--build-arg BUILT_BY=make-multiarch \
		-t quay.io/jparrill/michishirube:$$VERSION \
		-t quay.io/jparrill/michishirube:latest \
		--push \
		.

# Run with docker-compose (production)
docker-up:
	docker-compose up --build

# Run with docker-compose (development mode)
docker-dev:
	docker-compose --profile dev up --build

# Stop docker-compose services
docker-down:
	docker-compose down

# View docker-compose logs
docker-logs:
	docker-compose logs -f

# Clean Docker resources
docker-clean:
	docker-compose down -v
	docker system prune -f
	docker buildx rm michishirube-builder || true

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

# Release targets
release:
	@echo "Installing GoReleaser if needed..."
	@which goreleaser > /dev/null || go install github.com/goreleaser/goreleaser@latest
	@echo "Creating release..."
	goreleaser release --clean

# Check release configuration
release-check:
	@echo "Installing GoReleaser if needed..."
	@which goreleaser > /dev/null || go install github.com/goreleaser/goreleaser@latest
	@echo "Checking GoReleaser configuration..."
	goreleaser check

# Create snapshot release (for development)
release-snapshot:
	@echo "Installing GoReleaser if needed..."
	@which goreleaser > /dev/null || go install github.com/goreleaser/goreleaser@latest
	@echo "Creating snapshot release..."
	goreleaser release --snapshot --clean

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

# Show release help
release-help:
	@echo "Available release targets:"
	@echo "  make release           - Create and publish a release (requires git tag)"
	@echo "  make release-check     - Check GoReleaser configuration"
	@echo "  make release-snapshot  - Create snapshot build for development"
	@echo ""
	@echo "Release workflow:"
	@echo "  1. git tag v1.0.0      - Create a version tag"
	@echo "  2. make release-check  - Validate the configuration"
	@echo "  3. make release-snapshot - Test with a snapshot build"
	@echo "  4. make release        - Create and publish the release"

# Show Docker help
docker-help:
	@echo "Available Docker targets:"
	@echo "  make docker-build      - Build single architecture Docker image (quay.io/jparrill/michishirube)"
	@echo "  make docker-multiarch  - Build multiarch image with buildx (amd64, arm64) and push to registry"
	@echo "  make docker-up         - Run production environment with docker-compose"
	@echo "  make docker-dev        - Run development environment with hot reload"
	@echo "  make docker-down       - Stop all docker-compose services"
	@echo "  make docker-logs       - View logs from all services"
	@echo "  make docker-clean      - Clean all Docker resources (containers, volumes, buildx)"
	@echo ""
	@echo "Environment variables (create .env file):"
	@echo "  PORT=8080              - Application port"
	@echo "  LOG_LEVEL=info         - Log level (debug, info, warn, error)"
	@echo "  DB_PATH=./app.db       - Database file path"
	@echo "  DEV_PORT=8081          - Development mode port"
	@echo ""
	@echo "Docker workflow:"
	@echo "  1. cp .env.example .env           - Create environment config"
	@echo "  2. make docker-build              - Build the image locally"
	@echo "  3. make docker-up                 - Start production services"
	@echo "  4. make docker-dev                - Start development services"
	@echo "  5. make docker-multiarch          - Build and push multiarch image to quay.io"
	@echo ""
	@echo "Registry: quay.io/jparrill/michishirube:latest (or version tag)"

# Simulate CI/CD pipeline locally
ci-local:
	@echo "🚀 Running CI/CD pipeline simulation locally..."
	@echo "================================================"
	@echo "This simulates the same steps as GitHub Actions CI pipeline"
	@echo ""
	@echo "Step 1/8: Managing Go dependencies..."
	@make deps
	@echo "✅ Dependencies managed"
	@echo ""
	@echo "Step 2/8: Generating code and documentation..."
	@make generate
	@make docs
	@echo "✅ Code generation completed"
	@echo ""
	@echo "Step 3/8: Running linter..."
	@make lint
	@echo "✅ Linting passed"
	@echo ""
	@echo "Step 4/8: Running complete test suite..."
	@make test
	@echo "✅ All tests passed"
	@echo ""
	@echo "Step 5/8: Running tests with coverage analysis..."
	@make test-coverage
	@echo "✅ Coverage analysis completed"
	@echo ""
	@echo "Step 6/8: Building application..."
	@make build
	@echo "✅ Application built successfully"
	@echo ""
	@echo "Step 7/8: Testing binary execution..."
	@./build/michishirube --version
	@timeout 10s ./build/michishirube > /dev/null 2>&1 || [ $$? -eq 124 ]
	@echo "✅ Binary execution verified"
	@echo ""
	@echo "Step 8/8: Checking GoReleaser configuration..."
	@make release-check
	@echo "✅ GoReleaser configuration valid"
	@echo ""
	@echo "🎉 CI/CD pipeline simulation completed successfully!"
	@echo "All steps that run in GitHub Actions have been verified locally."
	@echo ""
	@echo "Coverage report available at: build/coverage.html"

# Show CI help
ci-help:
	@echo "CI/CD simulation targets:"
	@echo "  make ci-local          - Run complete CI/CD pipeline simulation locally"
	@echo "  make ci-help           - Show this help message"
	@echo ""
	@echo "The ci-local target simulates all steps from GitHub Actions:"
	@echo "  1. Manage Go dependencies (make deps)"
	@echo "  2. Generate code and docs (make generate && make docs)"
	@echo "  3. Run linter (make lint)"
	@echo "  4. Run complete test suite (make test)"
	@echo "  5. Run coverage analysis (make test-coverage)"
	@echo "  6. Build application (make build)"
	@echo "  7. Test binary execution (version + startup test)"
	@echo "  8. Validate GoReleaser config (make release-check)"
	@echo ""
	@echo "This helps catch issues before pushing to GitHub and triggering CI."

# Dependency management
deps:
	@echo "📦 Managing Go dependencies..."
	@echo "========================================="
	@echo "Downloading and verifying dependencies..."
	@go mod download
	@go mod verify
	@echo "✅ Dependencies downloaded and verified"
	@echo ""
	@echo "Tidying module files..."
	@go mod tidy
	@echo "✅ Module files tidied"
	@echo ""
	@echo "📋 Current dependency summary:"
	@go list -m -mod=readonly all | wc -l | sed 's/^/  Total modules: /'
	@echo "  Direct dependencies:"
	@go list -m -f '{{if not .Indirect}}  - {{.Path}} {{.Version}}{{end}}' all | grep -v "^  -$$"
	@echo ""
	@echo "🎉 Dependency management completed!"

# Update all dependencies to latest versions
deps-update:
	@echo "🔄 Updating all dependencies to latest versions..."
	@echo "=============================================="
	@echo "Getting latest versions for direct dependencies..."
	@go get -u ./...
	@echo "✅ Dependencies updated"
	@echo ""
	@echo "Tidying module files..."
	@go mod tidy
	@echo "✅ Module files tidied"
	@echo ""
	@echo "Verifying updated dependencies..."
	@go mod verify
	@echo "✅ Dependencies verified"
	@echo ""
	@echo "📋 Updated dependency summary:"
	@go list -m -mod=readonly all | wc -l | sed 's/^/  Total modules: /'
	@echo ""
	@echo "⚠️  IMPORTANT: Review changes and run tests before committing!"
	@echo "   Run: make test"

# Clean dependency cache and reinstall
deps-clean:
	@echo "🧹 Cleaning dependency cache..."
	@echo "================================="
	@echo "Cleaning module cache..."
	@go clean -modcache
	@echo "✅ Module cache cleaned"
	@echo ""
	@echo "Re-downloading dependencies..."
	@go mod download
	@echo "✅ Dependencies re-downloaded"
	@echo ""
	@echo "Verifying dependencies..."
	@go mod verify
	@echo "✅ Dependencies verified"
	@echo ""
	@echo "🎉 Dependency cache cleaned and rebuilt!"

# Verify dependency integrity and security
deps-verify:
	@echo "🔍 Verifying dependency integrity and security..."
	@echo "================================================"
	@echo "Verifying module integrity..."
	@go mod verify
	@echo "✅ Module integrity verified"
	@echo ""
	@echo "Checking for known vulnerabilities..."
	@if command -v govulncheck >/dev/null 2>&1; then \
		govulncheck ./...; \
		echo "✅ Vulnerability check completed"; \
	else \
		echo "⚠️  govulncheck not installed. Install with:"; \
		echo "   go install golang.org/x/vuln/cmd/govulncheck@latest"; \
	fi
	@echo ""
	@echo "Analyzing dependency graph..."
	@go mod graph | head -20
	@echo "  ... (showing first 20 dependencies)"
	@echo ""
	@echo "📊 Dependency statistics:"
	@echo "  Direct dependencies: $$(go list -m -f '{{if not .Indirect}}{{.Path}}{{end}}' all | grep -v '^$$' | wc -l | tr -d ' ')"
	@echo "  Indirect dependencies: $$(go list -m -f '{{if .Indirect}}{{.Path}}{{end}}' all | grep -v '^$$' | wc -l | tr -d ' ')"
	@echo "  Total modules: $$(go list -m all | wc -l | tr -d ' ')"

# Show dependency management help
deps-help:
	@echo "Dependency management targets:"
	@echo "  make deps              - Download, verify, and tidy dependencies"
	@echo "  make deps-update       - Update all dependencies to latest versions"
	@echo "  make deps-clean        - Clean cache and reinstall dependencies"
	@echo "  make deps-verify       - Verify integrity and check for vulnerabilities"
	@echo "  make deps-help         - Show this help message"
	@echo ""
	@echo "Common workflows:"
	@echo "  Initial setup:         make deps"
	@echo "  Regular maintenance:   make deps-update && make test"
	@echo "  Troubleshoot issues:   make deps-clean"
	@echo "  Security audit:        make deps-verify"
	@echo ""
	@echo "Additional tools (install separately):"
	@echo "  govulncheck:          go install golang.org/x/vuln/cmd/govulncheck@latest"
	@echo "  go mod outdated:      go install github.com/psampaz/go-mod-outdated@latest"
	@echo ""
	@echo "Tips:"
	@echo "  - Always run tests after updating dependencies"
	@echo "  - Review go.mod changes before committing"
	@echo "  - Use deps-verify regularly for security"

# Security scanning
security: security-install security-gosec security-govulncheck
	@echo ""
	@echo "🎉 Security scanning completed!"
	@echo "Review any findings above and address security issues before deployment."

# Install security scanning tools
security-install:
	@echo "🔧 Installing security scanning tools..."
	@echo "======================================="
	@echo "Installing gosec..."
	@if ! command -v gosec >/dev/null 2>&1; then \
		go install github.com/securego/gosec/v2/cmd/gosec@latest; \
		echo "✅ gosec installed"; \
	else \
		echo "✅ gosec already installed"; \
	fi
	@echo ""
	@echo "Installing govulncheck..."
	@if ! command -v govulncheck >/dev/null 2>&1; then \
		go install golang.org/x/vuln/cmd/govulncheck@latest; \
		echo "✅ govulncheck installed"; \
	else \
		echo "✅ govulncheck already installed"; \
	fi
	@echo ""
	@echo "🎉 Security tools installation completed!"

# Run gosec security scanner
security-gosec:
	@echo "🔍 Running gosec security scanner..."
	@echo "===================================="
	@if command -v gosec >/dev/null 2>&1; then \
		gosec ./...; \
		echo "✅ gosec scan completed"; \
	else \
		echo "❌ gosec not found. Run 'make security-install' first"; \
		exit 1; \
	fi

# Run security scan for CI (strict mode - fails on HIGH severity)
security-ci: security-install
	@echo "🔍 Running CI security scan (strict mode)..."
	@echo "============================================="
	@echo "This will fail the build if HIGH severity issues are found"
	@echo ""
	@echo "Running gosec with HIGH severity filter..."
	@if command -v gosec >/dev/null 2>&1; then \
		echo "Running gosec scan for HIGH severity issues..."; \
		if gosec -severity high -confidence medium -quiet ./... >/dev/null 2>&1; then \
			echo "✅ gosec HIGH severity scan passed"; \
		else \
			echo "❌ gosec found HIGH severity security issues!"; \
			gosec -severity high -confidence medium ./...; \
			exit 1; \
		fi; \
	else \
		echo "❌ gosec not found"; \
		exit 1; \
	fi
	@echo ""
	@echo "Running govulncheck..."
	@if command -v govulncheck >/dev/null 2>&1; then \
		govulncheck ./...; \
		echo "✅ govulncheck scan passed"; \
	else \
		echo "❌ govulncheck not found"; \
		exit 1; \
	fi
	@echo ""
	@echo "🎉 CI security scan passed! No HIGH severity issues found."

# Run security scan in strict mode (fails on MEDIUM+ severity)
security-strict: security-install
	@echo "🔍 Running security scan (strict mode)..."
	@echo "=========================================="
	@echo "This will fail the build if MEDIUM+ severity issues are found"
	@echo ""
	@echo "Running gosec with MEDIUM+ severity filter..."
	@if command -v gosec >/dev/null 2>&1; then \
		echo "Running gosec scan for MEDIUM+ severity issues..."; \
		if gosec -severity medium -confidence medium -quiet ./... >/dev/null 2>&1; then \
			echo "✅ gosec MEDIUM+ severity scan passed"; \
		else \
			echo "❌ gosec found MEDIUM+ severity security issues!"; \
			gosec -severity medium -confidence medium ./...; \
			exit 1; \
		fi; \
	else \
		echo "❌ gosec not found"; \
		exit 1; \
	fi
	@echo ""
	@echo "Running govulncheck..."
	@if command -v govulncheck >/dev/null 2>&1; then \
		govulncheck ./...; \
		echo "✅ govulncheck scan passed"; \
	else \
		echo "❌ govulncheck not found"; \
		exit 1; \
	fi
	@echo ""
	@echo "🎉 Strict security scan passed! No MEDIUM+ severity issues found."

# Run govulncheck vulnerability scanner
security-govulncheck:
	@echo ""
	@echo "🔍 Running govulncheck vulnerability scanner..."
	@echo "==============================================="
	@if command -v govulncheck >/dev/null 2>&1; then \
		govulncheck ./...; \
		echo "✅ govulncheck scan completed"; \
	else \
		echo "❌ govulncheck not found. Run 'make security-install' first"; \
		exit 1; \
	fi

# Show security help
security-help:
	@echo "Security scanning targets:"
	@echo "  make security              - Run complete security scan (gosec + govulncheck)"
	@echo "  make security-ci           - Run CI security scan (fails on HIGH severity)"
	@echo "  make security-strict       - Run strict security scan (fails on MEDIUM+ severity)"
	@echo "  make security-install      - Install security scanning tools"
	@echo "  make security-gosec        - Run gosec static analysis scanner"
	@echo "  make security-govulncheck  - Run govulncheck vulnerability scanner"
	@echo "  make security-help         - Show this help message"
	@echo ""
	@echo "Security tools overview:"
	@echo "  gosec:         Static analysis for Go security issues"
	@echo "                 - Detects hardcoded credentials, SQL injection, etc."
	@echo "                 - Analyzes Go AST for security vulnerabilities"
	@echo ""
	@echo "  govulncheck:   Vulnerability database checker"
	@echo "                 - Checks dependencies against Go vulnerability DB"
	@echo "                 - Reports known CVEs in used packages"
	@echo ""
	@echo "Common workflows:"
	@echo "  Initial setup:     make security-install"
	@echo "  Regular scanning:  make security"
	@echo "  CI/CD integration: make security-ci (fails on HIGH severity)"
	@echo "  Strict checking:   make security-strict (fails on MEDIUM+ severity)"
	@echo ""
	@echo "Tips:"
	@echo "  - Run security scans before each release"
	@echo "  - Address HIGH and MEDIUM severity issues"
	@echo "  - Use #nosec comments sparingly and with justification"
	@echo "  - Keep dependencies updated to avoid vulnerabilities"