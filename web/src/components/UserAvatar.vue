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
      </div>

      <div class="flex flex-col gap-3">
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
import { setToken, clearAdminSession } from '../api/client'
import { loginAccount, registerAccount } from '../api/visitors'
import { dicebearUrl } from '../utils/avatar'

const PersonIcon = PersonOutline
const { avatarUrl, openProfile, loginVisible, closeLogin, account: accountUser, setAccount } = useVisitor()

const showModal = ref(false)
watch(loginVisible, v => { if (v) { showModal.value = true; closeLogin() } })
const activeTab = ref('login')
const errorMsg = ref('')
const submitting = ref(false)
const form = ref({ username: '', password: '', confirmPwd: '' })

const canSubmit = computed(() => {
  if (!form.value.username || !form.value.password) return false
  if (activeTab.value === 'register' && form.value.password !== form.value.confirmPwd) return false
  return true
})

const isLoggedIn = computed(() => !!accountUser.value)

const userAvatar = computed(() => avatarUrl.value || dicebearUrl('lorelei', 'default'))

const isAdmin = computed(() => accountUser.value?.role === 'admin')

const dropdownOpts = computed(() => {
  const opts = [
    { label: accountUser.value?.username || '用户', key: 'username' },
    { label: '资料管理', key: 'profile' },
  ]
  // 仅管理员有后台入口（角色随登录令牌由服务端签发）
  if (isAdmin.value) {
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
    let data
    if (activeTab.value === 'login') {
      data = await loginAccount({ username: form.value.username, password: form.value.password })
    } else {
      const uuid = 'acct_' + Date.now().toString(36) + Math.random().toString(36).substring(2, 8)
      data = await registerAccount({ uuid, username: form.value.username, password: form.value.password })
    }
    // 统一账号体系：登录/注册即拿带角色的 JWT，后台访问靠它
    if (data.token) setToken(data.token)
    setAccount(data.visitor)
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
    openProfile()
  } else if (key === 'admin') {
    window.location.href = '/admin'
  }
}
</script>
