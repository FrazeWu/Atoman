import { ref } from 'vue'
import { defineStore } from 'pinia'
import type { Subscription, SubscriptionGroup, Notification } from '@/types'
import { useAuthStore } from '@/stores/auth'

const API_URL = import.meta.env.VITE_API_URL || '/api'

export const useFeedStore = defineStore('feed', () => {
  // Feed state
  const subscriptions = ref<Subscription[]>([])
  const groups = ref<SubscriptionGroup[]>([])
  const timeline = ref<any[]>([])
  const activeSource = ref<{ type: string; id: number } | null>(null)

  // Notification state (integrated, no separate notification.ts)
  const notifications = ref<Notification[]>([])
  const unreadCount = ref(0)

  let pollInterval: ReturnType<typeof setInterval> | null = null

  // --- Feed Actions ---

  const fetchSubscriptions = async () => {
    const authStore = useAuthStore()
    if (!authStore.isAuthenticated) return
    try {
      const res = await fetch(`${API_URL}/feed/subscriptions`, {
        headers: { Authorization: `Bearer ${authStore.token}` },
      })
      if (res.ok) {
        const data = await res.json()
        subscriptions.value = data.data || []
      }
    } catch (e) {
      console.error('Failed to fetch subscriptions', e)
    }
  }

  const fetchGroups = async () => {
    const authStore = useAuthStore()
    if (!authStore.isAuthenticated) return
    try {
      const res = await fetch(`${API_URL}/feed/groups`, {
        headers: { Authorization: `Bearer ${authStore.token}` },
      })
      if (res.ok) {
        const data = await res.json()
        groups.value = data.data || []
      }
    } catch (e) {
      console.error('Failed to fetch groups', e)
    }
  }

  const createGroup = async (name: string) => {
    const authStore = useAuthStore()
    try {
      const res = await fetch(`${API_URL}/feed/groups`, {
        method: 'POST',
        headers: { 'Content-Type': 'application/json', Authorization: `Bearer ${authStore.token}` },
        body: JSON.stringify({ name }),
      })
      if (res.ok) {
        await fetchGroups()
        return true
      }
    } catch (e) {
      console.error('Failed to create group', e)
    }
    return false
  }

  const updateGroup = async (id: string, name: string) => {
    const authStore = useAuthStore()
    try {
      const res = await fetch(`${API_URL}/feed/groups/${id}`, {
        method: 'PUT',
        headers: { 'Content-Type': 'application/json', Authorization: `Bearer ${authStore.token}` },
        body: JSON.stringify({ name }),
      })
      if (res.ok) {
        await fetchGroups()
        return true
      }
    } catch (e) {
      console.error('Failed to update group', e)
    }
    return false
  }

  const deleteGroup = async (id: string) => {
    const authStore = useAuthStore()
    try {
      await fetch(`${API_URL}/feed/groups/${id}`, {
        method: 'DELETE',
        headers: { Authorization: `Bearer ${authStore.token}` },
      })
      await fetchGroups()
      await fetchSubscriptions()
    } catch (e) {
      console.error('Failed to delete group', e)
    }
  }

  const setSubscriptionGroup = async (subId: number | string, groupId: string | null) => {
    const authStore = useAuthStore()
    try {
      await fetch(`${API_URL}/feed/subscriptions/${subId}/group`, {
        method: 'PUT',
        headers: { 'Content-Type': 'application/json', Authorization: `Bearer ${authStore.token}` },
        body: JSON.stringify({ group_id: groupId }),
      })
      await fetchSubscriptions()
    } catch (e) {
      console.error('Failed to set subscription group', e)
    }
  }

  const fetchTimeline = async (sourceType?: string, sourceId?: number) => {
    const authStore = useAuthStore()
    if (!authStore.isAuthenticated) return
    try {
      let url = `${API_URL}/feed/timeline?`
      if (sourceType && sourceId) url += `source_type=${sourceType}&source_id=${sourceId}`
      const res = await fetch(url, {
        headers: { Authorization: `Bearer ${authStore.token}` },
      })
      if (res.ok) {
        const data = await res.json()
        timeline.value = data.data || []
      }
    } catch (e) {
      console.error('Failed to fetch timeline', e)
    }
  }

  const subscribe = async (targetType: string, targetId: number, title?: string) => {
    const authStore = useAuthStore()
    try {
      const res = await fetch(`${API_URL}/feed/subscriptions`, {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
          Authorization: `Bearer ${authStore.token}`,
        },
        body: JSON.stringify({ target_type: targetType, target_id: targetId, title }),
      })
      if (res.ok) {
        await fetchSubscriptions()
      }
    } catch (e) {
      console.error('Failed to subscribe', e)
    }
  }

  const unsubscribe = async (subscriptionId: number) => {
    const authStore = useAuthStore()
    try {
      await fetch(`${API_URL}/feed/subscriptions/${subscriptionId}`, {
        method: 'DELETE',
        headers: { Authorization: `Bearer ${authStore.token}` },
      })
      await fetchSubscriptions()
    } catch (e) {
      console.error('Failed to unsubscribe', e)
    }
  }

  const markItemsRead = async (feedItemIds: string[]) => {
    const authStore = useAuthStore()
    if (!feedItemIds.length) return
    try {
      await fetch(`${API_URL}/feed/timeline/mark-read`, {
        method: 'POST',
        headers: { 'Content-Type': 'application/json', Authorization: `Bearer ${authStore.token}` },
        body: JSON.stringify({ feed_item_ids: feedItemIds }),
      })
    } catch (e) {
      console.error('Failed to mark items read', e)
    }
  }

  const markAllRead = async () => {
    const authStore = useAuthStore()
    try {
      await fetch(`${API_URL}/feed/timeline/mark-all-read`, {
        method: 'POST',
        headers: { Authorization: `Bearer ${authStore.token}` },
      })
    } catch (e) {
      console.error('Failed to mark all read', e)
    }
  }

  // --- Notification Actions (integrated here) ---

  const fetchUnreadCount = async () => {
    const authStore = useAuthStore()
    if (!authStore.isAuthenticated) return
    try {
      const res = await fetch(`${API_URL}/notifications/unread-count`, {
        headers: { Authorization: `Bearer ${authStore.token}` },
      })
      if (res.ok) {
        const data = await res.json()
        unreadCount.value = data.count ?? 0
      }
    } catch (e) {
      console.error('Failed to fetch unread count', e)
    }
  }

  const fetchNotifications = async () => {
    const authStore = useAuthStore()
    if (!authStore.isAuthenticated) return
    try {
      const res = await fetch(`${API_URL}/notifications`, {
        headers: { Authorization: `Bearer ${authStore.token}` },
      })
      if (res.ok) {
        const data = await res.json()
        notifications.value = data.data || []
        unreadCount.value = data.unread_count ?? 0
      }
    } catch (e) {
      console.error('Failed to fetch notifications', e)
    }
  }

  const markRead = async (id: number) => {
    const authStore = useAuthStore()
    try {
      await fetch(`${API_URL}/notifications/${id}/read`, {
        method: 'PUT',
        headers: { Authorization: `Bearer ${authStore.token}` },
      })
      const n = notifications.value.find((n) => n.id === id)
      if (n && !n.read_at) {
        n.read_at = new Date().toISOString()
        unreadCount.value = Math.max(0, unreadCount.value - 1)
      }
    } catch (e) {
      console.error('Failed to mark notification read', e)
    }
  }

  const markAllNotificationsRead = async () => {
    const authStore = useAuthStore()
    try {
      await fetch(`${API_URL}/notifications/read-all`, {
        method: 'PUT',
        headers: { Authorization: `Bearer ${authStore.token}` },
      })
      notifications.value.forEach((n) => {
        if (!n.read_at) n.read_at = new Date().toISOString()
      })
      unreadCount.value = 0
    } catch (e) {
      console.error('Failed to mark all notifications read', e)
    }
  }

  const startPolling = () => {
    const authStore = useAuthStore()
    if (!authStore.isAuthenticated) return
    fetchUnreadCount()
    if (!pollInterval) {
      pollInterval = setInterval(fetchUnreadCount, 60_000)
    }
  }

  const stopPolling = () => {
    if (pollInterval) {
      clearInterval(pollInterval)
      pollInterval = null
    }
  }

  return {
    // Feed
    subscriptions,
    groups,
    timeline,
    activeSource,
    fetchSubscriptions,
    fetchGroups,
    createGroup,
    updateGroup,
    deleteGroup,
    setSubscriptionGroup,
    fetchTimeline,
    subscribe,
    unsubscribe,
    markItemsRead,
    markAllFeedRead: markAllRead,
    // Notifications
    notifications,
    unreadCount,
    fetchNotifications,
    fetchUnreadCount,
    markRead,
    markAllRead: markAllNotificationsRead,
    startPolling,
    stopPolling,
  }
})
