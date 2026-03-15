<template>
  <div class="markdown-editor">
    <!-- Toolbar -->
    <div class="toolbar">
      <button
        v-for="btn in toolbarButtons"
        :key="btn.label"
        type="button"
        @click="btn.action"
        :title="btn.title"
        class="toolbar-btn"
        :class="btn.monospace ? 'mono' : ''"
      >
        {{ btn.label }}
      </button>
    </div>

    <!-- Typora body -->
    <div class="md-body" ref="bodyRef" @mousedown="onBodyMousedown">
      <div class="md-title-wrap">
        <input
          class="md-title-input"
          :value="titleText"
          :placeholder="titlePlaceholder"
          @input="onTitleInput"
          spellcheck="false"
        />
      </div>

      <div class="md-title-notch" aria-hidden="true"></div>

      <div
        v-for="(block, i) in contentBlocks"
        :key="i"
        class="md-block md-content-block"
        :class="{ focused: focusedIndex === i + 1 }"
      >
        <div class="md-line-nos" aria-hidden="true">
          <span v-for="n in blockLineCount(block)" :key="n">
            {{ blockStartLine(i + 1) + n - 1 }}
          </span>
        </div>

        <!-- Raw editing textarea: only shown when focused -->
        <textarea
          v-if="focusedIndex === i + 1"
          :ref="(el) => setBlockRef(el as HTMLTextAreaElement | null, i + 1)"
          class="md-raw"
          :value="block"
          @input="onBlockInput(i + 1, $event)"
          @blur="onBlockBlur(i + 1)"
          @keydown="onBlockKeydown(i + 1, $event)"
          spellcheck="false"
          rows="1"
        />
        <!-- Rendered preview: shown when not focused -->
        <div
          v-else
          class="md-rendered prose-output"
          :class="{ empty: !block.trim() }"
          v-html="renderBlock(block)"
          @mousedown="onRenderedMousedown(i + 1, $event)"
        />
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed, ref, watch, nextTick, onMounted } from 'vue'
import { useMarkdownRenderer } from '@/composables/useMarkdownRenderer'

const props = defineProps<{
  modelValue: string
  placeholder?: string
}>()

const emit = defineEmits<{
  'update:modelValue': [value: string]
}>()

const { renderMarkdown } = useMarkdownRenderer()

function splitBlocks(raw: string): string[] {
  const parts = raw.split(/\n\n+/)
  return parts.length ? parts : ['']
}

function normalizeBlocks(raw: string): string[] {
  const parts = splitBlocks(raw || '')
  if (!parts.length) return ['# ', '']

  const first = parts[0]?.trim() || ''
  if (/^#{1,6}\s+/.test(first)) {
    parts[0] = first
  } else {
    parts[0] = first ? `# ${first}` : '# '
  }

  if (parts.length === 1) {
    parts.push('')
  }

  return parts
}

