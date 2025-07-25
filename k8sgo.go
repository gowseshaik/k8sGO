package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/muesli/termenv"
	// OpenShift support
	routev1 "github.com/openshift/api/route/v1"
	openshiftclient "github.com/openshift/client-go/apps/clientset/versioned"
	routeclient "github.com/openshift/client-go/route/clientset/versioned"
	projectclient "github.com/openshift/client-go/project/clientset/versioned"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/util/homedir"
)

// ASCII Banner for k8sGo
const k8sGoBanner = `▗▖ ▗▖▄▄▄▄  ▄▄▄  ▗▄▄▖ ▗▄▖ 
▐▌▗▞▘█  █ ▀▄▄  ▐▌   ▐▌ ▐▌
▐▛▚▖ █▀▀█ ▄▄▄▀ ▐▌▝▜▌▐▌ ▐▌
▐▌ ▐▌█▄▄█      ▝▚▄▞▘▝▚▄▞▘

    🚀 Kubernetes Resource Monitor v2.0
         Enhanced Terminal Interface`

// Frame represents different frames in the multi-frame layout
type Frame int

const (
	ContextFrame Frame = iota // Kubernetes context selection
	ResourceFrame            // Resource display frame
	LogFrame                 // Logs display frame
	EventFrame               // Events display frame
	DetailFrame              // Detail information frame
)

// ViewType represents different view modes in the application
type ViewType int

const (
	KubernetesContextView  ViewType = iota // Kubernetes context selection (cluster/context)
	ClusterOrNamespaceView                 // Choose between cluster-level or namespace-level resources
	NamespaceView                          // Namespace selection view
	ResourceView                           // Resource type selection view
	DetailView                             // Detailed resource information view
	LogView                                // Resource logs view with scrolling
	EventView                              // Resource events view for errors/warnings
	MultiFrameView                         // Multi-frame layout for resources and logs
)

// ResourceScope defines whether resource is cluster-scoped or namespace-scoped
type ResourceScope int

const (
	ClusterScoped ResourceScope = iota
	NamespaceScoped
)

// ResourceType represents different Kubernetes resource types
type ResourceType int

const (
	// Cluster-scoped resources
	NodesResource ResourceType = iota
	PersistentVolumesResource
	StorageClassesResource
	ClusterRolesResource
	
	// Namespace-scoped resources
	PodsResource
	ServicesResource
	DeploymentsResource
	ConfigMapsResource
	SecretsResource
	IngressResource
	PersistentVolumeClaimsResource
	ReplicaSetsResource
	DaemonSetsResource
	StatefulSetsResource
	JobsResource
	CronJobsResource
	EventsResource
	
	// OpenShift-specific resources
	RoutesResource
	DeploymentConfigsResource
	ProjectsResource
	BuildConfigsResource
	BuildsResource
	ImageStreamsResource
	
	// Gateway API resources
	GatewaysResource
	HTTPRoutesResource
	GatewayClassesResource
	
	// Additional resources
	NetworkPoliciesResource
	HorizontalPodAutoscalersResource
	VerticalPodAutoscalersResource
)

// ResourceInfo contains metadata about resource types
type ResourceInfo struct {
	Name          string
	Scope         ResourceScope
	SupportsLogs  bool
	SupportsEvents bool
	Icon          string
}

// GetResourceInfo returns information about a resource type
func (rt ResourceType) GetResourceInfo() ResourceInfo {
	resourceInfoMap := map[ResourceType]ResourceInfo{
		// Cluster-scoped resources
		NodesResource:             {"Nodes", ClusterScoped, false, true, "🖥️"},
		PersistentVolumesResource: {"Persistent Volumes", ClusterScoped, false, true, "💾"},
		StorageClassesResource:    {"Storage Classes", ClusterScoped, false, false, "📀"},
		ClusterRolesResource:      {"Cluster Roles", ClusterScoped, false, false, "🔐"},
		
		// Namespace-scoped resources
		PodsResource:                   {"Pods", NamespaceScoped, true, true, "🐳"},
		ServicesResource:               {"Services", NamespaceScoped, false, true, "🌐"},
		DeploymentsResource:            {"Deployments", NamespaceScoped, false, true, "🚀"},
		ConfigMapsResource:             {"ConfigMaps", NamespaceScoped, false, true, "⚙️"},
		SecretsResource:                {"Secrets", NamespaceScoped, false, true, "🔒"},
		IngressResource:                {"Ingress", NamespaceScoped, false, true, "🌍"},
		PersistentVolumeClaimsResource: {"Persistent Volume Claims", NamespaceScoped, false, true, "💽"},
		ReplicaSetsResource:            {"ReplicaSets", NamespaceScoped, false, true, "📊"},
		DaemonSetsResource:             {"DaemonSets", NamespaceScoped, false, true, "⚡"},
		StatefulSetsResource:           {"StatefulSets", NamespaceScoped, false, true, "🏛️"},
		JobsResource:                   {"Jobs", NamespaceScoped, false, true, "⚡"},
		CronJobsResource:               {"CronJobs", NamespaceScoped, false, true, "⏰"},
		EventsResource:                 {"Events", NamespaceScoped, false, false, "📢"},
		
		// OpenShift-specific resources
		RoutesResource:                 {"Routes", NamespaceScoped, false, true, "🔗"},
		DeploymentConfigsResource:      {"DeploymentConfigs", NamespaceScoped, false, true, "⚙️"},
		ProjectsResource:               {"Projects", ClusterScoped, false, true, "📁"},
		BuildConfigsResource:           {"BuildConfigs", NamespaceScoped, false, true, "🔨"},
		BuildsResource:                 {"Builds", NamespaceScoped, true, true, "🏗️"},
		ImageStreamsResource:           {"ImageStreams", NamespaceScoped, false, true, "📸"},
		
		// Gateway API resources
		GatewaysResource:               {"Gateways", NamespaceScoped, false, true, "🚪"},
		HTTPRoutesResource:             {"HTTPRoutes", NamespaceScoped, false, true, "🛤️"},
		GatewayClassesResource:         {"GatewayClasses", ClusterScoped, false, false, "🏭"},
		
		// Additional resources
		NetworkPoliciesResource:        {"NetworkPolicies", NamespaceScoped, false, true, "🛡️"},
		HorizontalPodAutoscalersResource: {"HorizontalPodAutoscalers", NamespaceScoped, false, true, "📈"},
		VerticalPodAutoscalersResource:   {"VerticalPodAutoscalers", NamespaceScoped, false, true, "📊"},
	}
	
	if info, exists := resourceInfoMap[rt]; exists {
		return info
	}
	return ResourceInfo{"Unknown", NamespaceScoped, false, false, "❓"}
}

func (rt ResourceType) String() string {
	return rt.GetResourceInfo().Name
}

// K8sResource represents a generic Kubernetes resource for display
type K8sResource struct {
	Name         string
	Namespace    string
	Status       string
	Age          string
	Details      map[string]string // Additional resource-specific details
	ResourceType ResourceType      // Type of resource for logs functionality
	Errors       []string          // List of errors/issues with this resource
	Warnings     []string          // List of warnings for this resource
}

// LogEntry represents a single log line
type LogEntry struct {
	Timestamp time.Time
	Message   string
	Container string
	Level     string // INFO, WARN, ERROR, DEBUG
}

// EventEntry represents a Kubernetes event
type EventEntry struct {
	Timestamp    time.Time
	Type         string // Normal, Warning
	Reason       string
	Message      string
	Source       string
	Count        int32
}

// Color definitions for catchy UI
type ColorScheme struct {
	Primary     lipgloss.Color
	Secondary   lipgloss.Color
	Success     lipgloss.Color
	Warning     lipgloss.Color
	Error       lipgloss.Color
	Info        lipgloss.Color
	Accent      lipgloss.Color
	Background  lipgloss.Color
	Text        lipgloss.Color
	Muted       lipgloss.Color
}

var colors = ColorScheme{
	Primary:    lipgloss.Color("#4A9EFF"), // Soft Blue
	Secondary:  lipgloss.Color("#FF8C42"), // Warm Orange
	Success:    lipgloss.Color("#28A745"), // Forest Green
	Warning:    lipgloss.Color("#FFC107"), // Amber
	Error:      lipgloss.Color("#DC3545"), // Muted Red
	Info:       lipgloss.Color("#17A2B8"), // Teal
	Accent:     lipgloss.Color("#6F42C1"), // Deep Purple
	Background: lipgloss.Color("#1E1E1E"), // Dark Gray
	Text:       lipgloss.Color("#E5E5E5"), // Light Gray
	Muted:      lipgloss.Color("#3A3A3A"), // Darker Gray for dividers and header
}

// Model represents the application state using Bubble Tea pattern
type Model struct {
	// Kubernetes client
	clientset *kubernetes.Clientset
	ctx       context.Context
	
	// OpenShift clients (optional, will be nil if not available)
	openshiftAppsClient *openshiftclient.Clientset
	routeClient         *routeclient.Clientset
	projectClient       *projectclient.Clientset
	isOpenShift         bool
	
	// Navigation state
	currentView     ViewType
	currentFrame    Frame   // Current active frame in multi-frame view
	cursor          int
	viewStack       []ViewType // For navigation history
	logScrollOffset int        // For log scrolling
	eventScrollOffset int      // For event scrolling
	
	// Multi-frame layout state
	isMultiFrameMode bool
	frameWidth      int
	frameHeight     int

	// Data collections
	kubernetesContexts  []string      // Available Kubernetes contexts
	namespaces          []string
	resourceTypes       []ResourceType
	resources           []K8sResource
	logEntries          []LogEntry
	eventEntries        []EventEntry
	
	// Current selections
	selectedKubeContext  string         // Selected Kubernetes context
	selectedNamespace    string
	selectedResource     ResourceType
	selectedK8sResource  *K8sResource   // For logs/events
	selectedScope        ResourceScope  // Cluster or namespace scoped
	
	// UI state
	width         int
	height        int
	loading       bool
	errorMessage  string
	lastUpdate    time.Time
	
	// Auto-refresh
	autoRefresh   bool
	refreshTicker *time.Ticker
}

