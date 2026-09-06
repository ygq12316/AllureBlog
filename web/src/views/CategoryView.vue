<template>
  <div class="text-center min-h-[60vh]">
    <div class="mb-2"><h2 class="text-xl md:text-2xl font-light tracking-[0.2em] text-ink m-0 mb-1">分类</h2><p class="text-xs text-ink3 m-0 mb-4">旋转标签云 · 点击浏览</p></div>
    <TagCloud v-if="cloudTags.length" :tags="cloudTags" :active="slug" @select="onSelect" />
    <div v-if="!categories.length" class="py-16 text-ink3"><FolderIcon class="w-10 h-10 mx-auto" /><p class="mt-2 text-[13px]">暂无分类</p></div>

    <div v-if="slug" class="mt-6 text-left max-w-[640px] mx-auto">
      <div class="flex items-center gap-2 mb-4 text-[15px]">
        <LayersIcon class="w-4 h-4 text-accent" />
        <span class="tracking-widest text-ink capitalize">{{ slug }}</span>
        <span class="text-xs text-ink3">· {{ articles.length }} 篇</span>
        <InkButton variant="link" size="xs" @click="clearCat">清除筛选</InkButton>
      </div>
      <div v-if="!articles.length" class="flex items-center gap-2 justify-center py-10 text-[13px] text-ink3">
        <DocIcon class="w-6 h-6" /><span>该分类下还没有文章</span>
      </div>
      <div v-for="(a,i) in articles" :key="a.id" class="flex items-start gap-2.5 py-3 border-b border-line2 list-item" :style="{animationDelay:i*.06+'s'}">
        <BookIcon class="w-4 h-4 mt-0.5 text-accent shrink-0" />
        <time class="text-[11px] text-accent-strong whitespace-nowrap min-w-[40px] mt-1">{{ md(a.created_at) }}</time>
        <div class="flex-1 min-w-0">
          <router-link :to="'/posts/'+a.slug" class="block text-base font-light tracking-wide text-ink no-underline mb-1 transition-colors duration-700 hover:text-accent-strong">{{ a.title }}</router-link>
          <div class="flex gap-1 flex-wrap" v-if="a.tags"><span v-for="t in splitTags(a.tags)" :key="t" class="text-[10px] px-1.5 py-px border border-line2 text-ink3">{{ t }}</span></div>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, computed, onMounted, watch } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { FolderOpenOutline, LayersOutline, DocumentOutline, BookOutline } from '@vicons/ionicons5'
import TagCloud from '../components/TagCloud.vue'
import InkButton from '../components/ui/InkButton.vue'
import { listCategories } from '../api/categories'
import { listArticles } from '../api/articles'
const FolderIcon=FolderOpenOutline, LayersIcon=LayersOutline, DocIcon=DocumentOutline, BookIcon=BookOutline
const route=useRoute(), router=useRouter(), slug=ref(route.params.slug||''), categories=ref([]), articles=ref([])
const cloudTags=computed(()=>categories.value.map(c=>({name:c.name,slug:c.slug,count:c.article_count})))

onMounted(async()=>{
  try{categories.value=await listCategories()}catch(e){}
  loadArticles()
})
watch(()=>route.params.slug,s=>{slug.value=s||'';loadArticles()})
function onSelect(s){router.push('/category/'+s)};function clearCat(){router.push('/category')}
async function loadArticles(){if(!slug.value){articles.value=[];return};try{const data=await listArticles({category:slug.value,per_page:50});articles.value=data.articles||[]}catch(e){articles.value=[]}}
function md(d){return d?new Date(d).toLocaleDateString('zh-CN',{month:'2-digit',day:'2-digit'}):''}
function splitTags(t){return t?t.split(',').map(x=>x.trim()).filter(Boolean):[]}
</script>

<style scoped>
.list-item{animation:fadeUp .7s ease-in-out both}
@keyframes fadeUp{from{opacity:0}to{opacity:1}}
@media (prefers-reduced-motion: reduce){.list-item{animation:none}}
</style>
