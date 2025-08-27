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
	"errors"
	"fmt"
	"io"
	"log"
	"mime/multipart"
	"path/filepath"
	"strings"
	"time"
)

// StorageService 存储服务管理器
type StorageService struct {
	config       *StorageConfig
	primaryStore StorageProvider
	backupStores []StorageProvider
	allProviders map[string]StorageProvider
}

// NewStorageService 创建新的存储服务
func NewStorageService(config *StorageConfig) (*StorageService, error) {
	service := &StorageService{
		config:       config,
		backupStores: make([]StorageProvider, 0),
		allProviders: make(map[string]StorageProvider),
	}

	// 初始化所有提供商
	if err := service.initProviders(); err != nil {
		return nil, fmt.Errorf("failed to initialize providers: %v", err)
	}

	// 设置主存储提供商
	if err := service.setPrimaryProvider(); err != nil {
		return nil, fmt.Errorf("failed to set primary provider: %v", err)
	}

	// 设置备用存储提供商
	service.setBackupProviders()

	return service, nil
}

// initProviders 初始化所有存储提供商
func (s *StorageService) initProviders() error {
	// 初始化本地存储
	if localProvider, err := NewLocalProvider(s.config.Local); err == nil {
		s.allProviders["local"] = localProvider
	}

	// 初始化云数据库存储
	if s.config.CloudDB.URL != "" {
		if cloudDBProvider, err := NewCloudDBProvider(s.config.CloudDB); err == nil {
			s.allProviders["cloud_db"] = cloudDBProvider
		}
	}

	// 初始化S3存储提供商
	for _, s3Config := range s.config.S3.Providers {
		if s3Provider, err := NewS3Provider(s3Config); err == nil {
			s.allProviders["s3_"+s3Config.Name] = s3Provider
		}
	}

	// 初始化WebDAV存储
	if s.config.WebDAV.URL != "" {
		if webdavProvider, err := NewWebDAVProvider(s.config.WebDAV); err == nil {
			s.allProviders["webdav"] = webdavProvider
		}
	}

	// 初始化云盘存储
	for _, driveConfig := range s.config.CloudDrive.Providers {
		if driveProvider, err := NewCloudDriveProvider(driveConfig); err == nil {
			s.allProviders["drive_"+driveConfig.Name] = driveProvider
		}
	}

	return nil
}

// setPrimaryProvider 设置主存储提供商
func (s *StorageService) setPrimaryProvider() error {
	primary := s.config.Primary
	if primary == "" {
		primary = "local" // 默认使用本地存储
	}

	if provider, exists := s.allProviders[primary]; exists && provider.IsAvailable() {
		s.primaryStore = provider
		return nil
	}

	// 如果主存储不可用，尝试使用其他可用的存储
	for name, provider := range s.allProviders {
		if provider.IsAvailable() {
			log.Printf("Primary storage %s not available, fallback to %s", primary, name)
			s.primaryStore = provider
			return nil
		}
	}

	return errors.New("no available storage provider found")
}

// setBackupProviders 设置备用存储提供商
func (s *StorageService) setBackupProviders() {
	for name, provider := range s.allProviders {
		if provider != s.primaryStore && provider.IsAvailable() {
			s.backupStores = append(s.backupStores, provider)
			log.Printf("Added backup storage provider: %s", name)
		}
	}
}

// Upload 上传文件，自动选择最佳存储提供商
func (s *StorageService) Upload(file *multipart.FileHeader, path string) (*FileInfo, error) {
	// 验证文件
	if err := s.validateFile(file); err != nil {
		return nil, err
	}

	// 生成唯一文件路径
	if path == "" {
		path = s.generateFilePath(file.Filename)
	}

	// 尝试使用主存储
	if s.primaryStore != nil {
		quota, err := s.primaryStore.GetQuota()
		if err == nil && (quota.Unlimited || quota.Available > file.Size) {
			info, err := s.primaryStore.Upload(file, path)
			if err == nil {
				log.Printf("File uploaded to primary storage: %s", s.primaryStore.GetProviderName())
				return info, nil
			}
			log.Printf("Primary storage upload failed: %v", err)
		} else {
			log.Printf("Primary storage quota insufficient: %v", err)
		}
	}

	// 尝试使用备用存储
	for _, provider := range s.backupStores {
		quota, err := provider.GetQuota()
		if err == nil && (quota.Unlimited || quota.Available > file.Size) {
			info, err := provider.Upload(file, path)
			if err == nil {
				log.Printf("File uploaded to backup storage: %s", provider.GetProviderName())
				return info, nil
			}
			log.Printf("Backup storage upload failed: %v", err)
		}
	}

	return nil, errors.New("all storage providers failed or have insufficient space")
}

