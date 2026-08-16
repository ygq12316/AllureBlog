# 后台管理重构实现计划

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.
>
> **本项目特别约定：不派子代理，主会话用 executing-plans 内联执行**（用户要求，网关不稳定时子代理易失败）。

**Goal:** 评论并入随笔编辑器分开管理；两个编辑器操作框架统一；后台整体升级为「侧边栏 + 水墨金」的主流设计。

**Architecture:** 后端仅两处只读扩展（NoteRepo 子查询 `comment_count`、Stats 聚合计数），评论管理复用既有三接口零改动。前端重构 `AdminLayout` 为侧边栏骨架，11 个后台页面统一页头/面板/表格规范，`CommentList.vue` 删除。

**Tech Stack:** Go (Gin + GORM + glebarez SQLite)、Vue 3 (Naive UI + Vite)、Node 内置 fetch/WebSocket 冒烟。

**设计文档:** `docs/superpowers/specs/2026-08-16-admin-panel-redesign-design.md`

---

## 执行环境注意（Windows Git Bash 实测）

- `make`、`python` 不在 PATH。后端命令用 env 前缀纯命令，在 `blog/` 下执行。
- 后端测试必须注入环境变量，否则 handler 包 init 直接 Fatal：
  `JWT_SECRET=test-secret ADMIN_PASSWORD=test-pass go test ./... -count=1`
- 前端命令在 `blog/web/admin/` 下执行。
- 提交信息用中文，每任务一提交。

---

## Task 1: 后端 — 随笔列表 `comment_count`（TDD）

**Files:**
- Modify: `blog/internal/model/model.go:33-40`（Note 结构体）
- Modify: `blog/internal/repository/note.go:21-36`（ListPublished / ListAll）
- Create: `blog/internal/repository/note_test.go`

- [x] **Step 1: 写失败测试**

创建 `blog/internal/repository/note_test.go`：

```go
package repository

import (
	"testing"

	"blog/internal/model"

	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
)

// newTestDB 内存库：:memory: 下每个连接是独立库，必须强制单连接
func newTestDB(t *testing.T) *gorm.DB {
	t.Helper()
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("打开内存库失败: %v", err)
	}
	sqlDB, _ := db.DB()
	sqlDB.SetMaxOpenConns(1)
	if err := db.AutoMigrate(&model.Note{}, &model.Comment{}); err != nil {
		t.Fatalf("自动迁移失败: %v", err)
	}
	return db
}

func TestNoteRepoListAllCommentCount(t *testing.T) {
	db := newTestDB(t)
	nr := NewNoteRepo(db)
	cr := NewCommentRepo(db)

	n1 := &model.Note{Content: "随笔一", HTML: "随笔一", IsPublished: true}
	n2 := &model.Note{Content: "随笔二", HTML: "随笔二", IsPublished: false}
	if err := nr.Create(n1); err != nil {
		t.Fatal(err)
	}
	if err := nr.Create(n2); err != nil {
		t.Fatal(err)
	}

	for i := 0; i < 2; i++ {
		if err := cr.Create(&model.Comment{NoteID: n1.ID, VisitorUUID: "u1", Content: "评论"}); err != nil {
			t.Fatal(err)
		}
	}

	notes, total, err := nr.ListAll(1, 50)
	if err != nil {
		t.Fatalf("ListAll: %v", err)
	}
	if total != 2 {
		t.Errorf("total = %d, want 2", total)
	}
	byID := map[uint]int64{}
	for _, n := range notes {
		byID[n.ID] = n.CommentCount
	}
	if byID[n1.ID] != 2 {
		t.Errorf("随笔 %d comment_count = %d, want 2", n1.ID, byID[n1.ID])
	}
	if byID[n2.ID] != 0 {
		t.Errorf("随笔 %d comment_count = %d, want 0", n2.ID, byID[n2.ID])
	}
}

func TestNoteRepoListPublishedCommentCount(t *testing.T) {
	db := newTestDB(t)
	nr := NewNoteRepo(db)
	cr := NewCommentRepo(db)

	pub := &model.Note{Content: "已发布", HTML: "已发布", IsPublished: true}
	draft := &model.Note{Content: "草稿", HTML: "草稿", IsPublished: false}
	if err := nr.Create(pub); err != nil {
		t.Fatal(err)
	}
	if err := nr.Create(draft); err != nil {
		t.Fatal(err)
	}
	if err := cr.Create(&model.Comment{NoteID: pub.ID, VisitorUUID: "u1", Content: "评论"}); err != nil {
		t.Fatal(err)
	}

	notes, total, err := nr.ListPublished(1, 50)
	if err != nil {
		t.Fatalf("ListPublished: %v", err)
	}
	if total != 1 || len(notes) != 1 {
		t.Fatalf("total=%d len=%d, want 1/1（草稿不应出现）", total, len(notes))
	}
	if notes[0].CommentCount != 1 {
		t.Errorf("comment_count = %d, want 1", notes[0].CommentCount)
	}
}
```

- [x] **Step 2: 运行确认失败**

```bash
cd /e/pythonProject/web/blog && JWT_SECRET=test-secret ADMIN_PASSWORD=test-pass go test ./internal/repository/ -run TestNoteRepo -count=1 -v
```

期望：FAIL，编译错误 `n.CommentCount undefined`（字段尚不存在）。

- [x] **Step 3: 加 model 字段**

`blog/internal/model/model.go` 的 Note 结构体末尾（CreatedAt 之后）加一行：

```go
type Note struct {
	ID          uint      `gorm:"primaryKey" json:"id"`
	Content     string    `gorm:"not null" json:"content"`
	HTML        string    `json:"html"`
	Images      string    `json:"images"` // comma-separated URLs, max 9
	IsPublished bool      `gorm:"default:false" json:"is_published"`
	CreatedAt   time.Time `json:"created_at"`
	// CommentCount 由列表查询的子查询填充（-> 只读禁写 + -:migration 跳过建列）
	CommentCount int64 `json:"comment_count" gorm:"->;-:migration"`
}
```

- [x] **Step 4: 改两个列表查询**

`blog/internal/repository/note.go` 的 ListPublished / ListAll 整体替换为：

