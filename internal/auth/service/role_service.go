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

// RoleServiceInterface defines the interface for role service
type RoleServiceInterface interface {
	AssignRoleToUser(ctx context.Context, userID, roleID string) error
	GetUserRoles(ctx context.Context, userID string) ([]*domain.Role, error)
}

// RoleService implements the RoleServiceInterface
type RoleService struct {
	db *gorm.DB
}

// NewRoleService creates a new instance of RoleService
func NewRoleService() *RoleService {
	return &RoleService{
		db: database.GetDB(),
	}
}

// AssignRoleToUser assigns a role to a user
func (s *RoleService) AssignRoleToUser(ctx context.Context, userID, roleID string) error {
	// Check if the role exists
	var role domain.Role
	if err := s.db.Where("id = ?", roleID).First(&role).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errors.New("role not found")
		}
		logger.Error("failed to find role", "error", err)
		return errors.New("failed to assign role to user")
	}

	// Check if the user exists
	var user domain.User
	if err := s.db.Where("id = ?", userID).First(&user).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errors.New("user not found")
		}
		logger.Error("failed to find user", "error", err)
		return errors.New("failed to assign role to user")
	}

	// Create user role relationship
	userRole := &domain.UserRole{
		ID:        generateID(),
		UserID:    userID,
		Role:      role.Name,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	if err := s.db.Create(userRole).Error; err != nil {
		logger.Error("failed to assign role to user", "error", err)
		return errors.New("failed to assign role to user")
	}

	return nil
}

// GetUserRoles retrieves all roles for a user
func (s *RoleService) GetUserRoles(ctx context.Context, userID string) ([]*domain.Role, error) {
	// Get user role relationships
	var userRoles []domain.UserRole
	if err := s.db.Where("user_id = ?", userID).Find(&userRoles).Error; err != nil {
		logger.Error("failed to find user roles", "error", err)
		return nil, errors.New("failed to get user roles")
	}

	// Get role details
	var roles []*domain.Role
	for _, userRole := range userRoles {
		var role domain.Role
		if err := s.db.Where("name = ?", userRole.Role).First(&role).Error; err != nil {
			if !errors.Is(err, gorm.ErrRecordNotFound) {
				logger.Error("failed to find role", "error", err)
				return nil, errors.New("failed to get user roles")
			}
			// Skip roles that don't exist
			continue
		}
		roles = append(roles, &role)
	}

	return roles, nil
}