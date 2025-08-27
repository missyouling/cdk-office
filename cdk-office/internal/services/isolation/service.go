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

package isolation

import (
	"context"
	"errors"
	"fmt"
	"time"

	"gorm.io/gorm"

	"github.com/linux-do/cdk-office/internal/models"
)

// DataIsolationService 数据隔离服务
type DataIsolationService struct {
	db    *gorm.DB
	cache CacheInterface
}

// CacheInterface 缓存接口
type CacheInterface interface {
	Get(key string) (interface{}, error)
	Set(key string, value interface{}, expiration time.Duration) error
	Delete(key string) error
}

// AccessContext 访问上下文
type AccessContext struct {
	UserID        string            `json:"user_id"`
	TeamID        string            `json:"team_id"`
	UserRole      string            `json:"user_role"`
	ResourceID    string            `json:"resource_id"`
	ResourceType  string            `json:"resource_type"`
	ActionType    string            `json:"action_type"`
	OwnerTeamID   string            `json:"owner_team_id"`
	RequestSource string            `json:"request_source"`
	Metadata      map[string]string `json:"metadata"`
}

// AccessResult 访问结果
type AccessResult struct {
	Allowed          bool                   `json:"allowed"`
	Reason           string                 `json:"reason"`
	Restrictions     map[string]interface{} `json:"restrictions"`
	RequiresApproval bool                   `json:"requires_approval"`
	ExpiresAt        *time.Time             `json:"expires_at"`
}

// NewDataIsolationService 创建数据隔离服务
func NewDataIsolationService(db *gorm.DB, cache CacheInterface) *DataIsolationService {
	return &DataIsolationService{
		db:    db,
		cache: cache,
	}
}

// CheckAccess 检查数据访问权限
func (s *DataIsolationService) CheckAccess(ctx context.Context, accessCtx *AccessContext) (*AccessResult, error) {
	// 记录访问尝试
	defer s.logAccess(accessCtx, time.Now())

	// 1. 获取用户数据访问档案
	profile, err := s.getUserAccessProfile(accessCtx.UserID)
	if err != nil {
		return &AccessResult{Allowed: false, Reason: "Failed to get user profile"}, err
	}

	// 2. 检查用户是否被封禁
	if profile.SecuritySettings.IsBlocked {
		if profile.SecuritySettings.BlockExpires != nil && time.Now().After(*profile.SecuritySettings.BlockExpires) {
			// 解除封禁
			profile.SecuritySettings.IsBlocked = false
			profile.SecuritySettings.BlockExpires = nil
			s.db.Save(profile)
		} else {
			s.recordViolation(accessCtx, "access_denied", "User is blocked", "high")
			return &AccessResult{Allowed: false, Reason: "User is currently blocked"}, nil
		}
	}

	// 3. 检查每日访问限制
	if err := s.checkDailyLimits(profile, accessCtx); err != nil {
		s.recordViolation(accessCtx, "rate_limit_exceeded", err.Error(), "medium")
		return &AccessResult{Allowed: false, Reason: err.Error()}, nil
	}

	// 4. 获取团队隔离策略
	policy, err := s.getTeamIsolationPolicy(accessCtx.OwnerTeamID)
	if err != nil {
		return &AccessResult{Allowed: false, Reason: "Failed to get team policy"}, err
	}

	// 5. 执行访问权限检查
	result := s.evaluateAccess(accessCtx, policy, profile)

	// 6. 更新访问统计
	s.updateAccessStats(profile, accessCtx, result.Allowed)

	return result, nil
}

// evaluateAccess 评估访问权限
func (s *DataIsolationService) evaluateAccess(accessCtx *AccessContext, policy *models.TeamDataIsolationPolicy, profile *models.UserDataAccessProfile) *AccessResult {
	// 同团队访问
	if accessCtx.TeamID == accessCtx.OwnerTeamID {
		return &AccessResult{
			Allowed: true,
			Reason:  "Same team access",
		}
	}

	// 超级管理员权限
	if accessCtx.UserRole == "super_admin" {
		return &AccessResult{
			Allowed: true,
			Reason:  "Super admin access",
		}
	}

	// 检查严格隔离模式
	if policy.StrictIsolation {
		// 严格模式下只允许特定情况的跨团队访问
		return s.evaluateStrictModeAccess(accessCtx, policy, profile)
	}

	// 宽松模式下的权限检查
	return s.evaluateNormalModeAccess(accessCtx, policy, profile)
}

