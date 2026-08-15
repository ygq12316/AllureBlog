<template>
  <n-switch :value="isDark" @update:value="toggle" size="small" class="theme-switch">
    <template #checked-icon><n-icon :component="Moon" /></template>
    <template #unchecked-icon><n-icon :component="Sunny" /></template>
  </n-switch>
</template>

<script setup>
import { ref, onMounted } from 'vue'
import { MoonOutline, SunnyOutline } from '@vicons/ionicons5'

const Moon = MoonOutline, Sunny = SunnyOutline
const isDark = ref(false)

onMounted(() => {
  const s = localStorage.getItem('theme')
  if (s === 'dark' || (!s && matchMedia('(prefers-color-scheme:dark)').matches)) {
    isDark.value = true; document.documentElement.classList.add('dark')
  }
})

function toggle(v) { isDark.value = v; document.documentElement.classList.toggle('dark',v); localStorage.setItem('theme',v?'dark':'light') }
</script>

<style scoped>
.theme-switch {
  --n-rail-color: var(--border);
  --n-rail-color-active: #b8944c;
  --n-button-color: var(--bg);
  --n-button-box-shadow: 0 1px 3px rgba(0,0,0,0.15);
}
</style>
