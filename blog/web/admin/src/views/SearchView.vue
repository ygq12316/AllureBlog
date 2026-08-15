<template>
  <div class="page">
    <div class="search-hero">
      <n-icon size="40" color="var(--gold)" :component="SearchIcon" />
      <h2 class="search-title">搜索文章</h2>
      <form @submit.prevent="doSearch" class="search-form">
        <div class="search-box" :class="{'search-box--focused':focused}">
          <n-icon size="18" color="var(--muted)" :component="SearchIcon" />
          <input v-model="q" placeholder="试试搜 Go、咖啡、徒步..." class="search-input" @focus="focused=true" @blur="focused=false" />
          <button v-if="q" @click="q='';searched=false" type="button" class="search-clear"><n-icon size="14" :component="CloseIcon" /></button>
        </div>
      </form>
    </div>
    <div v-if="searched" class="search-results">
      <div v-if="results.length" class="results-count">
        <n-icon size="14" color="var(--gold)" :component="BookIcon" /> 找到 <strong>{{ results.length }}</strong> 篇关于 "<em>{{ q }}</em>" 的文章
      </div>
      <div v-for="a in results" :key="a.id" class="result-item">
        <div class="result-meta"><n-icon size="13" color="var(--gold)" :component="TimeIcon" /><time>{{ fmt(a.created_at) }}</time><span v-if="a.category" class="result-cat">{{ a.category }}</span></div>
        <router-link :to="'/posts/'+a.slug" class="result-title">{{ a.title }}</router-link>
        <p class="result-excerpt">{{ a.excerpt }}</p>
      </div>
      <div v-if="!results.length" class="no-results">
        <n-icon size="48" color="var(--border)" :component="TrashIcon" />
        <div class="no-results-title">没有找到 "{{ q }}"</div><div class="no-results-desc">换个关键词，或试试下面的标签</div>
        <WordCloud :tags="allTags" @select="q=$event;doSearch()" />
      </div>
    </div>
    <div v-else class="search-browse">
      <WordCloud :tags="allTags" @select="q=$event;doSearch()" />
    </div>
  </div>
</template>

<script setup>
import { ref, onMounted } from 'vue'
import { useRoute } from 'vue-router'
import { SearchOutline, CloseOutline, BookOutline, TimeOutline, TrashOutline } from '@vicons/ionicons5'
import WordCloud from '../components/WordCloud.vue'
import axios from 'axios'

const SearchIcon=SearchOutline, CloseIcon=CloseOutline, BookIcon=BookOutline, TimeIcon=TimeOutline, TrashIcon=TrashOutline
const route=useRoute(), q=ref(route.query.q||''), results=ref([]), searched=ref(false), focused=ref(false)
const allTags=ref([])

if(q.value){searched.value=true;doSearch()}

onMounted(async()=>{
  try{const{data}=await axios.get('/api/tags');allTags.value=data.tags||[]}catch(e){}
})

async function doSearch(){if(!q.value.trim())return;searched.value=true;try{const{data}=await axios.get('/api/articles/search?q='+encodeURIComponent(q.value));results.value=data.articles||[]}catch(e){results.value=[]}}
function fmt(d){return d?new Date(d).toLocaleDateString('zh-CN',{month:'short',day:'numeric'}):''}
</script>

<style scoped>
.page{padding:0;max-width:600px;margin:0 auto;text-align:center}
.search-hero{margin-bottom:40px}.search-title{font-size:clamp(22px,2.5vw,28px);font-weight:700;color:var(--text);margin:12px 0 24px}
.search-form{width:100%}.search-box{display:flex;align-items:center;background:var(--bg);border:2px solid var(--card-border);border-radius:50px;padding:12px 20px;transition:all .3s;gap:8px}
.search-box--focused{border-color:var(--gold);box-shadow:0 0 0 4px rgba(184,148,76,.1)}
.search-input{flex:1;border:none;background:transparent;font-family:'LXGW WenKai',serif;font-size:16px;color:var(--text);outline:none}.search-input::placeholder{color:var(--muted)}
.search-clear{background:none;border:none;color:var(--muted);cursor:pointer;padding:4px;display:flex;align-items:center}.search-clear:hover{color:var(--text)}
.search-results{text-align:left}.results-count{font-size:12px;color:var(--muted);margin-bottom:20px;text-align:center;display:flex;align-items:center;justify-content:center;gap:6px}
.results-count em{font-style:normal;color:var(--gold)}.result-item{padding:16px 0;border-bottom:1px solid var(--card-border)}
.result-meta{display:flex;align-items:center;gap:6px;margin-bottom:6px}.result-meta time{font-size:11px;color:var(--gold)}.result-cat{font-size:9px;padding:1px 7px;background:var(--tag-bg);color:var(--text2);border:1px solid var(--border);border-radius:2px}
.result-title{font-size:15px;font-weight:700;color:var(--text);text-decoration:none;display:block;margin-bottom:4px}.result-title:hover{color:var(--gold)}
.result-excerpt{font-size:12px;color:var(--muted);margin:0;line-height:1.6}
.no-results{text-align:center;padding:40px 0}.no-results-title{font-size:16px;font-weight:700;color:var(--text);margin:12px 0 8px}.no-results-desc{font-size:12px;color:var(--muted);margin-bottom:20px}
.search-browse{margin-top:12px}
</style>
