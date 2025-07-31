# k8sGO kubectl Support

k8sGO now supports both **OpenShift CLI (oc)** and **kubectl** with automatic detection and seamless operation.

## Auto-Detection Process

k8sGO automatically detects which CLI tool to use in the following order:

1. **Checks for `oc`** - If OpenShift CLI is available, uses it first
2. **Falls back to `kubectl`** - If `oc` is not found, uses kubectl
3. **Adapts behavior** - Features and commands adapt based on detected tool

## Supported Environments

### ✅ OpenShift Clusters (with `oc`)
- Full feature set including ImageStreams
- OpenShift-specific commands (`oc project`, `oc get is`)
- All resource types available
- Native OpenShift integration

### ✅ Kubernetes Clusters (with `kubectl`)
- Core Kubernetes resources (pods, services, deployments, etc.)
- Standard kubectl commands
- ImageStreams automatically hidden (not applicable)
- Namespace switching via `kubectl config set-context`

### ✅ Mixed Environments
- Works in environments with both tools installed
- Prefers `oc` when both are available
- Consistent user experience across both tools

## Feature Comparison

| Feature | OpenShift (`oc`) | Kubernetes (`kubectl`) |
|---------|------------------|------------------------|
| Pods | ✅ | ✅ |
| Services | ✅ | ✅ |
| Deployments | ✅ | ✅ |
| PVCs | ✅ | ✅ |
| ImageStreams | ✅ | ❌ (Hidden) |
| Secrets | ✅ | ✅ |
| ConfigMaps | ✅ | ✅ |
| Events | ✅ | ✅ |
| Context Switching | ✅ | ✅ |
| Namespace Switching | `oc project` | `kubectl config set-context` |
| Resource Describe | ✅ | ✅ |
| Tags View | ✅ (ImageStreams) | ❌ (N/A) |
| Resource Usage | ✅ | ✅ |

## Command Mapping

k8sGO automatically uses the appropriate commands based on the detected tool:

### Resource Operations
| Function | OpenShift (`oc`) | Kubernetes (`kubectl`) |
|----------|------------------|------------------------|
| Get Pods | `oc get pods` | `kubectl get pods` |
| Get Services | `oc get services` | `kubectl get services` |
| Get Deployments | `oc get deployments` | `kubectl get deployments` |
| Get PVCs | `oc get pvc` | `kubectl get pvc` |
| Get ImageStreams | `oc get is` | ❌ (Skipped) |
| Get Secrets | `oc get secrets` | `kubectl get secrets` |
| Get ConfigMaps | `oc get configmaps` | `kubectl get configmaps` |
| Get Events | `oc get events` | `kubectl get events` |
| Describe Resource | `oc describe <type> <name>` | `kubectl describe <type> <name>` |

### Context & Namespace Operations
| Function | OpenShift (`oc`) | Kubernetes (`kubectl`) |
|----------|------------------|------------------------|
| List Contexts | `oc config get-contexts` | `kubectl config get-contexts` |
| Switch Context | `oc config use-context` | `kubectl config use-context` |
| Current Context | `oc config current-context` | `kubectl config current-context` |
| Current Namespace | `oc project -q` | `kubectl config view --minify -o jsonpath={..namespace}` |
| Switch Namespace | `oc project <namespace>` | `kubectl config set-context --current --namespace=<namespace>` |
| List Namespaces | `oc get namespaces` | `kubectl get namespaces` |

## Installation & Usage

### Prerequisites
Either tool can be installed:

**OpenShift CLI:**
```bash
# Red Hat/CentOS/Fedora
sudo dnf install openshift-clients

# Ubuntu/Debian  
curl -L https://mirror.openshift.com/pub/openshift-v4/clients/ocp/latest/openshift-client-linux.tar.gz | tar xzv
sudo mv oc /usr/local/bin/

# macOS
brew install openshift-cli

# Windows
choco install openshift-cli
```

**kubectl:**
```bash
# Linux
curl -LO "https://dl.k8s.io/release/$(curl -L -s https://dl.k8s.io/release/stable.txt)/bin/linux/amd64/kubectl"
sudo install kubectl /usr/local/bin/

# macOS
brew install kubectl

# Windows
choco install kubernetes-cli
```

### Usage
k8sGO works identically regardless of which tool is detected:

```bash
# Start k8sGO (auto-detects available CLI tool)
./k8sgo

# Same interface and functionality
# Tool detection is shown in the UI
```

## UI Adaptations

### Dynamic Resource List
The numeric key bindings adapt based on available resources:

**With OpenShift (`oc`):**
- [1] Pods
- [2] Services  
- [3] Deployments
- [4] PVCs
- [5] ImageStreams ← OpenShift only
- [6] Secrets
- [7] ConfigMaps
- [8] Events

**With Kubernetes (`kubectl`):**
- [1] Pods
- [2] Services  
- [3] Deployments
- [4] PVCs
- [5] Secrets ← Numbers shift up
- [6] ConfigMaps
- [7] Events

### Tool Indicator
The UI shows which tool is being used:
- **Context Info**: Shows detected tool name
- **Banner**: Adapts description based on tool
  - `oc`: "OpenShift & Kubernetes CLI Tool"
  - `kubectl`: "Kubernetes CLI Tool"

## Troubleshooting

### Tool Not Detected
```bash
# Check if tools are available
which oc kubectl

# Install missing tool (see prerequisites above)

# Verify authentication
oc whoami  # for OpenShift
kubectl auth whoami  # for Kubernetes
```

### Mixed Tool Environment
```bash
# If both tools are installed, k8sGO prefers oc
# To force kubectl usage, temporarily rename/remove oc:
sudo mv /usr/local/bin/oc /usr/local/bin/oc.backup

# Run k8sGO (will use kubectl)
./k8sgo

# Restore oc when done
sudo mv /usr/local/bin/oc.backup /usr/local/bin/oc
```

### Context/Namespace Issues
```bash
# Verify current context
oc config current-context  # or kubectl config current-context

# List available contexts  
oc config get-contexts  # or kubectl config get-contexts

# Ensure proper authentication
oc login <cluster-url>  # for OpenShift
kubectl config use-context <context>  # for Kubernetes
```

## Development Notes

For developers working on k8sGO:

### Key Components
- **`pkg/utils/tools.go`**: Tool detection logic
- **`pkg/utils/oc.go`**: CLI command abstraction (CLICommands struct)
- **`pkg/ui/app.go`**: UI adaptations and resource type management
- **`pkg/ui/keyboard.go`**: Dynamic key binding handling

### Adding New Features
When adding features that may differ between `oc` and `kubectl`:

1. Check `a.state.Tool` to determine active CLI
2. Use appropriate commands for each tool
3. Consider whether feature applies to both environments
4. Update UI elements to reflect availability

### Testing
Test with both environments:
```bash
# Test with OpenShift
oc login <openshift-cluster>
./k8sgo

# Test with Kubernetes  
kubectl config use-context <k8s-context>
./k8sgo
```

---

**k8sGO now provides a unified experience across both OpenShift and Kubernetes environments!** 🚀