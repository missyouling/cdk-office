package handler

import (
	"net/http"
	"strconv"

	"cdk-office/internal/app/domain"
	"cdk-office/internal/app/service"
	"github.com/gin-gonic/gin"
)

// AppPermissionHandlerInterface defines the interface for application permission handler
type AppPermissionHandlerInterface interface {
	CreateAppPermission(c *gin.Context)
	UpdateAppPermission(c *gin.Context)
	DeleteAppPermission(c *gin.Context)
	ListAppPermissions(c *gin.Context)
	GetAppPermission(c *gin.Context)
	AssignPermissionToUser(c *gin.Context)
	RevokePermissionFromUser(c *gin.Context)
	ListUserPermissions(c *gin.Context)
	CheckUserPermission(c *gin.Context)
}

// AppPermissionHandler implements the AppPermissionHandlerInterface
type AppPermissionHandler struct {
	permissionService service.AppPermissionServiceInterface
}

// NewAppPermissionHandler creates a new instance of AppPermissionHandler
func NewAppPermissionHandler() *AppPermissionHandler {
	return &AppPermissionHandler{
		permissionService: service.NewAppPermissionService(),
	}
}

// CreateAppPermissionRequest represents the request for creating an application permission
type CreateAppPermissionRequest struct {
	AppID       string `json:"app_id" binding:"required"`
	Name        string `json:"name" binding:"required"`
	Description string `json:"description"`
	Permission  string `json:"permission" binding:"required"` // read, write, delete, manage
	CreatedBy   string `json:"created_by" binding:"required"`
}

// UpdateAppPermissionRequest represents the request for updating an application permission
type UpdateAppPermissionRequest struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	Permission  string `json:"permission"` // read, write, delete, manage
}

// AssignPermissionToUserRequest represents the request for assigning a permission to a user
type AssignPermissionToUserRequest struct {
	AppID        string `json:"app_id" binding:"required"`
	UserID       string `json:"user_id" binding:"required"`
	PermissionID string `json:"permission_id" binding:"required"`
	AssignedBy   string `json:"assigned_by" binding:"required"`
}

// RevokePermissionFromUserRequest represents the request for revoking a permission from a user
type RevokePermissionFromUserRequest struct {
	AppID        string `json:"app_id" binding:"required"`
	UserID       string `json:"user_id" binding:"required"`
	PermissionID string `json:"permission_id" binding:"required"`
}

// CheckUserPermissionRequest represents the request for checking a user's permission
type CheckUserPermissionRequest struct {
	AppID      string `json:"app_id" binding:"required"`
	UserID     string `json:"user_id" binding:"required"`
	Permission string `json:"permission" binding:"required"` // read, write, delete, manage
}

// ListAppPermissionsRequest represents the request for listing application permissions
type ListAppPermissionsRequest struct {
	AppID string `form:"app_id" binding:"required"`
	Page  int    `form:"page"`
	Size  int    `form:"size"`
}

// ListUserPermissionsRequest represents the request for listing user permissions
type ListUserPermissionsRequest struct {
	AppID  string `form:"app_id" binding:"required"`
	UserID string `form:"user_id" binding:"required"`
}

// CreateAppPermission handles creating a new application permission
func (h *AppPermissionHandler) CreateAppPermission(c *gin.Context) {
	var req CreateAppPermissionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Call service to create application permission
	permission, err := h.permissionService.CreateAppPermission(c.Request.Context(), &service.CreateAppPermissionRequest{
		AppID:       req.AppID,
		Name:        req.Name,
		Description: req.Description,
		Permission:  req.Permission,
		CreatedBy:   req.CreatedBy,
	})
	if err != nil {
		if err.Error() == "invalid permission" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid permission"})
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

// UpdateAppPermission handles updating an existing application permission
func (h *AppPermissionHandler) UpdateAppPermission(c *gin.Context) {
	permissionID := c.Param("id")
	if permissionID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "permission id is required"})
		return
	}

	var req UpdateAppPermissionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Call service to update application permission
	if err := h.permissionService.UpdateAppPermission(c.Request.Context(), permissionID, &service.UpdateAppPermissionRequest{
		Name:        req.Name,
		Description: req.Description,
		Permission:  req.Permission,
	}); err != nil {
		if err.Error() == "application permission not found" {
			c.JSON(http.StatusNotFound, gin.H{"error": "application permission not found"})
			return
		}
		if err.Error() == "invalid permission" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid permission"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "application permission updated successfully"})
}

