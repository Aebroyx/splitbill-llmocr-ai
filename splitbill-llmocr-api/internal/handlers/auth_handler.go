package handlers

import (
	"net/http"
	"time"

	"github.com/Aebroyx/splitbill-llmocr-api/internal/domain/models"
	"github.com/Aebroyx/splitbill-llmocr-api/internal/services"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
)

type AuthHandler struct {
	userService *services.UserService
	validate    *validator.Validate
}

func NewAuthHandler(userService *services.UserService) *AuthHandler {
	return &AuthHandler{
		userService: userService,
		validate:    validator.New(),
	}
}

// Register handles user registration
func (h *AuthHandler) Register(c *gin.Context) {
	// Using Gin's context
	var req models.RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}

	// Validate request
	if err := h.validate.Struct(req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Validation failed: " + err.Error()})
		return
	}

	// Register user
	user, err := h.userService.Register(&req)
	if err != nil {
		switch err.Error() {
		case "username already exists":
			c.JSON(http.StatusConflict, gin.H{"error": "Username already exists"})
		case "email already exists":
			c.JSON(http.StatusConflict, gin.H{"error": "Email already exists"})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error"})
		}
		return
	}

	// Return success response
	c.JSON(http.StatusCreated, user)
}

func (h *AuthHandler) Login(c *gin.Context) {
	var req models.LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}

	// Validate request
	if err := h.validate.Struct(req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Validation failed: " + err.Error()})
		return
	}

	// Login user
	response, err := h.userService.Login(&req)
	if err != nil {
		switch err.Error() {
		case "invalid username or password":
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid username or password"})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error"})
		}
		return
	}

	// Set access token cookie
	c.SetCookie(
		"access_token",
		response.Token.AccessToken,
		int(response.Token.ExpiresIn),
		"/",   // path
		"",    // domain (empty for current domain)
		false, // secure (set to false for development)
		true,  // httpOnly
	)

	// Set refresh token cookie (7 days)
	c.SetCookie(
		"refresh_token",
		response.Token.RefreshToken,
		int(7*24*time.Hour.Seconds()), // 7 days
		"/",                           // path
		"",                            // domain (empty for current domain)
		false,                         // secure (set to false for development)
		true,                          // httpOnly
	)

	// Return user data only (tokens are in cookies)
	c.JSON(http.StatusOK, gin.H{
		"user": response.User,
	})
}

func (h *AuthHandler) Logout(c *gin.Context) {
	// Clear access token cookie by setting it to expire immediately
	c.SetCookie(
		"access_token",
		"",
		-1,    // MaxAge -1 means delete immediately
		"/",   // path
		"",    // domain (empty for current domain)
		false, // secure (set to false for development)
		true,  // httpOnly
	)

	// Clear refresh token cookie
	c.SetCookie(
		"refresh_token",
		"",
		-1,    // MaxAge -1 means delete immediately
		"/",   // path
		"",    // domain (empty for current domain)
		false, // secure (set to false for development)
		true,  // httpOnly
	)

	c.JSON(http.StatusOK, gin.H{
		"message": "Logged out successfully",
	})
}

func (h *AuthHandler) GetMe(c *gin.Context) {
	user, exists := c.Get("user")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	c.JSON(http.StatusOK, user)
}
