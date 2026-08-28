<template>
  <PageShell title="弹幕管理">
    <div class="panel">
      <n-data-table :columns="cols" :data="list" :bordered="false" size="small" :row-key="r => r.id" />
    </div>
  </PageShell>
</template>

<script setup>
import { ref, onMounted, h } from 'vue'
import { NButton, NPopconfirm } from 'naive-ui'
import { listDanmaku, removeDanmaku } from '../../api/danmaku'
import PageShell from '../../components/admin/PageShell.vue'
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
    list.value = await listDanmaku()
  } catch (e) {}
}

onMounted(load)

async function del(id) {
  await removeDanmaku(id)
  list.value = list.value.filter(d => d.id !== id)
}
</script>
