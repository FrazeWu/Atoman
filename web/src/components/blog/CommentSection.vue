<template>
  <div class="mt-12 pt-8 border-t-2 border-black">
    <h3 class="text-2xl font-black tracking-tight mb-6">
      评论 <span class="text-gray-400 font-medium text-lg">({{ comments.length }})</span>
    </h3>

    <!-- Closed comments notice -->
    <div v-if="!allowComments" class="border-2 border-dashed border-gray-300 p-8 text-center text-gray-400 font-medium">
      作者已关闭评论
    </div>

    <template v-else>
      <!-- Comment input (logged in) -->
      <div v-if="authStore.isAuthenticated" class="mb-8">
        <textarea
          v-model="newComment"
          placeholder="写下你的评论..."
          rows="3"
          class="w-full border-2 border-black p-4 font-medium focus:shadow-[5px_5px_0px_0px_rgba(0,0,0,1)] outline-none transition-all resize-none"
        />
        <button
          @click="submitComment"
          :disabled="!newComment.trim() || submitting"
          class="mt-2 bg-black text-white px-6 py-2 font-black uppercase tracking-widest text-xs border-2 border-black hover:bg-white hover:text-black transition-all disabled:opacity-40"
        >
          {{ submitting ? '发送中...' : '发表评论' }}
        </button>
      </div>
      <div v-else class="mb-8 border-2 border-dashed border-gray-300 p-6 text-center text-gray-400">
        <RouterLink to="/login" class="font-black underline hover:opacity-70">登录</RouterLink> 后发表评论
      </div>

      <!-- Loading -->
      <div v-if="loading" class="space-y-4">
        <div v-for="i in 3" :key="i" class="h-20 bg-gray-100 border border-gray-200 animate-pulse rounded" />
      </div>

      <!-- Comment list -->
      <div v-else class="space-y-4">
        <div
          v-for="comment in comments"
          :key="comment.id"
          class="border-l-4 border-black pl-6 py-3"
        >
          <div class="flex items-center gap-3 mb-2">
            <div class="w-7 h-7 rounded-full bg-black flex items-center justify-center text-white text-xs font-black">
              {{ (comment.user?.display_name || comment.user?.username || '?').charAt(0).toUpperCase() }}
            </div>
            <span class="font-black text-sm">{{ comment.user?.display_name || comment.user?.username }}</span>
            <span class="text-xs text-gray-400 font-medium">{{ formatDate(comment.created_at) }}</span>
            <button
              v-if="canDelete(comment)"
              @click="deleteComment(comment.id)"
              class="ml-auto text-xs text-red-500 hover:text-red-700 font-bold"
            >
              删除
            </button>
          </div>
          <p class="text-gray-800 font-medium whitespace-pre-wrap">{{ comment.content }}</p>
        </div>

        <div v-if="!comments.length" class="text-gray-400 font-medium text-center py-8">
          还没有评论，来发表第一条吧
        </div>
      </div>
    </template>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { RouterLink } from 'vue-router'
import { useAuthStore } from '@/stores/auth'
import { useApi } from '@/composables/useApi'
import type { Comment } from '@/types'

const props = defineProps<{
  postId: number
  allowComments: boolean
  postOwnerId?: number
}>()

const authStore = useAuthStore()
const api = useApi()

const comments = ref<Comment[]>([])
const loading = ref(true)
const newComment = ref('')
const submitting = ref(false)

const formatDate = (d: string) => new Date(d).toLocaleDateString('zh-CN')

const canDelete = (c: Comment) => {
  if (!authStore.user) return false
  return authStore.user.id === c.user_id || authStore.user.id === props.postOwnerId || authStore.user.role === 'admin'
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

const deleteComment = async (id: number) => {
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

onMounted(fetchComments)
</script>
