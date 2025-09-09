package middleware

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"cdk-office/internal/auth/domain"
	"cdk-office/internal/auth/service"
	"cdk-office/pkg/jwt"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockPermissionService is a mock implementation of the PermissionServiceInterface
type MockPermissionService struct {
	mock.Mock
}

func (m *MockPermissionService) CreatePermission(ctx context.Context, req *service.CreatePermissionRequest) (*domain.Permission, error) {
	args := m.Called(ctx, req)
	return args.Get(0).(*domain.Permission), args.Error(1)
}

func (m *MockPermissionService) GetPermissionByName(ctx context.Context, name string) (*domain.Permission, error) {
	args := m.Called(ctx, name)
	return args.Get(0).(*domain.Permission), args.Error(1)
}

func (m *MockPermissionService) CreateRole(ctx context.Context, name, description string) (*domain.Role, error) {
	args := m.Called(ctx, name, description)
	return args.Get(0).(*domain.Role), args.Error(1)
}

func (m *MockPermissionService) GetRoleByName(ctx context.Context, name string) (*domain.Role, error) {
	args := m.Called(ctx, name)
	return args.Get(0).(*domain.Role), args.Error(1)
}

func (m *MockPermissionService) AssignPermissionToRole(ctx context.Context, roleID, permissionID string) error {
	args := m.Called(ctx, roleID, permissionID)
	return args.Error(0)
}

func (m *MockPermissionService) CheckPermission(ctx context.Context, userID, resource, action string) (bool, error) {
	args := m.Called(ctx, userID, resource, action)
	return args.Bool(0), args.Error(1)
}

func TestPermissionMiddleware_CheckPermission_UnauthenticatedUser(t *testing.T) {
	// Setup
	gin.SetMode(gin.TestMode)
	
	// Create middleware
	jwtManager := jwt.NewJWTManager(&jwt.JWTConfig{
		SecretKey:      "test_secret",
		AccessTokenExp: 1,
		RefreshTokenExp: 1,
	})
	mockPermissionService := new(MockPermissionService)
	permissionMiddleware := NewPermissionMiddleware(jwtManager, mockPermissionService)
	
	// Create a test handler
	handler := permissionMiddleware.CheckPermission("document", "read")
	
	// Create test request without user authentication
	req := httptest.NewRequest("GET", "/test", nil)
	w := httptest.NewRecorder()
	
	// Create Gin context
	c, _ := gin.CreateTestContext(w)
	c.Request = req
	
	// Perform request
	handler(c)
	
	// Assertions
	assert.Equal(t, http.StatusUnauthorized, w.Code)
	assert.Contains(t, w.Body.String(), "user not authenticated")
}

func TestPermissionMiddleware_CheckPermission_InvalidUserIDType(t *testing.T) {
	// Setup
	gin.SetMode(gin.TestMode)
	
	// Create middleware
	jwtManager := jwt.NewJWTManager(&jwt.JWTConfig{
		SecretKey:      "test_secret",
		AccessTokenExp: 1,
		RefreshTokenExp: 1,
	})
	mockPermissionService := new(MockPermissionService)
	permissionMiddleware := NewPermissionMiddleware(jwtManager, mockPermissionService)
	
	// Create a test handler
	handler := permissionMiddleware.CheckPermission("document", "read")
	
	// Create test request with invalid user ID type
	req := httptest.NewRequest("GET", "/test", nil)
	w := httptest.NewRecorder()
	
	// Create Gin context
	c, _ := gin.CreateTestContext(w)
	c.Request = req
	c.Set("user_id", 123) // Setting integer instead of string
	
	// Perform request
	handler(c)
	
	// Assertions
	assert.Equal(t, http.StatusInternalServerError, w.Code)
	assert.Contains(t, w.Body.String(), "invalid user ID type")
}

