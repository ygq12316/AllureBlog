import { ref, computed } from 'vue'
import { saveVisitor } from '../api/visitors'
import { dicebearUrl } from '../utils/avatar'

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
  // 访客 UUID 同时是 agent 会话隔离的 key,必须是全球唯一
  if (crypto?.randomUUID) return crypto.randomUUID()
  // 老浏览器降级:时间戳+随机数拼接
  return 'v-' + Date.now().toString(36) + '-' + Math.random().toString(36).slice(2, 10)
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
      await saveVisitor(visitor.value)
    } catch {}
    return visitor.value
  }

  async function update(fields) {
    if (!visitor.value) return
    Object.assign(visitor.value, fields)
    if (!visitor.value.setup) visitor.value.setup = true
    localStorage.setItem('blog_visitor', JSON.stringify(visitor.value))
    try {
      await saveVisitor(visitor.value)
    } catch {}
  }

  const avatarUrl = computed(() => {
    if (!visitor.value) return ''
    if (visitor.value.avatar_url) return visitor.value.avatar_url
    return dicebearUrl(visitor.value.avatar_style || 'lorelei', visitor.value.uuid)
  })

  const isSetUp = computed(() => visitor.value?.setup === true)

  return { visitor, account, avatarUrl, isSetUp, setupVisible, loginVisible, init, update, openSetup, closeSetup, openLogin, closeLogin, setAccount }
}
