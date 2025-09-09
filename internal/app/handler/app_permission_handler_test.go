package handler

import (
	"bytes"
	"context"
	"encoding/json"
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

// MockAppPermissionService is a mock implementation of AppPermissionServiceInterface
type MockAppPermissionService struct {
	mock.Mock
}

func (m *MockAppPermissionService) CreateAppPermission(ctx context.Context, req *service.CreateAppPermissionRequest) (*domain.AppPermission, error) {
	args := m.Called(ctx, req)
	return args.Get(0).(*domain.AppPermission), args.Error(1)
}

func (m *MockAppPermissionService) UpdateAppPermission(ctx context.Context, permissionID string, req *service.UpdateAppPermissionRequest) error {
	args := m.Called(ctx, permissionID, req)
	return args.Error(0)
}

func (m *MockAppPermissionService) DeleteAppPermission(ctx context.Context, permissionID string) error {
	args := m.Called(ctx, permissionID)
	return args.Error(0)
}

func (m *MockAppPermissionService) ListAppPermissions(ctx context.Context, appID string, page, size int) ([]*domain.AppPermission, int64, error) {
	args := m.Called(ctx, appID, page, size)
	return args.Get(0).([]*domain.AppPermission), args.Get(1).(int64), args.Error(2)
}

func (m *MockAppPermissionService) GetAppPermission(ctx context.Context, permissionID string) (*domain.AppPermission, error) {
	args := m.Called(ctx, permissionID)
	return args.Get(0).(*domain.AppPermission), args.Error(1)
}

func (m *MockAppPermissionService) AssignPermissionToUser(ctx context.Context, req *service.AssignPermissionToUserRequest) error {
	args := m.Called(ctx, req)
	return args.Error(0)
}

func (m *MockAppPermissionService) RevokePermissionFromUser(ctx context.Context, req *service.RevokePermissionFromUserRequest) error {
	args := m.Called(ctx, req)
	return args.Error(0)
}

func (m *MockAppPermissionService) ListUserPermissions(ctx context.Context, appID, userID string) ([]*domain.AppPermission, error) {
	args := m.Called(ctx, appID, userID)
	return args.Get(0).([]*domain.AppPermission), args.Error(1)
}

func (m *MockAppPermissionService) CheckUserPermission(ctx context.Context, appID, userID, permission string) (bool, error) {
	args := m.Called(ctx, appID, userID, permission)
	return args.Bool(0), args.Error(1)
}

// TestNewAppPermissionHandler tests the NewAppPermissionHandler function
func TestNewAppPermissionHandler(t *testing.T) {
	handler := NewAppPermissionHandler()
	assert.NotNil(t, handler)
	assert.NotNil(t, handler.permissionService)
}

// TestCreateAppPermission tests the CreateAppPermission handler
func TestCreateAppPermission(t *testing.T) {
	// Set up test environment
	gin.SetMode(gin.TestMode)
	
	// Create mock service
	mockService := new(MockAppPermissionService)
	
	// Create handler with mock service
	handler := &AppPermissionHandler{
		permissionService: mockService,
	}
	
	// Create test router
	router := gin.New()
	router.POST("/permissions", handler.CreateAppPermission)
	
	// Test successful creation
	t.Run("SuccessfulCreation", func(t *testing.T) {
		// Prepare test data
		reqBody := CreateAppPermissionRequest{
			AppID:       "app_123",
			Name:        "Test Permission",
			Description: "Test permission description",
			Permission:  "read",
			CreatedBy:   "user_123",
		}
		
		// Mock service response
		expectedPermission := &domain.AppPermission{
			ID:          "perm_123",
			AppID:       "app_123",
			Name:        "Test Permission",
			Description: "Test permission description",
			Permission:  "read",
			CreatedBy:   "user_123",
			CreatedAt:   time.Now(),
			UpdatedAt:   time.Now(),
		}
		
		mockService.On("CreateAppPermission", mock.Anything, mock.MatchedBy(func(req *service.CreateAppPermissionRequest) bool {
			return req.AppID == reqBody.AppID && req.Name == reqBody.Name && req.Permission == reqBody.Permission
		})).Return(expectedPermission, nil).Once()
		
		// Create request
		jsonValue, _ := json.Marshal(reqBody)
		req, _ := http.NewRequest(http.MethodPost, "/permissions", bytes.NewBuffer(jsonValue))
		req.Header.Set("Content-Type", "application/json")
		
		// Create response recorder
		w := httptest.NewRecorder()
		
		// Perform request
		router.ServeHTTP(w, req)
		
		// Assert response
		assert.Equal(t, http.StatusOK, w.Code)
		
		// Parse response
		var response domain.AppPermission
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.Equal(t, expectedPermission.ID, response.ID)
		assert.Equal(t, expectedPermission.Name, response.Name)
		
		// Assert mock expectations
		mockService.AssertExpectations(t)
	})
	
	// Test invalid request body
	t.Run("InvalidRequestBody", func(t *testing.T) {
		// Create request with invalid JSON
		req, _ := http.NewRequest(http.MethodPost, "/permissions", bytes.NewBufferString("{invalid json}"))
		req.Header.Set("Content-Type", "application/json")
		
		// Create response recorder
		w := httptest.NewRecorder()
		
		// Perform request
		router.ServeHTTP(w, req)
		
		// Assert response
		assert.Equal(t, http.StatusBadRequest, w.Code)
	})
	
	// Test service error - invalid permission
	t.Run("InvalidPermission", func(t *testing.T) {
		// Prepare test data
		reqBody := CreateAppPermissionRequest{
			AppID:       "app_123",
			Name:        "Test Permission",
			Description: "Test permission description",
			Permission:  "invalid",
			CreatedBy:   "user_123",
		}
		
		// Mock service response
		mockService.On("CreateAppPermission", mock.Anything, mock.MatchedBy(func(req *service.CreateAppPermissionRequest) bool {
			return req.Permission == "invalid"
		})).Return((*domain.AppPermission)(nil), testutils.NewError("invalid permission")).Once()
		
		// Create request
		jsonValue, _ := json.Marshal(reqBody)
		req, _ := http.NewRequest(http.MethodPost, "/permissions", bytes.NewBuffer(jsonValue))
		req.Header.Set("Content-Type", "application/json")
		
		// Create response recorder
		w := httptest.NewRecorder()
		
		// Perform request
		router.ServeHTTP(w, req)
		
		// Assert response
		assert.Equal(t, http.StatusBadRequest, w.Code)
		assert.Contains(t, w.Body.String(), "invalid permission")
		
		// Assert mock expectations
		mockService.AssertExpectations(t)
	})
	
	// Test service error - permission already exists
	t.Run("PermissionAlreadyExists", func(t *testing.T) {
		// Prepare test data
		reqBody := CreateAppPermissionRequest{
			AppID:       "app_123",
			Name:        "Existing Permission",
			Description: "Test permission description",
			Permission:  "read",
			CreatedBy:   "user_123",
		}
		
		// Mock service response
		mockService.On("CreateAppPermission", mock.Anything, mock.MatchedBy(func(req *service.CreateAppPermissionRequest) bool {
			return req.Name == "Existing Permission"
		})).Return((*domain.AppPermission)(nil), testutils.NewError("permission with this name already exists in the application")).Once()
		
		// Create request
		jsonValue, _ := json.Marshal(reqBody)
		req, _ := http.NewRequest(http.MethodPost, "/permissions", bytes.NewBuffer(jsonValue))
		req.Header.Set("Content-Type", "application/json")
		
		// Create response recorder
		w := httptest.NewRecorder()
		
		// Perform request
		router.ServeHTTP(w, req)
		
		// Assert response
		assert.Equal(t, http.StatusBadRequest, w.Code)
		assert.Contains(t, w.Body.String(), "permission with this name already exists in the application")
		
		// Assert mock expectations
		mockService.AssertExpectations(t)
	})
	
	// Test service error - internal error
	t.Run("InternalServerError", func(t *testing.T) {
		// Prepare test data
		reqBody := CreateAppPermissionRequest{
			AppID:       "app_123",
			Name:        "Test Permission",
			Description: "Test permission description",
			Permission:  "read",
			CreatedBy:   "user_123",
		}
		
		// Mock service response
		mockService.On("CreateAppPermission", mock.Anything, mock.MatchedBy(func(req *service.CreateAppPermissionRequest) bool {
			return req.Name == "Test Permission"
		})).Return((*domain.AppPermission)(nil), testutils.NewError("internal error")).Once()
		
		// Create request
		jsonValue, _ := json.Marshal(reqBody)
		req, _ := http.NewRequest(http.MethodPost, "/permissions", bytes.NewBuffer(jsonValue))
		req.Header.Set("Content-Type", "application/json")
		
		// Create response recorder
		w := httptest.NewRecorder()
		
		// Perform request
		router.ServeHTTP(w, req)
		
		// Assert response
		assert.Equal(t, http.StatusInternalServerError, w.Code)
		assert.Contains(t, w.Body.String(), "internal error")
		
		// Assert mock expectations
		mockService.AssertExpectations(t)
	})
}

// TestUpdateAppPermission tests the UpdateAppPermission handler
func TestUpdateAppPermission(t *testing.T) {
	// Set up test environment
	gin.SetMode(gin.TestMode)
	
	// Create mock service
	mockService := new(MockAppPermissionService)
	
	// Create handler with mock service
	handler := &AppPermissionHandler{
		permissionService: mockService,
	}
	
	// Create test router with route parameter
	router := gin.New()
	router.PUT("/permissions/:id", handler.UpdateAppPermission)
	
	// Test successful update
	t.Run("SuccessfulUpdate", func(t *testing.T) {
		// Prepare test data
		permissionID := "perm_123"
		reqBody := UpdateAppPermissionRequest{
			Name:        "Updated Permission",
			Description: "Updated permission description",
			Permission:  "write",
		}
		
		// Mock service response
		mockService.On("UpdateAppPermission", mock.Anything, permissionID, mock.MatchedBy(func(req *service.UpdateAppPermissionRequest) bool {
			return req.Name == "Updated Permission" && req.Permission == "write"
		})).Return(nil).Once()
		
		// Create request
		jsonValue, _ := json.Marshal(reqBody)
		req, _ := http.NewRequest(http.MethodPut, "/permissions/"+permissionID, bytes.NewBuffer(jsonValue))
		req.Header.Set("Content-Type", "application/json")
		
		// Create response recorder
		w := httptest.NewRecorder()
		
		// Perform request
		router.ServeHTTP(w, req)
		
		// Assert response
		assert.Equal(t, http.StatusOK, w.Code)
		assert.Contains(t, w.Body.String(), "application permission updated successfully")
		
		// Assert mock expectations
		mockService.AssertExpectations(t)
	})
	
	// Test missing permission ID
	t.Run("MissingPermissionID", func(t *testing.T) {
		// Create request without permission ID
		reqBody := UpdateAppPermissionRequest{
			Name: "Updated Permission",
		}
		
		jsonValue, _ := json.Marshal(reqBody)
		req, _ := http.NewRequest(http.MethodPut, "/permissions/", bytes.NewBuffer(jsonValue))
		req.Header.Set("Content-Type", "application/json")
		
		// Create response recorder
		w := httptest.NewRecorder()
		
		// Perform request
		router.ServeHTTP(w, req)
		
		// Assert response
		assert.Equal(t, http.StatusNotFound, w.Code) // Changed to StatusNotFound since the route doesn't match
	})
	
	// Test invalid request body
	t.Run("InvalidRequestBody", func(t *testing.T) {
		// Create request with invalid JSON
		req, _ := http.NewRequest(http.MethodPut, "/permissions/perm_123", bytes.NewBufferString("{invalid json}"))
		req.Header.Set("Content-Type", "application/json")
		
		// Create response recorder
		w := httptest.NewRecorder()
		
		// Perform request
		router.ServeHTTP(w, req)
		
		// Assert response
		assert.Equal(t, http.StatusBadRequest, w.Code)
	})
	
	// Test service error - permission not found
	t.Run("PermissionNotFound", func(t *testing.T) {
		// Prepare test data
		permissionID := "perm_456"
		reqBody := UpdateAppPermissionRequest{
			Name: "Updated Permission",
		}
		
		// Mock service response
		mockService.On("UpdateAppPermission", mock.Anything, permissionID, mock.MatchedBy(func(req *service.UpdateAppPermissionRequest) bool {
			return req.Name == "Updated Permission"
		})).Return(testutils.NewError("application permission not found")).Once()
		
		// Create request
		jsonValue, _ := json.Marshal(reqBody)
		req, _ := http.NewRequest(http.MethodPut, "/permissions/"+permissionID, bytes.NewBuffer(jsonValue))
		req.Header.Set("Content-Type", "application/json")
		
		// Create response recorder
		w := httptest.NewRecorder()
		
		// Perform request
		router.ServeHTTP(w, req)
		
		// Assert response
		assert.Equal(t, http.StatusNotFound, w.Code)
		assert.Contains(t, w.Body.String(), "application permission not found")
		
		// Assert mock expectations
		mockService.AssertExpectations(t)
	})
	
	// Test service error - invalid permission
	t.Run("InvalidPermission", func(t *testing.T) {
		// Prepare test data
		permissionID := "perm_123"
		reqBody := UpdateAppPermissionRequest{
			Permission: "invalid",
		}
		
		// Mock service response
		mockService.On("UpdateAppPermission", mock.Anything, permissionID, mock.MatchedBy(func(req *service.UpdateAppPermissionRequest) bool {
			return req.Permission == "invalid"
		})).Return(testutils.NewError("invalid permission")).Once()
		
		// Create request
		jsonValue, _ := json.Marshal(reqBody)
		req, _ := http.NewRequest(http.MethodPut, "/permissions/"+permissionID, bytes.NewBuffer(jsonValue))
		req.Header.Set("Content-Type", "application/json")
		
		// Create response recorder
		w := httptest.NewRecorder()
		
		// Perform request
		router.ServeHTTP(w, req)
		
		// Assert response
		assert.Equal(t, http.StatusBadRequest, w.Code)
		assert.Contains(t, w.Body.String(), "invalid permission")
		
		// Assert mock expectations
		mockService.AssertExpectations(t)
	})
	
	// Test service error - internal error
	t.Run("InternalServerError", func(t *testing.T) {
		// Prepare test data
		permissionID := "perm_123"
		reqBody := UpdateAppPermissionRequest{
			Name: "Updated Permission",
		}
		
		// Mock service response
		mockService.On("UpdateAppPermission", mock.Anything, permissionID, mock.MatchedBy(func(req *service.UpdateAppPermissionRequest) bool {
			return req.Name == "Updated Permission"
		})).Return(testutils.NewError("internal error")).Once()
		
		// Create request
		jsonValue, _ := json.Marshal(reqBody)
		req, _ := http.NewRequest(http.MethodPut, "/permissions/"+permissionID, bytes.NewBuffer(jsonValue))
		req.Header.Set("Content-Type", "application/json")
		
		// Create response recorder
		w := httptest.NewRecorder()
		
		// Perform request
		router.ServeHTTP(w, req)
		
		// Assert response
		assert.Equal(t, http.StatusInternalServerError, w.Code)
		assert.Contains(t, w.Body.String(), "internal error")
		
		// Assert mock expectations
		mockService.AssertExpectations(t)
	})
}

// TestDeleteAppPermission tests the DeleteAppPermission handler
func TestDeleteAppPermission(t *testing.T) {
	// Set up test environment
	gin.SetMode(gin.TestMode)
	
	// Create mock service
	mockService := new(MockAppPermissionService)
	
	// Create handler with mock service
	handler := &AppPermissionHandler{
		permissionService: mockService,
	}
	
	// Create test router with route parameter
	router := gin.New()
	router.DELETE("/permissions/:id", handler.DeleteAppPermission)
	
	// Test successful deletion
	t.Run("SuccessfulDeletion", func(t *testing.T) {
		// Prepare test data
		permissionID := "perm_123"
		
		// Mock service response
		mockService.On("DeleteAppPermission", mock.Anything, permissionID).Return(nil).Once()
		
		// Create request
		req, _ := http.NewRequest(http.MethodDelete, "/permissions/"+permissionID, nil)
		
		// Create response recorder
		w := httptest.NewRecorder()
		
		// Perform request
		router.ServeHTTP(w, req)
		
		// Assert response
		assert.Equal(t, http.StatusOK, w.Code)
		assert.Contains(t, w.Body.String(), "application permission deleted successfully")
		
		// Assert mock expectations
		mockService.AssertExpectations(t)
	})
	
	// Test missing permission ID
	t.Run("MissingPermissionID", func(t *testing.T) {
		// Create request without permission ID
		req, _ := http.NewRequest(http.MethodDelete, "/permissions/", nil)
		
		// Create response recorder
		w := httptest.NewRecorder()
		
		// Perform request
		router.ServeHTTP(w, req)
		
		// Assert response
		assert.Equal(t, http.StatusNotFound, w.Code) // Changed to StatusNotFound since the route doesn't match
	})
	
	// Test service error - permission not found
	t.Run("PermissionNotFound", func(t *testing.T) {
		// Prepare test data
		permissionID := "perm_456"
		
		// Mock service response
		mockService.On("DeleteAppPermission", mock.Anything, permissionID).Return(testutils.NewError("application permission not found")).Once()
		
		// Create request
		req, _ := http.NewRequest(http.MethodDelete, "/permissions/"+permissionID, nil)
		
		// Create response recorder
		w := httptest.NewRecorder()
		
		// Perform request
		router.ServeHTTP(w, req)
		
		// Assert response
		assert.Equal(t, http.StatusNotFound, w.Code)
		assert.Contains(t, w.Body.String(), "application permission not found")
		
		// Assert mock expectations
		mockService.AssertExpectations(t)
	})
	
	// Test service error - internal error
	t.Run("InternalServerError", func(t *testing.T) {
		// Prepare test data
		permissionID := "perm_123"
		
		// Mock service response
		mockService.On("DeleteAppPermission", mock.Anything, permissionID).Return(testutils.NewError("internal error")).Once()
		
		// Create request
		req, _ := http.NewRequest(http.MethodDelete, "/permissions/"+permissionID, nil)
		
		// Create response recorder
		w := httptest.NewRecorder()
		
		// Perform request
		router.ServeHTTP(w, req)
		
		// Assert response
		assert.Equal(t, http.StatusInternalServerError, w.Code)
		assert.Contains(t, w.Body.String(), "internal error")
		
		// Assert mock expectations
		mockService.AssertExpectations(t)
	})
}

// TestListAppPermissions tests the ListAppPermissions handler
func TestListAppPermissions(t *testing.T) {
	// Set up test environment
	gin.SetMode(gin.TestMode)
	
	// Create mock service
	mockService := new(MockAppPermissionService)
	
	// Create handler with mock service
	handler := &AppPermissionHandler{
		permissionService: mockService,
	}
	
	// Create test router
	router := gin.New()
	router.GET("/permissions", handler.ListAppPermissions)
	
	// Test successful listing
	t.Run("SuccessfulListing", func(t *testing.T) {
		// Prepare test data
		appID := "app_123"
		
		// Mock service response
		expectedPermissions := []*domain.AppPermission{
			{
				ID:          "perm_123",
				AppID:       appID,
				Name:        "Read Permission",
				Permission:  "read",
				CreatedBy:   "user_123",
				CreatedAt:   time.Now(),
				UpdatedAt:   time.Now(),
			},
			{
				ID:          "perm_456",
				AppID:       appID,
				Name:        "Write Permission",
				Permission:  "write",
				CreatedBy:   "user_123",
				CreatedAt:   time.Now(),
				UpdatedAt:   time.Now(),
			},
		}
		
		mockService.On("ListAppPermissions", mock.Anything, appID, 1, 10).Return(expectedPermissions, int64(2), nil).Once()
		
		// Create request
		req, _ := http.NewRequest(http.MethodGet, "/permissions?app_id="+appID, nil)
		
		// Create response recorder
		w := httptest.NewRecorder()
		
		// Perform request
		router.ServeHTTP(w, req)
		
		// Assert response
		assert.Equal(t, http.StatusOK, w.Code)
		
		// Parse response
		var response ListAppPermissionsResponse
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.Equal(t, int64(2), response.Total)
		assert.Equal(t, 2, len(response.Items))
		assert.Equal(t, "Read Permission", response.Items[0].Name)
		
		// Assert mock expectations
		mockService.AssertExpectations(t)
	})
	
	// Test missing app ID
	t.Run("MissingAppID", func(t *testing.T) {
		// Create request without app ID
		req, _ := http.NewRequest(http.MethodGet, "/permissions", nil)
		
		// Create response recorder
		w := httptest.NewRecorder()
		
		// Perform request
		router.ServeHTTP(w, req)
		
		// Assert response
		assert.Equal(t, http.StatusBadRequest, w.Code)
		assert.Contains(t, w.Body.String(), "app id is required")
	})
	
	// Test service error
	t.Run("ServiceError", func(t *testing.T) {
		// Prepare test data
		appID := "app_123"
		
		// Mock service response
		mockService.On("ListAppPermissions", mock.Anything, appID, 1, 10).Return([]*domain.AppPermission(nil), int64(0), testutils.NewError("internal error")).Once()
		
		// Create request
		req, _ := http.NewRequest(http.MethodGet, "/permissions?app_id="+appID, nil)
		
		// Create response recorder
		w := httptest.NewRecorder()
		
		// Perform request
		router.ServeHTTP(w, req)
		
		// Assert response
		assert.Equal(t, http.StatusInternalServerError, w.Code)
		assert.Contains(t, w.Body.String(), "internal error")
		
		// Assert mock expectations
		mockService.AssertExpectations(t)
	})
	
	// Test with custom pagination
	t.Run("CustomPagination", func(t *testing.T) {
		// Prepare test data
		appID := "app_123"
		
		// Mock service response
		mockService.On("ListAppPermissions", mock.Anything, appID, 2, 5).Return([]*domain.AppPermission{}, int64(0), nil).Once()
		
		// Create request with custom pagination
		req, _ := http.NewRequest(http.MethodGet, "/permissions?app_id="+appID+"&page=2&size=5", nil)
		
		// Create response recorder
		w := httptest.NewRecorder()
		
		// Perform request
		router.ServeHTTP(w, req)
		
		// Assert response
		assert.Equal(t, http.StatusOK, w.Code)
		
		// Assert mock expectations
		mockService.AssertExpectations(t)
	})
}

// TestGetAppPermission tests the GetAppPermission handler
func TestGetAppPermission(t *testing.T) {
	// Set up test environment
	gin.SetMode(gin.TestMode)
	
	// Create mock service
	mockService := new(MockAppPermissionService)
	
	// Create handler with mock service
	handler := &AppPermissionHandler{
		permissionService: mockService,
	}
	
	// Create test router with route parameter
	router := gin.New()
	router.GET("/permissions/:id", handler.GetAppPermission)
	
	// Test successful retrieval
	t.Run("SuccessfulRetrieval", func(t *testing.T) {
		// Prepare test data
		permissionID := "perm_123"
		
		// Mock service response
		expectedPermission := &domain.AppPermission{
			ID:          permissionID,
			AppID:       "app_123",
			Name:        "Test Permission",
			Permission:  "read",
			CreatedBy:   "user_123",
			CreatedAt:   time.Now(),
			UpdatedAt:   time.Now(),
		}
		
		mockService.On("GetAppPermission", mock.Anything, permissionID).Return(expectedPermission, nil).Once()
		
		// Create request
		req, _ := http.NewRequest(http.MethodGet, "/permissions/"+permissionID, nil)
		
		// Create response recorder
		w := httptest.NewRecorder()
		
		// Perform request
		router.ServeHTTP(w, req)
		
		// Assert response
		assert.Equal(t, http.StatusOK, w.Code)
		
		// Parse response
		var response domain.AppPermission
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.Equal(t, expectedPermission.ID, response.ID)
		assert.Equal(t, expectedPermission.Name, response.Name)
		
		// Assert mock expectations
		mockService.AssertExpectations(t)
	})
	
	// Test missing permission ID
	t.Run("MissingPermissionID", func(t *testing.T) {
		// Create request without permission ID
		req, _ := http.NewRequest(http.MethodGet, "/permissions/", nil)
		
		// Create response recorder
		w := httptest.NewRecorder()
		
		// Perform request
		router.ServeHTTP(w, req)
		
		// Assert response
		assert.Equal(t, http.StatusNotFound, w.Code) // Changed to StatusNotFound since the route doesn't match
	})
	
	// Test service error - permission not found
	t.Run("PermissionNotFound", func(t *testing.T) {
		// Prepare test data
		permissionID := "perm_456"
		
		// Mock service response
		mockService.On("GetAppPermission", mock.Anything, permissionID).Return((*domain.AppPermission)(nil), testutils.NewError("application permission not found")).Once()
		
		// Create request
		req, _ := http.NewRequest(http.MethodGet, "/permissions/"+permissionID, nil)
		
		// Create response recorder
		w := httptest.NewRecorder()
		
		// Perform request
		router.ServeHTTP(w, req)
		
		// Assert response
		assert.Equal(t, http.StatusNotFound, w.Code)
		assert.Contains(t, w.Body.String(), "application permission not found")
		
		// Assert mock expectations
		mockService.AssertExpectations(t)
	})
	
	// Test service error - internal error
	t.Run("InternalServerError", func(t *testing.T) {
		// Prepare test data
		permissionID := "perm_123"
		
		// Mock service response
		mockService.On("GetAppPermission", mock.Anything, permissionID).Return((*domain.AppPermission)(nil), testutils.NewError("internal error")).Once()
		
		// Create request
		req, _ := http.NewRequest(http.MethodGet, "/permissions/"+permissionID, nil)
		
		// Create response recorder
		w := httptest.NewRecorder()
		
		// Perform request
		router.ServeHTTP(w, req)
		
		// Assert response
		assert.Equal(t, http.StatusInternalServerError, w.Code)
		assert.Contains(t, w.Body.String(), "internal error")
		
		// Assert mock expectations
		mockService.AssertExpectations(t)
	})
}

// TestAssignPermissionToUser tests the AssignPermissionToUser handler
func TestAssignPermissionToUser(t *testing.T) {
	// Set up test environment
	gin.SetMode(gin.TestMode)
	
	// Create mock service
	mockService := new(MockAppPermissionService)
	
	// Create handler with mock service
	handler := &AppPermissionHandler{
		permissionService: mockService,
	}
	
	// Create test router
	router := gin.New()
	router.POST("/permissions/assign", handler.AssignPermissionToUser)
	
	// Test successful assignment
	t.Run("SuccessfulAssignment", func(t *testing.T) {
		// Prepare test data
		reqBody := AssignPermissionToUserRequest{
			AppID:        "app_123",
			UserID:       "user_123",
			PermissionID: "perm_123",
			AssignedBy:   "admin_123",
		}
		
		// Mock service response
		mockService.On("AssignPermissionToUser", mock.Anything, mock.MatchedBy(func(req *service.AssignPermissionToUserRequest) bool {
			return req.AppID == "app_123" && req.UserID == "user_123" && req.PermissionID == "perm_123"
		})).Return(nil).Once()
		
		// Create request
		jsonValue, _ := json.Marshal(reqBody)
		req, _ := http.NewRequest(http.MethodPost, "/permissions/assign", bytes.NewBuffer(jsonValue))
		req.Header.Set("Content-Type", "application/json")
		
		// Create response recorder
		w := httptest.NewRecorder()
		
		// Perform request
		router.ServeHTTP(w, req)
		
		// Assert response
		assert.Equal(t, http.StatusOK, w.Code)
		assert.Contains(t, w.Body.String(), "permission assigned to user successfully")
		
		// Assert mock expectations
		mockService.AssertExpectations(t)
	})
	
	// Test invalid request body
	t.Run("InvalidRequestBody", func(t *testing.T) {
		// Create request with invalid JSON
		req, _ := http.NewRequest(http.MethodPost, "/permissions/assign", bytes.NewBufferString("{invalid json}"))
		req.Header.Set("Content-Type", "application/json")
		
		// Create response recorder
		w := httptest.NewRecorder()
		
		// Perform request
		router.ServeHTTP(w, req)
		
		// Assert response
		assert.Equal(t, http.StatusBadRequest, w.Code)
	})
	
	// Test service error - permission not found
	t.Run("PermissionNotFound", func(t *testing.T) {
		// Prepare test data
		reqBody := AssignPermissionToUserRequest{
			AppID:        "app_123",
			UserID:       "user_123",
			PermissionID: "perm_456",
			AssignedBy:   "admin_123",
		}
		
		// Mock service response
		mockService.On("AssignPermissionToUser", mock.Anything, mock.MatchedBy(func(req *service.AssignPermissionToUserRequest) bool {
			return req.PermissionID == "perm_456"
		})).Return(testutils.NewError("application permission not found")).Once()
		
		// Create request
		jsonValue, _ := json.Marshal(reqBody)
		req, _ := http.NewRequest(http.MethodPost, "/permissions/assign", bytes.NewBuffer(jsonValue))
		req.Header.Set("Content-Type", "application/json")
		
		// Create response recorder
		w := httptest.NewRecorder()
		
		// Perform request
		router.ServeHTTP(w, req)
		
		// Assert response
		assert.Equal(t, http.StatusNotFound, w.Code)
		assert.Contains(t, w.Body.String(), "application permission not found")
		
		// Assert mock expectations
		mockService.AssertExpectations(t)
	})
	
	// Test service error - permission does not belong to app
	t.Run("PermissionDoesNotBelongToApp", func(t *testing.T) {
		// Prepare test data
		reqBody := AssignPermissionToUserRequest{
			AppID:        "app_123",
			UserID:       "user_123",
			PermissionID: "perm_456",
			AssignedBy:   "admin_123",
		}
		
		// Mock service response
		mockService.On("AssignPermissionToUser", mock.Anything, mock.MatchedBy(func(req *service.AssignPermissionToUserRequest) bool {
			return req.AppID == "app_123" && req.PermissionID == "perm_456"
		})).Return(testutils.NewError("application permission does not belong to the specified application")).Once()
		
		// Create request
		jsonValue, _ := json.Marshal(reqBody)
		req, _ := http.NewRequest(http.MethodPost, "/permissions/assign", bytes.NewBuffer(jsonValue))
		req.Header.Set("Content-Type", "application/json")
		
		// Create response recorder
		w := httptest.NewRecorder()
		
		// Perform request
		router.ServeHTTP(w, req)
		
		// Assert response
		assert.Equal(t, http.StatusBadRequest, w.Code)
		assert.Contains(t, w.Body.String(), "application permission does not belong to the specified application")
		
		// Assert mock expectations
		mockService.AssertExpectations(t)
	})
	
	// Test service error - permission already assigned
	t.Run("PermissionAlreadyAssigned", func(t *testing.T) {
		// Prepare test data
		reqBody := AssignPermissionToUserRequest{
			AppID:        "app_123",
			UserID:       "user_123",
			PermissionID: "perm_123",
			AssignedBy:   "admin_123",
		}
		
		// Mock service response
		mockService.On("AssignPermissionToUser", mock.Anything, mock.MatchedBy(func(req *service.AssignPermissionToUserRequest) bool {
			return req.PermissionID == "perm_123"
		})).Return(testutils.NewError("permission already assigned to user")).Once()
		
		// Create request
		jsonValue, _ := json.Marshal(reqBody)
		req, _ := http.NewRequest(http.MethodPost, "/permissions/assign", bytes.NewBuffer(jsonValue))
		req.Header.Set("Content-Type", "application/json")
		
		// Create response recorder
		w := httptest.NewRecorder()
		
		// Perform request
		router.ServeHTTP(w, req)
		
		// Assert response
		assert.Equal(t, http.StatusBadRequest, w.Code)
		assert.Contains(t, w.Body.String(), "permission already assigned to user")
		
		// Assert mock expectations
		mockService.AssertExpectations(t)
	})
	
	// Test service error - internal error
	t.Run("InternalServerError", func(t *testing.T) {
		// Prepare test data
		reqBody := AssignPermissionToUserRequest{
			AppID:        "app_123",
			UserID:       "user_123",
			PermissionID: "perm_123",
			AssignedBy:   "admin_123",
		}
		
		// Mock service response
		mockService.On("AssignPermissionToUser", mock.Anything, mock.MatchedBy(func(req *service.AssignPermissionToUserRequest) bool {
			return req.PermissionID == "perm_123"
		})).Return(testutils.NewError("internal error")).Once()
		
		// Create request
		jsonValue, _ := json.Marshal(reqBody)
		req, _ := http.NewRequest(http.MethodPost, "/permissions/assign", bytes.NewBuffer(jsonValue))
		req.Header.Set("Content-Type", "application/json")
		
		// Create response recorder
		w := httptest.NewRecorder()
		
		// Perform request
		router.ServeHTTP(w, req)
		
		// Assert response
		assert.Equal(t, http.StatusInternalServerError, w.Code)
		assert.Contains(t, w.Body.String(), "internal error")
		
		// Assert mock expectations
		mockService.AssertExpectations(t)
	})
}

// TestRevokePermissionFromUser tests the RevokePermissionFromUser handler
func TestRevokePermissionFromUser(t *testing.T) {
	// Set up test environment
	gin.SetMode(gin.TestMode)
	
	// Create mock service
	mockService := new(MockAppPermissionService)
	
	// Create handler with mock service
	handler := &AppPermissionHandler{
		permissionService: mockService,
	}
	
	// Create test router
	router := gin.New()
	router.POST("/permissions/revoke", handler.RevokePermissionFromUser)
	
	// Test successful revocation
	t.Run("SuccessfulRevocation", func(t *testing.T) {
		// Prepare test data
		reqBody := RevokePermissionFromUserRequest{
			AppID:        "app_123",
			UserID:       "user_123",
			PermissionID: "perm_123",
		}
		
		// Mock service response
		mockService.On("RevokePermissionFromUser", mock.Anything, mock.MatchedBy(func(req *service.RevokePermissionFromUserRequest) bool {
			return req.AppID == "app_123" && req.UserID == "user_123" && req.PermissionID == "perm_123"
		})).Return(nil).Once()
		
		// Create request
		jsonValue, _ := json.Marshal(reqBody)
		req, _ := http.NewRequest(http.MethodPost, "/permissions/revoke", bytes.NewBuffer(jsonValue))
		req.Header.Set("Content-Type", "application/json")
		
		// Create response recorder
		w := httptest.NewRecorder()
		
		// Perform request
		router.ServeHTTP(w, req)
		
		// Assert response
		assert.Equal(t, http.StatusOK, w.Code)
		assert.Contains(t, w.Body.String(), "permission revoked from user successfully")
		
		// Assert mock expectations
		mockService.AssertExpectations(t)
	})
	
	// Test invalid request body
	t.Run("InvalidRequestBody", func(t *testing.T) {
		// Create request with invalid JSON
		req, _ := http.NewRequest(http.MethodPost, "/permissions/revoke", bytes.NewBufferString("{invalid json}"))
		req.Header.Set("Content-Type", "application/json")
		
		// Create response recorder
		w := httptest.NewRecorder()
		
		// Perform request
		router.ServeHTTP(w, req)
		
		// Assert response
		assert.Equal(t, http.StatusBadRequest, w.Code)
	})
	
	// Test service error - internal error
	t.Run("InternalServerError", func(t *testing.T) {
		// Prepare test data
		reqBody := RevokePermissionFromUserRequest{
			AppID:        "app_123",
			UserID:       "user_123",
			PermissionID: "perm_123",
		}
		
		// Mock service response
		mockService.On("RevokePermissionFromUser", mock.Anything, mock.MatchedBy(func(req *service.RevokePermissionFromUserRequest) bool {
			return req.PermissionID == "perm_123"
		})).Return(testutils.NewError("internal error")).Once()
		
		// Create request
		jsonValue, _ := json.Marshal(reqBody)
		req, _ := http.NewRequest(http.MethodPost, "/permissions/revoke", bytes.NewBuffer(jsonValue))
		req.Header.Set("Content-Type", "application/json")
		
		// Create response recorder
		w := httptest.NewRecorder()
		
		// Perform request
		router.ServeHTTP(w, req)
		
		// Assert response
		assert.Equal(t, http.StatusInternalServerError, w.Code)
		assert.Contains(t, w.Body.String(), "internal error")
		
		// Assert mock expectations
		mockService.AssertExpectations(t)
	})
}

// TestListUserPermissions tests the ListUserPermissions handler
func TestListUserPermissions(t *testing.T) {
	// Set up test environment
	gin.SetMode(gin.TestMode)
	
	// Create mock service
	mockService := new(MockAppPermissionService)
	
	// Create handler with mock service
	handler := &AppPermissionHandler{
		permissionService: mockService,
	}
	
	// Create test router
	router := gin.New()
	router.GET("/permissions/user", handler.ListUserPermissions)
	
	// Test successful listing
	t.Run("SuccessfulListing", func(t *testing.T) {
		// Prepare test data
		appID := "app_123"
		userID := "user_123"
		
		// Mock service response
		expectedPermissions := []*domain.AppPermission{
			{
				ID:          "perm_123",
				AppID:       appID,
				Name:        "Read Permission",
				Permission:  "read",
				CreatedBy:   "admin_123",
				CreatedAt:   time.Now(),
				UpdatedAt:   time.Now(),
			},
			{
				ID:          "perm_456",
				AppID:       appID,
				Name:        "Write Permission",
				Permission:  "write",
				CreatedBy:   "admin_123",
				CreatedAt:   time.Now(),
				UpdatedAt:   time.Now(),
			},
		}
		
		mockService.On("ListUserPermissions", mock.Anything, appID, userID).Return(expectedPermissions, nil).Once()
		
		// Create request
		req, _ := http.NewRequest(http.MethodGet, "/permissions/user?app_id="+appID+"&user_id="+userID, nil)
		
		// Create response recorder
		w := httptest.NewRecorder()
		
		// Perform request
		router.ServeHTTP(w, req)
		
		// Assert response
		assert.Equal(t, http.StatusOK, w.Code)
		
		// Parse response
		var response []*domain.AppPermission
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.Equal(t, 2, len(response))
		assert.Equal(t, "Read Permission", response[0].Name)
		
		// Assert mock expectations
		mockService.AssertExpectations(t)
	})
	
	// Test missing app ID
	t.Run("MissingAppID", func(t *testing.T) {
		// Create request without app ID
		req, _ := http.NewRequest(http.MethodGet, "/permissions/user?user_id=user_123", nil)
		
		// Create response recorder
		w := httptest.NewRecorder()
		
		// Perform request
		router.ServeHTTP(w, req)
		
		// Assert response
		assert.Equal(t, http.StatusBadRequest, w.Code)
		assert.Contains(t, w.Body.String(), "app id is required")
	})
	
	// Test missing user ID
	t.Run("MissingUserID", func(t *testing.T) {
		// Create request without user ID
		req, _ := http.NewRequest(http.MethodGet, "/permissions/user?app_id=app_123", nil)
		
		// Create response recorder
		w := httptest.NewRecorder()
		
		// Perform request
		router.ServeHTTP(w, req)
		
		// Assert response
		assert.Equal(t, http.StatusBadRequest, w.Code)
		assert.Contains(t, w.Body.String(), "user id is required")
	})
	
	// Test service error
	t.Run("ServiceError", func(t *testing.T) {
		// Prepare test data
		appID := "app_123"
		userID := "user_123"
		
		// Mock service response
		mockService.On("ListUserPermissions", mock.Anything, appID, userID).Return([]*domain.AppPermission(nil), testutils.NewError("internal error")).Once()
		
		// Create request
		req, _ := http.NewRequest(http.MethodGet, "/permissions/user?app_id="+appID+"&user_id="+userID, nil)
		
		// Create response recorder
		w := httptest.NewRecorder()
		
		// Perform request
		router.ServeHTTP(w, req)
		
		// Assert response
		assert.Equal(t, http.StatusInternalServerError, w.Code)
		assert.Contains(t, w.Body.String(), "internal error")
		
		// Assert mock expectations
		mockService.AssertExpectations(t)
	})
}

// TestCheckUserPermission tests the CheckUserPermission handler
func TestCheckUserPermission(t *testing.T) {
	// Set up test environment
	gin.SetMode(gin.TestMode)
	
	// Create mock service
	mockService := new(MockAppPermissionService)
	
	// Create handler with mock service
	handler := &AppPermissionHandler{
		permissionService: mockService,
	}
	
	// Create test router
	router := gin.New()
	router.POST("/permissions/check", handler.CheckUserPermission)
	
	// Test user has permission
	t.Run("UserHasPermission", func(t *testing.T) {
		// Prepare test data
		reqBody := CheckUserPermissionRequest{
			AppID:      "app_123",
			UserID:     "user_123",
			Permission: "read",
		}
		
		// Mock service response
		mockService.On("CheckUserPermission", mock.Anything, "app_123", "user_123", "read").Return(true, nil).Once()
		
		// Create request
		jsonValue, _ := json.Marshal(reqBody)
		req, _ := http.NewRequest(http.MethodPost, "/permissions/check", bytes.NewBuffer(jsonValue))
		req.Header.Set("Content-Type", "application/json")
		
		// Create response recorder
		w := httptest.NewRecorder()
		
		// Perform request
		router.ServeHTTP(w, req)
		
		// Assert response
		assert.Equal(t, http.StatusOK, w.Code)
		
		// Parse response
		var response map[string]bool
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.True(t, response["has_permission"])
		
		// Assert mock expectations
		mockService.AssertExpectations(t)
	})
	
	// Test user does not have permission
	t.Run("UserDoesNotHavePermission", func(t *testing.T) {
		// Prepare test data
		reqBody := CheckUserPermissionRequest{
			AppID:      "app_123",
			UserID:     "user_456",
			Permission: "write",
		}
		
		// Mock service response
		mockService.On("CheckUserPermission", mock.Anything, "app_123", "user_456", "write").Return(false, nil).Once()
		
		// Create request
		jsonValue, _ := json.Marshal(reqBody)
		req, _ := http.NewRequest(http.MethodPost, "/permissions/check", bytes.NewBuffer(jsonValue))
		req.Header.Set("Content-Type", "application/json")
		
		// Create response recorder
		w := httptest.NewRecorder()
		
		// Perform request
		router.ServeHTTP(w, req)
		
		// Assert response
		assert.Equal(t, http.StatusOK, w.Code)
		
		// Parse response
		var response map[string]bool
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.False(t, response["has_permission"])
		
		// Assert mock expectations
		mockService.AssertExpectations(t)
	})
	
	// Test invalid request body
	t.Run("InvalidRequestBody", func(t *testing.T) {
		// Create request with invalid JSON
		req, _ := http.NewRequest(http.MethodPost, "/permissions/check", bytes.NewBufferString("{invalid json}"))
		req.Header.Set("Content-Type", "application/json")
		
		// Create response recorder
		w := httptest.NewRecorder()
		
		// Perform request
		router.ServeHTTP(w, req)
		
		// Assert response
		assert.Equal(t, http.StatusBadRequest, w.Code)
	})
	
	// Test service error - invalid permission
	t.Run("InvalidPermission", func(t *testing.T) {
		// Prepare test data
		reqBody := CheckUserPermissionRequest{
			AppID:      "app_123",
			UserID:     "user_123",
			Permission: "invalid",
		}
		
		// Mock service response
		mockService.On("CheckUserPermission", mock.Anything, "app_123", "user_123", "invalid").Return(false, testutils.NewError("invalid permission")).Once()
		
		// Create request
		jsonValue, _ := json.Marshal(reqBody)
		req, _ := http.NewRequest(http.MethodPost, "/permissions/check", bytes.NewBuffer(jsonValue))
		req.Header.Set("Content-Type", "application/json")
		
		// Create response recorder
		w := httptest.NewRecorder()
		
		// Perform request
		router.ServeHTTP(w, req)
		
		// Assert response
		assert.Equal(t, http.StatusBadRequest, w.Code)
		assert.Contains(t, w.Body.String(), "invalid permission")
		
		// Assert mock expectations
		mockService.AssertExpectations(t)
	})
	
	// Test service error - internal error
	t.Run("InternalServerError", func(t *testing.T) {
		// Prepare test data
		reqBody := CheckUserPermissionRequest{
			AppID:      "app_123",
			UserID:     "user_123",
			Permission: "read",
		}
		
		// Mock service response
		mockService.On("CheckUserPermission", mock.Anything, "app_123", "user_123", "read").Return(false, testutils.NewError("internal error")).Once()
		
		// Create request
		jsonValue, _ := json.Marshal(reqBody)
		req, _ := http.NewRequest(http.MethodPost, "/permissions/check", bytes.NewBuffer(jsonValue))
		req.Header.Set("Content-Type", "application/json")
		
		// Create response recorder
		w := httptest.NewRecorder()
		
		// Perform request
		router.ServeHTTP(w, req)
		
		// Assert response
		assert.Equal(t, http.StatusInternalServerError, w.Code)
		assert.Contains(t, w.Body.String(), "internal error")
		
		// Assert mock expectations
		mockService.AssertExpectations(t)
	})
}