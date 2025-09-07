package service

import (
	"context"
	"errors"
	"time"

	"cdk-office/internal/app/domain"
	"cdk-office/internal/shared/database"
	"cdk-office/internal/shared/utils"
	"cdk-office/pkg/logger"
	"gorm.io/gorm"
)

// AppPermissionServiceInterface defines the interface for application permission service
type AppPermissionServiceInterface interface {
	CreateAppPermission(ctx context.Context, req *CreateAppPermissionRequest) (*domain.AppPermission, error)
	UpdateAppPermission(ctx context.Context, permissionID string, req *UpdateAppPermissionRequest) error
	DeleteAppPermission(ctx context.Context, permissionID string) error
	ListAppPermissions(ctx context.Context, appID string, page, size int) ([]*domain.AppPermission, int64, error)
	GetAppPermission(ctx context.Context, permissionID string) (*domain.AppPermission, error)
	AssignPermissionToUser(ctx context.Context, req *AssignPermissionToUserRequest) error
	RevokePermissionFromUser(ctx context.Context, req *RevokePermissionFromUserRequest) error
	ListUserPermissions(ctx context.Context, appID, userID string) ([]*domain.AppPermission, error)
	CheckUserPermission(ctx context.Context, appID, userID, permission string) (bool, error)
}

// AppPermissionService implements the AppPermissionServiceInterface
type AppPermissionService struct {
	db *gorm.DB
}

// NewAppPermissionService creates a new instance of AppPermissionService
func NewAppPermissionService() *AppPermissionService {
	return &AppPermissionService{
		db: database.GetDB(),
	}
}

// CreateAppPermissionRequest represents the request for creating an application permission
type CreateAppPermissionRequest struct {
	AppID       string `json:"app_id" binding:"required"`
	Name        string `json:"name" binding:"required"`
	Description string `json:"description"`
	Permission  string `json:"permission" binding:"required"` // read, write, delete, manage
	CreatedBy   string `json:"created_by" binding:"required"`
}

// UpdateAppPermissionRequest represents the request for updating an application permission
type UpdateAppPermissionRequest struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	Permission  string `json:"permission"` // read, write, delete, manage
}

// AssignPermissionToUserRequest represents the request for assigning a permission to a user
type AssignPermissionToUserRequest struct {
	AppID        string `json:"app_id" binding:"required"`
	UserID       string `json:"user_id" binding:"required"`
	PermissionID string `json:"permission_id" binding:"required"`
	AssignedBy   string `json:"assigned_by" binding:"required"`
}

// RevokePermissionFromUserRequest represents the request for revoking a permission from a user
type RevokePermissionFromUserRequest struct {
	AppID        string `json:"app_id" binding:"required"`
	UserID       string `json:"user_id" binding:"required"`
	PermissionID string `json:"permission_id" binding:"required"`
}



// CreateAppPermission creates a new application permission
func (s *AppPermissionService) CreateAppPermission(ctx context.Context, req *CreateAppPermissionRequest) (*domain.AppPermission, error) {
	// Validate permission
	validPermissions := map[string]bool{
		"read":   true,
		"write":  true,
		"delete": true,
		"manage": true,
	}

	if !validPermissions[req.Permission] {
		return nil, errors.New("invalid permission")
	}

	// Check if permission with the same name already exists in the app
	var existingPermission domain.AppPermission
	if err := s.db.Table("app_permissions").Where("app_id = ? AND name = ?", req.AppID, req.Name).First(&existingPermission).Error; err == nil {
		return nil, errors.New("permission with this name already exists in the application")
	}

	// Create new application permission
	permission := &domain.AppPermission{
		ID:          utils.GeneratePermissionID(),
		AppID:       req.AppID,
		Name:        req.Name,
		Description: req.Description,
		Permission:  req.Permission,
		CreatedBy:   req.CreatedBy,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	// Save application permission to database
	if err := s.db.Table("app_permissions").Create(permission).Error; err != nil {
		logger.Error("failed to create application permission", "error", err)
		return nil, errors.New("failed to create application permission")
	}

	return permission, nil
}

// UpdateAppPermission updates an existing application permission
func (s *AppPermissionService) UpdateAppPermission(ctx context.Context, permissionID string, req *UpdateAppPermissionRequest) error {
	// Find application permission by ID
	var permission domain.AppPermission
	if err := s.db.Table("app_permissions").Where("id = ?", permissionID).First(&permission).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errors.New("application permission not found")
		}
		logger.Error("failed to find application permission", "error", err)
		return errors.New("failed to update application permission")
	}

	// Update application permission fields
	if req.Name != "" {
		permission.Name = req.Name
	}

	if req.Description != "" {
		permission.Description = req.Description
	}

	if req.Permission != "" {
		// Validate permission
		validPermissions := map[string]bool{
			"read":   true,
			"write":  true,
			"delete": true,
			"manage": true,
		}

		if !validPermissions[req.Permission] {
			return errors.New("invalid permission")
		}
		permission.Permission = req.Permission
	}

	permission.UpdatedAt = time.Now()

	// Save updated application permission to database
	if err := s.db.Table("app_permissions").Save(&permission).Error; err != nil {
		logger.Error("failed to update application permission", "error", err)
		return errors.New("failed to update application permission")
	}

	return nil
}

