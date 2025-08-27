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
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/linux-do/cdk-office/internal/models"
	"gorm.io/gorm"
)

// ShareApprovalService 知识分享审核服务
type ShareApprovalService struct {
	db *gorm.DB
}

// NewShareApprovalService 创建知识分享审核服务
func NewShareApprovalService(db *gorm.DB) *ShareApprovalService {
	return &ShareApprovalService{
		db: db,
	}
}

// SubmitShareApplication 提交分享申请
func (s *ShareApprovalService) SubmitShareApplication(ctx context.Context, req *SubmitShareApplicationRequest) (*models.PersonalKnowledgeShare, error) {
	// 检查知识是否存在且属于用户
	var knowledge models.PersonalKnowledgeBase
	if err := s.db.WithContext(ctx).Where("id = ? AND user_id = ?", req.KnowledgeID, req.UserID).First(&knowledge).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("knowledge not found")
		}
		return nil, fmt.Errorf("failed to find knowledge: %w", err)
	}

	// 检查是否已经有待审批的申请
	var existingShare models.PersonalKnowledgeShare
	if err := s.db.WithContext(ctx).Where("knowledge_id = ? AND status = 'pending'", req.KnowledgeID).First(&existingShare).Error; err == nil {
		return nil, fmt.Errorf("knowledge share application already pending")
	}

	// 创建分享申请记录
	share := &models.PersonalKnowledgeShare{
		KnowledgeID: req.KnowledgeID,
		UserID:      req.UserID,
		TeamID:      req.TeamID,
		ShareReason: req.ShareReason,
		Status:      "pending",
	}

	if err := s.db.WithContext(ctx).Create(share).Error; err != nil {
		return nil, fmt.Errorf("failed to create share application: %w", err)
	}

	// 创建审批工作流
	if req.CreateWorkflow {
		workflowID, err := s.createApprovalWorkflow(ctx, share)
		if err != nil {
			log.Printf("Failed to create approval workflow for share %s: %v", share.ID, err)
			// 不返回错误，因为分享申请已经创建成功
		} else {
			// 更新分享记录关联的审批工作流ID
			s.db.WithContext(ctx).Model(share).Update("approval_id", workflowID)
			log.Printf("Created approval workflow %s for share %s", workflowID, share.ID)
		}
	}

	log.Printf("Created knowledge share application: %s", share.ID)
	return share, nil
}

// createApprovalWorkflow 创建审批工作流
func (s *ShareApprovalService) createApprovalWorkflow(ctx context.Context, share *models.PersonalKnowledgeShare) (string, error) {
	// 获取知识详情用于工作流
	var knowledge models.PersonalKnowledgeBase
	if err := s.db.WithContext(ctx).Where("id = ?", share.KnowledgeID).First(&knowledge).Error; err != nil {
		return "", fmt.Errorf("failed to get knowledge for workflow: %w", err)
	}

	// 构建工作流定义
	workflowDefinition := map[string]interface{}{
		"id":          fmt.Sprintf("knowledge_share_%s", share.ID),
		"name":        "知识库分享审批",
		"description": fmt.Sprintf("审批用户 %s 的知识「%s」分享申请", share.UserID, knowledge.Title),
		"version":     "1.0",
		"steps": []map[string]interface{}{
			{
				"id":   "review",
				"name": "管理员审批",
				"type": "human_task",
				"config": map[string]interface{}{
					"assignees":      []string{"admin"}, // 这里应该根据团队配置获取审批人
					"timeout":        "7d",              // 7天超时
					"allow_reject":   true,
					"require_reason": true,
				},
			},
			{
				"id":   "sync_to_team",
				"name": "同步到团队知识库",
				"type": "system_task",
				"config": map[string]interface{}{
					"action": "sync_knowledge_to_team",
				},
				"condition": "review.result == 'approved'",
			},
		},
		"variables": map[string]interface{}{
			"share_id":     share.ID,
			"knowledge_id": share.KnowledgeID,
			"user_id":      share.UserID,
			"team_id":      share.TeamID,
			"share_reason": share.ShareReason,
		},
	}

	// 序列化工作流定义
	definitionJSON, err := json.Marshal(workflowDefinition)
	if err != nil {
		return "", fmt.Errorf("failed to marshal workflow definition: %w", err)
	}

	// 创建工作流定义记录
	workflow := &models.WorkflowDefinition{
		Name:        "知识库分享审批",
		Description: fmt.Sprintf("知识「%s」的分享审批流程", knowledge.Title),
		Definition:  string(definitionJSON),
		Version:     1,
		Status:      "active",
		CreatedBy:   share.UserID,
	}

	if err := s.db.WithContext(ctx).Create(workflow).Error; err != nil {
		return "", fmt.Errorf("failed to create workflow definition: %w", err)
	}

	// 启动工作流实例
	instance := &models.WorkflowInstance{
		WorkflowID: workflow.ID,
		Name:       fmt.Sprintf("知识分享审批 - %s", knowledge.Title),
		Status:     "running",
		Variables:  string(definitionJSON), // 复用变量
		CreatedBy:  share.UserID,
	}

	if err := s.db.WithContext(ctx).Create(instance).Error; err != nil {
		return "", fmt.Errorf("failed to create workflow instance: %w", err)
	}

	// 创建第一个任务（管理员审批）
	task := &models.WorkflowTask{
		InstanceID: instance.ID,
		StepID:     "review",
		Name:       "管理员审批",
		Type:       "human_task",
		Status:     "pending",
		Assignee:   "admin", // 这里应该根据团队配置分配给具体的管理员
		Variables:  string(definitionJSON),
		DueDate:    time.Now().Add(7 * 24 * time.Hour), // 7天后到期
	}

	if err := s.db.WithContext(ctx).Create(task).Error; err != nil {
		return "", fmt.Errorf("failed to create workflow task: %w", err)
	}

	return instance.ID, nil
}

