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

package knowledge

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/linux-do/cdk-office/internal/models"
	"gorm.io/gorm"
)

// WeChatService 微信聊天记录服务
type WeChatService struct {
	db        *gorm.DB
	uploadDir string
}

// NewWeChatService 创建微信聊天记录服务
func NewWeChatService(db *gorm.DB) *WeChatService {
	uploadDir := "uploads/wechat"
	if err := os.MkdirAll(uploadDir, 0755); err != nil {
		log.Printf("Failed to create upload directory: %v", err)
	}

	return &WeChatService{
		db:        db,
		uploadDir: uploadDir,
	}
}

// ProcessWeChatUpload 处理微信聊天记录上传
func (s *WeChatService) ProcessWeChatUpload(ctx context.Context, req *WeChatUploadRequest) (*WeChatUploadResponse, error) {
	response := &WeChatUploadResponse{
		ProcessedCount: 0,
		FailedCount:    0,
		Records:        []models.WeChatRecord{},
		Errors:         []WeChatProcessError{},
	}

	// 设置默认处理配置
	if req.ProcessConfig == nil {
		req.ProcessConfig = &WeChatProcessConfig{
			EnableOCR:         true,
			EnableAutoArchive: false,
			ExtractKeywords:   true,
			GroupBySession:    true,
		}
	}

	for _, recordData := range req.Records {
		record, err := s.processWeChatRecord(ctx, req.UserID, req.SessionName, recordData, req.ProcessConfig)
		if err != nil {
			response.FailedCount++
			response.Errors = append(response.Errors, WeChatProcessError{
				MessageID: recordData.MessageID,
				Error:     err.Error(),
			})
			continue
		}

		response.ProcessedCount++
		response.Records = append(response.Records, *record)
	}

	log.Printf("Processed WeChat records: %d success, %d failed", response.ProcessedCount, response.FailedCount)
	return response, nil
}

// processWeChatRecord 处理单条微信聊天记录
func (s *WeChatService) processWeChatRecord(ctx context.Context, userID, sessionName string, data WeChatRecordData, config *WeChatProcessConfig) (*models.WeChatRecord, error) {
	// 创建基础记录
	record := &models.WeChatRecord{
		UserID:        userID,
		SessionName:   sessionName,
		MessageType:   data.MessageType,
		MessageID:     data.MessageID,
		SenderName:    data.SenderName,
		SenderID:      data.SenderID,
		Content:       data.Content,
		ProcessStatus: "processing",
	}

	// 解析消息时间
	if data.MessageTime != "" {
		if messageTime, err := time.Parse("2006-01-02 15:04:05", data.MessageTime); err == nil {
			record.MessageTime = messageTime
		} else if messageTime, err := time.Parse(time.RFC3339, data.MessageTime); err == nil {
			record.MessageTime = messageTime
		} else {
			record.MessageTime = time.Now()
		}
	} else {
		record.MessageTime = time.Now()
	}

	// 处理文件数据
	if data.FileData != "" && data.FileName != "" {
		originalPath, processedPath, err := s.saveFileData(userID, data.FileName, data.FileData)
		if err != nil {
			return nil, fmt.Errorf("failed to save file: %w", err)
		}
		record.OriginalFile = originalPath
		record.ProcessedFile = processedPath
	}

	// 保存基础记录到数据库
	if err := s.db.WithContext(ctx).Create(record).Error; err != nil {
		return nil, fmt.Errorf("failed to create wechat record: %w", err)
	}

	// 异步处理OCR和其他增强功能
	go s.enhanceWeChatRecord(record.ID, config)

	record.ProcessStatus = "completed"
	s.db.WithContext(ctx).Model(record).Update("process_status", "completed")

	return record, nil
}

