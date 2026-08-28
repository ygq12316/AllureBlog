<template>
  <div class="page">
    <router-link to="/notes" class="back">← 返回随笔</router-link>
    <div v-if="loading" class="loading">加载中...</div>
    <div v-else-if="note" class="detail">
      <div class="header">
        <img v-if="authorAvatarUrl" :src="authorAvatarUrl" class="author-img" />
        <div v-else class="avatar"><n-icon size="18" :component="UserIcon" /></div>
        <div>
          <div class="name">{{ authorName }}</div>
          <div v-if="authorSignature" class="author-sig">{{ authorSignature }}</div>
          <time class="time">{{ rel(note.created_at) }}</time>
        </div>
      </div>
      <div class="body" v-html="note.html" />
      <div v-if="imgs.length" class="images" :class="'images--'+imgs.length"><img v-for="(u,i) in imgs" :key="i" :src="u" loading="lazy" @error="e=>e.target.style.display='none'" /></div>
    </div>
    <div v-else class="empty"><n-icon size="40" :component="TrashIcon" /><p>找不到这条随笔</p><router-link to="/notes">← 返回</router-link></div>
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
.page{padding:0;max-width:640px;margin:0 auto}.back{font-size:12px;color:var(--muted);text-decoration:none}.back:hover{color:var(--gold)}
.loading{text-align:center;padding:60px;color:var(--muted)}.detail{margin-top:24px}
.header{display:flex;align-items:center;gap:10px;margin-bottom:20px}
.avatar{width:42px;height:42px;border-radius:6px;background:var(--tag-bg);border:1px solid var(--card-border);display:flex;align-items:center;justify-content:center;color:var(--gold)}
.author-img{width:42px;height:42px;border-radius:50%;object-fit:cover;border:1px solid var(--card-border)}
.name{font-size:14px;font-weight:600;color:var(--text)}.author-sig{font-size:10px;color:var(--muted);margin-top:1px;line-height:1.3}.time{font-size:11px;color:var(--muted);margin-top:2px}
.body{font-size:clamp(14px,1.1vw,16px);line-height:2;color:var(--text)}
.images{display:grid;gap:4px;margin-top:16px;border-radius:4px;overflow:hidden}.images img{width:100%;height:100%;object-fit:cover}
.images--1{grid-template-columns:1fr;max-width:320px}.images--2{grid-template-columns:1fr 1fr;max-width:400px}.images--3{grid-template-columns:1fr 1fr;max-width:400px}.images--3 img:first-child{grid-column:1/-1}.images--4{grid-template-columns:1fr 1fr;max-width:400px}.images--5,.images--6{grid-template-columns:repeat(3,1fr);max-width:420px}.images--7,.images--8,.images--9{grid-template-columns:repeat(3,1fr);max-width:420px}
.empty{text-align:center;padding:80px 0;color:var(--muted)}.empty p{margin:12px 0}.empty a{color:var(--gold);text-decoration:none;font-size:13px}
</style>
