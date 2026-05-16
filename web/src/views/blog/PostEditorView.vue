<template>
  <div class="editor-page">
    <div class="editor-shell">
      <div v-if="error" class="editor-error a-error">{{ error }}</div>

      <div class="editor-layout">
        <!-- 主编辑区 -->
        <main class="col-center a-card-sm">
          <div v-if="contentReady" class="editor-workspace">
            <div class="editor-meta-row">
              <div class="meta-chip-group">
                <span class="meta-chip">{{ isEdit ? '编辑文章' : '新建文章' }}</span>
              </div>

              <div class="editor-meta-actions">
                <span class="draft-status" :class="`is-${draftStatus.tone}`">{{ draftStatus.text }}</span>

                <div v-if="contentSource !== 'manual'" class="import-actions">
                  <ABtn v-if="contentSource === 'empty'" tag="label" variant="secondary" size="sm">
                    <input type="file" accept=".md,.markdown,.txt" @change="handleFileUpload" class="hidden-file-input" />
                    导入 Markdown
                  </ABtn>
                  <template v-else-if="contentSource === 'imported'">
                    <ABtn type="button" variant="secondary" size="sm" @click="triggerReimport">
                      <input ref="fileInput" type="file" accept=".md,.markdown,.txt" @change="handleFileUpload" class="hidden-file-input" />
                      重新导入
                    </ABtn>
                    <ABtn type="button" variant="ghost" size="sm" @click="clearContent">清空内容</ABtn>
                  </template>
                  <span v-if="uploading" class="a-muted">读取中…</span>
                </div>
              </div>
            </div>

            <section class="editor-canvas">
              <div class="title-sticky">
                <textarea
                  ref="titleRef"
                  v-model="form.title"
                  class="title-input"
                  placeholder="文章标题…"
                  rows="1"
                  @input="autoResizeTitle"
                />
                <div class="title-divider" />
              </div>
              <div class="editor-body">
                <AEditor
                  v-model="form.content"
                  mode="sv"
                  :no-border="true"
                  :enable-embeds="true"
                  :enable-collab="isEdit"
                  :collab-room-id="isEdit ? String(route.params.id || '') : undefined"
                />
              </div>
            </section>
          </div>

          <div v-else class="editor-loading">加载中…</div>
        </main>

        <!-- 右侧面板 -->
        <aside class="col-right a-card-sm">
          <!-- 发布操作 -->
          <section class="right-section publish-section">
            <span class="a-label">发布</span>
            <div class="publish-actions">
              <ABtn
                variant="secondary"
                block
                :loading="saving === 'draft'"
                :disabled="!!saving"
                loading-text="保存中…"
                @click="save('draft')"
              >
                存草稿
              </ABtn>
              <ABtn
                variant="primary"
                block
                :loading="saving === 'published'"
                :disabled="!!saving"
                loading-text="发布中…"
                @click="save('published')"
              >
                发布文章
              </ABtn>
            </div>
            <ABtn v-if="hasDraftManagerAccess" type="button" variant="ghost" size="sm" block @click="openDraftManager">草稿管理</ABtn>
          </section>

          <!-- 字数统计 -->
          <section class="right-section stat-section">
            <span class="a-label">统计</span>
            <div class="stat-grid">
              <div class="stat-card">
                <span class="stat-num">{{ charCount }}</span>
                <span class="stat-unit">字数</span>
              </div>
              <div class="stat-card">
                <span class="stat-num">{{ readingMinutes }}</span>
                <span class="stat-unit">分钟</span>
              </div>
            </div>
          </section>

          <!-- 文章设置 -->
          <section class="right-section">
            <span class="a-label">文章设置</span>
            <div class="options-list">
              <div class="a-field">
                <label class="a-field-label">文章摘要</label>
                <textarea
                  v-model="form.summary"
                  placeholder="留空自动截取…"
                  rows="3"
                  class="a-textarea"
                />
              </div>
              <div class="a-field">
                <label class="a-field-label">封面图</label>
                <input
                  ref="coverInput"
                  type="file"
                  accept="image/jpeg,image/png,image/gif,image/webp"
                  class="hidden-file-input"
                  @change="handleCoverUpload"
                />
                <div class="cover-upload-card a-card-sm">
                  <div v-if="form.cover_url" class="cover-preview-wrap">
                    <img :src="form.cover_url" alt="封面预览" class="cover-preview-image" />
                  </div>
                  <div v-else class="cover-empty-state">
                    <strong>未上传封面</strong>
                    <p class="a-muted">上传后会自动填入文章封面地址</p>
                  </div>

                  <div class="cover-upload-actions">
                    <ABtn
                      type="button"
                      variant="secondary"
                      size="sm"
                      :loading="coverUploading"
                      :disabled="coverUploading"
                      loading-text="上传中…"
                      @click="triggerCoverUpload"
                    >
                      {{ form.cover_url ? '更换封面' : '上传封面' }}
                    </ABtn>
                    <ABtn
                      v-if="form.cover_url"
                      type="button"
                      variant="ghost"
                      size="sm"
                      :disabled="coverUploading"
                      @click="removeCover"
                    >
                      移除封面
                    </ABtn>
                  </div>

                  <p class="cover-upload-hint a-muted">支持 JPEG、PNG、GIF、WebP，最大 5MB。</p>
                  <p v-if="coverUploadError" class="cover-upload-error">{{ coverUploadError }}</p>
                </div>
              </div>
              <label class="option-check a-card-sm">
                <input type="checkbox" v-model="form.allow_comments" />
                <span>允许评论</span>
              </label>
            </div>
          </section>

          <!-- 目录 -->
          <section class="right-section toc-section">
            <div class="section-heading-row">
              <span class="a-label">目录</span>
              <span class="a-muted">{{ toc.length }} 个标题</span>
            </div>
            <div v-if="toc.length === 0" class="col-empty">加入 Markdown 标题后显示</div>
            <nav v-else class="toc-list">
              <a
                v-for="(item, i) in toc"
                :key="i"
                class="toc-item"
                :class="`toc-h${item.level}`"
                href="#"
                @click.prevent
              >
                <span class="toc-rail" aria-hidden="true" />
                <span class="toc-text">{{ item.text }}</span>
              </a>
            </nav>
          </section>
        </aside>
      </div>
    </div>

    <AModal v-if="recoveryModalVisible && pendingDraftCandidate" title="发现未恢复草稿" size="md" @close="keepCurrentContent">
      <div class="draft-recovery-body">
        <span class="a-label">{{ pendingDraftCandidate.source === 'server' ? '云端草稿' : '本地草稿' }}</span>
        <p class="draft-recovery-text">
          检测到一份较新的{{ pendingDraftCandidate.source === 'server' ? '云端' : '本地' }}草稿，保存于 {{ formatSavedTime(pendingDraftCandidate.savedAt) }}。
          恢复后会覆盖当前编辑区内容。
        </p>

        <div class="draft-recovery-preview a-card-sm">
          <strong>{{ pendingDraftCandidate.payload.title || '未命名草稿' }}</strong>
          <p class="a-muted">{{ draftRecoveryPreview }}</p>
        </div>
      </div>

      <template #footer>
        <div class="draft-recovery-actions">
          <ABtn type="button" variant="secondary" @click="keepCurrentContent">稍后处理</ABtn>
          <ABtn type="button" variant="ghost" @click="discardPendingDraft">丢弃草稿</ABtn>
          <ABtn type="button" variant="primary" @click="restorePendingDraft">恢复草稿</ABtn>
        </div>
      </template>
    </AModal>

    <AModal v-if="draftManagerVisible" title="草稿管理" size="md" @close="closeDraftManager">
      <div class="draft-manager-body">
        <div class="draft-manager-grid">
          <div class="draft-manager-card a-card-sm">
            <span class="a-label">本地草稿</span>
            <strong>{{ localDraftStatusText }}</strong>
            <p class="a-muted">保存在当前浏览器中，刷新页面后仍可恢复。</p>
          </div>

          <div class="draft-manager-card a-card-sm">
            <span class="a-label">云端草稿</span>
            <strong>{{ cloudDraftStatusText }}</strong>
            <p class="a-muted">登录状态下自动同步，可在其他会话中继续写作。</p>
          </div>
        </div>

        <div v-if="deferredDraftCandidate" class="draft-manager-card draft-manager-card-accent a-card-sm">
          <span class="a-label">待恢复草稿</span>
          <strong>{{ deferredDraftCandidate.payload.title || '未命名草稿' }}</strong>
          <p class="a-muted">{{ deferredDraftCandidate.source === 'server' ? '云端' : '本地' }}版本，保存于 {{ formatSavedTime(deferredDraftCandidate.savedAt) }}</p>
          <p class="draft-manager-preview">{{ deferredDraftSummary }}</p>
        </div>

        <div v-if="serverDraftState === 'error'" class="draft-manager-warning">
          云端草稿同步失败，当前变更仍保存在本地。你可以稍后重试同步，或继续在当前会话中编辑。
        </div>
      </div>

      <template #footer>
        <div class="draft-recovery-actions">
          <ABtn type="button" variant="secondary" @click="closeDraftManager">关闭</ABtn>
          <ABtn
            v-if="authStore.token && hasMeaningfulDraft(draftPayload)"
            type="button"
            variant="secondary"
            @click="syncDraftNow"
          >
            立即同步
          </ABtn>
          <ABtn
            v-if="hasDraftManagerAccess"
            type="button"
            variant="ghost"
            @click="clearSavedDrafts"
          >
            清除已保存草稿
          </ABtn>
          <ABtn
            v-if="deferredDraftCandidate"
            type="button"
            variant="primary"
            @click="restoreDeferredFromManager"
          >
            恢复最新草稿
          </ABtn>
        </div>
      </template>
    </AModal>

    <AModal v-if="leaveConfirmVisible" title="草稿仍在同步" size="sm" @close="cancelLeave">
      <div class="leave-confirm-body">
        <p class="leave-confirm-text">{{ leaveConfirmText }}</p>
        <p class="a-muted">继续离开会中断当前这次保存或同步，最新改动可能无法写入草稿。</p>
      </div>

      <template #footer>
        <div class="draft-recovery-actions">
          <ABtn type="button" variant="secondary" @click="cancelLeave">留在此页</ABtn>
          <ABtn type="button" variant="primary" @click="confirmLeave">继续离开</ABtn>
        </div>
      </template>
    </AModal>
  </div>
