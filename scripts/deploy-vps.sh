#!/bin/bash

# CDK-Office VPS部署脚本

set -e

# 配置变量
APP_NAME="cdk-office"
BUILD_DIR="./build"
CONFIG_FILE="./config.vps.yaml"
SERVICE_NAME="cdk-office"
SERVICE_FILE="/etc/systemd/system/${SERVICE_NAME}.service"

echo "Starting CDK-Office VPS deployment..."

# 1. 构建应用
echo "1. Building application..."
make build-vps

# 2. 停止现有服务
echo "2. Stopping existing service..."
systemctl stop ${SERVICE_NAME} || true

# 3. 备份现有文件
echo "3. Backing up existing files..."
if [ -f "/opt/${APP_NAME}/${APP_NAME}" ]; then
    mkdir -p /opt/${APP_NAME}/backup
    cp /opt/${APP_NAME}/${APP_NAME} /opt/${APP_NAME}/backup/${APP_NAME}.$(date +%Y%m%d_%H%M%S)
fi

# 4. 复制新文件
echo "4. Copying new files..."
mkdir -p /opt/${APP_NAME}
cp ${BUILD_DIR}/${APP_NAME}-vps /opt/${APP_NAME}/${APP_NAME}
cp ${CONFIG_FILE} /opt/${APP_NAME}/config.yaml

# 5. 设置权限
echo "5. Setting permissions..."
chmod +x /opt/${APP_NAME}/${APP_NAME}

# 6. 创建systemd服务文件
echo "6. Creating systemd service file..."
cat > ${SERVICE_FILE} << EOF
[Unit]
Description=CDK-Office Service
After=network.target

[Service]
Type=simple
User=root
WorkingDirectory=/opt/${APP_NAME}
ExecStart=/opt/${APP_NAME}/${APP_NAME} api
Restart=always
RestartSec=10
Environment=GIN_MODE=production

[Install]
WantedBy=multi-user.target
EOF

# 7. 重新加载systemd
echo "7. Reloading systemd..."
systemctl daemon-reload

# 8. 启动服务
echo "8. Starting service..."
systemctl enable ${SERVICE_NAME}
systemctl start ${SERVICE_NAME}

# 9. 检查服务状态
echo "9. Checking service status..."
systemctl status ${SERVICE_NAME} --no-pager

echo "CDK-Office VPS deployment completed!"
echo "Service is running at http://localhost:8000"