// saveFileData 保存文件数据
func (s *WeChatService) saveFileData(userID, fileName, fileData string) (string, string, error) {
	// 解码base64数据
	data, err := base64.StdEncoding.DecodeString(fileData)
	if err != nil {
		return "", "", fmt.Errorf("failed to decode file data: %w", err)
	}

	// 创建用户目录
	userDir := filepath.Join(s.uploadDir, userID)
	if err := os.MkdirAll(userDir, 0755); err != nil {
		return "", "", fmt.Errorf("failed to create user directory: %w", err)
	}

	// 生成文件路径
	timestamp := time.Now().Format("20060102_150405")
	originalPath := filepath.Join(userDir, fmt.Sprintf("%s_%s", timestamp, fileName))
	processedPath := filepath.Join(userDir, fmt.Sprintf("%s_processed_%s", timestamp, fileName))

	// 保存原始文件
	if err := os.WriteFile(originalPath, data, 0644); err != nil {
		return "", "", fmt.Errorf("failed to save original file: %w", err)
	}

	// 简单复制作为处理后文件（实际应用中这里会有图像处理逻辑）
	if err := os.WriteFile(processedPath, data, 0644); err != nil {
		return "", "", fmt.Errorf("failed to save processed file: %w", err)
	}

	return originalPath, processedPath, nil
}

// enhanceWeChatRecord 增强微信聊天记录（OCR、关键词提取等）
func (s *WeChatService) enhanceWeChatRecord(recordID string, config *WeChatProcessConfig) {
	ctx := context.Background()

	// 获取记录
	var record models.WeChatRecord
	if err := s.db.WithContext(ctx).Where("id = ?", recordID).First(&record).Error; err != nil {
		log.Printf("Failed to get wechat record for enhancement: %v", err)
		return
	}

	updateData := make(map[string]interface{})
	extractedInfo := make(map[string]interface{})

	// 如果是图片类型且启用了OCR
	if record.MessageType == "image" && config.EnableOCR && record.ProcessedFile != "" {
		ocrText, err := s.performOCR(record.ProcessedFile)
		if err != nil {
			log.Printf("OCR failed for record %s: %v", recordID, err)
		} else {
			updateData["ocr_text"] = ocrText
			extractedInfo["ocr_confidence"] = 0.85 // 模拟置信度
		}
	}

	// 关键词提取
	if config.ExtractKeywords {
		keywords := s.extractKeywords(record.Content + " " + record.OCRText)
		extractedInfo["keywords"] = keywords
	}

	// 内容分析
	contentAnalysis := s.analyzeContent(record.Content, record.OCRText)
	extractedInfo["content_analysis"] = contentAnalysis

	// 保存提取的信息
	if len(extractedInfo) > 0 {
		extractedJSON, _ := json.Marshal(extractedInfo)
		updateData["extracted_info"] = string(extractedJSON)
	}

	// 更新记录
	if len(updateData) > 0 {
		if err := s.db.WithContext(ctx).Model(&record).Updates(updateData).Error; err != nil {
			log.Printf("Failed to update wechat record enhancement: %v", err)
		}
	}

	log.Printf("Enhanced wechat record: %s", recordID)
}

// performOCR 执行OCR识别
func (s *WeChatService) performOCR(filePath string) (string, error) {
	// 这里应该调用实际的OCR服务
	// 暂时返回模拟数据
	return "OCR识别的文本内容", nil
}

// extractKeywords 提取关键词
func (s *WeChatService) extractKeywords(text string) []string {
	if text == "" {
		return []string{}
	}

	// 简单的关键词提取逻辑
	words := strings.Fields(text)
	keywordMap := make(map[string]int)

	for _, word := range words {
		word = strings.ToLower(strings.TrimSpace(word))
		if len(word) > 2 { // 忽略太短的词
			keywordMap[word]++
		}
	}

	var keywords []string
	for keyword, count := range keywordMap {
		if count >= 2 { // 出现2次以上的词作为关键词
			keywords = append(keywords, keyword)
		}
	}

	// 限制关键词数量
	if len(keywords) > 10 {
		keywords = keywords[:10]
	}

	return keywords
}

// analyzeContent 分析内容
func (s *WeChatService) analyzeContent(content, ocrText string) map[string]interface{} {
	analysis := make(map[string]interface{})

	fullText := content + " " + ocrText
	analysis["character_count"] = len(fullText)
	analysis["word_count"] = len(strings.Fields(fullText))

	// 简单的内容类型判断
	if strings.Contains(fullText, "@") && strings.Contains(fullText, ".") {
		analysis["contains_email"] = true
	}
	if strings.Contains(fullText, "http://") || strings.Contains(fullText, "https://") {
		analysis["contains_url"] = true
	}
	if strings.Contains(fullText, "电话") || strings.Contains(fullText, "手机") {
		analysis["contains_phone"] = true
	}

	// 情感分析（简化）
	positiveWords := []string{"好", "棒", "不错", "喜欢", "满意"}
	negativeWords := []string{"不好", "差", "讨厌", "不满意", "糟糕"}

	positiveCount := 0
	negativeCount := 0

	for _, word := range positiveWords {
		positiveCount += strings.Count(fullText, word)
	}
	for _, word := range negativeWords {
		negativeCount += strings.Count(fullText, word)
	}

	if positiveCount > negativeCount {
		analysis["sentiment"] = "positive"
	} else if negativeCount > positiveCount {
		analysis["sentiment"] = "negative"
	} else {
		analysis["sentiment"] = "neutral"
	}

	return analysis
}

