package ui

import (
	"encoding/json"
	"fmt"
	"os/exec"
	"sort"
	"strings"
	"time"

	"k8sgo/internal/types"
	"k8sgo/pkg/client"
	"k8sgo/pkg/config"
	"k8sgo/pkg/utils"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type App struct {
	state             types.AppState
	client            types.ClusterClient
	config            *config.ConfigManager
	toolDetector      *utils.ToolDetector
	bannerManager     *types.BannerManager
	paginator         *types.ResourcePaginator
	contextSelector   *ContextSelector
	namespaceSelector *NamespaceSelector
	ocCommands        *utils.OcCommands
	width             int
	height            int
}

type TickMsg time.Time

func NewApp(name, version, description string) *App {
	return &App{
		state: types.AppState{
			CurrentView:   types.ResourceView,
			ToolName:      name,
			Version:       version,
			ShowBanner:    true,
			PageSize:      50,
			CurrentPage:   0,
			ResourceType:  "pods",
			SortField:     "name",
			SortDirection: "asc",
		},
		bannerManager: &types.BannerManager{
			ToolName: name,
			Version:  version,
			Enabled:  true,
			Style: types.BannerStyle{
				Icon:        "🔮",
				ShowVersion: true,
				Compact:     true,
			},
		},
		paginator: &types.ResourcePaginator{
			PageSize:    50,
			CurrentPage: 0,
		},
		contextSelector: NewContextSelector(),
		ocCommands:      utils.NewOcCommands(),
	}
}

func (a *App) Run() error {
	if err := a.initialize(); err != nil {
		return err
	}

	p := tea.NewProgram(a, tea.WithAltScreen(), tea.WithMouseCellMotion())
	_, err := p.Run()
	return err
}

func (a *App) initialize() error {
	// Initialize config manager - make it optional
	configManager, err := config.NewConfigManager()
	if err != nil {
		// Continue without kubeconfig - we'll use oc commands directly
		a.config = &config.ConfigManager{
			Settings: &config.UserSettings{
				RefreshInterval: 30 * time.Second,
			},
		}
	} else {
		a.config = configManager
	}

	// Create Kubernetes client using kubeconfig - make it optional
	client, err := client.NewKubernetesClient("")
	if err != nil {
		// Continue without kubernetes client - we'll use oc commands directly
		a.client = nil
	} else {
		a.client = client
	}

	a.toolDetector = utils.NewToolDetector()

	// Get current context and namespace from oc command
	_, currentContext, err := a.ocCommands.GetContexts()
	if err == nil && currentContext != "" {
		a.state.Context = currentContext
	} else {
		a.state.Context = "unknown"
	}

	// Get current namespace from oc project command
	currentNamespace, err := a.ocCommands.GetCurrentNamespace()
	if err == nil && currentNamespace != "" {
		a.state.Namespace = currentNamespace
	} else {
		a.state.Namespace = "default"
	}

	// Set the tool to the one detected by OcCommands
	a.state.Tool = a.ocCommands.GetTool()

	// Try to get current user from kubeconfig context
	a.state.User = "user"
	if a.config.KubeConfig != nil && a.config.CurrentContext != "" {
		if context, exists := a.config.KubeConfig.Contexts[a.config.CurrentContext]; exists {
			if context.AuthInfo != "" {
				a.state.User = context.AuthInfo
			}
		}
	}

	return nil
}

func (a *App) Init() tea.Cmd {
	// Auto-fetch resources on startup and setup tick timer
	return tea.Batch(
		a.fetchResources,
		tea.Tick(a.config.Settings.RefreshInterval, func(t time.Time) tea.Msg {
			return TickMsg(t)
		}),
	)
}

func (a *App) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		a.width = msg.Width
		a.height = msg.Height
		return a, nil

	case tea.KeyMsg:
		return a.handleKeyPress(msg)

	case tea.MouseMsg:
		return a.handleMouseEvent(msg)

	case TickMsg:
		// Just reschedule the tick, don't auto-refresh resources
		return a, tea.Tick(a.config.Settings.RefreshInterval, func(t time.Time) tea.Msg {
			return TickMsg(t)
		})

	case resourcesMsg:
		a.state.Resources = msg.resources
		a.state.TotalItems = len(msg.resources)
		a.state.TotalPages = (len(msg.resources) + a.state.PageSize - 1) / a.state.PageSize
		a.state.LastUpdate = time.Now()
		return a, nil

	case describeMsg:
		a.state.DescribeOutput = msg.output
		return a, nil

	case tagsMsg:
		a.state.TagsOutput = msg.output
		return a, nil

	case containerResourcesMsg:
		a.state.ResourcesOutput = msg.output
		return a, nil

	case yamlMsg:
		a.state.YamlOutput = msg.output
		return a, nil

	case eventsMsg:
		a.state.EventsOutput = msg.output
		return a, nil

	case diagramMsg:
		a.state.DiagramOutput = msg.output
		return a, nil

	case FeedbackSubmittedMsg:
		a.state.FeedbackSubmitting = false
		if msg.success {
			// Feedback submitted successfully, quit the app
			return a, tea.Quit
		} else {
			// Error submitting feedback, but still quit
			return a, tea.Quit
		}
	}

	return a, nil
}

func (a *App) View() string {
	if a.width == 0 || a.height == 0 {
		return "Loading..."
	}

	switch a.state.CurrentView {
	case types.HelpView:
		return a.renderHelpView()
	case types.ContextView:
		return a.contextSelector.Render(a.width, a.height)
	case types.NamespaceView:
		return a.namespaceSelector.Render(a.width, a.height)
	case types.DescribeView:
		return a.renderDescribeView()
	case types.TagsView:
		return a.renderTagsView()
	case types.ResourcesView:
		return a.renderResourcesView()
	case types.YamlView:
		return a.renderYamlView()
	case types.EventsView:
		return a.renderEventsView()
	case types.DiagramView:
		return a.renderDiagramView()
	case types.FeedbackView:
		return a.renderFeedbackView()
	default:
		return a.renderMainView()
	}
}

func (a *App) renderMainView() string {
	var sections []string

	if a.state.ShowBanner {
		sections = append(sections, a.renderBanner())
	}

	sections = append(sections,
		a.renderContextInfo(),
		a.renderResourceTypeSelector(),
		a.renderResourceInfo(),
		a.renderResourceTable(),
		a.renderPagination(),
		a.renderKeyOptions(),
		a.renderStatusBar(),
	)

	content := lipgloss.JoinVertical(lipgloss.Left, sections...)

	style := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("240")).
		Padding(1).
		Width(a.width - 2).
		Height(a.height - 2)

	return style.Render(content)
}

