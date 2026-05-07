<template>
<div class="markdown-editor-wrapper" :class="{ 'fullscreen': isFullscreen }">
<div class="markdown-editor">

  <div class="editor-title-bar" v-if="!hideTitle">
    <input
      type="text"
      :value="localTitle"
      @input="onTitleInput"
      :placeholder="titlePlaceholder"
      class="title-input"
    />
  </div>

  <div class="editor-toolbar">
    <div class="toolbar-group">
      <button
        v-for="(btn, idx) in basicButtons"
        :key="idx"
        type="button"
        @click="btn.action"
        class="toolbar-btn"
        :title="btn.title"
      >{{ btn.label }}</button>
    </div>
    <div class="toolbar-divider"></div>
    <div class="toolbar-group">
      <button
        v-for="(btn, idx) in headingButtons"
        :key="idx"
        type="button"
        @click="btn.action"
        class="toolbar-btn toolbar-btn-heading"
        :title="btn.title"
      >{{ btn.label }}</button>
    </div>
    <div class="toolbar-divider"></div>
    <div class="toolbar-group">
      <button
        v-for="(btn, idx) in codeButtons"
        :key="idx"
        type="button"
        @click="btn.action"
        class="toolbar-btn toolbar-btn-code"
        :title="btn.title"
      >{{ btn.label }}</button>
    </div>
    <div class="toolbar-divider"></div>
    <div class="toolbar-group">
      <button
        type="button"
        @click="triggerImageUpload"
        class="toolbar-btn toolbar-btn-code"
        :disabled="imageUploader.isUploading.value"
        title="插入图片"
      >图片</button>
    </div>
    <div class="toolbar-spacer"></div>
    <div class="toolbar-group mode-switch">
      <button
        type="button"
        @click="setMode('source')"
        class="toolbar-btn toolbar-btn-mode"
        :class="{ active: mode === 'source' }"
        title="源码分屏模式"
      >源码</button>
      <button
        type="button"
        @click="setMode('render')"
        class="toolbar-btn toolbar-btn-mode"
        :class="{ active: mode === 'render' }"
        title="单栏所见所得模式"
      >所见</button>
    </div>
    <div class="toolbar-group">
      <button
        type="button"
        @click="toggleFullscreen"
        class="toolbar-btn toolbar-btn-icon"
        :title="isFullscreen ? '退出全屏' : '全屏模式'"
      >⛶</button>
    </div>
  </div>

  <div
    class="editor-content-wrapper"
    @dragover.prevent
    @drop="onDrop"
  >
    <div v-show="mode === 'source'" class="editor-panel">
      <div ref="editorContainerRef" class="codemirror-container"></div>
    </div>
    <div v-show="mode === 'source'" class="source-split-divider"></div>
    <div v-show="mode === 'source'" class="source-preview-panel" ref="sourcePreviewPanelRef">
      <MarkdownRender
        v-if="localContent.trim()"
        :content="localContent"
        custom-id="source-preview"
        class="markstream-preview"
      />
      <p v-else style="color:#9ca3af;padding:1rem">预览将显示在此处...</p>
    </div>

    <MarkdownEditorWysiwyg
      v-show="mode === 'render'"
      ref="wysiwygEditorRef"
      :model-value="localContent"
      :placeholder="placeholder"
      @update:model-value="onRenderContentUpdate"
    />
  </div>

  <div v-if="imageUploader.isUploading.value" class="upload-progress-bar">
    <div class="upload-progress-fill" :style="{ width: imageUploader.uploadProgress.value + '%' }"></div>
    <span class="upload-progress-text">上传中 {{ imageUploader.uploadProgress.value }}%</span>
  </div>

  <div class="editor-status-bar">
    <div class="status-left">
      <span class="status-item">字数：{{ wordCount }}</span>
      <span class="status-divider">|</span>
      <span class="status-item">阅读时间：{{ readingTime }}</span>
    </div>
    <div class="status-right">
      <span v-if="imageUploader.uploadError.value" class="status-item error">
        {{ imageUploader.uploadError.value }}
      </span>
      <span class="status-item saved">{{ autoSaved ? '已保存 ✓' : '未保存' }}</span>
    </div>
  </div>

  <input
    ref="imageFileInputRef"
    type="file"
    accept="image/jpeg,image/png,image/gif,image/webp"
    class="hidden"
    @change="imageUploader.handleFileInput"
  />

