package middleware

import (
	"net/http"
	"strings"

	"cdk-office/pkg/jwt"

	"github.com/gin-gonic/gin"
)

// AuthMiddlewareInterface defines the interface for authentication middleware
type AuthMiddlewareInterface interface {
	Authenticate() gin.HandlerFunc
}

// AuthMiddleware implements the AuthMiddlewareInterface
type AuthMiddleware struct {
	jwtManager      *jwt.JWTManager
	tokenBlacklist  *jwt.TokenBlacklist
}

// NewAuthMiddleware creates a new instance of AuthMiddleware
func NewAuthMiddleware(jwtManager *jwt.JWTManager) *AuthMiddleware {
	return &AuthMiddleware{
		jwtManager:     jwtManager,
		tokenBlacklist: jwt.NewTokenBlacklist(),
	}
}

// Authenticate returns a Gin middleware function for JWT authentication
func (m *AuthMiddleware) Authenticate() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Get the Authorization header
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "authorization header is required"})
			c.Abort()
			return
		}

		// Check if the header starts with "Bearer "
		if !strings.HasPrefix(authHeader, "Bearer ") {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "authorization header must start with Bearer"})
			c.Abort()
			return
		}

		// Extract the token
		tokenString := strings.TrimPrefix(authHeader, "Bearer ")

		// Check if token is blacklisted
		isBlacklisted, err := m.tokenBlacklist.IsBlacklisted(tokenString)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to check token status"})
			c.Abort()
			return
		}

		if isBlacklisted {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "token is invalid"})
			c.Abort()
			return
		}

		// Verify the token
		claims, err := m.jwtManager.VerifyToken(tokenString)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid token"})
			c.Abort()
			return
		}

		// Set user information in the context
		c.Set("user_id", claims.UserID)
		c.Set("username", claims.Username)
		c.Set("role", claims.Role)

		// Continue with the next handler
		c.Next()
	}
}