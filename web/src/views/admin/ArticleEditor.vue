<template>
  <div class="max-w-[1100px] mx-auto">
    <!-- 顶栏 -->
    <div class="editor-topbar">
      <InkButton variant="link" size="sm" @click="$router.push('/admin/articles')">
        <span class="inline-flex items-center gap-1"><ArrowBackIcon class="w-3.5 h-3.5" /> 返回</span>
      </InkButton>
      <span class="text-[13px] text-ink3 font-mono">{{ form.content.length }} 字</span>
      <div class="editor-topbar-side">
        <InkPopconfirm v-if="isEdit" text="确定删除这篇文章？此操作不可恢复。" @confirm="del">
          <template #trigger><InkButton variant="danger" size="sm">删除</InkButton></template>
        </InkPopconfirm>
        <InkButton variant="primary" @click="save" :disabled="!canPublish" :loading="saving">发布</InkButton>
      </div>
    </div>

    <!-- 标题区 -->
    <input v-model="form.title" placeholder="文章标题..." autofocus
      class="w-full border-0 bg-transparent text-[28px] font-light tracking-wide text-ink pb-4 outline-none caret-[var(--accent)] placeholder:text-ink3/50" />

    <!-- 元信息 -->
    <div class="flex flex-wrap items-center gap-5 pb-3.5 border-b border-line">
      <div class="flex items-center gap-2">
        <span class="text-xs text-ink3 whitespace-nowrap tracking-wider">分类</span>
        <span class="w-[140px] inline-block">
          <InkSelect v-model="form.category" placeholder="选择分类" clearable :options="cats" />
        </span>
      </div>
      <div class="flex items-center gap-2 flex-1 min-w-[200px]">
        <span class="text-xs text-ink3 whitespace-nowrap tracking-wider">标签</span>
        <span class="flex-1 max-w-[480px] inline-block">
          <TagInput v-model="tags" :suggestions="tagOptions" placeholder="添加标签..." />
        </span>
      </div>
    </div>

    <!-- 编辑区 -->
    <div class="flex min-h-[520px] border border-line mt-3 overflow-hidden">
      <div class="flex-1 p-4 px-5 flex flex-col border-r border-line">
        <div class="text-[10px] text-accent tracking-[2px] mb-2.5 shrink-0">MARKDOWN</div>
        <textarea v-model="form.content" placeholder="开始写作..." spellcheck="false"
          class="flex-1 w-full border-0 bg-transparent resize-none outline-none font-mono text-[13px] leading-[1.8] text-ink caret-[var(--accent)] placeholder:text-ink3/40" />
      </div>
      <div class="flex-1 p-4 px-5 flex flex-col bg-paper2">
        <div class="text-[10px] text-accent tracking-[2px] mb-2.5 shrink-0">预览</div>
        <div v-if="form.content" v-html="preview" class="preview-body" />
        <div v-else class="flex-1 flex items-center justify-center text-[13px] text-ink3/50">右侧编辑区的内容将实时渲染于此</div>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, computed, watch, onMounted } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { ArrowBackOutline } from '@vicons/ionicons5'
import { toast } from '../../composables/useToast'
import { listCategories } from '../../api/categories'
import { listTags } from '../../api/tags'
import { getArticle, createArticle, updateArticle, removeArticle } from '../../api/articles'

const ArrowBackIcon = ArrowBackOutline
const route = useRoute()
const router = useRouter()
const categories = ref([])
const tagList = ref([])
const tags = ref([])
const form = ref({ title: '', content: '', category: '', tags: '', is_published: false })
const preview = ref('')
const saving = ref(false)

const isEdit = computed(() => !!route.params.id)
const cats = computed(() => categories.value.map(c => ({ label: c.name, value: c.name })))
const canPublish = computed(() => form.value.title.trim() && form.value.content.trim())
const tagOptions = computed(() => tagList.value.map(t => t.name))

// 监听 tags 数组变化，同步到 form.tags 字符串
watch(tags, v => { form.value.tags = v.join(',') })
watch(() => form.value.tags, v => {
  if (!v && tags.value.length) tags.value = []
})

watch(() => form.value.content, v => {
  if (!v) { preview.value = ''; return }
  preview.value = v
    .replace(/^### (.+)$/gm, '<h3>$1</h3>')
    .replace(/^## (.+)$/gm, '<h2>$1</h2>')
    .replace(/^# (.+)$/gm, '<h1>$1</h1>')
    .replace(/\*\*(.+?)\*\*/g, '<strong>$1</strong>')
    .replace(/`(.+?)`/g, '<code>$1</code>')
    .replace(/\n\n/g, '</p><p>')
})

onMounted(async () => {
  const [cats, tagsAll] = await Promise.all([
    listCategories(),
    listTags()
  ])
  categories.value = cats
  tagList.value = tagsAll

  const id = route.params.id
  if (id) {
    const data = await getArticle(id)
    Object.assign(form.value, data)
    // 回填 tags 数组
    if (data.tags) tags.value = data.tags.split(',').filter(Boolean)
  }
})

async function save() {
  saving.value = true
  try {
    form.value.is_published = true
    const id = route.params.id
    if (id) await updateArticle(id, form.value)
    else await createArticle(form.value)
    router.push('/admin/articles')
  } catch (e) {
    toast.error(e.response?.data?.error || '发布失败')
    saving.value = false
  }
}

async function del() {
  const id = route.params.id
  if (!id) return
  try {
    await removeArticle(id)
    router.push('/admin/articles')
  } catch (e) {
    toast.error('删除失败')
  }
}
</script>

<style scoped>
.preview-body {
  flex: 1;
  font-size: 14px;
  line-height: 1.9;
  color: var(--ink);
  overflow-y: auto;
}
.preview-body :deep(h1) { font-size: 22px; font-weight: 400; margin: 0 0 12px; }
.preview-body :deep(h2) { font-size: 18px; font-weight: 400; margin: 16px 0 8px; color: #6b7b6e; }
.preview-body :deep(h3) { font-size: 15px; font-weight: 400; margin: 12px 0 6px; }
.preview-body :deep(p) { margin: 0 0 10px; }
.preview-body :deep(strong) { font-weight: 400; color: var(--accent-strong); border-bottom: 1px solid color-mix(in srgb, var(--accent) 40%, transparent); }
.preview-body :deep(code) {
  background: var(--paper);
  color: var(--accent-strong);
  padding: 1px 6px;
  font-size: 12px;
  font-family: 'JetBrains Mono', ui-monospace, monospace;
}
</style>
