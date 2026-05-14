<script setup lang="ts">
import { ref, onMounted, computed } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { useAuthStore } from '@/stores/auth'
import type { PodcastEpisode, Channel } from '@/types'

const API_URL = import.meta.env.VITE_API_URL || '/api'
const route = useRoute()
const router = useRouter()
const authStore = useAuthStore()

const isEdit = computed(() => !!route.params.id)
const saving = ref(false)
const error = ref('')
const channels = ref<Channel[]>([])

const form = ref({
  channel_id: '',
  title: '',
  shownotes: '',
  audio_url: '',
  duration_sec: 0,
  episode_cover_url: '',
  season_number: 1,
  episode_number: 0,
  status: 'draft' as 'draft' | 'published',
})

async function loadChannels() {
  if (!authStore.user) return
  const res = await fetch(
    `${API_URL}/blog/channels?user_id=${authStore.user.id}`,
    { headers: { Authorization: `Bearer ${authStore.token}` } }
  )
  if (res.ok) {
    const data = await res.json()
    // API returns { data: Channel[] }
    channels.value = data.data ?? data
    if (!form.value.channel_id && channels.value.length > 0) {
      form.value.channel_id = channels.value[0].id
    }
  }
}

async function loadEpisode() {
  const id = route.params.id as string
  const res = await fetch(`${API_URL}/podcast/episodes/${id}`, {
    headers: { Authorization: `Bearer ${authStore.token}` },
  })
  if (!res.ok) return
  const ep: PodcastEpisode = await res.json()
  form.value = {
    channel_id: ep.channel_id,
    title: ep.post?.title || '',
    shownotes: ep.post?.content || '',
    audio_url: ep.audio_url,
    duration_sec: ep.duration_sec,
    episode_cover_url: ep.episode_cover_url,
    season_number: ep.season_number,
    episode_number: ep.episode_number,
    status: (ep.post?.status as 'draft' | 'published') || 'draft',
  }
}

onMounted(async () => {
  await loadChannels()
  if (isEdit.value) await loadEpisode()
})

async function save(publish = false) {
  saving.value = true
  error.value = ''
  const payload = { ...form.value, status: publish ? 'published' : form.value.status }
  try {
    if (isEdit.value) {
      const res = await fetch(`${API_URL}/podcast/episodes/${route.params.id}`, {
        method: 'PUT',
        headers: {
          'Content-Type': 'application/json',
          Authorization: `Bearer ${authStore.token}`,
        },
        body: JSON.stringify(payload),
      })
      if (!res.ok) throw await res.json()
      router.push(`/podcast/${route.params.id}`)
    } else {
      const res = await fetch(`${API_URL}/podcast/episodes`, {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
          Authorization: `Bearer ${authStore.token}`,
        },
        body: JSON.stringify(payload),
      })
      if (!res.ok) throw await res.json()
      const ep: PodcastEpisode = await res.json()
      router.push(`/podcast/${ep.id}`)
    }
  } catch (e: any) {
    error.value = e?.error || '保存失败，请重试'
  } finally {
    saving.value = false
  }
}
</script>

