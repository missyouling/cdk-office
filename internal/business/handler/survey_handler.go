package handler

import (
	"net/http"
	"strconv"

	"cdk-office/internal/business/service"
	"github.com/gin-gonic/gin"
)

// SurveyHandlerInterface defines the interface for survey handler
type SurveyHandlerInterface interface {
	CreateSurvey(c *gin.Context)
	UpdateSurvey(c *gin.Context)
	DeleteSurvey(c *gin.Context)
	ListSurveys(c *gin.Context)
	GetSurvey(c *gin.Context)
	PublishSurvey(c *gin.Context)
	CloseSurvey(c *gin.Context)
	SubmitResponse(c *gin.Context)
	GetSurveyResponses(c *gin.Context)
}

// SurveyHandler implements the SurveyHandlerInterface
type SurveyHandler struct {
	surveyService service.SurveyServiceInterface
}

// NewSurveyHandler creates a new instance of SurveyHandler
func NewSurveyHandler() *SurveyHandler {
	return &SurveyHandler{
		surveyService: service.NewSurveyService(),
	}
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

// ListSurveysRequest represents the request for listing surveys
type ListSurveysRequest struct {
	TeamID string `form:"team_id" binding:"required"`
	Page   int    `form:"page"`
	Size   int    `form:"size"`
}

// CreateSurvey handles creating a new survey
func (h *SurveyHandler) CreateSurvey(c *gin.Context) {
	var req CreateSurveyRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Call service to create survey
	survey, err := h.surveyService.CreateSurvey(c.Request.Context(), &service.CreateSurveyRequest{
		TeamID:      req.TeamID,
		Title:       req.Title,
		Description: req.Description,
		Content:     req.Content,
		CreatedBy:   req.CreatedBy,
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, survey)
}

// UpdateSurvey handles updating an existing survey
func (h *SurveyHandler) UpdateSurvey(c *gin.Context) {
	surveyID := c.Param("id")
	if surveyID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "survey id is required"})
		return
	}

	var req UpdateSurveyRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Call service to update survey
	if err := h.surveyService.UpdateSurvey(c.Request.Context(), surveyID, &service.UpdateSurveyRequest{
		Title:       req.Title,
		Description: req.Description,
		Content:     req.Content,
	}); err != nil {
		if err.Error() == "survey not found" {
			c.JSON(http.StatusNotFound, gin.H{"error": "survey not found"})
			return
		}
		if err.Error() == "only draft surveys can be updated" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "only draft surveys can be updated"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "survey updated successfully"})
}

// DeleteSurvey handles deleting a survey
func (h *SurveyHandler) DeleteSurvey(c *gin.Context) {
	surveyID := c.Param("id")
	if surveyID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "survey id is required"})
		return
	}

	// Call service to delete survey
	if err := h.surveyService.DeleteSurvey(c.Request.Context(), surveyID); err != nil {
		if err.Error() == "survey not found" {
			c.JSON(http.StatusNotFound, gin.H{"error": "survey not found"})
			return
		}
		if err.Error() == "only draft surveys can be deleted" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "only draft surveys can be deleted"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "survey deleted successfully"})
}

// ListSurveys handles listing surveys with pagination
func (h *SurveyHandler) ListSurveys(c *gin.Context) {
	// Parse query parameters
	teamID := c.Query("team_id")
	if teamID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "team id is required"})
		return
	}

	pageStr := c.Query("page")
	sizeStr := c.Query("size")

	page := 1
	size := 10

	var err error
	if pageStr != "" {
		page, err = strconv.Atoi(pageStr)
		if err != nil || page < 1 {
			page = 1
		}
	}

	if sizeStr != "" {
		size, err = strconv.Atoi(sizeStr)
		if err != nil || size < 1 || size > 100 {
			size = 10
		}
	}

	// Call service to list surveys
	surveys, total, err := h.surveyService.ListSurveys(c.Request.Context(), teamID, page, size)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Build response
	response := ListSurveysResponse{
		Items: surveys,
		Total: total,
		Page:  page,
		Size:  size,
	}

	c.JSON(http.StatusOK, response)
}

// GetSurvey handles retrieving a survey by ID
func (h *SurveyHandler) GetSurvey(c *gin.Context) {
	surveyID := c.Param("id")
	if surveyID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "survey id is required"})
		return
	}

	// Call service to get survey
	survey, err := h.surveyService.GetSurvey(c.Request.Context(), surveyID)
	if err != nil {
		if err.Error() == "survey not found" {
			c.JSON(http.StatusNotFound, gin.H{"error": "survey not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, survey)
}

// PublishSurvey handles publishing a survey
func (h *SurveyHandler) PublishSurvey(c *gin.Context) {
	surveyID := c.Param("id")
	if surveyID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "survey id is required"})
		return
	}

	// Call service to publish survey
	if err := h.surveyService.PublishSurvey(c.Request.Context(), surveyID); err != nil {
		if err.Error() == "survey not found" {
			c.JSON(http.StatusNotFound, gin.H{"error": "survey not found"})
			return
		}
		if err.Error() == "only draft surveys can be published" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "only draft surveys can be published"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "survey published successfully"})
}

// CloseSurvey handles closing a survey
func (h *SurveyHandler) CloseSurvey(c *gin.Context) {
	surveyID := c.Param("id")
	if surveyID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "survey id is required"})
		return
	}

	// Call service to close survey
	if err := h.surveyService.CloseSurvey(c.Request.Context(), surveyID); err != nil {
		if err.Error() == "survey not found" {
			c.JSON(http.StatusNotFound, gin.H{"error": "survey not found"})
			return
		}
		if err.Error() == "only published surveys can be closed" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "only published surveys can be closed"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "survey closed successfully"})
}

// SubmitResponse handles submitting a survey response
func (h *SurveyHandler) SubmitResponse(c *gin.Context) {
	var req SubmitResponseRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Call service to submit response
	if err := h.surveyService.SubmitResponse(c.Request.Context(), &service.SubmitResponseRequest{
		SurveyID:   req.SurveyID,
		Respondent: req.Respondent,
		Content:    req.Content,
	}); err != nil {
		if err.Error() == "survey not found" {
			c.JSON(http.StatusNotFound, gin.H{"error": "survey not found"})
			return
		}
		if err.Error() == "survey is not published" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "survey is not published"})
			return
		}
		if err.Error() == "survey is closed" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "survey is closed"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "response submitted successfully"})
}

// GetSurveyResponses handles retrieving all responses for a survey
func (h *SurveyHandler) GetSurveyResponses(c *gin.Context) {
	surveyID := c.Param("id")
	if surveyID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "survey id is required"})
		return
	}

	// Call service to get survey responses
	responses, err := h.surveyService.GetSurveyResponses(c.Request.Context(), surveyID)
	if err != nil {
		if err.Error() == "survey not found" {
			c.JSON(http.StatusNotFound, gin.H{"error": "survey not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, responses)
}

// ListSurveysResponse represents the response for listing surveys
type ListSurveysResponse struct {
	Items []*service.Survey `json:"items"`
	Total int64             `json:"total"`
	Page  int               `json:"page"`
	Size  int               `json:"size"`
}