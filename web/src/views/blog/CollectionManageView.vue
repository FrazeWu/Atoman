<template>
  <div class="a-page" style="padding-bottom:12rem">
    <APageHeader title="合集管理" subtitle="创建和管理您的文章合集" />

    <!-- Loading -->
    <div v-if="loadingCollections" style="display:flex;flex-direction:column;gap:1.5rem">
      <div class="a-skeleton" style="height:8rem" />
      <div class="a-skeleton" style="height:8rem" />
      <div class="a-skeleton" style="height:8rem" />
    </div>

    <!-- Empty State -->
    <AEmpty v-else-if="collections.length === 0" title="暂无合集" description="点击下方按钮创建第一个合集" />

    <!-- Collections List -->
    <div v-else style="display:flex;flex-direction:column;gap:1.5rem">
      <div
        v-for="collection in collections"
        :key="collection.id"
        class="a-card"
        style="cursor:pointer;transition:transform 0.2s"
        @click="$router.push(`/collection/${collection.id}`)"
      >
        <div style="display:flex;justify-content:space-between;align-items:start">
          <div style="flex:1">
            <h3 style="font-size:1.25rem;font-weight:bold;margin-bottom:0.5rem">{{ collection.name }}</h3>
            <p v-if="collection.description" style="color:#666;margin-bottom:1rem">{{ collection.description }}</p>
            <div style="display:flex;gap:1rem;font-size:0.875rem;color:#999">
              <span>{{ collection.channel_name }}</span>
              <span>{{ collection.posts_count || 0 }}篇文章</span>
            </div>
          </div>
          <div style="display:flex;gap:0.5rem" @click.stop>
            <ABtn variant="secondary" size="sm" @click="showEditModal(collection)">编辑</ABtn>
            <ABtn variant="danger" size="sm" @click="showDeleteModal(collection)">删除</ABtn>
          </div>
        </div>
      </div>
    </div>

    <!-- Floating Action Button -->
    <button
      class="a-fab"
      style="position:fixed;bottom:2rem;right:2rem;width:4rem;height:4rem;border-radius:50%;background:black;color:white;font-size:2rem;display:flex;align-items:center;justify-content:center;border:none;cursor:pointer;box-shadow:5px 5px 0 rgba(0,0,0,1)"
      @click="showCreateModal"
    >
      +
    </button>

    <!-- Create/Edit Modal -->
    <AModal v-if="modalVisible" @close="closeModal" size="md">
      <div style="display:flex;flex-direction:column;gap:1.5rem">
        <div>
          <label style="display:block;font-weight:bold;margin-bottom:0.5rem">合集名称 *</label>
          <AInput v-model="formData.name" placeholder="输入合集名称" />
        </div>
        <div>
          <label style="display:block;font-weight:bold;margin-bottom:0.5rem">所属合集 *</label>
          <select v-model="formData.channel_id" style="width:100%;padding:0.75rem;border:2px solid black;font-family:inherit;font-size:1rem;background:white;outline:none;transition:all 0.2s">
            <option value="" disabled>选择合集</option>
            <option v-for="ch in channels" :key="ch.id" :value="ch.id">{{ ch.name }}</option>
          </select>
        </div>
        <div>
          <label style="display:block;font-weight:bold;margin-bottom:0.5rem">描述</label>
          <ATextarea v-model="formData.description" placeholder="合集描述（可选）" :rows="3" />
        </div>
      </div>
      <template #footer>
        <div style="display:flex;gap:1rem;justify-content:flex-end">
          <ABtn variant="secondary" @click="closeModal">取消</ABtn>
          <ABtn variant="primary" @click="handleSubmit" :disabled="submitting">{{ submitting ? '提交中...' : '确定' }}</ABtn>
        </div>
      </template>
    </AModal>

    <!-- Delete Confirmation Modal -->
    <AModal v-if="deleteModalVisible" @close="closeDeleteModal" size="sm">
      <p style="margin-bottom:1rem">确定要删除合集 <strong>{{ collectionToDelete?.name }}</strong>吗？</p>
      <p style="color:#666;font-size:0.875rem">删除后该合集下的文章将移至默认合集。</p>
      <template #footer>
        <div style="display:flex;gap:1rem;justify-content:flex-end">
          <ABtn variant="secondary" @click="closeDeleteModal">取消</ABtn>
          <ABtn variant="danger" @click="executeDelete" :disabled="deleting">{{ deleting ? '删除中...' : '确认删除' }}</ABtn>
        </div>
      </template>
    </AModal>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import APageHeader from '@/components/ui/APageHeader.vue'
