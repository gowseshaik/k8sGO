# Kubernetes Terminal Monitor Makefile

# Application name
APP_NAME := k8sgo

# Build variables
BUILD_DIR := .
VENDOR_DIR := vendor
GO_VERSION := 1.24.4
DIST_DIR := dist

# Cross-platform build targets
PLATFORMS := linux/amd64 linux/arm64 windows/amd64 windows/arm64 darwin/amd64 darwin/arm64
BINARY_EXT_windows := .exe

# Default target
.PHONY: all
all: build

# Build the application using vendor dependencies (offline capable)
.PHONY: build
build:
	@echo "Building $(APP_NAME) with vendor dependencies..."
	go build -mod=vendor -o $(APP_NAME) k8sgo.go
	@echo "Build completed: $(APP_NAME)"

# Download and vendor all dependencies for offline use
.PHONY: vendor
vendor:
	@echo "Downloading and vendoring dependencies..."
	go mod tidy
	go mod vendor
	@echo "Dependencies vendored to $(VENDOR_DIR)/"

# Clean build artifacts
.PHONY: clean
clean:
	@echo "Cleaning build artifacts..."
	rm -f $(APP_NAME)
	@echo "Clean completed"

# Clean everything including vendor directory
.PHONY: clean-all
clean-all: clean
	@echo "Removing vendor directory..."
	rm -rf $(VENDOR_DIR)
	@echo "Full clean completed"

# Run the application
.PHONY: run
run: build
	./$(APP_NAME)

# Verify Go version
.PHONY: check-go
check-go:
	@echo "Checking Go version..."
	@go version | grep -q "go$(GO_VERSION)" || (echo "Warning: Expected Go $(GO_VERSION), got $$(go version)" && exit 1)
	@echo "Go version check passed"

# Build for offline deployment (includes verification)
.PHONY: build-offline
build-offline: check-go vendor build
	@echo "Offline build completed successfully"
	@echo "You can now copy the $(APP_NAME) binary and vendor/ directory to offline systems"

# Install dependencies (requires internet)
.PHONY: deps
deps:
	go mod download
	go mod verify

# Development build with race detection
.PHONY: build-dev
build-dev:
	@echo "Building development version with race detection..."
	go build -mod=vendor -race -o $(APP_NAME)-dev k8sgo.go

# Cross-platform builds
.PHONY: build-all
build-all: vendor
	@echo "Building for all platforms..."
	@mkdir -p $(DIST_DIR)
	@for platform in $(PLATFORMS); do \
		OS=$$(echo $$platform | cut -d'/' -f1); \
		ARCH=$$(echo $$platform | cut -d'/' -f2); \
		EXT=$$(eval echo \$$BINARY_EXT_$$OS); \
		OUTPUT=$(DIST_DIR)/$(APP_NAME)-$$OS-$$ARCH$$EXT; \
		echo "Building $$OUTPUT..."; \
		GOOS=$$OS GOARCH=$$ARCH CGO_ENABLED=0 go build -mod=vendor -ldflags="-s -w" -o $$OUTPUT k8sgo.go; \
	done
	@echo "Cross-platform build completed. Binaries available in $(DIST_DIR)/"

# Build for specific platform
.PHONY: build-linux build-windows build-darwin
build-linux: vendor
	@echo "Building for Linux (amd64 and arm64)..."
	@mkdir -p $(DIST_DIR)
	GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build -mod=vendor -ldflags="-s -w" -o $(DIST_DIR)/$(APP_NAME)-linux-amd64 k8sgo.go
	GOOS=linux GOARCH=arm64 CGO_ENABLED=0 go build -mod=vendor -ldflags="-s -w" -o $(DIST_DIR)/$(APP_NAME)-linux-arm64 k8sgo.go

build-windows: vendor
	@echo "Building for Windows (amd64 and arm64)..."
	@mkdir -p $(DIST_DIR)
	GOOS=windows GOARCH=amd64 CGO_ENABLED=0 go build -mod=vendor -ldflags="-s -w" -o $(DIST_DIR)/$(APP_NAME)-windows-amd64.exe k8sgo.go
	GOOS=windows GOARCH=arm64 CGO_ENABLED=0 go build -mod=vendor -ldflags="-s -w" -o $(DIST_DIR)/$(APP_NAME)-windows-arm64.exe k8sgo.go

build-darwin: vendor
	@echo "Building for macOS (amd64 and arm64)..."
	@mkdir -p $(DIST_DIR)
	GOOS=darwin GOARCH=amd64 CGO_ENABLED=0 go build -mod=vendor -ldflags="-s -w" -o $(DIST_DIR)/$(APP_NAME)-darwin-amd64 k8sgo.go
	GOOS=darwin GOARCH=arm64 CGO_ENABLED=0 go build -mod=vendor -ldflags="-s -w" -o $(DIST_DIR)/$(APP_NAME)-darwin-arm64 k8sgo.go

# Create release packages
.PHONY: package
package: build-all
	@echo "Creating release packages..."
	@cd $(DIST_DIR) && for binary in $(APP_NAME)-*; do \
		if [[ "$$binary" == *".exe" ]]; then \
			zip "$$binary.zip" "$$binary"; \
		else \
			tar -czf "$$binary.tar.gz" "$$binary"; \
		fi; \
	done
	@echo "Release packages created in $(DIST_DIR)/"

# Clean dist directory
.PHONY: clean-dist
clean-dist:
	@echo "Cleaning distribution directory..."
	rm -rf $(DIST_DIR)

# Clean everything including vendor directory and dist
.PHONY: clean-all
clean-all: clean clean-dist
	@echo "Removing vendor directory..."
	rm -rf $(VENDOR_DIR)
	@echo "Full clean completed"

# Help target
.PHONY: help
help:
	@echo "Available targets:"
	@echo "  build         - Build the application using vendor dependencies"
	@echo "  build-all     - Build for all platforms (Linux, Windows, macOS)"
	@echo "  build-linux   - Build for Linux (amd64, arm64)"
	@echo "  build-windows - Build for Windows (amd64, arm64)"
	@echo "  build-darwin  - Build for macOS (amd64, arm64)"
	@echo "  package       - Create release packages (zip/tar.gz)"
	@echo "  vendor        - Download and vendor all dependencies"
	@echo "  build-offline - Complete offline build (vendor + build + verify)"
	@echo "  run           - Build and run the application"
	@echo "  clean         - Remove build artifacts"
	@echo "  clean-dist    - Remove distribution directory"
	@echo "  clean-all     - Remove build artifacts, vendor, and dist directories"
	@echo "  check-go      - Verify Go version"
	@echo "  build-dev     - Build development version with race detection"
	@echo "  help          - Show this help message"