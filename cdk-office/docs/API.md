# CDK-Office API 文档

## 概述

CDK-Office 是一个企业级办公系统，提供了丰富的API接口用于各种办公场景。本文档详细描述了系统的所有API接口。

## 基础信息

- **Base URL**: `https://your-domain.com/api/v1`
- **认证方式**: Bearer Token
- **数据格式**: JSON
- **字符编码**: UTF-8

## 认证

所有API请求都需要在Header中包含认证信息：

```http
Authorization: Bearer YOUR_TOKEN_HERE
X-User-ID: user_id
X-Team-ID: team_id
```

## 通用响应格式

### 成功响应

```json
{
  "code": 200,
  "message": "success",
  "data": {...},
  "timestamp": "2024-01-20T10:30:00Z"
}
```

### 错误响应

```json
{
  "code": 400,
  "message": "error message",
  "error": "detailed error description",
  "timestamp": "2024-01-20T10:30:00Z"
}
```

## API 接口

### 1. 数据隔离 API

#### 1.1 创建团队隔离策略

- **URL**: `/isolation/policies`
- **Method**: `POST`
- **描述**: 创建团队数据隔离策略

**请求参数**:
```json
{
  "team_id": 1,
  "isolation_level": "strict",
  "allow_cross_team_access": false,
  "data_sharing_rules": {},
  "access_control_rules": {},
  "audit_level": "full",
  "retention_period_days": 365,
  "enable_data_masking": true
}
```

**响应示例**:
```json
{
  "code": 200,
  "message": "success",
  "data": {
    "id": 1,
    "team_id": 1,
    "isolation_level": "strict",
    "created_at": "2024-01-20T10:30:00Z"
  }
}
```

#### 1.2 检查数据访问权限

- **URL**: `/isolation/check`
- **Method**: `GET`
- **描述**: 检查用户对特定资源的访问权限

**查询参数**:
- `resource_type`: 资源类型 (string, required)
- `resource_id`: 资源ID (string, required)
- `operation`: 操作类型 (string, required)

**响应示例**:
```json
{
  "code": 200,
  "message": "success",
  "data": {
    "allowed": true,
    "reason": "User has permission to read document",
    "restrictions": []
  }
}
```

#### 1.3 获取审计日志

- **URL**: `/isolation/audit-logs`
- **Method**: `GET`
- **描述**: 获取数据访问审计日志

**查询参数**:
- `page`: 页码 (int, default: 1)
- `limit`: 每页数量 (int, default: 20)
- `start_date`: 开始日期 (string, optional)
- `end_date`: 结束日期 (string, optional)

### 2. 系统优化 API

#### 2.1 获取优化模块状态

- **URL**: `/optimization/status`
- **Method**: `GET`
- **描述**: 获取系统优化模块的运行状态

**响应示例**:
```json
{
  "code": 200,
  "message": "success",
  "data": {
    "initialized": true,
    "components": {
      "circuit_breaker": true,
      "performance_monitor": true,
      "cache_optimizer": true,
      "rate_limit_manager": true
    }
  }
}
```

#### 2.2 获取性能指标

- **URL**: `/optimization/performance/metrics`
- **Method**: `GET`
- **描述**: 获取系统性能指标

**响应示例**:
```json
{
  "code": 200,
  "message": "success",
  "data": {
    "cpu": 0.45,
    "memory": 0.62,
    "total_requests": 15420,
    "avg_response_time": 245,
    "error_rate": 0.01
  }
}
```

#### 2.3 获取熔断器状态

- **URL**: `/optimization/circuit-breaker/stats`
- **Method**: `GET`
- **描述**: 获取熔断器统计信息

#### 2.4 重置熔断器

- **URL**: `/optimization/circuit-breaker/reset`
- **Method**: `POST`
- **描述**: 重置指定服务的熔断器

**查询参数**:
- `service`: 服务名称 (string, required)

#### 2.5 获取限流统计

