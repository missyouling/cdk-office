package client

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"time"

	"cdk-office/pkg/logger"
)

// DifyClientInterface defines the interface for Dify API client
type DifyClientInterface interface {
	CreateCompletionMessage(ctx context.Context, req *CompletionRequest) (*CompletionResponse, error)
	CreateChatMessage(ctx context.Context, req *ChatRequest) (*ChatResponse, error)
	UploadFile(ctx context.Context, req *FileUploadRequest) (*FileUploadResponse, error)
}

// DifyClient implements the DifyClientInterface
type DifyClient struct {
	baseURL    string
	apiKey     string
	httpClient *http.Client
}

// NewDifyClient creates a new instance of DifyClient
func NewDifyClient(baseURL, apiKey string) *DifyClient {
	return &DifyClient{
		baseURL: baseURL,
		apiKey:  apiKey,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// CompletionRequest represents the request for completion API
type CompletionRequest struct {
	Query          string                 `json:"query"`
	Inputs         map[string]interface{} `json:"inputs,omitempty"`
	ResponseMode   string                 `json:"response_mode"` // streaming or blocking
	ConversationID string                 `json:"conversation_id,omitempty"`
	User           string                 `json:"user"`
}

// CompletionResponse represents the response for completion API
type CompletionResponse struct {
	MessageID      string      `json:"message_id"`
	Mode           string      `json:"mode"`
	Answer         string      `json:"answer"`
	Metadata       Metadata    `json:"metadata"`
	CreatedAt      time.Time   `json:"created_at"`
	ConversationID string      `json:"conversation_id"`
}

// ChatRequest represents the request for chat API
type ChatRequest struct {
	Query          string                 `json:"query"`
	Inputs         map[string]interface{} `json:"inputs,omitempty"`
	ResponseMode   string                 `json:"response_mode"` // streaming or blocking
	ConversationID string                 `json:"conversation_id,omitempty"`
	User           string                 `json:"user"`
}

// ChatResponse represents the response for chat API
type ChatResponse struct {
	MessageID      string      `json:"message_id"`
	Mode           string      `json:"mode"`
	Answer         string      `json:"answer"`
	Metadata       Metadata    `json:"metadata"`
	CreatedAt      time.Time   `json:"created_at"`
	ConversationID string      `json:"conversation_id"`
}

// FileUploadRequest represents the request for file upload API
type FileUploadRequest struct {
	File     io.Reader `json:"-"`
	FileName string    `json:"-"`
	MimeType string    `json:"-"`
	User     string    `json:"user"`
}

// FileUploadResponse represents the response for file upload API
type FileUploadResponse struct {
	ID        string    `json:"id"`
	Name      string    `json:"name"`
	Size      int64     `json:"size"`
	MimeType  string    `json:"mime_type"`
	CreatedAt time.Time `json:"created_at"`
	CreatedBy string    `json:"created_by"`
}

// Metadata represents metadata in responses
type Metadata struct {
	Usage Usage `json:"usage"`
}

// Usage represents usage information
type Usage struct {
	PromptTokens     int `json:"prompt_tokens"`
	CompletionTokens int `json:"completion_tokens"`
	TotalTokens      int `json:"total_tokens"`
}

// CreateCompletionMessage sends a completion message to Dify
func (c *DifyClient) CreateCompletionMessage(ctx context.Context, req *CompletionRequest) (*CompletionResponse, error) {
	url := fmt.Sprintf("%s/completion-messages", c.baseURL)
	
	// Convert request to JSON
	jsonData, err := json.Marshal(req)
	if err != nil {
		logger.Error("failed to marshal completion request", "error", err)
		return nil, errors.New("failed to create completion message")
	}

	// Create HTTP request
	httpReq, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		logger.Error("failed to create completion HTTP request", "error", err)
		return nil, errors.New("failed to create completion message")
	}

	// Set headers
	httpReq.Header.Set("Authorization", "Bearer "+c.apiKey)
	httpReq.Header.Set("Content-Type", "application/json")

	// Send request
	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		logger.Error("failed to send completion request", "error", err)
		return nil, errors.New("failed to create completion message")
	}
	defer resp.Body.Close()

	// Check status code
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		logger.Error("completion request failed", "status", resp.StatusCode, "body", string(body))
		return nil, errors.New("failed to create completion message")
	}

	// Parse response
	var completionResp CompletionResponse
	if err := json.NewDecoder(resp.Body).Decode(&completionResp); err != nil {
		logger.Error("failed to decode completion response", "error", err)
		return nil, errors.New("failed to create completion message")
	}

	return &completionResp, nil
}

