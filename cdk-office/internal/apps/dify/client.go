/*
 * MIT License
 *
 * Copyright (c) 2025 CDK-Office
 *
 * Permission is hereby granted, free of charge, to any person obtaining a copy
 * of this software and associated documentation files (the "Software"), to deal
 * in the Software without restriction, including without limitation the rights
 * to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
 * copies of the Software, and to permit persons to whom the Software is
 * furnished to do so, subject to the following conditions:
 *
 * The above copyright notice and this permission notice shall be included in all
 * copies or substantial portions of the Software.
 *
 * THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
 * IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
 * FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
 * AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
 * LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
 * OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
 * SOFTWARE.
 */

package dify

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"strings"
	"time"
)

// Client Dify客户端
type Client struct {
	baseURL    string
	apiKey     string
	httpClient *http.Client
}

// NewClient 创建Dify客户端
func NewClient(baseURL, apiKey string) *Client {
	return &Client{
		baseURL:    strings.TrimSuffix(baseURL, "/"),
		apiKey:     apiKey,
		httpClient: &http.Client{Timeout: 30 * time.Second},
	}
}

// ChatRequest 聊天请求
type ChatRequest struct {
	Inputs       map[string]interface{} `json:"inputs"`
	Query        string                 `json:"query"`
	ResponseMode string                 `json:"response_mode"` // streaming, blocking
	User         string                 `json:"user"`
	Variables    map[string]interface{} `json:"variables,omitempty"`
}

// ChatResponse 聊天响应
type ChatResponse struct {
	Event          string                 `json:"event"`
	ConversationID string                 `json:"conversation_id"`
	MessageID      string                 `json:"message_id"`
	Answer         string                 `json:"answer"`
	Metadata       map[string]interface{} `json:"metadata"`
	CreatedAt      int64                  `json:"created_at"`
}

// CompletionRequest 文本补全请求
type CompletionRequest struct {
	Inputs       map[string]interface{} `json:"inputs"`
	ResponseMode string                 `json:"response_mode"`
	User         string                 `json:"user"`
}

// CompletionResponse 文本补全响应
type CompletionResponse struct {
	Answer    string                 `json:"answer"`
	MessageID string                 `json:"message_id"`
	Metadata  map[string]interface{} `json:"metadata"`
	Usage     UsageInfo              `json:"usage"`
}

// WorkflowRunRequest 工作流运行请求
type WorkflowRunRequest struct {
	Inputs       map[string]interface{} `json:"inputs"`
	ResponseMode string                 `json:"response_mode"`
	User         string                 `json:"user"`
}

// WorkflowRunResponse 工作流运行响应
type WorkflowRunResponse struct {
	WorkflowRunID string                 `json:"workflow_run_id"`
	TaskID        string                 `json:"task_id"`
	Data          map[string]interface{} `json:"data"`
	Error         string                 `json:"error"`
	Status        string                 `json:"status"`
}

// DocumentUploadRequest 文档上传请求
type DocumentUploadRequest struct {
	Data        []byte                 `json:"data"`
	Name        string                 `json:"name"`
	OriginalURL string                 `json:"original_url,omitempty"`
	ProcessRule map[string]interface{} `json:"process_rule,omitempty"`
}

// DocumentUploadResponse 文档上传响应
type DocumentUploadResponse struct {
	DocumentID string `json:"document_id"`
	BatchID    string `json:"batch_id"`
	Name       string `json:"name"`
	CreatedAt  int64  `json:"created_at"`
}

// DocumentListResponse 文档列表响应
type DocumentListResponse struct {
	Data    []DocumentInfo `json:"data"`
	HasMore bool           `json:"has_more"`
	Limit   int            `json:"limit"`
	Total   int            `json:"total"`
	Page    int            `json:"page"`
}

// DocumentInfo 文档信息
type DocumentInfo struct {
	ID                   string                 `json:"id"`
	Position             int                    `json:"position"`
	DataSourceType       string                 `json:"data_source_type"`
	DataSourceInfo       map[string]interface{} `json:"data_source_info"`
	DatasetProcessRuleID string                 `json:"dataset_process_rule_id"`
	Name                 string                 `json:"name"`
	CreatedFrom          string                 `json:"created_from"`
	CreatedBy            string                 `json:"created_by"`
	CreatedAt            int64                  `json:"created_at"`
	Tokens               int                    `json:"tokens"`
	IndexingStatus       string                 `json:"indexing_status"`
	Error                string                 `json:"error"`
	Enabled              bool                   `json:"enabled"`
	DisabledAt           int64                  `json:"disabled_at"`
	DisabledBy           string                 `json:"disabled_by"`
	Archived             bool                   `json:"archived"`
	DisplayStatus        string                 `json:"display_status"`
	WordCount            int                    `json:"word_count"`
	HitCount             int                    `json:"hit_count"`
	DocForm              string                 `json:"doc_form"`
}

