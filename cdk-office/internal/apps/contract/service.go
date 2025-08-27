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

package contract

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"github.com/linux-do/cdk-office/internal/models"
)

// Service 电子合同服务
type Service struct {
	db                *gorm.DB
	caProvider        CAProvider
	smsProvider       SMSProvider
	blockchainService BlockchainService
	config            *ContractConfig
}

// ContractConfig 合同服务配置
type ContractConfig struct {
	CAProvider         string `json:"ca_provider"`          // CA证书服务商
	SMSProvider        string `json:"sms_provider"`         // 短信服务商
	BlockchainEnabled  bool   `json:"blockchain_enabled"`   // 是否启用区块链存证
	DefaultExpireHours int    `json:"default_expire_hours"` // 默认过期时间(小时)
	MaxSigners         int    `json:"max_signers"`          // 最大签署人数
	AutoArchive        bool   `json:"auto_archive"`         // 是否自动归档
	FileStorageURL     string `json:"file_storage_url"`     // 文件存储服务URL
}

// CAProvider CA证书服务接口
type CAProvider interface {
	GenerateCertificate(ctx context.Context, userInfo *UserInfo) (*Certificate, error)
	ValidateCertificate(ctx context.Context, certID string) (*CertificateInfo, error)
	GenerateSignature(ctx context.Context, certID, content string) (*Signature, error)
}

// SMSProvider 短信服务接口
type SMSProvider interface {
	SendSignNotification(ctx context.Context, phone, contractTitle, signURL string) error
	SendCompletionNotification(ctx context.Context, phones []string, contractTitle string) error
	SendVerificationCode(ctx context.Context, phone string) (string, error)
}

// BlockchainService 区块链存证服务接口
type BlockchainService interface {
	StoreEvidence(ctx context.Context, data *EvidenceData) (*BlockchainEvidence, error)
	QueryEvidence(ctx context.Context, txHash string) (*EvidenceInfo, error)
	VerifyEvidence(ctx context.Context, txHash, originalHash string) (bool, error)
}

// UserInfo 用户信息
type UserInfo struct {
	Type        string `json:"type"` // person, company
	Name        string `json:"name"`
	IDCard      string `json:"id_card"`
	Phone       string `json:"phone"`
	Email       string `json:"email"`
	CompanyName string `json:"company_name"`
	UnifiedCode string `json:"unified_code"`
}

// Certificate CA证书
type Certificate struct {
	ID          string    `json:"id"`
	UserInfo    *UserInfo `json:"user_info"`
	PublicKey   string    `json:"public_key"`
	PrivateKey  string    `json:"private_key"`
	CertContent string    `json:"cert_content"`
	ValidFrom   time.Time `json:"valid_from"`
	ValidTo     time.Time `json:"valid_to"`
	IssuerCA    string    `json:"issuer_ca"`
}

// CertificateInfo CA证书信息
type CertificateInfo struct {
	ID       string    `json:"id"`
	Status   string    `json:"status"` // valid, invalid, expired
	UserInfo *UserInfo `json:"user_info"`
	ValidTo  time.Time `json:"valid_to"`
}

// Signature 数字签名
type Signature struct {
	CertID        string `json:"cert_id"`
	SignatureData string `json:"signature_data"`
	Algorithm     string `json:"algorithm"`
	Timestamp     int64  `json:"timestamp"`
}

// EvidenceData 存证数据
type EvidenceData struct {
	ContractID   string                 `json:"contract_id"`
	ContentHash  string                 `json:"content_hash"`
	SignerHashes []string               `json:"signer_hashes"`
	Timestamp    int64                  `json:"timestamp"`
	Metadata     map[string]interface{} `json:"metadata"`
}

// BlockchainEvidence 区块链存证结果
type BlockchainEvidence struct {
	TxHash      string    `json:"tx_hash"`
	BlockHeight int64     `json:"block_height"`
	Timestamp   time.Time `json:"timestamp"`
	EvidenceURL string    `json:"evidence_url"`
}

// EvidenceInfo 存证信息
type EvidenceInfo struct {
	TxHash      string                 `json:"tx_hash"`
	Status      string                 `json:"status"`
	ContentHash string                 `json:"content_hash"`
	Timestamp   time.Time              `json:"timestamp"`
	Metadata    map[string]interface{} `json:"metadata"`
}

