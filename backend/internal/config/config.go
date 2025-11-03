package config

import (
	"fmt"
	"os"
	"strconv"

	"smart-company-discovery/internal/models"
)

// LoadConfig loads configuration from environment variables
func LoadConfig() (*models.Config, error) {
	config := &models.Config{
		Server: models.ServerConfig{
			Port:        getEnvAsInt("SERVER_PORT", 8080),
			Host:        getEnv("SERVER_HOST", "0.0.0.0"),
			Environment: getEnv("SERVER_ENVIRONMENT", "development"),
		},
		Database: models.DatabaseConfig{
			Host:         getEnv("DB_HOST", "localhost"),
			Port:         getEnvAsInt("DB_PORT", 5432),
			User:         getEnv("DB_USER", "postgres"),
			Password:     getEnv("DB_PASSWORD", "postgres"),
			DBName:       getEnv("DB_NAME", "smart_discovery"),
			SSLMode:      getEnv("DB_SSLMODE", "disable"),
			MaxOpenConns: getEnvAsInt("DB_MAX_OPEN_CONNS", 25),
			MaxIdleConns: getEnvAsInt("DB_MAX_IDLE_CONNS", 5),
		},
		Pinecone: models.PineconeConfig{
			APIKey:      getEnv("PINECONE_API_KEY", ""),
			Environment: getEnv("PINECONE_ENVIRONMENT", ""),
			IndexName:   getEnv("PINECONE_INDEX_NAME", ""),
			Namespace:   getEnv("PINECONE_NAMESPACE", ""),
			Host:        getEnv("PINECONE_HOST", ""), // For Pinecone Local
		},
		GoogleEmbedding: models.GoogleEmbeddingConfig{
			APIKey:    getEnv("GOOGLE_API_KEY", ""),
			ProjectID: getEnv("GOOGLE_PROJECT_ID", ""),
			Location:  getEnv("GOOGLE_LOCATION", "us-central1"),
			Model:     getEnv("GOOGLE_EMBEDDING_MODEL", "text-embedding-004"),
		},
	}

	if err := validateConfig(config); err != nil {
		return nil, fmt.Errorf("configuration validation failed: %w", err)
	}

	return config, nil
}

func getEnv(key, defaultValue string) string {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}
	return value
}

func getEnvAsInt(key string, defaultValue int) int {
	valueStr := os.Getenv(key)
	if valueStr == "" {
		return defaultValue
	}
	value, err := strconv.Atoi(valueStr)
	if err != nil {
		return defaultValue
	}
	return value
}

func validateConfig(config *models.Config) error {
	if config.Server.Port < 1 || config.Server.Port > 65535 {
		return fmt.Errorf("invalid server port: %d", config.Server.Port)
	}

	if config.Database.Host == "" {
		return fmt.Errorf("database host is required")
	}

	return nil
}
