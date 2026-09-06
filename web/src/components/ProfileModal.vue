<template>
  <InkModal :show="show" :mask-closable="false" width="380px" @update:show="emit('close')">
    <h3 class="font-light tracking-wide text-lg text-center m-0 mb-5">编辑资料</h3>

    <!-- 裁剪模式 -->
    <AvatarCropper v-if="cropping" :src="cropSrc" @done="onCropped" @cancel="cancelCrop" />

    <!-- 正常模式 -->
    <template v-else>
      <div class="flex items-start justify-center gap-4 mb-5">
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
            <img :src="dicebearUrl(s, account?.uuid || 'demo')" class="w-full h-full object-cover" alt="" />
          </span>
        </div>
      </div>

      <div class="flex flex-col gap-3">
        <InkInput v-model="form.nickname" placeholder="昵称" :maxlength="20" />
        <InkInput v-model="form.signature" placeholder="个性签名（选填）" :maxlength="50" />
        <InkButton variant="primary" block @click="save" :disabled="!form.nickname.trim()">保存</InkButton>
      </div>
    </template>
  </InkModal>
</template>

<script setup>
import { reactive, ref, onMounted, computed } from 'vue'
import InkModal from './ui/InkModal.vue'
import InkInput from './ui/InkInput.vue'
import InkButton from './ui/InkButton.vue'
import InkFilePicker from './ui/InkFilePicker.vue'
import AvatarCropper from './AvatarCropper.vue'
import { useVisitor, setAccount, closeProfile } from '../composables/useVisitor'
import { saveVisitor } from '../api/visitors'
import { dicebearUrl } from '../utils/avatar'

const { account } = useVisitor()

const emit = defineEmits(['close'])
const show = ref(true)
const cropping = ref(false)
const cropSrc = ref('')
const styles = ['lorelei', 'thumbs', 'pixel-art', 'bottts', 'adventurer', 'identicon', 'initials', 'micah', 'avataaars', 'big-ears', 'big-smile', 'croodles', 'fun-emoji', 'shapes', 'notionists']

const form = reactive({
  nickname: '',
  avatar_style: 'lorelei',
  avatar_url: '',
  signature: '',
})

const previewAvatar = computed(() => form.avatar_url || dicebearUrl(form.avatar_style || 'lorelei', account.value?.uuid))

onMounted(() => {
  form.nickname = account.value?.nickname || ''
  form.avatar_style = account.value?.avatar_style || 'lorelei'
  form.avatar_url = account.value?.avatar_url || ''
  form.signature = account.value?.signature || ''
})

function pickImage(file) {
  cropSrc.value = URL.createObjectURL(file)
  cropping.value = true
}
function onCropped(url) {
  form.avatar_url = url
  form.avatar_style = ''
  URL.revokeObjectURL(cropSrc.value)
  cropping.value = false
}
function cancelCrop() {
  URL.revokeObjectURL(cropSrc.value)
  cropping.value = false
}

async function save() {
  if (!account.value) return
  const fields = {
    nickname: form.nickname.trim(),
    avatar_style: form.avatar_style,
    avatar_url: form.avatar_url,
    signature: form.signature.trim(),
  }
  try {
    await saveVisitor({ uuid: account.value.uuid, ...fields })
  } catch {}
  setAccount({ ...account.value, ...fields })
  show.value = false
  closeProfile()
  emit('close')
}
</script>
