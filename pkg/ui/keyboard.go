package ui

import (
	"fmt"
	"os/exec"
	"strings"

	"k8sgo/internal/types"

	"github.com/aymanbagabas/go-osc52/v2"
	tea "github.com/charmbracelet/bubbletea"
)

func (a *App) handleKeyPress(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	// Clear previous status messages on most key presses (except copy operations)
	if msg.String() != "x" {
		a.clearStatusMessage()
	}

	// Handle context view separately
	if a.state.CurrentView == types.ContextView {
		return a.handleContextViewKeys(msg)
	}

	// Handle describe view scrolling
	if a.state.CurrentView == types.DescribeView {
		return a.handleDescribeViewKeys(msg)
	}

	// Handle tags view scrolling
	if a.state.CurrentView == types.TagsView {
		return a.handleTagsViewKeys(msg)
	}

	// Handle top view scrolling
	if a.state.CurrentView == types.TopView {
		return a.handleTopViewKeys(msg)
	}

	// Handle resources view scrolling
	if a.state.CurrentView == types.ResourcesView {
		return a.handleResourcesViewKeys(msg)
	}

	// Handle YAML view scrolling
	if a.state.CurrentView == types.YamlView {
		return a.handleYamlViewKeys(msg)
	}

	// Handle Events view scrolling
	if a.state.CurrentView == types.EventsView {
		return a.handleEventsViewKeys(msg)
	}

	// Handle namespace view
	if a.state.CurrentView == types.NamespaceView {
		return a.handleNamespaceViewKeys(msg)
	}

	// Handle memory view
	if a.state.CurrentView == types.MemoryView {
		return a.handleMemoryViewKeys(msg)
	}

	// Handle logs view
	if a.state.CurrentView == types.LogsView {
		return a.handleLogsViewKeys(msg)
	}

	// Handle diagram view
	if a.state.CurrentView == types.DiagramView {
		return a.handleDiagramViewKeys(msg)
	}

	// Handle feedback view
	if a.state.CurrentView == types.FeedbackView {
		return a.handleFeedbackViewKeys(msg)
	}

	switch msg.String() {
	case "q":
		// Show feedback form before quitting
		a.state.CurrentView = types.FeedbackView
		a.state.FeedbackText = ""
		a.state.FeedbackCursorPos = 0
		a.state.ShowFeedbackForm = true
		a.state.FeedbackSubmitting = false
		return a, nil

	case "ctrl+c":
		// Force quit without feedback
		return a, tea.Quit

	case "?":
		if a.state.CurrentView == types.HelpView {
			a.state.CurrentView = types.ResourceView
		} else {
			a.state.CurrentView = types.HelpView
		}
		return a, nil

	case "esc":
		if a.state.CurrentView == types.HelpView || a.state.CurrentView == types.DescribeView {
			a.state.CurrentView = types.ResourceView
			return a, nil
		}
		// ESC anywhere else goes to context selection
		return a.handleContextSwitch()

	case "j", "down":
		return a.handleDownNavigation()

	case "k", "up":
		return a.handleUpNavigation()

	case "h", "left":
		return a.handleLeftNavigation()

	case "l", "right":
		return a.handleRightNavigation()

	case "/":
		return a.handleSearchFocus()

	case "r":
		return a, a.fetchResources

	case "enter":
		return a.handleResourceSelect()

	case "tab":
		// Tab goes to context selection from anywhere
		return a.handleContextSwitch()

	case "backspace":
		return a.handleBackNavigation()

	case "1", "2", "3", "4", "5", "6", "7", "8":
		return a.handleResourceTypeSwitchByIndex(msg.String())

	case "c":
		return a.handleContextSwitch()

	case "i":
		return a.handleDescribeResource()

	case "d":
		return a.handleShowDiagram()

	case "t":
		return a.handleShowTop()

	case "m":
		return a.handleShowMemory()

	case "y":
		return a.handleShowYaml()

	case "e":
		return a.handleShowEvents()

	case "n":
		return a.handleNamespaceSelector()

	case "L":
		return a.handleShowLogs()

	case "x":
		// Use 'x' for extract/copy
		return a.handleCopyContent()
	}

	return a, nil
}

// Mouse event handling
func (a *App) handleMouseEvent(msg tea.MouseMsg) (tea.Model, tea.Cmd) {
	switch msg.Type {
	case tea.MouseWheelUp:
		return a.handleMouseWheelUp()
	case tea.MouseWheelDown:
		return a.handleMouseWheelDown()
	case tea.MouseLeft:
		a.state.StatusMessage = fmt.Sprintf("Mouse click at (%d,%d)", msg.X, msg.Y)
		return a.handleMouseClick(msg)
	case tea.MouseRight:
		a.state.StatusMessage = fmt.Sprintf("Right click at (%d,%d)", msg.X, msg.Y)
		return a.handleRightClick(msg)
	case tea.MouseRelease:
		a.state.StatusMessage = fmt.Sprintf("Mouse release at (%d,%d)", msg.X, msg.Y)
		return a.handleMouseRelease(msg)
	case tea.MouseMotion:
		if a.state.SelectionActive {
			a.state.StatusMessage = fmt.Sprintf("Dragging to (%d,%d)", msg.X, msg.Y)
		}
		return a.handleMouseMotion(msg)
	}
	return a, nil
}

func (a *App) handleMouseWheelUp() (tea.Model, tea.Cmd) {
	switch a.state.CurrentView {
	case types.ResourceView:
		return a.handleUpNavigation()
	case types.TagsView:
		a.scrollTagsUp()
		return a, nil
	case types.TopView:
		a.scrollTopUp()
		return a, nil
	case types.ResourcesView:
		a.scrollResourcesUp()
		return a, nil
	case types.YamlView:
		a.scrollYamlUp()
		return a, nil
	case types.EventsView:
		a.scrollEventsUp()
		return a, nil
	case types.DiagramView:
		a.scrollDiagramUp()
		return a, nil
	}
	return a, nil
}

func (a *App) handleMouseWheelDown() (tea.Model, tea.Cmd) {
	switch a.state.CurrentView {
	case types.ResourceView:
		return a.handleDownNavigation()
	case types.TagsView:
		a.scrollTagsDown()
		return a, nil
	case types.TopView:
		a.scrollTopDown()
		return a, nil
	case types.ResourcesView:
		a.scrollResourcesDown()
		return a, nil
	case types.YamlView:
		a.scrollYamlDown()
		return a, nil
	case types.EventsView:
		a.scrollEventsDown()
		return a, nil
	case types.DiagramView:
		a.scrollDiagramDown()
		return a, nil
	}
	return a, nil
}

func (a *App) handleMouseClick(msg tea.MouseMsg) (tea.Model, tea.Cmd) {
	// Handle header clicks for sorting in resource view
	if a.state.CurrentView == types.ResourceView {
		// Check if click is on table header (approximate position)
		headerY := 7 // Approximate Y position of table header
		if msg.Y == headerY {
			// Determine which column was clicked based on X position
			columnIndex := a.getClickedColumn(msg.X)
			if columnIndex >= 0 {
				return a.handleColumnSort(columnIndex)
			}
		}

		// Handle clicks in resource table for row selection
		resourceRowStart := 8 // Approximate start of resource table
		if msg.Y >= resourceRowStart && msg.Y < resourceRowStart+len(a.state.Resources) {
			clickedIndex := msg.Y - resourceRowStart
			if clickedIndex >= 0 && clickedIndex < len(a.state.Resources) {
				a.state.SelectedIndex = clickedIndex
				return a, nil
			}
		}
	}

	// Only start text selection in views where it's useful (not in main resource view)
	if a.state.CurrentView == types.DescribeView ||
		a.state.CurrentView == types.YamlView ||
		a.state.CurrentView == types.TagsView ||
		a.state.CurrentView == types.TopView ||
		a.state.CurrentView == types.EventsView ||
		a.state.CurrentView == types.ResourcesView ||
		a.state.CurrentView == types.DiagramView {
		// Start text selection only on click, not on move
		a.state.SelectionActive = true
		a.state.SelectionStartX = msg.X
		a.state.SelectionStartY = msg.Y
		a.state.SelectionEndX = msg.X
		a.state.SelectionEndY = msg.Y
		a.state.SelectedText = ""
		a.state.StatusMessage = "Click and drag to select text, then right-click or press 'x' to copy"
	}

	return a, nil
}

