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
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/google/uuid"
	"github.com/linux-do/cdk-office/internal/models"
	"github.com/yeqown/go-qrcode/v2"
	"github.com/yeqown/go-qrcode/writer/standard"
	"gorm.io/gorm"
)

// QRCodeService 二维码服务
type QRCodeService struct {
	db          *gorm.DB
	storagePath string
	baseURL     string
}

// NewQRCodeService 创建二维码服务实例
func NewQRCodeService(db *gorm.DB, storagePath, baseURL string) *QRCodeService {
	return &QRCodeService{
		db:          db,
		storagePath: storagePath,
		baseURL:     baseURL,
	}
}

// FormRequest 表单创建请求
type FormRequest struct {
	FormName    string      `json:"form_name" binding:"required"`
	FormType    string      `json:"form_type" binding:"required"`
	Description string      `json:"description"`
	FormFields  []FieldInfo `json:"form_fields" binding:"required"`
}

// FieldInfo 字段信息
type FieldInfo struct {
	FieldKey     string   `json:"field_key" binding:"required"`
	FieldLabel   string   `json:"field_label" binding:"required"`
	FieldType    string   `json:"field_type" binding:"required"`
	IsRequired   bool     `json:"is_required"`
	DefaultValue string   `json:"default_value"`
	Options      []string `json:"options"`
	DisplayOrder int      `json:"display_order"`
}

// QRCodeRequest 二维码生成请求
type QRCodeRequest struct {
	FormID     string `json:"form_id" binding:"required"`
	Size       int    `json:"size"`
	FGColor    string `json:"fg_color"`
	BGColor    string `json:"bg_color"`
	ErrorLevel string `json:"error_level"`
	LogoPath   string `json:"logo_path"`
	ExpireTime int    `json:"expire_time"` // 过期时间（小时）
}

// SubmissionRequest 表单提交请求
type SubmissionRequest struct {
	Data map[string]interface{} `json:"data" binding:"required"`
}

// CreateForm 创建表单
func (s *QRCodeService) CreateForm(userID, teamID string, req *FormRequest) (*models.QRCodeForm, error) {
	// 创建表单
	form := &models.QRCodeForm{
		TeamID:      teamID,
		FormName:    req.FormName,
		FormType:    req.FormType,
		Description: req.Description,
		CreatedBy:   userID,
		UpdatedBy:   userID,
	}

	// 开始事务
	tx := s.db.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	// 创建表单
	if err := tx.Create(form).Error; err != nil {
		tx.Rollback()
		return nil, fmt.Errorf("failed to create form: %w", err)
	}

	// 创建表单字段
	for _, fieldInfo := range req.FormFields {
		field := &models.QRCodeFormField{
			FormID:       form.ID,
			FieldKey:     fieldInfo.FieldKey,
			FieldLabel:   fieldInfo.FieldLabel,
			FieldType:    fieldInfo.FieldType,
			IsRequired:   fieldInfo.IsRequired,
			DefaultValue: fieldInfo.DefaultValue,
			Options:      fieldInfo.Options,
			DisplayOrder: fieldInfo.DisplayOrder,
		}

		if err := tx.Create(field).Error; err != nil {
			tx.Rollback()
			return nil, fmt.Errorf("failed to create form field: %w", err)
		}
	}

	// 提交事务
	if err := tx.Commit().Error; err != nil {
		return nil, fmt.Errorf("failed to commit transaction: %w", err)
	}

	log.Printf("Created QRCode form: %s (ID: %s)", form.FormName, form.ID)
	return form, nil
}

// GetForm 获取表单详情
func (s *QRCodeService) GetForm(formID string) (*FormDetails, error) {
	var form models.QRCodeForm
	if err := s.db.First(&form, "id = ?", formID).Error; err != nil {
		return nil, fmt.Errorf("form not found: %w", err)
	}

	var fields []models.QRCodeFormField
	if err := s.db.Where("form_id = ?", formID).Order("display_order ASC").Find(&fields).Error; err != nil {
		return nil, fmt.Errorf("failed to get form fields: %w", err)
	}

	return &FormDetails{
		Form:   form,
		Fields: fields,
	}, nil
}

// FormDetails 表单详情
type FormDetails struct {
	Form   models.QRCodeForm        `json:"form"`
	Fields []models.QRCodeFormField `json:"fields"`
}

