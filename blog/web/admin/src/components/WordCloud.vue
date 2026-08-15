<template>
  <div class="wc-stage">
    <canvas ref="canvas" class="wc-canvas"></canvas>
  </div>
</template>

<script setup>
import { computed, ref, onMounted, watch } from 'vue'

const props = defineProps({ tags: { type: Array, default: () => [] } })
const emit = defineEmits(['select'])

const canvas = ref(null)

const sorted = computed(() =>
  [...props.tags].sort((a, b) => (b.article_count || 0) - (a.article_count || 0))
)

const palette = [
  '#b8944c', '#c4a060', '#d4b468', '#a08040',
  '#8a7a68', '#9a8a6a', '#786858', '#6a5a48',
  '#5c4a3a', '#6a5a4a', '#7a6a50', '#4a3a2a',
  '#a88a5c', '#b0906c', '#c0a878',
]

let WordCloud = null

function loadLib() {
  return new Promise(resolve => {
    if (window.WordCloud) { WordCloud = window.WordCloud; resolve(); return }
    const s = document.createElement('script')
    s.src = 'https://cdn.jsdelivr.net/npm/wordcloud@1.2.2/src/wordcloud2.min.js'
    s.onload = () => { WordCloud = window.WordCloud; resolve() }
    document.head.appendChild(s)
  })
}

async function render() {
  if (!canvas.value || !sorted.value.length) return
  if (!WordCloud) await loadLib()

  const dpr = window.devicePixelRatio || 1
  const w = canvas.value.clientWidth
  const h = 520
  canvas.value.width = w * dpr
  canvas.value.height = h * dpr
  canvas.value.style.height = h + 'px'

  const list = sorted.value.map(t => [t.name, (t.article_count || 1) * 10])

  WordCloud(canvas.value, {
    list: list,
    gridSize: 35,
    weightFactor: 2.5,
    fontFamily: 'LXGW WenKai, serif',
    color: () => palette[Math.floor(Math.random() * palette.length)],
    rotateRatio: 0,
    backgroundColor: 'transparent',
    minSize: 8,
    shape: 'circle',
    ellipticity: 0.8,
    clearCanvas: true,
    click: (item) => {
      emit('select', item[0])
    },
  })
}

onMounted(render)
watch(() => props.tags, render, { deep: true })
</script>

<style scoped>
.wc-section { text-align: center; }
.wc-title { font-size: 12px; color: var(--muted); font-weight: 400; margin: 0 0 8px; }
.wc-stage { max-width: 100%; margin: 0 auto; }
.wc-canvas { width: 100%; max-width: 100%; height: 520px; display: block; margin: 0 auto; }
</style>
