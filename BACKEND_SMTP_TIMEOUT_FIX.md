# Backend Guide: Fixing SMTP Port 587 i/o timeout on Render / Production

This guide addresses the specific SMTP connection error:
`failed to connect to SMTP server (timeout/connection error): dial tcp 54.157.71.137:587: i/o timeout`

---

## 1. Why Did This Error Occur?

1. **Hosting Provider Restrictions:** Most cloud hosting providers (including Render, Heroku, AWS, and DigitalOcean) block outbound traffic on SMTP Port **587** (TLS) and Port **25** by default to prevent spam relaying from their servers.
2. **Connection Hanging/Timeout:** Because the hosting network silently drops or blocks connections to port 587, the Go backend connection handshake times out (`i/o timeout` error) after your 10-second limit.

---

## 2. Recommended Solution: Switch to SSL Port 465

Instead of using port 587, use **SMTP Port 465** with SSL/TLS enabled. Hosting platforms generally allow outbound traffic on port 465.

### Environment Variable Updates on Render / Production

Update your backend Environment Variables to the following:

| Variable | Value | Description |
| :--- | :--- | :--- |
| `SMTP_HOST` | `smtp.gmail.com` | SMTP Server Hostname (Gmail or equivalent) |
| `SMTP_PORT` | `465` | Port 465 is commonly allowed on Render |
| `SMTP_SECURE` | `true` | Tells the mailer to use SSL/TLS directly |
| `SMTP_USER` | `your-email@gmail.com` | The email address sending the OTP |
| `SMTP_PASS` | `xxxx xxxx xxxx xxxx` | **Google App Password** (not your login password) |

---

## 3. Go SMTP Transporter Code Example

Ensure your Go SMTP email-sending code is configured to establish a secure SSL/TLS connection directly when using port 465.

Here is a secure implementation example using standard Go `crypto/tls` and `net/smtp`:

```go
package utils

import (
	"crypto/tls"
	"fmt"
	"net/smtp"
	"os"
)

// SendOTPEmail sends a 6-digit verification code to the target email
func SendOTPEmail(targetEmail string, otp string) error {
	smtpHost := os.Getenv("SMTP_HOST") // e.g., "smtp.gmail.com"
	smtpPort := os.Getenv("SMTP_PORT") // e.g., "465"
	smtpUser := os.Getenv("SMTP_USER") // e.g., "careers@yourcompany.com"
	smtpPass := os.Getenv("SMTP_PASS") // Gmail App Password
	
	// Create headers and HTML email template
	subject := "Subject: Password Reset OTP Code\n"
	mime := "MIME-version: 1.0;\nContent-Type: text/html; charset=\"UTF-8\";\n\n"
	body := fmt.Sprintf(`
		<div style="font-family: Arial, sans-serif; padding: 20px; border: 1px solid #eee; border-radius: 8px; max-width: 600px; margin: 0 auto;">
			<h2 style="color: #0f172a; text-align: center;">Reset Your Password</h2>
			<p>Hello,</p>
			<p>You requested a password reset for your account on Online Job Portal. Use the following OTP code to proceed:</p>
			<div style="background-color: #f1f5f9; text-align: center; padding: 15px; margin: 20px 0; border-radius: 8px;">
				<span style="font-size: 28px; font-weight: bold; letter-spacing: 5px; color: #0f172a;">%s</span>
			</div>
			<p style="color: #ef4444; font-size: 14px;"><strong>Note:</strong> This OTP is valid for 10 minutes only. Do not share this code with anyone.</p>
			<p>If you did not request this, please ignore this email.</p>
			<hr style="border: 0; border-top: 1px solid #eee; margin: 20px 0;" />
			<p style="font-size: 12px; color: #64748b; text-align: center;">Online Job Portal &copy; 2026. All rights reserved.</p>
		</div>
	`, otp)

	msg := []byte(subject + mime + body)
	addr := fmt.Sprintf("%s:%s", smtpHost, smtpPort)

	// Set up TLS Configuration (Required for Port 465 SSL)
	tlsConfig := &tls.Config{
		InsecureSkipVerify: false,
		ServerName:         smtpHost,
	}

	// Connect to the SMTP Server
	conn, err := tls.Dial("tcp", addr, tlsConfig)
	if err != nil {
		return fmt.Errorf("failed to dial SMTP server over TLS: %w", err)
	}
	defer conn.Close()

	client, err := smtp.NewClient(conn, smtpHost)
	if err != nil {
		return fmt.Errorf("failed to create SMTP client: %w", err)
	}
	defer client.Quit()

	// Authenticate
	auth := smtp.PlainAuth("", smtpUser, smtpPass, smtpHost)
	if err = client.Auth(auth); err != nil {
		return fmt.Errorf("SMTP authentication failed: %w", err)
	}

	// Set Sender and Recipient
	if err = client.Mail(smtpUser); err != nil {
		return fmt.Errorf("failed to set sender: %w", err)
	}
	if err = client.Rcpt(targetEmail); err != nil {
		return fmt.Errorf("failed to set recipient: %w", err)
	}

	// Write Email Body Data
	w, err := client.Data()
	if err != nil {
		return fmt.Errorf("failed to open SMTP data writer: %w", err)
	}
	defer w.Close()

	_, err = w.Write(msg)
	if err != nil {
		return fmt.Errorf("failed to write email body: %w", err)
	}

	return nil
}
```

---

## 4. How to Verify If the Backend Fix Works

1. **Verify Local SMTP Connection:** Run the Go server locally. Ensure it successfully dials the SMTP host on port 465 (or port 587 if your local network does not block outbound port 587).
2. **Check Render Logs:** Deploy the backend changes to Render and view the server logs console. 
3. **Trigger Request on Frontend:** Click the **"Send OTP"** button on the frontend. If port 465 connects correctly, you will instantly receive the success toast notification and transition to Step 2.
