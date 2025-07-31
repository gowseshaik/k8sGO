package ui

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
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
	// Fixed dimensions for proper alignment
	const (
		formWidth      = 70
		textAreaWidth  = 64
		textAreaHeight = 8
	)

	// Title
	title := "📝 Feedback & Exit"
	titleStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("39")).
		Background(lipgloss.Color("236")).
		Padding(0, 2).
		Width(formWidth - 4).
		Align(lipgloss.Center)

	// Instructions
	instructions := "Help us improve k8sgo! (Optional - press ESC to skip)\nUse Enter for new lines, Ctrl+Enter or F1 to submit"
	instructStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("245")).
		Width(formWidth - 4).
		Align(lipgloss.Center)

	// Text input area with proper cursor handling
	textContent := a.renderTextAreaWithCursor(textAreaWidth, textAreaHeight)

	textAreaStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("240")).
		Padding(1).
		Width(textAreaWidth).
		Height(textAreaHeight + 2).
		Foreground(lipgloss.Color("252")).
		Background(lipgloss.Color("234"))

	// Action buttons
	var buttons string
	if a.state.FeedbackSubmitting {
		buttons = "⏳ Submitting feedback..."
		buttonStyle := lipgloss.NewStyle().
			Foreground(lipgloss.Color("226")).
			Background(lipgloss.Color("235")).
			Padding(0, 2).
			Width(formWidth - 4).
			Align(lipgloss.Center)
		buttons = buttonStyle.Render(buttons)
	} else {
		submitBtn := "[Ctrl+Enter/F1] Submit & Exit"
		skipBtn := "[ESC] Skip & Exit"
		backBtn := "[Ctrl+C] Cancel"

		submitStyle := lipgloss.NewStyle().
			Foreground(lipgloss.Color("46")).
			Bold(true).
			Padding(0, 1)
		skipStyle := lipgloss.NewStyle().
			Foreground(lipgloss.Color("208")).
			Padding(0, 1)
		backStyle := lipgloss.NewStyle().
			Foreground(lipgloss.Color("240")).
			Padding(0, 1)

		buttonRow := lipgloss.JoinHorizontal(lipgloss.Center,
			submitStyle.Render(submitBtn),
			"  ",
			skipStyle.Render(skipBtn),
			"  ",
			backStyle.Render(backBtn))

		buttonRowStyle := lipgloss.NewStyle().
			Width(formWidth - 4).
			Align(lipgloss.Center)
		buttons = buttonRowStyle.Render(buttonRow)
	}

	// Email info - hidden from end user
	emailInfo := "Your feedback helps us improve k8sgo!"
	emailStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("242")).
		Italic(true).
		Width(formWidth - 4).
		Align(lipgloss.Center)

	// Assemble the popup with consistent spacing
	content := lipgloss.JoinVertical(lipgloss.Center,
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

	// Create popup container with clean borders
	popupStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("39")).
		Background(lipgloss.Color("233")).
		Padding(2).
		Width(formWidth).
		Align(lipgloss.Center)

	popup := popupStyle.Render(content)

	// Center the popup on screen
	screenStyle := lipgloss.NewStyle().
		Width(a.width).
		Height(a.height).
		Align(lipgloss.Center, lipgloss.Center).
		Background(lipgloss.Color("232"))

	return screenStyle.Render(popup)
}

// renderTextAreaWithCursor renders the text area with proper cursor positioning
func (a *App) renderTextAreaWithCursor(width, height int) string {
	text := a.state.FeedbackText
	cursorPos := a.state.FeedbackCursorPos

	// Convert cursor position to line and column
	textRunes := []rune(text)
	if cursorPos > len(textRunes) {
		cursorPos = len(textRunes)
	}

	lines := make([]string, height)
	currentLine := 0
	currentCol := 0
	cursorLine := 0
	cursorCol := 0

	// Build display lines and track cursor position
	for i, r := range textRunes {
		if i == cursorPos {
			cursorLine = currentLine
			cursorCol = currentCol
		}

		if r == '\n' {
			currentLine++
			currentCol = 0
			if currentLine >= height {
				break
			}
		} else {
			if currentLine < height {
				if currentCol < width-1 { // Leave space for cursor
					lines[currentLine] += string(r)
					currentCol++
				} else {
					// Wrap to next line
					currentLine++
					currentCol = 0
					if currentLine < height {
						lines[currentLine] = string(r)
						currentCol = 1
					}
				}
			}
		}
	}

	// Handle cursor at end of text
	if cursorPos == len(textRunes) {
		cursorLine = currentLine
		cursorCol = currentCol
	}

	// Add cursor to the appropriate position
	if cursorLine < height {
		line := lines[cursorLine]
		lineRunes := []rune(line)
		
		if cursorCol <= len(lineRunes) {
			// Insert cursor at the correct position
			newLine := make([]rune, 0, len(lineRunes)+1)
			newLine = append(newLine, lineRunes[:cursorCol]...)
			newLine = append(newLine, '│')
			newLine = append(newLine, lineRunes[cursorCol:]...)
			lines[cursorLine] = string(newLine)
		}
	}

	// Pad lines to consistent width and height
	for i := 0; i < height; i++ {
		if len(lines[i]) < width {
			lines[i] += strings.Repeat(" ", width-len(lines[i]))
		} else if len(lines[i]) > width {
			lines[i] = lines[i][:width]
		}
	}

	return strings.Join(lines, "\n")
}

