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

package qrcode

import (
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/linux-do/cdk-office/internal/models"
	"gorm.io/gorm"
)

// Router 二维码路由
type Router struct {
	service *QRCodeService
	db      *gorm.DB
}

// NewRouter 创建二维码路由
func NewRouter(db *gorm.DB, storagePath, baseURL string) *Router {
	service := NewQRCodeService(db, storagePath, baseURL)

	return &Router{
		service: service,
		db:      db,
	}
}

// RegisterRoutes 注册路由
func (r *Router) RegisterRoutes(rg *gin.RouterGroup) {
	qrcode := rg.Group("/qrcode")
	{
		// 表单管理（需要认证）
		qrcode.POST("/forms", r.CreateForm)
		qrcode.GET("/forms", r.ListForms)
		qrcode.GET("/forms/:id", r.GetForm)
		qrcode.PUT("/forms/:id", r.UpdateForm)
		qrcode.DELETE("/forms/:id", r.DeleteForm)

		// 二维码生成
		qrcode.POST("/generate", r.GenerateQRCode)
		qrcode.GET("/records", r.GetQRCodeRecords)

		// 数据管理
		qrcode.GET("/submissions/:formId", r.GetSubmissions)
		qrcode.GET("/analytics/:formId", r.GetFormAnalytics)

		// 模板管理
		qrcode.GET("/templates", r.GetFormTemplates)
		qrcode.POST("/forms/:id/clone", r.CloneForm)
	}

	// 公开访问路由（表单提交，无需认证）
	public := rg.Group("/public")
	{
		public.GET("/forms/:id", r.GetPublicForm)
		public.POST("/forms/:id/submit", r.SubmitForm)
		public.GET("/qrcodes/:filename", r.ServeQRCode)
	}
}

// CreateForm 创建表单
// @Summary 创建二维码表单
// @Description 创建新的二维码表单，包含字段定义
// @Tags QRCode
// @Accept json
// @Produce json
// @Param request body FormRequest true "表单创建请求"
// @Success 201 {object} models.QRCodeForm
// @Failure 400 {object} ErrorResponse
// @Failure 401 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /api/qrcode/forms [post]
func (r *Router) CreateForm(c *gin.Context) {
	userID := c.GetHeader("X-User-ID")
	teamID := c.GetHeader("X-Team-ID")

	if userID == "" || teamID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	var req FormRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	form, err := r.service.CreateForm(userID, teamID, &req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, form)
}

// ListForms 获取表单列表
// @Summary 获取表单列表
// @Description 获取当前团队的表单列表
// @Tags QRCode
// @Accept json
// @Produce json
// @Param page query int false "页码" default(1)
// @Param limit query int false "每页数量" default(20)
// @Success 200 {object} FormListResponse
// @Router /api/qrcode/forms [get]
func (r *Router) ListForms(c *gin.Context) {
	teamID := c.GetHeader("X-Team-ID")
	if teamID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Team not specified"})
		return
	}

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))

	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 20
	}

	result, err := r.service.ListForms(teamID, page, limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, result)
}

// GetForm 获取表单详情
// @Summary 获取表单详情
// @Description 获取指定表单的详细信息，包括字段定义
// @Tags QRCode
// @Accept json
// @Produce json
// @Param id path string true "表单ID"
// @Success 200 {object} FormDetails
// @Router /api/qrcode/forms/{id} [get]
func (r *Router) GetForm(c *gin.Context) {
	formID := c.Param("id")

	details, err := r.service.GetForm(formID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, details)
}

// UpdateForm 更新表单
// @Summary 更新表单
// @Description 更新表单信息和字段定义
// @Tags QRCode
// @Accept json
// @Produce json
// @Param id path string true "表单ID"
// @Param request body FormRequest true "表单更新请求"
// @Success 200 {object} SuccessResponse
// @Router /api/qrcode/forms/{id} [put]
func (r *Router) UpdateForm(c *gin.Context) {
	formID := c.Param("id")
	userID := c.GetHeader("X-User-ID")

	if userID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	var req FormRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := r.service.UpdateForm(formID, userID, &req); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Form updated successfully"})
}

