/*
 * MIT License
 *
 * Copyright (c) 2025 CDK-Office
 */

package optimization

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/go-redis/redis/v8"
	"gorm.io/gorm"
)

// ConfigManager 配置管理器
type ConfigManager struct {
	db          *gorm.DB
	redisClient *redis.Client
	configs     map[string]*Config
	watchers    map[string][]ConfigWatcher
	mutex       sync.RWMutex
	autoTuning  *AutoTuningEngine
}

// Config 配置项
type Config struct {
	ID           string                 `json:"id" gorm:"primaryKey"`
	Category     string                 `json:"category" gorm:"index"`         // 配置分类
	Name         string                 `json:"name" gorm:"index"`             // 配置名称
	Value        string                 `json:"value"`                         // 配置值(JSON格式)
	DefaultValue string                 `json:"default_value"`                 // 默认值
	Description  string                 `json:"description"`                   // 描述
	Type         string                 `json:"type"`                          // 类型: string, int, float, bool, object
	Constraints  map[string]interface{} `json:"constraints" gorm:"type:jsonb"` // 约束条件
	Metadata     map[string]interface{} `json:"metadata" gorm:"type:jsonb"`    // 元数据
	Version      int64                  `json:"version"`                       // 版本号
	UpdatedAt    time.Time              `json:"updated_at"`
	CreatedAt    time.Time              `json:"created_at"`
}

// ConfigWatcher 配置监听器
type ConfigWatcher interface {
	OnConfigChanged(config *Config, oldValue, newValue interface{})
}

// ConfigWatcherFunc 配置监听器函数
type ConfigWatcherFunc func(config *Config, oldValue, newValue interface{})

func (f ConfigWatcherFunc) OnConfigChanged(config *Config, oldValue, newValue interface{}) {
	f(config, oldValue, newValue)
}

// NewConfigManager 创建配置管理器
func NewConfigManager(db *gorm.DB, redisClient *redis.Client) *ConfigManager {
	cm := &ConfigManager{
		db:          db,
		redisClient: redisClient,
		configs:     make(map[string]*Config),
		watchers:    make(map[string][]ConfigWatcher),
	}

	// 初始化数据库表
	cm.initDatabase()

	// 加载配置
	cm.loadConfigs()

	// 初始化自动调优引擎
	cm.autoTuning = NewAutoTuningEngine(cm)

	// 启动配置同步
	go cm.startConfigSync()

	return cm
}

// initDatabase 初始化数据库
func (cm *ConfigManager) initDatabase() {
	if err := cm.db.AutoMigrate(&Config{}); err != nil {
		log.Printf("Failed to migrate config table: %v", err)
	}

	// 创建默认配置
	cm.createDefaultConfigs()
}

