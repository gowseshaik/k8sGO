package ui

import (
	"fmt"
	"strings"
	"time"

	"k8sgo/internal/types"

	"github.com/charmbracelet/lipgloss"
)

func (a *App) renderResourceTable() string {
	if len(a.state.Resources) == 0 {
		return a.renderEmptyTable()
	}

	paginatedResources := a.paginateResources()

	// Force table structure for all resource types
	header := a.renderTableHeader()
	rows := a.renderTableRows(paginatedResources.Items)

	// Ensure we have content
	if header == "" || rows == "" {
		return a.renderEmptyTable()
	}

	// Add table borders
	table := lipgloss.JoinVertical(lipgloss.Left, header, rows)

	// Calculate border width based on column widths (same as header)
	columnWidths := a.calculateColumnWidths()
	totalWidth := 2 // For "| " and " |"
	for i, width := range columnWidths {
		totalWidth += width
		if i < len(columnWidths)-1 {
			totalWidth += 3 // For " | "
		}
	}

	// Add bottom border matching the header width
	bottomBorder := "+" + strings.Repeat("-", totalWidth) + "+"

	return lipgloss.JoinVertical(lipgloss.Left, table, bottomBorder)
}

func (a *App) renderTableHeader() string {
	var headers []string

	// Use resource-specific headers
	switch a.state.ResourceType {
	case "is": // ImageStreams
		headers = []string{
			"NAME",
			"DOCKER REPO",
			"TAGS",
			"UPDATED",
			"", // Empty for alignment
			"", // Empty for alignment
			"", // Empty for alignment
		}
	case "pvc": // PersistentVolumeClaims
		headers = []string{
			"NAME",
			"STATUS",
			"VOLUME",
			"CAPACITY",
			"ACCESS MODES",
			"STORAGE CLASS",
			"AGE",
		}
	case "services":
		headers = []string{
			"NAME",
			"TYPE",
			"CLUSTER-IP",
			"EXTERNAL-IP",
			"PORT(S)",
			"",
			"AGE",
		}
	case "events":
		headers = []string{
			"NAME",
			"TYPE",
			"REASON",
			"OBJECT",
			"MESSAGE",
			"COUNT",
			"AGE",
		}
	case "deployments":
		headers = []string{
			"NAME",
			"READY",
			"UP-TO-DATE",
			"AVAILABLE",
			"",
			"",
			"AGE",
		}
	case "secrets":
		headers = []string{
			"NAME",
			"TYPE",
			"DATA",
			"AGE",
			"",
			"",
			"",
		}
	case "configmaps":
		headers = []string{
			"NAME",
			"DATA",
			"AGE",
			"",
			"",
			"",
			"",
		}
	default: // pods, events
		headers = []string{
			"NAME",
			"READY",
			"STATUS",
			"RESTARTS",
			"CPU",
			"MEM",
			"AGE",
		}
	}

	// Calculate dynamic column widths based on content
	columnWidths := a.calculateColumnWidths()

	var headerCells []string
	for i, header := range headers {
		// Add sorting indicators for sortable columns
		headerText := header
		if header != "" && a.isColumnSortable(header) {
			sortIcon := a.getSortIcon(header)
			headerText = header + " " + sortIcon
		}

		cell := lipgloss.NewStyle().
			Width(columnWidths[i]).
			Align(lipgloss.Left).
			Bold(true).
			Foreground(lipgloss.Color("39")).
			Render(headerText)
		headerCells = append(headerCells, cell)
	}

	headerRow := "| " + strings.Join(headerCells, " | ") + " |"

	// Calculate border width based on actual column widths
	totalWidth := 2 // For "| " and " |"
	for i, width := range columnWidths {
		totalWidth += width
		if i < len(columnWidths)-1 {
			totalWidth += 3 // For " | "
		}
	}

	// Add top border
	topBorder := "+" + strings.Repeat("-", totalWidth) + "+"
	separator := "+" + strings.Repeat("-", totalWidth) + "+"

	return lipgloss.JoinVertical(lipgloss.Left, topBorder, headerRow, separator)
}

func (a *App) renderTableRows(resources []types.Resource) string {
	if len(resources) == 0 {
		return a.renderEmptyRows()
	}

	columnWidths := a.calculateColumnWidths()
	var rows []string

	for i, resource := range resources {
		isSelected := i == a.getSelectedIndex()
		isEvenRow := i%2 == 0
		row := a.renderTableRow(resource, columnWidths, isSelected, isEvenRow)
		rows = append(rows, row)
	}

	remainingRows := a.state.PageSize - len(resources)
	for i := 0; i < remainingRows; i++ {
		emptyRow := a.renderEmptyRow(columnWidths)
		rows = append(rows, emptyRow)
	}

	return strings.Join(rows, "\n")
}