func (a *App) handleRightClick(msg tea.MouseMsg) (tea.Model, tea.Cmd) {
	// Right click - copy selection or current content
	if a.state.SelectionActive && a.state.SelectedText != "" {
		// Copy selected text using OSC52
		copyCmd := a.createCopyCommand(a.state.SelectedText)
		a.state.StatusMessage = fmt.Sprintf("✓ Copied selected text (%d chars) - use Ctrl+V to paste", len(a.state.SelectedText))

		// Clear selection after copying
		a.state.SelectionActive = false
		a.state.SelectedText = ""

		return a, copyCmd
	} else {
		// Copy current content (same as x key)
		return a.handleCopyContent()
	}
}

func (a *App) handleMouseRelease(msg tea.MouseMsg) (tea.Model, tea.Cmd) {
	if a.state.SelectionActive {
		a.state.SelectionEndX = msg.X
		a.state.SelectionEndY = msg.Y

		// Extract selected text based on coordinates
		a.extractSelectedText()

		// Keep selection active if text was selected, but provide better feedback
		if a.state.SelectedText == "" {
			a.state.SelectionActive = false
			a.state.StatusMessage = "No text selected - click and drag to select text"
		} else {
			// Clean up the selected text (remove excessive whitespace)
			a.state.SelectedText = strings.TrimSpace(a.state.SelectedText)
			if a.state.SelectedText == "" {
				a.state.SelectionActive = false
				a.state.StatusMessage = "No text selected - click and drag to select text"
			} else {
				a.state.StatusMessage = fmt.Sprintf("Selected %d chars - right-click or press 'x' to copy", len(a.state.SelectedText))
			}
		}
	}
	return a, nil
}

func (a *App) handleMouseMotion(msg tea.MouseMsg) (tea.Model, tea.Cmd) {
	// Only update selection if we're actively dragging (selection was started by click)
	// and we're in a view that supports text selection
	if a.state.SelectionActive &&
		(a.state.CurrentView == types.DescribeView ||
			a.state.CurrentView == types.YamlView ||
			a.state.CurrentView == types.TagsView ||
			a.state.CurrentView == types.TopView ||
			a.state.CurrentView == types.EventsView ||
			a.state.CurrentView == types.ResourcesView ||
			a.state.CurrentView == types.DiagramView) {
		a.state.SelectionEndX = msg.X
		a.state.SelectionEndY = msg.Y

		// Update selected text in real-time
		a.extractSelectedText()
	}
	return a, nil
}

func (a *App) handleDownNavigation() (tea.Model, tea.Cmd) {
	if len(a.state.Resources) > 0 {
		maxIndex := len(a.state.Resources) - 1
		if a.state.SelectedIndex < maxIndex {
			a.state.SelectedIndex++
			// Check if we need to go to next page
			pageStart := a.state.CurrentPage * a.state.PageSize
			pageEnd := pageStart + a.state.PageSize - 1
			if a.state.SelectedIndex > pageEnd {
				a.state.CurrentPage++
			}
		}
	}
	return a, nil
}

func (a *App) handleUpNavigation() (tea.Model, tea.Cmd) {
	if len(a.state.Resources) > 0 {
		if a.state.SelectedIndex > 0 {
			a.state.SelectedIndex--
			// Check if we need to go to previous page
			pageStart := a.state.CurrentPage * a.state.PageSize
			if a.state.SelectedIndex < pageStart {
				a.state.CurrentPage--
			}
		}
	}
	return a, nil
}

func (a *App) handleLeftNavigation() (tea.Model, tea.Cmd) {
	if a.state.CurrentPage > 0 {
		a.state.CurrentPage--
	}
	return a, nil
}

func (a *App) handleRightNavigation() (tea.Model, tea.Cmd) {
	paginatedResult := a.paginateResources()
	if paginatedResult.HasNext {
		a.state.CurrentPage++
	}
	return a, nil
}

func (a *App) handleSearchFocus() (tea.Model, tea.Cmd) {
	a.state.FocusedElement = "search"
	return a, nil
}

func (a *App) handleResourceSelect() (tea.Model, tea.Cmd) {
	return a, nil
}

func (a *App) handleBackNavigation() (tea.Model, tea.Cmd) {
	// Backspace goes back to previous view/state
	// Note: ResourceView = 0, so we need to check if PreviousView has been set at all
	// We'll use a different approach - check if we're not already in ResourceView
	if a.state.CurrentView != types.ResourceView {
		// Go back to the previous view, or ResourceView if no previous view
		if a.state.PreviousView != a.state.CurrentView {
			previousView := a.state.PreviousView
			a.state.PreviousView = a.state.CurrentView
			a.state.CurrentView = previousView
		} else {
			// Default fallback to ResourceView
			a.state.PreviousView = a.state.CurrentView
			a.state.CurrentView = types.ResourceView
		}

		// If going back to resource view and we have a previous resource type, restore it
		if a.state.CurrentView == types.ResourceView && a.state.PreviousResource != "" {
			previousResource := a.state.PreviousResource
			a.state.PreviousResource = a.state.ResourceType
			a.state.ResourceType = previousResource
			a.state.CurrentPage = 0
			a.state.SelectedIndex = 0
			return a, a.fetchResources
		}
	}
	return a, nil
}

func (a *App) handleNamespaceSwitch() (tea.Model, tea.Cmd) {
	// Get namespaces using oc command
	namespaces, err := a.ocCommands.GetNamespaces()
	if err != nil {
		return a, nil
	}

	// Find current namespace index
	currentIndex := 0
	for i, ns := range namespaces {
		if ns == a.state.Namespace {
			currentIndex = i
			break
		}
	}

	// Switch to next namespace
	nextIndex := (currentIndex + 1) % len(namespaces)
	newNamespace := namespaces[nextIndex]

	// Switch namespace using oc project command
	err = a.switchToNamespace(newNamespace)
	if err != nil {
		return a, nil
	}

	a.state.Namespace = newNamespace
	a.state.CurrentPage = 0 // Reset pagination

	return a, a.fetchResources
}

func (a *App) handleResourceTypeSwitch(resourceType string) (tea.Model, tea.Cmd) {
	// Track previous resource type for backspace navigation
	a.state.PreviousResource = a.state.ResourceType
	a.state.PreviousView = a.state.CurrentView

	a.state.ResourceType = resourceType
	a.state.CurrentPage = 0   // Reset pagination
	a.state.SelectedIndex = 0 // Reset selection
	return a, a.fetchResources
}

func (a *App) handleContextSwitch() (tea.Model, tea.Cmd) {
	if err := a.contextSelector.LoadContexts(); err != nil {
		return a, nil
	}
	// Track previous view for backspace navigation
	a.state.PreviousView = a.state.CurrentView
	a.state.CurrentView = types.ContextView
	return a, nil
}

func (a *App) handleContextViewKeys(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "q":
		// Show feedback form before quitting
		a.state.CurrentView = types.FeedbackView
		a.state.FeedbackText = ""
		a.state.FeedbackCursorPos = 0
		a.state.ShowFeedbackForm = true
		a.state.FeedbackSubmitting = false
		return a, nil
	case "ctrl+c":
		return a, tea.Quit
	case "esc":
		a.state.CurrentView = types.ResourceView
		return a, nil
	case "j", "down":
		a.contextSelector.MoveDown()
		return a, nil
	case "k", "up":
		a.contextSelector.MoveUp()
		return a, nil
	case "enter":
		if err := a.contextSelector.SelectContext(); err != nil {
			return a, nil
		}
		// Update current context and switch to namespace selection
		a.state.Context = a.contextSelector.GetSelectedContext()

		// Load namespaces for the selected context
		namespaces, err := a.ocCommands.GetNamespaces()
		if err != nil {
			// If namespace loading fails, go directly to resource view with default namespace
			a.state.Namespace = "default"
			a.state.CurrentView = types.ResourceView
			return a, a.fetchResources
		}

		// Get current namespace for this context
		currentNs, err := a.ocCommands.GetCurrentNamespace()
		if err != nil {
			currentNs = "default"
		}

		// Set up namespace selector
		a.namespaceSelector.SetNamespaces(namespaces, currentNs)
		a.state.Namespace = currentNs
		a.state.CurrentView = types.NamespaceView
		return a, nil
	}
	return a, nil
}

func (a *App) handleDescribeViewKeys(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "q":
		// Show feedback form before quitting
		a.state.CurrentView = types.FeedbackView
		a.state.FeedbackText = ""
		a.state.FeedbackCursorPos = 0
		a.state.ShowFeedbackForm = true
		a.state.FeedbackSubmitting = false
		return a, nil
	case "ctrl+c":
		return a, tea.Quit
	case "esc", "backspace":
		// Reset scroll position when leaving describe view
		a.state.DescribeScrollOffset = 0
		a.state.CurrentView = types.ResourceView
		return a, nil
	case "j", "down":
		// Scroll down
		a.scrollDescribeDown()
		return a, nil
	case "k", "up":
		// Scroll up
		a.scrollDescribeUp()
		return a, nil
	case "d", "ctrl+d":
		// Page down (half screen)
		a.pageDescribeDown()
		return a, nil
	case "u", "ctrl+u":
		// Page up (half screen)
		a.pageDescribeUp()
		return a, nil
	case "g":
		// Go to top
		a.state.DescribeScrollOffset = 0
		return a, nil
	case "G":
		// Go to bottom
		a.scrollDescribeToBottom()
		return a, nil
	case "x":
		// Copy describe content
		return a.handleCopyContent()
	case "L":
		// Show logs (if applicable)
		return a.handleShowLogs()
	}
	return a, nil
}

