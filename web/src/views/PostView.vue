<template>
  <div class="page">
    <router-link to="/" class="back-link">← 返回</router-link>
    <div v-if="loading" class="state-box">加载中...</div>
    <div v-else-if="article" class="post-layout">
      <div class="post-main" ref="postMain">
        <div class="post-meta">{{ fmt(article.created_at) }}<span v-if="article.category"> · {{ article.category }}</span></div>
        <h1 class="post-title">{{ article.title }}</h1>
        <div class="post-tags" v-if="article.tags"><span v-for="t in split(article.tags)" :key="t" class="tag">{{ t }}</span></div>
        <div ref="postBody" class="post-body" v-html="article.html" /><hr class="post-divider">
      </div>
      <aside class="post-sidebar" v-if="toc.length">
        <div class="toc-header">目录</div>
        <nav class="toc-nav">
          <a v-for="h in toc" :key="h.id" :href="'#'+h.id" class="toc-link" :class="{'toc-link--active':activeId===h.id,'toc-link--h3':h.level===3}" @click.prevent="scrollTo(h.id)">{{ h.text }}</a>
        </nav>
      </aside>
    </div>
    <div v-else class="state-box">
      <n-icon size="28" :component="TrashIcon" />
      <div class="state-title">这张纸是空的</div><div class="state-desc">找不到这篇文章</div>
      <div class="state-actions"><router-link to="/" class="btn-outline">← 返回首页</router-link><router-link to="/search" class="btn-solid">去搜索</router-link></div>
    </div>
  </div>
</template>
<script setup>
import { ref, onMounted, nextTick, onUnmounted } from 'vue'
import { useRoute } from 'vue-router'
import { TrashOutline } from '@vicons/ionicons5'
import { getArticleBySlug } from '../api/articles'
const TrashIcon=TrashOutline
const route=useRoute(), article=ref(null), loading=ref(true), postBody=ref(null), toc=ref([]), activeId=ref('')
let observer=null
onMounted(async()=>{try{const data=await getArticleBySlug(route.params.slug);article.value=data.article||data.articles?.[0]||null}catch(e){article.value=null};loading.value=false;await nextTick();buildTOC()})
function buildTOC(){if(!postBody.value)return;const headings=postBody.value.querySelectorAll('h2,h3'),items=[];let h2I=0,h3I=0;headings.forEach(h=>{const level=h.tagName==='H2'?2:3,id=h.id||(level===2?'s'+(++h2I):'s'+h2I+'-'+(++h3I));if(!h.id)h.id=id;items.push({id,text:h.textContent,level})});toc.value=items;observer=new IntersectionObserver(entries=>{for(const e of entries){if(e.isIntersecting){activeId.value=e.target.id;break}}},{rootMargin:'-80px 0px -70% 0px'});headings.forEach(h=>observer.observe(h))}
function scrollTo(id){const el=document.getElementById(id);if(el)el.scrollIntoView({behavior:'smooth',block:'start'})}
onUnmounted(()=>{if(observer)observer.disconnect()})
function fmt(d){return d?new Date(d).toLocaleDateString('zh-CN',{year:'numeric',month:'long',day:'numeric'}):''}
function split(t){return t?t.split(',').map(x=>x.trim()).filter(Boolean):[]}
</script>
<style scoped>
.page{padding:0}.back-link{font-size:12px;color:var(--muted);text-decoration:none;display:inline-block;margin-bottom:8px}.back-link:hover{color:var(--gold)}
.post-layout{display:flex;gap:clamp(20px,4vw,48px);margin-top:16px}.post-main{flex:7;min-width:0}
.post-meta{font-size:11px;color:var(--gold);letter-spacing:.1em;margin-bottom:8px}
.post-title{font-size:clamp(22px,2.5vw,32px);font-weight:700;color:var(--text);line-height:1.3;margin-bottom:12px}
.post-tags{display:flex;gap:5px;margin-bottom:24px;flex-wrap:wrap}
.tag{font-size:10px;padding:2px 8px;background:var(--tag-bg);color:var(--text2);border:1px solid var(--border);border-radius:2px}
.post-body{line-height:2.1;font-size:clamp(14px,1.05vw,16px);color:var(--text)}
.post-body :deep(pre){background:var(--tag-bg);padding:16px 20px;border-radius:4px;overflow-x:auto;font-size:13px}
.post-body :deep(code){font-family:'JetBrains Mono',monospace;font-size:13px}
.post-body :deep(blockquote){border-left:3px solid var(--gold);margin:16px 0;padding-left:14px;color:var(--muted)}
.post-body :deep(a){color:var(--gold)}.post-body :deep(h2),.post-body :deep(h3){color:var(--text);margin-top:28px;margin-bottom:10px;scroll-margin-top:80px}
.post-body :deep(img){max-width:100%;border-radius:4px}.post-divider{border:none;border-top:1px dashed var(--border);margin:clamp(24px,5vh,40px) 0}
.post-sidebar{flex:3;min-width:0;padding:20px 18px;background:transparent;border-left:1px solid var(--card-border);position:sticky;top:80px;align-self:flex-start;max-height:calc(100vh - 120px);overflow-y:auto}
.toc-header{font-size:13px;font-weight:700;color:var(--text);margin-bottom:14px;letter-spacing:.05em}
.toc-nav{display:flex;flex-direction:column}
.toc-link{font-size:12px;color:var(--muted);text-decoration:none;padding:5px 0;border-left:2px solid transparent;transition:color .2s,border-color .2s;line-height:1.5}
.toc-link:hover{color:var(--gold)}.toc-link--h3{padding-left:16px;font-size:11px}
.toc-link--active{color:var(--gold);border-left-color:var(--gold);font-weight:600;padding-left:10px}
.toc-link--active.toc-link--h3{padding-left:24px}
.state-box{text-align:center;padding:80px 20px}.state-title{font-size:18px;font-weight:700;color:var(--text);margin:10px 0 6px}.state-desc{font-size:13px;color:var(--muted);margin-bottom:20px}
.state-actions{display:flex;gap:10px;justify-content:center}
.btn-outline{padding:6px 16px;font-size:12px;background:var(--tag-bg);color:var(--text2);border:1px solid var(--border);border-radius:2px;text-decoration:none}
.btn-solid{padding:6px 16px;font-size:12px;background:var(--gold);color:#fff;border-radius:2px;text-decoration:none}
</style>
