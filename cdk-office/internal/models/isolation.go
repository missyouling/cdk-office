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

package models

import (
	"time"
)

// TeamDataIsolationPolicy 团队数据隔离策略
type TeamDataIsolationPolicy struct {
	ID     string `json:"id" gorm:"type:uuid;primary_key;default:gen_random_uuid()"`
	TeamID string `json:"team_id" gorm:"type:uuid;not null;uniqueIndex"`

	// 数据隔离配置
	StrictIsolation     bool `json:"strict_isolation" gorm:"default:true"`        // 严格隔离模式
	AllowCrossTeamView  bool `json:"allow_cross_team_view" gorm:"default:false"`  // 允许跨团队查看
	AllowCrossTeamShare bool `json:"allow_cross_team_share" gorm:"default:false"` // 允许跨团队分享

	// 可见性设置
	VisibilitySettings struct {
		SystemPublicAccess bool `json:"system_public_access" gorm:"default:true"` // 系统公开文档访问
		TeamPublicAccess   bool `json:"team_public_access" gorm:"default:true"`   // 团队公开文档访问
		PrivateDataAccess  bool `json:"private_data_access" gorm:"default:false"` // 私有数据跨团队访问
	} `json:"visibility_settings" gorm:"embedded;embeddedPrefix:vis_"`

	// 数据访问限制
	AccessRestrictions struct {
		DownloadRestriction bool     `json:"download_restriction" gorm:"default:true"` // 下载限制
		ShareRestriction    bool     `json:"share_restriction" gorm:"default:true"`    // 分享限制
		ExportRestriction   bool     `json:"export_restriction" gorm:"default:true"`   // 导出限制
		AllowedTeams        []string `json:"allowed_teams" gorm:"type:text[]"`         // 允许访问的团队ID列表
		RestrictedActions   []string `json:"restricted_actions" gorm:"type:text[]"`    // 受限制的操作
	} `json:"access_restrictions" gorm:"embedded;embeddedPrefix:restrict_"`

	// 审计设置
	AuditSettings struct {
		EnableAccessLog    bool `json:"enable_access_log" gorm:"default:true"`    // 启用访问日志
		EnableOperationLog bool `json:"enable_operation_log" gorm:"default:true"` // 启用操作日志
		LogRetentionDays   int  `json:"log_retention_days" gorm:"default:365"`    // 日志保留天数
		AlertOnViolation   bool `json:"alert_on_violation" gorm:"default:true"`   // 违规时告警
	} `json:"audit_settings" gorm:"embedded;embeddedPrefix:audit_"`

	CreatedBy string    `json:"created_by" gorm:"type:uuid;not null"`
	UpdatedBy string    `json:"updated_by" gorm:"type:uuid"`
	CreatedAt time.Time `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt time.Time `json:"updated_at" gorm:"autoUpdateTime"`
}

// DataAccessLog 数据访问日志
type DataAccessLog struct {
	ID string `json:"id" gorm:"type:uuid;primary_key;default:gen_random_uuid()"`

	// 访问者信息
	UserID   string `json:"user_id" gorm:"type:uuid;not null;index"`
	TeamID   string `json:"team_id" gorm:"type:uuid;not null;index"`
	UserRole string `json:"user_role" gorm:"size:50;not null"`

	// 资源信息
	ResourceID   string `json:"resource_id" gorm:"type:uuid;not null;index"`
	ResourceType string `json:"resource_type" gorm:"size:50;not null"` // document, knowledge, survey, etc.
	ResourceName string `json:"resource_name" gorm:"size:255"`
	OwnerTeamID  string `json:"owner_team_id" gorm:"type:uuid;not null;index"`

	// 访问详情
	ActionType   string `json:"action_type" gorm:"size:50;not null"` // view, download, share, edit, delete
	AccessMethod string `json:"access_method" gorm:"size:50"`        // web, api, mobile
	IsCrossTeam  bool   `json:"is_cross_team" gorm:"default:false"`  // 是否跨团队访问
	IsViolation  bool   `json:"is_violation" gorm:"default:false"`   // 是否违规访问
	ViolationMsg string `json:"violation_msg" gorm:"size:500"`       // 违规信息

	// 请求信息
	IPAddress string `json:"ip_address" gorm:"size:45"`
	UserAgent string `json:"user_agent" gorm:"size:500"`
	SessionID string `json:"session_id" gorm:"size:100"`

	// 结果信息
	Success  bool   `json:"success" gorm:"default:true"`
	ErrorMsg string `json:"error_msg" gorm:"size:500"`
	Duration int64  `json:"duration"` // 访问耗时(毫秒)

	CreatedAt time.Time `json:"created_at" gorm:"autoCreateTime"`
}

// SystemVisibilityConfig 系统可见性配置
type SystemVisibilityConfig struct {
	ID string `json:"id" gorm:"type:uuid;primary_key;default:gen_random_uuid()"`

	// 全局可见性设置
	GlobalSettings struct {
		AllowGlobalSearch    bool `json:"allow_global_search" gorm:"default:false"`     // 允许全局搜索
		AllowCrossTeamAccess bool `json:"allow_cross_team_access" gorm:"default:false"` // 允许跨团队访问
		RequireApproval      bool `json:"require_approval" gorm:"default:true"`         // 跨团队操作需审批
	} `json:"global_settings" gorm:"embedded;embeddedPrefix:global_"`

	// 角色权限映射
	RolePermissions struct {
		SuperAdminAccess   []string `json:"super_admin_access" gorm:"type:text[]"`  // 超级管理员权限
		TeamManagerAccess  []string `json:"team_manager_access" gorm:"type:text[]"` // 团队管理员权限
		CollaboratorAccess []string `json:"collaborator_access" gorm:"type:text[]"` // 协作用户权限
		NormalUserAccess   []string `json:"normal_user_access" gorm:"type:text[]"`  // 普通用户权限
	} `json:"role_permissions" gorm:"embedded;embeddedPrefix:role_"`

	// 数据分类配置
	DataClassification struct {
		SystemPublic map[string]interface{} `json:"system_public" gorm:"type:jsonb"` // 系统公开数据规则
		TeamPublic   map[string]interface{} `json:"team_public" gorm:"type:jsonb"`   // 团队公开数据规则
		Private      map[string]interface{} `json:"private" gorm:"type:jsonb"`       // 私有数据规则
		Confidential map[string]interface{} `json:"confidential" gorm:"type:jsonb"`  // 机密数据规则
	} `json:"data_classification" gorm:"embedded;embeddedPrefix:class_"`

	// 审计配置
	AuditEnabled       bool `json:"audit_enabled" gorm:"default:true"`
	AlertEnabled       bool `json:"alert_enabled" gorm:"default:true"`
	MaxViolationCount  int  `json:"max_violation_count" gorm:"default:5"`   // 最大违规次数
	ViolationBlockTime int  `json:"violation_block_time" gorm:"default:30"` // 违规封禁时间(分钟)

	CreatedBy string    `json:"created_by" gorm:"type:uuid;not null"`
	UpdatedBy string    `json:"updated_by" gorm:"type:uuid"`
	CreatedAt time.Time `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt time.Time `json:"updated_at" gorm:"autoUpdateTime"`
}

