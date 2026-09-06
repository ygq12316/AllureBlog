<template>
  <table class="w-full text-sm font-light border-collapse">
    <thead>
      <tr class="border-b border-line">
        <th v-if="selectable" class="w-10 px-3 py-3 text-left">
          <input type="checkbox" :checked="allChecked" @change="toggleAll" aria-label="全选"
            class="w-3.5 h-3.5 cursor-pointer align-middle" :style="{ accentColor: 'var(--ink)' }" />
        </th>
        <th v-for="c in columns" :key="c.key ?? c.title"
          class="px-3 py-3 text-left text-xs font-light tracking-widest text-ink3"
          :style="thStyle(c)">{{ c.title }}</th>
        <th v-if="$slots.actions" class="w-28 px-3 py-3" />
      </tr>
    </thead>
    <tbody>
      <tr v-if="!data.length">
        <td :colspan="colspan" class="px-3 py-2">
          <slot name="empty" />
        </td>
      </tr>
      <tr v-for="row in data" :key="rowKey(row)" class="border-b border-line2 transition-colors duration-700 hover:bg-paper2">
        <td v-if="selectable" class="px-3 py-3 align-middle">
          <input type="checkbox" :checked="checked.includes(rowKey(row))" @change="toggleRow(row)"
            :aria-label="`选中 ${rowKey(row)}`" class="w-3.5 h-3.5 cursor-pointer align-middle" :style="{ accentColor: 'var(--ink)' }" />
        </td>
        <td v-for="c in columns" :key="c.key ?? c.title" class="px-3 py-3 align-middle text-ink2"
          :style="tdStyle(c)">
          <slot v-if="c.slot" :name="'cell-' + c.key" :row="row" :value="row[c.key]" />
          <div v-else-if="c.ellipsis" class="truncate" :title="String(display(c, row) ?? '')">{{ display(c, row) }}</div>
          <template v-else>{{ display(c, row) }}</template>
        </td>
        <td v-if="$slots.actions" class="px-3 py-3 align-middle">
          <slot name="actions" :row="row" />
        </td>
      </tr>
    </tbody>
  </table>
</template>

<script setup>
import { computed } from 'vue'

const props = defineProps({
  columns: { type: Array, default: () => [] }, // {title, key, width, slot, ellipsis, render(row)}
  data: { type: Array, default: () => [] },
  rowKey: { type: Function, default: r => r.id },
  selectable: Boolean,
  checked: { type: Array, default: () => [] },
})
const emit = defineEmits(['update:checked'])

const colspan = computed(() => props.columns.length + (props.selectable ? 1 : 0))
const allChecked = computed(() => props.data.length > 0 && props.data.every(r => props.checked.includes(props.rowKey(r))))

function display(c, row) {
  if (c.render) return c.render(row)
  return row[c.key]
}
function thStyle(c) {
  return c.width && c.width !== '*' ? { width: c.width + 'px' } : null
}
function tdStyle(c) {
  if (!c.ellipsis) return null
  const mw = c.width === '*' ? '28rem' : c.width ? (c.width - 24) + 'px' : '12rem'
  return { maxWidth: mw }
}
function toggleAll() {
  emit('update:checked', allChecked.value ? [] : props.data.map(props.rowKey))
}
function toggleRow(row) {
  const k = props.rowKey(row)
  emit('update:checked', props.checked.includes(k) ? props.checked.filter(x => x !== k) : [...props.checked, k])
}
</script>
