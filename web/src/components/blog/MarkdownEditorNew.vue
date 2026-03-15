<template>
  <div class="md-editor-wrap">
    <!-- Left: CodeMirror editor -->
    <div class="md-editor-left" ref="editorContainerRef"></div>

    <!-- Right: TOC + Preview -->
    <div class="md-editor-right" :class="{ 'mobile-visible': mobilePreviewOpen }">
      <EditorTOC :content="modelValue" :preview-el="previewRef" />
      <div class="md-preview-pane" ref="previewRef" v-html="renderedHtml" />
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, watch, onMounted, onBeforeUnmount, shallowRef } from 'vue'
import {
  EditorView,
  lineNumbers,
  keymap,
  Decoration,
  ViewPlugin,
  type DecorationSet,
  type ViewUpdate,
} from '@codemirror/view'
import { EditorState } from '@codemirror/state'
import { defaultKeymap, historyKeymap, history, indentWithTab } from '@codemirror/commands'
import { markdown } from '@codemirror/lang-markdown'
import { languages } from '@codemirror/language-data'
import { RangeSetBuilder } from '@codemirror/state'
import EditorTOC from '@/components/blog/EditorTOC.vue'
import { useMarkdownRenderer } from '@/composables/useMarkdownRenderer'
import { useAuthStore } from '@/stores/auth'
import { useApi } from '@/composables/useApi'

const props = defineProps<{
  modelValue: string
  placeholder?: string
}>()

const emit = defineEmits<{
  'update:modelValue': [value: string]
}>()

const editorContainerRef = ref<HTMLDivElement | null>(null)
const previewRef = ref<HTMLDivElement | null>(null)
const mobilePreviewOpen = ref(false)
const editorView = shallowRef<EditorView | null>(null)

const authStore = useAuthStore()
const api = useApi()
const { renderMarkdown } = useMarkdownRenderer()

const renderedHtml = computed(() => renderMarkdown(props.modelValue))

// ─── Heading decoration extension ──────────────────────────────────────────
const headingLinePlugin = ViewPlugin.fromClass(
  class {
    decorations: DecorationSet
    constructor(view: EditorView) {
      this.decorations = buildDecorations(view)
    }
    update(update: ViewUpdate) {
      if (update.docChanged || update.viewportChanged) {
        this.decorations = buildDecorations(update.view)
      }
    }
  },
  { decorations: (v) => v.decorations },
)

function buildDecorations(view: EditorView): DecorationSet {
  const builder = new RangeSetBuilder<Decoration>()
  for (const { from, to } of view.visibleRanges) {
    let pos = from
    while (pos <= to) {
      const line = view.state.doc.lineAt(pos)
      const text = line.text
      let cls: string | null = null
      if (/^# /.test(text)) cls = 'cm-title-h1'
      else if (/^## /.test(text)) cls = 'cm-title-h2'
      else if (/^### /.test(text)) cls = 'cm-title-h3'
      else if (/^#### /.test(text)) cls = 'cm-title-h4'
      if (cls) {
        builder.add(line.from, line.from, Decoration.line({ class: cls }))
      }
      pos = line.to + 1
    }
  }
  return builder.finish()
}

// ─── Image upload helper ────────────────────────────────────────────────────
async function uploadImageFile(file: File): Promise<string | null> {
  const allowed = ['image/jpeg', 'image/png', 'image/gif', 'image/webp']
  if (!allowed.includes(file.type)) return null
  if (file.size > 5 * 1024 * 1024) return null

  const fd = new FormData()
  fd.append('image', file)

  try {
    const res = await fetch(api.blog.uploadImage, {
      method: 'POST',
      headers: authStore.token ? { Authorization: `Bearer ${authStore.token}` } : {},
      body: fd,
    })
    if (res.ok) {
      const d = await res.json()
      return d.url as string
    }
  } catch {}
  return null
}

function insertAtCursor(view: EditorView, text: string) {
  const { from } = view.state.selection.main
  view.dispatch({
    changes: { from, to: from, insert: text },
    selection: { anchor: from + text.length },
  })
}

// ─── Paste handler ──────────────────────────────────────────────────────────
function handlePaste(e: ClipboardEvent, view: EditorView) {
  const items = e.clipboardData?.items
  if (!items) return false
  for (const item of Array.from(items)) {
    if (item.kind === 'file' && item.type.startsWith('image/')) {
      const file = item.getAsFile()
      if (!file) continue
      e.preventDefault()
      uploadImageFile(file).then((url) => {
        if (url) insertAtCursor(view, `![](${url})`)
      })
      return true
    }
  }
  return false
}

// ─── Drop handler ───────────────────────────────────────────────────────────
function handleDrop(e: DragEvent, view: EditorView) {
  const files = e.dataTransfer?.files
  if (!files) return false
  for (const file of Array.from(files)) {
    if (file.type.startsWith('image/')) {
      e.preventDefault()
      const coords = view.posAtCoords({ x: e.clientX, y: e.clientY })
      const insertPos = coords ?? view.state.doc.length
      uploadImageFile(file).then((url) => {
        if (url) {
          const text = `![](${url})`
          view.dispatch({
            changes: { from: insertPos, to: insertPos, insert: text },
          })
        }
      })
      return true
    }
  }
  return false
}

// ─── Mount editor ───────────────────────────────────────────────────────────
onMounted(() => {
  if (!editorContainerRef.value) return

  const updateListener = EditorView.updateListener.of((update) => {
    if (update.docChanged) {
      const value = update.state.doc.toString()
      emit('update:modelValue', value)
    }
  })

  const pasteExt = EditorView.domEventHandlers({
    paste(e, view) {
      return handlePaste(e as ClipboardEvent, view)
    },
    drop(e, view) {
      return handleDrop(e as DragEvent, view)
    },
  })

  const state = EditorState.create({
    doc: props.modelValue || '',
    extensions: [
      lineNumbers(),
      history(),
      markdown({ codeLanguages: languages }),
      updateListener,
      pasteExt,
      headingLinePlugin,
      keymap.of([...defaultKeymap, ...historyKeymap, indentWithTab]),
      EditorView.theme({
        '&': { height: '100%' },
        '.cm-scroller': { overflow: 'auto' },
      }),
    ],
  })

  editorView.value = new EditorView({
    state,
    parent: editorContainerRef.value,
  })
})

onBeforeUnmount(() => {
  editorView.value?.destroy()
})

// Sync external model changes into editor (e.g. when loading post for edit)
watch(
  () => props.modelValue,
  (newVal) => {
    const view = editorView.value
    if (!view) return
    const current = view.state.doc.toString()
    if (current !== newVal) {
      view.dispatch({
        changes: { from: 0, to: current.length, insert: newVal },
      })
    }
  },
)

defineExpose({ mobilePreviewOpen })
</script>

<style>
@import '@/assets/editor.css';
</style>

<style scoped>
.md-editor-wrap {
  flex: 1;
  min-height: 0;
}
</style>
