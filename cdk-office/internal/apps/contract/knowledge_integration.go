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

package contract

import (
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/google/uuid"
	"github.com/linux-do/cdk-office/internal/models"
	"gorm.io/gorm"
)

// KnowledgeIntegrationService 知识库集成服务
type KnowledgeIntegrationService struct {
	db             *gorm.DB
	workflowEngine *WorkflowEngine
	approvalService *ApprovalService
	knowledgeBase  *KnowledgeBaseService
	notifyService  *NotificationService
	difyClient     *DifyClient
}

// NewKnowledgeIntegrationService 创建知识库集成服务
func NewKnowledgeIntegrationService(db *gorm.DB) *KnowledgeIntegrationService {
	return &KnowledgeIntegrationService{
		db:             db,
		workflowEngine: NewWorkflowEngine(),
		approvalService: NewApprovalService(),
		knowledgeBase:  NewKnowledgeBaseService(),
		notifyService:  NewNotificationService(),
		difyClient:     NewDifyClient(),
	}
}

// DifyProcessRequest Dify处理请求
type DifyProcessRequest struct {
	DocumentID   string                 `json:"document_id"`
	Title        string                 `json:"title"`
	Content      string                 `json:"content"`
	FileURL      string                 `json:"file_url"`
	DocumentType string                 `json:"document_type"`
	Metadata     map[string]interface{} `json:"metadata"`
}

// DifyProcessResponse Dify处理响应
type DifyProcessResponse struct {
	Summary  string   `json:"summary"`
	Keywords []string `json:"keywords"`
	Tags     []string `json:"tags"`
	Status   string   `json:"status"`
}

// ProcessContractCompletion 处理合同完成事件
func (s *KnowledgeIntegrationService) ProcessContractCompletion(contract *models.Contract) error {
	log.Printf("[KnowledgeIntegration] 开始处理合同完成事件: %s", contract.ID)

	// 1. 获取团队工作流配置
	workflow, err := s.getTeamContractWorkflow(contract.TeamID)
	if err != nil {
		log.Printf("[KnowledgeIntegration] 获取工作流配置失败: %v", err)
		return err
	}

	if workflow == nil || !workflow.AutoSubmitKnowledge {
		log.Printf("[KnowledgeIntegration] 团队未配置自动提交知识库")
		return nil
	}

	// 2. 创建知识库文档记录
	document, err := s.createKnowledgeDocument(contract, workflow)
	if err != nil {
		log.Printf("[KnowledgeIntegration] 创建知识库文档失败: %v", err)
		return err
	}

	// 3. 创建知识库提交记录
	submission, err := s.createKnowledgeSubmission(contract, document, workflow)
	if err != nil {
		log.Printf("[KnowledgeIntegration] 创建知识库提交记录失败: %v", err)
		return err
	}

	// 4. 判断是否需要审批
	if workflow.RequireApproval {
		log.Printf("[KnowledgeIntegration] 需要审批，创建审批任务")
		return s.createApprovalWorkflow(contract, document, submission, workflow)
	} else {
		log.Printf("[KnowledgeIntegration] 无需审批，直接处理")
		return s.processKnowledgeSubmission(submission)
	}
}

// createKnowledgeDocument 创建知识库文档
func (s *KnowledgeIntegrationService) createKnowledgeDocument(contract *models.Contract, workflow *models.ContractWorkflow) (*models.Document, error) {
	// 生成文档内容
	content := s.generateDocumentContent(contract)

	// 生成标签
	tags := s.generateDocumentTags(contract, workflow)

	// 创建文档记录
	document := &models.Document{
		ID:             uuid.New().String(),
		TeamID:         contract.TeamID,
		Name:           fmt.Sprintf("合同：%s", contract.Title),
		Description:    fmt.Sprintf("合同《%s》签署完成后自动生成的知识库文档", contract.Title),
		FileName:       fmt.Sprintf("%s_已签署.pdf", contract.Title),
		FilePath:       contract.FinalFileURL,
		FileSize:       0, // 需要从文件服务获取
		FileType:       "contract",
		MimeType:       "application/pdf",
		Status:         "pending_review",
		Version:        1,
		CreatedBy:      contract.CreatedBy,
		UpdatedBy:      contract.CreatedBy,
		Tags:           tags,
	}

	if err := s.db.Create(document).Error; err != nil {
		return nil, fmt.Errorf("创建知识库文档失败: %w", err)
	}

	return document, nil
}

