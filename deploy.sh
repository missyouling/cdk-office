#!/bin/bash

# CDK-Office VPS 部署脚本
# 支持 Ubuntu 20.04+, CentOS 8+, Debian 11+
# 用途：一键部署 CDK-Office 到 VPS 服务器

set -e  # 遇到错误立即退出

# 配置变量
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_NAME="cdk-office"
DOMAIN="${DOMAIN:-localhost}"
EMAIL="${EMAIL:-admin@example.com}"
DB_PASSWORD="${DB_PASSWORD:-$(openssl rand -base64 32)}"
REDIS_PASSWORD="${REDIS_PASSWORD:-$(openssl rand -base64 32)}"
GRAFANA_PASSWORD="${GRAFANA_PASSWORD:-$(openssl rand -base64 16)}"

# 颜色定义
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# 日志函数
log_info() {
    echo -e "${GREEN}[INFO]${NC} $1"
}

log_warn() {
    echo -e "${YELLOW}[WARN]${NC} $1"
}

log_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

log_step() {
    echo -e "${BLUE}[STEP]${NC} $1"
}

# 检查root权限
check_root() {
    if [[ $EUID -ne 0 ]]; then
        log_error "此脚本需要root权限运行"
        log_info "请使用: sudo $0"
        exit 1
    fi
}

# 检测操作系统
detect_os() {
    if [[ -f /etc/redhat-release ]]; then
        OS="centos"
        PACKAGE_MANAGER="yum"
    elif [[ -f /etc/lsb-release ]]; then
        OS="ubuntu"
        PACKAGE_MANAGER="apt"
    elif [[ -f /etc/debian_version ]]; then
        OS="debian"
        PACKAGE_MANAGER="apt"
    else
        log_error "不支持的操作系统"
        exit 1
    fi
    log_info "检测到操作系统: $OS"
}

# 更新系统
update_system() {
    log_step "更新系统..."
    case $PACKAGE_MANAGER in
        "apt")
            apt update -y
            apt upgrade -y
            apt install -y curl wget git unzip software-properties-common apt-transport-https ca-certificates gnupg lsb-release
            ;;
        "yum")
            yum update -y
            yum install -y curl wget git unzip epel-release
            ;;
    esac
}

# 安装Docker
install_docker() {
    log_step "安装Docker..."
    
    if command -v docker &> /dev/null; then
        log_info "Docker已安装，版本: $(docker --version)"
        return
    fi
    
    case $OS in
        "ubuntu"|"debian")
            # 添加Docker官方GPG密钥
            curl -fsSL https://download.docker.com/linux/$OS/gpg | gpg --dearmor -o /usr/share/keyrings/docker-archive-keyring.gpg
            
            # 添加Docker仓库
            echo "deb [arch=$(dpkg --print-architecture) signed-by=/usr/share/keyrings/docker-archive-keyring.gpg] https://download.docker.com/linux/$OS $(lsb_release -cs) stable" > /etc/apt/sources.list.d/docker.list
            
            # 安装Docker
            apt update -y
            apt install -y docker-ce docker-ce-cli containerd.io docker-compose-plugin
            ;;
        "centos")
            # 添加Docker仓库
            yum-config-manager --add-repo https://download.docker.com/linux/centos/docker-ce.repo
            
            # 安装Docker
            yum install -y docker-ce docker-ce-cli containerd.io docker-compose-plugin
            ;;
    esac
    
    # 启动Docker服务
    systemctl start docker
    systemctl enable docker
    
    # 添加当前用户到docker组（如果有非root用户）
    if [[ -n "${SUDO_USER}" ]]; then
        usermod -aG docker "${SUDO_USER}"
        log_info "已将用户 ${SUDO_USER} 添加到docker组"
    fi
    
    log_info "Docker安装完成，版本: $(docker --version)"
}

# 安装Docker Compose
install_docker_compose() {
    log_step "安装Docker Compose..."
    
    if command -v docker-compose &> /dev/null; then
        log_info "Docker Compose已安装，版本: $(docker-compose --version)"
        return
    fi
    
    # 获取最新版本
    COMPOSE_VERSION=$(curl -s https://api.github.com/repos/docker/compose/releases/latest | grep -Po '"tag_name": "\K.*?(?=")')
    
    # 下载并安装
    curl -L "https://github.com/docker/compose/releases/download/${COMPOSE_VERSION}/docker-compose-$(uname -s)-$(uname -m)" -o /usr/local/bin/docker-compose
    chmod +x /usr/local/bin/docker-compose
    
    # 创建软链接
    ln -sf /usr/local/bin/docker-compose /usr/bin/docker-compose
    
    log_info "Docker Compose安装完成，版本: $(docker-compose --version)"
}

# 配置防火墙
configure_firewall() {
    log_step "配置防火墙..."
    
    case $OS in
        "ubuntu"|"debian")
            if command -v ufw &> /dev/null; then
                ufw --force enable
                ufw allow ssh
                ufw allow 80/tcp
                ufw allow 443/tcp
                ufw allow 8000/tcp  # CDK-Office
                ufw reload
                log_info "UFW防火墙配置完成"
            fi
            ;;
        "centos")
            if command -v firewall-cmd &> /dev/null; then
                systemctl start firewalld
                systemctl enable firewalld
                firewall-cmd --permanent --add-service=ssh
                firewall-cmd --permanent --add-service=http
                firewall-cmd --permanent --add-service=https
                firewall-cmd --permanent --add-port=8000/tcp
                firewall-cmd --reload
                log_info "firewalld防火墙配置完成"
            fi
            ;;
    esac
}

