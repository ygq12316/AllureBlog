<template>
  <n-modal :show="show" :mask-closable="false" @update:show="emit('close')">
    <div class="setup-card">
      <h3 class="setup-title">{{ isEdit ? '编辑资料' : '👋 设定你的身份' }}</h3>

      <!-- 裁剪模式 -->
      <div v-if="cropMode" class="crop-section">
        <div class="crop-frame" ref="cropFrame"
          @mousedown="startDrag" @mousemove="onDrag" @mouseup="endDrag" @mouseleave="endDrag"
          @touchstart="startDrag" @touchmove="onDrag" @touchend="endDrag"
          @wheel.prevent="onWheel">
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
      <div v-else class="avatar-row">
        <div class="avatar-left">
          <img :src="previewAvatar" class="avatar-preview" />
          <n-upload :show-file-list="false" :custom-request="pickImage" accept="image/*" style="margin-top:8px">
            <n-button size="tiny" text>{{ form.avatar_url ? '更换头像' : '上传头像' }}</n-button>
          </n-upload>
        </div>
        <div class="avatar-styles">
          <span v-for="s in styles" :key="s" class="style-dot"
            :class="{ active: form.avatar_style === s && !form.avatar_url }"
            @click="form.avatar_style = s; form.avatar_url = ''">
            <img :src="dicebearUrl(s, visitor?.uuid || 'demo')" />
          </span>
        </div>
      </div>

      <n-input v-model:value="form.nickname" placeholder="昵称" maxlength="20" style="margin-bottom:12px" />
      <n-input v-model:value="form.signature" placeholder="个性签名（选填）" maxlength="50" style="margin-bottom:16px" />

      <n-button type="primary" block @click="save" :disabled="!form.nickname.trim()">开始使用</n-button>
    </div>
  </n-modal>
</template>

<script setup>
import { reactive, ref, onMounted, computed } from 'vue'
import { useVisitor } from '../composables/useVisitor'
import { useAvatarCrop } from '../composables/useAvatarCrop'
import { dicebearUrl } from '../utils/avatar'

const { visitor, avatarUrl, init, update, account, setAccount } = useVisitor()
const { cropMode, cropSrc, cropImgStyle, cropZoom, pickImage, onWheel, startDrag, onDrag, endDrag, applyCrop, cancelCrop } = useAvatarCrop()

const emit = defineEmits(['close'])
const show = ref(true)
const styles = ['lorelei', 'thumbs', 'pixel-art', 'bottts', 'adventurer', 'identicon', 'initials', 'micah', 'avataaars', 'big-ears', 'big-smile', 'croodles', 'fun-emoji', 'shapes', 'notionists']

const form = reactive({
  nickname: '',
  avatar_style: 'lorelei',
  avatar_url: '',
  signature: '',
})

const isEdit = computed(() => !!visitor.value?.nickname && !visitor.value?.nickname.startsWith('访客'))

const previewAvatar = computed(() => form.avatar_url || dicebearUrl(form.avatar_style || 'lorelei', visitor.value?.uuid))

onMounted(async () => {
  await init()
  form.nickname = visitor.value?.nickname || ''
  form.avatar_style = visitor.value?.avatar_style || 'lorelei'
  form.avatar_url = visitor.value?.avatar_url || ''
  form.signature = visitor.value?.signature || ''
})

async function applyAvatar() {
  try {
    form.avatar_url = await applyCrop()
    form.avatar_style = ''
  } catch {}
}

async function save() {
  await update({
    nickname: form.nickname.trim(),
    avatar_style: form.avatar_style,
    avatar_url: form.avatar_url,
    signature: form.signature.trim(),
  })
  // 同步到登录账号
  if (account.value) {
    setAccount({ ...account.value, nickname: form.nickname.trim(), avatar_url: form.avatar_url, avatar_style: form.avatar_style })
  }
  show.value = false
  emit('close')
}
</script>

<style scoped>
.crop-section { text-align: center; }
.crop-frame {
  width: 200px; height: 200px; margin: 0 auto 12px;
  border-radius: 50%; overflow: hidden; position: relative;
  border: 2px solid var(--gold); cursor: move;
}
.crop-img {
  position: absolute; top: 0; left: 0;
  user-select: none; pointer-events: none;
}
.crop-mask {
  position: absolute; inset: 0; border-radius: 50%;
  box-shadow: inset 0 0 0 999px rgba(0,0,0,.3);
  pointer-events: none;
}
.crop-zoom { display: flex; align-items: center; gap: 8px; justify-content: center; margin-bottom: 12px; }
.zoom-slider { width: 120px; accent-color: var(--gold); }
.zoom-label { font-size: 16px; color: var(--muted); font-weight: 700; }
.crop-actions { display: flex; gap: 8px; justify-content: center; }

.setup-card {
  background: var(--card);
  border: 1px solid var(--card-border);
  border-radius: 12px;
  padding: 32px 28px;
  max-width: 380px;
  margin: 0 auto;
  text-align: center;
}
.setup-title {
  font-size: 18px;
  font-weight: 700;
  color: var(--text);
  margin: 0 0 20px;
}
.avatar-row {
  display: flex;
  align-items: flex-start;
  gap: 16px;
  margin-bottom: 20px;
  justify-content: center;
}
.avatar-left {
  display: flex; flex-direction: column; align-items: center;
  flex-shrink: 0;
}
.avatar-preview {
  width: 60px; height: 60px;
  border-radius: 50%;
  border: 2px solid var(--gold);
  flex-shrink: 0;
  object-fit: cover;
}
.avatar-styles {
  display: flex;
  flex-wrap: wrap;
  gap: 6px;
  max-height: 140px;
  overflow-y: auto;
  justify-content: center;
}
.style-dot {
  width: 32px;
  height: 32px;
  border-radius: 50%;
  border: 2px solid transparent;
  cursor: pointer;
  overflow: hidden;
  transition: border-color .2s;
}
.style-dot img {
  width: 100%;
  height: 100%;
  object-fit: cover;
}
.style-dot.active {
  border-color: var(--gold);
}
</style>
