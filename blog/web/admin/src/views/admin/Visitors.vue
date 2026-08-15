<template>
  <div class="wrap">
    <h3 class="title">访客管理 ({{ visitors.length }})</h3>
    <div v-if="!visitors.length" class="empty">暂无访客</div>
    <div v-for="v in visitors" :key="v.uuid" class="row">
      <img :src="`https://api.dicebear.com/9.x/${v.avatar_style}/svg?seed=${v.uuid}`" class="row-avatar" />
      <div class="row-body">
        <div class="row-name">{{ v.nickname }}</div>
        <div class="row-sig">{{ v.signature || '无签名' }}</div>
        <div class="row-time">{{ v.created_at?.slice(0,10) }}</div>
      </div>
      <div class="row-actions">
        <n-button size="tiny" text @click="editVisitor(v)">编辑</n-button>
        <n-popconfirm v-if="!v.uuid.startsWith('admin_')" @positive-click="del(v.uuid)">
          <template #trigger><n-button size="tiny" text type="error">删除</n-button></template>
          确定删除此访客？其评论和弹幕会保留。
        </n-popconfirm>
      </div>
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
  </div>
</template>

<script setup>
import { ref, onMounted } from 'vue'
import axios from 'axios'

const visitors = ref([])
const editing = ref(false)
const editForm = ref({ uuid: '', nickname: '', signature: '', avatar: '' })

onMounted(async () => {
  try {
    const { data } = await axios.get('/api/admin/visitors')
    visitors.value = (data.visitors || []).filter(v => !v.uuid.startsWith('admin_'))
  } catch {}
})

function editVisitor(v) {
  editForm.value = {
    uuid: v.uuid,
    nickname: v.nickname,
    signature: v.signature,
    avatar: `https://api.dicebear.com/9.x/${v.avatar_style}/svg?seed=${v.uuid}`,
  }
  editing.value = true
}

async function saveEdit() {
  try {
    await axios.put(`/api/admin/visitors/${editForm.value.uuid}`, {
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
    await axios.delete(`/api/admin/visitors/${uuid}`)
    visitors.value = visitors.value.filter(v => v.uuid !== uuid)
  } catch {}
}
</script>

<style scoped>
.wrap { max-width: 640px; margin: 0 auto; }
.title { font-size: 17px; font-weight: 700; color: var(--text); margin: 0 0 20px; }
.empty { text-align: center; padding: 40px; color: var(--muted); font-size: 13px; }
.row { display: flex; align-items: center; gap: 12px; padding: 12px 0; border-bottom: 1px solid var(--card-border); }
.row-avatar { width: 36px; height: 36px; border-radius: 50%; background: var(--tag-bg); flex-shrink: 0; }
.row-body { flex: 1; min-width: 0; }
.row-name { font-size: 14px; font-weight: 600; color: var(--text); }
.row-sig { font-size: 11px; color: var(--muted); margin-top: 2px; }
.row-time { font-size: 10px; color: var(--muted); }
.row-actions { display: flex; gap: 4px; flex-shrink: 0; }

.edit-card { background: var(--card); border: 1px solid var(--card-border); border-radius: 12px; padding: 24px; max-width: 300px; margin: 0 auto; text-align: center; }
.edit-card h4 { font-size: 16px; margin: 0 0 16px; color: var(--text); }
.edit-avatar { width: 56px; height: 56px; border-radius: 50%; margin-bottom: 14px; border: 2px solid var(--gold); }
</style>
