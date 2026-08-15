<template>
  <div class="wrap">
    <div class="bar">
      <h2 class="bar-title">随笔管理</h2>
      <div class="bar-actions">
        <n-button v-if="selected.length" type="error" size="tiny" @click="batchDel">删除选中 ({{ selected.length }})</n-button>
        <n-button type="primary" @click="$router.push('/admin/notes/new')">+ 写随笔</n-button>
      </div>
    </div>
    <div v-if="!notes.length" class="empty">还没有随笔</div>
    <div v-for="n in notes" :key="n.id" class="item">
      <n-checkbox :checked="selected.includes(n.id)" @update:checked="toggle(n.id)" style="flex-shrink:0;margin-top:8px" />
      <div class="item-main">
        <div class="item-text" v-html="n.html" />
        <div v-if="imgs(n.images).length" class="item-imgs"><img v-for="(u,i) in imgs(n.images)" :key="i" :src="u" /></div>
        <div class="item-foot"><span>{{ rel(n.created_at) }}</span><n-button size="tiny" text @click="$router.push(`/admin/notes/${n.id}/edit`)">编辑</n-button><n-popconfirm @positive-click="del(n.id)"><template #trigger><n-button size="tiny" text type="error">删除</n-button></template>确定删除?</n-popconfirm></div>
      </div>
    </div>
  </div>
</template>
<script setup>
import { ref, onMounted } from 'vue';import { NCheckbox } from 'naive-ui';import axios from 'axios'
const notes=ref([]),selected=ref([])
onMounted(async()=>{const{data}=await axios.get('/api/notes?all=true');notes.value=data.notes||[]})
function toggle(id){const i=selected.value.indexOf(id);if(i>=0)selected.value.splice(i,1);else selected.value.push(id)}
async function del(id){await axios.delete(`/api/notes/${id}`);notes.value=notes.value.filter(n=>n.id!==id)}
async function batchDel(){for(const id of selected.value)await axios.delete(`/api/notes/${id}`);notes.value=notes.value.filter(n=>!selected.value.includes(n.id));selected.value=[]}
function imgs(s){return s?s.split(',').filter(Boolean):[]}
function rel(d){const s=Math.floor((Date.now()-new Date(d).getTime())/1000);if(s<60)return'刚刚';if(s<3600)return Math.floor(s/60)+'分钟前';if(s<86400)return Math.floor(s/3600)+'小时前';if(s<2592000)return Math.floor(s/86400)+'天前';return new Date(d).toLocaleDateString('zh-CN')}
</script>
<style scoped>
.wrap{max-width:600px;margin:0 auto}
.bar{display:flex;justify-content:space-between;align-items:center;margin-bottom:16px;gap:8px}.bar-title{font-size:18px;font-weight:700;color:var(--text);margin:0}
.bar-actions{display:flex;gap:8px}
.empty{text-align:center;padding:40px;color:var(--muted);font-size:13px}
.item{display:flex;gap:8px;padding:14px 0;border-bottom:1px solid var(--card-border)}
.item-main{flex:1;min-width:0}.item-text{font-size:13px;line-height:1.6;color:var(--text);margin-bottom:6px}
.item-imgs{display:flex;gap:4px;margin-bottom:6px}.item-imgs img{width:60px;height:60px;object-fit:cover;border-radius:3px}
.item-foot{display:flex;align-items:center;gap:10px;font-size:11px;color:var(--muted)}
</style>