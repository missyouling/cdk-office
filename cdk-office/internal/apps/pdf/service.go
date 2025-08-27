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
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"mime/multipart"
	"net/http"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"github.com/linux-do/cdk-office/internal/models"
)

// Service PDF处理服务
type Service struct {
	db             *gorm.DB
	stirlingPDFURL string
	httpClient     *http.Client
	config         *Config
}

// Config PDF服务配置
type Config struct {
	StirlingPDFURL    string   `json:"stirling_pdf_url"`
	Enabled           bool     `json:"enabled"`
	Timeout           int      `json:"timeout"`            // 请求超时时间（秒）
	MaxFileSize       int64    `json:"max_file_size"`      // 最大文件大小（字节）
	AllowedOperations []string `json:"allowed_operations"` // 允许的操作类型
}

// NewService 创建PDF处理服务
func NewService(config *Config, db *gorm.DB) *Service {
	if config.Timeout == 0 {
		config.Timeout = 120 // 默认2分钟超时
	}

	if config.MaxFileSize == 0 {
		config.MaxFileSize = 100 * 1024 * 1024 // 默认100MB
	}

	return &Service{
		db:             db,
		stirlingPDFURL: config.StirlingPDFURL,
		httpClient: &http.Client{
			Timeout: time.Duration(config.Timeout) * time.Second,
		},
		config: config,
	}
}

// PDFOperation PDF操作请求
type PDFOperation struct {
	Operation  string                 `json:"operation"` // merge, split, compress, rotate, watermark, convert, protect, extract-text, extract-images, repair, optimize, reorder, remove-pages
	Files      []FileInfo             `json:"files"`
	Parameters map[string]interface{} `json:"parameters"`
	UserID     string                 `json:"user_id"`
	TeamID     string                 `json:"team_id"`
}

// FileInfo 文件信息
type FileInfo struct {
	ID       string `json:"id"`
	Name     string `json:"name"`
	Size     int64  `json:"size"`
	MimeType string `json:"mime_type"`
	Path     string `json:"path"`
}

// PDFOperationResult PDF操作结果
type PDFOperationResult struct {
	Success     bool                   `json:"success"`
	ResultFile  *FileInfo              `json:"result_file,omitempty"`
	ResultFiles []FileInfo             `json:"result_files,omitempty"`
	Error       string                 `json:"error,omitempty"`
	TaskID      string                 `json:"task_id"`
	Metadata    map[string]interface{} `json:"metadata,omitempty"`
}

// MergePDFs 合并PDF文件
func (s *Service) MergePDFs(ctx context.Context, operation *PDFOperation) (*PDFOperationResult, error) {
	if !s.config.Enabled {
		return nil, fmt.Errorf("PDF processing service is disabled")
	}

	if len(operation.Files) < 2 {
		return nil, fmt.Errorf("merge operation requires at least 2 files")
	}

	// 记录操作请求
	taskRecord := &models.PDFTask{
		UserID:     operation.UserID,
		TeamID:     operation.TeamID,
		Operation:  operation.Operation,
		Status:     "processing",
		InputFiles: s.filesToJSON(operation.Files),
	}

	if err := s.db.Create(taskRecord).Error; err != nil {
		return nil, fmt.Errorf("failed to create task record: %w", err)
	}

	// 调用Stirling PDF API
	result, err := s.callStirlingPDFAPI("/api/v1/merge", operation)
	if err != nil {
		taskRecord.Status = "failed"
		taskRecord.ErrorMessage = err.Error()
		s.db.Save(taskRecord)
		return nil, fmt.Errorf("failed to merge PDFs: %w", err)
	}

	// 更新任务状态
	taskRecord.Status = "completed"
	if result.ResultFile != nil {
		taskRecord.OutputFiles = s.filesToJSON([]FileInfo{*result.ResultFile})
	}
	s.db.Save(taskRecord)

	result.TaskID = taskRecord.ID
	log.Printf("PDF merge completed: Task=%s, Files=%d", taskRecord.ID, len(operation.Files))

	return result, nil
}

