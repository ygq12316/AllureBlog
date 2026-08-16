<template>
  <div class="wrap page-narrow">
    <div class="page-head">
      <h2>标签管理</h2>
      <div class="page-head-actions" />
    </div>
    <div class="panel">
      <div class="add-row"><n-input v-model:value="n" placeholder="标签名称" @keyup.enter="add" /><n-button type="primary" @click="add">添加</n-button></div>
      <n-space><n-tag v-for="t in list" :key="t.id" closable @close="del(t.id)">{{ t.name }} ({{ t.article_count }})</n-tag></n-space>
    </div>
  </div>
</template>
<script setup>
import { ref, onMounted } from 'vue';import axios from 'axios'
const list=ref([]),n=ref('')
onMounted(async()=>{const{data}=await axios.get('/api/tags');list.value=data.tags||[]})
async function add(){if(!n.value.trim())return;await axios.post('/api/tags',{name:n.value});n.value='';const{data}=await axios.get('/api/tags');list.value=data.tags||[]}
async function del(id){await axios.delete(`/api/tags/${id}`);list.value=list.value.filter(t=>t.id!==id)}
</script>
<style scoped>
.add-row{display:flex;gap:8px;margin-bottom:16px}
</style>
