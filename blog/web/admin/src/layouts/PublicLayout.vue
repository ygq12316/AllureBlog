<template>
  <div class="blog-root" style="display:flex;flex-direction:column;min-height:100vh;background:var(--bg);color:var(--text);font-family:'LXGW WenKai',serif;line-height:1.8">
    <a href="#blog-main" class="skip-link">跳到主内容</a>
    <canvas ref="canvas" class="particle-canvas" />
    <div ref="glow" class="mouse-glow" />

    <!-- 经典三段式导航：品牌+主导航靠左成组，工具（搜索/主题/头像）靠右 -->
    <nav class="blog-nav" aria-label="主导航">
      <div class="nav-left">
        <router-link to="/" class="nav-brand">
          <n-icon size="18" color="var(--gold)" :component="EditIcon" /> 笔墨
        </router-link>
        <div class="nav-links">
          <router-link v-for="l in navs" :key="l.to" :to="l.to"
            class="nav-link" active-class="nav-link--active">{{ l.label }}</router-link>
        </div>
      </div>
      <div class="nav-right">
        <router-link to="/search" class="nav-icon-btn" aria-label="搜索文章" title="搜索">
          <n-icon size="18" :component="SearchIcon" />
        </router-link>
        <ThemeToggle />
        <UserAvatar />
        <button class="nav-icon-btn nav-burger" aria-label="打开导航菜单" :aria-expanded="drawerOpen" @click="drawerOpen = true">
          <n-icon size="20" :component="MenuIcon" />
        </button>
      </div>
    </nav>

    <!-- 移动端抽屉导航（≤768px 由汉堡展开） -->
    <n-drawer v-model:show="drawerOpen" placement="right" :width="drawerWidth" :auto-focus="true">
      <n-drawer-content title="导航" :native-scrollbar="false">
        <nav class="drawer-nav" aria-label="移动端导航">
          <router-link v-for="l in navs" :key="l.to" :to="l.to"
            class="drawer-link" active-class="drawer-link--active" @click="drawerOpen = false">{{ l.label }}</router-link>
          <router-link to="/search" class="drawer-link" @click="drawerOpen = false">搜索</router-link>
        </nav>
      </n-drawer-content>
    </n-drawer>

    <main id="blog-main" class="blog-main"><router-view /></main>

    <footer class="blog-footer">&copy; 2026 笔墨 &middot; 记录思考，分享生活</footer>

    <VisitorSetup v-if="setupVisible" @close="closeSetup" />
    <FairyChat />
  </div>
</template>

<script setup>
import { ref, onMounted } from 'vue'
import { CreateOutline, SearchOutline, MenuOutline } from '@vicons/ionicons5'
import ThemeToggle from '../components/ThemeToggle.vue'
import UserAvatar from '../components/UserAvatar.vue'
import VisitorSetup from '../components/VisitorSetup.vue'
import FairyChat from '../components/FairyChat.vue'
import { useVisitor } from '../composables/useVisitor'
import { useParticles } from '../composables/useParticles'

const { setupVisible, isSetUp, closeSetup, init } = useVisitor()

onMounted(async () => {
  await init()
  // 如果未设定身份，弹出设置
  if (!isSetUp.value) {
    setupVisible.value = true
  }
})

const EditIcon = CreateOutline, SearchIcon = SearchOutline, MenuIcon = MenuOutline
const drawerOpen = ref(false)
// 抽屉宽度随视口收缩（n-drawer 只接受数值），78vw 上限 300px
const drawerWidth = Math.min(300, Math.round(window.innerWidth * 0.78))
// 主页居首，内容板块（文章/随笔）次之，聚合页（分类）收尾；搜索已移作右侧图标
const navs = [
  { to: '/', label: '主页' },
  { to: '/articles', label: '文章' },
  { to: '/notes', label: '随笔' },
  { to: '/category', label: '分类' },
]

// 水墨粒子背景(与后台共用)
const { canvas, glow } = useParticles()
</script>

