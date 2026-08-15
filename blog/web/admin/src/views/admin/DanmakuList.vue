<template>
  <div class="wrap">
    <h3 class="title">弹幕管理</h3>
    <div v-if="!list.length" class="empty">暂无弹幕</div>
    <div v-for="d in list" :key="d.id" class="row">
      <div class="body">
        <span class="nick">{{ d.nickname || '匿名' }}</span>
        <span class="time">{{ d.created_at?.slice(0,16) }}</span>
        <span class="dot" :style="{background:d.color}" />
        <span class="content">{{ d.content }}</span>
      </div>
      <n-popconfirm @positive-click="del(d.id)">
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

async function load() {
  try {
    const resp = await axios.get('/api/danmaku')
    list.value = resp.data.danmaku || []
  } catch (e) {}
}

onMounted(load)

async function del(id) {
  await axios.delete('/api/admin/danmaku/' + id)
  list.value = list.value.filter(d => d.id !== id)
}
</script>

<style scoped>
.wrap { max-width: 640px; margin: 0 auto; }
.title { font-size: 17px; font-weight: 700; color: var(--text); margin: 0 0 20px; }
.empty { padding: 40px; text-align: center; color: var(--muted); font-size: 13px; }
.row { display: flex; align-items: center; gap: 8px; padding: 10px 0; border-bottom: 1px solid var(--card-border); }
.body { flex: 1; min-width: 0; display: flex; align-items: center; gap: 8px; flex-wrap: wrap; }
.nick { font-size: 12px; font-weight: 600; color: var(--gold); }
.time { font-size: 10px; color: var(--muted); }
.dot { width: 6px; height: 6px; border-radius: 50%; flex-shrink: 0; }
.content { font-size: 13px; color: var(--text); }
</style>