// Init initializes the model - required by Bubble Tea
func (m Model) Init() tea.Cmd {
	return tea.Batch(
		m.loadKubernetesContexts(),
		tea.Tick(time.Second*5, func(t time.Time) tea.Msg {
			return refreshMsg{}
		}),
	)
}

// Messages for Bubble Tea messaging system
type namespacesLoadedMsg struct {
	namespaces []string
	err        error
}

type resourcesLoadedMsg struct {
	resources []K8sResource
	err       error
}

type logsLoadedMsg struct {
	logs []LogEntry
	err  error
}

type eventsLoadedMsg struct {
	events []EventEntry
	err    error
}

type refreshMsg struct{}

type kubeContextsLoadedMsg struct {
	contexts []string
	err      error
}

type contextSwitchedMsg struct {
	contextName string
	err         error
}

type clientsReinitializedMsg struct {
	contextName         string
	clientset          *kubernetes.Clientset
	openshiftAppsClient *openshiftclient.Clientset
	routeClient        *routeclient.Clientset
	projectClient      *projectclient.Clientset
	isOpenShift        bool
	err                error
}

// isOpenShiftContext checks if a context is pointing to an OpenShift cluster
func isOpenShiftContext(contextName string) bool {
	// Try to create a client for this context and check for OpenShift APIs
	config, err := clientcmd.NewDefaultClientConfigLoadingRules().Load()
	if err != nil {
		return false
	}
	
	// Set the current context temporarily
	config.CurrentContext = contextName
	clientConfig := clientcmd.NewDefaultClientConfig(*config, &clientcmd.ConfigOverrides{})
	restConfig, err := clientConfig.ClientConfig()
	if err != nil {
		return false
	}
	
	// Try to create an OpenShift project client to test if it's OpenShift
	projectClient, err := projectclient.NewForConfig(restConfig)
	if err != nil {
		return false
	}
	
	// Try to list projects to confirm it's OpenShift
	_, err = projectClient.ProjectV1().Projects().List(context.Background(), metav1.ListOptions{Limit: 1})
	return err == nil
}

// loadKubernetesContexts creates a command to load all available contexts
func (m Model) loadKubernetesContexts() tea.Cmd {
	return func() tea.Msg {
		// Load all contexts from kubeconfig
		config, err := clientcmd.NewDefaultClientConfigLoadingRules().Load()
		if err != nil {
			return kubeContextsLoadedMsg{err: err}
		}
		
		var contexts []string
		for contextName := range config.Contexts {
			contexts = append(contexts, contextName)
		}
		
		// Sort contexts for consistent display
		sort.Strings(contexts)
		
		// If no contexts found, add current context
		if len(contexts) == 0 {
			if config.CurrentContext != "" {
				contexts = append(contexts, config.CurrentContext)
			} else {
				contexts = append(contexts, "default")
			}
		}
		
		return kubeContextsLoadedMsg{contexts: contexts}
	}
}

// switchKubernetesContext switches to the selected context using kubectl
func (m Model) switchKubernetesContext(contextName string) tea.Cmd {
	return func() tea.Msg {
		// Execute kubectl config use-context command
		cmd := exec.Command("kubectl", "config", "use-context", contextName)
		if err := cmd.Run(); err != nil {
			return contextSwitchedMsg{contextName: contextName, err: fmt.Errorf("failed to switch context: %v", err)}
		}
		
		return contextSwitchedMsg{contextName: contextName, err: nil}
	}
}

// execCommand executes a shell command
func execCommand(name string, args ...string) error {
	cmd := exec.Command(name, args...)
	return cmd.Run()
}

// loadNamespaces creates a command to asynchronously load namespaces
func (m Model) loadNamespaces() tea.Cmd {
	return func() tea.Msg {
		nsList, err := m.clientset.CoreV1().Namespaces().List(m.ctx, metav1.ListOptions{})
		if err != nil {
			return namespacesLoadedMsg{err: err}
		}
		
		var namespaces []string
		for _, ns := range nsList.Items {
			namespaces = append(namespaces, ns.Name)
		}
		sort.Strings(namespaces)
		
		return namespacesLoadedMsg{namespaces: namespaces}
	}
}

// loadResources creates a command to asynchronously load resources for current namespace and type
func (m Model) loadResources() tea.Cmd {
	return func() tea.Msg {
		var resources []K8sResource
		var err error
		
		switch m.selectedResource {
		// Cluster-scoped resources
		case NodesResource:
			resources, err = m.loadNodes()
		case PersistentVolumesResource:
			resources, err = m.loadPersistentVolumes()
		case StorageClassesResource:
			resources, err = m.loadStorageClasses()
		case ClusterRolesResource:
			resources, err = m.loadClusterRoles()
		
		// Namespace-scoped resources
		case PodsResource:
			resources, err = m.loadPods()
		case ServicesResource:
			resources, err = m.loadServices()
		case DeploymentsResource:
			resources, err = m.loadDeployments()
		case ConfigMapsResource:
			resources, err = m.loadConfigMaps()
		case SecretsResource:
			resources, err = m.loadSecrets()
		case IngressResource:
			resources, err = m.loadIngress()
		case PersistentVolumeClaimsResource:
			resources, err = m.loadPersistentVolumeClaims()
		case ReplicaSetsResource:
			resources, err = m.loadReplicaSets()
		case DaemonSetsResource:
			resources, err = m.loadDaemonSets()
		case StatefulSetsResource:
			resources, err = m.loadStatefulSets()
		case JobsResource:
			resources, err = m.loadJobs()
		case CronJobsResource:
			resources, err = m.loadCronJobs()
		case EventsResource:
			resources, err = m.loadEventsResource()
		
		// OpenShift-specific resources
		case RoutesResource:
			resources, err = m.loadRoutes()
		case DeploymentConfigsResource:
			resources, err = m.loadDeploymentConfigs()
		case ProjectsResource:
			resources, err = m.loadProjects()
		case BuildConfigsResource:
			resources, err = m.loadBuildConfigs()
		case BuildsResource:
			resources, err = m.loadBuilds()
		case ImageStreamsResource:
			resources, err = m.loadImageStreams()
		
		// Gateway API resources
		case GatewaysResource:
			resources, err = m.loadGateways()
		case HTTPRoutesResource:
			resources, err = m.loadHTTPRoutes()
		case GatewayClassesResource:
			resources, err = m.loadGatewayClasses()
		
		// Additional resources
		case NetworkPoliciesResource:
			resources, err = m.loadNetworkPolicies()
		case HorizontalPodAutoscalersResource:
			resources, err = m.loadHorizontalPodAutoscalers()
		case VerticalPodAutoscalersResource:
			resources, err = m.loadVerticalPodAutoscalers()
		}
		
		return resourcesLoadedMsg{resources: resources, err: err}
	}
}

// Update handles messages and state transitions - required by Bubble Tea
func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		return m, nil
		
	case tea.KeyMsg:
		return m.handleKeyPress(msg)
		
	case kubeContextsLoadedMsg:
		m.loading = false
		if msg.err != nil {
			m.errorMessage = fmt.Sprintf("Error loading contexts: %v", msg.err)
		} else {
			m.kubernetesContexts = msg.contexts
			m.errorMessage = ""
		}
		return m, nil
		
	case contextSwitchedMsg:
		m.loading = false
		if msg.err != nil {
			m.errorMessage = fmt.Sprintf("Error switching context: %v", msg.err)
		} else {
			// Reinitialize clients after context switch
			return m, m.reinitializeClients(msg.contextName)
		}
		return m, nil
		
	case clientsReinitializedMsg:
		m.loading = false
		if msg.err != nil {
			m.errorMessage = fmt.Sprintf("Error reinitializing clients: %v", msg.err)
		} else {
			// Update model with new clients
			m.clientset = msg.clientset
			m.openshiftAppsClient = msg.openshiftAppsClient
			m.routeClient = msg.routeClient
			m.projectClient = msg.projectClient
			m.isOpenShift = msg.isOpenShift
			m.selectedKubeContext = msg.contextName
			m.errorMessage = ""
			
			// Move to next view
			m.viewStack = append(m.viewStack, m.currentView)
			m.currentView = ClusterOrNamespaceView
			m.cursor = 0
		}
		return m, nil
		
	case namespacesLoadedMsg:
		m.loading = false
		if msg.err != nil {
			if strings.Contains(msg.err.Error(), "no such host") || strings.Contains(msg.err.Error(), "connection refused") {
				m.errorMessage = "❌ Cannot connect to cluster. Please check your kubeconfig and ensure you have a real Kubernetes cluster running."
			} else {
				m.errorMessage = fmt.Sprintf("Error loading namespaces: %v", msg.err)
			}
		} else {
			m.namespaces = msg.namespaces
			m.errorMessage = ""
		}
		return m, nil
		
	case resourcesLoadedMsg:
		m.loading = false
		if msg.err != nil {
			if strings.Contains(msg.err.Error(), "no such host") || strings.Contains(msg.err.Error(), "connection refused") {
				m.errorMessage = "❌ Cannot connect to cluster. Please check your kubeconfig and ensure you have a real Kubernetes cluster running."
			} else {
				m.errorMessage = fmt.Sprintf("Error loading resources: %v", msg.err)
			}
		} else {
			m.resources = msg.resources
			m.errorMessage = ""
			m.lastUpdate = time.Now()
		}
		return m, nil
		
	case logsLoadedMsg:
		m.loading = false
		if msg.err != nil {
			m.errorMessage = fmt.Sprintf("Error loading logs: %v", msg.err)
		} else {
			m.logEntries = msg.logs
			m.errorMessage = ""
			m.lastUpdate = time.Now()
		}
		return m, nil
		
	case eventsLoadedMsg:
		m.loading = false
		if msg.err != nil {
			m.errorMessage = fmt.Sprintf("Error loading events: %v", msg.err)
		} else {
			m.eventEntries = msg.events
			m.errorMessage = ""
			m.lastUpdate = time.Now()
		}
		return m, nil
		
	case refreshMsg:
		// Auto-refresh current view if enabled
		if m.autoRefresh {
			switch m.currentView {
			case DetailView:
				return m, m.loadResources()
			case LogView:
				return m, m.loadLogs()
			case EventView:
				return m, m.loadEventsCmd()
			}
		}
		return m, tea.Tick(time.Second*5, func(t time.Time) tea.Msg {
			return refreshMsg{}
		})
	}
	
	return m, nil
}