// createDefaultConfigs 创建默认配置
func (cm *ConfigManager) createDefaultConfigs() {
	defaultConfigs := []*Config{
		// 数据库优化配置
		{
			ID:           "db.max_open_conns",
			Category:     "database",
			Name:         "最大打开连接数",
			Value:        "25",
			DefaultValue: "25",
			Description:  "数据库最大打开连接数",
			Type:         "int",
			Constraints: map[string]interface{}{
				"min": 1,
				"max": 100,
			},
		},
		{
			ID:           "db.max_idle_conns",
			Category:     "database",
			Name:         "最大空闲连接数",
			Value:        "25",
			DefaultValue: "25",
			Description:  "数据库最大空闲连接数",
			Type:         "int",
			Constraints: map[string]interface{}{
				"min": 1,
				"max": 50,
			},
		},
		{
			ID:           "db.conn_max_lifetime",
			Category:     "database",
			Name:         "连接最大生命周期",
			Value:        "300", // 5分钟
			DefaultValue: "300",
			Description:  "数据库连接最大生命周期(秒)",
			Type:         "int",
			Constraints: map[string]interface{}{
				"min": 60,
				"max": 3600,
			},
		},
		{
			ID:           "db.slow_query_threshold",
			Category:     "database",
			Name:         "慢查询阈值",
			Value:        "200", // 200ms
			DefaultValue: "200",
			Description:  "慢查询阈值(毫秒)",
			Type:         "int",
			Constraints: map[string]interface{}{
				"min": 10,
				"max": 5000,
			},
		},

		// 限流配置
		{
			ID:           "rate_limit.api.limit",
			Category:     "rate_limit",
			Name:         "API限流阈值",
			Value:        "1000",
			DefaultValue: "1000",
			Description:  "API每分钟请求限制",
			Type:         "int",
			Constraints: map[string]interface{}{
				"min": 100,
				"max": 10000,
			},
		},
		{
			ID:           "rate_limit.upload.limit",
			Category:     "rate_limit",
			Name:         "上传限流阈值",
			Value:        "10",
			DefaultValue: "10",
			Description:  "上传每分钟请求限制",
			Type:         "int",
			Constraints: map[string]interface{}{
				"min": 1,
				"max": 100,
			},
		},

		// 缓存配置
		{
			ID:           "cache.default_ttl",
			Category:     "cache",
			Name:         "默认TTL",
			Value:        "300", // 5分钟
			DefaultValue: "300",
			Description:  "缓存默认过期时间(秒)",
			Type:         "int",
			Constraints: map[string]interface{}{
				"min": 60,
				"max": 3600,
			},
		},
		{
			ID:           "cache.max_size",
			Category:     "cache",
			Name:         "最大缓存大小",
			Value:        "1000",
			DefaultValue: "1000",
			Description:  "最大缓存条目数",
			Type:         "int",
			Constraints: map[string]interface{}{
				"min": 100,
				"max": 10000,
			},
		},

		// 熔断器配置
		{
			ID:           "circuit_breaker.failure_threshold",
			Category:     "circuit_breaker",
			Name:         "失败阈值",
			Value:        "5",
			DefaultValue: "5",
			Description:  "熔断器失败次数阈值",
			Type:         "int",
			Constraints: map[string]interface{}{
				"min": 1,
				"max": 20,
			},
		},
		{
			ID:           "circuit_breaker.timeout",
			Category:     "circuit_breaker",
			Name:         "超时时间",
			Value:        "60", // 60秒
			DefaultValue: "60",
			Description:  "熔断器超时时间(秒)",
			Type:         "int",
			Constraints: map[string]interface{}{
				"min": 10,
				"max": 300,
			},
		},

		// 并发控制配置
		{
			ID:           "concurrency.max_concurrent",
			Category:     "concurrency",
			Name:         "最大并发数",
			Value:        "1000",
			DefaultValue: "1000",
			Description:  "系统最大并发请求数",
			Type:         "int",
			Constraints: map[string]interface{}{
				"min": 100,
				"max": 5000,
			},
		},

		// 性能监控配置
		{
			ID:           "monitor.cpu_threshold",
			Category:     "monitor",
			Name:         "CPU阈值",
			Value:        "0.8",
			DefaultValue: "0.8",
			Description:  "CPU使用率告警阈值",
			Type:         "float",
			Constraints: map[string]interface{}{
				"min": 0.1,
				"max": 0.99,
			},
		},
		{
			ID:           "monitor.memory_threshold",
			Category:     "monitor",
			Name:         "内存阈值",
			Value:        "0.8",
			DefaultValue: "0.8",
			Description:  "内存使用率告警阈值",
			Type:         "float",
			Constraints: map[string]interface{}{
				"min": 0.1,
				"max": 0.99,
			},
		},
		{
			ID:           "monitor.response_time_threshold",
			Category:     "monitor",
			Name:         "响应时间阈值",
			Value:        "500", // 500ms
			DefaultValue: "500",
			Description:  "响应时间告警阈值(毫秒)",
			Type:         "int",
			Constraints: map[string]interface{}{
				"min": 100,
				"max": 5000,
			},
		},
	}

	for _, config := range defaultConfigs {
		var existingConfig Config
		if err := cm.db.Where("id = ?", config.ID).First(&existingConfig).Error; err != nil {
			if err == gorm.ErrRecordNotFound {
				config.CreatedAt = time.Now()
				config.UpdatedAt = time.Now()
				config.Version = 1
				if err := cm.db.Create(config).Error; err != nil {
					log.Printf("Failed to create default config %s: %v", config.ID, err)
				}
			}
		}
	}
}

