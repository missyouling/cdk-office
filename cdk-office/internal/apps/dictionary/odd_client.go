/*
 * MIT License
 *
 * Copyright (c) 2025 CDK-Office
 */

package dictionary

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"cdk-office/internal/models"
)

// ODDClient ODD平台客户端
type ODDClient struct {
	BaseURL    string
	APIKey     string
	HTTPClient *http.Client
}

// ODDEntityRequest ODD实体请求
type ODDEntityRequest struct {
	Name        string                 `json:"name"`
	Type        string                 `json:"type"`
	Description string                 `json:"description"`
	Metadata    map[string]interface{} `json:"metadata"`
	Owner       string                 `json:"owner"`
}

// ODDFieldRequest ODD字段请求
type ODDFieldRequest struct {
	EntityID    string                 `json:"entity_id"`
	Name        string                 `json:"name"`
	Type        string                 `json:"type"`
	Description string                 `json:"description"`
	Required    bool                   `json:"required"`
	Metadata    map[string]interface{} `json:"metadata"`
}

// ODDResponse ODD响应
type ODDResponse struct {
	Success bool                   `json:"success"`
	Data    map[string]interface{} `json:"data"`
	Error   string                 `json:"error,omitempty"`
}

// NewODDClient 创建ODD客户端
func NewODDClient(baseURL, apiKey string) *ODDClient {
	return &ODDClient{
		BaseURL: baseURL,
		APIKey:  apiKey,
		HTTPClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// CreateDataEntity 创建数据实体
func (c *ODDClient) CreateDataEntity(ctx context.Context, entity *models.DataEntity) error {
	url := fmt.Sprintf("%s/api/entities", c.BaseURL)

	request := ODDEntityRequest{
		Name:        entity.Name,
		Type:        entity.Type,
		Description: entity.Description,
		Metadata: map[string]interface{}{
			"display_name": entity.DisplayName,
			"source":       entity.Source,
			"module":       entity.Module,
			"team_id":      entity.TeamID,
			"created_by":   entity.CreatedBy,
		},
		Owner: entity.CreatedBy,
	}

	payload, err := json.Marshal(request)
	if err != nil {
		return fmt.Errorf("failed to marshal request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(payload))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+c.APIKey)
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		return fmt.Errorf("ODD API returned status: %d", resp.StatusCode)
	}

	var oddResp ODDResponse
	if err := json.NewDecoder(resp.Body).Decode(&oddResp); err != nil {
		return fmt.Errorf("failed to decode response: %w", err)
	}

	if !oddResp.Success {
		return fmt.Errorf("ODD API error: %s", oddResp.Error)
	}

	return nil
}

// AddFieldDefinition 添加字段定义
func (c *ODDClient) AddFieldDefinition(ctx context.Context, field *models.FieldDefinition) error {
	url := fmt.Sprintf("%s/api/fields", c.BaseURL)

	request := ODDFieldRequest{
		EntityID:    field.EntityID,
		Name:        field.FieldName,
		Type:        field.DataType,
		Description: field.Description,
		Required:    field.IsRequired,
		Metadata: map[string]interface{}{
			"display_name":    field.DisplayName,
			"default_value":   field.DefaultValue,
			"validation_rule": field.ValidationRule,
			"display_order":   field.DisplayOrder,
			"is_system_field": field.IsSystemField,
			"is_visible":      field.IsVisible,
			"is_editable":     field.IsEditable,
			"is_searchable":   field.IsSearchable,
		},
	}

	payload, err := json.Marshal(request)
	if err != nil {
		return fmt.Errorf("failed to marshal request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(payload))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+c.APIKey)
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		return fmt.Errorf("ODD API returned status: %d", resp.StatusCode)
	}

	return nil
}

// GetFieldDefinitions 获取实体的字段定义
func (c *ODDClient) GetFieldDefinitions(ctx context.Context, entityID string) ([]models.FieldDefinition, error) {
	url := fmt.Sprintf("%s/api/entities/%s/fields", c.BaseURL, entityID)

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+c.APIKey)

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("ODD API returned status: %d", resp.StatusCode)
	}

	var oddResp ODDResponse
	if err := json.NewDecoder(resp.Body).Decode(&oddResp); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	if !oddResp.Success {
		return nil, fmt.Errorf("ODD API error: %s", oddResp.Error)
	}

	// 这里需要根据实际的ODD API响应格式来解析字段定义
	// 暂时返回空列表，实际使用时需要根据ODD平台的具体API来实现
	var fields []models.FieldDefinition

	return fields, nil
}

// UpdateFieldDefinition 更新字段定义
func (c *ODDClient) UpdateFieldDefinition(ctx context.Context, fieldID string, field *models.FieldDefinition) error {
	url := fmt.Sprintf("%s/api/fields/%s", c.BaseURL, fieldID)

	request := ODDFieldRequest{
		EntityID:    field.EntityID,
		Name:        field.FieldName,
		Type:        field.DataType,
		Description: field.Description,
		Required:    field.IsRequired,
		Metadata: map[string]interface{}{
			"display_name":    field.DisplayName,
			"default_value":   field.DefaultValue,
			"validation_rule": field.ValidationRule,
			"display_order":   field.DisplayOrder,
			"is_system_field": field.IsSystemField,
			"is_visible":      field.IsVisible,
			"is_editable":     field.IsEditable,
			"is_searchable":   field.IsSearchable,
		},
	}

	payload, err := json.Marshal(request)
	if err != nil {
		return fmt.Errorf("failed to marshal request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "PUT", url, bytes.NewBuffer(payload))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+c.APIKey)
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("ODD API returned status: %d", resp.StatusCode)
	}

	return nil
}

// DeleteEntity 删除数据实体
func (c *ODDClient) DeleteEntity(ctx context.Context, entityID string) error {
	url := fmt.Sprintf("%s/api/entities/%s", c.BaseURL, entityID)

	req, err := http.NewRequestWithContext(ctx, "DELETE", url, nil)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+c.APIKey)

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusNoContent {
		return fmt.Errorf("ODD API returned status: %d", resp.StatusCode)
	}

	return nil
}

// ValidateConnection 验证ODD连接
func (c *ODDClient) ValidateConnection(ctx context.Context) error {
	url := fmt.Sprintf("%s/api/health", c.BaseURL)

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+c.APIKey)

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("ODD API health check failed with status: %d", resp.StatusCode)
	}

	return nil
}

// SyncEntity 同步实体到ODD平台
func (c *ODDClient) SyncEntity(ctx context.Context, entity *models.DataEntity) error {
	// 先创建实体
	if err := c.CreateDataEntity(ctx, entity); err != nil {
		return fmt.Errorf("failed to create entity: %w", err)
	}

	// 再同步字段定义
	for _, field := range entity.FieldDefinitions {
		if err := c.AddFieldDefinition(ctx, &field); err != nil {
			return fmt.Errorf("failed to add field %s: %w", field.FieldName, err)
		}
	}

	return nil
}

// BatchSyncEntities 批量同步实体
func (c *ODDClient) BatchSyncEntities(ctx context.Context, entities []*models.DataEntity) error {
	for _, entity := range entities {
		if err := c.SyncEntity(ctx, entity); err != nil {
			return fmt.Errorf("failed to sync entity %s: %w", entity.Name, err)
		}
	}

	return nil
}
