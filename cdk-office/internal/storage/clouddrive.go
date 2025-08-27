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
	"context"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"path"
	"strings"
	"time"

	"golang.org/x/oauth2"
	"google.golang.org/api/drive/v3"
	"google.golang.org/api/option"
)

// CloudDriveProvider 云盘存储提供商
type CloudDriveProvider struct {
	config       CloudDriveProvider
	httpClient   *http.Client
	driveService *drive.Service
	providerType string
}

// NewCloudDriveProvider 创建云盘存储提供商
func NewCloudDriveProvider(config CloudDriveProvider) (*CloudDriveProvider, error) {
	provider := &CloudDriveProvider{
		config:       config,
		providerType: config.Name,
	}

	switch strings.ToLower(config.Name) {
	case "google_drive":
		return provider.initGoogleDrive()
	case "aliyun_drive":
		return provider.initAliyunDrive()
	default:
		return nil, fmt.Errorf("unsupported cloud drive provider: %s", config.Name)
	}
}

// initGoogleDrive 初始化Google Drive
func (p *CloudDriveProvider) initGoogleDrive() (*CloudDriveProvider, error) {
	if p.config.ClientID == "" || p.config.ClientSecret == "" || p.config.RefreshToken == "" {
		return nil, fmt.Errorf("Google Drive requires client_id, client_secret and refresh_token")
	}

	// 配置OAuth2
	oauth2Config := &oauth2.Config{
		ClientID:     p.config.ClientID,
		ClientSecret: p.config.ClientSecret,
		Scopes:       []string{drive.DriveFileScope},
		Endpoint: oauth2.Endpoint{
			AuthURL:  "https://accounts.google.com/o/oauth2/auth",
			TokenURL: "https://oauth2.googleapis.com/token",
		},
	}

	// 创建Token
	token := &oauth2.Token{
		RefreshToken: p.config.RefreshToken,
		TokenType:    "Bearer",
	}

	// 创建HTTP客户端
	ctx := context.Background()
	p.httpClient = oauth2Config.Client(ctx, token)

	// 创建Drive服务
	driveService, err := drive.NewService(ctx, option.WithHTTPClient(p.httpClient))
	if err != nil {
		return nil, fmt.Errorf("failed to create Google Drive service: %v", err)
	}

	p.driveService = driveService

	// 测试连接
	if err := p.testGoogleDriveConnection(); err != nil {
		return nil, fmt.Errorf("Google Drive connection test failed: %v", err)
	}

	return p, nil
}

// initAliyunDrive 初始化阿里云盘
func (p *CloudDriveProvider) initAliyunDrive() (*CloudDriveProvider, error) {
	if p.config.RefreshToken == "" {
		return nil, fmt.Errorf("Aliyun Drive requires refresh_token")
	}

	// 创建HTTP客户端
	p.httpClient = &http.Client{
		Timeout: 30 * time.Second,
	}

	// 测试连接
	if err := p.testAliyunDriveConnection(); err != nil {
		return nil, fmt.Errorf("Aliyun Drive connection test failed: %v", err)
	}

	return p, nil
}

// testGoogleDriveConnection 测试Google Drive连接
func (p *CloudDriveProvider) testGoogleDriveConnection() error {
	_, err := p.driveService.About.Get().Fields("user").Do()
	return err
}

// testAliyunDriveConnection 测试阿里云盘连接
func (p *CloudDriveProvider) testAliyunDriveConnection() error {
	// 获取访问令牌
	_, err := p.getAliyunAccessToken()
	return err
}

// Upload 上传文件到云盘存储
func (p *CloudDriveProvider) Upload(file *multipart.FileHeader, path string) (*FileInfo, error) {
	switch p.providerType {
	case "google_drive":
		return p.uploadToGoogleDrive(file, path)
	case "aliyun_drive":
		return p.uploadToAliyunDrive(file, path)
	default:
		return nil, fmt.Errorf("unsupported provider type: %s", p.providerType)
	}
}

// uploadToGoogleDrive 上传文件到Google Drive
func (p *CloudDriveProvider) uploadToGoogleDrive(file *multipart.FileHeader, path string) (*FileInfo, error) {
	// 打开文件
	src, err := file.Open()
	if err != nil {
		return nil, fmt.Errorf("failed to open file: %v", err)
	}
	defer src.Close()

	// 获取或创建文件夹
	folderID, err := p.getOrCreateGoogleDriveFolder(p.config.FolderPath)
	if err != nil {
		return nil, fmt.Errorf("failed to get folder: %v", err)
	}

	// 创建文件元数据
	driveFile := &drive.File{
		Name:    file.Filename,
		Parents: []string{folderID},
	}

	// 上传文件
	result, err := p.driveService.Files.Create(driveFile).Media(src).Do()
	if err != nil {
		return nil, fmt.Errorf("failed to upload file: %v", err)
	}

	// 设置文件为公开可读（可选）
	permission := &drive.Permission{
		Type: "anyone",
		Role: "reader",
	}
	_, err = p.driveService.Permissions.Create(result.Id, permission).Do()
	if err != nil {
		// 忽略权限设置错误，文件已上传成功
	}

	return &FileInfo{
		Path:         path,
		Name:         file.Filename,
		Size:         file.Size,
		MimeType:     file.Header.Get("Content-Type"),
		URL:          fmt.Sprintf("https://drive.google.com/file/d/%s/view", result.Id),
		Provider:     p.GetProviderName(),
		LastModified: time.Now(),
	}, nil
}

