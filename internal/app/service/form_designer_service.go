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

// FormDesignerServiceInterface defines the interface for form designer service
type FormDesignerServiceInterface interface {
	CreateFormDesign(ctx context.Context, req *CreateFormDesignRequest) (*domain.FormDesign, error)
	UpdateFormDesign(ctx context.Context, formID string, req *UpdateFormDesignRequest) error
	DeleteFormDesign(ctx context.Context, formID string) error
	ListFormDesigns(ctx context.Context, appID string, page, size int) ([]*domain.FormDesign, int64, error)
	GetFormDesign(ctx context.Context, formID string) (*domain.FormDesign, error)
	PublishFormDesign(ctx context.Context, formID string) error
}

// FormDesignerService implements the FormDesignerServiceInterface
type FormDesignerService struct {
	db *gorm.DB
}

// NewFormDesignerService creates a new instance of FormDesignerService
func NewFormDesignerService() *FormDesignerService {
	return &FormDesignerService{
		db: database.GetDB(),
	}
}

// CreateFormDesignRequest represents the request for creating a form design
type CreateFormDesignRequest struct {
	AppID       string `json:"app_id" binding:"required"`
	Name        string `json:"name" binding:"required"`
	Description string `json:"description"`
	Schema      string `json:"schema" binding:"required"`
	Config      string `json:"config"`
	CreatedBy   string `json:"created_by" binding:"required"`
}

// UpdateFormDesignRequest represents the request for updating a form design
type UpdateFormDesignRequest struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	Schema      string `json:"schema"`
	Config      string `json:"config"`
	IsActive    *bool  `json:"is_active"`
}

// FormDesign represents the form design entity
type FormDesign struct {
	ID          string    `json:"id"`
	AppID       string    `json:"app_id"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	Schema      string    `json:"schema"`
	Config      string    `json:"config"`
	IsActive    bool      `json:"is_active"`
	IsPublished bool      `json:"is_published"`
	CreatedBy   string    `json:"created_by"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// CreateFormDesign creates a new form design
func (s *FormDesignerService) CreateFormDesign(ctx context.Context, req *CreateFormDesignRequest) (*domain.FormDesign, error) {
	// Create new form design
	form := &domain.FormDesign{
		ID:          utils.GenerateFormDesignID(),
		AppID:       req.AppID,
		Name:        req.Name,
		Description: req.Description,
		Schema:      req.Schema,
		Config:      req.Config,
		IsActive:    true,
		IsPublished: false,
		CreatedBy:   req.CreatedBy,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	// Save form design to database
	if err := s.db.Table("form_designs").Create(form).Error; err != nil {
		logger.Error("failed to create form design", "error", err)
		return nil, errors.New("failed to create form design")
	}

	return form, nil
}

// UpdateFormDesign updates an existing form design
func (s *FormDesignerService) UpdateFormDesign(ctx context.Context, formID string, req *UpdateFormDesignRequest) error {
	// Find form design by ID
	var form domain.FormDesign
	if err := s.db.Table("form_designs").Where("id = ?", formID).First(&form).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errors.New("form design not found")
		}
		logger.Error("failed to find form design", "error", err)
		return errors.New("failed to update form design")
	}

	// Check if form is published
	if form.IsPublished {
		return errors.New("cannot update published form design")
	}

	// Update form design fields
	if req.Name != "" {
		form.Name = req.Name
	}

	if req.Description != "" {
		form.Description = req.Description
	}

	if req.Schema != "" {
		form.Schema = req.Schema
	}

	if req.Config != "" {
		form.Config = req.Config
	}

	if req.IsActive != nil {
		form.IsActive = *req.IsActive
	}

	form.UpdatedAt = time.Now()

	// Save updated form design to database
	if err := s.db.Table("form_designs").Save(&form).Error; err != nil {
		logger.Error("failed to update form design", "error", err)
		return errors.New("failed to update form design")
	}

	return nil
}

// DeleteFormDesign deletes a form design
func (s *FormDesignerService) DeleteFormDesign(ctx context.Context, formID string) error {
	// Find form design by ID
	var form domain.FormDesign
	if err := s.db.Table("form_designs").Where("id = ?", formID).First(&form).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errors.New("form design not found")
		}
		logger.Error("failed to find form design", "error", err)
		return errors.New("failed to delete form design")
	}

	// Check if form is published
	if form.IsPublished {
		return errors.New("cannot delete published form design")
	}

	// Delete form design from database
	if err := s.db.Table("form_designs").Delete(&form).Error; err != nil {
		logger.Error("failed to delete form design", "error", err)
		return errors.New("failed to delete form design")
	}

	return nil
}

// ListFormDesigns lists form designs with pagination
func (s *FormDesignerService) ListFormDesigns(ctx context.Context, appID string, page, size int) ([]*domain.FormDesign, int64, error) {
	// Validate pagination parameters
	if page < 1 {
		page = 1
	}
	if size < 1 || size > 100 {
		size = 10
	}

	// Build query
	dbQuery := s.db.Table("form_designs").Where("app_id = ?", appID)

	// Count total results
	var total int64
	if err := dbQuery.Count(&total).Error; err != nil {
		logger.Error("failed to count form designs", "error", err)
		return nil, 0, errors.New("failed to list form designs")
	}

	// Apply pagination
	offset := (page - 1) * size
	dbQuery = dbQuery.Offset(offset).Limit(size).Order("created_at desc")

	// Execute query
	var forms []*domain.FormDesign
	if err := dbQuery.Find(&forms).Error; err != nil {
		logger.Error("failed to list form designs", "error", err)
		return nil, 0, errors.New("failed to list form designs")
	}

	return forms, total, nil
}

// GetFormDesign retrieves a form design by ID
func (s *FormDesignerService) GetFormDesign(ctx context.Context, formID string) (*domain.FormDesign, error) {
	var form domain.FormDesign
	if err := s.db.Table("form_designs").Where("id = ?", formID).First(&form).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("form design not found")
		}
		logger.Error("failed to find form design", "error", err)
		return nil, errors.New("failed to get form design")
	}

	return &form, nil
}

// PublishFormDesign publishes a form design
func (s *FormDesignerService) PublishFormDesign(ctx context.Context, formID string) error {
	// Find form design by ID
	var form domain.FormDesign
	if err := s.db.Table("form_designs").Where("id = ?", formID).First(&form).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errors.New("form design not found")
		}
		logger.Error("failed to find form design", "error", err)
		return errors.New("failed to publish form design")
	}

	// Check if form is already published
	if form.IsPublished {
		return errors.New("form design is already published")
	}

	// Update form design status to published
	form.IsPublished = true
	form.UpdatedAt = time.Now()

	// Save updated form design to database
	if err := s.db.Table("form_designs").Save(&form).Error; err != nil {
		logger.Error("failed to publish form design", "error", err)
		return errors.New("failed to publish form design")
	}

	return nil
}

// generateFormID generates a unique form ID
func generateFormID() string {
	// In a real application, use a proper ID generation library like uuid
	return "form_" + time.Now().Format("20060102150405")
}