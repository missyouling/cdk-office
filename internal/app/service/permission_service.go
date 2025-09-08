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

// PermissionServiceInterface defines the interface for permission service
type PermissionServiceInterface interface {
	CreatePermission(ctx context.Context, req *CreatePermissionRequest) (*domain.AppPermission, error)
	UpdatePermission(ctx context.Context, permissionID string, req *UpdatePermissionRequest) error
	DeletePermission(ctx context.Context, permissionID string) error
	ListPermissions(ctx context.Context, appID string, page, size int) ([]*domain.AppPermission, int64, error)
	GetPermission(ctx context.Context, permissionID string) (*domain.AppPermission, error)
	CheckPermission(ctx context.Context, appID, userID, action string) (bool, error)
	AssignPermission(ctx context.Context, req *AssignPermissionRequest) error
	RevokePermission(ctx context.Context, req *RevokePermissionRequest) error
	ListUserPermissions(ctx context.Context, appID, userID string) ([]*domain.AppPermission, error)
}

// PermissionService implements the PermissionServiceInterface
type PermissionService struct {
	db *gorm.DB
}

// NewPermissionService creates a new instance of PermissionService
func NewPermissionService() *PermissionService {
	return &PermissionService{
		db: database.GetDB(),
	}
}

// NewPermissionServiceWithDB creates a new instance of PermissionService with a specific database connection
func NewPermissionServiceWithDB(db *gorm.DB) *PermissionService {
	return &PermissionService{
		db: db,
	}
}

// CreatePermissionRequest represents the request for creating a permission
type CreatePermissionRequest struct {
	AppID       string `json:"app_id" binding:"required"`
	Name        string `json:"name" binding:"required"`
	Description string `json:"description"`
	Action      string `json:"action" binding:"required"`
	CreatedBy   string `json:"created_by" binding:"required"`
}

// UpdatePermissionRequest represents the request for updating a permission
type UpdatePermissionRequest struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	Action      string `json:"action"`
}

// AssignPermissionRequest represents the request for assigning a permission to a user
type AssignPermissionRequest struct {
	AppID        string `json:"app_id" binding:"required"`
	UserID       string `json:"user_id" binding:"required"`
	PermissionID string `json:"permission_id" binding:"required"`
	AssignedBy   string `json:"assigned_by" binding:"required"`
}

// RevokePermissionRequest represents the request for revoking a permission from a user
type RevokePermissionRequest struct {
	AppID        string `json:"app_id" binding:"required"`
	UserID       string `json:"user_id" binding:"required"`
	PermissionID string `json:"permission_id" binding:"required"`
}



// CreatePermission creates a new permission
func (s *PermissionService) CreatePermission(ctx context.Context, req *CreatePermissionRequest) (*domain.AppPermission, error) {
	// Validate action
	validActions := map[string]bool{
		"read":   true,
		"write":  true,
		"delete": true,
		"admin":  true,
	}

	if !validActions[req.Action] {
		return nil, errors.New("invalid permission action")
	}

	// Check if permission with the same name already exists in the app
	var existingPermission domain.AppPermission
	if err := s.db.Table("app_permissions").Where("app_id = ? AND name = ?", req.AppID, req.Name).First(&existingPermission).Error; err == nil {
		return nil, errors.New("permission with this name already exists in the application")
	}

	// Create new permission
	permission := &domain.AppPermission{
		ID:          utils.GeneratePermissionID(),
		AppID:       req.AppID,
		Name:        req.Name,
		Description: req.Description,
		Permission:  req.Action,
		CreatedBy:   req.CreatedBy,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	// Save permission to database
	if err := s.db.Table("app_permissions").Create(permission).Error; err != nil {
		logger.Error("failed to create permission", "error", err)
		return nil, errors.New("failed to create permission")
	}

	return permission, nil
}

// UpdatePermission updates an existing permission
func (s *PermissionService) UpdatePermission(ctx context.Context, permissionID string, req *UpdatePermissionRequest) error {
	// Find permission by ID
	var permission domain.AppPermission
	if err := s.db.Table("app_permissions").Where("id = ?", permissionID).First(&permission).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errors.New("permission not found")
		}
		logger.Error("failed to find permission", "error", err)
		return errors.New("failed to update permission")
	}

	// Update permission fields
	if req.Name != "" {
		permission.Name = req.Name
	}

	if req.Description != "" {
		permission.Description = req.Description
	}

	if req.Action != "" {
		// Validate permission
		validPermissions := map[string]bool{
			"read":   true,
			"write":  true,
			"delete": true,
			"admin":  true,
		}

		if !validPermissions[req.Action] {
			return errors.New("invalid permission")
		}
		permission.Permission = req.Action
	}

	permission.UpdatedAt = time.Now()

	// Save updated permission to database
	if err := s.db.Table("app_permissions").Save(&permission).Error; err != nil {
		logger.Error("failed to update permission", "error", err)
		return errors.New("failed to update permission")
	}

	return nil
}

// DeletePermission deletes a permission
func (s *PermissionService) DeletePermission(ctx context.Context, permissionID string) error {
	// Find permission by ID
	var permission domain.AppPermission
	if err := s.db.Table("app_permissions").Where("id = ?", permissionID).First(&permission).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errors.New("permission not found")
		}
		logger.Error("failed to find permission", "error", err)
		return errors.New("failed to delete permission")
	}

	// Delete permission from database
	if err := s.db.Table("app_permissions").Delete(&permission).Error; err != nil {
		logger.Error("failed to delete permission", "error", err)
		return errors.New("failed to delete permission")
	}

	// Also delete all user permissions associated with this permission
	if err := s.db.Table("app_user_permissions").Where("permission_id = ?", permissionID).Delete(&domain.AppUserPermission{}).Error; err != nil {
		logger.Error("failed to delete user permissions", "error", err)
		// Don't return error here as the main permission was deleted successfully
	}

	return nil
}

