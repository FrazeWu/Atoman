<script setup lang="ts">
import type { Video } from '@/types'

defineProps<{ video: Video }>()

function fmtDuration(sec: number): string {
  if (!sec) return ''
  const h = Math.floor(sec / 3600)
  const m = Math.floor((sec % 3600) / 60)
  const s = sec % 60
  if (h > 0) return `${h}:${m.toString().padStart(2, '0')}:${s.toString().padStart(2, '0')}`
  return `${m}:${s.toString().padStart(2, '0')}`
}

function fmtViews(n: number): string {
  if (n >= 10000) return `${(n / 10000).toFixed(1)}万`
  return n.toString()
}

function fmtDate(s: string): string {
  return new Date(s).toLocaleDateString('zh-CN')
}
</script>

<template>
  <RouterLink :to="`/video/${video.id}`" class="vc-card">
    <div class="vc-thumb">
      <img v-if="video.thumbnail_url" :src="video.thumbnail_url" :alt="video.title" class="vc-img" />
      <div v-else class="vc-thumb-placeholder">
        <svg width="32" height="32" viewBox="0 0 24 24" fill="currentColor" style="opacity:0.3">
          <path d="M8 5v14l11-7z"/>
        </svg>
      </div>
      <span v-if="video.duration_sec" class="vc-duration">{{ fmtDuration(video.duration_sec) }}</span>
    </div>
    <div class="vc-body">
      <p class="vc-title">{{ video.title }}</p>
      <p v-if="video.channel" class="vc-channel">{{ video.channel.name }}</p>
      <p class="vc-meta">{{ fmtViews(video.view_count) }} 次播放 · {{ fmtDate(video.created_at) }}</p>
    </div>
  </RouterLink>
</template>

<style scoped>
.vc-card {
  display: block;
  text-decoration: none;
  color: inherit;
  border-radius: 0.5rem;
  overflow: hidden;
  transition: opacity 0.15s;
}
.vc-card:hover { opacity: 0.85; }
.vc-card:hover .vc-title { text-decoration: underline; }

.vc-thumb {
  position: relative;
  aspect-ratio: 16/9;
  background: var(--a-color-surface, #f3f4f6);
  border-radius: 0.375rem;
  overflow: hidden;
}
.vc-img {
  width: 100%;
  height: 100%;
  object-fit: cover;
}
.vc-thumb-placeholder {
  position: absolute;
  inset: 0;
  display: flex;
  align-items: center;
  justify-content: center;
  color: var(--a-color-muted, #9ca3af);
}
.vc-duration {
  position: absolute;
  bottom: 0.3rem;
  right: 0.4rem;
  background: rgba(0, 0, 0, 0.8);
  color: #fff;
  font-size: 0.7rem;
  font-weight: 700;
  padding: 0.1rem 0.35rem;
  border-radius: 0.2rem;
  letter-spacing: 0.02em;
}

.vc-body {
  padding: 0.5rem 0.25rem;
}
.vc-title {
  font-size: 0.875rem;
  font-weight: 600;
  line-height: 1.4;
  display: -webkit-box;
  -webkit-line-clamp: 2;
  -webkit-box-orient: vertical;
  overflow: hidden;
  color: var(--a-color-fg);
  margin: 0 0 0.2rem 0;
}
.vc-channel {
  font-size: 0.75rem;
  color: var(--a-color-muted, #6b7280);
  margin: 0 0 0.1rem 0;
}
.vc-meta {
  font-size: 0.7rem;
  color: var(--a-color-muted-soft, #9ca3af);
  margin: 0;
}
</style>
