<template>
  <span class="relative inline-block" @mouseenter="show = true" @mouseleave="show = false">
    <slot />
    <Transition name="ink-fade">
      <div v-if="show"
        class="absolute right-0 top-full pt-2 z-40 min-w-[140px]">
        <ul class="bg-paper border border-line py-1 m-0 p-0 list-none" @mouseleave="show = false">
          <template v-for="o in options" :key="o.key ?? o.label">
            <li v-if="o.type === 'divider'" class="my-1 border-t border-line2" />
            <li v-else class="px-5 py-1.5 text-sm font-light cursor-pointer whitespace-nowrap transition-colors duration-500"
              :class="o.type === 'danger' ? 'text-cinnabar hover:bg-paper2' : 'text-ink2 hover:text-ink hover:bg-paper2'"
              @click="show = false; $emit('select', o.key)">{{ o.label }}</li>
          </template>
        </ul>
      </div>
    </Transition>
  </span>
</template>

<script setup>
import { ref } from 'vue'

defineProps({
  options: { type: Array, default: () => [] }, // [{label, key, type?: 'divider'|'danger'}]
  trigger: { type: String, default: 'hover' },
})
defineEmits(['select'])

const show = ref(false)
</script>

<style>
.ink-fade-enter-active, .ink-fade-leave-active { transition: opacity 0.5s ease-in-out; }
.ink-fade-enter-from, .ink-fade-leave-to { opacity: 0; }
@media (prefers-reduced-motion: reduce) {
  .ink-fade-enter-active, .ink-fade-leave-active { transition: none; }
}
</style>
