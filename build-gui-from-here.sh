#!/bin/bash

# Build the GUI version from the k8sgo directory
set -e

echo "🖥️  Building K8sGO GUI from k8sgo directory"
echo "============================================"

GUI_DIR="../k8sgo-gui"

if [ ! -d "$GUI_DIR" ]; then
    echo "❌ k8sgo-gui directory not found at $GUI_DIR"
    exit 1
fi

echo "📂 GUI project found at: $GUI_DIR"

# Create build directory
mkdir -p "$GUI_DIR/build"

echo "📦 Building GUI application..."

# Set GOPATH to include the GUI project
export GOPATH=$(go env GOPATH)

# Build the GUI application
(cd "$GUI_DIR" && go build -ldflags "-X main.Version=1.0.0-gui" -o build/k8sgo-gui ./cmd/k8sgo-gui)

if [ -f "$GUI_DIR/build/k8sgo-gui" ]; then
    echo "✅ GUI build successful!"
    echo ""
    echo "📍 Binary location: $GUI_DIR/build/k8sgo-gui"
    ls -la "$GUI_DIR/build/"
    echo ""
    echo "🚀 To run the GUI application:"
    echo "   $GUI_DIR/build/k8sgo-gui"
    echo ""
    echo "💡 GUI Features:"
    echo "   • Native desktop interface with Fyne"
    echo "   • Perfect copy/paste (Ctrl+C/Ctrl+V)"
    echo "   • Click and drag text selection"
    echo "   • Top command with one-click copy"
    echo "   • Resource browsing with tabs"
    echo "   • Context switching dropdown"
    echo ""
    echo "🎯 The GUI version completely solves clipboard and mouse selection issues!"
elif [ -f "$GUI_DIR/build/k8sgo-gui.exe" ]; then
    echo "✅ GUI build successful (Windows)!"
    echo "📍 Binary location: $GUI_DIR/build/k8sgo-gui.exe"
    ls -la "$GUI_DIR/build/"
else
    echo "❌ Build failed - binary not found"
    echo "Check for build errors above"
    exit 1
fi