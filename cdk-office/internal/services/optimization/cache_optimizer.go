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
	"sync"
	"time"
)

// CacheLevel 缓存级别
type CacheLevel int

const (
	CacheLevelL1 CacheLevel = iota // 内存缓存
	CacheLevelL2                   // Redis缓存
	CacheLevelL3                   // 数据库缓存
)

// CacheItem 缓存项
type CacheItem struct {
	Key        string      `json:"key"`
	Value      interface{} `json:"value"`
	ExpiresAt  time.Time   `json:"expires_at"`
	AccessTime time.Time   `json:"access_time"`
	HitCount   int64       `json:"hit_count"`
	Size       int64       `json:"size"`
}

// IsExpired 是否过期
func (ci *CacheItem) IsExpired() bool {
	return !ci.ExpiresAt.IsZero() && time.Now().After(ci.ExpiresAt)
}

// Touch 更新访问时间
func (ci *CacheItem) Touch() {
	ci.AccessTime = time.Now()
	ci.HitCount++
}

// MultiLevelCache 多级缓存
type MultiLevelCache struct {
	l1Cache *LRUCache
	l2Cache RedisInterface
	l3Cache DatabaseInterface
	stats   *CacheStats
	mutex   sync.RWMutex
	config  *CacheConfig
}

// CacheConfig 缓存配置
type CacheConfig struct {
	L1MaxSize    int           `json:"l1_max_size"`   // L1缓存最大条目数
	L1TTL        time.Duration `json:"l1_ttl"`        // L1缓存TTL
	L2TTL        time.Duration `json:"l2_ttl"`        // L2缓存TTL
	L3TTL        time.Duration `json:"l3_ttl"`        // L3缓存TTL
	EnableL1     bool          `json:"enable_l1"`     // 启用L1缓存
	EnableL2     bool          `json:"enable_l2"`     // 启用L2缓存
	EnableL3     bool          `json:"enable_l3"`     // 启用L3缓存
	WriteThrough bool          `json:"write_through"` // 写穿透
	WriteBack    bool          `json:"write_back"`    // 写回
}

// DefaultCacheConfig 默认缓存配置
func DefaultCacheConfig() *CacheConfig {
	return &CacheConfig{
		L1MaxSize:    1000,
		L1TTL:        5 * time.Minute,
		L2TTL:        30 * time.Minute,
		L3TTL:        2 * time.Hour,
		EnableL1:     true,
		EnableL2:     true,
		EnableL3:     false,
		WriteThrough: true,
		WriteBack:    false,
	}
}

// RedisInterface Redis接口
type RedisInterface interface {
	Get(key string) ([]byte, error)
	Set(key string, value []byte, expiration time.Duration) error
	Delete(key string) error
	Exists(key string) bool
}

// DatabaseInterface 数据库接口
type DatabaseInterface interface {
	Get(key string) ([]byte, error)
	Set(key string, value []byte, expiration time.Duration) error
	Delete(key string) error
}

// LRUCache LRU缓存
type LRUCache struct {
	capacity int
	items    map[string]*CacheItem
	head     *CacheItem
	tail     *CacheItem
	mutex    sync.RWMutex
}

// NewLRUCache 创建LRU缓存
func NewLRUCache(capacity int) *LRUCache {
	lru := &LRUCache{
		capacity: capacity,
		items:    make(map[string]*CacheItem),
	}

	// 创建虚拟头尾节点
	lru.head = &CacheItem{}
	lru.tail = &CacheItem{}
	lru.head.next = lru.tail
	lru.tail.prev = lru.head

	return lru
}

// 为CacheItem添加链表字段
type CacheItemNode struct {
	*CacheItem
	prev *CacheItemNode
	next *CacheItemNode
}

// Get 获取缓存项
func (lru *LRUCache) Get(key string) (*CacheItem, bool) {
	lru.mutex.Lock()
	defer lru.mutex.Unlock()

	if item, exists := lru.items[key]; exists {
		if item.IsExpired() {
			delete(lru.items, key)
			return nil, false
		}

		item.Touch()
		lru.moveToHead(item)
		return item, true
	}

	return nil, false
}

