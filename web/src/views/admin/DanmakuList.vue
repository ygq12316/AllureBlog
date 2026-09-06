<template>
  <PageShell title="弹幕管理">
    <div class="panel">
      <InkTable :columns="cols" :data="list" :row-key="r => r.id">
        <template #cell-color="{ row }">
          <span class="inline-block w-3 h-3 rounded-full align-middle" :style="{ background: row.color }" />
        </template>
        <template #actions="{ row }">
          <InkPopconfirm text="确定删除？" @confirm="del(row.id)">
            <template #trigger><InkButton variant="danger" size="xs">删除</InkButton></template>
          </InkPopconfirm>
        </template>
      </InkTable>
    </div>
  </PageShell>
</template>

<script setup>
import { ref, onMounted } from 'vue'
import { listDanmaku, removeDanmaku } from '../../api/danmaku'
import PageShell from '../../components/admin/PageShell.vue'
const list = ref([])

const cols = [
  { title: '昵称', key: 'nickname', width: 110, render(row) { return row.nickname || '匿名' } },
  { title: '内容', key: 'content', ellipsis: true },
  { title: '颜色', key: 'color', width: 60, slot: true },
  { title: '时间', key: 'created_at', width: 140, render(row) { return row.created_at?.slice(0, 16) } },
]

async function load() {
  try {
    list.value = await listDanmaku()
  } catch (e) {}
}

onMounted(load)

async function del(id) {
  await removeDanmaku(id)
  list.value = list.value.filter(d => d.id !== id)
}
</script>
