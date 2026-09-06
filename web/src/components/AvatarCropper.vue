<template>
  <div class="panel text-center">
    <!-- 裁剪框：200px 圆形视口；touch-none + 指针捕获，拖动在触屏上不滚页 -->
    <div ref="frame"
      class="relative w-[200px] h-[200px] mx-auto mb-2 rounded-full overflow-hidden border border-accent cursor-move touch-none select-none"
      @pointerdown="onDown" @pointermove="onMove" @pointerup="onUp" @pointercancel="onUp"
      @wheel.prevent="onWheel">
      <img ref="imgEl" :src="src" :style="imgStyle"
        class="absolute top-1/2 left-1/2 max-w-none pointer-events-none will-change-transform"
        alt="裁剪预览" @load="onImgLoad" @error="onImgError" />
      <div class="absolute inset-0 rounded-full pointer-events-none" style="box-shadow: inset 0 0 0 999px rgba(0,0,0,.35)" />
    </div>
    <p class="text-[11px] text-ink3 m-0 mb-2">拖动调整位置 · 滚轮或滑杆缩放</p>

    <div class="flex items-center justify-center gap-2 mb-3">
      <span class="text-base text-ink3">−</span>
      <input type="range" min="1" max="3" step="0.05" :value="zoom"
        @input="setZoom(Number($event.target.value))" class="w-[120px] accent-[var(--accent)]" aria-label="缩放" />
      <span class="text-base text-ink3">+</span>
    </div>

    <div class="flex justify-center gap-3">
      <InkButton size="xs" @click="emit('cancel')">取消</InkButton>
      <InkButton size="xs" variant="primary" :loading="uploading" @click="apply">{{ uploading ? '上传中…' : '裁剪并上传' }}</InkButton>
    </div>
    <p v-if="error" class="text-xs text-cinnabar mt-2 mb-0">{{ error }}</p>
  </div>
</template>

<script setup>
import { ref, computed, watch, onMounted } from 'vue'
import InkButton from './ui/InkButton.vue'
import { uploadFile } from '../api/upload'

const props = defineProps({ src: { type: String, required: true } })
const emit = defineEmits(['done', 'cancel'])

const SIZE = 200
const imgEl = ref(null)
const frame = ref(null)
const natural = ref({ w: 0, h: 0 }) // 原图尺寸
const zoom = ref(1) // 相对 cover 基准的倍率
const off = ref({ x: 0, y: 0 }) // 相对居中的位移（已钳位，图恒覆盖视口）
const uploading = ref(false)
const error = ref('')
let img = null // Image 对象，导出时取像素
let drag = null // { px, py, ox, oy }

const baseScale = computed(() => (natural.value.w ? SIZE / Math.min(natural.value.w, natural.value.h) : 0))
const dispW = computed(() => natural.value.w * baseScale.value * zoom.value)
const dispH = computed(() => natural.value.h * baseScale.value * zoom.value)

// 图心 = 视口中心 + off；可移动范围 = (显示尺寸 - 视口) / 2
const maxOff = computed(() => ({
  x: Math.max(0, (dispW.value - SIZE) / 2),
  y: Math.max(0, (dispH.value - SIZE) / 2),
}))
const clampOff = v => ({
  x: Math.max(-maxOff.value.x, Math.min(maxOff.value.x, v.x)),
  y: Math.max(-maxOff.value.y, Math.min(maxOff.value.y, v.y)),
})

const imgStyle = computed(() => ({
  width: dispW.value + 'px',
  height: dispH.value + 'px',
  transform: `translate(calc(-50% + ${off.value.x}px), calc(-50% + ${off.value.y}px))`,
}))

function onImgLoad() {
  img = imgEl.value
  natural.value = { w: img.naturalWidth, h: img.naturalHeight }
  off.value = clampOff(off.value)
}
function onImgError() {
  error.value = '图片读取失败，请换一张试试'
}

// zoom 变化后位移可能越界，统一钳位
watch(zoom, () => { off.value = clampOff(off.value) })

function setZoom(v) { zoom.value = Math.max(1, Math.min(3, v)) }
function onWheel(e) { setZoom(zoom.value + (e.deltaY > 0 ? -0.1 : 0.1)) }

function onDown(e) {
  e.preventDefault()
  try { frame.value.setPointerCapture(e.pointerId) } catch {} // 拖出框外/快速甩动不断线；合成事件无活动指针时忽略
  drag = { px: e.clientX, py: e.clientY, ox: off.value.x, oy: off.value.y }
}
function onMove(e) {
  if (!drag) return
  off.value = clampOff({ x: drag.ox + (e.clientX - drag.px), y: drag.oy + (e.clientY - drag.py) })
}
function onUp() { drag = null }

// 导出与显示 1:1：把图按当前显示位置画进 200×200 画布，圆形路径裁切
// 图心 = (SIZE/2 + ox, SIZE/2 + oy) → drawImage 左上角 = 图心 - 显示尺寸/2
async function apply() {
  if (!img || !natural.value.w) { error.value = '图片未就绪'; return }
  uploading.value = true
  error.value = ''
  try {
    const cv = document.createElement('canvas')
    cv.width = SIZE; cv.height = SIZE
    const ctx = cv.getContext('2d')
    ctx.beginPath()
    ctx.arc(SIZE / 2, SIZE / 2, SIZE / 2, 0, Math.PI * 2)
    ctx.clip()
    ctx.drawImage(img, SIZE / 2 + off.value.x - dispW.value / 2, SIZE / 2 + off.value.y - dispH.value / 2, dispW.value, dispH.value)
    const blob = await new Promise(r => cv.toBlob(r, 'image/png'))
    if (!blob) { error.value = '裁剪失败，请重试'; return }
    const fd = new FormData()
    fd.append('file', blob, 'avatar.png')
    const { url } = await uploadFile(fd)
    emit('done', url)
  } catch (e) {
    error.value = e.response?.data?.error || '上传失败，请重试'
  } finally {
    uploading.value = false
  }
}

onMounted(() => { zoom.value = 1; off.value = { x: 0, y: 0 } })
</script>
