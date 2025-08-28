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

package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"

	"github.com/linux-do/cdk-office/internal/models"
	"github.com/linux-do/cdk-office/internal/router"
)

func main() {
	// 初始化日志
	logrus.SetFormatter(&logrus.JSONFormatter{})
	logrus.SetLevel(logrus.InfoLevel)
	logger := logrus.New()

	// 初始化数据库连接
	db, err := initDatabase()
	if err != nil {
		logger.WithError(err).Fatal("数据库初始化失败")
	}

	// 自动迁移数据库表
	if err := db.AutoMigrate(&models.ServiceHealthStatus{}); err != nil {
		logger.WithError(err).Fatal("数据库迁移失败")
	}

	// 创建Gin引擎
	r := gin.New()
	r.Use(gin.Logger(), gin.Recovery())

	// 设置CORS中间件
	r.Use(func(c *gin.Context) {
		c.Header("Access-Control-Allow-Origin", "*")
		c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		c.Header("Access-Control-Allow-Headers", "Origin, Content-Type, Authorization")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}

		c.Next()
	})

	// 设置健康检查路由
	router.SetupHealthCheckRoutes(r, db, logger)

	// 启动定期健康检查
	healthChecker := router.SetupPeriodicHealthCheck(db, logger)

	// 手动执行一次初始健康检查
	ctx := context.Background()
	if _, err := healthChecker.CheckAllServices(ctx); err != nil {
		logger.WithError(err).Warn("初始健康检查执行失败")
	}

	// 创建HTTP服务器
	srv := &http.Server{
		Addr:    ":8080",
		Handler: r,
	}

	// 在goroutine中启动服务器
	go func() {
		logger.Info("服务器启动在端口 :8080")
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.WithError(err).Fatal("服务器启动失败")
		}
	}()

	// 等待中断信号以优雅关闭服务器
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	logger.Info("正在关闭服务器...")

	// 给服务器5秒钟来完成当前请求
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		logger.WithError(err).Fatal("服务器强制关闭")
	}

	logger.Info("服务器已退出")
}

// initDatabase 初始化数据库连接
func initDatabase() (*gorm.DB, error) {
	// 从环境变量获取数据库配置
	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		// 默认配置（用于开发环境）
		dsn = "host=localhost user=postgres password=password dbname=cdk_office port=5432 sslmode=disable TimeZone=Asia/Shanghai"
	}

	// 配置GORM日志
	gormLogger := logger.New(
		log.New(os.Stdout, "\r\n", log.LstdFlags),
		logger.Config{
			SlowThreshold:             time.Second,
			LogLevel:                  logger.Silent,
			IgnoreRecordNotFoundError: true,
			Colorful:                  false,
		},
	)

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{
		Logger: gormLogger,
	})
	if err != nil {
		return nil, err
	}

	// 配置连接池
	sqlDB, err := db.DB()
	if err != nil {
		return nil, err
	}

	sqlDB.SetMaxIdleConns(10)
	sqlDB.SetMaxOpenConns(100)
	sqlDB.SetConnMaxLifetime(time.Hour)

	return db, nil
}
