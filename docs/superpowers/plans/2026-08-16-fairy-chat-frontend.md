# 笔墨精灵前端聊天入口 · 实现计划

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** 为 Vue SPA 补齐笔墨精灵聊天入口：右下角悬浮墨滴球 + 弹出聊天窗，接入已有 `/chat/ws` WebSocket 服务端。

**Architecture:** composable（`useFairyChat.js`，纯 WS 状态机）+ 单组件（`FairyChat.vue`，球与窗同文件）挂 `PublicLayout`；惰性建连、常驻心跳、手动重连；会话仅活在页面生命周期。设计全文见 `docs/superpowers/specs/2026-08-16-fairy-chat-frontend-design.md`。

**Tech Stack:** Vue 3 `<script setup>`、原生 WebSocket、Naive UI（`n-icon`/`n-button` 已全局挂载）、`@vicons/ionicons5`（已有依赖）。

## Global Constraints

- 仅改前端（`blog/web/admin/`）与驱动脚本（`.claude/skills/run-blog/smoke.mjs`），不改任何服务端代码
- 不引入新 npm 依赖（用原生 WebSocket；`@vicons/ionicons5` 项目已有）
- 视觉只用现有 CSS 变量：`--gold`、`--bg`、`--text`、`--muted`、`--card-border`、`--tag-bg`；字体 `LXGW WenKai`
- 注释与提交信息均为中文；代码风格对齐现有组件（紧凑、短变量名、单文件三段式）
- 本机 Git Bash 无 `make`/`python`；命令一律纯形式 + 绝对路径（`/e/pythonProject/web/blog`）
- 项目前端无测试框架，验证 = `npm run build` + 冒烟驱动 + 浏览器手动（agent 本机大概率起不来，以 offline 路径验证为主）
- agent 本机现状：`docker compose build agent` 因 pip 无镜像加速超时未跑通；agent 起来需要 `agent/.env` 里的 `DEEPSEEK_API_KEY`

---

### Task 1: vite dev 代理补 `/chat`

开发模式下 Vite 只代理 `/api` 和 `/uploads`（`vite.config.js:8-12`），前端连 `/chat/ws` 会打到 5173 自身 404。补一条带 `ws: true` 的代理指向本地 agent :8000。

**Files:**
- Modify: `blog/web/admin/vite.config.js`

**Interfaces:**
- Produces: dev 模式下 `/chat/*`（含 WebSocket upgrade）转发到 `http://localhost:8000`，供 Task 2 的组件在 `npm run dev` 时使用。生产环境不走这里（Caddy 负责），行为互不影响。

- [ ] **Step 1: 修改 proxy 配置**

```js
  server: {
    proxy: {
      '/api': 'http://localhost:8080',
      '/uploads': 'http://localhost:8080',
      // 笔墨精灵 agent — 生产由 Caddy 转发，开发模式在此代理（含 WebSocket）
      '/chat': { target: 'http://localhost:8000', ws: true },
    },
  },
```

- [ ] **Step 2: 构建验证语法无误**

Run: `cd /e/pythonProject/web/blog/web/admin && npm run build`
Expected: `✓ built in ...`，无报错。

- [ ] **Step 3: Commit**

```bash
cd /e/pythonProject/web && git add blog/web/admin/vite.config.js
git commit -m "feat: vite dev 代理 /chat 到本地 agent（含 WebSocket）"
```

---

### Task 2: useFairyChat.js + FairyChat.vue + PublicLayout 挂载

一次性交付完整聊天入口（三个文件互相引用，拆开提交无法独立通过构建验证——composable/组件在未挂载时会被 tree-shake，build 不能证明其正确）。内部按步骤逐文件完成。

**Files:**
- Create: `blog/web/admin/src/composables/useFairyChat.js`
- Create: `blog/web/admin/src/components/FairyChat.vue`
- Modify: `blog/web/admin/src/layouts/PublicLayout.vue`（import + 模板挂载）

