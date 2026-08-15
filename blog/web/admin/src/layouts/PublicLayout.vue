<template>
  <div class="blog-root" style="display:flex;flex-direction:column;min-height:100vh;background:var(--bg);color:var(--text);font-family:'LXGW WenKai',serif;line-height:1.8">
    <canvas ref="canvas" class="particle-canvas" />
    <div ref="glow" class="mouse-glow" />

    <nav class="blog-nav">
      <router-link to="/" class="nav-brand">
        <n-icon size="18" color="var(--gold)" :component="EditIcon" /> 笔墨
      </router-link>
      <div class="nav-links">
        <router-link v-for="l in navs" :key="l.to" :to="l.to"
          class="nav-link" active-class="nav-link--active" :exact="l.exact">{{ l.label }}</router-link>
      </div>
      <div class="nav-right"><ThemeToggle /><UserAvatar /></div>
    </nav>

    <main class="blog-main"><router-view /></main>

    <footer class="blog-footer">&copy; 2026 笔墨 &middot; 记录思考，分享生活</footer>

    <VisitorSetup v-if="setupVisible" @close="closeSetup" />
    <FairyChat />
  </div>
</template>

<script setup>
import { ref, onMounted, onUnmounted } from 'vue'
import { CreateOutline } from '@vicons/ionicons5'
import ThemeToggle from '../components/ThemeToggle.vue'
import UserAvatar from '../components/UserAvatar.vue'
import VisitorSetup from '../components/VisitorSetup.vue'
import FairyChat from '../components/FairyChat.vue'
import { useVisitor } from '../composables/useVisitor'

const { setupVisible, isSetUp, closeSetup, init } = useVisitor()

onMounted(async () => {
  await init()
  // 如果未设定身份，弹出设置
  if (!isSetUp.value) {
    setupVisible.value = true
  }
})

const EditIcon = CreateOutline
const navs = [
  { to: '/search', label: '搜索' }, { to: '/', label: '主页', exact: true },
  { to: '/category', label: '分类' }, { to: '/notes', label: '随笔' }, { to: '/articles', label: '文章' },
]

// Particle system (same as before)
const canvas = ref(null), glow = ref(null)
let ctx, w, h, particles = [], mouse = { x: -999, y: -999 }, animId
class Particle {
  constructor() { this.reset(); this.y = Math.random() * h }
  reset() { this.x = Math.random() * w; this.y = h + 10; this.size = Math.random() * 2.5 + 1; this.speedY = -(Math.random() * 0.4 + 0.15); this.speedX = (Math.random() - 0.5) * 0.3; this.opacity = Math.random() * 0.4 + 0.15; this.hue = Math.random() > 0.5 ? '184,148,76' : '120,105,81' }
  update() { this.x += this.speedX; this.y += this.speedY; const dx = this.x - mouse.x, dy = this.y - mouse.y, dist = Math.sqrt(dx*dx+dy*dy); if (dist < 120) { const f = (120-dist)/120*1.5; this.x += (dx/dist)*f; this.y += (dy/dist)*f }; if (this.y < -10 || this.x < -10 || this.x > w+10) this.reset() }
  draw() { ctx.beginPath(); ctx.arc(this.x, this.y, this.size, 0, Math.PI*2); ctx.fillStyle = `rgba(${this.hue},${this.opacity})`; ctx.fill() }
}
function initCanvas() { if (!canvas.value) return; ctx = canvas.value.getContext('2d'); resize(); particles = Array.from({length:55},()=>new Particle()); animate() }
function resize() { w = window.innerWidth; h = window.innerHeight; canvas.value.width = w; canvas.value.height = h }
function animate() { ctx.clearRect(0,0,w,h); particles.forEach(p=>{p.update();p.draw()}); animId = requestAnimationFrame(animate) }
function onMM(e) { mouse.x = e.clientX; mouse.y = e.clientY; if (glow.value) { glow.value.style.opacity='1'; glow.value.style.transform=`translate3d(${e.clientX-200}px,${e.clientY-200}px,0)` } }
function onML() { mouse.x=-999; mouse.y=-999; if(glow.value) glow.value.style.opacity='0' }
onMounted(() => { initCanvas(); window.addEventListener('resize',resize); window.addEventListener('mousemove',onMM); document.addEventListener('mouseleave',onML) })
onUnmounted(() => { cancelAnimationFrame(animId); window.removeEventListener('resize',resize); window.removeEventListener('mousemove',onMM); document.removeEventListener('mouseleave',onML) })
</script>

<style>
.particle-canvas { position: fixed; top: 0; left: 0; width: 100vw; height: 100vh; pointer-events: none; z-index: 0; }
.mouse-glow { position: fixed; top: 0; left: 0; width: 400px; height: 400px; border-radius: 50%; background: radial-gradient(circle, rgba(184,148,76,0.06) 0%, transparent 70%); pointer-events: none; z-index: 0; opacity: 0; transition: opacity 0.6s; }
.blog-nav { display: flex; align-items: center; justify-content: space-between; height: 52px; padding: 0 clamp(16px, 5vw, 64px); border-bottom: 1px solid var(--card-border); position: sticky; top: 0; z-index: 50; background: var(--bg); backface-visibility: hidden; -webkit-backface-visibility: hidden; }
.nav-brand { font-size: clamp(14px, 1.5vw, 17px); font-weight: 700; color: var(--gold); text-decoration: none; display: flex; align-items: center; gap: 4px; }
.nav-links { display: flex; gap: clamp(20px, 4vw, 40px); }
.nav-link { font-size: clamp(12px, 1.1vw, 14px); color: var(--text); text-decoration: none; transition: color .2s; }
.nav-link:hover { color: var(--gold); }
.nav-link--active { color: var(--gold); font-weight: 600; }
.nav-right { display: flex; align-items: center; gap: 12px; flex-shrink: 0; }
.blog-main { flex: 1; position: relative; z-index: 1; max-width: 1100px; width: 100%; margin: 0 auto; padding: clamp(24px, 4vh, 48px) clamp(16px, 5vw, 64px); }
.blog-footer { text-align: center; padding: 16px clamp(16px, 5vw, 64px); border-top: 1px solid var(--card-border); font-size: clamp(10px, 0.8vw, 12px); color: var(--muted); position: relative; z-index: 1; }
</style>
