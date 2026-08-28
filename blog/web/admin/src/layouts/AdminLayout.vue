<template>
  <div style="min-height:100vh;background:var(--bg)">
    <!-- 粒子背景（主内容区后方） -->
    <canvas ref="canvas" class="particle-canvas" />
    <div ref="glow" class="mouse-glow" />

    <n-config-provider :theme-overrides="themeOverrides">
      <div class="shell">
        <!-- 侧边栏：固定深墨色，明暗主题下不变 -->
        <aside class="sider" :class="{ collapsed }">
          <div class="brand">
            <span class="brand-mark">✽</span>
            <span v-if="!collapsed" class="brand-name">笔墨后台</span>
          </div>
          <n-menu
            v-model:value="activeMenu"
            :options="menuOptions"
            :collapsed="collapsed"
            :collapsed-width="64"
            :collapsed-icon-size="20"
            class="sider-menu"
            @update:value="onMenuClick"
          />
          <div class="sider-foot">
            <ThemeToggle />
            <button v-if="!collapsed" class="sider-link" @click="goBlog">← 博客</button>
          </div>
        </aside>

        <!-- 折叠开关（贴主内容区左缘） -->
        <button class="collapse-btn" :class="{ shifted: collapsed }" @click="collapsed = !collapsed" :title="collapsed ? '展开菜单' : '收起菜单'">
          {{ collapsed ? '»' : '«' }}
        </button>

        <!-- 主内容区 -->
        <main class="admin-content"><router-view /></main>
      </div>
    </n-config-provider>
  </div>
</template>

<script setup>
import { ref, computed, onMounted, onUnmounted, h } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { NIcon } from 'naive-ui'
import {
  SpeedometerOutline, DocumentTextOutline, ChatbubbleEllipsesOutline,
  FilmOutline, FolderOpenOutline, PricetagsOutline, PeopleOutline, SettingsOutline,
} from '@vicons/ionicons5'
import ThemeToggle from '../components/ThemeToggle.vue'
import { useParticles } from '../composables/useParticles'

const route = useRoute(), router = useRouter()
const collapsed = ref(false)

// 菜单唯一事实源:新增后台页只需在此加一行(路由另注册一条)
const renderIcon = comp => () => h(NIcon, null, { default: () => h(comp) })
const MENUS = [
  { label: '仪表盘', key: 'dashboard', to: '/admin', icon: renderIcon(SpeedometerOutline) },
  { label: '文章', key: 'articles', to: '/admin/articles', icon: renderIcon(DocumentTextOutline) },
  { label: '随笔', key: 'notes', to: '/admin/notes', icon: renderIcon(ChatbubbleEllipsesOutline) },
  { label: '弹幕', key: 'danmakus', to: '/admin/danmakus', icon: renderIcon(FilmOutline) },
  { label: '分类', key: 'categories', to: '/admin/categories', icon: renderIcon(FolderOpenOutline) },
  { label: '标签', key: 'tags', to: '/admin/tags', icon: renderIcon(PricetagsOutline) },
  { label: '访客', key: 'visitors', to: '/admin/visitors', icon: renderIcon(PeopleOutline) },
  { label: '设置', key: 'settings', to: '/admin/settings', icon: renderIcon(SettingsOutline) },
]
const menuOptions = MENUS

// 高亮从路径推导:取 /admin 后的第一段,仪表盘精确匹配
const activeMenu = computed(() => {
  const p = route.path
  if (p === '/admin' || p === '/admin/') return 'dashboard'
  const seg = '/' + (p.split('/')[2] || '')
  return MENUS.find(m => m.to === '/admin' + seg)?.key ?? 'dashboard'
})
function onMenuClick(key) {
  const m = MENUS.find(m => m.key === key)
  if (m) router.push(m.to)
}
function goBlog() { window.location.href = '/' }

const themeOverrides = {
  common: { primaryColor: '#b8944c', primaryColorHover: '#d4b060', borderRadius: '4px' },
}

// 粒子背景(与公开侧共用 useParticles)
const { canvas, glow } = useParticles()
// ≤900px 自动收窄为图标栏
function onResize() { collapsed.value = window.innerWidth <= 900 }
onMounted(() => { onResize(); window.addEventListener('resize', onResize) })
onUnmounted(() => { window.removeEventListener('resize', onResize) })
</script>

