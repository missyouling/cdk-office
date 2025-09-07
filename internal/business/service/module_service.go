package service

import (
	"context"
	"encoding/json"
	"errors"
	"time"

	"cdk-office/internal/business/domain"
	"cdk-office/internal/shared/database"
	"cdk-office/internal/shared/utils"
	"cdk-office/pkg/logger"
	"gorm.io/gorm"
)

// ModuleServiceInterface defines the interface for business module service
type ModuleServiceInterface interface {
	CreateModule(ctx context.Context, req *CreateModuleRequest) (*domain.BusinessModule, error)
	UpdateModule(ctx context.Context, moduleID string, req *UpdateModuleRequest) error
	DeleteModule(ctx context.Context, moduleID string) error
	ListModules(ctx context.Context, isActive *bool) ([]*domain.BusinessModule, error)
	GetModule(ctx context.Context, moduleID string) (*domain.BusinessModule, error)
	ActivateModule(ctx context.Context, moduleID string) error
	DeactivateModule(ctx context.Context, moduleID string) error
}

// ModuleService implements the ModuleServiceInterface
type ModuleService struct {
	db *gorm.DB
}

// NewModuleService creates a new instance of ModuleService
func NewModuleService() *ModuleService {
	return &ModuleService{
		db: database.GetDB(),
	}
}

// CreateModuleRequest represents the request for creating a business module
type CreateModuleRequest struct {
	Name        string `json:"name" binding:"required"`
	Description string `json:"description"`
	Version     string `json:"version"`
	Config      string `json:"config"`
}

// UpdateModuleRequest represents the request for updating a business module
type UpdateModuleRequest struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	Version     string `json:"version"`
	Config      string `json:"config"`
}

// CreateModule creates a new business module
func (s *ModuleService) CreateModule(ctx context.Context, req *CreateModuleRequest) (*domain.BusinessModule, error) {
	// Check if module name already exists
	var existingModule domain.BusinessModule
	if err := s.db.Where("name = ?", req.Name).First(&existingModule).Error; err == nil {
		return nil, errors.New("module name already exists")
	}

	// Create new module
	module := &domain.BusinessModule{
		ID:          utils.GenerateModuleID(),
		Name:        req.Name,
		Description: req.Description,
		Version:     "1.0.0",
		IsActive:    true,
		Config:      "{}",
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	// Save module to database
	if err := s.db.Create(module).Error; err != nil {
		logger.Error("failed to create business module", "error", err)
		return nil, errors.New("failed to create business module")
	}

	return module, nil
}

// UpdateModule updates an existing business module
func (s *ModuleService) UpdateModule(ctx context.Context, moduleID string, req *UpdateModuleRequest) error {
	// Find module by ID
	var module domain.BusinessModule
	if err := s.db.Where("id = ?", moduleID).First(&module).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errors.New("module not found")
		}
		logger.Error("failed to find business module", "error", err)
		return errors.New("failed to update business module")
	}

	// Update module fields
	if req.Name != "" {
		// Check if new name already exists
		var existingModule domain.BusinessModule
		if err := s.db.Where("name = ? AND id != ?", req.Name, moduleID).First(&existingModule).Error; err == nil {
			return errors.New("module name already exists")
		}
		module.Name = req.Name
	}
	
	if req.Description != "" {
		module.Description = req.Description
	}
	
	if req.Version != "" {
		module.Version = req.Version
	}
	
	if req.Config != "" {
		module.Config = req.Config
	}
	
	module.UpdatedAt = time.Now()

	// Save updated module to database
	if err := s.db.Save(&module).Error; err != nil {
		logger.Error("failed to update business module", "error", err)
		return errors.New("failed to update business module")
	}

	return nil
}

// DeleteModule deletes a business module
func (s *ModuleService) DeleteModule(ctx context.Context, moduleID string) error {
	// Find module by ID
	var module domain.BusinessModule
	if err := s.db.Where("id = ?", moduleID).First(&module).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errors.New("module not found")
		}
		logger.Error("failed to find business module", "error", err)
		return errors.New("failed to delete business module")
	}

	// Delete module from database
	if err := s.db.Delete(&module).Error; err != nil {
		logger.Error("failed to delete business module", "error", err)
		return errors.New("failed to delete business module")
	}

	// Delete associated permissions
	if err := s.db.Where("module_id = ?", moduleID).Delete(&domain.BusinessModulePermission{}).Error; err != nil {
		logger.Error("failed to delete module permissions", "error", err)
		// Don't return error here as the module was successfully deleted
	}

	return nil
}

// ListModules lists business modules
func (s *ModuleService) ListModules(ctx context.Context, isActive *bool) ([]*domain.BusinessModule, error) {
	var modules []*domain.BusinessModule

	// Build query
	query := s.db.Model(&domain.BusinessModule{})

	// Add filter for active status if provided
	if isActive != nil {
		query = query.Where("is_active = ?", *isActive)
	}

	// Execute query
	if err := query.Order("created_at desc").Find(&modules).Error; err != nil {
		logger.Error("failed to list business modules", "error", err)
		return nil, errors.New("failed to list business modules")
	}

	return modules, nil
}

// GetModule retrieves a business module by ID
func (s *ModuleService) GetModule(ctx context.Context, moduleID string) (*domain.BusinessModule, error) {
	var module domain.BusinessModule
	if err := s.db.Where("id = ?", moduleID).First(&module).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("module not found")
		}
		logger.Error("failed to find business module", "error", err)
		return nil, errors.New("failed to get business module")
	}

	return &module, nil
}

// ActivateModule activates a business module
func (s *ModuleService) ActivateModule(ctx context.Context, moduleID string) error {
	// Find module by ID
	var module domain.BusinessModule
	if err := s.db.Where("id = ?", moduleID).First(&module).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errors.New("module not found")
		}
		logger.Error("failed to find business module", "error", err)
		return errors.New("failed to activate business module")
	}

	// Update module status
	module.IsActive = true
	module.UpdatedAt = time.Now()

	// Save updated module to database
	if err := s.db.Save(&module).Error; err != nil {
		logger.Error("failed to activate business module", "error", err)
		return errors.New("failed to activate business module")
	}

	return nil
}

// DeactivateModule deactivates a business module
func (s *ModuleService) DeactivateModule(ctx context.Context, moduleID string) error {
	// Find module by ID
	var module domain.BusinessModule
	if err := s.db.Where("id = ?", moduleID).First(&module).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errors.New("module not found")
		}
		logger.Error("failed to find business module", "error", err)
		return errors.New("failed to deactivate business module")
	}

	// Update module status
	module.IsActive = false
	module.UpdatedAt = time.Now()

	// Save updated module to database
	if err := s.db.Save(&module).Error; err != nil {
		logger.Error("failed to deactivate business module", "error", err)
		return errors.New("failed to deactivate business module")
	}

	return nil
}

// convertStringToJSON converts a string to JSON string
func convertStringToJSON(str string) string {
	// Convert string to JSON
	jsonData, err := json.Marshal(str)
	if err != nil {
		logger.Error("failed to marshal string to JSON", "error", err)
		return "\"\""
	}
	return string(jsonData)
}