import { ref, computed } from 'vue'
import { dicebearUrl } from '../utils/avatar'

// 唯一用户身份 = 登录账号（访客模式已移除）：未登录仅可浏览，
// 评论/弹幕/精灵对话一律要求登录。登录态存 blog_account（服务端 visitors 表）。
const account = ref(JSON.parse(localStorage.getItem('blog_account') || 'null'))
localStorage.removeItem('blog_visitor') // 清理旧版匿名访客模式的本地残留

const loginVisible = ref(false)
const profileVisible = ref(false)

export function setAccount(v) {
  account.value = v
  if (v) localStorage.setItem('blog_account', JSON.stringify(v))
  else localStorage.removeItem('blog_account')
}

export function openLogin() { loginVisible.value = true }
export function closeLogin() { loginVisible.value = false }
export function openProfile() { profileVisible.value = true }
export function closeProfile() { profileVisible.value = false }

export function useVisitor() {
  const avatarUrl = computed(() => {
    const a = account.value
    if (!a) return ''
    if (a.avatar_url) return a.avatar_url
    return dicebearUrl(a.avatar_style || 'lorelei', a.uuid)
  })

  return { account, avatarUrl, loginVisible, profileVisible, openLogin, closeLogin, openProfile, closeProfile, setAccount }
}
