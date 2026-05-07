<template>
  <!-- Full-page editor layout: topbar is 64px, this fills the rest -->
  <div class="editor-page" :class="{ 'toc-open': contentReady && tocOpen }">
    <!-- Top bar: title + meta -->
    <div class="editor-top">
      <div style="display:flex;align-items:center;justify-content:space-between;margin-bottom:1rem">
        <h1 class="a-title-sm" style="margin:0">{{ isEdit ? '编辑文章' : '写文章' }}</h1>
        <RouterLink to="/blog" class="a-link">← 返回</RouterLink>
      </div>

      <input
        v-model="form.title"
        class="editor-title-input"
        placeholder="输入文章标题..."
      />

      <div v-if="error" class="a-error" style="margin-top:.5rem">{{ error }}</div>
    </div>

    <!-- Editor: fills remaining space -->
    <div class="editor-body">
      <div v-if="!isEdit && !selectedChannelId" class="channel-picker-state">
        <div class="channel-picker-card">
          <h2 class="a-subtitle" style="margin-bottom:1rem">先选择合集</h2>
          <p class="a-muted" style="margin-bottom:1.5rem">文章需要从合集入口创建。先进入一个合集，再开始写作。</p>

          <div v-if="loadingChannels" class="channel-picker-grid">
            <div v-for="item in 3" :key="item" class="a-skeleton" style="height:8rem" />
          </div>

          <div v-else-if="channels.length" class="channel-picker-grid">
            <button
              v-for="channel in channels"
              :key="channel.id"
              class="a-card a-card-hover channel-picker-link"
              @click="selectChannel(channel.id)"
            >
              <span class="channel-picker-name">{{ channel.name }}</span>
              <span v-if="channel.description" class="a-muted">{{ channel.description }}</span>
            </button>
          </div>

          <div v-else class="channel-picker-empty">
            <p class="a-muted" style="margin-bottom:1rem">你还没有合集，先创建一个。</p>
          </div>

          <div class="channel-picker-actions">
            <RouterLink to="/blog?create=channel" class="a-btn">+ 新建合集</RouterLink>
            <RouterLink to="/blog" class="a-btn-outline-sm">返回合集页</RouterLink>
          </div>
        </div>
      </div>
      <MarkdownEditor v-else-if="contentReady" v-model="form.content" :hide-title="true" />
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

    <!-- Floating sidebar: TOC + channel controls -->
    <aside v-if="contentReady && tocOpen" class="toc-floating">
      <!-- Channel selector -->
      <div class="sidebar-section">
        <label class="sidebar-label">当前合集</label>
        <select v-model="currentChannelId" class="sidebar-select" @change="onChannelChange">
          <option v-for="ch in channels" :key="ch.id" :value="ch.id">{{ ch.name }}</option>
        </select>
      </div>

      <!-- More options (expanded by default in sidebar) -->
      <div class="sidebar-section">
        <details open>
          <summary class="sidebar-summary">更多选项</summary>
          <div class="sidebar-options">
            <div v-if="channelCollections.length" class="a-field" style="margin-bottom:0.75rem">
              <label class="a-field-label">归档合集</label>
              <div class="editor-collections-grid">
                <label v-for="collection in channelCollections" :key="collection.id" class="editor-collection-item">
                  <input
                    type="checkbox"
                    :checked="selectedCollectionIds.includes(collection.id)"
                    @change="onCollectionToggle(collection.id, $event)"
                  />
                  <span>{{ collection.name }}</span>
                  <span v-if="collection.is_default" class="a-badge a-badge-fill">默认</span>
                </label>
              </div>
            </div>

            <div class="a-field" style="margin-bottom:0.75rem">
              <label class="a-field-label">文章摘要</label>
              <textarea v-model="form.summary" placeholder="留空将自动截取..." rows="2" class="a-textarea" />
            </div>
            <div class="a-field" style="margin-bottom:0.75rem">
              <label class="a-field-label">封面图 URL</label>
              <input v-model="form.cover_url" placeholder="https://..." class="a-input" />
            </div>
            <label style="display:flex;align-items:center;gap:.5rem;cursor:pointer;font-weight:700;font-size:.75rem">
              <input type="checkbox" v-model="form.allow_comments" style="width:1rem;height:1rem" />
              允许评论
            </label>
          </div>
        </details>
      </div>

      <!-- TOC -->
      <div class="sidebar-section">
        <label class="sidebar-label">目录</label>
        <EditorTOC :content="form.content" />
      </div>

      <!-- Upload markdown: show only when empty or imported, hide when manually typed -->
      <div v-if="contentSource !== 'manual'" class="sidebar-section">
        <label class="sidebar-label">导入 Markdown</label>
        <div style="display:flex;align-items:center;gap:.5rem">
          <label v-if="contentSource === 'empty'" class="upload-btn">
            <input type="file" accept=".md,.markdown,.txt" @change="handleFileUpload" style="display:none" />
            <span class="upload-btn-text">选择文件</span>
          </label>
          <template v-else-if="contentSource === 'imported'">
            <button type="button" class="upload-btn" @click="triggerReimport">
              <input ref="fileInput" type="file" accept=".md,.markdown,.txt" @change="handleFileUpload" style="display:none" />
              <span class="upload-btn-text">重新导入</span>
            </button>
            <button type="button" class="a-btn-outline-sm" style="padding:.4rem .6rem;font-size:.7rem" @click="clearContent">
              清空
            </button>
          </template>
          <span v-if="uploading" class="a-muted" style="font-size:.75rem">上传中...</span>
        </div>
      </div>

      <!-- Action buttons -->
      <div class="sidebar-actions">
        <span class="word-count">{{ charCount }} 字 · {{ readingMinutes }} 分钟</span>
        <div style="display:flex;gap:.5rem">
          <ABtn outline size="sm" @click="save('draft')" :disabled="!!saving">
            {{ saving === 'draft' ? '...' : '保存' }}
          </ABtn>
          <ABtn size="sm" @click="save('published')" :disabled="!!saving">
            {{ saving === 'published' ? '...' : '发布' }}
          </ABtn>
        </div>
      </div>
    </aside>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted, defineAsyncComponent, watch } from 'vue'