func (a *App) renderBanner() string {
	// ASCII art for k8sGO (new design)
	asciiArt := `
 __   ___  _______    ________  _______     ______    
|/"| /  ")/"  _  \\  /"       )/" _   "|   /    " \   
(: |/   /|:  _ /  :|(:   \___/(: ( \___)  // ____  \  
|    __/  \___/___/  \___  \   \/ \      /  /    ) :) 
(// _  \  //  /_ \\   __/  \\  //  \ ___(: (____/ //  
|: | \  \|:  /_   :| /" \   :)(:   _(  _|\        /   
(__|  \__)\_______/ (_______/  \_______)  \"_____/    
                                                      `

	// Version and tool info
	toolName := "Kubernetes & OpenShift"
	if a.state.Tool == "kubectl" {
		toolName = "Kubernetes"
	} else if a.state.Tool == "oc" {
		toolName = "OpenShift & Kubernetes"
	}
	infoLine := fmt.Sprintf("v%s │ %s CLI Tool", a.bannerManager.Version, toolName)
	creditLine := "Developed By - GNA"

	// Style for ASCII art
	artStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("39")).
		Align(lipgloss.Center).
		Width(a.width - 4)

	// Style for info line
	infoStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("245")).
		Align(lipgloss.Center).
		Width(a.width - 4)

	// Style for credit line
	creditStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("242")).
		Align(lipgloss.Center).
		Width(a.width - 4)

	return lipgloss.JoinVertical(lipgloss.Center,
		artStyle.Render(asciiArt),
		infoStyle.Render(infoLine),
		creditStyle.Render(creditLine),
	)
}

func (a *App) renderContextInfo() string {
	// Context info is now shown in the resource info section, so this can be empty
	// or just show essential cluster connection status
	return ""
}

// getAvailableResourceTypes returns resource types based on the detected CLI tool
func (a *App) getAvailableResourceTypes() []string {
	// Base resource types that work with both oc and kubectl
	types := []string{"pods", "services", "deployments", "pvc"}

	// Add ImageStreams only if using oc (OpenShift)
	if a.state.Tool == "oc" {
		types = append(types, "is")
	}

	// Add remaining common types
	types = append(types, "secrets", "configmaps", "events")

	return types
}

// handleResourceTypeSwitchByIndex handles resource type switching by numeric key
func (a *App) handleResourceTypeSwitchByIndex(key string) (tea.Model, tea.Cmd) {
	// Convert key to index (1-8 -> 0-7)
	index := int(key[0] - '1')

	types := a.getAvailableResourceTypes()
	if index >= 0 && index < len(types) {
		return a.handleResourceTypeSwitch(types[index])
	}

	// Invalid index, do nothing
	return a, nil
}

func (a *App) renderResourceTypeSelector() string {
	types := a.getAvailableResourceTypes()
	var parts []string

	for i, resourceType := range types {
		keyNum := fmt.Sprintf("%d", i+1)
		var part string
		if resourceType == a.state.ResourceType {
			// Highlight current selection
			part = fmt.Sprintf("[%s] %s", keyNum, strings.ToUpper(resourceType))
			styled := lipgloss.NewStyle().
				Foreground(lipgloss.Color("39")).
				Bold(true).
				Render(part)
			parts = append(parts, styled)
		} else {
			part = fmt.Sprintf("[%s] %s", keyNum, resourceType)
			styled := lipgloss.NewStyle().
				Foreground(lipgloss.Color("246")).
				Render(part)
			parts = append(parts, styled)
		}
	}

	line := "Resource Type: " + strings.Join(parts, " │ ")

	style := lipgloss.NewStyle().
		Padding(0, 1)

	return style.Render(line)
}

func (a *App) renderResourceInfo() string {
	// Get the resource type display name
	resourceName := ""
	switch a.state.ResourceType {
	case "pods":
		resourceName = "Pods"
	case "services":
		resourceName = "Services"
	case "deployments":
		resourceName = "Deployments"
	case "pvc":
		resourceName = "PersistentVolumeClaims"
	case "is":
		resourceName = "ImageStreams"
	case "secrets":
		resourceName = "Secrets"
	case "configmaps":
		resourceName = "ConfigMaps"
	case "events":
		resourceName = "Events"
	default:
		resourceName = strings.Title(a.state.ResourceType)
	}

	line := fmt.Sprintf("Resource: %s │ Count: %d │ Namespace: %s │ Tool: %s",
		resourceName, len(a.state.Resources), a.state.Namespace, a.state.Tool)

	style := lipgloss.NewStyle().
		Foreground(lipgloss.Color("39")).
		Bold(true).
		Padding(0, 1)

	return style.Render(line)
}

func (a *App) renderKeyOptions() string {
	options := []string{
		"[c] Context",
		"[n] Namespace",
		"[1-8] Resource Type",
		"[i] Info/Describe",
		"[d] Diagram",
		"[y] YAML",
		"[e] Events",
		"[t] Tags",
		"[m] Memory",
		"[r] Refresh",
		"[x] Copy",
		"[?] Help",
		"[q] Quit",
	}

	line := strings.Join(options, " │ ")

	style := lipgloss.NewStyle().
		Foreground(lipgloss.Color("240")).
		Background(lipgloss.Color("235")).
		Padding(0, 1).
		Width(a.width - 4)

	return style.Render(line)
}

func (a *App) renderDescribeView() string {
	var sections []string

	// No banner for describe view to maximize content space

	// Resource info at top
	resourceInfo := fmt.Sprintf("Resource: %s | Type: %s | Tool: %s",
		a.state.SelectedResource.Name,
		strings.Title(a.state.ResourceType), a.state.Tool)

	resourceInfoStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("245")).
		Padding(0, 1)

	sections = append(sections, resourceInfoStyle.Render(resourceInfo))

	// Title
	title := fmt.Sprintf("Describe %s: %s",
		strings.Title(a.state.ResourceType),
		a.state.SelectedResource.Name)

	titleStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("39")).
		Background(lipgloss.Color("236")).
		Padding(0, 2).
		Width(a.width - 4).
		Align(lipgloss.Center)

	// Get scrollable content
	content := a.getScrollableDescribeContent()

	// Scrollable content area
	contentStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("252")).
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("240")).
		Padding(1).
		Width(a.width - 4).
		Height(a.height - 8)

	// Instructions with scroll info
	scrollInfo := a.getScrollInfo()
	instructions := fmt.Sprintf("Press [ESC] to return │ [j/k] scroll │ [d/u] page │ [g/G] top/bottom │ %s", scrollInfo)
	instructStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("240")).
		Background(lipgloss.Color("235")).
		Padding(0, 1).
		Width(a.width - 4).
		Align(lipgloss.Center)

	sections = append(sections,
		titleStyle.Render(title),
		"",
		contentStyle.Render(content),
		"",
		instructStyle.Render(instructions),
	)

	return lipgloss.JoinVertical(lipgloss.Left, sections...)
}