// NewService 创建电子合同服务
func NewService(db *gorm.DB) *Service {
	config := &ContractConfig{
		CAProvider:         "internal",
		SMSProvider:        "aliyun",
		BlockchainEnabled:  true,
		DefaultExpireHours: 72,
		MaxSigners:         10,
		AutoArchive:        true,
		FileStorageURL:     "http://localhost:8000/files",
	}

	// 初始化服务提供者
	caProvider := NewInternalCAProvider()
	smsProvider := NewAliyunSMSProvider()
	blockchainService := NewBlockchainService()

	return &Service{
		db:                db,
		caProvider:        caProvider,
		smsProvider:       smsProvider,
		blockchainService: blockchainService,
		config:            config,
	}
}

// CreateContractRequest 创建合同请求
type CreateContractRequest struct {
	TemplateID        string                `json:"template_id"`
	Title             string                `json:"title"`
	Description       string                `json:"description"`
	Content           string                `json:"content"`
	SignMode          string                `json:"sign_mode"` // sequential, parallel
	RequireCA         bool                  `json:"require_ca"`
	RequireBlockchain bool                  `json:"require_blockchain"`
	ExpireTime        time.Time             `json:"expire_time"`
	Signers           []CreateSignerRequest `json:"signers"`
}

// CreateSignerRequest 创建签署人请求
type CreateSignerRequest struct {
	SignerType   string `json:"signer_type"` // person, company
	Name         string `json:"name"`
	Email        string `json:"email"`
	Phone        string `json:"phone"`
	IDCard       string `json:"id_card"`
	CompanyName  string `json:"company_name"`
	UnifiedCode  string `json:"unified_code"`
	SignOrder    int    `json:"sign_order"`
	SignPosition string `json:"sign_position"`
	SignType     string `json:"sign_type"` // signature, seal
}

// CreateContract 创建合同
func (s *Service) CreateContract(ctx context.Context, req *CreateContractRequest, userID, teamID string) (*models.Contract, error) {
	// 验证签署人数量
	if len(req.Signers) > s.config.MaxSigners {
		return nil, fmt.Errorf("签署人数量超过限制，最多%d人", s.config.MaxSigners)
	}

	// 开始数据库事务
	tx := s.db.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	// 创建合同记录
	contract := &models.Contract{
		ID:                uuid.New().String(),
		TeamID:            teamID,
		Title:             req.Title,
		Description:       req.Description,
		TemplateID:        req.TemplateID,
		Content:           req.Content,
		Status:            "draft",
		SignMode:          req.SignMode,
		RequireCA:         req.RequireCA,
		RequireBlockchain: req.RequireBlockchain,
		ExpireTime:        req.ExpireTime,
		CreatedBy:         userID,
		UpdatedBy:         userID,
	}

	// 设置默认过期时间
	if contract.ExpireTime.IsZero() {
		contract.ExpireTime = time.Now().Add(time.Duration(s.config.DefaultExpireHours) * time.Hour)
	}

	// 生成合同文件哈希
	contract.FileHash = s.generateContentHash(req.Content)

	if err := tx.Create(contract).Error; err != nil {
		tx.Rollback()
		return nil, fmt.Errorf("创建合同失败: %w", err)
	}

	// 创建签署人记录
	for _, signerReq := range req.Signers {
		signer := &models.ContractSigner{
			ID:           uuid.New().String(),
			ContractID:   contract.ID,
			SignerType:   signerReq.SignerType,
			Name:         signerReq.Name,
			Email:        signerReq.Email,
			Phone:        signerReq.Phone,
			IDCard:       signerReq.IDCard,
			CompanyName:  signerReq.CompanyName,
			UnifiedCode:  signerReq.UnifiedCode,
			SignOrder:    signerReq.SignOrder,
			SignPosition: signerReq.SignPosition,
			SignType:     signerReq.SignType,
			Status:       "pending",
		}

		if err := tx.Create(signer).Error; err != nil {
			tx.Rollback()
			return nil, fmt.Errorf("创建签署人失败: %w", err)
		}
	}

	// 记录操作日志
	if err := s.logContractAction(tx, contract.ID, "create", "创建合同", userID, ""); err != nil {
		log.Printf("记录操作日志失败: %v", err)
	}

	tx.Commit()

	log.Printf("合同创建成功: ID=%s, Title=%s", contract.ID, contract.Title)
	return contract, nil
}

