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

package worker

import (
	"context"
	"log"
	"sync"
	"time"

	"github.com/linux-do/cdk-office/internal/config"
	"github.com/linux-do/cdk-office/internal/db"
	"github.com/linux-do/cdk-office/internal/logger"
)

// Worker 工作进程
type Worker struct {
	ctx    context.Context
	cancel context.CancelFunc
	wg     sync.WaitGroup
}

// NewWorker 创建新的工作进程
func NewWorker() *Worker {
	ctx, cancel := context.WithCancel(context.Background())
	return &Worker{
		ctx:    ctx,
		cancel: cancel,
	}
}

// Start 启动工作进程
func (w *Worker) Start() {
	// 初始化数据库连接
	db.Init()
	db.InitRedis()

	// 启动工作进程
	for i := 0; i < config.Config.Worker.Concurrency; i++ {
		w.wg.Add(1)
		go w.workerRoutine(i)
	}

	log.Printf("[WORKER] started %d worker routines", config.Config.Worker.Concurrency)
}

// Stop 停止工作进程
func (w *Worker) Stop() {
	log.Println("[WORKER] stopping worker...")
	w.cancel()
	w.wg.Wait()
	log.Println("[WORKER] worker stopped")
}

// workerRoutine 工作进程例程
func (w *Worker) workerRoutine(id int) {
	defer w.wg.Done()

	log.Printf("[WORKER] worker routine %d started", id)

	for {
		select {
		case <-w.ctx.Done():
			log.Printf("[WORKER] worker routine %d stopped", id)
			return
		default:
			// 处理任务
			if err := w.processTask(); err != nil {
				logger.Error("worker routine %d failed to process task: %v", id, err)
				// 等待一段时间后重试
				time.Sleep(1 * time.Second)
			}
		}
	}
}

// processTask 处理任务
func (w *Worker) processTask() error {
	// TODO: 实现任务处理逻辑
	// 这里可以从队列中获取任务并处理

	// 模拟任务处理
	time.Sleep(100 * time.Millisecond)

	return nil
}
