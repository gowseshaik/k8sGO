package ui

import (
	"fmt"
	"strings"

	"k8sgo/pkg/utils"

	"github.com/charmbracelet/lipgloss"
)

type NamespaceSelector struct {
	namespaces       []string
	currentNamespace string
	selectedIndex    int
	ocCommands       *utils.OcCommands
}

func NewNamespaceSelector() *NamespaceSelector {
	return &NamespaceSelector{
		ocCommands:    utils.NewOcCommands(),
		selectedIndex: 0,
	}
}

func (ns *NamespaceSelector) SetNamespaces(namespaces []string, currentNamespace string) {
	ns.namespaces = namespaces
	ns.currentNamespace = currentNamespace

	// Set selected index to current namespace
	for i, namespace := range namespaces {
		if namespace == currentNamespace {
			ns.selectedIndex = i
			break
		}
	}
}

func (ns *NamespaceSelector) MoveUp() {
	if ns.selectedIndex > 0 {
		ns.selectedIndex--
	}
}

func (ns *NamespaceSelector) MoveDown() {
	if ns.selectedIndex < len(ns.namespaces)-1 {
		ns.selectedIndex++
	}
}

func (ns *NamespaceSelector) GetSelectedNamespace() string {
	if ns.selectedIndex >= 0 && ns.selectedIndex < len(ns.namespaces) {
		return ns.namespaces[ns.selectedIndex]
	}
	return ""
}

func (ns *NamespaceSelector) Render(width, height int) string {
	// Title
	titleStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("39")).
		Align(lipgloss.Center).
		Width(width - 4).
		MarginBottom(1)

	title := titleStyle.Render("📂 Select Namespace/Project")

	// Instructions
	instructionsStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("245")).
		Align(lipgloss.Center).
		Width(width - 4).
		MarginBottom(2)

	instructions := instructionsStyle.Render("Use ↑/↓ or j/k to navigate • Enter to select • ESC to cancel")

	// Namespace list
	var namespaceItems []string
	maxVisible := height - 10 // Account for title, instructions, and borders

	// Calculate scroll offset to keep selected item visible
	startIndex := 0
	if len(ns.namespaces) > maxVisible {
		if ns.selectedIndex >= maxVisible/2 {
			startIndex = ns.selectedIndex - maxVisible/2
			if startIndex+maxVisible > len(ns.namespaces) {
				startIndex = len(ns.namespaces) - maxVisible
			}
		}
	}

	endIndex := startIndex + maxVisible
	if endIndex > len(ns.namespaces) {
		endIndex = len(ns.namespaces)
	}

	for i := startIndex; i < endIndex; i++ {
		namespace := ns.namespaces[i]

		// Check if this is the currently active namespace
		prefix := "  "
		if namespace == ns.currentNamespace {
			prefix = "● " // Current namespace indicator
		}

		// Style for selected item
		if i == ns.selectedIndex {
			selectedStyle := lipgloss.NewStyle().
				Foreground(lipgloss.Color("15")).
				Background(lipgloss.Color("39")).
				Bold(true).
				Padding(0, 1).
				Width(width - 8)

			namespaceItems = append(namespaceItems, selectedStyle.Render(fmt.Sprintf("%s%s", prefix, namespace)))
		} else {
			// Style for unselected items
			unselectedStyle := lipgloss.NewStyle().
				Foreground(lipgloss.Color("252")).
				Padding(0, 1).
				Width(width - 8)

			if namespace == ns.currentNamespace {
				// Highlight current namespace with different color
				unselectedStyle = unselectedStyle.Foreground(lipgloss.Color("82")) // Green for current
			}

			namespaceItems = append(namespaceItems, unselectedStyle.Render(fmt.Sprintf("%s%s", prefix, namespace)))
		}
	}

	// Add scroll indicators if needed
	if len(ns.namespaces) > maxVisible {
		scrollInfo := fmt.Sprintf("Showing %d-%d of %d namespaces", startIndex+1, endIndex, len(ns.namespaces))
		scrollStyle := lipgloss.NewStyle().
			Foreground(lipgloss.Color("240")).
			Align(lipgloss.Center).
			Width(width - 4).
			MarginTop(1)

		namespaceItems = append(namespaceItems, scrollStyle.Render(scrollInfo))
	}

	namespacesContent := strings.Join(namespaceItems, "\n")

	// Status info
	statusStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("245")).
		Align(lipgloss.Center).
		Width(width - 4).
		MarginTop(2)

	var statusText string
	if ns.currentNamespace != "" {
		statusText = fmt.Sprintf("Current: %s", ns.currentNamespace)
	} else {
		statusText = "No namespace selected"
	}
	status := statusStyle.Render(statusText)

	// Combine all parts
	content := lipgloss.JoinVertical(lipgloss.Center,
		title,
		instructions,
		namespacesContent,
		status,
	)

	// Wrap in border
	borderStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("240")).
		Padding(1).
		Width(width-2).
		Height(height-2).
		Align(lipgloss.Center, lipgloss.Center)

	return borderStyle.Render(content)
}