// createKnowledgeSubmission 创建知识库提交记录
func (s *KnowledgeIntegrationService) createKnowledgeSubmission(contract *models.Contract, document *models.Document, workflow *models.ContractWorkflow) (*models.KnowledgeSubmission, error) {
	submission := &models.KnowledgeSubmission{
		ID:                 uuid.New().String(),
		ContractID:         contract.ID,
		DocumentID:         document.ID,
		SubmissionType:     "auto",
		Status:             "pending",
		AutoProcessing:     true,
		ApprovalRequired:   workflow.RequireApproval,
		DifyWorkflowStatus: "pending",
		CreatedBy:          contract.CreatedBy,
	}

	if err := s.db.Create(submission).Error; err != nil {
		return nil, fmt.Errorf("创建知识库提交记录失败: %w", err)
	}

	return submission, nil
}

// createApprovalWorkflow 创建审批工作流
func (s *KnowledgeIntegrationService) createApprovalWorkflow(contract *models.Contract, document *models.Document, submission *models.KnowledgeSubmission, workflow *models.ContractWorkflow) error {
	// 1. 创建工作流实例
	workflowInstance := &models.WorkflowInstance{
		ID:            uuid.New().String(),
		WorkflowDefID: workflow.ID,
		Status:        "running",
		InputData:     s.serializeWorkflowInput(contract, document),
		CreatedBy:     contract.CreatedBy,
	}

	if err := s.db.Create(workflowInstance).Error; err != nil {
		return fmt.Errorf("创建工作流实例失败: %w", err)
	}

	// 2. 解析审批角色
	var approvalRoles []string
	if err := json.Unmarshal([]byte(workflow.ApprovalRoles), &approvalRoles); err != nil {
		return fmt.Errorf("解析审批角色失败: %w", err)
	}

	// 3. 创建审批任务
	for _, role := range approvalRoles {
		approvers, err := s.getApproversByRole(contract.TeamID, role)
		if err != nil {
			log.Printf("[KnowledgeIntegration] 获取审批人失败: %v", err)
			continue
		}

		for _, approver := range approvers {
			approvalTask := &models.ApprovalTask{
				ID:             uuid.New().String(),
				WorkflowInstID: workflowInstance.ID,
				Name:           fmt.Sprintf("合同知识库审批：%s", contract.Title),
				Description:    fmt.Sprintf("合同《%s》已完成签署，需要审批是否加入知识库", contract.Title),
				Assignee:       approver.ID,
				Status:         "pending",
				DueDate:        time.Now().Add(time.Duration(workflow.ApprovalTimeout) * time.Hour),
			}

			if err := s.db.Create(approvalTask).Error; err != nil {
				log.Printf("[KnowledgeIntegration] 创建审批任务失败: %v", err)
				continue
			}

			// 发送审批通知
			go s.sendApprovalNotification(approver, approvalTask, contract)
		}
	}

	// 4. 更新提交状态
	return s.db.Model(submission).Updates(map[string]interface{}{
		"approval_status": "pending",
		"status":          "pending_approval",
	}).Error
}

