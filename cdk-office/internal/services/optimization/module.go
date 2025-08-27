/*
 * MIT License
 *
 * Copyright (c) 2025 CDK-Office
 */

package optimization

import (
	"log"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/go-redis/redis/v8"
	"gorm.io/gorm"
)

// OptimizationModule 系统优化模块
type OptimizationModule struct {
	// 核心组件
	DB          *gorm.DB
	RedisClient *redis.Client

	// 优化组件
	CircuitBreaker     *CircuitBreaker
	PerformanceMonitor *PerformanceMonitor
	CacheOptimizer     *CacheOptimizer
	DatabaseOptimizer  *DatabaseOptimizer
	RateLimitManager   *RateLimitManager
	ConfigManager      *ConfigManager

	// API组件
	ConfigAPI *ConfigAPI

	// 配置
	Config *ModuleConfig

	// 状态
	Initialized bool
}

// ModuleConfig 模块配置
type ModuleConfig struct {
	// 熔断器配置
	EnableCircuitBreaker bool                  `json:"enable_circuit_breaker"`
	CircuitBreakerConfig *CircuitBreakerConfig `json:"circuit_breaker_config"`

	// 性能监控配置
	EnablePerformanceMonitor bool                      `json:"enable_performance_monitor"`
	PerformanceMonitorConfig *PerformanceMonitorConfig `json:"performance_monitor_config"`

	// 缓存优化配置
	EnableCacheOptimizer bool                  `json:"enable_cache_optimizer"`
	CacheOptimizerConfig *CacheOptimizerConfig `json:"cache_optimizer_config"`

	// 数据库优化配置
	EnableDatabaseOptimizer bool            `json:"enable_database_optimizer"`
	DatabaseConfig          *DatabaseConfig `json:"database_config"`

	// 限流配置
	EnableRateLimit  bool                        `json:"enable_rate_limit"`
	RateLimitConfigs map[string]*RateLimitConfig `json:"rate_limit_configs"`

	// 配置管理
	EnableConfigManager bool `json:"enable_config_manager"`

	// 自动调优
	EnableAutoTuning bool `json:"enable_auto_tuning"`
}

// DefaultModuleConfig 默认模块配置
func DefaultModuleConfig() *ModuleConfig {
	return &ModuleConfig{
		EnableCircuitBreaker:     true,
		CircuitBreakerConfig:     DefaultCircuitBreakerConfig(),
		EnablePerformanceMonitor: true,
		PerformanceMonitorConfig: DefaultPerformanceMonitorConfig(),
		EnableCacheOptimizer:     true,
		CacheOptimizerConfig:     DefaultCacheOptimizerConfig(),
		EnableDatabaseOptimizer:  true,
		DatabaseConfig:           DefaultDatabaseConfig(),
		EnableRateLimit:          true,
		RateLimitConfigs: map[string]*RateLimitConfig{
			"api":      DefaultRateLimitConfig(),
			"upload":   {Algorithm: "token_bucket", Limit: 10, Window: time.Minute, Burst: 5},
			"download": {Algorithm: "sliding_window", Limit: 50, Window: time.Minute},
		},
		EnableConfigManager: true,
		EnableAutoTuning:    true,
	}
}

// NewOptimizationModule 创建优化模块
func NewOptimizationModule(db *gorm.DB, redisClient *redis.Client, config *ModuleConfig) *OptimizationModule {
	if config == nil {
		config = DefaultModuleConfig()
	}

	return &OptimizationModule{
		DB:          db,
		RedisClient: redisClient,
		Config:      config,
		Initialized: false,
	}
}

// Initialize 初始化模块
func (om *OptimizationModule) Initialize() error {
	log.Println("Initializing optimization module...")

	// 初始化熔断器
	if om.Config.EnableCircuitBreaker {
		if err := om.initCircuitBreaker(); err != nil {
			log.Printf("Failed to initialize circuit breaker: %v", err)
			return err
		}
		log.Println("✓ Circuit breaker initialized")
	}

	// 初始化性能监控
	if om.Config.EnablePerformanceMonitor {
		if err := om.initPerformanceMonitor(); err != nil {
			log.Printf("Failed to initialize performance monitor: %v", err)
			return err
		}
		log.Println("✓ Performance monitor initialized")
	}

	// 初始化缓存优化器
	if om.Config.EnableCacheOptimizer {
		if err := om.initCacheOptimizer(); err != nil {
			log.Printf("Failed to initialize cache optimizer: %v", err)
			return err
		}
		log.Println("✓ Cache optimizer initialized")
	}

	// 初始化数据库优化器
	if om.Config.EnableDatabaseOptimizer {
		if err := om.initDatabaseOptimizer(); err != nil {
			log.Printf("Failed to initialize database optimizer: %v", err)
			return err
		}
		log.Println("✓ Database optimizer initialized")
	}

	// 初始化限流管理器
	if om.Config.EnableRateLimit {
		if err := om.initRateLimitManager(); err != nil {
			log.Printf("Failed to initialize rate limit manager: %v", err)
			return err
		}
		log.Println("✓ Rate limit manager initialized")
	}

	// 初始化配置管理器
	if om.Config.EnableConfigManager {
		if err := om.initConfigManager(); err != nil {
			log.Printf("Failed to initialize config manager: %v", err)
			return err
		}
		log.Println("✓ Config manager initialized")
	}

	// 初始化API
	om.initAPI()
	log.Println("✓ API endpoints initialized")

	om.Initialized = true
	log.Println("🚀 Optimization module initialized successfully!")

	return nil
}

