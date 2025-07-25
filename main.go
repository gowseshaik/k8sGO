package main

import (
	"bufio"
	"context"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/muesli/termenv"
	// TODO: OpenShift support - temporarily disabled due to client compatibility issues
	// routev1 "github.com/openshift/api/route/v1"
	// appsv1 "github.com/openshift/api/apps/v1"
	// projectv1 "github.com/openshift/api/project/v1"
	// openshiftclient "github.com/openshift/client-go/apps/clientset/versioned"
	// routeclient "github.com/openshift/client-go/route/clientset/versioned"
	// projectclient "github.com/openshift/client-go/project/clientset/versioned"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/util/homedir"
)

// ViewType represents different view modes in the application
type ViewType int

const (
	ClusterOrNamespaceView ViewType = iota // Choose between cluster-level or namespace-level resources
	NamespaceView                          // Namespace selection view
	ResourceView                           // Resource type selection view
	DetailView                             // Detailed resource information view
	LogView                                // Resource logs view with scrolling
	EventView                              // Resource events view for errors/warnings
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
	
	// TODO: OpenShift-specific resources (temporarily disabled)
	// RoutesResource
	// DeploymentConfigsResource
	// ProjectsResource  // OpenShift projects (cluster-scoped in functionality)
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
		
		// TODO: OpenShift-specific resources (temporarily disabled)
		// RoutesResource:           {"Routes", NamespaceScoped, false, true, "🔗"},
		// DeploymentConfigsResource: {"DeploymentConfigs", NamespaceScoped, false, true, "⚙️"},
		// ProjectsResource:         {"Projects", ClusterScoped, false, true, "📁"},
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
	Primary:    lipgloss.Color("#FF6B9D"), // Pink
	Secondary:  lipgloss.Color("#4ECDC4"), // Teal
	Success:    lipgloss.Color("#45E649"), // Green
	Warning:    lipgloss.Color("#FFA726"), // Orange
	Error:      lipgloss.Color("#F44336"), // Red
	Info:       lipgloss.Color("#2196F3"), // Blue
	Accent:     lipgloss.Color("#AB47BC"), // Purple
	Background: lipgloss.Color("#1A1A2E"), // Dark blue
	Text:       lipgloss.Color("#EAEAEA"), // Light gray
	Muted:      lipgloss.Color("#16213E"), // Darker blue
}

// Model represents the application state using Bubble Tea pattern
type Model struct {
	// Kubernetes client
	clientset *kubernetes.Clientset
	ctx       context.Context
	
	// OpenShift clients (optional, will be nil if not available)
	// TODO: OpenShift support - temporarily disabled due to client compatibility issues
	// openshiftAppsClient *openshiftclient.Clientset
	// routeClient         *routeclient.Clientset
	// projectClient       *projectclient.Clientset
	isOpenShift         bool

	// Navigation state
	currentView     ViewType
	cursor          int
	viewStack       []ViewType // For navigation history
	logScrollOffset int        // For log scrolling
	eventScrollOffset int      // For event scrolling

	// Data collections
	namespaces      []string
	resourceTypes   []ResourceType
	resources       []K8sResource
	logEntries      []LogEntry
	eventEntries    []EventEntry
	
	// Current selections
	selectedNamespace   string
	selectedResource    ResourceType
	selectedK8sResource *K8sResource // For logs/events
	selectedScope       ResourceScope  // Cluster or namespace scoped
	
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
		
		// TODO: OpenShift-specific resources (temporarily disabled)
		// case RoutesResource:
		//	resources, err = m.loadRoutes()
		// case DeploymentConfigsResource:
		//	resources, err = m.loadDeploymentConfigs()
		// case ProjectsResource:
		//	resources, err = m.loadProjects()
		}
		
		return resourcesLoadedMsg{resources: resources, err: err}
	}
}

// loadLogs creates a command to asynchronously load logs for selected resource
func (m Model) loadLogs() tea.Cmd {
	return func() tea.Msg {
		if m.selectedK8sResource == nil {
			return logsLoadedMsg{err: fmt.Errorf("no resource selected")}
		}

		// Only pods support logs currently
		if m.selectedK8sResource.ResourceType != PodsResource {
			return logsLoadedMsg{err: fmt.Errorf("logs not supported for this resource type")}
		}

		// Get pod logs with better options
		podLogOpts := corev1.PodLogOptions{
			TailLines:    int64Ptr(200), // More lines
			Follow:       false,
			Timestamps:   true, // Include timestamps
			SinceSeconds: int64Ptr(3600), // Last hour
		}

		req := m.clientset.CoreV1().Pods(m.selectedK8sResource.Namespace).GetLogs(m.selectedK8sResource.Name, &podLogOpts)
		podLogs, err := req.Stream(m.ctx)
		if err != nil {
			return logsLoadedMsg{err: fmt.Errorf("failed to get logs: %v", err)}
		}
		defer podLogs.Close()

		var logs []LogEntry
		scanner := bufio.NewScanner(podLogs)
		
		for scanner.Scan() {
			line := scanner.Text()
			if line == "" {
				continue
			}
			
			// Parse timestamp if present
			timestamp := time.Now()
			message := line
			level := "INFO"
			
			// Try to parse Kubernetes log format
			if strings.Contains(line, " ") {
				parts := strings.SplitN(line, " ", 2)
				if len(parts) == 2 {
					if parsedTime, err := time.Parse(time.RFC3339, parts[0]); err == nil {
						timestamp = parsedTime
						message = parts[1]
					}
				}
			}
			
			// Detect log level
			lowerMsg := strings.ToLower(message)
			if strings.Contains(lowerMsg, "error") || strings.Contains(lowerMsg, "err") {
				level = "ERROR"
			} else if strings.Contains(lowerMsg, "warn") {
				level = "WARN"
			} else if strings.Contains(lowerMsg, "debug") {
				level = "DEBUG"
			}
			
			logs = append(logs, LogEntry{
				Timestamp: timestamp,
				Message:   message,
				Container: "main", // Could be extracted from multi-container pods
				Level:     level,
			})
		}

		if err := scanner.Err(); err != nil {
			return logsLoadedMsg{err: fmt.Errorf("error reading logs: %v", err)}
		}

		if len(logs) == 0 {
			return logsLoadedMsg{logs: []LogEntry{{
				Timestamp: time.Now(),
				Message:   "No logs available for this pod",
				Container: "system",
				Level:     "INFO",
			}}}
		}

		return logsLoadedMsg{logs: logs}
	}
}

