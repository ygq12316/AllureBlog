<template>
  <div class="page">
    <div class="header">
      <h2 class="title">随笔 · 说说</h2>
      <p class="sub">一切随心，随手记下</p>
    </div>
    <div v-if="!notes.length" class="empty"><n-empty description="还没有随笔" /></div>
    <div v-else class="board">
      <router-link v-for="(n, i) in pinned" :key="n.id" :to="'/notes/'+n.id" class="note"
        :style="noteStyle(n.id, i)">
        <div class="pin" />
        <div class="tape" :class="'tape-'+((i*7+3)%4)" />
        <div v-if="firstImg(n.images)" class="note-img"><img :src="firstImg(n.images)" loading="lazy" @error="e=>e.target.style.display='none'" /></div>
        <div class="note-body" v-html="n.html" />
        <time class="note-time">{{ rel(n.created_at) }}</time>
      </router-link>
    </div>
  </div>
</template>

<script setup>
import { ref, computed, onMounted } from 'vue'
import axios from 'axios'

const notes = ref([])
onMounted(async () => {
  try { const { data } = await axios.get('/api/notes?per_page=100'); notes.value = data.notes || [] } catch {}
})

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

function hash(s, i) {
  let r = i * 17
  for (let c of s) r = (r << 5) - r + c.charCodeAt(0) | 0
  return Math.abs(r)
}

function noteStyle(id, i) {
  const c = paperColors[i % paperColors.length]
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

function firstImg(s) { return s ? s.split(',')[0]?.trim() || '' : '' }

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

/* 便签 */
.note {
  position: relative;
  max-width: 280px; width: 100%;
  justify-self: center;
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
  z-index: 10;
}

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

.note-img {
  width: 100%;
  max-height: 45%;
  overflow: hidden;
  border-radius: 2px;
  margin-bottom: 8px;
  flex-shrink: 0;
}
.note-img img {
  width: 100%;
  display: block;
  object-fit: cover;
}
.note-body {
  font-size: 13px;
  line-height: 1.6;
  display: -webkit-box;
  -webkit-line-clamp: 6;
  -webkit-box-orient: vertical;
  overflow: hidden;
  margin-bottom: 8px;
  flex: 0 0 auto;
}
.note-time {
  font-size: 10px;
  opacity: .6;
  margin-top: auto;
  flex-shrink: 0;
  text-align: right;
}
</style>
