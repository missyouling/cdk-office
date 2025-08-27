/*
 * MIT License
 *
 * Copyright (c) 2025 CDK-Office
 */

package dictionary

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"gorm.io/datatypes"
	"gorm.io/gorm"

	"cdk-office/internal/models"
)

// Service 数据字典服务
type Service struct {
	db        *gorm.DB
	oddClient *ODDClient
}

// NewService 创建数据字典服务
func NewService(db *gorm.DB, oddClient *ODDClient) *Service {
	return &Service{
		db:        db,
		oddClient: oddClient,
	}
}

// CreateEntity 创建数据实体
func (s *Service) CreateEntity(ctx context.Context, userID, teamID string, entity *models.DataEntity) error {
	// 设置基本信息
	entity.TeamID = teamID
	entity.CreatedBy = userID
	entity.UpdatedBy = userID
	entity.Status = "active"

	// 验证实体名称唯一性
	var existingEntity models.DataEntity
	err := s.db.Where("name = ? AND team_id = ?", entity.Name, teamID).First(&existingEntity).Error
	if err == nil {
		return errors.New("实体名称已存在")
	}
	if !errors.Is(err, gorm.ErrRecordNotFound) {
		return fmt.Errorf("检查实体名称失败: %w", err)
	}

	// 开始事务
	tx := s.db.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	// 创建实体
	if err := tx.Create(entity).Error; err != nil {
		tx.Rollback()
		return fmt.Errorf("创建实体失败: %w", err)
	}

	// 同步到ODD平台
	if s.oddClient != nil {
		if err := s.oddClient.CreateDataEntity(ctx, entity); err != nil {
			// ODD同步失败不回滚，但记录日志
			// 在生产环境中应该记录到日志系统
			fmt.Printf("Warning: Failed to sync entity to ODD: %v\n", err)
		}
	}

	// 记录变更日志
	changeLog := &models.DictionaryChangeLog{
		EntityID:          entity.ID,
		ChangeType:        "create",
		ChangeScope:       "entity",
		NewValue:          s.marshalToJSON(entity),
		ChangeReason:      "创建新的数据实体",
		ChangeDescription: fmt.Sprintf("创建实体: %s", entity.Name),
		RequestedBy:       userID,
		TeamID:            teamID,
		ApprovalStatus:    "approved", // 创建操作自动审批
		ApprovedBy:        userID,
		ExecutionStatus:   "executed",
	}
	now := time.Now()
	changeLog.ApprovedAt = &now
	changeLog.ExecutedAt = &now

	if err := tx.Create(changeLog).Error; err != nil {
		tx.Rollback()
		return fmt.Errorf("记录变更日志失败: %w", err)
	}

	return tx.Commit().Error
}

// GetEntity 获取数据实体
func (s *Service) GetEntity(ctx context.Context, entityID, teamID string) (*models.DataEntity, error) {
	var entity models.DataEntity
	err := s.db.Preload("FieldDefinitions.FieldExtensions").
		Where("id = ? AND team_id = ?", entityID, teamID).
		First(&entity).Error

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("数据实体不存在")
		}
		return nil, fmt.Errorf("获取数据实体失败: %w", err)
	}

	return &entity, nil
}

// ListEntities 获取数据实体列表
func (s *Service) ListEntities(ctx context.Context, teamID string, filter EntityFilter) ([]*models.DataEntity, int64, error) {
	query := s.db.Model(&models.DataEntity{}).Where("team_id = ?", teamID)

	// 应用过滤条件
	if filter.Type != "" {
		query = query.Where("type = ?", filter.Type)
	}
	if filter.Module != "" {
		query = query.Where("module = ?", filter.Module)
	}
	if filter.Status != "" {
		query = query.Where("status = ?", filter.Status)
	}
	if filter.Search != "" {
		searchPattern := "%" + filter.Search + "%"
		query = query.Where("name ILIKE ? OR display_name ILIKE ? OR description ILIKE ?",
			searchPattern, searchPattern, searchPattern)
	}

	// 获取总数
	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, fmt.Errorf("获取实体总数失败: %w", err)
	}

	// 应用分页和排序
	if filter.PageSize > 0 {
		offset := (filter.Page - 1) * filter.PageSize
		query = query.Offset(offset).Limit(filter.PageSize)
	}

	sortField := filter.SortBy
	if sortField == "" {
		sortField = "created_at"
	}
	sortOrder := filter.SortOrder
	if sortOrder == "" {
		sortOrder = "DESC"
	}
	query = query.Order(fmt.Sprintf("%s %s", sortField, sortOrder))

	// 预加载关联数据
	query = query.Preload("FieldDefinitions")

	var entities []*models.DataEntity
	if err := query.Find(&entities).Error; err != nil {
		return nil, 0, fmt.Errorf("获取实体列表失败: %w", err)
	}

	return entities, total, nil
}