// loadConfigs 加载配置
func (cm *ConfigManager) loadConfigs() {
	cm.mutex.Lock()
	defer cm.mutex.Unlock()

	var configs []*Config
	if err := cm.db.Find(&configs).Error; err != nil {
		log.Printf("Failed to load configs: %v", err)
		return
	}

	for _, config := range configs {
		cm.configs[config.ID] = config
	}

	log.Printf("Loaded %d configurations", len(configs))
}

// Get 获取配置值
func (cm *ConfigManager) Get(id string) (interface{}, error) {
	cm.mutex.RLock()
	config, exists := cm.configs[id]
	cm.mutex.RUnlock()

	if !exists {
		return nil, fmt.Errorf("config %s not found", id)
	}

	return cm.parseValue(config.Value, config.Type)
}

// GetString 获取字符串配置
func (cm *ConfigManager) GetString(id string) (string, error) {
	value, err := cm.Get(id)
	if err != nil {
		return "", err
	}
	if str, ok := value.(string); ok {
		return str, nil
	}
	return fmt.Sprintf("%v", value), nil
}

// GetInt 获取整数配置
func (cm *ConfigManager) GetInt(id string) (int, error) {
	value, err := cm.Get(id)
	if err != nil {
		return 0, err
	}
	if i, ok := value.(int); ok {
		return i, nil
	}
	return 0, fmt.Errorf("config %s is not an integer", id)
}

// GetFloat 获取浮点数配置
func (cm *ConfigManager) GetFloat(id string) (float64, error) {
	value, err := cm.Get(id)
	if err != nil {
		return 0, err
	}
	if f, ok := value.(float64); ok {
		return f, nil
	}
	return 0, fmt.Errorf("config %s is not a float", id)
}

// GetBool 获取布尔配置
func (cm *ConfigManager) GetBool(id string) (bool, error) {
	value, err := cm.Get(id)
	if err != nil {
		return false, err
	}
	if b, ok := value.(bool); ok {
		return b, nil
	}
	return false, fmt.Errorf("config %s is not a boolean", id)
}

// Set 设置配置值
func (cm *ConfigManager) Set(id string, value interface{}) error {
	cm.mutex.Lock()
	defer cm.mutex.Unlock()

	config, exists := cm.configs[id]
	if !exists {
		return fmt.Errorf("config %s not found", id)
	}

	// 验证值
	if err := cm.validateValue(value, config); err != nil {
		return err
	}

	oldValue, _ := cm.parseValue(config.Value, config.Type)

	// 转换值为字符串
	valueStr, err := cm.valueToString(value, config.Type)
	if err != nil {
		return err
	}

	// 更新数据库
	config.Value = valueStr
	config.Version++
	config.UpdatedAt = time.Now()

	if err := cm.db.Save(config).Error; err != nil {
		return err
	}

	// 更新内存缓存
	cm.configs[id] = config

	// 发布到Redis
	if cm.redisClient != nil {
		cm.publishConfigChange(config)
	}

	// 通知监听器
	cm.notifyWatchers(config, oldValue, value)

	return nil
}

// parseValue 解析配置值
func (cm *ConfigManager) parseValue(valueStr, typ string) (interface{}, error) {
	switch typ {
	case "string":
		return valueStr, nil
	case "int":
		var i int
		if err := json.Unmarshal([]byte(valueStr), &i); err != nil {
			return 0, err
		}
		return i, nil
	case "float":
		var f float64
		if err := json.Unmarshal([]byte(valueStr), &f); err != nil {
			return 0.0, err
		}
		return f, nil
	case "bool":
		var b bool
		if err := json.Unmarshal([]byte(valueStr), &b); err != nil {
			return false, err
		}
		return b, nil
	case "object":
		var obj interface{}
		if err := json.Unmarshal([]byte(valueStr), &obj); err != nil {
			return nil, err
		}
		return obj, nil
	default:
		return valueStr, nil
	}
}