// SendContract 发送合同供签署
func (s *Service) SendContract(ctx context.Context, contractID, userID string) error {
	// 获取合同信息
	var contract models.Contract
	if err := s.db.Preload("Signers").First(&contract, "id = ?", contractID).Error; err != nil {
		return fmt.Errorf("合同不存在: %w", err)
	}

	// 检查权限
	if contract.CreatedBy != userID {
		return fmt.Errorf("无权限操作此合同")
	}

	// 检查合同状态
	if contract.Status != "draft" {
		return fmt.Errorf("合同状态不允许发送")
	}

	// 更新合同状态
	contract.Status = "signing"
	contract.StartTime = time.Now()

	if err := s.db.Save(&contract).Error; err != nil {
		return fmt.Errorf("更新合同状态失败: %w", err)
	}

	// 根据签署模式发送通知
	if contract.SignMode == "sequential" {
		// 顺序签署：只通知第一个签署人
		return s.notifyNextSigner(ctx, &contract)
	} else {
		// 并行签署：通知所有签署人
		return s.notifyAllSigners(ctx, &contract)
	}
}

// SignContract 签署合同
func (s *Service) SignContract(ctx context.Context, req *SignContractRequest) error {
	// 获取签署人信息
	var signer models.ContractSigner
	if err := s.db.Preload("Contract").First(&signer, "id = ?", req.SignerID).Error; err != nil {
		return fmt.Errorf("签署人不存在: %w", err)
	}

	// 检查签署人状态
	if signer.Status != "pending" {
		return fmt.Errorf("签署人状态不允许签署")
	}

	// 检查合同是否过期
	if time.Now().After(signer.Contract.ExpireTime) {
		return fmt.Errorf("合同已过期")
	}

	// 开始数据库事务
	tx := s.db.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	// 生成CA证书（如果需要）
	var certificateID string
	if signer.Contract.RequireCA {
		userInfo := &UserInfo{
			Type:        signer.SignerType,
			Name:        signer.Name,
			IDCard:      signer.IDCard,
			Phone:       signer.Phone,
			Email:       signer.Email,
			CompanyName: signer.CompanyName,
			UnifiedCode: signer.UnifiedCode,
		}

		cert, err := s.caProvider.GenerateCertificate(ctx, userInfo)
		if err != nil {
			tx.Rollback()
			return fmt.Errorf("生成CA证书失败: %w", err)
		}
		certificateID = cert.ID
	}

	// 更新签署人状态
	now := time.Now()
	signer.Status = "signed"
	signer.SignTime = &now
	signer.SignIP = req.SignIP
	signer.SignLocation = req.SignLocation
	signer.CertificateID = certificateID
	signer.SignatureImage = req.SignatureImage

	if err := tx.Save(&signer).Error; err != nil {
		tx.Rollback()
		return fmt.Errorf("更新签署人状态失败: %w", err)
	}

	// 记录操作日志
	if err := s.logContractAction(tx, signer.ContractID, "sign",
		fmt.Sprintf("%s完成签署", signer.Name), "", req.SignIP); err != nil {
		log.Printf("记录操作日志失败: %v", err)
	}

	tx.Commit()

	// 检查是否所有人都已签署
	var pendingCount int64
	s.db.Model(&models.ContractSigner{}).Where("contract_id = ? AND status = ?",
		signer.ContractID, "pending").Count(&pendingCount)

	if pendingCount == 0 {
		// 所有人都已签署，完成合同
		return s.completeContract(ctx, signer.ContractID)
	} else if signer.Contract.SignMode == "sequential" {
		// 顺序签署：通知下一个签署人
		return s.notifyNextSigner(ctx, &signer.Contract)
	}

	log.Printf("合同签署成功: ContractID=%s, SignerID=%s", signer.ContractID, signer.ID)
	return nil
}

// SignContractRequest 签署合同请求
type SignContractRequest struct {
	SignerID       string `json:"signer_id"`
	SignatureImage string `json:"signature_image"`
	SignIP         string `json:"sign_ip"`
	SignLocation   string `json:"sign_location"`
}

