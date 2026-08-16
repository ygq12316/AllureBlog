<template>
  <div class="wrap page-narrow">
    <div class="page-head">
      <h2>弹幕管理</h2>
      <div class="page-head-actions" />
    </div>
    <div class="panel">
      <n-data-table :columns="cols" :data="list" :bordered="false" size="small" :row-key="r => r.id" />
    </div>
  </div>
</template>

<script setup>
import { ref, onMounted, h } from 'vue'
import { NButton, NPopconfirm } from 'naive-ui'
import axios from 'axios'
const list = ref([])

const cols = [
  { title: '昵称', key: 'nickname', width: 110, render(row) { return row.nickname || '匿名' } },
  { title: '内容', key: 'content', width: '*', ellipsis: { tooltip: true } },
  { title: '颜色', key: 'color', width: 60, render(row) { return h('span', { style: `display:inline-block;width:12px;height:12px;border-radius:50%;background:${row.color};vertical-align:middle` }) } },
  { title: '时间', key: 'created_at', width: 140, render(row) { return row.created_at?.slice(0, 16) } },
  {
    title: '', width: 60,
    render(row) {
      return h(NPopconfirm, { onPositiveClick: () => del(row.id) }, {
        trigger: () => h(NButton, { size: 'tiny', text: true, type: 'error' }, { default: () => '删除' }),
        default: () => '确定删除？',
      })
    },
  },
]

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
