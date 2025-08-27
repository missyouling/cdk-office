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

package filepreview

import (
	"encoding/json"
	"fmt"
	"os"
	"strconv"
	"strings"
)

// ConfigManager 配置管理器
type ConfigManager struct {
	config *Config
}

// NewConfigManager 创建配置管理器
func NewConfigManager() *ConfigManager {
	config := loadDefaultConfig()
	loadConfigFromEnv(config)

	return &ConfigManager{
		config: config,
	}
}

// GetConfig 获取配置
func (cm *ConfigManager) GetConfig() *Config {
	return cm.config
}

// UpdateConfig 更新配置
func (cm *ConfigManager) UpdateConfig(newConfig *Config) {
	cm.config = newConfig
}

// loadDefaultConfig 加载默认配置
func loadDefaultConfig() *Config {
	return &Config{
		Provider: "dify",
		DifyURL:  "http://dify-api:5001",
		KKFileView: KKFileViewConfig{
			Enabled: false,
			URL:     "http://kkfileview:8012",
			Timeout: 30,
		},
	}
}

// loadConfigFromEnv 从环境变量加载配置
func loadConfigFromEnv(config *Config) {
	// 文件预览提供者
	if provider := os.Getenv("FILE_PREVIEW_PROVIDER"); provider != "" {
		config.Provider = provider
	}

	// Dify配置
	if difyURL := os.Getenv("DIFY_URL"); difyURL != "" {
		config.DifyURL = difyURL
	}

	// KKFileView配置
	if kkEnabled := os.Getenv("KKFILEVIEW_ENABLED"); kkEnabled != "" {
		if enabled, err := strconv.ParseBool(kkEnabled); err == nil {
			config.KKFileView.Enabled = enabled
		}
	}

	if kkURL := os.Getenv("KKFILEVIEW_URL"); kkURL != "" {
		config.KKFileView.URL = kkURL
	}

	if kkTimeout := os.Getenv("KKFILEVIEW_TIMEOUT"); kkTimeout != "" {
		if timeout, err := strconv.Atoi(kkTimeout); err == nil {
			config.KKFileView.Timeout = timeout
		}
	}

	// 如果启用了KKFileView，则自动切换提供者
	if config.KKFileView.Enabled {
		config.Provider = "kkfileview"
	}
}

// SaveConfigToFile 保存配置到文件
func (cm *ConfigManager) SaveConfigToFile(filename string) error {
	data, err := json.MarshalIndent(cm.config, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(filename, data, 0644)
}

// LoadConfigFromFile 从文件加载配置
func (cm *ConfigManager) LoadConfigFromFile(filename string) error {
	data, err := os.ReadFile(filename)
	if err != nil {
		return err
	}

	return json.Unmarshal(data, &cm.config)
}

// ValidateConfig 验证配置
func (cm *ConfigManager) ValidateConfig() error {
	config := cm.config

	// 验证提供者
	if config.Provider != "dify" && config.Provider != "kkfileview" {
		return fmt.Errorf("invalid provider: %s, must be 'dify' or 'kkfileview'", config.Provider)
	}

	// 验证Dify配置
	if config.Provider == "dify" && config.DifyURL == "" {
		return fmt.Errorf("Dify URL is required when provider is 'dify'")
	}

	// 验证KKFileView配置
	if config.Provider == "kkfileview" {
		if !config.KKFileView.Enabled {
			return fmt.Errorf("KKFileView must be enabled when provider is 'kkfileview'")
		}
		if config.KKFileView.URL == "" {
			return fmt.Errorf("KKFileView URL is required when provider is 'kkfileview'")
		}
		if config.KKFileView.Timeout <= 0 {
			return fmt.Errorf("KKFileView timeout must be greater than 0")
		}
	}

	return nil
}

// GetProviderSpecificConfig 获取特定提供者的配置
func (cm *ConfigManager) GetProviderSpecificConfig() map[string]interface{} {
	config := cm.config
	result := make(map[string]interface{})

	switch config.Provider {
	case "dify":
		result["url"] = config.DifyURL
		result["features"] = []string{"basic_preview", "text_extraction"}
	case "kkfileview":
		result["url"] = config.KKFileView.URL
		result["enabled"] = config.KKFileView.Enabled
		result["timeout"] = config.KKFileView.Timeout
		result["features"] = []string{
			"enhanced_preview", "thumbnail", "zoom",
			"download", "print", "fullscreen", "multi_format",
		}
	}

	return result
}

// GetSupportedFormats 获取支持的文件格式
func (cm *ConfigManager) GetSupportedFormats() map[string][]string {
	formats := make(map[string][]string)

	switch cm.config.Provider {
	case "dify":
		formats["document"] = []string{"pdf", "txt", "md", "doc", "docx"}
		formats["image"] = []string{"jpg", "jpeg", "png", "gif", "bmp"}
		formats["total"] = append(formats["document"], formats["image"]...)
	case "kkfileview":
		formats["document"] = []string{
			"pdf", "doc", "docx", "xls", "xlsx", "ppt", "pptx",
			"txt", "md", "xml", "json", "csv",
		}
		formats["image"] = []string{
			"jpg", "jpeg", "png", "gif", "bmp", "tiff",
		}
		formats["video"] = []string{
			"mp4", "avi", "mov", "wmv", "flv", "mkv",
		}
		formats["audio"] = []string{
			"mp3", "wav", "aac", "flac",
		}
		formats["archive"] = []string{
			"zip", "rar", "7z", "tar", "gz",
		}
		formats["cad"] = []string{
			"dwg", "dxf",
		}
		formats["other"] = []string{
			"psd", "eps",
		}

		// 合并所有格式
		var all []string
		for _, formatList := range formats {
			all = append(all, formatList...)
		}
		formats["total"] = all
	}

	return formats
}

// IsFormatSupported 检查格式是否支持
func (cm *ConfigManager) IsFormatSupported(fileExtension string) bool {
	formats := cm.GetSupportedFormats()
	totalFormats := formats["total"]

	fileExtension = strings.ToLower(strings.TrimPrefix(fileExtension, "."))

	for _, format := range totalFormats {
		if format == fileExtension {
			return true
		}
	}

	return false
}

// GetPreviewCapabilities 获取预览能力
func (cm *ConfigManager) GetPreviewCapabilities() map[string]interface{} {
	capabilities := make(map[string]interface{})

	switch cm.config.Provider {
	case "dify":
		capabilities["provider"] = "dify"
		capabilities["type"] = "basic"
		capabilities["features"] = map[string]bool{
			"text_preview":  true,
			"image_preview": true,
			"pdf_preview":   true,
			"thumbnail":     false,
			"zoom":          false,
			"download":      true,
			"print":         false,
			"fullscreen":    false,
			"annotation":    false,
			"search":        false,
		}
	case "kkfileview":
		capabilities["provider"] = "kkfileview"
		capabilities["type"] = "enhanced"
		capabilities["features"] = map[string]bool{
			"text_preview":   true,
			"image_preview":  true,
			"pdf_preview":    true,
			"office_preview": true,
			"video_preview":  true,
			"audio_preview":  true,
			"thumbnail":      true,
			"zoom":           true,
			"download":       true,
			"print":          true,
			"fullscreen":     true,
			"annotation":     false,
			"search":         true,
			"watermark":      true,
		}
	}

	return capabilities
}
