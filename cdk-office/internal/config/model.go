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

package config

import "github.com/linux-do/cdk-office/internal/storage"

// Config 全局配置变量
var Config configModel

type configModel struct {
	App         appConfig             `mapstructure:"app"`
	Database    databaseConfig        `mapstructure:"database"`
	Redis       redisConfig           `mapstructure:"redis"`
	Log         logConfig             `mapstructure:"log"`
	Dify        difyConfig            `mapstructure:"dify"`
	OCR         ocrConfig             `mapstructure:"ocr"`
	WeChat      wechatConfig          `mapstructure:"wechat"`
	Schedule    scheduleConfig        `mapstructure:"schedule"`
	Worker      workerConfig          `mapstructure:"worker"`
	Contract    contractConfig        `mapstructure:"contract"`
	FileStorage fileStorageConfig     `mapstructure:"file_storage"`
	Storage     storage.StorageConfig `mapstructure:"storage"`
	Survey      surveyConfig          `mapstructure:"survey"`
}

// appConfig 应用基本配置
type appConfig struct {
	AppName           string `mapstructure:"app_name"`
	Env               string `mapstructure:"env"`
	Addr              string `mapstructure:"addr"`
	APIPrefix         string `mapstructure:"api_prefix"`
	SessionCookieName string `mapstructure:"session_cookie_name"`
	SessionSecret     string `mapstructure:"session_secret"`
	SessionDomain     string `mapstructure:"session_domain"`
	SessionAge        int    `mapstructure:"session_age"`
	SessionHttpOnly   bool   `mapstructure:"session_http_only"`
	SessionSecure     bool   `mapstructure:"session_secure"`
}

// databaseConfig 数据库配置
type databaseConfig struct {
	// 基础连接配置
	Provider string `mapstructure:"provider"` // local_postgres, supabase, memfire, neon, planetscale
	Host     string `mapstructure:"host"`
	Port     int    `mapstructure:"port"`
	Username string `mapstructure:"username"`
	Password string `mapstructure:"password"`
	Database string `mapstructure:"database"`

	// SSL配置
	SSLMode     string `mapstructure:"ssl_mode"`      // disable, require, verify-ca, verify-full
	SSLCert     string `mapstructure:"ssl_cert"`      // SSL证书文件路径
	SSLKey      string `mapstructure:"ssl_key"`       // SSL私钥文件路径
	SSLRootCert string `mapstructure:"ssl_root_cert"` // SSL根证书文件路径

	// 连接池配置
	MaxIdleConn     int `mapstructure:"max_idle_conn"`
	MaxOpenConn     int `mapstructure:"max_open_conn"`
	ConnMaxLifetime int `mapstructure:"conn_max_lifetime"`
	ConnMaxIdleTime int `mapstructure:"conn_max_idle_time"` // 连接最大空闲时间

	// 日志和监控配置
	LogLevel           string `mapstructure:"log_level"`
	SlowQueryThreshold int    `mapstructure:"slow_query_threshold"` // 慢查询阈值(毫秒)
	EnableMetrics      bool   `mapstructure:"enable_metrics"`       // 启用指标监控

	// Supabase专用配置
	Supabase supabaseConfig `mapstructure:"supabase"`

	// MemFire Cloud专用配置
	MemFire memfireConfig `mapstructure:"memfire"`

	// 多数据库支持(未来扩展)
	ReadReplicas   []replicaConfig `mapstructure:"read_replicas"`   // 读片数据库配置
	EnableSharding bool            `mapstructure:"enable_sharding"` // 是否启用分片
}

// redisConfig Redis配置
type redisConfig struct {
	Host         string `mapstructure:"host"`
	Port         int    `mapstructure:"port"`
	Username     string `mapstructure:"username"`
	Password     string `mapstructure:"password"`
	DB           int    `mapstructure:"db"`
	PoolSize     int    `mapstructure:"pool_size"`
	MinIdleConn  int    `mapstructure:"min_idle_conn"`
	DialTimeout  int    `mapstructure:"dial_timeout"`
	ReadTimeout  int    `mapstructure:"read_timeout"`
	WriteTimeout int    `mapstructure:"write_timeout"`
}

