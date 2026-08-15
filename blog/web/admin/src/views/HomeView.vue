<template>
  <div style="position:relative">
    <DanmakuLayer />

    <section class="hero-block">
      <div class="hero-center">
        <n-icon size="18" color="var(--border)" :component="StarIcon" />
        <p class="hero-typing">技术是笔，生活是墨。</p>
        <p class="hero-sub">—— 在这里写下两种颜色</p>
      </div>
      <a href="#main" class="hero-arrow"><n-icon :component="DownIcon" /></a>
    </section>

    <div id="main" class="main-content">
      <aside class="profile-sidebar">
        <img v-if="authorAvatar" :src="authorAvatar" class="profile-img" />
        <div v-else class="profile-avatar"><n-icon size="20" :component="UserIcon" /></div>
        <div class="profile-name">{{ authorName }}</div>
        <div v-if="authorBio" class="profile-bio">{{ authorBio }}</div>
        <hr class="profile-divider">
        <div class="profile-stats"><div><strong>{{ articleCount }}</strong><span>文章</span></div><div><strong>{{ noteCount }}</strong><span>随笔</span></div></div>
        <div class="profile-links"><a href="https://github.com" target="_blank">GitHub</a><a href="/feed.xml">RSS</a></div>
      </aside>
      <main class="articles-main">
        <div class="articles-header">最近更新</div>
        <div v-if="!articles.length" class="empty-block">
          <n-icon size="32" color="var(--gold)" :component="EditIcon" />
          <div class="empty-title">墨还未干</div>
          <div class="empty-desc">第一篇文字正在路上。趁这个时间，泡杯茶吧</div>
        </div>
        <article v-for="(a,i) in articles" :key="a.id" class="post-card" :style="{animationDelay:i*0.1+'s'}">
          <router-link :to="'/posts/'+a.slug" class="post-card-link">
            <time class="post-card-date">{{ fmt(a.created_at) }}</time>
            <h3 class="post-card-title">{{ a.title }}</h3>
            <p class="post-card-excerpt">{{ a.excerpt }}</p>
            <div class="post-card-tags" v-if="a.tags"><span v-for="t in splitTags(a.tags)" :key="t" class="post-tag">{{ t }}</span></div>
          </router-link>
        </article>
        <div v-if="articles.length" class="post-card-more"><router-link to="/articles">浏览全部文章 →</router-link></div>
      </main>
    </div>
  </div>
</template>

<script setup>
import { ref, onMounted } from 'vue'
import { StarOutline, ChevronDownOutline, PersonOutline, CreateOutline } from '@vicons/ionicons5'
import DanmakuLayer from '../components/DanmakuLayer.vue'
import axios from 'axios'
const StarIcon=StarOutline, DownIcon=ChevronDownOutline, UserIcon=PersonOutline, EditIcon=CreateOutline
const articles=ref([]), articleCount=ref(0), noteCount=ref(0), authorName=ref('Allure'), authorAvatar=ref(''), authorBio=ref('')
onMounted(async()=>{
  try{const[aR,nR]=await Promise.all([axios.get('/api/articles?per_page=8'),axios.get('/api/notes?per_page=50')]);articles.value=aR.data.articles||[];articleCount.value=aR.data.total||0;noteCount.value=nR.data.total||0}catch(e){}
  try{const{data}=await axios.get('/api/config');const cfg=data.config;authorName.value=cfg.author_name||'Allure';authorAvatar.value=cfg.author_avatar||''}catch(e){}
  try{const{data}=await axios.get('/api/visitor/admin_admin');authorBio.value=data.visitor?.signature||''}catch(e){}
})
function fmt(d){return d?new Date(d).toLocaleDateString('zh-CN',{month:'2-digit',day:'2-digit'}):''}
function splitTags(t){return t?t.split(',').map(x=>x.trim()).filter(Boolean):[]}
</script>

