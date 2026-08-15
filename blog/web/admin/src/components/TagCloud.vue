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
let slugs = []
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

async function build() {
  if (!el.value || !props.tags.length) return
  if (!TagCloud) await loadScript()
  instance?.destroy?.()
  el.value.innerHTML = ''

  slugs = props.tags.map(t => t.slug)
  const texts = props.tags.map(t => t.name)

  instance = TagCloud(el.value, texts, {
    radius: 120,
    maxSpeed: 'fast',
    initSpeed: 'normal',
    direction: 135,
    keep: false,
  })

  instance.maxSpeed = 5

  setTimeout(() => {
    const items = el.value?.querySelectorAll?.('.tagcloud--item')
    items?.forEach((el, i) => {
      el.style.cursor = 'pointer'
      el.onclick = () => { if (slugs[i]) emit('select', slugs[i]) }
    })
  }, 500)
}

onMounted(() => nextTick(build))
watch(() => props.tags, () => nextTick(build), { deep: true })
onUnmounted(() => instance?.destroy?.())
</script>

<style>
.cloud .tagcloud--item {
  color: #5c4a3a !important;
  font-weight: 500;
  font-family: 'LXGW WenKai', serif !important;
  font-size: 14px !important;
  padding: 4px 10px;
  border-radius: 4px;
  transition: color .15s, background .15s;
}
.cloud .tagcloud--item:hover {
  color: #b8944c !important;
  font-weight: 700 !important;
  background: rgba(184,148,76,.08);
}
</style>

<style scoped>
.cloud-wrap { position: relative; width: 460px; height: 460px; margin: 0 auto; }
.glow {
  position: absolute; top: 50%; left: 50%; transform: translate(-50%, -50%);
  width: 380px; height: 380px; border-radius: 50%;
  background: radial-gradient(circle, rgba(184,148,76,.08), rgba(184,148,76,.02) 50%, transparent 70%);
  box-shadow: 0 0 80px rgba(184,148,76,.05);
  pointer-events: none;
}
.cloud { position: absolute; inset: 0; display: flex; align-items: center; justify-content: center; }
</style>
