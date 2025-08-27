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

package contract

import (
	"crypto/sha256"
	"fmt"
	"io"
	"io/ioutil"
	"mime/multipart"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/linux-do/cdk-office/internal/config"
)

// FileService 文件存储服务
type FileService struct {
	config *config.FileStorageConfig
}

// FileInfo 文件信息
type FileInfo struct {
	ID       string    `json:"id"`
	Name     string    `json:"name"`
	Path     string    `json:"path"`
	URL      string    `json:"url"`
	Size     int64     `json:"size"`
	MimeType string    `json:"mime_type"`
	Hash     string    `json:"hash"`
	UploadAt time.Time `json:"upload_at"`
}

// StorageProvider 存储提供商接口
type StorageProvider interface {
	Save(filePath string, data []byte) error
	Get(filePath string) ([]byte, error)
	Delete(filePath string) error
	GetURL(filePath string) string
	Exists(filePath string) bool
}

// LocalStorageProvider 本地存储提供商
type LocalStorageProvider struct {
	basePath string
	baseURL  string
}

// OSSStorageProvider 阿里云OSS存储提供商
type OSSStorageProvider struct {
	endpoint        string
	accessKeyID     string
	accessKeySecret string
	bucket          string
}

// COSStorageProvider 腾讯云COS存储提供商
type COSStorageProvider struct {
	region    string
	secretID  string
	secretKey string
	bucket    string
}

// NewFileService 创建文件存储服务
func NewFileService() *FileService {
	return &FileService{
		config: &config.Config.FileStorage,
	}
}

// getStorageProvider 获取存储提供商
func (s *FileService) getStorageProvider() StorageProvider {
	switch s.config.Provider {
	case "oss":
		return &OSSStorageProvider{
			endpoint:        s.config.OSSEndpoint,
			accessKeyID:     s.config.OSSAccessKeyID,
			accessKeySecret: s.config.OSSAccessKeySecret,
			bucket:          s.config.OSSBucket,
		}
	case "cos":
		return &COSStorageProvider{
			region:    s.config.COSRegion,
			secretID:  s.config.COSSecretID,
			secretKey: s.config.COSSecretKey,
			bucket:    s.config.COSBucket,
		}
	default: // local
		return &LocalStorageProvider{
			basePath: s.config.LocalPath,
			baseURL:  s.config.BasePath,
		}
	}
}

// SaveFile 保存文件
func (s *FileService) SaveFile(file *multipart.FileHeader, category string) (*FileInfo, error) {
	// 检查文件大小
	if file.Size > s.config.MaxFileSize*1024*1024 {
		return nil, fmt.Errorf("文件大小超过限制: %dMB", s.config.MaxFileSize)
	}

	// 打开文件
	src, err := file.Open()
	if err != nil {
		return nil, fmt.Errorf("打开文件失败: %w", err)
	}
	defer src.Close()

	// 读取文件内容
	data, err := ioutil.ReadAll(src)
	if err != nil {
		return nil, fmt.Errorf("读取文件失败: %w", err)
	}

	return s.SaveFileData(data, file.Filename, category)
}

// SaveFileData 保存文件数据
func (s *FileService) SaveFileData(data []byte, fileName, category string) (*FileInfo, error) {
	// 生成文件ID和路径
	fileID := uuid.New().String()
	ext := filepath.Ext(fileName)
	if ext == "" {
		ext = ".bin"
	}

	// 构建文件路径: category/yyyy/mm/dd/uuid.ext
	now := time.Now()
	relativePath := filepath.Join(
		category,
		fmt.Sprintf("%04d", now.Year()),
		fmt.Sprintf("%02d", now.Month()),
		fmt.Sprintf("%02d", now.Day()),
		fileID+ext,
	)

	// 计算文件哈希
	hash := fmt.Sprintf("%x", sha256.Sum256(data))

	// 获取存储提供商并保存文件
	provider := s.getStorageProvider()
	if err := provider.Save(relativePath, data); err != nil {
		return nil, fmt.Errorf("保存文件失败: %w", err)
	}

	// 构建文件信息
	fileInfo := &FileInfo{
		ID:       fileID,
		Name:     fileName,
		Path:     relativePath,
		URL:      provider.GetURL(relativePath),
		Size:     int64(len(data)),
		MimeType: s.getMimeType(fileName),
		Hash:     hash,
		UploadAt: now,
	}

	return fileInfo, nil
}

// GetFile 获取文件
func (s *FileService) GetFile(filePath string) ([]byte, error) {
	provider := s.getStorageProvider()
	if !provider.Exists(filePath) {
		return nil, fmt.Errorf("文件不存在: %s", filePath)
	}

	data, err := provider.Get(filePath)
	if err != nil {
		return nil, fmt.Errorf("获取文件失败: %w", err)
	}

	return data, nil
}

// DeleteFile 删除文件
func (s *FileService) DeleteFile(filePath string) error {
	provider := s.getStorageProvider()
	if err := provider.Delete(filePath); err != nil {
		return fmt.Errorf("删除文件失败: %w", err)
	}

	return nil
}

// GetFileURL 获取文件URL
func (s *FileService) GetFileURL(filePath string) string {
	provider := s.getStorageProvider()
	return provider.GetURL(filePath)
}

