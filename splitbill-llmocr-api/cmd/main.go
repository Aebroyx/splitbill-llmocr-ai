package main

import (
	"log"

	"github.com/Aebroyx/splitbill-llmocr-api/internal/config"
	"github.com/Aebroyx/splitbill-llmocr-api/internal/database"
	"github.com/Aebroyx/splitbill-llmocr-api/internal/handlers"
	"github.com/Aebroyx/splitbill-llmocr-api/internal/middleware"
	"github.com/Aebroyx/splitbill-llmocr-api/internal/services"
	"github.com/gin-gonic/gin"
)

func main() {
	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	// Validate configuration
	if err := cfg.Validate(); err != nil {
		log.Fatalf("Invalid configuration: %v", err)
	}

	// Initialize database
	db, err := database.NewConnection(cfg)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}

	// Initialize services
	userService := services.NewUserService(db.DB, cfg)

	// Initialize handlers
	authHandler := handlers.NewAuthHandler(userService)

	// Initialize router
	router := gin.New() // Use gin.New() instead of gin.Default() to avoid default middleware

	// Add logger middleware
	router.Use(gin.Logger())

	// Add CORS middleware
	router.Use(func(c *gin.Context) {
		// Log incoming request
		log.Printf("Incoming request: %s %s", c.Request.Method, c.Request.URL.Path)

		// Get allowed origins from config
		allowedOrigin := cfg.CORSAllowedOrigins
		if allowedOrigin == "" {
			allowedOrigin = "http://localhost:3001" // fallback
		}

		// Set CORS headers
		c.Writer.Header().Set("Access-Control-Allow-Origin", allowedOrigin)
		c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization, X-Requested-With")
		c.Writer.Header().Set("Access-Control-Max-Age", "86400") // 24 hours

		// Handle preflight
		if c.Request.Method == "OPTIONS" {
			log.Printf("Handling OPTIONS request for: %s", c.Request.URL.Path)
			c.AbortWithStatus(204)
			return
		}

		c.Next()
	})

	// Public routes
	public := router.Group("/api")
	{
		// Auth routes
		auth := public.Group("/auth")
		{
			auth.POST("/register", authHandler.Register)
			auth.POST("/login", authHandler.Login)
		}
	}

	// Protected routes
	protected := router.Group("/api")
	protected.Use(middleware.Auth(cfg.JWTSecret, db.DB))
	{
		// Add your protected routes here
		protected.GET("/me", authHandler.GetMe)
		protected.POST("/auth/logout", authHandler.Logout)
	}

	// Start server
	log.Printf("Server starting on %s", cfg.GetServerAddr())
	if err := router.Run(cfg.GetServerAddr()); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
