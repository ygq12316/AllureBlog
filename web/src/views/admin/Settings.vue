<template>
  <PageShell title="博客设置" width="narrow">
    <!-- 裁剪模式 -->
    <div v-if="cropMode" class="panel text-center">
      <div class="relative w-[200px] h-[200px] mx-auto mb-3 rounded-full overflow-hidden border border-accent cursor-move"
        @mousedown="startDrag" @mousemove="onDrag" @mouseup="endDrag" @mouseleave="endDrag"
        @touchstart="startDrag" @touchmove="onDrag" @touchend="endDrag" @wheel.prevent="onWheel">
        <img :src="cropSrc" class="absolute top-0 left-0 select-none pointer-events-none" :style="cropImgStyle" alt="裁剪预览" />
        <div class="absolute inset-0 rounded-full pointer-events-none" style="box-shadow: inset 0 0 0 999px rgba(0,0,0,.3)" />
      </div>
      <div class="flex items-center justify-center gap-2 mb-3">
        <span class="text-base text-ink3">−</span>
        <input type="range" min="0.5" max="3" step="0.05" v-model.number="cropZoom" class="w-[120px] accent-[var(--accent)]" aria-label="缩放" />
        <span class="text-base text-ink3">+</span>
      </div>
      <div class="flex justify-center gap-3">
        <InkButton size="xs" @click="cancelCrop">取消</InkButton>
        <InkButton size="xs" variant="primary" @click="applyAvatar">裁剪</InkButton>
      </div>
    </div>

    <!-- 正常模式 -->
    <div v-else class="panel flex flex-col gap-4">
      <div class="flex flex-col gap-1.5">
        <label class="text-xs text-ink3 tracking-wider">作者昵称</label>
        <InkInput v-model="form.author_name" placeholder="作者昵称" />
      </div>
      <div class="flex flex-col gap-1.5">
        <label class="text-xs text-ink3 tracking-wider">作者头像</label>
        <div class="flex items-center gap-3">
          <img :src="previewAvatar" class="w-[60px] h-[60px] rounded-full object-cover border border-accent" alt="头像预览" />
          <InkFilePicker accept="image/*" @file="pickImage">
            <InkButton variant="link" size="sm">{{ form.author_avatar ? '更换头像' : '上传头像' }}</InkButton>
          </InkFilePicker>
        </div>
      </div>
      <InkButton variant="primary" @click="save" :loading="saving">保存</InkButton>
      <p v-if="msg" class="text-xs text-center m-0 transition-colors duration-700"
        :class="msgOk ? 'text-moss-deep' : 'text-cinnabar'">{{ msg }}</p>
    </div>
  </PageShell>
</template>

<script setup>
import { ref, reactive, onMounted, computed } from 'vue'
import InkButton from '../../components/ui/InkButton.vue'
import InkInput from '../../components/ui/InkInput.vue'
import InkFilePicker from '../../components/ui/InkFilePicker.vue'
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
const msgOk = ref(true)

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
    msgOk.value = false
    msg.value = e.response?.data?.error || '上传失败'
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
    msgOk.value = true
    msg.value = '保存成功'
  } catch (e) {
    msgOk.value = false
    msg.value = e.response?.data?.error || '保存失败'
  }
  saving.value = false; setTimeout(() => msg.value = '', 3000)
}
</script>
