/*
 * MIT License
 *
 * Copyright (c) 2025 CDK-Office
 */

package isolation

import (
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"github.com/linux-do/cdk-office/internal/models"
	"github.com/linux-do/cdk-office/internal/services/isolation"
)

// IsolationHandler 数据隔离管理处理器
type IsolationHandler struct {
	db               *gorm.DB
	isolationService *isolation.DataIsolationService
}

// NewIsolationHandler 创建数据隔离处理器
func NewIsolationHandler(db *gorm.DB, isolationService *isolation.DataIsolationService) *IsolationHandler {
	return &IsolationHandler{
		db:               db,
		isolationService: isolationService,
	}
}

// CreateTeamIsolationPolicy 创建团队隔离策略
func (h *IsolationHandler) CreateTeamIsolationPolicy(c *gin.Context) {
	type Request struct {
		TeamID              string   `json:"team_id" binding:"required"`
		StrictIsolation     bool     `json:"strict_isolation"`
		AllowCrossTeamView  bool     `json:"allow_cross_team_view"`
		AllowCrossTeamShare bool     `json:"allow_cross_team_share"`
		SystemPublicAccess  bool     `json:"system_public_access"`
		TeamPublicAccess    bool     `json:"team_public_access"`
		PrivateDataAccess   bool     `json:"private_data_access"`
		DownloadRestriction bool     `json:"download_restriction"`
		ShareRestriction    bool     `json:"share_restriction"`
		ExportRestriction   bool     `json:"export_restriction"`
		AllowedTeams        []string `json:"allowed_teams"`
		RestrictedActions   []string `json:"restricted_actions"`
		EnableAccessLog     bool     `json:"enable_access_log"`
		EnableOperationLog  bool     `json:"enable_operation_log"`
		LogRetentionDays    int      `json:"log_retention_days"`
		AlertOnViolation    bool     `json:"alert_on_violation"`
	}

	var req Request
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// 检查权限：只有超级管理员可以创建隔离策略
	userRole := c.GetString("role")
	if userRole != "super_admin" {
		c.JSON(http.StatusForbidden, gin.H{"error": "Only super administrators can create isolation policies"})
		return
	}

	// 检查策略是否已存在
	var existingPolicy models.TeamDataIsolationPolicy
	if err := h.db.Where("team_id = ?", req.TeamID).First(&existingPolicy).Error; err == nil {
		c.JSON(http.StatusConflict, gin.H{"error": "Isolation policy already exists for this team"})
		return
	}

	// 创建隔离策略
	policy := models.TeamDataIsolationPolicy{
		TeamID:              req.TeamID,
		StrictIsolation:     req.StrictIsolation,
		AllowCrossTeamView:  req.AllowCrossTeamView,
		AllowCrossTeamShare: req.AllowCrossTeamShare,
		CreatedBy:           c.GetString("user_id"),
	}

	if err := h.db.Create(&policy).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create isolation policy"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message": "Isolation policy created successfully",
		"policy":  policy,
	})
}

// GetTeamIsolationPolicy 获取团队隔离策略
func (h *IsolationHandler) GetTeamIsolationPolicy(c *gin.Context) {
	teamID := c.Param("team_id")
	userRole := c.GetString("role")
	userTeamID := c.GetString("team_id")

	// 权限检查：超级管理员可以查看所有，团队管理员只能查看本团队
	if userRole != "super_admin" && userTeamID != teamID {
		c.JSON(http.StatusForbidden, gin.H{"error": "Insufficient permissions"})
		return
	}

	var policy models.TeamDataIsolationPolicy
	if err := h.db.Where("team_id = ?", teamID).First(&policy).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "Isolation policy not found"})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get isolation policy"})
		}
		return
	}

	c.JSON(http.StatusOK, policy)
}

// UpdateTeamIsolationPolicy 更新团队隔离策略
func (h *IsolationHandler) UpdateTeamIsolationPolicy(c *gin.Context) {
	teamID := c.Param("team_id")
	userRole := c.GetString("role")

	// 只有超级管理员可以更新隔离策略
	if userRole != "super_admin" {
		c.JSON(http.StatusForbidden, gin.H{"error": "Only super administrators can update isolation policies"})
		return
	}

	type Request struct {
		StrictIsolation     *bool    `json:"strict_isolation,omitempty"`
		AllowCrossTeamView  *bool    `json:"allow_cross_team_view,omitempty"`
		AllowCrossTeamShare *bool    `json:"allow_cross_team_share,omitempty"`
		AllowedTeams        []string `json:"allowed_teams,omitempty"`
		RestrictedActions   []string `json:"restricted_actions,omitempty"`
	}

	var req Request
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var policy models.TeamDataIsolationPolicy
	if err := h.db.Where("team_id = ?", teamID).First(&policy).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "Isolation policy not found"})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get isolation policy"})
		}
		return
	}

	// 更新字段
	updates := make(map[string]interface{})

	if req.StrictIsolation != nil {
		updates["strict_isolation"] = *req.StrictIsolation
	}
	if req.AllowCrossTeamView != nil {
		updates["allow_cross_team_view"] = *req.AllowCrossTeamView
	}
	if req.AllowCrossTeamShare != nil {
		updates["allow_cross_team_share"] = *req.AllowCrossTeamShare
	}

	updates["updated_by"] = c.GetString("user_id")

	if err := h.db.Model(&policy).Updates(updates).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update isolation policy"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Isolation policy updated successfully"})
}

