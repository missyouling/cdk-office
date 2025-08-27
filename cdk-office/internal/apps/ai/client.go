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

package ai

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/linux-do/cdk-office/internal/models"
)

// AIClient AI客户端接口
type AIClient interface {
	Chat(ctx context.Context, request map[string]interface{}) (map[string]interface{}, error)
	Embedding(ctx context.Context, request map[string]interface{}) (map[string]interface{}, error)
	Translate(ctx context.Context, request map[string]interface{}) (map[string]interface{}, error)
}

// BaseAIClient 基础AI客户端
type BaseAIClient struct {
	config     *models.AIServiceConfig
	httpClient *http.Client
}

// NewBaseAIClient 创建基础AI客户端
func NewBaseAIClient(config *models.AIServiceConfig) *BaseAIClient {
	return &BaseAIClient{
		config: config,
		httpClient: &http.Client{
			Timeout: time.Duration(config.Timeout) * time.Second,
		},
	}
}

// makeRequest 发送HTTP请求
func (c *BaseAIClient) makeRequest(ctx context.Context, endpoint string, payload map[string]interface{}) (map[string]interface{}, error) {
	jsonData, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", endpoint, bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// 设置基础头部
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+c.config.APIKey)

	// 设置自定义头部
	if c.config.CustomHeaders != "" {
		var customHeaders map[string]string
		if err := json.Unmarshal([]byte(c.config.CustomHeaders), &customHeaders); err == nil {
			for key, value := range customHeaders {
				req.Header.Set(key, value)
			}
		}
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("API error: %d %s", resp.StatusCode, string(body))
	}

	var result map[string]interface{}
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	return result, nil
}

// OpenAIClient OpenAI客户端
type OpenAIClient struct {
	*BaseAIClient
}

// NewOpenAIClient 创建OpenAI客户端
func NewOpenAIClient(config *models.AIServiceConfig) *OpenAIClient {
	return &OpenAIClient{
		BaseAIClient: NewBaseAIClient(config),
	}
}

func (c *OpenAIClient) Chat(ctx context.Context, request map[string]interface{}) (map[string]interface{}, error) {
	endpoint := c.config.APIEndpoint
	if endpoint == "" {
		endpoint = "https://api.openai.com/v1/chat/completions"
	}

	// 设置默认模型
	if _, exists := request["model"]; !exists {
		request["model"] = "gpt-3.5-turbo"
	}

	return c.makeRequest(ctx, endpoint, request)
}

func (c *OpenAIClient) Embedding(ctx context.Context, request map[string]interface{}) (map[string]interface{}, error) {
	endpoint := "https://api.openai.com/v1/embeddings"

	// 设置默认模型
	if _, exists := request["model"]; !exists {
		request["model"] = "text-embedding-ada-002"
	}

	return c.makeRequest(ctx, endpoint, request)
}

func (c *OpenAIClient) Translate(ctx context.Context, request map[string]interface{}) (map[string]interface{}, error) {
	// OpenAI没有专门的翻译API，使用Chat API实现
	messages := []map[string]string{
		{
			"role": "system",
			"content": fmt.Sprintf("Translate the following text from %s to %s:",
				request["from"], request["to"]),
		},
		{
			"role":    "user",
			"content": request["text"].(string),
		},
	}

	chatRequest := map[string]interface{}{
		"model":    "gpt-3.5-turbo",
		"messages": messages,
	}

	return c.Chat(ctx, chatRequest)
}

// BaiduAIClient 百度AI客户端
type BaiduAIClient struct {
	*BaseAIClient
}

// NewBaiduAIClient 创建百度AI客户端
func NewBaiduAIClient(config *models.AIServiceConfig) *BaiduAIClient {
	return &BaiduAIClient{
		BaseAIClient: NewBaseAIClient(config),
	}
}

func (c *BaiduAIClient) Chat(ctx context.Context, request map[string]interface{}) (map[string]interface{}, error) {
	// 首先获取access_token
	accessToken, err := c.getAccessToken(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get access token: %w", err)
	}

	endpoint := fmt.Sprintf("%s?access_token=%s", c.config.APIEndpoint, accessToken)

	// 转换为百度API格式
	baiduRequest := c.convertToBaiduFormat(request)

	return c.makeRequest(ctx, endpoint, baiduRequest)
}

func (c *BaiduAIClient) Embedding(ctx context.Context, request map[string]interface{}) (map[string]interface{}, error) {
	accessToken, err := c.getAccessToken(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get access token: %w", err)
	}

	endpoint := fmt.Sprintf("https://aip.baidubce.com/rpc/2.0/ai_custom/v1/wenxinworkshop/embeddings/embedding-v1?access_token=%s", accessToken)

	return c.makeRequest(ctx, endpoint, request)
}

func (c *BaiduAIClient) Translate(ctx context.Context, request map[string]interface{}) (map[string]interface{}, error) {
	accessToken, err := c.getAccessToken(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get access token: %w", err)
	}

	endpoint := fmt.Sprintf("https://fanyi-api.baidu.com/api/trans/vip/translate?access_token=%s", accessToken)

	return c.makeRequest(ctx, endpoint, request)
}

