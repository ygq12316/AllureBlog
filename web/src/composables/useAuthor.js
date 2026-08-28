// 博主信息单例:昵称/头像(/api/config)+ 签名(/api/visitor/admin_admin 假访客)。
// 此前 4 个文件各自拉取无缓存,Settings 保存后也无法让其他页面感知。
import { ref } from 'vue'
import { getConfig } from '../api/config'
import { getVisitor } from '../api/visitors'

const author = ref({ name: 'Allure', avatar: '', signature: '' })
let loaded = false
let pending = null

async function load() {
  if (pending) return pending
  pending = (async () => {
    try {
      const cfg = await getConfig()
      author.value.name = cfg?.author_name || 'Allure'
      author.value.avatar = cfg?.author_avatar || ''
    } catch {}
    try {
      const v = await getVisitor('admin_admin')
      author.value.signature = v?.signature || ''
    } catch {}
    loaded = true
  })()
  return pending
}

// Settings 保存后调用,全站消费方随之更新
async function refresh() {
  if (!loaded && !pending) return load()
  pending = null
  return load()
}

export function useAuthor() {
  if (!loaded && !pending) load()
  return { author, load, refresh }
}
