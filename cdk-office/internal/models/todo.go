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

// TodoItem 待办事项模型
type TodoItem struct {
	ID        string     `json:"id" gorm:"type:uuid;primary_key;default:gen_random_uuid()"`
	UserID    string     `json:"user_id" gorm:"type:uuid;not null;index"`
	TeamID    string     `json:"team_id" gorm:"type:uuid;not null;index"`
	Title     string     `json:"title" gorm:"size:255;not null"`
	Completed bool       `json:"completed" gorm:"default:false"`
	DueDate   *time.Time `json:"due_date,omitempty" gorm:"type:timestamp"`
	CreatedAt time.Time  `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt time.Time  `json:"updated_at" gorm:"autoUpdateTime"`
}

// BeforeCreate 创建前钩子
func (t *TodoItem) BeforeCreate(tx *gorm.DB) error {
	if t.ID == "" {
		t.ID = uuid.New().String()
	}
	return nil
}

// CalendarEvent 日程事件模型
type CalendarEvent struct {
	ID          string    `json:"id" gorm:"type:uuid;primary_key;default:gen_random_uuid()"`
	UserID      string    `json:"user_id" gorm:"type:uuid;not null;index"`
	TeamID      string    `json:"team_id" gorm:"type:uuid;not null;index"`
	Title       string    `json:"title" gorm:"size:255;not null"`
	Description string    `json:"description,omitempty" gorm:"type:text"`
	StartTime   time.Time `json:"start_time" gorm:"type:timestamp;not null"`
	EndTime     time.Time `json:"end_time" gorm:"type:timestamp;not null"`
	AllDay      bool      `json:"all_day" gorm:"default:false"`
	CreatedAt   time.Time `json:"created_at" gorm:"autoCreateTime"`
}

// BeforeCreate 创建前钩子
func (c *CalendarEvent) BeforeCreate(tx *gorm.DB) error {
	if c.ID == "" {
		c.ID = uuid.New().String()
	}
	return nil
}

// TodoItemResponse 待办事项响应结构
type TodoItemResponse struct {
	ID        string     `json:"id"`
	Title     string     `json:"title"`
	Completed bool       `json:"completed"`
	DueDate   *time.Time `json:"due_date,omitempty"`
	CreatedAt time.Time  `json:"created_at"`
	UpdatedAt time.Time  `json:"updated_at"`
	IsOverdue bool       `json:"is_overdue"`
}

// CalendarEventResponse 日程事件响应结构
type CalendarEventResponse struct {
	ID          string    `json:"id"`
	Title       string    `json:"title"`
	Description string    `json:"description,omitempty"`
	StartTime   time.Time `json:"start_time"`
	EndTime     time.Time `json:"end_time"`
	AllDay      bool      `json:"all_day"`
	CreatedAt   time.Time `json:"created_at"`
}

// ToResponse 转换为响应结构
func (t *TodoItem) ToResponse() TodoItemResponse {
	isOverdue := false
	if t.DueDate != nil && t.DueDate.Before(time.Now()) && !t.Completed {
		isOverdue = true
	}

	return TodoItemResponse{
		ID:        t.ID,
		Title:     t.Title,
		Completed: t.Completed,
		DueDate:   t.DueDate,
		CreatedAt: t.CreatedAt,
		UpdatedAt: t.UpdatedAt,
		IsOverdue: isOverdue,
	}
}

// ToResponse 转换为响应结构
func (c *CalendarEvent) ToResponse() CalendarEventResponse {
	return CalendarEventResponse{
		ID:          c.ID,
		Title:       c.Title,
		Description: c.Description,
		StartTime:   c.StartTime,
		EndTime:     c.EndTime,
		AllDay:      c.AllDay,
		CreatedAt:   c.CreatedAt,
	}
}
