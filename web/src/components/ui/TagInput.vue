<template>
  <div>
    <div class="flex flex-wrap items-center gap-2 mb-2" v-if="modelValue.length">
      <span v-for="t in modelValue" :key="t"
        class="inline-flex items-center gap-1 border border-line px-2 py-0.5 text-xs font-light text-ink2 transition-colors duration-700 hover:border-accent hover:text-ink">
        {{ t }}
        <button type="button" class="bg-transparent border-0 p-0 cursor-pointer text-ink3 hover:text-cinnabar transition-colors duration-700"
          @click="remove(t)" :aria-label="`移除标签 ${t}`">
          <CloseOutline class="w-3 h-3" />
        </button>
      </span>
    </div>
    <input :value="draft" :placeholder="placeholder || '回车添加标签'" @input="draft = $event.target.value"
      @keydown.enter.prevent="add(draft)" @keydown.,.prevent="add(draft)" @blur="add(draft)"
      class="w-full bg-transparent border-b border-line py-2 text-sm font-light focus:outline-none focus:border-accent placeholder:text-ink3 transition-colors duration-700" />
    <div v-if="availableSuggestions.length" class="flex flex-wrap gap-2 mt-2">
      <button v-for="s in availableSuggestions" :key="s" type="button" @click="add(s)"
        class="bg-transparent border-0 p-0 cursor-pointer text-xs font-light text-ink3 hover:text-accent-strong border-b border-transparent hover:border-accent transition-colors duration-700">
        {{ s }}
      </button>
    </div>
  </div>
</template>

<script setup>
import { computed, ref } from 'vue'
import { CloseOutline } from '@vicons/ionicons5'

const props = defineProps({
  modelValue: { type: Array, default: () => [] },
  suggestions: { type: Array, default: () => [] },
  placeholder: String,
})
const emit = defineEmits(['update:modelValue'])

const draft = ref('')
const availableSuggestions = computed(() => props.suggestions.filter(s => !props.modelValue.includes(s)))

function add(v) {
  const t = String(v || '').trim()
  if (t && !props.modelValue.includes(t)) emit('update:modelValue', [...props.modelValue, t])
  draft.value = ''
}
function remove(t) { emit('update:modelValue', props.modelValue.filter(x => x !== t)) }
</script>