// loadEventsCmd creates a command to asynchronously load events for selected resource
func (m Model) loadEventsCmd() tea.Cmd {
	return func() tea.Msg {
		var events []EventEntry
		
		if m.selectedK8sResource != nil {
			// Load events for specific resource
			fieldSelector := fmt.Sprintf("involvedObject.name=%s", m.selectedK8sResource.Name)
			eventList, err := m.clientset.CoreV1().Events(m.selectedK8sResource.Namespace).List(m.ctx, metav1.ListOptions{
				FieldSelector: fieldSelector,
			})
			if err != nil {
				return eventsLoadedMsg{err: err}
			}
			
			for _, event := range eventList.Items {
				events = append(events, EventEntry{
					Timestamp: event.CreationTimestamp.Time,
					Type:      event.Type,
					Reason:    event.Reason,
					Message:   event.Message,
					Source:    event.Source.Component,
					Count:     event.Count,
				})
			}
		} else {
			// Load all events in namespace
			eventList, err := m.clientset.CoreV1().Events(m.selectedNamespace).List(m.ctx, metav1.ListOptions{})
			if err != nil {
				return eventsLoadedMsg{err: err}
			}
			
			for _, event := range eventList.Items {
				events = append(events, EventEntry{
					Timestamp: event.CreationTimestamp.Time,
					Type:      event.Type,
					Reason:    event.Reason,
					Message:   event.Message,
					Source:    event.Source.Component,
					Count:     event.Count,
				})
			}
		}
		
		// Sort events by timestamp (newest first)
		sort.Slice(events, func(i, j int) bool {
			return events[i].Timestamp.After(events[j].Timestamp)
		})
		
		return eventsLoadedMsg{events: events}
	}
}

// analyzeResourceHealth checks for errors and warnings in resources
func (m Model) analyzeResourceHealth(resource *K8sResource) {
	resource.Errors = []string{}
	resource.Warnings = []string{}
	
	switch resource.ResourceType {
	case PodsResource:
		// Analyze pod health
		if resource.Status == "Failed" || resource.Status == "Error" {
			resource.Errors = append(resource.Errors, "Pod is in failed state")
		}
		if resource.Status == "Pending" {
			resource.Warnings = append(resource.Warnings, "Pod is stuck in pending state")
		}
		if restarts := resource.Details["Restarts"]; restarts != "" && restarts != "0" {
			if r, err := strconv.Atoi(restarts); err == nil && r > 5 {
				resource.Warnings = append(resource.Warnings, fmt.Sprintf("High restart count: %s", restarts))
			}
		}
		
	case DeploymentsResource:
		// Analyze deployment health
		if resource.Status == "NotReady" {
			resource.Errors = append(resource.Errors, "Deployment replicas not ready")
		}
		
	case ServicesResource:
		// Analyze service health
		if endpoints := resource.Details["Endpoints"]; endpoints == "0" {
			resource.Warnings = append(resource.Warnings, "Service has no endpoints")
		}
		
	case JobsResource:
		// Analyze job health
		if resource.Status == "Failed" {
			resource.Errors = append(resource.Errors, "Job failed to complete")
		}
	}
}

// Resource loading functions with enhanced error detection

func (m Model) loadNodes() ([]K8sResource, error) {
	nodes, err := m.clientset.CoreV1().Nodes().List(m.ctx, metav1.ListOptions{})
	if err != nil {
		return nil, err
	}
	
	var resources []K8sResource
	now := time.Now()
	
	for _, node := range nodes.Items {
		age := humanAge(now.Sub(node.CreationTimestamp.Time))
		
		// Determine node status and conditions
		status := "Unknown"
		var nodeErrors []string
		var nodeWarnings []string
		
		for _, condition := range node.Status.Conditions {
			switch condition.Type {
			case corev1.NodeReady:
				if condition.Status == corev1.ConditionTrue {
					status = "Ready"
				} else {
					status = "NotReady"
					nodeErrors = append(nodeErrors, fmt.Sprintf("Node not ready: %s", condition.Message))
				}
			case corev1.NodeMemoryPressure:
				if condition.Status == corev1.ConditionTrue {
					nodeWarnings = append(nodeWarnings, "Node under memory pressure")
				}
			case corev1.NodeDiskPressure:
				if condition.Status == corev1.ConditionTrue {
					nodeWarnings = append(nodeWarnings, "Node under disk pressure")
				}
			case corev1.NodePIDPressure:
				if condition.Status == corev1.ConditionTrue {
					nodeWarnings = append(nodeWarnings, "Node under PID pressure")
				}
			}
		}
		
		// Get CPU and memory capacity
		cpu := node.Status.Capacity[corev1.ResourceCPU]
		memory := node.Status.Capacity[corev1.ResourceMemory]
		
		resource := K8sResource{
			Name:         node.Name,
			Namespace:    "", // Nodes are cluster-scoped
			Status:       status,
			Age:          age,
			ResourceType: NodesResource,
			Errors:       nodeErrors,
			Warnings:     nodeWarnings,
			Details: map[string]string{
				"CPU":      cpu.String(),
				"Memory":   formatMemory(memory),
				"OS":       node.Status.NodeInfo.OperatingSystem,
				"Arch":     node.Status.NodeInfo.Architecture,
				"Version":  node.Status.NodeInfo.KubeletVersion,
				"Runtime":  node.Status.NodeInfo.ContainerRuntimeVersion,
			},
		}
		resources = append(resources, resource)
	}
	
	return resources, nil
}

func (m Model) loadPods() ([]K8sResource, error) {
	pods, err := m.clientset.CoreV1().Pods(m.selectedNamespace).List(m.ctx, metav1.ListOptions{})
	if err != nil {
		return nil, err
	}
	
	var resources []K8sResource
	now := time.Now()
	
	for _, pod := range pods.Items {
		restarts := int32(0)
		var containerErrors []string
		var containerWarnings []string
		
		// Analyze container statuses
		for _, cs := range pod.Status.ContainerStatuses {
			restarts += cs.RestartCount
			
			if !cs.Ready {
				if cs.State.Waiting != nil {
					containerErrors = append(containerErrors, fmt.Sprintf("Container %s waiting: %s", cs.Name, cs.State.Waiting.Reason))
				}
				if cs.State.Terminated != nil && cs.State.Terminated.ExitCode != 0 {
					containerErrors = append(containerErrors, fmt.Sprintf("Container %s terminated with exit code %d", cs.Name, cs.State.Terminated.ExitCode))
				}
			}
			
			if cs.RestartCount > 3 {
				containerWarnings = append(containerWarnings, fmt.Sprintf("Container %s has high restart count: %d", cs.Name, cs.RestartCount))
			}
		}
		
		age := humanAge(now.Sub(pod.CreationTimestamp.Time))
		
		// Calculate resource requests/limits
		var cpuReq, memoryReq, cpuLimit, memoryLimit resource.Quantity
		for _, container := range pod.Spec.Containers {
			if container.Resources.Requests != nil {
				cpuReq.Add(container.Resources.Requests[corev1.ResourceCPU])
				memoryReq.Add(container.Resources.Requests[corev1.ResourceMemory])
			}
			if container.Resources.Limits != nil {
				cpuLimit.Add(container.Resources.Limits[corev1.ResourceCPU])
				memoryLimit.Add(container.Resources.Limits[corev1.ResourceMemory])
			}
		}
		
		// Analyze pod-level issues
		var podErrors []string
		var podWarnings []string
		
		switch pod.Status.Phase {
		case corev1.PodFailed:
			podErrors = append(podErrors, "Pod is in failed state")
		case corev1.PodPending:
			if age := now.Sub(pod.CreationTimestamp.Time); age > 5*time.Minute {
				podWarnings = append(podWarnings, "Pod stuck in pending state")
			}
		}
		
		// Combine all errors and warnings
		allErrors := append(containerErrors, podErrors...)
		allWarnings := append(containerWarnings, podWarnings...)
		
		resource := K8sResource{
			Name:         pod.Name,
			Namespace:    pod.Namespace,
			Status:       string(pod.Status.Phase),
			Age:          age,
			ResourceType: PodsResource,
			Errors:       allErrors,
			Warnings:     allWarnings,
			Details: map[string]string{
				"Node":        pod.Spec.NodeName,
				"Restarts":    fmt.Sprintf("%d", restarts),
				"IP":          pod.Status.PodIP,
				"Containers":  fmt.Sprintf("%d", len(pod.Spec.Containers)),
				"CPU Req":     cpuReq.String(),
				"Memory Req":  formatMemory(memoryReq),
				"CPU Limit":   cpuLimit.String(),
				"Memory Limit": formatMemory(memoryLimit),
				"QoS Class":   string(pod.Status.QOSClass),
			},
		}
		resources = append(resources, resource)
	}
	
	return resources, nil
}

