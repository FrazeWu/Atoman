<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import { useRoute, RouterLink } from 'vue-router'
import { useAuthStore } from '@/stores/auth'

const route = useRoute()
const authStore = useAuthStore()
const API_URL = import.meta.env.VITE_API_URL || '/api'

const albumId = computed(() => route.params.albumId as string)
const discussions = ref<any[]>([])
const loading = ref(true)
const error = ref('')
const total = ref(0)

// New discussion form
const newContent = ref('')
const submitting = ref(false)

// Active reply form
const replyingTo = ref<string | null>(null)
const replyContent = ref('')

const fetchDiscussions = async () => {
  loading.value = true
  error.value = ''
  try {
    const res = await fetch(`${API_URL}/albums/${albumId.value}/discussions?limit=50`, {
      headers: authStore.token ? { Authorization: `Bearer ${authStore.token}` } : {}
    })
    const data = await res.json()
    discussions.value = data.data || []
    total.value = data.total || 0
  } catch (e: any) {
    error.value = e.message || 'Failed to fetch discussions'
  } finally {
    loading.value = false
  }
}

const createDiscussion = async () => {
  if (!newContent.value.trim()) return
  submitting.value = true
  try {
    await fetch(`${API_URL}/albums/${albumId.value}/discussions`, {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
        Authorization: `Bearer ${authStore.token}`
      },
      body: JSON.stringify({ content: newContent.value })
    })
    newContent.value = ''
    fetchDiscussions()
  } catch (e: any) {
    alert(e.message || 'Failed to create discussion')
  } finally {
    submitting.value = false
  }
}

const replyTo = async (discussionId: string) => {
  if (!replyContent.value.trim()) return
  try {
    await fetch(`${API_URL}/albums/${albumId.value}/discussions/${discussionId}/reply`, {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
        Authorization: `Bearer ${authStore.token}`
      },
      body: JSON.stringify({ content: replyContent.value })
    })
    replyContent.value = ''
    replyingTo.value = null
    fetchDiscussions()
  } catch (e: any) {
    alert(e.message || 'Failed to reply')
  }
}

const deleteDiscussion = async (discussionId: string) => {
  if (!confirm('确定删除这条讨论？')) return
  try {
    await fetch(`${API_URL}/albums/${albumId.value}/discussions/${discussionId}`, {
      method: 'DELETE',
      headers: { Authorization: `Bearer ${authStore.token}` }
    })
    fetchDiscussions()
  } catch (e: any) {
    alert(e.message || 'Failed to delete')
  }
}

const formatDate = (date: string) => {
  return new Date(date).toLocaleString('zh-CN', {
    year: 'numeric',
    month: 'short',
    day: 'numeric',
    hour: '2-digit',
    minute: '2-digit'
  })
}

const canDelete = (discussion: any) => {
  return authStore.isAuthenticated && (
    authStore.user?.role === 'admin' ||
    discussion.user_id === authStore.user?.id
  )
}

onMounted(() => {
  fetchDiscussions()
})
</script>

<template>
  <div class="page-container">
    <RouterLink :to="`/music/albums/${albumId}`" class="back-link">
      ← 返回专辑
    </RouterLink>

    <div class="header">
      <h1 class="title">讨论</h1>
      <p class="subtitle">共 {{ total }} 条讨论</p>
    </div>

    <!-- New Discussion Form -->
    <div v-if="authStore.isAuthenticated" class="new-discussion">
      <h3 class="section-title">发起讨论</h3>
      <textarea
        v-model="newContent"
        class="textarea"
        placeholder="写下你的想法...（支持 Markdown）"
        rows="4"
      ></textarea>
      <button
        @click="createDiscussion"
        class="btn-submit"
        :disabled="!newContent.trim() || submitting"
      >
        {{ submitting ? '发布中...' : '发布' }}
      </button>
    </div>
    <div v-else class="login-prompt">
      <RouterLink to="/login" class="link">登录</RouterLink> 后参与讨论
    </div>

    <!-- Loading -->
    <div v-if="loading" class="loading">加载中...</div>

    <!-- Error -->
    <div v-else-if="error" class="error">{{ error }}</div>

    <!-- Discussion List -->
    <div v-else class="discussion-list">
      <div
        v-for="discussion in discussions"
        :key="discussion.id"
        class="discussion-item"
      >
        <div class="discussion-header">
          <div class="user-info">
            <div class="avatar">{{ discussion.user?.username?.[0]?.toUpperCase() || '?' }}</div>
            <div class="username">{{ discussion.user?.username || 'Unknown' }}</div>
          </div>
          <div class="meta">
            <span class="date">{{ formatDate(discussion.created_at) }}</span>
            <button
              v-if="canDelete(discussion)"
              @click="deleteDiscussion(discussion.id)"
              class="btn-delete"
            >
              删除
            </button>
          </div>
        </div>

        <div class="discussion-content" v-html="discussion.content"></div>

        <!-- Reply Button -->
        <button
          v-if="authStore.isAuthenticated"
          @click="replyingTo = discussion.id"
          class="btn-reply"
        >
          回复
        </button>

        <!-- Reply Form -->
        <div v-if="replyingTo === discussion.id" class="reply-form">
          <textarea
            v-model="replyContent"
            class="textarea"
            placeholder="写下你的回复..."
            rows="3"
          ></textarea>
          <div class="reply-actions">
            <button @click="replyingTo = null" class="btn-cancel">取消</button>
            <button
              @click="replyTo(discussion.id)"
              class="btn-submit"
              :disabled="!replyContent.trim()"
            >
              回复
            </button>
          </div>
        </div>

        <!-- Replies -->
        <div v-if="discussion.replies?.length" class="replies">
          <div
            v-for="reply in discussion.replies"
            :key="reply.id"
            class="reply-item"
          >
            <div class="reply-header">
              <div class="user-info">
                <div class="avatar small">{{ reply.user?.username?.[0]?.toUpperCase() || '?' }}</div>
                <div class="username">{{ reply.user?.username || 'Unknown' }}</div>
              </div>
              <span class="date">{{ formatDate(reply.created_at) }}</span>
            </div>
            <div class="reply-content">{{ reply.content }}</div>
          </div>
        </div>
      </div>

      <div v-if="discussions.length === 0" class="empty">
        暂无讨论，快来发起第一个话题吧！
      </div>
    </div>
  </div>
