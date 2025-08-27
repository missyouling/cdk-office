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
	"context"
	"fmt"
	"io"
	"mime/multipart"
	"strings"
	"time"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

// S3Provider S3兼容存储提供商
type S3Provider struct {
	config S3Provider
	client *minio.Client
	name   string
}

// NewS3Provider 创建S3兼容存储提供商
func NewS3Provider(config S3Provider) (*S3Provider, error) {
	if config.AccessKey == "" || config.SecretKey == "" || config.Bucket == "" {
		return nil, fmt.Errorf("S3 access key, secret key and bucket are required")
	}

	// 创建S3客户端
	client, err := minio.New(config.Endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(config.AccessKey, config.SecretKey, ""),
		Secure: config.UseSSL,
		Region: config.Region,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create S3 client: %v", err)
	}

	provider := &S3Provider{
		config: config,
		client: client,
		name:   config.Name,
	}

	// 测试连接并确保桶存在
	if err := provider.ensureBucketExists(); err != nil {
		return nil, fmt.Errorf("failed to ensure bucket exists: %v", err)
	}

	return provider, nil
}

// ensureBucketExists 确保存储桶存在
func (p *S3Provider) ensureBucketExists() error {
	ctx := context.Background()

	// 检查桶是否存在
	exists, err := p.client.BucketExists(ctx, p.config.Bucket)
	if err != nil {
		return fmt.Errorf("failed to check bucket existence: %v", err)
	}

	// 如果桶不存在，尝试创建（某些S3服务可能不允许）
	if !exists {
		err = p.client.MakeBucket(ctx, p.config.Bucket, minio.MakeBucketOptions{
			Region: p.config.Region,
		})
		if err != nil {
			return fmt.Errorf("bucket does not exist and failed to create: %v", err)
		}
	}

	return nil
}

// Upload 上传文件到S3存储
func (p *S3Provider) Upload(file *multipart.FileHeader, path string) (*FileInfo, error) {
	ctx := context.Background()

	// 打开文件
	src, err := file.Open()
	if err != nil {
		return nil, fmt.Errorf("failed to open file: %v", err)
	}
	defer src.Close()

	// 上传文件
	_, err = p.client.PutObject(ctx, p.config.Bucket, path, src, file.Size, minio.PutObjectOptions{
		ContentType: file.Header.Get("Content-Type"),
	})
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

// Download 从S3存储下载文件
func (p *S3Provider) Download(path string) (io.ReadCloser, error) {
	ctx := context.Background()

	object, err := p.client.GetObject(ctx, p.config.Bucket, path, minio.GetObjectOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to get object: %v", err)
	}

	return object, nil
}

// Delete 删除S3存储的文件
func (p *S3Provider) Delete(path string) error {
	ctx := context.Background()

	err := p.client.RemoveObject(ctx, p.config.Bucket, path, minio.RemoveObjectOptions{})
	if err != nil {
		return fmt.Errorf("failed to delete object: %v", err)
	}

	return nil
}

// GetURL 获取文件访问URL
func (p *S3Provider) GetURL(path string) (string, error) {
	ctx := context.Background()

	// 生成预签名URL（24小时有效）
	url, err := p.client.PresignedGetObject(ctx, p.config.Bucket, path, 24*time.Hour, nil)
	if err != nil {
		return "", fmt.Errorf("failed to generate presigned URL: %v", err)
	}

	return url.String(), nil
}

// GetQuota 获取S3存储空间信息
func (p *S3Provider) GetQuota() (*QuotaInfo, error) {
	// S3存储通常认为是无限制的（由配额和计费控制）
	return &QuotaInfo{
		Total:     0,
		Used:      0,
		Available: 0,
		Unlimited: true,
	}, nil
}

// ListFiles 列出S3存储中的文件
func (p *S3Provider) ListFiles(prefix string) ([]*FileInfo, error) {
	ctx := context.Background()

	objectCh := p.client.ListObjects(ctx, p.config.Bucket, minio.ListObjectsOptions{
		Prefix:    prefix,
		Recursive: true,
	})

	var files []*FileInfo
	for object := range objectCh {
		if object.Err != nil {
			return nil, fmt.Errorf("error listing objects: %v", object.Err)
		}

		// 生成访问URL
		url, err := p.GetURL(object.Key)
		if err != nil {
			continue // 跳过无法生成URL的文件
		}

		files = append(files, &FileInfo{
			Path:         object.Key,
			Name:         extractFileName(object.Key),
			Size:         object.Size,
			MimeType:     "", // S3 API可能不返回MIME类型
			URL:          url,
			Provider:     p.GetProviderName(),
			LastModified: object.LastModified,
		})
	}

	return files, nil
}

// GetProviderName 获取提供商名称
func (p *S3Provider) GetProviderName() string {
	if p.name != "" {
		return fmt.Sprintf("s3_%s", p.name)
	}
	return "s3"
}

// IsAvailable 检查S3提供商是否可用
func (p *S3Provider) IsAvailable() bool {
	ctx := context.Background()

	// 尝试列出桶来测试连接
	_, err := p.client.BucketExists(ctx, p.config.Bucket)
	return err == nil
}

// extractFileName 从路径中提取文件名
func extractFileName(path string) string {
	parts := strings.Split(path, "/")
	return parts[len(parts)-1]
}
