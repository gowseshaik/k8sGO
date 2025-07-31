package ui

import (
	"github.com/charmbracelet/lipgloss"
)

func (a *App) renderHelpView() string {
	var sections []string

	// Add banner
	if a.state.ShowBanner {
		sections = append(sections, a.renderBanner())
	}

	helpContent := `
Keyboard Shortcuts:

Navigation:
  ↑/k         Move up
  ↓/j         Move down  
  ←/h         Previous page
  →/l         Next page
  Tab         Navigate between sections

Actions:
  Enter       Select/view resource details
  /           Focus search box
  d           Describe selected resource
  r           Refresh resources
  c           Switch context
  ?           Toggle this help
  q           Quit application
  Esc         Return to previous view

Resource Management:
  1           Switch to Pods view
  2           Switch to Services view
  3           Switch to Deployments view
  4           Switch to PVC view
  5           Switch to ImageStreams view
  6           Switch to Secrets view
  7           Switch to ConfigMaps view
  8           Switch to Events view

Context & Namespace:
  Automatically detects current context
  Displays current namespace
  Shows available tools (kubectl/oc)

Status Indicators:
  ► symbol    Currently selected item
  Green       Running/Ready/Active resources
  Yellow      Pending/Partial resources  
  Red         Error/Failed resources
  Gray        Unknown status

Pagination:
  Fixed 50 items per page
  ◄ Prev / Next ► for navigation
  Page indicator shows current position

Press ? again or Esc to return to main view
`

	style := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("39")).
		Padding(2).
		Width(a.width - 4).
		Height(a.height - 4).
		Foreground(lipgloss.Color("252"))

	titleStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("39")).
		Align(lipgloss.Center).
		Width(a.width - 8)

	title := titleStyle.Render("🔮 K8sgo Help")
	helpContentStyled := lipgloss.JoinVertical(lipgloss.Left, title, helpContent)

	sections = append(sections, helpContentStyled)
	content := lipgloss.JoinVertical(lipgloss.Left, sections...)

	return style.Render(content)
}
