/*
 * MIT License
 *
 * Copyright (c) 2025 CDK-Office
 */

package dictionary

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"gorm.io/datatypes"

	"cdk-office/internal/models"
	"cdk-office/internal/utils"
)

// Handler 数据字典处理器
type Handler struct {
	service *Service
}

// NewHandler 创建数据字典处理器
func NewHandler(service *Service) *Handler {
	return &Handler{
		service: service,
	}
}

// CreateEntityRequest 创建实体请求
type CreateEntityRequest struct {
	Name        string                 `json:"name" binding:"required"`
	DisplayName string                 `json:"display_name"`
	Description string                 `json:"description"`
	Type        string                 `json:"type" binding:"required"`
	Source      string                 `json:"source"`
	Module      string                 `json:"module"`
	Metadata    map[string]interface{} `json:"metadata"`
}

// CreateFieldRequest 创建字段请求
type CreateFieldRequest struct {
	EntityID       string                 `json:"entity_id" binding:"required"`
	FieldName      string                 `json:"field_name" binding:"required"`
	DisplayName    string                 `json:"display_name"`
	DataType       string                 `json:"data_type" binding:"required"`
	Description    string                 `json:"description"`
	IsRequired     bool                   `json:"is_required"`
	IsUnique       bool                   `json:"is_unique"`
	DefaultValue   string                 `json:"default_value"`
	ValidationRule string                 `json:"validation_rule"`
	MinLength      int                    `json:"min_length"`
	MaxLength      int                    `json:"max_length"`
	MinValue       string                 `json:"min_value"`
	MaxValue       string                 `json:"max_value"`
	Options        map[string]interface{} `json:"options"`
	OptionsSource  string                 `json:"options_source"`
	DisplayOrder   int                    `json:"display_order"`
	IsVisible      bool                   `json:"is_visible"`
	IsEditable     bool                   `json:"is_editable"`
	IsSearchable   bool                   `json:"is_searchable"`
}

// CreateEntity 创建数据实体
func (h *Handler) CreateEntity(c *gin.Context) {
	var req CreateEntityRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "请求参数错误", err.Error())
		return
	}

	userID := c.GetString("user_id")
	teamID := c.GetString("team_id")

	if userID == "" || teamID == "" {
		utils.ErrorResponse(c, http.StatusUnauthorized, "用户信息不完整", "")
		return
	}

	// 创建实体对象
	entity := &models.DataEntity{
		Name:        req.Name,
		DisplayName: req.DisplayName,
		Description: req.Description,
		Type:        req.Type,
		Source:      req.Source,
		Module:      req.Module,
	}

	if req.Source == "" {
		entity.Source = "user_defined"
	}

	if req.Metadata != nil {
		metadata, _ := datatypes.JSON(nil).MarshalJSON()
		entity.Metadata = metadata
	}

	// 创建实体
	if err := h.service.CreateEntity(c.Request.Context(), userID, teamID, entity); err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "创建实体失败", err.Error())
		return
	}

	utils.SuccessResponse(c, "创建实体成功", gin.H{
		"entity_id": entity.ID,
		"name":      entity.Name,
	})
}

// GetEntity 获取数据实体
func (h *Handler) GetEntity(c *gin.Context) {
	entityID := c.Param("id")
	teamID := c.GetString("team_id")

	if teamID == "" {
		utils.ErrorResponse(c, http.StatusUnauthorized, "团队信息缺失", "")
		return
	}

	entity, err := h.service.GetEntity(c.Request.Context(), entityID, teamID)
	if err != nil {
		utils.ErrorResponse(c, http.StatusNotFound, "获取实体失败", err.Error())
		return
	}

	utils.SuccessResponse(c, "获取实体成功", entity)
}

