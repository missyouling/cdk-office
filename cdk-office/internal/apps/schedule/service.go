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

package schedule

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/robfig/cron/v3"
	"gorm.io/gorm"

	"github.com/linux-do/cdk-office/internal/models"
)

// ScheduleService 调度服务
type ScheduleService struct {
	db        *gorm.DB
	cron      *cron.Cron
	tasks     map[string]*TaskContext
	mutex     sync.RWMutex
	stopChan  chan struct{}
	isRunning bool
}

// TaskContext 任务上下文
type TaskContext struct {
	Task      *models.ScheduledTask
	EntryID   cron.EntryID
	IsRunning bool
	LastRun   time.Time
	NextRun   time.Time
	Context   context.Context
	Cancel    context.CancelFunc
}

// TaskConfig 任务配置接口
type TaskConfig interface {
	Execute(ctx context.Context) error
	Validate() error
}

// WorkflowTaskConfig 工作流任务配置
type WorkflowTaskConfig struct {
	WorkflowDefID string                 `json:"workflow_def_id"`
	InputData     map[string]interface{} `json:"input_data"`
}

// ScriptTaskConfig 脚本任务配置
type ScriptTaskConfig struct {
	ScriptType string            `json:"script_type"` // shell, python, nodejs
	Script     string            `json:"script"`
	WorkingDir string            `json:"working_dir"`
	Env        map[string]string `json:"env"`
}

// HTTPRequestTaskConfig HTTP请求任务配置
type HTTPRequestTaskConfig struct {
	URL        string            `json:"url"`
	Method     string            `json:"method"`
	Headers    map[string]string `json:"headers"`
	Body       string            `json:"body"`
	Timeout    int               `json:"timeout"`
	RetryCount int               `json:"retry_count"`
}

// EmailTaskConfig 邮件任务配置
type EmailTaskConfig struct {
	To          []string `json:"to"`
	CC          []string `json:"cc"`
	BCC         []string `json:"bcc"`
	Subject     string   `json:"subject"`
	Body        string   `json:"body"`
	IsHTML      bool     `json:"is_html"`
	Attachments []string `json:"attachments"`
}

// NewScheduleService 创建调度服务
func NewScheduleService(db *gorm.DB) *ScheduleService {
	location, _ := time.LoadLocation("UTC")
	c := cron.New(cron.WithLocation(location), cron.WithSeconds())

	return &ScheduleService{
		db:       db,
		cron:     c,
		tasks:    make(map[string]*TaskContext),
		stopChan: make(chan struct{}),
	}
}

// Start 启动调度服务
func (s *ScheduleService) Start() error {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	if s.isRunning {
		return fmt.Errorf("schedule service is already running")
	}

	// 加载所有启用的任务
	if err := s.loadTasks(); err != nil {
		return fmt.Errorf("failed to load tasks: %w", err)
	}

	// 启动cron调度器
	s.cron.Start()
	s.isRunning = true

	// 启动监控协程
	go s.monitorTasks()

	log.Println("Schedule service started")
	return nil
}

// Stop 停止调度服务
func (s *ScheduleService) Stop() error {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	if !s.isRunning {
		return nil
	}

	// 停止cron调度器
	ctx := s.cron.Stop()
	<-ctx.Done()

	// 取消所有正在运行的任务
	for _, taskCtx := range s.tasks {
		if taskCtx.Cancel != nil {
			taskCtx.Cancel()
		}
	}

	// 停止监控协程
	close(s.stopChan)

	s.isRunning = false
	s.tasks = make(map[string]*TaskContext)

	log.Println("Schedule service stopped")
	return nil
}

// AddTask 添加任务
func (s *ScheduleService) AddTask(task *models.ScheduledTask) error {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	// 验证任务配置
	if err := s.validateTask(task); err != nil {
		return fmt.Errorf("invalid task configuration: %w", err)
	}

	// 保存任务到数据库
	if err := s.db.Create(task).Error; err != nil {
		return fmt.Errorf("failed to save task: %w", err)
	}

	// 如果任务启用，添加到调度器
	if task.IsEnabled {
		if err := s.scheduleTask(task); err != nil {
			return fmt.Errorf("failed to schedule task: %w", err)
		}
	}

	log.Printf("Added task: %s (ID: %s)", task.Name, task.ID)
	return nil
}

