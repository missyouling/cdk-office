package handler

import (
	"bytes"
	"context"
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

// MockFormService is a mock implementation of FormServiceInterface
type MockFormService struct {
	mock.Mock
}

func (m *MockFormService) CreateForm(ctx context.Context, req *service.CreateFormRequest) (*domain.FormData, error) {
	args := m.Called(ctx, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.FormData), args.Error(1)
}

func (m *MockFormService) UpdateForm(ctx context.Context, formID string, req *service.UpdateFormRequest) error {
	args := m.Called(ctx, formID, req)
	return args.Error(0)
}

func (m *MockFormService) DeleteForm(ctx context.Context, formID string) error {
	args := m.Called(ctx, formID)
	return args.Error(0)
}

func (m *MockFormService) ListForms(ctx context.Context, appID string, page, size int) ([]*domain.FormData, int64, error) {
	args := m.Called(ctx, appID, page, size)
	if args.Get(0) == nil {
		return nil, 0, args.Error(2)
	}
	return args.Get(0).([]*domain.FormData), args.Get(1).(int64), args.Error(2)
}

func (m *MockFormService) GetForm(ctx context.Context, formID string) (*domain.FormData, error) {
	args := m.Called(ctx, formID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.FormData), args.Error(1)
}

func (m *MockFormService) SubmitFormData(ctx context.Context, req *service.SubmitFormDataRequest) (*domain.FormDataEntry, error) {
	args := m.Called(ctx, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.FormDataEntry), args.Error(1)
}

func (m *MockFormService) ListFormDataEntries(ctx context.Context, formID string, page, size int) ([]*domain.FormDataEntry, int64, error) {
	args := m.Called(ctx, formID, page, size)
	if args.Get(0) == nil {
		return nil, 0, args.Error(2)
	}
	return args.Get(0).([]*domain.FormDataEntry), args.Get(1).(int64), args.Error(2)
}

func TestNewFormHandler(t *testing.T) {
	handler := NewFormHandler()
	assert.NotNil(t, handler)
	assert.NotNil(t, handler.formService)
}

func TestFormHandler_CreateForm(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("successful creation", func(t *testing.T) {
		// Setup
		mockService := new(MockFormService)
		handler := &FormHandler{formService: mockService}
		
		// Mock data
		form := &domain.FormData{
			ID:          "form_123",
			AppID:       "app_123",
			Name:        "Test Form",
			Description: "Test Description",
			Schema:      "{}",
			IsActive:    true,
			CreatedBy:   "user_123",
			CreatedAt:   time.Now(),
			UpdatedAt:   time.Now(),
		}
		
		// Mock service
		mockService.On("CreateForm", mock.Anything, mock.MatchedBy(func(req *service.CreateFormRequest) bool {
			return req.AppID == "app_123" && req.Name == "Test Form"
		})).Return(form, nil)
		
		// Create request
		reqBody := `{"app_id":"app_123","name":"Test Form","description":"Test Description","schema":"{}","created_by":"user_123"}`
		req, _ := http.NewRequest(http.MethodPost, "/forms", bytes.NewBufferString(reqBody))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		
		// Create context and call handler
		c, _ := gin.CreateTestContext(w)
		c.Request = req
		
		handler.CreateForm(c)
		
		// Assertions
		assert.Equal(t, http.StatusOK, w.Code)
		mockService.AssertExpectations(t)
	})
	
	t.Run("invalid request body", func(t *testing.T) {
		// Setup
		handler := &FormHandler{}
		
		// Create request with invalid JSON
		reqBody := `{"app_id":"app_123","name":"Test Form","description":"Test Description","schema":"{}"` // Missing closing brace and missing created_by
		req, _ := http.NewRequest(http.MethodPost, "/forms", bytes.NewBufferString(reqBody))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		
		// Create context and call handler
		c, _ := gin.CreateTestContext(w)
		c.Request = req
		
		handler.CreateForm(c)
		
		// Assertions
		assert.Equal(t, http.StatusBadRequest, w.Code)
	})
	
	t.Run("service error", func(t *testing.T) {
		// Setup
		mockService := new(MockFormService)
		handler := &FormHandler{formService: mockService}
		
		// Mock service to return error
		mockService.On("CreateForm", mock.Anything, mock.MatchedBy(func(req *service.CreateFormRequest) bool {
			return req.AppID == "app_123" && req.Name == "Test Form"
		})).Return((*domain.FormData)(nil), testutils.NewError("service error"))
		
		// Create request
		reqBody := `{"app_id":"app_123","name":"Test Form","description":"Test Description","schema":"{}","created_by":"user_123"}`
		req, _ := http.NewRequest(http.MethodPost, "/forms", bytes.NewBufferString(reqBody))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		
		// Create context and call handler
		c, _ := gin.CreateTestContext(w)
		c.Request = req
		
		handler.CreateForm(c)
		
		// Assertions
		assert.Equal(t, http.StatusInternalServerError, w.Code)
		mockService.AssertExpectations(t)
	})
}

func TestFormHandler_UpdateForm(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("successful update", func(t *testing.T) {
		// Setup
		mockService := new(MockFormService)
		handler := &FormHandler{formService: mockService}
		
		// Mock service
		mockService.On("UpdateForm", mock.Anything, "form_123", mock.MatchedBy(func(req *service.UpdateFormRequest) bool {
			return req.Name != "" && req.Description != ""
		})).Return(nil)
		
		// Create request
		reqBody := `{"name":"Updated Form","description":"Updated Description"}`
		req, _ := http.NewRequest(http.MethodPut, "/forms/form_123", bytes.NewBufferString(reqBody))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		
		// Create context and call handler
		c, _ := gin.CreateTestContext(w)
		c.Request = req
		c.AddParam("id", "form_123")
		
		handler.UpdateForm(c)
		
		// Assertions
		assert.Equal(t, http.StatusOK, w.Code)
		mockService.AssertExpectations(t)
	})
	
	t.Run("missing form id", func(t *testing.T) {
		// Setup
		handler := &FormHandler{}
		
		// Create request without form ID
		reqBody := `{}`
		req, _ := http.NewRequest(http.MethodPut, "/forms/", bytes.NewBufferString(reqBody))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		
		// Create context and call handler
		c, _ := gin.CreateTestContext(w)
		c.Request = req
		
		handler.UpdateForm(c)
		
		// Assertions
		assert.Equal(t, http.StatusBadRequest, w.Code)
	})
	
	t.Run("form not found", func(t *testing.T) {
		// Setup
		mockService := new(MockFormService)
		handler := &FormHandler{formService: mockService}
		
		// Mock service to return "form not found" error
		mockService.On("UpdateForm", mock.Anything, "form_123", mock.MatchedBy(func(req *service.UpdateFormRequest) bool {
			return req.Name == "Updated Form"
		})).Return(testutils.NewError("form not found"))
		
		// Create request
		reqBody := `{"name":"Updated Form"}`
		req, _ := http.NewRequest(http.MethodPut, "/forms/form_123", bytes.NewBufferString(reqBody))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		
		// Create context and call handler
		c, _ := gin.CreateTestContext(w)
		c.Request = req
		c.AddParam("id", "form_123")
		
		handler.UpdateForm(c)
		
		// Assertions
		assert.Equal(t, http.StatusNotFound, w.Code)
		mockService.AssertExpectations(t)
	})
	
	t.Run("service error", func(t *testing.T) {
		// Setup
		mockService := new(MockFormService)
		handler := &FormHandler{formService: mockService}
		
		// Mock service to return error
		mockService.On("UpdateForm", mock.Anything, "form_123", mock.MatchedBy(func(req *service.UpdateFormRequest) bool {
			return req.Name == "Updated Form"
		})).Return(testutils.NewError("service error"))
		
		// Create request
		reqBody := `{"name":"Updated Form"}`
		req, _ := http.NewRequest(http.MethodPut, "/forms/form_123", bytes.NewBufferString(reqBody))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		
		// Create context and call handler
		c, _ := gin.CreateTestContext(w)
		c.Request = req
		c.AddParam("id", "form_123")
		
		handler.UpdateForm(c)
		
		// Assertions
		assert.Equal(t, http.StatusInternalServerError, w.Code)
		mockService.AssertExpectations(t)
	})
}

func TestFormHandler_DeleteForm(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("successful deletion", func(t *testing.T) {
		// Setup
		mockService := new(MockFormService)
		handler := &FormHandler{formService: mockService}
		
		// Mock service
		mockService.On("DeleteForm", mock.Anything, "form_123").Return(nil)
		
		// Create request
		req, _ := http.NewRequest(http.MethodDelete, "/forms/form_123", nil)
		w := httptest.NewRecorder()
		
		// Create context and call handler
		c, _ := gin.CreateTestContext(w)
		c.Request = req
		c.AddParam("id", "form_123")
		
		handler.DeleteForm(c)
		
		// Assertions
		assert.Equal(t, http.StatusOK, w.Code)
		mockService.AssertExpectations(t)
	})
	
	t.Run("missing form id", func(t *testing.T) {
		// Setup
		handler := &FormHandler{}
		
		// Create request without form ID
		req, _ := http.NewRequest(http.MethodDelete, "/forms/", nil)
		w := httptest.NewRecorder()
		
		// Create context and call handler
		c, _ := gin.CreateTestContext(w)
		c.Request = req
		
		handler.DeleteForm(c)
		
		// Assertions
		assert.Equal(t, http.StatusBadRequest, w.Code)
	})
	
	t.Run("form not found", func(t *testing.T) {
		// Setup
		mockService := new(MockFormService)
		handler := &FormHandler{formService: mockService}
		
		// Mock service to return "form not found" error
		mockService.On("DeleteForm", mock.Anything, "form_123").Return(testutils.NewError("form not found"))
		
		// Create request
		req, _ := http.NewRequest(http.MethodDelete, "/forms/form_123", nil)
		w := httptest.NewRecorder()
		
		// Create context and call handler
		c, _ := gin.CreateTestContext(w)
		c.Request = req
		c.AddParam("id", "form_123")
		
		handler.DeleteForm(c)
		
		// Assertions
		assert.Equal(t, http.StatusNotFound, w.Code)
		mockService.AssertExpectations(t)
	})
	
	t.Run("service error", func(t *testing.T) {
		// Setup
		mockService := new(MockFormService)
		handler := &FormHandler{formService: mockService}
		
		// Mock service to return error
		mockService.On("DeleteForm", mock.Anything, "form_123").Return(testutils.NewError("service error"))
		
		// Create request
		req, _ := http.NewRequest(http.MethodDelete, "/forms/form_123", nil)
		w := httptest.NewRecorder()
		
		// Create context and call handler
		c, _ := gin.CreateTestContext(w)
		c.Request = req
		c.AddParam("id", "form_123")
		
		handler.DeleteForm(c)
		
		// Assertions
		assert.Equal(t, http.StatusInternalServerError, w.Code)
		mockService.AssertExpectations(t)
	})
}

func TestFormHandler_ListForms(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("successful listing", func(t *testing.T) {
		// Setup
		mockService := new(MockFormService)
		handler := &FormHandler{formService: mockService}
		
		// Mock data
		forms := []*domain.FormData{
			{
				ID:          "form_1",
				AppID:       "app_123",
				Name:        "Form 1",
				Description: "Description 1",
				Schema:      "{}",
				IsActive:    true,
				CreatedBy:   "user_123",
				CreatedAt:   time.Now(),
				UpdatedAt:   time.Now(),
			},
			{
				ID:          "form_2",
				AppID:       "app_123",
				Name:        "Form 2",
				Description: "Description 2",
				Schema:      "{}",
				IsActive:    true,
				CreatedBy:   "user_123",
				CreatedAt:   time.Now(),
				UpdatedAt:   time.Now(),
			},
		}
		
		// Mock service
		mockService.On("ListForms", mock.Anything, "app_123", 1, 10).Return(forms, int64(2), nil)
		
		// Create request
		req, _ := http.NewRequest(http.MethodGet, "/forms?app_id=app_123", nil)
		w := httptest.NewRecorder()
		
		// Create context and call handler
		c, _ := gin.CreateTestContext(w)
		c.Request = req
		
		handler.ListForms(c)
		
		// Assertions
		assert.Equal(t, http.StatusOK, w.Code)
		mockService.AssertExpectations(t)
	})
	
	t.Run("missing app id", func(t *testing.T) {
		// Setup
		handler := &FormHandler{}
		
		// Create request without app_id
		req, _ := http.NewRequest(http.MethodGet, "/forms", nil)
		w := httptest.NewRecorder()
		
		// Create context and call handler
		c, _ := gin.CreateTestContext(w)
		c.Request = req
		
		handler.ListForms(c)
		
		// Assertions
		assert.Equal(t, http.StatusBadRequest, w.Code)
	})
	
	t.Run("service error", func(t *testing.T) {
		// Setup
		mockService := new(MockFormService)
		handler := &FormHandler{formService: mockService}
		
		// Mock service to return error
		mockService.On("ListForms", mock.Anything, "app_123", 1, 10).Return([]*domain.FormData(nil), int64(0), testutils.NewError("service error"))
		
		// Create request
		req, _ := http.NewRequest(http.MethodGet, "/forms?app_id=app_123", nil)
		w := httptest.NewRecorder()
		
		// Create context and call handler
		c, _ := gin.CreateTestContext(w)
		c.Request = req
		
		handler.ListForms(c)
		
		// Assertions
		assert.Equal(t, http.StatusInternalServerError, w.Code)
		mockService.AssertExpectations(t)
	})
}

func TestFormHandler_GetForm(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("successful retrieval", func(t *testing.T) {
		// Setup
		mockService := new(MockFormService)
		handler := &FormHandler{formService: mockService}
		
		// Mock data
		form := &domain.FormData{
			ID:          "form_123",
			AppID:       "app_123",
			Name:        "Test Form",
			Description: "Test Description",
			Schema:      "{}",
			IsActive:    true,
			CreatedBy:   "user_123",
			CreatedAt:   time.Now(),
			UpdatedAt:   time.Now(),
		}
		
		// Mock service
		mockService.On("GetForm", mock.Anything, "form_123").Return(form, nil)
		
		// Create request
		req, _ := http.NewRequest(http.MethodGet, "/forms/form_123", nil)
		w := httptest.NewRecorder()
		
		// Create context and call handler
		c, _ := gin.CreateTestContext(w)
		c.Request = req
		c.AddParam("id", "form_123")
		
		handler.GetForm(c)
		
		// Assertions
		assert.Equal(t, http.StatusOK, w.Code)
		mockService.AssertExpectations(t)
	})
	
	t.Run("missing form id", func(t *testing.T) {
		// Setup
		handler := &FormHandler{}
		
		// Create request without form ID
		req, _ := http.NewRequest(http.MethodGet, "/forms/", nil)
		w := httptest.NewRecorder()
		
		// Create context and call handler
		c, _ := gin.CreateTestContext(w)
		c.Request = req
		
		handler.GetForm(c)
		
		// Assertions
		assert.Equal(t, http.StatusBadRequest, w.Code)
	})
	
	t.Run("form not found", func(t *testing.T) {
		// Setup
		mockService := new(MockFormService)
		handler := &FormHandler{formService: mockService}
		
		// Mock service to return "form not found" error
		mockService.On("GetForm", mock.Anything, "form_123").Return((*domain.FormData)(nil), testutils.NewError("form not found"))
		
		// Create request
		req, _ := http.NewRequest(http.MethodGet, "/forms/form_123", nil)
		w := httptest.NewRecorder()
		
		// Create context and call handler
		c, _ := gin.CreateTestContext(w)
		c.Request = req
		c.AddParam("id", "form_123")
		
		handler.GetForm(c)
		
		// Assertions
		assert.Equal(t, http.StatusNotFound, w.Code)
		mockService.AssertExpectations(t)
	})
	
	t.Run("service error", func(t *testing.T) {
		// Setup
		mockService := new(MockFormService)
		handler := &FormHandler{formService: mockService}
		
		// Mock service to return error
		mockService.On("GetForm", mock.Anything, "form_123").Return((*domain.FormData)(nil), testutils.NewError("service error"))
		
		// Create request
		req, _ := http.NewRequest(http.MethodGet, "/forms/form_123", nil)
		w := httptest.NewRecorder()
		
		// Create context and call handler
		c, _ := gin.CreateTestContext(w)
		c.Request = req
		c.AddParam("id", "form_123")
		
		handler.GetForm(c)
		
		// Assertions
		assert.Equal(t, http.StatusInternalServerError, w.Code)
		mockService.AssertExpectations(t)
	})
}

func TestFormHandler_SubmitFormData(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("successful submission", func(t *testing.T) {
		// Setup
		mockService := new(MockFormService)
		handler := &FormHandler{formService: mockService}
		
		// Mock data
		entry := &domain.FormDataEntry{
			ID:        "entry_123",
			FormID:    "form_123",
			Data:      `{"name": "John", "age": 30}`,
			CreatedBy: "user_123",
			CreatedAt: time.Now(),
		}
		
		// Mock service
		mockService.On("SubmitFormData", mock.Anything, mock.MatchedBy(func(req *service.SubmitFormDataRequest) bool {
			return req.FormID == "form_123" && req.Data == `{"name": "John", "age": 30}`
		})).Return(entry, nil)
		
		// Create request
		reqBody := `{"form_id":"form_123","data":"{\"name\": \"John\", \"age\": 30}","created_by":"user_123"}`
		req, _ := http.NewRequest(http.MethodPost, "/forms/data", bytes.NewBufferString(reqBody))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		
		// Create context and call handler
		c, _ := gin.CreateTestContext(w)
		c.Request = req
		
		handler.SubmitFormData(c)
		
		// Assertions
		assert.Equal(t, http.StatusOK, w.Code)
		mockService.AssertExpectations(t)
	})
	
	t.Run("invalid request body", func(t *testing.T) {
		// Setup
		handler := &FormHandler{}
		
		// Create request with invalid JSON
		reqBody := `{"form_id":"form_123","data":"{\"name\": \"John\", \"age\": 30}"` // Missing closing brace and missing created_by
		req, _ := http.NewRequest(http.MethodPost, "/forms/data", bytes.NewBufferString(reqBody))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		
		// Create context and call handler
		c, _ := gin.CreateTestContext(w)
		c.Request = req
		
		handler.SubmitFormData(c)
		
		// Assertions
		assert.Equal(t, http.StatusBadRequest, w.Code)
	})
	
	t.Run("form not found or inactive", func(t *testing.T) {
		// Setup
		mockService := new(MockFormService)
		handler := &FormHandler{formService: mockService}
		
		// Mock service to return "form not found or inactive" error
		mockService.On("SubmitFormData", mock.Anything, mock.MatchedBy(func(req *service.SubmitFormDataRequest) bool {
			return req.FormID == "form_123"
		})).Return((*domain.FormDataEntry)(nil), testutils.NewError("form not found or inactive"))
		
		// Create request
		reqBody := `{"form_id":"form_123","data":"{\"name\": \"John\", \"age\": 30}","created_by":"user_123"}`
		req, _ := http.NewRequest(http.MethodPost, "/forms/data", bytes.NewBufferString(reqBody))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		
		// Create context and call handler
		c, _ := gin.CreateTestContext(w)
		c.Request = req
		
		handler.SubmitFormData(c)
		
		// Assertions
		assert.Equal(t, http.StatusNotFound, w.Code)
		mockService.AssertExpectations(t)
	})
	
	t.Run("service error", func(t *testing.T) {
		// Setup
		mockService := new(MockFormService)
		handler := &FormHandler{formService: mockService}
		
		// Mock service to return error
		mockService.On("SubmitFormData", mock.Anything, mock.MatchedBy(func(req *service.SubmitFormDataRequest) bool {
			return req.FormID == "form_123"
		})).Return((*domain.FormDataEntry)(nil), testutils.NewError("service error"))
		
		// Create request
		reqBody := `{"form_id":"form_123","data":"{\"name\": \"John\", \"age\": 30}","created_by":"user_123"}`
		req, _ := http.NewRequest(http.MethodPost, "/forms/data", bytes.NewBufferString(reqBody))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		
		// Create context and call handler
		c, _ := gin.CreateTestContext(w)
		c.Request = req
		
		handler.SubmitFormData(c)
		
		// Assertions
		assert.Equal(t, http.StatusInternalServerError, w.Code)
		mockService.AssertExpectations(t)
	})
}

func TestFormHandler_ListFormDataEntries(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("successful listing", func(t *testing.T) {
		// Setup
		mockService := new(MockFormService)
		handler := &FormHandler{formService: mockService}
		
		// Mock data
		entries := []*domain.FormDataEntry{
			{
				ID:        "entry_1",
				FormID:    "form_123",
				Data:      `{"name": "John", "age": 30}`,
				CreatedBy: "user_123",
				CreatedAt: time.Now(),
			},
			{
				ID:        "entry_2",
				FormID:    "form_123",
				Data:      `{"name": "Jane", "age": 25}`,
				CreatedBy: "user_456",
				CreatedAt: time.Now(),
			},
		}
		
		// Mock service
		mockService.On("ListFormDataEntries", mock.Anything, "form_123", 1, 10).Return(entries, int64(2), nil)
		
		// Create request
		req, _ := http.NewRequest(http.MethodGet, "/forms/data?form_id=form_123", nil)
		w := httptest.NewRecorder()
		
		// Create context and call handler
		c, _ := gin.CreateTestContext(w)
		c.Request = req
		
		handler.ListFormDataEntries(c)
		
		// Assertions
		assert.Equal(t, http.StatusOK, w.Code)
		mockService.AssertExpectations(t)
	})
	
	t.Run("missing form id", func(t *testing.T) {
		// Setup
		handler := &FormHandler{}
		
		// Create request without form_id
		req, _ := http.NewRequest(http.MethodGet, "/forms/data", nil)
		w := httptest.NewRecorder()
		
		// Create context and call handler
		c, _ := gin.CreateTestContext(w)
		c.Request = req
		
		handler.ListFormDataEntries(c)
		
		// Assertions
		assert.Equal(t, http.StatusBadRequest, w.Code)
	})
	
	t.Run("service error", func(t *testing.T) {
		// Setup
		mockService := new(MockFormService)
		handler := &FormHandler{formService: mockService}
		
		// Mock service to return error
		mockService.On("ListFormDataEntries", mock.Anything, "form_123", 1, 10).Return([]*domain.FormDataEntry(nil), int64(0), testutils.NewError("service error"))
		
		// Create request
		req, _ := http.NewRequest(http.MethodGet, "/forms/data?form_id=form_123", nil)
		w := httptest.NewRecorder()
		
		// Create context and call handler
		c, _ := gin.CreateTestContext(w)
		c.Request = req
		
		handler.ListFormDataEntries(c)
		
		// Assertions
		assert.Equal(t, http.StatusInternalServerError, w.Code)
		mockService.AssertExpectations(t)
	})
}