// CreateCrossTeamAccessRequest 创建跨团队访问申请
func (h *IsolationHandler) CreateCrossTeamAccessRequest(c *gin.Context) {
	type Request struct {
		TargetTeamID       string `json:"target_team_id" binding:"required"`
		TargetResourceID   string `json:"target_resource_id" binding:"required"`
		TargetResourceType string `json:"target_resource_type" binding:"required"`
		RequestType        string `json:"request_type" binding:"required"`
		RequestReason      string `json:"request_reason" binding:"required"`
		ExpectedDuration   int    `json:"expected_duration"`
	}

	var req Request
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	userID := c.GetString("user_id")
	userTeamID := c.GetString("team_id")

	// 检查是否是跨团队请求
	if userTeamID == req.TargetTeamID {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Cannot request access to your own team"})
		return
	}

	// 创建申请
	request := models.CrossTeamAccessRequest{
		RequesterID:        userID,
		RequesterTeamID:    userTeamID,
		TargetTeamID:       req.TargetTeamID,
		TargetResourceID:   req.TargetResourceID,
		TargetResourceType: req.TargetResourceType,
		RequestType:        req.RequestType,
		RequestReason:      req.RequestReason,
		ExpectedDuration:   req.ExpectedDuration,
		Status:             "pending",
	}

	if err := h.db.Create(&request).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create access request"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message": "Cross-team access request created successfully",
		"request": request,
	})
}

// GetDataAccessLogs 获取数据访问日志
func (h *IsolationHandler) GetDataAccessLogs(c *gin.Context) {
	userRole := c.GetString("role")
	userTeamID := c.GetString("team_id")
	userID := c.GetString("user_id")

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))

	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}

	var logs []models.DataAccessLog
	var total int64

	query := h.db.Model(&models.DataAccessLog{})

	// 根据用户角色过滤
	switch userRole {
	case "super_admin":
		// 超级管理员可以看所有日志
	case "team_manager":
		// 团队管理员可以看本团队的日志
		query = query.Where("team_id = ? OR owner_team_id = ?", userTeamID, userTeamID)
	default:
		// 普通用户只能看自己的日志
		query = query.Where("user_id = ?", userID)
	}

	// 计算总数
	if err := query.Count(&total).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to count logs"})
		return
	}

	// 分页查询
	offset := (page - 1) * pageSize
	if err := query.Order("created_at DESC").Offset(offset).Limit(pageSize).Find(&logs).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get logs"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"logs":        logs,
		"total":       total,
		"page":        page,
		"page_size":   pageSize,
		"total_pages": (total + int64(pageSize) - 1) / int64(pageSize),
	})
}

// GetViolations 获取违规记录
func (h *IsolationHandler) GetViolations(c *gin.Context) {
	userRole := c.GetString("role")
	userTeamID := c.GetString("team_id")

	// 只有管理员可以查看违规记录
	if userRole != "super_admin" && userRole != "team_manager" {
		c.JSON(http.StatusForbidden, gin.H{"error": "Insufficient permissions"})
		return
	}

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))

	var violations []models.DataIsolationViolation
	var total int64

	query := h.db.Model(&models.DataIsolationViolation{})

	// 团队管理员只能看本团队的违规
	if userRole == "team_manager" {
		query = query.Where("team_id = ?", userTeamID)
	}

	query.Count(&total)
	offset := (page - 1) * pageSize
	query.Order("created_at DESC").Offset(offset).Limit(pageSize).Find(&violations)

	c.JSON(http.StatusOK, gin.H{
		"violations": violations,
		"total":      total,
		"page":       page,
		"page_size":  pageSize,
	})
}

// GetCrossTeamAccessRequests 获取跨团队访问申请列表
func (h *IsolationHandler) GetCrossTeamAccessRequests(c *gin.Context) {
	userRole := c.GetString("role")
	userTeamID := c.GetString("team_id")
	userID := c.GetString("user_id")

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "10"))
	status := c.Query("status")

	var requests []models.CrossTeamAccessRequest
	var total int64

	query := h.db.Model(&models.CrossTeamAccessRequest{})

	// 根据用户角色过滤
	switch userRole {
	case "super_admin":
		// 超级管理员可以看所有申请
	case "team_manager":
		// 团队管理员可以看针对本团队的申请和本团队发出的申请
		query = query.Where("target_team_id = ? OR requester_team_id = ?", userTeamID, userTeamID)
	default:
		// 普通用户只能看自己的申请
		query = query.Where("requester_id = ?", userID)
	}

	if status != "" {
		query = query.Where("status = ?", status)
	}

	query.Count(&total)
	offset := (page - 1) * pageSize
	query.Order("created_at DESC").Offset(offset).Limit(pageSize).Find(&requests)

	c.JSON(http.StatusOK, gin.H{
		"requests":  requests,
		"total":     total,
		"page":      page,
		"page_size": pageSize,
	})
}

