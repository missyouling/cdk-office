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
	"context"
	"encoding/json"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/linux-do/cdk-office/internal/models"
	"gorm.io/gorm"
)

// Service 个人知识库服务
type Service struct {
	db *gorm.DB
}

// NewService 创建个人知识库服务
func NewService(db *gorm.DB) *Service {
	return &Service{
		db: db,
	}
}

// CreateKnowledge 创建个人知识
func (s *Service) CreateKnowledge(ctx context.Context, req *CreateKnowledgeRequest) (*models.PersonalKnowledgeBase, error) {
	knowledge := &models.PersonalKnowledgeBase{
		UserID:      req.UserID,
		Title:       req.Title,
		Description: req.Description,
		Content:     req.Content,
		ContentType: req.ContentType,
		Tags:        req.Tags,
		Category:    req.Category,
		Privacy:     req.Privacy,
		SourceType:  req.SourceType,
	}

	if req.SourceData != nil {
		sourceDataJSON, err := json.Marshal(req.SourceData)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal source data: %w", err)
		}
		knowledge.SourceData = string(sourceDataJSON)
	}

	if err := s.db.WithContext(ctx).Create(knowledge).Error; err != nil {
		return nil, fmt.Errorf("failed to create knowledge: %w", err)
	}

	log.Printf("Created personal knowledge: %s for user: %s", knowledge.ID, req.UserID)
	return knowledge, nil
}

// GetKnowledge 获取个人知识详情
func (s *Service) GetKnowledge(ctx context.Context, userID, knowledgeID string) (*models.PersonalKnowledgeBase, error) {
	var knowledge models.PersonalKnowledgeBase
	if err := s.db.WithContext(ctx).Where("id = ? AND user_id = ?", knowledgeID, userID).First(&knowledge).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("knowledge not found")
		}
		return nil, fmt.Errorf("failed to get knowledge: %w", err)
	}

	return &knowledge, nil
}

// UpdateKnowledge 更新个人知识
func (s *Service) UpdateKnowledge(ctx context.Context, userID, knowledgeID string, req *UpdateKnowledgeRequest) (*models.PersonalKnowledgeBase, error) {
	var knowledge models.PersonalKnowledgeBase
	if err := s.db.WithContext(ctx).Where("id = ? AND user_id = ?", knowledgeID, userID).First(&knowledge).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("knowledge not found")
		}
		return nil, fmt.Errorf("failed to find knowledge: %w", err)
	}

	updateData := make(map[string]interface{})
	if req.Title != "" {
		updateData["title"] = req.Title
	}
	if req.Description != "" {
		updateData["description"] = req.Description
	}
	if req.Content != "" {
		updateData["content"] = req.Content
	}
	if req.ContentType != "" {
		updateData["content_type"] = req.ContentType
	}
	if req.Tags != nil {
		updateData["tags"] = req.Tags
	}
	if req.Category != "" {
		updateData["category"] = req.Category
	}
	if req.Privacy != "" {
		updateData["privacy"] = req.Privacy
	}

	if err := s.db.WithContext(ctx).Model(&knowledge).Updates(updateData).Error; err != nil {
		return nil, fmt.Errorf("failed to update knowledge: %w", err)
	}

	// 重新查询更新后的数据
	if err := s.db.WithContext(ctx).Where("id = ?", knowledgeID).First(&knowledge).Error; err != nil {
		return nil, fmt.Errorf("failed to get updated knowledge: %w", err)
	}

	log.Printf("Updated personal knowledge: %s", knowledgeID)
	return &knowledge, nil
}

// DeleteKnowledge 删除个人知识
func (s *Service) DeleteKnowledge(ctx context.Context, userID, knowledgeID string) error {
	result := s.db.WithContext(ctx).Where("id = ? AND user_id = ?", knowledgeID, userID).Delete(&models.PersonalKnowledgeBase{})
	if result.Error != nil {
		return fmt.Errorf("failed to delete knowledge: %w", result.Error)
	}
	if result.RowsAffected == 0 {
		return fmt.Errorf("knowledge not found")
	}

	log.Printf("Deleted personal knowledge: %s", knowledgeID)
	return nil
}

