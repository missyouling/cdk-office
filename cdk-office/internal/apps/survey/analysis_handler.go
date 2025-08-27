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

package survey

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
	"gorm.io/datatypes"

	"cdk-office/internal/models"
)

// TriggerAnalysis 触发问卷分析
func (h *SurveyHandler) TriggerAnalysis(c *gin.Context) {
	surveyID := c.Param("id")
	userID := c.GetString("user_id")

	// 检查分析权限
	if !h.hasAnalyzePermission(surveyID, userID) {
		c.JSON(http.StatusForbidden, gin.H{"error": "No analyze permission"})
		return
	}

	var req TriggerAnalysisRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request data", "details": err.Error()})
		return
	}

	// 检查问卷是否存在
	var survey models.Survey
	if err := h.db.Where("survey_id = ?", surveyID).First(&survey).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "Survey not found"})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get survey"})
		}
		return
	}

	// 创建分析记录
	analysis := &models.SurveyAnalysis{
		SurveyID:       surveyID,
		AnalysisType:   req.AnalysisType,
		DifyWorkflowID: req.DifyWorkflowID,
		Status:         "pending",
	}

	if err := h.db.Create(analysis).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create analysis"})
		return
	}

	// 根据分析类型执行不同的分析
	switch req.AnalysisType {
	case "basic":
		go h.performBasicAnalysis(analysis)
	case "ai":
		if h.difyService != nil {
			go h.performAIAnalysis(analysis)
		} else {
			h.db.Model(analysis).Updates(map[string]interface{}{
				"status":        "failed",
				"error_message": "Dify service not configured",
			})
		}
	case "custom":
		go h.performCustomAnalysis(analysis, req.DifyWorkflowID)
	default:
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid analysis type"})
		return
	}

	c.JSON(http.StatusAccepted, AnalysisResponse{
		ID:             analysis.ID,
		SurveyID:       analysis.SurveyID,
		AnalysisType:   analysis.AnalysisType,
		DifyWorkflowID: analysis.DifyWorkflowID,
		Status:         analysis.Status,
		CreatedAt:      analysis.CreatedAt,
		UpdatedAt:      analysis.UpdatedAt,
	})
}

// GetAnalysis 获取问卷分析结果
func (h *SurveyHandler) GetAnalysis(c *gin.Context) {
	surveyID := c.Param("id")
	userID := c.GetString("user_id")

	// 检查分析权限
	if !h.hasAnalyzePermission(surveyID, userID) {
		c.JSON(http.StatusForbidden, gin.H{"error": "No analyze permission"})
		return
	}

	analysisType := c.DefaultQuery("type", "")
	
	query := h.db.Where("survey_id = ?", surveyID)
	if analysisType != "" {
		query = query.Where("analysis_type = ?", analysisType)
	}

	var analyses []models.SurveyAnalysis
	if err := query.Order("created_at DESC").Find(&analyses).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get analysis"})
		return
	}

	// 转换为响应格式
	analysisResponses := make([]AnalysisResponse, len(analyses))
	for i, analysis := range analyses {
		analysisResponses[i] = AnalysisResponse{
			ID:             analysis.ID,
			SurveyID:       analysis.SurveyID,
			AnalysisType:   analysis.AnalysisType,
			ResultData:     analysis.ResultData,
			DifyWorkflowID: analysis.DifyWorkflowID,
			RunID:          analysis.RunID,
			Status:         analysis.Status,
			ErrorMessage:   analysis.ErrorMessage,
			CreatedAt:      analysis.CreatedAt,
			UpdatedAt:      analysis.UpdatedAt,
		}
	}

	c.JSON(http.StatusOK, analysisResponses)
}

// ExportData 导出问卷数据
func (h *SurveyHandler) ExportData(c *gin.Context) {
	surveyID := c.Param("id")
	userID := c.GetString("user_id")

	// 检查导出权限
	if !h.hasExportPermission(surveyID, userID) {
		c.JSON(http.StatusForbidden, gin.H{"error": "No export permission"})
		return
	}

	var req ExportRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		// 如果没有请求体，使用默认设置
		req = ExportRequest{
			Format:      c.DefaultQuery("format", "csv"),
			IncludeUser: c.DefaultQuery("include_user", "false") == "true",
		}
	}

	// 获取问卷信息
	var survey models.Survey
	if err := h.db.Where("survey_id = ?", surveyID).First(&survey).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Survey not found"})
		return
	}

	// 获取响应数据
	query := h.db.Where("survey_id = ?", surveyID)
	
	// 应用日期过滤
	if req.DateFrom != "" {
		if dateFrom, err := time.Parse("2006-01-02", req.DateFrom); err == nil {
			query = query.Where("created_at >= ?", dateFrom)
		}
	}
	if req.DateTo != "" {
		if dateTo, err := time.Parse("2006-01-02", req.DateTo); err == nil {
			query = query.Where("created_at <= ?", dateTo.Add(24*time.Hour))
		}
	}

	var responses []models.SurveyResponse
	if err := query.Find(&responses).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get responses"})
		return
	}

	// 根据格式导出数据
	switch strings.ToLower(req.Format) {
	case "csv":
		h.exportCSV(c, survey, responses, req)
	case "xlsx":
		h.exportExcel(c, survey, responses, req)
	case "json":
		h.exportJSON(c, survey, responses, req)
	default:
		c.JSON(http.StatusBadRequest, gin.H{"error": "Unsupported export format"})
	}
}

