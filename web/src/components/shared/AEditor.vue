<template>
  <div class="a-editor" :class="{ 'no-border': noBorder, [`mode-${mode}`]: true }" @keydown="onContainerKeydown">

    <!-- ── SV 模式 ──────────────────────────────────────── -->
    <template v-if="mode === 'sv'">

      <!-- 协作 presence bar -->
      <div v-if="collabPeers.length > 0" class="a-editor-presence">
        <span class="a-label">协作中</span>
        <div class="presence-avatars">
          <div
            v-for="peer in collabPeers"
            :key="peer.clientId"
            class="presence-dot"
            :style="{ background: peer.color }"
            :title="peer.name"
          >{{ peer.name.charAt(0).toUpperCase() }}</div>
        </div>
      </div>

      <!-- 工具栏 sticky -->
      <div class="a-editor-toolbar">
        <button type="button" class="tb-btn" title="撤销" @click="sv_undo">↶</button>
        <button type="button" class="tb-btn" title="重做" @click="sv_redo">↷</button>
        <button type="button" class="tb-btn" title="插入代码块" @click="sv_insertCodeBlock">Code</button>
        <button type="button" class="tb-btn" title="清空内容" @click="sv_clearContent">Clear</button>
        <button
          type="button"
          class="tb-btn"
          :class="{ active: showLineNumbers }"
          title="显示或隐藏行号"
          @click="toggleLineNumbers"
        >
          LN
        </button>
        <span class="tb-sep" />
        <button type="button" class="tb-btn" title="二级标题" @click="sv_wrapLinePrefix('## ', '标题')">H2</button>
        <button type="button" class="tb-btn" title="三级标题" @click="sv_wrapLinePrefix('### ', '标题')">H3</button>
        <button type="button" class="tb-btn" title="粗体" @click="sv_wrap('**', '**', '粗体文字')">B</button>
        <button type="button" class="tb-btn" title="斜体" @click="sv_wrap('*', '*', '斜体文字')">I</button>
        <button type="button" class="tb-btn" title="删除线" @click="sv_wrap('~~', '~~', '删除线')">S</button>
        <button type="button" class="tb-btn" title="行内代码" @click="sv_wrap('`', '`', '代码')">code</button>
        <button type="button" class="tb-btn" title="插入链接" @click="sv_insertLink">Link</button>
        <button type="button" class="tb-btn" :class="{ uploading: uploadingImage }" title="上传图片" @click="triggerImageUpload">
          {{ uploadingImage ? '…' : 'Img' }}
        </button>
        <input ref="imageInputRef" type="file" accept="image/*" class="tb-hidden-input" @change="handleImageUploadFile" />
        <span class="tb-sep" />
        <button type="button" class="tb-btn" title="引用" @click="sv_wrapLinePrefix('> ', '引用内容')">Quote</button>
        <button type="button" class="tb-btn" title="无序列表" @click="sv_wrapLinePrefix('- ', '列表项')">·List</button>
        <button type="button" class="tb-btn" title="有序列表" @click="sv_wrapLinePrefix('1. ', '列表项')">1.List</button>
        <button type="button" class="tb-btn" title="插入表格" @click="sv_insertTable">Table</button>
        <button type="button" class="tb-btn" title="插入分割线" @click="sv_insertHr">HR</button>
        <template v-if="enableEmbeds">
          <span class="tb-sep" />
          <button type="button" class="tb-btn" @click="insertEmbed('post')">POST</button>
          <button type="button" class="tb-btn" @click="insertEmbed('music')">MUSIC</button>
          <button type="button" class="tb-btn" @click="insertEmbed('video')">VIDEO</button>
        </template>
      </div>

      <!-- 编辑区 + 预览区 -->
      <div
        class="a-editor-sv-body"
        @dragover.prevent="onDragOver"
        @dragleave="onDragLeave"
        @drop.prevent="onDrop"
        :class="{ dragging: isDragging }"
      >
        <div class="sv-pane sv-source">
          <div ref="cmContainerRef" class="cm-container" />
        </div>
        <div ref="previewPaneRef" class="sv-pane sv-preview prose-blog" v-html="svPreviewHtml" @scroll="onPreviewScroll" />
      </div>

      <!-- @提及下拉 -->
      <div
        v-if="mention.visible && mention.results.length > 0"
        class="a-mention-dropdown"
        :style="{ top: mention.y + 'px', left: mention.x + 'px' }"
      >
        <button
          v-for="(user, i) in mention.results"
          :key="user.username"
          type="button"
          class="mention-item"
          :class="{ 'is-active': i === mention.index }"
          @mousedown.prevent="applyMention(user)"
        >
          <span class="mention-name">{{ user.display_name || user.username }}</span>
          <span class="mention-username">@{{ user.username }}</span>
        </button>
      </div>
    </template>

    <!-- ── Plain 模式 ────────────────────────────────────── -->
    <template v-else>
      <div class="a-editor-plain-toolbar">
        <button type="button" class="tb-btn" :class="{ uploading: uploadingImage }" title="上传图片" @click="triggerPlainImageUpload">
          {{ uploadingImage ? '…' : 'Img' }}
        </button>
        <input ref="plainImageInputRef" type="file" accept="image/*" class="tb-hidden-input" @change="handlePlainImageUploadFile" />
      </div>
      <textarea
        ref="plainTextareaRef"
        v-model="plainValue"
        class="a-editor-plain-textarea"
        :placeholder="placeholder"
        @input="emit('update:modelValue', plainValue)"
      />
    </template>

  </div>
