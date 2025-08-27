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
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/google/uuid"
	"github.com/linux-do/cdk-office/internal/models"
	"gorm.io/gorm"
)

// WorkflowEngine 工作流引擎
type WorkflowEngine struct {
	db             *gorm.DB
	executor       *WorkflowExecutor
	activityRunner *ActivityRunner
}

// NewWorkflowEngine 创建工作流引擎
func NewWorkflowEngine(db *gorm.DB) *WorkflowEngine {
	engine := &WorkflowEngine{
		db: db,
	}

	engine.executor = NewWorkflowExecutor(db)
	engine.activityRunner = NewActivityRunner(db)

	// 注册内置活动
	engine.registerBuiltinActivities()

	return engine
}

// WorkflowDefinition 工作流定义
type WorkflowDefinition struct {
	ID          string                 `json:"id"`
	Name        string                 `json:"name"`
	Description string                 `json:"description"`
	Steps       []WorkflowStep         `json:"steps"`
	Variables   map[string]interface{} `json:"variables"`
	CreatedBy   string                 `json:"created_by"`
	TeamID      string                 `json:"team_id"`
	CreatedAt   time.Time              `json:"created_at"`
}

// WorkflowStep 工作流步骤
type WorkflowStep struct {
	ID          string                 `json:"id"`
	Name        string                 `json:"name"`
	Type        string                 `json:"type"` // approval, condition, activity, parallel
	Config      map[string]interface{} `json:"config"`
	NextSteps   []string               `json:"next_steps"`
	Conditions  []StepCondition        `json:"conditions"`
	Timeout     time.Duration          `json:"timeout"`
	RetryPolicy *RetryPolicy           `json:"retry_policy"`
}

// StepCondition 步骤条件
type StepCondition struct {
	Field    string      `json:"field"`
	Operator string      `json:"operator"` // eq, ne, gt, lt, contains, in
	Value    interface{} `json:"value"`
}

// RetryPolicy 重试策略
type RetryPolicy struct {
	MaxRetries  int           `json:"max_retries"`
	Delay       time.Duration `json:"delay"`
	BackoffRate float64       `json:"backoff_rate"`
}

// WorkflowInstance 工作流实例
type WorkflowInstance struct {
	ID            string                 `json:"id" gorm:"type:uuid;primary_key"`
	WorkflowDefID string                 `json:"workflow_def_id" gorm:"type:uuid;not null"`
	Name          string                 `json:"name" gorm:"size:255"`
	Status        string                 `json:"status"` // pending, running, completed, failed, cancelled
	CurrentStep   string                 `json:"current_step"`
	InputData     map[string]interface{} `json:"input_data" gorm:"type:jsonb"`
	OutputData    map[string]interface{} `json:"output_data" gorm:"type:jsonb"`
	Variables     map[string]interface{} `json:"variables" gorm:"type:jsonb"`
	StartedAt     time.Time              `json:"started_at"`
	CompletedAt   *time.Time             `json:"completed_at"`
	CreatedBy     string                 `json:"created_by" gorm:"type:uuid"`
	TeamID        string                 `json:"team_id" gorm:"type:uuid"`
	CreatedAt     time.Time              `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt     time.Time              `json:"updated_at" gorm:"autoUpdateTime"`
}

// StepInstance 步骤实例
type StepInstance struct {
	ID           string                 `json:"id" gorm:"type:uuid;primary_key"`
	WorkflowID   string                 `json:"workflow_id" gorm:"type:uuid;not null"`
	StepDefID    string                 `json:"step_def_id"`
	Name         string                 `json:"name" gorm:"size:255"`
	Type         string                 `json:"type"`
	Status       string                 `json:"status"` // pending, running, completed, failed, skipped
	InputData    map[string]interface{} `json:"input_data" gorm:"type:jsonb"`
	OutputData   map[string]interface{} `json:"output_data" gorm:"type:jsonb"`
	AssignedTo   string                 `json:"assigned_to" gorm:"type:uuid"`
	StartedAt    time.Time              `json:"started_at"`
	CompletedAt  *time.Time             `json:"completed_at"`
	ErrorMessage string                 `json:"error_message" gorm:"type:text"`
	RetryCount   int                    `json:"retry_count"`
	CreatedAt    time.Time              `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt    time.Time              `json:"updated_at" gorm:"autoUpdateTime"`
}