func (a *App) handleDescribeResource() (tea.Model, tea.Cmd) {
	// Check if we have resources and a valid selection
	if len(a.state.Resources) == 0 {
		return a, nil
	}

	// Ensure selected index is within bounds
	if a.state.SelectedIndex >= len(a.state.Resources) || a.state.SelectedIndex < 0 {
		a.state.SelectedIndex = 0
	}

	// Get the currently selected resource
	selectedResource := a.state.Resources[a.state.SelectedIndex]

	// Skip describe for "No resources found" or error entries
	if selectedResource.Type == "Info" || selectedResource.Type == "Error" {
		return a, nil
	}

	a.state.SelectedResource = selectedResource
	// Track previous view for backspace navigation
	a.state.PreviousView = a.state.CurrentView
	a.state.CurrentView = types.DescribeView
	// Reset scroll position for new resource
	a.state.DescribeScrollOffset = 0

	return a, a.fetchResourceDescription
}

func (a *App) switchToNamespace(namespace string) error {
	if a.state.Tool == "oc" {
		cmd := exec.Command("oc", "project", namespace)
		return cmd.Run()
	} else {
		// kubectl uses different approach - set namespace in context
		cmd := exec.Command("kubectl", "config", "set-context", "--current", "--namespace="+namespace)
		return cmd.Run()
	}
}

// Scrolling functions for describe view
func (a *App) scrollDescribeDown() {
	if a.state.DescribeScrollOffset < a.getMaxDescribeScroll() {
		a.state.DescribeScrollOffset++
	}
}

func (a *App) scrollDescribeUp() {
	if a.state.DescribeScrollOffset > 0 {
		a.state.DescribeScrollOffset--
	}
}

func (a *App) pageDescribeDown() {
	pageSize := a.getDescribeViewHeight() / 2
	if pageSize < 1 {
		pageSize = 1
	}
	maxScroll := a.getMaxDescribeScroll()
	a.state.DescribeScrollOffset += pageSize
	if a.state.DescribeScrollOffset > maxScroll {
		a.state.DescribeScrollOffset = maxScroll
	}
}

func (a *App) pageDescribeUp() {
	pageSize := a.getDescribeViewHeight() / 2
	if pageSize < 1 {
		pageSize = 1
	}
	a.state.DescribeScrollOffset -= pageSize
	if a.state.DescribeScrollOffset < 0 {
		a.state.DescribeScrollOffset = 0
	}
}

func (a *App) scrollDescribeToBottom() {
	a.state.DescribeScrollOffset = a.getMaxDescribeScroll()
}

func (a *App) getMaxDescribeScroll() int {
	if a.state.DescribeOutput == "" {
		return 0
	}

	lines := strings.Split(a.state.DescribeOutput, "\n")
	contentLines := len(lines)
	viewHeight := a.getDescribeViewHeight()

	maxScroll := contentLines - viewHeight + 2 // +2 for padding
	if maxScroll < 0 {
		maxScroll = 0
	}
	return maxScroll
}

// Tags view functions
func (a *App) handleShowTags() (tea.Model, tea.Cmd) {
	// Check if we have resources and a valid selection
	if len(a.state.Resources) == 0 || a.state.SelectedIndex >= len(a.state.Resources) {
		return a, nil
	}

	// Only allow tags for ImageStreams, and only if using oc
	if a.state.ResourceType != "is" || a.state.Tool != "oc" {
		return a, nil
	}

	// Get the currently selected resource
	selectedResource := a.state.Resources[a.state.SelectedIndex]
	a.state.SelectedResource = selectedResource
	// Track previous view for backspace navigation
	a.state.PreviousView = a.state.CurrentView
	a.state.CurrentView = types.TagsView
	// Reset scroll position for new resource
	a.state.TagsScrollOffset = 0

	return a, a.fetchResourceTags
}

// Top view functions
func (a *App) handleShowTop() (tea.Model, tea.Cmd) {
	// Track previous view for backspace navigation
	a.state.PreviousView = a.state.CurrentView
	a.state.CurrentView = types.TopView
	// Reset scroll position for new content
	a.state.TopScrollOffset = 0

	return a, a.fetchTopCommand
}

func (a *App) handleTagsViewKeys(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "q":
		// Show feedback form before quitting
		a.state.CurrentView = types.FeedbackView
		a.state.FeedbackText = ""
		a.state.FeedbackCursorPos = 0
		a.state.ShowFeedbackForm = true
		a.state.FeedbackSubmitting = false
		return a, nil
	case "ctrl+c":
		return a, tea.Quit
	case "esc", "backspace":
		// Reset scroll position when leaving tags view
		a.state.TagsScrollOffset = 0
		a.state.CurrentView = types.ResourceView
		return a, nil
	case "j", "down":
		// Scroll down
		a.scrollTagsDown()
		return a, nil
	case "k", "up":
		// Scroll up
		a.scrollTagsUp()
		return a, nil
	case "d", "ctrl+d":
		// Page down (half screen)
		a.pageTagsDown()
		return a, nil
	case "u", "ctrl+u":
		// Page up (half screen)
		a.pageTagsUp()
		return a, nil
	case "g":
		// Go to top
		a.state.TagsScrollOffset = 0
		return a, nil
	case "G":
		// Go to bottom
		a.scrollTagsToBottom()
		return a, nil
	case "x":
		// Copy tags content
		return a.handleCopyContent()
	case "L":
		// Show logs (if applicable)
		return a.handleShowLogs()
	}
	return a, nil
}

// Tags scrolling functions
func (a *App) scrollTagsDown() {
	if a.state.TagsScrollOffset < a.getMaxTagsScroll() {
		a.state.TagsScrollOffset++
	}
}

func (a *App) scrollTagsUp() {
	if a.state.TagsScrollOffset > 0 {
		a.state.TagsScrollOffset--
	}
}

func (a *App) pageTagsDown() {
	pageSize := a.getTagsViewHeight() / 2
	if pageSize < 1 {
		pageSize = 1
	}
	maxScroll := a.getMaxTagsScroll()
	a.state.TagsScrollOffset += pageSize
	if a.state.TagsScrollOffset > maxScroll {
		a.state.TagsScrollOffset = maxScroll
	}
}

func (a *App) pageTagsUp() {
	pageSize := a.getTagsViewHeight() / 2
	if pageSize < 1 {
		pageSize = 1
	}
	a.state.TagsScrollOffset -= pageSize
	if a.state.TagsScrollOffset < 0 {
		a.state.TagsScrollOffset = 0
	}
}

func (a *App) scrollTagsToBottom() {
	a.state.TagsScrollOffset = a.getMaxTagsScroll()
}

func (a *App) getMaxTagsScroll() int {
	if a.state.TagsOutput == "" {
		return 0
	}

	lines := strings.Split(a.state.TagsOutput, "\n")
	contentLines := len(lines)
	viewHeight := a.getTagsViewHeight()

	maxScroll := contentLines - viewHeight + 2 // +2 for padding
	if maxScroll < 0 {
		maxScroll = 0
	}
	return maxScroll
}

func (a *App) getTagsViewHeight() int {
	// Account for tags title, instructions, padding, and borders
	return a.height - 8
}

func (a *App) handleTopViewKeys(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "q":
		// Show feedback form before quitting
		a.state.CurrentView = types.FeedbackView
		a.state.FeedbackText = ""
		a.state.FeedbackCursorPos = 0
		a.state.ShowFeedbackForm = true
		a.state.FeedbackSubmitting = false
		return a, nil
	case "ctrl+c":
		return a, tea.Quit
	case "esc", "backspace":
		// Reset scroll position when leaving top view
		a.state.TopScrollOffset = 0
		a.state.CurrentView = types.ResourceView
		return a, nil
	case "j", "down":
		// Scroll down
		a.scrollTopDown()
		return a, nil
	case "k", "up":
		// Scroll up
		a.scrollTopUp()
		return a, nil
	case "d", "ctrl+d":
		// Page down (half screen)
		a.pageTopDown()
		return a, nil
	case "u", "ctrl+u":
		// Page up (half screen)
		a.pageTopUp()
		return a, nil
	case "g":
		// Go to top
		a.state.TopScrollOffset = 0
		return a, nil
	case "G":
		// Go to bottom
		a.scrollTopToBottom()
		return a, nil
	case "x":
		// Copy top content
		return a.handleCopyContent()
	case "L":
		// Show logs (if applicable)
		return a.handleShowLogs()
	}
	return a, nil
}

