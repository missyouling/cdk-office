#!/bin/bash

# VPS停止脚本
# 用于停止在VPS上运行的CDK应用

set -e

echo "停止CDK服务..."

# 检查PID文件是否存在
if [ ! -f cdk-api.pid ] || [ ! -f cdk-worker.pid ]; then
  echo "PID文件不存在，服务可能未运行"
  exit 1
fi

# 读取PID
API_PID=$(cat cdk-api.pid)
WORKER_PID=$(cat cdk-worker.pid)

# 停止服务
echo "停止API服务 (PID: $API_PID)..."
kill $API_PID || echo "API服务可能已经停止"

echo "停止Worker服务 (PID: $WORKER_PID)..."
kill $WORKER_PID || echo "Worker服务可能已经停止"

# 等待进程结束
sleep 3

# 检查进程是否仍在运行
if ps -p $API_PID > /dev/null; then
  echo "API服务仍在运行，强制终止..."
  kill -9 $API_PID
fi

if ps -p $WORKER_PID > /dev/null; then
  echo "Worker服务仍在运行，强制终止..."
  kill -9 $WORKER_PID
fi

# 删除PID文件
rm -f cdk-api.pid cdk-worker.pid

echo "CDK服务已停止"