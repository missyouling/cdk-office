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

package dify

import (
	"context"
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/linux-do/cdk-office/internal/models"
	"gorm.io/gorm"
)

// SyncManager Dify知识库同步管理器
type SyncManager struct {
	db         *gorm.DB
	client     *Client
	config     *Config
	syncQueue  chan *SyncTask
	retryQueue chan *SyncTask
	workers    int
	stopChan   chan struct{}
	wg         sync.WaitGroup
	mu         sync.RWMutex
	syncStats  *SyncStatistics
	isRunning  bool
}

// SyncTask 同步任务
type SyncTask struct {
	ID           string                 `json:"id"`
	Type         string                 `json:"type"` // create, update, delete
	DocumentID   string                 `json:"document_id"`
	Title        string                 `json:"title"`
	Content      string                 `json:"content"`
	DocumentType string                 `json:"document_type"`
	TeamID       string                 `json:"team_id"`
	UserID       string                 `json:"user_id"`
	Metadata     map[string]interface{} `json:"metadata"`
	Priority     int                    `json:"priority"` // 1-5, 5为最高优先级
	RetryCount   int                    `json:"retry_count"`
	MaxRetries   int                    `json:"max_retries"`
	CreatedAt    time.Time              `json:"created_at"`
	UpdatedAt    time.Time              `json:"updated_at"`
}

// SyncStatistics 同步统计信息
type SyncStatistics struct {
	TotalTasks     int64     `json:"total_tasks"`
	CompletedTasks int64     `json:"completed_tasks"`
	FailedTasks    int64     `json:"failed_tasks"`
	PendingTasks   int64     `json:"pending_tasks"`
	LastSyncTime   time.Time `json:"last_sync_time"`
	AverageTime    float64   `json:"average_time"` // 平均处理时间（秒）
	SuccessRate    float64   `json:"success_rate"`
	LastUpdateTime time.Time `json:"last_update_time"`
}

// SyncConfig 同步配置
type SyncConfig struct {
	AutoSyncEnabled     bool          `json:"auto_sync_enabled"`
	SyncInterval        time.Duration `json:"sync_interval"`
	BatchSize           int           `json:"batch_size"`
	MaxRetries          int           `json:"max_retries"`
	RetryDelay          time.Duration `json:"retry_delay"`
	WorkerCount         int           `json:"worker_count"`
	EnablePriorityQueue bool          `json:"enable_priority_queue"`
	MaxQueueSize        int           `json:"max_queue_size"`
}

// NewSyncManager 创建同步管理器
func NewSyncManager(db *gorm.DB, client *Client, config *Config) *SyncManager {
	syncConfig := &SyncConfig{
		AutoSyncEnabled:     config.EnableAutoSync,
		SyncInterval:        time.Duration(config.SyncInterval) * time.Minute,
		BatchSize:           10,
		MaxRetries:          3,
		RetryDelay:          time.Minute * 5,
		WorkerCount:         5,
		EnablePriorityQueue: true,
		MaxQueueSize:        1000,
	}

	return &SyncManager{
		db:         db,
		client:     client,
		config:     config,
		syncQueue:  make(chan *SyncTask, syncConfig.MaxQueueSize),
		retryQueue: make(chan *SyncTask, syncConfig.MaxQueueSize),
		workers:    syncConfig.WorkerCount,
		stopChan:   make(chan struct{}),
		syncStats: &SyncStatistics{
			LastUpdateTime: time.Now(),
		},
	}
}

// Start 启动同步管理器
func (sm *SyncManager) Start(ctx context.Context) error {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	if sm.isRunning {
		return fmt.Errorf("sync manager is already running")
	}

	sm.isRunning = true
	log.Printf("[SyncManager] Starting Dify sync manager with %d workers", sm.workers)

	// 启动工作协程
	for i := 0; i < sm.workers; i++ {
		sm.wg.Add(1)
		go sm.worker(ctx, i)
	}

	// 启动重试协程
	sm.wg.Add(1)
	go sm.retryWorker(ctx)

	// 启动自动同步协程
	if sm.config.EnableAutoSync {
		sm.wg.Add(1)
		go sm.autoSyncWorker(ctx)
	}

	// 启动统计更新协程
	sm.wg.Add(1)
	go sm.statsUpdater(ctx)

	// 加载待处理的同步任务
	if err := sm.loadPendingTasks(); err != nil {
		log.Printf("[SyncManager] Failed to load pending tasks: %v", err)
	}

	log.Printf("[SyncManager] Dify sync manager started successfully")
	return nil
}