func (a *App) renderTableRow(resource types.Resource, widths []int, isSelected bool, isEvenRow bool) string {
	age := formatDuration(resource.Age)

	// Prepare cell content based on resource type and truncate if necessary
	var cells []string

	switch a.state.ResourceType {
	case "is": // ImageStreams
		cells = []string{
			truncateString(resource.Name, widths[0]),
			truncateString(resource.Status, widths[1]), // Docker repo in Status field
			truncateString(resource.Ready, widths[2]),  // Tags in Ready field
			truncateString(age, widths[3]),             // Updated time
			"", "", "",
		}
	case "pvc": // PersistentVolumeClaims
		cells = []string{
			truncateString(resource.Name, widths[0]),
			truncateString(resource.Ready, widths[1]),        // STATUS
			truncateString(resource.Volume, widths[2]),       // VOLUME
			truncateString(resource.Capacity, widths[3]),     // CAPACITY
			truncateString(resource.AccessModes, widths[4]),  // ACCESS MODES
			truncateString(resource.StorageClass, widths[5]), // STORAGE CLASS
			truncateString(age, widths[6]),                   // AGE
		}
	case "services":
		cells = []string{
			truncateString(resource.Name, widths[0]),
			truncateString(resource.Ready, widths[1]),  // TYPE
			truncateString(resource.Status, widths[2]), // CLUSTER-IP
			truncateString(resource.CPU, widths[3]),    // EXTERNAL-IP (stored in CPU field)
			truncateString(resource.Memory, widths[4]), // PORT(S) (stored in Memory field)
			"", // Empty
			truncateString(age, widths[6]),
		}
	case "deployments":
		cells = []string{
			truncateString(resource.Name, widths[0]),
			truncateString(resource.Ready, widths[1]),                       // READY
			truncateString(resource.Status, widths[2]),                      // UP-TO-DATE
			truncateString(fmt.Sprintf("%d", resource.Restarts), widths[3]), // AVAILABLE
			"", "", // Empty columns
			truncateString(age, widths[6]),
		}
	case "secrets":
		cells = []string{
			truncateString(resource.Name, widths[0]),
			truncateString(resource.Ready, widths[1]),  // TYPE
			truncateString(resource.Status, widths[2]), // DATA
			truncateString(age, widths[3]),             // AGE
			"", "", "",                                 // Empty columns
		}
	case "configmaps":
		cells = []string{
			truncateString(resource.Name, widths[0]),
			truncateString(resource.Status, widths[1]), // DATA (stored in Status field)
			truncateString(age, widths[2]),             // AGE
			"", "", "", "",                             // Empty columns
		}
	default: // pods, events
		cells = []string{
			truncateString(resource.Name, widths[0]),
			truncateString(resource.Ready, widths[1]),
			truncateString(resource.Status, widths[2]),
			truncateString(fmt.Sprintf("%d", resource.Restarts), widths[3]),
			truncateString(resource.CPU, widths[4]),
			truncateString(resource.Memory, widths[5]),
			truncateString(age, widths[6]),
		}
	}

	var formattedCells []string
	for i, cell := range cells {
		// Use fixed width for all columns to maintain perfect alignment
		style := lipgloss.NewStyle().
			Width(widths[i]).
			Align(lipgloss.Left).
			Inline(true)

		if isSelected {
			style = style.Foreground(lipgloss.Color("0")).Background(lipgloss.Color("39"))
		} else {
			// Apply alternating row background colors with more visible contrast
			if isEvenRow {
				style = style.Background(lipgloss.Color("240")) // More visible gray for even rows
			} else {
				style = style.Background(lipgloss.Color("237")) // Darker gray for odd rows
			}

			// Apply status-based text colors with better contrast
			switch resource.Status {
			case "Running", "Ready", "Active", "Bound":
				style = style.Foreground(lipgloss.Color("82")) // Bright green
			case "Pending", "Partial":
				style = style.Foreground(lipgloss.Color("214")) // Orange
			case "Error", "Failed", "Not Ready":
				style = style.Foreground(lipgloss.Color("196")) // Red
			default:
				style = style.Foreground(lipgloss.Color("255")) // White text for better contrast
			}
		}

		formattedCells = append(formattedCells, style.Render(cell))
	}

	prefix := " "
	if isSelected {
		prefix = "►"
	}

	rowContent := fmt.Sprintf("|%s %s |", prefix, strings.Join(formattedCells, " | "))

	// Apply text selection highlighting if active and in resource view
	if a.state.SelectionActive && a.state.CurrentView == types.ResourceView {
		// This is a simplified approach - for more precise highlighting,
		// you'd need to calculate exact character positions within the table
		rowContent = a.applyRowSelectionHighlight(rowContent)
	}

	return rowContent
}

