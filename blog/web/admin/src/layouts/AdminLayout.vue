<template>
  <div style="min-height:100vh;background:var(--bg)">
    <!-- 粒子背景 -->
    <canvas ref="canvas" class="particle-canvas" />
    <div ref="glow" class="mouse-glow" />

    <n-config-provider :theme-overrides="themeOverrides">
      <n-layout style="min-height:100vh;background:transparent" :content-style="{background:'transparent'}">
        <n-layout-header bordered style="background:var(--bg)">
          <div class="nav-bar">
            <span class="nav-title">✽ 笔墨后台</span>
            <n-menu v-model:value="activeMenu" mode="horizontal" :options="menuOptions" class="nav-menu" @update:value="onMenuClick" />
            <div class="nav-right"><ThemeToggle /><n-button text size="small" tag="a" href="/" style="color:var(--muted)">← 博客</n-button></div>
          </div>
        </n-layout-header>
        <n-layout-content class="admin-content"><router-view /></n-layout-content>
      </n-layout>
    </n-config-provider>
  </div>
</template>

<script setup>
import { ref, computed, onMounted, onUnmounted } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import ThemeToggle from '../components/ThemeToggle.vue'

const route=useRoute(), router=useRouter()
const activeMenu=computed(()=>{const p=route.path;if(p.startsWith('/admin/articles'))return'articles';if(p.startsWith('/admin/notes'))return'notes';if(p.startsWith('/admin/categories'))return'categories';if(p.startsWith('/admin/tags'))return'tags';return'dashboard'})
function onMenuClick(key){const m={dashboard:'/admin',articles:'/admin/articles',notes:'/admin/notes',comments:'/admin/comments',danmakus:'/admin/danmakus',categories:'/admin/categories',tags:'/admin/tags',visitors:'/admin/visitors',settings:'/admin/settings'};if(m[key])router.push(m[key])}
const menuOptions=[{label:'仪表盘',key:'dashboard'},{label:'文章',key:'articles'},{label:'随笔',key:'notes'},{label:'评论',key:'comments'},{label:'弹幕',key:'danmakus'},{label:'分类',key:'categories'},{label:'标签',key:'tags'},{label:'访客',key:'visitors'},{label:'设置',key:'settings'}]
const themeOverrides={
  common:{primaryColor:'#b8944c',primaryColorHover:'#d4b060',borderRadius:'4px'},
}

// 粒子系统（复制自 PublicLayout）
const canvas=ref(null),glow=ref(null)
let ctx,w,h,particles=[],mouse={x:-999,y:-999},animId
class Particle{constructor(){this.reset();this.y=Math.random()*h}reset(){this.x=Math.random()*w;this.y=h+10;this.size=Math.random()*2.5+1;this.speedY=-(Math.random()*.4+.15);this.speedX=(Math.random()-.5)*.3;this.opacity=Math.random()*.4+.15;this.hue=Math.random()>.5?'184,148,76':'120,105,81'}update(){this.x+=this.speedX;this.y+=this.speedY;const dx=this.x-mouse.x,dy=this.y-mouse.y,dist=Math.sqrt(dx*dx+dy*dy);if(dist<120){const f=(120-dist)/120*1.5;this.x+=(dx/dist)*f;this.y+=(dy/dist)*f};if(this.y<-10||this.x<-10||this.x>w+10)this.reset()}draw(){ctx.beginPath();ctx.arc(this.x,this.y,this.size,0,Math.PI*2);ctx.fillStyle=`rgba(${this.hue},${this.opacity})`;ctx.fill()}}
function initC(){if(!canvas.value)return;ctx=canvas.value.getContext('2d');resize();particles=Array.from({length:55},()=>new Particle());animate()}
function resize(){w=window.innerWidth;h=window.innerHeight;canvas.value.width=w;canvas.value.height=h}
function animate(){ctx.clearRect(0,0,w,h);particles.forEach(p=>{p.update();p.draw()});animId=requestAnimationFrame(animate)}
function onMM(e){mouse.x=e.clientX;mouse.y=e.clientY;if(glow.value){glow.value.style.opacity='1';glow.value.style.transform=`translate3d(${e.clientX-200}px,${e.clientY-200}px,0)`}}
function onML(){mouse.x=-999;mouse.y=-999;if(glow.value)glow.value.style.opacity='0'}
onMounted(()=>{initC();window.addEventListener('resize',resize);window.addEventListener('mousemove',onMM);document.addEventListener('mouseleave',onML)})
onUnmounted(()=>{cancelAnimationFrame(animId);window.removeEventListener('resize',resize);window.removeEventListener('mousemove',onMM);document.removeEventListener('mouseleave',onML)})
</script>

<style scoped>
.particle-canvas{position:fixed;top:0;left:0;width:100vw;height:100vh;pointer-events:none;z-index:0}
.mouse-glow{position:fixed;top:0;left:0;width:400px;height:400px;border-radius:50%;background:radial-gradient(circle,rgba(184,148,76,.06) 0%,transparent 70%);pointer-events:none;z-index:0;opacity:0;transition:opacity .6s}
.nav-bar{display:flex;align-items:center;padding:0 clamp(16px,5vw,32px);height:52px;position:relative;z-index:10}
.nav-title{font-weight:700;font-size:15px;color:var(--text);flex-shrink:0;margin-right:auto}
.nav-menu{flex:1;display:flex;justify-content:center;--n-item-text-color:var(--text2);--n-item-text-color-hover:var(--gold)}
.nav-right{display:flex;align-items:center;gap:12px;flex-shrink:0;margin-left:auto}
.admin-content{padding:clamp(20px,3vh,32px) clamp(16px,5vw,48px);background:transparent;position:relative;z-index:1;font-size:15px}

/* 深色输入框：覆盖 Naive UI 默认白色背景 */
:deep(.n-input) { --n-color: transparent !important; --n-color-focus: transparent !important; --n-text-color: var(--text) !important; --n-placeholder-color: var(--muted) !important; --n-border: 1px solid var(--card-border) !important; --n-border-focus: 1px solid var(--gold) !important; --n-border-hover: 1px solid var(--gold) !important; --n-box-shadow-focus: 0 0 0 2px rgba(184,148,76,0.15) !important; }
:deep(.n-base-selection) { --n-color: transparent !important; --n-color-active: transparent !important; --n-text-color: var(--text) !important; --n-border: 1px solid var(--card-border) !important; --n-border-focus: 1px solid var(--gold) !important; --n-border-hover: 1px solid var(--gold) !important; }
:deep(.n-base-select-menu) { --n-color: var(--card) !important; }
:deep(.n-base-selection-label) { --n-text-color: var(--text) !important; background: transparent !important; }
:deep(.n-base-selection-tag) { --n-color: var(--tag-bg) !important; }
:deep(.n-dynamic-tags .n-input) { --n-color: transparent !important; --n-color-focus: transparent !important; }
</style>
