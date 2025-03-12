package config

import (
	"fmt"
	"os"
	"strconv"
	"time"

	"github.com/huberts90/restful-api/internal/storage"
)

// Config holds all application configuration
type Config struct {
	Server   ServerConfig
	Postgres storage.PostgresConfig
	IsProd   bool
}

// ServerConfig holds the HTTP server configuration
type ServerConfig struct {
	Port int
}

// LoadConfig loads configuration from environment variables
func LoadConfig() (*Config, error) {
	// Load server config
	port, err := loadIntEnv("SERVER_PORT", 8080)
	if err != nil {
		return nil, fmt.Errorf("invalid SERVER_PORT: %w", err)
	}

	// Load database config
	pgHost := loadEnv("POSTGRES_HOST", "localhost")
	pgPort, err := loadIntEnv("POSTGRES_PORT", 5432)
	if err != nil {
		return nil, fmt.Errorf("invalid POSTGRES_PORT: %w", err)
	}
	pgUser := loadEnv("POSTGRES_USER", "postgres")
	pgPassword := loadEnv("POSTGRES_PASSWORD", "postgres")
	pgDBName := loadEnv("POSTGRES_DB", "users_db")
	pgSSLMode := loadEnv("POSTGRES_SSLMODE", "disable")

	maxOpenConns, err := loadIntEnv("MAX_OPEN_CONNS", 25)
	if err != nil {
		return nil, fmt.Errorf("invalid MAX_OPEN_CONNS: %w", err)
	}
	maxIdleConns, err := loadIntEnv("MAX_IDLE_CONNS", 5)
	if err != nil {
		return nil, fmt.Errorf("invalid MAX_IDLE_CONNS: %w", err)
	}
	connMaxLifetime, err := loadTimeDurEnv("CONN_MAX_LIFETIME", 5*time.Minute)
	if err != nil {
		return nil, fmt.Errorf("invalid CONN_MAX_LIFETIME: %w", err)
	}
	connMaxIdletime, err := loadTimeDurEnv("CONN_MAX_IDLETIME", 5*time.Minute)
	if err != nil {
		return nil, fmt.Errorf("invalid CONN_MAX_IDLETIME: %w", err)
	}

	// Load environment mode
	isProd := loadEnv("ENV", "development") == "production"

	return &Config{
		Server: ServerConfig{
			Port: port,
		},
		Postgres: storage.PostgresConfig{
			Host:            pgHost,
			Port:            pgPort,
			User:            pgUser,
			Password:        pgPassword,
			DBName:          pgDBName,
			SSLMode:         pgSSLMode,
			MaxOpenConns:    maxOpenConns,
			MaxIdleConns:    maxIdleConns,
			ConnMaxLifetime: connMaxLifetime,
			ConnMaxIdleTime: connMaxIdletime,
		},
		IsProd: isProd,
	}, nil
}

// Helper to load string environment variables with defaults
func loadEnv(key, defaultValue string) string {
	val, exists := os.LookupEnv(key)
	if !exists {
		return defaultValue
	}
	return val
}

// Helper to load integer environment variables with defaults
func loadIntEnv(key string, defaultValue int) (int, error) {
	valStr := loadEnv(key, "")
	if valStr == "" {
		return defaultValue, nil
	}
	val, err := strconv.Atoi(valStr)
	if err != nil {
		return 0, err
	}
	return val, nil
}

func loadTimeDurEnv(key string, defaultValue time.Duration) (time.Duration, error) {
	valStr := loadEnv(key, "")
	if valStr == "" {
		return defaultValue, nil
	}
	val, err := time.ParseDuration(valStr)
	if err != nil {
		return 0, err
	}
	return val, nil
}
