<template>
  <nav class="editor-toc">
    <div class="toc-header">目录</div>
    <div v-if="headings.length === 0" class="toc-empty">暂无标题</div>
    <ul v-else class="toc-list">
      <li
        v-for="h in headings"
        :key="h.id"
        class="toc-item"
        :style="{ paddingLeft: `${(h.level - 1) * 12}px` }"
        @click="scrollTo(h.id)"
      >
        <span class="toc-bullet" :class="`toc-level-${h.level}`"></span>
        <span class="toc-text">{{ h.text }}</span>
      </li>
    </ul>
  </nav>
</template>

<script setup lang="ts">
import { computed } from 'vue'

const props = defineProps<{
  content: string
  previewEl?: HTMLElement | null
}>()

interface TocHeading {
  level: number
  text: string
  id: string
}

const headings = computed<TocHeading[]>(() => {
  const result: TocHeading[] = []
  const regex = /^(#{1,4})\s+(.+)$/gm
  let match
  while ((match = regex.exec(props.content)) !== null) {
    const level = match[1].length
    const text = match[2].trim()
    const id = text
      .toLowerCase()
      .replace(/[^\w\u4e00-\u9fa5]+/g, '-')
      .replace(/^-|-$/g, '')
    result.push({ level, text, id })
  }
  return result
})

function scrollTo(id: string) {
  if (props.previewEl) {
    const el = props.previewEl.querySelector(`#${CSS.escape(id)}`)
    if (el) {
      el.scrollIntoView({ behavior: 'smooth', block: 'start' })
      return
    }
  }
  // Fallback: scroll entire page
  const el = document.getElementById(id)
  if (el) el.scrollIntoView({ behavior: 'smooth', block: 'start' })
}
</script>

<style scoped>
.editor-toc {
  padding: 0.75rem 1rem;
  border-bottom: 2px solid #000;
  background: #f9fafb;
  flex-shrink: 0;
  max-height: 280px;
  overflow-y: auto;
}

.toc-header {
  font-size: 0.65rem;
  font-weight: 900;
  text-transform: uppercase;
  letter-spacing: 0.12em;
  color: #6b7280;
  margin-bottom: 0.5rem;
}

.toc-empty {
  font-size: 0.75rem;
  color: #9ca3af;
  font-style: italic;
}

.toc-list {
  list-style: none;
  margin: 0;
  padding: 0;
  display: flex;
  flex-direction: column;
  gap: 2px;
}

.toc-item {
  display: flex;
  align-items: center;
  gap: 0.4rem;
  cursor: pointer;
  padding-top: 2px;
  padding-bottom: 2px;
  border-radius: 2px;
  transition: background 0.1s;
}

.toc-item:hover {
  background: #e5e7eb;
}

.toc-bullet {
  flex-shrink: 0;
  border-radius: 9999px;
  background: #000;
}

.toc-level-1 { width: 6px; height: 6px; }
.toc-level-2 { width: 5px; height: 5px; background: #374151; }
.toc-level-3 { width: 4px; height: 4px; background: #6b7280; }
.toc-level-4 { width: 3px; height: 3px; background: #9ca3af; }

.toc-text {
  font-size: 0.78rem;
  font-weight: 600;
  line-height: 1.35;
  color: #111827;
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
  max-width: 180px;
}
</style>
