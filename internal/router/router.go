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
	"github.com/linux-do/cdk-office/internal/apps/dashboard"
	"github.com/linux-do/cdk-office/internal/apps/health"
	"github.com/linux-do/cdk-office/internal/apps/oauth"
	"github.com/linux-do/cdk-office/internal/apps/ocr"
	"github.com/linux-do/cdk-office/internal/apps/project"
	"github.com/linux-do/cdk-office/internal/apps/qrcode"
	"github.com/linux-do/cdk-office/internal/config"
	"github.com/linux-do/cdk-office/internal/otel_trace"
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

			// Dashboard
			dashboardRouter := apiV1Router.Group("/dashboard")
			dashboardRouter.Use(oauth.LoginRequired())
			{
				dashboardRouter.GET("/stats/all", dashboard.GetAllStats)
			}

			// AI (Dify集成)
			aiRouter := apiV1Router.Group("/ai")
			aiRouter.Use(oauth.LoginRequired())
			{
				aiRouter.POST("/chat", ai.Chat)
				aiRouter.POST("/knowledge-sync", ai.SyncKnowledge)
				aiRouter.GET("/knowledge-status", ai.GetKnowledgeSyncStatus)
			}

			// OCR
			ocrRouter := apiV1Router.Group("/ocr")
			ocrRouter.Use(oauth.LoginRequired())
			{
				ocrRouter.POST("/process", ocr.ProcessDocument)
				ocrRouter.GET("/providers", ocr.ListProviders)
				ocrRouter.POST("/providers/test", ocr.TestProvider)
			}

			// QRCode
			qrRouter := apiV1Router.Group("/qrcode")
			qrRouter.Use(oauth.LoginRequired())
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
