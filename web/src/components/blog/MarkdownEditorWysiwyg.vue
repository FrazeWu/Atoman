<template>
  <div class="ir-panel" @click.self="focusEnd">
    <div class="render-edit-container">
      <div class="ir-editor" @click.self="focusEnd">
        <template v-for="block in blocks" :key="block.id">
          <textarea
            v-if="focusedBlockId === block.id"
            :ref="(el) => setTextareaRef(block.id, el)"
            :value="block.raw"
            class="ir-source-block"
            :class="`ir-source-block--${block.type}`"
            rows="1"
            @input="onBlockInput(block.id, $event)"
            @blur="onBlockBlur"
            @keydown="onBlockKeydown(block.id, $event)"
          />
          <button
            v-else
            type="button"
            class="ir-render-block"
            :class="`ir-render-block--${block.type}`"
            @click="focusBlock(block.id)"
            v-html="block.html || '<p><br></p>'"
          />
        </template>
      </div>
      <p v-if="!modelValue.trim() && !blocks.length" class="ir-placeholder" @click="focusEnd">开始输入内容...</p>
    </div>
  </div>
</template>

<script setup lang="ts">
import { nextTick, ref, watch } from 'vue'
import { parseBlocks, renderToken } from '@/composables/useMarkdownRenderer'

interface Props {
  modelValue?: string
  placeholder?: string
}

interface EditorBlock {
  id: string
  raw: string
  type: string
  html: string
}

const props = withDefaults(defineProps<Props>(), {
  modelValue: '',
  placeholder: '',
})

const emit = defineEmits<{
  'update:modelValue': [value: string]
}>()

const focusedBlockId = ref<string | null>(null)
const textareaRefs = new Map<string, HTMLTextAreaElement>()
const blockIdCounter = ref(0)
const internalMarkdown = ref(props.modelValue)
const blocks = ref<EditorBlock[]>(createBlocksFromMarkdown(props.modelValue))

function nextBlockId() {
  blockIdCounter.value += 1
  return `block-${blockIdCounter.value}`
}

function normalizeBlockRaw(raw: string) {
  return raw.endsWith('\n') ? raw : `${raw}\n`
}

function createBlocksFromMarkdown(markdown: string) {
  const tokens = parseBlocks(markdown)
  const nextBlocks: EditorBlock[] = []

  for (const token of tokens) {
    if (token.type === 'space') continue
    const raw = typeof (token as { raw?: string }).raw === 'string' ? (token as { raw?: string }).raw || '' : ''
    nextBlocks.push({
      id: nextBlockId(),
      raw,
      type: token.type,
      html: renderToken(token),
    })
  }

  if (nextBlocks.length === 0 && markdown.trim()) {
    nextBlocks.push({
      id: nextBlockId(),
      raw: normalizeBlockRaw(markdown),
      type: 'paragraph',
      html: renderParagraphFallback(markdown),
    })
  }

  return nextBlocks
}

function renderParagraphFallback(text: string) {
  const escaped = text
    .replace(/&/g, '&amp;')
    .replace(/</g, '&lt;')
    .replace(/>/g, '&gt;')
    .replace(/\n/g, '<br>')
  return `<p>${escaped}</p>`
}

function rebuildMarkdown() {
  internalMarkdown.value = blocks.value.map((block) => block.raw).join('')
  emit('update:modelValue', internalMarkdown.value)
}

function syncBlocksFromMarkdown(markdown: string) {
  internalMarkdown.value = markdown
  blocks.value = createBlocksFromMarkdown(markdown)
}

function focusBlock(blockId: string) {
  focusedBlockId.value = blockId
  nextTick(() => {
    const textarea = textareaRefs.get(blockId)
    if (!textarea) return
    textarea.focus()
    textarea.selectionStart = textarea.value.length
    textarea.selectionEnd = textarea.value.length
    autoResize(textarea)
  })
}

function focusEnd() {
  if (blocks.value.length === 0) {
    blocks.value = [
      {
        id: nextBlockId(),
        raw: '',
        type: 'paragraph',
        html: '',
      },
    ]
  }
  focusBlock(blocks.value[blocks.value.length - 1].id)
}

function setTextareaRef(blockId: string, el: unknown) {
  if (el instanceof HTMLTextAreaElement) {
    textareaRefs.set(blockId, el)
    autoResize(el)
  } else {
    textareaRefs.delete(blockId)
  }
}

function autoResize(textarea: HTMLTextAreaElement) {
  textarea.style.height = 'auto'
  textarea.style.height = `${Math.max(textarea.scrollHeight, 36)}px`
}

