<template>
  <div class="a-page">
    <APageHeader title="发布话题">
      <template #action>
        <ABtn outline @click="router.back()">取消</ABtn>
      </template>
    </APageHeader>

    <div class="form-wrap">
      <!-- Category picker -->
      <div class="field">
        <label class="field-label">分类</label>
        <div v-if="!forumStore.categoriesLoaded" class="a-muted" style="font-size:0.875rem">
          加载中...
        </div>
        <div v-else-if="forumStore.categories.length === 0" style="font-size:0.875rem;font-weight:700;color:#6b7280;padding:.5rem 0">
          暂无分类，请联系管理员创建分类后再发帖
        </div>
        <div v-else class="category-grid">
          <button
            v-for="cat in forumStore.categories"
            :key="cat.id"
            type="button"
            class="category-btn"
            :class="{ selected: selectedCategoryId === cat.id }"
            @click="selectedCategoryId = cat.id"
          >
            <span class="category-dot" :style="{ background: cat.color || '#000' }" />
            <span>{{ cat.name }}</span>
          </button>
        </div>
        <p v-if="errors.category" class="field-error">{{ errors.category }}</p>
      </div>

      <!-- Tag input -->
      <div class="field">
        <label class="field-label">标签 <span style="font-weight:500;text-transform:none;letter-spacing:0">（可选，回车或逗号添加）</span></label>
        <div class="tag-input-wrap">
          <span
            v-for="(tag, i) in tags"
            :key="i"
            class="tag-chip"
          >
            {{ tag }}
            <button type="button" class="tag-remove" @click="removeTag(i)">×</button>
          </span>
          <input
            ref="tagInputRef"
            v-model="tagInput"
            type="text"
            placeholder="输入标签..."
            class="tag-input"
            @keydown.enter.prevent="addTag"
            @keydown.comma.prevent="addTag"
            @keydown.backspace="onTagBackspace"
          />
        </div>
      </div>

      <!-- Editor (title + content) -->
      <div class="field">
        <label class="field-label">
          标题 &amp; 正文
          <span v-if="draftSavedAt" style="font-weight:500;text-transform:none;letter-spacing:0;color:#6b7280">
            — 草稿已保存 {{ draftSavedAt }}
          </span>
        </label>
        <div class="editor-wrap" :class="{ 'editor-error': !!errors.editor }">
          <MarkdownEditor v-model="editorValue" placeholder="话题标题..." />
        </div>
        <p v-if="errors.editor" class="field-error">{{ errors.editor }}</p>
      </div>

      <!-- Submit -->
      <div class="form-actions">
        <ABtn :disabled="submitting" @click="submit">
          {{ submitting ? '发布中...' : '发布话题' }}
        </ABtn>
        <ABtn outline @click="router.back()">取消</ABtn>
        <button v-if="hasDraft" type="button" class="clear-draft-btn" @click="clearDraft">
          清除草稿
        </button>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted, onBeforeUnmount } from 'vue'
import { useRouter } from 'vue-router'
import ABtn from '@/components/ui/ABtn.vue'
import APageHeader from '@/components/ui/APageHeader.vue'
import MarkdownEditor from '@/components/blog/MarkdownEditor.vue'
import { useForumStore } from '@/stores/forum'
import { useAuthStore } from '@/stores/auth'

const DRAFT_KEY = 'new_topic'

const router = useRouter()
const forumStore = useForumStore()
const authStore = useAuthStore()

const selectedCategoryId = ref('')
const editorValue = ref('')
const tags = ref<string[]>([])
const tagInput = ref('')
const tagInputRef = ref<HTMLInputElement | null>(null)
const submitting = ref(false)
const errors = ref({ category: '', editor: '' })
const draftSavedAt = ref('')
const hasDraft = ref(false)

let autosaveTimer: ReturnType<typeof setInterval> | null = null

// ─── Tag management ──────────────────────────────────────────────────────────

const addTag = () => {
  const val = tagInput.value.replace(/,/g, '').trim()
  if (val && !tags.value.includes(val) && tags.value.length < 5) {
    tags.value.push(val)
  }
  tagInput.value = ''
}

const removeTag = (index: number) => {
  tags.value.splice(index, 1)
}

const onTagBackspace = () => {
  if (tagInput.value === '' && tags.value.length > 0) {
    tags.value.pop()
  }
}

// ─── Draft persistence ───────────────────────────────────────────────────────

const saveDraft = () => {
  const { title, content } = parseEditor()
  if (!title && !content) return
  forumStore.saveDraftLocal(DRAFT_KEY, {
    context_key: DRAFT_KEY,
    title,
    content,
    tags: tags.value.join(','),
  })
  const now = new Date()
  draftSavedAt.value = now.toLocaleTimeString('zh-CN', { hour: '2-digit', minute: '2-digit' })
  hasDraft.value = true
}

const restoreDraft = () => {
  const draft = forumStore.loadDraftLocal(DRAFT_KEY)
  if (!draft) return
  if (draft.content) {
    const titleLine = draft.title ? `# ${draft.title}\n\n` : ''
    editorValue.value = titleLine + (draft.content || '')
  }
  if (draft.tags) {
    tags.value = draft.tags.split(',').filter(Boolean)
  }
  hasDraft.value = true
}

