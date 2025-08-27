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

package pdf

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/linux-do/cdk-office/internal/db"
	"gorm.io/gorm"
)

// Handler PDF处理请求处理器
type Handler struct {
	db      *gorm.DB
	service *Service
}

// NewHandler 创建PDF处理请求处理器
func NewHandler() *Handler {
	database := db.GetDB()

	// 从配置中读取PDF服务配置
	config := &Config{
		StirlingPDFURL:    "http://stirling-pdf:8080",
		Enabled:           true,
		Timeout:           120,
		MaxFileSize:       100 * 1024 * 1024, // 100MB
		AllowedOperations: []string{"merge", "split", "compress", "rotate", "watermark", "convert", "protect", "extract-text", "extract-images", "repair", "optimize", "reorder", "remove-pages", "pdf-to-images", "pdf-info"},
	}

	service := NewService(config, database)

	return &Handler{
		db:      database,
		service: service,
	}
}

// RegisterRoutes 注册路由
func (h *Handler) RegisterRoutes(router *gin.RouterGroup) {
	pdf := router.Group("/pdf")
	{
		// PDF操作接口
		pdf.POST("/merge", h.MergePDFs)
		pdf.POST("/split", h.SplitPDF)
		pdf.POST("/compress", h.CompressPDF)
		pdf.POST("/rotate", h.RotatePDF)
		pdf.POST("/watermark", h.AddWatermark)
		pdf.POST("/convert", h.ConvertToPDF)

		// 新增的PDF工具功能
		pdf.POST("/protect", h.ProtectPDF)
		pdf.POST("/extract-text", h.ExtractText)
		pdf.POST("/extract-images", h.ExtractImages)
		pdf.POST("/repair", h.RepairPDF)
		pdf.POST("/optimize", h.OptimizePDF)
		pdf.POST("/reorder-pages", h.ReorderPages)
		pdf.POST("/remove-pages", h.RemovePages)
		pdf.POST("/pdf-to-images", h.ConvertToImages)
		pdf.POST("/pdf-info", h.GetPDFInfo)

		// 任务管理接口
		pdf.GET("/tasks", h.ListTasks)
		pdf.GET("/tasks/:id", h.GetTaskStatus)
		pdf.DELETE("/tasks/:id", h.DeleteTask)

		// 文件下载接口
		pdf.GET("/download/:taskId/:fileId", h.DownloadFile)

		// 服务状态接口
		pdf.GET("/health", h.HealthCheck)
		pdf.GET("/config", h.GetConfig)
	}
}

// MergePDFs 合并PDF文件
// @Summary 合并PDF文件
// @Description 将多个PDF文件合并为一个文件
// @Tags pdf
// @Accept multipart/form-data
// @Produce json
// @Param files formData file true "PDF文件（支持多个）"
// @Success 200 {object} PDFOperationResult
// @Failure 400 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /api/pdf/merge [post]
func (h *Handler) MergePDFs(c *gin.Context) {
	userID := c.GetString("user_id")
	teamID := c.GetString("team_id")

	// 解析上传的文件
	form, err := c.MultipartForm()
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Failed to parse form data"})
		return
	}

	files := form.File["files"]
	if len(files) < 2 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "At least 2 files are required for merge"})
		return
	}

	// 构建操作请求
	fileInfos := make([]FileInfo, len(files))
	for i, file := range files {
		fileInfos[i] = FileInfo{
			Name:     file.Filename,
			Size:     file.Size,
			MimeType: file.Header.Get("Content-Type"),
		}
	}

	operation := &PDFOperation{
		Operation: "merge",
		Files:     fileInfos,
		UserID:    userID,
		TeamID:    teamID,
	}

	// 执行操作
	result, err := h.service.MergePDFs(c.Request.Context(), operation)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, result)
}

