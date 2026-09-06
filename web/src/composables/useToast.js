// 轻量 toast：模块级 reactive 队列，App.vue 挂 InkToaster 消费
import { reactive } from 'vue'

const state = reactive({ items: [] })
let seq = 0

function push(type, content) {
  const id = ++seq
  state.items.push({ id, type, content })
  setTimeout(() => dismiss(id), 3000)
}

function dismiss(id) {
  const i = state.items.findIndex(t => t.id === id)
  if (i > -1) state.items.splice(i, 1)
}

export const toast = {
  success: c => push('success', c),
  error: c => push('error', c),
  info: c => push('info', c),
}

export function useToast() {
  return { items: state.items, dismiss }
}
