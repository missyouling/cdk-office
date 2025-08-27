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
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/linux-do/cdk-office/internal/db"
	"gorm.io/gorm"
)

// Handler 电子合同处理器
type Handler struct {
	db      *gorm.DB
	service *Service
}

// NewHandler 创建电子合同处理器
func NewHandler() *Handler {
	database := db.GetDB()

	// 从配置中读取电子合同配置
	config := &ContractConfig{
		Enabled:          true,
		MaxFileSize:      100 * 1024 * 1024, // 100MB
		AllowedFormats:   []string{"pdf", "doc", "docx"},
		EnableBlockchain: true,
		EnableCAService:  true,
		EnableSMSService: true,
		SignatureTimeout: 7 * 24 * 3600, // 7天
		MaxSignatories:   10,
	}

	service := NewService(config, database)

	return &Handler{
		db:      database,
		service: service,
	}
}

// RegisterRoutes 注册路由
func (h *Handler) RegisterRoutes(router *gin.RouterGroup) {
	contract := router.Group("/contract")
	{
		// 合同管理
		contract.POST("/create", h.CreateContract)
		contract.GET("/list", h.ListContracts)
		contract.GET("/:id", h.GetContract)
		contract.PUT("/:id", h.UpdateContract)
		contract.DELETE("/:id", h.DeleteContract)

		// 签署管理
		contract.POST("/:id/sign", h.SignContract)
		contract.POST("/:id/add-signatory", h.AddSignatory)
		contract.DELETE("/:id/signatory/:signatoryId", h.RemoveSignatory)
		contract.GET("/:id/signatures", h.GetSignatures)

		// CA证书管理
		contract.POST("/ca/generate", h.GenerateCAKey)
		contract.GET("/ca/certificate/:userId", h.GetCACertificate)
		contract.POST("/ca/revoke/:certificateId", h.RevokeCertificate)

		// 短信服务
		contract.POST("/sms/send-code", h.SendVerificationCode)
		contract.POST("/sms/verify-code", h.VerifyCode)

		// 区块链存证
		contract.POST("/:id/blockchain/store", h.StoreOnBlockchain)
		contract.GET("/:id/blockchain/verify", h.VerifyBlockchainRecord)

		// 合同模板
		contract.POST("/template", h.CreateTemplate)
		contract.GET("/templates", h.ListTemplates)
		contract.GET("/template/:id", h.GetTemplate)

		// 文件管理
		contract.POST("/upload", h.UploadFile)
		contract.GET("/download/:fileId", h.DownloadFile)

		// 状态和统计
		contract.GET("/status/:id", h.GetContractStatus)
		contract.GET("/statistics", h.GetStatistics)
		contract.GET("/health", h.HealthCheck)
	}
}

// CreateContract 创建合同
// @Summary 创建电子合同
// @Description 创建新的电子合同，支持文件上传和基本信息设置
// @Tags contract
// @Accept multipart/form-data
// @Produce json
// @Param title formData string true "合同标题"
// @Param description formData string false "合同描述"
// @Param template_id formData string false "模板ID"
// @Param file formData file false "合同文件"
// @Success 200 {object} ContractInfo
// @Failure 400 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /api/contract/create [post]
func (h *Handler) CreateContract(c *gin.Context) {
	userID := c.GetString("user_id")
	teamID := c.GetString("team_id")

	// 解析表单数据
	title := c.PostForm("title")
	description := c.PostForm("description")
	templateID := c.PostForm("template_id")

	if title == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "合同标题不能为空"})
		return
	}

	// 处理文件上传
	var fileURL string
	file, err := c.FormFile("file")
	if err == nil {
		// 验证文件格式和大小
		if !h.service.isValidFileFormat(file.Filename) {
			c.JSON(http.StatusBadRequest, gin.H{"error": "不支持的文件格式"})
			return
		}

		if file.Size > int64(h.service.config.MaxFileSize) {
			c.JSON(http.StatusBadRequest, gin.H{"error": "文件大小超过限制"})
			return
		}

		// 保存文件
		fileURL, err = h.service.saveContractFile(file)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "文件保存失败"})
			return
		}
	}

	// 创建合同请求
	req := &CreateContractRequest{
		Title:       title,
		Description: description,
		TemplateID:  templateID,
		FileURL:     fileURL,
		CreatorID:   userID,
		TeamID:      teamID,
	}

	// 创建合同
	contract, err := h.service.CreateContract(c.Request.Context(), req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, contract)
}

