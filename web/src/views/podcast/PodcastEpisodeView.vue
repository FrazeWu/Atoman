<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { useRoute } from 'vue-router'
import type { PodcastEpisode } from '@/types'

const API_URL = import.meta.env.VITE_API_URL || '/api'
const route = useRoute()
const ep = ref<PodcastEpisode | null>(null)
const loading = ref(true)
const error = ref('')

onMounted(async () => {
  const id = route.params.id as string
  try {
    const res = await fetch(`${API_URL}/podcast/episodes/${id}`)
    if (res.ok) {
      ep.value = await res.json()
    } else {
      error.value = '单集不存在'
    }
  } catch {
    error.value = '加载失败，请重试'
  } finally {
    loading.value = false
  }
})

function fmtDuration(sec: number) {
  if (!sec) return ''
  const h = Math.floor(sec / 3600)
  const m = Math.floor((sec % 3600) / 60)
  const s = sec % 60
  if (h > 0) return `${h}:${m.toString().padStart(2, '0')}:${s.toString().padStart(2, '0')}`
  return `${m}:${s.toString().padStart(2, '0')}`
}
</script>

<template>
  <div v-if="loading" class="pev-state">加载中…</div>
  <div v-else-if="error" class="pev-state pev-error">{{ error }}</div>
  <div v-else-if="ep" class="pev-wrap">
    <!-- 封面 + 标题区 -->
    <div class="pev-header">
      <img
        :src="ep.episode_cover_url || ep.channel?.cover_url || ''"
        class="pev-cover"
        :alt="ep.post?.title"
      />
      <div class="pev-info">
        <h1 class="pev-title">{{ ep.post?.title }}</h1>
        <RouterLink
          v-if="ep.channel"
          :to="`/podcast/show/${ep.channel.slug}`"
          class="pev-show"
        >{{ ep.channel.name }}</RouterLink>
        <div class="pev-meta">
          <span v-if="ep.season_number > 1">第 {{ ep.season_number }} 季</span>
          <span v-if="ep.episode_number">第 {{ ep.episode_number }} 集</span>
          <span v-if="ep.duration_sec">{{ fmtDuration(ep.duration_sec) }}</span>
        </div>
      </div>
    </div>

    <!-- 音频播放器 -->
    <audio
      :src="ep.audio_url"
      controls
      class="pev-player"
      preload="metadata"
    />

    <!-- Shownotes / 节目说明 -->
    <div v-if="ep.post?.content" class="pev-notes">
      <h2 class="pev-notes-title">节目说明</h2>
      <div class="pev-notes-body">{{ ep.post.content }}</div>
    </div>
  </div>
</template>

<style scoped>
.pev-state { text-align: center; padding: 6rem 0; color: #9ca3af; }
.pev-error { color: #ef4444; }
.pev-wrap { max-width: 40rem; margin: 0 auto; padding: 2rem 1rem; }
.pev-header { display: flex; gap: 1.5rem; margin-bottom: 1.5rem; }
.pev-cover { width: 7rem; height: 7rem; border-radius: 0.375rem; object-fit: cover; flex-shrink: 0; }
.pev-info { display: flex; flex-direction: column; gap: 0.25rem; }
.pev-title { font-size: 1.25rem; font-weight: 700; line-height: 1.3; }
.pev-show { font-size: 0.875rem; color: #6b7280; text-decoration: none; }
.pev-show:hover { text-decoration: underline; }
.pev-meta { display: flex; gap: 0.75rem; font-size: 0.75rem; color: #9ca3af; margin-top: 0.25rem; }
.pev-player { width: 100%; margin-top: 1rem; }
.pev-notes { margin-top: 2rem; }
.pev-notes-title { font-size: 1rem; font-weight: 600; margin-bottom: 0.5rem; }
.pev-notes-body { font-size: 0.875rem; color: #4b5563; white-space: pre-wrap; line-height: 1.7; }
</style>
