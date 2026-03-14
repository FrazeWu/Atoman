<template>
  <div class="max-w-4xl mx-auto px-8 py-12 pb-48">
    <!-- Header -->
    <div class="flex items-center justify-between mb-8">
      <h1 class="text-4xl font-black tracking-tighter">
        {{ isEdit ? '编辑文章' : '写文章' }}
      </h1>
      <RouterLink to="/blog" class="text-xs font-black uppercase tracking-widest border-b-2 border-black hover:opacity-60">
        ← 返回
      </RouterLink>
    </div>

    <div class="flex flex-col gap-6">
      <!-- Title -->
      <div>
        <label class="text-xs font-black uppercase tracking-widest text-gray-500 block mb-2">标题 *</label>
        <input
          v-model="form.title"
          placeholder="文章标题..."
          class="w-full border-2 border-black p-4 text-2xl font-black tracking-tight focus:shadow-[5px_5px_0px_0px_rgba(0,0,0,1)] outline-none transition-all"
        />
      </div>

      <!-- Editor -->
      <div>
        <div class="flex items-center justify-between mb-2">
          <label class="text-xs font-black uppercase tracking-widest text-gray-500">正文 *</label>
          <div class="flex gap-2">
            <button
              @click="editorMode = 'write'"
              class="text-xs font-black uppercase tracking-widest px-3 py-1 border border-black transition-all"
              :class="editorMode === 'write' ? 'bg-black text-white' : 'hover:bg-gray-100'"
            >
              编辑
            </button>
            <button
              @click="editorMode = 'preview'"
              class="text-xs font-black uppercase tracking-widest px-3 py-1 border border-black transition-all"
              :class="editorMode === 'preview' ? 'bg-black text-white' : 'hover:bg-gray-100'"
            >
              预览
            </button>
          </div>
        </div>

        <!-- Write mode -->
        <textarea
          v-show="editorMode === 'write'"
          v-model="form.content"
          placeholder="支持 Markdown 语法，开始写作..."
          rows="20"
          class="w-full border-2 border-black p-6 font-mono text-sm leading-relaxed focus:shadow-[5px_5px_0px_0px_rgba(0,0,0,1)] outline-none transition-all resize-none"
        />

        <!-- Preview mode -->
        <div
          v-show="editorMode === 'preview'"
          class="border-2 border-black p-6 min-h-[400px] prose-blog"
          v-html="renderedPreview"
        />
      </div>

      <!-- Optional fields -->
      <details class="border-2 border-black">
        <summary class="p-4 font-black uppercase tracking-widest text-sm cursor-pointer hover:bg-gray-50">
          更多选项
        </summary>
        <div class="p-6 border-t-2 border-black flex flex-col gap-5">
          <!-- Summary -->
          <div>
            <label class="text-xs font-black uppercase tracking-widest text-gray-500 block mb-2">文章摘要（可选）</label>
            <textarea
              v-model="form.summary"
              placeholder="留空将自动截取正文前150字..."
              rows="3"
              class="w-full border-2 border-black p-3 font-medium focus:outline-none resize-none"
            />
          </div>

          <!-- Cover URL -->
          <div>
            <label class="text-xs font-black uppercase tracking-widest text-gray-500 block mb-2">封面图 URL（可选）</label>
            <input
              v-model="form.cover_url"
              placeholder="https://example.com/cover.jpg"
              class="w-full border-2 border-black p-3 font-medium focus:outline-none"
            />
          </div>

          <!-- Allow comments -->
          <label class="flex items-center gap-3 cursor-pointer">
            <input type="checkbox" v-model="form.allow_comments" class="w-5 h-5 cursor-pointer" />
            <span class="font-black text-sm">允许评论</span>
          </label>
        </div>
      </details>

      <!-- Error -->
      <div v-if="error" class="border-2 border-red-500 bg-red-50 p-4 text-red-700 font-bold text-sm">
        {{ error }}
      </div>

      <!-- Action buttons -->
      <div class="flex gap-3 pt-2">
        <button
          @click="save('draft')"
          :disabled="!!saving"
          class="px-6 py-3 font-black uppercase tracking-widest text-sm border-2 border-black hover:bg-black hover:text-white transition-all disabled:opacity-40"
        >
          {{ saving === 'draft' ? '保存中...' : '保存草稿' }}
        </button>
        <button
          @click="save('published')"
          :disabled="!!saving"
          class="flex-1 py-3 font-black uppercase tracking-widest text-sm bg-black text-white border-2 border-black hover:bg-white hover:text-black transition-all disabled:opacity-40"
        >
          {{ saving === 'published' ? '发布中...' : (isEdit ? '更新并发布' : '立即发布') }}
        </button>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import { RouterLink, useRoute, useRouter } from 'vue-router'