<style scoped>
.hero-block{width:100%;min-height:100vh;display:flex;flex-direction:column;align-items:center;justify-content:center;position:relative;z-index:2;background:transparent}
.hero-center{text-align:center;margin-top:-8vh}.hero-typing{display:inline-block;overflow:hidden;white-space:nowrap;border-right:2px solid var(--gold);font-size:clamp(22px,3.5vw,34px);font-weight:700;color:var(--text);animation:type 2.8s steps(14).3s both;margin-bottom:12px}
.hero-sub{font-size:clamp(13px,1.2vw,16px);color:var(--muted);opacity:0;animation:appear .8s 3s both}
.hero-arrow{position:absolute;bottom:clamp(60px,10vh,120px);width:36px;height:36px;background:var(--tag-bg);border:2px solid var(--gold);border-radius:50%;display:flex;align-items:center;justify-content:center;color:var(--gold);text-decoration:none;opacity:1;animation:bounce 1.8s ease-in-out infinite}
@keyframes type{from{width:0}to{width:100%}}
@keyframes appear{from{opacity:0}to{opacity:1}}
@keyframes bounce{0%,100%{transform:translateY(0)}50%{transform:translateY(6px)}}
.main-content{display:flex;position:relative;z-index:2;padding:0;gap:clamp(24px,5vw,56px);min-height:50vh}
.profile-sidebar{width:clamp(160px,18vw,220px);flex-shrink:0}
.profile-avatar{width:48px;height:48px;border-radius:50%;background:var(--tag-bg);border:2px solid var(--border);display:flex;align-items:center;justify-content:center;margin-bottom:12px;color:var(--gold)}
.profile-img{width:48px;height:48px;border-radius:50%;object-fit:cover;border:2px solid var(--gold);margin-bottom:12px}
.profile-name{font-size:clamp(14px,1.1vw,16px);font-weight:700;color:var(--text);margin-bottom:2px}
.profile-bio{font-size:clamp(10px,.85vw,12px);color:var(--muted);line-height:1.6;margin-bottom:12px}
.profile-divider{border:none;border-top:1px dotted var(--card-border);margin-bottom:12px}
.profile-stats{display:flex;gap:16px;margin-bottom:12px}
.profile-stats strong{font-size:clamp(14px,1.1vw,16px);color:var(--text);margin-right:2px}
.profile-stats span{font-size:clamp(9px,.75vw,11px);color:var(--muted)}
.profile-links{font-size:clamp(10px,.85vw,12px)}.profile-links a{color:var(--gold);text-decoration:none;margin-right:10px}
.articles-main{flex:1;min-width:0}.articles-header{font-size:clamp(10px,.85vw,12px);color:var(--muted);letter-spacing:2px;margin-bottom:16px;text-transform:uppercase}
.post-card{margin-bottom:12px;animation:fadeUp .6s ease-out both}
@keyframes fadeUp{from{opacity:0;transform:translateY(16px)}to{opacity:1;transform:translateY(0)}}
.post-card-link{display:block;text-decoration:none;padding:clamp(14px,2vw,20px);background:var(--tag-bg);border:1px solid var(--card-border);border-radius:2px;transition:box-shadow .35s,transform .25s}
.post-card-link:hover{box-shadow:0 8px 32px rgba(120,105,81,.12);transform:translateY(-2px)}
.post-card-date{font-size:clamp(9px,.75vw,11px);color:var(--gold);letter-spacing:.08em;display:block;margin-bottom:4px}
.post-card-title{font-size:clamp(14px,1.1vw,17px);font-weight:700;color:var(--text);margin-bottom:4px;transition:color .2s}
.post-card-link:hover .post-card-title{color:var(--gold)}
.post-card-excerpt{font-size:clamp(11px,.9vw,13px);color:var(--muted);line-height:1.6;margin-bottom:8px}
.post-card-tags{display:flex;flex-wrap:wrap;gap:5px}
.post-tag{font-size:clamp(8px,.65vw,10px);padding:2px 8px;background:var(--tag-bg);color:var(--text2);border:1px solid var(--border);border-radius:2px}
.post-card-more{text-align:center;margin-top:16px}.post-card-more a{font-size:clamp(11px,.9vw,13px);color:var(--gold);text-decoration:none}
.empty-block{text-align:center;padding:60px 20px}.empty-icon{margin-bottom:12px;color:var(--gold)}.empty-title{font-size:18px;font-weight:700;color:var(--text);margin-bottom:6px}.empty-desc{font-size:13px;color:var(--muted)}
</style>
