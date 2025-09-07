package service

import (
	"context"
	"errors"
	"time"

	"cdk-office/internal/auth/domain"
	"cdk-office/internal/shared/database"
	"cdk-office/pkg/jwt"
	"cdk-office/pkg/logger"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

// PasswordServiceInterface defines the interface for password authentication service
type PasswordServiceInterface interface {
	PasswordLogin(ctx context.Context, req *PasswordLoginRequest) (*LoginResponse, error)
	ChangePassword(ctx context.Context, userID, oldPassword, newPassword string) error
}

// PasswordService implements the PasswordServiceInterface
type PasswordService struct {
	db *gorm.DB
}

// NewPasswordService creates a new instance of PasswordService
func NewPasswordService() *PasswordService {
	return &PasswordService{
		db: database.GetDB(),
	}
}

// PasswordLoginRequest represents the request for password login
type PasswordLoginRequest struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
}

// PasswordLogin authenticates a user via username and password
func (s *PasswordService) PasswordLogin(ctx context.Context, req *PasswordLoginRequest) (*LoginResponse, error) {
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
	jwtManager := jwt.NewJWTManager(&jwt.JWTConfig{
		SecretKey:       "cdk-office-secret-key",
		AccessTokenExp:  time.Hour * 2,
		RefreshTokenExp: time.Hour * 24 * 7,
	})
	
	accessToken, err := jwtManager.GenerateAccessToken(user.ID, user.Username, user.Role)
	if err != nil {
		logger.Error("failed to generate access token", "error", err)
		return nil, errors.New("failed to login")
	}

	// Generate refresh token
	refreshToken, err := jwtManager.GenerateRefreshToken(user.ID)
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

// ChangePassword changes a user's password
func (s *PasswordService) ChangePassword(ctx context.Context, userID, oldPassword, newPassword string) error {
	// Find user by ID
	var user domain.User
	if err := s.db.Where("id = ?", userID).First(&user).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errors.New("user not found")
		}
		logger.Error("failed to find user", "error", err)
		return errors.New("failed to change password")
	}

	// Check old password
	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(oldPassword)); err != nil {
		return errors.New("invalid old password")
	}

	// Hash new password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(newPassword), bcrypt.DefaultCost)
	if err != nil {
		logger.Error("failed to hash new password", "error", err)
		return errors.New("failed to change password")
	}

	// Update password
	user.Password = string(hashedPassword)
	user.UpdatedAt = time.Now()

	if err := s.db.Save(&user).Error; err != nil {
		logger.Error("failed to update user password", "error", err)
		return errors.New("failed to change password")
	}

	return nil
}