</template>

<script setup lang="ts">
import { computed, nextTick, onBeforeUnmount, onMounted, ref, watch } from 'vue'
import { EditorView, keymap, scrollPastEnd, lineNumbers, highlightActiveLineGutter } from '@codemirror/view'
import { Compartment, EditorState } from '@codemirror/state'
import { defaultKeymap, historyKeymap, history, indentWithTab, undo, redo } from '@codemirror/commands'
import { markdown } from '@codemirror/lang-markdown'
import { languages } from '@codemirror/language-data'
import * as Y from 'yjs'
import { WebsocketProvider } from 'y-websocket'
import { yCollab } from 'y-codemirror.next'

import { useMarkdownRenderer } from '@/composables/useMarkdownRenderer'
import { useAuthStore } from '@/stores/auth'

// ── Types ──────────────────────────────────────────────
interface Peer { clientId: number; name: string; color: string }
interface MentionUser { uuid: string; username: string; display_name: string; avatar_url: string }

interface Props {
  modelValue?: string
  mode: 'sv' | 'plain'
  placeholder?: string
  noBorder?: boolean
  // sv only
  enableImageUpload?: boolean
  enableMentions?: boolean
  enableEmbeds?: boolean
  enableCollab?: boolean
  collabRoomId?: string
}

const props = withDefaults(defineProps<Props>(), {
  modelValue: '',
  placeholder: '开始输入…',
  noBorder: false,
  enableImageUpload: true,
  enableMentions: false,
  enableEmbeds: false,
  enableCollab: false,
  collabRoomId: undefined,
})

const emit = defineEmits<{ 'update:modelValue': [value: string] }>()

const authStore = useAuthStore()
const { renderMarkdown } = useMarkdownRenderer()

// ── Plain mode ──────────────────────────────────────────
const plainValue = ref(props.modelValue)
watch(() => props.modelValue, (val) => {
  if (props.mode === 'plain') plainValue.value = val
})

// ── SV: preview ─────────────────────────────────────────
const svPreviewHtml = computed(() => renderMarkdown(props.modelValue))

// ── SV: CodeMirror ──────────────────────────────────────
const cmContainerRef = ref<HTMLElement | null>(null)
const previewPaneRef = ref<HTMLElement | null>(null)
const showLineNumbers = ref(false)
const lineNumberCompartment = new Compartment()
let cmView: EditorView | null = null

// ── Collab ──────────────────────────────────────────────
const CURSOR_COLORS = ['#e74c3c', '#3498db', '#2ecc71', '#f39c12', '#9b59b6', '#1abc9c']
const myColor = CURSOR_COLORS[Math.floor(Math.random() * CURSOR_COLORS.length)]
const myName = computed(() => authStore.user?.display_name || authStore.user?.username || '匿名')
const collabPeers = ref<Peer[]>([])

let ydoc: Y.Doc | null = null
let provider: WebsocketProvider | null = null
let ytext: Y.Text | null = null

// ── Scroll sync ─────────────────────────────────────────
let syncingScroll = false

