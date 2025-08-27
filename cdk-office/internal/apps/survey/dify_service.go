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
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"gorm.io/gorm"
	"gorm.io/datatypes"

	"cdk-office/internal/models"
)

// DifyConfig Dify平台配置
type DifyConfig struct {
	BaseURL               string `json:"base_url"`                 // Dify API基础URL
	APIKey                string `json:"api_key"`                  // Dify API密钥
	SurveyAnalysisWorkflowID string `json:"survey_analysis_workflow_id"` // 问卷分析工作流ID
	KnowledgeBaseID       string `json:"knowledge_base_id"`        // 知识库ID
	Timeout               time.Duration `json:"timeout"`           // 请求超时时间
}

// DifyService Dify服务
type DifyService struct {
	config DifyConfig
	client *http.Client
	db     *gorm.DB
}

// NewDifyService 创建Dify服务
func NewDifyService(config DifyConfig, db *gorm.DB) *DifyService {
	if config.Timeout == 0 {
		config.Timeout = 30 * time.Second
	}

	return &DifyService{
		config: config,
		client: &http.Client{Timeout: config.Timeout},
		db:     db,
	}
}

// DifyWorkflowRequest Dify工作流请求
type DifyWorkflowRequest struct {
	Inputs map[string]interface{} `json:"inputs"`
	User   string                 `json:"user,omitempty"`
}

// DifyWorkflowResponse Dify工作流响应
type DifyWorkflowResponse struct {
	WorkflowRunID string                 `json:"workflow_run_id"`
	TaskID        string                 `json:"task_id"`
	Status        string                 `json:"status"` // running, succeeded, failed
	Outputs       map[string]interface{} `json:"outputs"`
	Error         string                 `json:"error,omitempty"`
}

// DifyKnowledgeRequest Dify知识库请求
type DifyKnowledgeRequest struct {
	Name     string            `json:"name"`
	Text     string            `json:"text"`
	Metadata map[string]string `json:"metadata"`
}

// TriggerSurveyAnalysis 触发问卷分析
func (s *DifyService) TriggerSurveyAnalysis(response *models.SurveyResponse) error {
	// 获取问卷信息
	var survey models.Survey
	if err := s.db.Where("survey_id = ?", response.SurveyID).First(&survey).Error; err != nil {
		return fmt.Errorf("failed to get survey: %v", err)
	}

	// 构建工作流输入数据
	inputs := map[string]interface{}{
		"survey_id":       response.SurveyID,
		"survey_title":    survey.Title,
		"survey_description": survey.Description,
		"response_id":     response.ID,
		"response_data":   response.ResponseData,
		"user_id":         response.UserID,
		"team_id":         response.TeamID,
		"completed_at":    response.CompletedAt,
		"time_spent":      response.TimeSpent,
		"ip_address":      response.IPAddress,
	}

	// 调用Dify工作流
	result, err := s.runWorkflow(s.config.SurveyAnalysisWorkflowID, inputs, response.UserID)
	if err != nil {
		return fmt.Errorf("failed to run Dify workflow: %v", err)
	}

	// 保存分析结果
	analysis := &models.SurveyAnalysis{
		SurveyID:       response.SurveyID,
		AnalysisType:   "ai",
		ResultData:     datatypes.JSON(mustMarshal(result.Outputs)),
		DifyWorkflowID: s.config.SurveyAnalysisWorkflowID,
		RunID:          result.WorkflowRunID,
		Status:         result.Status,
		ErrorMessage:   result.Error,
	}

	if err := s.db.Create(analysis).Error; err != nil {
		return fmt.Errorf("failed to save analysis result: %v", err)
	}

	// 如果分析成功，将结果提交到知识库
	if result.Status == "succeeded" && s.config.KnowledgeBaseID != "" {
		go s.submitToKnowledgeBase(response.SurveyID, analysis)
	}

	return nil
}

// runWorkflow 运行Dify工作流
func (s *DifyService) runWorkflow(workflowID string, inputs map[string]interface{}, userID string) (*DifyWorkflowResponse, error) {
	url := fmt.Sprintf("%s/v1/workflows/%s/runs", s.config.BaseURL, workflowID)

	request := DifyWorkflowRequest{
		Inputs: inputs,
		User:   userID,
	}

	jsonData, err := json.Marshal(request)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %v", err)
	}

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %v", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+s.config.APIKey)

	resp, err := s.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %v", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %v", err)
	}

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		return nil, fmt.Errorf("workflow execution failed with status %d: %s", resp.StatusCode, string(body))
	}

	var response DifyWorkflowResponse
	if err := json.Unmarshal(body, &response); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %v", err)
	}

	return &response, nil
}

