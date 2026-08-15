import { ref } from 'vue'
import { useVisitor } from './useVisitor'

// 笔墨精灵聊天 — WebSocket 状态机（纯逻辑，视图在 FairyChat.vue）
export function useFairyChat() {
  const { visitor, init } = useVisitor()
  const status = ref('idle') // idle | connecting | ready | offline
  const messages = ref([])
  const streaming = ref(false)
  let ws = null
  let pingTimer = null

  // 找气泡：fromEnd=true 取最晚（token 追加目标），false 取最早（done 定稿目标）
  function openFairy(fromEnd) {
    const list = fromEnd ? [...messages.value].reverse() : messages.value
    return list.find(m => m.role === 'fairy' && !m.done)
  }

  function handle(m) {
    if (m.type === 'auth_result') {
      status.value = 'ready'
      if (m.greeting) messages.value.push({ role: 'fairy', content: m.greeting, done: true })
      startPing()
    } else if (m.type === 'token') {
      let b = openFairy(true)
      if (!b) { b = { role: 'fairy', content: '', done: false }; messages.value.push(b) }
      b.content += m.content
    } else if (m.type === 'done') {
      const b = openFairy(false)
      if (b) { b.done = true; b.interrupted = !!m.interrupted }
      streaming.value = false
    } else if (m.type === 'rejected' || m.type === 'error') {
      messages.value.push({ role: 'system', content: m.display || '精灵暂时无法回应' })
      streaming.value = false
    }
    // pong 仅保活，忽略
  }

  async function connect() {
    if (status.value === 'connecting' || status.value === 'ready') return
    await init() // uuid 由 useVisitor 自动生成并持久化，此处必可得
    if (!visitor.value) return
    status.value = 'connecting'
    const proto = location.protocol === 'https:' ? 'wss' : 'ws'
    try {
      ws = new WebSocket(`${proto}://${location.host}/chat/ws`)
    } catch {
      onDown()
      return
    }
    ws.onopen = () => {
      ws.send(JSON.stringify({
        type: 'auth',
        visitor_uuid: visitor.value.uuid,
        visitor_name: visitor.value.nickname,
      }))
    }
    ws.onmessage = ev => {
      try { handle(JSON.parse(ev.data)) } catch {}
    }
    ws.onclose = onDown
    ws.onerror = onDown
  }

  function send(content) {
    const text = (content || '').trim()
    if (!text || !ws || ws.readyState !== WebSocket.OPEN) return false
    messages.value.push({ role: 'user', content: text, done: true })
    messages.value.push({ role: 'fairy', content: '', done: false })
    streaming.value = true
    ws.send(JSON.stringify({ type: 'message', content: text }))
    return true
  }

  function interrupt() {
    if (ws && ws.readyState === WebSocket.OPEN && streaming.value) {
      ws.send(JSON.stringify({ type: 'interrupt' }))
    }
  }

  // 心跳：25s 低于常见代理 60s 空闲超时
  function startPing() {
    stopPing()
    pingTimer = setInterval(() => {
      if (ws && ws.readyState === WebSocket.OPEN) ws.send(JSON.stringify({ type: 'ping' }))
    }, 25000)
  }
  function stopPing() {
    if (pingTimer) { clearInterval(pingTimer); pingTimer = null }
  }

  function onDown() {
    if (status.value === 'offline') return // onclose/onerror 双触发防重
    status.value = 'offline'
    stopPing()
    const b = openFairy(false)
    if (b) { b.done = true; b.content += '（连接中断）' }
    streaming.value = false
    ws = null
  }

  function disconnect() {
    stopPing()
    if (ws) {
      ws.onclose = null; ws.onerror = null; ws.onmessage = null
      try { ws.close() } catch {}
      ws = null
    }
    status.value = 'idle'
  }

  return { status, messages, streaming, connect, send, interrupt, disconnect }
}
