package middleware

import (
	"github.com/gin-gonic/gin"
)

// CORSMiddlewareInterface defines the interface for CORS middleware
type CORSMiddlewareInterface interface {
	CORS() gin.HandlerFunc
}

// CORSMiddleware implements the CORSMiddlewareInterface
type CORSMiddleware struct{}

// NewCORSMiddleware creates a new instance of CORSMiddleware
func NewCORSMiddleware() *CORSMiddleware {
	return &CORSMiddleware{}
}

// CORS returns a Gin middleware function for handling CORS
func (m *CORSMiddleware) CORS() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Set CORS headers
		c.Header("Access-Control-Allow-Origin", "*")
		c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		c.Header("Access-Control-Allow-Headers", "Origin, Content-Type, Accept, Authorization, X-Requested-With")
		c.Header("Access-Control-Allow-Credentials", "true")
		c.Header("Access-Control-Max-Age", "86400")

		// Handle preflight requests
		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}

		// Continue with the next handler
		c.Next()
	}
}