<template>
  <label class="inline-flex items-center cursor-pointer">
    <input type="file" class="hidden" :accept="accept" @change="onChange" ref="el" />
    <slot />
  </label>
</template>

<script setup>
import { ref } from 'vue'

defineProps({ accept: { type: String, default: '' } })
const emit = defineEmits(['file'])
const el = ref(null)

function onChange(e) {
  const f = e.target.files && e.target.files[0]
  if (f) emit('file', f)
  e.target.value = '' // 允许重复选择同一文件
}
function open() { el.value && el.value.click() }
defineExpose({ open })
</script>