// DeleteForm 删除表单
// @Summary 删除表单
// @Description 删除表单及其相关数据
// @Tags QRCode
// @Accept json
// @Produce json
// @Param id path string true "表单ID"
// @Success 200 {object} SuccessResponse
// @Router /api/qrcode/forms/{id} [delete]
func (r *Router) DeleteForm(c *gin.Context) {
	formID := c.Param("id")
	userID := c.GetHeader("X-User-ID")

	if userID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	if err := r.service.DeleteForm(formID, userID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Form deleted successfully"})
}

// GenerateQRCode 生成二维码
// @Summary 生成二维码
// @Description 为指定表单生成二维码
// @Tags QRCode
// @Accept json
// @Produce json
// @Param request body QRCodeRequest true "二维码生成请求"
// @Success 200 {object} models.QRCodeRecord
// @Router /api/qrcode/generate [post]
func (r *Router) GenerateQRCode(c *gin.Context) {
	userID := c.GetHeader("X-User-ID")
	if userID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	var req QRCodeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	record, err := r.service.GenerateQRCode(userID, &req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, record)
}

// GetQRCodeRecords 获取二维码记录
// @Summary 获取二维码记录列表
// @Description 获取当前团队的二维码生成记录
// @Tags QRCode
// @Accept json
// @Produce json
// @Param page query int false "页码" default(1)
// @Param limit query int false "每页数量" default(20)
// @Success 200 {object} QRCodeRecordListResponse
// @Router /api/qrcode/records [get]
func (r *Router) GetQRCodeRecords(c *gin.Context) {
	teamID := c.GetHeader("X-Team-ID")
	if teamID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Team not specified"})
		return
	}

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))

	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 20
	}

	result, err := r.service.GetQRCodeRecords(teamID, page, limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, result)
}

// GetSubmissions 获取表单提交记录
// @Summary 获取表单提交记录
// @Description 获取指定表单的提交记录
// @Tags QRCode
// @Accept json
// @Produce json
// @Param formId path string true "表单ID"
// @Param page query int false "页码" default(1)
// @Param limit query int false "每页数量" default(20)
// @Success 200 {object} SubmissionListResponse
// @Router /api/qrcode/submissions/{formId} [get]
func (r *Router) GetSubmissions(c *gin.Context) {
	formID := c.Param("formId")

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))

	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 20
	}

	result, err := r.service.GetSubmissions(formID, page, limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, result)
}

// GetFormAnalytics 获取表单分析数据
// @Summary 获取表单分析数据
// @Description 获取表单的统计分析数据
// @Tags QRCode
// @Accept json
// @Produce json
// @Param formId path string true "表单ID"
// @Success 200 {object} FormAnalytics
// @Router /api/qrcode/analytics/{formId} [get]
func (r *Router) GetFormAnalytics(c *gin.Context) {
	formID := c.Param("formId")

	analytics, err := r.getFormAnalytics(formID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, analytics)
}

// getFormAnalytics 获取表单分析数据的具体实现
func (r *Router) getFormAnalytics(formID string) (*FormAnalytics, error) {
	// 获取表单信息
	details, err := r.service.GetForm(formID)
	if err != nil {
		return nil, err
	}

	// 统计提交数量
	var totalSubmissions int64
	if err := r.db.Model(&models.QRCodeFormSubmission{}).Where("form_id = ?", formID).Count(&totalSubmissions).Error; err != nil {
		return nil, err
	}

	// 统计今日提交数量
	var todaySubmissions int64
	if err := r.db.Model(&models.QRCodeFormSubmission{}).
		Where("form_id = ? AND DATE(created_at) = CURRENT_DATE", formID).
		Count(&todaySubmissions).Error; err != nil {
		return nil, err
	}

	// 统计二维码生成数量
	var qrCodeCount int64
	if err := r.db.Model(&models.QRCodeRecord{}).Where("form_id = ?", formID).Count(&qrCodeCount).Error; err != nil {
		return nil, err
	}

	analytics := &FormAnalytics{
		FormID:           formID,
		FormName:         details.Form.FormName,
		TotalSubmissions: int(totalSubmissions),
		TodaySubmissions: int(todaySubmissions),
		QRCodeCount:      int(qrCodeCount),
		CreatedAt:        details.Form.CreatedAt,
	}

	return analytics, nil
}

