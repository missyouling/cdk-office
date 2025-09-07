package main

import (
	"time"

	app_handler "cdk-office/internal/app/handler"
	auth_handler "cdk-office/internal/auth/handler"
	"cdk-office/internal/auth/service"
	document_handler "cdk-office/internal/document/handler"
	employee_handler "cdk-office/internal/employee/handler"
	business_handler "cdk-office/internal/business/handler"
	"cdk-office/internal/shared/cache"
	"cdk-office/internal/shared/middleware"
	"cdk-office/pkg/config"
	"cdk-office/pkg/jwt"

	"github.com/gin-gonic/gin"
)

func main() {
	// Initialize configuration
	config.Init()

	// Initialize logger
	// logger.Init() // Logger doesn't have Init function

	// Initialize database
	// db := config.InitDatabase() // Config doesn't have InitDatabase function
	// database.InitDB(db)

	// Initialize Redis cache
	cache.InitRedis()

	// Initialize JWT manager
	jwtConfig := &jwt.JWTConfig{
		SecretKey:       "cdk-office-secret-key", // In production, use environment variable
		AccessTokenExp:  time.Hour * 2,
		RefreshTokenExp: time.Hour * 24 * 7,
	}
	jwtManager := jwt.NewJWTManager(jwtConfig)

	// Create Gin engine
	r := gin.New()

	// Use middleware
	r.Use(gin.Logger())
	r.Use(gin.Recovery())
	securityMiddleware := middleware.NewSecurityHeadersMiddleware()
	r.Use(securityMiddleware.SecurityHeaders())
	corsMiddleware := middleware.NewCORSMiddleware()
	r.Use(corsMiddleware.CORS())

	// Initialize services for middleware
	permissionService := service.NewPermissionService()
	authMiddleware := middleware.NewAuthMiddleware(jwtManager)
	_ = middleware.NewPermissionMiddleware(jwtManager, permissionService) // Not used yet

	// API v1 group
	v1 := r.Group("/api/v1")
	{
		// Authentication routes
		auth := v1.Group("/auth")
		{
			authHandler := auth_handler.NewAuthHandler(jwtManager)
			auth.POST("/register", authHandler.Register)
			auth.POST("/login", authHandler.Login)
			auth.GET("/user/:id", authHandler.GetUserInfo)
			auth.POST("/logout", authHandler.Logout)
			auth.POST("/refresh", authHandler.RefreshToken)
			
			// WeChat login route
			wechatHandler := auth_handler.NewWeChatHandler()
			auth.POST("/wechat/login", wechatHandler.WeChatLogin)
		}

		// Document routes
		documents := v1.Group("/documents")
		documents.Use(authMiddleware.Authenticate())
		{
			documentHandler := document_handler.NewDocumentHandler()
			documents.POST("", documentHandler.Upload)
			documents.GET("/:id", documentHandler.GetDocument)
			documents.PUT("/:id", documentHandler.UpdateDocument)
			documents.DELETE("/:id", documentHandler.DeleteDocument)
			documents.GET("/:id/versions", documentHandler.GetDocumentVersions)
		}

		// Document category routes
		categories := v1.Group("/categories")
		{
			categoryHandler := document_handler.NewCategoryHandler()
			categories.POST("", categoryHandler.CreateCategory)
			categories.GET("/:id", categoryHandler.GetCategory)
			categories.PUT("/:id", categoryHandler.UpdateCategory)
			categories.DELETE("/:id", categoryHandler.DeleteCategory)
			categories.GET("", categoryHandler.ListCategories)
		}

		// Document version routes
		versions := v1.Group("/versions")
		{
			versionHandler := document_handler.NewVersionHandler()
			versions.POST("", versionHandler.CreateVersion)
			versions.GET("/:id", versionHandler.GetVersion)
			versions.GET("/document/:docId", versionHandler.ListVersions)
		}

		// Document search routes
		search := v1.Group("/search")
		{
			searchHandler := document_handler.NewSearchHandler()
			search.GET("", searchHandler.SearchDocuments)
		}

		// Employee routes
		employees := v1.Group("/employees")
		employees.Use(authMiddleware.Authenticate())
		{
			employeeHandler := employee_handler.NewEmployeeHandler()
			employees.POST("", employeeHandler.CreateEmployee)
			employees.GET("/:id", employeeHandler.GetEmployee)
			employees.PUT("/:id", employeeHandler.UpdateEmployee)
			employees.DELETE("/:id", employeeHandler.DeleteEmployee)
			employees.GET("", employeeHandler.ListEmployees)
		}

		// Department routes
		departments := v1.Group("/departments")
		{
			departmentHandler := employee_handler.NewDepartmentHandler()
			departments.POST("", departmentHandler.CreateDepartment)
			departments.GET("/:id", departmentHandler.GetDepartment)
			departments.PUT("/:id", departmentHandler.UpdateDepartment)
			departments.DELETE("/:id", departmentHandler.DeleteDepartment)
			departments.GET("", departmentHandler.ListDepartments)
		}

		// Employee analytics routes
		analytics := v1.Group("/analytics")
		{
			analyticsHandler := employee_handler.NewAnalyticsHandler()
			analytics.GET("/employee/count-by-department", analyticsHandler.GetEmployeeCountByDepartment)
			analytics.GET("/employee/count-by-position", analyticsHandler.GetEmployeeCountByPosition)
			analytics.GET("/employee/lifecycle-stats", analyticsHandler.GetEmployeeLifecycleStats)
			analytics.GET("/employee/age-distribution", analyticsHandler.GetEmployeeAgeDistribution)
			analytics.GET("/employee/performance-stats", analyticsHandler.GetEmployeePerformanceStats)
			analytics.GET("/employee/termination-analysis", analyticsHandler.GetTerminationAnalysis)
			analytics.GET("/employee/turnover-rate", analyticsHandler.GetEmployeeTurnoverRate)
			analytics.GET("/employee/survey-analysis", analyticsHandler.GetSurveyAnalysis)
		}

		// Employee lifecycle routes
		lifecycle := v1.Group("/lifecycle")
		{
			lifecycleHandler := employee_handler.NewLifecycleHandler()
			lifecycle.POST("/promote", lifecycleHandler.PromoteEmployee)
			lifecycle.POST("/transfer", lifecycleHandler.TransferEmployee)
			lifecycle.POST("/terminate", lifecycleHandler.TerminateEmployee)
			lifecycle.GET("/employee/:id", lifecycleHandler.GetEmployeeLifecycleHistory)
		}

		// Business module routes
		modules := v1.Group("/modules")
		{
			moduleHandler := business_handler.NewModuleHandler()
			modules.POST("", moduleHandler.CreateModule)
			modules.GET("/:id", moduleHandler.GetModule)
			modules.PUT("/:id", moduleHandler.UpdateModule)
			modules.DELETE("/:id", moduleHandler.DeleteModule)
			modules.GET("", moduleHandler.ListModules)
		}

		// Business plugin routes
		plugins := v1.Group("/plugins")
		{
			pluginHandler := business_handler.NewPluginHandler()
			plugins.POST("", pluginHandler.RegisterPlugin)
			plugins.GET("/:id", pluginHandler.GetPlugin)
			plugins.DELETE("/:id", pluginHandler.UnregisterPlugin)
			plugins.GET("", pluginHandler.ListPlugins)
			plugins.POST("/:id/enable", pluginHandler.EnablePlugin)
			plugins.POST("/:id/disable", pluginHandler.DisablePlugin)
		}

		// Business contract routes
		contracts := v1.Group("/contracts")
		{
			contractHandler := business_handler.NewContractHandler()
			contracts.POST("", contractHandler.CreateContract)
			contracts.GET("/:id", contractHandler.GetContract)
			contracts.PUT("/:id", contractHandler.UpdateContract)
			contracts.DELETE("/:id", contractHandler.DeleteContract)
			contracts.GET("", contractHandler.ListContracts)
		}

		// Business survey routes
		surveys := v1.Group("/surveys")
		{
			surveyHandler := business_handler.NewSurveyHandler()
			surveys.POST("", surveyHandler.CreateSurvey)
			surveys.GET("/:id", surveyHandler.GetSurvey)
			surveys.PUT("/:id", surveyHandler.UpdateSurvey)
			surveys.DELETE("/:id", surveyHandler.DeleteSurvey)
			surveys.GET("", surveyHandler.ListSurveys)
		}

		// Business permission routes
		// businessPermissions := v1.Group("/business-permissions")
		// {
		// 	// permissionHandler := business_handler.NewPermissionHandler() // This handler doesn't exist yet
		// 	// businessPermissions.POST("", permissionHandler.CreatePermission)
		// 	// businessPermissions.GET("/:id", permissionHandler.GetPermission)
		// 	// businessPermissions.PUT("/:id", permissionHandler.UpdatePermission)
		// 	// businessPermissions.DELETE("/:id", permissionHandler.DeletePermission)
		// 	// businessPermissions.GET("", permissionHandler.ListPermissions)
		// }

		// Application center routes
		apps := v1.Group("/apps")
		{
			appHandler := app_handler.NewAppHandler()
			apps.POST("", appHandler.CreateApplication)
			apps.GET("/:id", appHandler.GetApplication)
			apps.PUT("/:id", appHandler.UpdateApplication)
			apps.DELETE("/:id", appHandler.DeleteApplication)
			apps.GET("", appHandler.ListApplications)
		}

		// QR code routes
		qrcodes := v1.Group("/qrcodes")
		{
			qrCodeHandler := app_handler.NewQRCodeHandler()
			qrcodes.POST("", qrCodeHandler.CreateQRCode)
			qrcodes.GET("/:id", qrCodeHandler.GetQRCode)
			qrcodes.PUT("/:id", qrCodeHandler.UpdateQRCode)
			qrcodes.DELETE("/:id", qrCodeHandler.DeleteQRCode)
			qrcodes.GET("", qrCodeHandler.ListQRCodes)
			qrcodes.POST("/:id/generate", qrCodeHandler.GenerateQRCodeImage)
		}

		// Form routes
		forms := v1.Group("/forms")
		{
			formHandler := app_handler.NewFormHandler()
			forms.POST("", formHandler.CreateForm)
			forms.GET("/:id", formHandler.GetForm)
			forms.PUT("/:id", formHandler.UpdateForm)
			forms.DELETE("/:id", formHandler.DeleteForm)
			forms.GET("", formHandler.ListForms)
			
			// Form data submission routes
			formData := forms.Group("/data")
			{
				formData.POST("", formHandler.SubmitFormData)
				formData.GET("", formHandler.ListFormDataEntries)
			}
		}

		// Application permission routes
		appPermissions := v1.Group("/app-permissions")
		{
			permissionHandler := app_handler.NewAppPermissionHandler()
			appPermissions.POST("", permissionHandler.CreateAppPermission)
			appPermissions.GET("/:id", permissionHandler.GetAppPermission)
			appPermissions.PUT("/:id", permissionHandler.UpdateAppPermission)
			appPermissions.DELETE("/:id", permissionHandler.DeleteAppPermission)
			appPermissions.GET("", permissionHandler.ListAppPermissions)
			appPermissions.POST("/assign", permissionHandler.AssignPermissionToUser)
			appPermissions.POST("/revoke", permissionHandler.RevokePermissionFromUser)
			appPermissions.GET("/user", permissionHandler.ListUserPermissions)
			appPermissions.POST("/check", permissionHandler.CheckUserPermission)
		}

		// Batch QR code routes
		batchQRCodes := v1.Group("/batch-qrcodes")
		{
			batchHandler := app_handler.NewBatchQRCodeHandler()
			batchQRCodes.POST("", batchHandler.CreateBatchQRCode)
			batchQRCodes.GET("/:id", batchHandler.GetBatchQRCode)
			batchQRCodes.PUT("/:id", batchHandler.UpdateBatchQRCode)
			batchQRCodes.DELETE("/:id", batchHandler.DeleteBatchQRCode)
			batchQRCodes.GET("", batchHandler.ListBatchQRCodes)
			batchQRCodes.POST("/:id/generate", batchHandler.GenerateBatchQRCodes)
		}

		// Form designer routes
		formDesigns := v1.Group("/form-designs")
		{
			formDesignerHandler := app_handler.NewFormDesignerHandler()
			formDesigns.POST("", formDesignerHandler.CreateFormDesign)
			formDesigns.GET("/:id", formDesignerHandler.GetFormDesign)
			formDesigns.PUT("/:id", formDesignerHandler.UpdateFormDesign)
			formDesigns.DELETE("/:id", formDesignerHandler.DeleteFormDesign)
			formDesigns.GET("", formDesignerHandler.ListFormDesigns)
			formDesigns.POST("/:id/publish", formDesignerHandler.PublishFormDesign)
		}

		// Data collection routes
		dataCollections := v1.Group("/data-collections")
		{
			dataCollectionHandler := app_handler.NewDataCollectionHandler()
			dataCollections.POST("", dataCollectionHandler.CreateDataCollection)
			dataCollections.GET("/:id", dataCollectionHandler.GetDataCollection)
			dataCollections.PUT("/:id", dataCollectionHandler.UpdateDataCollection)
			dataCollections.DELETE("/:id", dataCollectionHandler.DeleteDataCollection)
			dataCollections.GET("", dataCollectionHandler.ListDataCollections)
			
			// Data entry routes
			dataEntries := dataCollections.Group("/entries")
			{
				dataEntries.POST("", dataCollectionHandler.SubmitDataEntry)
				dataEntries.GET("", dataCollectionHandler.ListDataEntries)
				dataEntries.GET("/export/:id", dataCollectionHandler.ExportDataEntries)
			}
		}
	}

	// Start server
	port := "8080" // Using default port since config.Get() is not available
	
	// logger.Info("Starting server on port " + port) // Logger doesn't have Info method
	r.Run(":" + port)
}