```go
func (r *NoteRepo) ListPublished(page, pageSize int) ([]model.Note, int64, error) {
	var notes []model.Note
	var total int64
	q := r.db.Model(&model.Note{}).Where("is_published = ?", true)
	q.Count(&total)
	err := q.Select("notes.*, (SELECT COUNT(*) FROM comments WHERE comments.note_id = notes.id) AS comment_count").
		Order("created_at DESC").Offset((page - 1) * pageSize).Limit(pageSize).Find(&notes).Error
	return notes, total, err
}

func (r *NoteRepo) ListAll(page, pageSize int) ([]model.Note, int64, error) {
	var notes []model.Note
	var total int64
	r.db.Model(&model.Note{}).Count(&total)
	err := r.db.Model(&model.Note{}).
		Select("notes.*, (SELECT COUNT(*) FROM comments WHERE comments.note_id = notes.id) AS comment_count").
		Order("created_at DESC").Offset((page - 1) * pageSize).Limit(pageSize).Find(&notes).Error
	return notes, total, err
}
```

- [x] **Step 5: 运行确认通过**

同 Step 2 命令。期望：两个测试 PASS。

- [x] **Step 6: 提交**

```bash
git add blog/internal/model/model.go blog/internal/repository/note.go blog/internal/repository/note_test.go
git commit -m "feat: 随笔列表返回 comment_count（子查询，迁移不建列）"
```

---

## Task 2: 后端 — Stats 增加 `comment_count` / `visitor_count`（TDD）

**Files:**
- Modify: `blog/internal/repository/comment.go`（加 CountAll）
- Modify: `blog/internal/repository/visitor.go`（加 CountAll）
- Modify: `blog/internal/handler/admin.go:452-461`（Stats）
- Create: `blog/internal/repository/stats_count_test.go`

- [x] **Step 1: 写失败测试**

创建 `blog/internal/repository/stats_count_test.go`：

```go
package repository

import (
	"testing"

	"blog/internal/model"
)

func TestCommentRepoCountAll(t *testing.T) {
	db := newTestDB(t) // note_test.go 中的助手（同包共享）
	cr := NewCommentRepo(db)

	n := &model.Note{Content: "随笔", HTML: "随笔", IsPublished: true}
	if err := NewNoteRepo(db).Create(n); err != nil {
		t.Fatal(err)
	}
	for i := 0; i < 3; i++ {
		if err := cr.Create(&model.Comment{NoteID: n.ID, VisitorUUID: "u1", Content: "评论"}); err != nil {
			t.Fatal(err)
		}
	}
	got, err := cr.CountAll()
	if err != nil {
		t.Fatalf("CountAll: %v", err)
	}
	if got != 3 {
		t.Errorf("CountAll = %d, want 3", got)
	}
}

func TestVisitorRepoCountAll(t *testing.T) {
	db, err := gormOpenMem()
	if err != nil {
		t.Fatal(err)
	}
	if err := db.AutoMigrate(&model.Visitor{}); err != nil {
		t.Fatal(err)
	}
	vr := NewVisitorRepo(db)
	for i := 0; i < 2; i++ {
		if err := vr.Create(&model.Visitor{UUID: "u" + string(rune('1'+i)), Nickname: "访客", AvatarStyle: "lorelei"}); err != nil {
			t.Fatal(err)
		}
	}
	got, err := vr.CountAll()
	if err != nil {
		t.Fatalf("CountAll: %v", err)
	}
	if got != 2 {
		t.Errorf("CountAll = %d, want 2", got)
	}
}
```

同时把 Task 1 里 `newTestDB` 拆出的裸连接助手加进 `note_test.go`（`gormOpenMem` 供本测试用）：

```go
// gormOpenMem 打开单连接内存库（不迁移）
func gormOpenMem() (*gorm.DB, error) {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		return nil, err
	}
	sqlDB, _ := db.DB()
	sqlDB.SetMaxOpenConns(1)
	return db, nil
}
```

并将 `newTestDB` 改为复用它：

```go
func newTestDB(t *testing.T) *gorm.DB {
	t.Helper()
	db, err := gormOpenMem()
	if err != nil {
		t.Fatalf("打开内存库失败: %v", err)
	}
	if err := db.AutoMigrate(&model.Note{}, &model.Comment{}); err != nil {
		t.Fatalf("自动迁移失败: %v", err)
	}
	return db
}
```

注：`model.Visitor` 字段以 `blog/internal/model/visitor.go` 实际定义为准（UUID/Nickname/AvatarStyle 是公开端注册一直在写的字段）。

- [x] **Step 2: 运行确认失败**

```bash
cd /e/pythonProject/web/blog && JWT_SECRET=test-secret ADMIN_PASSWORD=test-pass go test ./internal/repository/ -run 'CountAll' -count=1 -v
```

期望：FAIL，编译错误 `cr.CountAll undefined` / `vr.CountAll undefined`。

- [x] **Step 3: 实现两个 CountAll**

`blog/internal/repository/comment.go` 末尾追加：

```go
func (r *CommentRepo) CountAll() (int64, error) {
	var count int64
	err := r.db.Model(&model.Comment{}).Count(&count).Error
	return count, err
}
```

`blog/internal/repository/visitor.go` 末尾追加：

```go
func (r *VisitorRepo) CountAll() (int64, error) {
	var count int64
	err := r.db.Model(&model.Visitor{}).Count(&count).Error
	return count, err
}
```

- [x] **Step 4: Stats handler 接入**

`blog/internal/handler/admin.go` 的 `Stats` 整体替换为：

```go
func (h *AdminHandler) Stats(c *gin.Context) {
	_, articleTotal, _ := h.articleSvc.ListAll(1, 1)
	_, noteTotal, _ := h.noteSvc.ListAll(1, 1)
	categories, _ := h.categoryRepo.ListAll()
	commentTotal, _ := h.commentRepo.CountAll()
	visitorTotal, _ := h.visitorRepo.CountAll()
	c.JSON(http.StatusOK, gin.H{
		"article_count":  articleTotal,
		"note_count":     noteTotal,
		"category_count": len(categories),
		"comment_count":  commentTotal,
		"visitor_count":  visitorTotal,
	})
}
```

- [x] **Step 5: 全量后端测试 + vet**

```bash
cd /e/pythonProject/web/blog && JWT_SECRET=test-secret ADMIN_PASSWORD=test-pass go test ./... -count=1 && JWT_SECRET=test-secret ADMIN_PASSWORD=test-pass go vet ./...
```

期望：全部 PASS，vet 无输出。

- [x] **Step 6: 提交**

```bash
git add blog/internal/repository/comment.go blog/internal/repository/visitor.go blog/internal/repository/note_test.go blog/internal/repository/stats_count_test.go blog/internal/handler/admin.go
git commit -m "feat: /api/stats 增加 comment_count 与 visitor_count"
```

---

## Task 3: 前端 — 全局规范样式 + AdminLayout 侧边栏

