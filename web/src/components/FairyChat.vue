<template>
  <!-- 悬浮墨滴球 -->
  <div v-show="!visible" @click="open"
    class="fairy-ball fixed right-6 bottom-6 z-[900] w-14 h-14 rounded-full cursor-pointer flex items-center justify-center bg-ink text-paper transition-all duration-700 ease-in-out hover:ring-4 hover:ring-accent/20"
    :class="{ 'fairy-unread': unread }" role="button" aria-label="召唤笔墨精灵" :title="statusTitle[status]">
    <BrushIcon class="w-6 h-6" />
    <span v-if="unread" class="absolute top-0.5 right-0.5 w-2.5 h-2.5 rounded-full bg-cinnabar border-2 border-paper" />
  </div>

  <!-- 弹出聊天窗 -->
  <Transition name="fairy-pop">
    <div v-if="visible"
      class="fairy-win fixed right-6 bottom-6 z-[900] w-[380px] h-[560px] max-h-[calc(100vh-48px)] flex flex-col bg-paper border border-line shadow-sm overflow-hidden">
      <div class="flex items-center gap-1.5 shrink-0 h-12 px-3.5 bg-paper2 border-b border-line">
        <BrushIcon class="w-4 h-4 text-accent-strong" />
        <span class="text-sm font-light tracking-widest text-ink">笔墨精灵</span>
        <span class="w-2 h-2 rounded-full ml-0.5 bg-ink3 transition-colors duration-700"
          :class="{ 'bg-moss fairy-ready': status === 'ready', 'bg-cinnabar': status === 'offline' }" :title="statusTitle[status]" />
        <span v-if="status === 'ready' && remaining !== null" class="ml-auto text-[11px] text-ink3">今日余 {{ remaining }} 次</span>
        <button class="border-0 bg-transparent cursor-pointer text-ink3 hover:text-ink transition-colors duration-700 flex p-1 ml-auto"
          :class="status === 'ready' && remaining !== null ? '' : '!ml-0'" @click="visible = false" aria-label="关闭">
          <CloseIcon class="w-3.5 h-3.5" />
        </button>
      </div>

      <div ref="listEl" class="flex-1 overflow-y-auto p-3.5 flex flex-col gap-2.5" style="scrollbar-width: thin" @scroll="onScroll">
        <div v-if="status === 'offline'" class="fairy-sys">
          精灵云游去了… <button class="fairy-link" @click="connect">重新召唤</button>
        </div>
        <div v-else-if="status === 'connecting'" class="fairy-sys">精灵正在苏醒…</div>
        <div v-else-if="status === 'need-login'" class="fairy-sys">
          登录后可与笔墨精灵对话 <button class="fairy-link" @click="openLogin">去登录</button>
        </div>

        <div v-for="(m, i) in messages" :key="i" class="flex" :class="{ 'justify-end': m.role === 'user' }">
          <div v-if="m.role === 'system'" class="fairy-sys">{{ m.content }}</div>
          <div v-else class="max-w-[82%] px-3 py-2 text-sm leading-relaxed break-words whitespace-pre-wrap border transition-colors duration-700"
            :class="m.role === 'user' ? 'bg-transparent border-accent text-ink' : 'bg-paper2 border-line2 text-ink'">
            {{ m.content }}<span v-if="!m.done" class="fairy-caret" />
            <div v-if="m.done && m.interrupted" class="text-[11px] text-ink3 italic mt-0.5">（笔墨言未尽…）</div>
          </div>
        </div>
        <div v-if="retrieving" class="fairy-sys">精灵翻阅卷轴中…</div>
      </div>

      <div class="flex gap-2 items-end shrink-0 py-2.5 px-3 border-t border-line">
        <textarea v-model="draft" rows="1" maxlength="500" :disabled="status !== 'ready'"
          placeholder="与精灵对谈…" @input="autoGrow" @keydown.enter.ctrl.prevent="doSend"
          class="flex-1 min-w-0 resize-none focus:outline-none border border-line focus:border-accent bg-paper2 text-[13px] leading-relaxed py-2 px-2.5 max-h-[72px] placeholder:text-ink3 transition-colors duration-700 disabled:opacity-50" />
        <InkButton v-if="status === 'offline'" variant="ghost" size="sm" @click="connect">重新召唤</InkButton>
        <InkButton v-else-if="streaming" variant="ghost" size="sm" @click="interrupt">停笔</InkButton>
        <InkButton v-else variant="primary" size="sm" :disabled="!draft.trim() || status !== 'ready'" @click="doSend">发送</InkButton>
      </div>
    </div>
  </Transition>
</template>

<script setup>
import { ref, watch, nextTick, onUnmounted } from 'vue'
import { BrushOutline, CloseOutline } from '@vicons/ionicons5'
import InkButton from './ui/InkButton.vue'
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
/* 未读：缓慢呼吸不晃动 */
.fairy-unread { animation: fairy-breathe 4s ease-in-out infinite; }
@keyframes fairy-breathe { 0%, 100% { opacity: 1; } 50% { opacity: 0.75; } }
.fairy-ready { animation: fairy-pulse 2s ease-in-out infinite; }
@keyframes fairy-pulse { 0%, 100% { opacity: 1; } 50% { opacity: 0.4; } }

.fairy-sys { align-self: center; font-size: 12px; color: var(--ink3); font-style: italic; text-align: center; }
.fairy-link { border: none; background: none; cursor: pointer; padding: 0; color: var(--accent-strong); font-size: 12px; text-decoration: underline; }

/* 流式输出光标 */
.fairy-caret {
  display: inline-block; width: 2px; height: 14px; margin-left: 2px;
  background: var(--accent); vertical-align: -2px;
  animation: fairy-blink 1s step-end infinite;
}
@keyframes fairy-blink { 50% { opacity: 0; } }

/* 出入窗：纯墨色晕染，无位移无缩放 */
.fairy-pop-enter-active { transition: opacity 0.7s ease-in-out; }
.fairy-pop-leave-active { transition: opacity 0.7s ease-in-out; }
.fairy-pop-enter-from, .fairy-pop-leave-to { opacity: 0; }

/* 移动端：贴底 */
@media (max-width: 480px) {
  .fairy-ball { right: 16px; bottom: 16px; }
  .fairy-win { right: 16px; left: 16px; bottom: 16px; width: auto; height: 70vh; }
}
@media (prefers-reduced-motion: reduce) {
  .fairy-unread, .fairy-ready { animation: none; }
  .fairy-pop-enter-active, .fairy-pop-leave-active { transition: none; }
}
</style>