- **URL**: `/optimization/rate-limit/stats`
- **Method**: `GET`
- **描述**: 获取API限流统计信息

### 3. 审批流程 API

#### 3.1 工作流定义管理

##### 创建工作流定义

- **URL**: `/workflow/definitions`
- **Method**: `POST`
- **描述**: 创建新的工作流定义

**请求参数**:
```json
{
  "name": "Leave Application",
  "description": "Employee leave application workflow",
  "version": "1.0",
  "definition": {
    "steps": [
      {
        "id": "start",
        "type": "start",
        "name": "Start"
      }
    ],
    "flows": [
      {
        "from": "start",
        "to": "approve"
      }
    ]
  },
  "form_schema": {},
  "settings": {}
}
```

##### 获取工作流定义

- **URL**: `/workflow/definitions/{id}`
- **Method**: `GET`
- **描述**: 获取指定工作流定义详情

##### 更新工作流定义

- **URL**: `/workflow/definitions/{id}`
- **Method**: `PUT`
- **描述**: 更新工作流定义

##### 删除工作流定义

- **URL**: `/workflow/definitions/{id}`
- **Method**: `DELETE`
- **描述**: 删除工作流定义

##### 获取工作流定义列表

- **URL**: `/workflow/definitions`
- **Method**: `GET`
- **描述**: 获取工作流定义列表

**查询参数**:
- `page`: 页码 (int, default: 1)
- `limit`: 每页数量 (int, default: 20)
- `status`: 状态过滤 (string, optional)

#### 3.2 工作流实例管理

##### 启动工作流实例

- **URL**: `/workflow/instances`
- **Method**: `POST`
- **描述**: 启动新的工作流实例

**请求参数**:
```json
{
  "definition_id": 1,
  "name": "John's Leave Application",
  "variables": {
    "applicant": "John Doe",
    "leave_type": "annual",
    "days": 5
  }
}
```

##### 获取工作流实例

- **URL**: `/workflow/instances/{id}`
- **Method**: `GET`
- **描述**: 获取工作流实例详情

##### 暂停工作流实例

- **URL**: `/workflow/instances/{id}/suspend`
- **Method**: `POST`
- **描述**: 暂停工作流实例

##### 恢复工作流实例

- **URL**: `/workflow/instances/{id}/resume`
- **Method**: `POST`
- **描述**: 恢复暂停的工作流实例

##### 终止工作流实例

- **URL**: `/workflow/instances/{id}/terminate`
- **Method**: `POST`
- **描述**: 终止工作流实例

#### 3.3 任务管理

##### 获取用户任务

- **URL**: `/workflow/tasks/user`
- **Method**: `GET`
- **描述**: 获取当前用户的待办任务

**查询参数**:
- `status`: 任务状态 (string, optional)
- `page`: 页码 (int, default: 1)
- `limit`: 每页数量 (int, default: 20)

##### 完成任务

- **URL**: `/workflow/tasks/{id}/complete`
- **Method**: `POST`
- **描述**: 完成指定任务

**请求参数**:
```json
{
  "variables": {
    "approved": true,
    "comment": "Application approved"
  }
}
```

##### 委派任务

- **URL**: `/workflow/tasks/{id}/delegate`
- **Method**: `POST`
- **描述**: 将任务委派给其他用户

**请求参数**:
```json
{
  "assignee_id": "user_123",
  "comment": "Delegating to supervisor"
}
```

### 4. 知识库 API

#### 4.1 个人知识库管理

##### 创建知识库

- **URL**: `/knowledge/personal`
- **Method**: `POST`
- **描述**: 创建个人知识库

**请求参数**:
```json
{
  "name": "My Knowledge Base",
  "description": "Personal knowledge collection",
  "type": "document",
  "settings": {
    "auto_sync": true,
    "public": false
  }
}
```

##### 获取知识库列表

- **URL**: `/knowledge/personal`
- **Method**: `GET`
- **描述**: 获取用户的知识库列表

