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

package knowledge

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/linux-do/cdk-office/internal/db"
	"github.com/linux-do/cdk-office/internal/models"
)

// Handler 知识库处理器
type Handler struct {
	service         *Service
	wechatService   *WeChatService
	approvalService *ShareApprovalService
}

// NewHandler 创建知识库处理器
func NewHandler() *Handler {
	db := db.GetDB()
	return &Handler{
		service:         NewService(db),
		wechatService:   NewWeChatService(db),
		approvalService: NewShareApprovalService(db),
	}
}

// RegisterRoutes 注册路由
func (h *Handler) RegisterRoutes(r *gin.RouterGroup) {
	knowledgeGroup := r.Group("/knowledge")
	{
		// 个人知识库 CRUD
		knowledgeGroup.POST("", h.CreateKnowledge)
		knowledgeGroup.GET("", h.ListKnowledge)
		knowledgeGroup.GET("/:id", h.GetKnowledge)
		knowledgeGroup.PUT("/:id", h.UpdateKnowledge)
		knowledgeGroup.DELETE("/:id", h.DeleteKnowledge)

		// 搜索和统计
		knowledgeGroup.POST("/search", h.SearchKnowledge)
		knowledgeGroup.GET("/statistics", h.GetStatistics)
		knowledgeGroup.GET("/tags/popular", h.GetPopularTags)

		// 分享功能
		knowledgeGroup.POST("/:id/share", h.ShareToTeam)
		knowledgeGroup.GET("/:id/share-status", h.GetShareStatus)

		// 批量操作
		knowledgeGroup.POST("/batch/delete", h.BatchDelete)
		knowledgeGroup.POST("/batch/update", h.BatchUpdate)

		// 导入导出
		knowledgeGroup.POST("/import", h.ImportKnowledge)
		knowledgeGroup.POST("/export", h.ExportKnowledge)

		// 微信聊天记录
		wechatGroup := knowledgeGroup.Group("/wechat")
		{
			wechatGroup.POST("/upload", h.UploadWeChatRecords)
			wechatGroup.GET("/records", h.ListWeChatRecords)
			wechatGroup.GET("/records/:id", h.GetWeChatRecord)
			wechatGroup.DELETE("/records/:id", h.DeleteWeChatRecord)
			wechatGroup.POST("/records/:id/archive", h.ArchiveWeChatRecord)
		}
	
		// 分享审批流程
		shareGroup := knowledgeGroup.Group("/share")
		{
			shareGroup.POST("/applications", h.SubmitShareApplication)
			shareGroup.GET("/applications", h.ListShareApplications)
			shareGroup.GET("/applications/:id", h.GetShareApplicationDetail)
			shareGroup.POST("/applications/:id/review", h.ReviewShareApplication)
			shareGroup.GET("/statistics", h.GetShareStatistics)
		}
}

// CreateKnowledge 创建个人知识
// @Summary 创建个人知识
// @Description 创建一条新的个人知识记录
// @Tags knowledge
// @Accept json
// @Produce json
// @Param request body CreateKnowledgeRequest true "创建知识请求"
// @Success 200 {object} models.PersonalKnowledgeBase
// @Failure 400 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /api/v1/knowledge [post]
func (h *Handler) CreateKnowledge(c *gin.Context) {
	var req CreateKnowledgeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body", "details": err.Error()})
		return
	}

	// 从上下文获取用户ID
	userID := c.GetString("user_id")
	if userID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}
	req.UserID = userID

	// 设置默认值
	if req.ContentType == "" {
		req.ContentType = "markdown"
	}
	if req.Privacy == "" {
		req.Privacy = "private"
	}
	if req.SourceType == "" {
		req.SourceType = "manual"
	}

	knowledge, err := h.service.CreateKnowledge(c.Request.Context(), &req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create knowledge", "details": err.Error()})
		return
	}

	c.JSON(http.StatusOK, knowledge)
}