// DeleteAppPermission deletes an application permission
func (s *AppPermissionService) DeleteAppPermission(ctx context.Context, permissionID string) error {
	// Find application permission by ID
	var permission domain.AppPermission
	if err := s.db.Table("app_permissions").Where("id = ?", permissionID).First(&permission).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errors.New("application permission not found")
		}
		logger.Error("failed to find application permission", "error", err)
		return errors.New("failed to delete application permission")
	}

	// Delete all user permissions associated with this permission
	if err := s.db.Table("app_user_permissions").Where("permission_id = ?", permissionID).Delete(&domain.AppUserPermission{}).Error; err != nil {
		logger.Error("failed to delete user permissions", "error", err)
		return errors.New("failed to delete user permissions")
	}

	// Delete application permission from database
	if err := s.db.Table("app_permissions").Delete(&permission).Error; err != nil {
		logger.Error("failed to delete application permission", "error", err)
		return errors.New("failed to delete application permission")
	}

	return nil
}

// ListAppPermissions lists application permissions with pagination
func (s *AppPermissionService) ListAppPermissions(ctx context.Context, appID string, page, size int) ([]*domain.AppPermission, int64, error) {
	// Validate pagination parameters
	if page < 1 {
		page = 1
	}
	if size < 1 || size > 100 {
		size = 10
	}

	// Build query
	dbQuery := s.db.Table("app_permissions").Where("app_id = ?", appID)

	// Count total results
	var total int64
	if err := dbQuery.Count(&total).Error; err != nil {
		logger.Error("failed to count application permissions", "error", err)
		return nil, 0, errors.New("failed to list application permissions")
	}

	// Apply pagination
	offset := (page - 1) * size
	dbQuery = dbQuery.Offset(offset).Limit(size).Order("created_at desc")

	// Execute query
	var permissions []*domain.AppPermission
	if err := dbQuery.Find(&permissions).Error; err != nil {
		logger.Error("failed to list application permissions", "error", err)
		return nil, 0, errors.New("failed to list application permissions")
	}

	return permissions, total, nil
}

// GetAppPermission retrieves an application permission by ID
func (s *AppPermissionService) GetAppPermission(ctx context.Context, permissionID string) (*domain.AppPermission, error) {
	var permission domain.AppPermission
	if err := s.db.Table("app_permissions").Where("id = ?", permissionID).First(&permission).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("application permission not found")
		}
		logger.Error("failed to find application permission", "error", err)
		return nil, errors.New("failed to get application permission")
	}

	return &permission, nil
}

// AssignPermissionToUser assigns a permission to a user
func (s *AppPermissionService) AssignPermissionToUser(ctx context.Context, req *AssignPermissionToUserRequest) error {
	// Verify application permission exists
	var permission domain.AppPermission
	if err := s.db.Table("app_permissions").Where("id = ?", req.PermissionID).First(&permission).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errors.New("application permission not found")
		}
		logger.Error("failed to find application permission", "error", err)
		return errors.New("failed to assign permission to user")
	}

	// Verify permission belongs to the specified app
	if permission.AppID != req.AppID {
		return errors.New("application permission does not belong to the specified application")
	}

	// Check if user already has this permission
	var existing domain.AppUserPermission
	if err := s.db.Table("app_user_permissions").Where("app_id = ? AND user_id = ? AND permission_id = ?", req.AppID, req.UserID, req.PermissionID).First(&existing).Error; err == nil {
		// User already has this permission
		return nil
	}

	// Create new user permission
	userPermission := &domain.AppUserPermission{
		ID:           utils.GenerateUserPermissionID(),
		AppID:        req.AppID,
		UserID:       req.UserID,
		PermissionID: req.PermissionID,
		AssignedBy:   req.AssignedBy,
		CreatedAt:    time.Now(),
	}

	// Save user permission to database
	if err := s.db.Table("app_user_permissions").Create(userPermission).Error; err != nil {
		logger.Error("failed to assign permission to user", "error", err)
		return errors.New("failed to assign permission to user")
	}

	return nil
}