##### 上传文档

- **URL**: `/knowledge/documents/upload`
- **Method**: `POST`
- **Content-Type**: `multipart/form-data`
- **描述**: 上传文档到知识库

**表单参数**:
- `file`: 文件 (file, required)
- `kb_id`: 知识库ID (string, required)
- `category`: 分类 (string, optional)

#### 4.2 文档管理

##### 搜索文档

- **URL**: `/knowledge/documents/search`
- **Method**: `GET`
- **描述**: 搜索知识库文档

**查询参数**:
- `q`: 搜索关键词 (string, required)
- `kb_id`: 知识库ID (string, optional)
- `category`: 分类过滤 (string, optional)

#### 4.3 微信聊天记录

##### 上传聊天记录

- **URL**: `/knowledge/wechat-records`
- **Method**: `POST`
- **描述**: 上传微信聊天记录

**请求参数**:
```json
{
  "session_name": "工作群聊",
  "records": [
    {
      "sender_name": "张三",
      "message": "今天的会议改到下午2点",
      "timestamp": "2024-01-20T10:30:00Z"
    }
  ]
}
```

### 5. PDF处理 API

#### 5.1 PDF操作

##### 合并PDF

- **URL**: `/pdf/merge`
- **Method**: `POST`
- **Content-Type**: `multipart/form-data`
- **描述**: 合并多个PDF文件

**表单参数**:
- `files[]`: PDF文件列表 (file[], required)
- `output_name`: 输出文件名 (string, optional)

##### 拆分PDF

- **URL**: `/pdf/split`
- **Method**: `POST`
- **描述**: 拆分PDF文件

##### PDF转图片

- **URL**: `/pdf/to-images`
- **Method**: `POST`
- **描述**: 将PDF转换为图片

##### OCR识别

- **URL**: `/pdf/ocr`
- **Method**: `POST`
- **描述**: 对PDF进行OCR文字识别

#### 5.2 获取PDF工具分类

- **URL**: `/pdf/categories`
- **Method**: `GET`
- **描述**: 获取PDF工具分类列表

### 6. 文件预览 API

#### 6.1 文件预览

##### 预览文件

- **URL**: `/preview/file`
- **Method**: `POST`
- **描述**: 预览文件

**请求参数**:
```json
{
  "file_url": "https://example.com/document.pdf",
  "watermark": {
    "enabled": true,
    "text": "Confidential",
    "position": "center"
  }
}
```

##### 获取预览状态

- **URL**: `/preview/status/{session_id}`
- **Method**: `GET`
- **描述**: 获取预览会话状态

### 7. 通知推送 API

#### 7.1 发送通知

##### 发送单个通知

- **URL**: `/notification/send`
- **Method**: `POST`
- **描述**: 发送单个通知

**请求参数**:
```json
{
  "recipient_id": "user_123",
  "title": "会议提醒",
  "message": "您有一个会议将在30分钟后开始",
  "channels": ["email", "wechat"],
  "priority": "high"
}
```

##### 批量发送通知

- **URL**: `/notification/batch-send`
- **Method**: `POST`
- **描述**: 批量发送通知

### 8. 配置管理 API

#### 8.1 获取配置

- **URL**: `/optimization/config/{key}`
- **Method**: `GET`
- **描述**: 获取配置项

#### 8.2 更新配置

- **URL**: `/optimization/config/{key}`
- **Method**: `PUT`
- **描述**: 更新配置项

**请求参数**:
```json
{
  "value": "new_value"
}
```

#### 8.3 获取配置分类

- **URL**: `/optimization/config/categories`
- **Method**: `GET`
- **描述**: 获取所有配置分类

### 9. 健康检查 API

#### 9.1 系统健康检查

- **URL**: `/health`
- **Method**: `GET`
- **描述**: 检查系统健康状态