// GetKnowledge 获取个人知识详情
// @Summary 获取个人知识详情
// @Description 根据ID获取个人知识的详细信息
// @Tags knowledge
// @Produce json
// @Param id path string true "知识ID"
// @Success 200 {object} models.PersonalKnowledgeBase
// @Failure 404 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /api/v1/knowledge/{id} [get]
func (h *Handler) GetKnowledge(c *gin.Context) {
	knowledgeID := c.Param("id")
	userID := c.GetString("user_id")

	knowledge, err := h.service.GetKnowledge(c.Request.Context(), userID, knowledgeID)
	if err != nil {
		if err.Error() == "knowledge not found" {
			c.JSON(http.StatusNotFound, gin.H{"error": "Knowledge not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get knowledge", "details": err.Error()})
		return
	}

	c.JSON(http.StatusOK, knowledge)
}

// UpdateKnowledge 更新个人知识
// @Summary 更新个人知识
// @Description 更新指定ID的个人知识
// @Tags knowledge
// @Accept json
// @Produce json
// @Param id path string true "知识ID"
// @Param request body UpdateKnowledgeRequest true "更新知识请求"
// @Success 200 {object} models.PersonalKnowledgeBase
// @Failure 400 {object} map[string]interface{}
// @Failure 404 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /api/v1/knowledge/{id} [put]
func (h *Handler) UpdateKnowledge(c *gin.Context) {
	knowledgeID := c.Param("id")
	userID := c.GetString("user_id")

	var req UpdateKnowledgeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body", "details": err.Error()})
		return
	}

	knowledge, err := h.service.UpdateKnowledge(c.Request.Context(), userID, knowledgeID, &req)
	if err != nil {
		if err.Error() == "knowledge not found" {
			c.JSON(http.StatusNotFound, gin.H{"error": "Knowledge not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update knowledge", "details": err.Error()})
		return
	}

	c.JSON(http.StatusOK, knowledge)
}

// DeleteKnowledge 删除个人知识
// @Summary 删除个人知识
// @Description 删除指定ID的个人知识
// @Tags knowledge
// @Produce json
// @Param id path string true "知识ID"
// @Success 200 {object} map[string]interface{}
// @Failure 404 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /api/v1/knowledge/{id} [delete]
func (h *Handler) DeleteKnowledge(c *gin.Context) {
	knowledgeID := c.Param("id")
	userID := c.GetString("user_id")

	err := h.service.DeleteKnowledge(c.Request.Context(), userID, knowledgeID)
	if err != nil {
		if err.Error() == "knowledge not found" {
			c.JSON(http.StatusNotFound, gin.H{"error": "Knowledge not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete knowledge", "details": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Knowledge deleted successfully"})
}

// ListKnowledge 列出个人知识
// @Summary 列出个人知识
// @Description 获取当前用户的个人知识列表，支持分页和筛选
// @Tags knowledge
// @Produce json
// @Param page query int false "页码" default(1)
// @Param page_size query int false "每页大小" default(20)
// @Param category query string false "分类筛选"
// @Param privacy query string false "隐私级别筛选"
// @Param source_type query string false "来源类型筛选"
// @Param keyword query string false "关键词搜索"
// @Param sort_by query string false "排序字段" default("updated_at")
// @Success 200 {object} ListKnowledgeResponse
// @Failure 400 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /api/v1/knowledge [get]
func (h *Handler) ListKnowledge(c *gin.Context) {
	userID := c.GetString("user_id")

	// 解析查询参数
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))

	req := &ListKnowledgeRequest{
		UserID:     userID,
		Page:       page,
		PageSize:   pageSize,
		Category:   c.Query("category"),
		Privacy:    c.Query("privacy"),
		SourceType: c.Query("source_type"),
		Keyword:    c.Query("keyword"),
		SortBy:     c.DefaultQuery("sort_by", "updated_at"),
	}

	// 解析标签参数（多个标签用逗号分隔）
	if tagsParam := c.Query("tags"); tagsParam != "" {
		// 这里简化处理，实际可能需要更复杂的解析
		req.Tags = []string{tagsParam}
	}

	response, err := h.service.ListKnowledge(c.Request.Context(), req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to list knowledge", "details": err.Error()})
		return
	}

	c.JSON(http.StatusOK, response)
}

// SearchKnowledge 搜索个人知识
// @Summary 搜索个人知识
// @Description 在个人知识库中进行全文搜索
// @Tags knowledge
// @Accept json
// @Produce json
// @Param request body SearchKnowledgeRequest true "搜索请求"
// @Success 200 {object} SearchKnowledgeResponse
// @Failure 400 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /api/v1/knowledge/search [post]
func (h *Handler) SearchKnowledge(c *gin.Context) {
	var req SearchKnowledgeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body", "details": err.Error()})
		return
	}

	req.UserID = c.GetString("user_id")
	if req.Page <= 0 {
		req.Page = 1
	}
	if req.PageSize <= 0 || req.PageSize > 100 {
		req.PageSize = 20
	}

	response, err := h.service.SearchKnowledge(c.Request.Context(), &req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to search knowledge", "details": err.Error()})
		return
	}

	c.JSON(http.StatusOK, response)
}

// GetStatistics 获取知识库统计信息
// @Summary 获取知识库统计信息
// @Description 获取当前用户的知识库统计数据
// @Tags knowledge
// @Produce json
// @Success 200 {object} KnowledgeStatistics
// @Failure 500 {object} map[string]interface{}
// @Router /api/v1/knowledge/statistics [get]
func (h *Handler) GetStatistics(c *gin.Context) {
	userID := c.GetString("user_id")

	stats, err := h.service.GetKnowledgeStatistics(c.Request.Context(), userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get statistics", "details": err.Error()})
		return
	}

	c.JSON(http.StatusOK, stats)
}

// GetPopularTags 获取热门标签
// @Summary 获取热门标签
// @Description 获取用户最常用的标签列表
// @Tags knowledge
// @Produce json
// @Param limit query int false "返回数量限制" default(10)
// @Success 200 {array} TagStat
// @Failure 500 {object} map[string]interface{}
// @Router /api/v1/knowledge/tags/popular [get]
func (h *Handler) GetPopularTags(c *gin.Context) {
	userID := c.GetString("user_id")
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))

	tags, err := h.service.GetPopularTags(c.Request.Context(), userID, limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get popular tags", "details": err.Error()})
		return
	}

	c.JSON(http.StatusOK, tags)
}

