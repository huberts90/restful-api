.PHONY: build run test clean lint fmt migrate-up migrate-down help docker-up docker-down mocks integration-test

# Application parameters
APP_NAME = restful-api
API_BIN = ./bin/api
MIGRATE_BIN = ./bin/migrate
VERSION ?= $(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")

# Build parameters
GO = go
GOOS ?= $(shell go env GOOS)
GOARCH ?= $(shell go env GOARCH)
BUILD_DIR = ./bin

# Test parameters
TEST_FLAGS = -v -race -coverprofile coverage.out

# Database parameters
MIGRATION_PATH = ./migrations
MIGRATION_DIRECTION ?= up
MIGRATION_STEPS ?= 0

# Help output
help:
	@echo "Management commands for $(APP_NAME):"
	@echo
	@echo "Usage:"
	@echo "    make build            Build the application binary"
	@echo "    make run              Run the API server locally"
	@echo "    make test             Run all tests"
	@echo "    make clean            Clean build artifacts"
	@echo "    make lint             Run linters"
	@echo "    make fmt              Run code formatters"
	@echo "    make migrate-up       Apply all up migrations"
	@echo "    make migrate-down     Revert all migrations"
	@echo "    make mocks            Regenerate mock implementations"
	@echo "    make integration-test Run integration tests"
	@echo

# Build the API server binary
build:
	@echo "Building API server binary..."
	@mkdir -p $(BUILD_DIR)
	$(GO) build -o $(API_BIN) -ldflags "-X main.version=$(VERSION)" ./cmd/api
	$(GO) build -o $(MIGRATE_BIN) ./cmd/migrate
	@echo "Build complete"

# Run the API server
run: build
	@echo "Starting API server..."
	$(API_BIN)

# Run all tests
test:
	@echo "Running tests..."
	$(GO) test $(TEST_FLAGS) ./...

# Clean build artifacts
clean:
	@echo "Cleaning build artifacts..."
	rm -rf $(BUILD_DIR)
	rm -f coverage.out
	@echo "Clean complete"

# Run linters
lint:
	@echo "Running linters..."
	golangci-lint run ./...

# Format code
fmt:
	@echo "Formatting code..."
	$(GO) fmt ./...
	goimports -w .

# Run database migrations (up)
migrate-up: build
	@echo "Running up migrations..."
	$(MIGRATE_BIN) -direction up -steps $(MIGRATION_STEPS)

# Run database migrations (down)
migrate-down: build
	@echo "Running down migrations..."
	$(MIGRATE_BIN) -direction down -steps $(MIGRATION_STEPS)


# Regenerate mocks
mocks:
	@echo "Regenerating mock implementations..."
	mockery --config=.mockery.yaml
	@echo "Mock generation complete"

# Run integration tests with docker-compose
integration-test:
	@echo "Running integration tests with docker-compose..."
	docker-compose up -d migrate
	@echo "Waiting for PostgreSQL to be ready..."
	@sleep 5
	$(GO) test $(TEST_FLAGS) ./internal/handler -run TestUserHandlerIntegrationSuite