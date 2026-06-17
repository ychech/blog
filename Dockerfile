# 多阶段构建：第一阶段编译 Go 二进制，第二阶段运行
FROM golang:1.23-alpine AS builder

WORKDIR /app

# 安装依赖（cgo 需要 gcc，sqlite 需要 musl-dev）
RUN apk add --no-cache gcc musl-dev

# 先复制 go.mod 和 go.sum，利用 Docker 缓存层
COPY go.mod go.sum ./
RUN go mod download

# 复制源码并编译
COPY . .
RUN CGO_ENABLED=1 GOOS=linux go build -ldflags="-s -w" -o blog-server main.go

# 第二阶段：最小运行镜像
FROM alpine:latest

WORKDIR /app

# 安装 ca-certificates（用于 HTTPS 请求，如 SMTP、Meilisearch）
RUN apk add --no-cache ca-certificates tzdata

# 从构建阶段复制二进制
COPY --from=builder /app/blog-server /app/blog-server

# 暴露端口
EXPOSE 8080

# 健康检查
HEALTHCHECK --interval=30s --timeout=5s --start-period=10s --retries=3 \
  CMD wget -qO- http://localhost:8080/health || exit 1

# 默认启动命令
CMD ["./blog-server"]
