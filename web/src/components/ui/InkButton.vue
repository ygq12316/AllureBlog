<template>
  <button :type="type" :disabled="disabled || loading" @click="$emit('click', $event)"
    class="bg-transparent cursor-pointer font-light tracking-wide transition-all duration-700 ease-in-out disabled:opacity-40 disabled:cursor-not-allowed"
    :class="[variantClass, sizeClass, block ? 'w-full' : '']">
    <span v-if="loading" class="ink-loading mr-1.5 inline-block">···</span>
    <slot />
  </button>
</template>

<script setup>
import { computed } from 'vue'

const props = defineProps({
  variant: { type: String, default: 'ghost' }, // primary | ghost | danger | link
  size: { type: String, default: 'md' },       // md | sm | xs
  type: { type: String, default: 'button' },
  block: Boolean,
  loading: Boolean,
  disabled: Boolean,
})
defineEmits(['click'])

const variantClass = computed(() => ({
  primary: 'bg-ink text-paper border border-ink hover:bg-transparent hover:text-ink',
  ghost: 'border-b border-ink/30 text-ink hover:text-accent-strong hover:border-accent',
  danger: 'border-b border-cinnabar/40 text-cinnabar hover:border-cinnabar',
  link: 'text-accent-strong hover:text-ink',
}[props.variant]))

const sizeClass = computed(() => ({
  md: 'px-5 py-2 text-sm',
  sm: 'px-3 py-1.5 text-xs',
  xs: 'px-2 py-0.5 text-xs',
}[props.size]))
</script>

<style scoped>
.ink-loading { animation: ink-breathe 1.2s ease-in-out infinite; }
@keyframes ink-breathe { 0%, 100% { opacity: 0.3; } 50% { opacity: 0.9; } }
@media (prefers-reduced-motion: reduce) { .ink-loading { animation: none; } }
</style>
