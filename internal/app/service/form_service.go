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

// FormServiceInterface defines the interface for form service
type FormServiceInterface interface {
	CreateForm(ctx context.Context, req *CreateFormRequest) (*domain.FormData, error)
	UpdateForm(ctx context.Context, formID string, req *UpdateFormRequest) error
	DeleteForm(ctx context.Context, formID string) error
	ListForms(ctx context.Context, appID string, page, size int) ([]*domain.FormData, int64, error)
	GetForm(ctx context.Context, formID string) (*domain.FormData, error)
	SubmitFormData(ctx context.Context, req *SubmitFormDataRequest) (*domain.FormDataEntry, error)
	ListFormDataEntries(ctx context.Context, formID string, page, size int) ([]*domain.FormDataEntry, int64, error)
}

// FormService implements the FormServiceInterface
type FormService struct {
	db *gorm.DB
}

// NewFormService creates a new instance of FormService
func NewFormService() *FormService {
	return &FormService{
		db: database.GetDB(),
	}
}

// CreateFormRequest represents the request for creating a form
type CreateFormRequest struct {
	AppID       string `json:"app_id" binding:"required"`
	Name        string `json:"name" binding:"required"`
	Description string `json:"description"`
	Schema      string `json:"schema" binding:"required"`
	CreatedBy   string `json:"created_by" binding:"required"`
}

// UpdateFormRequest represents the request for updating a form
type UpdateFormRequest struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	Schema      string `json:"schema"`
	IsActive    *bool  `json:"is_active"`
}

// SubmitFormDataRequest represents the request for submitting form data
type SubmitFormDataRequest struct {
	FormID    string `json:"form_id" binding:"required"`
	Data      string `json:"data" binding:"required"`
	CreatedBy string `json:"created_by" binding:"required"`
}

// CreateForm creates a new form
func (s *FormService) CreateForm(ctx context.Context, req *CreateFormRequest) (*domain.FormData, error) {
	// Create new form
	form := &domain.FormData{
		ID:          utils.GenerateFormID(),
		AppID:       req.AppID,
		Name:        req.Name,
		Description: req.Description,
		Schema:      req.Schema,
		IsActive:    true,
		CreatedBy:   req.CreatedBy,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	// Save form to database
	if err := s.db.Create(form).Error; err != nil {
		logger.Error("failed to create form", "error", err)
		return nil, errors.New("failed to create form")
	}

	return form, nil
}

// UpdateForm updates an existing form
func (s *FormService) UpdateForm(ctx context.Context, formID string, req *UpdateFormRequest) error {
	// Find form by ID
	var form domain.FormData
	if err := s.db.Where("id = ?", formID).First(&form).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errors.New("form not found")
		}
		logger.Error("failed to find form", "error", err)
		return errors.New("failed to update form")
	}

	// Update form fields
	if req.Name != "" {
		form.Name = req.Name
	}

	if req.Description != "" {
		form.Description = req.Description
	}

	if req.Schema != "" {
		form.Schema = req.Schema
	}

	if req.IsActive != nil {
		form.IsActive = *req.IsActive
	}

	form.UpdatedAt = time.Now()

	// Save updated form to database
	if err := s.db.Save(&form).Error; err != nil {
		logger.Error("failed to update form", "error", err)
		return errors.New("failed to update form")
	}

	return nil
}

// DeleteForm deletes a form
func (s *FormService) DeleteForm(ctx context.Context, formID string) error {
	// Find form by ID
	var form domain.FormData
	if err := s.db.Where("id = ?", formID).First(&form).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errors.New("form not found")
		}
		logger.Error("failed to find form", "error", err)
		return errors.New("failed to delete form")
	}

	// Delete form from database
	if err := s.db.Delete(&form).Error; err != nil {
		logger.Error("failed to delete form", "error", err)
		return errors.New("failed to delete form")
	}

	return nil
}

// ListForms lists forms with pagination
func (s *FormService) ListForms(ctx context.Context, appID string, page, size int) ([]*domain.FormData, int64, error) {
	// Validate pagination parameters
	if page < 1 {
		page = 1
	}
	if size < 1 || size > 100 {
		size = 10
	}

	// Build query
	dbQuery := s.db.Model(&domain.FormData{}).Where("app_id = ?", appID)

	// Count total results
	var total int64
	if err := dbQuery.Count(&total).Error; err != nil {
		logger.Error("failed to count forms", "error", err)
		return nil, 0, errors.New("failed to list forms")
	}

	// Apply pagination
	offset := (page - 1) * size
	dbQuery = dbQuery.Offset(offset).Limit(size).Order("created_at desc")

	// Execute query
	var forms []*domain.FormData
	if err := dbQuery.Find(&forms).Error; err != nil {
		logger.Error("failed to list forms", "error", err)
		return nil, 0, errors.New("failed to list forms")
	}

	return forms, total, nil
}

// GetForm retrieves a form by ID
func (s *FormService) GetForm(ctx context.Context, formID string) (*domain.FormData, error) {
	var form domain.FormData
	if err := s.db.Where("id = ?", formID).First(&form).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("form not found")
		}
		logger.Error("failed to find form", "error", err)
		return nil, errors.New("failed to get form")
	}

	return &form, nil
}

// SubmitFormData submits form data
func (s *FormService) SubmitFormData(ctx context.Context, req *SubmitFormDataRequest) (*domain.FormDataEntry, error) {
	// Verify form exists and is active
	var form domain.FormData
	if err := s.db.Where("id = ? AND is_active = ?", req.FormID, true).First(&form).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("form not found or inactive")
		}
		logger.Error("failed to find form", "error", err)
		return nil, errors.New("failed to submit form data")
	}

	// Create form data entry
	entry := &domain.FormDataEntry{
		ID:        generateFormEntryID(),
		FormID:    req.FormID,
		Data:      req.Data,
		CreatedBy: req.CreatedBy,
		CreatedAt: time.Now(),
	}

	// Save form data entry to database
	if err := s.db.Create(entry).Error; err != nil {
		logger.Error("failed to submit form data", "error", err)
		return nil, errors.New("failed to submit form data")
	}

	return entry, nil
}

// ListFormDataEntries lists form data entries with pagination
func (s *FormService) ListFormDataEntries(ctx context.Context, formID string, page, size int) ([]*domain.FormDataEntry, int64, error) {
	// Validate pagination parameters
	if page < 1 {
		page = 1
	}
	if size < 1 || size > 100 {
		size = 10
	}

	// Build query
	dbQuery := s.db.Model(&domain.FormDataEntry{}).Where("form_id = ?", formID)

	// Count total results
	var total int64
	if err := dbQuery.Count(&total).Error; err != nil {
		logger.Error("failed to count form data entries", "error", err)
		return nil, 0, errors.New("failed to list form data entries")
	}

	// Apply pagination
	offset := (page - 1) * size
	dbQuery = dbQuery.Offset(offset).Limit(size).Order("created_at desc")

	// Execute query
	var entries []*domain.FormDataEntry
	if err := dbQuery.Find(&entries).Error; err != nil {
		logger.Error("failed to list form data entries", "error", err)
		return nil, 0, errors.New("failed to list form data entries")
	}

	return entries, total, nil
}



// generateFormEntryID generates a unique form entry ID
func generateFormEntryID() string {
	// In a real application, use a proper ID generation library like uuid
	return "entry_" + time.Now().Format("20060102150405")
}