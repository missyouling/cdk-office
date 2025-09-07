package service

import (
	"context"
	"errors"
	"time"

	"cdk-office/internal/auth/domain"
	"cdk-office/internal/shared/database"
	"cdk-office/pkg/logger"
	"gorm.io/gorm"
)

// PermissionServiceInterface defines the interface for permission service
type PermissionServiceInterface interface {
	CreatePermission(ctx context.Context, req *CreatePermissionRequest) (*domain.Permission, error)
	GetPermissionByName(ctx context.Context, name string) (*domain.Permission, error)
	CreateRole(ctx context.Context, name, description string) (*domain.Role, error)
	GetRoleByName(ctx context.Context, name string) (*domain.Role, error)
	AssignPermissionToRole(ctx context.Context, roleID, permissionID string) error
	CheckPermission(ctx context.Context, userID, resource, action string) (bool, error)
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

// CreatePermissionRequest represents the request for creating a permission
type CreatePermissionRequest struct {
	Name        string `json:"name" binding:"required"`
	Resource    string `json:"resource" binding:"required"`
	Action      string `json:"action" binding:"required"`
	Description string `json:"description"`
}

// CreatePermission creates a new permission
func (s *PermissionService) CreatePermission(ctx context.Context, req *CreatePermissionRequest) (*domain.Permission, error) {
	permission := &domain.Permission{
		ID:          generateID(),
		Name:        req.Name,
		Resource:    req.Resource,
		Action:      req.Action,
		Description: req.Description,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	if err := s.db.Create(permission).Error; err != nil {
		logger.Error("failed to create permission", "error", err)
		return nil, errors.New("failed to create permission")
	}

	return permission, nil
}

// GetPermissionByName retrieves a permission by name
func (s *PermissionService) GetPermissionByName(ctx context.Context, name string) (*domain.Permission, error) {
	var permission domain.Permission
	if err := s.db.Where("name = ?", name).First(&permission).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("permission not found")
		}
		logger.Error("failed to find permission", "error", err)
		return nil, errors.New("failed to get permission")
	}

	return &permission, nil
}

// CreateRole creates a new role
func (s *PermissionService) CreateRole(ctx context.Context, name, description string) (*domain.Role, error) {
	role := &domain.Role{
		ID:          generateID(),
		Name:        name,
		Description: description,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	if err := s.db.Create(role).Error; err != nil {
		logger.Error("failed to create role", "error", err)
		return nil, errors.New("failed to create role")
	}

	return role, nil
}

// GetRoleByName retrieves a role by name
func (s *PermissionService) GetRoleByName(ctx context.Context, name string) (*domain.Role, error) {
	var role domain.Role
	if err := s.db.Where("name = ?", name).First(&role).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("role not found")
		}
		logger.Error("failed to find role", "error", err)
		return nil, errors.New("failed to get role")
	}

	return &role, nil
}

// AssignPermissionToRole assigns a permission to a role
func (s *PermissionService) AssignPermissionToRole(ctx context.Context, roleID, permissionID string) error {
	rolePermission := &domain.RolePermission{
		ID:           generateID(),
		RoleID:       roleID,
		PermissionID: permissionID,
		CreatedAt:    time.Now(),
	}

	if err := s.db.Create(rolePermission).Error; err != nil {
		logger.Error("failed to assign permission to role", "error", err)
		return errors.New("failed to assign permission to role")
	}

	return nil
}

// CheckPermission checks if a user has a specific permission
func (s *PermissionService) CheckPermission(ctx context.Context, userID, resource, action string) (bool, error) {
	// 1. Get the user's roles
	var userRoles []domain.UserRole
	if err := s.db.Where("user_id = ?", userID).Find(&userRoles).Error; err != nil {
		logger.Error("failed to find user roles", "error", err)
		return false, errors.New("failed to check permission")
	}

	// If user has no roles, they have no permissions
	if len(userRoles) == 0 {
		return false, nil
	}

	// Extract role names
	roleNames := make([]string, len(userRoles))
	for i, userRole := range userRoles {
		roleNames[i] = userRole.Role
	}

	// 2. Get the roles by names
	var roles []domain.Role
	if err := s.db.Where("name IN ?", roleNames).Find(&roles).Error; err != nil {
		logger.Error("failed to find roles", "error", err)
		return false, errors.New("failed to check permission")
	}

	// Extract role IDs
	roleIDs := make([]string, len(roles))
	for i, role := range roles {
		roleIDs[i] = role.ID
	}

	// 3. Get role permissions
	var rolePermissions []domain.RolePermission
	if err := s.db.Where("role_id IN ?", roleIDs).Find(&rolePermissions).Error; err != nil {
		logger.Error("failed to find role permissions", "error", err)
		return false, errors.New("failed to check permission")
	}

	// If user's roles have no permissions, they have no permissions
	if len(rolePermissions) == 0 {
		return false, nil
	}

	// Extract permission IDs
	permissionIDs := make([]string, len(rolePermissions))
	for i, rp := range rolePermissions {
		permissionIDs[i] = rp.PermissionID
	}

	// 4. Get permissions by IDs
	var permissions []domain.Permission
	if err := s.db.Where("id IN ? AND resource = ? AND action = ?", permissionIDs, resource, action).Find(&permissions).Error; err != nil {
		logger.Error("failed to find permissions", "error", err)
		return false, errors.New("failed to check permission")
	}

	// If we found any permissions matching the resource and action, the user has permission
	return len(permissions) > 0, nil
}