// Stop 停止同步管理器
func (sm *SyncManager) Stop() {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	if !sm.isRunning {
		return
	}

	log.Printf("[SyncManager] Stopping Dify sync manager...")
	close(sm.stopChan)
	sm.wg.Wait()

	sm.isRunning = false
	log.Printf("[SyncManager] Dify sync manager stopped")
}

// AddSyncTask 添加同步任务
func (sm *SyncManager) AddSyncTask(task *SyncTask) error {
	if task.MaxRetries == 0 {
		task.MaxRetries = 3
	}
	if task.CreatedAt.IsZero() {
		task.CreatedAt = time.Now()
	}
	task.UpdatedAt = time.Now()

	// 保存到数据库
	syncRecord := &models.DifyDocumentSync{
		DocumentID:   task.DocumentID,
		Title:        task.Title,
		Content:      task.Content,
		DocumentType: task.DocumentType,
		TeamID:       task.TeamID,
		SyncStatus:   "pending",
		CreatedBy:    task.UserID,
	}

	if err := sm.db.Create(syncRecord).Error; err != nil {
		return fmt.Errorf("failed to save sync task: %w", err)
	}

	task.ID = syncRecord.ID

	// 添加到队列
	select {
	case sm.syncQueue <- task:
		sm.updateStats(func(stats *SyncStatistics) {
			stats.TotalTasks++
			stats.PendingTasks++
		})
		return nil
	default:
		return fmt.Errorf("sync queue is full")
	}
}

// worker 同步工作协程
func (sm *SyncManager) worker(ctx context.Context, workerID int) {
	defer sm.wg.Done()

	log.Printf("[SyncManager] Worker %d started", workerID)

	for {
		select {
		case <-ctx.Done():
			log.Printf("[SyncManager] Worker %d stopping due to context cancellation", workerID)
			return
		case <-sm.stopChan:
			log.Printf("[SyncManager] Worker %d stopping", workerID)
			return
		case task := <-sm.syncQueue:
			sm.processTask(ctx, task, workerID)
		}
	}
}

// processTask 处理同步任务
func (sm *SyncManager) processTask(ctx context.Context, task *SyncTask, workerID int) {
	startTime := time.Now()
	log.Printf("[SyncManager] Worker %d processing task %s (type: %s, doc: %s)",
		workerID, task.ID, task.Type, task.DocumentID)

	var err error
	var response *DocumentSyncResponse

	switch task.Type {
	case "create":
		response, err = sm.createDocument(ctx, task)
	case "update":
		response, err = sm.updateDocument(ctx, task)
	case "delete":
		err = sm.deleteDocument(ctx, task)
	default:
		err = fmt.Errorf("unknown task type: %s", task.Type)
	}

	duration := time.Since(startTime)

	if err != nil {
		log.Printf("[SyncManager] Worker %d failed to process task %s: %v", workerID, task.ID, err)
		sm.handleTaskFailure(task, err)
	} else {
		log.Printf("[SyncManager] Worker %d completed task %s in %v", workerID, task.ID, duration)
		sm.handleTaskSuccess(task, response, duration)
	}
}

