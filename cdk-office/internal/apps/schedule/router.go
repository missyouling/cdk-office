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
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/linux-do/cdk-office/internal/db"
	"github.com/linux-do/cdk-office/internal/models"
	"gorm.io/gorm"
)

// Handler 调度处理器
type Handler struct {
	db      *gorm.DB
	service *ScheduleService
}

// NewHandler 创建调度处理器
func NewHandler() *Handler {
	database := db.GetDB()
	service := NewScheduleService(database)

	return &Handler{
		db:      database,
		service: service,
	}
}

// StartService 启动调度服务
func (h *Handler) StartService() error {
	return h.service.Start()
}

// StopService 停止调度服务
func (h *Handler) StopService() error {
	return h.service.Stop()
}

// RegisterRoutes 注册路由
func (h *Handler) RegisterRoutes(router *gin.RouterGroup) {
	schedule := router.Group("/schedule")
	{
		// 任务管理
		schedule.POST("/tasks", h.CreateTask)
		schedule.GET("/tasks", h.ListTasks)
		schedule.GET("/tasks/:id", h.GetTask)
		schedule.PUT("/tasks/:id", h.UpdateTask)
		schedule.DELETE("/tasks/:id", h.DeleteTask)

		// 任务控制
		schedule.POST("/tasks/:id/enable", h.EnableTask)
		schedule.POST("/tasks/:id/disable", h.DisableTask)
		schedule.POST("/tasks/:id/execute", h.ExecuteTask)
		schedule.GET("/tasks/:id/status", h.GetTaskStatus)

		// 执行记录
		schedule.GET("/tasks/:id/executions", h.ListTaskExecutions)
		schedule.GET("/executions/:id", h.GetExecution)

		// 任务模板
		schedule.GET("/templates", h.ListTaskTemplates)
		schedule.POST("/templates", h.CreateTaskTemplate)
		schedule.GET("/templates/:id", h.GetTaskTemplate)

		// 统计信息
		schedule.GET("/statistics", h.GetStatistics)
		schedule.GET("/health", h.HealthCheck)

		// 任务依赖
		schedule.POST("/tasks/:id/dependencies", h.AddTaskDependency)
		schedule.GET("/tasks/:id/dependencies", h.ListTaskDependencies)
		schedule.DELETE("/dependencies/:id", h.RemoveTaskDependency)

		// 通知配置
		schedule.POST("/tasks/:id/notifications", h.AddTaskNotification)
		schedule.GET("/tasks/:id/notifications", h.ListTaskNotifications)
		schedule.PUT("/notifications/:id", h.UpdateTaskNotification)
		schedule.DELETE("/notifications/:id", h.DeleteTaskNotification)
	}
}

// CreateTask 创建任务
// @Summary 创建调度任务
// @Description 创建新的调度任务
// @Tags schedule
// @Accept json
// @Produce json
// @Param task body models.ScheduledTask true "调度任务"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]interface{}
// @Router /api/schedule/tasks [post]
func (h *Handler) CreateTask(c *gin.Context) {
	var req models.ScheduledTask
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}

	// 设置创建者信息
	userID := c.GetString("user_id")
	teamID := c.GetString("team_id")

	req.CreatedBy = userID
	req.TeamID = teamID

	// 添加任务
	if err := h.service.AddTask(&req); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"id":      req.ID,
		"message": "Task created successfully",
	})
}

// ListTasks 获取任务列表
// @Summary 获取调度任务列表
// @Description 获取团队的调度任务列表
// @Tags schedule
// @Produce json
// @Param page query int false "页码" default(1)
// @Param limit query int false "每页数量" default(10)
// @Param enabled query bool false "是否启用"
// @Param task_type query string false "任务类型"
// @Success 200 {object} map[string]interface{}
// @Router /api/schedule/tasks [get]
func (h *Handler) ListTasks(c *gin.Context) {
	teamID := c.GetString("team_id")
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))
	enabled := c.Query("enabled")
	taskType := c.Query("task_type")

	var tasks []models.ScheduledTask
	var total int64

	query := h.db.Model(&models.ScheduledTask{}).Where("team_id = ?", teamID)

	if enabled != "" {
		if enabledBool, err := strconv.ParseBool(enabled); err == nil {
			query = query.Where("is_enabled = ?", enabledBool)
		}
	}

	if taskType != "" {
		query = query.Where("task_type = ?", taskType)
	}

	// 获取总数
	query.Count(&total)

	// 分页查询
	offset := (page - 1) * limit
	if err := query.Offset(offset).Limit(limit).Order("created_at DESC").Find(&tasks).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch tasks"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"tasks": tasks,
		"total": total,
		"page":  page,
		"limit": limit,
	})
}

