/*
 * MIT License
 *
 * Copyright (c) 2025 CDK-Office
 */

package optimization

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/go-redis/redis/v8"
)

// RateLimiter 限流器接口
type RateLimiter interface {
	Allow(key string) bool
	AllowN(key string, n int) bool
	Reset(key string) error
	GetStats(key string) *RateLimitStats
}

// RateLimitStats 限流统计
type RateLimitStats struct {
	Key           string    `json:"key"`
	RequestCount  int64     `json:"request_count"`
	AllowedCount  int64     `json:"allowed_count"`
	RejectedCount int64     `json:"rejected_count"`
	ResetTime     time.Time `json:"reset_time"`
	Remaining     int64     `json:"remaining"`
	Limit         int64     `json:"limit"`
}

// RateLimitConfig 限流配置
type RateLimitConfig struct {
	// 基础配置
	Algorithm string        `json:"algorithm"` // 算法类型: token_bucket, sliding_window, fixed_window
	Limit     int64         `json:"limit"`     // 请求限制数量
	Window    time.Duration `json:"window"`    // 时间窗口
	Burst     int64         `json:"burst"`     // 突发流量限制

	// Redis配置
	RedisEnabled bool   `json:"redis_enabled"` // 是否启用Redis
	RedisAddr    string `json:"redis_addr"`    // Redis地址
	RedisDB      int    `json:"redis_db"`      // Redis数据库

	// 分布式配置
	Distributed bool   `json:"distributed"` // 是否分布式限流
	NodeID      string `json:"node_id"`     // 节点ID

	// 高级配置
	WhiteList   []string `json:"white_list"`   // 白名单
	BlackList   []string `json:"black_list"`   // 黑名单
	EnableAlert bool     `json:"enable_alert"` // 启用告警

	// 自适应配置
	Adaptive              bool          `json:"adaptive"`                // 自适应限流
	LoadThreshold         float64       `json:"load_threshold"`          // 负载阈值
	ResponseTimeThreshold time.Duration `json:"response_time_threshold"` // 响应时间阈值
}

// DefaultRateLimitConfig 默认限流配置
func DefaultRateLimitConfig() *RateLimitConfig {
	return &RateLimitConfig{
		Algorithm:             "sliding_window",
		Limit:                 1000,
		Window:                time.Minute,
		Burst:                 100,
		RedisEnabled:          false,
		Distributed:           false,
		EnableAlert:           true,
		Adaptive:              false,
		LoadThreshold:         0.8,
		ResponseTimeThreshold: 500 * time.Millisecond,
	}
}

// TokenBucketLimiter 令牌桶限流器
type TokenBucketLimiter struct {
	config  *RateLimitConfig
	buckets map[string]*TokenBucket
	mutex   sync.RWMutex
}

// TokenBucket 令牌桶
type TokenBucket struct {
	capacity   int64     // 桶容量
	tokens     int64     // 当前令牌数
	refillRate int64     // 填充速率 (令牌/秒)
	lastRefill time.Time // 上次填充时间
	mutex      sync.Mutex
	stats      *RateLimitStats
}

// NewTokenBucketLimiter 创建令牌桶限流器
func NewTokenBucketLimiter(config *RateLimitConfig) *TokenBucketLimiter {
	return &TokenBucketLimiter{
		config:  config,
		buckets: make(map[string]*TokenBucket),
	}
}

// Allow 检查是否允许请求
func (tbl *TokenBucketLimiter) Allow(key string) bool {
	return tbl.AllowN(key, 1)
}

// AllowN 检查是否允许N个请求
func (tbl *TokenBucketLimiter) AllowN(key string, n int) bool {
	tbl.mutex.Lock()
	bucket, exists := tbl.buckets[key]
	if !exists {
		bucket = &TokenBucket{
			capacity:   tbl.config.Burst,
			tokens:     tbl.config.Burst,
			refillRate: tbl.config.Limit / int64(tbl.config.Window.Seconds()),
			lastRefill: time.Now(),
			stats: &RateLimitStats{
				Key:       key,
				Limit:     tbl.config.Limit,
				ResetTime: time.Now().Add(tbl.config.Window),
			},
		}
		tbl.buckets[key] = bucket
	}
	tbl.mutex.Unlock()

	bucket.mutex.Lock()
	defer bucket.mutex.Unlock()

	// 填充令牌
	now := time.Now()
	elapsed := now.Sub(bucket.lastRefill)
	tokensToAdd := int64(elapsed.Seconds()) * bucket.refillRate

	if tokensToAdd > 0 {
		bucket.tokens += tokensToAdd
		if bucket.tokens > bucket.capacity {
			bucket.tokens = bucket.capacity
		}
		bucket.lastRefill = now
	}

	// 检查是否有足够令牌
	bucket.stats.RequestCount++
	if bucket.tokens >= int64(n) {
		bucket.tokens -= int64(n)
		bucket.stats.AllowedCount++
		bucket.stats.Remaining = bucket.tokens
		return true
	}

	bucket.stats.RejectedCount++
	bucket.stats.Remaining = bucket.tokens
	return false
}

