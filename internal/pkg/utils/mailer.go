package utils

import (
	"crypto/rand"
	"crypto/tls"
	"fmt"
	"math/big"
	"net"
	"net/smtp"
	"time"

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

// SendEmail sends an HTML formatted email using the SMTP configurations with a 10-second timeout.
func SendEmail(to string, subject string, htmlBody string) error {
	cfg := config.AppConfig

	addr := fmt.Sprintf("%s:%d", cfg.SMTPHost, cfg.SMTPPort)
	dialer := net.Dialer{
		Timeout: 10 * time.Second,
	}

	var conn net.Conn
	var err error

	// 1. Dial connection with a timeout
	if cfg.SMTPPort == 465 {
		// Use TLS Dial for Port 465 (Implicit SSL/TLS)
		tlsConfig := &tls.Config{
			ServerName: cfg.SMTPHost,
		}
		conn, err = tls.DialWithDialer(&dialer, "tcp", addr, tlsConfig)
	} else {
		// Use standard TCP Dial for Port 587 / 25
		conn, err = dialer.Dial("tcp", addr)
	}

	if err != nil {
		return fmt.Errorf("failed to connect to SMTP server (timeout/connection error): %w", err)
	}
	defer conn.Close()

	// 2. Set read/write deadlines on the connection
	_ = conn.SetDeadline(time.Now().Add(10 * time.Second))

	// 3. Create SMTP client
	client, err := smtp.NewClient(conn, cfg.SMTPHost)
	if err != nil {
		return fmt.Errorf("failed to create SMTP client: %w", err)
	}
	defer client.Quit()

	// 4. Send STARTTLS if supported (necessary for port 587, skip if already on TLS port 465)
	if cfg.SMTPPort == 587 {
		tlsConfig := &tls.Config{
			ServerName: cfg.SMTPHost,
		}
		if err := client.StartTLS(tlsConfig); err != nil {
			return fmt.Errorf("failed to start TLS: %w", err)
		}
	}

	// 5. Authenticate if credentials are provided
	if cfg.SMTPUser != "" {
		auth := smtp.PlainAuth("", cfg.SMTPUser, cfg.SMTPPass, cfg.SMTPHost)
		if err := client.Auth(auth); err != nil {
			return fmt.Errorf("failed to authenticate SMTP: %w", err)
		}
	}

	// 6. Set sender and recipient
	if err := client.Mail(cfg.SMTPSender); err != nil {
		return fmt.Errorf("failed to set sender: %w", err)
	}
	if err := client.Rcpt(to); err != nil {
		return fmt.Errorf("failed to add recipient: %w", err)
	}

	// 7. Write email headers and HTML body
	w, err := client.Data()
	if err != nil {
		return fmt.Errorf("failed to get data writer: %w", err)
	}
	defer w.Close()

	mime := "MIME-version: 1.0;\r\nContent-Type: text/html; charset=\"UTF-8\";\r\n\r\n"
	msg := fmt.Sprintf("From: %s\r\nTo: %s\r\nSubject: %s\r\n%s%s", cfg.SMTPSender, to, subject, mime, htmlBody)

	if _, err := w.Write([]byte(msg)); err != nil {
		return fmt.Errorf("failed to write body: %w", err)
	}

	return nil
}

