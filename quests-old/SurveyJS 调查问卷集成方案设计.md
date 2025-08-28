基于您的技术栈（Go + Next.js + React）和项目需求，为您设计一套详细的 **SurveyJS 集成方案**，包含前后端对接细节、数据流设计和权限控制。

# 🧩 SurveyJS 集成方案设计

## 系统架构设计

```mermaid
graph TB
    A[前端 Next.js/React] -->|提交问卷数据| B[Go 后端 API]
    B -->|存储问卷数据| C[PostgreSQL]
    B -->|调用 Dify 工作流| D[Dify 平台]
    D -->|智能分析| E[返回分析结果]
    B -->|数据持久化| C
```

## 一、前端集成方案 (Next.js/React)

### 1. 安装依赖

```bash
# 安装 SurveyJS 核心库和 React UI
pnpm add survey-core survey-react-ui
# 安装类型定义（如果使用 TypeScript）
pnpm add -D @types/survey-core
```

### 2. 基础问卷组件封装

```typescript
// components/SurveyComponent.tsx
import React, { useState, useCallback } from 'react';
import { Model } from 'survey-core';
import { Survey } from 'survey-react-ui';
import 'survey-core/defaultV2.min.css';

interface SurveyComponentProps {
  surveyJson: any; // SurveyJS JSON 定义
  apiEndpoint: string; // 提交API地址
  onComplete?: (data: any) => void; // 完成回调
  theme?: 'defaultV2' | 'modern' | 'sharp'; // 主题选择
}

export const SurveyComponent: React.FC<SurveyComponentProps> = ({
  surveyJson,
  apiEndpoint,
  onComplete,
  theme = 'defaultV2'
}) => {
  const [surveyModel] useState(new Model(surveyJson));
  const [isSubmitting, setIsSubmitting] = useState(false);

  // 处理问卷完成
  const handleComplete = useCallback(async (sender: any) => {
    setIsSubmitting(true);
    try {
      const response = await fetch(apiEndpoint, {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
          'Authorization': `Bearer ${localStorage.getItem('token')}`
        },
        body: JSON.stringify({
          surveyData: sender.data,
          surveyId: surveyJson.surveyId,
          completedAt: new Date().toISOString()
        })
      });

      if (!response.ok) throw new Error('提交失败');
      
      const result = await response.json();
      onComplete?.(result);
      
    } catch (error) {
      console.error('问卷提交错误:', error);
      alert('提交失败，请重试');
    } finally {
      setIsSubmitting(false);
    }
  }, [apiEndpoint, onComplete, surveyJson.surveyId]);

  // 配置survey模型
  surveyModel.applyTheme(theme);
  surveyModel.onComplete.add(handleComplete);

  return (
    <div className="survey-container">
      <Survey model={surveyModel} />
      {isSubmitting && <div className="survey-loading">提交中...</div>}
    </div>
  );
};
```

### 3. 问卷JSON定义示例

```json
// surveys/employee-satisfaction.json
{
  "surveyId": "employee-satisfaction-2024",
  "title": "员工满意度调查",
  "description": "感谢您参与本次问卷调查，您的反馈将帮助我们改进工作环境",
  "pages": [
    {
      "name": "work-environment",
      "title": "工作环境",
      "elements": [
        {
          "type": "rating",
          "name": "workplace_rating",
          "title": "您如何评价当前的工作环境？",
          "rateMin": 1,
          "rateMax": 10,
          "minRateDescription": "很不满意",
          "maxRateDescription": "非常满意"
        },
        {
          "type": "comment",
          "name": "workplace_suggestions",
          "title": "您对工作环境有什么改进建议？",
          "placeholder": "请分享您的想法..."
        }
      ]
    }
  ],
  "completedHtml": "<h3>感谢您的参与！</h3><p>您的反馈对我们非常重要。</p>"
}
```

### 4. 页面集成示例

```typescript
// app/surveys/[surveyId]/page.tsx
import { SurveyComponent } from '@/components/SurveyComponent';
import { getSurveyDefinition } from '@/lib/surveyData';

interface SurveyPageProps {
  params: {
    surveyId: string;
  };
}

export default async function SurveyPage({ params }: SurveyPageProps) {
  // 获取问卷定义（可以从数据库或文件加载）
  const surveyJson = await getSurveyDefinition(params.surveyId);
  
  return (
    <div className="container mx-auto px-4 py-8">
      <SurveyComponent
        surveyJson={surveyJson}
        apiEndpoint="/api/surveys/submit"
        onComplete={(data) => {
          // 处理完成逻辑
          console.log('问卷完成:', data);
        }}
        theme="modern"
      />
    </div>
  );
}
```

## 二、后端API设计 (Go)

### 1. 数据模型定义

