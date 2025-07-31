# k8sGO Installation Guide

## Quick Start

k8sGO is a terminal-based Kubernetes/OpenShift resource management tool with a beautiful interface that works with both `oc` and `kubectl`.

## Prerequisites

1. **CLI Tool** - Either **OpenShift CLI (oc)** or **kubectl** must be installed
2. **Active cluster connection** - Must be logged in via `oc login` or have valid kubeconfig for kubectl

## Platform-Specific Installation

### 🪟 Windows

#### Download and Install
```powershell
# Download the Windows version (choose appropriate architecture)
# For most Windows PCs (Intel/AMD):
curl -L -o k8sgo.exe https://github.com/your-repo/k8sgo/releases/latest/download/k8sgo-windows-amd64.exe

# For Windows on ARM:
curl -L -o k8sgo.exe https://github.com/your-repo/k8sgo/releases/latest/download/k8sgo-windows-arm64.exe

# Run directly
./k8sgo.exe
```

#### Install OpenShift CLI (if not installed)
```powershell
# Using Chocolatey
choco install openshift-cli

# Or download from OpenShift releases page
```

### 🐧 Linux (Ubuntu, Red Hat, CentOS, Fedora)

#### Download and Install
```bash
# For most Linux systems (Intel/AMD 64-bit)
curl -L -o k8sgo https://github.com/your-repo/k8sgo/releases/latest/download/k8sgo-linux-amd64

# For ARM-based Linux systems
curl -L -o k8sgo https://github.com/your-repo/k8sgo/releases/latest/download/k8sgo-linux-arm64

# Make executable
chmod +x k8sgo

# Run
./k8sgo

# Optional: Install system-wide
sudo mv k8sgo /usr/local/bin/
```

#### Install OpenShift CLI

**Red Hat/CentOS/Fedora:**
```bash
sudo dnf install openshift-clients
```

**Ubuntu/Debian:**
```bash
# Download latest release
curl -L https://mirror.openshift.com/pub/openshift-v4/clients/ocp/latest/openshift-client-linux.tar.gz | tar xzv
sudo mv oc /usr/local/bin/
```

### 🍎 macOS

#### Download and Install
```bash
# For Intel Macs
curl -L -o k8sgo https://github.com/your-repo/k8sgo/releases/latest/download/k8sgo-darwin-amd64

# For Apple Silicon (M1/M2)
curl -L -o k8sgo https://github.com/your-repo/k8sgo/releases/latest/download/k8sgo-darwin-arm64

# Make executable
chmod +x k8sgo

# Run (you may need to allow in Security & Privacy)
./k8sgo
```

#### Install OpenShift CLI
```bash
# Using Homebrew
brew install openshift-cli
```

## Configuration

### 1. Connect to your cluster
```bash
# Login to your OpenShift/Kubernetes cluster
oc login https://your-cluster-url:6443

# Verify connection
oc whoami
oc get projects
```

### 2. Set default project/namespace (optional)
```bash
oc project your-project-name
```

## Usage

Start k8sGO:
```bash
./k8sgo
```

### Key Bindings
- **[1]** - View Pods
- **[2]** - View Services
- **[3]** - View Deployments
- **[4]** - View PVCs (Persistent Volume Claims)
- **[5]** - View ImageStreams
- **[6]** - View Secrets
- **[7]** - View ConfigMaps
- **[8]** - View Events
- **[c]** - Switch Context/Cluster
- **[d]** - Describe selected resource
- **[t]** - Show tags (ImageStreams only)
- **[s]** - Show container resources (Pods/Deployments only)
- **[r]** - Refresh current view
- **[?]** - Show help
- **[q]** - Quit

### Navigation
- **↑/k** - Move up
- **↓/j** - Move down
- **←/h** - Previous page
- **→/l** - Next page
- **Esc** - Go back/Switch context
- **Enter** - Select resource

### Mouse Support
- **Scroll** - Navigate through resources
- **Click** - Select resource
- **Right-click** - Copy content

## Auto-Detection Features

k8sGO automatically detects your environment:

- **CLI Tool Detection**: Automatically uses `oc` if available, falls back to `kubectl`
- **Cluster Type Awareness**: Shows OpenShift-specific resources (ImageStreams) only when using `oc`
- **Context Adaptation**: Uses appropriate commands for namespace switching (`oc project` vs `kubectl config set-context`)
- **Dynamic UI**: Resource options adapt based on detected CLI tool

## Features

✨ **Beautiful ASCII Banner** - Professional k8sGO branding  
🔧 **Multi-CLI Support** - Works seamlessly with both `oc` and `kubectl`  
📊 **Multi-Resource Support** - Pods, Services, Deployments, PVCs, etc.  
🔄 **Real-time Updates** - Live resource monitoring  
🎯 **Context Switching** - Easy cluster/namespace management  
📝 **Detailed Descriptions** - Full resource information  
🏷️ **ImageStream Tags** - Container image tag management (OpenShift only)  
📋 **Resource Usage** - Container CPU/Memory requests  
⌨️ **Keyboard Shortcuts** - Efficient navigation  
🎨 **Color Coding** - Visual status indicators  
🖱️ **Mouse Support** - Click and scroll support  

## Troubleshooting

### Common Issues

**"Permission denied"**
```bash
chmod +x k8sgo
```

**"oc: command not found"**
- Install OpenShift CLI (see installation instructions above)

**"No resources found" or authentication errors**
```bash
# Check if you're logged in
oc whoami

# Login if needed
oc login https://your-cluster-url

# Check available projects
oc get projects
```

**"Context switching not working"**
```bash
# List available contexts
oc config get-contexts

# Switch context manually
oc config use-context <context-name>
```

### Platform-Specific Issues

**Windows: "Windows protected your PC"**
- Click "More info" → "Run anyway"
- Or add an exception in Windows Security

**macOS: "Cannot be opened because it is from an unidentified developer"**
- Go to System Preferences → Security & Privacy
- Click "Allow" for k8sgo
- Or run: `sudo spctl --master-disable`

**Linux: "Text rendering issues"**
- Ensure your terminal supports UTF-8
- Try a different terminal emulator (like `gnome-terminal` or `konsole`)

## Advanced Usage

### Custom Installation Path
```bash
# System-wide installation
sudo cp k8sgo /usr/local/bin/
sudo chmod +x /usr/local/bin/k8sgo

# User-specific installation
mkdir -p ~/.local/bin
cp k8sgo ~/.local/bin/
echo 'export PATH="$HOME/.local/bin:$PATH"' >> ~/.bashrc
source ~/.bashrc
```

### Integration with kubectl

While k8sGO is designed for OpenShift (`oc`), it can work with kubectl clusters. The tool will need access to cluster resources through `oc` commands.

## Development

To build from source:
```bash
git clone https://github.com/your-repo/k8sgo
cd k8sgo
go build -o k8sgo cmd/k8sgo/main.go
```

Cross-platform builds:
```bash
./build-all.sh  # Unix/Linux/macOS
build-all.bat   # Windows
```

## Support

- 🐛 **Bug Reports**: Open an issue on GitHub
- 💡 **Feature Requests**: Open an issue with enhancement label
- 📖 **Documentation**: Check the README and docs folder
- 💬 **Questions**: Start a discussion on GitHub

---

**k8sGO - Developed By GNA**  
*Kubernetes & OpenShift CLI Tool*