**Files:**
- Modify: `blog/web/admin/index.html`（追加全局类）
- Modify: `blog/web/admin/src/App.vue`（挂 n-message-provider，供 useMessage 使用）
- Modify: `blog/web/admin/src/layouts/AdminLayout.vue`（整文件重写）

- [x] **Step 0: App.vue 挂载 n-message-provider**

后续编辑器/列表页用 `useMessage` 做失败提示，Naive UI 要求外层必须有 provider。`App.vue` 整文件替换为：

```vue
<template>
  <n-config-provider :theme-overrides="themeOverrides">
    <n-message-provider>
      <router-view />
    </n-message-provider>
  </n-config-provider>
</template>

<script setup>
const themeOverrides = {
  common: { primaryColor: '#b8944c', primaryColorHover: '#d4b060', borderRadius: '4px' },
}
</script>
```

- [x] **Step 1: index.html 追加后台规范类**

在 `index.html` 的 `<style>` 块末尾（`.n-data-table .n-data-table-wrapper` 规则之后）追加：

```css
/* === 后台统一规范 === */
/* 页头：标题 + 主操作 */
.page-head { display: flex; justify-content: space-between; align-items: center; gap: 12px; flex-wrap: wrap; margin-bottom: 20px; }
.page-head h2 { font-size: 18px; font-weight: 700; color: var(--text); margin: 0; }
.page-head-actions { display: flex; gap: 8px; align-items: center; }
/* 内容面板卡片 */
.panel { background: var(--card); border: 1px solid var(--card-border); border-radius: 8px; padding: 20px; }
.panel .n-data-table { --n-merged-th-color: var(--card); --n-merged-td-color: var(--card); }
.panel .n-data-table .n-data-table-th { background: var(--card) !important; }
.panel .n-data-table .n-data-table-td { background: var(--card) !important; }
/* 页面容器宽度 */
.page-narrow { max-width: 720px; margin: 0 auto; }
.page-wide { max-width: 960px; margin: 0 auto; }
/* 编辑器顶栏（两个编辑器共用同构） */
.editor-topbar { display: flex; align-items: center; justify-content: space-between; margin-bottom: 24px; gap: 10px; }
.editor-topbar-side { display: flex; align-items: center; gap: 10px; }
```

- [x] **Step 2: 重写 AdminLayout.vue 为侧边栏**

`blog/web/admin/src/layouts/AdminLayout.vue` 整文件替换为：

```vue
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
            <button v-if="!collapsed" class="sider-link" tag="a" @click="goBlog">← 博客</button>
          </div>
        </aside>

        <!-- 折叠开关（贴主内容区左缘） -->
        <button class="collapse-btn" @click="collapsed = !collapsed" :title="collapsed ? '展开菜单' : '收起菜单'">
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

const route = useRoute(), router = useRouter()
const collapsed = ref(false)

const renderIcon = comp => () => h(NIcon, null, { default: () => h(comp) })
const menuOptions = [
  { label: '仪表盘', key: 'dashboard', icon: renderIcon(SpeedometerOutline) },
  { label: '文章', key: 'articles', icon: renderIcon(DocumentTextOutline) },
  { label: '随笔', key: 'notes', icon: renderIcon(ChatbubbleEllipsesOutline) },
  { label: '弹幕', key: 'danmakus', icon: renderIcon(FilmOutline) },
  { label: '分类', key: 'categories', icon: renderIcon(FolderOpenOutline) },
  { label: '标签', key: 'tags', icon: renderIcon(PricetagsOutline) },
  { label: '访客', key: 'visitors', icon: renderIcon(PeopleOutline) },
  { label: '设置', key: 'settings', icon: renderIcon(SettingsOutline) },
]

const activeMenu = computed(() => {
  const p = route.path
  if (p.startsWith('/admin/articles')) return 'articles'
  if (p.startsWith('/admin/notes')) return 'notes'
  if (p.startsWith('/admin/danmakus')) return 'danmakus'
  if (p.startsWith('/admin/categories')) return 'categories'
  if (p.startsWith('/admin/tags')) return 'tags'
  if (p.startsWith('/admin/visitors')) return 'visitors'
  if (p.startsWith('/admin/settings')) return 'settings'
  return 'dashboard'
})
function onMenuClick(key) {
  const m = { dashboard: '/admin', articles: '/admin/articles', notes: '/admin/notes', danmakus: '/admin/danmakus', categories: '/admin/categories', tags: '/admin/tags', visitors: '/admin/visitors', settings: '/admin/settings' }
  if (m[key]) router.push(m[key])
}
function goBlog() { window.location.href = '/' }

const themeOverrides = {
  common: { primaryColor: '#b8944c', primaryColorHover: '#d4b060', borderRadius: '4px' },
}

// 粒子系统（复制自原版）
const canvas = ref(null), glow = ref(null)
let ctx, w, h, particles = [], mouse = { x: -999, y: -999 }, animId
class Particle { constructor() { this.reset(); this.y = Math.random() * h } reset() { this.x = Math.random() * w; this.y = h + 10; this.size = Math.random() * 2.5 + 1; this.speedY = -(Math.random() * .4 + .15); this.speedX = (Math.random() - .5) * .3; this.opacity = Math.random() * .4 + .15; this.hue = Math.random() > .5 ? '184,148,76' : '120,105,81' } update() { this.x += this.speedX; this.y += this.speedY; const dx = this.x - mouse.x, dy = this.y - mouse.y, dist = Math.sqrt(dx * dx + dy * dy); if (dist < 120) { const f = (120 - dist) / 120 * 1.5; this.x += (dx / dist) * f; this.y += (dy / dist) * f }; if (this.y < -10 || this.x < -10 || this.x > w + 10) this.reset() } draw() { ctx.beginPath(); ctx.arc(this.x, this.y, this.size, 0, Math.PI * 2); ctx.fillStyle = `rgba(${this.hue},${this.opacity})`; ctx.fill() } }
function initC() { if (!canvas.value) return; ctx = canvas.value.getContext('2d'); resize(); particles = Array.from({ length: 55 }, () => new Particle()); animate() }
function resize() { w = window.innerWidth; h = window.innerHeight; canvas.value.width = w; canvas.value.height = h }
function animate() { ctx.clearRect(0, 0, w, h); particles.forEach(p => { p.update(); p.draw() }); animId = requestAnimationFrame(animate) }
function onMM(e) { mouse.x = e.clientX; mouse.y = e.clientY; if (glow.value) { glow.value.style.opacity = '1'; glow.value.style.transform = `translate3d(${e.clientX - 200}px,${e.clientY - 200}px,0)` } }
function onML() { mouse.x = -999; mouse.y = -999; if (glow.value) glow.value.style.opacity = '0' }
// ≤900px 自动收窄为图标栏
function onResize() { collapsed.value = window.innerWidth <= 900 }
onMounted(() => { initC(); onResize(); window.addEventListener('resize', resize); window.addEventListener('resize', onResize); window.addEventListener('mousemove', onMM); document.addEventListener('mouseleave', onML) })
onUnmounted(() => { cancelAnimationFrame(animId); window.removeEventListener('resize', resize); window.removeEventListener('resize', onResize); window.removeEventListener('mousemove', onMM); document.removeEventListener('mouseleave', onML) })
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
.sider.collapsed ~ .collapse-btn { left: 72px }

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
```