// CrossTeamAccessRequest 跨团队访问申请
type CrossTeamAccessRequest struct {
	ID string `json:"id" gorm:"type:uuid;primary_key;default:gen_random_uuid()"`

	// 申请者信息
	RequesterID     string `json:"requester_id" gorm:"type:uuid;not null;index"`
	RequesterTeamID string `json:"requester_team_id" gorm:"type:uuid;not null"`

	// 目标信息
	TargetTeamID       string `json:"target_team_id" gorm:"type:uuid;not null;index"`
	TargetResourceID   string `json:"target_resource_id" gorm:"type:uuid;not null"`
	TargetResourceType string `json:"target_resource_type" gorm:"size:50;not null"`

	// 申请详情
	RequestType      string `json:"request_type" gorm:"size:50;not null"`     // view, download, share, collaborate
	RequestReason    string `json:"request_reason" gorm:"type:text;not null"` // 申请原因
	ExpectedDuration int    `json:"expected_duration" gorm:"default:7"`       // 预期使用天数

	// 审批信息
	Status         string     `json:"status" gorm:"size:20;default:'pending'"` // pending, approved, rejected
	ApproverID     string     `json:"approver_id" gorm:"type:uuid"`
	ApprovalReason string     `json:"approval_reason" gorm:"type:text"`
	ApprovedAt     *time.Time `json:"approved_at"`
	ExpiresAt      *time.Time `json:"expires_at"`

	CreatedAt time.Time `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt time.Time `json:"updated_at" gorm:"autoUpdateTime"`
}

// DataIsolationViolation 数据隔离违规记录
type DataIsolationViolation struct {
	ID string `json:"id" gorm:"type:uuid;primary_key;default:gen_random_uuid()"`

	// 违规者信息
	UserID   string `json:"user_id" gorm:"type:uuid;not null;index"`
	TeamID   string `json:"team_id" gorm:"type:uuid;not null"`
	UserRole string `json:"user_role" gorm:"size:50;not null"`

	// 违规详情
	ViolationType  string `json:"violation_type" gorm:"size:50;not null"`  // access_denied, data_leak, unauthorized_share
	ViolationLevel string `json:"violation_level" gorm:"size:20;not null"` // low, medium, high, critical
	Description    string `json:"description" gorm:"type:text;not null"`
	ResourceID     string `json:"resource_id" gorm:"type:uuid"`
	ResourceType   string `json:"resource_type" gorm:"size:50"`
	TargetTeamID   string `json:"target_team_id" gorm:"type:uuid"`

	// 处理信息
	Status     string     `json:"status" gorm:"size:20;default:'open'"` // open, investigating, resolved, ignored
	HandlerID  string     `json:"handler_id" gorm:"type:uuid"`
	HandledAt  *time.Time `json:"handled_at"`
	Resolution string     `json:"resolution" gorm:"type:text"`

	// 自动处理
	AutoBlocked      bool `json:"auto_blocked" gorm:"default:false"`
	BlockDuration    int  `json:"block_duration"` // 自动封禁时长(分钟)
	NotificationSent bool `json:"notification_sent" gorm:"default:false"`

	CreatedAt time.Time `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt time.Time `json:"updated_at" gorm:"autoUpdateTime"`
}