// ListForms 获取表单列表
func (s *QRCodeService) ListForms(teamID string, page, limit int) (*FormListResponse, error) {
	var forms []models.QRCodeForm
	var total int64

	query := s.db.Where("team_id = ?", teamID)

	// 获取总数
	if err := query.Model(&models.QRCodeForm{}).Count(&total).Error; err != nil {
		return nil, fmt.Errorf("failed to count forms: %w", err)
	}

	// 获取分页数据
	offset := (page - 1) * limit
	if err := query.Offset(offset).Limit(limit).Order("created_at DESC").Find(&forms).Error; err != nil {
		return nil, fmt.Errorf("failed to list forms: %w", err)
	}

	return &FormListResponse{
		Forms: forms,
		Total: int(total),
		Page:  page,
		Limit: limit,
	}, nil
}

// FormListResponse 表单列表响应
type FormListResponse struct {
	Forms []models.QRCodeForm `json:"forms"`
	Total int                 `json:"total"`
	Page  int                 `json:"page"`
	Limit int                 `json:"limit"`
}

// UpdateForm 更新表单
func (s *QRCodeService) UpdateForm(formID, userID string, req *FormRequest) error {
	// 检查表单是否存在
	var form models.QRCodeForm
	if err := s.db.First(&form, "id = ?", formID).Error; err != nil {
		return fmt.Errorf("form not found: %w", err)
	}

	// 开始事务
	tx := s.db.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	// 更新表单基本信息
	updates := map[string]interface{}{
		"form_name":   req.FormName,
		"form_type":   req.FormType,
		"description": req.Description,
		"updated_by":  userID,
		"updated_at":  time.Now(),
	}

	if err := tx.Model(&form).Updates(updates).Error; err != nil {
		tx.Rollback()
		return fmt.Errorf("failed to update form: %w", err)
	}

	// 删除旧的字段
	if err := tx.Where("form_id = ?", formID).Delete(&models.QRCodeFormField{}).Error; err != nil {
		tx.Rollback()
		return fmt.Errorf("failed to delete old fields: %w", err)
	}

	// 创建新的字段
	for _, fieldInfo := range req.FormFields {
		field := &models.QRCodeFormField{
			FormID:       formID,
			FieldKey:     fieldInfo.FieldKey,
			FieldLabel:   fieldInfo.FieldLabel,
			FieldType:    fieldInfo.FieldType,
			IsRequired:   fieldInfo.IsRequired,
			DefaultValue: fieldInfo.DefaultValue,
			Options:      fieldInfo.Options,
			DisplayOrder: fieldInfo.DisplayOrder,
		}

		if err := tx.Create(field).Error; err != nil {
			tx.Rollback()
			return fmt.Errorf("failed to create new field: %w", err)
		}
	}

	// 提交事务
	if err := tx.Commit().Error; err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	log.Printf("Updated QRCode form: %s (ID: %s)", req.FormName, formID)
	return nil
}

// DeleteForm 删除表单
func (s *QRCodeService) DeleteForm(formID, userID string) error {
	// 开始事务
	tx := s.db.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	// 删除表单字段
	if err := tx.Where("form_id = ?", formID).Delete(&models.QRCodeFormField{}).Error; err != nil {
		tx.Rollback()
		return fmt.Errorf("failed to delete form fields: %w", err)
	}

	// 删除二维码记录
	if err := tx.Where("form_id = ?", formID).Delete(&models.QRCodeRecord{}).Error; err != nil {
		tx.Rollback()
		return fmt.Errorf("failed to delete qrcode records: %w", err)
	}

	// 删除提交记录
	if err := tx.Where("form_id = ?", formID).Delete(&models.QRCodeFormSubmission{}).Error; err != nil {
		tx.Rollback()
		return fmt.Errorf("failed to delete submissions: %w", err)
	}

	// 删除表单
	if err := tx.Delete(&models.QRCodeForm{}, "id = ?", formID).Error; err != nil {
		tx.Rollback()
		return fmt.Errorf("failed to delete form: %w", err)
	}

	// 提交事务
	if err := tx.Commit().Error; err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	log.Printf("Deleted QRCode form: %s", formID)
	return nil
}

