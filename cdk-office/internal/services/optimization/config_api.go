/*
 * MIT License
 *
 * Copyright (c) 2025 CDK-Office
 */

package optimization

import (
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
)

// ConfigAPI 配置管理API
type ConfigAPI struct {
	configManager *ConfigManager
}

// NewConfigAPI 创建配置API
func NewConfigAPI(configManager *ConfigManager) *ConfigAPI {
	return &ConfigAPI{
		configManager: configManager,
	}
}

// GetConfig 获取单个配置
func (api *ConfigAPI) GetConfig(c *gin.Context) {
	configID := c.Param("id")
	if configID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "配置ID不能为空"})
		return
	}

	value, err := api.configManager.Get(configID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	// 获取配置详情
	api.configManager.mutex.RLock()
	config := api.configManager.configs[configID]
	api.configManager.mutex.RUnlock()

	if config == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "配置不存在"})
		return
	}

	response := gin.H{
		"id":            config.ID,
		"category":      config.Category,
		"name":          config.Name,
		"value":         value,
		"default_value": config.DefaultValue,
		"description":   config.Description,
		"type":          config.Type,
		"constraints":   config.Constraints,
		"version":       config.Version,
		"updated_at":    config.UpdatedAt,
	}

	c.JSON(http.StatusOK, gin.H{"data": response})
}

// GetConfigs 获取配置列表
func (api *ConfigAPI) GetConfigs(c *gin.Context) {
	category := c.Query("category")

	var configs map[string]*Config
	if category != "" {
		configs = api.configManager.GetByCategory(category)
	} else {
		api.configManager.mutex.RLock()
		configs = make(map[string]*Config)
		for id, config := range api.configManager.configs {
			configs[id] = config
		}
		api.configManager.mutex.RUnlock()
	}

	result := make([]gin.H, 0, len(configs))
	for _, config := range configs {
		value, _ := api.configManager.Get(config.ID)
		result = append(result, gin.H{
			"id":            config.ID,
			"category":      config.Category,
			"name":          config.Name,
			"value":         value,
			"default_value": config.DefaultValue,
			"description":   config.Description,
			"type":          config.Type,
			"constraints":   config.Constraints,
			"version":       config.Version,
			"updated_at":    config.UpdatedAt,
		})
	}

	c.JSON(http.StatusOK, gin.H{
		"data":  result,
		"total": len(result),
	})
}

// UpdateConfig 更新配置
func (api *ConfigAPI) UpdateConfig(c *gin.Context) {
	configID := c.Param("id")
	if configID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "配置ID不能为空"})
		return
	}

	var request struct {
		Value interface{} `json:"value" binding:"required"`
	}

	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "请求参数错误: " + err.Error()})
		return
	}

	// 获取配置信息
	api.configManager.mutex.RLock()
	config := api.configManager.configs[configID]
	api.configManager.mutex.RUnlock()

	if config == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "配置不存在"})
		return
	}

	// 类型转换
	var value interface{}
	switch config.Type {
	case "int":
		if intValue, ok := request.Value.(float64); ok {
			value = int(intValue)
		} else if intStr, ok := request.Value.(string); ok {
			if intValue, err := strconv.Atoi(intStr); err == nil {
				value = intValue
			} else {
				c.JSON(http.StatusBadRequest, gin.H{"error": "无效的整数值"})
				return
			}
		} else {
			c.JSON(http.StatusBadRequest, gin.H{"error": "无效的整数值"})
			return
		}
	case "float":
		if floatValue, ok := request.Value.(float64); ok {
			value = floatValue
		} else if floatStr, ok := request.Value.(string); ok {
			if floatValue, err := strconv.ParseFloat(floatStr, 64); err == nil {
				value = floatValue
			} else {
				c.JSON(http.StatusBadRequest, gin.H{"error": "无效的浮点数值"})
				return
			}
		} else {
			c.JSON(http.StatusBadRequest, gin.H{"error": "无效的浮点数值"})
			return
		}
	case "bool":
		if boolValue, ok := request.Value.(bool); ok {
			value = boolValue
		} else if boolStr, ok := request.Value.(string); ok {
			if boolValue, err := strconv.ParseBool(boolStr); err == nil {
				value = boolValue
			} else {
				c.JSON(http.StatusBadRequest, gin.H{"error": "无效的布尔值"})
				return
			}
		} else {
			c.JSON(http.StatusBadRequest, gin.H{"error": "无效的布尔值"})
			return
		}
	default:
		value = request.Value
	}

	// 更新配置
	if err := api.configManager.Set(configID, value); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message":   "配置更新成功",
		"config_id": configID,
		"new_value": value,
	})
}

