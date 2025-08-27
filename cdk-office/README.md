# CDK-Office 企业级办公系统

[![Go Version](https://img.shields.io/badge/Go-1.24-blue.svg)](https://golang.org/)
[![Next.js](https://img.shields.io/badge/Next.js-15-black.svg)](https://nextjs.org/)
[![React](https://img.shields.io/badge/React-19-blue.svg)](https://reactjs.org/)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)
[![Docker](https://img.shields.io/badge/Docker-Ready-blue.svg)](https://docker.com/)

🚀 **现代化企业办公解决方案** - 集成数据隔离、智能审批、知识管理、PDF处理等核心功能

## 🌟 核心特性

### 🔒 数据隔离
- **多级隔离策略**: 支持无隔离、基础隔离、标准隔离、严格隔离、完全隔离
- **访问控制**: 基于角色的细粒度权限控制
- **审计日志**: 完整的数据访问审计追踪
- **数据脱敏**: 敏感信息自动脱敏保护

### 📋 智能审批
- **可视化设计器**: 拖拽式工作流程设计
- **多种节点类型**: 用户任务、服务任务、决策节点、并行网关
- **动态表单**: 灵活的表单设计和数据收集
- **智能路由**: 基于条件的智能流程路由

### 📚 知识管理
- **个人知识库**: 私有文档管理和组织
- **团队协作**: 文档分享和协同编辑
- **AI问答**: 基于Dify平台的智能问答
- **全文搜索**: 强大的内容搜索和推荐

### 📄 PDF工具
- **20+种操作**: 合并、拆分、转换、压缩、加密等
- **OCR识别**: 多语言文字识别
- **批量处理**: 高效的批量文档处理
- **安全保护**: 数字签名和权限控制

### ⚡ 系统优化
- **智能缓存**: 多级缓存架构
- **API限流**: 智能请求频率控制
- **服务熔断**: 自动故障恢复机制
- **性能监控**: 实时系统性能监控

## 🏗️ 技术架构

### 后端技术栈
- **Go 1.24+**: 高性能后端服务
- **Gin**: 轻量级Web框架
- **GORM**: ORM数据库操作
- **PostgreSQL**: 主数据库
- **Redis**: 缓存和会话存储
- **go-workflows**: 工作流引擎

### 前端技术栈
- **React 19**: 现代化前端框架
- **Next.js 15**: 全栈React框架
- **TypeScript**: 类型安全的JavaScript
- **Material-UI**: 组件库
- **SWR**: 数据获取和缓存

### 集成服务
- **Stirling PDF**: PDF处理服务
- **KKFileView**: 文件预览服务
- **Dify**: AI智能问答平台
- **Docker**: 容器化部署

## 🚀 快速开始

### 环境要求

- Docker 20.10.0+
- Docker Compose 1.29.0+
- 4GB+ RAM
- 20GB+ 存储空间

### 一键部署

```bash
# 克隆项目
git clone https://github.com/your-org/cdk-office.git
cd cdk-office

# 配置环境变量
cp .env.example .env
# 编辑 .env 文件，设置数据库密码等

# 启动所有服务
make deploy

# 等待服务启动完成
make status
```

### 验证部署

```bash
# 健康检查
curl http://localhost:8000/api/v1/health

# 访问前端界面
open http://localhost:8000
```

## 📖 文档

- [📘 API文档](docs/API.md) - 完整的API接口文档
- [👥 用户手册](docs/USER_MANUAL.md) - 详细的使用指南
- [🚀 部署文档](docs/DEPLOYMENT.md) - 部署和运维指南
- [🧪 测试文档](docs/TESTING.md) - 测试执行指南

## 🛠️ 开发指南

### 环境搭建

```bash
# 安装Go依赖
go mod download

# 安装前端依赖
cd frontend
npm install

# 启动开发环境
make dev
```

### 测试

```bash
# 运行所有测试
make test

# 运行单元测试
make test-unit

# 运行集成测试
make test-integration

# 生成覆盖率报告
make test-coverage
```

### 代码质量

```bash
# 代码格式化
make fmt

# 代码检查
make lint

# 安全扫描
make security
```

## 📊 系统监控

### 性能指标

访问 `http://localhost:8000/api/v1/optimization/performance/metrics` 查看：

- CPU 使用率
- 内存使用情况
- API 响应时间
- 数据库连接状态
- 缓存命中率

### 健康检查

```bash
# 基础健康检查
curl http://localhost:8000/api/v1/health

# 详细健康状态
curl http://localhost:8000/api/v1/health/detailed
```

## 🔧 配置选项

### 环境变量

| 变量名 | 说明 | 默认值 |
|--------|------|--------|
| `APP_ENV` | 应用环境 | `development` |
| `DB_HOST` | 数据库主机 | `localhost` |
| `REDIS_HOST` | Redis主机 | `localhost` |
| `DIFY_API_URL` | Dify API地址 | - |

### 功能开关

```yaml
# config.yaml
features:
  data_isolation: true
  workflow_engine: true
  knowledge_base: true
  pdf_tools: true
  file_preview: true
```

## 🐳 容器化部署

### Docker Compose

```bash
# 生产环境部署
docker-compose -f docker-compose.prod.yml up -d

# 开发环境部署
docker-compose -f docker-compose.dev.yml up -d
```

### Kubernetes

```bash
# 部署到Kubernetes
kubectl apply -f k8s-deployment.yaml

# 查看部署状态
kubectl get pods -n cdk-office
```

## 🔐 安全配置

### SSL/TLS

```bash
# 生成自签名证书（开发环境）
make ssl-cert

# 使用Let's Encrypt（生产环境）
certbot --nginx -d your-domain.com
```

### 访问控制

- 基于角色的权限控制(RBAC)
- 数据隔离策略
- API限流保护
- 请求审计日志

## 📈 性能优化

### 缓存策略

- L1缓存：内存缓存（最快）
- L2缓存：Redis缓存（快速）
- L3缓存：数据库缓存（持久）

### 数据库优化

- 连接池管理
- 查询性能监控
- 慢查询优化
- 索引优化建议

## 🔄 CI/CD 流程

### GitHub Actions

```yaml
# .github/workflows/ci.yml
name: CI
on: [push, pull_request]
jobs:
  test:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v2
      - name: Run tests
        run: make ci-test
```

### 自动部署

```bash
# 部署到测试环境
make deploy-test

# 部署到生产环境
make deploy-prod
```

## 🤝 贡献指南

我们欢迎各种形式的贡献！

### 提交代码

1. Fork 项目仓库
2. 创建功能分支 (`git checkout -b feature/amazing-feature`)
3. 提交更改 (`git commit -m 'Add amazing feature'`)
4. 推送到分支 (`git push origin feature/amazing-feature`)
5. 创建 Pull Request

### 报告问题

- 使用 [GitHub Issues](https://github.com/your-org/cdk-office/issues)
- 提供详细的问题描述和复现步骤
- 附上相关的日志和截图

### 功能请求

- 在 Issues 中描述新功能需求
- 说明使用场景和预期效果
- 参与功能设计讨论

## 📞 支持

### 社区支持

- 📧 邮件: support@your-domain.com
- 💬 讨论区: [GitHub Discussions](https://github.com/your-org/cdk-office/discussions)
- 📱 QQ群: 123456789
- 🎮 Discord: [CDK-Office Community](https://discord.gg/cdk-office)

### 商业支持

- 🏢 企业版咨询: enterprise@your-domain.com
- 📞 技术支持: +86-400-xxx-xxxx
- 🎯 定制开发服务
- 🚀 部署和运维服务

## 📄 许可证

本项目基于 [MIT License](LICENSE) 开源协议发布。

## 🙏 致谢

感谢以下开源项目和贡献者：

- [Gin](https://github.com/gin-gonic/gin) - Web框架
- [GORM](https://github.com/go-gorm/gorm) - ORM库
- [React](https://github.com/facebook/react) - 前端框架
- [Stirling PDF](https://github.com/Stirling-Tools/Stirling-PDF) - PDF处理
- [KKFileView](https://github.com/kekingcn/kkFileView) - 文件预览
- [Dify](https://github.com/langgenius/dify) - AI平台

## 🗓️ 更新日志

### v1.0.0 (2024-01-20)

#### 🎉 首个正式版本发布

**新功能：**
- ✨ 完整的数据隔离系统
- ✨ 智能审批工作流引擎
- ✨ 个人和团队知识库
- ✨ 强大的PDF处理工具
- ✨ 多格式文件预览
- ✨ 系统性能优化
- ✨ 多渠道通知推送

**技术特性：**
- 🔧 微服务架构设计
- 🔧 容器化部署支持
- 🔧 完整的API文档
- 🔧 全面的单元测试
- 🔧 性能监控和报警

---

// ... existing project structure and other content ...

---

<div align="center">

**⭐ 如果这个项目对你有帮助，请给它一个Star！⭐**

[🏠 官网](https://your-domain.com) |
[📖 文档](https://docs.your-domain.com) |
[🐛 报告问题](https://github.com/your-org/cdk-office/issues) |
[💡 功能建议](https://github.com/your-org/cdk-office/discussions)

</div>