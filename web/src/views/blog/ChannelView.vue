<template>
  <div class="a-page-xl">
    <AToast v-model="toastVisible" :message="toastMessage" />
    <div v-if="loading" class="a-grid-2" style="margin-top:1rem">
      <div v-for="item in 4" :key="item" class="a-skeleton" style="height:10rem" />
    </div>

    <AEmpty v-else-if="!channel" text="合集不存在或已被删除" />

    <template v-else>
      <APageHeader :title="channel.name" accent :sub="channel.description || '合集主页'">
        <template #action>
          <div style="display:flex;gap:.75rem;flex-wrap:wrap">
            <button
              v-if="authStore.isAuthenticated && !isOwner"
              @click="toggleChannelSubscribe"
              class="a-toggle-btn"
              :class="{ 'a-toggle-btn-active': channelSubscribed }"
              :disabled="channelSubscribeLoading"
            >
              {{ channelSubscribeLoading ? '加载中...' : (channelSubscribed ? '已订阅' : '订阅') }}
            </button>
            <!-- 仅外部RSS复制入口 -->
            <button
              v-if="channelRssUrl"
              @click="copyRssLink"
              class="a-btn-outline-sm"
            >复制RSS链接</button>
            <RouterLink to="/blog" class="a-btn-outline-sm">返回管理台</RouterLink>
            <RouterLink
              v-if="isOwner"
              :to="{ path: '/post/new', query: { channel: channel.id } }"
              class="a-btn"
            >写文章</RouterLink>
          </div>
        </template>
      </APageHeader>

      <div class="channel-meta-card">
        <div>
          <p class="a-label a-muted" style="margin-bottom:.4rem">合集作者</p>
          <p style="font-weight:900;font-size:1rem;margin:0">{{ channel.user?.display_name || channel.user?.username || '未知作者' }}</p>
        </div>
        <div>
          <p class="a-label a-muted" style="margin-bottom:.4rem">更新时间</p>
          <p style="font-weight:700;margin:0">{{ formatDate(channel.updated_at) }}</p>
        </div>
      </div>

      <section style="margin:2rem 0 2.5rem">
        <div class="section-headline">
          <h2 class="a-subtitle" style="margin:0">合集</h2>
          <span class="a-muted" style="font-size:.875rem">{{ collections.length }} 个</span>
        </div>

        <AEmpty v-if="!collections.length" text="当前合集暂无子合集" />
        <div v-else class="a-grid-3">
          <div v-for="collection in collections" :key="collection.id" class="a-card a-card-hover" style="position:relative">
            <div style="cursor:pointer" @click="$router.push(`/collection/${collection.id}`)">
              <div style="display:flex;align-items:center;justify-content:space-between;margin-bottom:.6rem">
                <span class="a-badge" :class="collection.is_default ? 'a-badge-fill' : ''">
                  {{ collection.is_default ? '默认合集' : '合集' }}
                </span>
                <span class="a-label a-muted">{{ postCountByCollection(collection.id) }}篇</span>
              </div>
              <h3 style="margin:0 0 .4rem;font-weight:900;font-size:1.05rem">{{ collection.name }}</h3>
              <p class="a-muted a-clamp-3" style="font-size:.875rem">{{ collection.description || '暂无合集描述' }}</p>
            </div>
            <div v-if="isOwner" style="display:flex;gap:.5rem;margin-top:1rem;padding-top:1rem;border-top:1px solid #f0f0f0">
              <button @click.stop="openCollectionModal(collection)" class="a-btn-outline-sm">编辑</button>
              <button v-if="!collection.is_default" @click.stop="showDeleteModal(collection)" class="a-btn-outline-sm" style="color:#ef4444;border-color:#ef4444">删除</button>
            </div>
          </div>
        </div>
      </section>

      <section>
        <div class="section-headline">
          <h2 class="a-subtitle" style="margin:0">文章</h2>
          <span class="a-muted" style="font-size:.875rem">{{ channelPosts.length }} 篇</span>
        </div>

        <AEmpty v-if="!channelPosts.length" text="当前合集还没有文章" />
        <div v-else class="a-grid-2">
          <div v-for="post in channelPosts" :key="post.id" class="a-card a-card-hover">
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

            <div style="display:flex;gap:.5rem;flex-wrap:wrap;margin-top:1rem">
              <span v-for="collection in post.collections || []" :key="collection.id" class="a-badge">
                {{ collection.name }}
              </span>
            </div>

            <div style="display:flex;gap:.75rem;margin-top:1rem;flex-wrap:wrap">
              <RouterLink :to="`/post/${post.id}`" class="a-btn-outline-sm">查看</RouterLink>
              <RouterLink
                v-if="isOwner"
                :to="{ path: `/post/${post.id}/edit`, query: { channel: channel.id } }"
                class="a-btn-outline-sm"
              >编辑</RouterLink>
            </div>
          </div>
        </div>
      </section>
    </template>

    <!-- Collection Modal -->
    <AModal v-model="collectionModalOpen" title="编辑合集">
      <div style="display:flex;flex-direction:column;gap:1rem">
        <AInput v-model="collectionForm.name" label="合集名称" placeholder="输入合集名称" />
        <ATextarea v-model="collectionForm.description" label="合集描述" placeholder="简短介绍这个合集" :rows="3" />
        <div style="display:flex;gap:.75rem;justify-content:flex-end;margin-top:.5rem">
          <ABtn outline @click="collectionModalOpen = false">取消</ABtn>
          <ABtn :disabled="!collectionForm.name.trim() || collectionSaving" @click="saveCollection">
            {{ collectionSaving ? '保存中...' : (editingCollection ? '更新' : '创建') }}
          </ABtn>
        </div>
      </div>
    </AModal>

    <!-- Delete Confirmation Modal -->
    <AModal v-if="deleteModalVisible" @close="closeDeleteModal">
      <div style="display:flex;flex-direction:column;gap:1rem">
        <p>确定要删除合集<strong>{{ collectionToDelete?.name }}</strong>吗？此操作不可恢复。</p>
        <div style="display:flex;gap:.75rem;justify-content:flex-end">
          <ABtn outline @click="closeDeleteModal">取消</ABtn>
          <ABtn danger @click="executeDelete">删除</ABtn>
        </div>
      </div>
    </AModal>
  </div>
