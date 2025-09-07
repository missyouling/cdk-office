package service

import (
	"context"
	"errors"
	"time"

	"cdk-office/internal/auth/domain"
	"cdk-office/internal/shared/database"
	"cdk-office/pkg/jwt"
	"cdk-office/pkg/logger"
	"gorm.io/gorm"
)

// WeChatServiceInterface defines the interface for WeChat authentication service
type WeChatServiceInterface interface {
	WeChatLogin(ctx context.Context, code string) (*LoginResponse, error)
}

// WeChatService implements the WeChatServiceInterface
type WeChatService struct {
	db *gorm.DB
}

// NewWeChatService creates a new instance of WeChatService
func NewWeChatService() *WeChatService {
	return &WeChatService{
		db: database.GetDB(),
	}
}

// WeChatLogin authenticates a user via WeChat
func (s *WeChatService) WeChatLogin(ctx context.Context, code string) (*LoginResponse, error) {
	// Exchange code for access token and openid (simplified implementation)
	// In a real application, you would call WeChat's API to exchange the code
	// For now, we'll simulate this process
	openid, err := s.exchangeCodeForOpenID(code)
	if err != nil {
		logger.Error("failed to exchange code for openid", "error", err)
		return nil, errors.New("failed to login with WeChat")
	}

	// Find user by WeChat openid
	var user domain.User
	if err := s.db.Where("wechat_openid = ?", openid).First(&user).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			// If user doesn't exist, create a new user
			user, err = s.createWeChatUser(ctx, openid)
			if err != nil {
				logger.Error("failed to create WeChat user", "error", err)
				return nil, errors.New("failed to login with WeChat")
			}
		} else {
			logger.Error("failed to find user by WeChat openid", "error", err)
			return nil, errors.New("failed to login with WeChat")
		}
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
		return nil, errors.New("failed to login with WeChat")
	}

	// Generate refresh token
	refreshToken, err := jwtManager.GenerateRefreshToken(user.ID)
	if err != nil {
		logger.Error("failed to generate refresh token", "error", err)
		return nil, errors.New("failed to login with WeChat")
	}

	return &LoginResponse{
		User:         &user,
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
	}, nil
}

// exchangeCodeForOpenID exchanges WeChat code for openid (simplified implementation)
func (s *WeChatService) exchangeCodeForOpenID(code string) (string, error) {
	// In a real application, you would call WeChat's API to exchange the code
	// For now, we'll just return a simulated openid
	// This is where you would make an HTTP request to WeChat's API
	return "wechat_openid_" + code, nil
}

// createWeChatUser creates a new user from WeChat information (simplified implementation)
func (s *WeChatService) createWeChatUser(ctx context.Context, openid string) (domain.User, error) {
	user := domain.User{
		ID:          generateID(),
		Username:    "wechat_user_" + openid,
		Email:       "", // WeChat login doesn't provide email by default
		Phone:       "", // WeChat login doesn't provide phone by default
		Password:    "", // WeChat login doesn't use password
		RealName:    "", // WeChat login doesn't provide real name by default
		IDCard:      "", // WeChat login doesn't provide ID card by default
		Role:        "user",
		Status:      "active",
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	// Save user to database
	if err := s.db.Create(&user).Error; err != nil {
		return domain.User{}, err
	}

	return user, nil
}