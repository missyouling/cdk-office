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

// PluginServiceInterface defines the interface for plugin service
type PluginServiceInterface interface {
	RegisterPlugin(ctx context.Context, req *RegisterPluginRequest) error
	UnregisterPlugin(ctx context.Context, pluginID string) error
	ListPlugins(ctx context.Context, teamID string) ([]*Plugin, error)
	GetPlugin(ctx context.Context, pluginID string) (*Plugin, error)
	EnablePlugin(ctx context.Context, pluginID string) error
	DisablePlugin(ctx context.Context, pluginID string) error
}

// PluginService implements the PluginServiceInterface
type PluginService struct {
	db *gorm.DB
}

// NewPluginService creates a new instance of PluginService
func NewPluginService() *PluginService {
	return &PluginService{
		db: database.GetDB(),
	}
}

// Plugin represents a plugin in the system
type Plugin struct {
	ID          string    `json:"id" gorm:"primaryKey"`
	TeamID      string    `json:"team_id" gorm:"index"`
	Name        string    `json:"name" gorm:"size:100;uniqueIndex:idx_team_plugin"`
	Description string    `json:"description" gorm:"type:text"`
	Version     string    `json:"version" gorm:"size:20"`
	EntryPoint  string    `json:"entry_point" gorm:"size:255"`
	IsActive    bool      `json:"is_active"`
	Config      string    `json:"config" gorm:"type:jsonb"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// RegisterPluginRequest represents the request for registering a plugin
type RegisterPluginRequest struct {
	TeamID      string `json:"team_id" binding:"required"`
	Name        string `json:"name" binding:"required"`
	Description string `json:"description"`
	Version     string `json:"version"`
	EntryPoint  string `json:"entry_point" binding:"required"`
	Config      string `json:"config"`
}

// RegisterPlugin registers a new plugin
func (s *PluginService) RegisterPlugin(ctx context.Context, req *RegisterPluginRequest) error {
	// Check if plugin name already exists
	var existingPlugin domain.Plugin
	if err := s.db.Where("name = ?", req.Name).First(&existingPlugin).Error; err == nil {
		return errors.New("plugin name already exists")
	}

	// Create new plugin
	plugin := &domain.Plugin{
		ID:          utils.GeneratePluginID(),
		Name:        req.Name,
		Description: req.Description,
		Version:     "1.0.0",
		Status:      "active",
		CreatedBy:   req.Name, // Using Name as placeholder since CreatedBy is not in request
		TeamID:      req.Name, // Using Name as placeholder since TeamID is not in request
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	// Save plugin to database
	if err := s.db.Create(plugin).Error; err != nil {
		logger.Error("failed to register plugin", "error", err)
		return errors.New("failed to register plugin")
	}

	return nil
}

// UnregisterPlugin unregisters a plugin
func (s *PluginService) UnregisterPlugin(ctx context.Context, pluginID string) error {
	// Find plugin by ID
	var plugin domain.Plugin
	if err := s.db.Where("id = ?", pluginID).First(&plugin).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errors.New("plugin not found")
		}
		logger.Error("failed to find plugin", "error", err)
		return errors.New("failed to unregister plugin")
	}

	// Delete plugin from database
	if err := s.db.Delete(&plugin).Error; err != nil {
		logger.Error("failed to unregister plugin", "error", err)
		return errors.New("failed to unregister plugin")
	}

	return nil
}

// ListPlugins lists plugins for a team
func (s *PluginService) ListPlugins(ctx context.Context, teamID string) ([]*Plugin, error) {
	var plugins []*Plugin

	// Build query
	query := s.db.Model(&Plugin{}).Where("team_id = ?", teamID)

	// Execute query
	if err := query.Order("created_at desc").Find(&plugins).Error; err != nil {
		logger.Error("failed to list plugins", "error", err)
		return nil, errors.New("failed to list plugins")
	}

	return plugins, nil
}

// GetPlugin retrieves a plugin by ID
func (s *PluginService) GetPlugin(ctx context.Context, pluginID string) (*Plugin, error) {
	var plugin Plugin
	if err := s.db.Where("id = ?", pluginID).First(&plugin).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("plugin not found")
		}
		logger.Error("failed to find plugin", "error", err)
		return nil, errors.New("failed to get plugin")
	}

	return &plugin, nil
}

// EnablePlugin enables a plugin
func (s *PluginService) EnablePlugin(ctx context.Context, pluginID string) error {
	// Find plugin by ID
	var plugin Plugin
	if err := s.db.Where("id = ?", pluginID).First(&plugin).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errors.New("plugin not found")
		}
		logger.Error("failed to find plugin", "error", err)
		return errors.New("failed to enable plugin")
	}

	// Update plugin status
	plugin.IsActive = true
	plugin.UpdatedAt = time.Now()

	// Save updated plugin to database
	if err := s.db.Save(&plugin).Error; err != nil {
		logger.Error("failed to enable plugin", "error", err)
		return errors.New("failed to enable plugin")
	}

	return nil
}

// DisablePlugin disables a plugin
func (s *PluginService) DisablePlugin(ctx context.Context, pluginID string) error {
	// Find plugin by ID
	var plugin Plugin
	if err := s.db.Where("id = ?", pluginID).First(&plugin).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errors.New("plugin not found")
		}
		logger.Error("failed to find plugin", "error", err)
		return errors.New("failed to disable plugin")
	}

	// Update plugin status
	plugin.IsActive = false
	plugin.UpdatedAt = time.Now()

	// Save updated plugin to database
	if err := s.db.Save(&plugin).Error; err != nil {
		logger.Error("failed to disable plugin", "error", err)
		return errors.New("failed to disable plugin")
	}

	return nil
}