function titleFromBlock(block: string): string {
  return (block || '').replace(/^#{1,6}\s*/, '').trim()
}

const localBlocks = ref<string[]>(normalizeBlocks(props.modelValue || ''))
const focusedIndex = ref<number | null>(null)
const bodyRef = ref<HTMLDivElement | null>(null)
let blurTimer: ReturnType<typeof setTimeout> | null = null

const titlePlaceholder = props.placeholder || '文章标题...'

const contentBlocks = computed(() => localBlocks.value.slice(1))
const titleText = computed(() => titleFromBlock(localBlocks.value[0] || ''))

// Store click coordinates to restore cursor position after textarea mounts
const pendingClickPos = ref<{ x: number; y: number } | null>(null)

const blockRefs: (HTMLTextAreaElement | null)[] = []
function setBlockRef(el: HTMLTextAreaElement | null, i: number) {
  blockRefs[i] = el
  // If we have a pending click position, use it to place the cursor
  if (el && pendingClickPos.value) {
    const { x, y } = pendingClickPos.value
    pendingClickPos.value = null
    nextTick(() => {
      autoResize(el)
      el.focus()
      const offset = getCaretOffsetFromPoint(el, x, y)
      if (offset !== null) {
        el.selectionStart = el.selectionEnd = offset
      }
    })
  }
}

watch(
  () => props.modelValue,
  (val) => {
    const current = localBlocks.value.join('\n\n')
    if (val !== current) {
      localBlocks.value = normalizeBlocks(val || '')
      focusedIndex.value = null
    }
  }
)

function onTitleInput(e: Event) {
  const raw = (e.target as HTMLInputElement).value || ''
  localBlocks.value[0] = raw.trim() ? `# ${raw.trim()}` : '# '
  syncToModel()
}

function syncToModel() {
  emit('update:modelValue', localBlocks.value.join('\n\n'))
}

function renderBlock(md: string): string {
  if (!md.trim()) return ''
  return renderMarkdown(md)
}

function autoResize(el: HTMLTextAreaElement) {
  el.style.height = 'auto'
  el.style.height = el.scrollHeight + 'px'
}

// Convert a pixel point to a textarea character offset using a hidden mirror div
function getCaretOffsetFromPoint(ta: HTMLTextAreaElement, x: number, y: number): number | null {
  // Use caretPositionFromPoint / caretRangeFromPoint on the textarea itself
  // Textarea text nodes are not in the DOM tree directly, so we use a mirror approach.
  // Simpler: use document.elementFromPoint after the textarea is positioned, then
  // fall back to the native caret APIs which work on rendered text nodes.
  try {
    // Standard (Firefox)
    if ('caretPositionFromPoint' in document) {
      const pos = (document as any).caretPositionFromPoint(x, y)
      if (pos) return Math.min(pos.offset, ta.value.length)
    }
    // WebKit/Blink
    if ('caretRangeFromPoint' in document) {
      const range = (document as any).caretRangeFromPoint(x, y)
      if (range) return Math.min(range.startOffset, ta.value.length)
    }
  } catch (_) {}

  // Fallback: place at end
  return ta.value.length
}

// Called when user clicks on md-body padding area (not on a block)
function onBodyMousedown(e: MouseEvent) {
  // If click target is the body itself (not a child block), focus last block
  if (e.target === bodyRef.value) {
    e.preventDefault()
    focusBlock(localBlocks.value.length - 1)
  }
}

// Called on mousedown so we capture coords before the rendered div is replaced by textarea
function onRenderedMousedown(i: number, e: MouseEvent) {
  if (blurTimer) {
    clearTimeout(blurTimer)
    blurTimer = null
  }
  e.preventDefault()
  pendingClickPos.value = { x: e.clientX, y: e.clientY }
  focusedIndex.value = i

  // Ensure we always re-enter edit mode, even if caret-from-point APIs fail.
  nextTick(() => {
    const el = blockRefs[i]
    if (!el) return
    autoResize(el)
    el.focus()
  })
}

function focusBlock(i: number) {
  if (blurTimer) {
    clearTimeout(blurTimer)
    blurTimer = null
  }
  pendingClickPos.value = null
  focusedIndex.value = i
  nextTick(() => {
    const el = blockRefs[i]
    if (el) {
      autoResize(el)
      el.focus()
      el.selectionStart = el.selectionEnd = el.value.length
    }
  })
}

function onBlockInput(i: number, e: Event) {
  const val = (e.target as HTMLTextAreaElement).value
  localBlocks.value[i] = val
  autoResize(e.target as HTMLTextAreaElement)
  syncToModel()
}

function onBlockBlur(i: number) {
  if (blurTimer) {
    clearTimeout(blurTimer)
  }

  blurTimer = setTimeout(() => {
    const root = bodyRef.value
    const active = document.activeElement as Node | null

    // If focus moved to another element inside editor body, keep editing state.
    if (root && active && root.contains(active)) {
      return
    }

    if (focusedIndex.value === i) {
      focusedIndex.value = null
    }
  }, 0)
}

function getLineCol(ta: HTMLTextAreaElement): { line: number; col: number; totalLines: number } {
  const val = ta.value
  const pos = ta.selectionStart
  const before = val.substring(0, pos)
  const lines = before.split('\n')
  const line = lines.length - 1
  const col = lines[line].length
  const totalLines = val.split('\n').length
  return { line, col, totalLines }
}

function focusBlockAtCol(targetI: number, col: number, fromBottom: boolean) {
  pendingClickPos.value = null
  focusedIndex.value = targetI
  nextTick(() => {
    const el = blockRefs[targetI]
    if (!el) return
    autoResize(el)
    el.focus()
    const lines = el.value.split('\n')
    const targetLine = fromBottom ? lines.length - 1 : 0
    const lineStart = lines.slice(0, targetLine).reduce((acc, l) => acc + l.length + 1, 0)
    const lineLen = lines[targetLine]?.length ?? 0
    el.selectionStart = el.selectionEnd = lineStart + Math.min(col, lineLen)
  })
}

function onBlockKeydown(i: number, e: KeyboardEvent) {
  const ta = e.target as HTMLTextAreaElement

  if (e.key === 'Tab') {
    e.preventDefault()
    const start = ta.selectionStart
    const end = ta.selectionEnd
    const val = ta.value
    const newVal = val.substring(0, start) + '  ' + val.substring(end)
    localBlocks.value[i] = newVal
    syncToModel()
    nextTick(() => {
      ta.selectionStart = ta.selectionEnd = start + 2
      autoResize(ta)
    })
    return
  }

  if (e.key === 'ArrowUp') {
    const { line, col } = getLineCol(ta)
    if (line === 0 && i > 1) {
      e.preventDefault()
      focusBlockAtCol(i - 1, col, true)
    }
    return
  }

  if (e.key === 'ArrowDown') {
    const { line, col, totalLines } = getLineCol(ta)
    if (line === totalLines - 1 && i < localBlocks.value.length - 1) {
      e.preventDefault()
      focusBlockAtCol(i + 1, col, false)
    }
    return
  }

  if (e.key === 'Enter' && !e.shiftKey) {
    const atEnd = ta.selectionStart === ta.value.length
    const inList = /^(\s*[-*+]|\s*\d+\.) /.test(ta.value.split('\n').at(-1) || '')
    const inCode = (ta.value.match(/```/g) || []).length % 2 !== 0
    if (atEnd && !inList && !inCode) {
      e.preventDefault()
      localBlocks.value.splice(i + 1, 0, '')
      syncToModel()
      nextTick(() => focusBlock(i + 1))
    }
    return
  }

  if (e.key === 'Backspace' && ta.value === '' && localBlocks.value.length > 2) {
    e.preventDefault()
    localBlocks.value.splice(i, 1)
    syncToModel()
    nextTick(() => focusBlock(Math.max(0, i - 1)))
    return
  }
}

function getActiveTa(): HTMLTextAreaElement | null {
  if (focusedIndex.value === null) return null
  return blockRefs[focusedIndex.value] ?? null
}

function wrap(before: string, after: string, placeholder: string) {
  const ta = getActiveTa()
  if (!ta) return
  const i = focusedIndex.value!
  const start = ta.selectionStart
  const end = ta.selectionEnd
  const selected = ta.value.substring(start, end) || placeholder
  const newVal = ta.value.substring(0, start) + before + selected + after + ta.value.substring(end)
  localBlocks.value[i] = newVal
  syncToModel()
  nextTick(() => {
    ta.focus()
    ta.selectionStart = start + before.length
    ta.selectionEnd = start + before.length + selected.length
    autoResize(ta)
  })
}

function insertLine(prefix: string) {
  const ta = getActiveTa()
  if (!ta) return
  const i = focusedIndex.value!
  const start = ta.selectionStart
  const val = ta.value
  const lineStart = val.lastIndexOf('\n', start - 1) + 1
  const newVal = val.substring(0, lineStart) + prefix + val.substring(lineStart)
  localBlocks.value[i] = newVal
  syncToModel()
  nextTick(() => {
    ta.focus()
    ta.selectionStart = ta.selectionEnd = start + prefix.length
    autoResize(ta)
  })
}

function ensureFocused() {
  if (focusedIndex.value === null) {
    focusBlock(localBlocks.value.length - 1)
  }
}

const toolbarButtons = [
  { label: 'B', title: '粗体', monospace: false, action: () => { ensureFocused(); nextTick(() => wrap('**', '**', '粗体文字')) } },
  { label: 'I', title: '斜体', monospace: false, action: () => { ensureFocused(); nextTick(() => wrap('*', '*', '斜体文字')) } },
  { label: '~~', title: '删除线', monospace: true, action: () => { ensureFocused(); nextTick(() => wrap('~~', '~~', '删除线')) } },
  { label: 'H1', title: '一级标题', monospace: false, action: () => { ensureFocused(); nextTick(() => insertLine('# ')) } },
  { label: 'H2', title: '二级标题', monospace: false, action: () => { ensureFocused(); nextTick(() => insertLine('## ')) } },
  { label: 'H3', title: '三级标题', monospace: false, action: () => { ensureFocused(); nextTick(() => insertLine('### ')) } },
  { label: '`code`', title: '行内代码', monospace: true, action: () => { ensureFocused(); nextTick(() => wrap('`', '`', 'code')) } },
  { label: '```', title: '代码块', monospace: true, action: () => { ensureFocused(); nextTick(() => wrap('```\n', '\n```', 'code')) } },
  { label: '> 引用', title: '引用', monospace: false, action: () => { ensureFocused(); nextTick(() => insertLine('> ')) } },
  { label: '— 分割线', title: '分割线', monospace: false, action: () => { ensureFocused(); nextTick(() => insertLine('\n---\n')) } },
  { label: '链接', title: '链接', monospace: false, action: () => { ensureFocused(); nextTick(() => wrap('[', '](url)', '链接文字')) } },
]

onMounted(() => {
  if (localBlocks.value.length === 0) localBlocks.value = ['# ', '']
})

function blockLineCount(text: string): number {
  return Math.max(1, text.split('\n').length)
}

function blockStartLine(blockIndex: number): number {
  // Line numbers are only for content blocks, title is excluded.
  let line = 1
  for (let i = 1; i < blockIndex; i++) {
    line += blockLineCount(localBlocks.value[i] || '') + 1
  }
  return line
}
</script>

<style scoped>
.markdown-editor {
  border: 2px solid #000;
  display: flex;
  flex-direction: column;
  height: 100%;
  min-height: 0;
}
.toolbar {
  display: flex;
  flex-wrap: wrap;
  gap: 0.25rem;
  padding: 0.375rem 0.75rem;
  border-bottom: 2px solid #000;
  background: #f9fafb;
  flex-shrink: 0;
}
.toolbar-btn {
  padding: 0.2rem 0.45rem;
  font-size: 0.75rem;
  font-weight: 900;
  border: 1px solid #000;
  background: #fff;
  cursor: pointer;
  transition: all 0.15s;
}
.toolbar-btn:hover { background: #000; color: #fff; }
.toolbar-btn.mono { font-family: ui-monospace, SFMono-Regular, Menlo, monospace; }
.md-body {
  flex: 1;
  overflow-y: auto;
  padding: 1.25rem 2.5rem 1.5rem;
  max-width: 800px;
  margin: 0 auto;
  width: 100%;
  box-sizing: border-box;
}
.md-title-wrap {
  padding: 0.25rem 0 0.4rem;
}
.md-title-input {
  width: 100%;
  border: none;
  outline: none;
  background: transparent;
  font-size: 2.1rem;
  font-weight: 900;
  letter-spacing: -0.04em;
  line-height: 1.1;
  padding: 0;
}
.md-title-input::placeholder {
  color: #d1d5db;
}
.md-title-notch {
  height: 0;
  border-top: 2px solid #000;
  margin: 0.3rem 0 0.8rem;
}
.md-block {
  position: relative;
}
.md-content-block {
  padding-left: 3rem;
}
.md-line-nos {
  position: absolute;
  left: 0;
  top: 0;
  width: 2.4rem;
  text-align: right;
  color: #9ca3af;
  font-size: 0.72rem;
  font-family: ui-monospace, SFMono-Regular, Menlo, monospace;
  line-height: 1.5;
  user-select: none;
  pointer-events: none;
  display: flex;
  flex-direction: column;
  gap: 0;
}
.md-raw {
  width: 100%;
  border: none;
  outline: none;
  resize: none;
  font-family: ui-monospace, SFMono-Regular, Menlo, monospace;
  font-size: 0.9rem;
  line-height: 1.5;
  padding: 1px 0;
  background: #fafaf7;
  overflow: hidden;
  display: block;
  box-sizing: border-box;
}
.md-block.focused .md-raw {
  background: #fffdf0;
}
.md-rendered {
  min-height: 1.4em;
  cursor: text;
  padding: 1px 0;
}
.md-rendered.empty::before {
  content: attr(data-placeholder);
  color: #d1d5db;
  font-style: italic;
}

@media (max-width: 900px) {
  .md-body {
    padding: 1rem 1rem 1.25rem;
  }

  .md-title-input {
    font-size: 1.6rem;
  }

  .md-content-block {
    padding-left: 2.35rem;
  }

  .md-line-nos {
    width: 1.8rem;
    font-size: 0.65rem;
  }
}
</style>

<style>
.prose-output {
  font-size: 0.9375rem;
  line-height: 1.55;
  word-break: break-word;
}

.prose-output h1 { font-size: 1.75rem; font-weight: 900; margin: 1rem 0 0.25rem; line-height: 1.2; }
.prose-output h2 { font-size: 1.375rem; font-weight: 900; margin: 0.875rem 0 0.25rem; line-height: 1.2; }
.prose-output h3 { font-size: 1.125rem; font-weight: 900; margin: 0.75rem 0 0.2rem; line-height: 1.2; }
.prose-output h4, .prose-output h5, .prose-output h6 { font-weight: 700; margin: 0.5rem 0 0.2rem; }
.prose-output p { margin: 0.35rem 0; }
.prose-output code {
  background: #f3f4f6;
  padding: 0.1em 0.35em;
  font-size: 0.85em;
  font-family: ui-monospace, SFMono-Regular, Menlo, monospace;
  border: 1px solid #e5e7eb;
}
.prose-output pre {
  background: #111;
  color: #f8f8f2;
  padding: 1rem;
  overflow-x: auto;
  margin: 0.625rem 0;
}
.prose-output pre code {
  background: none;
  padding: 0;
  color: inherit;
  font-size: 0.875rem;
  border: none;
}
.prose-output blockquote {
  border-left: 4px solid black;
  padding: 0.15rem 0 0.15rem 1rem;
  margin: 0.625rem 0;
  color: #555;
}
.prose-output blockquote > p { margin: 0.15rem 0; }
.prose-output ul { list-style-type: disc; padding-left: 1.5rem; margin: 0.375rem 0; }
.prose-output ol { list-style-type: decimal; padding-left: 1.5rem; margin: 0.375rem 0; }
.prose-output li { margin: 0.1rem 0; }
.prose-output li p { margin: 0; }
.prose-output hr { border: none; border-top: 2px solid black; margin: 1rem 0; }
.prose-output a { text-decoration: underline; color: inherit; }
.prose-output a:hover { opacity: 0.7; }
.prose-output img { max-width: 100%; border: 2px solid black; }
.prose-output strong { font-weight: 900; }
.prose-output em { font-style: italic; }
.prose-output table { border-collapse: collapse; width: 100%; margin: 0.625rem 0; }
.prose-output th, .prose-output td { border: 2px solid black; padding: 0.375rem 0.625rem; text-align: left; }
.prose-output th { font-weight: 900; background: #f3f4f6; }
.prose-output ::selection { background: rgba(0, 0, 0, 0.12); }
</style>
