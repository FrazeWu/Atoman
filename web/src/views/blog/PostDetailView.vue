<template>
  <div style="padding-bottom:12rem">
    <!-- Loading -->
    <div v-if="loading" class="a-page-md" style="padding-top:4rem">
      <div class="a-skeleton" style="height:3rem;width:75%;margin-bottom:1rem" />
      <div class="a-skeleton" style="height:1rem;width:33%;margin-bottom:1.5rem" />
      <div class="a-skeleton" style="aspect-ratio:16/9;margin-bottom:1.5rem" />
      <div v-for="i in 5" :key="i" class="a-skeleton" style="height:1rem;margin-bottom:.75rem" />
    </div>

    <!-- Not found -->
    <div v-else-if="errorStatus === 404" class="a-page-md" style="padding-top:6rem;text-align:center">
      <p style="font-size:3rem;font-weight:900;color:#e5e7eb;margin-bottom:1rem">404</p>
      <p class="a-muted" style="margin-bottom:1.5rem">文章不存在</p>
      <RouterLink to="/blog/explore" class="a-link">← 返回发现广场</RouterLink>
    </div>

    <!-- Draft (only visible to owner) -->
    <div v-else-if="errorStatus === 403" class="a-page-md" style="padding-top:6rem;text-align:center">
      <p style="font-size:3rem;font-weight:900;color:#e5e7eb;margin-bottom:1rem">草稿</p>
      <p class="a-muted" style="margin-bottom:1.5rem">该文章尚未发布，请登录后查看或编辑</p>
      <RouterLink :to="`/blog/posts/${route.params.id}/edit`" class="a-link">去编辑 →</RouterLink>
    </div>

    <!-- Post content -->
    <article v-else-if="post">
      <!-- Cover image -->
      <div v-if="post.cover_url" style="width:100%;max-height:20rem;overflow:hidden;border-bottom:2px solid #000">
        <img :src="post.cover_url" :alt="post.title" style="width:100%;object-fit:cover;max-height:20rem" />
      </div>

      <div class="a-page-md" style="padding-top:3rem">
        <!-- Breadcrumb -->
        <RouterLink to="/blog/explore" class="a-link">← 博客广场</RouterLink>

        <!-- Title -->
        <h1 class="a-title" style="margin-top:1.5rem;margin-bottom:1rem">{{ post.title }}</h1>

        <!-- Meta -->
        <div style="display:flex;flex-wrap:wrap;align-items:center;gap:1rem;padding-bottom:1.5rem;border-bottom:2px solid #000;margin-bottom:2.5rem">
          <RouterLink :to="`/blog/@${post.user?.username}`" style="display:flex;align-items:center;gap:.5rem;text-decoration:none">
            <div style="width:2rem;height:2rem;border-radius:9999px;background:#000;display:flex;align-items:center;justify-content:center;color:#fff;font-weight:900;font-size:.75rem">
              {{ (post.user?.display_name || post.user?.username || '?').charAt(0).toUpperCase() }}
            </div>
            <span style="font-weight:900;font-size:.875rem">{{ post.user?.display_name || post.user?.username }}</span>
          </RouterLink>
          <span class="a-label a-muted">{{ formatDate(post.created_at) }}</span>
          <RouterLink v-if="isOwner" :to="`/blog/posts/${post.id}/edit`" class="a-btn-outline-sm" style="margin-left:auto">编辑</RouterLink>
        </div>

        <!-- Markdown content -->
        <div class="prose-blog" style="margin-bottom:3rem" v-html="renderedContent" />

        <!-- Interaction bar -->
        <div style="display:flex;align-items:center;gap:1rem;padding:1.5rem 0;border-top:2px solid #000;border-bottom:2px solid #000;margin-bottom:3rem">
          <button
            @click="toggleLike"
            style="display:flex;align-items:center;gap:.5rem;font-weight:900;font-size:.875rem;border:2px solid #000;padding:.5rem 1rem;cursor:pointer;transition:all .2s"
            :style="liked ? 'background:#000;color:#fff' : 'background:#fff;color:#000'"
          >
            ♥ {{ likesCount }}
          </button>
          <button
            v-if="authStore.isAuthenticated"
            @click="toggleBookmark"
            style="display:flex;align-items:center;gap:.5rem;font-weight:900;font-size:.875rem;border:2px solid #000;padding:.5rem 1rem;cursor:pointer;transition:all .2s"
            :style="bookmarked ? 'background:#000;color:#fff' : 'background:#fff;color:#000'"
          >
            {{ bookmarked ? '★ 已收藏' : '☆ 收藏' }}
          </button>
          <a
            v-if="post.user?.username"
            :href="api.feed.rss(post.user.username)"
            target="_blank"
            class="a-link"
            style="margin-left:auto"
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

    <AConfirm
      :show="showUnbookmarkConfirm"
      title="取消收藏"
      message="确定将这篇文章移出收藏吗？"
      confirm-text="删除"
      cancel-text="取消"
      danger
      @confirm="confirmToggleBookmark"
      @cancel="showUnbookmarkConfirm = false"
    />
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import { RouterLink, useRoute } from 'vue-router'
import CommentSection from '@/components/blog/CommentSection.vue'
import AConfirm from '@/components/ui/AConfirm.vue'
import { useAuthStore } from '@/stores/auth'
import { useApi } from '@/composables/useApi'
import { useMarkdownRenderer } from '@/composables/useMarkdownRenderer'
import type { Post } from '@/types'

