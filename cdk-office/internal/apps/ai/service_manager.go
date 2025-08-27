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

package ai

import (
	"context"
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/linux-do/cdk-office/internal/models"
	"gorm.io/gorm"
)

// ServiceManager AI服务管理器
type ServiceManager struct {
	db            *gorm.DB
	services      map[string]*models.AIServiceConfig
	healthChecker *ServiceHealthChecker
	mutex         sync.RWMutex
}

// NewServiceManager 创建新的服务管理器
func NewServiceManager(db *gorm.DB) *ServiceManager {
	sm := &ServiceManager{
		db:       db,
		services: make(map[string]*models.AIServiceConfig),
	}

	sm.healthChecker = NewServiceHealthChecker(db, sm)
	sm.loadServices()

	return sm
}

// loadServices 加载所有启用的服务配置
func (sm *ServiceManager) loadServices() error {
	sm.mutex.Lock()
	defer sm.mutex.Unlock()

	var configs []models.AIServiceConfig
	if err := sm.db.Where("is_enabled = ?", true).Find(&configs).Error; err != nil {
		return err
	}

	for _, config := range configs {
		sm.services[config.ID] = &config
	}

	log.Printf("Loaded %d AI service configurations", len(sm.services))
	return nil
}

// GetDefaultService 获取指定类型的默认服务
func (sm *ServiceManager) GetDefaultService(serviceType string) (*models.AIServiceConfig, error) {
	sm.mutex.RLock()
	defer sm.mutex.RUnlock()

	// 先查找默认服务
	for _, service := range sm.services {
		if service.ServiceType == serviceType && service.IsDefault && service.IsEnabled {
			return service, nil
		}
	}

	// 如果没有默认服务，返回第一个可用的服务
	for _, service := range sm.services {
		if service.ServiceType == serviceType && service.IsEnabled {
			return service, nil
		}
	}

	return nil, fmt.Errorf("no available service for type: %s", serviceType)
}

// GetBackupServices 获取指定类型的备用服务列表
func (sm *ServiceManager) GetBackupServices(serviceType, excludeID string) ([]*models.AIServiceConfig, error) {
	sm.mutex.RLock()
	defer sm.mutex.RUnlock()

	var backups []*models.AIServiceConfig
	for _, service := range sm.services {
		if service.ServiceType == serviceType && service.ID != excludeID && service.IsEnabled {
			backups = append(backups, service)
		}
	}

	// 按优先级排序
	for i := 0; i < len(backups)-1; i++ {
		for j := i + 1; j < len(backups); j++ {
			if backups[i].Priority < backups[j].Priority {
				backups[i], backups[j] = backups[j], backups[i]
			}
		}
	}

	return backups, nil
}

// CreateServiceConfig 创建服务配置
func (sm *ServiceManager) CreateServiceConfig(adminID string, config *models.AIServiceConfig) error {
	if !sm.isSuperAdmin(adminID) {
		return fmt.Errorf("only super admin can configure AI services")
	}

	config.CreatedBy = adminID
	config.UpdatedBy = adminID

	// 如果设置为默认，取消其他同类型服务的默认状态
	if config.IsDefault {
		if err := sm.db.Model(&models.AIServiceConfig{}).
			Where("service_type = ? AND is_default = true", config.ServiceType).
			Update("is_default", false).Error; err != nil {
			return err
		}
	}

	if err := sm.db.Create(config).Error; err != nil {
		return err
	}

	// 重新加载服务配置
	return sm.loadServices()
}

// UpdateServiceConfig 更新服务配置
func (sm *ServiceManager) UpdateServiceConfig(adminID, configID string, updates map[string]interface{}) error {
	if !sm.isSuperAdmin(adminID) {
		return fmt.Errorf("only super admin can configure AI services")
	}

	updates["updated_by"] = adminID
	updates["updated_at"] = time.Now()

	// 如果设置为默认，取消其他同类型服务的默认状态
	if isDefault, ok := updates["is_default"].(bool); ok && isDefault {
		var config models.AIServiceConfig
		if err := sm.db.First(&config, "id = ?", configID).Error; err != nil {
			return err
		}

		if err := sm.db.Model(&models.AIServiceConfig{}).
			Where("service_type = ? AND is_default = true AND id != ?", config.ServiceType, configID).
			Update("is_default", false).Error; err != nil {
			return err
		}
	}

	if err := sm.db.Model(&models.AIServiceConfig{}).Where("id = ?", configID).Updates(updates).Error; err != nil {
		return err
	}

	// 重新加载服务配置
	return sm.loadServices()
}

