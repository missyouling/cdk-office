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

// MockBatchQRCodeService is a mock implementation of BatchQRCodeServiceInterface
type MockBatchQRCodeService struct {
	mock.Mock
}

func (m *MockBatchQRCodeService) CreateBatchQRCode(ctx context.Context, req *service.CreateBatchQRCodeRequest) (*service.BatchQRCode, error) {
	args := m.Called(ctx, req)
	return args.Get(0).(*service.BatchQRCode), args.Error(1)
}

func (m *MockBatchQRCodeService) UpdateBatchQRCode(ctx context.Context, batchID string, req *service.UpdateBatchQRCodeRequest) error {
	args := m.Called(ctx, batchID, req)
	return args.Error(0)
}

func (m *MockBatchQRCodeService) DeleteBatchQRCode(ctx context.Context, batchID string) error {
	args := m.Called(ctx, batchID)
	return args.Error(0)
}

func (m *MockBatchQRCodeService) ListBatchQRCodes(ctx context.Context, appID string, page, size int) ([]*service.BatchQRCode, int64, error) {
	args := m.Called(ctx, appID, page, size)
	return args.Get(0).([]*service.BatchQRCode), args.Get(1).(int64), args.Error(2)
}

func (m *MockBatchQRCodeService) GetBatchQRCode(ctx context.Context, batchID string) (*service.BatchQRCode, error) {
	args := m.Called(ctx, batchID)
	return args.Get(0).(*service.BatchQRCode), args.Error(1)
}

func (m *MockBatchQRCodeService) GenerateBatchQRCodes(ctx context.Context, batchID string) ([]*domain.QRCode, error) {
	args := m.Called(ctx, batchID)
	return args.Get(0).([]*domain.QRCode), args.Error(1)
}

// TestNewBatchQRCodeHandler tests the NewBatchQRCodeHandler function
func TestNewBatchQRCodeHandler(t *testing.T) {
	handler := NewBatchQRCodeHandler()
	assert.NotNil(t, handler)
	assert.NotNil(t, handler.batchService)
}

