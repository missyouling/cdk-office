# CDK-Office Makefile

# 变量定义
APP_NAME := cdk-office
VERSION := 1.0.0
BUILD_DIR := build
DIST_DIR := dist

# Go相关变量
GO := go
GO_BUILD := $(GO) build
GO_CLEAN := $(GO) clean
GO_TEST := $(GO) test
GO_GET := $(GO) get
GO_MOD := $(GO) mod
GO_FMT := $(GO) fmt
GO_VET := $(GO) vet

# 默认目标
.PHONY: all
all: clean fmt vet build

# 清理构建文件
.PHONY: clean
clean:
	@echo "Cleaning build files..."
	$(GO_CLEAN)
	rm -rf $(BUILD_DIR)
	rm -rf $(DIST_DIR)
	@echo "Clean completed"

# 格式化代码
.PHONY: fmt
fmt:
	@echo "Formatting code..."
	$(GO_FMT) ./...
	@echo "Format completed"

# 代码检查
.PHONY: vet
vet:
	@echo "Vetting code..."
	$(GO_VET) ./...
	@echo "Vet completed"

# 构建项目
.PHONY: build
build:
	@echo "Building $(APP_NAME)..."
	mkdir -p $(BUILD_DIR)
	$(GO_BUILD) -o $(BUILD_DIR)/$(APP_NAME) .
	@echo "Build completed"

# 构建VPS优化版本
.PHONY: build-vps
build-vps:
	@echo "Building $(APP_NAME) for VPS..."
	mkdir -p $(BUILD_DIR)
	$(GO_BUILD) -ldflags="-w -s" -o $(BUILD_DIR)/$(APP_NAME)-vps .
	@echo "VPS build completed"

# 运行开发服务器
.PHONY: run
run:
	@echo "Running $(APP_NAME) in development mode..."
	$(GO_RUN) .

# 运行API服务器
.PHONY: run-api
run-api:
	@echo "Running $(APP_NAME) API server..."
	$(GO_RUN) . api

# 运行调度器
.PHONY: run-scheduler
run-scheduler:
	@echo "Running $(APP_NAME) scheduler..."
	$(GO_RUN) . scheduler

# 运行工作进程
.PHONY: run-worker
run-worker:
	@echo "Running $(APP_NAME) worker..."
	$(GO_RUN) . worker

# 运行测试
.PHONY: test
test:
	@echo "Running tests..."
	$(GO_TEST) -v ./...

# 运行测试覆盖率
.PHONY: test-coverage
test-coverage:
	@echo "Running tests with coverage..."
	$(GO_TEST) -coverprofile=coverage.out ./...
	$(GO) tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report generated: coverage.html"

# 更新依赖
.PHONY: deps
deps:
	@echo "Updating dependencies..."
	$(GO_GET) -u ./...
	$(GO_MOD) tidy
	@echo "Dependencies updated"

# 安装依赖
.PHONY: install
install:
	@echo "Installing dependencies..."
	$(GO_MOD) download
	@echo "Dependencies installed"

# 生成文档
.PHONY: docs
docs:
	@echo "Generating documentation..."
	swag init -g main.go
	@echo "Documentation generated"

# 部署到VPS
.PHONY: deploy-vps
deploy-vps:
	@echo "Deploying to VPS..."
	./scripts/deploy-vps.sh
	@echo "Deployment completed"

# 停止VPS服务
.PHONY: stop-vps
stop-vps:
	@echo "Stopping VPS service..."
	./scripts/stop-vps.sh
	@echo "Service stopped"

# 帮助信息
.PHONY: help
help:
	@echo "CDK-Office Makefile"
	@echo ""
	@echo "Usage:"
	@echo "  make              # Clean, format, vet and build"
	@echo "  make clean        # Clean build files"
	@echo "  make fmt          # Format code"
	@echo "  make vet          # Vet code"
	@echo "  make build        # Build project"
	@echo "  make build-vps    # Build VPS optimized version"
	@echo "  make run          # Run development server"
	@echo "  make run-api      # Run API server"
	@echo "  make run-scheduler # Run scheduler"
	@echo "  make run-worker   # Run worker"
	@echo "  make test         # Run tests"
	@echo "  make test-coverage # Run tests with coverage"
	@echo "  make deps         # Update dependencies"
	@echo "  make install      # Install dependencies"
	@echo "  make docs         # Generate documentation"
	@echo "  make deploy-vps   # Deploy to VPS"
	@echo "  make stop-vps     # Stop VPS service"
	@echo "  make help         # Show this help

# 包含测试相关命令
include test.mk"