**Interfaces:**
- Consumes: `useVisitor()`（`blog/web/admin/src/composables/useVisitor.js`）返回的 `{ visitor, init }`；WS 协议 `auth/message/interrupt/ping` ↔ `auth_result/token/done/rejected/error/pong`（`blog/agent/ws/protocol.py`）
- Produces: `useFairyChat()` → `{ status, messages, streaming, connect, send, interrupt, disconnect }`：
  - `status: Ref<'idle'|'connecting'|'ready'|'offline'>`
  - `messages: Ref<Array<{role:'user'|'fairy'|'system', content:string, done:boolean, interrupted?:boolean}>>`
  - `streaming: Ref<boolean>`
  - `connect(): Promise<void>` — 幂等（connecting/ready 时直接返回）；offline 时重连
  - `send(content: string): boolean` — 成功发出返回 true（此时已本地追加 user 气泡与 fairy 占位气泡）；streaming 中**允许**发送（服务端把新 message 视作打断）
  - `interrupt(): void`、`disconnect(): void`

**关键状态机规则**（实现时勿偏离）：
- `token` 追加到**最后一条**未完成 fairy 气泡（无则新建）；`done` 定稿**最早一条**未完成 fairy 气泡（连发场景：服务端先发旧流的 `done(interrupted)` 再开新流，FIFO 定稿才能对准旧气泡）
- `onclose`/`onerror` 双触发防重（`status` 已是 offline 则跳过）
- 断线时 streaming 中的气泡定稿并追加"（连接中断）"

- [ ] **Step 1: 写 `useFairyChat.js`**

```js
import { ref } from 'vue'
import { useVisitor } from './useVisitor'

// 笔墨精灵聊天 — WebSocket 状态机（纯逻辑，视图在 FairyChat.vue）
export function useFairyChat() {
  const { visitor, init } = useVisitor()
  const status = ref('idle') // idle | connecting | ready | offline
  const messages = ref([])
  const streaming = ref(false)
  let ws = null
  let pingTimer = null

  // 找气泡：fromEnd=true 取最晚（token 追加目标），false 取最早（done 定稿目标）
  function openFairy(fromEnd) {
    const list = fromEnd ? [...messages.value].reverse() : messages.value
    return list.find(m => m.role === 'fairy' && !m.done)
  }

  function handle(m) {
    if (m.type === 'auth_result') {
      status.value = 'ready'
      if (m.greeting) messages.value.push({ role: 'fairy', content: m.greeting, done: true })
      startPing()
    } else if (m.type === 'token') {
      let b = openFairy(true)
      if (!b) { b = { role: 'fairy', content: '', done: false }; messages.value.push(b) }
      b.content += m.content
    } else if (m.type === 'done') {
      const b = openFairy(false)
      if (b) { b.done = true; b.interrupted = !!m.interrupted }
      streaming.value = false
    } else if (m.type === 'rejected' || m.type === 'error') {
      messages.value.push({ role: 'system', content: m.display || '精灵暂时无法回应' })
      streaming.value = false
    }
    // pong 仅保活，忽略
  }

  async function connect() {
    if (status.value === 'connecting' || status.value === 'ready') return
    await init() // uuid 由 useVisitor 自动生成并持久化，此处必可得
    if (!visitor.value) return
    status.value = 'connecting'
    const proto = location.protocol === 'https:' ? 'wss' : 'ws'
    try {
      ws = new WebSocket(`${proto}://${location.host}/chat/ws`)
    } catch {
      onDown()
      return
    }
    ws.onopen = () => {
      ws.send(JSON.stringify({
        type: 'auth',
        visitor_uuid: visitor.value.uuid,
        visitor_name: visitor.value.nickname,
      }))
    }
    ws.onmessage = ev => {
      try { handle(JSON.parse(ev.data)) } catch {}
    }
    ws.onclose = onDown
    ws.onerror = onDown
  }

  function send(content) {
    const text = (content || '').trim()
    if (!text || !ws || ws.readyState !== WebSocket.OPEN) return false
    messages.value.push({ role: 'user', content: text, done: true })
    messages.value.push({ role: 'fairy', content: '', done: false })
    streaming.value = true
    ws.send(JSON.stringify({ type: 'message', content: text }))
    return true
  }

  function interrupt() {
    if (ws && ws.readyState === WebSocket.OPEN && streaming.value) {
      ws.send(JSON.stringify({ type: 'interrupt' }))
    }
  }

  // 心跳：25s 低于常见代理 60s 空闲超时
  function startPing() {
    stopPing()
    pingTimer = setInterval(() => {
      if (ws && ws.readyState === WebSocket.OPEN) ws.send(JSON.stringify({ type: 'ping' }))
    }, 25000)
  }
  function stopPing() {
    if (pingTimer) { clearInterval(pingTimer); pingTimer = null }
  }

  function onDown() {
    if (status.value === 'offline') return // onclose/onerror 双触发防重
    status.value = 'offline'
    stopPing()
    const b = openFairy(false)
    if (b) { b.done = true; b.content += '（连接中断）' }
    streaming.value = false
    ws = null
  }

  function disconnect() {
    stopPing()
    if (ws) {
      ws.onclose = null; ws.onerror = null; ws.onmessage = null
      try { ws.close() } catch {}
      ws = null
    }
    status.value = 'idle'
  }

  return { status, messages, streaming, connect, send, interrupt, disconnect }
}
```

- [ ] **Step 2: 写 `FairyChat.vue`**

```vue
<template>
  <!-- 悬浮墨滴球 -->
  <div v-show="!visible" class="fairy-ball" :class="{ 'fairy-ball--unread': unread }" @click="open">
    <n-icon :component="BrushIcon" size="24" color="#fff" />
    <span v-if="unread" class="fairy-dot" />
  </div>

  <!-- 弹出聊天窗 -->
  <transition name="fairy-pop">
    <div v-if="visible" class="fairy-win">
      <div class="fairy-head">
        <n-icon :component="BrushIcon" size="16" color="var(--gold)" />
        <span class="fairy-title">笔墨精灵</span>
        <span class="fairy-status" :class="'fairy-status--' + status" :title="statusTitle" />
        <button class="fairy-close" @click="visible = false"><n-icon :component="CloseIcon" size="14" /></button>
      </div>

      <div ref="listEl" class="fairy-body" @scroll="onScroll">
        <div v-if="status === 'offline'" class="fairy-sys">
          精灵云游去了… <button class="fairy-link" @click="connect">重新召唤</button>
        </div>
        <div v-else-if="status === 'connecting'" class="fairy-sys">精灵正在苏醒…</div>

        <div v-for="(m, i) in messages" :key="i" class="fairy-row" :class="{ 'fairy-row--me': m.role === 'user' }">
          <div v-if="m.role === 'system'" class="fairy-sys">{{ m.content }}</div>
          <div v-else class="fairy-bubble" :class="{ 'fairy-bubble--me': m.role === 'user' }">
            {{ m.content }}<span v-if="!m.done" class="fairy-caret" />
            <div v-if="m.done && m.interrupted" class="fairy-cut">（笔墨言未尽…）</div>
          </div>
        </div>
      </div>

      <div class="fairy-input">
        <textarea v-model="draft" class="fairy-ta" rows="1" maxlength="500"
          :disabled="status !== 'ready'" placeholder="与精灵对谈…"
          @input="autoGrow" @keydown.enter.ctrl.prevent="doSend" />
        <button v-if="status === 'offline'" class="fairy-btn fairy-btn--ghost" @click="connect">重新召唤</button>
        <button v-else-if="streaming" class="fairy-btn fairy-btn--ghost" @click="interrupt">停笔</button>
        <button v-else class="fairy-btn" :disabled="!draft.trim() || status !== 'ready'" @click="doSend">发送</button>
      </div>
    </div>
  </transition>
