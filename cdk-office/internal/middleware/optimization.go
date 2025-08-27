/*
 * MIT License
 *
 * Copyright (c) 2025 CDK-Office
 */

package middleware

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/cdk-office/internal/services/optimization"
	"github.com/gin-gonic/gin"
)

// OptimizationMiddleware 系统优化中间件
type OptimizationMiddleware struct {
	circuitBreaker     *optimization.CircuitBreaker
	performanceMonitor *optimization.PerformanceMonitor
	cacheOptimizer     *optimization.CacheOptimizer
	rateLimitManager   *optimization.RateLimitManager
	concurrencyCtrl    *optimization.ConcurrencyController
	config             *OptimizationConfig
}

// OptimizationConfig 优化配置
type OptimizationConfig struct {
	// 熔断器配置
	CircuitBreakerEnabled bool                               `json:"circuit_breaker_enabled"`
	CircuitBreakerConfig  *optimization.CircuitBreakerConfig `json:"circuit_breaker_config"`

	// 性能监控配置
	PerformanceMonitorEnabled bool                                   `json:"performance_monitor_enabled"`
	PerformanceMonitorConfig  *optimization.PerformanceMonitorConfig `json:"performance_monitor_config"`

	// 缓存优化配置
	CacheOptimizerEnabled bool                               `json:"cache_optimizer_enabled"`
	CacheOptimizerConfig  *optimization.CacheOptimizerConfig `json:"cache_optimizer_config"`

	// 限流配置
	RateLimitEnabled bool                                     `json:"rate_limit_enabled"`
	RateLimitConfigs map[string]*optimization.RateLimitConfig `json:"rate_limit_configs"`

	// 并发控制配置
	ConcurrencyControlEnabled bool  `json:"concurrency_control_enabled"`
	MaxConcurrency            int64 `json:"max_concurrency"`

	// 性能阈值配置
	ResponseTimeThreshold time.Duration `json:"response_time_threshold"`
	CPUThreshold          float64       `json:"cpu_threshold"`
	MemoryThreshold       float64       `json:"memory_threshold"`

	// 告警配置
	AlertEnabled    bool             `json:"alert_enabled"`
	AlertInterval   time.Duration    `json:"alert_interval"`
	AlertThresholds *AlertThresholds `json:"alert_thresholds"`
}

// AlertThresholds 告警阈值
type AlertThresholds struct {
	ErrorRate          float64       `json:"error_rate"`           // 错误率阈值
	ResponseTime       time.Duration `json:"response_time"`        // 响应时间阈值
	ThroughputDrop     float64       `json:"throughput_drop"`      // 吞吐量下降阈值
	CircuitBreakerOpen bool          `json:"circuit_breaker_open"` // 熔断器打开告警
}

// DefaultOptimizationConfig 默认优化配置
func DefaultOptimizationConfig() *OptimizationConfig {
	return &OptimizationConfig{
		CircuitBreakerEnabled:     true,
		CircuitBreakerConfig:      optimization.DefaultCircuitBreakerConfig(),
		PerformanceMonitorEnabled: true,
		PerformanceMonitorConfig:  optimization.DefaultPerformanceMonitorConfig(),
		CacheOptimizerEnabled:     true,
		CacheOptimizerConfig:      optimization.DefaultCacheOptimizerConfig(),
		RateLimitEnabled:          true,
		RateLimitConfigs: map[string]*optimization.RateLimitConfig{
			"api":    optimization.DefaultRateLimitConfig(),
			"upload": {Algorithm: "token_bucket", Limit: 10, Window: time.Minute, Burst: 5},
		},
		ConcurrencyControlEnabled: true,
		MaxConcurrency:            1000,
		ResponseTimeThreshold:     500 * time.Millisecond,
		CPUThreshold:              0.8,
		MemoryThreshold:           0.8,
		AlertEnabled:              true,
		AlertInterval:             5 * time.Minute,
		AlertThresholds: &AlertThresholds{
			ErrorRate:          0.05, // 5% 错误率
			ResponseTime:       1 * time.Second,
			ThroughputDrop:     0.5, // 50% 吞吐量下降
			CircuitBreakerOpen: true,
		},
	}
}

