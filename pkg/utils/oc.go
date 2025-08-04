package utils

import (
	"os/exec"
	"strings"
)

// CLICommands provides an abstraction for both oc and kubectl commands
type CLICommands struct {
	tool string // "oc" or "kubectl"
}

// OcCommands is kept for backward compatibility, now wraps CLICommands
type OcCommands struct {
	*CLICommands
}

// NewCLICommands creates a new CLI commands instance with auto-detected tool
func NewCLICommands() *CLICommands {
	detector := NewToolDetector()
	tool, err := detector.detectBinaryTool()
	if err != nil {
		// Default to oc if no tool detected (for backward compatibility)
		tool = "oc"
	}
	return &CLICommands{tool: tool}
}

// NewOcCommands creates a new OcCommands instance (backward compatibility)
func NewOcCommands() *OcCommands {
	return &OcCommands{
		CLICommands: NewCLICommands(),
	}
}

// GetTool returns the currently detected CLI tool
func (c *CLICommands) GetTool() string {
	return c.tool
}

// GetContexts returns all available contexts using the detected CLI tool
func (oc *OcCommands) GetContexts() ([]string, string, error) {
	cmd := exec.Command(oc.tool, "config", "get-contexts", "-o", "name")
	output, err := cmd.Output()
	if err != nil {
		return nil, "", err
	}

	contexts := strings.Split(strings.TrimSpace(string(output)), "\n")

	// Get current context
	currentCmd := exec.Command(oc.tool, "config", "current-context")
	currentOutput, err := currentCmd.Output()
	if err != nil {
		return contexts, "", nil
	}

	currentContext := strings.TrimSpace(string(currentOutput))
	return contexts, currentContext, nil
}

// SwitchContext switches to the specified context using the detected CLI tool
func (oc *OcCommands) SwitchContext(contextName string) error {
	cmd := exec.Command(oc.tool, "config", "use-context", contextName)
	return cmd.Run()
}

// GetNamespaces returns all namespaces using the detected CLI tool
func (oc *OcCommands) GetNamespaces() ([]string, error) {
	var cmd *exec.Cmd

	// Use oc get projects for OpenShift (shows only accessible projects)
	// Use kubectl get namespaces for Kubernetes (shows all namespaces)
	if oc.tool == "oc" {
		cmd = exec.Command(oc.tool, "get", "projects", "-o", "name")
	} else {
		cmd = exec.Command(oc.tool, "get", "namespaces", "-o", "name")
	}

	output, err := cmd.Output()
	if err != nil {
		return nil, err
	}

	lines := strings.Split(strings.TrimSpace(string(output)), "\n")
	var namespaces []string

	for _, line := range lines {
		var ns string
		if oc.tool == "oc" {
			// Remove "project.project.openshift.io/" prefix for oc get projects
			ns = strings.TrimPrefix(line, "project.project.openshift.io/")
		} else {
			// Remove "namespace/" prefix for kubectl get namespaces
			ns = strings.TrimPrefix(line, "namespace/")
		}

		if ns != "" {
			namespaces = append(namespaces, ns)
		}
	}

	return namespaces, nil
}

// GetCurrentNamespace returns current namespace
func (oc *OcCommands) GetCurrentNamespace() (string, error) {
	// Use different commands for oc vs kubectl
	if oc.tool == "oc" {
		cmd := exec.Command(oc.tool, "project", "-q")
		output, err := cmd.Output()
		if err != nil {
			return "default", nil // fallback to default
		}
		return strings.TrimSpace(string(output)), nil
	} else {
		// kubectl uses different command
		cmd := exec.Command(oc.tool, "config", "view", "--minify", "-o", "jsonpath={..namespace}")
		output, err := cmd.Output()
		if err != nil {
			return "default", nil // fallback to default
		}
		namespace := strings.TrimSpace(string(output))
		if namespace == "" {
			return "default", nil
		}
		return namespace, nil
	}
}

// GetPods returns actual pod information parsed from the detected CLI tool
func (oc *OcCommands) GetPods(namespace string) ([]PodInfo, error) {
	var cmd *exec.Cmd

	if namespace == "" {
		cmd = exec.Command(oc.tool, "get", "pods")
	} else {
		cmd = exec.Command(oc.tool, "get", "pods", "-n", namespace)
	}

	output, err := cmd.Output()
	if err != nil {
		return nil, err
	}

	return oc.parsePodOutput(string(output)), nil
}

