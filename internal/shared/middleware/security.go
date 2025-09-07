package middleware

import (
	"github.com/gin-gonic/gin"
)

// SecurityHeadersMiddlewareInterface defines the interface for security headers middleware
type SecurityHeadersMiddlewareInterface interface {
	SecurityHeaders() gin.HandlerFunc
}

// SecurityHeadersMiddleware implements the SecurityHeadersMiddlewareInterface
type SecurityHeadersMiddleware struct{}

// NewSecurityHeadersMiddleware creates a new instance of SecurityHeadersMiddleware
func NewSecurityHeadersMiddleware() *SecurityHeadersMiddleware {
	return &SecurityHeadersMiddleware{}
}

// SecurityHeaders returns a Gin middleware function for setting security headers
func (m *SecurityHeadersMiddleware) SecurityHeaders() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Set security headers
		c.Header("X-Content-Type-Options", "nosniff")
		c.Header("X-Frame-Options", "DENY")
		c.Header("X-XSS-Protection", "1; mode=block")
		c.Header("Strict-Transport-Security", "max-age=31536000; includeSubDomains")
		c.Header("Referrer-Policy", "no-referrer")
		c.Header("Permissions-Policy", "geolocation=(), microphone=(), camera=()")

		// Continue with the next handler
		c.Next()
	}
}