// ShareToTeam 分享知识到团队
// @Summary 分享知识到团队
// @Description 将个人知识分享到团队知识库，需要审核
// @Tags knowledge
// @Accept json
// @Produce json
// @Param id path string true "知识ID"
// @Param request body ShareToTeamRequest true "分享请求"
// @Success 200 {object} models.PersonalKnowledgeShare
// @Failure 400 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /api/v1/knowledge/{id}/share [post]
func (h *Handler) ShareToTeam(c *gin.Context) {
	knowledgeID := c.Param("id")
	userID := c.GetString("user_id")

	var req ShareToTeamRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body", "details": err.Error()})
		return
	}

	req.KnowledgeID = knowledgeID
	req.UserID = userID

	share, err := h.service.ShareToTeam(c.Request.Context(), &req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to share knowledge", "details": err.Error()})
		return
	}

	c.JSON(http.StatusOK, share)
}

// GetShareStatus 获取分享状态
// @Summary 获取分享状态
// @Description 获取知识的分享状态和审核信息
// @Tags knowledge
// @Produce json
// @Param id path string true "知识ID"
// @Success 200 {object} models.PersonalKnowledgeShare
// @Failure 404 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /api/v1/knowledge/{id}/share-status [get]
func (h *Handler) GetShareStatus(c *gin.Context) {
	knowledgeID := c.Param("id")
	userID := c.GetString("user_id")

	share, err := h.service.GetShareStatus(c.Request.Context(), userID, knowledgeID)
	if err != nil {
		if err.Error() == "share record not found" {
			c.JSON(http.StatusNotFound, gin.H{"error": "Share record not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get share status", "details": err.Error()})
		return
	}

	c.JSON(http.StatusOK, share)
}

// BatchDelete 批量删除知识
// @Summary 批量删除知识
// @Description 批量删除多个知识条目
// @Tags knowledge
// @Accept json
// @Produce json
// @Param request body BatchDeleteRequest true "批量删除请求"
// @Success 200 {object} BatchDeleteResponse
// @Failure 400 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /api/v1/knowledge/batch/delete [post]
func (h *Handler) BatchDelete(c *gin.Context) {
	var req BatchDeleteRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body", "details": err.Error()})
		return
	}

	req.UserID = c.GetString("user_id")

	// 简化实现：逐个删除
	var successCount, failedCount int
	var failedIDs []string

	for _, knowledgeID := range req.KnowledgeIDs {
		if err := h.service.DeleteKnowledge(c.Request.Context(), req.UserID, knowledgeID); err != nil {
			failedCount++
			failedIDs = append(failedIDs, knowledgeID)
		} else {
			successCount++
		}
	}

	response := &BatchDeleteResponse{
		SuccessCount: successCount,
		FailedCount:  failedCount,
		FailedIDs:    failedIDs,
	}

	c.JSON(http.StatusOK, response)
}

