#!/bin/bash

# VPS部署脚本
# 用于在2C4G配置的VPS上部署CDK应用

set -e

echo "开始VPS优化部署..."

# 检查是否在VPS环境中
MEMORY_KB=$(grep MemTotal /proc/meminfo | awk '{print $2}')
MEMORY_GB=$((MEMORY_KB / 1024 / 1024))

if [ $MEMORY_GB -gt 4 ]; then
  echo "警告: 检测到系统内存大于4GB，可能不是VPS环境"
  echo "是否继续使用VPS优化配置？(y/N)"
  read -r response
  if [[ ! "$response" =~ ^([yY][eE][sS]|[yY])$ ]]; then
    echo "部署取消"
    exit 1
  fi
fi

# 构建优化的二进制文件
echo "构建优化的二进制文件..."
go build -ldflags="-s -w" -o cdk-server main.go

# 复制VPS优化配置
echo "应用VPS优化配置..."
cp config.vps.yaml config.yaml

# 设置环境变量
export GOGC=20
export GOMEMLIMIT=3GiB

# 创建日志目录
mkdir -p logs

# 启动服务
echo "启动API服务..."
nohup ./cdk-server api > logs/api.log 2>&1 &
API_PID=$!

echo "启动Worker服务..."
nohup ./cdk-server worker > logs/worker.log 2>&1 &
WORKER_PID=$!

echo "服务启动完成"
echo "API服务PID: $API_PID"
echo "Worker服务PID: $WORKER_PID"
echo "日志文件位于 logs/ 目录下"

# 保存PID到文件
echo $API_PID > cdk-api.pid
echo $WORKER_PID > cdk-worker.pid

echo "PID已保存到 cdk-api.pid 和 cdk-worker.pid"

# 显示服务状态
sleep 2
echo "服务状态:"
ps -p $API_PID && echo "API服务运行中" || echo "API服务启动失败"
ps -p $WORKER_PID && echo "Worker服务运行中" || echo "Worker服务启动失败"

echo "VPS部署完成"