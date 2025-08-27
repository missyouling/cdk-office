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

package contract

import (
	"github.com/gin-gonic/gin"
)

// RegisterRoutes 注册合同相关路由
func RegisterRoutes(router *gin.RouterGroup) {
	handler := NewHandler()

	// 合同管理路由组
	contracts := router.Group("/contracts")
	{
		// 基础CRUD操作
		contracts.POST("", handler.CreateContract)           // 创建合同
		contracts.GET("", handler.GetContracts)              // 获取合同列表
		contracts.GET("/:id", handler.GetContract)           // 获取合同详情
		contracts.PUT("/:id", handler.UpdateContract)        // 更新合同
		contracts.DELETE("/:id", handler.DeleteContract)     // 删除合同

		// 合同操作
		contracts.POST("/:id/send", handler.SendContract)    // 发送合同
		contracts.POST("/:id/sign", handler.SignContract)    // 签署合同
		contracts.POST("/:id/reject", handler.RejectContract) // 拒绝合同
		contracts.POST("/:id/cancel", handler.CancelContract) // 取消合同

		// 文件操作
		contracts.GET("/:id/download", handler.GetContractFile) // 下载合同文件

		// 日志查看
		contracts.GET("/:id/logs", handler.GetContractLogs) // 获取操作日志
	}

	// 合同模板路由组
	templates := router.Group("/contract-templates")
	{
		templates.POST("", handler.CreateTemplate)           // 创建模板
		templates.GET("", handler.GetTemplates)              // 获取模板列表
		templates.GET("/:id", handler.GetTemplate)           // 获取模板详情
		templates.PUT("/:id", handler.UpdateTemplate)        // 更新模板
		templates.DELETE("/:id", handler.DeleteTemplate)     // 删除模板
		templates.POST("/:id/copy", handler.CopyTemplate)    // 复制模板
	}

	// 合同统计路由组
	statistics := router.Group("/contract-statistics")
	{
		statistics.GET("/dashboard", handler.GetContractDashboard) // 合同仪表板
		statistics.GET("/overview", handler.GetContractOverview)   // 合同概览
		statistics.GET("/trends", handler.GetContractTrends)       // 合同趋势
	}

	// 合同配置路由组
	config := router.Group("/contract-config")
	{
		config.GET("/services", handler.GetServiceConfigs)       // 获取服务配置
		config.POST("/services", handler.CreateServiceConfig)    // 创建服务配置
		config.PUT("/services/:id", handler.UpdateServiceConfig) // 更新服务配置
		config.DELETE("/services/:id", handler.DeleteServiceConfig) // 删除服务配置
		config.POST("/services/:id/test", handler.TestServiceConfig) // 测试服务配置

		config.GET("/workflows", handler.GetWorkflowConfigs)     // 获取工作流配置
		config.POST("/workflows", handler.CreateWorkflowConfig)  // 创建工作流配置
		config.PUT("/workflows/:id", handler.UpdateWorkflowConfig) // 更新工作流配置
		config.DELETE("/workflows/:id", handler.DeleteWorkflowConfig) // 删除工作流配置
	}

	// 知识库集成路由组
	knowledge := router.Group("/contract-knowledge")
	{
		knowledge.GET("/submissions", handler.GetKnowledgeSubmissions)     // 获取知识库提交列表
		knowledge.POST("/submissions/:id/approve", handler.ApproveKnowledgeSubmission) // 审批知识库提交
		knowledge.POST("/submissions/:id/reject", handler.RejectKnowledgeSubmission)   // 拒绝知识库提交
		knowledge.POST("/sync", handler.SyncToKnowledgeBase)               // 手动同步到知识库
	}
}