// GetFormTemplates 获取表单模板
// @Summary 获取表单模板
// @Description 获取预定义的表单模板
// @Tags QRCode
// @Accept json
// @Produce json
// @Success 200 {object} []FormTemplate
// @Router /api/qrcode/templates [get]
func (r *Router) GetFormTemplates(c *gin.Context) {
	templates := GetPresetFormTemplates()
	c.JSON(http.StatusOK, templates)
}

// CloneForm 克隆表单
// @Summary 克隆表单
// @Description 基于现有表单创建副本
// @Tags QRCode
// @Accept json
// @Produce json
// @Param id path string true "源表单ID"
// @Param request body CloneFormRequest true "克隆请求"
// @Success 201 {object} models.QRCodeForm
// @Router /api/qrcode/forms/{id}/clone [post]
func (r *Router) CloneForm(c *gin.Context) {
	sourceFormID := c.Param("id")
	userID := c.GetHeader("X-User-ID")
	teamID := c.GetHeader("X-Team-ID")

	if userID == "" || teamID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	var req CloneFormRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// 获取源表单
	sourceDetails, err := r.service.GetForm(sourceFormID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Source form not found"})
		return
	}

	// 构建新表单请求
	newFormReq := &FormRequest{
		FormName:    req.NewFormName,
		FormType:    sourceDetails.Form.FormType,
		Description: sourceDetails.Form.Description,
		FormFields:  make([]FieldInfo, len(sourceDetails.Fields)),
	}

	// 复制字段定义
	for i, field := range sourceDetails.Fields {
		newFormReq.FormFields[i] = FieldInfo{
			FieldKey:     field.FieldKey,
			FieldLabel:   field.FieldLabel,
			FieldType:    field.FieldType,
			IsRequired:   field.IsRequired,
			DefaultValue: field.DefaultValue,
			Options:      field.Options,
			DisplayOrder: field.DisplayOrder,
		}
	}

	// 创建新表单
	newForm, err := r.service.CreateForm(userID, teamID, newFormReq)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, newForm)
}

// 公开访问接口

// GetPublicForm 获取公开表单（无需认证）
// @Summary 获取公开表单
// @Description 获取用于填写的公开表单
// @Tags Public
// @Accept json
// @Produce json
// @Param id path string true "表单ID"
// @Success 200 {object} PublicFormData
// @Router /api/public/forms/{id} [get]
func (r *Router) GetPublicForm(c *gin.Context) {
	formID := c.Param("id")

	details, err := r.service.GetForm(formID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Form not found"})
		return
	}

	// 构建公开表单数据（不包含敏感信息）
	publicForm := &PublicFormData{
		ID:          details.Form.ID,
		FormName:    details.Form.FormName,
		FormType:    details.Form.FormType,
		Description: details.Form.Description,
		Fields:      details.Fields,
	}

	c.JSON(http.StatusOK, publicForm)
}

