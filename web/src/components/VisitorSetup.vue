<template>
  <InkModal :show="show" :mask-closable="false" width="380px" @update:show="emit('close')">
    <h3 class="font-light tracking-wide text-lg text-center m-0 mb-5">{{ isEdit ? '编辑资料' : '设定你的身份' }}</h3>

    <!-- 裁剪模式 -->
    <div v-if="cropMode" class="text-center">
      <div class="relative w-[200px] h-[200px] mx-auto mb-3 rounded-full overflow-hidden border border-accent cursor-move"
        ref="cropFrame"
        @mousedown="startDrag" @mousemove="onDrag" @mouseup="endDrag" @mouseleave="endDrag"
        @touchstart="startDrag" @touchmove="onDrag" @touchend="endDrag"
        @wheel.prevent="onWheel">
        <img :src="cropSrc" class="absolute top-0 left-0 select-none pointer-events-none" :style="cropImgStyle" alt="裁剪预览" />
        <div class="absolute inset-0 rounded-full pointer-events-none" style="box-shadow: inset 0 0 0 999px rgba(0,0,0,.3)" />
      </div>
      <div class="flex items-center justify-center gap-2 mb-3">
        <span class="text-base text-ink3">−</span>
        <input type="range" min="0.5" max="3" step="0.05" v-model.number="cropZoom" class="w-30 accent-[var(--accent)]" aria-label="缩放" />
        <span class="text-base text-ink3">+</span>
      </div>
      <div class="flex justify-center gap-3">
        <InkButton size="xs" @click="cancelCrop">取消</InkButton>
        <InkButton size="xs" variant="primary" @click="applyAvatar">裁剪</InkButton>
      </div>
    </div>

    <!-- 正常模式 -->
    <div v-else class="flex items-start justify-center gap-4 mb-5">
      <div class="flex flex-col items-center shrink-0">
        <img :src="previewAvatar" class="w-[60px] h-[60px] rounded-full border border-accent object-cover" alt="头像预览" />
        <InkFilePicker accept="image/*" class="mt-2" @file="pickImage">
          <InkButton variant="link" size="xs">{{ form.avatar_url ? '更换头像' : '上传头像' }}</InkButton>
        </InkFilePicker>
      </div>
      <div class="flex flex-wrap justify-center gap-1.5 max-h-[140px] overflow-y-auto">
        <span v-for="s in styles" :key="s"
          class="w-8 h-8 rounded-full overflow-hidden cursor-pointer border transition-colors duration-700"
          :class="form.avatar_style === s && !form.avatar_url ? 'border-accent' : 'border-transparent hover:border-line'"
          @click="form.avatar_style = s; form.avatar_url = ''">
          <img :src="dicebearUrl(s, visitor?.uuid || 'demo')" class="w-full h-full object-cover" alt="" />
        </span>
      </div>
    </div>

    <div class="flex flex-col gap-3">
      <InkInput v-model="form.nickname" placeholder="昵称" :maxlength="20" />
      <InkInput v-model="form.signature" placeholder="个性签名（选填）" :maxlength="50" />
      <InkButton variant="primary" block @click="save" :disabled="!form.nickname.trim()">开始使用</InkButton>
    </div>
  </InkModal>
</template>

<script setup>
import { reactive, ref, onMounted, computed } from 'vue'
import InkModal from './ui/InkModal.vue'
import InkInput from './ui/InkInput.vue'
import InkButton from './ui/InkButton.vue'
import InkFilePicker from './ui/InkFilePicker.vue'
import { useVisitor } from '../composables/useVisitor'
import { useAvatarCrop } from '../composables/useAvatarCrop'
import { dicebearUrl } from '../utils/avatar'

const { visitor, avatarUrl, init, update, account, setAccount } = useVisitor()
const { cropMode, cropSrc, cropImgStyle, cropZoom, pickImage, onWheel, startDrag, onDrag, endDrag, applyCrop, cancelCrop } = useAvatarCrop()

const emit = defineEmits(['close'])
const show = ref(true)
const cropFrame = ref(null)
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