- [x] **Step 3: 构建验证**

```bash
cd /e/pythonProject/web/blog/web/admin && npm run build
```

期望：`✓ built` 无错误（路由里 CommentList 仍存在，本任务不动它）。

- [x] **Step 4: 提交**

```bash
git add blog/web/admin/index.html blog/web/admin/src/App.vue blog/web/admin/src/layouts/AdminLayout.vue
git commit -m "feat: 后台改为侧边栏布局（水墨金），新增统一页头/面板规范样式"
```

---

## Task 4: 前端 — 删除全局评论页

**Files:**
- Delete: `blog/web/admin/src/views/admin/CommentList.vue`
- Modify: `blog/web/admin/src/router/index.js:35`

- [x] **Step 1: 移除路由**

`blog/web/admin/src/router/index.js` 删除第 35 行：

```js
      { path: 'comments', name: 'AdminComments', component: () => import('../views/admin/CommentList.vue') },
```

- [x] **Step 2: 删除组件文件**

```bash
git rm blog/web/admin/src/views/admin/CommentList.vue
```

- [x] **Step 3: 全仓确认无残留引用**

```bash
grep -rn "CommentList\|AdminComments\|admin/comments" blog/web/admin/src/
```

期望：无输出。（AdminLayout 菜单已在 Task 3 去掉 comments 项。）

- [x] **Step 4: 构建 + 提交**

```bash
cd /e/pythonProject/web/blog/web/admin && npm run build
git add blog/web/admin/src/router/index.js
git commit -m "refactor: 移除全局评论页，评论改由随笔编辑页分条管理"
```

---

## Task 5: 前端 — NoteEditor 重构（顶栏 + 左编辑右评论）

**Files:**
- Modify: `blog/web/admin/src/views/admin/NoteEditor.vue`（整文件重写）

- [x] **Step 1: 整文件替换**

