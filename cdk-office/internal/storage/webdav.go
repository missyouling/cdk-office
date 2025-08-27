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
	"fmt"
	"io"
	"mime/multipart"
	"path"
	"strings"
	"time"

	"github.com/studio-b12/gowebdav"
)

// WebDAVProvider WebDAV存储提供商
type WebDAVProvider struct {
	config WebDAVConfig
	client *gowebdav.Client
}

// NewWebDAVProvider 创建WebDAV存储提供商
func NewWebDAVProvider(config WebDAVConfig) (*WebDAVProvider, error) {
	if config.URL == "" {
		return nil, fmt.Errorf("WebDAV URL is required")
	}

	// 创建WebDAV客户端
	client := gowebdav.NewClient(config.URL, config.Username, config.Password)

	// 设置超时
	client.SetTimeout(30 * time.Second)

	provider := &WebDAVProvider{
		config: config,
		client: client,
	}

	// 测试连接
	if err := provider.testConnection(); err != nil {
		return nil, fmt.Errorf("WebDAV connection test failed: %v", err)
	}

	return provider, nil
}

// testConnection 测试WebDAV连接
func (p *WebDAVProvider) testConnection() error {
	// 尝试读取根目录来测试连接
	_, err := p.client.ReadDir("/")
	if err != nil {
		return fmt.Errorf("failed to connect to WebDAV server: %v", err)
	}

	// 确保基础路径存在
	if p.config.BasePath != "" && p.config.BasePath != "/" {
		err = p.client.MkdirAll(p.config.BasePath, 0755)
		if err != nil {
			return fmt.Errorf("failed to create base path: %v", err)
		}
	}

	return nil
}

// Upload 上传文件到WebDAV存储
func (p *WebDAVProvider) Upload(file *multipart.FileHeader, path string) (*FileInfo, error) {
	// 构建完整路径
	fullPath := p.buildFullPath(path)

	// 确保目录存在
	dir := path
	if idx := strings.LastIndex(dir, "/"); idx != -1 {
		dir = dir[:idx]
	}
	if dir != "" {
		dirPath := p.buildFullPath(dir)
		if err := p.client.MkdirAll(dirPath, 0755); err != nil {
			return nil, fmt.Errorf("failed to create directory: %v", err)
		}
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

	// 上传文件
	err = p.client.Write(fullPath, content, 0644)
	if err != nil {
		return nil, fmt.Errorf("failed to upload file: %v", err)
	}

	// 生成访问URL
	url, err := p.GetURL(path)
	if err != nil {
		return nil, fmt.Errorf("failed to generate URL: %v", err)
	}

	return &FileInfo{
		Path:         path,
		Name:         file.Filename,
		Size:         file.Size,
		MimeType:     file.Header.Get("Content-Type"),
		URL:          url,
		Provider:     p.GetProviderName(),
		LastModified: time.Now(),
	}, nil
}

// Download 从WebDAV存储下载文件
func (p *WebDAVProvider) Download(path string) (io.ReadCloser, error) {
	fullPath := p.buildFullPath(path)

	// 检查文件是否存在
	info, err := p.client.Stat(fullPath)
	if err != nil {
		return nil, fmt.Errorf("file not found: %v", err)
	}

	if info.IsDir() {
		return nil, fmt.Errorf("path is a directory, not a file")
	}

	// 读取文件内容
	content, err := p.client.Read(fullPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read file: %v", err)
	}

	return io.NopCloser(strings.NewReader(string(content))), nil
}

// Delete 删除WebDAV存储的文件
func (p *WebDAVProvider) Delete(path string) error {
	fullPath := p.buildFullPath(path)

	// 检查文件是否存在
	_, err := p.client.Stat(fullPath)
	if err != nil {
		return nil // 文件不存在，视为删除成功
	}

	// 删除文件
	err = p.client.Remove(fullPath)
	if err != nil {
		return fmt.Errorf("failed to delete file: %v", err)
	}

	return nil
}

// GetURL 获取文件访问URL
func (p *WebDAVProvider) GetURL(path string) (string, error) {
	fullPath := p.buildFullPath(path)

	// 检查文件是否存在
	_, err := p.client.Stat(fullPath)
	if err != nil {
		return "", fmt.Errorf("file not found: %v", err)
	}

	// 构建访问URL
	baseURL := strings.TrimSuffix(p.config.URL, "/")
	return fmt.Sprintf("%s%s", baseURL, fullPath), nil
}

// GetQuota 获取WebDAV存储空间信息
func (p *WebDAVProvider) GetQuota() (*QuotaInfo, error) {
	// WebDAV协议本身不提供标准的配额查询方法
	// 这里返回无限制，实际限制由WebDAV服务器控制
	return &QuotaInfo{
		Total:     0,
		Used:      0,
		Available: 0,
		Unlimited: true,
	}, nil
}

// ListFiles 列出WebDAV存储中的文件
func (p *WebDAVProvider) ListFiles(prefix string) ([]*FileInfo, error) {
	fullPrefix := p.buildFullPath(prefix)

	// 列出目录内容
	infos, err := p.client.ReadDir(fullPrefix)
	if err != nil {
		return nil, fmt.Errorf("failed to list files: %v", err)
	}

	var files []*FileInfo
	for _, info := range infos {
		if !info.IsDir() {
			relativePath := strings.TrimPrefix(path.Join(fullPrefix, info.Name()), p.config.BasePath)
			relativePath = strings.TrimPrefix(relativePath, "/")

			// 生成访问URL
			url, err := p.GetURL(relativePath)
			if err != nil {
				continue // 跳过无法生成URL的文件
			}

			files = append(files, &FileInfo{
				Path:         relativePath,
				Name:         info.Name(),
				Size:         info.Size(),
				MimeType:     "", // WebDAV可能不提供MIME类型
				URL:          url,
				Provider:     p.GetProviderName(),
				LastModified: info.ModTime(),
			})
		}
	}

	return files, nil
}

// GetProviderName 获取提供商名称
func (p *WebDAVProvider) GetProviderName() string {
	return "webdav"
}

// IsAvailable 检查WebDAV提供商是否可用
func (p *WebDAVProvider) IsAvailable() bool {
	// 尝试读取根目录来测试连接
	_, err := p.client.ReadDir("/")
	return err == nil
}

// buildFullPath 构建完整路径
func (p *WebDAVProvider) buildFullPath(path string) string {
	basePath := strings.TrimSuffix(p.config.BasePath, "/")
	if basePath == "" {
		basePath = "/"
	}

	// 确保路径以/开头
	if !strings.HasPrefix(path, "/") {
		path = "/" + path
	}

	return basePath + path
}
