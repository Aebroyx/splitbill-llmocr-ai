package config

import (
	"fmt"
	"net/url"
	"os"
	"strings"
	"time"

	"github.com/joho/godotenv"
)

type Config struct {
	// Server config
	Environment string
	ServerPort  string
	ServerHost  string

	// Database config
	DBHost      string
	DBPort      string
	DBUser      string
	DBPassword  string
	DBName      string
	DBSSLMode   string
	DatabaseURL string

	// JWT config
	JWTSecret string
	JWTExpiry time.Duration

	// CORS config
	CORSAllowedOrigins string

	// Logging
	LogLevel string
}

// Load loads the configuration from environment variables
func Load() (*Config, error) {
	// Load .env file
	if err := godotenv.Load(); err != nil {
		return nil, fmt.Errorf("error loading .env file: %v", err)
	}

	// Parse JWT expiry duration
	jwtExpiry, err := time.ParseDuration(getEnv("JWT_EXPIRY", "24h"))
	if err != nil {
		return nil, fmt.Errorf("invalid JWT_EXPIRY format: %v", err)
	}

	environment := getEnv("APP_ENV", "development")

	// For production, prioritize DATABASE_URL
	var dbHost, dbPort, dbUser, dbPassword, dbName, dbSSLMode string
	databaseURL := getEnv("DATABASE_URL", "")

	if environment == "production" && databaseURL != "" {
		// Parse DATABASE_URL for production
		// Handle both standard PostgreSQL and Supabase connection strings
		if strings.HasPrefix(databaseURL, "postgresql://") || strings.HasPrefix(databaseURL, "postgres://") {
			// Parse as URL
			parsedURL, err := url.Parse(databaseURL)
			if err != nil {
				return nil, fmt.Errorf("invalid DATABASE_URL format: %v", err)
			}

			// Extract database connection details from URL
			dbHost = parsedURL.Hostname()
			if dbHost == "" {
				return nil, fmt.Errorf("invalid DATABASE_URL: missing hostname")
			}

			if parsedURL.Port() != "" {
				dbPort = parsedURL.Port()
			} else {
				dbPort = "5432" // Default PostgreSQL port
			}

			dbUser = parsedURL.User.Username()
			if dbUser == "" {
				return nil, fmt.Errorf("invalid DATABASE_URL: missing username")
			}

			if password, ok := parsedURL.User.Password(); ok {
				dbPassword = password
			} else {
				return nil, fmt.Errorf("invalid DATABASE_URL: missing password")
			}

			dbName = strings.TrimPrefix(parsedURL.Path, "/")
			if dbName == "" {
				return nil, fmt.Errorf("invalid DATABASE_URL: missing database name")
			}

			dbSSLMode = "require" // Supabase requires SSL
		} else {
			return nil, fmt.Errorf("invalid DATABASE_URL: must start with 'postgresql://' or 'postgres://'")
		}
	} else {
		// Use individual database parameters for development
		dbHost = getEnv("DB_HOST", "localhost")
		dbPort = getEnv("DB_PORT", "5432")
		dbUser = getEnv("DB_USER", "postgres")
		dbPassword = getEnv("DB_PASSWORD", "")
		dbName = getEnv("DB_NAME", "splitbill-llmocr.app")
		dbSSLMode = getEnv("DB_SSL_MODE", "disable")
	}

	return &Config{
		// Server config
		Environment: environment,
		ServerPort:  getEnv("SERVER_PORT", "8080"),
		ServerHost:  getEnv("SERVER_HOST", "localhost"),

		// Database config
		DBHost:      dbHost,
		DBPort:      dbPort,
		DBUser:      dbUser,
		DBPassword:  dbPassword,
		DBName:      dbName,
		DBSSLMode:   dbSSLMode,
		DatabaseURL: databaseURL,

		// JWT config
		JWTSecret: getEnv("JWT_SECRET", ""),
		JWTExpiry: jwtExpiry,

		// CORS config
		CORSAllowedOrigins: getEnv("CORS_ALLOWED_ORIGINS", "http://localhost:3000"),

		// Logging
		LogLevel: getEnv("LOG_LEVEL", "debug"),
	}, nil
}

// getEnv gets an environment variable or returns a default value
func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

// Validate checks if the configuration is valid
func (c *Config) Validate() error {
	if c.JWTSecret == "" {
		return fmt.Errorf("JWT_SECRET is required")
	}

	// For production, DATABASE_URL is required and must be valid
	if c.Environment == "production" {
		if c.DatabaseURL == "" {
			return fmt.Errorf("DATABASE_URL is required for production environment")
		}

		// Validate DATABASE_URL format
		if err := c.validateDatabaseURL(); err != nil {
			return fmt.Errorf("invalid DATABASE_URL: %v", err)
		}
	} else {
		// For development, individual database parameters are required
		if c.DBPassword == "" {
			return fmt.Errorf("DB_PASSWORD is required for development environment")
		}
	}

	return nil
}

// validateDatabaseURL validates the DATABASE_URL format
func (c *Config) validateDatabaseURL() error {
	if !strings.HasPrefix(c.DatabaseURL, "postgresql://") && !strings.HasPrefix(c.DatabaseURL, "postgres://") {
		return fmt.Errorf("must start with 'postgresql://' or 'postgres://'")
	}

	parsedURL, err := url.Parse(c.DatabaseURL)
	if err != nil {
		return fmt.Errorf("invalid URL format: %v", err)
	}

	if parsedURL.Hostname() == "" {
		return fmt.Errorf("missing hostname")
	}

	if parsedURL.User == nil || parsedURL.User.Username() == "" {
		return fmt.Errorf("missing username")
	}

	if parsedURL.Path == "" || strings.TrimPrefix(parsedURL.Path, "/") == "" {
		return fmt.Errorf("missing database name")
	}

	return nil
}

// GetDSN returns the database connection string
func (c *Config) GetDSN() string {
	// For production, use DATABASE_URL directly
	if c.Environment == "production" && c.DatabaseURL != "" {
		return c.DatabaseURL
	}

	// For development, build DSN from individual parameters
	return fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
		c.DBHost, c.DBPort, c.DBUser, c.DBPassword, c.DBName, c.DBSSLMode,
	)
}

// GetServerAddr returns the server address
func (c *Config) GetServerAddr() string {
	return fmt.Sprintf("%s:%s", c.ServerHost, c.ServerPort)
}