// SaveContractFile 保存合同文件
func (s *FileService) SaveContractFile(contractID, fileURL string) (string, error) {
	// 从URL读取文件内容
	data, err := s.readFileFromURL(fileURL)
	if err != nil {
		return "", fmt.Errorf("读取文件失败: %w", err)
	}

	// 保存到合同目录
	fileName := fmt.Sprintf("contract_%s.pdf", contractID)
	fileInfo, err := s.SaveFileData(data, fileName, "contracts")
	if err != nil {
		return "", err
	}

	return fileInfo.ID, nil
}

// getMimeType 获取MIME类型
func (s *FileService) getMimeType(fileName string) string {
	ext := strings.ToLower(filepath.Ext(fileName))
	switch ext {
	case ".pdf":
		return "application/pdf"
	case ".doc":
		return "application/msword"
	case ".docx":
		return "application/vnd.openxmlformats-officedocument.wordprocessingml.document"
	case ".jpg", ".jpeg":
		return "image/jpeg"
	case ".png":
		return "image/png"
	case ".gif":
		return "image/gif"
	case ".txt":
		return "text/plain"
	case ".html":
		return "text/html"
	case ".json":
		return "application/json"
	case ".xml":
		return "application/xml"
	default:
		return "application/octet-stream"
	}
}

// readFileFromURL 从URL读取文件（占位符实现）
func (s *FileService) readFileFromURL(fileURL string) ([]byte, error) {
	// 这里应该实现从URL读取文件的逻辑
	// 对于本地文件，直接读取文件系统
	// 对于HTTP URL，使用HTTP客户端下载
	// 暂时返回模拟数据
	return []byte("mock file content"), nil
}

// 本地存储提供商实现

// Save 保存文件到本地
func (p *LocalStorageProvider) Save(filePath string, data []byte) error {
	fullPath := filepath.Join(p.basePath, filePath)
	
	// 确保目录存在
	dir := filepath.Dir(fullPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("创建目录失败: %w", err)
	}

	// 写入文件
	if err := ioutil.WriteFile(fullPath, data, 0644); err != nil {
		return fmt.Errorf("写入文件失败: %w", err)
	}

	return nil
}

// Get 从本地获取文件
func (p *LocalStorageProvider) Get(filePath string) ([]byte, error) {
	fullPath := filepath.Join(p.basePath, filePath)
	data, err := ioutil.ReadFile(fullPath)
	if err != nil {
		return nil, fmt.Errorf("读取文件失败: %w", err)
	}
	return data, nil
}

// Delete 删除本地文件
func (p *LocalStorageProvider) Delete(filePath string) error {
	fullPath := filepath.Join(p.basePath, filePath)
	if err := os.Remove(fullPath); err != nil {
		return fmt.Errorf("删除文件失败: %w", err)
	}
	return nil
}

// GetURL 获取本地文件URL
func (p *LocalStorageProvider) GetURL(filePath string) string {
	return p.baseURL + "/" + strings.ReplaceAll(filePath, "\\", "/")
}

// Exists 检查本地文件是否存在
func (p *LocalStorageProvider) Exists(filePath string) bool {
	fullPath := filepath.Join(p.basePath, filePath)
	_, err := os.Stat(fullPath)
	return err == nil
}

// 阿里云OSS存储提供商实现（占位符）

// Save 保存文件到阿里云OSS
func (p *OSSStorageProvider) Save(filePath string, data []byte) error {
	// 这里应该实现阿里云OSS的文件上传逻辑
	// 使用 github.com/aliyun/aliyun-oss-go-sdk/oss
	return fmt.Errorf("OSS存储暂未实现")
}

// Get 从阿里云OSS获取文件
func (p *OSSStorageProvider) Get(filePath string) ([]byte, error) {
	return nil, fmt.Errorf("OSS存储暂未实现")
}

// Delete 删除阿里云OSS文件
func (p *OSSStorageProvider) Delete(filePath string) error {
	return fmt.Errorf("OSS存储暂未实现")
}

// GetURL 获取阿里云OSS文件URL
func (p *OSSStorageProvider) GetURL(filePath string) string {
	return fmt.Sprintf("https://%s.%s/%s", p.bucket, p.endpoint, filePath)
}

// Exists 检查阿里云OSS文件是否存在
func (p *OSSStorageProvider) Exists(filePath string) bool {
	return false
}

// 腾讯云COS存储提供商实现（占位符）

// Save 保存文件到腾讯云COS
func (p *COSStorageProvider) Save(filePath string, data []byte) error {
	// 这里应该实现腾讯云COS的文件上传逻辑
	// 使用 github.com/tencentyun/cos-go-sdk-v5
	return fmt.Errorf("COS存储暂未实现")
}

// Get 从腾讯云COS获取文件
func (p *COSStorageProvider) Get(filePath string) ([]byte, error) {
	return nil, fmt.Errorf("COS存储暂未实现")
}

// Delete 删除腾讯云COS文件
func (p *COSStorageProvider) Delete(filePath string) error {
	return fmt.Errorf("COS存储暂未实现")
}

// GetURL 获取腾讯云COS文件URL
func (p *COSStorageProvider) GetURL(filePath string) string {
	return fmt.Sprintf("https://%s.cos.%s.myqcloud.com/%s", p.bucket, p.region, filePath)
}

// Exists 检查腾讯云COS文件是否存在
func (p *COSStorageProvider) Exists(filePath string) bool {
	return false
}