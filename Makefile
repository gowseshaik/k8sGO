# K8sgo Makefile - Cross-platform Kubernetes/OpenShift CLI Tool

BINARY_NAME=k8sgo
VERSION=1.0.0
BUILD_DIR=build
LDFLAGS=-ldflags "-X main.Version=$(VERSION)" -buildvcs=false

# Default target
.PHONY: all
all: clean test build

# Clean build artifacts
.PHONY: clean
clean:
	rm -rf $(BUILD_DIR)
	go clean

# Run tests
.PHONY: test
test:
	go test -v ./...

# Build for current platform
.PHONY: build
build:
	go build $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME) ./cmd/k8sgo

# Build for all platforms
.PHONY: build-all
build-all: clean
	mkdir -p $(BUILD_DIR)
	
	# Windows
	GOOS=windows GOARCH=amd64 go build $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)-windows-amd64.exe ./cmd/k8sgo
	GOOS=windows GOARCH=386 go build $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)-windows-386.exe ./cmd/k8sgo
	
	# Linux
	GOOS=linux GOARCH=amd64 go build $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)-linux-amd64 ./cmd/k8sgo
	GOOS=linux GOARCH=386 go build $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)-linux-386 ./cmd/k8sgo
	GOOS=linux GOARCH=arm64 go build $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)-linux-arm64 ./cmd/k8sgo
	
	# macOS
	GOOS=darwin GOARCH=amd64 go build $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)-darwin-amd64 ./cmd/k8sgo
	GOOS=darwin GOARCH=arm64 go build $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)-darwin-arm64 ./cmd/k8sgo

# Install locally
.PHONY: install
install: build
	cp $(BUILD_DIR)/$(BINARY_NAME) /usr/local/bin/

# Run locally
.PHONY: run
run:
	go run ./cmd/k8sgo

# Format code
.PHONY: fmt
fmt:
	go fmt ./...

# Lint code
.PHONY: lint
lint:
	golangci-lint run

# Get dependencies
.PHONY: deps
deps:
	go mod download
	go mod tidy

# Development setup
.PHONY: dev-setup
dev-setup:
	go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest

# Help
.PHONY: help
help:
	@echo "Available targets:"
	@echo "  all         - Clean, test, and build"
	@echo "  clean       - Remove build artifacts"
	@echo "  test        - Run tests"
	@echo "  build       - Build for current platform"
	@echo "  build-all   - Build for all platforms"
	@echo "  install     - Install locally"
	@echo "  run         - Run locally"
	@echo "  fmt         - Format code"
	@echo "  lint        - Lint code"
	@echo "  deps        - Get dependencies"
	@echo "  dev-setup   - Setup development tools"
	@echo "  help        - Show this help"