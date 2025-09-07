# 员工管理模块分析功能增强文档

## 概述

本文档描述了员工管理模块中新增的分析功能，这些功能提供了更深入的人力资源洞察，帮助管理者更好地了解团队状况和员工表现。

## 新增分析功能

### 1. 员工绩效分析 (Employee Performance Analysis)

#### 功能描述
提供员工绩效统计数据，包括平均绩效评分、高绩效员工数量、低绩效员工数量等。

#### 数据结构
```go
type EmployeePerformanceStats struct {
    AverageScore   float64 `json:"average_score"`
    HighPerformers int64   `json:"high_performers"`
    LowPerformers  int64   `json:"low_performers"`
    TotalReviews   int64   `json:"total_reviews"`
}
```

#### API端点
```
GET /api/employee/analytics/performance
参数: team_id (必需)
```

### 2. 离职分析 (Termination Analysis)

#### 功能描述
分析员工离职情况，包括离职总数和离职原因统计。

#### 数据结构
```go
type TerminationAnalysis struct {
    TotalTerminations int64        `json:"total_terminations"`
    TerminationReasons []*ReasonCount `json:"termination_reasons"`
}

type ReasonCount struct {
    Reason string `json:"reason"`
    Count  int64  `json:"count"`
}
```

#### API端点
```
GET /api/employee/analytics/termination
参数: team_id (必需)
```

### 3. 员工流动率分析 (Employee Turnover Rate Analysis)

#### 功能描述
计算员工流动率，包括整体流动率、晋升率和调动率。

#### 数据结构
```go
type EmployeeTurnoverRate struct {
    TurnoverRate   float64 `json:"turnover_rate"`
    PromotionRate  float64 `json:"promotion_rate"`
    TransferRate   float64 `json:"transfer_rate"`
    TotalEmployees int64   `json:"total_employees"`
}
```

#### API端点
```
GET /api/employee/analytics/turnover
参数: team_id (必需)
```

### 4. 满意度调查分析 (Survey Analysis)

#### 功能描述
分析员工满意度调查结果，包括调查总数、响应数、响应率和平均评分。

#### 数据结构
```go
type SurveyAnalysis struct {
    TotalSurveys   int64   `json:"total_surveys"`
    TotalResponses int64   `json:"total_responses"`
    ResponseRate   float64 `json:"response_rate"`
    AverageScore   float64 `json:"average_score"`
}
```

#### API端点
```
GET /api/employee/analytics/survey
参数: team_id (必需)
```

## 实现细节

### 服务层 (Service Layer)
所有分析功能都在 `analytics_service.go` 文件中实现，通过以下方法提供：
- `GetEmployeePerformanceStats`
- `GetTerminationAnalysis`
- `GetEmployeeTurnoverRate`
- `GetSurveyAnalysis`

### 处理层 (Handler Layer)
所有API端点都在 `analytics_handler.go` 文件中实现，通过以下方法处理HTTP请求：
- `GetEmployeePerformanceStats`
- `GetTerminationAnalysis`
- `GetEmployeeTurnoverRate`
- `GetSurveyAnalysis`

## 数据依赖

这些分析功能依赖于以下数据模型：
- `PerformanceReview` - 绩效评估数据
- `TerminationRecord` - 离职记录数据
- `EmployeeLifecycleEvent` - 员工生命周期事件数据
- `EmployeeSurvey` 和 `SurveyResponse` - 员工满意度调查数据

## 使用示例

### 获取团队绩效统计
```bash
curl "http://localhost:8080/api/employee/analytics/performance?team_id=team123"
```

### 获取离职分析
```bash
curl "http://localhost:8080/api/employee/analytics/termination?team_id=team123"
```

### 获取员工流动率
```bash
curl "http://localhost:8080/api/employee/analytics/turnover?team_id=team123"
```

### 获取满意度调查分析
```bash
curl "http://localhost:8080/api/employee/analytics/survey?team_id=team123"
```