package utils

import (
	"crypto/rand"
	"fmt"
	"math/big"
	"os"

	"github.com/rakib/job-portal-api/internal/config"
	"github.com/resend/resend-go/v2"
)

// GenerateOTP generates a cryptographically secure 6-digit numeric OTP.
func GenerateOTP() (string, error) {
	max := big.NewInt(900000)
	n, err := rand.Int(rand.Reader, max)
	if err != nil {
		return "", err
	}
	otp := n.Int64() + 100000
	return fmt.Sprintf("%d", otp), nil
}

// SendEmail sends an HTML formatted email using the Resend HTTP API.
// This bypasses hosting provider port blocks (like Render's SMTP block) by sending requests over HTTPS (port 443).
func SendEmail(to string, subject string, htmlBody string) error {
	cfg := config.AppConfig

	// Support both SMTP_PASS and RESEND_API_KEY environment variables
	apiKey := cfg.SMTPPass
	if apiKey == "" {
		apiKey = os.Getenv("RESEND_API_KEY")
	}

	if apiKey == "" {
		return fmt.Errorf("resend API key is empty (check SMTP_PASS or RESEND_API_KEY in configuration)")
	}

	client := resend.NewClient(apiKey)

	params := &resend.SendEmailRequest{
		From:    cfg.SMTPSender, // e.g., "onboarding@resend.dev"
		To:      []string{to},
		Subject: subject,
		Html:    htmlBody,
	}

	_, err := client.Emails.Send(params)
	if err != nil {
		return fmt.Errorf("resend API call failed: %w", err)
	}

	return nil
}



