<template>
  <div class="a-page-xl">
    <AToast v-model="toastVisible" :message="toastMessage" />
    <div v-if="loading" class="a-grid-2" style="margin-top:1rem">
      <div v-for="i in 4" :key="i" class="a-skeleton" style="height:10rem" />
    </div>

    <AEmpty v-else-if="!channel" title="频道不存在" description="该频道已被删除或链接无效" />

    <template v-else>
      <!-- Channel header -->
      <APageHeader :title="channel.name" accent :sub="channel.description">
        <template #action>
          <div style="display:flex;gap:.75rem;flex-wrap:wrap">
            <button
              v-if="authStore.isAuthenticated && !isOwner"
              @click="toggleChannelSubscribe"
              class="a-toggle-btn"
              :class="{ 'a-toggle-btn-active': channelSubscribed }"
              :disabled="channelSubscribeLoading"
            >
              {{ channelSubscribeLoading ? '...' : (channelSubscribed ? '已订阅' : '订阅') }}
            </button>
            <button v-if="channelRssUrl" @click="copyRssLink" class="a-btn-outline-sm">RSS</button>
            <RouterLink
              v-if="isOwner"
              :to="`/channel/${channel.slug || channel.id}/manage`"
              class="a-btn-outline-sm"
            >管理</RouterLink>
            <RouterLink
              v-if="isOwner"
              :to="{ path: '/post/new', query: { channel: channel.id } }"
              class="a-btn"
            >写文章</RouterLink>
          </div>
        </template>
      </APageHeader>

      <!-- Author info -->
      <div class="channel-meta-card">
        <div>
          <p class="a-label a-muted" style="margin-bottom:.4rem">作者</p>
          <RouterLink
            :to="`/user/${channel.user?.username}`"
            style="font-weight:900;font-size:1rem;text-decoration:none;color:#000"
          >{{ channel.user?.display_name || channel.user?.username || '未知作者' }}</RouterLink>
        </div>
        <div>
          <p class="a-label a-muted" style="margin-bottom:.4rem">更新时间</p>
          <p style="font-weight:700;margin:0">{{ formatDate(channel.updated_at) }}</p>
        </div>
      </div>

      <!-- Two-column layout: left collections, right posts -->
      <div class="channel-body">
        <!-- Left: collection list -->
        <aside class="collection-sidebar">
          <div class="section-headline">
            <h2 class="a-subtitle" style="margin:0;font-size:.875rem">合集</h2>
            <button v-if="isOwner" @click="openCollectionModal()" class="a-btn-outline-xs">+</button>
          </div>

          <div class="collection-list">
            <button
              class="collection-item"
              :class="{ active: activeCollectionId === null }"
              @click="activeCollectionId = null"
            >
              <span>全部内容</span>
              <span class="count">{{ channelPosts.length }}</span>
            </button>
            <button
              v-for="col in collections"
              :key="col.id"
              class="collection-item"
              :class="{ active: activeCollectionId === col.id }"
              @click="activeCollectionId = col.id"
            >
              <span class="a-clamp-1">{{ col.name }}</span>
              <span class="count">{{ postCountByCollection(col.id) }}</span>
            </button>
          </div>
        </aside>

        <!-- Right: posts -->
        <main class="post-main">
          <AEmpty v-if="!filteredPosts.length" title="暂无内容" description="该合集还没有文章" />
          <div v-else class="a-grid-2">
            <div v-for="post in filteredPosts" :key="post.id" class="a-card a-card-hover">
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
                  :to="`/post/${post.id}/edit`"
                  class="a-btn-outline-sm"
                >编辑</RouterLink>
              </div>
            </div>
          </div>
        </main>
      </div>
    </template>

    <!-- Collection Modal -->
    <AModal v-if="collectionModalOpen" @close="collectionModalOpen = false">
      <h3 class="a-subtitle" style="margin-bottom:1.5rem">{{ editingCollection ? '编辑合集' : '新建合集' }}</h3>
      <div style="display:flex;flex-direction:column;gap:1rem">
        <input v-model="collectionForm.name" placeholder="合集名称*" class="a-input" />
        <textarea v-model="collectionForm.description" placeholder="合集描述（可选）" rows="3" class="a-textarea" />
      </div>
      <div style="display:flex;gap:.75rem;margin-top:1.5rem;justify-content:flex-end">
        <ABtn outline @click="collectionModalOpen = false">取消</ABtn>
        <ABtn :disabled="!collectionForm.name.trim() || collectionSaving" @click="saveCollection">
          {{ collectionSaving ? '保存中...' : (editingCollection ? '更新' : '创建') }}
        </ABtn>
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
const channelPosts = ref<Post[]>([])
const activeCollectionId = ref<string | null>(null)

const collectionModalOpen = ref(false)
const editingCollection = ref<Collection | null>(null)
const collectionForm = ref({ name: '', description: '' })
const collectionSaving = ref(false)

const channelSubscribed = ref(false)
const channelSubscribeLoading = ref(false)
const toastVisible = ref(false)
const toastMessage = ref('')

// Support both /channel/:slug (new) and /blog/channel/:id (legacy)
const routeParam = computed(() => (typeof route.params.slug === 'string' ? route.params.slug : typeof route.params.id === 'string' ? route.params.id : ''))
const isSlug = computed(() => !/^[0-9a-f-]{36}$/.test(routeParam.value))

const authHeader = computed(() => ({ Authorization: `Bearer ${authStore.token}` }))
const isOwner = computed(() => !!channel.value && channel.value.user_id === authStore.user?.uuid)
const channelRssUrl = computed(() => {
  if (!channel.value?.id) return ''
  return api.feed.rss(channel.value.user?.username || '')
})