</template>

<script setup lang="ts">
import { computed, onMounted, ref } from 'vue'
import { RouterLink, useRoute } from 'vue-router'
import AEmpty from '@/components/ui/AEmpty.vue'
import APageHeader from '@/components/ui/APageHeader.vue'
import AModal from '@/components/ui/AModal.vue'
import ABtn from '@/components/ui/ABtn.vue'
import AInput from '@/components/ui/AInput.vue'
import ATextarea from '@/components/ui/ATextarea.vue'
import type { Channel, Collection, Post } from '@/types'
import { useApi } from '@/composables/useApi'
import { useAuthStore } from '@/stores/auth'
import { useFeedStore } from '@/stores/feed'
import AToast from '@/components/ui/AToast.vue'

const route = useRoute()
const api = useApi()
const authStore = useAuthStore()
const feedStore = useFeedStore()

const loading = ref(true)
const channel = ref<Channel | null>(null)
const collections = ref<Collection[]>([])
const publishedPosts = ref<Post[]>([])
const draftPosts = ref<Post[]>([])

// Modal state
const collectionModalOpen = ref(false)
const editingCollection = ref<Collection | null>(null)
const collectionForm = ref({ name: '', description: '' })
const collectionSaving = ref(false)

// Delete modal state - using separate boolean flag
const deleteModalVisible = ref(false)
const collectionToDelete = ref<Collection | null>(null)

// Channel subscription state
const channelSubscribed = ref(false)
const channelSubscribeLoading = ref(false)
// 移除RSS订阅状态
const toastVisible = ref(false)
const toastMessage = ref('')

const channelId = computed(() => (typeof route.params.id === 'string' ? route.params.id : ''))
const authHeader = computed(() => ({ Authorization: `Bearer ${authStore.token}` }))
const isOwner = computed(() => !!channel.value && channel.value.user_id === authStore.user?.uuid)
const channelRssUrl = computed(() => {
  const username = channel.value?.user?.username
  if (!username) return ''
  return api.feed.rss(username)
})

const channelPosts = computed(() => {
  const merged = [...publishedPosts.value, ...draftPosts.value]
  const sorted = merged.sort(
    (left, right) => new Date(right.updated_at).getTime() - new Date(left.updated_at).getTime()
  )

  const seen = new Set<string>()
  return sorted.filter((post) => {
    if (seen.has(post.id)) return false
    seen.add(post.id)
    return true
  })
})

const formatDate = (date: string) => new Date(date).toLocaleDateString('zh-CN')

