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

	"gorm.io/datatypes"
)

// DataEntity 数据实体
type DataEntity struct {
	ID          string `json:"id" gorm:"type:uuid;primary_key;default:gen_random_uuid()"`
	Name        string `json:"name" gorm:"size:100;not null;uniqueIndex:idx_entity_name"`
	DisplayName string `json:"display_name" gorm:"size:100"`
	Description string `json:"description" gorm:"size:500"`
	Type        string `json:"type" gorm:"size:50;not null"`           // table, form, template, qrcode_form, print_template
	Source      string `json:"source" gorm:"size:100"`                 // 数据源：system, user_defined, imported
	Module      string `json:"module" gorm:"size:50"`                  // 所属模块：workflow, qrcode, survey, contract等
	Status      string `json:"status" gorm:"size:20;default:'active'"` // active, inactive, pending_approval

	// 关联信息
	TeamID    string `json:"team_id" gorm:"type:uuid;index"`
	CreatedBy string `json:"created_by" gorm:"type:uuid;not null"`
	UpdatedBy string `json:"updated_by" gorm:"type:uuid"`

	// 元数据
	Metadata datatypes.JSON `json:"metadata" gorm:"type:jsonb"` // 扩展元数据

	// 时间戳
	CreatedAt time.Time `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt time.Time `json:"updated_at" gorm:"autoUpdateTime"`

	// 关联字段定义
	FieldDefinitions []FieldDefinition `json:"field_definitions" gorm:"foreignKey:EntityID;constraint:OnDelete:CASCADE"`
}

// FieldDefinition 字段定义
type FieldDefinition struct {
	ID          string `json:"id" gorm:"type:uuid;primary_key;default:gen_random_uuid()"`
	EntityID    string `json:"entity_id" gorm:"type:uuid;not null;index"`
	FieldName   string `json:"field_name" gorm:"size:100;not null"`
	DisplayName string `json:"display_name" gorm:"size:100"`
	DataType    string `json:"data_type" gorm:"size:50;not null"` // string, number, boolean, date, json, array
	Description string `json:"description" gorm:"size:500"`

	// 验证规则
	IsRequired     bool   `json:"is_required" gorm:"default:false"`
	IsUnique       bool   `json:"is_unique" gorm:"default:false"`
	DefaultValue   string `json:"default_value" gorm:"size:255"`
	ValidationRule string `json:"validation_rule" gorm:"size:500"` // 正则表达式或其他验证规则
	MinLength      int    `json:"min_length" gorm:"default:0"`
	MaxLength      int    `json:"max_length" gorm:"default:0"`
	MinValue       string `json:"min_value" gorm:"size:100"` // 最小值（数字/日期）
	MaxValue       string `json:"max_value" gorm:"size:100"` // 最大值（数字/日期）

	// 选项配置（下拉框、单选、复选框等）
	Options       datatypes.JSON `json:"options" gorm:"type:jsonb"`      // 选项列表
	OptionsSource string         `json:"options_source" gorm:"size:100"` // 选项数据源

	// 显示配置
	DisplayOrder int  `json:"display_order" gorm:"default:0"`
	IsVisible    bool `json:"is_visible" gorm:"default:true"`
	IsEditable   bool `json:"is_editable" gorm:"default:true"`
	IsSearchable bool `json:"is_searchable" gorm:"default:false"`

	// 系统字段标识
	IsSystemField   bool   `json:"is_system_field" gorm:"default:false"`
	SystemFieldType string `json:"system_field_type" gorm:"size:50"` // created_at, updated_at, created_by等

	// 版本控制
	Version int    `json:"version" gorm:"default:1"`
	Status  string `json:"status" gorm:"size:20;default:'active'"` // active, inactive, pending_approval, pending_delete

	// 时间戳
	CreatedAt time.Time `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt time.Time `json:"updated_at" gorm:"autoUpdateTime"`

	// 关联扩展属性
	FieldExtensions []FieldExtension `json:"field_extensions" gorm:"foreignKey:FieldID;constraint:OnDelete:CASCADE"`
}

