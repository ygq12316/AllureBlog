<template>
  <!-- 用显式点击而非 label 包裹：label 内的 <button> 会吞掉点击，文件选择器永远弹不出来 -->
  <span class="inline-flex items-center cursor-pointer" @click="open()">
    <input type="file" class="hidden" :accept="accept" @change="onChange" ref="el" />
    <slot />
  </span>
</template>

<script setup>
import { ref } from 'vue'

defineProps({ accept: { type: String, default: '' } })
const emit = defineEmits(['file'])
const el = ref(null)

function open() { el.value && el.value.click() }
function onChange(e) {
  const f = e.target.files && e.target.files[0]
  if (f) emit('file', f)
  e.target.value = '' // 允许重复选择同一文件
}
defineExpose({ open })
</script>
