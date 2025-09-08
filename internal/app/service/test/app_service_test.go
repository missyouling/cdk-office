package test

import (
	"context"
	"testing"
	"time"

	"cdk-office/internal/app/service"
	"cdk-office/internal/shared/testutils"
	"github.com/stretchr/testify/assert"
)

func TestAppService_CreateApplication(t *testing.T) {
	// Setup
	db := testutils.SetupTestDB()
	appService := service.NewAppServiceWithDB(db)
	ctx := context.Background()

	// Test cases
	tests := []struct {
		name          string
		request       *service.CreateApplicationRequest
		expectError   bool
		errorMessage  string
	}{
		{
			name: "Valid application creation",
			request: &service.CreateApplicationRequest{
				TeamID:      "team1",
				Name:        "Test App",
				Description: "Test application description",
				Type:        "qrcode",
				Config:      "{}",
				CreatedBy:   "user1",
			},
			expectError: false,
		},
		{
			name: "Invalid application type",
			request: &service.CreateApplicationRequest{
				TeamID:      "team1",
				Name:        "Invalid App",
				Description: "Invalid application description",
				Type:        "invalid",
				Config:      "{}",
				CreatedBy:   "user1",
			},
			expectError:  true,
			errorMessage: "invalid application type",
		},
		{
			name: "Duplicate application name in same team",
			request: &service.CreateApplicationRequest{
				TeamID:      "team1",
				Name:        "Duplicate App",
				Description: "First application",
				Type:        "form",
				Config:      "{}",
				CreatedBy:   "user1",
			},
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// For the duplicate test case, create the first app
			if tt.name == "Duplicate application name in same team" {
				// Create first application
				_, err := appService.CreateApplication(ctx, tt.request)
				assert.NoError(t, err)

				// Try to create duplicate
				_, err = appService.CreateApplication(ctx, tt.request)
				assert.Error(t, err)
				assert.Equal(t, "application with this name already exists in the team", err.Error())
				return
			}

			// Execute
			app, err := appService.CreateApplication(ctx, tt.request)

			// Assert
			if tt.expectError {
				assert.Error(t, err)
				assert.Equal(t, tt.errorMessage, err.Error())
				assert.Nil(t, app)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, app)
				assert.NotEmpty(t, app.ID)
				assert.Equal(t, tt.request.TeamID, app.TeamID)
				assert.Equal(t, tt.request.Name, app.Name)
				assert.Equal(t, tt.request.Description, app.Description)
				assert.Equal(t, tt.request.Type, app.Type)
				assert.Equal(t, tt.request.Config, app.Config)
				assert.Equal(t, tt.request.CreatedBy, app.CreatedBy)
				assert.True(t, app.IsActive)
				assert.WithinDuration(t, time.Now(), app.CreatedAt, time.Second)
				assert.WithinDuration(t, time.Now(), app.UpdatedAt, time.Second)
			}
		})
	}
}

func TestAppService_UpdateApplication(t *testing.T) {
	// Setup
	db := testutils.SetupTestDB()
	appService := service.NewAppServiceWithDB(db)
	ctx := context.Background()

	// Create an application for testing
	createReq := &service.CreateApplicationRequest{
		TeamID:      "team1",
		Name:        "Test App",
		Description: "Test application description",
		Type:        "qrcode",
		Config:      "{}",
		CreatedBy:   "user1",
	}
	createdApp, err := appService.CreateApplication(ctx, createReq)
	assert.NoError(t, err)
	assert.NotNil(t, createdApp)

	// Test cases
	tests := []struct {
		name          string
		appID         string
		request       *service.UpdateApplicationRequest
		expectError   bool
		errorMessage  string
	}{
		{
			name:  "Valid application update",
			appID: createdApp.ID,
			request: &service.UpdateApplicationRequest{
				Name:        "Updated App Name",
				Description: "Updated description",
				Config:      "{\"updated\": true}",
				IsActive:    boolPtr(true),
			},
			expectError: false,
		},
		{
			name:  "Update non-existent application",
			appID: "non-existent-id",
			request: &service.UpdateApplicationRequest{
				Name: "Updated App Name",
			},
			expectError:  true,
			errorMessage: "application not found",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Execute
			err := appService.UpdateApplication(ctx, tt.appID, tt.request)

			// Assert
			if tt.expectError {
				assert.Error(t, err)
				assert.Equal(t, tt.errorMessage, err.Error())
			} else {
				assert.NoError(t, err)

				// Verify the update
				updatedApp, getErr := appService.GetApplication(ctx, tt.appID)
				assert.NoError(t, getErr)
				assert.NotNil(t, updatedApp)
				assert.Equal(t, tt.request.Name, updatedApp.Name)
				assert.Equal(t, tt.request.Description, updatedApp.Description)
				assert.Equal(t, tt.request.Config, updatedApp.Config)
				if tt.request.IsActive != nil {
					assert.Equal(t, *tt.request.IsActive, updatedApp.IsActive)
				}
			}
		})
	}
}