</template>

<script setup>
import { ref, watch, nextTick, onUnmounted } from 'vue'
import { BrushOutline, CloseOutline } from '@vicons/ionicons5'
import { useFairyChat } from '../composables/useFairyChat'

const BrushIcon = BrushOutline, CloseIcon = CloseOutline
const { status, messages, streaming, connect, send, interrupt, disconnect } = useFairyChat()

const visible = ref(false)
const unread = ref(false)
const draft = ref('')
const listEl = ref(null)
let follow = true // 流式期间自动跟随滚动，用户上滚则暂停

const statusTitle = { idle: '未连接', connecting: '连接中', ready: '在线', offline: '离线' }

function open() {
  visible.value = true
  unread.value = false
  connect() // 惰性建连：首次打开才连，之后常驻
}

function doSend() {
  // send 返回 false（离线等）时保留草稿，绝不静默丢弃输入
  if (send(draft.value)) draft.value = ''
}

function autoGrow(e) {
  const el = e.target
  el.style.height = 'auto'
  el.style.height = Math.min(el.scrollHeight, 72) + 'px'
}

function onScroll() {
  const el = listEl.value
  if (!el) return
  follow = el.scrollTop + el.clientHeight >= el.scrollHeight - 40
}

async function scrollToEnd() {
  await nextTick()
  if (listEl.value && follow) listEl.value.scrollTop = listEl.value.scrollHeight
}