// SplitPDF 拆分PDF文件
// @Summary 拆分PDF文件
// @Description 将一个PDF文件拆分为多个文件
// @Tags pdf
// @Accept multipart/form-data
// @Produce json
// @Param file formData file true "PDF文件"
// @Param pages formData string false "页面范围（如：1-3,5,7-9）"
// @Success 200 {object} PDFOperationResult
// @Failure 400 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /api/pdf/split [post]
func (h *Handler) SplitPDF(c *gin.Context) {
	userID := c.GetString("user_id")
	teamID := c.GetString("team_id")

	// 解析上传的文件
	file, err := c.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "No file uploaded"})
		return
	}

	// 获取分页参数
	pages := c.PostForm("pages")

	// 构建操作请求
	fileInfo := FileInfo{
		Name:     file.Filename,
		Size:     file.Size,
		MimeType: file.Header.Get("Content-Type"),
	}

	parameters := make(map[string]interface{})
	if pages != "" {
		parameters["pages"] = pages
	}

	operation := &PDFOperation{
		Operation:  "split",
		Files:      []FileInfo{fileInfo},
		Parameters: parameters,
		UserID:     userID,
		TeamID:     teamID,
	}

	// 执行操作
	result, err := h.service.SplitPDF(c.Request.Context(), operation)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, result)
}

// CompressPDF 压缩PDF文件
// @Summary 压缩PDF文件
// @Description 压缩PDF文件以减小文件大小
// @Tags pdf
// @Accept multipart/form-data
// @Produce json
// @Param file formData file true "PDF文件"
// @Param quality formData string false "压缩质量（low, medium, high）"
// @Success 200 {object} PDFOperationResult
// @Failure 400 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /api/pdf/compress [post]
func (h *Handler) CompressPDF(c *gin.Context) {
	userID := c.GetString("user_id")
	teamID := c.GetString("team_id")

	// 解析上传的文件
	file, err := c.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "No file uploaded"})
		return
	}

	// 获取压缩质量参数
	quality := c.DefaultPostForm("quality", "medium")

	// 构建操作请求
	fileInfo := FileInfo{
		Name:     file.Filename,
		Size:     file.Size,
		MimeType: file.Header.Get("Content-Type"),
	}

	parameters := map[string]interface{}{
		"quality": quality,
	}

	operation := &PDFOperation{
		Operation:  "compress",
		Files:      []FileInfo{fileInfo},
		Parameters: parameters,
		UserID:     userID,
		TeamID:     teamID,
	}

	// 执行操作
	result, err := h.service.CompressPDF(c.Request.Context(), operation)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, result)
}

// RotatePDF 旋转PDF文件
// @Summary 旋转PDF文件
// @Description 旋转PDF文件的页面
// @Tags pdf
// @Accept multipart/form-data
// @Produce json
// @Param file formData file true "PDF文件"
// @Param angle formData int true "旋转角度（90, 180, 270）"
// @Param pages formData string false "页面范围（如：1-3,5,7-9）"
// @Success 200 {object} PDFOperationResult
// @Failure 400 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /api/pdf/rotate [post]
func (h *Handler) RotatePDF(c *gin.Context) {
	userID := c.GetString("user_id")
	teamID := c.GetString("team_id")

	// 解析上传的文件
	file, err := c.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "No file uploaded"})
		return
	}

	// 获取旋转角度
	angleStr := c.PostForm("angle")
	angle, err := strconv.ParseFloat(angleStr, 64)
	if err != nil || (angle != 90 && angle != 180 && angle != 270) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid rotation angle, must be 90, 180, or 270"})
		return
	}

	// 获取页面范围
	pages := c.PostForm("pages")

	// 构建操作请求
	fileInfo := FileInfo{
		Name:     file.Filename,
		Size:     file.Size,
		MimeType: file.Header.Get("Content-Type"),
	}

	parameters := map[string]interface{}{
		"angle": angle,
	}
	if pages != "" {
		parameters["pages"] = pages
	}

	operation := &PDFOperation{
		Operation:  "rotate",
		Files:      []FileInfo{fileInfo},
		Parameters: parameters,
		UserID:     userID,
		TeamID:     teamID,
	}

	// 执行操作
	result, err := h.service.RotatePDF(c.Request.Context(), operation)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, result)
}

