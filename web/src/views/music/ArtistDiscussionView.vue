<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import { useRoute, RouterLink } from 'vue-router'
import { useAuthStore } from '@/stores/auth'
import AEditor from '@/components/shared/AEditor.vue'

const route = useRoute()
const authStore = useAuthStore()
const API_URL = import.meta.env.VITE_API_URL || '/api'

const artistId = computed(() => route.params.artistId as string)
const discussions = ref<any[]>([])
const loading = ref(true)
const error = ref('')
const total = ref(0)

const newContent = ref('')
const submitting = ref(false)
const replyingTo = ref<string | null>(null)
const replyContent = ref('')

const fetchDiscussions = async () => {
  loading.value = true
  error.value = ''
  try {
    const res = await fetch(`${API_URL}/artists/${artistId.value}/discussions?limit=50`, {
      headers: authStore.token ? { Authorization: `Bearer ${authStore.token}` } : {}
    })
    const data = await res.json()
    discussions.value = data.data || []
    total.value = data.total || 0
  } catch (e: any) {
    error.value = e.message || '加载失败'
  } finally {
    loading.value = false
  }
}

const createDiscussion = async () => {
  if (!newContent.value.trim()) return
  submitting.value = true
  try {
    await fetch(`${API_URL}/artists/${artistId.value}/discussions`, {
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
    alert(e.message || '发布失败')
  } finally {
    submitting.value = false
  }
}

const replyTo = async (discussionId: string) => {
  if (!replyContent.value.trim()) return
  try {
    await fetch(`${API_URL}/artists/${artistId.value}/discussions/${discussionId}/reply`, {
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
    alert(e.message || '回复失败')
  }
}

const deleteDiscussion = async (discussionId: string) => {
  if (!confirm('确定删除这条讨论？')) return
  try {
    await fetch(`${API_URL}/artists/${artistId.value}/discussions/${discussionId}`, {
      method: 'DELETE',
      headers: { Authorization: `Bearer ${authStore.token}` }
    })
    fetchDiscussions()
  } catch (e: any) {
    alert(e.message || '删除失败')
  }
}

const formatDate = (date: string) => {
  return new Date(date).toLocaleString('zh-CN', {
    year: 'numeric', month: 'short', day: 'numeric',
    hour: '2-digit', minute: '2-digit'
  })
}

const canDelete = (discussion: any) => {
  return authStore.isAuthenticated && (
    authStore.user?.role === 'admin' ||
    discussion.user_id === authStore.user?.id
  )
}

onMounted(() => fetchDiscussions())
</script>

<template>
  <div class="page-container">
    <RouterLink :to="`/music/artists/${artistId}`" class="back-link">
      ← 返回艺术家
    </RouterLink>

    <div class="header">
      <h1 class="title">讨论</h1>
      <p class="subtitle">共 {{ total }} 条讨论</p>
    </div>

    <div v-if="authStore.isAuthenticated" class="new-discussion">
      <h3 class="section-title">发起讨论</h3>
      <AEditor
        v-model="newContent"
        mode="sv"
        :enable-mentions="true"
        placeholder="写下你的想法…（支持 Markdown，@ 提及用户）"
      />
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

    <div v-if="loading" class="loading">加载中...</div>
    <div v-else-if="error" class="error">{{ error }}</div>
    <div v-else class="discussion-list">
      <div
        v-for="discussion in discussions"
        :key="discussion.id"
        class="discussion-item"
      >
        <div class="discussion-header">
          <div class="user-info">
            <div class="avatar">{{ discussion.user?.username?.[0]?.toUpperCase() || '?' }}</div>
            <div class="username">{{ discussion.user?.username || '未知用户' }}</div>
          </div>
          <div class="meta">
            <span class="date">{{ formatDate(discussion.created_at) }}</span>
            <button
              v-if="canDelete(discussion)"
              @click="deleteDiscussion(discussion.id)"
              class="btn-delete"
            >删除</button>
          </div>
        </div>
        <div class="discussion-content" v-html="discussion.content"></div>
        <button
          v-if="authStore.isAuthenticated"
          @click="replyingTo = discussion.id"
          class="btn-reply"
        >回复</button>
        <div v-if="replyingTo === discussion.id" class="reply-form">
          <AEditor
            v-model="replyContent"
            mode="sv"
            :enable-mentions="true"
            placeholder="写下你的回复…"
          />
          <div class="reply-actions">
            <button @click="replyingTo = null" class="btn-cancel">取消</button>
            <button
              @click="replyTo(discussion.id)"
              class="btn-submit"
              :disabled="!replyContent.trim()"
            >回复</button>
          </div>
        </div>
        <div v-if="discussion.replies?.length" class="replies">
          <div
            v-for="reply in discussion.replies"
            :key="reply.id"
            class="reply-item"
          >
            <div class="reply-header">
              <div class="user-info">
                <div class="avatar small">{{ reply.user?.username?.[0]?.toUpperCase() || '?' }}</div>
                <div class="username">{{ reply.user?.username || '未知用户' }}</div>
              </div>
              <span class="date">{{ formatDate(reply.created_at) }}</span>
            </div>
            <div class="reply-content">{{ reply.content }}</div>
          </div>
        </div>
      </div>
      <div v-if="discussions.length === 0" class="empty">暂无讨论，快来发起第一个话题吧！</div>
    </div>
  </div>
</template>

<style scoped>
.page-container {
  max-width: 900px;
  margin: 0 auto;
  padding: 2rem;
  padding-bottom: 10rem;
}
.back-link {
  display: inline-block;
  margin-bottom: 2rem;
  font-weight: 700;
  text-decoration: none;
  color: var(--a-color-fg);
}
.back-link:hover { text-decoration: underline; }
.header { margin-bottom: 2rem; }
.title { font-size: 2rem; font-weight: 900; margin: 0 0 0.25rem; }
.subtitle { color: var(--a-color-muted); margin: 0; }
.section-title {
  font-size: 0.75rem;
  font-weight: 900;
  text-transform: uppercase;
  letter-spacing: 0.1em;
  margin: 0 0 1rem;
}
.new-discussion {
  border: var(--a-border);
  padding: 1.5rem;
  margin-bottom: 2rem;
}
.login-prompt {
  border: var(--a-border);
  padding: 1rem;
  margin-bottom: 2rem;
  color: var(--a-color-muted);
}
.login-prompt .link { font-weight: 700; color: var(--a-color-fg); }
.btn-submit {
  margin-top: 0.75rem;
  padding: 0.5rem 1.25rem;
  background: var(--a-color-fg);
  color: var(--a-color-bg);
  border: none;
  font-weight: 700;
  cursor: pointer;
}
.btn-submit:disabled { opacity: 0.5; cursor: not-allowed; }
.loading, .error { padding: 2rem; text-align: center; color: var(--a-color-muted); }
.discussion-list { display: flex; flex-direction: column; gap: 1.5rem; }
.discussion-item {
  border: var(--a-border);
  padding: 1.5rem;
}
.discussion-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 0.75rem;
}
.user-info { display: flex; align-items: center; gap: 0.5rem; }
.avatar {
  width: 2rem; height: 2rem;
  display: flex; align-items: center; justify-content: center;
  background: var(--a-color-fg);
  color: var(--a-color-bg);
  font-weight: 700;
  font-size: 0.875rem;
}
.avatar.small { width: 1.5rem; height: 1.5rem; font-size: 0.75rem; }
.username { font-weight: 700; font-size: 0.875rem; }
.meta { display: flex; align-items: center; gap: 0.75rem; }
.date { font-size: 0.75rem; color: var(--a-color-muted); }
.btn-delete {
  font-size: 0.75rem;
  color: #dc2626;
  background: none;
  border: none;
  cursor: pointer;
}
.discussion-content { margin-bottom: 0.75rem; line-height: 1.6; }
.btn-reply {
  font-size: 0.75rem;
  font-weight: 700;
  background: none;
  border: none;
  cursor: pointer;
  color: var(--a-color-muted);
}
.reply-form { margin-top: 0.75rem; border-left: 3px solid var(--a-color-fg); padding-left: 1rem; }
.reply-actions { display: flex; gap: 0.5rem; margin-top: 0.5rem; }
.btn-cancel {
  padding: 0.375rem 0.75rem;
  background: none;
  border: var(--a-border);
  cursor: pointer;
  font-size: 0.875rem;
}
.replies { margin-top: 1rem; display: flex; flex-direction: column; gap: 0.75rem; }
.reply-item { border-left: 3px solid var(--a-color-border, #e5e7eb); padding-left: 1rem; }
.reply-header { display: flex; justify-content: space-between; margin-bottom: 0.5rem; }
.reply-content { font-size: 0.875rem; line-height: 1.5; }
.empty { color: var(--a-color-muted); padding: 2rem; text-align: center; }
</style>
