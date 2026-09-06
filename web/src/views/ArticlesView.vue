<template>
  <div class="w-full"><h2 class="text-xl md:text-2xl font-light tracking-[0.2em] text-ink m-0 mb-8 text-center">文章 · 时间轴</h2>
    <div v-if="!articles.length" class="text-center py-20 text-ink3 text-sm">还没有文章</div>
    <div v-else class="timeline">
      <template v-for="(a,i) in articles" :key="a.id">
        <div v-if="yearBreak(a)" class="year-marker"><span class="year-dot" /><span class="year-text">{{ year(a) }}</span><span class="year-dot" /></div>
        <div class="tl-row" :class="i%2===0?'tl-left':'tl-right'">
          <div class="tl-card" :style="{animationDelay:(i%10)*.08+'s'}">
            <router-link :to="'/posts/'+a.slug" class="group block no-underline p-4 md:p-5 bg-paper2 border border-line transition-colors duration-700 hover:border-accent">
              <div class="flex items-center gap-2 mb-1.5"><time class="text-[11px] text-accent-strong tracking-widest">{{ fmt(a.created_at) }}</time><span class="w-1.5 h-1.5 rounded-full" :class="'dot-'+catColor(a.category)" /></div>
              <h3 class="text-[15px] font-light tracking-wide text-ink m-0 mb-1.5 leading-snug transition-colors duration-700 group-hover:text-accent-strong">{{ a.title }}</h3>
              <p class="text-xs text-ink3 leading-relaxed m-0 mb-2">{{ a.excerpt }}</p>
              <div class="flex flex-wrap gap-1">
                <span v-if="a.category" class="text-[9px] px-1.5 py-px border border-accent/40 text-accent-strong">{{ a.category }}</span>
                <span v-for="t in tags(a.tags)" :key="t" class="text-[9px] px-1.5 py-px border border-line2 text-ink3">{{ t }}</span>
              </div>
            </router-link>
          </div>
        </div>
      </template>
    </div>
  </div>
</template>
<script setup>
import { ref, onMounted } from 'vue'
import { listArticles } from '../api/articles'
const articles=ref([]);let lastY=''
onMounted(async()=>{try{const data=await listArticles({per_page:100});articles.value=data.articles||[]}catch(e){}})
function year(a){return new Date(a.created_at).getFullYear()}
function yearBreak(a){const y=year(a).toString();if(y!==lastY){lastY=y;return true};return false}
function fmt(d){return d?new Date(d).toLocaleDateString('zh-CN',{month:'short',day:'numeric'}):''}
function tags(t){return t?t.split(',').map(x=>x.trim()).filter(Boolean).slice(0,3):[]}
function catColor(c){const m={'技术':'tech','生活':'life','读书':'read'};return m[c]||'default'}
</script>
<style scoped>
.timeline{position:relative;max-width:800px;margin:0 auto}
.timeline::before{content:'';position:absolute;left:50%;top:0;bottom:0;width:1px;background:var(--line);transform:translateX(-50%)}
.year-marker{display:flex;align-items:center;justify-content:center;gap:14px;margin:32px auto 20px;position:relative;z-index:2;background:var(--paper);width:fit-content;padding:0 12px}
.year-dot{width:6px;height:6px;border-radius:50%;background:var(--accent)}
.year-text{font-size:clamp(18px,2vw,24px);font-weight:400;color:var(--ink);letter-spacing:.15em}
.tl-row{position:relative;padding:0 0 24px;animation:fadeUp .7s ease-in-out both}
@keyframes fadeUp{from{opacity:0}to{opacity:1}}
.tl-left{margin-left:0;width:calc(50% - 28px)}.tl-right{margin-left:calc(50% + 28px);width:calc(50% - 28px)}
.tl-row::before{content:'';position:absolute;top:16px;width:8px;height:8px;border-radius:50%;background:var(--ink);border:2px solid var(--paper);z-index:1}
.tl-left::before{left:calc(100% + 28px);transform:translateX(-50%)}.tl-right::before{left:calc(0px - 28px);transform:translateX(-50%)}
.dot-tech{background:var(--accent)}.dot-life{background:#6b7b6e}.dot-read{background:#c4b9a8}.dot-default{background:var(--line)}
/* 移动端：单列，竖线居左 */
@media (max-width: 640px){
  .timeline::before{left:8px;transform:none}
  .tl-left,.tl-right{width:auto;margin-left:28px}
  .tl-left::before,.tl-right::before{left:-24px}
  .year-marker{margin-left:28px;margin-right:0;justify-content:flex-start}
}
@media (prefers-reduced-motion: reduce){.tl-row{animation:none}}
</style>
