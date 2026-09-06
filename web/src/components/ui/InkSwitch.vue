<template>
  <button type="button" role="switch" :aria-checked="modelValue" @click="$emit('update:modelValue', !modelValue)"
    class="relative inline-flex items-center cursor-pointer bg-transparent border-0 p-0 gap-2">
    <slot name="unchecked" />
    <span class="relative inline-block w-10 h-5 border border-line transition-colors duration-700"
      :class="modelValue ? 'bg-ink/10 border-accent' : 'bg-transparent'">
      <span class="absolute top-0.5 left-0.5 w-3.5 h-3.5 transition-transform duration-700 ease-in-out"
        :class="modelValue ? 'translate-x-[18px] bg-accent' : 'translate-x-0 bg-ink3'" />
    </span>
    <slot name="checked" />
  </button>
</template>

<script setup>
defineProps({ modelValue: Boolean })
defineEmits(['update:modelValue'])
</script>
