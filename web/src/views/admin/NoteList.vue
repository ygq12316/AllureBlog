<template>
  <PageShell title="随笔管理" width="wide">
    <template #actions>
      <InkButton v-if="selected.length" variant="danger" size="sm" @click="batchDel">删除选中 ({{ selected.length }})</InkButton>
      <InkButton variant="primary" size="sm" @click="$router.push('/admin/notes/new')">+ 写随笔</InkButton>
    </template>
    <div class="panel">
      <InkTable :columns="cols" :data="notes" :row-key="r => r.id" selectable
        v-model:checked="selected">
        <template #cell-images="{ row }">
          <div class="flex gap-1" v-if="imgs(row).length">
            <img v-for="u in imgs(row).slice(0, 3)" :key="u" :src="u"
              class="w-9 h-9 object-cover border border-line" alt="" />
          </div>
        </template>
        <template #cell-comment_count="{ row }">
          <a class="cursor-pointer transition-colors duration-700 hover:text-ink"
            :class="row.comment_count > 0 ? 'text-accent-strong' : 'text-ink3'"
            @click="router.push(`/admin/notes/${row.id}/edit`)">{{ row.comment_count ?? 0 }}</a>
        </template>
        <template #cell-status="{ row }">
          <InkTag :tone="row.is_published ? 'moss' : 'default'">{{ row.is_published ? '已发布' : '草稿' }}</InkTag>
        </template>
        <template #actions="{ row }">
          <span class="flex gap-2">
            <InkButton variant="link" size="xs" @click="router.push(`/admin/notes/${row.id}/edit`)">编辑</InkButton>
            <InkPopconfirm text="确定删除?" @confirm="del(row.id)">
              <template #trigger><InkButton variant="danger" size="xs">删除</InkButton></template>
            </InkPopconfirm>
          </span>
        </template>
      </InkTable>
    </div>
  </PageShell>
</template>
<script setup>
import { ref, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { toast } from '../../composables/useToast'
import { listNotes, removeNote } from '../../api/notes'
import PageShell from '../../components/admin/PageShell.vue'

const router = useRouter()
const notes = ref([]), selected = ref([])

const cols = [
  { title: '内容', key: 'content', ellipsis: true, render(row) { const t = (row.content || '').replace(/\s+/g, ' ').trim(); return t || '（无文字）' } },
  { title: '图片', key: 'images', width: 120, slot: true },
  { title: '评论', key: 'comment_count', width: 70, slot: true },
  { title: '状态', key: 'status', width: 80, slot: true },
  { title: '日期', key: 'created_at', width: 110, render(row) { return new Date(row.created_at).toLocaleDateString('zh-CN') } },
]

onMounted(async () => { notes.value = (await listNotes({ all: 'true' })).notes || [] })
function imgs(row) { return row.images ? row.images.split(',').filter(Boolean) : [] }
async function del(id) {
  try {
    await removeNote(id)
    notes.value = notes.value.filter(n => n.id !== id)
  } catch (e) { toast.error('删除失败') }
}
async function batchDel() {
  try {
    for (const id of selected.value) await removeNote(id)
    notes.value = notes.value.filter(n => !selected.value.includes(n.id))
    selected.value = []
  } catch (e) {
    toast.error('批量删除失败')
  }
}
</script>
