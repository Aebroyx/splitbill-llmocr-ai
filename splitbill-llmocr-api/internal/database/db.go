package database

import (
	"fmt"
	"log"

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
	// Configure GORM logger
	gormLogger := logger.New(
		log.New(log.Writer(), "\r\n", log.LstdFlags),
		logger.Config{
			LogLevel: logger.Info,
		},
	)

	// Open database connection
	db, err := gorm.Open(postgres.Open(cfg.GetDSN()), &gorm.Config{
		Logger: gormLogger,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %v", err)
	}

	// Auto-migrate models
	if err := db.AutoMigrate(&models.Users{}); err != nil {
		return nil, fmt.Errorf("failed to migrate database: %v", err)
	}

	return &DB{db}, nil
}
