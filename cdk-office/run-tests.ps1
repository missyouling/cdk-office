# CDK-Office 测试运行脚本 (Windows PowerShell)

Write-Host "🚀 开始运行 CDK-Office 测试套件..." -ForegroundColor Green

# 设置测试环境变量
$env:GO_ENV = "test"
$env:GIN_MODE = "test"

# 切换到项目根目录
Set-Location $PSScriptRoot

# 创建测试报告目录
if (!(Test-Path "test-reports")) {
    New-Item -ItemType Directory -Name "test-reports"
}

Write-Host "📋 检查Go模块状态..." -ForegroundColor Yellow

# 检查和整理Go模块
if (Test-Path "go.mod") {
    go mod tidy
    if ($LASTEXITCODE -ne 0) {
        Write-Host "❌ Go模块整理失败" -ForegroundColor Red
        exit 1
    }
} else {
    Write-Host "❌ 找不到go.mod文件" -ForegroundColor Red
    exit 1
}

Write-Host "🔍 运行ServiceHealthChecker测试..." -ForegroundColor Yellow
go test -v -race -coverprofile=test-reports/health_checker_coverage.out ./internal/service -run "TestHealthCheckerTestSuite" | Tee-Object test-reports/health_checker_test.log

if ($LASTEXITCODE -ne 0) {
    Write-Host "❌ ServiceHealthChecker 测试失败" -ForegroundColor Red
    exit 1
}

Write-Host "📄 运行DocumentSyncService测试..." -ForegroundColor Yellow
go test -v -race -coverprofile=test-reports/document_sync_coverage.out ./internal/apps/ai -run "TestDocumentSyncTestSuite" | Tee-Object test-reports/document_sync_test.log

if ($LASTEXITCODE -ne 0) {
    Write-Host "❌ DocumentSyncService 测试失败" -ForegroundColor Red
    exit 1
}

Write-Host "🔗 运行AI API集成测试..." -ForegroundColor Yellow
go test -v -race -coverprofile=test-reports/api_integration_coverage.out ./internal/apps/ai -run "TestAIAPIIntegrationTestSuite" | Tee-Object test-reports/api_integration_test.log

if ($LASTEXITCODE -ne 0) {
    Write-Host "❌ API集成测试失败" -ForegroundColor Red
    exit 1
}

Write-Host "⚡ 运行性能基准测试..." -ForegroundColor Yellow
go test -bench=. -benchmem ./internal/service ./internal/apps/ai | Tee-Object test-reports/benchmark_results.log

Write-Host "📊 生成测试覆盖率报告..." -ForegroundColor Yellow

# 合并覆盖率文件
$coverageFiles = @("test-reports/health_checker_coverage.out", "test-reports/document_sync_coverage.out", "test-reports/api_integration_coverage.out")
$totalCoverage = "test-reports/total_coverage.out"

"mode: atomic" | Out-File -FilePath $totalCoverage -Encoding UTF8

foreach ($file in $coverageFiles) {
    if (Test-Path $file) {
        Get-Content $file | Select-Object -Skip 1 | Add-Content $totalCoverage
    }
}

# 生成HTML覆盖率报告
go tool cover -html=$totalCoverage -o test-reports/coverage_report.html

# 显示覆盖率摘要
go tool cover -func=$totalCoverage | Tee-Object test-reports/coverage_summary.txt

Write-Host ""
Write-Host "📈 测试覆盖率摘要:" -ForegroundColor Cyan
Get-Content test-reports/coverage_summary.txt | Select-Object -Last 1

Write-Host ""
Write-Host "✅ 所有后端测试完成!" -ForegroundColor Green
Write-Host "📋 测试报告保存在: test-reports/" -ForegroundColor Cyan
Write-Host "🌐 HTML覆盖率报告: test-reports/coverage_report.html" -ForegroundColor Cyan

# 检查最低覆盖率要求 (80%)
$coverageLine = Get-Content test-reports/coverage_summary.txt | Select-Object -Last 1
$coverage = [regex]::Match($coverageLine, '(\d+\.?\d*)%').Groups[1].Value

if ($coverage -as [double] -ge 80) {
    Write-Host "🎉 覆盖率 ${coverage}% 达到要求 (>= 80%)" -ForegroundColor Green
    exit 0
} else {
    Write-Host "⚠️  覆盖率 ${coverage}% 低于要求 (>= 80%)" -ForegroundColor Yellow
    exit 1
}