<template>
  <div class="wrap">
    <!-- 顶栏：与文章编辑器同构 -->
    <div class="editor-topbar">
      <n-button text size="small" @click="$router.push('/admin/notes')">
        <n-icon size="14" :component="ArrowBackIcon" /> 返回
      </n-button>
      <span class="count" :class="{ 'count-warn': content.length > 400 }">{{ content.length }}/500</span>
      <div class="editor-topbar-side">
        <n-popconfirm v-if="isEdit" @positive-click="del" positive-text="确认删除" negative-text="取消">
          <template #trigger><n-button text size="small" type="error">删除</n-button></template>
          确定删除这条随笔？其下评论会一并失去归属。此操作不可恢复。
        </n-popconfirm>
        <n-button type="primary" @click="publish" :disabled="!canPublish" :loading="saving">发布</n-button>
      </div>
    </div>

    <div class="cols">
      <!-- 左栏：编辑卡（内容形态保留） -->
      <div class="card">
        <div class="card-top">
          <div class="avatar"><n-icon size="16" :component="PersonIcon" /></div>
          <span class="author">{{ authorName }}</span>
        </div>
        <textarea
          v-model="content"
          class="content-area"
          placeholder="写点什么..."
          maxlength="500"
          rows="8"
          autofocus
        />
        <div v-if="images.length" class="previews">
          <div v-for="(img, i) in images" :key="i" class="pv">
            <img :src="img" />
            <n-button @click="remove(i)" circle size="tiny" class="pv-rm">✕</n-button>
          </div>
        </div>
        <div class="card-foot">
          <n-upload
            v-if="images.length < 9"
            :show-file-list="false"
            :custom-request="upload"
            accept="image/*"
          >
            <n-button size="small" text>+ 添加图片</n-button>
          </n-upload>
        </div>
      </div>

      <!-- 右栏：评论面板（仅编辑态可用） -->
      <div class="panel comments-panel">
        <template v-if="isEdit">
          <div class="comments-head">
            <span class="comments-title">评论 · {{ comments.length }}</span>
            <span class="comments-sub">全部来自这条随笔</span>
          </div>
          <div class="comments-list">
            <div v-if="!comments.length" class="empty">还没有评论</div>
            <div v-for="c in comments" :key="c.id" class="comment-row">
              <img :src="avt(c)" class="c-avt" />
              <div class="c-body">
                <div class="c-meta">
                  <span class="c-nick">{{ c.nickname || '匿名' }}</span>
                  <span class="c-time">{{ c.created_at?.slice(0, 16) }}</span>
                </div>
                <p class="c-content">{{ c.content }}</p>
              </div>
              <n-popconfirm @positive-click="delComment(c.id)" positive-text="删除" negative-text="取消">
                <template #trigger><n-button size="tiny" text type="error">删除</n-button></template>
                确定删除这条评论？
              </n-popconfirm>
            </div>
          </div>
        </template>
        <div v-else class="empty comments-placeholder">发布随笔后<br />即可在此管理评论</div>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, computed, onMounted } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { PersonOutline, ArrowBackOutline } from '@vicons/ionicons5'
import { useMessage } from 'naive-ui'
import axios from 'axios'

const PersonIcon = PersonOutline
const ArrowBackIcon = ArrowBackOutline
const route = useRoute()
const router = useRouter()
const message = useMessage()
const content = ref('')
const images = ref([])
const authorName = ref('Allure')
const comments = ref([])
const saving = ref(false)

const isEdit = computed(() => !!route.params.id)
const canPublish = computed(() => content.value.trim().length > 0 || images.value.length > 0)

function remove(i) { images.value.splice(i, 1) }

function avt(c) {
  if (c.avatar_url) return c.avatar_url
  return 'https://api.dicebear.com/9.x/' + (c.avatar_style || 'lorelei') + '/svg?seed=' + c.visitor_uuid
}

async function upload({ file }) {
  try {
    const f = new FormData()
    f.append('file', file.file)
    const { data } = await axios.post('/api/upload', f)
    images.value.push(data.url)
  } catch (e) {
    message.error('图片上传失败')
  }
}

async function loadComments(id) {
  try {
    const { data } = await axios.get(`/api/notes/${id}/comments`)
    comments.value = data.comments || []
  } catch (e) {
    message.error('评论加载失败')
  }
}

