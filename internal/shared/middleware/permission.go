package middleware

import (
	"net/http"

	"cdk-office/internal/auth/service"
	"cdk-office/pkg/jwt"

	"github.com/gin-gonic/gin"
)

// PermissionMiddlewareInterface defines the interface for permission middleware
type PermissionMiddlewareInterface interface {
	CheckPermission(resource, action string) gin.HandlerFunc
}

// PermissionMiddleware implements the PermissionMiddlewareInterface
type PermissionMiddleware struct {
	jwtManager      *jwt.JWTManager
	permissionService service.PermissionServiceInterface
	tokenBlacklist  *jwt.TokenBlacklist
}

// NewPermissionMiddleware creates a new instance of PermissionMiddleware
func NewPermissionMiddleware(jwtManager *jwt.JWTManager, permissionService service.PermissionServiceInterface) *PermissionMiddleware {
	return &PermissionMiddleware{
		jwtManager:      jwtManager,
		permissionService: permissionService,
		tokenBlacklist:  jwt.NewTokenBlacklist(),
	}
}

// CheckPermission returns a Gin middleware function for permission checking
func (m *PermissionMiddleware) CheckPermission(resource, action string) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Get user ID from context (set by auth middleware)
		userID, exists := c.Get("user_id")
		if !exists {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "user not authenticated"})
			c.Abort()
			return
		}

		// Convert userID to string
		userIDStr, ok := userID.(string)
		if !ok {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "invalid user ID type"})
			c.Abort()
			return
		}

		// Check permission
		hasPermission, err := m.permissionService.CheckPermission(c.Request.Context(), userIDStr, resource, action)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to check permission"})
			c.Abort()
			return
		}

		if !hasPermission {
			c.JSON(http.StatusForbidden, gin.H{"error": "insufficient permissions"})
			c.Abort()
			return
		}

		// Continue with the next handler
		c.Next()
	}
}