#!/bin/bash

echo "Building k8sGO for multiple platforms..."
echo

cd cmd/k8sgo

echo "Building for Windows (amd64)..."
GOOS=windows GOARCH=amd64 go build -buildvcs=false -o ../../builds/windows/k8sgo-windows-amd64.exe
if [ $? -ne 0 ]; then
    echo "Failed to build for Windows amd64"
    exit 1
fi

echo "Building for Windows (arm64)..."
GOOS=windows GOARCH=arm64 go build -buildvcs=false -o ../../builds/windows/k8sgo-windows-arm64.exe
if [ $? -ne 0 ]; then
    echo "Failed to build for Windows arm64"
    exit 1
fi

echo "Building for Linux (amd64) - Ubuntu/Red Hat/CentOS..."
GOOS=linux GOARCH=amd64 go build -buildvcs=false -o ../../builds/linux/k8sgo-linux-amd64
if [ $? -ne 0 ]; then
    echo "Failed to build for Linux amd64"
    exit 1
fi

echo "Building for Linux (arm64)..."
GOOS=linux GOARCH=arm64 go build -buildvcs=false -o ../../builds/linux/k8sgo-linux-arm64
if [ $? -ne 0 ]; then
    echo "Failed to build for Linux arm64"
    exit 1
fi

echo "Building for macOS (amd64)..."
GOOS=darwin GOARCH=amd64 go build -buildvcs=false -o ../../builds/linux/k8sgo-darwin-amd64
if [ $? -ne 0 ]; then
    echo "Failed to build for macOS amd64"
    exit 1
fi

echo "Building for macOS (arm64) - Apple Silicon..."
GOOS=darwin GOARCH=arm64 go build -buildvcs=false -o ../../builds/linux/k8sgo-darwin-arm64
if [ $? -ne 0 ]; then
    echo "Failed to build for macOS arm64"
    exit 1
fi

cd ../..

echo
echo "==============================================="
echo "Build Summary:"
echo "==============================================="
echo "Windows amd64: builds/windows/k8sgo-windows-amd64.exe"
echo "Windows arm64: builds/windows/k8sgo-windows-arm64.exe"
echo "Linux amd64:   builds/linux/k8sgo-linux-amd64"
echo "Linux arm64:   builds/linux/k8sgo-linux-arm64"
echo "macOS amd64:   builds/linux/k8sgo-darwin-amd64"
echo "macOS arm64:   builds/linux/k8sgo-darwin-arm64"
echo "==============================================="
echo
echo "All builds completed successfully!"
echo
echo "Usage Instructions:"
echo "- Windows: Run k8sgo-windows-amd64.exe (or arm64 version)"
echo "- Linux/Red Hat/CentOS/Ubuntu: chmod +x k8sgo-linux-amd64 && ./k8sgo-linux-amd64"
echo "- macOS: chmod +x k8sgo-darwin-amd64 && ./k8sgo-darwin-amd64"
echo