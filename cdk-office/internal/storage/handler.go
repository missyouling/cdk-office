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
	"net/http"
	"path/filepath"
	"strconv"

	"github.com/gin-gonic/gin"
)

// StorageHandler 存储管理API处理器
type StorageHandler struct {
	storageService *StorageService
}

// NewStorageHandler 创建存储处理器
func NewStorageHandler(storageService *StorageService) *StorageHandler {
	return &StorageHandler{
		storageService: storageService,
	}
}

// UploadFile 上传文件
// @Summary 上传文件
// @Description 上传文件到混合云存储
// @Tags storage
// @Accept multipart/form-data
// @Produce json
// @Param file formData file true "上传的文件"
// @Param path formData string false "文件路径"
// @Success 200 {object} FileInfo
// @Failure 400 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /api/storage/upload [post]
func (h *StorageHandler) UploadFile(c *gin.Context) {
	// 获取上传的文件
	file, err := c.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "No file uploaded"})
		return
	}

	// 获取文件路径
	path := c.PostForm("path")
	if path == "" {
		// 使用原始文件名生成路径
		path = generateUniqueFilename(file.Filename)
	}

	// 上传文件
	fileInfo, err := h.storageService.Upload(file, path)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, fileInfo)
}

// DownloadFile 下载文件
// @Summary 下载文件
// @Description 从混合云存储下载文件
// @Tags storage
// @Produce octet-stream
// @Param path query string true "文件路径"
// @Param provider query string false "存储提供商"
// @Success 200 {file} binary
// @Failure 400 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /api/storage/download [get]
func (h *StorageHandler) DownloadFile(c *gin.Context) {
	path := c.Query("path")
	if path == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Path is required"})
		return
	}

	provider := c.Query("provider")

	// 下载文件
	reader, err := h.storageService.Download(path, provider)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}
	defer reader.Close()

	// 设置响应头
	filename := filepath.Base(path)
	c.Header("Content-Disposition", "attachment; filename=\""+filename+"\"")
	c.Header("Content-Type", "application/octet-stream")

	// 流式传输文件内容
	c.DataFromReader(http.StatusOK, -1, "application/octet-stream", reader, nil)
}

// DeleteFile 删除文件
// @Summary 删除文件
// @Description 从混合云存储删除文件
// @Tags storage
// @Produce json
// @Param path query string true "文件路径"
// @Param provider query string false "存储提供商"
// @Success 200 {object} map[string]string
// @Failure 400 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /api/storage/delete [delete]
func (h *StorageHandler) DeleteFile(c *gin.Context) {
	path := c.Query("path")
	if path == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Path is required"})
		return
	}

	provider := c.Query("provider")

	// 删除文件
	err := h.storageService.Delete(path, provider)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "File deleted successfully"})
}

// GetFileURL 获取文件访问URL
// @Summary 获取文件URL
// @Description 获取文件的访问URL
// @Tags storage
// @Produce json
// @Param path query string true "文件路径"
// @Param provider query string false "存储提供商"
// @Success 200 {object} map[string]string
// @Failure 400 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /api/storage/url [get]
func (h *StorageHandler) GetFileURL(c *gin.Context) {
	path := c.Query("path")
	if path == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Path is required"})
		return
	}

	provider := c.Query("provider")

	// 获取文件URL
	url, err := h.storageService.GetURL(path, provider)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"url": url})
}

// GetStorageInfo 获取存储信息
// @Summary 获取存储信息
// @Description 获取所有存储提供商的配额和使用情况
// @Tags storage
// @Produce json
// @Success 200 {object} map[string]QuotaInfo
// @Router /api/storage/info [get]
func (h *StorageHandler) GetStorageInfo(c *gin.Context) {
	info := h.storageService.GetStorageInfo()
	c.JSON(http.StatusOK, info)
}

// ListFiles 列出文件
// @Summary 列出文件
// @Description 列出指定路径下的文件
// @Tags storage
// @Produce json
// @Param prefix query string false "路径前缀"
// @Param provider query string false "存储提供商"
// @Param page query int false "页码" default(1)
// @Param limit query int false "每页数量" default(20)
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /api/storage/files [get]
func (h *StorageHandler) ListFiles(c *gin.Context) {
	prefix := c.Query("prefix")
	provider := c.Query("provider")

	// 分页参数
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))
	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 20
	}

	var allFiles []*FileInfo
	var err error

	if provider != "" {
		// 从指定提供商列出文件
		if p, exists := h.storageService.allProviders[provider]; exists {
			allFiles, err = p.ListFiles(prefix)
		} else {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Provider not found"})
			return
		}
	} else {
		// 从所有提供商列出文件
		for _, p := range h.storageService.allProviders {
			files, err := p.ListFiles(prefix)
			if err == nil {
				allFiles = append(allFiles, files...)
			}
		}
	}

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// 分页处理
	total := len(allFiles)
	start := (page - 1) * limit
	end := start + limit

	if start >= total {
		allFiles = []*FileInfo{}
	} else {
		if end > total {
			end = total
		}
		allFiles = allFiles[start:end]
	}

	c.JSON(http.StatusOK, gin.H{
		"files": allFiles,
		"pagination": gin.H{
			"page":  page,
			"limit": limit,
			"total": total,
		},
	})
}

// TestProviders 测试存储提供商
// @Summary 测试存储提供商
// @Description 测试所有存储提供商的可用性
// @Tags storage
// @Produce json
// @Success 200 {object} map[string]interface{}
// @Router /api/storage/test [get]
func (h *StorageHandler) TestProviders(c *gin.Context) {
	results := make(map[string]interface{})

	for name, provider := range h.storageService.allProviders {
		available := provider.IsAvailable()
		providerInfo := gin.H{
			"name":      provider.GetProviderName(),
			"available": available,
		}

		// 如果可用，获取配额信息
		if available {
			if quota, err := provider.GetQuota(); err == nil {
				providerInfo["quota"] = quota
			}
		}

		results[name] = providerInfo
	}

	c.JSON(http.StatusOK, gin.H{
		"providers": results,
		"primary":   h.storageService.primaryStore.GetProviderName(),
	})
}

// SwitchPrimaryProvider 切换主存储提供商
// @Summary 切换主存储提供商
// @Description 切换到指定的存储提供商作为主存储
// @Tags storage
// @Accept json
// @Produce json
// @Param request body map[string]string true "请求参数"
// @Success 200 {object} map[string]string
// @Failure 400 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /api/storage/switch [post]
func (h *StorageHandler) SwitchPrimaryProvider(c *gin.Context) {
	var req struct {
		Provider string `json:"provider" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// 检查提供商是否存在
	provider, exists := h.storageService.allProviders[req.Provider]
	if !exists {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Provider not found"})
		return
	}

	// 检查提供商是否可用
	if !provider.IsAvailable() {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Provider is not available"})
		return
	}

	// 切换主存储提供商
	oldPrimary := h.storageService.primaryStore.GetProviderName()
	h.storageService.primaryStore = provider

	// 重新设置备用存储
	h.storageService.backupStores = make([]StorageProvider, 0)
	for name, p := range h.storageService.allProviders {
		if name != req.Provider && p.IsAvailable() {
			h.storageService.backupStores = append(h.storageService.backupStores, p)
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"message":     "Primary provider switched successfully",
		"old_primary": oldPrimary,
		"new_primary": provider.GetProviderName(),
	})
}
