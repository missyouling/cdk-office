package handler

import (
	"net/http"

	"cdk-office/internal/auth/service"
	"github.com/gin-gonic/gin"
)

// PermissionHandlerInterface defines the interface for permission handler
type PermissionHandlerInterface interface {
	CreatePermission(c *gin.Context)
	CreateRole(c *gin.Context)
	AssignPermissionToRole(c *gin.Context)
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
	Name        string `json:"name" binding:"required"`
	Resource    string `json:"resource" binding:"required"`
	Action      string `json:"action" binding:"required"`
	Description string `json:"description"`
}

// CreateRoleRequest represents the request for creating a role
type CreateRoleRequest struct {
	Name        string `json:"name" binding:"required"`
	Description string `json:"description"`
}

// AssignPermissionToRoleRequest represents the request for assigning a permission to a role
type AssignPermissionToRoleRequest struct {
	RoleID       string `json:"role_id" binding:"required"`
	PermissionID string `json:"permission_id" binding:"required"`
}

// CheckPermissionRequest represents the request for checking a permission
type CheckPermissionRequest struct {
	UserID   string `json:"user_id" binding:"required"`
	Resource string `json:"resource" binding:"required"`
	Action   string `json:"action" binding:"required"`
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
		Name:        req.Name,
		Resource:    req.Resource,
		Action:      req.Action,
		Description: req.Description,
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, permission)
}

// CreateRole handles creating a new role
func (h *PermissionHandler) CreateRole(c *gin.Context) {
	var req CreateRoleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Call service to create role
	role, err := h.permissionService.CreateRole(c.Request.Context(), req.Name, req.Description)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, role)
}

// AssignPermissionToRole handles assigning a permission to a role
func (h *PermissionHandler) AssignPermissionToRole(c *gin.Context) {
	var req AssignPermissionToRoleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Call service to assign permission to role
	if err := h.permissionService.AssignPermissionToRole(c.Request.Context(), req.RoleID, req.PermissionID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "permission assigned to role successfully"})
}

// CheckPermission handles checking if a user has a specific permission
func (h *PermissionHandler) CheckPermission(c *gin.Context) {
	var req CheckPermissionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Call service to check permission
	hasPermission, err := h.permissionService.CheckPermission(c.Request.Context(), req.UserID, req.Resource, req.Action)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"has_permission": hasPermission})
}