// GetServices returns service information
func (oc *OcCommands) GetServices(namespace string) ([]PodInfo, error) {
	var cmd *exec.Cmd

	if namespace == "" {
		cmd = exec.Command(oc.tool, "get", "services")
	} else {
		cmd = exec.Command(oc.tool, "get", "services", "-n", namespace)
	}

	output, err := cmd.Output()
	if err != nil {
		return nil, err
	}

	return oc.parseServiceOutput(string(output)), nil
}

// GetDeployments returns deployment information
func (oc *OcCommands) GetDeployments(namespace string) ([]PodInfo, error) {
	var cmd *exec.Cmd

	if namespace == "" {
		cmd = exec.Command(oc.tool, "get", "deployments")
	} else {
		cmd = exec.Command(oc.tool, "get", "deployments", "-n", namespace)
	}

	output, err := cmd.Output()
	if err != nil {
		return nil, err
	}

	return oc.parseDeploymentOutput(string(output)), nil
}

// GetPVC returns persistent volume claims information
func (oc *OcCommands) GetPVC(namespace string) ([]PodInfo, error) {
	var cmd *exec.Cmd

	if namespace == "" {
		cmd = exec.Command(oc.tool, "get", "pvc")
	} else {
		cmd = exec.Command(oc.tool, "get", "pvc", "-n", namespace)
	}

	output, err := cmd.Output()
	if err != nil {
		return nil, err
	}

	return oc.parseGenericOutput(string(output)), nil
}

// GetImageStreams returns image streams information (OpenShift only)
func (oc *OcCommands) GetImageStreams(namespace string) ([]PodInfo, error) {
	// ImageStreams are OpenShift-specific, skip if using kubectl
	if oc.tool != "oc" {
		return []PodInfo{}, nil
	}
	var cmd *exec.Cmd

	if namespace == "" {
		cmd = exec.Command(oc.tool, "get", "is")
	} else {
		cmd = exec.Command(oc.tool, "get", "is", "-n", namespace)
	}

	output, err := cmd.Output()
	if err != nil {
		return nil, err
	}

	return oc.parseImageStreamOutput(string(output)), nil
}

// GetSecrets returns secrets information
func (oc *OcCommands) GetSecrets(namespace string) ([]PodInfo, error) {
	var cmd *exec.Cmd

	if namespace == "" {
		cmd = exec.Command(oc.tool, "get", "secrets")
	} else {
		cmd = exec.Command(oc.tool, "get", "secrets", "-n", namespace)
	}

	output, err := cmd.Output()
	if err != nil {
		return nil, err
	}

	return oc.parseGenericOutput(string(output)), nil
}

// GetConfigMaps returns configmaps information
func (oc *OcCommands) GetConfigMaps(namespace string) ([]PodInfo, error) {
	var cmd *exec.Cmd

	if namespace == "" {
		cmd = exec.Command(oc.tool, "get", "configmaps")
	} else {
		cmd = exec.Command(oc.tool, "get", "configmaps", "-n", namespace)
	}

	output, err := cmd.Output()
	if err != nil {
		return nil, err
	}

	return oc.parseGenericOutput(string(output)), nil
}

// GetEvents returns events information
func (oc *OcCommands) GetEvents(namespace string) ([]PodInfo, error) {
	// For events resource type, always show all namespaces
	cmd := exec.Command(oc.tool, "get", "events", "--all-namespaces")

	output, err := cmd.Output()
	if err != nil {
		return nil, err
	}

	return oc.parseEventsOutput(string(output)), nil
}

type PodInfo struct {
	Name     string
	Ready    string
	Status   string
	Restarts string
	Age      string
	CPU      string // For storing additional data like EXTERNAL-IP
	Memory   string // For storing additional data like PORT(S)
	// Additional fields for PVC and other resources
	Volume       string
	Capacity     string
	AccessModes  string
	StorageClass string
}

func (oc *OcCommands) parsePodOutput(output string) []PodInfo {
	lines := strings.Split(strings.TrimSpace(output), "\n")
	var pods []PodInfo

	// Skip header line (first line)
	for i, line := range lines {
		if i == 0 || strings.TrimSpace(line) == "" {
			continue
		}

		// Split by whitespace and extract pod info
		fields := strings.Fields(line)
		if len(fields) >= 5 {
			pods = append(pods, PodInfo{
				Name:     fields[0],
				Ready:    fields[1],
				Status:   fields[2],
				Restarts: fields[3],
				Age:      fields[4],
				CPU:      "", // Pods don't typically show CPU in basic listing
				Memory:   "", // Pods don't typically show Memory in basic listing
			})
		}
	}

	return pods
}

