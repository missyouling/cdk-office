package handler

import (
	"net/http"
	"strconv"

	"cdk-office/internal/app/domain"
	"cdk-office/internal/app/service"
	"github.com/gin-gonic/gin"
)

// PermissionHandlerInterface defines the interface for permission handler
type PermissionHandlerInterface interface {
	CreatePermission(c *gin.Context)
	UpdatePermission(c *gin.Context)
	DeletePermission(c *gin.Context)
	ListPermissions(c *gin.Context)
	GetPermission(c *gin.Context)
	AssignPermission(c *gin.Context)
	RevokePermission(c *gin.Context)
	ListUserPermissions(c *gin.Context)
	CheckPermission(c *gin.Context)
}

// PermissionHandler implements the PermissionHandlerInterface
type PermissionHandler struct {
	permissionService service.PermissionServiceInterface
}

// NewPermissionHandler creates a new instance of PermissionHandler
func NewPermissionHandler() *PermissionHandler {
	return &PermissionHandler{
		permissionService: service.NewPermissionService(),
	}
}

// CreatePermissionRequest represents the request for creating a permission
type CreatePermissionRequest struct {
	AppID       string `json:"app_id" binding:"required"`
	Name        string `json:"name" binding:"required"`
	Description string `json:"description"`
	Action      string `json:"action" binding:"required"`
	CreatedBy   string `json:"created_by" binding:"required"`
}

// UpdatePermissionRequest represents the request for updating a permission
type UpdatePermissionRequest struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	Action      string `json:"action"`
}

// AssignPermissionRequest represents the request for assigning a permission to a user
type AssignPermissionRequest struct {
	AppID        string `json:"app_id" binding:"required"`
	UserID       string `json:"user_id" binding:"required"`
	PermissionID string `json:"permission_id" binding:"required"`
	AssignedBy   string `json:"assigned_by" binding:"required"`
}

// RevokePermissionRequest represents the request for revoking a permission from a user
type RevokePermissionRequest struct {
	AppID        string `json:"app_id" binding:"required"`
	UserID       string `json:"user_id" binding:"required"`
	PermissionID string `json:"permission_id" binding:"required"`
}

// CheckPermissionRequest represents the request for checking a permission
type CheckPermissionRequest struct {
	AppID  string `json:"app_id" binding:"required"`
	UserID string `json:"user_id" binding:"required"`
	Action string `json:"action" binding:"required"`
}

// ListPermissionsRequest represents the request for listing permissions
type ListPermissionsRequest struct {
	AppID string `form:"app_id" binding:"required"`
	Page  int    `form:"page"`
	Size  int    `form:"size"`
}

// ListUserPermissionsRequest represents the request for listing user permissions
type PermissionListUserPermissionsRequest struct {
	AppID  string `form:"app_id" binding:"required"`
	UserID string `form:"user_id" binding:"required"`
}

// CreatePermission handles creating a new permission
func (h *PermissionHandler) CreatePermission(c *gin.Context) {
	var req CreatePermissionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Call service to create permission
	permission, err := h.permissionService.CreatePermission(c.Request.Context(), &service.CreatePermissionRequest{
		AppID:       req.AppID,
		Name:        req.Name,
		Description: req.Description,
		Action:      req.Action,
		CreatedBy:   req.CreatedBy,
	})
	if err != nil {
		if err.Error() == "invalid permission action" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid permission action"})
			return
		}
		if err.Error() == "permission with this name already exists in the application" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "permission with this name already exists in the application"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, permission)
}

// UpdatePermission handles updating an existing permission
func (h *PermissionHandler) UpdatePermission(c *gin.Context) {
	permissionID := c.Param("id")
	if permissionID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "permission id is required"})
		return
	}

	var req UpdatePermissionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Call service to update permission
	if err := h.permissionService.UpdatePermission(c.Request.Context(), permissionID, &service.UpdatePermissionRequest{
		Name:        req.Name,
		Description: req.Description,
		Action:      req.Action,
	}); err != nil {
		if err.Error() == "permission not found" {
			c.JSON(http.StatusNotFound, gin.H{"error": "permission not found"})
			return
		}
		if err.Error() == "invalid permission action" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid permission action"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "permission updated successfully"})
}