// loadEventsResource as a resource type  
func (m Model) loadEventsResource() ([]K8sResource, error) {
	events, err := m.clientset.CoreV1().Events(m.selectedNamespace).List(m.ctx, metav1.ListOptions{})
	if err != nil {
		return nil, err
	}
	
	var resources []K8sResource
	now := time.Now()
	
	// Sort events by creation time (newest first)
	sort.Slice(events.Items, func(i, j int) bool {
		return events.Items[i].CreationTimestamp.After(events.Items[j].CreationTimestamp.Time)
	})
	
	for _, event := range events.Items {
		age := humanAge(now.Sub(event.CreationTimestamp.Time))
		
		var eventErrors []string
		var eventWarnings []string
		
		if event.Type == "Warning" {
			eventWarnings = append(eventWarnings, event.Message)
		}
		
		resource := K8sResource{
			Name:         fmt.Sprintf("%s/%s", event.InvolvedObject.Kind, event.InvolvedObject.Name),
			Namespace:    event.Namespace,
			Status:       event.Type,
			Age:          age,
			ResourceType: EventsResource,
			Errors:       eventErrors,
			Warnings:     eventWarnings,
			Details: map[string]string{
				"Reason":  event.Reason,
				"Source":  event.Source.Component,
				"Count":   fmt.Sprintf("%d", event.Count),
				"Message": truncateString(event.Message, 100),
			},
		}
		resources = append(resources, resource)
	}
	
	return resources, nil
}

// Additional resource loading functions (simplified for brevity)
func (m Model) loadPersistentVolumes() ([]K8sResource, error) {
	pvs, err := m.clientset.CoreV1().PersistentVolumes().List(m.ctx, metav1.ListOptions{})
	if err != nil {
		return nil, err
	}
	
	var resources []K8sResource
	now := time.Now()
	
	for _, pv := range pvs.Items {
		age := humanAge(now.Sub(pv.CreationTimestamp.Time))
		capacity := pv.Spec.Capacity[corev1.ResourceStorage]
		
		var pvErrors []string
		if pv.Status.Phase == corev1.VolumeFailed {
			pvErrors = append(pvErrors, "Persistent Volume is in failed state")
		}
		
		resource := K8sResource{
			Name:         pv.Name,
			Namespace:    "",
			Status:       string(pv.Status.Phase),
			Age:          age,
			ResourceType: PersistentVolumesResource,
			Errors:       pvErrors,
			Details: map[string]string{
				"Capacity":     capacity.String(),
				"AccessModes":  strings.Join(accessModesToStrings(pv.Spec.AccessModes), ","),
				"ReclaimPolicy": string(pv.Spec.PersistentVolumeReclaimPolicy),
				"StorageClass":  pv.Spec.StorageClassName,
				"Claim":         formatPVClaim(pv.Spec.ClaimRef),
			},
		}
		resources = append(resources, resource)
	}
	
	return resources, nil
}

// Add other resource loading functions (abbreviated for space)
func (m Model) loadStorageClasses() ([]K8sResource, error) {
	scs, err := m.clientset.StorageV1().StorageClasses().List(m.ctx, metav1.ListOptions{})
	if err != nil {
		return nil, err
	}
	
	var resources []K8sResource
	now := time.Now()
	
	for _, sc := range scs.Items {
		age := humanAge(now.Sub(sc.CreationTimestamp.Time))
		
		resource := K8sResource{
			Name:         sc.Name,
			Namespace:    "",
			Status:       "Active",
			Age:          age,
			ResourceType: StorageClassesResource,
			Details: map[string]string{
				"Provisioner":          sc.Provisioner,
				"ReclaimPolicy":        reclaimPolicyPtrToString(sc.ReclaimPolicy),
				"VolumeBindingMode":    volumeBindingModePtrToString(sc.VolumeBindingMode),
				"AllowVolumeExpansion": boolPtrToString(sc.AllowVolumeExpansion),
			},
		}
		resources = append(resources, resource)
	}
	
	return resources, nil
}

func (m Model) loadClusterRoles() ([]K8sResource, error) {
	crs, err := m.clientset.RbacV1().ClusterRoles().List(m.ctx, metav1.ListOptions{})
	if err != nil {
		return nil, err
	}
	
	var resources []K8sResource
	now := time.Now()
	
	for _, cr := range crs.Items {
		age := humanAge(now.Sub(cr.CreationTimestamp.Time))
		
		resource := K8sResource{
			Name:         cr.Name,
			Namespace:    "",
			Status:       "Active",
			Age:          age,
			ResourceType: ClusterRolesResource,
			Details: map[string]string{
				"Rules": fmt.Sprintf("%d rules", len(cr.Rules)),
			},
		}
		resources = append(resources, resource)
	}
	
	return resources, nil
}

