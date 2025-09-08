package test

import (
	"context"
	"testing"
	"time"

	"cdk-office/internal/app/domain"
	"cdk-office/internal/app/service"
	"cdk-office/internal/shared/testutils"
	"github.com/stretchr/testify/assert"
)

func TestPermissionService_CreatePermission(t *testing.T) {
	// Setup
	db := testutils.SetupTestDB()
	defer db.Migrator().DropTable(&domain.AppPermission{}, &domain.Application{})

	permissionService := service.NewPermissionServiceWithDB(db)

	// Create a test application first
	app := &domain.Application{
		ID:        "app-001",
		TeamID:    "team-001",
		Name:      "Test App",
		Type:      "form",
		CreatedBy: "user-001",
	}
	err := db.Create(app).Error
	assert.NoError(t, err)

	// Test cases
	tests := []struct {
		name          string
		request       *service.CreatePermissionRequest
		expectError   bool
		errorMessage  string
	}{
		{
			name: "Valid permission creation",
			request: &service.CreatePermissionRequest{
				AppID:       "app-001",
				Name:        "Read Permission",
				Description: "Permission to read data",
				Action:      "read",
				CreatedBy:   "user-001",
			},
			expectError: false,
		},
		{
			name: "Invalid permission action",
			request: &service.CreatePermissionRequest{
				AppID:       "app-001",
				Name:        "Invalid Permission",
				Description: "Invalid permission",
				Action:      "invalid",
				CreatedBy:   "user-001",
			},
			expectError:  true,
			errorMessage: "invalid permission action",
		},
		{
			name: "Duplicate permission name in same app",
			request: &service.CreatePermissionRequest{
				AppID:       "app-001",
				Name:        "Read Permission",
				Description: "Another read permission",
				Action:      "read",
				CreatedBy:   "user-001",
			},
			expectError:  true,
			errorMessage: "permission with this name already exists in the application",
		},
		{
			name: "Non-existent application",
			request: &service.CreatePermissionRequest{
				AppID:       "non-existent-app",
				Name:        "Orphan Permission",
				Description: "Permission for non-existent app",
				Action:      "read",
				CreatedBy:   "user-001",
			},
			expectError: false, // Note: This doesn't validate app existence in the service
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Execute
			permission, err := permissionService.CreatePermission(context.Background(), tt.request)

			// Assert
			if tt.expectError {
				assert.Error(t, err)
				assert.Equal(t, tt.errorMessage, err.Error())
				assert.Nil(t, permission)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, permission)
				assert.NotEmpty(t, permission.ID)
				assert.Equal(t, tt.request.AppID, permission.AppID)
				assert.Equal(t, tt.request.Name, permission.Name)
				assert.Equal(t, tt.request.Description, permission.Description)
				assert.Equal(t, tt.request.Action, permission.Permission)
				assert.Equal(t, tt.request.CreatedBy, permission.CreatedBy)
				assert.WithinDuration(t, time.Now(), permission.CreatedAt, time.Second)
				assert.WithinDuration(t, time.Now(), permission.UpdatedAt, time.Second)
			}
		})
	}
}

