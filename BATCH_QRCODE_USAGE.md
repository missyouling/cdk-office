# 批量二维码生成功能使用指南

## 概述

批量二维码生成功能允许用户一次性创建多个二维码，这对于需要大量二维码的场景非常有用。

## 功能特性

1. 批量创建多个二维码
2. 支持自定义URL模板
3. 支持前缀命名
4. 支持静态和动态二维码类型
5. 支持配置参数
6. 异步生成二维码图像
7. 状态跟踪（待处理、生成中、已完成、失败）

## API使用

### 创建批量二维码批次

发送POST请求到 `/api/app/batch-qrcodes` 端点：

```json
{
  "app_id": "your_app_id",
  "name": "Batch Name",
  "description": "Batch description",
  "prefix": "ticket",
  "count": 100,
  "type": "static",
  "url_template": "https://example.com/ticket/{index}",
  "created_by": "user_id"
}
```

### 生成批量二维码

发送POST请求到 `/api/app/batch-qrcodes/{batch_id}/generate` 端点来开始生成二维码。

### 查询批次状态

发送GET请求到 `/api/app/batch-qrcodes/{batch_id}` 端点来查询批次状态。

## 使用示例

### 创建批次

```bash
curl -X POST http://localhost:8080/api/app/batch-qrcodes \
  -H "Content-Type: application/json" \
  -d '{
    "app_id": "app_123",
    "name": "Event Tickets",
    "description": "QR codes for event tickets",
    "prefix": "ticket",
    "count": 100,
    "type": "static",
    "url_template": "https://example.com/ticket/{index}",
    "created_by": "user_123"
  }'
```

### 生成二维码

```bash
curl -X POST http://localhost:8080/api/app/batch-qrcodes/batch_123/generate
```

## 注意事项

1. 单次批量生成最多支持10000个二维码
2. 二维码图像生成是异步过程
3. 如果生成过程中出现错误，批次状态将标记为"失败"
4. 已完成的批次不能重新生成