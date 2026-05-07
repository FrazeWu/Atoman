<template>
  <div class="a-page" style="padding-bottom:12rem">
    <div v-if="loading" style="display:flex;flex-direction:column;gap:1.5rem">
      <div class="a-skeleton" style="height:8rem" />
      <div class="a-skeleton" style="height:2rem;width:50%" />
    </div>

    <AEmpty v-else-if="!collection" text="合集不存在或已被删除" />

    <template v-else>
      <APageHeader :title="collection.name" accent :sub="collection.description || '合集详情'" style="margin-bottom:2.5rem">
        <template #action>
          <div style="display:flex;gap:.75rem;flex-wrap:wrap">
            <button
              v-if="authStore.isAuthenticated && !isOwner"
              @click="toggleCollectionSubscribe"
              class="a-toggle-btn"
              :class="{ 'a-toggle-btn-active': collectionSubscribed }"
              :disabled="collectionSubscribeLoading"
            >
              {{ collectionSubscribeLoading ? '加载中...' : (collectionSubscribed ? '已订阅' : '订阅合集') }}
            </button>
            <RouterLink :to="`/channel/${channelId}`" class="a-btn-outline-sm">返回频道</RouterLink>
            <RouterLink
              v-if="isOwner"
              :to="{ path: '/post/new', query: { channel: channelId, collection: collection.id } }"
              class="a-btn"
            >写文章</RouterLink>
          </div>
        </template>
      </APageHeader>

      <div class="a-card" style="margin-bottom:2.5rem">
        <div style="display:flex;align-items:center;justify-content:space-between;margin-bottom:1rem">
          <div>
            <p class="a-label a-muted" style="margin-bottom:.4rem">所属合集</p>
            <RouterLink :to="`/channel/${channelId}`" class="a-link" style="font-size:.875rem">
              {{ channel?.name || '加载中...' }}
            </RouterLink>
          </div>
          <div>
            <p class="a-label a-muted" style="margin-bottom:.4rem">文章数量</p>
            <p style="font-weight:900;margin:0">{{ posts.length }}篇</p>
          </div>
          <div v-if="isOwner" style="display:flex;gap:.5rem">
            <button @click="openEditModal" class="a-btn-outline-sm">编辑</button>
            <button @click="confirmDelete" class="a-btn-outline-sm" style="color:#ef4444;border-color:#ef4444">删除</button>
          </div>
        </div>
      </div>

      <section>
        <div class="section-headline">
          <h2 class="a-subtitle" style="margin:0">收录文章</h2>
          <span class="a-muted" style="font-size:.875rem">{{ posts.length }} 篇</span>
        </div>

        <AEmpty v-if="!posts.length" text="当前合集暂无文章" />
        <div v-else class="a-grid-2">
          <div v-for="post in posts" :key="post.id" class="a-card a-card-hover">
            <div style="display:flex;align-items:center;justify-content:space-between;margin-bottom:.6rem">
              <span class="a-badge" :class="post.status === 'published' ? 'a-badge-fill' : ''">
                {{ post.status === 'published' ? '已发布' : '草稿' }}
              </span>
              <span class="a-label a-muted">{{ formatDate(post.updated_at) }}</span>
            </div>

            <h3 class="a-clamp-2" style="font-weight:900;font-size:1.1rem;letter-spacing:-0.02em;margin-bottom:.5rem">
              {{ post.title }}
            </h3>
            <p class="a-muted a-clamp-3" style="font-size:.875rem">{{ post.summary || summarize(post.content) }}</p>

            <div style="display:flex;gap:.75rem;margin-top:1rem;flex-wrap:wrap">
              <RouterLink :to="`/post/${post.id}`" class="a-btn-outline-sm">查看</RouterLink>
              <RouterLink
                v-if="isOwner"
                :to="{ path: `/post/${post.id}/edit`, query: { channel: channelId } }"
                class="a-btn-outline-sm"
              >编辑</RouterLink>
            </div>
          </div>
        </div>
      </section>

      <!-- Edit Collection Modal -->
      <AModal v-model="editModalOpen" title="编辑合集">
        <div style="display:flex;flex-direction:column;gap:1rem">
          <AInput v-model="form.name" label="合集名称" placeholder="输入合集名称" />
          <ATextarea v-model="form.description" label="合集描述" placeholder="简短介绍这个合集" :rows="3" />
          <div style="display:flex;gap:.75rem;justify-content:flex-end;margin-top:.5rem">
            <ABtn outline @click="editModalOpen = false">取消</ABtn>
            <ABtn :disabled="!form.name.trim() || saving" @click="saveCollection">
              {{ saving ? '保存中...' : '更新' }}
            </ABtn>
          </div>
        </div>
      </AModal>

      <!-- Delete Confirmation Modal -->
      <AModal v-model="deleteModalOpen" title="确认删除合集">
        <div style="display:flex;flex-direction:column;gap:1rem">
          <p>确定要删除合集<strong>{{ collection.name }}</strong>吗？此操作不可恢复，但不会删除其中的文章。</p>
          <div style="display:flex;gap:.75rem;justify-content:flex-end">
            <ABtn outline @click="deleteModalOpen = false">取消</ABtn>
            <ABtn danger @click="deleteCollection">删除</ABtn>
          </div>
        </div>
      </AModal>
    </template>
  </div>
</template>