<style scoped>
.particle-canvas { position: fixed; top: 0; left: 0; width: 100vw; height: 100vh; pointer-events: none; z-index: 0 }
.mouse-glow { position: fixed; top: 0; left: 0; width: 400px; height: 400px; border-radius: 50%; background: radial-gradient(circle, rgba(184, 148, 76, .06) 0%, transparent 70%); pointer-events: none; z-index: 0; opacity: 0; transition: opacity .6s }

.shell { display: flex; min-height: 100vh; position: relative; z-index: 1 }

/* 侧边栏：深墨底 + 金色高亮，主题切换不影响 */
.sider {
  width: 220px; flex-shrink: 0; background: #26211a; border-right: 1px solid #3d3830;
  display: flex; flex-direction: column; position: sticky; top: 0; height: 100vh;
  transition: width .2s; overflow: hidden;
}
.sider.collapsed { width: 64px }
.brand { display: flex; align-items: center; gap: 10px; padding: 18px 20px 14px; color: #d8b96a; font-weight: 700; font-size: 15px; white-space: nowrap }
.brand-mark { font-size: 18px; flex-shrink: 0 }
.sider-menu { flex: 1; background: transparent; --n-item-text-color: #9c8f76; --n-item-text-color-hover: #e8dcc0; --n-item-text-color-active: #26211a; --n-item-text-color-active-hover: #26211a; --n-item-color-active: #d8b96a; --n-item-color-active-hover: #e2c678; --n-item-icon-color: #9c8f76; --n-item-icon-color-hover: #e8dcc0; --n-item-icon-color-active: #26211a; --n-item-color-active-collapsed: #d8b96a; --n-arrow-color: #9c8f76; --n-item-text-color-child-active: #d8b96a; --n-item-icon-color-child-active: #d8b96a; }
.sider-foot { padding: 14px 16px; border-top: 1px solid #3d3830; display: flex; flex-direction: column; gap: 10px; align-items: flex-start }
.sider-link { background: none; border: none; color: #9c8f76; font-size: 13px; cursor: pointer; font-family: inherit; padding: 0 }
.sider-link:hover { color: #e8dcc0 }

/* 折叠按钮 */
.collapse-btn {
  position: fixed; top: 14px; left: 228px; z-index: 20; width: 22px; height: 44px;
  background: var(--card); border: 1px solid var(--card-border); border-radius: 6px;
  color: var(--muted); cursor: pointer; transition: left .2s; font-size: 12px;
}
.collapse-btn:hover { color: var(--gold); border-color: var(--gold) }
.collapse-btn.shifted { left: 72px }

/* 主内容区 */
.admin-content { flex: 1; min-width: 0; padding: clamp(20px, 3vh, 32px) clamp(16px, 5vw, 48px); position: relative; z-index: 1; font-size: 15px; }

/* 深色输入框：覆盖 Naive UI 默认白色背景 */
:deep(.n-input) { --n-color: transparent !important; --n-color-focus: transparent !important; --n-text-color: var(--text) !important; --n-placeholder-color: var(--muted) !important; --n-border: 1px solid var(--card-border) !important; --n-border-focus: 1px solid var(--gold) !important; --n-border-hover: 1px solid var(--gold) !important; --n-box-shadow-focus: 0 0 0 2px rgba(184, 148, 76, 0.15) !important; }
:deep(.n-base-selection) { --n-color: transparent !important; --n-color-active: transparent !important; --n-text-color: var(--text) !important; --n-border: 1px solid var(--card-border) !important; --n-border-focus: 1px solid var(--gold) !important; --n-border-hover: 1px solid var(--gold) !important; }
:deep(.n-base-select-menu) { --n-color: var(--card) !important; }
:deep(.n-base-selection-label) { --n-text-color: var(--text) !important; background: transparent !important; }
:deep(.n-base-selection-tag) { --n-color: var(--tag-bg) !important; }
:deep(.n-dynamic-tags .n-input) { --n-color: transparent !important; --n-color-focus: transparent !important; }
</style>