// Reset 重置限流器
func (tbl *TokenBucketLimiter) Reset(key string) error {
	tbl.mutex.Lock()
	defer tbl.mutex.Unlock()
	delete(tbl.buckets, key)
	return nil
}

// GetStats 获取统计信息
func (tbl *TokenBucketLimiter) GetStats(key string) *RateLimitStats {
	tbl.mutex.RLock()
	defer tbl.mutex.RUnlock()

	if bucket, exists := tbl.buckets[key]; exists {
		return bucket.stats
	}
	return nil
}

// SlidingWindowLimiter 滑动窗口限流器
type SlidingWindowLimiter struct {
	config  *RateLimitConfig
	windows map[string]*SlidingWindow
	mutex   sync.RWMutex
}

// SlidingWindow 滑动窗口
type SlidingWindow struct {
	requests []time.Time // 请求时间戳列表
	mutex    sync.Mutex
	stats    *RateLimitStats
}

// NewSlidingWindowLimiter 创建滑动窗口限流器
func NewSlidingWindowLimiter(config *RateLimitConfig) *SlidingWindowLimiter {
	return &SlidingWindowLimiter{
		config:  config,
		windows: make(map[string]*SlidingWindow),
	}
}

// Allow 检查是否允许请求
func (swl *SlidingWindowLimiter) Allow(key string) bool {
	return swl.AllowN(key, 1)
}

// AllowN 检查是否允许N个请求
func (swl *SlidingWindowLimiter) AllowN(key string, n int) bool {
	swl.mutex.Lock()
	window, exists := swl.windows[key]
	if !exists {
		window = &SlidingWindow{
			requests: make([]time.Time, 0),
			stats: &RateLimitStats{
				Key:       key,
				Limit:     swl.config.Limit,
				ResetTime: time.Now().Add(swl.config.Window),
			},
		}
		swl.windows[key] = window
	}
	swl.mutex.Unlock()

	window.mutex.Lock()
	defer window.mutex.Unlock()

	now := time.Now()
	windowStart := now.Add(-swl.config.Window)

	// 清理过期请求
	validRequests := make([]time.Time, 0)
	for _, reqTime := range window.requests {
		if reqTime.After(windowStart) {
			validRequests = append(validRequests, reqTime)
		}
	}
	window.requests = validRequests

	// 检查是否超过限制
	window.stats.RequestCount++
	currentCount := int64(len(window.requests))

	if currentCount+int64(n) <= swl.config.Limit {
		// 添加新请求
		for i := 0; i < n; i++ {
			window.requests = append(window.requests, now)
		}
		window.stats.AllowedCount++
		window.stats.Remaining = swl.config.Limit - currentCount - int64(n)
		return true
	}

	window.stats.RejectedCount++
	window.stats.Remaining = swl.config.Limit - currentCount
	return false
}

// Reset 重置限流器
func (swl *SlidingWindowLimiter) Reset(key string) error {
	swl.mutex.Lock()
	defer swl.mutex.Unlock()
	delete(swl.windows, key)
	return nil
}

// GetStats 获取统计信息
func (swl *SlidingWindowLimiter) GetStats(key string) *RateLimitStats {
	swl.mutex.RLock()
	defer swl.mutex.RUnlock()

	if window, exists := swl.windows[key]; exists {
		return window.stats
	}
	return nil
}

// DistributedRateLimiter 分布式限流器
type DistributedRateLimiter struct {
	config       *RateLimitConfig
	redisClient  *redis.Client
	localLimiter RateLimiter
}