// AddWatermark 添加水印到PDF文件
// @Summary 添加水印到PDF文件
// @Description 为PDF文件添加文字水印
// @Tags pdf
// @Accept multipart/form-data
// @Produce json
// @Param file formData file true "PDF文件"
// @Param text formData string true "水印文字"
// @Param opacity formData number false "透明度（0-1）"
// @Param fontSize formData int false "字体大小"
// @Param position formData string false "位置（center, top-left, top-right, bottom-left, bottom-right）"
// @Success 200 {object} PDFOperationResult
// @Failure 400 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /api/pdf/watermark [post]
func (h *Handler) AddWatermark(c *gin.Context) {
	userID := c.GetString("user_id")
	teamID := c.GetString("team_id")

	// 解析上传的文件
	file, err := c.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "No file uploaded"})
		return
	}

	// 获取水印参数
	text := c.PostForm("text")
	if text == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Watermark text is required"})
		return
	}

	opacity := c.DefaultPostForm("opacity", "0.5")
	fontSize := c.DefaultPostForm("fontSize", "36")
	position := c.DefaultPostForm("position", "center")

	// 构建操作请求
	fileInfo := FileInfo{
		Name:     file.Filename,
		Size:     file.Size,
		MimeType: file.Header.Get("Content-Type"),
	}

	parameters := map[string]interface{}{
		"text":     text,
		"opacity":  opacity,
		"fontSize": fontSize,
		"position": position,
	}

	operation := &PDFOperation{
		Operation:  "watermark",
		Files:      []FileInfo{fileInfo},
		Parameters: parameters,
		UserID:     userID,
		TeamID:     teamID,
	}

	// 执行操作
	result, err := h.service.AddWatermark(c.Request.Context(), operation)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, result)
}

// ConvertToPDF 转换文件为PDF
// @Summary 转换文件为PDF
// @Description 将各种格式的文件转换为PDF
// @Tags pdf
// @Accept multipart/form-data
// @Produce json
// @Param file formData file true "要转换的文件"
// @Success 200 {object} PDFOperationResult
// @Failure 400 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /api/pdf/convert [post]
func (h *Handler) ConvertToPDF(c *gin.Context) {
	userID := c.GetString("user_id")
	teamID := c.GetString("team_id")

	// 解析上传的文件
	file, err := c.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "No file uploaded"})
		return
	}

	// 构建操作请求
	fileInfo := FileInfo{
		Name:     file.Filename,
		Size:     file.Size,
		MimeType: file.Header.Get("Content-Type"),
	}

	operation := &PDFOperation{
		Operation: "convert",
		Files:     []FileInfo{fileInfo},
		UserID:    userID,
		TeamID:    teamID,
	}

	// 执行操作
	result, err := h.service.ConvertToPDF(c.Request.Context(), operation)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, result)
}

// ListTasks 获取用户的PDF任务列表
// @Summary 获取PDF任务列表
// @Description 获取当前用户的PDF处理任务列表
// @Tags pdf
// @Produce json
// @Param page query int false "页码" default(1)
// @Param limit query int false "每页数量" default(10)
// @Success 200 {object} map[string]interface{}
// @Router /api/pdf/tasks [get]
func (h *Handler) ListTasks(c *gin.Context) {
	userID := c.GetString("user_id")
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))

	tasks, err := h.service.ListUserTasks(userID, page, limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch tasks"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"tasks": tasks,
		"page":  page,
		"limit": limit,
	})
}

// GetTaskStatus 获取任务状态
// @Summary 获取PDF任务状态
// @Description 根据任务ID获取PDF处理任务的状态
// @Tags pdf
// @Produce json
// @Param id path string true "任务ID"
// @Success 200 {object} models.PDFTask
// @Failure 404 {object} map[string]interface{}
// @Router /api/pdf/tasks/{id} [get]
func (h *Handler) GetTaskStatus(c *gin.Context) {
	taskID := c.Param("id")

	task, err := h.service.GetTaskStatus(taskID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Task not found"})
		return
	}

	c.JSON(http.StatusOK, task)
}

// DeleteTask 删除任务
func (h *Handler) DeleteTask(c *gin.Context) {
	taskID := c.Param("id")
	userID := c.GetString("user_id")

	// 验证任务所有权
	task, err := h.service.GetTaskStatus(taskID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Task not found"})
		return
	}

	if task.UserID != userID {
		c.JSON(http.StatusForbidden, gin.H{"error": "Access denied"})
		return
	}

	// 删除任务记录
	if err := h.db.Delete(task).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete task"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Task deleted successfully"})
}

// DownloadFile 下载处理后的文件
func (h *Handler) DownloadFile(c *gin.Context) {
	taskID := c.Param("taskId")
	fileID := c.Param("fileId")
	userID := c.GetString("user_id")

	// 验证任务所有权
	task, err := h.service.GetTaskStatus(taskID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Task not found"})
		return
	}

	if task.UserID != userID {
		c.JSON(http.StatusForbidden, gin.H{"error": "Access denied"})
		return
	}

	// 简化实现：返回文件下载链接
	c.JSON(http.StatusOK, gin.H{
		"download_url": "/files/pdf_results/" + fileID,
		"filename":     "result.pdf",
	})
}

