/*
 * MIT License
 *
 * Copyright (c) 2025 CDK-Office
 */

package dictionary

import (
	"github.com/gin-gonic/gin"
)

// Router 数据字典路由
type Router struct {
	handler *Handler
}

// NewRouter 创建数据字典路由
func NewRouter(handler *Handler) *Router {
	return &Router{
		handler: handler,
	}
}

// RegisterRoutes 注册路由
func (r *Router) RegisterRoutes(rg *gin.RouterGroup) {
	// 数据字典管理
	dictionary := rg.Group("/dictionary")
	{
		// 数据实体管理
		entities := dictionary.Group("/entities")
		{
			entities.POST("", r.handler.CreateEntity)                       // 创建数据实体
			entities.GET("", r.handler.ListEntities)                        // 获取实体列表
			entities.GET("/:id", r.handler.GetEntity)                       // 获取单个实体
			entities.PUT("/:id", r.handler.UpdateEntity)                    // 更新实体
			entities.DELETE("/:id", r.handler.DeleteEntity)                 // 删除实体
			entities.GET("/:entity_id/fields", r.handler.GetFieldsByEntity) // 获取实体字段列表
		}

		// 字段定义管理
		fields := dictionary.Group("/fields")
		{
			fields.POST("", r.handler.CreateField)             // 创建字段定义
			fields.PUT("/:field_id", r.handler.UpdateField)    // 更新字段定义
			fields.DELETE("/:field_id", r.handler.DeleteField) // 删除字段定义
		}

		// 元数据和模板
		meta := dictionary.Group("/meta")
		{
			meta.GET("/data-types", r.handler.GetDataTypes)           // 获取支持的数据类型
			meta.GET("/field-templates", r.handler.GetFieldTemplates) // 获取字段模板
		}
	}
}
