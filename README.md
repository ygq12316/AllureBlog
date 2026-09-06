# 笔墨 · Ink & Code

水墨画风的个人博客（allureblog.top）：Go 后端 + Vue 3 前端 + AI 访客伴侣「笔墨精灵」，由 Docker Compose 编排三个服务。

## 功能特性

- **博客**：文章（Markdown 渲染 + TOC + 时间轴）、随笔（九宫格图片 + 软木板便签）、分类（3D 标签云）、搜索（canvas 词云）、弹幕墙
- **账号体系**：统一登录/注册，角色分 `user` / `admin`；评论、弹幕、精灵对话均需登录，后台仅管理员可进（前端守卫 + 接口角色校验双重门）
- **评论**：嵌套回复（两层，自动归根）、WebSocket 实时推送、乐观上屏（发送即显示、失败自动回退）、断线重连自动对账
- **弹幕实时**：全局 WS 房间广播，发一条全网即时滚动
- **笔墨精灵**：登录用户专属的 AI 对话悬浮球，流式回复 + 博客内检索工具，带每日限额与短期记忆
- **管理后台**：文章/随笔编辑、评论两级管理与级联删除、用户管理（角色标识）、弹幕管理、博客设置（作者头像圆形裁剪）

## 技术栈

| 层 | 技术 |
|---|---|
| 后端 | Go · Gin · GORM · SQLite（glebarez 纯 Go 驱动，WAL 模式）· gorilla/websocket |
| 前端 | Vue 3 · Vite · Tailwind CSS v4 · vue-router（无重型 UI 库，交互组件手写于 `web/src/components/ui/`） |
| 智能体 | Python · FastAPI · openai SDK（直连 DeepSeek 兼容端点，手写 agent 循环） |
| 部署 | Docker Compose · Caddy（自动 HTTPS）· GitHub Actions |

## 架构

```
            ┌────────── Caddy (:80/:443, 自动 HTTPS) ──────────┐
            │   /chat/* ──▶ agent :8000                        │
浏览器 ────▶│   其余 ────▶ app :8080                           │
            └──────────────────────────────────────────────────┘
                               │
        ┌──────────────────────┴───────────────────────┐
        ▼                                              ▼
  app（Go，:8080）                              agent（Python，:8000）
  · REST API + WebSocket 房间                   · /chat/ws 精灵对话
  · 托管 web/dist 构建产物（SPA 回退）           · DeepSeek 流式生成
  · /uploads 静态资源（上传图片）                · 博客账号登录校验（经 BLOG_API_BASE）
  · SQLite（blog.db，单文件）                    · 进程内短期记忆 + 每日限额
```

## 目录结构

```
├── server/              # Go 后端
│   ├── cmd/server/      # 入口：装配依赖、路由、中间件
│   └── internal/
│       ├── handler/     # HTTP/WS 处理器（鉴权、评论、弹幕、后台）
│       ├── service/     # 业务规则（校验、归根、级联等都在这层）
│       ├── repository/  # GORM 数据访问
│       ├── model/       # 数据模型
│       ├── database/    # 初始化 + AutoMigrate
│       └── util/        # Markdown 渲染、slug
├── web/                 # Vue 3 前端（水墨画风，Tailwind v4 token 体系）
│   ├── src/components/  # 页面组件 + ui/（手写交互基件）
│   ├── src/composables/ # useVisitor（账号态）/ useWS（断线重连）/ useFairyChat
│   └── src/api/         # axios 领域模块（统一注入 token、401 处理）
├── agent/               # 笔墨精灵（FastAPI）
├── scripts/seed.js      # 向本地后端灌测试文章
├── deploy.sh            # 服务器端部署脚本（拉代码/镜像 → up -d → 自检）
└── docker-compose.yml   # 三服务编排（app / agent / caddy）
```

## 快速开始（本地开发）

依赖：Go ≥ 1.25、Node ≥ 21、Python + `uv`（可选，跑智能体才需要）。

