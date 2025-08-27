#!/bin/bash

# CDK-Office 测试执行脚本
# 运行单元测试和集成测试

set -e

echo "🧪 开始执行 CDK-Office 测试套件..."

# 颜色定义
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
RED='\033[0;31m'
NC='\033[0m' # No Color

# 项目根目录
PROJECT_ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
cd "$PROJECT_ROOT"

echo "📁 项目根目录: $PROJECT_ROOT"

# 检查Go环境
if ! command -v go &> /dev/null; then
    echo -e "${RED}❌ Go 未安装或不在PATH中${NC}"
    exit 1
fi

echo -e "${GREEN}✅ Go 版本: $(go version)${NC}"

# 下载依赖
echo "📦 下载依赖包..."
go mod download
go mod tidy

# 运行代码检查
echo "🔍 运行代码检查..."
if command -v golangci-lint &> /dev/null; then
    golangci-lint run --timeout=5m
else
    echo -e "${YELLOW}⚠️  golangci-lint 未安装，跳过代码检查${NC}"
fi

# 运行格式化检查
echo "🎨 检查代码格式..."
UNFORMATTED=$(gofmt -l .)
if [ -n "$UNFORMATTED" ]; then
    echo -e "${RED}❌ 以下文件需要格式化:${NC}"
    echo "$UNFORMATTED"
    echo "运行 'gofmt -w .' 来格式化代码"
    exit 1
fi

# 创建测试结果目录
mkdir -p test-results

# 运行单元测试
echo "🧪 运行单元测试..."
echo "==============================================="

# 数据隔离模块测试
echo -e "${YELLOW}🔒 测试数据隔离模块...${NC}"
go test -v -race -coverprofile=test-results/isolation-coverage.out \
    ./internal/services/isolation/... || {
    echo -e "${RED}❌ 数据隔离模块测试失败${NC}"
    exit 1
}

# 系统优化模块测试
echo -e "${YELLOW}⚡ 测试系统优化模块...${NC}"
go test -v -race -coverprofile=test-results/optimization-coverage.out \
    ./internal/services/optimization/... || {
    echo -e "${RED}❌ 系统优化模块测试失败${NC}"
    exit 1
}

# 审批流程模块测试
echo -e "${YELLOW}📋 测试审批流程模块...${NC}"
go test -v -race -coverprofile=test-results/workflow-coverage.out \
    ./internal/apps/workflow/... || {
    echo -e "${RED}❌ 审批流程模块测试失败${NC}"
    exit 1
}

# 知识库模块测试
echo -e "${YELLOW}📚 测试知识库模块...${NC}"
go test -v -race -coverprofile=test-results/knowledge-coverage.out \
    ./internal/apps/knowledge/... || {
    echo -e "${RED}❌ 知识库模块测试失败${NC}"
    exit 1
}

# 运行所有其他单元测试
echo -e "${YELLOW}🔄 运行其他单元测试...${NC}"
go test -v -race -coverprofile=test-results/unit-coverage.out \
    ./internal/... || {
    echo -e "${RED}❌ 单元测试失败${NC}"
    exit 1
}

# 运行集成测试
echo "🔗 运行集成测试..."
echo "==============================================="

export RUN_INTEGRATION_TESTS=1
go test -v -race -timeout=5m -coverprofile=test-results/integration-coverage.out \
    ./tests/... || {
    echo -e "${RED}❌ 集成测试失败${NC}"
    exit 1
}

# 生成测试覆盖率报告
echo "📊 生成测试覆盖率报告..."

# 合并覆盖率文件
echo "mode: atomic" > test-results/total-coverage.out
tail -n +2 test-results/*-coverage.out >> test-results/total-coverage.out

# 生成HTML报告
go tool cover -html=test-results/total-coverage.out -o test-results/coverage.html

# 显示覆盖率统计
COVERAGE=$(go tool cover -func=test-results/total-coverage.out | grep total | awk '{print $3}')
echo -e "${GREEN}📈 总体测试覆盖率: $COVERAGE${NC}"

# 运行性能基准测试
echo "🚀 运行性能基准测试..."
echo "==============================================="

go test -bench=. -benchmem -run=^$ ./tests/... > test-results/benchmark.txt || {
    echo -e "${YELLOW}⚠️  性能基准测试失败，但不影响主要测试${NC}"
}

# 运行竞态条件检测
echo "🏃 运行竞态条件检测..."
go test -race -short ./... || {
    echo -e "${RED}❌ 发现竞态条件${NC}"
    exit 1
}

# 运行内存泄漏检测（如果可用）
if command -v goleak &> /dev/null; then
    echo "🧠 运行内存泄漏检测..."
    go test -tags=leak ./internal/... || {
        echo -e "${YELLOW}⚠️  内存泄漏检测警告${NC}"
    }
fi

# 生成测试报告
echo "📋 生成测试报告..."
cat > test-results/test-report.md << EOF
# CDK-Office 测试报告

生成时间: $(date)

## 测试覆盖率
- 总体覆盖率: $COVERAGE
- 详细报告: [coverage.html](coverage.html)

## 测试模块

### ✅ 单元测试
- 数据隔离模块: 通过
- 系统优化模块: 通过  
- 审批流程模块: 通过
- 知识库模块: 通过
- 其他模块: 通过

### ✅ 集成测试
- API接口测试: 通过
- 模块间协作测试: 通过
- 数据库集成测试: 通过

### ✅ 性能测试
- 基准测试结果: [benchmark.txt](benchmark.txt)
- 竞态条件检测: 通过

## 测试文件
$(find . -name "*_test.go" | wc -l) 个测试文件
$(grep -r "func Test" --include="*_test.go" . | wc -l) 个测试函数

## 建议
1. 定期运行测试确保代码质量
2. 新功能必须包含相应的测试
3. 保持测试覆盖率在 80% 以上
4. 关注性能基准测试结果

EOF

echo "==============================================="
echo -e "${GREEN}🎉 所有测试执行完成！${NC}"
echo -e "${GREEN}📊 测试覆盖率: $COVERAGE${NC}"
echo -e "${GREEN}📋 测试报告: test-results/test-report.md${NC}"
echo -e "${GREEN}📈 覆盖率报告: test-results/coverage.html${NC}"

# 检查最低覆盖率要求
MIN_COVERAGE=70
COVERAGE_NUM=$(echo $COVERAGE | tr -d '%')
if (( $(echo "$COVERAGE_NUM < $MIN_COVERAGE" | bc -l) )); then
    echo -e "${YELLOW}⚠️  测试覆盖率 ($COVERAGE) 低于最低要求 (${MIN_COVERAGE}%)${NC}"
    exit 1
fi

echo -e "${GREEN}✅ 测试套件执行成功，覆盖率达标！${NC}"