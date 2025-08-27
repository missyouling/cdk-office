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

package ocr

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"time"

	"github.com/linux-do/cdk-office/internal/models"
)

// OCRClient OCR客户端接口
type OCRClient interface {
	RecognizeText(ctx context.Context, imageData []byte, options OCROptions) (*OCRResult, error)
	RecognizeTable(ctx context.Context, imageData []byte) (*TableResult, error)
	RecognizeHandwriting(ctx context.Context, imageData []byte) (*OCRResult, error)
}

// OCROptions OCR识别选项
type OCROptions struct {
	Language   string `json:"language"`    // 识别语言
	OutputType string `json:"output_type"` // 输出类型: text, json
	Accuracy   string `json:"accuracy"`    // 准确度: fast, accurate
}

// OCRResult OCR识别结果
type OCRResult struct {
	Text       string                 `json:"text"`
	Confidence float64                `json:"confidence"`
	Words      []WordResult           `json:"words,omitempty"`
	Lines      []LineResult           `json:"lines,omitempty"`
	Details    map[string]interface{} `json:"details,omitempty"`
}

// WordResult 单词识别结果
type WordResult struct {
	Text        string      `json:"text"`
	Confidence  float64     `json:"confidence"`
	BoundingBox BoundingBox `json:"bounding_box"`
}

// LineResult 行识别结果
type LineResult struct {
	Text        string       `json:"text"`
	Confidence  float64      `json:"confidence"`
	Words       []WordResult `json:"words"`
	BoundingBox BoundingBox  `json:"bounding_box"`
}

// BoundingBox 边界框
type BoundingBox struct {
	Left   int `json:"left"`
	Top    int `json:"top"`
	Width  int `json:"width"`
	Height int `json:"height"`
}

// TableResult 表格识别结果
type TableResult struct {
	Tables []Table `json:"tables"`
}

// Table 表格
type Table struct {
	Rows []TableRow `json:"rows"`
}

// TableRow 表格行
type TableRow struct {
	Cells []TableCell `json:"cells"`
}

// TableCell 表格单元格
type TableCell struct {
	Text        string      `json:"text"`
	RowSpan     int         `json:"row_span"`
	ColSpan     int         `json:"col_span"`
	BoundingBox BoundingBox `json:"bounding_box"`
}

// BaseOCRClient 基础OCR客户端
type BaseOCRClient struct {
	config     *models.AIServiceConfig
	httpClient *http.Client
}

// NewBaseOCRClient 创建基础OCR客户端
func NewBaseOCRClient(config *models.AIServiceConfig) *BaseOCRClient {
	return &BaseOCRClient{
		config: config,
		httpClient: &http.Client{
			Timeout: time.Duration(config.Timeout) * time.Second,
		},
	}
}

