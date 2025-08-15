package main

import (
	"log"
	"os"

	"github.com/Aebroyx/splitbill-llmocr-api/internal/config"
	"github.com/Aebroyx/splitbill-llmocr-api/internal/database"
	"github.com/Aebroyx/splitbill-llmocr-api/internal/handlers"
	"github.com/Aebroyx/splitbill-llmocr-api/internal/middleware"
	"github.com/Aebroyx/splitbill-llmocr-api/internal/services"
	"github.com/gin-gonic/gin"
)

func main() {
	// Set environment variable if not already set
	if os.Getenv("APP_ENV") == "" {
		os.Setenv("APP_ENV", "development")
	}

	// Set Gin mode based on environment
	if os.Getenv("APP_ENV") == "production" {
		gin.SetMode(gin.ReleaseMode)
		log.Println("Starting application in PRODUCTION mode")
	} else {
		log.Println("Starting application in DEVELOPMENT mode")
	}

	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	log.Printf("Environment: %s", cfg.Environment)
	log.Printf("Server will start on: %s", cfg.GetServerAddr())

	// Validate configuration
	if err := cfg.Validate(); err != nil {
		log.Fatalf("Invalid configuration: %v", err)
	}

	// Initialize database
	log.Println("Initializing database connection...")
	db, err := database.NewConnection(cfg)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}

	// Initialize services
	log.Println("Initializing services...")
	userService := services.NewUserService(db.DB, cfg)
	billService := services.NewBillService(db.DB)

	// Initialize handlers
	log.Println("Initializing handlers...")
	authHandler := handlers.NewAuthHandler(userService)
	billHandler := handlers.NewBillHandler(billService)

	// Initialize router
	router := gin.New() // Use gin.New() instead of gin.Default() to avoid default middleware

	// Add logger middleware
	router.Use(gin.Logger())

	// Add CORS middleware
	router.Use(func(c *gin.Context) {
		// Log incoming request
		log.Printf("Incoming request: %s %s", c.Request.Method, c.Request.URL.Path)

		// Get allowed origins from config
		allowedOrigins := cfg.CORSAllowedOrigins
		var allowedOrigin string

		// Handle multiple origins
		if len(allowedOrigins) == 0 {
			if cfg.Environment == "production" {
				allowedOrigin = "*" // Allow all origins in production if not specified
			} else {
				allowedOrigin = "http://localhost:3001" // fallback for development
			}
		} else if len(allowedOrigins) == 1 {
			allowedOrigin = allowedOrigins[0]
		} else {
			// Multiple origins - check if request origin is in the allowed list
			requestOrigin := c.Request.Header.Get("Origin")
			for _, origin := range allowedOrigins {
				if origin == requestOrigin {
					allowedOrigin = requestOrigin
					break
				}
			}
			// If no match found, don't set Access-Control-Allow-Origin header
		}

		// Set CORS headers (only if we have a valid origin)
		if allowedOrigin != "" {
			c.Writer.Header().Set("Access-Control-Allow-Origin", allowedOrigin)
		}
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

	// Serve static files (for uploaded images)
	uploadsPath := os.Getenv("UPLOADS_PATH")
	if uploadsPath == "" {
		uploadsPath = "./uploads"
	}
	router.Static("/uploads", uploadsPath)

	// All API routes
	api := router.Group("/api")
	{
		// Public routes
		auth := api.Group("/auth")
		{
			auth.POST("/register", authHandler.Register)
			auth.POST("/login", authHandler.Login)
		}

		bills := api.Group("/bills")
		{
			bills.POST("/", billHandler.CreateBill)
			bills.GET("/:id", billHandler.GetBill)
			bills.PUT("/:id", billHandler.UpdateBill)
			bills.GET("/:id/status", billHandler.GetBillStatus)
			bills.POST("/:id/image", billHandler.UploadBillImage)
			bills.GET("/:id/summary", billHandler.GetBillSummary)
			bills.GET("/:id/participants", billHandler.GetParticipants)
			bills.POST("/:id/participants", billHandler.AddParticipant)
			bills.DELETE("/:id/participants/:participantId", billHandler.DeleteParticipant)
			bills.GET("/:id/item-assignments", billHandler.GetItemAssignments)
			bills.POST("/:id/assign-items", billHandler.AssignItemToParticipant)
			bills.DELETE("/:id/assign-items", billHandler.DeleteItemAssignment)
			bills.POST("/:id/process-data", billHandler.ProcessExtractedData)
		}

		// Items routes
		items := api.Group("/items")
		{
			items.PUT("/:id", billHandler.UpdateItem)
		}

		// Protected routes (with auth middleware)
		protected := api.Group("")
		protected.Use(middleware.Auth(cfg.JWTSecret, db.DB))
		{
			protected.GET("/me", authHandler.GetMe)
			protected.POST("/auth/logout", authHandler.Logout)
		}
	}

	// Start server
	log.Printf("Server starting on %s", cfg.GetServerAddr())
	log.Println("Application is ready to handle requests!")
	if err := router.Run(cfg.GetServerAddr()); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
