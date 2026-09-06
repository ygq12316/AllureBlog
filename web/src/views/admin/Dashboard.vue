<template>
  <div class="max-w-[720px] mx-auto">
    <!-- 欢迎 -->
    <div class="flex items-center gap-3 mb-7">
      <img v-if="authorAvatar" :src="authorAvatar" class="w-10 h-10 rounded-full object-cover border border-accent" alt="博主头像" />
      <div v-else class="w-10 h-10 rounded-full bg-paper2 border border-line flex items-center justify-center text-ink3">
        <PersonIcon class="w-[18px] h-[18px]" />
      </div>
      <div>
        <div class="text-[17px] tracking-widest text-ink">欢迎回来，{{ authorName }}</div>
        <div v-if="authorSig" class="text-xs text-ink3">{{ authorSig }}</div>
      </div>
    </div>

    <!-- 统计 -->
    <div class="grid grid-cols-3 gap-3 mb-6">
      <div class="panel flex flex-col items-center gap-1 !py-4" v-for="s in statList" :key="s.label">
        <component :is="s.icon" class="w-[18px] h-[18px] text-accent" />
        <div class="text-2xl font-light text-ink">{{ s.value }}</div>
        <div class="text-[11px] text-ink3">{{ s.label }}</div>
      </div>
    </div>

    <!-- 快捷 -->
    <div class="flex gap-2.5 mb-7">
      <InkButton variant="primary" block @click="$router.push('/admin/articles/new')"><span class="inline-flex items-center gap-1.5"><CreateIcon class="w-4 h-4" /> 写文章</span></InkButton>
      <InkButton block @click="$router.push('/admin/notes/new')"><span class="inline-flex items-center gap-1.5"><ChatIcon class="w-4 h-4" /> 写随笔</span></InkButton>
    </div>

    <!-- 最近 -->
    <div class="grid md:grid-cols-2 gap-5">
      <div class="panel">
        <div class="text-[13px] tracking-widest text-ink mb-2.5">最近文章</div>
        <div v-if="!articles.length" class="text-xs text-ink3 py-4 text-center">暂无</div>
        <div v-for="a in articles" :key="a.id" @click="$router.push(`/admin/articles/${a.id}/edit`)"
          class="flex justify-between items-center py-2 border-b border-line2 last:border-b-0 cursor-pointer text-[13px] transition-colors duration-700 hover:text-accent-strong group">
          <span class="truncate flex-1 mr-2 text-ink group-hover:text-accent-strong">{{ a.title }}</span>
          <span class="text-[11px] text-ink3 whitespace-nowrap">{{ fmt(a.created_at) }}</span>
        </div>
      </div>
      <div class="panel">
        <div class="text-[13px] tracking-widest text-ink mb-2.5">最近随笔</div>
        <div v-if="!nlist.length" class="text-xs text-ink3 py-4 text-center">暂无</div>
        <div v-for="n in nlist" :key="n.id" @click="$router.push('/admin/notes')"
          class="flex justify-between items-center py-2 border-b border-line2 last:border-b-0 cursor-pointer text-[13px] transition-colors duration-700 group">
          <span class="truncate flex-1 mr-2 text-ink group-hover:text-accent-strong" v-html="trunc(n.html,24)" />
          <span class="text-[11px] text-ink3 whitespace-nowrap">{{ rel(n.created_at) }}</span>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, computed, onMounted } from 'vue'
import { PersonOutline, DocumentTextOutline, ChatbubbleOutline, ChatbubblesOutline, FolderOpenOutline, PricetagsOutline, CreateOutline, PeopleOutline } from '@vicons/ionicons5'
import InkButton from '../../components/ui/InkButton.vue'
import { getStats } from '../../api/stats'
import { listArticles } from '../../api/articles'
import { listNotes } from '../../api/notes'
import { listTags } from '../../api/tags'
import { useAuthor } from '../../composables/useAuthor'
import { rel, fmtDate, trunc } from '../../utils/format'

const PersonIcon=PersonOutline, DocIcon=DocumentTextOutline, ChatIcon=ChatbubbleOutline, BubblesIcon=ChatbubblesOutline, FolderIcon=FolderOpenOutline, PricetagIcon=PricetagsOutline, CreateIcon=CreateOutline, PeopleIcon=PeopleOutline

const stats=ref({article_count:0,note_count:0,category_count:0,tag_count:0,comment_count:0,visitor_count:0})
const articles=ref([]), nlist=ref([])
const { author } = useAuthor()
const authorName=computed(()=>author.value.name), authorAvatar=computed(()=>author.value.avatar), authorSig=computed(()=>author.value.signature)

const statList=computed(()=>[
  {icon:DocIcon, value:stats.value.article_count, label:'文章'},
  {icon:ChatIcon, value:stats.value.note_count, label:'随笔'},
  {icon:BubblesIcon, value:stats.value.comment_count ?? 0, label:'评论'},
  {icon:PeopleIcon, value:stats.value.visitor_count ?? 0, label:'用户'},
  {icon:FolderIcon, value:stats.value.category_count||0, label:'分类'},
  {icon:PricetagIcon, value:stats.value.tag_count||0, label:'标签'},
])

onMounted(async()=>{
  try{stats.value=await getStats()}catch(e){}
  try{articles.value=(await listArticles({all:'true',per_page:3})).articles||[]}catch(e){}
  try{nlist.value=(await listNotes({all:'true',per_page:3})).notes||[]}catch(e){}
  try{stats.value.tag_count=(await listTags()).length||0}catch(e){}
})

function fmt(d){return fmtDate(d)}
</script>