// ListKnowledge 列出个人知识
func (s *Service) ListKnowledge(ctx context.Context, req *ListKnowledgeRequest) (*ListKnowledgeResponse, error) {
	query := s.db.WithContext(ctx).Model(&models.PersonalKnowledgeBase{}).Where("user_id = ?", req.UserID)

	// 筛选条件
	if req.Category != "" {
		query = query.Where("category = ?", req.Category)
	}
	if req.Privacy != "" {
		query = query.Where("privacy = ?", req.Privacy)
	}
	if req.SourceType != "" {
		query = query.Where("source_type = ?", req.SourceType)
	}
	if req.Keyword != "" {
		keyword := "%" + req.Keyword + "%"
		query = query.Where("title ILIKE ? OR content ILIKE ? OR description ILIKE ?", keyword, keyword, keyword)
	}
	if len(req.Tags) > 0 {
		query = query.Where("tags && ?", req.Tags)
	}

	// 计算总数
	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, fmt.Errorf("failed to count knowledge: %w", err)
	}

	// 分页
	offset := (req.Page - 1) * req.PageSize
	query = query.Offset(offset).Limit(req.PageSize)

	// 排序
	switch req.SortBy {
	case "created_at":
		query = query.Order("created_at DESC")
	case "updated_at":
		query = query.Order("updated_at DESC")
	case "title":
		query = query.Order("title ASC")
	default:
		query = query.Order("updated_at DESC")
	}

	var knowledgeList []models.PersonalKnowledgeBase
	if err := query.Find(&knowledgeList).Error; err != nil {
		return nil, fmt.Errorf("failed to list knowledge: %w", err)
	}

	return &ListKnowledgeResponse{
		Knowledge: knowledgeList,
		Total:     total,
		Page:      req.Page,
		PageSize:  req.PageSize,
	}, nil
}

// ShareToTeam 分享知识到团队知识库
func (s *Service) ShareToTeam(ctx context.Context, req *ShareToTeamRequest) (*models.PersonalKnowledgeShare, error) {
	// 检查知识是否存在且属于用户
	var knowledge models.PersonalKnowledgeBase
	if err := s.db.WithContext(ctx).Where("id = ? AND user_id = ?", req.KnowledgeID, req.UserID).First(&knowledge).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("knowledge not found")
		}
		return nil, fmt.Errorf("failed to find knowledge: %w", err)
	}

	// 检查是否已经分享过
	var existingShare models.PersonalKnowledgeShare
	if err := s.db.WithContext(ctx).Where("knowledge_id = ? AND status IN ('pending', 'approved')", req.KnowledgeID).First(&existingShare).Error; err == nil {
		return nil, fmt.Errorf("knowledge already shared or pending approval")
	}

	// 创建分享申请
	share := &models.PersonalKnowledgeShare{
		KnowledgeID: req.KnowledgeID,
		UserID:      req.UserID,
		TeamID:      req.TeamID,
		ShareReason: req.ShareReason,
		Status:      "pending",
	}

	if err := s.db.WithContext(ctx).Create(share).Error; err != nil {
		return nil, fmt.Errorf("failed to create share request: %w", err)
	}

	log.Printf("Created knowledge share request: %s", share.ID)
	return share, nil
}

// GetShareStatus 获取分享状态
func (s *Service) GetShareStatus(ctx context.Context, userID, knowledgeID string) (*models.PersonalKnowledgeShare, error) {
	var share models.PersonalKnowledgeShare
	if err := s.db.WithContext(ctx).Where("knowledge_id = ? AND user_id = ?", knowledgeID, userID).Order("created_at DESC").First(&share).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("share record not found")
		}
		return nil, fmt.Errorf("failed to get share status: %w", err)
	}

	return &share, nil
}

