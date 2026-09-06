<template>
  <div class="text-center px-5 min-h-[60vh] flex flex-col items-center justify-center py-[10vh]">
    <div class="w-14 h-14 rounded-full bg-paper2 border border-line flex items-center justify-center text-ink3 mb-4">
      <TrashIcon class="w-6 h-6" />
    </div>
    <div class="text-lg font-light tracking-[0.2em] text-ink mb-2">这张纸是空的</div>
    <div class="text-[13px] text-ink3 mb-6">翻过头了，这篇手帖还没写下来</div>
    <div class="flex gap-2.5">
      <router-link to="/" class="px-4 py-1.5 text-xs no-underline border border-line text-ink2 flex items-center gap-1 transition-colors duration-700 hover:border-accent hover:text-accent-strong">
        <BackIcon class="w-3.5 h-3.5" /> 返回首页
      </router-link>
      <router-link to="/search" class="px-4 py-1.5 text-xs no-underline bg-ink text-paper flex items-center gap-1 transition-opacity duration-700">
        <SearchIcon class="w-3.5 h-3.5" /> 去搜索
      </router-link>
    </div>
  </div>
</template>
<script setup>
import { TrashOutline, ChevronBackOutline, SearchOutline } from '@vicons/ionicons5'
const TrashIcon=TrashOutline, BackIcon=ChevronBackOutline, SearchIcon=SearchOutline
</script>
