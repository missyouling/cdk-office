package handler

import (
	"net/http"
	"strconv"

	"cdk-office/internal/app/domain"
	"cdk-office/internal/app/service"
	"github.com/gin-gonic/gin"
)

// FormDesignerHandlerInterface defines the interface for form designer handler
type FormDesignerHandlerInterface interface {
	CreateFormDesign(c *gin.Context)
	UpdateFormDesign(c *gin.Context)
	DeleteFormDesign(c *gin.Context)
	ListFormDesigns(c *gin.Context)
	GetFormDesign(c *gin.Context)
	PublishFormDesign(c *gin.Context)
}

// FormDesignerHandler implements the FormDesignerHandlerInterface
type FormDesignerHandler struct {
	formService service.FormDesignerServiceInterface
}

// NewFormDesignerHandler creates a new instance of FormDesignerHandler
func NewFormDesignerHandler() *FormDesignerHandler {
	return &FormDesignerHandler{
		formService: service.NewFormDesignerService(),
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

// ListFormDesignsRequest represents the request for listing form designs
type ListFormDesignsRequest struct {
	AppID string `form:"app_id" binding:"required"`
	Page  int    `form:"page"`
	Size  int    `form:"size"`
}

// CreateFormDesign handles creating a new form design
func (h *FormDesignerHandler) CreateFormDesign(c *gin.Context) {
	var req CreateFormDesignRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Call service to create form design
	form, err := h.formService.CreateFormDesign(c.Request.Context(), &service.CreateFormDesignRequest{
		AppID:       req.AppID,
		Name:        req.Name,
		Description: req.Description,
		Schema:      req.Schema,
		Config:      req.Config,
		CreatedBy:   req.CreatedBy,
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, form)
}

// UpdateFormDesign handles updating an existing form design
func (h *FormDesignerHandler) UpdateFormDesign(c *gin.Context) {
	formID := c.Param("id")
	if formID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "form id is required"})
		return
	}

	var req UpdateFormDesignRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Call service to update form design
	if err := h.formService.UpdateFormDesign(c.Request.Context(), formID, &service.UpdateFormDesignRequest{
		Name:        req.Name,
		Description: req.Description,
		Schema:      req.Schema,
		Config:      req.Config,
		IsActive:    req.IsActive,
	}); err != nil {
		if err.Error() == "form design not found" {
			c.JSON(http.StatusNotFound, gin.H{"error": "form design not found"})
			return
		}
		if err.Error() == "cannot update published form design" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "cannot update published form design"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "form design updated successfully"})
}

// DeleteFormDesign handles deleting a form design
func (h *FormDesignerHandler) DeleteFormDesign(c *gin.Context) {
	formID := c.Param("id")
	if formID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "form id is required"})
		return
	}

	// Call service to delete form design
	if err := h.formService.DeleteFormDesign(c.Request.Context(), formID); err != nil {
		if err.Error() == "form design not found" {
			c.JSON(http.StatusNotFound, gin.H{"error": "form design not found"})
			return
		}
		if err.Error() == "cannot delete published form design" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "cannot delete published form design"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "form design deleted successfully"})
}

// ListFormDesigns handles listing form designs with pagination
func (h *FormDesignerHandler) ListFormDesigns(c *gin.Context) {
	// Parse query parameters
	appID := c.Query("app_id")
	if appID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "app id is required"})
		return
	}

	page := 1
	size := 10

	if pageStr := c.Query("page"); pageStr != "" {
		if p, err := strconv.Atoi(pageStr); err == nil && p > 0 {
			page = p
		}
	}

	if sizeStr := c.Query("size"); sizeStr != "" {
		if s, err := strconv.Atoi(sizeStr); err == nil && s > 0 && s <= 100 {
			size = s
		}
	}

	// Call service to list form designs
	forms, total, err := h.formService.ListFormDesigns(c.Request.Context(), appID, page, size)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Build response
	response := ListFormDesignsResponse{
		Items: forms,
		Total: total,
		Page:  page,
		Size:  size,
	}

	c.JSON(http.StatusOK, response)
}

// GetFormDesign handles retrieving a form design by ID
func (h *FormDesignerHandler) GetFormDesign(c *gin.Context) {
	formID := c.Param("id")
	if formID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "form id is required"})
		return
	}

	// Call service to get form design
	form, err := h.formService.GetFormDesign(c.Request.Context(), formID)
	if err != nil {
		if err.Error() == "form design not found" {
			c.JSON(http.StatusNotFound, gin.H{"error": "form design not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, form)
}

// PublishFormDesign handles publishing a form design
func (h *FormDesignerHandler) PublishFormDesign(c *gin.Context) {
	formID := c.Param("id")
	if formID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "form id is required"})
		return
	}

	// Call service to publish form design
	if err := h.formService.PublishFormDesign(c.Request.Context(), formID); err != nil {
		if err.Error() == "form design not found" {
			c.JSON(http.StatusNotFound, gin.H{"error": "form design not found"})
			return
		}
		if err.Error() == "form design is already published" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "form design is already published"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "form design published successfully"})
}

// ListFormDesignsResponse represents the response for listing form designs
type ListFormDesignsResponse struct {
	Items []*domain.FormDesign `json:"items"`
	Total int64                 `json:"total"`
	Page  int                   `json:"page"`
	Size  int                   `json:"size"`
}