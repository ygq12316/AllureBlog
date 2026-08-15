import { createApp } from 'vue'
import naive from 'naive-ui'
import axios from 'axios'
import App from './App.vue'
import router from './router'

// 自动附加 token
axios.interceptors.request.use(config => {
  const token = localStorage.getItem('token')
  if (token) config.headers.Authorization = `Bearer ${token}`
  return config
})

// 401 清凭证并跳转登录（登录页本身除外）
axios.interceptors.response.use(
  res => res,
  err => {
    if (err.response?.status === 401) {
      localStorage.removeItem('token')
      localStorage.removeItem('admin_user')
      if (router.currentRoute.value.path.startsWith('/admin') && router.currentRoute.value.path !== '/login') {
        router.push('/login')
      }
    }
    return Promise.reject(err)
  }
)

createApp(App).use(naive).use(router).mount('#app')