// valueToString 将值转换为字符串
func (cm *ConfigManager) valueToString(value interface{}, typ string) (string, error) {
	switch typ {
	case "string":
		return fmt.Sprintf("%v", value), nil
	default:
		data, err := json.Marshal(value)
		if err != nil {
			return "", err
		}
		return string(data), nil
	}
}

// validateValue 验证配置值
func (cm *ConfigManager) validateValue(value interface{}, config *Config) error {
	if config.Constraints == nil {
		return nil
	}

	switch config.Type {
	case "int":
		if i, ok := value.(int); ok {
			if min, exists := config.Constraints["min"]; exists {
				if minInt, ok := min.(float64); ok && float64(i) < minInt {
					return fmt.Errorf("value %d is less than minimum %v", i, min)
				}
			}
			if max, exists := config.Constraints["max"]; exists {
				if maxInt, ok := max.(float64); ok && float64(i) > maxInt {
					return fmt.Errorf("value %d is greater than maximum %v", i, max)
				}
			}
		}
	case "float":
		if f, ok := value.(float64); ok {
			if min, exists := config.Constraints["min"]; exists {
				if minFloat, ok := min.(float64); ok && f < minFloat {
					return fmt.Errorf("value %f is less than minimum %v", f, min)
				}
			}
			if max, exists := config.Constraints["max"]; exists {
				if maxFloat, ok := max.(float64); ok && f > maxFloat {
					return fmt.Errorf("value %f is greater than maximum %v", f, max)
				}
			}
		}
	}

	return nil
}

// Watch 添加配置监听器
func (cm *ConfigManager) Watch(configID string, watcher ConfigWatcher) {
	cm.mutex.Lock()
	defer cm.mutex.Unlock()

	cm.watchers[configID] = append(cm.watchers[configID], watcher)
}

// notifyWatchers 通知监听器
func (cm *ConfigManager) notifyWatchers(config *Config, oldValue, newValue interface{}) {
	cm.mutex.RLock()
	watchers := cm.watchers[config.ID]
	cm.mutex.RUnlock()

	for _, watcher := range watchers {
		go func(w ConfigWatcher) {
			defer func() {
				if r := recover(); r != nil {
					log.Printf("Config watcher panic: %v", r)
				}
			}()
			w.OnConfigChanged(config, oldValue, newValue)
		}(watcher)
	}
}

// publishConfigChange 发布配置变更到Redis
func (cm *ConfigManager) publishConfigChange(config *Config) {
	ctx := context.Background()
	data, err := json.Marshal(config)
	if err != nil {
		log.Printf("Failed to marshal config: %v", err)
		return
	}

	if err := cm.redisClient.Publish(ctx, "config_changes", data).Err(); err != nil {
		log.Printf("Failed to publish config change: %v", err)
	}
}

// startConfigSync 启动配置同步
func (cm *ConfigManager) startConfigSync() {
	if cm.redisClient == nil {
		return
	}

	ctx := context.Background()
	pubsub := cm.redisClient.Subscribe(ctx, "config_changes")
	defer pubsub.Close()

	ch := pubsub.Channel()
	for msg := range ch {
		var config Config
		if err := json.Unmarshal([]byte(msg.Payload), &config); err != nil {
			log.Printf("Failed to unmarshal config change: %v", err)
			continue
		}

		cm.mutex.Lock()
		oldConfig := cm.configs[config.ID]
		if oldConfig == nil || oldConfig.Version < config.Version {
			cm.configs[config.ID] = &config

			if oldConfig != nil {
				oldValue, _ := cm.parseValue(oldConfig.Value, oldConfig.Type)
				newValue, _ := cm.parseValue(config.Value, config.Type)
				cm.notifyWatchers(&config, oldValue, newValue)
			}
		}
		cm.mutex.Unlock()
	}
}

