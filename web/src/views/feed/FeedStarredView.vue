<template>
  <div class="a-page-xl" style="padding-bottom:12rem">
    <APageHeader title="收藏" accent sub="你收藏的 RSS 文章" style="margin-bottom:2.5rem">
      <template #action>
        <RouterLink to="/feed" class="a-btn-outline-sm" style="text-decoration:none">← 返回订阅</RouterLink>
      </template>
    </APageHeader>

    <div v-if="loading" style="display:flex;flex-direction:column;gap:1rem">
      <div v-for="i in 5" :key="i" class="a-skeleton" style="height:7rem" />
    </div>

    <AEmpty v-else-if="!items.length" text="还没有收藏任何文章" sub="在订阅时间线中点击「收藏」保存" />

    <div v-else style="display:flex;flex-direction:column;gap:1rem">
      <div
        v-for="item in items"
        :key="item.id"
        class="a-card"
        style="position:relative"
      >
        <div style="display:flex;gap:1rem;align-items:flex-start">
          <img
            v-if="item.image_url"
            :src="item.image_url"
            style="width:4rem;height:4rem;object-fit:cover;border:2px solid #000;filter:grayscale(100%);flex-shrink:0"
          />
          <span v-else class="a-badge" style="flex-shrink:0">外部</span>

          <div style="flex:1;min-width:0">
            <div style="display:flex;align-items:center;gap:.75rem;margin-bottom:.5rem;flex-wrap:wrap">
              <span class="a-label a-muted">{{ item.source_title || 'RSS' }}</span>
              <span v-if="item.author" class="a-label a-muted">· {{ item.author }}</span>
              <span style="font-size:.75rem;color:#d1d5db">{{ formatDate(item.published_at) }}</span>
            </div>
            <h3 style="font-weight:900;font-size:1.125rem;letter-spacing:-0.025em;margin-bottom:.5rem">
              <RouterLink
                :to="`/feed/item/${item.id}`"
                style="color:#000;text-decoration:none"
                class="hover-underline"
              >{{ item.title }}</RouterLink>
            </h3>
            <p v-if="item.summary" style="font-size:.875rem;color:#6b7280;display:-webkit-box;-webkit-line-clamp:2;-webkit-box-orient:vertical;overflow:hidden">
              {{ stripHtml(item.summary) }}
            </p>
            <div style="display:flex;align-items:center;gap:1rem;margin-top:.5rem">
              <a
                v-if="item.link"
                :href="item.link"
                target="_blank"
                rel="noopener noreferrer"
                style="font-size:.75rem;color:#6b7280;text-decoration:underline"
              >📄 查看原文</a>
              <button
                @click="unstar(item.id)"
                style="font-size:.7rem;font-weight:900;padding:.2rem .5rem;border:1px solid #fca5a5;background:#fff;color:#ef4444;cursor:pointer;text-transform:uppercase;letter-spacing:.05em"
                title="取消收藏"
              >取消收藏</button>
            </div>
          </div>
        </div>
      </div>

      <!-- Pagination -->
      <div v-if="hasMore" style="display:flex;justify-content:center;padding-top:1rem">
        <button
          @click="loadMore"
          :disabled="loadingMore"
          style="font-weight:900;font-size:.75rem;text-transform:uppercase;letter-spacing:.08em;padding:.75rem 2rem;border:2px solid #000;cursor:pointer;background:#fff"
          :style="loadingMore ? 'opacity:.5;cursor:not-allowed' : ''"
        >{{ loadingMore ? '加载中...' : '加载更多' }}</button>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { RouterLink } from 'vue-router'
import APageHeader from '@/components/ui/APageHeader.vue'
import AEmpty from '@/components/ui/AEmpty.vue'
import { useAuthStore } from '@/stores/auth'
import { useFeedStore } from '@/stores/feed'

interface StarredItem {
  id: string
  title: string
  link: string
  summary: string
  author: string
  published_at: string
  image_url?: string
  enclosure_url?: string
  enclosure_type?: string
  source_title: string
}

const authStore = useAuthStore()
const feedStore = useFeedStore()
const API_URL = import.meta.env.VITE_API_URL || '/api'
const authHeaders = () => ({ Authorization: `Bearer ${authStore.token}` })

const loading = ref(true)
const loadingMore = ref(false)
const items = ref<StarredItem[]>([])
const page = ref(1)
const hasMore = ref(false)
const pageLimit = 20

const formatDate = (d?: string) => {
  if (!d) return ''
  return new Date(d).toLocaleDateString('zh-CN', { year: 'numeric', month: 'short', day: 'numeric' })
}

const stripHtml = (html: string) =>
  html.replace(/<[^>]*>/g, '').replace(/&amp;/g, '&').replace(/&lt;/g, '<').replace(/&gt;/g, '>').replace(/&quot;/g, '"').trim()

const fetchStarred = async (p = 1, append = false) => {
  if (!authStore.isAuthenticated) return
  try {
    const res = await fetch(`${API_URL}/feed/stars?page=${p}&limit=${pageLimit}`, {
      headers: authHeaders(),
    })
    if (res.ok) {
      const data = await res.json()
      const newItems: StarredItem[] = data.items || []
      if (append) {
        items.value = [...items.value, ...newItems]
      } else {
        items.value = newItems
      }
      hasMore.value = newItems.length === pageLimit
    }
  } catch (e) {
    console.error('Failed to fetch starred items', e)
  }
}

const loadMore = async () => {
  loadingMore.value = true
  page.value++
  await fetchStarred(page.value, true)
  loadingMore.value = false
}

const unstar = async (feedItemId: string) => {
  const result = await feedStore.toggleStar(feedItemId)
  if (result === false) {
    items.value = items.value.filter((item) => item.id !== feedItemId)
  }
}

onMounted(async () => {
  await fetchStarred(1)
  loading.value = false
})
</script>

<style scoped>
.hover-underline:hover {
  text-decoration: underline;
}
</style>
