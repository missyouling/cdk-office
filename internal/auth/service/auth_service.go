package service

import (
	"context"
	"errors"
	"time"

	"cdk-office/internal/auth/domain"
	"cdk-office/internal/shared/database"
	"cdk-office/pkg/jwt"
	"cdk-office/pkg/logger"
	jwtLib "github.com/golang-jwt/jwt/v4"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

// AuthServiceInterface defines the interface for authentication service
type AuthServiceInterface interface {
	Register(ctx context.Context, req *RegisterRequest) error
	Login(ctx context.Context, req *LoginRequest) (*LoginResponse, error)
	GetUserInfo(ctx context.Context, userID string) (*domain.User, error)
	Logout(ctx context.Context, token string) error
	RefreshToken(ctx context.Context, refreshTokenString string) (*LoginResponse, error)
}

// AuthService implements the AuthServiceInterface
type AuthService struct {
	db        *gorm.DB
	jwtManager *jwt.JWTManager
}

// NewAuthService creates a new instance of AuthService
func NewAuthService(jwtManager *jwt.JWTManager) *AuthService {
	return &AuthService{
		db:        database.GetDB(),
		jwtManager: jwtManager,
	}
}

// RegisterRequest represents the request for user registration
type RegisterRequest struct {
	Username string `json:"username" binding:"required"`
	Email    string `json:"email" binding:"required,email"`
	Phone    string `json:"phone"`
	Password string `json:"password" binding:"required,min=6"`
	RealName string `json:"real_name"`
	IDCard   string `json:"id_card"`
}

// LoginRequest represents the request for user login
type LoginRequest struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
}

// LoginResponse represents the response for user login
type LoginResponse struct {
	User         *domain.User `json:"user"`
	AccessToken  string       `json:"access_token"`
	RefreshToken string       `json:"refresh_token"`
}

// Register registers a new user
func (s *AuthService) Register(ctx context.Context, req *RegisterRequest) error {
	// Check if user already exists
	var existingUser domain.User
	if err := s.db.Where("username = ? OR email = ?", req.Username, req.Email).First(&existingUser).Error; err == nil {
		return errors.New("user already exists")
	}

	// Hash the password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		logger.Error("failed to hash password", "error", err)
		return errors.New("failed to register user")
	}

	// Create new user
	user := &domain.User{
		ID:        generateID(),
		Username:  req.Username,
		Email:     req.Email,
		Phone:     req.Phone,
		Password:  string(hashedPassword),
		RealName:  req.RealName,
		IDCard:    req.IDCard,
		Role:      "user",
		Status:    "active",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	// Save user to database
	if err := s.db.Create(user).Error; err != nil {
		logger.Error("failed to create user", "error", err)
		return errors.New("failed to register user")
	}

	return nil
}

// Login authenticates a user
func (s *AuthService) Login(ctx context.Context, req *LoginRequest) (*LoginResponse, error) {
	// Find user by username
	var user domain.User
	if err := s.db.Where("username = ?", req.Username).First(&user).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("invalid username or password")
		}
		logger.Error("failed to find user", "error", err)
		return nil, errors.New("failed to login")
	}

	// Check password
	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(req.Password)); err != nil {
		return nil, errors.New("invalid username or password")
	}

	// Generate access token
	accessToken, err := s.jwtManager.GenerateAccessToken(user.ID, user.Username, user.Role)
	if err != nil {
		logger.Error("failed to generate access token", "error", err)
		return nil, errors.New("failed to login")
	}

	// Generate refresh token
	refreshToken, err := s.jwtManager.GenerateRefreshToken(user.ID)
	if err != nil {
		logger.Error("failed to generate refresh token", "error", err)
		return nil, errors.New("failed to login")
	}

	return &LoginResponse{
		User:         &user,
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
	}, nil
}

// GetUserInfo retrieves user information by user ID
func (s *AuthService) GetUserInfo(ctx context.Context, userID string) (*domain.User, error) {
	var user domain.User
	if err := s.db.Where("id = ?", userID).First(&user).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("user not found")
		}
		logger.Error("failed to find user", "error", err)
		return nil, errors.New("failed to get user info")
	}

	return &user, nil
}

// generateID generates a unique ID (simplified implementation)
func generateID() string {
	// In a real application, use a proper ID generation library like uuid
	return "user_" + time.Now().Format("20060102150405")
}

// RefreshToken generates new tokens using a refresh token
func (s *AuthService) RefreshToken(ctx context.Context, refreshTokenString string) (*LoginResponse, error) {
	// Verify and refresh the token
	newAccessToken, newRefreshToken, err := s.jwtManager.RefreshToken(refreshTokenString)
	if err != nil {
		logger.Error("failed to refresh token", "error", err)
		return nil, errors.New("failed to refresh token")
	}

	// Get user ID from refresh token
	refreshToken, err := jwtLib.ParseWithClaims(refreshTokenString, &jwtLib.RegisteredClaims{}, func(token *jwtLib.Token) (interface{}, error) {
		return s.jwtManager.SecretKey, nil
	})

	if err != nil {
		logger.Error("failed to parse refresh token", "error", err)
		return nil, errors.New("failed to refresh token")
	}

	claims, ok := refreshToken.Claims.(*jwtLib.RegisteredClaims)
	if !ok || !refreshToken.Valid {
		return nil, errors.New("invalid refresh token")
	}

	// Get user information
	var user domain.User
	if err := s.db.Where("id = ?", claims.Subject).First(&user).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("user not found")
		}
		logger.Error("failed to find user", "error", err)
		return nil, errors.New("failed to refresh token")
	}

	return &LoginResponse{
		User:         &user,
		AccessToken:  newAccessToken,
		RefreshToken: newRefreshToken,
	}, nil
}

// Logout invalidates a user's token
func (s *AuthService) Logout(ctx context.Context, tokenString string) error {
	// Verify the token first
	claims, err := s.jwtManager.VerifyToken(tokenString)
	if err != nil {
		return errors.New("invalid token")
	}

	// Get token expiration time
	exp := claims.ExpiresAt.Time

	// Add token to blacklist
	tokenBlacklist := jwt.NewTokenBlacklist()
	if err := tokenBlacklist.AddToBlacklist(tokenString, exp); err != nil {
		logger.Error("failed to add token to blacklist", "error", err)
		return errors.New("failed to logout")
	}

	return nil
}