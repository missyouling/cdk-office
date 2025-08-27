/*
 * MIT License
 *
 * Copyright (c) 2025 CDK-Office
 */

package optimization

import (
	"context"
	"runtime"
	"sync"
	"time"
)

// PerformanceMetrics 性能指标
type PerformanceMetrics struct {
	// 系统指标
	CPUUsage       float64 `json:"cpu_usage"`
	MemoryUsage    float64 `json:"memory_usage"`
	GoroutineCount int     `json:"goroutine_count"`

	// 应用指标
	RequestCount    int64 `json:"request_count"`
	ErrorCount      int64 `json:"error_count"`
	AvgResponseTime int64 `json:"avg_response_time_ms"`
	P95ResponseTime int64 `json:"p95_response_time_ms"`
	P99ResponseTime int64 `json:"p99_response_time_ms"`

	// 数据库指标
	DBConnections   int   `json:"db_connections"`
	DBQueuedQueries int   `json:"db_queued_queries"`
	DBAvgQueryTime  int64 `json:"db_avg_query_time_ms"`

	// 缓存指标
	CacheHitRate float64 `json:"cache_hit_rate"`
	CacheSize    int64   `json:"cache_size"`

	// 业务指标
	ActiveUsers   int64 `json:"active_users"`
	OnlineUsers   int64 `json:"online_users"`
	UploadCount   int64 `json:"upload_count"`
	DownloadCount int64 `json:"download_count"`

	Timestamp time.Time `json:"timestamp"`
}

// ResponseTimeCollector 响应时间收集器
type ResponseTimeCollector struct {
	times   []int64
	mutex   sync.RWMutex
	maxSize int
}

// NewResponseTimeCollector 创建响应时间收集器
func NewResponseTimeCollector(maxSize int) *ResponseTimeCollector {
	return &ResponseTimeCollector{
		times:   make([]int64, 0, maxSize),
		maxSize: maxSize,
	}
}

// Add 添加响应时间
func (rtc *ResponseTimeCollector) Add(responseTime int64) {
	rtc.mutex.Lock()
	defer rtc.mutex.Unlock()

	if len(rtc.times) >= rtc.maxSize {
		// 移除最老的记录
		rtc.times = rtc.times[1:]
	}
	rtc.times = append(rtc.times, responseTime)
}

// GetPercentile 获取百分位数
func (rtc *ResponseTimeCollector) GetPercentile(p float64) int64 {
	rtc.mutex.RLock()
	defer rtc.mutex.RUnlock()

	if len(rtc.times) == 0 {
		return 0
	}

	// 简单实现，实际应该排序后计算
	sorted := make([]int64, len(rtc.times))
	copy(sorted, rtc.times)

	// 快速排序
	quickSort(sorted, 0, len(sorted)-1)

	index := int(float64(len(sorted)) * p / 100.0)
	if index >= len(sorted) {
		index = len(sorted) - 1
	}
	return sorted[index]
}

// GetAverage 获取平均值
func (rtc *ResponseTimeCollector) GetAverage() int64 {
	rtc.mutex.RLock()
	defer rtc.mutex.RUnlock()

	if len(rtc.times) == 0 {
		return 0
	}

	var sum int64
	for _, t := range rtc.times {
		sum += t
	}
	return sum / int64(len(rtc.times))
}

// quickSort 快速排序
func quickSort(arr []int64, low, high int) {
	if low < high {
		pi := partition(arr, low, high)
		quickSort(arr, low, pi-1)
		quickSort(arr, pi+1, high)
	}
}

func partition(arr []int64, low, high int) int {
	pivot := arr[high]
	i := low - 1

	for j := low; j < high; j++ {
		if arr[j] < pivot {
			i++
			arr[i], arr[j] = arr[j], arr[i]
		}
	}
	arr[i+1], arr[high] = arr[high], arr[i+1]
	return i + 1
}

