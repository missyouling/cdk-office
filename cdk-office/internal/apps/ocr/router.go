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

package ocr

import (
	"context"
	"fmt"
	"io"
	"log"
	"net/http"
	"strconv"
	"sync"

	"github.com/gin-gonic/gin"
	"github.com/linux-do/cdk-office/internal/models"
	"gorm.io/gorm"
)

// OCRServiceManager OCR服务管理器
type OCRServiceManager struct {
	db       *gorm.DB
	services map[string]*models.AIServiceConfig
	mutex    sync.RWMutex
}

// NewOCRServiceManager 创建OCR服务管理器
func NewOCRServiceManager(db *gorm.DB) *OCRServiceManager {
	manager := &OCRServiceManager{
		db:       db,
		services: make(map[string]*models.AIServiceConfig),
	}

	manager.loadOCRServices()
	return manager
}

// loadOCRServices 加载OCR服务配置
func (m *OCRServiceManager) loadOCRServices() error {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	var configs []models.AIServiceConfig
	if err := m.db.Where("service_type LIKE ? AND is_enabled = ?", "ocr%", true).Find(&configs).Error; err != nil {
		return err
	}

	for _, config := range configs {
		m.services[config.ID] = &config
	}

	log.Printf("Loaded %d OCR service configurations", len(m.services))
	return nil
}

// GetDefaultOCRService 获取默认OCR服务
func (m *OCRServiceManager) GetDefaultOCRService(ocrType string) (*models.AIServiceConfig, error) {
	m.mutex.RLock()
	defer m.mutex.RUnlock()

	serviceType := "ocr"
	if ocrType != "" {
		serviceType = "ocr_" + ocrType
	}

	// 查找默认服务
	for _, service := range m.services {
		if service.ServiceType == serviceType && service.IsDefault && service.IsEnabled {
			return service, nil
		}
	}

	// 如果没有默认服务，返回第一个可用的服务
	for _, service := range m.services {
		if service.ServiceType == serviceType && service.IsEnabled {
			return service, nil
		}
	}

	return nil, fmt.Errorf("no available OCR service for type: %s", serviceType)
}

// TriggerOCRFallback 触发OCR服务降级
func (m *OCRServiceManager) TriggerOCRFallback(failedServiceID string) error {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	failedService, exists := m.services[failedServiceID]
	if !exists {
		return fmt.Errorf("failed OCR service not found: %s", failedServiceID)
	}

	log.Printf("Triggering OCR fallback for service: %s (%s)", failedService.ServiceName, failedService.ServiceType)

	// 查找备用服务
	var backupServices []*models.AIServiceConfig
	for _, service := range m.services {
		if service.ServiceType == failedService.ServiceType && service.ID != failedServiceID && service.IsEnabled {
			backupServices = append(backupServices, service)
		}
	}

	// 按优先级排序
	for i := 0; i < len(backupServices)-1; i++ {
		for j := i + 1; j < len(backupServices); j++ {
			if backupServices[i].Priority < backupServices[j].Priority {
				backupServices[i], backupServices[j] = backupServices[j], backupServices[i]
			}
		}
	}

	if len(backupServices) > 0 {
		// 切换到第一个备用服务
		backupService := backupServices[0]
		log.Printf("Switching to backup OCR service: %s", backupService.ServiceName)

		// 临时设置备用服务为默认
		if err := m.db.Model(&models.AIServiceConfig{}).
			Where("service_type = ?", failedService.ServiceType).
			Update("is_default", false).Error; err != nil {
			return err
		}

		if err := m.db.Model(&models.AIServiceConfig{}).
			Where("id = ?", backupService.ID).
			Update("is_default", true).Error; err != nil {
			return err
		}

		// 重新加载服务配置
		m.loadOCRServices()

		log.Printf("Successfully switched to backup OCR service: %s", backupService.ServiceName)
	} else {
		// 没有备用服务，禁用OCR功能
		log.Printf("No backup OCR services available for type: %s, disabling functionality", failedService.ServiceType)
		return m.disableOCRFeatures(failedService.ServiceType)
	}

	return nil
}

