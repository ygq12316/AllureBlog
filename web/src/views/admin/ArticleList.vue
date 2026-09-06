<template>
  <PageShell title="文章管理" width="wide">
    <template #actions>
      <InkButton v-if="selected.length" variant="danger" size="sm" @click="batchDel">删除选中 ({{ selected.length }})</InkButton>
      <InkButton variant="primary" size="sm" @click="$router.push('/admin/articles/new')">+ 写文章</InkButton>
    </template>
    <div class="panel">
      <InkTable :columns="cols" :data="articles" :row-key="r => r.id" selectable
        v-model:checked="selected">
        <template #cell-title="{ row }">
          <a class="cursor-pointer text-accent-strong transition-colors duration-700 hover:text-ink" @click="router.push(`/admin/articles/${row.id}/edit`)">{{ row.title }}</a>
        </template>
        <template #cell-category="{ row }">
          <InkTag v-if="row.category" tone="sand">{{ row.category }}</InkTag>
        </template>
        <template #cell-status="{ row }">
          <InkTag :tone="row.is_published ? 'moss' : 'default'">{{ row.is_published ? '已发布' : '草稿' }}</InkTag>
        </template>
        <template #actions="{ row }">
          <InkButton variant="link" size="xs" @click="router.push(`/admin/articles/${row.id}/edit`)">编辑</InkButton>
        </template>
      </InkTable>
    </div>
  </PageShell>
</template>
<script setup>
import { ref, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { toast } from '../../composables/useToast'
import { listArticles, removeArticle } from '../../api/articles'
import PageShell from '../../components/admin/PageShell.vue'

const router = useRouter()
const articles = ref([]), selected = ref([])

const cols = [
  { title: '标题', key: 'title', slot: true, ellipsis: true },
  { title: '分类', key: 'category', width: 90, slot: true },
  { title: '状态', key: 'status', width: 80, slot: true },
  { title: '日期', key: 'created_at', width: 110, render(row) { return new Date(row.created_at).toLocaleDateString('zh-CN') } },
]

onMounted(async () => { articles.value = (await listArticles({ all: 'true' })).articles || [] })
async function batchDel() {
  try {
    for (const id of selected.value) await removeArticle(id)
    articles.value = articles.value.filter(a => !selected.value.includes(a.id))
    selected.value = []
  } catch (e) {
    toast.error('批量删除失败')
  }
}
</script>