```bash
# 1. 后端（:8080，首次启动 AutoMigrate 建表；缺环境变量会拒绝启动）
cd server
JWT_SECRET=dev-secret ADMIN_PASSWORD=dev-pass go run ./cmd/server

# 2. 前端热更新（:5173，/api 与 /uploads 代理到 :8080，/chat 代理到 :8000）
cd web
npm install
npm run dev

# 3. 智能体（可选，:8000；先复制 agent/.env.example 为 agent/.env 填入 key）
cd agent
uv run uvicorn main:app --host 0.0.0.0 --port 8000
```

启动后访问 http://localhost:5173。管理员账号 = 环境变量 `ADMIN_USER` / `ADMIN_PASSWORD`（默认 `admin`），在统一登录框登录即可进后台。

### Docker 整体跑（本地演练）

```bash
cp .env.example .env        # 填入 JWT_SECRET / ADMIN_PASSWORD / DEEPSEEK_API_KEY
cp agent/.env.example agent/.env
docker compose -f docker-compose.yml -f docker-compose.local.yml up -d --build
```

本地 override 会构建当前代码而非拉远端镜像，并用 HTTP 版 Caddyfile 避开线上域名的证书签发。

## 生产部署

1. push 到 `master` 触发 GitHub Actions：并行构建 `allure12316/blog`（多阶段：Vue 构建 → Go 编译 → alpine 运行时）与 `allure12316/blog-agent` 两个镜像并推送
2. 服务器上执行 `./deploy.sh`：拉代码、拉镜像、`docker compose up -d`，随后自检（SPA 200 + `/chat/ws` 代理正常 + 容器状态）
3. 需要手工放置的文件（均在 .gitignore 中，不入库）：服务器 `.env` 与 `agent/.env`（参照两个 `.env.example`）

> 部署目录约定为 `/opt/blog`，可用环境变量 `DEPLOY_DIR` 覆盖。服务器若访问 Docker Hub 受限，可在本地 `docker save | ssh "docker load"` 直传镜像后 `docker compose up -d`。

## 环境变量

**app（`.env`，随 compose 注入）**

| 变量 | 必填 | 说明 |
|---|---|---|
| `JWT_SECRET` | ✓ | JWT 签名密钥，务必长随机串 |
| `ADMIN_USER` | | 管理员账号名（默认 `admin`） |
| `ADMIN_PASSWORD` | ✓ | 管理员密码；启动时作为种子同步进账号表 |
| `DB_PATH` | | SQLite 路径（默认仓库根 `blog.db`） |
| `AGENT_URL` | | 精灵服务地址（默认 `http://localhost:8000`） |

**agent（`agent/.env`）**

| 变量 | 必填 | 说明 |
|---|---|---|
| `DEEPSEEK_API_KEY` | ✓ | DeepSeek API Key |
| `DEEPSEEK_BASE_URL` / `DEEPSEEK_MODEL` | | 端点与模型（有默认值） |
| `BLOG_API_BASE` | | 校验博客登录用（容器内为 `http://app:8080`） |

> 密码/密钥只存在于两台机器的 `.env` 里，永远不进 git；`ADMIN_PASSWORD` 是「启动种子」——改密码 = 改 `.env` 后重启，账号表自动同步。

## 账号与角色

- 登录/注册统一入口（导航头像弹窗或 `/login`），成功即签发 7 天期带角色 JWT
- `user`：评论（含嵌套回复）、弹幕、精灵对话、资料编辑
- `admin`：以上全部 + 后台（服务端中间件校验角色，无角色令牌一律 403）
- 数据库没有该管理员账号时，启动种子会按 `ADMIN_USER` / `ADMIN_PASSWORD` 自动创建

## 常用命令

```bash
make run      # 起后端（等价：cd server && go run ./cmd/server）
make build    # 编译 server/blog.exe
make test     # go test ./...
make vet      # go vet ./...
make lint     # golangci-lint run（配置 server/.golangci.yml）
npm run build # 前端构建 → web/dist（生产由 Go 托管该目录）
node scripts/seed.js   # 向 localhost:8080 灌测试文章
```

无 `make` 的环境直接用注释里的等价命令。

## 更多

- 前端视觉遵循「水墨画风」token 体系（宣纸/墨色双主题），入口在 `web/src/assets/main.css`
- WebSocket 房间约定：`/api/notes/:id/ws` 为随笔房间（评论），`/api/ws` 为全局房间（弹幕）