// handleKeyPress processes keyboard input and navigation
func (m Model) handleKeyPress(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	
	case "ctrl+c", "q":
		return m, tea.Quit
		
	case "up", "k":
		if m.currentView == LogView {
			if m.logScrollOffset > 0 {
				m.logScrollOffset--
			}
		} else if m.currentView == EventView {
			if m.eventScrollOffset > 0 {
				m.eventScrollOffset--
			}
		} else if m.cursor > 0 {
			m.cursor--
		}
		
	case "down", "j":
		if m.currentView == LogView {
			maxScroll := len(m.logEntries) - (m.height - 10)
			if maxScroll < 0 {
				maxScroll = 0
			}
			if m.logScrollOffset < maxScroll {
				m.logScrollOffset++
			}
		} else if m.currentView == EventView {
			maxScroll := len(m.eventEntries) - (m.height - 10)
			if maxScroll < 0 {
				maxScroll = 0
			}
			if m.eventScrollOffset < maxScroll {
				m.eventScrollOffset++
			}
		} else {
			switch m.currentView {
			case KubernetesContextView:
				if m.cursor < len(m.kubernetesContexts)-1 {
					m.cursor++
				}
			case ClusterOrNamespaceView:
				if m.cursor < 1 {
					m.cursor++
				}
			case NamespaceView:
				if m.cursor < len(m.namespaces)-1 {
					m.cursor++
				}
			case ResourceView:
				if m.cursor < len(m.resourceTypes)-1 {
					m.cursor++
				}
			case DetailView:
				if m.cursor < len(m.resources)-1 {
					m.cursor++
				}
			}
		}
		
	case "enter", " ":
		return m.handleSelection()
		
	case "esc", "backspace":
		return m.navigateBack()
		
	case "l":
		// Show logs for selected resource (if supported)
		if m.currentView == DetailView && len(m.resources) > 0 && m.cursor < len(m.resources) {
			selectedResource := &m.resources[m.cursor]
			if selectedResource.ResourceType.GetResourceInfo().SupportsLogs {
				m.selectedK8sResource = selectedResource
				m.viewStack = append(m.viewStack, m.currentView)
				m.currentView = MultiFrameView
				m.currentFrame = LogFrame
				m.isMultiFrameMode = true
				m.cursor = 0
				m.logScrollOffset = 0
				m.loading = true
				return m, m.loadLogs()
			}
		}
		
	case "e":
		// Show events for selected resource (if supported)
		if m.currentView == DetailView && len(m.resources) > 0 && m.cursor < len(m.resources) {
			selectedResource := &m.resources[m.cursor]
			if selectedResource.ResourceType.GetResourceInfo().SupportsEvents {
				m.selectedK8sResource = selectedResource
				m.viewStack = append(m.viewStack, m.currentView)
				m.currentView = EventView
				m.cursor = 0
				m.eventScrollOffset = 0
				m.loading = true
				return m, m.loadEventsCmd()
			}
		}
		
	case "r":
		// Manual refresh
		switch m.currentView {
		case NamespaceView:
			m.loading = true
			return m, m.loadNamespaces()
		case DetailView:
			m.loading = true
			return m, m.loadResources()
		case LogView:
			m.loading = true
			return m, m.loadLogs()
		case EventView:
			m.loading = true
			return m, m.loadEventsCmd()
		}
		
	case "a":
		// Toggle auto-refresh
		m.autoRefresh = !m.autoRefresh
		if m.autoRefresh {
			return m, tea.Tick(time.Second*5, func(t time.Time) tea.Msg {
				return refreshMsg{}
			})
		} else {
			if m.refreshTicker != nil {
				m.refreshTicker.Stop()
				m.refreshTicker = nil
			}
		}
		
	case "m":
		// Toggle multi-frame mode when viewing resources
		if m.currentView == DetailView {
			m.viewStack = append(m.viewStack, m.currentView)
			m.currentView = MultiFrameView
			m.isMultiFrameMode = true
			m.currentFrame = ResourceFrame
			m.frameWidth = m.width / 2
			m.frameHeight = m.height - 10
		}
		
	case "tab":
		// Switch between frames in multi-frame mode
		if m.currentView == MultiFrameView {
			switch m.currentFrame {
			case ResourceFrame:
				m.currentFrame = LogFrame
			case LogFrame:
				m.currentFrame = EventFrame
			case EventFrame:
				m.currentFrame = ResourceFrame
			}
			m.cursor = 0
		}
	}
	
	return m, nil
}

// handleSelection processes enter/space key selections
func (m Model) handleSelection() (tea.Model, tea.Cmd) {
	switch m.currentView {
	
	case KubernetesContextView:
		if len(m.kubernetesContexts) > 0 && m.cursor < len(m.kubernetesContexts) {
			contextName := m.kubernetesContexts[m.cursor]
			
			// Check if this is already the current context
			cmd := exec.Command("kubectl", "config", "current-context")
			currentContextBytes, _ := cmd.Output()
			currentContext := strings.TrimSpace(string(currentContextBytes))
			
			if contextName == currentContext {
				// Already using this context, proceed directly
				m.selectedKubeContext = contextName
				m.viewStack = append(m.viewStack, m.currentView)
				m.currentView = ClusterOrNamespaceView
				m.cursor = 0
				return m, nil
			}
			
			m.selectedKubeContext = contextName
			m.loading = true
			
			// Switch to the selected context using kubectl
			return m, m.switchKubernetesContext(contextName)
		}
		return m, nil
	
	case ClusterOrNamespaceView:
		m.viewStack = append(m.viewStack, m.currentView)
		if m.cursor == 0 {
			// Cluster-scoped resources
			m.selectedScope = ClusterScoped
			m.currentView = ResourceView
			// Cluster-scoped resources
			m.resourceTypes = []ResourceType{NodesResource, PersistentVolumesResource, StorageClassesResource, ClusterRolesResource}
			// Add OpenShift Projects if available
			if m.isOpenShift {
				m.resourceTypes = append(m.resourceTypes, ProjectsResource)
			}
			// Add Gateway API cluster resources
			m.resourceTypes = append(m.resourceTypes, GatewayClassesResource)
		} else {
			// Namespace-scoped resources
			m.selectedScope = NamespaceScoped
			m.currentView = NamespaceView
			m.loading = true
			return m, m.loadNamespaces()
		}
		m.cursor = 0
		
	case NamespaceView:
		if len(m.namespaces) > 0 && m.cursor < len(m.namespaces) {
			m.selectedNamespace = m.namespaces[m.cursor]
			m.viewStack = append(m.viewStack, m.currentView)
			m.currentView = ResourceView
			m.cursor = 0
			
			// Initialize namespace-scoped resource types including Events
			m.resourceTypes = []ResourceType{
				PodsResource, ServicesResource, DeploymentsResource,
				ConfigMapsResource, SecretsResource, IngressResource,
				PersistentVolumeClaimsResource, ReplicaSetsResource,
				DaemonSetsResource, StatefulSetsResource,
				JobsResource, CronJobsResource, EventsResource,
			}
			// Add OpenShift-specific resources if available
			if m.isOpenShift {
				m.resourceTypes = append(m.resourceTypes, RoutesResource, DeploymentConfigsResource, BuildConfigsResource, BuildsResource, ImageStreamsResource)
			}
			// Add Gateway API resources if available
			m.resourceTypes = append(m.resourceTypes, GatewaysResource, HTTPRoutesResource)
			// Add additional common resources
			m.resourceTypes = append(m.resourceTypes, NetworkPoliciesResource, HorizontalPodAutoscalersResource)
		}
		
	case ResourceView:
		if len(m.resourceTypes) > 0 && m.cursor < len(m.resourceTypes) {
			m.selectedResource = m.resourceTypes[m.cursor]
			m.viewStack = append(m.viewStack, m.currentView)
			m.currentView = DetailView
			m.cursor = 0
			m.loading = true
			
			return m, m.loadResources()
		}
	}
	
	return m, nil
}