// ListWeChatRecords 列出微信聊天记录
func (s *WeChatService) ListWeChatRecords(ctx context.Context, req *ListWeChatRecordsRequest) (*ListWeChatRecordsResponse, error) {
	query := s.db.WithContext(ctx).Model(&models.WeChatRecord{}).Where("user_id = ?", req.UserID)

	// 筛选条件
	if req.SessionName != "" {
		query = query.Where("session_name = ?", req.SessionName)
	}
	if req.MessageType != "" {
		query = query.Where("message_type = ?", req.MessageType)
	}
	if req.StartDate != "" {
		if startDate, err := time.Parse("2006-01-02", req.StartDate); err == nil {
			query = query.Where("message_time >= ?", startDate)
		}
	}
	if req.EndDate != "" {
		if endDate, err := time.Parse("2006-01-02", req.EndDate); err == nil {
			query = query.Where("message_time <= ?", endDate.Add(24*time.Hour))
		}
	}
	if req.Keyword != "" {
		keyword := "%" + req.Keyword + "%"
		query = query.Where("content ILIKE ? OR ocr_text ILIKE ?", keyword, keyword)
	}

	// 计算总数
	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, fmt.Errorf("failed to count wechat records: %w", err)
	}

	// 分页
	offset := (req.Page - 1) * req.PageSize
	query = query.Offset(offset).Limit(req.PageSize)

	// 排序
	query = query.Order("message_time DESC")

	var records []models.WeChatRecord
	if err := query.Find(&records).Error; err != nil {
		return nil, fmt.Errorf("failed to list wechat records: %w", err)
	}

	return &ListWeChatRecordsResponse{
		Records:  records,
		Total:    total,
		Page:     req.Page,
		PageSize: req.PageSize,
	}, nil
}

// GetWeChatRecord 获取微信聊天记录详情
func (s *WeChatService) GetWeChatRecord(ctx context.Context, userID, recordID string) (*models.WeChatRecord, error) {
	var record models.WeChatRecord
	if err := s.db.WithContext(ctx).Where("id = ? AND user_id = ?", recordID, userID).First(&record).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("wechat record not found")
		}
		return nil, fmt.Errorf("failed to get wechat record: %w", err)
	}

	return &record, nil
}

// DeleteWeChatRecord 删除微信聊天记录
func (s *WeChatService) DeleteWeChatRecord(ctx context.Context, userID, recordID string) error {
	// 先获取记录以删除相关文件
	var record models.WeChatRecord
	if err := s.db.WithContext(ctx).Where("id = ? AND user_id = ?", recordID, userID).First(&record).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return fmt.Errorf("wechat record not found")
		}
		return fmt.Errorf("failed to find wechat record: %w", err)
	}

	// 删除相关文件
	if record.OriginalFile != "" {
		if err := os.Remove(record.OriginalFile); err != nil {
			log.Printf("Failed to delete original file %s: %v", record.OriginalFile, err)
		}
	}
	if record.ProcessedFile != "" {
		if err := os.Remove(record.ProcessedFile); err != nil {
			log.Printf("Failed to delete processed file %s: %v", record.ProcessedFile, err)
		}
	}

	// 删除数据库记录
	if err := s.db.WithContext(ctx).Delete(&record).Error; err != nil {
		return fmt.Errorf("failed to delete wechat record: %w", err)
	}

	log.Printf("Deleted wechat record: %s", recordID)
	return nil
}

