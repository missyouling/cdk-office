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

package storage

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"strconv"
	"strings"
	"time"
)

// CloudDBProvider 云数据库存储提供商
type CloudDBProvider struct {
	config      CloudDBConfig
	client      *http.Client
	maxFileSize int64
	maxStorage  int64
}

// NewCloudDBProvider 创建云数据库存储提供商
func NewCloudDBProvider(config CloudDBConfig) (*CloudDBProvider, error) {
	if config.URL == "" || config.AnonKey == "" {
		return nil, fmt.Errorf("cloud database URL and anon key are required")
	}
	
	provider := &CloudDBProvider{
		config:      config,
		client:      &http.Client{Timeout: 30 * time.Second},
		maxFileSize: parseSize(config.MaxFileSize),
		maxStorage:  parseSize(config.MaxStorage),
	}
	
	return provider, nil
}

// Upload 上传文件到云数据库存储
func (p *CloudDBProvider) Upload(file *multipart.FileHeader, path string) (*FileInfo, error) {
	// 检查文件大小限制
	if p.maxFileSize > 0 && file.Size > p.maxFileSize {
		return nil, fmt.Errorf("file size exceeds limit")
	}
	
	// 检查存储空间
	quota, err := p.GetQuota()
	if err == nil && !quota.Unlimited && quota.Available < file.Size {
		return nil, fmt.Errorf("insufficient storage space")
	}
	
	// 打开文件
	src, err := file.Open()
	if err != nil {
		return nil, fmt.Errorf("failed to open file: %v", err)
	}
	defer src.Close()
	
	// 读取文件内容
	content, err := io.ReadAll(src)
	if err != nil {
		return nil, fmt.Errorf("failed to read file content: %v", err)
	}
	
	// 构建请求URL
	bucket := p.config.Bucket
	if bucket == "" {
		bucket = "files"
	}
	
	url := fmt.Sprintf("%s/storage/v1/object/%s/%s", 
		strings.TrimSuffix(p.config.URL, "/"), bucket, path)
	
	// 创建HTTP请求
	req, err := http.NewRequest("POST", url, bytes.NewReader(content))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %v", err)
	}
	
	// 设置请求头
	req.Header.Set("Authorization", "Bearer "+p.config.AnonKey)
	req.Header.Set("Content-Type", file.Header.Get("Content-Type"))
	req.Header.Set("Content-Length", strconv.FormatInt(file.Size, 10))
	
	// 发送请求
	resp, err := p.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to upload file: %v", err)
	}
	defer resp.Body.Close()
	
	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("upload failed with status %d: %s", resp.StatusCode, string(body))
	}
	
	// 生成访问URL
	fileURL := fmt.Sprintf("%s/storage/v1/object/public/%s/%s", 
		strings.TrimSuffix(p.config.URL, "/"), bucket, path)
	
	return &FileInfo{
		Path:         path,
		Name:         file.Filename,
		Size:         file.Size,
		MimeType:     file.Header.Get("Content-Type"),
		URL:          fileURL,
		Provider:     p.GetProviderName(),
		LastModified: time.Now(),
	}, nil
}

// Download 从云数据库存储下载文件
func (p *CloudDBProvider) Download(path string) (io.ReadCloser, error) {
	bucket := p.config.Bucket
	if bucket == "" {
		bucket = "files"
	}
	
	url := fmt.Sprintf("%s/storage/v1/object/%s/%s", 
		strings.TrimSuffix(p.config.URL, "/"), bucket, path)
	
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %v", err)
	}
	
	req.Header.Set("Authorization", "Bearer "+p.config.AnonKey)
	
	resp, err := p.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to download file: %v", err)
	}
	
	if resp.StatusCode != http.StatusOK {
		resp.Body.Close()
		return nil, fmt.Errorf("download failed with status %d", resp.StatusCode)
	}
	
	return resp.Body, nil
}

// Delete 删除云数据库存储的文件
func (p *CloudDBProvider) Delete(path string) error {
	bucket := p.config.Bucket
	if bucket == "" {
		bucket = "files"
	}
	
	url := fmt.Sprintf("%s/storage/v1/object/%s/%s", 
		strings.TrimSuffix(p.config.URL, "/"), bucket, path)
	
	req, err := http.NewRequest("DELETE", url, nil)
	if err != nil {
		return fmt.Errorf("failed to create request: %v", err)
	}
	
	req.Header.Set("Authorization", "Bearer "+p.config.AnonKey)
	
	resp, err := p.client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to delete file: %v", err)
	}
	defer resp.Body.Close()
	
	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusNoContent {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("delete failed with status %d: %s", resp.StatusCode, string(body))
	}
	
	return nil
}

// GetURL 获取文件访问URL
func (p *CloudDBProvider) GetURL(path string) (string, error) {
	bucket := p.config.Bucket
	if bucket == "" {
		bucket = "files"
	}
	
	// 对于云数据库存储，通常返回公共访问URL
	url := fmt.Sprintf("%s/storage/v1/object/public/%s/%s", 
		strings.TrimSuffix(p.config.URL, "/"), bucket, path)
	
	return url, nil
}

