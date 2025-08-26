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

package task

import (
	"context"
	"encoding/json"
	"time"

	"github.com/linux-do/cdk-office/internal/db"
	"github.com/linux-do/cdk-office/internal/logger"
	"github.com/redis/go-redis/v9"
)

// Task 任务结构
type Task struct {
	ID        string      `json:"id"`
	Type      string      `json:"type"`
	Payload   interface{} `json:"payload"`
	Status    string      `json:"status"`
	CreatedAt time.Time   `json:"created_at"`
	UpdatedAt time.Time   `json:"updated_at"`
}

// EnqueueTask 入队任务
func EnqueueTask(ctx context.Context, queueName string, taskType string, payload interface{}) error {
	// 创建任务
	task := &Task{
		ID:        generateTaskID(),
		Type:      taskType,
		Payload:   payload,
		Status:    TaskPending,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	// 序列化任务
	taskData, err := json.Marshal(task)
	if err != nil {
		return err
	}

	// 入队
	redisClient := db.GetRedis()
	if err := redisClient.LPush(ctx, queueName, taskData).Err(); err != nil {
		return err
	}

	logger.Info("task enqueued: %s, type: %s, queue: %s", task.ID, taskType, queueName)
	return nil
}

// DequeueTask 出队任务
func DequeueTask(ctx context.Context, queueName string) (*Task, error) {
	// 出队
	redisClient := db.GetRedis()
	taskData, err := redisClient.BRPop(ctx, 0, queueName).Result()
	if err != nil {
		if err == redis.Nil {
			return nil, nil
		}
		return nil, err
	}

	// 反序列化任务
	var task Task
	if err := json.Unmarshal([]byte(taskData[1]), &task); err != nil {
		return nil, err
	}

	// 更新任务状态
	task.Status = TaskProcessing
	task.UpdatedAt = time.Now()

	logger.Info("task dequeued: %s, type: %s, queue: %s", task.ID, task.Type, queueName)
	return &task, nil
}

// CompleteTask 完成任务
func CompleteTask(ctx context.Context, task *Task) error {
	task.Status = TaskCompleted
	task.UpdatedAt = time.Now()

	logger.Info("task completed: %s, type: %s", task.ID, task.Type)
	return nil
}

// FailTask 失败任务
func FailTask(ctx context.Context, task *Task, errMsg string) error {
	task.Status = TaskFailed
	task.UpdatedAt = time.Now()

	logger.Error("task failed: %s, type: %s, error: %s", task.ID, task.Type, errMsg)
	return nil
}

// generateTaskID 生成任务ID
func generateTaskID() string {
	// 这里可以使用UUID或其他方式生成唯一ID
	return "task_" + time.Now().Format("20060102150405") + "_" + randomString(6)
}

// randomString 生成随机字符串
func randomString(length int) string {
	// 简单实现，实际项目中可以使用更好的随机字符串生成方法
	return "random"
}
