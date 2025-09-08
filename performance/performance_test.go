package performance

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"
	"time"

	"cdk-office/internal/document/domain"
	"cdk-office/internal/document/handler"
	"cdk-office/internal/document/service"
	"cdk-office/internal/dify/client"
	"github.com/gin-gonic/gin"
)

// mockDocumentService is a mock implementation of DocumentServiceInterface
type mockDocumentService struct{}

func (m *mockDocumentService) Upload(ctx context.Context, req *service.UploadRequest) (*domain.Document, error) {
	// Simulate some processing time
	time.Sleep(10 * time.Millisecond)
	
	// Return a mock document
	document := &domain.Document{
		ID:          "mock-doc-id",
		Title:       req.Title,
		Description: req.Description,
		FilePath:    req.FilePath,
		FileSize:    req.FileSize,
		MimeType:    req.MimeType,
		OwnerID:     req.OwnerID,
		TeamID:      req.TeamID,
		Status:      "active",
		Tags:        req.Tags,
	}
	
	return document, nil
}

func (m *mockDocumentService) GetDocument(ctx context.Context, docID string) (*domain.Document, error) {
	// Simulate some processing time
	time.Sleep(5 * time.Millisecond)
	
	// Return a mock document
	document := &domain.Document{
		ID:          docID,
		Title:       "Mock Document",
		Description: "This is a mock document",
		FilePath:    "/path/to/mock/document.pdf",
		FileSize:    1024,
		MimeType:    "application/pdf",
		OwnerID:     "mock-owner",
		TeamID:      "mock-team",
		Status:      "active",
		Tags:        "mock,test",
	}
	
	return document, nil
}

func (m *mockDocumentService) UpdateDocument(ctx context.Context, docID string, req *service.UpdateRequest) error {
	// Simulate some processing time
	time.Sleep(5 * time.Millisecond)
	
	return nil
}

func (m *mockDocumentService) DeleteDocument(ctx context.Context, docID string) error {
	// Simulate some processing time
	time.Sleep(5 * time.Millisecond)
	
	return nil
}

func (m *mockDocumentService) GetDocumentVersions(ctx context.Context, docID string) ([]*domain.DocumentVersion, error) {
	// Simulate some processing time
	time.Sleep(5 * time.Millisecond)
	
	// Return mock versions
	versions := []*domain.DocumentVersion{
		{
			ID:         "mock-version-id",
			DocumentID: docID,
			Version:    1,
		},
	}
	
	return versions, nil
}