// BatchUpdate 批量更新知识
// @Summary 批量更新知识
// @Description 批量更新多个知识条目的属性
// @Tags knowledge
// @Accept json
// @Produce json
// @Param request body BatchUpdateRequest true "批量更新请求"
// @Success 200 {object} BatchUpdateResponse
// @Failure 400 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /api/v1/knowledge/batch/update [post]
func (h *Handler) BatchUpdate(c *gin.Context) {
	var req BatchUpdateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body", "details": err.Error()})
		return
	}

	req.UserID = c.GetString("user_id")

	// 简化实现：直接数据库批量更新
	database := db.GetDB()
	result := database.Model(&models.PersonalKnowledgeBase{}).
		Where("id IN ? AND user_id = ?", req.KnowledgeIDs, req.UserID).
		Updates(req.Updates)

	if result.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to batch update", "details": result.Error.Error()})
		return
	}

	response := &BatchUpdateResponse{
		SuccessCount: int(result.RowsAffected),
		FailedCount:  len(req.KnowledgeIDs) - int(result.RowsAffected),
	}

	c.JSON(http.StatusOK, response)
}

// ImportKnowledge 导入知识
// @Summary 导入知识
// @Description 从外部格式导入知识数据
// @Tags knowledge
// @Accept json
// @Produce json
// @Param request body ImportKnowledgeRequest true "导入请求"
// @Success 200 {object} ImportKnowledgeResponse
// @Failure 400 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /api/v1/knowledge/import [post]
func (h *Handler) ImportKnowledge(c *gin.Context) {
	var req ImportKnowledgeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body", "details": err.Error()})
		return
	}

	req.UserID = c.GetString("user_id")

	// TODO: 实现具体的导入逻辑
	// 这里只是示例响应
	response := &ImportKnowledgeResponse{
		ImportedCount: 0,
		SkippedCount:  0,
		FailedCount:   0,
		Errors:        []string{"Import functionality not implemented yet"},
	}

	c.JSON(http.StatusOK, response)
}

// ExportKnowledge 导出知识
// @Summary 导出知识
// @Description 将知识数据导出为指定格式
// @Tags knowledge
// @Accept json
// @Produce json
// @Param request body ExportKnowledgeRequest true "导出请求"
// @Success 200 {object} ExportKnowledgeResponse
// @Failure 400 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /api/v1/knowledge/export [post]
func (h *Handler) ExportKnowledge(c *gin.Context) {
	var req ExportKnowledgeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body", "details": err.Error()})
		return
	}

	req.UserID = c.GetString("user_id")

	// TODO: 实现具体的导出逻辑
	// 这里只是示例响应
	response := &ExportKnowledgeResponse{
		ExportedCount: 0,
		Format:        req.Format,
		Data:          "",
		FileName:      "knowledge_export." + req.Format,
		FileSize:      0,
	}

	c.JSON(http.StatusOK, response)
}

// UploadWeChatRecords 上传微信聊天记录
// @Summary 上传微信聊天记录
// @Description 上传并处理微信聊天记录，支持OCR和内容分析
// @Tags knowledge
// @Accept json
// @Produce json
// @Param request body WeChatUploadRequest true "上传微信聊天记录请求"
// @Success 200 {object} WeChatUploadResponse
// @Failure 400 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /api/v1/knowledge/wechat/upload [post]
func (h *Handler) UploadWeChatRecords(c *gin.Context) {
	var req WeChatUploadRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body", "details": err.Error()})
		return
	}

	req.UserID = c.GetString("user_id")
	if req.UserID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	response, err := h.wechatService.ProcessWeChatUpload(c.Request.Context(), &req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to process WeChat upload", "details": err.Error()})
		return
	}

	c.JSON(http.StatusOK, response)
}