func TestPermissionService_UpdatePermission(t *testing.T) {
	// Setup
	db := testutils.SetupTestDB()
	defer db.Migrator().DropTable(&domain.AppPermission{}, &domain.Application{})

	permissionService := service.NewPermissionServiceWithDB(db)

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

	// Create a permission for testing
	permission := &domain.AppPermission{
		ID:          "perm-001",
		AppID:       "app-001",
		Name:        "Original Permission",
		Description: "Original description",
		Permission:  "read",
		CreatedBy:   "user-001",
		CreatedAt:   time.Now().Add(-time.Hour),
		UpdatedAt:   time.Now().Add(-time.Hour),
	}
	err = db.Create(permission).Error
	assert.NoError(t, err)

	// Test cases
	tests := []struct {
		name          string
		permissionID  string
		request       *service.UpdatePermissionRequest
		expectError   bool
		errorMessage  string
	}{
		{
			name:         "Valid permission update",
			permissionID: "perm-001",
			request: &service.UpdatePermissionRequest{
				Name:        "Updated Permission",
				Description: "Updated description",
				Action:      "write",
			},
			expectError: false,
		},
		{
			name:         "Update non-existent permission",
			permissionID: "non-existent-id",
			request: &service.UpdatePermissionRequest{
				Name: "Updated Name",
			},
			expectError:  true,
			errorMessage: "permission not found",
		},
		{
			name:         "Invalid permission action",
			permissionID: "perm-001",
			request: &service.UpdatePermissionRequest{
				Action: "invalid",
			},
			expectError:  true,
			errorMessage: "invalid permission",
		},
		{
			name:         "Partial update - name only",
			permissionID: "perm-001",
			request: &service.UpdatePermissionRequest{
				Name: "Name Only Update",
			},
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Execute
			err := permissionService.UpdatePermission(context.Background(), tt.permissionID, tt.request)

			// Assert
			if tt.expectError {
				assert.Error(t, err)
				assert.Equal(t, tt.errorMessage, err.Error())
			} else {
				assert.NoError(t, err)

				// Verify the update
				updatedPermission, getErr := permissionService.GetPermission(context.Background(), tt.permissionID)
				assert.NoError(t, getErr)
				assert.NotNil(t, updatedPermission)

				// Check updated fields
				if tt.request.Name != "" {
					assert.Equal(t, tt.request.Name, updatedPermission.Name)
				}
				if tt.request.Description != "" {
					assert.Equal(t, tt.request.Description, updatedPermission.Description)
				}
				if tt.request.Action != "" {
					assert.Equal(t, tt.request.Action, updatedPermission.Permission)
				}
				// UpdatedAt should be changed
				assert.True(t, updatedPermission.UpdatedAt.After(permission.UpdatedAt))
			}
		})
	}
}

func TestPermissionService_DeletePermission(t *testing.T) {
	// Setup
	db := testutils.SetupTestDB()
	defer db.Migrator().DropTable(&domain.AppPermission{}, &domain.AppUserPermission{}, &domain.Application{})

	permissionService := service.NewPermissionServiceWithDB(db)

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

	// Create a permission for testing
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

	// Test cases
	tests := []struct {
		name          string
		permissionID  string
		expectError   bool
		errorMessage  string
	}{
		{
			name:         "Delete non-existent permission",
			permissionID: "non-existent-id",
			expectError:  true,
			errorMessage: "permission not found",
		},
		{
			name:         "Valid permission deletion",
			permissionID: "perm-001",
			expectError:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Execute
			err := permissionService.DeletePermission(context.Background(), tt.permissionID)

			// Assert
			if tt.expectError {
				assert.Error(t, err)
				assert.Equal(t, tt.errorMessage, err.Error())
			} else {
				assert.NoError(t, err)

				// Verify permission is deleted
				_, getErr := permissionService.GetPermission(context.Background(), tt.permissionID)
				assert.Error(t, getErr)
				assert.Equal(t, "permission not found", getErr.Error())

				// Verify user permissions are also deleted
				var count int64
				countErr := db.Model(&domain.AppUserPermission{}).Where("permission_id = ?", tt.permissionID).Count(&count).Error
				assert.NoError(t, countErr)
				assert.Equal(t, int64(0), count)
			}
		})
	}
}

