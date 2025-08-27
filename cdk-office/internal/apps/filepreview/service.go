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

package filepreview

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"path/filepath"
	"strings"
	"time"

	"gorm.io/gorm"

	"github.com/linux-do/cdk-office/internal/models"
)

// Service 文件预览服务
type Service struct {
	db       *gorm.DB
	config   *Config
	provider PreviewProvider
}

// Config 文件预览服务配置
type Config struct {
	Provider   string           `json:"provider"` // dify, kkfileview
	DifyURL    string           `json:"dify_url"`
	KKFileView KKFileViewConfig `json:"kkfileview"`
}

// KKFileViewConfig KKFileView配置
type KKFileViewConfig struct {
	Enabled bool   `json:"enabled"`
	URL     string `json:"url"`
	Timeout int    `json:"timeout"` // 秒
}

// PreviewProvider 文件预览提供者接口
type PreviewProvider interface {
	Preview(ctx context.Context, req *PreviewRequest) (*PreviewResponse, error)
	GetPreviewURL(documentID, fileType string) string
	SupportedTypes() []string
	HealthCheck() error
}

// PreviewRequest 预览请求
type PreviewRequest struct {
	DocumentID string `json:"document_id"`
	FileName   string `json:"file_name"`
	FileType   string `json:"file_type"`
	FileURL    string `json:"file_url"`
	UserID     string `json:"user_id"`
	TeamID     string `json:"team_id"`
}

// PreviewResponse 预览响应
type PreviewResponse struct {
	PreviewURL   string                 `json:"preview_url"`
	ThumbnailURL string                 `json:"thumbnail_url,omitempty"`
	Provider     string                 `json:"provider"`
	Supported    bool                   `json:"supported"`
	Metadata     map[string]interface{} `json:"metadata,omitempty"`
}

// NewService 创建文件预览服务
func NewService(config *Config, db *gorm.DB) *Service {
	var provider PreviewProvider

	switch config.Provider {
	case "kkfileview":
		if config.KKFileView.Enabled {
			provider = NewKKFileViewProvider(&config.KKFileView)
		} else {
			log.Println("KKFileView is disabled, falling back to Dify")
			provider = NewDifyPreviewProvider(config.DifyURL)
		}
	default:
		provider = NewDifyPreviewProvider(config.DifyURL)
	}

	return &Service{
		db:       db,
		config:   config,
		provider: provider,
	}
}

// Preview 预览文件
func (s *Service) Preview(ctx context.Context, req *PreviewRequest) (*PreviewResponse, error) {
	// 记录预览请求
	previewRecord := &models.FilePreview{
		DocumentID: req.DocumentID,
		FileName:   req.FileName,
		FileType:   req.FileType,
		UserID:     req.UserID,
		TeamID:     req.TeamID,
		Provider:   s.config.Provider,
		Status:     "processing",
	}

	if err := s.db.Create(previewRecord).Error; err != nil {
		log.Printf("Failed to create preview record: %v", err)
	}

	// 调用预览提供者
	response, err := s.provider.Preview(ctx, req)
	if err != nil {
		previewRecord.Status = "failed"
		previewRecord.ErrorMessage = err.Error()
		s.db.Save(previewRecord)
		return nil, fmt.Errorf("preview failed: %w", err)
	}

	// 更新记录
	previewRecord.Status = "completed"
	previewRecord.PreviewURL = response.PreviewURL
	s.db.Save(previewRecord)

	log.Printf("File preview generated: %s (provider: %s)", req.DocumentID, response.Provider)

	return response, nil
}

// GetSupportedTypes 获取支持的文件类型
func (s *Service) GetSupportedTypes() []string {
	return s.provider.SupportedTypes()
}

// HealthCheck 健康检查
func (s *Service) HealthCheck() error {
	return s.provider.HealthCheck()
}

// DifyPreviewProvider Dify原生预览提供者
type DifyPreviewProvider struct {
	baseURL string
}

// NewDifyPreviewProvider 创建Dify预览提供者
func NewDifyPreviewProvider(baseURL string) *DifyPreviewProvider {
	return &DifyPreviewProvider{
		baseURL: baseURL,
	}
}

// Preview 实现Dify预览
func (p *DifyPreviewProvider) Preview(ctx context.Context, req *PreviewRequest) (*PreviewResponse, error) {
	previewURL := p.GetPreviewURL(req.DocumentID, req.FileType)

	return &PreviewResponse{
		PreviewURL: previewURL,
		Provider:   "dify",
		Supported:  p.isSupported(req.FileType),
	}, nil
}

// GetPreviewURL 获取预览URL
func (p *DifyPreviewProvider) GetPreviewURL(documentID, fileType string) string {
	return fmt.Sprintf("%s/api/documents/%s/preview", p.baseURL, documentID)
}

// SupportedTypes 支持的文件类型
func (p *DifyPreviewProvider) SupportedTypes() []string {
	return []string{
		"pdf", "txt", "md", "doc", "docx",
		"jpg", "jpeg", "png", "gif", "bmp",
	}
}

// HealthCheck 健康检查
func (p *DifyPreviewProvider) HealthCheck() error {
	// 简单的健康检查实现
	return nil
}

// isSupported 检查文件类型是否支持
func (p *DifyPreviewProvider) isSupported(fileType string) bool {
	supportedTypes := p.SupportedTypes()
	fileType = strings.ToLower(fileType)

	for _, t := range supportedTypes {
		if t == fileType {
			return true
		}
	}
	return false
}

// KKFileViewProvider KKFileView预览提供者
type KKFileViewProvider struct {
	config     *KKFileViewConfig
	httpClient *http.Client
}