// GetByCategory 按分类获取配置
func (cm *ConfigManager) GetByCategory(category string) map[string]*Config {
	cm.mutex.RLock()
	defer cm.mutex.RUnlock()

	result := make(map[string]*Config)
	for id, config := range cm.configs {
		if config.Category == category {
			result[id] = config
		}
	}
	return result
}

// ListCategories 列出所有分类
func (cm *ConfigManager) ListCategories() []string {
	cm.mutex.RLock()
	defer cm.mutex.RUnlock()

	categories := make(map[string]bool)
	for _, config := range cm.configs {
		categories[config.Category] = true
	}

	result := make([]string, 0, len(categories))
	for category := range categories {
		result = append(result, category)
	}
	return result
}

// AutoTuningEngine 自动调优引擎
type AutoTuningEngine struct {
	configManager *ConfigManager
	enabled       bool
	rules         []*TuningRule
	metrics       *PerformanceMetrics
	lastTuning    time.Time
	mutex         sync.Mutex
}

// TuningRule 调优规则
type TuningRule struct {
	ID          string
	Name        string
	Condition   func(*PerformanceMetrics) bool
	Action      func(*ConfigManager, *PerformanceMetrics) error
	CoolDown    time.Duration
	LastApplied time.Time
}

// PerformanceMetrics 性能指标
type PerformanceMetrics struct {
	CPU               float64
	Memory            float64
	AvgResponseTime   time.Duration
	RequestRate       float64
	ErrorRate         float64
	DatabaseConnUsage float64
	CacheHitRate      float64
	UpdatedAt         time.Time
}

// NewAutoTuningEngine 创建自动调优引擎
func NewAutoTuningEngine(cm *ConfigManager) *AutoTuningEngine {
	engine := &AutoTuningEngine{
		configManager: cm,
		enabled:       true,
		rules:         make([]*TuningRule, 0),
		lastTuning:    time.Now(),
	}

	// 添加默认调优规则
	engine.addDefaultRules()

	// 启动调优循环
	go engine.start()

	return engine
}

// addDefaultRules 添加默认调优规则
func (ate *AutoTuningEngine) addDefaultRules() {
	// CPU使用率过高时调整
	ate.rules = append(ate.rules, &TuningRule{
		ID:       "high_cpu_adjust",
		Name:     "高CPU使用率调整",
		CoolDown: 5 * time.Minute,
		Condition: func(metrics *PerformanceMetrics) bool {
			return metrics.CPU > 0.85
		},
		Action: func(cm *ConfigManager, metrics *PerformanceMetrics) error {
			// 降低并发数
			currentConcurrency, _ := cm.GetInt("concurrency.max_concurrent")
			newConcurrency := int(float64(currentConcurrency) * 0.8)
			if newConcurrency < 100 {
				newConcurrency = 100
			}
			return cm.Set("concurrency.max_concurrent", newConcurrency)
		},
	})

	// 内存使用率过高时调整
	ate.rules = append(ate.rules, &TuningRule{
		ID:       "high_memory_adjust",
		Name:     "高内存使用率调整",
		CoolDown: 5 * time.Minute,
		Condition: func(metrics *PerformanceMetrics) bool {
			return metrics.Memory > 0.85
		},
		Action: func(cm *ConfigManager, metrics *PerformanceMetrics) error {
			// 降低缓存大小
			currentCacheSize, _ := cm.GetInt("cache.max_size")
			newCacheSize := int(float64(currentCacheSize) * 0.8)
			if newCacheSize < 100 {
				newCacheSize = 100
			}
			return cm.Set("cache.max_size", newCacheSize)
		},
	})

	// 响应时间过长时调整
	ate.rules = append(ate.rules, &TuningRule{
		ID:       "high_response_time_adjust",
		Name:     "高响应时间调整",
		CoolDown: 3 * time.Minute,
		Condition: func(metrics *PerformanceMetrics) bool {
			return metrics.AvgResponseTime > 1*time.Second
		},
		Action: func(cm *ConfigManager, metrics *PerformanceMetrics) error {
			// 降低API限流
			currentLimit, _ := cm.GetInt("rate_limit.api.limit")
			newLimit := int(float64(currentLimit) * 0.9)
			if newLimit < 100 {
				newLimit = 100
			}
			return cm.Set("rate_limit.api.limit", newLimit)
		},
	})

	// 数据库连接使用率过高时调整
	ate.rules = append(ate.rules, &TuningRule{
		ID:       "high_db_conn_adjust",
		Name:     "高数据库连接使用率调整",
		CoolDown: 5 * time.Minute,
		Condition: func(metrics *PerformanceMetrics) bool {
			return metrics.DatabaseConnUsage > 0.9
		},
		Action: func(cm *ConfigManager, metrics *PerformanceMetrics) error {
			// 增加数据库连接数
			currentMaxOpen, _ := cm.GetInt("db.max_open_conns")
			newMaxOpen := currentMaxOpen + 5
			if newMaxOpen > 100 {
				newMaxOpen = 100
			}
			return cm.Set("db.max_open_conns", newMaxOpen)
		},
	})
}

