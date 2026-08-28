import { createApp } from 'vue'
import App from './App.vue'
import router from './router'
import { setUnauthorizedHandler } from './api/client'
import './assets/main.css'

// 401 清凭证(在 client 内完成)后跳转登录(登录页本身除外)
setUnauthorizedHandler(() => {
  const path = router.currentRoute.value.path
  if (path.startsWith('/admin') && path !== '/login') {
    router.push('/login')
  }
})

createApp(App).use(router).mount('#app')
