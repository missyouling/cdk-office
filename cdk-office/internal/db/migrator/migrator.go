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
	"github.com/linux-do/cdk-office/internal/models"
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
		// 原有模型
		&models.Document{},
		&models.DocumentVersion{},
		&models.DocumentTag{},
		&models.DocumentTagRelation{},
		&models.User{},
		&models.Approval{},
		&models.Notification{},
		&models.QRCode{},
		&models.QRCodeForm{},
		&models.QRCodeFormField{},
		&models.QRCodeRecord{},
		&models.Archive{},
		&models.ArchiveRule{},
		&models.ArchiveLog{},
		&models.AIService{},
		&models.AIServiceConfig{},
		&models.OCRTask{},
		&models.OCRResult{},
		
		// 新增合同相关模型
		&models.Contract{},
		&models.ContractSigner{},
		&models.ContractTemplate{},
		&models.ContractLog{},
		&models.ContractFile{},
		&models.KnowledgeSubmission{},
		&models.ContractServiceConfig{},
		&models.ContractWorkflow{},
		&models.ContractStatistics{},
		&models.ContractNotification{},
		
		// 新增问卷调查相关模型
		&models.Survey{},
		&models.SurveyResponse{},
		&models.SurveyAnalysis{},
		&models.SurveyPermission{},
		&models.SurveyTemplate{},
		&models.SurveyFile{},
	); err != nil {
		log.Fatalf("[MIGRATOR] failed to migrate database: %v", err)
	}
	
	log.Println("[MIGRATOR] database migration completed")
}