const summarize = (content: string) => {
  const text = content
    .replace(/```[\s\S]*?```/g, ' ')
    .replace(/^#+\s*/gm, '')
    .replace(/[>*_`\[\]()!-]/g, ' ')
    .replace(/\s+/g, ' ')
    .trim()
  return text.slice(0, 120) || '暂无摘要'
}

const postCountByCollection = (collectionId: string) =>
  channelPosts.value.filter((post) => (post.collections || []).some((c) => c.id === collectionId)).length

const fetchChannel = async () => {
  const res = await fetch(api.blog.channel(channelId.value))
  if (!res.ok) {
    channel.value = null
    return
  }

  const data = await res.json()
  channel.value = (data.data || null) as Channel | null
  
  // Check subscription status if authenticated
  if (authStore.isAuthenticated && channel.value) {
    channelSubscribeLoading.value = true
    channelSubscribed.value = await feedStore.isSubscribedToChannel(channel.value.id)
    channelSubscribeLoading.value = false

    // 不再检测RSS订阅状态
  }
}

const fetchCollections = async () => {
  const res = await fetch(api.blog.channelCollections(channelId.value))
  if (!res.ok) {
    collections.value = []
    return
  }

  const data = await res.json()
  collections.value = (data.data || []) as Collection[]
}

const fetchPublishedPosts = async () => {
  if (!channel.value) {
    publishedPosts.value = []
    return
  }

  const byCollection = await Promise.all(
    collections.value.map(async (collection) => {
      const res = await fetch(`${api.blog.posts}?collection_id=${collection.id}`)
      if (!res.ok) return [] as Post[]
      const data = await res.json()
      return (data.data || []) as Post[]
    })
  )

  publishedPosts.value = byCollection.flat()
}

const fetchDraftPostsIfOwner = async () => {
  if (!isOwner.value) {
    draftPosts.value = []
    return
  }

  const res = await fetch(api.blog.drafts, { headers: authHeader.value })
  if (!res.ok) {
    draftPosts.value = []
    return
  }

  const data = await res.json()
  const drafts = (data.data || []) as Post[]
  draftPosts.value = drafts.filter((post) =>
    (post.collections || []).some((collection) => collection.channel_id === channelId.value)
  )
}

// Collection modal methods
const openCollectionModal = (collection?: Collection) => {
  editingCollection.value = collection || null
  collectionForm.value = {
    name: collection?.name || '',
    description: collection?.description || ''
  }
  collectionModalOpen.value = true
}

const saveCollection = async () => {
  if (!collectionForm.value.name.trim()) return
  
  collectionSaving.value = true
  try {
    if (editingCollection.value) {
      // Update existing
      await fetch(api.blog.collection(editingCollection.value.id), {
        method: 'PUT',
        headers: { ...authHeader.value, 'Content-Type': 'application/json' },
        body: JSON.stringify(collectionForm.value)
      })
    } else {
      // Create new
      await fetch(api.blog.channelCollections(channelId.value), {
        method: 'POST',
        headers: { ...authHeader.value, 'Content-Type': 'application/json' },
        body: JSON.stringify(collectionForm.value)
      })
    }
    collectionModalOpen.value = false
    await fetchCollections()
    await fetchPublishedPosts()
  } catch (e) {
    console.error('Failed to save collection:', e)
  } finally {
    collectionSaving.value = false
  }
}

// Delete modal methods - explicitly controlled
const showDeleteModal = (collection: Collection) => {
  collectionToDelete.value = collection
  deleteModalVisible.value = true
}

const closeDeleteModal = () => {
  deleteModalVisible.value = false
  collectionToDelete.value = null
}

const executeDelete = async () => {
  if (!collectionToDelete.value) return
  
  try {
    await fetch(api.blog.collection(collectionToDelete.value.id), {
      method: 'DELETE',
      headers: authHeader.value
    })
    closeDeleteModal()
    await fetchCollections()
    await fetchPublishedPosts()
  } catch (e) {
    console.error('Failed to delete collection:', e)
  }
}

const toggleChannelSubscribe = async () => {
  if (!channel.value) return
  channelSubscribeLoading.value = true
  try {
    let success = false
    if (channelSubscribed.value) {
      success = await feedStore.unsubscribeFromChannel(channel.value.id)
    } else {
      success = await feedStore.subscribeToChannel(channel.value.id)
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
  if (!channelRssUrl.value) return
  try {
    await navigator.clipboard.writeText(channelRssUrl.value)
    toastMessage.value = '已复制 RSS 链接'
    toastVisible.value = true
  } catch (e) {
    console.error('Failed to copy RSS link:', e)
  }
}

onMounted(async () => {
  if (!channelId.value) {
    loading.value = false
    return
  }

  try {
    await fetchChannel()
    if (!channel.value) return

    await fetchCollections()
    await fetchPublishedPosts()
    await fetchDraftPostsIfOwner()
  } catch (e) {
    console.error(e)
  } finally {
    loading.value = false
  }
})
</script>

<style scoped>
.section-headline {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 0.75rem;
  margin-bottom: 1rem;
  flex-wrap: wrap;
}

.channel-meta-card {
  border: 2px solid #000;
  background: #fff;
  padding: 1rem 1.25rem;
  display: flex;
  gap: 1.5rem;
  align-items: flex-start;
  justify-content: space-between;
  flex-wrap: wrap;
}

@media (max-width: 640px) {
  .channel-meta-card {
    flex-direction: column;
  }
}
</style>
