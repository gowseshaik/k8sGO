package config

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"time"

	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/tools/clientcmd/api"
)

type ConfigManager struct {
	KubeConfig     *api.Config
	CurrentContext string
	AvailableTools []string
	Settings       *UserSettings
}

type UserSettings struct {
	DefaultNamespace string
	RefreshInterval  time.Duration
	ColorScheme      string
	KeyBindings      map[string]string
	ShowBanner       bool
	PageSize         int
}

func NewConfigManager() (*ConfigManager, error) {
	cm := &ConfigManager{
		Settings: &UserSettings{
			DefaultNamespace: "default",
			RefreshInterval:  5 * time.Second,
			ColorScheme:      "default",
			KeyBindings:      getDefaultKeyBindings(),
			ShowBanner:       true,
			PageSize:         50,
		},
	}

	if err := cm.loadKubeConfig(); err != nil {
		return nil, err
	}

	cm.detectAvailableTools()
	return cm, nil
}

func (cm *ConfigManager) loadKubeConfig() error {
	configPath := getKubeConfigPath()

	if configPath == "" {
		if runtime.GOOS == "windows" {
			return fmt.Errorf("kubeconfig not found - please ensure you have a valid kubeconfig file at %s\\.kube\\config", os.Getenv("USERPROFILE"))
		} else {
			return fmt.Errorf("kubeconfig not found - please ensure you have a valid kubeconfig file at $HOME/.kube/config")
		}
	}

	config, err := clientcmd.LoadFromFile(configPath)
	if err != nil {
		return fmt.Errorf("failed to load kubeconfig from %s: %w", configPath, err)
	}

	cm.KubeConfig = config
	cm.CurrentContext = config.CurrentContext
	return nil
}

func (cm *ConfigManager) detectAvailableTools() {
	var tools []string

	if isToolAvailable("oc") {
		tools = append(tools, "oc")
	}
	if isToolAvailable("kubectl") {
		tools = append(tools, "kubectl")
	}

	cm.AvailableTools = tools
}

func getKubeConfigPath() string {
	// Check KUBECONFIG environment variable first
	if kubeconfig := os.Getenv("KUBECONFIG"); kubeconfig != "" {
		if _, err := os.Stat(kubeconfig); err == nil {
			return kubeconfig
		}
	}

	var configPath string

	// Platform-specific paths
	if runtime.GOOS == "windows" {
		// Windows: %USERPROFILE%\.kube\config
		userProfile := os.Getenv("USERPROFILE")
		if userProfile != "" {
			configPath = filepath.Join(userProfile, ".kube", "config")
		}
	} else {
		// Linux/macOS: $HOME/.kube/config
		homeDir := os.Getenv("HOME")
		if homeDir != "" {
			configPath = filepath.Join(homeDir, ".kube", "config")
		}
	}

	// Fallback to os.UserHomeDir() if environment variables are not set
	if configPath == "" {
		homeDir, err := os.UserHomeDir()
		if err != nil {
			return ""
		}
		configPath = filepath.Join(homeDir, ".kube", "config")
	}

	// Check if config file exists
	if _, err := os.Stat(configPath); err == nil {
		return configPath
	}

	return ""
}

func isToolAvailable(tool string) bool {
	var toolName string
	if runtime.GOOS == "windows" {
		toolName = tool + ".exe"
	} else {
		toolName = tool
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

func getDefaultKeyBindings() map[string]string {
	return map[string]string{
		"up":     "k",
		"down":   "j",
		"left":   "h",
		"right":  "l",
		"search": "/",
		"help":   "?",
		"quit":   "q",
		"enter":  "enter",
		"escape": "esc",
	}
}

func (cm *ConfigManager) GetContexts() []string {
	var contexts []string
	for name := range cm.KubeConfig.Contexts {
		contexts = append(contexts, name)
	}
	return contexts
}

func (cm *ConfigManager) SwitchContext(contextName string) error {
	if _, exists := cm.KubeConfig.Contexts[contextName]; !exists {
		return clientcmd.ErrNoContext
	}

	cm.CurrentContext = contextName
	cm.KubeConfig.CurrentContext = contextName
	return nil
}
