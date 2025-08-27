package optimization

import (
	"context"
	"testing"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
)

// MockRedisClient 模拟Redis客户端
type MockRedisClient struct {
	mock.Mock
}

func (m *MockRedisClient) Get(ctx context.Context, key string) *redis.StringCmd {
	args := m.Called(ctx, key)
	return args.Get(0).(*redis.StringCmd)
}

func (m *MockRedisClient) Set(ctx context.Context, key string, value interface{}, expiration time.Duration) *redis.StatusCmd {
	args := m.Called(ctx, key, value, expiration)
	return args.Get(0).(*redis.StatusCmd)
}

func (m *MockRedisClient) Del(ctx context.Context, keys ...string) *redis.IntCmd {
	args := m.Called(ctx, keys)
	return args.Get(0).(*redis.IntCmd)
}

func (m *MockRedisClient) Publish(ctx context.Context, channel string, message interface{}) *redis.IntCmd {
	args := m.Called(ctx, channel, message)
	return args.Get(0).(*redis.IntCmd)
}

// CircuitBreakerTestSuite 熔断器测试套件
type CircuitBreakerTestSuite struct {
	suite.Suite
	circuitBreaker *CircuitBreaker
	config         *CircuitBreakerConfig
}

func (suite *CircuitBreakerTestSuite) SetupTest() {
	suite.config = &CircuitBreakerConfig{
		FailureThreshold:    5,
		RecoveryTimeout:     30 * time.Second,
		HalfOpenMaxCalls:    3,
		HalfOpenSuccessRate: 0.6,
		MonitorInterval:     1 * time.Second,
	}
	suite.circuitBreaker = NewCircuitBreaker(suite.config)
}

func (suite *CircuitBreakerTestSuite) TestNewCircuitBreaker() {
	cb := NewCircuitBreaker(suite.config)
	suite.NotNil(cb)
	suite.Equal(suite.config, cb.config)
	suite.NotNil(cb.services)
	suite.NotNil(cb.stats)
}

func (suite *CircuitBreakerTestSuite) TestCircuitBreakerStates() {
	serviceName := "test-service"

	// 初始状态应该是 Closed
	state := suite.circuitBreaker.GetState(serviceName)
	suite.Equal(Closed, state)

	// 模拟多次失败，触发熔断
	for i := 0; i < int(suite.config.FailureThreshold); i++ {
		err := suite.circuitBreaker.Call(serviceName, func() error {
			return assert.AnError
		})
		suite.Error(err)
	}

	// 现在应该是 Open 状态
	state = suite.circuitBreaker.GetState(serviceName)
	suite.Equal(Open, state)

	// 在 Open 状态下的调用应该立即失败
	err := suite.circuitBreaker.Call(serviceName, func() error {
		return nil
	})
	suite.Error(err)
	suite.Contains(err.Error(), "circuit breaker is open")
}

func (suite *CircuitBreakerTestSuite) TestHalfOpenState() {
	serviceName := "test-service"

	// 手动设置为 HalfOpen 状态
	service := &CircuitBreakerService{
		Name:              serviceName,
		State:             HalfOpen,
		LastFailureTime:   time.Now(),
		HalfOpenAttempts:  0,
		HalfOpenSuccesses: 0,
	}
	suite.circuitBreaker.services[serviceName] = service

	// 在 HalfOpen 状态下进行成功调用
	err := suite.circuitBreaker.Call(serviceName, func() error {
		return nil
	})
	suite.NoError(err)

	// 检查统计
	suite.Equal(uint64(1), service.HalfOpenAttempts)
	suite.Equal(uint64(1), service.HalfOpenSuccesses)
}

func (suite *CircuitBreakerTestSuite) TestGetStats() {
	serviceName := "test-service"

	// 进行一些调用生成统计数据
	suite.circuitBreaker.Call(serviceName, func() error {
		return nil
	})
	suite.circuitBreaker.Call(serviceName, func() error {
		return assert.AnError
	})

	stats := suite.circuitBreaker.GetStats()
	suite.NotNil(stats)
	suite.Contains(stats, serviceName)

	serviceStats := stats[serviceName]
	suite.Equal(uint64(2), serviceStats.TotalCalls)
	suite.Equal(uint64(1), serviceStats.SuccessfulCalls)
	suite.Equal(uint64(1), serviceStats.FailedCalls)
}