func (oc *OcCommands) parseServiceOutput(output string) []PodInfo {
	lines := strings.Split(strings.TrimSpace(output), "\n")
	var services []PodInfo

	// Skip header line (first line)
	for i, line := range lines {
		if i == 0 || strings.TrimSpace(line) == "" {
			continue
		}

		// Split by whitespace and extract service info
		// Expected format: NAME TYPE CLUSTER-IP EXTERNAL-IP PORT(S) AGE
		fields := strings.Fields(line)
		if len(fields) >= 5 {
			externalIP := "<none>"
			ports := ""

			if len(fields) >= 6 {
				externalIP = fields[3]
				ports = fields[4]
			} else if len(fields) == 5 {
				ports = fields[3]
			}

			services = append(services, PodInfo{
				Name:     fields[0],             // NAME
				Ready:    fields[1],             // TYPE -> Ready column
				Status:   fields[2],             // CLUSTER-IP -> Status column
				Restarts: "0",                   // Not used for services
				CPU:      externalIP,            // EXTERNAL-IP -> CPU column
				Memory:   ports,                 // PORT(S) -> Memory column
				Age:      fields[len(fields)-1], // AGE
			})
		}
	}

	return services
}

func (oc *OcCommands) parseDeploymentOutput(output string) []PodInfo {
	lines := strings.Split(strings.TrimSpace(output), "\n")
	var deployments []PodInfo

	// Skip header line (first line)
	for i, line := range lines {
		if i == 0 || strings.TrimSpace(line) == "" {
			continue
		}

		// Split by whitespace and extract deployment info
		// Expected format: NAME READY UP-TO-DATE AVAILABLE AGE
		fields := strings.Fields(line)
		if len(fields) >= 5 {
			deployments = append(deployments, PodInfo{
				Name:     fields[0],             // NAME
				Ready:    fields[1],             // READY
				Status:   fields[2],             // UP-TO-DATE -> Status column
				Restarts: fields[3],             // AVAILABLE -> Restarts column
				CPU:      "",                    // Empty -> CPU column
				Memory:   "",                    // Empty -> Memory column
				Age:      fields[len(fields)-1], // AGE
			})
		}
	}

	return deployments
}

// Generic parser for resources that follow standard format: NAME TYPE/STATUS DATA AGE
func (oc *OcCommands) parseGenericOutput(output string) []PodInfo {
	lines := strings.Split(strings.TrimSpace(output), "\n")
	var resources []PodInfo

	// Skip header line (first line)
	for i, line := range lines {
		if i == 0 || strings.TrimSpace(line) == "" {
			continue
		}

		// Split by whitespace and extract resource info
		fields := strings.Fields(line)
		if len(fields) >= 2 {
			var ready, status, restarts string

			// Try to extract meaningful fields based on what's available
			if len(fields) >= 4 {
				ready = fields[1]
				status = fields[2]
				restarts = fields[3]
			} else if len(fields) >= 3 {
				ready = fields[1]
				status = fields[2]
				restarts = "0"
			} else {
				ready = "N/A"
				status = fields[1]
				restarts = "0"
			}

			resources = append(resources, PodInfo{
				Name:     fields[0],
				Ready:    ready,
				Status:   status,
				Restarts: restarts,
				Age:      fields[len(fields)-1], // Last field is usually AGE
				CPU:      "",                    // Generic resources don't typically show CPU
				Memory:   "",                    // Generic resources don't typically show Memory
			})
		}
	}

	return resources
}

