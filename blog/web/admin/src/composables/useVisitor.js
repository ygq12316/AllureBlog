import { ref, computed } from 'vue'
import axios from 'axios'

// 单例 visitor 状态
const visitor = ref(null)
const initialized = ref(false)
const setupVisible = ref(false)
const loginVisible = ref(false)

// 已登录账号（响应式，跨组件共享：UserAvatar 登录/登出时更新，评论/弹幕实时感知）
const account = ref(JSON.parse(localStorage.getItem('blog_account') || 'null'))

export function setAccount(v) {
  account.value = v
  if (v) localStorage.setItem('blog_account', JSON.stringify(v))
  else localStorage.removeItem('blog_account')
}

export function openSetup() { setupVisible.value = true }
export function closeSetup() { setupVisible.value = false }
export function openLogin() { loginVisible.value = true }
export function closeLogin() { loginVisible.value = false }

function generateUUID() {
  const hex = () => Math.random().toString(36).substring(2, 10)
  return hex() + hex() + hex()
}

export function useVisitor() {
  async function init() {
    if (initialized.value) return visitor.value
    initialized.value = true

    // 从 localStorage 恢复
    const raw = localStorage.getItem('blog_visitor')
    if (raw) {
      try {
        visitor.value = JSON.parse(raw)
        return visitor.value
      } catch {}
    }

    // 新访客
    visitor.value = {
      uuid: generateUUID(),
      nickname: '访客' + Math.random().toString(36).substring(2, 6),
      avatar_style: 'lorelei',
      signature: '',
    }
    localStorage.setItem('blog_visitor', JSON.stringify(visitor.value))

    // 同步到服务端
    try {
      await axios.post('/api/visitor', visitor.value)
    } catch {}
    return visitor.value
  }

  async function update(fields) {
    if (!visitor.value) return
    Object.assign(visitor.value, fields)
    if (!visitor.value.setup) visitor.value.setup = true
    localStorage.setItem('blog_visitor', JSON.stringify(visitor.value))
    try {
      await axios.post('/api/visitor', visitor.value)
    } catch {}
  }

  const avatarUrl = computed(() => {
    if (!visitor.value) return ''
    if (visitor.value.avatar_url) return visitor.value.avatar_url
    return `https://api.dicebear.com/9.x/${visitor.value.avatar_style || 'lorelei'}/svg?seed=${encodeURIComponent(visitor.value.uuid)}`
  })

  const isSetUp = computed(() => visitor.value?.setup === true)

  function showSetup() {
    // 触发弹窗：清除标记让 PublicLayout 重新判定
    const raw = localStorage.getItem('blog_visitor')
    if (raw) {
      try {
        const v = JSON.parse(raw)
        if (v.setup) return true // 已设定
      } catch {}
    }
    return false
  }

  async function requireSetup() {
    if (!initialized.value) await init()
    return isSetUp.value
  }

  return { visitor, account, avatarUrl, isSetUp, setupVisible, loginVisible, init, update, requireSetup, openSetup, closeSetup, openLogin, closeLogin, setAccount }
}
