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

package migrator

import (
	"log"

	"github.com/linux-do/cdk-office/internal/db"
	"gorm.io/gorm"
)

// Migrate 执行数据库迁移
func Migrate() {
	// 初始化数据库连接
	db.Init()
	
	// 获取数据库实例
	database := db.GetDB()
	
	// 执行自动迁移
	if err := database.AutoMigrate(
		// 在这里添加需要迁移的模型
		// 例如：&models.User{}, &models.Project{},
	); err != nil {
		log.Fatalf("[MIGRATOR] failed to migrate database: %v", err)
	}
	
	log.Println("[MIGRATOR] database migration completed")
}