// ListWeChatRecords 列出微信聊天记录
// @Summary 列出微信聊天记录
// @Description 获取用户的微信聊天记录列表，支持分页和筛选
// @Tags knowledge
// @Produce json
// @Param session_name query string false "会话名称"
// @Param message_type query string false "消息类型"
// @Param start_date query string false "开始日期 (YYYY-MM-DD)"
// @Param end_date query string false "结束日期 (YYYY-MM-DD)"
// @Param keyword query string false "关键词搜索"
// @Param page query int false "页码" default(1)
// @Param page_size query int false "每页大小" default(20)
// @Success 200 {object} ListWeChatRecordsResponse
// @Failure 400 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /api/v1/knowledge/wechat/records [get]
func (h *Handler) ListWeChatRecords(c *gin.Context) {
	userID := c.GetString("user_id")

	// 解析查询参数
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))

	req := &ListWeChatRecordsRequest{
		UserID:      userID,
		SessionName: c.Query("session_name"),
		MessageType: c.Query("message_type"),
		StartDate:   c.Query("start_date"),
		EndDate:     c.Query("end_date"),
		Keyword:     c.Query("keyword"),
		Page:        page,
		PageSize:    pageSize,
	}

	response, err := h.wechatService.ListWeChatRecords(c.Request.Context(), req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to list WeChat records", "details": err.Error()})
		return
	}

	c.JSON(http.StatusOK, response)
}

// GetWeChatRecord 获取微信聊天记录详情
// @Summary 获取微信聊天记录详情
// @Description 根据ID获取微信聊天记录的详细信息
// @Tags knowledge
// @Produce json
// @Param id path string true "记录ID"
// @Success 200 {object} models.WeChatRecord
// @Failure 404 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /api/v1/knowledge/wechat/records/{id} [get]
func (h *Handler) GetWeChatRecord(c *gin.Context) {
	recordID := c.Param("id")
	userID := c.GetString("user_id")

	record, err := h.wechatService.GetWeChatRecord(c.Request.Context(), userID, recordID)
	if err != nil {
		if err.Error() == "wechat record not found" {
			c.JSON(http.StatusNotFound, gin.H{"error": "WeChat record not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get WeChat record", "details": err.Error()})
		return
	}

	c.JSON(http.StatusOK, record)
}

// DeleteWeChatRecord 删除微信聊天记录
// @Summary 删除微信聊天记录
// @Description 删除指定ID的微信聊天记录及相关文件
// @Tags knowledge
// @Produce json
// @Param id path string true "记录ID"
// @Success 200 {object} map[string]interface{}
// @Failure 404 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /api/v1/knowledge/wechat/records/{id} [delete]
func (h *Handler) DeleteWeChatRecord(c *gin.Context) {
	recordID := c.Param("id")
	userID := c.GetString("user_id")

	err := h.wechatService.DeleteWeChatRecord(c.Request.Context(), userID, recordID)
	if err != nil {
		if err.Error() == "wechat record not found" {
			c.JSON(http.StatusNotFound, gin.H{"error": "WeChat record not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete WeChat record", "details": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "WeChat record deleted successfully"})
}

// ArchiveWeChatRecord 归档微信聊天记录到知识库
// @Summary 归档微信聊天记录
// @Description 将微信聊天记录归档到个人知识库
// @Tags knowledge
// @Accept json
// @Produce json
// @Param id path string true "记录ID"
// @Param request body ArchiveWeChatRecordRequest true "归档请求"
// @Success 200 {object} models.PersonalKnowledgeBase
// @Failure 400 {object} map[string]interface{}
// @Failure 404 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /api/v1/knowledge/wechat/records/{id}/archive [post]
func (h *Handler) ArchiveWeChatRecord(c *gin.Context) {
	recordID := c.Param("id")
	userID := c.GetString("user_id")

	var req ArchiveWeChatRecordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body", "details": err.Error()})
		return
	}

	req.RecordID = recordID
	req.UserID = userID

	knowledge, err := h.wechatService.ArchiveWeChatRecord(c.Request.Context(), &req)
	if err != nil {
		if err.Error() == "wechat record not found" {
			c.JSON(http.StatusNotFound, gin.H{"error": "WeChat record not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to archive WeChat record", "details": err.Error()})
		return
	}

	c.JSON(http.StatusOK, knowledge)
}

// SubmitShareApplication 提交分享申请
// @Summary 提交分享申请
// @Description 提交个人知识分享到团队知识库的申请
// @Tags knowledge
// @Accept json
// @Produce json
// @Param request body SubmitShareApplicationRequest true "提交分享申请"
// @Success 200 {object} models.PersonalKnowledgeShare
// @Failure 400 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /api/v1/knowledge/share/applications [post]
func (h *Handler) SubmitShareApplication(c *gin.Context) {
	var req SubmitShareApplicationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body", "details": err.Error()})
		return
	}

	req.UserID = c.GetString("user_id")
	if req.UserID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	// 默认创建审批工作流
	if !req.CreateWorkflow {
		req.CreateWorkflow = true
	}

	share, err := h.approvalService.SubmitShareApplication(c.Request.Context(), &req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to submit share application", "details": err.Error()})
		return
	}

	c.JSON(http.StatusOK, share)
}

// ListShareApplications 列出分享申请
// @Summary 列出分享申请
// @Description 获取分享申请列表，支持管理员和普通用户视图
// @Tags knowledge
// @Produce json
// @Param team_id query string false "团队ID（管理员用）"
// @Param status query string false "申请状态"
// @Param page query int false "页码" default(1)
// @Param page_size query int false "每页大小" default(20)
// @Success 200 {object} ListShareApplicationsResponse
// @Failure 400 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /api/v1/knowledge/share/applications [get]
func (h *Handler) ListShareApplications(c *gin.Context) {
	userID := c.GetString("user_id")
	userRole := c.GetString("user_role") // 从中间件获取用户角色

	// 解析查询参数
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))

	req := &ListShareApplicationsRequest{
		UserID:   userID,
		TeamID:   c.Query("team_id"),
		Role:     userRole,
		Status:   c.Query("status"),
		Page:     page,
		PageSize: pageSize,
	}

	response, err := h.approvalService.ListShareApplications(c.Request.Context(), req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to list share applications", "details": err.Error()})
		return
	}

	c.JSON(http.StatusOK, response)
}