</div>
</div>
</template>

<script setup lang="ts">
import { ref, watch, onMounted, onBeforeUnmount, computed, nextTick, shallowRef } from 'vue'
import {
  EditorView,
  lineNumbers,
  keymap,
  highlightActiveLine,
} from '@codemirror/view'
import { EditorState } from '@codemirror/state'
import { defaultKeymap, historyKeymap, history, indentWithTab } from '@codemirror/commands'
import { markdown } from '@codemirror/lang-markdown'
import { languages } from '@codemirror/language-data'
import 'highlight.js/styles/github-dark.css'
import MarkdownRender from 'markstream-vue'
import 'markstream-vue/index.css'
import MarkdownEditorWysiwyg from './MarkdownEditorWysiwyg.vue'
import { useAutoSave } from './composables/useAutoSave'
import { useImageUpload } from './composables/useImageUpload'

interface Props {
  modelValue?: string
  hideTitle?: boolean
  titleText?: string
  titlePlaceholder?: string
  postId?: string
  placeholder?: string
}

interface MarkdownEditorWysiwygExposed {
  focusEnd: () => void
  getMarkdown: () => string
  setContent: (markdown: string) => void
  insertMarkdownAtCursor: (before: string, after: string, placeholder: string) => void
  insertLinePrefixAtCursor: (prefix: string) => void
  replaceInMarkdown: (search: string, replacement: string) => void
  mount: () => void
  unmount: () => void
}

const props = withDefaults(defineProps<Props>(), {
  modelValue: '',
  hideTitle: false,
  titleText: '',
  titlePlaceholder: '请输入标题',
  postId: '',
  placeholder: '',
})

const emit = defineEmits<{
  'update:modelValue': [value: string]
  'update:titleText': [value: string]
}>()

const editorContainerRef = ref<HTMLDivElement | null>(null)
const sourcePreviewPanelRef = ref<HTMLElement | null>(null)
const imageFileInputRef = ref<HTMLInputElement | null>(null)
const wysiwygEditorRef = ref<MarkdownEditorWysiwygExposed | null>(null)
const editorView = shallowRef<EditorView | null>(null)
const localContent = ref(props.modelValue)
const localTitle = ref(props.titleText)
const isFullscreen = ref(false)
const mode = ref<'source' | 'render'>('source')

function emitContent(value: string) {
  localContent.value = value
  emit('update:modelValue', value)
}

function withWysiwygEditor(callback: (editor: MarkdownEditorWysiwygExposed) => void) {
  const editor = wysiwygEditorRef.value
  if (editor) callback(editor)
}

function onRenderContentUpdate(value: string) {
  emitContent(value)
}

const { autoSaved, triggerAutoSave, checkForDraft, clearDraft } = useAutoSave(
  () => localContent.value,
  () => localTitle.value,
  () => props.postId,
  (content, title) => {
    localContent.value = content
    localTitle.value = title
    emit('update:modelValue', content)
    emit('update:titleText', title)
    syncToActiveEditor()
  },
)

watch(localContent, triggerAutoSave)
watch(localTitle, triggerAutoSave)

const imageUploader = useImageUpload(
  (text: string) => {
    if (mode.value === 'source') {
      const view = editorView.value
      if (!view) return
      const { from, to } = view.state.selection.main
      view.dispatch({
        changes: { from, to, insert: text },
        selection: { anchor: from + text.length },
      })
      view.focus()
      emitContent(view.state.doc.toString())
    } else {
      withWysiwygEditor((editor) => {
        editor.insertMarkdownAtCursor('', '', text)
      })
    }
  },
  (search: string, replacement: string) => {
    if (mode.value === 'source') {
      const view = editorView.value
      if (!view) return
      const doc = view.state.doc.toString()
      const idx = doc.indexOf(search)
      if (idx >= 0) {
        view.dispatch({
          changes: { from: idx, to: idx + search.length, insert: replacement },
        })
        emitContent(view.state.doc.toString())
      }
    } else {
      withWysiwygEditor((editor) => {
        editor.replaceInMarkdown(search, replacement)
      })
    }
  },
)

