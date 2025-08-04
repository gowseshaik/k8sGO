package ui

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/smtp"
	"strings"
	"time"

	"k8sgo/internal/types"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// FeedbackSubmittedMsg indicates feedback was successfully submitted
type FeedbackSubmittedMsg struct {
	success bool
	error   string
}

// renderFeedbackView renders the feedback popup dialog
func (a *App) renderFeedbackView() string {
	var sections []string

	// Title
	title := "📝 Feedback & Exit"
	titleStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("39")).
		Background(lipgloss.Color("236")).
		Padding(0, 2).
		Width(60).
		Align(lipgloss.Center)

	// Instructions
	instructions := "Help us improve k8sgo! (Optional - press ESC to skip)"
	instructStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("245")).
		Padding(0, 1).
		Width(60).
		Align(lipgloss.Center)

	// Text input area
	textAreaHeight := 8
	textAreaWidth := 56

	// Split feedback text into lines for display
	lines := strings.Split(a.state.FeedbackText, "\n")

	// Handle cursor position and text wrapping
	displayLines := make([]string, textAreaHeight)
	currentLine := 0

	for i, line := range lines {
		if i < textAreaHeight {
			if len(line) > textAreaWidth {
				// Simple text wrapping
				displayLines[i] = line[:textAreaWidth]
			} else {
				displayLines[i] = line
			}
			currentLine = i
		}
	}

	// Add cursor if we're on the current line
	if currentLine < textAreaHeight && len(displayLines[currentLine]) < textAreaWidth {
		if a.state.FeedbackCursorPos >= len(a.state.FeedbackText) {
			displayLines[currentLine] += "│" // Cursor at end
		}
	}

	// Pad remaining lines
	for i := len(lines); i < textAreaHeight; i++ {
		displayLines[i] = ""
	}

	textContent := strings.Join(displayLines, "\n")

	textAreaStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("240")).
		Padding(1).
		Width(textAreaWidth + 4).
		Height(textAreaHeight + 2).
		Foreground(lipgloss.Color("252"))

	// Action buttons
	var buttons string
	if a.state.FeedbackSubmitting {
		buttons = "⏳ Submitting feedback..."
		buttonStyle := lipgloss.NewStyle().
			Foreground(lipgloss.Color("226")).
			Background(lipgloss.Color("235")).
			Padding(0, 2).
			Width(60).
			Align(lipgloss.Center)
		buttons = buttonStyle.Render(buttons)
	} else {
		submitBtn := "[Enter] Submit & Exit"
		skipBtn := "[ESC] Skip & Exit"
		backBtn := "[Ctrl+C] Cancel"

		submitStyle := lipgloss.NewStyle().
			Foreground(lipgloss.Color("46")).
			Bold(true)
		skipStyle := lipgloss.NewStyle().
			Foreground(lipgloss.Color("208"))
		backStyle := lipgloss.NewStyle().
			Foreground(lipgloss.Color("240"))

		buttons = fmt.Sprintf("%s  %s  %s",
			submitStyle.Render(submitBtn),
			skipStyle.Render(skipBtn),
			backStyle.Render(backBtn))
	}

	// Email info
	emailInfo := "Feedback will be sent to: k8sgo-feedback@example.com"
	emailStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("242")).
		Italic(true).
		Width(60).
		Align(lipgloss.Center)

	// Assemble the popup
	sections = append(sections,
		titleStyle.Render(title),
		"",
		instructStyle.Render(instructions),
		"",
		textAreaStyle.Render(textContent),
		"",
		buttons,
		"",
		emailStyle.Render(emailInfo),
	)

	content := lipgloss.JoinVertical(lipgloss.Center, sections...)

	// Create popup container
	popupStyle := lipgloss.NewStyle().
		Border(lipgloss.ThickBorder()).
		BorderForeground(lipgloss.Color("39")).
		Background(lipgloss.Color("233")).
		Padding(2).
		Width(70).
		Align(lipgloss.Center)

	popup := popupStyle.Render(content)

	// Center the popup on screen
	screenStyle := lipgloss.NewStyle().
		Width(a.width).
		Height(a.height).
		Align(lipgloss.Center, lipgloss.Center).
		Background(lipgloss.Color("232")) // Semi-transparent background

	return screenStyle.Render(popup)
}