```go
// models/survey.go
package models

import (
    "time"
    "github.com/google/uuid"
)

// SurveyDefinition 问卷定义
type SurveyDefinition struct {
    ID          uuid.UUID `gorm:"type:uuid;primary_key" json:"id"`
    SurveyID    string    `gorm:"uniqueIndex" json:"survey_id"`
    Title       string    `gorm:"size:255" json:"title"`
    Description string    `gorm:"type:text" json:"description"`
    JsonDefinition map[string]interface{} `gorm:"type:jsonb" json:"json_definition"`
    Version     string    `gorm:"size:50" json:"version"`
    IsActive    bool      `gorm:"default:true" json:"is_active"`
    CreatedBy   uuid.UUID `gorm:"type:uuid" json:"created_by"`
    CreatedAt   time.Time `json:"created_at"`
    UpdatedAt   time.Time `json:"updated_at"`
}

// SurveyResponse 问卷响应
type SurveyResponse struct {
    ID          uuid.UUID `gorm:"type:uuid;primary_key" json:"id"`
    SurveyID    string    `gorm:"index" json:"survey_id"`
    UserID      uuid.UUID `gorm:"type:uuid;index" json:"user_id"`
    TeamID      uuid.UUID `gorm:"type:uuid;index" json:"team_id"`
    ResponseData map[string]interface{} `gorm:"type:jsonb" json:"response_data"`
    CompletedAt time.Time `json:"completed_at"`
    TimeSpent   int       `json:"time_spent"` // 完成耗时（秒）
    IPAddress   string    `gorm:"size:45" json:"ip_address"`
    UserAgent   string    `gorm:"type:text" json:"user_agent"`
    CreatedAt   time.Time `json:"created_at"`
}

// SurveyAnalysis 问卷分析结果
type SurveyAnalysis struct {
    ID          uuid.UUID `gorm:"type:uuid;primary_key" json:"id"`
    SurveyID    string    `gorm:"index" json:"survey_id"`
    ResponseID  uuid.UUID `gorm:"type:uuid;index" json:"response_id"`
    AnalysisData map[string]interface{} `gorm:"type:jsonb" json:"analysis_data"`
    DifyWorkflowID string `gorm:"size:100" json:"dify_workflow_id"`
    CreatedAt   time.Time `json:"created_at"`
}
```

### 2. API路由设计

```go
// routes/survey.go
package routes

import (
    "net/http"
    "github.com/gin-gonic/gin"
    "github.com/google/uuid"
    "your-app/models"
    "your-app/middleware"
)

func RegisterSurveyRoutes(router *gin.Engine) {
    group := router.Group("/api/surveys")
    {
        group.GET("/:surveyId", middleware.Auth(), getSurveyDefinition)
        group.POST("/submit", middleware.Auth(), submitSurveyResponse)
        group.GET("/responses/:surveyId", middleware.Auth(), getSurveyResponses)
        group.GET("/analysis/:responseId", middleware.Auth(), getSurveyAnalysis)
    }
}

// 获取问卷定义
func getSurveyDefinition(c *gin.Context) {
    surveyId := c.Param("surveyId")
    
    var survey models.SurveyDefinition
    if err := db.Where("survey_id = ? AND is_active = true", surveyId).First(&survey).Error; err != nil {
        c.JSON(http.StatusNotFound, gin.H{"error": "问卷不存在"})
        return
    }
    
    c.JSON(http.StatusOK, survey.JsonDefinition)
}

// 提交问卷响应
func submitSurveyResponse(c *gin.Context) {
    var request struct {
        SurveyID    string         `json:"surveyId" binding:"required"`
        SurveyData  map[string]interface{} `json:"surveyData" binding:"required"`
        CompletedAt string         `json:"completedAt"`
        TimeSpent   int            `json:"timeSpent"`
    }
    
    if err := c.ShouldBindJSON(&request); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": "无效的请求数据"})
        return
    }
    
    // 获取用户信息
    userID, _ := uuid.Parse(c.GetString("user_id"))
    teamID, _ := uuid.Parse(c.GetString("team_id"))
    
    // 保存响应数据
    response := models.SurveyResponse{
        ID:          uuid.New(),
        SurveyID:    request.SurveyID,
        UserID:      userID,
        TeamID:      teamID,
        ResponseData: request.SurveyData,
        CompletedAt: time.Now(),
        TimeSpent:   request.TimeSpent,
        IPAddress:   c.ClientIP(),
        UserAgent:   c.Request.UserAgent(),
        CreatedAt:   time.Now(),
    }
    
    if err := db.Create(&response).Error; err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": "保存失败"})
        return
    }
    
    // 触发Dify工作流进行智能分析（异步）
    go triggerDifyAnalysis(response)
    
    c.JSON(http.StatusOK, gin.H{
        "message": "提交成功",
        "responseId": response.ID,
    })
}

// 获取问卷响应列表
func getSurveyResponses(c *gin.Context) {
    surveyId := c.Param("surveyId")
    teamID := c.GetString("team_id")
    
    var responses []models.SurveyResponse
    if err := db.Where("survey_id = ? AND team_id = ?", surveyId, teamID).
        Order("created_at DESC").
        Find(&responses).Error; err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": "获取数据失败"})
        return
    }
    
    c.JSON(http.StatusOK, responses)
}
```

