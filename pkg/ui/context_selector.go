package ui

import (
	"fmt"
	"strings"

	"k8sgo/pkg/utils"

	"github.com/charmbracelet/lipgloss"
)

type ContextSelector struct {
	contexts       []string
	currentContext string
	selectedIndex  int
	ocCommands     *utils.OcCommands
}

func NewContextSelector() *ContextSelector {
	return &ContextSelector{
		ocCommands:    utils.NewOcCommands(),
		selectedIndex: 0,
	}
}

func (cs *ContextSelector) LoadContexts() error {
	contexts, current, err := cs.ocCommands.GetContexts()
	if err != nil {
		return fmt.Errorf("failed to get contexts: %w", err)
	}

	cs.contexts = contexts
	cs.currentContext = current

	// Set selected index to current context
	for i, ctx := range contexts {
		if ctx == current {
			cs.selectedIndex = i
			break
		}
	}

	return nil
}

func (cs *ContextSelector) MoveUp() {
	if cs.selectedIndex > 0 {
		cs.selectedIndex--
	}
}

func (cs *ContextSelector) MoveDown() {
	if cs.selectedIndex < len(cs.contexts)-1 {
		cs.selectedIndex++
	}
}

func (cs *ContextSelector) SelectContext() error {
	if cs.selectedIndex < 0 || cs.selectedIndex >= len(cs.contexts) {
		return fmt.Errorf("invalid context selection")
	}

	selectedContext := cs.contexts[cs.selectedIndex]
	err := cs.ocCommands.SwitchContext(selectedContext)
	if err != nil {
		return fmt.Errorf("failed to switch context: %w", err)
	}

	cs.currentContext = selectedContext
	return nil
}

func (cs *ContextSelector) Render(width, height int) string {
	var sections []string

	// Add banner at top
	sections = append(sections, cs.renderBanner())

	if len(cs.contexts) == 0 {
		// Show a more informative message with border
		message := `No contexts available
		
Please check your kubeconfig file:
- ~/.kube/config (Linux/Mac)
- %USERPROFILE%\.kube\config (Windows)

Press Esc to continue anyway.`

		style := lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("196")).
			Padding(2).
			Width(width - 4).
			Align(lipgloss.Center)

		sections = append(sections, style.Render(message))
		return lipgloss.JoinVertical(lipgloss.Left, sections...)
	}

	var lines []string

	// Title
	title := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("39")).
		Align(lipgloss.Center).
		Width(width - 4).
		Render("Select Context")

	lines = append(lines, title)
	lines = append(lines, "")

	// Context list
	for i, context := range cs.contexts {
		prefix := "  "
		style := lipgloss.NewStyle().Foreground(lipgloss.Color("252"))

		if i == cs.selectedIndex {
			prefix = "► "
			style = style.Background(lipgloss.Color("39")).Foreground(lipgloss.Color("0"))
		}

		if context == cs.currentContext {
			context = fmt.Sprintf("%s (current)", context)
			if i != cs.selectedIndex {
				style = style.Foreground(lipgloss.Color("46"))
			}
		}

		line := fmt.Sprintf("%s%s", prefix, context)
		lines = append(lines, style.Width(width-6).Render(line))
	}

	lines = append(lines, "")
	lines = append(lines, "↑/k: Move up  ↓/j: Move down  Enter: Select  Esc: Cancel")

	content := strings.Join(lines, "\n")

	frame := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("39")).
		Padding(1).
		Width(width - 2).
		Height(height - 10) // Adjust height for banner

	sections = append(sections, frame.Render(content))
	return lipgloss.JoinVertical(lipgloss.Left, sections...)
}

func (cs *ContextSelector) renderBanner() string {
	// ASCII art for k8sGO (same as main app)
	asciiArt := `
 __   ___  _______    ________  _______     ______    
|/"| /  ")/"  _  \  /"       )/" _   "|   /    " \   
(: |/   /|:  _ /  :|(:   \___/(: ( \___)  // ____  \  
|    __/  \___/___/  \___  \   \/ \      /  /    ) :) 
(// _  \  //  /_ \\   __/  \\  //  \ ___(: (____/ //  
|: | \  \|:  /_   :| /" \   :)(:   _(  _|\        /   
(__|  \__)\_______/ (_______/  \_______)  \"_____/    
                                                      `

	// Style for ASCII art
	artStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("39")).
		Align(lipgloss.Center)

	return artStyle.Render(asciiArt)
}

func (cs *ContextSelector) GetSelectedContext() string {
	if cs.selectedIndex < 0 || cs.selectedIndex >= len(cs.contexts) {
		return ""
	}
	return cs.contexts[cs.selectedIndex]
}