// SplitPDF 拆分PDF文件
func (s *Service) SplitPDF(ctx context.Context, operation *PDFOperation) (*PDFOperationResult, error) {
	if !s.config.Enabled {
		return nil, fmt.Errorf("PDF processing service is disabled")
	}

	if len(operation.Files) != 1 {
		return nil, fmt.Errorf("split operation requires exactly 1 file")
	}

	// 记录操作请求
	taskRecord := &models.PDFTask{
		UserID:     operation.UserID,
		TeamID:     operation.TeamID,
		Operation:  operation.Operation,
		Status:     "processing",
		InputFiles: s.filesToJSON(operation.Files),
	}

	if err := s.db.Create(taskRecord).Error; err != nil {
		return nil, fmt.Errorf("failed to create task record: %w", err)
	}

	// 调用Stirling PDF API
	result, err := s.callStirlingPDFAPI("/api/v1/split", operation)
	if err != nil {
		taskRecord.Status = "failed"
		taskRecord.ErrorMessage = err.Error()
		s.db.Save(taskRecord)
		return nil, fmt.Errorf("failed to split PDF: %w", err)
	}

	// 更新任务状态
	taskRecord.Status = "completed"
	if len(result.ResultFiles) > 0 {
		taskRecord.OutputFiles = s.filesToJSON(result.ResultFiles)
	}
	s.db.Save(taskRecord)

	result.TaskID = taskRecord.ID
	log.Printf("PDF split completed: Task=%s, Output files=%d", taskRecord.ID, len(result.ResultFiles))

	return result, nil
}

// CompressPDF 压缩PDF文件
func (s *Service) CompressPDF(ctx context.Context, operation *PDFOperation) (*PDFOperationResult, error) {
	if !s.config.Enabled {
		return nil, fmt.Errorf("PDF processing service is disabled")
	}

	if len(operation.Files) != 1 {
		return nil, fmt.Errorf("compress operation requires exactly 1 file")
	}

	// 记录操作请求
	taskRecord := &models.PDFTask{
		UserID:     operation.UserID,
		TeamID:     operation.TeamID,
		Operation:  operation.Operation,
		Status:     "processing",
		InputFiles: s.filesToJSON(operation.Files),
	}

	if err := s.db.Create(taskRecord).Error; err != nil {
		return nil, fmt.Errorf("failed to create task record: %w", err)
	}

	// 调用Stirling PDF API
	result, err := s.callStirlingPDFAPI("/api/v1/compress", operation)
	if err != nil {
		taskRecord.Status = "failed"
		taskRecord.ErrorMessage = err.Error()
		s.db.Save(taskRecord)
		return nil, fmt.Errorf("failed to compress PDF: %w", err)
	}

	// 更新任务状态
	taskRecord.Status = "completed"
	if result.ResultFile != nil {
		taskRecord.OutputFiles = s.filesToJSON([]FileInfo{*result.ResultFile})
	}
	s.db.Save(taskRecord)

	result.TaskID = taskRecord.ID
	log.Printf("PDF compress completed: Task=%s", taskRecord.ID)

	return result, nil
}