// navigateBack handles back navigation
func (m Model) navigateBack() (tea.Model, tea.Cmd) {
	if len(m.viewStack) > 0 {
		// Pop the last view from stack
		lastView := m.viewStack[len(m.viewStack)-1]
		m.viewStack = m.viewStack[:len(m.viewStack)-1]
		m.currentView = lastView
		m.cursor = 0
		m.logScrollOffset = 0
		m.eventScrollOffset = 0
		m.errorMessage = ""
	}
	return m, nil
}

// View renders the UI - required by Bubble Tea
func (m Model) View() string {

	// Define catchy styles using the color scheme - using darker background
	headerStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(colors.Text).
		Background(colors.Muted).
		Padding(0, 1).
		Margin(0, 0, 1, 0).
		Width(m.width).
		Align(lipgloss.Center)
		
	selectedStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(colors.Background).
		Background(colors.Primary).
		Padding(0, 1)
		
	normalStyle := lipgloss.NewStyle().
		Foreground(colors.Text)
		
	errorStyle := lipgloss.NewStyle().
		Foreground(colors.Error).
		Bold(true)
		
	successStyle := lipgloss.NewStyle().
		Foreground(colors.Success).
		Bold(true)
		
	infoStyle := lipgloss.NewStyle().
		Foreground(colors.Info)
		
	helpStyle := lipgloss.NewStyle().
		Foreground(colors.Muted).
		Italic(true)

	var content strings.Builder
	
	// Always show the ASCII banner at the top, centered with gradient effect
	bannerStyle := lipgloss.NewStyle().
		Foreground(colors.Primary).
		Bold(true).
		Align(lipgloss.Center).
		Width(m.width)
	
	// Split banner into ASCII art and title sections
	bannerParts := strings.Split(k8sGoBanner, "\n\n")
	
	// Render ASCII art part with primary color
	if len(bannerParts) > 0 {
		asciiPart := bannerParts[0]
		content.WriteString(bannerStyle.Render(asciiPart) + "\n\n")
	}
	
	// Render title part with accent color
	if len(bannerParts) > 1 {
		titleStyle := lipgloss.NewStyle().
			Foreground(colors.Accent).
			Bold(true).
			Align(lipgloss.Center).
			Width(m.width)
		titlePart := bannerParts[1]
		content.WriteString(titleStyle.Render(titlePart) + "\n")
	}
	
	// Header with updated styling
	content.WriteString(headerStyle.Render(m.buildHeader()) + "\n")
	
	// Error display
	if m.errorMessage != "" {
		content.WriteString(errorStyle.Render("❌ " + m.errorMessage) + "\n\n")
	}
	
	// Loading indicator
	if m.loading {
		content.WriteString(infoStyle.Render("⏳ Loading...") + "\n\n")
	}
	
	// Main content based on current view
	switch m.currentView {
	case KubernetesContextView:
		content.WriteString(successStyle.Render("🔧 Select Context:") + "\n\n")
		
		// Get current context to show active indicator
		cmd := exec.Command("kubectl", "config", "current-context")
		currentContextBytes, _ := cmd.Output()
		currentContext := strings.TrimSpace(string(currentContextBytes))
		
		for i, ctx := range m.kubernetesContexts {
			prefix := "  "
			style := normalStyle
			if i == m.cursor {
				prefix = "▶ "
				style = selectedStyle
			}
			
			// Add active indicator for current context
			contextDisplay := ctx
			if ctx == currentContext {
				contextDisplay = fmt.Sprintf("%s (active)", ctx)
				// Keep the same color instead of changing to green
			}
			
			content.WriteString(prefix + style.Render(contextDisplay) + "\n")
		}
		
	case ClusterOrNamespaceView:
		content.WriteString(successStyle.Render("🌐 Select Resource Scope:") + "\n\n")
		options := []string{"🏢 Cluster Resources", "📁 Namespace Resources"}
		for i, option := range options {
			prefix := "  "
			style := normalStyle
			if i == m.cursor {
				prefix = "▶ "
				style = selectedStyle
			}
			content.WriteString(prefix + style.Render(option) + "\n")
		}
		
	case NamespaceView:
		content.WriteString(successStyle.Render("📁 Select Namespace:") + "\n\n")
		for i, ns := range m.namespaces {
			prefix := "  "
			style := normalStyle
			if i == m.cursor {
				prefix = "▶ "
				style = selectedStyle
			}
			content.WriteString(prefix + style.Render(ns) + "\n")
		}
		
	case ResourceView:
		scopeStr := "cluster"
		if m.selectedScope == NamespaceScoped {
			scopeStr = fmt.Sprintf("namespace '%s'", m.selectedNamespace)
		}
		content.WriteString(successStyle.Render(fmt.Sprintf("📦 Resources in %s:", scopeStr)) + "\n\n")
		for i, rt := range m.resourceTypes {
			prefix := "  "
			style := normalStyle
			if i == m.cursor {
				prefix = "▶ "
				style = selectedStyle
			}
			info := rt.GetResourceInfo()
			content.WriteString(prefix + style.Render(fmt.Sprintf("%s %s", info.Icon, info.Name)) + "\n")
		}
		
	case DetailView:
		content.WriteString(m.renderResourceDetails())
		
	case LogView:
		content.WriteString(m.renderLogs())
		
	case EventView:
		content.WriteString(m.renderEvents())
		
	case MultiFrameView:
		content.WriteString(m.renderMultiFrameView())
	}
	
	// Help section with feature options and commands - using darker dividers
	dividerWidth := m.width
	if dividerWidth == 0 {
		dividerWidth = 80 // fallback width
	}
	dividerStyle := lipgloss.NewStyle().Foreground(colors.Muted)
	content.WriteString("\n" + dividerStyle.Render(strings.Repeat("─", dividerWidth)) + "\n")
	
	// Show available features and commands
	featuresText := m.buildFeaturesText()
	if featuresText != "" {
		content.WriteString(featuresText + "\n")
		content.WriteString(dividerStyle.Render(strings.Repeat("─", dividerWidth)) + "\n")
	}
	
	helpText := m.buildHelpText()
	
	// Center the help text
	if m.width > 0 {
		helpPadding := (m.width - len(helpText)) / 2
		if helpPadding > 0 {
			helpText = strings.Repeat(" ", helpPadding) + helpText
		}
	}
	content.WriteString(helpStyle.Render(helpText))
	
	return content.String()
}

// buildHeader creates the application header with context information
func (m Model) buildHeader() string {
	title := "🚀 k8sGo - Kubernetes Resource Monitor v2.0"
	
	var context []string
	if m.selectedScope == ClusterScoped {
		context = append(context, "Scope: Cluster")
	} else if m.selectedNamespace != "" {
		context = append(context, fmt.Sprintf("Namespace: %s", m.selectedNamespace))
	}
	
	if m.selectedResource.String() != "Unknown" {
		info := m.selectedResource.GetResourceInfo()
		context = append(context, fmt.Sprintf("Resource: %s %s", info.Icon, info.Name))
	}
	
	if m.autoRefresh {
		context = append(context, "Auto-refresh: ON")
	}
	
	// Create the full header text (lipgloss will handle centering and width)
	if len(context) > 0 {
		return title + " | " + strings.Join(context, " | ")
	}
	return title
}