// processKnowledgeSubmission 处理知识库提交
func (s *KnowledgeIntegrationService) processKnowledgeSubmission(submission *models.KnowledgeSubmission) error {
	log.Printf("[KnowledgeIntegration] 开始处理知识库提交: %s", submission.ID)

	// 1. 调用Dify API处理文档
	if submission.AutoProcessing {
		if err := s.processByDify(submission); err != nil {
			log.Printf("[KnowledgeIntegration] Dify处理失败: %v", err)
			return err
		}
	}

	// 2. 更新文档状态
	if err := s.db.Model(&models.Document{}).Where("id = ?", submission.DocumentID).Updates(map[string]interface{}{
		"status":     "published",
		"updated_at": time.Now(),
	}).Error; err != nil {
		return fmt.Errorf("更新文档状态失败: %w", err)
	}

	// 3. 更新提交状态
	if err := s.db.Model(submission).Updates(map[string]interface{}{
		"status":     "completed",
		"updated_at": time.Now(),
	}).Error; err != nil {
		return fmt.Errorf("更新提交状态失败: %w", err)
	}

	// 4. 发送完成通知
	go s.notifyKnowledgeBaseUpdate(submission)

	log.Printf("[KnowledgeIntegration] 知识库提交处理完成: %s", submission.ID)
	return nil
}

// processByDify 通过Dify处理文档
func (s *KnowledgeIntegrationService) processByDify(submission *models.KnowledgeSubmission) error {
	// 获取文档信息
	document, err := s.getDocument(submission.DocumentID)
	if err != nil {
		return fmt.Errorf("获取文档失败: %w", err)
	}

	// 调用Dify API进行文档向量化和知识抽取
	result, err := s.difyClient.ProcessDocument(DifyProcessRequest{
		DocumentID:   document.ID,
		Title:        document.Name,
		Content:      "", // 从文件中提取
		FileURL:      document.FilePath,
		DocumentType: "contract",
		Metadata: map[string]interface{}{
			"contract_id":      submission.ContractID,
			"file_type":        document.FileType,
			"created_at":       document.CreatedAt,
			"team_id":          document.TeamID,
		},
	})

	if err != nil {
		// 更新处理状态为失败
		s.db.Model(submission).Updates(map[string]interface{}{
			"dify_workflow_status": "failed",
			"updated_at":           time.Now(),
		})
		return fmt.Errorf("Dify处理文档失败: %w", err)
	}

	// 更新处理结果
	s.db.Model(submission).Updates(map[string]interface{}{
		"dify_workflow_status": "completed",
		"dify_result":          s.serializeDifyResult(result),
		"updated_at":           time.Now(),
	})

	// 更新文档的AI处理结果
	s.db.Model(document).Updates(map[string]interface{}{
		"last_sync_at": time.Now(),
		"updated_at":   time.Now(),
	})

	return nil
}

// ApproveKnowledgeSubmission 审批知识库提交
func (s *KnowledgeIntegrationService) ApproveKnowledgeSubmission(taskID, approverID, comments string, approved bool) error {
	// 获取审批任务
	var task models.ApprovalTask
	if err := s.db.Where("id = ? AND assignee = ?", taskID, approverID).First(&task).Error; err != nil {
		return fmt.Errorf("审批任务不存在或无权限")
	}

	if task.Status != "pending" {
		return fmt.Errorf("任务状态异常，无法审批")
	}

	tx := s.db.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	// 更新审批任务状态
	status := "rejected"
	if approved {
		status = "approved"
	}

	if err := tx.Model(&task).Updates(map[string]interface{}{
		"status":     status,
		"comments":   comments,
		"updated_at": time.Now(),
	}).Error; err != nil {
		tx.Rollback()
		return fmt.Errorf("更新审批任务失败: %w", err)
	}

	// 检查所有审批任务是否完成
	allCompleted, allApproved, err := s.checkAllApprovalsCompleted(task.WorkflowInstID)
	if err != nil {
		tx.Rollback()
		return err
	}

	if allCompleted {
		// 获取知识库提交记录
		submission, err := s.getSubmissionByWorkflow(task.WorkflowInstID)
		if err != nil {
			tx.Rollback()
			return err
		}

		if allApproved {
			// 所有审批通过，处理知识库提交
			if err := tx.Model(submission).Updates(map[string]interface{}{
				"approval_status": "approved",
				"status":          "approved",
				"approval_by":     approverID,
				"approval_at":     time.Now(),
				"approval_comments": comments,
			}).Error; err != nil {
				tx.Rollback()
				return fmt.Errorf("更新提交状态失败: %w", err)
			}

			tx.Commit()

			// 异步处理知识库提交
			go s.processKnowledgeSubmission(submission)
		} else {
			// 审批被拒绝
			if err := tx.Model(submission).Updates(map[string]interface{}{
				"approval_status": "rejected",
				"status":          "rejected",
				"approval_by":     approverID,
				"approval_at":     time.Now(),
				"approval_comments": comments,
			}).Error; err != nil {
				tx.Rollback()
				return fmt.Errorf("更新提交状态失败: %w", err)
			}

			tx.Commit()

			// 发送拒绝通知
			go s.notifyKnowledgeSubmissionRejected(submission, comments)
		}
	} else {
		tx.Commit()
	}

	return nil
}