// DeleteServiceConfig 删除服务配置
func (sm *ServiceManager) DeleteServiceConfig(adminID, configID string) error {
	if !sm.isSuperAdmin(adminID) {
		return fmt.Errorf("only super admin can configure AI services")
	}

	if err := sm.db.Delete(&models.AIServiceConfig{}, "id = ?", configID).Error; err != nil {
		return err
	}

	// 重新加载服务配置
	return sm.loadServices()
}

// TestServiceConnection 测试服务连接
func (sm *ServiceManager) TestServiceConnection(adminID, configID string) error {
	if !sm.isSuperAdmin(adminID) {
		return fmt.Errorf("permission denied")
	}

	sm.mutex.RLock()
	config, exists := sm.services[configID]
	sm.mutex.RUnlock()

	if !exists {
		return fmt.Errorf("service config not found")
	}

	// 根据服务类型测试连接
	switch config.ServiceType {
	case "ai_chat":
		return sm.testAIChatService(config)
	case "ai_embedding":
		return sm.testAIEmbeddingService(config)
	case "ai_translation":
		return sm.testAITranslationService(config)
	default:
		return fmt.Errorf("unsupported service type: %s", config.ServiceType)
	}
}

// testAIChatService 测试AI对话服务
func (sm *ServiceManager) testAIChatService(config *models.AIServiceConfig) error {
	client := sm.createClient(config)

	// 发送测试请求
	testReq := map[string]interface{}{
		"messages": []map[string]string{
			{"role": "user", "content": "Hello, this is a test message."},
		},
		"max_tokens": 10,
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(config.Timeout)*time.Second)
	defer cancel()

	_, err := client.Chat(ctx, testReq)
	return err
}

// testAIEmbeddingService 测试AI向量化服务
func (sm *ServiceManager) testAIEmbeddingService(config *models.AIServiceConfig) error {
	client := sm.createClient(config)

	// 发送测试请求
	testReq := map[string]interface{}{
		"input": "This is a test text for embedding.",
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(config.Timeout)*time.Second)
	defer cancel()

	_, err := client.Embedding(ctx, testReq)
	return err
}

// testAITranslationService 测试AI翻译服务
func (sm *ServiceManager) testAITranslationService(config *models.AIServiceConfig) error {
	client := sm.createClient(config)

	// 发送测试请求
	testReq := map[string]interface{}{
		"text": "Hello, world!",
		"from": "en",
		"to":   "zh",
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(config.Timeout)*time.Second)
	defer cancel()

	_, err := client.Translate(ctx, testReq)
	return err
}

// createClient 根据配置创建客户端
func (sm *ServiceManager) createClient(config *models.AIServiceConfig) AIClient {
	switch config.Provider {
	case "openai":
		return NewOpenAIClient(config)
	case "baidu":
		return NewBaiduAIClient(config)
	case "tencent":
		return NewTencentAIClient(config)
	case "aliyun":
		return NewAliyunAIClient(config)
	default:
		return NewGenericAIClient(config)
	}
}

// isSuperAdmin 检查是否为超级管理员
func (sm *ServiceManager) isSuperAdmin(userID string) bool {
	// 这里应该实现真实的权限检查逻辑
	// 暂时简化处理，实际应该查询用户角色
	return true
}

// StartHealthCheckRoutine 启动健康检查例程
func (sm *ServiceManager) StartHealthCheckRoutine() {
	go sm.healthChecker.StartHealthCheckRoutine()
}

// TriggerFallback 触发服务降级
func (sm *ServiceManager) TriggerFallback(failedServiceID string) error {
	sm.mutex.Lock()
	defer sm.mutex.Unlock()

	failedService, exists := sm.services[failedServiceID]
	if !exists {
		return fmt.Errorf("failed service not found: %s", failedServiceID)
	}

	log.Printf("Triggering fallback for service: %s (%s)", failedService.ServiceName, failedService.ServiceType)

	// 获取备用服务
	backupServices, err := sm.GetBackupServices(failedService.ServiceType, failedServiceID)
	if err != nil {
		return err
	}

	if len(backupServices) > 0 {
		// 切换到第一个备用服务
		backupService := backupServices[0]
		log.Printf("Switching to backup service: %s", backupService.ServiceName)

		// 临时设置备用服务为默认
		if err := sm.db.Model(&models.AIServiceConfig{}).
			Where("service_type = ?", failedService.ServiceType).
			Update("is_default", false).Error; err != nil {
			return err
		}

		if err := sm.db.Model(&models.AIServiceConfig{}).
			Where("id = ?", backupService.ID).
			Update("is_default", true).Error; err != nil {
			return err
		}

		// 重新加载服务配置
		sm.loadServices()

		log.Printf("Successfully switched to backup service: %s", backupService.ServiceName)
	} else {
		// 没有备用服务，禁用功能
		log.Printf("No backup services available for type: %s, disabling functionality", failedService.ServiceType)
		return sm.disableServiceType(failedService.ServiceType)
	}

	return nil
}

