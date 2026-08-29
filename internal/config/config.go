package config

import (
	"bufio"
	"os"
	"strconv"
	"strings"
	"time"
)

type Config struct {
	AppEnv            string
	AppPort           string
	AppBaseURL        string
	MongoURI          string
	MongoDB           string
	JWTAccessSecret   string
	JWTRefreshSecret  string
	JWTAccessTTL      time.Duration
	JWTRefreshTTL     time.Duration
	CORSOrigins       []string
	UploadDriver      string
	UploadDir         string
	RateLimitRPM      int
	SMTPHost          string
	SMTPPort          int
	SMTPUser          string
	SMTPPass          string
	SMTPSender        string
	GmailClientID     string
	GmailClientSecret string
	GmailRefreshToken string
	GmailSenderEmail  string
}

var AppConfig *Config

func LoadConfig() *Config {
	// Try loading .env file manually
	loadEnvFile(".env")

	accessTTL, err := time.ParseDuration(getEnv("JWT_ACCESS_TTL", "60m"))
	if err != nil {
		accessTTL = 60 * time.Minute
	}
	refreshTTL, err := time.ParseDuration(getEnv("JWT_REFRESH_TTL", "168h"))
	if err != nil {
		refreshTTL = 7 * 24 * time.Hour
	}

	corsOriginsStr := getEnv("CORS_ORIGINS", "http://localhost:5173,http://localhost:5174")
	corsOrigins := strings.Split(corsOriginsStr, ",")
	for i, origin := range corsOrigins {
		corsOrigins[i] = strings.TrimSpace(origin)
	}

	rpm, err := strconv.Atoi(getEnv("RATE_LIMIT_RPM", "60"))
	if err != nil {
		rpm = 60
	}

	smtpPort, err := strconv.Atoi(getEnv("SMTP_PORT", "587"))
	if err != nil {
		smtpPort = 587
	}

	AppConfig = &Config{
		AppEnv:            getEnv("APP_ENV", "development"),
		AppPort:           getEnv("PORT", getEnv("APP_PORT", "8080")),
		AppBaseURL:        getEnv("APP_BASE_URL", "http://localhost:8080"),
		MongoURI:          getEnv("MONGO_URI", "mongodb+srv://Job_portal_db:rakib74@cluster0.j9djoaf.mongodb.net/?appName=Cluster0"),
		MongoDB:           getEnv("MONGO_DB", "job_portal"),
		JWTAccessSecret:   getEnv("JWT_ACCESS_SECRET", "change-me-access-super-secret"),
		JWTRefreshSecret:  getEnv("JWT_REFRESH_SECRET", "change-me-refresh-super-secret"),
		JWTAccessTTL:      accessTTL,
		JWTRefreshTTL:     refreshTTL,
		CORSOrigins:       corsOrigins,
		UploadDriver:      getEnv("UPLOAD_DRIVER", "local"),
		UploadDir:         getEnv("UPLOAD_DIR", "./uploads"),
		RateLimitRPM:      rpm,
		SMTPHost:          getEnv("SMTP_HOST", "smtp.resend.com"),
		SMTPPort:          smtpPort,
		SMTPUser:          getEnv("SMTP_USER", "resend"),
		SMTPPass:          getEnv("SMTP_PASS", ""),
		SMTPSender:        getEnv("SMTP_SENDER", "onboarding@resend.dev"),
		GmailClientID:     getEnv("GMAIL_CLIENT_ID", ""),
		GmailClientSecret: getEnv("GMAIL_CLIENT_SECRET", ""),
		GmailRefreshToken: getEnv("GMAIL_REFRESH_TOKEN", ""),
		GmailSenderEmail:  getEnv("GMAIL_SENDER_EMAIL", ""),
	}

	return AppConfig
}

func getEnv(key, defaultVal string) string {
	if val, ok := os.LookupEnv(key); ok {
		val = strings.TrimSpace(val)
		if len(val) >= 2 && ((val[0] == '"' && val[len(val)-1] == '"') || (val[0] == '\'' && val[len(val)-1] == '\'')) {
			val = val[1 : len(val)-1]
		}
		return val
	}
	return defaultVal
}

func loadEnvFile(filename string) {
	file, err := os.Open(filename)
	if err != nil {
		return // Ignore error, environment variables might be set directly
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		parts := strings.SplitN(line, "=", 2)
		if len(parts) == 2 {
			key := strings.TrimSpace(parts[0])
			val := strings.TrimSpace(parts[1])
			// Strip quotes if present
			if len(val) >= 2 && ((val[0] == '"' && val[len(val)-1] == '"') || (val[0] == '\'' && val[len(val)-1] == '\'')) {
				val = val[1 : len(val)-1]
			}
			if os.Getenv(key) == "" {
				os.Setenv(key, val)
			}
		}
	}
}