func (oc *OcCommands) parsePVCOutput(output string) []PodInfo {
	lines := strings.Split(strings.TrimSpace(output), "\n")
	var pvcs []PodInfo

	// If no output or only header, return empty
	if len(lines) <= 1 {
		return pvcs
	}

	// Skip header line (first line)
	for i, line := range lines {
		if i == 0 || strings.TrimSpace(line) == "" {
			continue
		}

		// Split by whitespace and extract PVC info
		// PVC format: NAME STATUS VOLUME CAPACITY ACCESS_MODES STORAGECLASS VOLUMEATTRIBUTESCLASS AGE
		fields := strings.Fields(line)
		if len(fields) >= 2 {
			var status, volume, capacity, accessModes, storageClass, age string

			// Handle different field counts gracefully
			name := fields[0]
			status = fields[1]

			if len(fields) >= 8 {
				volume = fields[2]       // VOLUME
				capacity = fields[3]     // CAPACITY
				accessModes = fields[4]  // ACCESS MODES
				storageClass = fields[5] // STORAGECLASS
				age = fields[7]          // AGE (skip VOLUMEATTRIBUTESCLASS)
			} else if len(fields) >= 6 {
				volume = fields[2]
				capacity = fields[3]
				accessModes = fields[4]
				storageClass = fields[5]
				age = fields[len(fields)-1]
			} else {
				// Minimal fields available
				volume = "N/A"
				capacity = "N/A"
				accessModes = "N/A"
				storageClass = "N/A"
				age = fields[len(fields)-1]
			}

			// No truncation - show full names and fields

			pvcs = append(pvcs, PodInfo{
				Name:         name,
				Ready:        status,   // STATUS (Bound/Pending)
				Status:       volume,   // Store volume in Status field
				Restarts:     capacity, // Store capacity in Restarts field
				Age:          age,
				CPU:          "", // PVCs don't have CPU data
				Memory:       "", // PVCs don't have Memory data
				Volume:       volume,
				Capacity:     capacity,
				AccessModes:  accessModes,
				StorageClass: storageClass,
			})
		}
	}

	return pvcs
}

func (oc *OcCommands) parseEventsOutput(output string) []PodInfo {
	lines := strings.Split(strings.TrimSpace(output), "\n")
	var events []PodInfo

	// Skip header line (first line)
	for i, line := range lines {
		if i == 0 || strings.TrimSpace(line) == "" {
			continue
		}

		// Split by whitespace and extract event info
		// Event format with --all-namespaces: NAMESPACE   LAST SEEN   TYPE      REASON    OBJECT    MESSAGE
		fields := strings.Fields(line)
		if len(fields) >= 6 {
			// For events with --all-namespaces, we have an extra NAMESPACE field
			namespace := fields[0]
			lastSeen := fields[1]
			eventType := fields[2]
			reason := fields[3]
			object := fields[4]
			// Message might be multiple words, so join the rest
			message := strings.Join(fields[5:], " ")

			// Include namespace in the object name for better identification
			objectWithNamespace := namespace + "/" + object

			events = append(events, PodInfo{
				Name:     objectWithNamespace,
				Ready:    eventType,
				Status:   reason,
				Restarts: message,
				Age:      lastSeen,
				CPU:      "", // Events don't have CPU data
				Memory:   "", // Events don't have Memory data
			})
		}
	}

	return events
}

func (oc *OcCommands) parseImageStreamOutput(output string) []PodInfo {
	lines := strings.Split(strings.TrimSpace(output), "\n")
	var imageStreams []PodInfo

	// Skip header line (first line)
	for i, line := range lines {
		if i == 0 || strings.TrimSpace(line) == "" {
			continue
		}

		// Split by whitespace and extract imagestream info
		// ImageStream format: NAME    IMAGE_REPOSITORY    TAGS    UPDATED
		fields := strings.Fields(line)
		if len(fields) >= 2 {
			var name, imageRepo, tags, updated string

			name = fields[0]

			// Handle different field counts gracefully
			if len(fields) >= 4 {
				imageRepo = fields[1]
				tags = fields[2]
				updated = fields[3]
			} else if len(fields) >= 3 {
				imageRepo = fields[1]
				tags = fields[2]
				updated = "N/A"
			} else {
				// Only name and one other field
				imageRepo = fields[1]
				tags = "N/A"
				updated = "N/A"
			}

			// Map to PodInfo fields appropriately
			// Name = NAME, Ready = TAGS, Status = IMAGE_REPOSITORY, Age = UPDATED
			imageStreams = append(imageStreams, PodInfo{
				Name:     name,
				Ready:    tags,      // Show tags in Ready column
				Status:   imageRepo, // Show image repository in Status column
				Restarts: "0",       // ImageStreams don't have restarts
				Age:      updated,   // Show last updated time
				CPU:      "",        // ImageStreams don't have CPU data
				Memory:   "",        // ImageStreams don't have Memory data
			})
		}
	}

	return imageStreams
}
