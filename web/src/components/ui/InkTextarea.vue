<template>
  <textarea :value="modelValue" :rows="rows" :placeholder="placeholder" :maxlength="maxlength" :disabled="disabled"
    @input="$emit('update:modelValue', $event.target.value)"
    class="w-full bg-transparent border border-line focus:outline-none focus:border-accent p-3 text-sm font-light leading-relaxed placeholder:text-ink3 disabled:opacity-40 transition-colors duration-700 resize-y" />
</template>

<script setup>
defineProps({
  modelValue: { type: String, default: '' },
  rows: { type: [String, Number], default: 4 },
  placeholder: String,
  maxlength: [String, Number],
  disabled: Boolean,
})
defineEmits(['update:modelValue'])
</script>
