<template>
  <div class="a-page" style="padding-bottom:12rem">
    <AToast v-model="toastVisible" :message="toastMessage" />
    <!-- Loading -->
    <div v-if="loading" style="display:flex;flex-direction:column;gap:1.5rem">
      <div class="a-skeleton" style="height:8rem" />
      <div class="a-skeleton" style="height:2rem;width:50%" />
    </div>

    <!-- Not found -->
    <div v-else-if="!profile" style="text-align:center;padding:6rem 0">
      <p class="a-title a-muted" style="margin-bottom:1rem">用户不存在</p>
      <RouterLink to="/blog" class="a-link">← 博客首页</RouterLink>
    </div>

    <template v-else>
      <!-- Profile header -->
      <div class="a-card" style="display:flex;flex-wrap:wrap;gap:1.5rem;align-items:flex-start;margin-bottom:2rem">
        <!-- Avatar -->
        <div style="width:5rem;height:5rem;border-radius:9999px;background:#000;display:flex;align-items:center;justify-content:center;color:#fff;font-size:2rem;font-weight:900;flex-shrink:0">
          {{ (profile.display_name || profile.username).charAt(0).toUpperCase() }}
        </div>

        <div style="flex:1">
          <div style="display:flex;align-items:flex-start;justify-content:space-between;gap:1rem;flex-wrap:wrap;margin-bottom:1rem">
            <div>
              <h1 style="font-size:1.875rem;font-weight:900;letter-spacing:-0.025em">{{ profile.display_name || profile.username }}</h1>
              <p class="a-muted" style="font-size:.875rem">@{{ profile.username }}</p>
            </div>
            <div style="display:flex;gap:.5rem">
              <button
                v-if="authStore.isAuthenticated && !isSelf"
                @click="toggleFollow"
                class="a-toggle-btn"
                :class="{ 'a-toggle-btn-active': following }"
              >
                {{ following ? '已订阅' : '订阅' }}
              </button>
              <!-- 仅外部RSS复制入口 -->
              <button
                v-if="rssUrl"
                @click="copyRssLink"
                class="a-btn-outline-sm"
              >复制RSS链接</button>
              <button
                v-if="authStore.isAuthenticated && !isSelf && userChannelId"
                @click="toggleChannelSubscribe"
                class="a-toggle-btn"
                :class="{ 'a-toggle-btn-active': channelSubscribed }"
                :disabled="channelSubscribeLoading"
              >
                {{ channelSubscribeLoading ? '加载中...' : (channelSubscribed ? '已订阅频道' : '订阅频道') }}
              </button>
              <RouterLink v-if="isSelf" to="/blog/settings" class="a-btn-outline-sm">编辑资料</RouterLink>
            </div>
          </div>

          <!-- Stats -->
          <div style="display:flex;gap:1.5rem;font-weight:900;font-size:.875rem;margin-bottom:.75rem">
            <span><span style="font-size:1.25rem">{{ profile.posts_count ?? posts.length }}</span> 篇文章</span>
            <span><span style="font-size:1.25rem">{{ profile.followers_count ?? 0 }}</span> 订阅者</span>
            <span><span style="font-size:1.25rem">{{ profile.following_count ?? 0 }}</span> 已订阅</span>
          </div>

          <p v-if="profile.bio" class="a-muted">{{ profile.bio }}</p>
          <a v-if="profile.website" :href="profile.website" target="_blank" class="a-link" style="font-size:.875rem">{{ profile.website }}</a>
        </div>
      </div>

      <!-- Channel info -->
      <div v-if="userChannel" class="a-card" style="margin-bottom:2rem">
        <div style="display:flex;align-items:center;justify-content:space-between;margin-bottom:1rem">
          <div>
            <h3 style="font-weight:900;font-size:1.25rem;margin-bottom:.25rem">{{ userChannel.name }}</h3>
            <p v-if="userChannel.description" class="a-muted" style="font-size:.875rem">{{ userChannel.description }}</p>
          </div>
          <RouterLink :to="`/blog/channel/${userChannel.id}`" class="a-btn-outline-sm">查看频道</RouterLink>
        </div>
      </div>

      <!-- Posts -->
      <div style="display:flex;align-items:center;justify-content:space-between;margin-bottom:1.5rem">
        <h2 class="a-subtitle">文章</h2>
      </div>
      <div v-if="loadingPosts" class="a-grid-2">
        <div v-for="i in 4" :key="i" class="a-skeleton" style="height:10rem" />
      </div>
      <AEmpty v-else-if="!posts.length" text="还没有发布文章" />
      <div v-else class="a-grid-2">
        <PostCard v-for="post in posts" :key="post.id" :post="post" />
      </div>
    </template>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import { RouterLink, useRoute } from 'vue-router'