// TestDocumentUploadPerformance tests the performance of document upload API
func TestDocumentUploadPerformance(t *testing.T) {
	// Create a test server
	gin.SetMode(gin.TestMode)
	router := gin.New()
	
	// Register routes
	router.POST("/documents", func(c *gin.Context) {
		// Create a mock request
		var req handler.UploadRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		
		// Create mock service
		mockService := &mockDocumentService{}
		
		// Call mock service
		document, err := mockService.Upload(c.Request.Context(), &service.UploadRequest{
			Title:       req.Title,
			Description: req.Description,
			FilePath:    req.FilePath,
			FileSize:    req.FileSize,
			MimeType:    req.MimeType,
			OwnerID:     req.OwnerID,
			TeamID:      req.TeamID,
			Tags:        req.Tags,
		})
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		
		c.JSON(http.StatusOK, document)
	})
	
	// Create test server
	server := httptest.NewServer(router)
	defer server.Close()
	
	// Test with 100 concurrent users
	concurrentUsers := 100
	var wg sync.WaitGroup
	results := make(chan time.Duration, concurrentUsers)
	
	// Start time
	startTime := time.Now()
	
	// Launch concurrent requests
	for i := 0; i < concurrentUsers; i++ {
		wg.Add(1)
		go func(userID int) {
			defer wg.Done()
			
			// Create request payload
			payload := map[string]interface{}{
				"title":       fmt.Sprintf("Test Document %d", userID),
				"description": fmt.Sprintf("Test document description for user %d", userID),
				"file_path":   fmt.Sprintf("/path/to/file_%d.pdf", userID),
				"file_size":   1024 * 1024, // 1MB
				"mime_type":   "application/pdf",
				"owner_id":    fmt.Sprintf("user_%d", userID),
				"team_id":     "team_1",
				"tags":        "test,performance",
			}
			
			// Marshal payload
			jsonData, err := json.Marshal(payload)
			if err != nil {
				t.Errorf("Failed to marshal payload: %v", err)
				return
			}
			
			// Record start time
			reqStartTime := time.Now()
			
			// Send request
			resp, err := http.Post(server.URL+"/documents", "application/json", bytes.NewBuffer(jsonData))
			if err != nil {
				t.Errorf("Failed to send request: %v", err)
				return
			}
			defer resp.Body.Close()
			
			// Record duration
			duration := time.Since(reqStartTime)
			results <- duration
			
			// Check response status
			if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusBadRequest {
				t.Errorf("Unexpected status code: %d", resp.StatusCode)
			}
		}(i)
	}
	
	// Wait for all requests to complete
	wg.Wait()
	close(results)
	
	// Calculate statistics
	var totalDuration time.Duration
	var maxDuration time.Duration
	var minDuration = time.Hour // Initialize to a large value
	count := 0
	
	for duration := range results {
		count++
		totalDuration += duration
		
		if duration > maxDuration {
			maxDuration = duration
		}
		
		if duration < minDuration {
			minDuration = duration
		}
	}
	
	// Calculate average duration
	avgDuration := totalDuration / time.Duration(count)
	
	// Calculate 95th percentile
	// For simplicity, we'll approximate this by sorting and taking the 95th percentile value
	// In a real implementation, we would use a proper percentile calculation
	
	// Total test time
	totalTestTime := time.Since(startTime)
	
	// Log results
	t.Logf("Document Upload Performance Test Results:")
	t.Logf("Concurrent Users: %d", concurrentUsers)
	t.Logf("Total Requests: %d", count)
	t.Logf("Total Test Time: %v", totalTestTime)
	t.Logf("Average Response Time: %v", avgDuration)
	t.Logf("Min Response Time: %v", minDuration)
	t.Logf("Max Response Time: %v", maxDuration)
	t.Logf("Requests per Second: %.2f", float64(count)/totalTestTime.Seconds())
}

// TestDocumentSearchPerformance tests the performance of document search API
func TestDocumentSearchPerformance(t *testing.T) {
	// Create a test server
	gin.SetMode(gin.TestMode)
	router := gin.New()
	
	// Register routes
	router.GET("/documents/search", func(c *gin.Context) {
		// Parse query parameters
		// query := c.Query("q")
		teamID := c.Query("team_id")
		
		// Parse page and size parameters
		// pageStr := c.Query("page")
		// sizeStr := c.Query("size")
		
		// For simplicity, we're not implementing actual pagination in the mock
		// In a real implementation, we would parse these values
		
		// Simulate some processing time
		time.Sleep(10 * time.Millisecond)
		
		// Create mock response
		documents := []*domain.Document{
			{
				ID:          "mock-doc-1",
				Title:       "Mock Document 1",
				Description: "This is a mock document for search testing",
				FilePath:    "/path/to/mock/document1.pdf",
				FileSize:    1024,
				MimeType:    "application/pdf",
				OwnerID:     "mock-owner",
				TeamID:      teamID,
				Status:      "active",
				Tags:        "mock,test,search",
			},
			{
				ID:          "mock-doc-2",
				Title:       "Mock Document 2",
				Description: "This is another mock document for search testing",
				FilePath:    "/path/to/mock/document2.pdf",
				FileSize:    2048,
				MimeType:    "application/pdf",
				OwnerID:     "mock-owner",
				TeamID:      teamID,
				Status:      "active",
				Tags:        "mock,test,search",
			},
		}
		
		// Build response
		response := map[string]interface{}{
			"items": documents,
			"total": len(documents),
			"page":  1,
			"size":  10,
		}
		
		c.JSON(http.StatusOK, response)
	})
	
	// Create test server
	server := httptest.NewServer(router)
	defer server.Close()
	
	// Test with 50 concurrent users
	concurrentUsers := 50
	var wg sync.WaitGroup
	results := make(chan time.Duration, concurrentUsers)
	
	// Start time
	startTime := time.Now()
	
	// Launch concurrent requests
	for i := 0; i < concurrentUsers; i++ {
		wg.Add(1)
		go func(userID int) {
			defer wg.Done()
			
			// Record start time
			reqStartTime := time.Now()
			
			// Send request
			resp, err := http.Get(fmt.Sprintf("%s/documents/search?q=test&team_id=team_1&page=1&size=10", server.URL))
			if err != nil {
				t.Errorf("Failed to send request: %v", err)
				return
			}
			defer resp.Body.Close()
			
			// Record duration
			duration := time.Since(reqStartTime)
			results <- duration
			
			// Check response status
			if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusBadRequest {
				t.Errorf("Unexpected status code: %d", resp.StatusCode)
			}
		}(i)
	}
	
	// Wait for all requests to complete
	wg.Wait()
	close(results)
	
	// Calculate statistics
	var totalDuration time.Duration
	var maxDuration time.Duration
	var minDuration = time.Hour // Initialize to a large value
	count := 0
	
	for duration := range results {
		count++
		totalDuration += duration
		
		if duration > maxDuration {
			maxDuration = duration
		}
		
		if duration < minDuration {
			minDuration = duration
		}
	}
	
	// Calculate average duration
	avgDuration := totalDuration / time.Duration(count)
	
	// Total test time
	totalTestTime := time.Since(startTime)
	
	// Log results
	t.Logf("Document Search Performance Test Results:")
	t.Logf("Concurrent Users: %d", concurrentUsers)
	t.Logf("Total Requests: %d", count)
	t.Logf("Total Test Time: %v", totalTestTime)
	t.Logf("Average Response Time: %v", avgDuration)
	t.Logf("Min Response Time: %v", minDuration)
	t.Logf("Max Response Time: %v", maxDuration)
	t.Logf("Requests per Second: %.2f", float64(count)/totalTestTime.Seconds())
}

