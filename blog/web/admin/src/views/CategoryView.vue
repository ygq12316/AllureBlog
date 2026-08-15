<template>
  <div class="page">
    <div class="cat-header"><h2 class="cat-title">分类</h2><p class="cat-sub">旋转标签云 · 点击浏览</p></div>
    <TagCloud v-if="cloudTags.length" :tags="cloudTags" :active="slug" @select="onSelect" />
    <div v-if="!categories.length" class="cat-empty"><n-icon size="40" :component="FolderIcon" color="var(--muted)" /><p>暂无分类</p></div>

    <div v-if="slug" class="cat-list">
      <div class="list-head"><n-icon :component="LayersIcon" color="var(--gold)" /><span class="list-label">{{ slug }}</span><span class="list-count">· {{ articles.length }} 篇</span><n-button text size="small" @click="clearCat">清除筛选</n-button></div>
      <div v-if="!articles.length" class="list-empty"><n-icon size="24" :component="DocIcon" /><span>该分类下还没有文章</span></div>
      <div v-for="(a,i) in articles" :key="a.id" class="list-item" :style="{animationDelay:i*.06+'s'}">
        <n-icon class="item-icon" :component="BookIcon" /><time class="item-date">{{ md(a.created_at) }}</time>
        <div class="item-main"><router-link :to="'/posts/'+a.slug" class="item-title">{{ a.title }}</router-link><div class="item-tags" v-if="a.tags"><span v-for="t in splitTags(a.tags)" :key="t" class="item-tag">{{ t }}</span></div></div>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, computed, onMounted, watch } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { FolderOpenOutline, LayersOutline, DocumentOutline, BookOutline } from '@vicons/ionicons5'
import TagCloud from '../components/TagCloud.vue'
import axios from 'axios'
const FolderIcon=FolderOpenOutline, LayersIcon=LayersOutline, DocIcon=DocumentOutline, BookIcon=BookOutline
const route=useRoute(), router=useRouter(), slug=ref(route.params.slug||''), categories=ref([]), articles=ref([])
const cloudTags=computed(()=>categories.value.map(c=>({name:c.name,slug:c.slug,count:c.article_count})))

onMounted(async()=>{
  try{const{data}=await axios.get('/api/categories');categories.value=data.categories||[]}catch(e){}
  loadArticles()
})
watch(()=>route.params.slug,s=>{slug.value=s||'';loadArticles()})
function onSelect(s){router.push('/category/'+s)};function clearCat(){router.push('/category')}
async function loadArticles(){if(!slug.value){articles.value=[];return};try{const{data}=await axios.get('/api/articles?category='+slug.value+'&per_page=50');articles.value=data.articles||[]}catch(e){articles.value=[]}}
function md(d){return d?new Date(d).toLocaleDateString('zh-CN',{month:'2-digit',day:'2-digit'}):''}
function splitTags(t){return t?t.split(',').map(x=>x.trim()).filter(Boolean):[]}
</script>

<style scoped>
.page{padding:0;text-align:center;min-height:60vh}
.cat-header{margin-bottom:8px}.cat-title{font-size:clamp(22px,2.5vw,28px);font-weight:700;color:var(--text);margin:0 0 4px}.cat-sub{font-size:clamp(11px,.9vw,13px);color:var(--muted);margin:0 0 16px}
.cat-empty{padding:60px 0;color:var(--muted)}.cat-empty p{margin-top:8px;font-size:13px}
.cat-list{margin-top:24px;text-align:left;max-width:640px;margin-left:auto;margin-right:auto}
.list-head{display:flex;align-items:center;gap:8px;margin-bottom:16px;font-size:clamp(14px,1.1vw,16px)}.list-label{font-weight:700;color:var(--text);text-transform:capitalize}.list-count{font-size:12px;color:var(--muted)}
.list-empty{display:flex;align-items:center;gap:8px;justify-content:center;padding:40px 0;font-size:13px;color:var(--muted)}
.list-item{display:flex;align-items:flex-start;gap:10px;padding:12px 0;border-bottom:1px solid var(--card-border);animation:fadeUp .5s ease-out both}
@keyframes fadeUp{from{opacity:0;transform:translateY(10px)}to{opacity:1;transform:translateY(0)}}
.item-icon{margin-top:2px;color:var(--gold)}.item-date{font-size:11px;color:var(--gold);white-space:nowrap;min-width:40px}
.item-main{flex:1;min-width:0}.item-title{font-size:clamp(15px,1.3vw,18px);font-weight:700;color:var(--text);text-decoration:none;display:block;margin-bottom:4px}.item-title:hover{color:var(--gold)}
.item-tags{display:flex;gap:4px;flex-wrap:wrap}.item-tag{font-size:11px;padding:2px 8px;background:var(--tag-bg);color:var(--text2);border:1px solid var(--border);border-radius:2px}

</style>
