<template>
  <div class="max-w-[640px] mx-auto">
    <router-link to="/notes" class="text-xs text-ink3 no-underline transition-colors duration-700 hover:text-accent-strong">← 返回随笔</router-link>
    <div v-if="loading" class="text-center py-16 text-sm text-ink3">加载中...</div>
    <div v-else-if="note" class="mt-6">
      <div class="flex items-center gap-2.5 mb-5">
        <img v-if="authorAvatarUrl" :src="authorAvatarUrl" class="w-[42px] h-[42px] rounded-full object-cover border border-line" alt="博主头像" />
        <div v-else class="w-[42px] h-[42px] rounded-full bg-paper2 border border-line flex items-center justify-center text-ink3">
          <UserIcon class="w-[18px] h-[18px]" />
        </div>
        <div>
          <div class="text-sm tracking-widest text-ink">{{ authorName }}</div>
          <div v-if="authorSignature" class="text-[10px] text-ink3 mt-0.5 leading-snug">{{ authorSignature }}</div>
          <time class="text-[11px] text-ink3 mt-0.5 block">{{ rel(note.created_at) }}</time>
        </div>
      </div>
      <div class="text-[15px] md:text-base leading-[2] text-ink" v-html="note.html" />
      <div v-if="imgs.length" class="grid gap-1 mt-4" :class="'images--'+imgs.length"><img v-for="(u,i) in imgs" :key="i" :src="u" loading="lazy" alt="" class="w-full h-full object-cover" @error="e=>e.target.style.display='none'" /></div>
    </div>
    <div v-else class="text-center py-20 text-ink3">
      <TrashIcon class="w-10 h-10 mx-auto" />
      <p class="my-3">找不到这条随笔</p>
      <router-link to="/notes" class="text-[13px] text-accent-strong no-underline transition-colors duration-700 hover:text-ink">← 返回</router-link>
    </div>
    <CommentSection v-if="note" :note-id="note.id" />
  </div>
</template>
<script setup>
import { ref, computed, onMounted } from 'vue'
import { useRoute } from 'vue-router'
import { PersonOutline, TrashOutline } from '@vicons/ionicons5'
import CommentSection from '../components/CommentSection.vue'
import { getNote } from '../api/notes'
import { useAuthor } from '../composables/useAuthor'
import { rel } from '../utils/format'
const UserIcon=PersonOutline, TrashIcon=TrashOutline
const route=useRoute(), note=ref(null), loading=ref(true)
const { author } = useAuthor()
const authorName = computed(() => author.value.name), authorAvatarUrl = computed(() => author.value.avatar), authorSignature = computed(() => author.value.signature)
const imgs=computed(()=>note.value?.images?note.value.images.split(',').filter(Boolean):[])
onMounted(async()=>{
  try{note.value=await getNote(route.params.id)}catch(e){}
  loading.value=false
})
</script>
<style scoped>
/* 九宫格图片：1–9 张不同排布（类名由 JS 依据图片数拼接，勿改名） */
.images { border: 1px solid var(--line2); }
.images--1 { grid-template-columns: 1fr; max-width: 320px; }
.images--2 { grid-template-columns: 1fr 1fr; max-width: 400px; }
.images--3 { grid-template-columns: 1fr 1fr; max-width: 400px; }
.images--3 img:first-child { grid-column: 1 / -1; }
.images--4 { grid-template-columns: 1fr 1fr; max-width: 400px; }
.images--5, .images--6 { grid-template-columns: repeat(3, 1fr); max-width: 420px; }
.images--7, .images--8, .images--9 { grid-template-columns: repeat(3, 1fr); max-width: 420px; }
</style>
