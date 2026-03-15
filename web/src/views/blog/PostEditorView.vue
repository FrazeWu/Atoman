<template>
  <!-- Full-page editor layout: topbar is 64px, this fills the rest -->
  <div class="editor-page" :class="{ 'toc-open': contentReady && tocOpen }">
    <!-- Top bar: title + meta -->
    <div class="editor-top">
      <div style="display:flex;align-items:center;justify-content:space-between;margin-bottom:1rem">
        <h1 class="a-title-sm" style="margin:0">{{ isEdit ? '编辑文章' : '写文章' }}</h1>
        <RouterLink to="/blog" class="a-link">← 返回</RouterLink>
      </div>

      <!-- Optional fields (collapsed by default) -->
      <details style="border:2px solid #000">
        <summary style="padding:0.75rem 1rem;font-weight:900;text-transform:uppercase;letter-spacing:.1em;font-size:.75rem;cursor:pointer">
          更多选项
        </summary>
        <div style="padding:1.25rem;border-top:2px solid #000;display:flex;flex-direction:column;gap:1rem">
          <div class="a-field">
            <label class="a-field-label">文章摘要（可选）</label>
            <textarea v-model="form.summary" placeholder="留空将自动截取正文前150字..." rows="2" class="a-textarea" />
          </div>
          <div class="a-field">
            <label class="a-field-label">封面图 URL（可选）</label>
            <input v-model="form.cover_url" placeholder="https://example.com/cover.jpg" class="a-input" />
          </div>
          <label style="display:flex;align-items:center;gap:.75rem;cursor:pointer;font-weight:700;font-size:.875rem">
            <input type="checkbox" v-model="form.allow_comments" style="width:1.25rem;height:1.25rem;cursor:pointer" />
            允许评论
          </label>
        </div>
      </details>

      <div v-if="error" class="a-error" style="margin-top:.5rem">{{ error }}</div>
    </div>

    <!-- Editor: fills remaining space -->
    <div class="editor-body">
      <MarkdownEditor v-if="contentReady" v-model="form.content" />
      <div v-else style="height:100%;display:flex;align-items:center;justify-content:center;border:2px solid #000">
        <span class="a-muted" style="font-size:.875rem">加载中...</span>
      </div>
    </div>

    <button
      v-if="contentReady"
      type="button"
      class="toc-toggle"
      @click="tocOpen = !tocOpen"
    >
      {{ tocOpen ? '隐藏目录' : '显示目录' }}
    </button>

    <aside v-if="contentReady && tocOpen" class="toc-floating">
      <EditorTOC :content="form.content" />
    </aside>

    <!-- Sticky action bar at bottom -->
    <div class="editor-actions">
      <span class="word-count">{{ charCount }} 字 · 约 {{ readingMinutes }} 分钟</span>
      <ABtn outline @click="save('draft')" :disabled="!!saving">
        {{ saving === 'draft' ? '保存中...' : '保存草稿' }}
      </ABtn>
      <ABtn style="flex:1;max-width:200px" @click="save('published')" :disabled="!!saving">
        {{ saving === 'published' ? '发布中...' : (isEdit ? '更新并发布' : '立即发布') }}
      </ABtn>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted, defineAsyncComponent } from 'vue'
import { RouterLink, useRoute, useRouter } from 'vue-router'
import ABtn from '@/components/ui/ABtn.vue'
import EditorTOC from '@/components/blog/EditorTOC.vue'
import { useAuthStore } from '@/stores/auth'
import { useApi } from '@/composables/useApi'

const MarkdownEditor = defineAsyncComponent(() => import('@/components/blog/MarkdownEditor.vue'))

const route = useRoute()
const router = useRouter()
const authStore = useAuthStore()
const api = useApi()

const isEdit = computed(() => !!route.params.id)
const contentReady = ref(!route.params.id) // true immediately for new posts
const tocOpen = ref(true)