// NewOptimizationMiddleware 创建优化中间件
func NewOptimizationMiddleware(config *OptimizationConfig) *OptimizationMiddleware {
	middleware := &OptimizationMiddleware{
		config: config,
	}

	// 初始化熔断器
	if config.CircuitBreakerEnabled {
		middleware.circuitBreaker = optimization.NewCircuitBreaker(config.CircuitBreakerConfig)
	}

	// 初始化性能监控器
	if config.PerformanceMonitorEnabled {
		middleware.performanceMonitor = optimization.NewPerformanceMonitor(config.PerformanceMonitorConfig)
	}

	// 初始化缓存优化器
	if config.CacheOptimizerEnabled {
		middleware.cacheOptimizer = optimization.NewCacheOptimizer(config.CacheOptimizerConfig)
	}

	// 初始化限流管理器
	if config.RateLimitEnabled {
		middleware.rateLimitManager = optimization.NewRateLimitManager()
		for name, rateLimitConfig := range config.RateLimitConfigs {
			var limiter optimization.RateLimiter
			switch rateLimitConfig.Algorithm {
			case "token_bucket":
				limiter = optimization.NewTokenBucketLimiter(rateLimitConfig)
			case "sliding_window":
				limiter = optimization.NewSlidingWindowLimiter(rateLimitConfig)
			default:
				limiter = optimization.NewSlidingWindowLimiter(rateLimitConfig)
			}
			middleware.rateLimitManager.RegisterLimiter(name, limiter, rateLimitConfig)
		}
	}

	// 初始化并发控制器
	if config.ConcurrencyControlEnabled {
		middleware.concurrencyCtrl = optimization.NewConcurrencyController(config.MaxConcurrency)
	}

	return middleware
}

// PerformanceMonitoringMiddleware 性能监控中间件
func (om *OptimizationMiddleware) PerformanceMonitoringMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		if !om.config.PerformanceMonitorEnabled || om.performanceMonitor == nil {
			c.Next()
			return
		}

		startTime := time.Now()

		// 处理请求
		c.Next()

		// 记录性能指标
		duration := time.Since(startTime)
		statusCode := c.Writer.Status()

		// 构建请求信息
		requestInfo := &optimization.RequestInfo{
			Method:     c.Request.Method,
			Path:       c.Request.URL.Path,
			StatusCode: statusCode,
			Duration:   duration,
			Timestamp:  startTime,
			UserAgent:  c.Request.UserAgent(),
			ClientIP:   c.ClientIP(),
		}

		// 记录请求
		om.performanceMonitor.RecordRequest(requestInfo)

		// 添加性能头信息
		c.Header("X-Response-Time", fmt.Sprintf("%.2fms", float64(duration.Nanoseconds())/1e6))
	}
}

// CircuitBreakerMiddleware 熔断器中间件
func (om *OptimizationMiddleware) CircuitBreakerMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		if !om.config.CircuitBreakerEnabled || om.circuitBreaker == nil {
			c.Next()
			return
		}

		serviceName := fmt.Sprintf("%s:%s", c.Request.Method, c.Request.URL.Path)

		// 检查熔断器状态
		if !om.circuitBreaker.Allow(serviceName) {
			c.JSON(503, gin.H{
				"error":   "Service temporarily unavailable",
				"message": "Circuit breaker is open",
				"service": serviceName,
			})
			c.Abort()
			return
		}

		startTime := time.Now()

		// 处理请求
		c.Next()

		// 记录执行结果
		duration := time.Since(startTime)
		success := c.Writer.Status() < 500

		if success {
			om.circuitBreaker.RecordSuccess(serviceName, duration)
		} else {
			om.circuitBreaker.RecordFailure(serviceName, duration)
		}
	}
}

// RateLimitMiddleware 限流中间件
func (om *OptimizationMiddleware) RateLimitMiddleware(limiterName string) gin.HandlerFunc {
	return func(c *gin.Context) {
		if !om.config.RateLimitEnabled || om.rateLimitManager == nil {
			c.Next()
			return
		}

		limiter, exists := om.rateLimitManager.GetLimiter(limiterName)
		if !exists {
			// 使用默认限流器
			limiter, _ = om.rateLimitManager.GetLimiter("api")
		}

		if limiter == nil {
			c.Next()
			return
		}

		// 生成限流键
		key := fmt.Sprintf("%s:%s", c.ClientIP(), c.Request.URL.Path)

		// 检查限流
		if !limiter.Allow(key) {
			stats := limiter.GetStats(key)

			// 设置限流头信息
			if stats != nil {
				c.Header("X-RateLimit-Limit", fmt.Sprintf("%d", stats.Limit))
				c.Header("X-RateLimit-Remaining", fmt.Sprintf("%d", stats.Remaining))
				c.Header("X-RateLimit-Reset", fmt.Sprintf("%d", stats.ResetTime.Unix()))
			}

			c.JSON(429, gin.H{
				"error":       "Rate limit exceeded",
				"message":     "Too many requests",
				"retry_after": 60, // 秒
			})
			c.Abort()
			return
		}

		c.Next()
	}
}