// NewDistributedRateLimiter 创建分布式限流器
func NewDistributedRateLimiter(config *RateLimitConfig) *DistributedRateLimiter {
	var redisClient *redis.Client
	if config.RedisEnabled {
		redisClient = redis.NewClient(&redis.Options{
			Addr: config.RedisAddr,
			DB:   config.RedisDB,
		})
	}

	// 创建本地限流器作为备份
	var localLimiter RateLimiter
	switch config.Algorithm {
	case "token_bucket":
		localLimiter = NewTokenBucketLimiter(config)
	case "sliding_window":
		localLimiter = NewSlidingWindowLimiter(config)
	default:
		localLimiter = NewSlidingWindowLimiter(config)
	}

	return &DistributedRateLimiter{
		config:       config,
		redisClient:  redisClient,
		localLimiter: localLimiter,
	}
}

// Allow 检查是否允许请求
func (drl *DistributedRateLimiter) Allow(key string) bool {
	return drl.AllowN(key, 1)
}

// AllowN 检查是否允许N个请求
func (drl *DistributedRateLimiter) AllowN(key string, n int) bool {
	if drl.redisClient == nil {
		return drl.localLimiter.AllowN(key, n)
	}

	ctx := context.Background()
	now := time.Now()
	window := int64(drl.config.Window.Seconds())

	// Redis Lua脚本实现滑动窗口限流
	luaScript := `
		local key = KEYS[1]
		local window = tonumber(ARGV[1])
		local limit = tonumber(ARGV[2])
		local now = tonumber(ARGV[3])
		local count = tonumber(ARGV[4])
		
		-- 清理过期记录
		redis.call('ZREMRANGEBYSCORE', key, '-inf', now - window)
		
		-- 获取当前窗口内的请求数
		local current = redis.call('ZCARD', key)
		
		if current + count <= limit then
			-- 添加新请求
			for i=1,count do
				redis.call('ZADD', key, now, now .. '-' .. i)
			end
			redis.call('EXPIRE', key, window)
			return {1, limit - current - count}
		else
			return {0, limit - current}
		end
	`

	result, err := drl.redisClient.Eval(ctx, luaScript, []string{key},
		window, drl.config.Limit, now.Unix(), n).Result()

	if err != nil {
		// Redis错误时使用本地限流器
		return drl.localLimiter.AllowN(key, n)
	}

	if resultSlice, ok := result.([]interface{}); ok && len(resultSlice) >= 1 {
		if allowed, ok := resultSlice[0].(int64); ok {
			return allowed == 1
		}
	}

	return false
}

// Reset 重置限流器
func (drl *DistributedRateLimiter) Reset(key string) error {
	if drl.redisClient != nil {
		ctx := context.Background()
		return drl.redisClient.Del(ctx, key).Err()
	}
	return drl.localLimiter.Reset(key)
}

// GetStats 获取统计信息
func (drl *DistributedRateLimiter) GetStats(key string) *RateLimitStats {
	// 从Redis或本地获取统计信息
	if drl.redisClient != nil {
		ctx := context.Background()
		count, _ := drl.redisClient.ZCard(ctx, key).Result()
		return &RateLimitStats{
			Key:       key,
			Limit:     drl.config.Limit,
			Remaining: drl.config.Limit - count,
		}
	}
	return drl.localLimiter.GetStats(key)
}

// AdaptiveRateLimiter 自适应限流器
type AdaptiveRateLimiter struct {
	baseLimiter  RateLimiter
	config       *RateLimitConfig
	monitor      *PerformanceMonitor
	currentLimit int64
	mutex        sync.RWMutex
}

// NewAdaptiveRateLimiter 创建自适应限流器
func NewAdaptiveRateLimiter(config *RateLimitConfig, monitor *PerformanceMonitor) *AdaptiveRateLimiter {
	var baseLimiter RateLimiter
	if config.Distributed {
		baseLimiter = NewDistributedRateLimiter(config)
	} else {
		switch config.Algorithm {
		case "token_bucket":
			baseLimiter = NewTokenBucketLimiter(config)
		default:
			baseLimiter = NewSlidingWindowLimiter(config)
		}
	}

	return &AdaptiveRateLimiter{
		baseLimiter:  baseLimiter,
		config:       config,
		monitor:      monitor,
		currentLimit: config.Limit,
	}
}

// Allow 检查是否允许请求
func (arl *AdaptiveRateLimiter) Allow(key string) bool {
	if arl.config.Adaptive {
		arl.adjustLimit()
	}
	return arl.baseLimiter.Allow(key)
}

// AllowN 检查是否允许N个请求
func (arl *AdaptiveRateLimiter) AllowN(key string, n int) bool {
	if arl.config.Adaptive {
		arl.adjustLimit()
	}
	return arl.baseLimiter.AllowN(key, n)
}