// renderResourceDetails creates the detailed resource view with enhanced error display
func (m Model) renderResourceDetails() string {
	if len(m.resources) == 0 {
		scopeStr := "cluster"
		if m.selectedScope == NamespaceScoped {
			scopeStr = fmt.Sprintf("namespace '%s'", m.selectedNamespace)
		}
		return fmt.Sprintf("No %s found in %s", 
			strings.ToLower(m.selectedResource.String()), scopeStr)
	}
	
	var content strings.Builder
	scopeStr := "cluster"
	if m.selectedScope == NamespaceScoped {
		scopeStr = fmt.Sprintf("namespace '%s'", m.selectedNamespace)
	}
	
	info := m.selectedResource.GetResourceInfo()
	
	// Count resources with errors/warnings for summary
	errorCount := 0
	warningCount := 0
	for _, resource := range m.resources {
		if len(resource.Errors) > 0 {
			errorCount++
		}
		if len(resource.Warnings) > 0 {
			warningCount++
		}
	}
	
	// Header with error/warning summary
	headerText := fmt.Sprintf("📋 %s %s in %s (%d items)", 
		info.Icon, m.selectedResource.String(), scopeStr, len(m.resources))
	
	if errorCount > 0 || warningCount > 0 {
		headerText += fmt.Sprintf(" - ")
		if errorCount > 0 {
			headerText += lipgloss.NewStyle().Foreground(colors.Error).Render(fmt.Sprintf("%d errors", errorCount))
		}
		if errorCount > 0 && warningCount > 0 {
			headerText += ", "
		}
		if warningCount > 0 {
			headerText += lipgloss.NewStyle().Foreground(colors.Warning).Render(fmt.Sprintf("%d warnings", warningCount))
		}
	}
	
	successStyle := lipgloss.NewStyle().Foreground(colors.Success).Bold(true)
	normalStyle := lipgloss.NewStyle().Foreground(colors.Text)
	selectedStyle := lipgloss.NewStyle().Bold(true).Foreground(colors.Background).Background(colors.Secondary).Padding(0, 1)
	
	content.WriteString(successStyle.Render(headerText) + "\n\n")
	
	// Table header
	var headerFormat, rowFormat string
	if m.selectedScope == NamespaceScoped {
		headerFormat = "%-30s %-20s %-15s %-10s"
		rowFormat = "%-30s %-20s %-15s %-10s"
		header := fmt.Sprintf(headerFormat, "NAME", "NAMESPACE", "STATUS", "AGE")
		content.WriteString(normalStyle.Render(header) + "\n")
	} else {
		headerFormat = "%-30s %-15s %-10s"
		rowFormat = "%-30s %-15s %-10s"
		header := fmt.Sprintf(headerFormat, "NAME", "STATUS", "AGE")
		content.WriteString(normalStyle.Render(header) + "\n")
	}
	dividerWidth := m.width
	if dividerWidth == 0 {
		dividerWidth = 70 // fallback width
	}
	dividerStyle := lipgloss.NewStyle().Foreground(colors.Muted)
	content.WriteString(dividerStyle.Render(strings.Repeat("─", dividerWidth)) + "\n")
	
	// Table rows with color-coded status
	for i, resource := range m.resources {
		var row string
		if m.selectedScope == NamespaceScoped {
			row = fmt.Sprintf(rowFormat, 
				truncateString(resource.Name, 30),
				truncateString(resource.Namespace, 20),
				truncateString(resource.Status, 15),
				resource.Age)
		} else {
			row = fmt.Sprintf(rowFormat, 
				truncateString(resource.Name, 30),
				truncateString(resource.Status, 15),
				resource.Age)
		}
		
		// Choose style based on resource health
		var style lipgloss.Style
		if i == m.cursor {
			style = selectedStyle
		} else if len(resource.Errors) > 0 {
			style = lipgloss.NewStyle().Foreground(colors.Error)
		} else if len(resource.Warnings) > 0 {
			style = lipgloss.NewStyle().Foreground(colors.Warning)
		} else {
			style = normalStyle
		}
		
		content.WriteString(style.Render(row) + "\n")
		
		// Show errors and warnings for selected resource
		if i == m.cursor {
			if len(resource.Errors) > 0 {
				for _, err := range resource.Errors {
					content.WriteString(lipgloss.NewStyle().Foreground(colors.Error).Render("  ❌ " + err) + "\n")
				}
			}
			if len(resource.Warnings) > 0 {
				for _, warning := range resource.Warnings {
					content.WriteString(lipgloss.NewStyle().Foreground(colors.Warning).Render("  ⚠️  " + warning) + "\n")
				}
			}
			
			// Show additional details
			if len(resource.Details) > 0 {
				content.WriteString(lipgloss.NewStyle().Foreground(colors.Info).Render("  📝 Details:") + "\n")
				for key, value := range resource.Details {
					content.WriteString(fmt.Sprintf("     %s: %s\n", key, value))
				}
			}
		}
	}
	
	return content.String()
}

// buildFeaturesText shows available features and actions for current view
func (m Model) buildFeaturesText() string {
	var features []string
	featureStyle := lipgloss.NewStyle().Foreground(colors.Info).Bold(true)
	actionStyle := lipgloss.NewStyle().Foreground(colors.Success)
	
	switch m.currentView {
	case DetailView:
		if len(m.resources) > 0 && m.cursor < len(m.resources) {
			selectedResource := &m.resources[m.cursor]
			info := selectedResource.ResourceType.GetResourceInfo()
			
			features = append(features, featureStyle.Render("Available Actions:"))
			
			if info.SupportsLogs {
				features = append(features, actionStyle.Render("  📜 Press 'l' - View logs for selected resource"))
			}
			if info.SupportsEvents {
				features = append(features, actionStyle.Render("  📢 Press 'e' - View events for selected resource"))
			}
			features = append(features, actionStyle.Render("  🔲 Press 'm' - Switch to multi-frame view"))
			features = append(features, actionStyle.Render("  🔄 Press 'r' - Refresh resource list"))
			features = append(features, actionStyle.Render("  ⚡ Press 'a' - Toggle auto-refresh"))
		}
	case MultiFrameView:
		features = append(features, featureStyle.Render("Multi-Frame Features:"))
		features = append(features, actionStyle.Render("  🔄 Press 'tab' - Switch between Resource, Log, and Event frames"))
		features = append(features, actionStyle.Render("  📦 Resource Frame - Navigate and select resources"))
		features = append(features, actionStyle.Render("  📜 Log Frame - View real-time logs"))
		features = append(features, actionStyle.Render("  📢 Event Frame - View Kubernetes events"))
		features = append(features, actionStyle.Render("  🔄 Press 'r' - Refresh current frame"))
	case KubernetesContextView:
		features = append(features, featureStyle.Render("Context Selection:"))
		features = append(features, actionStyle.Render("  🔧 Select cluster context to connect"))
		features = append(features, actionStyle.Render("  ✅ Current active context is marked"))
		features = append(features, actionStyle.Render("  🚀 Auto-switches kubectl context on selection"))
	case ResourceView:
		features = append(features, featureStyle.Render("Resource Types Available:"))
		if m.isOpenShift {
			features = append(features, actionStyle.Render("  🔗 OpenShift Routes, DeploymentConfigs, Projects"))
		}
		features = append(features, actionStyle.Render("  🚪 Gateway API resources (Gateways, HTTPRoutes)"))
		features = append(features, actionStyle.Render("  🛡️ Network Policies, Autoscalers"))
		features = append(features, actionStyle.Render("  🐳 Pods, Services, Deployments, and more"))
	}
	
	if len(features) == 0 {
		return ""
	}
	
	// Center the features text
	var centeredFeatures []string
	for _, feature := range features {
		if m.width > 0 {
			// Remove styles for length calculation
			plainText := lipgloss.NewStyle().Render(feature)
			padding := (m.width - len(plainText)) / 2
			if padding > 0 {
				feature = strings.Repeat(" ", padding) + feature
			}
		}
		centeredFeatures = append(centeredFeatures, feature)
	}
	
	return strings.Join(centeredFeatures, "\n")
}

// buildHelpText creates context-sensitive help information
func (m Model) buildHelpText() string {
	var help []string
	
	switch m.currentView {
	case KubernetesContextView:
		help = []string{
			"↑/k: up", "↓/j: down", "enter/space: select", "q: quit",
		}
	case LogView:
		help = []string{
			"↑/k: scroll up", "↓/j: scroll down", "esc: back", 
			"r: refresh logs", "a: toggle auto-refresh", "q: quit",
		}
	case EventView:
		help = []string{
			"↑/k: scroll up", "↓/j: scroll down", "esc: back", 
			"r: refresh events", "a: toggle auto-refresh", "q: quit",
		}
	case DetailView:
		if len(m.resources) > 0 && m.cursor < len(m.resources) {
			selectedResource := &m.resources[m.cursor]
			info := selectedResource.ResourceType.GetResourceInfo()
			helpItems := []string{"↑/k: up", "↓/j: down", "enter: select", "esc: back", "r: refresh", "a: toggle auto-refresh", "m: multi-frame", "q: quit"}
			if info.SupportsLogs {
				helpItems = append(helpItems[:3], append([]string{"l: view logs"}, helpItems[3:]...)...)
			}
			if info.SupportsEvents {
				helpItems = append(helpItems[:4], append([]string{"e: view events"}, helpItems[4:]...)...)
			}
			help = helpItems
		} else {
			help = []string{
				"↑/k: up", "↓/j: down", "enter: select", "esc: back",
				"r: refresh", "a: toggle auto-refresh", "m: multi-frame", "q: quit",
			}
		}
	case MultiFrameView:
		help = []string{
			"↑/k: up", "↓/j: down", "tab: switch frame", "esc: back", 
			"r: refresh", "a: toggle auto-refresh", "q: quit",
		}
	default:
		help = []string{
			"↑/k: up", "↓/j: down", "enter/space: select", "esc: back",
			"r: refresh", "a: toggle auto-refresh", "q: quit",
		}
	}
	
	return strings.Join(help, " | ")
}

// Placeholder functions for resource loading - implement based on your needs
func (m Model) loadNodes() ([]K8sResource, error) {
	nodes, err := m.clientset.CoreV1().Nodes().List(m.ctx, metav1.ListOptions{})
	if err != nil {
		return nil, err
	}
	
	var resources []K8sResource
	now := time.Now()
	
	for _, node := range nodes.Items {
		age := humanAge(now.Sub(node.CreationTimestamp.Time))
		
		status := "Ready"
		var nodeErrors []string
		var nodeWarnings []string
		
		// Check node conditions
		for _, condition := range node.Status.Conditions {
			if condition.Type == corev1.NodeReady {
				if condition.Status != corev1.ConditionTrue {
					status = "NotReady"
					nodeErrors = append(nodeErrors, fmt.Sprintf("Node is not ready: %s", condition.Message))
				}
			} else if condition.Status == corev1.ConditionTrue {
				switch condition.Type {
				case corev1.NodeMemoryPressure:
					nodeWarnings = append(nodeWarnings, "Memory pressure detected")
				case corev1.NodeDiskPressure:
					nodeWarnings = append(nodeWarnings, "Disk pressure detected")
				case corev1.NodePIDPressure:
					nodeWarnings = append(nodeWarnings, "PID pressure detected")
				case corev1.NodeNetworkUnavailable:
					nodeErrors = append(nodeErrors, "Network unavailable")
				}
			}
		}
		
		// Get resource usage info
		allocatable := node.Status.Allocatable
		cpu := allocatable[corev1.ResourceCPU]
		memory := allocatable[corev1.ResourceMemory]
		
		resource := K8sResource{
			Name:         node.Name,
			Namespace:    "", // Nodes are cluster-scoped
			Status:       status,
			Age:          age,
			ResourceType: NodesResource,
			Errors:       nodeErrors,
			Warnings:     nodeWarnings,
			Details: map[string]string{
				"CPU":            cpu.String(),
				"Memory":         memory.String(),
				"OS":             node.Status.NodeInfo.OperatingSystem,
				"Kernel":         node.Status.NodeInfo.KernelVersion,
				"Container Runtime": node.Status.NodeInfo.ContainerRuntimeVersion,
			},
		}
		resources = append(resources, resource)
	}
	
	return resources, nil
}

