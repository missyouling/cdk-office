package handler

import (
	"bytes"
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"cdk-office/internal/app/domain"
	"cdk-office/internal/app/service"
	"cdk-office/internal/shared/testutils"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockPermissionService is a mock implementation of PermissionServiceInterface
type MockPermissionService struct {
	mock.Mock
}

func (m *MockPermissionService) CreatePermission(ctx context.Context, req *service.CreatePermissionRequest) (*domain.AppPermission, error) {
	args := m.Called(ctx, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.AppPermission), args.Error(1)
}

func (m *MockPermissionService) UpdatePermission(ctx context.Context, permissionID string, req *service.UpdatePermissionRequest) error {
	args := m.Called(ctx, permissionID, req)
	return args.Error(0)
}

func (m *MockPermissionService) DeletePermission(ctx context.Context, permissionID string) error {
	args := m.Called(ctx, permissionID)
	return args.Error(0)
}

func (m *MockPermissionService) ListPermissions(ctx context.Context, appID string, page, size int) ([]*domain.AppPermission, int64, error) {
	args := m.Called(ctx, appID, page, size)
	if args.Get(0) == nil {
		return nil, 0, args.Error(2)
	}
	return args.Get(0).([]*domain.AppPermission), args.Get(1).(int64), args.Error(2)
}

func (m *MockPermissionService) GetPermission(ctx context.Context, permissionID string) (*domain.AppPermission, error) {
	args := m.Called(ctx, permissionID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.AppPermission), args.Error(1)
}

func (m *MockPermissionService) CheckPermission(ctx context.Context, appID, userID, action string) (bool, error) {
	args := m.Called(ctx, appID, userID, action)
	return args.Bool(0), args.Error(1)
}

func (m *MockPermissionService) AssignPermission(ctx context.Context, req *service.AssignPermissionRequest) error {
	args := m.Called(ctx, req)
	return args.Error(0)
}

func (m *MockPermissionService) RevokePermission(ctx context.Context, req *service.RevokePermissionRequest) error {
	args := m.Called(ctx, req)
	return args.Error(0)
}

func (m *MockPermissionService) ListUserPermissions(ctx context.Context, appID, userID string) ([]*domain.AppPermission, error) {
	args := m.Called(ctx, appID, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*domain.AppPermission), args.Error(1)
}

func TestNewPermissionHandler(t *testing.T) {
	handler := NewPermissionHandler()
	assert.NotNil(t, handler)
	assert.NotNil(t, handler.permissionService)
}

func TestPermissionHandler_CreatePermission(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("successful creation", func(t *testing.T) {
		// Setup
		mockService := new(MockPermissionService)
		handler := &PermissionHandler{permissionService: mockService}
		
		// Mock data
		permission := &domain.AppPermission{
			ID:          "perm_123",
			AppID:       "app_123",
			Name:        "Test Permission",
			Description: "Test Description",
			Permission:  "read",
			CreatedBy:   "user_123",
			CreatedAt:   time.Now(),
			UpdatedAt:   time.Now(),
		}
		
		// Mock service
		mockService.On("CreatePermission", mock.Anything, mock.MatchedBy(func(req *service.CreatePermissionRequest) bool {
			return req.AppID == "app_123" && req.Name == "Test Permission"
		})).Return(permission, nil)
		
		// Create request
		reqBody := `{"app_id":"app_123","name":"Test Permission","description":"Test Description","action":"read","created_by":"user_123"}`
		req, _ := http.NewRequest(http.MethodPost, "/permissions", bytes.NewBufferString(reqBody))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		
		// Create context and call handler
		c, _ := gin.CreateTestContext(w)
		c.Request = req
		
		handler.CreatePermission(c)
		
		// Assertions
		assert.Equal(t, http.StatusOK, w.Code)
		mockService.AssertExpectations(t)
	})
	
	t.Run("invalid request body", func(t *testing.T) {
		// Setup
		handler := &PermissionHandler{}
		
		// Create request with invalid JSON
		reqBody := `{"app_id":"app_123","name":"Test Permission","description":"Test Description","action":"read"` // Missing closing brace and missing created_by
		req, _ := http.NewRequest(http.MethodPost, "/permissions", bytes.NewBufferString(reqBody))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		
		// Create context and call handler
		c, _ := gin.CreateTestContext(w)
		c.Request = req
		
		handler.CreatePermission(c)
		
		// Assertions
		assert.Equal(t, http.StatusBadRequest, w.Code)
	})
	
	t.Run("invalid permission action", func(t *testing.T) {
		// Setup
		mockService := new(MockPermissionService)
		handler := &PermissionHandler{permissionService: mockService}
		
		// Mock service to return "invalid permission action" error
		mockService.On("CreatePermission", mock.Anything, mock.MatchedBy(func(req *service.CreatePermissionRequest) bool {
			return req.AppID == "app_123" && req.Name == "Test Permission"
		})).Return((*domain.AppPermission)(nil), testutils.NewError("invalid permission action"))
		
		// Create request
		reqBody := `{"app_id":"app_123","name":"Test Permission","description":"Test Description","action":"invalid","created_by":"user_123"}`
		req, _ := http.NewRequest(http.MethodPost, "/permissions", bytes.NewBufferString(reqBody))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		
		// Create context and call handler
		c, _ := gin.CreateTestContext(w)
		c.Request = req
		
		handler.CreatePermission(c)
		
		// Assertions
		assert.Equal(t, http.StatusBadRequest, w.Code)
		mockService.AssertExpectations(t)
	})
	
	t.Run("permission with this name already exists in the application", func(t *testing.T) {
		// Setup
		mockService := new(MockPermissionService)
		handler := &PermissionHandler{permissionService: mockService}
		
		// Mock service to return "permission with this name already exists in the application" error
		mockService.On("CreatePermission", mock.Anything, mock.MatchedBy(func(req *service.CreatePermissionRequest) bool {
			return req.AppID == "app_123" && req.Name == "Test Permission"
		})).Return((*domain.AppPermission)(nil), testutils.NewError("permission with this name already exists in the application"))
		
		// Create request
		reqBody := `{"app_id":"app_123","name":"Test Permission","description":"Test Description","action":"read","created_by":"user_123"}`
		req, _ := http.NewRequest(http.MethodPost, "/permissions", bytes.NewBufferString(reqBody))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		
		// Create context and call handler
		c, _ := gin.CreateTestContext(w)
		c.Request = req
		
		handler.CreatePermission(c)
		
		// Assertions
		assert.Equal(t, http.StatusBadRequest, w.Code)
		mockService.AssertExpectations(t)
	})
	
	t.Run("service error", func(t *testing.T) {
		// Setup
		mockService := new(MockPermissionService)
		handler := &PermissionHandler{permissionService: mockService}
		
		// Mock service to return error
		mockService.On("CreatePermission", mock.Anything, mock.MatchedBy(func(req *service.CreatePermissionRequest) bool {
			return req.AppID == "app_123" && req.Name == "Test Permission"
		})).Return((*domain.AppPermission)(nil), testutils.NewError("service error"))
		
		// Create request
		reqBody := `{"app_id":"app_123","name":"Test Permission","description":"Test Description","action":"read","created_by":"user_123"}`
		req, _ := http.NewRequest(http.MethodPost, "/permissions", bytes.NewBufferString(reqBody))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		
		// Create context and call handler
		c, _ := gin.CreateTestContext(w)
		c.Request = req
		
		handler.CreatePermission(c)
		
		// Assertions
		assert.Equal(t, http.StatusInternalServerError, w.Code)
		mockService.AssertExpectations(t)
	})
}

func TestPermissionHandler_UpdatePermission(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("successful update", func(t *testing.T) {
		// Setup
		mockService := new(MockPermissionService)
		handler := &PermissionHandler{permissionService: mockService}
		
		// Mock service
		mockService.On("UpdatePermission", mock.Anything, "perm_123", mock.MatchedBy(func(req *service.UpdatePermissionRequest) bool {
			return req.Name != "" && req.Description != ""
		})).Return(nil)
		
		// Create request
		reqBody := `{"name":"Updated Permission","description":"Updated Description","action":"write"}`
		req, _ := http.NewRequest(http.MethodPut, "/permissions/perm_123", bytes.NewBufferString(reqBody))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		
		// Create context and call handler
		c, _ := gin.CreateTestContext(w)
		c.Request = req
		c.AddParam("id", "perm_123")
		
		handler.UpdatePermission(c)
		
		// Assertions
		assert.Equal(t, http.StatusOK, w.Code)
		mockService.AssertExpectations(t)
	})
	
	t.Run("missing permission id", func(t *testing.T) {
		// Setup
		handler := &PermissionHandler{}
		
		// Create request without permission ID
		reqBody := `{}`
		req, _ := http.NewRequest(http.MethodPut, "/permissions/", bytes.NewBufferString(reqBody))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		
		// Create context and call handler
		c, _ := gin.CreateTestContext(w)
		c.Request = req
		
		handler.UpdatePermission(c)
		
		// Assertions
		assert.Equal(t, http.StatusBadRequest, w.Code)
	})
	
	t.Run("permission not found", func(t *testing.T) {
		// Setup
		mockService := new(MockPermissionService)
		handler := &PermissionHandler{permissionService: mockService}
		
		// Mock service to return "permission not found" error
		mockService.On("UpdatePermission", mock.Anything, "perm_123", mock.MatchedBy(func(req *service.UpdatePermissionRequest) bool {
			return req.Name == "Updated Permission"
		})).Return(testutils.NewError("permission not found"))
		
		// Create request
		reqBody := `{"name":"Updated Permission"}`
		req, _ := http.NewRequest(http.MethodPut, "/permissions/perm_123", bytes.NewBufferString(reqBody))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		
		// Create context and call handler
		c, _ := gin.CreateTestContext(w)
		c.Request = req
		c.AddParam("id", "perm_123")
		
		handler.UpdatePermission(c)
		
		// Assertions
		assert.Equal(t, http.StatusNotFound, w.Code)
		mockService.AssertExpectations(t)
	})
	
	t.Run("invalid permission action", func(t *testing.T) {
		// Setup
		mockService := new(MockPermissionService)
		handler := &PermissionHandler{permissionService: mockService}
		
		// Mock service to return "invalid permission action" error
		mockService.On("UpdatePermission", mock.Anything, "perm_123", mock.MatchedBy(func(req *service.UpdatePermissionRequest) bool {
			return req.Action == "invalid"
		})).Return(testutils.NewError("invalid permission action"))
		
		// Create request
		reqBody := `{"action":"invalid"}`
		req, _ := http.NewRequest(http.MethodPut, "/permissions/perm_123", bytes.NewBufferString(reqBody))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		
		// Create context and call handler
		c, _ := gin.CreateTestContext(w)
		c.Request = req
		c.AddParam("id", "perm_123")
		
		handler.UpdatePermission(c)
		
		// Assertions
		assert.Equal(t, http.StatusBadRequest, w.Code)
		mockService.AssertExpectations(t)
	})
	
	t.Run("service error", func(t *testing.T) {
		// Setup
		mockService := new(MockPermissionService)
		handler := &PermissionHandler{permissionService: mockService}
		
		// Mock service to return error
		mockService.On("UpdatePermission", mock.Anything, "perm_123", mock.MatchedBy(func(req *service.UpdatePermissionRequest) bool {
			return req.Name == "Updated Permission"
		})).Return(testutils.NewError("service error"))
		
		// Create request
		reqBody := `{"name":"Updated Permission"}`
		req, _ := http.NewRequest(http.MethodPut, "/permissions/perm_123", bytes.NewBufferString(reqBody))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		
		// Create context and call handler
		c, _ := gin.CreateTestContext(w)
		c.Request = req
		c.AddParam("id", "perm_123")
		
		handler.UpdatePermission(c)
		
		// Assertions
		assert.Equal(t, http.StatusInternalServerError, w.Code)
		mockService.AssertExpectations(t)
	})
}

func TestPermissionHandler_DeletePermission(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("successful deletion", func(t *testing.T) {
		// Setup
		mockService := new(MockPermissionService)
		handler := &PermissionHandler{permissionService: mockService}
		
		// Mock service
		mockService.On("DeletePermission", mock.Anything, "perm_123").Return(nil)
		
		// Create request
		req, _ := http.NewRequest(http.MethodDelete, "/permissions/perm_123", nil)
		w := httptest.NewRecorder()
		
		// Create context and call handler
		c, _ := gin.CreateTestContext(w)
		c.Request = req
		c.AddParam("id", "perm_123")
		
		handler.DeletePermission(c)
		
		// Assertions
		assert.Equal(t, http.StatusOK, w.Code)
		mockService.AssertExpectations(t)
	})
	
	t.Run("missing permission id", func(t *testing.T) {
		// Setup
		handler := &PermissionHandler{}
		
		// Create request without permission ID
		req, _ := http.NewRequest(http.MethodDelete, "/permissions/", nil)
		w := httptest.NewRecorder()
		
		// Create context and call handler
		c, _ := gin.CreateTestContext(w)
		c.Request = req
		
		handler.DeletePermission(c)
		
		// Assertions
		assert.Equal(t, http.StatusBadRequest, w.Code)
	})
	
	t.Run("permission not found", func(t *testing.T) {
		// Setup
		mockService := new(MockPermissionService)
		handler := &PermissionHandler{permissionService: mockService}
		
		// Mock service to return "permission not found" error
		mockService.On("DeletePermission", mock.Anything, "perm_123").Return(testutils.NewError("permission not found"))
		
		// Create request
		req, _ := http.NewRequest(http.MethodDelete, "/permissions/perm_123", nil)
		w := httptest.NewRecorder()
		
		// Create context and call handler
		c, _ := gin.CreateTestContext(w)
		c.Request = req
		c.AddParam("id", "perm_123")
		
		handler.DeletePermission(c)
		
		// Assertions
		assert.Equal(t, http.StatusNotFound, w.Code)
		mockService.AssertExpectations(t)
	})
	
	t.Run("service error", func(t *testing.T) {
		// Setup
		mockService := new(MockPermissionService)
		handler := &PermissionHandler{permissionService: mockService}
		
		// Mock service to return error
		mockService.On("DeletePermission", mock.Anything, "perm_123").Return(testutils.NewError("service error"))
		
		// Create request
		req, _ := http.NewRequest(http.MethodDelete, "/permissions/perm_123", nil)
		w := httptest.NewRecorder()
		
		// Create context and call handler
		c, _ := gin.CreateTestContext(w)
		c.Request = req
		c.AddParam("id", "perm_123")
		
		handler.DeletePermission(c)
		
		// Assertions
		assert.Equal(t, http.StatusInternalServerError, w.Code)
		mockService.AssertExpectations(t)
	})
}

func TestPermissionHandler_ListPermissions(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("successful listing", func(t *testing.T) {
		// Setup
		mockService := new(MockPermissionService)
		handler := &PermissionHandler{permissionService: mockService}
		
		// Mock data
		permissions := []*domain.AppPermission{
			{
				ID:          "perm_1",
				AppID:       "app_123",
				Name:        "Permission 1",
				Description: "Description 1",
				Permission:  "read",
				CreatedBy:   "user_123",
				CreatedAt:   time.Now(),
				UpdatedAt:   time.Now(),
			},
			{
				ID:          "perm_2",
				AppID:       "app_123",
				Name:        "Permission 2",
				Description: "Description 2",
				Permission:  "write",
				CreatedBy:   "user_123",
				CreatedAt:   time.Now(),
				UpdatedAt:   time.Now(),
			},
		}
		
		// Mock service
		mockService.On("ListPermissions", mock.Anything, "app_123", 1, 10).Return(permissions, int64(2), nil)
		
		// Create request
		req, _ := http.NewRequest(http.MethodGet, "/permissions?app_id=app_123", nil)
		w := httptest.NewRecorder()
		
		// Create context and call handler
		c, _ := gin.CreateTestContext(w)
		c.Request = req
		
		handler.ListPermissions(c)
		
		// Assertions
		assert.Equal(t, http.StatusOK, w.Code)
		mockService.AssertExpectations(t)
	})
	
	t.Run("missing app id", func(t *testing.T) {
		// Setup
		handler := &PermissionHandler{}
		
		// Create request without app_id
		req, _ := http.NewRequest(http.MethodGet, "/permissions", nil)
		w := httptest.NewRecorder()
		
		// Create context and call handler
		c, _ := gin.CreateTestContext(w)
		c.Request = req
		
		handler.ListPermissions(c)
		
		// Assertions
		assert.Equal(t, http.StatusBadRequest, w.Code)
	})
	
	t.Run("service error", func(t *testing.T) {
		// Setup
		mockService := new(MockPermissionService)
		handler := &PermissionHandler{permissionService: mockService}
		
		// Mock service to return error
		mockService.On("ListPermissions", mock.Anything, "app_123", 1, 10).Return([]*domain.AppPermission(nil), int64(0), testutils.NewError("service error"))
		
		// Create request
		req, _ := http.NewRequest(http.MethodGet, "/permissions?app_id=app_123", nil)
		w := httptest.NewRecorder()
		
		// Create context and call handler
		c, _ := gin.CreateTestContext(w)
		c.Request = req
		
		handler.ListPermissions(c)
		
		// Assertions
		assert.Equal(t, http.StatusInternalServerError, w.Code)
		mockService.AssertExpectations(t)
	})
}

func TestPermissionHandler_GetPermission(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("successful retrieval", func(t *testing.T) {
		// Setup
		mockService := new(MockPermissionService)
		handler := &PermissionHandler{permissionService: mockService}
		
		// Mock data
		permission := &domain.AppPermission{
			ID:          "perm_123",
			AppID:       "app_123",
			Name:        "Test Permission",
			Description: "Test Description",
			Permission:  "read",
			CreatedBy:   "user_123",
			CreatedAt:   time.Now(),
			UpdatedAt:   time.Now(),
		}
		
		// Mock service
		mockService.On("GetPermission", mock.Anything, "perm_123").Return(permission, nil)
		
		// Create request
		req, _ := http.NewRequest(http.MethodGet, "/permissions/perm_123", nil)
		w := httptest.NewRecorder()
		
		// Create context and call handler
		c, _ := gin.CreateTestContext(w)
		c.Request = req
		c.AddParam("id", "perm_123")
		
		handler.GetPermission(c)
		
		// Assertions
		assert.Equal(t, http.StatusOK, w.Code)
		mockService.AssertExpectations(t)
	})
	
	t.Run("missing permission id", func(t *testing.T) {
		// Setup
		handler := &PermissionHandler{}
		
		// Create request without permission ID
		req, _ := http.NewRequest(http.MethodGet, "/permissions/", nil)
		w := httptest.NewRecorder()
		
		// Create context and call handler
		c, _ := gin.CreateTestContext(w)
		c.Request = req
		
		handler.GetPermission(c)
		
		// Assertions
		assert.Equal(t, http.StatusBadRequest, w.Code)
	})
	
	t.Run("permission not found", func(t *testing.T) {
		// Setup
		mockService := new(MockPermissionService)
		handler := &PermissionHandler{permissionService: mockService}
		
		// Mock service to return "permission not found" error
		mockService.On("GetPermission", mock.Anything, "perm_123").Return((*domain.AppPermission)(nil), testutils.NewError("permission not found"))
		
		// Create request
		req, _ := http.NewRequest(http.MethodGet, "/permissions/perm_123", nil)
		w := httptest.NewRecorder()
		
		// Create context and call handler
		c, _ := gin.CreateTestContext(w)
		c.Request = req
		c.AddParam("id", "perm_123")
		
		handler.GetPermission(c)
		
		// Assertions
		assert.Equal(t, http.StatusNotFound, w.Code)
		mockService.AssertExpectations(t)
	})
	
	t.Run("service error", func(t *testing.T) {
		// Setup
		mockService := new(MockPermissionService)
		handler := &PermissionHandler{permissionService: mockService}
		
		// Mock service to return error
		mockService.On("GetPermission", mock.Anything, "perm_123").Return((*domain.AppPermission)(nil), testutils.NewError("service error"))
		
		// Create request
		req, _ := http.NewRequest(http.MethodGet, "/permissions/perm_123", nil)
		w := httptest.NewRecorder()
		
		// Create context and call handler
		c, _ := gin.CreateTestContext(w)
		c.Request = req
		c.AddParam("id", "perm_123")
		
		handler.GetPermission(c)
		
		// Assertions
		assert.Equal(t, http.StatusInternalServerError, w.Code)
		mockService.AssertExpectations(t)
	})
}

func TestPermissionHandler_AssignPermission(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("successful assignment", func(t *testing.T) {
		// Setup
		mockService := new(MockPermissionService)
		handler := &PermissionHandler{permissionService: mockService}
		
		// Mock service
		mockService.On("AssignPermission", mock.Anything, mock.MatchedBy(func(req *service.AssignPermissionRequest) bool {
			return req.AppID == "app_123" && req.UserID == "user_123" && req.PermissionID == "perm_123"
		})).Return(nil)
		
		// Create request
		reqBody := `{"app_id":"app_123","user_id":"user_123","permission_id":"perm_123","assigned_by":"admin_123"}`
		req, _ := http.NewRequest(http.MethodPost, "/permissions/assign", bytes.NewBufferString(reqBody))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		
		// Create context and call handler
		c, _ := gin.CreateTestContext(w)
		c.Request = req
		
		handler.AssignPermission(c)
		
		// Assertions
		assert.Equal(t, http.StatusOK, w.Code)
		mockService.AssertExpectations(t)
	})
	
	t.Run("invalid request body", func(t *testing.T) {
		// Setup
		handler := &PermissionHandler{}
		
		// Create request with invalid JSON
		reqBody := `{"app_id":"app_123","user_id":"user_123","permission_id":"perm_123"` // Missing closing brace and missing assigned_by
		req, _ := http.NewRequest(http.MethodPost, "/permissions/assign", bytes.NewBufferString(reqBody))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		
		// Create context and call handler
		c, _ := gin.CreateTestContext(w)
		c.Request = req
		
		handler.AssignPermission(c)
		
		// Assertions
		assert.Equal(t, http.StatusBadRequest, w.Code)
	})
	
	t.Run("permission not found", func(t *testing.T) {
		// Setup
		mockService := new(MockPermissionService)
		handler := &PermissionHandler{permissionService: mockService}
		
		// Mock service to return "permission not found" error
		mockService.On("AssignPermission", mock.Anything, mock.MatchedBy(func(req *service.AssignPermissionRequest) bool {
			return req.AppID == "app_123" && req.UserID == "user_123" && req.PermissionID == "perm_123"
		})).Return(testutils.NewError("permission not found"))
		
		// Create request
		reqBody := `{"app_id":"app_123","user_id":"user_123","permission_id":"perm_123","assigned_by":"admin_123"}`
		req, _ := http.NewRequest(http.MethodPost, "/permissions/assign", bytes.NewBufferString(reqBody))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		
		// Create context and call handler
		c, _ := gin.CreateTestContext(w)
		c.Request = req
		
		handler.AssignPermission(c)
		
		// Assertions
		assert.Equal(t, http.StatusNotFound, w.Code)
		mockService.AssertExpectations(t)
	})
	
	t.Run("permission does not belong to this application", func(t *testing.T) {
		// Setup
		mockService := new(MockPermissionService)
		handler := &PermissionHandler{permissionService: mockService}
		
		// Mock service to return "permission does not belong to this application" error
		mockService.On("AssignPermission", mock.Anything, mock.MatchedBy(func(req *service.AssignPermissionRequest) bool {
			return req.AppID == "app_123" && req.UserID == "user_123" && req.PermissionID == "perm_123"
		})).Return(testutils.NewError("permission does not belong to this application"))
		
		// Create request
		reqBody := `{"app_id":"app_123","user_id":"user_123","permission_id":"perm_123","assigned_by":"admin_123"}`
		req, _ := http.NewRequest(http.MethodPost, "/permissions/assign", bytes.NewBufferString(reqBody))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		
		// Create context and call handler
		c, _ := gin.CreateTestContext(w)
		c.Request = req
		
		handler.AssignPermission(c)
		
		// Assertions
		assert.Equal(t, http.StatusBadRequest, w.Code)
		mockService.AssertExpectations(t)
	})
	
	t.Run("service error", func(t *testing.T) {
		// Setup
		mockService := new(MockPermissionService)
		handler := &PermissionHandler{permissionService: mockService}
		
		// Mock service to return error
		mockService.On("AssignPermission", mock.Anything, mock.MatchedBy(func(req *service.AssignPermissionRequest) bool {
			return req.AppID == "app_123" && req.UserID == "user_123" && req.PermissionID == "perm_123"
		})).Return(testutils.NewError("service error"))
		
		// Create request
		reqBody := `{"app_id":"app_123","user_id":"user_123","permission_id":"perm_123","assigned_by":"admin_123"}`
		req, _ := http.NewRequest(http.MethodPost, "/permissions/assign", bytes.NewBufferString(reqBody))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		
		// Create context and call handler
		c, _ := gin.CreateTestContext(w)
		c.Request = req
		
		handler.AssignPermission(c)
		
		// Assertions
		assert.Equal(t, http.StatusInternalServerError, w.Code)
		mockService.AssertExpectations(t)
	})
}

func TestPermissionHandler_RevokePermission(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("successful revocation", func(t *testing.T) {
		// Setup
		mockService := new(MockPermissionService)
		handler := &PermissionHandler{permissionService: mockService}
		
		// Mock service
		mockService.On("RevokePermission", mock.Anything, mock.MatchedBy(func(req *service.RevokePermissionRequest) bool {
			return req.AppID == "app_123" && req.UserID == "user_123" && req.PermissionID == "perm_123"
		})).Return(nil)
		
		// Create request
		reqBody := `{"app_id":"app_123","user_id":"user_123","permission_id":"perm_123"}`
		req, _ := http.NewRequest(http.MethodPost, "/permissions/revoke", bytes.NewBufferString(reqBody))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		
		// Create context and call handler
		c, _ := gin.CreateTestContext(w)
		c.Request = req
		
		handler.RevokePermission(c)
		
		// Assertions
		assert.Equal(t, http.StatusOK, w.Code)
		mockService.AssertExpectations(t)
	})
	
	t.Run("invalid request body", func(t *testing.T) {
		// Setup
		handler := &PermissionHandler{}
		
		// Create request with invalid JSON
		reqBody := `{"app_id":"app_123","user_id":"user_123","permission_id":"perm_123"` // Missing closing brace
		req, _ := http.NewRequest(http.MethodPost, "/permissions/revoke", bytes.NewBufferString(reqBody))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		
		// Create context and call handler
		c, _ := gin.CreateTestContext(w)
		c.Request = req
		
		handler.RevokePermission(c)
		
		// Assertions
		assert.Equal(t, http.StatusBadRequest, w.Code)
	})
	
	t.Run("service error", func(t *testing.T) {
		// Setup
		mockService := new(MockPermissionService)
		handler := &PermissionHandler{permissionService: mockService}
		
		// Mock service to return error
		mockService.On("RevokePermission", mock.Anything, mock.MatchedBy(func(req *service.RevokePermissionRequest) bool {
			return req.AppID == "app_123" && req.UserID == "user_123" && req.PermissionID == "perm_123"
		})).Return(testutils.NewError("service error"))
		
		// Create request
		reqBody := `{"app_id":"app_123","user_id":"user_123","permission_id":"perm_123"}`
		req, _ := http.NewRequest(http.MethodPost, "/permissions/revoke", bytes.NewBufferString(reqBody))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		
		// Create context and call handler
		c, _ := gin.CreateTestContext(w)
		c.Request = req
		
		handler.RevokePermission(c)
		
		// Assertions
		assert.Equal(t, http.StatusInternalServerError, w.Code)
		mockService.AssertExpectations(t)
	})
}

func TestPermissionHandler_ListUserPermissions(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("successful listing", func(t *testing.T) {
		// Setup
		mockService := new(MockPermissionService)
		handler := &PermissionHandler{permissionService: mockService}
		
		// Mock data
		permissions := []*domain.AppPermission{
			{
				ID:          "perm_1",
				AppID:       "app_123",
				Name:        "Permission 1",
				Description: "Description 1",
				Permission:  "read",
				CreatedBy:   "user_123",
				CreatedAt:   time.Now(),
				UpdatedAt:   time.Now(),
			},
			{
				ID:          "perm_2",
				AppID:       "app_123",
				Name:        "Permission 2",
				Description: "Description 2",
				Permission:  "write",
				CreatedBy:   "user_123",
				CreatedAt:   time.Now(),
				UpdatedAt:   time.Now(),
			},
		}
		
		// Mock service
		mockService.On("ListUserPermissions", mock.Anything, "app_123", "user_123").Return(permissions, nil)
		
		// Create request
		req, _ := http.NewRequest(http.MethodGet, "/permissions/user?app_id=app_123&user_id=user_123", nil)
		w := httptest.NewRecorder()
		
		// Create context and call handler
		c, _ := gin.CreateTestContext(w)
		c.Request = req
		
		handler.ListUserPermissions(c)
		
		// Assertions
		assert.Equal(t, http.StatusOK, w.Code)
		mockService.AssertExpectations(t)
	})
	
	t.Run("missing app id", func(t *testing.T) {
		// Setup
		handler := &PermissionHandler{}
		
		// Create request without app_id
		req, _ := http.NewRequest(http.MethodGet, "/permissions/user?user_id=user_123", nil)
		w := httptest.NewRecorder()
		
		// Create context and call handler
		c, _ := gin.CreateTestContext(w)
		c.Request = req
		
		handler.ListUserPermissions(c)
		
		// Assertions
		assert.Equal(t, http.StatusBadRequest, w.Code)
	})
	
	t.Run("missing user id", func(t *testing.T) {
		// Setup
		handler := &PermissionHandler{}
		
		// Create request without user_id
		req, _ := http.NewRequest(http.MethodGet, "/permissions/user?app_id=app_123", nil)
		w := httptest.NewRecorder()
		
		// Create context and call handler
		c, _ := gin.CreateTestContext(w)
		c.Request = req
		
		handler.ListUserPermissions(c)
		
		// Assertions
		assert.Equal(t, http.StatusBadRequest, w.Code)
	})
	
	t.Run("service error", func(t *testing.T) {
		// Setup
		mockService := new(MockPermissionService)
		handler := &PermissionHandler{permissionService: mockService}
		
		// Mock service to return error
		mockService.On("ListUserPermissions", mock.Anything, "app_123", "user_123").Return([]*domain.AppPermission(nil), testutils.NewError("service error"))
		
		// Create request
		req, _ := http.NewRequest(http.MethodGet, "/permissions/user?app_id=app_123&user_id=user_123", nil)
		w := httptest.NewRecorder()
		
		// Create context and call handler
		c, _ := gin.CreateTestContext(w)
		c.Request = req
		
		handler.ListUserPermissions(c)
		
		// Assertions
		assert.Equal(t, http.StatusInternalServerError, w.Code)
		mockService.AssertExpectations(t)
	})
}

func TestPermissionHandler_CheckPermission(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("successful check with permission", func(t *testing.T) {
		// Setup
		mockService := new(MockPermissionService)
		handler := &PermissionHandler{permissionService: mockService}
		
		// Mock service
		mockService.On("CheckPermission", mock.Anything, "app_123", "user_123", "read").Return(true, nil)
		
		// Create request
		reqBody := `{"app_id":"app_123","user_id":"user_123","action":"read"}`
		req, _ := http.NewRequest(http.MethodPost, "/permissions/check", bytes.NewBufferString(reqBody))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		
		// Create context and call handler
		c, _ := gin.CreateTestContext(w)
		c.Request = req
		
		handler.CheckPermission(c)
		
		// Assertions
		assert.Equal(t, http.StatusOK, w.Code)
		assert.Contains(t, w.Body.String(), `"has_permission":true`)
		mockService.AssertExpectations(t)
	})
	
	t.Run("successful check without permission", func(t *testing.T) {
		// Setup
		mockService := new(MockPermissionService)
		handler := &PermissionHandler{permissionService: mockService}
		
		// Mock service
		mockService.On("CheckPermission", mock.Anything, "app_123", "user_123", "read").Return(false, nil)
		
		// Create request
		reqBody := `{"app_id":"app_123","user_id":"user_123","action":"read"}`
		req, _ := http.NewRequest(http.MethodPost, "/permissions/check", bytes.NewBufferString(reqBody))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		
		// Create context and call handler
		c, _ := gin.CreateTestContext(w)
		c.Request = req
		
		handler.CheckPermission(c)
		
		// Assertions
		assert.Equal(t, http.StatusOK, w.Code)
		assert.Contains(t, w.Body.String(), `"has_permission":false`)
		mockService.AssertExpectations(t)
	})
	
	t.Run("invalid request body", func(t *testing.T) {
		// Setup
		handler := &PermissionHandler{}
		
		// Create request with invalid JSON
		reqBody := `{"app_id":"app_123","user_id":"user_123","action":"read"` // Missing closing brace
		req, _ := http.NewRequest(http.MethodPost, "/permissions/check", bytes.NewBufferString(reqBody))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		
		// Create context and call handler
		c, _ := gin.CreateTestContext(w)
		c.Request = req
		
		handler.CheckPermission(c)
		
		// Assertions
		assert.Equal(t, http.StatusBadRequest, w.Code)
	})
	
	t.Run("service error", func(t *testing.T) {
		// Setup
		mockService := new(MockPermissionService)
		handler := &PermissionHandler{permissionService: mockService}
		
		// Mock service to return error
		mockService.On("CheckPermission", mock.Anything, "app_123", "user_123", "read").Return(false, testutils.NewError("service error"))
		
		// Create request
		reqBody := `{"app_id":"app_123","user_id":"user_123","action":"read"}`
		req, _ := http.NewRequest(http.MethodPost, "/permissions/check", bytes.NewBufferString(reqBody))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		
		// Create context and call handler
		c, _ := gin.CreateTestContext(w)
		c.Request = req
		
		handler.CheckPermission(c)
		
		// Assertions
		assert.Equal(t, http.StatusInternalServerError, w.Code)
		mockService.AssertExpectations(t)
	})
}