// 内容变化：窗关且有新消息 → 未读点；跟随滚动
watch(() => messages.value.map(m => m.content).join('\u0000'), () => {
  const last = messages.value[messages.value.length - 1]
  if (!visible.value && last && last.role !== 'user') unread.value = true
  scrollToEnd()
})

watch(visible, v => { if (v) { unread.value = false; scrollToEnd() } })

// 进 admin / 整页刷新时 PublicLayout 卸载，组件随之销毁 → 清理连接
onUnmounted(disconnect)
</script>

<style scoped>
/* ── 悬浮球 ── */
.fairy-ball {
  position: fixed; right: 24px; bottom: 24px; z-index: 900;
  width: 56px; height: 56px; border-radius: 50%; cursor: pointer;
  display: flex; align-items: center; justify-content: center;
  background: radial-gradient(circle at 35% 30%, rgba(184,148,76,0.95), rgba(120,105,81,0.92) 70%);
  box-shadow: 0 4px 18px rgba(120,105,81,0.45);
  transition: transform .25s;
  animation: fairy-breathe 4s ease-in-out infinite;
}
.fairy-ball:hover { transform: scale(1.08); }
.fairy-ball--unread { animation: fairy-breathe 4s ease-in-out infinite, fairy-shake .5s ease; }
.fairy-dot {
  position: absolute; top: 2px; right: 2px; width: 10px; height: 10px;
  border-radius: 50%; background: var(--gold); border: 2px solid var(--bg);
}
@keyframes fairy-breathe {
  0%, 100% { box-shadow: 0 4px 18px rgba(120,105,81,0.45), 0 0 0 0 rgba(184,148,76,0.35); }
  50% { box-shadow: 0 4px 18px rgba(120,105,81,0.45), 0 0 0 12px rgba(184,148,76,0); }
}
@keyframes fairy-shake {
  0%, 100% { transform: rotate(0); }
  25% { transform: rotate(-8deg); }
  75% { transform: rotate(8deg); }
}

/* ── 聊天窗 ── */
.fairy-win {
  position: fixed; right: 24px; bottom: 24px; z-index: 900;
  width: 380px; height: 560px; max-height: calc(100vh - 48px);
  display: flex; flex-direction: column;
  background: var(--bg); border: 1px solid var(--card-border); border-radius: 14px;
  box-shadow: 0 12px 40px rgba(0,0,0,0.18);
  overflow: hidden; transform-origin: bottom right;
}
.fairy-pop-enter-active { transition: opacity .25s, transform .25s; }
.fairy-pop-leave-active { transition: opacity .15s, transform .15s; }
.fairy-pop-enter-from, .fairy-pop-leave-to { opacity: 0; transform: scale(0.6); }

.fairy-head {
  display: flex; align-items: center; gap: 6px; flex-shrink: 0;
  height: 48px; padding: 0 14px;
  background: var(--tag-bg); border-bottom: 1px solid var(--card-border);
  font-family: 'LXGW WenKai', serif;
}
.fairy-title { font-size: 14px; font-weight: 700; color: var(--text); }
.fairy-status { width: 8px; height: 8px; border-radius: 50%; margin-left: 2px; background: var(--muted); }
.fairy-status--ready { background: var(--gold); animation: fairy-pulse 2s infinite; }
@keyframes fairy-pulse { 0%,100% { opacity: 1; } 50% { opacity: .4; } }
.fairy-close {
  margin-left: auto; border: none; background: none; cursor: pointer;
  color: var(--muted); display: flex; padding: 4px; border-radius: 4px;
}
.fairy-close:hover { color: var(--gold); }

.fairy-body {
  flex: 1; overflow-y: auto; padding: 14px;
  display: flex; flex-direction: column; gap: 10px;
  scrollbar-width: thin;
}
.fairy-row { display: flex; }
.fairy-row--me { justify-content: flex-end; }
.fairy-bubble {
  max-width: 82%; padding: 8px 12px; border-radius: 10px 10px 10px 2px;
  background: var(--tag-bg); color: var(--text);
  font-family: 'LXGW WenKai', serif; font-size: 14px; line-height: 1.7;
  word-break: break-word; white-space: pre-wrap;
}
.fairy-bubble--me {
  border-radius: 10px 10px 2px 10px;
  background: transparent; border: 1px solid var(--gold);
}
.fairy-caret {
  display: inline-block; width: 2px; height: 14px; margin-left: 2px;
  background: var(--gold); vertical-align: -2px;
  animation: fairy-blink 1s step-end infinite;
}
@keyframes fairy-blink { 50% { opacity: 0; } }
.fairy-cut { font-size: 11px; color: var(--muted); font-style: italic; margin-top: 2px; }
.fairy-sys {
  align-self: center; font-size: 12px; color: var(--muted);
  font-style: italic; text-align: center;
}
.fairy-link {
  border: none; background: none; cursor: pointer; padding: 0;
  color: var(--gold); font-size: 12px; text-decoration: underline;
}