function triggerImageUpload() {
  if (imageFileInputRef.value) {
    imageUploader.openFilePicker(imageFileInputRef.value)
  }
}

function onDrop(e: DragEvent) {
  imageUploader.handleDrop(e)
}

function onEditorPaste(e: ClipboardEvent) {
  imageUploader.handlePaste(e)
}

function mountRenderEditor() {
  withWysiwygEditor((editor) => {
    editor.mount()
    if (editor.getMarkdown() !== localContent.value) {
      editor.setContent(localContent.value)
    }
  })
}

function setMode(newMode: 'source' | 'render') {
  if (mode.value === newMode) return

  if (mode.value === 'source' && editorView.value) {
    localContent.value = editorView.value.state.doc.toString()
  } else if (mode.value === 'render') {
    withWysiwygEditor((editor) => {
      emitContent(editor.getMarkdown())
      editor.unmount()
    })

    const view = editorView.value
    if (view) {
      const current = view.state.doc.toString()
      if (current !== localContent.value) {
        view.dispatch({ changes: { from: 0, to: current.length, insert: localContent.value } })
      }
    }
  }

  mode.value = newMode

  if (newMode === 'render') {
    nextTick(() => {
      mountRenderEditor()
      withWysiwygEditor((editor) => {
        editor.focusEnd()
      })
    })
  } else {
    nextTick(() => editorView.value?.focus())
  }
}

function syncToActiveEditor() {
  if (mode.value === 'source') {
    const view = editorView.value
    if (view && view.state.doc.toString() !== localContent.value) {
      view.dispatch({
        changes: { from: 0, to: view.state.doc.length, insert: localContent.value },
      })
    }
  } else {
    mountRenderEditor()
  }
}

function wrap(before: string, after: string, placeholder: string) {
  if (mode.value === 'source') {
    const view = editorView.value
    if (!view) return
    const { from, to } = view.state.selection.main
    const selected = view.state.sliceDoc(from, to) || placeholder
    view.dispatch({
      changes: { from, to, insert: before + selected + after },
      selection: { anchor: from + before.length, head: from + before.length + selected.length },
    })
    view.focus()
    emitContent(view.state.doc.toString())
  } else {
    withWysiwygEditor((editor) => {
      editor.insertMarkdownAtCursor(before, after, placeholder)
    })
  }
}

function insertLine(prefix: string) {
  if (mode.value === 'source') {
    const view = editorView.value
    if (!view) return
    const { from } = view.state.selection.main
    const line = view.state.doc.lineAt(from)
    view.dispatch({
      changes: { from: line.from, to: line.from, insert: prefix },
      selection: { anchor: line.from + prefix.length },
    })
    view.focus()
    emitContent(view.state.doc.toString())
  } else {
    withWysiwygEditor((editor) => {
      editor.insertLinePrefixAtCursor(prefix)
    })
  }
}

const basicButtons = [
  { label: '粗体', title: '粗体 (Ctrl+B)', action: () => wrap('**', '**', '粗体文字') },
  { label: '斜体', title: '斜体', action: () => wrap('*', '*', '斜体文字') },
  { label: '删除线', title: '删除线', action: () => wrap('~~', '~~', '删除线') },
]

const headingButtons = [
  { label: 'H1', title: '一级标题', action: () => insertLine('# ') },
  { label: 'H2', title: '二级标题', action: () => insertLine('## ') },
  { label: 'H3', title: '三级标题', action: () => insertLine('### ') },
]

const codeButtons = [
  { label: '行内代码', title: '行内代码', action: () => wrap('`', '`', 'code') },
  { label: '代码块', title: '代码块', action: () => wrap('```\n', '\n```', 'code') },
  { label: '引用', title: '引用', action: () => insertLine('> ') },
  { label: '分割线', title: '分割线', action: () => insertLine('\n---\n') },
  { label: '链接', title: '链接', action: () => wrap('[', '](url)', '链接文字') },
]