// createDocument 创建文档
func (sm *SyncManager) createDocument(ctx context.Context, task *SyncTask) (*DocumentSyncResponse, error) {
	uploadResp, err := sm.client.UploadDocumentByText(ctx, sm.config.DefaultDatasetID, task.Title, task.Content, task.Metadata)
	if err != nil {
		return nil, fmt.Errorf("failed to upload document to Dify: %w", err)
	}

	response := &DocumentSyncResponse{
		DifyDocumentID: uploadResp.DocumentID,
		BatchID:        uploadResp.BatchID,
		Status:         "created",
		IndexingStatus: "processing",
	}

	// 更新数据库记录
	err = sm.db.Model(&models.DifyDocumentSync{}).Where("id = ?", task.ID).Updates(map[string]interface{}{
		"dify_document_id": uploadResp.DocumentID,
		"sync_status":      "synced",
		"indexing_status":  "processing",
		"updated_at":       time.Now(),
	}).Error

	return response, err
}

// updateDocument 更新文档
func (sm *SyncManager) updateDocument(ctx context.Context, task *SyncTask) (*DocumentSyncResponse, error) {
	// 获取现有的Dify文档ID
	var syncRecord models.DifyDocumentSync
	err := sm.db.Where("document_id = ?", task.DocumentID).First(&syncRecord).Error
	if err != nil {
		return nil, fmt.Errorf("failed to find existing sync record: %w", err)
	}

	err = sm.client.UpdateDocument(ctx, sm.config.DefaultDatasetID, syncRecord.DifyDocumentID, task.Title, task.Content, task.Metadata)
	if err != nil {
		return nil, fmt.Errorf("failed to update document in Dify: %w", err)
	}

	response := &DocumentSyncResponse{
		DifyDocumentID: syncRecord.DifyDocumentID,
		Status:         "updated",
		IndexingStatus: "processing",
	}

	// 更新数据库记录
	err = sm.db.Model(&syncRecord).Updates(map[string]interface{}{
		"title":           task.Title,
		"content":         task.Content,
		"sync_status":     "synced",
		"indexing_status": "processing",
		"updated_at":      time.Now(),
	}).Error

	return response, err
}

// deleteDocument 删除文档
func (sm *SyncManager) deleteDocument(ctx context.Context, task *SyncTask) error {
	var syncRecord models.DifyDocumentSync
	err := sm.db.Where("document_id = ?", task.DocumentID).First(&syncRecord).Error
	if err != nil {
		return fmt.Errorf("failed to find existing sync record: %w", err)
	}

	err = sm.client.DeleteDocument(ctx, sm.config.DefaultDatasetID, syncRecord.DifyDocumentID)
	if err != nil {
		return fmt.Errorf("failed to delete document from Dify: %w", err)
	}

	// 删除数据库记录
	err = sm.db.Delete(&syncRecord).Error
	return err
}

// handleTaskSuccess 处理任务成功
func (sm *SyncManager) handleTaskSuccess(task *SyncTask, response *DocumentSyncResponse, duration time.Duration) {
	sm.updateStats(func(stats *SyncStatistics) {
		stats.CompletedTasks++
		stats.PendingTasks--
		stats.LastSyncTime = time.Now()

		// 更新平均处理时间
		if stats.CompletedTasks == 1 {
			stats.AverageTime = duration.Seconds()
		} else {
			stats.AverageTime = (stats.AverageTime*float64(stats.CompletedTasks-1) + duration.Seconds()) / float64(stats.CompletedTasks)
		}

		// 更新成功率
		stats.SuccessRate = float64(stats.CompletedTasks) / float64(stats.TotalTasks) * 100
	})
}

// handleTaskFailure 处理任务失败
func (sm *SyncManager) handleTaskFailure(task *SyncTask, err error) {
	task.RetryCount++

	if task.RetryCount < task.MaxRetries {
		// 添加到重试队列
		select {
		case sm.retryQueue <- task:
			log.Printf("[SyncManager] Task %s added to retry queue (attempt %d/%d)",
				task.ID, task.RetryCount, task.MaxRetries)
		default:
			log.Printf("[SyncManager] Retry queue is full, dropping task %s", task.ID)
			sm.markTaskFailed(task, err)
		}
	} else {
		sm.markTaskFailed(task, err)
	}
}

