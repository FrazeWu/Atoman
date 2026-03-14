<template>
  <div class="pb-48">
    <!-- Loading -->
    <div v-if="loading" class="max-w-3xl mx-auto px-8 py-16 space-y-6">
      <div class="h-12 bg-gray-100 border-2 border-black animate-pulse w-3/4" />
      <div class="h-4 bg-gray-100 animate-pulse w-1/3" />
      <div class="aspect-video bg-gray-100 border-2 border-black animate-pulse" />
      <div class="space-y-3">
        <div v-for="i in 5" :key="i" class="h-4 bg-gray-100 animate-pulse" :style="{ width: `${70 + Math.random() * 30}%` }" />
      </div>
    </div>

    <!-- Not found -->
    <div v-else-if="!post" class="max-w-3xl mx-auto px-8 py-24 text-center">
      <p class="text-5xl font-black text-gray-200 mb-4">404</p>
      <p class="text-gray-500 font-medium">文章不存在</p>
      <RouterLink to="/blog/explore" class="mt-6 inline-block font-black border-b-2 border-black hover:opacity-60">← 返回发现广场</RouterLink>
    </div>

    <!-- Post content -->
    <article v-else>
      <!-- Header with cover -->
      <div v-if="post.cover_url" class="w-full max-h-80 overflow-hidden border-b-2 border-black">
        <img :src="post.cover_url" :alt="post.title" class="w-full object-cover max-h-80" />
      </div>

      <div class="max-w-3xl mx-auto px-8 py-12">
        <!-- Breadcrumb -->
        <RouterLink to="/blog/explore" class="text-xs font-black uppercase tracking-widest text-gray-400 hover:text-black transition-colors">
          ← 博客广场
        </RouterLink>

        <!-- Title -->
        <h1 class="text-4xl md:text-5xl font-black tracking-tighter mt-6 mb-4 leading-tight">
          {{ post.title }}
        </h1>

        <!-- Meta -->
        <div class="flex flex-wrap items-center gap-4 pb-6 border-b-2 border-black mb-10">
          <RouterLink
            :to="`/blog/@${post.user?.username}`"
            class="flex items-center gap-2 group"
          >
            <div class="w-8 h-8 rounded-full bg-black flex items-center justify-center text-white font-black text-sm">
              {{ (post.user?.display_name || post.user?.username || '?').charAt(0).toUpperCase() }}
            </div>
            <span class="font-black text-sm group-hover:underline">
              {{ post.user?.display_name || post.user?.username }}
            </span>
          </RouterLink>

          <span class="text-xs font-black uppercase tracking-widest text-gray-400">
            {{ formatDate(post.created_at) }}
          </span>

          <!-- Edit button (owner only) -->
          <RouterLink
            v-if="isOwner"
            :to="`/blog/posts/${post.id}/edit`"
            class="ml-auto text-xs font-black uppercase tracking-widest border-2 border-black px-4 py-1 hover:bg-black hover:text-white transition-all"
          >
            编辑
          </RouterLink>
        </div>

        <!-- Markdown Content -->
        <div class="prose-blog mb-12" v-html="renderedContent" />

        <!-- Interaction bar -->
        <div class="flex items-center gap-4 py-6 border-y-2 border-black mb-12">
          <!-- Like -->
          <button
            @click="toggleLike"
            class="flex items-center gap-2 font-black text-sm transition-all border-2 px-4 py-2"
            :class="liked ? 'bg-black text-white border-black' : 'border-black hover:bg-black hover:text-white'"
          >
            ♥ {{ likesCount }}
          </button>

          <!-- Bookmark -->
          <button
            v-if="authStore.isAuthenticated"
            @click="toggleBookmark"
            class="flex items-center gap-2 font-black text-sm transition-all border-2 border-black px-4 py-2"
            :class="bookmarked ? 'bg-black text-white' : 'hover:bg-black hover:text-white'"
          >
            {{ bookmarked ? '★ 已收藏' : '☆ 收藏' }}
          </button>

          <!-- RSS link -->
          <a
            v-if="post.user?.username"
            :href="api.orbit.rss(post.user.username)"
            target="_blank"
            class="ml-auto flex items-center gap-1 text-xs font-black uppercase tracking-widest text-gray-400 hover:text-black transition-colors"
          >
            RSS ↗
          </a>
        </div>

        <!-- Comments -->
        <CommentSection
          :post-id="post.id"
          :allow-comments="post.allow_comments"
          :post-owner-id="post.user_id"
        />
      </div>
    </article>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import { RouterLink, useRoute } from 'vue-router'
import { marked } from 'marked'
import CommentSection from '@/components/blog/CommentSection.vue'
import { useAuthStore } from '@/stores/auth'
import { useApi } from '@/composables/useApi'
import type { Post } from '@/types'

const route = useRoute()
const authStore = useAuthStore()
const api = useApi()

