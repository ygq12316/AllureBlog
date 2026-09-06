<template>
  <div class="flex items-center">
    <div v-if="!isLoggedIn" @click="showModal = true" title="登录"
      class="w-[34px] h-[34px] rounded-full border border-line bg-paper2 flex items-center justify-center cursor-pointer text-ink3 overflow-hidden shrink-0 transition-colors duration-700 hover:border-accent hover:text-accent-strong">
      <PersonIcon class="w-[18px] h-[18px]" />
    </div>
    <InkDropdown v-else :options="dropdownOpts" @select="onSelect">
      <div class="w-[34px] h-[34px] rounded-full border border-accent bg-paper2 flex items-center justify-center cursor-pointer overflow-hidden shrink-0 transition-colors duration-700 hover:border-accent-strong">
        <img :src="userAvatar" class="w-full h-full object-cover" alt="头像" />
      </div>
    </InkDropdown>

    <InkModal :show="showModal" @update:show="showModal = $event" width="320px">
      <div class="flex items-center justify-center gap-2 mb-5">
        <span class="text-[13px] cursor-pointer transition-colors duration-700"
          :class="activeTab === 'login' ? 'text-ink border-b border-accent pb-0.5' : 'text-ink3 hover:text-ink2'"
          @click="activeTab = 'login'">登录</span>
        <span class="text-[11px] text-line">|</span>
        <span class="text-[13px] cursor-pointer transition-colors duration-700"
          :class="activeTab === 'register' ? 'text-ink border-b border-accent pb-0.5' : 'text-ink3 hover:text-ink2'"
          @click="activeTab = 'register'">注册</span>
        <span class="text-[11px] text-line">|</span>
        <span class="text-[13px] cursor-pointer transition-colors duration-700"
          :class="activeTab === 'admin' ? 'text-ink border-b border-accent pb-0.5' : 'text-ink3 hover:text-ink2'"
          @click="activeTab = 'admin'">管理员</span>
      </div>

      <!-- 管理员登录 -->
      <div v-if="activeTab === 'admin'" class="flex flex-col gap-3">
        <InkInput v-model="adminForm.admin_user" placeholder="管理员账号" />
        <InkInput v-model="adminForm.admin_pass" type="password" placeholder="管理员密码" @keydown.enter="doAdminLogin" />
        <InkButton variant="primary" block size="sm" @click="doAdminLogin" :loading="adminLoading"
          :disabled="!adminForm.admin_user || !adminForm.admin_pass">管理员登录</InkButton>
      </div>

      <!-- 访客登录/注册 -->
      <div v-else class="flex flex-col gap-3">
        <InkInput v-model="form.username" placeholder="用户名" :maxlength="20" clearable />
        <InkInput v-model="form.password" type="password" placeholder="密码" :maxlength="50" @keydown.enter="doSubmit" />
        <InkInput v-if="activeTab === 'register'" v-model="form.confirmPwd" type="password" placeholder="确认密码" :maxlength="50" />
        <InkButton variant="primary" block size="sm" @click="doSubmit" :loading="submitting" :disabled="!canSubmit">
          {{ activeTab === 'login' ? '登录' : '注册' }}
        </InkButton>
      </div>
      <p v-if="errorMsg" class="text-xs text-cinnabar text-center mt-3 leading-snug">{{ errorMsg }}</p>
    </InkModal>
  </div>
</template>

<script setup>
import { ref, computed, watch } from 'vue'
import { PersonOutline } from '@vicons/ionicons5'
import InkDropdown from './ui/InkDropdown.vue'
import InkModal from './ui/InkModal.vue'
import InkInput from './ui/InkInput.vue'
import InkButton from './ui/InkButton.vue'
import { useVisitor } from '../composables/useVisitor'
import { setToken, setAdminUser, hasAdminToken, clearAdminSession } from '../api/client'
import { login } from '../api/auth'
import { loginAccount, registerAccount } from '../api/visitors'
import { dicebearUrl } from '../utils/avatar'

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
    const data = await login(adminForm.value.admin_user, adminForm.value.admin_pass)
    setToken(data.token)
    setAdminUser(data.user)
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
  const v = visitor.value, a = accountUser.value
  if (v?.avatar_url) return v.avatar_url
  if (a?.avatar_url) return a.avatar_url
  return dicebearUrl(v?.avatar_style || a?.avatar_style || 'lorelei', v?.uuid || a?.uuid || 'default')
})

const dropdownOpts = computed(() => {
  const opts = [
    { label: accountUser.value?.username || '用户', key: 'username' },
    { label: '资料管理', key: 'profile' },
  ]
  if (hasAdminToken()) {
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
      const data = await loginAccount({ username: form.value.username, password: form.value.password })
      setAccount(data.visitor)
      await update({ uuid: data.visitor.uuid, nickname: data.visitor.nickname, avatar_style: data.visitor.avatar_style, avatar_url: data.visitor.avatar_url || '' })
    } else {
      const uuid = 'acct_' + Date.now().toString(36) + Math.random().toString(36).substring(2, 8)
      const data = await registerAccount({ uuid, username: form.value.username, password: form.value.password })
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
    clearAdminSession()
  } else if (key === 'profile') {
    openSetup()
  } else if (key === 'admin') {
    window.location.href = '/admin'
  }
}
</script>
