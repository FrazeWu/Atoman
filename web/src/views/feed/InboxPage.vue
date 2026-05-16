<template>
  <div class="a-page">
    <div class="a-section-header inbox-header">
      <div>
        <h1 class="a-title">收件箱</h1>
        <p class="a-muted">通知与私信统一收纳在这里。</p>
      </div>
    </div>

    <div class="inbox-layout">
      <aside class="a-card inbox-sidebar">
        <div class="inbox-tabs">
          <ABtn
            v-for="tab in tabs"
            :key="tab.key"
            :variant="activeTab === tab.key ? 'primary' : 'secondary'"
            size="sm"
            block
            @click="switchTab(tab.key)"
          >{{ tab.label }}</ABtn>
        </div>

        <div v-if="activeTab !== 'dm'" class="sidebar-list">
          <ABtn variant="secondary" size="sm" block @click="markCurrentNotificationsRead">当前全部已读</ABtn>
          <button
            v-for="item in notificationStore.notifications"
            :key="item.id"
            class="sidebar-item"
            :class="{ unread: !item.read_at, selected: selectedNotificationId === item.id }"
            @click="openNotification(item.id)"
          >
            <div class="sidebar-item-title">{{ formatNotificationTitle(item) }}</div>
            <div class="sidebar-item-body a-muted">{{ formatNotificationBody(item) }}</div>
            <div class="sidebar-item-time">{{ formatTime(item.created_at) }}</div>
          </button>
          <AEmpty v-if="!notificationStore.loading && notificationStore.notifications.length === 0" title="暂无通知" description="这里会显示回复、点赞和提及。" />
        </div>

        <div v-else class="sidebar-list">
          <ABtn variant="secondary" size="sm" block @click="dmStore.fetchConversations">刷新会话</ABtn>
          <button
            v-for="conversation in dmStore.conversations"
            :key="conversation.conversation_id"
            class="sidebar-item"
            :class="{ unread: conversation.unread_count > 0, selected: dmStore.activeConversation === conversation.other_username }"
            @click="openConversation(conversation.other_username)"
          >
            <div class="sidebar-item-title">
              <span>{{ conversation.other_username }}</span>
              <span v-if="conversation.unread_count > 0" class="sidebar-badge">{{ conversation.unread_count }}</span>
            </div>
            <div class="sidebar-item-body a-muted">{{ conversation.preview }}</div>
            <div class="sidebar-item-time">{{ formatTime(conversation.last_message_at) }}</div>
          </button>
          <AEmpty v-if="!dmStore.loading && dmStore.conversations.length === 0" title="暂无私信" description="你还没有任何私信会话。" />
        </div>
      </aside>

      <section class="a-card inbox-detail">
        <template v-if="activeTab !== 'dm'">
          <div v-if="selectedNotification" class="detail-card">
            <h2 class="a-subtitle">{{ formatNotificationTitle(selectedNotification) }}</h2>
            <p class="detail-body a-muted">{{ formatNotificationBody(selectedNotification) }}</p>
            <p class="detail-time">{{ formatTime(selectedNotification.created_at) }}</p>
            <ABtn @click="jumpToNotification(selectedNotification)">前往来源内容</ABtn>
          </div>
          <AEmpty v-else title="选择一条通知" description="点击左侧通知查看详情。" />
        </template>

        <template v-else>
          <div v-if="dmStore.activeConversation" class="detail-card detail-card-dm">
            <div class="dm-header">
              <h2 class="a-subtitle">与 {{ dmStore.activeConversation }} 的对话</h2>
            </div>

            <div class="dm-messages">
              <div
                v-for="message in dmStore.messages"
                :key="message.id"
                class="dm-message"
                :class="{ self: message.sender_id === authStore.user?.uuid }"
              >
                <div class="dm-bubble">
                  <p v-if="message.content">{{ message.content }}</p>
                  <img v-if="message.image_url" :src="message.image_url" alt="dm image" class="dm-image" />
                </div>
              </div>
            </div>

            <form class="dm-composer" @submit.prevent="submitDM">
              <ATextarea v-model="dmContent" label="消息内容" :rows="3" placeholder="输入私信内容" :error="dmError || undefined" />
              <div v-if="dmImageUrl" class="dm-upload-preview">
                <img :src="dmImageUrl" alt="preview" class="dm-image" />
              </div>
              <div class="dm-actions">
                <input ref="fileInput" type="file" accept="image/*" class="dm-file-input" @change="uploadDMImage" />
                <ABtn variant="secondary" type="button" @click="fileInput?.click()">上传图片</ABtn>
                <ABtn type="submit" :loading="dmSending" loadingText="发送中...">发送</ABtn>
              </div>
            </form>
          </div>
          <AEmpty v-else :title="dmOpenError ? '无法打开会话' : '选择一个会话'" :description="dmOpenError || '点击左侧私信会话开始聊天。'" />
        </template>
      </section>
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed, onMounted, ref, watch } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import ABtn from '@/components/ui/ABtn.vue'
import AEmpty from '@/components/ui/AEmpty.vue'
import ATextarea from '@/components/ui/ATextarea.vue'
import { useInboxStore } from '@/stores/inbox'
import { useNotificationStore } from '@/stores/notification'
import { useDMStore } from '@/stores/dm'
import { useAuthStore } from '@/stores/auth'
import type { InboxTab, Notification, NotificationFilterType } from '@/types'