// ListContracts 获取合同列表
// @Summary 获取合同列表
// @Description 获取用户或团队的合同列表，支持分页和筛选
// @Tags contract
// @Produce json
// @Param page query int false "页码" default(1)
// @Param limit query int false "每页数量" default(10)
// @Param status query string false "合同状态"
// @Param team_id query string false "团队ID"
// @Success 200 {object} map[string]interface{}
// @Router /api/contract/list [get]
func (h *Handler) ListContracts(c *gin.Context) {
	userID := c.GetString("user_id")
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))
	status := c.Query("status")
	teamID := c.Query("team_id")

	req := &ListContractsRequest{
		UserID: userID,
		TeamID: teamID,
		Status: status,
		Page:   page,
		Limit:  limit,
	}

	contracts, total, err := h.service.ListContracts(c.Request.Context(), req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"contracts": contracts,
		"total":     total,
		"page":      page,
		"limit":     limit,
	})
}

// GetContract 获取合同详情
// @Summary 获取合同详情
// @Description 根据合同ID获取详细信息
// @Tags contract
// @Produce json
// @Param id path string true "合同ID"
// @Success 200 {object} ContractInfo
// @Failure 404 {object} map[string]interface{}
// @Router /api/contract/{id} [get]
func (h *Handler) GetContract(c *gin.Context) {
	contractID := c.Param("id")
	userID := c.GetString("user_id")

	contract, err := h.service.GetContract(c.Request.Context(), contractID, userID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "合同不存在"})
		return
	}

	c.JSON(http.StatusOK, contract)
}

// UpdateContract 更新合同
func (h *Handler) UpdateContract(c *gin.Context) {
	contractID := c.Param("id")
	userID := c.GetString("user_id")

	var req UpdateContractRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "参数错误"})
		return
	}

	req.ContractID = contractID
	req.UserID = userID

	err := h.service.UpdateContract(c.Request.Context(), &req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "合同更新成功"})
}

// DeleteContract 删除合同
func (h *Handler) DeleteContract(c *gin.Context) {
	contractID := c.Param("id")
	userID := c.GetString("user_id")

	err := h.service.DeleteContract(c.Request.Context(), contractID, userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "合同删除成功"})
}

// SignContract 签署合同
// @Summary 签署合同
// @Description 对指定合同进行数字签名
// @Tags contract
// @Accept json
// @Produce json
// @Param id path string true "合同ID"
// @Param request body SignContractRequest true "签署请求"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /api/contract/{id}/sign [post]
func (h *Handler) SignContract(c *gin.Context) {
	contractID := c.Param("id")
	userID := c.GetString("user_id")

	var req SignContractRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "参数错误"})
		return
	}

	req.ContractID = contractID
	req.SignerID = userID

	signature, err := h.service.SignContract(c.Request.Context(), &req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message":   "签署成功",
		"signature": signature,
	})
}

// AddSignatory 添加签署人
func (h *Handler) AddSignatory(c *gin.Context) {
	contractID := c.Param("id")
	userID := c.GetString("user_id")

	var req AddSignatoryRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "参数错误"})
		return
	}

	req.ContractID = contractID
	req.RequesterID = userID

	err := h.service.AddSignatory(c.Request.Context(), &req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "签署人添加成功"})
}

// RemoveSignatory 移除签署人
func (h *Handler) RemoveSignatory(c *gin.Context) {
	contractID := c.Param("id")
	signatoryID := c.Param("signatoryId")
	userID := c.GetString("user_id")

	err := h.service.RemoveSignatory(c.Request.Context(), contractID, signatoryID, userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "签署人移除成功"})
}

// GetSignatures 获取签名列表
func (h *Handler) GetSignatures(c *gin.Context) {
	contractID := c.Param("id")
	userID := c.GetString("user_id")

	signatures, err := h.service.GetSignatures(c.Request.Context(), contractID, userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"signatures": signatures})
}

// GenerateCAKey 生成CA密钥
func (h *Handler) GenerateCAKey(c *gin.Context) {
	userID := c.GetString("user_id")

	var req GenerateCAKeyRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "参数错误"})
		return
	}

	req.UserID = userID

	keyPair, err := h.service.GenerateCAKey(c.Request.Context(), &req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "CA密钥生成成功",
		"key_id":  keyPair.ID,
	})
}

