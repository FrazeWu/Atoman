<template>
  <div class="a-page" style="padding-bottom:12rem">
    <div v-if="loading" style="display:flex;flex-direction:column;gap:1.5rem">
      <div class="a-skeleton" style="height:4rem" />
      <div class="a-skeleton" style="height:20rem" />
    </div>

    <AEmpty v-else-if="!channel" title="频道不存在" description="该频道已被删除或您没有权限管理" />

    <template v-else>
      <APageHeader :title="`管理：${channel.name}`" accent>
        <template #action>
          <RouterLink :to="`/channel/${channel.slug || channel.id}`" class="a-btn-outline-sm">← 查看频道</RouterLink>
        </template>
      </APageHeader>

      <!-- Nav tabs -->
      <div class="manage-tabs">
        <button v-for="tab in tabs" :key="tab.id" class="manage-tab" :class="{ active: activeTab === tab.id }" @click="activeTab = tab.id">
          {{ tab.label }}
        </button>
      </div>

      <!-- Basic Info -->
      <section v-if="activeTab === 'info'" class="a-card manage-section">
        <h3 class="a-subtitle" style="margin-bottom:1.5rem">基本信息</h3>
        <div style="display:flex;flex-direction:column;gap:1rem;max-width:32rem">
          <div>
            <label class="a-label" style="margin-bottom:.4rem;display:block">频道名称 *</label>
            <input v-model="infoForm.name" class="a-input" placeholder="频道名称" />
          </div>
          <div>
            <label class="a-label" style="margin-bottom:.4rem;display:block">Slug（URL标识）</label>
            <input v-model="infoForm.slug" class="a-input" placeholder="my-channel" />
            <p class="a-muted" style="font-size:.8rem;margin-top:.25rem">访问路径：/channel/{{ infoForm.slug || channel.slug }}</p>
          </div>
          <div>
            <label class="a-label" style="margin-bottom:.4rem;display:block">简介</label>
            <textarea v-model="infoForm.description" class="a-textarea" rows="3" placeholder="频道简介" />
          </div>
          <div style="display:flex;gap:.75rem">
            <ABtn :disabled="infoSaving" @click="saveInfo">{{ infoSaving ? '保存中...' : '保存' }}</ABtn>
          </div>
        </div>
      </section>

      <!-- Collections -->
      <section v-if="activeTab === 'collections'" class="manage-section">
        <div style="display:flex;align-items:center;justify-content:space-between;margin-bottom:1.5rem">
          <h3 class="a-subtitle" style="margin:0">合集管理</h3>
          <ABtn size="sm" @click="openCollectionModal()">+ 新建合集</ABtn>
        </div>
        <AEmpty v-if="!collections.length" title="暂无合集" />
        <div v-else style="display:flex;flex-direction:column;gap:1rem">
          <div v-for="col in collections" :key="col.id" class="a-card" style="display:flex;align-items:center;justify-content:space-between;gap:1rem">
            <div>
              <span class="a-badge" :class="col.is_default ? 'a-badge-fill' : ''" style="margin-right:.5rem">
                {{ col.is_default ? '默认' : '合集' }}
              </span>
              <strong>{{ col.name }}</strong>
              <p v-if="col.description" class="a-muted" style="font-size:.8rem;margin:.25rem 0 0">{{ col.description }}</p>
            </div>
            <div style="display:flex;gap:.5rem;flex-shrink:0">
              <button class="a-btn-outline-sm" @click="openCollectionModal(col)">编辑</button>
              <button v-if="!col.is_default" class="a-btn-outline-sm" style="color:#ef4444;border-color:#ef4444" @click="confirmDeleteCollection(col)">删除</button>
            </div>
          </div>
        </div>
      </section>

      <!-- Posts -->
      <section v-if="activeTab === 'posts'" class="manage-section">
        <h3 class="a-subtitle" style="margin-bottom:1.5rem">内容管理</h3>
        <div v-if="loadingPosts" style="display:flex;flex-direction:column;gap:1rem">
          <div v-for="i in 4" :key="i" class="a-skeleton" style="height:4rem" />
        </div>
        <AEmpty v-else-if="!posts.length" title="暂无内容" />
        <div v-else style="display:flex;flex-direction:column;gap:.75rem">
          <div v-for="post in posts" :key="post.id" class="a-card" style="display:flex;align-items:center;justify-content:space-between;gap:1rem">
            <div style="flex:1;min-width:0">
              <div style="display:flex;align-items:center;gap:.5rem;margin-bottom:.25rem">
                <span class="a-badge" :class="post.status === 'published' ? 'a-badge-fill' : ''">
                  {{ post.status === 'published' ? '已发布' : '草稿' }}
                </span>
                <strong class="a-clamp-1" style="font-size:.9375rem">{{ post.title }}</strong>
              </div>
              <p class="a-muted" style="font-size:.8rem">{{ formatDate(post.updated_at) }}</p>
            </div>
            <div style="display:flex;gap:.5rem;flex-shrink:0">
              <RouterLink :to="`/post/${post.id}/edit`" class="a-btn-outline-sm">编辑</RouterLink>
              <RouterLink :to="`/post/${post.id}`" class="a-btn-outline-sm">查看</RouterLink>
            </div>
          </div>
        </div>
      </section>

      <!-- Danger zone -->
      <section v-if="activeTab === 'danger'" class="manage-section">
        <h3 class="a-subtitle" style="margin-bottom:1.5rem;color:#ef4444">危险操作</h3>
        <div class="a-card" style="border-color:#ef4444">
          <h4 style="font-weight:900;margin-bottom:.5rem">删除频道</h4>
          <p class="a-muted" style="margin-bottom:1rem;font-size:.875rem">删除后所有内容将被永久删除，此操作不可恢复。</p>
          <ABtn variant="danger" @click="showDeleteChannel = true">删除频道</ABtn>
        </div>
      </section>
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

    <!-- Delete channel confirm -->
    <AModal v-if="showDeleteChannel" @close="showDeleteChannel = false">
      <h3 class="a-subtitle" style="margin-bottom:1rem">确认删除频道</h3>
      <p class="a-muted" style="margin-bottom:1rem">请输入频道名称 <strong>{{ channel?.name }}</strong> 以确认删除：</p>
      <input v-model="deleteConfirmName" class="a-input" placeholder="输入频道名称" style="margin-bottom:1.5rem" />
      <div style="display:flex;gap:.75rem;justify-content:flex-end">
        <ABtn outline @click="showDeleteChannel = false">取消</ABtn>
        <ABtn
          :disabled="deleteConfirmName !== channel?.name || deleting"
          @click="deleteChannel"
          style="background:#ef4444;border-color:#ef4444;color:#fff"
        >{{ deleting ? '删除中...' : '确认删除' }}</ABtn>
      </div>
    </AModal>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import { RouterLink, useRoute, useRouter } from 'vue-router'
