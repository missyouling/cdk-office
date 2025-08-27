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

package workflows

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/linux-do/cdk-office/internal/models"
	"gorm.io/gorm"
)

// WorkflowExecutor 工作流执行器
type WorkflowExecutor struct {
	db               *gorm.DB
	runningWorkflows map[string]*ExecutionContext
	mutex            sync.RWMutex
	stopChan         chan struct{}
	ticker           *time.Ticker
}

// ExecutionContext 执行上下文
type ExecutionContext struct {
	Instance   *WorkflowInstance
	Definition *WorkflowDefinition
	Context    context.Context
	Cancel     context.CancelFunc
	LastUpdate time.Time
}

// NewWorkflowExecutor 创建工作流执行器
func NewWorkflowExecutor(db *gorm.DB) *WorkflowExecutor {
	executor := &WorkflowExecutor{
		db:               db,
		runningWorkflows: make(map[string]*ExecutionContext),
		stopChan:         make(chan struct{}),
		ticker:           time.NewTicker(30 * time.Second), // 每30秒检查一次
	}

	// 启动监控协程
	go executor.monitorWorkflows()

	// 恢复正在运行的工作流
	go executor.recoverRunningWorkflows()

	return executor
}

// AddRunningWorkflow 添加正在运行的工作流
func (e *WorkflowExecutor) AddRunningWorkflow(instance *WorkflowInstance, definition *WorkflowDefinition) {
	e.mutex.Lock()
	defer e.mutex.Unlock()

	ctx, cancel := context.WithCancel(context.Background())

	execCtx := &ExecutionContext{
		Instance:   instance,
		Definition: definition,
		Context:    ctx,
		Cancel:     cancel,
		LastUpdate: time.Now(),
	}

	e.runningWorkflows[instance.ID] = execCtx
	log.Printf("Added running workflow: %s (ID: %s)", instance.Name, instance.ID)
}

// RemoveRunningWorkflow 移除正在运行的工作流
func (e *WorkflowExecutor) RemoveRunningWorkflow(instanceID string) {
	e.mutex.Lock()
	defer e.mutex.Unlock()

	if execCtx, exists := e.runningWorkflows[instanceID]; exists {
		execCtx.Cancel()
		delete(e.runningWorkflows, instanceID)
		log.Printf("Removed running workflow: %s", instanceID)
	}
}

// GetRunningWorkflow 获取正在运行的工作流
func (e *WorkflowExecutor) GetRunningWorkflow(instanceID string) (*ExecutionContext, bool) {
	e.mutex.RLock()
	defer e.mutex.RUnlock()

	execCtx, exists := e.runningWorkflows[instanceID]
	return execCtx, exists
}

// ListRunningWorkflows 获取所有正在运行的工作流
func (e *WorkflowExecutor) ListRunningWorkflows() []*ExecutionContext {
	e.mutex.RLock()
	defer e.mutex.RUnlock()

	var workflows []*ExecutionContext
	for _, execCtx := range e.runningWorkflows {
		workflows = append(workflows, execCtx)
	}

	return workflows
}

// PauseWorkflow 暂停工作流
func (e *WorkflowExecutor) PauseWorkflow(instanceID string) error {
	e.mutex.Lock()
	defer e.mutex.Unlock()

	execCtx, exists := e.runningWorkflows[instanceID]
	if !exists {
		return fmt.Errorf("workflow not found: %s", instanceID)
	}

	// 更新数据库状态
	execCtx.Instance.Status = "paused"
	if err := e.db.Save(execCtx.Instance).Error; err != nil {
		return fmt.Errorf("failed to pause workflow: %w", err)
	}

	log.Printf("Paused workflow: %s (ID: %s)", execCtx.Instance.Name, instanceID)
	return nil
}

// ResumeWorkflow 恢复工作流
func (e *WorkflowExecutor) ResumeWorkflow(instanceID string) error {
	e.mutex.Lock()
	defer e.mutex.Unlock()

	execCtx, exists := e.runningWorkflows[instanceID]
	if !exists {
		return fmt.Errorf("workflow not found: %s", instanceID)
	}

	if execCtx.Instance.Status != "paused" {
		return fmt.Errorf("workflow is not paused: %s", instanceID)
	}

	// 更新数据库状态
	execCtx.Instance.Status = "running"
	if err := e.db.Save(execCtx.Instance).Error; err != nil {
		return fmt.Errorf("failed to resume workflow: %w", err)
	}

	log.Printf("Resumed workflow: %s (ID: %s)", execCtx.Instance.Name, instanceID)
	return nil
}