func (a *App) renderTagsView() string {
	var sections []string

	// Add banner
	if a.state.ShowBanner {
		sections = append(sections, a.renderBanner())
	}

	// Resource info at top
	resourceInfo := fmt.Sprintf("Resource: %s | Type: %s | Tool: %s",
		a.state.SelectedResource.Name,
		strings.Title(a.state.ResourceType), a.state.Tool)

	resourceInfoStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("245")).
		Padding(0, 1)

	sections = append(sections, resourceInfoStyle.Render(resourceInfo))

	// Title
	title := fmt.Sprintf("Tags for ImageStream: %s", a.state.SelectedResource.Name)

	titleStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("39")).
		Background(lipgloss.Color("236")).
		Padding(0, 2).
		Width(a.width - 4).
		Align(lipgloss.Center)

	// Get scrollable content
	content := a.getScrollableTagsContent()

	// Scrollable content area
	contentStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("252")).
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("240")).
		Padding(1).
		Width(a.width - 4).
		Height(a.height - 8)

	// Instructions with scroll info
	scrollInfo := a.getTagsScrollInfo()
	instructions := fmt.Sprintf("Press [ESC] to return │ [j/k] scroll │ [d/u] page │ [g/G] top/bottom │ %s", scrollInfo)
	instructStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("240")).
		Background(lipgloss.Color("235")).
		Padding(0, 1).
		Width(a.width - 4).
		Align(lipgloss.Center)

	sections = append(sections,
		titleStyle.Render(title),
		"",
		contentStyle.Render(content),
		"",
		instructStyle.Render(instructions),
	)

	return lipgloss.JoinVertical(lipgloss.Left, sections...)
}

func (a *App) getScrollableTagsContent() string {
	if a.state.TagsOutput == "" {
		return "Loading ImageStream tags..."
	}

	lines := strings.Split(a.state.TagsOutput, "\n")
	viewHeight := a.getTagsViewHeight() - 2 // Account for padding

	// Calculate visible range
	startLine := a.state.TagsScrollOffset
	endLine := startLine + viewHeight

	if startLine >= len(lines) {
		startLine = len(lines) - 1
		if startLine < 0 {
			startLine = 0
		}
	}

	if endLine > len(lines) {
		endLine = len(lines)
	}

	if startLine >= endLine {
		return "-- End of content --"
	}

	visibleLines := lines[startLine:endLine]

	// Pad with empty lines if content is shorter than view
	for len(visibleLines) < viewHeight {
		visibleLines = append(visibleLines, "")
	}

	content := strings.Join(visibleLines, "\n")

	// Apply selection highlighting if active
	if a.state.SelectionActive && a.state.CurrentView == types.TagsView {
		content = a.renderWithSelectionHighlight(content, 3) // viewStartY = 3
	}

	return content
}

func (a *App) getTagsScrollInfo() string {
	if a.state.TagsOutput == "" {
		return "Loading..."
	}

	lines := strings.Split(a.state.TagsOutput, "\n")
	totalLines := len(lines)
	viewHeight := a.getTagsViewHeight() - 2
	currentLine := a.state.TagsScrollOffset + 1

	if totalLines <= viewHeight {
		return "All"
	}

	endLine := currentLine + viewHeight - 1
	if endLine > totalLines {
		endLine = totalLines
	}

	return fmt.Sprintf("Lines %d-%d of %d", currentLine, endLine, totalLines)
}

func (a *App) getScrollableDescribeContent() string {
	if a.state.DescribeOutput == "" {
		return "Loading resource description..."
	}

	lines := strings.Split(a.state.DescribeOutput, "\n")
	viewHeight := a.getDescribeViewHeight() - 2 // Account for padding

	// Calculate visible range
	startLine := a.state.DescribeScrollOffset
	endLine := startLine + viewHeight

	if startLine >= len(lines) {
		startLine = len(lines) - 1
		if startLine < 0 {
			startLine = 0
		}
	}

	if endLine > len(lines) {
		endLine = len(lines)
	}

	if startLine >= endLine {
		return "-- End of content --"
	}

	visibleLines := lines[startLine:endLine]

	// Pad with empty lines if content is shorter than view
	for len(visibleLines) < viewHeight {
		visibleLines = append(visibleLines, "")
	}

	content := strings.Join(visibleLines, "\n")

	return content
}

func (a *App) getScrollInfo() string {
	if a.state.DescribeOutput == "" {
		return "Loading..."
	}

	lines := strings.Split(a.state.DescribeOutput, "\n")
	totalLines := len(lines)
	viewHeight := a.getDescribeViewHeight() - 2
	currentLine := a.state.DescribeScrollOffset + 1

	if totalLines <= viewHeight {
		return "All"
	}

	endLine := currentLine + viewHeight - 1
	if endLine > totalLines {
		endLine = totalLines
	}

	return fmt.Sprintf("Lines %d-%d of %d", currentLine, endLine, totalLines)
}

func (a *App) getDescribeViewHeight() int {
	// Account for title, instructions, padding, and borders
	return a.height - 8
}

type resourcesMsg struct {
	resources []types.Resource
}

type describeMsg struct {
	output string
}

type tagsMsg struct {
	output string
}

type containerResourcesMsg struct {
	output string
}

type yamlMsg struct {
	output string
}

type eventsMsg struct {
	output string
}

type diagramMsg struct {
	output string
}

// ImageStream JSON structures for native parsing
type ImageStreamStatus struct {
	Tags []ImageStreamTag `json:"tags"`
}

type ImageStreamTag struct {
	Tag   string               `json:"tag"`
	Items []ImageStreamTagItem `json:"items"`
}

type ImageStreamTagItem struct {
	Created string `json:"created"`
}

type ImageStreamJSON struct {
	Status ImageStreamStatus `json:"status"`
}

func (a *App) fetchResources() tea.Msg {
	var resourceData []utils.PodInfo
	var err error
	var resourceTypeName string

	// Get resources based on selected type
	switch a.state.ResourceType {
	case "services":
		resourceData, err = a.ocCommands.GetServices(a.state.Namespace)
		resourceTypeName = "Services"
	case "deployments":
		resourceData, err = a.ocCommands.GetDeployments(a.state.Namespace)
		resourceTypeName = "Deployments"
	case "pvc":
		resourceData, err = a.ocCommands.GetPVC(a.state.Namespace)
		resourceTypeName = "PersistentVolumeClaims"
	case "is":
		resourceData, err = a.ocCommands.GetImageStreams(a.state.Namespace)
		resourceTypeName = "ImageStreams"
	case "secrets":
		resourceData, err = a.ocCommands.GetSecrets(a.state.Namespace)
		resourceTypeName = "Secrets"
	case "configmaps":
		resourceData, err = a.ocCommands.GetConfigMaps(a.state.Namespace)
		resourceTypeName = "ConfigMaps"
	case "events":
		resourceData, err = a.ocCommands.GetEvents(a.state.Namespace)
		resourceTypeName = "Events"
	default: // pods
		resourceData, err = a.ocCommands.GetPods(a.state.Namespace)
		resourceTypeName = "Pods"
	}

	if err != nil {
		return resourcesMsg{resources: []types.Resource{
			{
				Name:      fmt.Sprintf("Error fetching %s", a.state.ResourceType),
				Namespace: a.state.Namespace,
				Type:      "Error",
				Status:    err.Error(),
				Ready:     "N/A",
				Age:       0,
			},
		}}
	}

	// Convert resource data to display format
	var resources []types.Resource
	for _, item := range resourceData {
		resources = append(resources, types.Resource{
			Name:      item.Name,
			Namespace: a.state.Namespace,
			Type:      resourceTypeName,
			Status:    item.Status,
			Ready:     item.Ready,
			Restarts:  parseRestarts(item.Restarts),
			Age:       parseAge(item.Age),
			CPU:       "N/A",
			Memory:    "N/A",
			// PVC-specific fields
			Volume:       item.Volume,
			Capacity:     item.Capacity,
			AccessModes:  item.AccessModes,
			StorageClass: item.StorageClass,
		})
	}

	// Apply manual sorting based on user preference
	a.sortResources(resources)

	if len(resources) == 0 {
		resources = append(resources, types.Resource{
			Name:      fmt.Sprintf("No %s found", a.state.ResourceType),
			Namespace: a.state.Namespace,
			Type:      "Info",
			Status:    "Empty",
			Ready:     "N/A",
			Age:       0,
		})
	}

	return resourcesMsg{resources: resources}
}

