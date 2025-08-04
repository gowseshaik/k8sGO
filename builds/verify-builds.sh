#!/bin/bash

echo "k8sGO Build Verification"
echo "========================"
echo

# Check if builds directory exists
if [ ! -d "." ]; then
    echo "Error: Please run this script from the builds directory"
    exit 1
fi

echo "Checking Windows builds..."
if [ -f "windows/k8sgo-windows-amd64.exe" ]; then
    size=$(ls -lh windows/k8sgo-windows-amd64.exe | awk '{print $5}')
    echo "✅ Windows amd64: $size"
else
    echo "❌ Windows amd64: Missing"
fi

if [ -f "windows/k8sgo-windows-arm64.exe" ]; then
    size=$(ls -lh windows/k8sgo-windows-arm64.exe | awk '{print $5}')
    echo "✅ Windows arm64: $size"
else
    echo "❌ Windows arm64: Missing"
fi

echo
echo "Checking Linux builds..."
if [ -f "linux/k8sgo-linux-amd64" ]; then
    size=$(ls -lh linux/k8sgo-linux-amd64 | awk '{print $5}')
    echo "✅ Linux amd64: $size"
else
    echo "❌ Linux amd64: Missing"
fi

if [ -f "linux/k8sgo-linux-arm64" ]; then
    size=$(ls -lh linux/k8sgo-linux-arm64 | awk '{print $5}')
    echo "✅ Linux arm64: $size"
else
    echo "❌ Linux arm64: Missing"
fi

echo
echo "Checking macOS builds..."
if [ -f "linux/k8sgo-darwin-amd64" ]; then
    size=$(ls -lh linux/k8sgo-darwin-amd64 | awk '{print $5}')
    echo "✅ macOS amd64: $size"
else
    echo "❌ macOS amd64: Missing"
fi

if [ -f "linux/k8sgo-darwin-arm64" ]; then
    size=$(ls -lh linux/k8sgo-darwin-arm64 | awk '{print $5}')
    echo "✅ macOS arm64: $size"
else
    echo "❌ macOS arm64: Missing"
fi

echo
echo "Build verification complete!"
echo
echo "Usage examples:"
echo "  Windows: k8sgo-windows-amd64.exe"
echo "  Linux:   chmod +x linux/k8sgo-linux-amd64 && ./linux/k8sgo-linux-amd64"
echo "  macOS:   chmod +x linux/k8sgo-darwin-amd64 && ./linux/k8sgo-darwin-amd64"