// TestCreateBatchQRCode tests the CreateBatchQRCode handler
func TestCreateBatchQRCode(t *testing.T) {
	// Set up test environment
	gin.SetMode(gin.TestMode)

	// Create mock service
	mockService := new(MockBatchQRCodeService)

	// Create handler with mock service
	handler := &BatchQRCodeHandler{
		batchService: mockService,
	}

	// Create test router
	router := gin.New()
	router.POST("/batches", handler.CreateBatchQRCode)

	// Test successful creation
	t.Run("SuccessfulCreation", func(t *testing.T) {
		// Prepare test data
		reqBody := CreateBatchQRCodeRequest{
			AppID:       "app_123",
			Name:        "Test Batch",
			Description: "Test batch description",
			Prefix:      "TB",
			Count:       10,
			Type:        "static",
			URLTemplate: "https://example.com/{index}",
			Config:      map[string]string{"color": "black"},
			CreatedBy:   "user_123",
		}

		// Mock service response
		expectedBatch := &service.BatchQRCode{
			ID:          "batch_123",
			AppID:       "app_123",
			Name:        "Test Batch",
			Description: "Test batch description",
			Prefix:      "TB",
			Count:       10,
			Type:        "static",
			URLTemplate: "https://example.com/{index}",
			Status:      "pending",
			CreatedBy:   "user_123",
			CreatedAt:   time.Now(),
			UpdatedAt:   time.Now(),
		}

		mockService.On("CreateBatchQRCode", mock.Anything, mock.MatchedBy(func(req *service.CreateBatchQRCodeRequest) bool {
			return req.AppID == reqBody.AppID && req.Name == reqBody.Name && req.Type == reqBody.Type
		})).Return(expectedBatch, nil).Once()

		// Create request
		jsonValue, _ := json.Marshal(reqBody)
		req, _ := http.NewRequest(http.MethodPost, "/batches", bytes.NewBuffer(jsonValue))
		req.Header.Set("Content-Type", "application/json")

		// Create response recorder
		w := httptest.NewRecorder()

		// Perform request
		router.ServeHTTP(w, req)

		// Assert response
		assert.Equal(t, http.StatusOK, w.Code)

		// Parse response
		var response service.BatchQRCode
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.Equal(t, expectedBatch.ID, response.ID)
		assert.Equal(t, expectedBatch.Name, response.Name)

		// Assert mock expectations
		mockService.AssertExpectations(t)
	})

	// Test invalid request body
	t.Run("InvalidRequestBody", func(t *testing.T) {
		// Create request with invalid JSON
		req, _ := http.NewRequest(http.MethodPost, "/batches", bytes.NewBufferString("{invalid json}"))
		req.Header.Set("Content-Type", "application/json")

		// Create response recorder
		w := httptest.NewRecorder()

		// Perform request
		router.ServeHTTP(w, req)

		// Assert response
		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	// Test service error - invalid QR code type
	t.Run("InvalidQRCodeType", func(t *testing.T) {
		// Prepare test data
		reqBody := CreateBatchQRCodeRequest{
			AppID:       "app_123",
			Name:        "Test Batch",
			Description: "Test batch description",
			Count:       10,
			Type:        "invalid",
			CreatedBy:   "user_123",
		}

		// Mock service response
		mockService.On("CreateBatchQRCode", mock.Anything, mock.MatchedBy(func(req *service.CreateBatchQRCodeRequest) bool {
			return req.Type == "invalid"
		})).Return((*service.BatchQRCode)(nil), testutils.NewError("invalid QR code type")).Once()

		// Create request
		jsonValue, _ := json.Marshal(reqBody)
		req, _ := http.NewRequest(http.MethodPost, "/batches", bytes.NewBuffer(jsonValue))
		req.Header.Set("Content-Type", "application/json")

		// Create response recorder
		w := httptest.NewRecorder()

		// Perform request
		router.ServeHTTP(w, req)

		// Assert response
		assert.Equal(t, http.StatusBadRequest, w.Code)
		assert.Contains(t, w.Body.String(), "invalid QR code type")

		// Assert mock expectations
		mockService.AssertExpectations(t)
	})

	// Test service error - internal error
	t.Run("InternalServerError", func(t *testing.T) {
		// Prepare test data
		reqBody := CreateBatchQRCodeRequest{
			AppID:       "app_123",
			Name:        "Test Batch",
			Description: "Test batch description",
			Count:       100, // Valid count
			Type:        "static",
			CreatedBy:   "user_123",
		}

		// Mock service response
		mockService.On("CreateBatchQRCode", mock.Anything, mock.MatchedBy(func(req *service.CreateBatchQRCodeRequest) bool {
			return req.Name == "Test Batch"
		})).Return((*service.BatchQRCode)(nil), testutils.NewError("internal error")).Once()

		// Create request
		jsonValue, _ := json.Marshal(reqBody)
		req, _ := http.NewRequest(http.MethodPost, "/batches", bytes.NewBuffer(jsonValue))
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

// TestUpdateBatchQRCode tests the UpdateBatchQRCode handler
func TestUpdateBatchQRCode(t *testing.T) {
	// Set up test environment
	gin.SetMode(gin.TestMode)

	// Create mock service
	mockService := new(MockBatchQRCodeService)

	// Create handler with mock service
	handler := &BatchQRCodeHandler{
		batchService: mockService,
	}

	// Create test router with route parameter
	router := gin.New()
	router.PUT("/batches/:id", handler.UpdateBatchQRCode)

	// Test successful update
	t.Run("SuccessfulUpdate", func(t *testing.T) {
		// Prepare test data
		batchID := "batch_123"
		reqBody := UpdateBatchQRCodeRequest{
			Name:        "Updated Batch",
			Description: "Updated batch description",
			Prefix:      "UB",
			URLTemplate: "https://updated-example.com/{index}",
			Config:      "{\"color\": \"blue\"}",
		}

		// Mock service response
		mockService.On("UpdateBatchQRCode", mock.Anything, batchID, mock.MatchedBy(func(req *service.UpdateBatchQRCodeRequest) bool {
			return req.Name == "Updated Batch" && req.Prefix == "UB"
		})).Return(nil).Once()

		// Create request
		jsonValue, _ := json.Marshal(reqBody)
		req, _ := http.NewRequest(http.MethodPut, "/batches/"+batchID, bytes.NewBuffer(jsonValue))
		req.Header.Set("Content-Type", "application/json")

		// Create response recorder
		w := httptest.NewRecorder()

		// Perform request
		router.ServeHTTP(w, req)

		// Assert response
		assert.Equal(t, http.StatusOK, w.Code)
		assert.Contains(t, w.Body.String(), "batch QR code updated successfully")

		// Assert mock expectations
		mockService.AssertExpectations(t)
	})

	// Test missing batch ID
	t.Run("MissingBatchID", func(t *testing.T) {
		// Create request without batch ID
		reqBody := UpdateBatchQRCodeRequest{
			Name: "Updated Batch",
		}

		jsonValue, _ := json.Marshal(reqBody)
		req, _ := http.NewRequest(http.MethodPut, "/batches/", bytes.NewBuffer(jsonValue))
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
		req, _ := http.NewRequest(http.MethodPut, "/batches/batch_123", bytes.NewBufferString("{invalid json}"))
		req.Header.Set("Content-Type", "application/json")

		// Create response recorder
		w := httptest.NewRecorder()

		// Perform request
		router.ServeHTTP(w, req)

		// Assert response
		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	// Test service error - batch not found
	t.Run("BatchNotFound", func(t *testing.T) {
		// Prepare test data
		batchID := "batch_456"
		reqBody := UpdateBatchQRCodeRequest{
			Name: "Updated Batch",
		}

		// Mock service response
		mockService.On("UpdateBatchQRCode", mock.Anything, batchID, mock.MatchedBy(func(req *service.UpdateBatchQRCodeRequest) bool {
			return req.Name == "Updated Batch"
		})).Return(testutils.NewError("batch QR code not found")).Once()

		// Create request
		jsonValue, _ := json.Marshal(reqBody)
		req, _ := http.NewRequest(http.MethodPut, "/batches/"+batchID, bytes.NewBuffer(jsonValue))
		req.Header.Set("Content-Type", "application/json")

		// Create response recorder
		w := httptest.NewRecorder()

		// Perform request
		router.ServeHTTP(w, req)

		// Assert response
		assert.Equal(t, http.StatusNotFound, w.Code)
		assert.Contains(t, w.Body.String(), "batch QR code not found")

		// Assert mock expectations
		mockService.AssertExpectations(t)
	})

	// Test service error - internal error
	t.Run("InternalServerError", func(t *testing.T) {
		// Prepare test data
		batchID := "batch_123"
		reqBody := UpdateBatchQRCodeRequest{
			Name: "Updated Batch",
		}

		// Mock service response
		mockService.On("UpdateBatchQRCode", mock.Anything, batchID, mock.MatchedBy(func(req *service.UpdateBatchQRCodeRequest) bool {
			return req.Name == "Updated Batch"
		})).Return(testutils.NewError("internal error")).Once()

		// Create request
		jsonValue, _ := json.Marshal(reqBody)
		req, _ := http.NewRequest(http.MethodPut, "/batches/"+batchID, bytes.NewBuffer(jsonValue))
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

// TestDeleteBatchQRCode tests the DeleteBatchQRCode handler
func TestDeleteBatchQRCode(t *testing.T) {
	// Set up test environment
	gin.SetMode(gin.TestMode)

	// Create mock service
	mockService := new(MockBatchQRCodeService)

	// Create handler with mock service
	handler := &BatchQRCodeHandler{
		batchService: mockService,
	}

	// Create test router with route parameter
	router := gin.New()
	router.DELETE("/batches/:id", handler.DeleteBatchQRCode)

	// Test successful deletion
	t.Run("SuccessfulDeletion", func(t *testing.T) {
		// Prepare test data
		batchID := "batch_123"

		// Mock service response
		mockService.On("DeleteBatchQRCode", mock.Anything, batchID).Return(nil).Once()

		// Create request
		req, _ := http.NewRequest(http.MethodDelete, "/batches/"+batchID, nil)

		// Create response recorder
		w := httptest.NewRecorder()

		// Perform request
		router.ServeHTTP(w, req)

		// Assert response
		assert.Equal(t, http.StatusOK, w.Code)
		assert.Contains(t, w.Body.String(), "batch QR code deleted successfully")

		// Assert mock expectations
		mockService.AssertExpectations(t)
	})

	// Test missing batch ID
	t.Run("MissingBatchID", func(t *testing.T) {
		// Create request without batch ID
		req, _ := http.NewRequest(http.MethodDelete, "/batches/", nil)

		// Create response recorder
		w := httptest.NewRecorder()

		// Perform request
		router.ServeHTTP(w, req)

		// Assert response
		assert.Equal(t, http.StatusNotFound, w.Code) // Changed to StatusNotFound since the route doesn't match
	})

	// Test service error - batch not found
	t.Run("BatchNotFound", func(t *testing.T) {
		// Prepare test data
		batchID := "batch_456"

		// Mock service response
		mockService.On("DeleteBatchQRCode", mock.Anything, batchID).Return(testutils.NewError("batch QR code not found")).Once()

		// Create request
		req, _ := http.NewRequest(http.MethodDelete, "/batches/"+batchID, nil)

		// Create response recorder
		w := httptest.NewRecorder()

		// Perform request
		router.ServeHTTP(w, req)

		// Assert response
		assert.Equal(t, http.StatusNotFound, w.Code)
		assert.Contains(t, w.Body.String(), "batch QR code not found")

		// Assert mock expectations
		mockService.AssertExpectations(t)
	})

	// Test service error - internal error
	t.Run("InternalServerError", func(t *testing.T) {
		// Prepare test data
		batchID := "batch_123"

		// Mock service response
		mockService.On("DeleteBatchQRCode", mock.Anything, batchID).Return(testutils.NewError("internal error")).Once()

		// Create request
		req, _ := http.NewRequest(http.MethodDelete, "/batches/"+batchID, nil)

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

// TestListBatchQRCodes tests the ListBatchQRCodes handler
func TestListBatchQRCodes(t *testing.T) {
	// Set up test environment
	gin.SetMode(gin.TestMode)

	// Create mock service
	mockService := new(MockBatchQRCodeService)

	// Create handler with mock service
	handler := &BatchQRCodeHandler{
		batchService: mockService,
	}

	// Create test router
	router := gin.New()
	router.GET("/batches", handler.ListBatchQRCodes)

	// Test successful listing
	t.Run("SuccessfulListing", func(t *testing.T) {
		// Prepare test data
		appID := "app_123"

		// Mock service response
		expectedBatches := []*service.BatchQRCode{
			{
				ID:          "batch_123",
				AppID:       appID,
				Name:        "Test Batch 1",
				Count:       10,
				Type:        "static",
				Status:      "completed",
				CreatedBy:   "user_123",
				CreatedAt:   time.Now(),
				UpdatedAt:   time.Now(),
			},
			{
				ID:          "batch_456",
				AppID:       appID,
				Name:        "Test Batch 2",
				Count:       20,
				Type:        "dynamic",
				Status:      "pending",
				CreatedBy:   "user_123",
				CreatedAt:   time.Now(),
				UpdatedAt:   time.Now(),
			},
		}

		mockService.On("ListBatchQRCodes", mock.Anything, appID, 1, 10).Return(expectedBatches, int64(2), nil).Once()

		// Create request
		req, _ := http.NewRequest(http.MethodGet, "/batches?app_id="+appID, nil)

		// Create response recorder
		w := httptest.NewRecorder()

		// Perform request
		router.ServeHTTP(w, req)

		// Assert response
		assert.Equal(t, http.StatusOK, w.Code)

		// Parse response
		var response ListBatchQRCodesResponse
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.Equal(t, int64(2), response.Total)
		assert.Equal(t, 2, len(response.Items))
		assert.Equal(t, "Test Batch 1", response.Items[0].Name)

		// Assert mock expectations
		mockService.AssertExpectations(t)
	})

	// Test missing app ID
	t.Run("MissingAppID", func(t *testing.T) {
		// Create request without app ID
		req, _ := http.NewRequest(http.MethodGet, "/batches", nil)

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
		mockService.On("ListBatchQRCodes", mock.Anything, appID, 1, 10).Return([]*service.BatchQRCode(nil), int64(0), testutils.NewError("internal error")).Once()

		// Create request
		req, _ := http.NewRequest(http.MethodGet, "/batches?app_id="+appID, nil)

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
		mockService.On("ListBatchQRCodes", mock.Anything, appID, 2, 5).Return([]*service.BatchQRCode{}, int64(0), nil).Once()

		// Create request with custom pagination
		req, _ := http.NewRequest(http.MethodGet, "/batches?app_id="+appID+"&page=2&size=5", nil)

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

// TestGetBatchQRCode tests the GetBatchQRCode handler
func TestGetBatchQRCode(t *testing.T) {
	// Set up test environment
	gin.SetMode(gin.TestMode)

	// Create mock service
	mockService := new(MockBatchQRCodeService)

	// Create handler with mock service
	handler := &BatchQRCodeHandler{
		batchService: mockService,
	}

	// Create test router with route parameter
	router := gin.New()
	router.GET("/batches/:id", handler.GetBatchQRCode)

	// Test successful retrieval
	t.Run("SuccessfulRetrieval", func(t *testing.T) {
		// Prepare test data
		batchID := "batch_123"

		// Mock service response
		expectedBatch := &service.BatchQRCode{
			ID:          batchID,
			AppID:       "app_123",
			Name:        "Test Batch",
			Count:       10,
			Type:        "static",
			Status:      "completed",
			CreatedBy:   "user_123",
			CreatedAt:   time.Now(),
			UpdatedAt:   time.Now(),
		}

		mockService.On("GetBatchQRCode", mock.Anything, batchID).Return(expectedBatch, nil).Once()

		// Create request
		req, _ := http.NewRequest(http.MethodGet, "/batches/"+batchID, nil)

		// Create response recorder
		w := httptest.NewRecorder()

		// Perform request
		router.ServeHTTP(w, req)

		// Assert response
		assert.Equal(t, http.StatusOK, w.Code)

		// Parse response
		var response service.BatchQRCode
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.Equal(t, expectedBatch.ID, response.ID)
		assert.Equal(t, expectedBatch.Name, response.Name)

		// Assert mock expectations
		mockService.AssertExpectations(t)
	})

	// Test missing batch ID
	t.Run("MissingBatchID", func(t *testing.T) {
		// Create request without batch ID
		req, _ := http.NewRequest(http.MethodGet, "/batches/", nil)

		// Create response recorder
		w := httptest.NewRecorder()

		// Perform request
		router.ServeHTTP(w, req)

		// Assert response
		assert.Equal(t, http.StatusNotFound, w.Code) // Changed to StatusNotFound since the route doesn't match
	})

	// Test service error - batch not found
	t.Run("BatchNotFound", func(t *testing.T) {
		// Prepare test data
		batchID := "batch_456"

		// Mock service response
		mockService.On("GetBatchQRCode", mock.Anything, batchID).Return((*service.BatchQRCode)(nil), testutils.NewError("batch QR code not found")).Once()

		// Create request
		req, _ := http.NewRequest(http.MethodGet, "/batches/"+batchID, nil)

		// Create response recorder
		w := httptest.NewRecorder()

		// Perform request
		router.ServeHTTP(w, req)

		// Assert response
		assert.Equal(t, http.StatusNotFound, w.Code)
		assert.Contains(t, w.Body.String(), "batch QR code not found")

		// Assert mock expectations
		mockService.AssertExpectations(t)
	})

	// Test service error - internal error
	t.Run("InternalServerError", func(t *testing.T) {
		// Prepare test data
		batchID := "batch_123"

		// Mock service response
		mockService.On("GetBatchQRCode", mock.Anything, batchID).Return((*service.BatchQRCode)(nil), testutils.NewError("internal error")).Once()

		// Create request
		req, _ := http.NewRequest(http.MethodGet, "/batches/"+batchID, nil)

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

// TestGenerateBatchQRCodes tests the GenerateBatchQRCodes handler
func TestGenerateBatchQRCodes(t *testing.T) {
	// Set up test environment
	gin.SetMode(gin.TestMode)

	// Create mock service
	mockService := new(MockBatchQRCodeService)

	// Create handler with mock service
	handler := &BatchQRCodeHandler{
		batchService: mockService,
	}

	// Create test router with route parameter
	router := gin.New()
	router.POST("/batches/:id/generate", handler.GenerateBatchQRCodes)

	// Test successful generation
	t.Run("SuccessfulGeneration", func(t *testing.T) {
		// Prepare test data
		batchID := "batch_123"

		// Mock service response
		expectedQRCodes := []*domain.QRCode{
			{
				ID:        "qr_123",
				AppID:     "app_123",
				Name:      "Test QR 1",
				Content:   "https://example.com/batch_123/1",
				Type:      "static",
				URL:       "https://example.com/batch_123/1",
				CreatedBy: "user_123",
				CreatedAt: time.Now(),
				UpdatedAt: time.Now(),
			},
			{
				ID:        "qr_456",
				AppID:     "app_123",
				Name:      "Test QR 2",
				Content:   "https://example.com/batch_123/2",
				Type:      "static",
				URL:       "https://example.com/batch_123/2",
				CreatedBy: "user_123",
				CreatedAt: time.Now(),
				UpdatedAt: time.Now(),
			},
		}

		mockService.On("GenerateBatchQRCodes", mock.Anything, batchID).Return(expectedQRCodes, nil).Once()

		// Create request
		req, _ := http.NewRequest(http.MethodPost, "/batches/"+batchID+"/generate", nil)

		// Create response recorder
		w := httptest.NewRecorder()

		// Perform request
		router.ServeHTTP(w, req)

		// Assert response
		assert.Equal(t, http.StatusOK, w.Code)

		// Parse response
		var response []*domain.QRCode
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.Equal(t, 2, len(response))
		assert.Equal(t, "Test QR 1", response[0].Name)

		// Assert mock expectations
		mockService.AssertExpectations(t)
	})

	// Test missing batch ID
	t.Run("MissingBatchID", func(t *testing.T) {
		// Create request without batch ID
		req, _ := http.NewRequest(http.MethodPost, "/batches//generate", nil)

		// Create response recorder
		w := httptest.NewRecorder()

		// Perform request
		router.ServeHTTP(w, req)

		// Assert response
		assert.Equal(t, http.StatusBadRequest, w.Code) // Changed to StatusBadRequest since the handler checks for empty ID
	})

	// Test service error - batch not found
	t.Run("BatchNotFound", func(t *testing.T) {
		// Prepare test data
		batchID := "batch_456"

		// Mock service response
		mockService.On("GenerateBatchQRCodes", mock.Anything, batchID).Return([]*domain.QRCode(nil), testutils.NewError("batch QR code not found")).Once()

		// Create request
		req, _ := http.NewRequest(http.MethodPost, "/batches/"+batchID+"/generate", nil)

		// Create response recorder
		w := httptest.NewRecorder()

		// Perform request
		router.ServeHTTP(w, req)

		// Assert response
		assert.Equal(t, http.StatusNotFound, w.Code)
		assert.Contains(t, w.Body.String(), "batch QR code not found")

		// Assert mock expectations
		mockService.AssertExpectations(t)
	})

	// Test service error - internal error
	t.Run("InternalServerError", func(t *testing.T) {
		// Prepare test data
		batchID := "batch_123"

		// Mock service response
		mockService.On("GenerateBatchQRCodes", mock.Anything, batchID).Return([]*domain.QRCode(nil), testutils.NewError("internal error")).Once()

		// Create request
		req, _ := http.NewRequest(http.MethodPost, "/batches/"+batchID+"/generate", nil)

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