// 辅助方法

// getTeamContractWorkflow 获取团队合同工作流配置
func (s *KnowledgeIntegrationService) getTeamContractWorkflow(teamID string) (*models.ContractWorkflow, error) {
	var workflow models.ContractWorkflow
	err := s.db.Where("team_id = ? AND is_active = true", teamID).
		Order("is_default DESC, created_at DESC").First(&workflow).Error
	
	if err == gorm.ErrRecordNotFound {
		return nil, nil // 没有配置工作流
	}
	
	return &workflow, err
}

// generateDocumentContent 生成文档内容
func (s *KnowledgeIntegrationService) generateDocumentContent(contract *models.Contract) string {
	content := fmt.Sprintf(`合同标题：%s

合同描述：%s

签署状态：已完成
签署时间：%s
合同期限：%s

签署方信息：
`, contract.Title, contract.Description, 
		contract.CompletedAt.Format("2006-01-02 15:04:05"),
		contract.ExpireTime.Format("2006-01-02"))

	// 添加签署人信息
	for _, signer := range contract.Signers {
		content += fmt.Sprintf("- %s (%s) - %s\n", 
			signer.Name, signer.SignerType, signer.SignTime.Format("2006-01-02 15:04:05"))
	}

	if contract.RequireBlockchain && contract.EvidenceURL != "" {
		content += fmt.Sprintf("\n区块链存证：\n- 交易哈希：%s\n- 存证报告：%s\n", 
			contract.BlockchainTxHash, contract.EvidenceURL)
	}

	return content
}

// generateDocumentTags 生成文档标签
func (s *KnowledgeIntegrationService) generateDocumentTags(contract *models.Contract, workflow *models.ContractWorkflow) []string {
	tags := []string{"合同", "已签署", contract.Status}

	// 添加工作流配置的标签
	if workflow.DifyTags != "" {
		var workflowTags []string
		if err := json.Unmarshal([]byte(workflow.DifyTags), &workflowTags); err == nil {
			tags = append(tags, workflowTags...)
		}
	}

	// 根据签署人类型添加标签
	hasPersonSigner := false
	hasCompanySigner := false
	for _, signer := range contract.Signers {
		if signer.SignerType == "person" {
			hasPersonSigner = true
		} else if signer.SignerType == "company" {
			hasCompanySigner = true
		}
	}

	if hasPersonSigner {
		tags = append(tags, "个人签署")
	}
	if hasCompanySigner {
		tags = append(tags, "企业签署")
	}

	if contract.RequireCA {
		tags = append(tags, "CA认证")
	}

	if contract.RequireBlockchain {
		tags = append(tags, "区块链存证")
	}

	return tags
}

// getApproversByRole 根据角色获取审批人
func (s *KnowledgeIntegrationService) getApproversByRole(teamID, role string) ([]models.User, error) {
	var users []models.User
	// 这里需要根据实际的用户角色系统实现
	// 暂时返回空列表
	return users, nil
}

// getDocument 获取文档
func (s *KnowledgeIntegrationService) getDocument(documentID string) (*models.Document, error) {
	var document models.Document
	err := s.db.First(&document, "id = ?", documentID).Error
	return &document, err
}

