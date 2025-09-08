package service

import (
	"context"
	"testing"
	
	"github.com/stretchr/testify/assert"
	"cdk-office/internal/shared/testutils"
)

// TestAppService tests the AppService
func TestAppService(t *testing.T) {
	// Set up test environment
	// logger.InitTestLogger()
	
	// Initialize the database connection for testing
	testDB := testutils.SetupTestDB()
	
	// Create app service with database connection
	appService := NewAppServiceWithDB(testDB)
	
	// Test CreateApplication
	t.Run("CreateApplication", func(t *testing.T) {
		// Prepare test data
		ctx := context.Background()
		req := &CreateApplicationRequest{
			TeamID:      "team_123",
			Name:        "Test App",
			Type:        "qrcode",
			CreatedBy:   "user_123",
			Description: "Test application description",
			Config:      "{}",
		}
		
		// Call the method under test
		app, err := appService.CreateApplication(ctx, req)
		
		// Assert results
		assert.NoError(t, err)
		assert.NotNil(t, app)
	})
	
	// Test UpdateApplication
	t.Run("UpdateApplication", func(t *testing.T) {
		// First create an application to update
		ctx := context.Background()
		createReq := &CreateApplicationRequest{
			TeamID:      "team_123",
			Name:        "Test App 2",
			Type:        "qrcode",
			CreatedBy:   "user_123",
			Description: "Test application description",
			Config:      "{}",
		}
		
		app, err := appService.CreateApplication(ctx, createReq)
		assert.NoError(t, err)
		assert.NotNil(t, app)
		
		// Now update the application
		updateReq := &UpdateApplicationRequest{
			Name: "Updated App",
		}
		
		err = appService.UpdateApplication(ctx, app.ID, updateReq)
		assert.NoError(t, err)
	})
	
	// Test DeleteApplication
	t.Run("DeleteApplication", func(t *testing.T) {
		// First create an application to delete
		ctx := context.Background()
		createReq := &CreateApplicationRequest{
			TeamID:      "team_123",
			Name:        "Test App 3",
			Type:        "qrcode",
			CreatedBy:   "user_123",
			Description: "Test application description",
			Config:      "{}",
		}
		
		app, err := appService.CreateApplication(ctx, createReq)
		assert.NoError(t, err)
		assert.NotNil(t, app)
		
		// Now delete the application
		err = appService.DeleteApplication(ctx, app.ID)
		assert.NoError(t, err)
	})
}