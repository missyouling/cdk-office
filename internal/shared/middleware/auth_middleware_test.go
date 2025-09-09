package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"cdk-office/pkg/jwt"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockTokenBlacklist is a mock implementation of the TokenBlacklistInterface
type MockTokenBlacklist struct {
	mock.Mock
}

func (m *MockTokenBlacklist) IsBlacklisted(token string) (bool, error) {
	args := m.Called(token)
	return args.Bool(0), args.Error(1)
}

func TestAuthMiddleware_Authenticate_NoAuthHeader(t *testing.T) {
	// Setup
	gin.SetMode(gin.TestMode)
	router := gin.New()
	
	// Create middleware
	jwtManager := jwt.NewJWTManager(&jwt.JWTConfig{
		SecretKey:      "test_secret",
		AccessTokenExp: time.Hour,
		RefreshTokenExp: time.Hour * 24,
	})
	
	// Create mock blacklist that returns false for all tokens
	mockBlacklist := new(MockTokenBlacklist)
	mockBlacklist.On("IsBlacklisted", mock.Anything).Return(false, nil)
	
	authMiddleware := NewAuthMiddlewareWithBlacklist(jwtManager, mockBlacklist)
	
	// Add middleware to router
	router.Use(authMiddleware.Authenticate())
	
	// Create test request
	req := httptest.NewRequest("GET", "/test", nil)
	w := httptest.NewRecorder()
	
	// Perform request
	router.ServeHTTP(w, req)
	
	// Assertions
	assert.Equal(t, http.StatusUnauthorized, w.Code)
	assert.Contains(t, w.Body.String(), "authorization header is required")
}

func TestAuthMiddleware_Authenticate_InvalidAuthHeader(t *testing.T) {
	// Setup
	gin.SetMode(gin.TestMode)
	router := gin.New()
	
	// Create middleware
	jwtManager := jwt.NewJWTManager(&jwt.JWTConfig{
		SecretKey:      "test_secret",
		AccessTokenExp: time.Hour,
		RefreshTokenExp: time.Hour * 24,
	})
	
	// Create mock blacklist that returns false for all tokens
	mockBlacklist := new(MockTokenBlacklist)
	mockBlacklist.On("IsBlacklisted", mock.Anything).Return(false, nil)
	
	authMiddleware := NewAuthMiddlewareWithBlacklist(jwtManager, mockBlacklist)
	
	// Add middleware to router
	router.Use(authMiddleware.Authenticate())
	
	// Create test request with invalid auth header
	req := httptest.NewRequest("GET", "/test", nil)
	req.Header.Set("Authorization", "InvalidToken")
	w := httptest.NewRecorder()
	
	// Perform request
	router.ServeHTTP(w, req)
	
	// Assertions
	assert.Equal(t, http.StatusUnauthorized, w.Code)
	assert.Contains(t, w.Body.String(), "authorization header must start with Bearer")
}

func TestAuthMiddleware_Authenticate_BlacklistedToken(t *testing.T) {
	// Setup
	gin.SetMode(gin.TestMode)
	router := gin.New()
	
	// Create middleware
	jwtManager := jwt.NewJWTManager(&jwt.JWTConfig{
		SecretKey:      "test_secret",
		AccessTokenExp: time.Hour,
		RefreshTokenExp: time.Hour * 24,
	})
	
	// Create mock blacklist that returns true for the test token
	mockBlacklist := new(MockTokenBlacklist)
	mockBlacklist.On("IsBlacklisted", "blacklisted_token").Return(true, nil)
	
	authMiddleware := NewAuthMiddlewareWithBlacklist(jwtManager, mockBlacklist)
	
	// Add middleware to router
	router.Use(authMiddleware.Authenticate())
	
	// Create test request with blacklisted token
	req := httptest.NewRequest("GET", "/test", nil)
	req.Header.Set("Authorization", "Bearer blacklisted_token")
	w := httptest.NewRecorder()
	
	// Perform request
	router.ServeHTTP(w, req)
	
	// Assertions
	assert.Equal(t, http.StatusUnauthorized, w.Code)
	assert.Contains(t, w.Body.String(), "token is invalid")
}

func TestAuthMiddleware_Authenticate_InvalidToken(t *testing.T) {
	// Setup
	gin.SetMode(gin.TestMode)
	router := gin.New()
	
	// Create middleware
	jwtManager := jwt.NewJWTManager(&jwt.JWTConfig{
		SecretKey:      "test_secret",
		AccessTokenExp: time.Hour,
		RefreshTokenExp: time.Hour * 24,
	})
	
	// Create mock blacklist that returns false for all tokens
	mockBlacklist := new(MockTokenBlacklist)
	mockBlacklist.On("IsBlacklisted", mock.Anything).Return(false, nil)
	
	authMiddleware := NewAuthMiddlewareWithBlacklist(jwtManager, mockBlacklist)
	
	// Add middleware to router
	router.Use(authMiddleware.Authenticate())
	
	// Create test request with invalid token
	req := httptest.NewRequest("GET", "/test", nil)
	req.Header.Set("Authorization", "Bearer invalid_token")
	w := httptest.NewRecorder()
	
	// Perform request
	router.ServeHTTP(w, req)
	
	// Assertions
	assert.Equal(t, http.StatusUnauthorized, w.Code)
	assert.Contains(t, w.Body.String(), "invalid token")
}

func TestAuthMiddleware_Authenticate_ValidToken(t *testing.T) {
	// Setup
	gin.SetMode(gin.TestMode)
	router := gin.New()
	
	// Create JWT manager
	jwtManager := jwt.NewJWTManager(&jwt.JWTConfig{
		SecretKey:      "test_secret",
		AccessTokenExp: time.Hour,
		RefreshTokenExp: time.Hour * 24,
	})
	
	// Generate a valid token
	token, err := jwtManager.GenerateAccessToken("user123", "testuser", "user")
	assert.NoError(t, err)
	
	// Create mock blacklist that returns false for all tokens
	mockBlacklist := new(MockTokenBlacklist)
	mockBlacklist.On("IsBlacklisted", mock.Anything).Return(false, nil)
	
	// Create middleware
	authMiddleware := NewAuthMiddlewareWithBlacklist(jwtManager, mockBlacklist)
	
	// Create a handler to test that context values are set
	router.Use(authMiddleware.Authenticate())
	router.GET("/test", func(c *gin.Context) {
		userID, _ := c.Get("user_id")
		username, _ := c.Get("username")
		role, _ := c.Get("role")
		
		c.JSON(http.StatusOK, gin.H{
			"user_id":  userID,
			"username": username,
			"role":     role,
		})
	})
	
	// Create test request with valid token
	req := httptest.NewRequest("GET", "/test", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	w := httptest.NewRecorder()
	
	// Perform request
	router.ServeHTTP(w, req)
	
	// Assertions
	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), "user123")
	assert.Contains(t, w.Body.String(), "testuser")
	assert.Contains(t, w.Body.String(), "user")
}