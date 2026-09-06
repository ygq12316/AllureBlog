<template>
  <div class="min-h-screen flex items-center justify-center bg-paper">
    <div class="w-[360px] max-w-[90vw] text-center px-9 py-12 bg-paper border border-line">
      <CreateOutline class="w-8 h-8 text-accent mx-auto" />
      <h1 class="text-2xl font-light tracking-[0.3em] text-ink mt-3 mb-1">笔墨后台</h1>
      <p class="text-[13px] text-ink3 m-0 mb-6">请输入密码以继续</p>
      <div class="flex flex-col gap-4 mb-2">
        <InkInput v-model="username" placeholder="账号" @keydown.enter="doLogin" />
        <InkInput v-model="password" type="password" placeholder="密码" :error="!!error" @keydown.enter="doLogin" />
        <InkButton variant="primary" block :loading="loading" @click="doLogin">登录</InkButton>
      </div>
      <p v-if="error" class="text-xs text-cinnabar mt-2 mb-0">{{ error }}</p>
      <router-link to="/" class="inline-block mt-4 text-xs text-ink3 no-underline transition-colors duration-700 hover:text-accent-strong">← 返回博客</router-link>
    </div>
  </div>
</template>

<script setup>
import { ref } from 'vue'
import { useRouter, useRoute } from 'vue-router'
import { CreateOutline } from '@vicons/ionicons5'
import InkInput from '../components/ui/InkInput.vue'
import InkButton from '../components/ui/InkButton.vue'
import { setToken } from '../api/client'
import { loginAccount } from '../api/visitors'
import { setAccount } from '../composables/useVisitor'
import { toast } from '../composables/useToast'

const router = useRouter()
const route = useRoute()
const username = ref('')
const password = ref('')
const loading = ref(false)
const error = ref('')

async function doLogin() {
  if (!username.value || !password.value) { error.value = '请输入账号和密码'; return }
  loading.value = true; error.value = ''
  try {
    // 统一账号登录：管理员与普通用户同入口，角色由服务端签发进令牌
    const data = await loginAccount({ username: username.value, password: password.value })
    setToken(data.token)
    setAccount(data.visitor)
    if (data.visitor?.role === 'admin') {
      router.push(route.query.redirect || '/admin')
    } else {
      toast.info('该账号无后台权限，已进入博客')
      router.push('/')
    }
  } catch (e) {
    error.value = e.response?.data?.error || '登录失败'
  }
  loading.value = false
}
</script>
