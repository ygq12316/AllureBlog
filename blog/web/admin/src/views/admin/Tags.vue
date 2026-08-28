<template>
  <PageShell title="标签管理">
    <div class="panel">
      <div class="add-row"><n-input v-model:value="n" placeholder="标签名称" @keyup.enter="add" /><n-button type="primary" @click="add">添加</n-button></div>
      <n-space><n-tag v-for="t in list" :key="t.id" closable @close="del(t.id)">{{ t.name }} ({{ t.article_count }})</n-tag></n-space>
    </div>
  </PageShell>
</template>
<script setup>
import { ref, onMounted } from 'vue';import { listTags, createTag, removeTag } from '../../api/tags'
import PageShell from '../../components/admin/PageShell.vue'
const list=ref([]),n=ref('')
onMounted(async()=>{list.value=await listTags()})
async function add(){if(!n.value.trim())return;await createTag(n.value);n.value='';list.value=await listTags()}
async function del(id){await removeTag(id);list.value=list.value.filter(t=>t.id!==id)}
</script>
<style scoped>
.add-row{display:flex;gap:8px;margin-bottom:16px}
</style>
