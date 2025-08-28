# Go环境设置脚本 - 简化版

Write-Host "🔧 设置Go环境..." -ForegroundColor Green

# 确保在正确目录
Set-Location $PSScriptRoot

# 设置Go代理环境变量
Write-Host "🌐 设置Go代理..." -ForegroundColor Yellow

# 方法1: 使用国内代理
go env -w GOPROXY=https://goproxy.cn,direct
go env -w GOSUMDB=sum.golang.google.cn

Write-Host "✅ Go代理设置完成" -ForegroundColor Green
Write-Host "GOPROXY: $(go env GOPROXY)" -ForegroundColor Cyan

# 显示Go环境
Write-Host ""
Write-Host "📋 当前Go环境:" -ForegroundColor Cyan
go version
Write-Host "GOPROXY: $(go env GOPROXY)" -ForegroundColor White
Write-Host "GOPATH: $(go env GOPATH)" -ForegroundColor White
Write-Host "GOROOT: $(go env GOROOT)" -ForegroundColor White

Write-Host ""
Write-Host "💡 后续步骤:" -ForegroundColor Yellow
Write-Host "1. 现在可以尝试运行: go mod tidy" -ForegroundColor White
Write-Host "2. 如果仍有问题，可以尝试: go clean -modcache" -ForegroundColor White
Write-Host "3. 或者使用: go mod download" -ForegroundColor White

Write-Host ""
Write-Host "✅ Go环境配置完成!" -ForegroundColor Green