// Word count stats
const charCount = computed(() => {
  const text = form.value.content.replace(/```[\s\S]*?```/g, '').replace(/[#*`>~_\[\]()]/g, '').trim()
  return text.replace(/\s+/g, '').length
})
const readingMinutes = computed(() => Math.max(1, Math.ceil(charCount.value / 350)))
const saving = ref<'draft' | 'published' | null>(null)
const error = ref('')

const form = ref({
  title: '',
  content: '',
  summary: '',
  cover_url: '',
  allow_comments: true,
})

const loadPost = async () => {
  if (!isEdit.value) return
  try {
    const postId = String(route.params.id || '')
    if (!postId) return

    const res = await fetch(api.blog.post(postId), {
      headers: authStore.token ? { Authorization: `Bearer ${authStore.token}` } : {},
    })
    if (res.ok) {
      const d = await res.json()
      const p = d.data || d
      let content = p.content || ''
      if (!content.trimStart().startsWith('#')) {
        content = `# ${p.title}\n\n${content}`
      }
      form.value = {
        title: p.title,
        content,
        summary: p.summary || '',
        cover_url: p.cover_url || '',
        allow_comments: p.allow_comments,
      }
    }
  } catch (e) {
    console.error(e)
  } finally {
    contentReady.value = true
  }
}

const save = async (status: 'draft' | 'published') => {
  const rawFirstLine = form.value.content.split('\n')[0].trim()
  const derivedTitle = rawFirstLine.replace(/^#+\s*/, '')
  if (!derivedTitle || !form.value.content.trim()) {
    error.value = '请在第一行写上文章标题'
    return
  }
  error.value = ''
  saving.value = status

  const payload = { ...form.value, title: derivedTitle, status }
  try {
    let res: Response
    if (isEdit.value) {
      const postId = String(route.params.id || '')
      if (!postId) {
        error.value = '文章 ID 无效'
        saving.value = null
        return
      }

      res = await fetch(api.blog.post(postId), {
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
.editor-page {
  display: flex;
  flex-direction: column;
  height: calc(100vh - 64px);
  max-width: 1400px;
  margin: 0 auto;
  padding: 1.25rem 1.5rem 0;
  box-sizing: border-box;
}
.editor-top {
  flex-shrink: 0;
  margin-bottom: 0.75rem;
}
.editor-body {
  flex: 1;
  min-height: 0;
  display: flex;
  flex-direction: column;
}
.editor-body > * {
  flex: 1;
  min-height: 0;
}
.editor-actions {
  flex-shrink: 0;
  display: flex;
  gap: 0.75rem;
  align-items: center;
  padding: 0.75rem 0;
  border-top: 2px solid #000;
  background: #fff;
}

.word-count {
  font-size: 0.75rem;
  font-weight: 700;
  color: #6b7280;
  margin-right: auto;
  user-select: none;
}

.toc-toggle {
  position: fixed;
  right: 1.5rem;
  top: 7.5rem;
  z-index: 45;
  border: 2px solid #000;
  background: #000;
  color: #fff;
  padding: 0.45rem 0.7rem;
  font-size: 0.7rem;
  font-weight: 900;
  text-transform: uppercase;
  letter-spacing: 0.08em;
  cursor: pointer;
  transition: all 0.15s;
}

.toc-toggle:hover {
  background: #fff;
  color: #000;
}

.toc-floating {
  position: fixed;
  right: 1.5rem;
  top: 10rem;
  width: 260px;
  max-height: min(62vh, 520px);
  border: 2px solid #000;
  background: #fff;
  box-shadow: 10px 10px 0px 0px rgba(0,0,0,1);
  overflow: hidden;
  z-index: 44;
}

@media (max-width: 1100px) {
  .toc-floating,
  .toc-toggle {
    display: none;
  }
}

@media (min-width: 1101px) {
  .editor-page.toc-open {
    padding-right: 21rem;
  }
}
</style>