// Top scrolling functions
func (a *App) scrollTopDown() {
	if a.state.TopScrollOffset < a.getMaxTopScroll() {
		a.state.TopScrollOffset++
	}
}

func (a *App) scrollTopUp() {
	if a.state.TopScrollOffset > 0 {
		a.state.TopScrollOffset--
	}
}

func (a *App) pageTopDown() {
	pageSize := a.getTopViewHeight() / 2
	if pageSize < 1 {
		pageSize = 1
	}
	maxScroll := a.getMaxTopScroll()
	a.state.TopScrollOffset += pageSize
	if a.state.TopScrollOffset > maxScroll {
		a.state.TopScrollOffset = maxScroll
	}
}

func (a *App) pageTopUp() {
	pageSize := a.getTopViewHeight() / 2
	if pageSize < 1 {
		pageSize = 1
	}
	a.state.TopScrollOffset -= pageSize
	if a.state.TopScrollOffset < 0 {
		a.state.TopScrollOffset = 0
	}
}

func (a *App) scrollTopToBottom() {
	a.state.TopScrollOffset = a.getMaxTopScroll()
}

func (a *App) getMaxTopScroll() int {
	if a.state.TopOutput == "" {
		return 0
	}

	lines := strings.Split(a.state.TopOutput, "\n")
	contentLines := len(lines)
	viewHeight := a.getTopViewHeight()

	maxScroll := contentLines - viewHeight + 2 // +2 for padding
	if maxScroll < 0 {
		maxScroll = 0
	}
	return maxScroll
}

func (a *App) getTopViewHeight() int {
	// Account for top title, instructions, padding, and borders
	return a.height - 8
}

// Copy functionality
func (a *App) handleCopyContent() (tea.Model, tea.Cmd) {
	var content string

	// First check if there's selected text from mouse selection
	if a.state.SelectionActive && a.state.SelectedText != "" {
		content = a.state.SelectedText
		a.state.StatusMessage = fmt.Sprintf("✓ Copied selected text (%d chars) to clipboard", len(content))
	} else {
		// Fall back to copying entire view content
		switch a.state.CurrentView {
		case types.ResourceView:
			// Copy selected resource name or entire resource info
			if len(a.state.Resources) > 0 && a.state.SelectedIndex < len(a.state.Resources) {
				selectedResource := a.state.Resources[a.state.SelectedIndex]
				age := formatDuration(selectedResource.Age)
				content = fmt.Sprintf("Name: %s\nNamespace: %s\nStatus: %s\nReady: %s\nRestarts: %d\nAge: %s",
					selectedResource.Name, selectedResource.Namespace, selectedResource.Status,
					selectedResource.Ready, selectedResource.Restarts, age)
			}
		case types.DescribeView:
			// Copy entire describe output
			content = a.state.DescribeOutput
		case types.TagsView:
			// Copy entire tags output
			content = a.state.TagsOutput
		case types.TopView:
			// Copy entire top output
			content = a.state.TopOutput
		case types.ResourcesView:
			// Copy entire resources output
			content = a.state.ResourcesOutput
		case types.YamlView:
			// Copy entire YAML output
			content = a.state.YamlOutput
		case types.EventsView:
			// Copy entire events output
			content = a.state.EventsOutput
		case types.DiagramView:
			// Copy entire diagram output
			content = a.state.DiagramOutput
		case types.MemoryView:
			// Copy entire memory output
			content = a.state.MemoryOutput
		case types.LogsView:
			// Copy entire logs output
			content = a.state.LogsOutput
		}

		if content != "" {
			a.state.StatusMessage = fmt.Sprintf("✓ Copied %d chars to clipboard (use Ctrl+V to paste)", len(content))
		}
	}

	if content != "" {
		// Clear selection after copying
		a.state.SelectionActive = false
		a.state.SelectedText = ""

		// Return a command that will copy the content using OSC52
		copyCmd := a.createCopyCommand(content)
		return a, copyCmd
	} else {
		a.state.StatusMessage = "✗ No content to copy"
		return a, nil
	}
}

// createCopyCommand creates a Bubbletea command to copy content using OSC52
func (a *App) createCopyCommand(text string) tea.Cmd {
	// Use OSC52 terminal sequences - works without system dependencies
	clipboardSeq := osc52.New(text)
	primarySeq := osc52.New(text).Primary()

	// Send both clipboard and primary selection sequences
	return tea.Printf("%s%s", clipboardSeq.String(), primarySeq.String())
}

// extractSelectedText extracts text content based on mouse selection coordinates
func (a *App) extractSelectedText() {
	if !a.state.SelectionActive {
		a.state.SelectedText = ""
		return
	}

	// Normalize selection coordinates (ensure start is top-left, end is bottom-right)
	startX, startY := a.state.SelectionStartX, a.state.SelectionStartY
	endX, endY := a.state.SelectionEndX, a.state.SelectionEndY

	if startY > endY || (startY == endY && startX > endX) {
		// Swap coordinates if selection was made backwards
		startX, endX = endX, startX
		startY, endY = endY, startY
	}

	// Validate coordinates
	if startX < 0 || startY < 0 || endX < 0 || endY < 0 {
		a.state.SelectedText = ""
		return
	}

	// If start and end are the same, no selection
	if startX == endX && startY == endY {
		a.state.SelectedText = ""
		return
	}

	var content string
	switch a.state.CurrentView {
	case types.ResourceView:
		content = a.extractTextFromResourceView(startX, startY, endX, endY)
	case types.TagsView:
		content = a.extractTextFromTagsView(startX, startY, endX, endY)
	case types.TopView:
		content = a.extractTextFromScrollableView(a.state.TopOutput, a.state.TopScrollOffset, startX, startY, endX, endY)
	case types.DescribeView:
		content = a.extractTextFromDescribeView(startX, startY, endX, endY)
	case types.YamlView:
		content = a.extractTextFromScrollableView(a.state.YamlOutput, a.state.YamlScrollOffset, startX, startY, endX, endY)
	case types.EventsView:
		content = a.extractTextFromScrollableView(a.state.EventsOutput, a.state.EventsScrollOffset, startX, startY, endX, endY)
	case types.ResourcesView:
		content = a.extractTextFromScrollableView(a.state.ResourcesOutput, a.state.ResourcesScrollOffset, startX, startY, endX, endY)
	case types.DiagramView:
		content = a.extractTextFromScrollableView(a.state.DiagramOutput, a.state.DiagramScrollOffset, startX, startY, endX, endY)
	case types.MemoryView:
		content = a.extractTextFromScrollableView(a.state.MemoryOutput, a.state.MemoryScrollOffset, startX, startY, endX, endY)
	case types.LogsView:
		content = a.extractTextFromScrollableView(a.state.LogsOutput, a.state.LogsScrollOffset, startX, startY, endX, endY)
	default:
		content = ""
	}

	// Clean up the content - remove excessive whitespace but preserve structure
	lines := strings.Split(content, "\n")
	var cleanLines []string
	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if trimmed != "" {
			cleanLines = append(cleanLines, trimmed)
		}
	}

	a.state.SelectedText = strings.Join(cleanLines, "\n")
}

// extractTextFromResourceView extracts text from the resource table view
func (a *App) extractTextFromResourceView(startX, startY, endX, endY int) string {
	if len(a.state.Resources) == 0 {
		return ""
	}

	// Get paginated resources
	paginatedResult := a.paginateResources()
	resources := paginatedResult.Items

	// Calculate approximate positions of table elements
	// This is simplified - in a real implementation you'd need precise layout calculations
	headerOffset := 6               // Approximate lines for banner, context, search, etc.
	tableStartY := headerOffset + 3 // Header + borders

	var selectedLines []string

	// Extract from table rows
	for i, resource := range resources {
		rowY := tableStartY + i
		if rowY >= startY && rowY <= endY {
			// Extract relevant parts of the row based on X coordinates
			age := formatDuration(resource.Age)
			rowText := fmt.Sprintf("%s %s %s %d %s %s %s",
				resource.Name, resource.Ready, resource.Status,
				resource.Restarts, resource.CPU, resource.Memory, age)

			// Simple character-based extraction
			if startY == endY && startY == rowY {
				// Single line selection
				if startX < len(rowText) {
					end := endX
					if end > len(rowText) {
						end = len(rowText)
					}
					if startX < end {
						selectedLines = append(selectedLines, rowText[startX:end])
					}
				}
			} else {
				// Multi-line selection
				if rowY == startY {
					// First line
					if startX < len(rowText) {
						selectedLines = append(selectedLines, rowText[startX:])
					}
				} else if rowY == endY {
					// Last line
					end := endX
					if end > len(rowText) {
						end = len(rowText)
					}
					if end > 0 {
						selectedLines = append(selectedLines, rowText[:end])
					}
				} else {
					// Middle lines
					selectedLines = append(selectedLines, rowText)
				}
			}
		}
	}

	return strings.Join(selectedLines, "\n")
}