// disableOCRFeatures 禁用OCR功能
func (m *OCRServiceManager) disableOCRFeatures(serviceType string) error {
	// 这里可以实现功能禁用逻辑
	log.Printf("OCR service type %s has been disabled due to lack of available services", serviceType)
	return nil
}

// RecognizeWithFallback 带降级的OCR识别
func (m *OCRServiceManager) RecognizeWithFallback(ctx context.Context, imageData []byte, ocrType string, options OCROptions) (*OCRResult, error) {
	service, err := m.GetDefaultOCRService(ocrType)
	if err != nil {
		return nil, err
	}

	client := CreateOCRClient(service)

	// 尝试识别
	result, err := client.RecognizeText(ctx, imageData, options)
	if err != nil {
		// 如果失败，尝试降级
		log.Printf("OCR recognition failed with service %s: %v", service.ServiceName, err)

		if fallbackErr := m.TriggerOCRFallback(service.ID); fallbackErr == nil {
			// 重新获取服务并重试
			if fallbackService, getErr := m.GetDefaultOCRService(ocrType); getErr == nil {
				fallbackClient := CreateOCRClient(fallbackService)
				if result, err = fallbackClient.RecognizeText(ctx, imageData, options); err == nil {
					log.Printf("OCR recognition succeeded with fallback service: %s", fallbackService.ServiceName)
					return result, nil
				}
			}
		}

		return nil, fmt.Errorf("OCR recognition failed and fallback unsuccessful: %w", err)
	}

	return result, nil
}

// Router OCR路由
type Router struct {
	serviceManager *OCRServiceManager
	db             *gorm.DB
}

// NewRouter 创建OCR路由
func NewRouter(db *gorm.DB) *Router {
	return &Router{
		serviceManager: NewOCRServiceManager(db),
		db:             db,
	}
}

// RegisterRoutes 注册路由
func (r *Router) RegisterRoutes(rg *gin.RouterGroup) {
	ocr := rg.Group("/ocr")
	{
		// OCR识别功能
		ocr.POST("/text", r.RecognizeText)
		ocr.POST("/table", r.RecognizeTable)
		ocr.POST("/handwriting", r.RecognizeHandwriting)
		ocr.POST("/batch", r.BatchRecognize)

		// 图像预处理功能
		ocr.POST("/preprocess", r.PreprocessImage)
		ocr.POST("/analyze-quality", r.AnalyzeImageQuality)
		ocr.POST("/enhanced-text", r.RecognizeTextWithEnhancement)

		// OCR服务配置管理
		ocr.GET("/services", r.GetOCRServices)
		ocr.POST("/services", r.CreateOCRService)
		ocr.PUT("/services/:id", r.UpdateOCRService)
		ocr.DELETE("/services/:id", r.DeleteOCRService)
		ocr.POST("/services/:id/test", r.TestOCRService)

		// 预设OCR服务商
		ocr.GET("/providers", r.GetOCRProviders)
	}
}