// CreateWorkflow 创建工作流定义
func (e *WorkflowEngine) CreateWorkflow(definition *WorkflowDefinition) error {
	// 验证工作流定义
	if err := e.validateWorkflowDefinition(definition); err != nil {
		return fmt.Errorf("invalid workflow definition: %w", err)
	}

	// 保存到数据库
	definitionJSON, err := json.Marshal(definition)
	if err != nil {
		return fmt.Errorf("failed to marshal workflow definition: %w", err)
	}

	workflowRecord := &models.WorkflowDefinition{
		Name:        definition.Name,
		Description: definition.Description,
		Definition:  string(definitionJSON),
		CreatedBy:   definition.CreatedBy,
		TeamID:      definition.TeamID,
	}

	if err := e.db.Create(workflowRecord).Error; err != nil {
		return fmt.Errorf("failed to save workflow definition: %w", err)
	}

	definition.ID = workflowRecord.ID
	log.Printf("Created workflow definition: %s (ID: %s)", definition.Name, definition.ID)
	return nil
}

// StartWorkflow 启动工作流实例
func (e *WorkflowEngine) StartWorkflow(workflowDefID string, inputData map[string]interface{}, createdBy, teamID string) (*WorkflowInstance, error) {
	// 获取工作流定义
	var workflowRecord models.WorkflowDefinition
	if err := e.db.First(&workflowRecord, "id = ?", workflowDefID).Error; err != nil {
		return nil, fmt.Errorf("workflow definition not found: %w", err)
	}

	var definition WorkflowDefinition
	if err := json.Unmarshal([]byte(workflowRecord.Definition), &definition); err != nil {
		return nil, fmt.Errorf("failed to unmarshal workflow definition: %w", err)
	}

	// 创建工作流实例
	instance := &WorkflowInstance{
		ID:            uuid.New().String(),
		WorkflowDefID: workflowDefID,
		Name:          definition.Name,
		Status:        "running",
		InputData:     inputData,
		Variables:     definition.Variables,
		StartedAt:     time.Now(),
		CreatedBy:     createdBy,
		TeamID:        teamID,
	}

	// 保存实例
	if err := e.db.Create(instance).Error; err != nil {
		return nil, fmt.Errorf("failed to create workflow instance: %w", err)
	}

	// 启动第一个步骤
	if len(definition.Steps) > 0 {
		go e.executeNextStep(instance, &definition, definition.Steps[0].ID)
	}

	log.Printf("Started workflow instance: %s (ID: %s)", instance.Name, instance.ID)
	return instance, nil
}

// executeNextStep 执行下一个步骤
func (e *WorkflowEngine) executeNextStep(instance *WorkflowInstance, definition *WorkflowDefinition, stepID string) {
	// 查找步骤定义
	var step *WorkflowStep
	for _, s := range definition.Steps {
		if s.ID == stepID {
			step = &s
			break
		}
	}

	if step == nil {
		e.failWorkflow(instance, fmt.Sprintf("step not found: %s", stepID))
		return
	}

	// 创建步骤实例
	stepInstance := &StepInstance{
		ID:         uuid.New().String(),
		WorkflowID: instance.ID,
		StepDefID:  step.ID,
		Name:       step.Name,
		Type:       step.Type,
		Status:     "running",
		InputData:  instance.Variables,
		StartedAt:  time.Now(),
	}

	// 更新工作流当前步骤
	instance.CurrentStep = step.ID
	e.db.Save(instance)

	// 保存步骤实例
	if err := e.db.Create(stepInstance).Error; err != nil {
		e.failWorkflow(instance, fmt.Sprintf("failed to create step instance: %v", err))
		return
	}

	// 执行步骤
	if err := e.executeStep(instance, definition, step, stepInstance); err != nil {
		e.failStep(stepInstance, err.Error())
		e.failWorkflow(instance, fmt.Sprintf("step execution failed: %v", err))
		return
	}
}

// executeStep 执行具体步骤
func (e *WorkflowEngine) executeStep(instance *WorkflowInstance, definition *WorkflowDefinition, step *WorkflowStep, stepInstance *StepInstance) error {
	switch step.Type {
	case "approval":
		return e.executeApprovalStep(instance, step, stepInstance)
	case "condition":
		return e.executeConditionStep(instance, definition, step, stepInstance)
	case "activity":
		return e.executeActivityStep(instance, step, stepInstance)
	case "parallel":
		return e.executeParallelStep(instance, definition, step, stepInstance)
	default:
		return fmt.Errorf("unknown step type: %s", step.Type)
	}
}

