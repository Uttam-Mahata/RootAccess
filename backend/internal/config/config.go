package config

import (
	"log"
	"os"
	"strconv"

	"github.com/joho/godotenv"
)

type Config struct {
	Port           string
	Environment    string // "development" or "production"
	MongoURI       string
	DBName         string
	JWTSecret      string
	FrontendURL    string
	TrustedProxies string
	SMTPHost       string
	SMTPPort       int
	SMTPUser       string
	SMTPPass       string
	SMTPFrom       string
	RedisAddr      string
	RedisPassword  string
	RedisDB        int
	GoogleClientID     string
	GoogleClientSecret string
	GoogleRedirectURL  string
	GitHubClientID     string
	GitHubClientSecret string
	GitHubRedirectURL  string
	DiscordClientID     string
	DiscordClientSecret string
	DiscordRedirectURL  string
	// Registration access control
	RegistrationMode           string // "open" | "domain" | "disabled"
	RegistrationAllowedDomains string // comma-separated, e.g. "college.edu,university.org"
}

func LoadConfig() *Config {
	err := godotenv.Load()
	if err != nil {
		log.Println("Warning: .env file not found, relying on environment variables")
	}

	smtpPort, _ := strconv.Atoi(getEnv("SMTP_PORT", "587"))
	redisDB, _ := strconv.Atoi(getEnv("REDIS_DB", "0"))

	return &Config{
		Port:           getEnv("PORT", "8080"),
		Environment:    getEnv("APP_ENV", "development"),
		MongoURI:       getEnv("MONGO_URI", "mongodb://localhost:27017"),
		DBName:         getEnv("DB_NAME", "go_ctf"),
		JWTSecret:      getEnv("JWT_SECRET", "default_secret"),
		FrontendURL:    getEnv("FRONTEND_URL", "http://localhost:4200"),
		TrustedProxies: getEnv("TRUSTED_PROXIES", ""),
		SMTPHost:       getEnv("SMTP_HOST", "smtp.gmail.com"),
		SMTPPort:     smtpPort,
		SMTPUser:     getEnv("SMTP_USER", ""),
		SMTPPass:     getEnv("SMTP_PASS", ""),
		SMTPFrom:     getEnv("SMTP_FROM", "noreply@rootaccess.live"),
		RedisAddr:    getEnv("REDIS_ADDR", "localhost:6379"),
		RedisPassword: getEnv("REDIS_PASSWORD", ""),
		RedisDB:      redisDB,
		GoogleClientID:     getEnv("GOOGLE_CLIENT_ID", ""),
		GoogleClientSecret: getEnv("GOOGLE_CLIENT_SECRET", ""),
		GoogleRedirectURL:  getEnv("GOOGLE_REDIRECT_URL", "https://rootaccess.live/auth/google/callback"),
		GitHubClientID:     getEnv("GITHUB_CLIENT_ID", ""),
		GitHubClientSecret: getEnv("GITHUB_CLIENT_SECRET", ""),
		GitHubRedirectURL:  getEnv("GITHUB_REDIRECT_URL", "https://rootaccess.live/auth/github/callback"),
		DiscordClientID:     getEnv("DISCORD_CLIENT_ID", ""),
		DiscordClientSecret: getEnv("DISCORD_CLIENT_SECRET", ""),
		DiscordRedirectURL:  getEnv("DISCORD_REDIRECT_URL", "https://rootaccess.live/auth/discord/callback"),
		RegistrationMode:           getEnv("REGISTRATION_MODE", "open"),
		RegistrationAllowedDomains: getEnv("REGISTRATION_ALLOWED_DOMAINS", ""),
	}
}

func getEnv(key, fallback string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return fallback
}