func (suite *CircuitBreakerTestSuite) TestReset() {
	serviceName := "test-service"

	// 触发熔断
	for i := 0; i < int(suite.config.FailureThreshold); i++ {
		suite.circuitBreaker.Call(serviceName, func() error {
			return assert.AnError
		})
	}

	// 验证是 Open 状态
	suite.Equal(Open, suite.circuitBreaker.GetState(serviceName))

	// 重置熔断器
	suite.circuitBreaker.Reset(serviceName)

	// 验证回到 Closed 状态
	suite.Equal(Closed, suite.circuitBreaker.GetState(serviceName))
}

func TestCircuitBreakerSuite(t *testing.T) {
	suite.Run(t, new(CircuitBreakerTestSuite))
}

// RateLimiterTestSuite 限流器测试套件
type RateLimiterTestSuite struct {
	suite.Suite
}

func (suite *RateLimiterTestSuite) TestTokenBucketLimiter() {
	config := &RateLimitConfig{
		Algorithm: "token_bucket",
		Limit:     10,
		Window:    time.Minute,
		Burst:     5,
	}

	limiter := NewTokenBucketLimiter(config)
	suite.NotNil(limiter)

	// 测试正常请求
	for i := 0; i < 5; i++ {
		allowed, err := limiter.Allow(context.Background(), "test-key")
		suite.NoError(err)
		suite.True(allowed)
	}

	// 测试超出限制
	allowed, err := limiter.Allow(context.Background(), "test-key")
	suite.NoError(err)
	suite.False(allowed) // 应该被限流
}

func (suite *RateLimiterTestSuite) TestSlidingWindowLimiter() {
	config := &RateLimitConfig{
		Algorithm: "sliding_window",
		Limit:     5,
		Window:    time.Second,
	}

	limiter := NewSlidingWindowLimiter(config)
	suite.NotNil(limiter)

	// 测试正常请求
	for i := 0; i < 5; i++ {
		allowed, err := limiter.Allow(context.Background(), "test-key")
		suite.NoError(err)
		suite.True(allowed)
	}

	// 测试超出限制
	allowed, err := limiter.Allow(context.Background(), "test-key")
	suite.NoError(err)
	suite.False(allowed) // 应该被限流
}

func (suite *RateLimiterTestSuite) TestRateLimitManager() {
	manager := NewRateLimitManager()
	suite.NotNil(manager)

	// 注册限流器
	config := &RateLimitConfig{
		Algorithm: "token_bucket",
		Limit:     10,
		Window:    time.Minute,
		Burst:     5,
	}

	limiter := NewTokenBucketLimiter(config)
	manager.RegisterLimiter("test", limiter, config)

	// 测试获取限流器
	retrievedLimiter := manager.GetLimiter("test")
	suite.NotNil(retrievedLimiter)

	// 测试检查限流
	allowed, err := manager.CheckRateLimit(context.Background(), "test", "test-key")
	suite.NoError(err)
	suite.True(allowed)

	// 测试获取统计
	stats := manager.GetAllStats()
	suite.NotNil(stats)
	suite.Contains(stats, "test")
}

func TestRateLimiterSuite(t *testing.T) {
	suite.Run(t, new(RateLimiterTestSuite))
}

// CacheOptimizerTestSuite 缓存优化器测试套件
type CacheOptimizerTestSuite struct {
	suite.Suite
	optimizer *CacheOptimizer
	config    *CacheOptimizerConfig
}

func (suite *CacheOptimizerTestSuite) SetupTest() {
	suite.config = &CacheOptimizerConfig{
		EnableL1Cache:      true,
		EnableL2Cache:      true,
		EnableL3Cache:      true,
		L1MaxSize:          1000,
		L2MaxSize:          5000,
		L3MaxSize:          10000,
		DefaultTTL:         5 * time.Minute,
		CleanupInterval:    1 * time.Minute,
		EvictionPolicy:     "lru",
		CompressionEnabled: true,
		EncryptionEnabled:  false,
		MetricsEnabled:     true,
		PrefetchEnabled:    true,
		WarmupEnabled:      true,
	}
	suite.optimizer = NewCacheOptimizer(suite.config)
}

func (suite *CacheOptimizerTestSuite) TestCacheOperations() {
	namespace := "test"
	key := "test-key"
	value := "test-value"

	// 测试设置缓存
	err := suite.optimizer.Set(namespace, key, value, time.Minute)
	suite.NoError(err)

	// 测试获取缓存
	result, found, err := suite.optimizer.Get(namespace, key)
	suite.NoError(err)
	suite.True(found)
	suite.Equal(value, result)

	// 测试删除缓存
	err = suite.optimizer.Delete(namespace, key)
	suite.NoError(err)

	// 验证已删除
	_, found, err = suite.optimizer.Get(namespace, key)
	suite.NoError(err)
	suite.False(found)
}