const route = useRoute()
const router = useRouter()
const authStore = useAuthStore()
const inboxStore = useInboxStore()
const notificationStore = useNotificationStore()
const dmStore = useDMStore()

const tabs: Array<{ key: InboxTab; label: string }> = [
  { key: 'reply', label: '回复我的' },
  { key: 'like', label: '给我的赞' },
  { key: 'mention', label: '@我的' },
  { key: 'dm', label: '私信' },
]

const selectedNotificationId = ref<string | null>(null)
const dmContent = ref('')
const dmImageUrl = ref('')
const dmSending = ref(false)
const dmError = ref('')
const dmOpenError = ref('')
const fileInput = ref<HTMLInputElement | null>(null)

const activeTab = computed<InboxTab>(() => {
  const tab = route.query.tab
  if (tab === 'like' || tab === 'mention' || tab === 'dm') return tab
  return 'reply'
})

const selectedNotification = computed(() => notificationStore.notifications.find((item) => item.id === selectedNotificationId.value) || null)

const notificationTypeByTab: Record<'reply' | 'like' | 'mention', NotificationFilterType> = {
  reply: 'forum_reply',
  like: 'forum_like',
  mention: 'forum_mention',
}

const switchTab = async (tab: InboxTab) => {
  await router.push({ path: '/inbox', query: tab === 'reply' ? {} : { tab } })
}

const loadTab = async () => {
  if (activeTab.value === 'dm') {
    dmOpenError.value = ''
    await dmStore.fetchConversations()
    const user = typeof route.query.user === 'string' ? route.query.user : ''
    if (user) {
      try {
        await openConversation(user)
      } catch (error) {
        dmOpenError.value = error instanceof Error ? error.message : '打开私信失败'
      }
    }
    return
  }

  const type = notificationTypeByTab[activeTab.value]
  await notificationStore.fetchNotifications(type, 1)
  selectedNotificationId.value = notificationStore.notifications[0]?.id || null
}

const openNotification = async (id: string) => {
  selectedNotificationId.value = id
  await notificationStore.markRead(id)
}

const markCurrentNotificationsRead = async () => {
  if (activeTab.value === 'dm') return
  await notificationStore.markAllRead(notificationTypeByTab[activeTab.value])
}

const openConversation = async (username: string) => {
  dmOpenError.value = ''
  await router.replace({ path: '/inbox', query: { tab: 'dm', user: username } })
  await dmStore.openConversation(username)
}

const submitDM = async () => {
  if (!dmStore.activeConversation || (!dmContent.value.trim() && !dmImageUrl.value)) return
  dmSending.value = true
  dmError.value = ''
  try {
    await dmStore.sendMessage(dmStore.activeConversation, dmContent.value.trim(), dmImageUrl.value)
    dmContent.value = ''
    dmImageUrl.value = ''
  } catch (error) {
    dmError.value = error instanceof Error ? error.message : '发送失败'
  } finally {
    dmSending.value = false
  }
}

const uploadDMImage = async (event: Event) => {
  const target = event.target as HTMLInputElement
  const file = target.files?.[0]
  if (!file) return
  try {
    dmImageUrl.value = await dmStore.uploadImage(file)
  } catch (error) {
    dmError.value = error instanceof Error ? error.message : '上传失败'
  } finally {
    target.value = ''
  }
}

const jumpToNotification = async (notification: Notification) => {
  await notificationStore.markRead(notification.id)
  const topicId = notification.meta.topic_id
  if (notification.source_type === 'forum_reply' && topicId) {
    await router.push(`/topic/${topicId}#reply-${notification.source_id}`)
    return
  }
  if (notification.source_type === 'forum_topic' && notification.source_id) {
    await router.push(`/topic/${notification.source_id}`)
    return
  }
}