// ListEntities 获取数据实体列表
func (h *Handler) ListEntities(c *gin.Context) {
	teamID := c.GetString("team_id")

	if teamID == "" {
		utils.ErrorResponse(c, http.StatusUnauthorized, "团队信息缺失", "")
		return
	}

	// 解析查询参数
	filter := EntityFilter{
		Type:      c.Query("type"),
		Module:    c.Query("module"),
		Status:    c.Query("status"),
		Search:    c.Query("search"),
		SortBy:    c.Query("sort_by"),
		SortOrder: c.Query("sort_order"),
		Page:      1,
		PageSize:  20,
	}

	if page := c.Query("page"); page != "" {
		if p, err := strconv.Atoi(page); err == nil && p > 0 {
			filter.Page = p
		}
	}

	if pageSize := c.Query("page_size"); pageSize != "" {
		if ps, err := strconv.Atoi(pageSize); err == nil && ps > 0 && ps <= 100 {
			filter.PageSize = ps
		}
	}

	entities, total, err := h.service.ListEntities(c.Request.Context(), teamID, filter)
	if err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "获取实体列表失败", err.Error())
		return
	}

	utils.SuccessResponse(c, "获取实体列表成功", gin.H{
		"entities":  entities,
		"total":     total,
		"page":      filter.Page,
		"page_size": filter.PageSize,
	})
}

// UpdateEntity 更新数据实体
func (h *Handler) UpdateEntity(c *gin.Context) {
	entityID := c.Param("id")
	userID := c.GetString("user_id")
	teamID := c.GetString("team_id")

	if userID == "" || teamID == "" {
		utils.ErrorResponse(c, http.StatusUnauthorized, "用户信息不完整", "")
		return
	}

	var updates map[string]interface{}
	if err := c.ShouldBindJSON(&updates); err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "请求参数错误", err.Error())
		return
	}

	// 过滤不允许更新的字段
	allowedFields := []string{"display_name", "description", "metadata"}
	filteredUpdates := make(map[string]interface{})
	for _, field := range allowedFields {
		if value, exists := updates[field]; exists {
			filteredUpdates[field] = value
		}
	}

	if len(filteredUpdates) == 0 {
		utils.ErrorResponse(c, http.StatusBadRequest, "没有有效的更新字段", "")
		return
	}

	if err := h.service.UpdateEntity(c.Request.Context(), userID, teamID, entityID, filteredUpdates); err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "更新实体失败", err.Error())
		return
	}

	utils.SuccessResponse(c, "更新实体成功", nil)
}

// DeleteEntity 删除数据实体
func (h *Handler) DeleteEntity(c *gin.Context) {
	entityID := c.Param("id")
	userID := c.GetString("user_id")
	teamID := c.GetString("team_id")

	if userID == "" || teamID == "" {
		utils.ErrorResponse(c, http.StatusUnauthorized, "用户信息不完整", "")
		return
	}

	if err := h.service.DeleteEntity(c.Request.Context(), userID, teamID, entityID); err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "删除实体失败", err.Error())
		return
	}

	utils.SuccessResponse(c, "删除实体成功", nil)
}

// CreateField 创建字段定义
func (h *Handler) CreateField(c *gin.Context) {
	var req CreateFieldRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "请求参数错误", err.Error())
		return
	}

	userID := c.GetString("user_id")
	teamID := c.GetString("team_id")

	if userID == "" || teamID == "" {
		utils.ErrorResponse(c, http.StatusUnauthorized, "用户信息不完整", "")
		return
	}

	// 创建字段对象
	field := &models.FieldDefinition{
		EntityID:       req.EntityID,
		FieldName:      req.FieldName,
		DisplayName:    req.DisplayName,
		DataType:       req.DataType,
		Description:    req.Description,
		IsRequired:     req.IsRequired,
		IsUnique:       req.IsUnique,
		DefaultValue:   req.DefaultValue,
		ValidationRule: req.ValidationRule,
		MinLength:      req.MinLength,
		MaxLength:      req.MaxLength,
		MinValue:       req.MinValue,
		MaxValue:       req.MaxValue,
		OptionsSource:  req.OptionsSource,
		DisplayOrder:   req.DisplayOrder,
		IsVisible:      req.IsVisible,
		IsEditable:     req.IsEditable,
		IsSearchable:   req.IsSearchable,
		IsSystemField:  false,
	}

	// 设置默认值
	if field.DisplayName == "" {
		field.DisplayName = field.FieldName
	}

	if req.Options != nil {
		optionsJSON, _ := datatypes.JSON(nil).MarshalJSON()
		field.Options = optionsJSON
	}

	// 验证字段
	if err := h.service.ValidateField(field); err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "字段验证失败", err.Error())
		return
	}

	// 创建字段
	if err := h.service.CreateField(c.Request.Context(), userID, teamID, field); err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "创建字段失败", err.Error())
		return
	}

	utils.SuccessResponse(c, "创建字段成功", gin.H{
		"field_id":   field.ID,
		"field_name": field.FieldName,
	})
}