function onBlockInput(blockId: string, event: Event) {
  const target = event.target as HTMLTextAreaElement
  const value = target.value
  autoResize(target)
  const index = blocks.value.findIndex((block) => block.id === blockId)
  if (index === -1) return
  blocks.value[index] = {
    ...blocks.value[index],
    raw: value,
    html: renderParagraphFallback(value),
  }
}

function onBlockBlur() {
  commitFocusedBlock()
}

function commitFocusedBlock() {
  if (!focusedBlockId.value) return
  const currentId = focusedBlockId.value
  const index = blocks.value.findIndex((block) => block.id === currentId)
  focusedBlockId.value = null
  if (index === -1) return

  const mergedMarkdown = blocks.value.map((block) => block.raw).join('')
  syncBlocksFromMarkdown(mergedMarkdown)
  rebuildMarkdown()
}

function onBlockKeydown(blockId: string, event: KeyboardEvent) {
  const textarea = event.target as HTMLTextAreaElement
  const index = blocks.value.findIndex((block) => block.id === blockId)
  if (index === -1) return

  if (event.key === 'Escape') {
    event.preventDefault()
    commitFocusedBlock()
    return
  }

  if (event.key === 'Enter' && !event.shiftKey) {
    event.preventDefault()
    const { selectionStart, selectionEnd, value } = textarea
    const before = value.slice(0, selectionStart)
    const after = value.slice(selectionEnd)
    const currentRaw = before
    const nextRaw = after

    blocks.value[index] = {
      ...blocks.value[index],
      raw: currentRaw,
      html: renderParagraphFallback(currentRaw),
    }

    const newBlock: EditorBlock = {
      id: nextBlockId(),
      raw: nextRaw,
      type: 'paragraph',
      html: renderParagraphFallback(nextRaw),
    }

    blocks.value.splice(index + 1, 0, newBlock)
    rebuildMarkdown()
    focusBlock(newBlock.id)
    return
  }

  if (event.key === 'Backspace' && textarea.selectionStart === 0 && textarea.selectionEnd === 0 && index > 0) {
    event.preventDefault()
    const prev = blocks.value[index - 1]
    const current = blocks.value[index]
    const merged = `${prev.raw}${current.raw}`
    blocks.value[index - 1] = {
      ...prev,
      raw: merged,
      html: renderParagraphFallback(merged),
    }
    blocks.value.splice(index, 1)
    rebuildMarkdown()
    focusBlock(prev.id)
  }
}

function getMarkdown() {
  if (focusedBlockId.value) {
    commitFocusedBlock()
  }
  return internalMarkdown.value
}

function setContent(markdown: string) {
  focusedBlockId.value = null
  syncBlocksFromMarkdown(markdown)
}

function insertMarkdownAtCursor(before: string, after: string, placeholder: string) {
  const blockId = focusedBlockId.value
  if (!blockId) {
    focusEnd()
    nextTick(() => insertMarkdownAtCursor(before, after, placeholder))
    return
  }

  const textarea = textareaRefs.get(blockId)
  if (!textarea) return
  const selected = textarea.value.slice(textarea.selectionStart, textarea.selectionEnd) || placeholder
  const nextValue =
    textarea.value.slice(0, textarea.selectionStart) +
    before +
    selected +
    after +
    textarea.value.slice(textarea.selectionEnd)

  textarea.value = nextValue
  const start = textarea.selectionStart + before.length
  const end = start + selected.length
  textarea.selectionStart = start
  textarea.selectionEnd = end
  onBlockInput(blockId, { target: textarea } as unknown as Event)
  rebuildMarkdown()
}

function insertLinePrefixAtCursor(prefix: string) {
  const blockId = focusedBlockId.value
  if (!blockId) {
    focusEnd()
    nextTick(() => insertLinePrefixAtCursor(prefix))
    return
  }

  const textarea = textareaRefs.get(blockId)
  if (!textarea) return
  const start = textarea.selectionStart
  const lineStart = textarea.value.lastIndexOf('\n', start - 1) + 1
  const nextValue = textarea.value.slice(0, lineStart) + prefix + textarea.value.slice(lineStart)
  textarea.value = nextValue
  textarea.selectionStart = start + prefix.length
  textarea.selectionEnd = start + prefix.length
  onBlockInput(blockId, { target: textarea } as unknown as Event)
  rebuildMarkdown()
}

function replaceInMarkdown(search: string, replacement: string) {
  const nextMarkdown = getMarkdown().replace(search, replacement)
  setContent(nextMarkdown)
  rebuildMarkdown()
}