// uploadToAliyunDrive 上传文件到阿里云盘
func (p *CloudDriveProvider) uploadToAliyunDrive(file *multipart.FileHeader, path string) (*FileInfo, error) {
	// 获取访问令牌
	accessToken, err := p.getAliyunAccessToken()
	if err != nil {
		return nil, fmt.Errorf("failed to get access token: %v", err)
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

	// 创建上传任务（这是简化实现，实际阿里云盘API更复杂）
	uploadURL := "https://api.aliyundrive.com/v2/file/create"

	createReq := map[string]interface{}{
		"drive_id":        "default",
		"parent_file_id":  "root",
		"name":            file.Filename,
		"type":            "file",
		"check_name_mode": "auto_rename",
		"size":            len(content),
	}

	jsonData, _ := json.Marshal(createReq)

	req, err := http.NewRequest("POST", uploadURL, bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %v", err)
	}

	req.Header.Set("Authorization", "Bearer "+accessToken)
	req.Header.Set("Content-Type", "application/json")

	resp, err := p.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to create upload task: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		return nil, fmt.Errorf("upload task creation failed with status: %d", resp.StatusCode)
	}

	// 注意：这是简化实现，实际需要处理阿里云盘的分片上传流程

	return &FileInfo{
		Path:         path,
		Name:         file.Filename,
		Size:         file.Size,
		MimeType:     file.Header.Get("Content-Type"),
		URL:          "", // 阿里云盘文件URL需要特殊处理
		Provider:     p.GetProviderName(),
		LastModified: time.Now(),
	}, nil
}

// Download 从云盘存储下载文件
func (p *CloudDriveProvider) Download(path string) (io.ReadCloser, error) {
	return nil, fmt.Errorf("cloud drive download not implemented yet")
}

// Delete 删除云盘存储的文件
func (p *CloudDriveProvider) Delete(path string) error {
	return fmt.Errorf("cloud drive delete not implemented yet")
}

// GetURL 获取文件访问URL
func (p *CloudDriveProvider) GetURL(path string) (string, error) {
	return "", fmt.Errorf("cloud drive URL generation not implemented yet")
}

// GetQuota 获取云盘存储空间信息
func (p *CloudDriveProvider) GetQuota() (*QuotaInfo, error) {
	switch p.providerType {
	case "google_drive":
		return p.getGoogleDriveQuota()
	case "aliyun_drive":
		return p.getAliyunDriveQuota()
	default:
		return &QuotaInfo{Unlimited: true}, nil
	}
}

// getGoogleDriveQuota 获取Google Drive配额
func (p *CloudDriveProvider) getGoogleDriveQuota() (*QuotaInfo, error) {
	about, err := p.driveService.About.Get().Fields("storageQuota").Do()
	if err != nil {
		return nil, fmt.Errorf("failed to get quota: %v", err)
	}

	quota := about.StorageQuota
	return &QuotaInfo{
		Total:     quota.Limit,
		Used:      quota.Usage,
		Available: quota.Limit - quota.Usage,
		Unlimited: quota.Limit == 0,
	}, nil
}

// getAliyunDriveQuota 获取阿里云盘配额
func (p *CloudDriveProvider) getAliyunDriveQuota() (*QuotaInfo, error) {
	// 阿里云盘配额查询需要特殊API
	return &QuotaInfo{Unlimited: true}, nil
}

// ListFiles 列出云盘存储中的文件
func (p *CloudDriveProvider) ListFiles(prefix string) ([]*FileInfo, error) {
	return nil, fmt.Errorf("cloud drive file listing not implemented yet")
}

// GetProviderName 获取提供商名称
func (p *CloudDriveProvider) GetProviderName() string {
	return fmt.Sprintf("drive_%s", p.config.Name)
}

// IsAvailable 检查云盘提供商是否可用
func (p *CloudDriveProvider) IsAvailable() bool {
	switch p.providerType {
	case "google_drive":
		return p.testGoogleDriveConnection() == nil
	case "aliyun_drive":
		return p.testAliyunDriveConnection() == nil
	default:
		return false
	}
}

// getOrCreateGoogleDriveFolder 获取或创建Google Drive文件夹
func (p *CloudDriveProvider) getOrCreateGoogleDriveFolder(folderPath string) (string, error) {
	if folderPath == "" || folderPath == "/" {
		return "root", nil
	}

	// 简化实现：直接搜索文件夹名称
	folderName := path.Base(folderPath)

	query := fmt.Sprintf("name='%s' and mimeType='application/vnd.google-apps.folder' and trashed=false", folderName)
	fileList, err := p.driveService.Files.List().Q(query).Do()
	if err != nil {
		return "", err
	}

	if len(fileList.Files) > 0 {
		return fileList.Files[0].Id, nil
	}

	// 创建文件夹
	folder := &drive.File{
		Name:     folderName,
		MimeType: "application/vnd.google-apps.folder",
		Parents:  []string{"root"},
	}

	result, err := p.driveService.Files.Create(folder).Do()
	if err != nil {
		return "", err
	}

	return result.Id, nil
}

// getAliyunAccessToken 获取阿里云盘访问令牌
func (p *CloudDriveProvider) getAliyunAccessToken() (string, error) {
	// 使用refresh_token获取access_token
	tokenURL := "https://auth.aliyundrive.com/v2/account/token"

	reqData := map[string]string{
		"refresh_token": p.config.RefreshToken,
		"grant_type":    "refresh_token",
	}

	jsonData, _ := json.Marshal(reqData)

	req, err := http.NewRequest("POST", tokenURL, bytes.NewBuffer(jsonData))
	if err != nil {
		return "", err
	}

	req.Header.Set("Content-Type", "application/json")

	resp, err := p.httpClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("failed to get access token, status: %d", resp.StatusCode)
	}

	var tokenResp struct {
		AccessToken string `json:"access_token"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&tokenResp); err != nil {
		return "", err
	}

	return tokenResp.AccessToken, nil
}