// ResetConfig 重置配置到默认值
func (api *ConfigAPI) ResetConfig(c *gin.Context) {
	configID := c.Param("id")
	if configID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "配置ID不能为空"})
		return
	}

	// 获取配置信息
	api.configManager.mutex.RLock()
	config := api.configManager.configs[configID]
	api.configManager.mutex.RUnlock()

	if config == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "配置不存在"})
		return
	}

	// 解析默认值
	defaultValue, err := api.configManager.parseValue(config.DefaultValue, config.Type)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "解析默认值失败: " + err.Error()})
		return
	}

	// 设置为默认值
	if err := api.configManager.Set(configID, defaultValue); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message":       "配置已重置为默认值",
		"config_id":     configID,
		"default_value": defaultValue,
	})
}

// GetCategories 获取配置分类列表
func (api *ConfigAPI) GetCategories(c *gin.Context) {
	categories := api.configManager.ListCategories()

	// 统计每个分类的配置数量
	categoryStats := make(map[string]int)
	api.configManager.mutex.RLock()
	for _, config := range api.configManager.configs {
		categoryStats[config.Category]++
	}
	api.configManager.mutex.RUnlock()

	result := make([]gin.H, 0, len(categories))
	for _, category := range categories {
		result = append(result, gin.H{
			"category": category,
			"count":    categoryStats[category],
		})
	}

	c.JSON(http.StatusOK, gin.H{
		"data":  result,
		"total": len(result),
	})
}

// GetAutoTuningStatus 获取自动调优状态
func (api *ConfigAPI) GetAutoTuningStatus(c *gin.Context) {
	if api.configManager.autoTuning == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "自动调优引擎未启用"})
		return
	}

	status := api.configManager.autoTuning.GetStatus()
	c.JSON(http.StatusOK, gin.H{"data": status})
}

// EnableAutoTuning 启用自动调优
func (api *ConfigAPI) EnableAutoTuning(c *gin.Context) {
	if api.configManager.autoTuning == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "自动调优引擎未启用"})
		return
	}

	api.configManager.autoTuning.Enable()
	c.JSON(http.StatusOK, gin.H{"message": "自动调优已启用"})
}

// DisableAutoTuning 禁用自动调优
func (api *ConfigAPI) DisableAutoTuning(c *gin.Context) {
	if api.configManager.autoTuning == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "自动调优引擎未启用"})
		return
	}

	api.configManager.autoTuning.Disable()
	c.JSON(http.StatusOK, gin.H{"message": "自动调优已禁用"})
}

// GetPerformanceMetrics 获取性能指标
func (api *ConfigAPI) GetPerformanceMetrics(c *gin.Context) {
	if api.configManager.autoTuning == nil || api.configManager.autoTuning.metrics == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "性能指标不可用"})
		return
	}

	metrics := api.configManager.autoTuning.metrics
	c.JSON(http.StatusOK, gin.H{
		"data": gin.H{
			"cpu":                 metrics.CPU,
			"memory":              metrics.Memory,
			"avg_response_time":   metrics.AvgResponseTime.Milliseconds(),
			"request_rate":        metrics.RequestRate,
			"error_rate":          metrics.ErrorRate,
			"database_conn_usage": metrics.DatabaseConnUsage,
			"cache_hit_rate":      metrics.CacheHitRate,
			"updated_at":          metrics.UpdatedAt,
		},
	})
}