// GetFieldsByEntity 获取实体的字段列表
func (h *Handler) GetFieldsByEntity(c *gin.Context) {
	entityID := c.Param("entity_id")
	teamID := c.GetString("team_id")

	if teamID == "" {
		utils.ErrorResponse(c, http.StatusUnauthorized, "团队信息缺失", "")
		return
	}

	fields, err := h.service.GetFieldsByEntity(c.Request.Context(), entityID, teamID)
	if err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "获取字段列表失败", err.Error())
		return
	}

	utils.SuccessResponse(c, "获取字段列表成功", gin.H{
		"fields": fields,
	})
}

// UpdateField 更新字段定义
func (h *Handler) UpdateField(c *gin.Context) {
	fieldID := c.Param("field_id")
	userID := c.GetString("user_id")
	teamID := c.GetString("team_id")

	if userID == "" || teamID == "" {
		utils.ErrorResponse(c, http.StatusUnauthorized, "用户信息不完整", "")
		return
	}

	var updates map[string]interface{}
	if err := c.ShouldBindJSON(&updates); err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "请求参数错误", err.Error())
		return
	}

	// 过滤不允许更新的字段
	allowedFields := []string{
		"display_name", "description", "is_required", "is_unique",
		"default_value", "validation_rule", "min_length", "max_length",
		"min_value", "max_value", "options", "options_source",
		"display_order", "is_visible", "is_editable", "is_searchable",
	}

	filteredUpdates := make(map[string]interface{})
	for _, field := range allowedFields {
		if value, exists := updates[field]; exists {
			filteredUpdates[field] = value
		}
	}

	if len(filteredUpdates) == 0 {
		utils.ErrorResponse(c, http.StatusBadRequest, "没有有效的更新字段", "")
		return
	}

	if err := h.service.UpdateField(c.Request.Context(), userID, teamID, fieldID, filteredUpdates); err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "更新字段失败", err.Error())
		return
	}

	utils.SuccessResponse(c, "更新字段成功", nil)
}

// DeleteField 删除字段定义
func (h *Handler) DeleteField(c *gin.Context) {
	fieldID := c.Param("field_id")
	userID := c.GetString("user_id")
	teamID := c.GetString("team_id")

	if userID == "" || teamID == "" {
		utils.ErrorResponse(c, http.StatusUnauthorized, "用户信息不完整", "")
		return
	}

	if err := h.service.DeleteField(c.Request.Context(), userID, teamID, fieldID); err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "删除字段失败", err.Error())
		return
	}

	utils.SuccessResponse(c, "删除字段成功", nil)
}

