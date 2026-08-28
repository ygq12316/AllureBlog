<template>
  <div class="comment-section">
    <h4 class="cs-title">💬 评论 ({{ total }})</h4>

    <!-- 评论列表 -->
    <div v-if="comments.length" class="cs-list">
      <div v-for="c in displayComments" :key="c.id" class="cs-item">
        <img :src="commentAvatar(c)" class="cs-avatar" />
        <div class="cs-body">
          <div class="cs-head">
            <span class="cs-nick">{{ c.nickname || '匿名' }}</span>
            <span v-if="c.visitor_uuid && c.visitor_uuid.startsWith('admin_')" class="cs-admin-badge">博主</span>
            <span v-if="c.signature" class="cs-sig">{{ c.signature }}</span>
            <time class="cs-time">{{ rel(c.created_at) }}</time>
          </div>
          <p class="cs-content">{{ c.content }}</p>
        </div>
      </div>
    </div>
    <div v-else class="cs-empty">还没有评论，来说两句吧</div>

    <!-- 折叠/展开 -->
    <div v-if="comments.length > 10" class="cs-toggle" @click="expanded = !expanded">
      <n-icon :component="expanded ? ChevronUp : ChevronDown" size="14" />
      {{ expanded ? '收起' : `展开更多 (${comments.length - 10}条)` }}
    </div>

    <!-- 未登录提示 -->
    <div v-if="!loggedIn" class="cs-input-row">
      <div class="cs-login-hint" @click="openLogin">👋 登录后参与评论</div>
    </div>
    <!-- 已登录输入区 -->
    <div v-else class="cs-input-row">
      <img :src="myAvatar" class="cs-avatar" />
      <div class="cs-input-wrap">
        <textarea v-model="content" class="cs-textarea" placeholder="写下你的想法..." maxlength="500" rows="2"
          @keydown.enter.ctrl="submit" />
        <div class="cs-input-foot">
          <span class="cs-count">{{ content.length }}/500</span>
          <n-button size="small" type="primary" @click="submit" :disabled="!content.trim()">发送</n-button>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, computed, onMounted, onUnmounted } from 'vue'
import { ChevronDown, ChevronUp } from '@vicons/ionicons5'
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

<style scoped>
.comment-section {
  max-width: 640px;
  margin: 48px auto 0;
  padding-top: 32px;
  border-top: 1px solid var(--card-border);
}
.cs-title {
  font-size: 15px;
  font-weight: 700;
  color: var(--text);
  margin: 0 0 20px;
}

.cs-list {
  margin-bottom: 16px;
}
.cs-item {
  display: flex;
  gap: 10px;
  padding: 12px 0;
  border-bottom: 1px solid var(--card-border);
}
.cs-avatar {
  width: 32px;
  height: 32px;
  border-radius: 50%;
  flex-shrink: 0;
  background: var(--tag-bg);
}
.cs-body {
  flex: 1;
  min-width: 0;
}
.cs-head {
  display: flex;
  align-items: center;
  gap: 8px;
  flex-wrap: wrap;
  margin-bottom: 4px;
}
.cs-nick {
  font-size: 13px;
  font-weight: 600;
  color: var(--gold);
}
.cs-admin-badge {
  font-size: 9px;
  padding: 1px 5px;
  background: var(--gold);
  color: #fff;
  border-radius: 3px;
}
.cs-sig {
  font-size: 10px;
  color: var(--muted);
}
.cs-time {
  font-size: 10px;
  color: var(--muted);
  margin-left: auto;
}
.cs-content {
  font-size: 14px;
  color: var(--text);
  margin: 0;
  line-height: 1.7;
  word-break: break-word;
}

.cs-empty {
  font-size: 13px;
  color: var(--muted);
  text-align: center;
  padding: 24px 0;
}

.cs-toggle {
  font-size: 12px;
  color: var(--gold);
  cursor: pointer;
  display: flex;
  align-items: center;
  gap: 4px;
  justify-content: center;
  padding: 8px 0 16px;
}
.cs-toggle:hover {
  text-decoration: underline;
}

.cs-input-row {
  display: flex;
  gap: 10px;
  align-items: flex-start;
}
.cs-input-wrap {
  flex: 1;
  min-width: 0;
}
.cs-textarea {
  width: 100%;
  border: 1px solid var(--card-border);
  border-radius: 8px;
  background: var(--tag-bg);
  color: var(--text);
  font-family: 'LXGW WenKai', serif;
  font-size: 13px;
  line-height: 1.6;
  padding: 10px 12px;
  resize: none;
  outline: none;
  caret-color: var(--gold);
}
.cs-textarea:focus {
  border-color: var(--gold);
}
.cs-textarea::placeholder {
  color: var(--muted);
  opacity: 0.5;
}
.cs-input-foot {
  display: flex;
  align-items: center;
  justify-content: space-between;
  margin-top: 6px;
}
.cs-count {
  font-size: 11px;
  color: var(--muted);
}
.cs-login-hint {
  flex: 1;
  padding: 14px;
  text-align: center;
  background: var(--tag-bg);
  border: 1px solid var(--card-border);
  border-radius: 8px;
  font-size: 13px;
  color: var(--muted);
  cursor: pointer;
  transition: border-color .2s, color .2s;
}
.cs-login-hint:hover { border-color: var(--gold); color: var(--gold); }
</style>