// GetTask 获取任务详情
// @Summary 获取调度任务详情
// @Description 根据ID获取调度任务详情
// @Tags schedule
// @Produce json
// @Param id path string true "任务ID"
// @Success 200 {object} map[string]interface{}
// @Failure 404 {object} map[string]interface{}
// @Router /api/schedule/tasks/{id} [get]
func (h *Handler) GetTask(c *gin.Context) {
	id := c.Param("id")
	teamID := c.GetString("team_id")

	var task models.ScheduledTask
	if err := h.db.Where("id = ? AND team_id = ?", id, teamID).First(&task).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "Task not found"})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch task"})
		}
		return
	}

	// 获取任务状态
	status, _ := h.service.GetTaskStatus(id)

	c.JSON(http.StatusOK, gin.H{
		"task":   task,
		"status": status,
	})
}

// UpdateTask 更新任务
func (h *Handler) UpdateTask(c *gin.Context) {
	id := c.Param("id")
	teamID := c.GetString("team_id")
	userID := c.GetString("user_id")

	var req models.ScheduledTask
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}

	// 查找现有任务
	var task models.ScheduledTask
	if err := h.db.Where("id = ? AND team_id = ?", id, teamID).First(&task).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "Task not found"})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch task"})
		}
		return
	}

	// 更新任务
	req.ID = id
	req.TeamID = teamID
	req.CreatedBy = task.CreatedBy
	req.UpdatedBy = userID

	if err := h.service.UpdateTask(&req); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Task updated successfully"})
}

// DeleteTask 删除任务
func (h *Handler) DeleteTask(c *gin.Context) {
	id := c.Param("id")
	teamID := c.GetString("team_id")

	// 验证任务存在
	var task models.ScheduledTask
	if err := h.db.Where("id = ? AND team_id = ?", id, teamID).First(&task).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "Task not found"})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch task"})
		}
		return
	}

	// 删除任务
	if err := h.service.RemoveTask(id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Task deleted successfully"})
}

// EnableTask 启用任务
func (h *Handler) EnableTask(c *gin.Context) {
	id := c.Param("id")

	if err := h.service.EnableTask(id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Task enabled successfully"})
}

// DisableTask 禁用任务
func (h *Handler) DisableTask(c *gin.Context) {
	id := c.Param("id")

	if err := h.service.DisableTask(id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Task disabled successfully"})
}

// ExecuteTask 立即执行任务
func (h *Handler) ExecuteTask(c *gin.Context) {
	id := c.Param("id")

	if err := h.service.ExecuteTaskNow(id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Task execution started"})
}

// GetTaskStatus 获取任务状态
func (h *Handler) GetTaskStatus(c *gin.Context) {
	id := c.Param("id")

	status, err := h.service.GetTaskStatus(id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"status": status})
}

// ListTaskExecutions 获取任务执行记录
func (h *Handler) ListTaskExecutions(c *gin.Context) {
	taskID := c.Param("id")
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))
	status := c.Query("status")

	var executions []models.TaskExecution
	var total int64

	query := h.db.Model(&models.TaskExecution{}).Where("task_id = ?", taskID)

	if status != "" {
		query = query.Where("status = ?", status)
	}

	// 获取总数
	query.Count(&total)

	// 分页查询
	offset := (page - 1) * limit
	if err := query.Offset(offset).Limit(limit).Order("start_time DESC").Find(&executions).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch executions"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"executions": executions,
		"total":      total,
		"page":       page,
		"limit":      limit,
	})
}

// GetExecution 获取执行记录详情
func (h *Handler) GetExecution(c *gin.Context) {
	id := c.Param("id")

	var execution models.TaskExecution
	if err := h.db.First(&execution, "id = ?", id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "Execution not found"})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch execution"})
		}
		return
	}

	c.JSON(http.StatusOK, gin.H{"execution": execution})
}

// ListTaskTemplates 获取任务模板列表
func (h *Handler) ListTaskTemplates(c *gin.Context) {
	teamID := c.GetString("team_id")
	category := c.Query("category")
	taskType := c.Query("task_type")

	var templates []models.TaskTemplate
	query := h.db.Where("team_id = ? OR is_public = true", teamID)

	if category != "" {
		query = query.Where("category = ?", category)
	}

	if taskType != "" {
		query = query.Where("task_type = ?", taskType)
	}

	if err := query.Order("use_count DESC, created_at DESC").Find(&templates).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch templates"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"templates": templates})
}

// CreateTaskTemplate 创建任务模板
func (h *Handler) CreateTaskTemplate(c *gin.Context) {
	var req models.TaskTemplate
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}

	userID := c.GetString("user_id")
	teamID := c.GetString("team_id")

	req.CreatedBy = userID
	req.TeamID = teamID

	if err := h.db.Create(&req).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create template"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"id":      req.ID,
		"message": "Template created successfully",
	})
}

