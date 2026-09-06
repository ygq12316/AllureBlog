// 水墨粒子背景:两个 Layout 共用(此前整段复制,且后台版本丢了无障碍处理)。
// prefers-reduced-motion:只画一帧静景,不起动画循环。
import { ref, onMounted, onUnmounted } from 'vue'

export function useParticles() {
  const canvas = ref(null)
  const glow = ref(null)
  let ctx, w, h, particles = [], mouse = { x: -999, y: -999 }, animId
  const reduceMotion = matchMedia('(prefers-reduced-motion: reduce)').matches

  class Particle {
    constructor() { this.reset(); this.y = Math.random() * h }
    reset() {
      this.x = Math.random() * w; this.y = h + 10
      this.size = Math.random() * 2.5 + 1
      this.speedY = -(Math.random() * 0.4 + 0.15)
      this.speedX = (Math.random() - 0.5) * 0.3
      this.opacity = Math.random() * 0.4 + 0.15
      // 淡墨双色：苔绿与茶棕，明暗主题下皆可读
      this.hue = Math.random() > 0.5 ? '107,123,110' : '168,146,121'
    }
    update() {
      this.x += this.speedX; this.y += this.speedY
      const dx = this.x - mouse.x, dy = this.y - mouse.y, dist = Math.sqrt(dx * dx + dy * dy)
      if (dist < 120) { const f = (120 - dist) / 120 * 1.5; this.x += (dx / dist) * f; this.y += (dy / dist) * f }
      if (this.y < -10 || this.x < -10 || this.x > w + 10) this.reset()
    }
    draw() {
      ctx.beginPath(); ctx.arc(this.x, this.y, this.size, 0, Math.PI * 2)
      ctx.fillStyle = `rgba(${this.hue},${this.opacity})`; ctx.fill()
    }
  }

  function drawFrame() { ctx.clearRect(0, 0, w, h); particles.forEach(p => p.draw()) }
  function resize() {
    w = window.innerWidth; h = window.innerHeight
    canvas.value.width = w; canvas.value.height = h
    if (reduceMotion && particles.length) drawFrame()
  }
  function animate() { drawFrame(); particles.forEach(p => p.update()); animId = requestAnimationFrame(animate) }
  function initC() {
    if (!canvas.value) return
    ctx = canvas.value.getContext('2d')
    resize()
    particles = Array.from({ length: 55 }, () => new Particle())
    if (reduceMotion) { drawFrame(); return }
    animate()
  }
  function onMM(e) {
    mouse.x = e.clientX; mouse.y = e.clientY
    if (glow.value) { glow.value.style.opacity = '1'; glow.value.style.transform = `translate3d(${e.clientX - 200}px,${e.clientY - 200}px,0)` }
  }
  function onML() {
    mouse.x = -999; mouse.y = -999
    if (glow.value) glow.value.style.opacity = '0'
  }

  onMounted(() => {
    initC()
    window.addEventListener('resize', resize)
    window.addEventListener('mousemove', onMM)
    document.addEventListener('mouseleave', onML)
  })
  onUnmounted(() => {
    if (animId) cancelAnimationFrame(animId)
    window.removeEventListener('resize', resize)
    window.removeEventListener('mousemove', onMM)
    document.removeEventListener('mouseleave', onML)
  })

  return { canvas, glow }
}
