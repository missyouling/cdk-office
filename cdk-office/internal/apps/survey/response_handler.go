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
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"cdk-office/internal/models"
)

// SubmitResponse 提交问卷响应
func (h *SurveyHandler) SubmitResponse(c *gin.Context) {
	surveyID := c.Param("id")
	userID := c.GetString("user_id") // 可能为空（匿名用户）
	teamID := c.GetString("team_id")

	var req SubmitResponseRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request data", "details": err.Error()})
		return
	}

	// 检查问卷是否存在且可以响应
	var survey models.Survey
	if err := h.db.Where("survey_id = ?", surveyID).First(&survey).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "Survey not found"})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get survey"})
		}
		return
	}

	// 检查问卷状态
	if survey.Status != "active" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Survey is not active"})
		return
	}

	// 检查时间限制
	now := time.Now()
	if survey.StartTime != nil && now.Before(*survey.StartTime) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Survey has not started yet"})
		return
	}
	if survey.EndTime != nil && now.After(*survey.EndTime) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Survey has ended"})
		return
	}

	// 检查最大响应数量限制
	if survey.MaxResponses > 0 && survey.ResponseCount >= survey.MaxResponses {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Survey has reached maximum responses"})
		return
	}

	// 检查用户是否已经响应过（如果不是匿名）
	if userID != "" {
		var existingResponse models.SurveyResponse
		if err := h.db.Where("survey_id = ? AND user_id = ?", surveyID, userID).First(&existingResponse).Error; err == nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "You have already responded to this survey"})
			return
		}
	}

	// 创建响应记录
	response := &models.SurveyResponse{
		SurveyID:     surveyID,
		UserID:       userID,
		TeamID:       teamID,
		ResponseData: req.ResponseData,
		TimeSpent:    req.TimeSpent,
		IPAddress:    c.ClientIP(),
		UserAgent:    req.UserAgent,
		IsCompleted:  true,
		CompletedAt:  time.Now(),
	}

	if err := h.db.Create(response).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to submit response"})
		return
	}

	// 更新问卷响应计数
	h.db.Model(&survey).UpdateColumn("response_count", gorm.Expr("response_count + ?", 1))

	// 异步触发AI分析
	go h.triggerDifyAnalysis(response)

	c.JSON(http.StatusCreated, ResponseResponse{
		ID:           response.ID,
		SurveyID:     response.SurveyID,
		UserID:       response.UserID,
		ResponseData: response.ResponseData,
		TimeSpent:    response.TimeSpent,
		IsCompleted:  response.IsCompleted,
		CompletedAt:  response.CompletedAt,
		CreatedAt:    response.CreatedAt,
	})
}

// SubmitPublicResponse 提交公开问卷响应（无需认证）
func (h *SurveyHandler) SubmitPublicResponse(c *gin.Context) {
	surveyID := c.Param("id")

	var req SubmitResponseRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request data", "details": err.Error()})
		return
	}

	// 检查问卷是否存在且是公开的
	var survey models.Survey
	if err := h.db.Where("survey_id = ? AND is_public = ?", surveyID, true).First(&survey).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "Public survey not found"})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get survey"})
		}
		return
	}

	// 检查问卷状态和时间限制（与SubmitResponse相同的逻辑）
	if survey.Status != "active" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Survey is not active"})
		return
	}

	now := time.Now()
	if survey.StartTime != nil && now.Before(*survey.StartTime) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Survey has not started yet"})
		return
	}
	if survey.EndTime != nil && now.After(*survey.EndTime) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Survey has ended"})
		return
	}

	if survey.MaxResponses > 0 && survey.ResponseCount >= survey.MaxResponses {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Survey has reached maximum responses"})
		return
	}

	// 创建匿名响应记录
	response := &models.SurveyResponse{
		SurveyID:     surveyID,
		TeamID:       survey.TeamID,
		ResponseData: req.ResponseData,
		TimeSpent:    req.TimeSpent,
		IPAddress:    c.ClientIP(),
		UserAgent:    req.UserAgent,
		IsCompleted:  true,
		CompletedAt:  time.Now(),
	}

	if err := h.db.Create(response).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to submit response"})
		return
	}

	// 更新问卷响应计数
	h.db.Model(&survey).UpdateColumn("response_count", gorm.Expr("response_count + ?", 1))

	// 异步触发AI分析
	go h.triggerDifyAnalysis(response)

	c.JSON(http.StatusCreated, gin.H{"message": "Response submitted successfully", "response_id": response.ID})
}

