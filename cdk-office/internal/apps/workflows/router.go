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
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/linux-do/cdk-office/internal/db"
	"github.com/linux-do/cdk-office/internal/models"
	"gorm.io/gorm"
)

// Handler 工作流处理器
type Handler struct {
	db     *gorm.DB
	engine *WorkflowEngine
}

// NewHandler 创建工作流处理器
func NewHandler() *Handler {
	database := db.GetDB()
	engine := NewWorkflowEngine(database)

	return &Handler{
		db:     database,
		engine: engine,
	}
}

// RegisterRoutes 注册路由
func (h *Handler) RegisterRoutes(router *gin.RouterGroup) {
	workflow := router.Group("/workflows")
	{
		// 工作流定义管理
		workflow.POST("/definitions", h.CreateWorkflowDefinition)
		workflow.GET("/definitions", h.ListWorkflowDefinitions)
		workflow.GET("/definitions/:id", h.GetWorkflowDefinition)
		workflow.PUT("/definitions/:id", h.UpdateWorkflowDefinition)
		workflow.DELETE("/definitions/:id", h.DeleteWorkflowDefinition)

		// 工作流实例管理
		workflow.POST("/instances", h.StartWorkflowInstance)
		workflow.GET("/instances", h.ListWorkflowInstances)
		workflow.GET("/instances/:id", h.GetWorkflowInstance)
		workflow.POST("/instances/:id/cancel", h.CancelWorkflowInstance)
		workflow.POST("/instances/:id/pause", h.PauseWorkflowInstance)
		workflow.POST("/instances/:id/resume", h.ResumeWorkflowInstance)

		// 审批任务管理
		workflow.GET("/tasks", h.ListApprovalTasks)
		workflow.GET("/tasks/:id", h.GetApprovalTask)
		workflow.POST("/tasks/:id/approve", h.ApproveTask)
		workflow.POST("/tasks/:id/reject", h.RejectTask)

		// 活动管理
		workflow.GET("/activities", h.ListActivities)
		workflow.POST("/activities/test", h.TestActivity)

		// 统计信息
		workflow.GET("/statistics", h.GetStatistics)
		workflow.GET("/health", h.HealthCheck)
	}
}

// CreateWorkflowDefinition 创建工作流定义
// @Summary 创建工作流定义
// @Description 创建新的工作流定义
// @Tags workflows
// @Accept json
// @Produce json
// @Param workflow body WorkflowDefinition true "工作流定义"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]interface{}
// @Router /api/workflows/definitions [post]
func (h *Handler) CreateWorkflowDefinition(c *gin.Context) {
	var req WorkflowDefinition
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}

	// 设置创建者信息
	userID := c.GetString("user_id")
	teamID := c.GetString("team_id")

	req.CreatedBy = userID
	req.TeamID = teamID

	// 创建工作流定义
	if err := h.engine.CreateWorkflow(&req); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"id":      req.ID,
		"message": "Workflow definition created successfully",
	})
}

// ListWorkflowDefinitions 获取工作流定义列表
// @Summary 获取工作流定义列表
// @Description 获取团队的工作流定义列表
// @Tags workflows
// @Produce json
// @Param page query int false "页码" default(1)
// @Param limit query int false "每页数量" default(10)
// @Success 200 {object} map[string]interface{}
// @Router /api/workflows/definitions [get]
func (h *Handler) ListWorkflowDefinitions(c *gin.Context) {
	teamID := c.GetString("team_id")
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))

	var definitions []models.WorkflowDefinition
	var total int64

	query := h.db.Model(&models.WorkflowDefinition{}).Where("team_id = ?", teamID)

	// 获取总数
	query.Count(&total)

	// 分页查询
	offset := (page - 1) * limit
	if err := query.Offset(offset).Limit(limit).Order("created_at DESC").Find(&definitions).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch workflow definitions"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"definitions": definitions,
		"total":       total,
		"page":        page,
		"limit":       limit,
	})
}

