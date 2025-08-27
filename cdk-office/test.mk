# 测试相关命令
test: ## 运行所有测试
	@echo "$(GREEN)运行测试套件...$(NC)"
	chmod +x scripts/run-tests.sh
	./scripts/run-tests.sh

test-unit: ## 仅运行单元测试
	@echo "$(GREEN)运行单元测试...$(NC)"
	go test -v -race ./internal/...

test-integration: ## 仅运行集成测试
	@echo "$(GREEN)运行集成测试...$(NC)"
	export RUN_INTEGRATION_TESTS=1 && go test -v -race -timeout=5m ./tests/...

test-coverage: ## 运行测试并生成覆盖率报告
	@echo "$(GREEN)生成测试覆盖率报告...$(NC)"
	mkdir -p test-results
	go test -v -coverprofile=test-results/coverage.out ./...
	go tool cover -html=test-results/coverage.out -o test-results/coverage.html
	go tool cover -func=test-results/coverage.out
	@echo "$(GREEN)覆盖率报告已生成: test-results/coverage.html$(NC)"

test-bench: ## 运行性能基准测试
	@echo "$(GREEN)运行性能基准测试...$(NC)"
	mkdir -p test-results
	go test -bench=. -benchmem -run=^$ ./... | tee test-results/benchmark.txt

test-race: ## 运行竞态条件检测
	@echo "$(GREEN)运行竞态条件检测...$(NC)"
	go test -race ./...

test-short: ## 运行快速测试（跳过长时间运行的测试）
	@echo "$(GREEN)运行快速测试...$(NC)"
	go test -short ./...

test-clean: ## 清理测试结果
	@echo "$(YELLOW)清理测试结果...$(NC)"
	rm -rf test-results/
	go clean -testcache

test-watch: ## 监视文件变化并自动运行测试
	@echo "$(GREEN)启动测试监视模式...$(NC)"
	@if command -v watchexec >/dev/null 2>&1; then \
		watchexec -e go -r "make test-unit"; \
	elif command -v inotifywait >/dev/null 2>&1; then \
		while inotifywait -e modify -r .; do make test-unit; done; \
	else \
		echo "$(RED)需要安装 watchexec 或 inotify-tools$(NC)"; \
	fi

# 代码质量检查
quality: ## 运行代码质量检查
	@echo "$(GREEN)运行代码质量检查...$(NC)"
	@if command -v golangci-lint >/dev/null 2>&1; then \
		golangci-lint run --timeout=5m; \
	else \
		echo "$(YELLOW)golangci-lint 未安装，跳过代码检查$(NC)"; \
		echo "安装命令: curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b \$$(go env GOPATH)/bin v1.54.2"; \
	fi

security: ## 运行安全扫描
	@echo "$(GREEN)运行安全扫描...$(NC)"
	@if command -v gosec >/dev/null 2>&1; then \
		gosec ./...; \
	else \
		echo "$(YELLOW)gosec 未安装，跳过安全扫描$(NC)"; \
		echo "安装命令: go install github.com/securecodewarrior/gosec/v2/cmd/gosec@latest"; \
	fi

# 持续集成相关
ci-test: ## CI环境测试
	@echo "$(GREEN)运行CI测试...$(NC)"
	go version
	go mod download
	go mod verify
	go vet ./...
	go test -v -race -coverprofile=coverage.out ./...
	go tool cover -func=coverage.out

ci-build: ## CI环境构建
	@echo "$(GREEN)CI环境构建...$(NC)"
	CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o cdk-office .

# 测试环境管理
test-env-up: ## 启动测试环境
	@echo "$(GREEN)启动测试环境...$(NC)"
	docker-compose -f docker-compose.test.yml up -d

test-env-down: ## 停止测试环境
	@echo "$(YELLOW)停止测试环境...$(NC)"
	docker-compose -f docker-compose.test.yml down

test-env-logs: ## 查看测试环境日志
	docker-compose -f docker-compose.test.yml logs -f

# 测试数据管理
test-data-generate: ## 生成测试数据
	@echo "$(GREEN)生成测试数据...$(NC)"
	go run scripts/generate-test-data.go

test-data-clean: ## 清理测试数据
	@echo "$(YELLOW)清理测试数据...$(NC)"
	go run scripts/clean-test-data.go

# Mock生成
mock-generate: ## 生成Mock文件
	@echo "$(GREEN)生成Mock文件...$(NC)"
	@if command -v mockgen >/dev/null 2>&1; then \
		go generate ./...; \
	else \
		echo "$(YELLOW)mockgen 未安装$(NC)"; \
		echo "安装命令: go install github.com/golang/mock/mockgen@latest"; \
	fi

# 性能分析
profile-cpu: ## CPU性能分析
	@echo "$(GREEN)运行CPU性能分析...$(NC)"
	go test -cpuprofile=cpu.prof -bench=. ./...
	go tool pprof cpu.prof

profile-mem: ## 内存性能分析
	@echo "$(GREEN)运行内存性能分析...$(NC)"
	go test -memprofile=mem.prof -bench=. ./...
	go tool pprof mem.prof

# 添加到主要目标
.PHONY: test test-unit test-integration test-coverage test-bench test-race test-short test-clean test-watch quality security ci-test ci-build test-env-up test-env-down test-env-logs test-data-generate test-data-clean mock-generate profile-cpu profile-mem