// logConfig 日志配置
type logConfig struct {
	Level      string `mapstructure:"level"`
	Format     string `mapstructure:"format"`
	Output     string `mapstructure:"output"`
	FilePath   string `mapstructure:"file_path"`
	MaxSize    int    `mapstructure:"max_size"`
	MaxAge     int    `mapstructure:"max_age"`
	MaxBackups int    `mapstructure:"max_backups"`
	Compress   bool   `mapstructure:"compress"`
}

// difyConfig Dify AI服务配置
type difyConfig struct {
	APIKey             string `mapstructure:"api_key"`
	APIEndpoint        string `mapstructure:"api_endpoint"`
	ChatEndpoint       string `mapstructure:"chat_endpoint"`
	CompletionEndpoint string `mapstructure:"completion_endpoint"`
	DatasetsEndpoint   string `mapstructure:"datasets_endpoint"`
	DocumentsEndpoint  string `mapstructure:"documents_endpoint"`
}

// ocrProviderConfig OCR服务商配置
type ocrProviderConfig struct {
	APIKey    string `mapstructure:"api_key"`
	SecretKey string `mapstructure:"secret_key"`
	SecretID  string `mapstructure:"secret_id"`
	Region    string `mapstructure:"region"`
	Endpoint  string `mapstructure:"endpoint"`
}

// ocrConfig OCR服务配置
type ocrConfig struct {
	DefaultProvider string                       `mapstructure:"default_provider"`
	Providers       map[string]ocrProviderConfig `mapstructure:"providers"`
}

// wechatConfig 微信配置
type wechatConfig struct {
	AppID             string `mapstructure:"app_id"`
	AppSecret         string `mapstructure:"app_secret"`
	MiniProgramAppID  string `mapstructure:"mini_program_app_id"`
	MiniProgramSecret string `mapstructure:"mini_program_secret"`
}

// scheduleConfig 定时任务配置
type scheduleConfig struct {
	DocumentSyncCron     string `mapstructure:"document_sync_cron"`
	HealthCheckCron      string `mapstructure:"health_check_cron"`
	ArchiveCron          string `mapstructure:"archive_cron"`
	CalendarReminderCron string `mapstructure:"calendar_reminder_cron"` // 日程提醒任务
}

// workerConfig 工作配置
type workerConfig struct {
	Concurrency int `mapstructure:"concurrency"`
}

// contractConfig 合同服务配置
type contractConfig struct {
	CAProvider         string `mapstructure:"ca_provider"`          // CA证书服务商
	SMSProvider        string `mapstructure:"sms_provider"`         // 短信服务商
	BlockchainEnabled  bool   `mapstructure:"blockchain_enabled"`   // 是否启用区块链存证
	DefaultExpireHours int    `mapstructure:"default_expire_hours"` // 默认过期时间(小时)
	MaxSigners         int    `mapstructure:"max_signers"`          // 最大签署人数
	AutoArchive        bool   `mapstructure:"auto_archive"`         // 是否自动归档
}

// fileStorageConfig 文件存储配置
type fileStorageConfig struct {
	Provider    string `mapstructure:"provider"`      // local, oss, cos, s3
	BasePath    string `mapstructure:"base_path"`     // 基础路径
	MaxFileSize int64  `mapstructure:"max_file_size"` // 最大文件大小(MB)

	// 本地存储配置
	LocalPath string `mapstructure:"local_path"`

	// 阿里云OSS配置
	OSSEndpoint        string `mapstructure:"oss_endpoint"`
	OSSAccessKeyID     string `mapstructure:"oss_access_key_id"`
	OSSAccessKeySecret string `mapstructure:"oss_access_key_secret"`
	OSSBucket          string `mapstructure:"oss_bucket"`

	// 腾讯云COS配置
	COSRegion    string `mapstructure:"cos_region"`
	COSSecretID  string `mapstructure:"cos_secret_id"`
	COSSecretKey string `mapstructure:"cos_secret_key"`
	COSBucket    string `mapstructure:"cos_bucket"`
}

