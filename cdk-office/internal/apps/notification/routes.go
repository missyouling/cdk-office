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

package notification

import (
	"github.com/gin-gonic/gin"
)

// RegisterRoutes 注册通知中心路由
func RegisterRoutes(router *gin.RouterGroup) {
	// 创建服务实例
	service := NewService()

	// 创建处理函数实例
	handler := NewHandler(service)

	// 通知相关路由
	notifications := router.Group("/notifications")
	{
		notifications.POST("", handler.CreateNotification)             // 创建通知
		notifications.GET("/:id", handler.GetNotificationByID)         // 获取通知详情
		notifications.GET("", handler.ListNotifications)               // 获取通知列表
		notifications.PUT("/:id/read", handler.MarkAsRead)             // 标记通知为已读
		notifications.PUT("/read", handler.MarkMultipleAsRead)         // 批量标记通知为已读
		notifications.PUT("/read-all", handler.MarkAllAsRead)          // 标记所有通知为已读
		notifications.PUT("/:id/archive", handler.ArchiveNotification) // 归档通知
		notifications.DELETE("/:id", handler.DeleteNotification)       // 删除通知
	}

	// 通知模板相关路由
	templates := router.Group("/templates")
	{
		templates.POST("", handler.CreateTemplate)       // 创建通知模板
		templates.GET("", handler.ListTemplates)         // 获取通知模板列表
		templates.GET("/:id", handler.GetTemplateByID)   // 获取通知模板详情
		templates.PUT("/:id", handler.UpdateTemplate)    // 更新通知模板
		templates.DELETE("/:id", handler.DeleteTemplate) // 删除通知模板
	}

	// 用户偏好设置相关路由
	preferences := router.Group("/preferences")
	{
		preferences.GET("", handler.GetUserPreference)    // 获取用户通知偏好设置
		preferences.PUT("", handler.UpdateUserPreference) // 更新用户通知偏好设置
	}

	// 通知渠道相关路由
	channels := router.Group("/channels")
	{
		channels.POST("", handler.CreateChannel)       // 创建通知渠道
		channels.GET("", handler.ListChannels)         // 获取通知渠道列表
		channels.GET("/:id", handler.GetChannelByID)   // 获取通知渠道详情
		channels.PUT("/:id", handler.UpdateChannel)    // 更新通知渠道
		channels.DELETE("/:id", handler.DeleteChannel) // 删除通知渠道
	}

	// 其他相关路由
	router.GET("/unread-count", handler.GetUnreadCount) // 获取未读通知数量
}
