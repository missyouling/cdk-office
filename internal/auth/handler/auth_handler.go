package handler

import (
	"net/http"

	"cdk-office/internal/auth/service"
	"cdk-office/pkg/jwt"
	"github.com/gin-gonic/gin"
)

// AuthHandlerInterface defines the interface for authentication handler
type AuthHandlerInterface interface {
	Register(c *gin.Context)
	Login(c *gin.Context)
	GetUserInfo(c *gin.Context)
	Logout(c *gin.Context)
	RefreshToken(c *gin.Context)
}

// AuthHandler implements the AuthHandlerInterface
type AuthHandler struct {
	authService service.AuthServiceInterface
}

// NewAuthHandler creates a new instance of AuthHandler
func NewAuthHandler(jwtManager *jwt.JWTManager) *AuthHandler {
	return &AuthHandler{
		authService: service.NewAuthService(jwtManager),
	}
}

// RegisterRequest represents the request for user registration
type RegisterRequest struct {
	Username string `json:"username" binding:"required"`
	Email    string `json:"email" binding:"required,email"`
	Phone    string `json:"phone"`
	Password string `json:"password" binding:"required,min=6"`
	RealName string `json:"real_name"`
	IDCard   string `json:"id_card"`
}

// LoginRequest represents the request for user login
type LoginRequest struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
}

// Register handles user registration
func (h *AuthHandler) Register(c *gin.Context) {
	var req RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Call service to register user
	if err := h.authService.Register(c.Request.Context(), &service.RegisterRequest{
		Username: req.Username,
		Email:    req.Email,
		Phone:    req.Phone,
		Password: req.Password,
		RealName: req.RealName,
		IDCard:   req.IDCard,
	}); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "user registered successfully"})
}

// Login handles user login
func (h *AuthHandler) Login(c *gin.Context) {
	var req LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Call service to login user
	resp, err := h.authService.Login(c.Request.Context(), &service.LoginRequest{
		Username: req.Username,
		Password: req.Password,
	})
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, resp)
}

// GetUserInfo handles getting user information
func (h *AuthHandler) GetUserInfo(c *gin.Context) {
	userID := c.Param("id")
	if userID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "user id is required"})
		return
	}

	// Call service to get user info
	user, err := h.authService.GetUserInfo(c.Request.Context(), userID)
	if err != nil {
		if err.Error() == "user not found" {
			c.JSON(http.StatusNotFound, gin.H{"error": "user not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, user)
}

// LogoutRequest represents the request for user logout
type LogoutRequest struct {
	Token string `json:"token" binding:"required"`
}

// Logout handles user logout
func (h *AuthHandler) Logout(c *gin.Context) {
	var req LogoutRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Call service to logout user
	if err := h.authService.Logout(c.Request.Context(), req.Token); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "logout successful"})
}

// RefreshTokenRequest represents the request for token refresh
type RefreshTokenRequest struct {
	RefreshToken string `json:"refresh_token" binding:"required"`
}

// RefreshToken handles token refresh
func (h *AuthHandler) RefreshToken(c *gin.Context) {
	var req RefreshTokenRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Call service to refresh token
	resp, err := h.authService.RefreshToken(c.Request.Context(), req.RefreshToken)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, resp)
}