// DeleteAppPermission handles deleting an application permission
func (h *AppPermissionHandler) DeleteAppPermission(c *gin.Context) {
	permissionID := c.Param("id")
	if permissionID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "permission id is required"})
		return
	}

	// Call service to delete application permission
	if err := h.permissionService.DeleteAppPermission(c.Request.Context(), permissionID); err != nil {
		if err.Error() == "application permission not found" {
			c.JSON(http.StatusNotFound, gin.H{"error": "application permission not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "application permission deleted successfully"})
}

// ListAppPermissions handles listing application permissions with pagination
func (h *AppPermissionHandler) ListAppPermissions(c *gin.Context) {
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

	// Call service to list application permissions
	permissions, total, err := h.permissionService.ListAppPermissions(c.Request.Context(), appID, page, size)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Build response
	response := ListAppPermissionsResponse{
		Items: permissions,
		Total: total,
		Page:  page,
		Size:  size,
	}

	c.JSON(http.StatusOK, response)
}

// GetAppPermission handles retrieving an application permission by ID
func (h *AppPermissionHandler) GetAppPermission(c *gin.Context) {
	permissionID := c.Param("id")
	if permissionID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "permission id is required"})
		return
	}

	// Call service to get application permission
	permission, err := h.permissionService.GetAppPermission(c.Request.Context(), permissionID)
	if err != nil {
		if err.Error() == "application permission not found" {
			c.JSON(http.StatusNotFound, gin.H{"error": "application permission not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, permission)
}

// AssignPermissionToUser handles assigning a permission to a user
func (h *AppPermissionHandler) AssignPermissionToUser(c *gin.Context) {
	var req AssignPermissionToUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Call service to assign permission to user
	if err := h.permissionService.AssignPermissionToUser(c.Request.Context(), &service.AssignPermissionToUserRequest{
		AppID:        req.AppID,
		UserID:       req.UserID,
		PermissionID: req.PermissionID,
		AssignedBy:   req.AssignedBy,
	}); err != nil {
		if err.Error() == "application permission not found" {
			c.JSON(http.StatusNotFound, gin.H{"error": "application permission not found"})
			return
		}
		if err.Error() == "application permission does not belong to the specified application" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "application permission does not belong to the specified application"})
			return
		}
		if err.Error() == "permission already assigned to user" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "permission already assigned to user"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "permission assigned to user successfully"})
}

// RevokePermissionFromUser handles revoking a permission from a user
func (h *AppPermissionHandler) RevokePermissionFromUser(c *gin.Context) {
	var req RevokePermissionFromUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Call service to revoke permission from user
	if err := h.permissionService.RevokePermissionFromUser(c.Request.Context(), &service.RevokePermissionFromUserRequest{
		AppID:        req.AppID,
		UserID:       req.UserID,
		PermissionID: req.PermissionID,
	}); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "permission revoked from user successfully"})
}

// ListUserPermissions handles listing all permissions assigned to a user for an application
func (h *AppPermissionHandler) ListUserPermissions(c *gin.Context) {
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

// CheckUserPermission handles checking if a user has a specific permission for an application
func (h *AppPermissionHandler) CheckUserPermission(c *gin.Context) {
	var req CheckUserPermissionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Call service to check user permission
	hasPermission, err := h.permissionService.CheckUserPermission(c.Request.Context(), req.AppID, req.UserID, req.Permission)
	if err != nil {
		if err.Error() == "invalid permission" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid permission"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"has_permission": hasPermission})
}

// ListAppPermissionsResponse represents the response for listing application permissions
type ListAppPermissionsResponse struct {
	Items []*domain.AppPermission `json:"items"`
	Total int64                    `json:"total"`
	Page  int                      `json:"page"`
	Size  int                      `json:"size"`
}