# 创建项目目录
create_project_structure() {
    log_step "创建项目目录结构..."
    
    PROJECT_ROOT="/opt/$PROJECT_NAME"
    mkdir -p "$PROJECT_ROOT"/{logs,uploads,pdf_results,cache,backup,ssl,monitoring,scripts}
    mkdir -p "$PROJECT_ROOT"/nginx/{conf.d,ssl,logs}
    mkdir -p "$PROJECT_ROOT"/stirling-pdf/{configs,logs,customFiles}
    mkdir -p "$PROJECT_ROOT"/kkfileview/{files,logs,config}
    mkdir -p "$PROJECT_ROOT"/monitoring/{prometheus,grafana,rules}
    mkdir -p "$PROJECT_ROOT"/elasticsearch/config
    mkdir -p "$PROJECT_ROOT"/redis
    
    # 设置权限
    chown -R 1001:1001 "$PROJECT_ROOT"/{logs,uploads,pdf_results,cache}
    chmod -R 755 "$PROJECT_ROOT"
    
    log_info "项目目录结构创建完成: $PROJECT_ROOT"
}

# 生成环境配置文件
generate_env_config() {
    log_step "生成环境配置文件..."
    
    cat > "$PROJECT_ROOT/.env" << EOF
# CDK-Office 生产环境配置
# 生成时间: $(date)

# 基础配置
COMPOSE_PROJECT_NAME=$PROJECT_NAME
DOMAIN=$DOMAIN
VERSION=latest
COMMIT_SHA=$(git rev-parse --short HEAD 2>/dev/null || echo "unknown")
BUILD_TIME=$(date -u +"%Y-%m-%dT%H:%M:%SZ")

# 端口配置
HTTP_PORT=80
HTTPS_PORT=443
APP_PORT=8000
DB_PORT=5432
REDIS_PORT=6379
STIRLING_PDF_PORT=8081
KKFILEVIEW_PORT=8012
PROMETHEUS_PORT=9090
GRAFANA_PORT=3000
ELASTICSEARCH_PORT=9200

# 数据库配置
DB_NAME=cdk_office
DB_USER=cdk_office
DB_PASSWORD=$DB_PASSWORD

# Redis配置
REDIS_PASSWORD=$REDIS_PASSWORD

# 监控配置
GRAFANA_PASSWORD=$GRAFANA_PASSWORD
GRAFANA_SECRET_KEY=$(openssl rand -base64 32)

# SSL配置
SSL_EMAIL=$EMAIL
EOF
    
    log_info "环境配置文件生成完成: $PROJECT_ROOT/.env"
    log_warn "请妥善保管数据库密码: $DB_PASSWORD"
    log_warn "请妥善保管Redis密码: $REDIS_PASSWORD"
    log_warn "请妥善保管Grafana密码: $GRAFANA_PASSWORD"
}

# 生成SSL证书
generate_ssl_cert() {
    log_step "生成SSL证书..."
    
    if [[ "$DOMAIN" == "localhost" ]]; then
        # 生成自签名证书
        openssl req -x509 -nodes -days 365 -newkey rsa:2048 \
            -keyout "$PROJECT_ROOT/ssl/key.pem" \
            -out "$PROJECT_ROOT/ssl/cert.pem" \
            -subj "/C=CN/ST=State/L=City/O=Organization/CN=$DOMAIN"
        log_info "自签名SSL证书生成完成"
    else
        # 使用Let's Encrypt（需要域名指向服务器）
        if command -v certbot &> /dev/null; then
            certbot certonly --standalone --agree-tos --email "$EMAIL" -d "$DOMAIN"
            ln -sf "/etc/letsencrypt/live/$DOMAIN/fullchain.pem" "$PROJECT_ROOT/ssl/cert.pem"
            ln -sf "/etc/letsencrypt/live/$DOMAIN/privkey.pem" "$PROJECT_ROOT/ssl/key.pem"
            log_info "Let's Encrypt SSL证书配置完成"
        else
            log_warn "certbot未安装，生成自签名证书"
            generate_ssl_cert
        fi
    fi
}

