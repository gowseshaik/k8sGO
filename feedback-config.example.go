package main

const (
	// Webhook Configuration (Recommended)
	FEEDBACK_WEBHOOK_URL = "https://your-api.example.com/k8sgo-feedback"

	// Email Configuration (Fallback)
	SMTP_HOST      = "smtp.gmail.com"
	SMTP_PORT      = "587"
	FROM_EMAIL     = "k8sgo-feedback@your-domain.com"
	TO_EMAIL       = "your-email@your-domain.com"
	EMAIL_PASSWORD = "your-app-specific-password" // Use app-specific password for Gmail

	// Local fallback file (last resort)
	FEEDBACK_FILE_PATH = "/tmp/k8sgo-feedback.log"
)


