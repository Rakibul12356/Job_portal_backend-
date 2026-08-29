package utils

import (
	"crypto/rand"
	"fmt"
	"math/big"
	"net/smtp"

	"github.com/rakib/job-portal-api/internal/config"
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

// SendEmail sends an HTML formatted email using the SMTP configurations loaded in config.go.
func SendEmail(to string, subject string, htmlBody string) error {
	cfg := config.AppConfig

	// Create authentication for SMTP
	auth := smtp.PlainAuth("", cfg.SMTPUser, cfg.SMTPPass, cfg.SMTPHost)

	// Set headers for HTML email
	mime := "MIME-version: 1.0;\nContent-Type: text/html; charset=\"UTF-8\";\n\n"
	msg := []byte("From: " + cfg.SMTPSender + "\r\n" +
		"To: " + to + "\r\n" +
		"Subject: " + subject + "\r\n" +
		mime +
		htmlBody)

	addr := fmt.Sprintf("%s:%d", cfg.SMTPHost, cfg.SMTPPort)

	// SendMail handles establishing the connection, switching to TLS if STARTTLS is supported, and sending the mail.
	err := smtp.SendMail(addr, auth, cfg.SMTPSender, []string{to}, msg)
	if err != nil {
		return fmt.Errorf("failed to send email: %w", err)
	}

	return nil
}