// Set 设置缓存项
func (lru *LRUCache) Set(key string, value interface{}, ttl time.Duration) {
	lru.mutex.Lock()
	defer lru.mutex.Unlock()

	item := &CacheItem{
		Key:        key,
		Value:      value,
		ExpiresAt:  time.Now().Add(ttl),
		AccessTime: time.Now(),
		HitCount:   0,
		Size:       int64(len(fmt.Sprintf("%v", value))),
	}

	if existing, exists := lru.items[key]; exists {
		existing.Value = value
		existing.ExpiresAt = item.ExpiresAt
		existing.Touch()
		lru.moveToHead(existing)
	} else {
		if len(lru.items) >= lru.capacity {
			lru.removeTail()
		}
		lru.items[key] = item
		lru.addToHead(item)
	}
}

// Delete 删除缓存项
func (lru *LRUCache) Delete(key string) {
	lru.mutex.Lock()
	defer lru.mutex.Unlock()

	if item, exists := lru.items[key]; exists {
		delete(lru.items, key)
		lru.removeItem(item)
	}
}

// 内部方法（简化实现，实际需要完整的双向链表操作）
func (lru *LRUCache) moveToHead(item *CacheItem) {}
func (lru *LRUCache) addToHead(item *CacheItem)  {}
func (lru *LRUCache) removeItem(item *CacheItem) {}
func (lru *LRUCache) removeTail()                {}

// CacheStats 缓存统计
type CacheStats struct {
	L1Hits    int64 `json:"l1_hits"`
	L1Misses  int64 `json:"l1_misses"`
	L2Hits    int64 `json:"l2_hits"`
	L2Misses  int64 `json:"l2_misses"`
	L3Hits    int64 `json:"l3_hits"`
	L3Misses  int64 `json:"l3_misses"`
	TotalGets int64 `json:"total_gets"`
	TotalSets int64 `json:"total_sets"`
	Evictions int64 `json:"evictions"`
	mutex     sync.RWMutex
}

// IncrementL1Hit L1命中+1
func (cs *CacheStats) IncrementL1Hit() {
	cs.mutex.Lock()
	defer cs.mutex.Unlock()
	cs.L1Hits++
	cs.TotalGets++
}

// IncrementL1Miss L1未命中+1
func (cs *CacheStats) IncrementL1Miss() {
	cs.mutex.Lock()
	defer cs.mutex.Unlock()
	cs.L1Misses++
}

// 其他统计方法...
func (cs *CacheStats) IncrementL2Hit() {
	cs.mutex.Lock()
	defer cs.mutex.Unlock()
	cs.L2Hits++
	cs.TotalGets++
}
func (cs *CacheStats) IncrementL2Miss() { cs.mutex.Lock(); defer cs.mutex.Unlock(); cs.L2Misses++ }
func (cs *CacheStats) IncrementL3Hit() {
	cs.mutex.Lock()
	defer cs.mutex.Unlock()
	cs.L3Hits++
	cs.TotalGets++
}
func (cs *CacheStats) IncrementL3Miss()   { cs.mutex.Lock(); defer cs.mutex.Unlock(); cs.L3Misses++ }
func (cs *CacheStats) IncrementSet()      { cs.mutex.Lock(); defer cs.mutex.Unlock(); cs.TotalSets++ }
func (cs *CacheStats) IncrementEviction() { cs.mutex.Lock(); defer cs.mutex.Unlock(); cs.Evictions++ }

// GetStats 获取统计信息
func (cs *CacheStats) GetStats() map[string]interface{} {
	cs.mutex.RLock()
	defer cs.mutex.RUnlock()

	totalRequests := cs.TotalGets
	if totalRequests == 0 {
		totalRequests = 1
	}

	return map[string]interface{}{
		"l1_hits":          cs.L1Hits,
		"l1_misses":        cs.L1Misses,
		"l1_hit_rate":      float64(cs.L1Hits) / float64(cs.L1Hits+cs.L1Misses+1) * 100,
		"l2_hits":          cs.L2Hits,
		"l2_misses":        cs.L2Misses,
		"l2_hit_rate":      float64(cs.L2Hits) / float64(cs.L2Hits+cs.L2Misses+1) * 100,
		"l3_hits":          cs.L3Hits,
		"l3_misses":        cs.L3Misses,
		"l3_hit_rate":      float64(cs.L3Hits) / float64(cs.L3Hits+cs.L3Misses+1) * 100,
		"total_gets":       cs.TotalGets,
		"total_sets":       cs.TotalSets,
		"evictions":        cs.Evictions,
		"overall_hit_rate": float64(cs.L1Hits+cs.L2Hits+cs.L3Hits) / float64(totalRequests) * 100,
	}
}

