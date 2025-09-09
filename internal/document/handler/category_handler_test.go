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

// MockCategoryService is a mock implementation of CategoryServiceInterface
type MockCategoryService struct {
	mock.Mock
}

func (m *MockCategoryService) CreateCategory(ctx context.Context, name, description, parentID string) (*domain.DocumentCategory, error) {
	args := m.Called(ctx, name, description, parentID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.DocumentCategory), args.Error(1)
}

func (m *MockCategoryService) GetCategory(ctx context.Context, categoryID string) (*domain.DocumentCategory, error) {
	args := m.Called(ctx, categoryID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.DocumentCategory), args.Error(1)
}

func (m *MockCategoryService) UpdateCategory(ctx context.Context, categoryID, name, description string) error {
	args := m.Called(ctx, categoryID, name, description)
	return args.Error(0)
}

func (m *MockCategoryService) DeleteCategory(ctx context.Context, categoryID string) error {
	args := m.Called(ctx, categoryID)
	return args.Error(0)
}

func (m *MockCategoryService) ListCategories(ctx context.Context, parentID string) ([]*domain.DocumentCategory, error) {
	args := m.Called(ctx, parentID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*domain.DocumentCategory), args.Error(1)
}

func (m *MockCategoryService) AssignDocumentToCategory(ctx context.Context, documentID, categoryID string) error {
	args := m.Called(ctx, documentID, categoryID)
	return args.Error(0)
}

func (m *MockCategoryService) RemoveDocumentFromCategory(ctx context.Context, documentID, categoryID string) error {
	args := m.Called(ctx, documentID, categoryID)
	return args.Error(0)
}

func (m *MockCategoryService) GetDocumentCategories(ctx context.Context, documentID string) ([]*domain.DocumentCategory, error) {
	args := m.Called(ctx, documentID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*domain.DocumentCategory), args.Error(1)
}

// TestNewCategoryHandler tests the NewCategoryHandler function
func TestNewCategoryHandler(t *testing.T) {
	handler := NewCategoryHandler()
	assert.NotNil(t, handler)
	assert.NotNil(t, handler.categoryService)
}