// ListShareApplications 列出分享申请
func (s *ShareApprovalService) ListShareApplications(ctx context.Context, req *ListShareApplicationsRequest) (*ListShareApplicationsResponse, error) {
	query := s.db.WithContext(ctx).Model(&models.PersonalKnowledgeShare{})

	// 根据角色筛选
	if req.Role == "admin" {
		// 管理员可以看到指定团队的所有申请
		if req.TeamID != "" {
			query = query.Where("team_id = ?", req.TeamID)
		}
	} else {
		// 普通用户只能看到自己的申请
		query = query.Where("user_id = ?", req.UserID)
	}

	// 状态筛选
	if req.Status != "" {
		query = query.Where("status = ?", req.Status)
	}

	// 计算总数
	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, fmt.Errorf("failed to count share applications: %w", err)
	}

	// 分页
	offset := (req.Page - 1) * req.PageSize
	query = query.Offset(offset).Limit(req.PageSize)

	// 排序
	query = query.Order("created_at DESC")

	var applications []models.PersonalKnowledgeShare
	if err := query.Find(&applications).Error; err != nil {
		return nil, fmt.Errorf("failed to list share applications: %w", err)
	}

	// 获取关联的知识信息
	var enrichedApplications []ShareApplicationWithKnowledge
	for _, app := range applications {
		var knowledge models.PersonalKnowledgeBase
		if err := s.db.WithContext(ctx).Where("id = ?", app.KnowledgeID).First(&knowledge).Error; err != nil {
			log.Printf("Failed to get knowledge %s for application %s: %v", app.KnowledgeID, app.ID, err)
			continue
		}

		enrichedApplications = append(enrichedApplications, ShareApplicationWithKnowledge{
			PersonalKnowledgeShare: app,
			Knowledge:              &knowledge,
		})
	}

	return &ListShareApplicationsResponse{
		Applications: enrichedApplications,
		Total:        total,
		Page:         req.Page,
		PageSize:     req.PageSize,
	}, nil
}