// RecognizeText 文字识别
// @Summary 文字识别
// @Description 识别图片中的文字内容
// @Tags OCR
// @Accept multipart/form-data
// @Produce json
// @Param image formData file true "图片文件"
// @Param language formData string false "识别语言"
// @Param accuracy formData string false "准确度模式"
// @Success 200 {object} OCRResult
// @Router /api/ocr/text [post]
func (r *Router) RecognizeText(c *gin.Context) {
	// 获取上传的文件
	file, err := c.FormFile("image")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "No image file provided"})
		return
	}

	// 读取文件内容
	src, err := file.Open()
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Failed to open image file"})
		return
	}
	defer src.Close()

	imageData, err := io.ReadAll(src)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Failed to read image file"})
		return
	}

	// 获取识别选项
	options := OCROptions{
		Language:   c.DefaultPostForm("language", "auto"),
		OutputType: c.DefaultPostForm("output_type", "text"),
		Accuracy:   c.DefaultPostForm("accuracy", "accurate"),
	}

	// 检查是否需要图像预处理
	enhanceImage := c.DefaultPostForm("enhance_image", "false") == "true"
	if enhanceImage {
		// 进行图像增强预处理
		processedData, _, err := r.preprocessImage(c.Request.Context(), imageData)
		if err != nil {
			log.Printf("Image preprocessing failed, using original image: %v", err)
		} else {
			imageData = processedData
		}
	}

	// 执行OCR识别
	result, err := r.serviceManager.RecognizeWithFallback(c.Request.Context(), imageData, "", options)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, result)
}

// RecognizeTable 表格识别
// @Summary 表格识别
// @Description 识别图片中的表格内容
// @Tags OCR
// @Accept multipart/form-data
// @Produce json
// @Param image formData file true "图片文件"
// @Success 200 {object} TableResult
// @Router /api/ocr/table [post]
func (r *Router) RecognizeTable(c *gin.Context) {
	file, err := c.FormFile("image")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "No image file provided"})
		return
	}

	src, err := file.Open()
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Failed to open image file"})
		return
	}
	defer src.Close()

	imageData, err := io.ReadAll(src)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Failed to read image file"})
		return
	}

	// 获取表格OCR服务
	service, err := r.serviceManager.GetDefaultOCRService("table")
	if err != nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "Table OCR service not available"})
		return
	}

	client := CreateOCRClient(service)
	result, err := client.RecognizeTable(c.Request.Context(), imageData)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, result)
}

// RecognizeHandwriting 手写文字识别
// @Summary 手写文字识别
// @Description 识别手写文字内容
// @Tags OCR
// @Accept multipart/form-data
// @Produce json
// @Param image formData file true "图片文件"
// @Success 200 {object} OCRResult
// @Router /api/ocr/handwriting [post]
func (r *Router) RecognizeHandwriting(c *gin.Context) {
	file, err := c.FormFile("image")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "No image file provided"})
		return
	}

	src, err := file.Open()
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Failed to open image file"})
		return
	}
	defer src.Close()

	imageData, err := io.ReadAll(src)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Failed to read image file"})
		return
	}

	options := OCROptions{
		Language:   "auto",
		OutputType: "text",
		Accuracy:   "accurate",
	}

	result, err := r.serviceManager.RecognizeWithFallback(c.Request.Context(), imageData, "handwriting", options)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, result)
}

// BatchRecognize 批量识别
// @Summary 批量OCR识别
// @Description 批量识别多个图片文件
// @Tags OCR
// @Accept multipart/form-data
// @Produce json
// @Param images formData file true "图片文件数组"
// @Success 200 {object} BatchOCRResult
// @Router /api/ocr/batch [post]
func (r *Router) BatchRecognize(c *gin.Context) {
	form, err := c.MultipartForm()
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Failed to parse multipart form"})
		return
	}

	files := form.File["images"]
	if len(files) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "No image files provided"})
		return
	}

	results := make([]BatchOCRItem, 0, len(files))

	for i, file := range files {
		src, err := file.Open()
		if err != nil {
			results = append(results, BatchOCRItem{
				Index:    i,
				Filename: file.Filename,
				Error:    fmt.Sprintf("Failed to open file: %v", err),
			})
			continue
		}

		imageData, err := io.ReadAll(src)
		src.Close()

		if err != nil {
			results = append(results, BatchOCRItem{
				Index:    i,
				Filename: file.Filename,
				Error:    fmt.Sprintf("Failed to read file: %v", err),
			})
			continue
		}

		options := OCROptions{
			Language:   "auto",
			OutputType: "text",
			Accuracy:   "accurate",
		}

		result, err := r.serviceManager.RecognizeWithFallback(c.Request.Context(), imageData, "", options)
		if err != nil {
			results = append(results, BatchOCRItem{
				Index:    i,
				Filename: file.Filename,
				Error:    err.Error(),
			})
		} else {
			results = append(results, BatchOCRItem{
				Index:    i,
				Filename: file.Filename,
				Result:   result,
			})
		}
	}

	c.JSON(http.StatusOK, BatchOCRResult{
		Results: results,
		Total:   len(files),
		Success: len(files) - countErrors(results),
	})
}