// Simplified versions of other resource loaders
func (m Model) loadServices() ([]K8sResource, error) {
	services, err := m.clientset.CoreV1().Services(m.selectedNamespace).List(m.ctx, metav1.ListOptions{})
	if err != nil {
		return nil, err
	}
	
	var resources []K8sResource
	now := time.Now()
	
	for _, svc := range services.Items {
		age := humanAge(now.Sub(svc.CreationTimestamp.Time))
		
		endpoints, _ := m.clientset.CoreV1().Endpoints(m.selectedNamespace).Get(m.ctx, svc.Name, metav1.GetOptions{})
		endpointCount := 0
		if endpoints != nil {
			for _, subset := range endpoints.Subsets {
				endpointCount += len(subset.Addresses)
			}
		}
		
		var svcWarnings []string
		if endpointCount == 0 {
			svcWarnings = append(svcWarnings, "Service has no endpoints")
		}
		
		resource := K8sResource{
			Name:         svc.Name,
			Namespace:    svc.Namespace,
			Status:       string(svc.Spec.Type),
			Age:          age,
			ResourceType: ServicesResource,
			Warnings:     svcWarnings,
			Details: map[string]string{
				"ClusterIP": svc.Spec.ClusterIP,
				"Endpoints": fmt.Sprintf("%d", endpointCount),
				"Ports":     fmt.Sprintf("%d ports", len(svc.Spec.Ports)),
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
	
	for _, deploy := range deployments.Items {
		age := humanAge(now.Sub(deploy.CreationTimestamp.Time))
		
		ready := fmt.Sprintf("%d/%d", deploy.Status.ReadyReplicas, deploy.Status.Replicas)
		status := "Running"
		var deployErrors []string
		
		if deploy.Status.ReadyReplicas != deploy.Status.Replicas {
			status = "NotReady"
			deployErrors = append(deployErrors, "Not all replicas are ready")
		}
		
		resource := K8sResource{
			Name:         deploy.Name,
			Namespace:    deploy.Namespace,
			Status:       status,
			Age:          age,
			ResourceType: DeploymentsResource,
			Errors:       deployErrors,
			Details: map[string]string{
				"Ready":     ready,
				"Available": fmt.Sprintf("%d", deploy.Status.AvailableReplicas),
				"Strategy":  string(deploy.Spec.Strategy.Type),
			},
		}
		resources = append(resources, resource)
	}
	
	return resources, nil
}

// Add simplified versions of other resource loaders
func (m Model) loadConfigMaps() ([]K8sResource, error) {
	cms, err := m.clientset.CoreV1().ConfigMaps(m.selectedNamespace).List(m.ctx, metav1.ListOptions{})
	if err != nil {
		return nil, err
	}
	
	var resources []K8sResource
	now := time.Now()
	
	for _, cm := range cms.Items {
		age := humanAge(now.Sub(cm.CreationTimestamp.Time))
		
		resource := K8sResource{
			Name:         cm.Name,
			Namespace:    cm.Namespace,
			Status:       "Active",
			Age:          age,
			ResourceType: ConfigMapsResource,
			Details: map[string]string{
				"Data Keys": fmt.Sprintf("%d", len(cm.Data)),
				"Size":      fmt.Sprintf("%d bytes", calculateConfigMapSize(&cm)),
			},
		}
		resources = append(resources, resource)
	}
	
	return resources, nil
}

func (m Model) loadSecrets() ([]K8sResource, error) {
	secrets, err := m.clientset.CoreV1().Secrets(m.selectedNamespace).List(m.ctx, metav1.ListOptions{})
	if err != nil {
		return nil, err
	}
	
	var resources []K8sResource
	now := time.Now()
	
	for _, secret := range secrets.Items {
		age := humanAge(now.Sub(secret.CreationTimestamp.Time))
		
		resource := K8sResource{
			Name:         secret.Name,
			Namespace:    secret.Namespace,
			Status:       "Active",
			Age:          age,
			ResourceType: SecretsResource,
			Details: map[string]string{
				"Type":      string(secret.Type),
				"Data Keys": fmt.Sprintf("%d", len(secret.Data)),
			},
		}
		resources = append(resources, resource)
	}
	
	return resources, nil
}

func (m Model) loadIngress() ([]K8sResource, error) {
	ingresses, err := m.clientset.NetworkingV1().Ingresses(m.selectedNamespace).List(m.ctx, metav1.ListOptions{})
	if err != nil {
		return nil, err
	}
	
	var resources []K8sResource
	now := time.Now()
	
	for _, ing := range ingresses.Items {
		age := humanAge(now.Sub(ing.CreationTimestamp.Time))
		
		var hosts []string
		for _, rule := range ing.Spec.Rules {
			if rule.Host != "" {
				hosts = append(hosts, rule.Host)
			}
		}
		
		resource := K8sResource{
			Name:         ing.Name,
			Namespace:    ing.Namespace,
			Status:       "Active",
			Age:          age,
			ResourceType: IngressResource,
			Details: map[string]string{
				"Hosts": strings.Join(hosts, ","),
				"Rules": fmt.Sprintf("%d", len(ing.Spec.Rules)),
				"Class": stringValue(ing.Spec.IngressClassName),
			},
		}
		resources = append(resources, resource)
	}
	
	return resources, nil
}

func (m Model) loadPersistentVolumeClaims() ([]K8sResource, error) {
	pvcs, err := m.clientset.CoreV1().PersistentVolumeClaims(m.selectedNamespace).List(m.ctx, metav1.ListOptions{})
	if err != nil {
		return nil, err
	}
	
	var resources []K8sResource
	now := time.Now()
	
	for _, pvc := range pvcs.Items {
		age := humanAge(now.Sub(pvc.CreationTimestamp.Time))
		
		capacity := resource.Quantity{}
		if pvc.Status.Capacity != nil {
			capacity = pvc.Status.Capacity[corev1.ResourceStorage]
		}
		
		var pvcErrors []string
		if pvc.Status.Phase == corev1.ClaimPending {
			pvcErrors = append(pvcErrors, "PVC is stuck in pending state")
		}
		
		resource := K8sResource{
			Name:         pvc.Name,
			Namespace:    pvc.Namespace,
			Status:       string(pvc.Status.Phase),
			Age:          age,
			ResourceType: PersistentVolumeClaimsResource,
			Errors:       pvcErrors,
			Details: map[string]string{
				"Capacity":     capacity.String(),
				"AccessModes":  strings.Join(accessModesToStrings(pvc.Spec.AccessModes), ","),
				"StorageClass": stringPtrValue(pvc.Spec.StorageClassName),
				"Volume":       pvc.Spec.VolumeName,
			},
		}
		resources = append(resources, resource)
	}
	
	return resources, nil
}

func (m Model) loadReplicaSets() ([]K8sResource, error) {
	rss, err := m.clientset.AppsV1().ReplicaSets(m.selectedNamespace).List(m.ctx, metav1.ListOptions{})
	if err != nil {
		return nil, err
	}
	
	var resources []K8sResource
	now := time.Now()
	
	for _, rs := range rss.Items {
		age := humanAge(now.Sub(rs.CreationTimestamp.Time))
		
		ready := fmt.Sprintf("%d/%d", rs.Status.ReadyReplicas, rs.Status.Replicas)
		status := "Running"
		var rsErrors []string
		
		if rs.Status.ReadyReplicas != rs.Status.Replicas {
			status = "NotReady"
			rsErrors = append(rsErrors, "Not all replicas are ready")
		}
		
		resource := K8sResource{
			Name:         rs.Name,
			Namespace:    rs.Namespace,
			Status:       status,
			Age:          age,
			ResourceType: ReplicaSetsResource,
			Errors:       rsErrors,
			Details: map[string]string{
				"Ready":     ready,
				"Available": fmt.Sprintf("%d", rs.Status.AvailableReplicas),
			},
		}
		resources = append(resources, resource)
	}
	
	return resources, nil
}

func (m Model) loadDaemonSets() ([]K8sResource, error) {
	dss, err := m.clientset.AppsV1().DaemonSets(m.selectedNamespace).List(m.ctx, metav1.ListOptions{})
	if err != nil {
		return nil, err
	}
	
	var resources []K8sResource
	now := time.Now()
	
	for _, ds := range dss.Items {
		age := humanAge(now.Sub(ds.CreationTimestamp.Time))
		
		desired := ds.Status.DesiredNumberScheduled
		ready := ds.Status.NumberReady
		status := "Running"
		var dsErrors []string
		
		if ready != desired {
			status = "NotReady"
			dsErrors = append(dsErrors, "Not all pods are ready")
		}
		
		resource := K8sResource{
			Name:         ds.Name,
			Namespace:    ds.Namespace,
			Status:       status,
			Age:          age,
			ResourceType: DaemonSetsResource,
			Errors:       dsErrors,
			Details: map[string]string{
				"Desired": fmt.Sprintf("%d", desired),
				"Ready":   fmt.Sprintf("%d", ready),
			},
		}
		resources = append(resources, resource)
	}
	
	return resources, nil
}

func (m Model) loadStatefulSets() ([]K8sResource, error) {
	sss, err := m.clientset.AppsV1().StatefulSets(m.selectedNamespace).List(m.ctx, metav1.ListOptions{})
	if err != nil {
		return nil, err
	}
	
	var resources []K8sResource
	now := time.Now()
	
	for _, ss := range sss.Items {
		age := humanAge(now.Sub(ss.CreationTimestamp.Time))
		
		ready := fmt.Sprintf("%d/%d", ss.Status.ReadyReplicas, ss.Status.Replicas)
		status := "Running"
		var ssErrors []string
		
		if ss.Status.ReadyReplicas != ss.Status.Replicas {
			status = "NotReady"
			ssErrors = append(ssErrors, "Not all replicas are ready")
		}
		
		resource := K8sResource{
			Name:         ss.Name,
			Namespace:    ss.Namespace,
			Status:       status,
			Age:          age,
			ResourceType: StatefulSetsResource,
			Errors:       ssErrors,
			Details: map[string]string{
				"Ready":   ready,
				"Service": ss.Spec.ServiceName,
			},
		}
		resources = append(resources, resource)
	}
	
	return resources, nil
}

func (m Model) loadJobs() ([]K8sResource, error) {
	jobs, err := m.clientset.BatchV1().Jobs(m.selectedNamespace).List(m.ctx, metav1.ListOptions{})
	if err != nil {
		return nil, err
	}
	
	var resources []K8sResource
	now := time.Now()
	
	for _, job := range jobs.Items {
		age := humanAge(now.Sub(job.CreationTimestamp.Time))
		
		completions := fmt.Sprintf("%d/%d", job.Status.Succeeded, *job.Spec.Completions)
		status := "Running"
		var jobErrors []string
		
		if job.Status.Succeeded == *job.Spec.Completions {
			status = "Complete"
		} else if job.Status.Failed > 0 {
			status = "Failed"
			jobErrors = append(jobErrors, "Job has failed pods")
		}
		
		resource := K8sResource{
			Name:         job.Name,
			Namespace:    job.Namespace,
			Status:       status,
			Age:          age,
			ResourceType: JobsResource,
			Errors:       jobErrors,
			Details: map[string]string{
				"Completions": completions,
				"Failed":      fmt.Sprintf("%d", job.Status.Failed),
			},
		}
		resources = append(resources, resource)
	}
	
	return resources, nil
}

func (m Model) loadCronJobs() ([]K8sResource, error) {
	cronJobs, err := m.clientset.BatchV1().CronJobs(m.selectedNamespace).List(m.ctx, metav1.ListOptions{})
	if err != nil {
		return nil, err
	}
	
	var resources []K8sResource
	now := time.Now()
	
	for _, cj := range cronJobs.Items {
		age := humanAge(now.Sub(cj.CreationTimestamp.Time))
		
		status := "Active"
		if cj.Spec.Suspend != nil && *cj.Spec.Suspend {
			status = "Suspended"
		}
		
		lastSchedule := "Never"
		if cj.Status.LastScheduleTime != nil {
			lastSchedule = humanAge(now.Sub(cj.Status.LastScheduleTime.Time)) + " ago"
		}
		
		resource := K8sResource{
			Name:         cj.Name,
			Namespace:    cj.Namespace,
			Status:       status,
			Age:          age,
			ResourceType: CronJobsResource,
			Details: map[string]string{
				"Schedule":     cj.Spec.Schedule,
				"LastSchedule": lastSchedule,
			},
		}
		resources = append(resources, resource)
	}
	
	return resources, nil
}

// Update handles messages and updates model state - required by Bubble Tea
func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		return m, nil
		
	case tea.KeyMsg:
		return m.handleKeyPress(msg)
		
	case namespacesLoadedMsg:
		m.loading = false
		if msg.err != nil {
			m.errorMessage = fmt.Sprintf("Error loading namespaces: %v", msg.err)
		} else {
			m.namespaces = msg.namespaces
			m.errorMessage = ""
		}
		return m, nil
		
	case resourcesLoadedMsg:
		m.loading = false
		if msg.err != nil {
			m.errorMessage = fmt.Sprintf("Error loading resources: %v", msg.err)
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
				m.currentView = LogView
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
			if m.selectedK8sResource != nil {
				m.loading = true
				return m, m.loadLogs()
			}
		case EventView:
			m.loading = true
			return m, m.loadEventsCmd()
		}
		
	case "a":
		// Toggle auto-refresh
		m.autoRefresh = !m.autoRefresh
	}
	
	return m, nil
}

// handleSelection processes enter/space key selections
func (m Model) handleSelection() (tea.Model, tea.Cmd) {
	switch m.currentView {
	
	case ClusterOrNamespaceView:
		m.viewStack = append(m.viewStack, m.currentView)
		if m.cursor == 0 {
			// Cluster-scoped resources
			m.selectedScope = ClusterScoped
			m.currentView = ResourceView
			// Cluster-scoped resources
			m.resourceTypes = []ResourceType{NodesResource, PersistentVolumesResource, StorageClassesResource, ClusterRolesResource}
			// TODO: Add OpenShift Projects if available (temporarily disabled)
			// if m.isOpenShift {
			//	m.resourceTypes = append(m.resourceTypes, ProjectsResource)
			// }
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
			// TODO: Add OpenShift-specific resources if available (temporarily disabled)
			// if m.isOpenShift {
			//	m.resourceTypes = append(m.resourceTypes, RoutesResource, DeploymentConfigsResource)
			// }
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

// navigateBack handles ESC/backspace navigation
func (m Model) navigateBack() (tea.Model, tea.Cmd) {
	if len(m.viewStack) > 0 {
		// Pop the last view from stack
		m.currentView = m.viewStack[len(m.viewStack)-1]
		m.viewStack = m.viewStack[:len(m.viewStack)-1]
		m.cursor = 0
		m.logScrollOffset = 0
		m.eventScrollOffset = 0
		m.errorMessage = ""
	}
	return m, nil
}

// View renders the UI - required by Bubble Tea
func (m Model) View() string {
	// Define catchy styles using the color scheme
	headerStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(colors.Text).
		Background(colors.Primary).
		Padding(0, 1).
		Margin(0, 0, 1, 0)
		
	selectedStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(colors.Text).
		Background(colors.Accent)
		
	normalStyle := lipgloss.NewStyle().
		Foreground(colors.Text)
		
	errorStyle := lipgloss.NewStyle().
		Foreground(colors.Error).
		Bold(true)
		
	_ = lipgloss.NewStyle().
		Foreground(colors.Warning).
		Bold(true)
		
	successStyle := lipgloss.NewStyle().
		Foreground(colors.Success).
		Bold(true)
		
	infoStyle := lipgloss.NewStyle().
		Foreground(colors.Info)
		
	helpStyle := lipgloss.NewStyle().
		Foreground(colors.Secondary).
		Italic(true).
		Margin(1, 0, 0, 0)

	var content strings.Builder
	
	// Header with current context
	header := m.buildHeader()
	content.WriteString(headerStyle.Render(header) + "\n")
	
	// Error message if present
	if m.errorMessage != "" {
		content.WriteString(errorStyle.Render("⚠ " + m.errorMessage) + "\n\n")
	}
	
	// Loading indicator
	if m.loading {
		content.WriteString(infoStyle.Render("⏳ Loading...") + "\n\n")
	}
	
	// Main content based on current view
	switch m.currentView {
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
			content.WriteString(style.Render(prefix + option) + "\n")
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
			content.WriteString(style.Render(prefix + ns) + "\n")
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
			content.WriteString(style.Render(prefix + info.Icon + " " + rt.String()) + "\n")
		}
		
	case DetailView:
		content.WriteString(m.renderResourceDetails())
		
	case LogView:
		content.WriteString(m.renderLogs())
		
	case EventView:
		content.WriteString(m.renderEvents())
	}
	
	// Help text
	content.WriteString("\n" + strings.Repeat("─", min(m.width, 80)) + "\n")
	helpText := m.buildHelpText()
	content.WriteString(helpStyle.Render(helpText))
	
	return content.String()
}

// buildHeader creates the application header with context information
func (m Model) buildHeader() string {
	title := "🚀 Kubernetes Resource Monitor"
	
	var context []string
	if m.selectedScope == ClusterScoped {
		context = append(context, "Scope: Cluster")
	} else if m.selectedNamespace != "" {
		context = append(context, "NS: "+m.selectedNamespace)
	}
	
	if m.currentView == DetailView || m.currentView == LogView || m.currentView == EventView {
		info := m.selectedResource.GetResourceInfo()
		context = append(context, "Type: "+info.Icon+" "+m.selectedResource.String())
		if !m.lastUpdate.IsZero() {
			context = append(context, "Updated: "+m.lastUpdate.Format("15:04:05"))
		}
	}
	
	if (m.currentView == LogView || m.currentView == EventView) && m.selectedK8sResource != nil {
		viewType := "Logs"
		if m.currentView == EventView {
			viewType = "Events"
		}
		context = append(context, viewType+": "+m.selectedK8sResource.Name)
	}
	
	if len(context) > 0 {
		title += " | " + strings.Join(context, " | ")
	}
	
	if m.autoRefresh {
		title += " | Auto-refresh: ON"
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
	
	content.WriteString(headerText + "\n\n")
	
	// Create enhanced table with proper column formatting
	selectedStyle := lipgloss.NewStyle().Bold(true).Foreground(colors.Text).Background(colors.Accent)
	normalStyle := lipgloss.NewStyle().Foreground(colors.Text)
	errorRowStyle := lipgloss.NewStyle().Foreground(colors.Error)
	warningRowStyle := lipgloss.NewStyle().Foreground(colors.Warning)
	successRowStyle := lipgloss.NewStyle().Foreground(colors.Success)
	
	// Table header
	var headerFormat, rowFormat string
	if m.selectedScope == ClusterScoped {
		headerFormat = "%-35s %-15s %-10s"
		rowFormat = "%-35s %-15s %-10s"
		header := fmt.Sprintf(headerFormat, "NAME", "STATUS", "AGE")
		content.WriteString(normalStyle.Render(header) + "\n")
	} else {
		headerFormat = "%-30s %-15s %-10s"
		rowFormat = "%-30s %-15s %-10s"
		header := fmt.Sprintf(headerFormat, "NAME", "STATUS", "AGE")
		content.WriteString(normalStyle.Render(header) + "\n")
	}
	content.WriteString(strings.Repeat("─", 70) + "\n")
	
	// Table rows with color-coded status
	for i, resource := range m.resources {
		row := fmt.Sprintf(rowFormat, 
			truncateString(resource.Name, 33),
			truncateString(resource.Status, 13),
			resource.Age)
		
		// Choose style based on resource health
		var style lipgloss.Style
		if i == m.cursor {
			style = selectedStyle
			row = "▶ " + row
		} else {
			row = "  " + row
			
			// Color code based on status and errors
			if len(resource.Errors) > 0 {
				style = errorRowStyle
			} else if len(resource.Warnings) > 0 {
				style = warningRowStyle
			} else if resource.Status == "Running" || resource.Status == "Ready" || resource.Status == "Active" {
				style = successRowStyle
			} else {
				style = normalStyle
			}
		}
		
		content.WriteString(style.Render(row) + "\n")
		
		// Show errors, warnings, and details for selected resource
		if i == m.cursor {
			content.WriteString("\n")
			
			// Show errors first
			if len(resource.Errors) > 0 {
				errorStyle := lipgloss.NewStyle().Foreground(colors.Error).Bold(true)
				content.WriteString(errorStyle.Render("    🚨 ERRORS:") + "\n")
				for _, err := range resource.Errors {
					content.WriteString(lipgloss.NewStyle().Foreground(colors.Error).Render("      • " + err) + "\n")
				}
				content.WriteString("\n")
			}
			
			// Show warnings
			if len(resource.Warnings) > 0 {
				warningStyle := lipgloss.NewStyle().Foreground(colors.Warning).Bold(true)
				content.WriteString(warningStyle.Render("    ⚠️  WARNINGS:") + "\n")
				for _, warn := range resource.Warnings {
					content.WriteString(lipgloss.NewStyle().Foreground(colors.Warning).Render("      • " + warn) + "\n")
				}
				content.WriteString("\n")
			}
			
			// Show details
			if len(resource.Details) > 0 {
				detailStyle := lipgloss.NewStyle().Foreground(colors.Info).Bold(true)
				content.WriteString(detailStyle.Render("    📋 DETAILS:") + "\n")
				
				keys := make([]string, 0, len(resource.Details))
				for k := range resource.Details {
					keys = append(keys, k)
				}
				sort.Strings(keys)
				
				// Display details in two columns
				for j := 0; j < len(keys); j += 2 {
					leftKey := keys[j]
					leftValue := resource.Details[leftKey]
					leftDetail := fmt.Sprintf("      %-15s: %-20s", leftKey, truncateString(leftValue, 18))
					
					if j+1 < len(keys) {
						rightKey := keys[j+1]
						rightValue := resource.Details[rightKey]
						rightDetail := fmt.Sprintf("%-15s: %s", rightKey, truncateString(rightValue, 18))
						content.WriteString(normalStyle.Render(leftDetail + rightDetail) + "\n")
					} else {
						content.WriteString(normalStyle.Render(leftDetail) + "\n")
					}
				}
			}
			content.WriteString("\n")
		}
	}
	
	return content.String()
}

// renderLogs creates the enhanced logs view with color-coding
func (m Model) renderLogs() string {
	if m.selectedK8sResource == nil {
		return "No resource selected for logs"
	}
	
	var content strings.Builder
	headerStyle := lipgloss.NewStyle().Foreground(colors.Success).Bold(true)
	content.WriteString(headerStyle.Render(fmt.Sprintf("📋 Logs for %s '%s':", 
		m.selectedK8sResource.ResourceType.String(), m.selectedK8sResource.Name)) + "\n\n")
	
	if len(m.logEntries) == 0 {
		content.WriteString("No logs available or logs loading...\n")
		return content.String()
	}
	
	// Calculate visible area for logs
	visibleLines := m.height - 10
	if visibleLines < 5 {
		visibleLines = 5
	}
	
	// Show scroll position indicator
	totalLines := len(m.logEntries)
	if totalLines > visibleLines {
		infoStyle := lipgloss.NewStyle().Foreground(colors.Info)
		content.WriteString(infoStyle.Render(fmt.Sprintf("Showing lines %d-%d of %d (scroll with ↑/↓)",
			m.logScrollOffset+1,
			min(m.logScrollOffset+visibleLines, totalLines),
			totalLines)) + "\n\n")
	} else {
		infoStyle := lipgloss.NewStyle().Foreground(colors.Info)
		content.WriteString(infoStyle.Render(fmt.Sprintf("Showing all %d log lines:", totalLines)) + "\n\n")
	}
	
	// Render visible log lines with color coding
	start := m.logScrollOffset
	end := min(start+visibleLines, len(m.logEntries))
	
	for i := start; i < end; i++ {
		logLine := m.logEntries[i]
		
		// Choose color based on log level
		var logStyle lipgloss.Style
		switch logLine.Level {
		case "ERROR":
			logStyle = lipgloss.NewStyle().Foreground(colors.Error)
		case "WARN":
			logStyle = lipgloss.NewStyle().Foreground(colors.Warning)
		case "DEBUG":
			logStyle = lipgloss.NewStyle().Foreground(colors.Muted)
		default:
			logStyle = lipgloss.NewStyle().Foreground(colors.Text)
		}
		
		// Format timestamp
		timeStyle := lipgloss.NewStyle().Foreground(colors.Secondary)
		levelStyle := lipgloss.NewStyle().Foreground(colors.Info).Bold(true)
		
		line := fmt.Sprintf("%s [%s] %s", 
			timeStyle.Render(logLine.Timestamp.Format("15:04:05")),
			levelStyle.Render(logLine.Level),
			logLine.Message)
		content.WriteString(logStyle.Render(line) + "\n")
	}
	
	return content.String()
}

// renderEvents creates the events view
func (m Model) renderEvents() string {
	var content strings.Builder
	headerStyle := lipgloss.NewStyle().Foreground(colors.Success).Bold(true)
	
	if m.selectedK8sResource != nil {
		content.WriteString(headerStyle.Render(fmt.Sprintf("📢 Events for %s '%s':", 
			m.selectedK8sResource.ResourceType.String(), m.selectedK8sResource.Name)) + "\n\n")
	} else {
		content.WriteString(headerStyle.Render(fmt.Sprintf("📢 Events in namespace '%s':", m.selectedNamespace)) + "\n\n")
	}
	
	if len(m.eventEntries) == 0 {
		content.WriteString("No events found\n")
		return content.String()
	}
	
	// Calculate visible area
	visibleLines := m.height - 10
	if visibleLines < 5 {
		visibleLines = 5
	}
	
	// Show scroll position indicator
	totalLines := len(m.eventEntries)
	if totalLines > visibleLines {
		infoStyle := lipgloss.NewStyle().Foreground(colors.Info)
		content.WriteString(infoStyle.Render(fmt.Sprintf("Showing events %d-%d of %d (scroll with ↑/↓)",
			m.eventScrollOffset+1,
			min(m.eventScrollOffset+visibleLines, totalLines),
			totalLines)) + "\n\n")
	} else {
		infoStyle := lipgloss.NewStyle().Foreground(colors.Info)
		content.WriteString(infoStyle.Render(fmt.Sprintf("Showing all %d events:", totalLines)) + "\n\n")
	}
	
	// Render visible events
	start := m.eventScrollOffset
	end := min(start+visibleLines, len(m.eventEntries))
	
	for i := start; i < end; i++ {
		event := m.eventEntries[i]
		
		// Choose color based on event type
		var eventStyle lipgloss.Style
		if event.Type == "Warning" {
			eventStyle = lipgloss.NewStyle().Foreground(colors.Warning)
		} else {
			eventStyle = lipgloss.NewStyle().Foreground(colors.Success)
		}
		
		timeStyle := lipgloss.NewStyle().Foreground(colors.Secondary)
		reasonStyle := lipgloss.NewStyle().Foreground(colors.Info).Bold(true)
		
		line := fmt.Sprintf("%s [%s] %s: %s", 
			timeStyle.Render(event.Timestamp.Format("15:04:05")),
			eventStyle.Render(event.Type),
			reasonStyle.Render(event.Reason),
			event.Message)
		content.WriteString(eventStyle.Render(line) + "\n")
	}
	
	return content.String()
}

// buildHelpText creates context-sensitive help information
func (m Model) buildHelpText() string {
	var help []string
	
	switch m.currentView {
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
			helpItems := []string{"↑/k: up", "↓/j: down", "enter: select", "esc: back", "r: refresh", "a: toggle auto-refresh", "q: quit"}
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
				"r: refresh", "a: toggle auto-refresh", "q: quit",
			}
		}
	default:
		help = []string{
			"↑/k: up", "↓/j: down", "enter/space: select", "esc: back",
			"r: refresh", "a: toggle auto-refresh", "q: quit",
		}
	}
	
	return strings.Join(help, " | ")
}

// Utility functions

// Helper functions for resource data formatting
func formatMemory(q resource.Quantity) string {
	if q.IsZero() {
		return "0"
	}
	return q.String()
}

func formatSelector(selector map[string]string) string {
	if len(selector) == 0 {
		return "none"
	}
	var parts []string
	for k, v := range selector {
		parts = append(parts, fmt.Sprintf("%s=%s", k, v))
	}
	return strings.Join(parts, ",")
}

func formatPVClaim(claim *corev1.ObjectReference) string {
	if claim == nil {
		return "none"
	}
	return fmt.Sprintf("%s/%s", claim.Namespace, claim.Name)
}

func formatDuration(start, end *metav1.Time) string {
	if start == nil {
		return "not started"
	}
	if end == nil {
		return "running"
	}
	return end.Sub(start.Time).String()
}

func accessModesToStrings(modes []corev1.PersistentVolumeAccessMode) []string {
	var result []string
	for _, mode := range modes {
		result = append(result, string(mode))
	}
	return result
}

func calculateConfigMapSize(cm *corev1.ConfigMap) int {
	size := 0
	for _, v := range cm.Data {
		size += len(v)
	}
	for _, v := range cm.BinaryData {
		size += len(v)
	}
	return size
}

func stringValue(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}

func stringPtrValue(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}

func reclaimPolicyPtrToString(p *corev1.PersistentVolumeReclaimPolicy) string {
	if p == nil {
		return ""
	}
	return string(*p)
}

func volumeBindingModePtrToString(v interface{}) string {
	if v == nil {
		return ""
	}
	return fmt.Sprintf("%v", v)
}

func boolPtrToString(b *bool) string {
	if b == nil {
		return "false"
	}
	return strconv.FormatBool(*b)
}

func int64Ptr(i int64) *int64 {
	return &i
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

// truncateString truncates string to maxLen with ellipsis
func truncateString(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen-3] + "..."
}

// initializeKubernetesClient sets up the Kubernetes client with proper error handling
func initializeKubernetesClient() (*kubernetes.Clientset, error) {
	// Try in-cluster config first (for when running inside cluster)
	config, err := clientcmd.BuildConfigFromFlags("", "")
	if err != nil {
		// Fall back to kubeconfig file
		kubeconfig := filepath.Join(homedir.HomeDir(), ".kube", "config")
		if _, err := os.Stat(kubeconfig); os.IsNotExist(err) {
			return nil, fmt.Errorf("kubeconfig not found at %s and in-cluster config failed: %v", kubeconfig, err)
		}
		
		config, err = clientcmd.BuildConfigFromFlags("", kubeconfig)
		if err != nil {
			return nil, fmt.Errorf("failed to load kubeconfig: %v", err)
		}
	}

	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		return nil, fmt.Errorf("failed to create Kubernetes clientset: %v", err)
	}

	return clientset, nil
}

// getKubernetesConfig returns the Kubernetes client configuration
func getKubernetesConfig() (*rest.Config, error) {
	var config *rest.Config
	var err error

	// Try in-cluster config first
	config, err = rest.InClusterConfig()
	if err != nil {
		// Fall back to kubeconfig file
		var kubeconfig string
		if home := homedir.HomeDir(); home != "" {
			kubeconfig = filepath.Join(home, ".kube", "config")
		}
		
		// Use environment variable if set
		if envConfig := os.Getenv("KUBECONFIG"); envConfig != "" {
			kubeconfig = envConfig
		}

		config, err = clientcmd.BuildConfigFromFlags("", kubeconfig)
		if err != nil {
			return nil, fmt.Errorf("failed to build config from kubeconfig: %v", err)
		}
	}

	return config, nil
}

// TODO: OpenShift-specific resource loading functions (temporarily disabled due to client compatibility issues)

// TODO: loadRoutes loads OpenShift Routes (disabled)
/*
func (m Model) loadRoutes() ([]K8sResource, error) {
	if m.routeClient == nil {
		return nil, fmt.Errorf("OpenShift Route client not available")
	}
	// ... implementation commented out for compatibility
	return nil, fmt.Errorf("OpenShift support temporarily disabled")
}
*/

// TODO: loadDeploymentConfigs loads OpenShift DeploymentConfigs (disabled)
/*
func (m Model) loadDeploymentConfigs() ([]K8sResource, error) {
	if m.openshiftAppsClient == nil {
		return nil, fmt.Errorf("OpenShift Apps client not available")
	}
	// ... implementation commented out for compatibility  
	return nil, fmt.Errorf("OpenShift support temporarily disabled")
}
*/

// TODO: loadProjects loads OpenShift Projects (disabled)
/*
func (m Model) loadProjects() ([]K8sResource, error) {
	if m.projectClient == nil {
		return nil, fmt.Errorf("OpenShift Project client not available")
	}
	// ... implementation commented out for compatibility
	return nil, fmt.Errorf("OpenShift support temporarily disabled")
}
*/

// main function - application entry point
func main() {
	// Set color profile for better terminal colors
	lipgloss.SetColorProfile(termenv.ANSI256)
	
	// Initialize Kubernetes client
	clientset, err := initializeKubernetesClient()
	if err != nil {
		log.Fatalf("Failed to initialize Kubernetes client: %v", err)
	}
	
	// TODO: Initialize OpenShift clients (temporarily disabled due to client compatibility issues)
	// var openshiftAppsClient *openshiftclient.Clientset
	// var routeClient *routeclient.Clientset  
	// var projectClient *projectclient.Clientset
	var isOpenShift bool = false // Set to false for now
	
	// TODO: Try to initialize OpenShift clients
	// config, err := getKubernetesConfig()
	// if err == nil {
	//	openshiftAppsClient, _ = openshiftclient.NewForConfig(config)
	//	routeClient, _ = routeclient.NewForConfig(config)
	//	projectClient, _ = projectclient.NewForConfig(config)
	//	
	//	// Test if we're on OpenShift by trying to list projects
	//	if projectClient != nil {
	//		_, err := projectClient.ProjectV1().Projects().List(context.Background(), metav1.ListOptions{Limit: 1})
	//		isOpenShift = err == nil
	//	}
	// }
	
	// Create initial model
	initialModel := Model{
		clientset:           clientset,
		ctx:                context.Background(),
		// TODO: OpenShift clients (temporarily disabled)
		// openshiftAppsClient: openshiftAppsClient,
		// routeClient:         routeClient,
		// projectClient:       projectClient,
		isOpenShift:         isOpenShift,
		currentView:         ClusterOrNamespaceView,
		cursor:              0,
		viewStack:           make([]ViewType, 0),
		namespaces:          make([]string, 0),
		resourceTypes:       make([]ResourceType, 0),
		resources:           make([]K8sResource, 0),
		logEntries:          make([]LogEntry, 0),
		eventEntries:        make([]EventEntry, 0),
		loading:             false,
		autoRefresh:         true,
	}
	
	// Start the Bubble Tea program
	program := tea.NewProgram(initialModel, tea.WithAltScreen())
	
	if _, err := program.Run(); err != nil {
		log.Fatalf("Error running program: %v", err)
	}
}