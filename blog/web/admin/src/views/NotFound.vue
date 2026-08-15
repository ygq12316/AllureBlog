<template>
  <div class="box">
    <div class="icon"><n-icon size="24" :component="TrashIcon" /></div>
    <div class="title">这张纸是空的</div>
    <div class="desc">翻过头了，这篇手帖还没写下来</div>
    <div class="btns">
      <router-link to="/" class="btn-outline"><n-icon size="14" :component="BackIcon" /> 返回首页</router-link>
      <router-link to="/search" class="btn-solid"><n-icon size="14" :component="SearchIcon" /> 去搜索</router-link>
    </div>
  </div>
</template>
<script setup>
import { TrashOutline, ChevronBackOutline, SearchOutline } from '@vicons/ionicons5'
const TrashIcon=TrashOutline, BackIcon=ChevronBackOutline, SearchIcon=SearchOutline
</script>
<style scoped>
.box{text-align:center;padding:clamp(60px,15vh,120px) 20px;min-height:60vh;display:flex;flex-direction:column;align-items:center;justify-content:center}
.icon{width:56px;height:56px;border-radius:50%;background:var(--tag-bg);border:1px solid var(--border);display:flex;align-items:center;justify-content:center;color:var(--gold);margin-bottom:16px}
.title{font-size:clamp(16px,1.5vw,20px);font-weight:700;color:var(--text);margin-bottom:8px}
.desc{font-size:clamp(12px,.95vw,14px);color:var(--muted);margin-bottom:24px}
.btns{display:flex;gap:10px}
.btn-outline{padding:6px 18px;font-size:12px;background:var(--tag-bg);color:var(--text2);border:1px solid var(--border);border-radius:2px;text-decoration:none;display:flex;align-items:center;gap:4px}
.btn-solid{padding:6px 18px;font-size:12px;background:var(--gold);color:#fff;border-radius:2px;text-decoration:none;display:flex;align-items:center;gap:4px}
</style>