// sortResources sorts the resources array based on the current sort settings
func (a *App) sortResources(resources []types.Resource) {
	if len(resources) <= 1 {
		return
	}

	sort.Slice(resources, func(i, j int) bool {
		var less bool

		switch a.state.SortField {
		case "name":
			less = resources[i].Name < resources[j].Name
		case "age":
			less = resources[i].Age < resources[j].Age
		case "status":
			less = resources[i].Status < resources[j].Status
		case "ready":
			less = resources[i].Ready < resources[j].Ready
		case "restarts":
			less = resources[i].Restarts < resources[j].Restarts
		default:
			less = resources[i].Name < resources[j].Name // Default to name
		}

		if a.state.SortDirection == "desc" {
			return !less
		}
		return less
	})
}

func (a *App) fetchResourceDescription() tea.Msg {
	// Get the resource type for the oc describe command
	var resourceType string
	switch a.state.ResourceType {
	case "pods":
		resourceType = "pod"
	case "services":
		resourceType = "service"
	case "deployments":
		resourceType = "deployment"
	case "pvc":
		resourceType = "pvc"
	case "is":
		resourceType = "imagestream"
	case "secrets":
		resourceType = "secret"
	case "configmaps":
		resourceType = "configmap"
	case "events":
		resourceType = "event"
	default:
		resourceType = "pod"
	}

	// Execute oc describe command
	cmd := exec.Command("oc", "describe", resourceType, a.state.SelectedResource.Name, "-n", a.state.Namespace)
	output, err := cmd.Output()

	if err != nil {
		return describeMsg{output: fmt.Sprintf("Error describing %s: %s", a.state.SelectedResource.Name, err.Error())}
	}

	return describeMsg{output: string(output)}
}

func (a *App) fetchResourceTags() tea.Msg {
	// Only works for ImageStreams
	if a.state.ResourceType != "is" {
		return tagsMsg{output: "Tags are only available for ImageStreams"}
	}

	// Execute oc command to get ImageStream JSON
	cmd := exec.Command("oc", "get", "is", a.state.SelectedResource.Name, "-n", a.state.Namespace, "-o", "json")
	output, err := cmd.Output()

	if err != nil {
		return tagsMsg{output: fmt.Sprintf("Error getting ImageStream %s: %s", a.state.SelectedResource.Name, err.Error())}
	}

	// Parse JSON using native Go
	var imageStream ImageStreamJSON
	if err := json.Unmarshal(output, &imageStream); err != nil {
		return tagsMsg{output: fmt.Sprintf("Error parsing ImageStream JSON: %s", err.Error())}
	}

	// Check if there are any tags
	if len(imageStream.Status.Tags) == 0 {
		return tagsMsg{output: "No tags found for this ImageStream"}
	}

	// Process tags and sort by creation date (newest first)
	type TagWithDate struct {
		Tag        string
		Created    time.Time
		CreatedStr string
	}

	var tags []TagWithDate
	for _, tag := range imageStream.Status.Tags {
		if len(tag.Items) > 0 {
			// Parse the creation timestamp
			created, err := time.Parse(time.RFC3339, tag.Items[0].Created)
			if err != nil {
				// If parsing fails, use current time as fallback
				created = time.Now()
			}

			// Format the date as "YYYY-MM-DD HH:MM:SS"
			createdStr := created.Format("2006-01-02 15:04:05")

			tags = append(tags, TagWithDate{
				Tag:        tag.Tag,
				Created:    created,
				CreatedStr: createdStr,
			})
		}
	}

	// Sort by creation date (newest first)
	sort.Slice(tags, func(i, j int) bool {
		return tags[i].Created.After(tags[j].Created)
	})

	// Format output as a table
	var result strings.Builder
	result.WriteString("TAG        CREATED\n")
	result.WriteString("----------------------------------------\n")

	for _, tag := range tags {
		// Use fixed-width formatting for better alignment
		result.WriteString(fmt.Sprintf("%-10s %s\n", tag.Tag, tag.CreatedStr))
	}

	return tagsMsg{output: result.String()}
}

func (a *App) fetchResourceRequests() tea.Msg {
	// Only works for Pods and Deployments
	if a.state.ResourceType != "pods" && a.state.ResourceType != "deployments" {
		return containerResourcesMsg{output: "Resource requests/limits are only available for Pods and Deployments"}
	}

	var cmd *exec.Cmd
	if a.state.ResourceType == "pods" {
		// Get container resource requests and limits for a pod
		cmd = exec.Command("oc", "get", "pod", a.state.SelectedResource.Name, "-n", a.state.Namespace, "-o", "jsonpath={range .spec.containers[*]}{.name}{\"\\t\"}{.resources.requests.cpu}{\"\\t\"}{.resources.requests.memory}{\"\\t\"}{.resources.limits.cpu}{\"\\t\"}{.resources.limits.memory}{\"\\n\"}{end}")
	} else {
		// Get container resource requests and limits for a deployment
		cmd = exec.Command("oc", "get", "deployment", a.state.SelectedResource.Name, "-n", a.state.Namespace, "-o", "jsonpath={range .spec.template.spec.containers[*]}{.name}{\"\\t\"}{.resources.requests.cpu}{\"\\t\"}{.resources.requests.memory}{\"\\t\"}{.resources.limits.cpu}{\"\\t\"}{.resources.limits.memory}{\"\\n\"}{end}")
	}

	output, err := cmd.Output()
	if err != nil {
		return containerResourcesMsg{output: fmt.Sprintf("Error getting container resources for %s: %s", a.state.SelectedResource.Name, err.Error())}
	}

	resourceOutput := string(output)
	if strings.TrimSpace(resourceOutput) == "" {
		return containerResourcesMsg{output: "No container resource requests/limits configured for this resource"}
	}

	// Format the output as a table
	var result strings.Builder
	result.WriteString("CONTAINER        CPU REQ    MEM REQ     CPU LIM    MEM LIM\n")
	result.WriteString("----------------------------------------------------------------\n")

	lines := strings.Split(strings.TrimSpace(resourceOutput), "\n")
	for _, line := range lines {
		if line != "" {
			parts := strings.Split(line, "\t")
			if len(parts) >= 5 {
				// Pad and format columns
				container := padString(parts[0], 16)
				cpuReq := padString(parts[1], 10)
				memReq := padString(parts[2], 11)
				cpuLim := padString(parts[3], 10)
				memLim := padString(parts[4], 10)

				result.WriteString(fmt.Sprintf("%s %s %s %s %s\n", container, cpuReq, memReq, cpuLim, memLim))
			}
		}
	}

	return containerResourcesMsg{output: result.String()}
}

func padString(s string, length int) string {
	if s == "" || s == "<none>" {
		s = "-"
	}
	if len(s) > length {
		return s[:length-3] + "..."
	}
	for len(s) < length {
		s += " "
	}
	return s
}

