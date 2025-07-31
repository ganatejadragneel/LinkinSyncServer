# Makefile for LinkinSync Backend

.PHONY: help build run test test-unit test-integration test-coverage clean deps

# Default target
help:
	@echo "Available commands:"
	@echo "  build           - Build the application"
	@echo "  run             - Run the application"
	@echo "  test            - Run all tests"
	@echo "  test-unit       - Run unit tests only"
	@echo "  test-integration - Run integration tests only"
	@echo "  test-coverage   - Run tests with coverage report"
	@echo "  clean           - Clean build artifacts"
	@echo "  deps            - Download dependencies"
	@echo "  lint            - Run linter"
	@echo "  fmt             - Format code"

# Build the application
build:
	@echo "Building application..."
	go build -o bin/server ./server

# Run the application
run:
	@echo "Starting server..."
	go run ./server/main.go

# Run all tests
test:
	@echo "Running all tests..."
	go test -v ./tests/...

# Run unit tests only
test-unit:
	@echo "Running unit tests..."
	go test -v ./tests/unit/...

# Run integration tests only
test-integration:
	@echo "Running integration tests..."
	go test -v ./tests/integration/...

# Run tests with coverage
test-coverage:
	@echo "Running tests with coverage..."
	go test -v -coverprofile=coverage.out ./tests/...
	go tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report generated: coverage.html"

# Run tests for specific package
test-models:
	@echo "Running model tests..."
	go test -v ./tests/unit/models/...

test-handlers:
	@echo "Running handler tests..."
	go test -v ./tests/unit/handlers/...

test-repositories:
	@echo "Running repository tests..."
	go test -v ./tests/unit/repositories/...

test-services:
	@echo "Running service tests..."
	go test -v ./tests/unit/services/...

# Clean build artifacts
clean:
	@echo "Cleaning..."
	rm -rf bin/
	rm -f coverage.out coverage.html

# Download dependencies
deps:
	@echo "Downloading dependencies..."
	go mod download
	go mod tidy

# Run linter (requires golangci-lint)
lint:
	@echo "Running linter..."
	golangci-lint run

# Format code
fmt:
	@echo "Formatting code..."
	go fmt ./...

# Run security check (requires gosec)
security:
	@echo "Running security check..."
	gosec ./...

# Run all quality checks
quality: fmt lint security test

# Install development tools
install-tools:
	@echo "Installing development tools..."
	go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
	go install github.com/securecodewarrior/gosec/v2/cmd/gosec@latest

# Docker commands
docker-build:
	@echo "Building Docker image..."
	docker build -t linkin-sync-backend .

docker-run:
	@echo "Running Docker container..."
	docker run -p 8080:8080 --env-file .env linkin-sync-backend

# Database commands (if needed)
db-migrate:
	@echo "Running database migrations..."
	# Add your migration commands here

# Development server with auto-reload (requires air)
dev:
	@echo "Starting development server with auto-reload..."
	air

# Benchmark tests
benchmark:
	@echo "Running benchmark tests..."
	go test -bench=. -benchmem ./tests/...

# Profile the application
profile:
	@echo "Running with profiling..."
	go run ./server/main.go -cpuprofile=cpu.prof -memprofile=mem.prof