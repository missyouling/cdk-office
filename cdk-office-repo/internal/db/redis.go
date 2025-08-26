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

package db

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/linux-do/cdk-office/internal/config"
)

var (
	// RedisClient 全局Redis客户端实例
	RedisClient *redis.Client
)

// InitRedis 初始化Redis连接
func InitRedis() {
	// 创建Redis客户端
	RedisClient = redis.NewClient(&redis.Options{
		Addr:         fmt.Sprintf("%s:%d", config.Config.Redis.Host, config.Config.Redis.Port),
		Username:     config.Config.Redis.Username,
		Password:     config.Config.Redis.Password,
		DB:           config.Config.Redis.DB,
		PoolSize:     config.Config.Redis.PoolSize,
		MinIdleConns: config.Config.Redis.MinIdleConn,
		DialTimeout:  time.Duration(config.Config.Redis.DialTimeout) * time.Second,
		ReadTimeout:  time.Duration(config.Config.Redis.ReadTimeout) * time.Second,
		WriteTimeout: time.Duration(config.Config.Redis.WriteTimeout) * time.Second,
	})

	// 测试连接
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	_, err := RedisClient.Ping(ctx).Result()
	if err != nil {
		log.Fatalf("[REDIS] failed to connect to Redis: %v", err)
	}

	log.Println("[REDIS] Redis connection initialized")
}

// GetRedis 获取Redis客户端实例
func GetRedis() *redis.Client {
	return RedisClient
}

// CloseRedis 关闭Redis连接
func CloseRedis() {
	if RedisClient != nil {
		if err := RedisClient.Close(); err != nil {
			log.Printf("[REDIS] error closing Redis connection: %v", err)
		}
	}
	log.Println("[REDIS] Redis connection closed")
}
