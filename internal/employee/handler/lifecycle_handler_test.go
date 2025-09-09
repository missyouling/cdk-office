package handler

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"cdk-office/internal/employee/service"
	"cdk-office/internal/shared/testutils"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockLifecycleService is a mock implementation of LifecycleServiceInterface
type MockLifecycleService struct {
	mock.Mock
}

func (m *MockLifecycleService) PromoteEmployee(ctx context.Context, employeeID, newPosition string) error {
	args := m.Called(ctx, employeeID, newPosition)
	return args.Error(0)
}

func (m *MockLifecycleService) TransferEmployee(ctx context.Context, employeeID, newDeptID string) error {
	args := m.Called(ctx, employeeID, newDeptID)
	return args.Error(0)
}

func (m *MockLifecycleService) TerminateEmployee(ctx context.Context, employeeID string, terminationDate time.Time, reason string) error {
	args := m.Called(ctx, employeeID, terminationDate, reason)
	return args.Error(0)
}

func (m *MockLifecycleService) GetEmployeeLifecycleHistory(ctx context.Context, employeeID string) ([]*service.EmployeeLifecycleEvent, error) {
	args := m.Called(ctx, employeeID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*service.EmployeeLifecycleEvent), args.Error(1)
}

// TestNewLifecycleHandler tests the NewLifecycleHandler function
func TestNewLifecycleHandler(t *testing.T) {
	handler := NewLifecycleHandler()
	assert.NotNil(t, handler)
	assert.NotNil(t, handler.lifecycleService)
}