func (a *App) renderEmptyRow(widths []int) string {
	var cells []string
	for _, width := range widths {
		cell := lipgloss.NewStyle().
			Width(width).
			Foreground(lipgloss.Color("240")).
			Render("")
		cells = append(cells, cell)
	}

	return "|  " + strings.Join(cells, " | ") + " |"
}

func (a *App) renderEmptyTable() string {
	message := "No resources found"
	style := lipgloss.NewStyle().
		Foreground(lipgloss.Color("240")).
		Align(lipgloss.Center).
		Width(a.width - 6)

	return style.Render(message)
}

func (a *App) renderEmptyRows() string {
	var rows []string
	columnWidths := a.calculateColumnWidths()

	for i := 0; i < a.state.PageSize; i++ {
		row := a.renderEmptyRow(columnWidths)
		rows = append(rows, row)
	}

	return strings.Join(rows, "\n")
}

func (a *App) paginateResources() types.PaginatedResult {
	totalItems := len(a.state.Resources)
	totalPages := (totalItems + a.state.PageSize - 1) / a.state.PageSize

	if totalPages == 0 {
		totalPages = 1
	}

	start := a.state.CurrentPage * a.state.PageSize
	end := start + a.state.PageSize

	if start >= totalItems {
		return types.PaginatedResult{
			Items:       []types.Resource{},
			CurrentPage: a.state.CurrentPage,
			TotalPages:  totalPages,
			TotalItems:  totalItems,
			HasNext:     false,
			HasPrev:     a.state.CurrentPage > 0,
		}
	}

	if end > totalItems {
		end = totalItems
	}

	return types.PaginatedResult{
		Items:       a.state.Resources[start:end],
		CurrentPage: a.state.CurrentPage,
		TotalPages:  totalPages,
		TotalItems:  totalItems,
		HasNext:     a.state.CurrentPage < totalPages-1,
		HasPrev:     a.state.CurrentPage > 0,
	}
}

func (a *App) getSelectedIndex() int {
	// Get the index within the current page
	pageStart := a.state.CurrentPage * a.state.PageSize
	return a.state.SelectedIndex - pageStart
}

func formatDuration(d time.Duration) string {
	if d < time.Minute {
		return fmt.Sprintf("%ds", int(d.Seconds()))
	}
	if d < time.Hour {
		return fmt.Sprintf("%dm", int(d.Minutes()))
	}
	if d < 24*time.Hour {
		return fmt.Sprintf("%dh", int(d.Hours()))
	}
	return fmt.Sprintf("%dd", int(d.Hours()/24))
}

func truncateString(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	if maxLen <= 3 {
		return s[:maxLen]
	}
	return s[:maxLen-3] + "..."
}

// applyRowSelectionHighlight applies text selection highlighting to table rows
func (a *App) applyRowSelectionHighlight(rowContent string) string {
	// For table rows, we'll use a simpler approach since the exact character mapping
	// is complex due to styled cells. This provides basic visual feedback.

	// Normalize selection coordinates
	startX, startY := a.state.SelectionStartX, a.state.SelectionStartY
	endX, endY := a.state.SelectionEndX, a.state.SelectionEndY

	if startY > endY || (startY == endY && startX > endX) {
		startX, endX = endX, startX
		startY, endY = endY, startY
	}

	// Calculate approximate table area (this would need refinement for precise positioning)
	headerOffset := 6               // Approximate lines for banner, context, search, etc.
	tableStartY := headerOffset + 3 // Header + borders

	// For now, just show a selection indicator if the row might be in selection
	// A more sophisticated implementation would require parsing the styled content
	_ = startX
	_ = endX
	_ = startY
	_ = endY
	_ = tableStartY

	// Return the row as-is for now - the main text selection will work in describe/tags views
	// Table selection is handled by the existing row selection mechanism
	return rowContent
}

// getSortIcon returns the appropriate sort icon for a column
func (a *App) getSortIcon(header string) string {
	fieldName := a.getFieldNameFromHeader(header)

	if a.state.SortField == fieldName {
		if a.state.SortDirection == "asc" {
			return "▲"
		} else {
			return "▼"
		}
	}
	return "↕" // Neutral sort icon
}
