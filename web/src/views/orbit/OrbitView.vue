<template>
  <div class="max-w-7xl mx-auto px-8 py-12 pb-48">
    <!-- Header -->
    <div class="flex items-center justify-between mb-10">
      <div>
        <h1 class="text-5xl font-black tracking-tighter border-l-8 border-black pl-6">订阅</h1>
        <p class="mt-2 text-gray-500 font-medium pl-8">聚合你感兴趣的 RSS 订阅源</p>
      </div>
      <button
        v-if="authStore.isAuthenticated"
        @click="showAddModal = true"
        class="bg-black text-white px-6 py-3 font-black uppercase tracking-widest border-2 border-black hover:bg-white hover:text-black transition-all"
      >
        + 添加订阅
      </button>
    </div>

    <!-- Not logged in -->
    <div v-if="!authStore.isAuthenticated" class="min-h-[50vh] flex flex-col items-center justify-center text-center">
      <p class="text-6xl font-black tracking-tighter text-gray-200 mb-6">订阅</p>
      <p class="text-gray-500 font-medium max-w-md mb-8">登录后即可添加 RSS 源，构建你的个性化信息流。</p>
      <RouterLink
        to="/login"
        class="bg-black text-white px-8 py-4 font-black uppercase tracking-widest border-2 border-black hover:bg-white hover:text-black transition-all"
      >
        登录
      </RouterLink>
    </div>

    <template v-else>
      <div class="flex gap-8">
        <!-- Left: Subscription list -->
        <div class="w-72 flex-shrink-0">
          <div class="border-2 border-black">
            <button
              @click="activeSourceId = null"
              class="w-full text-left px-5 py-4 font-black text-sm uppercase tracking-widest border-b-2 border-black transition-all"
              :class="activeSourceId === null ? 'bg-black text-white hover:bg-black/90' : 'hover:underline hover:decoration-2 hover:underline-offset-4'"
            >
              全部订阅
            </button>
            <div v-if="loadingSubscriptions" class="p-4">
              <div v-for="i in 4" :key="i" class="h-12 bg-gray-100 animate-pulse mb-2" />
            </div>
            <div v-else-if="!subscriptions.length" class="p-6 text-center text-gray-400 text-sm font-medium">
              还没有订阅
            </div>
            <button
              v-for="sub in subscriptions"
              :key="sub.id"
              @click="activeSourceId = sub.id"
              class="w-full text-left px-5 py-4 border-b border-gray-100 transition-all flex items-start justify-between group"
              :class="activeSourceId === sub.id ? 'bg-black text-white' : 'hover:underline hover:decoration-2 hover:underline-offset-4'"
            >
              <div class="flex-1 min-w-0">
                <!-- Source type badge -->
                <span
                  class="text-xs font-black uppercase tracking-widest block mb-1"
                  :class="activeSourceId === sub.id ? 'text-gray-300' : 'text-gray-400'"
                >
                  {{ sourceTypeLabel(sub.feed_source?.source_type || '') }}
                </span>
                <span class="font-bold text-sm leading-tight block truncate">
                  {{ sub.title || sub.feed_source?.title || '未命名' }}
                </span>
              </div>
              <button
                @click.stop="unsubscribeSource(sub.id)"
                class="ml-2 opacity-0 group-hover:opacity-100 transition-opacity text-xs font-black flex-shrink-0 mt-1"
                :class="activeSourceId === sub.id ? 'text-gray-300' : 'text-red-500'"
              >
                ✕
              </button>
            </button>
          </div>
        </div>

        <!-- Right: Timeline -->
        <div class="flex-1 min-w-0">
          <!-- Loading -->
          <div v-if="loadingTimeline" class="space-y-4">
            <div v-for="i in 5" :key="i" class="border-2 border-black p-6 flex gap-4 animate-pulse">
              <div class="w-12 h-12 bg-gray-100 flex-shrink-0" />
              <div class="flex-1 space-y-2">
                <div class="h-5 bg-gray-100 w-3/4" />
                <div class="h-4 bg-gray-100 w-1/2" />
                <div class="h-4 bg-gray-100 w-full" />
              </div>
            </div>
          </div>

          <!-- Empty state -->
          <div v-else-if="!timeline.length" class="border-2 border-dashed border-gray-300 py-24 text-center">
            <p class="text-3xl font-black tracking-tighter text-gray-300 mb-3">
              {{ subscriptions.length ? '暂无内容' : '订阅后开始探索' }}
            </p>
            <p class="text-gray-400 font-medium text-sm">
              {{ subscriptions.length ? '订阅源暂无更新' : '点击右上角 + 添加订阅' }}
            </p>
          </div>

          <!-- Timeline items -->
          <div v-else class="space-y-4">
            <template v-for="item in timeline" :key="itemKey(item)">
              <!-- Internal Post -->
              <RouterLink
                v-if="item.type === 'post' && item.post"
                :to="`/blog/posts/${item.post.id}`"
                class="block border-2 border-black p-6 hover:shadow-[8px_8px_0px_0px_rgba(0,0,0,1)] transition-all duration-300 group"
              >
                <div class="flex gap-4 items-start">
                  <div class="px-2 py-1 border-2 border-black flex items-center justify-center font-black bg-gray-50 group-hover:bg-black group-hover:text-white transition-all flex-shrink-0 text-xs uppercase tracking-tighter">
                    博客
                  </div>
                  <div class="flex-1 min-w-0">
                    <div class="flex items-center gap-3 mb-2 flex-wrap">
                      <span class="text-xs font-black uppercase tracking-widest text-gray-400">
                        {{ item.post.user?.display_name || item.post.user?.username || '未知作者' }}
                      </span>
                      <span class="text-xs text-gray-300 font-medium">
                        {{ formatDate(item.published_at) }}
                      </span>
                      <span class="ml-auto text-xs font-black uppercase tracking-widest text-gray-200 group-hover:text-black transition-colors">
                        内部文章 →
                      </span>
                    </div>
                    <h3 class="text-lg font-black tracking-tight leading-tight mb-2 group-hover:underline line-clamp-2">
                      {{ item.post.title }}
                    </h3>
                    <p v-if="item.post.summary" class="text-sm text-gray-600 font-medium leading-relaxed line-clamp-2">
                      {{ item.post.summary }}
                    </p>
                  </div>
                </div>
              </RouterLink>

              <!-- External RSS Item -->
              <a
                v-else-if="item.type === 'orbit_item' && item.orbit_item"
                :href="item.orbit_item.link"
                target="_blank"
                rel="noopener noreferrer"
                class="block border-2 border-black p-6 hover:shadow-[8px_8px_0px_0px_rgba(0,0,0,1)] transition-all duration-300 group cursor-pointer"
              >
                <div class="flex gap-4 items-start">
                  <div class="px-2 py-1 border-2 border-black flex items-center justify-center font-black bg-gray-50 group-hover:bg-black group-hover:text-white transition-all flex-shrink-0 text-xs uppercase tracking-tighter">
                    外部
                  </div>
                  <div class="flex-1 min-w-0">
                    <div class="flex items-center gap-3 mb-2 flex-wrap">
                      <span class="text-xs font-black uppercase tracking-widest text-gray-400">
                        {{ item.orbit_item.author || item.orbit_item.feed_source?.title || 'RSS' }}
                      </span>
                      <span class="text-xs text-gray-300 font-medium">
                        {{ formatDate(item.orbit_item.published_at) }}
                      </span>
                      <span class="ml-auto text-xs font-black uppercase tracking-widest text-gray-300 group-hover:text-black transition-colors">
                        ↗ 外部链接
                      </span>
                    </div>
                    <h3 class="text-lg font-black tracking-tight leading-tight mb-2 group-hover:underline line-clamp-2">
                      {{ item.orbit_item.title }}
                    </h3>
                    <p v-if="item.orbit_item.summary" class="text-sm text-gray-600 font-medium leading-relaxed line-clamp-2">
                      {{ stripHtml(item.orbit_item.summary) }}
                    </p>
                  </div>
                </div>
              </a>
            </template>
          </div>
        </div>
      </div>
    </template>

    <!-- Add Subscription Modal -->
    <div v-if="showAddModal" class="fixed inset-0 z-50 flex items-center justify-center bg-black/50" @click.self="showAddModal = false">
      <div class="bg-white border-2 border-black p-8 w-full max-w-lg shadow-[20px_20px_0px_0px_rgba(0,0,0,1)]">
        <h3 class="text-2xl font-black tracking-tight mb-6">添加订阅</h3>

        <div class="flex flex-col gap-4">
          <div>
            <label class="text-xs font-black uppercase tracking-widest text-gray-500 block mb-2">RSS 地址 *</label>
            <input
              v-model="newRssUrl"
              placeholder="https://example.com/feed.xml"
              class="w-full border-2 border-black p-3 font-medium focus:shadow-[4px_4px_0px_0px_rgba(0,0,0,1)] outline-none transition-all"
            />
          </div>
          <div>
            <label class="text-xs font-black uppercase tracking-widest text-gray-500 block mb-2">自定义名称（可选）</label>
            <input
              v-model="newRssTitle"
              placeholder="例如：GitHub Blog"
              class="w-full border-2 border-black p-3 font-medium focus:outline-none"
            />
          </div>
        </div>

        <!-- Error -->
        <div v-if="addError" class="mt-4 border-2 border-red-500 bg-red-50 p-3 text-red-700 text-sm font-bold">
          {{ addError }}
        </div>

        <!-- Actions -->
        <div class="flex gap-3 mt-6">
          <button
            @click="addSubscription"
            :disabled="adding"
            class="flex-1 bg-black text-white py-3 font-black uppercase tracking-widest text-sm border-2 border-black hover:bg-white hover:text-black transition-all disabled:opacity-40"
          >
            {{ adding ? '添加中...' : '添加' }}
          </button>
          <button
            @click="showAddModal = false"
            class="px-6 py-3 font-black uppercase tracking-widest text-sm border-2 border-black hover:bg-black hover:text-white transition-all"
          >
            取消
          </button>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, watch, onMounted } from 'vue'
