# PDF处理功能使用指南

## 概述

CDK-Office集成了Stirling PDF处理功能，为用户提供强大的PDF文档处理能力。该功能支持PDF合并、拆分、压缩、旋转、添加水印、格式转换等多种操作。

## 功能特性

### 支持的PDF操作

1. **PDF合并** - 将多个PDF文件合并为一个文件
2. **PDF拆分** - 将一个PDF文件拆分为多个文件
3. **PDF压缩** - 压缩PDF文件以减小文件大小
4. **PDF旋转** - 旋转PDF页面（90°、180°、270°）
5. **添加水印** - 为PDF文件添加文字水印
6. **格式转换** - 将其他格式文件转换为PDF

### 技术架构

- **后端服务**: Go语言实现的PDF处理API
- **PDF引擎**: 集成Stirling PDF开源项目
- **数据库**: PostgreSQL存储任务记录
- **缓存**: Redis缓存处理结果
- **容器化**: Docker部署，易于扩展

## 部署说明

### 开发环境部署

1. **启动基础服务**
```bash
# 启动PostgreSQL、Redis、Stirling PDF
docker-compose -f docker-compose.dev.yml up -d

# 检查服务状态
docker-compose -f docker-compose.dev.yml ps
```

2. **服务地址**
- CDK-Office: http://localhost:8000
- Stirling PDF: http://localhost:8081
- PostgreSQL: localhost:5433
- Redis: localhost:6380

### 生产环境部署

1. **完整部署**
```bash
# 启动所有服务
docker-compose up -d

# 启用文件预览功能（可选）
docker-compose --profile file-preview up -d

# 启用监控功能（可选）
docker-compose --profile monitoring up -d
```

2. **服务地址**
- CDK-Office: http://localhost:8000
- Stirling PDF: http://localhost:8081
- KKFileView: http://localhost:8012
- Nginx: http://localhost
- Prometheus: http://localhost:9090
- Grafana: http://localhost:3000

## API接口使用

### 1. PDF合并

```bash
curl -X POST "http://localhost:8000/api/v1/pdf/merge" \
  -H "Content-Type: multipart/form-data" \
  -F "files=@file1.pdf" \
  -F "files=@file2.pdf" \
  -H "Authorization: Bearer YOUR_TOKEN"
```

### 2. PDF拆分

```bash
curl -X POST "http://localhost:8000/api/v1/pdf/split" \
  -H "Content-Type: multipart/form-data" \
  -F "file=@document.pdf" \
  -F "pages=1-3,5,7-9" \
  -H "Authorization: Bearer YOUR_TOKEN"
```

### 3. PDF压缩

```bash
curl -X POST "http://localhost:8000/api/v1/pdf/compress" \
  -H "Content-Type: multipart/form-data" \
  -F "file=@document.pdf" \
  -F "quality=medium" \
  -H "Authorization: Bearer YOUR_TOKEN"
```

### 4. PDF旋转

```bash
curl -X POST "http://localhost:8000/api/v1/pdf/rotate" \
  -H "Content-Type: multipart/form-data" \
  -F "file=@document.pdf" \
  -F "angle=90" \
  -F "pages=1-5" \
  -H "Authorization: Bearer YOUR_TOKEN"
```

### 5. 添加水印

```bash
curl -X POST "http://localhost:8000/api/v1/pdf/watermark" \
  -H "Content-Type: multipart/form-data" \
  -F "file=@document.pdf" \
  -F "text=机密文件" \
  -F "opacity=0.5" \
  -F "fontSize=36" \
  -F "position=center" \
  -H "Authorization: Bearer YOUR_TOKEN"
```

### 6. 格式转换

```bash
curl -X POST "http://localhost:8000/api/v1/pdf/convert" \
  -H "Content-Type: multipart/form-data" \
  -F "file=@document.docx" \
  -H "Authorization: Bearer YOUR_TOKEN"
```

### 7. 查看任务状态

```bash
# 获取任务列表
curl -X GET "http://localhost:8000/api/v1/pdf/tasks" \
  -H "Authorization: Bearer YOUR_TOKEN"

# 获取特定任务状态
curl -X GET "http://localhost:8000/api/v1/pdf/tasks/{task_id}" \
  -H "Authorization: Bearer YOUR_TOKEN"
```

### 8. 健康检查

```bash
curl -X GET "http://localhost:8000/api/v1/pdf/health"
```

## 配置说明

### 环境变量

```bash
# Stirling PDF服务地址
STIRLING_PDF_URL=http://stirling-pdf:8080

# PDF服务配置
PDF_ENABLED=true
PDF_TIMEOUT=120
PDF_MAX_FILE_SIZE=104857600  # 100MB
```

### 配置文件

```yaml
# config.yaml
pdf:
  enabled: true
  stirling_pdf_url: "http://stirling-pdf:8080"
  timeout: 120
  max_file_size: 104857600
  allowed_operations:
    - merge
    - split
    - compress
    - rotate
    - watermark
    - convert
```

## 性能优化

### 文件大小限制

- 默认最大文件大小: 100MB
- 可通过配置调整限制
- 建议根据服务器性能设置合理限制

### 并发处理

- 支持多任务并发处理
- 任务队列管理
- 自动重试机制

### 缓存策略

- Redis缓存处理结果
- 24小时自动清理
- 支持手动清理缓存

## 故障排除

### 常见问题

1. **Stirling PDF服务无法访问**
```bash
# 检查服务状态
docker logs cdk-office-stirling-pdf

# 重启服务
docker-compose restart stirling-pdf
```

2. **文件上传失败**
- 检查文件大小是否超出限制
- 确认文件格式是否支持
- 检查网络连接

3. **任务处理超时**
- 增加超时配置
- 检查服务器资源使用情况
- 优化文件大小

### 日志查看

```bash
# 查看CDK-Office日志
docker logs cdk-office

# 查看Stirling PDF日志
docker logs cdk-office-stirling-pdf

# 查看实时日志
docker-compose logs -f
```

## 安全注意事项

1. **文件访问控制**
- 确保用户只能访问自己的文件
- 实施适当的权限验证
- 定期清理临时文件

2. **服务隔离**
- Stirling PDF运行在独立容器中
- 网络隔离配置
- 防止恶意文件攻击

3. **数据保护**
- 敏感文件加密存储
- 安全的文件传输
- 合规的数据处理

## 扩展开发

### 添加新的PDF操作

1. 在`service.go`中添加新方法
2. 在`router.go`中添加对应路由
3. 更新数据模型和API文档
4. 编写单元测试

### 集成其他PDF工具

1. 实现新的PDF处理客户端
2. 添加配置选项
3. 实现服务降级机制
4. 更新文档

## 支持与反馈

如有问题或建议，请通过以下方式联系：

- 项目仓库: https://github.com/linux-do/cdk-office
- 问题反馈: 创建GitHub Issue
- 技术讨论: 参与社区讨论