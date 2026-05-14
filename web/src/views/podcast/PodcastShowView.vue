<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { useRoute } from 'vue-router'
import type { PodcastEpisode, Channel } from '@/types'

const API_URL = import.meta.env.VITE_API_URL || '/api'
const route = useRoute()
const channel = ref<Channel | null>(null)
const episodes = ref<PodcastEpisode[]>([])
const loading = ref(true)

onMounted(async () => {
  const slug = route.params.channelSlug as string
  try {
    const res = await fetch(`${API_URL}/podcast/shows/${slug}/episodes`)
    if (res.ok) {
      const data = await res.json()
      channel.value = data.channel
      episodes.value = data.episodes
    }
  } finally {
    loading.value = false
  }
})
</script>

<template>
  <div v-if="loading" class="ps-state">加载中…</div>
  <div v-else-if="!channel" class="ps-state">节目不存在</div>
  <div v-else class="ps-wrap">
    <header class="ps-header">
      <img :src="channel.cover_url || ''" class="ps-cover" :alt="channel.name" />
      <div>
        <h1 class="ps-name">{{ channel.name }}</h1>
        <p v-if="channel.description" class="ps-desc">{{ channel.description }}</p>
        <a :href="`${API_URL}/channels/${channel.slug}/rss/podcast`" class="ps-rss" target="_blank">RSS 订阅</a>
      </div>
    </header>

    <div v-if="episodes.length === 0" class="ps-empty">暂无单集</div>
    <ul v-else class="ps-list">
      <li v-for="ep in episodes" :key="ep.id" class="ps-ep">
        <span v-if="ep.episode_number" class="ps-ep-num">第 {{ ep.episode_number }} 集</span>
        <RouterLink :to="`/podcast/${ep.id}`" class="ps-ep-title">
          {{ ep.post?.title }}
        </RouterLink>
      </li>
    </ul>
  </div>
</template>

<style scoped>
.ps-state { text-align: center; padding: 6rem 0; color: #9ca3af; }
.ps-wrap { max-width: 48rem; margin: 0 auto; padding: 2rem 1rem; }
.ps-header { display: flex; gap: 1.5rem; margin-bottom: 2rem; }
.ps-cover { width: 8rem; height: 8rem; border-radius: 0.25rem; object-fit: cover; flex-shrink: 0; }
.ps-name { font-size: 1.5rem; font-weight: 700; }
.ps-desc { font-size: 0.875rem; color: #6b7280; margin-top: 0.25rem; }
.ps-rss { font-size: 0.75rem; color: #9ca3af; margin-top: 0.5rem; display: inline-block; }
.ps-rss:hover { text-decoration: underline; }
.ps-empty { color: #9ca3af; padding: 2rem 0; }
.ps-list { list-style: none; padding: 0; display: flex; flex-direction: column; gap: 0.5rem; }
.ps-ep { display: flex; align-items: baseline; gap: 0.75rem; padding: 0.5rem 0; border-bottom: 1px solid #f3f4f6; }
.ps-ep-num { font-size: 0.75rem; color: #9ca3af; width: 3rem; flex-shrink: 0; }
.ps-ep-title { font-size: 0.875rem; text-decoration: none; color: inherit; }
.ps-ep-title:hover { text-decoration: underline; }
</style>