func parseRestarts(restartStr string) int {
	var restarts int
	fmt.Sscanf(restartStr, "%d", &restarts)
	return restarts
}

func parseAge(ageStr string) time.Duration {
	// Simple age parsing - you can enhance this
	if strings.HasSuffix(ageStr, "d") {
		var days int
		fmt.Sscanf(ageStr, "%dd", &days)
		return time.Duration(days) * 24 * time.Hour
	} else if strings.HasSuffix(ageStr, "h") {
		var hours int
		fmt.Sscanf(ageStr, "%dh", &hours)
		return time.Duration(hours) * time.Hour
	} else if strings.HasSuffix(ageStr, "m") {
		var minutes int
		fmt.Sscanf(ageStr, "%dm", &minutes)
		return time.Duration(minutes) * time.Minute
	} else if strings.HasSuffix(ageStr, "s") {
		var seconds int
		fmt.Sscanf(ageStr, "%ds", &seconds)
		return time.Duration(seconds) * time.Second
	}
	return 0
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// renderWithSelectionHighlight applies visual highlighting to selected text
func (a *App) renderWithSelectionHighlight(content string, viewStartY int) string {
	if !a.state.SelectionActive {
		return content
	}

	lines := strings.Split(content, "\n")

	// Normalize selection coordinates
	startX, startY := a.state.SelectionStartX, a.state.SelectionStartY
	endX, endY := a.state.SelectionEndX, a.state.SelectionEndY

	if startY > endY || (startY == endY && startX > endX) {
		startX, endX = endX, startX
		startY, endY = endY, startY
	}

	// Apply highlighting to each line
	for i, line := range lines {
		screenY := viewStartY + i

		if screenY >= startY && screenY <= endY {
			var highlightedLine string

			if startY == endY && startY == screenY {
				// Single line selection
				if startX < len(line) && endX > startX {
					beforeSelection := ""
					if startX > 0 {
						beforeSelection = line[:startX]
					}

					selectionEnd := endX
					if selectionEnd > len(line) {
						selectionEnd = len(line)
					}

					selectedText := line[startX:selectionEnd]
					afterSelection := ""
					if selectionEnd < len(line) {
						afterSelection = line[selectionEnd:]
					}

					// Style the selected portion with inverted colors
					selectedStyled := lipgloss.NewStyle().
						Foreground(lipgloss.Color("0")).
						Background(lipgloss.Color("39")).
						Render(selectedText)

					highlightedLine = beforeSelection + selectedStyled + afterSelection
				} else {
					highlightedLine = line
				}
			} else {
				// Multi-line selection
				if screenY == startY {
					// First line of selection
					if startX < len(line) {
						beforeSelection := ""
						if startX > 0 {
							beforeSelection = line[:startX]
						}
						selectedText := line[startX:]

						selectedStyled := lipgloss.NewStyle().
							Foreground(lipgloss.Color("0")).
							Background(lipgloss.Color("39")).
							Render(selectedText)

						highlightedLine = beforeSelection + selectedStyled
					} else {
						highlightedLine = line
					}
				} else if screenY == endY {
					// Last line of selection
					selectionEnd := endX
					if selectionEnd > len(line) {
						selectionEnd = len(line)
					}

					if selectionEnd > 0 {
						selectedText := line[:selectionEnd]
						afterSelection := ""
						if selectionEnd < len(line) {
							afterSelection = line[selectionEnd:]
						}

						selectedStyled := lipgloss.NewStyle().
							Foreground(lipgloss.Color("0")).
							Background(lipgloss.Color("39")).
							Render(selectedText)

						highlightedLine = selectedStyled + afterSelection
					} else {
						highlightedLine = line
					}
				} else {
					// Middle lines - highlight entire line
					if line != "" {
						highlightedLine = lipgloss.NewStyle().
							Foreground(lipgloss.Color("0")).
							Background(lipgloss.Color("39")).
							Render(line)
					} else {
						// Empty line - still highlight with a space to show selection
						highlightedLine = lipgloss.NewStyle().
							Foreground(lipgloss.Color("0")).
							Background(lipgloss.Color("39")).
							Render(" ")
					}
				}
			}

			lines[i] = highlightedLine
		}
	}

	return strings.Join(lines, "\n")
}

func (a *App) renderResourcesView() string {
	var sections []string

	// Add banner
	if a.state.ShowBanner {
		sections = append(sections, a.renderBanner())
	}

	// Resource info at top
	resourceInfo := fmt.Sprintf("Resource: %s | Type: %s | Tool: %s",
		a.state.SelectedResource.Name,
		strings.Title(a.state.ResourceType), a.state.Tool)

	resourceInfoStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("245")).
		Padding(0, 1)

	sections = append(sections, resourceInfoStyle.Render(resourceInfo))

	// Title
	title := fmt.Sprintf("Container Resources for %s: %s",
		strings.Title(a.state.ResourceType),
		a.state.SelectedResource.Name)

	titleStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("39")).
		Background(lipgloss.Color("236")).
		Padding(0, 2).
		Width(a.width - 4).
		Align(lipgloss.Center)

	// Get scrollable content
	content := a.getScrollableResourcesContent()

	// Scrollable content area
	contentStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("252")).
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("240")).
		Padding(1).
		Width(a.width - 4).
		Height(a.height - 8)

	// Instructions with scroll info
	scrollInfo := a.getResourcesScrollInfo()
	instructions := fmt.Sprintf("Press [ESC] to return │ [j/k] scroll │ [d/u] page │ [g/G] top/bottom │ %s", scrollInfo)
	instructStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("240")).
		Background(lipgloss.Color("235")).
		Padding(0, 1).
		Width(a.width - 4).
		Align(lipgloss.Center)

	sections = append(sections,
		titleStyle.Render(title),
		"",
		contentStyle.Render(content),
		"",
		instructStyle.Render(instructions),
	)

	return lipgloss.JoinVertical(lipgloss.Left, sections...)
}

func (a *App) getScrollableResourcesContent() string {
	if a.state.ResourcesOutput == "" {
		return "Loading container resources..."
	}

	lines := strings.Split(a.state.ResourcesOutput, "\n")
	viewHeight := a.getResourcesViewHeight() - 2 // Account for padding

	// Calculate visible range
	startLine := a.state.ResourcesScrollOffset
	endLine := startLine + viewHeight

	if startLine >= len(lines) {
		startLine = len(lines) - 1
		if startLine < 0 {
			startLine = 0
		}
	}

	if endLine > len(lines) {
		endLine = len(lines)
	}

	if startLine >= endLine {
		return "-- End of content --"
	}

	visibleLines := lines[startLine:endLine]

	// Pad with empty lines if content is shorter than view
	for len(visibleLines) < viewHeight {
		visibleLines = append(visibleLines, "")
	}

	content := strings.Join(visibleLines, "\n")

	// Apply selection highlighting if active
	if a.state.SelectionActive && a.state.CurrentView == types.ResourcesView {
		content = a.renderWithSelectionHighlight(content, 3) // viewStartY = 3
	}

	return content
}

