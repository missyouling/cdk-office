# CDK-Office Makefile
# 提供便捷的构建、部署和管理命令

# 默认变量
PROJECT_NAME := cdk-office
VERSION := $(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
COMMIT_SHA := $(shell git rev-parse --short HEAD 2>/dev/null || echo "unknown")
BUILD_TIME := $(shell date -u +"%Y-%m-%dT%H:%M:%SZ")

# Docker相关变量
DOCKER_REGISTRY := your-registry.com  # 替换为实际仓库地址
DOCKER_IMAGE := $(DOCKER_REGISTRY)/$(PROJECT_NAME)
DOCKER_TAG := $(VERSION)

# 环境变量
DOMAIN := localhost
EMAIL := admin@example.com

# 颜色定义
GREEN := \033[0;32m
YELLOW := \033[1;33m
RED := \033[0;31m
NC := \033[0m # No Color

.PHONY: help build push deploy clean logs status backup restore test

# 默认目标
.DEFAULT_GOAL := help

# 帮助信息
help: ## 显示帮助信息
	@echo "CDK-Office 部署工具"
	@echo "用法: make <target>"
	@echo ""
	@echo "可用命令:"
	@awk 'BEGIN {FS = ":.*?## "} /^[a-zA-Z_-]+:.*?## / {printf "  $(GREEN)%-15s$(NC) %s\n", $$1, $$2}' $(MAKEFILE_LIST)

# 构建相关命令
build: ## 构建Docker镜像
	@echo "$(GREEN)构建Docker镜像...$(NC)"
	docker build \
		--build-arg VERSION=$(VERSION) \
		--build-arg COMMIT_SHA=$(COMMIT_SHA) \
		--build-arg BUILD_TIME=$(BUILD_TIME) \
		-t $(PROJECT_NAME):$(DOCKER_TAG) \
		-t $(PROJECT_NAME):latest \
		-f Dockerfile .
	@echo "$(GREEN)构建完成: $(PROJECT_NAME):$(DOCKER_TAG)$(NC)"

build-prod: ## 构建生产环境镜像
	@echo "$(GREEN)构建生产环境镜像...$(NC)"
	docker build \
		--build-arg VERSION=$(VERSION) \
		--build-arg COMMIT_SHA=$(COMMIT_SHA) \
		--build-arg BUILD_TIME=$(BUILD_TIME) \
		--target runtime \
		-t $(DOCKER_IMAGE):$(DOCKER_TAG) \
		-t $(DOCKER_IMAGE):latest \
		-f Dockerfile .
	@echo "$(GREEN)生产镜像构建完成$(NC)"

push: build-prod ## 推送镜像到仓库
	@echo "$(GREEN)推送镜像到仓库...$(NC)"
	docker push $(DOCKER_IMAGE):$(DOCKER_TAG)
	docker push $(DOCKER_IMAGE):latest
	@echo "$(GREEN)镜像推送完成$(NC)"

# 本地开发命令
dev: ## 启动开发环境
	@echo "$(GREEN)启动开发环境...$(NC)"
	docker-compose -f docker-compose.dev.yml up -d
	@echo "$(GREEN)开发环境已启动$(NC)"
	@echo "访问地址: http://localhost:8000"

dev-down: ## 停止开发环境
	@echo "$(YELLOW)停止开发环境...$(NC)"
	docker-compose -f docker-compose.dev.yml down

dev-logs: ## 查看开发环境日志
	docker-compose -f docker-compose.dev.yml logs -f

# 生产部署命令
deploy: ## 部署到生产环境（Docker Compose）
	@echo "$(GREEN)部署到生产环境...$(NC)"
	@if [ ! -f .env ]; then \
		echo "创建环境配置文件..."; \
		cp .env.example .env; \
	fi
	docker-compose -f docker-compose.prod.yml pull
	docker-compose -f docker-compose.prod.yml up -d
	@echo "$(GREEN)生产环境部署完成$(NC)"

deploy-core: ## 仅部署核心服务
	@echo "$(GREEN)部署核心服务...$(NC)"
	docker-compose -f docker-compose.prod.yml --profile core up -d

deploy-full: ## 部署完整服务栈
	@echo "$(GREEN)部署完整服务栈...$(NC)"
	docker-compose -f docker-compose.prod.yml --profile full --profile monitoring up -d

deploy-vps: ## 使用脚本部署到VPS
	@echo "$(GREEN)部署到VPS...$(NC)"
	chmod +x deploy.sh
	./deploy.sh --domain $(DOMAIN) --email $(EMAIL) --full

# Kubernetes部署命令
k8s-deploy: ## 部署到Kubernetes
	@echo "$(GREEN)部署到Kubernetes...$(NC)"
	kubectl apply -f k8s-deployment.yaml
	@echo "$(GREEN)Kubernetes部署完成$(NC)"

k8s-status: ## 查看Kubernetes部署状态
	kubectl get all -n cdk-office

k8s-logs: ## 查看Kubernetes日志
	kubectl logs -f -n cdk-office -l app=cdk-office-app

k8s-delete: ## 删除Kubernetes部署
	@echo "$(YELLOW)删除Kubernetes部署...$(NC)"
	kubectl delete -f k8s-deployment.yaml

# 服务管理命令
start: ## 启动服务
	@echo "$(GREEN)启动服务...$(NC)"
	docker-compose up -d

stop: ## 停止服务
	@echo "$(YELLOW)停止服务...$(NC)"
	docker-compose down

restart: ## 重启服务
	@echo "$(YELLOW)重启服务...$(NC)"
	docker-compose restart

status: ## 查看服务状态
	@echo "$(GREEN)服务状态:$(NC)"
	docker-compose ps

logs: ## 查看服务日志
	docker-compose logs -f

logs-app: ## 仅查看应用日志
	docker-compose logs -f cdk-office

# 数据管理命令
backup: ## 备份数据
	@echo "$(GREEN)开始备份数据...$(NC)"
	@mkdir -p backup
	@BACKUP_FILE="backup/cdk-office-$(shell date +%Y%m%d_%H%M%S).sql"; \
	docker-compose exec -T postgres pg_dump -U cdk_office cdk_office > $$BACKUP_FILE; \
	echo "数据库备份完成: $$BACKUP_FILE"
	@tar -czf backup/files-$(shell date +%Y%m%d_%H%M%S).tar.gz uploads pdf_results
	@echo "$(GREEN)文件备份完成$(NC)"

restore: ## 恢复数据 (需要指定备份文件: make restore BACKUP_FILE=backup.sql)
	@if [ -z "$(BACKUP_FILE)" ]; then \
		echo "$(RED)请指定备份文件: make restore BACKUP_FILE=backup.sql$(NC)"; \
		exit 1; \
	fi
	@echo "$(YELLOW)恢复数据库: $(BACKUP_FILE)$(NC)"
	docker-compose exec -T postgres psql -U cdk_office -d cdk_office < $(BACKUP_FILE)
	@echo "$(GREEN)数据恢复完成$(NC)"

db-shell: ## 连接数据库Shell
	docker-compose exec postgres psql -U cdk_office -d cdk_office

redis-shell: ## 连接Redis Shell
	docker-compose exec redis redis-cli

# 维护命令
clean: ## 清理Docker资源
	@echo "$(YELLOW)清理Docker资源...$(NC)"
	docker system prune -f
	docker volume prune -f
	@echo "$(GREEN)清理完成$(NC)"

clean-all: ## 清理所有Docker资源（包括镜像）
	@echo "$(RED)清理所有Docker资源...$(NC)"
	docker system prune -af
	docker volume prune -f
	@echo "$(GREEN)清理完成$(NC)"

update: ## 更新服务到最新版本
	@echo "$(GREEN)更新服务...$(NC)"
	$(MAKE) backup
	docker-compose pull
	docker-compose up -d
	@echo "$(GREEN)更新完成$(NC)"

# 健康检查
health: ## 检查服务健康状态
	@echo "$(GREEN)检查服务健康状态...$(NC)"
	@curl -f http://localhost:8000/api/v1/health || echo "$(RED)服务不健康$(NC)"
	@docker-compose ps

# 监控命令
monitor: ## 打开监控面板
	@echo "$(GREEN)监控服务地址:$(NC)"
	@echo "Grafana: http://localhost:3000"
	@echo "Prometheus: http://localhost:9090"

# 测试命令
test: ## 运行测试
	@echo "$(GREEN)运行测试...$(NC)"
	go test -v ./...

test-coverage: ## 运行测试并生成覆盖率报告
	@echo "$(GREEN)生成测试覆盖率报告...$(NC)"
	go test -v -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out -o coverage.html
	@echo "覆盖率报告: coverage.html"

# 代码质量
lint: ## 代码检查
	@echo "$(GREEN)进行代码检查...$(NC)"
	golangci-lint run

fmt: ## 格式化代码
	@echo "$(GREEN)格式化代码...$(NC)"
	go fmt ./...
	goimports -w .

# 开发工具
docs: ## 生成API文档
	@echo "$(GREEN)生成API文档...$(NC)"
	swag init -g main.go
	@echo "API文档已生成"

deps: ## 安装依赖
	@echo "$(GREEN)安装依赖...$(NC)"
	go mod download
	go mod tidy

init: ## 初始化项目（首次运行）
	@echo "$(GREEN)初始化项目...$(NC)"
	$(MAKE) deps
	cp .env.example .env
	@echo "$(YELLOW)请编辑 .env 文件配置环境变量$(NC)"
	@echo "$(GREEN)初始化完成，可以运行: make dev$(NC)"

# SSL证书管理
ssl-cert: ## 生成自签名SSL证书
	@echo "$(GREEN)生成SSL证书...$(NC)"
	@mkdir -p nginx/ssl
	openssl req -x509 -nodes -days 365 -newkey rsa:2048 \
		-keyout nginx/ssl/key.pem \
		-out nginx/ssl/cert.pem \
		-subj "/C=CN/ST=State/L=City/O=Organization/CN=$(DOMAIN)"
	@echo "$(GREEN)SSL证书生成完成$(NC)"

# 配置模板
config: ## 生成配置文件
	@echo "$(GREEN)生成配置文件...$(NC)"
	@if [ ! -f .env ]; then \
		cp .env.example .env; \
		echo "已创建 .env 文件"; \
	fi
	@if [ ! -f nginx/ssl/cert.pem ]; then \
		$(MAKE) ssl-cert; \
	fi

# 发布相关
release: ## 创建发布版本
	@echo "$(GREEN)创建发布版本...$(NC)"
	@if [ -z "$(TAG)" ]; then \
		echo "$(RED)请指定版本号: make release TAG=v1.0.0$(NC)"; \
		exit 1; \
	fi
	git tag -a $(TAG) -m "Release $(TAG)"
	git push origin $(TAG)
	@echo "$(GREEN)发布 $(TAG) 完成$(NC)"

# 快速命令别名
up: start     ## 启动服务（别名）
down: stop    ## 停止服务（别名）
ps: status    ## 查看状态（别名）

# 完整工作流
full-deploy: config build deploy health ## 完整部署流程
	@echo "$(GREEN)完整部署流程完成！$(NC)"
	@echo "访问地址: https://$(DOMAIN)"

# 信息显示
info: ## 显示项目信息
	@echo "$(GREEN)CDK-Office 项目信息:$(NC)"
	@echo "项目名称: $(PROJECT_NAME)"
	@echo "版本: $(VERSION)"
	@echo "提交: $(COMMIT_SHA)"
	@echo "构建时间: $(BUILD_TIME)"
	@echo "Docker镜像: $(DOCKER_IMAGE):$(DOCKER_TAG)"
	@echo "域名: $(DOMAIN)"