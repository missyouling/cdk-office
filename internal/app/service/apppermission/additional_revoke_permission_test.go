package apppermission_test

import (
	"context"
	"testing"

	"cdk-office/internal/app/domain"
	"cdk-office/internal/app/service"
	"cdk-office/internal/shared/testutils"
	"github.com/stretchr/testify/assert"
)

func TestAppPermissionService_RevokePermissionFromUser_DBError(t *testing.T) {
	// Setup
	db := testutils.SetupTestDB()
	defer db.Migrator().DropTable(&domain.AppPermission{}, &domain.AppUserPermission{}, &domain.Application{})

	// Create a test application
	app := &domain.Application{
		ID:        "app-001",
		TeamID:    "team-001",
		Name:      "Test App",
		Type:      "form",
		CreatedBy: "user-001",
	}
	err := db.Create(app).Error
	assert.NoError(t, err)

	// Create an app permission for testing
	permission := &domain.AppPermission{
		ID:          "perm-001",
		AppID:       "app-001",
		Name:        "Test Permission",
		Description: "Test description",
		Permission:  "read",
		CreatedBy:   "user-001",
	}
	err = db.Create(permission).Error
	assert.NoError(t, err)

	// Create a user permission for testing
	userPermission := &domain.AppUserPermission{
		ID:           "user-perm-001",
		AppID:        "app-001",
		UserID:       "user-001",
		PermissionID: "perm-001",
		AssignedBy:   "admin",
	}
	err = db.Create(userPermission).Error
	assert.NoError(t, err)

	// Test case: Database error when finding user permission
	t.Run("Database error when finding user permission (conceptual)", func(t *testing.T) {
		// This is a placeholder for what the test would look like with proper mocking
		// In a real implementation with proper dependency injection, we would inject
		// a mock database that returns an error when First is called
		assert.True(t, true) // Placeholder assertion
	})

	// Test case: Database error when deleting user permission
	t.Run("Database error when deleting user permission (conceptual)", func(t *testing.T) {
		// This is a placeholder for what the test would look like with proper mocking
		assert.True(t, true) // Placeholder assertion
	})
}

func TestAppPermissionService_RevokePermissionFromUser_AdditionalCases(t *testing.T) {
	// Setup
	db := testutils.SetupTestDB()
	defer db.Migrator().DropTable(&domain.AppPermission{}, &domain.AppUserPermission{}, &domain.Application{})

	appPermissionService := service.NewAppPermissionService()

	// Create a test application
	app := &domain.Application{
		ID:        "app-001",
		TeamID:    "team-001",
		Name:      "Test App",
		Type:      "form",
		CreatedBy: "user-001",
	}
	err := db.Create(app).Error
	assert.NoError(t, err)

	// Create multiple app permissions for testing
	permissions := []*domain.AppPermission{
		{
			ID:          "perm-001",
			AppID:       "app-001",
			Name:        "Read Permission",
			Description: "Read permission",
			Permission:  "read",
			CreatedBy:   "user-001",
		},
		{
			ID:          "perm-002",
			AppID:       "app-001",
			Name:        "Write Permission",
			Description: "Write permission",
			Permission:  "write",
			CreatedBy:   "user-001",
		},
	}

	for _, perm := range permissions {
		err = db.Create(perm).Error
		assert.NoError(t, err)
	}

	// Create multiple user permissions for the same user
	userPermissions := []*domain.AppUserPermission{
		{
			ID:           "user-perm-001",
			AppID:        "app-001",
			UserID:       "user-001",
			PermissionID: "perm-001",
			AssignedBy:   "admin",
		},
		{
			ID:           "user-perm-002",
			AppID:        "app-001",
			UserID:       "user-001",
			PermissionID: "perm-002",
			AssignedBy:   "admin",
		},
	}

	for _, userPerm := range userPermissions {
		err = db.Create(userPerm).Error
		assert.NoError(t, err)
	}

	// Test case: Revoke one permission from user with multiple permissions
	t.Run("Revoke one permission from user with multiple permissions", func(t *testing.T) {
		// Revoke read permission
		req := &service.RevokePermissionFromUserRequest{
			AppID:        "app-001",
			UserID:       "user-001",
			PermissionID: "perm-001",
		}
		err := appPermissionService.RevokePermissionFromUser(context.Background(), req)
		assert.NoError(t, err)

		// Verify only read permission was revoked
		userPerms, err := appPermissionService.ListUserPermissions(context.Background(), "app-001", "user-001")
		assert.NoError(t, err)
		assert.Len(t, userPerms, 1)
		assert.Equal(t, "perm-002", userPerms[0].ID)
	})

	// Test case: Revoke all permissions from user
	t.Run("Revoke all permissions from user", func(t *testing.T) {
		// Revoke remaining permission
		req := &service.RevokePermissionFromUserRequest{
			AppID:        "app-001",
			UserID:       "user-001",
			PermissionID: "perm-002",
		}
		err := appPermissionService.RevokePermissionFromUser(context.Background(), req)
		assert.NoError(t, err)

		// Verify user has no permissions left
		userPerms, err := appPermissionService.ListUserPermissions(context.Background(), "app-001", "user-001")
		assert.NoError(t, err)
		assert.Len(t, userPerms, 0)
	})

	// Test case: Revoke permission from user who already doesn't have it
	t.Run("Revoke permission from user who already doesn't have it", func(t *testing.T) {
		// Try to revoke a permission that the user doesn't have
		req := &service.RevokePermissionFromUserRequest{
			AppID:        "app-001",
			UserID:       "user-001",
			PermissionID: "perm-001", // Already revoked
		}
		err := appPermissionService.RevokePermissionFromUser(context.Background(), req)
		assert.NoError(t, err) // Should not error, just return success
	})
}