// start 启动自动调优
func (ate *AutoTuningEngine) start() {
	ticker := time.NewTicker(30 * time.Second) // 每30秒检查一次
	defer ticker.Stop()

	for range ticker.C {
		if !ate.enabled {
			continue
		}

		ate.tune()
	}
}

// tune 执行调优
func (ate *AutoTuningEngine) tune() {
	ate.mutex.Lock()
	defer ate.mutex.Unlock()

	// 获取当前性能指标
	metrics := ate.getCurrentMetrics()
	if metrics == nil {
		return
	}

	ate.metrics = metrics

	// 应用调优规则
	for _, rule := range ate.rules {
		if time.Since(rule.LastApplied) < rule.CoolDown {
			continue
		}

		if rule.Condition(metrics) {
			if err := rule.Action(ate.configManager, metrics); err != nil {
				log.Printf("Failed to apply tuning rule %s: %v", rule.Name, err)
			} else {
				log.Printf("Applied tuning rule: %s", rule.Name)
				rule.LastApplied = time.Now()
			}
		}
	}

	ate.lastTuning = time.Now()
}

// getCurrentMetrics 获取当前性能指标
func (ate *AutoTuningEngine) getCurrentMetrics() *PerformanceMetrics {
	// 这里应该从性能监控系统获取实际指标
	// 暂时返回模拟数据
	return &PerformanceMetrics{
		CPU:               0.75,
		Memory:            0.65,
		AvgResponseTime:   300 * time.Millisecond,
		RequestRate:       100.0,
		ErrorRate:         0.01,
		DatabaseConnUsage: 0.7,
		CacheHitRate:      0.85,
		UpdatedAt:         time.Now(),
	}
}

// Enable 启用自动调优
func (ate *AutoTuningEngine) Enable() {
	ate.mutex.Lock()
	defer ate.mutex.Unlock()
	ate.enabled = true
}

// Disable 禁用自动调优
func (ate *AutoTuningEngine) Disable() {
	ate.mutex.Lock()
	defer ate.mutex.Unlock()
	ate.enabled = false
}

// GetStatus 获取调优状态
func (ate *AutoTuningEngine) GetStatus() map[string]interface{} {
	ate.mutex.Lock()
	defer ate.mutex.Unlock()

	return map[string]interface{}{
		"enabled":     ate.enabled,
		"last_tuning": ate.lastTuning,
		"metrics":     ate.metrics,
		"rules_count": len(ate.rules),
	}
}

// 全局配置管理器
var GlobalConfigManager *ConfigManager

// InitConfigManager 初始化配置管理器
func InitConfigManager(db *gorm.DB, redisClient *redis.Client) {
	GlobalConfigManager = NewConfigManager(db, redisClient)
	log.Println("Configuration manager initialized successfully")
}
