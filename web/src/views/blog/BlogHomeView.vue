<template>
  <div class="max-w-6xl mx-auto px-8 py-12 pb-48">
    <!-- Logged-out state -->
    <div v-if="!authStore.isAuthenticated" class="min-h-[60vh] flex flex-col items-center justify-center text-center">
      <h1 class="text-6xl font-black tracking-tighter mb-6">ATOMAN<br />BLOG</h1>
      <p class="text-gray-500 font-medium max-w-md mb-8">
        创作你的博客，订阅他人内容，构建属于你的知识图谱。
      </p>
      <div class="flex gap-4">
        <RouterLink
          to="/login"
          class="bg-black text-white px-8 py-4 font-black uppercase tracking-widest border-2 border-black hover:bg-white hover:text-black transition-all"
        >
          登录
        </RouterLink>
        <RouterLink
          to="/blog/explore"
          class="px-8 py-4 font-black uppercase tracking-widest border-2 border-black hover:bg-black hover:text-white transition-all"
        >
          浏览文章
        </RouterLink>
      </div>
    </div>

    <!-- Logged-in state -->
    <template v-else>
      <div class="flex items-center justify-between mb-10">
        <h1 class="text-5xl font-black tracking-tighter border-l-8 border-black pl-6">
          我的博客
        </h1>
        <RouterLink
          to="/blog/posts/new"
          class="bg-black text-white px-6 py-3 font-black uppercase tracking-widest border-2 border-black hover:bg-white hover:text-black transition-all"
        >
          + 写文章
        </RouterLink>
      </div>

      <!-- Channels section -->
      <section class="mb-12">
        <div class="flex items-center justify-between mb-4">
          <h2 class="text-2xl font-black tracking-tight">我的频道</h2>
          <button
            @click="showCreateChannel = true"
            class="text-xs font-black uppercase tracking-widest border-2 border-black px-4 py-2 hover:bg-black hover:text-white transition-colors"
          >
            + 新建频道
          </button>
        </div>

        <div v-if="!channels.length" class="border-2 border-dashed border-gray-300 p-8 text-center text-gray-400 font-medium">
          还没有频道，创建一个来组织你的文章
        </div>
        <div v-else class="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-4">
          <RouterLink
            v-for="ch in channels"
            :key="ch.id"
            :to="`/blog/explore?channel=${ch.id}`"
            class="border-2 border-black p-5 hover:shadow-[6px_6px_0px_0px_rgba(0,0,0,1)] transition-all duration-300 flex flex-col gap-1"
          >
            <span class="font-black text-lg tracking-tight">{{ ch.name }}</span>
            <span v-if="ch.description" class="text-sm text-gray-500 font-medium line-clamp-2">{{ ch.description }}</span>
          </RouterLink>
        </div>
      </section>

      <!-- Recent posts -->
      <section>
        <div class="flex items-center justify-between mb-4">
          <h2 class="text-2xl font-black tracking-tight">最近文章</h2>
          <RouterLink to="/blog/explore" class="text-xs font-black uppercase tracking-widest border-b-2 border-black hover:opacity-60">
            发现广场 →
          </RouterLink>
        </div>

        <div v-if="loadingPosts" class="grid grid-cols-1 md:grid-cols-2 gap-6">
          <div v-for="i in 4" :key="i" class="h-40 bg-gray-100 border-2 border-black animate-pulse" />
        </div>
        <div v-else-if="!recentPosts.length" class="border-2 border-dashed border-gray-300 p-8 text-center text-gray-400 font-medium">
          你还没有发布任何文章
        </div>
        <div v-else class="grid grid-cols-1 md:grid-cols-2 gap-6">
          <PostCard v-for="post in recentPosts" :key="post.id" :post="post" />
        </div>
      </section>
    </template>

    <!-- Create Channel Modal -->
    <div v-if="showCreateChannel" class="fixed inset-0 z-50 flex items-center justify-center bg-black/50">
      <div class="bg-white border-2 border-black p-8 w-full max-w-md shadow-[20px_20px_0px_0px_rgba(0,0,0,1)]">
        <h3 class="text-2xl font-black tracking-tight mb-6">创建频道</h3>
        <div class="flex flex-col gap-4">
          <input
            v-model="newChannelName"
            placeholder="频道名称*"
            class="w-full border-2 border-black p-3 font-medium focus:shadow-[4px_4px_0px_0px_rgba(0,0,0,1)] outline-none transition-all"
          />
          <textarea
            v-model="newChannelDesc"
            placeholder="频道描述（可选）"
            rows="3"
            class="w-full border-2 border-black p-3 font-medium focus:shadow-[4px_4px_0px_0px_rgba(0,0,0,1)] outline-none transition-all resize-none"
          />
        </div>
        <div class="flex gap-3 mt-6">
          <button
            @click="createChannel"
            class="flex-1 bg-black text-white py-3 font-black uppercase tracking-widest border-2 border-black hover:bg-white hover:text-black transition-all"
          >
            创建
          </button>
          <button
            @click="showCreateChannel = false"
            class="px-6 py-3 font-black uppercase tracking-widest border-2 border-black hover:bg-black hover:text-white transition-all"
          >
            取消
          </button>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { RouterLink } from 'vue-router'
import PostCard from '@/components/blog/PostCard.vue'
import { useAuthStore } from '@/stores/auth'
import { useApi } from '@/composables/useApi'
import type { Channel, Post } from '@/types'

const authStore = useAuthStore()
const api = useApi()

const channels = ref<Channel[]>([])
const recentPosts = ref<Post[]>([])
const loadingPosts = ref(true)
const showCreateChannel = ref(false)
const newChannelName = ref('')
const newChannelDesc = ref('')

const fetchMyData = async () => {
  if (!authStore.isAuthenticated) return
  try {
    const [chRes, postRes] = await Promise.all([
      fetch(api.blog.channels, { headers: { Authorization: `Bearer ${authStore.token}` } }),
      fetch(`${api.blog.posts}?user_id=${authStore.user?.id}&limit=6`, { headers: { Authorization: `Bearer ${authStore.token}` } })
    ])
    if (chRes.ok) channels.value = (await chRes.json()).data || []
    if (postRes.ok) {
      const d = await postRes.json()
      recentPosts.value = (d.data || []).filter((p: Post) => p.status === 'published')
    }
  } catch (e) {
    console.error(e)
  } finally {
    loadingPosts.value = false
  }
}

const createChannel = async () => {
  if (!newChannelName.value.trim()) return
  try {
    const res = await fetch(api.blog.channels, {
      method: 'POST',
      headers: { 'Content-Type': 'application/json', Authorization: `Bearer ${authStore.token}` },
      body: JSON.stringify({ name: newChannelName.value, description: newChannelDesc.value })
    })
    if (res.ok) {
      showCreateChannel.value = false
      newChannelName.value = ''
      newChannelDesc.value = ''
      await fetchMyData()
    }
  } catch (e) {
    console.error(e)
  }
}

onMounted(fetchMyData)
</script>
