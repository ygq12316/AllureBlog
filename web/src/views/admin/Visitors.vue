<template>
  <PageShell title="用户管理" width="wide">
    <div class="panel">
      <InkTable :columns="cols" :data="visitors" :row-key="r => r.uuid">
        <template #cell-nickname="{ row }">
          <span class="flex items-center gap-2">
            <img :src="entityAvatar(row)" class="w-7 h-7 rounded-full bg-paper2" alt="" />
            <span class="tracking-wide text-ink">{{ row.nickname }}</span>
          </span>
        </template>
        <template #cell-role="{ row }">
          <InkTag :tone="row.role === 'admin' ? 'cinnabar' : 'sand'">{{ row.role === 'admin' ? '管理员' : '用户' }}</InkTag>
        </template>
        <template #actions="{ row }">
          <span class="flex gap-2">
            <InkButton variant="link" size="xs" @click="editVisitor(row)">编辑</InkButton>
            <InkPopconfirm v-if="!row.uuid.startsWith('admin_')" text="确定删除？评论和弹幕会保留。" @confirm="del(row.uuid)">
              <template #trigger><InkButton variant="danger" size="xs">删除</InkButton></template>
            </InkPopconfirm>
          </span>
        </template>
      </InkTable>
    </div>

    <!-- 编辑弹窗 -->
    <InkModal :show="editing" @update:show="editing = $event" title="编辑用户" width="320px">
      <div class="text-center">
        <img :src="editForm.avatar" class="w-14 h-14 rounded-full mx-auto mb-4 border border-accent" alt="用户头像" />
        <div class="flex flex-col gap-3 text-left">
          <InkInput v-model="editForm.nickname" placeholder="昵称" />
          <InkInput v-model="editForm.signature" placeholder="签名" />
          <InkButton variant="primary" block @click="saveEdit">保存</InkButton>
        </div>
      </div>
    </InkModal>
  </PageShell>
</template>

<script setup>
import { ref, onMounted } from 'vue'
import { listVisitors, updateVisitor, removeVisitor } from '../../api/visitors'
import PageShell from '../../components/admin/PageShell.vue'
import InkTable from '../../components/ui/InkTable.vue'
import InkModal from '../../components/ui/InkModal.vue'
import InkInput from '../../components/ui/InkInput.vue'
import InkButton from '../../components/ui/InkButton.vue'
import InkPopconfirm from '../../components/ui/InkPopconfirm.vue'
import InkTag from '../../components/ui/InkTag.vue'
import { entityAvatar } from '../../utils/avatar'

const visitors = ref([])
const editing = ref(false)
const editForm = ref({ uuid: '', nickname: '', signature: '', avatar: '' })

const cols = [
  { title: '用户', key: 'nickname', width: 180, slot: true },
  { title: '角色', key: 'role', width: 80, slot: true },
  { title: '签名', key: 'signature', ellipsis: true, render(row) { return row.signature || '—' } },
  { title: '注册时间', key: 'created_at', width: 110, render(row) { return row.created_at?.slice(0, 10) } },
]

onMounted(async () => {
  try {
    visitors.value = await listVisitors()
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