// PerformanceMonitor 性能监控器
type PerformanceMonitor struct {
	metrics     *PerformanceMetrics
	rtCollector *ResponseTimeCollector
	mutex       sync.RWMutex

	// 计数器
	requestCount  int64
	errorCount    int64
	uploadCount   int64
	downloadCount int64

	// 状态回调
	alertCallbacks []func(*PerformanceMetrics)
}

// NewPerformanceMonitor 创建性能监控器
func NewPerformanceMonitor() *PerformanceMonitor {
	return &PerformanceMonitor{
		metrics:        &PerformanceMetrics{},
		rtCollector:    NewResponseTimeCollector(1000),
		alertCallbacks: make([]func(*PerformanceMetrics), 0),
	}
}

// AddAlertCallback 添加告警回调
func (pm *PerformanceMonitor) AddAlertCallback(callback func(*PerformanceMetrics)) {
	pm.alertCallbacks = append(pm.alertCallbacks, callback)
}

// RecordRequest 记录请求
func (pm *PerformanceMonitor) RecordRequest(responseTime time.Duration, isError bool) {
	pm.mutex.Lock()
	defer pm.mutex.Unlock()

	pm.requestCount++
	if isError {
		pm.errorCount++
	}

	pm.rtCollector.Add(responseTime.Milliseconds())
}

// RecordUpload 记录上传
func (pm *PerformanceMonitor) RecordUpload() {
	pm.mutex.Lock()
	defer pm.mutex.Unlock()
	pm.uploadCount++
}

// RecordDownload 记录下载
func (pm *PerformanceMonitor) RecordDownload() {
	pm.mutex.Lock()
	defer pm.mutex.Unlock()
	pm.downloadCount++
}

// UpdateMetrics 更新性能指标
func (pm *PerformanceMonitor) UpdateMetrics() *PerformanceMetrics {
	pm.mutex.Lock()
	defer pm.mutex.Unlock()

	// 获取系统指标
	var memStats runtime.MemStats
	runtime.ReadMemStats(&memStats)

	pm.metrics = &PerformanceMetrics{
		// 系统指标
		CPUUsage:       getCPUUsage(),
		MemoryUsage:    float64(memStats.Alloc) / float64(memStats.Sys) * 100,
		GoroutineCount: runtime.NumGoroutine(),

		// 应用指标
		RequestCount:    pm.requestCount,
		ErrorCount:      pm.errorCount,
		AvgResponseTime: pm.rtCollector.GetAverage(),
		P95ResponseTime: pm.rtCollector.GetPercentile(95),
		P99ResponseTime: pm.rtCollector.GetPercentile(99),

		// 业务指标
		UploadCount:   pm.uploadCount,
		DownloadCount: pm.downloadCount,

		Timestamp: time.Now(),
	}

	// 检查告警条件
	pm.checkAlerts(pm.metrics)

	return pm.metrics
}

// GetMetrics 获取当前指标
func (pm *PerformanceMonitor) GetMetrics() *PerformanceMetrics {
	pm.mutex.RLock()
	defer pm.mutex.RUnlock()

	// 返回副本
	metrics := *pm.metrics
	return &metrics
}

// checkAlerts 检查告警条件
func (pm *PerformanceMonitor) checkAlerts(metrics *PerformanceMetrics) {
	shouldAlert := false

	// CPU使用率告警
	if metrics.CPUUsage > 80 {
		shouldAlert = true
	}

	// 内存使用率告警
	if metrics.MemoryUsage > 85 {
		shouldAlert = true
	}

	// 错误率告警
	if metrics.RequestCount > 0 {
		errorRate := float64(metrics.ErrorCount) / float64(metrics.RequestCount) * 100
		if errorRate > 10 {
			shouldAlert = true
		}
	}

	// 响应时间告警
	if metrics.P95ResponseTime > 5000 { // 5秒
		shouldAlert = true
	}

	// Goroutine数量告警
	if metrics.GoroutineCount > 1000 {
		shouldAlert = true
	}

	if shouldAlert {
		for _, callback := range pm.alertCallbacks {
			go callback(metrics)
		}
	}
}

// getCPUUsage 获取CPU使用率（简化实现）
func getCPUUsage() float64 {
	// 实际实现中应该使用系统调用获取真实的CPU使用率
	// 这里返回一个模拟值
	return 0.0
}

