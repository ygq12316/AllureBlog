<template>
  <div class="wrap page-wide">
    <div class="page-head">
      <h2>文章管理</h2>
      <div class="page-head-actions">
        <n-button v-if="selected.length" type="error" size="small" @click="batchDel">删除选中 ({{ selected.length }})</n-button>
        <n-button type="primary" @click="$router.push('/admin/articles/new')">+ 写文章</n-button>
      </div>
    </div>
    <div class="panel">
      <n-data-table :columns="cols" :data="articles" :bordered="false" size="small"
        :row-key="r => r.id" @update:checked-row-keys="selected = $event" />
    </div>
  </div>
</template>
<script setup>
import { ref, onMounted, h } from 'vue'
import { useRouter } from 'vue-router'
import { NButton, NTag, useMessage } from 'naive-ui'
import axios from 'axios'

const router = useRouter()
const message = useMessage()
const articles = ref([]), selected = ref([])

const cols = [
  { type: 'selection', width: 40 },
  { title: '标题', key: 'title', width: '*', render(row) { return h('a', { style: 'color:var(--gold);cursor:pointer', onClick: () => router.push(`/admin/articles/${row.id}/edit`) }, row.title) } },
  { title: '分类', key: 'category', width: 80, render(row) { return row.category ? h(NTag, { size: 'tiny', bordered: false }, { default: () => row.category }) : '' } },
  { title: '状态', width: 70, render(row) { return h(NTag, { size: 'tiny', type: row.is_published ? 'success' : 'warning', bordered: false }, { default: () => row.is_published ? '已发布' : '草稿' }) } },
  { title: '日期', width: 110, render(row) { return new Date(row.created_at).toLocaleDateString('zh-CN') } },
  { title: '', width: 50, render(row) { return h(NButton, { size: 'tiny', onClick: () => router.push(`/admin/articles/${row.id}/edit`) }, { default: () => '编辑' }) } },
]

onMounted(async () => { const { data } = await axios.get('/api/articles?all=true'); articles.value = data.articles || [] })
async function batchDel() {
  try {
    for (const id of selected.value) await axios.delete(`/api/articles/${id}`)
    articles.value = articles.value.filter(a => !selected.value.includes(a.id))
    selected.value = []
  } catch (e) {
    message.error('批量删除失败')
  }
}
</script>
