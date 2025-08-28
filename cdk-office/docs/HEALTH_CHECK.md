# 服务健康检查系统

CDK-Office 的服务健康检查系统提供了完整的服务监控和状态管理功能。

## 功能特性

- **全面的服务监控**：支持数据库、Redis、AI服务、OCR服务、微信API等多种服务类型
- **智能健康判断**：基于响应时间、状态码和模拟请求的综合健康评估
- **状态持久化**：将检查结果存储到PostgreSQL数据库
- **RESTful API**：提供完整的管理API接口
- **定期检查**：支持后台定期执行健康检查
- **清理功能**：自动清理过期的历史记录

## 核心组件

### 1. ServiceHealthChecker
主要的健康检查服务，负责执行检查逻辑和状态管理。

```go
// 创建健康检查器
healthChecker := service.NewServiceHealthChecker(db, logger)

// 检查所有服务
results, err := healthChecker.CheckAllServices(ctx)

// 获取服务状态
statuses, err := healthChecker.GetAllServiceStatuses(ctx)
```

### 2. ServiceStatus 模型
服务状态的数据模型，存储在数据库中。

```go
type ServiceStatus struct {
    ID           uuid.UUID `json:"id"`
    ServiceName  string    `json:"service_name"`
    Status       string    `json:"status"`        // healthy, unhealthy, degraded
    ResponseTime int64     `json:"response_time"` // 毫秒
    StatusCode   int       `json:"status_code"`
    ErrorMessage string    `json:"error_message"`
    Details      string    `json:"details"`       // JSON格式
    CheckedAt    time.Time `json:"checked_at"`
    CreatedAt    time.Time `json:"created_at"`
    UpdatedAt    time.Time `json:"updated_at"`
}
```

### 3. HealthCheckHandler
HTTP API处理器，提供REST接口。

## API 端点

### 获取所有服务状态
```http
GET /api/admin/service-status
Authorization: Bearer admin-token
```

**响应示例：**
```json
{
  "success": true,
  "data": {
    "services": [
      {
        "id": "123e4567-e89b-12d3-a456-426614174000",
        "service_name": "postgresql_database",
        "status": "healthy",
        "response_time": 50,
        "status_code": 200,
        "error_message": "",
        "details": "",
        "checked_at": "2025-01-20T12:00:00Z"
      }
    ],
    "summary": {
      "overall_status": "healthy",
      "total_services": 6,
      "healthy_count": 5,
      "degraded_count": 1,
      "unhealthy_count": 0,
      "avg_response_time_ms": 150
    }
  },
  "timestamp": "2025-01-20T12:00:00Z"
}
```

### 获取指定服务状态
```http
GET /api/admin/service-status/{service_name}
Authorization: Bearer admin-token
```

### 手动触发健康检查
```http
POST /api/admin/service-status/check
Authorization: Bearer admin-token
```

### 清理旧记录
```http
DELETE /api/admin/service-status/cleanup?days=30
Authorization: Bearer admin-token
```

### 获取健康摘要
```http
GET /api/admin/service-status/summary
Authorization: Bearer admin-token
```

## 公共健康检查端点

这些端点不需要认证，适用于负载均衡器和监控系统：

```http
GET /health          # 基础健康检查
GET /health/ready    # 就绪检查
GET /health/live     # 存活检查
```

## 支持的服务类型

1. **数据库服务** (`database`)
   - PostgreSQL连接检查
   - 简单SQL查询测试

2. **Redis缓存** (`redis`)
   - Redis连接检查
   - Ping操作测试

3. **AI服务** (`ai_service`)
   - HTTP端点可达性
   - API响应时间检查
   - 模拟请求测试

4. **OCR服务** (`ocr_service`)
   - 服务端点检查
   - 认证接口测试

5. **微信API** (`wechat_api`)
   - 微信接口可达性
   - Token获取测试

6. **存储服务** (`storage`)
   - Supabase存储检查
   - 存储桶访问测试

## 健康状态定义

