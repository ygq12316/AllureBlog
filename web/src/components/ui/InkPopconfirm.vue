<template>
  <span class="relative inline-block" ref="root">
    <span @click="open = !open"><slot name="trigger" /></span>
    <Transition name="ink-fade">
      <span v-if="open"
        class="absolute bottom-full mb-2 left-1/2 -translate-x-1/2 z-40 w-max max-w-[240px] bg-paper border border-line px-4 py-3 text-xs font-light text-ink2 shadow-none block">
        <slot>{{ text }}</slot>
        <span class="flex justify-end gap-3 mt-2">
          <button type="button" @click="open = false"
            class="bg-transparent border-0 p-0 cursor-pointer text-ink3 hover:text-ink transition-colors duration-700 text-xs font-light">取消</button>
          <button type="button" @click="confirm"
            class="bg-transparent border-0 p-0 cursor-pointer text-cinnabar hover:opacity-70 transition-opacity duration-700 text-xs font-light">确定</button>
        </span>
      </span>
    </Transition>
  </span>
</template>

<script setup>
import { onBeforeUnmount, onMounted, ref } from 'vue'

defineProps({ text: { type: String, default: '确定执行该操作？' } })
const emit = defineEmits(['confirm'])

const open = ref(false)
const root = ref(null)

function confirm() {
  open.value = false
  emit('confirm')
}
function onOutside(e) { if (root.value && !root.value.contains(e.target)) open.value = false }
onMounted(() => document.addEventListener('click', onOutside))
onBeforeUnmount(() => document.removeEventListener('click', onOutside))
</script>

<style>
.ink-fade-enter-active, .ink-fade-leave-active { transition: opacity 0.5s ease-in-out; }
.ink-fade-enter-from, .ink-fade-leave-to { opacity: 0; }
@media (prefers-reduced-motion: reduce) {
  .ink-fade-enter-active, .ink-fade-leave-active { transition: none; }
}
</style>