function onCmScroll() {
  if (syncingScroll || !cmView || !previewPaneRef.value) return
  syncingScroll = true
  const dom = cmView.scrollDOM
  const ratio = dom.scrollTop / (dom.scrollHeight - dom.clientHeight || 1)
  const preview = previewPaneRef.value
  preview.scrollTop = ratio * (preview.scrollHeight - preview.clientHeight)
  requestAnimationFrame(() => { syncingScroll = false })
}

function onPreviewScroll() {
  if (syncingScroll || !cmView || !previewPaneRef.value) return
  syncingScroll = true
  const preview = previewPaneRef.value
  const ratio = preview.scrollTop / (preview.scrollHeight - preview.clientHeight || 1)
  const dom = cmView.scrollDOM
  dom.scrollTop = ratio * (dom.scrollHeight - dom.clientHeight)
  requestAnimationFrame(() => { syncingScroll = false })
}

// ── CodeMirror init ─────────────────────────────────────
onMounted(() => {
  if (props.mode !== 'sv') return
  initCodeMirror()
})

function initCodeMirror() {
  if (!cmContainerRef.value) return

  const extensions = [
    history(),
    keymap.of([...defaultKeymap, ...historyKeymap, indentWithTab]),
    markdown({ codeLanguages: languages }),
    lineNumberCompartment.of([]),
    EditorView.lineWrapping,
    scrollPastEnd(),
    EditorView.domEventHandlers({
      scroll: onCmScroll,
      paste: onCmPaste,
      drop: (e) => { e.preventDefault(); handleDropFiles(e.dataTransfer?.files) },
    }),
    EditorView.updateListener.of((update) => {
      if (update.docChanged) {
        const val = update.state.doc.toString()
        emit('update:modelValue', val)
      }
      if (props.enableMentions && (update.docChanged || update.selectionSet || update.viewportChanged || update.focusChanged)) {
        detectMentionFromCm(update)
      }
    }),
    EditorView.theme({
      '&': { height: '100%', fontSize: '0.875rem' },
      '.cm-scroller': { fontFamily: "'SFMono-Regular','Consolas','Liberation Mono',monospace", lineHeight: '1.75', padding: '1.5rem 1.5rem 2rem', overflow: 'auto' },
      '.cm-content': { caretColor: '#000' },
      '.cm-cursor': { borderLeftColor: '#000' },
      '.cm-selectionBackground, ::selection': { backgroundColor: '#d4e0ff' },
      '.cm-focused .cm-selectionBackground': { backgroundColor: '#b3ccff' },
      '.cm-line': { padding: '0' },
      '&.cm-focused': { outline: 'none' },
    }),
  ]

  // Collab with Yjs
  if (props.enableCollab && props.collabRoomId) {
    ydoc = new Y.Doc()
    const proto = location.protocol === 'https:' ? 'wss:' : 'ws:'
    provider = new WebsocketProvider(
      `${proto}//${location.host}/api/collab/ws/${props.collabRoomId}`,
      props.collabRoomId,
      ydoc,
      { connect: true },
    )
    ytext = ydoc.getText('codemirror')

    provider.awareness.on('change', () => {
      const list: Peer[] = []
      provider!.awareness.getStates().forEach((state, clientId) => {
        if (clientId === provider!.awareness.clientID) return
        if (state.user) list.push({ clientId, name: state.user.name as string, color: state.user.color as string })
      })
      collabPeers.value = list
    })
    provider.awareness.setLocalStateField('user', { name: myName.value, color: myColor })

    extensions.push(yCollab(ytext, provider.awareness))

    cmView = new EditorView({
      parent: cmContainerRef.value,
      extensions,
    })
  } else {
    cmView = new EditorView({
      state: EditorState.create({
        doc: props.modelValue,
        extensions,
      }),
      parent: cmContainerRef.value,
    })
  }
}

// Watch external modelValue changes (non-collab mode)
watch(() => props.modelValue, (val) => {
  if (props.mode !== 'sv' || !cmView || props.enableCollab) return
  const current = cmView.state.doc.toString()
  if (current !== val) {
    cmView.dispatch({
      changes: { from: 0, to: current.length, insert: val },
    })
  }
})

