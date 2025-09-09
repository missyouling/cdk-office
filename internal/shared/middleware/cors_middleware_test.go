package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func TestCORSMiddleware_CORS_OptionsRequest(t *testing.T) {
	// Setup
	gin.SetMode(gin.TestMode)
	router := gin.New()
	
	// Create middleware
	corsMiddleware := NewCORSMiddleware()
	
	// Add middleware to router
	router.Use(corsMiddleware.CORS())
	
	// Create a test handler
	router.GET("/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "test"})
	})
	
	// Create OPTIONS request
	req := httptest.NewRequest("OPTIONS", "/test", nil)
	w := httptest.NewRecorder()
	
	// Perform request
	router.ServeHTTP(w, req)
	
	// Assertions
	assert.Equal(t, http.StatusNoContent, w.Code)
	
	// Check CORS headers
	assert.Equal(t, "*", w.Header().Get("Access-Control-Allow-Origin"))
	assert.Equal(t, "GET, POST, PUT, DELETE, OPTIONS", w.Header().Get("Access-Control-Allow-Methods"))
	assert.Equal(t, "Origin, Content-Type, Accept, Authorization, X-Requested-With", w.Header().Get("Access-Control-Allow-Headers"))
	assert.Equal(t, "true", w.Header().Get("Access-Control-Allow-Credentials"))
	assert.Equal(t, "86400", w.Header().Get("Access-Control-Max-Age"))
}

func TestCORSMiddleware_CORS_GetRequest(t *testing.T) {
	// Setup
	gin.SetMode(gin.TestMode)
	router := gin.New()
	
	// Create middleware
	corsMiddleware := NewCORSMiddleware()
	
	// Add middleware to router
	router.Use(corsMiddleware.CORS())
	
	// Create a test handler
	router.GET("/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "test"})
	})
	
	// Create GET request
	req := httptest.NewRequest("GET", "/test", nil)
	w := httptest.NewRecorder()
	
	// Perform request
	router.ServeHTTP(w, req)
	
	// Assertions
	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), "test")
	
	// Check CORS headers
	assert.Equal(t, "*", w.Header().Get("Access-Control-Allow-Origin"))
	assert.Equal(t, "GET, POST, PUT, DELETE, OPTIONS", w.Header().Get("Access-Control-Allow-Methods"))
	assert.Equal(t, "Origin, Content-Type, Accept, Authorization, X-Requested-With", w.Header().Get("Access-Control-Allow-Headers"))
	assert.Equal(t, "true", w.Header().Get("Access-Control-Allow-Credentials"))
	assert.Equal(t, "86400", w.Header().Get("Access-Control-Max-Age"))
}

func TestCORSMiddleware_CORS_PostRequest(t *testing.T) {
	// Setup
	gin.SetMode(gin.TestMode)
	router := gin.New()
	
	// Create middleware
	corsMiddleware := NewCORSMiddleware()
	
	// Add middleware to router
	router.Use(corsMiddleware.CORS())
	
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
	
	// Check CORS headers
	assert.Equal(t, "*", w.Header().Get("Access-Control-Allow-Origin"))
	assert.Equal(t, "GET, POST, PUT, DELETE, OPTIONS", w.Header().Get("Access-Control-Allow-Methods"))
	assert.Equal(t, "Origin, Content-Type, Accept, Authorization, X-Requested-With", w.Header().Get("Access-Control-Allow-Headers"))
	assert.Equal(t, "true", w.Header().Get("Access-Control-Allow-Credentials"))
	assert.Equal(t, "86400", w.Header().Get("Access-Control-Max-Age"))
}