// ArchiveWeChatRecord 归档微信聊天记录到个人知识库
func (s *WeChatService) ArchiveWeChatRecord(ctx context.Context, req *ArchiveWeChatRecordRequest) (*models.PersonalKnowledgeBase, error) {
	// 获取微信记录
	var record models.WeChatRecord
	if err := s.db.WithContext(ctx).Where("id = ? AND user_id = ?", req.RecordID, req.UserID).First(&record).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("wechat record not found")
		}
		return nil, fmt.Errorf("failed to get wechat record: %w", err)
	}

	// 构建知识库内容
	content := s.buildKnowledgeContent(&record)

	// 创建个人知识库记录
	knowledge := &models.PersonalKnowledgeBase{
		UserID:      req.UserID,
		Title:       req.Title,
		Description: req.Description,
		Content:     content,
		ContentType: "markdown",
		Tags:        req.Tags,
		Category:    req.Category,
		Privacy:     "private",
		SourceType:  "wechat",
		SourceData:  fmt.Sprintf(`{"wechat_record_id": "%s"}`, req.RecordID),
	}

	if err := s.db.WithContext(ctx).Create(knowledge).Error; err != nil {
		return nil, fmt.Errorf("failed to create knowledge from wechat record: %w", err)
	}

	// 更新微信记录的归档状态
	updateData := map[string]interface{}{
		"is_archived": true,
		"archived_to": knowledge.ID,
	}
	if err := s.db.WithContext(ctx).Model(&record).Updates(updateData).Error; err != nil {
		log.Printf("Failed to update wechat record archive status: %v", err)
	}

	log.Printf("Archived wechat record %s to knowledge %s", req.RecordID, knowledge.ID)
	return knowledge, nil
}

// buildKnowledgeContent 构建知识库内容
func (s *WeChatService) buildKnowledgeContent(record *models.WeChatRecord) string {
	var content strings.Builder

	content.WriteString(fmt.Sprintf("# 微信聊天记录\n\n"))
	content.WriteString(fmt.Sprintf("**会话名称:** %s\n\n", record.SessionName))
	content.WriteString(fmt.Sprintf("**发送者:** %s\n\n", record.SenderName))
	content.WriteString(fmt.Sprintf("**消息类型:** %s\n\n", record.MessageType))
	content.WriteString(fmt.Sprintf("**发送时间:** %s\n\n", record.MessageTime.Format("2006-01-02 15:04:05")))

	if record.Content != "" {
		content.WriteString(fmt.Sprintf("## 消息内容\n\n%s\n\n", record.Content))
	}

	if record.OCRText != "" {
		content.WriteString(fmt.Sprintf("## OCR识别内容\n\n%s\n\n", record.OCRText))
	}

	if record.ExtractedInfo != "" {
		content.WriteString(fmt.Sprintf("## 提取信息\n\n```json\n%s\n```\n\n", record.ExtractedInfo))
	}

	return content.String()
}

// GetWeChatStatistics 获取微信聊天记录统计
func (s *WeChatService) GetWeChatStatistics(ctx context.Context, userID string) (*WeChatStatistics, error) {
	var stats WeChatStatistics

	// 总记录数
	if err := s.db.WithContext(ctx).Model(&models.WeChatRecord{}).Where("user_id = ?", userID).Count(&stats.TotalRecords).Error; err != nil {
		return nil, fmt.Errorf("failed to count total records: %w", err)
	}

	// 按消息类型统计
	var typeStats []WeChatTypeStat
	if err := s.db.WithContext(ctx).Model(&models.WeChatRecord{}).
		Select("message_type, COUNT(*) as count").
		Where("user_id = ?", userID).
		Group("message_type").
		Find(&typeStats).Error; err != nil {
		return nil, fmt.Errorf("failed to get type statistics: %w", err)
	}
	stats.ByType = typeStats

	// 已归档数量
	if err := s.db.WithContext(ctx).Model(&models.WeChatRecord{}).Where("user_id = ? AND is_archived = true", userID).Count(&stats.ArchivedRecords).Error; err != nil {
		return nil, fmt.Errorf("failed to count archived records: %w", err)
	}

	// 本周新增
	weekAgo := time.Now().AddDate(0, 0, -7)
	if err := s.db.WithContext(ctx).Model(&models.WeChatRecord{}).Where("user_id = ? AND created_at >= ?", userID, weekAgo).Count(&stats.WeeklyAdded).Error; err != nil {
		return nil, fmt.Errorf("failed to count weekly records: %w", err)
	}

	return &stats, nil
}
