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

// 墨分五色：热度越高墨色越淡雅（浓墨 → 淡彩）
const palette = [
  '#2c2c2c', '#454540', '#55524b', '#5c6158',
  '#6b7b6e', '#7a8578', '#8a8072', '#96897a',
  '#a89279', '#b0a08a', '#bcae97', '#c4b9a8',
]
// 暗色模式换浅墨系，避免深墨字沉进暗背景
const darkPalette = [
  '#8a8072', '#96907e', '#a89e88', '#b5ab99',
  '#bcae97', '#c4b9a8', '#cec2ad', '#d8d2c4',
  '#ddd6c6', '#e2dccd', '#e8e2d6', '#ede8dc',
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
  const h = 480
  canvas.value.width = w * dpr
  canvas.value.height = h * dpr
  canvas.value.style.height = h + 'px'

  // 字号压缩：平方根 + 封顶，热门标签醒目但不出格（约 16–48px）
  const weightOf = c => Math.min(Math.sqrt(c || 1), 3.1)
  const counts = sorted.value.map(t => t.article_count || 1)
  const lo = weightOf(Math.min(...counts))
  const hi = weightOf(Math.max(...counts))
  const span = hi - lo
  const list = sorted.value.map(t => [t.name, weightOf(t.article_count)])
  const colors = document.documentElement.classList.contains('dark') ? darkPalette : palette

  WordCloud(canvas.value, {
    list,
    gridSize: 30,
    weightFactor: 30,
    fontFamily: 'LXGW WenKai, serif',
    // 热度映射取色：冷门偏浓墨、热门偏淡彩，刷新不跳色
    color: (word, weight) => {
      const norm = span ? (Math.min(weight, hi) - lo) / span : 0.6
      return colors[Math.round(norm * (colors.length - 1))]
    },
    rotateRatio: 0,
    backgroundColor: 'transparent',
    minSize: 9,
    shape: 'circle',
    ellipticity: 0.8,
    shrinkToFit: true,
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
.wc-stage { max-width: 100%; margin: 0 auto; }
.wc-canvas { width: 100%; max-width: 100%; height: 480px; display: block; margin: 0 auto; cursor: pointer; }
</style>
