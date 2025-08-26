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

package db

import (
	"database/sql"
	"fmt"
	"log"
	"time"

	"github.com/linux-do/cdk-office/internal/config"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

var (
	// DB 全局数据库连接实例
	DB *gorm.DB

	// SQLDB 全局SQL数据库连接实例
	SQLDB *sql.DB
)

// Init 初始化数据库连接
func Init() {
	var err error

	// 构建PostgreSQL连接字符串
	dsn := fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%d sslmode=disable TimeZone=Asia/Shanghai",
		config.Config.Database.Host,
		config.Config.Database.Username,
		config.Config.Database.Password,
		config.Config.Database.Database,
		config.Config.Database.Port,
	)

	// 配置GORM日志级别
	gormLogger := logger.Default
	switch config.Config.Database.LogLevel {
	case "silent":
		gormLogger = logger.Default.LogMode(logger.Silent)
	case "error":
		gormLogger = logger.Default.LogMode(logger.Error)
	case "warn":
		gormLogger = logger.Default.LogMode(logger.Warn)
	case "info":
		gormLogger = logger.Default.LogMode(logger.Info)
	}

	// 连接数据库
	DB, err = gorm.Open(postgres.Open(dsn), &gorm.Config{
		Logger: gormLogger,
	})
	if err != nil {
		log.Fatalf("[DB] failed to connect database: %v", err)
	}

	// 获取底层sql.DB连接
	SQLDB, err = DB.DB()
	if err != nil {
		log.Fatalf("[DB] failed to get database instance: %v", err)
	}

	// 配置连接池
	SQLDB.SetMaxIdleConns(config.Config.Database.MaxIdleConn)
	SQLDB.SetMaxOpenConns(config.Config.Database.MaxOpenConn)
	SQLDB.SetConnMaxLifetime(time.Duration(config.Config.Database.ConnMaxLifetime) * time.Second)

	log.Println("[DB] database connection initialized")
}

// GetDB 获取数据库连接实例
func GetDB() *gorm.DB {
	return DB
}

// GetSQLDB 获取SQL数据库连接实例
func GetSQLDB() *sql.DB {
	return SQLDB
}

// Close 关闭数据库连接
func Close() {
	if SQLDB != nil {
		if err := SQLDB.Close(); err != nil {
			log.Printf("[DB] error closing database connection: %v", err)
		}
	}
	log.Println("[DB] database connection closed")
}
