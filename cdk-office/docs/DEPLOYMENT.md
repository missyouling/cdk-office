# CDK-Office 部署文档

## 概述

本文档详细介绍了CDK-Office系统的部署方法，包括开发环境、测试环境和生产环境的部署步骤。

## 系统要求

### 最低配置要求

- **CPU**: 2核心
- **内存**: 4GB RAM
- **存储**: 20GB 可用空间
- **网络**: 1Mbps 带宽

### 推荐配置要求

- **CPU**: 4核心或更多
- **内存**: 8GB RAM 或更多
- **存储**: 50GB SSD 存储
- **网络**: 10Mbps 带宽或更多

### 软件依赖

- **操作系统**: Linux (Ubuntu 20.04+, CentOS 7+) 或 Windows 10+
- **Docker**: 20.10.0+
- **Docker Compose**: 1.29.0+
- **Go**: 1.19+ (开发环境)
- **Node.js**: 16.0+ (前端开发)

## 快速部署 (Docker Compose)

### 1. 克隆项目

```bash
git clone https://github.com/your-org/cdk-office.git
cd cdk-office
```

### 2. 配置环境变量

```bash
cp .env.example .env
# 编辑 .env 文件，配置数据库密码等敏感信息
vim .env
```

### 3. 启动服务

```bash
# 启动所有服务
make deploy

# 或使用docker-compose
docker-compose -f docker-compose.prod.yml up -d
```

### 4. 验证部署

```bash
# 检查服务状态
make status

# 查看日志
make logs

# 健康检查
curl http://localhost:8000/api/v1/health
```

## 详细部署步骤

### 1. 准备环境

#### Ubuntu/Debian 系统

```bash
# 更新系统
sudo apt update && sudo apt upgrade -y

# 安装 Docker
curl -fsSL https://get.docker.com -o get-docker.sh
sudo sh get-docker.sh

# 安装 Docker Compose
sudo curl -L "https://github.com/docker/compose/releases/latest/download/docker-compose-$(uname -s)-$(uname -m)" -o /usr/local/bin/docker-compose
sudo chmod +x /usr/local/bin/docker-compose

# 启动 Docker 服务
sudo systemctl enable docker
sudo systemctl start docker
```

#### CentOS/RHEL 系统

```bash
# 更新系统
sudo yum update -y

# 安装 Docker
sudo yum install -y yum-utils
sudo yum-config-manager --add-repo https://download.docker.com/linux/centos/docker-ce.repo
sudo yum install -y docker-ce docker-ce-cli containerd.io

# 安装 Docker Compose
sudo curl -L "https://github.com/docker/compose/releases/latest/download/docker-compose-$(uname -s)-$(uname -m)" -o /usr/local/bin/docker-compose
sudo chmod +x /usr/local/bin/docker-compose

# 启动 Docker 服务
sudo systemctl enable docker
sudo systemctl start docker
```

### 2. 配置文件

#### 环境配置文件 (.env)

```bash
# 应用配置
APP_ENV=production
GIN_MODE=release
APP_PORT=8000

# 数据库配置
DB_HOST=postgres
DB_PORT=5432
DB_NAME=cdk_office
DB_USER=cdk_office
DB_PASSWORD=your_secure_password

# Redis配置
REDIS_HOST=redis
REDIS_PORT=6379
REDIS_PASSWORD=your_redis_password

# 外部服务配置
STIRLING_PDF_URL=http://stirling-pdf:8080
KKFILEVIEW_URL=http://kkfileview:8012

# Dify集成配置
DIFY_API_URL=https://api.dify.ai
DIFY_API_KEY=your_dify_api_key
DIFY_DATASET_ID=your_dataset_id

# SSL配置
SSL_CERT_PATH=/etc/ssl/certs/cert.pem
SSL_KEY_PATH=/etc/ssl/private/key.pem

# 域名配置
DOMAIN=your-domain.com
EMAIL=admin@your-domain.com
```

#### Nginx 配置

如果使用自定义Nginx配置，编辑 `nginx/nginx.conf`：

```nginx
server {
    listen 443 ssl http2;
    server_name your-domain.com;
    
    ssl_certificate /etc/ssl/certs/cert.pem;
    ssl_certificate_key /etc/ssl/private/key.pem;
    
    # 其他配置...
}
```

