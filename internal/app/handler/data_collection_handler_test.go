package handler

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"cdk-office/internal/app/service"
	"cdk-office/internal/shared/testutils"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockDataCollectionService is a mock implementation of DataCollectionServiceInterface
type MockDataCollectionService struct {
	mock.Mock
}

func (m *MockDataCollectionService) CreateDataCollection(ctx context.Context, req *service.CreateDataCollectionRequest) (*service.DataCollection, error) {
	args := m.Called(ctx, req)
	return args.Get(0).(*service.DataCollection), args.Error(1)
}

func (m *MockDataCollectionService) UpdateDataCollection(ctx context.Context, collectionID string, req *service.UpdateDataCollectionRequest) error {
	args := m.Called(ctx, collectionID, req)
	return args.Error(0)
}

func (m *MockDataCollectionService) DeleteDataCollection(ctx context.Context, collectionID string) error {
	args := m.Called(ctx, collectionID)
	return args.Error(0)
}

func (m *MockDataCollectionService) ListDataCollections(ctx context.Context, appID string, page, size int) ([]*service.DataCollection, int64, error) {
	args := m.Called(ctx, appID, page, size)
	return args.Get(0).([]*service.DataCollection), args.Get(1).(int64), args.Error(2)
}

func (m *MockDataCollectionService) GetDataCollection(ctx context.Context, collectionID string) (*service.DataCollection, error) {
	args := m.Called(ctx, collectionID)
	return args.Get(0).(*service.DataCollection), args.Error(1)
}

func (m *MockDataCollectionService) SubmitDataEntry(ctx context.Context, req *service.SubmitDataEntryRequest) (*service.DataCollectionEntry, error) {
	args := m.Called(ctx, req)
	return args.Get(0).(*service.DataCollectionEntry), args.Error(1)
}

func (m *MockDataCollectionService) ListDataEntries(ctx context.Context, collectionID string, page, size int) ([]*service.DataCollectionEntry, int64, error) {
	args := m.Called(ctx, collectionID, page, size)
	return args.Get(0).([]*service.DataCollectionEntry), args.Get(1).(int64), args.Error(2)
}

func (m *MockDataCollectionService) ExportDataEntries(ctx context.Context, collectionID string) ([]*service.DataCollectionEntry, error) {
	args := m.Called(ctx, collectionID)
	return args.Get(0).([]*service.DataCollectionEntry), args.Error(1)
}

// TestNewDataCollectionHandler tests the NewDataCollectionHandler function
func TestNewDataCollectionHandler(t *testing.T) {
	handler := NewDataCollectionHandler()
	assert.NotNil(t, handler)
	assert.NotNil(t, handler.dataService)
}

