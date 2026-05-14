<template>
  <div class="a-page" style="padding-bottom:12rem">
    <APageHeader title="博客管理" subtitle="管理您的合集和文章">
      <template #actions>
        <button @click="showCreateCollectionModal" class="a-btn-secondary-sm">新建合集</button>
        <button @click="showCreateChannelModal" class="a-btn-secondary-sm">新建合集</button>
        <RouterLink to="/channels" class="a-btn-primary-sm">合集管理</RouterLink>
      </template>
    </APageHeader>

    <!-- Loading -->
    <div v-if="loadingChannels" style="display:flex;flex-direction:column;gap:1.5rem">
      <div class="a-skeleton" style="height:12rem" />
      <div class="a-skeleton" style="height:12rem" />
      <div class="a-skeleton" style="height:12rem" />
    </div>

    <!-- Empty state -->
    <AEmpty v-else-if="channels.length === 0" title="还没有创建合集" description="点击下方按钮创建第一个合集">
      <template #actions>
        <RouterLink to="/channels" class="a-btn-primary">创建合集</RouterLink>
      </template>
    </AEmpty>

    <!-- Channels with collections -->
    <div v-else style="display:flex;flex-direction:column;gap:3rem">
      <div v-for="channel in channels" :key="channel.id" style="display:flex;flex-direction:column;gap:1.5rem">
        <!-- Channel header -->
        <div style="display:flex;align-items:center;justify-content:space-between;gap:1rem">
          <div style="display:flex;align-items:center;gap:1rem">
            <h2 style="font-size:1.5rem;font-weight:900;margin:0">{{ channel.name }}</h2>
            <span class="a-badge">{{ channel.collections_count || 0 }}个合集 · {{ channel.posts_count || 0 }}篇文章</span>
          </div>
          <div style="display:flex;gap:.5rem">
            <button @click="editChannel(channel)" class="a-btn-outline-sm">管理合集</button>
          </div>
        </div>

        <!-- Collections under this channel -->
        <div v-if="channel.collections && channel.collections.length > 0" style="display:flex;flex-direction:column;gap:1rem">
          <div
            v-for="collection in channel.collections"
            :key="collection.id"
            class="a-card"
            style="transition:transform .2s"
          >
            <div style="display:flex;align-items:center;justify-content:space-between;gap:1rem">
              <div style="flex:1;min-width:0">
                <div style="display:flex;align-items:center;gap:.75rem;margin-bottom:.5rem">
                  <h3 style="font-size:1.125rem;font-weight:900;margin:0">{{ collection.name }}</h3>
                  <span class="a-badge a-badge-muted">{{ collection.posts_count || 0 }}篇文章</span>
                </div>
                <p v-if="collection.description" class="a-muted" style="font-size:.875rem;margin:0">{{ collection.description }}</p>
              </div>
              <div style="display:flex;gap:.5rem;flex-shrink:0">
                <RouterLink :to="`/collection/${collection.id}`" class="a-btn-outline-sm">查看合集</RouterLink>
              </div>
            </div>
          </div>
        </div>

        <!-- No collections message -->
        <div v-else style="padding:2rem;text-align:center;border:2px dashed #e5e7eb;border-radius:8px">
          <p class="a-muted" style="margin:0">该合集下还没有合集</p>
        </div>
      </div>
    </div>

    <!-- Floating Action Button -->
    <RouterLink
      to="/channels"
      class="a-fab"
      style="position:fixed;bottom:2rem;right:2rem;width:4rem;height:4rem;border-radius:50%;background:black;color:white;font-size:2rem;display:flex;align-items:center;justify-content:center;text-decoration:none;box-shadow:5px 5px 0 rgba(0,0,0,1);transition:all .2s"
      title="创建合集"
    >
      +
    </RouterLink>

    <!-- Create Collection Modal -->
    <AModal v-if="createCollectionModalVisible" @close="closeCreateCollectionModal" size="md">
      <div style="display:flex;flex-direction:column;gap:1.5rem">
        <div>
          <h3 style="font-size:1.25rem;font-weight:900;margin:0 0 1.5rem 0">创建合集</h3>
          <div style="display:flex;flex-direction:column;gap:1rem">
            <div>
              <label style="display:block;font-weight:bold;margin-bottom:0.5rem">合集名称 *</label>
              <AInput v-model="collectionFormData.name" placeholder="输入合集名称" />
            </div>
            <div>
              <label style="display:block;font-weight:bold;margin-bottom:0.5rem">所属合集 *</label>
              <ASelect v-model="collectionFormData.channel_id" :options="channelOptions" placeholder="选择合集" />
            </div>
            <div>
              <label style="display:block;font-weight:bold;margin-bottom:0.5rem">描述</label>
              <ATextarea v-model="collectionFormData.description" placeholder="合集描述（可选）" :rows="3" />
            </div>
          </div>
        </div>
        <div style="display:flex;gap:1rem;justify-content:flex-end;margin-top:1.5rem">
          <ABtn variant="secondary" @click="closeCreateCollectionModal">取消</ABtn>
          <ABtn variant="primary" @click="handleCreateCollection" :disabled="submitting">{{ submitting ? '创建中...' : '创建' }}</ABtn>
        </div>
      </div>
    </AModal>

    <!-- Create Channel Modal -->
    <AModal v-if="createModalVisible" @close="closeCreateModal" size="md">
      <div style="display:flex;flex-direction:column;gap:1.5rem">
        <div>
          <h3 style="font-size:1.25rem;font-weight:900;margin:0 0 1.5rem 0">创建合集</h3>
          <div style="display:flex;flex-direction:column;gap:1rem">
            <div>
              <label style="display:block;font-weight:bold;margin-bottom:0.5rem">合集名称 *</label>
              <AInput v-model="formData.name" placeholder="输入合集名称" />
            </div>
            <div>
              <label style="display:block;font-weight:bold;margin-bottom:0.5rem">描述</label>
              <ATextarea v-model="formData.description" placeholder="合集描述（可选）" :rows="3" />
            </div>
          </div>
        </div>
        <div style="display:flex;gap:1rem;justify-content:flex-end;margin-top:1.5rem">
          <ABtn variant="secondary" @click="closeCreateModal">取消</ABtn>
          <ABtn variant="primary" @click="handleCreateChannel" :disabled="submitting">{{ submitting ? '创建中...' : '创建' }}</ABtn>
        </div>
      </div>
    </AModal>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted, computed } from 'vue'