import { RouterLink, useRoute, useRouter } from 'vue-router'
import ABtn from '@/components/ui/ABtn.vue'
import EditorTOC from '@/components/blog/EditorTOC.vue'
import { useAuthStore } from '@/stores/auth'
import { useApi } from '@/composables/useApi'
import type { Channel, Collection } from '@/types'

const MarkdownEditor = defineAsyncComponent(() => import('@/components/blog/MarkdownEditor.vue'))

const route = useRoute()
const router = useRouter()
const authStore = useAuthStore()
const api = useApi()

const isEdit = computed(() => !!route.params.id)
const contentReady = ref(!route.params.id) // true immediately for new posts
const tocOpen = ref(true)
const channels = ref<Channel[]>([])
const channelCollections = ref<Collection[]>([])
const selectedCollectionIds = ref<string[]>([])
const existingCollectionIds = ref<string[]>([])
const loadingChannels = ref(false)

// Word count stats
const charCount = computed(() => {
  const text = form.value.content.replace(/```[\s\S]*?```/g, '').replace(/[#*`>~_\[\]()]/g, '').trim()
  return text.replace(/\s+/g, '').length
})
const readingMinutes = computed(() => Math.max(1, Math.ceil(charCount.value / 350)))
const saving = ref<'draft' | 'published' | null>(null)
const uploading = ref(false)
const error = ref('')

// Track content source: 'empty' | 'imported' | 'manual'
// - 'empty': no content, show import option
// - 'imported': content from markdown file import, allow re-import or clear
// - 'manual': user manually typed, hide import option
const contentSource = ref<'empty' | 'imported' | 'manual'>('empty')

const selectedChannelId = computed(() => {
  const raw = route.query.channel
  return typeof raw === 'string' && raw ? raw : ''
})
const selectedChannel = computed(() => channels.value.find(channel => String(channel.id) === selectedChannelId.value) || null)

const currentChannelId = ref<string>('')

const form = ref({
  title: '',
  content: '',
  summary: '',
  cover_url: '',
  allow_comments: true,
})

const authHeaders = computed(() => ({ Authorization: `Bearer ${authStore.token}` }))

const selectChannel = (channelId: string) => {
  router.replace({ path: '/post/new', query: { channel: channelId } })
}

const onChannelChange = async () => {
  if (!currentChannelId.value) return
  
  // Update route
  selectChannel(currentChannelId.value)
  
  // Load collections for the selected channel and auto-select default
  await loadChannelCollections()
}

const ensureDefaultSelection = () => {
  const defaultCollection = channelCollections.value.find(collection => collection.is_default)
  if (!defaultCollection) return
  if (!selectedCollectionIds.value.includes(defaultCollection.id)) {
    selectedCollectionIds.value = [defaultCollection.id, ...selectedCollectionIds.value]
  }
}

const loadChannels = async () => {
  if (!authStore.isAuthenticated) return

  loadingChannels.value = true
  try {
    const res = await fetch(`${api.blog.channels}?user_id=${authStore.user?.uuid}`, {
      headers: authHeaders.value,
    })
    if (res.ok) {
      const data = await res.json()
      channels.value = data.data || []

      if (selectedChannelId.value) {
        currentChannelId.value = selectedChannelId.value
      }

      if (selectedChannelId.value && !selectedChannel.value) {
        error.value = '所选合集不存在或无权访问，请重新选择合集'
      }
    }
  } catch (e) {
    console.error(e)
  } finally {
    loadingChannels.value = false
  }
}

const loadChannelCollections = async () => {
  if (!authStore.isAuthenticated || !selectedChannelId.value) {
    channelCollections.value = []
    selectedCollectionIds.value = []
    existingCollectionIds.value = []
    return
  }

  try {
    const res = await fetch(api.blog.channelCollections(selectedChannelId.value), {
      headers: authHeaders.value,
    })
    if (!res.ok) {
      error.value = '加载合集失败'
      return
    }

    const data = await res.json()
    channelCollections.value = data.data || []

    if (!isEdit.value) {
      const defaultCollection = channelCollections.value.find(collection => collection.is_default) || channelCollections.value[0]
      selectedCollectionIds.value = defaultCollection ? [defaultCollection.id] : []
    }

    if (isEdit.value) {
      const allowed = channelCollections.value.map(collection => collection.id)
      selectedCollectionIds.value = existingCollectionIds.value.filter(id => allowed.includes(id))
      ensureDefaultSelection()
    }
  } catch (e) {
    console.error(e)
    error.value = '加载合集失败'
  }
}

const onCollectionToggle = (collectionId: string, event: Event) => {
  const target = event.target as HTMLInputElement | null
  const checked = !!target?.checked

  if (checked) {
    if (!selectedCollectionIds.value.includes(collectionId)) {
      selectedCollectionIds.value = [...selectedCollectionIds.value, collectionId]
    }
    return
  }

  selectedCollectionIds.value = selectedCollectionIds.value.filter(id => id !== collectionId)
  ensureDefaultSelection()
}

const syncPostCollections = async (postId: string) => {
  if (!selectedChannelId.value) return

  const targetIDs = Array.from(new Set(selectedCollectionIds.value))
  const existingIDs = Array.from(new Set(existingCollectionIds.value))
  const toAdd = targetIDs.filter(id => !existingIDs.includes(id))
  const toRemove = existingIDs.filter(id => !targetIDs.includes(id))

  for (const collectionID of toAdd) {
    const res = await fetch(api.blog.postCollections(postId), {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
        ...authHeaders.value,
      },
      body: JSON.stringify({ collection_id: collectionID }),
    })
    if (!res.ok) {
      throw new Error('添加文章合集失败')
    }
  }

  for (const collectionID of toRemove) {
    const res = await fetch(api.blog.postCollection(postId, collectionID), {
      method: 'DELETE',
      headers: authHeaders.value,
    })
    if (!res.ok) {
      throw new Error('移除文章合集失败')
    }
  }

  existingCollectionIds.value = [...targetIDs]
}

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
      form.value = {
        title: p.title,
        content: p.content || '',
        summary: p.summary || '',
        cover_url: p.cover_url || '',
        allow_comments: p.allow_comments,
      }

      // Loaded from server, treat as manual input
      contentSource.value = 'manual'

      if (!selectedChannelId.value) {
        const fallbackChannelID = p.collections?.[0]?.channel_id
        if (fallbackChannelID) {
          await router.replace({
            path: `/post/${String(route.params.id || '')}/edit`,
            query: { channel: fallbackChannelID },
          })
        }
      }

      existingCollectionIds.value = (p.collections || [])
        .filter((collection: Collection) => collection.channel_id === selectedChannelId.value)
        .map((collection: Collection) => collection.id)
      selectedCollectionIds.value = [...existingCollectionIds.value]
    }
  } catch (e) {
    console.error(e)
  } finally {
    contentReady.value = true
  }
}