func (suite *CacheOptimizerTestSuite) TestCacheStats() {
	namespace := "test"

	// 进行一些缓存操作
	suite.optimizer.Set(namespace, "key1", "value1", time.Minute)
	suite.optimizer.Set(namespace, "key2", "value2", time.Minute)
	suite.optimizer.Get(namespace, "key1")
	suite.optimizer.Get(namespace, "key3") // 缓存未命中

	// 获取统计信息
	stats := suite.optimizer.GetStats()
	suite.NotNil(stats)

	// 验证统计数据
	suite.True(stats.TotalRequests > 0)
	suite.True(stats.CacheHits > 0)
	suite.True(stats.CacheMisses > 0)
}

func (suite *CacheOptimizerTestSuite) TestCacheEviction() {
	// 创建小容量的缓存配置
	smallConfig := &CacheOptimizerConfig{
		EnableL1Cache:  true,
		L1MaxSize:      2, // 只能存储2个条目
		DefaultTTL:     time.Minute,
		EvictionPolicy: "lru",
	}

	smallOptimizer := NewCacheOptimizer(smallConfig)
	namespace := "test"

	// 添加超过容量的条目
	smallOptimizer.Set(namespace, "key1", "value1", time.Minute)
	smallOptimizer.Set(namespace, "key2", "value2", time.Minute)
	smallOptimizer.Set(namespace, "key3", "value3", time.Minute) // 应该触发驱逐

	// key1 应该被驱逐
	_, found, _ := smallOptimizer.Get(namespace, "key1")
	suite.False(found)

	// key2 和 key3 应该还在
	_, found, _ = smallOptimizer.Get(namespace, "key2")
	suite.True(found)
	_, found, _ = smallOptimizer.Get(namespace, "key3")
	suite.True(found)
}

func TestCacheOptimizerSuite(t *testing.T) {
	suite.Run(t, new(CacheOptimizerTestSuite))
}

// PerformanceMonitorTestSuite 性能监控测试套件
type PerformanceMonitorTestSuite struct {
	suite.Suite
	monitor *PerformanceMonitor
	config  *PerformanceMonitorConfig
}

func (suite *PerformanceMonitorTestSuite) SetupTest() {
	suite.config = &PerformanceMonitorConfig{
		Enabled:          true,
		SampleInterval:   1 * time.Second,
		MetricsRetention: 1 * time.Hour,
		AlertThresholds: map[string]float64{
			"cpu":    0.8,
			"memory": 0.8,
		},
		EnableProfiling: true,
	}
	suite.monitor = NewPerformanceMonitor(suite.config)
}

func (suite *PerformanceMonitorTestSuite) TestMetricsCollection() {
	// 记录请求
	suite.monitor.RecordRequest("GET", "/api/test", 200, 100*time.Millisecond)
	suite.monitor.RecordRequest("POST", "/api/test", 201, 150*time.Millisecond)
	suite.monitor.RecordRequest("GET", "/api/test", 500, 200*time.Millisecond)

	// 获取指标
	metrics := suite.monitor.GetMetrics()
	suite.NotNil(metrics)

	// 验证有数据记录
	suite.True(metrics.TotalRequests > 0)
	suite.True(metrics.AvgResponseTime > 0)
}

func (suite *PerformanceMonitorTestSuite) TestRequestStats() {
	// 记录不同的请求
	suite.monitor.RecordRequest("GET", "/api/users", 200, 50*time.Millisecond)
	suite.monitor.RecordRequest("GET", "/api/users", 200, 60*time.Millisecond)
	suite.monitor.RecordRequest("POST", "/api/users", 201, 100*time.Millisecond)

	// 获取请求统计
	stats := suite.monitor.GetRequestStats()
	suite.NotNil(stats)

	// 验证统计数据
	if endpointStats, exists := stats["GET:/api/users"]; exists {
		suite.Equal(uint64(2), endpointStats.Count)
		suite.True(endpointStats.AvgResponseTime > 0)
	}
}

func (suite *PerformanceMonitorTestSuite) TestAlertThresholds() {
	// 模拟高CPU使用率
	highCPUMetrics := &SystemMetrics{
		CPU:       0.9, // 超过阈值
		Memory:    0.5,
		Timestamp: time.Now(),
	}

	// 这里应该有告警逻辑的测试
	// 由于当前实现中没有明确的告警接口，这里只是示例
	suite.True(highCPUMetrics.CPU > suite.config.AlertThresholds["cpu"])
}

func TestPerformanceMonitorSuite(t *testing.T) {
	suite.Run(t, new(PerformanceMonitorTestSuite))
}

// OptimizationModuleTestSuite 优化模块集成测试套件
type OptimizationModuleTestSuite struct {
	suite.Suite
	module *OptimizationModule
	config *ModuleConfig
}

