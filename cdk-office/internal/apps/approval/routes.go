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

package approval

import (
	"cdk-office/internal/apps/approval"

	"github.com/gin-gonic/gin"
)

// RegisterRoutes 注册审批流程路由
func RegisterRoutes(router *gin.RouterGroup) {
	// 创建审批服务和处理函数
	service := approval.NewService()
	handler := approval.NewHandler(service)

	// 审批流程相关路由
	approvalGroup := router.Group("/approvals")
	{
		// 创建审批流程
		approvalGroup.POST("", handler.CreateApproval)

		// 获取审批流程列表
		approvalGroup.GET("", handler.ListApprovals)

		// 根据ID获取审批流程
		approvalGroup.GET("/:id", handler.GetApprovalByID)

		// 更新审批状态
		approvalGroup.PUT("/:id/status", handler.UpdateApprovalStatus)

		// 获取审批历史
		approvalGroup.GET("/:id/history", handler.GetApprovalHistory)
	}

	// 审批模板相关路由
	templateGroup := router.Group("/approval-templates")
	{
		// 创建审批模板
		templateGroup.POST("", handler.CreateApprovalTemplate)

		// 获取审批模板列表
		templateGroup.GET("", handler.ListApprovalTemplates)
	}

	// 审批通知相关路由
	notificationGroup := router.Group("/approval-notifications")
	{
		// 创建审批通知
		notificationGroup.POST("", handler.CreateNotification)

		// 获取用户通知列表
		notificationGroup.GET("", handler.ListNotifications)

		// 标记通知为已读
		notificationGroup.PUT("/:id/read", handler.MarkNotificationAsRead)
	}
}
