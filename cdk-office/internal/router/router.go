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

package router

import (
	"context"
	"fmt"
	"log"
	"strconv"

	"github.com/gin-contrib/sessions"
	"github.com/gin-contrib/sessions/redis"
	"github.com/gin-gonic/gin"
	_ "github.com/linux-do/cdk-office/docs"
	"github.com/linux-do/cdk-office/internal/apps/admin"
	"github.com/linux-do/cdk-office/internal/apps/ai"
	"github.com/linux-do/cdk-office/internal/apps/approval"
	"github.com/linux-do/cdk-office/internal/apps/contract"
	"github.com/linux-do/cdk-office/internal/apps/dashboard"
	"github.com/linux-do/cdk-office/internal/apps/filepreview"
	"github.com/linux-do/cdk-office/internal/apps/health"
	"github.com/linux-do/cdk-office/internal/apps/knowledge"
	"github.com/linux-do/cdk-office/internal/apps/notification"
	"github.com/linux-do/cdk-office/internal/apps/oauth"
	"github.com/linux-do/cdk-office/internal/apps/ocr"
	"github.com/linux-do/cdk-office/internal/apps/pdf"
	"github.com/linux-do/cdk-office/internal/apps/project"
	"github.com/linux-do/cdk-office/internal/apps/qrcode"
	"github.com/linux-do/cdk-office/internal/apps/schedule"
	"github.com/linux-do/cdk-office/internal/apps/survey"
	"github.com/linux-do/cdk-office/internal/apps/workflows"
	"github.com/linux-do/cdk-office/internal/config"
	"github.com/linux-do/cdk-office/internal/db"
	"github.com/linux-do/cdk-office/internal/middleware"
	"github.com/linux-do/cdk-office/internal/otel_trace"
	"github.com/linux-do/cdk-office/internal/services/optimization"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
	"go.opentelemetry.io/contrib/instrumentation/github.com/gin-gonic/gin/otelgin"
)

