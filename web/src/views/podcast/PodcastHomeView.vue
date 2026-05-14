<script setup lang="ts">
import { ref, onMounted } from 'vue'
import type { PodcastEpisode } from '@/types'

const API_URL = import.meta.env.VITE_API_URL || '/api'
const episodes = ref<PodcastEpisode[]>([])
const loading = ref(false)

onMounted(async () => {
  loading.value = true
  try {
    const res = await fetch(`${API_URL}/podcast/episodes`)
    if (res.ok) episodes.value = await res.json()
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
  <div class="ph-wrap">
    <h1 class="ph-title">播客</h1>

    <div v-if="loading" class="ph-state">加载中…</div>
    <div v-else-if="episodes.length === 0" class="ph-state">暂无节目</div>

    <ul v-else class="ph-list">
      <li v-for="ep in episodes" :key="ep.id" class="ph-item">
        <img
          :src="ep.episode_cover_url || ep.channel?.cover_url || ''"
          class="ph-cover"
          :alt="ep.post?.title"
        />
        <div class="ph-body">
          <RouterLink :to="`/podcast/${ep.id}`" class="ph-ep-title">
            {{ ep.post?.title }}
          </RouterLink>
          <RouterLink
            v-if="ep.channel"
            :to="`/podcast/show/${ep.channel.slug}`"
            class="ph-show-name"
          >{{ ep.channel.name }}</RouterLink>
          <span v-if="ep.duration_sec" class="ph-duration">{{ fmtDuration(ep.duration_sec) }}</span>
        </div>
      </li>
    </ul>
  </div>
</template>

<style scoped>
.ph-wrap { max-width: 48rem; margin: 0 auto; padding: 2rem 1rem; }
.ph-title { font-size: 1.5rem; font-weight: 700; margin-bottom: 1.5rem; }
.ph-state { text-align: center; padding: 4rem 0; color: #9ca3af; }
.ph-list { display: flex; flex-direction: column; gap: 0.75rem; list-style: none; padding: 0; }
.ph-item { display: flex; gap: 0.75rem; align-items: flex-start; border-bottom: 1px solid #f3f4f6; padding-bottom: 0.75rem; }
.ph-cover { width: 4rem; height: 4rem; border-radius: 0.25rem; object-fit: cover; flex-shrink: 0; }
.ph-body { display: flex; flex-direction: column; gap: 0.125rem; }
.ph-ep-title { font-size: 0.875rem; font-weight: 500; text-decoration: none; color: inherit; }
.ph-ep-title:hover { text-decoration: underline; }
.ph-show-name { font-size: 0.75rem; color: #6b7280; text-decoration: none; }
.ph-show-name:hover { text-decoration: underline; }
.ph-duration { font-size: 0.75rem; color: #9ca3af; }
</style>