const save = async (status: 'draft' | 'published') => {
  if (!isEdit.value && !selectedChannelId.value) {
    error.value = '请先选择合集再开始写作'
    return
  }

  if (!form.value.title.trim()) {
    error.value = '请输入文章标题'
    return
  }

  if (!form.value.content.trim()) {
    error.value = '请输入文章内容'
    return
  }

  error.value = ''
  saving.value = status

  const payload = { ...form.value, status }
  const uniqueCollectionIDs = Array.from(new Set(selectedCollectionIds.value))
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
      console.log('[PostEditor] Token:', authStore.token ? 'Present (' + authStore.token.substring(0, 20) + '...)' : 'NULL');
      console.log('[PostEditor] AuthStore:', {
        token: authStore.token ? 'present' : 'missing',
        isAuthenticated: authStore.isAuthenticated
      });
      res = await fetch(api.blog.posts, {
        method: 'POST',
        headers: { 'Content-Type': 'application/json', Authorization: `Bearer ${authStore.token}` },
        body: JSON.stringify({
          ...payload,
          channel_id: selectedChannelId.value,
          collection_ids: uniqueCollectionIDs,
        })
      })
    }

    if (res.ok) {
      const d = await res.json()
      const savedPost = d.data || d

      if (isEdit.value && selectedChannelId.value) {
        await syncPostCollections(String(savedPost.id))
      }

      router.push(`/post/${savedPost.id}`)
    } else {
      const err = await res.json()
      error.value = err.error || '保存失败，请重试'
    }
  } catch (e) {
    error.value = e instanceof Error ? e.message : '网络错误，请重试'
  } finally {
    saving.value = null
  }
}

