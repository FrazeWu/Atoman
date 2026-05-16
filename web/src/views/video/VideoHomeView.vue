<script setup lang="ts">
import { ref, onMounted, watch } from 'vue'
import ABtn from '@/components/ui/ABtn.vue'
import { useAuthStore } from '@/stores/auth'
import type { Video } from '@/types'
import VideoCard from '@/components/shared/VideoCard.vue'

const API_URL = import.meta.env.VITE_API_URL || '/api'
const authStore = useAuthStore()
const videos = ref<Video[]>([])
const loading = ref(false)
const sort = ref<'latest' | 'popular'>('latest')

async function fetchVideos() {
  loading.value = true
  try {
    const res = await fetch(`${API_URL}/videos?sort=${sort.value}`)
    if (res.ok) videos.value = await res.json()
  } finally {
    loading.value = false
  }
}

onMounted(fetchVideos)
watch(sort, fetchVideos)
</script>

<template>
  <div class="vh-wrap">
    <!-- Sticky filter bar (YouTube style) -->
    <div class="vh-bar">
      <div class="vh-bar-inner">
        <button
          v-for="s in [{ v: 'latest', label: '最新上传' }, { v: 'popular', label: '最热播放' }]"
          :key="s.v"
          class="vh-chip"
          :class="{ 'vh-chip--active': sort === s.v }"
          @click="sort = s.v as 'latest' | 'popular'"
        >{{ s.label }}</button>
      </div>
      <div class="vh-bar-action">
        <ABtn v-if="authStore.isAuthenticated" to="/video/new" variant="primary" size="sm">+ 上传</ABtn>
      </div>
    </div>

    <!-- Skeleton -->
    <div v-if="loading" class="vh-grid">
      <div v-for="i in 12" :key="i" class="vh-skel">
        <div class="vh-skel-thumb" />
        <div class="vh-skel-info">
          <div class="vh-skel-avatar" />
          <div class="vh-skel-lines">
            <div class="vh-skel-line" style="width:85%" />
            <div class="vh-skel-line" style="width:55%" />
            <div class="vh-skel-line" style="width:40%" />
          </div>
        </div>
      </div>
    </div>

    <div v-else-if="videos.length === 0" class="vh-empty">暂无视频</div>

    <div v-else class="vh-grid">
      <VideoCard v-for="v in videos" :key="v.id" :video="v" />
    </div>
  </div>
</template>

<style scoped>
.vh-wrap {
  max-width: 82rem;
  margin: 0 auto;
  padding: 0 1rem 6rem;
}

/* Sticky filter bar */
.vh-bar {
  position: sticky;
  top: 0;
  z-index: 10;
  background: var(--a-color-bg);
  display: flex;
  align-items: center;
  gap: 0.75rem;
  padding: 0.75rem 0;
  border-bottom: 1px solid var(--a-color-border, #e5e7eb);
  margin-bottom: 1.5rem;
}
.vh-bar-inner {
  display: flex;
  gap: 0.4rem;
  overflow-x: auto;
  flex: 1;
  scrollbar-width: none;
}
.vh-bar-inner::-webkit-scrollbar { display: none; }

.vh-chip {
  flex-shrink: 0;
  padding: 0.3rem 0.75rem;
  font-size: 0.8rem;
  font-weight: 600;
  border: none;
  border-radius: 0.5rem;
  background: var(--a-color-surface);
  cursor: pointer;
  color: var(--a-color-fg);
  transition: background 0.12s;
  white-space: nowrap;
}
.vh-chip:hover { background: var(--a-color-border); }
.vh-chip--active {
  background: var(--a-color-fg);
  color: var(--a-color-bg);
}

/* Grid */
.vh-grid {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(17rem, 1fr));
  gap: 1rem 1rem;
}

/* Skeleton card */
.vh-skel { display: flex; flex-direction: column; gap: 0; }
.vh-skel-thumb {
  width: 100%;
  aspect-ratio: 16/9;
  background: var(--a-color-surface);
  border-radius: 0.75rem;
  animation: pulse 1.4s ease-in-out infinite;
}
.vh-skel-info {
  display: flex;
  gap: 0.65rem;
  padding: 0.6rem 0 0;
}
.vh-skel-avatar {
  flex-shrink: 0;
  width: 2.25rem;
  height: 2.25rem;
  border-radius: 50%;
  background: var(--a-color-surface);
  animation: pulse 1.4s ease-in-out infinite;
}
.vh-skel-lines {
  flex: 1;
  display: flex;
  flex-direction: column;
  gap: 0.35rem;
  padding-top: 0.1rem;
}
.vh-skel-line {
  height: 0.75rem;
  background: var(--a-color-surface);
  border-radius: 0.25rem;
  animation: pulse 1.4s ease-in-out infinite;
}

@keyframes pulse {
  0%, 100% { opacity: 1; }
  50% { opacity: 0.45; }
}

.vh-empty {
  text-align: center;
  padding: 6rem 0;
  color: var(--a-color-muted);
  font-size: 0.9rem;
}
</style>
