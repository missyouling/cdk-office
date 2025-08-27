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

package isolation

import (
	"context"
	"encoding/json"
	"time"

	"github.com/go-redis/redis/v8"
)

// RedisCache Redis缓存实现
type RedisCache struct {
	client *redis.Client
	ctx    context.Context
}

// NewRedisCache 创建Redis缓存
func NewRedisCache(addr, password string, db int) *RedisCache {
	rdb := redis.NewClient(&redis.Options{
		Addr:     addr,
		Password: password,
		DB:       db,
	})

	return &RedisCache{
		client: rdb,
		ctx:    context.Background(),
	}
}

// Get 获取缓存值
func (c *RedisCache) Get(key string) (interface{}, error) {
	val, err := c.client.Get(c.ctx, key).Result()
	if err != nil {
		if err == redis.Nil {
			return nil, nil // 键不存在
		}
		return nil, err
	}

	// 尝试反序列化为通用对象
	var result interface{}
	if err := json.Unmarshal([]byte(val), &result); err != nil {
		// 如果反序列化失败，返回原始字符串
		return val, nil
	}

	return result, nil
}

// Set 设置缓存值
func (c *RedisCache) Set(key string, value interface{}, expiration time.Duration) error {
	// 序列化值
	data, err := json.Marshal(value)
	if err != nil {
		return err
	}

	return c.client.Set(c.ctx, key, data, expiration).Err()
}

// Delete 删除缓存值
func (c *RedisCache) Delete(key string) error {
	return c.client.Del(c.ctx, key).Err()
}

// MemoryCache 内存缓存实现（用于测试）
type MemoryCache struct {
	data map[string]cacheItem
}

type cacheItem struct {
	value     interface{}
	expiresAt time.Time
}

// NewMemoryCache 创建内存缓存
func NewMemoryCache() *MemoryCache {
	return &MemoryCache{
		data: make(map[string]cacheItem),
	}
}

// Get 获取缓存值
func (c *MemoryCache) Get(key string) (interface{}, error) {
	item, exists := c.data[key]
	if !exists {
		return nil, nil
	}

	// 检查是否过期
	if time.Now().After(item.expiresAt) {
		delete(c.data, key)
		return nil, nil
	}

	return item.value, nil
}

// Set 设置缓存值
func (c *MemoryCache) Set(key string, value interface{}, expiration time.Duration) error {
	expiresAt := time.Now().Add(expiration)
	c.data[key] = cacheItem{
		value:     value,
		expiresAt: expiresAt,
	}
	return nil
}

// Delete 删除缓存值
func (c *MemoryCache) Delete(key string) error {
	delete(c.data, key)
	return nil
}
