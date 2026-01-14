package config

import (
	"log"
	"os"
	"strconv"

	"github.com/joho/godotenv"
)

type Config struct {
	Port         string
	MongoURI     string
	DBName       string
	JWTSecret    string
	FrontendURL  string
	SMTPHost     string
	SMTPPort     int
	SMTPUser     string
	SMTPPass     string
	SMTPFrom     string
	RedisAddr    string
	RedisPassword string
	RedisDB      int
}

func LoadConfig() *Config {
	err := godotenv.Load()
	if err != nil {
		log.Println("Warning: .env file not found, relying on environment variables")
	}

	smtpPort, _ := strconv.Atoi(getEnv("SMTP_PORT", "587"))
	redisDB, _ := strconv.Atoi(getEnv("REDIS_DB", "0"))

	return &Config{
		Port:         getEnv("PORT", "8080"),
		MongoURI:     getEnv("MONGO_URI", "mongodb://localhost:27017"),
		DBName:       getEnv("DB_NAME", "go_ctf"),
		JWTSecret:    getEnv("JWT_SECRET", "default_secret"),
		FrontendURL:  getEnv("FRONTEND_URL", "http://localhost:4200"),
		SMTPHost:     getEnv("SMTP_HOST", "smtp.gmail.com"),
		SMTPPort:     smtpPort,
		SMTPUser:     getEnv("SMTP_USER", ""),
		SMTPPass:     getEnv("SMTP_PASS", ""),
		SMTPFrom:     getEnv("SMTP_FROM", "noreply@rootaccess.ctf"),
		RedisAddr:    getEnv("REDIS_ADDR", "localhost:6379"),
		RedisPassword: getEnv("REDIS_PASSWORD", ""),
		RedisDB:      redisDB,
	}
}

func getEnv(key, fallback string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return fallback
}