const filteredPosts = computed(() => {
  if (activeCollectionId.value === null) return channelPosts.value
  return channelPosts.value.filter(p =>
    (p.collections || []).some(c => c.id === activeCollectionId.value)
  )
})

const formatDate = (date: string) => new Date(date).toLocaleDateString('zh-CN')
const summarize = (content: string) =>
  content.replace(/```[\s\S]*?```/g, ' ').replace(/[>#*_`\[\]()!-]/g, ' ').replace(/\s+/g, ' ').trim().slice(0, 120) || '暂无摘要'
const postCountByCollection = (cid: string) =>
  channelPosts.value.filter(p => (p.collections || []).some(c => c.id === cid)).length

const fetchChannel = async () => {
  const url = isSlug.value
    ? api.blog.channelBySlug(routeParam.value)
    : api.blog.channel(routeParam.value)
  const res = await fetch(url)
  if (!res.ok) { channel.value = null; return }
  const data = await res.json()
  channel.value = (data.data || null) as Channel | null

  if (authStore.isAuthenticated && channel.value) {
    channelSubscribeLoading.value = true
    channelSubscribed.value = await feedStore.isSubscribedToChannel(channel.value.id)
    channelSubscribeLoading.value = false
  }
}

const fetchCollections = async () => {
  if (!channel.value) return
  const url = isSlug.value
    ? api.blog.channelCollectionsBySlug(routeParam.value)
    : api.blog.channelCollections(channel.value.id)
  const res = await fetch(url)
  if (res.ok) collections.value = (await res.json()).data || []
}

const fetchPosts = async () => {
  if (!channel.value) return
  const params = new URLSearchParams({ channel_id: channel.value.id, limit: '100' })
  const headers: Record<string, string> = {}
  if (authStore.token) headers['Authorization'] = `Bearer ${authStore.token}`
  const res = await fetch(`${api.blog.posts}?${params}`, { headers })
  if (res.ok) channelPosts.value = (await res.json()).data || []
}

const openCollectionModal = (collection?: Collection) => {
  editingCollection.value = collection || null
  collectionForm.value = { name: collection?.name || '', description: collection?.description || '' }
  collectionModalOpen.value = true
}

const saveCollection = async () => {
  if (!collectionForm.value.name.trim() || !channel.value) return
  collectionSaving.value = true
  try {
    if (editingCollection.value) {
      await fetch(api.blog.collection(editingCollection.value.id), {
        method: 'PUT',
        headers: { ...authHeader.value, 'Content-Type': 'application/json' },
        body: JSON.stringify(collectionForm.value)
      })
    } else {
      await fetch(api.blog.channelCollections(channel.value.id), {
        method: 'POST',
        headers: { ...authHeader.value, 'Content-Type': 'application/json' },
        body: JSON.stringify(collectionForm.value)
      })
    }
    collectionModalOpen.value = false
    await fetchCollections()
  } catch (e) { console.error(e) } finally { collectionSaving.value = false }
}

const toggleChannelSubscribe = async () => {
  if (!channel.value) return
  channelSubscribeLoading.value = true
  try {
    const success = channelSubscribed.value
      ? await feedStore.unsubscribeFromChannel(channel.value.id)
      : await feedStore.subscribeToChannel(channel.value.id)
    if (success) channelSubscribed.value = !channelSubscribed.value
  } finally { channelSubscribeLoading.value = false }
}

const copyRssLink = async () => {
  if (!channelRssUrl.value) return
  await navigator.clipboard.writeText(channelRssUrl.value)
  toastMessage.value = '已复制 RSS 链接'
  toastVisible.value = true
}

onMounted(async () => {
  try {
    await fetchChannel()
    if (!channel.value) return
    await Promise.all([fetchCollections(), fetchPosts()])
  } finally { loading.value = false }
})
</script>

<style scoped>
.channel-body {
  display: flex;
  gap: 2rem;
  margin-top: 2rem;
  align-items: flex-start;
}

.collection-sidebar {
  width: 13rem;
  flex-shrink: 0;
  position: sticky;
  top: 5rem;
}

.post-main { flex: 1; min-width: 0; }

.collection-list { display: flex; flex-direction: column; gap: .25rem; }

.collection-item {
  display: flex;
  justify-content: space-between;
  align-items: center;
  gap: .5rem;
  padding: .5rem .75rem;
  border: 2px solid transparent;
  background: none;
  cursor: pointer;
  text-align: left;
  font-size: .875rem;
  font-weight: 700;
  width: 100%;
  transition: border-color .1s, background .1s;
}
.collection-item:hover { border-color: #000; }
.collection-item.active { border-color: #000; background: #000; color: #fff; }
.collection-item .count {
  font-size: .75rem;
  font-weight: 400;
  opacity: .6;
  flex-shrink: 0;
}

.channel-meta-card {
  border: 2px solid #000;
  background: #fff;
  padding: 1rem 1.25rem;
  display: flex;
  gap: 1.5rem;
  align-items: flex-start;
  flex-wrap: wrap;
  margin-bottom: 1.5rem;
}

.section-headline {
  display: flex;
  align-items: center;
  justify-content: space-between;
  margin-bottom: .75rem;
}

.a-btn-outline-xs {
  padding: .1rem .4rem;
  font-size: .8rem;
  font-weight: 700;
  border: 2px solid #000;
  background: none;
  cursor: pointer;
}
.a-btn-outline-xs:hover { background: #000; color: #fff; }

@media (max-width: 768px) {
  .channel-body { flex-direction: column; }
  .collection-sidebar { width: 100%; position: static; }
  .collection-list { flex-direction: row; overflow-x: auto; padding-bottom: .5rem; }
  .collection-item { white-space: nowrap; }
}
</style>