import PostCard from '@/components/blog/PostCard.vue'
import AEmpty from '@/components/ui/AEmpty.vue'
import { useAuthStore } from '@/stores/auth'
import { useFeedStore } from '@/stores/feed'
import AToast from '@/components/ui/AToast.vue'
import { useApi } from '@/composables/useApi'
import type { UserProfile, Post, Channel } from '@/types'

const route = useRoute()
const authStore = useAuthStore()
const feedStore = useFeedStore()
const api = useApi()

const profile = ref<UserProfile | null>(null)
const posts = ref<Post[]>([])
const loading = ref(true)
const loadingPosts = ref(true)
const following = ref(false)

// Channel subscription state
const userChannel = ref<Channel | null>(null)
const channelSubscribed = ref(false)
const channelSubscribeLoading = ref(false)
// 移除RSS订阅状态
const toastVisible = ref(false)
const toastMessage = ref('')

const username = computed(() => route.params.username as string)
const isSelf = computed(() => authStore.user?.username === username.value)
const userChannelId = computed(() => userChannel.value?.id || '')
const rssUrl = computed(() => {
  if (!profile.value?.username) return ''
  return api.feed.rss(profile.value.username)
})

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
    const res = await fetch(`${api.blog.posts}?user_id=${profile.value.uuid}`)
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

const fetchFollowingState = async () => {
  if (!profile.value || !authStore.isAuthenticated || isSelf.value) return
  try {
    const res = await fetch(api.users.following(authStore.user?.uuid || ''), {
      headers: { Authorization: `Bearer ${authStore.token}` },
    })
    if (res.ok) {
      const d = await res.json()
      const followingUsers = d.data || []
      following.value = followingUsers.some((user: { uuid?: string }) => user.uuid === profile.value?.uuid)
    }
  } catch (e) {
    console.error('Failed to fetch following state:', e)
  }
}

const fetchUserChannel = async () => {
  if (!profile.value) return
  try {
    const res = await fetch(`${api.blog.channels}?user_id=${profile.value.uuid}`)
    if (res.ok) {
      const d = await res.json()
      const channels = d.data || []
      if (channels.length > 0) {
        userChannel.value = channels[0]
        if (authStore.isAuthenticated && userChannel.value) {
          channelSubscribed.value = await feedStore.isSubscribedToChannel(userChannel.value.id)
        }
      }
    }
  } catch (e) {
    console.error('Failed to fetch user channel:', e)
  }
}

const toggleFollow = async () => {
  if (!profile.value) return
  const method = following.value ? 'DELETE' : 'POST'
  try {
    const res = await fetch(api.users.follow(profile.value.uuid), {
      method,
      headers: { Authorization: `Bearer ${authStore.token}` }
    })
    if (res.ok) following.value = !following.value
  } catch (e) {
    console.error(e)
  }
}

const toggleChannelSubscribe = async () => {
  if (!userChannel.value) return
  channelSubscribeLoading.value = true
  try {
    let success = false
    if (channelSubscribed.value) {
      success = await feedStore.unsubscribeFromChannel(userChannel.value.id)
    } else {
      success = await feedStore.subscribeToChannel(userChannel.value.id)
    }
    if (success) {
      channelSubscribed.value = !channelSubscribed.value
    }
  } catch (e) {
    console.error('Failed to toggle channel subscription:', e)
  } finally {
    channelSubscribeLoading.value = false
  }
}



const copyRssLink = async () => {
  if (!rssUrl.value) return
  try {
    await navigator.clipboard.writeText(rssUrl.value)
    toastMessage.value = '已复制 RSS 链接'
    toastVisible.value = true
  } catch (e) {
    console.error('Failed to copy RSS link:', e)
  }
}

onMounted(async () => {
  await fetchProfile()
  await fetchFollowingState()
  await fetchUserChannel()
  await fetchPosts()
})
</script>
