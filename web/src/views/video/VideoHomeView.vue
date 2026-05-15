<script setup lang="ts">
import { ref, onMounted, watch } from 'vue'
import APageHeader from '@/components/ui/APageHeader.vue'
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
    <APageHeader title="视频" accent sub="探索视频内容">
      <template #action>
        <ABtn v-if="authStore.isAuthenticated" to="/video/new" variant="primary">+ 上传视频</ABtn>
      </template>
    </APageHeader>

    <!-- Sort tabs -->
    <div class="vh-tabs">
      <button
        v-for="s in [{ v: 'latest', label: '最新' }, { v: 'popular', label: '最热' }]"
        :key="s.v"
        class="vh-tab"
        :class="{ 'vh-tab--active': sort === s.v }"
        @click="sort = s.v as 'latest' | 'popular'"
      >{{ s.label }}</button>
    </div>

    <div v-if="loading" class="vh-grid">
      <div v-for="i in 8" :key="i" class="vh-skeleton" />
    </div>
    <div v-else-if="videos.length === 0" class="vh-empty">暂无视频</div>
    <div v-else class="vh-grid">
      <VideoCard v-for="v in videos" :key="v.id" :video="v" />
    </div>
  </div>
</template>

<style scoped>
.vh-wrap {
  max-width: 80rem;
  margin: 0 auto;
  padding: 2rem 1.5rem 6rem;
}

.vh-tabs {
  display: flex;
  gap: 0.5rem;
  margin-bottom: 1.5rem;
}
.vh-tab {
  padding: 0.35rem 0.875rem;
  font-size: 0.8rem;
  font-weight: 700;
  border: 1px solid var(--a-color-disabled-border, #e5e7eb);
  border-radius: 9999px;
  background: none;
  cursor: pointer;
  color: var(--a-color-muted, #6b7280);
  transition: all 0.15s;
  letter-spacing: 0.02em;
}
.vh-tab:hover {
  border-color: var(--a-color-muted, #6b7280);
  color: var(--a-color-fg);
}
.vh-tab--active {
  background: var(--a-color-fg);
  color: var(--a-color-bg);
  border-color: var(--a-color-fg);
}

.vh-grid {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(15rem, 1fr));
  gap: 1.25rem;
}

.vh-skeleton {
  aspect-ratio: 16/9;
  background: var(--a-color-surface, #f3f4f6);
  border-radius: 0.375rem;
  animation: pulse 1.5s ease-in-out infinite;
}
@keyframes pulse {
  0%, 100% { opacity: 1; }
  50% { opacity: 0.5; }
}

.vh-empty {
  text-align: center;
  padding: 6rem 0;
  color: var(--a-color-muted, #9ca3af);
}
</style>
