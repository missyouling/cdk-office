# CDK-Office 测试文档

## 概述

本文档介绍了CDK-Office项目的测试策略、测试执行方法和测试覆盖率要求。

## 测试策略

### 测试金字塔

```
        /\
       /  \
      /    \
     /  E2E  \     端到端测试 (少量)
    /________\
   /          \
  /Integration \   集成测试 (适量)
 /_____________\
/               \
/   Unit Tests   \  单元测试 (大量)
/________________\
```

### 测试类型

1. **单元测试 (Unit Tests)**
   - 测试单个函数或方法
   - 覆盖核心业务逻辑
   - 快速执行，无外部依赖

2. **集成测试 (Integration Tests)**
   - 测试模块间交互
   - 测试数据库操作
   - 测试API接口

3. **端到端测试 (E2E Tests)**
   - 测试完整用户流程
   - 浏览器自动化测试
   - 关键业务场景验证

## 测试工具

### Go 后端测试
- **测试框架**: Go标准库 `testing`
- **断言库**: `github.com/stretchr/testify`
- **Mock库**: `github.com/stretchr/testify/mock`
- **HTTP测试**: `net/http/httptest`
- **数据库测试**: 内存SQLite数据库

### 前端测试
- **测试框架**: Jest
- **组件测试**: React Testing Library
- **E2E测试**: Playwright
- **Mock库**: MSW (Mock Service Worker)

## 测试执行

### 快速开始

```bash
# 运行所有测试
make test

# 运行单元测试
make test-unit

# 运行集成测试
make test-integration

# 生成覆盖率报告
make test-coverage
```

### 详细命令

#### 后端测试

```bash
# 运行所有后端测试
go test ./...

# 运行特定包的测试
go test ./internal/services/isolation/...

# 运行测试并显示详细输出
go test -v ./...

# 运行测试并生成覆盖率
go test -cover ./...

# 运行测试并生成HTML覆盖率报告
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out -o coverage.html

# 运行基准测试
go test -bench=. ./...

# 运行竞态条件检测
go test -race ./...
```

#### 前端测试

```bash
cd frontend

# 运行单元测试
npm test

# 运行测试并生成覆盖率
npm run test:coverage

# 运行E2E测试
npm run test:e2e

# 运行组件测试
npm run test:component
```

### 使用测试脚本

项目提供了自动化测试脚本：

```bash
# 运行完整测试套件
./scripts/run-tests.sh

# 仅运行快速测试
./scripts/run-tests.sh --quick

# 运行测试并生成报告
./scripts/run-tests.sh --report
```

## 测试覆盖率

### 覆盖率要求

- **整体覆盖率**: ≥ 80%
- **核心业务逻辑**: ≥ 90%
- **API接口**: ≥ 85%
- **数据库操作**: ≥ 85%

### 查看覆盖率报告

```bash
# 生成覆盖率报告
make test-coverage

# 在浏览器中查看HTML报告
open test-results/coverage.html

# 查看控制台覆盖率统计
go tool cover -func=test-results/total-coverage.out
```

## 测试数据管理

### 测试数据库

```bash
# 设置测试数据库
export DB_NAME=cdk_office_test
export DB_HOST=localhost

# 运行数据库迁移
go run . migrate --env=test

# 清理测试数据
go run . clean-test-data
```

### Mock数据

测试中使用Mock数据避免外部依赖：

```go
// 示例：Mock数据库操作
func TestUserService_CreateUser(t *testing.T) {
    mockDB := &MockDB{}
    service := &UserService{db: mockDB}
    
    user := &User{Name: "Test User"}
    mockDB.On("Create", user).Return(nil)
    
    err := service.CreateUser(user)
    assert.NoError(t, err)
    mockDB.AssertExpectations(t)
}
```

## 编写测试

### 单元测试示例