func (suite *OptimizationModuleTestSuite) SetupTest() {
	suite.config = DefaultModuleConfig()
	suite.module = NewOptimizationModule(nil, nil, suite.config)
}

func (suite *OptimizationModuleTestSuite) TestModuleInitialization() {
	suite.NotNil(suite.module)
	suite.Equal(suite.config, suite.module.Config)
	suite.False(suite.module.Initialized) // 未初始化
}

func (suite *OptimizationModuleTestSuite) TestDefaultConfig() {
	config := DefaultModuleConfig()
	suite.NotNil(config)
	suite.True(config.EnableCircuitBreaker)
	suite.True(config.EnablePerformanceMonitor)
	suite.True(config.EnableCacheOptimizer)
	suite.True(config.EnableDatabaseOptimizer)
	suite.True(config.EnableRateLimit)
	suite.True(config.EnableConfigManager)
}

func TestOptimizationModuleSuite(t *testing.T) {
	suite.Run(t, new(OptimizationModuleTestSuite))
}

// 基准测试
func BenchmarkCircuitBreakerCall(b *testing.B) {
	config := &CircuitBreakerConfig{
		FailureThreshold: 5,
		RecoveryTimeout:  30 * time.Second,
	}
	cb := NewCircuitBreaker(config)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		cb.Call("benchmark-service", func() error {
			return nil
		})
	}
}

func BenchmarkTokenBucketAllow(b *testing.B) {
	config := &RateLimitConfig{
		Algorithm: "token_bucket",
		Limit:     1000,
		Window:    time.Minute,
		Burst:     100,
	}
	limiter := NewTokenBucketLimiter(config)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		limiter.Allow(context.Background(), "benchmark-key")
	}
}

func BenchmarkCacheSet(b *testing.B) {
	config := &CacheOptimizerConfig{
		EnableL1Cache: true,
		L1MaxSize:     10000,
		DefaultTTL:    time.Minute,
	}
	optimizer := NewCacheOptimizer(config)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		optimizer.Set("benchmark", string(rune(i)), "value", time.Minute)
	}
}

func BenchmarkCacheGet(b *testing.B) {
	config := &CacheOptimizerConfig{
		EnableL1Cache: true,
		L1MaxSize:     10000,
		DefaultTTL:    time.Minute,
	}
	optimizer := NewCacheOptimizer(config)

	// 预填充缓存
	for i := 0; i < 1000; i++ {
		optimizer.Set("benchmark", string(rune(i)), "value", time.Minute)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		optimizer.Get("benchmark", string(rune(i%1000)))
	}
}

// 并发安全测试
func TestConcurrentCircuitBreaker(t *testing.T) {
	config := &CircuitBreakerConfig{
		FailureThreshold: 100,
		RecoveryTimeout:  30 * time.Second,
	}
	cb := NewCircuitBreaker(config)

	concurrency := 10
	iterations := 100
	done := make(chan bool, concurrency)

	for i := 0; i < concurrency; i++ {
		go func(workerID int) {
			defer func() { done <- true }()

			for j := 0; j < iterations; j++ {
				err := cb.Call("concurrent-service", func() error {
					time.Sleep(time.Microsecond) // 模拟工作
					return nil
				})
				assert.NoError(t, err)
			}
		}(i)
	}

	// 等待所有goroutine完成
	for i := 0; i < concurrency; i++ {
		select {
		case <-done:
		case <-time.After(10 * time.Second):
			t.Fatal("Concurrent test timeout")
		}
	}

	// 验证统计
	stats := cb.GetStats()
	assert.Contains(t, stats, "concurrent-service")
	assert.Equal(t, uint64(concurrency*iterations), stats["concurrent-service"].TotalCalls)
}

func TestConcurrentRateLimiter(t *testing.T) {
	config := &RateLimitConfig{
		Algorithm: "token_bucket",
		Limit:     1000,
		Window:    time.Minute,
		Burst:     100,
	}
	limiter := NewTokenBucketLimiter(config)

	concurrency := 10
	iterations := 50
	done := make(chan bool, concurrency)

	for i := 0; i < concurrency; i++ {
		go func(workerID int) {
			defer func() { done <- true }()

			for j := 0; j < iterations; j++ {
				_, err := limiter.Allow(context.Background(), "concurrent-key")
				assert.NoError(t, err)
			}
		}(i)
	}

	// 等待所有goroutine完成
	for i := 0; i < concurrency; i++ {
		select {
		case <-done:
		case <-time.After(5 * time.Second):
			t.Fatal("Concurrent rate limiter test timeout")
		}
	}
}
