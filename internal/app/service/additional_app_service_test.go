package service

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"cdk-office/internal/shared/testutils"
)

// TestAppServiceAdditional tests additional scenarios for the AppService
func TestAppServiceAdditional(t *testing.T) {
	// Set up test environment
	testDB := testutils.SetupTestDB()

	// Create app service with database connection
	appService := NewAppServiceWithDB(testDB)

	// Test CreateApplication with invalid type
	t.Run("CreateApplicationInvalidType", func(t *testing.T) {
		ctx := context.Background()
		req := &CreateApplicationRequest{
			TeamID:    "team_123",
			Name:      "Invalid App",
			Type:      "invalid",
			CreatedBy: "user_123",
		}

		app, err := appService.CreateApplication(ctx, req)

		assert.Error(t, err)
		assert.Nil(t, app)
		assert.Equal(t, "invalid application type", err.Error())
	})

	// Test CreateApplication with duplicate name
	t.Run("CreateApplicationDuplicateName", func(t *testing.T) {
		ctx := context.Background()

		// Create first application
		req1 := &CreateApplicationRequest{
			TeamID:    "team_123",
			Name:      "Duplicate Test App",
			Type:      "qrcode",
			CreatedBy: "user_123",
		}

		app1, err1 := appService.CreateApplication(ctx, req1)
		assert.NoError(t, err1)
		assert.NotNil(t, app1)

		// Try to create another application with the same name
		req2 := &CreateApplicationRequest{
			TeamID:    "team_123",
			Name:      "Duplicate Test App",
			Type:      "form",
			CreatedBy: "user_123",
		}

		app2, err2 := appService.CreateApplication(ctx, req2)

		assert.Error(t, err2)
		assert.Nil(t, app2)
		assert.Equal(t, "application with this name already exists in the team", err2.Error())
	})

	// Test UpdateApplication with non-existent ID
	t.Run("UpdateApplicationNotFound", func(t *testing.T) {
		ctx := context.Background()
		req := &UpdateApplicationRequest{
			Name: "Updated App",
		}

		err := appService.UpdateApplication(ctx, "non-existent-id", req)

		assert.Error(t, err)
		assert.Equal(t, "application not found", err.Error())
	})

	// Test DeleteApplication with non-existent ID
	t.Run("DeleteApplicationNotFound", func(t *testing.T) {
		ctx := context.Background()

		err := appService.DeleteApplication(ctx, "non-existent-id")

		assert.Error(t, err)
		assert.Equal(t, "application not found", err.Error())
	})

	// Test GetApplication
	t.Run("GetApplication", func(t *testing.T) {
		ctx := context.Background()

		// First create an application
		createReq := &CreateApplicationRequest{
			TeamID:    "team_123",
			Name:      "Get Test App",
			Type:      "qrcode",
			CreatedBy: "user_123",
		}

		createdApp, err := appService.CreateApplication(ctx, createReq)
		assert.NoError(t, err)
		assert.NotNil(t, createdApp)

		// Now get the application
		retrievedApp, err := appService.GetApplication(ctx, createdApp.ID)
		assert.NoError(t, err)
		assert.NotNil(t, retrievedApp)
		assert.Equal(t, createdApp.ID, retrievedApp.ID)
		assert.Equal(t, createdApp.Name, retrievedApp.Name)
	})

	// Test GetApplication with non-existent ID
	t.Run("GetApplicationNotFound", func(t *testing.T) {
		ctx := context.Background()

		app, err := appService.GetApplication(ctx, "non-existent-id")

		assert.Error(t, err)
		assert.Nil(t, app)
		assert.Equal(t, "application not found", err.Error())
	})

	// Test ListApplications
	t.Run("ListApplications", func(t *testing.T) {
		ctx := context.Background()

		// Create a few applications
		for i := 1; i <= 3; i++ {
			req := &CreateApplicationRequest{
				TeamID:    "team_list",
				Name:      "List Test App " + string(rune(i+'0')),
				Type:      "qrcode",
				CreatedBy: "user_123",
			}

			_, err := appService.CreateApplication(ctx, req)
			assert.NoError(t, err)
		}

		// List applications
		apps, total, err := appService.ListApplications(ctx, "team_list", 1, 10)
		assert.NoError(t, err)
		assert.NotNil(t, apps)
		assert.GreaterOrEqual(t, total, int64(3))
		assert.GreaterOrEqual(t, len(apps), 3)
	})

	// Test ListApplications with invalid pagination
	t.Run("ListApplicationsInvalidPagination", func(t *testing.T) {
		ctx := context.Background()

		// Test with page = 0
		apps, _, err := appService.ListApplications(ctx, "team_list", 0, 10)
		assert.NoError(t, err)
		assert.NotNil(t, apps)
		// Just check it doesn't panic

		// Test with size = 0
		apps, _, err = appService.ListApplications(ctx, "team_list", 1, 0)
		assert.NoError(t, err)
		assert.NotNil(t, apps)
		// Default size should be 10

		// Test with size > 100
		apps, _, err = appService.ListApplications(ctx, "team_list", 1, 150)
		assert.NoError(t, err)
		assert.NotNil(t, apps)
		// Default size should be 10
	})

	// Test UpdateApplication with all fields
	t.Run("UpdateApplicationAllFields", func(t *testing.T) {
		ctx := context.Background()

		// Create an application
		createReq := &CreateApplicationRequest{
			TeamID:    "team_123",
			Name:      "Update All Fields App",
			Type:      "qrcode",
			CreatedBy: "user_123",
		}

		app, err := appService.CreateApplication(ctx, createReq)
		assert.NoError(t, err)
		assert.NotNil(t, app)

		// Update all fields
		isActive := false
		updateReq := &UpdateApplicationRequest{
			Name:        "Fully Updated App",
			Description: "Updated description",
			Config:      "{\"updated\": true}",
			IsActive:    &isActive,
		}

		err = appService.UpdateApplication(ctx, app.ID, updateReq)
		assert.NoError(t, err)

		// Verify the update
		updatedApp, err := appService.GetApplication(ctx, app.ID)
		assert.NoError(t, err)
		assert.NotNil(t, updatedApp)
		assert.Equal(t, "Fully Updated App", updatedApp.Name)
		assert.Equal(t, "Updated description", updatedApp.Description)
		assert.Equal(t, "{\"updated\": true}", updatedApp.Config)
		assert.Equal(t, false, updatedApp.IsActive)
	})
}