func TestAppService_DeleteApplication(t *testing.T) {
	// Setup
	db := testutils.SetupTestDB()
	appService := service.NewAppServiceWithDB(db)
	ctx := context.Background()

	// Create an application for testing
	createReq := &service.CreateApplicationRequest{
		TeamID:      "team1",
		Name:        "Test App",
		Description: "Test application description",
		Type:        "qrcode",
		Config:      "{}",
		CreatedBy:   "user1",
	}
	createdApp, err := appService.CreateApplication(ctx, createReq)
	assert.NoError(t, err)
	assert.NotNil(t, createdApp)

	// Test cases
	tests := []struct {
		name         string
		appID        string
		expectError  bool
		errorMessage string
	}{
		{
			name:        "Delete non-existent application",
			appID:       "non-existent-id",
			expectError: true,
		},
		{
			name:        "Valid application deletion",
			appID:       createdApp.ID,
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Execute
			err := appService.DeleteApplication(ctx, tt.appID)

			// Assert
			if tt.expectError {
				assert.Error(t, err)
				if tt.errorMessage != "" {
					assert.Equal(t, tt.errorMessage, err.Error())
				}
			} else {
				assert.NoError(t, err)

				// Verify the deletion
				_, getErr := appService.GetApplication(ctx, tt.appID)
				assert.Error(t, getErr)
				assert.Equal(t, "application not found", getErr.Error())
			}
		})
	}
}

func TestAppService_ListApplications(t *testing.T) {
	// Setup
	db := testutils.SetupTestDB()
	appService := service.NewAppServiceWithDB(db)
	ctx := context.Background()

	// Create test applications
	apps := []*service.CreateApplicationRequest{
		{
			TeamID:      "team1",
			Name:        "App 1",
			Description: "First application",
			Type:        "qrcode",
			Config:      "{}",
			CreatedBy:   "user1",
		},
		{
			TeamID:      "team1",
			Name:        "App 2",
			Description: "Second application",
			Type:        "form",
			Config:      "{}",
			CreatedBy:   "user1",
		},
		{
			TeamID:      "team2",
			Name:        "App 3",
			Description: "Third application",
			Type:        "survey",
			Config:      "{}",
			CreatedBy:   "user2",
		},
	}

	for _, appReq := range apps {
		_, err := appService.CreateApplication(ctx, appReq)
		assert.NoError(t, err)
	}

	// Test cases
	tests := []struct {
		name              string
		teamID            string
		page              int
		size              int
		expectedCount     int
		expectedListCount int
	}{
		{
			name:              "List applications for team1",
			teamID:            "team1",
			page:              1,
			size:              10,
			expectedCount:     2,
			expectedListCount: 2,
		},
		{
			name:              "List applications for team2",
			teamID:            "team2",
			page:              1,
			size:              10,
			expectedCount:     1,
			expectedListCount: 1,
		},
		{
			name:              "List applications with pagination",
			teamID:            "team1",
			page:              1,
			size:              1,
			expectedCount:     2, // Total count is still 2
			expectedListCount: 1, // But we only get 1 item due to pagination
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Execute
			applications, total, err := appService.ListApplications(ctx, tt.teamID, tt.page, tt.size)

			// Assert
			assert.NoError(t, err)
			assert.NotNil(t, applications)
			assert.Equal(t, int64(tt.expectedCount), total)
			assert.Len(t, applications, tt.expectedListCount)
		})
	}
}

func TestAppService_GetApplication(t *testing.T) {
	// Setup
	db := testutils.SetupTestDB()
	appService := service.NewAppServiceWithDB(db)
	ctx := context.Background()

	// Create an application for testing
	createReq := &service.CreateApplicationRequest{
		TeamID:      "team1",
		Name:        "Test App",
		Description: "Test application description",
		Type:        "qrcode",
		Config:      "{}",
		CreatedBy:   "user1",
	}
	createdApp, err := appService.CreateApplication(ctx, createReq)
	assert.NoError(t, err)
	assert.NotNil(t, createdApp)

	// Test cases
	tests := []struct {
		name         string
		appID        string
		expectError  bool
		errorMessage string
	}{
		{
			name:         "Get non-existent application",
			appID:        "non-existent-id",
			expectError:  true,
			errorMessage: "application not found",
		},
		{
			name:        "Get existing application",
			appID:       createdApp.ID,
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Execute
			app, err := appService.GetApplication(ctx, tt.appID)

			// Assert
			if tt.expectError {
				assert.Error(t, err)
				assert.Equal(t, tt.errorMessage, err.Error())
				assert.Nil(t, app)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, app)
				assert.Equal(t, createdApp.ID, app.ID)
				assert.Equal(t, createdApp.TeamID, app.TeamID)
				assert.Equal(t, createdApp.Name, app.Name)
				assert.Equal(t, createdApp.Description, app.Description)
				assert.Equal(t, createdApp.Type, app.Type)
				assert.Equal(t, createdApp.Config, app.Config)
				assert.Equal(t, createdApp.IsActive, app.IsActive)
				assert.Equal(t, createdApp.CreatedBy, app.CreatedBy)
			}
		})
	}
}

// Helper function to create a pointer to a bool
func boolPtr(b bool) *bool {
	return &b
}