// ── SV 工具栏操作 ───────────────────────────────────────
function getCmSelection(): { from: number; to: number; selectedText: string } {
  if (!cmView) return { from: 0, to: 0, selectedText: '' }
  const { from, to } = cmView.state.selection.main
  const selectedText = cmView.state.sliceDoc(from, to)
  return { from, to, selectedText }
}

function cmInsert(from: number, to: number, text: string, cursorFrom?: number, cursorTo?: number) {
  if (!cmView) return
  cmView.dispatch({
    changes: { from, to, insert: text },
    selection: cursorFrom !== undefined
      ? { anchor: cursorFrom, head: cursorTo ?? cursorFrom }
      : undefined,
  })
  cmView.focus()
}

function sv_wrap(before: string, after: string, placeholder: string) {
  const { from, to, selectedText } = getCmSelection()
  const inserted = selectedText || placeholder
  const newText = before + inserted + after
  cmInsert(from, to, newText, from + before.length, from + before.length + inserted.length)
}

function sv_wrapLinePrefix(prefix: string, placeholder: string) {
  if (!cmView) return
  const { from, to, selectedText } = getCmSelection()
  const line = cmView.state.doc.lineAt(from)
  const lineStart = line.from
  const selected = selectedText || placeholder
  cmInsert(lineStart, to, prefix + selected, lineStart + prefix.length, lineStart + prefix.length + selected.length)
}

function sv_insertLink() {
  const url = window.prompt('输入链接 URL')?.trim()
  if (!url) return
  const { from, to, selectedText } = getCmSelection()
  const text = selectedText || '链接文字'
  const md = `[${text}](${url})`
  cmInsert(from, to, md, from + 1, from + 1 + text.length)
}

function sv_insertTable() {
  const { from } = getCmSelection()
  const table = '\n| 标题 | 标题 | 标题 |\n| --- | --- | --- |\n| 内容 | 内容 | 内容 |\n'
  cmInsert(from, from, table)
}

function sv_insertHr() {
  const { from } = getCmSelection()
  cmInsert(from, from, '\n---\n')
}

function sv_insertCodeBlock() {
  const { from, to, selectedText } = getCmSelection()
  const body = selectedText || '代码'
  const newText = `\n\`\`\`txt\n${body}\n\`\`\`\n`
  const bodyStart = from + '\n```txt\n'.length
  cmInsert(from, to, newText, bodyStart, bodyStart + body.length)
}

function sv_undo() {
  if (!cmView) return
  undo({ state: cmView.state, dispatch: cmView.dispatch.bind(cmView) })
  cmView.focus()
}

function sv_redo() {
  if (!cmView) return
  redo({ state: cmView.state, dispatch: cmView.dispatch.bind(cmView) })
  cmView.focus()
}

function sv_clearContent() {
  if (!cmView) return
  if (!window.confirm('确认清空内容？')) return
  const current = cmView.state.doc.toString()
  if (!current) return
  cmInsert(0, current.length, '')
}

function lineNumberExtensions() {
  return showLineNumbers.value ? [lineNumbers(), highlightActiveLineGutter()] : []
}

function toggleLineNumbers() {
  if (!cmView) return
  showLineNumbers.value = !showLineNumbers.value
  cmView.dispatch({
    effects: lineNumberCompartment.reconfigure(lineNumberExtensions()),
  })
}

function insertEmbed(kind: 'post' | 'music' | 'video') {
  const labels = { post: '文章', music: '音乐/专辑', video: '视频' }
  const id = window.prompt(`输入要引用的${labels[kind]} UUID`)?.trim()
  if (!id) return
  const { from } = getCmSelection()
  const md = `\n:::${kind}{id="${id}"}\n:::\n`
  cmInsert(from, from, md)
}

// ── 图片上传 ─────────────────────────────────────────────
const imageInputRef = ref<HTMLInputElement | null>(null)
const plainImageInputRef = ref<HTMLInputElement | null>(null)
const plainTextareaRef = ref<HTMLTextAreaElement | null>(null)
const uploadingImage = ref(false)

function triggerImageUpload() {
  imageInputRef.value?.click()
}

async function handleImageUploadFile(e: Event) {
  const file = (e.target as HTMLInputElement).files?.[0]
  if (!file) return
  if (imageInputRef.value) imageInputRef.value.value = ''
  await uploadImage(file)
}