// GetWorkflowDefinition 获取工作流定义详情
// @Summary 获取工作流定义详情
// @Description 根据ID获取工作流定义详情
// @Tags workflows
// @Produce json
// @Param id path string true "工作流定义ID"
// @Success 200 {object} map[string]interface{}
// @Failure 404 {object} map[string]interface{}
// @Router /api/workflows/definitions/{id} [get]
func (h *Handler) GetWorkflowDefinition(c *gin.Context) {
	id := c.Param("id")
	teamID := c.GetString("team_id")

	var definition models.WorkflowDefinition
	if err := h.db.Where("id = ? AND team_id = ?", id, teamID).First(&definition).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "Workflow definition not found"})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch workflow definition"})
		}
		return
	}

	// 解析工作流定义
	var workflowDef WorkflowDefinition
	if err := json.Unmarshal([]byte(definition.Definition), &workflowDef); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to parse workflow definition"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"definition": workflowDef,
		"metadata": map[string]interface{}{
			"id":         definition.ID,
			"created_at": definition.CreatedAt,
			"updated_at": definition.UpdatedAt,
			"created_by": definition.CreatedBy,
		},
	})
}

// UpdateWorkflowDefinition 更新工作流定义
func (h *Handler) UpdateWorkflowDefinition(c *gin.Context) {
	id := c.Param("id")
	teamID := c.GetString("team_id")
	userID := c.GetString("user_id")

	var req WorkflowDefinition
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}

	// 查找现有定义
	var definition models.WorkflowDefinition
	if err := h.db.Where("id = ? AND team_id = ?", id, teamID).First(&definition).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "Workflow definition not found"})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch workflow definition"})
		}
		return
	}

	// 更新定义
	req.ID = id
	req.TeamID = teamID
	req.CreatedBy = definition.CreatedBy

	definitionJSON, err := json.Marshal(req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to serialize workflow definition"})
		return
	}

	definition.Name = req.Name
	definition.Description = req.Description
	definition.Definition = string(definitionJSON)
	definition.UpdatedBy = userID

	if err := h.db.Save(&definition).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update workflow definition"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Workflow definition updated successfully"})
}

// DeleteWorkflowDefinition 删除工作流定义
func (h *Handler) DeleteWorkflowDefinition(c *gin.Context) {
	id := c.Param("id")
	teamID := c.GetString("team_id")

	// 检查是否有正在运行的实例
	var count int64
	h.db.Model(&WorkflowInstance{}).Where("workflow_def_id = ? AND status IN ?", id, []string{"running", "pending"}).Count(&count)
	if count > 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Cannot delete workflow definition with running instances"})
		return
	}

	// 删除定义
	if err := h.db.Where("id = ? AND team_id = ?", id, teamID).Delete(&models.WorkflowDefinition{}).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete workflow definition"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Workflow definition deleted successfully"})
}

