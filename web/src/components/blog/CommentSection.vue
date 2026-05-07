<template>
  <div class="comment-section">
    <h3 class="section-title">
      评论 <span class="comment-count">({{ comments.length }})</span>
    </h3>

    <!-- Closed comments notice -->
    <div v-if="!allowComments" class="closed-notice">
      作者已关闭评论
    </div>

    <template v-else>
      <!-- Comment input (logged in) -->
      <div v-if="authStore.isAuthenticated" class="comment-form">
        <textarea
          v-model="newComment"
          placeholder="写下你的评论..."
          rows="3"
          class="comment-input"
        />
        <button
          @click="submitComment"
          :disabled="!newComment.trim() || submitting"
          class="submit-btn"
        >
          {{ submitting ? '发送中...' : '发表评论' }}
        </button>
      </div>
      <div v-else class="login-prompt">
        <RouterLink to="/login" class="login-link">登录</RouterLink> 后发表评论
      </div>

      <!-- Loading -->
      <div v-if="loading" class="loading-list">
        <div v-for="i in 3" :key="i" class="loading-item" />
      </div>

      <!-- Comment list -->
      <div v-else class="comment-list">
        <div v-for="comment in comments" :key="comment.id" class="comment-item">
          <div class="comment-header">
            <div class="comment-avatar">
              {{ (comment.user?.display_name || comment.user?.username || '?').charAt(0).toUpperCase() }}
            </div>
            <span class="comment-author">{{ comment.user?.display_name || comment.user?.username }}</span>
            <span class="comment-time">{{ formatDate(comment.created_at) }}</span>
            <button
              v-if="canDelete(comment)"
              @click="requestDeleteComment(comment.id)"
              class="delete-btn"
            >
              删除
            </button>
          </div>
          <p class="comment-content">{{ comment.content }}</p>
        </div>

        <div v-if="!comments.length" class="empty-comments">
          还没有评论，来发表第一条吧
        </div>
      </div>
    </template>

    <AConfirm
      :show="showDeleteConfirm"
      title="删除评论"
      message="确定删除这条评论吗？"
      confirm-text="删除"
      cancel-text="取消"
      danger
      @confirm="confirmDeleteComment"
      @cancel="cancelDeleteComment"
    />
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { RouterLink } from 'vue-router'
import AConfirm from '@/components/ui/AConfirm.vue'
import { useAuthStore } from '@/stores/auth'
import { useApi } from '@/composables/useApi'
import type { Comment } from '@/types'

const props = defineProps<{
  postId: string
  allowComments: boolean
  postOwnerId?: string
}>()

const authStore = useAuthStore()
const api = useApi()

const comments = ref<Comment[]>([])
const loading = ref(true)
const newComment = ref('')
const submitting = ref(false)
const showDeleteConfirm = ref(false)
const pendingDeleteCommentId = ref<string | null>(null)

const formatDate = (d: string) => new Date(d).toLocaleDateString('zh-CN')

const canDelete = (c: Comment) => {
  if (!authStore.user) return false
  return authStore.user.uuid === c.user_id || authStore.user.uuid === props.postOwnerId || authStore.user.role === 'admin'
}

const fetchComments = async () => {
  loading.value = true
  try {
    const res = await fetch(api.blog.postComments(props.postId))
    if (res.ok) {
      const d = await res.json()
      comments.value = d.data || []
    }
  } catch (e) {
    console.error(e)
  } finally {
    loading.value = false
  }
}

const submitComment = async () => {
  if (!newComment.value.trim()) return
  submitting.value = true
  try {
    const res = await fetch(api.blog.postComments(props.postId), {
      method: 'POST',
      headers: { 'Content-Type': 'application/json', Authorization: `Bearer ${authStore.token}` },
      body: JSON.stringify({ content: newComment.value })
    })
    if (res.ok) {
      newComment.value = ''
      await fetchComments()
    }
  } catch (e) {
    console.error(e)
  } finally {
    submitting.value = false
  }
}

const deleteComment = async (id: string) => {
  try {
    const res = await fetch(`${api.blog.comments}/${id}`, {
      method: 'DELETE',
      headers: { Authorization: `Bearer ${authStore.token}` }
    })
    if (res.ok) await fetchComments()
  } catch (e) {
    console.error(e)
  }
}

