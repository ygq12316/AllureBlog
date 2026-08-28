<template>
  <div class="login-page">
    <div class="login-card">
      <n-icon size="32" color="var(--gold)" :component="CreateOutline" />
      <h1 class="login-title">笔墨后台</h1>
      <p class="login-sub">请输入密码以继续</p>
      <n-input v-model:value="username" placeholder="账号" size="large" class="login-input" />
      <n-input
        v-model:value="password"
        type="password"
        placeholder="密码"
        size="large"
        :status="error ? 'error' : undefined"
        @keyup.enter="doLogin"
        class="login-input" />
      <n-button
        type="primary"
        size="large"
        block
        :loading="loading"
        @click="doLogin"
        class="login-btn">登录</n-button>
      <p v-if="error" class="login-error">{{ error }}</p>
      <router-link to="/" class="login-back">← 返回博客</router-link>
    </div>
  </div>
</template>

<script setup>
import { ref } from 'vue'
import { useRouter } from 'vue-router'
import { CreateOutline } from '@vicons/ionicons5'
import { setToken } from '../api/client'
import { login } from '../api/auth'

const router = useRouter()
const username = ref('')
const password = ref('')
const loading = ref(false)
const error = ref('')

async function doLogin() {
  if (!username.value || !password.value) { error.value = '请输入账号和密码'; return }
  loading.value = true; error.value = ''
  try {
    const data = await login(username.value, password.value)
    setToken(data.token)
    router.push('/admin')
  } catch (e) {
    error.value = e.response?.data?.error || '登录失败'
  }
  loading.value = false
}
</script>

<style scoped>
.login-page {
  min-height: 100vh; display: flex; align-items: center; justify-content: center;
  background: linear-gradient(135deg, var(--bg) 0%, var(--card) 100%);
}
.login-card {
  width: 360px; text-align: center;
  padding: 48px 36px 36px;
  background: var(--bg); border: 1px solid var(--card-border);
  border-radius: 8px; box-shadow: 0 4px 24px rgba(0,0,0,.06);
}
.login-title { font-size: 22px; font-weight: 700; color: var(--text); margin: 12px 0 4px; }
.login-sub { font-size: 13px; color: var(--muted); margin: 0 0 24px; }
.login-input { margin-bottom: 16px; }
.login-btn { margin-bottom: 8px; }
.login-error { font-size: 12px; color: #d03050; margin: 8px 0 0; }
.login-back { display: inline-block; margin-top: 16px; font-size: 12px; color: var(--muted); text-decoration: none; }
.login-back:hover { color: var(--gold); }
</style>