// NewMultiLevelCache 创建多级缓存
func NewMultiLevelCache(config *CacheConfig, l2Cache RedisInterface, l3Cache DatabaseInterface) *MultiLevelCache {
	return &MultiLevelCache{
		l1Cache: NewLRUCache(config.L1MaxSize),
		l2Cache: l2Cache,
		l3Cache: l3Cache,
		stats:   &CacheStats{},
		config:  config,
	}
}

// Get 获取缓存值
func (mlc *MultiLevelCache) Get(ctx context.Context, key string) (interface{}, error) {
	// L1缓存查找
	if mlc.config.EnableL1 {
		if item, found := mlc.l1Cache.Get(key); found {
			mlc.stats.IncrementL1Hit()
			return item.Value, nil
		}
		mlc.stats.IncrementL1Miss()
	}

	// L2缓存查找
	if mlc.config.EnableL2 && mlc.l2Cache != nil {
		if data, err := mlc.l2Cache.Get(key); err == nil {
			mlc.stats.IncrementL2Hit()

			var value interface{}
			if err := json.Unmarshal(data, &value); err == nil {
				// 回填L1缓存
				if mlc.config.EnableL1 {
					mlc.l1Cache.Set(key, value, mlc.config.L1TTL)
				}
				return value, nil
			}
		}
		mlc.stats.IncrementL2Miss()
	}

	// L3缓存查找
	if mlc.config.EnableL3 && mlc.l3Cache != nil {
		if data, err := mlc.l3Cache.Get(key); err == nil {
			mlc.stats.IncrementL3Hit()

			var value interface{}
			if err := json.Unmarshal(data, &value); err == nil {
				// 回填上级缓存
				if mlc.config.EnableL2 && mlc.l2Cache != nil {
					if jsonData, err := json.Marshal(value); err == nil {
						mlc.l2Cache.Set(key, jsonData, mlc.config.L2TTL)
					}
				}
				if mlc.config.EnableL1 {
					mlc.l1Cache.Set(key, value, mlc.config.L1TTL)
				}
				return value, nil
			}
		}
		mlc.stats.IncrementL3Miss()
	}

	return nil, fmt.Errorf("cache miss for key: %s", key)
}

// Set 设置缓存值
func (mlc *MultiLevelCache) Set(ctx context.Context, key string, value interface{}, ttl time.Duration) error {
	mlc.stats.IncrementSet()

	// 写穿透模式：同时写入所有级别
	if mlc.config.WriteThrough {
		// L1缓存
		if mlc.config.EnableL1 {
			l1TTL := ttl
			if l1TTL > mlc.config.L1TTL {
				l1TTL = mlc.config.L1TTL
			}
			mlc.l1Cache.Set(key, value, l1TTL)
		}

		// L2缓存
		if mlc.config.EnableL2 && mlc.l2Cache != nil {
			if jsonData, err := json.Marshal(value); err == nil {
				l2TTL := ttl
				if l2TTL > mlc.config.L2TTL {
					l2TTL = mlc.config.L2TTL
				}
				mlc.l2Cache.Set(key, jsonData, l2TTL)
			}
		}

		// L3缓存
		if mlc.config.EnableL3 && mlc.l3Cache != nil {
			if jsonData, err := json.Marshal(value); err == nil {
				l3TTL := ttl
				if l3TTL > mlc.config.L3TTL {
					l3TTL = mlc.config.L3TTL
				}
				mlc.l3Cache.Set(key, jsonData, l3TTL)
			}
		}
	} else {
		// 写回模式：只写L1，异步写回其他级别
		if mlc.config.EnableL1 {
			mlc.l1Cache.Set(key, value, ttl)
		}

		// 异步写回
		go func() {
			if mlc.config.EnableL2 && mlc.l2Cache != nil {
				if jsonData, err := json.Marshal(value); err == nil {
					mlc.l2Cache.Set(key, jsonData, ttl)
				}
			}
		}()
	}

	return nil
}

