import { createRouter, createWebHistory } from 'vue-router'
import { getToken } from '../api/client'

const routes = [
  // === 公开页面 ===
  {
    path: '/',
    component: () => import('../layouts/PublicLayout.vue'),
    children: [
      { path: '', name: 'Home', component: () => import('../views/HomeView.vue') },
      { path: 'posts/:slug', name: 'Post', component: () => import('../views/PostView.vue') },
      { path: 'articles', name: 'Articles', component: () => import('../views/ArticlesView.vue') },
      { path: 'notes', name: 'Notes', component: () => import('../views/NotesView.vue') },
      { path: 'notes/:id', name: 'NoteDetail', component: () => import('../views/NoteDetail.vue') },
      { path: 'search', name: 'Search', component: () => import('../views/SearchView.vue') },
      { path: 'category', name: 'Categories', component: () => import('../views/CategoryView.vue') },
      { path: 'category/:slug', name: 'Category', component: () => import('../views/CategoryView.vue') },
    ],
  },

  // === 管理后台 ===
  {
    path: '/admin',
    component: () => import('../layouts/AdminLayout.vue'),
    children: [
      { path: '', name: 'Dashboard', component: () => import('../views/admin/Dashboard.vue') },
      { path: 'articles', name: 'AdminArticles', component: () => import('../views/admin/ArticleList.vue') },
      { path: 'articles/new', name: 'AdminNewArticle', component: () => import('../views/admin/ArticleEditor.vue') },
      { path: 'articles/:id/edit', name: 'AdminEditArticle', component: () => import('../views/admin/ArticleEditor.vue'), props: true },
      { path: 'notes', name: 'AdminNotes', component: () => import('../views/admin/NoteList.vue') },
      { path: 'notes/new', name: 'AdminNewNote', component: () => import('../views/admin/NoteEditor.vue') },
      { path: 'notes/:id/edit', name: 'AdminEditNote', component: () => import('../views/admin/NoteEditor.vue'), props: true },
      { path: 'categories', name: 'AdminCategories', component: () => import('../views/admin/Categories.vue') },
      { path: 'tags', name: 'AdminTags', component: () => import('../views/admin/Tags.vue') },
      { path: 'visitors', name: 'AdminVisitors', component: () => import('../views/admin/Visitors.vue') },
      { path: 'danmakus', name: 'AdminDanmakus', component: () => import('../views/admin/DanmakuList.vue') },
      { path: 'settings', name: 'AdminSettings', component: () => import('../views/admin/Settings.vue') },
    ],
  },

  // 登录页（无布局，独立全屏）
  { path: '/login', name: 'Login', component: () => import('../views/LoginView.vue') },

  { path: '/:pathMatch(.*)*', name: 'NotFound', component: () => import('../views/NotFound.vue') },
]

const router = createRouter({ history: createWebHistory(), routes })

// 管理后台登录守卫：无 token 不进入后台界面（写操作另有后端 401 兜底）
router.beforeEach(to => {
  if (to.path.startsWith('/admin') && !getToken()) {
    return { path: '/login', query: { redirect: to.fullPath } }
  }
})

export default router