<style>
.particle-canvas { position: fixed; top: 0; left: 0; width: 100vw; height: 100vh; pointer-events: none; z-index: 0; }
.mouse-glow { position: fixed; top: 0; left: 0; width: 400px; height: 400px; border-radius: 50%; background: radial-gradient(circle, rgba(184,148,76,0.06) 0%, transparent 70%); pointer-events: none; z-index: 0; opacity: 0; transition: opacity 0.6s; }
.skip-link { position: absolute; left: -9999px; top: 8px; z-index: 100; padding: 8px 16px; background: var(--card); border: 1px solid var(--gold); border-radius: 2px; color: var(--gold); font-size: 13px; text-decoration: none; }
.skip-link:focus { left: 16px; }
.blog-nav { display: flex; align-items: center; justify-content: space-between; gap: 16px; height: 52px; padding: 0 clamp(16px, 5vw, 64px); border-bottom: 1px solid var(--card-border); position: sticky; top: 0; z-index: 50; background: var(--bg); backface-visibility: hidden; -webkit-backface-visibility: hidden; }
.nav-left { display: flex; align-items: center; gap: clamp(20px, 4vw, 48px); min-width: 0; }
.nav-brand { font-size: clamp(14px, 1.5vw, 17px); font-weight: 700; color: var(--gold); text-decoration: none; display: flex; align-items: center; gap: 4px; flex-shrink: 0; }
.nav-links { display: flex; gap: clamp(18px, 3vw, 36px); }
.nav-link { font-size: clamp(12px, 1.1vw, 14px); color: var(--text); text-decoration: none; padding: 6px 2px; transition: color .2s; touch-action: manipulation; -webkit-tap-highlight-color: transparent; }
.nav-link:hover { color: var(--gold); }
.nav-link--active { color: var(--gold); font-weight: 600; }
.nav-right { display: flex; align-items: center; gap: 12px; flex-shrink: 0; }
.nav-icon-btn { width: 40px; height: 40px; display: flex; align-items: center; justify-content: center; border-radius: 50%; color: var(--text2); background: none; border: none; cursor: pointer; text-decoration: none; transition: color .2s, background-color .2s; touch-action: manipulation; -webkit-tap-highlight-color: transparent; }
.nav-icon-btn:hover { color: var(--gold); }
.nav-burger { display: none; }
.blog-main { flex: 1; position: relative; z-index: 1; max-width: 1100px; width: 100%; margin: 0 auto; padding: clamp(24px, 4vh, 48px) clamp(16px, 5vw, 64px); scroll-margin-top: 60px; }
.blog-footer { text-align: center; padding: 16px clamp(16px, 5vw, 64px); border-top: 1px solid var(--card-border); font-size: clamp(10px, 0.8vw, 12px); color: var(--muted); position: relative; z-index: 1; }
.drawer-nav { display: flex; flex-direction: column; padding-top: 4px; }
.drawer-link { padding: 14px 8px; font-size: 15px; color: var(--text); text-decoration: none; border-bottom: 1px dotted var(--card-border); touch-action: manipulation; -webkit-tap-highlight-color: transparent; }
.drawer-link--active { color: var(--gold); font-weight: 600; }
/* 键盘焦点环：金色细描边，呼应水墨印章 */
.nav-link:focus-visible, .nav-brand:focus-visible, .nav-icon-btn:focus-visible, .drawer-link:focus-visible, .skip-link:focus {
  outline: 2px solid var(--gold); outline-offset: 3px; border-radius: 2px;
}
/* 移动端：主导航收进抽屉，工具区压缩间距 */
@media (max-width: 768px) {
  .nav-links { display: none; }
  .nav-burger { display: flex; }
  .nav-right { gap: 6px; }
  .drawer-link:focus-visible { outline-offset: -2px; }
}
/* 减少动态效果：停用光晕与装饰动画 */
@media (prefers-reduced-motion: reduce) {
  .mouse-glow { display: none; }
  .nav-link, .nav-icon-btn { transition: none; }
}
</style>
