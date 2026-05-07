<template>
  <div class="a-page" style="padding-bottom:12rem">
    <APageHeader title="发现广场" accent sub="浏览所有人的公开文章" style="margin-bottom:2.5rem" />

    <!-- Loading -->
    <div v-if="loading" class="a-grid-3">
      <div v-for="i in 6" :key="i" class="a-skeleton" style="height:16rem" />
    </div>

    <!-- Empty -->
    <div v-else-if="!posts.length" style="text-align:center;padding:6rem 0">
      <p class="a-title a-muted" style="margin-bottom:1rem">还没有文章</p>
      <p class="a-muted" style="margin-bottom:1.5rem">成为第一个发布的人吧</p>
      <ABtn v-if="authStore.isAuthenticated" to="/blog">去选合集写文章</ABtn>
    </div>

    <!-- Post Grid -->
    <div v-else class="a-grid-3">
      <PostCard v-for="post in posts" :key="post.id" :post="post" />
    </div>

    <!-- Pagination -->
    <div v-if="totalPages > 1" style="display:flex;justify-content:center;gap:.5rem;margin-top:3rem">
      <button
        v-for="p in totalPages"
        :key="p"
        @click="loadPage(p)"
        class="a-tab-btn"
        :class="{ 'a-tab-btn-active': p === currentPage }"
        style="width:2.5rem;height:2.5rem;border:2px solid #000"
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
import ABtn from '@/components/ui/ABtn.vue'
import APageHeader from '@/components/ui/APageHeader.vue'
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
