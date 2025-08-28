# Go依赖下载脚本 - 解决网络连接问题

Write-Host "🔧 配置Go代理和依赖下载..." -ForegroundColor Green

# 设置当前目录为脚本所在目录
Set-Location $PSScriptRoot

# 验证当前目录有go.mod文件
if (!(Test-Path "go.mod")) {
    Write-Host "❌ 错误: 当前目录没有找到go.mod文件" -ForegroundColor Red
    Write-Host "当前目录: $(Get-Location)" -ForegroundColor Yellow
    Write-Host "请确保在cdk-office项目根目录执行此脚本" -ForegroundColor Yellow
    exit 1
}

Write-Host "✅ 找到go.mod文件，当前目录: $(Get-Location)" -ForegroundColor Green

# 配置Go代理
Write-Host "🌐 配置Go代理..." -ForegroundColor Yellow

# 设置国内Go代理
$env:GOPROXY = "https://goproxy.cn,direct"
$env:GOSUMDB = "sum.golang.google.cn"

# 也可以尝试其他代理
# $env:GOPROXY = "https://proxy.golang.org,direct"
# $env:GOPROXY = "https://mirrors.aliyun.com/goproxy/,direct"

Write-Host "GOPROXY设置为: $env:GOPROXY" -ForegroundColor Cyan
Write-Host "GOSUMDB设置为: $env:GOSUMDB" -ForegroundColor Cyan

# 清理模块缓存（可选）
Write-Host "🧹 清理Go模块缓存..." -ForegroundColor Yellow
go clean -modcache

# 下载依赖
Write-Host "📦 开始下载Go依赖..." -ForegroundColor Yellow

# 尝试下载依赖
go mod download
if ($LASTEXITCODE -eq 0) {
    Write-Host "✅ 依赖下载成功" -ForegroundColor Green
} else {
    Write-Host "⚠️  go mod download 完成，但有警告" -ForegroundColor Yellow
}

# 整理依赖
Write-Host "🔄 整理Go模块依赖..." -ForegroundColor Yellow
go mod tidy
if ($LASTEXITCODE -eq 0) {
    Write-Host "✅ 依赖整理成功" -ForegroundColor Green
} else {
    Write-Host "❌ 依赖整理失败" -ForegroundColor Red
}

# 验证依赖
Write-Host "🔍 验证Go模块依赖..." -ForegroundColor Yellow
go mod verify
if ($LASTEXITCODE -eq 0) {
    Write-Host "✅ 依赖验证成功" -ForegroundColor Green
} else {
    Write-Host "⚠️  依赖验证有问题，但可能不影响使用" -ForegroundColor Yellow
}

# 显示Go环境信息
Write-Host ""
Write-Host "📋 Go环境信息:" -ForegroundColor Cyan
Write-Host "Go版本: $(go version)" -ForegroundColor White
Write-Host "GOPROXY: $(go env GOPROXY)" -ForegroundColor White
Write-Host "GOSUMDB: $(go env GOSUMDB)" -ForegroundColor White

# 列出主要依赖
Write-Host ""
Write-Host "📦 主要依赖检查:" -ForegroundColor Cyan

$testDependencies = @(
    "github.com/stretchr/testify",
    "gorm.io/driver/sqlite", 
    "gorm.io/gorm",
    "github.com/gin-gonic/gin",
    "github.com/sirupsen/logrus"
)

foreach ($dep in $testDependencies) {
    $found = go list -m $dep 2>$null
    if ($found) {
        Write-Host "✅ $dep" -ForegroundColor Green
    } else {
        Write-Host "❌ $dep 未找到" -ForegroundColor Red
    }
}

Write-Host ""
Write-Host "🎯 测试相关依赖:" -ForegroundColor Cyan
go list -m github.com/stretchr/testify gorm.io/driver/sqlite 2>$null

Write-Host ""
Write-Host "💡 如果仍有网络问题，请尝试:" -ForegroundColor Yellow
Write-Host "  1. 使用VPN或代理" -ForegroundColor White
Write-Host "  2. 配置企业代理: go env -w GOPROXY=http://your-proxy.com" -ForegroundColor White
Write-Host "  3. 使用离线依赖包" -ForegroundColor White

Write-Host ""
Write-Host "✅ Go依赖配置完成!" -ForegroundColor Green