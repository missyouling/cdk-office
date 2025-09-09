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

// MockQRCodeService is a mock implementation of QRCodeServiceInterface
type MockQRCodeService struct {
	mock.Mock
}

func (m *MockQRCodeService) CreateQRCode(ctx context.Context, req *service.CreateQRCodeRequest) (*domain.QRCode, error) {
	args := m.Called(ctx, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.QRCode), args.Error(1)
}

func (m *MockQRCodeService) UpdateQRCode(ctx context.Context, qrCodeID string, req *service.UpdateQRCodeRequest) error {
	args := m.Called(ctx, qrCodeID, req)
	return args.Error(0)
}

func (m *MockQRCodeService) DeleteQRCode(ctx context.Context, qrCodeID string) error {
	args := m.Called(ctx, qrCodeID)
	return args.Error(0)
}

func (m *MockQRCodeService) ListQRCodes(ctx context.Context, appID string, page, size int) ([]*domain.QRCode, int64, error) {
	args := m.Called(ctx, appID, page, size)
	if args.Get(0) == nil {
		return nil, 0, args.Error(2)
	}
	return args.Get(0).([]*domain.QRCode), args.Get(1).(int64), args.Error(2)
}

func (m *MockQRCodeService) GetQRCode(ctx context.Context, qrCodeID string) (*domain.QRCode, error) {
	args := m.Called(ctx, qrCodeID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.QRCode), args.Error(1)
}

func (m *MockQRCodeService) GenerateQRCodeImage(ctx context.Context, qrCodeID string) (string, error) {
	args := m.Called(ctx, qrCodeID)
	return args.String(0), args.Error(1)
}

// TestNewQRCodeHandler tests the NewQRCodeHandler function
func TestNewQRCodeHandler(t *testing.T) {
	handler := NewQRCodeHandler()
	assert.NotNil(t, handler)
	assert.NotNil(t, handler.qrCodeService)
}