func (a *App) getResourcesScrollInfo() string {
	if a.state.ResourcesOutput == "" {
		return "Loading..."
	}

	lines := strings.Split(a.state.ResourcesOutput, "\n")
	totalLines := len(lines)
	viewHeight := a.getResourcesViewHeight() - 2
	currentLine := a.state.ResourcesScrollOffset + 1

	if totalLines <= viewHeight {
		return "All"
	}

	endLine := currentLine + viewHeight - 1
	if endLine > totalLines {
		endLine = totalLines
	}

	return fmt.Sprintf("Lines %d-%d of %d", currentLine, endLine, totalLines)
}

func (a *App) fetchResourceYaml() tea.Msg {
	// Get the resource type for the oc get command
	var resourceType string
	switch a.state.ResourceType {
	case "pods":
		resourceType = "pod"
	case "services":
		resourceType = "service"
	case "deployments":
		resourceType = "deployment"
	case "pvc":
		resourceType = "pvc"
	case "is":
		resourceType = "imagestream"
	case "secrets":
		resourceType = "secret"
	case "configmaps":
		resourceType = "configmap"
	case "events":
		resourceType = "event"
	default:
		resourceType = "pod"
	}

	// Execute oc get command with YAML output
	cmd := exec.Command("oc", "get", resourceType, a.state.SelectedResource.Name, "-n", a.state.Namespace, "-o", "yaml")
	output, err := cmd.Output()

	if err != nil {
		return yamlMsg{output: fmt.Sprintf("Error getting YAML for %s: %s", a.state.SelectedResource.Name, err.Error())}
	}

	return yamlMsg{output: string(output)}
}

func (a *App) renderYamlView() string {
	var sections []string

	// No banner for YAML view to maximize content space

	// Resource info at top
	resourceInfo := fmt.Sprintf("Resource: %s | Type: %s | Tool: %s",
		a.state.SelectedResource.Name,
		strings.Title(a.state.ResourceType), a.state.Tool)

	resourceInfoStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("245")).
		Padding(0, 1)

	sections = append(sections, resourceInfoStyle.Render(resourceInfo))

	// Title
	title := fmt.Sprintf("YAML for %s: %s",
		strings.Title(a.state.ResourceType),
		a.state.SelectedResource.Name)

	titleStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("39")).
		Background(lipgloss.Color("236")).
		Padding(0, 2).
		Width(a.width - 4).
		Align(lipgloss.Center)

	// Get scrollable content
	content := a.getScrollableYamlContent()

	// Scrollable content area
	contentStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("252")).
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("240")).
		Padding(1).
		Width(a.width - 4).
		Height(a.height - 8)

	// Instructions with scroll info
	scrollInfo := a.getYamlScrollInfo()
	instructions := fmt.Sprintf("Press [ESC] to return │ [j/k] scroll │ [d/u] page │ [g/G] top/bottom │ [x] copy │ %s", scrollInfo)
	instructStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("240")).
		Background(lipgloss.Color("235")).
		Padding(0, 1).
		Width(a.width - 4).
		Align(lipgloss.Center)

	sections = append(sections,
		titleStyle.Render(title),
		"",
		contentStyle.Render(content),
		"",
		instructStyle.Render(instructions),
	)

	return lipgloss.JoinVertical(lipgloss.Left, sections...)
}

func (a *App) getScrollableYamlContent() string {
	if a.state.YamlOutput == "" {
		return "Loading YAML..."
	}

	lines := strings.Split(a.state.YamlOutput, "\n")
	viewHeight := a.getYamlViewHeight() - 2 // Account for padding

	// Calculate visible range
	startLine := a.state.YamlScrollOffset
	endLine := startLine + viewHeight

	if startLine >= len(lines) {
		startLine = len(lines) - 1
		if startLine < 0 {
			startLine = 0
		}
	}

	if endLine > len(lines) {
		endLine = len(lines)
	}

	if startLine >= endLine {
		return "-- End of content --"
	}

	visibleLines := lines[startLine:endLine]

	// Pad with empty lines if content is shorter than view
	for len(visibleLines) < viewHeight {
		visibleLines = append(visibleLines, "")
	}

	content := strings.Join(visibleLines, "\n")

	// Apply selection highlighting if active
	if a.state.SelectionActive && a.state.CurrentView == types.YamlView {
		content = a.renderWithSelectionHighlight(content, 3) // viewStartY = 3
	}

	return content
}

func (a *App) getYamlScrollInfo() string {
	if a.state.YamlOutput == "" {
		return "Loading..."
	}

	lines := strings.Split(a.state.YamlOutput, "\n")
	totalLines := len(lines)
	viewHeight := a.getYamlViewHeight() - 2
	currentLine := a.state.YamlScrollOffset + 1

	if totalLines <= viewHeight {
		return "All"
	}

	endLine := currentLine + viewHeight - 1
	if endLine > totalLines {
		endLine = totalLines
	}

	return fmt.Sprintf("Lines %d-%d of %d", currentLine, endLine, totalLines)
}

func (a *App) getYamlViewHeight() int {
	// Account for title, instructions, padding, and borders
	return a.height - 8
}

func (a *App) fetchResourceEvents() tea.Msg {
	// Get the resource type for the oc command
	var resourceType string
	switch a.state.ResourceType {
	case "pods":
		resourceType = "pod"
	case "services":
		resourceType = "service"
	case "deployments":
		resourceType = "deployment"
	case "pvc":
		resourceType = "pvc"
	case "is":
		resourceType = "imagestream"
	case "secrets":
		resourceType = "secret"
	case "configmaps":
		resourceType = "configmap"
	case "events":
		resourceType = "event"
	default:
		resourceType = "pod"
	}

	// Method 1: Try to get events related to this specific resource
	cmd := exec.Command("oc", "get", "events", "-n", a.state.Namespace,
		"--field-selector", fmt.Sprintf("involvedObject.name=%s,involvedObject.kind=%s",
			a.state.SelectedResource.Name, strings.ToUpper(string(resourceType[0]))+resourceType[1:]),
		"--sort-by=.lastTimestamp")
	output, err := cmd.Output()

	if err != nil || strings.TrimSpace(string(output)) == "" || strings.Contains(string(output), "No resources found") {
		// Method 2: Fallback - extract events from describe output
		return a.fetchEventsFromDescribe()
	}

	return eventsMsg{output: string(output)}
}