const requestDeleteComment = (id: string) => {
  pendingDeleteCommentId.value = id
  showDeleteConfirm.value = true
}

const cancelDeleteComment = () => {
  showDeleteConfirm.value = false
  pendingDeleteCommentId.value = null
}

const confirmDeleteComment = async () => {
  const id = pendingDeleteCommentId.value
  cancelDeleteComment()
  if (id !== null) {
    await deleteComment(id)
  }
}

onMounted(fetchComments)
</script>

<style scoped>
.comment-section {
  margin-top: 3rem;
  padding-top: 2rem;
  border-top: 2px solid #000;
}
.section-title {
  font-size: 1.5rem;
  font-weight: 900;
  letter-spacing: -0.03em;
  margin: 0 0 1.5rem;
}
.comment-count { color: #9ca3af; font-weight: 500; font-size: 1.125rem; }
.closed-notice {
  border: 2px dashed #d1d5db;
  padding: 2rem;
  text-align: center;
  color: #9ca3af;
  font-weight: 500;
}
.comment-form { margin-bottom: 2rem; }
.comment-input {
  width: 100%;
  border: 2px solid #000;
  padding: 1rem;
  font-size: 0.875rem;
  font-weight: 500;
  resize: none;
  outline: none;
  transition: box-shadow 0.2s;
  box-sizing: border-box;
  font-family: inherit;
}
.comment-input:focus { box-shadow: 5px 5px 0px 0px rgba(0,0,0,1); }
.submit-btn {
  margin-top: 0.5rem;
  background: #000;
  color: #fff;
  padding: 0.5rem 1.5rem;
  font-size: 0.75rem;
  font-weight: 900;
  text-transform: uppercase;
  letter-spacing: 0.1em;
  border: 2px solid #000;
  cursor: pointer;
  transition: all 0.2s;
}
.submit-btn:hover:not(:disabled) { background: #fff; color: #000; }
.submit-btn:disabled { opacity: 0.4; cursor: not-allowed; }
.login-prompt {
  margin-bottom: 2rem;
  border: 2px dashed #d1d5db;
  padding: 1.5rem;
  text-align: center;
  color: #9ca3af;
  font-weight: 500;
}
.login-link { font-weight: 900; text-decoration: underline; color: #000; }
.login-link:hover { opacity: 0.7; }
.loading-list { display: flex; flex-direction: column; gap: 1rem; }
.loading-item {
  height: 80px;
  background: #f3f4f6;
  border: 1px solid #e5e7eb;
  animation: pulse 2s infinite;
}
@keyframes pulse {
  0%, 100% { opacity: 1; }
  50% { opacity: 0.5; }
}
.comment-list { display: flex; flex-direction: column; gap: 1rem; }
.comment-item {
  border-left: 4px solid #000;
  padding: 0.75rem 0 0.75rem 1.5rem;
}
.comment-header {
  display: flex;
  align-items: center;
  gap: 0.75rem;
  margin-bottom: 0.5rem;
  flex-wrap: wrap;
}
.comment-avatar {
  width: 28px;
  height: 28px;
  border-radius: 9999px;
  background: #000;
  color: #fff;
  font-size: 0.75rem;
  font-weight: 900;
  display: flex;
  align-items: center;
  justify-content: center;
  flex-shrink: 0;
}
.comment-author { font-size: 0.875rem; font-weight: 900; }
.comment-time { font-size: 0.75rem; color: #9ca3af; font-weight: 500; }
.delete-btn {
  margin-left: auto;
  background: none;
  border: none;
  cursor: pointer;
  font-size: 0.75rem;
  color: #ef4444;
  font-weight: 700;
}
.delete-btn:hover { color: #b91c1c; }
.comment-content {
  font-size: 0.875rem;
  color: #1f2937;
  font-weight: 500;
  white-space: pre-wrap;
  margin: 0;
  line-height: 1.6;
}
.empty-comments {
  text-align: center;
  padding: 2rem;
  color: #9ca3af;
  font-weight: 500;
}
</style>