// GetResponses 获取问卷响应列表
func (h *SurveyHandler) GetResponses(c *gin.Context) {
	surveyID := c.Param("id")
	userID := c.GetString("user_id")

	// 检查查看权限
	if !h.hasViewPermission(surveyID, userID) {
		c.JSON(http.StatusForbidden, gin.H{"error": "No view permission"})
		return
	}

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))
	includeData := c.DefaultQuery("include_data", "false") == "true"

	offset := (page - 1) * pageSize

	query := h.db.Where("survey_id = ?", surveyID)

	var responses []models.SurveyResponse
	var total int64

	query.Model(&models.SurveyResponse{}).Count(&total)

	selectFields := "id, survey_id, user_id, time_spent, is_completed, completed_at, created_at"
	if includeData {
		selectFields += ", response_data"
	}

	if err := query.Select(selectFields).Order("created_at DESC").Offset(offset).Limit(pageSize).Find(&responses).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get responses"})
		return
	}

	// 转换为响应格式
	responseList := make([]ResponseResponse, len(responses))
	for i, resp := range responses {
		responseList[i] = ResponseResponse{
			ID:          resp.ID,
			SurveyID:    resp.SurveyID,
			UserID:      resp.UserID,
			TimeSpent:   resp.TimeSpent,
			IsCompleted: resp.IsCompleted,
			CompletedAt: resp.CompletedAt,
			CreatedAt:   resp.CreatedAt,
		}
		if includeData {
			responseList[i].ResponseData = resp.ResponseData
		}
	}

	c.JSON(http.StatusOK, ListResponse{
		Data:  responseList,
		Total: total,
		Page:  page,
		Pages: (total + int64(pageSize) - 1) / int64(pageSize),
	})
}

// GetResponse 获取单个响应详情
func (h *SurveyHandler) GetResponse(c *gin.Context) {
	responseID := c.Param("responseId")
	userID := c.GetString("user_id")

	var response models.SurveyResponse
	if err := h.db.Where("id = ?", responseID).First(&response).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "Response not found"})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get response"})
		}
		return
	}

	// 检查权限（只有问卷管理者或响应者本人可以查看）
	if !h.hasViewPermission(response.SurveyID, userID) && response.UserID != userID {
		c.JSON(http.StatusForbidden, gin.H{"error": "Access denied"})
		return
	}

	c.JSON(http.StatusOK, ResponseResponse{
		ID:           response.ID,
		SurveyID:     response.SurveyID,
		UserID:       response.UserID,
		ResponseData: response.ResponseData,
		TimeSpent:    response.TimeSpent,
		IsCompleted:  response.IsCompleted,
		CompletedAt:  response.CompletedAt,
		CreatedAt:    response.CreatedAt,
	})
}

// DeleteResponse 删除响应
func (h *SurveyHandler) DeleteResponse(c *gin.Context) {
	responseID := c.Param("responseId")
	userID := c.GetString("user_id")

	var response models.SurveyResponse
	if err := h.db.Where("id = ?", responseID).First(&response).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "Response not found"})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get response"})
		}
		return
	}

	// 检查删除权限（只有问卷管理者可以删除）
	if !h.hasManagePermission(response.SurveyID, userID) {
		c.JSON(http.StatusForbidden, gin.H{"error": "No delete permission"})
		return
	}

	if err := h.db.Delete(&response).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete response"})
		return
	}

	// 更新问卷响应计数
	var survey models.Survey
	if err := h.db.Where("survey_id = ?", response.SurveyID).First(&survey).Error; err == nil {
		h.db.Model(&survey).UpdateColumn("response_count", gorm.Expr("response_count - ?", 1))
	}

	c.JSON(http.StatusOK, gin.H{"message": "Response deleted successfully"})
}

// GetPublicSurvey 获取公开问卷（无需认证）
func (h *SurveyHandler) GetPublicSurvey(c *gin.Context) {
	surveyID := c.Param("id")

	var survey models.Survey
	if err := h.db.Where("survey_id = ? AND is_public = ?", surveyID, true).First(&survey).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "Public survey not found"})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get survey"})
		}
		return
	}

	// 检查问卷状态
	if survey.Status != "active" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Survey is not available"})
		return
	}

	// 增加浏览次数
	h.db.Model(&survey).UpdateColumn("view_count", gorm.Expr("view_count + ?", 1))

	// 返回公开信息（不包含敏感数据）
	response := SurveyResponse{
		ID:             survey.ID,
		SurveyID:       survey.SurveyID,
		Title:          survey.Title,
		Description:    survey.Description,
		Status:         survey.Status,
		IsPublic:       survey.IsPublic,
		Tags:           survey.Tags,
		CreatedAt:      survey.CreatedAt,
		JsonDefinition: &survey.JsonDefinition,
		StartTime:      survey.StartTime,
		EndTime:        survey.EndTime,
	}

	c.JSON(http.StatusOK, response)
}

// triggerDifyAnalysis 触发Dify AI分析（异步）
func (h *SurveyHandler) triggerDifyAnalysis(response *models.SurveyResponse) {
	// 这里将实现Dify AI分析的触发逻辑
	// 暂时只记录日志
	fmt.Printf("Triggering Dify analysis for response %s in survey %s\n", response.ID, response.SurveyID)
	
	// TODO: 实现Dify API调用
	// 1. 构建分析请求数据
	// 2. 调用Dify工作流
	// 3. 保存分析结果到数据库
}

// 权限检查辅助方法
func (h *SurveyHandler) hasViewPermission(surveyID, userID string) bool {
	var permission models.SurveyPermission
	err := h.db.Where("survey_id = ? AND user_id = ? AND can_view = ?", surveyID, userID, true).First(&permission).Error
	return err == nil
}