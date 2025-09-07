package handler

import (
	"net/http"

	"cdk-office/internal/auth/service"
	"github.com/gin-gonic/gin"
)

// RoleHandlerInterface defines the interface for role handler
type RoleHandlerInterface interface {
	AssignRoleToUser(c *gin.Context)
	GetUserRoles(c *gin.Context)
}

// RoleHandler implements the RoleHandlerInterface
type RoleHandler struct {
	roleService service.RoleServiceInterface
}

// NewRoleHandler creates a new instance of RoleHandler
func NewRoleHandler() *RoleHandler {
	return &RoleHandler{
		roleService: service.NewRoleService(),
	}
}

// AssignRoleToUserRequest represents the request for assigning a role to a user
type AssignRoleToUserRequest struct {
	UserID string `json:"user_id" binding:"required"`
	RoleID string `json:"role_id" binding:"required"`
}

// AssignRoleToUser handles assigning a role to a user
func (h *RoleHandler) AssignRoleToUser(c *gin.Context) {
	var req AssignRoleToUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Call service to assign role to user
	if err := h.roleService.AssignRoleToUser(c.Request.Context(), req.UserID, req.RoleID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "role assigned to user successfully"})
}

// GetUserRoles handles getting all roles for a user
func (h *RoleHandler) GetUserRoles(c *gin.Context) {
	userID := c.Param("id")
	if userID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "user id is required"})
		return
	}

	// Call service to get user roles
	roles, err := h.roleService.GetUserRoles(c.Request.Context(), userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, roles)
}