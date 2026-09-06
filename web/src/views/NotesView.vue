<template>
  <div class="flex flex-col h-full">
    <div class="text-center shrink-0 pt-8 pb-4">
      <h2 class="text-xl md:text-2xl font-light tracking-[0.2em] text-ink m-0 mb-1">随笔 · 说说</h2>
      <p class="text-xs text-ink3 m-0">一切随心，随手记下</p>
    </div>
    <div v-if="!notes.length" class="flex-1 flex items-center justify-center"><InkEmpty description="还没有随笔" /></div>
    <div v-else class="board">
      <div v-for="(n, i) in pinned" :key="n.id" class="note-wrap" :style="{ animationDelay: (i * 80) + 'ms' }">
        <router-link :to="'/notes/'+n.id" class="note" :style="noteStyle(n.id, i)">
          <div class="pin" />
          <div class="tape" :class="'tape-'+((i*7+3)%4)" />
          <div v-if="imgs(n.images).length" class="note-imgs" :class="'note-imgs--'+Math.min(imgs(n.images).length, 4)">
            <div v-for="(u, j) in imgs(n.images).slice(0, 4)" :key="j" class="img-cell">
              <img :src="u" loading="lazy" alt="" @error="e=>e.target.parentNode.style.display='none'" />
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
import { listNotes } from '../api/notes'
import { rel } from '../utils/format'
import InkEmpty from '../components/ui/InkEmpty.vue'

const notes = ref([])

// 主题切换时重算纸色：暗色下软木板换深色纸，避免亮便签刺眼
const darkMode = ref(document.documentElement.classList.contains('dark'))
let themeObs = null
onMounted(async () => {
  themeObs = new MutationObserver(() => {
    darkMode.value = document.documentElement.classList.contains('dark')
  })
  themeObs.observe(document.documentElement, { attributes: true, attributeFilter: ['class'] })
  try { const data = await listNotes({ per_page: 100 }); notes.value = data.notes || [] } catch {}
})
onUnmounted(() => themeObs?.disconnect())

const pinned = computed(() => notes.value.slice(0, 12))

// 宣纸色系便签：淡墨、苔青、茶赭、灰紫，浓淡相间
const paperColors = [
  { bg: '#f7f2e8', edge: '#e5dcc8', ink: '#4a4438' },
  { bg: '#eef0e6', edge: '#d5dbc8', ink: '#44503f' },
  { bg: '#efe9dd', edge: '#ddd2bd', ink: '#55483a' },
  { bg: '#e9ede9', edge: '#d2d9d2', ink: '#3f4a42' },
  { bg: '#f4ece0', edge: '#e2d5c2', ink: '#54463a' },
  { bg: '#efe9ec', edge: '#dbd2de', ink: '#4d4252' },
  { bg: '#f7f3e4', edge: '#e6dfc5', ink: '#55503c' },
  { bg: '#f0efeb', edge: '#dedcd4', ink: '#4a4844' },
]

// 暗色纸系：深纸、柔边、浅墨
const paperColorsDark = [
  { bg: '#2e2a24', edge: '#3d372e', ink: '#cfc5b2' },
  { bg: '#2a2e28', edge: '#3a4038', ink: '#b8c4b2' },
  { bg: '#302b26', edge: '#423a32', ink: '#c4b8a2' },
  { bg: '#262b2d', edge: '#353d40', ink: '#aebfc4' },
  { bg: '#2d2830', edge: '#3e3646', ink: '#c0b2c4' },
  { bg: '#322e26', edge: '#454033', ink: '#ccc2a6' },
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
</script>

<style scoped>
/* 软木板 */
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
  animation: noteIn .7s ease-in-out both;
}
.note-wrap:hover { z-index: 10; }
@keyframes noteIn { from { opacity: 0; } to { opacity: 1; } }

/* 便签 */
.note {
  position: relative;
  width: 100%;
  border: 1px solid;
  text-decoration: none;
  transition: box-shadow .7s ease-in-out, filter .7s ease-in-out;
  padding: 20px 14px 14px;
  display: flex; flex-direction: column;
  overflow: visible;
}
.note:hover {
  box-shadow: 0 0 0 1px color-mix(in srgb, currentColor 30%, transparent);
  filter: brightness(.98);
}
.note:focus-visible { outline: 1px solid var(--accent); outline-offset: 3px; }

/* 图钉：墨点 */
.pin {
  position: absolute;
  top: -6px; left: 50%; transform: translateX(-50%);
  width: 10px; height: 10px;
  border-radius: 50%;
  background: var(--ink);
  opacity: .85;
  z-index: 2;
}

/* 胶带 */
.tape {
  position: absolute;
  width: 36px; height: 14px;
  background: rgba(255, 255, 255, .4);
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
  background: rgba(0, 0, 0, .04);
}
.note-imgs--1 .img-cell { aspect-ratio: 4 / 3; }
.img-cell img { width: 100%; height: 100%; object-fit: cover; display: block; }
/* 第 4 格叠加剩余张数 */
.img-more {
  position: absolute; inset: 0;
  display: flex; align-items: center; justify-content: center;
  background: rgba(0, 0, 0, .45);
  color: #fff; font-size: 15px; font-weight: 400;
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
  font-size: 11px;
  opacity: 0;
  transition: opacity .7s ease-in-out;
}
.note:hover .note-more,
.note:focus-visible .note-more { opacity: .9; }
/* 触屏无 hover，常显弱化版 */
@media (hover: none) {
  .note-more { opacity: .55; }
}

@media (prefers-reduced-motion: reduce) {
  .note-wrap { animation: none; }
  .note, .note-more { transition: none; }
}
</style>