// ApproveCrossTeamAccessRequest 审批跨团队访问申请
func (h *IsolationHandler) ApproveCrossTeamAccessRequest(c *gin.Context) {
	requestID := c.Param("request_id")

	type Request struct {
		Action string `json:"action" binding:"required,oneof=approve reject"`
		Reason string `json:"reason"`
	}

	var req Request
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	userRole := c.GetString("role")
	userTeamID := c.GetString("team_id")
	userID := c.GetString("user_id")

	// 获取申请
	var request models.CrossTeamAccessRequest
	if err := h.db.Where("id = ?", requestID).First(&request).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "Access request not found"})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get access request"})
		}
		return
	}

	// 检查申请状态
	if request.Status != "pending" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Request is not pending"})
		return
	}

	// 权限检查
	if userRole != "super_admin" && request.TargetTeamID != userTeamID {
		c.JSON(http.StatusForbidden, gin.H{"error": "Insufficient permissions"})
		return
	}

	// 更新申请状态
	updates := map[string]interface{}{
		"status":          req.Action + "d", // approved 或 rejected
		"approver_id":     userID,
		"approval_reason": req.Reason,
		"approved_at":     time.Now(),
	}

	if req.Action == "approve" {
		// 设置过期时间
		expiresAt := time.Now().AddDate(0, 0, request.ExpectedDuration)
		updates["expires_at"] = expiresAt
	}

	if err := h.db.Model(&request).Updates(updates).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update request"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Request " + req.Action + "d successfully"})
}

// Placeholder methods for router compatibility
func (h *IsolationHandler) GetUserAccessProfile(c *gin.Context) {
	c.JSON(http.StatusNotImplemented, gin.H{"message": "Not implemented yet"})
}

func (h *IsolationHandler) UpdateUserAccessProfile(c *gin.Context) {
	c.JSON(http.StatusNotImplemented, gin.H{"message": "Not implemented yet"})
}

func (h *IsolationHandler) ListTeamIsolationPolicies(c *gin.Context) {
	c.JSON(http.StatusNotImplemented, gin.H{"message": "Not implemented yet"})
}

func (h *IsolationHandler) GetSystemVisibilityConfig(c *gin.Context) {
	c.JSON(http.StatusNotImplemented, gin.H{"message": "Not implemented yet"})
}

func (h *IsolationHandler) UpdateSystemVisibilityConfig(c *gin.Context) {
	c.JSON(http.StatusNotImplemented, gin.H{"message": "Not implemented yet"})
}

func (h *IsolationHandler) GetViolationsSummary(c *gin.Context) {
	c.JSON(http.StatusNotImplemented, gin.H{"message": "Not implemented yet"})
}

func (h *IsolationHandler) BlockUser(c *gin.Context) {
	c.JSON(http.StatusNotImplemented, gin.H{"message": "Not implemented yet"})
}

func (h *IsolationHandler) UnblockUser(c *gin.Context) {
	c.JSON(http.StatusNotImplemented, gin.H{"message": "Not implemented yet"})
}

func (h *IsolationHandler) GetDetailedAccessLogs(c *gin.Context) {
	c.JSON(http.StatusNotImplemented, gin.H{"message": "Not implemented yet"})
}

func (h *IsolationHandler) GetOperationLogs(c *gin.Context) {
	c.JSON(http.StatusNotImplemented, gin.H{"message": "Not implemented yet"})
}

func (h *IsolationHandler) GetCrossTeamActivities(c *gin.Context) {
	c.JSON(http.StatusNotImplemented, gin.H{"message": "Not implemented yet"})
}

func (h *IsolationHandler) GetUserActivities(c *gin.Context) {
	c.JSON(http.StatusNotImplemented, gin.H{"message": "Not implemented yet"})
}

func (h *IsolationHandler) ExportAccessLogs(c *gin.Context) {
	c.JSON(http.StatusNotImplemented, gin.H{"message": "Not implemented yet"})
}

func (h *IsolationHandler) ExportViolations(c *gin.Context) {
	c.JSON(http.StatusNotImplemented, gin.H{"message": "Not implemented yet"})
}

func (h *IsolationHandler) ExportTeamData(c *gin.Context) {
	c.JSON(http.StatusNotImplemented, gin.H{"message": "Not implemented yet"})
}
