<template>
  <span class="relative inline-flex w-full items-center">
    <input ref="el" :type="revealed ? 'text' : type" :value="modelValue" :placeholder="placeholder"
      :maxlength="maxlength" :disabled="disabled" @input="onInput" @keydown="$emit('keydown', $event)"
      class="w-full bg-transparent border-b py-2 pr-7 text-sm font-light focus:outline-none placeholder:text-ink3 disabled:opacity-40 transition-colors duration-700"
      :class="error ? 'border-cinnabar' : 'border-line focus:border-accent'" />
    <button v-if="type === 'password'" type="button" @click="revealed = !revealed"
      class="absolute right-0 text-ink3 hover:text-ink transition-colors duration-700 cursor-pointer bg-transparent border-0 p-0"
      :aria-label="revealed ? '隐藏密码' : '显示密码'">
      <EyeOutline v-if="revealed" class="w-4 h-4" />
      <EyeOffOutline v-else class="w-4 h-4" />
    </button>
    <button v-else-if="clearable && modelValue" type="button" @click="clear"
      class="absolute right-0 text-ink3 hover:text-ink transition-colors duration-700 cursor-pointer bg-transparent border-0 p-0" aria-label="清空">
      <CloseOutline class="w-4 h-4" />
    </button>
  </span>
</template>

<script setup>
import { ref } from 'vue'
import { EyeOutline, EyeOffOutline, CloseOutline } from '@vicons/ionicons5'

const props = defineProps({
  modelValue: { type: [String, Number], default: '' },
  type: { type: String, default: 'text' }, // text | password | number
  placeholder: String,
  maxlength: [String, Number],
  clearable: Boolean,
  disabled: Boolean,
  error: Boolean,
})
const emit = defineEmits(['update:modelValue', 'keydown'])

const revealed = ref(false)
function onInput(e) { emit('update:modelValue', e.target.value) }
function clear() { emit('update:modelValue', '') }
</script>