const post = ref<Post | null>(null)
const loading = ref(true)
const liked = ref(false)
const likesCount = ref(0)
const bookmarked = ref(false)

const isOwner = computed(() => authStore.user?.id === post.value?.user_id)

const formatDate = (d: string) => new Date(d).toLocaleDateString('zh-CN', { year: 'numeric', month: 'long', day: 'numeric' })

const renderedContent = computed(() => {
  if (!post.value?.content) return ''
  return marked(post.value.content)
})

const fetchPost = async () => {
  loading.value = true
  try {
    const id = route.params.id
    const res = await fetch(api.blog.post(Number(id)))
    if (res.ok) {
      const d = await res.json()
      post.value = d.data || d
      likesCount.value = post.value?.likes_count ?? 0
    }
  } catch (e) {
    console.error(e)
  } finally {
    loading.value = false
  }
}

const toggleLike = async () => {
  if (!authStore.isAuthenticated || !post.value) return
  const method = liked.value ? 'DELETE' : 'POST'
  try {
    const res = await fetch(api.blog.likes, {
      method,
      headers: { 'Content-Type': 'application/json', Authorization: `Bearer ${authStore.token}` },
      body: JSON.stringify({ target_type: 'post', target_id: post.value.id })
    })
    if (res.ok) {
      liked.value = !liked.value
      likesCount.value += liked.value ? 1 : -1
    }
  } catch (e) {
    console.error(e)
  }
}

const toggleBookmark = async () => {
  if (!authStore.isAuthenticated || !post.value) return
  try {
    if (bookmarked.value) {
      // fetch bookmarks to find the id
      const bRes = await fetch(api.blog.bookmarks, { headers: { Authorization: `Bearer ${authStore.token}` } })
      if (bRes.ok) {
        const d = await bRes.json()
        const bm = (d.data || []).find((b: any) => b.post_id === post.value?.id)
        if (bm) {
          await fetch(api.blog.bookmark(bm.id), { method: 'DELETE', headers: { Authorization: `Bearer ${authStore.token}` } })
          bookmarked.value = false
        }
      }
    } else {
      const res = await fetch(api.blog.bookmarks, {
        method: 'POST',
        headers: { 'Content-Type': 'application/json', Authorization: `Bearer ${authStore.token}` },
        body: JSON.stringify({ post_id: post.value.id })
      })
      if (res.ok) bookmarked.value = true
    }
  } catch (e) {
    console.error(e)
  }
}

onMounted(fetchPost)
</script>

<style scoped>
.prose-blog :deep(h1),
.prose-blog :deep(h2),
.prose-blog :deep(h3),
.prose-blog :deep(h4) {
  font-weight: 900;
  letter-spacing: -0.025em;
  margin: 2rem 0 1rem;
  line-height: 1.2;
}
.prose-blog :deep(h1) { font-size: 2rem; }
.prose-blog :deep(h2) { font-size: 1.5rem; border-left: 6px solid black; padding-left: 1rem; }
.prose-blog :deep(h3) { font-size: 1.2rem; }
.prose-blog :deep(p) { margin: 1rem 0; line-height: 1.8; font-size: 1.05rem; color: #1a1a1a; }
.prose-blog :deep(a) { font-weight: 700; text-decoration: underline; }
.prose-blog :deep(a:hover) { opacity: 0.6; }
.prose-blog :deep(code) {
  background: #f3f4f6;
  border: 1px solid #e5e7eb;
  padding: 0.15em 0.4em;
  font-size: 0.9em;
  font-family: 'JetBrains Mono', 'Fira Code', monospace;
}
.prose-blog :deep(pre) {
  background: #111;
  color: #f8f8f2;
  padding: 1.5rem;
  border: 2px solid black;
  overflow-x: auto;
  margin: 1.5rem 0;
}
.prose-blog :deep(pre code) { background: none; border: none; padding: 0; color: inherit; }
.prose-blog :deep(blockquote) {
  border-left: 4px solid black;
  padding: 0.5rem 1.5rem;
  margin: 1.5rem 0;
  font-style: italic;
  color: #555;
  background: #f9f9f9;
}
.prose-blog :deep(ul), .prose-blog :deep(ol) {
  padding-left: 1.5rem;
  margin: 1rem 0;
  line-height: 1.8;
}
.prose-blog :deep(li) { margin: 0.4rem 0; }
.prose-blog :deep(img) { border: 2px solid black; width: 100%; margin: 1.5rem 0; }
.prose-blog :deep(hr) { border: 0; border-top: 2px solid black; margin: 2rem 0; }
.prose-blog :deep(table) { border-collapse: collapse; width: 100%; margin: 1.5rem 0; }
.prose-blog :deep(th), .prose-blog :deep(td) { border: 2px solid black; padding: 0.6rem 1rem; }
.prose-blog :deep(th) { background: black; color: white; font-weight: 900; text-align: left; }
</style>