</template>

<script setup lang="ts">
import { computed, onBeforeUnmount, onMounted, ref, watch, nextTick } from 'vue'
import { onBeforeRouteLeave, useRoute, useRouter } from 'vue-router'

import AEditor from '@/components/shared/AEditor.vue'
import { useAutoSave } from '@/composables/useAutoSave'
import ABtn from '@/components/ui/ABtn.vue'
import AModal from '@/components/ui/AModal.vue'
import { useApi } from '@/composables/useApi'
import { useAuthStore } from '@/stores/auth'
import type { BlogDraft, Channel, Collection } from '@/types'

const route = useRoute()
const router = useRouter()
const authStore = useAuthStore()
const api = useApi()

// ── 布局 ─────────────────────────────────────────────────
type DraftSyncState = 'idle' | 'syncing' | 'synced' | 'error'
type SaveTarget = 'draft' | 'published'
type EditorDraftPayload = {
  context_key: string
  source_post_id?: string
  title: string
  content: string
  summary: string
  cover_url: string
  allow_comments: boolean
  channel_id?: string
  collection_ids: string[]
}
type DraftCandidate = {
  source: 'local' | 'server'
  payload: EditorDraftPayload
  savedAt: number
}

const saving = ref<'draft' | 'published' | null>(null)

// ── 状态 ─────────────────────────────────────────────────
const isEdit = computed(() => !!route.params.id)
const contentReady = ref(!route.params.id)
const channels = ref<Channel[]>([])
const channelCollections = ref<Collection[]>([])
const selectedCollectionIds = ref<string[]>([])
const existingCollectionIds = ref<string[]>([])
const loadingChannels = ref(false)
const uploading = ref(false)
const coverUploading = ref(false)
const error = ref('')
const coverUploadError = ref('')
const titleRef = ref<HTMLTextAreaElement | null>(null)
const fileInput = ref<HTMLInputElement | null>(null)
const coverInput = ref<HTMLInputElement | null>(null)
const contentSource = ref<'empty' | 'imported' | 'manual'>('empty')
const rightPanelMode = ref('outline')
const loadedPostUpdatedAt = ref(0)
const recoveryModalVisible = ref(false)
const pendingDraftCandidate = ref<DraftCandidate | null>(null)
const deferredDraftCandidate = ref<DraftCandidate | null>(null)
const draftManagerVisible = ref(false)
const leaveConfirmVisible = ref(false)
const serverDraftState = ref<DraftSyncState>('idle')
const serverDraftSavedAt = ref<number | null>(null)
const draftWatchEnabled = ref(false)
const isApplyingDraft = ref(false)
const pendingLeavePath = ref<string | null>(null)
const allowRouteLeaveOnce = ref(false)