func TestPermissionService_ListPermissions(t *testing.T) {
	// Setup
	db := testutils.SetupTestDB()
	defer db.Migrator().DropTable(&domain.AppPermission{}, &domain.Application{})

	permissionService := service.NewPermissionServiceWithDB(db)

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

	// Create test permissions
	permissions := []*domain.AppPermission{
		{
			ID:          "perm-001",
			AppID:       "app-001",
			Name:        "Permission 1",
			Description: "First permission",
			Permission:  "read",
			CreatedBy:   "user-001",
			CreatedAt:   time.Now().Add(-2 * time.Hour),
		},
		{
			ID:          "perm-002",
			AppID:       "app-001",
			Name:        "Permission 2",
			Description: "Second permission",
			Permission:  "write",
			CreatedBy:   "user-001",
			CreatedAt:   time.Now().Add(-1 * time.Hour),
		},
		{
			ID:          "perm-003",
			AppID:       "app-001",
			Name:        "Permission 3",
			Description: "Third permission",
			Permission:  "delete",
			CreatedBy:   "user-001",
			CreatedAt:   time.Now(),
		},
	}

	for _, perm := range permissions {
		err := db.Create(perm).Error
		assert.NoError(t, err)
	}

	// Test cases
	tests := []struct {
		name              string
		appID             string
		page              int
		size              int
		expectedCount     int
		totalCount        int64
		expectError       bool
	}{
		{
			name:          "List first page",
			appID:         "app-001",
			page:          1,
			size:          2,
			expectedCount: 2,
			totalCount:    3,
			expectError:   false,
		},
		{
			name:          "List second page",
			appID:         "app-001",
			page:          2,
			size:          2,
			expectedCount: 1,
			totalCount:    3,
			expectError:   false,
		},
		{
			name:          "List with large page size",
			appID:         "app-001",
			page:          1,
			size:          10,
			expectedCount: 3,
			totalCount:    3,
			expectError:   false,
		},
		{
			name:          "List with zero page",
			appID:         "app-001",
			page:          0,
			size:          10,
			expectedCount: 3,
			totalCount:    3,
			expectError:   false,
		},
		{
			name:          "List non-existent app",
			appID:         "non-existent-app",
			page:          1,
			size:          10,
			expectedCount: 0,
			totalCount:    0,
			expectError:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Execute
			result, total, err := permissionService.ListPermissions(context.Background(), tt.appID, tt.page, tt.size)

			// Assert
			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedCount, len(result))
				assert.Equal(t, tt.totalCount, total)

				// Verify ordering (should be by created_at desc)
				if len(result) > 1 {
					for i := 0; i < len(result)-1; i++ {
						assert.True(t, result[i].CreatedAt.After(result[i+1].CreatedAt) || 
							result[i].CreatedAt.Equal(result[i+1].CreatedAt))
					}
				}
			}
		})
	}
}

func TestPermissionService_GetPermission(t *testing.T) {
	// Setup
	db := testutils.SetupTestDB()
	defer db.Migrator().DropTable(&domain.AppPermission{}, &domain.Application{})

	permissionService := service.NewPermissionServiceWithDB(db)

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

	// Create a permission for testing
	permission := &domain.AppPermission{
		ID:          "perm-001",
		AppID:       "app-001",
		Name:        "Test Permission",
		Description: "Test description",
		Permission:  "read",
		CreatedBy:   "user-001",
		CreatedAt:   time.Now().Add(-time.Hour),
		UpdatedAt:   time.Now().Add(-time.Hour),
	}
	err = db.Create(permission).Error
	assert.NoError(t, err)

	// Test cases
	tests := []struct {
		name          string
		permissionID  string
		expectError   bool
		errorMessage  string
	}{
		{
			name:          "Get non-existent permission",
			permissionID:  "non-existent-id",
			expectError:   true,
			errorMessage:  "permission not found",
		},
		{
			name:          "Get existing permission",
			permissionID:  "perm-001",
			expectError:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Execute
			result, err := permissionService.GetPermission(context.Background(), tt.permissionID)

			// Assert
			if tt.expectError {
				assert.Error(t, err)
				assert.Equal(t, tt.errorMessage, err.Error())
				assert.Nil(t, result)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, result)
				assert.Equal(t, permission.ID, result.ID)
				assert.Equal(t, permission.AppID, result.AppID)
				assert.Equal(t, permission.Name, result.Name)
				assert.Equal(t, permission.Description, result.Description)
				assert.Equal(t, permission.Permission, result.Permission)
				assert.Equal(t, permission.CreatedBy, result.CreatedBy)
				assert.Equal(t, permission.CreatedAt.Unix(), result.CreatedAt.Unix())
				assert.Equal(t, permission.UpdatedAt.Unix(), result.UpdatedAt.Unix())
			}
		})
	}
}

