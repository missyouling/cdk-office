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

package optimization

import (
	"context"
	"errors"
	"sync"
	"time"
)

// CircuitBreakerState 熔断器状态
type CircuitBreakerState int

const (
	StateClosed CircuitBreakerState = iota
	StateHalfOpen
	StateOpen
)

// CircuitBreakerConfig 熔断器配置
type CircuitBreakerConfig struct {
	Name             string        `json:"name"`              // 熔断器名称
	MaxRequests      uint32        `json:"max_requests"`      // 半开状态最大请求数
	Interval         time.Duration `json:"interval"`          // 统计时间窗口
	Timeout          time.Duration `json:"timeout"`           // 开启状态持续时间
	FailureThreshold uint32        `json:"failure_threshold"` // 失败阈值
	SuccessThreshold uint32        `json:"success_threshold"` // 成功阈值
	FailureRate      float64       `json:"failure_rate"`      // 失败率阈值
	MinRequestCount  uint32        `json:"min_request_count"` // 最小请求数
}

// DefaultCircuitBreakerConfig 默认熔断器配置
func DefaultCircuitBreakerConfig(name string) *CircuitBreakerConfig {
	return &CircuitBreakerConfig{
		Name:             name,
		MaxRequests:      10,
		Interval:         60 * time.Second,
		Timeout:          30 * time.Second,
		FailureThreshold: 5,
		SuccessThreshold: 3,
		FailureRate:      0.6,
		MinRequestCount:  5,
	}
}

// CircuitBreaker 熔断器
type CircuitBreaker struct {
	config        *CircuitBreakerConfig
	mutex         sync.RWMutex
	state         CircuitBreakerState
	generation    uint64
	expiry        time.Time
	requests      uint32
	totalFailures uint32
	totalRequests uint32
	onStateChange func(name string, from CircuitBreakerState, to CircuitBreakerState)
}

// NewCircuitBreaker 创建熔断器
func NewCircuitBreaker(config *CircuitBreakerConfig) *CircuitBreaker {
	cb := &CircuitBreaker{
		config:     config,
		state:      StateClosed,
		generation: 0,
		expiry:     time.Now().Add(config.Interval),
	}
	return cb
}

// SetStateChangeCallback 设置状态变更回调
func (cb *CircuitBreaker) SetStateChangeCallback(callback func(name string, from CircuitBreakerState, to CircuitBreakerState)) {
	cb.onStateChange = callback
}

// Execute 执行函数并应用熔断逻辑
func (cb *CircuitBreaker) Execute(fn func() (interface{}, error)) (interface{}, error) {
	generation, err := cb.beforeRequest()
	if err != nil {
		return nil, err
	}

	defer func() {
		if e := recover(); e != nil {
			cb.afterRequest(generation, false)
			panic(e)
		}
	}()

	result, err := fn()
	cb.afterRequest(generation, err == nil)
	return result, err
}

// beforeRequest 请求前检查
func (cb *CircuitBreaker) beforeRequest() (uint64, error) {
	cb.mutex.Lock()
	defer cb.mutex.Unlock()

	now := time.Now()
	state, generation := cb.currentState(now)

	if state == StateOpen {
		return generation, errors.New("circuit breaker is open")
	} else if state == StateHalfOpen {
		if cb.requests >= cb.config.MaxRequests {
			return generation, errors.New("too many requests in half-open state")
		}
	}

	cb.requests++
	cb.totalRequests++
	return generation, nil
}

// afterRequest 请求后处理
func (cb *CircuitBreaker) afterRequest(before uint64, success bool) {
	cb.mutex.Lock()
	defer cb.mutex.Unlock()

	now := time.Now()
	state, generation := cb.currentState(now)
	if generation != before {
		return
	}

	if success {
		cb.onSuccess(state, now)
	} else {
		cb.onFailure(state, now)
	}
}

// currentState 获取当前状态
func (cb *CircuitBreaker) currentState(now time.Time) (CircuitBreakerState, uint64) {
	switch cb.state {
	case StateClosed:
		if !cb.expiry.IsZero() && cb.expiry.Before(now) {
			cb.toNewGeneration(now)
		}
	case StateOpen:
		if cb.expiry.Before(now) {
			cb.setState(StateHalfOpen, now)
		}
	}
	return cb.state, cb.generation
}

