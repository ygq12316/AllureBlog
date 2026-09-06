<template>
  <div class="max-w-[600px] mx-auto text-center">
    <div class="mb-10">
      <SearchIcon class="w-10 h-10 text-accent mx-auto" />
      <h2 class="text-xl md:text-2xl font-light tracking-[0.2em] text-ink my-3 mb-6">搜索文章</h2>
      <form @submit.prevent="doSearch" class="w-full">
        <div class="flex items-center bg-paper border px-5 py-3 gap-2 transition-colors duration-700"
          :class="focused ? 'border-accent' : 'border-line'">
          <SearchIcon class="w-[18px] h-[18px] text-ink3 shrink-0" />
          <input v-model="q" placeholder="试试搜 Go、咖啡、徒步..." class="flex-1 border-0 bg-transparent font-serif text-base text-ink outline-none placeholder:text-ink3" @focus="focused=true" @blur="focused=false" />
          <button v-if="q" @click="q='';searched=false" type="button" class="bg-transparent border-0 p-1 cursor-pointer text-ink3 hover:text-ink transition-colors duration-700 flex items-center" aria-label="清空搜索">
            <CloseIcon class="w-3.5 h-3.5" />
          </button>
        </div>
      </form>
    </div>
    <div v-if="searched" class="text-left">
      <div v-if="results.length" class="text-xs text-ink3 mb-5 text-center flex items-center justify-center gap-1.5">
        <BookIcon class="w-3.5 h-3.5 text-accent" /> 找到 <span class="text-ink">{{ results.length }}</span> 篇关于 "<span class="text-accent-strong">{{ q }}</span>" 的文章
      </div>
      <div v-for="a in results" :key="a.id" class="py-4 border-b border-line2 group">
        <div class="flex items-center gap-1.5 mb-1.5">
          <TimeIcon class="w-3.5 h-3.5 text-accent" />
          <time class="text-[11px] text-accent-strong tracking-widest">{{ fmt(a.created_at) }}</time>
          <span v-if="a.category" class="text-[9px] px-1.5 py-px border border-line2 text-ink3">{{ a.category }}</span>
        </div>
        <router-link :to="'/posts/'+a.slug" class="block text-[15px] font-light tracking-wide text-ink no-underline mb-1 transition-colors duration-700 group-hover:text-accent-strong">{{ a.title }}</router-link>
        <p class="text-xs text-ink3 m-0 leading-relaxed">{{ a.excerpt }}</p>
      </div>
      <div v-if="!results.length" class="text-center py-10">
        <TrashIcon class="w-12 h-12 text-line mx-auto" />
        <div class="text-base font-light tracking-widest text-ink mt-3 mb-2">没有找到 "{{ q }}"</div>
        <div class="text-xs text-ink3 mb-5">换个关键词，或试试下面的标签</div>
        <WordCloud :tags="allTags" @select="q=$event;doSearch()" />
      </div>
    </div>
    <div v-else class="mt-3">
      <WordCloud :tags="allTags" @select="q=$event;doSearch()" />
    </div>
  </div>
</template>

<script setup>
import { ref, onMounted } from 'vue'
import { useRoute } from 'vue-router'
import { SearchOutline, CloseOutline, BookOutline, TimeOutline, TrashOutline } from '@vicons/ionicons5'
import WordCloud from '../components/WordCloud.vue'
import { listTags } from '../api/tags'
import { searchArticles } from '../api/articles'

const SearchIcon=SearchOutline, CloseIcon=CloseOutline, BookIcon=BookOutline, TimeIcon=TimeOutline, TrashIcon=TrashOutline
const route=useRoute(), q=ref(route.query.q||''), results=ref([]), searched=ref(false), focused=ref(false)
const allTags=ref([])

if(q.value){searched.value=true;doSearch()}

onMounted(async()=>{
  try{allTags.value=await listTags()}catch(e){}
})

async function doSearch(){if(!q.value.trim())return;searched.value=true;try{const data=await searchArticles(q.value);results.value=data.articles||[]}catch(e){results.value=[]}}
function fmt(d){return d?new Date(d).toLocaleDateString('zh-CN',{month:'short',day:'numeric'}):''}
</script>
