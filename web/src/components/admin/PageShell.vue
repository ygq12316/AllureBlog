<template>
  <div class="wrap" :class="widthClass">
    <div class="page-head">
      <h2>{{ title }}</h2>
      <div class="page-head-actions"><slot name="actions" /></div>
    </div>
    <slot />
  </div>
</template>

<script setup>
import { computed } from 'vue'

// 后台页骨架:页头(标题+主操作)与页面宽度的唯一来源,
// 页头/面板规范(index.html 内联 CSS)由本组件而非各视图照抄。
// width: narrow(720px)| wide(960px)| full(不限)
const props = defineProps({
  title: { type: String, required: true },
  width: { type: String, default: 'narrow' },
})

const widthClass = computed(() => (props.width === 'full' ? '' : `page-${props.width}`))
</script>
