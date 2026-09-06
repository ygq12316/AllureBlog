<template>
  <PageShell title="标签管理">
    <div class="panel">
      <div class="flex gap-2 mb-4">
        <InkInput v-model="n" placeholder="标签名称" @keydown.enter="add" class="flex-1" />
        <InkButton variant="primary" size="sm" @click="add">添加</InkButton>
      </div>
      <div class="flex flex-wrap gap-2">
        <InkTag v-for="t in list" :key="t.id" closable @close="del(t.id)">{{ t.name }} ({{ t.article_count }})</InkTag>
      </div>
    </div>
  </PageShell>
</template>
<script setup>
import { ref, onMounted } from 'vue'
import { listTags, createTag, removeTag } from '../../api/tags'
import PageShell from '../../components/admin/PageShell.vue'
const list=ref([]),n=ref('')
onMounted(async()=>{list.value=await listTags()})
async function add(){if(!n.value.trim())return;await createTag(n.value);n.value='';list.value=await listTags()}
async function del(id){await removeTag(id);list.value=list.value.filter(t=>t.id!==id)}
</script>