// extractTextFromTagsView extracts text from the tags output
func (a *App) extractTextFromTagsView(startX, startY, endX, endY int) string {
	if a.state.TagsOutput == "" {
		return ""
	}

	lines := strings.Split(a.state.TagsOutput, "\n")

	// Account for scroll offset and view positioning
	viewStartY := 3 // Title and padding

	var selectedLines []string

	for lineIdx, line := range lines {
		// Calculate screen Y position accounting for scroll
		screenY := viewStartY + lineIdx - a.state.TagsScrollOffset

		if screenY >= startY && screenY <= endY {
			if startY == endY && startY == screenY {
				// Single line selection
				if startX < len(line) {
					end := endX
					if end > len(line) {
						end = len(line)
					}
					if startX < end {
						selectedLines = append(selectedLines, line[startX:end])
					}
				}
			} else {
				// Multi-line selection
				if screenY == startY {
					// First line
					if startX < len(line) {
						selectedLines = append(selectedLines, line[startX:])
					}
				} else if screenY == endY {
					// Last line
					end := endX
					if end > len(line) {
						end = len(line)
					}
					if end > 0 {
						selectedLines = append(selectedLines, line[:end])
					}
				} else {
					// Middle lines
					selectedLines = append(selectedLines, line)
				}
			}
		}
	}

	return strings.Join(selectedLines, "\n")
}

// extractTextFromScrollableView extracts text from any scrollable view (YAML, Events, Logs, etc.)
func (a *App) extractTextFromScrollableView(output string, scrollOffset int, startX, startY, endX, endY int) string {
	if output == "" {
		return ""
	}

	lines := strings.Split(output, "\n")

	// Account for scroll offset and view positioning
	viewStartY := 3 // Title and padding

	var selectedLines []string

	for lineIdx, line := range lines {
		// Calculate screen Y position accounting for scroll
		screenY := viewStartY + lineIdx - scrollOffset

		if screenY >= startY && screenY <= endY {
			if startY == endY && startY == screenY {
				// Single line selection
				if startX < len(line) {
					end := endX
					if end > len(line) {
						end = len(line)
					}
					if startX < end {
						selectedLines = append(selectedLines, line[startX:end])
					}
				}
			} else {
				// Multi-line selection
				if screenY == startY {
					// First line
					if startX < len(line) {
						selectedLines = append(selectedLines, line[startX:])
					}
				} else if screenY == endY {
					// Last line
					end := endX
					if end > len(line) {
						end = len(line)
					}
					if end > 0 {
						selectedLines = append(selectedLines, line[:end])
					}
				} else {
					// Middle lines
					selectedLines = append(selectedLines, line)
				}
			}
		}
	}

	return strings.Join(selectedLines, "\n")
}

// Container Resources view functions
func (a *App) handleShowResources() (tea.Model, tea.Cmd) {
	// Check if we have resources and a valid selection
	if len(a.state.Resources) == 0 || a.state.SelectedIndex >= len(a.state.Resources) {
		return a, nil
	}

	// Only allow resources for Pods and Deployments
	if a.state.ResourceType != "pods" && a.state.ResourceType != "deployments" {
		return a, nil
	}

	// Get the currently selected resource
	selectedResource := a.state.Resources[a.state.SelectedIndex]
	a.state.SelectedResource = selectedResource
	// Track previous view for backspace navigation
	a.state.PreviousView = a.state.CurrentView
	a.state.CurrentView = types.ResourcesView
	// Reset scroll position for new resource
	a.state.ResourcesScrollOffset = 0

	return a, a.fetchResourceRequests
}

func (a *App) handleResourcesViewKeys(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "q":
		// Show feedback form before quitting
		a.state.CurrentView = types.FeedbackView
		a.state.FeedbackText = ""
		a.state.FeedbackCursorPos = 0
		a.state.ShowFeedbackForm = true
		a.state.FeedbackSubmitting = false
		return a, nil
	case "ctrl+c":
		return a, tea.Quit
	case "esc", "backspace":
		// Reset scroll position when leaving resources view
		a.state.ResourcesScrollOffset = 0
		a.state.CurrentView = types.ResourceView
		return a, nil
	case "j", "down":
		// Scroll down
		a.scrollResourcesDown()
		return a, nil
	case "k", "up":
		// Scroll up
		a.scrollResourcesUp()
		return a, nil
	case "d", "ctrl+d":
		// Page down (half screen)
		a.pageResourcesDown()
		return a, nil
	case "u", "ctrl+u":
		// Page up (half screen)
		a.pageResourcesUp()
		return a, nil
	case "g":
		// Go to top
		a.state.ResourcesScrollOffset = 0
		return a, nil
	case "G":
		// Go to bottom
		a.scrollResourcesToBottom()
		return a, nil
	case "x":
		// Copy resources content
		return a.handleCopyContent()
	case "L":
		// Show logs (if applicable)
		return a.handleShowLogs()
	}
	return a, nil
}

// Resources scrolling functions
func (a *App) scrollResourcesDown() {
	if a.state.ResourcesScrollOffset < a.getMaxResourcesScroll() {
		a.state.ResourcesScrollOffset++
	}
}

func (a *App) scrollResourcesUp() {
	if a.state.ResourcesScrollOffset > 0 {
		a.state.ResourcesScrollOffset--
	}
}

func (a *App) pageResourcesDown() {
	pageSize := a.getResourcesViewHeight() / 2
	if pageSize < 1 {
		pageSize = 1
	}
	maxScroll := a.getMaxResourcesScroll()
	a.state.ResourcesScrollOffset += pageSize
	if a.state.ResourcesScrollOffset > maxScroll {
		a.state.ResourcesScrollOffset = maxScroll
	}
}

func (a *App) pageResourcesUp() {
	pageSize := a.getResourcesViewHeight() / 2
	if pageSize < 1 {
		pageSize = 1
	}
	a.state.ResourcesScrollOffset -= pageSize
	if a.state.ResourcesScrollOffset < 0 {
		a.state.ResourcesScrollOffset = 0
	}
}

func (a *App) scrollResourcesToBottom() {
	a.state.ResourcesScrollOffset = a.getMaxResourcesScroll()
}

func (a *App) getMaxResourcesScroll() int {
	if a.state.ResourcesOutput == "" {
		return 0
	}

	lines := strings.Split(a.state.ResourcesOutput, "\n")
	contentLines := len(lines)
	viewHeight := a.getResourcesViewHeight()

	maxScroll := contentLines - viewHeight + 2 // +2 for padding
	if maxScroll < 0 {
		maxScroll = 0
	}
	return maxScroll
}

func (a *App) getResourcesViewHeight() int {
	// Account for resources title, instructions, padding, and borders
	return a.height - 8
}

// extractTextFromDescribeView extracts text from the describe output
func (a *App) extractTextFromDescribeView(startX, startY, endX, endY int) string {
	if a.state.DescribeOutput == "" {
		return ""
	}

	lines := strings.Split(a.state.DescribeOutput, "\n")

	// Account for scroll offset and view positioning
	viewStartY := 3 // Title and padding

	var selectedLines []string

	for lineIdx, line := range lines {
		// Calculate screen Y position accounting for scroll
		screenY := viewStartY + lineIdx - a.state.DescribeScrollOffset

		if screenY >= startY && screenY <= endY {
			if startY == endY && startY == screenY {
				// Single line selection
				if startX < len(line) {
					end := endX
					if end > len(line) {
						end = len(line)
					}
					if startX < end {
						selectedLines = append(selectedLines, line[startX:end])
					}
				}
			} else {
				// Multi-line selection
				if screenY == startY {
					// First line
					if startX < len(line) {
						selectedLines = append(selectedLines, line[startX:])
					}
				} else if screenY == endY {
					// Last line
					end := endX
					if end > len(line) {
						end = len(line)
					}
					if end > 0 {
						selectedLines = append(selectedLines, line[:end])
					}
				} else {
					// Middle lines
					selectedLines = append(selectedLines, line)
				}
			}
		}
	}

	return strings.Join(selectedLines, "\n")
}