// TestCreateQRCode tests the CreateQRCode handler
func TestCreateQRCode(t *testing.T) {
	// Set up test environment
	gin.SetMode(gin.TestMode)

	// Create mock service
	mockService := new(MockQRCodeService)

	// Create handler with mock service
	handler := &QRCodeHandler{
		qrCodeService: mockService,
	}

	// Create test router
	router := gin.New()
	router.POST("/qrcodes", handler.CreateQRCode)

	// Test successful creation
	t.Run("SuccessfulCreation", func(t *testing.T) {
		// Prepare test data
		reqBody := CreateQRCodeRequest{
			AppID:     "app_123",
			Name:      "Test QR Code",
			Content:   "Test content",
			Type:      "static",
			URL:       "https://example.com",
			CreatedBy: "user_123",
		}

		// Mock service response
		expectedQRCode := &domain.QRCode{
			ID:        "qr_123",
			AppID:     "app_123",
			Name:      "Test QR Code",
			Content:   "Test content",
			Type:      "static",
			URL:       "https://example.com",
			CreatedBy: "user_123",
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		}

		mockService.On("CreateQRCode", mock.Anything, mock.MatchedBy(func(req *service.CreateQRCodeRequest) bool {
			return req.AppID == reqBody.AppID && req.Name == reqBody.Name
		})).Return(expectedQRCode, nil).Once()

		// Create request
		jsonValue, _ := json.Marshal(reqBody)
		req, _ := http.NewRequest(http.MethodPost, "/qrcodes", bytes.NewBuffer(jsonValue))
		req.Header.Set("Content-Type", "application/json")

		// Create response recorder
		w := httptest.NewRecorder()

		// Perform request
		router.ServeHTTP(w, req)

		// Assert response
		assert.Equal(t, http.StatusOK, w.Code)

		// Parse response
		var response domain.QRCode
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.Equal(t, expectedQRCode.ID, response.ID)
		assert.Equal(t, expectedQRCode.Name, response.Name)

		// Assert mock expectations
		mockService.AssertExpectations(t)
	})

	// Test invalid request body
	t.Run("InvalidRequestBody", func(t *testing.T) {
		// Create request with invalid JSON
		req, _ := http.NewRequest(http.MethodPost, "/qrcodes", bytes.NewBufferString("{invalid json}"))
		req.Header.Set("Content-Type", "application/json")

		// Create response recorder
		w := httptest.NewRecorder()

		// Perform request
		router.ServeHTTP(w, req)

		// Assert response
		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	// Test invalid QR code type
	t.Run("InvalidQRCodeType", func(t *testing.T) {
		// Prepare test data
		reqBody := CreateQRCodeRequest{
			AppID:     "app_123",
			Name:      "Test QR Code",
			Content:   "Test content",
			Type:      "invalid",
			CreatedBy: "user_123",
		}

		// Mock service response
		mockService.On("CreateQRCode", mock.Anything, mock.MatchedBy(func(req *service.CreateQRCodeRequest) bool {
			return req.Type == "invalid"
		})).Return((*domain.QRCode)(nil), testutils.NewError("invalid QR code type")).Once()

		// Create request
		jsonValue, _ := json.Marshal(reqBody)
		req, _ := http.NewRequest(http.MethodPost, "/qrcodes", bytes.NewBuffer(jsonValue))
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

	// Test service error
	t.Run("ServiceError", func(t *testing.T) {
		// Prepare test data
		reqBody := CreateQRCodeRequest{
			AppID:     "app_123",
			Name:      "Test QR Code",
			Content:   "Test content",
			Type:      "static",
			CreatedBy: "user_123",
		}

		// Mock service response
		mockService.On("CreateQRCode", mock.Anything, mock.MatchedBy(func(req *service.CreateQRCodeRequest) bool {
			return req.Name == "Test QR Code"
		})).Return((*domain.QRCode)(nil), testutils.NewError("internal error")).Once()

		// Create request
		jsonValue, _ := json.Marshal(reqBody)
		req, _ := http.NewRequest(http.MethodPost, "/qrcodes", bytes.NewBuffer(jsonValue))
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

// TestUpdateQRCode tests the UpdateQRCode handler
func TestUpdateQRCode(t *testing.T) {
	// Set up test environment
	gin.SetMode(gin.TestMode)

	// Create mock service
	mockService := new(MockQRCodeService)

	// Create handler with mock service
	handler := &QRCodeHandler{
		qrCodeService: mockService,
	}

	// Create test router with route parameter
	router := gin.New()
	router.PUT("/qrcodes/:id", handler.UpdateQRCode)

	// Test successful update
	t.Run("SuccessfulUpdate", func(t *testing.T) {
		// Prepare test data
		qrCodeID := "qr_123"
		reqBody := UpdateQRCodeRequest{
			Name:    "Updated QR Code",
			Content: "Updated content",
			URL:     "https://updated-example.com",
		}

		// Mock service response
		mockService.On("UpdateQRCode", mock.Anything, qrCodeID, mock.MatchedBy(func(req *service.UpdateQRCodeRequest) bool {
			return req.Name == "Updated QR Code" && req.Content == "Updated content"
		})).Return(nil).Once()

		// Create request
		jsonValue, _ := json.Marshal(reqBody)
		req, _ := http.NewRequest(http.MethodPut, "/qrcodes/"+qrCodeID, bytes.NewBuffer(jsonValue))
		req.Header.Set("Content-Type", "application/json")

		// Create response recorder
		w := httptest.NewRecorder()

		// Perform request
		router.ServeHTTP(w, req)

		// Assert response
		assert.Equal(t, http.StatusOK, w.Code)
		assert.Contains(t, w.Body.String(), "QR code updated successfully")

		// Assert mock expectations
		mockService.AssertExpectations(t)
	})

	// Test missing QR code ID
	t.Run("MissingQRCodeID", func(t *testing.T) {
		// Create request without QR code ID
		reqBody := UpdateQRCodeRequest{
			Name: "Updated QR Code",
		}

		jsonValue, _ := json.Marshal(reqBody)
		req, _ := http.NewRequest(http.MethodPut, "/qrcodes/", bytes.NewBuffer(jsonValue))
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
		req, _ := http.NewRequest(http.MethodPut, "/qrcodes/qr_123", bytes.NewBufferString("{invalid json}"))
		req.Header.Set("Content-Type", "application/json")

		// Create response recorder
		w := httptest.NewRecorder()

		// Perform request
		router.ServeHTTP(w, req)

		// Assert response
		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	// Test service error - QR code not found
	t.Run("QRCodeNotFound", func(t *testing.T) {
		// Prepare test data
		qrCodeID := "qr_456"
		reqBody := UpdateQRCodeRequest{
			Name: "Updated QR Code",
		}

		// Mock service response
		mockService.On("UpdateQRCode", mock.Anything, qrCodeID, mock.MatchedBy(func(req *service.UpdateQRCodeRequest) bool {
			return req.Name == "Updated QR Code"
		})).Return(testutils.NewError("QR code not found")).Once()

		// Create request
		jsonValue, _ := json.Marshal(reqBody)
		req, _ := http.NewRequest(http.MethodPut, "/qrcodes/"+qrCodeID, bytes.NewBuffer(jsonValue))
		req.Header.Set("Content-Type", "application/json")

		// Create response recorder
		w := httptest.NewRecorder()

		// Perform request
		router.ServeHTTP(w, req)

		// Assert response
		assert.Equal(t, http.StatusNotFound, w.Code)
		assert.Contains(t, w.Body.String(), "QR code not found")

		// Assert mock expectations
		mockService.AssertExpectations(t)
	})

	// Test service error - internal error
	t.Run("InternalServerError", func(t *testing.T) {
		// Prepare test data
		qrCodeID := "qr_123"
		reqBody := UpdateQRCodeRequest{
			Name: "Updated QR Code",
		}

		// Mock service response
		mockService.On("UpdateQRCode", mock.Anything, qrCodeID, mock.MatchedBy(func(req *service.UpdateQRCodeRequest) bool {
			return req.Name == "Updated QR Code"
		})).Return(testutils.NewError("internal error")).Once()

		// Create request
		jsonValue, _ := json.Marshal(reqBody)
		req, _ := http.NewRequest(http.MethodPut, "/qrcodes/"+qrCodeID, bytes.NewBuffer(jsonValue))
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

// TestDeleteQRCode tests the DeleteQRCode handler
func TestDeleteQRCode(t *testing.T) {
	// Set up test environment
	gin.SetMode(gin.TestMode)

	// Create mock service
	mockService := new(MockQRCodeService)

	// Create handler with mock service
	handler := &QRCodeHandler{
		qrCodeService: mockService,
	}

	// Create test router with route parameter
	router := gin.New()
	router.DELETE("/qrcodes/:id", handler.DeleteQRCode)

	// Test successful deletion
	t.Run("SuccessfulDeletion", func(t *testing.T) {
		// Prepare test data
		qrCodeID := "qr_123"

		// Mock service response
		mockService.On("DeleteQRCode", mock.Anything, qrCodeID).Return(nil).Once()

		// Create request
		req, _ := http.NewRequest(http.MethodDelete, "/qrcodes/"+qrCodeID, nil)

		// Create response recorder
		w := httptest.NewRecorder()

		// Perform request
		router.ServeHTTP(w, req)

		// Assert response
		assert.Equal(t, http.StatusOK, w.Code)
		assert.Contains(t, w.Body.String(), "QR code deleted successfully")

		// Assert mock expectations
		mockService.AssertExpectations(t)
	})

	// Test missing QR code ID
	t.Run("MissingQRCodeID", func(t *testing.T) {
		// Create request without QR code ID
		req, _ := http.NewRequest(http.MethodDelete, "/qrcodes/", nil)

		// Create response recorder
		w := httptest.NewRecorder()

		// Perform request
		router.ServeHTTP(w, req)

		// Assert response
		assert.Equal(t, http.StatusNotFound, w.Code) // Changed to StatusNotFound since the route doesn't match
	})

	// Test service error - QR code not found
	t.Run("QRCodeNotFound", func(t *testing.T) {
		// Prepare test data
		qrCodeID := "qr_456"

		// Mock service response
		mockService.On("DeleteQRCode", mock.Anything, qrCodeID).Return(testutils.NewError("QR code not found")).Once()

		// Create request
		req, _ := http.NewRequest(http.MethodDelete, "/qrcodes/"+qrCodeID, nil)

		// Create response recorder
		w := httptest.NewRecorder()

		// Perform request
		router.ServeHTTP(w, req)

		// Assert response
		assert.Equal(t, http.StatusNotFound, w.Code)
		assert.Contains(t, w.Body.String(), "QR code not found")

		// Assert mock expectations
		mockService.AssertExpectations(t)
	})

	// Test service error - internal error
	t.Run("InternalServerError", func(t *testing.T) {
		// Prepare test data
		qrCodeID := "qr_123"

		// Mock service response
		mockService.On("DeleteQRCode", mock.Anything, qrCodeID).Return(testutils.NewError("internal error")).Once()

		// Create request
		req, _ := http.NewRequest(http.MethodDelete, "/qrcodes/"+qrCodeID, nil)

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

// TestListQRCodes tests the ListQRCodes handler
func TestListQRCodes(t *testing.T) {
	// Set up test environment
	gin.SetMode(gin.TestMode)

	// Create mock service
	mockService := new(MockQRCodeService)

	// Create handler with mock service
	handler := &QRCodeHandler{
		qrCodeService: mockService,
	}

	// Create test router
	router := gin.New()
	router.GET("/qrcodes", handler.ListQRCodes)

	// Test successful listing
	t.Run("SuccessfulListing", func(t *testing.T) {
		// Prepare test data
		appID := "app_123"

		// Mock service response
		expectedQRCodes := []*domain.QRCode{
			{
				ID:        "qr_123",
				AppID:     appID,
				Name:      "Test QR Code 1",
				Content:   "Test content 1",
				Type:      "static",
				URL:       "https://example1.com",
				CreatedBy: "user_123",
				CreatedAt: time.Now(),
				UpdatedAt: time.Now(),
			},
			{
				ID:        "qr_456",
				AppID:     appID,
				Name:      "Test QR Code 2",
				Content:   "Test content 2",
				Type:      "dynamic",
				URL:       "https://example2.com",
				CreatedBy: "user_123",
				CreatedAt: time.Now(),
				UpdatedAt: time.Now(),
			},
		}

		mockService.On("ListQRCodes", mock.Anything, appID, 1, 10).Return(expectedQRCodes, int64(2), nil).Once()

		// Create request
		req, _ := http.NewRequest(http.MethodGet, "/qrcodes?app_id="+appID, nil)

		// Create response recorder
		w := httptest.NewRecorder()

		// Perform request
		router.ServeHTTP(w, req)

		// Assert response
		assert.Equal(t, http.StatusOK, w.Code)

		// Parse response
		var response ListQRCodesResponse
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.Equal(t, int64(2), response.Total)
		assert.Equal(t, 2, len(response.Items))
		assert.Equal(t, "Test QR Code 1", response.Items[0].Name)

		// Assert mock expectations
		mockService.AssertExpectations(t)
	})

	// Test missing app ID
	t.Run("MissingAppID", func(t *testing.T) {
		// Create request without app ID
		req, _ := http.NewRequest(http.MethodGet, "/qrcodes", nil)

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
		mockService.On("ListQRCodes", mock.Anything, appID, 1, 10).Return([]*domain.QRCode(nil), int64(0), testutils.NewError("internal error")).Once()

		// Create request
		req, _ := http.NewRequest(http.MethodGet, "/qrcodes?app_id="+appID, nil)

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
		mockService.On("ListQRCodes", mock.Anything, appID, 2, 5).Return([]*domain.QRCode{}, int64(0), nil).Once()

		// Create request with custom pagination
		req, _ := http.NewRequest(http.MethodGet, "/qrcodes?app_id="+appID+"&page=2&size=5", nil)

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

// TestGetQRCode tests the GetQRCode handler
func TestGetQRCode(t *testing.T) {
	// Set up test environment
	gin.SetMode(gin.TestMode)

	// Create mock service
	mockService := new(MockQRCodeService)

	// Create handler with mock service
	handler := &QRCodeHandler{
		qrCodeService: mockService,
	}

	// Create test router with route parameter
	router := gin.New()
	router.GET("/qrcodes/:id", handler.GetQRCode)

	// Test successful retrieval
	t.Run("SuccessfulRetrieval", func(t *testing.T) {
		// Prepare test data
		qrCodeID := "qr_123"

		// Mock service response
		expectedQRCode := &domain.QRCode{
			ID:        qrCodeID,
			AppID:     "app_123",
			Name:      "Test QR Code",
			Content:   "Test content",
			Type:      "static",
			URL:       "https://example.com",
			CreatedBy: "user_123",
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		}

		mockService.On("GetQRCode", mock.Anything, qrCodeID).Return(expectedQRCode, nil).Once()

		// Create request
		req, _ := http.NewRequest(http.MethodGet, "/qrcodes/"+qrCodeID, nil)

		// Create response recorder
		w := httptest.NewRecorder()

		// Perform request
		router.ServeHTTP(w, req)

		// Assert response
		assert.Equal(t, http.StatusOK, w.Code)

		// Parse response
		var response domain.QRCode
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.Equal(t, expectedQRCode.ID, response.ID)
		assert.Equal(t, expectedQRCode.Name, response.Name)

		// Assert mock expectations
		mockService.AssertExpectations(t)
	})

	// Test missing QR code ID
	t.Run("MissingQRCodeID", func(t *testing.T) {
		// Create request without QR code ID
		req, _ := http.NewRequest(http.MethodGet, "/qrcodes/", nil)

		// Create response recorder
		w := httptest.NewRecorder()

		// Perform request
		router.ServeHTTP(w, req)

		// Assert response
		assert.Equal(t, http.StatusNotFound, w.Code) // Changed to StatusNotFound since the route doesn't match
	})

	// Test service error - QR code not found
	t.Run("QRCodeNotFound", func(t *testing.T) {
		// Prepare test data
		qrCodeID := "qr_456"

		// Mock service response
		mockService.On("GetQRCode", mock.Anything, qrCodeID).Return((*domain.QRCode)(nil), testutils.NewError("QR code not found")).Once()

		// Create request
		req, _ := http.NewRequest(http.MethodGet, "/qrcodes/"+qrCodeID, nil)

		// Create response recorder
		w := httptest.NewRecorder()

		// Perform request
		router.ServeHTTP(w, req)

		// Assert response
		assert.Equal(t, http.StatusNotFound, w.Code)
		assert.Contains(t, w.Body.String(), "QR code not found")

		// Assert mock expectations
		mockService.AssertExpectations(t)
	})

	// Test service error - internal error
	t.Run("InternalServerError", func(t *testing.T) {
		// Prepare test data
		qrCodeID := "qr_123"

		// Mock service response
		mockService.On("GetQRCode", mock.Anything, qrCodeID).Return((*domain.QRCode)(nil), testutils.NewError("internal error")).Once()

		// Create request
		req, _ := http.NewRequest(http.MethodGet, "/qrcodes/"+qrCodeID, nil)

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

// TestGenerateQRCodeImage tests the GenerateQRCodeImage handler
func TestGenerateQRCodeImage(t *testing.T) {
	// Set up test environment
	gin.SetMode(gin.TestMode)

	// Create mock service
	mockService := new(MockQRCodeService)

	// Create handler with mock service
	handler := &QRCodeHandler{
		qrCodeService: mockService,
	}

	// Create test router with route parameter
	router := gin.New()
	router.GET("/qrcodes/:id/image", handler.GenerateQRCodeImage)

	// Test successful generation
	t.Run("SuccessfulGeneration", func(t *testing.T) {
		// Prepare test data
		qrCodeID := "qr_123"
		expectedImagePath := "/path/to/qrcode/image.png"

		// Mock service response
		mockService.On("GenerateQRCodeImage", mock.Anything, qrCodeID).Return(expectedImagePath, nil).Once()

		// Create request
		req, _ := http.NewRequest(http.MethodGet, "/qrcodes/"+qrCodeID+"/image", nil)

		// Create response recorder
		w := httptest.NewRecorder()

		// Perform request
		router.ServeHTTP(w, req)

		// Assert response
		assert.Equal(t, http.StatusOK, w.Code)
		assert.Contains(t, w.Body.String(), expectedImagePath)

		// Assert mock expectations
		mockService.AssertExpectations(t)
	})

	// Test missing QR code ID
	t.Run("MissingQRCodeID", func(t *testing.T) {
		// Create request without QR code ID
		req, _ := http.NewRequest(http.MethodGet, "/qrcodes//image", nil)

		// Create response recorder
		w := httptest.NewRecorder()

		// Perform request
		router.ServeHTTP(w, req)

		// Assert response
		assert.Equal(t, http.StatusBadRequest, w.Code)
		assert.Contains(t, w.Body.String(), "QR code id is required")
	})

	// Test service error - QR code not found
	t.Run("QRCodeNotFound", func(t *testing.T) {
		// Prepare test data
		qrCodeID := "qr_456"

		// Mock service response
		mockService.On("GenerateQRCodeImage", mock.Anything, qrCodeID).Return("", testutils.NewError("QR code not found")).Once()

		// Create request
		req, _ := http.NewRequest(http.MethodGet, "/qrcodes/"+qrCodeID+"/image", nil)

		// Create response recorder
		w := httptest.NewRecorder()

		// Perform request
		router.ServeHTTP(w, req)

		// Assert response
		assert.Equal(t, http.StatusNotFound, w.Code)
		assert.Contains(t, w.Body.String(), "QR code not found")

		// Assert mock expectations
		mockService.AssertExpectations(t)
	})

	// Test service error - internal error
	t.Run("InternalServerError", func(t *testing.T) {
		// Prepare test data
		qrCodeID := "qr_123"

		// Mock service response
		mockService.On("GenerateQRCodeImage", mock.Anything, qrCodeID).Return("", testutils.NewError("internal error")).Once()

		// Create request
		req, _ := http.NewRequest(http.MethodGet, "/qrcodes/"+qrCodeID+"/image", nil)

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