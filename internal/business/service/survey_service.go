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

// SurveyServiceInterface defines the interface for survey service
type SurveyServiceInterface interface {
	CreateSurvey(ctx context.Context, req *CreateSurveyRequest) (*Survey, error)
	UpdateSurvey(ctx context.Context, surveyID string, req *UpdateSurveyRequest) error
	DeleteSurvey(ctx context.Context, surveyID string) error
	ListSurveys(ctx context.Context, teamID string, page, size int) ([]*Survey, int64, error)
	GetSurvey(ctx context.Context, surveyID string) (*Survey, error)
	PublishSurvey(ctx context.Context, surveyID string) error
	CloseSurvey(ctx context.Context, surveyID string) error
	SubmitResponse(ctx context.Context, req *SubmitResponseRequest) error
	GetSurveyResponses(ctx context.Context, surveyID string) ([]*SurveyResponse, error)
}

// SurveyService implements the SurveyServiceInterface
type SurveyService struct {
	db *gorm.DB
}

// NewSurveyService creates a new instance of SurveyService
func NewSurveyService() *SurveyService {
	return &SurveyService{
		db: database.GetDB(),
	}
}

// Survey represents a survey in the system
type Survey struct {
	ID          string    `json:"id" gorm:"primaryKey"`
	TeamID      string    `json:"team_id" gorm:"index"`
	Title       string    `json:"title" gorm:"size:200"`
	Description string    `json:"description" gorm:"type:text"`
	Content     string    `json:"content" gorm:"type:jsonb"`
	Status      string    `json:"status" gorm:"size:20"`
	CreatedBy   string    `json:"created_by" gorm:"size:50"`
	Questions   string    `json:"questions" gorm:"type:jsonb"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
	PublishedAt time.Time `json:"published_at"`
	ClosedAt    time.Time `json:"closed_at"`
}

// SurveyResponse represents a survey response in the system
type SurveyResponse struct {
	ID         string    `json:"id" gorm:"primaryKey"`
	SurveyID   string    `json:"survey_id" gorm:"index"`
	Respondent string    `json:"respondent" gorm:"size:50"`
	Content    string    `json:"content" gorm:"type:jsonb"`
	CreatedAt  time.Time `json:"created_at"`
}

// CreateSurveyRequest represents the request for creating a survey
type CreateSurveyRequest struct {
	TeamID      string `json:"team_id" binding:"required"`
	Title       string `json:"title" binding:"required"`
	Description string `json:"description"`
	Content     string `json:"content" binding:"required"`
	CreatedBy   string `json:"created_by" binding:"required"`
}

// UpdateSurveyRequest represents the request for updating a survey
type UpdateSurveyRequest struct {
	Title       string `json:"title"`
	Description string `json:"description"`
	Content     string `json:"content"`
}

// SubmitResponseRequest represents the request for submitting a survey response
type SubmitResponseRequest struct {
	SurveyID   string `json:"survey_id" binding:"required"`
	Respondent string `json:"respondent" binding:"required"`
	Content    string `json:"content" binding:"required"`
}

// CreateSurvey creates a new survey
func (s *SurveyService) CreateSurvey(ctx context.Context, req *CreateSurveyRequest) (*Survey, error) {
	// Create new survey
	survey := &Survey{
		ID:          utils.GenerateSurveyID(),
		Title:       req.Title,
		Description: req.Description,
		Status:      "draft",
		CreatedBy:   req.CreatedBy,
		TeamID:      req.TeamID,
		Questions:   convertStringToJSON(req.Content),
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	// Save survey to database
	if err := s.db.Create(survey).Error; err != nil {
		logger.Error("failed to create survey", "error", err)
		return nil, errors.New("failed to create survey")
	}

	return survey, nil
}

// UpdateSurvey updates an existing survey
func (s *SurveyService) UpdateSurvey(ctx context.Context, surveyID string, req *UpdateSurveyRequest) error {
	// Find survey by ID
	var survey Survey
	if err := s.db.Where("id = ?", surveyID).First(&survey).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errors.New("survey not found")
		}
		logger.Error("failed to find survey", "error", err)
		return errors.New("failed to update survey")
	}

	// Check if survey is in draft status
	if survey.Status != "draft" {
		return errors.New("only draft surveys can be updated")
	}

	// Update survey fields
	if req.Title != "" {
		survey.Title = req.Title
	}
	
	if req.Description != "" {
		survey.Description = req.Description
	}
	
	if req.Content != "" {
		survey.Content = req.Content
	}
	
	survey.UpdatedAt = time.Now()

	// Save updated survey to database
	if err := s.db.Save(&survey).Error; err != nil {
		logger.Error("failed to update survey", "error", err)
		return errors.New("failed to update survey")
	}

	return nil
}

// DeleteSurvey deletes a survey
func (s *SurveyService) DeleteSurvey(ctx context.Context, surveyID string) error {
	// Find survey by ID
	var survey Survey
	if err := s.db.Where("id = ?", surveyID).First(&survey).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errors.New("survey not found")
		}
		logger.Error("failed to find survey", "error", err)
		return errors.New("failed to delete survey")
	}

	// Check if survey is in draft status
	if survey.Status != "draft" {
		return errors.New("only draft surveys can be deleted")
	}

	// Delete survey from database
	if err := s.db.Delete(&survey).Error; err != nil {
		logger.Error("failed to delete survey", "error", err)
		return errors.New("failed to delete survey")
	}

	// Delete associated responses
	if err := s.db.Where("survey_id = ?", surveyID).Delete(&SurveyResponse{}).Error; err != nil {
		logger.Error("failed to delete survey responses", "error", err)
		// Don't return error here as the survey was successfully deleted
	}

	return nil
}

// ListSurveys lists surveys with pagination
func (s *SurveyService) ListSurveys(ctx context.Context, teamID string, page, size int) ([]*Survey, int64, error) {
	// Validate pagination parameters
	if page < 1 {
		page = 1
	}
	if size < 1 || size > 100 {
		size = 10
	}

	// Build query
	dbQuery := s.db.Model(&Survey{}).Where("team_id = ?", teamID)

	// Count total results
	var total int64
	if err := dbQuery.Count(&total).Error; err != nil {
		logger.Error("failed to count surveys", "error", err)
		return nil, 0, errors.New("failed to list surveys")
	}

	// Apply pagination
	offset := (page - 1) * size
	dbQuery = dbQuery.Offset(offset).Limit(size).Order("created_at desc")

	// Execute query
	var surveys []*Survey
	if err := dbQuery.Find(&surveys).Error; err != nil {
		logger.Error("failed to list surveys", "error", err)
		return nil, 0, errors.New("failed to list surveys")
	}

	return surveys, total, nil
}

// GetSurvey retrieves a survey by ID
func (s *SurveyService) GetSurvey(ctx context.Context, surveyID string) (*Survey, error) {
	var survey Survey
	if err := s.db.Where("id = ?", surveyID).First(&survey).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("survey not found")
		}
		logger.Error("failed to find survey", "error", err)
		return nil, errors.New("failed to get survey")
	}

	return &survey, nil
}

// PublishSurvey publishes a survey
func (s *SurveyService) PublishSurvey(ctx context.Context, surveyID string) error {
	// Find survey by ID
	var survey Survey
	if err := s.db.Where("id = ?", surveyID).First(&survey).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errors.New("survey not found")
		}
		logger.Error("failed to find survey", "error", err)
		return errors.New("failed to publish survey")
	}

	// Check if survey is in draft status
	if survey.Status != "draft" {
		return errors.New("only draft surveys can be published")
	}

	// Update survey status
	survey.Status = "published"
	survey.PublishedAt = time.Now()
	survey.UpdatedAt = time.Now()

	// Save updated survey to database
	if err := s.db.Save(&survey).Error; err != nil {
		logger.Error("failed to publish survey", "error", err)
		return errors.New("failed to publish survey")
	}

	return nil
}

// CloseSurvey closes a survey
func (s *SurveyService) CloseSurvey(ctx context.Context, surveyID string) error {
	// Find survey by ID
	var survey Survey
	if err := s.db.Where("id = ?", surveyID).First(&survey).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errors.New("survey not found")
		}
		logger.Error("failed to find survey", "error", err)
		return errors.New("failed to close survey")
	}

	// Check if survey is in published status
	if survey.Status != "published" {
		return errors.New("only published surveys can be closed")
	}

	// Update survey status
	survey.Status = "closed"
	survey.ClosedAt = time.Now()
	survey.UpdatedAt = time.Now()

	// Save updated survey to database
	if err := s.db.Save(&survey).Error; err != nil {
		logger.Error("failed to close survey", "error", err)
		return errors.New("failed to close survey")
	}

	return nil
}

// SubmitResponse submits a response to a survey
func (s *SurveyService) SubmitResponse(ctx context.Context, req *SubmitResponseRequest) error {
	// Find survey by ID
	var survey domain.Survey
	if err := s.db.Where("id = ?", req.SurveyID).First(&survey).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errors.New("survey not found")
		}
		logger.Error("failed to find survey", "error", err)
		return errors.New("failed to submit response")
	}

	// Check if survey is open
	if survey.Status == "closed" {
		return errors.New("survey is closed")
	}

	// Create new response
	response := &domain.SurveyResponse{
		ID:        utils.GenerateSurveyResponseID(),
		SurveyID:  req.SurveyID,
		UserID:    req.Respondent,
		Answers:   convertStringToJSON(req.Content),
		CreatedAt: time.Now(),
	}

	// Save response to database
	if err := s.db.Create(response).Error; err != nil {
		logger.Error("failed to create survey response", "error", err)
		return errors.New("failed to submit response")
	}

	return nil
}

// GetSurveyResponses retrieves all responses for a survey
func (s *SurveyService) GetSurveyResponses(ctx context.Context, surveyID string) ([]*SurveyResponse, error) {
	// Check if survey exists
	var survey Survey
	if err := s.db.Where("id = ?", surveyID).First(&survey).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("survey not found")
		}
		logger.Error("failed to find survey", "error", err)
		return nil, errors.New("failed to get survey responses")
	}

	// Get survey responses
	var responses []*SurveyResponse
	if err := s.db.Where("survey_id = ?", surveyID).Order("created_at asc").Find(&responses).Error; err != nil {
		logger.Error("failed to find survey responses", "error", err)
		return nil, errors.New("failed to get survey responses")
	}

	return responses, nil
}