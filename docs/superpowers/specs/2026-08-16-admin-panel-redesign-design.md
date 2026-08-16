# 后台管理重构设计

- 日期：2026-08-16
- 状态：已与用户确认
- 分支基线：feature/fairy-chat-frontend
- 范围：`blog/web/admin/`（管理后台全部页面）+ 后端两处只读接口扩展

## 背景与目标

后台现状存在三类问题：

1. **评论管理错位**：全局「评论」页（`CommentList.vue`）靠遍历所有文章和随笔逐个拉取 `/api/notes/:id/comments` 合并展示，其中拿文章 ID 查随笔评论的循环是无效代码（文章没有评论，ID 碰撞时还会错乱显示）；管理员无法按随笔维度管理评论。
2. **两个编辑器不同构**：`ArticleEditor` 有顶栏删除按钮、MD 双栏预览；`NoteEditor` 是极简发布卡片，缺删除按钮，无评论管理，操作框架与文章编辑器不一致。
3. **整体设计非主流**：顶部水平菜单 + 各页面宽窄不一（400px～1100px）、样式各异，部分页面（弹幕/访客）还是简单行列表。

目标：评论并入随笔编辑器按条分开管理；两个编辑器操作框架统一；后台整体升级为侧边栏布局的主流设计，保留水墨金品牌风格。

## 已确认的关键决策

| 决策点 | 结论 |
|---|---|
| 编辑器统一程度 | 统一**操作框架**（顶栏/操作/视觉），各自保留内容形态（文章：标题+分类标签+MD；随笔：500 字短文+图片） |
| 全局评论页 | **移除**（删 `CommentList.vue`、路由、菜单项），评论只在随笔编辑页内管理 |
| 评论管理功能 | 仅**查看 + 删除**，不做后台回复 |
| 布局方向 | **侧边栏 + 水墨金**（主流骨架 + 现有宣纸底色与金色点缀） |
| 改造范围 | **全部后台页面**深做 |
| 随笔编辑器布局 | **左右结构**：左编辑卡、右评论面板，容器放宽至与文章编辑器同宽（1100px） |

## 1. 布局骨架 — `AdminLayout.vue`

- 左侧固定侧边栏 220px，深墨底色（`#26211a` 系），顶部 Logo「✽ 笔墨」，金色 `#b8944c` 高亮当前项
- 垂直 `n-menu`，菜单项：仪表盘 / 文章 / 随笔 / 弹幕 / 分类 / 标签 / 访客 / 设置（**移除「评论」**）
- 侧栏底部：主题切换 + 「← 博客」返回前台
- 主内容区保留粒子背景与宣纸底色；内容宽度由各页面自定
- 响应式：≤900px 侧栏收窄为纯图标栏（不做抽屉）
- 侧栏在明/暗主题下均保持深墨色；内容区沿用现有 `--bg/--card/--text` CSS 变量

## 2. `NoteEditor.vue` 重构（核心）

- 顶栏与文章编辑器同构：`返回 | 字数 n/500 | 删除（popconfirm，仅编辑态）| 发布`（补上缺失的删除按钮）
- 左栏：现有编辑卡原样保留（头像+作者、500 字 textarea、图片九宫格、+ 添加图片）
- 右栏（固定约 380px）评论面板：
  - 标题「评论 · N」；行结构：头像 / 昵称 / 时间 / 内容 / 删除（popconfirm）
  - 面板内部独立滚动，高度与左栏编辑区对齐
  - 空态「还没有评论」；新建模式（无 id）显示占位「发布随笔后即可在此管理评论」
- 数据流：`onMounted` 时 `Promise.all` 拉 `GET /api/notes/:id` 与 `GET /api/notes/:id/comments`；删除评论调 `DELETE /api/admin/comments/:id` 后本地移除
- 评论部分后端零改动（三个接口均已存在）

## 3. `ArticleEditor.vue` 统一化

- 结构与功能不动（标题 / 分类标签元信息行 / MD 双栏 / 删除 / 发布）
- 与 NoteEditor 共用同一套顶栏样式与间距规范，视觉细节对齐（字号、边框、按钮尺寸）
- 已知限制（不在本次范围）：MD 预览为简陋正则替换，不支持列表/链接/代码块等

## 4. 列表页统一 + 后端 `comment_count`

- `NoteList.vue` 表格化（与 `ArticleList` 同构）：`n-data-table`，列：内容摘要（纯文本截断）/ 图片缩略 / 评论数（badge，点击跳编辑页）/ 状态 / 日期 / 编辑·删除；保留 selection 批量删除
- 后端改动一：`GET /api/notes` 列表响应增加 `comment_count`（GORM 子查询 COUNT）
- `ArticleList.vue`：保持表格，样式规整
- 删除 `CommentList.vue`，移除路由 `/admin/comments`

## 5. 其余页面深做

- **Dashboard**：统计卡升级为卡片容器（金 icon + 数字 + 标签），新增「评论」「访客」两张卡 → 后端改动二：`GET /api/stats` 增加 `comment_count`/`visitor_count`；快捷按钮与「最近文章/随笔」列卡片化
- **DanmakuList**：改 `n-data-table`（昵称 / 内容 / 颜色点 / 时间 / 删除）+ 统一页头
- **Visitors**：改 `n-data-table`（头像+昵称 / 签名 / 注册时间 / 编辑·删除）；编辑弹窗保留、样式规整
- **Categories / Tags**：卡片容器 + 统一页头，交互不变
- **Settings**：表单卡片化，头像裁剪逻辑原样保留
- 统一「页标题 + 主操作」页头规范（`.page-head`）与内容宽度：内容型页面 ~720px、表格型页面 ~960px

## 6. 错误处理

- 现状大量 `catch(e){}` 静默吞错 → 核心页面（两个编辑器、两个列表、评论删除）统一改用 `n-message` 提示失败

## 7. 测试与验证

- 后端 Go 测试：`comment_count`（建随笔 + 2 条评论 → ListNotes 断言）、stats 新字段（`internal/handler` 现有测试基建）
- 前端：`npm run build` 通过
- `/run-blog` 冒烟：现有 10 项断言全过；`smoke.mjs` 新增一条断言——创建评论后 `GET /api/notes` 返回的 `comment_count` 为 1

## 改动面清单

| 层 | 文件 | 动作 |
|---|---|---|
| 前端 | `src/layouts/AdminLayout.vue` | 侧边栏重构 |
| 前端 | `src/views/admin/NoteEditor.vue` | 重构（顶栏 + 左编辑右评论） |
| 前端 | `src/views/admin/ArticleEditor.vue` | 样式统一 |
| 前端 | `src/views/admin/NoteList.vue` | 表格化 |
| 前端 | `src/views/admin/ArticleList.vue` | 样式规整 |
| 前端 | `src/views/admin/CommentList.vue` | **删除** |
| 前端 | `src/views/admin/Dashboard.vue` | 统计卡 + 卡片化 |
| 前端 | `src/views/admin/DanmakuList.vue` | 表格化 |
| 前端 | `src/views/admin/Visitors.vue` | 表格化 |
| 前端 | `src/views/admin/Categories.vue`、`Tags.vue`、`Settings.vue` | 样式规整 |
| 前端 | `src/router/index.js` | 移除 `/admin/comments` |
| 后端 | `internal/handler/admin.go`（ListNotes） | `comment_count` |
| 后端 | `internal/handler/admin.go`（Stats） | `comment_count`/`visitor_count` |
| 工具 | `.claude/skills/run-blog/smoke.mjs` | 新增 comment_count 断言 |
