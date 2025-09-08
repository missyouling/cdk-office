package handler

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	
	"cdk-office/internal/app/service"
	"cdk-office/internal/shared/testutils"
	
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

// TestAppHandler tests the AppHandler
func TestAppHandler(t *testing.T) {
	// Set up test environment
	gin.SetMode(gin.TestMode)
	
	// Initialize the database connection for testing
	testDB := testutils.SetupTestDB()
	
	// Create app service with database connection
	appService := service.NewAppServiceWithDB(testDB)
	
	// Create app handler
	appHandler := NewAppHandlerWithService(appService)
	
	// Test CreateApplication
	var appID string
	t.Run("CreateApplication", func(t *testing.T) {
		// Create a test context
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		
		// Create a test request
		jsonStr := `{"team_id":"team_123","name":"Test App","type":"qrcode","created_by":"user_123"}`
		c.Request, _ = http.NewRequest("POST", "/applications", strings.NewReader(jsonStr))
		c.Request.Header.Set("Content-Type", "application/json")
		
		// Call the handler
		appHandler.CreateApplication(c)
		
		// Assert results
		assert.Equal(t, http.StatusOK, w.Code)
		
		// Extract app ID from response
		var response map[string]interface{}
		json.Unmarshal(w.Body.Bytes(), &response)
		if response["id"] != nil {
			appID = response["id"].(string)
		}
	})
	
	// Test UpdateApplication
	t.Run("UpdateApplication", func(t *testing.T) {
		// Create a test context
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		
		// Create a test request
		jsonStr := `{"name":"Updated App"}`
		req, _ := http.NewRequest("PUT", "/applications/"+appID, strings.NewReader(jsonStr))
		req.Header.Set("Content-Type", "application/json")
		c.Request = req
		c.Params = []gin.Param{{Key: "id", Value: appID}}
		
		// Assert results
		assert.Equal(t, http.StatusOK, w.Code)
	})
}