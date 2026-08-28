<template>
  <div class="cloud-wrap">
    <div class="glow" />
    <div ref="el" class="cloud" />
  </div>
</template>

<script setup>
import { ref, onMounted, onUnmounted, watch, nextTick } from 'vue'

const props = defineProps({
  tags: { type: Array, default: () => [] },
  active: { type: String, default: '' },
})
const emit = defineEmits(['select'])

const el = ref(null)
let instance = null
let TagCloud = null

function loadScript() {
  return new Promise(resolve => {
    if (window.TagCloud) { TagCloud = window.TagCloud; resolve(); return }
    const s = document.createElement('script')
    s.src = '/TagCloud.min.js'
    s.onload = () => { TagCloud = window.TagCloud; resolve() }
    document.head.appendChild(s)
  })
}

// 用户偏好减弱动效时不旋转，仅静态呈现
const reducedMotion = window.matchMedia('(prefers-reduced-motion: reduce)').matches

async function build() {
  if (!el.value || !props.tags.length) return
  if (!TagCloud) await loadScript()
  instance?.destroy?.()
  el.value.innerHTML = ''

  const texts = props.tags.map(t => t.name)
  // 球体半径随容器缩放，大屏更饱满，移动端不溢出
  const radius = Math.round(el.value.clientWidth * 0.36)

  instance = TagCloud(el.value, texts, {
    radius,
    maxSpeed: reducedMotion ? 0 : 3.5,
    initSpeed: reducedMotion ? 0 : 'normal',
    direction: 135,
    keep: false,
  })

  bindItems()
}

// 库同步渲染 item，直接绑定点击与键盘事件
function bindItems() {
  const items = el.value?.querySelectorAll?.('.tagcloud--item')
  if (!items?.length) return
  const counts = props.tags.map(t => Number(t.count) || 0)
  const min = Math.min(...counts), max = Math.max(...counts)
  const span = max - min || 1
  items.forEach((item, i) => {
    const tag = props.tags[i]
    // 按文章数分级字号，热门分类更醒目
    const w = (counts[i] - min) / span
    item.style.fontSize = (14 + Math.round(w * 7)) + 'px'
    item.style.cursor = 'pointer'
    item.tabIndex = 0
    item.setAttribute('role', 'link')
    const pick = () => { if (tag?.slug) emit('select', tag.slug) }
    item.onclick = pick
    item.onkeydown = e => {
      if (e.key === 'Enter' || e.key === ' ') { e.preventDefault(); pick() }
    }
  })
}

// 视口变化时防抖重建，保持球体铺满容器
let resizeTimer = null
function onResize() {
  clearTimeout(resizeTimer)
  resizeTimer = setTimeout(build, 200)
}

onMounted(() => {
  nextTick(build)
  window.addEventListener('resize', onResize)
})
watch(() => props.tags, () => nextTick(build), { deep: true })
onUnmounted(() => {
  window.removeEventListener('resize', onResize)
  clearTimeout(resizeTimer)
  instance?.destroy?.()
})
</script>

<style>
.cloud .tagcloud--item {
  color: var(--text) !important;
  font-weight: 500;
  font-family: 'LXGW WenKai', serif !important;
  padding: 5px 12px;
  border-radius: 999px;
  background: color-mix(in srgb, var(--gold) 6%, var(--bg));
  border: 1px solid color-mix(in srgb, var(--border) 60%, transparent);
  transition: color .15s, background-color .15s, border-color .15s;
}
.cloud .tagcloud--item:hover,
.cloud .tagcloud--item:focus-visible {
  color: var(--gold) !important;
  background: color-mix(in srgb, var(--gold) 12%, var(--bg));
  border-color: color-mix(in srgb, var(--gold) 50%, transparent);
  text-shadow: 0 0 14px color-mix(in srgb, var(--gold) 40%, transparent);
}
</style>

<style scoped>
.cloud-wrap { position: relative; width: min(560px, 92vw); aspect-ratio: 1; margin: 0 auto; }
.glow {
  position: absolute; inset: 4%; border-radius: 50%;
  /* 墨晕圈：中心微染，球缘一圈晕开的金环，勾出球体轮廓 */
  background: radial-gradient(circle,
    color-mix(in srgb, var(--gold) 5%, transparent) 0%,
    color-mix(in srgb, var(--gold) 3%, transparent) 50%,
    color-mix(in srgb, var(--gold) 18%, transparent) 64%,
    color-mix(in srgb, var(--gold) 6%, transparent) 74%,
    transparent 80%);
  border: 1px solid color-mix(in srgb, var(--gold) 22%, transparent);
  box-shadow:
    0 0 90px color-mix(in srgb, var(--gold) 7%, transparent),
    inset 0 0 70px color-mix(in srgb, var(--gold) 5%, transparent);
  pointer-events: none;
}
.cloud {
  position: absolute; inset: 0;
  display: flex; align-items: center; justify-content: center;
  touch-action: manipulation;
}
</style>
