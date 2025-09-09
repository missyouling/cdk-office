package client

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

// TestDifyClient tests the DifyClient
func TestDifyClient(t *testing.T) {
	// Create Dify client
	difyClient := NewDifyClient("http://localhost:8000", "test_api_key")

	// Test CreateCompletionMessage success
	t.Run("CreateCompletionMessageSuccess", func(t *testing.T) {
		// Create mock HTTP server
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Verify request
			assert.Equal(t, "POST", r.Method)
			assert.Equal(t, "/completion-messages", r.URL.Path)
			assert.Equal(t, "Bearer test_api_key", r.Header.Get("Authorization"))
			assert.Equal(t, "application/json", r.Header.Get("Content-Type"))

			// Send mock response
			response := CompletionResponse{
				MessageID:      "msg_123",
				Mode:           "completion",
				Answer:         "Hello! How can I help you?",
				Metadata:       Metadata{Usage: Usage{PromptTokens: 10, CompletionTokens: 20, TotalTokens: 30}},
				CreatedAt:      time.Now(),
				ConversationID: "conv_123",
			}
			json.NewEncoder(w).Encode(response)
		}))
		defer server.Close()

		// Update client to use mock server
		difyClient.baseURL = server.URL

		// Prepare test data
		ctx := context.Background()
		req := &CompletionRequest{
			Query:        "Hello, world!",
			ResponseMode: "blocking",
			User:         "test_user",
		}

		// Call the method under test
		resp, err := difyClient.CreateCompletionMessage(ctx, req)

		// Assert results
		assert.NoError(t, err)
		assert.NotNil(t, resp)
		assert.Equal(t, "msg_123", resp.MessageID)
		assert.Equal(t, "Hello! How can I help you?", resp.Answer)
	})

	// Test CreateCompletionMessage with HTTP error
	t.Run("CreateCompletionMessageHTTPError", func(t *testing.T) {
		// Create mock HTTP server that returns error
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		}))
		defer server.Close()

		// Update client to use mock server
		difyClient.baseURL = server.URL

		// Prepare test data
		ctx := context.Background()
		req := &CompletionRequest{
			Query:        "Hello, world!",
			ResponseMode: "blocking",
			User:         "test_user",
		}

		// Call the method under test
		resp, err := difyClient.CreateCompletionMessage(ctx, req)

		// Assert results
		assert.Error(t, err)
		assert.Nil(t, resp)
		assert.Equal(t, "failed to create completion message", err.Error())
	})

	// Test CreateChatMessage success
	t.Run("CreateChatMessageSuccess", func(t *testing.T) {
		// Create mock HTTP server
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Verify request
			assert.Equal(t, "POST", r.Method)
			assert.Equal(t, "/chat-messages", r.URL.Path)
			assert.Equal(t, "Bearer test_api_key", r.Header.Get("Authorization"))
			assert.Equal(t, "application/json", r.Header.Get("Content-Type"))

			// Send mock response
			response := ChatResponse{
				MessageID:      "msg_456",
				Mode:           "chat",
				Answer:         "Hello! I'm here to help you with your questions.",
				Metadata:       Metadata{Usage: Usage{PromptTokens: 15, CompletionTokens: 25, TotalTokens: 40}},
				CreatedAt:      time.Now(),
				ConversationID: "conv_456",
			}
			json.NewEncoder(w).Encode(response)
		}))
		defer server.Close()

		// Update client to use mock server
		difyClient.baseURL = server.URL

		// Prepare test data
		ctx := context.Background()
		req := &ChatRequest{
			Query:        "Hello, world!",
			ResponseMode: "blocking",
			User:         "test_user",
		}

		// Call the method under test
		resp, err := difyClient.CreateChatMessage(ctx, req)

		// Assert results
		assert.NoError(t, err)
		assert.NotNil(t, resp)
		assert.Equal(t, "msg_456", resp.MessageID)
		assert.Equal(t, "Hello! I'm here to help you with your questions.", resp.Answer)
	})

	// Test CreateChatMessage with HTTP error
	t.Run("CreateChatMessageHTTPError", func(t *testing.T) {
		// Create mock HTTP server that returns error
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		}))
		defer server.Close()

		// Update client to use mock server
		difyClient.baseURL = server.URL

		// Prepare test data
		ctx := context.Background()
		req := &ChatRequest{
			Query:        "Hello, world!",
			ResponseMode: "blocking",
			User:         "test_user",
		}

		// Call the method under test
		resp, err := difyClient.CreateChatMessage(ctx, req)

		// Assert results
		assert.Error(t, err)
		assert.Nil(t, resp)
		assert.Equal(t, "failed to create chat message", err.Error())
	})

	// Test UploadFile success
	t.Run("UploadFileSuccess", func(t *testing.T) {
		// Create mock HTTP server
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Verify request
			assert.Equal(t, "POST", r.Method)
			assert.Equal(t, "/files/upload", r.URL.Path)
			assert.Equal(t, "Bearer test_api_key", r.Header.Get("Authorization"))
			assert.Contains(t, r.Header.Get("Content-Type"), "multipart/form-data")

			// Send mock response
			response := FileUploadResponse{
				ID:        "file_123",
				Name:      "test.txt",
				Size:      1024,
				MimeType:  "text/plain",
				CreatedAt: time.Now(),
				CreatedBy: "test_user",
			}
			json.NewEncoder(w).Encode(response)
		}))
		defer server.Close()

		// Update client to use mock server
		difyClient.baseURL = server.URL

		// Prepare test data
		ctx := context.Background()
		fileContent := "This is a test file content"
		req := &FileUploadRequest{
			File:     bytes.NewBufferString(fileContent),
			FileName: "test.txt",
			MimeType: "text/plain",
			User:     "test_user",
		}

		// Call the method under test
		resp, err := difyClient.UploadFile(ctx, req)

		// Assert results
		assert.NoError(t, err)
		assert.NotNil(t, resp)
		assert.Equal(t, "file_123", resp.ID)
		assert.Equal(t, "test.txt", resp.Name)
	})

	// Test UploadFile with HTTP error
	t.Run("UploadFileHTTPError", func(t *testing.T) {
		// Create mock HTTP server that returns error
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		}))
		defer server.Close()

		// Update client to use mock server
		difyClient.baseURL = server.URL

		// Prepare test data
		ctx := context.Background()
		fileContent := "This is a test file content"
		req := &FileUploadRequest{
			File:     bytes.NewBufferString(fileContent),
			FileName: "test.txt",
			MimeType: "text/plain",
			User:     "test_user",
		}

		// Call the method under test
		resp, err := difyClient.UploadFile(ctx, req)

		// Assert results
		assert.Error(t, err)
		assert.Nil(t, resp)
		assert.Equal(t, "failed to upload file", err.Error())
	})

	// Test NewDifyClient
	t.Run("NewDifyClient", func(t *testing.T) {
		client := NewDifyClient("http://test.com", "test_key")
		assert.NotNil(t, client)
		assert.Equal(t, "http://test.com", client.baseURL)
		assert.Equal(t, "test_key", client.apiKey)
		assert.NotNil(t, client.httpClient)
	})
}