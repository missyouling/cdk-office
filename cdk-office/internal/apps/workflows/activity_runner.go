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
	"fmt"
	"log"
	"time"

	"gorm.io/gorm"
)

// ActivityFunction 活动函数类型
type ActivityFunction func(ctx context.Context, config map[string]interface{}, variables map[string]interface{}) (map[string]interface{}, error)

// ActivityRunner 活动执行器
type ActivityRunner struct {
	db         *gorm.DB
	activities map[string]ActivityFunction
}

// NewActivityRunner 创建活动执行器
func NewActivityRunner(db *gorm.DB) *ActivityRunner {
	return &ActivityRunner{
		db:         db,
		activities: make(map[string]ActivityFunction),
	}
}

// RegisterActivity 注册活动
func (r *ActivityRunner) RegisterActivity(name string, fn ActivityFunction) {
	r.activities[name] = fn
	log.Printf("Registered activity: %s", name)
}

// ExecuteActivity 执行活动
func (r *ActivityRunner) ExecuteActivity(name string, config map[string]interface{}, variables map[string]interface{}) (map[string]interface{}, error) {
	fn, exists := r.activities[name]
	if !exists {
		return nil, fmt.Errorf("activity not found: %s", name)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	log.Printf("Executing activity: %s", name)

	result, err := fn(ctx, config, variables)
	if err != nil {
		log.Printf("Activity %s failed: %v", name, err)
		return nil, err
	}

	log.Printf("Activity %s completed successfully", name)
	return result, nil
}

// ListActivities 获取已注册的活动列表
func (r *ActivityRunner) ListActivities() []string {
	var activities []string
	for name := range r.activities {
		activities = append(activities, name)
	}
	return activities
}

// 内置活动实现

// SendEmailActivity 发送邮件活动
func SendEmailActivity(ctx context.Context, config map[string]interface{}, variables map[string]interface{}) (map[string]interface{}, error) {
	// 获取邮件配置
	to, ok := config["to"].(string)
	if !ok {
		// 从变量中获取收件人
		if recipient, exists := variables["email_recipient"]; exists {
			to = fmt.Sprintf("%v", recipient)
		} else {
			return nil, fmt.Errorf("email recipient not specified")
		}
	}

	subject, _ := config["subject"].(string)
	if subject == "" {
		subject = "Workflow Notification"
	}

	body, _ := config["body"].(string)
	if body == "" {
		body = "This is a notification from the workflow system."
	}

	// 替换模板变量
	body = replaceTemplateVariables(body, variables)
	subject = replaceTemplateVariables(subject, variables)

	// 模拟发送邮件（实际实现中需要集成真实的邮件服务）
	log.Printf("Sending email to: %s, subject: %s", to, subject)
	log.Printf("Email body: %s", body)

	// 这里应该集成实际的邮件发送服务
	// 例如：SMTP、SendGrid、AWS SES等

	return map[string]interface{}{
		"status":    "sent",
		"recipient": to,
		"sent_at":   time.Now(),
	}, nil
}

// SendNotificationActivity 发送通知活动
func SendNotificationActivity(ctx context.Context, config map[string]interface{}, variables map[string]interface{}) (map[string]interface{}, error) {
	// 获取通知配置
	userID, ok := config["user_id"].(string)
	if !ok {
		if recipient, exists := variables["notification_recipient"]; exists {
			userID = fmt.Sprintf("%v", recipient)
		} else {
			return nil, fmt.Errorf("notification recipient not specified")
		}
	}

	title, _ := config["title"].(string)
	if title == "" {
		title = "Workflow Notification"
	}

	message, _ := config["message"].(string)
	if message == "" {
		message = "You have a new notification from the workflow system."
	}

	// 替换模板变量
	title = replaceTemplateVariables(title, variables)
	message = replaceTemplateVariables(message, variables)

	// 模拟发送通知（实际实现中需要集成通知系统）
	log.Printf("Sending notification to user: %s", userID)
	log.Printf("Title: %s", title)
	log.Printf("Message: %s", message)

	// 这里应该集成实际的通知系统
	// 例如：推送通知、站内信、短信等

	return map[string]interface{}{
		"status":  "sent",
		"user_id": userID,
		"sent_at": time.Now(),
	}, nil
}

// UpdateDocumentActivity 更新文档活动
func UpdateDocumentActivity(ctx context.Context, config map[string]interface{}, variables map[string]interface{}) (map[string]interface{}, error) {
	// 获取文档ID
	documentID, ok := config["document_id"].(string)
	if !ok {
		if docID, exists := variables["document_id"]; exists {
			documentID = fmt.Sprintf("%v", docID)
		} else {
			return nil, fmt.Errorf("document ID not specified")
		}
	}

	// 获取更新状态
	status, ok := config["status"].(string)
	if !ok {
		status = "approved"
	}

	// 模拟更新文档状态
	log.Printf("Updating document %s status to: %s", documentID, status)

	// 这里应该集成实际的文档管理系统
	// 例如：更新数据库中的文档状态

	return map[string]interface{}{
		"status":      "updated",
		"document_id": documentID,
		"new_status":  status,
		"updated_at":  time.Now(),
	}, nil
}

// LogEventActivity 记录事件活动
func LogEventActivity(ctx context.Context, config map[string]interface{}, variables map[string]interface{}) (map[string]interface{}, error) {
	// 获取事件信息
	event, _ := config["event"].(string)
	if event == "" {
		event = "workflow_event"
	}

	level, _ := config["level"].(string)
	if level == "" {
		level = "info"
	}

	message, _ := config["message"].(string)
	if message == "" {
		message = "Workflow event occurred"
	}

	// 替换模板变量
	message = replaceTemplateVariables(message, variables)

	// 记录事件
	log.Printf("[%s] %s: %s", level, event, message)

	// 这里应该集成实际的日志系统
	// 例如：写入日志文件、发送到日志聚合服务等

	return map[string]interface{}{
		"status":    "logged",
		"event":     event,
		"level":     level,
		"message":   message,
		"logged_at": time.Now(),
	}, nil
}

// HTTPRequestActivity HTTP请求活动
func HTTPRequestActivity(ctx context.Context, config map[string]interface{}, variables map[string]interface{}) (map[string]interface{}, error) {
	// 获取请求配置
	url, ok := config["url"].(string)
	if !ok {
		return nil, fmt.Errorf("URL not specified")
	}

	method, _ := config["method"].(string)
	if method == "" {
		method = "GET"
	}

	// 替换URL中的模板变量
	url = replaceTemplateVariables(url, variables)

	// 模拟HTTP请求
	log.Printf("Making %s request to: %s", method, url)

	// 这里应该实现真实的HTTP请求
	// 例如：使用http.Client发送请求

	return map[string]interface{}{
		"status":      "completed",
		"method":      method,
		"url":         url,
		"response":    "success",
		"executed_at": time.Now(),
	}, nil
}

// DelayActivity 延迟活动
func DelayActivity(ctx context.Context, config map[string]interface{}, variables map[string]interface{}) (map[string]interface{}, error) {
	// 获取延迟时间
	durationStr, ok := config["duration"].(string)
	if !ok {
		return nil, fmt.Errorf("duration not specified")
	}

	duration, err := time.ParseDuration(durationStr)
	if err != nil {
		return nil, fmt.Errorf("invalid duration format: %v", err)
	}

	log.Printf("Delaying for: %v", duration)

	// 执行延迟
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	case <-time.After(duration):
		// 延迟完成
	}

	log.Printf("Delay completed")

	return map[string]interface{}{
		"status":       "completed",
		"duration":     duration.String(),
		"completed_at": time.Now(),
	}, nil
}

