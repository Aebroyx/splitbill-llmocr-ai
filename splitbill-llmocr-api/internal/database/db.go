package database

import (
	"fmt"
	"log"
	"time"

	"github.com/Aebroyx/splitbill-llmocr-api/internal/config"
	"github.com/Aebroyx/splitbill-llmocr-api/internal/domain/models"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

type DB struct {
	*gorm.DB
}

func NewConnection(cfg *config.Config) (*DB, error) {
	// Configure GORM logger based on environment
	var gormLogger logger.Interface
	if cfg.Environment == "production" {
		// Production: minimal logging
		gormLogger = logger.New(
			log.New(log.Writer(), "\r\n", log.LstdFlags),
			logger.Config{
				LogLevel: logger.Error, // Only log errors in production
			},
		)
	} else {
		// Development: verbose logging
		gormLogger = logger.New(
			log.New(log.Writer(), "\r\n", log.LstdFlags),
			logger.Config{
				LogLevel: logger.Info,
			},
		)
	}

	// Get database connection string
	dsn := cfg.GetDSN()

	// Log connection attempt (without sensitive info in production)
	if cfg.Environment == "production" {
		log.Printf("Connecting to production database...")
		// Log parsed connection details for debugging (without password)
		log.Printf("Database host: %s, port: %s, user: %s, database: %s, sslmode: %s",
			cfg.DBHost, cfg.DBPort, cfg.DBUser, cfg.DBName, cfg.DBSSLMode)
	} else {
		log.Printf("Connecting to database: %s:%s/%s", cfg.DBHost, cfg.DBPort, cfg.DBName)
	}

	// Open database connection
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{
		Logger: gormLogger,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %v", err)
	}

	// Configure connection pool settings to prevent connection timeouts
	sqlDB, err := db.DB()
	if err != nil {
		return nil, fmt.Errorf("failed to get underlying sql.DB: %v", err)
	}

	// Set connection pool parameters
	sqlDB.SetMaxOpenConns(25)                  // Maximum number of open connections
	sqlDB.SetMaxIdleConns(5)                   // Maximum number of idle connections
	sqlDB.SetConnMaxLifetime(15 * time.Minute) // Maximum lifetime of a connection (15 minutes)
	sqlDB.SetConnMaxIdleTime(8 * time.Minute)  // Maximum idle time for a connection (8 minutes)

	log.Printf("Successfully connected to database with connection pool configured")

	// Auto-migrate models
	log.Printf("Running database migrations...")
	if err := db.AutoMigrate(&models.Users{}, &models.Bills{}, &models.Items{}, &models.Participants{}, &models.ItemAssignments{}); err != nil {
		return nil, fmt.Errorf("failed to migrate database: %v", err)
	}
	log.Printf("Database migrations completed successfully")

	return &DB{db}, nil
}

// HealthCheck performs a database health check by pinging the database
func (d *DB) HealthCheck() error {
	sqlDB, err := d.DB.DB()
	if err != nil {
		return fmt.Errorf("failed to get underlying sql.DB: %v", err)
	}

	// Ping the database to check connectivity
	if err := sqlDB.Ping(); err != nil {
		return fmt.Errorf("database ping failed: %v", err)
	}

	return nil
}
