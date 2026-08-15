#!/bin/bash
set -e

# 部署目录可用环境变量覆盖
DEPLOY_DIR="${DEPLOY_DIR:-/opt/blog/blog}"

echo "=== Bimo Ink & Code Deploy ==="

cd "$DEPLOY_DIR"

echo "1. Pull latest code..."
git pull

echo "2. Pull latest image..."
docker pull allure12316/blog:latest

echo "3. Start services..."
docker compose up -d --remove-orphans

echo "4. Clean up old images..."
docker image prune -f

echo "=== Deploy Complete ==="
docker compose ps
