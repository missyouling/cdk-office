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

// MockFormDesignerService is a mock implementation of FormDesignerServiceInterface
type MockFormDesignerService struct {
	mock.Mock
}

func (m *MockFormDesignerService) CreateFormDesign(ctx context.Context, req *service.CreateFormDesignRequest) (*domain.FormDesign, error) {
	args := m.Called(ctx, req)
	return args.Get(0).(*domain.FormDesign), args.Error(1)
}

func (m *MockFormDesignerService) UpdateFormDesign(ctx context.Context, formID string, req *service.UpdateFormDesignRequest) error {
	args := m.Called(ctx, formID, req)
	return args.Error(0)
}

func (m *MockFormDesignerService) DeleteFormDesign(ctx context.Context, formID string) error {
	args := m.Called(ctx, formID)
	return args.Error(0)
}

func (m *MockFormDesignerService) ListFormDesigns(ctx context.Context, appID string, page, size int) ([]*domain.FormDesign, int64, error) {
	args := m.Called(ctx, appID, page, size)
	return args.Get(0).([]*domain.FormDesign), args.Get(1).(int64), args.Error(2)
}

func (m *MockFormDesignerService) GetFormDesign(ctx context.Context, formID string) (*domain.FormDesign, error) {
	args := m.Called(ctx, formID)
	return args.Get(0).(*domain.FormDesign), args.Error(1)
}

func (m *MockFormDesignerService) PublishFormDesign(ctx context.Context, formID string) error {
	args := m.Called(ctx, formID)
	return args.Error(0)
}

// TestNewFormDesignerHandler tests the NewFormDesignerHandler function
func TestNewFormDesignerHandler(t *testing.T) {
	handler := NewFormDesignerHandler()
	assert.NotNil(t, handler)
	assert.NotNil(t, handler.formService)
}