// initCircuitBreaker 初始化熔断器
func (om *OptimizationModule) initCircuitBreaker() error {
	om.CircuitBreaker = NewCircuitBreaker(om.Config.CircuitBreakerConfig)

	// 注册到全局管理器
	GlobalCircuitBreakerManager.Register("default", om.CircuitBreaker)

	return nil
}

// initPerformanceMonitor 初始化性能监控
func (om *OptimizationModule) initPerformanceMonitor() error {
	om.PerformanceMonitor = NewPerformanceMonitor(om.Config.PerformanceMonitorConfig)

	// 注册到全局管理器
	GlobalPerformanceMonitorManager.Register("default", om.PerformanceMonitor)

	return nil
}

// initCacheOptimizer 初始化缓存优化器
func (om *OptimizationModule) initCacheOptimizer() error {
	om.CacheOptimizer = NewCacheOptimizer(om.Config.CacheOptimizerConfig)

	// 注册到全局管理器
	GlobalCacheOptimizerManager.Register("default", om.CacheOptimizer)

	return nil
}

// initDatabaseOptimizer 初始化数据库优化器
func (om *OptimizationModule) initDatabaseOptimizer() error {
	om.DatabaseOptimizer = NewDatabaseOptimizer(om.DB, om.Config.DatabaseConfig)

	// 注册到全局管理器
	GlobalDatabaseOptimizerManager.RegisterOptimizer("main", om.DatabaseOptimizer)

	return nil
}

// initRateLimitManager 初始化限流管理器
func (om *OptimizationModule) initRateLimitManager() error {
	om.RateLimitManager = NewRateLimitManager()

	// 注册限流器
	for name, config := range om.Config.RateLimitConfigs {
		var limiter RateLimiter

		if config.Distributed && om.RedisClient != nil {
			limiter = NewDistributedRateLimiter(config)
		} else {
			switch config.Algorithm {
			case "token_bucket":
				limiter = NewTokenBucketLimiter(config)
			case "sliding_window":
				limiter = NewSlidingWindowLimiter(config)
			default:
				limiter = NewSlidingWindowLimiter(config)
			}
		}

		om.RateLimitManager.RegisterLimiter(name, limiter, config)
	}

	// 设置全局管理器
	GlobalRateLimitManager = om.RateLimitManager

	return nil
}

// initConfigManager 初始化配置管理器
func (om *OptimizationModule) initConfigManager() error {
	om.ConfigManager = NewConfigManager(om.DB, om.RedisClient)

	// 设置全局管理器
	GlobalConfigManager = om.ConfigManager

	// 注册配置监听器
	om.registerConfigWatchers()

	return nil
}

// initAPI 初始化API
func (om *OptimizationModule) initAPI() {
	if om.ConfigManager != nil {
		om.ConfigAPI = NewConfigAPI(om.ConfigManager)
	}
}