// RemoveTask 移除任务
func (s *ScheduleService) RemoveTask(taskID string) error {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	// 从调度器中移除
	if taskCtx, exists := s.tasks[taskID]; exists {
		s.cron.Remove(taskCtx.EntryID)
		if taskCtx.Cancel != nil {
			taskCtx.Cancel()
		}
		delete(s.tasks, taskID)
	}

	// 从数据库中删除
	if err := s.db.Delete(&models.ScheduledTask{}, "id = ?", taskID).Error; err != nil {
		return fmt.Errorf("failed to delete task: %w", err)
	}

	log.Printf("Removed task: %s", taskID)
	return nil
}

// UpdateTask 更新任务
func (s *ScheduleService) UpdateTask(task *models.ScheduledTask) error {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	// 验证任务配置
	if err := s.validateTask(task); err != nil {
		return fmt.Errorf("invalid task configuration: %w", err)
	}

	// 从调度器中移除旧任务
	if taskCtx, exists := s.tasks[task.ID]; exists {
		s.cron.Remove(taskCtx.EntryID)
		if taskCtx.Cancel != nil {
			taskCtx.Cancel()
		}
		delete(s.tasks, task.ID)
	}

	// 更新数据库
	if err := s.db.Save(task).Error; err != nil {
		return fmt.Errorf("failed to update task: %w", err)
	}

	// 如果任务启用，重新添加到调度器
	if task.IsEnabled {
		if err := s.scheduleTask(task); err != nil {
			return fmt.Errorf("failed to reschedule task: %w", err)
		}
	}

	log.Printf("Updated task: %s (ID: %s)", task.Name, task.ID)
	return nil
}

// EnableTask 启用任务
func (s *ScheduleService) EnableTask(taskID string) error {
	return s.toggleTask(taskID, true)
}

// DisableTask 禁用任务
func (s *ScheduleService) DisableTask(taskID string) error {
	return s.toggleTask(taskID, false)
}

// toggleTask 切换任务状态
func (s *ScheduleService) toggleTask(taskID string, enabled bool) error {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	// 获取任务
	var task models.ScheduledTask
	if err := s.db.First(&task, "id = ?", taskID).Error; err != nil {
		return fmt.Errorf("task not found: %w", err)
	}

	if task.IsEnabled == enabled {
		return nil // 状态没有变化
	}

	task.IsEnabled = enabled
	if err := s.db.Save(&task).Error; err != nil {
		return fmt.Errorf("failed to update task status: %w", err)
	}

	if enabled {
		// 启用任务 - 添加到调度器
		if err := s.scheduleTask(&task); err != nil {
			return fmt.Errorf("failed to schedule task: %w", err)
		}
	} else {
		// 禁用任务 - 从调度器中移除
		if taskCtx, exists := s.tasks[taskID]; exists {
			s.cron.Remove(taskCtx.EntryID)
			if taskCtx.Cancel != nil {
				taskCtx.Cancel()
			}
			delete(s.tasks, taskID)
		}
	}

	log.Printf("Task %s %s: %s", taskID, map[bool]string{true: "enabled", false: "disabled"}[enabled], task.Name)
	return nil
}

// ExecuteTaskNow 立即执行任务
func (s *ScheduleService) ExecuteTaskNow(taskID string) error {
	s.mutex.RLock()
	taskCtx, exists := s.tasks[taskID]
	s.mutex.RUnlock()

	if !exists {
		return fmt.Errorf("task not found: %s", taskID)
	}

	// 如果任务正在运行，返回错误
	if taskCtx.IsRunning {
		return fmt.Errorf("task is already running: %s", taskID)
	}

	// 在新的协程中执行任务
	go s.executeTask(taskCtx.Task)

	return nil
}

// GetTaskStatus 获取任务状态
func (s *ScheduleService) GetTaskStatus(taskID string) (*TaskStatus, error) {
	s.mutex.RLock()
	taskCtx, exists := s.tasks[taskID]
	s.mutex.RUnlock()

	if !exists {
		return nil, fmt.Errorf("task not found: %s", taskID)
	}

	status := &TaskStatus{
		TaskID:    taskID,
		IsRunning: taskCtx.IsRunning,
		LastRun:   taskCtx.LastRun,
		NextRun:   taskCtx.NextRun,
	}

	return status, nil
}

// TaskStatus 任务状态
type TaskStatus struct {
	TaskID    string    `json:"task_id"`
	IsRunning bool      `json:"is_running"`
	LastRun   time.Time `json:"last_run"`
	NextRun   time.Time `json:"next_run"`
}