const wordCount = computed(() => {
  const text = localContent.value.trim()
  if (!text) return 0
  const chinese = (text.match(/[一-龥]/g) || []).length
  const english = (text.match(/[a-zA-Z]+/g) || []).length
  return chinese + english
})

const readingTime = computed(() => {
  const count = wordCount.value
  const minutes = Math.ceil(count / 400)
  return minutes < 1 ? '< 1分钟' : `${minutes}分钟`
})

function onTitleInput(e: Event) {
  const target = e.target as HTMLInputElement
  localTitle.value = target.value
  emit('update:titleText', target.value)
}

function toggleFullscreen() {
  isFullscreen.value = !isFullscreen.value
  nextTick(() => {
    if (mode.value === 'source') {
      editorView.value?.focus()
    } else {
      mountRenderEditor()
      withWysiwygEditor((editor) => {
        editor.focusEnd()
      })
    }
  })
}

function onEditorScroll() {
  const scroller = editorView.value?.scrollDOM
  const preview = sourcePreviewPanelRef.value
  if (!scroller || !preview) return
  const scrollableHeight = scroller.scrollHeight - scroller.clientHeight
  if (scrollableHeight <= 0) return
  const ratio = scroller.scrollTop / scrollableHeight
  preview.scrollTop = ratio * (preview.scrollHeight - preview.clientHeight)
}

onMounted(() => {
  if (!editorContainerRef.value) return

  checkForDraft()

  const updateListener = EditorView.updateListener.of((update) => {
    if (update.docChanged) {
      emitContent(update.state.doc.toString())
    }
  })

  editorView.value = new EditorView({
    state: EditorState.create({
      doc: localContent.value,
      extensions: [
        lineNumbers(),
        history(),
        highlightActiveLine(),
        EditorView.lineWrapping,
        markdown({ codeLanguages: languages }),
        updateListener,
        keymap.of([...defaultKeymap, ...historyKeymap, indentWithTab]),
        EditorView.theme({
          '&': { height: '100%' },
          '.cm-scroller': {
            overflow: 'auto',
            fontFamily: 'JetBrains Mono, Fira Code, monospace',
            fontSize: '0.9375rem',
          },
          '.cm-line': { padding: '0 4px' },
          '.cm-gutters': { backgroundColor: '#f9fafb', borderRight: '2px solid #000' },
        }),
        EditorView.domEventHandlers({
          paste: (e) => {
            onEditorPaste(e)
          },
        }),
      ],
    }),
    parent: editorContainerRef.value,
  })

  editorView.value.scrollDOM.addEventListener('scroll', onEditorScroll, { passive: true })
})

onBeforeUnmount(() => {
  editorView.value?.scrollDOM.removeEventListener('scroll', onEditorScroll)
  editorView.value?.destroy()
  withWysiwygEditor((editor) => {
    editor.unmount()
  })
})

watch(
  () => props.modelValue,
  (newVal) => {
    if (newVal === localContent.value) return
    localContent.value = newVal
    syncToActiveEditor()
  },
)

defineExpose({ clearDraft })
</script>

<style scoped>
.markdown-editor-wrapper {
  position: relative;
  width: 100%;
  height: 100%;
  min-height: 600px;
}

.markdown-editor-wrapper.fullscreen {
  position: fixed;
  top: 0;
  left: 0;
  right: 0;
  bottom: 0;
  z-index: 9999;
  background: white;
}

.markdown-editor {
  display: flex;
  flex-direction: column;
  height: 100%;
  background: #fff;
  border: 2px solid #000;
}

.editor-title-bar {
  border-bottom: 2px solid #000;
  padding: 1rem;
  background: #fff;
  flex-shrink: 0;
}

.title-input {
  width: 100%;
  font-size: 1.5rem;
  font-weight: 900;
  border: 2px solid #000;
  padding: 0.75rem 1rem;
  outline: none;
  transition: box-shadow 0.2s;
}

.title-input:focus {
  box-shadow: 5px 5px 0 0 rgba(0, 0, 0, 1);
}

