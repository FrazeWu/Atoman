<template>
  <div class="a-page">
    <APageHeader title="发现" accent>
      <template #action>
        <ABtn v-if="authStore.isAuthenticated" to="/post/new">+ 写文章</ABtn>
        <ABtn v-else to="/login" outline>登录</ABtn>
      </template>
    </APageHeader>

    <!-- Filters -->
    <div style="display:flex;align-items:center;gap:1rem;margin-bottom:1.5rem;flex-wrap:wrap">
      <div style="display:flex;gap:.5rem">
        <button
          v-for="t in typeOptions"
          :key="t.value"
          class="filter-btn"
          :class="{ active: typeFilter === t.value }"
          @click="typeFilter = t.value; fetchPosts()"
        >{{ t.label }}</button>
      </div>
      <div style="margin-left:auto;display:flex;gap:.5rem">
        <button
          v-for="s in sortOptions"
          :key="s.value"
          class="filter-btn"
          :class="{ active: sortBy === s.value }"
          @click="sortBy = s.value; fetchPosts()"
        >{{ s.label }}</button>
      </div>
    </div>

    <!-- Posts grid -->
    <div v-if="loading" class="a-grid-2">
      <div v-for="i in 6" :key="i" class="a-skeleton" style="height:12rem" />
    </div>
    <AEmpty v-else-if="!posts.length" title="暂无内容" description="还没有发布任何内容" />
    <div v-else class="a-grid-2">
      <PostCard v-for="post in posts" :key="post.id" :post="post" show-channel />
    </div>

    <!-- Load more -->
    <div v-if="hasMore && !loading" style="display:flex;justify-content:center;margin-top:2rem">
      <ABtn outline @click="loadMore">加载更多</ABtn>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import PostCard from '@/components/blog/PostCard.vue'
import ABtn from '@/components/ui/ABtn.vue'
import AEmpty from '@/components/ui/AEmpty.vue'
import APageHeader from '@/components/ui/APageHeader.vue'
import { useAuthStore } from '@/stores/auth'
import { useApi } from '@/composables/useApi'
import type { Post } from '@/types'

const authStore = useAuthStore()
const api = useApi()

const posts = ref<Post[]>([])
const loading = ref(true)
const page = ref(1)
const hasMore = ref(false)
const typeFilter = ref('all')
const sortBy = ref('latest')

const typeOptions = [
  { label: '全部', value: 'all' },
  { label: '文章', value: 'post' },
]

const sortOptions = [
  { label: '最新', value: 'latest' },
  { label: '最热', value: 'popular' },
]

const fetchPosts = async (append = false) => {
  loading.value = true
  if (!append) page.value = 1
  try {
    const params = new URLSearchParams({
      page: String(page.value),
      limit: '12',
      status: 'published',
    })
    if (sortBy.value === 'popular') params.set('sort', 'popular')

    const headers: Record<string, string> = {}
    if (authStore.token) headers['Authorization'] = `Bearer ${authStore.token}`

    const res = await fetch(`${api.blog.posts}?${params}`, { headers })
    if (res.ok) {
      const d = await res.json()
      const data: Post[] = d.data || []
      if (append) {
        posts.value = [...posts.value, ...data]
      } else {
        posts.value = data
      }
      hasMore.value = data.length === 12
    }
  } catch (e) {
    console.error(e)
  } finally {
    loading.value = false
  }
}

const loadMore = () => {
  page.value++
  fetchPosts(true)
}

onMounted(() => fetchPosts())
</script>

<style scoped>
.filter-btn {
  padding: .375rem .75rem;
  font-size: .8125rem;
  font-weight: 700;
  text-transform: uppercase;
  letter-spacing: .05em;
  border: 2px solid #000;
  background: #fff;
  cursor: pointer;
  transition: background .1s, color .1s;
}
.filter-btn:hover,
.filter-btn.active {
  background: #000;
  color: #fff;
}
</style>