// SubmitForm 提交表单（无需认证）
// @Summary 提交表单
// @Description 提交表单数据
// @Tags Public
// @Accept json
// @Produce json
// @Param id path string true "表单ID"
// @Param request body SubmissionRequest true "提交数据"
// @Success 200 {object} SuccessResponse
// @Router /api/public/forms/{id}/submit [post]
func (r *Router) SubmitForm(c *gin.Context) {
	formID := c.Param("id")
	clientIP := c.ClientIP()

	var req SubmissionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := r.service.SubmitForm(formID, clientIP, &req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Form submitted successfully"})
}

// ServeQRCode 提供二维码图片文件
// @Summary 提供二维码图片
// @Description 返回二维码图片文件
// @Tags Public
// @Param filename path string true "文件名"
// @Success 200 {file} image/jpeg
// @Router /api/public/qrcodes/{filename} [get]
func (r *Router) ServeQRCode(c *gin.Context) {
	filename := c.Param("filename")
	filePath := fmt.Sprintf("%s/%s", r.service.storagePath, filename)

	c.File(filePath)
}

// 数据类型定义

// FormAnalytics 表单分析数据
type FormAnalytics struct {
	FormID           string    `json:"form_id"`
	FormName         string    `json:"form_name"`
	TotalSubmissions int       `json:"total_submissions"`
	TodaySubmissions int       `json:"today_submissions"`
	QRCodeCount      int       `json:"qrcode_count"`
	CreatedAt        time.Time `json:"created_at"`
}

// PublicFormData 公开表单数据
type PublicFormData struct {
	ID          string                   `json:"id"`
	FormName    string                   `json:"form_name"`
	FormType    string                   `json:"form_type"`
	Description string                   `json:"description"`
	Fields      []models.QRCodeFormField `json:"fields"`
}

// CloneFormRequest 克隆表单请求
type CloneFormRequest struct {
	NewFormName string `json:"new_form_name" binding:"required"`
}

// FormTemplate 表单模板
type FormTemplate struct {
	ID          string      `json:"id"`
	Name        string      `json:"name"`
	Type        string      `json:"type"`
	Description string      `json:"description"`
	Fields      []FieldInfo `json:"fields"`
}

// ErrorResponse 错误响应
type ErrorResponse struct {
	Error string `json:"error"`
}

// SuccessResponse 成功响应
type SuccessResponse struct {
	Message string `json:"message"`
}

// GetPresetFormTemplates 获取预设表单模板
func GetPresetFormTemplates() []FormTemplate {
	return []FormTemplate{
		{
			ID:          "signin_template",
			Name:        "员工签到表单",
			Type:        "registration",
			Description: "用于员工日常签到的表单模板",
			Fields: []FieldInfo{
				{
					FieldKey:     "name",
					FieldLabel:   "姓名",
					FieldType:    "text",
					IsRequired:   true,
					DisplayOrder: 1,
				},
				{
					FieldKey:     "employee_id",
					FieldLabel:   "员工编号",
					FieldType:    "text",
					IsRequired:   true,
					DisplayOrder: 2,
				},
				{
					FieldKey:     "department",
					FieldLabel:   "部门",
					FieldType:    "select",
					IsRequired:   true,
					Options:      []string{"技术部", "产品部", "运营部", "人力资源部"},
					DisplayOrder: 3,
				},
				{
					FieldKey:     "signin_time",
					FieldLabel:   "签到时间",
					FieldType:    "text",
					IsRequired:   false,
					DisplayOrder: 4,
				},
			},
		},
		{
			ID:          "feedback_template",
			Name:        "客户反馈表单",
			Type:        "feedback",
			Description: "用于收集客户反馈的表单模板",
			Fields: []FieldInfo{
				{
					FieldKey:     "customer_name",
					FieldLabel:   "客户姓名",
					FieldType:    "text",
					IsRequired:   true,
					DisplayOrder: 1,
				},
				{
					FieldKey:     "contact_phone",
					FieldLabel:   "联系电话",
					FieldType:    "text",
					IsRequired:   false,
					DisplayOrder: 2,
				},
				{
					FieldKey:     "satisfaction",
					FieldLabel:   "满意度评价",
					FieldType:    "radio",
					IsRequired:   true,
					Options:      []string{"非常满意", "满意", "一般", "不满意", "非常不满意"},
					DisplayOrder: 3,
				},
				{
					FieldKey:     "feedback_content",
					FieldLabel:   "反馈内容",
					FieldType:    "text",
					IsRequired:   false,
					DisplayOrder: 4,
				},
			},
		},
		{
			ID:          "event_registration_template",
			Name:        "活动报名表单",
			Type:        "registration",
			Description: "用于活动报名的表单模板",
			Fields: []FieldInfo{
				{
					FieldKey:     "participant_name",
					FieldLabel:   "参与者姓名",
					FieldType:    "text",
					IsRequired:   true,
					DisplayOrder: 1,
				},
				{
					FieldKey:     "participant_phone",
					FieldLabel:   "联系电话",
					FieldType:    "text",
					IsRequired:   true,
					DisplayOrder: 2,
				},
				{
					FieldKey:     "participant_email",
					FieldLabel:   "邮箱地址",
					FieldType:    "text",
					IsRequired:   false,
					DisplayOrder: 3,
				},
				{
					FieldKey:     "dietary_requirements",
					FieldLabel:   "饮食要求",
					FieldType:    "checkbox",
					IsRequired:   false,
					Options:      []string{"无特殊要求", "素食", "清真", "过敏食物说明"},
					DisplayOrder: 4,
				},
			},
		},
	}
}
