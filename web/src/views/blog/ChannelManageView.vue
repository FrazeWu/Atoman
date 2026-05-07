<template>
  <div class="a-page" style="padding-bottom:12rem">
    <APageHeader title="合集管理" subtitle="创建和管理您的合集">
      <template #action>
        <ABtn size="sm" @click="showCreateModal">新建合集</ABtn>
      </template>
    </APageHeader>

    <!-- Loading -->
    <div v-if="loadingChannels" style="display:flex;flex-direction:column;gap:1.5rem">
      <div class="a-skeleton" style="height:8rem" />
      <div class="a-skeleton" style="height:8rem" />
      <div class="a-skeleton" style="height:8rem" />
    </div>

    <!-- Empty State -->
    <AEmpty v-else-if="channels.length === 0" title="暂无合集" description="点击右上角按钮创建第一个合集" />

    <!-- Channels List -->
    <div v-else style="display:flex;flex-direction:column;gap:1.5rem">
      <div
        v-for="channel in channels"
        :key="channel.id"
        class="a-card"
      >
        <div style="display:flex;justify-content:space-between;align-items:start">
          <div style="flex:1;cursor:default">
            <h3 style="font-size:1.25rem;font-weight:bold;margin-bottom:0.5rem">{{ channel.name }}</h3>
            <p v-if="channel.description" style="color:#666;margin-bottom:1rem">{{ channel.description }}</p>
            <div style="display:flex;gap:1rem;font-size:0.875rem;color:#999">
              <span>{{ channel.collections_count || 0 }} 个合集</span>
              <span>{{ channel.posts_count || 0 }}篇文章</span>
            </div>
          </div>
          <div style="display:flex;gap:0.5rem">
            <ABtn outline size="sm" @click="showEditModal(channel)">编辑</ABtn>
            <ABtn v-if="!channel.is_default" outline size="sm" style="color:#ef4444;border-color:#ef4444" @click="showDeleteModal(channel)">删除</ABtn>
          </div>
        </div>
      </div>
    </div>

    <!-- Create/Edit Modal -->
    <AModal v-if="modalVisible" @close="closeModal" size="md">
      <h3 style="font-size:1.25rem;font-weight:900;margin:0 0 1.5rem 0">
        {{ modalMode === 'create' ? '创建合集' : '编辑合集' }}
      </h3>
      <div style="display:flex;flex-direction:column;gap:1.5rem">
        <div>
          <label style="display:block;font-weight:bold;margin-bottom:0.5rem">合集名称 *</label>
          <AInput v-model="formData.name" placeholder="输入合集名称" />
        </div>
        <div>
          <label style="display:block;font-weight:bold;margin-bottom:0.5rem">描述</label>
          <ATextarea v-model="formData.description" placeholder="合集描述（可选）" :rows="3" />
        </div>
      </div>
      <template #footer>
        <div style="display:flex;gap:1rem;justify-content:flex-end">
          <ABtn outline @click="closeModal">取消</ABtn>
          <ABtn @click="handleSubmit" :disabled="submitting">{{ submitting ? '提交中...' : '确定' }}</ABtn>
        </div>
      </template>
    </AModal>

    <!-- Delete Confirmation Modal -->
    <AModal v-if="deleteModalVisible" @close="closeDeleteModal" size="sm">
      <h3 style="font-size:1.125rem;font-weight:900;margin:0 0 1rem 0">确认删除</h3>
      <p style="margin-bottom:1rem">确定要删除合集 <strong>{{ channelToDelete?.name }}</strong>吗？</p>
      <p style="color:#666;font-size:0.875rem;margin-bottom:1rem">删除后该合集下的所有内容将被转移至默认合集。</p>
      <div>
        <label style="display:block;font-weight:bold;margin-bottom:0.5rem">请输入密码确认</label>
        <AInput v-model="deletePassword" type="password" placeholder="输入您的密码" />
      </div>
      <template #footer>
        <div style="display:flex;gap:1rem;justify-content:flex-end;margin-top:1rem">
          <ABtn outline @click="closeDeleteModal">取消</ABtn>
          <ABtn @click="executeDelete" :disabled="deleting || !deletePassword">{{ deleting ? '删除中...' : '确认删除' }}</ABtn>
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