// Download 下载文件
func (s *StorageService) Download(path string, provider string) (io.ReadCloser, error) {
	// 如果指定了提供商，直接使用
	if provider != "" {
		if p, exists := s.allProviders[provider]; exists {
			return p.Download(path)
		}
		return nil, fmt.Errorf("provider %s not found", provider)
	}

	// 尝试从主存储下载
	if s.primaryStore != nil {
		reader, err := s.primaryStore.Download(path)
		if err == nil {
			return reader, nil
		}
	}

	// 尝试从备用存储下载
	for _, p := range s.backupStores {
		reader, err := p.Download(path)
		if err == nil {
			return reader, nil
		}
	}

	return nil, errors.New("file not found in any storage provider")
}

// Delete 删除文件
func (s *StorageService) Delete(path string, provider string) error {
	// 如果指定了提供商，只删除指定提供商的文件
	if provider != "" {
		if p, exists := s.allProviders[provider]; exists {
			return p.Delete(path)
		}
		return fmt.Errorf("provider %s not found", provider)
	}

	// 从所有提供商中删除文件
	var lastErr error
	deleted := false

	for _, p := range s.allProviders {
		err := p.Delete(path)
		if err == nil {
			deleted = true
		} else {
			lastErr = err
		}
	}

	if !deleted && lastErr != nil {
		return lastErr
	}

	return nil
}

// GetURL 获取文件访问URL
func (s *StorageService) GetURL(path string, provider string) (string, error) {
	// 如果指定了提供商，直接使用
	if provider != "" {
		if p, exists := s.allProviders[provider]; exists {
			return p.GetURL(path)
		}
		return "", fmt.Errorf("provider %s not found", provider)
	}

	// 尝试从主存储获取URL
	if s.primaryStore != nil {
		url, err := s.primaryStore.GetURL(path)
		if err == nil {
			return url, nil
		}
	}

	// 尝试从备用存储获取URL
	for _, p := range s.backupStores {
		url, err := p.GetURL(path)
		if err == nil {
			return url, nil
		}
	}

	return "", errors.New("file not found in any storage provider")
}

// GetStorageInfo 获取存储信息
func (s *StorageService) GetStorageInfo() map[string]*QuotaInfo {
	info := make(map[string]*QuotaInfo)

	for name, provider := range s.allProviders {
		if quota, err := provider.GetQuota(); err == nil {
			info[name] = quota
		}
	}

	return info
}

// validateFile 验证文件
func (s *StorageService) validateFile(file *multipart.FileHeader) error {
	// 检查文件大小
	maxSize := parseSize(s.config.Global.MaxFileSize)
	if maxSize > 0 && file.Size > maxSize {
		return fmt.Errorf("file size %d exceeds maximum allowed size %d", file.Size, maxSize)
	}

	// 检查文件类型
	if len(s.config.Global.AllowedTypes) > 0 {
		ext := strings.ToLower(filepath.Ext(file.Filename))
		allowed := false
		for _, allowedType := range s.config.Global.AllowedTypes {
			if ext == allowedType || strings.HasSuffix(allowedType, ext) {
				allowed = true
				break
			}
		}
		if !allowed {
			return fmt.Errorf("file type %s is not allowed", ext)
		}
	}

	return nil
}

// generateFilePath 生成文件路径
func (s *StorageService) generateFilePath(filename string) string {
	ext := filepath.Ext(filename)
	name := strings.TrimSuffix(filename, ext)
	timestamp := time.Now().Format("20060102150405")
	return fmt.Sprintf("surveys/%s_%s%s", name, timestamp, ext)
}