import { marked } from 'marked'
import { useAuthStore } from '@/stores/auth'
import { useApi } from '@/composables/useApi'

const route = useRoute()
const router = useRouter()
const authStore = useAuthStore()
const api = useApi()

const isEdit = computed(() => !!route.params.id)
const editorMode = ref<'write' | 'preview'>('write')
const saving = ref<'draft' | 'published' | null>(null)
const error = ref('')

const form = ref({
  title: '',
  content: '',
  summary: '',
  cover_url: '',
  allow_comments: true,
})

const renderedPreview = computed(() => marked(form.value.content || ''))

const loadPost = async () => {
  if (!isEdit.value) return
  try {
    const res = await fetch(api.blog.post(Number(route.params.id)))
    if (res.ok) {
      const d = await res.json()
      const p = d.data || d
      form.value = {
        title: p.title,
        content: p.content,
        summary: p.summary || '',
        cover_url: p.cover_url || '',
        allow_comments: p.allow_comments,
      }
    }
  } catch (e) {
    console.error(e)
  }
}

const save = async (status: 'draft' | 'published') => {
  if (!form.value.title.trim() || !form.value.content.trim()) {
    error.value = '标题和正文不能为空'
    return
  }
  error.value = ''
  saving.value = status

  const payload = { ...form.value, status }
  try {
    let res: Response
    if (isEdit.value) {
      res = await fetch(api.blog.post(Number(route.params.id)), {
        method: 'PUT',
        headers: { 'Content-Type': 'application/json', Authorization: `Bearer ${authStore.token}` },
        body: JSON.stringify(payload)
      })
    } else {
      res = await fetch(api.blog.posts, {
        method: 'POST',
        headers: { 'Content-Type': 'application/json', Authorization: `Bearer ${authStore.token}` },
        body: JSON.stringify(payload)
      })
    }

    if (res.ok) {
      const d = await res.json()
      const savedPost = d.data || d
      router.push(`/blog/posts/${savedPost.id}`)
    } else {
      const err = await res.json()
      error.value = err.error || '保存失败，请重试'
    }
  } catch (e) {
    error.value = '网络错误，请重试'
  } finally {
    saving.value = null
  }
}

onMounted(loadPost)
</script>

<style scoped>
.prose-blog :deep(h1), .prose-blog :deep(h2), .prose-blog :deep(h3) { font-weight: 900; margin: 1.5rem 0 0.75rem; }
.prose-blog :deep(p) { margin: 0.75rem 0; line-height: 1.8; }
.prose-blog :deep(code) { background: #f3f4f6; padding: 0.15em 0.4em; font-size: 0.9em; }
.prose-blog :deep(pre) { background: #111; color: #f8f8f2; padding: 1rem; overflow-x: auto; margin: 1rem 0; }
.prose-blog :deep(pre code) { background: none; padding: 0; }
.prose-blog :deep(blockquote) { border-left: 4px solid black; padding-left: 1rem; margin: 1rem 0; color: #555; }
</style>
