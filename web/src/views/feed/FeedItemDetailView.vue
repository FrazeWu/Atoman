<template>
  <div class="a-page" style="padding-bottom:12rem">
    <!-- Loading -->
    <div v-if="loading" style="display:flex;flex-direction:column;gap:1.5rem;max-width:800px;margin:0 auto">
      <div class="a-skeleton" style="height:3rem" />
      <div class="a-skeleton" style="height:1.5rem;width:50%" />
      <div class="a-skeleton" style="height:20rem" />
    </div>

    <!-- Content -->
    <article v-else-if="item" style="max-width:800px;margin:0 auto">
      <!-- Header -->
      <header style="margin-bottom:2rem">
        <div style="display:flex;align-items:center;gap:.75rem;margin-bottom:1rem;flex-wrap:wrap">
          <span class="a-badge">RSS 订阅</span>
          <span class="a-label a-muted">{{ item.feed_source?.title || '未知来源' }}</span>
          <span v-if="item.author" class="a-label a-muted">· {{ item.author }}</span>
          <span style="font-size:.75rem;color:#9ca3af">{{ formatDate(item.published_at) }}</span>
        </div>
        
        <h1 style="font-size:2rem;font-weight:900;line-height:1.3;margin-bottom:1rem">
          {{ item.title }}
        </h1>

        <!-- Original blog link -->
        <a
          v-if="item.link"
          :href="item.link"
          target="_blank"
          rel="noopener noreferrer"
          class="a-btn-outline-sm"
          style="display:inline-flex;align-items:center;gap:.5rem;text-decoration:none"
        >
          📄 访问原博客
        </a>
      </header>

      <!-- Featured Image -->
      <figure v-if="item.image_url" style="margin:0 0 2rem">
        <img
          :src="item.image_url"
          :alt="item.title"
          style="width:100%;height:auto;border:2px solid #000;filter:grayscale(100%)"
        />
        <figcaption v-if="item.image_caption" class="a-muted" style="font-size:.75rem;margin-top:.5rem;text-align:center">
          {{ item.image_caption }}
        </figcaption>
      </figure>

      <!-- Content -->
      <div
        class="markdown-body"
        style="font-size:1rem;line-height:1.7"
        v-html="renderContent(item.content || item.summary || '')"
      ></div>

      <!-- Podcast Player -->
      <div
        v-if="item.enclosure_url && item.enclosure_type?.startsWith('audio/') || item.duration"
        style="margin-top:2rem;padding:1.5rem;border:2px solid #000;background:#fafafa"
      >
        <h3 style="font-size:1.125rem;font-weight:900;margin-bottom:1rem">音频内容</h3>
        <div style="display:flex;align-items:center;gap:1rem">
          <button
            @click="togglePlay"
            style="font-weight:900;font-size:.7rem;text-transform:uppercase;letter-spacing:.08em;padding:.75rem 1.5rem;border:2px solid #000;cursor:pointer;transition:all .2s"
            :style="isPlaying ? 'background:#000;color:#fff' : 'background:#fff;color:#000'"
          >
            {{ isPlaying ? '⏸ 暂停' : '▶ 播放' }}
          </button>
          <span v-if="item.duration" style="font-size:.875rem;font-weight:700">时长：{{ item.duration }}</span>
        </div>
        <audio
          v-if="item.enclosure_url"
          ref="audioRef"
          :src="item.enclosure_url"
          @ended="onEnded"
          style="width:100%;margin-top:1rem"
        />
      </div>

      <!-- Back button -->
      <div style="margin-top:3rem;padding-top:2rem;border-top:2px solid #000">
        <RouterLink to="/feed" class="a-btn-outline-sm">← 返回订阅</RouterLink>
      </div>
    </article>

    <!-- Empty/Error -->
    <AEmpty v-else text="内容不存在或已被删除" />
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted, onUnmounted } from 'vue'
import { useRoute, RouterLink } from 'vue-router'
import DOMPurify from 'dompurify'
import AEmpty from '@/components/ui/AEmpty.vue'

interface FeedItem {
  id: string
  title: string
  content?: string
  summary?: string
  link?: string
  author?: string
  published_at: string
  image_url?: string
  image_caption?: string
  enclosure_url?: string
  enclosure_type?: string
  duration?: string
  feed_source?: {
    id: string
    title: string
    rss_url: string
  }
}

const route = useRoute()

const loading = ref(true)
const item = ref<FeedItem | null>(null)
const audioRef = ref<HTMLAudioElement | null>(null)
const isPlaying = ref(false)

const formatDate = (dateStr: string) => {
  return new Date(dateStr).toLocaleDateString('zh-CN', {
    year: 'numeric',
    month: 'long',
    day: 'numeric',
    hour: '2-digit',
    minute: '2-digit'
  })
}

const renderContent = (html: string) => {
  // Sanitize HTML to prevent XSS, then add img styling
  const clean = DOMPurify.sanitize(html, {
    USE_PROFILES: { html: true },
    ADD_ATTR: ['target', 'rel'],
  })
  return clean.replace(/<img/g, '<img style="max-width:100%;height:auto"')
}

const togglePlay = () => {
  if (!audioRef.value || !item.value?.enclosure_url) return
  
  if (isPlaying.value) {
    audioRef.value.pause()
    isPlaying.value = false
  } else {
    audioRef.value.play()
    isPlaying.value = true
  }
}

const onEnded = () => {
  isPlaying.value = false
}

// Fetch feed item detail
const fetchItem = async () => {
  loading.value = true
  try {
    const API_URL = import.meta.env.VITE_API_URL || '/api'
    const res = await fetch(`${API_URL}/feed/items/${route.params.id}`, {
      headers: {
        Authorization: `Bearer ${localStorage.getItem('token')}`
      }
    })
    
    if (res.ok) {
      const data = await res.json()
      item.value = data.data
    }
  } catch (e) {
    console.error('Failed to fetch feed item:', e)
  } finally {
    loading.value = false
  }
}

onMounted(() => {
  fetchItem()
})

onUnmounted(() => {
  if (audioRef.value) {
    audioRef.value.pause()
    audioRef.value = null
  }
})
</script>

<style scoped>
.markdown-body :deep(p) {
  margin: 1rem 0;
}

.markdown-body :deep(h1),
.markdown-body :deep(h2),
.markdown-body :deep(h3) {
  margin-top: 2rem;
  margin-bottom: 1rem;
  font-weight: 900;
}

.markdown-body :deep(a) {
  color: inherit;
  text-decoration: underline;
}

.markdown-body :deep(blockquote) {
  border-left: 4px solid #000;
  padding-left: 1rem;
  margin: 1rem 0;
  color: #6b7280;
}

.markdown-body :deep(code) {
  background: #f3f4f6;
  padding: 0.2rem 0.4rem;
  border-radius: 0.25rem;
  font-family: monospace;
}

.markdown-body :deep(pre) {
  background: #1f2937;
  color: #fff;
  padding: 1rem;
  border-radius: 0.5rem;
  overflow-x: auto;
}
</style>
