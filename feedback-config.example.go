package main

// Feedback Configuration Example
// Copy this file to pkg/ui/feedback_config.go and update with your actual values

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

// Usage Instructions:
// 1. Set up a webhook endpoint to receive feedback (recommended)
// 2. Or configure SMTP settings for email delivery
// 3. Update the URLs/credentials above
// 4. Rename this file to feedback_config.go in pkg/ui/
// 5. Rebuild the application

// Example webhook handler (Node.js/Express):
/*
app.post('/k8sgo-feedback', (req, res) => {
    const feedback = req.body;
    console.log('Received feedback:', feedback);

    // Save to database, send email, etc.
    // Your logic here

    res.status(200).json({ success: true });
});
*/

// Example Gmail App Password setup:
// 1. Enable 2FA on your Google account
// 2. Go to Google Account settings
// 3. Security > App passwords
// 4. Generate app password for "Mail"
// 5. Use that password in EMAIL_PASSWORD