async function publish() {
  saving.value = true
  try {
    const id = route.params.id
    const payload = { content: content.value, images: images.value.join(','), is_published: true }
    if (id) await axios.put(`/api/notes/${id}`, payload)
    else await axios.post('/api/notes', payload)
    router.push('/admin/notes')
  } catch (e) {
    message.error(e.response?.data?.error || '发布失败')
    saving.value = false
  }
}

async function del() {
  const id = route.params.id
  if (!id) return
  try {
    await axios.delete(`/api/notes/${id}`)
    router.push('/admin/notes')
  } catch (e) {
    message.error('删除失败')
  }
}

async function delComment(cid) {
  try {
    await axios.delete(`/api/admin/comments/${cid}`)
    comments.value = comments.value.filter(c => c.id !== cid)
  } catch (e) {
    message.error('评论删除失败')
  }
}

onMounted(async () => {
  try { const { data } = await axios.get('/api/config'); authorName.value = data.config?.author_name || 'Allure' } catch (e) {}
  const id = route.params.id
  if (!id) return
  try {
    const { data } = await axios.get(`/api/notes/${id}`)
    content.value = data.content || ''
    images.value = data.images ? data.images.split(',').filter(Boolean) : []
  } catch (e) {
    message.error('随笔加载失败')
  }
  loadComments(id)
})
</script>

<style scoped>
.wrap { max-width: 1100px; margin: 0 auto; }

.count {
  font-size: 13px;
  color: var(--muted);
  font-family: 'JetBrains Mono', monospace;
}
.count-warn { color: #c97a4a; }

/* 左右两栏 */
.cols { display: flex; gap: 20px; align-items: stretch; }
@media (max-width: 900px) { .cols { flex-direction: column; } }

/* 左栏：编辑卡 */
.card {
  flex: 1;
  border: 1px solid var(--card-border);
  border-radius: 8px;
  padding: 20px;
  background: var(--card);
  display: flex;
  flex-direction: column;
}
.card-top { display: flex; align-items: center; gap: 10px; margin-bottom: 14px; }
.avatar {
  width: 34px; height: 34px; border-radius: 6px;
  background: var(--tag-bg); border: 1px solid var(--card-border);
  display: flex; align-items: center; justify-content: center;
  flex-shrink: 0; color: var(--gold);
}
.author { font-size: 14px; font-weight: 600; color: var(--text); }
.content-area {
  width: 100%; border: none; background: transparent; resize: none; outline: none;
  font-family: 'LXGW WenKai', serif; font-size: 15px; line-height: 1.8;
  color: var(--text); flex: 1; min-height: 200px; caret-color: var(--gold);
}
.content-area::placeholder { color: var(--muted); opacity: 0.45; }
.previews { display: flex; gap: 8px; flex-wrap: wrap; margin-top: 14px; }
.pv { position: relative; width: 72px; height: 72px; border-radius: 6px; overflow: hidden; border: 1px solid var(--card-border); }
.pv img { width: 100%; height: 100%; object-fit: cover; }
.pv-rm { position: absolute; top: 2px; right: 2px; }
.card-foot { margin-top: 14px; padding-top: 12px; border-top: 1px solid var(--card-border); }

/* 右栏：评论面板 */
.comments-panel { width: 380px; flex-shrink: 0; display: flex; flex-direction: column; max-height: 640px; }
@media (max-width: 900px) { .comments-panel { width: auto; max-height: 360px; } }
.comments-head {
  display: flex; justify-content: space-between; align-items: baseline;
  padding-bottom: 12px; border-bottom: 1px solid var(--card-border); flex-shrink: 0;
}
.comments-title { font-size: 14px; font-weight: 700; color: var(--text); }
.comments-sub { font-size: 11px; color: var(--muted); }
.comments-list { flex: 1; overflow-y: auto; }
.comment-row { display: flex; align-items: flex-start; gap: 10px; padding: 12px 0; border-bottom: 1px solid var(--card-border); }
.comment-row:last-child { border-bottom: none; }
.c-avt { width: 28px; height: 28px; border-radius: 50%; flex-shrink: 0; background: var(--tag-bg); }
.c-body { flex: 1; min-width: 0; }
.c-meta { display: flex; align-items: baseline; gap: 8px; }
.c-nick { font-size: 13px; font-weight: 600; color: var(--gold); }
.c-time { font-size: 10px; color: var(--muted); }
.c-content { font-size: 13px; color: var(--text); margin: 4px 0 0; word-break: break-word; }
.empty { padding: 40px 0; text-align: center; color: var(--muted); font-size: 13px; }
.comments-placeholder { line-height: 2; flex: 1; display: flex; align-items: center; justify-content: center; padding: 0; }
</style>