// TestCreateDataCollection tests the CreateDataCollection handler
func TestCreateDataCollection(t *testing.T) {
	// Set up test environment
	gin.SetMode(gin.TestMode)

	// Create mock service
	mockService := new(MockDataCollectionService)

	// Create handler with mock service
	handler := &DataCollectionHandler{
		dataService: mockService,
	}

	// Create test router
	router := gin.New()
	router.POST("/collections", handler.CreateDataCollection)

	// Test successful creation
	t.Run("SuccessfulCreation", func(t *testing.T) {
		// Prepare test data
		reqBody := CreateDataCollectionRequest{
			AppID:       "app_123",
			Name:        "Test Collection",
			Description: "Test collection description",
			Schema:      "{\"type\": \"object\"}",
			Config:      "{\"required\": [\"name\"]}",
			CreatedBy:   "user_123",
		}

		// Mock service response
		expectedCollection := &service.DataCollection{
			ID:          "col_123",
			AppID:       "app_123",
			Name:        "Test Collection",
			Description: "Test collection description",
			Schema:      "{\"type\": \"object\"}",
			Config:      "{\"required\": [\"name\"]}",
			IsActive:    true,
			CreatedBy:   "user_123",
			CreatedAt:   time.Now(),
			UpdatedAt:   time.Now(),
		}

		mockService.On("CreateDataCollection", mock.Anything, mock.MatchedBy(func(req *service.CreateDataCollectionRequest) bool {
			return req.AppID == reqBody.AppID && req.Name == reqBody.Name
		})).Return(expectedCollection, nil).Once()

		// Create request
		jsonValue, _ := json.Marshal(reqBody)
		req, _ := http.NewRequest(http.MethodPost, "/collections", bytes.NewBuffer(jsonValue))
		req.Header.Set("Content-Type", "application/json")

		// Create response recorder
		w := httptest.NewRecorder()

		// Perform request
		router.ServeHTTP(w, req)

		// Assert response
		assert.Equal(t, http.StatusOK, w.Code)

		// Parse response
		var response service.DataCollection
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.Equal(t, expectedCollection.ID, response.ID)
		assert.Equal(t, expectedCollection.Name, response.Name)

		// Assert mock expectations
		mockService.AssertExpectations(t)
	})

	// Test invalid request body
	t.Run("InvalidRequestBody", func(t *testing.T) {
		// Create request with invalid JSON
		req, _ := http.NewRequest(http.MethodPost, "/collections", bytes.NewBufferString("{invalid json}"))
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
		reqBody := CreateDataCollectionRequest{
			AppID:       "app_123",
			Name:        "Test Collection",
			Description: "Test collection description",
			Schema:      "{\"type\": \"object\"}",
			CreatedBy:   "user_123",
		}

		// Mock service response
		mockService.On("CreateDataCollection", mock.Anything, mock.MatchedBy(func(req *service.CreateDataCollectionRequest) bool {
			return req.Name == "Test Collection"
		})).Return((*service.DataCollection)(nil), testutils.NewError("internal error")).Once()

		// Create request
		jsonValue, _ := json.Marshal(reqBody)
		req, _ := http.NewRequest(http.MethodPost, "/collections", bytes.NewBuffer(jsonValue))
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

// TestUpdateDataCollection tests the UpdateDataCollection handler
func TestUpdateDataCollection(t *testing.T) {
	// Set up test environment
	gin.SetMode(gin.TestMode)

	// Create mock service
	mockService := new(MockDataCollectionService)

	// Create handler with mock service
	handler := &DataCollectionHandler{
		dataService: mockService,
	}

	// Create test router with route parameter
	router := gin.New()
	router.PUT("/collections/:id", handler.UpdateDataCollection)

	// Test successful update
	t.Run("SuccessfulUpdate", func(t *testing.T) {
		// Prepare test data
		collectionID := "col_123"
		isActive := true
		reqBody := UpdateDataCollectionRequest{
			Name:        "Updated Collection",
			Description: "Updated collection description",
			Schema:      "{\"type\": \"object\", \"properties\": {\"name\": {\"type\": \"string\"}}}",
			Config:      "{\"required\": [\"name\", \"email\"]}",
			IsActive:    &isActive,
		}

		// Mock service response
		mockService.On("UpdateDataCollection", mock.Anything, collectionID, mock.MatchedBy(func(req *service.UpdateDataCollectionRequest) bool {
			return req.Name == "Updated Collection" && *req.IsActive == true
		})).Return(nil).Once()

		// Create request
		jsonValue, _ := json.Marshal(reqBody)
		req, _ := http.NewRequest(http.MethodPut, "/collections/"+collectionID, bytes.NewBuffer(jsonValue))
		req.Header.Set("Content-Type", "application/json")

		// Create response recorder
		w := httptest.NewRecorder()

		// Perform request
		router.ServeHTTP(w, req)

		// Assert response
		assert.Equal(t, http.StatusOK, w.Code)
		assert.Contains(t, w.Body.String(), "data collection updated successfully")

		// Assert mock expectations
		mockService.AssertExpectations(t)
	})

	// Test missing collection ID
	t.Run("MissingCollectionID", func(t *testing.T) {
		// Create request without collection ID
		reqBody := UpdateDataCollectionRequest{
			Name: "Updated Collection",
		}

		jsonValue, _ := json.Marshal(reqBody)
		req, _ := http.NewRequest(http.MethodPut, "/collections/", bytes.NewBuffer(jsonValue))
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
		req, _ := http.NewRequest(http.MethodPut, "/collections/col_123", bytes.NewBufferString("{invalid json}"))
		req.Header.Set("Content-Type", "application/json")

		// Create response recorder
		w := httptest.NewRecorder()

		// Perform request
		router.ServeHTTP(w, req)

		// Assert response
		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	// Test service error - collection not found
	t.Run("CollectionNotFound", func(t *testing.T) {
		// Prepare test data
		collectionID := "col_456"
		reqBody := UpdateDataCollectionRequest{
			Name: "Updated Collection",
		}

		// Mock service response
		mockService.On("UpdateDataCollection", mock.Anything, collectionID, mock.MatchedBy(func(req *service.UpdateDataCollectionRequest) bool {
			return req.Name == "Updated Collection"
		})).Return(testutils.NewError("data collection not found")).Once()

		// Create request
		jsonValue, _ := json.Marshal(reqBody)
		req, _ := http.NewRequest(http.MethodPut, "/collections/"+collectionID, bytes.NewBuffer(jsonValue))
		req.Header.Set("Content-Type", "application/json")

		// Create response recorder
		w := httptest.NewRecorder()

		// Perform request
		router.ServeHTTP(w, req)

		// Assert response
		assert.Equal(t, http.StatusNotFound, w.Code)
		assert.Contains(t, w.Body.String(), "data collection not found")

		// Assert mock expectations
		mockService.AssertExpectations(t)
	})

	// Test service error - internal error
	t.Run("InternalServerError", func(t *testing.T) {
		// Prepare test data
		collectionID := "col_123"
		reqBody := UpdateDataCollectionRequest{
			Name: "Updated Collection",
		}

		// Mock service response
		mockService.On("UpdateDataCollection", mock.Anything, collectionID, mock.MatchedBy(func(req *service.UpdateDataCollectionRequest) bool {
			return req.Name == "Updated Collection"
		})).Return(testutils.NewError("internal error")).Once()

		// Create request
		jsonValue, _ := json.Marshal(reqBody)
		req, _ := http.NewRequest(http.MethodPut, "/collections/"+collectionID, bytes.NewBuffer(jsonValue))
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

// TestDeleteDataCollection tests the DeleteDataCollection handler
func TestDeleteDataCollection(t *testing.T) {
	// Set up test environment
	gin.SetMode(gin.TestMode)

	// Create mock service
	mockService := new(MockDataCollectionService)

	// Create handler with mock service
	handler := &DataCollectionHandler{
		dataService: mockService,
	}

	// Create test router with route parameter
	router := gin.New()
	router.DELETE("/collections/:id", handler.DeleteDataCollection)

	// Test successful deletion
	t.Run("SuccessfulDeletion", func(t *testing.T) {
		// Prepare test data
		collectionID := "col_123"

		// Mock service response
		mockService.On("DeleteDataCollection", mock.Anything, collectionID).Return(nil).Once()

		// Create request
		req, _ := http.NewRequest(http.MethodDelete, "/collections/"+collectionID, nil)

		// Create response recorder
		w := httptest.NewRecorder()

		// Perform request
		router.ServeHTTP(w, req)

		// Assert response
		assert.Equal(t, http.StatusOK, w.Code)
		assert.Contains(t, w.Body.String(), "data collection deleted successfully")

		// Assert mock expectations
		mockService.AssertExpectations(t)
	})

	// Test missing collection ID
	t.Run("MissingCollectionID", func(t *testing.T) {
		// Create request without collection ID
		req, _ := http.NewRequest(http.MethodDelete, "/collections/", nil)

		// Create response recorder
		w := httptest.NewRecorder()

		// Perform request
		router.ServeHTTP(w, req)

		// Assert response
		assert.Equal(t, http.StatusNotFound, w.Code) // Changed to StatusNotFound since the route doesn't match
	})

	// Test service error - collection not found
	t.Run("CollectionNotFound", func(t *testing.T) {
		// Prepare test data
		collectionID := "col_456"

		// Mock service response
		mockService.On("DeleteDataCollection", mock.Anything, collectionID).Return(testutils.NewError("data collection not found")).Once()

		// Create request
		req, _ := http.NewRequest(http.MethodDelete, "/collections/"+collectionID, nil)

		// Create response recorder
		w := httptest.NewRecorder()

		// Perform request
		router.ServeHTTP(w, req)

		// Assert response
		assert.Equal(t, http.StatusNotFound, w.Code)
		assert.Contains(t, w.Body.String(), "data collection not found")

		// Assert mock expectations
		mockService.AssertExpectations(t)
	})

	// Test service error - internal error
	t.Run("InternalServerError", func(t *testing.T) {
		// Prepare test data
		collectionID := "col_123"

		// Mock service response
		mockService.On("DeleteDataCollection", mock.Anything, collectionID).Return(testutils.NewError("internal error")).Once()

		// Create request
		req, _ := http.NewRequest(http.MethodDelete, "/collections/"+collectionID, nil)

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

// TestListDataCollections tests the ListDataCollections handler
func TestListDataCollections(t *testing.T) {
	// Set up test environment
	gin.SetMode(gin.TestMode)

	// Create mock service
	mockService := new(MockDataCollectionService)

	// Create handler with mock service
	handler := &DataCollectionHandler{
		dataService: mockService,
	}

	// Create test router
	router := gin.New()
	router.GET("/collections", handler.ListDataCollections)

	// Test successful listing
	t.Run("SuccessfulListing", func(t *testing.T) {
		// Prepare test data
		appID := "app_123"

		// Mock service response
		expectedCollections := []*service.DataCollection{
			{
				ID:          "col_123",
				AppID:       appID,
				Name:        "Test Collection 1",
				Description: "Test collection 1 description",
				Schema:      "{\"type\": \"object\"}",
				IsActive:    true,
				CreatedBy:   "user_123",
				CreatedAt:   time.Now(),
				UpdatedAt:   time.Now(),
			},
			{
				ID:          "col_456",
				AppID:       appID,
				Name:        "Test Collection 2",
				Description: "Test collection 2 description",
				Schema:      "{\"type\": \"array\"}",
				IsActive:    false,
				CreatedBy:   "user_123",
				CreatedAt:   time.Now(),
				UpdatedAt:   time.Now(),
			},
		}

		mockService.On("ListDataCollections", mock.Anything, appID, 1, 10).Return(expectedCollections, int64(2), nil).Once()

		// Create request
		req, _ := http.NewRequest(http.MethodGet, "/collections?app_id="+appID, nil)

		// Create response recorder
		w := httptest.NewRecorder()

		// Perform request
		router.ServeHTTP(w, req)

		// Assert response
		assert.Equal(t, http.StatusOK, w.Code)

		// Parse response
		var response ListDataCollectionsResponse
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.Equal(t, int64(2), response.Total)
		assert.Equal(t, 2, len(response.Items))
		assert.Equal(t, "Test Collection 1", response.Items[0].Name)

		// Assert mock expectations
		mockService.AssertExpectations(t)
	})

	// Test missing app ID
	t.Run("MissingAppID", func(t *testing.T) {
		// Create request without app ID
		req, _ := http.NewRequest(http.MethodGet, "/collections", nil)

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
		mockService.On("ListDataCollections", mock.Anything, appID, 1, 10).Return([]*service.DataCollection(nil), int64(0), testutils.NewError("internal error")).Once()

		// Create request
		req, _ := http.NewRequest(http.MethodGet, "/collections?app_id="+appID, nil)

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
		mockService.On("ListDataCollections", mock.Anything, appID, 2, 5).Return([]*service.DataCollection{}, int64(0), nil).Once()

		// Create request with custom pagination
		req, _ := http.NewRequest(http.MethodGet, "/collections?app_id="+appID+"&page=2&size=5", nil)

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

// TestGetDataCollection tests the GetDataCollection handler
func TestGetDataCollection(t *testing.T) {
	// Set up test environment
	gin.SetMode(gin.TestMode)

	// Create mock service
	mockService := new(MockDataCollectionService)

	// Create handler with mock service
	handler := &DataCollectionHandler{
		dataService: mockService,
	}

	// Create test router with route parameter
	router := gin.New()
	router.GET("/collections/:id", handler.GetDataCollection)

	// Test successful retrieval
	t.Run("SuccessfulRetrieval", func(t *testing.T) {
		// Prepare test data
		collectionID := "col_123"

		// Mock service response
		expectedCollection := &service.DataCollection{
			ID:          collectionID,
			AppID:       "app_123",
			Name:        "Test Collection",
			Description: "Test collection description",
			Schema:      "{\"type\": \"object\"}",
			IsActive:    true,
			CreatedBy:   "user_123",
			CreatedAt:   time.Now(),
			UpdatedAt:   time.Now(),
		}

		mockService.On("GetDataCollection", mock.Anything, collectionID).Return(expectedCollection, nil).Once()

		// Create request
		req, _ := http.NewRequest(http.MethodGet, "/collections/"+collectionID, nil)

		// Create response recorder
		w := httptest.NewRecorder()

		// Perform request
		router.ServeHTTP(w, req)

		// Assert response
		assert.Equal(t, http.StatusOK, w.Code)

		// Parse response
		var response service.DataCollection
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.Equal(t, expectedCollection.ID, response.ID)
		assert.Equal(t, expectedCollection.Name, response.Name)

		// Assert mock expectations
		mockService.AssertExpectations(t)
	})

	// Test missing collection ID
	t.Run("MissingCollectionID", func(t *testing.T) {
		// Create request without collection ID
		req, _ := http.NewRequest(http.MethodGet, "/collections/", nil)

		// Create response recorder
		w := httptest.NewRecorder()

		// Perform request
		router.ServeHTTP(w, req)

		// Assert response
		assert.Equal(t, http.StatusNotFound, w.Code) // Changed to StatusNotFound since the route doesn't match
	})

	// Test service error - collection not found
	t.Run("CollectionNotFound", func(t *testing.T) {
		// Prepare test data
		collectionID := "col_456"

		// Mock service response
		mockService.On("GetDataCollection", mock.Anything, collectionID).Return((*service.DataCollection)(nil), testutils.NewError("data collection not found")).Once()

		// Create request
		req, _ := http.NewRequest(http.MethodGet, "/collections/"+collectionID, nil)

		// Create response recorder
		w := httptest.NewRecorder()

		// Perform request
		router.ServeHTTP(w, req)

		// Assert response
		assert.Equal(t, http.StatusNotFound, w.Code)
		assert.Contains(t, w.Body.String(), "data collection not found")

		// Assert mock expectations
		mockService.AssertExpectations(t)
	})

	// Test service error - internal error
	t.Run("InternalServerError", func(t *testing.T) {
		// Prepare test data
		collectionID := "col_123"

		// Mock service response
		mockService.On("GetDataCollection", mock.Anything, collectionID).Return((*service.DataCollection)(nil), testutils.NewError("internal error")).Once()

		// Create request
		req, _ := http.NewRequest(http.MethodGet, "/collections/"+collectionID, nil)

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

// TestSubmitDataEntry tests the SubmitDataEntry handler
func TestSubmitDataEntry(t *testing.T) {
	// Set up test environment
	gin.SetMode(gin.TestMode)

	// Create mock service
	mockService := new(MockDataCollectionService)

	// Create handler with mock service
	handler := &DataCollectionHandler{
		dataService: mockService,
	}

	// Create test router
	router := gin.New()
	router.POST("/entries", handler.SubmitDataEntry)

	// Test successful submission
	t.Run("SuccessfulSubmission", func(t *testing.T) {
		// Prepare test data
		reqBody := SubmitDataEntryRequest{
			CollectionID: "col_123",
			Data:         "{\"name\": \"John Doe\", \"email\": \"john@example.com\"}",
			CreatedBy:    "user_123",
		}

		// Mock service response
		expectedEntry := &service.DataCollectionEntry{
			ID:           "entry_123",
			CollectionID: "col_123",
			Data:         "{\"name\": \"John Doe\", \"email\": \"john@example.com\"}",
			CreatedBy:    "user_123",
			CreatedAt:    time.Now(),
		}

		mockService.On("SubmitDataEntry", mock.Anything, mock.MatchedBy(func(req *service.SubmitDataEntryRequest) bool {
			return req.CollectionID == "col_123" && req.CreatedBy == "user_123"
		})).Return(expectedEntry, nil).Once()

		// Create request
		jsonValue, _ := json.Marshal(reqBody)
		req, _ := http.NewRequest(http.MethodPost, "/entries", bytes.NewBuffer(jsonValue))
		req.Header.Set("Content-Type", "application/json")

		// Create response recorder
		w := httptest.NewRecorder()

		// Perform request
		router.ServeHTTP(w, req)

		// Assert response
		assert.Equal(t, http.StatusOK, w.Code)

		// Parse response
		var response service.DataCollectionEntry
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.Equal(t, expectedEntry.ID, response.ID)
		assert.Equal(t, expectedEntry.CollectionID, response.CollectionID)

		// Assert mock expectations
		mockService.AssertExpectations(t)
	})

	// Test invalid request body
	t.Run("InvalidRequestBody", func(t *testing.T) {
		// Create request with invalid JSON
		req, _ := http.NewRequest(http.MethodPost, "/entries", bytes.NewBufferString("{invalid json}"))
		req.Header.Set("Content-Type", "application/json")

		// Create response recorder
		w := httptest.NewRecorder()

		// Perform request
		router.ServeHTTP(w, req)

		// Assert response
		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	// Test service error - collection not found or inactive
	t.Run("CollectionNotFoundOrInactive", func(t *testing.T) {
		// Prepare test data
		reqBody := SubmitDataEntryRequest{
			CollectionID: "col_456",
			Data:         "{\"name\": \"John Doe\"}",
			CreatedBy:    "user_123",
		}

		// Mock service response
		mockService.On("SubmitDataEntry", mock.Anything, mock.MatchedBy(func(req *service.SubmitDataEntryRequest) bool {
			return req.CollectionID == "col_456"
		})).Return((*service.DataCollectionEntry)(nil), testutils.NewError("data collection not found or inactive")).Once()

		// Create request
		jsonValue, _ := json.Marshal(reqBody)
		req, _ := http.NewRequest(http.MethodPost, "/entries", bytes.NewBuffer(jsonValue))
		req.Header.Set("Content-Type", "application/json")

		// Create response recorder
		w := httptest.NewRecorder()

		// Perform request
		router.ServeHTTP(w, req)

		// Assert response
		assert.Equal(t, http.StatusNotFound, w.Code)
		assert.Contains(t, w.Body.String(), "data collection not found or inactive")

		// Assert mock expectations
		mockService.AssertExpectations(t)
	})

	// Test service error - internal error
	t.Run("InternalServerError", func(t *testing.T) {
		// Prepare test data
		reqBody := SubmitDataEntryRequest{
			CollectionID: "col_123",
			Data:         "{\"name\": \"John Doe\"}",
			CreatedBy:    "user_123",
		}

		// Mock service response
		mockService.On("SubmitDataEntry", mock.Anything, mock.MatchedBy(func(req *service.SubmitDataEntryRequest) bool {
			return req.CollectionID == "col_123"
		})).Return((*service.DataCollectionEntry)(nil), testutils.NewError("internal error")).Once()

		// Create request
		jsonValue, _ := json.Marshal(reqBody)
		req, _ := http.NewRequest(http.MethodPost, "/entries", bytes.NewBuffer(jsonValue))
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

// TestListDataEntries tests the ListDataEntries handler
func TestListDataEntries(t *testing.T) {
	// Set up test environment
	gin.SetMode(gin.TestMode)

	// Create mock service
	mockService := new(MockDataCollectionService)

	// Create handler with mock service
	handler := &DataCollectionHandler{
		dataService: mockService,
	}

	// Create test router
	router := gin.New()
	router.GET("/entries", handler.ListDataEntries)

	// Test successful listing
	t.Run("SuccessfulListing", func(t *testing.T) {
		// Prepare test data
		collectionID := "col_123"

		// Mock service response
		expectedEntries := []*service.DataCollectionEntry{
			{
				ID:           "entry_123",
				CollectionID: collectionID,
				Data:         "{\"name\": \"John Doe\", \"email\": \"john@example.com\"}",
				CreatedBy:    "user_123",
				CreatedAt:    time.Now(),
			},
			{
				ID:           "entry_456",
				CollectionID: collectionID,
				Data:         "{\"name\": \"Jane Smith\", \"email\": \"jane@example.com\"}",
				CreatedBy:    "user_456",
				CreatedAt:    time.Now(),
			},
		}

		mockService.On("ListDataEntries", mock.Anything, collectionID, 1, 10).Return(expectedEntries, int64(2), nil).Once()

		// Create request
		req, _ := http.NewRequest(http.MethodGet, "/entries?collection_id="+collectionID, nil)

		// Create response recorder
		w := httptest.NewRecorder()

		// Perform request
		router.ServeHTTP(w, req)

		// Assert response
		assert.Equal(t, http.StatusOK, w.Code)

		// Parse response
		var response ListDataEntriesResponse
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.Equal(t, int64(2), response.Total)
		assert.Equal(t, 2, len(response.Items))
		assert.Equal(t, "entry_123", response.Items[0].ID)

		// Assert mock expectations
		mockService.AssertExpectations(t)
	})

	// Test missing collection ID
	t.Run("MissingCollectionID", func(t *testing.T) {
		// Create request without collection ID
		req, _ := http.NewRequest(http.MethodGet, "/entries", nil)

		// Create response recorder
		w := httptest.NewRecorder()

		// Perform request
		router.ServeHTTP(w, req)

		// Assert response
		assert.Equal(t, http.StatusBadRequest, w.Code)
		assert.Contains(t, w.Body.String(), "collection id is required")
	})

	// Test service error - collection not found
	t.Run("CollectionNotFound", func(t *testing.T) {
		// Prepare test data
		collectionID := "col_456"

		// Mock service response
		mockService.On("ListDataEntries", mock.Anything, collectionID, 1, 10).Return([]*service.DataCollectionEntry(nil), int64(0), testutils.NewError("data collection not found")).Once()

		// Create request
		req, _ := http.NewRequest(http.MethodGet, "/entries?collection_id="+collectionID, nil)

		// Create response recorder
		w := httptest.NewRecorder()

		// Perform request
		router.ServeHTTP(w, req)

		// Assert response
		assert.Equal(t, http.StatusNotFound, w.Code)
		assert.Contains(t, w.Body.String(), "data collection not found")

		// Assert mock expectations
		mockService.AssertExpectations(t)
	})

	// Test service error - internal error
	t.Run("InternalServerError", func(t *testing.T) {
		// Prepare test data
		collectionID := "col_123"

		// Mock service response
		mockService.On("ListDataEntries", mock.Anything, collectionID, 1, 10).Return([]*service.DataCollectionEntry(nil), int64(0), testutils.NewError("internal error")).Once()

		// Create request
		req, _ := http.NewRequest(http.MethodGet, "/entries?collection_id="+collectionID, nil)

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
		collectionID := "col_123"

		// Mock service response
		mockService.On("ListDataEntries", mock.Anything, collectionID, 2, 5).Return([]*service.DataCollectionEntry{}, int64(0), nil).Once()

		// Create request with custom pagination
		req, _ := http.NewRequest(http.MethodGet, "/entries?collection_id="+collectionID+"&page=2&size=5", nil)

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

// TestExportDataEntries tests the ExportDataEntries handler
func TestExportDataEntries(t *testing.T) {
	// Set up test environment
	gin.SetMode(gin.TestMode)

	// Create mock service
	mockService := new(MockDataCollectionService)

	// Create handler with mock service
	handler := &DataCollectionHandler{
		dataService: mockService,
	}

	// Create test router with route parameter
	router := gin.New()
	router.GET("/collections/:id/export", handler.ExportDataEntries)

	// Test successful export
	t.Run("SuccessfulExport", func(t *testing.T) {
		// Prepare test data
		collectionID := "col_123"

		// Mock service response
		expectedEntries := []*service.DataCollectionEntry{
			{
				ID:           "entry_123",
				CollectionID: collectionID,
				Data:         "{\"name\": \"John Doe\", \"email\": \"john@example.com\"}",
				CreatedBy:    "user_123",
				CreatedAt:    time.Now(),
			},
			{
				ID:           "entry_456",
				CollectionID: collectionID,
				Data:         "{\"name\": \"Jane Smith\", \"email\": \"jane@example.com\"}",
				CreatedBy:    "user_456",
				CreatedAt:    time.Now(),
			},
		}

		mockService.On("ExportDataEntries", mock.Anything, collectionID).Return(expectedEntries, nil).Once()

		// Create request
		req, _ := http.NewRequest(http.MethodGet, "/collections/"+collectionID+"/export", nil)

		// Create response recorder
		w := httptest.NewRecorder()

		// Perform request
		router.ServeHTTP(w, req)

		// Assert response
		assert.Equal(t, http.StatusOK, w.Code)

		// Parse response
		var response []*service.DataCollectionEntry
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.Equal(t, 2, len(response))
		assert.Equal(t, "entry_123", response[0].ID)

		// Assert mock expectations
		mockService.AssertExpectations(t)
	})

	// Test missing collection ID
	t.Run("MissingCollectionID", func(t *testing.T) {
		// Create request without collection ID
		req, _ := http.NewRequest(http.MethodGet, "/collections//export", nil)

		// Create response recorder
		w := httptest.NewRecorder()

		// Perform request
		router.ServeHTTP(w, req)

		// Assert response
		assert.Equal(t, http.StatusBadRequest, w.Code) // Changed to StatusBadRequest since the handler checks for empty ID
	})

	// Test service error - collection not found
	t.Run("CollectionNotFound", func(t *testing.T) {
		// Prepare test data
		collectionID := "col_456"

		// Mock service response
		mockService.On("ExportDataEntries", mock.Anything, collectionID).Return([]*service.DataCollectionEntry(nil), testutils.NewError("data collection not found")).Once()

		// Create request
		req, _ := http.NewRequest(http.MethodGet, "/collections/"+collectionID+"/export", nil)

		// Create response recorder
		w := httptest.NewRecorder()

		// Perform request
		router.ServeHTTP(w, req)

		// Assert response
		assert.Equal(t, http.StatusNotFound, w.Code)
		assert.Contains(t, w.Body.String(), "data collection not found")

		// Assert mock expectations
		mockService.AssertExpectations(t)
	})

	// Test service error - internal error
	t.Run("InternalServerError", func(t *testing.T) {
		// Prepare test data
		collectionID := "col_123"

		// Mock service response
		mockService.On("ExportDataEntries", mock.Anything, collectionID).Return([]*service.DataCollectionEntry(nil), testutils.NewError("internal error")).Once()

		// Create request
		req, _ := http.NewRequest(http.MethodGet, "/collections/"+collectionID+"/export", nil)

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