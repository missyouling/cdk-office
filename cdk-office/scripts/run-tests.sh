#!/bin/bash

# Go 测试运行脚本
# 用于运行所有后端单元测试和集成测试

echo "🚀 开始运行 CDK-Office 后端测试套件..."

# 设置测试环境变量
export GO_ENV=test
export GIN_MODE=test

# 确保在项目根目录
cd "$(dirname "$0")"

# 创建测试报告目录
mkdir -p test-reports

echo "📋 运行单元测试..."

# 运行健康检查器测试
echo "🔍 测试 ServiceHealthChecker..."
go test -v -race -coverprofile=test-reports/health_checker_coverage.out ./internal/service -run "TestHealthCheckerTestSuite" 2>&1 | tee test-reports/health_checker_test.log

if [ $? -ne 0 ]; then
    echo "❌ ServiceHealthChecker 测试失败"
    exit 1
fi

# 运行文档同步服务测试
echo "📄 测试 DocumentSyncService..."
go test -v -race -coverprofile=test-reports/document_sync_coverage.out ./internal/apps/ai -run "TestDocumentSyncTestSuite" 2>&1 | tee test-reports/document_sync_test.log

if [ $? -ne 0 ]; then
    echo "❌ DocumentSyncService 测试失败"
    exit 1
fi

# 运行API集成测试
echo "🔗 测试 AI API 集成..."
go test -v -race -coverprofile=test-reports/api_integration_coverage.out ./internal/apps/ai -run "TestAIAPIIntegrationTestSuite" 2>&1 | tee test-reports/api_integration_test.log

if [ $? -ne 0 ]; then
    echo "❌ API集成测试失败"
    exit 1
fi

# 运行基准测试
echo "⚡ 运行性能基准测试..."
go test -bench=. -benchmem ./internal/service ./internal/apps/ai 2>&1 | tee test-reports/benchmark_results.log

# 生成总体测试覆盖率报告
echo "📊 生成测试覆盖率报告..."

# 合并覆盖率文件
echo "mode: atomic" > test-reports/total_coverage.out
tail -n +2 test-reports/health_checker_coverage.out >> test-reports/total_coverage.out
tail -n +2 test-reports/document_sync_coverage.out >> test-reports/total_coverage.out  
tail -n +2 test-reports/api_integration_coverage.out >> test-reports/total_coverage.out

# 生成HTML覆盖率报告
go tool cover -html=test-reports/total_coverage.out -o test-reports/coverage_report.html

# 显示覆盖率摘要
go tool cover -func=test-reports/total_coverage.out | tee test-reports/coverage_summary.txt

echo ""
echo "📈 测试覆盖率摘要:"
tail -1 test-reports/coverage_summary.txt

echo ""
echo "✅ 所有后端测试完成!"
echo "📋 测试报告保存在: test-reports/"
echo "🌐 HTML覆盖率报告: test-reports/coverage_report.html"

# 检查最低覆盖率要求 (80%)
COVERAGE=$(tail -1 test-reports/coverage_summary.txt | grep -o '[0-9.]*%' | tr -d '%')
THRESHOLD=80

if (( $(echo "$COVERAGE >= $THRESHOLD" | bc -l) )); then
    echo "🎉 覆盖率 ${COVERAGE}% 达到要求 (>= ${THRESHOLD}%)"
    exit 0
else
    echo "⚠️  覆盖率 ${COVERAGE}% 低于要求 (>= ${THRESHOLD}%)"
    exit 1
fi