.fairy-input {
  display: flex; gap: 8px; align-items: flex-end; flex-shrink: 0;
  padding: 10px 12px; border-top: 1px solid var(--card-border);
}
.fairy-ta {
  flex: 1; min-width: 0; resize: none; outline: none;
  border: 1px solid var(--card-border); border-radius: 8px;
  background: var(--tag-bg); color: var(--text);
  font-family: 'LXGW WenKai', serif; font-size: 13px; line-height: 1.6;
  padding: 8px 10px; max-height: 72px; caret-color: var(--gold);
}
.fairy-ta:focus { border-color: var(--gold); }
.fairy-ta:disabled { opacity: .5; }
.fairy-btn {
  flex-shrink: 0; border: none; border-radius: 8px; cursor: pointer;
  padding: 8px 14px; font-size: 13px;
  background: var(--gold); color: #fff;
  transition: opacity .2s;
}
.fairy-btn:disabled { opacity: .4; cursor: default; }
.fairy-btn--ghost {
  background: transparent; color: var(--gold);
  border: 1px solid var(--gold);
}

/* ── 移动端：贴底抽屉 ── */
@media (max-width: 480px) {
  .fairy-ball { right: 16px; bottom: 16px; }
  .fairy-win {
    right: 16px; left: 16px; bottom: 16px; width: auto; height: 70vh;
    transform-origin: bottom center;
  }
}
</style>
```

- [ ] **Step 3: 挂载到 `PublicLayout.vue`**

`blog/web/admin/src/layouts/PublicLayout.vue` 两处修改：

模板——`<VisitorSetup>` 之后（第 21 行后）加：

```html
    <FairyChat />
```

script——import 区（`VisitorSetup` import 之后）加：

```js
import FairyChat from '../components/FairyChat.vue'
```

- [ ] **Step 4: 构建验证**

Run: `cd /e/pythonProject/web/blog/web/admin && npm run build`
Expected: `✓ built in ...`，无报错、无新 warning。

- [ ] **Step 5: 起后端做浏览器验证（offline 路径）**

agent 本机大概率起不来（见 Global Constraints），offline 路径恰好是本次可完整验证的路径：

```bash
cd /e/pythonProject/web/blog
JWT_SECRET=dev-secret ADMIN_PASSWORD=dev-pass nohup go run ./cmd/server > /tmp/blog-server.log 2>&1 &
sleep 16 && grep -E "已启动" /tmp/blog-server.log
```

截图确认（Edge 无头，`QQBrowser user data path not found` 为已知噪音）：

```bash
"/c/Program Files (x86)/Microsoft/Edge/Application/msedge.exe" --headless --disable-gpu \
  "--screenshot=C:/Users/Allure/AppData/Local/Temp/blog-fairy.png" --window-size=1280,2400 \
  http://localhost:8080/
```

Expected: 截图中右下角出现墨滴悬浮球。交互验证（`npm run dev` + 浏览器手点）：点球 → 窗展开 → agent 未起时显示"精灵云游去了… 重新召唤"，输入框禁用（offline 态 textarea `:disabled` 生效）。若 agent 已跑通（:8000/health 通），追加验证：auth 问候语出现、发消息流式输出、停笔按钮、Ctrl+Enter 发送。

- [ ] **Step 6: Commit**

```bash
cd /e/pythonProject/web && git add blog/web/admin/src/composables/useFairyChat.js blog/web/admin/src/components/FairyChat.vue blog/web/admin/src/layouts/PublicLayout.vue
git commit -m "feat: 笔墨精灵前端聊天入口 — 悬浮墨滴球 + 弹出聊天窗