interface Channel {
  id: string
  name: string
  description?: string
  is_default: boolean
  collections_count?: number
  posts_count?: number
}

const router = useRouter()
const api = useApi()
const authStore = useAuthStore()

const loadingChannels = ref(true)
const channels = ref<Channel[]>([])

const modalVisible = ref(false)
const modalMode = ref<'create' | 'edit'>('create')
const formData = ref({ name: '', description: '' })
const submitting = ref(false)

const deleteModalVisible = ref(false)
const channelToDelete = ref<Channel | null>(null)
const deletePassword = ref('')
const deleting = ref(false)

const channelToEdit = ref<Channel | null>(null)

const loadChannels = async () => {
  loadingChannels.value = true
  try {
    const res = await fetch(`${api.blog.channels}?user_id=${authStore.user?.uuid}`, { headers: { Authorization: `Bearer ${authStore.token}` } })
    const data = await res.json()
    channels.value = data.data || []

    if (channels.value.length === 0) {
      const ensureRes = await fetch(api.blog.channelEnsureDefault, {
        method: 'POST',
        headers: { Authorization: `Bearer ${authStore.token}` }
      })
      if (ensureRes.ok) {
        const ensureData = await ensureRes.json()
        channels.value = [ensureData.data]
      }
    }
  } catch (err) {
    console.error('Failed to load channels:', err)
  } finally {
    loadingChannels.value = false
  }
}

const showCreateModal = () => {
  modalMode.value = 'create'
  formData.value = { name: '', description: '' }
  modalVisible.value = true
}

const showEditModal = (channel: Channel) => {
  modalMode.value = 'edit'
  formData.value = { name: channel.name, description: channel.description || '' }
  channelToEdit.value = channel
  modalVisible.value = true
}

const closeModal = () => {
  modalVisible.value = false
  channelToEdit.value = null
}

const handleSubmit = async () => {
  if (!formData.value.name.trim()) {
    alert('请输入合集名称')
    return
  }

  submitting.value = true
  try {
    const url = modalMode.value === 'create' 
      ? api.blog.channels 
      : api.blog.channel(channelToEdit.value!.id)
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
    channelToEdit.value = null
    await loadChannels()
  } catch (err: any) {
    alert(err.message || '操作失败')
  } finally {
    submitting.value = false
  }
}

const showDeleteModal = (channel: Channel) => {
  channelToDelete.value = channel
  deleteModalVisible.value = true
}

const closeDeleteModal = () => {
  deleteModalVisible.value = false
  channelToDelete.value = null
  deletePassword.value = ''
}

const executeDelete = async () => {
  if (!channelToDelete.value || !deletePassword.value) return
  
  deleting.value = true
  try {
    const res = await fetch(api.blog.channel(channelToDelete.value.id), {
      method: 'DELETE',
      headers: {
        'Content-Type': 'application/json',
        Authorization: `Bearer ${authStore.token}`
      },
      body: JSON.stringify({ password: deletePassword.value, move_content: false })
    })
    
    const data = await res.json()
    if (!res.ok) throw new Error(data.error || '删除失败')
    
    deleteModalVisible.value = false
    deletePassword.value = ''
    await loadChannels()
  } catch (err: any) {
    alert(err.message || '删除失败')
  } finally {
    deleting.value = false
  }
}

onMounted(() => {
  loadChannels()
})
</script>

<style scoped>
.a-fab:hover {
  transform:scale(1.05);
  box-shadow:7px 7px 0 rgba(0,0,0,1);
}
</style>