// Additional placeholder resource loading functions
func (m Model) loadPods() ([]K8sResource, error) {
	pods, err := m.clientset.CoreV1().Pods(m.selectedNamespace).List(m.ctx, metav1.ListOptions{})
	if err != nil {
		return nil, err
	}
	
	var resources []K8sResource
	now := time.Now()
	
	for _, pod := range pods.Items {
		age := humanAge(now.Sub(pod.CreationTimestamp.Time))
		
		status := string(pod.Status.Phase)
		var podErrors []string
		var podWarnings []string
		
		// Check pod conditions and container statuses
		for _, condition := range pod.Status.Conditions {
			if condition.Status == corev1.ConditionFalse {
				switch condition.Type {
				case corev1.PodReady:
					podErrors = append(podErrors, "Pod is not ready")
				case corev1.PodScheduled:
					if condition.Reason == "Unschedulable" {
						podErrors = append(podErrors, "Pod cannot be scheduled")
					}
				}
			}
		}
		
		// Check container statuses
		restartCount := int32(0)
		for _, containerStatus := range pod.Status.ContainerStatuses {
			restartCount += containerStatus.RestartCount
			if containerStatus.State.Waiting != nil {
				podWarnings = append(podWarnings, fmt.Sprintf("Container %s is waiting: %s", 
					containerStatus.Name, containerStatus.State.Waiting.Reason))
			}
			if containerStatus.State.Terminated != nil && containerStatus.State.Terminated.ExitCode != 0 {
				podErrors = append(podErrors, fmt.Sprintf("Container %s terminated with exit code %d", 
					containerStatus.Name, containerStatus.State.Terminated.ExitCode))
			}
		}
		
		if restartCount > 0 {
			podWarnings = append(podWarnings, fmt.Sprintf("Container(s) restarted %d times", restartCount))
		}
		
		ready := 0
		for _, containerStatus := range pod.Status.ContainerStatuses {
			if containerStatus.Ready {
				ready++
			}
		}
		
		resource := K8sResource{
			Name:         pod.Name,
			Namespace:    pod.Namespace,
			Status:       status,
			Age:          age,
			ResourceType: PodsResource,
			Errors:       podErrors,
			Warnings:     podWarnings,
			Details: map[string]string{
				"Ready":         fmt.Sprintf("%d/%d", ready, len(pod.Spec.Containers)),
				"Restarts":      strconv.FormatInt(int64(restartCount), 10),
				"Node":          pod.Spec.NodeName,
				"Pod IP":        pod.Status.PodIP,
			},
		}
		resources = append(resources, resource)
	}
	
	return resources, nil
}

// Add more resource loading functions as needed...
func (m Model) loadPersistentVolumes() ([]K8sResource, error) { return []K8sResource{}, nil }
func (m Model) loadStorageClasses() ([]K8sResource, error) { return []K8sResource{}, nil }
func (m Model) loadClusterRoles() ([]K8sResource, error) { return []K8sResource{}, nil }
func (m Model) loadServices() ([]K8sResource, error) {
	services, err := m.clientset.CoreV1().Services(m.selectedNamespace).List(m.ctx, metav1.ListOptions{})
	if err != nil {
		return nil, err
	}
	
	var resources []K8sResource
	now := time.Now()
	
	for _, service := range services.Items {
		age := humanAge(now.Sub(service.CreationTimestamp.Time))
		
		status := "Active"
		clusterIP := service.Spec.ClusterIP
		if clusterIP == "" {
			clusterIP = "None"
		}
		
		var ports []string
		for _, port := range service.Spec.Ports {
			if port.NodePort != 0 {
				ports = append(ports, fmt.Sprintf("%d:%d/%s", port.Port, port.NodePort, port.Protocol))
			} else {
				ports = append(ports, fmt.Sprintf("%d/%s", port.Port, port.Protocol))
			}
		}
		portsStr := strings.Join(ports, ",")
		if portsStr == "" {
			portsStr = "None"
		}
		
		resource := K8sResource{
			Name:         service.Name,
			Namespace:    service.Namespace,
			Status:       status,
			Age:          age,
			ResourceType: ServicesResource,
			Details: map[string]string{
				"Type":       string(service.Spec.Type),
				"Cluster-IP": clusterIP,
				"Ports":      portsStr,
				"Selector":   fmt.Sprintf("%v", service.Spec.Selector),
			},
		}
		resources = append(resources, resource)
	}
	
	return resources, nil
}
func (m Model) loadDeployments() ([]K8sResource, error) {
	deployments, err := m.clientset.AppsV1().Deployments(m.selectedNamespace).List(m.ctx, metav1.ListOptions{})
	if err != nil {
		return nil, err
	}
	
	var resources []K8sResource
	now := time.Now()
	
	for _, deployment := range deployments.Items {
		age := humanAge(now.Sub(deployment.CreationTimestamp.Time))
		
		status := "Available"
		var deploymentErrors []string
		var deploymentWarnings []string
		
		// Check deployment conditions
		for _, condition := range deployment.Status.Conditions {
			if condition.Type == "Available" && condition.Status != "True" {
				status = "NotAvailable"
				deploymentErrors = append(deploymentErrors, "Deployment not available")
			}
			if condition.Type == "Progressing" && condition.Status != "True" {
				deploymentWarnings = append(deploymentWarnings, "Deployment not progressing")
			}
		}
		
		// Check replica status
		ready := fmt.Sprintf("%d/%d", deployment.Status.ReadyReplicas, deployment.Status.Replicas)
		if deployment.Status.ReadyReplicas != deployment.Status.Replicas {
			status = "NotReady"
			deploymentWarnings = append(deploymentWarnings, "Not all replicas are ready")
		}
		
		resource := K8sResource{
			Name:         deployment.Name,
			Namespace:    deployment.Namespace,
			Status:       status,
			Age:          age,
			ResourceType: DeploymentsResource,
			Errors:       deploymentErrors,
			Warnings:     deploymentWarnings,
			Details: map[string]string{
				"Ready":           ready,
				"Up-to-date":      fmt.Sprintf("%d", deployment.Status.UpdatedReplicas),
				"Available":       fmt.Sprintf("%d", deployment.Status.AvailableReplicas),
				"Strategy":        string(deployment.Spec.Strategy.Type),
			},
		}
		resources = append(resources, resource)
	}
	
	return resources, nil
}
func (m Model) loadConfigMaps() ([]K8sResource, error) { return []K8sResource{}, nil }
func (m Model) loadSecrets() ([]K8sResource, error) { return []K8sResource{}, nil }
func (m Model) loadIngress() ([]K8sResource, error) { return []K8sResource{}, nil }
func (m Model) loadPersistentVolumeClaims() ([]K8sResource, error) { return []K8sResource{}, nil }
func (m Model) loadReplicaSets() ([]K8sResource, error) { return []K8sResource{}, nil }
func (m Model) loadDaemonSets() ([]K8sResource, error) { return []K8sResource{}, nil }
func (m Model) loadStatefulSets() ([]K8sResource, error) { return []K8sResource{}, nil }
func (m Model) loadJobs() ([]K8sResource, error) { return []K8sResource{}, nil }
func (m Model) loadCronJobs() ([]K8sResource, error) { return []K8sResource{}, nil }
func (m Model) loadEventsResource() ([]K8sResource, error) { return []K8sResource{}, nil }

// Additional OpenShift resource loading functions
func (m Model) loadBuildConfigs() ([]K8sResource, error) { return []K8sResource{}, nil }
func (m Model) loadBuilds() ([]K8sResource, error) { return []K8sResource{}, nil }
func (m Model) loadImageStreams() ([]K8sResource, error) { return []K8sResource{}, nil }

// Gateway API resource loading functions
func (m Model) loadGateways() ([]K8sResource, error) { return []K8sResource{}, nil }
func (m Model) loadHTTPRoutes() ([]K8sResource, error) { return []K8sResource{}, nil }
func (m Model) loadGatewayClasses() ([]K8sResource, error) { return []K8sResource{}, nil }

// Additional resource loading functions
func (m Model) loadNetworkPolicies() ([]K8sResource, error) { return []K8sResource{}, nil }
func (m Model) loadHorizontalPodAutoscalers() ([]K8sResource, error) { return []K8sResource{}, nil }
func (m Model) loadVerticalPodAutoscalers() ([]K8sResource, error) { return []K8sResource{}, nil }