// CreateChatMessage sends a chat message to Dify
func (c *DifyClient) CreateChatMessage(ctx context.Context, req *ChatRequest) (*ChatResponse, error) {
	url := fmt.Sprintf("%s/chat-messages", c.baseURL)
	
	// Convert request to JSON
	jsonData, err := json.Marshal(req)
	if err != nil {
		logger.Error("failed to marshal chat request", "error", err)
		return nil, errors.New("failed to create chat message")
	}

	// Create HTTP request
	httpReq, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		logger.Error("failed to create chat HTTP request", "error", err)
		return nil, errors.New("failed to create chat message")
	}

	// Set headers
	httpReq.Header.Set("Authorization", "Bearer "+c.apiKey)
	httpReq.Header.Set("Content-Type", "application/json")

	// Send request
	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		logger.Error("failed to send chat request", "error", err)
		return nil, errors.New("failed to create chat message")
	}
	defer resp.Body.Close()

	// Check status code
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		logger.Error("chat request failed", "status", resp.StatusCode, "body", string(body))
		return nil, errors.New("failed to create chat message")
	}

	// Parse response
	var chatResp ChatResponse
	if err := json.NewDecoder(resp.Body).Decode(&chatResp); err != nil {
		logger.Error("failed to decode chat response", "error", err)
		return nil, errors.New("failed to create chat message")
	}

	return &chatResp, nil
}

// UploadFile uploads a file to Dify
func (c *DifyClient) UploadFile(ctx context.Context, req *FileUploadRequest) (*FileUploadResponse, error) {
	url := fmt.Sprintf("%s/files/upload", c.baseURL)
	
	// Create a buffer to write our multipart form
	var body bytes.Buffer
	writer := multipart.NewWriter(&body)
	
	// Create the file field
	fileWriter, err := writer.CreateFormFile("file", req.FileName)
	if err != nil {
		logger.Error("failed to create form file", "error", err)
		return nil, errors.New("failed to upload file")
	}
	
	// Copy the file data to the file field
	_, err = io.Copy(fileWriter, req.File)
	if err != nil {
		logger.Error("failed to copy file data", "error", err)
		return nil, errors.New("failed to upload file")
	}
	
	// Add the user field
	err = writer.WriteField("user", req.User)
	if err != nil {
		logger.Error("failed to write user field", "error", err)
		return nil, errors.New("failed to upload file")
	}
	
	// Close the multipart writer
	err = writer.Close()
	if err != nil {
		logger.Error("failed to close multipart writer", "error", err)
		return nil, errors.New("failed to upload file")
	}
	
	// Create HTTP request
	httpReq, err := http.NewRequestWithContext(ctx, "POST", url, &body)
	if err != nil {
		logger.Error("failed to create file upload HTTP request", "error", err)
		return nil, errors.New("failed to upload file")
	}
	
	// Set headers
	httpReq.Header.Set("Authorization", "Bearer "+c.apiKey)
	httpReq.Header.Set("Content-Type", writer.FormDataContentType())
	
	// Send request
	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		logger.Error("failed to send file upload request", "error", err)
		return nil, errors.New("failed to upload file")
	}
	defer resp.Body.Close()
	
	// Check status code
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		logger.Error("file upload request failed", "status", resp.StatusCode, "body", string(body))
		return nil, errors.New("failed to upload file")
	}
	
	// Parse response
	var fileResp FileUploadResponse
	if err := json.NewDecoder(resp.Body).Decode(&fileResp); err != nil {
		logger.Error("failed to decode file upload response", "error", err)
		return nil, errors.New("failed to upload file")
	}
	
	return &fileResp, nil
}