<template>
  <span class="inline-flex items-center gap-1 border px-2 py-0.5 text-xs font-light transition-colors duration-700"
    :class="toneClass">
    <slot />
    <button v-if="closable" type="button" @click="$emit('close')"
      class="bg-transparent border-0 p-0 cursor-pointer opacity-60 hover:opacity-100 transition-opacity duration-700"
      aria-label="移除">
      <CloseOutline class="w-3 h-3" />
    </button>
  </span>
</template>

<script setup>
import { computed } from 'vue'
import { CloseOutline } from '@vicons/ionicons5'

const props = defineProps({
  tone: { type: String, default: 'default' }, // default | moss | sand | cinnabar
  closable: Boolean,
})
defineEmits(['close'])

const toneClass = computed(() => ({
  default: 'border-line text-ink2',
  moss: 'border-moss/40 text-moss-deep',
  sand: 'border-sand text-accent-strong',
  cinnabar: 'border-cinnabar/40 text-cinnabar',
}[props.tone]))
</script>