// UpdateEntity 更新数据实体
func (s *Service) UpdateEntity(ctx context.Context, userID, teamID, entityID string, updates map[string]interface{}) error {
	// 获取原实体
	originalEntity, err := s.GetEntity(ctx, entityID, teamID)
	if err != nil {
		return err
	}

	// 检查权限（可以在这里添加更详细的权限检查）
	if originalEntity.CreatedBy != userID {
		// 这里可以检查用户是否有管理员权限
	}

	// 开始事务
	tx := s.db.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	// 更新实体
	updates["updated_by"] = userID
	updates["updated_at"] = time.Now()

	if err := tx.Model(&models.DataEntity{}).Where("id = ? AND team_id = ?", entityID, teamID).
		Updates(updates).Error; err != nil {
		tx.Rollback()
		return fmt.Errorf("更新实体失败: %w", err)
	}

	// 获取更新后的实体
	var updatedEntity models.DataEntity
	if err := tx.Where("id = ?", entityID).First(&updatedEntity).Error; err != nil {
		tx.Rollback()
		return fmt.Errorf("获取更新后实体失败: %w", err)
	}

	// 记录变更日志
	changeLog := &models.DictionaryChangeLog{
		EntityID:          entityID,
		ChangeType:        "update",
		ChangeScope:       "entity",
		OldValue:          s.marshalToJSON(originalEntity),
		NewValue:          s.marshalToJSON(&updatedEntity),
		ChangeReason:      "更新数据实体",
		ChangeDescription: fmt.Sprintf("更新实体: %s", updatedEntity.Name),
		RequestedBy:       userID,
		TeamID:            teamID,
		ApprovalStatus:    "approved", // 简单更新自动审批
		ApprovedBy:        userID,
		ExecutionStatus:   "executed",
	}
	now := time.Now()
	changeLog.ApprovedAt = &now
	changeLog.ExecutedAt = &now

	if err := tx.Create(changeLog).Error; err != nil {
		tx.Rollback()
		return fmt.Errorf("记录变更日志失败: %w", err)
	}

	return tx.Commit().Error
}

// DeleteEntity 删除数据实体
func (s *Service) DeleteEntity(ctx context.Context, userID, teamID, entityID string) error {
	// 获取实体
	entity, err := s.GetEntity(ctx, entityID, teamID)
	if err != nil {
		return err
	}

	// 检查是否为系统实体
	if entity.Source == "system" {
		return errors.New("系统实体不能删除")
	}

	// 开始事务
	tx := s.db.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	// 软删除实体（将状态设为inactive）
	if err := tx.Model(&models.DataEntity{}).Where("id = ? AND team_id = ?", entityID, teamID).
		Updates(map[string]interface{}{
			"status":     "inactive",
			"updated_by": userID,
			"updated_at": time.Now(),
		}).Error; err != nil {
		tx.Rollback()
		return fmt.Errorf("删除实体失败: %w", err)
	}

	// 记录变更日志
	changeLog := &models.DictionaryChangeLog{
		EntityID:          entityID,
		ChangeType:        "delete",
		ChangeScope:       "entity",
		OldValue:          s.marshalToJSON(entity),
		ChangeReason:      "删除数据实体",
		ChangeDescription: fmt.Sprintf("删除实体: %s", entity.Name),
		RequestedBy:       userID,
		TeamID:            teamID,
		ApprovalStatus:    "approved",
		ApprovedBy:        userID,
		ExecutionStatus:   "executed",
	}
	now := time.Now()
	changeLog.ApprovedAt = &now
	changeLog.ExecutedAt = &now

	if err := tx.Create(changeLog).Error; err != nil {
		tx.Rollback()
		return fmt.Errorf("记录变更日志失败: %w", err)
	}

	return tx.Commit().Error
}

