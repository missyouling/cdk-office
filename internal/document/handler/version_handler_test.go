package handler

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"cdk-office/internal/document/domain"
	"cdk-office/internal/shared/testutils"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockVersionService is a mock implementation of VersionServiceInterface
type MockVersionService struct {
	mock.Mock
}

func (m *MockVersionService) CreateVersion(ctx context.Context, documentID, filePath string, fileSize int64) (*domain.DocumentVersion, error) {
	args := m.Called(ctx, documentID, filePath, fileSize)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.DocumentVersion), args.Error(1)
}

func (m *MockVersionService) GetVersion(ctx context.Context, versionID string) (*domain.DocumentVersion, error) {
	args := m.Called(ctx, versionID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.DocumentVersion), args.Error(1)
}

func (m *MockVersionService) ListVersions(ctx context.Context, documentID string) ([]*domain.DocumentVersion, error) {
	args := m.Called(ctx, documentID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*domain.DocumentVersion), args.Error(1)
}

func (m *MockVersionService) GetLatestVersion(ctx context.Context, documentID string) (*domain.DocumentVersion, error) {
	args := m.Called(ctx, documentID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.DocumentVersion), args.Error(1)
}

func (m *MockVersionService) RestoreVersion(ctx context.Context, versionID string) error {
	args := m.Called(ctx, versionID)
	return args.Error(0)
}

// TestNewVersionHandler tests the NewVersionHandler function
func TestNewVersionHandler(t *testing.T) {
	handler := NewVersionHandler()
	assert.NotNil(t, handler)
	assert.NotNil(t, handler.versionService)
}

