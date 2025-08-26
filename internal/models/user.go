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

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// User 用户模型
type User struct {
	ID            string    `json:"id" gorm:"type:uuid;primary_key;default:gen_random_uuid()"`
	Username      string    `json:"username" gorm:"size:50;uniqueIndex;not null"`
	Email         string    `json:"email" gorm:"size:100;uniqueIndex"`
	Phone         string    `json:"phone" gorm:"size:20;uniqueIndex"`
	PasswordHash  string    `json:"-" gorm:"size:255"` // 不返回给前端
	Nickname      string    `json:"nickname" gorm:"size:100"`
	AvatarURL     string    `json:"avatar_url" gorm:"size:500"`
	WeChatOpenID  string    `json:"wechat_open_id" gorm:"size:100;uniqueIndex"`
	WeComUserID   string    `json:"wecom_user_id" gorm:"size:100;uniqueIndex"`
	CurrentTeamID string    `json:"current_team_id" gorm:"type:uuid"`
	IsActive      bool      `json:"is_active" gorm:"default:true"`
	LastLoginAt   time.Time `json:"last_login_at"`
	CreatedAt     time.Time `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt     time.Time `json:"updated_at" gorm:"autoUpdateTime"`
}

// BeforeCreate 创建前钩子
func (u *User) BeforeCreate(tx *gorm.DB) error {
	if u.ID == "" {
		u.ID = uuid.New().String()
	}
	return nil
}

// UserRole 用户角色模型
type UserRole struct {
	ID        string    `json:"id" gorm:"type:uuid;primary_key;default:gen_random_uuid()"`
	UserID    string    `json:"user_id" gorm:"type:uuid;not null;index"`
	TeamID    string    `json:"team_id" gorm:"type:uuid;not null;index"`
	Role      string    `json:"role" gorm:"size:20;not null"` // super_admin, team_manager, collaborator, normal_user
	CreatedAt time.Time `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt time.Time `json:"updated_at" gorm:"autoUpdateTime"`
}

// Team 团队模型
type Team struct {
	ID          string    `json:"id" gorm:"type:uuid;primary_key;default:gen_random_uuid()"`
	Name        string    `json:"name" gorm:"size:100;not null"`
	Description string    `json:"description" gorm:"type:text"`
	CreatedBy   string    `json:"created_by" gorm:"type:uuid;not null"`
	IsActive    bool      `json:"is_active" gorm:"default:true"`
	CreatedAt   time.Time `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt   time.Time `json:"updated_at" gorm:"autoUpdateTime"`
}

// TeamMember 团队成员模型
type TeamMember struct {
	ID        string    `json:"id" gorm:"type:uuid;primary_key;default:gen_random_uuid()"`
	TeamID    string    `json:"team_id" gorm:"type:uuid;not null;index"`
	UserID    string    `json:"user_id" gorm:"type:uuid;not null;index"`
	Role      string    `json:"role" gorm:"size:20;not null"` // owner, admin, member
	JoinedAt  time.Time `json:"joined_at" gorm:"autoCreateTime"`
	CreatedAt time.Time `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt time.Time `json:"updated_at" gorm:"autoUpdateTime"`
}