// evaluateStrictModeAccess 严格模式访问评估
func (s *DataIsolationService) evaluateStrictModeAccess(accessCtx *AccessContext, policy *models.TeamDataIsolationPolicy, profile *models.UserDataAccessProfile) *AccessResult {
	// 检查是否在允许的团队列表中
	for _, allowedTeam := range policy.AccessRestrictions.AllowedTeams {
		if allowedTeam == accessCtx.TeamID {
			return &AccessResult{
				Allowed: true,
				Reason:  "Team in allowed list",
			}
		}
	}

	// 检查资源可见性
	visibility := s.getResourceVisibility(accessCtx.ResourceID, accessCtx.ResourceType)

	switch visibility {
	case "system_public":
		if policy.VisibilitySettings.SystemPublicAccess {
			return &AccessResult{
				Allowed: true,
				Reason:  "System public resource",
			}
		}
	case "team_public":
		if policy.VisibilitySettings.TeamPublicAccess {
			// 团队公开资源允许查看，但可能限制其他操作
			restrictions := make(map[string]interface{})
			if accessCtx.ActionType == "download" && policy.AccessRestrictions.DownloadRestriction {
				restrictions["download"] = false
			}
			if accessCtx.ActionType == "share" && policy.AccessRestrictions.ShareRestriction {
				restrictions["share"] = false
			}

			return &AccessResult{
				Allowed:      true,
				Reason:       "Team public resource with restrictions",
				Restrictions: restrictions,
			}
		}
	}

	// 检查是否需要申请审批
	if profile.SecuritySettings.RequireApproval {
		// 检查是否有有效的跨团队访问申请
		hasValidRequest := s.hasValidCrossTeamRequest(accessCtx)
		if hasValidRequest {
			return &AccessResult{
				Allowed: true,
				Reason:  "Valid cross-team access request",
			}
		}

		return &AccessResult{
			Allowed:          false,
			Reason:           "Cross-team access requires approval",
			RequiresApproval: true,
		}
	}

	// 默认拒绝访问
	s.recordViolation(accessCtx, "access_denied", "Strict isolation policy violation", "medium")
	return &AccessResult{
		Allowed: false,
		Reason:  "Access denied by strict isolation policy",
	}
}

// evaluateNormalModeAccess 普通模式访问评估
func (s *DataIsolationService) evaluateNormalModeAccess(accessCtx *AccessContext, policy *models.TeamDataIsolationPolicy, profile *models.UserDataAccessProfile) *AccessResult {
	// 检查跨团队查看权限
	if accessCtx.ActionType == "view" && policy.AllowCrossTeamView {
		return &AccessResult{
			Allowed: true,
			Reason:  "Cross-team view allowed",
		}
	}

	// 检查跨团队分享权限
	if accessCtx.ActionType == "share" && policy.AllowCrossTeamShare {
		// 团队管理员及以上才能跨团队分享
		if accessCtx.UserRole == "team_manager" || accessCtx.UserRole == "super_admin" {
			return &AccessResult{
				Allowed: true,
				Reason:  "Cross-team share allowed for managers",
			}
		}
	}

	// 按角色检查权限
	return s.evaluateRoleBasedAccess(accessCtx, policy, profile)
}

