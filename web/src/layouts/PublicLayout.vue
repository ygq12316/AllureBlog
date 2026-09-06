<template>
  <div class="flex flex-col min-h-screen bg-paper text-ink font-serif leading-relaxed transition-colors duration-700">
    <a href="#blog-main" class="skip-link">跳到主内容</a>
    <canvas ref="canvas" class="particle-canvas" />
    <div ref="glow" class="mouse-glow" />

    <!-- 三段式导航：品牌+主导航靠左成组，工具（搜索/主题/头像）靠右 -->
    <nav class="sticky top-0 z-50 flex items-center justify-between gap-4 h-14 px-6 md:px-16 bg-paper border-b border-line"
      aria-label="主导航">
      <div class="flex items-center gap-8 md:gap-12 min-w-0">
        <router-link to="/" class="flex items-center gap-1.5 shrink-0 text-base tracking-[0.3em] text-ink no-underline transition-colors duration-700 hover:text-accent-strong">
          <EditIcon class="w-4 h-4 text-accent" /> 笔墨
        </router-link>
        <div class="hidden md:flex items-center gap-8">
          <router-link v-for="l in navs" :key="l.to" :to="l.to"
            class="px-0.5 py-1.5 text-sm font-light text-ink2 no-underline border-b border-transparent transition-colors duration-700 hover:text-accent-strong hover:border-accent/40"
            :class="{ '!text-accent-strong !border-accent/60': isActive(l.to) }">{{ l.label }}</router-link>
        </div>
      </div>
      <div class="flex items-center gap-1.5 md:gap-3 shrink-0">
        <router-link to="/search" class="w-9 h-9 flex items-center justify-center text-ink2 transition-colors duration-700 hover:text-accent-strong" aria-label="搜索文章" title="搜索">
          <SearchIcon class="w-[18px] h-[18px]" />
        </router-link>
        <ThemeToggle />
        <UserAvatar />
        <button class="md:hidden w-9 h-9 flex items-center justify-center text-ink2 transition-colors duration-700 hover:text-accent-strong bg-transparent border-0 cursor-pointer"
          aria-label="打开导航菜单" :aria-expanded="drawerOpen" @click="drawerOpen = true">
          <MenuIcon class="w-5 h-5" />
        </button>
      </div>
    </nav>

    <!-- 移动端抽屉导航 -->
    <InkDrawer v-model:show="drawerOpen" title="导航" max-width="300px">
      <nav class="flex flex-col py-2" aria-label="移动端导航">
        <router-link v-for="l in [...navs, { to: '/search', label: '搜索' }]" :key="l.to" :to="l.to"
          class="px-8 py-3.5 text-[15px] font-light text-ink2 no-underline border-b border-line2 transition-colors duration-700 hover:text-accent-strong"
          :class="{ '!text-accent-strong': isActive(l.to) }" @click="drawerOpen = false">{{ l.label }}</router-link>
      </nav>
    </InkDrawer>

    <main id="blog-main" class="flex-1 relative z-[1] w-full max-w-[1100px] mx-auto px-6 md:px-16 py-10 md:py-14 scroll-mt-16">
      <router-view />
    </main>

    <footer class="relative z-[1] text-center py-10 px-6 md:px-16 border-t border-line text-xs font-light tracking-widest text-ink3">
      &copy; 2026 笔墨 &middot; 记录思考，分享生活
    </footer>

    <ProfileModal v-if="profileVisible" @close="closeProfile" />
    <FairyChat />
  </div>
</template>

<script setup>
import { ref } from 'vue'
import { useRoute } from 'vue-router'
import { CreateOutline, SearchOutline, MenuOutline } from '@vicons/ionicons5'
import ThemeToggle from '../components/ThemeToggle.vue'
import UserAvatar from '../components/UserAvatar.vue'
import ProfileModal from '../components/ProfileModal.vue'
import FairyChat from '../components/FairyChat.vue'
import { useVisitor } from '../composables/useVisitor'
import { useParticles } from '../composables/useParticles'

const route = useRoute()
const { profileVisible, closeProfile } = useVisitor()

const EditIcon = CreateOutline, SearchIcon = SearchOutline, MenuIcon = MenuOutline
const drawerOpen = ref(false)
// 主页居首，内容板块（文章/随笔）次之，聚合页（分类）收尾；搜索已移作右侧图标
const navs = [
  { to: '/', label: '主页' },
  { to: '/articles', label: '文章' },
  { to: '/notes', label: '随笔' },
  { to: '/category', label: '分类' },
]
// 高亮：主页精确匹配，其余前缀匹配（覆盖 /posts/:slug 等详情页）
function isActive(to) {
  if (to === '/') return route.path === '/'
  return route.path === to || route.path.startsWith(to + '/')
}

// 水墨粒子背景(与后台共用)
const { canvas, glow } = useParticles()
</script>

<style>
/* 键盘焦点环 */
.nav :focus-visible { outline: 1px solid var(--accent); outline-offset: 3px; }
/* 移动端：主导航收进抽屉 */
@media (max-width: 768px) {
  .nav-links { display: none; }
}
@media (prefers-reduced-motion: reduce) {
  nav a, nav button { transition: none; }
}
</style>