const handleFileUpload = async (event: Event) => {
  const target = event.target as HTMLInputElement
  const file = target.files?.[0]
  if (!file) return

  uploading.value = true
  try {
    const text = await file.text()
    
    // Try to extract title from first markdown heading
    const lines = text.split('\n')
    let title = ''
    let content = text
    
    for (const line of lines) {
      const trimmed = line.trim()
      if (trimmed.startsWith('# ')) {
        title = trimmed.slice(2).trim()
        break
      }
    }
    
    if (title) {
      // Remove the first heading line from content
      const contentLines = text.split('\n')
      const firstHeadingIndex = contentLines.findIndex(l => l.trim().startsWith('# '))
      if (firstHeadingIndex !== -1) {
        content = contentLines.slice(firstHeadingIndex + 1).join('\n').trim()
      }
    }
    
    form.value.title = title || file.name.replace(/\.(md|markdown|txt)$/i, '')
    form.value.content = content
    contentSource.value = 'imported'
  } catch (e) {
    error.value = '读取文件失败'
    console.error(e)
  } finally {
    uploading.value = false
    target.value = '' // Reset input
  }
}

// Watch content changes to detect manual input
watch(() => form.value.content, (newContent, oldContent) => {
  // If content was empty and now has content, and it's not from import
  if (!oldContent && newContent && contentSource.value === 'empty') {
    // Check if user manually typed in the editor (not imported)
    // We'll detect this by checking if title was also manually entered
    if (form.value.title && !form.value.title.includes('.md')) {
      contentSource.value = 'manual'
    }
  }
})

// Watch title changes to detect manual input
watch(() => form.value.title, (newTitle, oldTitle) => {
  // If title changed from empty to non-empty, and content is already there or user is typing
  if (!oldTitle && newTitle && contentSource.value === 'empty') {
    contentSource.value = 'manual'
  }
})

const clearContent = () => {
  form.value.content = ''
  form.value.title = ''
  contentSource.value = 'empty'
}

const fileInput = ref<HTMLInputElement | null>(null)

const triggerReimport = () => {
  fileInput.value?.click()
}

watch(() => route.query.channel, () => {
  if (!isEdit.value) {
    error.value = ''
  }
})

watch(selectedChannelId, async () => {
  await loadChannelCollections()
})

onMounted(async () => {
  await loadChannels()
  await loadPost()
  await loadChannelCollections()
})
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

.editor-title-input {
  width: 100%;
  padding: 1rem;
  border: 2px solid #000;
  font-size: 1.5rem;
  font-weight: 900;
  letter-spacing: -0.02em;
  outline: none;
  background: #fff;
  font-family: inherit;
}

.editor-title-input::placeholder {
  color: #9ca3af;
}

.editor-title-input:focus {
  box-shadow: 5px 5px 0px 0px rgba(0,0,0,1);
}

.selected-channel-banner {
  display: flex;
  align-items: flex-start;
  justify-content: space-between;
  gap: 1rem;
  border: 2px solid #000;
  padding: 1rem 1.25rem;
  margin-bottom: 0.75rem;
}

.selected-channel-label {
  font-size: 0.7rem;
  font-weight: 900;
  letter-spacing: 0.12em;
  text-transform: uppercase;
  color: #6b7280;
}

.selected-channel-name {
  font-size: 1rem;
  font-weight: 900;
  letter-spacing: -0.02em;
}