// evaluateRoleBasedAccess 基于角色的访问评估
func (s *DataIsolationService) evaluateRoleBasedAccess(accessCtx *AccessContext, policy *models.TeamDataIsolationPolicy, profile *models.UserDataAccessProfile) *AccessResult {
	visibility := s.getResourceVisibility(accessCtx.ResourceID, accessCtx.ResourceType)

	switch accessCtx.UserRole {
	case "team_manager":
		// 团队管理员可以访问团队公开和系统公开资源
		if visibility == "system_public" || visibility == "team_public" {
			return &AccessResult{
				Allowed: true,
				Reason:  "Team manager access to public resources",
			}
		}

	case "collaborator":
		// 协作用户可以访问系统公开资源
		if visibility == "system_public" {
			return &AccessResult{
				Allowed: true,
				Reason:  "Collaborator access to system public resources",
			}
		}

		// 如果在用户的允许团队列表中
		for _, allowedTeam := range profile.AllowedTeams {
			if allowedTeam == accessCtx.OwnerTeamID {
				return &AccessResult{
					Allowed: true,
					Reason:  "Collaborator access to allowed team resources",
				}
			}
		}

	case "normal_user":
		// 普通用户只能访问系统公开资源
		if visibility == "system_public" {
			return &AccessResult{
				Allowed: true,
				Reason:  "Normal user access to system public resources",
			}
		}
	}

	// 默认拒绝
	s.recordViolation(accessCtx, "access_denied", "Insufficient role permissions", "low")
	return &AccessResult{
		Allowed: false,
		Reason:  "Insufficient permissions for cross-team access",
	}
}

// getResourceVisibility 获取资源可见性
func (s *DataIsolationService) getResourceVisibility(resourceID, resourceType string) string {
	// 从缓存获取
	cacheKey := fmt.Sprintf("resource_visibility:%s:%s", resourceType, resourceID)
	if cached, err := s.cache.Get(cacheKey); err == nil {
		if visibility, ok := cached.(string); ok {
			return visibility
		}
	}

	// 从数据库查询
	var visibility string
	switch resourceType {
	case "document":
		s.db.Model(&models.PersonalKnowledgeBase{}).
			Where("id = ?", resourceID).
			Select("privacy").
			Scan(&visibility)
	case "survey":
		s.db.Raw("SELECT CASE WHEN is_public THEN 'team_public' ELSE 'private' END FROM surveys WHERE survey_id = ?", resourceID).
			Scan(&visibility)
	default:
		visibility = "private"
	}

	// 缓存结果
	s.cache.Set(cacheKey, visibility, 5*time.Minute)

	return visibility
}

// getUserAccessProfile 获取用户访问档案
func (s *DataIsolationService) getUserAccessProfile(userID string) (*models.UserDataAccessProfile, error) {
	var profile models.UserDataAccessProfile

	err := s.db.Where("user_id = ?", userID).First(&profile).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			// 创建默认档案
			var user models.User
			if err := s.db.First(&user, "id = ?", userID).Error; err != nil {
				return nil, err
			}

			profile = models.UserDataAccessProfile{
				UserID:          userID,
				TeamID:          user.TeamID,
				PermissionLevel: "normal",
				SecuritySettings: struct {
					RequireApproval    bool       `json:"require_approval" gorm:"default:false"`
					MaxDailyAccess     int        `json:"max_daily_access" gorm:"default:100"`
					MaxCrossTeamAccess int        `json:"max_cross_team_access" gorm:"default:10"`
					IsBlocked          bool       `json:"is_blocked" gorm:"default:false"`
					BlockExpires       *time.Time `json:"block_expires"`
				}{
					RequireApproval:    false,
					MaxDailyAccess:     100,
					MaxCrossTeamAccess: 10,
					IsBlocked:          false,
				},
			}

			if err := s.db.Create(&profile).Error; err != nil {
				return nil, err
			}
		} else {
			return nil, err
		}
	}

	return &profile, nil
}

