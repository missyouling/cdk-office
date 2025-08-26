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

// QRCodeForm 二维码表单模型
type QRCodeForm struct {
	ID          string    `json:"id" gorm:"type:uuid;primary_key;default:gen_random_uuid()"`
	TeamID      string    `json:"team_id" gorm:"type:uuid;not null;index"`
	FormName    string    `json:"form_name" gorm:"size:100;not null"`
	FormType    string    `json:"form_type" gorm:"size:50"` // survey, registration, feedback
	Description string    `json:"description" gorm:"type:text"`
	CreatedBy   string    `json:"created_by" gorm:"type:uuid;not null"`
	UpdatedBy   string    `json:"updated_by" gorm:"type:uuid"`
	CreatedAt   time.Time `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt   time.Time `json:"updated_at" gorm:"autoUpdateTime"`
}

// BeforeCreate 创建前钩子
func (q *QRCodeForm) BeforeCreate(tx *gorm.DB) error {
	if q.ID == "" {
		q.ID = uuid.New().String()
	}
	return nil
}

// QRCodeFormField 表单字段定义模型
type QRCodeFormField struct {
	ID           string    `json:"id" gorm:"type:uuid;primary_key;default:gen_random_uuid()"`
	FormID       string    `json:"form_id" gorm:"type:uuid;not null;index"`
	FieldKey     string    `json:"field_key" gorm:"size:100;not null"`
	FieldLabel   string    `json:"field_label" gorm:"size:255;not null"`
	FieldType    string    `json:"field_type" gorm:"size:50"` // text, number, select, radio, checkbox
	IsRequired   bool      `json:"is_required" gorm:"default:false"`
	DefaultValue string    `json:"default_value" gorm:"size:255"`
	Options      []string  `json:"options" gorm:"type:text[]"`
	DisplayOrder int       `json:"display_order" gorm:"default:0"`
	CreatedAt    time.Time `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt    time.Time `json:"updated_at" gorm:"autoUpdateTime"`
}

// QRCodeRecord 二维码生成记录模型
type QRCodeRecord struct {
	ID        string    `json:"id" gorm:"type:uuid;primary_key;default:gen_random_uuid()"`
	FormID    string    `json:"form_id" gorm:"type:uuid;not null;index"`
	Content   string    `json:"content" gorm:"type:text;not null"`
	QRCodeURL string    `json:"qrcode_url" gorm:"size:500"`
	ExpireAt  time.Time `json:"expire_at"`
	CreatedBy string    `json:"created_by" gorm:"type:uuid;not null"`
	CreatedAt time.Time `json:"created_at" gorm:"autoCreateTime"`
}

// QRCodeFormSubmission 表单提交记录模型
type QRCodeFormSubmission struct {
	ID         string    `json:"id" gorm:"type:uuid;primary_key;default:gen_random_uuid()"`
	FormID     string    `json:"form_id" gorm:"type:uuid;not null;index"`
	SubmitData string    `json:"submit_data" gorm:"type:text"` // JSON格式的提交数据
	SubmitIP   string    `json:"submit_ip" gorm:"size:45"`
	CreatedAt  time.Time `json:"created_at" gorm:"autoCreateTime"`
}
