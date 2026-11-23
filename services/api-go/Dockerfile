# Build stage
FROM golang:1.23-alpine AS builder

WORKDIR /build

# 安装基础编译依赖
RUN apk add --no-cache git ca-certificates tzdata

# 依赖缓存加速
COPY go.mod go.sum ./
RUN go mod download

# 复制源码
COPY . .

# 编译纯静态 Linux 二进制（关闭 CGO，剥离符号表以最小化体积）
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build \
    -ldflags="-s -w -X main.version=0.1.0" \
    -o cyclecompass-go ./cmd/server

# Production stage
FROM alpine:3.20

WORKDIR /app

# 安装根证书与时区配置
RUN apk add --no-cache ca-certificates tzdata curl && \
    cp /usr/share/zoneinfo/Asia/Shanghai /etc/localtime && \
    echo "Asia/Shanghai" > /etc/timezone

# 安全配置：创建非 root 运行账户
RUN addgroup -g 10001 -S appgroup && \
    adduser -u 10001 -S appuser -G appgroup

# 从构建层拷贝二进制
COPY --from=builder /build/cyclecompass-go /app/cyclecompass-go

# 创建数据挂载目录
RUN mkdir -p /app/data /app/references && chown -R appuser:appgroup /app

USER appuser:appgroup

ENV PORT=8002 \
    DATA_DIR=/app/references/strategy_original/index_rotation/data \
    DB_PATH=/app/data/cyclecompass.db

EXPOSE 8002

HEALTHCHECK --interval=15s --timeout=3s --start-period=5s --retries=3 \
  CMD curl -f http://127.0.0.1:8002/api/v1/health || exit 1

ENTRYPOINT ["/app/cyclecompass-go"]