<template>
  <div class="pe-wrap">
    <h1 class="pe-title">{{ isEdit ? '编辑单集' : '发布新单集' }}</h1>

    <form class="pe-form" @submit.prevent="save()">
      <!-- 节目频道 -->
      <label class="pe-label">
        节目（频道）*
        <select v-model="form.channel_id" required class="pe-input">
          <option disabled value="">请选择频道</option>
          <option v-for="ch in channels" :key="ch.id" :value="ch.id">{{ ch.name }}</option>
        </select>
      </label>

      <!-- 标题 -->
      <label class="pe-label">
        单集标题 *
        <input v-model="form.title" required class="pe-input" placeholder="单集标题" />
      </label>

      <!-- 音频 URL -->
      <label class="pe-label">
        音频文件 URL *
        <input v-model="form.audio_url" required class="pe-input" placeholder="https://..." />
        <span class="pe-hint">请先通过上传接口获取音频 URL，再填入此处</span>
      </label>

      <!-- 预览播放器（仅填写了 URL 时显示） -->
      <div v-if="form.audio_url" class="pe-preview">
        <span class="pe-label-text">预览</span>
        <audio :src="form.audio_url" controls class="pe-audio-preview" preload="none" />
      </div>

      <!-- 时长 -->
      <label class="pe-label">
        时长（秒）
        <input v-model.number="form.duration_sec" type="number" min="0" class="pe-input" />
      </label>

      <!-- 单集封面 -->
      <label class="pe-label">
        单集封面 URL（可选，不填则使用节目封面）
        <input v-model="form.episode_cover_url" class="pe-input" placeholder="https://..." />
      </label>

      <!-- 季 / 集编号 -->
      <div class="pe-row">
        <label class="pe-label pe-row-item">
          季
          <input v-model.number="form.season_number" type="number" min="1" class="pe-input" />
        </label>
        <label class="pe-label pe-row-item">
          集
          <input v-model.number="form.episode_number" type="number" min="0" class="pe-input" />
        </label>
      </div>

      <!-- Shownotes -->
      <label class="pe-label">
        节目说明（Shownotes）
        <textarea
          v-model="form.shownotes"
          class="pe-textarea"
          rows="6"
          placeholder="节目内容简介、时间轴、相关链接等"
        />
      </label>

      <p v-if="error" class="pe-error">{{ error }}</p>

      <div class="pe-actions">
        <button type="submit" :disabled="saving" class="pe-btn pe-btn--ghost">
          {{ saving ? '保存中…' : '保存草稿' }}
        </button>
        <button type="button" :disabled="saving" class="pe-btn pe-btn--primary" @click="save(true)">
          发布
        </button>
      </div>
    </form>
  </div>
</template>

<style scoped>
.pe-wrap { max-width: 40rem; margin: 0 auto; padding: 2rem 1rem; }
.pe-title { font-size: 1.5rem; font-weight: 700; margin-bottom: 1.5rem; }
.pe-form { display: flex; flex-direction: column; gap: 1rem; }
.pe-label { display: flex; flex-direction: column; gap: 0.25rem; font-size: 0.875rem; font-weight: 500; }
.pe-label-text { font-size: 0.875rem; font-weight: 500; margin-bottom: 0.25rem; }
.pe-input, .pe-textarea {
  padding: 0.5rem 0.75rem;
  border: 1px solid #d1d5db;
  border-radius: 0.375rem;
  font-size: 0.875rem;
  background: white;
  outline: none;
  transition: border-color 0.15s;
}
.pe-input:focus, .pe-textarea:focus { border-color: #374151; }
.pe-textarea { resize: vertical; }
.pe-hint { font-size: 0.75rem; color: #9ca3af; }
.pe-preview { display: flex; flex-direction: column; gap: 0.25rem; }
.pe-audio-preview { width: 100%; }
.pe-row { display: flex; gap: 1rem; }
.pe-row-item { flex: 1; }
.pe-error { font-size: 0.875rem; color: #ef4444; }
.pe-actions { display: flex; gap: 0.75rem; padding-top: 0.5rem; }
.pe-btn {
  padding: 0.5rem 1.25rem;
  border-radius: 0.375rem;
  font-size: 0.875rem;
  font-weight: 500;
  cursor: pointer;
  transition: all 0.15s;
}
.pe-btn:disabled { opacity: 0.5; cursor: not-allowed; }
.pe-btn--primary { background: #111827; color: white; border: 1px solid #111827; }
.pe-btn--primary:hover:not(:disabled) { background: #374151; }
.pe-btn--ghost { background: transparent; border: 1px solid #d1d5db; color: #374151; }
.pe-btn--ghost:hover:not(:disabled) { border-color: #374151; }
</style>
