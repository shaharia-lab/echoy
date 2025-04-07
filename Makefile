# Variables
GO := go
GOTEST := $(GO) test
GOVET := $(GO) vet
GOFMT := gofmt
GOLINT := golangci-lint
GORELEASE := goreleaser

# Binary output
BINARY_NAME := echoy
BUILD_DIR := build

# Go module paths
PKG := $(shell $(GO) list -m)
PKGS := $(shell $(GO) list ./...)

# Build variables
LDFLAGS := -ldflags "-s -w"
VERSION ?= $(shell git describe --tags --always --dirty)
COMMIT_HASH ?= $(shell git rev-parse --short HEAD 2>/dev/null)
BUILD_TIME ?= $(shell date +%FT%T%z)

# Set default goal
.DEFAULT_GOAL := build

# Phony targets
.PHONY: all build clean test test-unit test-coverage lint fmt vet vendor-deps gorelease-test help

# Build the application
build:
	@echo "Building $(BINARY_NAME)..."
	$(GO) build $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME) .

# Clean build artifacts
clean:
	@echo "Cleaning up..."
	rm -rf $(BUILD_DIR)/
	rm -f coverage.out

# Run all tests
test:
	@echo "Running tests..."
	$(GOTEST) -v ./...

# Run unit tests only
test-unit:
	@echo "Running unit tests..."
	$(GOTEST) -v -short ./...

# Run tests with coverage
test-coverage:
	@echo "Running tests with coverage..."
	$(GOTEST) -v -race -coverprofile=coverage.out -covermode=atomic ./...
	$(GO) tool cover -html=coverage.out -o coverage.html

# Run linter
lint:
	@echo "Running linter..."
	$(GOLINT) run

# Format code
fmt:
	@echo "Formatting code..."
	$(GOFMT) -s -w .

# Run go vet
vet:
	@echo "Running go vet..."
	$(GOVET) ./...

# Install dependencies
deps:
	@echo "Installing dependencies..."
	$(GO) mod download

# Install development tools
tools:
	@echo "Installing development tools..."
	$(GO) install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
	$(GO) install github.com/goreleaser/goreleaser/v2@latest

# Test goreleaser configuration
gorelease-test:
	@echo "Testing goreleaser configuration..."
	$(GORELEASE) check
	$(GORELEASE) release --skip=publish --clean --fail-fast --verbose

# All-in-one target for CI
ci: deps lint test-unit vet gorelease-test

# Help command
help:
	@echo "Makefile targets:"
	@echo "  build          - Build the application"
	@echo "  clean          - Remove build artifacts"
	@echo "  test           - Run all tests"
	@echo "  test-unit      - Run unit tests only"
	@echo "  test-coverage  - Run tests with coverage"
	@echo "  lint           - Run linter"
	@echo "  fmt            - Format code"
	@echo "  vet            - Run go vet"
	@echo "  deps           - Install dependencies"
	@echo "  tools          - Install development tools"
	@echo "  gorelease-test - Test goreleaser configuration"
	@echo "  ci             - Run linting, testing and other CI tasks"
	@echo "  help           - Show this help message"