func TestPermissionService_CheckPermission(t *testing.T) {
	// Setup
	db := testutils.SetupTestDB()
	defer db.Migrator().DropTable(&domain.AppPermission{}, &domain.AppUserPermission{}, &domain.Application{})

	permissionService := service.NewPermissionServiceWithDB(db)

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

	// Create a permission for testing
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

	// Test cases
	tests := []struct {
		name          string
		appID         string
		userID        string
		permission    string
		expected      bool
		expectError   bool
		errorMessage  string
	}{
		{
			name:        "User has permission",
			appID:       "app-001",
			userID:      "user-001",
			permission:  "read",
			expected:    true,
			expectError: false,
		},
		{
			name:        "User doesn't have permission",
			appID:       "app-001",
			userID:      "user-002",
			permission:  "read",
			expected:    false,
			expectError: false,
		},
		{
			name:        "User has no permissions for app",
			appID:       "app-001",
			userID:      "user-003",
			permission:  "read",
			expected:    false,
			expectError: false,
		},
		{
			name:        "Non-existent permission",
			appID:       "app-001",
			userID:      "user-001",
			permission:  "admin",
			expected:    false,
			expectError: false,
		},
		{
			name:        "Non-existent app",
			appID:       "non-existent-app",
			userID:      "user-001",
			permission:  "read",
			expected:    false,
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Execute
			result, err := permissionService.CheckPermission(context.Background(), tt.appID, tt.userID, tt.permission)

			// Assert
			if tt.expectError {
				assert.Error(t, err)
				assert.Equal(t, tt.errorMessage, err.Error())
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expected, result)
			}
		})
	}
}

func TestPermissionService_AssignPermission(t *testing.T) {
	// Setup
	db := testutils.SetupTestDB()
	defer db.Migrator().DropTable(&domain.AppPermission{}, &domain.AppUserPermission{}, &domain.Application{})

	permissionService := service.NewPermissionServiceWithDB(db)

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

	// Create a permission for testing
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

	// Test cases
	tests := []struct {
		name          string
		request       *service.AssignPermissionRequest
		expectError   bool
		errorMessage  string
	}{
		{
			name: "Valid permission assignment",
			request: &service.AssignPermissionRequest{
				AppID:        "app-001",
				UserID:       "user-001",
				PermissionID: "perm-001",
				AssignedBy:   "admin",
			},
			expectError: false,
		},
		{
			name: "Assign non-existent permission",
			request: &service.AssignPermissionRequest{
				AppID:        "app-001",
				UserID:       "user-001",
				PermissionID: "non-existent-id",
				AssignedBy:   "admin",
			},
			expectError:  true,
			errorMessage: "permission not found",
		},
		{
			name: "Assign permission to wrong app",
			request: &service.AssignPermissionRequest{
				AppID:        "wrong-app",
				UserID:       "user-001",
				PermissionID: "perm-001",
				AssignedBy:   "admin",
			},
			expectError:  true,
			errorMessage: "permission does not belong to this application",
		},
		{
			name: "Assign already assigned permission",
			request: &service.AssignPermissionRequest{
				AppID:        "app-001",
				UserID:       "user-001",
				PermissionID: "perm-001",
				AssignedBy:   "admin",
			},
			expectError: false, // Should not error, just return success
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Execute
			err := permissionService.AssignPermission(context.Background(), tt.request)

			// Assert
			if tt.expectError {
				assert.Error(t, err)
				assert.Equal(t, tt.errorMessage, err.Error())
			} else {
				assert.NoError(t, err)

				// For the first valid assignment, verify it was created
				if tt.name == "Valid permission assignment" {
					var count int64
					countErr := db.Model(&domain.AppUserPermission{}).Where("app_id = ? AND user_id = ? AND permission_id = ?", 
						tt.request.AppID, tt.request.UserID, tt.request.PermissionID).Count(&count).Error
					assert.NoError(t, countErr)
					assert.Equal(t, int64(1), count)
				}
			}
		})
	}
}