// Serve 启动HTTP服务
func Serve() {
	defer otel_trace.Shutdown(context.Background())

	// 运行模式
	if config.Config.App.Env == "production" {
		gin.SetMode(gin.ReleaseMode)
	}

	// 初始化路由
	r := gin.New()
	r.Use(gin.Recovery())

	// Session
	sessionStore, err := redis.NewStoreWithDB(
		config.Config.Redis.MinIdleConn,
		"tcp",
		fmt.Sprintf("%s:%d", config.Config.Redis.Host, config.Config.Redis.Port),
		config.Config.Redis.Username,
		config.Config.Redis.Password,
		strconv.Itoa(config.Config.Redis.DB),
		[]byte(config.Config.App.SessionSecret),
	)
	if err != nil {
		log.Fatalf("[API] init session store failed: %v\n", err)
	}
	sessionStore.Options(
		sessions.Options{
			Path:     "/",
			Domain:   config.Config.App.SessionDomain,
			MaxAge:   config.Config.App.SessionAge,
			HttpOnly: config.Config.App.SessionHttpOnly,
			Secure:   config.Config.App.SessionSecure,
		},
	)
	r.Use(sessions.Sessions(config.Config.App.SessionCookieName, sessionStore))

	// 性能优化中间件
	optimizationConfig := middleware.DefaultOptimizationConfig()
	optimizationMiddleware := middleware.SetupOptimizationMiddleware(r, optimizationConfig)

	// 初始化优化模块
	if err := optimization.InitOptimizationModule(db.GetDB(), nil, nil); err != nil {
		log.Printf("[WARNING] Failed to initialize optimization module: %v", err)
	}

	// 补充中间件
	r.Use(otelgin.Middleware(config.Config.App.AppName), loggerMiddleware())

	apiGroup := r.Group(config.Config.App.APIPrefix)
	{
		if config.Config.App.Env == "development" {
			// Swagger
			apiGroup.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))
		}

		// API V1
		apiV1Router := apiGroup.Group("/v1")
		{
			// Health
			apiV1Router.GET("/health", health.Health)

			// OAuth
			apiV1Router.GET("/oauth/login", oauth.GetLoginURL)
			apiV1Router.GET("/oauth/logout", oauth.LoginRequired(), oauth.Logout)
			apiV1Router.POST("/oauth/callback", oauth.Callback)
			apiV1Router.GET("/oauth/user-info", oauth.LoginRequired(), oauth.UserInfo)

			// Project
			projectRouter := apiV1Router.Group("/projects")
			projectRouter.Use(oauth.LoginRequired())
			{
				projectRouter.GET("/mine", project.ListMyProjects)
				projectRouter.GET("", project.ListProjects)
				projectRouter.POST("", project.CreateProject)
				projectRouter.PUT("/:id", project.ProjectCreatorPermMiddleware(), project.UpdateProject)
				projectRouter.DELETE("/:id", project.ProjectCreatorPermMiddleware(), project.DeleteProject)
				projectRouter.POST("/:id/receive", project.ReceiveProjectMiddleware(), project.ReceiveProject)
				projectRouter.POST("/:id/report", project.ReportProject)
				projectRouter.GET("/received", project.ListReceiveHistory)
				projectRouter.GET("/:id", project.GetProject)
			}

			// Tag
			tagRouter := apiV1Router.Group("/tags")
			tagRouter.Use(oauth.LoginRequired())
			{
				tagRouter.GET("", project.ListTags)
			}

			// Dashboard (待办事项和日程)
			dashboardService := dashboard.NewService(db.GetDB())
			dashboardHandler := dashboard.NewHandler(dashboardService)
			dashboardRouter := apiV1Router.Group("")
			dashboardRouter.Use(oauth.LoginRequired())
			{
				dashboard.RegisterRoutes(dashboardRouter, dashboardHandler)
			}

			// AI Services (智能问答和知识库同步)
			aiRouter, err := ai.NewAIRouter(db.GetDB())
			if err != nil {
				log.Printf("[WARNING] Failed to initialize AI router: %v", err)
			} else {
				aiRouterGroup := apiV1Router.Group("")
				aiRouterGroup.Use(oauth.LoginRequired())
				// 应用标准API限流
				optimizationMiddleware.ApplyRateLimiting(aiRouterGroup, "api")
				aiRouter.RegisterRoutes(aiRouterGroup)
			}

			// OCR Services (with fallback support)
			ocrRouterInstance := ocr.NewRouter(db.GetDB())
			ocrRouterInstance.RegisterRoutes(apiV1Router)

			// QRCode
			qrRouter := apiV1Router.Group("/qrcode")
			qrRouter.Use(oauth.LoginRequired())
			// 应用标准API限流
			optimizationMiddleware.ApplyRateLimiting(qrRouter, "api")
			{
				// 表单管理
				qrRouter.POST("/forms", qrcode.CreateForm)
				qrRouter.GET("/forms", qrcode.ListForms)
				qrRouter.GET("/forms/:id", qrcode.GetForm)
				qrRouter.PUT("/forms/:id", qrcode.UpdateForm)
				qrRouter.DELETE("/forms/:id", qrcode.DeleteForm)

				// 二维码生成
				qrRouter.POST("/generate", qrcode.GenerateQRCode)
				qrRouter.GET("/records", qrcode.ListRecords)

				// 表单提交
				qrRouter.POST("/submit/:formId", qrcode.SubmitForm)
				qrRouter.GET("/submissions/:formId", qrcode.ListSubmissions)
			}

			// Approval
			approvalRouter := apiV1Router.Group("/approval")
			approvalRouter.Use(oauth.LoginRequired())
			{
				approval.RegisterRoutes(approvalRouter)
			}

			// Notification
			notificationRouter := apiV1Router.Group("/notification")
			notificationRouter.Use(oauth.LoginRequired())
			{
				notification.RegisterRoutes(notificationRouter)
			}

			// Contract (电子合同)
			contractRouter := apiV1Router.Group("/contract")
			contractRouter.Use(oauth.LoginRequired())
			{
				contract.RegisterRoutes(contractRouter)
			}

			// Survey (调查问卷)
			surveyRouter := apiV1Router.Group("/surveys")
			surveyRouter.Use(oauth.LoginRequired())
			{
				survey.RegisterRoutes(surveyRouter)
			}

			// Survey 公开访问路由（无需认证）
			publicSurveyRouter := apiV1Router.Group("/public/surveys")
			{
				survey.RegisterPublicRoutes(publicSurveyRouter)
			}

			// Workflows (工作流)
			workflowHandler := workflows.NewHandler()
			workflowRouter := apiV1Router.Group("/workflows")
			workflowRouter.Use(oauth.LoginRequired())
			{
				workflowHandler.RegisterRoutes(workflowRouter)
			}

			// Schedule (调度系统)
			scheduleHandler := schedule.NewHandler()
			scheduleHandler.StartService() // 启动调度服务
			scheduleRouter := apiV1Router.Group("/schedule")
			scheduleRouter.Use(oauth.LoginRequired())
			{
				scheduleHandler.RegisterRoutes(scheduleRouter)
			}

			// PDF Processing (PDF处理)
			pdfHandler := pdf.NewHandler()
			pdfRouter := apiV1Router.Group("/pdf")
			pdfRouter.Use(oauth.LoginRequired())
			// 应用上传限流和缓存
			optimizationMiddleware.ApplyRateLimiting(pdfRouter, "upload")
			optimizationMiddleware.ApplyCaching(pdfRouter)
			{
				pdfHandler.RegisterRoutes(pdfRouter)
			}

			// Knowledge Base (个人知识库)
			knowledgeHandler := knowledge.NewHandler()
			knowledgeRouter := apiV1Router.Group("/")
			knowledgeRouter.Use(oauth.LoginRequired())
			// 应用缓存中间件提高查询性能
			optimizationMiddleware.ApplyCaching(knowledgeRouter)
			{
				knowledgeHandler.RegisterRoutes(knowledgeRouter)
			}

			// File Preview (文件预览)
			filePreviewHandler := filepreview.NewHandler()
			filePreviewRouter := apiV1Router.Group("/preview")
			filePreviewRouter.Use(oauth.LoginRequired())
			// 应用缓存和限流
			optimizationMiddleware.ApplyRateLimiting(filePreviewRouter, "api")
			optimizationMiddleware.ApplyCaching(filePreviewRouter)
			{
				filePreviewHandler.RegisterRoutes(filePreviewRouter)
			}

			// 系统优化管理
			optimization.RegisterOptimizationRoutes(apiV1Router)

			// Admin
			adminRouter := apiV1Router.Group("/admin")
			adminRouter.Use(oauth.LoginRequired(), admin.LoginAdminRequired())
			{
				// Project
				projectAdminRouter := adminRouter.Group("/projects")
				{
					projectAdminRouter.GET("", admin.GetProjectsList)
					projectAdminRouter.PUT("/:id/review", admin.ReviewProject)
				}

				// AI服务管理
				aiAdminRouter := adminRouter.Group("/ai")
				{
					aiAdminRouter.GET("/services", admin.ListAIServices)
					aiAdminRouter.POST("/services", admin.CreateAIService)
					aiAdminRouter.PUT("/services/:id", admin.UpdateAIService)
					aiAdminRouter.DELETE("/services/:id", admin.DeleteAIService)
					aiAdminRouter.POST("/services/:id/test", admin.TestAIService)
					aiAdminRouter.PUT("/services/:id/toggle", admin.ToggleAIService)
				}

				// OCR服务管理
				ocrAdminRouter := adminRouter.Group("/ocr")
				{
					ocrAdminRouter.GET("/services", admin.ListOCRServices)
					ocrAdminRouter.POST("/services", admin.CreateOCRService)
					ocrAdminRouter.PUT("/services/:id", admin.UpdateOCRService)
					ocrAdminRouter.DELETE("/services/:id", admin.DeleteOCRService)
					ocrAdminRouter.POST("/services/:id/test", admin.TestOCRService)
					ocrAdminRouter.PUT("/services/:id/toggle", admin.ToggleOCRService)
				}

				// 系统状态
				adminRouter.GET("/system/status", admin.GetSystemStatus)
			}
		}
	}

	// Serve
	log.Printf("[API] starting server on %s", config.Config.App.Addr)
	if err := r.Run(config.Config.App.Addr); err != nil {
		log.Fatalf("[API] serve api failed: %v\n", err)
	}
}