// RotatePDF 旋转PDF文件
func (s *Service) RotatePDF(ctx context.Context, operation *PDFOperation) (*PDFOperationResult, error) {
	if !s.config.Enabled {
		return nil, fmt.Errorf("PDF processing service is disabled")
	}

	if len(operation.Files) != 1 {
		return nil, fmt.Errorf("rotate operation requires exactly 1 file")
	}

	// 验证旋转角度参数
	angle, ok := operation.Parameters["angle"].(float64)
	if !ok || (angle != 90 && angle != 180 && angle != 270) {
		return nil, fmt.Errorf("invalid rotation angle, must be 90, 180, or 270")
	}

	// 记录操作请求
	taskRecord := &models.PDFTask{
		UserID:     operation.UserID,
		TeamID:     operation.TeamID,
		Operation:  operation.Operation,
		Status:     "processing",
		InputFiles: s.filesToJSON(operation.Files),
		Parameters: s.parametersToJSON(operation.Parameters),
	}

	if err := s.db.Create(taskRecord).Error; err != nil {
		return nil, fmt.Errorf("failed to create task record: %w", err)
	}

	// 调用Stirling PDF API
	result, err := s.callStirlingPDFAPI("/api/v1/rotate", operation)
	if err != nil {
		taskRecord.Status = "failed"
		taskRecord.ErrorMessage = err.Error()
		s.db.Save(taskRecord)
		return nil, fmt.Errorf("failed to rotate PDF: %w", err)
	}

	// 更新任务状态
	taskRecord.Status = "completed"
	if result.ResultFile != nil {
		taskRecord.OutputFiles = s.filesToJSON([]FileInfo{*result.ResultFile})
	}
	s.db.Save(taskRecord)

	result.TaskID = taskRecord.ID
	log.Printf("PDF rotate completed: Task=%s, Angle=%.0f", taskRecord.ID, angle)

	return result, nil
}

// AddWatermark 添加水印到PDF文件
func (s *Service) AddWatermark(ctx context.Context, operation *PDFOperation) (*PDFOperationResult, error) {
	if !s.config.Enabled {
		return nil, fmt.Errorf("PDF processing service is disabled")
	}

	if len(operation.Files) != 1 {
		return nil, fmt.Errorf("watermark operation requires exactly 1 file")
	}

	// 验证水印参数
	text, hasText := operation.Parameters["text"].(string)
	if !hasText || text == "" {
		return nil, fmt.Errorf("watermark text is required")
	}

	// 记录操作请求
	taskRecord := &models.PDFTask{
		UserID:     operation.UserID,
		TeamID:     operation.TeamID,
		Operation:  operation.Operation,
		Status:     "processing",
		InputFiles: s.filesToJSON(operation.Files),
		Parameters: s.parametersToJSON(operation.Parameters),
	}

	if err := s.db.Create(taskRecord).Error; err != nil {
		return nil, fmt.Errorf("failed to create task record: %w", err)
	}

	// 调用Stirling PDF API
	result, err := s.callStirlingPDFAPI("/api/v1/watermark", operation)
	if err != nil {
		taskRecord.Status = "failed"
		taskRecord.ErrorMessage = err.Error()
		s.db.Save(taskRecord)
		return nil, fmt.Errorf("failed to add watermark: %w", err)
	}

	// 更新任务状态
	taskRecord.Status = "completed"
	if result.ResultFile != nil {
		taskRecord.OutputFiles = s.filesToJSON([]FileInfo{*result.ResultFile})
	}
	s.db.Save(taskRecord)

	result.TaskID = taskRecord.ID
	log.Printf("PDF watermark completed: Task=%s, Text=%s", taskRecord.ID, text)

	return result, nil
}

// ConvertToPDF 转换文件为PDF
func (s *Service) ConvertToPDF(ctx context.Context, operation *PDFOperation) (*PDFOperationResult, error) {
	if !s.config.Enabled {
		return nil, fmt.Errorf("PDF processing service is disabled")
	}

	if len(operation.Files) != 1 {
		return nil, fmt.Errorf("convert operation requires exactly 1 file")
	}

	// 记录操作请求
	taskRecord := &models.PDFTask{
		UserID:     operation.UserID,
		TeamID:     operation.TeamID,
		Operation:  operation.Operation,
		Status:     "processing",
		InputFiles: s.filesToJSON(operation.Files),
	}

	if err := s.db.Create(taskRecord).Error; err != nil {
		return nil, fmt.Errorf("failed to create task record: %w", err)
	}

	// 调用Stirling PDF API
	result, err := s.callStirlingPDFAPI("/api/v1/convert", operation)
	if err != nil {
		taskRecord.Status = "failed"
		taskRecord.ErrorMessage = err.Error()
		s.db.Save(taskRecord)
		return nil, fmt.Errorf("failed to convert to PDF: %w", err)
	}

	// 更新任务状态
	taskRecord.Status = "completed"
	if result.ResultFile != nil {
		taskRecord.OutputFiles = s.filesToJSON([]FileInfo{*result.ResultFile})
	}
	s.db.Save(taskRecord)

	result.TaskID = taskRecord.ID
	log.Printf("PDF convert completed: Task=%s", taskRecord.ID)

	return result, nil
}