// performBasicAnalysis 执行基础分析
func (h *SurveyHandler) performBasicAnalysis(analysis *models.SurveyAnalysis) {
	h.db.Model(analysis).Update("status", "running")

	// 获取问卷响应
	var responses []models.SurveyResponse
	if err := h.db.Where("survey_id = ?", analysis.SurveyID).Find(&responses).Error; err != nil {
		h.db.Model(analysis).Updates(map[string]interface{}{
			"status":        "failed",
			"error_message": fmt.Sprintf("Failed to get responses: %v", err),
		})
		return
	}

	// 计算基础统计
	totalResponses := len(responses)
	var totalTimeSpent int
	for _, resp := range responses {
		totalTimeSpent += resp.TimeSpent
	}

	avgTimeSpent := 0
	if totalResponses > 0 {
		avgTimeSpent = totalTimeSpent / totalResponses
	}

	// 分析响应数据
	questionStats := h.analyzeQuestionStats(responses)

	result := map[string]interface{}{
		"total_responses":     totalResponses,
		"avg_time_spent":      avgTimeSpent,
		"completion_rate":     100.0, // 假设都是完成的响应
		"question_statistics": questionStats,
		"generated_at":        time.Now(),
	}

	resultData, _ := json.Marshal(result)
	h.db.Model(analysis).Updates(map[string]interface{}{
		"status":      "completed",
		"result_data": datatypes.JSON(resultData),
	})
}

// performAIAnalysis 执行AI分析
func (h *SurveyHandler) performAIAnalysis(analysis *models.SurveyAnalysis) {
	h.db.Model(analysis).Update("status", "running")

	// 获取所有响应进行批量分析
	if err := h.difyService.BatchAnalyzeSurvey(analysis.SurveyID); err != nil {
		h.db.Model(analysis).Updates(map[string]interface{}{
			"status":        "failed",
			"error_message": fmt.Sprintf("AI analysis failed: %v", err),
		})
		return
	}

	// AI分析的结果会由DifyService异步更新
}

// performCustomAnalysis 执行自定义分析
func (h *SurveyHandler) performCustomAnalysis(analysis *models.SurveyAnalysis, workflowID string) {
	h.db.Model(analysis).Update("status", "running")

	if h.difyService == nil || workflowID == "" {
		h.db.Model(analysis).Updates(map[string]interface{}{
			"status":        "failed",
			"error_message": "Custom analysis requires Dify service and workflow ID",
		})
		return
	}

	// 获取响应数据
	var responses []models.SurveyResponse
	if err := h.db.Where("survey_id = ?", analysis.SurveyID).Find(&responses).Error; err != nil {
		h.db.Model(analysis).Updates(map[string]interface{}{
			"status":        "failed",
			"error_message": fmt.Sprintf("Failed to get responses: %v", err),
		})
		return
	}

	// 构建自定义工作流输入
	inputs := map[string]interface{}{
		"survey_id": analysis.SurveyID,
		"responses": responses,
		"analysis_type": "custom",
	}

	// 运行自定义工作流
	result, err := h.difyService.runWorkflow(workflowID, inputs, "system")
	if err != nil {
		h.db.Model(analysis).Updates(map[string]interface{}{
			"status":        "failed",
			"error_message": fmt.Sprintf("Custom workflow failed: %v", err),
		})
		return
	}

	resultData, _ := json.Marshal(result.Outputs)
	h.db.Model(analysis).Updates(map[string]interface{}{
		"status":      result.Status,
		"result_data": datatypes.JSON(resultData),
		"run_id":      result.WorkflowRunID,
	})
}

