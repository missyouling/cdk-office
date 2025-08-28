# 文件预览功能使用指南

## 概述

CDK-Office集成了灵活的文件预览功能，支持两种预览模式：

1. **Dify原生预览** - 轻量级的基础预览功能
2. **KKFileView增强预览** - 功能丰富的专业预览服务

系统默认使用Dify原生预览，可根据需要配置启用KKFileView增强预览。

## 功能特性

### Dify原生预览

- **支持格式**: PDF、文本、Markdown、基础Office文档、图片
- **特点**: 轻量级、快速响应、低资源消耗
- **适用场景**: 基础文档浏览、资源受限环境

### KKFileView增强预览

- **支持格式**: 50+种文件格式
- **特点**: 功能丰富、专业预览、高兼容性
- **适用场景**: 专业办公环境、多格式文档处理

#### 支持的文件格式

| 类别 | 格式 |
|------|------|
| Office文档 | doc, docx, xls, xlsx, ppt, pptx |
| PDF文档 | pdf |
| 文本文件 | txt, md, xml, json, csv |
| 图片文件 | jpg, jpeg, png, gif, bmp, tiff |
| 视频文件 | mp4, avi, mov, wmv, flv, mkv |
| 音频文件 | mp3, wav, aac, flac |
| 压缩文件 | zip, rar, 7z, tar, gz |
| CAD文件 | dwg, dxf |
| 其他格式 | psd, eps |

## 配置说明

### 环境变量配置

```bash
# 文件预览提供者选择
FILE_PREVIEW_PROVIDER=dify  # 或 kkfileview

# Dify配置
DIFY_URL=http://dify-api:5001

# KKFileView配置
KKFILEVIEW_ENABLED=false
KKFILEVIEW_URL=http://kkfileview:8012
KKFILEVIEW_TIMEOUT=30
```

### 配置文件设置

```yaml
# config.yaml
file_preview:
  provider: dify  # dify 或 kkfileview
  dify_url: "http://dify-api:5001"
  kkfileview:
    enabled: false
    url: "http://kkfileview:8012"
    timeout: 30
```

### 动态配置切换

系统支持运行时配置切换，无需重启服务：

1. **启用KKFileView**:
   - 设置 `KKFILEVIEW_ENABLED=true`
   - 系统自动切换到KKFileView预览

2. **禁用KKFileView**:
   - 设置 `KKFILEVIEW_ENABLED=false`
   - 系统自动回退到Dify预览

## 部署指南

### 默认部署（仅Dify预览）

```bash
# 使用默认配置启动
docker-compose up -d
```

### 增强部署（启用KKFileView）

```bash
# 启用文件预览增强功能
docker-compose --profile file-preview up -d

# 或者修改环境变量
export KKFILEVIEW_ENABLED=true
docker-compose up -d
```

### Docker Compose配置

```yaml
# docker-compose.yml
services:
  # KKFileView服务（可选）
  kkfileview:
    image: keking/kkfileview:latest
    container_name: cdk-office-kkfileview
    ports:
      - "8012:8012"
    environment:
      - SERVER_PORT=8012
      - FILE_UPLOAD_ENABLED=true
      - FILE_UPLOAD_SIZE_MAX=104857600  # 100MB
    volumes:
      - ./kkfileview/files:/opt/kkFileView/files
      - ./kkfileview/logs:/opt/kkFileView/logs
    restart: unless-stopped
    profiles: ["file-preview"]
```

## API接口使用

### 1. 生成文件预览

```bash
curl -X POST "http://localhost:8000/api/v1/preview/generate" \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer YOUR_TOKEN" \
  -d '{
    "document_id": "doc-123",
    "file_name": "document.pdf",
    "file_url": "http://example.com/files/document.pdf",
    "file_type": "pdf"
  }'
```

**响应示例**:
```json
{
  "preview_url": "http://kkfileview:8012/onlinePreview?url=...",
  "thumbnail_url": "http://kkfileview:8012/picturesPreview?url=...",
  "provider": "kkfileview",
  "supported": true,
  "metadata": {
    "enhanced": true,
    "features": ["zoom", "download", "print", "fullscreen"]
  }
}
```

### 2. 获取预览URL

```bash
curl -X GET "http://localhost:8000/api/v1/preview/url/doc-123?fileType=pdf" \
  -H "Authorization: Bearer YOUR_TOKEN"
```

### 3. 检查文件支持

```bash
curl -X GET "http://localhost:8000/api/v1/preview/check/document.xlsx" \
  -H "Authorization: Bearer YOUR_TOKEN"
```

