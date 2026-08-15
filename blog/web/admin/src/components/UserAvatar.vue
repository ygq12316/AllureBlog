<template>
  <div class="ua-wrap">
    <div v-if="!isLoggedIn" class="ua-avatar" @click="showModal = true" title="登录">
      <n-icon size="18" :component="PersonIcon" />
    </div>
    <n-dropdown v-else trigger="hover" :options="dropdownOpts" @select="onSelect">
      <div class="ua-avatar ua-avatar--logged">
        <img :src="userAvatar" class="ua-img" />
      </div>
    </n-dropdown>

    <n-modal :show="showModal" @update:show="showModal = $event" transform-origin="center">
      <div class="auth-card">
        <div class="auth-tabs">
          <span class="auth-tab" :class="{ active: activeTab==='login' }" @click="activeTab='login'">登录</span>
          <span class="auth-tab-sep">|</span>
          <span class="auth-tab" :class="{ active: activeTab==='register' }" @click="activeTab='register'">注册</span>
          <span class="auth-tab-sep">|</span>
          <span class="auth-tab" :class="{ active: activeTab==='admin' }" @click="activeTab='admin'">管理员</span>
        </div>

        <!-- 管理员登录 -->
        <div v-if="activeTab==='admin'" class="auth-form">
          <n-input v-model:value="adminForm.admin_user" placeholder="管理员账号" size="small" />
          <n-input v-model:value="adminForm.admin_pass" type="password" placeholder="管理员密码" size="small" @keydown.enter="doAdminLogin" show-password-on="click" />
          <n-button type="warning" block size="small" @click="doAdminLogin" :loading="adminLoading" :disabled="!adminForm.admin_user||!adminForm.admin_pass">管理员登录</n-button>
        </div>

        <!-- 访客登录/注册 -->
        <div v-else class="auth-form">
          <n-input v-model:value="form.username" placeholder="用户名" size="small" :maxlength="20" clearable />
          <n-input v-model:value="form.password" type="password" placeholder="密码" size="small" :maxlength="50" @keydown.enter="doSubmit" show-password-on="click" />
          <n-input v-if="activeTab==='register'" v-model:value="form.confirmPwd" type="password" placeholder="确认密码" size="small" :maxlength="50" show-password-on="click" />
          <n-button type="primary" block size="small" @click="doSubmit" :loading="submitting" :disabled="!canSubmit">
            {{ activeTab === 'login' ? '登录' : '注册' }}
          </n-button>
        </div>
        <p v-if="errorMsg" class="auth-err">{{ errorMsg }}</p>
      </div>
    </n-modal>
  </div>
</template>

<script setup>
import { ref, computed, watch } from 'vue'
import { PersonOutline } from '@vicons/ionicons5'
import { useVisitor } from '../composables/useVisitor'
import axios from 'axios'

const PersonIcon = PersonOutline
const { visitor, isSetUp, init, update, openSetup, loginVisible, closeLogin, account: accountUser, setAccount } = useVisitor()

const showModal = ref(false)
watch(loginVisible, v => { if (v) { showModal.value = true; closeLogin() } })
const activeTab = ref('login')
const errorMsg = ref('')
const submitting = ref(false)
const adminLoading = ref(false)
const form = ref({ username: '', password: '', confirmPwd: '' })
const adminForm = ref({ admin_user: '', admin_pass: '' })

async function doAdminLogin() {
  errorMsg.value = ''; adminLoading.value = true
  try {
    const { data } = await axios.post('/api/login', { username: adminForm.value.admin_user, password: adminForm.value.admin_pass })
    localStorage.setItem('token', data.token)
    localStorage.setItem('admin_user', data.user)
    adminForm.value.admin_pass = ''
    showModal.value = false
    window.location.href = '/admin'
  } catch (e) {
    errorMsg.value = e.response?.data?.error || '管理员登录失败'
  } finally { adminLoading.value = false }
}

const canSubmit = computed(() => {
  if (!form.value.username || !form.value.password) return false
  if (activeTab.value === 'register' && form.value.password !== form.value.confirmPwd) return false
  return true
})

const isLoggedIn = computed(() => !!accountUser.value)

