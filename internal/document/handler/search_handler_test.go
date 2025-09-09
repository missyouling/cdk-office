package handler

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"

	"cdk-office/internal/document/domain"
	"cdk-office/internal/shared/testutils"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockSearchService is a mock implementation of SearchServiceInterface
type MockSearchService struct {
	mock.Mock
}

func (m *MockSearchService) SearchDocuments(ctx context.Context, query, teamID string, page, size int) ([]*domain.Document, int64, error) {
	args := m.Called(ctx, query, teamID, page, size)
	if args.Get(0) == nil {
		return nil, 0, args.Error(2)
	}
	return args.Get(0).([]*domain.Document), args.Get(1).(int64), args.Error(2)
}

// TestNewSearchHandler tests the NewSearchHandler function
func TestNewSearchHandler(t *testing.T) {
	handler := NewSearchHandler()
	assert.NotNil(t, handler)
	assert.NotNil(t, handler.searchService)
}

// TestSearchDocuments tests the SearchDocuments handler
func TestSearchDocuments(t *testing.T) {
	// Set up test environment
	gin.SetMode(gin.TestMode)

	// Create mock service
	mockService := new(MockSearchService)

	// Create handler with mock service
	handler := &SearchHandler{
		searchService: mockService,
	}

	// Create test router
	router := gin.New()
	router.GET("/search", handler.SearchDocuments)

	// Test successful search
	t.Run("SuccessfulSearch", func(t *testing.T) {
		// Prepare test data
		query := "test"
		teamID := "team_123"
		page := 1
		size := 10

		// Mock data
		expectedDocuments := []*domain.Document{
			{
				ID:     "doc_123",
				TeamID: teamID,
				Title:  "Test Document 1",
			},
			{
				ID:     "doc_456",
				TeamID: teamID,
				Title:  "Test Document 2",
			},
		}
		total := int64(2)

		// Mock service response
		mockService.On("SearchDocuments", mock.Anything, query, teamID, page, size).Return(expectedDocuments, total, nil).Once()

		// Create request
		req, _ := http.NewRequest(http.MethodGet, "/search?q="+query+"&team_id="+teamID+"&page="+strconv.Itoa(page)+"&size="+strconv.Itoa(size), nil)

		// Create response recorder
		w := httptest.NewRecorder()

		// Perform request
		router.ServeHTTP(w, req)

		// Assert response
		assert.Equal(t, http.StatusOK, w.Code)

		// Parse response
		var response SearchDocumentsResponse
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.Equal(t, expectedDocuments, response.Items)
		assert.Equal(t, total, response.Total)
		assert.Equal(t, page, response.Page)
		assert.Equal(t, size, response.Size)

		// Assert mock expectations
		mockService.AssertExpectations(t)
	})

	// Test search with default pagination
	t.Run("SearchWithDefaultPagination", func(t *testing.T) {
		// Prepare test data
		query := "test"

		// Mock data
		expectedDocuments := []*domain.Document{
			{
				ID:    "doc_123",
				Title: "Test Document 1",
			},
		}
		total := int64(1)

		// Mock service response with default page=1 and size=10
		mockService.On("SearchDocuments", mock.Anything, query, "", 1, 10).Return(expectedDocuments, total, nil).Once()

		// Create request
		req, _ := http.NewRequest(http.MethodGet, "/search?q="+query, nil)

		// Create response recorder
		w := httptest.NewRecorder()

		// Perform request
		router.ServeHTTP(w, req)

		// Assert response
		assert.Equal(t, http.StatusOK, w.Code)

		// Parse response
		var response SearchDocumentsResponse
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.Equal(t, expectedDocuments, response.Items)
		assert.Equal(t, total, response.Total)
		assert.Equal(t, 1, response.Page)
		assert.Equal(t, 10, response.Size)

		// Assert mock expectations
		mockService.AssertExpectations(t)
	})

	// Test search with custom pagination
	t.Run("SearchWithCustomPagination", func(t *testing.T) {
		// Prepare test data
		query := "test"
		page := 2
		size := 20

		// Mock data
		expectedDocuments := []*domain.Document{}
		total := int64(0)

		// Mock service response
		mockService.On("SearchDocuments", mock.Anything, query, "", page, size).Return(expectedDocuments, total, nil).Once()

		// Create request
		req, _ := http.NewRequest(http.MethodGet, "/search?q="+query+"&page="+strconv.Itoa(page)+"&size="+strconv.Itoa(size), nil)

		// Create response recorder
		w := httptest.NewRecorder()

		// Perform request
		router.ServeHTTP(w, req)

		// Assert response
		assert.Equal(t, http.StatusOK, w.Code)

		// Parse response
		var response SearchDocumentsResponse
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.Equal(t, expectedDocuments, response.Items)
		assert.Equal(t, total, response.Total)
		assert.Equal(t, page, response.Page)
		assert.Equal(t, size, response.Size)

		// Assert mock expectations
		mockService.AssertExpectations(t)
	})

	// Test search with invalid page parameter
	t.Run("SearchWithInvalidPage", func(t *testing.T) {
		// Prepare test data
		query := "test"

		// Mock data
		expectedDocuments := []*domain.Document{}
		total := int64(0)

		// Mock service response with default page=1
		mockService.On("SearchDocuments", mock.Anything, query, "", 1, 10).Return(expectedDocuments, total, nil).Once()

		// Create request with invalid page
		req, _ := http.NewRequest(http.MethodGet, "/search?q="+query+"&page=invalid", nil)

		// Create response recorder
		w := httptest.NewRecorder()

		// Perform request
		router.ServeHTTP(w, req)

		// Assert response
		assert.Equal(t, http.StatusOK, w.Code)

		// Parse response
		var response SearchDocumentsResponse
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.Equal(t, expectedDocuments, response.Items)
		assert.Equal(t, total, response.Total)
		assert.Equal(t, 1, response.Page) // Should default to 1
		assert.Equal(t, 10, response.Size)

		// Assert mock expectations
		mockService.AssertExpectations(t)
	})

	// Test search with invalid size parameter
	t.Run("SearchWithInvalidSize", func(t *testing.T) {
		// Prepare test data
		query := "test"

		// Mock data
		expectedDocuments := []*domain.Document{}
		total := int64(0)

		// Mock service response with default size=10
		mockService.On("SearchDocuments", mock.Anything, query, "", 1, 10).Return(expectedDocuments, total, nil).Once()

		// Create request with invalid size
		req, _ := http.NewRequest(http.MethodGet, "/search?q="+query+"&size=invalid", nil)

		// Create response recorder
		w := httptest.NewRecorder()

		// Perform request
		router.ServeHTTP(w, req)

		// Assert response
		assert.Equal(t, http.StatusOK, w.Code)

		// Parse response
		var response SearchDocumentsResponse
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.Equal(t, expectedDocuments, response.Items)
		assert.Equal(t, total, response.Total)
		assert.Equal(t, 1, response.Page)
		assert.Equal(t, 10, response.Size) // Should default to 10

		// Assert mock expectations
		mockService.AssertExpectations(t)
	})

	// Test search with size exceeding maximum
	t.Run("SearchWithSizeExceedingMaximum", func(t *testing.T) {
		// Prepare test data
		query := "test"

		// Mock data
		expectedDocuments := []*domain.Document{}
		total := int64(0)

		// Mock service response with default size=10
		mockService.On("SearchDocuments", mock.Anything, query, "", 1, 10).Return(expectedDocuments, total, nil).Once()

		// Create request with size exceeding maximum
		req, _ := http.NewRequest(http.MethodGet, "/search?q="+query+"&size=150", nil)

		// Create response recorder
		w := httptest.NewRecorder()

		// Perform request
		router.ServeHTTP(w, req)

		// Assert response
		assert.Equal(t, http.StatusOK, w.Code)

		// Parse response
		var response SearchDocumentsResponse
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.Equal(t, expectedDocuments, response.Items)
		assert.Equal(t, total, response.Total)
		assert.Equal(t, 1, response.Page)
		assert.Equal(t, 10, response.Size) // Should default to 10 when exceeding maximum of 100

		// Assert mock expectations
		mockService.AssertExpectations(t)
	})

	// Test service error
	t.Run("ServiceError", func(t *testing.T) {
		// Prepare test data
		query := "test"

		// Mock service response
		mockService.On("SearchDocuments", mock.Anything, query, "", 1, 10).Return([]*domain.Document(nil), int64(0), testutils.NewError("internal error")).Once()

		// Create request
		req, _ := http.NewRequest(http.MethodGet, "/search?q="+query, nil)

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