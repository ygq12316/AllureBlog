<template>
  <div class="max-w-[1100px] mx-auto">
    <!-- 顶栏：与文章编辑器同构 -->
    <div class="editor-topbar">
      <InkButton variant="link" size="sm" @click="$router.push('/admin/notes')">
        <span class="inline-flex items-center gap-1"><ArrowBackIcon class="w-3.5 h-3.5" /> 返回</span>
      </InkButton>
      <span class="text-[13px] font-mono transition-colors duration-700" :class="content.length > 400 ? 'text-cinnabar' : 'text-ink3'">{{ content.length }}/500</span>
      <div class="editor-topbar-side">
        <InkPopconfirm v-if="isEdit" text="确定删除这条随笔？其下评论会一并失去归属。此操作不可恢复。" @confirm="del">
          <template #trigger><InkButton variant="danger" size="sm">删除</InkButton></template>
        </InkPopconfirm>
        <InkButton variant="primary" @click="publish" :disabled="!canPublish" :loading="saving">发布</InkButton>
      </div>
    </div>

    <div class="flex flex-col lg:flex-row gap-5 items-stretch">
      <!-- 左栏：编辑卡（内容形态保留） -->
      <div class="flex-1 border border-line p-5 bg-card flex flex-col">
        <div class="flex items-center gap-2.5 mb-3.5">
          <div class="w-[34px] h-[34px] rounded-full bg-paper2 border border-line flex items-center justify-center shrink-0 text-ink3">
            <PersonIcon class="w-4 h-4" />
          </div>
          <span class="text-sm tracking-widest text-ink">{{ authorName }}</span>
        </div>
        <textarea v-model="content" placeholder="写点什么..." maxlength="500" rows="8" autofocus
          class="w-full border-0 bg-transparent resize-none outline-none text-[15px] leading-[1.8] text-ink flex-1 min-h-[200px] caret-[var(--accent)] placeholder:text-ink3/45" />
        <div v-if="images.length" class="flex gap-2 flex-wrap mt-3.5">
          <div v-for="(img, i) in images" :key="i" class="relative w-[72px] h-[72px] overflow-hidden border border-line">
            <img :src="img" class="w-full h-full object-cover" alt="" />
            <button @click="remove(i)" aria-label="移除图片"
              class="absolute top-0.5 right-0.5 w-4 h-4 flex items-center justify-center bg-ink/70 text-paper border-0 cursor-pointer text-[10px] leading-none p-0 transition-colors duration-700 hover:bg-cinnabar">✕</button>
          </div>
        </div>
        <div class="mt-3.5 pt-3 border-t border-line2">
          <InkFilePicker v-if="images.length < 9" accept="image/*" @file="upload">
            <InkButton variant="link" size="sm">+ 添加图片</InkButton>
          </InkFilePicker>
        </div>
      </div>

      <!-- 右栏：评论面板（仅编辑态可用） -->
      <div class="panel lg:w-[380px] lg:shrink-0 flex flex-col max-h-[640px]">
        <template v-if="isEdit">
          <div class="flex justify-between items-baseline pb-3 border-b border-line2 shrink-0">
            <span class="text-sm tracking-widest text-ink">评论 · {{ comments.length }}</span>
            <span class="text-[11px] text-ink3">全部来自这条随笔</span>
          </div>
          <div class="flex-1 overflow-y-auto">
            <div v-if="!comments.length" class="py-10 text-center text-[13px] text-ink3">还没有评论</div>
            <div v-for="c in comments" :key="c.id" class="flex items-start gap-2.5 py-3 border-b border-line2 last:border-b-0">
              <img :src="avt(c)" class="w-7 h-7 rounded-full shrink-0 bg-paper2" alt="" />
              <div class="flex-1 min-w-0">
                <div class="flex items-baseline gap-2">
                  <span class="text-[13px] text-accent-strong">{{ c.nickname || '匿名' }}</span>
                  <span class="text-[10px] text-ink3">{{ c.created_at?.slice(0, 16) }}</span>
                </div>
                <p class="text-[13px] text-ink mt-1 mb-0 break-words">{{ c.content }}</p>
              </div>
              <InkPopconfirm text="确定删除这条评论？" @confirm="delComment(c.id)">
                <template #trigger><InkButton variant="danger" size="xs">删除</InkButton></template>
              </InkPopconfirm>
            </div>
          </div>
        </template>
        <div v-else class="flex-1 flex items-center justify-center py-10 text-center text-[13px] text-ink3 leading-loose">发布随笔后<br />即可在此管理评论</div>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, computed, onMounted } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { PersonOutline, ArrowBackOutline } from '@vicons/ionicons5'
import InkButton from '../../components/ui/InkButton.vue'
import InkPopconfirm from '../../components/ui/InkPopconfirm.vue'
import InkFilePicker from '../../components/ui/InkFilePicker.vue'
import { toast } from '../../composables/useToast'
import { uploadFile } from '../../api/upload'
import { listComments, removeComment } from '../../api/comments'
import { getNote, createNote, updateNote, removeNote } from '../../api/notes'
import { useAuthor } from '../../composables/useAuthor'
import { entityAvatar } from '../../utils/avatar'

const PersonIcon = PersonOutline
const ArrowBackIcon = ArrowBackOutline
const route = useRoute()
const router = useRouter()
const content = ref('')
const images = ref([])
const { author } = useAuthor()
const authorName = computed(() => author.value.name)
const comments = ref([])
const saving = ref(false)

const isEdit = computed(() => !!route.params.id)
const canPublish = computed(() => content.value.trim().length > 0 || images.value.length > 0)

function remove(i) { images.value.splice(i, 1) }

function avt(c) { return entityAvatar(c, c.visitor_uuid) }

async function upload(file) {
  try {
    const { url } = await uploadFile(file)
    images.value.push(url)
  } catch (e) {
    toast.error('图片上传失败')
  }
}

async function loadComments(id) {
  try {
    const data = await listComments(id)
    comments.value = data.comments || []
  } catch (e) {
    toast.error('评论加载失败')
  }
}

async function publish() {
  saving.value = true
  try {
    const id = route.params.id
    const payload = { content: content.value, images: images.value.join(','), is_published: true }
    if (id) await updateNote(id, payload)
    else await createNote(payload)
    router.push('/admin/notes')
  } catch (e) {
    toast.error(e.response?.data?.error || '发布失败')
    saving.value = false
  }
}

async function del() {
  const id = route.params.id
  if (!id) return
  try {
    await removeNote(id)
    router.push('/admin/notes')
  } catch (e) {
    toast.error('删除失败')
  }
}

async function delComment(cid) {
  try {
    await removeComment(cid)
    comments.value = comments.value.filter(c => c.id !== cid)
  } catch (e) {
    toast.error('评论删除失败')
  }
}

onMounted(async () => {
  const id = route.params.id
  if (!id) return
  try {
    const data = await getNote(id)
    content.value = data.content || ''
    images.value = data.images ? data.images.split(',').filter(Boolean) : []
  } catch (e) {
    toast.error('随笔加载失败')
  }
  loadComments(id)
})
</script>