func (a *App) fetchEventsFromDescribe() tea.Msg {
	// Get the resource type for the oc describe command
	var resourceType string
	switch a.state.ResourceType {
	case "pods":
		resourceType = "pod"
	case "services":
		resourceType = "service"
	case "deployments":
		resourceType = "deployment"
	case "pvc":
		resourceType = "pvc"
	case "is":
		resourceType = "imagestream"
	case "secrets":
		resourceType = "secret"
	case "configmaps":
		resourceType = "configmap"
	case "events":
		resourceType = "event"
	default:
		resourceType = "pod"
	}

	// Execute oc describe command
	cmd := exec.Command("oc", "describe", resourceType, a.state.SelectedResource.Name, "-n", a.state.Namespace)
	output, err := cmd.Output()

	if err != nil {
		return eventsMsg{output: fmt.Sprintf("Error getting events for %s: %s", a.state.SelectedResource.Name, err.Error())}
	}

	// Extract the Events section from describe output
	describeOutput := string(output)
	lines := strings.Split(describeOutput, "\n")

	var eventsSection []string
	inEventsSection := false

	for _, line := range lines {
		if strings.HasPrefix(line, "Events:") {
			inEventsSection = true
			eventsSection = append(eventsSection, line)
			continue
		}

		if inEventsSection {
			// Continue until we hit another section (starts with capital letter and colon) or end
			if strings.Contains(line, ":") && len(line) > 0 &&
				strings.ToUpper(string(line[0])) == string(line[0]) &&
				!strings.HasPrefix(line, "  ") && !strings.HasPrefix(line, "\t") {
				// This looks like a new section header, stop here
				break
			}
			eventsSection = append(eventsSection, line)
		}
	}

	if len(eventsSection) == 0 {
		return eventsMsg{output: "No events found for this resource"}
	}

	return eventsMsg{output: strings.Join(eventsSection, "\n")}
}

func (a *App) fetchResourceDiagram() tea.Msg {
	// Only generate diagrams for deployments
	if a.state.ResourceType != "deployments" {
		return diagramMsg{output: "Diagrams are only available for deployments"}
	}

	// First get the YAML of the deployment
	cmd := exec.Command("oc", "get", "deployment", a.state.SelectedResource.Name, "-n", a.state.Namespace, "-o", "yaml")
	yamlOutput, err := cmd.Output()
	if err != nil {
		return diagramMsg{output: fmt.Sprintf("Error getting deployment YAML: %s", err.Error())}
	}

	// Generate ASCII flowchart from deployment YAML
	diagram := a.generateDeploymentDiagram(string(yamlOutput))
	return diagramMsg{output: diagram}
}

func (a *App) renderEventsView() string {
	var sections []string

	// No banner for events view to maximize content space

	// Resource info at top
	resourceInfo := fmt.Sprintf("Resource: %s | Type: %s | Tool: %s",
		a.state.SelectedResource.Name,
		strings.Title(a.state.ResourceType), a.state.Tool)

	resourceInfoStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("245")).
		Padding(0, 1)

	sections = append(sections, resourceInfoStyle.Render(resourceInfo))

	// Title
	title := fmt.Sprintf("Events for %s: %s",
		strings.Title(a.state.ResourceType),
		a.state.SelectedResource.Name)

	titleStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("39")).
		Background(lipgloss.Color("236")).
		Padding(0, 2).
		Width(a.width - 4).
		Align(lipgloss.Center)

	// Get scrollable content
	content := a.getScrollableEventsContent()

	// Scrollable content area
	contentStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("252")).
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("240")).
		Padding(1).
		Width(a.width - 4).
		Height(a.height - 8)

	// Instructions with scroll info
	scrollInfo := a.getEventsScrollInfo()
	instructions := fmt.Sprintf("Press [ESC] to return │ [j/k] scroll │ [d/u] page │ [g/G] top/bottom │ [x] copy │ %s", scrollInfo)
	instructStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("240")).
		Background(lipgloss.Color("235")).
		Padding(0, 1).
		Width(a.width - 4).
		Align(lipgloss.Center)

	sections = append(sections,
		titleStyle.Render(title),
		"",
		contentStyle.Render(content),
		"",
		instructStyle.Render(instructions),
	)

	return lipgloss.JoinVertical(lipgloss.Left, sections...)
}

func (a *App) renderDiagramView() string {
	var sections []string

	// Resource info at top
	resourceInfo := fmt.Sprintf("Resource: %s | Type: %s | Tool: %s",
		a.state.SelectedResource.Name,
		strings.Title(a.state.ResourceType), a.state.Tool)

	resourceInfoStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("245")).
		Padding(0, 1)

	sections = append(sections, resourceInfoStyle.Render(resourceInfo))

	// Title
	title := fmt.Sprintf("Deployment Diagram: %s", a.state.SelectedResource.Name)

	titleStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("39")).
		Background(lipgloss.Color("236")).
		Padding(0, 2).
		Width(a.width - 4).
		Align(lipgloss.Center)

	// Get scrollable content
	content := a.getScrollableDiagramContent()

	// Scrollable content area
	contentStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("252")).
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("240")).
		Padding(1).
		Width(a.width - 4).
		Height(a.height - 8)

	// Instructions with scroll info
	scrollInfo := fmt.Sprintf("Line %d-%d of %d",
		a.state.DiagramScrollOffset+1,
		min(a.state.DiagramScrollOffset+a.getDiagramViewHeight(), a.getDiagramLineCount()),
		a.getDiagramLineCount())

	instructions := fmt.Sprintf("[j/k] Scroll • [d/u] Page • [g/G] Top/Bottom • [x] Copy • [ESC] Back | %s", scrollInfo)

	instructStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("240")).
		Background(lipgloss.Color("235")).
		Padding(0, 1).
		Width(a.width - 4).
		Align(lipgloss.Center)

	sections = append(sections,
		titleStyle.Render(title),
		"",
		contentStyle.Render(content),
		"",
		instructStyle.Render(instructions),
	)

	return lipgloss.JoinVertical(lipgloss.Left, sections...)
}

func (a *App) getScrollableDiagramContent() string {
	if a.state.DiagramOutput == "" {
		return "Loading diagram..."
	}

	lines := strings.Split(a.state.DiagramOutput, "\n")
	viewHeight := a.getDiagramViewHeight() - 2 // Account for padding

	// Calculate visible range
	startLine := a.state.DiagramScrollOffset
	endLine := startLine + viewHeight

	if startLine >= len(lines) {
		startLine = len(lines) - 1
		if startLine < 0 {
			startLine = 0
		}
	}

	if endLine > len(lines) {
		endLine = len(lines)
	}

	visibleLines := lines[startLine:endLine]
	return strings.Join(visibleLines, "\n")
}

func (a *App) getDiagramLineCount() int {
	if a.state.DiagramOutput == "" {
		return 0
	}
	return len(strings.Split(a.state.DiagramOutput, "\n"))
}