// RevokePermissionFromUser revokes a permission from a user
func (s *AppPermissionService) RevokePermissionFromUser(ctx context.Context, req *RevokePermissionFromUserRequest) error {
	// Find user permission
	var userPermission domain.AppUserPermission
	if err := s.db.Table("app_user_permissions").Where("app_id = ? AND user_id = ? AND permission_id = ?", req.AppID, req.UserID, req.PermissionID).First(&userPermission).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			// User doesn't have this permission
			return nil
		}
		logger.Error("failed to find user permission", "error", err)
		return errors.New("failed to revoke permission from user")
	}

	// Delete user permission from database
	if err := s.db.Table("app_user_permissions").Delete(&userPermission).Error; err != nil {
		logger.Error("failed to revoke permission from user", "error", err)
		return errors.New("failed to revoke permission from user")
	}

	return nil
}

// ListUserPermissions lists all permissions assigned to a user for an application
func (s *AppPermissionService) ListUserPermissions(ctx context.Context, appID, userID string) ([]*domain.AppPermission, error) {
	// Get all permission IDs assigned to this user for this app
	var userPermissions []domain.AppUserPermission
	if err := s.db.Table("app_user_permissions").Where("app_id = ? AND user_id = ?", appID, userID).Find(&userPermissions).Error; err != nil {
		logger.Error("failed to list user permissions", "error", err)
		return nil, errors.New("failed to list user permissions")
	}

	// If user has no permissions, return empty list
	if len(userPermissions) == 0 {
		return []*domain.AppPermission{}, nil
	}

	// Extract permission IDs
	permissionIDs := make([]string, len(userPermissions))
	for i, up := range userPermissions {
		permissionIDs[i] = up.PermissionID
	}

	// Get all permissions
	var permissions []*domain.AppPermission
	if err := s.db.Table("app_permissions").Where("id IN ?", permissionIDs).Find(&permissions).Error; err != nil {
		logger.Error("failed to get permissions", "error", err)
		return nil, errors.New("failed to list user permissions")
	}

	return permissions, nil
}

// CheckUserPermission checks if a user has a specific permission for an application
func (s *AppPermissionService) CheckUserPermission(ctx context.Context, appID, userID, permission string) (bool, error) {
	// Validate permission
	validPermissions := map[string]bool{
		"read":   true,
		"write":  true,
		"delete": true,
		"manage": true,
	}

	if !validPermissions[permission] {
		return false, errors.New("invalid permission")
	}

	// First check if user has any permissions for this app
	var count int64
	if err := s.db.Table("app_user_permissions").Where("app_id = ? AND user_id = ?", appID, userID).Count(&count).Error; err != nil {
		logger.Error("failed to check user permissions", "error", err)
		return false, errors.New("failed to check user permission")
	}

	// If user has no permissions for this app, return false
	if count == 0 {
		return false, nil
	}

	// Get the permission ID for the specified permission name
	var appPermission domain.AppPermission
	if err := s.db.Table("app_permissions").Where("app_id = ? AND permission = ?", appID, permission).First(&appPermission).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			// Permission doesn't exist for this app
			return false, nil
		}
		logger.Error("failed to find application permission", "error", err)
		return false, errors.New("failed to check user permission")
	}

	// Check if user has this specific permission
	if err := s.db.Table("app_user_permissions").Where("app_id = ? AND user_id = ? AND permission_id = ?", appID, userID, appPermission.ID).First(&domain.AppUserPermission{}).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return false, nil
		}
		logger.Error("failed to check user permission", "error", err)
		return false, errors.New("failed to check user permission")
	}

	return true, nil
}