func TestPermissionMiddleware_CheckPermission_PermissionCheckFailure(t *testing.T) {
	// Setup
	gin.SetMode(gin.TestMode)
	
	// Create middleware
	jwtManager := jwt.NewJWTManager(&jwt.JWTConfig{
		SecretKey:      "test_secret",
		AccessTokenExp: 1,
		RefreshTokenExp: 1,
	})
	mockPermissionService := new(MockPermissionService)
	
	// Mock the permission service to return an error
	mockPermissionService.On("CheckPermission", mock.Anything, "user123", "document", "read").Return(false, assert.AnError)
	
	permissionMiddleware := NewPermissionMiddleware(jwtManager, mockPermissionService)
	
	// Create a test handler
	handler := permissionMiddleware.CheckPermission("document", "read")
	
	// Create test request with valid user ID
	req := httptest.NewRequest("GET", "/test", nil)
	w := httptest.NewRecorder()
	
	// Create Gin context
	c, _ := gin.CreateTestContext(w)
	c.Request = req
	c.Set("user_id", "user123")
	
	// Perform request
	handler(c)
	
	// Assertions
	assert.Equal(t, http.StatusInternalServerError, w.Code)
	assert.Contains(t, w.Body.String(), "failed to check permission")
}

func TestPermissionMiddleware_CheckPermission_InsufficientPermissions(t *testing.T) {
	// Setup
	gin.SetMode(gin.TestMode)
	
	// Create middleware
	jwtManager := jwt.NewJWTManager(&jwt.JWTConfig{
		SecretKey:      "test_secret",
		AccessTokenExp: 1,
		RefreshTokenExp: 1,
	})
	mockPermissionService := new(MockPermissionService)
	
	// Mock the permission service to return false for permission check
	mockPermissionService.On("CheckPermission", mock.Anything, "user123", "document", "read").Return(false, nil)
	
	permissionMiddleware := NewPermissionMiddleware(jwtManager, mockPermissionService)
	
	// Create a test handler
	handler := permissionMiddleware.CheckPermission("document", "read")
	
	// Create test request with valid user ID
	req := httptest.NewRequest("GET", "/test", nil)
	w := httptest.NewRecorder()
	
	// Create Gin context
	c, _ := gin.CreateTestContext(w)
	c.Request = req
	c.Set("user_id", "user123")
	
	// Perform request
	handler(c)
	
	// Assertions
	assert.Equal(t, http.StatusForbidden, w.Code)
	assert.Contains(t, w.Body.String(), "insufficient permissions")
}

func TestPermissionMiddleware_CheckPermission_SufficientPermissions(t *testing.T) {
	// Setup
	gin.SetMode(gin.TestMode)
	
	// Create middleware
	jwtManager := jwt.NewJWTManager(&jwt.JWTConfig{
		SecretKey:      "test_secret",
		AccessTokenExp: 1,
		RefreshTokenExp: 1,
	})
	mockPermissionService := new(MockPermissionService)
	
	// Mock the permission service to return true for permission check
	mockPermissionService.On("CheckPermission", mock.Anything, "user123", "document", "read").Return(true, nil)
	
	permissionMiddleware := NewPermissionMiddleware(jwtManager, mockPermissionService)
	
	// Create a test handler that will be called if permissions are sufficient
	called := false
	var nextHandler gin.HandlerFunc = func(c *gin.Context) {
		called = true
		c.JSON(http.StatusOK, gin.H{"message": "access granted"})
	}
	
	// Create the middleware handler
	middlewareHandler := permissionMiddleware.CheckPermission("document", "read")
	
	// Create test request with valid user ID
	req := httptest.NewRequest("GET", "/test", nil)
	w := httptest.NewRecorder()
	
	// Create Gin context
	c, _ := gin.CreateTestContext(w)
	c.Request = req
	c.Set("user_id", "user123")
	
	// Chain the middleware with the next handler
	// First call the middleware handler, then the next handler
	middlewareHandler(c)
	
	// If the middleware allows the request to proceed, call the next handler
	if !c.IsAborted() {
		nextHandler(c)
	}
	
	// Assertions
	assert.True(t, called)
	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), "access granted")
}