watch(() => visitor.value, (v) => {
  if (v && accountUser.value) {
    const changed = v.avatar_url !== accountUser.value.avatar_url || v.avatar_style !== accountUser.value.avatar_style
    if (changed) {
      setAccount({ ...accountUser.value, avatar_url: v.avatar_url, avatar_style: v.avatar_style })
    }
  }
}, { deep: true })

const userAvatar = computed(() => {
  const v = visitor.value
  if (v?.avatar_url) return v.avatar_url
  const a = accountUser.value
  if (a?.avatar_url) return a.avatar_url
  const u = v?.uuid || a?.uuid || 'default'
  const style = v?.avatar_style || a?.avatar_style || 'lorelei'
  return `https://api.dicebear.com/9.x/${style}/svg?seed=${encodeURIComponent(u)}`
})

const dropdownOpts = computed(() => {
  const opts = [
    { label: accountUser.value?.username || '用户', key: 'username' },
    { label: '资料管理', key: 'profile' },
  ]
  if (localStorage.getItem('token')) {
    opts.push({ label: '进入后台', key: 'admin' })
  }
  opts.push({ type: 'divider', key: 'd1' }, { label: '退出登录', key: 'logout' })
  return opts
})

async function doSubmit() {
  errorMsg.value = ''
  if (!canSubmit.value) return
  submitting.value = true
  try {
    if (activeTab.value === 'login') {
      const { data } = await axios.post('/api/visitor/login', { username: form.value.username, password: form.value.password })
      setAccount(data.visitor)
      await update({ uuid: data.visitor.uuid, nickname: data.visitor.nickname, avatar_style: data.visitor.avatar_style, avatar_url: data.visitor.avatar_url || '' })
    } else {
      const uuid = 'acct_' + Date.now().toString(36) + Math.random().toString(36).substring(2, 8)
      const { data } = await axios.post('/api/visitor/register', { uuid, username: form.value.username, password: form.value.password })
      setAccount(data.visitor)
      await update({ uuid: data.visitor.uuid, nickname: data.visitor.nickname, avatar_style: data.visitor.avatar_style, avatar_url: data.visitor.avatar_url || '' })
    }
    showModal.value = false
    form.value = { username: '', password: '', confirmPwd: '' }
  } catch (e) {
    errorMsg.value = e.response?.data?.error || '操作失败'
  } finally {
    submitting.value = false
  }
}

function onSelect(key) {
  if (key === 'logout') {
    setAccount(null)
    localStorage.removeItem('token')
    localStorage.removeItem('admin_user')
  } else if (key === 'profile') {
    openSetup()
  } else if (key === 'admin') {
    window.location.href = '/admin'
  }
}
</script>

<style scoped>
.ua-wrap { display: flex; align-items: center; }
.ua-avatar { width: 34px; height: 34px; border-radius: 50%; background: var(--tag-bg); border: 2px solid var(--border); display: flex; align-items: center; justify-content: center; cursor: pointer; color: var(--muted); flex-shrink: 0; overflow: hidden; transition: all .25s; }
.ua-avatar:hover { border-color: var(--gold); color: var(--gold); transform: scale(1.05); }
.ua-avatar--logged { border-color: var(--gold); }
.ua-img { width: 100%; height: 100%; object-fit: cover; }
.auth-card { width: 280px; max-width: 86vw; border-radius: 10px; background: var(--card); border: 1px solid var(--card-border); box-shadow: 0 8px 32px rgba(0,0,0,.1); padding: 20px 20px 16px; }
.auth-tabs { display: flex; align-items: center; gap: 8px; justify-content: center; margin-bottom: 14px; }
.auth-tab { font-size: 13px; color: var(--muted); cursor: pointer; transition: color .2s; }
.auth-tab.active { color: var(--gold); font-weight: 600; }
.auth-tab-sep { font-size: 11px; color: var(--border); }
.auth-form { display: flex; flex-direction: column; gap: 8px; }
.auth-err { font-size: 11px; color: #c97a4a; text-align: center; margin: 8px 0 0; line-height: 1.4; }
.auth-foot { display: flex; justify-content: center; margin-top: 10px; padding-top: 10px; border-top: 1px solid var(--card-border); font-size: 11px; }
.auth-foot a { color: var(--muted); cursor: pointer; transition: color .2s; }
.auth-foot a:hover { color: var(--gold); }
</style>