const formatNotificationTitle = (notification: Notification) => {
  const actor = notification.actor?.display_name || notification.actor?.username || '有人'
  switch (notification.type) {
    case 'forum_reply':
      return `${actor} 回复了你`
    case 'forum_like':
      return `${actor} 赞了你`
    case 'forum_mention':
      return `${actor} 提到了你`
    case 'forum_solved':
      return `${actor} 采纳了你的回复`
    default:
      return '新通知'
  }
}

const formatNotificationBody = (notification: Notification) => {
  return notification.meta.reply_excerpt || notification.meta.topic_title || '查看详情'
}

const formatTime = (value?: string | null) => {
  if (!value) return '刚刚'
  return new Date(value).toLocaleString('zh-CN')
}

watch(() => route.fullPath, () => {
  loadTab()
})

onMounted(async () => {
  await inboxStore.bootstrap()
  await loadTab()
})
</script>

<style scoped>
.inbox-header {
  margin-bottom: var(--a-space-5);
}

.inbox-layout {
  display: grid;
  grid-template-columns: 18rem minmax(0, 1fr);
  gap: var(--a-space-4);
  align-items: start;
}

.inbox-sidebar,
.inbox-detail {
  min-height: 70vh;
}

.inbox-sidebar {
  display: flex;
  flex-direction: column;
  gap: var(--a-space-4);
}

.inbox-tabs {
  display: flex;
  flex-direction: column;
  gap: var(--a-space-2);
}

.sidebar-list {
  display: flex;
  flex-direction: column;
  gap: var(--a-space-2);
}

.sidebar-item {
  border: var(--a-border);
  background: var(--a-color-bg);
  padding: var(--a-space-3);
  text-align: left;
  cursor: pointer;
  transition: transform 0.12s ease, box-shadow 0.12s ease;
}

.sidebar-item:hover {
  box-shadow: var(--a-shadow-button);
}

.sidebar-item.selected {
  border-color: var(--a-color-fg);
}

.sidebar-item.unread .sidebar-item-title {
  font-weight: var(--a-font-weight-black);
}

.sidebar-item-title {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: var(--a-space-2);
  font-size: var(--a-text-sm);
  font-weight: var(--a-font-weight-strong);
}

.sidebar-item-body {
  margin-top: var(--a-space-1);
  font-size: var(--a-text-sm);
}

.sidebar-item-time,
.detail-time {
  margin-top: var(--a-space-1);
  font-size: var(--a-text-xs);
  color: var(--a-color-muted);
}

.sidebar-badge {
  min-width: 1.5rem;
  padding: 0 var(--a-space-2);
  border: var(--a-border);
  background: var(--a-color-fg);
  color: var(--a-color-bg);
  font-size: var(--a-text-xs);
  font-weight: var(--a-font-weight-black);
  text-align: center;
}

.detail-card {
  display: flex;
  flex-direction: column;
  gap: var(--a-space-4);
}

.detail-body {
  white-space: pre-wrap;
}

.detail-card-dm {
  min-height: 60vh;
}

.dm-header {
  padding-bottom: var(--a-space-2);
  border-bottom: var(--a-border);
}

.dm-messages {
  display: flex;
  flex-direction: column;
  gap: var(--a-space-3);
  min-height: 20rem;
}

.dm-message {
  display: flex;
}

.dm-message.self {
  justify-content: flex-end;
}

.dm-bubble {
  max-width: 70%;
  border: var(--a-border);
  padding: var(--a-space-3);
  background: var(--a-color-bg);
}

.dm-message.self .dm-bubble {
  background: var(--a-color-surface);
}

.dm-image {
  display: block;
  max-width: 14rem;
  border: var(--a-border);
}

.dm-composer {
  display: flex;
  flex-direction: column;
  gap: var(--a-space-3);
  padding-top: var(--a-space-3);
  border-top: var(--a-border);
}

.dm-actions {
  display: flex;
  flex-wrap: wrap;
  gap: var(--a-space-2);
  justify-content: space-between;
}

.dm-file-input {
  display: none;
}

.dm-upload-preview {
  display: flex;
}

@media (max-width: 960px) {
  .inbox-layout {
    grid-template-columns: 1fr;
  }

  .inbox-sidebar,
  .inbox-detail {
    min-height: auto;
  }
}

@media (max-width: 640px) {
  .dm-bubble {
    max-width: 100%;
  }

  .dm-actions {
    flex-direction: column;
  }
}

</style>
