// 全站唯一的 axios 实例与凭证读写点。
// 视图统一 `import axios from '../api/client'`(保持调用写法不变),
// token 注入、401 处理、凭证存取只在这里发生。
import axios from 'axios'

const TOKEN_KEY = 'token'
const ADMIN_USER_KEY = 'admin_user'

export const getToken = () => localStorage.getItem(TOKEN_KEY)
export const setToken = t => localStorage.setItem(TOKEN_KEY, t)
export const setAdminUser = u => localStorage.setItem(ADMIN_USER_KEY, u)
export const hasAdminToken = () => !!getToken()

// 退出登录:清空管理员凭证(访客身份 blog_visitor/blog_account 由 useVisitor 管理)
export function clearAdminSession() {
  localStorage.removeItem(TOKEN_KEY)
  localStorage.removeItem(ADMIN_USER_KEY)
}

// 401 时的跳转策略由 main.js 注入,避免 client <-> router 循环依赖
let onUnauthorized = null
export function setUnauthorizedHandler(fn) { onUnauthorized = fn }

// 自动附加 token
axios.interceptors.request.use(config => {
  const token = getToken()
  if (token) config.headers.Authorization = `Bearer ${token}`
  return config
})

// 401 清凭证并通知(登录页本身除外)
axios.interceptors.response.use(
  res => res,
  err => {
    if (err.response?.status === 401) {
      clearAdminSession()
      onUnauthorized?.()
    }
    return Promise.reject(err)
  }
)

export default axios
