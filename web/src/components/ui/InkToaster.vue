<template>
  <slot />
  <Teleport to="body">
    <div class="fixed top-6 left-1/2 -translate-x-1/2 z-[1000] flex flex-col items-center gap-2 pointer-events-none">
      <TransitionGroup name="ink-toast">
        <div v-for="t in items" :key="t.id"
          class="pointer-events-auto bg-paper2 border px-5 py-2.5 text-sm font-light tracking-wide"
          :class="t.type === 'error' ? 'text-cinnabar border-cinnabar/40' : t.type === 'success' ? 'text-moss-deep border-moss/40' : 'text-ink border-line'">
          {{ t.content }}
        </div>
      </TransitionGroup>
    </div>
  </Teleport>
</template>

<script setup>
import { useToast } from '../../composables/useToast'
const { items } = useToast()
</script>

<style>
.ink-toast-enter-active, .ink-toast-leave-active { transition: opacity 0.7s ease-in-out; }
.ink-toast-enter-from, .ink-toast-leave-to { opacity: 0; }
@media (prefers-reduced-motion: reduce) {
  .ink-toast-enter-active, .ink-toast-leave-active { transition: none; }
}
</style>
