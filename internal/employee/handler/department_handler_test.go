package handler

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"cdk-office/internal/employee/domain"
	"cdk-office/internal/employee/service"
	"cdk-office/internal/shared/testutils"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockDepartmentService is a mock implementation of DepartmentServiceInterface
type MockDepartmentService struct {
	mock.Mock
}

func (m *MockDepartmentService) CreateDepartment(ctx context.Context, req *service.CreateDepartmentRequest) error {
	args := m.Called(ctx, req)
	return args.Error(0)
}

func (m *MockDepartmentService) UpdateDepartment(ctx context.Context, deptID string, req *service.UpdateDepartmentRequest) error {
	args := m.Called(ctx, deptID, req)
	return args.Error(0)
}

func (m *MockDepartmentService) DeleteDepartment(ctx context.Context, deptID string) error {
	args := m.Called(ctx, deptID)
	return args.Error(0)
}

func (m *MockDepartmentService) ListDepartments(ctx context.Context, teamID string) ([]*domain.Department, error) {
	args := m.Called(ctx, teamID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*domain.Department), args.Error(1)
}

func (m *MockDepartmentService) GetDepartment(ctx context.Context, deptID string) (*domain.Department, error) {
	args := m.Called(ctx, deptID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.Department), args.Error(1)
}

// TestNewDepartmentHandler tests the NewDepartmentHandler function
func TestNewDepartmentHandler(t *testing.T) {
	handler := NewDepartmentHandler()
	assert.NotNil(t, handler)
	assert.NotNil(t, handler.departmentService)
}

