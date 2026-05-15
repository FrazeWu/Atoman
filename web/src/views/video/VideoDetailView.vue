<script setup lang="ts">
import { ref, onMounted, computed, watch } from 'vue'
import { useRoute } from 'vue-router'
import type { Video } from '@/types'
import VideoCard from '@/components/shared/VideoCard.vue'

const API_URL = import.meta.env.VITE_API_URL || '/api'
const route = useRoute()

const video = ref<Video | null>(null)
const recommended = ref<Video[]>([])
const loading = ref(true)
const error = ref('')

// Detect embed type from URL
const embedSrc = computed(() => {
  const url = video.value?.video_url || ''
  const ytMatch = url.match(/(?:youtube\.com\/watch\?v=|youtu\.be\/)([A-Za-z0-9_-]{11})/)
  if (ytMatch) return `https://www.youtube.com/embed/${ytMatch[1]}?rel=0`
  const biliMatch = url.match(/bilibili\.com\/video\/(BV[A-Za-z0-9]+)/)
  if (biliMatch) return `https://player.bilibili.com/player.html?bvid=${biliMatch[1]}&autoplay=0`
  return null
})

function fmtDuration(sec: number): string {
  if (!sec) return ''
  const h = Math.floor(sec / 3600)
  const m = Math.floor((sec % 3600) / 60)
  const s = sec % 60
  if (h > 0) return `${h}:${m.toString().padStart(2, '0')}:${s.toString().padStart(2, '0')}`
  return `${m}:${s.toString().padStart(2, '0')}`
}

function fmtDate(s: string): string {
  return new Date(s).toLocaleDateString('zh-CN')
}

async function load(id: string) {
  loading.value = true
  error.value = ''
  video.value = null
  recommended.value = []
  try {
    const [vRes, rRes] = await Promise.all([
      fetch(`${API_URL}/videos/${id}`),
      fetch(`${API_URL}/videos/${id}/recommended`),
    ])
    if (!vRes.ok) { error.value = '视频不存在'; return }
    video.value = await vRes.json()
    if (rRes.ok) recommended.value = await rRes.json()
    // Fire-and-forget view count increment
    fetch(`${API_URL}/videos/${id}/view`, { method: 'POST' })
  } catch {
    error.value = '加载失败，请重试'
  } finally {
    loading.value = false
  }
}

onMounted(() => load(route.params.id as string))
watch(() => route.params.id, (id) => { if (id) load(id as string) })
</script>

<template>
  <div class="vd-page">
    <div v-if="loading" class="vd-loading">
      <div class="vd-loading-player" />
      <div class="vd-loading-info">
        <div class="vd-loading-line vd-loading-line--lg" />
        <div class="vd-loading-line vd-loading-line--sm" />
      </div>
    </div>

    <div v-else-if="error" class="vd-error">{{ error }}</div>

    <div v-else-if="video" class="vd-layout">
      <!-- Main: player + info -->
      <div class="vd-main">
        <!-- Player -->
        <div class="vd-player-wrap">
          <iframe
            v-if="embedSrc"
            :src="embedSrc"
            class="vd-embed"
            allow="accelerometer; autoplay; clipboard-write; encrypted-media; gyroscope; picture-in-picture; web-share"
            allowfullscreen
          />
          <video
            v-else-if="video.storage_type === 'local'"
            :src="video.video_url"
            :poster="video.thumbnail_url || undefined"
            controls
            class="vd-native"
            preload="metadata"
          />
          <div v-else class="vd-external">
            <a :href="video.video_url" target="_blank" rel="noopener noreferrer" class="vd-external-link">
              在外部平台观看 →
            </a>
          </div>
        </div>

        <!-- Title -->
        <h1 class="vd-title">{{ video.title }}</h1>

        <!-- Meta row -->
        <div class="vd-meta-row">
          <RouterLink v-if="video.channel" :to="`/channel/${video.channel.slug || video.channel_id}`" class="vd-channel">
            {{ video.channel.name }}
          </RouterLink>
          <div class="vd-stats">
            <span>{{ video.view_count.toLocaleString() }} 次播放</span>
            <span>{{ fmtDate(video.created_at) }}</span>
            <span v-if="video.duration_sec">{{ fmtDuration(video.duration_sec) }}</span>
          </div>
        </div>

        <!-- Tags -->
        <div v-if="video.tags && video.tags.length" class="vd-tags">
          <span v-for="tag in video.tags" :key="tag.id" class="vd-tag"># {{ tag.name }}</span>
        </div>

        <!-- Description -->
        <div v-if="video.description" class="vd-desc">
          <pre class="vd-desc-text">{{ video.description }}</pre>
        </div>
      </div>

      <!-- Sidebar: recommended -->
      <aside class="vd-sidebar">
        <h2 class="vd-sidebar-title">相关推荐</h2>
        <div v-if="recommended.length" class="vd-recommended">
          <VideoCard v-for="v in recommended" :key="v.id" :video="v" />
        </div>
        <p v-else class="vd-no-rec">暂无推荐视频</p>
      </aside>
    </div>
  </div>
