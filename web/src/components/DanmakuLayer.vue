<template>
  <div class="danmaku-wrap" ref="wrap">
    <!-- 弹幕轨道 -->
    <div class="dm-track" v-for="track in tracks" :key="track">
      <span v-for="d in trackDanmaku[track]" :key="d.id" class="dm-item"
        :style="{ color: d.color, animationDuration: animDuration(d) + 's', animationDelay: animDelay(d, track) + 's' }">
        <span class="dm-dot" :style="{ background: d.color }" />
        {{ d.content }}
      </span>
    </div>

    <!-- 右侧入口 -->
    <div class="dm-sidebar">
      <div v-if="!isLoggedIn" class="dm-side-pill" @click="openLogin">
        发弹幕
      </div>
      <div v-else class="dm-side-pill dm-pill-input">
        <input v-model="dmText" class="dm-pill-text" placeholder="弹幕..." maxlength="50"
          @keydown.enter="sendDm" />
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, reactive, computed, onMounted, onUnmounted } from 'vue'
import { useVisitor } from '../composables/useVisitor'
import { useWS } from '../composables/useWS'
import { listDanmaku, createDanmaku } from '../api/danmaku'

defineEmits(['showLogin'])
const { account, openLogin } = useVisitor()
const isLoggedIn = computed(() => !!account.value)

const tracks = [0, 1, 2, 3, 4]
const danmaku = ref([])
const dmText = ref('')
const wrap = ref(null)
const trackDanmaku = reactive({})
tracks.forEach(t => trackDanmaku[t] = [])

let counter = 0, timer = null
const shownIds = new Set() // 本地自插 / WS 回环 / 轮询补漏三路共用的去重

function animDuration() { return 6 + Math.random() * 10 }
function animDelay() { return (counter++) * 0.3 }

async function sendDm() {
  const text = dmText.value.trim()
  if (!text || !account.value) return
  try {
    const data = await createDanmaku({
      visitor_uuid: account.value.uuid,
      content: text,
      color: '#a89279',
    })
    addToTrack(data.danmaku)
  } catch {
    openLogin() // 登录态失效：拉起登录
  }
  dmText.value = ''
}

function addToTrack(d) {
  if (!d || shownIds.has(d.id)) return
  shownIds.add(d.id)
  const t = Math.floor(Math.random() * tracks.length)
  trackDanmaku[t].push(d)
}

async function loadDanmaku() {
  try {
    danmaku.value = await listDanmaku()
  } catch {}
}

// 首屏铺底：均分五轨；之后的增量（WS / 轮询补漏）走 addToTrack
function seedTracks() {
  tracks.forEach(t => trackDanmaku[t] = [])
  danmaku.value.forEach((d, i) => {
    shownIds.add(d.id)
    trackDanmaku[i % tracks.length].push(d)
  })
}

// 轮询仅作对账兜底：WS 掉线期间的漏网弹幕在此补齐
function reconcile() {
  loadDanmaku().then(() => danmaku.value.forEach(addToTrack))
}

onMounted(() => {
  loadDanmaku().then(seedTracks)
  // 全局房间：实时接收所有人的新弹幕
  const { connect } = useWS('/api/ws', {
    onMessage: msg => { if (msg.type === 'danmaku' && msg.danmaku) addToTrack(msg.danmaku) },
  })
  connect()
  timer = setInterval(reconcile, 60000)
})
onUnmounted(() => clearInterval(timer))
</script>

<style scoped>
.danmaku-wrap {
  position: absolute; top: 25vh; left: 0;
  width: 100%; height: 130px;
  overflow: hidden; pointer-events: none;
  background: transparent; z-index: 3;
}

.dm-track {
  position: absolute; left: 0; right: 60px;
  height: 26px; overflow: hidden; white-space: nowrap;
}
.dm-track:nth-child(1) { top: 0; }
.dm-track:nth-child(2) { top: 24px; }
.dm-track:nth-child(3) { top: 48px; }
.dm-track:nth-child(4) { top: 72px; }
.dm-track:nth-child(5) { top: 96px; }

.dm-item {
  position: absolute; right: -100%; white-space: nowrap;
  font-size: 13px; font-family: 'LXGW WenKai', serif;
  animation: dm-scroll linear infinite; opacity: .7;
}
@keyframes dm-scroll {
  from { right: -100%; }
  to { right: 110%; }
}
.dm-dot {
  display: inline-block; width: 4px; height: 4px;
  border-radius: 50%; margin-right: 3px; vertical-align: middle;
}

/* 右侧入口 */
.dm-sidebar {
  position: fixed; right: 16px; top: 60px;
  pointer-events: all; z-index: 60;
}
.dm-side-pill {
  background: var(--paper);
  border: 1px solid var(--line);
  padding: 8px 16px;
  font-size: 12px; color: var(--ink3);
  cursor: pointer;
  transition: border-color .7s ease-in-out, color .7s ease-in-out;
  font-family: 'LXGW WenKai', serif;
}
.dm-side-pill:hover {
  border-color: var(--accent); color: var(--accent-strong);
}
.dm-pill-input {
  padding: 4px 6px 4px 14px;
  cursor: default;
}
.dm-pill-text {
  border: none; background: transparent;
  color: var(--ink); font-size: 12px;
  font-family: 'LXGW WenKai', serif;
  outline: none;
  caret-color: var(--accent);
  width: 80px;
}
.dm-pill-text::placeholder { color: var(--ink3); }
@media (prefers-reduced-motion: reduce) {
  .dm-side-pill { transition: none; }
}
</style>
