<template>
  <div class="a-page">
    <APageHeader title="发布话题">
      <template #action>
        <ABtn outline @click="router.back()">取消</ABtn>
      </template>
    </APageHeader>

    <div class="form-wrap">
      <!-- Category picker -->
      <div class="a-field">
        <label class="a-field-label">分类</label>
        <div v-if="!forumStore.categoriesLoaded" class="a-muted" style="font-size:0.875rem">
          加载中...
        </div>
        <div v-else-if="forumStore.categories.length === 0" class="a-muted empty-category-note">
          暂无分类，请联系管理员创建分类后再发帖
        </div>
        <div v-else class="category-grid">
          <ABtn
            v-for="cat in forumStore.categories"
            :key="cat.id"
            outline
            size="sm"
            class="category-btn"
            :class="{ 'category-btn-selected': selectedCategoryId === cat.id }"
            @click="selectedCategoryId = cat.id"
          >
            <span class="category-dot" :style="{ background: cat.color || 'var(--a-color-fg)' }" />
            <span>{{ cat.name }}</span>
          </ABtn>
        </div>
        <p v-if="errors.category" class="a-field-error">{{ errors.category }}</p>
      </div>

      <!-- Tag input -->
      <div class="a-field">
        <label class="a-field-label">
          标签
          <span class="field-label-meta">（可选，回车或逗号添加）</span>
        </label>
        <div class="tag-input-wrap">
          <span
            v-for="(tag, i) in tags"
            :key="i"
            class="tag-chip a-badge a-badge-fill"
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
      <div class="a-field">
        <label class="a-field-label">
          标题 &amp; 正文
          <span v-if="draftSavedAt" class="field-label-meta field-label-saved">
            — 草稿已保存 {{ draftSavedAt }}
          </span>
        </label>
        <div class="editor-wrap" :class="{ 'editor-error': !!errors.editor }">
          <AEditor v-model="editorValue" mode="sv" :enable-mentions="true" placeholder="话题标题…" />
        </div>
        <p v-if="errors.editor" class="a-field-error">{{ errors.editor }}</p>
      </div>

      <!-- Submit -->
      <div class="form-actions">
        <ABtn :disabled="submitting" @click="submit">
          {{ submitting ? '发布中...' : '发布话题' }}
        </ABtn>
        <ABtn outline @click="router.back()">取消</ABtn>
        <ABtn v-if="hasDraft" outline size="sm" @click="clearDraft">清除草稿</ABtn>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted, onBeforeUnmount } from 'vue'
import { useRouter } from 'vue-router'
import ABtn from '@/components/ui/ABtn.vue'
import APageHeader from '@/components/ui/APageHeader.vue'
import AEditor from '@/components/shared/AEditor.vue'
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

.field-label-meta {
  font-weight: var(--a-font-weight-normal);
  text-transform: none;
  letter-spacing: 0;
}

.field-label-saved {
  color: var(--a-color-muted);
}

.empty-category-note {
  padding: 0.5rem 0;
  font-size: 0.875rem;
  font-weight: var(--a-font-weight-strong);
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
  box-shadow: none;
}

.category-btn-selected {
  background: var(--a-color-fg);
  color: var(--a-color-bg);
  border-color: var(--a-color-fg);
}

.category-btn-selected .category-dot {
  outline: 2px solid var(--a-color-bg);
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
  border: var(--a-border);
  padding: 0.5rem 0.75rem;
  background: var(--a-color-bg);
  min-height: 2.75rem;
  align-items: center;
  cursor: text;
}

.tag-chip {
  display: inline-flex;
  align-items: center;
  gap: 0.25rem;
  font-size: 0.75rem;
}

.tag-remove {
  background: none;
  border: none;
  color: currentColor;
  cursor: pointer;
  font-size: 1rem;
  line-height: 1;
  padding: 0;
  margin-left: 0.1rem;
}

.tag-input {
  flex: 1;
  min-width: 120px;
  font-size: var(--a-text-sm);
  font-family: inherit;
  background: transparent;
  color: var(--a-color-fg);
}

/* Editor */
.editor-wrap {
  height: 520px;
  display: flex;
  flex-direction: column;
}

.editor-error :deep(.markdown-editor) {
  border-color: var(--a-color-danger);
}

/* Form actions */
.form-actions {
  display: flex;
  gap: 0.75rem;
  align-items: center;
}

</style>