</template>

<style scoped>
.vd-page {
  max-width: 80rem;
  margin: 0 auto;
  padding: 1.5rem 1.5rem 6rem;
}

/* Loading skeleton */
.vd-loading { display: flex; flex-direction: column; gap: 1rem; }
.vd-loading-player {
  width: 100%;
  aspect-ratio: 16/9;
  background: var(--a-color-surface, #f3f4f6);
  border-radius: 0.5rem;
  animation: pulse 1.5s ease-in-out infinite;
}
.vd-loading-info { display: flex; flex-direction: column; gap: 0.5rem; }
.vd-loading-line {
  height: 1rem;
  background: var(--a-color-surface, #f3f4f6);
  border-radius: 0.25rem;
  animation: pulse 1.5s ease-in-out infinite;
}
.vd-loading-line--lg { width: 60%; }
.vd-loading-line--sm { width: 30%; }

.vd-error {
  text-align: center;
  padding: 6rem 0;
  color: var(--a-color-danger, #ef4444);
}

/* Main layout */
.vd-layout {
  display: flex;
  gap: 1.5rem;
  align-items: flex-start;
}

.vd-main { flex: 1; min-width: 0; }

/* Player */
.vd-player-wrap {
  width: 100%;
  background: #000;
  border-radius: 0.5rem;
  overflow: hidden;
  margin-bottom: 1rem;
}
.vd-embed {
  display: block;
  width: 100%;
  aspect-ratio: 16/9;
  border: none;
}
.vd-native {
  display: block;
  width: 100%;
  aspect-ratio: 16/9;
  background: #000;
}
.vd-external {
  aspect-ratio: 16/9;
  display: flex;
  align-items: center;
  justify-content: center;
  background: var(--a-color-surface, #111);
}
.vd-external-link {
  color: #fff;
  font-size: 1rem;
  font-weight: 600;
  text-decoration: none;
}
.vd-external-link:hover { text-decoration: underline; }

/* Info */
.vd-title {
  font-size: 1.25rem;
  font-weight: 700;
  line-height: 1.4;
  color: var(--a-color-fg);
  margin: 0 0 0.75rem 0;
}
.vd-meta-row {
  display: flex;
  align-items: center;
  gap: 1rem;
  flex-wrap: wrap;
  margin-bottom: 0.75rem;
}
.vd-channel {
  font-weight: 700;
  font-size: 0.9rem;
  color: var(--a-color-fg);
  text-decoration: none;
}
.vd-channel:hover { text-decoration: underline; }
.vd-stats {
  display: flex;
  gap: 0.75rem;
  font-size: 0.8rem;
  color: var(--a-color-muted, #6b7280);
}
.vd-tags {
  display: flex;
  flex-wrap: wrap;
  gap: 0.4rem;
  margin-bottom: 0.75rem;
}
.vd-tag {
  font-size: 0.75rem;
  padding: 0.2rem 0.5rem;
  background: var(--a-color-surface, #f3f4f6);
  border-radius: 9999px;
  color: var(--a-color-muted, #6b7280);
  cursor: pointer;
}
.vd-tag:hover { color: var(--a-color-fg); }

.vd-desc {
  margin-top: 0.75rem;
  padding: 0.75rem 1rem;
  background: var(--a-color-surface, #f9fafb);
  border-radius: 0.5rem;
}
.vd-desc-text {
  font-size: 0.875rem;
  color: var(--a-color-fg);
  white-space: pre-wrap;
  word-break: break-word;
  font-family: inherit;
  margin: 0;
  line-height: 1.6;
}

/* Sidebar */
.vd-sidebar {
  width: 22rem;
  flex-shrink: 0;
}
.vd-sidebar-title {
  font-size: 0.875rem;
  font-weight: 700;
  text-transform: uppercase;
  letter-spacing: 0.06em;
  color: var(--a-color-muted, #6b7280);
  margin: 0 0 0.75rem 0;
}
.vd-recommended {
  display: flex;
  flex-direction: column;
  gap: 0.75rem;
}
.vd-no-rec {
  font-size: 0.8rem;
  color: var(--a-color-muted, #9ca3af);
  text-align: center;
  padding: 2rem 0;
}

/* Responsive */
@media (max-width: 1024px) {
  .vd-sidebar { width: 18rem; }
}
@media (max-width: 768px) {
  .vd-layout { flex-direction: column; }
  .vd-sidebar { width: 100%; }
  .vd-recommended {
    display: grid;
    grid-template-columns: repeat(2, 1fr);
    gap: 0.75rem;
  }
}

@keyframes pulse {
  0%, 100% { opacity: 1; }
  50% { opacity: 0.5; }
}
</style>