// NewKKFileViewProvider 创建KKFileView预览提供者
func NewKKFileViewProvider(config *KKFileViewConfig) *KKFileViewProvider {
	timeout := 30
	if config.Timeout > 0 {
		timeout = config.Timeout
	}

	return &KKFileViewProvider{
		config: config,
		httpClient: &http.Client{
			Timeout: time.Duration(timeout) * time.Second,
		},
	}
}

// Preview 实现KKFileView预览
func (p *KKFileViewProvider) Preview(ctx context.Context, req *PreviewRequest) (*PreviewResponse, error) {
	if !p.config.Enabled {
		return nil, fmt.Errorf("KKFileView is disabled")
	}

	if !p.isSupported(req.FileType) {
		return &PreviewResponse{
			Provider:  "kkfileview",
			Supported: false,
		}, nil
	}

	// 构建预览URL
	previewURL := p.GetPreviewURL(req.FileURL, req.FileType)

	// 生成缩略图URL（如果支持）
	thumbnailURL := ""
	if p.supportsThumbnail(req.FileType) {
		thumbnailURL = p.getThumbnailURL(req.FileURL)
	}

	return &PreviewResponse{
		PreviewURL:   previewURL,
		ThumbnailURL: thumbnailURL,
		Provider:     "kkfileview",
		Supported:    true,
		Metadata: map[string]interface{}{
			"enhanced": true,
			"features": []string{"zoom", "download", "print", "fullscreen"},
		},
	}, nil
}

// GetPreviewURL 获取KKFileView预览URL
func (p *KKFileViewProvider) GetPreviewURL(fileURL, fileType string) string {
	encodedURL := url.QueryEscape(fileURL)
	return fmt.Sprintf("%s/onlinePreview?url=%s", p.config.URL, encodedURL)
}

// getThumbnailURL 获取缩略图URL
func (p *KKFileViewProvider) getThumbnailURL(fileURL string) string {
	encodedURL := url.QueryEscape(fileURL)
	return fmt.Sprintf("%s/picturesPreview?url=%s", p.config.URL, encodedURL)
}

// SupportedTypes KKFileView支持的文件类型
func (p *KKFileViewProvider) SupportedTypes() []string {
	return []string{
		// Office文档
		"doc", "docx", "xls", "xlsx", "ppt", "pptx",
		// PDF文档
		"pdf",
		// 文本文件
		"txt", "md", "xml", "json", "csv",
		// 图片文件
		"jpg", "jpeg", "png", "gif", "bmp", "tiff",
		// 视频文件
		"mp4", "avi", "mov", "wmv", "flv", "mkv",
		// 音频文件
		"mp3", "wav", "aac", "flac",
		// 压缩文件
		"zip", "rar", "7z", "tar", "gz",
		// 其他格式
		"dwg", "dxf", "psd", "eps",
	}
}

// HealthCheck 健康检查
func (p *KKFileViewProvider) HealthCheck() error {
	if !p.config.Enabled {
		return fmt.Errorf("KKFileView is disabled")
	}

	// 检查KKFileView服务是否可用
	resp, err := p.httpClient.Get(p.config.URL + "/index")
	if err != nil {
		return fmt.Errorf("KKFileView service unavailable: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("KKFileView service unhealthy: status %d", resp.StatusCode)
	}

	return nil
}

// isSupported 检查文件类型是否支持
func (p *KKFileViewProvider) isSupported(fileType string) bool {
	supportedTypes := p.SupportedTypes()
	fileType = strings.ToLower(fileType)

	// 去除文件扩展名中的点
	fileType = strings.TrimPrefix(fileType, ".")

	for _, t := range supportedTypes {
		if t == fileType {
			return true
		}
	}
	return false
}

// supportsThumbnail 检查是否支持缩略图
func (p *KKFileViewProvider) supportsThumbnail(fileType string) bool {
	thumbnailTypes := []string{
		"jpg", "jpeg", "png", "gif", "bmp", "tiff",
		"pdf", "doc", "docx", "ppt", "pptx",
	}

	fileType = strings.ToLower(fileType)
	fileType = strings.TrimPrefix(fileType, ".")

	for _, t := range thumbnailTypes {
		if t == fileType {
			return true
		}
	}
	return false
}

// GetFileExtension 获取文件扩展名
func GetFileExtension(filename string) string {
	ext := filepath.Ext(filename)
	if ext != "" {
		ext = strings.ToLower(ext[1:]) // 移除点号并转换为小写
	}
	return ext
}

// IsPreviewSupported 检查文件是否支持预览
func (s *Service) IsPreviewSupported(filename string) bool {
	fileType := GetFileExtension(filename)
	supportedTypes := s.GetSupportedTypes()

	for _, t := range supportedTypes {
		if t == fileType {
			return true
		}
	}
	return false
}

// GetPreviewHistory 获取用户的预览历史
func (s *Service) GetPreviewHistory(userID string, page, limit int) ([]models.FilePreview, error) {
	var previews []models.FilePreview
	offset := (page - 1) * limit

	err := s.db.Where("user_id = ?", userID).
		Order("created_at DESC").
		Offset(offset).
		Limit(limit).
		Find(&previews).Error

	return previews, err
}

// CleanupPreviewHistory 清理预览历史记录
func (s *Service) CleanupPreviewHistory(olderThanDays int) error {
	cutoffDate := time.Now().AddDate(0, 0, -olderThanDays)

	result := s.db.Where("created_at < ?", cutoffDate).Delete(&models.FilePreview{})
	if result.Error != nil {
		return result.Error
	}

	log.Printf("Cleaned up %d preview history records older than %d days", result.RowsAffected, olderThanDays)
	return nil
}