// DeletePermission handles deleting a permission
func (h *PermissionHandler) DeletePermission(c *gin.Context) {
	permissionID := c.Param("id")
	if permissionID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "permission id is required"})
		return
	}

	// Call service to delete permission
	if err := h.permissionService.DeletePermission(c.Request.Context(), permissionID); err != nil {
		if err.Error() == "permission not found" {
			c.JSON(http.StatusNotFound, gin.H{"error": "permission not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "permission deleted successfully"})
}

// ListPermissions handles listing permissions with pagination
func (h *PermissionHandler) ListPermissions(c *gin.Context) {
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

	// Call service to list permissions
	permissions, total, err := h.permissionService.ListPermissions(c.Request.Context(), appID, page, size)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Build response
	response := ListPermissionsResponse{
		Items: permissions,
		Total: total,
		Page:  page,
		Size:  size,
	}

	c.JSON(http.StatusOK, response)
}

// GetPermission handles retrieving a permission by ID
func (h *PermissionHandler) GetPermission(c *gin.Context) {
	permissionID := c.Param("id")
	if permissionID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "permission id is required"})
		return
	}

	// Call service to get permission
	permission, err := h.permissionService.GetPermission(c.Request.Context(), permissionID)
	if err != nil {
		if err.Error() == "permission not found" {
			c.JSON(http.StatusNotFound, gin.H{"error": "permission not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, permission)
}

// AssignPermission handles assigning a permission to a user
func (h *PermissionHandler) AssignPermission(c *gin.Context) {
	var req AssignPermissionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Call service to assign permission
	if err := h.permissionService.AssignPermission(c.Request.Context(), &service.AssignPermissionRequest{
		AppID:        req.AppID,
		UserID:       req.UserID,
		PermissionID: req.PermissionID,
		AssignedBy:   req.AssignedBy,
	}); err != nil {
		if err.Error() == "permission not found" {
			c.JSON(http.StatusNotFound, gin.H{"error": "permission not found"})
			return
		}
		if err.Error() == "permission does not belong to this application" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "permission does not belong to this application"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "permission assigned successfully"})
}

// RevokePermission handles revoking a permission from a user
func (h *PermissionHandler) RevokePermission(c *gin.Context) {
	var req RevokePermissionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Call service to revoke permission
	if err := h.permissionService.RevokePermission(c.Request.Context(), &service.RevokePermissionRequest{
		AppID:        req.AppID,
		UserID:       req.UserID,
		PermissionID: req.PermissionID,
	}); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "permission revoked successfully"})
}

// ListUserPermissions handles listing all permissions assigned to a user for an application
func (h *PermissionHandler) ListUserPermissions(c *gin.Context) {
	// Parse query parameters
	appID := c.Query("app_id")
	if appID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "app id is required"})
		return
	}

	userID := c.Query("user_id")
	if userID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "user id is required"})
		return
	}

	// Call service to list user permissions
	permissions, err := h.permissionService.ListUserPermissions(c.Request.Context(), appID, userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, permissions)
}

// CheckPermission handles checking if a user has a specific permission for an application
func (h *PermissionHandler) CheckPermission(c *gin.Context) {
	var req CheckPermissionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Call service to check permission
	hasPermission, err := h.permissionService.CheckPermission(c.Request.Context(), req.AppID, req.UserID, req.Action)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"has_permission": hasPermission})
}

// ListPermissionsResponse represents the response for listing permissions
type ListPermissionsResponse struct {
	Items []*domain.AppPermission `json:"items"`
	Total int64                    `json:"total"`
	Page  int                      `json:"page"`
	Size  int                      `json:"size"`
}