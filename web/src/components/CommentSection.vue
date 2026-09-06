<template>
  <div class="max-w-[640px] mx-auto mt-12 pt-8 border-t border-line">
    <h4 class="text-[15px] font-light tracking-widest text-ink m-0 mb-6">评论 · {{ total }} 则</h4>

    <!-- 留言手札：根评论倒序（新札在上），回复顺序缩进其下 -->
    <div v-if="comments.length">
      <article v-for="node in displayTree" :key="node.c.id" class="py-5 border-b border-line2 first:pt-0">
        <div class="group flex gap-3">
          <img :src="commentAvatar(node.c)" class="w-9 h-9 rounded-full shrink-0 bg-paper2" alt="" />
          <div class="flex-1 min-w-0">
            <div class="flex items-baseline gap-2 flex-wrap">
              <span class="text-[13px] text-accent-strong">{{ node.c.nickname || '匿名' }}</span>
              <span v-if="isAdmin(node.c)" class="text-[9px] px-1 py-px border border-cinnabar/70 text-cinnabar -rotate-3 inline-block select-none">博主印</span>
              <span v-if="node.c.signature" class="text-[10px] text-ink3">{{ node.c.signature }}</span>
              <time class="text-[10px] text-ink3 ml-auto" :title="new Date(node.c.created_at).toLocaleString('zh-CN')">
                {{ node.c.temp ? '墨迹未干…' : rel(node.c.created_at) }}
              </time>
              <span class="text-[10px] text-ink3 tabular-nums">#{{ floorOf.get(node.c.id) }}</span>
            </div>
            <p class="text-sm text-ink m-0 mt-1.5 leading-relaxed break-words">{{ node.c.content }}</p>
            <button v-if="account" @click="startReply(node.c)"
              class="mt-1.5 text-[11px] text-ink3 hover:text-accent-strong bg-transparent border-0 cursor-pointer p-0 transition-colors duration-700">回复</button>
          </div>
        </div>

        <!-- 二级回复 -->
        <div v-if="node.replies.length" class="ml-10 mt-3 pl-3.5 border-l border-line2">
          <div v-for="r in node.replies" :key="r.c.id" class="flex gap-2.5 py-2.5 border-b border-line2/60 last:border-b-0">
            <img :src="commentAvatar(r.c)" class="w-7 h-7 rounded-full shrink-0 bg-paper2" alt="" />
            <div class="flex-1 min-w-0">
              <div class="flex items-baseline gap-2 flex-wrap">
                <span class="text-xs text-accent-strong">{{ r.c.nickname || '匿名' }}</span>
                <span v-if="isAdmin(r.c)" class="text-[9px] px-1 py-px border border-cinnabar/70 text-cinnabar -rotate-3 inline-block select-none">博主印</span>
                <time class="text-[10px] text-ink3 ml-auto" :title="new Date(r.c.created_at).toLocaleString('zh-CN')">
                  {{ r.c.temp ? '墨迹未干…' : rel(r.c.created_at) }}
                </time>
              </div>
              <p class="text-[13px] text-ink m-0 mt-1 leading-relaxed break-words">
                <span v-if="replyNick(r.c)" class="text-ink3">回复 @{{ replyNick(r.c) }}：</span>{{ r.c.content }}
              </p>
              <button v-if="account" @click="startReply(r.c)"
                class="mt-1 text-[11px] text-ink3 hover:text-accent-strong bg-transparent border-0 cursor-pointer p-0 transition-colors duration-700">回复</button>
            </div>
          </div>
        </div>
      </article>
    </div>
    <div v-else class="text-[13px] text-ink3 text-center py-8">手札空白，来落第一笔吧</div>

    <!-- 折叠/展开 -->
    <div v-if="tree.length > 10" @click="expanded = !expanded"
      class="text-xs text-accent-strong cursor-pointer flex items-center justify-center gap-1 py-3 transition-colors duration-700 hover:text-ink">
      {{ expanded ? '收起' : `展开更早的手札 (${tree.length - 10})` }}
    </div>

    <!-- 回复目标 -->
    <div v-if="replyTarget" class="flex items-center gap-2 mb-2 text-[11px] text-ink2">
      <span class="px-2 py-1 border border-line bg-paper2">回复 @{{ replyTarget.nickname }}</span>
      <button @click="replyTarget = null"
        class="text-ink3 hover:text-cinnabar bg-transparent border-0 cursor-pointer p-0 transition-colors duration-700" aria-label="取消回复">×</button>
    </div>

    <!-- 未登录 -->
    <div v-if="!account" class="flex">
      <div class="flex-1 py-3.5 text-center bg-paper2 border border-line text-[13px] font-light text-ink2 cursor-pointer transition-colors duration-700 hover:border-accent hover:text-accent-strong"
        @click="openLogin">登录后落笔</div>
    </div>
    <!-- 已登录输入区 -->
    <div v-else class="flex gap-3 items-start mt-5">
      <img :src="myAvatar" class="w-9 h-9 rounded-full shrink-0 bg-paper2" alt="" />
      <div class="flex-1 min-w-0">
        <textarea v-model="content" :placeholder="replyTarget ? `回复 @${replyTarget.nickname}…` : '留下你的手札…'" maxlength="500" rows="2"
          @keydown.enter.ctrl="submit"
          class="w-full bg-paper2 border border-line focus:outline-none focus:border-accent text-[13px] leading-relaxed p-2.5 resize-none placeholder:text-ink3 transition-colors duration-700" />
        <div class="flex items-center justify-between mt-1.5">
          <span class="text-[11px] text-ink3">{{ content.length }}/500</span>
          <InkButton variant="primary" size="sm" @click="submit" :disabled="!content.trim()">寄出</InkButton>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, computed, onMounted } from 'vue'