// GetTaskStatus 获取任务状态
func (s *Service) GetTaskStatus(taskID string) (*models.PDFTask, error) {
	var task models.PDFTask
	if err := s.db.First(&task, "id = ?", taskID).Error; err != nil {
		return nil, fmt.Errorf("task not found: %w", err)
	}
	return &task, nil
}

// ListUserTasks 获取用户的PDF任务列表
func (s *Service) ListUserTasks(userID string, page, limit int) ([]models.PDFTask, error) {
	var tasks []models.PDFTask
	offset := (page - 1) * limit

	err := s.db.Where("user_id = ?", userID).
		Order("created_at DESC").
		Offset(offset).
		Limit(limit).
		Find(&tasks).Error

	return tasks, err
}

// 辅助方法

// callStirlingPDFAPI 调用Stirling PDF API
func (s *Service) callStirlingPDFAPI(endpoint string, operation *PDFOperation) (*PDFOperationResult, error) {
	url := s.stirlingPDFURL + endpoint

	// 创建multipart form数据
	var requestBody bytes.Buffer
	writer := multipart.NewWriter(&requestBody)

	// 添加文件
	for _, file := range operation.Files {
		// 这里应该从文件存储系统读取文件内容
		// 简化实现，假设我们有文件路径
		fileContent, err := s.readFile(file.Path)
		if err != nil {
			return nil, fmt.Errorf("failed to read file %s: %w", file.Name, err)
		}

		fileWriter, err := writer.CreateFormFile("files", file.Name)
		if err != nil {
			return nil, fmt.Errorf("failed to create form file: %w", err)
		}

		if _, err := io.Copy(fileWriter, bytes.NewReader(fileContent)); err != nil {
			return nil, fmt.Errorf("failed to copy file content: %w", err)
		}
	}

	// 添加参数
	for key, value := range operation.Parameters {
		writer.WriteField(key, fmt.Sprintf("%v", value))
	}

	writer.Close()

	// 创建HTTP请求
	req, err := http.NewRequest("POST", url, &requestBody)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", writer.FormDataContentType())

	// 发送请求
	resp, err := s.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("Stirling PDF API error: %d, %s", resp.StatusCode, string(body))
	}

	// 处理响应
	// 这里简化处理，实际应该根据API响应格式解析
	resultFile := &FileInfo{
		ID:   uuid.New().String(),
		Name: "result_" + operation.Files[0].Name,
		Path: "/tmp/pdf_results/" + uuid.New().String() + ".pdf",
	}

	result := &PDFOperationResult{
		Success:    true,
		ResultFile: resultFile,
	}

	return result, nil
}

// readFile 读取文件内容（简化实现）
func (s *Service) readFile(filePath string) ([]byte, error) {
	// 这里应该从实际的文件存储系统读取文件
	// 暂时返回空内容
	return []byte{}, nil
}

// filesToJSON 将文件信息转换为JSON字符串
func (s *Service) filesToJSON(files []FileInfo) string {
	data, _ := json.Marshal(files)
	return string(data)
}

// parametersToJSON 将参数转换为JSON字符串
func (s *Service) parametersToJSON(params map[string]interface{}) string {
	data, _ := json.Marshal(params)
	return string(data)
}

