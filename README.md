# 🔮 K8sgo - Kubernetes & OpenShift CLI Tool

A cross-platform, terminal-based user interface tool for managing both Kubernetes and OpenShift clusters with an intuitive, unified experience.

## Features

- **Cross-Platform**: Supports Windows, Linux (Ubuntu, RHEL, CentOS, Fedora), and macOS
- **Auto-Detection**: Automatically detects and uses kubectl or oc based on cluster type
- **Rich Terminal UI**: Beautiful terminal interface with colors, frames, and branding
- **Resource Management**: View Pods, Services, Deployments with real-time updates
- **Pagination**: Fixed 50 items per page with navigation controls
- **Search & Filter**: Real-time search and filtering capabilities
- **Keyboard Navigation**: Full keyboard navigation with accessibility support
- **Context Management**: Automatic kubeconfig detection and context switching

## Prerequisites

- Go 1.24+ (for building from source)
- kubectl and/or oc CLI tools installed
- Access to a Kubernetes or OpenShift cluster
- Valid kubeconfig file (~/.kube/config)

## Installation

### From Source

```bash
# Clone the repository
git clone <repository-url>
cd k8sgo

# Build for your platform
make build

# Install locally (Linux/macOS)
make install

# Or run directly
make run
```

### Cross-Platform Builds

```bash
# Build for all platforms
make build-all

# Builds will be available in build/ directory:
# - k8sgo-windows-amd64.exe
# - k8sgo-linux-amd64
# - k8sgo-darwin-amd64
# - k8sgo-darwin-arm64
# etc.
```

## Usage

### Basic Usage

```bash
# Run k8sgo
./k8sgo

# Or if installed globally
k8sgo
```

### Keyboard Shortcuts

| Key | Action |
|-----|--------|
| `↑/k` | Move up |
| `↓/j` | Move down |
| `←/h` | Previous page |
| `→/l` | Next page |
| `Enter` | Select resource |
| `/` | Focus search |
| `r` | Refresh |
| `?` | Toggle help |
| `q` | Quit |
| `Esc` | Return to previous view |

### Configuration

K8sgo automatically:
- Detects kubeconfig from `~/.kube/config` or `$KUBECONFIG`
- Identifies available kubectl/oc tools
- Determines cluster type (Kubernetes vs OpenShift)
- Uses appropriate CLI tool based on cluster

## Interface Layout

```
┌─────────────────────────────────────────────────────────────────────┐
│  🔮 K8sgo │ v1.0.0 │ Kubernetes & OpenShift CLI Tool                │
├─ Context: cluster-name │ User: username ─────────────────────────────┤
├─ Namespace: default │ Tool: kubectl/oc ──────────────────────────────┤
├─────────────────────────────────────────────────────────────────────┤
│ Search: [_______________] │ Filter: [All] │ Sort: [Name] │ Refresh   │
├─────────────────────────────────────────────────────────────────────┤
│ NAME               │ READY │ STATUS  │ RESTARTS │ CPU │ MEM │ AGE    │
│ ► pod-1            │ 1/1   │ Running │ 0        │ 10m │ 64Mi│ 5d     │
│   pod-2            │ 0/1   │ Pending │ 0        │ 0   │ 0   │ 2m     │
│   pod-3            │ 1/1   │ Running │ 1        │ 5m  │ 32Mi│ 1h     │
├─────────────────────────────────────────────────────────────────────┤
│ ◄ Prev │ Page 1/3 (50/142 items) │ Next ► │ [?] Help │ [q] Quit     │
└─────────────────────────────────────────────────────────────────────┘
Status: Ready │ Total: 142 │ Displayed: 50 │ Selected: pod-1
```

## Development

### Setup Development Environment

```bash
# Install development dependencies
make dev-setup

# Install project dependencies
make deps

# Format code
make fmt

# Run linter
make lint

# Run tests
make test
```

### Project Structure

```
k8sgo/
├── cmd/k8sgo/          # Main application entry point
├── pkg/
│   ├── client/         # Kubernetes/OpenShift client abstraction
│   ├── config/         # Configuration management
│   ├── ui/             # Terminal UI components
│   └── utils/          # Utility functions
├── internal/types/     # Internal type definitions
├── Makefile           # Build configuration
└── README.md          # This file
```

## Contributing

1. Fork the repository
2. Create a feature branch
3. Make your changes
4. Run tests: `make test`
5. Run linter: `make lint`
6. Submit a pull request

## License

[Add your license here]

## Support

For issues and questions:
1. Check the help system with `?` key
2. Review keyboard shortcuts
3. Check that kubectl/oc is properly installed
4. Verify kubeconfig is accessible