import { useRouter } from 'vue-router'
import APageHeader from '@/components/ui/APageHeader.vue'
import AEmpty from '@/components/ui/AEmpty.vue'
import ABtn from '@/components/ui/ABtn.vue'
import AModal from '@/components/ui/AModal.vue'
import AInput from '@/components/ui/AInput.vue'
import ATextarea from '@/components/ui/ATextarea.vue'
import ASelect from '@/components/ui/ASelect.vue'
import { useAuthStore } from '@/stores/auth'
import { useApi } from '@/composables/useApi'

interface Collection {
  id: string
  name: string
  description?: string
  posts_count?: number
}

interface Channel {
  id: string
  name: string
  description?: string
  collections_count?: number
  posts_count?: number
  collections?: Collection[]
}

const authStore = useAuthStore()
const api = useApi()
const router = useRouter()

const loadingChannels = ref(false)
const channels = ref<Channel[]>([])
const channelOptions = computed(() => channels.value.map(ch => ({ label: ch.name, value: ch.id })))

// Create channel modal state
const createModalVisible = ref(false)
const formData = ref({ name: '', description: '' })
const submitting = ref(false)

// Create collection modal state
const createCollectionModalVisible = ref(false)
const collectionFormData = ref({ name: '', description: '', channel_id: '' })

const loadChannels = async () => {
  loadingChannels.value = true
  try {
    // Load only current user's channels
    const channelsRes = await fetch(`${api.blog.channels}?user_id=${authStore.user?.uuid}`, { headers: { Authorization: `Bearer ${authStore.token}` } })
    if (channelsRes.ok) {
      const channelsData = await channelsRes.json()
      const channelList = channelsData.data || []
      
      // Load collections for each channel
      for (const channel of channelList) {
        const collectionsRes = await fetch(api.blog.channelCollections(channel.id), { headers: { Authorization: `Bearer ${authStore.token}` } })
        if (collectionsRes.ok) {
          const collectionsData = await collectionsRes.json()
          channel.collections = collectionsData.data || []
        }
      }
      
      channels.value = channelList
    }
  } catch (e) {
    console.error('Failed to load channels', e)
  } finally {
    loadingChannels.value = false
  }
}

const editChannel = (channel: Channel) => {
  // Navigate to channel management page
  router.push('/channels')
}

const showCreateChannelModal = () => {
  formData.value = { name: '', description: '' }
  createModalVisible.value = true
}

const showCreateCollectionModal = () => {
  if (channels.value.length === 0) {
    alert('请先创建合集')
    return
  }
  collectionFormData.value = { name: '', description: '', channel_id: '' }
  createCollectionModalVisible.value = true
}

const closeCreateModal = () => {
  createModalVisible.value = false
}

const closeCreateCollectionModal = () => {
  createCollectionModalVisible.value = false
}

const handleCreateChannel = async () => {
  if (!formData.value.name.trim()) {
    alert('请输入合集名称')
    return
  }

  submitting.value = true
  try {
    const res = await fetch(api.blog.channels, {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
        Authorization: `Bearer ${authStore.token}`
      },
      body: JSON.stringify({
        name: formData.value.name.trim(),
        description: formData.value.description.trim()
      })
    })

    if (!res.ok) {
      const error = await res.json()
      throw new Error(error.error || '创建失败')
    }

    // Success - reload channels and close modal
    await loadChannels()
    closeCreateModal()
    alert('合集创建成功！已自动生成默认合集。')
  } catch (err: any) {
    console.error('Failed to create channel:', err)
    alert(err.message || '创建失败，请重试')
  } finally {
    submitting.value = false
  }
}

const handleCreateCollection = async () => {
  if (!collectionFormData.value.name.trim()) {
    alert('请输入合集名称')
    return
  }
  if (!collectionFormData.value.channel_id) {
    alert('请选择所属合集')
    return
  }

  submitting.value = true
  try {
    const res = await fetch(api.blog.channelCollections(collectionFormData.value.channel_id), {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
        Authorization: `Bearer ${authStore.token}`
      },
      body: JSON.stringify({
        name: collectionFormData.value.name.trim(),
        description: collectionFormData.value.description.trim()
      })
    })

    if (!res.ok) {
      const error = await res.json()
      throw new Error(error.error || '创建失败')
    }

    // Success - reload channels and close modal
    await loadChannels()
    closeCreateCollectionModal()
    alert('合集创建成功！')
  } catch (err: any) {
    console.error('Failed to create collection:', err)
    alert(err.message || '创建失败，请重试')
  } finally {
    submitting.value = false
  }
}

onMounted(() => {
  loadChannels()
})
</script>

<style scoped>
.a-page {
  max-width: 1200px;
  margin: 0 auto;
  padding: 2rem 1rem;
}
</style>
