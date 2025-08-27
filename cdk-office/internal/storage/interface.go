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
	"io"
	"mime/multipart"
	"time"
)

// StorageProvider 存储提供商接口
type StorageProvider interface {
	// Upload 上传文件
	Upload(file *multipart.FileHeader, path string) (*FileInfo, error)
	
	// Download 下载文件
	Download(path string) (io.ReadCloser, error)
	
	// Delete 删除文件
	Delete(path string) error
	
	// GetURL 获取文件访问URL
	GetURL(path string) (string, error)
	
	// GetQuota 获取存储空间配额信息
	GetQuota() (*QuotaInfo, error)
	
	// ListFiles 列出文件
	ListFiles(prefix string) ([]*FileInfo, error)
	
	// GetProviderName 获取提供商名称
	GetProviderName() string
	
	// IsAvailable 检查提供商是否可用
	IsAvailable() bool
}

// FileInfo 文件信息
type FileInfo struct {
	Path         string    `json:"path"`          // 文件路径
	Name         string    `json:"name"`          // 文件名
	Size         int64     `json:"size"`          // 文件大小
	MimeType     string    `json:"mime_type"`     // MIME类型
	URL          string    `json:"url"`           // 访问URL
	Provider     string    `json:"provider"`      // 存储提供商
	LastModified time.Time `json:"last_modified"` // 最后修改时间
}

// QuotaInfo 存储配额信息
type QuotaInfo struct {
	Total     int64 `json:"total"`     // 总容量 (bytes)
	Used      int64 `json:"used"`      // 已使用 (bytes)
	Available int64 `json:"available"` // 可用空间 (bytes)
	Unlimited bool  `json:"unlimited"` // 是否无限制
}

// UploadOptions 上传选项
type UploadOptions struct {
	MaxFileSize   int64    `json:"max_file_size"`   // 最大文件大小
	AllowedTypes  []string `json:"allowed_types"`   // 允许的MIME类型
	GenerateName  bool     `json:"generate_name"`   // 是否生成新文件名
	OverwriteMode string   `json:"overwrite_mode"`  // 覆盖模式: error, overwrite, rename
}

// StorageConfig 存储配置
type StorageConfig struct {
	// 主存储策略
	Primary string `yaml:"primary" json:"primary"` // cloud_db, s3, local, webdav
	
	// 云数据库存储配置
	CloudDB CloudDBConfig `yaml:"cloud_db" json:"cloud_db"`
	
	// S3兼容存储配置
	S3 S3Config `yaml:"s3" json:"s3"`
	
	// 云盘存储配置
	CloudDrive CloudDriveConfig `yaml:"cloud_drive" json:"cloud_drive"`
	
	// WebDAV存储配置
	WebDAV WebDAVConfig `yaml:"webdav" json:"webdav"`
	
	// 本地存储配置
	Local LocalConfig `yaml:"local" json:"local"`
	
	// 全局配置
	Global GlobalConfig `yaml:"global" json:"global"`
}

// CloudDBConfig 云数据库存储配置
type CloudDBConfig struct {
	Provider    string `yaml:"provider" json:"provider"`       // supabase, memfire
	URL         string `yaml:"url" json:"url"`                 // API URL
	AnonKey     string `yaml:"anon_key" json:"anon_key"`       // 匿名密钥
	ServiceKey  string `yaml:"service_key" json:"service_key"` // 服务密钥
	Bucket      string `yaml:"bucket" json:"bucket"`           // 存储桶名称
	MaxFileSize string `yaml:"max_file_size" json:"max_file_size"` // 最大文件大小
	MaxStorage  string `yaml:"max_storage" json:"max_storage"`     // 最大存储空间
}

// S3Config S3兼容存储配置
type S3Config struct {
	Providers []S3Provider `yaml:"providers" json:"providers"`
}

// S3Provider S3提供商配置
type S3Provider struct {
	Name      string `yaml:"name" json:"name"`           // 提供商名称
	AccessKey string `yaml:"access_key" json:"access_key"` // 访问密钥
	SecretKey string `yaml:"secret_key" json:"secret_key"` // 秘密密钥
	Bucket    string `yaml:"bucket" json:"bucket"`       // 存储桶
	Endpoint  string `yaml:"endpoint" json:"endpoint"`   // 端点URL
	Region    string `yaml:"region" json:"region"`       // 区域
	UseSSL    bool   `yaml:"use_ssl" json:"use_ssl"`     // 是否使用SSL
}

// CloudDriveConfig 云盘存储配置
type CloudDriveConfig struct {
	Providers []CloudDriveProvider `yaml:"providers" json:"providers"`
}

// CloudDriveProvider 云盘提供商配置
type CloudDriveProvider struct {
	Name         string `yaml:"name" json:"name"`                   // 提供商名称
	ClientID     string `yaml:"client_id" json:"client_id"`         // 客户端ID
	ClientSecret string `yaml:"client_secret" json:"client_secret"` // 客户端密钥
	RefreshToken string `yaml:"refresh_token" json:"refresh_token"` // 刷新令牌
	FolderPath   string `yaml:"folder_path" json:"folder_path"`     // 文件夹路径
}

// WebDAVConfig WebDAV存储配置
type WebDAVConfig struct {
	URL      string `yaml:"url" json:"url"`           // WebDAV URL
	Username string `yaml:"username" json:"username"` // 用户名
	Password string `yaml:"password" json:"password"` // 密码
	BasePath string `yaml:"base_path" json:"base_path"` // 基础路径
}

// LocalConfig 本地存储配置
type LocalConfig struct {
	Path    string `yaml:"path" json:"path"`         // 存储路径
	MaxSize string `yaml:"max_size" json:"max_size"` // 最大存储空间
	BaseURL string `yaml:"base_url" json:"base_url"` // 基础访问URL
}

// GlobalConfig 全局存储配置
type GlobalConfig struct {
	MaxFileSize   string   `yaml:"max_file_size" json:"max_file_size"`     // 全局最大文件大小
	AllowedTypes  []string `yaml:"allowed_types" json:"allowed_types"`     // 全局允许的文件类型
	EnableCleanup bool     `yaml:"enable_cleanup" json:"enable_cleanup"`   // 是否启用清理
	CleanupDays   int      `yaml:"cleanup_days" json:"cleanup_days"`       // 清理天数
}