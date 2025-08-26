# LINUX DO CDK

🚀 Linux Do 社区 CDK (Content Distribution Kit) 快速分享平台

[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)
[![Go Version](https://img.shields.io/badge/Go-1.24-blue.svg)](https://golang.org/)
[![Next.js](https://img.shields.io/badge/Next.js-15-black.svg)](https://nextjs.org/)
[![React](https://img.shields.io/badge/React-19-blue.svg)](https://reactjs.org/)

[![GitHub release](https://img.shields.io/github/v/release/linux-do/cdk?include_prereleases)](https://github.com/linux-do/cdk/releases)
[![GitHub stars](https://img.shields.io/github/stars/linux-do/cdk)](https://github.com/linux-do/cdk/stargazers)
[![GitHub forks](https://img.shields.io/github/forks/linux-do/cdk)](https://github.com/linux-do/cdk/network)
[![GitHub issues](https://img.shields.io/github/issues/linux-do/cdk)](https://github.com/linux-do/cdk/issues)
[![GitHub pull requests](https://img.shields.io/github/issues-pr/linux-do/cdk)](https://github.com/linux-do/cdk/pulls)
[![GitHub contributors](https://img.shields.io/github/contributors/linux-do/cdk)](https://github.com/linux-do/cdk/graphs/contributors)

[![Backend Build](https://github.com/linux-do/cdk/actions/workflows/build_backend.yml/badge.svg)](https://github.com/linux-do/cdk/actions/workflows/build_backend.yml)
[![Frontend Build](https://github.com/linux-do/cdk/actions/workflows/build_frontend.yml/badge.svg)](https://github.com/linux-do/cdk/actions/workflows/build_frontend.yml)
[![Docker Build](https://github.com/linux-do/cdk/actions/workflows/build_image.yml/badge.svg)](https://github.com/linux-do/cdk/actions/workflows/build_image.yml)
[![CodeQL](https://github.com/linux-do/cdk/actions/workflows/codeql.yml/badge.svg)](https://github.com/linux-do/cdk/actions/workflows/codeql.yml)
[![ESLint](https://github.com/linux-do/cdk/actions/workflows/eslint.yml/badge.svg)](https://github.com/linux-do/cdk/actions/workflows/eslint.yml)

## 📖 项目简介

LINUX DO CDK 是一个为 Linux Do 社区打造的内容分发工具包，旨在提供快速、安全、便捷的 CDK 分享服务。平台支持多种分发方式，具备完善的用户权限管理和风险控制机制。

### ✨ 主要特性

- 🔐 **OAuth2 认证** - 集成 Linux Do 社区账号系统
- 🎯 **多种分发模式** - 支持不同的 CDK 分发策略
- 🛡️ **风险控制** - 完善的信任等级和风险评估系统
- 📊 **实时监控** - 详细的分发统计和用户行为分析
- 🎨 **现代化界面** - 基于 Next.js 15 和 React 19 的响应式设计
- ⚡ **高性能** - Go 后端 + Redis 缓存 + MySQL 数据库

## 🏗️ 架构概览

```
┌─────────────────┐    ┌─────────────────┐    ┌─────────────────┐
│   Frontend      │    │    Backend      │    │   Database      │
│   (Next.js)     │◄──►│     (Go)        │◄──►│  (MySQL/Redis)  │
│                 │    │                 │    │                 │
│ • React 19      │    │ • Gin Framework │    │ • MySQL         │
│ • TypeScript    │    │ • OAuth2        │    │ • Redis Cache   │
│ • Tailwind CSS  │    │ • OpenTelemetry │    │ • Session Store │
│ • Shadcn UI     │    │ • Swagger API   │    │                 │
└─────────────────┘    └─────────────────┘    └─────────────────┘
```

## 🛠️ 技术栈

### 后端
- **Go 1.24** - 主要开发语言
- **Gin** - Web 框架
- **GORM** - ORM 框架
- **Redis** - 缓存和会话存储
- **MySQL** - 主数据库
- **OpenTelemetry** - 可观测性
- **Swagger** - API 文档

### 前端
- **Next.js 15** - React 框架
- **React 19** - UI 库
- **TypeScript** - 类型安全
- **Tailwind CSS 4** - 样式框架
- **Shadcn UI** - 组件库
- **Lucide Icons** - 图标库

## 📋 环境要求

- **Go** >= 1.24
- **Node.js** >= 18.0
- **MySQL** >= 8.0
- **Redis** >= 6.0
- **pnpm** >= 8.0 (推荐)

## 🚀 快速开始

```
# 克隆项目
git clone https://github.com/linux-do/cdk.git
cd cdk

# 复制配置文件
cp config.example.yaml config.yaml

# 编辑配置文件
vim config.yaml

# 运行项目
go run main.go api
```

## VPS部署优化

对于2C4G配置的VPS环境，建议使用优化的配置文件以减少内存占用：

```
# 使用VPS优化配置
cp config.vps.yaml config.yaml

# 根据实际环境修改配置文件中的敏感信息
vim config.yaml

# 构建项目
go build -o cdk-server main.go

# 运行API服务
./cdk-server api &

# 运行Worker服务
./cdk-server worker &
```

优化配置主要调整了以下参数：
1. 数据库连接池：减少最大连接数和空闲连接数
2. Redis连接池：减少连接池大小和空闲连接数
3. 工作进程并发：降低并发数以减少内存使用
4. 日志级别：调整为警告级别以减少日志内存占用
5. 连接超时：优化连接生命周期

## ⚙️ 配置说明

### 主要配置项

| 配置项 | 说明 | 示例 |
|--------|------|------|
| `app.addr` | 后端服务监听地址 | `:8000` |
| `oauth2.client_id` | OAuth2 客户端 ID | `your_client_id` |
| `database.host` | MySQL 数据库地址 | `127.0.0.1` |
| `redis.host` | Redis 服务器地址 | `127.0.0.1` |

详细配置说明请参考 `config.example.yaml` 文件。

## 🔧 开发指南

### 后端开发

```
# 运行 API 服务器
go run main.go api

# 运行任务调度器
go run main.go scheduler

# 运行工作队列
go run main.go worker

# 生成 Swagger 文档
make swagger

# 代码格式化和检查
make tidy
```

### 前端开发

```
cd frontend

# 开发模式（使用 Turbopack）
pnpm dev

# 构建生产版本
pnpm build

# 启动生产服务
pnpm start

# 代码检查和格式化
pnpm lint
pnpm format
```

## 📚 API 文档

API 文档通过 Swagger 自动生成，启动后端服务后可访问：

```
http://localhost:8000/swagger/index.html
```

### 主要 API 端点

- `GET /api/health` - 健康检查
- `GET /api/oauth2/login` - OAuth2 登录
- `GET /api/projects` - 获取项目列表
- `POST /api/projects` - 创建新项目

## 🧪 测试

```
# 后端测试
go test ./...

# 前端测试
cd frontend
pnpm test
```

## 🚀 部署

### Docker 部署

```
# 构建镜像
docker build -t linux-do-cdk .

# 运行容器
docker run -d -p 8000:8000 linux-do-cdk
```

### 生产环境部署

1. 构建前端资源：
   ```bash
   cd frontend && pnpm build
   ```

2. 编译后端程序：
   ```bash
   go build -o cdk main.go
   ```

3. 配置生产环境的 `config.yaml`

4. 启动服务：
   ```bash
   ./cdk api
   ```

## 🤝 贡献指南

我们欢迎社区贡献！请在提交代码前阅读：

- [贡献指南](CONTRIBUTING.md)
- [行为准则](CODE_OF_CONDUCT.md)
- [贡献者许可协议](CLA.md)

### 提交流程

1. Fork 本仓库
2. 创建特性分支 (`git checkout -b feature/your-feature`)
3. 提交更改 (`git commit -am 'Add your feature'`)
4. 推送到分支 (`git push origin feature/your-feature`)
5. 创建 Pull Request

## 📄 许可证

本项目基于 [MIT 许可证](LICENSE) 开源。

## 🔗 相关链接

- [Linux Do 社区](https://linux.do)
- [问题反馈](https://github.com/linux-do/cdk/issues)
- [功能请求](https://github.com/linux-do/cdk/issues/new?template=feature_request.md)

## ❤️ 致谢

感谢所有为本项目做出贡献的开发者和 Linux Do 社区的支持！