// TestCreateFormDesign tests the CreateFormDesign handler
func TestCreateFormDesign(t *testing.T) {
	// Set up test environment
	gin.SetMode(gin.TestMode)

	// Create mock service
	mockService := new(MockFormDesignerService)

	// Create handler with mock service
	handler := &FormDesignerHandler{
		formService: mockService,
	}

	// Create test router
	router := gin.New()
	router.POST("/forms", handler.CreateFormDesign)

	// Test successful creation
	t.Run("SuccessfulCreation", func(t *testing.T) {
		// Prepare test data
		reqBody := CreateFormDesignRequest{
			AppID:       "app_123",
			Name:        "Test Form",
			Description: "Test form description",
			Schema:      "{\"type\": \"object\"}",
			Config:      "{\"required\": [\"name\"]}",
			CreatedBy:   "user_123",
		}

		// Mock service response
		expectedForm := &domain.FormDesign{
			ID:          "form_123",
			AppID:       "app_123",
			Name:        "Test Form",
			Description: "Test form description",
			Schema:      "{\"type\": \"object\"}",
			Config:      "{\"required\": [\"name\"]}",
			IsActive:    true,
			IsPublished: false,
			CreatedBy:   "user_123",
			CreatedAt:   time.Now(),
			UpdatedAt:   time.Now(),
		}

		mockService.On("CreateFormDesign", mock.Anything, mock.MatchedBy(func(req *service.CreateFormDesignRequest) bool {
			return req.AppID == reqBody.AppID && req.Name == reqBody.Name
		})).Return(expectedForm, nil).Once()

		// Create request
		jsonValue, _ := json.Marshal(reqBody)
		req, _ := http.NewRequest(http.MethodPost, "/forms", bytes.NewBuffer(jsonValue))
		req.Header.Set("Content-Type", "application/json")

		// Create response recorder
		w := httptest.NewRecorder()

		// Perform request
		router.ServeHTTP(w, req)

		// Assert response
		assert.Equal(t, http.StatusOK, w.Code)

		// Parse response
		var response domain.FormDesign
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.Equal(t, expectedForm.ID, response.ID)
		assert.Equal(t, expectedForm.Name, response.Name)

		// Assert mock expectations
		mockService.AssertExpectations(t)
	})

	// Test invalid request body
	t.Run("InvalidRequestBody", func(t *testing.T) {
		// Create request with invalid JSON
		req, _ := http.NewRequest(http.MethodPost, "/forms", bytes.NewBufferString("{invalid json}"))
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
		reqBody := CreateFormDesignRequest{
			AppID:       "app_123",
			Name:        "Test Form",
			Description: "Test form description",
			Schema:      "{\"type\": \"object\"}",
			CreatedBy:   "user_123",
		}

		// Mock service response
		mockService.On("CreateFormDesign", mock.Anything, mock.MatchedBy(func(req *service.CreateFormDesignRequest) bool {
			return req.Name == "Test Form"
		})).Return((*domain.FormDesign)(nil), testutils.NewError("internal error")).Once()

		// Create request
		jsonValue, _ := json.Marshal(reqBody)
		req, _ := http.NewRequest(http.MethodPost, "/forms", bytes.NewBuffer(jsonValue))
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

// TestUpdateFormDesign tests the UpdateFormDesign handler
func TestUpdateFormDesign(t *testing.T) {
	// Set up test environment
	gin.SetMode(gin.TestMode)

	// Create mock service
	mockService := new(MockFormDesignerService)

	// Create handler with mock service
	handler := &FormDesignerHandler{
		formService: mockService,
	}

	// Create test router with route parameter
	router := gin.New()
	router.PUT("/forms/:id", handler.UpdateFormDesign)

	// Test successful update
	t.Run("SuccessfulUpdate", func(t *testing.T) {
		// Prepare test data
		formID := "form_123"
		isActive := true
		reqBody := UpdateFormDesignRequest{
			Name:        "Updated Form",
			Description: "Updated form description",
			Schema:      "{\"type\": \"object\", \"properties\": {\"name\": {\"type\": \"string\"}}}",
			Config:      "{\"required\": [\"name\", \"email\"]}",
			IsActive:    &isActive,
		}

		// Mock service response
		mockService.On("UpdateFormDesign", mock.Anything, formID, mock.MatchedBy(func(req *service.UpdateFormDesignRequest) bool {
			return req.Name == "Updated Form" && *req.IsActive == true
		})).Return(nil).Once()

		// Create request
		jsonValue, _ := json.Marshal(reqBody)
		req, _ := http.NewRequest(http.MethodPut, "/forms/"+formID, bytes.NewBuffer(jsonValue))
		req.Header.Set("Content-Type", "application/json")

		// Create response recorder
		w := httptest.NewRecorder()

		// Perform request
		router.ServeHTTP(w, req)

		// Assert response
		assert.Equal(t, http.StatusOK, w.Code)
		assert.Contains(t, w.Body.String(), "form design updated successfully")

		// Assert mock expectations
		mockService.AssertExpectations(t)
	})

	// Test missing form ID
	t.Run("MissingFormID", func(t *testing.T) {
		// Create request without form ID
		reqBody := UpdateFormDesignRequest{
			Name: "Updated Form",
		}

		jsonValue, _ := json.Marshal(reqBody)
		req, _ := http.NewRequest(http.MethodPut, "/forms/", bytes.NewBuffer(jsonValue))
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
		req, _ := http.NewRequest(http.MethodPut, "/forms/form_123", bytes.NewBufferString("{invalid json}"))
		req.Header.Set("Content-Type", "application/json")

		// Create response recorder
		w := httptest.NewRecorder()

		// Perform request
		router.ServeHTTP(w, req)

		// Assert response
		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	// Test service error - form not found
	t.Run("FormNotFound", func(t *testing.T) {
		// Prepare test data
		formID := "form_456"
		reqBody := UpdateFormDesignRequest{
			Name: "Updated Form",
		}

		// Mock service response
		mockService.On("UpdateFormDesign", mock.Anything, formID, mock.MatchedBy(func(req *service.UpdateFormDesignRequest) bool {
			return req.Name == "Updated Form"
		})).Return(testutils.NewError("form design not found")).Once()

		// Create request
		jsonValue, _ := json.Marshal(reqBody)
		req, _ := http.NewRequest(http.MethodPut, "/forms/"+formID, bytes.NewBuffer(jsonValue))
		req.Header.Set("Content-Type", "application/json")

		// Create response recorder
		w := httptest.NewRecorder()

		// Perform request
		router.ServeHTTP(w, req)

		// Assert response
		assert.Equal(t, http.StatusNotFound, w.Code)
		assert.Contains(t, w.Body.String(), "form design not found")

		// Assert mock expectations
		mockService.AssertExpectations(t)
	})

	// Test service error - cannot update published form
	t.Run("CannotUpdatePublishedForm", func(t *testing.T) {
		// Prepare test data
		formID := "form_123"
		reqBody := UpdateFormDesignRequest{
			Name: "Updated Form",
		}

		// Mock service response
		mockService.On("UpdateFormDesign", mock.Anything, formID, mock.MatchedBy(func(req *service.UpdateFormDesignRequest) bool {
			return req.Name == "Updated Form"
		})).Return(testutils.NewError("cannot update published form design")).Once()

		// Create request
		jsonValue, _ := json.Marshal(reqBody)
		req, _ := http.NewRequest(http.MethodPut, "/forms/"+formID, bytes.NewBuffer(jsonValue))
		req.Header.Set("Content-Type", "application/json")

		// Create response recorder
		w := httptest.NewRecorder()

		// Perform request
		router.ServeHTTP(w, req)

		// Assert response
		assert.Equal(t, http.StatusBadRequest, w.Code)
		assert.Contains(t, w.Body.String(), "cannot update published form design")

		// Assert mock expectations
		mockService.AssertExpectations(t)
	})

	// Test service error - internal error
	t.Run("InternalServerError", func(t *testing.T) {
		// Prepare test data
		formID := "form_123"
		reqBody := UpdateFormDesignRequest{
			Name: "Updated Form",
		}

		// Mock service response
		mockService.On("UpdateFormDesign", mock.Anything, formID, mock.MatchedBy(func(req *service.UpdateFormDesignRequest) bool {
			return req.Name == "Updated Form"
		})).Return(testutils.NewError("internal error")).Once()

		// Create request
		jsonValue, _ := json.Marshal(reqBody)
		req, _ := http.NewRequest(http.MethodPut, "/forms/"+formID, bytes.NewBuffer(jsonValue))
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

// TestDeleteFormDesign tests the DeleteFormDesign handler
func TestDeleteFormDesign(t *testing.T) {
	// Set up test environment
	gin.SetMode(gin.TestMode)

	// Create mock service
	mockService := new(MockFormDesignerService)

	// Create handler with mock service
	handler := &FormDesignerHandler{
		formService: mockService,
	}

	// Create test router with route parameter
	router := gin.New()
	router.DELETE("/forms/:id", handler.DeleteFormDesign)

	// Test successful deletion
	t.Run("SuccessfulDeletion", func(t *testing.T) {
		// Prepare test data
		formID := "form_123"

		// Mock service response
		mockService.On("DeleteFormDesign", mock.Anything, formID).Return(nil).Once()

		// Create request
		req, _ := http.NewRequest(http.MethodDelete, "/forms/"+formID, nil)

		// Create response recorder
		w := httptest.NewRecorder()

		// Perform request
		router.ServeHTTP(w, req)

		// Assert response
		assert.Equal(t, http.StatusOK, w.Code)
		assert.Contains(t, w.Body.String(), "form design deleted successfully")

		// Assert mock expectations
		mockService.AssertExpectations(t)
	})

	// Test missing form ID
	t.Run("MissingFormID", func(t *testing.T) {
		// Create request without form ID
		req, _ := http.NewRequest(http.MethodDelete, "/forms/", nil)

		// Create response recorder
		w := httptest.NewRecorder()

		// Perform request
		router.ServeHTTP(w, req)

		// Assert response
		assert.Equal(t, http.StatusNotFound, w.Code) // Changed to StatusNotFound since the route doesn't match
	})

	// Test service error - form not found
	t.Run("FormNotFound", func(t *testing.T) {
		// Prepare test data
		formID := "form_456"

		// Mock service response
		mockService.On("DeleteFormDesign", mock.Anything, formID).Return(testutils.NewError("form design not found")).Once()

		// Create request
		req, _ := http.NewRequest(http.MethodDelete, "/forms/"+formID, nil)

		// Create response recorder
		w := httptest.NewRecorder()

		// Perform request
		router.ServeHTTP(w, req)

		// Assert response
		assert.Equal(t, http.StatusNotFound, w.Code)
		assert.Contains(t, w.Body.String(), "form design not found")

		// Assert mock expectations
		mockService.AssertExpectations(t)
	})

	// Test service error - cannot delete published form
	t.Run("CannotDeletePublishedForm", func(t *testing.T) {
		// Prepare test data
		formID := "form_123"

		// Mock service response
		mockService.On("DeleteFormDesign", mock.Anything, formID).Return(testutils.NewError("cannot delete published form design")).Once()

		// Create request
		req, _ := http.NewRequest(http.MethodDelete, "/forms/"+formID, nil)

		// Create response recorder
		w := httptest.NewRecorder()

		// Perform request
		router.ServeHTTP(w, req)

		// Assert response
		assert.Equal(t, http.StatusBadRequest, w.Code)
		assert.Contains(t, w.Body.String(), "cannot delete published form design")

		// Assert mock expectations
		mockService.AssertExpectations(t)
	})

	// Test service error - internal error
	t.Run("InternalServerError", func(t *testing.T) {
		// Prepare test data
		formID := "form_123"

		// Mock service response
		mockService.On("DeleteFormDesign", mock.Anything, formID).Return(testutils.NewError("internal error")).Once()

		// Create request
		req, _ := http.NewRequest(http.MethodDelete, "/forms/"+formID, nil)

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

// TestListFormDesigns tests the ListFormDesigns handler
func TestListFormDesigns(t *testing.T) {
	// Set up test environment
	gin.SetMode(gin.TestMode)

	// Create mock service
	mockService := new(MockFormDesignerService)

	// Create handler with mock service
	handler := &FormDesignerHandler{
		formService: mockService,
	}

	// Create test router
	router := gin.New()
	router.GET("/forms", handler.ListFormDesigns)

	// Test successful listing
	t.Run("SuccessfulListing", func(t *testing.T) {
		// Prepare test data
		appID := "app_123"

		// Mock service response
		expectedForms := []*domain.FormDesign{
			{
				ID:          "form_123",
				AppID:       appID,
				Name:        "Test Form 1",
				Description: "Test form 1 description",
				Schema:      "{\"type\": \"object\"}",
				IsActive:    true,
				IsPublished: false,
				CreatedBy:   "user_123",
				CreatedAt:   time.Now(),
				UpdatedAt:   time.Now(),
			},
			{
				ID:          "form_456",
				AppID:       appID,
				Name:        "Test Form 2",
				Description: "Test form 2 description",
				Schema:      "{\"type\": \"array\"}",
				IsActive:    false,
				IsPublished: true,
				CreatedBy:   "user_123",
				CreatedAt:   time.Now(),
				UpdatedAt:   time.Now(),
			},
		}

		mockService.On("ListFormDesigns", mock.Anything, appID, 1, 10).Return(expectedForms, int64(2), nil).Once()

		// Create request
		req, _ := http.NewRequest(http.MethodGet, "/forms?app_id="+appID, nil)

		// Create response recorder
		w := httptest.NewRecorder()

		// Perform request
		router.ServeHTTP(w, req)

		// Assert response
		assert.Equal(t, http.StatusOK, w.Code)

		// Parse response
		var response ListFormDesignsResponse
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.Equal(t, int64(2), response.Total)
		assert.Equal(t, 2, len(response.Items))
		assert.Equal(t, "Test Form 1", response.Items[0].Name)

		// Assert mock expectations
		mockService.AssertExpectations(t)
	})

	// Test missing app ID
	t.Run("MissingAppID", func(t *testing.T) {
		// Create request without app ID
		req, _ := http.NewRequest(http.MethodGet, "/forms", nil)

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
		mockService.On("ListFormDesigns", mock.Anything, appID, 1, 10).Return([]*domain.FormDesign(nil), int64(0), testutils.NewError("internal error")).Once()

		// Create request
		req, _ := http.NewRequest(http.MethodGet, "/forms?app_id="+appID, nil)

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
		mockService.On("ListFormDesigns", mock.Anything, appID, 2, 5).Return([]*domain.FormDesign{}, int64(0), nil).Once()

		// Create request with custom pagination
		req, _ := http.NewRequest(http.MethodGet, "/forms?app_id="+appID+"&page=2&size=5", nil)

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

// TestGetFormDesign tests the GetFormDesign handler
func TestGetFormDesign(t *testing.T) {
	// Set up test environment
	gin.SetMode(gin.TestMode)

	// Create mock service
	mockService := new(MockFormDesignerService)

	// Create handler with mock service
	handler := &FormDesignerHandler{
		formService: mockService,
	}

	// Create test router with route parameter
	router := gin.New()
	router.GET("/forms/:id", handler.GetFormDesign)

	// Test successful retrieval
	t.Run("SuccessfulRetrieval", func(t *testing.T) {
		// Prepare test data
		formID := "form_123"

		// Mock service response
		expectedForm := &domain.FormDesign{
			ID:          formID,
			AppID:       "app_123",
			Name:        "Test Form",
			Description: "Test form description",
			Schema:      "{\"type\": \"object\"}",
			IsActive:    true,
			IsPublished: false,
			CreatedBy:   "user_123",
			CreatedAt:   time.Now(),
			UpdatedAt:   time.Now(),
		}

		mockService.On("GetFormDesign", mock.Anything, formID).Return(expectedForm, nil).Once()

		// Create request
		req, _ := http.NewRequest(http.MethodGet, "/forms/"+formID, nil)

		// Create response recorder
		w := httptest.NewRecorder()

		// Perform request
		router.ServeHTTP(w, req)

		// Assert response
		assert.Equal(t, http.StatusOK, w.Code)

		// Parse response
		var response domain.FormDesign
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.Equal(t, expectedForm.ID, response.ID)
		assert.Equal(t, expectedForm.Name, response.Name)

		// Assert mock expectations
		mockService.AssertExpectations(t)
	})

	// Test missing form ID
	t.Run("MissingFormID", func(t *testing.T) {
		// Create request without form ID
		req, _ := http.NewRequest(http.MethodGet, "/forms/", nil)

		// Create response recorder
		w := httptest.NewRecorder()

		// Perform request
		router.ServeHTTP(w, req)

		// Assert response
		assert.Equal(t, http.StatusNotFound, w.Code) // Changed to StatusNotFound since the route doesn't match
	})

	// Test service error - form not found
	t.Run("FormNotFound", func(t *testing.T) {
		// Prepare test data
		formID := "form_456"

		// Mock service response
		mockService.On("GetFormDesign", mock.Anything, formID).Return((*domain.FormDesign)(nil), testutils.NewError("form design not found")).Once()

		// Create request
		req, _ := http.NewRequest(http.MethodGet, "/forms/"+formID, nil)

		// Create response recorder
		w := httptest.NewRecorder()

		// Perform request
		router.ServeHTTP(w, req)

		// Assert response
		assert.Equal(t, http.StatusNotFound, w.Code)
		assert.Contains(t, w.Body.String(), "form design not found")

		// Assert mock expectations
		mockService.AssertExpectations(t)
	})

	// Test service error - internal error
	t.Run("InternalServerError", func(t *testing.T) {
		// Prepare test data
		formID := "form_123"

		// Mock service response
		mockService.On("GetFormDesign", mock.Anything, formID).Return((*domain.FormDesign)(nil), testutils.NewError("internal error")).Once()

		// Create request
		req, _ := http.NewRequest(http.MethodGet, "/forms/"+formID, nil)

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

// TestPublishFormDesign tests the PublishFormDesign handler
func TestPublishFormDesign(t *testing.T) {
	// Set up test environment
	gin.SetMode(gin.TestMode)

	// Create mock service
	mockService := new(MockFormDesignerService)

	// Create handler with mock service
	handler := &FormDesignerHandler{
		formService: mockService,
	}

	// Create test router with route parameter
	router := gin.New()
	router.POST("/forms/:id/publish", handler.PublishFormDesign)

	// Test successful publish
	t.Run("SuccessfulPublish", func(t *testing.T) {
		// Prepare test data
		formID := "form_123"

		// Mock service response
		mockService.On("PublishFormDesign", mock.Anything, formID).Return(nil).Once()

		// Create request
		req, _ := http.NewRequest(http.MethodPost, "/forms/"+formID+"/publish", nil)

		// Create response recorder
		w := httptest.NewRecorder()

		// Perform request
		router.ServeHTTP(w, req)

		// Assert response
		assert.Equal(t, http.StatusOK, w.Code)
		assert.Contains(t, w.Body.String(), "form design published successfully")

		// Assert mock expectations
		mockService.AssertExpectations(t)
	})

	// Test missing form ID
	t.Run("MissingFormID", func(t *testing.T) {
		// Create request without form ID
		req, _ := http.NewRequest(http.MethodPost, "/forms//publish", nil)

		// Create response recorder
		w := httptest.NewRecorder()

		// Perform request
		router.ServeHTTP(w, req)

		// Assert response
		assert.Equal(t, http.StatusBadRequest, w.Code)
		assert.Contains(t, w.Body.String(), "form id is required")
	})

	// Test service error - form not found
	t.Run("FormNotFound", func(t *testing.T) {
		// Prepare test data
		formID := "form_456"

		// Mock service response
		mockService.On("PublishFormDesign", mock.Anything, formID).Return(testutils.NewError("form design not found")).Once()

		// Create request
		req, _ := http.NewRequest(http.MethodPost, "/forms/"+formID+"/publish", nil)

		// Create response recorder
		w := httptest.NewRecorder()

		// Perform request
		router.ServeHTTP(w, req)

		// Assert response
		assert.Equal(t, http.StatusNotFound, w.Code)
		assert.Contains(t, w.Body.String(), "form design not found")

		// Assert mock expectations
		mockService.AssertExpectations(t)
	})

	// Test service error - form already published
	t.Run("FormAlreadyPublished", func(t *testing.T) {
		// Prepare test data
		formID := "form_123"

		// Mock service response
		mockService.On("PublishFormDesign", mock.Anything, formID).Return(testutils.NewError("form design is already published")).Once()

		// Create request
		req, _ := http.NewRequest(http.MethodPost, "/forms/"+formID+"/publish", nil)

		// Create response recorder
		w := httptest.NewRecorder()

		// Perform request
		router.ServeHTTP(w, req)

		// Assert response
		assert.Equal(t, http.StatusBadRequest, w.Code)
		assert.Contains(t, w.Body.String(), "form design is already published")

		// Assert mock expectations
		mockService.AssertExpectations(t)
	})

	// Test service error - internal error
	t.Run("InternalServerError", func(t *testing.T) {
		// Prepare test data
		formID := "form_123"

		// Mock service response
		mockService.On("PublishFormDesign", mock.Anything, formID).Return(testutils.NewError("internal error")).Once()

		// Create request
		req, _ := http.NewRequest(http.MethodPost, "/forms/"+formID+"/publish", nil)

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