// getTeamIsolationPolicy 获取团队隔离策略
func (s *DataIsolationService) getTeamIsolationPolicy(teamID string) (*models.TeamDataIsolationPolicy, error) {
	// 从缓存获取
	cacheKey := fmt.Sprintf("team_isolation_policy:%s", teamID)
	if cached, err := s.cache.Get(cacheKey); err == nil {
		if policy, ok := cached.(*models.TeamDataIsolationPolicy); ok {
			return policy, nil
		}
	}

	var policy models.TeamDataIsolationPolicy
	err := s.db.Where("team_id = ?", teamID).First(&policy).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			// 创建默认策略
			policy = models.TeamDataIsolationPolicy{
				TeamID:              teamID,
				StrictIsolation:     true,
				AllowCrossTeamView:  false,
				AllowCrossTeamShare: false,
				VisibilitySettings: struct {
					SystemPublicAccess bool `json:"system_public_access" gorm:"default:true"`
					TeamPublicAccess   bool `json:"team_public_access" gorm:"default:true"`
					PrivateDataAccess  bool `json:"private_data_access" gorm:"default:false"`
				}{
					SystemPublicAccess: true,
					TeamPublicAccess:   true,
					PrivateDataAccess:  false,
				},
				AccessRestrictions: struct {
					DownloadRestriction bool     `json:"download_restriction" gorm:"default:true"`
					ShareRestriction    bool     `json:"share_restriction" gorm:"default:true"`
					ExportRestriction   bool     `json:"export_restriction" gorm:"default:true"`
					AllowedTeams        []string `json:"allowed_teams" gorm:"type:text[]"`
					RestrictedActions   []string `json:"restricted_actions" gorm:"type:text[]"`
				}{
					DownloadRestriction: true,
					ShareRestriction:    true,
					ExportRestriction:   true,
				},
				AuditSettings: struct {
					EnableAccessLog    bool `json:"enable_access_log" gorm:"default:true"`
					EnableOperationLog bool `json:"enable_operation_log" gorm:"default:true"`
					LogRetentionDays   int  `json:"log_retention_days" gorm:"default:365"`
					AlertOnViolation   bool `json:"alert_on_violation" gorm:"default:true"`
				}{
					EnableAccessLog:    true,
					EnableOperationLog: true,
					LogRetentionDays:   365,
					AlertOnViolation:   true,
				},
			}

			if err := s.db.Create(&policy).Error; err != nil {
				return nil, err
			}
		} else {
			return nil, err
		}
	}

	// 缓存结果
	s.cache.Set(cacheKey, &policy, 10*time.Minute)

	return &policy, nil
}

// checkDailyLimits 检查每日访问限制
func (s *DataIsolationService) checkDailyLimits(profile *models.UserDataAccessProfile, accessCtx *AccessContext) error {
	today := time.Now().Format("2006-01-02")

	// 检查每日总访问次数
	var totalCount int64
	s.db.Model(&models.DataAccessLog{}).
		Where("user_id = ? AND DATE(created_at) = ?", profile.UserID, today).
		Count(&totalCount)

	if int(totalCount) >= profile.SecuritySettings.MaxDailyAccess {
		return fmt.Errorf("Daily access limit exceeded: %d/%d", totalCount, profile.SecuritySettings.MaxDailyAccess)
	}

	// 如果是跨团队访问，检查跨团队访问限制
	if accessCtx.TeamID != accessCtx.OwnerTeamID {
		var crossTeamCount int64
		s.db.Model(&models.DataAccessLog{}).
			Where("user_id = ? AND DATE(created_at) = ? AND is_cross_team = true", profile.UserID, today).
			Count(&crossTeamCount)

		if int(crossTeamCount) >= profile.SecuritySettings.MaxCrossTeamAccess {
			return fmt.Errorf("Daily cross-team access limit exceeded: %d/%d", crossTeamCount, profile.SecuritySettings.MaxCrossTeamAccess)
		}
	}

	return nil
}

// hasValidCrossTeamRequest 检查是否有有效的跨团队访问申请
func (s *DataIsolationService) hasValidCrossTeamRequest(accessCtx *AccessContext) bool {
	var request models.CrossTeamAccessRequest
	err := s.db.Where(`
		requester_id = ? AND 
		target_team_id = ? AND 
		target_resource_id = ? AND 
		status = 'approved' AND 
		expires_at > ?
	`, accessCtx.UserID, accessCtx.OwnerTeamID, accessCtx.ResourceID, time.Now()).
		First(&request).Error

	return err == nil
}