function triggerPlainImageUpload() {
  plainImageInputRef.value?.click()
}

async function handlePlainImageUploadFile(e: Event) {
  const file = (e.target as HTMLInputElement).files?.[0]
  if (!file) return
  if (plainImageInputRef.value) plainImageInputRef.value.value = ''
  await uploadImagePlain(file)
}

async function uploadImagePlain(file: File) {
  const placeholder = `![上传中]()`
  const ta = plainTextareaRef.value
  const insertStart = ta ? (ta.selectionStart ?? plainValue.value.length) : plainValue.value.length
  const insertEnd = ta ? (ta.selectionEnd ?? insertStart) : insertStart

  // 插入占位符
  plainValue.value = plainValue.value.slice(0, insertStart) + placeholder + plainValue.value.slice(insertEnd)
  emit('update:modelValue', plainValue.value)
  await nextTick()
  if (ta) {
    ta.focus()
    ta.selectionStart = ta.selectionEnd = insertStart + placeholder.length
  }

  uploadingImage.value = true
  try {
    const formData = new FormData()
    formData.append('image', file)
    const res = await fetch('/api/blog/upload-image', {
      method: 'POST',
      headers: authStore.token ? { Authorization: `Bearer ${authStore.token}` } : {},
      body: formData,
    })
    if (!res.ok) throw new Error('upload failed')
    const data = await res.json()
    const url: string = data.url
    const md = `![图片](${url})`

    // 替换占位符
    const idx = plainValue.value.indexOf(placeholder)
    if (idx !== -1) {
      plainValue.value = plainValue.value.slice(0, idx) + md + plainValue.value.slice(idx + placeholder.length)
    } else {
      plainValue.value += md
    }
    emit('update:modelValue', plainValue.value)
    await nextTick()
    const textarea = plainTextareaRef.value
    if (textarea) {
      textarea.focus()
      const cursorPos = idx !== -1 ? idx + md.length : plainValue.value.length
      textarea.selectionStart = textarea.selectionEnd = cursorPos
    }
  } catch (err) {
    console.error('Image upload failed', err)
    // 移除占位符
    const idx = plainValue.value.indexOf(placeholder)
    if (idx !== -1) {
      plainValue.value = plainValue.value.slice(0, idx) + plainValue.value.slice(idx + placeholder.length)
      emit('update:modelValue', plainValue.value)
    }
  } finally {
    uploadingImage.value = false
  }
}

async function uploadImage(file: File) {
  if (!cmView) return
  const uploadId = Math.random().toString(36).slice(2, 8)
  const placeholder = `![上传中-${uploadId}]()`

  // 插入占位符
  const { from } = getCmSelection()
  cmInsert(from, from, placeholder)

  uploadingImage.value = true
  try {
    const formData = new FormData()
    formData.append('image', file)
    const res = await fetch('/api/blog/upload-image', {
      method: 'POST',
      headers: authStore.token ? { Authorization: `Bearer ${authStore.token}` } : {},
      body: formData,
    })
    if (!res.ok) throw new Error('upload failed')
    const data = await res.json()
    const url: string = data.url

    // 替换占位符
    const doc = cmView.state.doc.toString()
    const idx = doc.indexOf(placeholder)
    if (idx !== -1) {
      const finalMd = `![图片](${url})`
      cmView.dispatch({
        changes: { from: idx, to: idx + placeholder.length, insert: finalMd },
      })
    }
  } catch (err) {
    console.error('Image upload failed', err)
    // 移除占位符
    const doc = cmView.state.doc.toString()
    const idx = doc.indexOf(placeholder)
    if (idx !== -1) {
      cmView.dispatch({ changes: { from: idx, to: idx + placeholder.length, insert: '' } })
    }
  } finally {
    uploadingImage.value = false
  }
}

// ── 粘贴图片 ─────────────────────────────────────────────
function onCmPaste(e: ClipboardEvent) {
  if (!props.enableImageUpload) return
  const files = Array.from(e.clipboardData?.files ?? []).filter(f => f.type.startsWith('image/'))
  if (files.length === 0) return
  e.preventDefault()
  files.forEach(f => uploadImage(f))
}

// ── 拖拽图片 ─────────────────────────────────────────────
const isDragging = ref(false)

