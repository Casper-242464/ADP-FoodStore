package config

import (
	"database/sql"
	"fmt"
	"os"
	_ "github.com/lib/pq"
)

// Config holds configuration values for the application (e.g., DB credentials, server port).
type Config struct {
	DBHost        string
	DBPort        string
	DBUser        string
	DBPassword    string
	DBName        string
	ServerAddress string
}

// GetConfig reads configuration from environment variables or uses defaults.
func GetConfig() *Config {
	return &Config{
		DBHost:        getEnv("DB_HOST", "localhost"),
		DBPort:        getEnv("DB_PORT", "5432"),
		DBUser:        getEnv("DB_USER", "postgres"),
		DBPassword:    getEnv("DB_PASSWORD", "123456789"),
		DBName:        getEnv("DB_NAME", "foodstore"),
		ServerAddress: getEnv("SERVER_ADDR", ":8080"),  // HTTP listen address
	}
}

// ConnectDB opens a connection to the Postgres database using settings from Config.
func ConnectDB(cfg *Config) (*sql.DB, error) {
	// Construct DSN (Data Source Name) for PostgreSQL
	dsn := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		cfg.DBHost, cfg.DBPort, cfg.DBUser, cfg.DBPassword, cfg.DBName)
	db, err := sql.Open("postgres", dsn)
	if err != nil {
		return nil, err
	}
	// Verify the connection is live
	if err := db.Ping(); err != nil {
		return nil, err
	}
	return db, nil
}

// getEnv is a helper to read an environment variable or return a default value.
func getEnv(key string, defaultVal string) string {
	val := os.Getenv(key)
	if val == "" {
		return defaultVal
	}
	return val
}