# 生成监控配置
generate_monitoring_config() {
    log_step "生成监控配置..."
    
    # Prometheus配置
    cat > "$PROJECT_ROOT/monitoring/prometheus.yml" << EOF
global:
  scrape_interval: 15s
  evaluation_interval: 15s

rule_files:
  - "/etc/prometheus/rules/*.yml"

scrape_configs:
  - job_name: 'prometheus'
    static_configs:
      - targets: ['localhost:9090']

  - job_name: 'cdk-office'
    static_configs:
      - targets: ['cdk-office-app:8000']
    metrics_path: '/api/v1/metrics'

  - job_name: 'nginx'
    static_configs:
      - targets: ['nginx:8080']
    metrics_path: '/metrics'

  - job_name: 'postgres'
    static_configs:
      - targets: ['postgres:5432']

  - job_name: 'redis'
    static_configs:
      - targets: ['redis:6379']
EOF
    
    # Grafana数据源配置
    mkdir -p "$PROJECT_ROOT/monitoring/grafana/provisioning"/{datasources,dashboards}
    
    cat > "$PROJECT_ROOT/monitoring/grafana/provisioning/datasources/prometheus.yml" << EOF
apiVersion: 1
datasources:
  - name: Prometheus
    type: prometheus
    access: proxy
    url: http://prometheus:9090
    isDefault: true
EOF
    
    log_info "监控配置生成完成"
}

# 部署应用
deploy_application() {
    log_step "部署CDK-Office应用..."
    
    cd "$PROJECT_ROOT"
    
    # 复制配置文件到项目目录
    if [[ -f "$SCRIPT_DIR/docker-compose.prod.yml" ]]; then
        cp "$SCRIPT_DIR/docker-compose.prod.yml" "$PROJECT_ROOT/docker-compose.yml"
    fi
    
    if [[ -f "$SCRIPT_DIR/nginx/nginx.conf" ]]; then
        cp "$SCRIPT_DIR/nginx/nginx.conf" "$PROJECT_ROOT/nginx/"
    fi
    
    # 拉取镜像
    docker-compose pull
    
    # 启动核心服务
    docker-compose --profile core up -d
    
    # 等待服务启动
    log_info "等待服务启动..."
    sleep 30
    
    # 检查服务状态
    if docker-compose ps | grep -q "Up"; then
        log_info "CDK-Office核心服务启动成功"
    else
        log_error "服务启动失败，请检查日志"
        docker-compose logs
        exit 1
    fi
}

# 启动完整服务（可选）
deploy_full_stack() {
    log_step "部署完整服务栈（包括监控）..."
    
    cd "$PROJECT_ROOT"
    
    # 启动所有服务
    docker-compose --profile full --profile monitoring up -d
    
    log_info "完整服务栈部署完成"
    log_info "服务访问地址："
    log_info "  - CDK-Office: https://$DOMAIN"
    log_info "  - PDF工具: https://$DOMAIN/pdf-tools"
    log_info "  - 文件预览: https://$DOMAIN/preview"
    log_info "  - Grafana监控: http://$DOMAIN:3000 (admin/$GRAFANA_PASSWORD)"
    log_info "  - Prometheus: http://$DOMAIN:9090"
}