// TestPromoteEmployee tests the PromoteEmployee handler
func TestPromoteEmployee(t *testing.T) {
	// Set up test environment
	gin.SetMode(gin.TestMode)

	// Create mock service
	mockService := new(MockLifecycleService)

	// Create handler with mock service
	handler := &LifecycleHandler{
		lifecycleService: mockService,
	}

	// Create test router
	router := gin.New()
	router.POST("/lifecycle/promote", handler.PromoteEmployee)

	// Test successful promotion
	t.Run("SuccessfulPromotion", func(t *testing.T) {
		// Prepare test data
		reqBody := PromoteEmployeeRequest{
			EmployeeID:  "emp_123",
			NewPosition: "Senior Engineer",
		}

		// Mock service response
		mockService.On("PromoteEmployee", mock.Anything, "emp_123", "Senior Engineer").Return(nil).Once()

		// Create request
		jsonValue, _ := json.Marshal(reqBody)
		req, _ := http.NewRequest(http.MethodPost, "/lifecycle/promote", bytes.NewBuffer(jsonValue))
		req.Header.Set("Content-Type", "application/json")

		// Create response recorder
		w := httptest.NewRecorder()

		// Perform request
		router.ServeHTTP(w, req)

		// Assert response
		assert.Equal(t, http.StatusOK, w.Code)
		assert.Contains(t, w.Body.String(), "employee promoted successfully")

		// Assert mock expectations
		mockService.AssertExpectations(t)
	})

	// Test invalid request body
	t.Run("InvalidRequestBody", func(t *testing.T) {
		// Create request with invalid JSON
		req, _ := http.NewRequest(http.MethodPost, "/lifecycle/promote", bytes.NewBufferString("{invalid json}"))
		req.Header.Set("Content-Type", "application/json")

		// Create response recorder
		w := httptest.NewRecorder()

		// Perform request
		router.ServeHTTP(w, req)

		// Assert response
		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	// Test service error - employee not found
	t.Run("EmployeeNotFound", func(t *testing.T) {
		// Prepare test data
		reqBody := PromoteEmployeeRequest{
			EmployeeID:  "emp_456",
			NewPosition: "Senior Engineer",
		}

		// Mock service response
		mockService.On("PromoteEmployee", mock.Anything, "emp_456", "Senior Engineer").Return(testutils.NewError("employee not found")).Once()

		// Create request
		jsonValue, _ := json.Marshal(reqBody)
		req, _ := http.NewRequest(http.MethodPost, "/lifecycle/promote", bytes.NewBuffer(jsonValue))
		req.Header.Set("Content-Type", "application/json")

		// Create response recorder
		w := httptest.NewRecorder()

		// Perform request
		router.ServeHTTP(w, req)

		// Assert response
		assert.Equal(t, http.StatusNotFound, w.Code)
		assert.Contains(t, w.Body.String(), "employee not found")

		// Assert mock expectations
		mockService.AssertExpectations(t)
	})

	// Test service error - internal error
	t.Run("InternalServerError", func(t *testing.T) {
		// Prepare test data
		reqBody := PromoteEmployeeRequest{
			EmployeeID:  "emp_123",
			NewPosition: "Senior Engineer",
		}

		// Mock service response
		mockService.On("PromoteEmployee", mock.Anything, "emp_123", "Senior Engineer").Return(testutils.NewError("internal error")).Once()

		// Create request
		jsonValue, _ := json.Marshal(reqBody)
		req, _ := http.NewRequest(http.MethodPost, "/lifecycle/promote", bytes.NewBuffer(jsonValue))
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

// TestTransferEmployee tests the TransferEmployee handler
func TestTransferEmployee(t *testing.T) {
	// Set up test environment
	gin.SetMode(gin.TestMode)

	// Create mock service
	mockService := new(MockLifecycleService)

	// Create handler with mock service
	handler := &LifecycleHandler{
		lifecycleService: mockService,
	}

	// Create test router
	router := gin.New()
	router.POST("/lifecycle/transfer", handler.TransferEmployee)

	// Test successful transfer
	t.Run("SuccessfulTransfer", func(t *testing.T) {
		// Prepare test data
		reqBody := TransferEmployeeRequest{
			EmployeeID: "emp_123",
			NewDeptID:  "dept_456",
		}

		// Mock service response
		mockService.On("TransferEmployee", mock.Anything, "emp_123", "dept_456").Return(nil).Once()

		// Create request
		jsonValue, _ := json.Marshal(reqBody)
		req, _ := http.NewRequest(http.MethodPost, "/lifecycle/transfer", bytes.NewBuffer(jsonValue))
		req.Header.Set("Content-Type", "application/json")

		// Create response recorder
		w := httptest.NewRecorder()

		// Perform request
		router.ServeHTTP(w, req)

		// Assert response
		assert.Equal(t, http.StatusOK, w.Code)
		assert.Contains(t, w.Body.String(), "employee transferred successfully")

		// Assert mock expectations
		mockService.AssertExpectations(t)
	})

	// Test invalid request body
	t.Run("InvalidRequestBody", func(t *testing.T) {
		// Create request with invalid JSON
		req, _ := http.NewRequest(http.MethodPost, "/lifecycle/transfer", bytes.NewBufferString("{invalid json}"))
		req.Header.Set("Content-Type", "application/json")

		// Create response recorder
		w := httptest.NewRecorder()

		// Perform request
		router.ServeHTTP(w, req)

		// Assert response
		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	// Test service error - employee not found
	t.Run("EmployeeNotFound", func(t *testing.T) {
		// Prepare test data
		reqBody := TransferEmployeeRequest{
			EmployeeID: "emp_456",
			NewDeptID:  "dept_456",
		}

		// Mock service response
		mockService.On("TransferEmployee", mock.Anything, "emp_456", "dept_456").Return(testutils.NewError("employee not found")).Once()

		// Create request
		jsonValue, _ := json.Marshal(reqBody)
		req, _ := http.NewRequest(http.MethodPost, "/lifecycle/transfer", bytes.NewBuffer(jsonValue))
		req.Header.Set("Content-Type", "application/json")

		// Create response recorder
		w := httptest.NewRecorder()

		// Perform request
		router.ServeHTTP(w, req)

		// Assert response
		assert.Equal(t, http.StatusNotFound, w.Code)
		assert.Contains(t, w.Body.String(), "employee not found")

		// Assert mock expectations
		mockService.AssertExpectations(t)
	})

	// Test service error - department not found
	t.Run("DepartmentNotFound", func(t *testing.T) {
		// Prepare test data
		reqBody := TransferEmployeeRequest{
			EmployeeID: "emp_123",
			NewDeptID:  "dept_789",
		}

		// Mock service response
		mockService.On("TransferEmployee", mock.Anything, "emp_123", "dept_789").Return(testutils.NewError("department not found")).Once()

		// Create request
		jsonValue, _ := json.Marshal(reqBody)
		req, _ := http.NewRequest(http.MethodPost, "/lifecycle/transfer", bytes.NewBuffer(jsonValue))
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

	// Test service error - internal error
	t.Run("InternalServerError", func(t *testing.T) {
		// Prepare test data
		reqBody := TransferEmployeeRequest{
			EmployeeID: "emp_123",
			NewDeptID:  "dept_456",
		}

		// Mock service response
		mockService.On("TransferEmployee", mock.Anything, "emp_123", "dept_456").Return(testutils.NewError("internal error")).Once()

		// Create request
		jsonValue, _ := json.Marshal(reqBody)
		req, _ := http.NewRequest(http.MethodPost, "/lifecycle/transfer", bytes.NewBuffer(jsonValue))
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

// TestTerminateEmployee tests the TerminateEmployee handler
func TestTerminateEmployee(t *testing.T) {
	// Set up test environment
	gin.SetMode(gin.TestMode)

	// Create mock service
	mockService := new(MockLifecycleService)

	// Create handler with mock service
	handler := &LifecycleHandler{
		lifecycleService: mockService,
	}

	// Create test router
	router := gin.New()
	router.POST("/lifecycle/terminate", handler.TerminateEmployee)

	// Test successful termination
	t.Run("SuccessfulTermination", func(t *testing.T) {
		// Prepare test data
		terminationDate := "2023-12-31"
		reqBody := TerminateEmployeeRequest{
			EmployeeID:      "emp_123",
			TerminationDate: terminationDate,
			Reason:          "Resignation",
		}

		// Parse the date for the mock
		parsedDate, _ := time.Parse("2006-01-02", terminationDate)

		// Mock service response
		mockService.On("TerminateEmployee", mock.Anything, "emp_123", parsedDate, "Resignation").Return(nil).Once()

		// Create request
		jsonValue, _ := json.Marshal(reqBody)
		req, _ := http.NewRequest(http.MethodPost, "/lifecycle/terminate", bytes.NewBuffer(jsonValue))
		req.Header.Set("Content-Type", "application/json")

		// Create response recorder
		w := httptest.NewRecorder()

		// Perform request
		router.ServeHTTP(w, req)

		// Assert response
		assert.Equal(t, http.StatusOK, w.Code)
		assert.Contains(t, w.Body.String(), "employee terminated successfully")

		// Assert mock expectations
		mockService.AssertExpectations(t)
	})

	// Test invalid request body
	t.Run("InvalidRequestBody", func(t *testing.T) {
		// Create request with invalid JSON
		req, _ := http.NewRequest(http.MethodPost, "/lifecycle/terminate", bytes.NewBufferString("{invalid json}"))
		req.Header.Set("Content-Type", "application/json")

		// Create response recorder
		w := httptest.NewRecorder()

		// Perform request
		router.ServeHTTP(w, req)

		// Assert response
		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	// Test invalid date format
	t.Run("InvalidDateFormat", func(t *testing.T) {
		// Prepare test data with invalid date format
		reqBody := TerminateEmployeeRequest{
			EmployeeID:      "emp_123",
			TerminationDate: "invalid-date",
			Reason:          "Resignation",
		}

		// Create request
		jsonValue, _ := json.Marshal(reqBody)
		req, _ := http.NewRequest(http.MethodPost, "/lifecycle/terminate", bytes.NewBuffer(jsonValue))
		req.Header.Set("Content-Type", "application/json")

		// Create response recorder
		w := httptest.NewRecorder()

		// Perform request
		router.ServeHTTP(w, req)

		// Assert response
		assert.Equal(t, http.StatusBadRequest, w.Code)
		assert.Contains(t, w.Body.String(), "invalid termination_date format")
	})

	// Test service error - employee not found
	t.Run("EmployeeNotFound", func(t *testing.T) {
		// Prepare test data
		terminationDate := "2023-12-31"
		reqBody := TerminateEmployeeRequest{
			EmployeeID:      "emp_456",
			TerminationDate: terminationDate,
			Reason:          "Resignation",
		}

		// Parse the date for the mock
		parsedDate, _ := time.Parse("2006-01-02", terminationDate)

		// Mock service response
		mockService.On("TerminateEmployee", mock.Anything, "emp_456", parsedDate, "Resignation").Return(testutils.NewError("employee not found")).Once()

		// Create request
		jsonValue, _ := json.Marshal(reqBody)
		req, _ := http.NewRequest(http.MethodPost, "/lifecycle/terminate", bytes.NewBuffer(jsonValue))
		req.Header.Set("Content-Type", "application/json")

		// Create response recorder
		w := httptest.NewRecorder()

		// Perform request
		router.ServeHTTP(w, req)

		// Assert response
		assert.Equal(t, http.StatusNotFound, w.Code)
		assert.Contains(t, w.Body.String(), "employee not found")

		// Assert mock expectations
		mockService.AssertExpectations(t)
	})

	// Test service error - internal error
	t.Run("InternalServerError", func(t *testing.T) {
		// Prepare test data
		terminationDate := "2023-12-31"
		reqBody := TerminateEmployeeRequest{
			EmployeeID:      "emp_123",
			TerminationDate: terminationDate,
			Reason:          "Resignation",
		}

		// Parse the date for the mock
		parsedDate, _ := time.Parse("2006-01-02", terminationDate)

		// Mock service response
		mockService.On("TerminateEmployee", mock.Anything, "emp_123", parsedDate, "Resignation").Return(testutils.NewError("internal error")).Once()

		// Create request
		jsonValue, _ := json.Marshal(reqBody)
		req, _ := http.NewRequest(http.MethodPost, "/lifecycle/terminate", bytes.NewBuffer(jsonValue))
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

// TestGetEmployeeLifecycleHistory tests the GetEmployeeLifecycleHistory handler
func TestGetEmployeeLifecycleHistory(t *testing.T) {
	// Set up test environment
	gin.SetMode(gin.TestMode)

	// Create mock service
	mockService := new(MockLifecycleService)

	// Create handler with mock service
	handler := &LifecycleHandler{
		lifecycleService: mockService,
	}

	// Create test router with route parameter
	router := gin.New()
	router.GET("/lifecycle/history/:id", handler.GetEmployeeLifecycleHistory)

	// Test successful retrieval
	t.Run("SuccessfulRetrieval", func(t *testing.T) {
		// Prepare test data
		empID := "emp_123"
		expectedEvents := []*service.EmployeeLifecycleEvent{
			{
				ID:            "event_123",
				EmployeeID:    empID,
				EventType:     "promotion",
				OldValue:      "Engineer",
				NewValue:      "Senior Engineer",
				EffectiveDate: time.Now(),
				Reason:        "Promoted to Senior Engineer",
				CreatedAt:     time.Now(),
			},
			{
				ID:            "event_456",
				EmployeeID:    empID,
				EventType:     "transfer",
				OldValue:      "dept_123",
				NewValue:      "dept_456",
				EffectiveDate: time.Now(),
				Reason:        "Transferred to Engineering Department",
				CreatedAt:     time.Now(),
			},
		}

		// Mock service response
		mockService.On("GetEmployeeLifecycleHistory", mock.Anything, empID).Return(expectedEvents, nil).Once()

		// Create request
		req, _ := http.NewRequest(http.MethodGet, "/lifecycle/history/"+empID, nil)

		// Create response recorder
		w := httptest.NewRecorder()

		// Perform request
		router.ServeHTTP(w, req)

		// Assert response
		assert.Equal(t, http.StatusOK, w.Code)

		// Parse response
		var response []*service.EmployeeLifecycleEvent
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)
		// Since we're comparing time fields, we'll just check the length and some fields
		assert.Equal(t, len(expectedEvents), len(response))
		assert.Equal(t, expectedEvents[0].EventType, response[0].EventType)
		assert.Equal(t, expectedEvents[1].EventType, response[1].EventType)

		// Assert mock expectations
		mockService.AssertExpectations(t)
	})

	// Test missing employee ID
	t.Run("MissingEmployeeID", func(t *testing.T) {
		// Create request without employee ID
		req, _ := http.NewRequest(http.MethodGet, "/lifecycle/history/", nil)

		// Create response recorder
		w := httptest.NewRecorder()

		// Perform request
		router.ServeHTTP(w, req)

		// Assert response
		assert.Equal(t, http.StatusNotFound, w.Code) // Changed to StatusNotFound since the route doesn't match
	})

	// Test service error - employee not found
	t.Run("EmployeeNotFound", func(t *testing.T) {
		// Prepare test data
		empID := "emp_456"

		// Mock service response
		mockService.On("GetEmployeeLifecycleHistory", mock.Anything, empID).Return([]*service.EmployeeLifecycleEvent(nil), testutils.NewError("employee not found")).Once()

		// Create request
		req, _ := http.NewRequest(http.MethodGet, "/lifecycle/history/"+empID, nil)

		// Create response recorder
		w := httptest.NewRecorder()

		// Perform request
		router.ServeHTTP(w, req)

		// Assert response
		assert.Equal(t, http.StatusNotFound, w.Code)
		assert.Contains(t, w.Body.String(), "employee not found")

		// Assert mock expectations
		mockService.AssertExpectations(t)
	})

	// Test service error - internal error
	t.Run("InternalServerError", func(t *testing.T) {
		// Prepare test data
		empID := "emp_123"

		// Mock service response
		mockService.On("GetEmployeeLifecycleHistory", mock.Anything, empID).Return([]*service.EmployeeLifecycleEvent(nil), testutils.NewError("internal error")).Once()

		// Create request
		req, _ := http.NewRequest(http.MethodGet, "/lifecycle/history/"+empID, nil)

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