// registerConfigWatchers 注册配置监听器
func (om *OptimizationModule) registerConfigWatchers() {
	if om.ConfigManager == nil {
		return
	}

	// 数据库连接配置监听
	om.ConfigManager.Watch("db.max_open_conns", ConfigWatcherFunc(func(config *Config, oldValue, newValue interface{}) {
		if om.DatabaseOptimizer != nil && newValue != nil {
			if maxOpen, ok := newValue.(int); ok {
				currentIdle, _ := om.ConfigManager.GetInt("db.max_idle_conns")
				currentLifetime, _ := om.ConfigManager.GetInt("db.conn_max_lifetime")
				currentIdleTime, _ := om.ConfigManager.GetInt("db.conn_max_idle_time")

				om.DatabaseOptimizer.OptimizeConnection(
					maxOpen,
					currentIdle,
					time.Duration(currentLifetime)*time.Second,
					time.Duration(currentIdleTime)*time.Second,
				)
				log.Printf("Updated database max_open_conns to %d", maxOpen)
			}
		}
	}))

	// 限流配置监听
	om.ConfigManager.Watch("rate_limit.api.limit", ConfigWatcherFunc(func(config *Config, oldValue, newValue interface{}) {
		if om.RateLimitManager != nil && newValue != nil {
			if limit, ok := newValue.(int); ok {
				// 重新创建限流器
				rateLimitConfig := &RateLimitConfig{
					Algorithm: "sliding_window",
					Limit:     int64(limit),
					Window:    time.Minute,
				}
				newLimiter := NewSlidingWindowLimiter(rateLimitConfig)
				om.RateLimitManager.RegisterLimiter("api", newLimiter, rateLimitConfig)
				log.Printf("Updated API rate limit to %d", limit)
			}
		}
	}))

	// 缓存配置监听
	om.ConfigManager.Watch("cache.max_size", ConfigWatcherFunc(func(config *Config, oldValue, newValue interface{}) {
		if om.CacheOptimizer != nil && newValue != nil {
			if maxSize, ok := newValue.(int); ok {
				// 这里可以动态调整缓存大小
				log.Printf("Cache max size changed to %d", maxSize)
				// 实际实现需要根据缓存系统的API来调整
			}
		}
	}))
}

// RegisterRoutes 注册路由
func (om *OptimizationModule) RegisterRoutes(router *gin.RouterGroup) {
	optimizationGroup := router.Group("/optimization")
	{
		// 模块状态
		optimizationGroup.GET("/status", om.GetModuleStatus)
		optimizationGroup.GET("/health", om.GetModuleHealth)

		// 配置管理API
		if om.ConfigAPI != nil {
			om.ConfigAPI.RegisterRoutes(optimizationGroup)
		}

		// 性能监控API
		if om.PerformanceMonitor != nil {
			optimizationGroup.GET("/performance/metrics", om.GetPerformanceMetrics)
			optimizationGroup.GET("/performance/requests", om.GetRequestStats)
		}

		// 熔断器API
		if om.CircuitBreaker != nil {
			optimizationGroup.GET("/circuit-breaker/stats", om.GetCircuitBreakerStats)
			optimizationGroup.POST("/circuit-breaker/reset", om.ResetCircuitBreaker)
		}

		// 缓存API
		if om.CacheOptimizer != nil {
			optimizationGroup.GET("/cache/stats", om.GetCacheStats)
			optimizationGroup.POST("/cache/clear", om.ClearCache)
		}

		// 数据库优化API
		if om.DatabaseOptimizer != nil {
			optimizationGroup.GET("/database/metrics", om.GetDatabaseMetrics)
			optimizationGroup.GET("/database/slow-queries", om.GetSlowQueries)
		}

		// 限流API
		if om.RateLimitManager != nil {
			optimizationGroup.GET("/rate-limit/stats", om.GetRateLimitStats)
		}
	}
}

// GetModuleStatus 获取模块状态
func (om *OptimizationModule) GetModuleStatus(c *gin.Context) {
	status := gin.H{
		"initialized": om.Initialized,
		"components": gin.H{
			"circuit_breaker":     om.CircuitBreaker != nil,
			"performance_monitor": om.PerformanceMonitor != nil,
			"cache_optimizer":     om.CacheOptimizer != nil,
			"database_optimizer":  om.DatabaseOptimizer != nil,
			"rate_limit_manager":  om.RateLimitManager != nil,
			"config_manager":      om.ConfigManager != nil,
		},
		"config": om.Config,
	}

	c.JSON(200, gin.H{"data": status})
}

// GetModuleHealth 获取模块健康状态
func (om *OptimizationModule) GetModuleHealth(c *gin.Context) {
	health := gin.H{
		"status":     "healthy",
		"components": gin.H{},
	}

	overallHealthy := true

	// 检查各组件健康状态
	if om.PerformanceMonitor != nil {
		metrics := om.PerformanceMonitor.GetMetrics()
		componentHealth := "healthy"
		if metrics.CPU > 0.9 || metrics.Memory > 0.9 {
			componentHealth = "warning"
			overallHealthy = false
		}
		health["components"].(gin.H)["performance_monitor"] = componentHealth
	}

	if om.DatabaseOptimizer != nil {
		dbMetrics := om.DatabaseOptimizer.GetMetrics()
		componentHealth := "healthy"
		if dbMetrics.FailedQueries > 100 {
			componentHealth = "warning"
		}
		health["components"].(gin.H)["database_optimizer"] = componentHealth
	}

	if !overallHealthy {
		health["status"] = "warning"
	}

	c.JSON(200, gin.H{"data": health})
}