### 3. 数据库初始化

#### 使用 Docker 部署

数据库会自动初始化，但可以手动执行：

```bash
# 进入数据库容器
docker-compose exec postgres psql -U cdk_office -d cdk_office

# 或者执行SQL文件
docker-compose exec -T postgres psql -U cdk_office -d cdk_office < scripts/init.sql
```

#### 使用外部数据库

如果使用外部PostgreSQL数据库：

```sql
-- 创建数据库和用户
CREATE DATABASE cdk_office;
CREATE USER cdk_office WITH ENCRYPTED PASSWORD 'your_password';
GRANT ALL PRIVILEGES ON DATABASE cdk_office TO cdk_office;

-- 设置权限
\c cdk_office
GRANT ALL ON SCHEMA public TO cdk_office;
```

### 4. SSL证书配置

#### 使用 Let's Encrypt

```bash
# 安装 certbot
sudo apt install certbot python3-certbot-nginx

# 获取证书
sudo certbot --nginx -d your-domain.com

# 自动续期
sudo crontab -e
# 添加: 0 12 * * * /usr/bin/certbot renew --quiet
```

#### 使用自签名证书（开发环境）

```bash
# 生成自签名证书
make ssl-cert

# 或手动生成
openssl req -x509 -nodes -days 365 -newkey rsa:2048 \
    -keyout nginx/ssl/key.pem \
    -out nginx/ssl/cert.pem \
    -subj "/C=CN/ST=State/L=City/O=Organization/CN=your-domain.com"
```

## 高级部署选项

### Kubernetes 部署

#### 1. 准备 Kubernetes 集群

```bash
# 创建命名空间
kubectl create namespace cdk-office

# 应用配置
kubectl apply -f k8s-deployment.yaml
```

#### 2. 配置 ConfigMap 和 Secret

```bash
# 创建 Secret
kubectl create secret generic cdk-office-secrets \
    --from-literal=db-password=your_db_password \
    --from-literal=redis-password=your_redis_password \
    -n cdk-office

# 应用 ConfigMap
kubectl apply -f k8s/configmap.yaml
```

#### 3. 部署应用

```bash
# 部署所有服务
kubectl apply -f k8s-deployment.yaml

# 检查部署状态
kubectl get pods -n cdk-office
kubectl get services -n cdk-office
```

### VPS 一键部署

使用提供的脚本进行VPS部署：

```bash
# 下载部署脚本
curl -O https://raw.githubusercontent.com/your-org/cdk-office/main/deploy.sh
chmod +x deploy.sh

# 执行部署
./deploy.sh --domain your-domain.com --email admin@your-domain.com --full
```

部署脚本会自动：
- 安装Docker和Docker Compose
- 配置防火墙
- 设置SSL证书
- 启动所有服务
- 配置自动备份

## 生产环境优化

### 1. 性能优化

#### 数据库优化

```sql
-- PostgreSQL 配置优化
ALTER SYSTEM SET shared_buffers = '256MB';
ALTER SYSTEM SET effective_cache_size = '1GB';
ALTER SYSTEM SET maintenance_work_mem = '64MB';
ALTER SYSTEM SET checkpoint_completion_target = 0.9;
ALTER SYSTEM SET wal_buffers = '16MB';
ALTER SYSTEM SET default_statistics_target = 100;

-- 重启数据库应用配置
SELECT pg_reload_conf();
```

#### Redis 优化

```redis
# redis.conf 优化
maxmemory 512mb
maxmemory-policy allkeys-lru
save 900 1
save 300 10
save 60 10000
```

#### 应用优化

```yaml
# docker-compose.prod.yml 资源限制
services:
  cdk-office:
    deploy:
      resources:
        limits:
          memory: 2G
          cpus: '2.0'
        reservations:
          memory: 1G
          cpus: '1.0'
```

### 2. 安全配置

#### 防火墙设置

```bash
# Ubuntu UFW
sudo ufw allow 22/tcp
sudo ufw allow 80/tcp
sudo ufw allow 443/tcp
sudo ufw enable

# CentOS FirewallD
sudo firewall-cmd --permanent --add-service=ssh
sudo firewall-cmd --permanent --add-service=http
sudo firewall-cmd --permanent --add-service=https
sudo firewall-cmd --reload
```

