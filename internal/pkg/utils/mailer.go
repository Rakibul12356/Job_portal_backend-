package utils

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"math/big"

	"github.com/rakib/job-portal-api/internal/config"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/gmail/v1"
	"google.golang.org/api/option"
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

// SendEmail sends an HTML formatted email using the Google Gmail REST API (OAuth 2.0).
// This bypasses hosting provider port blocks (like Render's SMTP block) by sending requests over HTTPS (port 443).
func SendEmail(to string, subject string, htmlBody string) error {
	cfg := config.AppConfig

	// Configure OAuth2 client config
	oauthConfig := &oauth2.Config{
		ClientID:     cfg.GmailClientID,
		ClientSecret: cfg.GmailClientSecret,
		Endpoint:     google.Endpoint,
	}

	// Create oauth2 token utilizing the RefreshToken
	token := &oauth2.Token{
		RefreshToken: cfg.GmailRefreshToken,
	}

	ctx := context.Background()

	// Get a client that automatically refreshes the token using the refresh token
	httpClient := oauthConfig.Client(ctx, token)

	srv, err := gmail.NewService(ctx, option.WithHTTPClient(httpClient))
	if err != nil {
		return fmt.Errorf("unable to retrieve Gmail client: %v", err)
	}

	// Assemble MIME email
	from := cfg.GmailSenderEmail

	header := make(map[string]string)
	header["From"] = from
	header["To"] = to
	// UTF-8 base64 encoded subject is required for non-ASCII/special characters
	header["Subject"] = fmt.Sprintf("=?UTF-8?B?%s?=", base64.StdEncoding.EncodeToString([]byte(subject)))
	header["MIME-Version"] = "1.0"
	header["Content-Type"] = "text/html; charset=UTF-8"
	header["Content-Transfer-Encoding"] = "base64"

	var messageBody string
	for k, v := range header {
		messageBody += fmt.Sprintf("%s: %s\r\n", k, v)
	}
	messageBody += "\r\n" + base64.StdEncoding.EncodeToString([]byte(htmlBody))

	// Base64 URL safe encoding (without padding or with padding are both fine for Gmail API)
	rawEncoded := base64.URLEncoding.EncodeToString([]byte(messageBody))

	msg := &gmail.Message{
		Raw: rawEncoded,
	}

	_, err = srv.Users.Messages.Send("me", msg).Do()
	if err != nil {
		return fmt.Errorf("gmail Send API call failed: %w", err)
	}

	return nil
}