import { RouterLink } from 'vue-router'
import { useAuthStore } from '@/stores/auth'
import type { Subscription, OrbitItem, TimelineItem } from '@/types'

const authStore = useAuthStore()

const API_URL = import.meta.env.VITE_API_URL || '/api'

const authHeaders = () => ({ Authorization: `Bearer ${authStore.token}` })

// State
const subscriptions = ref<Subscription[]>([])
const timeline = ref<TimelineItem[]>([])
const activeSourceId = ref<number | null>(null)
const loadingSubscriptions = ref(true)
const loadingTimeline = ref(false)

// Modal state
const showAddModal = ref(false)
const newRssUrl = ref('')
const newRssTitle = ref('')
const addError = ref('')
const adding = ref(false)

// Helpers
const itemKey = (item: TimelineItem) => {
  if (item.type === 'post' && item.post) return `post-${item.post.id}`
  if (item.type === 'orbit_item' && item.orbit_item) return `orbit-${item.orbit_item.id}`
  return Math.random().toString()
}

const sourceTypeLabel = (type: string) => {
  if (type === 'external_rss') return 'RSS'
  if (type === 'internal_user') return '用户'
  if (type === 'internal_channel') return '频道'
  if (type === 'internal_collection') return '合集'
  return type
}

const sourceTypeIcon = (type: string) => {
  if (type === 'external_rss') return '📡'
  if (type === 'internal_user') return '👤'
  if (type === 'internal_channel') return '📺'
  return '📄'
}