// HealthCheck 健康检查
func (s *Service) HealthCheck() error {
	if !s.config.Enabled {
		return fmt.Errorf("PDF service is disabled")
	}

	// 检查Stirling PDF服务是否可用
	resp, err := s.httpClient.Get(s.stirlingPDFURL + "/api/v1/info")
	if err != nil {
		return fmt.Errorf("Stirling PDF service unavailable: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("Stirling PDF service unhealthy: status %d", resp.StatusCode)
	}

	return nil
}

// ProtectPDF 为PDF文件添加密码保护
func (s *Service) ProtectPDF(ctx context.Context, operation *PDFOperation) (*PDFOperationResult, error) {
	if !s.config.Enabled {
		return nil, fmt.Errorf("PDF processing service is disabled")
	}

	if len(operation.Files) != 1 {
		return nil, fmt.Errorf("protect operation requires exactly 1 file")
	}

	// 验证密码参数
	password, hasPassword := operation.Parameters["password"].(string)
	if !hasPassword || password == "" {
		return nil, fmt.Errorf("password is required for protection")
	}

	// 记录操作请求
	taskRecord := &models.PDFTask{
		UserID:     operation.UserID,
		TeamID:     operation.TeamID,
		Operation:  operation.Operation,
		Status:     "processing",
		InputFiles: s.filesToJSON(operation.Files),
		Parameters: s.parametersToJSON(operation.Parameters),
	}

	if err := s.db.Create(taskRecord).Error; err != nil {
		return nil, fmt.Errorf("failed to create task record: %w", err)
	}

	// 调用Stirling PDF API
	result, err := s.callStirlingPDFAPI("/api/v1/add-password", operation)
	if err != nil {
		taskRecord.Status = "failed"
		taskRecord.ErrorMessage = err.Error()
		s.db.Save(taskRecord)
		return nil, fmt.Errorf("failed to protect PDF: %w", err)
	}

	// 更新任务状态
	taskRecord.Status = "completed"
	if result.ResultFile != nil {
		taskRecord.OutputFiles = s.filesToJSON([]FileInfo{*result.ResultFile})
	}
	s.db.Save(taskRecord)

	result.TaskID = taskRecord.ID
	log.Printf("PDF protect completed: Task=%s", taskRecord.ID)

	return result, nil
}

// ExtractTextFromPDF 从 PDF 提取文本
func (s *Service) ExtractTextFromPDF(ctx context.Context, operation *PDFOperation) (*PDFOperationResult, error) {
	if !s.config.Enabled {
		return nil, fmt.Errorf("PDF processing service is disabled")
	}

	if len(operation.Files) != 1 {
		return nil, fmt.Errorf("text extraction requires exactly 1 file")
	}

	// 记录操作请求
	taskRecord := &models.PDFTask{
		UserID:     operation.UserID,
		TeamID:     operation.TeamID,
		Operation:  operation.Operation,
		Status:     "processing",
		InputFiles: s.filesToJSON(operation.Files),
	}

	if err := s.db.Create(taskRecord).Error; err != nil {
		return nil, fmt.Errorf("failed to create task record: %w", err)
	}

	// 调用Stirling PDF API
	result, err := s.callStirlingPDFAPI("/api/v1/extract-text", operation)
	if err != nil {
		taskRecord.Status = "failed"
		taskRecord.ErrorMessage = err.Error()
		s.db.Save(taskRecord)
		return nil, fmt.Errorf("failed to extract text from PDF: %w", err)
	}

	// 更新任务状态
	taskRecord.Status = "completed"
	if result.ResultFile != nil {
		taskRecord.OutputFiles = s.filesToJSON([]FileInfo{*result.ResultFile})
	}
	s.db.Save(taskRecord)

	result.TaskID = taskRecord.ID
	log.Printf("PDF text extraction completed: Task=%s", taskRecord.ID)

	return result, nil
}

// ExtractImagesFromPDF 从 PDF 提取图像
func (s *Service) ExtractImagesFromPDF(ctx context.Context, operation *PDFOperation) (*PDFOperationResult, error) {
	if !s.config.Enabled {
		return nil, fmt.Errorf("PDF processing service is disabled")
	}

	if len(operation.Files) != 1 {
		return nil, fmt.Errorf("image extraction requires exactly 1 file")
	}

	// 记录操作请求
	taskRecord := &models.PDFTask{
		UserID:     operation.UserID,
		TeamID:     operation.TeamID,
		Operation:  operation.Operation,
		Status:     "processing",
		InputFiles: s.filesToJSON(operation.Files),
	}

	if err := s.db.Create(taskRecord).Error; err != nil {
		return nil, fmt.Errorf("failed to create task record: %w", err)
	}

	// 调用Stirling PDF API
	result, err := s.callStirlingPDFAPI("/api/v1/extract-images", operation)
	if err != nil {
		taskRecord.Status = "failed"
		taskRecord.ErrorMessage = err.Error()
		s.db.Save(taskRecord)
		return nil, fmt.Errorf("failed to extract images from PDF: %w", err)
	}

	// 更新任务状态
	taskRecord.Status = "completed"
	if len(result.ResultFiles) > 0 {
		taskRecord.OutputFiles = s.filesToJSON(result.ResultFiles)
	}
	s.db.Save(taskRecord)

	result.TaskID = taskRecord.ID
	log.Printf("PDF image extraction completed: Task=%s, Images=%d", taskRecord.ID, len(result.ResultFiles))

	return result, nil
}

// RepairPDF 修复PDF文件
func (s *Service) RepairPDF(ctx context.Context, operation *PDFOperation) (*PDFOperationResult, error) {
	if !s.config.Enabled {
		return nil, fmt.Errorf("PDF processing service is disabled")
	}

	if len(operation.Files) != 1 {
		return nil, fmt.Errorf("repair operation requires exactly 1 file")
	}

	// 记录操作请求
	taskRecord := &models.PDFTask{
		UserID:     operation.UserID,
		TeamID:     operation.TeamID,
		Operation:  operation.Operation,
		Status:     "processing",
		InputFiles: s.filesToJSON(operation.Files),
	}

	if err := s.db.Create(taskRecord).Error; err != nil {
		return nil, fmt.Errorf("failed to create task record: %w", err)
	}

	// 调用Stirling PDF API
	result, err := s.callStirlingPDFAPI("/api/v1/repair", operation)
	if err != nil {
		taskRecord.Status = "failed"
		taskRecord.ErrorMessage = err.Error()
		s.db.Save(taskRecord)
		return nil, fmt.Errorf("failed to repair PDF: %w", err)
	}

	// 更新任务状态
	taskRecord.Status = "completed"
	if result.ResultFile != nil {
		taskRecord.OutputFiles = s.filesToJSON([]FileInfo{*result.ResultFile})
	}
	s.db.Save(taskRecord)

	result.TaskID = taskRecord.ID
	log.Printf("PDF repair completed: Task=%s", taskRecord.ID)

	return result, nil
}

// OptimizePDF 优化PDF文件
func (s *Service) OptimizePDF(ctx context.Context, operation *PDFOperation) (*PDFOperationResult, error) {
	if !s.config.Enabled {
		return nil, fmt.Errorf("PDF processing service is disabled")
	}

	if len(operation.Files) != 1 {
		return nil, fmt.Errorf("optimize operation requires exactly 1 file")
	}

	// 记录操作请求
	taskRecord := &models.PDFTask{
		UserID:     operation.UserID,
		TeamID:     operation.TeamID,
		Operation:  operation.Operation,
		Status:     "processing",
		InputFiles: s.filesToJSON(operation.Files),
		Parameters: s.parametersToJSON(operation.Parameters),
	}

	if err := s.db.Create(taskRecord).Error; err != nil {
		return nil, fmt.Errorf("failed to create task record: %w", err)
	}

	// 调用Stirling PDF API
	result, err := s.callStirlingPDFAPI("/api/v1/optimize", operation)
	if err != nil {
		taskRecord.Status = "failed"
		taskRecord.ErrorMessage = err.Error()
		s.db.Save(taskRecord)
		return nil, fmt.Errorf("failed to optimize PDF: %w", err)
	}

	// 更新任务状态
	taskRecord.Status = "completed"
	if result.ResultFile != nil {
		taskRecord.OutputFiles = s.filesToJSON([]FileInfo{*result.ResultFile})
	}
	s.db.Save(taskRecord)

	result.TaskID = taskRecord.ID
	log.Printf("PDF optimize completed: Task=%s", taskRecord.ID)

	return result, nil
}

// ReorderPDFPages 重新排列PDF页面
func (s *Service) ReorderPDFPages(ctx context.Context, operation *PDFOperation) (*PDFOperationResult, error) {
	if !s.config.Enabled {
		return nil, fmt.Errorf("PDF processing service is disabled")
	}

	if len(operation.Files) != 1 {
		return nil, fmt.Errorf("reorder operation requires exactly 1 file")
	}

	// 验证页面顺序参数
	pageOrder, hasOrder := operation.Parameters["page_order"].(string)
	if !hasOrder || pageOrder == "" {
		return nil, fmt.Errorf("page_order parameter is required")
	}

	// 记录操作请求
	taskRecord := &models.PDFTask{
		UserID:     operation.UserID,
		TeamID:     operation.TeamID,
		Operation:  operation.Operation,
		Status:     "processing",
		InputFiles: s.filesToJSON(operation.Files),
		Parameters: s.parametersToJSON(operation.Parameters),
	}

	if err := s.db.Create(taskRecord).Error; err != nil {
		return nil, fmt.Errorf("failed to create task record: %w", err)
	}

	// 调用Stirling PDF API
	result, err := s.callStirlingPDFAPI("/api/v1/reorder-pages", operation)
	if err != nil {
		taskRecord.Status = "failed"
		taskRecord.ErrorMessage = err.Error()
		s.db.Save(taskRecord)
		return nil, fmt.Errorf("failed to reorder PDF pages: %w", err)
	}

	// 更新任务状态
	taskRecord.Status = "completed"
	if result.ResultFile != nil {
		taskRecord.OutputFiles = s.filesToJSON([]FileInfo{*result.ResultFile})
	}
	s.db.Save(taskRecord)

	result.TaskID = taskRecord.ID
	log.Printf("PDF page reorder completed: Task=%s", taskRecord.ID)

	return result, nil
}

// RemovePDFPages 删除PDF页面
func (s *Service) RemovePDFPages(ctx context.Context, operation *PDFOperation) (*PDFOperationResult, error) {
	if !s.config.Enabled {
		return nil, fmt.Errorf("PDF processing service is disabled")
	}

	if len(operation.Files) != 1 {
		return nil, fmt.Errorf("remove pages operation requires exactly 1 file")
	}

	// 验证页面参数
	pagesToRemove, hasPages := operation.Parameters["pages"].(string)
	if !hasPages || pagesToRemove == "" {
		return nil, fmt.Errorf("pages parameter is required")
	}

	// 记录操作请求
	taskRecord := &models.PDFTask{
		UserID:     operation.UserID,
		TeamID:     operation.TeamID,
		Operation:  operation.Operation,
		Status:     "processing",
		InputFiles: s.filesToJSON(operation.Files),
		Parameters: s.parametersToJSON(operation.Parameters),
	}

	if err := s.db.Create(taskRecord).Error; err != nil {
		return nil, fmt.Errorf("failed to create task record: %w", err)
	}

	// 调用Stirling PDF API
	result, err := s.callStirlingPDFAPI("/api/v1/remove-pages", operation)
	if err != nil {
		taskRecord.Status = "failed"
		taskRecord.ErrorMessage = err.Error()
		s.db.Save(taskRecord)
		return nil, fmt.Errorf("failed to remove PDF pages: %w", err)
	}

	// 更新任务状态
	taskRecord.Status = "completed"
	if result.ResultFile != nil {
		taskRecord.OutputFiles = s.filesToJSON([]FileInfo{*result.ResultFile})
	}
	s.db.Save(taskRecord)

	result.TaskID = taskRecord.ID
	log.Printf("PDF page removal completed: Task=%s", taskRecord.ID)

	return result, nil
}

// ConvertPDFToImages 将PDF转换为图像
func (s *Service) ConvertPDFToImages(ctx context.Context, operation *PDFOperation) (*PDFOperationResult, error) {
	if !s.config.Enabled {
		return nil, fmt.Errorf("PDF processing service is disabled")
	}

	if len(operation.Files) != 1 {
		return nil, fmt.Errorf("PDF to images conversion requires exactly 1 file")
	}

	// 记录操作请求
	taskRecord := &models.PDFTask{
		UserID:     operation.UserID,
		TeamID:     operation.TeamID,
		Operation:  operation.Operation,
		Status:     "processing",
		InputFiles: s.filesToJSON(operation.Files),
		Parameters: s.parametersToJSON(operation.Parameters),
	}

	if err := s.db.Create(taskRecord).Error; err != nil {
		return nil, fmt.Errorf("failed to create task record: %w", err)
	}

	// 调用Stirling PDF API
	result, err := s.callStirlingPDFAPI("/api/v1/pdf-to-img", operation)
	if err != nil {
		taskRecord.Status = "failed"
		taskRecord.ErrorMessage = err.Error()
		s.db.Save(taskRecord)
		return nil, fmt.Errorf("failed to convert PDF to images: %w", err)
	}

	// 更新任务状态
	taskRecord.Status = "completed"
	if len(result.ResultFiles) > 0 {
		taskRecord.OutputFiles = s.filesToJSON(result.ResultFiles)
	}
	s.db.Save(taskRecord)

	result.TaskID = taskRecord.ID
	log.Printf("PDF to images conversion completed: Task=%s, Images=%d", taskRecord.ID, len(result.ResultFiles))

	return result, nil
}

// GetPDFInfo 获取PDF文件信息
func (s *Service) GetPDFInfo(ctx context.Context, operation *PDFOperation) (*PDFOperationResult, error) {
	if !s.config.Enabled {
		return nil, fmt.Errorf("PDF processing service is disabled")
	}

	if len(operation.Files) != 1 {
		return nil, fmt.Errorf("PDF info requires exactly 1 file")
	}

	// 记录操作请求
	taskRecord := &models.PDFTask{
		UserID:     operation.UserID,
		TeamID:     operation.TeamID,
		Operation:  operation.Operation,
		Status:     "processing",
		InputFiles: s.filesToJSON(operation.Files),
	}

	if err := s.db.Create(taskRecord).Error; err != nil {
		return nil, fmt.Errorf("failed to create task record: %w", err)
	}

	// 调用Stirling PDF API
	result, err := s.callStirlingPDFAPI("/api/v1/pdf-info", operation)
	if err != nil {
		taskRecord.Status = "failed"
		taskRecord.ErrorMessage = err.Error()
		s.db.Save(taskRecord)
		return nil, fmt.Errorf("failed to get PDF info: %w", err)
	}

	// 更新任务状态
	taskRecord.Status = "completed"
	s.db.Save(taskRecord)

	result.TaskID = taskRecord.ID
	log.Printf("PDF info retrieval completed: Task=%s", taskRecord.ID)

	return result, nil
}