const clearDraft = () => {
  forumStore.clearDraftLocal(DRAFT_KEY)
  hasDraft.value = false
  draftSavedAt.value = ''
}

// Start autosave every 3s while on the page
const startAutosave = () => {
  autosaveTimer = setInterval(saveDraft, 3000)
}

onBeforeUnmount(() => {
  if (autosaveTimer) clearInterval(autosaveTimer)
})

// ─── Submission ──────────────────────────────────────────────────────────────

function parseEditor(): { title: string; content: string } {
  const blocks = editorValue.value.split(/\n\n+/)
  const titleBlock = blocks[0] || ''
  const title = titleBlock.replace(/^#{1,6}\s*/, '').trim()
  const content = blocks.slice(1).join('\n\n').trim()
  return { title, content }
}

const validate = () => {
  let ok = true
  errors.value = { category: '', editor: '' }
  if (!selectedCategoryId.value) {
    errors.value.category = '请选择分类'
    ok = false
  }
  const { title, content } = parseEditor()
  if (!title) {
    errors.value.editor = '标题不能为空'
    ok = false
  } else if (!content) {
    errors.value.editor = '正文不能为空'
    ok = false
  }
  return ok
}

const submit = async () => {
  if (!validate()) return
  submitting.value = true
  try {
    const { title, content } = parseEditor()
    const topic = await forumStore.createTopic({
      category_id: selectedCategoryId.value,
      title,
      content,
      tags: tags.value,
    })
    if (topic) {
      clearDraft()
      router.push(`/topic/${topic.id}`)
    }
  } finally {
    submitting.value = false
  }
}

// ─── Lifecycle ───────────────────────────────────────────────────────────────

onMounted(async () => {
  if (!authStore.isAuthenticated) {
    router.push('/login')
    return
  }
  if (!forumStore.categoriesLoaded) {
    await forumStore.fetchCategories()
  }
  if (forumStore.categories.length > 0 && !selectedCategoryId.value) {
    selectedCategoryId.value = forumStore.categories[0].id
  }
  restoreDraft()
  startAutosave()
})
</script>

<style scoped>
.form-wrap {
  max-width: 860px;
  display: flex;
  flex-direction: column;
  gap: 2rem;
}

.field {
  display: flex;
  flex-direction: column;
  gap: 0.5rem;
}

.field-label {
  font-size: 0.75rem;
  font-weight: 900;
  text-transform: uppercase;
  letter-spacing: 0.1em;
  color: #000;
}

.field-error {
  font-size: 0.8rem;
  font-weight: 700;
  color: #dc2626;
  margin: 0;
}

/* Category grid */
.category-grid {
  display: flex;
  flex-wrap: wrap;
  gap: 0.5rem;
}

.category-btn {
  display: flex;
  align-items: center;
  gap: 0.4rem;
  padding: 0.375rem 0.875rem;
  border: 2px solid #000;
  background: #fff;
  font-weight: 700;
  font-size: 0.875rem;
  cursor: pointer;
  transition: all 0.15s;
}

.category-btn:hover {
  background: #f3f4f6;
}

.category-btn.selected {
  background: #000;
  color: #fff;
}

.category-btn.selected .category-dot {
  outline: 2px solid #fff;
  outline-offset: 1px;
}

.category-dot {
  width: 8px;
  height: 8px;
  border-radius: 9999px;
  flex-shrink: 0;
}

/* Tag input */
.tag-input-wrap {
  display: flex;
  flex-wrap: wrap;
  gap: 0.4rem;
  border: 2px solid #000;
  padding: 0.5rem 0.75rem;
  background: #fff;
  min-height: 2.75rem;
  align-items: center;
  cursor: text;
}

.tag-chip {
  display: inline-flex;
  align-items: center;
  gap: 0.25rem;
  padding: 0.15rem 0.5rem;
  background: #000;
  color: #fff;
  font-size: 0.75rem;
  font-weight: 900;
  letter-spacing: 0.05em;
}

.tag-remove {
  background: none;
  border: none;
  color: #fff;
  cursor: pointer;
  font-size: 1rem;
  line-height: 1;
  padding: 0;
  margin-left: 0.1rem;
}

.tag-input {
  flex: 1;
  min-width: 120px;
  border: none;
  outline: none;
  font-size: 0.875rem;
  font-family: inherit;
  background: transparent;
}

/* Editor */
.editor-wrap {
  height: 520px;
  display: flex;
  flex-direction: column;
}

.editor-error :deep(.markdown-editor) {
  border-color: #dc2626;
}

/* Form actions */
.form-actions {
  display: flex;
  gap: 0.75rem;
  align-items: center;
}

.clear-draft-btn {
  font-size: 0.7rem;
  font-weight: 900;
  text-transform: uppercase;
  letter-spacing: 0.08em;
  color: #9ca3af;
  background: none;
  border: none;
  cursor: pointer;
  padding: 0;
}

.clear-draft-btn:hover {
  color: #ef4444;
}
</style>