// GetOCRServices 获取OCR服务列表
func (r *Router) GetOCRServices(c *gin.Context) {
	var services []models.AIServiceConfig
	if err := r.db.Where("service_type LIKE ?", "ocr%").Find(&services).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"services": services,
		"total":    len(services),
	})
}

// CreateOCRService 创建OCR服务配置
func (r *Router) CreateOCRService(c *gin.Context) {
	userID := c.GetHeader("X-User-ID")
	if userID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	var config models.AIServiceConfig
	if err := c.ShouldBindJSON(&config); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// 确保是OCR类型的服务
	if config.ServiceType == "" {
		config.ServiceType = "ocr"
	}

	config.CreatedBy = userID
	config.UpdatedBy = userID

	if err := r.db.Create(&config).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// 重新加载服务配置
	r.serviceManager.loadOCRServices()

	c.JSON(http.StatusCreated, config)
}

// UpdateOCRService 更新OCR服务配置
func (r *Router) UpdateOCRService(c *gin.Context) {
	userID := c.GetHeader("X-User-ID")
	configID := c.Param("id")

	if userID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	var updates map[string]interface{}
	if err := c.ShouldBindJSON(&updates); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	updates["updated_by"] = userID

	if err := r.db.Model(&models.AIServiceConfig{}).Where("id = ?", configID).Updates(updates).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// 重新加载服务配置
	r.serviceManager.loadOCRServices()

	c.JSON(http.StatusOK, gin.H{"message": "OCR service config updated successfully"})
}

// DeleteOCRService 删除OCR服务配置
func (r *Router) DeleteOCRService(c *gin.Context) {
	userID := c.GetHeader("X-User-ID")
	configID := c.Param("id")

	if userID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	if err := r.db.Delete(&models.AIServiceConfig{}, "id = ?", configID).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// 重新加载服务配置
	r.serviceManager.loadOCRServices()

	c.JSON(http.StatusOK, gin.H{"message": "OCR service config deleted successfully"})
}

// TestOCRService 测试OCR服务
func (r *Router) TestOCRService(c *gin.Context) {
	configID := c.Param("id")

	var config models.AIServiceConfig
	if err := r.db.First(&config, "id = ?", configID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Service config not found"})
		return
	}

	client := CreateOCRClient(&config)

	// 使用测试图片数据（这里应该是一个小的测试图片）
	testImageData := []byte{} // 这里应该放置真实的测试图片数据

	options := OCROptions{
		Language:   "auto",
		OutputType: "text",
		Accuracy:   "fast",
	}

	_, err := client.RecognizeText(c.Request.Context(), testImageData, options)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "OCR service test successful"})
}

// GetOCRProviders 获取OCR服务商预设配置
func (r *Router) GetOCRProviders(c *gin.Context) {
	providers := GetPresetOCRProviders()
	c.JSON(http.StatusOK, providers)
}

// 辅助类型和函数

type BatchOCRResult struct {
	Results []BatchOCRItem `json:"results"`
	Total   int            `json:"total"`
	Success int            `json:"success"`
}

type BatchOCRItem struct {
	Index    int        `json:"index"`
	Filename string     `json:"filename"`
	Result   *OCRResult `json:"result,omitempty"`
	Error    string     `json:"error,omitempty"`
}