// ConcurrencyControlMiddleware 并发控制中间件
func (om *OptimizationMiddleware) ConcurrencyControlMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		if !om.config.ConcurrencyControlEnabled || om.concurrencyCtrl == nil {
			c.Next()
			return
		}

		// 尝试获取并发许可
		if !om.concurrencyCtrl.AcquireWithTimeout(5 * time.Second) {
			c.JSON(503, gin.H{
				"error":   "Service overloaded",
				"message": "Too many concurrent requests",
			})
			c.Abort()
			return
		}

		// 确保在请求完成后释放许可
		defer om.concurrencyCtrl.Release()

		c.Next()
	}
}

// CacheMiddleware 缓存中间件
func (om *OptimizationMiddleware) CacheMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		if !om.config.CacheOptimizerEnabled || om.cacheOptimizer == nil {
			c.Next()
			return
		}

		// 只对GET请求使用缓存
		if c.Request.Method != "GET" {
			c.Next()
			return
		}

		// 生成缓存键
		cacheKey := fmt.Sprintf("http_cache:%s:%s", c.Request.Method, c.Request.URL.String())

		// 尝试从缓存获取
		if data, err := om.cacheOptimizer.Get(context.Background(), "http", cacheKey); err == nil && data != nil {
			if responseData, ok := data.(map[string]interface{}); ok {
				// 设置缓存头信息
				c.Header("X-Cache", "HIT")
				c.Header("X-Cache-TTL", "300") // 5分钟

				if statusCode, ok := responseData["status_code"].(int); ok {
					c.JSON(statusCode, responseData["body"])
				} else {
					c.JSON(200, responseData["body"])
				}
				c.Abort()
				return
			}
		}

		// 缓存未命中，处理请求
		c.Header("X-Cache", "MISS")

		// 创建响应写入器来捕获响应
		originalWriter := c.Writer
		responseWriter := &responseWriter{
			ResponseWriter: originalWriter,
			body:           make([]byte, 0),
			statusCode:     200,
		}
		c.Writer = responseWriter

		c.Next()

		// 缓存响应（仅对成功的GET请求）
		if c.Request.Method == "GET" && responseWriter.statusCode == 200 {
			responseData := map[string]interface{}{
				"status_code": responseWriter.statusCode,
				"body":        string(responseWriter.body),
			}

			// 异步缓存，避免影响响应时间
			go func() {
				ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
				defer cancel()
				om.cacheOptimizer.Set(ctx, "http", cacheKey, responseData, 5*time.Minute)
			}()
		}
	}
}

// responseWriter 自定义响应写入器
type responseWriter struct {
	gin.ResponseWriter
	body       []byte
	statusCode int
}

func (rw *responseWriter) Write(data []byte) (int, error) {
	rw.body = append(rw.body, data...)
	return rw.ResponseWriter.Write(data)
}

func (rw *responseWriter) WriteHeader(statusCode int) {
	rw.statusCode = statusCode
	rw.ResponseWriter.WriteHeader(statusCode)
}

// HealthCheckMiddleware 健康检查中间件
func (om *OptimizationMiddleware) HealthCheckMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		if c.Request.URL.Path == "/health" {
			healthStatus := om.GetHealthStatus()

			if healthStatus["status"] == "healthy" {
				c.JSON(200, healthStatus)
			} else {
				c.JSON(503, healthStatus)
			}
			c.Abort()
			return
		}
		c.Next()
	}
}

// GetHealthStatus 获取健康状态
func (om *OptimizationMiddleware) GetHealthStatus() map[string]interface{} {
	status := map[string]interface{}{
		"status":    "healthy",
		"timestamp": time.Now().Unix(),
		"services":  make(map[string]interface{}),
	}

	overallHealthy := true

	// 检查熔断器状态
	if om.circuitBreaker != nil {
		circuitStatus := om.circuitBreaker.GetStats()
		status["services"].(map[string]interface{})["circuit_breaker"] = map[string]interface{}{
			"status": "healthy",
			"stats":  circuitStatus,
		}
	}

	// 检查性能监控状态
	if om.performanceMonitor != nil {
		metrics := om.performanceMonitor.GetMetrics()
		healthy := metrics.CPU < om.config.CPUThreshold && metrics.Memory < om.config.MemoryThreshold

		status["services"].(map[string]interface{})["performance_monitor"] = map[string]interface{}{
			"status":  map[bool]string{true: "healthy", false: "unhealthy"}[healthy],
			"metrics": metrics,
		}

		if !healthy {
			overallHealthy = false
		}
	}

	// 检查缓存优化器状态
	if om.cacheOptimizer != nil {
		cacheStats := om.cacheOptimizer.GetStats()
		status["services"].(map[string]interface{})["cache_optimizer"] = map[string]interface{}{
			"status": "healthy",
			"stats":  cacheStats,
		}
	}

	// 检查并发控制状态
	if om.concurrencyCtrl != nil {
		concurrencyStats := om.concurrencyCtrl.GetStats()
		if usageRate, ok := concurrencyStats["usage_rate"].(float64); ok && usageRate > 90 {
			status["services"].(map[string]interface{})["concurrency_control"] = map[string]interface{}{
				"status": "warning",
				"stats":  concurrencyStats,
			}
		} else {
			status["services"].(map[string]interface{})["concurrency_control"] = map[string]interface{}{
				"status": "healthy",
				"stats":  concurrencyStats,
			}
		}
	}

	if !overallHealthy {
		status["status"] = "unhealthy"
	}

	return status
}