const form = ref({
  title: '',
  content: '',
  summary: '',
  cover_url: '',
  allow_comments: true,
})

let serverSyncTimer: ReturnType<typeof setTimeout> | null = null

// ── 字数统计 ─────────────────────────────────────────────
const charCount = computed(() => {
  const text = form.value.content
    .replace(/```[\s\S]*?```/g, '')
    .replace(/[#*`>~_\[\]()]/g, '')
    .trim()
  return text.replace(/\s+/g, '').length
})
const readingMinutes = computed(() => Math.max(1, Math.ceil(charCount.value / 350)))

// ── 目录提取 ─────────────────────────────────────────────
const toc = computed(() => {
  const lines = form.value.content.split('\n')
  const items: { level: number; text: string }[] = []
  for (const line of lines) {
    const m = line.match(/^(#{1,4})\s+(.+)/)
    if (m) items.push({ level: m[1].length, text: m[2].trim() })
  }
  return items
})

// ── 合集 ─────────────────────────────────────────────────
const selectedChannelId = computed(() => {
  const raw = route.query.channel
  return typeof raw === 'string' && raw ? raw : ''
})

const currentChannelId = ref<string>('')

const authHeaders = computed(() => {
  const headers: Record<string, string> = {}
  if (authStore.token) {
    headers.Authorization = `Bearer ${authStore.token}`
  }
  return headers
})
const editorRoutePath = computed(() => isEdit.value ? `/post/${String(route.params.id || '')}/edit` : '/post/new')
const draftContextKey = computed(() => isEdit.value ? `blog:post:${String(route.params.id || '')}` : 'blog:new')
const draftPayload = computed<EditorDraftPayload>(() => ({
  context_key: draftContextKey.value,
  source_post_id: isEdit.value ? String(route.params.id || '') : undefined,
  title: form.value.title,
  content: form.value.content,
  summary: form.value.summary,
  cover_url: form.value.cover_url,
  allow_comments: form.value.allow_comments,
  channel_id: currentChannelId.value || selectedChannelId.value || undefined,
  collection_ids: Array.from(new Set(selectedCollectionIds.value)),
}))

const {
  autoSaveState,
  lastSavedAt,
  triggerAutoSave,
  loadDraft,
  clearDraft: clearLocalDraft,
} = useAutoSave<EditorDraftPayload>({
  getDraftKey: () => `blog_editor_${draftContextKey.value}`,
  getPayload: () => draftPayload.value,
  shouldPersist: (payload) => hasMeaningfulDraft(payload),
})

const draftStatus = computed(() => {
  if (saving.value === 'published') {
    return { tone: 'warn', text: '发布中…' }
  }
  if (saving.value === 'draft') {
    return { tone: 'warn', text: '草稿保存中…' }
  }
  if (serverDraftState.value === 'error') {
    return { tone: 'warn', text: '云端草稿同步失败，当前仅保存在本地' }
  }
  if (autoSaveState.value === 'saving' || serverDraftState.value === 'syncing') {
    return { tone: 'warn', text: '草稿同步中…' }
  }
  if (deferredDraftCandidate.value) {
    return { tone: 'warn', text: '检测到可恢复草稿' }
  }
  if (lastSavedAt.value || serverDraftSavedAt.value) {
    const labels = []
    if (lastSavedAt.value) labels.push(`本地 ${formatSavedTime(lastSavedAt.value)}`)
    if (serverDraftSavedAt.value) labels.push(`云端 ${formatSavedTime(serverDraftSavedAt.value)}`)
    return { tone: 'ok', text: `草稿已保存 · ${labels.join(' · ')}` }
  }
  return { tone: 'muted', text: '开始写作后会自动保存草稿' }
})

const draftRecoveryPreview = computed(() => {
  const candidate = pendingDraftCandidate.value
  if (!candidate) return ''

  const sourceText = candidate.payload.summary || candidate.payload.content
  const plainText = sourceText.replace(/[#*_>`~\-\[\]()]/g, ' ').replace(/\s+/g, ' ').trim()
  return plainText || '这份草稿还没有正文预览。'
})

const deferredDraftSummary = computed(() => {
  const candidate = deferredDraftCandidate.value
  if (!candidate) return ''

  const sourceText = candidate.payload.summary || candidate.payload.content
  const plainText = sourceText.replace(/[#*_>`~\-\[\]()]/g, ' ').replace(/\s+/g, ' ').trim()
  return plainText || '这份草稿还没有正文预览。'
})

const hasDraftManagerAccess = computed(() => (
  !!deferredDraftCandidate.value
  || !!lastSavedAt.value
  || !!serverDraftSavedAt.value
  || hasMeaningfulDraft(draftPayload.value)
  || serverDraftState.value === 'error'
))

const localDraftStatusText = computed(() => {
  if (lastSavedAt.value) {
    return `最近保存于 ${formatSavedTime(lastSavedAt.value)}`
  }
  if (autoSaveState.value === 'saving') {
    return '正在保存中…'
  }
  if (hasMeaningfulDraft(draftPayload.value)) {
    return '已有编辑内容，等待下一次自动保存'
  }
  return '暂无本地草稿'
})

const cloudDraftStatusText = computed(() => {
  if (!authStore.token) {
    return '未登录，未启用云端草稿'
  }
  if (serverDraftState.value === 'syncing') {
    return '正在同步到云端…'
  }
  if (serverDraftState.value === 'error') {
    return '同步失败，等待手动重试'
  }
  if (serverDraftSavedAt.value) {
    return `最近同步于 ${formatSavedTime(serverDraftSavedAt.value)}`
  }
  if (hasMeaningfulDraft(draftPayload.value)) {
    return '尚未生成云端草稿'
  }
  return '暂无云端草稿'
})

const leaveConfirmText = computed(() => {
  if (saving.value === 'published') {
    return '文章正在发布中，离开后可能无法确认本次发布结果。'
  }
  if (saving.value === 'draft') {
    return '文章正在保存草稿，离开后本次保存可能无法完成。'
  }
  if (serverDraftState.value === 'syncing') {
    return '云端草稿仍在同步中，离开后最新改动可能只保留在本地。'
  }
  return '本地草稿仍在写入中，离开后最新改动可能不会进入已保存草稿。'
})

const currentDraftSignature = computed(() => JSON.stringify(draftPayload.value))
const hasPendingPersistence = computed(() => (
  autoSaveState.value === 'saving' || serverDraftState.value === 'syncing' || !!saving.value
))

const selectChannel = (channelId: string) => {
  router.replace({ path: '/post/new', query: { channel: channelId } })
}

const onChannelChange = async () => {
  if (!currentChannelId.value) return
  selectChannel(currentChannelId.value)
  await loadChannelCollections()
}

const ensureDefaultSelection = () => {
  const def = channelCollections.value.find(c => c.is_default)
  if (def && !selectedCollectionIds.value.includes(def.id)) {
    selectedCollectionIds.value = [def.id, ...selectedCollectionIds.value]
  }
}

const hasMeaningfulDraft = (payload: EditorDraftPayload) => {
  return Boolean(
    payload.title.trim()
    || payload.content.trim()
    || payload.summary.trim()
    || payload.cover_url.trim()
    || payload.channel_id
    || payload.collection_ids.length
  )
}

const parseTimestamp = (value?: string | null) => {
  if (!value) return 0
  const parsed = Date.parse(value)
  return Number.isNaN(parsed) ? 0 : parsed
}

const formatSavedTime = (value?: number | null) => {
  if (!value) return '--:--'
  return new Date(value).toLocaleTimeString('zh-CN', { hour: '2-digit', minute: '2-digit' })
}

const triggerCoverUpload = () => {
  coverInput.value?.click()
}

const handleCoverUpload = async (event: Event) => {
  const input = event.target as HTMLInputElement
  const file = input.files?.[0]
  input.value = ''

  if (!file) return

  if (!authStore.token) {
    coverUploadError.value = '请先登录后再上传封面'
    return
  }

  const allowedTypes = new Set(['image/jpeg', 'image/png', 'image/gif', 'image/webp'])
  const maxSize = 5 * 1024 * 1024

  if (!allowedTypes.has(file.type)) {
    coverUploadError.value = '只支持 JPEG、PNG、GIF、WebP 格式的图片'
    return
  }
  if (file.size > maxSize) {
    coverUploadError.value = '图片不能超过 5MB'
    return
  }

  coverUploading.value = true
  coverUploadError.value = ''

  try {
    const formData = new FormData()
    formData.append('image', file)

    const res = await fetch(api.blog.uploadImage, {
      method: 'POST',
      headers: authHeaders.value,
      body: formData,
    })

    const data = await res.json().catch(() => null)
    if (!res.ok) {
      throw new Error(data?.error || '封面上传失败')
    }
    if (!data?.url) {
      throw new Error('服务器没有返回封面地址')
    }

    form.value.cover_url = data.url
    if (contentSource.value === 'empty') {
      contentSource.value = 'manual'
    }
  } catch (e) {
    coverUploadError.value = e instanceof Error ? e.message : '封面上传失败'
  } finally {
    coverUploading.value = false
  }
}

const removeCover = () => {
  form.value.cover_url = ''
  coverUploadError.value = ''
}

const clearServerSyncTimer = () => {
  if (serverSyncTimer) {
    clearTimeout(serverSyncTimer)
    serverSyncTimer = null
  }
}

const blogDraftToPayload = (draft: BlogDraft): EditorDraftPayload => ({
  context_key: draft.context_key,
  source_post_id: draft.source_post_id,
  title: draft.title || '',
  content: draft.content || '',
  summary: draft.summary || '',
  cover_url: draft.cover_url || '',
  allow_comments: draft.allow_comments,
  channel_id: draft.channel_id,
  collection_ids: draft.collection_ids || [],
})

const updateEditorChannel = async (channelId?: string) => {
  if (channelId) {
    await router.replace({ path: editorRoutePath.value, query: { channel: channelId } })
    currentChannelId.value = channelId
    return
  }

  if (!isEdit.value && selectedChannelId.value) {
    await router.replace({ path: editorRoutePath.value })
  }
  currentChannelId.value = ''
}

const fetchServerDraft = async () => {
  if (!authStore.token) return null

  try {
    const res = await fetch(`${api.blog.draft}?context_key=${encodeURIComponent(draftContextKey.value)}`, {
      headers: authHeaders.value,
    })
    if (!res.ok) return null

    const data = await res.json()
    return (data.data || null) as BlogDraft | null
  } catch (e) {
    console.error('Failed to fetch blog draft:', e)
    return null
  }
}

const deleteServerDraft = async () => {
  clearServerSyncTimer()
  serverDraftState.value = 'idle'
  serverDraftSavedAt.value = null

  if (!authStore.token) return

  try {
    await fetch(`${api.blog.draft}?context_key=${encodeURIComponent(draftContextKey.value)}`, {
      method: 'DELETE',
      headers: authHeaders.value,
    })
  } catch (e) {
    console.error('Failed to delete blog draft:', e)
  }
}

const syncServerDraft = async () => {
  if (!authStore.token || !draftWatchEnabled.value) return

  const payload = draftPayload.value
  if (!hasMeaningfulDraft(payload)) {
    await deleteServerDraft()
    return
  }

  serverDraftState.value = 'syncing'
  try {
    const res = await fetch(api.blog.draft, {
      method: 'PUT',
      headers: { 'Content-Type': 'application/json', ...authHeaders.value },
      body: JSON.stringify(payload),
    })
    if (!res.ok) throw new Error('Failed to sync draft')

    const data = await res.json()
    const draft = (data.data || null) as BlogDraft | null
    serverDraftSavedAt.value = draft ? parseTimestamp(draft.updated_at) : Date.now()
    serverDraftState.value = 'synced'
  } catch (e) {
    console.error('Failed to sync blog draft:', e)
    serverDraftState.value = 'error'
  }
}

const scheduleServerDraftSync = () => {
  clearServerSyncTimer()
  if (!authStore.token || !draftWatchEnabled.value || isApplyingDraft.value) return
  serverSyncTimer = setTimeout(() => {
    void syncServerDraft()
  }, 1800)
}

const clearAllDrafts = async () => {
  clearLocalDraft()
  await deleteServerDraft()
  pendingDraftCandidate.value = null
  deferredDraftCandidate.value = null
  recoveryModalVisible.value = false
}

const applyDraftPayload = async (payload: EditorDraftPayload) => {
  isApplyingDraft.value = true
  try {
    form.value = {
      title: payload.title,
      content: payload.content,
      summary: payload.summary,
      cover_url: payload.cover_url,
      allow_comments: payload.allow_comments,
    }

    contentSource.value = hasMeaningfulDraft(payload) ? 'manual' : 'empty'

    const draftChannelId = payload.channel_id?.trim() || ''
    if (!isEdit.value || (draftChannelId && draftChannelId !== selectedChannelId.value)) {
      await updateEditorChannel(draftChannelId || undefined)
    }

    if (draftChannelId || selectedChannelId.value) {
      await loadChannelCollections()
      const allowed = new Set(channelCollections.value.map(collection => collection.id))
      selectedCollectionIds.value = payload.collection_ids.filter(collectionId => allowed.has(collectionId))
      if (draftChannelId && selectedCollectionIds.value.length === 0) {
        ensureDefaultSelection()
      }
    } else {
      selectedCollectionIds.value = [...payload.collection_ids]
    }

    await nextTick()
    autoResizeTitle()
  } finally {
    isApplyingDraft.value = false
  }
}

const evaluateDraftRecovery = async () => {
  const localDraft = loadDraft()
  const serverDraft = await fetchServerDraft()

  const candidates: DraftCandidate[] = []
  if (localDraft && hasMeaningfulDraft(localDraft.payload)) {
    lastSavedAt.value = localDraft.saved_at
    candidates.push({
      source: 'local',
      payload: localDraft.payload,
      savedAt: localDraft.saved_at,
    })
  }
  if (serverDraft) {
    const payload = blogDraftToPayload(serverDraft)
    if (hasMeaningfulDraft(payload)) {
      const savedAt = parseTimestamp(serverDraft.updated_at)
      serverDraftSavedAt.value = savedAt || serverDraftSavedAt.value
      candidates.push({
        source: 'server',
        payload,
        savedAt,
      })
    }
  }

  if (!candidates.length) return

  candidates.sort((left, right) => right.savedAt - left.savedAt)
  const latestCandidate = candidates[0]
  if (isEdit.value && latestCandidate.savedAt <= loadedPostUpdatedAt.value) {
    return
  }

  pendingDraftCandidate.value = latestCandidate
  deferredDraftCandidate.value = latestCandidate
  recoveryModalVisible.value = true
}

const keepCurrentContent = () => {
  recoveryModalVisible.value = false
}

const openDraftManager = () => {
  draftManagerVisible.value = true
}

const closeDraftManager = () => {
  draftManagerVisible.value = false
}

const reopenDraftRecovery = () => {
  if (!deferredDraftCandidate.value) return
  pendingDraftCandidate.value = deferredDraftCandidate.value
  recoveryModalVisible.value = true
}

const restorePendingDraft = async () => {
  const candidate = pendingDraftCandidate.value
  if (!candidate) return

  recoveryModalVisible.value = false
  deferredDraftCandidate.value = null
  pendingDraftCandidate.value = null
  await applyDraftPayload(candidate.payload)
}

const discardPendingDraft = async () => {
  await clearAllDrafts()
}

const restoreDeferredFromManager = async () => {
  if (!deferredDraftCandidate.value) return
  pendingDraftCandidate.value = deferredDraftCandidate.value
  draftManagerVisible.value = false
  await restorePendingDraft()
}

const syncDraftNow = async () => {
  clearServerSyncTimer()
  await syncServerDraft()
}

const clearSavedDrafts = async () => {
  await clearAllDrafts()
  draftManagerVisible.value = false
}

const cancelLeave = () => {
  leaveConfirmVisible.value = false
  pendingLeavePath.value = null
}

const confirmLeave = async () => {
  const targetPath = pendingLeavePath.value
  leaveConfirmVisible.value = false
  pendingLeavePath.value = null
  if (!targetPath) return

  allowRouteLeaveOnce.value = true
  await router.push(targetPath)
}

const handleBeforeUnload = (event: BeforeUnloadEvent) => {
  if (!hasPendingPersistence.value) return
  event.preventDefault()
  event.returnValue = ''
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
      if (selectedChannelId.value) currentChannelId.value = selectedChannelId.value
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
    if (!res.ok) { error.value = '加载合集失败'; return }
    const data = await res.json()
    channelCollections.value = data.data || []
    if (!isEdit.value) {
      const def = channelCollections.value.find(c => c.is_default) || channelCollections.value[0]
      selectedCollectionIds.value = def ? [def.id] : []
    }
    if (isEdit.value) {
      const allowed = channelCollections.value.map(c => c.id)
      selectedCollectionIds.value = existingCollectionIds.value.filter(id => allowed.includes(id))
      ensureDefaultSelection()
    }
  } catch (e) {
    console.error(e)
    error.value = '加载合集失败'
  }
}

const onCollectionToggle = (id: string, event: Event) => {
  const checked = !!(event.target as HTMLInputElement)?.checked
  if (checked) {
    if (!selectedCollectionIds.value.includes(id))
      selectedCollectionIds.value = [...selectedCollectionIds.value, id]
  } else {
    selectedCollectionIds.value = selectedCollectionIds.value.filter(x => x !== id)
    ensureDefaultSelection()
  }
}

// ── 同步合集 ─────────────────────────────────────────────
const syncPostCollections = async (postId: string) => {
  if (!selectedChannelId.value) return
  const target = Array.from(new Set(selectedCollectionIds.value))
  const existing = Array.from(new Set(existingCollectionIds.value))
  const toAdd = target.filter(id => !existing.includes(id))
  const toRemove = existing.filter(id => !target.includes(id))
  for (const id of toAdd) {
    const res = await fetch(api.blog.postCollections(postId), {
      method: 'POST',
      headers: { 'Content-Type': 'application/json', ...authHeaders.value },
      body: JSON.stringify({ collection_id: id }),
    })
    if (!res.ok) throw new Error('添加文章合集失败')
  }
  for (const id of toRemove) {
    const res = await fetch(api.blog.postCollection(postId, id), {
      method: 'DELETE', headers: authHeaders.value,
    })
    if (!res.ok) throw new Error('移除文章合集失败')
  }
  existingCollectionIds.value = [...target]
}

// ── 加载文章 ─────────────────────────────────────────────
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
      loadedPostUpdatedAt.value = parseTimestamp(p.updated_at)
      contentSource.value = 'manual'
      if (!selectedChannelId.value) {
        const fallback = p.collections?.[0]?.channel_id
        if (fallback) {
          await router.replace({ path: `/post/${postId}/edit`, query: { channel: fallback } })
        }
      }
      existingCollectionIds.value = (p.collections || [])
        .filter((c: Collection) => c.channel_id === selectedChannelId.value)
        .map((c: Collection) => c.id)
      selectedCollectionIds.value = [...existingCollectionIds.value]
    }
  } catch (e) {
    console.error(e)
  } finally {
    contentReady.value = true
    await nextTick()
    autoResizeTitle()
  }
}

// ── 保存 ─────────────────────────────────────────────────
const save = async (status: SaveTarget) => {
  if (!isEdit.value && !selectedChannelId.value) {
    error.value = '请先选择合集再开始写作'; return
  }
  if (!form.value.title.trim()) { error.value = '请输入文章标题'; return }
  if (!form.value.content.trim()) { error.value = '请输入文章内容'; return }
  error.value = ''
  saving.value = status
  const payload = { ...form.value, status }
  try {
    let res: Response
    if (isEdit.value) {
      const postId = String(route.params.id || '')
      res = await fetch(api.blog.post(postId), {
        method: 'PUT',
        headers: { 'Content-Type': 'application/json', Authorization: `Bearer ${authStore.token}` },
        body: JSON.stringify(payload),
      })
    } else {
      res = await fetch(api.blog.posts, {
        method: 'POST',
        headers: { 'Content-Type': 'application/json', Authorization: `Bearer ${authStore.token}` },
        body: JSON.stringify({
          ...payload,
          channel_id: selectedChannelId.value,
          collection_ids: Array.from(new Set(selectedCollectionIds.value)),
        }),
      })
    }
    if (res.ok) {
      const d = await res.json()
      const savedPost = d.data || d
      await clearAllDrafts()
      if (isEdit.value && selectedChannelId.value) await syncPostCollections(String(savedPost.id))
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

// ── 导入 Markdown ─────────────────────────────────────────
const handleFileUpload = async (event: Event) => {
  const target = event.target as HTMLInputElement
  const file = target.files?.[0]
  if (!file) return
  uploading.value = true
  try {
    const text = await file.text()
    const lines = text.split('\n')
    let title = ''
    let content = text
    for (const line of lines) {
      const trimmed = line.trim()
      if (trimmed.startsWith('# ')) { title = trimmed.slice(2).trim(); break }
    }
    if (title) {
      const idx = text.split('\n').findIndex(l => l.trim().startsWith('# '))
      if (idx !== -1) content = text.split('\n').slice(idx + 1).join('\n').trim()
    }
    form.value.title = title || file.name.replace(/\.(md|markdown|txt)$/i, '')
    form.value.content = content
    contentSource.value = 'imported'
  } catch (e) {
    error.value = '读取文件失败'
  } finally {
    uploading.value = false
    target.value = ''
  }
}

const clearContent = () => {
  form.value.content = ''
  form.value.title = ''
  contentSource.value = 'empty'
}

const triggerReimport = () => { fileInput.value?.click() }

// ── 标题自动扩展高度 ─────────────────────────────────────
const autoResizeTitle = () => {
  const el = titleRef.value
  if (!el) return
  el.style.height = 'auto'
  el.style.height = el.scrollHeight + 'px'
}

// ── 内容变化检测 ─────────────────────────────────────────
watch(() => form.value.title, (nv, ov) => {
  if (!ov && nv && contentSource.value === 'empty') contentSource.value = 'manual'
  nextTick(autoResizeTitle)
})

watch(currentDraftSignature, () => {
  if (!draftWatchEnabled.value || isApplyingDraft.value || !contentReady.value) return
  if (contentSource.value === 'empty' && hasMeaningfulDraft(draftPayload.value)) {
    contentSource.value = 'manual'
  }
  triggerAutoSave()
  scheduleServerDraftSync()
})

watch(() => selectedChannelId.value, loadChannelCollections)

onBeforeRouteLeave((to) => {
  if (allowRouteLeaveOnce.value) {
    allowRouteLeaveOnce.value = false
    return true
  }
  if (!hasPendingPersistence.value) return true
  pendingLeavePath.value = to.fullPath
  leaveConfirmVisible.value = true
  return false
})

// ── 初始化 ───────────────────────────────────────────────
onMounted(async () => {
  window.addEventListener('beforeunload', handleBeforeUnload)
  await loadChannels()
  await loadPost()
  await loadChannelCollections()
  await evaluateDraftRecovery()
  draftWatchEnabled.value = true
})

onBeforeUnmount(() => {
  window.removeEventListener('beforeunload', handleBeforeUnload)
  clearServerSyncTimer()
})
</script>

<style scoped>
.editor-page {
  height: calc(100vh - 64px);
  background: var(--a-color-surface);
  overflow: hidden;
}

.editor-shell {
  display: flex;
  flex-direction: column;
  gap: 1rem;
  padding: 1rem;
  height: 100%;
  box-sizing: border-box;
}

.import-actions,
.meta-chip-group {
  display: flex;
  align-items: center;
  gap: 0.5rem;
  flex-wrap: wrap;
}

.editor-error {
  padding: 0.9rem 1rem;
}

.editor-layout {
  display: grid;
  grid-template-columns: minmax(0, 1fr) 300px;
  gap: 1rem;
  flex: 1;
  min-height: 0;
}

.col-left,
.col-center,
.col-right {
  background: var(--a-color-bg);
  min-height: 0;
}

.col-left,
.col-right {
  display: flex;
  flex-direction: column;
  gap: 0;
  overflow-y: auto;
}

.col-center {
  display: flex;
  flex-direction: column;
  min-width: 0;
  overflow: hidden;
}

/* 右侧 section 分区 */
.right-section {
  display: flex;
  flex-direction: column;
  gap: 0.75rem;
  padding: 1.25rem;
  border-bottom: var(--a-border);
}

.right-section:last-child {
  border-bottom: none;
}

/* 发布区 */
.publish-section {
  gap: 0.75rem;
}

.publish-actions {
  display: flex;
  flex-direction: column;
  gap: 0.5rem;
}

.toc-list,
.hidden-file-input {
  width: 100%;
}

.collection-item {
  display: grid;
  grid-template-columns: auto minmax(0, 1fr) auto;
  align-items: center;
  gap: 0.75rem;
  padding: 0.9rem 1rem;
  border: var(--a-border);
  background: var(--a-color-bg);
  cursor: pointer;
}

.collection-item.selected {
  background: var(--a-color-fg);
  color: var(--a-color-bg);
}

.collection-item.selected .badge-default {
  border-color: var(--a-color-bg);
  color: var(--a-color-bg);
}

.collection-name {
  min-width: 0;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
  font-size: 0.86rem;
  font-weight: 800;
}

.badge-default {
  padding: 0.2rem 0.4rem;
  border: 1.5px solid var(--a-color-border);
  font-size: 0.65rem;
  font-weight: 900;
  letter-spacing: 0.05em;
  text-transform: uppercase;
}

.col-empty,
.col-loading,
.editor-loading {
  color: var(--a-color-muted);
  font-size: 0.82rem;
  font-weight: 700;
}

.section-heading-row {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 0.75rem;
}

.editor-workspace {
  display: flex;
  flex: 1;
  flex-direction: column;
  gap: 1rem;
  min-height: 0;
  padding: 1.25rem;
}

.editor-meta-row {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 1rem;
  flex-wrap: wrap;
}

.editor-meta-actions {
  display: flex;
  align-items: center;
  gap: 0.75rem;
  flex-wrap: wrap;
}

.draft-status {
  display: inline-flex;
  align-items: center;
  padding: 0.35rem 0.6rem;
  border: var(--a-border);
  font-size: 0.7rem;
  font-weight: 800;
  letter-spacing: 0.04em;
  background: var(--a-color-bg);
}

.draft-status.is-ok {
  border-color: var(--a-color-success);
  color: var(--a-color-success);
  background: color-mix(in srgb, var(--a-color-success) 8%, var(--a-color-bg));
}

.draft-status.is-warn {
  border-color: var(--a-color-danger);
  color: var(--a-color-danger);
  background: color-mix(in srgb, var(--a-color-danger) 8%, var(--a-color-bg));
}

.draft-status.is-muted {
  border-color: var(--a-color-disabled-border);
  color: var(--a-color-muted);
  background: var(--a-color-surface);
}


.meta-chip {
  display: inline-flex;
  align-items: center;
  padding: 0.4rem 0.65rem;
  border: var(--a-border);
  font-size: 0.72rem;
  font-weight: 900;
  letter-spacing: 0.08em;
  text-transform: uppercase;
  background: var(--a-color-bg);
}

.hidden-file-input {
  display: none;
}

.editor-canvas {
  display: flex;
  flex: 1;
  min-height: 0;
  flex-direction: column;
  overflow: auto;
  border: var(--a-border);
  background: var(--a-color-bg);
}

.title-sticky {
  position: sticky;
  top: 0;
  z-index: 8;
  background: var(--a-color-bg);
  flex-shrink: 0;
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

.title-input {
  width: 100%;
  padding: 2rem 2rem 1rem;
  border: none;
  outline: none;
  resize: none;
  overflow: hidden;
  box-sizing: border-box;
  font-family: inherit;
  font-size: clamp(2rem, 3vw, 2.75rem);
  font-weight: 900;
  letter-spacing: -0.04em;
  line-height: 1.12;
  color: var(--a-color-fg);
  background: var(--a-color-bg);
}

.title-input::placeholder {
  color: var(--a-color-muted-soft);
}

.title-divider {
  height: 2px;
  margin: 0 2rem;
  background: var(--a-color-disabled-border);
  flex-shrink: 0;
}

.draft-recovery-body {
  display: flex;
  flex-direction: column;
  gap: 0.9rem;
  padding: 1.25rem;
}

.draft-recovery-text {
  margin: 0;
  font-size: 0.95rem;
  line-height: 1.7;
  color: var(--a-color-fg);
}

.draft-recovery-preview {
  display: flex;
  flex-direction: column;
  gap: 0.55rem;
  padding: 1rem;
  background: var(--a-color-surface);
}

.draft-recovery-preview strong {
  font-size: 1rem;
  font-weight: 900;
  letter-spacing: -0.02em;
}

.draft-recovery-preview .a-muted {
  margin: 0;
}

.draft-recovery-actions {
  display: flex;
  justify-content: flex-end;
  gap: 0.75rem;
  padding: 0 1.25rem 1.25rem;
  flex-wrap: wrap;
}

.draft-manager-body,
.leave-confirm-body {
  display: flex;
  flex-direction: column;
  gap: 1rem;
  padding: 1.25rem;
}

.draft-manager-grid {
  display: grid;
  grid-template-columns: repeat(2, minmax(0, 1fr));
  gap: 0.75rem;
}

.draft-manager-card {
  display: flex;
  flex-direction: column;
  gap: 0.55rem;
  padding: 1rem;
}

.draft-manager-card strong {
  font-size: 1rem;
  font-weight: 900;
  letter-spacing: -0.02em;
}

.draft-manager-card .a-muted {
  margin: 0;
}

.draft-manager-card-accent {
  background: var(--a-color-surface);
}

.draft-manager-preview,
.leave-confirm-text {
  margin: 0;
  font-size: 0.94rem;
  line-height: 1.7;
  color: var(--a-color-fg);
}

.draft-manager-warning {
  padding: 0.95rem 1rem;
  border: 2px solid var(--a-color-danger);
  background: color-mix(in srgb, var(--a-color-danger) 8%, var(--a-color-bg));
  color: var(--a-color-danger);
  font-size: 0.85rem;
  font-weight: 700;
  line-height: 1.6;
}

.leave-confirm-body .a-muted {
  margin: 0;
}

.stat-grid {
  display: grid;
  grid-template-columns: repeat(2, minmax(0, 1fr));
  gap: 0.75rem;
}

.stat-card {
  display: flex;
  flex-direction: column;
  gap: 0.35rem;
  padding: 1rem;
  border: var(--a-border);
  background: var(--a-color-surface);
}

.stat-num {
  font-size: 1.75rem;
  font-weight: 900;
  letter-spacing: -0.04em;
  line-height: 1;
}

.stat-unit {
  font-size: 0.72rem;
  font-weight: 800;
  letter-spacing: 0.08em;
  text-transform: uppercase;
  color: var(--a-color-muted);
}

.toc-section {
  flex: 1;
}

.toc-list {
  overflow-y: auto;
}

.toc-item {
  --toc-depth: 0;
  display: grid;
  grid-template-columns: auto minmax(0, 1fr);
  align-items: start;
  gap: 0.65rem;
  padding: 0.55rem 0.7rem;
  color: var(--a-color-muted);
  text-decoration: none;
  font-size: 0.8rem;
  font-weight: 700;
  line-height: 1.4;
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
  border-left: 2px solid transparent;
}

.toc-item:hover {
  border-left-color: var(--a-color-border);
  background: var(--a-color-surface);
  color: var(--a-color-fg);
}

.toc-rail {
  width: calc(var(--toc-depth) * 0.8rem + 1px);
  min-height: 1.2rem;
  background-image: repeating-linear-gradient(
    to right,
    color-mix(in srgb, var(--a-color-border) 52%, transparent) 0 1px,
    transparent 1px 0.8rem
  );
  background-repeat: no-repeat;
  opacity: 0.9;
}

.toc-text {
  min-width: 0;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.toc-h1 { --toc-depth: 0; font-weight: 900; color: var(--a-color-fg); }
.toc-h2 { --toc-depth: 1; color: var(--a-color-muted); }
.toc-h3 { --toc-depth: 2; color: var(--a-color-muted); font-weight: 700; }
.toc-h4 { --toc-depth: 3; color: var(--a-color-muted-soft); font-weight: 600; }

.cover-upload-card {
  display: flex;
  flex-direction: column;
  gap: 0.85rem;
  padding: 0.9rem;
}

.cover-preview-wrap {
  overflow: hidden;
  border: var(--a-border);
  background: var(--a-color-surface);
}

.cover-preview-image {
  display: block;
  width: 100%;
  aspect-ratio: 16 / 9;
  object-fit: cover;
}

.cover-empty-state {
  display: flex;
  flex-direction: column;
  gap: 0.35rem;
  padding: 1rem;
  border: var(--a-border);
  background: var(--a-color-surface);
}

.cover-empty-state strong {
  font-size: 0.92rem;
  font-weight: 900;
}

.cover-empty-state .a-muted,
.cover-upload-hint {
  margin: 0;
}

.cover-upload-actions {
  display: flex;
  gap: 0.6rem;
  flex-wrap: wrap;
}

.cover-upload-error {
  margin: 0;
  color: var(--a-color-danger);
  font-size: 0.8rem;
  font-weight: 700;
}

.options-list {
  display: flex;
  flex-direction: column;
  gap: 1rem;
}

.option-check {
  display: flex;
  align-items: center;
  gap: 0.6rem;
  font-size: 0.82rem;
  font-weight: 700;
  color: var(--a-color-fg);
}

.option-check {
  justify-content: space-between;
  padding: 0.875rem 1rem;
  background: var(--a-color-bg);
  cursor: pointer;
}

.option-check input {
  width: 1rem;
  height: 1rem;
  margin: 0;
  accent-color: var(--a-color-fg);
}

@media (max-width: 1100px) {
  .editor-layout {
    grid-template-columns: minmax(0, 1fr);
  }

  .col-right {
    display: none;
  }
}

@media (max-width: 800px) {
  .editor-shell {
    padding: 0.75rem;
  }
}

@media (max-width: 640px) {
  .editor-workspace,
  .col-right {
    padding: 1rem;
  }

  .editor-meta-actions,
  .draft-recovery-actions {
    width: 100%;
  }

  .draft-status,
  .publish-actions :deep(.a-btn),
  .editor-meta-actions :deep(.a-btn) {
    width: 100%;
    justify-content: center;
  }

  .title-input {
    padding: 1.25rem 1.25rem 0.85rem;
    font-size: 1.8rem;
  }

  .title-divider {
    margin: 0 1.25rem;
  }

  .stat-grid {
    grid-template-columns: minmax(0, 1fr);
  }

  .draft-manager-grid {
    grid-template-columns: minmax(0, 1fr);
  }
}
</style>
