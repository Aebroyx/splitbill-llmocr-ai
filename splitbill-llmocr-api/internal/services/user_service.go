package services

import (
	"errors"
	"time"

	"github.com/Aebroyx/splitbill-llmocr-api/internal/config"
	"github.com/Aebroyx/splitbill-llmocr-api/internal/domain/models"
	"github.com/golang-jwt/jwt/v5"

	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

type UserService struct {
	db     *gorm.DB
	config *config.Config
}

func NewUserService(db *gorm.DB, config *config.Config) *UserService {
	return &UserService{
		db:     db,
		config: config,
	}
}

// Register creates a new user with the provided registration data
func (s *UserService) Register(req *models.RegisterRequest) (*models.RegisterResponse, error) {
	// Check if username already exists
	var existingUser models.Users
	if err := s.db.Where("username = ?", req.Username).First(&existingUser).Error; err == nil {
		return nil, errors.New("username already exists")
	} else if !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, err
	}

	// Check if email already exists
	if err := s.db.Where("email = ?", req.Email).First(&existingUser).Error; err == nil {
		return nil, errors.New("email already exists")
	} else if !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, err
	}

	// Hash password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}

	// Create new user
	user := models.Users{
		Username: req.Username,
		Email:    req.Email,
		Password: string(hashedPassword),
		Name:     req.Name,
		Role:     "user", // Default role
	}

	if err := s.db.Create(&user).Error; err != nil {
		return nil, err
	}

	// Return user data without password
	return &models.RegisterResponse{
		ID:       user.ID,
		Username: user.Username,
		Email:    user.Email,
		Name:     user.Name,
		Role:     user.Role,
	}, nil
}

// Login authenticates a user and returns tokens
func (s *UserService) Login(req *models.LoginRequest) (*models.LoginResponse, error) {
	// Find user by username
	var user models.Users
	if err := s.db.Where("username = ?", req.Username).First(&user).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("invalid username or password")
		}
		return nil, err
	}

	// Verify password
	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(req.Password)); err != nil {
		return nil, errors.New("invalid username or password")
	}

	// Generate tokens
	accessToken, accessExp, err := s.generateToken(user, s.config.JWTExpiry)
	if err != nil {
		return nil, err
	}

	refreshToken, _, err := s.generateToken(user, 24*7*time.Hour) // 7 days
	if err != nil {
		return nil, err
	}

	// Create response
	return &models.LoginResponse{
		User: models.RegisterResponse{
			ID:       user.ID,
			Username: user.Username,
			Email:    user.Email,
			Name:     user.Name,
			Role:     user.Role,
		},
		Token: models.TokenResponse{
			AccessToken:  accessToken,
			RefreshToken: refreshToken,
			TokenType:    "Bearer",
			ExpiresIn:    int64(time.Until(accessExp).Seconds()),
		},
	}, nil
}

// generateToken generates a JWT token for the user
func (s *UserService) generateToken(user models.Users, expiry time.Duration) (string, time.Time, error) {
	expirationTime := time.Now().Add(expiry)
	claims := &models.Claims{
		UserID:   user.ID,
		Username: user.Username,
		Email:    user.Email,
		Role:     user.Role,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expirationTime),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			NotBefore: jwt.NewNumericDate(time.Now()),
			Issuer:    "splitbill-llmocr-api",
			Subject:   user.Username,
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString([]byte(s.config.JWTSecret))
	if err != nil {
		return "", time.Time{}, err
	}

	return tokenString, expirationTime, nil
}
