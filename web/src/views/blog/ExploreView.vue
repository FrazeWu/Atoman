<template>
  <div class="max-w-6xl mx-auto px-8 py-12 pb-48">
    <div class="mb-10">
      <h1 class="text-5xl font-black tracking-tighter border-l-8 border-black pl-6">
        发现广场
      </h1>
      <p class="mt-3 text-gray-500 font-medium pl-8">浏览所有人的公开文章</p>
    </div>

    <!-- Loading -->
    <div v-if="loading" class="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-6">
      <div
        v-for="i in 6"
        :key="i"
        class="bg-gray-100 border-2 border-black h-64 animate-pulse"
      />
    </div>

    <!-- Empty -->
    <div v-else-if="!posts.length" class="text-center py-24">
      <p class="text-4xl font-black tracking-tighter text-gray-300 mb-4">还没有文章</p>
      <p class="text-gray-500 font-medium">成为第一个发布的人吧</p>
      <RouterLink
        v-if="authStore.isAuthenticated"
        to="/blog/posts/new"
        class="inline-block mt-6 bg-black text-white px-8 py-3 font-black uppercase tracking-widest hover:bg-white hover:text-black border-2 border-black transition-all"
      >
        写文章
      </RouterLink>
    </div>

    <!-- Post Grid -->
    <div v-else class="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-6">
      <PostCard v-for="post in posts" :key="post.id" :post="post" />
    </div>

    <!-- Pagination -->
    <div v-if="totalPages > 1" class="flex justify-center gap-2 mt-12">
      <button
        v-for="p in totalPages"
        :key="p"
        @click="loadPage(p)"
        class="w-10 h-10 border-2 border-black font-black text-sm transition-all"
        :class="p === currentPage ? 'bg-black text-white' : 'hover:bg-black hover:text-white'"
      >
        {{ p }}
      </button>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { RouterLink } from 'vue-router'
import PostCard from '@/components/blog/PostCard.vue'
import { useAuthStore } from '@/stores/auth'
import { useApi } from '@/composables/useApi'
import type { Post } from '@/types'

const authStore = useAuthStore()
const api = useApi()

const posts = ref<Post[]>([])
const loading = ref(true)
const currentPage = ref(1)
const totalPages = ref(1)
const pageSize = 12

const loadPage = async (page: number) => {
  loading.value = true
  currentPage.value = page
  try {
    const res = await fetch(`${api.blog.explore}?page=${page}&limit=${pageSize}`)
    if (res.ok) {
      const data = await res.json()
      posts.value = data.data || []
      totalPages.value = Math.ceil((data.total || posts.value.length) / pageSize) || 1
    }
  } catch (e) {
    console.error('Failed to fetch explore posts', e)
  } finally {
    loading.value = false
  }
}

onMounted(() => loadPage(1))
</script>