// loadTasks 加载所有启用的任务
func (s *ScheduleService) loadTasks() error {
	var tasks []models.ScheduledTask
	if err := s.db.Where("is_enabled = ?", true).Find(&tasks).Error; err != nil {
		return err
	}

	for _, task := range tasks {
		if err := s.scheduleTask(&task); err != nil {
			log.Printf("Failed to schedule task %s: %v", task.ID, err)
			continue
		}
	}

	log.Printf("Loaded %d enabled tasks", len(tasks))
	return nil
}

// scheduleTask 调度任务
func (s *ScheduleService) scheduleTask(task *models.ScheduledTask) error {
	entryID, err := s.cron.AddFunc(task.CronExpr, func() {
		s.executeTask(task)
	})
	if err != nil {
		return err
	}

	ctx, cancel := context.WithCancel(context.Background())

	taskCtx := &TaskContext{
		Task:    task,
		EntryID: entryID,
		Context: ctx,
		Cancel:  cancel,
	}

	// 更新下次运行时间
	entries := s.cron.Entries()
	for _, entry := range entries {
		if entry.ID == entryID {
			taskCtx.NextRun = entry.Next
			break
		}
	}

	s.tasks[task.ID] = taskCtx
	return nil
}

// executeTask 执行任务
func (s *ScheduleService) executeTask(task *models.ScheduledTask) {
	s.mutex.Lock()
	taskCtx, exists := s.tasks[task.ID]
	if !exists {
		s.mutex.Unlock()
		return
	}

	if taskCtx.IsRunning {
		s.mutex.Unlock()
		log.Printf("Task %s is already running, skipping", task.ID)
		return
	}

	taskCtx.IsRunning = true
	taskCtx.LastRun = time.Now()
	s.mutex.Unlock()

	// 创建任务执行记录
	execution := &models.TaskExecution{
		TaskID:    task.ID,
		Status:    "running",
		StartTime: time.Now(),
	}
	s.db.Create(execution)

	// 执行任务
	err := s.doExecuteTask(task, execution)

	// 更新执行记录
	execution.EndTime = &execution.StartTime
	if execution.EndTime != nil {
		execution.Duration = int(execution.EndTime.Sub(execution.StartTime).Milliseconds())
	}

	if err != nil {
		execution.Status = "failed"
		execution.ErrorMessage = err.Error()
		log.Printf("Task %s failed: %v", task.ID, err)

		// 更新任务失败计数
		s.db.Model(task).UpdateColumn("failure_count", gorm.Expr("failure_count + 1"))
		s.db.Model(task).UpdateColumn("last_error", err.Error())
	} else {
		execution.Status = "completed"
		log.Printf("Task %s completed successfully", task.ID)

		// 更新任务成功计数
		s.db.Model(task).UpdateColumn("success_count", gorm.Expr("success_count + 1"))
	}

	// 更新任务运行计数和最后运行时间
	s.db.Model(task).UpdateColumn("run_count", gorm.Expr("run_count + 1"))
	s.db.Model(task).UpdateColumn("last_run_at", time.Now())

	s.db.Save(execution)

	// 重置运行状态
	s.mutex.Lock()
	if taskCtx, exists := s.tasks[task.ID]; exists {
		taskCtx.IsRunning = false
	}
	s.mutex.Unlock()
}

// doExecuteTask 实际执行任务
func (s *ScheduleService) doExecuteTask(task *models.ScheduledTask, execution *models.TaskExecution) error {
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(task.Timeout)*time.Second)
	defer cancel()

	switch task.TaskType {
	case "workflow":
		return s.executeWorkflowTask(ctx, task, execution)
	case "script":
		return s.executeScriptTask(ctx, task, execution)
	case "http_request":
		return s.executeHTTPRequestTask(ctx, task, execution)
	case "email":
		return s.executeEmailTask(ctx, task, execution)
	default:
		return fmt.Errorf("unsupported task type: %s", task.TaskType)
	}
}

// executeWorkflowTask 执行工作流任务
func (s *ScheduleService) executeWorkflowTask(ctx context.Context, task *models.ScheduledTask, execution *models.TaskExecution) error {
	var config WorkflowTaskConfig
	if err := json.Unmarshal([]byte(task.TaskConfig), &config); err != nil {
		return fmt.Errorf("invalid workflow task config: %w", err)
	}

	// 这里应该调用工作流引擎
	log.Printf("Executing workflow task: %s, workflow: %s", task.ID, config.WorkflowDefID)

	// 模拟工作流执行
	execution.ExecutionID = uuid.New().String()
	execution.Output = "Workflow execution completed"

	return nil
}