**响应示例**:
```json
{
  "filename": "document.xlsx",
  "file_type": "xlsx",
  "supported": true,
  "provider": "kkfileview"
}
```

### 4. 获取支持的文件类型

```bash
curl -X GET "http://localhost:8000/api/v1/preview/supported-types" \
  -H "Authorization: Bearer YOUR_TOKEN"
```

### 5. 查看预览历史

```bash
curl -X GET "http://localhost:8000/api/v1/preview/history?page=1&limit=10" \
  -H "Authorization: Bearer YOUR_TOKEN"
```

### 6. 健康检查

```bash
curl -X GET "http://localhost:8000/api/v1/preview/health"
```

### 7. 获取服务配置

```bash
curl -X GET "http://localhost:8000/api/v1/preview/config" \
  -H "Authorization: Bearer YOUR_TOKEN"
```

## 功能对比

| 功能 | Dify原生预览 | KKFileView增强预览 |
|------|-------------|-------------------|
| PDF预览 | ✅ | ✅ |
| Office文档 | 基础 | ✅ 完整支持 |
| 图片预览 | ✅ | ✅ |
| 视频播放 | ❌ | ✅ |
| 音频播放 | ❌ | ✅ |
| 缩略图 | ❌ | ✅ |
| 缩放功能 | ❌ | ✅ |
| 下载功能 | ✅ | ✅ |
| 打印功能 | ❌ | ✅ |
| 全屏预览 | ❌ | ✅ |
| 文字搜索 | ❌ | ✅ |
| 水印支持 | ❌ | ✅ |
| 资源消耗 | 低 | 中等 |
| 响应速度 | 快 | 中等 |

## 性能优化

### 缓存策略

1. **预览记录缓存** - 24小时本地缓存
2. **缩略图缓存** - 7天缓存时间
3. **配置缓存** - 内存缓存，实时更新

### 资源限制

```yaml
# 推荐配置
kkfileview:
  resources:
    limits:
      memory: 2Gi
      cpu: 1000m
    requests:
      memory: 1Gi
      cpu: 500m
```

### 文件大小限制

- **默认限制**: 100MB
- **Dify预览**: 50MB
- **KKFileView**: 100MB
- **可配置**: 根据服务器性能调整

## 故障排除

### 常见问题

1. **KKFileView服务无法访问**
```bash
# 检查服务状态
docker logs cdk-office-kkfileview

# 重启服务
docker-compose restart kkfileview
```

2. **预览生成失败**
```bash
# 检查文件格式支持
curl -X GET "http://localhost:8000/api/v1/preview/check/filename.ext"

# 检查服务健康状态
curl -X GET "http://localhost:8000/api/v1/preview/health"
```

3. **配置切换不生效**
```bash
# 重新加载配置
docker-compose restart cdk-office

# 检查环境变量
docker exec cdk-office env | grep KKFILEVIEW
```

### 日志分析

```bash
# 查看预览服务日志
docker logs cdk-office | grep "preview"

# 查看KKFileView日志
docker logs cdk-office-kkfileview

# 实时监控
docker-compose logs -f kkfileview
```

## 安全注意事项

1. **访问控制**
   - 确保用户只能预览有权限的文件
   - 实施适当的身份验证机制

2. **文件安全**
   - 验证文件来源的合法性
   - 防止恶意文件上传

3. **网络安全**
   - 使用HTTPS传输
   - 配置适当的CORS策略

4. **数据隐私**
   - 敏感文档加密存储
   - 定期清理预览缓存

## 扩展开发

### 添加新的预览提供者

1. 实现`PreviewProvider`接口
2. 在`ConfigManager`中添加配置选项
3. 更新路由和文档

### 自定义预览模板

1. 创建自定义预览组件
2. 配置模板映射关系
3. 实现预览参数传递

### 集成第三方服务

1. 评估服务兼容性
2. 实现适配器模式
3. 配置降级机制

## 监控与维护

### 性能指标

- 预览生成成功率
- 平均响应时间
- 错误率统计
- 资源使用情况

### 定期维护

```bash
# 清理预览历史（保留30天）
curl -X POST "http://localhost:8000/api/v1/preview/cleanup" \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer ADMIN_TOKEN" \
  -d '{"older_than_days": 30}'
```

### 升级指南

1. 备份配置和数据
2. 更新Docker镜像
3. 验证服务可用性
4. 恢复配置和数据

## 支持与反馈

如有问题或建议，请通过以下方式联系：

- 项目仓库: https://github.com/linux-do/cdk-office
- 问题反馈: 创建GitHub Issue
- 技术讨论: 参与社区讨论