```vue
<template>
  <div class="wrap">
    <!-- 顶栏：与文章编辑器同构 -->
    <div class="editor-topbar">
      <n-button text size="small" @click="$router.push('/admin/notes')">
        <n-icon size="14" :component="ArrowBackIcon" /> 返回
      </n-button>
      <span class="count" :class="{ 'count-warn': content.length > 400 }">{{ content.length }}/500</span>
      <div class="editor-topbar-side">
        <n-popconfirm v-if="isEdit" @positive-click="del" positive-text="确认删除" negative-text="取消">
          <template #trigger><n-button text size="small" type="error">删除</n-button></template>
          确定删除这条随笔？其下评论会一并失去归属。此操作不可恢复。
        </n-popconfirm>
        <n-button type="primary" @click="publish" :disabled="!canPublish" :loading="saving">发布</n-button>
      </div>
    </div>

    <div class="cols">
      <!-- 左栏：编辑卡（内容形态保留） -->
      <div class="card">
        <div class="card-top">
          <div class="avatar"><n-icon size="16" :component="PersonIcon" /></div>
          <span class="author">{{ authorName }}</span>
        </div>
        <textarea
          v-model="content"
          class="content-area"
          placeholder="写点什么..."
          maxlength="500"
          rows="8"
          autofocus
        />
        <div v-if="images.length" class="previews">
          <div v-for="(img, i) in images" :key="i" class="pv">
            <img :src="img" />
            <n-button @click="remove(i)" circle size="tiny" class="pv-rm">✕</n-button>
          </div>
        </div>
        <div class="card-foot">
          <n-upload
            v-if="images.length < 9"
            :show-file-list="false"
            :custom-request="upload"
            accept="image/*"
          >
            <n-button size="small" text>+ 添加图片</n-button>
          </n-upload>
        </div>
      </div>

      <!-- 右栏：评论面板（仅编辑态可用） -->
      <div class="panel comments-panel">
        <template v-if="isEdit">
          <div class="comments-head">
            <span class="comments-title">评论 · {{ comments.length }}</span>
            <span class="comments-sub">全部来自这条随笔</span>
          </div>
          <div class="comments-list">
            <div v-if="!comments.length" class="empty">还没有评论</div>
            <div v-for="c in comments" :key="c.id" class="comment-row">
              <img :src="avt(c)" class="c-avt" />
              <div class="c-body">
                <div class="c-meta">
                  <span class="c-nick">{{ c.nickname || '匿名' }}</span>
                  <span class="c-time">{{ c.created_at?.slice(0, 16) }}</span>
                </div>
                <p class="c-content">{{ c.content }}</p>
              </div>
              <n-popconfirm @positive-click="delComment(c.id)" positive-text="删除" negative-text="取消">
                <template #trigger><n-button size="tiny" text type="error">删除</n-button></template>
                确定删除这条评论？
              </n-popconfirm>
            </div>
          </div>
        </template>
        <div v-else class="empty comments-placeholder">发布随笔后<br />即可在此管理评论</div>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, computed, onMounted } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { PersonOutline, ArrowBackOutline } from '@vicons/ionicons5'
import { useMessage } from 'naive-ui'
import axios from 'axios'

const PersonIcon = PersonOutline
const ArrowBackIcon = ArrowBackOutline
const route = useRoute()
const router = useRouter()
const message = useMessage()
const content = ref('')
const images = ref([])
const authorName = ref('Allure')
const comments = ref([])
const saving = ref(false)

const isEdit = computed(() => !!route.params.id)
const canPublish = computed(() => content.value.trim().length > 0 || images.value.length > 0)

function remove(i) { images.value.splice(i, 1) }

function avt(c) {
  if (c.avatar_url) return c.avatar_url
  return 'https://api.dicebear.com/9.x/' + (c.avatar_style || 'lorelei') + '/svg?seed=' + c.visitor_uuid
}

async function upload({ file }) {
  try {
    const f = new FormData()
    f.append('file', file.file)
    const { data } = await axios.post('/api/upload', f)
    images.value.push(data.url)
  } catch (e) {
    message.error('图片上传失败')
  }
}

async function loadComments(id) {
  try {
    const { data } = await axios.get(`/api/notes/${id}/comments`)
    comments.value = data.comments || []
  } catch (e) {
    message.error('评论加载失败')
  }
}

async function publish() {
  saving.value = true
  try {
    const id = route.params.id
    const payload = { content: content.value, images: images.value.join(','), is_published: true }
    if (id) await axios.put(`/api/notes/${id}`, payload)
    else await axios.post('/api/notes', payload)
    router.push('/admin/notes')
  } catch (e) {
    message.error(e.response?.data?.error || '发布失败')
    saving.value = false
  }
}

async function del() {
  const id = route.params.id
  if (!id) return
  try {
    await axios.delete(`/api/notes/${id}`)
    router.push('/admin/notes')
  } catch (e) {
    message.error('删除失败')
  }
}

async function delComment(cid) {
  try {
    await axios.delete(`/api/admin/comments/${cid}`)
    comments.value = comments.value.filter(c => c.id !== cid)
  } catch (e) {
    message.error('评论删除失败')
  }
}

onMounted(async () => {
  try { const { data } = await axios.get('/api/config'); authorName.value = data.config?.author_name || 'Allure' } catch (e) {}
  const id = route.params.id
  if (!id) return
  try {
    const { data } = await axios.get(`/api/notes/${id}`)
    content.value = data.content || ''
    images.value = data.images ? data.images.split(',').filter(Boolean) : []
  } catch (e) {
    message.error('随笔加载失败')
  }
  loadComments(id)
})
</script>

<style scoped>
.wrap { max-width: 1100px; margin: 0 auto; }

.count {
  font-size: 13px;
  color: var(--muted);
  font-family: 'JetBrains Mono', monospace;
}
.count-warn { color: #c97a4a; }

/* 左右两栏 */
.cols { display: flex; gap: 20px; align-items: stretch; }
@media (max-width: 900px) { .cols { flex-direction: column; } }

/* 左栏：编辑卡 */
.card {
  flex: 1;
  border: 1px solid var(--card-border);
  border-radius: 8px;
  padding: 20px;
  background: var(--card);
  display: flex;
  flex-direction: column;
}
.card-top { display: flex; align-items: center; gap: 10px; margin-bottom: 14px; }
.avatar {
  width: 34px; height: 34px; border-radius: 6px;
  background: var(--tag-bg); border: 1px solid var(--card-border);
  display: flex; align-items: center; justify-content: center;
  flex-shrink: 0; color: var(--gold);
}
.author { font-size: 14px; font-weight: 600; color: var(--text); }
.content-area {
  width: 100%; border: none; background: transparent; resize: none; outline: none;
  font-family: 'LXGW WenKai', serif; font-size: 15px; line-height: 1.8;
  color: var(--text); flex: 1; min-height: 200px; caret-color: var(--gold);
}
.content-area::placeholder { color: var(--muted); opacity: 0.45; }
.previews { display: flex; gap: 8px; flex-wrap: wrap; margin-top: 14px; }
.pv { position: relative; width: 72px; height: 72px; border-radius: 6px; overflow: hidden; border: 1px solid var(--card-border); }
.pv img { width: 100%; height: 100%; object-fit: cover; }
.pv-rm { position: absolute; top: 2px; right: 2px; }
.card-foot { margin-top: 14px; padding-top: 12px; border-top: 1px solid var(--card-border); }

/* 右栏：评论面板 */
.comments-panel { width: 380px; flex-shrink: 0; display: flex; flex-direction: column; max-height: 640px; }
@media (max-width: 900px) { .comments-panel { width: auto; max-height: 360px; } }
.comments-head {
  display: flex; justify-content: space-between; align-items: baseline;
  padding-bottom: 12px; border-bottom: 1px solid var(--card-border); flex-shrink: 0;
}
.comments-title { font-size: 14px; font-weight: 700; color: var(--text); }
.comments-sub { font-size: 11px; color: var(--muted); }
.comments-list { flex: 1; overflow-y: auto; }
.comment-row { display: flex; align-items: flex-start; gap: 10px; padding: 12px 0; border-bottom: 1px solid var(--card-border); }
.comment-row:last-child { border-bottom: none; }
.c-avt { width: 28px; height: 28px; border-radius: 50%; flex-shrink: 0; background: var(--tag-bg); }
.c-body { flex: 1; min-width: 0; }
.c-meta { display: flex; align-items: baseline; gap: 8px; }
.c-nick { font-size: 13px; font-weight: 600; color: var(--gold); }
.c-time { font-size: 10px; color: var(--muted); }
.c-content { font-size: 13px; color: var(--text); margin: 4px 0 0; word-break: break-word; }
.empty { padding: 40px 0; text-align: center; color: var(--muted); font-size: 13px; }
.comments-placeholder { line-height: 2; flex: 1; display: flex; align-items: center; justify-content: center; padding: 0; }
</style>
```

- [x] **Step 2: 构建验证**

```bash
cd /e/pythonProject/web/blog/web/admin && npm run build
```

期望：`✓ built`。

- [x] **Step 3: 提交**

```bash
git add blog/web/admin/src/views/admin/NoteEditor.vue
git commit -m "feat: 随笔编辑器重构 — 顶栏对齐文章编辑器 + 右侧评论面板（查看/删除）"
```

---

## Task 6: 前端 — NoteList 表格化

**Files:**
- Modify: `blog/web/admin/src/views/admin/NoteList.vue`（整文件重写）

- [x] **Step 1: 整文件替换**

