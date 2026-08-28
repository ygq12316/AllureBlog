<template>
  <PageShell title="博客设置" width="narrow">
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
        <n-button size="tiny" @click="cancelCrop">取消</n-button>
        <n-button size="tiny" type="primary" @click="applyAvatar">裁剪</n-button>
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
  </PageShell>
</template>

<script setup>
import { ref, reactive, onMounted, computed } from 'vue'
import { updateConfig } from '../../api/config'
import { saveVisitor } from '../../api/visitors'
import PageShell from '../../components/admin/PageShell.vue'
import { useAvatarCrop } from '../../composables/useAvatarCrop'
import { useAuthor } from '../../composables/useAuthor'

const { author, load, refresh } = useAuthor()
const { cropMode, cropSrc, cropImgStyle, cropZoom, pickImage, onWheel, startDrag, onDrag, endDrag, applyCrop, cancelCrop } = useAvatarCrop()

const form = reactive({ author_name: '', author_avatar: '' })
const saving = ref(false)
const msg = ref('')

const previewAvatar = computed(() => form.author_avatar || 'data:image/svg+xml,<svg xmlns="http://www.w3.org/2000/svg" width="60" height="60"><rect fill="%23ddd" width="60" height="60"/><text x="30" y="35" text-anchor="middle" fill="%23999" font-size="24">?</text></svg>')

onMounted(async () => {
  await load()
  form.author_name = author.value.name
  form.author_avatar = author.value.avatar
})

async function applyAvatar() {
  try {
    form.author_avatar = await applyCrop()
  } catch (e) {
    msg.value = '❌ ' + (e.response?.data?.error || '上传失败')
  }
}

async function save() {
  saving.value = true; msg.value = ''
  try {
    await updateConfig({ author_name: form.author_name, author_avatar: form.author_avatar })
    // 同步博主假访客(admin_admin)的头像与昵称,供前台签名展示
    if (form.author_avatar) {
      try {
        await saveVisitor({ uuid: 'admin_admin', nickname: form.author_name, avatar_url: form.author_avatar, avatar_style: '', signature: '' })
      } catch {}
    }
    await refresh() // 全站博主信息(昵称/头像/签名)即时更新
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