useFairyChat composable 管理 WS 状态机（惰性建连/25s 心跳/流式拼接/
手动重连），FairyChat.vue 承载球与窗，挂 PublicLayout 跨公开路由常驻。
会话仅活在页面生命周期，刷新即清空。"
```

---

### Task 3: smoke.mjs 增补精灵 WS 冒烟段

**Files:**
- Modify: `.claude/skills/run-blog/smoke.mjs`（在第 85 行 `ok('评论 WebSocket 广播', ...)` 之后、末尾汇总之前插入）

**Interfaces:**
- Consumes: Task 2 落地后的 `/chat/ws` 协议（auth → auth_result）；agent 健康端点 `GET /health`
- Produces: 冒烟新增 1 项断言「精灵 WebSocket auth」；agent 不可达时打印 `SKIP` 不计失败（保持现有冒烟在无 agent 环境全绿）

- [ ] **Step 1: 先跑现有冒烟确认基线全绿**

前置：后端已启动（Task 2 Step 5 的进程还在跑）。

Run: `node /e/pythonProject/web/.claude/skills/run-blog/smoke.mjs`
Expected: `全部通过`（10 项 PASS）。

- [ ] **Step 2: 插入精灵冒烟段**

```js
// ── 6. 笔墨精灵 agent（可选：未部署则跳过，不计失败）──
const AGENT = process.argv.includes('--agent')
  ? process.argv[process.argv.indexOf('--agent') + 1]
  : 'http://localhost:8000';

let agentUp = false;
try {
  const h = await fetch(AGENT + '/health', { signal: AbortSignal.timeout(1500) });
  agentUp = h.ok;
} catch {}

if (!agentUp) {
  console.log(`SKIP  笔墨精灵 agent（${AGENT}/health 不可达）`);
} else {
  const fairy = await new Promise(resolve => {
    const ws = new WebSocket(AGENT.replace(/^http/, 'ws') + '/chat/ws');
    const timer = setTimeout(() => { try { ws.close(); } catch {} resolve({ ok: false, why: 'timeout' }); }, 8000);
    ws.onopen = () => ws.send(JSON.stringify({ type: 'auth', visitor_uuid: 'smoke-' + uuid, visitor_name: 'smoke' }));
    ws.onmessage = e => {
      clearTimeout(timer);
      const data = JSON.parse(String(e.data));
      try { ws.close(); } catch {}
      resolve({ ok: data.type === 'auth_result' && data.success === true, why: 'type=' + data.type });
    };
    ws.onerror = () => { clearTimeout(timer); resolve({ ok: false, why: 'ws error' }); };
  });
  ok('精灵 WebSocket auth', fairy.ok, fairy.why || '');
}
```

同时更新文件头注释第 5 行覆盖说明为：

```js
// 覆盖: 健康页面 → 公开 API → 管理员登录 → 鉴权写操作 → 访客注册 → 评论 + WebSocket 实时广播 → 精灵 agent（可选）
```

- [ ] **Step 3: 无 agent 环境验证 SKIP 路径**

Run: `node /e/pythonProject/web/.claude/skills/run-blog/smoke.mjs`
Expected: 原有 10 项 PASS + 一行 `SKIP  笔墨精灵 agent（http://localhost:8000/health 不可达）`，退出码 0。

- [ ] **Step 4: （agent 可用时）验证 PASS 路径**

前置：`docker compose up -d redis` 且 agent 构建成功并监听 :8000（本机若仍未跑通则跳过本步，在交付说明中注明）。auth_result 由 handler 直接返回，不消耗 DeepSeek 调用。

Run: `node /e/pythonProject/web/.claude/skills/run-blog/smoke.mjs`
Expected: `精灵 WebSocket auth  PASS  type=auth_result`，整体 `全部通过`。

- [ ] **Step 5: Commit**

```bash
cd /e/pythonProject/web && git add .claude/skills/run-blog/smoke.mjs
git commit -m "test: 冒烟驱动增补精灵 WS auth 检查（agent 不可达时 SKIP）"
```

---

## 计划自审记录

- **Spec 覆盖**：形态/形象/历史/组织/重连五项决策 → Task 2；vite dev 代理（spec 未列，属实现必要补充）→ Task 1；冒烟增补 → Task 3；错误矩阵各条均落在 composable `handle`/`onDown` 与组件 disabled/占位逻辑中 ✓
- **占位符扫描**：无 TBD/TODO；所有代码步骤含完整代码 ✓
- **类型一致性**：`useFairyChat` 返回值签名在 Task 2 Interfaces 与组件解构一致；消息对象 `{role, content, done, interrupted?}` 在 composable/组件/watch 中一致 ✓