.title-input::placeholder {
  color: #9ca3af;
  font-weight: 700;
}

.editor-toolbar {
  display: flex;
  align-items: center;
  gap: 0.5rem;
  padding: 0.5rem 1rem;
  border-bottom: 2px solid #000;
  background: #fff;
  flex-shrink: 0;
  flex-wrap: wrap;
}

.toolbar-group {
  display: flex;
  gap: 0.375rem;
}

.toolbar-divider {
  width: 2px;
  height: 2rem;
  background: #000;
  margin: 0 0.5rem;
}

.toolbar-spacer {
  flex: 1;
}

.mode-switch {
  border: 2px solid #000;
  gap: 0;
}

.toolbar-btn-mode {
  border: 0;
  border-right: 2px solid #000;
}

.toolbar-btn-mode:last-child {
  border-right: 0;
}

.toolbar-btn {
  padding: 0.5rem 0.875rem;
  font-size: 0.8125rem;
  font-weight: 700;
  text-transform: uppercase;
  letter-spacing: 0.05em;
  border: 2px solid #000;
  background: #fff;
  cursor: pointer;
  transition: all 0.15s;
}

.toolbar-btn:hover {
  background: #000;
  color: #fff;
}

.toolbar-btn:active {
  transform: translate(1px, 1px);
}

.toolbar-btn.active {
  background: #000;
  color: #fff;
}

.toolbar-btn:disabled {
  opacity: 0.5;
  cursor: not-allowed;
}

.mode-switch .toolbar-btn-mode {
  border: 0;
  border-right: 2px solid #000;
}

.mode-switch .toolbar-btn-mode:last-child {
  border-right: 0;
}

.toolbar-btn-heading {
  font-family: monospace;
  font-size: 0.75rem;
}

.toolbar-btn-code {
  font-family: monospace;
}

.toolbar-btn-icon {
  font-size: 1rem;
  padding: 0.5rem 0.625rem;
}

.editor-content-wrapper {
  display: flex;
  flex: 1;
  overflow: hidden;
  background: #f9fafb;
}

.editor-panel {
  flex: 1;
  display: flex;
  overflow: hidden;
  background: #fff;
}

.source-split-divider {
  width: 2px;
  background: #000;
  flex-shrink: 0;
}

.source-preview-panel {
  flex: 1;
  overflow-y: auto;
  padding: 1.5rem 2rem;
  background: #fafafa;
}

.codemirror-container {
  flex: 1;
  height: 100%;
  overflow: hidden;
}

.markstream-preview {
  font-size: 0.9375rem;
  line-height: 1.75;
}

.upload-progress-bar {
  position: relative;
  height: 1.75rem;
  background: #f3f4f6;
  border-top: 2px solid #000;
  overflow: hidden;
  flex-shrink: 0;
  display: flex;
  align-items: center;
}

.upload-progress-fill {
  position: absolute;
  left: 0;
  top: 0;
  bottom: 0;
  background: #000;
  transition: width 0.2s;
}

.upload-progress-text {
  position: relative;
  z-index: 1;
  font-size: 0.75rem;
  font-weight: 700;
  color: #fff;
  mix-blend-mode: difference;
  padding: 0 0.75rem;
}

.editor-status-bar {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 0.5rem 1rem;
  border-top: 2px solid #000;
  background: #f9fafb;
  font-size: 0.8125rem;
  font-weight: 700;
  flex-shrink: 0;
}

.status-left {
  display: flex;
  align-items: center;
  gap: 0.75rem;
}

.status-right {
  display: flex;
  align-items: center;
  gap: 0.75rem;
}

.status-item {
  color: #374151;
}

.status-divider {
  color: #9ca3af;
}

.status-item.saved {
  color: #059669;
}

.status-item.error {
  color: #dc2626;
  font-size: 0.75rem;
}

.hidden {
  display: none;
}

@media (max-width: 768px) {
  .editor-toolbar {
    gap: 0.25rem;
    padding: 0.375rem 0.5rem;
  }

  .toolbar-divider {
    display: none;
  }

  .title-input {
    font-size: 1.25rem;
  }
}
</style>