// OpenShift-specific resource loading functions
func (m Model) loadRoutes() ([]K8sResource, error) {
	if m.routeClient == nil {
		return nil, fmt.Errorf("OpenShift Route client not available")
	}
	
	routes, err := m.routeClient.RouteV1().Routes(m.selectedNamespace).List(m.ctx, metav1.ListOptions{})
	if err != nil {
		return nil, err
	}
	
	var resources []K8sResource
	now := time.Now()
	
	for _, route := range routes.Items {
		age := humanAge(now.Sub(route.CreationTimestamp.Time))
		
		status := "Active"
		var routeErrors []string
		var routeWarnings []string
		
		// Check route conditions
		for _, condition := range route.Status.Ingress {
			for _, cond := range condition.Conditions {
				if cond.Type == routev1.RouteAdmitted && cond.Status != corev1.ConditionTrue {
					status = "NotAdmitted"
					routeErrors = append(routeErrors, fmt.Sprintf("Route not admitted: %s", cond.Message))
				}
			}
		}
		
		// Extract host
		host := route.Spec.Host
		if host == "" && len(route.Status.Ingress) > 0 {
			host = route.Status.Ingress[0].Host
		}
		
		resource := K8sResource{
			Name:         route.Name,
			Namespace:    route.Namespace,
			Status:       status,
			Age:          age,
			ResourceType: RoutesResource,
			Errors:       routeErrors,
			Warnings:     routeWarnings,
			Details: map[string]string{
				"Host":    host,
				"Path":    route.Spec.Path,
				"Service": route.Spec.To.Name,
				"Port":    fmt.Sprintf("%v", route.Spec.Port),
			},
		}
		resources = append(resources, resource)
	}
	
	return resources, nil
}

func (m Model) loadDeploymentConfigs() ([]K8sResource, error) {
	if m.openshiftAppsClient == nil {
		return nil, fmt.Errorf("OpenShift Apps client not available")
	}
	
	dcs, err := m.openshiftAppsClient.AppsV1().DeploymentConfigs(m.selectedNamespace).List(m.ctx, metav1.ListOptions{})
	if err != nil {
		return nil, err
	}
	
	var resources []K8sResource
	now := time.Now()
	
	for _, dc := range dcs.Items {
		age := humanAge(now.Sub(dc.CreationTimestamp.Time))
		
		ready := fmt.Sprintf("%d/%d", dc.Status.ReadyReplicas, dc.Status.Replicas)
		status := "Running"
		var dcErrors []string
		
		if dc.Status.ReadyReplicas != dc.Status.Replicas {
			status = "NotReady"
			dcErrors = append(dcErrors, "Not all replicas are ready")
		}
		
		resource := K8sResource{
			Name:         dc.Name,
			Namespace:    dc.Namespace,
			Status:       status,
			Age:          age,
			ResourceType: DeploymentConfigsResource,
			Errors:       dcErrors,
			Details: map[string]string{
				"Ready":           ready,
				"Available":       fmt.Sprintf("%d", dc.Status.AvailableReplicas),
				"Latest Version":  fmt.Sprintf("%d", dc.Status.LatestVersion),
				"Observed Gen":    fmt.Sprintf("%d", dc.Status.ObservedGeneration),
			},
		}
		resources = append(resources, resource)
	}
	
	return resources, nil
}

func (m Model) loadProjects() ([]K8sResource, error) {
	if m.projectClient == nil {
		return nil, fmt.Errorf("OpenShift Project client not available")
	}
	
	projects, err := m.projectClient.ProjectV1().Projects().List(m.ctx, metav1.ListOptions{})
	if err != nil {
		return nil, err
	}
	
	var resources []K8sResource
	now := time.Now()
	
	for _, project := range projects.Items {
		age := humanAge(now.Sub(project.CreationTimestamp.Time))
		
		status := string(project.Status.Phase)
		var projectErrors []string
		
		if project.Status.Phase != corev1.NamespaceActive {
			projectErrors = append(projectErrors, fmt.Sprintf("Project is in %s state", project.Status.Phase))
		}
		
		resource := K8sResource{
			Name:         project.Name,
			Namespace:    "", // Projects are cluster-scoped
			Status:       status,
			Age:          age,
			ResourceType: ProjectsResource,
			Errors:       projectErrors,
			Details: map[string]string{
				"Display Name": project.Annotations["openshift.io/display-name"],
				"Description":  project.Annotations["openshift.io/description"],
				"Requester":    project.Annotations["openshift.io/requester"],
			},
		}
		resources = append(resources, resource)
	}
	
	return resources, nil
}

// renderMultiFrameView creates the multi-frame layout with resources, logs, and events
func (m Model) renderMultiFrameView() string {
	if m.frameWidth == 0 {
		m.frameWidth = m.width / 3
		m.frameHeight = m.height - 10
	}
	
	var content strings.Builder
	
	// Frame headers
	headerStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(colors.Text).
		Background(colors.Primary).
		Padding(0, 1).
		Width(m.frameWidth - 2)
	
	// Left frame: Resources
	resourceHeaderText := "📦 Resources"
	if m.currentFrame == ResourceFrame {
		resourceHeaderText = "▶ " + resourceHeaderText
	}
	resourceHeader := headerStyle.Render(resourceHeaderText)
	
	// Middle frame: Logs
	logHeaderText := "📜 Logs"
	if m.selectedK8sResource != nil {
		logHeaderText = fmt.Sprintf("📜 Logs - %s", m.selectedK8sResource.Name)
	}
	if m.currentFrame == LogFrame {
		logHeaderText = "▶ " + logHeaderText
	}
	logHeader := headerStyle.Render(logHeaderText)
	
	// Right frame: Events
	eventHeaderText := "📢 Events"
	if m.selectedK8sResource != nil {
		eventHeaderText = fmt.Sprintf("📢 Events - %s", m.selectedK8sResource.Name)
	}
	if m.currentFrame == EventFrame {
		eventHeaderText = "▶ " + eventHeaderText
	}
	eventHeader := headerStyle.Render(eventHeaderText)
	
	// Render frame headers
	content.WriteString(lipgloss.JoinHorizontal(lipgloss.Top, resourceHeader, " ", logHeader, " ", eventHeader) + "\n")
	
	// Frame content styles
	frameStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(colors.Secondary).
		Width(m.frameWidth - 4).
		Height(m.frameHeight - 4).
		Padding(1)
	
	activeFrameStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(colors.Primary).
		Width(m.frameWidth - 4).
		Height(m.frameHeight - 4).
		Padding(1)
	
	// Left frame: Resources content
	var resourceContent string
	if len(m.resources) == 0 {
		resourceContent = "No resources found"
	} else {
		var resourceLines strings.Builder
		for i, resource := range m.resources {
			prefix := "  "
			style := lipgloss.NewStyle().Foreground(colors.Text)
			if i == m.cursor && m.currentFrame == ResourceFrame {
				prefix = "▶ "
				style = lipgloss.NewStyle().Bold(true).Foreground(colors.Background).Background(colors.Secondary).Padding(0, 1)
			}
			
			// Resource status color
			if len(resource.Errors) > 0 {
				style = style.Foreground(colors.Error)
			} else if len(resource.Warnings) > 0 {
				style = style.Foreground(colors.Warning)
			}
			
			resourceLines.WriteString(prefix + style.Render(fmt.Sprintf("%-25s %s", 
				truncateString(resource.Name, 25), resource.Status)) + "\n")
		}
		resourceContent = resourceLines.String()
	}
	
	// Middle frame: Logs content
	var logContent string
	if m.selectedK8sResource == nil {
		logContent = "Select a resource to view logs"
	} else if len(m.logEntries) == 0 {
		if m.loading {
			logContent = "Loading logs..."
		} else {
			logContent = "No logs available for this resource"
		}
	} else {
		var logLines strings.Builder
		startIdx := m.logScrollOffset
		endIdx := min(startIdx + m.frameHeight - 6, len(m.logEntries))
		
		for i := startIdx; i < endIdx; i++ {
			entry := m.logEntries[i]
			timestamp := entry.Timestamp.Format("15:04:05")
			logLines.WriteString(fmt.Sprintf("[%s] %s\n", timestamp, entry.Message))
		}
		logContent = logLines.String()
	}
	
	// Right frame: Events content
	var eventContent string
	if m.selectedK8sResource == nil {
		eventContent = "Select a resource to view events"
	} else if len(m.eventEntries) == 0 {
		if m.loading {
			eventContent = "Loading events..."
		} else {
			eventContent = "No events found for this resource"
		}
	} else {
		var eventLines strings.Builder
		startIdx := m.eventScrollOffset
		endIdx := min(startIdx + m.frameHeight - 6, len(m.eventEntries))
		
		for i := startIdx; i < endIdx; i++ {
			event := m.eventEntries[i]
			timestamp := event.Timestamp.Format("15:04:05")
			eventLines.WriteString(fmt.Sprintf("[%s] %s: %s\n", timestamp, event.Reason, event.Message))
		}
		eventContent = eventLines.String()
	}
	
	// Apply frame styles
	var leftFrame, middleFrame, rightFrame string
	if m.currentFrame == ResourceFrame {
		leftFrame = activeFrameStyle.Render(resourceContent)
		middleFrame = frameStyle.Render(logContent)
		rightFrame = frameStyle.Render(eventContent)
	} else if m.currentFrame == LogFrame {
		leftFrame = frameStyle.Render(resourceContent)
		middleFrame = activeFrameStyle.Render(logContent)
		rightFrame = frameStyle.Render(eventContent)
	} else {
		leftFrame = frameStyle.Render(resourceContent)
		middleFrame = frameStyle.Render(logContent)
		rightFrame = activeFrameStyle.Render(eventContent)
	}
	
	// Join frames horizontally
	content.WriteString(lipgloss.JoinHorizontal(lipgloss.Top, leftFrame, " ", middleFrame, " ", rightFrame))
	
	return content.String()
}