// ReviewShareApplication 审批分享申请
func (s *ShareApprovalService) ReviewShareApplication(ctx context.Context, req *ReviewShareApplicationRequest) (*models.PersonalKnowledgeShare, error) {
	// 获取分享申请
	var share models.PersonalKnowledgeShare
	if err := s.db.WithContext(ctx).Where("id = ?", req.ShareID).First(&share).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("share application not found")
		}
		return nil, fmt.Errorf("failed to get share application: %w", err)
	}

	// 检查申请状态
	if share.Status != "pending" {
		return nil, fmt.Errorf("share application is not pending, current status: %s", share.Status)
	}

	// 更新审批结果
	updateData := map[string]interface{}{
		"status":        req.Decision,
		"reviewer_id":   req.ReviewerID,
		"review_reason": req.ReviewReason,
		"reviewed_at":   time.Now(),
	}

	if err := s.db.WithContext(ctx).Model(&share).Updates(updateData).Error; err != nil {
		return nil, fmt.Errorf("failed to update share application: %w", err)
	}

	// 如果审批通过，执行分享操作
	if req.Decision == "approved" {
		if err := s.executeShareToTeam(ctx, &share); err != nil {
			log.Printf("Failed to execute share to team for %s: %v", share.ID, err)
			// 将状态改为失败
			s.db.WithContext(ctx).Model(&share).Update("status", "failed")
			return nil, fmt.Errorf("failed to execute share operation: %w", err)
		}
	}

	// 更新关联的工作流任务状态
	if share.ApprovalID != "" {
		s.updateWorkflowTaskStatus(ctx, share.ApprovalID, req.Decision, req.ReviewReason)
	}

	// 重新获取更新后的数据
	if err := s.db.WithContext(ctx).Where("id = ?", req.ShareID).First(&share).Error; err != nil {
		return nil, fmt.Errorf("failed to get updated share application: %w", err)
	}

	log.Printf("Reviewed share application %s: %s", req.ShareID, req.Decision)
	return &share, nil
}

// executeShareToTeam 执行分享到团队知识库的操作
func (s *ShareApprovalService) executeShareToTeam(ctx context.Context, share *models.PersonalKnowledgeShare) error {
	// 获取原始知识
	var knowledge models.PersonalKnowledgeBase
	if err := s.db.WithContext(ctx).Where("id = ?", share.KnowledgeID).First(&knowledge).Error; err != nil {
		return fmt.Errorf("failed to get original knowledge: %w", err)
	}

	// 创建团队文档记录
	teamDocument := &models.Document{
		TeamID:      share.TeamID,
		Name:        knowledge.Title,
		Description: knowledge.Description,
		FileName:    fmt.Sprintf("%s.md", knowledge.Title),
		FilePath:    fmt.Sprintf("team_knowledge/%s/%s.md", share.TeamID, knowledge.ID),
		FileSize:    int64(len(knowledge.Content)),
		FileType:    "markdown",
		MimeType:    "text/markdown",
		Status:      "active",
		Version:     1,
		CreatedBy:   share.UserID,
		Tags:        knowledge.Tags,
	}

	if err := s.db.WithContext(ctx).Create(teamDocument).Error; err != nil {
		return fmt.Errorf("failed to create team document: %w", err)
	}

	// 创建Dify同步记录
	difySync := &models.DifyDocumentSync{
		DocumentID:     teamDocument.ID,
		DifyDocumentID: "", // 实际同步时会填充
		DatasetID:      "", // 需要根据团队配置获取
		Title:          knowledge.Title,
		Content:        knowledge.Content,
		DocumentType:   "markdown",
		TeamID:         share.TeamID,
		SyncStatus:     "pending",
		CreatedBy:      share.UserID,
	}

	if err := s.db.WithContext(ctx).Create(difySync).Error; err != nil {
		return fmt.Errorf("failed to create dify sync record: %w", err)
	}

	// 更新原始知识的分享状态
	updateData := map[string]interface{}{
		"is_shared": true,
		"shared_at": time.Now(),
	}
	if err := s.db.WithContext(ctx).Model(&knowledge).Updates(updateData).Error; err != nil {
		log.Printf("Failed to update original knowledge share status: %v", err)
	}

	// TODO: 触发实际的Dify同步操作
	// 这里应该调用Dify API将文档同步到知识库

	log.Printf("Successfully shared knowledge %s to team %s", knowledge.ID, share.TeamID)
	return nil
}