// GetShareApplicationDetail 获取分享申请详情
// @Summary 获取分享申请详情
// @Description 根据ID获取分享申请的详细信息，包括审批流程
// @Tags knowledge
// @Produce json
// @Param id path string true "申请ID"
// @Success 200 {object} ShareApplicationDetail
// @Failure 404 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /api/v1/knowledge/share/applications/{id} [get]
func (h *Handler) GetShareApplicationDetail(c *gin.Context) {
	shareID := c.Param("id")
	userID := c.GetString("user_id")
	userRole := c.GetString("user_role")

	detail, err := h.approvalService.GetShareApplicationDetail(c.Request.Context(), shareID, userID, userRole)
	if err != nil {
		if err.Error() == "share application not found" {
			c.JSON(http.StatusNotFound, gin.H{"error": "Share application not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get share application detail", "details": err.Error()})
		return
	}

	c.JSON(http.StatusOK, detail)
}

// ReviewShareApplication 审批分享申请
// @Summary 审批分享申请
// @Description 管理员审批知识分享申请（通过或拒绝）
// @Tags knowledge
// @Accept json
// @Produce json
// @Param id path string true "申请ID"
// @Param request body ReviewShareApplicationRequest true "审批请求"
// @Success 200 {object} models.PersonalKnowledgeShare
// @Failure 400 {object} map[string]interface{}
// @Failure 403 {object} map[string]interface{}
// @Failure 404 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /api/v1/knowledge/share/applications/{id}/review [post]
func (h *Handler) ReviewShareApplication(c *gin.Context) {
	shareID := c.Param("id")
	userID := c.GetString("user_id")
	userRole := c.GetString("user_role")

	// 检查权限（只有管理员可以审批）
	if userRole != "admin" && userRole != "super_admin" {
		c.JSON(http.StatusForbidden, gin.H{"error": "Insufficient permissions"})
		return
	}

	var req ReviewShareApplicationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body", "details": err.Error()})
		return
	}

	req.ShareID = shareID
	req.ReviewerID = userID

	share, err := h.approvalService.ReviewShareApplication(c.Request.Context(), &req)
	if err != nil {
		if err.Error() == "share application not found" {
			c.JSON(http.StatusNotFound, gin.H{"error": "Share application not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to review share application", "details": err.Error()})
		return
	}

	c.JSON(http.StatusOK, share)
}

// GetShareStatistics 获取分享统计信息
// @Summary 获取分享统计信息
// @Description 获取知识分享的统计数据
// @Tags knowledge
// @Produce json
// @Param team_id query string false "团队ID（管理员用）"
// @Success 200 {object} ShareStatistics
// @Failure 500 {object} map[string]interface{}
// @Router /api/v1/knowledge/share/statistics [get]
func (h *Handler) GetShareStatistics(c *gin.Context) {
	userID := c.GetString("user_id")
	userRole := c.GetString("user_role")
	teamID := c.Query("team_id")

	stats, err := h.approvalService.GetShareStatistics(c.Request.Context(), userID, teamID, userRole)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get share statistics", "details": err.Error()})
		return
	}

	c.JSON(http.StatusOK, stats)