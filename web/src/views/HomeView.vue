<template>
  <div class="relative">
    <DanmakuLayer />

    <section class="w-full min-h-screen flex flex-col items-center justify-center relative z-[2]">
      <div class="text-center -mt-[8vh]">
        <StarIcon class="w-[18px] h-[18px] text-line mx-auto" />
        <p class="hero-typing">技术是笔，生活是墨。</p>
        <p class="hero-sub">—— 在这里写下两种颜色</p>
      </div>
      <a href="#main" class="hero-arrow" aria-label="向下浏览"><DownIcon class="w-5 h-5" /></a>
    </section>

    <div id="main" class="flex flex-col md:flex-row relative z-[2] gap-6 md:gap-12 min-h-[50vh] scroll-mt-16">
      <aside class="w-full md:w-[200px] shrink-0 flex flex-wrap md:flex-col items-center md:items-start gap-x-3.5 gap-y-2 border-b md:border-b-0 border-line2 pb-4 md:pb-0">
        <img v-if="authorAvatar" :src="authorAvatar" class="w-12 h-12 rounded-full object-cover border border-accent" alt="博主头像" />
        <div v-else class="w-12 h-12 rounded-full bg-paper2 border border-line flex items-center justify-center text-ink3">
          <UserIcon class="w-5 h-5" />
        </div>
        <div class="text-[15px] tracking-widest text-ink">{{ authorName }}</div>
        <div v-if="authorBio" class="basis-full text-xs text-ink3 leading-relaxed">{{ authorBio }}</div>
        <hr class="hidden md:block w-full border-0 border-t border-dotted border-line2 my-2">
        <div class="flex gap-4">
          <span class="text-[15px] text-ink">{{ articleCount }}<span class="text-[10px] text-ink3 ml-1">文章</span></span>
          <span class="text-[15px] text-ink">{{ noteCount }}<span class="text-[10px] text-ink3 ml-1">随笔</span></span>
        </div>
        <div class="text-xs"><a href="https://github.com" target="_blank" class="text-accent-strong no-underline mr-2.5 transition-colors duration-700 hover:text-ink">GitHub</a><a href="/feed.xml" class="text-accent-strong no-underline transition-colors duration-700 hover:text-ink">RSS</a></div>
      </aside>
      <main class="flex-1 min-w-0">
        <div class="text-xs text-ink3 tracking-[0.2em] uppercase mb-5">最近更新</div>
        <div v-if="!articles.length" class="text-center py-16">
          <EditIcon class="w-8 h-8 text-accent mx-auto mb-3" />
          <div class="text-lg font-light tracking-widest text-ink mb-1.5">墨还未干</div>
          <div class="text-[13px] text-ink3">第一篇文字正在路上。趁这个时间，泡杯茶吧</div>
        </div>
        <article v-for="(a,i) in articles" :key="a.id" class="post-card mb-3" :style="{animationDelay:i*0.12+'s'}">
          <router-link :to="'/posts/'+a.slug" class="group block no-underline p-5 md:p-6 bg-paper2 border border-line transition-colors duration-700 hover:border-accent">
            <time class="block text-[11px] text-accent-strong tracking-widest mb-1">{{ fmt(a.created_at) }}</time>
            <h3 class="text-base md:text-lg font-light tracking-wide text-ink mb-1 transition-colors duration-700 group-hover:text-accent-strong">{{ a.title }}</h3>
            <p class="text-[13px] text-ink3 leading-relaxed mb-2">{{ a.excerpt }}</p>
            <div class="flex flex-wrap gap-1.5" v-if="a.tags"><span v-for="t in splitTags(a.tags)" :key="t" class="text-[10px] px-2 py-px border border-line2 text-ink3">{{ t }}</span></div>
          </router-link>
        </article>
        <div v-if="articles.length" class="text-center mt-5"><router-link to="/articles" class="text-[13px] text-accent-strong no-underline border-b border-transparent transition-colors duration-700 hover:text-ink hover:border-accent/50">浏览全部文章 →</router-link></div>
      </main>
    </div>
  </div>
</template>

<script setup>
import { ref, computed, onMounted } from 'vue'
import { StarOutline, ChevronDownOutline, PersonOutline, CreateOutline } from '@vicons/ionicons5'
import DanmakuLayer from '../components/DanmakuLayer.vue'
import { listArticles } from '../api/articles'
import { listNotes } from '../api/notes'
import { useAuthor } from '../composables/useAuthor'
import { fmtDate } from '../utils/format'
const StarIcon=StarOutline, DownIcon=ChevronDownOutline, UserIcon=PersonOutline, EditIcon=CreateOutline
const articles=ref([]), articleCount=ref(0), noteCount=ref(0)
const { author } = useAuthor()
const authorName = computed(() => author.value.name), authorAvatar = computed(() => author.value.avatar), authorBio = computed(() => author.value.signature)
onMounted(async()=>{
  try{const[aD,nD]=await Promise.all([listArticles({per_page:8}),listNotes({per_page:50})]);articles.value=aD.articles||[];articleCount.value=aD.total||0;noteCount.value=nD.total||0}catch(e){}
})
function fmt(d){return fmtDate(d,{month:'2-digit',day:'2-digit'})}
function splitTags(t){return t?t.split(',').map(x=>x.trim()).filter(Boolean):[]}
</script>

<style scoped>
.hero-typing{display:inline-block;overflow:hidden;white-space:nowrap;border-right:1px solid var(--accent);font-size:clamp(26px,4vw,44px);font-weight:400;font-family:'LXGW WenKai',serif;color:var(--ink);letter-spacing:.08em;animation:type 2.8s steps(14).3s both;margin:10px 0 14px}
.hero-sub{font-size:clamp(13px,1.2vw,16px);color:var(--ink3);font-family:'LXGW WenKai',serif;opacity:0;animation:appear 1s 3s both}
.hero-arrow{position:absolute;bottom:clamp(60px,10vh,120px);width:36px;height:36px;display:flex;align-items:center;justify-content:center;color:var(--ink3);text-decoration:none;transition:color .7s ease-in-out;animation:breathe 4s ease-in-out infinite}
.hero-arrow:hover{color:var(--accent-strong)}
@keyframes type{from{width:0}to{width:100%}}
@keyframes appear{from{opacity:0}to{opacity:1}}
@keyframes breathe{0%,100%{opacity:1}50%{opacity:.4}}
.post-card{animation:fadeUp .7s ease-in-out both}
@keyframes fadeUp{from{opacity:0}to{opacity:1}}
/* 减少动态效果：打字/呼吸/渐入动画停用，直接呈现终态 */
@media (prefers-reduced-motion: reduce){
  .hero-typing{animation:none;border-right:none;width:auto}
  .hero-sub{animation:none;opacity:1}
  .hero-arrow{animation:none}
  .post-card{animation:none}
}
</style>
