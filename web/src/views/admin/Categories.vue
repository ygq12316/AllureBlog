<template>
  <PageShell title="分类管理">
    <div class="panel">
      <div class="flex gap-2 mb-4">
        <InkInput v-model="n" placeholder="分类名称" @keydown.enter="add" class="flex-1" />
        <InkButton variant="primary" size="sm" @click="add">添加</InkButton>
      </div>
      <InkTable :columns="cols" :data="list">
        <template #actions="{ row }">
          <InkPopconfirm text="确定?" @confirm="del(row.id)">
            <template #trigger><InkButton variant="danger" size="xs">删除</InkButton></template>
          </InkPopconfirm>
        </template>
      </InkTable>
    </div>
  </PageShell>
</template>
<script setup>
import { ref, onMounted } from 'vue'
import { listCategories, createCategory, removeCategory } from '../../api/categories'
import PageShell from '../../components/admin/PageShell.vue'
const list=ref([]),n=ref('')
const cols=[{title:'名称',key:'name'},{title:'文章数',key:'article_count',width:80}]
onMounted(async()=>{list.value=await listCategories()})
async function add(){if(!n.value.trim())return;await createCategory(n.value);n.value='';list.value=await listCategories()}
async function del(id){await removeCategory(id);list.value=list.value.filter(c=>c.id!==id)}
</script>