// YAML view functions
func (a *App) handleShowYaml() (tea.Model, tea.Cmd) {
	// Check if we have resources and a valid selection
	if len(a.state.Resources) == 0 || a.state.SelectedIndex >= len(a.state.Resources) {
		return a, nil
	}

	// Get the currently selected resource
	selectedResource := a.state.Resources[a.state.SelectedIndex]

	// Skip YAML for "No resources found" or error entries
	if selectedResource.Type == "Info" || selectedResource.Type == "Error" {
		return a, nil
	}

	a.state.SelectedResource = selectedResource
	// Track previous view for backspace navigation
	a.state.PreviousView = a.state.CurrentView
	a.state.CurrentView = types.YamlView
	// Reset scroll position for new resource
	a.state.YamlScrollOffset = 0

	return a, a.fetchResourceYaml
}

func (a *App) handleYamlViewKeys(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "q":
		// Show feedback form before quitting
		a.state.CurrentView = types.FeedbackView
		a.state.FeedbackText = ""
		a.state.FeedbackCursorPos = 0
		a.state.ShowFeedbackForm = true
		a.state.FeedbackSubmitting = false
		return a, nil
	case "ctrl+c":
		return a, tea.Quit
	case "esc", "backspace":
		// Reset scroll position when leaving YAML view
		a.state.YamlScrollOffset = 0
		a.state.CurrentView = types.ResourceView
		return a, nil
	case "j", "down":
		// Scroll down
		a.scrollYamlDown()
		return a, nil
	case "k", "up":
		// Scroll up
		a.scrollYamlUp()
		return a, nil
	case "d", "ctrl+d":
		// Page down (half screen)
		a.pageYamlDown()
		return a, nil
	case "u", "ctrl+u":
		// Page up (half screen)
		a.pageYamlUp()
		return a, nil
	case "g":
		// Go to top
		a.state.YamlScrollOffset = 0
		return a, nil
	case "G":
		// Go to bottom
		a.scrollYamlToBottom()
		return a, nil
	case "x":
		// Copy YAML content
		return a.handleCopyContent()
	case "L":
		// Show logs (if applicable)
		return a.handleShowLogs()
	}
	return a, nil
}

// YAML scrolling functions
func (a *App) scrollYamlDown() {
	if a.state.YamlScrollOffset < a.getMaxYamlScroll() {
		a.state.YamlScrollOffset++
	}
}

func (a *App) scrollYamlUp() {
	if a.state.YamlScrollOffset > 0 {
		a.state.YamlScrollOffset--
	}
}

func (a *App) pageYamlDown() {
	pageSize := a.getYamlViewHeight() / 2
	if pageSize < 1 {
		pageSize = 1
	}
	maxScroll := a.getMaxYamlScroll()
	a.state.YamlScrollOffset += pageSize
	if a.state.YamlScrollOffset > maxScroll {
		a.state.YamlScrollOffset = maxScroll
	}
}

func (a *App) pageYamlUp() {
	pageSize := a.getYamlViewHeight() / 2
	if pageSize < 1 {
		pageSize = 1
	}
	a.state.YamlScrollOffset -= pageSize
	if a.state.YamlScrollOffset < 0 {
		a.state.YamlScrollOffset = 0
	}
}

func (a *App) scrollYamlToBottom() {
	a.state.YamlScrollOffset = a.getMaxYamlScroll()
}

func (a *App) getMaxYamlScroll() int {
	if a.state.YamlOutput == "" {
		return 0
	}

	lines := strings.Split(a.state.YamlOutput, "\n")
	contentLines := len(lines)
	viewHeight := a.getYamlViewHeight()

	maxScroll := contentLines - viewHeight + 2 // +2 for padding
	if maxScroll < 0 {
		maxScroll = 0
	}
	return maxScroll
}

// Events view functions
func (a *App) handleShowEvents() (tea.Model, tea.Cmd) {
	// Check if we have resources and a valid selection
	if len(a.state.Resources) == 0 || a.state.SelectedIndex >= len(a.state.Resources) {
		return a, nil
	}

	// Get the currently selected resource
	selectedResource := a.state.Resources[a.state.SelectedIndex]

	// Skip events for "No resources found" or error entries
	if selectedResource.Type == "Info" || selectedResource.Type == "Error" {
		return a, nil
	}

	a.state.SelectedResource = selectedResource
	// Track previous view for backspace navigation
	a.state.PreviousView = a.state.CurrentView
	a.state.CurrentView = types.EventsView
	// Reset scroll position for new resource
	a.state.EventsScrollOffset = 0

	return a, a.fetchResourceEvents
}

func (a *App) handleEventsViewKeys(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "q":
		// Show feedback form before quitting
		a.state.CurrentView = types.FeedbackView
		a.state.FeedbackText = ""
		a.state.FeedbackCursorPos = 0
		a.state.ShowFeedbackForm = true
		a.state.FeedbackSubmitting = false
		return a, nil
	case "ctrl+c":
		return a, tea.Quit
	case "esc", "backspace":
		// Reset scroll position when leaving events view
		a.state.EventsScrollOffset = 0
		a.state.CurrentView = types.ResourceView
		return a, nil
	case "j", "down":
		// Scroll down
		a.scrollEventsDown()
		return a, nil
	case "k", "up":
		// Scroll up
		a.scrollEventsUp()
		return a, nil
	case "d", "ctrl+d":
		// Page down (half screen)
		a.pageEventsDown()
		return a, nil
	case "u", "ctrl+u":
		// Page up (half screen)
		a.pageEventsUp()
		return a, nil
	case "g":
		// Go to top
		a.state.EventsScrollOffset = 0
		return a, nil
	case "G":
		// Go to bottom
		a.scrollEventsToBottom()
		return a, nil
	case "x":
		// Copy events content
		return a.handleCopyContent()
	case "L":
		// Show logs (if applicable)
		return a.handleShowLogs()
	}
	return a, nil
}

// Events scrolling functions
func (a *App) scrollEventsDown() {
	if a.state.EventsScrollOffset < a.getMaxEventsScroll() {
		a.state.EventsScrollOffset++
	}
}

func (a *App) scrollEventsUp() {
	if a.state.EventsScrollOffset > 0 {
		a.state.EventsScrollOffset--
	}
}

func (a *App) pageEventsDown() {
	pageSize := a.getEventsViewHeight() / 2
	if pageSize < 1 {
		pageSize = 1
	}
	maxScroll := a.getMaxEventsScroll()
	a.state.EventsScrollOffset += pageSize
	if a.state.EventsScrollOffset > maxScroll {
		a.state.EventsScrollOffset = maxScroll
	}
}

func (a *App) pageEventsUp() {
	pageSize := a.getEventsViewHeight() / 2
	if pageSize < 1 {
		pageSize = 1
	}
	a.state.EventsScrollOffset -= pageSize
	if a.state.EventsScrollOffset < 0 {
		a.state.EventsScrollOffset = 0
	}
}

func (a *App) scrollEventsToBottom() {
	a.state.EventsScrollOffset = a.getMaxEventsScroll()
}

func (a *App) getMaxEventsScroll() int {
	if a.state.EventsOutput == "" {
		return 0
	}

	lines := strings.Split(a.state.EventsOutput, "\n")
	contentLines := len(lines)
	viewHeight := a.getEventsViewHeight()

	maxScroll := contentLines - viewHeight + 2 // +2 for padding
	if maxScroll < 0 {
		maxScroll = 0
	}
	return maxScroll
}

// Namespace selector functions
func (a *App) handleNamespaceSelector() (tea.Model, tea.Cmd) {
	// Get namespaces using oc command
	namespaces, err := a.ocCommands.GetNamespaces()
	if err != nil {
		return a, nil
	}

	// Initialize namespace selector
	if a.namespaceSelector == nil {
		a.namespaceSelector = NewNamespaceSelector()
	}

	a.namespaceSelector.SetNamespaces(namespaces, a.state.Namespace)

	// Track previous view for backspace navigation
	a.state.PreviousView = a.state.CurrentView
	a.state.CurrentView = types.NamespaceView
	return a, nil
}

func (a *App) handleNamespaceViewKeys(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "q":
		// Show feedback form before quitting
		a.state.CurrentView = types.FeedbackView
		a.state.FeedbackText = ""
		a.state.FeedbackCursorPos = 0
		a.state.ShowFeedbackForm = true
		a.state.FeedbackSubmitting = false
		return a, nil
	case "ctrl+c":
		return a, tea.Quit
	case "esc", "backspace":
		// Go back to context selection
		a.state.CurrentView = types.ContextView
		return a, nil
	case "j", "down":
		a.namespaceSelector.MoveDown()
		return a, nil
	case "k", "up":
		a.namespaceSelector.MoveUp()
		return a, nil
	case "enter":
		selectedNamespace := a.namespaceSelector.GetSelectedNamespace()
		if selectedNamespace != "" {
			// Switch to the selected namespace
			err := a.switchToNamespace(selectedNamespace)
			if err != nil {
				return a, nil
			}
			a.state.Namespace = selectedNamespace
			a.state.CurrentPage = 0   // Reset pagination
			a.state.SelectedIndex = 0 // Reset selection
		}
		a.state.CurrentView = types.ResourceView
		return a, a.fetchResources
	}
	return a, nil
}

