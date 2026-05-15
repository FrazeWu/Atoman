<script setup lang="ts">
import { ref, onMounted, computed } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { useAuthStore } from '@/stores/auth'
import APageHeader from '@/components/ui/APageHeader.vue'
import ABtn from '@/components/ui/ABtn.vue'
import AInput from '@/components/ui/AInput.vue'
import ATextarea from '@/components/ui/ATextarea.vue'
import ASelect from '@/components/ui/ASelect.vue'
import type { Video, Channel } from '@/types'

const API_URL = import.meta.env.VITE_API_URL || '/api'
const route = useRoute()
const router = useRouter()
const authStore = useAuthStore()

const isEdit = computed(() => !!route.params.id)
const saving = ref(false)
const errorMsg = ref('')
const channels = ref<Channel[]>([])

const form = ref({
  channel_id: '' as string,
  title: '',
  description: '',
  storage_type: 'external' as 'local' | 'external',
  video_url: '',
  thumbnail_url: '',
  duration_sec: 0,
  visibility: 'public' as 'public' | 'followers' | 'private',
  status: 'draft' as 'draft' | 'published',
  tags: '',
})

const channelOptions = computed(() => [
  { label: '不关联频道', value: '' },
  ...channels.value.map(ch => ({ label: ch.name, value: ch.id })),
])

const storageOptions = [
  { label: '外部链接（YouTube / Bilibili / 其他）', value: 'external' },
  { label: '本地上传（S3/MinIO）', value: 'local' },
]

const visibilityOptions = [
  { label: '公开', value: 'public' },
  { label: '仅关注者', value: 'followers' },
  { label: '私密', value: 'private' },
]

async function loadChannels() {
  if (!authStore.user) return
  const res = await fetch(
    `${API_URL}/blog/channels?user_id=${authStore.user.id}`,
    { headers: { Authorization: `Bearer ${authStore.token}` } }
  )
  if (res.ok) {
    const data = await res.json()
    channels.value = data.data ?? data
    if (!form.value.channel_id && channels.value.length > 0) {
      form.value.channel_id = channels.value[0].id
    }
  }
}

async function loadVideo() {
  const id = route.params.id as string
  const res = await fetch(`${API_URL}/videos/${id}`, {
    headers: { Authorization: `Bearer ${authStore.token}` },
  })
  if (!res.ok) return
  const v: Video = await res.json()
  form.value = {
    channel_id: v.channel_id ?? '',
    title: v.title,
    description: v.description,
    storage_type: v.storage_type,
    video_url: v.video_url,
    thumbnail_url: v.thumbnail_url,
    duration_sec: v.duration_sec,
    visibility: v.visibility,
    status: v.status,
    tags: v.tags?.map(t => t.name).join(', ') ?? '',
  }
}

onMounted(async () => {
  await loadChannels()
  if (isEdit.value) await loadVideo()
})

async function save(publish = false) {
  if (!form.value.title.trim() || !form.value.video_url.trim()) return
  saving.value = true
  errorMsg.value = ''
  const payload = {
    ...form.value,
    channel_id: form.value.channel_id || null,
    duration_sec: Number(form.value.duration_sec) || 0,
    status: publish ? 'published' : form.value.status,
    tags: form.value.tags.split(',').map(t => t.trim()).filter(Boolean),
  }
  try {
    if (isEdit.value) {
      const res = await fetch(`${API_URL}/videos/${route.params.id}`, {
        method: 'PUT',
        headers: { 'Content-Type': 'application/json', Authorization: `Bearer ${authStore.token}` },
        body: JSON.stringify(payload),
      })
      if (!res.ok) throw await res.json()
      router.push(`/video/${route.params.id}`)
    } else {
      const res = await fetch(`${API_URL}/videos`, {
        method: 'POST',
        headers: { 'Content-Type': 'application/json', Authorization: `Bearer ${authStore.token}` },
        body: JSON.stringify(payload),
      })
      if (!res.ok) throw await res.json()
      const v: Video = await res.json()
      router.push(`/video/${v.id}`)
    }
  } catch (e: any) {
    errorMsg.value = e?.error || '保存失败，请重试'
  } finally {
    saving.value = false
  }
}
</script>

<template>
  <div class="ve-wrap">
    <APageHeader :title="isEdit ? '编辑视频' : '上传视频'" accent style="margin-bottom:2rem" />

    <form class="ve-form" @submit.prevent="save()">
      <!-- 频道 -->
      <ASelect
        label="频道"
        :model-value="form.channel_id"
        :options="channelOptions"
        @update:model-value="form.channel_id = $event as string"
      />

      <!-- 来源类型 -->
      <ASelect
        label="视频来源"
        :model-value="form.storage_type"
        :options="storageOptions"
        @update:model-value="form.storage_type = $event as 'local' | 'external'"
      />

      <!-- 标题 -->
      <AInput
        v-model="form.title"
        label="标题 *"
        placeholder="视频标题"
      />

      <!-- 视频链接 -->
      <AInput
        v-model="form.video_url"
        label="视频链接 *"
        :placeholder="form.storage_type === 'external' ? 'https://youtube.com/watch?v=...' : 'S3 对象路径'"
        hint="外部链接支持 YouTube / Bilibili 自动嵌入"
      />

      <!-- 封面图 -->
      <AInput
        v-model="form.thumbnail_url"
        label="封面图 URL"
        placeholder="https://..."
      />

      <!-- 时长（秒） -->
      <AInput
        v-model="form.duration_sec"
        label="时长（秒）"
        type="number"
        :min="0"
      />

      <!-- 简介 -->
      <ATextarea
        v-model="form.description"
        label="视频简介"
        placeholder="视频内容简介"
        :rows="4"
      />

      <!-- 标签 -->
      <AInput
        v-model="form.tags"
        label="标签"
        placeholder="music, tutorial, vlog（逗号分隔）"
      />

      <!-- 可见范围 -->
      <ASelect
        label="可见范围"
        :model-value="form.visibility"
        :options="visibilityOptions"
        @update:model-value="form.visibility = $event as 'public' | 'followers' | 'private'"
      />

      <!-- 错误提示 -->
      <p v-if="errorMsg" class="ve-error">{{ errorMsg }}</p>

      <!-- 操作按钮 -->
      <div class="ve-actions">
        <ABtn type="submit" variant="secondary" :disabled="saving">
          {{ saving ? '保存中…' : '保存草稿' }}
        </ABtn>
        <ABtn type="button" variant="primary" :disabled="saving" @click="save(true)">
          {{ saving ? '发布中…' : '发布' }}
        </ABtn>
      </div>
    </form>
  </div>
</template>

<style scoped>
.ve-wrap {
  max-width: 40rem;
  margin: 0 auto;
  padding: 2rem 1.5rem 6rem;
}
.ve-form {
  display: flex;
  flex-direction: column;
  gap: 1.25rem;
}
.ve-error {
  font-size: 0.875rem;
  color: var(--a-color-danger, #ef4444);
}
.ve-actions {
  display: flex;
  gap: 0.75rem;
  padding-top: 0.5rem;
}
</style>
