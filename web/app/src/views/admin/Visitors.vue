<template>
  <PageShell title="访客管理" width="wide">
    <div class="panel">
      <n-data-table :columns="cols" :data="visitors" :bordered="false" size="small" :row-key="r => r.uuid" />
    </div>

    <!-- 编辑弹窗 -->
    <n-modal :show="editing" @update:show="editing = $event">
      <div class="edit-card">
        <h4>编辑访客</h4>
        <img :src="editForm.avatar" class="edit-avatar" />
        <n-input v-model:value="editForm.nickname" placeholder="昵称" style="margin-bottom:8px" />
        <n-input v-model:value="editForm.signature" placeholder="签名" style="margin-bottom:12px" />
        <n-button type="primary" block @click="saveEdit">保存</n-button>
      </div>
    </n-modal>
  </PageShell>
</template>

<script setup>
import { ref, onMounted, h } from 'vue'
import { NButton, NPopconfirm } from 'naive-ui'
import { listVisitors, updateVisitor, removeVisitor } from '../../api/visitors'
import PageShell from '../../components/admin/PageShell.vue'
import { entityAvatar } from '../../utils/avatar'

const visitors = ref([])
const editing = ref(false)
const editForm = ref({ uuid: '', nickname: '', signature: '', avatar: '' })

const cols = [
  {
    title: '访客', key: 'nickname', width: 200,
    render(row) {
      return h('div', { style: 'display:flex;align-items:center;gap:8px' }, [
        h('img', { src: entityAvatar(row), style: 'width:28px;height:28px;border-radius:50%;background:var(--tag-bg)' }),
        h('span', { style: 'font-weight:600' }, row.nickname),
      ])
    },
  },
  { title: '签名', key: 'signature', width: '*', ellipsis: { tooltip: true }, render(row) { return row.signature || '—' } },
  { title: '注册时间', key: 'created_at', width: 110, render(row) { return row.created_at?.slice(0, 10) } },
  {
    title: '', width: 90,
    render(row) {
      const btns = [h(NButton, { size: 'tiny', onClick: () => editVisitor(row) }, { default: () => '编辑' })]
      if (!row.uuid.startsWith('admin_')) {
        btns.push(h(NPopconfirm, { onPositiveClick: () => del(row.uuid) }, {
          trigger: () => h(NButton, { size: 'tiny', text: true, type: 'error' }, { default: () => '删除' }),
          default: () => '确定删除？评论和弹幕会保留。',
        }))
      }
      return h('div', { style: 'display:flex;gap:2px' }, btns)
    },
  },
]

onMounted(async () => {
  try {
    visitors.value = (await listVisitors()).filter(v => !v.uuid.startsWith('admin_'))
  } catch {}
})

function editVisitor(v) {
  editForm.value = {
    uuid: v.uuid,
    nickname: v.nickname,
    signature: v.signature,
    avatar: entityAvatar(v),
  }
  editing.value = true
}

async function saveEdit() {
  try {
    await updateVisitor(editForm.value.uuid, {
      nickname: editForm.value.nickname,
      signature: editForm.value.signature,
    })
    const v = visitors.value.find(v => v.uuid === editForm.value.uuid)
    if (v) {
      v.nickname = editForm.value.nickname
      v.signature = editForm.value.signature
    }
    editing.value = false
  } catch {}
}

async function del(uuid) {
  try {
    await removeVisitor(uuid)
    visitors.value = visitors.value.filter(v => v.uuid !== uuid)
  } catch {}
}
</script>

<style scoped>
.edit-card { background: var(--card); border: 1px solid var(--card-border); border-radius: 12px; padding: 24px; max-width: 300px; margin: 0 auto; text-align: center; }
.edit-card h4 { font-size: 16px; margin: 0 0 16px; color: var(--text); }
.edit-avatar { width: 56px; height: 56px; border-radius: 50%; margin-bottom: 14px; border: 2px solid var(--gold); }
</style>
