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

package logger

import (
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"time"

	"github.com/linux-do/cdk-office/internal/config"
	"gopkg.in/natefinch/lumberjack.v2"
)

var (
	// Logger 全局日志实例
	Logger *log.Logger
	
	// writer 日志写入器
	writer io.Writer
)

// Init 初始化日志
func Init() {
	// 根据配置设置日志输出
	switch config.Config.Log.Output {
	case "file":
		// 确保日志目录存在
		if config.Config.Log.FilePath != "" {
			logDir := filepath.Dir(config.Config.Log.FilePath)
			if _, err := os.Stat(logDir); os.IsNotExist(err) {
				if err := os.MkdirAll(logDir, 0755); err != nil {
					log.Printf("[LOGGER] failed to create log directory: %v", err)
				}
			}
		}
		
		// 使用lumberjack进行日志轮转
		writer = &lumberjack.Logger{
			Filename:   config.Config.Log.FilePath,
			MaxSize:    config.Config.Log.MaxSize,    // MB
			MaxAge:     config.Config.Log.MaxAge,     // days
			MaxBackups: config.Config.Log.MaxBackups, // 保留文件个数
			Compress:   config.Config.Log.Compress,   // 是否压缩
		}
	case "stdout":
		fallthrough
	default:
		writer = os.Stdout
	}
	
	// 创建日志实例
	Logger = log.New(writer, "", log.LstdFlags|log.Lshortfile)
	
	// 设置日志前缀
	Logger.SetPrefix("[CDK-OFFICE] ")
	
	log.Println("[LOGGER] logger initialized")
}

// GetLogger 获取日志实例
func GetLogger() *log.Logger {
	return Logger
}

// Info 记录INFO级别日志
func Info(format string, v ...interface{}) {
	Logger.Printf("[INFO] "+format, v...)
}

// Warn 记录WARN级别日志
func Warn(format string, v ...interface{}) {
	Logger.Printf("[WARN] "+format, v...)
}

// Error 记录ERROR级别日志
func Error(format string, v ...interface{}) {
	Logger.Printf("[ERROR] "+format, v...)
}

// Debug 记录DEBUG级别日志
func Debug(format string, v ...interface{}) {
	if config.Config.Log.Level == "debug" {
		Logger.Printf("[DEBUG] "+format, v...)
	}
}

// Fatal 记录FATAL级别日志并退出
func Fatal(format string, v ...interface{}) {
	Logger.Printf("[FATAL] "+format, v...)
	os.Exit(1)
}