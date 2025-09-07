package handler

import (
	"net/http"
	"strconv"

	"cdk-office/internal/app/domain"
	"cdk-office/internal/app/service"
	"github.com/gin-gonic/gin"
)

// FormHandlerInterface defines the interface for form handler
type FormHandlerInterface interface {
	CreateForm(c *gin.Context)
	UpdateForm(c *gin.Context)
	DeleteForm(c *gin.Context)
	ListForms(c *gin.Context)
	GetForm(c *gin.Context)
	SubmitFormData(c *gin.Context)
	ListFormDataEntries(c *gin.Context)
}

// FormHandler implements the FormHandlerInterface
type FormHandler struct {
	formService service.FormServiceInterface
}

// NewFormHandler creates a new instance of FormHandler
func NewFormHandler() *FormHandler {
	return &FormHandler{
		formService: service.NewFormService(),
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

// ListFormsRequest represents the request for listing forms
type ListFormsRequest struct {
	AppID string `form:"app_id" binding:"required"`
	Page  int    `form:"page"`
	Size  int    `form:"size"`
}

// ListFormDataEntriesRequest represents the request for listing form data entries
type ListFormDataEntriesRequest struct {
	FormID string `form:"form_id" binding:"required"`
	Page   int    `form:"page"`
	Size   int    `form:"size"`
}

// CreateForm handles creating a new form
func (h *FormHandler) CreateForm(c *gin.Context) {
	var req CreateFormRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Call service to create form
	form, err := h.formService.CreateForm(c.Request.Context(), &service.CreateFormRequest{
		AppID:       req.AppID,
		Name:        req.Name,
		Description: req.Description,
		Schema:      req.Schema,
		CreatedBy:   req.CreatedBy,
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, form)
}

// UpdateForm handles updating an existing form
func (h *FormHandler) UpdateForm(c *gin.Context) {
	formID := c.Param("id")
	if formID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "form id is required"})
		return
	}

	var req UpdateFormRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Call service to update form
	if err := h.formService.UpdateForm(c.Request.Context(), formID, &service.UpdateFormRequest{
		Name:        req.Name,
		Description: req.Description,
		Schema:      req.Schema,
		IsActive:    req.IsActive,
	}); err != nil {
		if err.Error() == "form not found" {
			c.JSON(http.StatusNotFound, gin.H{"error": "form not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "form updated successfully"})
}

// DeleteForm handles deleting a form
func (h *FormHandler) DeleteForm(c *gin.Context) {
	formID := c.Param("id")
	if formID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "form id is required"})
		return
	}

	// Call service to delete form
	if err := h.formService.DeleteForm(c.Request.Context(), formID); err != nil {
		if err.Error() == "form not found" {
			c.JSON(http.StatusNotFound, gin.H{"error": "form not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "form deleted successfully"})
}

// ListForms handles listing forms with pagination
func (h *FormHandler) ListForms(c *gin.Context) {
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

	// Call service to list forms
	forms, total, err := h.formService.ListForms(c.Request.Context(), appID, page, size)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Build response
	response := ListFormsResponse{
		Items: forms,
		Total: total,
		Page:  page,
		Size:  size,
	}

	c.JSON(http.StatusOK, response)
}

// GetForm handles retrieving a form by ID
func (h *FormHandler) GetForm(c *gin.Context) {
	formID := c.Param("id")
	if formID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "form id is required"})
		return
	}

	// Call service to get form
	form, err := h.formService.GetForm(c.Request.Context(), formID)
	if err != nil {
		if err.Error() == "form not found" {
			c.JSON(http.StatusNotFound, gin.H{"error": "form not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, form)
}

// SubmitFormData handles submitting form data
func (h *FormHandler) SubmitFormData(c *gin.Context) {
	var req SubmitFormDataRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Call service to submit form data
	entry, err := h.formService.SubmitFormData(c.Request.Context(), &service.SubmitFormDataRequest{
		FormID:    req.FormID,
		Data:      req.Data,
		CreatedBy: req.CreatedBy,
	})
	if err != nil {
		if err.Error() == "form not found or inactive" {
			c.JSON(http.StatusNotFound, gin.H{"error": "form not found or inactive"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, entry)
}

// ListFormDataEntries handles listing form data entries with pagination
func (h *FormHandler) ListFormDataEntries(c *gin.Context) {
	// Parse query parameters
	formID := c.Query("form_id")
	if formID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "form id is required"})
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

	// Call service to list form data entries
	entries, total, err := h.formService.ListFormDataEntries(c.Request.Context(), formID, page, size)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Build response
	response := ListFormDataEntriesResponse{
		Items: entries,
		Total: total,
		Page:  page,
		Size:  size,
	}

	c.JSON(http.StatusOK, response)
}

// ListFormsResponse represents the response for listing forms
type ListFormsResponse struct {
	Items []*domain.FormData `json:"items"`
	Total int64               `json:"total"`
	Page  int                 `json:"page"`
	Size  int                 `json:"size"`
}

// ListFormDataEntriesResponse represents the response for listing form data entries
type ListFormDataEntriesResponse struct {
	Items []*domain.FormDataEntry `json:"items"`
	Total int64                    `json:"total"`
	Page  int                      `json:"page"`
	Size  int                      `json:"size"`
}