// logAccess 记录访问日志
func (s *DataIsolationService) logAccess(accessCtx *AccessContext, startTime time.Time) {
	duration := time.Since(startTime).Milliseconds()

	log := models.DataAccessLog{
		UserID:       accessCtx.UserID,
		TeamID:       accessCtx.TeamID,
		UserRole:     accessCtx.UserRole,
		ResourceID:   accessCtx.ResourceID,
		ResourceType: accessCtx.ResourceType,
		ResourceName: accessCtx.Metadata["resource_name"],
		OwnerTeamID:  accessCtx.OwnerTeamID,
		ActionType:   accessCtx.ActionType,
		AccessMethod: accessCtx.RequestSource,
		IsCrossTeam:  accessCtx.TeamID != accessCtx.OwnerTeamID,
		IPAddress:    accessCtx.Metadata["ip_address"],
		UserAgent:    accessCtx.Metadata["user_agent"],
		SessionID:    accessCtx.Metadata["session_id"],
		Duration:     duration,
		Success:      true,
	}

	// 异步写入日志
	go func() {
		if err := s.db.Create(&log).Error; err != nil {
			// 记录错误但不影响主流程
			fmt.Printf("Failed to log access: %v\n", err)
		}
	}()
}

// recordViolation 记录违规行为
func (s *DataIsolationService) recordViolation(accessCtx *AccessContext, violationType, description, level string) {
	violation := models.DataIsolationViolation{
		UserID:           accessCtx.UserID,
		TeamID:           accessCtx.TeamID,
		UserRole:         accessCtx.UserRole,
		ViolationType:    violationType,
		ViolationLevel:   level,
		Description:      description,
		ResourceID:       accessCtx.ResourceID,
		ResourceType:     accessCtx.ResourceType,
		TargetTeamID:     accessCtx.OwnerTeamID,
		Status:           "open",
		AutoBlocked:      false,
		NotificationSent: false,
	}

	// 异步处理违规记录
	go func() {
		if err := s.db.Create(&violation).Error; err != nil {
			fmt.Printf("Failed to record violation: %v\n", err)
			return
		}

		// 检查是否需要自动封禁
		s.checkAutoBlock(accessCtx.UserID, level)

		// 发送告警通知
		s.sendViolationAlert(&violation)
	}()
}

// updateAccessStats 更新访问统计
func (s *DataIsolationService) updateAccessStats(profile *models.UserDataAccessProfile, accessCtx *AccessContext, allowed bool) {
	updates := map[string]interface{}{
		"stats_last_access_time": time.Now(),
	}

	if allowed {
		updates["stats_total_access"] = gorm.Expr("stats_total_access + 1")
		if accessCtx.TeamID != accessCtx.OwnerTeamID {
			updates["stats_cross_team_access"] = gorm.Expr("stats_cross_team_access + 1")
		}
	} else {
		updates["stats_violation_count"] = gorm.Expr("stats_violation_count + 1")
		updates["stats_last_violation_time"] = time.Now()
	}

	s.db.Model(profile).Updates(updates)
}

// checkAutoBlock 检查是否需要自动封禁
func (s *DataIsolationService) checkAutoBlock(userID, violationLevel string) {
	// 获取系统配置
	var config models.SystemVisibilityConfig
	s.db.First(&config)

	// 查询最近的违规次数
	var recentViolations int64
	s.db.Model(&models.DataIsolationViolation{}).
		Where("user_id = ? AND created_at > ? AND violation_level IN ('high', 'critical')",
			userID, time.Now().Add(-24*time.Hour)).
		Count(&recentViolations)

	if int(recentViolations) >= config.MaxViolationCount {
		// 自动封禁用户
		var profile models.UserDataAccessProfile
		if err := s.db.Where("user_id = ?", userID).First(&profile).Error; err == nil {
			blockExpires := time.Now().Add(time.Duration(config.ViolationBlockTime) * time.Minute)
			profile.SecuritySettings.IsBlocked = true
			profile.SecuritySettings.BlockExpires = &blockExpires
			s.db.Save(&profile)
		}
	}
}

// sendViolationAlert 发送违规告警
func (s *DataIsolationService) sendViolationAlert(violation *models.DataIsolationViolation) {
	// 这里可以集成通知系统发送告警
	// 例如：邮件、短信、企业微信等
	fmt.Printf("SECURITY ALERT: Violation detected - %s by user %s\n",
		violation.ViolationType, violation.UserID)
}
