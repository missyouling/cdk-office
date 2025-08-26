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

// Document 文档模型
type Document struct {
	ID             string    `json:"id" gorm:"type:uuid;primary_key;default:gen_random_uuid()"`
	TeamID         string    `json:"team_id" gorm:"type:uuid;not null;index"`
	Name           string    `json:"name" gorm:"size:255;not null"`
	Description    string    `json:"description" gorm:"type:text"`
	FileName       string    `json:"file_name" gorm:"size:255;not null"`
	FilePath       string    `json:"file_path" gorm:"size:500;not null"`
	FileSize       int64     `json:"file_size" gorm:"not null"`
	FileType       string    `json:"file_type" gorm:"size:50;not null"`
	MimeType       string    `json:"mime_type" gorm:"size:100"`
	Status         string    `json:"status" gorm:"size:20;default:'active'"` // active, archived, deleted
	Version        int       `json:"version" gorm:"default:1"`
	CreatedBy      string    `json:"created_by" gorm:"type:uuid;not null"`
	UpdatedBy      string    `json:"updated_by" gorm:"type:uuid"`
	DifyDocumentID string    `json:"dify_document_id" gorm:"size:100;index"` // Dify知识库中的文档ID
	LastSyncAt     time.Time `json:"last_sync_at"`                           // 最后同步时间
	Tags           []string  `json:"tags" gorm:"type:text[]"`
	CreatedAt      time.Time `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt      time.Time `json:"updated_at" gorm:"autoUpdateTime"`
}

// BeforeCreate 创建前钩子
func (d *Document) BeforeCreate(tx *gorm.DB) error {
	if d.ID == "" {
		d.ID = uuid.New().String()
	}
	return nil
}

// DocumentVersion 文档版本模型
type DocumentVersion struct {
	ID         string    `json:"id" gorm:"type:uuid;primary_key;default:gen_random_uuid()"`
	DocumentID string    `json:"document_id" gorm:"type:uuid;not null;index"`
	Version    int       `json:"version" gorm:"not null"`
	FileName   string    `json:"file_name" gorm:"size:255;not null"`
	FilePath   string    `json:"file_path" gorm:"size:500;not null"`
	FileSize   int64     `json:"file_size" gorm:"not null"`
	CreatedBy  string    `json:"created_by" gorm:"type:uuid;not null"`
	CreatedAt  time.Time `json:"created_at" gorm:"autoCreateTime"`
}

// DocumentTag 文档标签模型
type DocumentTag struct {
	ID        string    `json:"id" gorm:"type:uuid;primary_key;default:gen_random_uuid()"`
	TeamID    string    `json:"team_id" gorm:"type:uuid;not null;index"`
	Name      string    `json:"name" gorm:"size:50;not null;uniqueIndex:idx_team_tag"`
	Color     string    `json:"color" gorm:"size:20"` // 标签颜色
	CreatedBy string    `json:"created_by" gorm:"type:uuid;not null"`
	CreatedAt time.Time `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt time.Time `json:"updated_at" gorm:"autoUpdateTime"`
}

// DocumentTagRelation 文档标签关联模型
type DocumentTagRelation struct {
	ID         string    `json:"id" gorm:"type:uuid;primary_key;default:gen_random_uuid()"`
	DocumentID string    `json:"document_id" gorm:"type:uuid;not null;index:idx_doc_tag"`
	TagID      string    `json:"tag_id" gorm:"type:uuid;not null;index:idx_doc_tag"`
	CreatedAt  time.Time `json:"created_at" gorm:"autoCreateTime"`
}