// markTaskFailed 标记任务失败
func (sm *SyncManager) markTaskFailed(task *SyncTask, err error) {
	sm.db.Model(&models.DifyDocumentSync{}).Where("id = ?", task.ID).Updates(map[string]interface{}{
		"sync_status":   "failed",
		"error_message": err.Error(),
		"updated_at":    time.Now(),
	})

	sm.updateStats(func(stats *SyncStatistics) {
		stats.FailedTasks++
		stats.PendingTasks--
		stats.SuccessRate = float64(stats.CompletedTasks) / float64(stats.TotalTasks) * 100
	})

	log.Printf("[SyncManager] Task %s permanently failed: %v", task.ID, err)
}

// retryWorker 重试工作协程
func (sm *SyncManager) retryWorker(ctx context.Context) {
	defer sm.wg.Done()

	ticker := time.NewTicker(time.Minute * 5) // 每5分钟检查重试队列
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-sm.stopChan:
			return
		case <-ticker.C:
			sm.processRetryQueue(ctx)
		case task := <-sm.retryQueue:
			// 延迟重试
			select {
			case <-time.After(time.Duration(task.RetryCount) * time.Minute):
				select {
				case sm.syncQueue <- task:
					log.Printf("[SyncManager] Retrying task %s (attempt %d)", task.ID, task.RetryCount)
				default:
					log.Printf("[SyncManager] Sync queue full, re-queuing retry task %s", task.ID)
					sm.retryQueue <- task
				}
			case <-ctx.Done():
				return
			case <-sm.stopChan:
				return
			}
		}
	}
}

// processRetryQueue 处理重试队列
func (sm *SyncManager) processRetryQueue(ctx context.Context) {
	// 检查数据库中失败的任务，重新加入重试队列
	var failedRecords []models.DifyDocumentSync
	err := sm.db.Where("sync_status = ? AND updated_at < ?", "failed", time.Now().Add(-time.Hour)).
		Limit(10).Find(&failedRecords).Error

	if err != nil {
		log.Printf("[SyncManager] Failed to fetch failed sync records: %v", err)
		return
	}

	for _, record := range failedRecords {
		task := &SyncTask{
			ID:           record.ID,
			Type:         "create", // 默认重试创建
			DocumentID:   record.DocumentID,
			Title:        record.Title,
			Content:      record.Content,
			DocumentType: record.DocumentType,
			TeamID:       record.TeamID,
			UserID:       record.CreatedBy,
			RetryCount:   0,
			MaxRetries:   3,
		}

		select {
		case sm.retryQueue <- task:
			log.Printf("[SyncManager] Added failed task %s to retry queue", task.ID)
		default:
			log.Printf("[SyncManager] Retry queue full, skipping task %s", task.ID)
		}
	}
}

// autoSyncWorker 自动同步工作协程
func (sm *SyncManager) autoSyncWorker(ctx context.Context) {
	defer sm.wg.Done()

	ticker := time.NewTicker(sm.config.SyncInterval)
	defer ticker.Stop()

	log.Printf("[SyncManager] Auto sync enabled with interval: %v", sm.config.SyncInterval)

	for {
		select {
		case <-ctx.Done():
			return
		case <-sm.stopChan:
			return
		case <-ticker.C:
			sm.performAutoSync(ctx)
		}
	}
}

// performAutoSync 执行自动同步
func (sm *SyncManager) performAutoSync(ctx context.Context) {
	log.Printf("[SyncManager] Starting auto sync...")

	// 同步个人知识库文档
	if err := sm.syncPersonalKnowledgeBase(ctx); err != nil {
		log.Printf("[SyncManager] Failed to sync personal knowledge base: %v", err)
	}

	// 同步扫描文档
	if err := sm.syncScannedDocuments(ctx); err != nil {
		log.Printf("[SyncManager] Failed to sync scanned documents: %v", err)
	}

	log.Printf("[SyncManager] Auto sync completed")
}