// onSuccess 成功处理
func (cb *CircuitBreaker) onSuccess(state CircuitBreakerState, now time.Time) {
	switch state {
	case StateClosed:
		// 闭合状态保持
	case StateHalfOpen:
		if cb.requests >= cb.config.SuccessThreshold {
			cb.setState(StateClosed, now)
		}
	}
}

// onFailure 失败处理
func (cb *CircuitBreaker) onFailure(state CircuitBreakerState, now time.Time) {
	cb.totalFailures++
	switch state {
	case StateClosed:
		if cb.readyToTrip() {
			cb.setState(StateOpen, now)
		}
	case StateHalfOpen:
		cb.setState(StateOpen, now)
	}
}

// readyToTrip 是否准备熔断
func (cb *CircuitBreaker) readyToTrip() bool {
	return cb.totalRequests >= cb.config.MinRequestCount &&
		float64(cb.totalFailures)/float64(cb.totalRequests) >= cb.config.FailureRate
}

// setState 设置状态
func (cb *CircuitBreaker) setState(state CircuitBreakerState, now time.Time) {
	if cb.state == state {
		return
	}

	prev := cb.state
	cb.state = state

	cb.toNewGeneration(now)

	if state == StateOpen {
		cb.expiry = now.Add(cb.config.Timeout)
	} else {
		cb.expiry = time.Time{}
	}

	if cb.onStateChange != nil {
		cb.onStateChange(cb.config.Name, prev, state)
	}
}

// toNewGeneration 新的统计周期
func (cb *CircuitBreaker) toNewGeneration(now time.Time) {
	cb.generation++
	cb.requests = 0
	cb.totalFailures = 0
	cb.totalRequests = 0

	var zero time.Time
	switch cb.state {
	case StateClosed:
		if cb.expiry == zero {
			cb.expiry = now.Add(cb.config.Interval)
		}
	case StateOpen:
		cb.expiry = now.Add(cb.config.Timeout)
	default: // StateHalfOpen
		cb.expiry = zero
	}
}

// State 获取当前状态
func (cb *CircuitBreaker) State() CircuitBreakerState {
	cb.mutex.RLock()
	defer cb.mutex.RUnlock()

	state, _ := cb.currentState(time.Now())
	return state
}

// Metrics 获取统计信息
func (cb *CircuitBreaker) Metrics() map[string]interface{} {
	cb.mutex.RLock()
	defer cb.mutex.RUnlock()

	return map[string]interface{}{
		"name":           cb.config.Name,
		"state":          cb.state,
		"generation":     cb.generation,
		"requests":       cb.requests,
		"total_failures": cb.totalFailures,
		"total_requests": cb.totalRequests,
		"failure_rate":   float64(cb.totalFailures) / float64(cb.totalRequests),
	}
}

// ServiceDegrader 服务降级器
type ServiceDegrader struct {
	circuitBreakers  map[string]*CircuitBreaker
	fallbackHandlers map[string]func() (interface{}, error)
	mutex            sync.RWMutex
}

// NewServiceDegrader 创建服务降级器
func NewServiceDegrader() *ServiceDegrader {
	return &ServiceDegrader{
		circuitBreakers:  make(map[string]*CircuitBreaker),
		fallbackHandlers: make(map[string]func() (interface{}, error)),
	}
}

// RegisterService 注册服务
func (sd *ServiceDegrader) RegisterService(name string, config *CircuitBreakerConfig, fallback func() (interface{}, error)) {
	sd.mutex.Lock()
	defer sd.mutex.Unlock()

	if config == nil {
		config = DefaultCircuitBreakerConfig(name)
	}

	cb := NewCircuitBreaker(config)
	cb.SetStateChangeCallback(sd.onStateChange)

	sd.circuitBreakers[name] = cb
	if fallback != nil {
		sd.fallbackHandlers[name] = fallback
	}
}

// Execute 执行服务调用
func (sd *ServiceDegrader) Execute(serviceName string, fn func() (interface{}, error)) (interface{}, error) {
	sd.mutex.RLock()
	cb, exists := sd.circuitBreakers[serviceName]
	fallback, hasFallback := sd.fallbackHandlers[serviceName]
	sd.mutex.RUnlock()

	if !exists {
		return fn()
	}

	result, err := cb.Execute(fn)
	if err != nil && hasFallback {
		// 使用降级处理
		return fallback()
	}

	return result, err
}

