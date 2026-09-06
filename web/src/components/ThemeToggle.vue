<template>
  <span class="inline-flex items-center gap-1.5">
    <SunnyIcon v-if="!isDark" class="w-3.5 h-3.5 text-accent-strong" />
    <MoonIcon v-else class="w-3.5 h-3.5 text-accent-strong" />
    <InkSwitch :model-value="isDark" @update:model-value="toggle" aria-label="切换明暗主题" />
  </span>
</template>

<script setup>
import { ref, onMounted } from 'vue'
import { MoonOutline, SunnyOutline } from '@vicons/ionicons5'
import InkSwitch from './ui/InkSwitch.vue'

const MoonIcon = MoonOutline, SunnyIcon = SunnyOutline
const isDark = ref(false)

onMounted(() => {
  const s = localStorage.getItem('theme')
  if (s === 'dark' || (!s && matchMedia('(prefers-color-scheme:dark)').matches)) {
    isDark.value = true; document.documentElement.classList.add('dark')
  }
})

function toggle(v) { isDark.value = v; document.documentElement.classList.toggle('dark', v); localStorage.setItem('theme', v ? 'dark' : 'light') }
</script>