// HealthCheck 健康检查
// @Summary PDF服务健康检查
// @Description 检查PDF处理服务的健康状态
// @Tags pdf
// @Produce json
// @Success 200 {object} map[string]interface{}
// @Failure 503 {object} map[string]interface{}
// @Router /api/pdf/health [get]
func (h *Handler) HealthCheck(c *gin.Context) {
	err := h.service.HealthCheck()
	if err != nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{
			"status": "unhealthy",
			"error":  err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status":    "healthy",
		"timestamp": gin.H{"checked_at": gin.H{}},
	})
}

// GetConfig 获取PDF服务配置
func (h *Handler) GetConfig(c *gin.Context) {
	config := map[string]interface{}{
		"enabled":            h.service.config.Enabled,
		"max_file_size":      h.service.config.MaxFileSize,
		"allowed_operations": h.service.config.AllowedOperations,
		"timeout":            h.service.config.Timeout,
	}

	c.JSON(http.StatusOK, gin.H{
		"config": config,
	})
}

// ProtectPDF 为PDF文件添加密码保护
// @Summary 为PDF文件添加密码保护
// @Description 为PDF文件添加密码保护，防止未经授权访问
// @Tags pdf
// @Accept multipart/form-data
// @Produce json
// @Param file formData file true "PDF文件"
// @Param password formData string true "密码"
// @Param permissions formData string false "权限设置（print,copy,modify,extract）"
// @Success 200 {object} PDFOperationResult
// @Failure 400 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /api/pdf/protect [post]
func (h *Handler) ProtectPDF(c *gin.Context) {
	userID := c.GetString("user_id")
	teamID := c.GetString("team_id")

	// 解析上传的文件
	file, err := c.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "No file uploaded"})
		return
	}

	// 获取密码参数
	password := c.PostForm("password")
	if password == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Password is required"})
		return
	}

	permissions := c.DefaultPostForm("permissions", "print,copy")

	// 构建操作请求
	fileInfo := FileInfo{
		Name:     file.Filename,
		Size:     file.Size,
		MimeType: file.Header.Get("Content-Type"),
	}

	parameters := map[string]interface{}{
		"password":    password,
		"permissions": permissions,
	}

	operation := &PDFOperation{
		Operation:  "protect",
		Files:      []FileInfo{fileInfo},
		Parameters: parameters,
		UserID:     userID,
		TeamID:     teamID,
	}

	// 执行操作
	result, err := h.service.ProtectPDF(c.Request.Context(), operation)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, result)
}

// ExtractText 从 PDF 提取文本
// @Summary 从 PDF 提取文本
// @Description 从 PDF 文件中提取文本内容
// @Tags pdf
// @Accept multipart/form-data
// @Produce json
// @Param file formData file true "PDF文件"
// @Param pages formData string false "页面范围（如：1-3,5,7-9）"
// @Success 200 {object} PDFOperationResult
// @Failure 400 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /api/pdf/extract-text [post]
func (h *Handler) ExtractText(c *gin.Context) {
	userID := c.GetString("user_id")
	teamID := c.GetString("team_id")

	// 解析上传的文件
	file, err := c.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "No file uploaded"})
		return
	}

	// 获取页面参数
	pages := c.PostForm("pages")

	// 构建操作请求
	fileInfo := FileInfo{
		Name:     file.Filename,
		Size:     file.Size,
		MimeType: file.Header.Get("Content-Type"),
	}

	parameters := make(map[string]interface{})
	if pages != "" {
		parameters["pages"] = pages
	}

	operation := &PDFOperation{
		Operation:  "extract-text",
		Files:      []FileInfo{fileInfo},
		Parameters: parameters,
		UserID:     userID,
		TeamID:     teamID,
	}

	// 执行操作
	result, err := h.service.ExtractTextFromPDF(c.Request.Context(), operation)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, result)
}

