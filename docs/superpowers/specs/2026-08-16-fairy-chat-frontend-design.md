# 笔墨精灵前端聊天入口 · 设计文档

日期：2026-08-16
状态：已与用户逐节确认

## 背景与目标

Agent 服务（`blog/agent/`，笔墨精灵）服务端代码与部署链路完整（`/chat/ws` WebSocket 端点、Caddy 路由、compose 编排均就绪），但 Vue SPA 前端从未接入——无路由、无组件、无 WebSocket 客户端代码。本设计补齐前端部分：聊天 UI 与 `/chat/ws` 客户端。

**范围：仅前端**（`blog/web/admin/`），不改服务端任何代码。

## 已确认的关键决策

| 决策点 | 结论 |
|---|---|
| UI 形态 | 右下角悬浮精灵球 + 弹出聊天窗（不做独立页面） |
| 精灵形象 | 水墨图标球：金褐墨色径向渐变圆球 + 书法感毛笔 SVG 图标（`BrushOutline`），无人脸/具象角色 |
| 会话历史 | 仅活在当前页面生命周期，刷新即清空（不存 localStorage） |
| 代码组织 | composable（纯逻辑）+ 单组件（球与窗同文件），挂 `PublicLayout` |
| 重连 | 不自动重连，UI 提供手动"重新召唤" |

历史决策的依据：服务端 agent 未挂 checkpointer（`agent/core.py` 的 `thread_id` 实际不生效），Redis 记忆只写不读仅作审计——精灵本来就无跨请求记忆，且无历史拉取端点，保留旧消息反而误导用户。

## 架构

### 文件结构

```
src/composables/useFairyChat.js   # 纯逻辑层：WS 状态机
src/components/FairyChat.vue       # 悬浮球 + 弹出窗（视图层）
src/layouts/PublicLayout.vue       # 挂载点（admin 路由不挂）
```

挂 `PublicLayout` 的理由：布局级组件跨公开路由持久——切换公开页面连接不断开；进入 `/admin/*` 时 PublicLayout 卸载、组件随之销毁并自动断连，行为天然合理。

### `useFairyChat.js` 接口

```js
const { status, messages, streaming, connect, send, interrupt, disconnect } = useFairyChat()
// status: 'idle' | 'connecting' | 'ready' | 'offline'
// messages: [{ role: 'user' | 'fairy', content, done: bool, interrupted?: bool }]
```

### 连接生命周期（惰性 + 常驻）

1. **首次打开聊天窗**才建连（不预连）：`ws(s)://{host}/chat/ws`，协议自适应（对齐 `CommentSection.vue` 现有写法）
2. `onopen` → 发 `auth`，`visitor_uuid`/`visitor_name` 取自 `useVisitor`（`connect()` 内先 `await init()`，uuid 必定存在）
3. 收 `auth_result` → `status='ready'`，`greeting` 作为第一条 fairy 气泡
4. 连接建立后**常驻**，每 25s 发 `ping` 保活（低于常见代理 60s 空闲超时）
5. 组件卸载 → `disconnect()`：清心跳 interval、关连接

### 消息收发与流式拼接

- `send(content)` → 本地立即追加 user 气泡 + 占位 fairy 气泡（`streaming=true`），发 `{type:'message', content}`
- 收 `token` → 追加到**最后一条未完成的 fairy 气泡**，若无则新建（鲁棒处理服务端打断旧流后立即开新流的情况）
- 收 `done` → 定稿当前气泡；`interrupted=true` 时气泡尾部渲染斜体小字"笔墨言未尽…"
- `rejected`/`error` → 其 `display` 字段渲染为斜体系统气泡，连接保持
- **streaming 中允许直接发新消息**：服务端 `_handle_urgent` 把新 message 视作打断；前端直接发送，token 按上述规则落到新气泡
- `interrupt()` → 发 `{type:'interrupt'}`，本地仅等 `done` 收尾，不强行截断（以服务端 done 为准）

### 断线