### 3. Dify工作流集成

```go
// services/dify_service.go
package services

import (
    "encoding/json"
    "fmt"
    "net/http"
    "time"
    "your-app/models"
)

// triggerDifyAnalysis 触发Dify工作流进行问卷分析
func triggerDifyAnalysis(response models.SurveyResponse) error {
    // 准备Dify请求数据
    requestData := map[string]interface{}{
        "inputs": map[string]interface{}{
            "survey_data": response.ResponseData,
            "survey_id":   response.SurveyID,
            "user_id":     response.UserID.String(),
        },
        "response_mode": "blocking",
        "user":         response.UserID.String(),
    }
    
    // 转换为JSON
    jsonData, err := json.Marshal(requestData)
    if err != nil {
        return err
    }
    
    // 创建HTTP请求
    req, err := http.NewRequest("POST", "https://api.dify.ai/v1/workflows/run", bytes.NewBuffer(jsonData))
    if err != nil {
        return err
    }
    
    req.Header.Set("Authorization", "Bearer "+os.Getenv("DIFY_API_KEY"))
    req.Header.Set("Content-Type", "application/json")
    
    // 发送请求
    client := &http.Client{Timeout: 30 * time.Second}
    resp, err := client.Do(req)
    if err != nil {
        return err
    }
    defer resp.Body.Close()
    
    // 解析响应
    var result struct {
        Data struct {
            Outputs map[string]interface{} `json:"outputs"`
            WorkflowID string `json:"workflow_id"`
        } `json:"data"`
    }
    
    if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
        return err
    }
    
    // 保存分析结果
    analysis := models.SurveyAnalysis{
        ID:          uuid.New(),
        SurveyID:    response.SurveyID,
        ResponseID:  response.ID,
        AnalysisData: result.Data.Outputs,
        DifyWorkflowID: result.Data.WorkflowID,
        CreatedAt:   time.Now(),
    }
    
    return db.Create(&analysis).Error
}

// GetSurveyStats 获取问卷统计信息
func GetSurveyStats(surveyID string, teamID uuid.UUID) (map[string]interface{}, error) {
    var stats struct {
        TotalResponses  int64 `json:"total_responses"`
        AverageTimeSpent float64 `json:"average_time_spent"`
        CompletionRate float64 `json:"completion_rate"`
    }
    
    // 获取总响应数
    if err := db.Model(&models.SurveyResponse{}).
        Where("survey_id = ? AND team_id = ?", surveyID, teamID).
        Count(&stats.TotalResponses).Error; err != nil {
        return nil, err
    }
    
    // 获取平均耗时
    if err := db.Model(&models.SurveyResponse{}).
        Where("survey_id = ? AND team_id = ?", surveyID, teamID).
        Select("AVG(time_spent) as average_time_spent").
        Scan(&stats.AverageTimeSpent).Error; err != nil {
        return nil, err
    }
    
    // 这里可以添加更复杂的统计逻辑
    
    return map[string]interface{}{
        "total_responses": stats.TotalResponses,
        "average_time_spent": stats.AverageTimeSpent,
        "completion_rate": stats.CompletionRate,
    }, nil
}
```

## 三、数据库迁移脚本

```sql
-- migrations/001_create_survey_tables.sql
CREATE TABLE survey_definitions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    survey_id VARCHAR(255) UNIQUE NOT NULL,
    title VARCHAR(255) NOT NULL,
    description TEXT,
    json_definition JSONB NOT NULL,
    version VARCHAR(50) DEFAULT '1.0',
    is_active BOOLEAN DEFAULT TRUE,
    created_by UUID NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE survey_responses (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    survey_id VARCHAR(255) NOT NULL,
    user_id UUID NOT NULL,
    team_id UUID NOT NULL,
    response_data JSONB NOT NULL,
    completed_at TIMESTAMP NOT NULL,
    time_spent INTEGER DEFAULT 0,
    ip_address VARCHAR(45),
    user_agent TEXT,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE survey_analysis (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    survey_id VARCHAR(255) NOT NULL,
    response_id UUID NOT NULL,
    analysis_data JSONB NOT NULL,
    dify_workflow_id VARCHAR(100),
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- 创建索引
CREATE INDEX idx_survey_responses_survey_id ON survey_responses(survey_id);
CREATE INDEX idx_survey_responses_team_id ON survey_responses(team_id);
CREATE INDEX idx_survey_responses_user_id ON survey_responses(user_id);
CREATE INDEX idx_survey_analysis_response_id ON survey_analysis(response_id);
```