// GetDataTypes 获取支持的数据类型列表
func (h *Handler) GetDataTypes(c *gin.Context) {
	dataTypes := []map[string]interface{}{
		{
			"type":        "string",
			"name":        "字符串",
			"description": "文本数据",
			"options":     []string{"max_length", "min_length", "pattern"},
		},
		{
			"type":        "text",
			"name":        "长文本",
			"description": "多行文本数据",
			"options":     []string{"max_length", "min_length"},
		},
		{
			"type":        "number",
			"name":        "数字",
			"description": "数值数据",
			"options":     []string{"min_value", "max_value", "decimal_places"},
		},
		{
			"type":        "boolean",
			"name":        "布尔值",
			"description": "真/假值",
			"options":     []string{},
		},
		{
			"type":        "date",
			"name":        "日期",
			"description": "日期时间数据",
			"options":     []string{"format", "min_date", "max_date"},
		},
		{
			"type":        "email",
			"name":        "邮箱",
			"description": "电子邮件地址",
			"options":     []string{"domain_restriction"},
		},
		{
			"type":        "url",
			"name":        "网址",
			"description": "URL链接",
			"options":     []string{"protocol"},
		},
		{
			"type":        "phone",
			"name":        "电话",
			"description": "电话号码",
			"options":     []string{"country_code", "format"},
		},
		{
			"type":        "json",
			"name":        "JSON",
			"description": "JSON格式数据",
			"options":     []string{"schema"},
		},
		{
			"type":        "array",
			"name":        "数组",
			"description": "数组类型数据",
			"options":     []string{"item_type", "min_items", "max_items"},
		},
	}

	utils.SuccessResponse(c, "获取数据类型成功", gin.H{
		"data_types": dataTypes,
	})
}

// GetFieldTemplates 获取字段模板
func (h *Handler) GetFieldTemplates(c *gin.Context) {
	entityType := c.Query("entity_type")
	module := c.Query("module")

	// 根据实体类型和模块返回预定义的字段模板
	templates := h.getFieldTemplatesByType(entityType, module)

	utils.SuccessResponse(c, "获取字段模板成功", gin.H{
		"templates": templates,
	})
}

// getFieldTemplatesByType 根据类型获取字段模板
func (h *Handler) getFieldTemplatesByType(entityType, module string) []map[string]interface{} {
	var templates []map[string]interface{}

	// 通用字段模板
	commonTemplates := []map[string]interface{}{
		{
			"name":            "id",
			"display_name":    "ID",
			"data_type":       "string",
			"description":     "唯一标识符",
			"is_required":     true,
			"is_unique":       true,
			"is_system_field": true,
		},
		{
			"name":         "name",
			"display_name": "名称",
			"data_type":    "string",
			"description":  "名称字段",
			"is_required":  true,
			"max_length":   100,
		},
		{
			"name":         "description",
			"display_name": "描述",
			"data_type":    "text",
			"description":  "详细描述",
			"is_required":  false,
		},
		{
			"name":            "created_at",
			"display_name":    "创建时间",
			"data_type":       "date",
			"description":     "记录创建时间",
			"is_system_field": true,
		},
		{
			"name":            "updated_at",
			"display_name":    "更新时间",
			"data_type":       "date",
			"description":     "记录更新时间",
			"is_system_field": true,
		},
	}

	templates = append(templates, commonTemplates...)

	// 根据具体类型添加专用模板
	switch entityType {
	case "qrcode_form":
		templates = append(templates, []map[string]interface{}{
			{
				"name":         "qr_code",
				"display_name": "二维码内容",
				"data_type":    "string",
				"description":  "二维码包含的内容",
				"is_required":  true,
			},
			{
				"name":          "scan_count",
				"display_name":  "扫描次数",
				"data_type":     "number",
				"description":   "二维码被扫描的次数",
				"default_value": "0",
			},
		}...)
	case "survey_form":
		templates = append(templates, []map[string]interface{}{
			{
				"name":          "response_count",
				"display_name":  "响应数量",
				"data_type":     "number",
				"description":   "问卷响应数量",
				"default_value": "0",
			},
			{
				"name":          "completion_rate",
				"display_name":  "完成率",
				"data_type":     "number",
				"description":   "问卷完成率",
				"default_value": "0",
			},
		}...)
	case "workflow_form":
		templates = append(templates, []map[string]interface{}{
			{
				"name":         "workflow_status",
				"display_name": "工作流状态",
				"data_type":    "string",
				"description":  "当前工作流状态",
				"options":      []string{"pending", "in_progress", "completed", "rejected"},
			},
			{
				"name":         "assignee",
				"display_name": "负责人",
				"data_type":    "string",
				"description":  "当前负责人",
			},
		}...)
	}

	return templates
}