// GetQuota 获取云数据库存储空间信息
func (p *CloudDBProvider) GetQuota() (*QuotaInfo, error) {
	// 获取存储桶使用情况
	used, err := p.getBucketUsage()
	if err != nil {
		// 如果无法获取准确的使用情况，使用估算值
		used = 0
	}
	
	// 如果配置了最大存储空间
	if p.maxStorage > 0 {
		available := p.maxStorage - used
		if available < 0 {
			available = 0
		}
		
		return &QuotaInfo{
			Total:     p.maxStorage,
			Used:      used,
			Available: available,
			Unlimited: false,
		}, nil
	}
	
	// 默认假设有足够空间（免费层通常有限制但难以准确获取）
	defaultLimit := int64(500 * 1024 * 1024) // 500MB 默认限制
	available := defaultLimit - used
	if available < 0 {
		available = 0
	}
	
	return &QuotaInfo{
		Total:     defaultLimit,
		Used:      used,
		Available: available,
		Unlimited: false,
	}, nil
}

// ListFiles 列出文件
func (p *CloudDBProvider) ListFiles(prefix string) ([]*FileInfo, error) {
	bucket := p.config.Bucket
	if bucket == "" {
		bucket = "files"
	}
	
	url := fmt.Sprintf("%s/storage/v1/object/list/%s", 
		strings.TrimSuffix(p.config.URL, "/"), bucket)
	
	// 构建请求体
	requestBody := map[string]interface{}{
		"limit":  1000,
		"offset": 0,
	}
	
	if prefix != "" {
		requestBody["prefix"] = prefix
	}
	
	jsonData, err := json.Marshal(requestBody)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %v", err)
	}
	
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %v", err)
	}
	
	req.Header.Set("Authorization", "Bearer "+p.config.AnonKey)
	req.Header.Set("Content-Type", "application/json")
	
	resp, err := p.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to list files: %v", err)
	}
	defer resp.Body.Close()
	
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("list files failed with status %d", resp.StatusCode)
	}
	
	var objects []map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&objects); err != nil {
		return nil, fmt.Errorf("failed to decode response: %v", err)
	}
	
	var files []*FileInfo
	for _, obj := range objects {
		name, _ := obj["name"].(string)
		if name == "" {
			continue
		}
		
		size, _ := obj["metadata"].(map[string]interface{})
		fileSize := int64(0)
		if size != nil {
			if s, ok := size["size"].(float64); ok {
				fileSize = int64(s)
			}
		}
		
		files = append(files, &FileInfo{
			Path:         name,
			Name:         name,
			Size:         fileSize,
			URL:          fmt.Sprintf("%s/storage/v1/object/public/%s/%s", p.config.URL, bucket, name),
			Provider:     p.GetProviderName(),
			LastModified: time.Now(), // 云存储通常会提供具体时间
		})
	}
	
	return files, nil
}

// GetProviderName 获取提供商名称
func (p *CloudDBProvider) GetProviderName() string {
	return "cloud_db_" + p.config.Provider
}

// IsAvailable 检查提供商是否可用
func (p *CloudDBProvider) IsAvailable() bool {
	// 尝试获取存储桶信息来检查连接
	bucket := p.config.Bucket
	if bucket == "" {
		bucket = "files"
	}
	
	url := fmt.Sprintf("%s/storage/v1/bucket/%s", 
		strings.TrimSuffix(p.config.URL, "/"), bucket)
	
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return false
	}
	
	req.Header.Set("Authorization", "Bearer "+p.config.AnonKey)
	
	resp, err := p.client.Do(req)
	if err != nil {
		return false
	}
	defer resp.Body.Close()
	
	return resp.StatusCode == http.StatusOK
}

// getBucketUsage 获取存储桶使用情况
func (p *CloudDBProvider) getBucketUsage() (int64, error) {
	// 这里需要根据具体的云数据库提供商API来实现
	// Supabase和MemFire可能有不同的API端点
	
	// 由于API限制，这里返回0，实际使用中需要根据具体提供商实现
	return 0, nil
}

// 实现缺失的提供商构造函数（占位符实现）
func NewS3Provider(config S3Provider) (StorageProvider, error) {
	// TODO: 实现S3存储提供商
	return nil, fmt.Errorf("S3 provider not implemented yet")
}

func NewWebDAVProvider(config WebDAVConfig) (StorageProvider, error) {
	// TODO: 实现WebDAV存储提供商
	return nil, fmt.Errorf("WebDAV provider not implemented yet")
}

func NewCloudDriveProvider(config CloudDriveProvider) (StorageProvider, error) {
	// TODO: 实现云盘存储提供商
	return nil, fmt.Errorf("Cloud drive provider not implemented yet")
}