// syncPersonalKnowledgeBase 同步个人知识库
func (sm *SyncManager) syncPersonalKnowledgeBase(ctx context.Context) error {
	var knowledgeList []models.PersonalKnowledgeBase
	err := sm.db.Where("is_shared = ? AND updated_at > ?", true, time.Now().Add(-time.Hour)).
		Limit(50).Find(&knowledgeList).Error

	if err != nil {
		return fmt.Errorf("failed to fetch personal knowledge: %w", err)
	}

	for _, knowledge := range knowledgeList {
		task := &SyncTask{
			Type:         "create",
			DocumentID:   knowledge.ID,
			Title:        knowledge.Title,
			Content:      knowledge.Content,
			DocumentType: "knowledge",
			TeamID:       knowledge.UserID, // 使用用户ID作为团队ID
			UserID:       knowledge.UserID,
			Metadata: map[string]interface{}{
				"source_type": knowledge.SourceType,
				"category":    knowledge.Category,
				"tags":        knowledge.Tags,
				"privacy":     knowledge.Privacy,
			},
			Priority:   3,
			MaxRetries: 3,
		}

		if err := sm.AddSyncTask(task); err != nil {
			log.Printf("[SyncManager] Failed to add knowledge sync task: %v", err)
		}
	}

	return nil
}

// syncScannedDocuments 同步扫描文档
func (sm *SyncManager) syncScannedDocuments(ctx context.Context) error {
	// 这里可以添加扫描文档的同步逻辑
	return nil
}

// statsUpdater 统计更新协程
func (sm *SyncManager) statsUpdater(ctx context.Context) {
	defer sm.wg.Done()

	ticker := time.NewTicker(time.Minute)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-sm.stopChan:
			return
		case <-ticker.C:
			sm.updateStatsFromDB()
		}
	}
}

// updateStatsFromDB 从数据库更新统计信息
func (sm *SyncManager) updateStatsFromDB() {
	var stats struct {
		Total   int64 `json:"total"`
		Synced  int64 `json:"synced"`
		Failed  int64 `json:"failed"`
		Pending int64 `json:"pending"`
	}

	sm.db.Model(&models.DifyDocumentSync{}).Select(
		"COUNT(*) as total",
		"COUNT(CASE WHEN sync_status = 'synced' THEN 1 END) as synced",
		"COUNT(CASE WHEN sync_status = 'failed' THEN 1 END) as failed",
		"COUNT(CASE WHEN sync_status = 'pending' THEN 1 END) as pending",
	).Scan(&stats)

	sm.updateStats(func(s *SyncStatistics) {
		s.TotalTasks = stats.Total
		s.CompletedTasks = stats.Synced
		s.FailedTasks = stats.Failed
		s.PendingTasks = stats.Pending
		s.LastUpdateTime = time.Now()

		if s.TotalTasks > 0 {
			s.SuccessRate = float64(s.CompletedTasks) / float64(s.TotalTasks) * 100
		}
	})
}

// updateStats 更新统计信息
func (sm *SyncManager) updateStats(fn func(*SyncStatistics)) {
	sm.mu.Lock()
	defer sm.mu.Unlock()
	fn(sm.syncStats)
}

// GetStats 获取统计信息
func (sm *SyncManager) GetStats() *SyncStatistics {
	sm.mu.RLock()
	defer sm.mu.RUnlock()

	statsCopy := *sm.syncStats
	return &statsCopy
}

// loadPendingTasks 加载待处理任务
func (sm *SyncManager) loadPendingTasks() error {
	var pendingRecords []models.DifyDocumentSync
	err := sm.db.Where("sync_status = ?", "pending").Limit(100).Find(&pendingRecords).Error

	if err != nil {
		return fmt.Errorf("failed to load pending tasks: %w", err)
	}

	for _, record := range pendingRecords {
		task := &SyncTask{
			ID:           record.ID,
			Type:         "create",
			DocumentID:   record.DocumentID,
			Title:        record.Title,
			Content:      record.Content,
			DocumentType: record.DocumentType,
			TeamID:       record.TeamID,
			UserID:       record.CreatedBy,
			MaxRetries:   3,
		}

		select {
		case sm.syncQueue <- task:
		default:
			log.Printf("[SyncManager] Queue full, skipping pending task %s", task.ID)
		}
	}

	log.Printf("[SyncManager] Loaded %d pending tasks", len(pendingRecords))
	return nil
}