func TestPermissionService_RevokePermission(t *testing.T) {
	// Setup
	db := testutils.SetupTestDB()
	defer db.Migrator().DropTable(&domain.AppPermission{}, &domain.AppUserPermission{}, &domain.Application{})

	permissionService := service.NewPermissionServiceWithDB(db)

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

	// Create a permission for testing
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

	// Test cases
	tests := []struct {
		name          string
		request       *service.RevokePermissionRequest
		expectError   bool
		errorMessage  string
	}{
		{
			name: "Revoke non-existent permission",
			request: &service.RevokePermissionRequest{
				AppID:        "app-001",
				UserID:       "user-001",
				PermissionID: "non-existent-id",
			},
			expectError: false, // Should not error, just return success
		},
		{
			name: "Valid permission revocation",
			request: &service.RevokePermissionRequest{
				AppID:        "app-001",
				UserID:       "user-001",
				PermissionID: "perm-001",
			},
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Execute
			err := permissionService.RevokePermission(context.Background(), tt.request)

			// Assert
			if tt.expectError {
				assert.Error(t, err)
				assert.Equal(t, tt.errorMessage, err.Error())
			} else {
				assert.NoError(t, err)

				// For the valid revocation, verify it was deleted
				if tt.name == "Valid permission revocation" {
					var count int64
					countErr := db.Model(&domain.AppUserPermission{}).Where("app_id = ? AND user_id = ? AND permission_id = ?", 
						tt.request.AppID, tt.request.UserID, tt.request.PermissionID).Count(&count).Error
					assert.NoError(t, countErr)
					assert.Equal(t, int64(0), count)
				}
			}
		})
	}
}

func TestPermissionService_ListUserPermissions(t *testing.T) {
	// Setup
	db := testutils.SetupTestDB()
	defer db.Migrator().DropTable(&domain.AppPermission{}, &domain.AppUserPermission{}, &domain.Application{})

	permissionService := service.NewPermissionServiceWithDB(db)

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

	// Create test permissions
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
		err := db.Create(perm).Error
		assert.NoError(t, err)
	}

	// Create user permissions for testing
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
		err := db.Create(userPerm).Error
		assert.NoError(t, err)
	}

	// Test cases
	tests := []struct {
		name          string
		appID         string
		userID        string
		expectedCount int
		expectError   bool
		errorMessage  string
	}{
		{
			name:          "List user permissions",
			appID:         "app-001",
			userID:        "user-001",
			expectedCount: 2,
			expectError:   false,
		},
		{
			name:          "List permissions for user with no permissions",
			appID:         "app-001",
			userID:        "user-002",
			expectedCount: 0,
			expectError:   false,
		},
		{
			name:          "List permissions for non-existent app",
			appID:         "non-existent-app",
			userID:        "user-001",
			expectedCount: 0,
			expectError:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Execute
			result, err := permissionService.ListUserPermissions(context.Background(), tt.appID, tt.userID)

			// Assert
			if tt.expectError {
				assert.Error(t, err)
				assert.Equal(t, tt.errorMessage, err.Error())
				assert.Nil(t, result)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedCount, len(result))

				// Verify that the returned permissions are correct
				if tt.expectedCount > 0 {
					permissionIDs := make(map[string]bool)
					for _, perm := range permissions {
						permissionIDs[perm.ID] = true
					}

					for _, perm := range result {
						assert.True(t, permissionIDs[perm.ID])
					}
				}
			}
		})
	}
}