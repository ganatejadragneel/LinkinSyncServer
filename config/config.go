package config

import (
	"fmt"
	"os"

	"github.com/joho/godotenv"
)

// Config holds all application configuration
type Config struct {
	Server   ServerConfig
	Database DatabaseConfig
	Spotify  SpotifyConfig
	Genius   GeniusConfig
	Ollama   OllamaConfig
}

// ServerConfig holds server configuration
type ServerConfig struct {
	Port string
}

// DatabaseConfig holds database configuration
type DatabaseConfig struct {
	Host     string
	Port     string
	User     string
	Password string
	DBName   string
	SSLMode  string
}

// SpotifyConfig holds Spotify API configuration
type SpotifyConfig struct {
	ClientID     string
	ClientSecret string
}

// GeniusConfig holds Genius API configuration
type GeniusConfig struct {
	AccessToken string
}

// OllamaConfig holds Ollama configuration
type OllamaConfig struct {
	BaseURL     string
	Model       string
	Temperature float64
	TopP        float64
	TopK        int
}

// Load loads configuration from environment variables
func Load() (*Config, error) {
	// Load .env file if it exists
	if err := godotenv.Load(); err != nil {
		// It's okay if .env doesn't exist in production
		fmt.Println("Warning: .env file not found, using system environment variables")
	}

	cfg := &Config{
		Server: ServerConfig{
			Port: getEnvWithDefault("PORT", "8080"),
		},
		Database: DatabaseConfig{
			Host:     getEnvWithDefault("DB_HOST", "localhost"),
			Port:     getEnvWithDefault("DB_PORT", "5432"),
			User:     getEnvRequired("DB_USER"),
			Password: getEnvRequired("DB_PASSWORD"),
			DBName:   getEnvRequired("DB_NAME"),
			SSLMode:  getEnvWithDefault("DB_SSL_MODE", "disable"),
		},
		Spotify: SpotifyConfig{
			ClientID:     getEnvRequired("SPOTIFY_CLIENT_ID"),
			ClientSecret: getEnvRequired("SPOTIFY_CLIENT_SECRET"),
		},
		Genius: GeniusConfig{
			AccessToken: getEnvRequired("GENIUS_ACCESS_TOKEN"),
		},
		Ollama: OllamaConfig{
			BaseURL:     getEnvWithDefault("OLLAMA_BASE_URL", "http://localhost:11434"),
			Model:       getEnvWithDefault("OLLAMA_MODEL", "llama3.2:3b"),
			Temperature: 0.7,
			TopP:        0.9,
			TopK:        40,
		},
	}

	return cfg, nil
}

// getEnvRequired gets a required environment variable
func getEnvRequired(key string) string {
	value := os.Getenv(key)
	if value == "" {
		panic(fmt.Sprintf("Required environment variable %s is not set", key))
	}
	return value
}

// getEnvWithDefault gets an environment variable with a default value
func getEnvWithDefault(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

// GetDatabaseURL returns the formatted database connection string
func (c *Config) GetDatabaseURL() string {
	return fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=%s",
		c.Database.User,
		c.Database.Password,
		c.Database.Host,
		c.Database.Port,
		c.Database.DBName,
		c.Database.SSLMode,
	)
}