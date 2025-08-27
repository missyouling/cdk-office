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

package isolation

import (
	"log"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"github.com/linux-do/cdk-office/internal/apps/isolation"
	"github.com/linux-do/cdk-office/internal/middleware"
	"github.com/linux-do/cdk-office/internal/models"
	isolationService "github.com/linux-do/cdk-office/internal/services/isolation"
	"github.com/linux-do/cdk-office/internal/services/isolation/cache"
)

// InitIsolationModule 初始化数据隔离模块
func InitIsolationModule(db *gorm.DB, router *gin.Engine, oauth middleware.OAuthMiddleware) error {
	// 数据库迁移
	if err := migrateDatabase(db); err != nil {
		return err
	}

	// 初始化缓存
	cache := cache.NewMemoryCache() // 可以根据配置选择Redis或内存缓存

	// 创建数据隔离服务
	isolationSvc := isolationService.NewDataIsolationService(db, cache)

	// 创建路由
	isolationRouter := isolation.NewIsolationRouter(db, isolationSvc)

	// 注册路由
	apiGroup := router.Group("/api/v1")
	isolationRouter.RegisterRoutes(apiGroup, oauth)

	// 初始化默认策略
	if err := isolationRouter.InitializeDefaultPolicies(); err != nil {
		log.Printf("Warning: Failed to initialize default policies: %v", err)
	}

	log.Println("Data isolation module initialized successfully")
	return nil
}

// migrateDatabase 执行数据库迁移
func migrateDatabase(db *gorm.DB) error {
	log.Println("Migrating data isolation tables...")

	// 自动迁移数据隔离相关表
	if err := db.AutoMigrate(
		&models.TeamDataIsolationPolicy{},
		&models.DataAccessLog{},
		&models.SystemVisibilityConfig{},
		&models.CrossTeamAccessRequest{},
		&models.DataIsolationViolation{},
		&models.UserDataAccessProfile{},
	); err != nil {
		return err
	}

	log.Println("Data isolation tables migrated successfully")
	return nil
}

// IsolationConfig 数据隔离配置
type IsolationConfig struct {
	EnableStrictMode    bool   `json:"enable_strict_mode"`
	DefaultCacheTimeout int    `json:"default_cache_timeout"` // 分钟
	MaxViolationCount   int    `json:"max_violation_count"`
	ViolationBlockTime  int    `json:"violation_block_time"` // 分钟
	EnableAuditLog      bool   `json:"enable_audit_log"`
	EnableAlert         bool   `json:"enable_alert"`
	RedisAddr           string `json:"redis_addr"`
	RedisPassword       string `json:"redis_password"`
	RedisDB             int    `json:"redis_db"`
}

// DefaultIsolationConfig 默认数据隔离配置
func DefaultIsolationConfig() *IsolationConfig {
	return &IsolationConfig{
		EnableStrictMode:    true,
		DefaultCacheTimeout: 10,
		MaxViolationCount:   5,
		ViolationBlockTime:  30,
		EnableAuditLog:      true,
		EnableAlert:         true,
		RedisAddr:           "localhost:6379",
		RedisPassword:       "",
		RedisDB:             0,
	}
}

// InitWithConfig 使用配置初始化数据隔离模块
func InitWithConfig(db *gorm.DB, router *gin.Engine, oauth middleware.OAuthMiddleware, config *IsolationConfig) error {
	// 数据库迁移
	if err := migrateDatabase(db); err != nil {
		return err
	}

	// 根据配置选择缓存实现
	var cacheImpl isolationService.CacheInterface
	if config.RedisAddr != "" {
		cacheImpl = cache.NewRedisCache(config.RedisAddr, config.RedisPassword, config.RedisDB)
	} else {
		cacheImpl = cache.NewMemoryCache()
	}

	// 创建数据隔离服务
	isolationSvc := isolationService.NewDataIsolationService(db, cacheImpl)

	// 创建路由
	isolationRouter := isolation.NewIsolationRouter(db, isolationSvc)

	// 注册路由
	apiGroup := router.Group("/api/v1")
	isolationRouter.RegisterRoutes(apiGroup, oauth)

	// 初始化默认策略
	if err := isolationRouter.InitializeDefaultPolicies(); err != nil {
		log.Printf("Warning: Failed to initialize default policies: %v", err)
	}

	// 初始化系统配置
	if err := initSystemConfig(db, config); err != nil {
		log.Printf("Warning: Failed to initialize system config: %v", err)
	}

	log.Println("Data isolation module initialized successfully with custom config")
	return nil
}

// initSystemConfig 初始化系统配置
func initSystemConfig(db *gorm.DB, config *IsolationConfig) error {
	var sysConfig models.SystemVisibilityConfig
	if err := db.First(&sysConfig).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			// 创建默认系统配置
			sysConfig = models.SystemVisibilityConfig{
				GlobalSettings: struct {
					AllowGlobalSearch    bool `json:"allow_global_search" gorm:"default:false"`
					AllowCrossTeamAccess bool `json:"allow_cross_team_access" gorm:"default:false"`
					RequireApproval      bool `json:"require_approval" gorm:"default:true"`
				}{
					AllowGlobalSearch:    false,
					AllowCrossTeamAccess: false,
					RequireApproval:      true,
				},
				AuditEnabled:       config.EnableAuditLog,
				AlertEnabled:       config.EnableAlert,
				MaxViolationCount:  config.MaxViolationCount,
				ViolationBlockTime: config.ViolationBlockTime,
				CreatedBy:          "system",
			}

			return db.Create(&sysConfig).Error
		}
		return err
	}

	return nil
}
