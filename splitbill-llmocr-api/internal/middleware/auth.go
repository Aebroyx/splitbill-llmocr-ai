package middleware

import (
	"net/http"

	"log"

	"github.com/Aebroyx/splitbill-llmocr-api/internal/domain/models"
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"gorm.io/gorm"
)

func Auth(jwtSecret string, db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Get access token from cookie
		accessToken, err := c.Cookie("access_token")
		if err != nil {
			// If access token is not found, try to refresh using refresh token
			if _, err := c.Cookie("refresh_token"); err != nil {
				c.JSON(http.StatusUnauthorized, gin.H{"error": "Authentication required"})
				c.Abort()
				return
			}

			// TODO: Implement token refresh logic
			// For now, just return unauthorized
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Access token expired"})
			c.Abort()
			return
		}

		// Parse and validate token
		claims := &models.Claims{}
		token, err := jwt.ParseWithClaims(accessToken, claims, func(token *jwt.Token) (interface{}, error) {
			return []byte(jwtSecret), nil
		})

		if err != nil {
			if err == jwt.ErrSignatureInvalid {
				c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid token signature"})
			} else if err == jwt.ErrTokenExpired {
				c.JSON(http.StatusUnauthorized, gin.H{"error": "Token has expired"})
			} else {
				c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid token"})
			}
			c.Abort()
			return
		}

		if !token.Valid {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid token"})
			c.Abort()
			return
		}

		// Get user from database
		var user models.Users
		if err := db.First(&user, claims.UserID).Error; err != nil {
			log.Printf("Auth middleware: user not found in database for ID %d", claims.UserID)
			c.JSON(http.StatusUnauthorized, gin.H{"error": "User not found"})
			c.Abort()
			return
		}

		// Create user response object
		userResponse := models.RegisterResponse{
			ID:       user.ID,
			Username: user.Username,
			Email:    user.Email,
			Name:     user.Name,
			Role:     user.Role,
		}

		log.Printf("Auth middleware: setting user in context: %+v", userResponse)

		// Set user in context
		c.Set("user", userResponse)

		c.Next()
	}
}