import AEmpty from '@/components/ui/AEmpty.vue'
import APageHeader from '@/components/ui/APageHeader.vue'
import AModal from '@/components/ui/AModal.vue'
import ABtn from '@/components/ui/ABtn.vue'
import type { Channel, Collection, Post } from '@/types'
import { useApi } from '@/composables/useApi'
import { useAuthStore } from '@/stores/auth'

const route = useRoute()
const router = useRouter()
const api = useApi()
const authStore = useAuthStore()

const channel = ref<Channel | null>(null)
const collections = ref<Collection[]>([])
const posts = ref<Post[]>([])
const loading = ref(true)
const loadingPosts = ref(false)

const activeTab = ref('info')
const tabs = [
  { id: 'info', label: '基本信息' },
  { id: 'collections', label: '合集' },
  { id: 'posts', label: '内容' },
  { id: 'danger', label: '危险操作' },
]

const infoForm = ref({ name: '', slug: '', description: '' })
const infoSaving = ref(false)

const collectionModalOpen = ref(false)
const editingCollection = ref<Collection | null>(null)
const collectionForm = ref({ name: '', description: '' })
const collectionSaving = ref(false)

const showDeleteChannel = ref(false)
const deleteConfirmName = ref('')
const deleting = ref(false)

const slug = computed(() => route.params.slug as string)
const authHeader = computed(() => ({ Authorization: `Bearer ${authStore.token}` }))

