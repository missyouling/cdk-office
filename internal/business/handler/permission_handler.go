package handler

import (
	"net/http"

	"cdk-office/internal/business/service"
	"github.com/gin-gonic/gin"
)

// BusinessPermissionHandlerInterface defines the interface for business module permission handler
type BusinessPermissionHandlerInterface interface {
	AssignPermissionToRole(c *gin.Context)
	RevokePermissionFromRole(c *gin.Context)
	ListRolePermissions(c *gin.Context)
	CheckRolePermission(c *gin.Context)
	ListModulePermissions(c *gin.Context)
}

// BusinessPermissionHandler implements the BusinessPermissionHandlerInterface
type BusinessPermissionHandler struct {
	permissionService service.BusinessPermissionServiceInterface
}

// NewBusinessPermissionHandler creates a new instance of BusinessPermissionHandler
func NewBusinessPermissionHandler() *BusinessPermissionHandler {
	return &BusinessPermissionHandler{
		permissionService: service.NewBusinessPermissionService(),
	}
}

// AssignPermissionToRoleRequest represents the request for assigning a permission to a role
type AssignPermissionToRoleRequest struct {
	ModuleID   string `json:"module_id" binding:"required"`
	RoleID     string `json:"role_id" binding:"required"`
	Permission string `json:"permission" binding:"required"`
}

// RevokePermissionFromRoleRequest represents the request for revoking a permission from a role
type RevokePermissionFromRoleRequest struct {
	ModuleID   string `json:"module_id" binding:"required"`
	RoleID     string `json:"role_id" binding:"required"`
	Permission string `json:"permission" binding:"required"`
}

// CheckRolePermissionRequest represents the request for checking a role permission
type CheckRolePermissionRequest struct {
	RoleID     string `json:"role_id" binding:"required"`
	ModuleID   string `json:"module_id" binding:"required"`
	Permission string `json:"permission" binding:"required"`
}

// AssignPermissionToRole handles assigning a permission to a role
func (h *BusinessPermissionHandler) AssignPermissionToRole(c *gin.Context) {
	var req AssignPermissionToRoleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Call service to assign permission to role
	if err := h.permissionService.AssignPermissionToRole(c.Request.Context(), &service.AssignPermissionToRoleRequest{
		ModuleID:   req.ModuleID,
		RoleID:     req.RoleID,
		Permission: req.Permission,
	}); err != nil {
		if err.Error() == "module not found" {
			c.JSON(http.StatusNotFound, gin.H{"error": "module not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "permission assigned to role successfully"})
}

// RevokePermissionFromRole handles revoking a permission from a role
func (h *BusinessPermissionHandler) RevokePermissionFromRole(c *gin.Context) {
	var req RevokePermissionFromRoleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Call service to revoke permission from role
	if err := h.permissionService.RevokePermissionFromRole(c.Request.Context(), &service.RevokePermissionFromRoleRequest{
		ModuleID:   req.ModuleID,
		RoleID:     req.RoleID,
		Permission: req.Permission,
	}); err != nil {
		if err.Error() == "permission not found for this role and module" {
			c.JSON(http.StatusNotFound, gin.H{"error": "permission not found for this role and module"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "permission revoked from role successfully"})
}

// ListRolePermissions handles listing all permissions for a role
func (h *BusinessPermissionHandler) ListRolePermissions(c *gin.Context) {
	roleID := c.Query("role_id")
	if roleID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "role id is required"})
		return
	}

	// Call service to list role permissions
	permissions, err := h.permissionService.ListRolePermissions(c.Request.Context(), roleID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, permissions)
}

// CheckRolePermission handles checking if a role has a specific permission
func (h *BusinessPermissionHandler) CheckRolePermission(c *gin.Context) {
	var req CheckRolePermissionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Call service to check role permission
	hasPermission, err := h.permissionService.CheckRolePermission(c.Request.Context(), req.RoleID, req.ModuleID, req.Permission)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"has_permission": hasPermission})
}

// ListModulePermissions handles listing all permissions for a module
func (h *BusinessPermissionHandler) ListModulePermissions(c *gin.Context) {
	moduleID := c.Query("module_id")
	if moduleID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "module id is required"})
		return
	}

	// Call service to list module permissions
	permissions, err := h.permissionService.ListModulePermissions(c.Request.Context(), moduleID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, permissions)
}