// updateWorkflowTaskStatus 更新工作流任务状态
func (s *ShareApprovalService) updateWorkflowTaskStatus(ctx context.Context, instanceID, decision, reason string) {
	// 查找相关的工作流任务
	var task models.WorkflowTask
	if err := s.db.WithContext(ctx).Where("instance_id = ? AND step_id = 'review' AND status = 'pending'", instanceID).First(&task).Error; err != nil {
		log.Printf("Failed to find workflow task for instance %s: %v", instanceID, err)
		return
	}

	// 更新任务状态
	taskStatus := "completed"
	if decision == "rejected" {
		taskStatus = "rejected"
	}

	updateData := map[string]interface{}{
		"status":       taskStatus,
		"completed_at": time.Now(),
		"result":       decision,
		"comments":     reason,
	}

	if err := s.db.WithContext(ctx).Model(&task).Updates(updateData).Error; err != nil {
		log.Printf("Failed to update workflow task status: %v", err)
		return
	}

	// 更新工作流实例状态
	instanceStatus := "completed"
	if decision == "rejected" {
		instanceStatus = "terminated"
	}

	if err := s.db.WithContext(ctx).Model(&models.WorkflowInstance{}).
		Where("id = ?", instanceID).
		Updates(map[string]interface{}{
			"status":       instanceStatus,
			"completed_at": time.Now(),
		}).Error; err != nil {
		log.Printf("Failed to update workflow instance status: %v", err)
	}

	log.Printf("Updated workflow task %s status to %s", task.ID, taskStatus)
}

// GetShareApplicationDetail 获取分享申请详情
func (s *ShareApprovalService) GetShareApplicationDetail(ctx context.Context, shareID, userID string, role string) (*ShareApplicationDetail, error) {
	var share models.PersonalKnowledgeShare
	query := s.db.WithContext(ctx).Where("id = ?", shareID)

	// 权限检查
	if role != "admin" {
		query = query.Where("user_id = ?", userID)
	}

	if err := query.First(&share).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("share application not found")
		}
		return nil, fmt.Errorf("failed to get share application: %w", err)
	}

	// 获取关联的知识
	var knowledge models.PersonalKnowledgeBase
	if err := s.db.WithContext(ctx).Where("id = ?", share.KnowledgeID).First(&knowledge).Error; err != nil {
		return nil, fmt.Errorf("failed to get knowledge: %w", err)
	}

	// 获取工作流信息（如果有）
	var workflowTasks []models.WorkflowTask
	if share.ApprovalID != "" {
		s.db.WithContext(ctx).Where("instance_id = ?", share.ApprovalID).Order("created_at ASC").Find(&workflowTasks)
	}

	detail := &ShareApplicationDetail{
		PersonalKnowledgeShare: share,
		Knowledge:              &knowledge,
		WorkflowTasks:          workflowTasks,
	}

	return detail, nil
}

// GetShareStatistics 获取分享统计信息
func (s *ShareApprovalService) GetShareStatistics(ctx context.Context, userID, teamID string, role string) (*ShareStatistics, error) {
	var stats ShareStatistics

	baseQuery := s.db.WithContext(ctx).Model(&models.PersonalKnowledgeShare{})

	// 根据角色设置查询范围
	if role == "admin" && teamID != "" {
		baseQuery = baseQuery.Where("team_id = ?", teamID)
	} else {
		baseQuery = baseQuery.Where("user_id = ?", userID)
	}

	// 总申请数
	if err := baseQuery.Count(&stats.TotalApplications).Error; err != nil {
		return nil, fmt.Errorf("failed to count total applications: %w", err)
	}

	// 待审批数
	if err := baseQuery.Where("status = 'pending'").Count(&stats.PendingApplications).Error; err != nil {
		return nil, fmt.Errorf("failed to count pending applications: %w", err)
	}

	// 已通过数
	if err := baseQuery.Where("status = 'approved'").Count(&stats.ApprovedApplications).Error; err != nil {
		return nil, fmt.Errorf("failed to count approved applications: %w", err)
	}

	// 已拒绝数
	if err := baseQuery.Where("status = 'rejected'").Count(&stats.RejectedApplications).Error; err != nil {
		return nil, fmt.Errorf("failed to count rejected applications: %w", err)
	}

	// 本周新增
	weekAgo := time.Now().AddDate(0, 0, -7)
	if err := baseQuery.Where("created_at >= ?", weekAgo).Count(&stats.WeeklyApplications).Error; err != nil {
		return nil, fmt.Errorf("failed to count weekly applications: %w", err)
	}

	return &stats, nil
}