// Reset 重置限流器
func (arl *AdaptiveRateLimiter) Reset(key string) error {
	return arl.baseLimiter.Reset(key)
}

// GetStats 获取统计信息
func (arl *AdaptiveRateLimiter) GetStats(key string) *RateLimitStats {
	return arl.baseLimiter.GetStats(key)
}

// adjustLimit 调整限流阈值
func (arl *AdaptiveRateLimiter) adjustLimit() {
	arl.mutex.Lock()
	defer arl.mutex.Unlock()

	metrics := arl.monitor.GetMetrics()

	// 根据系统负载调整限流
	if metrics.CPU > arl.config.LoadThreshold ||
		metrics.Memory > arl.config.LoadThreshold {
		// 系统负载高，降低限流阈值
		newLimit := int64(float64(arl.config.Limit) * 0.7)
		if newLimit < arl.currentLimit {
			arl.currentLimit = newLimit
		}
	} else if metrics.CPU < arl.config.LoadThreshold*0.5 &&
		metrics.Memory < arl.config.LoadThreshold*0.5 {
		// 系统负载低，提高限流阈值
		newLimit := int64(float64(arl.config.Limit) * 1.2)
		if newLimit > arl.currentLimit && newLimit <= arl.config.Limit*2 {
			arl.currentLimit = newLimit
		}
	}

	// 根据响应时间调整
	if metrics.AvgResponseTime > arl.config.ResponseTimeThreshold {
		// 响应时间过长，降低限流阈值
		arl.currentLimit = int64(float64(arl.currentLimit) * 0.9)
	}

	// 更新配置
	arl.config.Limit = arl.currentLimit
}

// ConcurrencyController 并发控制器
type ConcurrencyController struct {
	maxConcurrency int64
	current        int64
	mutex          sync.RWMutex
	semaphore      chan struct{}
}

// NewConcurrencyController 创建并发控制器
func NewConcurrencyController(maxConcurrency int64) *ConcurrencyController {
	return &ConcurrencyController{
		maxConcurrency: maxConcurrency,
		semaphore:      make(chan struct{}, maxConcurrency),
	}
}

// Acquire 获取并发许可
func (cc *ConcurrencyController) Acquire() bool {
	select {
	case cc.semaphore <- struct{}{}:
		cc.mutex.Lock()
		cc.current++
		cc.mutex.Unlock()
		return true
	default:
		return false
	}
}

// AcquireWithTimeout 带超时获取并发许可
func (cc *ConcurrencyController) AcquireWithTimeout(timeout time.Duration) bool {
	select {
	case cc.semaphore <- struct{}{}:
		cc.mutex.Lock()
		cc.current++
		cc.mutex.Unlock()
		return true
	case <-time.After(timeout):
		return false
	}
}

// Release 释放并发许可
func (cc *ConcurrencyController) Release() {
	select {
	case <-cc.semaphore:
		cc.mutex.Lock()
		cc.current--
		cc.mutex.Unlock()
	default:
		// 防止重复释放
	}
}

// GetStats 获取并发统计
func (cc *ConcurrencyController) GetStats() map[string]interface{} {
	cc.mutex.RLock()
	defer cc.mutex.RUnlock()

	return map[string]interface{}{
		"max_concurrency":     cc.maxConcurrency,
		"current_concurrency": cc.current,
		"available":           cc.maxConcurrency - cc.current,
		"usage_rate":          float64(cc.current) / float64(cc.maxConcurrency) * 100,
	}
}

// RateLimitMiddleware 限流中间件
type RateLimitMiddleware struct {
	limiter  RateLimiter
	config   *RateLimitConfig
	keyFunc  func(*gin.Context) string
	onReject func(*gin.Context, *RateLimitStats)
}

// NewRateLimitMiddleware 创建限流中间件
func NewRateLimitMiddleware(limiter RateLimiter, config *RateLimitConfig) *RateLimitMiddleware {
	return &RateLimitMiddleware{
		limiter:  limiter,
		config:   config,
		keyFunc:  DefaultKeyFunc,
		onReject: DefaultRejectHandler,
	}
}

// DefaultKeyFunc 默认键函数
func DefaultKeyFunc(c *gin.Context) string {
	// 使用IP地址作为限流键
	return c.ClientIP()
}

