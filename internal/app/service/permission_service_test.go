package service

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"cdk-office/internal/shared/testutils"
)

// TestPermissionService tests the PermissionService
func TestPermissionService(t *testing.T) {
	// Set up test environment
	testDB := testutils.SetupTestDB()

	// Create permission service with database connection
	permissionService := NewPermissionServiceWithDB(testDB)

	// Test CreatePermission
	t.Run("CreatePermission", func(t *testing.T) {
		ctx := context.Background()
		req := &CreatePermissionRequest{
			AppID:       "app_123",
			Name:        "Read Permission",
			Action:      "read",
			Description: "Permission to read data",
			CreatedBy:   "user_123",
		}

		permission, err := permissionService.CreatePermission(ctx, req)

		assert.NoError(t, err)
		assert.NotNil(t, permission)
		assert.Equal(t, "app_123", permission.AppID)
		assert.Equal(t, "Read Permission", permission.Name)
		assert.Equal(t, "read", permission.Permission)
	})

	// Test UpdatePermission
	t.Run("UpdatePermission", func(t *testing.T) {
		ctx := context.Background()

		// First create a permission
		createReq := &CreatePermissionRequest{
			AppID:     "app_123",
			Name:      "Update Test Permission",
			Action:    "read",
			CreatedBy: "user_123",
		}

		permission, err := permissionService.CreatePermission(ctx, createReq)
		assert.NoError(t, err)
		assert.NotNil(t, permission)

		// Now update the permission
		updateReq := &UpdatePermissionRequest{
			Name:        "Updated Permission",
			Description: "Updated description",
			Action:      "write",
		}

		err = permissionService.UpdatePermission(ctx, permission.ID, updateReq)
		assert.NoError(t, err)
	})

	// Test DeletePermission
	t.Run("DeletePermission", func(t *testing.T) {
		ctx := context.Background()

		// First create a permission
		createReq := &CreatePermissionRequest{
			AppID:     "app_123",
			Name:      "Delete Test Permission",
			Action:    "read",
			CreatedBy: "user_123",
		}

		permission, err := permissionService.CreatePermission(ctx, createReq)
		assert.NoError(t, err)
		assert.NotNil(t, permission)

		// Now delete the permission
		err = permissionService.DeletePermission(ctx, permission.ID)
		assert.NoError(t, err)
	})

	// Test ListPermissions
	t.Run("ListPermissions", func(t *testing.T) {
		ctx := context.Background()

		// Create a few permissions
		for i := 1; i <= 3; i++ {
			req := &CreatePermissionRequest{
				AppID:     "app_list",
				Name:      "List Test Permission " + string(rune(i+'0')),
				Action:    "read",
				CreatedBy: "user_123",
			}

			_, err := permissionService.CreatePermission(ctx, req)
			assert.NoError(t, err)
		}

		// List permissions
		permissions, total, err := permissionService.ListPermissions(ctx, "app_list", 1, 10)
		assert.NoError(t, err)
		assert.NotNil(t, permissions)
		assert.GreaterOrEqual(t, total, int64(3))
		assert.GreaterOrEqual(t, len(permissions), 3)
	})

	// Test GetPermission
	t.Run("GetPermission", func(t *testing.T) {
		ctx := context.Background()

		// First create a permission
		createReq := &CreatePermissionRequest{
			AppID:     "app_123",
			Name:      "Get Test Permission",
			Action:    "read",
			CreatedBy: "user_123",
		}

		createdPermission, err := permissionService.CreatePermission(ctx, createReq)
		assert.NoError(t, err)
		assert.NotNil(t, createdPermission)

		// Now get the permission
		retrievedPermission, err := permissionService.GetPermission(ctx, createdPermission.ID)
		assert.NoError(t, err)
		assert.NotNil(t, retrievedPermission)
		assert.Equal(t, createdPermission.ID, retrievedPermission.ID)
		assert.Equal(t, createdPermission.Name, retrievedPermission.Name)
	})
}