// ListPermissions lists permissions with pagination
func (s *PermissionService) ListPermissions(ctx context.Context, appID string, page, size int) ([]*domain.AppPermission, int64, error) {
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
		logger.Error("failed to count permissions", "error", err)
		return nil, 0, errors.New("failed to list permissions")
	}

	// Apply pagination
	offset := (page - 1) * size
	dbQuery = dbQuery.Offset(offset).Limit(size).Order("created_at desc")

	// Execute query
	var permissions []*domain.AppPermission
	if err := dbQuery.Find(&permissions).Error; err != nil {
		logger.Error("failed to list permissions", "error", err)
		return nil, 0, errors.New("failed to list permissions")
	}

	return permissions, total, nil
}

// GetPermission retrieves a permission by ID
func (s *PermissionService) GetPermission(ctx context.Context, permissionID string) (*domain.AppPermission, error) {
	var permission domain.AppPermission
	if err := s.db.Table("app_permissions").Where("id = ?", permissionID).First(&permission).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("permission not found")
		}
		logger.Error("failed to find permission", "error", err)
		return nil, errors.New("failed to get permission")
	}

	return &permission, nil
}

// CheckPermission checks if a user has a specific permission for an application
func (s *PermissionService) CheckPermission(ctx context.Context, appID, userID, permission string) (bool, error) {
	// First check if user has any permissions for this app
	var count int64
	if err := s.db.Table("app_user_permissions").Where("app_id = ? AND user_id = ?", appID, userID).Count(&count).Error; err != nil {
		logger.Error("failed to check user permissions", "error", err)
		return false, errors.New("failed to check permission")
	}

	// If user has no permissions for this app, return false
	if count == 0 {
		return false, nil
	}

	// Check if user has the specific permission
	var perm domain.AppPermission
	if err := s.db.Table("app_permissions").Where("app_id = ? AND permission = ?", appID, permission).First(&perm).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			// Permission doesn't exist, return false
			return false, nil
		}
		logger.Error("failed to find permission", "error", err)
		return false, errors.New("failed to check permission")
	}

	// Check if user has this specific permission
	if err := s.db.Table("app_user_permissions").Where("app_id = ? AND user_id = ? AND permission_id = ?", appID, userID, perm.ID).First(&domain.AppUserPermission{}).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return false, nil
		}
		logger.Error("failed to check user permission", "error", err)
		return false, errors.New("failed to check permission")
	}

	return true, nil
}

// AssignPermission assigns a permission to a user
func (s *PermissionService) AssignPermission(ctx context.Context, req *AssignPermissionRequest) error {
	// Check if permission exists
	var permission domain.AppPermission
	if err := s.db.Table("app_permissions").Where("id = ?", req.PermissionID).First(&permission).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errors.New("permission not found")
		}
		logger.Error("failed to find permission", "error", err)
		return errors.New("failed to assign permission")
	}

	// Check if permission is for the correct app
	if permission.AppID != req.AppID {
		return errors.New("permission does not belong to this application")
	}

	// Check if user already has this permission
	var existing domain.AppUserPermission
	if err := s.db.Table("app_user_permissions").Where("app_id = ? AND user_id = ? AND permission_id = ?", req.AppID, req.UserID, req.PermissionID).First(&existing).Error; err == nil {
		// User already has this permission
		return nil
	} else if !errors.Is(err, gorm.ErrRecordNotFound) {
		logger.Error("failed to check existing user permission", "error", err)
		return errors.New("failed to assign permission")
	}

	// Create new user permission
	userPermission := &domain.AppUserPermission{
		ID:           generateUserPermissionID(),
		AppID:        req.AppID,
		UserID:       req.UserID,
		PermissionID: req.PermissionID,
		AssignedBy:   req.AssignedBy,
		CreatedAt:    time.Now(),
	}

	// Save user permission to database
	if err := s.db.Table("app_user_permissions").Create(userPermission).Error; err != nil {
		logger.Error("failed to assign permission", "error", err)
		return errors.New("failed to assign permission")
	}

	return nil
}

// RevokePermission revokes a permission from a user
func (s *PermissionService) RevokePermission(ctx context.Context, req *RevokePermissionRequest) error {
	// Check if user has this permission
	var userPermission domain.AppUserPermission
	if err := s.db.Table("app_user_permissions").Where("app_id = ? AND user_id = ? AND permission_id = ?", req.AppID, req.UserID, req.PermissionID).First(&userPermission).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			// User doesn't have this permission, nothing to revoke
			return nil
		}
		logger.Error("failed to find user permission", "error", err)
		return errors.New("failed to revoke permission")
	}

	// Delete user permission from database
	if err := s.db.Table("app_user_permissions").Delete(&userPermission).Error; err != nil {
		logger.Error("failed to revoke permission", "error", err)
		return errors.New("failed to revoke permission")
	}

	return nil
}

// ListUserPermissions lists all permissions assigned to a user for an application
func (s *PermissionService) ListUserPermissions(ctx context.Context, appID, userID string) ([]*domain.AppPermission, error) {
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

// generatePermissionID generates a unique permission ID
func generatePermissionID() string {
	// In a real application, use a proper ID generation library like uuid
	return "perm_" + time.Now().Format("20060102150405")
}

// generateUserPermissionID generates a unique user permission ID
func generateUserPermissionID() string {
	// In a real application, use a proper ID generation library like uuid
	return "user_perm_" + time.Now().Format("20060102150405")
}