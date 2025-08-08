package models

import (
	"time"

	"github.com/golang-jwt/jwt/v5"
	"gorm.io/gorm"
)

type Users struct {
	ID        uint           `json:"id" gorm:"primaryKey"`
	Username  string         `json:"username" gorm:"unique;not null;size:50"`
	Email     string         `json:"email" gorm:"unique;not null;size:255"`
	Password  string         `json:"-" gorm:"not null"` // "-" means don't include in JSON
	Name      string         `json:"name" gorm:"not null;size:100"`
	Role      string         `json:"role" gorm:"not null;default:'user';size:20"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `json:"-" gorm:"index"`
	IsDeleted bool           `json:"is_deleted" gorm:"default:false"`
}

// RegisterRequest represents the registration request payload
type RegisterRequest struct {
	Username string `json:"username" validate:"required,min=3,max=50"`
	Email    string `json:"email" validate:"required,email,max=255"`
	Password string `json:"password" validate:"required,min=6"`
	Name     string `json:"name" validate:"required,max=100"`
}

// RegisterResponse represents the registration response payload
type RegisterResponse struct {
	ID       uint   `json:"id"`
	Username string `json:"username"`
	Email    string `json:"email"`
	Name     string `json:"name"`
	Role     string `json:"role"`
}

// LoginRequest represents the login request payload
type LoginRequest struct {
	Username string `json:"username" validate:"required"`
	Password string `json:"password" validate:"required"`
}

// TokenResponse represents the token response payload
type TokenResponse struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	TokenType    string `json:"token_type"`
	ExpiresIn    int64  `json:"expires_in"`
}

// LoginResponse represents the login response payload
type LoginResponse struct {
	User  RegisterResponse `json:"user"`
	Token TokenResponse    `json:"token"`
}

// Claims represents the JWT claims
type Claims struct {
	UserID   uint   `json:"user_id"`
	Username string `json:"username"`
	Email    string `json:"email"`
	Role     string `json:"role"`
	jwt.RegisteredClaims
}
