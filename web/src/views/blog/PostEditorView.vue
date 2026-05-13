<template>
  <div class="editor-page">
    <div class="editor-shell">
      <div v-if="error" class="editor-error">{{ error }}</div>

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
                  <label v-if="contentSource === 'empty'" class="import-btn">
                    <input type="file" accept=".md,.markdown,.txt" @change="handleFileUpload" class="hidden-file-input" />
                    导入 Markdown
                  </label>
                  <template v-else-if="contentSource === 'imported'">
                    <button type="button" class="import-btn" @click="triggerReimport">
                      <input ref="fileInput" type="file" accept=".md,.markdown,.txt" @change="handleFileUpload" class="hidden-file-input" />
                      重新导入
                    </button>
                    <button type="button" class="import-btn-clear" @click="clearContent">清空内容</button>
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
              <button class="btn-save" @click="save('draft')" :disabled="!!saving">
                {{ saving === 'draft' ? '保存中…' : '存草稿' }}
              </button>
              <button class="btn-publish" @click="save('published')" :disabled="!!saving">
                {{ saving === 'published' ? '发布中…' : '发布文章' }}
              </button>
            </div>
            <button v-if="hasDraftManagerAccess" type="button" class="draft-status-btn" @click="openDraftManager">
              草稿管理
            </button>
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
                <label class="a-field-label">封面图 URL</label>
                <input v-model="form.cover_url" placeholder="https://…" class="a-input" />
              </div>
              <label class="option-check">
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
              >{{ item.text }}</a>
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

        <div class="draft-recovery-preview">
          <strong>{{ pendingDraftCandidate.payload.title || '未命名草稿' }}</strong>
          <p class="a-muted">{{ draftRecoveryPreview }}</p>
        </div>
      </div>

      <template #footer>
        <div class="draft-recovery-actions">
          <button type="button" class="btn-save" @click="keepCurrentContent">稍后处理</button>
          <button type="button" class="import-btn-clear" @click="discardPendingDraft">丢弃草稿</button>
          <button type="button" class="btn-publish" @click="restorePendingDraft">恢复草稿</button>
        </div>
      </template>
    </AModal>

    <AModal v-if="draftManagerVisible" title="草稿管理" size="md" @close="closeDraftManager">
      <div class="draft-manager-body">
        <div class="draft-manager-grid">
          <div class="draft-manager-card">
            <span class="a-label">本地草稿</span>
            <strong>{{ localDraftStatusText }}</strong>
            <p class="a-muted">保存在当前浏览器中，刷新页面后仍可恢复。</p>
          </div>

          <div class="draft-manager-card">
            <span class="a-label">云端草稿</span>
            <strong>{{ cloudDraftStatusText }}</strong>
            <p class="a-muted">登录状态下自动同步，可在其他会话中继续写作。</p>
          </div>
        </div>

        <div v-if="deferredDraftCandidate" class="draft-manager-card draft-manager-card-accent">
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
          <button type="button" class="btn-save" @click="closeDraftManager">关闭</button>
          <button
            v-if="authStore.token && hasMeaningfulDraft(draftPayload)"
            type="button"
            class="import-btn"
            @click="syncDraftNow"
          >
            立即同步
          </button>
          <button
            v-if="hasDraftManagerAccess"
            type="button"
            class="import-btn-clear"
            @click="clearSavedDrafts"
          >
            清除已保存草稿
          </button>
          <button
            v-if="deferredDraftCandidate"
            type="button"
            class="btn-publish"
            @click="restoreDeferredFromManager"
          >
            恢复最新草稿
          </button>
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
          <button type="button" class="btn-save" @click="cancelLeave">留在此页</button>
          <button type="button" class="btn-publish" @click="confirmLeave">继续离开</button>
        </div>
      </template>
    </AModal>
  </div>
</template>

<script setup lang="ts">
import { computed, onBeforeUnmount, onMounted, ref, watch, nextTick } from 'vue'
import { RouterLink, onBeforeRouteLeave, useRoute, useRouter } from 'vue-router'

import AEditor from '@/components/shared/AEditor.vue'
import { useAutoSave } from '@/components/blog/composables/useAutoSave'
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
const error = ref('')
const titleRef = ref<HTMLTextAreaElement | null>(null)
const fileInput = ref<HTMLInputElement | null>(null)
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
  background: #f6f6f3;
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

