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

package survey

import (
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"cdk-office/internal/config"
	"cdk-office/internal/db"
	"cdk-office/internal/storage"
)

var (
	handler     *SurveyHandler
	difyService *DifyService
)

// InitSurveyModule 初始化Survey模块
func InitSurveyModule() error {
	// 初始化存储服务
	storageConfig := &storage.StorageConfig{
		Primary: config.Config.Storage.Primary,
		CloudDB: storage.CloudDBConfig{
			Provider:    config.Config.Storage.CloudDB.Provider,
			URL:         config.Config.Storage.CloudDB.URL,
			AnonKey:     config.Config.Storage.CloudDB.AnonKey,
			ServiceKey:  config.Config.Storage.CloudDB.ServiceKey,
			Bucket:      config.Config.Storage.CloudDB.Bucket,
			MaxFileSize: config.Config.Storage.CloudDB.MaxFileSize,
			MaxStorage:  config.Config.Storage.CloudDB.MaxStorage,
		},
		S3: storage.S3Config{
			Providers: make([]storage.S3Provider, len(config.Config.Storage.S3.Providers)),
		},
		WebDAV: storage.WebDAVConfig{
			URL:      config.Config.Storage.WebDAV.URL,
			Username: config.Config.Storage.WebDAV.Username,
			Password: config.Config.Storage.WebDAV.Password,
			BasePath: config.Config.Storage.WebDAV.BasePath,
		},
		Local: storage.LocalConfig{
			Path:    config.Config.Storage.Local.Path,
			MaxSize: config.Config.Storage.Local.MaxSize,
			BaseURL: config.Config.Storage.Local.BaseURL,
		},
		Global: storage.GlobalConfig{
			MaxFileSize:   config.Config.Storage.Global.MaxFileSize,
			AllowedTypes:  config.Config.Storage.Global.AllowedTypes,
			EnableCleanup: config.Config.Storage.Global.EnableCleanup,
			CleanupDays:   config.Config.Storage.Global.CleanupDays,
		},
	}

	// 复制S3配置
	for i, provider := range config.Config.Storage.S3.Providers {
		storageConfig.S3.Providers[i] = storage.S3Provider{
			Name:      provider.Name,
			AccessKey: provider.AccessKey,
			SecretKey: provider.SecretKey,
			Bucket:    provider.Bucket,
			Endpoint:  provider.Endpoint,
			Region:    provider.Region,
			UseSSL:    provider.UseSSL,
		}
	}

	storageService, err := storage.NewStorageService(storageConfig)
	if err != nil {
		return err
	}

	// 初始化Dify服务
	if config.Config.Dify.APIKey != "" && config.Config.Dify.SurveyAnalysisWorkflowID != "" {
		difyConfig := DifyConfig{
			BaseURL:                  config.Config.Dify.APIEndpoint,
			APIKey:                   config.Config.Dify.APIKey,
			SurveyAnalysisWorkflowID: config.Config.Dify.SurveyAnalysisWorkflowID,
			KnowledgeBaseID:          config.Config.Dify.KnowledgeBaseID,
		}
		difyService = NewDifyService(difyConfig, db.GetDB())
	}

	// 创建处理器
	handler = NewSurveyHandler(db.GetDB(), storageService)
	handler.difyService = difyService

	return nil
}

// RegisterRoutes 注册Survey路由
func RegisterRoutes(router *gin.RouterGroup) {
	// 确保模块已初始化
	if handler == nil {
		if err := InitSurveyModule(); err != nil {
			panic("Failed to initialize Survey module: " + err.Error())
		}
	}

	// 注册路由
	handler.RegisterRoutes(router)
}

// GetHandler 获取Survey处理器（用于测试）
func GetHandler() *SurveyHandler {
	return handler
}

// RegisterPublicRoutes 注册Survey公开访问路由（无需认证）
func RegisterPublicRoutes(router *gin.RouterGroup) {
	// 确保模块已初始化
	if handler == nil {
		if err := InitSurveyModule(); err != nil {
			panic("Failed to initialize Survey module: " + err.Error())
		}
	}

	// 注册公开访问路由
	handler.RegisterPublicRoutes(router)
}

// GetHandler 获取Survey处理器（用于测试）
func GetHandler() *SurveyHandler {
	return handler
}

// GetDifyService 获取Dify服务（用于测试）
func GetDifyService() *DifyService {
	return difyService
}