#### 访问控制

```nginx
# Nginx IP 白名单
location /admin {
    allow 192.168.1.0/24;
    allow 10.0.0.0/8;
    deny all;
    
    proxy_pass http://cdk_office_backend;
}
```

### 3. 监控配置

#### Prometheus + Grafana

```yaml
# docker-compose.monitoring.yml
version: '3.8'
services:
  prometheus:
    image: prom/prometheus
    ports:
      - "9090:9090"
    volumes:
      - ./monitoring/prometheus.yml:/etc/prometheus/prometheus.yml
  
  grafana:
    image: grafana/grafana
    ports:
      - "3000:3000"
    environment:
      - GF_SECURITY_ADMIN_PASSWORD=admin
```

#### 日志聚合

```yaml
# docker-compose.logging.yml
services:
  elasticsearch:
    image: docker.elastic.co/elasticsearch/elasticsearch:7.14.0
    
  logstash:
    image: docker.elastic.co/logstash/logstash:7.14.0
    
  kibana:
    image: docker.elastic.co/kibana/kibana:7.14.0
    ports:
      - "5601:5601"
```

## 备份和恢复

### 自动备份

```bash
# 设置自动备份
crontab -e

# 每天凌晨2点备份
0 2 * * * /path/to/cdk-office/scripts/backup.sh

# 备份脚本内容
#!/bin/bash
BACKUP_DIR="/backup/cdk-office"
DATE=$(date +%Y%m%d_%H%M%S)

# 备份数据库
docker-compose exec -T postgres pg_dump -U cdk_office cdk_office > $BACKUP_DIR/db_$DATE.sql

# 备份文件
tar -czf $BACKUP_DIR/files_$DATE.tar.gz uploads pdf_results

# 清理30天前的备份
find $BACKUP_DIR -name "*.sql" -mtime +30 -delete
find $BACKUP_DIR -name "*.tar.gz" -mtime +30 -delete
```

### 数据恢复

```bash
# 恢复数据库
docker-compose exec -T postgres psql -U cdk_office -d cdk_office < backup/db_20240120_020000.sql

# 恢复文件
tar -xzf backup/files_20240120_020000.tar.gz
```

## 故障排除

### 常见问题

#### 1. 服务启动失败

```bash
# 查看日志
docker-compose logs cdk-office
docker-compose logs postgres
docker-compose logs redis

# 检查端口占用
netstat -tulpn | grep :8000
```

#### 2. 数据库连接失败

```bash
# 检查数据库状态
docker-compose exec postgres pg_isready -U cdk_office

# 手动连接测试
docker-compose exec postgres psql -U cdk_office -d cdk_office
```

#### 3. 文件上传失败

```bash
# 检查存储空间
df -h

# 检查文件权限
ls -la uploads/
chmod 755 uploads/
```

#### 4. SSL证书问题

```bash
# 检查证书有效性
openssl x509 -in nginx/ssl/cert.pem -text -noout

# 重新生成证书
make ssl-cert
```

### 性能监控

```bash
# 查看系统资源
htop
iotop
nethogs

# 查看容器资源使用
docker stats

# 数据库性能监控
docker-compose exec postgres psql -U cdk_office -d cdk_office -c "
SELECT query, mean_time, calls 
FROM pg_stat_statements 
ORDER BY mean_time DESC 
LIMIT 10;"
```

## 升级指南

### 滚动升级

```bash
# 备份当前版本
make backup

# 拉取最新镜像
docker-compose pull

# 滚动更新
docker-compose up -d --force-recreate

# 验证升级
make health
```

### 版本回滚

```bash
# 回滚到上一个版本
docker-compose down
docker-compose up -d --force-recreate

# 如果需要，恢复数据库
make restore BACKUP_FILE=backup_file.sql
```

## 支持和维护

### 技术支持

- 文档: https://docs.your-domain.com
- Issues: https://github.com/your-org/cdk-office/issues
- 邮箱: support@your-domain.com

### 定期维护

- 每周检查系统日志和性能指标
- 每月更新系统和依赖包
- 每季度进行安全审计
- 每年进行灾难恢复演练