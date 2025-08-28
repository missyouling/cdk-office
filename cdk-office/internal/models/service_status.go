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

package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// ServiceHealthStatus 服务健康状态模型
type ServiceHealthStatus struct {
	ID           uuid.UUID `json:"id" gorm:"type:uuid;primary_key;default:gen_random_uuid()"`
	ServiceName  string    `json:"service_name" gorm:"size:100;not null;index"`
	Status       string    `json:"status" gorm:"size:20;not null"` // healthy, unhealthy, degraded
	ResponseTime int64     `json:"response_time"`                  // 响应时间（毫秒）
	StatusCode   int       `json:"status_code"`                    // HTTP状态码
	ErrorMessage string    `json:"error_message" gorm:"type:text"`
	Details      string    `json:"details" gorm:"type:text"` // JSON格式的详细信息
	CheckedAt    time.Time `json:"checked_at" gorm:"not null;index"`
	CreatedAt    time.Time `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt    time.Time `json:"updated_at" gorm:"autoUpdateTime"`
}

// TableName 指定表名
func (ServiceHealthStatus) TableName() string {
	return "service_statuses"
}

// BeforeCreate GORM钩子：创建前生成UUID
func (s *ServiceHealthStatus) BeforeCreate(tx *gorm.DB) error {
	if s.ID == uuid.Nil {
		s.ID = uuid.New()
	}
	return nil
}

// IsHealthy 判断服务是否健康
func (s *ServiceHealthStatus) IsHealthy() bool {
	return s.Status == "healthy"
}

// IsCritical 判断是否为关键服务
func (s *ServiceHealthStatus) IsCritical() bool {
	criticalServices := map[string]bool{
		"postgresql_database": true,
		"redis_cache":         true,
		"wechat_api":          true,
		"supabase_storage":    true,
	}
	return criticalServices[s.ServiceName]
}

// GetStatusLevel 获取状态级别（用于排序和展示）
func (s *ServiceHealthStatus) GetStatusLevel() int {
	switch s.Status {
	case "healthy":
		return 0
	case "degraded":
		return 1
	case "unhealthy":
		return 2
	default:
		return 3
	}
}

// GetResponseTimeCategory 获取响应时间分类
func (s *ServiceHealthStatus) GetResponseTimeCategory() string {
	if s.ResponseTime < 100 {
		return "excellent"
	} else if s.ResponseTime < 300 {
		return "good"
	} else if s.ResponseTime < 500 {
		return "acceptable"
	} else {
		return "slow"
	}
}
