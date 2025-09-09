package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func TestSecurityHeadersMiddleware_SecurityHeaders(t *testing.T) {
	// Setup
	gin.SetMode(gin.TestMode)
	router := gin.New()
	
	// Create middleware
	securityMiddleware := NewSecurityHeadersMiddleware()
	
	// Add middleware to router
	router.Use(securityMiddleware.SecurityHeaders())
	
	// Create a test handler
	router.GET("/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "test"})
	})
	
	// Create test request
	req := httptest.NewRequest("GET", "/test", nil)
	w := httptest.NewRecorder()
	
	// Perform request
	router.ServeHTTP(w, req)
	
	// Assertions
	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), "test")
	
	// Check security headers
	assert.Equal(t, "nosniff", w.Header().Get("X-Content-Type-Options"))
	assert.Equal(t, "DENY", w.Header().Get("X-Frame-Options"))
	assert.Equal(t, "1; mode=block", w.Header().Get("X-XSS-Protection"))
	assert.Equal(t, "max-age=31536000; includeSubDomains", w.Header().Get("Strict-Transport-Security"))
	assert.Equal(t, "no-referrer", w.Header().Get("Referrer-Policy"))
	assert.Equal(t, "geolocation=(), microphone=(), camera=()", w.Header().Get("Permissions-Policy"))
}

func TestSecurityHeadersMiddleware_SecurityHeaders_PostRequest(t *testing.T) {
	// Setup
	gin.SetMode(gin.TestMode)
	router := gin.New()
	
	// Create middleware
	securityMiddleware := NewSecurityHeadersMiddleware()
	
	// Add middleware to router
	router.Use(securityMiddleware.SecurityHeaders())
	
	// Create a test handler
	router.POST("/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "test post"})
	})
	
	// Create POST request
	req := httptest.NewRequest("POST", "/test", nil)
	w := httptest.NewRecorder()
	
	// Perform request
	router.ServeHTTP(w, req)
	
	// Assertions
	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), "test post")
	
	// Check security headers
	assert.Equal(t, "nosniff", w.Header().Get("X-Content-Type-Options"))
	assert.Equal(t, "DENY", w.Header().Get("X-Frame-Options"))
	assert.Equal(t, "1; mode=block", w.Header().Get("X-XSS-Protection"))
	assert.Equal(t, "max-age=31536000; includeSubDomains", w.Header().Get("Strict-Transport-Security"))
	assert.Equal(t, "no-referrer", w.Header().Get("Referrer-Policy"))
	assert.Equal(t, "geolocation=(), microphone=(), camera=()", w.Header().Get("Permissions-Policy"))
}