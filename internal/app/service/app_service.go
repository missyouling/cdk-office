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

// AppServiceInterface defines the interface for application service
type AppServiceInterface interface {
	CreateApplication(ctx context.Context, req *CreateApplicationRequest) (*domain.Application, error)
	UpdateApplication(ctx context.Context, appID string, req *UpdateApplicationRequest) error
	DeleteApplication(ctx context.Context, appID string) error
	ListApplications(ctx context.Context, teamID string, page, size int) ([]*domain.Application, int64, error)
	GetApplication(ctx context.Context, appID string) (*domain.Application, error)
}

// AppService implements the AppServiceInterface
type AppService struct {
	db *gorm.DB
}

// NewAppService creates a new instance of AppService
func NewAppService() *AppService {
	return &AppService{
		db: database.GetDB(),
	}
}

// NewAppServiceWithDB creates a new instance of AppService with a specific database connection
func NewAppServiceWithDB(db *gorm.DB) *AppService {
	return &AppService{
		db: db,
	}
}

// CreateApplicationRequest represents the request for creating an application
type CreateApplicationRequest struct {
	TeamID      string `json:"team_id" binding:"required"`
	Name        string `json:"name" binding:"required"`
	Description string `json:"description"`
	Type        string `json:"type" binding:"required"`
	Config      string `json:"config"`
	CreatedBy   string `json:"created_by" binding:"required"`
}

// UpdateApplicationRequest represents the request for updating an application
type UpdateApplicationRequest struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	Config      string `json:"config"`
	IsActive    *bool  `json:"is_active"`
}

// Application represents the application entity
type Application struct {
	ID          string    `json:"id"`
	TeamID      string    `json:"team_id"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	Type        string    `json:"type"`
	Config      string    `json:"config"`
	IsActive    bool      `json:"is_active"`
	CreatedBy   string    `json:"created_by"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// CreateApplication creates a new application
func (s *AppService) CreateApplication(ctx context.Context, req *CreateApplicationRequest) (*domain.Application, error) {
	// Validate application type
	validTypes := map[string]bool{
		"qrcode": true,
		"form":   true,
		"survey": true,
	}

	if !validTypes[req.Type] {
		return nil, errors.New("invalid application type")
	}

	// Check if application with the same name already exists in the team
	var existingApp domain.Application
	if err := s.db.Where("team_id = ? AND name = ?", req.TeamID, req.Name).First(&existingApp).Error; err == nil {
		return nil, errors.New("application with this name already exists in the team")
	} else if !errors.Is(err, gorm.ErrRecordNotFound) {
		logger.Error("failed to check existing application", "error", err)
		return nil, errors.New("failed to create application")
	}

	// Create new application
	app := &domain.Application{
		ID:          utils.GenerateAppID(),
		TeamID:      req.TeamID,
		Name:        req.Name,
		Description: req.Description,
		Type:        req.Type,
		Config:      req.Config,
		IsActive:    true,
		CreatedBy:   req.CreatedBy,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	// Save application to database
	if err := s.db.Create(app).Error; err != nil {
		logger.Error("failed to create application", "error", err)
		return nil, errors.New("failed to create application")
	}

	return app, nil
}

// UpdateApplication updates an existing application
func (s *AppService) UpdateApplication(ctx context.Context, appID string, req *UpdateApplicationRequest) error {
	// Find application by ID
	var app domain.Application
	if err := s.db.Where("id = ?", appID).First(&app).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errors.New("application not found")
		}
		logger.Error("failed to find application", "error", err)
		return errors.New("failed to update application")
	}

	// Update application fields
	if req.Name != "" {
		app.Name = req.Name
	}

	if req.Description != "" {
		app.Description = req.Description
	}

	if req.Config != "" {
		app.Config = req.Config
	}

	if req.IsActive != nil {
		app.IsActive = *req.IsActive
	}

	app.UpdatedAt = time.Now()

	// Save updated application to database
	if err := s.db.Save(&app).Error; err != nil {
		logger.Error("failed to update application", "error", err)
		return errors.New("failed to update application")
	}

	return nil
}

// DeleteApplication deletes an application
func (s *AppService) DeleteApplication(ctx context.Context, appID string) error {
	// Find application by ID
	var app domain.Application
	if err := s.db.Where("id = ?", appID).First(&app).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errors.New("application not found")
		}
		logger.Error("failed to find application", "error", err)
		return errors.New("failed to delete application")
	}

	// Delete application from database
	if err := s.db.Delete(&app).Error; err != nil {
		logger.Error("failed to delete application", "error", err)
		return errors.New("failed to delete application")
	}

	return nil
}

// ListApplications lists applications with pagination
func (s *AppService) ListApplications(ctx context.Context, teamID string, page, size int) ([]*domain.Application, int64, error) {
	// Validate pagination parameters
	if page < 1 {
		page = 1
	}
	if size < 1 || size > 100 {
		size = 10
	}

	// Build query
	dbQuery := s.db.Model(&domain.Application{}).Where("team_id = ?", teamID)

	// Count total results
	var total int64
	if err := dbQuery.Count(&total).Error; err != nil {
		logger.Error("failed to count applications", "error", err)
		return nil, 0, errors.New("failed to list applications")
	}

	// Apply pagination
	offset := (page - 1) * size
	dbQuery = dbQuery.Offset(offset).Limit(size).Order("created_at desc")

	// Execute query
	var apps []*domain.Application
	if err := dbQuery.Find(&apps).Error; err != nil {
		logger.Error("failed to list applications", "error", err)
		return nil, 0, errors.New("failed to list applications")
	}

	return apps, total, nil
}

// GetApplication retrieves an application by ID
func (s *AppService) GetApplication(ctx context.Context, appID string) (*domain.Application, error) {
	var app domain.Application
	if err := s.db.Where("id = ?", appID).First(&app).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("application not found")
		}
		logger.Error("failed to find application", "error", err)
		return nil, errors.New("failed to get application")
	}

	return &app, nil
}