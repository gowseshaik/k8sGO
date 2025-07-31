#!/bin/bash

# Build the portable web GUI version
set -e

echo "🌐 K8sGO Portable Web GUI Builder"
echo "================================="

WEB_DIR="../k8sgo-web"

if [ ! -d "$WEB_DIR" ]; then
    echo "❌ k8sgo-web directory not found at $WEB_DIR"
    exit 1
fi

echo "📂 Web GUI project found at: $WEB_DIR"

# Create build directory
mkdir -p "$WEB_DIR/build"

echo "📦 Installing dependencies..."
(cd "$WEB_DIR" && go mod tidy)

echo "🔨 Building portable web GUI..."

# Build the web application
(cd "$WEB_DIR" && go build -ldflags "-X main.Version=1.0.0-web" -o build/k8sgo-web ./cmd/k8sgo-web)

if [ -f "$WEB_DIR/build/k8sgo-web" ]; then
    echo "✅ Web GUI build successful!"
    echo ""
    echo "📍 Binary location: $WEB_DIR/build/k8sgo-web"
    ls -la "$WEB_DIR/build/"
    echo ""
    echo "🚀 To run the web GUI:"
    echo "   $WEB_DIR/build/k8sgo-web"
    echo "   (Opens browser automatically at http://localhost:8080)"
    echo ""
    echo "🌐 Web GUI Features:"
    echo "   • ✅ PORTABLE - Single binary, no system dependencies"
    echo "   • ✅ Perfect copy/paste - Native browser clipboard (Ctrl+C/Ctrl+V)"
    echo "   • ✅ Text selection - Click and drag anywhere"
    echo "   • ✅ Top command integration - One-click copy"
    echo "   • ✅ Resource browsing - Pods, Services, Deployments"
    echo "   • ✅ Context switching - Dropdown menu"
    echo "   • ✅ Auto-opens browser - Ready to use immediately"
    echo "   • ✅ Cross-platform - Works on Windows/Linux/macOS"
    echo ""
    echo "🎯 This completely solves all clipboard and mouse selection issues!"
    echo "💡 Just run the binary and use your browser - no GUI libraries needed!"
elif [ -f "$WEB_DIR/build/k8sgo-web.exe" ]; then
    echo "✅ Web GUI build successful (Windows)!"
    echo "📍 Binary location: $WEB_DIR/build/k8sgo-web.exe"
    ls -la "$WEB_DIR/build/"
else
    echo "❌ Build failed - binary not found"
    echo "Check for build errors above"
    exit 1
fi