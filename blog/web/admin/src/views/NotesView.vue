<template>
  <div class="page">
    <div class="header">
      <h2 class="title">随笔 · 说说</h2>
      <p class="sub">一切随心，随手记下</p>
    </div>
    <div v-if="!notes.length" class="empty"><n-empty description="还没有随笔" /></div>
    <div v-else class="board">
      <div v-for="(n, i) in pinned" :key="n.id" class="note-wrap" :style="{ animationDelay: (i * 70) + 'ms' }">
        <router-link :to="'/notes/'+n.id" class="note" :style="noteStyle(n.id, i)">
          <div class="pin" />
          <div class="tape" :class="'tape-'+((i*7+3)%4)" />
          <div v-if="imgs(n.images).length" class="note-imgs" :class="'note-imgs--'+Math.min(imgs(n.images).length, 4)">
            <div v-for="(u, j) in imgs(n.images).slice(0, 4)" :key="j" class="img-cell">
              <img :src="u" loading="lazy" @error="e=>e.target.parentNode.style.display='none'" />
              <span v-if="j === 3 && imgs(n.images).length > 4" class="img-more">+{{ imgs(n.images).length - 4 }}</span>
            </div>
          </div>
          <div class="note-body" v-html="n.html" />
          <div class="note-foot">
            <time class="note-time">{{ rel(n.created_at) }}</time>
            <span class="note-more">阅读全文 ›</span>
          </div>
        </router-link>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, computed, onMounted, onUnmounted } from 'vue'
import axios from 'axios'

const notes = ref([])

// 主题切换时重算纸色：暗色下软木板换深色纸，避免亮便签刺眼
const darkMode = ref(document.documentElement.classList.contains('dark'))
let themeObs = null
onMounted(async () => {
  themeObs = new MutationObserver(() => {
    darkMode.value = document.documentElement.classList.contains('dark')
  })
  themeObs.observe(document.documentElement, { attributes: true, attributeFilter: ['class'] })
  try { const { data } = await axios.get('/api/notes?per_page=100'); notes.value = data.notes || [] } catch {}
})
onUnmounted(() => themeObs?.disconnect())

const pinned = computed(() => notes.value.slice(0, 12))

const paperColors = [
  { bg: '#fef9e7', edge: '#f0e5c5', ink: '#5c4a3a' },
  { bg: '#fdf2f2', edge: '#f0d8d8', ink: '#5a3a3a' },
  { bg: '#f2f7fd', edge: '#d8e5f0', ink: '#3a4a5a' },
  { bg: '#f5faf3', edge: '#e0ece0', ink: '#3a5a3a' },
  { bg: '#fef8f0', edge: '#f0e0d0', ink: '#5a4a3a' },
  { bg: '#faf5fe', edge: '#e8d8f0', ink: '#4a3a5a' },
  { bg: '#fdfaf0', edge: '#f0ecd8', ink: '#5a5a3a' },
  { bg: '#f5f5f0', edge: '#e8e5d8', ink: '#4a4a3a' },
]

// 暗色纸系：深纸、柔边、浅墨
const paperColorsDark = [
  { bg: '#3a342a', edge: '#4c4334', ink: '#d8cba8' },
  { bg: '#3c2f2f', edge: '#4e3d3d', ink: '#d8bfb0' },
  { bg: '#2f353c', edge: '#3d454e', ink: '#bcc8d8' },
  { bg: '#2f3a32', edge: '#3d4c40', ink: '#b8d4c2' },
  { bg: '#3a332a', edge: '#4c4234', ink: '#d4c4a4' },
  { bg: '#34303c', edge: '#443e50', ink: '#ccc0dc' },
]

function hash(s, i) {
  let r = i * 17
  for (let c of s) r = (r << 5) - r + c.charCodeAt(0) | 0
  return Math.abs(r)
}

function noteStyle(id, i) {
  const colors = darkMode.value ? paperColorsDark : paperColors
  const c = colors[i % colors.length]
  const h = hash(String(id), i)
  const rot = (h % 7) - 3
  const h2 = (h * 13 + 7) % 20
  const minH = 160 + h2 * 5
  return {
    background: c.bg,
    borderColor: c.edge,
    color: c.ink,
    transform: `rotate(${rot}deg)`,
    minHeight: minH + 'px',
  }
}

function imgs(s) { return s ? s.split(',').map(x => x.trim()).filter(Boolean) : [] }

function rel(d) {
  const s = Math.floor((Date.now() - new Date(d).getTime()) / 1000)
  if (s < 60) return '刚刚'
  if (s < 3600) return Math.floor(s / 60) + '分钟前'
  if (s < 86400) return Math.floor(s / 3600) + '小时前'
  if (s < 2592000) return Math.floor(s / 86400) + '天前'
  return new Date(d).toLocaleDateString('zh-CN', { month: 'short', day: 'numeric' })
}
</script>