// UsageInfo 使用信息
type UsageInfo struct {
	PromptTokens        int     `json:"prompt_tokens"`
	PromptUnitPrice     string  `json:"prompt_unit_price"`
	PromptPrice         string  `json:"prompt_price"`
	CompletionTokens    int     `json:"completion_tokens"`
	CompletionUnitPrice string  `json:"completion_unit_price"`
	CompletionPrice     string  `json:"completion_price"`
	TotalTokens         int     `json:"total_tokens"`
	TotalPrice          string  `json:"total_price"`
	Currency            string  `json:"currency"`
	Latency             float64 `json:"latency"`
}

// Chat 发送聊天消息
func (c *Client) Chat(ctx context.Context, req *ChatRequest) (*ChatResponse, error) {
	url := fmt.Sprintf("%s/v1/chat-messages", c.baseURL)
	return c.doRequest(ctx, "POST", url, req, &ChatResponse{})
}

// Completion 文本补全
func (c *Client) Completion(ctx context.Context, req *CompletionRequest) (*CompletionResponse, error) {
	url := fmt.Sprintf("%s/v1/completion-messages", c.baseURL)
	return c.doRequest(ctx, "POST", url, req, &CompletionResponse{})
}

// RunWorkflow 运行工作流
func (c *Client) RunWorkflow(ctx context.Context, req *WorkflowRunRequest) (*WorkflowRunResponse, error) {
	url := fmt.Sprintf("%s/v1/workflows/run", c.baseURL)
	return c.doRequest(ctx, "POST", url, req, &WorkflowRunResponse{})
}

// UploadDocument 上传文档到知识库
func (c *Client) UploadDocument(ctx context.Context, datasetID string, file io.Reader, filename string, originalURL string) (*DocumentUploadResponse, error) {
	url := fmt.Sprintf("%s/v1/datasets/%s/document/create_by_file", c.baseURL, datasetID)

	var buf bytes.Buffer
	writer := multipart.NewWriter(&buf)

	// 添加文件
	fileWriter, err := writer.CreateFormFile("file", filename)
	if err != nil {
		return nil, fmt.Errorf("failed to create form file: %w", err)
	}

	if _, err := io.Copy(fileWriter, file); err != nil {
		return nil, fmt.Errorf("failed to copy file data: %w", err)
	}

	// 添加其他字段
	if originalURL != "" {
		writer.WriteField("original_url", originalURL)
	}

	writer.Close()

	req, err := http.NewRequestWithContext(ctx, "POST", url, &buf)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+c.apiKey)
	req.Header.Set("Content-Type", writer.FormDataContentType())

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("API request failed with status %d: %s", resp.StatusCode, string(body))
	}

	var result DocumentUploadResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &result, nil
}

// UploadDocumentByText 通过文本上传文档到知识库
func (c *Client) UploadDocumentByText(ctx context.Context, datasetID string, name, text string, processRule map[string]interface{}) (*DocumentUploadResponse, error) {
	url := fmt.Sprintf("%s/v1/datasets/%s/document/create_by_text", c.baseURL, datasetID)

	req := map[string]interface{}{
		"name": name,
		"text": text,
	}

	if processRule != nil {
		req["process_rule"] = processRule
	}

	return c.doRequest(ctx, "POST", url, req, &DocumentUploadResponse{})
}

// ListDocuments 获取知识库文档列表
func (c *Client) ListDocuments(ctx context.Context, datasetID string, keyword string, page, limit int) (*DocumentListResponse, error) {
	url := fmt.Sprintf("%s/v1/datasets/%s/documents?page=%d&limit=%d", c.baseURL, datasetID, page, limit)

	if keyword != "" {
		url += "&keyword=" + keyword
	}

	return c.doRequest(ctx, "GET", url, nil, &DocumentListResponse{})
}

// UpdateDocument 更新文档
func (c *Client) UpdateDocument(ctx context.Context, datasetID, documentID string, name string, text string, processRule map[string]interface{}) error {
	url := fmt.Sprintf("%s/v1/datasets/%s/documents/%s", c.baseURL, datasetID, documentID)

	req := map[string]interface{}{
		"name": name,
		"text": text,
	}

	if processRule != nil {
		req["process_rule"] = processRule
	}

	_, err := c.doRequest(ctx, "POST", url, req, &map[string]interface{}{})
	return err
}