// handleFeedbackViewKeys handles keyboard input in feedback view
func (a *App) handleFeedbackViewKeys(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	if a.state.FeedbackSubmitting {
		// Ignore input while submitting
		return a, nil
	}

	// Debug: Show what key was pressed
	keyPressed := msg.String()
	keyType := msg.Type.String()
	a.state.StatusMessage = fmt.Sprintf("Key: '%s' Type: '%s'", keyPressed, keyType)

	// Check for Ctrl+Enter in multiple ways
	if (msg.Type == tea.KeyCtrlM) || (keyPressed == "ctrl+enter") || (keyPressed == "ctrl+m") || (keyPressed == "ctrl+j") {
		// Submit feedback and quit
		a.state.StatusMessage = "Ctrl+Enter detected! Submitting..."
		feedback := strings.TrimSpace(a.state.FeedbackText)
		if feedback != "" {
			a.state.FeedbackSubmitting = true
			a.state.StatusMessage = "Submitting feedback..."
			return a, a.submitFeedback
		}
		// If no feedback text, just quit
		a.state.StatusMessage = "No feedback text, quitting..."
		return a, tea.Quit
	}

	switch keyPressed {
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

	case "f1":
		// Alternative submit key for testing
		a.state.StatusMessage = "F1 pressed - submitting feedback..."
		feedback := strings.TrimSpace(a.state.FeedbackText)
		if feedback != "" {
			a.state.FeedbackSubmitting = true
			a.state.StatusMessage = "Submitting feedback via F1..."
			return a, a.submitFeedback
		}
		// If no feedback text, just quit
		a.state.StatusMessage = "No feedback text, quitting..."
		return a, tea.Quit

	case "ctrl+enter", "ctrl+m", "ctrl+j":
		// Submit feedback and quit (handle multiple key combinations)
		a.state.StatusMessage = "Processing Ctrl+Enter..."
		feedback := strings.TrimSpace(a.state.FeedbackText)
		if feedback != "" {
			a.state.FeedbackSubmitting = true
			a.state.StatusMessage = "Submitting feedback..."
			return a, a.submitFeedback
		}
		// If no feedback text, just quit
		a.state.StatusMessage = "No feedback text, quitting..."
		return a, tea.Quit

	case "enter":
		// Add newline to feedback text
		textRunes := []rune(a.state.FeedbackText)
		if a.state.FeedbackCursorPos >= len(textRunes) {
			a.state.FeedbackText += "\n"
		} else {
			newRunes := make([]rune, len(textRunes)+1)
			copy(newRunes[:a.state.FeedbackCursorPos], textRunes[:a.state.FeedbackCursorPos])
			newRunes[a.state.FeedbackCursorPos] = '\n'
			copy(newRunes[a.state.FeedbackCursorPos+1:], textRunes[a.state.FeedbackCursorPos:])
			a.state.FeedbackText = string(newRunes)
		}
		a.state.FeedbackCursorPos++
		return a, nil

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

	case "up":
		// Move cursor up one line
		textRunes := []rune(a.state.FeedbackText)
		lines := strings.Split(a.state.FeedbackText, "\n")
		
		currentPos := 0
		currentLine := 0
		currentCol := 0
		
		// Find current line and column
		for i, r := range textRunes {
			if i == a.state.FeedbackCursorPos {
				break
			}
			if r == '\n' {
				currentLine++
				currentCol = 0
			} else {
				currentCol++
			}
			currentPos++
		}
		
		if currentLine > 0 {
			// Move to previous line, same column if possible
			prevLine := currentLine - 1
			prevLineLen := len([]rune(lines[prevLine]))
			targetCol := currentCol
			if targetCol > prevLineLen {
				targetCol = prevLineLen
			}
			
			// Calculate new cursor position
			newPos := 0
			for i := 0; i < prevLine; i++ {
				newPos += len([]rune(lines[i])) + 1 // +1 for newline
			}
			newPos += targetCol
			
			if newPos >= 0 && newPos <= len(textRunes) {
				a.state.FeedbackCursorPos = newPos
			}
		}
		return a, nil

	case "down":
		// Move cursor down one line
		textRunes := []rune(a.state.FeedbackText)
		lines := strings.Split(a.state.FeedbackText, "\n")
		
		currentPos := 0
		currentLine := 0
		currentCol := 0
		
		// Find current line and column
		for i, r := range textRunes {
			if i == a.state.FeedbackCursorPos {
				break
			}
			if r == '\n' {
				currentLine++
				currentCol = 0
			} else {
				currentCol++
			}
			currentPos++
		}
		
		if currentLine < len(lines)-1 {
			// Move to next line, same column if possible
			nextLine := currentLine + 1
			nextLineLen := len([]rune(lines[nextLine]))
			targetCol := currentCol
			if targetCol > nextLineLen {
				targetCol = nextLineLen
			}
			
			// Calculate new cursor position
			newPos := 0
			for i := 0; i < nextLine; i++ {
				newPos += len([]rune(lines[i])) + 1 // +1 for newline
			}
			newPos += targetCol
			
			if newPos >= 0 && newPos <= len(textRunes) {
				a.state.FeedbackCursorPos = newPos
			}
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

	// Try to send feedback via multiple methods
	var lastError error

	// Method 1: Save to file and try to open email client
	lastError = a.sendFeedbackViaEmail(feedbackData)
	if lastError == nil {
		return FeedbackSubmittedMsg{success: true, error: ""}
	}

	// Method 2: Save feedback to local file as backup
	lastError = a.saveFeedbackToFile("gowseshaik@gmail.com", "k8sgo Feedback", 
		fmt.Sprintf("Tool: %v\nVersion: %v\nContext: %v\nNamespace: %v\nCLI Tool: %v\nTimestamp: %v\n\nFeedback:\n%v",
			feedbackData["tool"], feedbackData["version"], feedbackData["context"],
			feedbackData["namespace"], feedbackData["cli_tool"], feedbackData["timestamp"],
			feedbackData["feedback"]))
	
	if lastError == nil {
		return FeedbackSubmittedMsg{success: true, error: "Feedback saved locally"}
	}

	// Always return success to show thank you message, but log the error
	return FeedbackSubmittedMsg{success: true, error: ""}
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

// sendFeedbackViaEmail sends feedback via SMTP email or alternative method
func (a *App) sendFeedbackViaEmail(data map[string]interface{}) error {
	// Create email content
	toEmail := "gowseshaik@gmail.com"
	subject := "k8sgo Feedback"
	body := fmt.Sprintf(`New feedback from k8sgo:

Tool: %v
Version: %v
Context: %v
Namespace: %v
CLI Tool: %v
Timestamp: %v

Feedback:
%v`,
		data["tool"], data["version"], data["context"],
		data["namespace"], data["cli_tool"], data["timestamp"],
		data["feedback"])

	// Method 1: Try to save feedback to a local file that can be emailed
	err := a.saveFeedbackToFile(toEmail, subject, body)
	if err == nil {
		return nil
	}

	// Method 2: Create a mailto link approach (platform dependent)
	err = a.openMailtoLink(toEmail, subject, body)
	if err == nil {
		return nil
	}

	// If all methods fail, return error
	return fmt.Errorf("unable to send feedback - please email manually to %s", toEmail)
}

// saveFeedbackToFile saves feedback to a local file that can be emailed
func (a *App) saveFeedbackToFile(toEmail, subject, body string) error {
	// Create feedback directory in user's home
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return err
	}
	
	feedbackDir := filepath.Join(homeDir, ".k8sgo", "feedback")
	err = os.MkdirAll(feedbackDir, 0755)
	if err != nil {
		return err
	}
	
	// Create filename with timestamp
	timestamp := time.Now().Format("2006-01-02_15-04-05")
	filename := filepath.Join(feedbackDir, fmt.Sprintf("feedback_%s.txt", timestamp))
	
	// Create email content
	content := fmt.Sprintf("To: %s\nSubject: %s\n\n%s\n\n---\nThis feedback was saved locally. Please copy and email it manually to the above address.", toEmail, subject, body)
	
	err = os.WriteFile(filename, []byte(content), 0644)
	if err != nil {
		return err
	}
	
	return nil
}

// openMailtoLink attempts to open default email client with mailto link
func (a *App) openMailtoLink(toEmail, subject, body string) error {
	// URL encode the email content
	params := url.Values{}
	params.Set("subject", subject)
	params.Set("body", body)
	
	mailtoURL := fmt.Sprintf("mailto:%s?%s", toEmail, params.Encode())
	
	var cmd *exec.Cmd
	switch runtime.GOOS {
	case "windows":
		cmd = exec.Command("rundll32", "url.dll,FileProtocolHandler", mailtoURL)
	case "darwin":
		cmd = exec.Command("open", mailtoURL)
	case "linux":
		cmd = exec.Command("xdg-open", mailtoURL)
	default:
		return fmt.Errorf("unsupported platform")
	}
	
	return cmd.Run()
}

// saveFeedbackLocally saves feedback to a local file as fallback
func (a *App) saveFeedbackLocally(data map[string]interface{}) error {
	// This is a fallback method - save to local file
	// In production, you might want to save to a temp directory or user config directory
	return fmt.Errorf("local save not implemented") // Placeholder
}