<style scoped>
.page { height: 100%; display: flex; flex-direction: column; padding: 0; }
.header { text-align: center; flex-shrink: 0; padding: clamp(28px, 5vh, 40px) 0 16px; }
.title { font-size: clamp(20px, 2vw, 26px); font-weight: 700; color: var(--text); margin: 0 0 4px; }
.sub { font-size: clamp(11px, .9vw, 13px); color: var(--muted); margin: 0; }
.empty { flex: 1; display: flex; align-items: center; justify-content: center; }

/* 软木板效果 */
.board {
  flex: 1;
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(230px, 1fr));
  gap: 56px 44px;
  padding: 32px;
  align-content: start;
}

/* 便签挂位：入场动画与层级放在外包层，避免覆盖便签自身的随机旋转 */
.note-wrap {
  position: relative;
  justify-self: center;
  max-width: 280px; width: 100%;
  animation: noteIn .5s cubic-bezier(.4, 0, .2, 1) both;
}
.note-wrap:hover { z-index: 10; }
@keyframes noteIn { from { opacity: 0; transform: translateY(16px); } to { opacity: 1; transform: translateY(0); } }

/* 便签 */
.note {
  position: relative;
  width: 100%;
  border: 2px solid;
  border-radius: 2px 2px 4px 4px;
  text-decoration: none;
  box-shadow: 2px 3px 8px rgba(0, 0, 0, .08), 0 1px 3px rgba(0, 0, 0, .05);
  transition: transform .3s cubic-bezier(.4, 0, .2, 1), box-shadow .3s;
  padding: 20px 14px 14px;
  display: flex; flex-direction: column;
  overflow: visible;
}
.note:hover {
  transform: scale(1.06) rotate(0deg) !important;
  box-shadow: 4px 8px 24px rgba(0, 0, 0, .15);
}
.note:active { transform: scale(1.02) rotate(0deg) !important; }
.note:focus-visible { outline: 2px solid var(--gold); outline-offset: 3px; }

/* 图钉 */
.pin {
  position: absolute;
  top: -8px; left: 50%; transform: translateX(-50%);
  width: 14px; height: 14px;
  border-radius: 50%;
  background: radial-gradient(circle at 35% 35%, #f0d0a0, #c89040 50%, #8a6020);
  box-shadow: 0 2px 4px rgba(0, 0, 0, .2), inset 0 1px 0 rgba(255, 255, 255, .3);
  z-index: 2;
}

/* 胶带 */
.tape {
  position: absolute;
  width: 36px; height: 14px;
  background: rgba(255, 255, 255, .45);
  border: 1px solid rgba(0, 0, 0, .04);
  z-index: 1;
}
.tape-0 { top: -6px; left: 16px; transform: rotate(-15deg); }
.tape-1 { top: -5px; right: 20px; transform: rotate(20deg); }
.tape-2 { top: -7px; left: 50%; transform: translateX(-50%) rotate(-8deg); width: 28px; }
.tape-3 { top: -4px; right: 28px; transform: rotate(-25deg); width: 30px; }

/* 多图网格：固定纵横比，加载不跳动 */
.note-imgs {
  display: grid; gap: 3px;
  margin-bottom: 8px;
  flex-shrink: 0;
}
.note-imgs--2, .note-imgs--4 { grid-template-columns: 1fr 1fr; }
.note-imgs--3 { grid-template-columns: repeat(3, 1fr); }
.img-cell {
  position: relative;
  aspect-ratio: 1 / 1;
  overflow: hidden;
  border-radius: 2px;
  background: rgba(0, 0, 0, .04);
}
.note-imgs--1 .img-cell { aspect-ratio: 4 / 3; }
.img-cell img { width: 100%; height: 100%; object-fit: cover; display: block; }
/* 第 4 格叠加剩余张数 */
.img-more {
  position: absolute; inset: 0;
  display: flex; align-items: center; justify-content: center;
  background: rgba(0, 0, 0, .45);
  color: #fff; font-size: 15px; font-weight: 600;
}

.note-body {
  font-size: 13px;
  line-height: 1.65;
  display: -webkit-box;
  -webkit-line-clamp: 6;
  -webkit-box-orient: vertical;
  overflow: hidden;
  margin-bottom: 10px;
  flex: 0 0 auto;
}

/* 底部信息条：左时间、右阅读引导 */
.note-foot {
  margin-top: auto;
  display: flex; align-items: center; justify-content: space-between; gap: 8px;
  padding-top: 8px;
  border-top: 1px dashed color-mix(in srgb, currentColor 25%, transparent);
}
.note-time { font-size: 10px; opacity: .65; }
.note-more {
  font-size: 11px; font-weight: 600;
  opacity: 0; transform: translateX(-4px);
  transition: opacity .2s, transform .2s;
}
.note:hover .note-more,
.note:focus-visible .note-more { opacity: .9; transform: none; }
/* 触屏无 hover，常显弱化版 */
@media (hover: none) {
  .note-more { opacity: .55; transform: none; }
}

@media (prefers-reduced-motion: reduce) {
  .note-wrap { animation: none; }
  .note, .note-more { transition: none; }
  .note:hover, .note:active { transform: none; }
}
</style>