```vue
<template>
  <div class="wrap page-wide">
    <div class="page-head">
      <h2>随笔管理</h2>
      <div class="page-head-actions">
        <n-button v-if="selected.length" type="error" size="small" @click="batchDel">删除选中 ({{ selected.length }})</n-button>
        <n-button type="primary" @click="$router.push('/admin/notes/new')">+ 写随笔</n-button>
      </div>
    </div>
    <div class="panel">
      <n-data-table :columns="cols" :data="notes" :bordered="false" size="small"
        :row-key="r => r.id" @update:checked-row-keys="selected = $event" />
    </div>
  </div>
</template>
<script setup>
import { ref, onMounted, h } from 'vue'
import { useRouter } from 'vue-router'
import { NButton, NTag, NPopconfirm, useMessage } from 'naive-ui'
import axios from 'axios'

const router = useRouter()
const message = useMessage()
const notes = ref([]), selected = ref([])

const cols = [
  { type: 'selection', width: 40 },
  {
    title: '内容', key: 'content', width: '*', ellipsis: { tooltip: true },
    render(row) { const t = (row.content || '').replace(/\s+/g, ' ').trim(); return t || '（无文字）' },
  },
  {
    title: '图片', key: 'images', width: 120,
    render(row) {
      const imgs = row.images ? row.images.split(',').filter(Boolean) : []
      if (!imgs.length) return ''
      return h('div', { style: 'display:flex;gap:4px' },
        imgs.slice(0, 3).map(u => h('img', { src: u, style: 'width:36px;height:36px;object-fit:cover;border-radius:4px;border:1px solid var(--card-border)' })))
    },
  },
  {
    title: '评论', key: 'comment_count', width: 70,
    render(row) {
      return h('a', {
        style: 'cursor:pointer;' + (row.comment_count > 0 ? 'color:var(--gold)' : 'color:var(--muted)'),
        onClick: () => router.push(`/admin/notes/${row.id}/edit`),
      }, String(row.comment_count ?? 0))
    },
  },
  { title: '状态', width: 70, render(row) { return h(NTag, { size: 'tiny', type: row.is_published ? 'success' : 'warning', bordered: false }, { default: () => row.is_published ? '已发布' : '草稿' }) } },
  { title: '日期', width: 110, render(row) { return new Date(row.created_at).toLocaleDateString('zh-CN') } },
  {
    title: '', width: 90,
    render(row) {
      return h('div', { style: 'display:flex;gap:2px' }, [
        h(NButton, { size: 'tiny', onClick: () => router.push(`/admin/notes/${row.id}/edit`) }, { default: () => '编辑' }),
        h(NPopconfirm, { onPositiveClick: () => del(row.id) }, {
          trigger: () => h(NButton, { size: 'tiny', text: true, type: 'error' }, { default: () => '删除' }),
          default: () => '确定删除?',
        }),
      ])
    },
  },
]

onMounted(async () => { const { data } = await axios.get('/api/notes?all=true'); notes.value = data.notes || [] })
async function del(id) {
  try {
    await axios.delete(`/api/notes/${id}`)
    notes.value = notes.value.filter(n => n.id !== id)
  } catch (e) { message.error('删除失败') }
}
async function batchDel() {
  try {
    for (const id of selected.value) await axios.delete(`/api/notes/${id}`)
    notes.value = notes.value.filter(n => !selected.value.includes(n.id))
    selected.value = []
  } catch (e) { message.error('批量删除失败'); }
}
</script>
<style scoped>
.wrap { margin: 0 auto; }
</style>
```

- [x] **Step 2: 构建验证 + 提交**

```bash
cd /e/pythonProject/web/blog/web/admin && npm run build
git add blog/web/admin/src/views/admin/NoteList.vue
git commit -m "feat: 随笔列表表格化，含评论数列（点击跳编辑页评论区）"
```

---

## Task 7: 前端 — ArticleEditor / ArticleList 规整

**Files:**
- Modify: `blog/web/admin/src/views/admin/ArticleEditor.vue`（顶栏类与发布加载态）
- Modify: `blog/web/admin/src/views/admin/ArticleList.vue`（页头 + 面板）

- [x] **Step 1: ArticleEditor 顶栏统一**

`ArticleEditor.vue` 模板中 `<div class="topbar">` 改为全局规范类（`.editor-topbar`），发布按钮加 loading，字数计数类名对齐：

```html
    <!-- 顶栏 -->
    <div class="editor-topbar">
      <n-button text size="small" @click="$router.push('/admin/articles')">
        <n-icon size="14" :component="ArrowBackIcon" /> 返回
      </n-button>
      <span class="count">{{ form.content.length }} 字</span>
      <div class="editor-topbar-side">
        <n-popconfirm v-if="isEdit" @positive-click="del" positive-text="确认删除" negative-text="取消">
          <template #trigger><n-button text size="small" type="error">删除</n-button></template>
          确定删除这篇文章？此操作不可恢复。
        </n-popconfirm>
        <n-button type="primary" @click="save" :disabled="!canPublish" :loading="saving">发布</n-button>
      </div>
    </div>
```

script 中补 `const saving = ref(false)`、`import { useMessage } from 'naive-ui'` 与 `const message = useMessage()`；`save()` / `del()` 替换为：

```js
async function save() {
  saving.value = true
  try {
    form.value.is_published = true
    const id = route.params.id
    if (id) await axios.put(`/api/articles/${id}`, form.value)
    else await axios.post('/api/articles', form.value)
    router.push('/admin/articles')
  } catch (e) {
    message.error(e.response?.data?.error || '发布失败')
    saving.value = false
  }
}

async function del() {
  const id = route.params.id
  if (!id) return
  try {
    await axios.delete(`/api/articles/${id}`)
    router.push('/admin/articles')
  } catch (e) {
    message.error('删除失败')
  }
}
```

样式：删除 `.topbar`、`.topbar-actions`、`.word-count` 三条 scoped 样式，替换为：

```css
.count { font-size: 13px; color: var(--muted); font-family: 'JetBrains Mono', monospace; }
```

- [x] **Step 2: ArticleList 页头/面板规整**

`ArticleList.vue` 模板整体替换为：

```vue
<template>
  <div class="wrap page-wide">
    <div class="page-head">
      <h2>文章管理</h2>
      <div class="page-head-actions">
        <n-button v-if="selected.length" type="error" size="small" @click="batchDel">删除选中 ({{ selected.length }})</n-button>
        <n-button type="primary" @click="$router.push('/admin/articles/new')">+ 写文章</n-button>
      </div>
    </div>
    <div class="panel">
      <n-data-table :columns="cols" :data="articles" :bordered="false" size="small"
        :row-key="r => r.id" @update:checked-row-keys="selected = $event" />
    </div>
  </div>
</template>
```

script 部分 cols 里标题列渲染、其余保持原状；`batchDel` 加失败提示（`try/catch` + `useMessage`，与 Task 6 的 NoteList 同构）；scoped 样式整段删除（改用全局规范类）。

- [x] **Step 3: 构建 + 提交**

```bash
cd /e/pythonProject/web/blog/web/admin && npm run build
git add blog/web/admin/src/views/admin/ArticleEditor.vue blog/web/admin/src/views/admin/ArticleList.vue
git commit -m "refactor: 文章编辑器/列表对齐后台统一规范（页头、面板、加载态、失败提示）"
```

---