// checkAllApprovalsCompleted 检查所有审批是否完成
func (s *KnowledgeIntegrationService) checkAllApprovalsCompleted(workflowInstID string) (bool, bool, error) {
	var pendingCount int64
	if err := s.db.Model(&models.ApprovalTask{}).
		Where("workflow_inst_id = ? AND status = ?", workflowInstID, "pending").
		Count(&pendingCount).Error; err != nil {
		return false, false, err
	}

	if pendingCount > 0 {
		return false, false, nil // 还有待审批的任务
	}

	var rejectedCount int64
	if err := s.db.Model(&models.ApprovalTask{}).
		Where("workflow_inst_id = ? AND status = ?", workflowInstID, "rejected").
		Count(&rejectedCount).Error; err != nil {
		return false, false, err
	}

	return true, rejectedCount == 0, nil
}

// getSubmissionByWorkflow 根据工作流ID获取提交记录
func (s *KnowledgeIntegrationService) getSubmissionByWorkflow(workflowInstID string) (*models.KnowledgeSubmission, error) {
	// 这里需要通过工作流实例找到对应的提交记录
	// 暂时使用简化实现
	var submission models.KnowledgeSubmission
	err := s.db.First(&submission, "status = ?", "pending_approval").Error
	return &submission, err
}

// serializeWorkflowInput 序列化工作流输入
func (s *KnowledgeIntegrationService) serializeWorkflowInput(contract *models.Contract, document *models.Document) string {
	data := map[string]interface{}{
		"contract_id":  contract.ID,
		"document_id":  document.ID,
		"contract_title": contract.Title,
		"document_name": document.Name,
	}
	
	jsonData, _ := json.Marshal(data)
	return string(jsonData)
}

// serializeDifyResult 序列化Dify处理结果
func (s *KnowledgeIntegrationService) serializeDifyResult(result *DifyProcessResponse) string {
	jsonData, _ := json.Marshal(result)
	return string(jsonData)
}

// 通知方法 (占位符实现)
func (s *KnowledgeIntegrationService) sendApprovalNotification(approver *models.User, task *models.ApprovalTask, contract *models.Contract) {
	// 实现发送审批通知逻辑
}

func (s *KnowledgeIntegrationService) notifyKnowledgeBaseUpdate(submission *models.KnowledgeSubmission) {
	// 实现知识库更新通知逻辑
}

func (s *KnowledgeIntegrationService) notifyKnowledgeSubmissionRejected(submission *models.KnowledgeSubmission, reason string) {
	// 实现知识库提交拒绝通知逻辑
}

// 需要的模型定义 (占位符)
type WorkflowInstance struct {
	ID            string    `json:"id" gorm:"type:uuid;primary_key;default:gen_random_uuid()"`
	WorkflowDefID string    `json:"workflow_def_id" gorm:"type:uuid;not null"`
	Status        string    `json:"status" gorm:"size:20;default:'pending'"`
	InputData     string    `json:"input_data" gorm:"type:text"`
	CreatedBy     string    `json:"created_by" gorm:"type:uuid;not null"`
	CreatedAt     time.Time `json:"created_at" gorm:"autoCreateTime"`
}

type ApprovalTask struct {
	ID             string     `json:"id" gorm:"type:uuid;primary_key;default:gen_random_uuid()"`
	WorkflowInstID string     `json:"workflow_inst_id" gorm:"type:uuid;not null"`
	Name           string     `json:"name" gorm:"size:255;not null"`
	Description    string     `json:"description" gorm:"type:text"`
	Assignee       string     `json:"assignee" gorm:"type:uuid;not null"`
	Status         string     `json:"status" gorm:"size:20;default:'pending'"`
	Comments       string     `json:"comments" gorm:"type:text"`
	DueDate        time.Time  `json:"due_date"`
	CreatedAt      time.Time  `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt      time.Time  `json:"updated_at" gorm:"autoUpdateTime"`
}

// 依赖服务的占位符定义 (这些已在service.go中定义)
type ApprovalService struct{}
func NewApprovalService() *ApprovalService { return &ApprovalService{} }