// executeApprovalStep 执行审批步骤
func (e *WorkflowEngine) executeApprovalStep(instance *WorkflowInstance, step *WorkflowStep, stepInstance *StepInstance) error {
	approvers, ok := step.Config["approvers"].([]interface{})
	if !ok {
		return fmt.Errorf("approvers not specified in approval step")
	}

	// 创建审批任务
	for _, approver := range approvers {
		approverID, ok := approver.(string)
		if !ok {
			continue
		}

		approval := &models.ApprovalProcess{
			TeamID:       instance.TeamID,
			Name:         fmt.Sprintf("%s - %s", instance.Name, step.Name),
			Description:  fmt.Sprintf("Approval required for workflow: %s", instance.Name),
			RequestorID:  instance.CreatedBy,
			ApproverID:   approverID,
			Status:       "pending",
			ApprovalType: "workflow_step",
			CreatedBy:    instance.CreatedBy,
		}

		if err := e.db.Create(approval).Error; err != nil {
			return fmt.Errorf("failed to create approval task: %w", err)
		}

		// 发送通知
		e.sendApprovalNotification(approval)
	}

	// 等待审批完成（通过回调处理）
	stepInstance.Status = "pending"
	stepInstance.AssignedTo = approvers[0].(string) // 分配给第一个审批人
	e.db.Save(stepInstance)

	return nil
}

// executeConditionStep 执行条件步骤
func (e *WorkflowEngine) executeConditionStep(instance *WorkflowInstance, definition *WorkflowDefinition, step *WorkflowStep, stepInstance *StepInstance) error {
	// 评估条件
	conditionMet := true
	for _, condition := range step.Conditions {
		if !e.evaluateCondition(condition, instance.Variables) {
			conditionMet = false
			break
		}
	}

	// 标记步骤完成
	now := time.Now()
	stepInstance.Status = "completed"
	stepInstance.CompletedAt = &now
	stepInstance.OutputData = map[string]interface{}{
		"condition_met": conditionMet,
	}
	e.db.Save(stepInstance)

	// 根据条件结果决定下一步骤
	var nextSteps []string
	if conditionMet {
		nextSteps = step.NextSteps
	} else {
		// 寻找失败分支
		if failureSteps, ok := step.Config["failure_steps"].([]string); ok {
			nextSteps = failureSteps
		}
	}

	// 执行下一步骤
	if len(nextSteps) > 0 {
		for _, nextStepID := range nextSteps {
			go e.executeNextStep(instance, definition, nextStepID)
		}
	} else {
		// 没有下一步骤，完成工作流
		e.completeWorkflow(instance)
	}

	return nil
}

// executeActivityStep 执行活动步骤
func (e *WorkflowEngine) executeActivityStep(instance *WorkflowInstance, step *WorkflowStep, stepInstance *StepInstance) error {
	activityName, ok := step.Config["activity"].(string)
	if !ok {
		return fmt.Errorf("activity name not specified")
	}

	// 执行活动
	result, err := e.activityRunner.ExecuteActivity(activityName, step.Config, instance.Variables)
	if err != nil {
		return fmt.Errorf("activity execution failed: %w", err)
	}

	// 更新步骤实例
	now := time.Now()
	stepInstance.Status = "completed"
	stepInstance.CompletedAt = &now
	stepInstance.OutputData = result
	e.db.Save(stepInstance)

	// 更新工作流变量
	if output, ok := result["output"].(map[string]interface{}); ok {
		for key, value := range output {
			instance.Variables[key] = value
		}
		e.db.Save(instance)
	}

	return nil
}

// executeParallelStep 执行并行步骤
func (e *WorkflowEngine) executeParallelStep(instance *WorkflowInstance, definition *WorkflowDefinition, step *WorkflowStep, stepInstance *StepInstance) error {
	parallelSteps, ok := step.Config["parallel_steps"].([]interface{})
	if !ok {
		return fmt.Errorf("parallel steps not specified")
	}

	// 启动所有并行步骤
	for _, parallelStepID := range parallelSteps {
		stepID, ok := parallelStepID.(string)
		if !ok {
			continue
		}
		go e.executeNextStep(instance, definition, stepID)
	}

	// 标记并行步骤启动完成
	now := time.Now()
	stepInstance.Status = "completed"
	stepInstance.CompletedAt = &now
	e.db.Save(stepInstance)

	return nil
}

// evaluateCondition 评估条件
func (e *WorkflowEngine) evaluateCondition(condition StepCondition, variables map[string]interface{}) bool {
	value, exists := variables[condition.Field]
	if !exists {
		return false
	}

	switch condition.Operator {
	case "eq":
		return value == condition.Value
	case "ne":
		return value != condition.Value
	case "gt":
		if v1, ok := value.(float64); ok {
			if v2, ok := condition.Value.(float64); ok {
				return v1 > v2
			}
		}
	case "lt":
		if v1, ok := value.(float64); ok {
			if v2, ok := condition.Value.(float64); ok {
				return v1 < v2
			}
		}
	case "contains":
		if v1, ok := value.(string); ok {
			if v2, ok := condition.Value.(string); ok {
				return fmt.Sprintf("%v", v1) == v2
			}
		}
	}

	return false
}