const formatDate = (d?: string) => {
  if (!d) return ''
  const date = new Date(d)
  return date.toLocaleDateString('zh-CN', { month: 'short', day: 'numeric' })
}

const stripHtml = (html: string) => {
  return html.replace(/<[^>]*>/g, '').replace(/&amp;/g, '&').replace(/&lt;/g, '<').replace(/&gt;/g, '>').replace(/&quot;/g, '"').trim()
}

// Fetch
const fetchSubscriptions = async () => {
  if (!authStore.isAuthenticated) return
  loadingSubscriptions.value = true
  try {
    const res = await fetch(`${API_URL}/feed/subscriptions`, { headers: authHeaders() })
    if (res.ok) {
      const d = await res.json()
      subscriptions.value = d.data || []
    }
  } catch (e) {
    console.error(e)
  } finally {
    loadingSubscriptions.value = false
  }
}

const fetchTimeline = async () => {
  if (!authStore.isAuthenticated) return
  loadingTimeline.value = true
  try {
    let url = `${API_URL}/feed/timeline`
    if (activeSourceId.value !== null) {
      url += `?source_id=${activeSourceId.value}`
    }
    const res = await fetch(url, { headers: authHeaders() })
    if (res.ok) {
      const d = await res.json()
      timeline.value = d.data || []
    }
  } catch (e) {
    console.error(e)
  } finally {
    loadingTimeline.value = false
  }
}

const unsubscribeSource = async (id: number) => {
  try {
    await fetch(`${API_URL}/feed/subscriptions/${id}`, {
      method: 'DELETE',
      headers: authHeaders()
    })
    if (activeSourceId.value === id) activeSourceId.value = null
    await fetchSubscriptions()
    await fetchTimeline()
  } catch (e) {
    console.error(e)
  }
}

const addSubscription = async () => {
  addError.value = ''
  adding.value = true
  try {
    if (!newRssUrl.value.trim()) {
      addError.value = 'RSS 地址不能为空'
      return
    }
    const body = {
      target_type: 'external_rss',
      rss_url: newRssUrl.value.trim(),
      title: newRssTitle.value.trim() // Empty title allows backend/frontend fallback to RSS title
    }

    const res = await fetch(`${API_URL}/feed/subscriptions`, {
      method: 'POST',
      headers: { 'Content-Type': 'application/json', ...authHeaders() },
      body: JSON.stringify(body)
    })

    if (res.ok) {
      showAddModal.value = false
      newRssUrl.value = ''
      newRssTitle.value = ''
      await fetchSubscriptions()
      await fetchTimeline()
    } else {
      const err = await res.json()
      addError.value = err.error || '添加失败'
    }
  } catch (e) {
    addError.value = '网络错误，请重试'
  } finally {
    adding.value = false
  }
}

watch(activeSourceId, fetchTimeline)
watch(showAddModal, (v) => { if (!v) addError.value = '' })

onMounted(async () => {
  if (authStore.isAuthenticated) {
    await fetchSubscriptions()
    await fetchTimeline()
  }
})
</script>