// TestCreateVersion tests the CreateVersion handler
func TestCreateVersion(t *testing.T) {
	// Set up test environment
	gin.SetMode(gin.TestMode)

	// Create mock service
	mockService := new(MockVersionService)

	// Create handler with mock service
	handler := &VersionHandler{
		versionService: mockService,
	}

	// Create test router
	router := gin.New()
	router.POST("/versions", handler.CreateVersion)

	// Test successful creation
	t.Run("SuccessfulCreation", func(t *testing.T) {
		// Prepare test data
		reqBody := CreateVersionRequest{
			DocumentID: "doc_123",
			FilePath:   "/path/to/file.pdf",
			FileSize:   1024,
		}

		// Mock data
		expectedVersion := &domain.DocumentVersion{
			ID:         "ver_123",
			DocumentID: "doc_123",
			Version:    1,
			FilePath:   "/path/to/file.pdf",
			FileSize:   1024,
		}

		// Mock service response
		mockService.On("CreateVersion", mock.Anything, "doc_123", "/path/to/file.pdf", int64(1024)).Return(expectedVersion, nil).Once()

		// Create request
		jsonValue, _ := json.Marshal(reqBody)
		req, _ := http.NewRequest(http.MethodPost, "/versions", bytes.NewBuffer(jsonValue))
		req.Header.Set("Content-Type", "application/json")

		// Create response recorder
		w := httptest.NewRecorder()

		// Perform request
		router.ServeHTTP(w, req)

		// Assert response
		assert.Equal(t, http.StatusOK, w.Code)

		// Parse response
		var response domain.DocumentVersion
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.Equal(t, expectedVersion.ID, response.ID)
		assert.Equal(t, expectedVersion.DocumentID, response.DocumentID)

		// Assert mock expectations
		mockService.AssertExpectations(t)
	})

	// Test invalid request body
	t.Run("InvalidRequestBody", func(t *testing.T) {
		// Create request with invalid JSON
		req, _ := http.NewRequest(http.MethodPost, "/versions", bytes.NewBufferString("{invalid json}"))
		req.Header.Set("Content-Type", "application/json")

		// Create response recorder
		w := httptest.NewRecorder()

		// Perform request
		router.ServeHTTP(w, req)

		// Assert response
		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	// Test service error - document not found
	t.Run("DocumentNotFound", func(t *testing.T) {
		// Prepare test data
		reqBody := CreateVersionRequest{
			DocumentID: "doc_456",
			FilePath:   "/path/to/file.pdf",
			FileSize:   1024,
		}

		// Mock service response
		mockService.On("CreateVersion", mock.Anything, "doc_456", "/path/to/file.pdf", int64(1024)).Return((*domain.DocumentVersion)(nil), testutils.NewError("document not found")).Once()

		// Create request
		jsonValue, _ := json.Marshal(reqBody)
		req, _ := http.NewRequest(http.MethodPost, "/versions", bytes.NewBuffer(jsonValue))
		req.Header.Set("Content-Type", "application/json")

		// Create response recorder
		w := httptest.NewRecorder()

		// Perform request
		router.ServeHTTP(w, req)

		// Assert response
		assert.Equal(t, http.StatusNotFound, w.Code)
		assert.Contains(t, w.Body.String(), "document not found")

		// Assert mock expectations
		mockService.AssertExpectations(t)
	})

	// Test service error - internal error
	t.Run("InternalServerError", func(t *testing.T) {
		// Prepare test data
		reqBody := CreateVersionRequest{
			DocumentID: "doc_123",
			FilePath:   "/path/to/file.pdf",
			FileSize:   1024,
		}

		// Mock service response
		mockService.On("CreateVersion", mock.Anything, "doc_123", "/path/to/file.pdf", int64(1024)).Return((*domain.DocumentVersion)(nil), testutils.NewError("internal error")).Once()

		// Create request
		jsonValue, _ := json.Marshal(reqBody)
		req, _ := http.NewRequest(http.MethodPost, "/versions", bytes.NewBuffer(jsonValue))
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

// TestGetVersion tests the GetVersion handler
func TestGetVersion(t *testing.T) {
	// Set up test environment
	gin.SetMode(gin.TestMode)

	// Create mock service
	mockService := new(MockVersionService)

	// Create handler with mock service
	handler := &VersionHandler{
		versionService: mockService,
	}

	// Create test router with route parameter
	router := gin.New()
	router.GET("/versions/:id", handler.GetVersion)

	// Test successful retrieval
	t.Run("SuccessfulRetrieval", func(t *testing.T) {
		// Prepare test data
		versionID := "ver_123"
		expectedVersion := &domain.DocumentVersion{
			ID:         versionID,
			DocumentID: "doc_123",
			Version:    1,
			FilePath:   "/path/to/file.pdf",
			FileSize:   1024,
		}

		// Mock service response
		mockService.On("GetVersion", mock.Anything, versionID).Return(expectedVersion, nil).Once()

		// Create request
		req, _ := http.NewRequest(http.MethodGet, "/versions/"+versionID, nil)

		// Create response recorder
		w := httptest.NewRecorder()

		// Perform request
		router.ServeHTTP(w, req)

		// Assert response
		assert.Equal(t, http.StatusOK, w.Code)

		// Parse response
		var response domain.DocumentVersion
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.Equal(t, expectedVersion.ID, response.ID)
		assert.Equal(t, expectedVersion.DocumentID, response.DocumentID)

		// Assert mock expectations
		mockService.AssertExpectations(t)
	})

	// Test missing version ID
	t.Run("MissingVersionID", func(t *testing.T) {
		// Create request without version ID
		req, _ := http.NewRequest(http.MethodGet, "/versions/", nil)

		// Create response recorder
		w := httptest.NewRecorder()

		// Perform request
		router.ServeHTTP(w, req)

		// Assert response
		assert.Equal(t, http.StatusNotFound, w.Code) // Changed to StatusNotFound since the route doesn't match
	})

	// Test service error - version not found
	t.Run("VersionNotFound", func(t *testing.T) {
		// Prepare test data
		versionID := "ver_456"

		// Mock service response
		mockService.On("GetVersion", mock.Anything, versionID).Return((*domain.DocumentVersion)(nil), testutils.NewError("version not found")).Once()

		// Create request
		req, _ := http.NewRequest(http.MethodGet, "/versions/"+versionID, nil)

		// Create response recorder
		w := httptest.NewRecorder()

		// Perform request
		router.ServeHTTP(w, req)

		// Assert response
		assert.Equal(t, http.StatusNotFound, w.Code)
		assert.Contains(t, w.Body.String(), "version not found")

		// Assert mock expectations
		mockService.AssertExpectations(t)
	})

	// Test service error - internal error
	t.Run("InternalServerError", func(t *testing.T) {
		// Prepare test data
		versionID := "ver_123"

		// Mock service response
		mockService.On("GetVersion", mock.Anything, versionID).Return((*domain.DocumentVersion)(nil), testutils.NewError("internal error")).Once()

		// Create request
		req, _ := http.NewRequest(http.MethodGet, "/versions/"+versionID, nil)

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

// TestListVersions tests the ListVersions handler
func TestListVersions(t *testing.T) {
	// Set up test environment
	gin.SetMode(gin.TestMode)

	// Create mock service
	mockService := new(MockVersionService)

	// Create handler with mock service
	handler := &VersionHandler{
		versionService: mockService,
	}

	// Create test router with route parameter
	router := gin.New()
	router.GET("/versions/document/:document_id", handler.ListVersions)

	// Test successful listing
	t.Run("SuccessfulListing", func(t *testing.T) {
		// Prepare test data
		documentID := "doc_123"
		expectedVersions := []*domain.DocumentVersion{
			{
				ID:         "ver_123",
				DocumentID: documentID,
				Version:    1,
				FilePath:   "/path/to/file_v1.pdf",
				FileSize:   1024,
			},
			{
				ID:         "ver_456",
				DocumentID: documentID,
				Version:    2,
				FilePath:   "/path/to/file_v2.pdf",
				FileSize:   2048,
			},
		}

		// Mock service response
		mockService.On("ListVersions", mock.Anything, documentID).Return(expectedVersions, nil).Once()

		// Create request
		req, _ := http.NewRequest(http.MethodGet, "/versions/document/"+documentID, nil)

		// Create response recorder
		w := httptest.NewRecorder()

		// Perform request
		router.ServeHTTP(w, req)

		// Assert response
		assert.Equal(t, http.StatusOK, w.Code)

		// Parse response
		var response []*domain.DocumentVersion
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.Equal(t, expectedVersions, response)

		// Assert mock expectations
		mockService.AssertExpectations(t)
	})

	// Test missing document ID
	t.Run("MissingDocumentID", func(t *testing.T) {
		// Create request without document ID
		req, _ := http.NewRequest(http.MethodGet, "/versions/document/", nil)

		// Create response recorder
		w := httptest.NewRecorder()

		// Perform request
		router.ServeHTTP(w, req)

		// Assert response
		assert.Equal(t, http.StatusNotFound, w.Code) // Changed to StatusNotFound since the route doesn't match
	})

	// Test service error - document not found
	t.Run("DocumentNotFound", func(t *testing.T) {
		// Prepare test data
		documentID := "doc_456"

		// Mock service response
		mockService.On("ListVersions", mock.Anything, documentID).Return([]*domain.DocumentVersion(nil), testutils.NewError("document not found")).Once()

		// Create request
		req, _ := http.NewRequest(http.MethodGet, "/versions/document/"+documentID, nil)

		// Create response recorder
		w := httptest.NewRecorder()

		// Perform request
		router.ServeHTTP(w, req)

		// Assert response
		assert.Equal(t, http.StatusNotFound, w.Code)
		assert.Contains(t, w.Body.String(), "document not found")

		// Assert mock expectations
		mockService.AssertExpectations(t)
	})

	// Test service error - internal error
	t.Run("InternalServerError", func(t *testing.T) {
		// Prepare test data
		documentID := "doc_123"

		// Mock service response
		mockService.On("ListVersions", mock.Anything, documentID).Return([]*domain.DocumentVersion(nil), testutils.NewError("internal error")).Once()

		// Create request
		req, _ := http.NewRequest(http.MethodGet, "/versions/document/"+documentID, nil)

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

// TestGetLatestVersion tests the GetLatestVersion handler
func TestGetLatestVersion(t *testing.T) {
	// Set up test environment
	gin.SetMode(gin.TestMode)

	// Create mock service
	mockService := new(MockVersionService)

	// Create handler with mock service
	handler := &VersionHandler{
		versionService: mockService,
	}

	// Create test router with route parameter
	router := gin.New()
	router.GET("/versions/latest/:document_id", handler.GetLatestVersion)

	// Test successful retrieval
	t.Run("SuccessfulRetrieval", func(t *testing.T) {
		// Prepare test data
		documentID := "doc_123"
		expectedVersion := &domain.DocumentVersion{
			ID:         "ver_456",
			DocumentID: documentID,
			Version:    2,
			FilePath:   "/path/to/file_v2.pdf",
			FileSize:   2048,
		}

		// Mock service response
		mockService.On("GetLatestVersion", mock.Anything, documentID).Return(expectedVersion, nil).Once()

		// Create request
		req, _ := http.NewRequest(http.MethodGet, "/versions/latest/"+documentID, nil)

		// Create response recorder
		w := httptest.NewRecorder()

		// Perform request
		router.ServeHTTP(w, req)

		// Assert response
		assert.Equal(t, http.StatusOK, w.Code)

		// Parse response
		var response domain.DocumentVersion
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.Equal(t, expectedVersion.ID, response.ID)
		assert.Equal(t, expectedVersion.Version, response.Version)

		// Assert mock expectations
		mockService.AssertExpectations(t)
	})

	// Test missing document ID
	t.Run("MissingDocumentID", func(t *testing.T) {
		// Create request without document ID
		req, _ := http.NewRequest(http.MethodGet, "/versions/latest/", nil)

		// Create response recorder
		w := httptest.NewRecorder()

		// Perform request
		router.ServeHTTP(w, req)

		// Assert response
		assert.Equal(t, http.StatusNotFound, w.Code) // Changed to StatusNotFound since the route doesn't match
	})

	// Test service error - document not found
	t.Run("DocumentNotFound", func(t *testing.T) {
		// Prepare test data
		documentID := "doc_456"

		// Mock service response
		mockService.On("GetLatestVersion", mock.Anything, documentID).Return((*domain.DocumentVersion)(nil), testutils.NewError("document not found")).Once()

		// Create request
		req, _ := http.NewRequest(http.MethodGet, "/versions/latest/"+documentID, nil)

		// Create response recorder
		w := httptest.NewRecorder()

		// Perform request
		router.ServeHTTP(w, req)

		// Assert response
		assert.Equal(t, http.StatusNotFound, w.Code)
		assert.Contains(t, w.Body.String(), "document not found")

		// Assert mock expectations
		mockService.AssertExpectations(t)
	})

	// Test service error - no versions found for document
	t.Run("NoVersionsFound", func(t *testing.T) {
		// Prepare test data
		documentID := "doc_123"

		// Mock service response
		mockService.On("GetLatestVersion", mock.Anything, documentID).Return((*domain.DocumentVersion)(nil), testutils.NewError("no versions found for document")).Once()

		// Create request
		req, _ := http.NewRequest(http.MethodGet, "/versions/latest/"+documentID, nil)

		// Create response recorder
		w := httptest.NewRecorder()

		// Perform request
		router.ServeHTTP(w, req)

		// Assert response
		assert.Equal(t, http.StatusNotFound, w.Code)
		assert.Contains(t, w.Body.String(), "no versions found for document")

		// Assert mock expectations
		mockService.AssertExpectations(t)
	})

	// Test service error - internal error
	t.Run("InternalServerError", func(t *testing.T) {
		// Prepare test data
		documentID := "doc_123"

		// Mock service response
		mockService.On("GetLatestVersion", mock.Anything, documentID).Return((*domain.DocumentVersion)(nil), testutils.NewError("internal error")).Once()

		// Create request
		req, _ := http.NewRequest(http.MethodGet, "/versions/latest/"+documentID, nil)

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

// TestRestoreVersion tests the RestoreVersion handler
func TestRestoreVersion(t *testing.T) {
	// Set up test environment
	gin.SetMode(gin.TestMode)

	// Create mock service
	mockService := new(MockVersionService)

	// Create handler with mock service
	handler := &VersionHandler{
		versionService: mockService,
	}

	// Create test router
	router := gin.New()
	router.POST("/versions/restore", handler.RestoreVersion)

	// Test successful restoration
	t.Run("SuccessfulRestoration", func(t *testing.T) {
		// Prepare test data
		reqBody := RestoreVersionRequest{
			VersionID: "ver_123",
		}

		// Mock service response
		mockService.On("RestoreVersion", mock.Anything, "ver_123").Return(nil).Once()

		// Create request
		jsonValue, _ := json.Marshal(reqBody)
		req, _ := http.NewRequest(http.MethodPost, "/versions/restore", bytes.NewBuffer(jsonValue))
		req.Header.Set("Content-Type", "application/json")

		// Create response recorder
		w := httptest.NewRecorder()

		// Perform request
		router.ServeHTTP(w, req)

		// Assert response
		assert.Equal(t, http.StatusOK, w.Code)
		assert.Contains(t, w.Body.String(), "version restored successfully")

		// Assert mock expectations
		mockService.AssertExpectations(t)
	})

	// Test invalid request body
	t.Run("InvalidRequestBody", func(t *testing.T) {
		// Create request with invalid JSON
		req, _ := http.NewRequest(http.MethodPost, "/versions/restore", bytes.NewBufferString("{invalid json}"))
		req.Header.Set("Content-Type", "application/json")

		// Create response recorder
		w := httptest.NewRecorder()

		// Perform request
		router.ServeHTTP(w, req)

		// Assert response
		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	// Test service error - version not found
	t.Run("VersionNotFound", func(t *testing.T) {
		// Prepare test data
		reqBody := RestoreVersionRequest{
			VersionID: "ver_456",
		}

		// Mock service response
		mockService.On("RestoreVersion", mock.Anything, "ver_456").Return(testutils.NewError("version not found")).Once()

		// Create request
		jsonValue, _ := json.Marshal(reqBody)
		req, _ := http.NewRequest(http.MethodPost, "/versions/restore", bytes.NewBuffer(jsonValue))
		req.Header.Set("Content-Type", "application/json")

		// Create response recorder
		w := httptest.NewRecorder()

		// Perform request
		router.ServeHTTP(w, req)

		// Assert response
		assert.Equal(t, http.StatusNotFound, w.Code)
		assert.Contains(t, w.Body.String(), "version not found")

		// Assert mock expectations
		mockService.AssertExpectations(t)
	})

	// Test service error - document not found
	t.Run("DocumentNotFound", func(t *testing.T) {
		// Prepare test data
		reqBody := RestoreVersionRequest{
			VersionID: "ver_123",
		}

		// Mock service response
		mockService.On("RestoreVersion", mock.Anything, "ver_123").Return(testutils.NewError("document not found")).Once()

		// Create request
		jsonValue, _ := json.Marshal(reqBody)
		req, _ := http.NewRequest(http.MethodPost, "/versions/restore", bytes.NewBuffer(jsonValue))
		req.Header.Set("Content-Type", "application/json")

		// Create response recorder
		w := httptest.NewRecorder()

		// Perform request
		router.ServeHTTP(w, req)

		// Assert response
		assert.Equal(t, http.StatusNotFound, w.Code)
		assert.Contains(t, w.Body.String(), "document not found")

		// Assert mock expectations
		mockService.AssertExpectations(t)
	})

	// Test service error - internal error
	t.Run("InternalServerError", func(t *testing.T) {
		// Prepare test data
		reqBody := RestoreVersionRequest{
			VersionID: "ver_123",
		}

		// Mock service response
		mockService.On("RestoreVersion", mock.Anything, "ver_123").Return(testutils.NewError("internal error")).Once()

		// Create request
		jsonValue, _ := json.Marshal(reqBody)
		req, _ := http.NewRequest(http.MethodPost, "/versions/restore", bytes.NewBuffer(jsonValue))
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