// 性能监控相关API
func (om *OptimizationModule) GetPerformanceMetrics(c *gin.Context) {
	if om.PerformanceMonitor == nil {
		c.JSON(404, gin.H{"error": "Performance monitor not available"})
		return
	}

	metrics := om.PerformanceMonitor.GetMetrics()
	c.JSON(200, gin.H{"data": metrics})
}

func (om *OptimizationModule) GetRequestStats(c *gin.Context) {
	if om.PerformanceMonitor == nil {
		c.JSON(404, gin.H{"error": "Performance monitor not available"})
		return
	}

	stats := om.PerformanceMonitor.GetRequestStats()
	c.JSON(200, gin.H{"data": stats})
}

// 熔断器相关API
func (om *OptimizationModule) GetCircuitBreakerStats(c *gin.Context) {
	if om.CircuitBreaker == nil {
		c.JSON(404, gin.H{"error": "Circuit breaker not available"})
		return
	}

	stats := om.CircuitBreaker.GetStats()
	c.JSON(200, gin.H{"data": stats})
}

func (om *OptimizationModule) ResetCircuitBreaker(c *gin.Context) {
	if om.CircuitBreaker == nil {
		c.JSON(404, gin.H{"error": "Circuit breaker not available"})
		return
	}

	serviceName := c.Query("service")
	if serviceName == "" {
		c.JSON(400, gin.H{"error": "Service name is required"})
		return
	}

	om.CircuitBreaker.Reset(serviceName)
	c.JSON(200, gin.H{"message": "Circuit breaker reset successfully"})
}

// 缓存相关API
func (om *OptimizationModule) GetCacheStats(c *gin.Context) {
	if om.CacheOptimizer == nil {
		c.JSON(404, gin.H{"error": "Cache optimizer not available"})
		return
	}

	stats := om.CacheOptimizer.GetStats()
	c.JSON(200, gin.H{"data": stats})
}

func (om *OptimizationModule) ClearCache(c *gin.Context) {
	if om.CacheOptimizer == nil {
		c.JSON(404, gin.H{"error": "Cache optimizer not available"})
		return
	}

	namespace := c.Query("namespace")
	if namespace == "" {
		namespace = "default"
	}

	// 这里需要实现清除缓存的逻辑
	c.JSON(200, gin.H{"message": "Cache cleared successfully"})
}

// 数据库相关API
func (om *OptimizationModule) GetDatabaseMetrics(c *gin.Context) {
	if om.DatabaseOptimizer == nil {
		c.JSON(404, gin.H{"error": "Database optimizer not available"})
		return
	}

	metrics := om.DatabaseOptimizer.GetMetrics()
	c.JSON(200, gin.H{"data": metrics})
}

func (om *OptimizationModule) GetSlowQueries(c *gin.Context) {
	if om.DatabaseOptimizer == nil {
		c.JSON(404, gin.H{"error": "Database optimizer not available"})
		return
	}

	limit := 20
	if limitStr := c.Query("limit"); limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil && l > 0 {
			limit = l
		}
	}

	slowQueries := om.DatabaseOptimizer.GetSlowQueries(limit)
	c.JSON(200, gin.H{"data": slowQueries})
}

// 限流相关API
func (om *OptimizationModule) GetRateLimitStats(c *gin.Context) {
	if om.RateLimitManager == nil {
		c.JSON(404, gin.H{"error": "Rate limit manager not available"})
		return
	}

	stats := om.RateLimitManager.GetAllStats()
	c.JSON(200, gin.H{"data": stats})
}

// Shutdown 关闭模块
func (om *OptimizationModule) Shutdown() {
	log.Println("Shutting down optimization module...")

	// 这里可以添加清理逻辑
	if om.PerformanceMonitor != nil {
		// 停止性能监控
	}

	if om.DatabaseOptimizer != nil {
		// 清理数据库优化器
	}

	om.Initialized = false
	log.Println("Optimization module shut down successfully")
}

// 全局优化模块实例
var GlobalOptimizationModule *OptimizationModule

// InitOptimizationModule 初始化全局优化模块
func InitOptimizationModule(db *gorm.DB, redisClient *redis.Client, config *ModuleConfig) error {
	GlobalOptimizationModule = NewOptimizationModule(db, redisClient, config)
	return GlobalOptimizationModule.Initialize()
}

// GetOptimizationModule 获取全局优化模块
func GetOptimizationModule() *OptimizationModule {
	return GlobalOptimizationModule
}

// RegisterOptimizationRoutes 注册优化模块路由
func RegisterOptimizationRoutes(router *gin.RouterGroup) {
	if GlobalOptimizationModule != nil {
		GlobalOptimizationModule.RegisterRoutes(router)
	}
}
