/*
 * MIT License
 *
 * Copyright (c) 2025 CDK-Office
 */

package db

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/linux-do/cdk-office/internal/config"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// DatabaseManager 数据库管理器
type DatabaseManager struct {
	primary *gorm.DB
	sqlDB   *sql.DB
}

// NewDatabaseManager 创建数据库管理器
func NewDatabaseManager() (*DatabaseManager, error) {
	manager := &DatabaseManager{}

	if err := manager.initPrimaryDatabase(); err != nil {
		return nil, fmt.Errorf("failed to initialize database: %v", err)
	}

	return manager, nil
}

// initPrimaryDatabase 初始化主数据库
func (m *DatabaseManager) initPrimaryDatabase() error {
	dsn := m.buildDSN()

	// 配置GORM日志级别
	gormLogger := m.configureLogger(config.Config.Database.LogLevel)

	// 配置GORM
	gormConfig := &gorm.Config{
		Logger:                                   gormLogger,
		DisableForeignKeyConstraintWhenMigrating: true,
		PrepareStmt:                              true,
	}

	// 连接数据库
	db, err := gorm.Open(postgres.Open(dsn), gormConfig)
	if err != nil {
		return fmt.Errorf("failed to connect to database: %v", err)
	}

	// 获取底层sql.DB连接
	sqlDB, err := db.DB()
	if err != nil {
		return fmt.Errorf("failed to get database instance: %v", err)
	}

	// 配置连接池
	m.configureConnectionPool(sqlDB)

	// 测试连接
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := sqlDB.PingContext(ctx); err != nil {
		return fmt.Errorf("failed to ping database: %v", err)
	}

	m.primary = db
	m.sqlDB = sqlDB

	log.Printf("[DB] Successfully connected to %s database", config.Config.Database.Provider)
	return nil
}

// buildDSN 构建数据库连接字符串
func (m *DatabaseManager) buildDSN() string {
	dbConfig := &config.Config.Database

	// 根据数据库提供商构建DSN
	switch strings.ToLower(dbConfig.Provider) {
	case "supabase":
		return m.buildSupabaseDSN(dbConfig)
	case "memfire":
		return m.buildMemFireDSN(dbConfig)
	case "local_postgres", "postgres", "postgresql":
		return m.buildPostgresDSN(dbConfig)
	default:
		return m.buildPostgresDSN(dbConfig)
	}
}

// buildSupabaseDSN 构建Supabase数据库连接字符串
func (m *DatabaseManager) buildSupabaseDSN(dbConfig *config.databaseConfig) string {
	supabaseConfig := dbConfig.Supabase

	// 优先使用Pooler URL
	if supabaseConfig.PoolerURL != "" {
		return supabaseConfig.PoolerURL
	}

	sslMode := dbConfig.SSLMode
	if sslMode == "" {
		sslMode = "require" // Supabase默认要求SSL
	}

	return fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%d sslmode=%s TimeZone=UTC",
		dbConfig.Host,
		dbConfig.Username,
		dbConfig.Password,
		dbConfig.Database,
		dbConfig.Port,
		sslMode,
	)
}

// buildMemFireDSN 构建MemFire Cloud数据库连接字符串
func (m *DatabaseManager) buildMemFireDSN(dbConfig *config.databaseConfig) string {
	memfireConfig := dbConfig.MemFire

	// 优先使用Pooler URL
	if memfireConfig.PoolerURL != "" {
		return memfireConfig.PoolerURL
	}

	// 直连URL
	if memfireConfig.DirectURL != "" {
		return memfireConfig.DirectURL
	}

	sslMode := dbConfig.SSLMode
	if sslMode == "" {
		sslMode = "require" // MemFire Cloud默认要求SSL
	}

	// 使用中国时区配置
	timeZone := "Asia/Shanghai"
	if memfireConfig.Region != "" && memfireConfig.Region != "cn-shanghai" {
		timeZone = "UTC" // 国外区域使用UTC
	}

	return fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%d sslmode=%s TimeZone=%s",
		dbConfig.Host,
		dbConfig.Username,
		dbConfig.Password,
		dbConfig.Database,
		dbConfig.Port,
		sslMode,
		timeZone,
	)
}

