<template>
  <div class="page"><h2 class="page-title">文章 · 时间轴</h2>
    <div v-if="!articles.length" class="empty">还没有文章</div>
    <div v-else class="timeline">
      <template v-for="(a,i) in articles" :key="a.id">
        <div v-if="yearBreak(a)" class="year-marker"><span class="year-dot" /><span class="year-text">{{ year(a) }}</span><span class="year-dot" /></div>
        <div class="tl-row" :class="i%2===0?'tl-left':'tl-right'">
          <div class="tl-card" :style="{animationDelay:(i%10)*.06+'s'}">
            <router-link :to="'/posts/'+a.slug" class="card-inner">
              <div class="card-head"><time class="card-date">{{ fmt(a.created_at) }}</time><span class="card-dot" :class="'dot-'+catColor(a.category)" /></div>
              <h3 class="card-title">{{ a.title }}</h3><p class="card-excerpt">{{ a.excerpt }}</p>
              <div class="card-foot"><span v-if="a.category" class="card-cat">{{ a.category }}</span><span v-for="t in tags(a.tags)" :key="t" class="card-tag">{{ t }}</span></div>
            </router-link>
          </div>
        </div>
      </template>
    </div>
  </div>
</template>
<script setup>
import { ref, onMounted } from 'vue'
import axios from 'axios'
const articles=ref([]);let lastY=''
onMounted(async()=>{try{const{data}=await axios.get('/api/articles?per_page=100');articles.value=data.articles||[]}catch(e){}})
function year(a){return new Date(a.created_at).getFullYear()}
function yearBreak(a){const y=year(a).toString();if(y!==lastY){lastY=y;return true};return false}
function fmt(d){return d?new Date(d).toLocaleDateString('zh-CN',{month:'short',day:'numeric'}):''}
function tags(t){return t?t.split(',').map(x=>x.trim()).filter(Boolean).slice(0,3):[]}
function catColor(c){const m={'技术':'tech','生活':'life','读书':'read'};return m[c]||'default'}
</script>
<style scoped>
.page{padding:0;width:100%}.page-title{font-size:clamp(20px,2vw,26px);font-weight:700;color:var(--text);margin:0 0 32px;text-align:center}.empty{text-align:center;padding:80px 0;color:var(--muted);font-size:14px}
.timeline{position:relative;max-width:800px;margin:0 auto}.timeline::before{content:'';position:absolute;left:50%;top:0;bottom:0;width:1px;background:var(--border);transform:translateX(-50%)}
.year-marker{display:flex;align-items:center;justify-content:center;gap:14px;margin:32px 0 20px;position:relative;z-index:2;background:var(--bg);width:fit-content;margin-left:auto;margin-right:auto;padding:0 12px}
.year-dot{width:8px;height:8px;border-radius:50%;background:var(--gold)}.year-text{font-size:clamp(18px,2vw,24px);font-weight:700;color:var(--gold);letter-spacing:.06em}
.tl-row{position:relative;padding:0 0 24px;animation:fadeUp .5s ease-out both}
@keyframes fadeUp{from{opacity:0;transform:translateY(12px)}to{opacity:1;transform:translateY(0)}}
.tl-left{margin-left:0;width:calc(50% - 28px);padding-right:0}.tl-right{margin-left:calc(50% + 28px);width:calc(50% - 28px);padding-left:0}
.tl-row::before{content:'';position:absolute;top:16px;width:10px;height:10px;border-radius:50%;background:var(--gold);border:2px solid var(--bg);z-index:1}
.tl-left::before{left:calc(100% + 28px);transform:translateX(-50%)}.tl-right::before{left:calc(0px - 28px);transform:translateX(-50%)}
.tl-card{transition:transform .25s,box-shadow .25s}.tl-card:hover{transform:translateY(-1px)}
.card-inner{display:block;text-decoration:none;padding:16px 20px;background:var(--tag-bg);border:1px solid var(--card-border);border-radius:4px;transition:box-shadow .3s}
.card-inner:hover{box-shadow:0 4px 20px rgba(120,105,81,.08)}
.card-head{display:flex;align-items:center;gap:8px;margin-bottom:6px}.card-date{font-size:11px;color:var(--gold)}
.card-dot{width:6px;height:6px;border-radius:50%}.dot-tech{background:var(--gold)}.dot-life{background:var(--muted)}.dot-read{background:var(--border)}.dot-default{background:var(--border)}
.card-title{font-size:15px;font-weight:700;color:var(--text);margin:0 0 6px;line-height:1.4}
.card-excerpt{font-size:12px;color:var(--muted);line-height:1.6;margin:0 0 8px}
.card-foot{display:flex;flex-wrap:wrap;gap:4px}.card-cat{font-size:9px;padding:1px 7px;background:rgba(184,148,76,.12);color:var(--gold);border-radius:2px}.card-tag{font-size:9px;padding:1px 6px;background:var(--tag-bg);color:var(--text2);border:1px solid var(--card-border);border-radius:2px}
</style>
