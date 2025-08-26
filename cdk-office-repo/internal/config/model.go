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

// Config 全局配置变量
var Config configModel

type configModel struct {
	App      appConfig      `mapstructure:"app"`
	Database databaseConfig `mapstructure:"database"`
	Redis    redisConfig    `mapstructure:"redis"`
	Log      logConfig      `mapstructure:"log"`
	Dify     difyConfig     `mapstructure:"dify"`
	OCR      ocrConfig      `mapstructure:"ocr"`
	WeChat   wechatConfig   `mapstructure:"wechat"`
	Schedule scheduleConfig `mapstructure:"schedule"`
	Worker   workerConfig   `mapstructure:"worker"`
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
	Host            string `mapstructure:"host"`
	Port            int    `mapstructure:"port"`
	Username        string `mapstructure:"username"`
	Password        string `mapstructure:"password"`
	Database        string `mapstructure:"database"`
	MaxIdleConn     int    `mapstructure:"max_idle_conn"`
	MaxOpenConn     int    `mapstructure:"max_open_conn"`
	ConnMaxLifetime int    `mapstructure:"conn_max_lifetime"`
	LogLevel        string `mapstructure:"log_level"`
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
	DocumentSyncCron string `mapstructure:"document_sync_cron"`
	HealthCheckCron  string `mapstructure:"health_check_cron"`
	ArchiveCron      string `mapstructure:"archive_cron"`
}

// workerConfig 工作配置
type workerConfig struct {
	Concurrency int `mapstructure:"concurrency"`
}
