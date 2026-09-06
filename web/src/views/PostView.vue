<template>
  <div>
    <router-link to="/" class="text-xs text-ink3 no-underline transition-colors duration-700 hover:text-accent-strong">← 返回</router-link>
    <div v-if="loading" class="text-center py-20 text-sm text-ink3">加载中...</div>
    <div v-else-if="article" class="flex flex-col md:flex-row gap-5 md:gap-12 mt-4">
      <div class="flex-[7] min-w-0" ref="postMain">
        <div class="text-[11px] text-accent-strong tracking-[0.1em] mb-2">{{ fmt(article.created_at) }}<span v-if="article.category"> · {{ article.category }}</span></div>
        <h1 class="text-2xl md:text-4xl font-light tracking-wide text-ink leading-snug mb-3 mt-0">{{ article.title }}</h1>
        <div class="flex gap-1.5 mb-6 flex-wrap" v-if="article.tags"><span v-for="t in split(article.tags)" :key="t" class="text-[10px] px-2 py-px border border-line2 text-ink3">{{ t }}</span></div>
        <div ref="postBody" class="post-body" v-html="article.html" /><hr class="border-0 border-t border-dashed border-line2 my-8 md:my-12">
      </div>
      <aside class="md:flex-[3] md:min-w-0 sticky top-20 self-start max-h-[calc(100vh-120px)] overflow-y-auto py-5 px-0 md:px-4.5 md:border-l border-line" v-if="toc.length">
        <div class="text-[13px] tracking-widest text-ink mb-3.5">目录</div>
        <nav class="flex flex-col">
          <a v-for="h in toc" :key="h.id" :href="'#'+h.id"
            class="text-xs no-underline py-1 border-l transition-colors duration-500 leading-relaxed"
            :class="[
              activeId===h.id
                ? 'text-accent-strong border-accent pl-2.5'
                : 'text-ink3 border-transparent hover:text-accent-strong pl-0',
              h.level===3 ? 'text-[11px]' : '',
            ]"
            @click.prevent="scrollTo(h.id)">
            <span :class="h.level===3 ? 'pl-4 block' : 'block'">{{ h.text }}</span>
          </a>
        </nav>
      </aside>
    </div>
    <div v-else class="text-center py-20">
      <TrashIcon class="w-7 h-7 text-ink3 mx-auto" />
      <div class="text-lg font-light tracking-widest text-ink mt-2.5 mb-1.5">这张纸是空的</div>
      <div class="text-[13px] text-ink3 mb-5">找不到这篇文章</div>
      <div class="flex gap-2.5 justify-center">
        <router-link to="/" class="px-4 py-1.5 text-xs no-underline border border-line text-ink2 transition-colors duration-700 hover:border-accent hover:text-accent-strong">← 返回首页</router-link>
        <router-link to="/search" class="px-4 py-1.5 text-xs no-underline bg-ink text-paper transition-opacity duration-700">去搜索</router-link>
      </div>
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
.post-body{line-height:2.1;font-size:clamp(15px,1.05vw,17px);color:var(--ink);font-weight:400}
.post-body :deep(pre){background:var(--paper2);border:1px solid var(--line);padding:16px 20px;overflow-x:auto;font-size:13px}
.post-body :deep(code){font-family:'JetBrains Mono',ui-monospace,monospace;font-size:13px}
.post-body :deep(blockquote){border-left:2px solid var(--color-moss,#6b7b6e);margin:16px 0;padding-left:14px;color:var(--ink2)}
.post-body :deep(a){color:var(--accent-strong);border-bottom:1px solid color-mix(in srgb,var(--accent) 40%,transparent);text-decoration:none;transition:color .7s ease-in-out}
.post-body :deep(a:hover){color:var(--ink)}
.post-body :deep(h2),.post-body :deep(h3){color:var(--ink);font-weight:400;letter-spacing:.05em;margin-top:28px;margin-bottom:10px;scroll-margin-top:80px}
.post-body :deep(h2){font-size:1.35em}
.post-body :deep(h3){font-size:1.15em}
.post-body :deep(img){max-width:100%}
</style>
