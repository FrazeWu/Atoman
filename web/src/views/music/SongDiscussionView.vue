<script setup lang="ts">
import { computed, onMounted, ref } from 'vue'
import { useRoute, RouterLink } from 'vue-router'
import { useAuthStore } from '@/stores/auth'

const route = useRoute()
const authStore = useAuthStore()
const API_URL = import.meta.env.VITE_API_URL || '/api'

const songId = computed(() => route.params.songId as string)
const discussions = ref<any[]>([])
const loading = ref(true)
const error = ref('')
const total = ref(0)
const newContent = ref('')
const submitting = ref(false)
const replyingTo = ref<string | null>(null)
const replyContent = ref('')

const authHeaders = () => authStore.token ? { Authorization: `Bearer ${authStore.token}` } : {}

const fetchDiscussions = async () => {
  loading.value = true
  error.value = ''
  try {
    const res = await fetch(`${API_URL}/songs/${songId.value}/discussions?limit=50`, {
      headers: authHeaders(),
    })
    const data = await res.json()
    if (!res.ok) throw new Error(data.error || '加载讨论失败')
    discussions.value = data.data || []
    total.value = data.total || 0
  } catch (e: any) {
    error.value = e.message || '加载讨论失败'
  } finally {
    loading.value = false
  }
}

const createDiscussion = async () => {
  if (!newContent.value.trim()) return
  submitting.value = true
  try {
    const res = await fetch(`${API_URL}/songs/${songId.value}/discussions`, {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
        Authorization: `Bearer ${authStore.token}`,
      },
      body: JSON.stringify({ content: newContent.value }),
    })
    const data = await res.json().catch(() => ({}))
    if (!res.ok) throw new Error(data.error || '发布失败')
    newContent.value = ''
    await fetchDiscussions()
  } catch (e: any) {
    alert(e.message || '发布失败')
  } finally {
    submitting.value = false
  }
}

const replyTo = async (discussionId: string) => {
  if (!replyContent.value.trim()) return
  const res = await fetch(`${API_URL}/songs/${songId.value}/discussions/${discussionId}/reply`, {
    method: 'POST',
    headers: {
      'Content-Type': 'application/json',
      Authorization: `Bearer ${authStore.token}`,
    },
    body: JSON.stringify({ content: replyContent.value }),
  })
  const data = await res.json().catch(() => ({}))
  if (!res.ok) {
    alert(data.error || '回复失败')
    return
  }
  replyContent.value = ''
  replyingTo.value = null
  await fetchDiscussions()
}

const deleteDiscussion = async (discussionId: string) => {
  if (!confirm('确定删除这条讨论？')) return
  const res = await fetch(`${API_URL}/songs/${songId.value}/discussions/${discussionId}`, {
    method: 'DELETE',
    headers: authHeaders(),
  })
  const data = await res.json().catch(() => ({}))
  if (!res.ok) {
    alert(data.error || '删除失败')
    return
  }
  await fetchDiscussions()
}

const canDelete = (discussion: any) => {
  return authStore.isAuthenticated && (
    authStore.user?.role === 'admin' ||
    discussion.user_id === authStore.user?.uuid
  )
}

const formatDate = (date: string) => {
  return new Date(date).toLocaleString('zh-CN', {
    year: 'numeric',
    month: 'short',
    day: 'numeric',
    hour: '2-digit',
    minute: '2-digit',
  })
}

onMounted(fetchDiscussions)
</script>