// BatchUpdateConfigs 批量更新配置
func (api *ConfigAPI) BatchUpdateConfigs(c *gin.Context) {
	var request struct {
		Configs map[string]interface{} `json:"configs" binding:"required"`
	}

	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "请求参数错误: " + err.Error()})
		return
	}

	results := make(map[string]interface{})
	errors := make(map[string]string)

	for configID, value := range request.Configs {
		// 获取配置信息
		api.configManager.mutex.RLock()
		config := api.configManager.configs[configID]
		api.configManager.mutex.RUnlock()

		if config == nil {
			errors[configID] = "配置不存在"
			continue
		}

		// 类型转换和验证
		var convertedValue interface{}
		switch config.Type {
		case "int":
			if intValue, ok := value.(float64); ok {
				convertedValue = int(intValue)
			} else {
				errors[configID] = "无效的整数值"
				continue
			}
		case "float":
			if floatValue, ok := value.(float64); ok {
				convertedValue = floatValue
			} else {
				errors[configID] = "无效的浮点数值"
				continue
			}
		case "bool":
			if boolValue, ok := value.(bool); ok {
				convertedValue = boolValue
			} else {
				errors[configID] = "无效的布尔值"
				continue
			}
		default:
			convertedValue = value
		}

		// 更新配置
		if err := api.configManager.Set(configID, convertedValue); err != nil {
			errors[configID] = err.Error()
		} else {
			results[configID] = convertedValue
		}
	}

	response := gin.H{
		"message":       "批量更新完成",
		"success":       results,
		"total_configs": len(request.Configs),
		"success_count": len(results),
	}

	if len(errors) > 0 {
		response["errors"] = errors
		response["error_count"] = len(errors)
	}

	c.JSON(http.StatusOK, response)
}

// GetConfigHistory 获取配置变更历史
func (api *ConfigAPI) GetConfigHistory(c *gin.Context) {
	configID := c.Param("id")
	if configID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "配置ID不能为空"})
		return
	}

	// 这里应该从数据库查询配置变更历史
	// 目前返回模拟数据
	history := []gin.H{
		{
			"id":         1,
			"config_id":  configID,
			"old_value":  "1000",
			"new_value":  "800",
			"changed_by": "auto_tuning",
			"reason":     "CPU使用率过高自动调整",
			"changed_at": time.Now().Add(-1 * time.Hour),
		},
		{
			"id":         2,
			"config_id":  configID,
			"old_value":  "800",
			"new_value":  "1000",
			"changed_by": "admin",
			"reason":     "手动调整",
			"changed_at": time.Now().Add(-30 * time.Minute),
		},
	}

	c.JSON(http.StatusOK, gin.H{
		"data":  history,
		"total": len(history),
	})
}

// ExportConfigs 导出配置
func (api *ConfigAPI) ExportConfigs(c *gin.Context) {
	category := c.Query("category")

	var configs map[string]*Config
	if category != "" {
		configs = api.configManager.GetByCategory(category)
	} else {
		api.configManager.mutex.RLock()
		configs = make(map[string]*Config)
		for id, config := range api.configManager.configs {
			configs[id] = config
		}
		api.configManager.mutex.RUnlock()
	}

	exportData := gin.H{
		"exported_at": time.Now(),
		"category":    category,
		"configs":     make([]gin.H, 0, len(configs)),
	}

	for _, config := range configs {
		value, _ := api.configManager.Get(config.ID)
		exportData["configs"] = append(exportData["configs"].([]gin.H), gin.H{
			"id":            config.ID,
			"category":      config.Category,
			"name":          config.Name,
			"value":         value,
			"default_value": config.DefaultValue,
			"description":   config.Description,
			"type":          config.Type,
			"constraints":   config.Constraints,
		})
	}

	c.Header("Content-Disposition", "attachment; filename=configs.json")
	c.JSON(http.StatusOK, exportData)
}

// RegisterRoutes 注册路由
func (api *ConfigAPI) RegisterRoutes(router *gin.RouterGroup) {
	configGroup := router.Group("/configs")
	{
		// 配置管理
		configGroup.GET("", api.GetConfigs)
		configGroup.GET("/:id", api.GetConfig)
		configGroup.PUT("/:id", api.UpdateConfig)
		configGroup.POST("/:id/reset", api.ResetConfig)
		configGroup.PUT("/batch", api.BatchUpdateConfigs)

		// 配置分类
		configGroup.GET("/categories", api.GetCategories)

		// 配置历史
		configGroup.GET("/:id/history", api.GetConfigHistory)

		// 导出配置
		configGroup.GET("/export", api.ExportConfigs)
	}

	// 自动调优
	tuningGroup := router.Group("/auto-tuning")
	{
		tuningGroup.GET("/status", api.GetAutoTuningStatus)
		tuningGroup.POST("/enable", api.EnableAutoTuning)
		tuningGroup.POST("/disable", api.DisableAutoTuning)
		tuningGroup.GET("/metrics", api.GetPerformanceMetrics)
	}
}