func countErrors(results []BatchOCRItem) int {
	count := 0
	for _, result := range results {
		if result.Error != "" {
			count++
		}
	}
	return count
}

// GetPresetOCRProviders 获取预设OCR服务商配置
func GetPresetOCRProviders() map[string]interface{} {
	return map[string]interface{}{
		"ocr_providers": []map[string]interface{}{
			{
				"name":         "baidu_ocr",
				"provider":     "baidu",
				"display_name": "百度OCR",
				"logo":         "/assets/providers/baidu-logo.png",
				"description":  "百度智能云文字识别服务",
				"config_template": map[string]interface{}{
					"api_endpoint": "https://aip.baidubce.com/rest/2.0/ocr/v1/general_basic",
				},
				"required_fields": []string{"api_key", "secret_key"},
				"supported_types": []string{"text", "table", "handwriting", "numbers"},
			},
			{
				"name":         "tencent_ocr",
				"provider":     "tencent",
				"display_name": "腾讯云OCR",
				"logo":         "/assets/providers/tencent-logo.png",
				"description":  "腾讯云文字识别服务",
				"config_template": map[string]interface{}{
					"api_endpoint": "https://ocr.tencentcloudapi.com",
				},
				"required_fields": []string{"secret_id", "secret_key", "region"},
				"supported_types": []string{"text", "table", "handwriting", "id_card", "business_card"},
			},
			{
				"name":         "aliyun_ocr",
				"provider":     "aliyun",
				"display_name": "阿里云OCR",
				"logo":         "/assets/providers/aliyun-logo.png",
				"description":  "阿里云文字识别服务",
				"config_template": map[string]interface{}{
					"api_endpoint": "https://ocr-api.cn-hangzhou.aliyuncs.com",
				},
				"required_fields": []string{"access_key_id", "access_key_secret", "region"},
				"supported_types": []string{"text", "scene", "vehicle", "face"},
			},
		},
	}
}

// PreprocessImage 图像预处理
// @Summary 图像预处理
// @Description 对图像进行降噪、亮度调整、对比度调整、透视矫正等预处理
// @Tags OCR
// @Accept multipart/form-data
// @Produce json
// @Param image formData file true "图片文件"
// @Param enable_denoising formData bool false "启用降噪"
// @Param enable_sharpening formData bool false "启用锐化"
// @Param enable_contrast_adjust formData bool false "启用对比度调整"
// @Param enable_brightness_adjust formData bool false "启用亮度调整"
// @Param enable_perspective_correction formData bool false "启用透视矫正"
// @Param contrast_factor formData number false "对比度因子"
// @Param brightness_factor formData number false "亮度因子"
// @Param sharpness_factor formData number false "锐化因子"
// @Param denoising_strength formData number false "降噪强度"
// @Success 200 {object} map[string]interface{}
// @Router /api/ocr/preprocess [post]
func (r *Router) PreprocessImage(c *gin.Context) {
	// 获取上传的文件
	file, err := c.FormFile("image")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "No image file provided"})
		return
	}

	// 读取文件内容
	src, err := file.Open()
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Failed to open image file"})
		return
	}
	defer src.Close()

	imageData, err := io.ReadAll(src)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Failed to read image file"})
		return
	}

	// 构建处理配置
	config := r.buildImageProcessorConfig(c)

	// 执行图像预处理
	processedData, result, err := r.preprocessImageWithConfig(c.Request.Context(), imageData, config)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// 返回处理结果（包含处理步骤信息）
	c.JSON(http.StatusOK, gin.H{
		"processed_image_size": len(processedData),
		"processing_result":    result,
		"message":              "Image preprocessing completed successfully",
	})
}