// GetCACertificate 获取CA证书
func (h *Handler) GetCACertificate(c *gin.Context) {
	userID := c.Param("userId")
	requesterID := c.GetString("user_id")

	certificate, err := h.service.GetCACertificate(c.Request.Context(), userID, requesterID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "证书不存在"})
		return
	}

	c.JSON(http.StatusOK, certificate)
}

// RevokeCertificate 撤销证书
func (h *Handler) RevokeCertificate(c *gin.Context) {
	certificateID := c.Param("certificateId")
	userID := c.GetString("user_id")

	err := h.service.RevokeCertificate(c.Request.Context(), certificateID, userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "证书撤销成功"})
}

// SendVerificationCode 发送验证码
func (h *Handler) SendVerificationCode(c *gin.Context) {
	var req SendVerificationCodeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "参数错误"})
		return
	}

	err := h.service.SendVerificationCode(c.Request.Context(), &req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "验证码发送成功"})
}

// VerifyCode 验证验证码
func (h *Handler) VerifyCode(c *gin.Context) {
	var req VerifyCodeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "参数错误"})
		return
	}

	valid, err := h.service.VerifyCode(c.Request.Context(), &req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	if !valid {
		c.JSON(http.StatusBadRequest, gin.H{"error": "验证码错误"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "验证成功"})
}

// StoreOnBlockchain 区块链存证
func (h *Handler) StoreOnBlockchain(c *gin.Context) {
	contractID := c.Param("id")
	userID := c.GetString("user_id")

	record, err := h.service.StoreOnBlockchain(c.Request.Context(), contractID, userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "区块链存证成功",
		"record":  record,
	})
}

// VerifyBlockchainRecord 验证区块链记录
func (h *Handler) VerifyBlockchainRecord(c *gin.Context) {
	contractID := c.Param("id")

	valid, record, err := h.service.VerifyBlockchainRecord(c.Request.Context(), contractID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"valid":  valid,
		"record": record,
	})
}

// CreateTemplate 创建合同模板
func (h *Handler) CreateTemplate(c *gin.Context) {
	userID := c.GetString("user_id")

	var req CreateTemplateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "参数错误"})
		return
	}

	req.CreatorID = userID

	template, err := h.service.CreateTemplate(c.Request.Context(), &req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, template)
}

// ListTemplates 获取模板列表
func (h *Handler) ListTemplates(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))

	templates, total, err := h.service.ListTemplates(c.Request.Context(), page, limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"templates": templates,
		"total":     total,
		"page":      page,
		"limit":     limit,
	})
}

// GetTemplate 获取模板详情
func (h *Handler) GetTemplate(c *gin.Context) {
	templateID := c.Param("id")

	template, err := h.service.GetTemplate(c.Request.Context(), templateID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "模板不存在"})
		return
	}

	c.JSON(http.StatusOK, template)
}

// UploadFile 上传文件
func (h *Handler) UploadFile(c *gin.Context) {
	file, err := c.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "文件上传失败"})
		return
	}

	fileURL, err := h.service.saveContractFile(file)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "文件保存失败"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message":  "文件上传成功",
		"file_url": fileURL,
	})
}

// DownloadFile 下载文件
func (h *Handler) DownloadFile(c *gin.Context) {
	fileID := c.Param("fileId")
	userID := c.GetString("user_id")

	fileURL, filename, err := h.service.GetFileDownloadURL(c.Request.Context(), fileID, userID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "文件不存在"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"download_url": fileURL,
		"filename":     filename,
	})
}

// GetContractStatus 获取合同状态
func (h *Handler) GetContractStatus(c *gin.Context) {
	contractID := c.Param("id")
	userID := c.GetString("user_id")

	status, err := h.service.GetContractStatus(c.Request.Context(), contractID, userID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "合同不存在"})
		return
	}

	c.JSON(http.StatusOK, status)
}

// GetStatistics 获取统计信息
func (h *Handler) GetStatistics(c *gin.Context) {
	userID := c.GetString("user_id")
	teamID := c.GetString("team_id")

	statistics, err := h.service.GetStatistics(c.Request.Context(), userID, teamID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, statistics)
}

// HealthCheck 健康检查
func (h *Handler) HealthCheck(c *gin.Context) {
	err := h.service.HealthCheck()
	if err != nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{
			"status": "unhealthy",
			"error":  err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status":    "healthy",
		"timestamp": gin.H{"checked_at": gin.H{}},
	})
}