// TestCreateDepartment tests the CreateDepartment handler
func TestCreateDepartment(t *testing.T) {
	// Set up test environment
	gin.SetMode(gin.TestMode)

	// Create mock service
	mockService := new(MockDepartmentService)

	// Create handler with mock service
	handler := &DepartmentHandler{
		departmentService: mockService,
	}

	// Create test router
	router := gin.New()
	router.POST("/departments", handler.CreateDepartment)

	// Test successful creation
	t.Run("SuccessfulCreation", func(t *testing.T) {
		// Prepare test data
		reqBody := CreateDepartmentRequest{
			Name:        "Engineering",
			Description: "Engineering Department",
			TeamID:      "team_123",
			ParentID:    "dept_456",
		}

		// Mock service response
		mockService.On("CreateDepartment", mock.Anything, mock.MatchedBy(func(req *service.CreateDepartmentRequest) bool {
			return req.Name == "Engineering" && req.TeamID == "team_123"
		})).Return(nil).Once()

		// Create request
		jsonValue, _ := json.Marshal(reqBody)
		req, _ := http.NewRequest(http.MethodPost, "/departments", bytes.NewBuffer(jsonValue))
		req.Header.Set("Content-Type", "application/json")

		// Create response recorder
		w := httptest.NewRecorder()

		// Perform request
		router.ServeHTTP(w, req)

		// Assert response
		assert.Equal(t, http.StatusOK, w.Code)
		assert.Contains(t, w.Body.String(), "department created successfully")

		// Assert mock expectations
		mockService.AssertExpectations(t)
	})

	// Test invalid request body
	t.Run("InvalidRequestBody", func(t *testing.T) {
		// Create request with invalid JSON
		req, _ := http.NewRequest(http.MethodPost, "/departments", bytes.NewBufferString("{invalid json}"))
		req.Header.Set("Content-Type", "application/json")

		// Create response recorder
		w := httptest.NewRecorder()

		// Perform request
		router.ServeHTTP(w, req)

		// Assert response
		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	// Test service error
	t.Run("ServiceError", func(t *testing.T) {
		// Prepare test data
		reqBody := CreateDepartmentRequest{
			Name:        "Engineering",
			Description: "Engineering Department",
			TeamID:      "team_123",
		}

		// Mock service response
		mockService.On("CreateDepartment", mock.Anything, mock.MatchedBy(func(req *service.CreateDepartmentRequest) bool {
			return req.Name == "Engineering"
		})).Return(testutils.NewError("internal error")).Once()

		// Create request
		jsonValue, _ := json.Marshal(reqBody)
		req, _ := http.NewRequest(http.MethodPost, "/departments", bytes.NewBuffer(jsonValue))
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

// TestUpdateDepartment tests the UpdateDepartment handler
func TestUpdateDepartment(t *testing.T) {
	// Set up test environment
	gin.SetMode(gin.TestMode)

	// Create mock service
	mockService := new(MockDepartmentService)

	// Create handler with mock service
	handler := &DepartmentHandler{
		departmentService: mockService,
	}

	// Create test router with route parameter
	router := gin.New()
	router.PUT("/departments/:id", handler.UpdateDepartment)

	// Test successful update
	t.Run("SuccessfulUpdate", func(t *testing.T) {
		// Prepare test data
		deptID := "dept_123"
		reqBody := UpdateDepartmentRequest{
			Name:        "Updated Engineering",
			Description: "Updated Engineering Department",
			ParentID:    "dept_456",
		}

		// Mock service response
		mockService.On("UpdateDepartment", mock.Anything, deptID, mock.MatchedBy(func(req *service.UpdateDepartmentRequest) bool {
			return req.Name == "Updated Engineering"
		})).Return(nil).Once()

		// Create request
		jsonValue, _ := json.Marshal(reqBody)
		req, _ := http.NewRequest(http.MethodPut, "/departments/"+deptID, bytes.NewBuffer(jsonValue))
		req.Header.Set("Content-Type", "application/json")

		// Create response recorder
		w := httptest.NewRecorder()

		// Perform request
		router.ServeHTTP(w, req)

		// Assert response
		assert.Equal(t, http.StatusOK, w.Code)
		assert.Contains(t, w.Body.String(), "department updated successfully")

		// Assert mock expectations
		mockService.AssertExpectations(t)
	})

	// Test missing department ID
	t.Run("MissingDepartmentID", func(t *testing.T) {
		// Create request without department ID
		reqBody := UpdateDepartmentRequest{
			Name: "Updated Engineering",
		}

		jsonValue, _ := json.Marshal(reqBody)
		req, _ := http.NewRequest(http.MethodPut, "/departments/", bytes.NewBuffer(jsonValue))
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
		req, _ := http.NewRequest(http.MethodPut, "/departments/dept_123", bytes.NewBufferString("{invalid json}"))
		req.Header.Set("Content-Type", "application/json")

		// Create response recorder
		w := httptest.NewRecorder()

		// Perform request
		router.ServeHTTP(w, req)

		// Assert response
		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	// Test service error - department not found
	t.Run("DepartmentNotFound", func(t *testing.T) {
		// Prepare test data
		deptID := "dept_456"
		reqBody := UpdateDepartmentRequest{
			Name: "Updated Engineering",
		}

		// Mock service response
		mockService.On("UpdateDepartment", mock.Anything, deptID, mock.MatchedBy(func(req *service.UpdateDepartmentRequest) bool {
			return req.Name == "Updated Engineering"
		})).Return(testutils.NewError("department not found")).Once()

		// Create request
		jsonValue, _ := json.Marshal(reqBody)
		req, _ := http.NewRequest(http.MethodPut, "/departments/"+deptID, bytes.NewBuffer(jsonValue))
		req.Header.Set("Content-Type", "application/json")

		// Create response recorder
		w := httptest.NewRecorder()

		// Perform request
		router.ServeHTTP(w, req)

		// Assert response
		assert.Equal(t, http.StatusNotFound, w.Code)
		assert.Contains(t, w.Body.String(), "department not found")

		// Assert mock expectations
		mockService.AssertExpectations(t)
	})

	// Test service error - parent department not found
	t.Run("ParentDepartmentNotFound", func(t *testing.T) {
		// Prepare test data
		deptID := "dept_123"
		reqBody := UpdateDepartmentRequest{
			Name:     "Updated Engineering",
			ParentID: "dept_789",
		}

		// Mock service response
		mockService.On("UpdateDepartment", mock.Anything, deptID, mock.MatchedBy(func(req *service.UpdateDepartmentRequest) bool {
			return req.Name == "Updated Engineering" && req.ParentID == "dept_789"
		})).Return(testutils.NewError("parent department not found")).Once()

		// Create request
		jsonValue, _ := json.Marshal(reqBody)
		req, _ := http.NewRequest(http.MethodPut, "/departments/"+deptID, bytes.NewBuffer(jsonValue))
		req.Header.Set("Content-Type", "application/json")

		// Create response recorder
		w := httptest.NewRecorder()

		// Perform request
		router.ServeHTTP(w, req)

		// Assert response
		assert.Equal(t, http.StatusNotFound, w.Code)
		assert.Contains(t, w.Body.String(), "parent department not found")

		// Assert mock expectations
		mockService.AssertExpectations(t)
	})

	// Test service error - internal error
	t.Run("InternalServerError", func(t *testing.T) {
		// Prepare test data
		deptID := "dept_123"
		reqBody := UpdateDepartmentRequest{
			Name: "Updated Engineering",
		}

		// Mock service response
		mockService.On("UpdateDepartment", mock.Anything, deptID, mock.MatchedBy(func(req *service.UpdateDepartmentRequest) bool {
			return req.Name == "Updated Engineering"
		})).Return(testutils.NewError("internal error")).Once()

		// Create request
		jsonValue, _ := json.Marshal(reqBody)
		req, _ := http.NewRequest(http.MethodPut, "/departments/"+deptID, bytes.NewBuffer(jsonValue))
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

// TestDeleteDepartment tests the DeleteDepartment handler
func TestDeleteDepartment(t *testing.T) {
	// Set up test environment
	gin.SetMode(gin.TestMode)

	// Create mock service
	mockService := new(MockDepartmentService)

	// Create handler with mock service
	handler := &DepartmentHandler{
		departmentService: mockService,
	}

	// Create test router with route parameter
	router := gin.New()
	router.DELETE("/departments/:id", handler.DeleteDepartment)

	// Test successful deletion
	t.Run("SuccessfulDeletion", func(t *testing.T) {
		// Prepare test data
		deptID := "dept_123"

		// Mock service response
		mockService.On("DeleteDepartment", mock.Anything, deptID).Return(nil).Once()

		// Create request
		req, _ := http.NewRequest(http.MethodDelete, "/departments/"+deptID, nil)

		// Create response recorder
		w := httptest.NewRecorder()

		// Perform request
		router.ServeHTTP(w, req)

		// Assert response
		assert.Equal(t, http.StatusOK, w.Code)
		assert.Contains(t, w.Body.String(), "department deleted successfully")

		// Assert mock expectations
		mockService.AssertExpectations(t)
	})

	// Test missing department ID
	t.Run("MissingDepartmentID", func(t *testing.T) {
		// Create request without department ID
		req, _ := http.NewRequest(http.MethodDelete, "/departments/", nil)

		// Create response recorder
		w := httptest.NewRecorder()

		// Perform request
		router.ServeHTTP(w, req)

		// Assert response
		assert.Equal(t, http.StatusNotFound, w.Code) // Changed to StatusNotFound since the route doesn't match
	})

	// Test service error - department not found
	t.Run("DepartmentNotFound", func(t *testing.T) {
		// Prepare test data
		deptID := "dept_456"

		// Mock service response
		mockService.On("DeleteDepartment", mock.Anything, deptID).Return(testutils.NewError("department not found")).Once()

		// Create request
		req, _ := http.NewRequest(http.MethodDelete, "/departments/"+deptID, nil)

		// Create response recorder
		w := httptest.NewRecorder()

		// Perform request
		router.ServeHTTP(w, req)

		// Assert response
		assert.Equal(t, http.StatusNotFound, w.Code)
		assert.Contains(t, w.Body.String(), "department not found")

		// Assert mock expectations
		mockService.AssertExpectations(t)
	})

	// Test service error - cannot delete department with child departments
	t.Run("CannotDeleteWithChildDepartments", func(t *testing.T) {
		// Prepare test data
		deptID := "dept_123"

		// Mock service response
		mockService.On("DeleteDepartment", mock.Anything, deptID).Return(testutils.NewError("cannot delete department with child departments")).Once()

		// Create request
		req, _ := http.NewRequest(http.MethodDelete, "/departments/"+deptID, nil)

		// Create response recorder
		w := httptest.NewRecorder()

		// Perform request
		router.ServeHTTP(w, req)

		// Assert response
		assert.Equal(t, http.StatusBadRequest, w.Code)
		assert.Contains(t, w.Body.String(), "cannot delete department with child departments")

		// Assert mock expectations
		mockService.AssertExpectations(t)
	})

	// Test service error - cannot delete department with employees
	t.Run("CannotDeleteWithEmployees", func(t *testing.T) {
		// Prepare test data
		deptID := "dept_123"

		// Mock service response
		mockService.On("DeleteDepartment", mock.Anything, deptID).Return(testutils.NewError("cannot delete department with employees")).Once()

		// Create request
		req, _ := http.NewRequest(http.MethodDelete, "/departments/"+deptID, nil)

		// Create response recorder
		w := httptest.NewRecorder()

		// Perform request
		router.ServeHTTP(w, req)

		// Assert response
		assert.Equal(t, http.StatusBadRequest, w.Code)
		assert.Contains(t, w.Body.String(), "cannot delete department with employees")

		// Assert mock expectations
		mockService.AssertExpectations(t)
	})

	// Test service error - internal error
	t.Run("InternalServerError", func(t *testing.T) {
		// Prepare test data
		deptID := "dept_123"

		// Mock service response
		mockService.On("DeleteDepartment", mock.Anything, deptID).Return(testutils.NewError("internal error")).Once()

		// Create request
		req, _ := http.NewRequest(http.MethodDelete, "/departments/"+deptID, nil)

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

// TestListDepartments tests the ListDepartments handler
func TestListDepartments(t *testing.T) {
	// Set up test environment
	gin.SetMode(gin.TestMode)

	// Create mock service
	mockService := new(MockDepartmentService)

	// Create handler with mock service
	handler := &DepartmentHandler{
		departmentService: mockService,
	}

	// Create test router
	router := gin.New()
	router.GET("/departments", handler.ListDepartments)

	// Test successful listing
	t.Run("SuccessfulListing", func(t *testing.T) {
		// Prepare test data
		teamID := "team_123"
		expectedDepartments := []*domain.Department{
			{
				ID:          "dept_123",
				Name:        "Engineering",
				Description: "Engineering Department",
				TeamID:      teamID,
				ParentID:    "",
			},
			{
				ID:          "dept_456",
				Name:        "Marketing",
				Description: "Marketing Department",
				TeamID:      teamID,
				ParentID:    "",
			},
		}

		// Mock service response
		mockService.On("ListDepartments", mock.Anything, teamID).Return(expectedDepartments, nil).Once()

		// Create request
		req, _ := http.NewRequest(http.MethodGet, "/departments?team_id="+teamID, nil)

		// Create response recorder
		w := httptest.NewRecorder()

		// Perform request
		router.ServeHTTP(w, req)

		// Assert response
		assert.Equal(t, http.StatusOK, w.Code)

		// Parse response
		var response []*domain.Department
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.Equal(t, expectedDepartments, response)

		// Assert mock expectations
		mockService.AssertExpectations(t)
	})

	// Test missing team ID
	t.Run("MissingTeamID", func(t *testing.T) {
		// Create request without team ID
		req, _ := http.NewRequest(http.MethodGet, "/departments", nil)

		// Create response recorder
		w := httptest.NewRecorder()

		// Perform request
		router.ServeHTTP(w, req)

		// Assert response
		assert.Equal(t, http.StatusBadRequest, w.Code)
		assert.Contains(t, w.Body.String(), "team id is required")
	})

	// Test service error
	t.Run("ServiceError", func(t *testing.T) {
		// Prepare test data
		teamID := "team_123"

		// Mock service response
		mockService.On("ListDepartments", mock.Anything, teamID).Return([]*domain.Department(nil), testutils.NewError("internal error")).Once()

		// Create request
		req, _ := http.NewRequest(http.MethodGet, "/departments?team_id="+teamID, nil)

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

// TestGetDepartment tests the GetDepartment handler
func TestGetDepartment(t *testing.T) {
	// Set up test environment
	gin.SetMode(gin.TestMode)

	// Create mock service
	mockService := new(MockDepartmentService)

	// Create handler with mock service
	handler := &DepartmentHandler{
		departmentService: mockService,
	}

	// Create test router with route parameter
	router := gin.New()
	router.GET("/departments/:id", handler.GetDepartment)

	// Test successful retrieval
	t.Run("SuccessfulRetrieval", func(t *testing.T) {
		// Prepare test data
		deptID := "dept_123"
		expectedDepartment := &domain.Department{
			ID:          deptID,
			Name:        "Engineering",
			Description: "Engineering Department",
			TeamID:      "team_123",
			ParentID:    "",
		}

		// Mock service response
		mockService.On("GetDepartment", mock.Anything, deptID).Return(expectedDepartment, nil).Once()

		// Create request
		req, _ := http.NewRequest(http.MethodGet, "/departments/"+deptID, nil)

		// Create response recorder
		w := httptest.NewRecorder()

		// Perform request
		router.ServeHTTP(w, req)

		// Assert response
		assert.Equal(t, http.StatusOK, w.Code)

		// Parse response
		var response domain.Department
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.Equal(t, *expectedDepartment, response)

		// Assert mock expectations
		mockService.AssertExpectations(t)
	})

	// Test missing department ID
	t.Run("MissingDepartmentID", func(t *testing.T) {
		// Create request without department ID
		req, _ := http.NewRequest(http.MethodGet, "/departments/", nil)

		// Create response recorder
		w := httptest.NewRecorder()

		// Perform request
		router.ServeHTTP(w, req)

		// Assert response
		assert.Equal(t, http.StatusNotFound, w.Code) // Changed to StatusNotFound since the route doesn't match
	})

	// Test service error - department not found
	t.Run("DepartmentNotFound", func(t *testing.T) {
		// Prepare test data
		deptID := "dept_456"

		// Mock service response
		mockService.On("GetDepartment", mock.Anything, deptID).Return((*domain.Department)(nil), testutils.NewError("department not found")).Once()

		// Create request
		req, _ := http.NewRequest(http.MethodGet, "/departments/"+deptID, nil)

		// Create response recorder
		w := httptest.NewRecorder()

		// Perform request
		router.ServeHTTP(w, req)

		// Assert response
		assert.Equal(t, http.StatusNotFound, w.Code)
		assert.Contains(t, w.Body.String(), "department not found")

		// Assert mock expectations
		mockService.AssertExpectations(t)
	})

	// Test service error - internal error
	t.Run("InternalServerError", func(t *testing.T) {
		// Prepare test data
		deptID := "dept_123"

		// Mock service response
		mockService.On("GetDepartment", mock.Anything, deptID).Return((*domain.Department)(nil), testutils.NewError("internal error")).Once()

		// Create request
		req, _ := http.NewRequest(http.MethodGet, "/departments/"+deptID, nil)

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