```go
package isolation

import (
    "testing"
    "github.com/stretchr/testify/assert"
)

func TestValidatePermissionLevel(t *testing.T) {
    testCases := []struct {
        name           string
        userRole       string
        requiredLevel  PermissionLevel
        expectedResult bool
    }{
        {
            name:           "SuperAdmin can access SystemAdmin",
            userRole:       "super_admin",
            requiredLevel:  SystemAdmin,
            expectedResult: true,
        },
        {
            name:           "TeamMember cannot access TeamAdmin",
            userRole:       "team_member",
            requiredLevel:  TeamAdmin,
            expectedResult: false,
        },
    }

    service := &IsolationService{}
    
    for _, tc := range testCases {
        t.Run(tc.name, func(t *testing.T) {
            result := service.validatePermissionLevel(tc.userRole, tc.requiredLevel)
            assert.Equal(t, tc.expectedResult, result)
        })
    }
}
```

### 集成测试示例

```go
func TestWorkflowIntegration(t *testing.T) {
    // 设置测试环境
    gin.SetMode(gin.TestMode)
    
    // 创建测试路由
    router := setupTestRouter()
    
    // 测试创建工作流
    workflowData := map[string]interface{}{
        "name": "Test Workflow",
        "definition": map[string]interface{}{
            "steps": []map[string]interface{}{
                {"id": "start", "type": "start"},
            },
        },
    }
    
    jsonData, _ := json.Marshal(workflowData)
    req := httptest.NewRequest("POST", "/api/v1/workflow/definitions", 
        bytes.NewBuffer(jsonData))
    req.Header.Set("Content-Type", "application/json")
    
    w := httptest.NewRecorder()
    router.ServeHTTP(w, req)
    
    assert.Equal(t, http.StatusCreated, w.Code)
}
```

### Mock服务示例

```go
type MockPDFService struct {
    mock.Mock
}

func (m *MockPDFService) MergePDFs(files []string) (*PDFResult, error) {
    args := m.Called(files)
    return args.Get(0).(*PDFResult), args.Error(1)
}

func TestPDFController_Merge(t *testing.T) {
    mockService := new(MockPDFService)
    controller := &PDFController{service: mockService}
    
    expectedResult := &PDFResult{FilePath: "/merged.pdf"}
    mockService.On("MergePDFs", []string{"file1.pdf", "file2.pdf"}).
        Return(expectedResult, nil)
    
    // 执行测试...
    
    mockService.AssertExpectations(t)
}
```

## 性能测试

### 基准测试

```go
func BenchmarkDataAccessCheck(b *testing.B) {
    service := &IsolationService{}
    
    b.ResetTimer()
    for i := 0; i < b.N; i++ {
        allowed, _ := service.CheckDataAccess(
            context.Background(), 
            1, 1, "document", "doc_123", "read")
        _ = allowed
    }
}

func BenchmarkCacheOperation(b *testing.B) {
    cache := NewCacheOptimizer(DefaultConfig())
    
    b.ResetTimer()
    for i := 0; i < b.N; i++ {
        cache.Set("test", "key", "value", time.Minute)
        cache.Get("test", "key")
    }
}
```

### 负载测试

```bash
# 使用Apache Bench进行负载测试
ab -n 1000 -c 10 http://localhost:8000/api/v1/health

# 使用wrk进行高并发测试
wrk -t12 -c400 -d30s http://localhost:8000/api/v1/health
```

## 测试环境

### Docker测试环境

```yaml
# docker-compose.test.yml
version: '3.8'
services:
  postgres-test:
    image: postgres:15-alpine
    environment:
      POSTGRES_DB: cdk_office_test
      POSTGRES_USER: test_user
      POSTGRES_PASSWORD: test_password
    ports:
      - "5433:5432"
  
  redis-test:
    image: redis:7-alpine
    ports:
      - "6380:6379"
```

```bash
# 启动测试环境
docker-compose -f docker-compose.test.yml up -d

# 运行测试
make test

# 清理测试环境
docker-compose -f docker-compose.test.yml down
```

### 环境隔离

```bash
# 设置测试环境变量
export APP_ENV=test
export DB_NAME=cdk_office_test
export REDIS_DB=1

# 运行测试
go test ./...
```

## 持续集成

### GitHub Actions

