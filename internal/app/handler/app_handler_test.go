package handler

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"cdk-office/internal/app/domain"
	"cdk-office/internal/app/service"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

// mockAppService is a mock implementation of the AppServiceInterface
type mockAppService struct {
	apps   map[string]*domain.Application
	nextID int
}

func newMockAppService() *mockAppService {
	return &mockAppService{
		apps:   make(map[string]*domain.Application),
		nextID: 1,
	}
}

func (m *mockAppService) CreateApplication(ctx context.Context, req *service.CreateApplicationRequest) (*domain.Application, error) {
	// Check if application with same name already exists in team
	for _, app := range m.apps {
		if app.Name == req.Name && app.TeamID == req.TeamID {
			return nil, errors.New("application with this name already exists in the team")
		}
	}

	// Check if application type is valid
	validTypes := map[string]bool{
		"qrcode": true,
		"form":   true,
		"survey": true,
	}
	if !validTypes[req.Type] {
		return nil, errors.New("invalid application type")
	}

	// Create new application
	app := &domain.Application{
		ID:          "app_" + string(rune(m.nextID+'0')),
		TeamID:      req.TeamID,
		Name:        req.Name,
		Description: req.Description,
		Type:        req.Type,
		Config:      req.Config,
		CreatedBy:   req.CreatedBy,
		IsActive:    true,
	}

	m.apps[app.ID] = app
	m.nextID++
	return app, nil
}

func (m *mockAppService) UpdateApplication(ctx context.Context, id string, req *service.UpdateApplicationRequest) error {
	app, exists := m.apps[id]
	if !exists {
		return errors.New("application not found")
	}

	if req.Name != "" {
		app.Name = req.Name
	}
	if req.Description != "" {
		app.Description = req.Description
	}
	if req.Config != "" {
		app.Config = req.Config
	}
	if req.IsActive != nil {
		app.IsActive = *req.IsActive
	}

	return nil
}

func (m *mockAppService) DeleteApplication(ctx context.Context, id string) error {
	_, exists := m.apps[id]
	if !exists {
		return errors.New("application not found")
	}

	delete(m.apps, id)
	return nil
}

func (m *mockAppService) ListApplications(ctx context.Context, teamID string, page, size int) ([]*domain.Application, int64, error) {
	var apps []*domain.Application
	for _, app := range m.apps {
		if app.TeamID == teamID {
			apps = append(apps, app)
		}
	}

	// Simple pagination
	start := (page - 1) * size
	end := start + size
	if start >= len(apps) {
		return []*domain.Application{}, int64(len(apps)), nil
	}
	if end > len(apps) {
		end = len(apps)
	}

	return apps[start:end], int64(len(apps)), nil
}

func (m *mockAppService) GetApplication(ctx context.Context, id string) (*domain.Application, error) {
	app, exists := m.apps[id]
	if !exists {
		return nil, errors.New("application not found")
	}
	return app, nil
}