// completeContract 完成合同签署
func (s *Service) completeContract(ctx context.Context, contractID string) error {
	// 获取合同信息
	var contract models.Contract
	if err := s.db.Preload("Signers").First(&contract, "id = ?", contractID).Error; err != nil {
		return fmt.Errorf("合同不存在: %w", err)
	}

	// 开始数据库事务
	tx := s.db.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	// 更新合同状态
	now := time.Now()
	contract.Status = "completed"
	contract.CompletedAt = &now

	// 生成最终合同文件哈希
	finalContent := s.generateFinalContractContent(&contract)
	contract.FileHash = s.generateContentHash(finalContent)

	// 区块链存证（如果需要）
	if contract.RequireBlockchain && s.config.BlockchainEnabled {
		evidenceData := &EvidenceData{
			ContractID:   contract.ID,
			ContentHash:  contract.FileHash,
			SignerHashes: s.getSignerHashes(contract.Signers),
			Timestamp:    now.Unix(),
			Metadata: map[string]interface{}{
				"title":        contract.Title,
				"signer_count": len(contract.Signers),
				"sign_mode":    contract.SignMode,
			},
		}

		evidence, err := s.blockchainService.StoreEvidence(ctx, evidenceData)
		if err != nil {
			log.Printf("区块链存证失败: %v", err)
		} else {
			contract.BlockchainTxHash = evidence.TxHash
			contract.EvidenceURL = evidence.EvidenceURL
		}
	}

	if err := tx.Save(&contract).Error; err != nil {
		tx.Rollback()
		return fmt.Errorf("更新合同状态失败: %w", err)
	}

	// 记录操作日志
	if err := s.logContractAction(tx, contract.ID, "complete", "合同签署完成", "", ""); err != nil {
		log.Printf("记录操作日志失败: %v", err)
	}

	tx.Commit()

	// 发送完成通知
	phones := make([]string, 0, len(contract.Signers))
	for _, signer := range contract.Signers {
		if signer.Phone != "" {
			phones = append(phones, signer.Phone)
		}
	}

	if len(phones) > 0 {
		if err := s.smsProvider.SendCompletionNotification(ctx, phones, contract.Title); err != nil {
			log.Printf("发送完成通知失败: %v", err)
		}
	}

	// 自动归档（如果配置启用）
	if s.config.AutoArchive {
		go s.archiveContract(context.Background(), contractID)
	}

	log.Printf("合同签署完成: ID=%s, Title=%s", contract.ID, contract.Title)
	return nil
}

// notifyNextSigner 通知下一个签署人
func (s *Service) notifyNextSigner(ctx context.Context, contract *models.Contract) error {
	// 查找下一个待签署的人
	var nextSigner models.ContractSigner
	err := s.db.Where("contract_id = ? AND status = ?", contract.ID, "pending").
		Order("sign_order ASC").First(&nextSigner).Error

	if err != nil {
		return fmt.Errorf("查找下一个签署人失败: %w", err)
	}

	// 发送短信通知
	signURL := fmt.Sprintf("%s/contract/sign/%s", s.config.FileStorageURL, nextSigner.ID)
	return s.smsProvider.SendSignNotification(ctx, nextSigner.Phone, contract.Title, signURL)
}

// notifyAllSigners 通知所有签署人
func (s *Service) notifyAllSigners(ctx context.Context, contract *models.Contract) error {
	var signers []models.ContractSigner
	if err := s.db.Where("contract_id = ? AND status = ?", contract.ID, "pending").Find(&signers).Error; err != nil {
		return fmt.Errorf("查找签署人失败: %w", err)
	}

	// 并发发送通知
	for _, signer := range signers {
		go func(s models.ContractSigner) {
			signURL := fmt.Sprintf("%s/contract/sign/%s", s.config.FileStorageURL, s.ID)
			if err := s.smsProvider.SendSignNotification(ctx, s.Phone, contract.Title, signURL); err != nil {
				log.Printf("发送签署通知失败: SignerID=%s, Error=%v", s.ID, err)
			}
		}(signer)
	}

	return nil
}