// Log and event loading functions
func (m Model) loadLogs() tea.Cmd { 
	return func() tea.Msg {
		// Placeholder - implement actual log loading
		var logs []LogEntry
		
		if m.selectedK8sResource != nil && m.selectedK8sResource.ResourceType == PodsResource {
			// For pods, we could load actual logs
			logs = []LogEntry{
				{Timestamp: time.Now().Add(-5 * time.Minute), Message: "Pod started successfully", Container: "main", Level: "INFO"},
				{Timestamp: time.Now().Add(-3 * time.Minute), Message: "Application initialized", Container: "main", Level: "INFO"},
				{Timestamp: time.Now().Add(-1 * time.Minute), Message: "Ready to accept connections", Container: "main", Level: "INFO"},
			}
		}
		
		return logsLoadedMsg{logs: logs, err: nil}
	}
}

func (m Model) loadEventsCmd() tea.Cmd {
	return func() tea.Msg {
		var events []EventEntry
		
		if m.selectedK8sResource == nil {
			return eventsLoadedMsg{events: events, err: nil}
		}
		
		// Load events related to the selected resource
		eventList, err := m.clientset.CoreV1().Events(m.selectedNamespace).List(m.ctx, metav1.ListOptions{
			FieldSelector: fmt.Sprintf("involvedObject.name=%s", m.selectedK8sResource.Name),
		})
		
		if err != nil {
			return eventsLoadedMsg{events: events, err: err}
		}
		
		// Convert Kubernetes events to our EventEntry format
		for _, event := range eventList.Items {
			entry := EventEntry{
				Timestamp: event.CreationTimestamp.Time,
				Type:      event.Type,
				Reason:    event.Reason,
				Message:   event.Message,
				Source:    event.Source.Component,
				Count:     event.Count,
			}
			events = append(events, entry)
		}
		
		// Sort events by timestamp (newest first)
		sort.Slice(events, func(i, j int) bool {
			return events[i].Timestamp.After(events[j].Timestamp)
		})
		
		return eventsLoadedMsg{events: events, err: nil}
	}
}
func (m Model) renderLogs() string { return "Logs view not implemented yet" }
func (m Model) renderEvents() string {
	var content strings.Builder
	
	// Header
	headerText := "📢 Events"
	if m.selectedK8sResource != nil {
		headerText = fmt.Sprintf("📢 Events - %s", m.selectedK8sResource.Name)
	}
	
	headerStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(colors.Success).
		Margin(0, 0, 1, 0)
	
	content.WriteString(headerStyle.Render(headerText) + "\n")
	
	if m.selectedK8sResource == nil {
		content.WriteString("No resource selected for events\n")
		return content.String()
	}
	
	if len(m.eventEntries) == 0 {
		if m.loading {
			content.WriteString("Loading events...\n")
		} else {
			content.WriteString("No events found for this resource\n")
		}
		return content.String()
	}
	
	// Calculate visible range based on scroll offset
	maxVisible := m.height - 10
	if maxVisible < 5 {
		maxVisible = 5
	}
	
	startIdx := m.eventScrollOffset
	endIdx := min(startIdx + maxVisible, len(m.eventEntries))
	
	// Render events
	for i := startIdx; i < endIdx; i++ {
		event := m.eventEntries[i]
		
		// Format timestamp
		timestamp := event.Timestamp.Format("15:04:05")
		
		// Choose color based on event type
		var eventStyle lipgloss.Style
		switch event.Type {
		case "Warning":
			eventStyle = lipgloss.NewStyle().Foreground(colors.Warning)
		case "Normal":
			eventStyle = lipgloss.NewStyle().Foreground(colors.Success)
		default:
			eventStyle = lipgloss.NewStyle().Foreground(colors.Info)
		}
		
		// Format event entry
		eventLine := fmt.Sprintf("[%s] %s: %s", timestamp, event.Reason, event.Message)
		if event.Count > 1 {
			eventLine = fmt.Sprintf("[%s] %s: %s (x%d)", timestamp, event.Reason, event.Message, event.Count)
		}
		
		content.WriteString(eventStyle.Render(eventLine) + "\n")
	}
	
	// Show scroll indicator if there are more events
	if len(m.eventEntries) > maxVisible {
		scrollInfo := fmt.Sprintf("Showing %d-%d of %d events", startIdx+1, endIdx, len(m.eventEntries))
		scrollStyle := lipgloss.NewStyle().Foreground(colors.Muted).Italic(true)
		content.WriteString("\n" + scrollStyle.Render(scrollInfo) + "\n")
	}
	
	return content.String()
}

// Utility functions
func truncateString(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen-3] + "..."
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// humanAge formats duration into human-readable age string
func humanAge(d time.Duration) string {
	s := int(d.Seconds())
	m := s / 60
	h := m / 60
	days := h / 24

	switch {
	case days > 0:
		return fmt.Sprintf("%dd", days)
	case h > 0:
		return fmt.Sprintf("%dh", h)
	case m > 0:
		return fmt.Sprintf("%dm", m)
	default:
		return fmt.Sprintf("%ds", s)
	}
}

// initializeKubernetesClient creates a Kubernetes client from kubeconfig
func initializeKubernetesClient() (*kubernetes.Clientset, error) {
	config, err := getKubernetesConfig()
	if err != nil {
		return nil, err
	}

	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		return nil, fmt.Errorf("failed to create Kubernetes client: %v", err)
	}

	return clientset, nil
}

// getKubernetesConfig loads Kubernetes configuration from kubeconfig or in-cluster config
func getKubernetesConfig() (*rest.Config, error) {
	// Try in-cluster config first
	config, err := rest.InClusterConfig()
	if err == nil {
		return config, nil
	}

	// Fall back to kubeconfig
	var kubeconfig string
	if home := homedir.HomeDir(); home != "" {
		kubeconfig = filepath.Join(home, ".kube", "config")
	}
	
	if _, err := os.Stat(kubeconfig); err == nil {
		config, err := clientcmd.BuildConfigFromFlags("", kubeconfig)
		if err != nil {
			return nil, fmt.Errorf("failed to build config from kubeconfig: %v", err)
		}
		return config, nil
	}

	return nil, fmt.Errorf("unable to load Kubernetes configuration")
}

// reinitializeClients creates new clients after context switch
func (m Model) reinitializeClients(contextName string) tea.Cmd {
	return func() tea.Msg {
		// Initialize new Kubernetes client with the switched context
		clientset, err := initializeKubernetesClient()
		if err != nil {
			return clientsReinitializedMsg{contextName: contextName, err: fmt.Errorf("failed to initialize Kubernetes client: %v", err)}
		}

		// Try to initialize OpenShift clients with the new context
		config, err := getKubernetesConfig()
		var openshiftAppsClient *openshiftclient.Clientset
		var routeClient *routeclient.Clientset  
		var projectClient *projectclient.Clientset
		var isOpenShift bool = false
		
		if err == nil {
			openshiftAppsClient, _ = openshiftclient.NewForConfig(config)
			routeClient, _ = routeclient.NewForConfig(config)
			projectClient, _ = projectclient.NewForConfig(config)
			
			// Test if we're on OpenShift by trying to list projects
			if projectClient != nil {
				_, err := projectClient.ProjectV1().Projects().List(context.Background(), metav1.ListOptions{Limit: 1})
				isOpenShift = err == nil
			}
		}

		// Return message with all the new client information
		return clientsReinitializedMsg{
			contextName:         contextName,
			clientset:          clientset,
			openshiftAppsClient: openshiftAppsClient,
			routeClient:        routeClient,
			projectClient:      projectClient,
			isOpenShift:        isOpenShift,
			err:                nil,
		}
	}
}

// main function - application entry point
func main() {
	// Set color profile for better terminal colors
	lipgloss.SetColorProfile(termenv.ANSI256)
	
	// Banner will be shown in the UI on every page
	
	// Initialize Kubernetes client
	clientset, err := initializeKubernetesClient()
	if err != nil {
		log.Fatalf("Failed to initialize Kubernetes client: %v", err)
	}
	
	// Try to initialize OpenShift clients
	config, err := getKubernetesConfig()
	var openshiftAppsClient *openshiftclient.Clientset
	var routeClient *routeclient.Clientset  
	var projectClient *projectclient.Clientset
	var isOpenShift bool = false
	
	if err == nil {
		openshiftAppsClient, _ = openshiftclient.NewForConfig(config)
		routeClient, _ = routeclient.NewForConfig(config)
		projectClient, _ = projectclient.NewForConfig(config)
		
		// Test if we're on OpenShift by trying to list projects
		if projectClient != nil {
			_, err := projectClient.ProjectV1().Projects().List(context.Background(), metav1.ListOptions{Limit: 1})
			isOpenShift = err == nil
		}
	}
	
	// Create initial model
	initialModel := Model{
		clientset:           clientset,
		ctx:                context.Background(),
		openshiftAppsClient: openshiftAppsClient,
		routeClient:         routeClient,
		projectClient:       projectClient,
		isOpenShift:         isOpenShift,
		currentView:         KubernetesContextView, // Will automatically switch to ClusterOrNamespaceView
		cursor:              0,
		viewStack:           make([]ViewType, 0),
		kubernetesContexts:  make([]string, 0),
		namespaces:          make([]string, 0),
		resourceTypes:       make([]ResourceType, 0),
		resources:           make([]K8sResource, 0),
		logEntries:          make([]LogEntry, 0),
		eventEntries:        make([]EventEntry, 0),
		loading:             true,
		autoRefresh:         true,
	}
	
	// Start the Bubble Tea program
	program := tea.NewProgram(initialModel, tea.WithAltScreen())
	
	if _, err := program.Run(); err != nil {
		log.Fatalf("Error running program: %v", err)
	}
}