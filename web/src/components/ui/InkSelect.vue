<template>
  <span ref="root" class="relative inline-block w-full">
    <button type="button" @click="open = !open" @keydown.down.prevent="move(1)" @keydown.up.prevent="move(-1)"
      @keydown.enter.prevent="open && choose(hovered)" @keydown.esc="open = false"
      class="w-full bg-transparent border-b py-2 pr-7 text-left text-sm font-light focus:outline-none flex items-center justify-between gap-2 transition-colors duration-700"
      :class="open ? 'border-accent' : 'border-line'">
      <span :class="current ? 'text-ink' : 'text-ink3'">{{ current ? current.label : (placeholder || '请选择') }}</span>
      <CloseOutline v-if="clearable && current" class="w-4 h-4 text-ink3 hover:text-ink shrink-0"
        @click.stop="open = false; $emit('update:modelValue', '')" />
    </button>
    <Transition name="ink-fade">
      <ul v-if="open"
        class="absolute left-0 right-0 top-full mt-1 z-30 max-h-56 overflow-y-auto bg-paper border border-line py-1 m-0 p-0 list-none shadow-none">
        <li v-for="(o, i) in options" :key="o.value"
          class="px-4 py-1.5 text-sm font-light cursor-pointer transition-colors duration-500"
          :class="o.value === modelValue ? 'text-ink bg-paper2' : i === hovered ? 'text-accent-strong' : 'text-ink2'"
          @mouseenter="hovered = i" @click="choose(i)">{{ o.label }}</li>
      </ul>
    </Transition>
  </span>
</template>

<script setup>
import { computed, onBeforeUnmount, onMounted, ref } from 'vue'
import { CloseOutline } from '@vicons/ionicons5'

const props = defineProps({
  modelValue: { type: [String, Number], default: '' },
  options: { type: Array, default: () => [] }, // [{label, value}]
  placeholder: String,
  clearable: Boolean,
})
const emit = defineEmits(['update:modelValue'])

const open = ref(false)
const hovered = ref(0)
const root = ref(null)
const current = computed(() => props.options.find(o => o.value === props.modelValue))

function choose(i) {
  emit('update:modelValue', props.options[i].value)
  open.value = false
}
function move(d) {
  if (!open.value) { open.value = true; return }
  hovered.value = (hovered.value + d + props.options.length) % props.options.length
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
