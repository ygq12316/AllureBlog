<template>
  <div class="py-16 text-center text-sm font-light text-ink3 tracking-wide">
    <slot>{{ description }}</slot>
  </div>
</template>

<script setup>
defineProps({ description: { type: String, default: '暂无内容' } })
</script>
