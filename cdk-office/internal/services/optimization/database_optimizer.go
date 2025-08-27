/*
 * MIT License
 *
 * Copyright (c) 2025 CDK-Office
 */

package optimization

import (
	"context"
	"database/sql"
	"log"
	"sync"
	"time"

	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// DatabaseOptimizer 数据库优化器
type DatabaseOptimizer struct {
	db          *gorm.DB
	monitor     *QueryMonitor
	poolManager *ConnectionPoolManager
	slowLogger  *SlowQueryLogger
	config      *DatabaseConfig
	metrics     *DatabaseMetrics
	mutex       sync.RWMutex
}

// DatabaseConfig 数据库配置
type DatabaseConfig struct {
	// 连接池配置
	MaxOpenConns    int           `json:"max_open_conns"`     // 最大打开连接数
	MaxIdleConns    int           `json:"max_idle_conns"`     // 最大空闲连接数
	ConnMaxLifetime time.Duration `json:"conn_max_lifetime"`  // 连接最大生命周期
	ConnMaxIdleTime time.Duration `json:"conn_max_idle_time"` // 连接最大空闲时间

	// 性能监控配置
	SlowQueryThreshold      time.Duration `json:"slow_query_threshold"`      // 慢查询阈值
	EnableQueryLogging      bool          `json:"enable_query_logging"`      // 启用查询日志
	EnableMetricsCollection bool          `json:"enable_metrics_collection"` // 启用指标收集
	MetricsInterval         time.Duration `json:"metrics_interval"`          // 指标收集间隔

	// 优化配置
	EnablePreparedStatements bool `json:"enable_prepared_statements"` // 启用预处理语句
	EnableQueryCache         bool `json:"enable_query_cache"`         // 启用查询缓存
	CacheSize                int  `json:"cache_size"`                 // 缓存大小
}

// DefaultDatabaseConfig 默认数据库配置
func DefaultDatabaseConfig() *DatabaseConfig {
	return &DatabaseConfig{
		MaxOpenConns:             25,
		MaxIdleConns:             25,
		ConnMaxLifetime:          5 * time.Minute,
		ConnMaxIdleTime:          5 * time.Minute,
		SlowQueryThreshold:       200 * time.Millisecond,
		EnableQueryLogging:       true,
		EnableMetricsCollection:  true,
		MetricsInterval:          30 * time.Second,
		EnablePreparedStatements: true,
		EnableQueryCache:         true,
		CacheSize:                1000,
	}
}

// ConnectionPoolManager 连接池管理器
type ConnectionPoolManager struct {
	db     *sql.DB
	config *DatabaseConfig
	stats  *PoolStats
	mutex  sync.RWMutex
}

// PoolStats 连接池统计
type PoolStats struct {
	OpenConnections   int       `json:"open_connections"`
	InUseConnections  int       `json:"in_use_connections"`
	IdleConnections   int       `json:"idle_connections"`
	WaitCount         int64     `json:"wait_count"`
	WaitDuration      int64     `json:"wait_duration_ms"`
	MaxIdleClosed     int64     `json:"max_idle_closed"`
	MaxLifetimeClosed int64     `json:"max_lifetime_closed"`
	UpdatedAt         time.Time `json:"updated_at"`
}

// NewConnectionPoolManager 创建连接池管理器
func NewConnectionPoolManager(db *sql.DB, config *DatabaseConfig) *ConnectionPoolManager {
	manager := &ConnectionPoolManager{
		db:     db,
		config: config,
		stats:  &PoolStats{},
	}

	// 配置连接池
	manager.configurePool()

	// 启动监控
	go manager.startMonitoring()

	return manager
}

// configurePool 配置连接池
func (cpm *ConnectionPoolManager) configurePool() {
	cpm.db.SetMaxOpenConns(cpm.config.MaxOpenConns)
	cpm.db.SetMaxIdleConns(cpm.config.MaxIdleConns)
	cpm.db.SetConnMaxLifetime(cpm.config.ConnMaxLifetime)
	cpm.db.SetConnMaxIdleTime(cpm.config.ConnMaxIdleTime)
}

// startMonitoring 启动监控
func (cpm *ConnectionPoolManager) startMonitoring() {
	ticker := time.NewTicker(cpm.config.MetricsInterval)
	defer ticker.Stop()

	for range ticker.C {
		cpm.updateStats()
	}
}

// updateStats 更新统计信息
func (cpm *ConnectionPoolManager) updateStats() {
	cpm.mutex.Lock()
	defer cpm.mutex.Unlock()

	stats := cpm.db.Stats()
	cpm.stats = &PoolStats{
		OpenConnections:   stats.OpenConnections,
		InUseConnections:  stats.InUse,
		IdleConnections:   stats.Idle,
		WaitCount:         stats.WaitCount,
		WaitDuration:      stats.WaitDuration.Milliseconds(),
		MaxIdleClosed:     stats.MaxIdleClosed,
		MaxLifetimeClosed: stats.MaxLifetimeClosed,
		UpdatedAt:         time.Now(),
	}
}

// GetStats 获取连接池统计
func (cpm *ConnectionPoolManager) GetStats() *PoolStats {
	cpm.mutex.RLock()
	defer cpm.mutex.RUnlock()
	return cpm.stats
}

// QueryMonitor 查询监控器
type QueryMonitor struct {
	config      *DatabaseConfig
	queryStats  map[string]*QueryStat
	slowQueries []*SlowQuery
	mutex       sync.RWMutex
}

// QueryStat 查询统计
type QueryStat struct {
	SQL           string        `json:"sql"`
	Count         int64         `json:"count"`
	TotalDuration time.Duration `json:"total_duration"`
	AvgDuration   time.Duration `json:"avg_duration"`
	MinDuration   time.Duration `json:"min_duration"`
	MaxDuration   time.Duration `json:"max_duration"`
	LastExecuted  time.Time     `json:"last_executed"`
	ErrorCount    int64         `json:"error_count"`
}

// SlowQuery 慢查询
type SlowQuery struct {
	SQL       string        `json:"sql"`
	Duration  time.Duration `json:"duration"`
	Timestamp time.Time     `json:"timestamp"`
	Error     string        `json:"error,omitempty"`
}

// NewQueryMonitor 创建查询监控器
func NewQueryMonitor(config *DatabaseConfig) *QueryMonitor {
	return &QueryMonitor{
		config:      config,
		queryStats:  make(map[string]*QueryStat),
		slowQueries: make([]*SlowQuery, 0),
	}
}

// RecordQuery 记录查询
func (qm *QueryMonitor) RecordQuery(sql string, duration time.Duration, err error) {
	qm.mutex.Lock()
	defer qm.mutex.Unlock()

	// 更新查询统计
	stat, exists := qm.queryStats[sql]
	if !exists {
		stat = &QueryStat{
			SQL:           sql,
			Count:         0,
			TotalDuration: 0,
			MinDuration:   duration,
			MaxDuration:   duration,
		}
		qm.queryStats[sql] = stat
	}

	stat.Count++
	stat.TotalDuration += duration
	stat.AvgDuration = stat.TotalDuration / time.Duration(stat.Count)
	stat.LastExecuted = time.Now()

	if duration < stat.MinDuration {
		stat.MinDuration = duration
	}
	if duration > stat.MaxDuration {
		stat.MaxDuration = duration
	}

	if err != nil {
		stat.ErrorCount++
	}

	// 记录慢查询
	if duration >= qm.config.SlowQueryThreshold {
		slowQuery := &SlowQuery{
			SQL:       sql,
			Duration:  duration,
			Timestamp: time.Now(),
		}
		if err != nil {
			slowQuery.Error = err.Error()
		}
		qm.slowQueries = append(qm.slowQueries, slowQuery)

		// 限制慢查询记录数量
		if len(qm.slowQueries) > 1000 {
			qm.slowQueries = qm.slowQueries[500:]
		}
	}
}

// GetTopSlowQueries 获取最慢的查询
func (qm *QueryMonitor) GetTopSlowQueries(limit int) []*SlowQuery {
	qm.mutex.RLock()
	defer qm.mutex.RUnlock()

	if len(qm.slowQueries) == 0 {
		return []*SlowQuery{}
	}

	// 简单排序获取最慢的查询
	queries := make([]*SlowQuery, len(qm.slowQueries))
	copy(queries, qm.slowQueries)

	// 这里应该实现更复杂的排序，暂时返回最近的慢查询
	if len(queries) > limit {
		return queries[len(queries)-limit:]
	}

	return queries
}

// GetQueryStats 获取查询统计
func (qm *QueryMonitor) GetQueryStats() map[string]*QueryStat {
	qm.mutex.RLock()
	defer qm.mutex.RUnlock()

	stats := make(map[string]*QueryStat)
	for sql, stat := range qm.queryStats {
		stats[sql] = stat
	}
	return stats
}

// SlowQueryLogger 慢查询日志记录器
type SlowQueryLogger struct {
	config *DatabaseConfig
	logger logger.Interface
}

// NewSlowQueryLogger 创建慢查询日志记录器
func NewSlowQueryLogger(config *DatabaseConfig) *SlowQueryLogger {
	loggerConfig := logger.Config{
		SlowThreshold:             config.SlowQueryThreshold,
		LogLevel:                  logger.Warn,
		IgnoreRecordNotFoundError: true,
		Colorful:                  false,
	}

	return &SlowQueryLogger{
		config: config,
		logger: logger.New(log.New(log.Writer(), "\r\n", log.LstdFlags), loggerConfig),
	}
}

// DatabaseMetrics 数据库指标
type DatabaseMetrics struct {
	TotalQueries    int64         `json:"total_queries"`
	SlowQueries     int64         `json:"slow_queries"`
	FailedQueries   int64         `json:"failed_queries"`
	AvgQueryTime    time.Duration `json:"avg_query_time"`
	ConnectionStats *PoolStats    `json:"connection_stats"`
	LastUpdated     time.Time     `json:"last_updated"`
}

// NewDatabaseOptimizer 创建数据库优化器
func NewDatabaseOptimizer(db *gorm.DB, config *DatabaseConfig) *DatabaseOptimizer {
	sqlDB, _ := db.DB()

	optimizer := &DatabaseOptimizer{
		db:          db,
		monitor:     NewQueryMonitor(config),
		poolManager: NewConnectionPoolManager(sqlDB, config),
		slowLogger:  NewSlowQueryLogger(config),
		config:      config,
		metrics:     &DatabaseMetrics{},
	}

	// 启用查询监控
	if config.EnableQueryLogging {
		optimizer.enableQueryLogging()
	}

	// 启动指标收集
	if config.EnableMetricsCollection {
		go optimizer.startMetricsCollection()
	}

	return optimizer
}

// enableQueryLogging 启用查询日志
func (do *DatabaseOptimizer) enableQueryLogging() {
	// 使用自定义logger包装原有的logger
	originalLogger := do.db.Logger

	customLogger := &CustomLogger{
		original: originalLogger,
		monitor:  do.monitor,
	}

	do.db.Logger = customLogger
}

// CustomLogger 自定义日志记录器
type CustomLogger struct {
	original logger.Interface
	monitor  *QueryMonitor
}

// LogMode 日志模式
func (cl *CustomLogger) LogMode(level logger.LogLevel) logger.Interface {
	return &CustomLogger{
		original: cl.original.LogMode(level),
		monitor:  cl.monitor,
	}
}

// Info 信息日志
func (cl *CustomLogger) Info(ctx context.Context, msg string, data ...interface{}) {
	cl.original.Info(ctx, msg, data...)
}

// Warn 警告日志
func (cl *CustomLogger) Warn(ctx context.Context, msg string, data ...interface{}) {
	cl.original.Warn(ctx, msg, data...)
}

// Error 错误日志
func (cl *CustomLogger) Error(ctx context.Context, msg string, data ...interface{}) {
	cl.original.Error(ctx, msg, data...)
}

// Trace 跟踪日志
func (cl *CustomLogger) Trace(ctx context.Context, begin time.Time, fc func() (sql string, rowsAffected int64), err error) {
	duration := time.Since(begin)
	sql, _ := fc()

	// 记录查询到监控器
	cl.monitor.RecordQuery(sql, duration, err)

	// 调用原始logger
	cl.original.Trace(ctx, begin, fc, err)
}

// startMetricsCollection 启动指标收集
func (do *DatabaseOptimizer) startMetricsCollection() {
	ticker := time.NewTicker(do.config.MetricsInterval)
	defer ticker.Stop()

	for range ticker.C {
		do.updateMetrics()
	}
}

// updateMetrics 更新指标
func (do *DatabaseOptimizer) updateMetrics() {
	do.mutex.Lock()
	defer do.mutex.Unlock()

	stats := do.monitor.GetQueryStats()
	slowQueries := do.monitor.GetTopSlowQueries(10)
	poolStats := do.poolManager.GetStats()

	var totalQueries int64
	var totalDuration time.Duration
	var failedQueries int64

	for _, stat := range stats {
		totalQueries += stat.Count
		totalDuration += stat.TotalDuration
		failedQueries += stat.ErrorCount
	}

	avgQueryTime := time.Duration(0)
	if totalQueries > 0 {
		avgQueryTime = totalDuration / time.Duration(totalQueries)
	}

	do.metrics = &DatabaseMetrics{
		TotalQueries:    totalQueries,
		SlowQueries:     int64(len(slowQueries)),
		FailedQueries:   failedQueries,
		AvgQueryTime:    avgQueryTime,
		ConnectionStats: poolStats,
		LastUpdated:     time.Now(),
	}
}

// GetMetrics 获取指标
func (do *DatabaseOptimizer) GetMetrics() *DatabaseMetrics {
	do.mutex.RLock()
	defer do.mutex.RUnlock()
	return do.metrics
}

// GetSlowQueries 获取慢查询
func (do *DatabaseOptimizer) GetSlowQueries(limit int) []*SlowQuery {
	return do.monitor.GetTopSlowQueries(limit)
}

// GetQueryStats 获取查询统计
func (do *DatabaseOptimizer) GetQueryStats() map[string]*QueryStat {
	return do.monitor.GetQueryStats()
}

// OptimizeConnection 优化连接配置
func (do *DatabaseOptimizer) OptimizeConnection(maxOpen, maxIdle int, maxLifetime, maxIdleTime time.Duration) error {
	sqlDB, err := do.db.DB()
	if err != nil {
		return err
	}

	sqlDB.SetMaxOpenConns(maxOpen)
	sqlDB.SetMaxIdleConns(maxIdle)
	sqlDB.SetConnMaxLifetime(maxLifetime)
	sqlDB.SetConnMaxIdleTime(maxIdleTime)

	// 更新配置
	do.config.MaxOpenConns = maxOpen
	do.config.MaxIdleConns = maxIdle
	do.config.ConnMaxLifetime = maxLifetime
	do.config.ConnMaxIdleTime = maxIdleTime

	return nil
}

// AnalyzePerformance 分析性能
func (do *DatabaseOptimizer) AnalyzePerformance() *PerformanceReport {
	metrics := do.GetMetrics()
	queryStats := do.GetQueryStats()
	slowQueries := do.GetSlowQueries(20)

	report := &PerformanceReport{
		Timestamp:       time.Now(),
		Metrics:         metrics,
		QueryStats:      queryStats,
		SlowQueries:     slowQueries,
		Recommendations: do.generateRecommendations(metrics, queryStats),
	}

	return report
}

// PerformanceReport 性能报告
type PerformanceReport struct {
	Timestamp       time.Time             `json:"timestamp"`
	Metrics         *DatabaseMetrics      `json:"metrics"`
	QueryStats      map[string]*QueryStat `json:"query_stats"`
	SlowQueries     []*SlowQuery          `json:"slow_queries"`
	Recommendations []string              `json:"recommendations"`
}

// generateRecommendations 生成建议
func (do *DatabaseOptimizer) generateRecommendations(metrics *DatabaseMetrics, queryStats map[string]*QueryStat) []string {
	recommendations := []string{}

	// 检查连接池使用情况
	if metrics.ConnectionStats != nil {
		poolUsage := float64(metrics.ConnectionStats.InUseConnections) / float64(do.config.MaxOpenConns)
		if poolUsage > 0.8 {
			recommendations = append(recommendations, "连接池使用率过高，建议增加最大连接数")
		}

		if metrics.ConnectionStats.WaitCount > 100 {
			recommendations = append(recommendations, "连接等待次数过多，建议优化连接池配置")
		}
	}

	// 检查慢查询
	if metrics.SlowQueries > 10 {
		recommendations = append(recommendations, "存在大量慢查询，建议优化SQL语句或添加索引")
	}

	// 检查错误率
	if metrics.FailedQueries > 0 {
		errorRate := float64(metrics.FailedQueries) / float64(metrics.TotalQueries)
		if errorRate > 0.01 { // 错误率超过1%
			recommendations = append(recommendations, "数据库查询错误率较高，建议检查SQL语句和数据库状态")
		}
	}

	// 检查平均查询时间
	if metrics.AvgQueryTime > 100*time.Millisecond {
		recommendations = append(recommendations, "平均查询时间较长，建议优化查询语句")
	}

	return recommendations
}

// DatabaseOptimizerManager 数据库优化器管理器
type DatabaseOptimizerManager struct {
	optimizers map[string]*DatabaseOptimizer
	mutex      sync.RWMutex
}

// NewDatabaseOptimizerManager 创建数据库优化器管理器
func NewDatabaseOptimizerManager() *DatabaseOptimizerManager {
	return &DatabaseOptimizerManager{
		optimizers: make(map[string]*DatabaseOptimizer),
	}
}

// RegisterOptimizer 注册优化器
func (dom *DatabaseOptimizerManager) RegisterOptimizer(name string, optimizer *DatabaseOptimizer) {
	dom.mutex.Lock()
	defer dom.mutex.Unlock()
	dom.optimizers[name] = optimizer
}

// GetOptimizer 获取优化器
func (dom *DatabaseOptimizerManager) GetOptimizer(name string) (*DatabaseOptimizer, bool) {
	dom.mutex.RLock()
	defer dom.mutex.RUnlock()
	optimizer, exists := dom.optimizers[name]
	return optimizer, exists
}

// GetAllMetrics 获取所有数据库指标
func (dom *DatabaseOptimizerManager) GetAllMetrics() map[string]*DatabaseMetrics {
	dom.mutex.RLock()
	defer dom.mutex.RUnlock()

	metrics := make(map[string]*DatabaseMetrics)
	for name, optimizer := range dom.optimizers {
		metrics[name] = optimizer.GetMetrics()
	}
	return metrics
}

// GetGlobalReport 获取全局性能报告
func (dom *DatabaseOptimizerManager) GetGlobalReport() *GlobalPerformanceReport {
	dom.mutex.RLock()
	defer dom.mutex.RUnlock()

	report := &GlobalPerformanceReport{
		Timestamp: time.Now(),
		Databases: make(map[string]*PerformanceReport),
	}

	for name, optimizer := range dom.optimizers {
		report.Databases[name] = optimizer.AnalyzePerformance()
	}

	return report
}

// GlobalPerformanceReport 全局性能报告
type GlobalPerformanceReport struct {
	Timestamp time.Time                     `json:"timestamp"`
	Databases map[string]*PerformanceReport `json:"databases"`
}

// 全局数据库优化器管理器
var GlobalDatabaseOptimizerManager = NewDatabaseOptimizerManager()

// InitDatabaseOptimization 初始化数据库优化
func InitDatabaseOptimization(mainDB *gorm.DB) {
	config := DefaultDatabaseConfig()

	// 主数据库优化器
	mainOptimizer := NewDatabaseOptimizer(mainDB, config)
	GlobalDatabaseOptimizerManager.RegisterOptimizer("main", mainOptimizer)

	log.Println("Database optimization initialized successfully")
}