<script setup lang="ts">
import { computed, onMounted, ref } from 'vue'
import { RouterLink, useRoute, useRouter } from 'vue-router'
import AEmpty from '@/components/ui/AEmpty.vue'
import APageHeader from '@/components/ui/APageHeader.vue'
import AModal from '@/components/ui/AModal.vue'
import ABtn from '@/components/ui/ABtn.vue'
import AInput from '@/components/ui/AInput.vue'
import ATextarea from '@/components/ui/ATextarea.vue'
import type { Collection, Post, Channel } from '@/types'
import { useApi } from '@/composables/useApi'
import { useAuthStore } from '@/stores/auth'
import { useFeedStore } from '@/stores/feed'

const route = useRoute()
const router = useRouter()
const api = useApi()
const authStore = useAuthStore()
const feedStore = useFeedStore()

const loading = ref(true)
const collection = ref<Collection | null>(null)
const channel = ref<Channel | null>(null)
const posts = ref<Post[]>([])

const editModalOpen = ref(false)
const deleteModalOpen = ref(false)
const form = ref({ name: '', description: '' })
const saving = ref(false)
const collectionSubscribed = ref(false)
const collectionSubscribeLoading = ref(false)

const collectionId = computed(() => (typeof route.params.id === 'string' ? route.params.id : ''))
const channelId = computed(() => collection.value?.channel_id || '')
const authHeader = computed(() => ({ Authorization: `Bearer ${authStore.token}` }))
const isOwner = computed(() => {
  if (!collection.value) return false
  // Check ownership through channel since collections belong to channels
  return channel.value?.user_id === authStore.user?.uuid
})

const formatDate = (dateStr: string) => {
  const d = new Date(dateStr)
  const now = new Date()
  const diff = now.getTime() - d.getTime()
  const days = Math.floor(diff / (1000 * 60 * 60 * 24))
  
  if (days === 0) return '今天'
  if (days === 1) return '昨天'
  if (days < 7) return `${days}天前`
  
  const month = d.getMonth() + 1
  const day = d.getDate()
  const year = d.getFullYear()
  return year === now.getFullYear() ? `${month}/${day}` : `${year}/${month}/${day}`
}

const summarize = (content: string) => {
  const text = content.replace(/[#*_~`]/g, '').replace(/\n/g, ' ')
  return text.length > 120 ? text.slice(0, 120) + '...' : text
}

const fetchCollection = async () => {
  loading.value = true
  try {
    const res = await fetch(api.blog.collection(collectionId.value))
    if (res.ok) {
      const data = await res.json()
      collection.value = data.data
      if (collection.value?.channel_id) {
        await fetchChannel()
        await fetchPosts()
      }

      if (authStore.isAuthenticated && collection.value?.id) {
        collectionSubscribeLoading.value = true
        collectionSubscribed.value = await feedStore.isSubscribedToCollection(collection.value.id)
        collectionSubscribeLoading.value = false
      }
    }
  } catch (e) {
    console.error('Failed to fetch collection:', e)
  } finally {
    loading.value = false
  }
}

const fetchChannel = async () => {
  if (!channelId.value) return
  try {
    const res = await fetch(api.blog.channel(channelId.value))
    if (res.ok) {
      const data = await res.json()
      channel.value = data.data
    }
  } catch (e) {
    console.error('Failed to fetch channel:', e)
  }
}

const fetchPosts = async () => {
  if (!channelId.value) return
  try {
    const res = await fetch(`${api.blog.posts}?channel_id=${channelId.value}&limit=100`)
    if (res.ok) {
      const data = await res.json()
      const allPosts = (data.data || []) as Post[]
      posts.value = allPosts.filter(post => 
        (post.collections || []).some(c => c.id === collectionId.value)
      )
    }
  } catch (e) {
    console.error('Failed to fetch posts:', e)
  }
}

const openEditModal = () => {
  form.value = {
    name: collection.value?.name || '',
    description: collection.value?.description || ''
  }
  editModalOpen.value = true
}

const saveCollection = async () => {
  if (!form.value.name.trim() || !collection.value) return
  
  saving.value = true
  try {
    await fetch(api.blog.collection(collection.value.id), {
      method: 'PUT',
      headers: { ...authHeader.value, 'Content-Type': 'application/json' },
      body: JSON.stringify(form.value)
    })
    editModalOpen.value = false
    await fetchCollection()
  } catch (e) {
    console.error('Failed to save collection:', e)
  } finally {
    saving.value = false
  }
}

const confirmDelete = () => {
  deleteModalOpen.value = true
}

const deleteCollection = async () => {
  if (!collection.value) return
  
  try {
    await fetch(api.blog.collection(collection.value.id), {
      method: 'DELETE',
      headers: authHeader.value
    })
    deleteModalOpen.value = false
    router.push(`/channel/${channelId.value}`)
  } catch (e) {
    console.error('Failed to delete collection:', e)
  }
}

const toggleCollectionSubscribe = async () => {
  if (!collection.value) return
  collectionSubscribeLoading.value = true
  try {
    let success = false
    if (collectionSubscribed.value) {
      success = await feedStore.unsubscribeFromCollection(collection.value.id)
    } else {
      success = await feedStore.subscribeToCollection(collection.value.id)
    }

    if (success) {
      collectionSubscribed.value = !collectionSubscribed.value
    }
  } catch (e) {
    console.error('Failed to toggle collection subscription:', e)
  } finally {
    collectionSubscribeLoading.value = false
  }
}

onMounted(() => {
  fetchCollection()
})
</script>

<style scoped>
.section-headline {
  display: flex; align-items: baseline; gap: 0.75rem;
  margin-bottom: 1.5rem; padding-bottom: 0.75rem;
  border-bottom: 2px solid #000;
}
</style>