## Task 8: 前端 — Dashboard 卡片化 + 评论/访客统计

**Files:**
- Modify: `blog/web/admin/src/views/admin/Dashboard.vue`

- [x] **Step 1: 统计卡升级**

模板的统计区 `<div class="stats">` 改为：

```html
    <!-- 统计 -->
    <div class="stats">
      <div class="panel stat" v-for="s in statList" :key="s.label">
        <n-icon size="18" color="var(--gold)" :component="s.icon" />
        <div class="stat-num">{{ s.value }}</div>
        <div class="stat-label">{{ s.label }}</div>
      </div>
    </div>
```

script 的 `statList` 扩展（图标 import 处补 `ChatbubblesOutline`、`PeopleOutline`）：

```js
const statList = computed(() => [
  { icon: DocIcon, value: stats.value.article_count, label: '文章' },
  { icon: ChatIcon, value: stats.value.note_count, label: '随笔' },
  { icon: BubblesIcon, value: stats.value.comment_count ?? 0, label: '评论' },
  { icon: PeopleIcon, value: stats.value.visitor_count ?? 0, label: '访客' },
  { icon: FolderIcon, value: stats.value.category_count || 0, label: '分类' },
  { icon: PricetagIcon, value: stats.value.tag_count || 0, label: '标签' },
])
```

（`BubblesIcon = ChatbubblesOutline`、`PeopleIcon = PeopleOutline`，与现有 `PersonIcon=PersonOutline` 同一行风格声明。`stats` 初值同步加 `comment_count: 0, visitor_count: 0`。）

- [x] **Step 2: 布局与样式调整**

- `.dash` 容器改为 `page-narrow`：`<div class="dash">` 保留，scoped 里 `.dash { max-width: 720px; margin: 0 auto; }`
- `.stats` 改 3×2 网格：`grid-template-columns: repeat(3, 1fr);`
- `.stat` 调整：`text-align: center; display: flex; flex-direction: column; align-items: center; gap: 4px;`
- 「最近」两列卡片包 `.panel`：`.recent .col` 外层各包 `<div class="panel">`，`.col` 内边距去掉

- [x] **Step 3: 构建 + 提交**

```bash
cd /e/pythonProject/web/blog/web/admin && npm run build
git add blog/web/admin/src/views/admin/Dashboard.vue
git commit -m "feat: 仪表盘统计卡片化，新增评论/访客统计"
```

---

## Task 9: 前端 — DanmakuList 表格化

**Files:**
- Modify: `blog/web/admin/src/views/admin/DanmakuList.vue`（整文件重写）

- [x] **Step 1: 整文件替换**

```vue
<template>
  <div class="wrap page-narrow">
    <div class="page-head">
      <h2>弹幕管理</h2>
      <div class="page-head-actions" />
    </div>
    <div class="panel">
      <n-data-table :columns="cols" :data="list" :bordered="false" size="small" :row-key="r => r.id" />
    </div>
  </div>
</template>

<script setup>
import { ref, onMounted, h } from 'vue'
import { NButton, NPopconfirm } from 'naive-ui'
import axios from 'axios'
const list = ref([])

const cols = [
  { title: '昵称', key: 'nickname', width: 110, render(row) { return row.nickname || '匿名' } },
  { title: '内容', key: 'content', width: '*', ellipsis: { tooltip: true } },
  { title: '颜色', key: 'color', width: 60, render(row) { return h('span', { style: `display:inline-block;width:12px;height:12px;border-radius:50%;background:${row.color};vertical-align:middle` }) } },
  { title: '时间', key: 'created_at', width: 140, render(row) { return row.created_at?.slice(0, 16) } },
  {
    title: '', width: 60,
    render(row) {
      return h(NPopconfirm, { onPositiveClick: () => del(row.id) }, {
        trigger: () => h(NButton, { size: 'tiny', text: true, type: 'error' }, { default: () => '删除' }),
        default: () => '确定删除？',
      })
    },
  },
]

async function load() {
  try {
    const resp = await axios.get('/api/danmaku')
    list.value = resp.data.danmaku || []
  } catch (e) {}
}

onMounted(load)

async function del(id) {
  await axios.delete('/api/admin/danmaku/' + id)
  list.value = list.value.filter(d => d.id !== id)
}
</script>
```

- [x] **Step 2: 构建 + 提交**

```bash
cd /e/pythonProject/web/blog/web/admin && npm run build
git add blog/web/admin/src/views/admin/DanmakuList.vue
git commit -m "refactor: 弹幕列表表格化，对齐后台统一规范"
```

---

## Task 10: 前端 — Visitors 表格化（保留编辑弹窗）

**Files:**
- Modify: `blog/web/admin/src/views/admin/Visitors.vue`（模板与列表部分重写，弹窗逻辑不动）

- [x] **Step 1: 模板替换**

```vue
<template>
  <div class="wrap page-wide">
    <div class="page-head">
      <h2>访客管理 ({{ visitors.length }})</h2>
      <div class="page-head-actions" />
    </div>
    <div class="panel">
      <n-data-table :columns="cols" :data="visitors" :bordered="false" size="small" :row-key="r => r.uuid" />
    </div>

    <!-- 编辑弹窗（逻辑保留） -->
    <n-modal :show="editing" @update:show="editing = $event">
      <div class="edit-card">
        <h4>编辑访客</h4>
        <img :src="editForm.avatar" class="edit-avatar" />
        <n-input v-model:value="editForm.nickname" placeholder="昵称" style="margin-bottom:8px" />
        <n-input v-model:value="editForm.signature" placeholder="签名" style="margin-bottom:12px" />
        <n-button type="primary" block @click="saveEdit">保存</n-button>
      </div>
    </n-modal>
  </div>
</template>
```

- [x] **Step 2: script 列定义**

script 里新增列（复用既有 `editVisitor`/`del`/`saveEdit` 函数与数据加载逻辑，不改动）：

