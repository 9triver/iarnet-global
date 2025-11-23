# 多阶段构建 Dockerfile for iarnet-global
# 在同一个容器中启动前后端

# ============================================================================
# 阶段 1: 构建 Go 后端
# ============================================================================
FROM golang:1.25-alpine AS backend-builder

ENV GOPROXY=https://goproxy.cn,direct
ENV GOSUMDB=sum.golang.org

# 安装必要的工具和 SQLite 开发库（go-sqlite3 需要 CGO）
RUN apk add --no-cache git ca-certificates tzdata \
    gcc musl-dev \
    sqlite-dev

# 设置工作目录
WORKDIR /build

# 复制整个项目
COPY . /build/iarnet-global

# 切换到项目根目录
WORKDIR /build/iarnet-global

# 下载依赖
RUN go mod download

# 构建后端应用（启用 CGO，因为 go-sqlite3 需要）
RUN CGO_ENABLED=1 GOOS=linux GOARCH=amd64 go build \
    -ldflags='-w -s' \
    -a \
    -o iarnet-global ./cmd/main.go

# ============================================================================
# 阶段 2: 构建 Next.js 前端
# ============================================================================
FROM node:20-alpine AS frontend-builder

# 设置工作目录
WORKDIR /build

# 复制前端项目
COPY web /build/web

# 切换到前端目录
WORKDIR /build/web

# 确保 public 目录存在（Next.js 需要，即使为空）
RUN mkdir -p public || true

# 安装依赖
RUN npm install --legacy-peer-deps

# 构建前端（生产模式）
RUN npm run build

# ============================================================================
# 阶段 3: 运行阶段
# ============================================================================
FROM alpine:latest

# 安装必要的运行时依赖（包括 SQLite 运行时库）
RUN apk add --no-cache \
    ca-certificates \
    tzdata \
    nodejs \
    npm \
    netcat-openbsd \
    sqlite-libs

# 创建非 root 用户
RUN addgroup -S appuser && adduser -S -G appuser appuser

# 设置工作目录
WORKDIR /app

# 从构建阶段复制后端二进制文件
COPY --from=backend-builder /build/iarnet-global/iarnet-global /app/iarnet-global

# 从构建阶段复制前端构建产物
COPY --from=frontend-builder /build/web/.next /app/web/.next
# 复制 public 目录（在构建阶段已确保存在）
COPY --from=frontend-builder /build/web/public /app/web/public
COPY --from=frontend-builder /build/web/package.json /app/web/package.json
COPY --from=frontend-builder /build/web/next.config.mjs /app/web/next.config.mjs
COPY --from=frontend-builder /build/web/node_modules /app/web/node_modules

# 复制配置文件
COPY --from=backend-builder /build/iarnet-global/config.yaml /app/config.yaml

# 创建数据目录
RUN mkdir -p /app/data && \
    chown -R appuser:appuser /app

# 创建启动脚本
RUN cat > /app/entrypoint.sh << 'EOF'
#!/bin/sh
set -e

# 设置环境变量
export BACKEND_URL="${BACKEND_URL:-http://localhost:8080}"

# 启动后端服务（后台运行）
echo "启动 iarnet-global 后端服务..."
cd /app
/app/iarnet-global -config /app/config.yaml &
BACKEND_PID=$!

# 等待后端启动
echo "等待后端服务启动..."
for i in $(seq 1 30); do
    if nc -z localhost 8080 2>/dev/null; then
        echo "后端服务已启动"
        break
    fi
    if [ $i -eq 30 ]; then
        echo "警告: 后端服务启动超时，继续启动前端..."
    fi
    sleep 1
done

# 启动前端服务（前台运行，保持容器运行）
echo "启动 iarnet-global 前端服务..."
cd /app/web
exec npm start
EOF

RUN chmod +x /app/entrypoint.sh

# 暴露端口
# 8080: 后端 HTTP API
# 50010: RPC 服务端口
# 3000: 前端 Next.js
EXPOSE 8080 50010 3000

# 健康检查：检查 RPC 端口 50010 是否可用
HEALTHCHECK --interval=10s --timeout=3s --start-period=15s --retries=3 \
    CMD nc -z localhost 50010 && pgrep -f iarnet-global || exit 1

# 使用启动脚本作为入口点
ENTRYPOINT ["/app/entrypoint.sh"]

