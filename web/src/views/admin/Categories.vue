<template>
  <PageShell title="分类管理">
    <div class="panel">
      <div class="add-row"><n-input v-model:value="n" placeholder="分类名称" @keyup.enter="add" /><n-button type="primary" @click="add">添加</n-button></div>
      <n-data-table :columns="cols" :data="list" :bordered="false" size="small" />
    </div>
  </PageShell>
</template>
<script setup>
import { ref, onMounted, h } from 'vue';import { NButton, NPopconfirm } from 'naive-ui';import { listCategories, createCategory, removeCategory } from '../../api/categories'
import PageShell from '../../components/admin/PageShell.vue'
const list=ref([]),n=ref('')
const cols=[{title:'名称',key:'name'},{title:'文章数',key:'article_count',width:80},{title:'',key:'actions',width:70,render(row){return h(NPopconfirm,{onPositiveClick:()=>del(row.id)},{trigger:()=>h(NButton,{size:'tiny',text:true,type:'error'},{default:()=>'删除'}),default:()=>'确定?'})}}]
onMounted(async()=>{list.value=await listCategories()})
async function add(){if(!n.value.trim())return;await createCategory(n.value);n.value='';list.value=await listCategories()}
async function del(id){await removeCategory(id);list.value=list.value.filter(c=>c.id!==id)}
</script>
<style scoped>
.add-row{display:flex;gap:8px;margin-bottom:16px}
</style>
