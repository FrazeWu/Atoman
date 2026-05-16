import { defineStore } from 'pinia'
import { ref } from 'vue'
import { useApi } from '@/composables/useApi'
import { useAuthStore } from '@/stores/auth'
import type { DMConversation, DMMessage, DMRealtimePayload } from '@/types'

export const useDMStore = defineStore('dm', () => {
  const api = useApi()
  const authStore = useAuthStore()

  const unreadCount = ref(0)
  const conversations = ref<DMConversation[]>([])
  const activeConversation = ref<string | null>(null)
  const messages = ref<DMMessage[]>([])
  const loading = ref(false)
  const total = ref(0)

  const authHeaders = () => ({
    Authorization: `Bearer ${authStore.token}`,
    'Content-Type': 'application/json',
  })

  const fetchUnreadCount = async () => {
    if (!authStore.token) return
    const res = await fetch(api.dm.unreadCount, { headers: authHeaders() })
    if (!res.ok) return
    const data = await res.json()
    unreadCount.value = data.count || 0
  }

  const fetchConversations = async () => {
    if (!authStore.token) return
    loading.value = true
    try {
      const res = await fetch(api.dm.conversations, { headers: authHeaders() })
      if (!res.ok) throw new Error('获取私信会话失败')
      const data = await res.json()
      conversations.value = data.data || []
    } finally {
      loading.value = false
    }
  }

  const openConversation = async (username: string, page = 1) => {
    if (!authStore.token) return
    loading.value = true
    activeConversation.value = username
    try {
      const res = await fetch(`${api.dm.conversation(username)}?page=${page}`, { headers: authHeaders() })
      const data = await res.json().catch(() => ({}))
      if (!res.ok) {
        throw new Error(data.message || data.error || '获取私信消息失败')
      }
      messages.value = data.data || []
      total.value = data.total || 0
      await markRead(username)
    } finally {
      loading.value = false
    }
  }

  const sendMessage = async (username: string, content: string, imageUrl = '') => {
    if (!authStore.token) return
    const res = await fetch(api.dm.conversation(username), {
      method: 'POST',
      headers: authHeaders(),
      body: JSON.stringify({ content, image_url: imageUrl }),
    })
    const data = await res.json().catch(() => ({}))
    if (!res.ok) {
      throw new Error(data.message || data.error || '发送私信失败')
    }
    messages.value = [...messages.value, data.data]
    await fetchConversations()
    return data.data as DMMessage
  }

  const uploadImage = async (file: File) => {
    if (!authStore.token) throw new Error('未登录')
    const formData = new FormData()
    formData.append('image', file)
    const res = await fetch(api.dm.upload, {
      method: 'POST',
      headers: { Authorization: `Bearer ${authStore.token}` },
      body: formData,
    })
    const data = await res.json().catch(() => ({}))
    if (!res.ok) {
      throw new Error(data.error || '图片上传失败')
    }
    return data.image_url as string
  }

  const markRead = async (username: string) => {
    if (!authStore.token) return
    const res = await fetch(api.dm.markRead(username), {
      method: 'PUT',
      headers: authHeaders(),
    })
    if (!res.ok) return
    const target = conversations.value.find((item) => item.other_username === username)
    if (target) {
      unreadCount.value = Math.max(0, unreadCount.value - target.unread_count)
      target.unread_count = 0
    }
  }

  const receiveDM = (payload: DMRealtimePayload) => {
    const active = activeConversation.value === payload.sender_username
    if (!active) {
      unreadCount.value += 1
    }

    const existingConversation = conversations.value.find((item) => item.conversation_id === payload.conversation_id)
    if (existingConversation) {
      existingConversation.preview = payload.content || '[图片]'
      existingConversation.last_message_at = payload.created_at
      if (!active) existingConversation.unread_count += 1
      conversations.value = [existingConversation, ...conversations.value.filter((item) => item.conversation_id !== payload.conversation_id)]
    }

    if (active) {
      messages.value = [...messages.value, {
        id: payload.message_id,
        conversation_id: payload.conversation_id,
        sender_id: payload.sender_id,
        sender: { username: payload.sender_username, email: '' },
        content: payload.content,
        image_url: payload.image_url,
        created_at: payload.created_at,
        updated_at: payload.created_at,
      }]
    }
  }

  return {
    unreadCount,
    conversations,
    activeConversation,
    messages,
    loading,
    total,
    fetchUnreadCount,
    fetchConversations,
    openConversation,
    sendMessage,
    uploadImage,
    markRead,
    receiveDM,
  }
})

