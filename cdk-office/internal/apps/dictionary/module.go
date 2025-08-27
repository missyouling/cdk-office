/*
 * MIT License
 *
 * Copyright (c) 2025 CDK-Office
 */

package dictionary

import (
	"fmt"

	"gorm.io/gorm"

	"cdk-office/internal/config"
	"cdk-office/internal/models"
)

var (
	service *Service
	handler *Handler
	router  *Router
)

// InitDictionaryModule 初始化数据字典模块
func InitDictionaryModule(db *gorm.DB) error {
	// 自动迁移数据库表
	if err := db.AutoMigrate(
		&models.DataEntity{},
		&models.FieldDefinition{},
		&models.FieldExtension{},
		&models.DictionaryChangeLog{},
		&models.DictionaryTemplate{},
		&models.DictionaryMapping{},
	); err != nil {
		return fmt.Errorf("failed to migrate dictionary tables: %w", err)
	}

	// 初始化ODD客户端
	var oddClient *ODDClient
	if config.Config.ODD.Enabled {
		oddClient = NewODDClient(config.Config.ODD.BaseURL, config.Config.ODD.APIKey)
	}

	// 初始化服务
	service = NewService(db, oddClient)

	// 初始化处理器
	handler = NewHandler(service)

	// 初始化路由
	router = NewRouter(handler)

	// 初始化系统数据
	if err := initSystemEntities(db); err != nil {
		return fmt.Errorf("failed to init system entities: %w", err)
	}

	return nil
}

// GetHandler 获取处理器
func GetHandler() *Handler {
	return handler
}

// GetRouter 获取路由
func GetRouter() *Router {
	return router
}

// GetService 获取服务
func GetService() *Service {
	return service
}

// initSystemEntities 初始化系统实体
func initSystemEntities(db *gorm.DB) error {
	systemEntities := []*models.DataEntity{
		{
			Name:        "user",
			DisplayName: "用户",
			Description: "系统用户实体",
			Type:        "table",
			Source:      "system",
			Module:      "auth",
			Status:      "active",
			CreatedBy:   "system",
			UpdatedBy:   "system",
			FieldDefinitions: []models.FieldDefinition{
				{
					FieldName:       "id",
					DisplayName:     "用户ID",
					DataType:        "string",
					Description:     "用户唯一标识",
					IsRequired:      true,
					IsUnique:        true,
					IsSystemField:   true,
					SystemFieldType: "primary_key",
					DisplayOrder:    1,
				},
				{
					FieldName:    "username",
					DisplayName:  "用户名",
					DataType:     "string",
					Description:  "登录用户名",
					IsRequired:   true,
					IsUnique:     true,
					MaxLength:    50,
					DisplayOrder: 2,
				},
				{
					FieldName:    "email",
					DisplayName:  "邮箱",
					DataType:     "email",
					Description:  "用户邮箱地址",
					IsRequired:   true,
					IsUnique:     true,
					DisplayOrder: 3,
				},
				{
					FieldName:    "phone",
					DisplayName:  "手机号",
					DataType:     "phone",
					Description:  "用户手机号码",
					IsRequired:   false,
					IsUnique:     true,
					DisplayOrder: 4,
				},
				{
					FieldName:    "status",
					DisplayName:  "状态",
					DataType:     "string",
					Description:  "用户状态",
					IsRequired:   true,
					DefaultValue: "active",
					DisplayOrder: 5,
				},
				{
					FieldName:       "created_at",
					DisplayName:     "创建时间",
					DataType:        "date",
					Description:     "账户创建时间",
					IsRequired:      true,
					IsSystemField:   true,
					SystemFieldType: "created_at",
					DisplayOrder:    6,
				},
				{
					FieldName:       "updated_at",
					DisplayName:     "更新时间",
					DataType:        "date",
					Description:     "账户更新时间",
					IsRequired:      true,
					IsSystemField:   true,
					SystemFieldType: "updated_at",
					DisplayOrder:    7,
				},
			},
		},
		{
			Name:        "team",
			DisplayName: "团队",
			Description: "系统团队实体",
			Type:        "table",
			Source:      "system",
			Module:      "team",
			Status:      "active",
			CreatedBy:   "system",
			UpdatedBy:   "system",
			FieldDefinitions: []models.FieldDefinition{
				{
					FieldName:       "id",
					DisplayName:     "团队ID",
					DataType:        "string",
					Description:     "团队唯一标识",
					IsRequired:      true,
					IsUnique:        true,
					IsSystemField:   true,
					SystemFieldType: "primary_key",
					DisplayOrder:    1,
				},
				{
					FieldName:    "name",
					DisplayName:  "团队名称",
					DataType:     "string",
					Description:  "团队名称",
					IsRequired:   true,
					MaxLength:    100,
					DisplayOrder: 2,
				},
				{
					FieldName:    "description",
					DisplayName:  "团队描述",
					DataType:     "text",
					Description:  "团队详细描述",
					IsRequired:   false,
					DisplayOrder: 3,
				},
				{
					FieldName:    "status",
					DisplayName:  "状态",
					DataType:     "string",
					Description:  "团队状态",
					IsRequired:   true,
					DefaultValue: "active",
					DisplayOrder: 4,
				},
				{
					FieldName:       "created_at",
					DisplayName:     "创建时间",
					DataType:        "date",
					Description:     "团队创建时间",
					IsRequired:      true,
					IsSystemField:   true,
					SystemFieldType: "created_at",
					DisplayOrder:    5,
				},
			},
		},
		{
			Name:        "qrcode_form",
			DisplayName: "二维码表单",
			Description: "二维码表单实体模板",
			Type:        "form",
			Source:      "system",
			Module:      "qrcode",
			Status:      "active",
			CreatedBy:   "system",
			UpdatedBy:   "system",
			FieldDefinitions: []models.FieldDefinition{
				{
					FieldName:       "id",
					DisplayName:     "表单ID",
					DataType:        "string",
					Description:     "表单唯一标识",
					IsRequired:      true,
					IsUnique:        true,
					IsSystemField:   true,
					SystemFieldType: "primary_key",
					DisplayOrder:    1,
				},
				{
					FieldName:    "form_name",
					DisplayName:  "表单名称",
					DataType:     "string",
					Description:  "表单名称",
					IsRequired:   true,
					MaxLength:    100,
					DisplayOrder: 2,
				},
				{
					FieldName:    "form_type",
					DisplayName:  "表单类型",
					DataType:     "string",
					Description:  "表单类型",
					IsRequired:   true,
					DisplayOrder: 3,
				},
				{
					FieldName:    "qr_content",
					DisplayName:  "二维码内容",
					DataType:     "string",
					Description:  "二维码包含的内容",
					IsRequired:   true,
					DisplayOrder: 4,
				},
				{
					FieldName:    "scan_count",
					DisplayName:  "扫描次数",
					DataType:     "number",
					Description:  "二维码被扫描的次数",
					IsRequired:   false,
					DefaultValue: "0",
					DisplayOrder: 5,
				},
			},
		},
	}

	// 检查并创建系统实体
	for _, entity := range systemEntities {
		var existingEntity models.DataEntity
		err := db.Where("name = ? AND source = ?", entity.Name, "system").First(&existingEntity).Error
		if err != nil {
			if err == gorm.ErrRecordNotFound {
				// 实体不存在，创建它
				if err := db.Create(entity).Error; err != nil {
					return fmt.Errorf("failed to create system entity %s: %w", entity.Name, err)
				}
			} else {
				return fmt.Errorf("failed to check system entity %s: %w", entity.Name, err)
			}
		}
	}

	return nil
}