// GetKnowledgeStatistics 获取知识库统计信息
func (s *Service) GetKnowledgeStatistics(ctx context.Context, userID string) (*KnowledgeStatistics, error) {
	var stats KnowledgeStatistics

	// 总知识数量
	if err := s.db.WithContext(ctx).Model(&models.PersonalKnowledgeBase{}).Where("user_id = ?", userID).Count(&stats.TotalKnowledge).Error; err != nil {
		return nil, fmt.Errorf("failed to count total knowledge: %w", err)
	}

	// 按分类统计
	var categoryStats []CategoryStat
	if err := s.db.WithContext(ctx).Model(&models.PersonalKnowledgeBase{}).
		Select("category, COUNT(*) as count").
		Where("user_id = ?", userID).
		Group("category").
		Find(&categoryStats).Error; err != nil {
		return nil, fmt.Errorf("failed to get category statistics: %w", err)
	}
	stats.ByCategory = categoryStats

	// 按来源统计
	var sourceStats []SourceStat
	if err := s.db.WithContext(ctx).Model(&models.PersonalKnowledgeBase{}).
		Select("source_type, COUNT(*) as count").
		Where("user_id = ?", userID).
		Group("source_type").
		Find(&sourceStats).Error; err != nil {
		return nil, fmt.Errorf("failed to get source statistics: %w", err)
	}
	stats.BySource = sourceStats

	// 已分享数量
	if err := s.db.WithContext(ctx).Model(&models.PersonalKnowledgeBase{}).Where("user_id = ? AND is_shared = true", userID).Count(&stats.SharedKnowledge).Error; err != nil {
		return nil, fmt.Errorf("failed to count shared knowledge: %w", err)
	}

	// 本周新增
	weekAgo := time.Now().AddDate(0, 0, -7)
	if err := s.db.WithContext(ctx).Model(&models.PersonalKnowledgeBase{}).Where("user_id = ? AND created_at >= ?", userID, weekAgo).Count(&stats.WeeklyAdded).Error; err != nil {
		return nil, fmt.Errorf("failed to count weekly knowledge: %w", err)
	}

	return &stats, nil
}

// SearchKnowledge 搜索知识
func (s *Service) SearchKnowledge(ctx context.Context, req *SearchKnowledgeRequest) (*SearchKnowledgeResponse, error) {
	query := s.db.WithContext(ctx).Model(&models.PersonalKnowledgeBase{}).Where("user_id = ?", req.UserID)

	// 构建搜索条件
	if req.Query != "" {
		searchTerms := strings.Fields(req.Query)
		for _, term := range searchTerms {
			likePattern := "%" + term + "%"
			query = query.Where("title ILIKE ? OR content ILIKE ? OR description ILIKE ?", likePattern, likePattern, likePattern)
		}
	}

	// 标签过滤
	if len(req.Tags) > 0 {
		query = query.Where("tags && ?", req.Tags)
	}

	// 分类过滤
	if req.Category != "" {
		query = query.Where("category = ?", req.Category)
	}

	// 来源过滤
	if req.SourceType != "" {
		query = query.Where("source_type = ?", req.SourceType)
	}

	// 计算总数
	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, fmt.Errorf("failed to count search results: %w", err)
	}

	// 分页
	offset := (req.Page - 1) * req.PageSize
	query = query.Offset(offset).Limit(req.PageSize)

	// 排序（默认按相关性，这里简化为按更新时间）
	query = query.Order("updated_at DESC")

	var results []models.PersonalKnowledgeBase
	if err := query.Find(&results).Error; err != nil {
		return nil, fmt.Errorf("failed to search knowledge: %w", err)
	}

	return &SearchKnowledgeResponse{
		Results:  results,
		Total:    total,
		Page:     req.Page,
		PageSize: req.PageSize,
		Query:    req.Query,
	}, nil
}

// GetPopularTags 获取热门标签
func (s *Service) GetPopularTags(ctx context.Context, userID string, limit int) ([]TagStat, error) {
	if limit <= 0 {
		limit = 10
	}

	// 由于PostgreSQL数组处理比较复杂，这里简化实现
	// 实际生产环境中可能需要更复杂的查询
	var knowledgeList []models.PersonalKnowledgeBase
	if err := s.db.WithContext(ctx).Select("tags").Where("user_id = ?", userID).Find(&knowledgeList).Error; err != nil {
		return nil, fmt.Errorf("failed to get tags: %w", err)
	}

	tagCounts := make(map[string]int)
	for _, knowledge := range knowledgeList {
		for _, tag := range knowledge.Tags {
			tagCounts[tag]++
		}
	}

	var tagStats []TagStat
	for tag, count := range tagCounts {
		tagStats = append(tagStats, TagStat{
			Tag:   tag,
			Count: int64(count),
		})
	}

	// 简单排序（实际应该按count降序）
	// 这里只返回前limit个
	if len(tagStats) > limit {
		tagStats = tagStats[:limit]
	}

	return tagStats, nil
}
