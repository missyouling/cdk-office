package service

import (
	"context"
	"errors"
	"time"

	"cdk-office/internal/business/domain"
	"cdk-office/internal/shared/database"
	"cdk-office/internal/shared/utils"
	"cdk-office/pkg/logger"
	"gorm.io/gorm"
)

// BusinessPermissionServiceInterface defines the interface for business module permission service
type BusinessPermissionServiceInterface interface {
	AssignPermissionToRole(ctx context.Context, req *AssignPermissionToRoleRequest) error
	RevokePermissionFromRole(ctx context.Context, req *RevokePermissionFromRoleRequest) error
	ListRolePermissions(ctx context.Context, roleID string) ([]*domain.BusinessModulePermission, error)
	CheckRolePermission(ctx context.Context, roleID, moduleID, permission string) (bool, error)
	ListModulePermissions(ctx context.Context, moduleID string) ([]*domain.BusinessModulePermission, error)
}

// BusinessPermissionService implements the BusinessPermissionServiceInterface
type BusinessPermissionService struct {
	db *gorm.DB
}

// NewBusinessPermissionService creates a new instance of BusinessPermissionService
func NewBusinessPermissionService() *BusinessPermissionService {
	return &BusinessPermissionService{
		db: database.GetDB(),
	}
}

// AssignPermissionToRoleRequest represents the request for assigning a permission to a role
type AssignPermissionToRoleRequest struct {
	ModuleID   string `json:"module_id" binding:"required"`
	RoleID     string `json:"role_id" binding:"required"`
	Permission string `json:"permission" binding:"required"`
}

// RevokePermissionFromRoleRequest represents the request for revoking a permission from a role
type RevokePermissionFromRoleRequest struct {
	ModuleID   string `json:"module_id" binding:"required"`
	RoleID     string `json:"role_id" binding:"required"`
	Permission string `json:"permission" binding:"required"`
}

// AssignPermissionToRole assigns a permission to a role for a specific module
func (s *BusinessPermissionService) AssignPermissionToRole(ctx context.Context, req *AssignPermissionToRoleRequest) error {
	// Check if the module exists
	var module domain.BusinessModule
	if err := s.db.Where("id = ?", req.ModuleID).First(&module).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errors.New("module not found")
		}
		logger.Error("failed to find module", "error", err)
		return errors.New("failed to assign permission to role")
	}

	// Check if the permission already exists for this role and module
	var existingPermission domain.BusinessModulePermission
	if err := s.db.Where("module_id = ? AND role_id = ? AND permission = ?", req.ModuleID, req.RoleID, req.Permission).First(&existingPermission).Error; err == nil {
		// Permission already exists, nothing to do
		return nil
	}

	// Create new permission
	permission := &domain.BusinessModulePermission{
		ID:         utils.GenerateBusinessPermissionID(),
		ModuleID:   req.ModuleID,
		RoleID:     req.RoleID,
		Permission: req.Permission,
		CreatedAt:  time.Now(),
	}

	// Save permission to database
	if err := s.db.Create(permission).Error; err != nil {
		logger.Error("failed to assign permission to role", "error", err)
		return errors.New("failed to assign permission to role")
	}

	return nil
}

// RevokePermissionFromRole revokes a permission from a role for a specific module
func (s *BusinessPermissionService) RevokePermissionFromRole(ctx context.Context, req *RevokePermissionFromRoleRequest) error {
	// Delete permission from database
	result := s.db.Where("module_id = ? AND role_id = ? AND permission = ?", req.ModuleID, req.RoleID, req.Permission).Delete(&domain.BusinessModulePermission{})

	if result.Error != nil {
		logger.Error("failed to revoke permission from role", "error", result.Error)
		return errors.New("failed to revoke permission from role")
	}

	if result.RowsAffected == 0 {
		return errors.New("permission not found for this role and module")
	}

	return nil
}

// ListRolePermissions lists all permissions for a role
func (s *BusinessPermissionService) ListRolePermissions(ctx context.Context, roleID string) ([]*domain.BusinessModulePermission, error) {
	var permissions []*domain.BusinessModulePermission

	// Build query
	query := s.db.Model(&domain.BusinessModulePermission{}).Where("role_id = ?", roleID)

	// Execute query
	if err := query.Find(&permissions).Error; err != nil {
		logger.Error("failed to list role permissions", "error", err)
		return nil, errors.New("failed to list role permissions")
	}

	return permissions, nil
}

// CheckRolePermission checks if a role has a specific permission for a module
func (s *BusinessPermissionService) CheckRolePermission(ctx context.Context, roleID, moduleID, permission string) (bool, error) {
	// Check if the permission exists for this role and module
	var existingPermission domain.BusinessModulePermission
	if err := s.db.Where("module_id = ? AND role_id = ? AND permission = ?", moduleID, roleID, permission).First(&existingPermission).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return false, nil
		}
		logger.Error("failed to check role permission", "error", err)
		return false, errors.New("failed to check role permission")
	}

	return true, nil
}

// ListModulePermissions lists all permissions for a module
func (s *BusinessPermissionService) ListModulePermissions(ctx context.Context, moduleID string) ([]*domain.BusinessModulePermission, error) {
	var permissions []*domain.BusinessModulePermission

	// Build query
	query := s.db.Model(&domain.BusinessModulePermission{}).Where("module_id = ?", moduleID)

	// Execute query
	if err := query.Find(&permissions).Error; err != nil {
		logger.Error("failed to list module permissions", "error", err)
		return nil, errors.New("failed to list module permissions")
	}

	return permissions, nil
}