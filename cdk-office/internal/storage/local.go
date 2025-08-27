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
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"syscall"
	"time"
)

// LocalProvider 本地存储提供商
type LocalProvider struct {
	config  LocalConfig
	basePath string
	maxSize int64
}

// NewLocalProvider 创建本地存储提供商
func NewLocalProvider(config LocalConfig) (*LocalProvider, error) {
	basePath := config.Path
	if basePath == "" {
		basePath = "./storage"
	}
	
	// 创建存储目录
	if err := os.MkdirAll(basePath, 0755); err != nil {
		return nil, fmt.Errorf("failed to create storage directory: %v", err)
	}
	
	provider := &LocalProvider{
		config:   config,
		basePath: basePath,
		maxSize:  parseSize(config.MaxSize),
	}
	
	return provider, nil
}

// Upload 上传文件到本地存储
func (p *LocalProvider) Upload(file *multipart.FileHeader, path string) (*FileInfo, error) {
	// 检查存储空间
	quota, err := p.GetQuota()
	if err != nil {
		return nil, fmt.Errorf("failed to check storage quota: %v", err)
	}
	
	if !quota.Unlimited && quota.Available < file.Size {
		return nil, fmt.Errorf("insufficient storage space")
	}
	
	// 构建完整文件路径
	fullPath := filepath.Join(p.basePath, path)
	dir := filepath.Dir(fullPath)
	
	// 创建目录
	if err := os.MkdirAll(dir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create directory: %v", err)
	}
	
	// 打开上传的文件
	src, err := file.Open()
	if err != nil {
		return nil, fmt.Errorf("failed to open source file: %v", err)
	}
	defer src.Close()
	
	// 创建目标文件
	dst, err := os.Create(fullPath)
	if err != nil {
		return nil, fmt.Errorf("failed to create destination file: %v", err)
	}
	defer dst.Close()
	
	// 复制文件内容
	if _, err := io.Copy(dst, src); err != nil {
		os.Remove(fullPath) // 清理失败的文件
		return nil, fmt.Errorf("failed to copy file content: %v", err)
	}
	
	// 获取文件信息
	stat, err := dst.Stat()
	if err != nil {
		return nil, fmt.Errorf("failed to get file stats: %v", err)
	}
	
	// 生成访问URL
	url := p.generateURL(path)
	
	return &FileInfo{
		Path:         path,
		Name:         file.Filename,
		Size:         stat.Size(),
		MimeType:     file.Header.Get("Content-Type"),
		URL:          url,
		Provider:     p.GetProviderName(),
		LastModified: stat.ModTime(),
	}, nil
}

// Download 从本地存储下载文件
func (p *LocalProvider) Download(path string) (io.ReadCloser, error) {
	fullPath := filepath.Join(p.basePath, path)
	
	// 检查文件是否存在
	if _, err := os.Stat(fullPath); os.IsNotExist(err) {
		return nil, fmt.Errorf("file not found: %s", path)
	}
	
	// 打开文件
	file, err := os.Open(fullPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open file: %v", err)
	}
	
	return file, nil
}

// Delete 删除本地存储的文件
func (p *LocalProvider) Delete(path string) error {
	fullPath := filepath.Join(p.basePath, path)
	
	// 检查文件是否存在
	if _, err := os.Stat(fullPath); os.IsNotExist(err) {
		return nil // 文件不存在，视为删除成功
	}
	
	// 删除文件
	if err := os.Remove(fullPath); err != nil {
		return fmt.Errorf("failed to delete file: %v", err)
	}
	
	// 尝试删除空目录
	p.cleanupEmptyDirs(filepath.Dir(fullPath))
	
	return nil
}

// GetURL 获取文件访问URL
func (p *LocalProvider) GetURL(path string) (string, error) {
	fullPath := filepath.Join(p.basePath, path)
	
	// 检查文件是否存在
	if _, err := os.Stat(fullPath); os.IsNotExist(err) {
		return "", fmt.Errorf("file not found: %s", path)
	}
	
	return p.generateURL(path), nil
}