// UserDataAccessProfile 用户数据访问档案
type UserDataAccessProfile struct {
	ID     string `json:"id" gorm:"type:uuid;primary_key;default:gen_random_uuid()"`
	UserID string `json:"user_id" gorm:"type:uuid;not null;uniqueIndex"`
	TeamID string `json:"team_id" gorm:"type:uuid;not null"`

	// 访问统计
	AccessStats struct {
		TotalAccess       int64     `json:"total_access"`        // 总访问次数
		CrossTeamAccess   int64     `json:"cross_team_access"`   // 跨团队访问次数
		ViolationCount    int64     `json:"violation_count"`     // 违规次数
		LastAccessTime    time.Time `json:"last_access_time"`    // 最后访问时间
		LastViolationTime time.Time `json:"last_violation_time"` // 最后违规时间
	} `json:"access_stats" gorm:"embedded;embeddedPrefix:stats_"`

	// 权限级别
	PermissionLevel   string   `json:"permission_level" gorm:"size:20;default:'normal'"` // normal, elevated, restricted
	AllowedTeams      []string `json:"allowed_teams" gorm:"type:text[]"`                 // 允许访问的团队
	RestrictedActions []string `json:"restricted_actions" gorm:"type:text[]"`            // 受限操作

	// 安全设置
	SecuritySettings struct {
		RequireApproval    bool       `json:"require_approval" gorm:"default:false"`   // 需要审批
		MaxDailyAccess     int        `json:"max_daily_access" gorm:"default:100"`     // 每日最大访问次数
		MaxCrossTeamAccess int        `json:"max_cross_team_access" gorm:"default:10"` // 每日最大跨团队访问次数
		IsBlocked          bool       `json:"is_blocked" gorm:"default:false"`         // 是否被封禁
		BlockExpires       *time.Time `json:"block_expires"`                           // 封禁到期时间
	} `json:"security_settings" gorm:"embedded;embeddedPrefix:security_"`

	CreatedAt time.Time `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt time.Time `json:"updated_at" gorm:"autoUpdateTime"`
}
