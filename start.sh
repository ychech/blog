#!/usr/bin/env bash
# 博客服务启动脚本
# 用法：./start.sh [up|down|logs|build|infra|full]

set -e

COMPOSE_FILE="docker-compose.yml"
COMPOSE_FULL_FILE="docker-compose.full.yml"
CMD="${1:-infra}"

case "$CMD" in
  infra|up)
    echo "启动博客基础设施（本地开发模式）..."
    docker compose -f "$COMPOSE_FILE" up -d
    echo "基础设施已启动，请在另一个终端运行：APP_ENV=dev go run main.go"
    ;;
  full)
    echo "启动博客全栈服务（含后端镜像构建）..."
    docker compose -f "$COMPOSE_FULL_FILE" up -d --build
    ;;
  down)
    echo "停止博客服务..."
    docker compose -f "$COMPOSE_FILE" down
    docker compose -f "$COMPOSE_FULL_FILE" down 2>/dev/null || true
    ;;
  logs)
    echo "查看日志..."
    docker compose -f "$COMPOSE_FILE" logs -f
    ;;
  build)
    echo "构建后端镜像（需要能拉取 golang/alpine 镜像）..."
    docker compose -f "$COMPOSE_FULL_FILE" build blog
    ;;
  *)
    echo "用法：./start.sh [infra|full|down|logs|build]"
    echo "  infra  - 仅启动基础设施（默认，本地开发推荐）"
    echo "  full   - 全栈启动，含 blog 后端镜像构建"
    echo "  down   - 停止并移除容器"
    echo "  logs   - 查看基础设施日志"
    echo "  build  - 仅构建 blog 后端镜像"
    exit 1
    ;;
esac
