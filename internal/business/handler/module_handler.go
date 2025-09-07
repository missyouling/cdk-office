package handler

import (
	"net/http"
	"strconv"

	"cdk-office/internal/business/service"
	"github.com/gin-gonic/gin"
)

// ModuleHandlerInterface defines the interface for business module handler
type ModuleHandlerInterface interface {
	CreateModule(c *gin.Context)
	UpdateModule(c *gin.Context)
	DeleteModule(c *gin.Context)
	ListModules(c *gin.Context)
	GetModule(c *gin.Context)
	ActivateModule(c *gin.Context)
	DeactivateModule(c *gin.Context)
}

// ModuleHandler implements the ModuleHandlerInterface
type ModuleHandler struct {
	moduleService service.ModuleServiceInterface
}

// NewModuleHandler creates a new instance of ModuleHandler
func NewModuleHandler() *ModuleHandler {
	return &ModuleHandler{
		moduleService: service.NewModuleService(),
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

// CreateModule handles creating a new business module
func (h *ModuleHandler) CreateModule(c *gin.Context) {
	var req CreateModuleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Call service to create module
	module, err := h.moduleService.CreateModule(c.Request.Context(), &service.CreateModuleRequest{
		Name:        req.Name,
		Description: req.Description,
		Version:     req.Version,
		Config:      req.Config,
	})
	if err != nil {
		if err.Error() == "module name already exists" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "module name already exists"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, module)
}

// UpdateModule handles updating an existing business module
func (h *ModuleHandler) UpdateModule(c *gin.Context) {
	moduleID := c.Param("id")
	if moduleID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "module id is required"})
		return
	}

	var req UpdateModuleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Call service to update module
	if err := h.moduleService.UpdateModule(c.Request.Context(), moduleID, &service.UpdateModuleRequest{
		Name:        req.Name,
		Description: req.Description,
		Version:     req.Version,
		Config:      req.Config,
	}); err != nil {
		if err.Error() == "module not found" {
			c.JSON(http.StatusNotFound, gin.H{"error": "module not found"})
			return
		}
		if err.Error() == "module name already exists" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "module name already exists"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "module updated successfully"})
}

// DeleteModule handles deleting a business module
func (h *ModuleHandler) DeleteModule(c *gin.Context) {
	moduleID := c.Param("id")
	if moduleID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "module id is required"})
		return
	}

	// Call service to delete module
	if err := h.moduleService.DeleteModule(c.Request.Context(), moduleID); err != nil {
		if err.Error() == "module not found" {
			c.JSON(http.StatusNotFound, gin.H{"error": "module not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "module deleted successfully"})
}

// ListModules handles listing business modules
func (h *ModuleHandler) ListModules(c *gin.Context) {
	// Parse query parameters
	isActiveStr := c.Query("is_active")
	var isActive *bool
	
	if isActiveStr != "" {
		isActiveVal, err := strconv.ParseBool(isActiveStr)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid is_active parameter"})
			return
		}
		isActive = &isActiveVal
	}

	// Call service to list modules
	modules, err := h.moduleService.ListModules(c.Request.Context(), isActive)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, modules)
}

// GetModule handles retrieving a business module by ID
func (h *ModuleHandler) GetModule(c *gin.Context) {
	moduleID := c.Param("id")
	if moduleID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "module id is required"})
		return
	}

	// Call service to get module
	module, err := h.moduleService.GetModule(c.Request.Context(), moduleID)
	if err != nil {
		if err.Error() == "module not found" {
			c.JSON(http.StatusNotFound, gin.H{"error": "module not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, module)
}

// ActivateModule handles activating a business module
func (h *ModuleHandler) ActivateModule(c *gin.Context) {
	moduleID := c.Param("id")
	if moduleID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "module id is required"})
		return
	}

	// Call service to activate module
	if err := h.moduleService.ActivateModule(c.Request.Context(), moduleID); err != nil {
		if err.Error() == "module not found" {
			c.JSON(http.StatusNotFound, gin.H{"error": "module not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "module activated successfully"})
}

// DeactivateModule handles deactivating a business module
func (h *ModuleHandler) DeactivateModule(c *gin.Context) {
	moduleID := c.Param("id")
	if moduleID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "module id is required"})
		return
	}

	// Call service to deactivate module
	if err := h.moduleService.DeactivateModule(c.Request.Context(), moduleID); err != nil {
		if err.Error() == "module not found" {
			c.JSON(http.StatusNotFound, gin.H{"error": "module not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "module deactivated successfully"})
}