// MetricsMiddleware 指标收集中间件
func (om *OptimizationMiddleware) MetricsMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		if c.Request.URL.Path == "/metrics" {
			metrics := om.GetMetrics()
			c.JSON(200, metrics)
			c.Abort()
			return
		}
		c.Next()
	}
}

// GetMetrics 获取指标
func (om *OptimizationMiddleware) GetMetrics() map[string]interface{} {
	metrics := map[string]interface{}{
		"timestamp": time.Now().Unix(),
	}

	// 性能监控指标
	if om.performanceMonitor != nil {
		metrics["performance"] = om.performanceMonitor.GetMetrics()
		metrics["requests"] = om.performanceMonitor.GetRequestStats()
	}

	// 熔断器指标
	if om.circuitBreaker != nil {
		metrics["circuit_breaker"] = om.circuitBreaker.GetStats()
	}

	// 缓存指标
	if om.cacheOptimizer != nil {
		metrics["cache"] = om.cacheOptimizer.GetStats()
	}

	// 限流指标
	if om.rateLimitManager != nil {
		metrics["rate_limit"] = om.rateLimitManager.GetAllStats()
	}

	// 并发控制指标
	if om.concurrencyCtrl != nil {
		metrics["concurrency"] = om.concurrencyCtrl.GetStats()
	}

	return metrics
}

// AlertMiddleware 告警中间件
func (om *OptimizationMiddleware) AlertMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 在请求处理后检查告警条件
		defer func() {
			if om.config.AlertEnabled {
				go om.checkAlerts()
			}
		}()

		c.Next()
	}
}

// checkAlerts 检查告警条件
func (om *OptimizationMiddleware) checkAlerts() {
	if om.performanceMonitor == nil {
		return
	}

	metrics := om.performanceMonitor.GetMetrics()

	// 检查CPU使用率
	if metrics.CPU > om.config.AlertThresholds.ErrorRate {
		om.sendAlert("High CPU Usage", fmt.Sprintf("CPU usage: %.2f%%", metrics.CPU*100))
	}

	// 检查内存使用率
	if metrics.Memory > om.config.MemoryThreshold {
		om.sendAlert("High Memory Usage", fmt.Sprintf("Memory usage: %.2f%%", metrics.Memory*100))
	}

	// 检查平均响应时间
	if metrics.AvgResponseTime > om.config.AlertThresholds.ResponseTime {
		om.sendAlert("High Response Time", fmt.Sprintf("Average response time: %v", metrics.AvgResponseTime))
	}
}

// sendAlert 发送告警
func (om *OptimizationMiddleware) sendAlert(alertType, message string) {
	// 这里可以集成实际的告警系统，如邮件、短信、钉钉等
	log.Printf("ALERT [%s]: %s", alertType, message)

	// 可以添加更复杂的告警逻辑，如告警聚合、去重等
}

// SetupOptimizationMiddleware 设置优化中间件
func SetupOptimizationMiddleware(router *gin.Engine, config *OptimizationConfig) *OptimizationMiddleware {
	middleware := NewOptimizationMiddleware(config)

	// 全局中间件（按顺序添加很重要）
	router.Use(middleware.HealthCheckMiddleware())           // 健康检查
	router.Use(middleware.MetricsMiddleware())               // 指标收集
	router.Use(middleware.PerformanceMonitoringMiddleware()) // 性能监控
	router.Use(middleware.ConcurrencyControlMiddleware())    // 并发控制
	router.Use(middleware.CircuitBreakerMiddleware())        // 熔断器
	router.Use(middleware.AlertMiddleware())                 // 告警

	return middleware
}

// ApplyRateLimiting 应用限流到特定路由组
func (om *OptimizationMiddleware) ApplyRateLimiting(routerGroup *gin.RouterGroup, limiterName string) {
	if om.config.RateLimitEnabled {
		routerGroup.Use(om.RateLimitMiddleware(limiterName))
	}
}

// ApplyCaching 应用缓存到特定路由组
func (om *OptimizationMiddleware) ApplyCaching(routerGroup *gin.RouterGroup) {
	if om.config.CacheOptimizerEnabled {
		routerGroup.Use(om.CacheMiddleware())
	}
}