// ExecuteWithContext 带上下文执行服务调用
func (sd *ServiceDegrader) ExecuteWithContext(ctx context.Context, serviceName string, fn func(context.Context) (interface{}, error)) (interface{}, error) {
	return sd.Execute(serviceName, func() (interface{}, error) {
		return fn(ctx)
	})
}

// GetServiceState 获取服务状态
func (sd *ServiceDegrader) GetServiceState(serviceName string) CircuitBreakerState {
	sd.mutex.RLock()
	defer sd.mutex.RUnlock()

	if cb, exists := sd.circuitBreakers[serviceName]; exists {
		return cb.State()
	}
	return StateClosed
}

// GetAllServiceMetrics 获取所有服务指标
func (sd *ServiceDegrader) GetAllServiceMetrics() map[string]interface{} {
	sd.mutex.RLock()
	defer sd.mutex.RUnlock()

	metrics := make(map[string]interface{})
	for name, cb := range sd.circuitBreakers {
		metrics[name] = cb.Metrics()
	}
	return metrics
}

// onStateChange 状态变更回调
func (sd *ServiceDegrader) onStateChange(name string, from CircuitBreakerState, to CircuitBreakerState) {
	// 这里可以发送告警通知
	// log.Printf("Service %s state changed from %v to %v", name, from, to)
}

// 全局服务降级器实例
var GlobalServiceDegrader = NewServiceDegrader()

// InitServiceDegradation 初始化服务降级
func InitServiceDegradation() {
	// 注册核心服务的熔断器

	// Dify服务
	GlobalServiceDegrader.RegisterService("dify", &CircuitBreakerConfig{
		Name:             "dify",
		MaxRequests:      5,
		Interval:         30 * time.Second,
		Timeout:          10 * time.Second,
		FailureThreshold: 3,
		SuccessThreshold: 2,
		FailureRate:      0.5,
		MinRequestCount:  3,
	}, func() (interface{}, error) {
		return map[string]string{"message": "Dify service is temporarily unavailable"}, nil
	})

	// 数据库服务
	GlobalServiceDegrader.RegisterService("database", &CircuitBreakerConfig{
		Name:             "database",
		MaxRequests:      20,
		Interval:         60 * time.Second,
		Timeout:          30 * time.Second,
		FailureThreshold: 10,
		SuccessThreshold: 5,
		FailureRate:      0.7,
		MinRequestCount:  10,
	}, func() (interface{}, error) {
		return nil, errors.New("database service degraded")
	})

	// Redis缓存服务
	GlobalServiceDegrader.RegisterService("redis", &CircuitBreakerConfig{
		Name:             "redis",
		MaxRequests:      15,
		Interval:         30 * time.Second,
		Timeout:          15 * time.Second,
		FailureThreshold: 5,
		SuccessThreshold: 3,
		FailureRate:      0.6,
		MinRequestCount:  5,
	}, func() (interface{}, error) {
		return nil, nil // 缓存降级，返回空值
	})

	// 文件存储服务
	GlobalServiceDegrader.RegisterService("storage", &CircuitBreakerConfig{
		Name:             "storage",
		MaxRequests:      10,
		Interval:         45 * time.Second,
		Timeout:          20 * time.Second,
		FailureThreshold: 4,
		SuccessThreshold: 2,
		FailureRate:      0.5,
		MinRequestCount:  4,
	}, func() (interface{}, error) {
		return map[string]string{"message": "File storage temporarily unavailable"}, nil
	})

	// PDF处理服务
	GlobalServiceDegrader.RegisterService("pdf", &CircuitBreakerConfig{
		Name:             "pdf",
		MaxRequests:      8,
		Interval:         60 * time.Second,
		Timeout:          30 * time.Second,
		FailureThreshold: 3,
		SuccessThreshold: 2,
		FailureRate:      0.4,
		MinRequestCount:  3,
	}, func() (interface{}, error) {
		return map[string]string{"message": "PDF processing service degraded"}, nil
	})
}