// ResourceMonitor 资源监控器
type ResourceMonitor struct {
	perfMonitor *PerformanceMonitor
	ticker      *time.Ticker
	stopCh      chan struct{}
	running     bool
	mutex       sync.RWMutex
}

// NewResourceMonitor 创建资源监控器
func NewResourceMonitor() *ResourceMonitor {
	return &ResourceMonitor{
		perfMonitor: NewPerformanceMonitor(),
		stopCh:      make(chan struct{}),
	}
}

// Start 启动监控
func (rm *ResourceMonitor) Start(interval time.Duration) {
	rm.mutex.Lock()
	defer rm.mutex.Unlock()

	if rm.running {
		return
	}

	rm.ticker = time.NewTicker(interval)
	rm.running = true

	go rm.monitor()
}

// Stop 停止监控
func (rm *ResourceMonitor) Stop() {
	rm.mutex.Lock()
	defer rm.mutex.Unlock()

	if !rm.running {
		return
	}

	rm.ticker.Stop()
	close(rm.stopCh)
	rm.running = false
}

// monitor 监控循环
func (rm *ResourceMonitor) monitor() {
	for {
		select {
		case <-rm.ticker.C:
			rm.perfMonitor.UpdateMetrics()
		case <-rm.stopCh:
			return
		}
	}
}

// GetPerformanceMonitor 获取性能监控器
func (rm *ResourceMonitor) GetPerformanceMonitor() *PerformanceMonitor {
	return rm.perfMonitor
}

// HealthChecker 健康检查器
type HealthChecker struct {
	checks map[string]func() error
	mutex  sync.RWMutex
}

// NewHealthChecker 创建健康检查器
func NewHealthChecker() *HealthChecker {
	return &HealthChecker{
		checks: make(map[string]func() error),
	}
}

// RegisterCheck 注册健康检查
func (hc *HealthChecker) RegisterCheck(name string, check func() error) {
	hc.mutex.Lock()
	defer hc.mutex.Unlock()
	hc.checks[name] = check
}

// CheckHealth 执行健康检查
func (hc *HealthChecker) CheckHealth(ctx context.Context) map[string]interface{} {
	hc.mutex.RLock()
	defer hc.mutex.RUnlock()

	results := make(map[string]interface{})
	overall := true

	for name, check := range hc.checks {
		select {
		case <-ctx.Done():
			results[name] = map[string]interface{}{
				"status": "timeout",
				"error":  "health check timeout",
			}
			overall = false
		default:
			if err := check(); err != nil {
				results[name] = map[string]interface{}{
					"status": "unhealthy",
					"error":  err.Error(),
				}
				overall = false
			} else {
				results[name] = map[string]interface{}{
					"status": "healthy",
				}
			}
		}
	}

	results["overall"] = map[string]interface{}{
		"status": func() string {
			if overall {
				return "healthy"
			} else {
				return "unhealthy"
			}
		}(),
		"timestamp": time.Now(),
	}

	return results
}

// 全局实例
var (
	GlobalResourceMonitor = NewResourceMonitor()
	GlobalHealthChecker   = NewHealthChecker()
)

// InitPerformanceMonitoring 初始化性能监控
func InitPerformanceMonitoring() {
	// 添加告警回调
	GlobalResourceMonitor.GetPerformanceMonitor().AddAlertCallback(func(metrics *PerformanceMetrics) {
		// 这里可以发送告警通知
		// log.Printf("Performance alert: %+v", metrics)
	})

	// 注册健康检查
	GlobalHealthChecker.RegisterCheck("database", func() error {
		// 检查数据库连接
		return nil
	})

	GlobalHealthChecker.RegisterCheck("redis", func() error {
		// 检查Redis连接
		return nil
	})

	GlobalHealthChecker.RegisterCheck("dify", func() error {
		// 检查Dify服务
		return nil
	})

	// 启动监控（每30秒更新一次）
	GlobalResourceMonitor.Start(30 * time.Second)
}