// TestCreateCategory tests the CreateCategory handler
func TestCreateCategory(t *testing.T) {
	// Set up test environment
	gin.SetMode(gin.TestMode)

	// Create mock service
	mockService := new(MockCategoryService)

	// Create handler with mock service
	handler := &CategoryHandler{
		categoryService: mockService,
	}

	// Create test router
	router := gin.New()
	router.POST("/categories", handler.CreateCategory)

	// Test successful creation
	t.Run("SuccessfulCreation", func(t *testing.T) {
		// Prepare test data
		reqBody := CreateCategoryRequest{
			Name:        "Test Category",
			Description: "Test Description",
			ParentID:    "parent_123",
		}

		// Mock data
		expectedCategory := &domain.DocumentCategory{
			ID:          "cat_123",
			Name:        "Test Category",
			Description: "Test Description",
			ParentID:    "parent_123",
		}

		// Mock service response
		mockService.On("CreateCategory", mock.Anything, "Test Category", "Test Description", "parent_123").Return(expectedCategory, nil).Once()

		// Create request
		jsonValue, _ := json.Marshal(reqBody)
		req, _ := http.NewRequest(http.MethodPost, "/categories", bytes.NewBuffer(jsonValue))
		req.Header.Set("Content-Type", "application/json")

		// Create response recorder
		w := httptest.NewRecorder()

		// Perform request
		router.ServeHTTP(w, req)

		// Assert response
		assert.Equal(t, http.StatusOK, w.Code)

		// Parse response
		var response domain.DocumentCategory
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.Equal(t, expectedCategory.ID, response.ID)
		assert.Equal(t, expectedCategory.Name, response.Name)

		// Assert mock expectations
		mockService.AssertExpectations(t)
	})

	// Test invalid request body
	t.Run("InvalidRequestBody", func(t *testing.T) {
		// Create request with invalid JSON
		req, _ := http.NewRequest(http.MethodPost, "/categories", bytes.NewBufferString("{invalid json}"))
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
		reqBody := CreateCategoryRequest{
			Name:        "Test Category",
			Description: "Test Description",
		}

		// Mock service response
		mockService.On("CreateCategory", mock.Anything, "Test Category", "Test Description", "").Return((*domain.DocumentCategory)(nil), testutils.NewError("internal error")).Once()

		// Create request
		jsonValue, _ := json.Marshal(reqBody)
		req, _ := http.NewRequest(http.MethodPost, "/categories", bytes.NewBuffer(jsonValue))
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

// TestGetCategory tests the GetCategory handler
func TestGetCategory(t *testing.T) {
	// Set up test environment
	gin.SetMode(gin.TestMode)

	// Create mock service
	mockService := new(MockCategoryService)

	// Create handler with mock service
	handler := &CategoryHandler{
		categoryService: mockService,
	}

	// Create test router with route parameter
	router := gin.New()
	router.GET("/categories/:id", handler.GetCategory)

	// Test successful retrieval
	t.Run("SuccessfulRetrieval", func(t *testing.T) {
		// Prepare test data
		categoryID := "cat_123"
		expectedCategory := &domain.DocumentCategory{
			ID:          categoryID,
			Name:        "Test Category",
			Description: "Test Description",
			ParentID:    "parent_123",
		}

		// Mock service response
		mockService.On("GetCategory", mock.Anything, categoryID).Return(expectedCategory, nil).Once()

		// Create request
		req, _ := http.NewRequest(http.MethodGet, "/categories/"+categoryID, nil)

		// Create response recorder
		w := httptest.NewRecorder()

		// Perform request
		router.ServeHTTP(w, req)

		// Assert response
		assert.Equal(t, http.StatusOK, w.Code)

		// Parse response
		var response domain.DocumentCategory
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.Equal(t, expectedCategory.ID, response.ID)
		assert.Equal(t, expectedCategory.Name, response.Name)

		// Assert mock expectations
		mockService.AssertExpectations(t)
	})

	// Test missing category ID
	t.Run("MissingCategoryID", func(t *testing.T) {
		// Create request without category ID
		req, _ := http.NewRequest(http.MethodGet, "/categories/", nil)

		// Create response recorder
		w := httptest.NewRecorder()

		// Perform request
		router.ServeHTTP(w, req)

		// Assert response
		assert.Equal(t, http.StatusNotFound, w.Code) // Changed to StatusNotFound since the route doesn't match
	})

	// Test service error - category not found
	t.Run("CategoryNotFound", func(t *testing.T) {
		// Prepare test data
		categoryID := "cat_456"

		// Mock service response
		mockService.On("GetCategory", mock.Anything, categoryID).Return((*domain.DocumentCategory)(nil), testutils.NewError("category not found")).Once()

		// Create request
		req, _ := http.NewRequest(http.MethodGet, "/categories/"+categoryID, nil)

		// Create response recorder
		w := httptest.NewRecorder()

		// Perform request
		router.ServeHTTP(w, req)

		// Assert response
		assert.Equal(t, http.StatusNotFound, w.Code)
		assert.Contains(t, w.Body.String(), "category not found")

		// Assert mock expectations
		mockService.AssertExpectations(t)
	})

	// Test service error - internal error
	t.Run("InternalServerError", func(t *testing.T) {
		// Prepare test data
		categoryID := "cat_123"

		// Mock service response
		mockService.On("GetCategory", mock.Anything, categoryID).Return((*domain.DocumentCategory)(nil), testutils.NewError("internal error")).Once()

		// Create request
		req, _ := http.NewRequest(http.MethodGet, "/categories/"+categoryID, nil)

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

// TestUpdateCategory tests the UpdateCategory handler
func TestUpdateCategory(t *testing.T) {
	// Set up test environment
	gin.SetMode(gin.TestMode)

	// Create mock service
	mockService := new(MockCategoryService)

	// Create handler with mock service
	handler := &CategoryHandler{
		categoryService: mockService,
	}

	// Create test router with route parameter
	router := gin.New()
	router.PUT("/categories/:id", handler.UpdateCategory)

	// Test successful update
	t.Run("SuccessfulUpdate", func(t *testing.T) {
		// Prepare test data
		categoryID := "cat_123"
		reqBody := UpdateCategoryRequest{
			Name:        "Updated Category",
			Description: "Updated Description",
		}

		// Mock service response
		mockService.On("UpdateCategory", mock.Anything, categoryID, "Updated Category", "Updated Description").Return(nil).Once()

		// Create request
		jsonValue, _ := json.Marshal(reqBody)
		req, _ := http.NewRequest(http.MethodPut, "/categories/"+categoryID, bytes.NewBuffer(jsonValue))
		req.Header.Set("Content-Type", "application/json")

		// Create response recorder
		w := httptest.NewRecorder()

		// Perform request
		router.ServeHTTP(w, req)

		// Assert response
		assert.Equal(t, http.StatusOK, w.Code)
		assert.Contains(t, w.Body.String(), "category updated successfully")

		// Assert mock expectations
		mockService.AssertExpectations(t)
	})

	// Test missing category ID
	t.Run("MissingCategoryID", func(t *testing.T) {
		// Create request without category ID
		reqBody := UpdateCategoryRequest{
			Name: "Updated Category",
		}

		jsonValue, _ := json.Marshal(reqBody)
		req, _ := http.NewRequest(http.MethodPut, "/categories/", bytes.NewBuffer(jsonValue))
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
		req, _ := http.NewRequest(http.MethodPut, "/categories/cat_123", bytes.NewBufferString("{invalid json}"))
		req.Header.Set("Content-Type", "application/json")

		// Create response recorder
		w := httptest.NewRecorder()

		// Perform request
		router.ServeHTTP(w, req)

		// Assert response
		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	// Test service error - category not found
	t.Run("CategoryNotFound", func(t *testing.T) {
		// Prepare test data
		categoryID := "cat_456"
		reqBody := UpdateCategoryRequest{
			Name: "Updated Category",
		}

		// Mock service response
		mockService.On("UpdateCategory", mock.Anything, categoryID, "Updated Category", "").Return(testutils.NewError("category not found")).Once()

		// Create request
		jsonValue, _ := json.Marshal(reqBody)
		req, _ := http.NewRequest(http.MethodPut, "/categories/"+categoryID, bytes.NewBuffer(jsonValue))
		req.Header.Set("Content-Type", "application/json")

		// Create response recorder
		w := httptest.NewRecorder()

		// Perform request
		router.ServeHTTP(w, req)

		// Assert response
		assert.Equal(t, http.StatusNotFound, w.Code)
		assert.Contains(t, w.Body.String(), "category not found")

		// Assert mock expectations
		mockService.AssertExpectations(t)
	})

	// Test service error - internal error
	t.Run("InternalServerError", func(t *testing.T) {
		// Prepare test data
		categoryID := "cat_123"
		reqBody := UpdateCategoryRequest{
			Name: "Updated Category",
		}

		// Mock service response
		mockService.On("UpdateCategory", mock.Anything, categoryID, "Updated Category", "").Return(testutils.NewError("internal error")).Once()

		// Create request
		jsonValue, _ := json.Marshal(reqBody)
		req, _ := http.NewRequest(http.MethodPut, "/categories/"+categoryID, bytes.NewBuffer(jsonValue))
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

// TestDeleteCategory tests the DeleteCategory handler
func TestDeleteCategory(t *testing.T) {
	// Set up test environment
	gin.SetMode(gin.TestMode)

	// Create mock service
	mockService := new(MockCategoryService)

	// Create handler with mock service
	handler := &CategoryHandler{
		categoryService: mockService,
	}

	// Create test router with route parameter
	router := gin.New()
	router.DELETE("/categories/:id", handler.DeleteCategory)

	// Test successful deletion
	t.Run("SuccessfulDeletion", func(t *testing.T) {
		// Prepare test data
		categoryID := "cat_123"

		// Mock service response
		mockService.On("DeleteCategory", mock.Anything, categoryID).Return(nil).Once()

		// Create request
		req, _ := http.NewRequest(http.MethodDelete, "/categories/"+categoryID, nil)

		// Create response recorder
		w := httptest.NewRecorder()

		// Perform request
		router.ServeHTTP(w, req)

		// Assert response
		assert.Equal(t, http.StatusOK, w.Code)
		assert.Contains(t, w.Body.String(), "category deleted successfully")

		// Assert mock expectations
		mockService.AssertExpectations(t)
	})

	// Test missing category ID
	t.Run("MissingCategoryID", func(t *testing.T) {
		// Create request without category ID
		req, _ := http.NewRequest(http.MethodDelete, "/categories/", nil)

		// Create response recorder
		w := httptest.NewRecorder()

		// Perform request
		router.ServeHTTP(w, req)

		// Assert response
		assert.Equal(t, http.StatusNotFound, w.Code) // Changed to StatusNotFound since the route doesn't match
	})

	// Test service error - category not found
	t.Run("CategoryNotFound", func(t *testing.T) {
		// Prepare test data
		categoryID := "cat_456"

		// Mock service response
		mockService.On("DeleteCategory", mock.Anything, categoryID).Return(testutils.NewError("category not found")).Once()

		// Create request
		req, _ := http.NewRequest(http.MethodDelete, "/categories/"+categoryID, nil)

		// Create response recorder
		w := httptest.NewRecorder()

		// Perform request
		router.ServeHTTP(w, req)

		// Assert response
		assert.Equal(t, http.StatusNotFound, w.Code)
		assert.Contains(t, w.Body.String(), "category not found")

		// Assert mock expectations
		mockService.AssertExpectations(t)
	})

	// Test service error - cannot delete category with child categories
	t.Run("CannotDeleteWithChildCategories", func(t *testing.T) {
		// Prepare test data
		categoryID := "cat_123"

		// Mock service response
		mockService.On("DeleteCategory", mock.Anything, categoryID).Return(testutils.NewError("cannot delete category with child categories")).Once()

		// Create request
		req, _ := http.NewRequest(http.MethodDelete, "/categories/"+categoryID, nil)

		// Create response recorder
		w := httptest.NewRecorder()

		// Perform request
		router.ServeHTTP(w, req)

		// Assert response
		assert.Equal(t, http.StatusBadRequest, w.Code)
		assert.Contains(t, w.Body.String(), "cannot delete category with child categories")

		// Assert mock expectations
		mockService.AssertExpectations(t)
	})

	// Test service error - internal error
	t.Run("InternalServerError", func(t *testing.T) {
		// Prepare test data
		categoryID := "cat_123"

		// Mock service response
		mockService.On("DeleteCategory", mock.Anything, categoryID).Return(testutils.NewError("internal error")).Once()

		// Create request
		req, _ := http.NewRequest(http.MethodDelete, "/categories/"+categoryID, nil)

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

// TestListCategories tests the ListCategories handler
func TestListCategories(t *testing.T) {
	// Set up test environment
	gin.SetMode(gin.TestMode)

	// Create mock service
	mockService := new(MockCategoryService)

	// Create handler with mock service
	handler := &CategoryHandler{
		categoryService: mockService,
	}

	// Create test router
	router := gin.New()
	router.GET("/categories", handler.ListCategories)

	// Test successful listing
	t.Run("SuccessfulListing", func(t *testing.T) {
		// Prepare test data
		expectedCategories := []*domain.DocumentCategory{
			{
				ID:          "cat_123",
				Name:        "Category 1",
				Description: "Description 1",
				ParentID:    "",
			},
			{
				ID:          "cat_456",
				Name:        "Category 2",
				Description: "Description 2",
				ParentID:    "",
			},
		}

		// Mock service response
		mockService.On("ListCategories", mock.Anything, "").Return(expectedCategories, nil).Once()

		// Create request
		req, _ := http.NewRequest(http.MethodGet, "/categories", nil)

		// Create response recorder
		w := httptest.NewRecorder()

		// Perform request
		router.ServeHTTP(w, req)

		// Assert response
		assert.Equal(t, http.StatusOK, w.Code)

		// Parse response
		var response []*domain.DocumentCategory
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.Equal(t, expectedCategories, response)

		// Assert mock expectations
		mockService.AssertExpectations(t)
	})

	// Test successful listing with parent_id
	t.Run("SuccessfulListingWithParentID", func(t *testing.T) {
		// Prepare test data
		parentID := "parent_123"
		expectedCategories := []*domain.DocumentCategory{
			{
				ID:          "cat_123",
				Name:        "Subcategory 1",
				Description: "Subcategory Description 1",
				ParentID:    parentID,
			},
		}

		// Mock service response
		mockService.On("ListCategories", mock.Anything, parentID).Return(expectedCategories, nil).Once()

		// Create request
		req, _ := http.NewRequest(http.MethodGet, "/categories?parent_id="+parentID, nil)

		// Create response recorder
		w := httptest.NewRecorder()

		// Perform request
		router.ServeHTTP(w, req)

		// Assert response
		assert.Equal(t, http.StatusOK, w.Code)

		// Parse response
		var response []*domain.DocumentCategory
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.Equal(t, expectedCategories, response)

		// Assert mock expectations
		mockService.AssertExpectations(t)
	})

	// Test service error
	t.Run("ServiceError", func(t *testing.T) {
		// Mock service response
		mockService.On("ListCategories", mock.Anything, "").Return([]*domain.DocumentCategory(nil), testutils.NewError("internal error")).Once()

		// Create request
		req, _ := http.NewRequest(http.MethodGet, "/categories", nil)

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

// TestAssignDocumentToCategory tests the AssignDocumentToCategory handler
func TestAssignDocumentToCategory(t *testing.T) {
	// Set up test environment
	gin.SetMode(gin.TestMode)

	// Create mock service
	mockService := new(MockCategoryService)

	// Create handler with mock service
	handler := &CategoryHandler{
		categoryService: mockService,
	}

	// Create test router
	router := gin.New()
	router.POST("/categories/assign", handler.AssignDocumentToCategory)

	// Test successful assignment
	t.Run("SuccessfulAssignment", func(t *testing.T) {
		// Prepare test data
		reqBody := AssignDocumentToCategoryRequest{
			DocumentID: "doc_123",
			CategoryID: "cat_123",
		}

		// Mock service response
		mockService.On("AssignDocumentToCategory", mock.Anything, "doc_123", "cat_123").Return(nil).Once()

		// Create request
		jsonValue, _ := json.Marshal(reqBody)
		req, _ := http.NewRequest(http.MethodPost, "/categories/assign", bytes.NewBuffer(jsonValue))
		req.Header.Set("Content-Type", "application/json")

		// Create response recorder
		w := httptest.NewRecorder()

		// Perform request
		router.ServeHTTP(w, req)

		// Assert response
		assert.Equal(t, http.StatusOK, w.Code)
		assert.Contains(t, w.Body.String(), "document assigned to category successfully")

		// Assert mock expectations
		mockService.AssertExpectations(t)
	})

	// Test invalid request body
	t.Run("InvalidRequestBody", func(t *testing.T) {
		// Create request with invalid JSON
		req, _ := http.NewRequest(http.MethodPost, "/categories/assign", bytes.NewBufferString("{invalid json}"))
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
		reqBody := AssignDocumentToCategoryRequest{
			DocumentID: "doc_456",
			CategoryID: "cat_123",
		}

		// Mock service response
		mockService.On("AssignDocumentToCategory", mock.Anything, "doc_456", "cat_123").Return(testutils.NewError("document not found")).Once()

		// Create request
		jsonValue, _ := json.Marshal(reqBody)
		req, _ := http.NewRequest(http.MethodPost, "/categories/assign", bytes.NewBuffer(jsonValue))
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

	// Test service error - category not found
	t.Run("CategoryNotFound", func(t *testing.T) {
		// Prepare test data
		reqBody := AssignDocumentToCategoryRequest{
			DocumentID: "doc_123",
			CategoryID: "cat_456",
		}

		// Mock service response
		mockService.On("AssignDocumentToCategory", mock.Anything, "doc_123", "cat_456").Return(testutils.NewError("category not found")).Once()

		// Create request
		jsonValue, _ := json.Marshal(reqBody)
		req, _ := http.NewRequest(http.MethodPost, "/categories/assign", bytes.NewBuffer(jsonValue))
		req.Header.Set("Content-Type", "application/json")

		// Create response recorder
		w := httptest.NewRecorder()

		// Perform request
		router.ServeHTTP(w, req)

		// Assert response
		assert.Equal(t, http.StatusNotFound, w.Code)
		assert.Contains(t, w.Body.String(), "category not found")

		// Assert mock expectations
		mockService.AssertExpectations(t)
	})

	// Test service error - internal error
	t.Run("InternalServerError", func(t *testing.T) {
		// Prepare test data
		reqBody := AssignDocumentToCategoryRequest{
			DocumentID: "doc_123",
			CategoryID: "cat_123",
		}

		// Mock service response
		mockService.On("AssignDocumentToCategory", mock.Anything, "doc_123", "cat_123").Return(testutils.NewError("internal error")).Once()

		// Create request
		jsonValue, _ := json.Marshal(reqBody)
		req, _ := http.NewRequest(http.MethodPost, "/categories/assign", bytes.NewBuffer(jsonValue))
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

// TestRemoveDocumentFromCategory tests the RemoveDocumentFromCategory handler
func TestRemoveDocumentFromCategory(t *testing.T) {
	// Set up test environment
	gin.SetMode(gin.TestMode)

	// Create mock service
	mockService := new(MockCategoryService)

	// Create handler with mock service
	handler := &CategoryHandler{
		categoryService: mockService,
	}

	// Create test router
	router := gin.New()
	router.POST("/categories/remove", handler.RemoveDocumentFromCategory)

	// Test successful removal
	t.Run("SuccessfulRemoval", func(t *testing.T) {
		// Prepare test data
		reqBody := RemoveDocumentFromCategoryRequest{
			DocumentID: "doc_123",
			CategoryID: "cat_123",
		}

		// Mock service response
		mockService.On("RemoveDocumentFromCategory", mock.Anything, "doc_123", "cat_123").Return(nil).Once()

		// Create request
		jsonValue, _ := json.Marshal(reqBody)
		req, _ := http.NewRequest(http.MethodPost, "/categories/remove", bytes.NewBuffer(jsonValue))
		req.Header.Set("Content-Type", "application/json")

		// Create response recorder
		w := httptest.NewRecorder()

		// Perform request
		router.ServeHTTP(w, req)

		// Assert response
		assert.Equal(t, http.StatusOK, w.Code)
		assert.Contains(t, w.Body.String(), "document removed from category successfully")

		// Assert mock expectations
		mockService.AssertExpectations(t)
	})

	// Test invalid request body
	t.Run("InvalidRequestBody", func(t *testing.T) {
		// Create request with invalid JSON
		req, _ := http.NewRequest(http.MethodPost, "/categories/remove", bytes.NewBufferString("{invalid json}"))
		req.Header.Set("Content-Type", "application/json")

		// Create response recorder
		w := httptest.NewRecorder()

		// Perform request
		router.ServeHTTP(w, req)

		// Assert response
		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	// Test service error - document-category relation not found
	t.Run("RelationNotFound", func(t *testing.T) {
		// Prepare test data
		reqBody := RemoveDocumentFromCategoryRequest{
			DocumentID: "doc_123",
			CategoryID: "cat_456",
		}

		// Mock service response
		mockService.On("RemoveDocumentFromCategory", mock.Anything, "doc_123", "cat_456").Return(testutils.NewError("document-category relation not found")).Once()

		// Create request
		jsonValue, _ := json.Marshal(reqBody)
		req, _ := http.NewRequest(http.MethodPost, "/categories/remove", bytes.NewBuffer(jsonValue))
		req.Header.Set("Content-Type", "application/json")

		// Create response recorder
		w := httptest.NewRecorder()

		// Perform request
		router.ServeHTTP(w, req)

		// Assert response
		assert.Equal(t, http.StatusNotFound, w.Code)
		assert.Contains(t, w.Body.String(), "document-category relation not found")

		// Assert mock expectations
		mockService.AssertExpectations(t)
	})

	// Test service error - internal error
	t.Run("InternalServerError", func(t *testing.T) {
		// Prepare test data
		reqBody := RemoveDocumentFromCategoryRequest{
			DocumentID: "doc_123",
			CategoryID: "cat_123",
		}

		// Mock service response
		mockService.On("RemoveDocumentFromCategory", mock.Anything, "doc_123", "cat_123").Return(testutils.NewError("internal error")).Once()

		// Create request
		jsonValue, _ := json.Marshal(reqBody)
		req, _ := http.NewRequest(http.MethodPost, "/categories/remove", bytes.NewBuffer(jsonValue))
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

// TestGetDocumentCategories tests the GetDocumentCategories handler
func TestGetDocumentCategories(t *testing.T) {
	// Set up test environment
	gin.SetMode(gin.TestMode)

	// Create mock service
	mockService := new(MockCategoryService)

	// Create handler with mock service
	handler := &CategoryHandler{
		categoryService: mockService,
	}

	// Create test router with route parameter
	router := gin.New()
	router.GET("/categories/document/:document_id", handler.GetDocumentCategories)

	// Test successful retrieval
	t.Run("SuccessfulRetrieval", func(t *testing.T) {
		// Prepare test data
		documentID := "doc_123"
		expectedCategories := []*domain.DocumentCategory{
			{
				ID:          "cat_123",
				Name:        "Category 1",
				Description: "Description 1",
				ParentID:    "",
			},
			{
				ID:          "cat_456",
				Name:        "Category 2",
				Description: "Description 2",
				ParentID:    "",
			},
		}

		// Mock service response
		mockService.On("GetDocumentCategories", mock.Anything, documentID).Return(expectedCategories, nil).Once()

		// Create request
		req, _ := http.NewRequest(http.MethodGet, "/categories/document/"+documentID, nil)

		// Create response recorder
		w := httptest.NewRecorder()

		// Perform request
		router.ServeHTTP(w, req)

		// Assert response
		assert.Equal(t, http.StatusOK, w.Code)

		// Parse response
		var response []*domain.DocumentCategory
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.Equal(t, expectedCategories, response)

		// Assert mock expectations
		mockService.AssertExpectations(t)
	})

	// Test missing document ID
	t.Run("MissingDocumentID", func(t *testing.T) {
		// Create request without document ID
		req, _ := http.NewRequest(http.MethodGet, "/categories/document/", nil)

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
		mockService.On("GetDocumentCategories", mock.Anything, documentID).Return([]*domain.DocumentCategory(nil), testutils.NewError("document not found")).Once()

		// Create request
		req, _ := http.NewRequest(http.MethodGet, "/categories/document/"+documentID, nil)

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
		mockService.On("GetDocumentCategories", mock.Anything, documentID).Return([]*domain.DocumentCategory(nil), testutils.NewError("internal error")).Once()

		// Create request
		req, _ := http.NewRequest(http.MethodGet, "/categories/document/"+documentID, nil)

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