// analyzeQuestionStats 分析问题统计数据
func (h *SurveyHandler) analyzeQuestionStats(responses []models.SurveyResponse) map[string]interface{} {
	stats := make(map[string]interface{})
	
	if len(responses) == 0 {
		return stats
	}

	// 统计每个问题的回答情况
	questionAnswers := make(map[string][]interface{})
	
	for _, response := range responses {
		var data map[string]interface{}
		if err := json.Unmarshal(response.ResponseData, &data); err == nil {
			for question, answer := range data {
				if answer != nil {
					questionAnswers[question] = append(questionAnswers[question], answer)
				}
			}
		}
	}

	// 为每个问题生成统计
	for question, answers := range questionAnswers {
		questionStat := map[string]interface{}{
			"total_answers": len(answers),
			"response_rate": float64(len(answers)) / float64(len(responses)) * 100,
		}

		// 分析答案类型和分布
		if len(answers) > 0 {
			switch answers[0].(type) {
			case string:
				// 文本或选择题
				answerCounts := make(map[string]int)
				for _, answer := range answers {
					if str, ok := answer.(string); ok {
						answerCounts[str]++
					}
				}
				questionStat["answer_distribution"] = answerCounts
				
			case float64:
				// 数值题
				var sum, min, max float64
				min = answers[0].(float64)
				max = answers[0].(float64)
				
				for _, answer := range answers {
					if num, ok := answer.(float64); ok {
						sum += num
						if num < min {
							min = num
						}
						if num > max {
							max = num
						}
					}
				}
				
				questionStat["average"] = sum / float64(len(answers))
				questionStat["min"] = min
				questionStat["max"] = max
			}
		}

		stats[question] = questionStat
	}

	return stats
}

// exportCSV 导出CSV格式
func (h *SurveyHandler) exportCSV(c *gin.Context, survey models.Survey, responses []models.SurveyResponse, req ExportRequest) {
	c.Header("Content-Type", "text/csv")
	c.Header("Content-Disposition", fmt.Sprintf("attachment; filename=\"%s_responses.csv\"", survey.SurveyID))

	writer := csv.NewWriter(c.Writer)
	defer writer.Flush()

	// 写入标题行
	headers := []string{"ResponseID", "CompletedAt", "TimeSpent"}
	if req.IncludeUser {
		headers = append(headers, "UserID")
	}

	// 获取所有问题字段
	questionFields := h.extractQuestionFields(responses)
	headers = append(headers, questionFields...)

	writer.Write(headers)

	// 写入数据行
	for _, response := range responses {
		row := []string{
			response.ID,
			response.CompletedAt.Format("2006-01-02 15:04:05"),
			strconv.Itoa(response.TimeSpent),
		}
		
		if req.IncludeUser {
			row = append(row, response.UserID)
		}

		// 添加答案数据
		var data map[string]interface{}
		if json.Unmarshal(response.ResponseData, &data) == nil {
			for _, field := range questionFields {
				if answer, exists := data[field]; exists && answer != nil {
					row = append(row, fmt.Sprintf("%v", answer))
				} else {
					row = append(row, "")
				}
			}
		}

		writer.Write(row)
	}
}

// exportJSON 导出JSON格式
func (h *SurveyHandler) exportJSON(c *gin.Context, survey models.Survey, responses []models.SurveyResponse, req ExportRequest) {
	c.Header("Content-Type", "application/json")
	c.Header("Content-Disposition", fmt.Sprintf("attachment; filename=\"%s_responses.json\"", survey.SurveyID))

	exportData := map[string]interface{}{
		"survey_info": map[string]interface{}{
			"id":          survey.SurveyID,
			"title":       survey.Title,
			"description": survey.Description,
			"exported_at": time.Now(),
		},
		"responses": responses,
	}

	c.JSON(http.StatusOK, exportData)
}

// exportExcel 导出Excel格式（简化实现）
func (h *SurveyHandler) exportExcel(c *gin.Context, survey models.Survey, responses []models.SurveyResponse, req ExportRequest) {
	// 这里应该使用Excel库，暂时返回CSV格式
	h.exportCSV(c, survey, responses, req)
}

// extractQuestionFields 提取问题字段
func (h *SurveyHandler) extractQuestionFields(responses []models.SurveyResponse) []string {
	fieldSet := make(map[string]bool)
	
	for _, response := range responses {
		var data map[string]interface{}
		if json.Unmarshal(response.ResponseData, &data) == nil {
			for field := range data {
				fieldSet[field] = true
			}
		}
	}

	fields := make([]string, 0, len(fieldSet))
	for field := range fieldSet {
		fields = append(fields, field)
	}

	return fields
}

// 权限检查辅助方法
func (h *SurveyHandler) hasAnalyzePermission(surveyID, userID string) bool {
	var permission models.SurveyPermission
	err := h.db.Where("survey_id = ? AND user_id = ? AND can_analyze = ?", surveyID, userID, true).First(&permission).Error
	return err == nil
}

func (h *SurveyHandler) hasExportPermission(surveyID, userID string) bool {
	var permission models.SurveyPermission
	err := h.db.Where("survey_id = ? AND user_id = ? AND can_export = ?", surveyID, userID, true).First(&permission).Error
	return err == nil
}