// DefaultRejectHandler 默认拒绝处理器
func DefaultRejectHandler(c *gin.Context, stats *RateLimitStats) {
	c.Header("X-RateLimit-Limit", fmt.Sprintf("%d", stats.Limit))
	c.Header("X-RateLimit-Remaining", fmt.Sprintf("%d", stats.Remaining))
	c.Header("X-RateLimit-Reset", fmt.Sprintf("%d", stats.ResetTime.Unix()))

	c.JSON(429, gin.H{
		"error":       "Rate limit exceeded",
		"message":     "Too many requests",
		"retry_after": stats.ResetTime.Sub(time.Now()).Seconds(),
	})
	c.Abort()
}

// WithKeyFunc 设置键函数
func (rlm *RateLimitMiddleware) WithKeyFunc(keyFunc func(*gin.Context) string) *RateLimitMiddleware {
	rlm.keyFunc = keyFunc
	return rlm
}

// WithRejectHandler 设置拒绝处理器
func (rlm *RateLimitMiddleware) WithRejectHandler(onReject func(*gin.Context, *RateLimitStats)) *RateLimitMiddleware {
	rlm.onReject = onReject
	return rlm
}

// Handler 中间件处理器
func (rlm *RateLimitMiddleware) Handler() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 检查白名单
		clientIP := c.ClientIP()
		for _, ip := range rlm.config.WhiteList {
			if ip == clientIP {
				c.Next()
				return
			}
		}

		// 检查黑名单
		for _, ip := range rlm.config.BlackList {
			if ip == clientIP {
				c.JSON(403, gin.H{"error": "IP blocked"})
				c.Abort()
				return
			}
		}

		// 获取限流键
		key := rlm.keyFunc(c)

		// 检查限流
		if !rlm.limiter.Allow(key) {
			stats := rlm.limiter.GetStats(key)
			if stats == nil {
				stats = &RateLimitStats{Limit: rlm.config.Limit}
			}
			rlm.onReject(c, stats)
			return
		}

		// 添加限流头信息
		stats := rlm.limiter.GetStats(key)
		if stats != nil {
			c.Header("X-RateLimit-Limit", fmt.Sprintf("%d", stats.Limit))
			c.Header("X-RateLimit-Remaining", fmt.Sprintf("%d", stats.Remaining))
		}

		c.Next()
	}
}

// RateLimitManager 限流管理器
type RateLimitManager struct {
	limiters map[string]RateLimiter
	configs  map[string]*RateLimitConfig
	mutex    sync.RWMutex
}

// NewRateLimitManager 创建限流管理器
func NewRateLimitManager() *RateLimitManager {
	return &RateLimitManager{
		limiters: make(map[string]RateLimiter),
		configs:  make(map[string]*RateLimitConfig),
	}
}

// RegisterLimiter 注册限流器
func (rlm *RateLimitManager) RegisterLimiter(name string, limiter RateLimiter, config *RateLimitConfig) {
	rlm.mutex.Lock()
	defer rlm.mutex.Unlock()
	rlm.limiters[name] = limiter
	rlm.configs[name] = config
}

// GetLimiter 获取限流器
func (rlm *RateLimitManager) GetLimiter(name string) (RateLimiter, bool) {
	rlm.mutex.RLock()
	defer rlm.mutex.RUnlock()
	limiter, exists := rlm.limiters[name]
	return limiter, exists
}

// GetAllStats 获取所有限流统计
func (rlm *RateLimitManager) GetAllStats() map[string]interface{} {
	rlm.mutex.RLock()
	defer rlm.mutex.RUnlock()

	stats := make(map[string]interface{})
	for name, config := range rlm.configs {
		stats[name] = map[string]interface{}{
			"algorithm": config.Algorithm,
			"limit":     config.Limit,
			"window":    config.Window.String(),
		}
	}
	return stats
}

// 全局限流管理器
var GlobalRateLimitManager = NewRateLimitManager()

// InitRateLimiting 初始化限流
func InitRateLimiting() {
	// API限流配置
	apiConfig := DefaultRateLimitConfig()
	apiConfig.Limit = 1000
	apiConfig.Window = time.Minute

	apiLimiter := NewSlidingWindowLimiter(apiConfig)
	GlobalRateLimitManager.RegisterLimiter("api", apiLimiter, apiConfig)

	// 上传限流配置
	uploadConfig := &RateLimitConfig{
		Algorithm: "token_bucket",
		Limit:     10,
		Window:    time.Minute,
		Burst:     5,
	}

	uploadLimiter := NewTokenBucketLimiter(uploadConfig)
	GlobalRateLimitManager.RegisterLimiter("upload", uploadLimiter, uploadConfig)

	fmt.Println("Rate limiting initialized successfully")
}
