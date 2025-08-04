package ui

import (
	"fmt"

	"github.com/charmbracelet/lipgloss"
)

func (a *App) renderPagination() string {
	paginatedResult := a.paginateResources()

	prevBtn := "◄ Prev"
	if !paginatedResult.HasPrev {
		prevBtn = lipgloss.NewStyle().
			Foreground(lipgloss.Color("240")).
			Render("◄ Prev")
	} else {
		prevBtn = lipgloss.NewStyle().
			Foreground(lipgloss.Color("39")).
			Bold(true).
			Render("◄ Prev")
	}

	nextBtn := "Next ►"
	if !paginatedResult.HasNext {
		nextBtn = lipgloss.NewStyle().
			Foreground(lipgloss.Color("240")).
			Render("Next ►")
	} else {
		nextBtn = lipgloss.NewStyle().
			Foreground(lipgloss.Color("39")).
			Bold(true).
			Render("Next ►")
	}

	pageInfo := fmt.Sprintf("Page %d/%d (%d/%d items)",
		paginatedResult.CurrentPage+1,
		paginatedResult.TotalPages,
		len(paginatedResult.Items),
		paginatedResult.TotalItems)

	helpInfo := "[?] Help │ [q] Quit"

	paginationLine := fmt.Sprintf("%s │ %s │ %s │ %s",
		prevBtn, pageInfo, nextBtn, helpInfo)

	style := lipgloss.NewStyle().
		Foreground(lipgloss.Color("246")).
		Padding(0, 1)

	return style.Render(paginationLine)
}

func (a *App) renderStatusBar() string {
	status := "Ready"
	if a.state.LastUpdate.IsZero() {
		status = "Loading..."
	}

	selectedInfo := ""
	if len(a.state.Resources) > 0 {
		paginatedResult := a.paginateResources()
		if len(paginatedResult.Items) > 0 {
			selectedResource := paginatedResult.Items[0]
			selectedInfo = fmt.Sprintf("Selected: %s", selectedResource.Name)
		}
	}

	statusLine := fmt.Sprintf("Status: %s │ Total: %d │ Displayed: %d │ %s",
		status,
		a.state.TotalItems,
		len(a.paginateResources().Items),
		selectedInfo)

	style := lipgloss.NewStyle().
		Foreground(lipgloss.Color("245")).
		Background(lipgloss.Color("236")).
		Padding(0, 1).
		Width(a.width - 4)

	return style.Render(statusLine)
}
