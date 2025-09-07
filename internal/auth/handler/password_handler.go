package handler

import (
	"net/http"

	"cdk-office/internal/auth/service"
	"github.com/gin-gonic/gin"
)

// PasswordHandlerInterface defines the interface for password authentication handler
type PasswordHandlerInterface interface {
	PasswordLogin(c *gin.Context)
	ChangePassword(c *gin.Context)
}

// PasswordHandler implements the PasswordHandlerInterface
type PasswordHandler struct {
	passwordService service.PasswordServiceInterface
}

// NewPasswordHandler creates a new instance of PasswordHandler
func NewPasswordHandler() *PasswordHandler {
	return &PasswordHandler{
		passwordService: service.NewPasswordService(),
	}
}

// PasswordLoginRequest represents the request for password login
type PasswordLoginRequest struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
}

// ChangePasswordRequest represents the request for changing password
type ChangePasswordRequest struct {
	OldPassword string `json:"old_password" binding:"required"`
	NewPassword string `json:"new_password" binding:"required,min=6"`
}

// PasswordLogin handles password login
func (h *PasswordHandler) PasswordLogin(c *gin.Context) {
	var req PasswordLoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Call service to login user via password
	resp, err := h.passwordService.PasswordLogin(c.Request.Context(), &service.PasswordLoginRequest{
		Username: req.Username,
		Password: req.Password,
	})
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, resp)
}

// ChangePassword handles changing password
func (h *PasswordHandler) ChangePassword(c *gin.Context) {
	userID := c.GetString("user_id") // Assuming user ID is set in context after authentication
	if userID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	var req ChangePasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Call service to change password
	if err := h.passwordService.ChangePassword(c.Request.Context(), userID, req.OldPassword, req.NewPassword); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "password changed successfully"})
}