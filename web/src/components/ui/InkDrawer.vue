<template>
  <Teleport to="body">
    <Transition name="ink-drawer">
      <div v-if="show" class="fixed inset-0 z-[960] bg-[rgba(30,28,25,0.45)]" @click.self="close">
        <div class="absolute top-0 right-0 h-full bg-paper border-l border-line overflow-y-auto flex flex-col"
          :style="{ width, maxWidth: props.maxWidth }" role="dialog" aria-modal="true">
          <div v-if="title" class="flex items-center justify-between px-8 pt-7 pb-4 border-b border-line2">
            <h3 class="font-light tracking-wide text-lg m-0">{{ title }}</h3>
            <button type="button" @click="close"
              class="bg-transparent border-0 p-0 cursor-pointer text-ink3 hover:text-ink transition-colors duration-700" aria-label="关闭">
              <CloseOutline class="w-5 h-5" />
            </button>
          </div>
          <div class="flex-1 overflow-y-auto">
            <slot />
          </div>
        </div>
      </div>
    </Transition>
  </Teleport>
</template>

<script setup>
import { onBeforeUnmount, onMounted, watch } from 'vue'
import { CloseOutline } from '@vicons/ionicons5'

const props = defineProps({
  show: Boolean,
  title: String,
  width: { type: String, default: '78vw' },
  maxWidth: { type: String, default: '85vw' },
})
const emit = defineEmits(['update:show'])

function close() { emit('update:show', false) }
function onKey(e) { if (e.key === 'Escape' && props.show) close() }

watch(() => props.show, v => {
  document.body.style.overflow = v ? 'hidden' : ''
})
onMounted(() => document.addEventListener('keydown', onKey))
onBeforeUnmount(() => {
  document.removeEventListener('keydown', onKey)
  document.body.style.overflow = ''
})
</script>

<style>
.ink-drawer-enter-active, .ink-drawer-leave-active { transition: opacity 0.7s ease-in-out; }
.ink-drawer-enter-active > div, .ink-drawer-leave-active > div { transition: transform 0.7s ease-in-out; }
.ink-drawer-enter-from, .ink-drawer-leave-to { opacity: 0; }
.ink-drawer-enter-from > div, .ink-drawer-leave-to > div { transform: translateX(100%); }
@media (prefers-reduced-motion: reduce) {
  .ink-drawer-enter-active, .ink-drawer-leave-active,
  .ink-drawer-enter-active > div, .ink-drawer-leave-active > div { transition: none; }
}
</style>