function onDragOver() { isDragging.value = true }
function onDragLeave() { isDragging.value = false }
function onDrop(e: DragEvent) {
  isDragging.value = false
  if (!props.enableImageUpload) return
  handleDropFiles(e.dataTransfer?.files)
}

function handleDropFiles(files?: FileList | null) {
  if (!files) return
  const imageFiles = Array.from(files).filter(f => f.type.startsWith('image/'))
  imageFiles.forEach(f => uploadImage(f))
}

// ── @提及 ────────────────────────────────────────────────
const mention = ref({
  visible: false,
  query: '',
  index: 0,
  x: 0,
  y: 0,
  results: [] as MentionUser[],
  startPos: -1,
})

let mentionDebounce: ReturnType<typeof setTimeout> | null = null

function detectMentionFromCm(update: import('@codemirror/view').ViewUpdate) {
  if (!cmView) return
  const pos = update.state.selection.main.head
  const doc = update.state.doc
  const line = doc.lineAt(pos)
  const textBefore = line.text.slice(0, pos - line.from)
  const match = textBefore.match(/@([\w一-龥]*)$/)

  if (!match) { closeMention(); return }

  const query = match[1]
  mention.value.startPos = pos - match[0].length
  mention.value.query = query

  // 定位下拉（相对于容器）
  const coords = cmView.coordsAtPos(mention.value.startPos)
  if (coords) {
    mention.value.x = coords.left
    mention.value.y = coords.bottom + 4
  }

  if (mentionDebounce) clearTimeout(mentionDebounce)
  mentionDebounce = setTimeout(() => fetchMentionUsers(query), 120)
}

async function fetchMentionUsers(q: string) {
  try {
    const headers: Record<string, string> = {}
    if (authStore.token) headers.Authorization = `Bearer ${authStore.token}`
    const res = await fetch(`/api/users/search?scope=mention&q=${encodeURIComponent(q)}&limit=5`, { headers })
    if (!res.ok) return
    const data = await res.json()
    mention.value.results = data.data || []
    mention.value.visible = mention.value.results.length > 0
    mention.value.index = 0
  } catch { /* ignore */ }
}

function applyMention(user: MentionUser) {
  if (!cmView) return
  const pos = cmView.state.selection.main.head
  const insertText = `@${user.username}`
  cmView.dispatch({
    changes: { from: mention.value.startPos, to: pos, insert: insertText },
    selection: { anchor: mention.value.startPos + insertText.length },
  })
  cmView.focus()
  closeMention()
}

function closeMention() {
  mention.value.visible = false
  mention.value.results = []
  mention.value.startPos = -1
}

// Mention keyboard nav via CodeMirror keymap — handled via global keydown on container
function onContainerKeydown(e: KeyboardEvent) {
  if (!mention.value.visible) return
  const items = mention.value.results
  if (e.key === 'ArrowDown') { e.preventDefault(); mention.value.index = (mention.value.index + 1) % items.length }
  else if (e.key === 'ArrowUp') { e.preventDefault(); mention.value.index = (mention.value.index - 1 + items.length) % items.length }
  else if (e.key === 'Enter' || e.key === 'Tab') { e.preventDefault(); applyMention(items[mention.value.index]) }
  else if (e.key === 'Escape') closeMention()
}

// ── Lifecycle ────────────────────────────────────────────
onBeforeUnmount(() => {
  if (mentionDebounce) clearTimeout(mentionDebounce)
  provider?.destroy()
  ydoc?.destroy()
  cmView?.destroy()
})
</script>

<style scoped>
.a-editor {
  width: 100%;
  display: flex;
  flex-direction: column;
  background: #fff;
  position: relative;
}

.a-editor:not(.no-border) {
  border: 2px solid #000;
}

/* ── Presence ────────────────────────────────────────── */
.a-editor-presence {
  display: flex;
  align-items: center;
  gap: 0.75rem;
  padding: 0.4rem 1.25rem;
  border-bottom: 2px solid #000;
  background: #f9f9f9;
  flex-shrink: 0;
}

.presence-avatars { display: flex; gap: 0.4rem; }

.presence-dot {
  width: 1.6rem;
  height: 1.6rem;
  border-radius: 9999px;
  display: flex;
  align-items: center;
  justify-content: center;
  color: #fff;
  font-weight: 900;
  font-size: 0.65rem;
  border: 2px solid #000;
  flex-shrink: 0;
}