// 工作流状态管理方法

func (e *WorkflowEngine) completeWorkflow(instance *WorkflowInstance) {
	now := time.Now()
	instance.Status = "completed"
	instance.CompletedAt = &now
	e.db.Save(instance)

	log.Printf("Workflow completed: %s (ID: %s)", instance.Name, instance.ID)
}

func (e *WorkflowEngine) failWorkflow(instance *WorkflowInstance, errorMessage string) {
	now := time.Now()
	instance.Status = "failed"
	instance.CompletedAt = &now
	instance.OutputData = map[string]interface{}{
		"error": errorMessage,
	}
	e.db.Save(instance)

	log.Printf("Workflow failed: %s (ID: %s) - %s", instance.Name, instance.ID, errorMessage)
}

func (e *WorkflowEngine) failStep(stepInstance *StepInstance, errorMessage string) {
	now := time.Now()
	stepInstance.Status = "failed"
	stepInstance.CompletedAt = &now
	stepInstance.ErrorMessage = errorMessage
	e.db.Save(stepInstance)
}

// 辅助方法

func (e *WorkflowEngine) validateWorkflowDefinition(definition *WorkflowDefinition) error {
	if definition.Name == "" {
		return fmt.Errorf("workflow name is required")
	}

	if len(definition.Steps) == 0 {
		return fmt.Errorf("workflow must have at least one step")
	}

	// 验证步骤定义
	stepIDs := make(map[string]bool)
	for _, step := range definition.Steps {
		if step.ID == "" {
			return fmt.Errorf("step ID is required")
		}
		if stepIDs[step.ID] {
			return fmt.Errorf("duplicate step ID: %s", step.ID)
		}
		stepIDs[step.ID] = true
	}

	return nil
}

func (e *WorkflowEngine) sendApprovalNotification(approval *models.ApprovalProcess) {
	notification := &models.ApprovalNotification{
		ApprovalID:       approval.ID,
		UserID:           approval.ApproverID,
		NotificationType: "submit",
		Title:            fmt.Sprintf("New approval request: %s", approval.Name),
		Content:          approval.Description,
	}

	e.db.Create(notification)
}

func (e *WorkflowEngine) registerBuiltinActivities() {
	// 注册内置活动
	e.activityRunner.RegisterActivity("send_email", SendEmailActivity)
	e.activityRunner.RegisterActivity("send_notification", SendNotificationActivity)
	e.activityRunner.RegisterActivity("update_document", UpdateDocumentActivity)
	e.activityRunner.RegisterActivity("log_event", LogEventActivity)
}

// GetWorkflowInstance 获取工作流实例
func (e *WorkflowEngine) GetWorkflowInstance(instanceID string) (*WorkflowInstance, error) {
	var instance WorkflowInstance
	if err := e.db.First(&instance, "id = ?", instanceID).Error; err != nil {
		return nil, fmt.Errorf("workflow instance not found: %w", err)
	}
	return &instance, nil
}

// ListWorkflowInstances 获取工作流实例列表
func (e *WorkflowEngine) ListWorkflowInstances(teamID string, status string, page, limit int) ([]WorkflowInstance, error) {
	var instances []WorkflowInstance
	query := e.db.Where("team_id = ?", teamID)

	if status != "" {
		query = query.Where("status = ?", status)
	}

	offset := (page - 1) * limit
	if err := query.Offset(offset).Limit(limit).Order("created_at DESC").Find(&instances).Error; err != nil {
		return nil, err
	}

	return instances, nil
}

// CancelWorkflow 取消工作流
func (e *WorkflowEngine) CancelWorkflow(instanceID, userID string) error {
	instance, err := e.GetWorkflowInstance(instanceID)
	if err != nil {
		return err
	}

	if instance.Status != "running" {
		return fmt.Errorf("workflow is not running")
	}

	now := time.Now()
	instance.Status = "cancelled"
	instance.CompletedAt = &now

	if err := e.db.Save(instance).Error; err != nil {
		return fmt.Errorf("failed to cancel workflow: %w", err)
	}

	log.Printf("Workflow cancelled: %s (ID: %s) by user: %s", instance.Name, instance.ID, userID)
	return nil
}