// GenerateQRCode 生成二维码
func (s *QRCodeService) GenerateQRCode(userID string, req *QRCodeRequest) (*models.QRCodeRecord, error) {
	// 检查表单是否存在
	var form models.QRCodeForm
	if err := s.db.First(&form, "id = ?", req.FormID).Error; err != nil {
		return nil, fmt.Errorf("form not found: %w", err)
	}

	// 构建二维码内容URL
	content := fmt.Sprintf("%s/forms/%s", s.baseURL, req.FormID)

	// 设置默认值
	if req.Size == 0 {
		req.Size = 256
	}
	if req.FGColor == "" {
		req.FGColor = "#000000"
	}
	if req.BGColor == "" {
		req.BGColor = "#FFFFFF"
	}
	if req.ErrorLevel == "" {
		req.ErrorLevel = "M"
	}

	// 生成二维码文件
	qrCodeFileName, err := s.generateQRCodeFile(content, req)
	if err != nil {
		return nil, fmt.Errorf("failed to generate qrcode file: %w", err)
	}

	// 计算过期时间
	expireAt := time.Now().Add(time.Duration(req.ExpireTime) * time.Hour)
	if req.ExpireTime == 0 {
		expireAt = time.Now().Add(24 * time.Hour) // 默认24小时过期
	}

	// 保存二维码记录
	record := &models.QRCodeRecord{
		FormID:    req.FormID,
		Content:   content,
		QRCodeURL: fmt.Sprintf("/qrcodes/%s", qrCodeFileName),
		ExpireAt:  expireAt,
		CreatedBy: userID,
	}

	if err := s.db.Create(record).Error; err != nil {
		return nil, fmt.Errorf("failed to create qrcode record: %w", err)
	}

	log.Printf("Generated QRCode for form: %s (Record ID: %s)", req.FormID, record.ID)
	return record, nil
}

// generateQRCodeFile 生成二维码图片文件
func (s *QRCodeService) generateQRCodeFile(content string, req *QRCodeRequest) (string, error) {
	// 确定容错级别
	var ecLevel qrcode.ErrorCorrectionLevel
	switch req.ErrorLevel {
	case "L":
		ecLevel = qrcode.ErrorCorrectionLow
	case "M":
		ecLevel = qrcode.ErrorCorrectionMedium
	case "Q":
		ecLevel = qrcode.ErrorCorrectionQuart
	case "H":
		ecLevel = qrcode.ErrorCorrectionHighest
	default:
		ecLevel = qrcode.ErrorCorrectionMedium
	}

	// 创建二维码
	qr, err := qrcode.New(content, qrcode.WithErrorCorrectionLevel(ecLevel))
	if err != nil {
		return "", fmt.Errorf("failed to create qrcode: %w", err)
	}

	// 生成文件名
	fileName := fmt.Sprintf("%s.jpeg", uuid.New().String())
	filePath := fmt.Sprintf("%s/%s", s.storagePath, fileName)

	// 配置写入器选项
	var options []standard.ImageOption

	// 设置大小
	if req.Size > 0 {
		options = append(options, standard.WithWidth(req.Size))
	}

	// 添加Logo（如果有）
	if req.LogoPath != "" {
		options = append(options, standard.WithLogoImageFileJPEG(req.LogoPath))
	}

	// 创建标准写入器
	writer, err := standard.New(filePath, options...)
	if err != nil {
		return "", fmt.Errorf("failed to create writer: %w", err)
	}

	// 保存二维码
	if err = qr.Save(writer); err != nil {
		return "", fmt.Errorf("failed to save qrcode: %w", err)
	}

	return fileName, nil
}

// SubmitForm 提交表单
func (s *QRCodeService) SubmitForm(formID, clientIP string, req *SubmissionRequest) error {
	// 检查表单是否存在
	var form models.QRCodeForm
	if err := s.db.First(&form, "id = ?", formID).Error; err != nil {
		return fmt.Errorf("form not found: %w", err)
	}

	// 获取表单字段定义
	var fields []models.QRCodeFormField
	if err := s.db.Where("form_id = ?", formID).Find(&fields).Error; err != nil {
		return fmt.Errorf("failed to get form fields: %w", err)
	}

	// 验证提交数据
	if err := s.validateSubmissionData(fields, req.Data); err != nil {
		return fmt.Errorf("validation failed: %w", err)
	}

	// 序列化提交数据
	submitData, err := json.Marshal(req.Data)
	if err != nil {
		return fmt.Errorf("failed to marshal submit data: %w", err)
	}

	// 保存提交记录
	submission := &models.QRCodeFormSubmission{
		FormID:     formID,
		SubmitData: string(submitData),
		SubmitIP:   clientIP,
	}

	if err := s.db.Create(submission).Error; err != nil {
		return fmt.Errorf("failed to create submission: %w", err)
	}

	log.Printf("Form submission created for form: %s (Submission ID: %s)", formID, submission.ID)
	return nil
}

