<template>
  <div class="min-h-screen bg-paper">
    <!-- 粒子背景（主内容区后方） -->
    <canvas ref="canvas" class="particle-canvas" />
    <div ref="glow" class="mouse-glow" />

    <div class="flex min-h-screen relative z-[1]">
      <!-- 侧边栏：宣纸底 + 细墨线 -->
      <aside class="sticky top-0 h-screen flex flex-col shrink-0 overflow-hidden border-r border-line bg-paper2 transition-[width] duration-700 ease-in-out"
        :style="{ width: collapsed ? '64px' : '220px' }">
        <div class="flex items-center gap-2.5 px-5 pt-6 pb-3 whitespace-nowrap text-[15px] tracking-[0.25em] text-ink">
          <span class="text-lg leading-none text-accent shrink-0">✽</span>
          <span v-if="!collapsed">笔墨后台</span>
        </div>
        <nav class="flex-1 flex flex-col gap-0.5 px-2.5 py-2" aria-label="后台导航">
          <router-link v-for="m in MENUS" :key="m.key" :to="m.to"
            class="flex items-center gap-3 px-3.5 py-2.5 text-sm font-light no-underline transition-colors duration-700"
            :class="activeKey === m.key
              ? 'bg-ink/5 text-ink'
              : 'text-ink3 hover:text-ink hover:bg-ink/5'"
            :title="m.label">
            <component :is="m.icon" class="w-[18px] h-[18px] shrink-0" :class="activeKey === m.key ? 'text-accent-strong' : ''" />
            <span v-if="!collapsed">{{ m.label }}</span>
          </router-link>
        </nav>
        <div class="flex flex-col items-start gap-3 px-4 py-4 border-t border-line2">
          <ThemeToggle />
          <button v-if="!collapsed" class="bg-transparent border-0 p-0 cursor-pointer text-sm font-light text-ink3 hover:text-ink transition-colors duration-700" @click="goBlog">← 博客</button>
        </div>
      </aside>

      <!-- 折叠开关（贴主内容区左缘） -->
      <button class="fixed top-3.5 z-20 w-[22px] h-11 bg-card border border-line text-ink3 cursor-pointer text-xs transition-[left] duration-700 ease-in-out hover:text-accent-strong hover:border-accent"
        :style="{ left: collapsed ? '72px' : '228px' }" @click="collapsed = !collapsed"
        :title="collapsed ? '展开菜单' : '收起菜单'" :aria-label="collapsed ? '展开菜单' : '收起菜单'">
        {{ collapsed ? '»' : '«' }}
      </button>

      <!-- 主内容区 -->
      <main class="admin-content flex-1 min-w-0 relative z-[1] px-6 md:px-12 py-8 md:py-10 text-[15px]">
        <router-view />
      </main>
    </div>
  </div>
</template>

<script setup>
import { ref, computed, onMounted, onUnmounted } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import {
  SpeedometerOutline, DocumentTextOutline, ChatbubbleEllipsesOutline,
  FilmOutline, FolderOpenOutline, PricetagsOutline, PeopleOutline, SettingsOutline,
} from '@vicons/ionicons5'
import ThemeToggle from '../components/ThemeToggle.vue'
import { useParticles } from '../composables/useParticles'

const route = useRoute(), router = useRouter()
const collapsed = ref(false)

// 菜单唯一事实源:新增后台页只需在此加一行(路由另注册一条)
const MENUS = [
  { label: '仪表盘', key: 'dashboard', to: '/admin', icon: SpeedometerOutline },
  { label: '文章', key: 'articles', to: '/admin/articles', icon: DocumentTextOutline },
  { label: '随笔', key: 'notes', to: '/admin/notes', icon: ChatbubbleEllipsesOutline },
  { label: '弹幕', key: 'danmakus', to: '/admin/danmakus', icon: FilmOutline },
  { label: '分类', key: 'categories', to: '/admin/categories', icon: FolderOpenOutline },
  { label: '标签', key: 'tags', to: '/admin/tags', icon: PricetagsOutline },
  { label: '用户', key: 'visitors', to: '/admin/visitors', icon: PeopleOutline },
  { label: '设置', key: 'settings', to: '/admin/settings', icon: SettingsOutline },
]

// 高亮从路径推导:取 /admin 后的第一段,仪表盘精确匹配
const activeKey = computed(() => {
  const p = route.path
  if (p === '/admin' || p === '/admin/') return 'dashboard'
  const seg = '/' + (p.split('/')[2] || '')
  return MENUS.find(m => m.to === '/admin' + seg)?.key ?? 'dashboard'
})
function goBlog() { window.location.href = '/' }

// 粒子背景(与公开侧共用 useParticles)
const { canvas, glow } = useParticles()
// ≤900px 自动收窄为图标栏
function onResize() { collapsed.value = window.innerWidth <= 900 }
onMounted(() => { onResize(); window.addEventListener('resize', onResize) })
onUnmounted(() => { window.removeEventListener('resize', onResize) })
</script>

<style scoped>
@media (prefers-reduced-motion: reduce) {
  aside, .admin-content { transition: none; }
}
</style>
