<template>
  <div class="wrap">
    <!-- 极简顶栏 -->
    <div class="topbar">
      <n-button text size="small" @click="$router.push('/admin/notes')">
        <n-icon size="14" :component="ArrowBackIcon" /> 返回
      </n-button>
      <span class="count" :class="{ 'count-warn': content.length > 400 }">{{ content.length }}/500</span>
      <n-button type="primary" @click="publish" :disabled="!canPublish">发布</n-button>
    </div>

    <!-- 卡片 -->
    <div class="card">
      <div class="card-top">
        <div class="avatar"><n-icon size="16" :component="PersonIcon" /></div>
        <span class="author">{{ authorName }}</span>
      </div>
      <textarea
        v-model="content"
        class="content-area"
        placeholder="写点什么..."
        maxlength="500"
        rows="4"
        autofocus
      />
      <div v-if="images.length" class="previews">
        <div v-for="(img, i) in images" :key="i" class="pv">
          <img :src="img" />
          <n-button @click="remove(i)" circle size="tiny" class="pv-rm">✕</n-button>
        </div>
      </div>
      <div class="card-foot">
        <n-upload
          v-if="images.length < 9"
          :show-file-list="false"
          :custom-request="upload"
          accept="image/*"
        >
          <n-button size="small" text>+ 添加图片</n-button>
        </n-upload>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, computed, onMounted } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { PersonOutline, ArrowBackOutline } from '@vicons/ionicons5'
import axios from 'axios'

const PersonIcon = PersonOutline
const ArrowBackIcon = ArrowBackOutline
const route = useRoute()
const router = useRouter()
const content = ref('')
const images = ref([])
const authorName = ref('Allure')
const canPublish = computed(() => content.value.trim().length > 0 || images.value.length > 0)

function remove(i) { images.value.splice(i, 1) }

async function upload({ file }) {
  const f = new FormData()
  f.append('file', file.file)
  const { data } = await axios.post('/api/upload', f)
  images.value.push(data.url)
}

async function publish() {
  const id = route.params.id
  const payload = { content: content.value, images: images.value.join(','), is_published: true }
  if (id) await axios.put(`/api/notes/${id}`, payload)
  else await axios.post('/api/notes', payload)
  router.push('/admin/notes')
}

onMounted(async () => {
  try{const{data}=await axios.get('/api/config');authorName.value=data.config?.author_name||'Allure'}catch(e){}
  const id = route.params.id
  if (!id) return
  try {
    const { data } = await axios.get(`/api/notes/${id}`)
    content.value = data.content || ''
    images.value = data.images ? data.images.split(',').filter(Boolean) : []
  } catch (e) {}
})
</script>

<style scoped>
.wrap { max-width: 540px; margin: 0 auto; }

/* 顶栏 */
.topbar {
  display: flex;
  align-items: center;
  justify-content: space-between;
  margin-bottom: 20px;
}
.count {
  font-size: 13px;
  color: var(--muted);
  font-family: 'JetBrains Mono', monospace;
}
.count-warn { color: #c97a4a; }

/* 卡片 */
.card {
  border: 1px solid var(--card-border);
  border-radius: 6px;
  padding: 20px;
  background: var(--card);
}
.card-top {
  display: flex;
  align-items: center;
  gap: 10px;
  margin-bottom: 14px;
}
.avatar {
  width: 34px;
  height: 34px;
  border-radius: 6px;
  background: var(--tag-bg);
  border: 1px solid var(--card-border);
  display: flex;
  align-items: center;
  justify-content: center;
  flex-shrink: 0;
  color: var(--gold);
}
.author {
  font-size: 14px;
  font-weight: 600;
  color: var(--text);
}

/* 内容 */
.content-area {
  width: 100%;
  border: none;
  background: transparent;
  resize: none;
  outline: none;
  font-family: 'LXGW WenKai', serif;
  font-size: 15px;
  line-height: 1.8;
  color: var(--text);
  min-height: 120px;
  caret-color: var(--gold);
}
.content-area::placeholder {
  color: var(--muted);
  opacity: 0.45;
}

/* 图片预览 */
.previews {
  display: flex;
  gap: 8px;
  flex-wrap: wrap;
  margin-top: 14px;
}
.pv {
  position: relative;
  width: 72px;
  height: 72px;
  border-radius: 6px;
  overflow: hidden;
  border: 1px solid var(--card-border);
}
.pv img {
  width: 100%;
  height: 100%;
  object-fit: cover;
}
.pv-rm {
  position: absolute;
  top: 2px;
  right: 2px;
}

/* 底部 */
.card-foot {
  margin-top: 14px;
  padding-top: 12px;
  border-top: 1px solid var(--card-border);
}
</style>
