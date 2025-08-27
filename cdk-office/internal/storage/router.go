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
	"github.com/gin-gonic/gin"
)

// RegisterStorageRoutes 注册存储相关路由
func RegisterStorageRoutes(router *gin.RouterGroup, storageService *StorageService) {
	handler := NewStorageHandler(storageService)

	// 存储管理路由组
	storage := router.Group("/storage")
	{
		// 文件操作
		storage.POST("/upload", handler.UploadFile)    // 上传文件
		storage.GET("/download", handler.DownloadFile) // 下载文件
		storage.DELETE("/delete", handler.DeleteFile)  // 删除文件
		storage.GET("/url", handler.GetFileURL)        // 获取文件URL
		storage.GET("/files", handler.ListFiles)       // 列出文件

		// 存储管理
		storage.GET("/info", handler.GetStorageInfo)           // 获取存储信息
		storage.GET("/test", handler.TestProviders)            // 测试存储提供商
		storage.POST("/switch", handler.SwitchPrimaryProvider) // 切换主存储提供商
	}
}
