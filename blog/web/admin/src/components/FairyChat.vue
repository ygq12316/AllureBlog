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
        <span v-if="status === 'ready' && remaining !== null" class="fairy-quota">今日余 {{ remaining }} 次</span>
        <button class="fairy-close" @click="visible = false"><n-icon :component="CloseIcon" size="14" /></button>
      </div>

      <div ref="listEl" class="fairy-body" @scroll="onScroll">
        <div v-if="status === 'offline'" class="fairy-sys">
          精灵云游去了… <button class="fairy-link" @click="connect">重新召唤</button>
        </div>
        <div v-else-if="status === 'connecting'" class="fairy-sys">精灵正在苏醒…</div>
        <div v-else-if="status === 'need-login'" class="fairy-sys">
          登录后可与笔墨精灵对话 <button class="fairy-link" @click="openLogin">去登录</button>
        </div>

        <div v-for="(m, i) in messages" :key="i" class="fairy-row" :class="{ 'fairy-row--me': m.role === 'user' }">
          <div v-if="m.role === 'system'" class="fairy-sys">{{ m.content }}</div>
          <div v-else class="fairy-bubble" :class="{ 'fairy-bubble--me': m.role === 'user' }">
            {{ m.content }}<span v-if="!m.done" class="fairy-caret" />
            <div v-if="m.done && m.interrupted" class="fairy-cut">（笔墨言未尽…）</div>
          </div>
        </div>
        <div v-if="retrieving" class="fairy-sys">精灵翻阅卷轴中…</div>
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
import { useVisitor } from '../composables/useVisitor'

const BrushIcon = BrushOutline, CloseIcon = CloseOutline
const { status, messages, streaming, remaining, retrieving, connect, send, interrupt, disconnect } = useFairyChat()
const { openLogin } = useVisitor()

const visible = ref(false)
const unread = ref(false)
const draft = ref('')
const listEl = ref(null)
let follow = true // 流式期间自动跟随滚动，用户上滚则暂停

const statusTitle = { idle: '未连接', connecting: '连接中', ready: '在线', 'need-login': '未登录', offline: '离线' }

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
watch(() => messages.value.map(m => m.content).join(''), () => {
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
.fairy-quota { margin-left: auto; font-size: 11px; color: var(--muted); }
.fairy-close {
  border: none; background: none; cursor: pointer;
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
