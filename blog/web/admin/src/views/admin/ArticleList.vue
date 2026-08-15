<template>
  <div class="wrap">
    <div class="bar">
      <h2 class="bar-title">文章管理</h2>
      <div class="bar-actions">
        <n-button v-if="selected.length" type="error" size="tiny" @click="batchDel">删除选中 ({{ selected.length }})</n-button>
        <n-button type="primary" @click="$router.push('/admin/articles/new')">+ 写文章</n-button>
      </div>
    </div>
    <n-data-table :columns="cols" :data="articles" :bordered="false" size="small"
      :row-key="r=>r.id" @update:checked-row-keys="selected=$event" />
  </div>
</template>
<script setup>
import { ref, onMounted, h } from 'vue';import { useRouter } from 'vue-router';import { NButton, NTag, NCheckbox } from 'naive-ui';import axios from 'axios'
const router=useRouter(),articles=ref([]),selected=ref([])
const cols=[
  {type:'selection',width:40},
  {title:'标题',key:'title',width:'*',render(row){return h('a',{style:'color:var(--gold);cursor:pointer',onClick:()=>router.push(`/admin/articles/${row.id}/edit`)},row.title)}},
  {title:'分类',key:'category',width:80,render(row){return row.category?h(NTag,{size:'tiny',bordered:false},{default:()=>row.category}):''}},
  {title:'状态',width:70,render(row){return h(NTag,{size:'tiny',type:row.is_published?'success':'warning',bordered:false},{default:()=>row.is_published?'已发布':'草稿'})}},
  {title:'日期',width:110,render(row){return new Date(row.created_at).toLocaleDateString('zh-CN')}},
  {title:'',width:50,render(row){return h(NButton,{size:'tiny',onClick:()=>router.push(`/admin/articles/${row.id}/edit`)},{default:()=>'编辑'})}},
]
onMounted(async()=>{const{data}=await axios.get('/api/articles?all=true');articles.value=data.articles||[]})
async function batchDel(){for(const id of selected.value)await axios.delete(`/api/articles/${id}`);articles.value=articles.value.filter(a=>!selected.value.includes(a.id));selected.value=[]}
</script>
<style scoped>
.wrap{max-width:800px;margin:0 auto}
.bar{display:flex;justify-content:space-between;align-items:center;margin-bottom:16px;flex-wrap:wrap;gap:8px}
.bar-title{font-size:18px;font-weight:700;color:var(--text);margin:0}
.bar-actions{display:flex;gap:8px}
</style>