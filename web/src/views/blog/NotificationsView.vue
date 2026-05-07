<template>
  <div class="notif-page">
    <div class="notif-header">
      <h1 class="notif-title">通知</h1>
      <button v-if="unread > 0" class="btn-markall" @click="markAllRead">全部标为已读</button>
    </div>

    <div v-if="loading" class="notif-empty">加载中...</div>
    <div v-else-if="notifications.length === 0" class="notif-empty">暂无通知</div>
    <div v-else class="notif-list">
      <div
        v-for="n in notifications"
        :key="n.id"
        class="notif-item"
        :class="{ unread: !n.read_at }"
        @click="handleClick(n)"
      >
        <div class="notif-dot" v-if="!n.read_at" />
        <div class="notif-body">
          <p class="notif-content">{{ n.content }}</p>
          <span class="notif-time">{{ formatTime(n.created_at) }}</span>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { useFeedStore } from '@/stores/feed'
import type { Notification } from '@/types'

const feedStore = useFeedStore()
const router = useRouter()

const loading = ref(true)

const notifications = computed(() => feedStore.notifications)
const unread = computed(() => feedStore.unreadCount)

onMounted(async () => {
  loading.value = true
  try {
    await feedStore.fetchNotifications()
    if (feedStore.notifications.some((n) => !n.read_at)) {
      await feedStore.markAllRead()
    }
  } finally {
    loading.value = false
  }
})

async function markAllRead() {
  await feedStore.markAllRead()
}

function handleClick(n: Notification) {
  if (!n.read_at) {
    feedStore.markRead(n.id).catch(() => {})
  }
  if (n.target_type === 'post' && n.target_id) {
    router.push(`/post/${n.target_id}`)
  }
}

function formatTime(iso: string) {
  const d = new Date(iso)
  const now = Date.now()
  const diff = now - d.getTime()
  if (diff < 60000) return '刚刚'
  if (diff < 3600000) return `${Math.floor(diff / 60000)} 分钟前`
  if (diff < 86400000) return `${Math.floor(diff / 3600000)} 小时前`
  return d.toLocaleDateString('zh-CN')
}
</script>

<style scoped>
.notif-page {
  max-width: 640px;
  margin: 0 auto;
  padding: 4rem 2rem 8rem;
}
.notif-header {
  display: flex;
  align-items: baseline;
  justify-content: space-between;
  margin-bottom: 2rem;
  border-bottom: 2px solid #000;
  padding-bottom: 1rem;
}
.notif-title {
  font-size: 2rem;
  font-weight: 900;
  letter-spacing: -0.04em;
  margin: 0;
}
.btn-markall {
  font-size: 0.75rem;
  font-weight: 700;
  background: none;
  border: none;
  cursor: pointer;
  color: #6b7280;
  text-decoration: none;
}
.btn-markall:hover { text-decoration: underline; color: #000; }
.notif-empty {
  color: #9ca3af;
  font-size: 0.875rem;
  padding: 3rem 0;
  text-align: center;
}
.notif-list {
  display: flex;
  flex-direction: column;
}
.notif-item {
  display: flex;
  align-items: flex-start;
  gap: 0.75rem;
  padding: 1rem 0;
  border-bottom: 1px solid #e5e7eb;
  cursor: pointer;
}
.notif-item:hover .notif-content { text-decoration: underline; }
.notif-dot {
  width: 8px;
  height: 8px;
  border-radius: 9999px;
  background: #000;
  flex-shrink: 0;
  margin-top: 0.375rem;
}
.notif-body { flex: 1; }
.notif-content {
  font-size: 0.875rem;
  font-weight: 500;
  margin: 0 0 0.25rem;
  line-height: 1.4;
}
.notif-item.unread .notif-content { font-weight: 700; }
.notif-time {
  font-size: 0.75rem;
  color: #9ca3af;
}
</style>