.layout-btn,
.right-tab-btn,
.btn-save,
.btn-publish,
.import-btn,
.import-btn-clear {
  border: 2px solid #000;
  background: #fff;
  color: #000;
  font-family: inherit;
  font-size: 0.75rem;
  font-weight: 900;
  letter-spacing: 0.08em;
  text-transform: uppercase;
  cursor: pointer;
  transition: background 0.15s ease, color 0.15s ease, transform 0.15s ease;
}

.layout-btn,
.right-tab-btn {
  border: none;
  padding: 0.5rem 0.85rem;
}

.layout-btn.active,
.right-tab-btn.active,
.btn-publish {
  background: #000;
  color: #fff;
}

.right-panel-tabs,
.import-actions,
.meta-chip-group {
  display: flex;
  align-items: center;
  gap: 0.5rem;
  flex-wrap: wrap;
}

.btn-save,
.btn-publish,
.import-btn,
.import-btn-clear {
  padding: 0.65rem 1rem;
}

.btn-save:hover,
.import-btn:hover,
.import-btn-clear:hover,
.layout-btn:hover,
.right-tab-btn:hover {
  background: #000;
  color: #fff;
}

.btn-publish:hover {
  background: #1f2937;
}

.btn-save:disabled,
.btn-publish:disabled {
  opacity: 0.5;
  cursor: not-allowed;
}

.editor-error {
  padding: 0.9rem 1rem;
  border: 2px solid #fca5a5;
  background: #fff2f2;
  font-size: 0.85rem;
  font-weight: 700;
  color: #b91c1c;
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
  background: #fff;
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

.panel-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 1rem;
}

.panel-title {
  margin: 0.2rem 0 0;
  font-size: 1rem;
  font-weight: 900;
  letter-spacing: -0.02em;
}