# 创建管理脚本
create_management_scripts() {
    log_step "创建管理脚本..."
    
    # 备份脚本
    cat > "$PROJECT_ROOT/scripts/backup.sh" << 'EOF'
#!/bin/bash
# CDK-Office 备份脚本

BACKUP_DIR="/opt/cdk-office/backup"
DATE=$(date +%Y%m%d_%H%M%S)

# 创建备份目录
mkdir -p "$BACKUP_DIR"

# 备份数据库
docker-compose exec -T postgres pg_dump -U cdk_office cdk_office > "$BACKUP_DIR/database_$DATE.sql"

# 备份文件
tar -czf "$BACKUP_DIR/files_$DATE.tar.gz" -C /opt/cdk-office uploads pdf_results

# 清理旧备份（保留7天）
find "$BACKUP_DIR" -name "*.sql" -mtime +7 -delete
find "$BACKUP_DIR" -name "*.tar.gz" -mtime +7 -delete

echo "备份完成: $DATE"
EOF
    
    # 更新脚本
    cat > "$PROJECT_ROOT/scripts/update.sh" << 'EOF'
#!/bin/bash
# CDK-Office 更新脚本

cd /opt/cdk-office

# 备份
./scripts/backup.sh

# 拉取最新镜像
docker-compose pull

# 重启服务
docker-compose down
docker-compose up -d

echo "更新完成"
EOF
    
    # 日志清理脚本
    cat > "$PROJECT_ROOT/scripts/cleanup.sh" << 'EOF'
#!/bin/bash
# CDK-Office 日志清理脚本

# 清理Docker日志
docker system prune -f

# 清理应用日志（保留30天）
find /opt/cdk-office/logs -name "*.log" -mtime +30 -delete

# 清理Nginx日志
find /opt/cdk-office/nginx/logs -name "*.log" -mtime +7 -delete

echo "清理完成"
EOF
    
    chmod +x "$PROJECT_ROOT/scripts"/*.sh
    
    log_info "管理脚本创建完成"
}

# 配置系统服务
setup_systemd_service() {
    log_step "配置系统服务..."
    
    cat > /etc/systemd/system/cdk-office.service << EOF
[Unit]
Description=CDK-Office Docker Compose Service
Requires=docker.service
After=docker.service

[Service]
Type=oneshot
RemainAfterExit=yes
WorkingDirectory=/opt/cdk-office
ExecStart=/usr/local/bin/docker-compose up -d
ExecStop=/usr/local/bin/docker-compose down
TimeoutStartSec=0

[Install]
WantedBy=multi-user.target
EOF
    
    systemctl daemon-reload
    systemctl enable cdk-office.service
    
    log_info "系统服务配置完成"
}

# 配置定时任务
setup_cron_jobs() {
    log_step "配置定时任务..."
    
    # 添加cron任务
    (crontab -l 2>/dev/null; echo "0 2 * * * /opt/cdk-office/scripts/backup.sh") | crontab -
    (crontab -l 2>/dev/null; echo "0 3 * * 0 /opt/cdk-office/scripts/cleanup.sh") | crontab -
    
    log_info "定时任务配置完成（每天2点备份，每周日3点清理）"
}

# 显示部署信息
show_deployment_info() {
    log_step "部署完成！"
    
    echo
    echo "========================="
    echo "  CDK-Office 部署信息"
    echo "========================="
    echo "项目目录: $PROJECT_ROOT"
    echo "域名: $DOMAIN"
    echo "数据库密码: $DB_PASSWORD"
    echo "Redis密码: $REDIS_PASSWORD"
    echo "Grafana密码: $GRAFANA_PASSWORD"
    echo
    echo "服务访问地址："
    echo "  - 主应用: https://$DOMAIN"
    echo "  - API文档: https://$DOMAIN/api/swagger/"
    echo "  - 健康检查: https://$DOMAIN/api/v1/health"
    echo
    echo "管理命令："
    echo "  - 查看状态: docker-compose ps"
    echo "  - 查看日志: docker-compose logs -f"
    echo "  - 重启服务: docker-compose restart"
    echo "  - 停止服务: docker-compose down"
    echo "  - 备份数据: ./scripts/backup.sh"
    echo
    echo "配置文件位置："
    echo "  - 环境配置: $PROJECT_ROOT/.env"
    echo "  - Nginx配置: $PROJECT_ROOT/nginx/nginx.conf"
    echo "  - 监控配置: $PROJECT_ROOT/monitoring/"
    echo
    log_info "请妥善保管密码信息！"
}

# 主函数
main() {
    log_info "开始部署CDK-Office到VPS..."
    
    # 检查参数
    while [[ $# -gt 0 ]]; do
        case $1 in
            --domain)
                DOMAIN="$2"
                shift 2
                ;;
            --email)
                EMAIL="$2"
                shift 2
                ;;
            --full)
                DEPLOY_FULL=true
                shift
                ;;
            --help)
                echo "使用方法: $0 [选项]"
                echo "选项："
                echo "  --domain DOMAIN   设置域名（默认: localhost）"
                echo "  --email EMAIL     设置邮箱（用于SSL证书）"
                echo "  --full           部署完整服务栈（包括监控）"
                echo "  --help           显示帮助"
                exit 0
                ;;
            *)
                log_error "未知参数: $1"
                exit 1
                ;;
        esac
    done
    
    check_root
    detect_os
    update_system
    install_docker
    install_docker_compose
    configure_firewall
    create_project_structure
    generate_env_config
    generate_ssl_cert
    generate_monitoring_config
    create_management_scripts
    setup_systemd_service
    setup_cron_jobs
    
    # 部署应用
    deploy_application
    
    if [[ "$DEPLOY_FULL" == "true" ]]; then
        deploy_full_stack
    fi
    
    show_deployment_info
    
    log_info "CDK-Office部署完成！"
}

# 脚本入口
if [[ "${BASH_SOURCE[0]}" == "${0}" ]]; then
    main "$@"
fi