// makeRequest 发送HTTP请求
func (c *BaseOCRClient) makeRequest(ctx context.Context, endpoint string, imageData []byte, extraParams map[string]string) (map[string]interface{}, error) {
	var buf bytes.Buffer
	writer := multipart.NewWriter(&buf)

	// 添加图片文件
	part, err := writer.CreateFormFile("image", "image.jpg")
	if err != nil {
		return nil, fmt.Errorf("failed to create form file: %w", err)
	}

	if _, err := part.Write(imageData); err != nil {
		return nil, fmt.Errorf("failed to write image data: %w", err)
	}

	// 添加额外参数
	for key, value := range extraParams {
		if err := writer.WriteField(key, value); err != nil {
			return nil, fmt.Errorf("failed to write field %s: %w", key, err)
		}
	}

	writer.Close()

	req, err := http.NewRequestWithContext(ctx, "POST", endpoint, &buf)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", writer.FormDataContentType())
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

// BaiduOCRClient 百度OCR客户端
type BaiduOCRClient struct {
	*BaseOCRClient
}

// NewBaiduOCRClient 创建百度OCR客户端
func NewBaiduOCRClient(config *models.AIServiceConfig) *BaiduOCRClient {
	return &BaiduOCRClient{
		BaseOCRClient: NewBaseOCRClient(config),
	}
}

func (c *BaiduOCRClient) RecognizeText(ctx context.Context, imageData []byte, options OCROptions) (*OCRResult, error) {
	// 获取access_token
	accessToken, err := c.getAccessToken(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get access token: %w", err)
	}

	endpoint := fmt.Sprintf("https://aip.baidubce.com/rest/2.0/ocr/v1/general_basic?access_token=%s", accessToken)

	// 百度OCR使用base64编码的图片
	imageBase64 := base64.StdEncoding.EncodeToString(imageData)

	params := map[string]string{
		"image": imageBase64,
	}

	if options.Language != "" {
		params["language_type"] = options.Language
	}

	response, err := c.makeRequestWithForm(ctx, endpoint, params)
	if err != nil {
		return nil, err
	}

	return c.parseBaiduResponse(response)
}

func (c *BaiduOCRClient) RecognizeTable(ctx context.Context, imageData []byte) (*TableResult, error) {
	accessToken, err := c.getAccessToken(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get access token: %w", err)
	}

	endpoint := fmt.Sprintf("https://aip.baidubce.com/rest/2.0/ocr/v1/table?access_token=%s", accessToken)

	imageBase64 := base64.StdEncoding.EncodeToString(imageData)
	params := map[string]string{
		"image": imageBase64,
	}

	response, err := c.makeRequestWithForm(ctx, endpoint, params)
	if err != nil {
		return nil, err
	}

	return c.parseBaiduTableResponse(response)
}

func (c *BaiduOCRClient) RecognizeHandwriting(ctx context.Context, imageData []byte) (*OCRResult, error) {
	accessToken, err := c.getAccessToken(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get access token: %w", err)
	}

	endpoint := fmt.Sprintf("https://aip.baidubce.com/rest/2.0/ocr/v1/handwriting?access_token=%s", accessToken)

	imageBase64 := base64.StdEncoding.EncodeToString(imageData)
	params := map[string]string{
		"image": imageBase64,
	}

	response, err := c.makeRequestWithForm(ctx, endpoint, params)
	if err != nil {
		return nil, err
	}

	return c.parseBaiduResponse(response)
}

func (c *BaiduOCRClient) getAccessToken(ctx context.Context) (string, error) {
	tokenURL := "https://aip.baidubce.com/oauth/2.0/token"
	params := map[string]string{
		"grant_type":    "client_credentials",
		"client_id":     c.config.APIKey,
		"client_secret": c.config.SecretKey,
	}

	result, err := c.makeRequestWithForm(ctx, tokenURL, params)
	if err != nil {
		return "", err
	}

	token, ok := result["access_token"].(string)
	if !ok {
		return "", fmt.Errorf("invalid access token response")
	}

	return token, nil
}

func (c *BaiduOCRClient) makeRequestWithForm(ctx context.Context, endpoint string, params map[string]string) (map[string]interface{}, error) {
	var buf bytes.Buffer
	writer := multipart.NewWriter(&buf)

	for key, value := range params {
		if err := writer.WriteField(key, value); err != nil {
			return nil, fmt.Errorf("failed to write field %s: %w", key, err)
		}
	}
	writer.Close()

	req, err := http.NewRequestWithContext(ctx, "POST", endpoint, &buf)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", writer.FormDataContentType())

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

func (c *BaiduOCRClient) parseBaiduResponse(response map[string]interface{}) (*OCRResult, error) {
	result := &OCRResult{
		Words: make([]WordResult, 0),
		Lines: make([]LineResult, 0),
	}

	wordsResult, ok := response["words_result"].([]interface{})
	if !ok {
		return nil, fmt.Errorf("invalid response format")
	}

	var fullText string
	for _, wordItem := range wordsResult {
		word, ok := wordItem.(map[string]interface{})
		if !ok {
			continue
		}

		text, _ := word["words"].(string)
		fullText += text + "\n"

		// 解析位置信息（如果有）
		if location, ok := word["location"].(map[string]interface{}); ok {
			bbox := BoundingBox{
				Left:   int(location["left"].(float64)),
				Top:    int(location["top"].(float64)),
				Width:  int(location["width"].(float64)),
				Height: int(location["height"].(float64)),
			}

			result.Words = append(result.Words, WordResult{
				Text:        text,
				Confidence:  1.0, // 百度OCR不返回置信度
				BoundingBox: bbox,
			})
		}
	}

	result.Text = fullText
	result.Confidence = 0.9 // 默认置信度

	return result, nil
}

func (c *BaiduOCRClient) parseBaiduTableResponse(response map[string]interface{}) (*TableResult, error) {
	result := &TableResult{
		Tables: make([]Table, 0),
	}

	// 百度表格OCR的响应格式解析
	// 这里需要根据实际API响应格式来解析

	return result, nil
}

// TencentOCRClient 腾讯OCR客户端
type TencentOCRClient struct {
	*BaseOCRClient
}

// NewTencentOCRClient 创建腾讯OCR客户端
func NewTencentOCRClient(config *models.AIServiceConfig) *TencentOCRClient {
	return &TencentOCRClient{
		BaseOCRClient: NewBaseOCRClient(config),
	}
}

func (c *TencentOCRClient) RecognizeText(ctx context.Context, imageData []byte, options OCROptions) (*OCRResult, error) {
	endpoint := "https://ocr.tencentcloudapi.com"

	params := map[string]string{
		"ImageBase64": base64.StdEncoding.EncodeToString(imageData),
	}

	response, err := c.makeRequest(ctx, endpoint, imageData, params)
	if err != nil {
		return nil, err
	}

	return c.parseTencentResponse(response)
}

func (c *TencentOCRClient) RecognizeTable(ctx context.Context, imageData []byte) (*TableResult, error) {
	// 腾讯云表格OCR实现
	return &TableResult{}, nil
}

func (c *TencentOCRClient) RecognizeHandwriting(ctx context.Context, imageData []byte) (*OCRResult, error) {
	// 腾讯云手写文字OCR实现
	return c.RecognizeText(ctx, imageData, OCROptions{})
}

func (c *TencentOCRClient) parseTencentResponse(response map[string]interface{}) (*OCRResult, error) {
	// 解析腾讯云OCR响应
	result := &OCRResult{
		Words: make([]WordResult, 0),
		Lines: make([]LineResult, 0),
	}

	// 根据腾讯云API文档解析响应

	return result, nil
}

// AliyunOCRClient 阿里云OCR客户端
type AliyunOCRClient struct {
	*BaseOCRClient
}

// NewAliyunOCRClient 创建阿里云OCR客户端
func NewAliyunOCRClient(config *models.AIServiceConfig) *AliyunOCRClient {
	return &AliyunOCRClient{
		BaseOCRClient: NewBaseOCRClient(config),
	}
}

func (c *AliyunOCRClient) RecognizeText(ctx context.Context, imageData []byte, options OCROptions) (*OCRResult, error) {
	endpoint := "https://ocr-api.cn-hangzhou.aliyuncs.com"

	params := map[string]string{
		"image": base64.StdEncoding.EncodeToString(imageData),
	}

	response, err := c.makeRequest(ctx, endpoint, imageData, params)
	if err != nil {
		return nil, err
	}

	return c.parseAliyunResponse(response)
}

func (c *AliyunOCRClient) RecognizeTable(ctx context.Context, imageData []byte) (*TableResult, error) {
	// 阿里云表格OCR实现
	return &TableResult{}, nil
}

func (c *AliyunOCRClient) RecognizeHandwriting(ctx context.Context, imageData []byte) (*OCRResult, error) {
	// 阿里云手写文字OCR实现
	return c.RecognizeText(ctx, imageData, OCROptions{})
}

func (c *AliyunOCRClient) parseAliyunResponse(response map[string]interface{}) (*OCRResult, error) {
	// 解析阿里云OCR响应
	result := &OCRResult{
		Words: make([]WordResult, 0),
		Lines: make([]LineResult, 0),
	}

	// 根据阿里云API文档解析响应

	return result, nil
}

// GenericOCRClient 通用OCR客户端
type GenericOCRClient struct {
	*BaseOCRClient
}

// NewGenericOCRClient 创建通用OCR客户端
func NewGenericOCRClient(config *models.AIServiceConfig) *GenericOCRClient {
	return &GenericOCRClient{
		BaseOCRClient: NewBaseOCRClient(config),
	}
}

func (c *GenericOCRClient) RecognizeText(ctx context.Context, imageData []byte, options OCROptions) (*OCRResult, error) {
	params := map[string]string{
		"language": options.Language,
		"accuracy": options.Accuracy,
	}

	response, err := c.makeRequest(ctx, c.config.APIEndpoint, imageData, params)
	if err != nil {
		return nil, err
	}

	return c.parseGenericResponse(response)
}

func (c *GenericOCRClient) RecognizeTable(ctx context.Context, imageData []byte) (*TableResult, error) {
	// 通用表格OCR实现
	return &TableResult{}, nil
}

func (c *GenericOCRClient) RecognizeHandwriting(ctx context.Context, imageData []byte) (*OCRResult, error) {
	return c.RecognizeText(ctx, imageData, OCROptions{})
}

func (c *GenericOCRClient) parseGenericResponse(response map[string]interface{}) (*OCRResult, error) {
	// 解析通用OCR响应
	result := &OCRResult{}

	if text, ok := response["text"].(string); ok {
		result.Text = text
	}

	if confidence, ok := response["confidence"].(float64); ok {
		result.Confidence = confidence
	}

	return result, nil
}

// CreateOCRClient 根据配置创建OCR客户端
func CreateOCRClient(config *models.AIServiceConfig) OCRClient {
	switch config.Provider {
	case "baidu":
		return NewBaiduOCRClient(config)
	case "tencent":
		return NewTencentOCRClient(config)
	case "aliyun":
		return NewAliyunOCRClient(config)
	default:
		return NewGenericOCRClient(config)
	}
}
