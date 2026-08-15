<template>
  <div class="wrap">
    <!-- 极简顶栏 -->
    <div class="topbar">
      <n-button text size="small" @click="$router.push('/admin/articles')">
        <n-icon size="14" :component="ArrowBackIcon" /> 返回
      </n-button>
      <span class="word-count">{{ form.content.length }} 字</span>
      <div class="topbar-actions">
        <n-popconfirm v-if="isEdit" @positive-click="del" positive-text="确认删除" negative-text="取消">
          <template #trigger><n-button text size="small" type="error">删除</n-button></template>
          确定删除这篇文章？此操作不可恢复。
        </n-popconfirm>
        <n-button type="primary" @click="save" :disabled="!canPublish">发布</n-button>
      </div>
    </div>

    <!-- 标题区 -->
    <input
      v-model="form.title"
      class="title-input"
      placeholder="文章标题..."
      autofocus
    />

    <!-- 元信息 -->
    <div class="meta-row">
      <div class="meta-item">
        <span class="meta-label">分类</span>
        <n-select
          v-model:value="form.category"
          placeholder="选择分类"
          clearable
          :options="cats"
          size="small"
          class="meta-select"
        />
      </div>
      <div class="meta-item meta-tags">
        <span class="meta-label">标签</span>
        <n-select
          v-model:value="tags"
          placeholder="添加标签..."
          multiple
          filterable
          tag
          clearable
          :options="tagOptions"
          size="small"
          class="meta-select-tags"
        />
      </div>
    </div>

    <!-- 编辑区 -->
    <div class="editor">
      <div class="editor-pane">
        <div class="pane-label">MARKDOWN</div>
        <textarea
          v-model="form.content"
          class="editor-textarea"
          placeholder="开始写作..."
          spellcheck="false"
        />
      </div>
      <div class="editor-pane preview">
        <div class="pane-label">预览</div>
        <div v-if="form.content" v-html="preview" class="preview-body" />
        <div v-else class="preview-placeholder">右侧编辑区的内容将实时渲染于此</div>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, computed, watch, onMounted } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { ArrowBackOutline } from '@vicons/ionicons5'
import axios from 'axios'

const ArrowBackIcon = ArrowBackOutline
const route = useRoute()
const router = useRouter()
const categories = ref([])
const tagList = ref([])
const tags = ref([])
const form = ref({ title: '', content: '', category: '', tags: '', is_published: false })
const preview = ref('')

const isEdit = computed(() => !!route.params.id)
const cats = computed(() => categories.value.map(c => ({ label: c.name, value: c.name })))
const canPublish = computed(() => form.value.title.trim() && form.value.content.trim())
const tagOptions = computed(() => tagList.value.map(t => ({ label: t.name, value: t.name })))

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
  const [catRes, tagRes] = await Promise.all([
    axios.get('/api/categories'),
    axios.get('/api/tags')
  ])
  categories.value = catRes.data.categories || []
  tagList.value = tagRes.data.tags || []

  const id = route.params.id
  if (id) {
    const { data } = await axios.get(`/api/articles/${id}`)
    Object.assign(form.value, data)
    // 回填 tags 数组
    if (data.tags) tags.value = data.tags.split(',').filter(Boolean)
  }
})

async function save() {
  form.value.is_published = true
  const id = route.params.id
  if (id) await axios.put(`/api/articles/${id}`, form.value)
  else await axios.post('/api/articles', form.value)
  router.push('/admin/articles')
}

async function del() {
  const id = route.params.id
  if (id) {
    await axios.delete(`/api/articles/${id}`)
    router.push('/admin/articles')
  }
}
</script>

<style scoped>
.wrap { max-width: 1100px; margin: 0 auto; }

/* 顶栏 */
.topbar {
  display: flex;
  align-items: center;
  justify-content: space-between;
  margin-bottom: 28px;
}
.word-count {
  font-size: 12px;
  color: var(--muted);
  font-family: 'JetBrains Mono', monospace;
}
.topbar-actions {
  display: flex;
  align-items: center;
  gap: 10px;
}

/* 标题 */
.title-input {
  width: 100%;
  border: none;
  background: transparent;
  font-family: 'LXGW WenKai', serif;
  font-size: 28px;
  font-weight: 700;
  color: var(--text);
  padding: 0 0 16px 0;
  outline: none;
  caret-color: var(--gold);
}
.title-input::placeholder {
  color: var(--muted);
  opacity: 0.5;
}

/* 元信息行 */
.meta-row {
  display: flex;
  align-items: center;
  gap: 20px;
  padding-bottom: 14px;
  border-bottom: 1px solid var(--card-border);
  margin-bottom: 0;
}
.meta-item {
  display: flex;
  align-items: center;
  gap: 8px;
}
.meta-label {
  font-size: 12px;
  color: var(--muted);
  white-space: nowrap;
  letter-spacing: 0.5px;
}
.meta-select {
  width: 140px;
}
.meta-tags {
  flex: 1;
  min-width: 200px;
}
.meta-select-tags {
  flex: 1;
  max-width: 480px;
}

/* 编辑&预览区 */
.editor {
  display: flex;
  gap: 0;
  min-height: 520px;
  border: 1px solid var(--card-border);
  border-radius: 4px;
  overflow: hidden;
  margin-top: 12px;
}
.editor-pane {
  flex: 1;
  padding: 16px 20px;
  display: flex;
  flex-direction: column;
}
.editor-pane:first-child {
  border-right: 1px solid var(--card-border);
}
.preview {
  background: var(--card);
}
.pane-label {
  font-size: 10px;
  color: var(--gold);
  letter-spacing: 2px;
  margin-bottom: 10px;
  flex-shrink: 0;
}
.editor-textarea {
  flex: 1;
  width: 100%;
  border: none;
  background: transparent;
  resize: none;
  outline: none;
  font-family: 'JetBrains Mono', monospace;
  font-size: 13px;
  line-height: 1.8;
  color: var(--text);
  caret-color: var(--gold);
}
.editor-textarea::placeholder {
  color: var(--muted);
  opacity: 0.4;
}
.preview-body {
  flex: 1;
  font-family: 'LXGW WenKai', serif;
  font-size: 14px;
  line-height: 1.9;
  color: var(--text);
  overflow-y: auto;
}
.preview-body :deep(h1) { font-size: 22px; margin: 0 0 12px; }
.preview-body :deep(h2) { font-size: 18px; margin: 16px 0 8px; color: var(--gold); }
.preview-body :deep(h3) { font-size: 15px; margin: 12px 0 6px; }
.preview-body :deep(p) { margin: 0 0 10px; }
.preview-body :deep(strong) { color: var(--gold); }
.preview-body :deep(code) {
  background: var(--tag-bg);
  color: var(--gold);
  padding: 1px 6px;
  border-radius: 3px;
  font-size: 12px;
  font-family: 'JetBrains Mono', monospace;
}
.preview-placeholder {
  flex: 1;
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 13px;
  color: var(--muted);
  opacity: 0.5;
  font-family: 'LXGW WenKai', serif;
}
</style>
