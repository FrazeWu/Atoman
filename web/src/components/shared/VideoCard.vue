<script setup lang="ts">
import type { Video } from '@/types'

const props = defineProps<{ video: Video }>()

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
  return String(n)
}

function timeAgo(s: string): string {
  const diff = (Date.now() - new Date(s).getTime()) / 1000
  if (diff < 60) return '刚刚'
  if (diff < 3600) return `${Math.floor(diff / 60)} 分钟前`
  if (diff < 86400) return `${Math.floor(diff / 3600)} 小时前`
  if (diff < 2592000) return `${Math.floor(diff / 86400)} 天前`
  if (diff < 31536000) return `${Math.floor(diff / 2592000)} 个月前`
  return `${Math.floor(diff / 31536000)} 年前`
}

const avatarLetter = () =>
  (props.video.channel?.name ?? props.video.user?.username ?? '?')[0].toUpperCase()
</script>

<template>
  <RouterLink :to="`/video/${video.id}`" class="vc-card">
    <!-- Thumbnail -->
    <div class="vc-thumb">
      <img v-if="video.thumbnail_url" :src="video.thumbnail_url" :alt="video.title" class="vc-img" loading="lazy" />
      <div v-else class="vc-thumb-placeholder">
        <svg width="28" height="28" viewBox="0 0 24 24" fill="currentColor" style="opacity:0.25">
          <path d="M8 5v14l11-7z"/>
        </svg>
      </div>
      <span v-if="video.duration_sec" class="vc-duration">{{ fmtDuration(video.duration_sec) }}</span>
    </div>

    <!-- Info row: avatar + text -->
    <div class="vc-info">
      <div class="vc-avatar" aria-hidden="true">
        <img v-if="video.channel?.cover_url" :src="video.channel.cover_url" :alt="video.channel.name" />
        <span v-else>{{ avatarLetter() }}</span>
      </div>
      <div class="vc-text">
        <p class="vc-title">{{ video.title }}</p>
        <p v-if="video.channel" class="vc-channel">{{ video.channel.name }}</p>
        <p class="vc-meta">
          <span>{{ fmtViews(video.view_count) }} 次播放</span>
          <span class="vc-dot">·</span>
          <span>{{ timeAgo(video.created_at) }}</span>
        </p>
      </div>
    </div>
  </RouterLink>
</template>

<style scoped>
.vc-card {
  display: block;
  text-decoration: none;
  color: inherit;
}
.vc-card:hover .vc-title { text-decoration: underline; }

/* Thumbnail */
.vc-thumb {
  position: relative;
  aspect-ratio: 16/9;
  background: var(--a-color-surface);
  border-radius: 0.75rem;
  overflow: hidden;
}
.vc-thumb:hover .vc-img { transform: scale(1.02); }
.vc-img {
  width: 100%;
  height: 100%;
  object-fit: cover;
  transition: transform 0.2s ease;
  display: block;
}
.vc-thumb-placeholder {
  position: absolute;
  inset: 0;
  display: flex;
  align-items: center;
  justify-content: center;
  color: var(--a-color-muted);
}
.vc-duration {
  position: absolute;
  bottom: 0.35rem;
  right: 0.45rem;
  background: rgba(0, 0, 0, 0.82);
  color: #fff;
  font-size: 0.7rem;
  font-weight: 700;
  padding: 0.1rem 0.4rem;
  border-radius: 0.25rem;
  letter-spacing: 0.03em;
}

/* Info */
.vc-info {
  display: flex;
  gap: 0.65rem;
  padding: 0.6rem 0 0;
}
.vc-avatar {
  flex-shrink: 0;
  width: 2.25rem;
  height: 2.25rem;
  border-radius: 50%;
  overflow: hidden;
  background: var(--a-color-accent, #6366f1);
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 0.875rem;
  font-weight: 700;
  color: #fff;
}
.vc-avatar img {
  width: 100%;
  height: 100%;
  object-fit: cover;
}
.vc-text { min-width: 0; flex: 1; }
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
  font-size: 0.775rem;
  color: var(--a-color-muted);
  margin: 0 0 0.15rem 0;
}
.vc-meta {
  font-size: 0.75rem;
  color: var(--a-color-muted);
  margin: 0;
  display: flex;
  align-items: center;
  gap: 0.25rem;
}
.vc-dot { opacity: 0.5; }
</style>
