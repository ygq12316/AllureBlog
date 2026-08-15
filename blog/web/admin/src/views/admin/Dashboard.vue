<template>
  <div class="dash">
    <!-- 欢迎 -->
    <div class="welcome">
      <img v-if="authorAvatar" :src="authorAvatar" class="welcome-img" />
      <div v-else class="avatar"><n-icon size="18" :component="PersonIcon" /></div>
      <div>
        <div class="w-name">欢迎回来，{{ authorName }}</div>
        <div v-if="authorSig" class="w-sub">{{ authorSig }}</div>
      </div>
    </div>

    <!-- 统计 -->
    <div class="stats">
      <div class="stat" v-for="s in statList" :key="s.label">
        <n-icon size="18" color="var(--gold)" :component="s.icon" />
        <div class="stat-num">{{ s.value }}</div>
        <div class="stat-label">{{ s.label }}</div>
      </div>
    </div>

    <!-- 快捷 -->
    <div class="quick">
      <n-button type="primary" @click="$router.push('/admin/articles/new')"><n-icon :component="CreateIcon" /> 写文章</n-button>
      <n-button @click="$router.push('/admin/notes/new')"><n-icon :component="ChatIcon" /> 写随笔</n-button>
    </div>

    <!-- 最近 -->
    <div class="recent">
      <div class="col">
        <div class="col-title">最近文章</div>
        <div v-if="!articles.length" class="col-empty">暂无</div>
        <div v-for="a in articles" :key="a.id" class="col-item" @click="$router.push(`/admin/articles/${a.id}/edit`)">
          <span class="ci-title">{{ a.title }}</span><span class="ci-date">{{ fmt(a.created_at) }}</span>
        </div>
      </div>
      <div class="col">
        <div class="col-title">最近随笔</div>
        <div v-if="!nlist.length" class="col-empty">暂无</div>
        <div v-for="n in nlist" :key="n.id" class="col-item" @click="$router.push('/admin/notes')">
          <span class="ci-title" v-html="trunc(n.html,24)" /><span class="ci-date">{{ rel(n.created_at) }}</span>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, computed, onMounted } from 'vue'
import { PersonOutline, DocumentTextOutline, ChatbubbleOutline, FolderOpenOutline, PricetagsOutline, CreateOutline } from '@vicons/ionicons5'
import axios from 'axios'

const PersonIcon=PersonOutline, DocIcon=DocumentTextOutline, ChatIcon=ChatbubbleOutline, FolderIcon=FolderOpenOutline, PricetagIcon=PricetagsOutline, CreateIcon=CreateOutline

const stats=ref({article_count:0,note_count:0,category_count:0,tag_count:0})
const articles=ref([]), nlist=ref([])
const authorName=ref(''), authorAvatar=ref(''), authorSig=ref('')

const statList=computed(()=>[
  {icon:DocIcon, value:stats.value.article_count, label:'文章'},
  {icon:ChatIcon, value:stats.value.note_count, label:'随笔'},
  {icon:FolderIcon, value:stats.value.category_count||0, label:'分类'},
  {icon:PricetagIcon, value:stats.value.tag_count||0, label:'标签'},
])

onMounted(async()=>{
  try{stats.value=(await axios.get('/api/stats')).data}catch(e){}
  try{articles.value=(await axios.get('/api/articles?all=true&per_page=3')).data.articles||[]}catch(e){}
  try{nlist.value=(await axios.get('/api/notes?all=true&per_page=3')).data.notes||[]}catch(e){}
  try{stats.value.tag_count=(await axios.get('/api/tags')).data.tags?.length||0}catch(e){}
  try{const{data}=await axios.get('/api/config');const c=data.config;authorName.value=c.author_name||'Allure';authorAvatar.value=c.author_avatar||''}catch(e){}
  try{const{data}=await axios.get('/api/visitor/admin_admin');authorSig.value=data.visitor?.signature||''}catch(e){}
})

function fmt(d){return d?new Date(d).toLocaleDateString('zh-CN',{month:'short',day:'numeric'}):''}
function trunc(h,n){const t=h.replace(/<[^>]*>/g,'');return t.length>n?t.slice(0,n)+'…':t}
function rel(d){const s=Math.floor((Date.now()-new Date(d).getTime())/1000);if(s<60)return'刚刚';if(s<3600)return Math.floor(s/60)+'m';if(s<86400)return Math.floor(s/3600)+'h';if(s<2592000)return Math.floor(s/86400)+'d';return fmt(d)}
</script>

<style scoped>
.dash{max-width:680px;margin:0 auto}
.welcome{display:flex;align-items:center;gap:12px;margin-bottom:28px}
.avatar{width:40px;height:40px;border-radius:50%;background:var(--tag-bg);border:2px solid var(--gold);display:flex;align-items:center;justify-content:center;color:var(--gold)}
.welcome-img{width:40px;height:40px;border-radius:50%;object-fit:cover;border:2px solid var(--gold)}
.w-name{font-size:17px;font-weight:700;color:var(--text)}.w-sub{font-size:12px;color:var(--muted)}

.stats{display:grid;grid-template-columns:repeat(4,1fr);gap:12px;margin-bottom:24px}
.stat{text-align:center;padding:12px 8px}.stat-num{font-size:24px;font-weight:700;color:var(--text)}.stat-label{font-size:11px;color:var(--muted);margin-top:2px}

.quick{display:flex;gap:10px;margin-bottom:28px}.quick .n-button{flex:1;height:40px}

.recent{display:grid;grid-template-columns:1fr 1fr;gap:20px}
.col-title{font-size:13px;font-weight:700;color:var(--text);margin-bottom:10px}
.col-empty{font-size:12px;color:var(--muted);padding:16px 0;text-align:center}
.col-item{display:flex;justify-content:space-between;align-items:center;padding:8px 0;border-bottom:1px solid var(--card-border);cursor:pointer;font-size:13px}.col-item:hover{color:var(--gold)}
.ci-title{overflow:hidden;text-overflow:ellipsis;white-space:nowrap;flex:1;margin-right:8px;color:var(--text)}.ci-date{font-size:11px;color:var(--muted);white-space:nowrap}
</style>
