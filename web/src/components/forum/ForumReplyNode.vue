<template>
  <div class="reply-node">
    <!-- Reply header -->
    <div class="reply-header">
      <div class="reply-meta">
        <span class="reply-author">{{ displayName }}</span>
        <span class="reply-floor">{{ floorLabel }}</span>
        <span class="reply-time">{{ formatTime(reply.created_at) }}</span>
      </div>
      <div class="reply-actions">
        <button
          v-if="isAuthenticated"
          class="reply-btn"
          @click="$emit('quote', reply)"
        >引用</button>
        <button
          v-if="isAuthenticated"
          class="reply-btn"
          :class="{ 'reply-btn-liked': reply.is_liked }"
          @click="$emit('toggle-like', reply.id)"
        >
          赞 {{ reply.like_count }}
        </button>
        <button
          v-if="isOwn"
          class="reply-btn reply-btn-danger"
          @click="$emit('delete', reply.id)"
        >删除</button>
      </div>
    </div>

    <!-- Reply body: Markdown rendered -->
    <div v-if="quotedReply" class="reply-quote">
      <div class="reply-quote-meta">
        <span>引用 {{ quotedDisplayName }}</span>
        <span v-if="quotedReply.floor_number > 0">#{{ quotedReply.floor_number }}</span>
      </div>
      <p class="reply-quote-text">{{ quotePreview }}</p>
    </div>

    <div
      class="reply-body markdown-body"
      v-html="renderMarkdown(reply.content)"
    />
  </div>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import type { ForumReply } from '@/types'
import { useMarkdownRenderer } from '@/composables/useMarkdownRenderer'

const props = defineProps<{
  reply: ForumReply
  quotedReply?: ForumReply | null
  authUserId?: string
  isAuthenticated?: boolean
}>()

defineEmits<{
  (e: 'quote', reply: ForumReply): void
  (e: 'delete', id: string): void
  (e: 'toggle-like', id: string): void
}>()

const { renderMarkdown } = useMarkdownRenderer()

const displayName = computed(
  () => props.reply.user?.display_name || props.reply.user?.username || '匿名',
)

const quotedDisplayName = computed(() => {
  if (!props.quotedReply) return ''
  return props.quotedReply.user?.display_name || props.quotedReply.user?.username || '匿名'
})

const quotePreview = computed(() => {
  if (!props.quotedReply) return ''
  return props.quotedReply.content.replace(/\s+/g, ' ').trim().slice(0, 140)
})

const floorLabel = computed(() =>
  props.reply.floor_number > 0 ? `#${props.reply.floor_number}` : '',
)

const isOwn = computed(
  () => props.authUserId != null && props.reply.user_id === props.authUserId,
)

const formatTime = (iso: string) => {
  const d = new Date(iso)
  const diff = Date.now() - d.getTime()
  const mins = Math.floor(diff / 60000)
  if (mins < 60) return `${mins} 分钟前`
  const hours = Math.floor(mins / 60)
  if (hours < 24) return `${hours} 小时前`
  const days = Math.floor(hours / 24)
  if (days < 30) return `${days} 天前`
  return d.toLocaleDateString('zh-CN')
}
</script>

<style scoped>
.reply-node {
  border: 2px solid #000;
  background: #fff;
  padding: 1.25rem 1.5rem;
  margin-bottom: 0.75rem;
}

.reply-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  margin-bottom: 0.75rem;
  flex-wrap: wrap;
  gap: 0.5rem;
}

.reply-meta {
  display: flex;
  align-items: center;
  gap: 0.75rem;
}

.reply-author {
  font-weight: 900;
  font-size: 0.85rem;
}

.reply-floor {
  font-size: 0.7rem;
  font-weight: 900;
  text-transform: uppercase;
  letter-spacing: 0.08em;
  padding: 0.1rem 0.35rem;
  border: 1.5px solid #000;
  color: #000;
  background: transparent;
}

.reply-time {
  font-size: 0.7rem;
  font-weight: 500;
  color: #9ca3af;
}

.reply-actions {
  display: flex;
  gap: 0.4rem;
}

.reply-btn {
  font-size: 0.65rem;
  font-weight: 900;
  text-transform: uppercase;
  letter-spacing: 0.08em;
  padding: 0.25rem 0.5rem;
  border: 1.5px solid #000;
  cursor: pointer;
  transition: all 0.15s;
  background: #fff;
  color: #000;
}

.reply-btn:hover {
  background: #000;
  color: #fff;
}

.reply-btn-liked {
  background: #000;
  color: #fff;
}

.reply-btn-liked:hover {
  background: #fff;
  color: #000;
}

.reply-btn-danger {
  border-color: #ef4444;
  color: #ef4444;
}

.reply-btn-danger:hover {
  background: #ef4444;
  color: #fff;
}

.reply-body {
  font-size: 0.9rem;
  font-weight: 500;
  line-height: 1.7;
  word-break: break-word;
}

.reply-quote {
  margin-bottom: 0.9rem;
  padding: 0.85rem 1rem;
  background: #f3f4f6;
  border-left: 3px solid #000;
}

.reply-quote-meta {
  display: flex;
  align-items: center;
  gap: 0.5rem;
  margin-bottom: 0.35rem;
  font-size: 0.7rem;
  font-weight: 900;
  text-transform: uppercase;
  letter-spacing: 0.08em;
}

.reply-quote-text {
  margin: 0;
  font-size: 0.8rem;
  line-height: 1.6;
  color: #4b5563;
  word-break: break-word;
}

/* Markdown styles within reply body */
.reply-body :deep(h1),
.reply-body :deep(h2),
.reply-body :deep(h3) {
  font-weight: 900;
  letter-spacing: -0.03em;
  margin: 1.2em 0 0.6em;
}

.reply-body :deep(p) {
  margin: 0.5em 0;
}

.reply-body :deep(pre) {
  background: #f3f4f6;
  border: 2px solid #000;
  padding: 0.875rem;
  overflow-x: auto;
  margin: 0.75rem 0;
  font-size: 0.85em;
}

.reply-body :deep(code) {
  font-family: monospace;
  font-size: 0.875em;
  background: #f3f4f6;
  padding: 0.1em 0.3em;
}

.reply-body :deep(pre code) {
  background: transparent;
  padding: 0;
}

.reply-body :deep(blockquote) {
  border-left: 3px solid #000;
  padding-left: 1rem;
  margin: 0.75rem 0;
  color: #6b7280;
}

.reply-body :deep(a) {
  color: #000;
  text-decoration: underline;
}

.reply-body :deep(ul),
.reply-body :deep(ol) {
  padding-left: 1.5rem;
  margin: 0.5em 0;
}
</style>