function mount() {}
function unmount() {}

watch(
  () => props.modelValue,
  (newVal) => {
    if (newVal === internalMarkdown.value) return
    setContent(newVal)
  },
)

defineExpose({
  focusEnd,
  getMarkdown,
  setContent,
  insertMarkdownAtCursor,
  insertLinePrefixAtCursor,
  replaceInMarkdown,
  mount,
  unmount,
})
</script>

<style scoped>
.ir-panel {
  flex: 1;
  overflow-y: auto;
  background: #fff;
  cursor: text;
}

.render-edit-container {
  min-height: 100%;
  padding: 2rem 3rem;
  max-width: 860px;
  margin: 0 auto;
  position: relative;
}

.ir-editor {
  min-height: 100%;
}

.ir-render-block {
  display: block;
  width: 100%;
  border: none;
  background: transparent;
  text-align: left;
  padding: 0.35rem 0.5rem;
  margin: 0;
  cursor: text;
}

.ir-render-block:hover {
  background: #fafafa;
}

.ir-source-block {
  display: block;
  width: 100%;
  resize: none;
  overflow: hidden;
  border: none;
  outline: none;
  background: #f9fafb;
  border-left: 3px solid #000;
  padding: 0.35rem 0.75rem;
  font-family: 'JetBrains Mono', 'Fira Code', monospace;
  font-size: 0.95rem;
  line-height: 1.75;
  color: #111827;
}

.ir-placeholder {
  position: absolute;
  top: 2rem;
  left: 3rem;
  color: #9ca3af;
  pointer-events: none;
  font-size: 1rem;
  margin: 0;
}

@media (max-width: 768px) {
  .render-edit-container {
    padding: 1rem;
  }

  .ir-placeholder {
    left: 1rem;
    top: 1rem;
  }
}
</style>

<style>
.ir-render-block h1,
.ir-render-block h2,
.ir-render-block h3,
.ir-render-block h4,
.ir-render-block h5,
.ir-render-block h6 {
  font-weight: 900;
  letter-spacing: -0.03em;
  margin: 1.25rem 0 0.5rem;
  line-height: 1.2;
}

.ir-render-block h1 { font-size: 2rem; }
.ir-render-block h2 { font-size: 1.5rem; }
.ir-render-block h3 { font-size: 1.25rem; }
.ir-render-block h4 { font-size: 1.1rem; }

.ir-render-block p {
  margin: 0.5rem 0;
  line-height: 1.75;
}

.ir-render-block code {
  font-family: 'JetBrains Mono', 'Fira Code', monospace;
  font-size: 0.875em;
  background: #f3f4f6;
  border: 1px solid #e5e7eb;
  padding: 0.1em 0.3em;
}

.ir-render-block pre {
  background: #1f2937;
  color: #f9fafb;
  padding: 1rem;
  overflow-x: auto;
  border: 2px solid #000;
  margin: 0.75rem 0;
}

.ir-render-block pre code {
  background: none;
  border: none;
  padding: 0;
  color: inherit;
  font-size: 0.875rem;
}

.ir-render-block blockquote {
  border-left: 4px solid #000;
  margin: 0.75rem 0;
  padding-left: 1rem;
  color: #374151;
}

.ir-render-block ul,
.ir-render-block ol {
  padding-left: 1.5rem;
  margin: 0.5rem 0;
}

.ir-render-block li {
  margin: 0.25rem 0;
}

.ir-render-block a {
  color: #000;
  text-decoration: underline;
  text-underline-offset: 2px;
}

.ir-render-block a:hover {
  background: #000;
  color: #fff;
}

.ir-render-block img {
  max-width: 100%;
  filter: grayscale(100%);
  border: 2px solid #000;
  display: block;
  margin: 1rem 0;
}

.ir-render-block hr {
  border: none;
  border-top: 2px solid #000;
  margin: 1.5rem 0;
}

.ir-render-block strong {
  font-weight: 900;
}

.ir-render-block del {
  text-decoration: line-through;
  text-decoration-thickness: 2px;
}

.ir-render-block table {
  width: 100%;
  border-collapse: collapse;
  margin: 1rem 0;
  table-layout: fixed;
}

.ir-render-block th,
.ir-render-block td {
  border: 2px solid #000;
  padding: 0.65rem 0.75rem;
  text-align: left;
  vertical-align: top;
}

.ir-render-block th {
  font-weight: 900;
  background: #f3f4f6;
}

.ir-render-block tbody tr:nth-child(even) {
  background: #fafafa;
}
</style>
