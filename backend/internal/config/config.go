package config

import (
	"context"
	"encoding/json"
	"log"
	"os"
	"strconv"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/secretsmanager"
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
	// Upstash Redis URLs (rediss://default:token@host:6379)
	RedisURLAuth       string
	RedisURLScoreboard string
	RedisURLChallenge  string
	RedisURLWebSocket  string
	RedisURLAnalytics  string
	RedisURLGeneral    string
	GoogleClientID     string
	GoogleClientSecret string
	GoogleRedirectURL  string
	GitHubClientID     string
	GitHubClientSecret string
	GitHubRedirectURL  string
	DiscordClientID     string
	DiscordClientSecret string
	DiscordRedirectURL  string
	// WebSocket Callback URL for AWS Lambda
	WsCallbackURL string
	// CORS configuration
	CORSAllowedOrigins string // comma-separated list of allowed origins
	// Registration access control
	RegistrationMode           string // "open" | "domain" | "disabled"
	RegistrationAllowedDomains string // comma-separated, e.g. "college.edu,university.org"
}

func LoadConfig() *Config {
	err := godotenv.Load()
	if err != nil {
		log.Println("Warning: .env file not found, relying on environment variables")
	}

	// Try to load secrets from AWS Secrets Manager if secret name is provided
	awsSecretName := os.Getenv("AWS_SECRET_NAME")
	if awsSecretName != "" {
		log.Printf("Attempting to load secrets from AWS Secrets Manager: %s", awsSecretName)
		loadAWSSecrets(awsSecretName)
	}

	smtpPort, _ := strconv.Atoi(getEnv("SMTP_PORT", "587"))

	jwtSecret := getEnv("JWT_SECRET", "default_secret")
	environment := getEnv("APP_ENV", "development")

	if environment == "production" && jwtSecret == "default_secret" {
		log.Fatal("FATAL: JWT_SECRET must be set in production environment")
	}

	return &Config{
		Port:           getEnv("PORT", "8080"),
		Environment:    environment,
		MongoURI:       getEnv("MONGO_URI", "mongodb://localhost:27017"),
		DBName:         getEnv("DB_NAME", "go_ctf"),
		JWTSecret:      jwtSecret,
		FrontendURL:    getEnv("FRONTEND_URL", "http://localhost:4200"),
		TrustedProxies: getEnv("TRUSTED_PROXIES", ""),
		SMTPHost:       getEnv("SMTP_HOST", "smtp.gmail.com"),
		SMTPPort:     smtpPort,
		SMTPUser:     getEnv("SMTP_USER", ""),
		SMTPPass:     getEnv("SMTP_PASS", ""),
		SMTPFrom:     getEnv("SMTP_FROM", "noreply@rootaccess.live"),
		RedisURLAuth:       getEnv("REDIS_URL_AUTH", ""),
		RedisURLScoreboard: getEnv("REDIS_URL_SCOREBOARD", ""),
		RedisURLChallenge:  getEnv("REDIS_URL_CHALLENGE", ""),
		RedisURLWebSocket:  getEnv("REDIS_URL_WEBSOCKET", ""),
		RedisURLAnalytics:  getEnv("REDIS_URL_ANALYTICS", ""),
		RedisURLGeneral:    getEnv("REDIS_URL_GENERAL", ""),
		GoogleClientID:     getEnv("GOOGLE_CLIENT_ID", ""),
		GoogleClientSecret: getEnv("GOOGLE_CLIENT_SECRET", ""),
		GoogleRedirectURL:  getEnv("GOOGLE_REDIRECT_URL", "https://rootaccess.live/auth/google/callback"),
		GitHubClientID:     getEnv("GITHUB_CLIENT_ID", ""),
		GitHubClientSecret: getEnv("GITHUB_CLIENT_SECRET", ""),
		GitHubRedirectURL:  getEnv("GITHUB_REDIRECT_URL", "https://rootaccess.live/auth/github/callback"),
		DiscordClientID:     getEnv("DISCORD_CLIENT_ID", ""),
		DiscordClientSecret: getEnv("DISCORD_CLIENT_SECRET", ""),
		DiscordRedirectURL:  getEnv("DISCORD_REDIRECT_URL", "https://rootaccess.live/auth/discord/callback"),
		WsCallbackURL:       getEnv("WS_CALLBACK_URL", ""),
		CORSAllowedOrigins:  getEnv("CORS_ALLOWED_ORIGINS", ""),
		RegistrationMode:           getEnv("REGISTRATION_MODE", "open"),
		RegistrationAllowedDomains: getEnv("REGISTRATION_ALLOWED_DOMAINS", ""),
	}
}

func loadAWSSecrets(secretName string) {
	region := os.Getenv("AWS_REGION")
	if region == "" {
		region = "us-east-1"
	}

	cfg, err := config.LoadDefaultConfig(context.TODO(), config.WithRegion(region))
	if err != nil {
		log.Printf("Error loading AWS config: %v", err)
		return
	}

	svc := secretsmanager.NewFromConfig(cfg)
	input := &secretsmanager.GetSecretValueInput{
		SecretId:     aws.String(secretName),
		VersionStage: aws.String("AWSCURRENT"),
	}

	result, err := svc.GetSecretValue(context.TODO(), input)
	if err != nil {
		log.Printf("Error getting secret value: %v", err)
		return
	}

	var secrets map[string]string
	err = json.Unmarshal([]byte(*result.SecretString), &secrets)
	if err != nil {
		log.Printf("Error unmarshaling secret: %v", err)
		return
	}

	for key, value := range secrets {
		os.Setenv(key, value)
	}
	log.Println("Successfully loaded secrets from AWS Secrets Manager")
}

func getEnv(key, fallback string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return fallback
}