/* ── Toolbar sticky ──────────────────────────────────── */
.a-editor-toolbar {
  display: flex;
  flex-wrap: wrap;
  align-items: center;
  gap: 0.25rem;
  padding: 0.6rem 1rem;
  border-bottom: 2px solid #000;
  background: #fff;
  flex-shrink: 0;
  position: sticky;
  top: 0;
  z-index: 10;
}

.tb-btn {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  padding: 0.35rem 0.6rem;
  border: 2px solid #000;
  background: #fff;
  color: #000;
  font-size: 0.72rem;
  font-weight: 900;
  font-family: inherit;
  cursor: pointer;
  letter-spacing: 0.02em;
  line-height: 1;
  white-space: nowrap;
  transition: none;
}

.tb-btn:hover,
.tb-btn.active {
  background: #000;
  color: #fff;
}

.tb-btn.uploading {
  opacity: 0.5;
  cursor: not-allowed;
}

.tb-sep {
  display: inline-block;
  width: 1px;
  height: 1.2rem;
  background: #d1d5db;
  margin: 0 0.2rem;
  flex-shrink: 0;
}

.tb-hidden-input { display: none; }

/* ── SV body ─────────────────────────────────────────── */
.a-editor-sv-body {
  display: grid;
  grid-template-columns: minmax(0, 1fr) minmax(0, 1fr);
  flex: 1;
  min-height: 0;
  position: relative;
}

.a-editor-sv-body.dragging::after {
  content: '松开鼠标上传图片';
  position: absolute;
  inset: 0;
  background: rgba(0, 0, 0, 0.08);
  border: 3px dashed #000;
  display: flex;
  align-items: center;
  justify-content: center;
  font-weight: 900;
  font-size: 1rem;
  color: #000;
  pointer-events: none;
  z-index: 5;
}

.sv-pane {
  min-height: 0;
  overflow: auto;
}

.sv-source {
  border-right: 2px solid #000;
  display: flex;
  flex-direction: column;
}

.cm-container {
  flex: 1;
  height: 100%;
  min-height: 16rem;
}

/* Override CM internal height */
:deep(.cm-editor) {
  height: 100%;
}
:deep(.cm-scroller) {
  overflow: auto;
}

.sv-preview {
  padding: 1.5rem 1.5rem 2rem;
  background: #fff;
  overflow: auto;
  min-height: 16rem;
}

/* ── Plain mode ──────────────────────────────────────── */
.a-editor-plain-toolbar {
  display: flex;
  align-items: center;
  gap: 0.25rem;
  padding: 0.4rem 0.75rem;
  border-bottom: 2px solid #000;
  background: #fff;
}

.a-editor-plain-textarea {
  width: 100%;
  min-height: 8rem;
  border: none;
  padding: 1rem 1.25rem;
  resize: vertical;
  outline: none;
  font-family: inherit;
  font-size: 0.9rem;
  line-height: 1.75;
  box-sizing: border-box;
  background: #fff;
}

/* ── @提及下拉 ───────────────────────────────────────── */
.a-mention-dropdown {
  position: fixed;
  z-index: 9999;
  background: #fff;
  border: 2px solid #000;
  min-width: 200px;
  box-shadow: 6px 6px 0 0 rgba(0, 0, 0, 1);
  max-height: 200px;
  overflow-y: auto;
}

.mention-item {
  display: flex;
  align-items: center;
  gap: 0.6rem;
  width: 100%;
  padding: 0.6rem 0.9rem;
  border: none;
  background: #fff;
  text-align: left;
  cursor: pointer;
  font-family: inherit;
}

.mention-item:hover,
.mention-item.is-active {
  background: #000;
  color: #fff;
}

.mention-name {
  font-size: 0.82rem;
  font-weight: 800;
}

.mention-username {
  font-size: 0.72rem;
  color: #9ca3af;
}

.mention-item.is-active .mention-username {
  color: #d1d5db;
}

/* ── 响应式 ──────────────────────────────────────────── */
@media (max-width: 700px) {
  .a-editor-sv-body {
    grid-template-columns: 1fr;
  }

  .sv-source {
    border-right: none;
    border-bottom: 2px solid #000;
  }

  .cm-container {
    min-height: 10rem;
  }
}
</style>