// Diagram view functions
func (a *App) handleShowDiagram() (tea.Model, tea.Cmd) {
	// Check if we have resources and a valid selection
	if len(a.state.Resources) == 0 || a.state.SelectedIndex >= len(a.state.Resources) {
		return a, nil
	}

	// Only allow diagrams for Deployments
	if a.state.ResourceType != "deployments" {
		return a, nil
	}

	// Get the currently selected resource
	selectedResource := a.state.Resources[a.state.SelectedIndex]

	// Skip diagram for "No resources found" or error entries
	if selectedResource.Type == "Info" || selectedResource.Type == "Error" {
		return a, nil
	}

	a.state.SelectedResource = selectedResource
	// Track previous view for backspace navigation
	a.state.PreviousView = a.state.CurrentView
	a.state.CurrentView = types.DiagramView
	// Reset scroll position for new resource
	a.state.DiagramScrollOffset = 0

	return a, a.fetchResourceDiagram
}

func (a *App) handleShowMemory() (tea.Model, tea.Cmd) {
	// Check if we have resources and a valid selection
	if len(a.state.Resources) == 0 || a.state.SelectedIndex >= len(a.state.Resources) {
		return a, nil
	}

	// Only allow memory for Pods
	if a.state.ResourceType != "pods" {
		return a, nil
	}

	// Get the currently selected resource
	selectedResource := a.state.Resources[a.state.SelectedIndex]

	// Skip memory for "No resources found" or error entries
	if selectedResource.Type == "Info" || selectedResource.Type == "Error" {
		return a, nil
	}

	a.state.SelectedResource = selectedResource
	// Track previous view for backspace navigation
	a.state.PreviousView = a.state.CurrentView
	a.state.CurrentView = types.MemoryView
	// Reset scroll position for new resource
	a.state.MemoryScrollOffset = 0

	return a, a.fetchPodMemory
}

func (a *App) handleShowLogs() (tea.Model, tea.Cmd) {
	// Check if we have resources - if not, try to fetch them first
	if len(a.state.Resources) == 0 {
		a.state.StatusMessage = "Loading resources..."
		return a, a.fetchResources
	}

	// Ensure we have a valid selection index
	if a.state.SelectedIndex >= len(a.state.Resources) {
		a.state.SelectedIndex = 0 // Reset to first resource
	}

	// Ensure we're viewing pods for logs functionality
	if a.state.ResourceType != "pods" {
		// Switch to pods resource type and fetch
		a.state.ResourceType = "pods"
		a.state.CurrentPage = 0
		a.state.SelectedIndex = 0
		a.state.StatusMessage = "Switching to pods view for logs..."
		return a, a.fetchResources
	}

	// Get the currently selected resource
	selectedResource := a.state.Resources[a.state.SelectedIndex]

	// Skip logs for "No resources found" or error entries
	if selectedResource.Type == "Info" || selectedResource.Type == "Error" {
		a.state.StatusMessage = "No valid pod selected for logs"
		return a, nil
	}

	a.state.SelectedResource = selectedResource
	// Track previous view for backspace navigation
	a.state.PreviousView = a.state.CurrentView
	a.state.CurrentView = types.LogsView
	// Reset scroll position for new resource
	a.state.LogsScrollOffset = 0
	a.state.StatusMessage = fmt.Sprintf("Loading logs for pod: %s", selectedResource.Name)

	return a, a.fetchPodLogs
}

func (a *App) handleDiagramViewKeys(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "q":
		// Show feedback form before quitting
		a.state.CurrentView = types.FeedbackView
		a.state.FeedbackText = ""
		a.state.FeedbackCursorPos = 0
		a.state.ShowFeedbackForm = true
		a.state.FeedbackSubmitting = false
		return a, nil
	case "ctrl+c":
		return a, tea.Quit
	case "esc", "backspace":
		// Reset scroll position when leaving diagram view
		a.state.DiagramScrollOffset = 0
		a.state.CurrentView = types.ResourceView
		return a, nil
	case "j", "down":
		// Scroll down
		a.scrollDiagramDown()
		return a, nil
	case "k", "up":
		// Scroll up
		a.scrollDiagramUp()
		return a, nil
	case "d", "ctrl+d":
		// Page down (half screen)
		a.pageDiagramDown()
		return a, nil
	case "u", "ctrl+u":
		// Page up (half screen)
		a.pageDiagramUp()
		return a, nil
	case "g":
		// Go to top
		a.state.DiagramScrollOffset = 0
		return a, nil
	case "G":
		// Go to bottom
		a.scrollDiagramToBottom()
		return a, nil
	case "x":
		// Copy diagram content
		return a.handleCopyContent()
	case "L":
		// Show logs (if applicable)
		return a.handleShowLogs()
	}
	return a, nil
}

func (a *App) handleMemoryViewKeys(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "q":
		// Show feedback form before quitting
		a.state.CurrentView = types.FeedbackView
		a.state.FeedbackText = ""
		a.state.FeedbackCursorPos = 0
		a.state.ShowFeedbackForm = true
		a.state.FeedbackSubmitting = false
		return a, nil
	case "ctrl+c":
		return a, tea.Quit
	case "esc", "backspace":
		// Reset scroll position when leaving memory view
		a.state.MemoryScrollOffset = 0
		a.state.CurrentView = types.ResourceView
		return a, nil
	case "j", "down":
		// Scroll down
		a.scrollMemoryDown()
		return a, nil
	case "k", "up":
		// Scroll up
		a.scrollMemoryUp()
		return a, nil
	case "d", "ctrl+d":
		// Page down (half screen)
		a.pageMemoryDown()
		return a, nil
	case "u", "ctrl+u":
		// Page up (half screen)
		a.pageMemoryUp()
		return a, nil
	case "g":
		// Go to top
		a.state.MemoryScrollOffset = 0
		return a, nil
	case "G":
		// Go to bottom
		a.scrollMemoryToBottom()
		return a, nil
	case "x":
		// Copy memory content
		return a.handleCopyContent()
	case "L":
		// Show logs (if applicable)
		return a.handleShowLogs()
	}
	return a, nil
}

func (a *App) handleLogsViewKeys(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "q":
		// Show feedback form before quitting
		a.state.CurrentView = types.FeedbackView
		a.state.FeedbackText = ""
		a.state.FeedbackCursorPos = 0
		a.state.ShowFeedbackForm = true
		a.state.FeedbackSubmitting = false
		return a, nil
	case "ctrl+c":
		return a, tea.Quit
	case "esc", "backspace":
		// Reset scroll position when leaving logs view
		a.state.LogsScrollOffset = 0
		a.state.CurrentView = types.ResourceView
		return a, nil
	case "j", "down":
		// Scroll down
		a.scrollLogsDown()
		return a, nil
	case "k", "up":
		// Scroll up
		a.scrollLogsUp()
		return a, nil
	case "d", "ctrl+d":
		// Page down (half screen)
		a.pageLogsDown()
		return a, nil
	case "u", "ctrl+u":
		// Page up (half screen)
		a.pageLogsUp()
		return a, nil
	case "g":
		// Go to top
		a.state.LogsScrollOffset = 0
		return a, nil
	case "G":
		// Go to bottom
		a.scrollLogsToBottom()
		return a, nil
	case "x":
		// Copy logs content
		return a.handleCopyContent()
	}
	return a, nil
}

// Diagram scrolling functions
func (a *App) scrollDiagramDown() {
	if a.state.DiagramScrollOffset < a.getMaxDiagramScroll() {
		a.state.DiagramScrollOffset++
	}
}

func (a *App) scrollDiagramUp() {
	if a.state.DiagramScrollOffset > 0 {
		a.state.DiagramScrollOffset--
	}
}

func (a *App) pageDiagramDown() {
	pageSize := a.getDiagramViewHeight() / 2
	if pageSize < 1 {
		pageSize = 1
	}
	maxScroll := a.getMaxDiagramScroll()
	a.state.DiagramScrollOffset += pageSize
	if a.state.DiagramScrollOffset > maxScroll {
		a.state.DiagramScrollOffset = maxScroll
	}
}

