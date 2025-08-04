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

	if isToolAvailable("oc") {
		tools = append(tools, "oc")
	}
	if isToolAvailable("kubectl") {
		tools = append(tools, "kubectl")
	}

	td.availableTools = tools
}

func (td *ToolDetector) GetAvailableTools() []string {
	return td.availableTools
}

func (td *ToolDetector) DetectClusterTool(client kubernetes.Interface) (string, error) {
	if client == nil {
		return td.detectBinaryTool()
	}

	clusterType, err := detectClusterType(client)
	if err != nil {
		return td.detectBinaryTool()
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
	if isToolAvailable("oc") {
		return "oc", nil
	}
	if isToolAvailable("kubectl") {
		return "kubectl", nil
	}
	return "", errors.New("no compatible tool found")
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
