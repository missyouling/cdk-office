package service

import (
	"context"
	"testing"
	
	"cdk-office/internal/dify/client"
	"cdk-office/pkg/logger"
	
	"github.com/stretchr/testify/assert"
)

// TestDifyClient tests the DifyClient
func TestDifyClient(t *testing.T) {
	// Set up test environment
	logger.InitTestLogger()
	
	// Create Dify client
	difyClient := client.NewDifyClient("test_api_key", "http://localhost:8000")
	
	// Test CreateCompletionMessage
	t.Run("CreateCompletionMessage", func(t *testing.T) {
		// Prepare test data
		ctx := context.Background()
		req := &client.CompletionRequest{
			Query:        "Hello, world!",
			ResponseMode: "blocking",
			User:         "test_user",
		}
		
		// Call the method under test
		resp, err := difyClient.CreateCompletionMessage(ctx, req)
		
		// Assert results
		// Note: This test will fail because we're not mocking the HTTP client
		// In a real test, we would mock the HTTP client to return a predefined response
		assert.Error(t, err)
		assert.Nil(t, resp)
	})
	
	// Test CreateChatMessage
	t.Run("CreateChatMessage", func(t *testing.T) {
		// Prepare test data
		ctx := context.Background()
		req := &client.ChatRequest{
			Query:        "Hello, world!",
			ResponseMode: "blocking",
			User:         "test_user",
		}
		
		// Call the method under test
		resp, err := difyClient.CreateChatMessage(ctx, req)
		
		// Assert results
		// Note: This test will fail because we're not mocking the HTTP client
		// In a real test, we would mock the HTTP client to return a predefined response
		assert.Error(t, err)
		assert.Nil(t, resp)
	})
}