// buildPostgresDSN 构建标准PostgreSQL连接字符串
func (m *DatabaseManager) buildPostgresDSN(dbConfig *config.databaseConfig) string {
	sslMode := dbConfig.SSLMode
	if sslMode == "" {
		sslMode = "disable" // 本地开发默认禁用SSL
	}

	return fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%d sslmode=%s TimeZone=Asia/Shanghai",
		dbConfig.Host,
		dbConfig.Username,
		dbConfig.Password,
		dbConfig.Database,
		dbConfig.Port,
		sslMode,
	)
}

// configureConnectionPool 配置连接池
func (m *DatabaseManager) configureConnectionPool(sqlDB *sql.DB) {
	dbConfig := &config.Config.Database

	maxOpenConns := dbConfig.MaxOpenConn
	maxIdleConns := dbConfig.MaxIdleConn
	connMaxLifetime := time.Duration(dbConfig.ConnMaxLifetime) * time.Second
	connMaxIdleTime := time.Duration(dbConfig.ConnMaxIdleTime) * time.Second

	// 针对云数据库的优化
	switch strings.ToLower(dbConfig.Provider) {
	case "supabase":
		// Supabase免费层限制连接数
		if maxOpenConns > 10 {
			maxOpenConns = 10
			log.Println("[DB] Adjusted max open connections for Supabase free tier")
		}
		if maxIdleConns > 5 {
			maxIdleConns = 5
		}
	case "memfire":
		// MemFire Cloud连接数优化
		memfireConfig := dbConfig.MemFire
		if memfireConfig.MaxConnections > 0 {
			// 使用MemFire配置的最大连接数
			if maxOpenConns > memfireConfig.MaxConnections {
				maxOpenConns = memfireConfig.MaxConnections
				log.Printf("[DB] Adjusted max open connections for MemFire Cloud: %d", maxOpenConns)
			}
			// 空闲连接数为最大连接数的一半
			if maxIdleConns > maxOpenConns/2 {
				maxIdleConns = maxOpenConns / 2
			}
		} else {
			// 默认MemFire Cloud连接数限制
			if maxOpenConns > 20 {
				maxOpenConns = 20
				log.Println("[DB] Adjusted max open connections for MemFire Cloud default")
			}
			if maxIdleConns > 10 {
				maxIdleConns = 10
			}
		}

		// MemFire Cloud在中国，适当延长连接生命周期
		if connMaxLifetime < 30*time.Minute {
			connMaxLifetime = 30 * time.Minute
		}
	}

	sqlDB.SetMaxOpenConns(maxOpenConns)
	sqlDB.SetMaxIdleConns(maxIdleConns)
	sqlDB.SetConnMaxLifetime(connMaxLifetime)

	if connMaxIdleTime > 0 {
		sqlDB.SetConnMaxIdleTime(connMaxIdleTime)
	}
}

// configureLogger 配置日志级别
func (m *DatabaseManager) configureLogger(logLevel string) logger.Interface {
	switch strings.ToLower(logLevel) {
	case "silent":
		return logger.Default.LogMode(logger.Silent)
	case "error":
		return logger.Default.LogMode(logger.Error)
	case "warn":
		return logger.Default.LogMode(logger.Warn)
	case "info":
		return logger.Default.LogMode(logger.Info)
	default:
		return logger.Default.LogMode(logger.Info)
	}
}

// GetPrimary 获取主数据库连接
func (m *DatabaseManager) GetPrimary() *gorm.DB {
	return m.primary
}

// GetSQLDB 获取底层SQL数据库连接
func (m *DatabaseManager) GetSQLDB() *sql.DB {
	return m.sqlDB
}

// Close 关闭数据库连接
func (m *DatabaseManager) Close() error {
	if m.sqlDB != nil {
		if err := m.sqlDB.Close(); err != nil {
			return err
		}
	}

	log.Println("[DB] Database connections closed")
	return nil
}

// HealthCheck 数据库健康检查
func (m *DatabaseManager) HealthCheck() error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := m.sqlDB.PingContext(ctx); err != nil {
		return fmt.Errorf("primary database health check failed: %v", err)
	}

	return nil
}

// GetConnectionStats 获取连接统计信息
func (m *DatabaseManager) GetConnectionStats() sql.DBStats {
	return m.sqlDB.Stats()
}