func (c *BaiduAIClient) getAccessToken(ctx context.Context) (string, error) {
	tokenURL := "https://aip.baidubce.com/oauth/2.0/token"
	payload := map[string]interface{}{
		"grant_type":    "client_credentials",
		"client_id":     c.config.APIKey,
		"client_secret": c.config.SecretKey,
	}

	result, err := c.makeRequest(ctx, tokenURL, payload)
	if err != nil {
		return "", err
	}

	token, ok := result["access_token"].(string)
	if !ok {
		return "", fmt.Errorf("invalid access token response")
	}

	return token, nil
}

func (c *BaiduAIClient) convertToBaiduFormat(request map[string]interface{}) map[string]interface{} {
	// 将OpenAI格式转换为百度格式
	messages, ok := request["messages"].([]map[string]string)
	if !ok {
		return request
	}

	// 百度API需要不同的格式
	return map[string]interface{}{
		"messages": messages,
	}
}

// TencentAIClient 腾讯AI客户端
type TencentAIClient struct {
	*BaseAIClient
}

// NewTencentAIClient 创建腾讯AI客户端
func NewTencentAIClient(config *models.AIServiceConfig) *TencentAIClient {
	return &TencentAIClient{
		BaseAIClient: NewBaseAIClient(config),
	}
}

func (c *TencentAIClient) Chat(ctx context.Context, request map[string]interface{}) (map[string]interface{}, error) {
	endpoint := c.config.APIEndpoint
	if endpoint == "" {
		endpoint = "https://hunyuan.tencentcloudapi.com"
	}

	// 腾讯云API需要特殊的签名和头部
	c.setTencentHeaders(request)

	return c.makeRequest(ctx, endpoint, request)
}

func (c *TencentAIClient) Embedding(ctx context.Context, request map[string]interface{}) (map[string]interface{}, error) {
	// 腾讯云目前没有公开的embedding API
	return nil, fmt.Errorf("embedding not supported by Tencent AI")
}

func (c *TencentAIClient) Translate(ctx context.Context, request map[string]interface{}) (map[string]interface{}, error) {
	endpoint := "https://tmt.tencentcloudapi.com"
	c.setTencentHeaders(request)

	return c.makeRequest(ctx, endpoint, request)
}

func (c *TencentAIClient) setTencentHeaders(request map[string]interface{}) {
	// 这里应该实现腾讯云的签名算法
	// 简化处理，实际应该使用腾讯云SDK
}

// AliyunAIClient 阿里云AI客户端
type AliyunAIClient struct {
	*BaseAIClient
}

// NewAliyunAIClient 创建阿里云AI客户端
func NewAliyunAIClient(config *models.AIServiceConfig) *AliyunAIClient {
	return &AliyunAIClient{
		BaseAIClient: NewBaseAIClient(config),
	}
}

func (c *AliyunAIClient) Chat(ctx context.Context, request map[string]interface{}) (map[string]interface{}, error) {
	endpoint := c.config.APIEndpoint
	if endpoint == "" {
		endpoint = "https://dashscope.aliyuncs.com/api/v1/services/aigc/text-generation/generation"
	}

	// 设置阿里云API格式
	aliyunRequest := c.convertToAliyunFormat(request)

	return c.makeRequest(ctx, endpoint, aliyunRequest)
}

func (c *AliyunAIClient) Embedding(ctx context.Context, request map[string]interface{}) (map[string]interface{}, error) {
	endpoint := "https://dashscope.aliyuncs.com/api/v1/services/embeddings/text-embedding/text-embedding"

	return c.makeRequest(ctx, endpoint, request)
}

func (c *AliyunAIClient) Translate(ctx context.Context, request map[string]interface{}) (map[string]interface{}, error) {
	endpoint := "https://mt.cn-hangzhou.aliyuncs.com"

	return c.makeRequest(ctx, endpoint, request)
}

func (c *AliyunAIClient) convertToAliyunFormat(request map[string]interface{}) map[string]interface{} {
	// 将通用格式转换为阿里云格式
	return map[string]interface{}{
		"model": "qwen-turbo",
		"input": request,
	}
}

// GenericAIClient 通用AI客户端
type GenericAIClient struct {
	*BaseAIClient
}

// NewGenericAIClient 创建通用AI客户端
func NewGenericAIClient(config *models.AIServiceConfig) *GenericAIClient {
	return &GenericAIClient{
		BaseAIClient: NewBaseAIClient(config),
	}
}

func (c *GenericAIClient) Chat(ctx context.Context, request map[string]interface{}) (map[string]interface{}, error) {
	return c.makeRequest(ctx, c.config.APIEndpoint, request)
}

func (c *GenericAIClient) Embedding(ctx context.Context, request map[string]interface{}) (map[string]interface{}, error) {
	return c.makeRequest(ctx, c.config.APIEndpoint, request)
}

func (c *GenericAIClient) Translate(ctx context.Context, request map[string]interface{}) (map[string]interface{}, error) {
	return c.makeRequest(ctx, c.config.APIEndpoint, request)
}
