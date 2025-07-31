package utils

import (
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"

	"k8s.io/client-go/kubernetes"
)

type ToolDetector struct {
	availableTools []string
}

func NewToolDetector() *ToolDetector {
	td := &ToolDetector{}
	td.detectAvailableTools()
	return td
}

func (td *ToolDetector) detectAvailableTools() {
	var tools []string

	// Prioritize kubectl over oc for broader Kubernetes compatibility
	if isToolAvailable("kubectl") {
		tools = append(tools, "kubectl")
	}
	if isToolAvailable("oc") {
		tools = append(tools, "oc")
	}

	td.availableTools = tools
}

func (td *ToolDetector) GetAvailableTools() []string {
	return td.availableTools
}

func (td *ToolDetector) DetectClusterTool(client kubernetes.Interface) (string, error) {
	if client == nil {
		return td.DetectSmartBinaryTool()
	}

	clusterType, err := detectClusterType(client)
	if err != nil {
		return td.DetectSmartBinaryTool()
	}

	if clusterType == "openshift" && isToolAvailable("oc") {
		return "oc", nil
	}

	if isToolAvailable("kubectl") {
		return "kubectl", nil
	}

	return "", errors.New("no compatible tool found")
}

func (td *ToolDetector) detectBinaryTool() (string, error) {
	// Prioritize kubectl over oc for broader Kubernetes compatibility
	if isToolAvailable("kubectl") {
		return "kubectl", nil
	}
	if isToolAvailable("oc") {
		return "oc", nil
	}
	return "", errors.New("no compatible tool found")
}

// DetectSmartBinaryTool detects the best tool based on cluster type
func (td *ToolDetector) DetectSmartBinaryTool() (string, error) {
	// Check if both tools are available
	kubectlAvailable := isToolAvailable("kubectl")
	ocAvailable := isToolAvailable("oc")

	if !kubectlAvailable && !ocAvailable {
		return "", errors.New("no compatible tool found")
	}

	// If only one tool is available, use it
	if kubectlAvailable && !ocAvailable {
		return "kubectl", nil
	}
	if ocAvailable && !kubectlAvailable {
		return "oc", nil
	}

	// Both tools available - detect cluster type using kubectl api-resources
	if kubectlAvailable {
		if td.isOpenShiftCluster() {
			return "oc", nil // Use oc for OpenShift clusters
		}
		return "kubectl", nil // Use kubectl for standard Kubernetes
	}

	return "kubectl", nil // Default fallback
}

// isOpenShiftCluster checks if the cluster has OpenShift-specific resources
func (td *ToolDetector) isOpenShiftCluster() bool {
	cmd := exec.Command("kubectl", "api-resources")
	output, err := cmd.Output()
	if err != nil {
		return false
	}

	// Check for route resources which are OpenShift-specific
	return strings.Contains(strings.ToLower(string(output)), "route")
}

func detectClusterType(client kubernetes.Interface) (string, error) {
	_, err := client.Discovery().ServerResourcesForGroupVersion("route.openshift.io/v1")
	if err == nil {
		return "openshift", nil
	}

	groups, err := client.Discovery().ServerGroups()
	if err != nil {
		return "", err
	}

	for _, group := range groups.Groups {
		if strings.Contains(group.Name, "openshift") {
			return "openshift", nil
		}
	}

	return "kubernetes", nil
}

func isToolAvailable(tool string) bool {
	var toolName string
	if runtime.GOOS == "windows" {
		toolName = tool + ".exe"
	} else {
		toolName = tool
	}

	if _, err := exec.LookPath(toolName); err == nil {
		return true
	}

	path := os.Getenv("PATH")
	for _, dir := range filepath.SplitList(path) {
		fullPath := filepath.Join(dir, toolName)
		if _, err := os.Stat(fullPath); err == nil {
			return true
		}
	}
	return false
}

func GetToolVersion(tool string) (string, error) {
	cmd := exec.Command(tool, "version", "--client", "--short")
	output, err := cmd.Output()
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(output)), nil
}