## 四、权限控制中间件

```go
// middleware/auth.go
package middleware

import (
    "net/http"
    "github.com/gin-gonic/gin"
    "github.com/golang-jwt/jwt/v4"
)

// Auth 身份验证中间件
func Auth() gin.HandlerFunc {
    return func(c *gin.Context) {
        tokenString := c.GetHeader("Authorization")
        if tokenString == "" {
            c.JSON(http.StatusUnauthorized, gin.H{"error": "需要认证"})
            c.Abort()
            return
        }
        
        // 解析JWT token
        token, err := jwt.Parse(tokenString[7:], func(token *jwt.Token) (interface{}, error) {
            return []byte(os.Getenv("JWT_SECRET")), nil
        })
        
        if err != nil || !token.Valid {
            c.JSON(http.StatusUnauthorized, gin.H{"error": "无效的token"})
            c.Abort()
            return
        }
        
        claims := token.Claims.(jwt.MapClaims)
        c.Set("user_id", claims["user_id"])
        c.Set("team_id", claims["team_id"])
        c.Set("role", claims["role"])
        
        c.Next()
    }
}

// SurveyPermission 问卷权限控制
func SurveyPermission() gin.HandlerFunc {
    return func(c *gin.Context) {
        userRole := c.GetString("role")
        surveyID := c.Param("surveyId")
        
        // 超级管理员和团队管理员有全部权限
        if userRole == "super_admin" || userRole == "team_manager" {
            c.Next()
            return
        }
        
        // 普通用户只能访问活跃的问卷
        var survey models.SurveyDefinition
        if err := db.Where("survey_id = ? AND is_active = true", surveyID).First(&survey).Error; err != nil {
            c.JSON(http.StatusNotFound, gin.H{"error": "问卷不存在或无权访问"})
            c.Abort()
            return
        }
        
        c.Next()
    }
}
```

## 五、部署配置

### 1. 环境变量配置

```env
# .env.production
DATABASE_URL=postgresql://user:password@localhost:5432/cdk_office
JWT_SECRET=your-super-secret-jwt-key
DIFY_API_KEY=app-your-dify-api-key
DIFY_WORKFLOW_ID=your-survey-analysis-workflow-id

# SurveyJS 配置
SURVEYJS_THEME=modern
SURVEYJS_CACHE_ENABLED=true
SURVEYJS_CACHE_TTL=3600
```

### 2. Docker部署配置

```dockerfile
# Dockerfile
FROM golang:1.21-alpine AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o survey-api .

FROM alpine:latest
RUN apk --no-cache add ca-certificates
WORKDIR /root/
COPY --from=builder /app/survey-api .
COPY --from=builder /app/templates ./templates
COPY --from=builder /app/static ./static
EXPOSE 8080
CMD ["./survey-api"]
```

## 六、API接口文档

### 1. 获取问卷定义
- **URL**: `GET /api/surveys/:surveyId`
- **权限**: 需要认证
- **参数**: 
  - `surveyId` (路径参数) - 问卷ID
- **响应**: 
  ```json
  {
    "surveyId": "employee-satisfaction-2024",
    "title": "员工满意度调查",
    "pages": [...]
  }
  ```

### 2. 提交问卷响应
- **URL**: `POST /api/surveys/submit`
- **权限**: 需要认证
- **请求体**:
  ```json
  {
    "surveyId": "employee-satisfaction-2024",
    "surveyData": {
      "workplace_rating": 8,
      "workplace_suggestions": "建议改善办公设施"
    },
    "timeSpent": 120
  }
  ```
- **响应**:
  ```json
  {
    "message": "提交成功",
    "responseId": "a1b2c3d4-e5f6-7890-abcd-ef1234567890"
  }
  ```

## 七、实施建议

1.  **分阶段实施**：
    -   第一阶段：基础问卷功能（创建、发布、提交）
    -   第二阶段：集成Dify智能分析
    -   第三阶段：高级统计和报表功能

2.  **性能优化**：
    -   使用Redis缓存问卷定义
    -   数据库查询优化，添加适当索引
    -   异步处理Dify工作流调用

3.  **安全考虑**：
    -   实施CSRF保护
    -   问卷数据验证和清理
    -   防止重复提交

4.  **监控和日志**：
    -   记录问卷提交统计
    -   监控Dify工作流执行状态
    -   错误日志和告警

这个集成方案提供了从前端到后端的完整实现细节，您可以根据实际需求进行调整和扩展。SurveyJS的灵活性和您现有技术栈的兼容性使得这个集成相对 straightforward，同时提供了强大的问卷功能和智能分析能力。