// GetQuota 获取本地存储空间信息
func (p *LocalProvider) GetQuota() (*QuotaInfo, error) {
	var stat syscall.Statfs_t
	err := syscall.Statfs(p.basePath, &stat)
	if err != nil {
		return nil, fmt.Errorf("failed to get filesystem stats: %v", err)
	}
	
	// 计算磁盘空间信息
	blockSize := stat.Bsize
	totalSpace := int64(stat.Blocks) * blockSize
	freeSpace := int64(stat.Bavail) * blockSize
	
	// 如果配置了最大空间限制
	if p.maxSize > 0 {
		used, err := p.calculateUsedSpace()
		if err != nil {
			return nil, err
		}
		
		available := p.maxSize - used
		if available < 0 {
			available = 0
		}
		
		return &QuotaInfo{
			Total:     p.maxSize,
			Used:      used,
			Available: available,
			Unlimited: false,
		}, nil
	}
	
	// 使用磁盘空间信息
	usedSpace := totalSpace - freeSpace
	
	return &QuotaInfo{
		Total:     totalSpace,
		Used:      usedSpace,
		Available: freeSpace,
		Unlimited: false,
	}, nil
}

// ListFiles 列出文件
func (p *LocalProvider) ListFiles(prefix string) ([]*FileInfo, error) {
	searchPath := filepath.Join(p.basePath, prefix)
	var files []*FileInfo
	
	err := filepath.Walk(searchPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		
		if info.IsDir() {
			return nil
		}
		
		// 计算相对路径
		relPath, err := filepath.Rel(p.basePath, path)
		if err != nil {
			return err
		}
		
		// 生成文件信息
		fileInfo := &FileInfo{
			Path:         relPath,
			Name:         info.Name(),
			Size:         info.Size(),
			URL:          p.generateURL(relPath),
			Provider:     p.GetProviderName(),
			LastModified: info.ModTime(),
		}
		
		files = append(files, fileInfo)
		return nil
	})
	
	return files, err
}

// GetProviderName 获取提供商名称
func (p *LocalProvider) GetProviderName() string {
	return "local"
}

// IsAvailable 检查提供商是否可用
func (p *LocalProvider) IsAvailable() bool {
	// 检查存储目录是否可写
	testFile := filepath.Join(p.basePath, ".test")
	file, err := os.Create(testFile)
	if err != nil {
		return false
	}
	file.Close()
	os.Remove(testFile)
	
	return true
}

// generateURL 生成文件访问URL
func (p *LocalProvider) generateURL(path string) string {
	baseURL := p.config.BaseURL
	if baseURL == "" {
		baseURL = "/api/files"
	}
	
	return fmt.Sprintf("%s/%s", strings.TrimSuffix(baseURL, "/"), path)
}

// calculateUsedSpace 计算已使用的存储空间
func (p *LocalProvider) calculateUsedSpace() (int64, error) {
	var totalSize int64
	
	err := filepath.Walk(p.basePath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		
		if !info.IsDir() {
			totalSize += info.Size()
		}
		
		return nil
	})
	
	return totalSize, err
}

// cleanupEmptyDirs 清理空目录
func (p *LocalProvider) cleanupEmptyDirs(dir string) {
	// 不要删除根存储目录
	if dir == p.basePath {
		return
	}
	
	// 检查目录是否为空
	entries, err := os.ReadDir(dir)
	if err != nil || len(entries) > 0 {
		return
	}
	
	// 删除空目录
	if err := os.Remove(dir); err == nil {
		// 递归清理父目录
		p.cleanupEmptyDirs(filepath.Dir(dir))
	}
}

// parseSize 解析大小字符串的辅助函数
func parseSize(sizeStr string) int64 {
	if sizeStr == "" {
		return 0
	}
	
	sizeStr = strings.ToUpper(strings.TrimSpace(sizeStr))
	
	// 提取数字和单位
	var numStr string
	var unit string
	
	for i, r := range sizeStr {
		if r >= '0' && r <= '9' || r == '.' {
			numStr += string(r)
		} else {
			unit = sizeStr[i:]
			break
		}
	}
	
	num, err := strconv.ParseFloat(numStr, 64)
	if err != nil {
		return 0
	}
	
	// 转换单位
	switch unit {
	case "B", "BYTES":
		return int64(num)
	case "KB", "K":
		return int64(num * 1024)
	case "MB", "M":
		return int64(num * 1024 * 1024)
	case "GB", "G":
		return int64(num * 1024 * 1024 * 1024)
	case "TB", "T":
		return int64(num * 1024 * 1024 * 1024 * 1024)
	default:
		return int64(num) // 默认为字节
	}
}