func (a *App) generateDeploymentDiagram(yamlContent string) string {
	// Parse deployment YAML structure properly
	lines := strings.Split(yamlContent, "\n")

	var name, namespace, image, replicas string
	var selector map[string]string
	var ports []string
	var resources map[string]string

	name = "Unknown"
	namespace = "Unknown"
	replicas = "1"
	image = "Unknown"
	selector = make(map[string]string)
	resources = make(map[string]string)

	// Simplified YAML parsing - look for specific patterns
	inMetadata := false
	inSpec := false
	inSelector := false
	inMatchLabels := false
	inContainers := false
	inPorts := false
	inResources := false

	for _, line := range lines {
		originalLine := line
		line = strings.TrimSpace(line)
		indent := len(originalLine) - len(strings.TrimLeft(originalLine, " "))

		// Track sections
		if line == "metadata:" {
			inMetadata = true
			inSpec = false
			continue
		} else if line == "spec:" {
			inMetadata = false
			inSpec = true
			inSelector = false
			inMatchLabels = false
			continue
		}

		// Parse metadata
		if inMetadata {
			if strings.HasPrefix(line, "name:") && indent <= 4 {
				name = strings.TrimSpace(strings.TrimPrefix(line, "name:"))
				name = strings.Trim(name, "\"'")
			} else if strings.HasPrefix(line, "namespace:") && indent <= 4 {
				namespace = strings.TrimSpace(strings.TrimPrefix(line, "namespace:"))
				namespace = strings.Trim(namespace, "\"'")
			}
		}

		// Parse spec section
		if inSpec {
			if strings.HasPrefix(line, "replicas:") && indent <= 4 {
				replicas = strings.TrimSpace(strings.TrimPrefix(line, "replicas:"))
			} else if line == "selector:" && indent <= 4 {
				inSelector = true
				inMatchLabels = false
			} else if inSelector && line == "matchLabels:" {
				inMatchLabels = true
			} else if inMatchLabels && strings.Contains(line, ":") && indent >= 4 {
				parts := strings.SplitN(line, ":", 2)
				if len(parts) == 2 {
					key := strings.TrimSpace(parts[0])
					value := strings.TrimSpace(parts[1])
					value = strings.Trim(value, "\"'")
					selector[key] = value
				}
			} else if line == "containers:" {
				inContainers = true
				inPorts = false
				inResources = false
			} else if inContainers {
				// Look for image field anywhere in containers section
				if strings.HasPrefix(line, "image:") {
					newImage := strings.TrimSpace(strings.TrimPrefix(line, "image:"))
					newImage = strings.Trim(newImage, "\"'")
					// Take the first valid image found
					if newImage != "" && image == "Unknown" {
						image = newImage
					}
				} else if line == "ports:" {
					inPorts = true
					inResources = false
				} else if inPorts && strings.HasPrefix(line, "- containerPort:") {
					port := strings.TrimSpace(strings.TrimPrefix(line, "- containerPort:"))
					ports = append(ports, port)
				} else if line == "resources:" {
					inResources = true
					inPorts = false
				} else if inResources {
					if strings.HasPrefix(line, "cpu:") {
						resources["cpu"] = strings.TrimSpace(strings.TrimPrefix(line, "cpu:"))
					} else if strings.HasPrefix(line, "memory:") {
						resources["memory"] = strings.TrimSpace(strings.TrimPrefix(line, "memory:"))
					}
				}
			}
		}
	}

	// Build selector string
	selectorStr := "Unknown"
	if len(selector) > 0 {
		var selectorParts []string
		for k, v := range selector {
			selectorParts = append(selectorParts, fmt.Sprintf("%s=%s", k, v))
		}
		selectorStr = strings.Join(selectorParts, ", ")
	}

	// Build ports string
	portsStr := "None"
	if len(ports) > 0 {
		portsStr = strings.Join(ports, ", ")
	}

	// Build resources string
	resourcesStr := "Not specified"
	if len(resources) > 0 {
		var resParts []string
		for k, v := range resources {
			resParts = append(resParts, fmt.Sprintf("%s: %s", k, v))
		}
		resourcesStr = strings.Join(resParts, ", ")
	}

	// Generate ASCII flowchart with actual parsed data
	diagram := fmt.Sprintf(`
╔══════════════════════════════════════════════════════════════════════════════╗
║                              DEPLOYMENT DIAGRAM                             ║
╚══════════════════════════════════════════════════════════════════════════════╝

┌─────────────────────────────────────────────────────────────────────────────┐
│                           Deployment: %-35s                │
├─────────────────────────────────────────────────────────────────────────────┤
│  Namespace: %-62s │
│  Replicas:  %-62s │
│  Selector:  %-62s │
└─────────────────────────────────────────────────────────────────────────────┘
                                       │
                                    creates
                                       ▼
┌─────────────────────────────────────────────────────────────────────────────┐
│                              ReplicaSet                                    │
├─────────────────────────────────────────────────────────────────────────────┤
│  Ensures %s Pod(s) are running                                            │
│  Matches Pods with selector: %-48s │
└─────────────────────────────────────────────────────────────────────────────┘
                                       │
                                   manages
                                       ▼
┌─────────────────────────────────────────────────────────────────────────────┐
│                               Pod(s) x%s                                   │
├─────────────────────────────────────────────────────────────────────────────┤
│  Each Pod contains container(s) running:                                   │
│  Image: %-64s │
│  Ports: %-64s │
│  Resources: %-58s │
└─────────────────────────────────────────────────────────────────────────────┘

DEPLOYMENT SPECIFICATION:
═══════════════════════════════════════════════════════════════════════════════
• Name: %s
• Namespace: %s
• Desired Replicas: %s
• Container Image: %s
• Container Ports: %s
• Resource Limits: %s
• Pod Selector: %s

KUBERNETES OBJECTS CREATED:
═══════════════════════════════════════════════════════════════════════════════
1. Deployment: %s (namespace: %s)
2. ReplicaSet: %s-xxxxxxxxxx (managed by deployment)
3. Pods: %s-xxxxxxxxxx-yyyyy (managed by replicaset, count: %s)
`,
		truncateString(name, 35),
		truncateString(namespace, 62),
		truncateString(replicas, 62),
		truncateString(selectorStr, 62),
		replicas,
		truncateString(selectorStr, 48),
		replicas,
		truncateString(image, 64),
		truncateString(portsStr, 64),
		truncateString(resourcesStr, 58),
		name, namespace, replicas, image, portsStr, resourcesStr, selectorStr,
		name, namespace, name, name, replicas)

	return diagram
}

func (a *App) getScrollableEventsContent() string {
	if a.state.EventsOutput == "" {
		return "Loading events..."
	}

	lines := strings.Split(a.state.EventsOutput, "\n")
	viewHeight := a.getEventsViewHeight() - 2 // Account for padding

	// Calculate visible range
	startLine := a.state.EventsScrollOffset
	endLine := startLine + viewHeight

	if startLine >= len(lines) {
		startLine = len(lines) - 1
		if startLine < 0 {
			startLine = 0
		}
	}

	if endLine > len(lines) {
		endLine = len(lines)
	}

	if startLine >= endLine {
		return "-- End of content --"
	}

	visibleLines := lines[startLine:endLine]

	// Pad with empty lines if content is shorter than view
	for len(visibleLines) < viewHeight {
		visibleLines = append(visibleLines, "")
	}

	content := strings.Join(visibleLines, "\n")

	// Apply selection highlighting if active
	if a.state.SelectionActive && a.state.CurrentView == types.EventsView {
		content = a.renderWithSelectionHighlight(content, 3) // viewStartY = 3
	}

	return content
}

func (a *App) getEventsScrollInfo() string {
	if a.state.EventsOutput == "" {
		return "Loading..."
	}

	lines := strings.Split(a.state.EventsOutput, "\n")
	totalLines := len(lines)
	viewHeight := a.getEventsViewHeight() - 2
	currentLine := a.state.EventsScrollOffset + 1

	if totalLines <= viewHeight {
		return "All"
	}

	endLine := currentLine + viewHeight - 1
	if endLine > totalLines {
		endLine = totalLines
	}

	return fmt.Sprintf("Lines %d-%d of %d", currentLine, endLine, totalLines)
}

func (a *App) getEventsViewHeight() int {
	// Account for title, instructions, padding, and borders
	return a.height - 8
}