const route = useRoute()
const authStore = useAuthStore()
const api = useApi()
const { renderMarkdown } = useMarkdownRenderer()

const post = ref<Post | null>(null)
const loading = ref(true)
const errorStatus = ref<number | null>(null)
const liked = ref(false)
const likesCount = ref(0)
const bookmarked = ref(false)
const showUnbookmarkConfirm = ref(false)

const isOwner = computed(() => authStore.user?.id === post.value?.user_id)

const formatDate = (d: string) => new Date(d).toLocaleDateString('zh-CN', { year: 'numeric', month: 'long', day: 'numeric' })

const renderedContent = computed(() => renderMarkdown(post.value?.content ?? ''))

const fetchPost = async () => {
  loading.value = true
  errorStatus.value = null
  try {
    const id = String(route.params.id || '')
    if (!id) {
      errorStatus.value = 404
      return
    }
    const headers: Record<string, string> = {}
    if (authStore.token) headers['Authorization'] = `Bearer ${authStore.token}`
    const res = await fetch(api.blog.post(id), { headers })
    if (res.ok) {
      const d = await res.json()
      post.value = d.data || d
      likesCount.value = post.value?.likes_count ?? 0
    } else {
      errorStatus.value = res.status
    }
  } catch (e) {
    console.error(e)
    errorStatus.value = 500
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
  if (bookmarked.value) {
    showUnbookmarkConfirm.value = true
    return
  }
  await runToggleBookmark()
}

const confirmToggleBookmark = async () => {
  showUnbookmarkConfirm.value = false
  await runToggleBookmark()
}

const runToggleBookmark = async () => {
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

/* KaTeX math rendering */
.prose-blog :deep(.katex-display) { margin: 1.5rem 0; overflow-x: auto; }
.prose-blog :deep(.katex) { font-size: 1.05rem; }

/* highlight.js code theme (inside dark pre) */
.prose-blog :deep(.hljs-keyword),
.prose-blog :deep(.hljs-built_in) { color: #ff79c6; }
.prose-blog :deep(.hljs-string) { color: #f1fa8c; }
.prose-blog :deep(.hljs-number) { color: #bd93f9; }
.prose-blog :deep(.hljs-comment) { color: #6272a4; font-style: italic; }
.prose-blog :deep(.hljs-function),
.prose-blog :deep(.hljs-title) { color: #50fa7b; }
.prose-blog :deep(.hljs-variable),
.prose-blog :deep(.hljs-attr) { color: #8be9fd; }
</style>
