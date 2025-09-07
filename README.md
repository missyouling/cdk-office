# CDK-Office

CDK-Office 是基于开源项目 [linux-do/cdk](https://github.com/linux-do/cdk) 进行二次开发的企业内容管理平台，旨在与 Dify AI 平台深度集成，实现智能文档管理、AI 问答和知识库管理功能。

## 项目目标

- 基于 CDK-Office 权限体系实现 Dify 知识库的分级访问控制
- 将 CDK-Office 文档自动同步到 Dify 知识库，实现向量化存储和语义检索
- 基于 Dify 平台实现智能问答功能
- 深度集成 Dify 自动化编排功能，构建智能文档处理工作流
- 利用 Dify AI Agent 能力，提供智能助手服务
- 集成 Dify RAG Pipeline，实现多级检索和智能问答
- 优化 2C4G VPS 部署，提供多级缓存和搜索优化

## 技术架构

### 架构模式
采用模块化单体（Modular Monolith）架构
- 单一入口，统一部署
- 模块化设计，高内聚低耦合
- 为未来微服务化预留扩展空间
- 适配2C4G VPS性能优化

### 技术栈
| 分类 | 技术 | 版本 | 说明 |
|------|------|------|------|
| 后端 | Go | 1.24 | 主要开发语言 |
| Web框架 | Gin | 最新稳定版 | HTTP服务框架 |
| ORM | GORM | 最新稳定版 | 数据库访问 |
| 缓存 | Redis | 6.0+ | 缓存和会话存储 |
| 数据库 | Supabase (本地部署) | 最新版 | 主数据库，可选配置云服务 |
| 工作流引擎 | Temporal + Asynq + gocron | 最新版 | 分层工作流架构 |
| **AI平台** | **Dify** | **最新版** | **AI编排、RAG、Agent服务** |
| **智能工作流** | **Dify Workflow** | **最新版** | **可视化AI工作流设计** |
| **知识库** | **Dify RAG Pipeline** | **最新版** | **多级检索和智能问答** |
| 前端 | Next.js | 15 | React框架 |
| UI框架 | React | 19 | 用户界面库 |
| 类型系统 | TypeScript | 最新稳定版 | 类型安全 |
| 样式 | Tailwind CSS | 4 | 样式框架 |
| 组件库 | Shadcn UI + 业务组件库 | 最新版 | 基础组件+封装业务组件 |
| 表格组件 | TanStack Table | v8 | 高性能数据表格 |
| 二维码 | qrcode.js + canvas | 最新版 | 高质量二维码生成 |
| PDF处理 | jsPDF + PDF-lib | 最新版 | 前后端协同PDF处理 |
| 图像处理 | OpenCV + ImageMagick | 最新版 | 文档扫描和图像优化 |
| 图标 | Lucide Icons | 最新版 | 图标库 |

## 项目结构

```
cdk-office/
├── cmd/
│   └── server/                    // 单一入口
│       └── main.go
├── internal/                      // 内部模块
│   ├── auth/                     // 认证模块
│   │   ├── domain/               // 领域模型
│   │   ├── service/              // 业务逻辑
│   │   └── handler/              // HTTP处理器
│   ├── document/                 // 文档模块
│   │   ├── domain/
│   │   ├── service/
│   │   └── handler/
│   ├── workflow/                 // 工作流模块
│   │   ├── domain/
│   │   ├── service/
│   │   └── handler/
│   ├── employee/                 // 员工管理模块
│   │   ├── domain/
│   │   ├── service/
│   │   └── handler/
│   ├── business/                 // 业务中心模块
│   │   ├── domain/
│   │   ├── service/
│   │   └── handler/
│   ├── dify/                     // Dify集成模块
│   │   ├── workflow/             // 工作流集成
│   │   ├── rag/                  // RAG检索集成
│   │   ├── agent/                // AI Agent集成
│   │   └── client/               // Dify API客户端
│   └── shared/                   // 共享组件
│       ├── database/             // 数据库连接
│       ├── cache/                // 缓存管理
│       ├── middleware/           // 中间件
│       └── utils/                // 工具函数
├── pkg/                          // 公共包
│   ├── config/                   // 配置管理
│   ├── logger/                   // 日志管理
│   └── errors/                   // 错误处理
└── frontend/                     // 前端项目
    ├── components/               // 组件库
    │   ├── ui/                   // 基础UI组件 (Shadcn)
    │   ├── business/             // 业务组件
    │   └── ai/                   // AI相关组件
    ├── pages/                    // 页面
    └── lib/                      // 工具库
        └── dify/                 // Dify前端集成
```

## 开发环境搭建

### 环境要求
- Go >= 1.24
- Node.js >= 18.0
- Redis >= 6.0
- pnpm >= 8.0 (推荐)
- Docker >= 20.0
- Docker Compose >= 1.29

### 本地开发环境搭建步骤

1. 克隆项目代码：
   ```bash
   git clone <repository-url>
   cd cdk-office
   ```

2. 安装 Go 依赖：
   ```bash
   go mod tidy
   ```

3. 安装前端依赖：
   ```bash
   cd frontend
   pnpm install
   ```

4. 启动开发环境：
   ```bash
   docker-compose up -d
   ```

5. 启动后端服务：
   ```bash
   go run cmd/server/main.go
   ```

6. 启动前端服务：
   ```bash
   cd frontend
   pnpm dev
   ```

## 测试

### 运行测试

```bash
# 运行所有测试
make test

# 运行测试并生成覆盖率报告
make test-coverage

# 运行测试并查看覆盖率（在浏览器中）
make test-coverage-html
```

### 测试配置

测试配置文件位于 `test.env`，可以根据需要修改测试环境变量。

## 部署

### Docker部署（推荐）

1. 构建并启动服务：
   ```bash
   docker-compose up -d
   ```

2. 查看服务状态：
   ```bash
   docker-compose ps
   ```

3. 查看服务日志：
   ```bash
   docker-compose logs -f
   ```

### 脚本部署

项目提供了部署脚本 `deploy.sh` 来简化部署过程：

```bash
# 构建应用
./deploy.sh build

# 启动应用
./deploy.sh start

# 停止应用
./deploy.sh stop

# 重启应用
./deploy.sh restart

# 查看应用状态
./deploy.sh status

# 查看应用日志
./deploy.sh logs

# 创建数据库备份
./deploy.sh backup

# 从备份恢复数据库
./deploy.sh restore <backup_file>

# 清理旧备份
./deploy.sh cleanup
```

### 配置文件

生产环境配置文件位于 `config/production.yaml`，可以根据实际环境进行修改。

### 环境变量

部署时需要设置以下环境变量：
- `DIFY_API_KEY`: Dify API密钥
- `DIFY_BASE_URL`: Dify API基础URL
- `JWT_SECRET`: JWT密钥

## 项目开发优先级

根据项目需求的依赖关系、核心功能重要性和实施难度，制定以下开发优先级：

### 第一优先级（P0）- 核心基础架构和认证模块
这些功能是整个系统的基础，其他所有模块都依赖于它们：
1. 环境搭建和项目初始化
2. 认证模块开发

### 第二优先级（P1）- 核心业务模块
这些功能构成了系统的核心业务流程：
1. 文档管理模块开发
2. Dify集成模块开发
3. 员工管理模块开发

### 第三优先级（P2）- 扩展业务模块
这些功能扩展了系统的能力，提供了更多业务场景支持：
1. 业务中心模块开发
2. 应用中心模块开发

### 第四优先级（P3）- 移动端开发
移动端功能提供了更便捷的用户体验：
1. 微信小程序开发
2. 移动端功能集成

### 第五优先级（P4）- 系统优化
这些功能提升了系统的性能和安全性：
1. 性能优化
2. 安全优化
3. 用户体验优化

### 第六优先级（P5）- 测试和部署
这些功能确保系统的质量和稳定性：
1. 测试阶段
2. 部署上线阶段
3. 项目收尾阶段

## 项目状态

✅ 所有功能开发已完成
✅ 所有组件占位符已替换为实际实现
✅ 测试框架已建立
✅ 部署方案已制定

## 文档

- [标准需求文档](../docs/standard-requirements-document.md)
- [项目任务清单](../docs/project-task-list.md)
- [项目优先级计划](../docs/project-priority-plan.md)
- [系统设计文档](../docs/system-design-document.md)

## 许可证

[MIT License](LICENSE)