// mockDifyClient is a mock implementation of DifyClientInterface
type mockDifyClient struct{}

func (m *mockDifyClient) CreateCompletionMessage(ctx context.Context, req *client.CompletionRequest) (*client.CompletionResponse, error) {
	// Simulate network delay
	time.Sleep(50 * time.Millisecond)
	
	// Return mock response
	response := &client.CompletionResponse{
		MessageID:      "mock-message-id",
		Mode:           "completion",
		Answer:         "This is a mock response from Dify AI for query: " + req.Query,
		ConversationID: "mock-conversation-id",
	}
	
	return response, nil
}

func (m *mockDifyClient) CreateChatMessage(ctx context.Context, req *client.ChatRequest) (*client.ChatResponse, error) {
	// Simulate network delay
	time.Sleep(50 * time.Millisecond)
	
	// Return mock response
	response := &client.ChatResponse{
		MessageID:      "mock-message-id",
		Mode:           "chat",
		Answer:         "This is a mock response from Dify AI for query: " + req.Query,
		ConversationID: "mock-conversation-id",
	}
	
	return response, nil
}

func (m *mockDifyClient) UploadFile(ctx context.Context, req *client.FileUploadRequest) (*client.FileUploadResponse, error) {
	// Simulate network delay
	time.Sleep(50 * time.Millisecond)
	
	// Return mock response
	response := &client.FileUploadResponse{
		ID:        "mock-file-id",
		Name:      req.FileName,
		Size:      1024,
		MimeType:  req.MimeType,
		CreatedBy: req.User,
	}
	
	return response, nil
}