// StopWorkflow 停止工作流
func (e *WorkflowExecutor) StopWorkflow(instanceID string, reason string) error {
	e.mutex.Lock()
	defer e.mutex.Unlock()

	execCtx, exists := e.runningWorkflows[instanceID]
	if !exists {
		return fmt.Errorf("workflow not found: %s", instanceID)
	}

	// 取消执行上下文
	execCtx.Cancel()

	// 更新数据库状态
	now := time.Now()
	execCtx.Instance.Status = "cancelled"
	execCtx.Instance.CompletedAt = &now
	if reason != "" {
		execCtx.Instance.OutputData = map[string]interface{}{
			"cancellation_reason": reason,
		}
	}

	if err := e.db.Save(execCtx.Instance).Error; err != nil {
		return fmt.Errorf("failed to stop workflow: %w", err)
	}

	// 从运行列表中移除
	delete(e.runningWorkflows, instanceID)

	log.Printf("Stopped workflow: %s (ID: %s) - %s", execCtx.Instance.Name, instanceID, reason)
	return nil
}

// UpdateWorkflowProgress 更新工作流进度
func (e *WorkflowExecutor) UpdateWorkflowProgress(instanceID, currentStep string, variables map[string]interface{}) error {
	e.mutex.Lock()
	defer e.mutex.Unlock()

	execCtx, exists := e.runningWorkflows[instanceID]
	if !exists {
		return fmt.Errorf("workflow not found: %s", instanceID)
	}

	// 更新实例信息
	execCtx.Instance.CurrentStep = currentStep
	if variables != nil {
		execCtx.Instance.Variables = variables
	}
	execCtx.LastUpdate = time.Now()

	// 保存到数据库
	if err := e.db.Save(execCtx.Instance).Error; err != nil {
		return fmt.Errorf("failed to update workflow progress: %w", err)
	}

	return nil
}

// CheckTimeout 检查工作流超时
func (e *WorkflowExecutor) CheckTimeout(instanceID string, timeout time.Duration) bool {
	e.mutex.RLock()
	defer e.mutex.RUnlock()

	execCtx, exists := e.runningWorkflows[instanceID]
	if !exists {
		return false
	}

	return time.Since(execCtx.LastUpdate) > timeout
}

// GetWorkflowStatistics 获取工作流统计信息
func (e *WorkflowExecutor) GetWorkflowStatistics() map[string]interface{} {
	e.mutex.RLock()
	defer e.mutex.RUnlock()

	stats := map[string]interface{}{
		"total_running":     len(e.runningWorkflows),
		"running_workflows": make([]map[string]interface{}, 0),
	}

	for instanceID, execCtx := range e.runningWorkflows {
		workflowInfo := map[string]interface{}{
			"instance_id":  instanceID,
			"name":         execCtx.Instance.Name,
			"status":       execCtx.Instance.Status,
			"current_step": execCtx.Instance.CurrentStep,
			"started_at":   execCtx.Instance.StartedAt,
			"last_update":  execCtx.LastUpdate,
		}
		stats["running_workflows"] = append(stats["running_workflows"].([]map[string]interface{}), workflowInfo)
	}

	return stats
}

// RetryFailedStep 重试失败的步骤
func (e *WorkflowExecutor) RetryFailedStep(instanceID, stepID string) error {
	// 获取步骤实例
	var stepInstance StepInstance
	if err := e.db.Where("workflow_id = ? AND step_def_id = ? AND status = ?", instanceID, stepID, "failed").First(&stepInstance).Error; err != nil {
		return fmt.Errorf("failed step not found: %w", err)
	}

	// 更新步骤状态
	stepInstance.Status = "pending"
	stepInstance.RetryCount++
	stepInstance.ErrorMessage = ""

	if err := e.db.Save(&stepInstance).Error; err != nil {
		return fmt.Errorf("failed to update step for retry: %w", err)
	}

	log.Printf("Retrying failed step: %s (workflow: %s)", stepID, instanceID)

	// 触发步骤重新执行
	// 这里需要通知工作流引擎重新执行该步骤

	return nil
}

