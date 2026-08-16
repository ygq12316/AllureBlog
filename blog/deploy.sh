#!/bin/bash
set -e

# 部署目录可用环境变量覆盖
DEPLOY_DIR="${DEPLOY_DIR:-/opt/blog/blog}"

echo "=== Bimo Ink & Code Deploy ==="

cd "$DEPLOY_DIR"

echo "1. Pull latest code..."
git pull

echo "2. Pull latest images..."
docker pull allure12316/blog:latest
docker pull allure12316/blog-agent:latest

echo "3. Start services..."
docker compose up -d --remove-orphans

echo "4. Verify deployment..."
# 首页 SPA 应经 caddy → app 返回 200（含页面标题）
for i in $(seq 1 15); do
  if curl -sf http://localhost/ | grep -q '<title>'; then
    echo "  SPA OK"
    break
  fi
  [ "$i" = 15 ] && { echo "  SPA 验证失败"; exit 1; }
  sleep 2
done

# /chat/ws 无 Upgrade 头应返回非 HTML（404=uvicorn 正常响应；200+HTML 即 SPA 回退，代理缺失）
chat_body=$(curl -s http://localhost/chat/ws || true)
if echo "$chat_body" | grep -qi '<!DOCTYPE html\|<html'; then
  echo "  /chat 代理异常（返回了 SPA 页面）"; exit 1
fi
echo "  Chat proxy OK"

# 容器健康状态（app/caddy 无 HEALTHCHECK 看 running，agent 带 HEALTHCHECK 看 health）
docker compose ps
echo "=== Deploy Complete ==="