// AnalyzeImageQuality 分析图像质量
// @Summary 分析图像质量
// @Description 分析图像的亮度、对比度、清晰度、噪声等质量指标
// @Tags OCR
// @Accept multipart/form-data
// @Produce json
// @Param image formData file true "图片文件"
// @Success 200 {object} ImageQualityAnalysis
// @Router /api/ocr/analyze-quality [post]
func (r *Router) AnalyzeImageQuality(c *gin.Context) {
	// 获取上传的文件
	file, err := c.FormFile("image")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "No image file provided"})
		return
	}

	// 读取文件内容
	src, err := file.Open()
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Failed to open image file"})
		return
	}
	defer src.Close()

	imageData, err := io.ReadAll(src)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Failed to read image file"})
		return
	}

	// 创建图像处理器
	processor := NewImageProcessor(nil) // 使用默认配置

	// 分析图像质量
	analysis, err := processor.AnalyzeImageQuality(imageData)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, analysis)
}

// RecognizeTextWithEnhancement 带图像增强的文字识别
// @Summary 带图像增强的文字识别
// @Description 先对图像进行预处理增强，然后执行OCR文字识别
// @Tags OCR
// @Accept multipart/form-data
// @Produce json
// @Param image formData file true "图片文件"
// @Param language formData string false "识别语言"
// @Param accuracy formData string false "准确度模式"
// @Param enable_preprocessing formData bool false "启用图像预处理"
// @Success 200 {object} map[string]interface{}
// @Router /api/ocr/enhanced-text [post]
func (r *Router) RecognizeTextWithEnhancement(c *gin.Context) {
	// 获取上传的文件
	file, err := c.FormFile("image")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "No image file provided"})
		return
	}

	// 读取文件内容
	src, err := file.Open()
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Failed to open image file"})
		return
	}
	defer src.Close()

	imageData, err := io.ReadAll(src)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Failed to read image file"})
		return
	}

	// 检查是否启用预处理
	enablePreprocessing := c.DefaultPostForm("enable_preprocessing", "true") == "true"
	var processingResult *ImageProcessingResult
	originalSize := len(imageData)

	if enablePreprocessing {
		// 先分析图像质量
		processor := NewImageProcessor(nil)
		qualityAnalysis, _ := processor.AnalyzeImageQuality(imageData)

		// 根据质量分析结果调整处理参数
		config := r.buildAdaptiveConfig(qualityAnalysis)

		// 执行图像预处理
		processedData, result, err := r.preprocessImageWithConfig(c.Request.Context(), imageData, config)
		if err != nil {
			log.Printf("Image preprocessing failed, using original image: %v", err)
		} else {
			imageData = processedData
			processingResult = result
		}
	}

	// 获取识别选项
	options := OCROptions{
		Language:   c.DefaultPostForm("language", "auto"),
		OutputType: c.DefaultPostForm("output_type", "text"),
		Accuracy:   c.DefaultPostForm("accuracy", "accurate"),
	}

	// 执行OCR识别
	result, err := r.serviceManager.RecognizeWithFallback(c.Request.Context(), imageData, "", options)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// 返回增强后的结果
	response := gin.H{
		"ocr_result":            result,
		"original_size":         originalSize,
		"processed_size":        len(imageData),
		"preprocessing_enabled": enablePreprocessing,
	}

	if processingResult != nil {
		response["processing_steps"] = processingResult.ProcessingSteps
		response["compression_ratio"] = processingResult.CompressionRatio
	}

	c.JSON(http.StatusOK, response)
}

// preprocessImage 预处理图像（内部方法）
func (r *Router) preprocessImage(ctx context.Context, imageData []byte) ([]byte, *ImageProcessingResult, error) {
	// 使用默认配置
	processor := NewImageProcessor(nil)
	return processor.ProcessImage(ctx, imageData)
}

// preprocessImageWithConfig 使用指定配置预处理图像
func (r *Router) preprocessImageWithConfig(ctx context.Context, imageData []byte, config *ImageProcessorConfig) ([]byte, *ImageProcessingResult, error) {
	processor := NewImageProcessor(config)
	return processor.ProcessImage(ctx, imageData)
}

