# K8sGO Web GUI Makefile
.PHONY: build run clean deps build-all install

VERSION := 1.0.0-web

# Default target
build:
	go build -ldflags "-X main.Version=$(VERSION)" -o build/k8sgo-web ./cmd/k8sgo-web

# Run the application
run:
	go run ./cmd/k8sgo-web

# Clean build artifacts
clean:
	rm -rf build

# Download dependencies
deps:
	go mod tidy
	go mod download

# Cross-compile for all platforms
build-all: clean
	mkdir -p build
	# Windows
	GOOS=windows GOARCH=amd64 go build -ldflags "-X main.Version=$(VERSION)" -o build/k8sgo-web-windows-amd64.exe ./cmd/k8sgo-web
	GOOS=windows GOARCH=386 go build -ldflags "-X main.Version=$(VERSION)" -o build/k8sgo-web-windows-386.exe ./cmd/k8sgo-web
	# Linux
	GOOS=linux GOARCH=amd64 go build -ldflags "-X main.Version=$(VERSION)" -o build/k8sgo-web-linux-amd64 ./cmd/k8sgo-web
	GOOS=linux GOARCH=386 go build -ldflags "-X main.Version=$(VERSION)" -o build/k8sgo-web-linux-386 ./cmd/k8sgo-web
	GOOS=linux GOARCH=arm64 go build -ldflags "-X main.Version=$(VERSION)" -o build/k8sgo-web-linux-arm64 ./cmd/k8sgo-web
	# macOS
	GOOS=darwin GOARCH=amd64 go build -ldflags "-X main.Version=$(VERSION)" -o build/k8sgo-web-darwin-amd64 ./cmd/k8sgo-web
	GOOS=darwin GOARCH=arm64 go build -ldflags "-X main.Version=$(VERSION)" -o build/k8sgo-web-darwin-arm64 ./cmd/k8sgo-web

# Install to system (Unix only)
install: build
	sudo cp build/k8sgo-web /usr/local/bin/

# Test the application
test:
	go test -v ./...

# Run with verbose output
debug:
	go run -race ./cmd/k8sgo-web