`onclose`/`onerror` → `status='offline'`；streaming 中的气泡定稿并标注"（连接中断）"。不自动重连，UI 提供手动"重新召唤"按钮。

## UI 视觉与交互

### 悬浮球（收起态）

- `fixed` 右下角（`right: 24px; bottom: 24px`，移动端 16px），56px 圆
- 视觉：金褐墨色径向渐变球（`rgba(184,148,76)` 系）+ 中心白色 `BrushOutline` 图标
- 动效：待机呼吸（外圈墨晕 4s 循环缓慢扩散）、hover 放大 1.08 + 图标微旋
- `z-index: 900`（内容之上，低于 Naive UI 弹窗层级）

### 弹出聊天窗（展开态）

- 从球上方展开（`transform-origin: bottom right`，scale + 淡入 250ms）
- 尺寸：桌面 380×560px；移动端（<480px）贴底全宽抽屉（宽 `100vw - 32px`、高 70vh，底部滑入）
- 三段式布局：
  - **头部**（48px）：墨色底，毛笔图标 + "笔墨精灵"标题，在线状态点（ready=金点脉动 / offline=灰点）+ 收起按钮
  - **消息区**：fairy 气泡居左（`--tag-bg` 底、墨尖角标、14px LXGW WenKai）；user 气泡居右（金色描边浅底）；不显示头像（窗窄保持轻）；系统消息居中斜体灰色小字
  - **输入区**：自适应 textarea（1-3 行，maxlength 500，Ctrl+Enter 发送）+ 右侧按钮
- 按钮状态机：空闲=`发送`（金色实心）；streaming=`停笔`（描边样式，点击发 interrupt）；offline=`重新召唤`
- 流式中的 fairy 气泡尾部闪烁墨点（`…` 动画）表示"运笔中"

### 触达细节

- 新消息到达且窗收起 → 球上未读金点 + 轻微晃动一次
- 打开窗自动滚到底；流式期间持续跟随滚动，用户手动上滚则暂停跟随，出现"回到底部"小按钮

## 错误处理矩阵

| 场景 | 表现 | 恢复 |
|---|---|---|
| WS 建连失败（agent 未部署/挂了） | `status='offline'`，窗内占位"精灵云游去了…" | "重新召唤"重连 |
| auth 被拒（rejected） | 斜体系统气泡显示 `display` | 连接保持，可继续发消息 |
| 流式中 agent_error | "笔墨精灵思绪受阻"斜体气泡，当前气泡定稿 | 再发一条即可 |
| 流式中断线 | 当前气泡定稿 + "（连接中断）"标注 | 重新召唤；输入保留 |
| 发送时 offline | 输入保留不清空，顶部离线提示条 | 重连后手动重发 |

原则：**绝不静默丢弃用户输入**（对齐 `CommentSection.vue` 先例）。

## 边界情况

- 快速连发：直接发送，服务端打断旧流，token 落到新气泡
- 窗收起 ≠ 断连：连接保持、心跳继续；仅组件卸载断连
- 进 admin：PublicLayout 卸载 → `onUnmounted` 清理（心跳、socket），无泄漏
- connecting 中点发送：禁用输入区直到 ready（简化状态机）

## 不做的事（YAGNI）

- 自动重连
- localStorage 历史
- 消息撤回/复制/表情
- 服务端改动

## 已知限制（后续改进方向，不在本次范围）

服务端 agent 无 checkpointer，精灵无上下文记忆；Redis 短期记忆只写不读。后续可挂 SQLite checkpointer + 历史拉取端点，届时前端仅需加载时拉取一次。

## 验证方式

1. `npm run build` 通过（`blog/web/admin/`）
2. `/run-blog` skill 起全栈（app + agent/redis 容器），浏览器实测：悬浮球展开、auth 问候、流式输出、停笔、断线占位、移动端宽度模拟
3. `smoke.mjs` 增补自动检查：WS 连 `/chat/ws` 发 auth 收 `auth_result`（Node 原生 WebSocket，不引依赖）
