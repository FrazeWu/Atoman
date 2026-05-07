<template>
  <div class="a-page-xl" style="padding-bottom:12rem">
    <APageHeader title="稍后阅读" accent sub="你保存的 RSS 阅读队列" style="margin-bottom:2.5rem">
      <template #action>
        <RouterLink to="/feed" class="a-btn-outline-sm" style="text-decoration:none">← 返回订阅</RouterLink>
      </template>
    </APageHeader>

    <div v-if="loading" style="display:flex;flex-direction:column;gap:1rem">
      <div v-for="i in 5" :key="i" class="a-skeleton" style="height:7rem" />
    </div>

    <AEmpty v-else-if="!items.length" text="阅读列表为空" sub="在订阅时间线中点击「稍后读」保存" />

    <div v-else style="display:flex;flex-direction:column;gap:1rem">
      <div v-for="entry in items" :key="entry.feed_item_id" class="a-card">
        <div v-if="entry.feed_item" style="display:flex;gap:1rem;align-items:flex-start">
          <img
            v-if="entry.feed_item.image_url"
            :src="entry.feed_item.image_url"
            style="width:4rem;height:4rem;object-fit:cover;border:2px solid #000;filter:grayscale(100%);flex-shrink:0"
          />
          <span v-else class="a-badge" style="flex-shrink:0">RSS</span>

          <div style="flex:1;min-width:0">
            <div style="display:flex;align-items:center;gap:.75rem;margin-bottom:.5rem;flex-wrap:wrap">
              <span class="a-label a-muted">{{ entry.feed_item.feed_source?.title || 'RSS' }}</span>
              <span v-if="entry.feed_item.author" class="a-label a-muted">· {{ entry.feed_item.author }}</span>
              <span style="font-size:.75rem;color:#d1d5db">{{ formatDate(entry.feed_item.published_at) }}</span>
            </div>

            <h3 style="font-weight:900;font-size:1.125rem;letter-spacing:-0.025em;margin-bottom:.5rem">
              <RouterLink :to="`/feed/item/${entry.feed_item.id}`" class="item-link">
                {{ entry.feed_item.title }}
              </RouterLink>
            </h3>

            <p v-if="entry.feed_item.summary" class="summary">
              {{ stripHtml(entry.feed_item.summary) }}
            </p>

            <div style="display:flex;align-items:center;gap:1rem;margin-top:.75rem;flex-wrap:wrap">
              <a
                v-if="entry.feed_item.link"
                :href="entry.feed_item.link"
                target="_blank"
                rel="noopener noreferrer"
                style="font-size:.75rem;color:#6b7280;text-decoration:underline"
              >查看原文</a>
              <button class="remove-btn" @click="remove(entry.feed_item_id)">移出列表</button>
            </div>
          </div>
        </div>
      </div>

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
import type { FeedItem } from '@/types'

interface ReadingListEntry {
  feed_item_id: string
  feed_item?: FeedItem
  created_at: string
}

const authStore = useAuthStore()
const feedStore = useFeedStore()
const API_URL = import.meta.env.VITE_API_URL || '/api'
const authHeaders = () => ({ Authorization: `Bearer ${authStore.token}` })

const loading = ref(true)
const loadingMore = ref(false)
const items = ref<ReadingListEntry[]>([])
const page = ref(1)
const hasMore = ref(false)
const pageLimit = 20

const formatDate = (d?: string) => {
  if (!d) return ''
  return new Date(d).toLocaleDateString('zh-CN', { year: 'numeric', month: 'short', day: 'numeric' })
}

const stripHtml = (html: string) =>
  html.replace(/<[^>]*>/g, '').replace(/&amp;/g, '&').replace(/&lt;/g, '<').replace(/&gt;/g, '>').replace(/&quot;/g, '"').trim()

const fetchItems = async (p = 1, append = false) => {
  if (!authStore.isAuthenticated) return
  const res = await fetch(`${API_URL}/feed/reading-list?page=${p}&limit=${pageLimit}`, {
    headers: authHeaders(),
  })
  if (!res.ok) return

  const data = await res.json()
  const nextItems: ReadingListEntry[] = data.items || []
  items.value = append ? [...items.value, ...nextItems] : nextItems
  hasMore.value = items.value.length < (data.total || 0)
}

const loadMore = async () => {
  loadingMore.value = true
  page.value++
  await fetchItems(page.value, true)
  loadingMore.value = false
}

const remove = async (feedItemId: string) => {
  const res = await fetch(`${API_URL}/feed/reading-list/${feedItemId}`, {
    method: 'DELETE',
    headers: authHeaders(),
  })
  if (!res.ok) return

  items.value = items.value.filter((item) => item.feed_item_id !== feedItemId)
  feedStore.readingListItemIds.delete(feedItemId)
}

onMounted(async () => {
  await fetchItems(1)
  loading.value = false
})
</script>

<style scoped>
.item-link {
  color: #000;
  text-decoration: none;
}

.item-link:hover {
  text-decoration: underline;
}

.summary {
  font-size: .875rem;
  color: #6b7280;
  display: -webkit-box;
  -webkit-line-clamp: 2;
  -webkit-box-orient: vertical;
  overflow: hidden;
}

.remove-btn {
  font-size: .7rem;
  font-weight: 900;
  padding: .2rem .5rem;
  border: 1px solid #fca5a5;
  background: #fff;
  color: #ef4444;
  cursor: pointer;
  text-transform: uppercase;
  letter-spacing: .05em;
}
</style>