// CancelContract 取消合同
func (s *Service) CancelContract(ctx context.Context, contractID, userID, reason string) error {
	// 获取合同信息
	var contract models.Contract
	if err := s.db.First(&contract, "id = ?", contractID).Error; err != nil {
		return fmt.Errorf("合同不存在: %w", err)
	}

	// 检查权限
	if contract.CreatedBy != userID {
		return fmt.Errorf("无权限操作此合同")
	}

	// 检查合同状态
	if contract.Status == "completed" || contract.Status == "cancelled" {
		return fmt.Errorf("合同状态不允许取消")
	}

	// 更新合同状态
	contract.Status = "cancelled"
	contract.CompletedAt = &[]time.Time{time.Now()}[0]

	if err := s.db.Save(&contract).Error; err != nil {
		return fmt.Errorf("更新合同状态失败: %w", err)
	}

	// 记录操作日志
	if err := s.logContractAction(s.db, contract.ID, "cancel", "取消合同: "+reason, userID, ""); err != nil {
		log.Printf("记录操作日志失败: %v", err)
	}

	log.Printf("合同已取消: ID=%s, Reason=%s", contract.ID, reason)
	return nil
}

// GetContractDetail 获取合同详情
func (s *Service) GetContractDetail(ctx context.Context, contractID, userID string) (*ContractDetail, error) {
	var contract models.Contract
	if err := s.db.Preload("Signers").Preload("Logs").First(&contract, "id = ?", contractID).Error; err != nil {
		return nil, fmt.Errorf("合同不存在: %w", err)
	}

	// 检查权限（创建者或签署人可以查看）
	hasPermission := contract.CreatedBy == userID
	if !hasPermission {
		for _, signer := range contract.Signers {
			if signer.Email == userID || signer.Phone == userID {
				hasPermission = true
				break
			}
		}
	}

	if !hasPermission {
		return nil, fmt.Errorf("无权限查看此合同")
	}

	return &ContractDetail{
		Contract: &contract,
		Progress: s.calculateProgress(&contract),
	}, nil
}

// ContractDetail 合同详情
type ContractDetail struct {
	Contract *models.Contract `json:"contract"`
	Progress int              `json:"progress"`
}

// calculateProgress 计算签署进度
func (s *Service) calculateProgress(contract *models.Contract) int {
	if len(contract.Signers) == 0 {
		return 0
	}

	signedCount := 0
	for _, signer := range contract.Signers {
		if signer.Status == "signed" {
			signedCount++
		}
	}

	return (signedCount * 100) / len(contract.Signers)
}

// 辅助方法

// generateContentHash 生成内容哈希
func (s *Service) generateContentHash(content string) string {
	hash := sha256.Sum256([]byte(content))
	return hex.EncodeToString(hash[:])
}

// generateFinalContractContent 生成最终合同内容
func (s *Service) generateFinalContractContent(contract *models.Contract) string {
	// 这里应该生成包含所有签名的最终合同内容
	// 简化处理，返回原内容加上签署信息
	var builder strings.Builder
	builder.WriteString(contract.Content)
	builder.WriteString("\n\n=== 签署信息 ===\n")

	for _, signer := range contract.Signers {
		if signer.Status == "signed" {
			builder.WriteString(fmt.Sprintf("签署人: %s, 签署时间: %v\n",
				signer.Name, signer.SignTime))
		}
	}

	return builder.String()
}

// getSignerHashes 获取签署人哈希列表
func (s *Service) getSignerHashes(signers []models.ContractSigner) []string {
	hashes := make([]string, len(signers))
	for i, signer := range signers {
		signerData := fmt.Sprintf("%s:%s:%s", signer.Name, signer.Phone, signer.IDCard)
		hash := sha256.Sum256([]byte(signerData))
		hashes[i] = hex.EncodeToString(hash[:])
	}
	return hashes
}

// logContractAction 记录合同操作日志
func (s *Service) logContractAction(tx *gorm.DB, contractID, action, description, operatorID, ipAddress string) error {
	log := &models.ContractLog{
		ID:          uuid.New().String(),
		ContractID:  contractID,
		Action:      action,
		Description: description,
		OperatorID:  operatorID,
		IPAddress:   ipAddress,
	}

	return tx.Create(log).Error
}

// archiveContract 归档合同
func (s *Service) archiveContract(ctx context.Context, contractID string) {
	// 这里实现合同归档逻辑
	// 例如：生成PDF、上传到存储服务、更新状态等
	log.Printf("开始归档合同: %s", contractID)

	// 简化实现
	time.Sleep(5 * time.Second)

	s.db.Model(&models.Contract{}).Where("id = ?", contractID).Update("status", "archived")
	log.Printf("合同归档完成: %s", contractID)
}
