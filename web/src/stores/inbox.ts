import { computed, ref } from 'vue'
import { defineStore } from 'pinia'
import { useAuthStore } from '@/stores/auth'
import { useNotificationStore } from '@/stores/notification'
import { useDMStore } from '@/stores/dm'

export const useInboxStore = defineStore('inbox', () => {
  const authStore = useAuthStore()
  const notificationStore = useNotificationStore()
  const dmStore = useDMStore()

  const connected = ref(false)
  const polling = ref(false)
  const initialized = ref(false)
  let socket: WebSocket | null = null
  let pollingTimer: number | null = null

  const totalUnread = computed(() => notificationStore.unreadCount + dmStore.unreadCount)

  const stopPolling = () => {
    if (pollingTimer) {
      window.clearInterval(pollingTimer)
      pollingTimer = null
    }
    polling.value = false
  }

  const startPolling = () => {
    if (pollingTimer || !authStore.isAuthenticated) return
    polling.value = true
    pollingTimer = window.setInterval(async () => {
      await Promise.all([
        notificationStore.fetchUnreadCount(),
        dmStore.fetchUnreadCount(),
      ])
    }, 60000)
  }

  const disconnect = () => {
    stopPolling()
    connected.value = false
    if (socket) {
      socket.close()
      socket = null
    }
  }

  const connect = async () => {
    if (!authStore.token || socket) return
    const apiBase = (import.meta.env.VITE_API_URL || '/api').replace(/\/api$/, '')
    const protocol = window.location.protocol === 'https:' ? 'wss:' : 'ws:'
    const host = apiBase.startsWith('http')
      ? apiBase.replace(/^http:/, 'ws:').replace(/^https:/, 'wss:')
      : `${protocol}//${window.location.host}`
    socket = new WebSocket(`${host}/ws/user?token=${encodeURIComponent(authStore.token)}`)

    socket.onopen = () => {
      connected.value = true
      stopPolling()
    }
    socket.onclose = () => {
      connected.value = false
      socket = null
      startPolling()
    }
    socket.onerror = () => {
      connected.value = false
      startPolling()
    }
    socket.onmessage = async (event) => {
      const payload = JSON.parse(event.data)
      if (payload.event === 'notification') {
        notificationStore.receiveNotification(payload.data)
      }
      if (payload.event === 'dm') {
        dmStore.receiveDM(payload.data)
        if (dmStore.activeConversation === payload.data.sender_username) {
          await dmStore.markRead(payload.data.sender_username)
        }
      }
    }
  }

  const bootstrap = async () => {
    if (!authStore.isAuthenticated) {
      disconnect()
      initialized.value = false
      return
    }
    await Promise.all([
      notificationStore.fetchUnreadCount(),
      dmStore.fetchUnreadCount(),
    ])
    await connect()
    initialized.value = true
  }

  return {
    connected,
    polling,
    initialized,
    totalUnread,
    bootstrap,
    connect,
    disconnect,
  }
})