// HandleApprovalResult 处理审批结果
func (e *WorkflowExecutor) HandleApprovalResult(approvalID string, approved bool, comments string) error {
	// 查找相关的审批流程
	var approval models.ApprovalProcess
	if err := e.db.First(&approval, "id = ?", approvalID).Error; err != nil {
		return fmt.Errorf("approval not found: %w", err)
	}

	// 查找相关的步骤实例
	var stepInstance StepInstance
	if err := e.db.Where("type = ? AND status = ?", "approval", "pending").First(&stepInstance).Error; err != nil {
		log.Printf("No pending approval step found for approval: %s", approvalID)
		return nil
	}

	// 更新步骤状态
	now := time.Now()
	if approved {
		stepInstance.Status = "completed"
	} else {
		stepInstance.Status = "failed"
		stepInstance.ErrorMessage = "Approval rejected: " + comments
	}
	stepInstance.CompletedAt = &now
	stepInstance.OutputData = map[string]interface{}{
		"approved": approved,
		"comments": comments,
	}

	if err := e.db.Save(&stepInstance).Error; err != nil {
		return fmt.Errorf("failed to update step instance: %w", err)
	}

	// 通知工作流引擎继续执行
	execCtx, exists := e.GetRunningWorkflow(stepInstance.WorkflowID)
	if exists {
		// 继续执行下一步
		go e.continueWorkflowExecution(execCtx, approved)
	}

	log.Printf("Handled approval result: %s (approved: %v)", approvalID, approved)
	return nil
}

// monitorWorkflows 监控工作流执行
func (e *WorkflowExecutor) monitorWorkflows() {
	for {
		select {
		case <-e.stopChan:
			return
		case <-e.ticker.C:
			e.performHealthCheck()
		}
	}
}

// performHealthCheck 执行健康检查
func (e *WorkflowExecutor) performHealthCheck() {
	e.mutex.RLock()
	workflows := make([]*ExecutionContext, 0, len(e.runningWorkflows))
	for _, execCtx := range e.runningWorkflows {
		workflows = append(workflows, execCtx)
	}
	e.mutex.RUnlock()

	for _, execCtx := range workflows {
		// 检查工作流是否超时
		if time.Since(execCtx.LastUpdate) > 1*time.Hour {
			log.Printf("Workflow %s appears to be stuck, last update: %v", execCtx.Instance.ID, execCtx.LastUpdate)

			// 可以选择自动取消超时的工作流
			// e.StopWorkflow(execCtx.Instance.ID, "timeout")
		}

		// 检查工作流状态是否与数据库一致
		var dbInstance WorkflowInstance
		if err := e.db.First(&dbInstance, "id = ?", execCtx.Instance.ID).Error; err == nil {
			if dbInstance.Status != execCtx.Instance.Status {
				log.Printf("Workflow %s status mismatch: memory=%s, db=%s", execCtx.Instance.ID, execCtx.Instance.Status, dbInstance.Status)
				// 同步状态
				execCtx.Instance.Status = dbInstance.Status
			}
		}
	}
}

// recoverRunningWorkflows 恢复正在运行的工作流
func (e *WorkflowExecutor) recoverRunningWorkflows() {
	var instances []WorkflowInstance
	if err := e.db.Where("status IN ?", []string{"running", "pending"}).Find(&instances).Error; err != nil {
		log.Printf("Failed to recover running workflows: %v", err)
		return
	}

	for _, instance := range instances {
		// 获取工作流定义
		var defRecord models.WorkflowDefinition
		if err := e.db.First(&defRecord, "id = ?", instance.WorkflowDefID).Error; err != nil {
			log.Printf("Failed to load workflow definition for instance %s: %v", instance.ID, err)
			continue
		}

		var definition WorkflowDefinition
		if err := json.Unmarshal([]byte(defRecord.Definition), &definition); err != nil {
			log.Printf("Failed to unmarshal workflow definition for instance %s: %v", instance.ID, err)
			continue
		}

		// 添加到运行列表
		e.AddRunningWorkflow(&instance, &definition)
		log.Printf("Recovered running workflow: %s (ID: %s)", instance.Name, instance.ID)
	}
}

// continueWorkflowExecution 继续工作流执行
func (e *WorkflowExecutor) continueWorkflowExecution(execCtx *ExecutionContext, approved bool) {
	// 这里应该实现继续执行工作流的逻辑
	// 根据审批结果决定下一步操作

	if approved {
		log.Printf("Continuing workflow %s after approval", execCtx.Instance.ID)
		// 继续执行下一步
	} else {
		log.Printf("Stopping workflow %s due to rejection", execCtx.Instance.ID)
		// 停止工作流或执行拒绝分支
		e.StopWorkflow(execCtx.Instance.ID, "approval rejected")
	}
}

// Shutdown 关闭执行器
func (e *WorkflowExecutor) Shutdown() {
	close(e.stopChan)
	e.ticker.Stop()

	e.mutex.Lock()
	defer e.mutex.Unlock()

	// 取消所有正在运行的工作流
	for instanceID, execCtx := range e.runningWorkflows {
		execCtx.Cancel()
		log.Printf("Cancelled workflow on shutdown: %s", instanceID)
	}

	e.runningWorkflows = make(map[string]*ExecutionContext)
	log.Println("Workflow executor shutdown completed")
}
