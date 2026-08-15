<template>
  <div class="wrap">
    <h3 class="title">评论管理</h3>
    <div v-if="!list.length" class="empty">暂无评论</div>
    <div v-for="c in list" :key="c.id" class="row">
      <img :src="avt(c)" class="avt" />
      <div class="body">
        <span class="nick">{{ c.nickname || '匿名' }}</span>
        <span class="time">{{ c.created_at?.slice(0,16) }}</span>
        <p class="content">{{ c.content }}</p>
      </div>
      <n-popconfirm @positive-click="del(c.id)">
        <template #trigger><n-button size="tiny" text type="error">删除</n-button></template>
        确定删除？
      </n-popconfirm>
    </div>
  </div>
</template>

<script setup>
import { ref, onMounted } from 'vue'
import axios from 'axios'
const list = ref([])

async function loadComments() {
  try {
    let all = []
    const res1 = await axios.get('/api/articles?all=true&per_page=100')
    const aids = (res1.data.articles || []).map(a => a.id)
    for (const id of aids) {
      try {
        const r = await axios.get('/api/notes/' + id + '/comments')
        if (r.data.comments.length) {
          all = all.concat(r.data.comments)
        }
      } catch (e) {}
    }
    const res2 = await axios.get('/api/notes?all=true&per_page=100')
    for (const n of (res2.data.notes || [])) {
      try {
        const r = await axios.get('/api/notes/' + n.id + '/comments')
        if (r.data.comments.length) {
          all = all.concat(r.data.comments)
        }
      } catch (e) {}
    }
    all.sort((a, b) => new Date(b.created_at) - new Date(a.created_at))
    list.value = all
  } catch (e) {}
}

onMounted(loadComments)

async function del(id) {
  try {
    await axios.delete('/api/admin/comments/' + id)
  } catch (e) {}
  list.value = list.value.filter(c => c.id !== id)
}

function avt(c) {
  if (c.avatar_url) return c.avatar_url
  return 'https://api.dicebear.com/9.x/' + (c.avatar_style || 'lorelei') + '/svg?seed=' + c.visitor_uuid
}
</script>

<style scoped>
.wrap { max-width: 640px; margin: 0 auto; }
.title { font-size: 17px; font-weight: 700; color: var(--text); margin: 0 0 20px; }
.empty { padding: 40px; text-align: center; color: var(--muted); font-size: 13px; }
.row { display: flex; align-items: flex-start; gap: 10px; padding: 12px 0; border-bottom: 1px solid var(--card-border); }
.avt { width: 28px; height: 28px; border-radius: 50%; flex-shrink: 0; background: var(--tag-bg); }
.body { flex: 1; min-width: 0; }
.nick { font-size: 13px; font-weight: 600; color: var(--gold); margin-right: 8px; }
.time { font-size: 10px; color: var(--muted); }
.content { font-size: 13px; color: var(--text); margin: 4px 0 0; word-break: break-word; }
</style>