import AEmpty from '@/components/ui/AEmpty.vue'
import ABtn from '@/components/ui/ABtn.vue'
import AModal from '@/components/ui/AModal.vue'
import AInput from '@/components/ui/AInput.vue'
import ATextarea from '@/components/ui/ATextarea.vue'
import { useApi } from '@/composables/useApi'
import { useAuthStore } from '@/stores/auth'

interface Collection {
  id: string
  name: string
  description?: string
  channel_id: string
  channel_name: string
  posts_count?: number
}

interface Channel {
  id: string
  name: string
}

const router = useRouter()
const api = useApi()
const authStore = useAuthStore()

const loadingCollections = ref(true)
const collections = ref<Collection[]>([])
const channels = ref<Channel[]>([])

const modalVisible = ref(false)
const modalMode = ref<'create' | 'edit'>('create')
const formData = ref({ name: '', channel_id: '', description: '' })
const submitting = ref(false)

const deleteModalVisible = ref(false)
const collectionToDelete = ref<Collection | null>(null)
const deleting = ref(false)

const loadChannels = async () => {
  try {
    const res = await fetch(`${api.blog.channels}?user_id=${authStore.user?.uuid}`, { headers: { Authorization: `Bearer ${authStore.token}` } })
    const data = await res.json()
    channels.value = data.data || []
  } catch (err) {
    console.error('Failed to load channels:', err)
  }
}

const loadCollections = async () => {
  loadingCollections.value = true
  try {
    const res = await fetch(api.blog.collections)
    const data = await res.json()
    collections.value = data.data || []
  } catch (err) {
    console.error('Failed to load collections:', err)
  } finally {
    loadingCollections.value = false
  }
}

const showCreateModal = () => {
  modalMode.value = 'create'
  formData.value = { name: '', channel_id: '', description: '' }
  modalVisible.value = true
}

const showEditModal = (collection: Collection) => {
  modalMode.value = 'edit'
  formData.value = { 
    name: collection.name, 
    channel_id: collection.channel_id, 
    description: collection.description || '' 
  }
  collectionToDelete.value = collection
  modalVisible.value = true
}

const closeModal = () => {
  modalVisible.value = false
  collectionToDelete.value = null
}

const handleSubmit = async () => {
  if (!formData.value.name.trim()) {
    alert('请输入合集名称')
    return
  }
  if (!formData.value.channel_id) {
    alert('请选择所属合集')
    return
  }

  submitting.value = true
  try {
    const url = modalMode.value === 'create' 
      ? api.blog.collections 
      : api.blog.collection(collectionToDelete.value!.id)
    const method = modalMode.value === 'create' ? 'POST' : 'PUT'
    
    const res = await fetch(url, {
      method,
      headers: {
        'Content-Type': 'application/json',
        Authorization: `Bearer ${authStore.token}`
      },
      body: JSON.stringify(formData.value)
    })
    
    const data = await res.json()
    if (!res.ok) throw new Error(data.error || '操作失败')
    
    modalVisible.value = false
    await loadCollections()
  } catch (err: any) {
    alert(err.message || '操作失败')
  } finally {
    submitting.value = false
  }
}

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
  
  deleting.value = true
  try {
    const res = await fetch(api.blog.collection(collectionToDelete.value.id), {
      method: 'DELETE',
      headers: {
        Authorization: `Bearer ${authStore.token}`
      }
    })
    
    const data = await res.json()
    if (!res.ok) throw new Error(data.error || '删除失败')
    
    deleteModalVisible.value = false
    await loadCollections()
  } catch (err: any) {
    alert(err.message || '删除失败')
  } finally {
    deleting.value = false
  }
}

onMounted(() => {
  loadChannels()
  loadCollections()
})
</script>

<style scoped>
.a-fab:hover {
  transform:scale(1.05);
  box-shadow:7px 7px 0 rgba(0,0,0,1);
}
</style>
