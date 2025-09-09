package service

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"cdk-office/internal/app/domain"
	"cdk-office/internal/shared/testutils"
)

// TestAppPermissionService tests the AppPermissionService
func TestAppPermissionService(t *testing.T) {
	// Set up test environment
	testDB := testutils.SetupTestDB()

	// Create app permission service with database connection
	appPermissionService := &AppPermissionService{
		db: testDB,
	}

	// Test CreateAppPermission
	t.Run("CreateAppPermission", func(t *testing.T) {
		ctx := context.Background()
		req := &CreateAppPermissionRequest{
			AppID:       "app_123",
			Name:        "Test Permission",
			Description: "A test permission",
			Permission:  "read",
			CreatedBy:   "user_123",
		}

		permission, err := appPermissionService.CreateAppPermission(ctx, req)

		assert.NoError(t, err)
		assert.NotNil(t, permission)
		assert.Equal(t, "app_123", permission.AppID)
		assert.Equal(t, "Test Permission", permission.Name)
		assert.Equal(t, "A test permission", permission.Description)
		assert.Equal(t, "read", permission.Permission)
		assert.Equal(t, "user_123", permission.CreatedBy)
	})

	// Test UpdateAppPermission
	t.Run("UpdateAppPermission", func(t *testing.T) {
		ctx := context.Background()

		// First create an app permission
		createReq := &CreateAppPermissionRequest{
			AppID:       "app_123",
			Name:        "Update Test Permission",
			Description: "A test permission to update",
			Permission:  "read",
			CreatedBy:   "user_123",
		}

		permission, err := appPermissionService.CreateAppPermission(ctx, createReq)
		assert.NoError(t, err)
		assert.NotNil(t, permission)

		// Now update the app permission
		updateReq := &UpdateAppPermissionRequest{
			Name:        "Updated Test Permission",
			Description: "An updated test permission",
			Permission:  "write",
		}

		err = appPermissionService.UpdateAppPermission(ctx, permission.ID, updateReq)
		assert.NoError(t, err)
	})

	// Test DeleteAppPermission
	t.Run("DeleteAppPermission", func(t *testing.T) {
		ctx := context.Background()

		// First create an app permission
		createReq := &CreateAppPermissionRequest{
			AppID:       "app_123",
			Name:        "Delete Test Permission",
			Description: "A test permission to delete",
			Permission:  "read",
			CreatedBy:   "user_123",
		}

		permission, err := appPermissionService.CreateAppPermission(ctx, createReq)
		assert.NoError(t, err)
		assert.NotNil(t, permission)

		// Now delete the app permission
		err = appPermissionService.DeleteAppPermission(ctx, permission.ID)
		assert.NoError(t, err)
	})

	// Test ListAppPermissions
	t.Run("ListAppPermissions", func(t *testing.T) {
		ctx := context.Background()

		// Create a few app permissions
		permissionsData := []struct {
			appID       string
			name        string
			description string
			permission  string
		}{
			{"app_201", "Permission 1", "First permission", "read"},
			{"app_201", "Permission 2", "Second permission", "write"},
			{"app_202", "Permission 3", "Third permission", "delete"},
		}

		for _, data := range permissionsData {
			req := &CreateAppPermissionRequest{
				AppID:       data.appID,
				Name:        data.name,
				Description: data.description,
				Permission:  data.permission,
				CreatedBy:   "user_123",
			}

			_, err := appPermissionService.CreateAppPermission(ctx, req)
			assert.NoError(t, err)
		}

		// List app permissions
		permissions, total, err := appPermissionService.ListAppPermissions(ctx, "app_201", 1, 10)
		assert.NoError(t, err)
		assert.NotNil(t, permissions)
		assert.GreaterOrEqual(t, total, int64(2))
		assert.GreaterOrEqual(t, len(permissions), 2)
	})

	// Test GetAppPermission
	t.Run("GetAppPermission", func(t *testing.T) {
		ctx := context.Background()

		// First create an app permission
		createReq := &CreateAppPermissionRequest{
			AppID:       "app_123",
			Name:        "Get Test Permission",
			Description: "A test permission to get",
			Permission:  "read",
			CreatedBy:   "user_123",
		}

		createdPermission, err := appPermissionService.CreateAppPermission(ctx, createReq)
		assert.NoError(t, err)
		assert.NotNil(t, createdPermission)

		// Now get the app permission
		retrievedPermission, err := appPermissionService.GetAppPermission(ctx, createdPermission.ID)
		assert.NoError(t, err)
		assert.NotNil(t, retrievedPermission)
		assert.Equal(t, createdPermission.ID, retrievedPermission.ID)
		assert.Equal(t, createdPermission.Name, retrievedPermission.Name)
	})
}