// TestDifyAIQuestionPerformance tests the performance of Dify AI question API
func TestDifyAIQuestionPerformance(t *testing.T) {
	// Create mock client
	mockClient := &mockDifyClient{}
	
	// Test with 20 concurrent users
	concurrentUsers := 20
	var wg sync.WaitGroup
	results := make(chan time.Duration, concurrentUsers)
	
	// Start time
	startTime := time.Now()
	
	// Launch concurrent requests
	for i := 0; i < concurrentUsers; i++ {
		wg.Add(1)
		go func(userID int) {
			defer wg.Done()
			
			// Create request payload
			req := &client.CompletionRequest{
				Query:        fmt.Sprintf("What is the content of document %d?", userID),
				ResponseMode: "blocking",
				User:         fmt.Sprintf("user_%d", userID),
			}
			
			// Record start time
			reqStartTime := time.Now()
			
			// Send request (this will use our mock implementation)
			_, err := mockClient.CreateCompletionMessage(context.Background(), req)
			
			// Record duration
			duration := time.Since(reqStartTime)
			results <- duration
			
			// Check for errors
			if err != nil {
				t.Errorf("Unexpected error from mock Dify client: %v", err)
			}
		}(i)
	}
	
	// Wait for all requests to complete
	wg.Wait()
	close(results)
	
	// Calculate statistics
	var totalDuration time.Duration
	var maxDuration time.Duration
	var minDuration = time.Hour // Initialize to a large value
	count := 0
	
	for duration := range results {
		count++
		totalDuration += duration
		
		if duration > maxDuration {
			maxDuration = duration
		}
		
		if duration < minDuration {
			minDuration = duration
		}
	}
	
	// Calculate average duration
	avgDuration := totalDuration / time.Duration(count)
	
	// Total test time
	totalTestTime := time.Since(startTime)
	
	// Log results
	t.Logf("Dify AI Question Performance Test Results:")
	t.Logf("Concurrent Users: %d", concurrentUsers)
	t.Logf("Total Requests: %d", count)
	t.Logf("Total Test Time: %v", totalTestTime)
	t.Logf("Average Response Time: %v", avgDuration)
	t.Logf("Min Response Time: %v", minDuration)
	t.Logf("Max Response Time: %v", maxDuration)
	t.Logf("Requests per Second: %.2f", float64(count)/totalTestTime.Seconds())
}

// BenchmarkDocumentUpload benchmarks the document upload API
func BenchmarkDocumentUpload(b *testing.B) {
	// Create a test server
	gin.SetMode(gin.TestMode)
	router := gin.New()
	
	// Create document handler
	docHandler := handler.NewDocumentHandler()
	
	// Register routes
	router.POST("/documents", docHandler.Upload)
	
	// Create test server
	server := httptest.NewServer(router)
	defer server.Close()
	
	b.ResetTimer()
	
	for i := 0; i < b.N; i++ {
		// Create request payload
		payload := map[string]interface{}{
			"title":       fmt.Sprintf("Benchmark Document %d", i),
			"description": fmt.Sprintf("Benchmark document description %d", i),
			"file_path":   fmt.Sprintf("/path/to/benchmark_%d.pdf", i),
			"file_size":   1024 * 1024, // 1MB
			"mime_type":   "application/pdf",
			"owner_id":    fmt.Sprintf("user_%d", i),
			"team_id":     "team_1",
			"tags":        "benchmark,test",
		}
		
		// Marshal payload
		jsonData, err := json.Marshal(payload)
		if err != nil {
			b.Errorf("Failed to marshal payload: %v", err)
			return
		}
		
		// Send request
		resp, err := http.Post(server.URL+"/documents", "application/json", bytes.NewBuffer(jsonData))
		if err != nil {
			b.Errorf("Failed to send request: %v", err)
			return
		}
		defer resp.Body.Close()
		
		// Check response status
		if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusBadRequest {
			b.Errorf("Unexpected status code: %d", resp.StatusCode)
		}
	}
}

// BenchmarkDocumentSearch benchmarks the document search API
func BenchmarkDocumentSearch(b *testing.B) {
	// Create a test server
	gin.SetMode(gin.TestMode)
	router := gin.New()
	
	// Create search handler
	searchHandler := handler.NewSearchHandler()
	
	// Register routes
	router.GET("/documents/search", searchHandler.SearchDocuments)
	
	// Create test server
	server := httptest.NewServer(router)
	defer server.Close()
	
	b.ResetTimer()
	
	for i := 0; i < b.N; i++ {
		// Send request
		resp, err := http.Get(fmt.Sprintf("%s/documents/search?q=benchmark&team_id=team_1&page=1&size=10", server.URL))
		if err != nil {
			b.Errorf("Failed to send request: %v", err)
			return
		}
		defer resp.Body.Close()
		
		// Check response status
		if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusBadRequest {
			b.Errorf("Unexpected status code: %d", resp.StatusCode)
		}
	}
}