// CreateField 创建字段定义
func (s *Service) CreateField(ctx context.Context, userID, teamID string, field *models.FieldDefinition) error {
	// 验证实体存在
	entity, err := s.GetEntity(ctx, field.EntityID, teamID)
	if err != nil {
		return err
	}

	// 验证字段名称唯一性
	var existingField models.FieldDefinition
	err = s.db.Where("entity_id = ? AND field_name = ?", field.EntityID, field.FieldName).
		First(&existingField).Error
	if err == nil {
		return errors.New("字段名称已存在")
	}
	if !errors.Is(err, gorm.ErrRecordNotFound) {
		return fmt.Errorf("检查字段名称失败: %w", err)
	}

	// 设置基本信息
	field.Status = "active"
	field.Version = 1

	// 开始事务
	tx := s.db.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	// 创建字段
	if err := tx.Create(field).Error; err != nil {
		tx.Rollback()
		return fmt.Errorf("创建字段失败: %w", err)
	}

	// 同步到ODD平台
	if s.oddClient != nil {
		if err := s.oddClient.AddFieldDefinition(ctx, field); err != nil {
			fmt.Printf("Warning: Failed to sync field to ODD: %v\n", err)
		}
	}

	// 记录变更日志
	changeLog := &models.DictionaryChangeLog{
		EntityID:          field.EntityID,
		FieldID:           field.ID,
		ChangeType:        "create",
		ChangeScope:       "field",
		NewValue:          s.marshalToJSON(field),
		ChangeReason:      "创建新字段",
		ChangeDescription: fmt.Sprintf("在实体 %s 中创建字段: %s", entity.Name, field.FieldName),
		RequestedBy:       userID,
		TeamID:            teamID,
		ApprovalStatus:    "approved",
		ApprovedBy:        userID,
		ExecutionStatus:   "executed",
	}
	now := time.Now()
	changeLog.ApprovedAt = &now
	changeLog.ExecutedAt = &now

	if err := tx.Create(changeLog).Error; err != nil {
		tx.Rollback()
		return fmt.Errorf("记录变更日志失败: %w", err)
	}

	return tx.Commit().Error
}

// GetFieldsByEntity 获取实体的字段列表
func (s *Service) GetFieldsByEntity(ctx context.Context, entityID, teamID string) ([]*models.FieldDefinition, error) {
	// 验证实体存在
	if _, err := s.GetEntity(ctx, entityID, teamID); err != nil {
		return nil, err
	}

	var fields []*models.FieldDefinition
	err := s.db.Preload("FieldExtensions").
		Where("entity_id = ? AND status = ?", entityID, "active").
		Order("display_order ASC, created_at ASC").
		Find(&fields).Error

	if err != nil {
		return nil, fmt.Errorf("获取字段列表失败: %w", err)
	}

	return fields, nil
}

// UpdateField 更新字段定义
func (s *Service) UpdateField(ctx context.Context, userID, teamID, fieldID string, updates map[string]interface{}) error {
	// 获取原字段
	var originalField models.FieldDefinition
	err := s.db.Preload("FieldExtensions").Where("id = ?", fieldID).First(&originalField).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errors.New("字段不存在")
		}
		return fmt.Errorf("获取字段失败: %w", err)
	}

	// 验证实体权限
	if _, err := s.GetEntity(ctx, originalField.EntityID, teamID); err != nil {
		return err
	}

	// 开始事务
	tx := s.db.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	// 更新字段版本
	updates["version"] = originalField.Version + 1
	updates["updated_at"] = time.Now()

	if err := tx.Model(&models.FieldDefinition{}).Where("id = ?", fieldID).
		Updates(updates).Error; err != nil {
		tx.Rollback()
		return fmt.Errorf("更新字段失败: %w", err)
	}

	// 获取更新后的字段
	var updatedField models.FieldDefinition
	if err := tx.Preload("FieldExtensions").Where("id = ?", fieldID).First(&updatedField).Error; err != nil {
		tx.Rollback()
		return fmt.Errorf("获取更新后字段失败: %w", err)
	}

	// 记录变更日志
	changeLog := &models.DictionaryChangeLog{
		EntityID:          originalField.EntityID,
		FieldID:           fieldID,
		ChangeType:        "update",
		ChangeScope:       "field",
		OldValue:          s.marshalToJSON(&originalField),
		NewValue:          s.marshalToJSON(&updatedField),
		ChangeReason:      "更新字段定义",
		ChangeDescription: fmt.Sprintf("更新字段: %s", updatedField.FieldName),
		RequestedBy:       userID,
		TeamID:            teamID,
		ApprovalStatus:    "approved",
		ApprovedBy:        userID,
		ExecutionStatus:   "executed",
	}
	now := time.Now()
	changeLog.ApprovedAt = &now
	changeLog.ExecutedAt = &now

	if err := tx.Create(changeLog).Error; err != nil {
		tx.Rollback()
		return fmt.Errorf("记录变更日志失败: %w", err)
	}

	return tx.Commit().Error
}