.selected-channel-desc {
  margin-top: 0.25rem;
  font-size: 0.875rem;
  color: #6b7280;
}

.editor-body {
  flex: 1;
  min-height: 0;
  display: flex;
  flex-direction: column;
}

.channel-picker-state {
  display: flex;
  align-items: center;
  justify-content: center;
  height: 100%;
  border: 2px solid #000;
  padding: 1.5rem;
}

.channel-picker-card {
  width: min(100%, 56rem);
}

.channel-picker-grid {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(220px, 1fr));
  gap: 1rem;
}

.channel-picker-link {
  display: flex;
  flex-direction: column;
  gap: 0.5rem;
  min-height: 8rem;
  color: #000;
  text-decoration: none;
  cursor: pointer;
  text-align: left;
}

.channel-picker-name {
  font-size: 1rem;
  font-weight: 900;
  letter-spacing: -0.02em;
}

.channel-picker-empty {
  padding: 1rem 0 0.5rem;
}

.channel-picker-actions {
  display: flex;
  gap: 0.75rem;
  align-items: center;
  margin-top: 1.5rem;
}

.editor-collections-grid {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(180px, 1fr));
  gap: 0.75rem;
}

.editor-collection-item {
  display: flex;
  align-items: center;
  gap: 0.5rem;
  border: 2px solid #000;
  padding: 0.6rem 0.75rem;
  font-weight: 700;
}

.editor-collection-item input {
  width: 1rem;
  height: 1rem;
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
  top: 5rem;
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
  top: 5.5rem;
  width: 280px;
  max-height: calc(100vh - 64px - 6rem);
  border: 2px solid #000;
  background: #fff;
  box-shadow: 10px 10px 0px 0px rgba(0,0,0,1);
  overflow-y: auto;
  z-index: 44;
  display: flex;
  flex-direction: column;
}

.sidebar-section {
  padding: 1rem;
  border-bottom: 2px solid #000;
}

.sidebar-section:last-of-type {
  border-bottom: none;
}

.sidebar-label {
  display: block;
  font-size: 0.7rem;
  font-weight: 900;
  letter-spacing: 0.12em;
  text-transform: uppercase;
  color: #6b7280;
  margin-bottom: 0.5rem;
}

.sidebar-select {
  width: 100%;
  padding: 0.6rem 0.75rem;
  border: 2px solid #000;
  font-family: inherit;
  font-size: 0.875rem;
  font-weight: 700;
  background: #fff;
  cursor: pointer;
  outline: none;
  appearance: none;
  background-image: url("data:image/svg+xml,%3Csvg xmlns='http://www.w3.org/2000/svg' width='12' height='12' viewBox='0 0 12 12'%3E%3Cpath fill='%23000' d='M6 8L1 3h10z'/%3E%3C/svg%3E");
  background-repeat: no-repeat;
  background-position: right 0.75rem center;
}

.sidebar-select:focus {
  box-shadow: 3px 3px 0px 0px rgba(0,0,0,1);
}

.sidebar-summary {
  font-size: 0.75rem;
  font-weight: 900;
  text-transform: uppercase;
  letter-spacing: 0.1em;
  cursor: pointer;
  padding: 0.25rem 0;
}

.sidebar-options {
  padding-top: 0.75rem;
  display: flex;
  flex-direction: column;
  gap: 0.75rem;
}

.sidebar-actions {
  padding: 1rem;
  border-top: 2px solid #000;
  display: flex;
  justify-content: space-between;
  align-items: center;
  gap: 0.5rem;
  flex-shrink: 0;
}

.sidebar-actions .word-count {
  margin-right: 0;
}

.upload-btn {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  padding: 0.5rem 1rem;
  border: 2px solid #000;
  background: #fff;
  color: #000;
  font-size: 0.75rem;
  font-weight: 900;
  text-transform: uppercase;
  letter-spacing: 0.1em;
  cursor: pointer;
  transition: all 0.2s;
}

.upload-btn:hover {
  background: #000;
  color: #fff;
}

.upload-btn-text {
  font-family: inherit;
}

@media (max-width: 1100px) {
  .toc-floating,
  .toc-toggle {
    display: none;
  }
}

@media (max-width: 640px) {
  .selected-channel-banner,
  .channel-picker-actions,
  .editor-actions {
    flex-direction: column;
    align-items: stretch;
  }

  .channel-picker-actions > * {
    text-align: center;
  }
}

@media (min-width: 1101px) {
  .editor-page.toc-open {
    padding-right: 21rem;
  }
}
</style>
