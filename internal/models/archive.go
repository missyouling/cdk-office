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

// ArchiveRule 归档规则模型
type ArchiveRule struct {
	ID         string    `json:"id" gorm:"type:uuid;primary_key;default:gen_random_uuid()"`
	TeamID     string    `json:"team_id" gorm:"type:uuid;not null;index"`
	RuleName   string    `json:"rule_name" gorm:"size:100;not null"`
	RuleType   string    `json:"rule_type" gorm:"size:50"`     // time_based, size_based, tag_based, custom
	RuleConfig string    `json:"rule_config" gorm:"type:text"` // JSON格式的规则配置
	TargetPath string    `json:"target_path" gorm:"size:500"`
	IsActive   bool      `json:"is_active" gorm:"default:true"`
	LastRun    time.Time `json:"last_run"`
	NextRun    time.Time `json:"next_run"`
	CreatedBy  string    `json:"created_by" gorm:"type:uuid"`
	UpdatedBy  string    `json:"updated_by" gorm:"type:uuid"`
	CreatedAt  time.Time `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt  time.Time `json:"updated_at" gorm:"autoUpdateTime"`
}

// BeforeCreate 创建前钩子
func (a *ArchiveRule) BeforeCreate(tx *gorm.DB) error {
	if a.ID == "" {
		a.ID = uuid.New().String()
	}
	return nil
}

// ArchiveRecord 归档记录模型
type ArchiveRecord struct {
	ID            string    `json:"id" gorm:"type:uuid;primary_key;default:gen_random_uuid()"`
	RuleID        string    `json:"rule_id" gorm:"type:uuid;not null;index"`
	DocumentID    string    `json:"document_id" gorm:"type:uuid;not null;index"`
	OriginalPath  string    `json:"original_path" gorm:"size:500"`
	ArchivePath   string    `json:"archive_path" gorm:"size:500"`
	ArchiveStatus string    `json:"archive_status" gorm:"size:20;default:'pending'"` // pending, processing, completed, failed
	ArchiveDate   time.Time `json:"archive_date" gorm:"autoCreateTime"`
	Metadata      string    `json:"metadata" gorm:"type:text"` // JSON格式的元数据
}

// ArchiveCatalog 归档目录模型
type ArchiveCatalog struct {
	ID            string    `json:"id" gorm:"type:uuid;primary_key;default:gen_random_uuid()"`
	TeamID        string    `json:"team_id" gorm:"type:uuid;not null;index"`
	CatalogName   string    `json:"catalog_name" gorm:"size:255"`
	CatalogPath   string    `json:"catalog_path" gorm:"size:500"`
	DocumentCount int       `json:"document_count" gorm:"default:0"`
	TotalSize     int64     `json:"total_size" gorm:"default:0"`
	CreatedDate   time.Time `json:"created_date" gorm:"autoCreateTime"`
	Metadata      string    `json:"metadata" gorm:"type:text"` // JSON格式的目录元数据
}
