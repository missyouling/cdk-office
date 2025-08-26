#!/bin/bash

# CDK-Office VPS停止脚本

set -e

# 配置变量
SERVICE_NAME="cdk-office"

echo "Stopping CDK-Office service..."

# 停止服务
systemctl stop ${SERVICE_NAME} || true

# 禁用服务
systemctl disable ${SERVICE_NAME} || true

# 删除服务文件
rm -f /etc/systemd/system/${SERVICE_NAME}.service

# 重新加载systemd
systemctl daemon-reload

echo "CDK-Office service stopped and disabled!"