// disableServiceType 禁用指定类型的服务功能
func (sm *ServiceManager) disableServiceType(serviceType string) error {
	// 这里可以实现功能禁用逻辑
	// 例如设置系统参数、发送通知等
	log.Printf("Service type %s has been disabled due to lack of available services", serviceType)
	return nil
}

// GetServiceList 获取服务列表（用于管理界面）
func (sm *ServiceManager) GetServiceList(serviceType string) ([]*models.AIServiceConfig, error) {
	sm.mutex.RLock()
	defer sm.mutex.RUnlock()

	var services []*models.AIServiceConfig
	for _, service := range sm.services {
		if serviceType == "" || service.ServiceType == serviceType {
			services = append(services, service)
		}
	}

	return services, nil
}

// GetPresetProviders 获取预设服务商配置
func (sm *ServiceManager) GetPresetProviders() *PresetServiceProviders {
	return &PresetServiceProviders{
		AIProviders: []ProviderTemplate{
			{
				Name:        "baidu_ai",
				Provider:    "baidu",
				DisplayName: "百度千帆",
				Logo:        "/assets/providers/baidu-logo.png",
				Description: "百度千帆大模型平台",
				ConfigTemplate: map[string]interface{}{
					"api_endpoint": "https://aip.baidubce.com/rpc/2.0/ai_custom/v1/wenxinworkshop/chat/completions",
				},
				RequiredFields: []string{"api_key", "secret_key"},
				SupportedTypes: []string{"ai_chat", "ai_embedding", "ai_translation"},
			},
			{
				Name:        "tencent_hunyuan",
				Provider:    "tencent",
				DisplayName: "腾讯混元",
				Logo:        "/assets/providers/tencent-logo.png",
				Description: "腾讯混元大模型",
				ConfigTemplate: map[string]interface{}{
					"api_endpoint": "https://hunyuan.tencentcloudapi.com",
				},
				RequiredFields: []string{"secret_id", "secret_key", "region"},
				SupportedTypes: []string{"ai_chat"},
			},
			{
				Name:        "aliyun_tongyi",
				Provider:    "aliyun",
				DisplayName: "阿里通义千问",
				Logo:        "/assets/providers/aliyun-logo.png",
				Description: "阿里云通义千问大模型",
				ConfigTemplate: map[string]interface{}{
					"api_endpoint": "https://dashscope.aliyuncs.com/api/v1/services/aigc/text-generation/generation",
				},
				RequiredFields: []string{"api_key"},
				SupportedTypes: []string{"ai_chat", "ai_embedding"},
			},
			{
				Name:        "openai",
				Provider:    "openai",
				DisplayName: "OpenAI",
				Logo:        "/assets/providers/openai-logo.png",
				Description: "OpenAI GPT模型",
				ConfigTemplate: map[string]interface{}{
					"api_endpoint": "https://api.openai.com/v1/chat/completions",
				},
				RequiredFields: []string{"api_key"},
				SupportedTypes: []string{"ai_chat", "ai_embedding", "ai_translation"},
			},
		},
	}
}

// PresetServiceProviders 预设服务商配置
type PresetServiceProviders struct {
	AIProviders []ProviderTemplate `json:"ai_providers"`
}

// ProviderTemplate 服务商模板
type ProviderTemplate struct {
	Name           string                 `json:"name"`
	Provider       string                 `json:"provider"`
	DisplayName    string                 `json:"display_name"`
	Logo           string                 `json:"logo"`
	Description    string                 `json:"description"`
	ConfigTemplate map[string]interface{} `json:"config_template"`
	RequiredFields []string               `json:"required_fields"`
	SupportedTypes []string               `json:"supported_types"`
}
