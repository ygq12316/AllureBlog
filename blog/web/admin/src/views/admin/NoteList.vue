<template>
  <PageShell title="随笔管理" width="wide">
    <template #actions>
      <n-button v-if="selected.length" type="error" size="small" @click="batchDel">删除选中 ({{ selected.length }})</n-button>
      <n-button type="primary" @click="$router.push('/admin/notes/new')">+ 写随笔</n-button>
    </template>
    <div class="panel">
      <n-data-table :columns="cols" :data="notes" :bordered="false" size="small"
        :row-key="r => r.id" @update:checked-row-keys="selected = $event" />
    </div>
  </PageShell>
</template>
<script setup>
import { ref, onMounted, h } from 'vue'
import { useRouter } from 'vue-router'
import { NButton, NTag, NPopconfirm, useMessage } from 'naive-ui'
import { listNotes, removeNote } from '../../api/notes'
import PageShell from '../../components/admin/PageShell.vue'

const router = useRouter()
const message = useMessage()
const notes = ref([]), selected = ref([])

const cols = [
  { type: 'selection', width: 40 },
  {
    title: '内容', key: 'content', width: '*', ellipsis: { tooltip: true },
    render(row) { const t = (row.content || '').replace(/\s+/g, ' ').trim(); return t || '（无文字）' },
  },
  {
    title: '图片', key: 'images', width: 120,
    render(row) {
      const imgs = row.images ? row.images.split(',').filter(Boolean) : []
      if (!imgs.length) return ''
      return h('div', { style: 'display:flex;gap:4px' },
        imgs.slice(0, 3).map(u => h('img', { src: u, style: 'width:36px;height:36px;object-fit:cover;border-radius:4px;border:1px solid var(--card-border)' })))
    },
  },
  {
    title: '评论', key: 'comment_count', width: 70,
    render(row) {
      return h('a', {
        style: 'cursor:pointer;' + (row.comment_count > 0 ? 'color:var(--gold)' : 'color:var(--muted)'),
        onClick: () => router.push(`/admin/notes/${row.id}/edit`),
      }, String(row.comment_count ?? 0))
    },
  },
  { title: '状态', width: 70, render(row) { return h(NTag, { size: 'tiny', type: row.is_published ? 'success' : 'warning', bordered: false }, { default: () => row.is_published ? '已发布' : '草稿' }) } },
  { title: '日期', width: 110, render(row) { return new Date(row.created_at).toLocaleDateString('zh-CN') } },
  {
    title: '', width: 90,
    render(row) {
      return h('div', { style: 'display:flex;gap:2px' }, [
        h(NButton, { size: 'tiny', onClick: () => router.push(`/admin/notes/${row.id}/edit`) }, { default: () => '编辑' }),
        h(NPopconfirm, { onPositiveClick: () => del(row.id) }, {
          trigger: () => h(NButton, { size: 'tiny', text: true, type: 'error' }, { default: () => '删除' }),
          default: () => '确定删除?',
        }),
      ])
    },
  },
]

onMounted(async () => { notes.value = (await listNotes({ all: 'true' })).notes || [] })
async function del(id) {
  try {
    await removeNote(id)
    notes.value = notes.value.filter(n => n.id !== id)
  } catch (e) { message.error('删除失败') }
}
async function batchDel() {
  try {
    for (const id of selected.value) await removeNote(id)
    notes.value = notes.value.filter(n => !selected.value.includes(n.id))
    selected.value = []
  } catch (e) {
    message.error('批量删除失败')
  }
}
</script>
