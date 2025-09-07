# 批量二维码生成功能文档

## 概述

批量二维码生成功能允许用户一次性创建多个二维码，这对于需要大量二维码的场景非常有用，例如活动门票、产品标签、资产标识等。

## 功能特性

1. 批量创建多个二维码
2. 支持自定义URL模板
3. 支持前缀命名
4. 支持静态和动态二维码类型
5. 支持配置参数
6. 异步生成二维码图像
7. 状态跟踪（待处理、生成中、已完成、失败）

## API端点

### 创建批量二维码批次

```
POST /api/app/batch-qrcodes
```

**请求参数：**
- `app_id` (string, 必需): 应用ID
- `name` (string, 必需): 批次名称
- `description` (string, 可选): 批次描述
- `prefix` (string, 可选): 二维码名称前缀
- `count` (int, 必需): 生成二维码数量 (1-10000)
- `type` (string, 必需): 二维码类型 (static 或 dynamic)
- `url_template` (string, 可选): URL模板，使用`{index}`作为占位符
- `config` (map, 可选): 配置参数
- `created_by` (string, 必需): 创建者ID

**响应：**
```json
{
  "id": "batch_xxxxxxxxxxxxxxxx",
  "app_id": "app_xxxxxxxxxxxxxxxx",
  "name": "My Batch",
  "description": "My batch description",
  "prefix": "ticket",
  "count": 100,
  "type": "static",
  "url_template": "https://example.com/ticket/{index}",
  "config": "{}",
  "status": "pending",
  "created_by": "user_xxxxxxxxxxxxxxxx",
  "created_at": "2023-01-01T00:00:00Z",
  "updated_at": "2023-01-01T00:00:00Z"
}
```

### 更新批量二维码批次

```
PUT /api/app/batch-qrcodes/{id}
```

**请求参数：**
- `name` (string, 可选): 批次名称
- `description` (string, 可选): 批次描述
- `prefix` (string, 可选): 二维码名称前缀
- `url_template` (string, 可选): URL模板
- `config` (string, 可选): 配置参数

### 删除批量二维码批次

```
DELETE /api/app/batch-qrcodes/{id}
```

### 列出批量二维码批次

```
GET /api/app/batch-qrcodes?app_id={app_id}&page={page}&size={size}
```

**查询参数：**
- `app_id` (string, 必需): 应用ID
- `page` (int, 可选): 页码，默认为1
- `size` (int, 可选): 每页大小，默认为10，最大为100

### 获取批量二维码批次详情

```
GET /api/app/batch-qrcodes/{id}
```

### 生成批量二维码

```
POST /api/app/batch-qrcodes/{id}/generate
```

## 使用示例

### 创建批量二维码批次

```bash
curl -X POST http://localhost:8080/api/app/batch-qrcodes \
  -H "Content-Type: application/json" \
  -d '{
    "app_id": "app_20230101000000",
    "name": "Event Tickets",
    "description": "QR codes for event tickets",
    "prefix": "ticket",
    "count": 100,
    "type": "static",
    "url_template": "https://example.com/ticket/{index}",
    "created_by": "user_20230101000000"
  }'
```

### 生成批量二维码

```bash
curl -X POST http://localhost:8080/api/app/batch-qrcodes/batch_20230101000000/generate
```

### 列出批量二维码批次

```bash
curl "http://localhost:8080/api/app/batch-qrcodes?app_id=app_20230101000000&page=1&size=10"
```

## 数据模型

### BatchQRCode (批量二维码批次)

| 字段 | 类型 | 描述 |
|------|------|------|
| ID | string | 批次ID |
| AppID | string | 应用ID |
| Name | string | 批次名称 |
| Description | string | 批次描述 |
| Prefix | string | 二维码名称前缀 |
| Count | int | 二维码数量 |
| Type | string | 二维码类型 (static/dynamic) |
| URLTemplate | string | URL模板 |
| Config | string | 配置参数 |
| Status | string | 状态 (pending/generating/completed/failed) |
| CreatedBy | string | 创建者ID |
| CreatedAt | time.Time | 创建时间 |
| UpdatedAt | time.Time | 更新时间 |

### BatchQRCodeItem (批量二维码项)

| 字段 | 类型 | 描述 |
|------|------|------|
| ID | string | 二维码项ID |
| BatchID | string | 批次ID |
| Name | string | 二维码名称 |
| Content | string | 二维码内容 |
| URL | string | 二维码URL |
| ImagePath | string | 二维码图像路径 |
| Status | string | 状态 (pending/generating/completed/failed) |
| CreatedAt | time.Time | 创建时间 |
| UpdatedAt | time.Time | 更新时间 |

## 实现细节

### 服务层

批量二维码生成功能在 `batch_qrcode_service.go` 文件中实现，主要方法包括：
- `CreateBatchQRCode`: 创建批量二维码批次
- `UpdateBatchQRCode`: 更新批量二维码批次
- `DeleteBatchQRCode`: 删除批量二维码批次
- `ListBatchQRCodes`: 列出批量二维码批次
- `GetBatchQRCode`: 获取批量二维码批次详情
- `GenerateBatchQRCodes`: 生成批量二维码

### 处理层

API端点在 `batch_qrcode_handler.go` 文件中实现，处理HTTP请求并调用相应的服务方法。

### 生成过程

1. 用户创建批量二维码批次
2. 系统创建批次记录，状态为"待处理"
3. 用户调用生成接口
4. 系统更新批次状态为"生成中"
5. 系统逐个创建二维码记录
6. 系统为每个二维码生成图像
7. 系统更新批次状态为"已完成"或"失败"

## 注意事项

1. 单次批量生成最多支持10000个二维码
2. 二维码图像生成是异步过程
3. 如果生成过程中出现错误，批次状态将标记为"失败"
4. 已完成的批次不能重新生成