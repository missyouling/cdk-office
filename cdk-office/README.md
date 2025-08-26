# CDK-Office 企业内容管理平台

[![Go Version](https://img.shields.io/badge/Go-1.24-blue.svg)](https://golang.org/)
[![Next.js](https://img.shields.io/badge/Next.js-15-black.svg)](https://nextjs.org/)
[![React](https://img.shields.io/badge/React-19-blue.svg)](https://reactjs.org/)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)

CDK-Office是一个集成了Dify AI平台的企业内容管理平台，实现智能文档管理、AI问答和知识库管理功能。

## 🌟 主要特性

- 🔐 **Casbin权限控制** - 基于Casbin实现强大的访问控制功能
- 📄 **gopdf文档打印** - 用于文档打印和PDF生成功能
- ⏰ **gocron日程规划** - 用于日程规划和定时任务管理
- 🔄 **go-workflows审批流程** - 用于构建审批工作流引擎
- 📚 **ODD数据字典** - 用于数据字典管理，统一管理各功能模块的字段定义
- 🤖 **Dify AI集成** - 智能问答、文档处理和知识管理能力
- 📱 **二维码应用系统** - 支持动态表单、员工签到、在线订餐、问卷调查和访客登记等应用场景

## 🏗️ 技术架构

```
┌─────────────────┐    ┌─────────────────┐    ┌─────────────────┐
│   Frontend      │    │    Backend      │    │   Dify Platform │
│   (Next.js)     │◄──►│     (Go)        │◄──►│  (AI Services)  │
│                 │    │                 │    │                 │
│ • React 19      │    │ • Gin Framework │    │ • Knowledge Base│
│ • TypeScript    │    │ • Casbin        │    │ • Workflows     │
│ • Tailwind CSS  │    │ • gopdf         │    │ • App Engine    │
│ • Shadcn UI     │    │ • gocron        │    │                 │
└─────────────────┘    └─────────────────┘    └─────────────────┘
                              │
                              ▼
                    ┌─────────────────┐
                    │   Data Storage  │
                    │                 │
                    │ • PostgreSQL    │
                    │ • Redis Cache   │
                    │ • Object Store  │
                    └─────────────────┘
```

## 🚀 快速开始

### 环境要求

- Go 1.24+
- PostgreSQL 13+
- Redis 6+
- Node.js 18+ (前端)

### 安装依赖

```bash
# 克隆项目
git clone <repository-url>
cd cdk-office

# 安装Go依赖
go mod download

# 安装前端依赖
cd frontend
npm install
```

### 配置

1. 复制配置文件：
   ```bash
   cp config.example.yaml config.yaml
   ```

2. 编辑配置文件，设置数据库、Redis等连接信息

### 运行

```bash
# 运行API服务
make run-api

# 运行调度器
make run-scheduler

# 运行工作进程
make run-worker

# 运行前端
cd frontend
npm run dev
```

## 📁 项目结构

```
cdk-office/
├── internal/              # 核心代码
│   ├── apps/             # 应用模块
│   ├── cmd/              # 命令行接口
│   ├── config/           # 配置管理
│   ├── db/               # 数据库连接
│   ├── logger/           # 日志管理
│   ├── models/           # 数据模型
│   ├── otel_trace/       # 链路追踪
│   ├── router/           # 路由管理
│   ├── task/             # 任务管理
│   └── utils/            # 工具函数
├── frontend/             # 前端代码
├── scripts/              # 脚本文件
├── config.example.yaml   # 配置示例
├── config.vps.yaml       # VPS配置
├── Dockerfile            # Docker配置
├── Makefile              # 构建脚本
└── README.md             # 项目说明
```

## 🐳 Docker部署

```bash
# 构建镜像
docker build -t cdk-office .

# 运行容器
docker run -d \
  --name cdk-office \
  -p 8000:8000 \
  -v /path/to/config.yaml:/root/config.yaml \
  cdk-office
```

## 🛠️ 开发指南

### 代码格式化

```bash
make fmt
```

### 代码检查

```bash
make vet
```

### 运行测试

```bash
make test
```

## 📄 许可证

本项目采用MIT许可证，详情请见[LICENSE](LICENSE)文件。

## 🤝 贡献

欢迎提交Issue和Pull Request来改进项目！

## 📞 联系我们

如有问题，请联系项目维护者。