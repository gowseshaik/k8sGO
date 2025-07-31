# k8sGO Cross-Platform Builds

This directory contains pre-built binaries of k8sGO for multiple platforms and architectures.

## Available Builds

### Windows
- **k8sgo-windows-amd64.exe** - Windows 64-bit (Intel/AMD)
- **k8sgo-windows-arm64.exe** - Windows 64-bit (ARM, for newer Windows on ARM devices)

### Linux (Ubuntu, Red Hat, CentOS, Fedora, etc.)
- **k8sgo-linux-amd64** - Linux 64-bit (Intel/AMD)
- **k8sgo-linux-arm64** - Linux 64-bit (ARM, for ARM-based Linux systems)

### macOS
- **k8sgo-darwin-amd64** - macOS 64-bit (Intel)
- **k8sgo-darwin-arm64** - macOS 64-bit (Apple Silicon M1/M2)

## Installation Instructions

### Windows
1. Download the appropriate version:
   - For most Windows PCs: `k8sgo-windows-amd64.exe`
   - For Windows on ARM: `k8sgo-windows-arm64.exe`
2. Place the executable in a directory in your PATH or run it directly
3. Run: `k8sgo-windows-amd64.exe`

### Linux (Ubuntu, Red Hat, CentOS, Fedora)
1. Download the appropriate version:
   - For most Linux systems: `k8sgo-linux-amd64`
   - For ARM-based Linux: `k8sgo-linux-arm64`
2. Make it executable: `chmod +x k8sgo-linux-amd64`
3. Run: `./k8sgo-linux-amd64`
4. Optionally, move to `/usr/local/bin/` for system-wide access:
   ```bash
   sudo mv k8sgo-linux-amd64 /usr/local/bin/k8sgo
   sudo chmod +x /usr/local/bin/k8sgo
   ```

### macOS
1. Download the appropriate version:
   - For Intel Macs: `k8sgo-darwin-amd64`
   - For Apple Silicon (M1/M2): `k8sgo-darwin-arm64`
2. Make it executable: `chmod +x k8sgo-darwin-amd64`
3. Run: `./k8sgo-darwin-amd64`
4. You may need to allow the app in System Preferences > Security & Privacy

## Prerequisites

- **CLI Tool**: Either **OpenShift CLI (oc)** or **kubectl** must be installed and configured
- Access to an OpenShift or Kubernetes cluster  
- Proper authentication (logged in via `oc login` or `kubectl` with valid kubeconfig)

## Platform Compatibility

### Tested Platforms
- ✅ Windows 10/11 (64-bit)
- ✅ Ubuntu 18.04+ (64-bit)
- ✅ Red Hat Enterprise Linux 7+ (64-bit)
- ✅ CentOS 7+ (64-bit)
- ✅ Fedora 30+ (64-bit)
- ✅ macOS 10.15+ (64-bit)

### Architecture Support
- **amd64** - Intel/AMD 64-bit processors (most common)
- **arm64** - ARM 64-bit processors (Apple Silicon, ARM servers, etc.)

## Usage

Once installed and with CLI tool configured:

```bash
# Start the application
./k8sgo-linux-amd64

# Auto-detects available CLI tool (oc or kubectl)
# Key bindings:
# [1] Pods
# [2] Services  
# [3] Deployments
# [4] PVCs
# [5] ImageStreams (OpenShift/oc only)
# [6] Secrets
# [7] ConfigMaps
# [8] Events
# [c] Switch Context
# [d] Describe Resource
# [t] Show Tags (ImageStreams only, oc only)
# [s] Show Resources (Pods/Deployments only)
# [r] Refresh
# [q] Quit
```

## Features

- 🎨 **Beautiful ASCII Banner** - Professional k8sGO branding
- 🔧 **Multi-CLI Support** - Auto-detects and works with both `oc` and `kubectl`
- 📊 **Resource Management** - View pods, services, deployments, PVCs, etc.
- 🔄 **Real-time Updates** - Refresh resources with [r]
- 🎯 **Context Switching** - Easy cluster/namespace switching
- 📝 **Resource Description** - Detailed resource information
- 🏷️ **ImageStream Tags** - View available image tags (OpenShift only)
- 📋 **Resource Requests** - View container resource usage
- ⌨️ **Keyboard Navigation** - Efficient keyboard shortcuts
- 🎨 **Color-coded Status** - Visual status indicators
- 🔀 **Cross-Platform** - Works on Windows, Linux, and macOS

## Troubleshooting

### Permission Denied (Linux/macOS)
```bash
chmod +x k8sgo-linux-amd64
```

### "oc command not found"
Install OpenShift CLI:
- **Red Hat/CentOS/Fedora**: `sudo dnf install openshift-clients`
- **Ubuntu/Debian**: Download from OpenShift releases
- **macOS**: `brew install openshift-cli`

### Context/Authentication Issues
```bash
oc login <your-cluster-url>
oc whoami  # Verify authentication
```

## Build Information

- Built with Go 1.21+
- Cross-compiled for multiple platforms
- No external dependencies (statically linked)
- Optimized for terminal environments

## Support

For issues, feature requests, or contributions, please refer to the main project documentation.

---
**Developed By - GNA**