// ExtractImages 从 PDF 提取图像
// @Summary 从 PDF 提取图像
// @Description 从 PDF 文件中提取所有图像
// @Tags pdf
// @Accept multipart/form-data
// @Produce json
// @Param file formData file true "PDF文件"
// @Param format formData string false "图像格式（png, jpg, gif）"
// @Success 200 {object} PDFOperationResult
// @Failure 400 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /api/pdf/extract-images [post]
func (h *Handler) ExtractImages(c *gin.Context) {
	userID := c.GetString("user_id")
	teamID := c.GetString("team_id")

	// 解析上传的文件
	file, err := c.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "No file uploaded"})
		return
	}

	// 获取图像格式参数
	format := c.DefaultPostForm("format", "png")

	// 构建操作请求
	fileInfo := FileInfo{
		Name:     file.Filename,
		Size:     file.Size,
		MimeType: file.Header.Get("Content-Type"),
	}

	parameters := map[string]interface{}{
		"format": format,
	}

	operation := &PDFOperation{
		Operation:  "extract-images",
		Files:      []FileInfo{fileInfo},
		Parameters: parameters,
		UserID:     userID,
		TeamID:     teamID,
	}

	// 执行操作
	result, err := h.service.ExtractImagesFromPDF(c.Request.Context(), operation)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, result)
}

// RepairPDF 修复PDF文件
// @Summary 修复PDF文件
// @Description 修复损坏或格式错误的PDF文件
// @Tags pdf
// @Accept multipart/form-data
// @Produce json
// @Param file formData file true "PDF文件"
// @Success 200 {object} PDFOperationResult
// @Failure 400 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /api/pdf/repair [post]
func (h *Handler) RepairPDF(c *gin.Context) {
	userID := c.GetString("user_id")
	teamID := c.GetString("team_id")

	// 解析上传的文件
	file, err := c.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "No file uploaded"})
		return
	}

	// 构建操作请求
	fileInfo := FileInfo{
		Name:     file.Filename,
		Size:     file.Size,
		MimeType: file.Header.Get("Content-Type"),
	}

	operation := &PDFOperation{
		Operation: "repair",
		Files:     []FileInfo{fileInfo},
		UserID:    userID,
		TeamID:    teamID,
	}

	// 执行操作
	result, err := h.service.RepairPDF(c.Request.Context(), operation)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, result)
}

// OptimizePDF 优化PDF文件
// @Summary 优化PDF文件
// @Description 优化PDF文件以提高性能和减小文件大小
// @Tags pdf
// @Accept multipart/form-data
// @Produce json
// @Param file formData file true "PDF文件"
// @Param optimization_level formData string false "优化级别（low, medium, high）"
// @Success 200 {object} PDFOperationResult
// @Failure 400 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /api/pdf/optimize [post]
func (h *Handler) OptimizePDF(c *gin.Context) {
	userID := c.GetString("user_id")
	teamID := c.GetString("team_id")

	// 解析上传的文件
	file, err := c.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "No file uploaded"})
		return
	}

	// 获取优化级别参数
	optimizationLevel := c.DefaultPostForm("optimization_level", "medium")

	// 构建操作请求
	fileInfo := FileInfo{
		Name:     file.Filename,
		Size:     file.Size,
		MimeType: file.Header.Get("Content-Type"),
	}

	parameters := map[string]interface{}{
		"optimization_level": optimizationLevel,
	}

	operation := &PDFOperation{
		Operation:  "optimize",
		Files:      []FileInfo{fileInfo},
		Parameters: parameters,
		UserID:     userID,
		TeamID:     teamID,
	}

	// 执行操作
	result, err := h.service.OptimizePDF(c.Request.Context(), operation)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, result)
}

// ReorderPages 重新排列PDF页面
// @Summary 重新排列PDF页面
// @Description 重新排列PDF文件中的页面顺序
// @Tags pdf
// @Accept multipart/form-data
// @Produce json
// @Param file formData file true "PDF文件"
// @Param page_order formData string true "新的页面顺序（如：3,1,4,2）"
// @Success 200 {object} PDFOperationResult
// @Failure 400 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /api/pdf/reorder-pages [post]
func (h *Handler) ReorderPages(c *gin.Context) {
	userID := c.GetString("user_id")
	teamID := c.GetString("team_id")

	// 解析上传的文件
	file, err := c.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "No file uploaded"})
		return
	}

	// 获取页面顺序参数
	pageOrder := c.PostForm("page_order")
	if pageOrder == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Page order is required"})
		return
	}

	// 构建操作请求
	fileInfo := FileInfo{
		Name:     file.Filename,
		Size:     file.Size,
		MimeType: file.Header.Get("Content-Type"),
	}

	parameters := map[string]interface{}{
		"page_order": pageOrder,
	}

	operation := &PDFOperation{
		Operation:  "reorder",
		Files:      []FileInfo{fileInfo},
		Parameters: parameters,
		UserID:     userID,
		TeamID:     teamID,
	}

	// 执行操作
	result, err := h.service.ReorderPDFPages(c.Request.Context(), operation)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, result)
}