const formatDate = (d: string) => new Date(d).toLocaleDateString('zh-CN')

const fetchChannel = async () => {
  const res = await fetch(api.blog.channelBySlug(slug.value))
  if (!res.ok) { channel.value = null; return }
  channel.value = (await res.json()).data || null
  if (channel.value) {
    infoForm.value = {
      name: channel.value.name,
      slug: channel.value.slug,
      description: channel.value.description || '',
    }
    // Verify ownership
    if (channel.value.user_id !== authStore.user?.uuid) {
      router.push(`/channel/${slug.value}`)
    }
  }
}

const fetchCollections = async () => {
  if (!channel.value) return
  const res = await fetch(api.blog.channelCollections(channel.value.id))
  if (res.ok) collections.value = (await res.json()).data || []
}

const fetchPosts = async () => {
  if (!channel.value) return
  loadingPosts.value = true
  try {
    const res = await fetch(`${api.blog.posts}?channel_id=${channel.value.id}&limit=50`, { headers: authHeader.value })
    if (res.ok) posts.value = (await res.json()).data || []
  } finally { loadingPosts.value = false }
}

const saveInfo = async () => {
  if (!channel.value || !infoForm.value.name.trim()) return
  infoSaving.value = true
  try {
    const res = await fetch(api.blog.channel(channel.value.id), {
      method: 'PUT',
      headers: { ...authHeader.value, 'Content-Type': 'application/json' },
      body: JSON.stringify(infoForm.value),
    })
    if (res.ok) {
      const newSlug = infoForm.value.slug
      channel.value = (await res.json()).data || channel.value
      if (newSlug && newSlug !== slug.value) {
        router.replace(`/channel/${newSlug}/manage`)
      }
    }
  } finally { infoSaving.value = false }
}

const openCollectionModal = (col?: Collection) => {
  editingCollection.value = col || null
  collectionForm.value = { name: col?.name || '', description: col?.description || '' }
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
        body: JSON.stringify(collectionForm.value),
      })
    } else {
      await fetch(api.blog.channelCollections(channel.value.id), {
        method: 'POST',
        headers: { ...authHeader.value, 'Content-Type': 'application/json' },
        body: JSON.stringify(collectionForm.value),
      })
    }
    collectionModalOpen.value = false
    await fetchCollections()
  } finally { collectionSaving.value = false }
}

const confirmDeleteCollection = async (col: Collection) => {
  if (!confirm(`确认删除合集「${col.name}」？`)) return
  await fetch(api.blog.collection(col.id), { method: 'DELETE', headers: authHeader.value })
  await fetchCollections()
}

const deleteChannel = async () => {
  if (!channel.value || deleteConfirmName.value !== channel.value.name) return
  deleting.value = true
  try {
    const res = await fetch(api.blog.channel(channel.value.id), { method: 'DELETE', headers: authHeader.value })
    if (res.ok) router.push('/blog')
  } finally { deleting.value = false }
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
.manage-tabs {
  display: flex;
  gap: 0;
  border-bottom: 2px solid #000;
  margin-bottom: 2rem;
  overflow-x: auto;
}
.manage-tab {
  padding: .625rem 1.25rem;
  font-size: .8125rem;
  font-weight: 700;
  text-transform: uppercase;
  letter-spacing: .05em;
  border: none;
  background: none;
  cursor: pointer;
  border-bottom: 3px solid transparent;
  margin-bottom: -2px;
  white-space: nowrap;
}
.manage-tab:hover { background: #f5f5f5; }
.manage-tab.active { border-bottom-color: #000; }
.manage-section { padding-bottom: 2rem; }
</style>