```yaml
name: Tests
on: [push, pull_request]

jobs:
  test:
    runs-on: ubuntu-latest
    
    services:
      postgres:
        image: postgres:15
        env:
          POSTGRES_PASSWORD: postgres
          POSTGRES_DB: cdk_office_test
        options: >-
          --health-cmd pg_isready
          --health-interval 10s
          --health-timeout 5s
          --health-retries 5
        ports:
          - 5432:5432
      
      redis:
        image: redis:7
        options: >-
          --health-cmd "redis-cli ping"
          --health-interval 10s
          --health-timeout 5s
          --health-retries 5
        ports:
          - 6379:6379
    
    steps:
    - uses: actions/checkout@v3
    
    - name: Set up Go
      uses: actions/setup-go@v3
      with:
        go-version: 1.24
    
    - name: Install dependencies
      run: go mod download
    
    - name: Run tests
      run: make test
      env:
        DB_HOST: localhost
        DB_PORT: 5432
        DB_NAME: cdk_office_test
        DB_USER: postgres
        DB_PASSWORD: postgres
        REDIS_HOST: localhost
        REDIS_PORT: 6379
    
    - name: Upload coverage
      uses: codecov/codecov-action@v3
      with:
        file: ./coverage.out
```

## 测试最佳实践

### 测试命名规范

```go
// 好的测试命名
func TestUserService_CreateUser_WithValidData_ShouldReturnNoError(t *testing.T) {
    // ...
}

func TestUserService_CreateUser_WithInvalidEmail_ShouldReturnValidationError(t *testing.T) {
    // ...
}

// 表格驱动测试
func TestValidateEmail(t *testing.T) {
    testCases := []struct {
        name     string
        email    string
        expected bool
    }{
        {"valid email", "user@example.com", true},
        {"invalid email", "invalid-email", false},
        {"empty email", "", false},
    }
    
    for _, tc := range testCases {
        t.Run(tc.name, func(t *testing.T) {
            result := ValidateEmail(tc.email)
            assert.Equal(t, tc.expected, result)
        })
    }
}
```

### 测试数据管理

```go
// 使用工厂模式创建测试数据
func NewTestUser() *User {
    return &User{
        ID:    1,
        Name:  "Test User",
        Email: "test@example.com",
    }
}

// 使用Builder模式创建复杂测试数据
type UserBuilder struct {
    user *User
}

func NewUserBuilder() *UserBuilder {
    return &UserBuilder{
        user: &User{},
    }
}

func (b *UserBuilder) WithName(name string) *UserBuilder {
    b.user.Name = name
    return b
}

func (b *UserBuilder) WithEmail(email string) *UserBuilder {
    b.user.Email = email
    return b
}

func (b *UserBuilder) Build() *User {
    return b.user
}
```

### 测试组织

```go
// 使用子测试组织相关测试
func TestUserService(t *testing.T) {
    service := NewUserService()
    
    t.Run("CreateUser", func(t *testing.T) {
        t.Run("with valid data", func(t *testing.T) {
            // 测试逻辑
        })
        
        t.Run("with invalid data", func(t *testing.T) {
            // 测试逻辑
        })
    })
    
    t.Run("GetUser", func(t *testing.T) {
        // GetUser相关测试
    })
}
```

## 故障排除

### 常见测试问题

1. **测试数据库连接失败**
   ```bash
   # 检查测试数据库状态
   docker-compose ps postgres-test
   
   # 查看数据库日志
   docker-compose logs postgres-test
   ```

2. **测试超时**
   ```go
   // 设置测试超时
   func TestLongRunningOperation(t *testing.T) {
       ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
       defer cancel()
       
       // 测试逻辑
   }
   ```

3. **竞态条件检测**
   ```bash
   # 运行竞态条件检测
   go test -race ./...
   
   # 如果发现竞态条件，使用sync包进行同步
   ```

### 调试测试

```bash
# 运行特定测试
go test -run TestSpecificFunction ./...

# 运行测试并显示详细输出
go test -v -run TestSpecificFunction ./...

# 使用delve调试器
dlv test ./internal/services/isolation -- -test.run TestSpecificFunction
```

## 总结

- 遵循测试金字塔原则，以单元测试为主
- 保持测试覆盖率在80%以上
- 使用Mock避免外部依赖
- 编写清晰、可维护的测试代码
- 定期运行测试确保代码质量
- 在CI/CD流程中集成自动化测试