// executeScriptTask 执行脚本任务
func (s *ScheduleService) executeScriptTask(ctx context.Context, task *models.ScheduledTask, execution *models.TaskExecution) error {
	var config ScriptTaskConfig
	if err := json.Unmarshal([]byte(task.TaskConfig), &config); err != nil {
		return fmt.Errorf("invalid script task config: %w", err)
	}

	log.Printf("Executing script task: %s, type: %s", task.ID, config.ScriptType)

	// 这里应该实现真实的脚本执行逻辑
	// 例如：使用os/exec包执行脚本
	execution.Output = "Script execution completed"

	return nil
}

// executeHTTPRequestTask 执行HTTP请求任务
func (s *ScheduleService) executeHTTPRequestTask(ctx context.Context, task *models.ScheduledTask, execution *models.TaskExecution) error {
	var config HTTPRequestTaskConfig
	if err := json.Unmarshal([]byte(task.TaskConfig), &config); err != nil {
		return fmt.Errorf("invalid HTTP request task config: %w", err)
	}

	log.Printf("Executing HTTP request task: %s, URL: %s", task.ID, config.URL)

	// 这里应该实现真实的HTTP请求逻辑
	execution.Output = "HTTP request completed"

	return nil
}

// executeEmailTask 执行邮件任务
func (s *ScheduleService) executeEmailTask(ctx context.Context, task *models.ScheduledTask, execution *models.TaskExecution) error {
	var config EmailTaskConfig
	if err := json.Unmarshal([]byte(task.TaskConfig), &config); err != nil {
		return fmt.Errorf("invalid email task config: %w", err)
	}

	log.Printf("Executing email task: %s, to: %v", task.ID, config.To)

	// 这里应该实现真实的邮件发送逻辑
	execution.Output = "Email sent successfully"

	return nil
}

// validateTask 验证任务配置
func (s *ScheduleService) validateTask(task *models.ScheduledTask) error {
	if task.Name == "" {
		return fmt.Errorf("task name is required")
	}

	if task.CronExpr == "" {
		return fmt.Errorf("cron expression is required")
	}

	// 验证cron表达式
	if _, err := cron.ParseStandard(task.CronExpr); err != nil {
		return fmt.Errorf("invalid cron expression: %w", err)
	}

	// 验证任务配置
	switch task.TaskType {
	case "workflow":
		var config WorkflowTaskConfig
		if err := json.Unmarshal([]byte(task.TaskConfig), &config); err != nil {
			return fmt.Errorf("invalid workflow task config: %w", err)
		}
		if config.WorkflowDefID == "" {
			return fmt.Errorf("workflow definition ID is required")
		}
	case "script":
		var config ScriptTaskConfig
		if err := json.Unmarshal([]byte(task.TaskConfig), &config); err != nil {
			return fmt.Errorf("invalid script task config: %w", err)
		}
		if config.Script == "" {
			return fmt.Errorf("script content is required")
		}
	case "http_request":
		var config HTTPRequestTaskConfig
		if err := json.Unmarshal([]byte(task.TaskConfig), &config); err != nil {
			return fmt.Errorf("invalid HTTP request task config: %w", err)
		}
		if config.URL == "" {
			return fmt.Errorf("URL is required")
		}
	case "email":
		var config EmailTaskConfig
		if err := json.Unmarshal([]byte(task.TaskConfig), &config); err != nil {
			return fmt.Errorf("invalid email task config: %w", err)
		}
		if len(config.To) == 0 {
			return fmt.Errorf("email recipients are required")
		}
	default:
		return fmt.Errorf("unsupported task type: %s", task.TaskType)
	}

	return nil
}

// monitorTasks 监控任务状态
func (s *ScheduleService) monitorTasks() {
	ticker := time.NewTicker(1 * time.Minute)
	defer ticker.Stop()

	for {
		select {
		case <-s.stopChan:
			return
		case <-ticker.C:
			s.updateTaskSchedules()
		}
	}
}

// updateTaskSchedules 更新任务调度时间
func (s *ScheduleService) updateTaskSchedules() {
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	entries := s.cron.Entries()
	for _, taskCtx := range s.tasks {
		for _, entry := range entries {
			if entry.ID == taskCtx.EntryID {
				taskCtx.NextRun = entry.Next

				// 更新数据库中的下次运行时间
				s.db.Model(&models.ScheduledTask{}).Where("id = ?", taskCtx.Task.ID).
					UpdateColumn("next_run_at", entry.Next)
				break
			}
		}
	}
}
