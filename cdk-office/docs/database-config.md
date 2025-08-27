# 数据库配置指南

CDK-Office 完整支持多种数据库提供商，包括 **MemFire Cloud** 和 **Supabase**。

## 支持的数据库提供商

- `local_postgres` - 本地 PostgreSQL 数据库
- `supabase` - Supabase 云数据库服务
- `memfire` - MemFire Cloud（中国版 Supabase）
- `neon` - Neon 数据库（未来支持）
- `planetscale` - PlanetScale 数据库（未来支持）

## Supabase 配置

Supabase 是一个开源的 Firebase 替代品，提供实时数据库、认证、即时API等功能。

### 基础配置

```yaml
database:
  provider: "supabase"
  host: "db.your-project.supabase.co"
  port: 5432
  username: "postgres"
  password: "your-database-password"
  database: "postgres"
  ssl_mode: "require"
  
  # Supabase专用配置
  supabase:
    url: "https://your-project.supabase.co"
    anon_key: "your-anon-key"
    service_role_key: "your-service-role-key"
    jwt_secret: "your-jwt-secret"
    region: "us-east-1"
    
    # 连接优化
    pooler_url: "postgresql://postgres.your-project:password@aws-0-us-east-1.pooler.supabase.com:5432/postgres"
    direct_url: "postgresql://postgres.your-project:password@db.your-project.supabase.co:5432/postgres"
    
    # 性能配置（免费层优化）
    max_connections: 10
    timeout: 30
    
    # 实时功能
    enable_realtime: false
    storage_bucket: "cdk-office"
```

### 特点

- 免费层提供 500MB 存储空间
- 最大连接数限制为 10（免费层）
- 提供实时数据同步功能
- 全球CDN加速

## MemFire Cloud 配置

MemFire Cloud 是国内的云数据库服务，基于 Supabase 技术，专为中国用户优化。

### 基础配置

```yaml
database:
  provider: "memfire"
  host: "db.memfiredb.com"
  port: 5432
  username: "postgres"  
  password: "your-database-password"
  database: "postgres"
  ssl_mode: "require"
  
  # MemFire Cloud专用配置
  memfire:
    url: "https://your-project.memfiredb.com"
    api_key: "your-api-key"
    service_key: "your-service-key"
    jwt_secret: "your-jwt-secret"
    region: "cn-shanghai"
    
    # 连接优化
    pooler_url: "postgresql://postgres.your-project:password@pooler.memfiredb.com:5432/postgres"
    direct_url: "postgresql://postgres.your-project:password@db.memfiredb.com:5432/postgres"
    
    # 性能配置
    max_connections: 20
    timeout: 30
    
    # 实时功能
    enable_realtime: false
    storage_bucket: "cdk-office"
    cdn_url: "https://cdn.memfiredb.com"
    
    # 中国特色配置
    enable_icp: true
    custom_domain: "your-domain.com"
    enable_https: true
```

### 特点

- 专为中国用户优化的网络连接
- 支持ICP备案模式
- 提供CDN加速服务
- 更高的连接数限制
- 中国时区配置（Asia/Shanghai）

## 连接池优化

系统会根据不同的数据库提供商自动优化连接池配置：

### Supabase 优化

- 最大连接数限制为 10（适配免费层）
- 空闲连接数限制为 5
- 标准连接生命周期

### MemFire Cloud 优化

- 最大连接数可配置（默认20）
- 自动调整空闲连接数为最大连接数的一半
- 延长连接生命周期（30分钟）适应中国网络环境
- 使用中国时区配置

### 本地 PostgreSQL

- 灵活的连接数配置
- 默认禁用SSL（开发环境）
- 亚洲/上海时区

## 配置示例

### 开发环境

```yaml
database:
  provider: "local_postgres"
  host: "127.0.0.1"
  port: 5432
  username: "postgres"
  password: ""
  database: "cdk_office"
  ssl_mode: "disable"
  max_idle_conn: 10
  max_open_conn: 100
```

### 生产环境（Supabase）

```yaml
database:
  provider: "supabase"
  ssl_mode: "require"
  max_idle_conn: 5
  max_open_conn: 10
  
  supabase:
    url: "${SUPABASE_URL}"
    anon_key: "${SUPABASE_ANON_KEY}"
    service_role_key: "${SUPABASE_SERVICE_ROLE_KEY}"
    pooler_url: "${SUPABASE_POOLER_URL}"
    max_connections: 10
```

### 生产环境（MemFire Cloud）

```yaml
database:
  provider: "memfire"
  ssl_mode: "require"
  max_idle_conn: 10
  max_open_conn: 20
  
  memfire:
    url: "${MEMFIRE_URL}"
    api_key: "${MEMFIRE_API_KEY}"
    service_key: "${MEMFIRE_SERVICE_KEY}"
    pooler_url: "${MEMFIRE_POOLER_URL}"
    max_connections: 20
    enable_icp: true
```

## 环境变量

推荐在生产环境中使用环境变量来配置敏感信息：

```bash
# Supabase
export SUPABASE_URL="https://your-project.supabase.co"
export SUPABASE_ANON_KEY="your-anon-key"
export SUPABASE_SERVICE_ROLE_KEY="your-service-role-key"
export SUPABASE_JWT_SECRET="your-jwt-secret"

# MemFire Cloud
export MEMFIRE_URL="https://your-project.memfiredb.com"
export MEMFIRE_API_KEY="your-api-key"
export MEMFIRE_SERVICE_KEY="your-service-key"
export MEMFIRE_JWT_SECRET="your-jwt-secret"
```

## 注意事项

1. **连接数限制**：云数据库通常有连接数限制，请根据套餐选择合适的配置
2. **SSL连接**：生产环境强烈建议启用SSL连接
3. **连接池**：合理配置连接池大小以平衡性能和资源消耗
4. **时区设置**：MemFire Cloud 默认使用中国时区，Supabase 使用UTC时区
5. **网络延迟**：MemFire Cloud 在中国网络环境下延迟更低

## 故障排查

### 常见问题

1. **连接超时**：检查网络连接和防火墙设置
2. **SSL错误**：确认SSL模式配置正确
3. **认证失败**：检查用户名、密码和API密钥
4. **连接数超限**：调整连接池配置或升级套餐

### 日志配置

```yaml
database:
  log_level: "info"  # debug, info, warn, error, silent
  slow_query_threshold: 200  # 慢查询阈值(毫秒)
  enable_metrics: true       # 启用监控指标
```

通过以上配置，CDK-Office 可以无缝支持 MemFire Cloud 和 Supabase，为不同地区和需求的用户提供最佳的数据库连接体验。