// FieldExtension 字段扩展属性
type FieldExtension struct {
	ID          string    `json:"id" gorm:"type:uuid;primary_key;default:gen_random_uuid()"`
	FieldID     string    `json:"field_id" gorm:"type:uuid;not null;index"`
	Key         string    `json:"key" gorm:"size:100;not null"`
	Value       string    `json:"value" gorm:"type:text"`
	DataType    string    `json:"data_type" gorm:"size:50;default:'string'"` // string, number, boolean, json
	Description string    `json:"description" gorm:"size:255"`
	CreatedAt   time.Time `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt   time.Time `json:"updated_at" gorm:"autoUpdateTime"`
}

// DictionaryChangeLog 数据字典变更记录
type DictionaryChangeLog struct {
	ID          string `json:"id" gorm:"type:uuid;primary_key;default:gen_random_uuid()"`
	EntityID    string `json:"entity_id" gorm:"type:uuid;index"`
	FieldID     string `json:"field_id" gorm:"type:uuid;index"`
	ChangeType  string `json:"change_type" gorm:"size:20;not null"`  // create, update, delete
	ChangeScope string `json:"change_scope" gorm:"size:20;not null"` // entity, field, extension

	// 变更内容
	OldValue          datatypes.JSON `json:"old_value" gorm:"type:jsonb"` // 变更前的值
	NewValue          datatypes.JSON `json:"new_value" gorm:"type:jsonb"` // 变更后的值
	ChangeReason      string         `json:"change_reason" gorm:"size:500"`
	ChangeDescription string         `json:"change_description" gorm:"type:text"`

	// 申请信息
	RequestedBy string    `json:"requested_by" gorm:"type:uuid;not null"`
	RequestedAt time.Time `json:"requested_at" gorm:"autoCreateTime"`
	TeamID      string    `json:"team_id" gorm:"type:uuid;index"`

	// 审批信息
	ApprovalStatus  string     `json:"approval_status" gorm:"size:20;default:'pending'"` // pending, approved, rejected
	ApprovedBy      string     `json:"approved_by" gorm:"type:uuid"`
	ApprovedAt      *time.Time `json:"approved_at"`
	ApprovalComment string     `json:"approval_comment" gorm:"type:text"`

	// 执行信息
	ExecutionStatus string     `json:"execution_status" gorm:"size:20;default:'pending'"` // pending, executed, failed
	ExecutedAt      *time.Time `json:"executed_at"`
	ExecutionError  string     `json:"execution_error" gorm:"type:text"`
}

// DictionaryTemplate 数据字典模板
type DictionaryTemplate struct {
	ID          string `json:"id" gorm:"type:uuid;primary_key;default:gen_random_uuid()"`
	Name        string `json:"name" gorm:"size:100;not null"`
	DisplayName string `json:"display_name" gorm:"size:100"`
	Description string `json:"description" gorm:"size:500"`
	Category    string `json:"category" gorm:"size:50"` // system, business, form, report
	Type        string `json:"type" gorm:"size:50"`     // qrcode_form, survey_form, workflow_form, print_template

	// 模板内容
	EntityTemplate datatypes.JSON `json:"entity_template" gorm:"type:jsonb"` // 实体模板
	FieldTemplates datatypes.JSON `json:"field_templates" gorm:"type:jsonb"` // 字段模板列表

	// 使用统计
	UsageCount int        `json:"usage_count" gorm:"default:0"`
	LastUsedAt *time.Time `json:"last_used_at"`

	// 版本信息
	Version  string `json:"version" gorm:"size:20;default:'1.0'"`
	IsPublic bool   `json:"is_public" gorm:"default:false"`
	IsSystem bool   `json:"is_system" gorm:"default:false"`

	CreatedBy string    `json:"created_by" gorm:"type:uuid;not null"`
	CreatedAt time.Time `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt time.Time `json:"updated_at" gorm:"autoUpdateTime"`
}

// DictionaryMapping 数据字典映射关系
type DictionaryMapping struct {
	ID             string         `json:"id" gorm:"type:uuid;primary_key;default:gen_random_uuid()"`
	SourceEntityID string         `json:"source_entity_id" gorm:"type:uuid;not null;index"`
	SourceFieldID  string         `json:"source_field_id" gorm:"type:uuid;not null;index"`
	TargetEntityID string         `json:"target_entity_id" gorm:"type:uuid;not null;index"`
	TargetFieldID  string         `json:"target_field_id" gorm:"type:uuid;not null;index"`
	MappingType    string         `json:"mapping_type" gorm:"size:50;not null"` // direct, transform, lookup
	MappingRule    datatypes.JSON `json:"mapping_rule" gorm:"type:jsonb"`       // 映射规则配置
	IsActive       bool           `json:"is_active" gorm:"default:true"`
	CreatedBy      string         `json:"created_by" gorm:"type:uuid;not null"`
	CreatedAt      time.Time      `json:"created_at" gorm:"autoCreateTime"`
}

// TableName 指定表名
func (DataEntity) TableName() string {
	return "data_entities"
}

func (FieldDefinition) TableName() string {
	return "field_definitions"
}

func (FieldExtension) TableName() string {
	return "field_extensions"
}

func (DictionaryChangeLog) TableName() string {
	return "dictionary_change_logs"
}

func (DictionaryTemplate) TableName() string {
	return "dictionary_templates"
}

func (DictionaryMapping) TableName() string {
	return "dictionary_mappings"
}