```js
import { h } from 'vue'
import { NButton, NPopconfirm } from 'naive-ui'

const cols = [
  {
    title: '访客', key: 'nickname', width: 200,
    render(row) {
      return h('div', { style: 'display:flex;align-items:center;gap:8px' }, [
        h('img', { src: `https://api.dicebear.com/9.x/${row.avatar_style}/svg?seed=${row.uuid}`, style: 'width:28px;height:28px;border-radius:50%;background:var(--tag-bg)' }),
        h('span', { style: 'font-weight:600' }, row.nickname),
      ])
    },
  },
  { title: '签名', key: 'signature', width: '*', ellipsis: { tooltip: true }, render(row) { return row.signature || '—' } },
  { title: '注册时间', key: 'created_at', width: 110, render(row) { return row.created_at?.slice(0, 10) } },
  {
    title: '', width: 90,
    render(row) {
      const btns = [h(NButton, { size: 'tiny', onClick: () => editVisitor(row) }, { default: () => '编辑' })]
      if (!row.uuid.startsWith('admin_')) {
        btns.push(h(NPopconfirm, { onPositiveClick: () => del(row.uuid) }, {
          trigger: () => h(NButton, { size: 'tiny', text: true, type: 'error' }, { default: () => '删除' }),
          default: () => '确定删除？评论和弹幕会保留。',
        }))
      }
      return h('div', { style: 'display:flex;gap:2px' }, btns)
    },
  },
]
```

原模板的行循环与 `.row` 系列样式删除；`.edit-card`/`.edit-avatar` 样式保留。

- [x] **Step 3: 构建 + 提交**

```bash
cd /e/pythonProject/web/blog/web/admin && npm run build
git add blog/web/admin/src/views/admin/Visitors.vue
git commit -m "refactor: 访客列表表格化，对齐后台统一规范"
```

---

## Task 11: 前端 — Categories / Tags / Settings 规整

**Files:**
- Modify: `blog/web/admin/src/views/admin/Categories.vue`
- Modify: `blog/web/admin/src/views/admin/Tags.vue`
- Modify: `blog/web/admin/src/views/admin/Settings.vue`

- [x] **Step 1: Categories 模板替换**

```vue
<template>
  <div class="wrap page-narrow">
    <div class="page-head">
      <h2>分类管理</h2>
      <div class="page-head-actions" />
    </div>
    <div class="panel">
      <div class="add-row"><n-input v-model:value="n" placeholder="分类名称" @keyup.enter="add" /><n-button type="primary" @click="add">添加</n-button></div>
      <n-data-table :columns="cols" :data="list" :bordered="false" size="small" />
    </div>
  </div>
</template>
```

script 不动；scoped 样式只留 `.add-row { display: flex; gap: 8px; margin-bottom: 16px; }`。

- [x] **Step 2: Tags 模板替换**

```vue
<template>
  <div class="wrap page-narrow">
    <div class="page-head">
      <h2>标签管理</h2>
      <div class="page-head-actions" />
    </div>
    <div class="panel">
      <div class="add-row"><n-input v-model:value="n" placeholder="标签名称" @keyup.enter="add" /><n-button type="primary" @click="add">添加</n-button></div>
      <n-space><n-tag v-for="t in list" :key="t.id" closable @close="del(t.id)">{{ t.name }} ({{ t.article_count }})</n-tag></n-space>
    </div>
  </div>
</template>
```

script 不动；scoped 样式只留 `.add-row`。

- [x] **Step 3: Settings 模板骨架替换**

外层结构改为：

```vue
<template>
  <div class="wrap page-narrow">
    <div class="page-head">
      <h2>博客设置</h2>
      <div class="page-head-actions" />
    </div>

    <!-- 裁剪模式（逻辑原样保留） -->
    <div v-if="cropMode" class="panel crop-section">
      ...原 crop-section 内容不变...
    </div>

    <!-- 正常模式 -->
    <div v-else class="panel form">
      ...原 form 内容不变...
    </div>
  </div>
</template>
```

script 与裁剪逻辑零改动；scoped 删除 `.wrap`/`.title`，`.form` 里补 `gap` 布局沿用。

- [x] **Step 4: 构建 + 提交**

```bash
cd /e/pythonProject/web/blog/web/admin && npm run build
git add blog/web/admin/src/views/admin/Categories.vue blog/web/admin/src/views/admin/Tags.vue blog/web/admin/src/views/admin/Settings.vue
git commit -m "refactor: 分类/标签/设置页套用统一页头与面板规范"
```

---

## Task 12: 冒烟断言 + 全链路验证

**Files:**
- Modify: `.claude/skills/run-blog/smoke.mjs`

- [x] **Step 1: 加 comment_count 断言**

在「评论 WebSocket 广播」断言（`ok('评论 WebSocket 广播', ...)`）之后插入：

```js
// ── 5.5 评论计数回读（随笔列表 comment_count）──
const notesAfter = await (await fetch(BASE + '/api/notes')).json();
const counted = (notesAfter.notes || []).find(n => n.id === noteId);
ok('随笔列表 comment_count', typeof counted?.comment_count === 'number' && counted.comment_count >= 1,
   `noteId=${noteId}, count=${counted?.comment_count}`);
```

- [x] **Step 2: 构建前端 + 启动后端**

```bash
cd /e/pythonProject/web/blog/web/admin && npm run build
cd /e/pythonProject/web/blog && JWT_SECRET=dev-secret ADMIN_PASSWORD=dev-pass nohup go run ./cmd/server > /tmp/blog-server.log 2>&1 &
sleep 16 && grep -E "已启动" /tmp/blog-server.log
```

期望日志：`笔墨 · Ink & Code 已启动 → http://localhost:8080`。

- [x] **Step 3: 跑冒烟**

```bash
node /e/pythonProject/web/.claude/skills/run-blog/smoke.mjs
```

期望：全部 PASS（原 10 项 + 新增 `随笔列表 comment_count`），结尾「全部通过」。

- [x] **Step 4: 后台截图人工检查**

```bash
"/c/Program Files (x86)/Microsoft/Edge/Application/msedge.exe" --headless --disable-gpu \
  "--screenshot=C:/Users/Allure/AppData/Local/Temp/admin-notes.png" --window-size=1280,900 \
  "http://localhost:8080/admin/notes"
```

注：路由守卫无 token 会跳登录页，截图管理员页面前需先用脚本登录拿 token 写 localStorage——简化做法：先截图 `/login` 确认样式正常；核心交互（侧边栏/评论面板）请用户在浏览器人工过一遍。

- [x] **Step 5: 全量后端测试收尾**

```bash
cd /e/pythonProject/web/blog && JWT_SECRET=test-secret ADMIN_PASSWORD=test-pass go test ./... -count=1
```

- [x] **Step 6: 提交**

```bash
git add .claude/skills/run-blog/smoke.mjs
git commit -m "test: 冒烟驱动增加随笔列表 comment_count 断言"
```

---

## 完成标准

- [ ] `go test ./...` 全绿
- [ ] `npm run build` 通过
- [ ] smoke.mjs 全部 PASS（含新断言）
- [ ] 后台人工过一遍：侧边栏导航、随笔编辑（评论面板）、文章编辑、各列表页无白屏