// buildImageProcessorConfig 根据请求参数构建图像处理器配置
func (r *Router) buildImageProcessorConfig(c *gin.Context) *ImageProcessorConfig {
	config := &ImageProcessorConfig{
		EnableDenoising:             c.DefaultPostForm("enable_denoising", "true") == "true",
		EnableSharpening:            c.DefaultPostForm("enable_sharpening", "true") == "true",
		EnableContrastAdjust:        c.DefaultPostForm("enable_contrast_adjust", "true") == "true",
		EnableBrightnessAdjust:      c.DefaultPostForm("enable_brightness_adjust", "true") == "true",
		EnablePerspectiveCorrection: c.DefaultPostForm("enable_perspective_correction", "true") == "true",
		AutoDetectEdges:             c.DefaultPostForm("auto_detect_edges", "true") == "true",
		ContrastFactor:              1.2,
		BrightnessFactor:            10,
		SharpnessFactor:             1.1,
		DenoisingStrength:           0.3,
	}

	// 解析数值参数
	if contrast := c.PostForm("contrast_factor"); contrast != "" {
		if val, err := strconv.ParseFloat(contrast, 64); err == nil && val > 0.5 && val < 2.0 {
			config.ContrastFactor = val
		}
	}

	if brightness := c.PostForm("brightness_factor"); brightness != "" {
		if val, err := strconv.ParseFloat(brightness, 64); err == nil && val >= -100 && val <= 100 {
			config.BrightnessFactor = val
		}
	}

	if sharpness := c.PostForm("sharpness_factor"); sharpness != "" {
		if val, err := strconv.ParseFloat(sharpness, 64); err == nil && val >= 0.0 && val <= 2.0 {
			config.SharpnessFactor = val
		}
	}

	if denoising := c.PostForm("denoising_strength"); denoising != "" {
		if val, err := strconv.ParseFloat(denoising, 64); err == nil && val >= 0.0 && val <= 1.0 {
			config.DenoisingStrength = val
		}
	}

	return config
}

// buildAdaptiveConfig 根据图像质量分析结果构建自适应配置
func (r *Router) buildAdaptiveConfig(analysis *ImageQualityAnalysis) *ImageProcessorConfig {
	config := &ImageProcessorConfig{
		EnableDenoising:             true,
		EnableSharpening:            true,
		EnableContrastAdjust:        true,
		EnableBrightnessAdjust:      true,
		EnablePerspectiveCorrection: true,
		AutoDetectEdges:             true,
		ContrastFactor:              1.2,
		BrightnessFactor:            10,
		SharpnessFactor:             1.1,
		DenoisingStrength:           0.3,
	}

	if analysis != nil {
		// 根据亮度调整亮度因子
		if brightness, ok := analysis.Metrics["brightness"]; ok {
			if brightness < 80 {
				config.BrightnessFactor = 30 // 增加更多亮度
			} else if brightness > 200 {
				config.BrightnessFactor = -20 // 降低亮度
			}
		}

		// 根据对比度调整对比度因子
		if contrast, ok := analysis.Metrics["contrast"]; ok {
			if contrast < 50 {
				config.ContrastFactor = 1.5 // 增加对比度
			}
		}

		// 根据清晰度调整锐化因子
		if sharpness, ok := analysis.Metrics["sharpness"]; ok {
			if sharpness < 0.3 {
				config.SharpnessFactor = 1.5 // 增加锐化
			}
		}

		// 根据噪声水平调整降噪强度
		if noiseLevel, ok := analysis.Metrics["noise_level"]; ok {
			if noiseLevel > 0.6 {
				config.DenoisingStrength = 0.5 // 增加降噪强度
			} else if noiseLevel < 0.2 {
				config.DenoisingStrength = 0.1 // 减少降噪强度
			}
		}
	}

	return config
}