// DeleteField 删除字段定义
func (s *Service) DeleteField(ctx context.Context, userID, teamID, fieldID string) error {
	// 获取字段
	var field models.FieldDefinition
	err := s.db.Where("id = ?", fieldID).First(&field).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errors.New("字段不存在")
		}
		return fmt.Errorf("获取字段失败: %w", err)
	}

	// 验证实体权限
	if _, err := s.GetEntity(ctx, field.EntityID, teamID); err != nil {
		return err
	}

	// 检查是否为系统字段
	if field.IsSystemField {
		return errors.New("系统字段不能删除")
	}

	// 开始事务
	tx := s.db.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	// 软删除字段
	if err := tx.Model(&models.FieldDefinition{}).Where("id = ?", fieldID).
		Updates(map[string]interface{}{
			"status":     "inactive",
			"updated_at": time.Now(),
		}).Error; err != nil {
		tx.Rollback()
		return fmt.Errorf("删除字段失败: %w", err)
	}

	// 记录变更日志
	changeLog := &models.DictionaryChangeLog{
		EntityID:          field.EntityID,
		FieldID:           fieldID,
		ChangeType:        "delete",
		ChangeScope:       "field",
		OldValue:          s.marshalToJSON(&field),
		ChangeReason:      "删除字段",
		ChangeDescription: fmt.Sprintf("删除字段: %s", field.FieldName),
		RequestedBy:       userID,
		TeamID:            teamID,
		ApprovalStatus:    "approved",
		ApprovedBy:        userID,
		ExecutionStatus:   "executed",
	}
	now := time.Now()
	changeLog.ApprovedAt = &now
	changeLog.ExecutedAt = &now

	if err := tx.Create(changeLog).Error; err != nil {
		tx.Rollback()
		return fmt.Errorf("记录变更日志失败: %w", err)
	}

	return tx.Commit().Error
}

// marshalToJSON 将对象序列化为JSON
func (s *Service) marshalToJSON(data interface{}) datatypes.JSON {
	jsonData, _ := json.Marshal(data)
	return datatypes.JSON(jsonData)
}

// EntityFilter 实体过滤器
type EntityFilter struct {
	Type      string `json:"type"`
	Module    string `json:"module"`
	Status    string `json:"status"`
	Search    string `json:"search"`
	Page      int    `json:"page"`
	PageSize  int    `json:"page_size"`
	SortBy    string `json:"sort_by"`
	SortOrder string `json:"sort_order"`
}

// ValidateField 验证字段定义
func (s *Service) ValidateField(field *models.FieldDefinition) error {
	if field.FieldName == "" {
		return errors.New("字段名称不能为空")
	}

	if field.DataType == "" {
		return errors.New("数据类型不能为空")
	}

	// 验证数据类型
	validTypes := []string{"string", "number", "boolean", "date", "json", "array", "text", "email", "url", "phone"}
	found := false
	for _, validType := range validTypes {
		if field.DataType == validType {
			found = true
			break
		}
	}
	if !found {
		return fmt.Errorf("无效的数据类型: %s", field.DataType)
	}

	// 验证字段名称格式（只允许字母、数字、下划线）
	if !isValidFieldName(field.FieldName) {
		return errors.New("字段名称只能包含字母、数字和下划线，且不能以数字开头")
	}

	return nil
}

// isValidFieldName 验证字段名称格式
func isValidFieldName(name string) bool {
	if len(name) == 0 {
		return false
	}

	// 不能以数字开头
	if name[0] >= '0' && name[0] <= '9' {
		return false
	}

	// 只能包含字母、数字、下划线
	for _, char := range name {
		if !((char >= 'a' && char <= 'z') ||
			(char >= 'A' && char <= 'Z') ||
			(char >= '0' && char <= '9') ||
			char == '_') {
			return false
		}
	}

	return true
}