// submitToKnowledgeBase 将分析结果提交到知识库
func (s *DifyService) submitToKnowledgeBase(surveyID string, analysis *models.SurveyAnalysis) error {
	// 获取问卷信息
	var survey models.Survey
	if err := s.db.Where("survey_id = ?", surveyID).First(&survey).Error; err != nil {
		return fmt.Errorf("failed to get survey: %v", err)
	}

	// 生成知识库文档
	document := s.generateAnalysisDocument(&survey, analysis)

	url := fmt.Sprintf("%s/v1/datasets/%s/documents", s.config.BaseURL, s.config.KnowledgeBaseID)

	jsonData, err := json.Marshal(document)
	if err != nil {
		return fmt.Errorf("failed to marshal document: %v", err)
	}

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return fmt.Errorf("failed to create request: %v", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+s.config.APIKey)

	resp, err := s.client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send request: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("knowledge base submission failed with status %d: %s", resp.StatusCode, string(body))
	}

	return nil
}

// generateAnalysisDocument 生成分析文档
func (s *DifyService) generateAnalysisDocument(survey *models.Survey, analysis *models.SurveyAnalysis) DifyKnowledgeRequest {
	// 提取分析结果数据
	var outputs map[string]interface{}
	json.Unmarshal(analysis.ResultData, &outputs)

	// 生成文档内容
	content := fmt.Sprintf(`# 问卷分析报告

## 问卷信息
- 标题: %s
- 描述: %s
- ID: %s
- 创建时间: %s

## 分析结果
类型: %s
分析时间: %s
工作流ID: %s

## 详细分析数据
%s

## 标签
%s
`, 
		survey.Title,
		survey.Description,
		survey.SurveyID,
		survey.CreatedAt.Format("2006-01-02 15:04:05"),
		analysis.AnalysisType,
		analysis.CreatedAt.Format("2006-01-02 15:04:05"),
		analysis.DifyWorkflowID,
		mustMarshalIndent(outputs),
		survey.Tags,
	)

	return DifyKnowledgeRequest{
		Name: fmt.Sprintf("问卷分析报告_%s_%s", survey.Title, analysis.CreatedAt.Format("20060102150405")),
		Text: content,
		Metadata: map[string]string{
			"type":        "survey_analysis",
			"survey_id":   survey.SurveyID,
			"analysis_id": analysis.ID,
			"team_id":     survey.TeamID,
			"tags":        survey.Tags,
		},
	}
}

// GetAnalysisStatus 获取分析状态
func (s *DifyService) GetAnalysisStatus(runID string) (*DifyWorkflowResponse, error) {
	url := fmt.Sprintf("%s/v1/workflows/runs/%s", s.config.BaseURL, runID)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %v", err)
	}

	req.Header.Set("Authorization", "Bearer "+s.config.APIKey)

	resp, err := s.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %v", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %v", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("get status failed with status %d: %s", resp.StatusCode, string(body))
	}

	var response DifyWorkflowResponse
	if err := json.Unmarshal(body, &response); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %v", err)
	}

	return &response, nil
}

// BatchAnalyzeSurvey 批量分析问卷的所有响应
func (s *DifyService) BatchAnalyzeSurvey(surveyID string) error {
	// 获取所有响应
	var responses []models.SurveyResponse
	if err := s.db.Where("survey_id = ?", surveyID).Find(&responses).Error; err != nil {
		return fmt.Errorf("failed to get responses: %v", err)
	}

	// 构建批量分析输入
	inputs := map[string]interface{}{
		"survey_id":    surveyID,
		"responses":    responses,
		"analysis_type": "batch",
	}

	// 调用批量分析工作流
	result, err := s.runWorkflow(s.config.SurveyAnalysisWorkflowID, inputs, "system")
	if err != nil {
		return fmt.Errorf("failed to run batch analysis: %v", err)
	}

	// 保存批量分析结果
	analysis := &models.SurveyAnalysis{
		SurveyID:       surveyID,
		AnalysisType:   "batch_ai",
		ResultData:     datatypes.JSON(mustMarshal(result.Outputs)),
		DifyWorkflowID: s.config.SurveyAnalysisWorkflowID,
		RunID:          result.WorkflowRunID,
		Status:         result.Status,
		ErrorMessage:   result.Error,
	}

	if err := s.db.Create(analysis).Error; err != nil {
		return fmt.Errorf("failed to save batch analysis result: %v", err)
	}

	return nil
}

// ValidateConnection 验证Dify连接
func (s *DifyService) ValidateConnection() error {
	url := fmt.Sprintf("%s/v1/datasets", s.config.BaseURL)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return fmt.Errorf("failed to create request: %v", err)
	}

	req.Header.Set("Authorization", "Bearer "+s.config.APIKey)

	resp, err := s.client.Do(req)
	if err != nil {
		return fmt.Errorf("connection failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusUnauthorized {
		return fmt.Errorf("invalid API key")
	}

	if resp.StatusCode >= 400 {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("validation failed with status %d: %s", resp.StatusCode, string(body))
	}

	return nil
}

// 辅助函数
func mustMarshal(v interface{}) []byte {
	data, err := json.Marshal(v)
	if err != nil {
		return []byte("{}")
	}
	return data
}

func mustMarshalIndent(v interface{}) string {
	data, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		return "{}"
	}
	return string(data)
}