/* 右侧 section 分区 */
.right-section {
  display: flex;
  flex-direction: column;
  gap: 0.75rem;
  padding: 1.25rem;
  border-bottom: 2px solid #000;
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

.publish-actions .btn-save,
.publish-actions .btn-publish {
  width: 100%;
  justify-content: center;
}

.left-panel-section,
.info-section {
  display: flex;
  flex-direction: column;
  gap: 0.75rem;
}

.collections-section {
  flex: 1;
  min-height: 0;
}

.collection-list,
.toc-list,
.status-list {
  display: flex;
  flex-direction: column;
  gap: 0.6rem;
}

.channel-select,
.hidden-file-input {
  width: 100%;
}

.channel-select {
  padding: 0.8rem 0.9rem;
  border: 2px solid #000;
  font-family: inherit;
  font-size: 0.9rem;
  font-weight: 700;
  background: #fff;
  appearance: none;
  cursor: pointer;
  outline: none;
  background-image: url("data:image/svg+xml,%3Csvg xmlns='http://www.w3.org/2000/svg' width='10' height='10' viewBox='0 0 12 12'%3E%3Cpath fill='%23000' d='M6 8L1 3h10z'/%3E%3C/svg%3E");
  background-repeat: no-repeat;
  background-position: right 0.8rem center;
}

.collection-item {
  display: grid;
  grid-template-columns: auto minmax(0, 1fr) auto;
  align-items: center;
  gap: 0.75rem;
  padding: 0.9rem 1rem;
  border: 2px solid #000;
  background: #fff;
  cursor: pointer;
}

.collection-item.selected {
  background: #000;
  color: #fff;
}

.collection-item.selected .badge-default {
  border-color: #fff;
  color: #fff;
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
  border: 1.5px solid #000;
  font-size: 0.65rem;
  font-weight: 900;
  letter-spacing: 0.05em;
  text-transform: uppercase;
}

.col-empty,
.col-loading,
.editor-loading {
  color: #6b7280;
  font-size: 0.82rem;
  font-weight: 700;
}

.section-heading-row {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 0.75rem;
}

.channel-picker-state {
  flex: 1;
  display: flex;
  align-items: center;
  justify-content: center;
  padding: 2rem;
}

.channel-picker-card {
  width: min(100%, 60rem);
  padding: 2rem;
}

.channel-picker-card .a-subtitle,
.channel-picker-card .a-muted {
  margin: 0 0 1rem;
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
  padding: 1.25rem;
  color: #000;
  text-align: left;
}

.channel-picker-name {
  font-size: 1rem;
  font-weight: 900;
  letter-spacing: -0.02em;
}

.channel-picker-actions {
  display: flex;
  gap: 0.75rem;
  flex-wrap: wrap;
  margin-top: 1.5rem;
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

.editor-meta-actions,
.draft-status-group {
  display: flex;
  align-items: center;
  gap: 0.75rem;
  flex-wrap: wrap;
}

.draft-status {
  display: inline-flex;
  align-items: center;
  padding: 0.35rem 0.6rem;
  border: 2px solid #000;
  font-size: 0.7rem;
  font-weight: 800;
  letter-spacing: 0.04em;
  background: #fff;
}

.draft-status.is-ok {
  border-color: #166534;
  color: #166534;
  background: #f0fdf4;
}

.draft-status.is-warn {
  border-color: #92400e;
  color: #92400e;
  background: #fffbeb;
}

.draft-status.is-muted {
  border-color: #d1d5db;
  color: #6b7280;
  background: #fafaf9;
}

.draft-status-btn {
  border: 2px solid #000;
  background: #fff;
  color: #000;
  font-family: inherit;
  font-size: 0.72rem;
  font-weight: 900;
  letter-spacing: 0.08em;
  text-transform: uppercase;
  padding: 0.55rem 0.85rem;
  cursor: pointer;
  width: 100%;
  text-align: center;
}

.draft-status-btn:hover {
  background: #000;
  color: #fff;
}

.editor-mode-group {
  display: inline-flex;
  gap: 0.5rem;
  align-items: center;
}

.draft-status-btn.is-active {
  background: #000;
  color: #fff;
}


.meta-chip {
  display: inline-flex;
  align-items: center;
  padding: 0.4rem 0.65rem;
  border: 2px solid #000;
  font-size: 0.72rem;
  font-weight: 900;
  letter-spacing: 0.08em;
  text-transform: uppercase;
}

.meta-chip-active {
  background: #000;
  color: #fff;
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
  border: 2px solid #000;
  background: #fff;
}

.title-sticky {
  position: sticky;
  top: 0;
  z-index: 8;
  background: #fff;
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
  color: #000;
  background: #fff;
}

.title-input::placeholder {
  color: #d1d5db;
}

.title-divider {
  height: 2px;
  margin: 0 2rem;
  background: #e5e7eb;
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
  color: #111827;
}

.draft-recovery-preview {
  display: flex;
  flex-direction: column;
  gap: 0.55rem;
  padding: 1rem;
  border: 2px solid #000;
  background: #fafaf9;
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
  border: 2px solid #000;
  background: #fff;
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
  background: #fafaf9;
}

.draft-manager-preview,
.leave-confirm-text {
  margin: 0;
  font-size: 0.94rem;
  line-height: 1.7;
  color: #111827;
}

.draft-manager-warning {
  padding: 0.95rem 1rem;
  border: 2px solid #92400e;
  background: #fffbeb;
  color: #92400e;
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
  border: 2px solid #000;
  background: #fafafa;
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
  color: #6b7280;
}

.toc-section {
  flex: 1;
}

.toc-list {
  overflow-y: auto;
}

.toc-item {
  display: flex;
  align-items: baseline;
  gap: 0.4rem;
  padding: 0.55rem 0.7rem;
  color: #374151;
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
  border-left-color: #000;
  background: #f3f4f6;
  color: #000;
}

/* Vertical-bar depth indicator */
.toc-item::before {
  content: '';
  display: inline-block;
  flex-shrink: 0;
  width: 0;
  height: 0.9em;
  border-left: none;
}

.toc-h1::before { width: 0; }
.toc-h2::before { width: 0.7rem; border-left: 2px solid #d1d5db; }
.toc-h3::before { width: 1.4rem; border-left: 2px solid #d1d5db; box-shadow: -6px 0 0 0 #d1d5db; }
.toc-h4::before { width: 2.1rem; border-left: 2px solid #d1d5db; box-shadow: -6px 0 0 0 #d1d5db, -12px 0 0 0 #d1d5db; }

.toc-h1 { font-weight: 900; color: #111; }
.toc-h2 { color: #374151; }
.toc-h3 { color: #6b7280; font-weight: 700; }
.toc-h4 { color: #9ca3af; font-weight: 600; }

.options-list {
  display: flex;
  flex-direction: column;
  gap: 1rem;
}

.option-check,
.status-row {
  display: flex;
  align-items: center;
  gap: 0.6rem;
  font-size: 0.82rem;
  font-weight: 700;
  color: #374151;
}

.status-dot {
  width: 0.55rem;
  height: 0.55rem;
  border-radius: 9999px;
  flex-shrink: 0;
}

.status-dot.ok { background: #16a34a; }
.status-dot.warn { background: #f59e0b; }

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
  .draft-status-group,
  .draft-recovery-actions {
    width: 100%;
  }

  .draft-status,
  .draft-status-btn {
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