func (a *App) pageDiagramUp() {
	pageSize := a.getDiagramViewHeight() / 2
	if pageSize < 1 {
		pageSize = 1
	}
	a.state.DiagramScrollOffset -= pageSize
	if a.state.DiagramScrollOffset < 0 {
		a.state.DiagramScrollOffset = 0
	}
}

func (a *App) scrollDiagramToBottom() {
	a.state.DiagramScrollOffset = a.getMaxDiagramScroll()
}

func (a *App) getMaxDiagramScroll() int {
	if a.state.DiagramOutput == "" {
		return 0
	}

	lines := strings.Split(a.state.DiagramOutput, "\n")
	contentLines := len(lines)
	viewHeight := a.getDiagramViewHeight()

	maxScroll := contentLines - viewHeight + 2 // +2 for padding
	if maxScroll < 0 {
		maxScroll = 0
	}
	return maxScroll
}

func (a *App) getDiagramViewHeight() int {
	// Account for diagram title, instructions, padding, and borders
	return a.height - 8
}

// Memory scrolling functions
func (a *App) scrollMemoryDown() {
	if a.state.MemoryScrollOffset < a.getMaxMemoryScroll() {
		a.state.MemoryScrollOffset++
	}
}

func (a *App) scrollMemoryUp() {
	if a.state.MemoryScrollOffset > 0 {
		a.state.MemoryScrollOffset--
	}
}

func (a *App) pageMemoryDown() {
	pageSize := a.getMemoryViewHeight() / 2
	if pageSize < 1 {
		pageSize = 1
	}
	newOffset := a.state.MemoryScrollOffset + pageSize
	maxScroll := a.getMaxMemoryScroll()
	if newOffset > maxScroll {
		newOffset = maxScroll
	}
	a.state.MemoryScrollOffset = newOffset
}

func (a *App) pageMemoryUp() {
	pageSize := a.getMemoryViewHeight() / 2
	if pageSize < 1 {
		pageSize = 1
	}
	newOffset := a.state.MemoryScrollOffset - pageSize
	if newOffset < 0 {
		newOffset = 0
	}
	a.state.MemoryScrollOffset = newOffset
}

func (a *App) scrollMemoryToBottom() {
	a.state.MemoryScrollOffset = a.getMaxMemoryScroll()
}

func (a *App) getMaxMemoryScroll() int {
	if a.state.MemoryOutput == "" {
		return 0
	}

	lines := strings.Split(a.state.MemoryOutput, "\n")
	contentLines := len(lines)
	viewHeight := a.getMemoryViewHeight()

	maxScroll := contentLines - viewHeight + 2 // +2 for padding
	if maxScroll < 0 {
		maxScroll = 0
	}
	return maxScroll
}

// Logs scrolling functions
func (a *App) scrollLogsDown() {
	if a.state.LogsScrollOffset < a.getMaxLogsScroll() {
		a.state.LogsScrollOffset++
	}
}

func (a *App) scrollLogsUp() {
	if a.state.LogsScrollOffset > 0 {
		a.state.LogsScrollOffset--
	}
}

func (a *App) pageLogsDown() {
	pageSize := a.getLogsViewHeight() / 2
	if pageSize < 1 {
		pageSize = 1
	}
	newOffset := a.state.LogsScrollOffset + pageSize
	maxScroll := a.getMaxLogsScroll()
	if newOffset > maxScroll {
		newOffset = maxScroll
	}
	a.state.LogsScrollOffset = newOffset
}

func (a *App) pageLogsUp() {
	pageSize := a.getLogsViewHeight() / 2
	if pageSize < 1 {
		pageSize = 1
	}
	newOffset := a.state.LogsScrollOffset - pageSize
	if newOffset < 0 {
		newOffset = 0
	}
	a.state.LogsScrollOffset = newOffset
}

func (a *App) scrollLogsToBottom() {
	a.state.LogsScrollOffset = a.getMaxLogsScroll()
}

func (a *App) getMaxLogsScroll() int {
	if a.state.LogsOutput == "" {
		return 0
	}

	lines := strings.Split(a.state.LogsOutput, "\n")
	contentLines := len(lines)
	viewHeight := a.getLogsViewHeight()

	maxScroll := contentLines - viewHeight + 2 // +2 for padding
	if maxScroll < 0 {
		maxScroll = 0
	}
	return maxScroll
}

// getClickedColumn determines which column was clicked based on X coordinate
func (a *App) getClickedColumn(x int) int {
	columnWidths := a.calculateColumnWidths()
	currentX := 2 // Start after "| "

	for i, width := range columnWidths {
		if x >= currentX && x < currentX+width {
			return i
		}
		currentX += width + 3 // Add width + " | "
	}
	return -1 // No column found
}

// calculateColumnWidths gets column widths from table.go
func (a *App) calculateColumnWidths() []int {
	// Use completely fixed column widths to prevent alignment issues
	fixedWidths := []int{55, 10, 20, 8, 8, 8, 10} // NAME, READY, STATUS, RESTARTS, CPU, MEM, AGE

	// If terminal is very narrow, use smaller fixed widths
	if a.width > 0 && a.width < 100 {
		fixedWidths = []int{40, 8, 15, 6, 6, 6, 8} // Compact mode
	}

	// If terminal is very wide, use larger fixed widths
	if a.width > 140 {
		fixedWidths = []int{65, 12, 25, 10, 10, 10, 12} // Wide mode
	}

	return fixedWidths
}

// handleColumnSort handles sorting when a column header is clicked
func (a *App) handleColumnSort(columnIndex int) (tea.Model, tea.Cmd) {
	// Get column headers based on current resource type
	var headers []string
	switch a.state.ResourceType {
	case "is": // ImageStreams
		headers = []string{"NAME", "DOCKER REPO", "TAGS", "UPDATED", "", "", ""}
	case "pvc": // PersistentVolumeClaims
		headers = []string{"NAME", "STATUS", "VOLUME", "CAPACITY", "ACCESS MODES", "STORAGE CLASS", "AGE"}
	case "services":
		headers = []string{"NAME", "TYPE", "CLUSTER-IP", "EXTERNAL-IP", "PORT(S)", "", "AGE"}
	case "events":
		headers = []string{"NAME", "TYPE", "REASON", "OBJECT", "MESSAGE", "COUNT", "AGE"}
	case "deployments":
		headers = []string{"NAME", "READY", "UP-TO-DATE", "AVAILABLE", "", "", "AGE"}
	case "secrets":
		headers = []string{"NAME", "TYPE", "DATA", "AGE", "", "", ""}
	case "configmaps":
		headers = []string{"NAME", "DATA", "AGE", "", "", "", ""}
	default: // pods
		headers = []string{"NAME", "READY", "STATUS", "RESTARTS", "CPU", "MEM", "AGE"}
	}

	if columnIndex >= len(headers) || headers[columnIndex] == "" {
		return a, nil // Invalid or empty column
	}

	headerName := headers[columnIndex]
	if !a.isColumnSortable(headerName) {
		return a, nil // Column not sortable
	}

	fieldName := a.getFieldNameFromHeader(headerName)

	// Toggle sort direction if same field, otherwise set to ascending
	if a.state.SortField == fieldName {
		if a.state.SortDirection == "asc" {
			a.state.SortDirection = "desc"
		} else {
			a.state.SortDirection = "asc"
		}
	} else {
		a.state.SortField = fieldName
		a.state.SortDirection = "asc"
	}

	// Re-fetch resources to apply new sorting
	return a, a.fetchResources
}

// isColumnSortable and getFieldNameFromHeader are duplicated from table.go
// In a real implementation, these should be moved to a shared location
func (a *App) isColumnSortable(header string) bool {
	sortableHeaders := map[string]bool{
		"NAME":     true,
		"AGE":      true,
		"STATUS":   true,
		"READY":    true,
		"RESTARTS": true,
		"UPDATED":  true,
		"DATA":     true,
		"TYPE":     true,
	}
	return sortableHeaders[header]
}

func (a *App) getFieldNameFromHeader(header string) string {
	headerToField := map[string]string{
		"NAME":     "name",
		"AGE":      "age",
		"STATUS":   "status",
		"READY":    "ready",
		"RESTARTS": "restarts",
		"UPDATED":  "age",
		"DATA":     "ready",
		"TYPE":     "ready",
	}

	if field, exists := headerToField[header]; exists {
		return field
	}
	return "name"
}

// clearStatusMessage clears any existing status message and text selection
func (a *App) clearStatusMessage() {
	// Only clear status message if it's not a recent copy operation message
	if !strings.Contains(a.state.StatusMessage, "Copied") {
		a.state.StatusMessage = ""
	}

	// Clear text selection unless it's active
	if !a.state.SelectionActive {
		a.state.SelectedText = ""
	}
}
