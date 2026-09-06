<template>
  <Teleport to="body">
    <Transition name="ink-modal">
      <div v-if="show" class="fixed inset-0 z-[950] flex items-center justify-center p-4 bg-[rgba(30,28,25,0.45)]"
        @click.self="maskClose">
        <div class="bg-paper border border-line max-h-[85vh] overflow-y-auto w-full max-w-[calc(100vw-2rem)]"
          :style="{ maxWidth: width }" role="dialog" aria-modal="true">
          <div v-if="title || $slots.header"
            class="flex items-center justify-between gap-3 px-8 pt-7 pb-4 border-b border-line2">
            <slot name="header">
              <h3 class="font-light tracking-wide text-lg m-0">{{ title }}</h3>
            </slot>
            <button v-if="closable" type="button" @click="close"
              class="bg-transparent border-0 p-0 cursor-pointer text-ink3 hover:text-ink transition-colors duration-700" aria-label="关闭">
              <CloseOutline class="w-4 h-4" />
            </button>
          </div>
          <div class="px-8 py-6">
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
  width: { type: String, default: '480px' },
  maskClosable: { type: Boolean, default: true },
  closable: { type: Boolean, default: true },
})
const emit = defineEmits(['update:show'])

function close() { emit('update:show', false) }
function maskClose() { if (props.maskClosable) close() }
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
.ink-modal-enter-active, .ink-modal-leave-active { transition: opacity 0.7s ease-in-out; }
.ink-modal-enter-from, .ink-modal-leave-to { opacity: 0; }
@media (prefers-reduced-motion: reduce) {
  .ink-modal-enter-active, .ink-modal-leave-active { transition: none; }
}
</style>