// RemovePages 删除PDF页面
// @Summary 删除PDF页面
// @Description 从 PDF 文件中删除指定的页面
// @Tags pdf
// @Accept multipart/form-data
// @Produce json
// @Param file formData file true "PDF文件"
// @Param pages formData string true "要删除的页面（如：1,3,5-7）"
// @Success 200 {object} PDFOperationResult
// @Failure 400 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /api/pdf/remove-pages [post]
func (h *Handler) RemovePages(c *gin.Context) {
	userID := c.GetString("user_id")
	teamID := c.GetString("team_id")

	// 解析上传的文件
	file, err := c.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "No file uploaded"})
		return
	}

	// 获取要删除的页面参数
	pages := c.PostForm("pages")
	if pages == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Pages to remove are required"})
		return
	}

	// 构建操作请求
	fileInfo := FileInfo{
		Name:     file.Filename,
		Size:     file.Size,
		MimeType: file.Header.Get("Content-Type"),
	}

	parameters := map[string]interface{}{
		"pages": pages,
	}

	operation := &PDFOperation{
		Operation:  "remove-pages",
		Files:      []FileInfo{fileInfo},
		Parameters: parameters,
		UserID:     userID,
		TeamID:     teamID,
	}

	// 执行操作
	result, err := h.service.RemovePDFPages(c.Request.Context(), operation)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, result)
}

// ConvertToImages 将PDF转换为图像
// @Summary 将PDF转换为图像
// @Description 将PDF文件的每一页转换为图像文件
// @Tags pdf
// @Accept multipart/form-data
// @Produce json
// @Param file formData file true "PDF文件"
// @Param format formData string false "图像格式（png, jpg）"
// @Param dpi formData int false "分辨率DPI"
// @Param quality formData int false "图像质量（1-100）"
// @Success 200 {object} PDFOperationResult
// @Failure 400 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /api/pdf/pdf-to-images [post]
func (h *Handler) ConvertToImages(c *gin.Context) {
	userID := c.GetString("user_id")
	teamID := c.GetString("team_id")

	// 解析上传的文件
	file, err := c.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "No file uploaded"})
		return
	}

	// 获取转换参数
	format := c.DefaultPostForm("format", "png")
	dpi := c.DefaultPostForm("dpi", "150")
	quality := c.DefaultPostForm("quality", "95")

	// 构建操作请求
	fileInfo := FileInfo{
		Name:     file.Filename,
		Size:     file.Size,
		MimeType: file.Header.Get("Content-Type"),
	}

	parameters := map[string]interface{}{
		"format":  format,
		"dpi":     dpi,
		"quality": quality,
	}

	operation := &PDFOperation{
		Operation:  "pdf-to-images",
		Files:      []FileInfo{fileInfo},
		Parameters: parameters,
		UserID:     userID,
		TeamID:     teamID,
	}

	// 执行操作
	result, err := h.service.ConvertPDFToImages(c.Request.Context(), operation)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, result)
}

// GetPDFInfo 获取PDF文件信息
// @Summary 获取PDF文件信息
// @Description 获取PDF文件的详细信息（页数、尺寸、作者等）
// @Tags pdf
// @Accept multipart/form-data
// @Produce json
// @Param file formData file true "PDF文件"
// @Success 200 {object} PDFOperationResult
// @Failure 400 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /api/pdf/pdf-info [post]
func (h *Handler) GetPDFInfo(c *gin.Context) {
	userID := c.GetString("user_id")
	teamID := c.GetString("team_id")

	// 解析上传的文件
	file, err := c.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "No file uploaded"})
		return
	}

	// 构建操作请求
	fileInfo := FileInfo{
		Name:     file.Filename,
		Size:     file.Size,
		MimeType: file.Header.Get("Content-Type"),
	}

	operation := &PDFOperation{
		Operation: "pdf-info",
		Files:     []FileInfo{fileInfo},
		UserID:    userID,
		TeamID:    teamID,
	}

	// 执行操作
	result, err := h.service.GetPDFInfo(c.Request.Context(), operation)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, result)
}
