package config

import (
	"fmt"
	"os"

	"github.com/joho/godotenv"
)

// Config holds the configuration values for the application.
type Config struct {
	ServerPort string
	DBHost     string
	DBPort     string
	DBPath     string
	DBUser     string
	DBPassword string
}

// LoadConfig loads the configuration from environment variables and .env file.
func LoadConfig() (*Config, error) {
	// Attempt to load .env file. It's okay if it fails (e.g., in Docker/production environment)
	// but we should log it or handle it.
	_ = godotenv.Load()

	cfg := &Config{
		ServerPort: getEnv("SERVER_PORT", "8080"),
		DBHost:     getEnv("DB_HOST", "127.0.0.1"),
		DBPort:     getEnv("DB_PORT", "3051"),
		DBPath:     getEnv("DB_PATH", ""),
		DBUser:     getEnv("DB_USER", "pos"),
		DBPassword: getEnv("DB_PASSWORD", "pos"),
	}

	if cfg.DBPath == "" {
		return nil, fmt.Errorf("DB_PATH environment variable is required")
	}

	return cfg, nil
}

// getEnv gets an environment variable or returns a default value.
func getEnv(key, defaultValue string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return defaultValue
}
