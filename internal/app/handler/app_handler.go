package handler

import (
	"net/http"
	"strconv"

	"cdk-office/internal/app/domain"
	"cdk-office/internal/app/service"
	"github.com/gin-gonic/gin"
)

// AppHandlerInterface defines the interface for application handler
type AppHandlerInterface interface {
	CreateApplication(c *gin.Context)
	UpdateApplication(c *gin.Context)
	DeleteApplication(c *gin.Context)
	ListApplications(c *gin.Context)
	GetApplication(c *gin.Context)
}

// AppHandler implements the AppHandlerInterface
type AppHandler struct {
	appService service.AppServiceInterface
}

// NewAppHandler creates a new instance of AppHandler
func NewAppHandler() *AppHandler {
	return &AppHandler{
		appService: service.NewAppService(),
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

// ListApplicationsRequest represents the request for listing applications
type ListApplicationsRequest struct {
	TeamID string `form:"team_id" binding:"required"`
	Page   int    `form:"page"`
	Size   int    `form:"size"`
}

// CreateApplication handles creating a new application
func (h *AppHandler) CreateApplication(c *gin.Context) {
	var req CreateApplicationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Call service to create application
	app, err := h.appService.CreateApplication(c.Request.Context(), &service.CreateApplicationRequest{
		TeamID:      req.TeamID,
		Name:        req.Name,
		Description: req.Description,
		Type:        req.Type,
		Config:      req.Config,
		CreatedBy:   req.CreatedBy,
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, app)
}

// UpdateApplication handles updating an existing application
func (h *AppHandler) UpdateApplication(c *gin.Context) {
	appID := c.Param("id")
	if appID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "application id is required"})
		return
	}

	var req UpdateApplicationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Call service to update application
	if err := h.appService.UpdateApplication(c.Request.Context(), appID, &service.UpdateApplicationRequest{
		Name:        req.Name,
		Description: req.Description,
		Config:      req.Config,
		IsActive:    req.IsActive,
	}); err != nil {
		if err.Error() == "application not found" {
			c.JSON(http.StatusNotFound, gin.H{"error": "application not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "application updated successfully"})
}

// DeleteApplication handles deleting an application
func (h *AppHandler) DeleteApplication(c *gin.Context) {
	appID := c.Param("id")
	if appID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "application id is required"})
		return
	}

	// Call service to delete application
	if err := h.appService.DeleteApplication(c.Request.Context(), appID); err != nil {
		if err.Error() == "application not found" {
			c.JSON(http.StatusNotFound, gin.H{"error": "application not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "application deleted successfully"})
}

// ListApplications handles listing applications with pagination
func (h *AppHandler) ListApplications(c *gin.Context) {
	// Parse query parameters
	teamID := c.Query("team_id")
	if teamID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "team id is required"})
		return
	}

	page := 1
	size := 10

	if pageStr := c.Query("page"); pageStr != "" {
		if p, err := parseInt(pageStr); err == nil && p > 0 {
			page = p
		}
	}

	if sizeStr := c.Query("size"); sizeStr != "" {
		if s, err := parseInt(sizeStr); err == nil && s > 0 && s <= 100 {
			size = s
		}
	}

	// Call service to list applications
	apps, total, err := h.appService.ListApplications(c.Request.Context(), teamID, page, size)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Build response
	response := ListApplicationsResponse{
		Items: apps,
		Total: total,
		Page:  page,
		Size:  size,
	}

	c.JSON(http.StatusOK, response)
}

// GetApplication handles retrieving an application by ID
func (h *AppHandler) GetApplication(c *gin.Context) {
	appID := c.Param("id")
	if appID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "application id is required"})
		return
	}

	// Call service to get application
	app, err := h.appService.GetApplication(c.Request.Context(), appID)
	if err != nil {
		if err.Error() == "application not found" {
			c.JSON(http.StatusNotFound, gin.H{"error": "application not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, app)
}

// ListApplicationsResponse represents the response for listing applications
type ListApplicationsResponse struct {
	Items []*domain.Application `json:"items"`
	Total int64                  `json:"total"`
	Page  int                    `json:"page"`
	Size  int                    `json:"size"`
}

// parseInt converts string to int, returns error if conversion fails
func parseInt(s string) (int, error) {
	// Convert string to int
	return strconv.Atoi(s)
}