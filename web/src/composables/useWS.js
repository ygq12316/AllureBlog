import { ref, onUnmounted } from 'vue'

// 通用 WebSocket 订阅：断线指数退避重连，重连成功以 isRetry=true 回调
// onOpen，由调用方触发一次对账（覆盖离线期间错过的广播）。消息一律 JSON。
export function useWS(path, { onMessage, onOpen } = {}) {
  const connected = ref(false)
  let ws = null
  let retry = 0
  let timer = null
  let closed = false

  function connect() {
    if (closed || (ws && ws.readyState <= WebSocket.OPEN)) return
    const proto = location.protocol === 'https:' ? 'wss' : 'ws'
    try {
      ws = new WebSocket(`${proto}://${location.host}${path}`)
    } catch {
      scheduleRetry()
      return
    }
    ws.onopen = () => {
      const isRetry = retry > 0
      retry = 0
      connected.value = true
      onOpen?.(isRetry)
    }
    ws.onmessage = ev => {
      try {
        const msg = JSON.parse(ev.data)
        if (msg) onMessage?.(msg)
      } catch {}
    }
    ws.onclose = () => { connected.value = false; scheduleRetry() }
    ws.onerror = () => { connected.value = false }
  }

  function scheduleRetry() {
    if (closed) return
    clearTimeout(timer)
    timer = setTimeout(connect, Math.min(1000 * 2 ** retry, 15000))
    retry++
  }

  function disconnect() {
    closed = true
    clearTimeout(timer)
    if (ws) {
      ws.onopen = ws.onmessage = ws.onclose = ws.onerror = null
      try { ws.close() } catch {}
      ws = null
    }
    connected.value = false
  }

  onUnmounted(disconnect)
  return { connected, connect, disconnect }
}