// StartWorkflowInstance 启动工作流实例
// @Summary 启动工作流实例
// @Description 根据工作流定义启动新的实例
// @Tags workflows
// @Accept json
// @Produce json
// @Param request body map[string]interface{} true "启动参数"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]interface{}
// @Router /api/workflows/instances [post]
func (h *Handler) StartWorkflowInstance(c *gin.Context) {
	var req struct {
		WorkflowDefID string                 `json:"workflow_def_id" binding:"required"`
		InputData     map[string]interface{} `json:"input_data"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}

	userID := c.GetString("user_id")
	teamID := c.GetString("team_id")

	// 启动工作流实例
	instance, err := h.engine.StartWorkflow(req.WorkflowDefID, req.InputData, userID, teamID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"instance_id": instance.ID,
		"status":      instance.Status,
		"message":     "Workflow instance started successfully",
	})
}

// ListWorkflowInstances 获取工作流实例列表
func (h *Handler) ListWorkflowInstances(c *gin.Context) {
	teamID := c.GetString("team_id")
	status := c.Query("status")
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))

	instances, err := h.engine.ListWorkflowInstances(teamID, status, page, limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch workflow instances"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"instances": instances,
		"page":      page,
		"limit":     limit,
	})
}

// GetWorkflowInstance 获取工作流实例详情
func (h *Handler) GetWorkflowInstance(c *gin.Context) {
	id := c.Param("id")

	instance, err := h.engine.GetWorkflowInstance(id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Workflow instance not found"})
		return
	}

	// 获取步骤实例
	var steps []StepInstance
	h.db.Where("workflow_id = ?", id).Order("created_at ASC").Find(&steps)

	c.JSON(http.StatusOK, gin.H{
		"instance": instance,
		"steps":    steps,
	})
}

// CancelWorkflowInstance 取消工作流实例
func (h *Handler) CancelWorkflowInstance(c *gin.Context) {
	id := c.Param("id")
	userID := c.GetString("user_id")

	if err := h.engine.CancelWorkflow(id, userID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Workflow instance cancelled successfully"})
}

// PauseWorkflowInstance 暂停工作流实例
func (h *Handler) PauseWorkflowInstance(c *gin.Context) {
	id := c.Param("id")

	if err := h.engine.executor.PauseWorkflow(id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Workflow instance paused successfully"})
}

// ResumeWorkflowInstance 恢复工作流实例
func (h *Handler) ResumeWorkflowInstance(c *gin.Context) {
	id := c.Param("id")

	if err := h.engine.executor.ResumeWorkflow(id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Workflow instance resumed successfully"})
}

// ListApprovalTasks 获取审批任务列表
func (h *Handler) ListApprovalTasks(c *gin.Context) {
	userID := c.GetString("user_id")
	status := c.DefaultQuery("status", "pending")

	var approvals []models.ApprovalProcess
	query := h.db.Where("approver_id = ?", userID)

	if status != "" {
		query = query.Where("status = ?", status)
	}

	if err := query.Order("created_at DESC").Find(&approvals).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch approval tasks"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"approvals": approvals,
	})
}

// GetApprovalTask 获取审批任务详情
func (h *Handler) GetApprovalTask(c *gin.Context) {
	id := c.Param("id")
	userID := c.GetString("user_id")

	var approval models.ApprovalProcess
	if err := h.db.Where("id = ? AND approver_id = ?", id, userID).First(&approval).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "Approval task not found"})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch approval task"})
		}
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"approval": approval,
	})
}

// ApproveTask 批准任务
func (h *Handler) ApproveTask(c *gin.Context) {
	id := c.Param("id")

	var req struct {
		Comments string `json:"comments"`
	}
	c.ShouldBindJSON(&req)

	// 处理审批结果
	if err := h.engine.executor.HandleApprovalResult(id, true, req.Comments); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Task approved successfully"})
}

// RejectTask 拒绝任务
func (h *Handler) RejectTask(c *gin.Context) {
	id := c.Param("id")

	var req struct {
		Comments string `json:"comments" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Comments are required for rejection"})
		return
	}

	// 处理审批结果
	if err := h.engine.executor.HandleApprovalResult(id, false, req.Comments); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Task rejected successfully"})
}

// ListActivities 获取活动列表
func (h *Handler) ListActivities(c *gin.Context) {
	activities := h.engine.activityRunner.ListActivities()

	c.JSON(http.StatusOK, gin.H{
		"activities": activities,
	})
}

// TestActivity 测试活动
func (h *Handler) TestActivity(c *gin.Context) {
	var req struct {
		ActivityName string                 `json:"activity_name" binding:"required"`
		Config       map[string]interface{} `json:"config"`
		Variables    map[string]interface{} `json:"variables"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}

	result, err := h.engine.activityRunner.ExecuteActivity(req.ActivityName, req.Config, req.Variables)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"result": result,
	})
}

// GetStatistics 获取统计信息
func (h *Handler) GetStatistics(c *gin.Context) {
	teamID := c.GetString("team_id")

	// 工作流定义统计
	var defCount int64
	h.db.Model(&models.WorkflowDefinition{}).Where("team_id = ?", teamID).Count(&defCount)

	// 工作流实例统计
	var instanceStats map[string]int64 = make(map[string]int64)
	statuses := []string{"running", "completed", "failed", "cancelled"}

	for _, status := range statuses {
		var count int64
		h.db.Model(&WorkflowInstance{}).Where("team_id = ? AND status = ?", teamID, status).Count(&count)
		instanceStats[status] = count
	}

	// 执行器统计
	executorStats := h.engine.executor.GetWorkflowStatistics()

	c.JSON(http.StatusOK, gin.H{
		"definitions":    defCount,
		"instance_stats": instanceStats,
		"executor_stats": executorStats,
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

	// 检查执行器状态
	executorStats := h.engine.executor.GetWorkflowStatistics()

	c.JSON(http.StatusOK, gin.H{
		"status":         "healthy",
		"executor_stats": executorStats,
		"timestamp":      gin.H{"checked_at": gin.H{}},
	})
}