// ScriptActivity 脚本执行活动
func ScriptActivity(ctx context.Context, config map[string]interface{}, variables map[string]interface{}) (map[string]interface{}, error) {
	// 获取脚本配置
	script, ok := config["script"].(string)
	if !ok {
		return nil, fmt.Errorf("script not specified")
	}

	scriptType, _ := config["type"].(string)
	if scriptType == "" {
		scriptType = "shell"
	}

	// 替换脚本中的模板变量
	script = replaceTemplateVariables(script, variables)

	log.Printf("Executing %s script: %s", scriptType, script)

	// 这里应该实现真实的脚本执行
	// 例如：使用os/exec包执行shell脚本
	// 注意：需要考虑安全性，限制可执行的命令

	return map[string]interface{}{
		"status":      "completed",
		"script_type": scriptType,
		"output":      "script executed successfully",
		"executed_at": time.Now(),
	}, nil
}

// 辅助函数

// replaceTemplateVariables 替换模板变量
func replaceTemplateVariables(template string, variables map[string]interface{}) string {
	result := template

	// 简单的模板变量替换实现
	// 实际项目中可以使用更强大的模板引擎，如text/template
	for key, value := range variables {
		// 暂时简化实现，实际应用中需要实现字符串替换逻辑
		_ = key
		_ = value
	}

	return result
}
