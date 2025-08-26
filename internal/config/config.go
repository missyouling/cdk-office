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

package config

import (
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/spf13/viper"
)

// Load 加载配置
func Load() {
	// 设置配置文件名称
	viper.SetConfigName("config")

	// 添加配置文件搜索路径
	viper.AddConfigPath(".")
	viper.AddConfigPath("./config")
	viper.AddConfigPath("../config")

	// 设置配置文件类型
	viper.SetConfigType("yaml")

	// 设置环境变量前缀
	viper.SetEnvPrefix("CDK_OFFICE")

	// 自动绑定环境变量
	viper.AutomaticEnv()

	// 尝试读取配置文件
	if err := viper.ReadInConfig(); err != nil {
		// 如果没有找到配置文件，尝试使用默认配置
		if _, ok := err.(viper.ConfigFileNotFoundError); ok {
			log.Println("[CONFIG] config file not found, using default config")

			// 尝试使用VPS配置
			viper.SetConfigName("config.vps")
			if err := viper.ReadInConfig(); err != nil {
				log.Printf("[CONFIG] vps config file not found: %v", err)

				// 尝试使用示例配置
				viper.SetConfigName("config.example")
				if err := viper.ReadInConfig(); err != nil {
					log.Printf("[CONFIG] example config file not found: %v", err)
				}
			}
		} else {
			log.Fatalf("[CONFIG] fatal error config file: %v", err)
		}
	}

	// 将配置绑定到结构体
	if err := viper.Unmarshal(&Config); err != nil {
		log.Fatalf("[CONFIG] unable to decode into struct: %v", err)
	}

	// 确保日志目录存在
	if Config.Log.FilePath != "" {
		logDir := filepath.Dir(Config.Log.FilePath)
		if _, err := os.Stat(logDir); os.IsNotExist(err) {
			if err := os.MkdirAll(logDir, 0755); err != nil {
				log.Printf("[CONFIG] failed to create log directory: %v", err)
			}
		}
	}

	log.Printf("[CONFIG] using config file: %s", viper.ConfigFileUsed())
}

// GetConfig 获取配置实例
func GetConfig() configModel {
	return Config
}

// IsDevelopment 检查是否为开发环境
func IsDevelopment() bool {
	return Config.App.Env == "development"
}

// IsProduction 检查是否为生产环境
func IsProduction() bool {
	return Config.App.Env == "production"
}

// IsTesting 检查是否为测试环境
func IsTesting() bool {
	return Config.App.Env == "testing"
}

// GetDatabaseDSN 获取数据库连接字符串
func GetDatabaseDSN() string {
	return fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=disable",
		Config.Database.Host,
		Config.Database.Port,
		Config.Database.Username,
		Config.Database.Password,
		Config.Database.Database)
}

// GetRedisAddr 获取Redis地址
func GetRedisAddr() string {
	return fmt.Sprintf("%s:%d", Config.Redis.Host, Config.Redis.Port)
}