import InkButton from './ui/InkButton.vue'
import { useVisitor } from '../composables/useVisitor'
import { useWS } from '../composables/useWS'
import { toast } from '../composables/useToast'
import { listComments, createComment } from '../api/comments'
import { rel } from '../utils/format'
import { dicebearUrl } from '../utils/avatar'

const props = defineProps({ noteId: { type: Number, required: true } })

const { account, avatarUrl: myAvatar, openLogin } = useVisitor()

const comments = ref([]) // 平铺 ASC（含回复与临时条目），树在 computed 里建
const total = ref(0)
const content = ref('')
const expanded = ref(false)
const replyTarget = ref(null) // { id, nickname }
let tempSeq = 0
const tempTimers = new Map()

// 根评论倒序渲染，回复按时间正序缩进挂根下；父不在列表的孤儿按根兜底
const tree = computed(() => {
  const byId = new Map()
  for (const c of comments.value) byId.set(c.id, { c, replies: [] })
  const roots = []
  for (const c of comments.value) {
    const node = byId.get(c.id)
    if (c.parent_id && byId.has(c.parent_id)) byId.get(c.parent_id).replies.push(node)
    else roots.push(node)
  }
  return roots.reverse()
})
const displayTree = computed(() => (expanded.value ? tree.value : tree.value.slice(0, 10)))

// 楼层号 = 平铺时间序（老楼永不漂移）
const floorOf = computed(() => {
  const m = new Map()
  comments.value.forEach((c, i) => m.set(c.id, i + 1))
  return m
})

// 被回复人昵称：新评论优先用后端带回的 reply_to（归根后 parent_id 指向根），旧数据回退按 parent 映射
const replyToOf = computed(() => {
  const m = new Map()
  for (const c of comments.value) if (c.reply_to) m.set(c.id, c.reply_to)
  return m
})
const nickOf = computed(() => {
  const m = new Map()
  for (const c of comments.value) m.set(c.id, c.nickname || '匿名')
  return m
})
function replyNick(c) {
  return replyToOf.value.get(c.id) || (c.parent_id ? nickOf.value.get(c.parent_id) : '')
}

function isAdmin(c) {
  return !!c.visitor_uuid && c.visitor_uuid.startsWith('admin_')
}
function commentAvatar(c) {
  return c.avatar_url || dicebearUrl(c.avatar_style || 'lorelei', c.visitor_uuid)
}

function startReply(c) {
  replyTarget.value = { id: c.id, nickname: c.nickname || '匿名' }
}

async function load() {
  try {
    const data = await listComments(props.noteId)
    comments.value = data.comments || []
    total.value = data.total || comments.value.length
  } catch {}
}

onMounted(() => {
  load()
  const { connect } = useWS(`/api/notes/${props.noteId}/ws`, {
    onMessage: msg => { if (msg.type === 'comment' && msg.comment) upsert({ ...msg.comment, reply_to: msg.reply_to || '' }) },
    // 断线重连成功：拉一次全量，补齐离线期间错过的评论
    onOpen: isRetry => { if (isRetry) load() },
  })
  connect()
})

// WS 到达 / HTTP 返回共用：替换匹配的临时条目，否则按 id 去重后追加
function upsert(real) {
  const i = comments.value.findIndex(c => c.temp && c.visitor_uuid === real.visitor_uuid && c.content === real.content)
  if (i >= 0) {
    const tempId = comments.value[i].id
    comments.value.splice(i, 1, real)
    clearTimeout(tempTimers.get(tempId))
    tempTimers.delete(tempId)
  } else if (!comments.value.some(c => c.id === real.id)) {
    comments.value.push(real)
    total.value++
  }
  expanded.value = true
}

// 乐观上屏：立即插入「墨迹未干」条目，WS 先到则原位落定；
// 2.5s 无广播（WS 掉线/丢消息）兜底重拉一次；HTTP 失败撤条、还输入
function submit() {
  if (!account.value) { openLogin(); return }
  const text = content.value.trim()
  if (!text) return
  const parentId = replyTarget.value?.id ?? null
  const restore = { text, target: replyTarget.value }

  const temp = {
    id: 'temp-' + ++tempSeq, temp: true,
    note_id: props.noteId, visitor_uuid: account.value.uuid, content: text, parent_id: parentId,
    created_at: new Date().toISOString(),
    nickname: account.value.nickname, avatar_style: account.value.avatar_style,
    avatar_url: account.value.avatar_url, signature: account.value.signature,
  }
  comments.value.push(temp)
  total.value++
  expanded.value = true
  content.value = ''
  replyTarget.value = null

  tempTimers.set(temp.id, setTimeout(() => {
    if (comments.value.some(c => c.id === temp.id)) load()
  }, 2500))

  createComment(props.noteId, {
    visitor_uuid: account.value.uuid,
    content: text,
    parent_id: parentId,
  }).then(data => {
    upsert({ ...data.comment, reply_to: data.reply_to || '' })
  }).catch(() => {
    const i = comments.value.findIndex(c => c.id === temp.id)
    if (i >= 0) comments.value.splice(i, 1)
    total.value--
    clearTimeout(tempTimers.get(temp.id))
    tempTimers.delete(temp.id)
    content.value = restore.text
    replyTarget.value = restore.target
    toast.error('发送失败，请重试')
  })
}
</script>