// handleFeedbackViewKeys handles keyboard input in feedback view
func (a *App) handleFeedbackViewKeys(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	if a.state.FeedbackSubmitting {
		// Ignore input while submitting
		return a, nil
	}

	switch msg.String() {
	case "ctrl+c":
		// Cancel and return to main view
		a.state.CurrentView = types.ResourceView
		a.state.FeedbackText = ""
		a.state.FeedbackCursorPos = 0
		a.state.ShowFeedbackForm = false
		return a, nil

	case "esc":
		// Skip feedback and quit
		return a, tea.Quit

	case "enter":
		// Submit feedback and quit
		if strings.TrimSpace(a.state.FeedbackText) != "" {
			a.state.FeedbackSubmitting = true
			return a, a.submitFeedback
		}
		// If no feedback text, just quit
		return a, tea.Quit

	case "backspace":
		// Remove character
		if len(a.state.FeedbackText) > 0 && a.state.FeedbackCursorPos > 0 {
			textRunes := []rune(a.state.FeedbackText)
			if a.state.FeedbackCursorPos <= len(textRunes) {
				textRunes = append(textRunes[:a.state.FeedbackCursorPos-1], textRunes[a.state.FeedbackCursorPos:]...)
				a.state.FeedbackText = string(textRunes)
				a.state.FeedbackCursorPos--
			}
		}
		return a, nil

	case "left":
		if a.state.FeedbackCursorPos > 0 {
			a.state.FeedbackCursorPos--
		}
		return a, nil

	case "right":
		if a.state.FeedbackCursorPos < len([]rune(a.state.FeedbackText)) {
			a.state.FeedbackCursorPos++
		}
		return a, nil

	default:
		// Add character to feedback text
		if len(msg.String()) == 1 && msg.String() != "\x00" {
			char := msg.String()
			textRunes := []rune(a.state.FeedbackText)

			// Insert character at cursor position
			if a.state.FeedbackCursorPos >= len(textRunes) {
				a.state.FeedbackText += char
			} else {
				newRunes := make([]rune, len(textRunes)+1)
				copy(newRunes[:a.state.FeedbackCursorPos], textRunes[:a.state.FeedbackCursorPos])
				newRunes[a.state.FeedbackCursorPos] = []rune(char)[0]
				copy(newRunes[a.state.FeedbackCursorPos+1:], textRunes[a.state.FeedbackCursorPos:])
				a.state.FeedbackText = string(newRunes)
			}
			a.state.FeedbackCursorPos++

			// Limit feedback length
			if len(a.state.FeedbackText) > 500 {
				a.state.FeedbackText = a.state.FeedbackText[:500]
				a.state.FeedbackCursorPos = 500
			}
		}
		return a, nil
	}
}

// submitFeedback sends the feedback via email
func (a *App) submitFeedback() tea.Msg {
	feedback := strings.TrimSpace(a.state.FeedbackText)
	if feedback == "" {
		return FeedbackSubmittedMsg{success: true, error: ""} // No feedback to send
	}

	// Create feedback data
	feedbackData := map[string]interface{}{
		"tool":      "k8sgo",
		"version":   a.state.Version,
		"feedback":  feedback,
		"timestamp": time.Now().Format(time.RFC3339),
		"context":   a.state.Context,
		"namespace": a.state.Namespace,
		"cli_tool":  a.state.Tool,
	}

	// Try multiple methods to send feedback
	var err error

	// Method 1: Try webhook/API endpoint (most reliable)
	err = a.sendFeedbackViaWebhook(feedbackData)
	if err == nil {
		return FeedbackSubmittedMsg{success: true, error: ""}
	}

	// Method 2: Try email (fallback)
	err = a.sendFeedbackViaEmail(feedbackData)
	if err == nil {
		return FeedbackSubmittedMsg{success: true, error: ""}
	}

	// Method 3: Save locally (last resort)
	err = a.saveFeedbackLocally(feedbackData)
	if err == nil {
		return FeedbackSubmittedMsg{success: true, error: "Feedback saved locally"}
	}

	return FeedbackSubmittedMsg{success: false, error: err.Error()}
}

// sendFeedbackViaWebhook sends feedback to a webhook endpoint
func (a *App) sendFeedbackViaWebhook(data map[string]interface{}) error {
	// This would be configured to point to your feedback collection service
	webhookURL := "https://api.example.com/k8sgo-feedback" // Replace with your endpoint

	jsonData, err := json.Marshal(data)
	if err != nil {
		return err
	}

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Post(webhookURL, "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("webhook returned status %d", resp.StatusCode)
	}

	return nil
}

// sendFeedbackViaEmail sends feedback via SMTP email
func (a *App) sendFeedbackViaEmail(data map[string]interface{}) error {
	// Email configuration - you would need to set these up
	smtpHost := "smtp.gmail.com"
	smtpPort := "587"
	fromEmail := "k8sgo-feedback@example.com"
	toEmail := "developer@example.com"
	password := "your-app-password" // Use app-specific password for Gmail

	// Create email message
	subject := "k8sgo Feedback"
	body := fmt.Sprintf(`
New feedback from k8sgo:

Tool: %v
Version: %v
Context: %v
Namespace: %v
CLI Tool: %v
Timestamp: %v

Feedback:
%v
`,
		data["tool"], data["version"], data["context"],
		data["namespace"], data["cli_tool"], data["timestamp"],
		data["feedback"])

	message := fmt.Sprintf("To: %s\r\nSubject: %s\r\n\r\n%s", toEmail, subject, body)

	// Send email
	auth := smtp.PlainAuth("", fromEmail, password, smtpHost)
	err := smtp.SendMail(smtpHost+":"+smtpPort, auth, fromEmail, []string{toEmail}, []byte(message))

	return err
}

// saveFeedbackLocally saves feedback to a local file as fallback
func (a *App) saveFeedbackLocally(data map[string]interface{}) error {
	// This is a fallback method - save to local file
	// In production, you might want to save to a temp directory or user config directory
	return fmt.Errorf("local save not implemented") // Placeholder
}