// Delete 删除缓存值
func (mlc *MultiLevelCache) Delete(ctx context.Context, key string) error {
	// 删除所有级别
	if mlc.config.EnableL1 {
		mlc.l1Cache.Delete(key)
	}

	if mlc.config.EnableL2 && mlc.l2Cache != nil {
		mlc.l2Cache.Delete(key)
	}

	if mlc.config.EnableL3 && mlc.l3Cache != nil {
		mlc.l3Cache.Delete(key)
	}

	return nil
}

// GetStats 获取缓存统计
func (mlc *MultiLevelCache) GetStats() map[string]interface{} {
	return mlc.stats.GetStats()
}

// Preload 预加载缓存
func (mlc *MultiLevelCache) Preload(ctx context.Context, keys []string, loader func(string) (interface{}, error)) error {
	for _, key := range keys {
		if _, err := mlc.Get(ctx, key); err != nil {
			// 缓存未命中，加载数据
			if value, err := loader(key); err == nil {
				mlc.Set(ctx, key, value, mlc.config.L1TTL)
			}
		}
	}
	return nil
}

// Warm 缓存预热
func (mlc *MultiLevelCache) Warm(ctx context.Context, warmupData map[string]interface{}) error {
	for key, value := range warmupData {
		mlc.Set(ctx, key, value, mlc.config.L1TTL)
	}
	return nil
}

// CacheManager 缓存管理器
type CacheManager struct {
	caches map[string]*MultiLevelCache
	mutex  sync.RWMutex
}

// NewCacheManager 创建缓存管理器
func NewCacheManager() *CacheManager {
	return &CacheManager{
		caches: make(map[string]*MultiLevelCache),
	}
}

// RegisterCache 注册缓存
func (cm *CacheManager) RegisterCache(name string, cache *MultiLevelCache) {
	cm.mutex.Lock()
	defer cm.mutex.Unlock()
	cm.caches[name] = cache
}

// GetCache 获取缓存
func (cm *CacheManager) GetCache(name string) (*MultiLevelCache, bool) {
	cm.mutex.RLock()
	defer cm.mutex.RUnlock()
	cache, exists := cm.caches[name]
	return cache, exists
}

// GetAllStats 获取所有缓存统计
func (cm *CacheManager) GetAllStats() map[string]interface{} {
	cm.mutex.RLock()
	defer cm.mutex.RUnlock()

	stats := make(map[string]interface{})
	for name, cache := range cm.caches {
		stats[name] = cache.GetStats()
	}
	return stats
}

// 全局缓存管理器
var GlobalCacheManager = NewCacheManager()

// InitCacheOptimization 初始化缓存优化
func InitCacheOptimization(l2Cache RedisInterface, l3Cache DatabaseInterface) {
	// 主缓存
	mainConfig := DefaultCacheConfig()
	mainConfig.L1MaxSize = 2000
	mainConfig.L1TTL = 10 * time.Minute
	mainCache := NewMultiLevelCache(mainConfig, l2Cache, l3Cache)
	GlobalCacheManager.RegisterCache("main", mainCache)

	// 用户会话缓存
	sessionConfig := &CacheConfig{
		L1MaxSize:    500,
		L1TTL:        30 * time.Minute,
		L2TTL:        2 * time.Hour,
		EnableL1:     true,
		EnableL2:     true,
		EnableL3:     false,
		WriteThrough: true,
	}
	sessionCache := NewMultiLevelCache(sessionConfig, l2Cache, nil)
	GlobalCacheManager.RegisterCache("session", sessionCache)

	// 文档元数据缓存
	docConfig := &CacheConfig{
		L1MaxSize:    1000,
		L1TTL:        5 * time.Minute,
		L2TTL:        30 * time.Minute,
		EnableL1:     true,
		EnableL2:     true,
		EnableL3:     false,
		WriteThrough: false,
		WriteBack:    true,
	}
	docCache := NewMultiLevelCache(docConfig, l2Cache, nil)
	GlobalCacheManager.RegisterCache("document", docCache)
}
