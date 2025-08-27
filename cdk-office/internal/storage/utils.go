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
	"crypto/md5"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

// parseSize 解析大小字符串（如 "100MB", "1GB"）
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

// formatSize 格式化文件大小为可读字符串
func formatSize(size int64) string {
	const unit = 1024
	if size < unit {
		return fmt.Sprintf("%d B", size)
	}
	div, exp := int64(unit), 0
	for n := size / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %cB", float64(size)/float64(div), "KMGTPE"[exp])
}

// generateUniqueFilename 生成唯一文件名
func generateUniqueFilename(originalName string) string {
	ext := filepath.Ext(originalName)
	name := strings.TrimSuffix(originalName, ext)
	timestamp := time.Now().Format("20060102150405")

	// 添加时间戳确保唯一性
	return fmt.Sprintf("%s_%s%s", name, timestamp, ext)
}

// generateHashedPath 生成基于文件内容哈希的路径
func generateHashedPath(filename string, content []byte) string {
	hash := md5.Sum(content)
	hashStr := fmt.Sprintf("%x", hash)

	// 创建分层目录结构
	dir1 := hashStr[:2]
	dir2 := hashStr[2:4]

	ext := filepath.Ext(filename)
	name := strings.TrimSuffix(filename, ext)

	return fmt.Sprintf("%s/%s/%s_%s%s", dir1, dir2, name, hashStr[:8], ext)
}

// sanitizeFilename 清理文件名，移除不安全字符
func sanitizeFilename(filename string) string {
	// 替换不安全字符
	filename = strings.ReplaceAll(filename, " ", "_")
	filename = strings.ReplaceAll(filename, "/", "_")
	filename = strings.ReplaceAll(filename, "\\", "_")
	filename = strings.ReplaceAll(filename, ":", "_")
	filename = strings.ReplaceAll(filename, "*", "_")
	filename = strings.ReplaceAll(filename, "?", "_")
	filename = strings.ReplaceAll(filename, "\"", "_")
	filename = strings.ReplaceAll(filename, "<", "_")
	filename = strings.ReplaceAll(filename, ">", "_")
	filename = strings.ReplaceAll(filename, "|", "_")

	// 限制文件名长度
	if len(filename) > 255 {
		ext := filepath.Ext(filename)
		name := strings.TrimSuffix(filename, ext)
		name = name[:255-len(ext)]
		filename = name + ext
	}

	return filename
}

// getMimeType 根据文件扩展名获取MIME类型
func getMimeType(filename string) string {
	ext := strings.ToLower(filepath.Ext(filename))

	mimeTypes := map[string]string{
		".txt":  "text/plain",
		".html": "text/html",
		".css":  "text/css",
		".js":   "application/javascript",
		".json": "application/json",
		".xml":  "application/xml",
		".pdf":  "application/pdf",
		".doc":  "application/msword",
		".docx": "application/vnd.openxmlformats-officedocument.wordprocessingml.document",
		".xls":  "application/vnd.ms-excel",
		".xlsx": "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet",
		".ppt":  "application/vnd.ms-powerpoint",
		".pptx": "application/vnd.openxmlformats-officedocument.presentationml.presentation",
		".zip":  "application/zip",
		".rar":  "application/x-rar-compressed",
		".7z":   "application/x-7z-compressed",
		".tar":  "application/x-tar",
		".gz":   "application/gzip",
		".jpg":  "image/jpeg",
		".jpeg": "image/jpeg",
		".png":  "image/png",
		".gif":  "image/gif",
		".bmp":  "image/bmp",
		".webp": "image/webp",
		".svg":  "image/svg+xml",
		".ico":  "image/x-icon",
		".mp3":  "audio/mpeg",
		".wav":  "audio/wav",
		".ogg":  "audio/ogg",
		".mp4":  "video/mp4",
		".avi":  "video/x-msvideo",
		".mov":  "video/quicktime",
		".wmv":  "video/x-ms-wmv",
		".flv":  "video/x-flv",
		".webm": "video/webm",
	}

	if mimeType, exists := mimeTypes[ext]; exists {
		return mimeType
	}

	return "application/octet-stream"
}

// isValidFileType 检查文件类型是否被允许
func isValidFileType(filename string, allowedTypes []string) bool {
	if len(allowedTypes) == 0 {
		return true // 如果没有限制，则允许所有类型
	}

	ext := strings.ToLower(filepath.Ext(filename))

	for _, allowedType := range allowedTypes {
		allowedType = strings.ToLower(allowedType)
		if ext == allowedType || strings.HasSuffix(allowedType, ext) {
			return true
		}
	}

	return false
}

// calculateDirectorySize 计算目录大小
func calculateDirectorySize(path string) (int64, error) {
	var size int64

	err := filepath.Walk(path, func(_ string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() {
			size += info.Size()
		}
		return nil
	})

	return size, err
}

// cleanupPath 清理路径，确保路径格式正确
func cleanupPath(path string) string {
	// 替换Windows路径分隔符
	path = strings.ReplaceAll(path, "\\", "/")

	// 移除重复的斜杠
	for strings.Contains(path, "//") {
		path = strings.ReplaceAll(path, "//", "/")
	}

	// 移除开头的斜杠（除非是根路径）
	if path != "/" && strings.HasPrefix(path, "/") {
		path = strings.TrimPrefix(path, "/")
	}

	// 移除结尾的斜杠
	path = strings.TrimSuffix(path, "/")

	return path
}

// validateStorageConfig 验证存储配置
func validateStorageConfig(config *StorageConfig) error {
	if config == nil {
		return fmt.Errorf("storage config is nil")
	}

	// 检查主存储配置
	if config.Primary == "" {
		config.Primary = "local" // 默认使用本地存储
	}

	// 验证本地存储配置
	if config.Primary == "local" && config.Local.Path == "" {
		config.Local.Path = "./storage" // 默认存储路径
	}

	// 验证云数据库配置
	if config.Primary == "cloud_db" {
		if config.CloudDB.URL == "" || config.CloudDB.AnonKey == "" {
			return fmt.Errorf("cloud database requires URL and anon_key")
		}
	}

	// 验证S3配置
	if config.Primary == "s3" && len(config.S3.Providers) == 0 {
		return fmt.Errorf("s3 storage requires at least one provider")
	}

	// 验证WebDAV配置
	if config.Primary == "webdav" && config.WebDAV.URL == "" {
		return fmt.Errorf("webdav storage requires URL")
	}

	// 设置默认全局配置
	if config.Global.MaxFileSize == "" {
		config.Global.MaxFileSize = "100MB"
	}

	if len(config.Global.AllowedTypes) == 0 {
		config.Global.AllowedTypes = []string{
			".jpg", ".jpeg", ".png", ".gif", ".pdf",
			".doc", ".docx", ".xls", ".xlsx", ".ppt", ".pptx",
			".txt", ".csv", ".zip", ".rar",
		}
	}

	return nil
}
