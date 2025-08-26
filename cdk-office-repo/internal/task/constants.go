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

// 任务类型常量
const (
	// DocumentSyncTask 文档同步任务
	DocumentSyncTask = "document_sync"

	// OCRProcessTask OCR处理任务
	OCRProcessTask = "ocr_process"

	// AITask AI处理任务
	AITask = "ai_process"

	// ArchiveTask 归档任务
	ArchiveTask = "archive"

	// NotificationTask 通知任务
	NotificationTask = "notification"
)

// 任务状态常量
const (
	// TaskPending 待处理
	TaskPending = "pending"

	// TaskProcessing 处理中
	TaskProcessing = "processing"

	// TaskCompleted 已完成
	TaskCompleted = "completed"

	// TaskFailed 失败
	TaskFailed = "failed"

	// TaskCancelled 已取消
	TaskCancelled = "cancelled"
)

// 队列名称常量
const (
	// DocumentQueue 文档处理队列
	DocumentQueue = "document_queue"

	// OCRQueue OCR处理队列
	OCRQueue = "ocr_queue"

	// AIQueue AI处理队列
	AIQueue = "ai_queue"

	// NotificationQueue 通知队列
	NotificationQueue = "notification_queue"
)