// surveyConfig 问卷调查配置
type surveyConfig struct {
	EnablePublicAccess      bool     `mapstructure:"enable_public_access"`      // 是否允许公开访问问卷
	MaxResponsePerSurvey    int      `mapstructure:"max_response_per_survey"`   // 每个问卷最大响应数
	DefaultExpireDays       int      `mapstructure:"default_expire_days"`       // 默认问卷过期天数
	EnableAnonymousResponse bool     `mapstructure:"enable_anonymous_response"` // 是否允许匿名响应
	AutoAnalysis            bool     `mapstructure:"auto_analysis"`             // 是否自动触发AI分析
	ExportFormats           []string `mapstructure:"export_formats"`            // 支持的导出格式
	TemplateCategories      []string `mapstructure:"template_categories"`       // 模板分类
}

// supabaseConfig Supabase专用配置
type supabaseConfig struct {
	URL            string `mapstructure:"url"`              // Supabase项目URL
	AnonKey        string `mapstructure:"anon_key"`         // 匿名密钥
	ServiceRoleKey string `mapstructure:"service_role_key"` // 服务角色密钥
	JWTSecret      string `mapstructure:"jwt_secret"`       // JWT签名密钥
	Region         string `mapstructure:"region"`           // 区域
	APIVersion     string `mapstructure:"api_version"`      // API版本

	// 连接配置
	PoolerURL string `mapstructure:"pooler_url"` // 连接池URL(用于高并发)
	DirectURL string `mapstructure:"direct_url"` // 直连接URL(用于迁移等)

	// 性能配置
	MaxConnections int `mapstructure:"max_connections"` // 最大连接数(适合免费层)
	Timeout        int `mapstructure:"timeout"`         // 连接超时(秒)

	// 实时数据配置
	EnableRealtime bool   `mapstructure:"enable_realtime"` // 是否启用实时数据同步
	RealtimeURL    string `mapstructure:"realtime_url"`    // 实时数据服务URL

	// 存储配置
	StorageBucket string `mapstructure:"storage_bucket"` // 默认存储桶
}

// replicaConfig 读副本数据库配置
type replicaConfig struct {
	Name     string `mapstructure:"name"` // 副本名称
	Host     string `mapstructure:"host"`
	Port     int    `mapstructure:"port"`
	Username string `mapstructure:"username"`
	Password string `mapstructure:"password"`
	Database string `mapstructure:"database"`
	SSLMode  string `mapstructure:"ssl_mode"`
	Weight   int    `mapstructure:"weight"`  // 负载均衡权重
	Enabled  bool   `mapstructure:"enabled"` // 是否启用
}

// memfireConfig MemFire Cloud专用配置
type memfireConfig struct {
	URL        string `mapstructure:"url"`         // MemFire项目URL
	APIKey     string `mapstructure:"api_key"`     // API密钥
	ServiceKey string `mapstructure:"service_key"` // 服务密钥
	JWTSecret  string `mapstructure:"jwt_secret"`  // JWT签名密钥
	Region     string `mapstructure:"region"`      // 区域(默认为cn-shanghai)
	APIVersion string `mapstructure:"api_version"` // API版本

	// 连接配置
	PoolerURL string `mapstructure:"pooler_url"` // 连接池URL(高并发场景)
	DirectURL string `mapstructure:"direct_url"` // 直连URL(迁移等)

	// 性能配置
	MaxConnections int `mapstructure:"max_connections"` // 最大连接数
	Timeout        int `mapstructure:"timeout"`         // 连接超时(秒)

	// 实时功能配置
	EnableRealtime bool   `mapstructure:"enable_realtime"` // 是否启用实时数据同步
	RealtimeURL    string `mapstructure:"realtime_url"`    // 实时数据服务URL

	// 存储配置
	StorageBucket string `mapstructure:"storage_bucket"` // 默认存储桶
	CDNURL        string `mapstructure:"cdn_url"`        // CDN加速URL

	// 中国特色配置
	EnableICP    bool   `mapstructure:"enable_icp"`    // 是否启用ICP备案模式
	CustomDomain string `mapstructure:"custom_domain"` // 自定义域名
	EnableHTTPS  bool   `mapstructure:"enable_https"`  // 强制HTTPS
}