// GetTaskTemplate 获取任务模板详情
func (h *Handler) GetTaskTemplate(c *gin.Context) {
	id := c.Param("id")
	teamID := c.GetString("team_id")

	var template models.TaskTemplate
	if err := h.db.Where("id = ? AND (team_id = ? OR is_public = true)", id, teamID).First(&template).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "Template not found"})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch template"})
		}
		return
	}

	c.JSON(http.StatusOK, gin.H{"template": template})
}

// AddTaskDependency 添加任务依赖
func (h *Handler) AddTaskDependency(c *gin.Context) {
	taskID := c.Param("id")

	var req models.TaskDependency
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}

	req.TaskID = taskID

	if err := h.db.Create(&req).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to add dependency"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Dependency added successfully"})
}

// ListTaskDependencies 获取任务依赖列表
func (h *Handler) ListTaskDependencies(c *gin.Context) {
	taskID := c.Param("id")

	var dependencies []models.TaskDependency
	if err := h.db.Where("task_id = ?", taskID).Find(&dependencies).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch dependencies"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"dependencies": dependencies})
}

// RemoveTaskDependency 移除任务依赖
func (h *Handler) RemoveTaskDependency(c *gin.Context) {
	id := c.Param("id")

	if err := h.db.Delete(&models.TaskDependency{}, "id = ?", id).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to remove dependency"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Dependency removed successfully"})
}

// AddTaskNotification 添加任务通知
func (h *Handler) AddTaskNotification(c *gin.Context) {
	taskID := c.Param("id")
	userID := c.GetString("user_id")

	var req models.TaskNotification
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}

	req.TaskID = taskID
	req.CreatedBy = userID

	if err := h.db.Create(&req).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to add notification"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Notification added successfully"})
}

// ListTaskNotifications 获取任务通知列表
func (h *Handler) ListTaskNotifications(c *gin.Context) {
	taskID := c.Param("id")

	var notifications []models.TaskNotification
	if err := h.db.Where("task_id = ?", taskID).Find(&notifications).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch notifications"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"notifications": notifications})
}

// UpdateTaskNotification 更新任务通知
func (h *Handler) UpdateTaskNotification(c *gin.Context) {
	id := c.Param("id")

	var req models.TaskNotification
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}

	if err := h.db.Model(&models.TaskNotification{}).Where("id = ?", id).Updates(&req).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update notification"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Notification updated successfully"})
}

// DeleteTaskNotification 删除任务通知
func (h *Handler) DeleteTaskNotification(c *gin.Context) {
	id := c.Param("id")

	if err := h.db.Delete(&models.TaskNotification{}, "id = ?", id).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete notification"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Notification deleted successfully"})
}

// GetStatistics 获取统计信息
func (h *Handler) GetStatistics(c *gin.Context) {
	teamID := c.GetString("team_id")

	// 任务统计
	var taskStats map[string]int64 = make(map[string]int64)
	var totalTasks int64
	h.db.Model(&models.ScheduledTask{}).Where("team_id = ?", teamID).Count(&totalTasks)
	taskStats["total"] = totalTasks

	var enabledTasks int64
	h.db.Model(&models.ScheduledTask{}).Where("team_id = ? AND is_enabled = true", teamID).Count(&enabledTasks)
	taskStats["enabled"] = enabledTasks

	// 执行统计
	var executionStats map[string]int64 = make(map[string]int64)
	statuses := []string{"running", "completed", "failed"}

	for _, status := range statuses {
		var count int64
		h.db.Model(&models.TaskExecution{}).
			Joins("JOIN scheduled_tasks ON task_executions.task_id = scheduled_tasks.id").
			Where("scheduled_tasks.team_id = ? AND task_executions.status = ?", teamID, status).
			Count(&count)
		executionStats[status] = count
	}

	// 今日执行统计
	today := time.Now().Truncate(24 * time.Hour)
	var todayExecutions int64
	h.db.Model(&models.TaskExecution{}).
		Joins("JOIN scheduled_tasks ON task_executions.task_id = scheduled_tasks.id").
		Where("scheduled_tasks.team_id = ? AND task_executions.start_time >= ?", teamID, today).
		Count(&todayExecutions)

	c.JSON(http.StatusOK, gin.H{
		"task_stats":       taskStats,
		"execution_stats":  executionStats,
		"today_executions": todayExecutions,
	})
}

// HealthCheck 健康检查
func (h *Handler) HealthCheck(c *gin.Context) {
	// 检查数据库连接
	sqlDB, err := h.db.DB()
	if err != nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"status": "unhealthy", "error": "database connection failed"})
		return
	}

	if err := sqlDB.Ping(); err != nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"status": "unhealthy", "error": "database ping failed"})
		return
	}

	// 检查调度服务状态
	isRunning := h.service.isRunning

	c.JSON(http.StatusOK, gin.H{
		"status":            "healthy",
		"scheduler_running": isRunning,
		"timestamp":         time.Now(),
	})
}