// TestAppHandler tests the AppHandler
func TestAppHandler(t *testing.T) {
	// Set up test environment
	gin.SetMode(gin.TestMode)

	// Create mock service
	mockService := newMockAppService()

	// Create app handler with mock service
	appHandler := NewAppHandlerWithService(mockService)

	// Test CreateApplication
	t.Run("CreateApplication", func(t *testing.T) {
		// Create test request
		reqBody := CreateApplicationRequest{
			TeamID:    "team_123",
			Name:      "Test App",
			Type:      "qrcode",
			CreatedBy: "user_123",
		}
		jsonReq, _ := json.Marshal(reqBody)

		// Create HTTP request and response recorder
		req, _ := http.NewRequest("POST", "/applications", bytes.NewBuffer(jsonReq))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		// Create gin context and call handler
		c, _ := gin.CreateTestContext(w)
		c.Request = req
		appHandler.CreateApplication(c)

		// Assert response
		assert.Equal(t, http.StatusOK, w.Code)
	})

	// Test CreateApplication with invalid JSON
	t.Run("CreateApplicationInvalidJSON", func(t *testing.T) {
		// Create HTTP request with invalid JSON
		req, _ := http.NewRequest("POST", "/applications", bytes.NewBuffer([]byte("{invalid json}")))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		// Create gin context and call handler
		c, _ := gin.CreateTestContext(w)
		c.Request = req
		appHandler.CreateApplication(c)

		// Assert response
		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	// Test CreateApplication with invalid type
	t.Run("CreateApplicationInvalidType", func(t *testing.T) {
		// Create test request with invalid type
		reqBody := CreateApplicationRequest{
			TeamID:    "team_123",
			Name:      "Invalid App",
			Type:      "invalid",
			CreatedBy: "user_123",
		}
		jsonReq, _ := json.Marshal(reqBody)

		// Create HTTP request and response recorder
		req, _ := http.NewRequest("POST", "/applications", bytes.NewBuffer(jsonReq))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		// Create gin context and call handler
		c, _ := gin.CreateTestContext(w)
		c.Request = req
		appHandler.CreateApplication(c)

		// Assert response
		assert.Equal(t, http.StatusInternalServerError, w.Code)
	})

	// Test UpdateApplication
	t.Run("UpdateApplication", func(t *testing.T) {
		// First create an application
		createReq := &service.CreateApplicationRequest{
			TeamID:    "team_123",
			Name:      "Update Test App",
			Type:      "qrcode",
			CreatedBy: "user_123",
		}
		app, _ := mockService.CreateApplication(context.Background(), createReq)

		// Create update request
		updateReqBody := UpdateApplicationRequest{
			Name: "Updated App",
		}
		jsonReq, _ := json.Marshal(updateReqBody)

		// Create HTTP request and response recorder
		req, _ := http.NewRequest("PUT", "/applications/"+app.ID, bytes.NewBuffer(jsonReq))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		// Create gin context and call handler
		c, _ := gin.CreateTestContext(w)
		c.Request = req
		c.AddParam("id", app.ID)
		appHandler.UpdateApplication(c)

		// Assert response
		assert.Equal(t, http.StatusOK, w.Code)
	})

	// Test UpdateApplication with non-existent ID
	t.Run("UpdateApplicationNotFound", func(t *testing.T) {
		// Create update request
		updateReqBody := UpdateApplicationRequest{
			Name: "Updated App",
		}
		jsonReq, _ := json.Marshal(updateReqBody)

		// Create HTTP request and response recorder
		req, _ := http.NewRequest("PUT", "/applications/non-existent", bytes.NewBuffer(jsonReq))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		// Create gin context and call handler
		c, _ := gin.CreateTestContext(w)
		c.Request = req
		c.AddParam("id", "non-existent")
		appHandler.UpdateApplication(c)

		// Assert response
		assert.Equal(t, http.StatusNotFound, w.Code)
	})

	// Test UpdateApplication with invalid JSON
	t.Run("UpdateApplicationInvalidJSON", func(t *testing.T) {
		// Create HTTP request with invalid JSON
		req, _ := http.NewRequest("PUT", "/applications/app_1", bytes.NewBuffer([]byte("{invalid json}")))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		// Create gin context and call handler
		c, _ := gin.CreateTestContext(w)
		c.Request = req
		c.AddParam("id", "app_1")
		appHandler.UpdateApplication(c)

		// Assert response
		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	// Test DeleteApplication
	t.Run("DeleteApplication", func(t *testing.T) {
		// First create an application
		createReq := &service.CreateApplicationRequest{
			TeamID:    "team_123",
			Name:      "Delete Test App",
			Type:      "qrcode",
			CreatedBy: "user_123",
		}
		app, _ := mockService.CreateApplication(context.Background(), createReq)

		// Create HTTP request and response recorder
		req, _ := http.NewRequest("DELETE", "/applications/"+app.ID, nil)
		w := httptest.NewRecorder()

		// Create gin context and call handler
		c, _ := gin.CreateTestContext(w)
		c.Request = req
		c.AddParam("id", app.ID)
		appHandler.DeleteApplication(c)

		// Assert response
		assert.Equal(t, http.StatusOK, w.Code)
	})

	// Test DeleteApplication with non-existent ID
	t.Run("DeleteApplicationNotFound", func(t *testing.T) {
		// Create HTTP request and response recorder
		req, _ := http.NewRequest("DELETE", "/applications/non-existent", nil)
		w := httptest.NewRecorder()

		// Create gin context and call handler
		c, _ := gin.CreateTestContext(w)
		c.Request = req
		c.AddParam("id", "non-existent")
		appHandler.DeleteApplication(c)

		// Assert response
		assert.Equal(t, http.StatusNotFound, w.Code)
	})

	// Test ListApplications
	t.Run("ListApplications", func(t *testing.T) {
		// Create a few applications
		for i := 1; i <= 3; i++ {
			createReq := &service.CreateApplicationRequest{
				TeamID:    "team_list",
				Name:      "List Test App " + string(rune(i+'0')),
				Type:      "qrcode",
				CreatedBy: "user_123",
			}
			_, _ = mockService.CreateApplication(context.Background(), createReq)
		}

		// Create HTTP request and response recorder
		req, _ := http.NewRequest("GET", "/applications?team_id=team_list", nil)
		w := httptest.NewRecorder()

		// Create gin context and call handler
		c, _ := gin.CreateTestContext(w)
		c.Request = req
		appHandler.ListApplications(c)

		// Assert response
		assert.Equal(t, http.StatusOK, w.Code)
	})

	// Test ListApplications with missing team_id
	t.Run("ListApplicationsMissingTeamID", func(t *testing.T) {
		// Create HTTP request and response recorder
		req, _ := http.NewRequest("GET", "/applications", nil)
		w := httptest.NewRecorder()

		// Create gin context and call handler
		c, _ := gin.CreateTestContext(w)
		c.Request = req
		appHandler.ListApplications(c)

		// Assert response
		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	// Test GetApplication
	t.Run("GetApplication", func(t *testing.T) {
		// First create an application
		createReq := &service.CreateApplicationRequest{
			TeamID:    "team_123",
			Name:      "Get Test App",
			Type:      "qrcode",
			CreatedBy: "user_123",
		}
		app, _ := mockService.CreateApplication(context.Background(), createReq)

		// Create HTTP request and response recorder
		req, _ := http.NewRequest("GET", "/applications/"+app.ID, nil)
		w := httptest.NewRecorder()

		// Create gin context and call handler
		c, _ := gin.CreateTestContext(w)
		c.Request = req
		c.AddParam("id", app.ID)
		appHandler.GetApplication(c)

		// Assert response
		assert.Equal(t, http.StatusOK, w.Code)
	})

	// Test GetApplication with non-existent ID
	t.Run("GetApplicationNotFound", func(t *testing.T) {
		// Create HTTP request and response recorder
		req, _ := http.NewRequest("GET", "/applications/non-existent", nil)
		w := httptest.NewRecorder()

		// Create gin context and call handler
		c, _ := gin.CreateTestContext(w)
		c.Request = req
		c.AddParam("id", "non-existent")
		appHandler.GetApplication(c)

		// Assert response
		assert.Equal(t, http.StatusNotFound, w.Code)
	})

	// Test parseInt function
	t.Run("ParseInt", func(t *testing.T) {
		// Test valid integer
		result, err := parseInt("123")
		assert.NoError(t, err)
		assert.Equal(t, 123, result)

		// Test invalid integer
		_, err = parseInt("abc")
		assert.Error(t, err)
	})
}