**响应示例**:
```json
{
  "code": 200,
  "message": "success",
  "data": {
    "status": "healthy",
    "database": "connected",
    "redis": "connected",
    "services": {
      "pdf": "available",
      "preview": "available"
    },
    "uptime": "72h30m45s"
  }
}
```

#### 9.2 详细健康检查

- **URL**: `/health/detailed`
- **Method**: `GET`
- **描述**: 获取详细的健康检查信息

## 错误码说明

| 错误码 | 说明 |
|--------|------|
| 200 | 成功 |
| 400 | 请求参数错误 |
| 401 | 未认证 |
| 403 | 权限不足 |
| 404 | 资源不存在 |
| 409 | 资源冲突 |
| 422 | 数据验证失败 |
| 429 | 请求频率限制 |
| 500 | 服务器内部错误 |
| 502 | 网关错误 |
| 503 | 服务不可用 |

## 频率限制

为了保护系统稳定性，API实施了频率限制：

- **一般API**: 1000次/分钟
- **上传API**: 10次/分钟
- **批量操作**: 5次/分钟

当超过限制时，会返回HTTP状态码429。

## SDK和示例

### JavaScript/Node.js 示例

```javascript
const axios = require('axios');

const client = axios.create({
  baseURL: 'https://your-domain.com/api/v1',
  headers: {
    'Authorization': 'Bearer YOUR_TOKEN',
    'X-User-ID': 'user_id',
    'X-Team-ID': 'team_id'
  }
});

// 创建知识库
async function createKnowledgeBase() {
  try {
    const response = await client.post('/knowledge/personal', {
      name: 'My Knowledge Base',
      type: 'document'
    });
    console.log(response.data);
  } catch (error) {
    console.error(error.response.data);
  }
}
```

### Python 示例

```python
import requests

class CDKOfficeClient:
    def __init__(self, base_url, token, user_id, team_id):
        self.base_url = base_url
        self.headers = {
            'Authorization': f'Bearer {token}',
            'X-User-ID': user_id,
            'X-Team-ID': team_id,
            'Content-Type': 'application/json'
        }
    
    def create_workflow_instance(self, definition_id, name, variables):
        data = {
            'definition_id': definition_id,
            'name': name,
            'variables': variables
        }
        response = requests.post(
            f'{self.base_url}/workflow/instances',
            json=data,
            headers=self.headers
        )
        return response.json()

# 使用示例
client = CDKOfficeClient(
    base_url='https://your-domain.com/api/v1',
    token='your_token',
    user_id='user_id',
    team_id='team_id'
)

result = client.create_workflow_instance(
    definition_id=1,
    name='Test Workflow',
    variables={'key': 'value'}
)
```

### Go 示例

```go
package main

import (
    "bytes"
    "encoding/json"
    "fmt"
    "net/http"
)

type Client struct {
    BaseURL string
    Token   string
    UserID  string
    TeamID  string
}

func (c *Client) CreateKnowledgeBase(name, kbType string) error {
    data := map[string]interface{}{
        "name": name,
        "type": kbType,
    }
    
    jsonData, _ := json.Marshal(data)
    req, _ := http.NewRequest("POST", c.BaseURL+"/knowledge/personal", bytes.NewBuffer(jsonData))
    req.Header.Set("Authorization", "Bearer "+c.Token)
    req.Header.Set("X-User-ID", c.UserID)
    req.Header.Set("X-Team-ID", c.TeamID)
    req.Header.Set("Content-Type", "application/json")
    
    client := &http.Client{}
    resp, err := client.Do(req)
    if err != nil {
        return err
    }
    defer resp.Body.Close()
    
    fmt.Println("Status:", resp.Status)
    return nil
}
```

## 更新日志

### v1.0.0 (2024-01-20)
- 初始API版本发布
- 支持数据隔离、系统优化、审批流程等核心功能
- 完善的错误处理和频率限制

### 联系方式

如有API相关问题，请联系：
- 邮箱: api-support@your-domain.com
- 文档: https://docs.your-domain.com
- GitHub: https://github.com/your-org/cdk-office