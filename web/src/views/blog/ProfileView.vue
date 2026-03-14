<template>
  <div class="max-w-5xl mx-auto px-8 py-12 pb-48">
    <!-- Loading -->
    <div v-if="loading" class="space-y-6">
      <div class="h-32 bg-gray-100 border-2 border-black animate-pulse" />
      <div class="h-8 bg-gray-100 animate-pulse w-1/2" />
    </div>

    <!-- Not found -->
    <div v-else-if="!profile" class="text-center py-24">
      <p class="text-4xl font-black text-gray-200">用户不存在</p>
      <RouterLink to="/blog" class="mt-6 inline-block font-black border-b-2 border-black">← 博客首页</RouterLink>
    </div>

    <template v-else>
      <!-- Profile header -->
      <div class="border-2 border-black p-8 mb-8 flex flex-col md:flex-row gap-6 items-start">
        <!-- Avatar -->
        <div class="w-20 h-20 rounded-full bg-black flex items-center justify-center text-white text-3xl font-black flex-shrink-0">
          {{ (profile.display_name || profile.username).charAt(0).toUpperCase() }}
        </div>

        <div class="flex-1">
          <div class="flex items-start justify-between gap-4 flex-wrap">
            <div>
              <h1 class="text-3xl font-black tracking-tight">{{ profile.display_name || profile.username }}</h1>
              <p class="text-gray-500 font-medium text-sm">@{{ profile.username }}</p>
            </div>

            <!-- Follow / Edit button -->
            <div class="flex gap-2">
              <button
                v-if="authStore.isAuthenticated && !isSelf"
                @click="toggleFollow"
                class="font-black uppercase tracking-widest text-sm border-2 border-black px-5 py-2 transition-all"
                :class="following ? 'bg-black text-white hover:bg-white hover:text-black' : 'hover:bg-black hover:text-white'"
              >
                {{ following ? '已订阅' : '订阅' }}
              </button>
              <RouterLink
                v-if="isSelf"
                to="/blog/settings"
                class="font-black uppercase tracking-widest text-sm border-2 border-black px-5 py-2 hover:bg-black hover:text-white transition-all"
              >
                编辑资料
              </RouterLink>
            </div>
          </div>

          <!-- Stats -->
          <div class="flex gap-6 mt-4 text-sm font-black">
            <span><span class="text-xl">{{ profile.posts_count ?? posts.length }}</span> 篇文章</span>
            <span><span class="text-xl">{{ profile.followers_count ?? 0 }}</span> 订阅者</span>
            <span><span class="text-xl">{{ profile.following_count ?? 0 }}</span> 已订阅</span>
          </div>

          <p v-if="profile.bio" class="mt-3 text-gray-600 font-medium">{{ profile.bio }}</p>
          <a v-if="profile.website" :href="profile.website" target="_blank" class="mt-1 inline-block text-sm font-black underline hover:opacity-60">
            {{ profile.website }}
          </a>
        </div>
      </div>

      <!-- Posts -->
      <div class="flex items-center justify-between mb-6">
        <h2 class="text-2xl font-black tracking-tight">文章</h2>
      </div>

      <div v-if="loadingPosts" class="grid grid-cols-1 md:grid-cols-2 gap-6">
        <div v-for="i in 4" :key="i" class="h-40 bg-gray-100 border-2 border-black animate-pulse" />
      </div>
      <div v-else-if="!posts.length" class="border-2 border-dashed border-gray-300 p-10 text-center text-gray-400 font-medium">
        还没有发布文章
      </div>
      <div v-else class="grid grid-cols-1 md:grid-cols-2 gap-6">
        <PostCard v-for="post in posts" :key="post.id" :post="post" />
      </div>
    </template>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import { RouterLink, useRoute } from 'vue-router'
import PostCard from '@/components/blog/PostCard.vue'
import { useAuthStore } from '@/stores/auth'
import { useApi } from '@/composables/useApi'
import type { UserProfile, Post } from '@/types'

const route = useRoute()
const authStore = useAuthStore()
const api = useApi()

const profile = ref<UserProfile | null>(null)
const posts = ref<Post[]>([])
const loading = ref(true)
const loadingPosts = ref(true)
const following = ref(false)

const username = computed(() => route.params.username as string)
const isSelf = computed(() => authStore.user?.username === username.value)

const fetchProfile = async () => {
  loading.value = true
  try {
    const res = await fetch(api.users.profile(username.value))
    if (res.ok) {
      const d = await res.json()
      profile.value = d.data || d
    }
  } catch (e) {
    console.error(e)
  } finally {
    loading.value = false
  }
}

const fetchPosts = async () => {
  if (!profile.value) return
  loadingPosts.value = true
  try {
    const res = await fetch(`${api.blog.posts}?user_id=${profile.value.id}&status=published`)
    if (res.ok) {
      const d = await res.json()
      posts.value = d.data || []
    }
  } catch (e) {
    console.error(e)
  } finally {
    loadingPosts.value = false
  }
}

const toggleFollow = async () => {
  if (!profile.value) return
  const method = following.value ? 'DELETE' : 'POST'
  try {
    const res = await fetch(api.users.follow(profile.value.id), {
      method,
      headers: { Authorization: `Bearer ${authStore.token}` }
    })
    if (res.ok) following.value = !following.value
  } catch (e) {
    console.error(e)
  }
}

onMounted(async () => {
  await fetchProfile()
  await fetchPosts()
})
</script>
