<template>
  <div class="max-w-[640px] mx-auto mt-12 pt-8 border-t border-line">
    <h4 class="text-[15px] font-light tracking-widest text-ink m-0 mb-5">评论 ({{ total }})</h4>

    <!-- 评论列表 -->
    <div v-if="comments.length" class="mb-4">
      <div v-for="c in displayComments" :key="c.id" class="flex gap-2.5 py-3 border-b border-line2">
        <img :src="commentAvatar(c)" class="w-8 h-8 rounded-full shrink-0 bg-paper2" alt="" />
        <div class="flex-1 min-w-0">
          <div class="flex items-center gap-2 flex-wrap mb-1">
            <span class="text-[13px] text-accent-strong">{{ c.nickname || '匿名' }}</span>
            <span v-if="c.visitor_uuid && c.visitor_uuid.startsWith('admin_')"
              class="text-[9px] px-1.5 py-px border border-cinnabar/60 text-cinnabar">博主</span>
            <span v-if="c.signature" class="text-[10px] text-ink3">{{ c.signature }}</span>
            <time class="text-[10px] text-ink3 ml-auto">{{ rel(c.created_at) }}</time>
          </div>
          <p class="text-sm text-ink m-0 leading-relaxed break-words">{{ c.content }}</p>
        </div>
      </div>
    </div>
    <div v-else class="text-[13px] text-ink3 text-center py-6">还没有评论，来说两句吧</div>

    <!-- 折叠/展开 -->
    <div v-if="comments.length > 10" @click="expanded = !expanded"
      class="text-xs text-accent-strong cursor-pointer flex items-center justify-center gap-1 py-2 pb-4 transition-colors duration-700 hover:text-ink">
      <component :is="expanded ? ChevronUp : ChevronDown" class="w-3.5 h-3.5" />
      {{ expanded ? '收起' : `展开更多 (${comments.length - 10}条)` }}
    </div>

    <!-- 未登录提示 -->
    <div v-if="!loggedIn" class="flex">
      <div class="flex-1 py-3.5 text-center bg-paper2 border border-line text-[13px] font-light text-ink2 cursor-pointer transition-colors duration-700 hover:border-accent hover:text-accent-strong"
        @click="openLogin">登录后参与评论</div>
    </div>
    <!-- 已登录输入区 -->
    <div v-else class="flex gap-2.5 items-start">
      <img :src="myAvatar" class="w-8 h-8 rounded-full shrink-0 bg-paper2" alt="" />
      <div class="flex-1 min-w-0">
        <textarea v-model="content" placeholder="写下你的想法..." maxlength="500" rows="2"
          @keydown.enter.ctrl="submit"
          class="w-full bg-paper2 border border-line focus:outline-none focus:border-accent text-[13px] leading-relaxed p-2.5 resize-none placeholder:text-ink3 transition-colors duration-700" />
        <div class="flex items-center justify-between mt-1.5">
          <span class="text-[11px] text-ink3">{{ content.length }}/500</span>
          <InkButton variant="primary" size="sm" @click="submit" :disabled="!content.trim()">发送</InkButton>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, computed, onMounted, onUnmounted } from 'vue'
import { ChevronDown, ChevronUp } from '@vicons/ionicons5'
import InkButton from './ui/InkButton.vue'
import { useVisitor } from '../composables/useVisitor'
import { listComments, createComment } from '../api/comments'
import { rel } from '../utils/format'
import { dicebearUrl } from '../utils/avatar'

const props = defineProps({ noteId: { type: Number, required: true } })

const { visitor, account, avatarUrl: myAvatar, isSetUp, init, update, openSetup, openLogin } = useVisitor()

const loggedIn = computed(() => !!account.value)

const comments = ref([])
const total = ref(0)
const content = ref('')
const expanded = ref(false)
let socket = null

const displayComments = computed(() => {
  if (expanded.value) return comments.value
  return comments.value.slice(-10)
})

onMounted(async () => {
  await init()
  try {
    const data = await listComments(props.noteId)
    comments.value = data.comments || []
    total.value = data.total || comments.value.length
  } catch {}

  // WebSocket 实时同步
  const proto = location.protocol === 'https:' ? 'wss' : 'ws'
  const wsUrl = `${proto}://${location.host}/api/notes/${props.noteId}/ws`
  socket = new WebSocket(wsUrl)
  socket.onmessage = (ev) => {
    try {
      const msg = JSON.parse(ev.data)
      if (msg.type === 'comment' && msg.comment) {
        if (!comments.value.find(c => c.id === msg.comment.id)) {
          comments.value.push(msg.comment)
          total.value++
          expanded.value = true
        }
      }
    } catch {}
  }
})

// 组件卸载时关闭连接，避免跨路由累积泄漏
onUnmounted(() => {
  if (socket) {
    socket.onmessage = null
    socket.close()
    socket = null
  }
})

async function submit() {
  if (!content.value.trim()) return
  try {
    await createComment(props.noteId, {
      visitor_uuid: visitor.value.uuid,
      content: content.value.trim(),
    })
  } catch {
    return // 发送失败：保留输入内容，不静默清空
  }
  content.value = ''
  // 刷新
  const data = await listComments(props.noteId)
  comments.value = data.comments || []
  total.value = data.total || comments.value.length
  expanded.value = true
}

function commentAvatar(c) {
  return c.avatar_url || dicebearUrl(c.avatar_style || 'lorelei', c.visitor_uuid)
}
</script>