// DeleteDocument 删除文档
func (c *Client) DeleteDocument(ctx context.Context, datasetID, documentID string) error {
	url := fmt.Sprintf("%s/v1/datasets/%s/documents/%s", c.baseURL, datasetID, documentID)
	_, err := c.doRequest(ctx, "DELETE", url, nil, &map[string]interface{}{})
	return err
}

// GetDocumentIndexingStatus 获取文档索引状态
func (c *Client) GetDocumentIndexingStatus(ctx context.Context, datasetID, batchID string) (*DocumentInfo, error) {
	url := fmt.Sprintf("%s/v1/datasets/%s/documents/%s/indexing-status", c.baseURL, datasetID, batchID)
	return c.doRequest(ctx, "GET", url, nil, &DocumentInfo{})
}

// doRequest 执行HTTP请求
func (c *Client) doRequest(ctx context.Context, method, url string, reqBody interface{}, respBody interface{}) (interface{}, error) {
	var body io.Reader

	if reqBody != nil {
		jsonData, err := json.Marshal(reqBody)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal request body: %w", err)
		}
		body = bytes.NewReader(jsonData)
	}

	req, err := http.NewRequestWithContext(ctx, method, url, body)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+c.apiKey)
	if reqBody != nil {
		req.Header.Set("Content-Type", "application/json")
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("API request failed with status %d: %s", resp.StatusCode, string(bodyBytes))
	}

	if respBody != nil {
		if err := json.NewDecoder(resp.Body).Decode(respBody); err != nil {
			return nil, fmt.Errorf("failed to decode response: %w", err)
		}
		return respBody, nil
	}

	return nil, nil
}

// StreamingChatResponse 流式聊天响应
type StreamingChatResponse struct {
	Event          string                 `json:"event"`
	ConversationID string                 `json:"conversation_id"`
	MessageID      string                 `json:"message_id"`
	Answer         string                 `json:"answer"`
	Metadata       map[string]interface{} `json:"metadata"`
	CreatedAt      int64                  `json:"created_at"`
}

// StreamingChat 流式聊天
func (c *Client) StreamingChat(ctx context.Context, req *ChatRequest, callback func(*StreamingChatResponse) error) error {
	req.ResponseMode = "streaming"

	url := fmt.Sprintf("%s/v1/chat-messages", c.baseURL)

	jsonData, err := json.Marshal(req)
	if err != nil {
		return fmt.Errorf("failed to marshal request: %w", err)
	}

	httpReq, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewReader(jsonData))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	httpReq.Header.Set("Authorization", "Bearer "+c.apiKey)
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Accept", "text/event-stream")

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("API request failed with status %d: %s", resp.StatusCode, string(bodyBytes))
	}

	// 处理Server-Sent Events
	return c.processSSEStream(resp.Body, callback)
}

// processSSEStream 处理SSE流
func (c *Client) processSSEStream(reader io.Reader, callback func(*StreamingChatResponse) error) error {
	buf := make([]byte, 4096)
	var buffer strings.Builder

	for {
		n, err := reader.Read(buf)
		if err != nil {
			if err == io.EOF {
				break
			}
			return fmt.Errorf("failed to read stream: %w", err)
		}

		buffer.Write(buf[:n])

		// 处理完整的事件
		for {
			data := buffer.String()
			if !strings.Contains(data, "\n\n") {
				break
			}

			eventEnd := strings.Index(data, "\n\n")
			event := data[:eventEnd]
			buffer.Reset()
			buffer.WriteString(data[eventEnd+2:])

			if err := c.processSSEEvent(event, callback); err != nil {
				return err
			}
		}
	}

	return nil
}

// processSSEEvent 处理单个SSE事件
func (c *Client) processSSEEvent(event string, callback func(*StreamingChatResponse) error) error {
	lines := strings.Split(event, "\n")
	var data string

	for _, line := range lines {
		if strings.HasPrefix(line, "data: ") {
			data = line[6:] // 移除"data: "前缀
			break
		}
	}

	if data == "" || data == "[DONE]" {
		return nil
	}

	var response StreamingChatResponse
	if err := json.Unmarshal([]byte(data), &response); err != nil {
		return fmt.Errorf("failed to unmarshal SSE data: %w", err)
	}

	return callback(&response)
}
