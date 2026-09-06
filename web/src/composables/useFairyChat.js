import { ref, watch } from 'vue'
import { useVisitor } from './useVisitor'

// 笔墨精灵聊天 — WebSocket 状态机（纯逻辑，视图在 FairyChat.vue）
export function useFairyChat() {
  const { account } = useVisitor()
  const status = ref('idle') // idle | connecting | ready | need-login | offline
  const messages = ref([])
  const streaming = ref(false)
  const remaining = ref(null) // 今日剩余对话次数（服务端在 auth_result/done 带回）
  const retrieving = ref(false) // 精灵正在检索博客
  let ws = null
  let pingTimer = null

  // 找气泡：fromEnd=true 取最晚（token 追加目标），false 取最早（done 定稿目标）
  function openFairy(fromEnd) {
    const list = fromEnd ? [...messages.value].reverse() : messages.value
    return list.find(m => m.role === 'fairy' && !m.done)
  }

  function handle(m) {
    if (m.type === 'auth_result') {
      if (!m.success) {
        // 未登录账号 / 身份验证失败：提示后置离线
        messages.value.push({ role: 'system', content: m.greeting || '请先登录后再与精灵对话' })
        status.value = 'offline'
        return
      }
      status.value = 'ready'
      if (m.greeting) messages.value.push({ role: 'fairy', content: m.greeting, done: true })
      if (typeof m.remaining === 'number') remaining.value = m.remaining
      startPing()
    } else if (m.type === 'token') {
      retrieving.value = false
      let b = openFairy(true)
      if (!b) { b = { role: 'fairy', content: '', done: false }; messages.value.push(b) }
      b.content += m.content
    } else if (m.type === 'tool_call') {
      retrieving.value = true // 检索中，首个 token 到来自动消除
    } else if (m.type === 'done') {
      const b = openFairy(false)
      if (b) {
        b.done = true
        b.interrupted = !!m.interrupted
        if (typeof m.final === 'string') b.content = m.final // 服务端过滤后的定稿
      }
      if (typeof m.remaining === 'number') remaining.value = m.remaining
      retrieving.value = false
      streaming.value = false
    } else if (m.type === 'error') {
      messages.value.push({ role: 'system', content: m.display || '精灵暂时无法回应' })
      if (m.code === 'limit_exceeded') remaining.value = 0
      streaming.value = false
    }
    // pong 仅保活，忽略
  }

  async function connect() {
    if (status.value === 'connecting' || status.value === 'ready') return
    // 仅账号登录用户可聊（服务端校验 username）
    if (!account.value) { status.value = 'need-login'; return }
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
        visitor_uuid: account.value.uuid,
        visitor_name: account.value.nickname,
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
    retrieving.value = false
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

  // 登录/登出联动：登录成功自动重连，登出即断开
  watch(account, v => {
    if (v && status.value === 'need-login') connect()
    if (!v && status.value !== 'idle') disconnect()
  })

  return { status, messages, streaming, remaining, retrieving, connect, send, interrupt, disconnect }
}
