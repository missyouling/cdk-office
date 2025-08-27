# CDK-Office 优化的多阶段 Dockerfile
# 支持生产环境部署和性能优化

# 基础阶段 - 依赖管理
FROM golang:1.21-alpine AS deps
WORKDIR /app

# 安装构建依赖
RUN apk add --no-cache git ca-certificates tzdata

# 复制依赖文件
COPY go.mod go.sum ./

# 下载依赖（利用Docker缓存层）
RUN go mod download
RUN go mod verify

# 构建阶段
FROM deps AS builder

# 复制源代码
COPY . .

# 设置构建参数
ARG VERSION=dev
ARG COMMIT_SHA=unknown
ARG BUILD_TIME

# 构建优化的二进制文件
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build \
    -a -installsuffix cgo \
    -ldflags="-s -w -X main.Version=${VERSION} -X main.CommitSHA=${COMMIT_SHA} -X main.BuildTime=${BUILD_TIME}" \
    -o cdk-office .

# 运行时阶段 - 最小化镜像
FROM alpine:3.19 AS runtime

# 创建非root用户
RUN addgroup -g 1001 -S appgroup && \
    adduser -u 1001 -S appuser -G appgroup

# 安装运行时依赖
RUN apk --no-cache add \
    ca-certificates \
    tzdata \
    curl \
    wget && \
    update-ca-certificates

# 设置时区
ENV TZ=Asia/Shanghai
RUN ln -snf /usr/share/zoneinfo/$TZ /etc/localtime && echo $TZ > /etc/timezone

# 创建必要目录
RUN mkdir -p /app/logs /app/uploads /app/pdf_results /app/cache && \
    chown -R appuser:appgroup /app

# 设置工作目录
WORKDIR /app

# 从构建阶段复制二进制文件
COPY --from=builder --chown=appuser:appgroup /app/cdk-office .

# 复制配置文件
COPY --chown=appuser:appgroup config.vps.yaml ./config.yaml

# 切换到非root用户
USER appuser

# 暴露端口
EXPOSE 8000

# 健康检查
HEALTHCHECK --interval=30s --timeout=10s --start-period=30s --retries=3 \
    CMD curl -f http://localhost:8000/api/v1/health || exit 1

# 设置环境变量
ENV GIN_MODE=release
ENV APP_ENV=production

# 启动命令
CMD ["./cdk-office", "api"]