// TestAppPermissionServiceAdditional tests additional scenarios for the AppPermissionService
func TestAppPermissionServiceAdditional(t *testing.T) {
	// Set up test environment
	testDB := testutils.SetupTestDB()

	// Create app permission service with database connection
	appPermissionService := &AppPermissionService{
		db: testDB,
	}

	// Test CreateAppPermission with invalid permission
	t.Run("CreateAppPermissionInvalidPermission", func(t *testing.T) {
		ctx := context.Background()
		req := &CreateAppPermissionRequest{
			AppID:       "app_123",
			Name:        "Invalid Permission",
			Description: "An invalid permission",
			Permission:  "invalid",
			CreatedBy:   "user_123",
		}

		permission, err := appPermissionService.CreateAppPermission(ctx, req)

		assert.Error(t, err)
		assert.Nil(t, permission)
		assert.Equal(t, "invalid permission", err.Error())
	})

	// Test CreateAppPermission with duplicate name
	t.Run("CreateAppPermissionDuplicateName", func(t *testing.T) {
		ctx := context.Background()

		// Create first app permission
		req1 := &CreateAppPermissionRequest{
			AppID:       "app_123",
			Name:        "Duplicate Test Permission",
			Description: "First permission",
			Permission:  "read",
			CreatedBy:   "user_123",
		}

		permission1, err1 := appPermissionService.CreateAppPermission(ctx, req1)
		assert.NoError(t, err1)
		assert.NotNil(t, permission1)

		// Try to create another app permission with the same name
		req2 := &CreateAppPermissionRequest{
			AppID:       "app_123",
			Name:        "Duplicate Test Permission",
			Description: "Second permission",
			Permission:  "write",
			CreatedBy:   "user_123",
		}

		permission2, err2 := appPermissionService.CreateAppPermission(ctx, req2)

		assert.Error(t, err2)
		assert.Nil(t, permission2)
		assert.Equal(t, "permission with this name already exists in the application", err2.Error())
	})

	// Test UpdateAppPermission with non-existent ID
	t.Run("UpdateAppPermissionNotFound", func(t *testing.T) {
		ctx := context.Background()
		req := &UpdateAppPermissionRequest{
			Name: "Updated Permission",
		}

		err := appPermissionService.UpdateAppPermission(ctx, "non-existent-id", req)

		assert.Error(t, err)
		assert.Equal(t, "application permission not found", err.Error())
	})

	// Test DeleteAppPermission with non-existent ID
	t.Run("DeleteAppPermissionNotFound", func(t *testing.T) {
		ctx := context.Background()

		err := appPermissionService.DeleteAppPermission(ctx, "non-existent-id")

		assert.Error(t, err)
		assert.Equal(t, "application permission not found", err.Error())
	})

	// Test GetAppPermission with non-existent ID
	t.Run("GetAppPermissionNotFound", func(t *testing.T) {
		ctx := context.Background()

		permission, err := appPermissionService.GetAppPermission(ctx, "non-existent-id")

		assert.Error(t, err)
		assert.Nil(t, permission)
		assert.Equal(t, "application permission not found", err.Error())
	})

	// Test ListAppPermissions with invalid pagination
	t.Run("ListAppPermissionsInvalidPagination", func(t *testing.T) {
		ctx := context.Background()

		// Test with page = 0
		permissions, _, err := appPermissionService.ListAppPermissions(ctx, "app_list", 0, 10)
		assert.NoError(t, err)
		assert.NotNil(t, permissions)
		// Just check it doesn't panic

		// Test with size = 0
		permissions, _, err = appPermissionService.ListAppPermissions(ctx, "app_list", 1, 0)
		assert.NoError(t, err)
		assert.NotNil(t, permissions)
		// Default size should be 10

		// Test with size > 100
		permissions, _, err = appPermissionService.ListAppPermissions(ctx, "app_list", 1, 150)
		assert.NoError(t, err)
		assert.NotNil(t, permissions)
		// Default size should be 10
	})

	// Test UpdateAppPermission with invalid permission
	t.Run("UpdateAppPermissionInvalidPermission", func(t *testing.T) {
		ctx := context.Background()

		// Create an app permission
		createReq := &CreateAppPermissionRequest{
			AppID:       "app_123",
			Name:        "Update Invalid Permission Test",
			Description: "A test permission to update with invalid permission",
			Permission:  "read",
			CreatedBy:   "user_123",
		}

		permission, err := appPermissionService.CreateAppPermission(ctx, createReq)
		assert.NoError(t, err)
		assert.NotNil(t, permission)

		// Try to update with invalid permission
		updateReq := &UpdateAppPermissionRequest{
			Permission: "invalid",
		}

		err = appPermissionService.UpdateAppPermission(ctx, permission.ID, updateReq)

		assert.Error(t, err)
		assert.Equal(t, "invalid permission", err.Error())
	})

	// Test AssignPermissionToUser and related functionality
	t.Run("AssignAndCheckPermissionToUser", func(t *testing.T) {
		ctx := context.Background()

		// Create an application and permission
		appID := "app_assign_test"
		userID := "user_assign_test"

		// Create an app permission
		createReq := &CreateAppPermissionRequest{
			AppID:       appID,
			Name:        "Assign Test Permission",
			Description: "A test permission to assign",
			Permission:  "read",
			CreatedBy:   "user_123",
		}

		permission, err := appPermissionService.CreateAppPermission(ctx, createReq)
		assert.NoError(t, err)
		assert.NotNil(t, permission)

		// Assign permission to user
		assignReq := &AssignPermissionToUserRequest{
			AppID:        appID,
			UserID:       userID,
			PermissionID: permission.ID,
			AssignedBy:   "user_123",
		}

		err = appPermissionService.AssignPermissionToUser(ctx, assignReq)
		assert.NoError(t, err)

		// Check if user has permission
		hasPermission, err := appPermissionService.CheckUserPermission(ctx, appID, userID, "read")
		assert.NoError(t, err)
		assert.True(t, hasPermission)

		// Check if user has non-existent permission
		hasPermission, err = appPermissionService.CheckUserPermission(ctx, appID, userID, "write")
		assert.NoError(t, err)
		assert.False(t, hasPermission)

		// List user permissions
		userPermissions, err := appPermissionService.ListUserPermissions(ctx, appID, userID)
		assert.NoError(t, err)
		assert.NotNil(t, userPermissions)
		assert.Len(t, userPermissions, 1)
		assert.Equal(t, permission.ID, userPermissions[0].ID)
	})

	// Test RevokePermissionFromUser
	t.Run("RevokePermissionFromUser", func(t *testing.T) {
		ctx := context.Background()

		// Create an application and permission
		appID := "app_revoke_test_" + time.Now().Format("20060102150405")
		userID := "user_revoke_test_" + time.Now().Format("20060102150405")

		// Create an app permission
		createReq := &CreateAppPermissionRequest{
			AppID:       appID,
			Name:        "Revoke Test Permission",
			Description: "A test permission to revoke",
			Permission:  "read",
			CreatedBy:   "user_123",
		}

		permission, err := appPermissionService.CreateAppPermission(ctx, createReq)
		assert.NoError(t, err)
		assert.NotNil(t, permission)

		// Assign permission to user
		assignReq := &AssignPermissionToUserRequest{
			AppID:        appID,
			UserID:       userID,
			PermissionID: permission.ID,
			AssignedBy:   "user_123",
		}

		err = appPermissionService.AssignPermissionToUser(ctx, assignReq)
		assert.NoError(t, err)

		// Check if user has permission
		hasPermission, err := appPermissionService.CheckUserPermission(ctx, appID, userID, "read")
		assert.NoError(t, err)
		assert.True(t, hasPermission)

		// Revoke permission from user
		revokeReq := &RevokePermissionFromUserRequest{
			AppID:        appID,
			UserID:       userID,
			PermissionID: permission.ID,
		}

		err = appPermissionService.RevokePermissionFromUser(ctx, revokeReq)
		assert.NoError(t, err)

		// Check if user still has permission
		hasPermission, err = appPermissionService.CheckUserPermission(ctx, appID, userID, "read")
		assert.NoError(t, err)
		assert.False(t, hasPermission)
	})

	// Test AssignPermissionToUser with non-existent permission
	t.Run("AssignPermissionToUserNotFound", func(t *testing.T) {
		ctx := context.Background()

		assignReq := &AssignPermissionToUserRequest{
			AppID:        "app_123",
			UserID:       "user_123",
			PermissionID: "non-existent-id",
			AssignedBy:   "user_123",
		}

		err := appPermissionService.AssignPermissionToUser(ctx, assignReq)

		assert.Error(t, err)
		assert.Equal(t, "application permission not found", err.Error())
	})

	// Test RevokePermissionFromUser with non-existent permission (should not error)
	t.Run("RevokePermissionFromUserNotFound", func(t *testing.T) {
		ctx := context.Background()

		revokeReq := &RevokePermissionFromUserRequest{
			AppID:        "app_123",
			UserID:       "user_123",
			PermissionID: "non-existent-id",
		}

		err := appPermissionService.RevokePermissionFromUser(ctx, revokeReq)

		assert.NoError(t, err) // Should not error when permission doesn't exist
	})

	// Test multiple app permission operations
	t.Run("MultipleAppPermissionOperations", func(t *testing.T) {
		ctx := context.Background()

		// Create multiple app permissions
		permissionsData := []struct {
			appID       string
			name        string
			description string
			permission  string
		}{
			{"app_501", "Multi Test 1", "First multi test permission", "read"},
			{"app_501", "Multi Test 2", "Second multi test permission", "write"},
			{"app_501", "Multi Test 3", "Third multi test permission", "delete"},
		}

		var createdPermissions []*domain.AppPermission
		for _, data := range permissionsData {
			req := &CreateAppPermissionRequest{
				AppID:       data.appID,
				Name:        data.name,
				Description: data.description,
				Permission:  data.permission,
				CreatedBy:   "user_123",
			}

			permission, err := appPermissionService.CreateAppPermission(ctx, req)
			assert.NoError(t, err)
			assert.NotNil(t, permission)
			createdPermissions = append(createdPermissions, permission)
		}

		// Update all app permissions
		for _, permission := range createdPermissions {
			updateReq := &UpdateAppPermissionRequest{
				Description: "Updated " + permission.Description,
			}

			err := appPermissionService.UpdateAppPermission(ctx, permission.ID, updateReq)
			assert.NoError(t, err)
		}

		// Verify updates
		for _, permission := range createdPermissions {
			updatedPermission, err := appPermissionService.GetAppPermission(ctx, permission.ID)
			assert.NoError(t, err)
			assert.NotNil(t, updatedPermission)
			assert.Contains(t, updatedPermission.Description, "Updated")
		}

		// Delete all app permissions
		for _, permission := range createdPermissions {
			err := appPermissionService.DeleteAppPermission(ctx, permission.ID)
			assert.NoError(t, err)
		}

		// Verify deletions
		for _, permission := range createdPermissions {
			_, err := appPermissionService.GetAppPermission(ctx, permission.ID)
			assert.Error(t, err)
			assert.Equal(t, "application permission not found", err.Error())
		}
	})
}