// validateSubmissionData 验证提交数据
func (s *QRCodeService) validateSubmissionData(fields []models.QRCodeFormField, data map[string]interface{}) error {
	for _, field := range fields {
		value, exists := data[field.FieldKey]

		// 检查必填字段
		if field.IsRequired && (!exists || value == nil || value == "") {
			return fmt.Errorf("required field '%s' is missing or empty", field.FieldKey)
		}

		// 跳过空值字段的进一步验证
		if !exists || value == nil {
			continue
		}

		// 根据字段类型验证数据
		switch field.FieldType {
		case "number":
			if _, ok := value.(float64); !ok {
				return fmt.Errorf("field '%s' must be a number", field.FieldKey)
			}
		case "select", "radio":
			// 验证选项值是否在预定义的选项中
			if len(field.Options) > 0 {
				valueStr := fmt.Sprintf("%v", value)
				found := false
				for _, option := range field.Options {
					if option == valueStr {
						found = true
						break
					}
				}
				if !found {
					return fmt.Errorf("field '%s' value '%s' is not in allowed options", field.FieldKey, valueStr)
				}
			}
		case "checkbox":
			// 检查多选值
			if valueSlice, ok := value.([]interface{}); ok {
				if len(field.Options) > 0 {
					for _, v := range valueSlice {
						valueStr := fmt.Sprintf("%v", v)
						found := false
						for _, option := range field.Options {
							if option == valueStr {
								found = true
								break
							}
						}
						if !found {
							return fmt.Errorf("field '%s' value '%s' is not in allowed options", field.FieldKey, valueStr)
						}
					}
				}
			}
		}
	}

	return nil
}

// GetSubmissions 获取表单提交记录
func (s *QRCodeService) GetSubmissions(formID string, page, limit int) (*SubmissionListResponse, error) {
	var submissions []models.QRCodeFormSubmission
	var total int64

	query := s.db.Where("form_id = ?", formID)

	// 获取总数
	if err := query.Model(&models.QRCodeFormSubmission{}).Count(&total).Error; err != nil {
		return nil, fmt.Errorf("failed to count submissions: %w", err)
	}

	// 获取分页数据
	offset := (page - 1) * limit
	if err := query.Offset(offset).Limit(limit).Order("created_at DESC").Find(&submissions).Error; err != nil {
		return nil, fmt.Errorf("failed to list submissions: %w", err)
	}

	// 解析提交数据
	var submissionData []SubmissionData
	for _, submission := range submissions {
		var data map[string]interface{}
		if err := json.Unmarshal([]byte(submission.SubmitData), &data); err != nil {
			log.Printf("Failed to unmarshal submission data: %v", err)
			continue
		}

		submissionData = append(submissionData, SubmissionData{
			ID:        submission.ID,
			Data:      data,
			SubmitIP:  submission.SubmitIP,
			CreatedAt: submission.CreatedAt,
		})
	}

	return &SubmissionListResponse{
		Submissions: submissionData,
		Total:       int(total),
		Page:        page,
		Limit:       limit,
	}, nil
}

// SubmissionData 提交数据
type SubmissionData struct {
	ID        string                 `json:"id"`
	Data      map[string]interface{} `json:"data"`
	SubmitIP  string                 `json:"submit_ip"`
	CreatedAt time.Time              `json:"created_at"`
}

// SubmissionListResponse 提交记录列表响应
type SubmissionListResponse struct {
	Submissions []SubmissionData `json:"submissions"`
	Total       int              `json:"total"`
	Page        int              `json:"page"`
	Limit       int              `json:"limit"`
}

// GetQRCodeRecords 获取二维码记录列表
func (s *QRCodeService) GetQRCodeRecords(teamID string, page, limit int) (*QRCodeRecordListResponse, error) {
	var records []QRCodeRecordWithForm
	var total int64

	// 构建查询，关联表单信息
	query := s.db.Table("qr_code_records r").
		Select("r.*, f.form_name, f.form_type").
		Joins("JOIN qr_code_forms f ON r.form_id = f.id").
		Where("f.team_id = ?", teamID)

	// 获取总数
	if err := query.Count(&total).Error; err != nil {
		return nil, fmt.Errorf("failed to count records: %w", err)
	}

	// 获取分页数据
	offset := (page - 1) * limit
	if err := query.Offset(offset).Limit(limit).Order("r.created_at DESC").Scan(&records).Error; err != nil {
		return nil, fmt.Errorf("failed to list records: %w", err)
	}

	return &QRCodeRecordListResponse{
		Records: records,
		Total:   int(total),
		Page:    page,
		Limit:   limit,
	}, nil
}

// QRCodeRecordWithForm 带表单信息的二维码记录
type QRCodeRecordWithForm struct {
	models.QRCodeRecord
	FormName string `json:"form_name"`
	FormType string `json:"form_type"`
}

// QRCodeRecordListResponse 二维码记录列表响应
type QRCodeRecordListResponse struct {
	Records []QRCodeRecordWithForm `json:"records"`
	Total   int                    `json:"total"`
	Page    int                    `json:"page"`
	Limit   int                    `json:"limit"`
}