</template>

<style scoped>
.page-container {
  max-width: 900px;
  margin: 0 auto;
  padding: 2rem;
}

.back-link {
  display: inline-flex;
  align-items: center;
  gap: 0.5rem;
  margin-bottom: 2rem;
  font-weight: 700;
  text-decoration: none;
  color: #000;
}
.back-link:hover { text-decoration: underline; }

.header {
  margin-bottom: 2rem;
}

.title {
  font-size: 2rem;
  font-weight: 900;
  letter-spacing: -0.05em;
  margin: 0 0 0.5rem 0;
}

.subtitle {
  color: #6b7280;
  margin: 0;
}

/* New Discussion */
.new-discussion {
  margin-bottom: 2rem;
  padding: 1.5rem;
  border: 2px solid #000;
  background: #fff;
}

.section-title {
  font-size: 0.875rem;
  font-weight: 900;
  text-transform: uppercase;
  letter-spacing: 0.1em;
  margin: 0 0 1rem 0;
}

.textarea {
  width: 100%;
  padding: 1rem;
  border: 2px solid #000;
  font-size: 0.875rem;
  font-family: inherit;
  resize: vertical;
  margin-bottom: 1rem;
}

.textarea:focus {
  outline: none;
  box-shadow: 4px 4px 0px 0px rgba(0,0,0,1);
}

.btn-submit {
  padding: 0.75rem 2rem;
  background: #000;
  color: #fff;
  border: none;
  font-weight: 700;
  text-transform: uppercase;
  letter-spacing: 0.1em;
  cursor: pointer;
}

.btn-submit:disabled {
  background: #9ca3af;
  cursor: not-allowed;
}

.login-prompt {
  margin-bottom: 2rem;
  padding: 1rem;
  background: #f9fafb;
  border: 2px solid #000;
  text-align: center;
}

.link {
  color: #000;
  font-weight: 700;
}

/* Discussion List */
.discussion-list {
  display: flex;
  flex-direction: column;
  gap: 1.5rem;
}

.discussion-item {
  padding: 1.5rem;
  border: 2px solid #000;
  background: #fff;
}

.discussion-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 1rem;
}

.user-info {
  display: flex;
  align-items: center;
  gap: 0.75rem;
}

.avatar {
  width: 2.5rem;
  height: 2.5rem;
  background: #000;
  color: #fff;
  display: flex;
  align-items: center;
  justify-content: center;
  font-weight: 900;
  font-size: 1rem;
}

.avatar.small {
  width: 2rem;
  height: 2rem;
  font-size: 0.75rem;
}

.username {
  font-weight: 700;
}

.meta {
  display: flex;
  align-items: center;
  gap: 1rem;
}

.date {
  font-size: 0.75rem;
  color: #6b7280;
}

.btn-delete {
  font-size: 0.75rem;
  color: #dc2626;
  background: none;
  border: none;
  cursor: pointer;
  text-decoration: underline;
}

.discussion-content {
  line-height: 1.6;
  white-space: pre-wrap;
}

.btn-reply {
  margin-top: 1rem;
  padding: 0.5rem 1rem;
  background: none;
  border: 2px solid #000;
  font-weight: 700;
  font-size: 0.75rem;
  text-transform: uppercase;
  cursor: pointer;
}

.btn-reply:hover {
  background: #000;
  color: #fff;
}

/* Reply Form */
.reply-form {
  margin-top: 1rem;
  padding: 1rem;
  background: #f9fafb;
  border: 2px solid #000;
}

.reply-actions {
  display: flex;
  gap: 1rem;
  margin-top: 0.75rem;
}

.btn-cancel {
  padding: 0.5rem 1rem;
  background: #fff;
  border: 2px solid #000;
  font-weight: 700;
  cursor: pointer;
}

/* Replies */
.replies {
  margin-top: 1rem;
  padding-left: 1rem;
  border-left: 4px solid #f3f4f6;
}

.reply-item {
  padding: 1rem;
  margin-bottom: 0.75rem;
  background: #f9fafb;
}

.reply-item:last-child {
  margin-bottom: 0;
}

.reply-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 0.5rem;
}

.reply-content {
  font-size: 0.875rem;
  line-height: 1.5;
}

.loading,
.error,
.empty {
  text-align: center;
  padding: 3rem;
  color: #6b7280;
}

.error {
  color: #dc2626;
}
</style>