- **healthy**: 服务正常，响应时间 < 500ms
- **degraded**: 服务可用但性能下降，响应时间 >= 500ms 或部分功能异常
- **unhealthy**: 服务不可用或出现错误

## 配置说明

### 服务配置
在 `ServiceHealthChecker.GetServiceConfigs()` 中配置需要监控的服务：

```go
ServiceConfig{
    Name:       "openai_service",
    Type:       "ai_service",
    Endpoint:   "https://api.openai.com",
    HealthPath: "/v1/models",
    Timeout:    10 * time.Second,
    Headers: map[string]string{
        "Authorization": "Bearer YOUR_API_KEY",
    },
    Critical: false,
}
```

### 数据库迁移
运行数据库迁移创建必要的表：

```sql
-- 执行迁移文件
psql -d your_database -f migrations/20250120_create_service_statuses_table.sql
```

## 使用示例

### 集成到现有应用

```go
package main

import (
    "github.com/linux-do/cdk-office/internal/router"
    "github.com/linux-do/cdk-office/internal/service"
)

func main() {
    // 初始化数据库和日志
    db := initDatabase()
    logger := initLogger()
    
    // 创建Gin路由
    r := gin.New()
    
    // 设置健康检查路由
    router.SetupHealthCheckRoutes(r, db, logger)
    
    // 启动定期健康检查
    router.SetupPeriodicHealthCheck(db, logger)
    
    // 启动服务器
    r.Run(":8080")
}
```

### 手动执行健康检查

```go
// 创建健康检查器
healthChecker := service.NewServiceHealthChecker(db, logger)

// 检查所有服务
ctx := context.Background()
results, err := healthChecker.CheckAllServices(ctx)
if err != nil {
    log.Printf("健康检查失败: %v", err)
    return
}

// 处理结果
for _, result := range results {
    log.Printf("服务 %s 状态: %s, 响应时间: %v", 
        result.ServiceName, result.Status, result.ResponseTime)
}
```

### 自定义服务检查

```go
// 添加自定义服务配置
customConfig := ServiceConfig{
    Name:       "custom_api",
    Type:       "ai_service",
    Endpoint:   "https://your-api.com",
    HealthPath: "/health",
    Timeout:    5 * time.Second,
    Critical:   true,
}

// 执行单个服务检查
result := healthChecker.performHealthCheck(ctx, customConfig)
```

## 监控和告警

### 监控指标

- 服务响应时间
- 服务可用性百分比
- 错误率统计
- 关键服务状态

### 告警建议

1. **关键服务不可用**：立即告警
2. **服务响应时间过长**：警告级别
3. **多个服务同时异常**：严重告警
4. **健康检查本身失败**：系统级告警

## 性能考虑

1. **检查频率**：默认5分钟，可根据需要调整
2. **超时设置**：不同服务类型设置不同的超时时间
3. **数据清理**：定期清理30天前的历史记录
4. **并发控制**：支持并发检查多个服务

## 故障排除

### 常见问题

1. **数据库连接失败**
   - 检查数据库连接字符串
   - 验证数据库权限

2. **服务检查超时**
   - 调整服务配置中的超时时间
   - 检查网络连接

3. **权限认证失败**
   - 验证API密钥配置
   - 检查服务端点URL

### 日志分析

健康检查系统使用结构化日志，方便分析：

```json
{
  "level": "info",
  "msg": "服务健康检查完成",
  "service_name": "postgresql_database",
  "status": "healthy",
  "response_time": "50ms",
  "timestamp": "2025-01-20T12:00:00Z"
}
```

## 扩展指南

### 添加新的服务类型

1. 在 `GetServiceConfigs()` 中添加新配置
2. 在 `performHealthCheck()` 中添加新的检查逻辑
3. 实现特定的检查方法

### 自定义健康判断逻辑

重写 `performHealthCheck()` 方法来实现自定义的健康判断逻辑。

### 集成外部监控系统

通过API端点可以轻松集成Prometheus、Grafana等监控系统。