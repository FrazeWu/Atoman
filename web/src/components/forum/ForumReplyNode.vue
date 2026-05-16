<template>
  <div class="reply-node" :class="{ 'reply-node-solution': isSolution }">
    <!-- Solution badge -->
    <div v-if="isSolution" class="reply-solution-badge a-badge">✓ 采纳答案</div>

    <!-- Reply header -->
    <div class="reply-header">
      <div class="reply-meta">
        <span class="reply-author">{{ displayName }}</span>
        <span v-if="floorLabel" class="reply-floor a-badge">{{ floorLabel }}</span>
        <span class="reply-time">{{ formatTime(reply.created_at) }}</span>
      </div>
      <div class="reply-actions">
        <ABtn
          v-if="isAuthenticated"
          outline
          size="sm"
          class="reply-btn"
          @click="$emit('quote', reply)"
        >引用</ABtn>
        <ABtn
          v-if="isAuthenticated"
          outline
          size="sm"
          class="reply-btn"
          :class="{ 'reply-btn-active': reply.is_liked }"
          @click="$emit('toggle-like', reply.id)"
        >
          赞 {{ reply.like_count }}
        </ABtn>
        <ABtn
          v-if="isOwn"
          danger
          size="sm"
          class="reply-btn"
          @click="$emit('delete', reply.id)"
        >删除</ABtn>
        <ABtn
          v-if="isAuthenticated && !isOwn"
          outline
          size="sm"
          class="reply-btn"
          @click="$emit('report', reply.id)"
        >举报</ABtn>
        <ABtn
          v-if="isTopicOwner && isSolution"
          outline
          size="sm"
          class="reply-btn"
          @click="$emit('unsolve', reply.id)"
        >取消采纳</ABtn>
        <ABtn
          v-if="isTopicOwner && !topicIsSolved && !isSolution"
          outline
          size="sm"
          class="reply-btn"
          @click="$emit('solve', reply.id)"
        >采纳</ABtn>
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
import ABtn from '@/components/ui/ABtn.vue'

const props = defineProps<{
  reply: ForumReply
  quotedReply?: ForumReply | null
  authUserId?: string
  isAuthenticated?: boolean
  isSolution?: boolean
  isTopicOwner?: boolean
  topicIsSolved?: boolean
}>()

defineEmits<{
  (e: 'quote', reply: ForumReply): void
  (e: 'delete', id: string): void
  (e: 'toggle-like', id: string): void
  (e: 'report', id: string): void
  (e: 'solve', id: string): void
  (e: 'unsolve', id: string): void
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
  border: var(--a-border);
  background: var(--a-color-bg);
  padding: 1.25rem 1.5rem;
  margin-bottom: 0.75rem;
}

.reply-node-solution {
  border-color: var(--a-color-success);
  border-left-width: 4px;
}

.reply-solution-badge {
  background: var(--a-color-success);
  color: var(--a-color-bg);
  border-color: var(--a-color-success);
  margin-bottom: 0.6rem;
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
  font-weight: var(--a-font-weight-black);
  font-size: 0.85rem;
}

.reply-floor {
  padding: 0.1rem 0.35rem;
}

.reply-time {
  font-size: 0.7rem;
  font-weight: var(--a-font-weight-normal);
  color: var(--a-color-muted-soft);
}

.reply-actions {
  display: flex;
  flex-wrap: wrap;
  gap: 0.4rem;
}

.reply-btn {
  box-shadow: none;
}

.reply-btn-active {
  background: var(--a-color-fg);
  color: var(--a-color-bg);
  border-color: var(--a-color-fg);
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
  background: var(--a-color-disabled-bg);
  border-left: 3px solid var(--a-color-fg);
}

.reply-quote-meta {
  display: flex;
  align-items: center;
  gap: 0.5rem;
  margin-bottom: 0.35rem;
  font-size: 0.7rem;
  font-weight: var(--a-font-weight-black);
  text-transform: uppercase;
  letter-spacing: var(--a-letter-spacing-wide);
}

.reply-quote-text {
  margin: 0;
  font-size: 0.8rem;
  line-height: 1.6;
  color: var(--a-color-muted);
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
  background: var(--a-color-disabled-bg);
  border: var(--a-border);
  padding: 0.875rem;
  overflow-x: auto;
  margin: 0.75rem 0;
  font-size: 0.85em;
}

.reply-body :deep(code) {
  font-family: monospace;
  font-size: 0.875em;
  background: var(--a-color-disabled-bg);
  padding: 0.1em 0.3em;
}

.reply-body :deep(pre code) {
  background: transparent;
  padding: 0;
}

.reply-body :deep(blockquote) {
  border-left: 3px solid var(--a-color-fg);
  padding-left: 1rem;
  margin: 0.75rem 0;
  color: var(--a-color-muted);
}

.reply-body :deep(a) {
  color: var(--a-color-fg);
  text-decoration: underline;
}

.reply-body :deep(ul),
.reply-body :deep(ol) {
  padding-left: 1.5rem;
  margin: 0.5em 0;
}
</style>
