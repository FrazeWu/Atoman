<template>
  <div class="a-page">
    <!-- Logged-out hero -->
    <div
      v-if="!authStore.isAuthenticated"
      style="min-height:60vh;display:flex;flex-direction:column;align-items:center;justify-content:center;text-align:center"
    >
      <h1 class="a-title-lg" style="margin-bottom:1.5rem">ATOMAN<br />BLOG</h1>
      <p class="a-muted" style="max-width:28rem;margin-bottom:2rem">创作你的博客，订阅他人内容，构建属于你的知识图谱。</p>
      <div style="display:flex;gap:1rem">
        <ABtn to="/login" size="lg">登录</ABtn>
        <ABtn outline to="/blog/explore" size="lg">浏览文章</ABtn>
      </div>
    </div>

    <!-- Logged-in -->
    <template v-else>
      <APageHeader title="我的博客" accent>
        <template #action>
          <ABtn to="/blog/posts/new">+ 写文章</ABtn>
        </template>
      </APageHeader>

      <!-- Channels -->
      <section style="margin-bottom:3rem">
        <div style="display:flex;align-items:center;justify-content:space-between;margin-bottom:1rem">
          <h2 class="a-subtitle">我的频道</h2>
          <ABtn size="sm" outline @click="showCreateChannel = true">+ 新建频道</ABtn>
        </div>
        <AEmpty v-if="!channels.length" text="还没有频道，创建一个来组织你的文章" />
        <div v-else class="a-grid-3">
          <RouterLink
            v-for="ch in channels"
            :key="ch.id"
            :to="`/blog/explore?channel=${ch.id}`"
            class="a-card a-card-hover"
            style="display:flex;flex-direction:column;gap:.25rem;text-decoration:none;color:#000"
          >
            <span style="font-weight:900;font-size:1.125rem;letter-spacing:-0.025em">{{ ch.name }}</span>
            <span v-if="ch.description" class="a-muted a-clamp-2" style="font-size:.875rem">{{ ch.description }}</span>
          </RouterLink>
        </div>
      </section>

      <!-- Recent posts -->
      <section style="margin-bottom:3rem">
        <div style="display:flex;align-items:center;justify-content:space-between;margin-bottom:1rem">
          <h2 class="a-subtitle">最近文章</h2>
          <RouterLink to="/blog/explore" class="a-link">发现广场 →</RouterLink>
        </div>
        <div v-if="loadingPosts" class="a-grid-2">
          <div v-for="i in 4" :key="i" class="a-skeleton" />
        </div>
        <AEmpty v-else-if="!recentPosts.length" text="你还没有发布任何文章" />
        <div v-else class="a-grid-2">
          <PostCard v-for="post in recentPosts" :key="post.id" :post="post" />
        </div>
      </section>
    </template>

    <!-- Create Channel Modal -->
    <AModal v-if="showCreateChannel" @close="showCreateChannel = false" size="sm">
      <h3 class="a-subtitle" style="margin-bottom:1.5rem">创建频道</h3>
      <div style="display:flex;flex-direction:column;gap:1rem">
        <input v-model="newChannelName" placeholder="频道名称*" class="a-input" />
        <textarea v-model="newChannelDesc" placeholder="频道描述（可选）" rows="3" class="a-textarea" />
      </div>
      <div style="display:flex;gap:.75rem;margin-top:1.5rem">
        <ABtn style="flex:1" @click="createChannel">创建</ABtn>
        <ABtn outline @click="showCreateChannel = false">取消</ABtn>
      </div>
    </AModal>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { RouterLink } from 'vue-router'
import PostCard from '@/components/blog/PostCard.vue'
import ABtn from '@/components/ui/ABtn.vue'
import AModal from '@/components/ui/AModal.vue'
import AEmpty from '@/components/ui/AEmpty.vue'
import APageHeader from '@/components/ui/APageHeader.vue'
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

<style scoped>
</style>
