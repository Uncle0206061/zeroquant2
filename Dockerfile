# ============ Stage 1: Build ============
FROM golang:1.26-alpine AS builder

# 安装编译依赖
RUN apk add --no-cache git ca-certificates tzdata

WORKDIR /build

# 先复制依赖文件，利用 Docker 缓存层
COPY go.mod go.sum ./
RUN go mod download

# 复制源码
COPY . .

# 编译（静态链接，无 CGO）
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 \
    go build -ldflags="-s -w" -o /app/server ./cmd/server/

# ============ Stage 2: Runtime ============
FROM alpine:3.19

# 安装运行时依赖
RUN apk add --no-cache ca-certificates tzdata bash \
    && cp /usr/share/zoneinfo/Asia/Shanghai /etc/localtime \
    && echo "Asia/Shanghai" > /etc/timezone

WORKDIR /app

# 从 builder 阶段复制二进制
COPY --from=builder /app/server .
COPY config.example.yaml ./config.yaml

# 创建日志目录
RUN mkdir -p /app/logs

# 暴露端口
EXPOSE 8080

# 健康检查
HEALTHCHECK --interval=30s --timeout=5s --start-period=10s --retries=3 \
    CMD wget -qO- http://localhost:8080/api/v1/health || exit 1

# 非root运行
RUN addgroup -S appgroup && adduser -S appuser -G appgroup
RUN chown -R appuser:appgroup /app
USER appuser

ENTRYPOINT ["./server"]