<template>
  <div class="page-container">
    <RouterLink to="/music" class="back-link">← 返回音乐库</RouterLink>

    <header class="header">
      <h1 class="title">歌曲讨论</h1>
      <p class="subtitle">共 {{ total }} 条讨论</p>
    </header>

    <section v-if="authStore.isAuthenticated" class="new-discussion">
      <h2 class="section-title">发起讨论</h2>
      <textarea
        v-model="newContent"
        class="textarea"
        placeholder="写下你的想法..."
        rows="4"
      />
      <button
        @click="createDiscussion"
        class="btn-submit"
        :disabled="!newContent.trim() || submitting"
      >
        {{ submitting ? '发布中...' : '发布' }}
      </button>
    </section>
    <div v-else class="login-prompt">
      <RouterLink to="/login">登录</RouterLink> 后参与讨论
    </div>

    <div v-if="loading" class="loading">加载中...</div>
    <div v-else-if="error" class="error">{{ error }}</div>

    <section v-else class="discussion-list">
      <article v-for="discussion in discussions" :key="discussion.id" class="discussion-item">
        <div class="discussion-header">
          <div class="user-info">
            <div class="avatar">{{ discussion.user?.username?.[0]?.toUpperCase() || '?' }}</div>
            <strong>{{ discussion.user?.username || 'Unknown' }}</strong>
          </div>
          <div class="meta">
            <span>{{ formatDate(discussion.created_at) }}</span>
            <button v-if="canDelete(discussion)" @click="deleteDiscussion(discussion.id)" class="btn-delete">
              删除
            </button>
          </div>
        </div>

        <p class="discussion-content">{{ discussion.content }}</p>

        <button v-if="authStore.isAuthenticated" @click="replyingTo = discussion.id" class="btn-reply">
          回复
        </button>

        <div v-if="replyingTo === discussion.id" class="reply-form">
          <textarea v-model="replyContent" class="textarea" placeholder="写下你的回复..." rows="3" />
          <div class="reply-actions">
            <button @click="replyingTo = null" class="btn-cancel">取消</button>
            <button @click="replyTo(discussion.id)" class="btn-submit" :disabled="!replyContent.trim()">回复</button>
          </div>
        </div>

        <div v-if="discussion.replies?.length" class="replies">
          <div v-for="reply in discussion.replies" :key="reply.id" class="reply-item">
            <div class="reply-header">
              <div class="user-info">
                <div class="avatar small">{{ reply.user?.username?.[0]?.toUpperCase() || '?' }}</div>
                <strong>{{ reply.user?.username || 'Unknown' }}</strong>
              </div>
              <span>{{ formatDate(reply.created_at) }}</span>
            </div>
            <p class="reply-content">{{ reply.content }}</p>
          </div>
        </div>
      </article>

      <div v-if="discussions.length === 0" class="empty">暂无讨论</div>
    </section>
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
  margin-bottom: 2rem;
  font-weight: 800;
  color: #000;
  text-decoration: none;
}
.back-link:hover { text-decoration: underline; }
.header { margin-bottom: 2rem; }
.title {
  margin: 0;
  font-size: 2rem;
  font-weight: 900;
  letter-spacing: 0;
}
.subtitle { color: #666; margin-top: .5rem; }
.new-discussion,
.login-prompt,
.discussion-item,
.loading,
.error,
.empty {
  border: 2px solid #000;
  background: #fff;
}
.new-discussion {
  padding: 1rem;
  margin-bottom: 1.5rem;
}
.section-title {
  font-size: 1.1rem;
  margin: 0 0 .75rem;
}
.textarea {
  width: 100%;
  border: 2px solid #000;
  padding: .75rem;
  resize: vertical;
  font: inherit;
}
.btn-submit,
.btn-cancel,
.btn-reply,
.btn-delete {
  border: 2px solid #000;
  padding: .45rem .8rem;
  font-weight: 900;
  cursor: pointer;
}
.btn-submit {
  margin-top: .75rem;
  background: #000;
  color: #fff;
}
.btn-submit:disabled {
  opacity: .45;
  cursor: not-allowed;
}
.btn-cancel,
.btn-reply {
  background: #fff;
  color: #000;
}
.btn-delete {
  background: #fff;
  color: #dc2626;
}
.login-prompt,
.loading,
.error,
.empty {
  padding: 1.5rem;
  text-align: center;
}
.error { color: #dc2626; }
.discussion-list { display: flex; flex-direction: column; gap: 1rem; }
.discussion-item { padding: 1rem; }
.discussion-header,
.reply-header,
.meta,
.reply-actions {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: .75rem;
  flex-wrap: wrap;
}
.user-info {
  display: flex;
  align-items: center;
  gap: .5rem;
}
.avatar {
  display: grid;
  place-items: center;
  width: 2rem;
  height: 2rem;
  border: 2px solid #000;
  font-weight: 900;
}
.avatar.small {
  width: 1.5rem;
  height: 1.5rem;
  font-size: .75rem;
}
.discussion-content,
.reply-content {
  white-space: pre-wrap;
  line-height: 1.7;
}
.reply-form {
  margin-top: 1rem;
}
.replies {
  margin-top: 1rem;
  padding-left: 1rem;
  border-left: 2px solid #000;
}
.reply-item {
  padding: .75rem 0;
  border-bottom: 1px solid #e5e7eb;
}
.reply-item:last-child {
  border-bottom: 0;
}
</style>
