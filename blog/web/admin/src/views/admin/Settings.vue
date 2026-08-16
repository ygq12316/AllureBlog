<template>
  <div class="wrap page-narrow">
    <div class="page-head">
      <h2>博客设置</h2>
      <div class="page-head-actions" />
    </div>

    <!-- 裁剪模式 -->
    <div v-if="cropMode" class="panel crop-section">
      <div class="crop-frame" @mousedown="startDrag" @mousemove="onDrag" @mouseup="endDrag" @mouseleave="endDrag"
        @touchstart="startDrag" @touchmove="onDrag" @touchend="endDrag" @wheel.prevent="onWheel">
        <img :src="cropSrc" class="crop-img" :style="cropImgStyle" />
        <div class="crop-mask" />
      </div>
      <div class="crop-zoom">
        <span class="zoom-label">−</span>
        <input type="range" min="0.5" max="3" step="0.05" v-model.number="cropZoom" class="zoom-slider" />
        <span class="zoom-label">+</span>
      </div>
      <div class="crop-actions">
        <n-button size="tiny" @click="cropMode = false">取消</n-button>
        <n-button size="tiny" type="primary" @click="applyCrop">裁剪</n-button>
      </div>
    </div>

    <!-- 正常模式 -->
    <div v-else class="panel form">
      <div class="field">
        <label>作者昵称</label>
        <n-input v-model:value="form.author_name" placeholder="作者昵称" size="small" />
      </div>
      <div class="field">
        <label>作者头像</label>
        <div class="avatar-row">
          <img :src="previewAvatar" class="preview" />
          <n-upload :show-file-list="false" :custom-request="pickImage" accept="image/*">
            <n-button size="small" text>{{ form.author_avatar ? '更换头像' : '上传头像' }}</n-button>
          </n-upload>
        </div>
      </div>
      <n-button type="primary" @click="save" :loading="saving">保存</n-button>
      <p v-if="msg" class="msg">{{ msg }}</p>
    </div>
  </div>
</template>

<script setup>
import { ref, reactive, onMounted, computed } from 'vue'
import axios from 'axios'

const form = reactive({ author_name: '', author_avatar: '' })
const saving = ref(false)
const msg = ref('')

const previewAvatar = computed(() => form.author_avatar || 'data:image/svg+xml,<svg xmlns="http://www.w3.org/2000/svg" width="60" height="60"><rect fill="%23ddd" width="60" height="60"/><text x="30" y="35" text-anchor="middle" fill="%23999" font-size="24">?</text></svg>')

onMounted(async () => {
  try {
    const { data } = await axios.get('/api/admin/config')
    form.author_name = data.config.author_name || 'Allure'
    form.author_avatar = data.config.author_avatar || ''
  } catch (e) { msg.value = '加载失败' }
})

// -- crop --
const cropMode = ref(false)
const cropSrc = ref('')
const cropImgStyle = ref({})
const cropZoom = ref(1)
let cropImg = null, dragging = false, dragStart = { x: 0, y: 0 }, imgPos = { x: 0, y: 0 }

async function pickImage({ file }) {
  cropSrc.value = URL.createObjectURL(file.file)
  cropImg = new Image(); cropImg.src = cropSrc.value
  await new Promise(r => cropImg.onload = r)
  imgPos = { x: 0, y: 0 }; cropZoom.value = 1; updateCropStyle(); cropMode.value = true
}

function updateCropStyle() {
  if (!cropImg) return
  const scale = (200 / Math.min(cropImg.width, cropImg.height)) * cropZoom.value
  cropImgStyle.value = { width: (cropImg.width * scale) + 'px', height: (cropImg.height * scale) + 'px', transform: `translate(${imgPos.x}px, ${imgPos.y}px)` }
}

function onWheel(e) { cropZoom.value = Math.max(0.5, Math.min(3, cropZoom.value + (e.deltaY > 0 ? -0.1 : 0.1))) }
function startDrag(e) { dragging = true; const ev = e.touches ? e.touches[0] : e; dragStart = { x: ev.clientX - imgPos.x, y: ev.clientY - imgPos.y } }
function onDrag(e) { if (!dragging) return; const ev = e.touches ? e.touches[0] : e; imgPos.x = ev.clientX - dragStart.x; imgPos.y = ev.clientY - dragStart.y; updateCropStyle() }
function endDrag() { dragging = false }

async function applyCrop() {
  if (!cropImg) return
  const size = 200, canvas = document.createElement('canvas')
  canvas.width = size; canvas.height = size
  const ctx = canvas.getContext('2d')
  ctx.beginPath(); ctx.arc(size / 2, size / 2, size / 2, 0, Math.PI * 2); ctx.clip()
  const scale = (size / Math.min(cropImg.width, cropImg.height)) * cropZoom.value
  const sx = -imgPos.x / scale, sy = -imgPos.y / scale, sw = size / scale
  ctx.drawImage(cropImg, sx, sy, sw, sw, 0, 0, size, size)
  const blob = await new Promise(r => canvas.toBlob(r, 'image/png'))
  const fd = new FormData(); fd.append('file', blob, 'avatar.png')
  const { data } = await axios.post('/api/upload', fd)
  form.author_avatar = data.url; cropMode.value = false; URL.revokeObjectURL(cropSrc.value)
}

async function save() {
  saving.value = true; msg.value = ''
  try {
    await axios.put('/api/admin/config', { author_name: form.author_name, author_avatar: form.author_avatar })
    // 同步到管理员访客头像
    if (form.author_avatar) {
      try {
        await axios.post('/api/visitor', { uuid: 'admin_admin', nickname: form.author_name, avatar_url: form.author_avatar, avatar_style: '', signature: '' })
        localStorage.setItem('blog_visitor', JSON.stringify({ uuid: 'admin_admin', nickname: form.author_name, avatar_url: form.author_avatar, avatar_style: '' }))
      } catch {}
    }
    msg.value = '✅ 保存成功'
  } catch (e) { msg.value = '❌ ' + (e.response?.data?.error || '保存失败') }
  saving.value = false; setTimeout(() => msg.value = '', 3000)
}
</script>

<style scoped>
.form { display: flex; flex-direction: column; gap: 16px; }
.field { display: flex; flex-direction: column; gap: 6px; }
.field label { font-size: 12px; color: var(--muted); }
.avatar-row { display: flex; align-items: center; gap: 12px; }
.preview { width: 60px; height: 60px; border-radius: 50%; object-fit: cover; border: 2px solid var(--gold); }
.msg { font-size: 12px; text-align: center; color: var(--text); margin: 0; }

.crop-section { text-align: center; }
.crop-frame { width: 200px; height: 200px; margin: 0 auto 12px; border-radius: 50%; overflow: hidden; position: relative; border: 2px solid var(--gold); cursor: move; }
.crop-img { position: absolute; top: 0; left: 0; user-select: none; pointer-events: none; }
.crop-mask { position: absolute; inset: 0; border-radius: 50%; box-shadow: inset 0 0 0 999px rgba(0,0,0,.3); pointer-events: none; }
.crop-zoom { display: flex; align-items: center; gap: 8px; justify-content: center; margin-bottom: 12px; }
.zoom-slider { width: 120px; accent-color: var(--gold); }
.zoom-label { font-size: 16px; color: var(--muted); font-weight: 700; }
.crop-actions { display: flex; gap: 8px; justify-content: center; }
</style>