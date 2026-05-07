<template>
  <div class="a-page-xl" style="padding-bottom:8rem">
    <div v-if="forumStore.loading" style="padding:4rem 0;text-align:center;font-weight:900;letter-spacing:.1em;font-size:.75rem;text-transform:uppercase">
      加载中...
    </div>

    <template v-else-if="forumStore.currentTopic">
      <!-- Breadcrumb -->
      <div style="display:flex;align-items:center;gap:.75rem;margin-bottom:1.5rem;flex-wrap:wrap">
        <RouterLink
          to="/forum"
          style="font-weight:900;font-size:.75rem;text-transform:uppercase;letter-spacing:.1em;text-decoration:none;color:#6b7280;border-bottom:1px solid transparent;transition:border-color .2s"
          @mouseenter="($event.currentTarget as HTMLElement).style.borderBottomColor='#6b7280'"
          @mouseleave="($event.currentTarget as HTMLElement).style.borderBottomColor='transparent'"
        >论坛</RouterLink>
        <span style="color:#d1d5db">/</span>
        <span
          v-if="forumStore.currentTopic.category"
          style="font-size:.65rem;font-weight:900;text-transform:uppercase;letter-spacing:.08em;padding:.15rem .5rem;border:1.5px solid;cursor:pointer"
          :style="{ borderColor: forumStore.currentTopic.category.color, color: forumStore.currentTopic.category.color }"
          @click="router.push(`/forum?category=${forumStore.currentTopic.category_id}`)"
        >{{ forumStore.currentTopic.category.name }}</span>
      </div>

      <!-- Topic header -->
      <div class="topic-header">
        <h1 class="topic-title">
          <span v-if="forumStore.currentTopic.pinned" class="badge-pinned">置顶</span>
          <span v-if="forumStore.currentTopic.closed" class="badge-closed">已关闭</span>
          {{ forumStore.currentTopic.title }}
        </h1>

        <!-- Tags -->
        <div v-if="(forumStore.currentTopic.tags || []).length > 0" style="display:flex;flex-wrap:wrap;gap:.4rem;margin-bottom:.75rem">
          <span
            v-for="tag in forumStore.currentTopic.tags"
            :key="tag"
            style="font-size:.65rem;font-weight:700;padding:.1rem .5rem;border:1.5px solid #d1d5db;color:#6b7280"
          >{{ tag }}</span>
        </div>

        <!-- Meta bar -->
        <div class="topic-meta">
          <span>{{ forumStore.currentTopic.user?.display_name || forumStore.currentTopic.user?.username || '匿名' }}</span>
          <span>{{ formatTime(forumStore.currentTopic.created_at) }}</span>
          <span>{{ forumStore.currentTopic.view_count }} 浏览</span>
          <span>{{ forumStore.currentTopic.reply_count }} 回复</span>

          <!-- Like button -->
          <button
            v-if="authStore.isAuthenticated"
            @click="forumStore.toggleTopicLike(forumStore.currentTopic!.id)"
            class="action-btn"
            :class="{ 'action-btn-active': forumStore.currentTopic.is_liked }"
          >{{ forumStore.currentTopic.is_liked ? '已赞' : '点赞' }} {{ forumStore.currentTopic.like_count }}</button>
          <span v-else>{{ forumStore.currentTopic.like_count }} 赞</span>

          <!-- Bookmark button -->
          <button
            v-if="authStore.isAuthenticated"
            @click="forumStore.toggleTopicBookmark(forumStore.currentTopic!.id)"
            class="action-btn"
            :class="{ 'action-btn-active': forumStore.currentTopic.is_bookmarked }"
          >{{ forumStore.currentTopic.is_bookmarked ? '已收藏' : '收藏' }}</button>
        </div>
      </div>

      <!-- Layout: content + optional ToC -->
      <div class="topic-layout">
        <div class="topic-main">
          <!-- Topic content (Markdown rendered) -->
          <div
            class="markdown-body"
            style="border:2px solid #000;padding:2rem;margin-bottom:2.5rem;background:#fff"
            v-html="renderMarkdown(forumStore.currentTopic.content)"
            ref="contentRef"
          />

          <!-- Reply sort -->
          <div style="display:flex;align-items:center;justify-content:space-between;margin-bottom:1.25rem">
            <h2 style="font-weight:900;font-size:.75rem;text-transform:uppercase;letter-spacing:.15em;margin:0;padding-bottom:.75rem;border-bottom:2px solid #000;flex:1">
              {{ forumStore.currentTopic.reply_count }} 条回复
            </h2>
            <div style="display:flex;gap:0;border:2px solid #000;margin-left:1rem">
              <button
                v-for="(label, s) in replySortOptions"
                :key="s"
                class="a-tab-btn"
                :class="{ 'a-tab-btn-active': replySort === s }"
                style="padding:.35rem .875rem;font-size:.7rem;border-right:2px solid #000"
                :style="s === 'best' ? 'border-right:none' : ''"
                @click="setReplySort(s as 'oldest' | 'best')"
              >{{ label }}</button>
            </div>
          </div>

          <!-- Replies -->
          <AEmpty v-if="forumStore.replies.length === 0" text="还没有回复，来说第一句" />
          <div style="display:flex;flex-direction:column;gap:.75rem">
            <ForumReplyNode
              v-for="reply in forumStore.replies"
              :key="reply.id"
              :reply="reply"
              :quoted-reply="getQuotedReply(reply)"
              :auth-user-id="authStore.user?.uuid"
              :is-authenticated="authStore.isAuthenticated"
              @quote="setQuote"
              @delete="handleDeleteReply"
              @toggle-like="forumStore.toggleReplyLike"
            />
          </div>

          <!-- Reply form -->
          <div id="reply-form" style="margin-top:2.5rem">
            <div v-if="forumStore.currentTopic.closed" class="reply-closed-notice">
              该话题已关闭，不允许回复
            </div>

            <div v-else-if="!authStore.isAuthenticated" class="reply-login-notice">
              <p style="font-weight:700;font-size:.9rem;margin:0 0 1rem">登录后即可参与讨论</p>
              <ABtn to="/login">登录</ABtn>
            </div>

            <div v-else class="reply-form-wrap">
              <h3 style="font-weight:900;font-size:.7rem;text-transform:uppercase;letter-spacing:.15em;margin:0 0 1rem">
                {{ quotedReply ? `引用 @${getReplyAuthor(quotedReply)}` : '发表回复' }}
              </h3>

              <!-- Quote indicator -->
              <div
                v-if="quotedReply"
                style="display:flex;align-items:flex-start;justify-content:space-between;gap:1rem;padding:.8rem 1rem;background:#f3f4f6;border-left:3px solid #000;margin-bottom:1rem;font-size:.8rem;font-weight:700"
              >
                <div>
                  <div>引用 #{{ quotedReply.floor_number }} @{{ getReplyAuthor(quotedReply) }} 的回复</div>
                  <div style="margin-top:.35rem;font-size:.75rem;font-weight:500;color:#6b7280;line-height:1.6">
                    {{ getReplyPreview(quotedReply.content) }}
                  </div>
                </div>
                <button
                  @click="clearQuote"
                  style="font-size:.7rem;font-weight:900;text-transform:uppercase;letter-spacing:.08em;background:none;border:none;cursor:pointer"
                >取消引用</button>
              </div>

              <!-- Draft restore notice -->
              <div
                v-if="draftRestored"
                style="display:flex;align-items:center;justify-content:space-between;padding:.6rem 1rem;background:#fef3c7;border:1.5px solid #f59e0b;margin-bottom:1rem;font-size:.8rem;font-weight:700"
              >
                <span>已恢复草稿</span>
                <button
                  @click="clearReplyDraft"
                  style="background:none;border:none;font-size:.7rem;font-weight:900;text-transform:uppercase;letter-spacing:.08em;cursor:pointer;color:#92400e"
                >清除</button>
              </div>

              <!-- MarkdownEditor for reply -->
              <div class="reply-editor-wrap">
                <MarkdownEditor
                  v-model="replyContent"
                  :hide-title="true"
                  placeholder="写下你的回复...（支持 Markdown，@用户名 可以 @人）"
                />
              </div>

              <div style="display:flex;justify-content:flex-end;margin-top:.75rem;gap:.75rem">
                <button
                  v-if="replyContent"
                  type="button"
                  style="font-size:.7rem;font-weight:900;text-transform:uppercase;letter-spacing:.08em;background:none;border:none;cursor:pointer;color:#9ca3af"
                  @click="clearReplyDraft"
                >清除草稿</button>
                <ABtn @click="submitReply" :loading="submitting" :disabled="!replyContent.trim()">提交回复</ABtn>
              </div>
            </div>
          </div>
        </div>

        <!-- Table of Contents sidebar (desktop only) -->
        <aside v-if="tocItems.length > 0" class="topic-toc">
          <div style="font-size:.65rem;font-weight:900;text-transform:uppercase;letter-spacing:.12em;margin-bottom:.75rem;color:#9ca3af">目录</div>
          <nav>
            <a
              v-for="item in tocItems"
              :key="item.id"
              :href="`#${item.id}`"
              class="toc-item"
              :style="`padding-left: ${(item.level - 1) * 0.75}rem`"
            >{{ item.text }}</a>
          </nav>
        </aside>
      </div>
    </template>

    <div v-else-if="!forumStore.loading" style="padding:4rem 0;text-align:center;font-weight:900;font-size:.8rem;text-transform:uppercase;letter-spacing:.1em;color:#9ca3af">
      话题不存在
    </div>

    <!-- Back to top button -->
    <button
      v-if="showBackTop"
      class="back-to-top"
      @click="scrollToTop"
    >↑ 顶部</button>
  </div>
</template>

<script setup lang="ts">
import { computed, ref, onMounted, onBeforeUnmount, watch } from 'vue'
import { useRoute, useRouter, RouterLink } from 'vue-router'
import { useForumStore } from '@/stores/forum'
import { useAuthStore } from '@/stores/auth'
import { useMarkdownRenderer } from '@/composables/useMarkdownRenderer'
import type { ForumReply } from '@/types'
import ABtn from '@/components/ui/ABtn.vue'
import AEmpty from '@/components/ui/AEmpty.vue'
import MarkdownEditor from '@/components/blog/MarkdownEditor.vue'
import ForumReplyNode from '@/components/forum/ForumReplyNode.vue'

const route = useRoute()
const router = useRouter()
const forumStore = useForumStore()
const authStore = useAuthStore()
const { renderMarkdown } = useMarkdownRenderer()

const replyContent = ref('')
const submitting = ref(false)
const quotedReply = ref<ForumReply | null>(null)
const replySort = ref<'oldest' | 'best'>('oldest')
const showBackTop = ref(false)
const contentRef = ref<HTMLElement | null>(null)
const tocItems = ref<Array<{ id: string; text: string; level: number }>>([])
const draftRestored = ref(false)

const replyLookup = computed(() => {
  const lookup = new Map<string, ForumReply>()
  forumStore.replies.forEach((reply) => lookup.set(reply.id, reply))
  return lookup
})

const replySortOptions: Record<string, string> = {
  oldest: '时间',
  best: '最赞',
}

const REPLY_DRAFT_KEY = () => `reply:${route.params.id}`

// ─── Draft persistence for replies ──────────────────────────────────────────

let autosaveTimer: ReturnType<typeof setInterval> | null = null

const saveReplyDraft = () => {
  if (replyContent.value.trim()) {
    forumStore.saveDraftLocal(REPLY_DRAFT_KEY(), { context_key: REPLY_DRAFT_KEY(), content: replyContent.value })
  }
}

const clearReplyDraft = () => {
  forumStore.clearDraftLocal(REPLY_DRAFT_KEY())
  replyContent.value = ''
  draftRestored.value = false
}

// ─── Table of Contents ───────────────────────────────────────────────────────

const buildToC = () => {
  if (!contentRef.value) return
  const headings = contentRef.value.querySelectorAll('h1, h2, h3')
  const items: typeof tocItems.value = []
  headings.forEach((el, i) => {
    const level = parseInt(el.tagName.slice(1))
    const text = el.textContent || ''
    const id = `toc-heading-${i}`
    el.id = id
    items.push({ id, text, level })
  })
  tocItems.value = items
}

// ─── Actions ─────────────────────────────────────────────────────────────────

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

const getReplyAuthor = (reply: ForumReply) => {
  return reply.user?.display_name || reply.user?.username || '匿名'
}

const getReplyPreview = (content: string) => {
  return content.replace(/\s+/g, ' ').trim().slice(0, 140)
}

const getQuotedReply = (reply: ForumReply) => {
  if (!reply.parent_reply_id) return null
  return replyLookup.value.get(reply.parent_reply_id) ?? null
}

const setQuote = (reply: ForumReply) => {
  quotedReply.value = reply
  document.getElementById('reply-form')?.scrollIntoView({ behavior: 'smooth' })
}

const clearQuote = () => {
  quotedReply.value = null
}

const setReplySort = async (s: 'oldest' | 'best') => {
  replySort.value = s
  await forumStore.fetchReplies(route.params.id as string, s)
}

const submitReply = async () => {
  if (!replyContent.value.trim()) return
  submitting.value = true
  const newReply = await forumStore.createReply(
    route.params.id as string,
    replyContent.value.trim(),
    quotedReply.value?.id,
  )
  if (newReply) {
    clearReplyDraft()
    clearQuote()
    await forumStore.fetchReplies(route.params.id as string, replySort.value)
    if (forumStore.currentTopic) forumStore.currentTopic.reply_count++
  }
  submitting.value = false
}

const handleDeleteReply = async (replyId: string) => {
  await forumStore.deleteReply(replyId)
  await forumStore.fetchReplies(route.params.id as string, replySort.value)
  if (forumStore.currentTopic) forumStore.currentTopic.reply_count--
}

const scrollToTop = () => window.scrollTo({ top: 0, behavior: 'smooth' })

const onScroll = () => {
  showBackTop.value = window.scrollY > 300
}

// ─── Lifecycle ───────────────────────────────────────────────────────────────

onMounted(async () => {
  const id = route.params.id as string
  await forumStore.fetchTopic(id)
  await forumStore.fetchReplies(id)

  // Build ToC after content renders
  setTimeout(buildToC, 100)

  // Restore reply draft
  const draft = forumStore.loadDraftLocal(REPLY_DRAFT_KEY())
  if (draft?.content) {
    replyContent.value = draft.content
    draftRestored.value = true
  }

  // Autosave every 2s
  autosaveTimer = setInterval(saveReplyDraft, 2000)
  window.addEventListener('scroll', onScroll)
})

onBeforeUnmount(() => {
  if (autosaveTimer) clearInterval(autosaveTimer)
  window.removeEventListener('scroll', onScroll)
})

watch(
  () => forumStore.currentTopic,
  () => setTimeout(buildToC, 100),
)
</script>

<style scoped>
/* ── Topic header ─────────────────────────────────────────────────────────── */
.topic-header {
  margin-bottom: 2rem;
}

.topic-title {
  font-size: 2rem;
  font-weight: 900;
  letter-spacing: -0.04em;
  line-height: 1.15;
  margin: 0 0 0.75rem;
}

.badge-pinned {
  font-size: 0.7rem;
  font-weight: 900;
  text-transform: uppercase;
  letter-spacing: 0.1em;
  padding: 0.15rem 0.4rem;
  border: 1.5px solid #000;
  margin-right: 0.6rem;
  vertical-align: middle;
}

.badge-closed {
  font-size: 0.7rem;
  font-weight: 900;
  text-transform: uppercase;
  letter-spacing: 0.1em;
  padding: 0.15rem 0.4rem;
  border: 1.5px solid #9ca3af;
  color: #9ca3af;
  margin-right: 0.6rem;
  vertical-align: middle;
}

.topic-meta {
  display: flex;
  align-items: center;
  gap: 1.5rem;
  font-size: 0.8rem;
  font-weight: 700;
  color: #6b7280;
  flex-wrap: wrap;
}

.action-btn {
  font-size: 0.75rem;
  font-weight: 900;
  text-transform: uppercase;
  letter-spacing: 0.08em;
  padding: 0.35rem 0.75rem;
  border: 1.5px solid #000;
  cursor: pointer;
  transition: all 0.15s;
  background: #fff;
  color: #000;
}

.action-btn:hover {
  background: #000;
  color: #fff;
}

.action-btn-active {
  background: #000;
  color: #fff;
}

.action-btn-active:hover {
  background: #fff;
  color: #000;
}

/* ── Layout ───────────────────────────────────────────────────────────────── */
.topic-layout {
  display: flex;
  gap: 2rem;
  align-items: flex-start;
}

.topic-main {
  flex: 1;
  min-width: 0;
}

/* ── Table of Contents ────────────────────────────────────────────────────── */
.topic-toc {
  width: 14rem;
  flex-shrink: 0;
  position: sticky;
  top: 5rem;
  border: 2px solid #000;
  padding: 1rem 1.25rem;
  background: #fff;
  max-height: 70vh;
  overflow-y: auto;
}

.toc-item {
  display: block;
  font-size: 0.75rem;
  font-weight: 700;
  color: #6b7280;
  text-decoration: none;
  padding: 0.25rem 0;
  border-bottom: 1px solid #f3f4f6;
  transition: color 0.15s;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.toc-item:hover {
  color: #000;
}

@media (max-width: 1024px) {
  .topic-toc {
    display: none;
  }
}

/* ── Reply form ───────────────────────────────────────────────────────────── */
.reply-closed-notice {
  border: 2px solid #9ca3af;
  padding: 1.25rem 1.5rem;
  text-align: center;
  font-weight: 900;
  font-size: 0.8rem;
  text-transform: uppercase;
  letter-spacing: 0.1em;
  color: #6b7280;
}

.reply-login-notice {
  border: 2px solid #000;
  padding: 1.25rem 1.5rem;
  text-align: center;
}

.reply-form-wrap {
  border: 2px solid #000;
  padding: 1.5rem;
}

.reply-editor-wrap {
  height: 300px;
  display: flex;
  flex-direction: column;
}

/* ── Markdown body ────────────────────────────────────────────────────────── */
.markdown-body :deep(h1),
.markdown-body :deep(h2),
.markdown-body :deep(h3) {
  font-weight: 900;
  letter-spacing: -0.03em;
  margin: 1.5em 0 0.75em;
  scroll-margin-top: 5rem;
}
.markdown-body :deep(p) {
  line-height: 1.75;
  margin: 0.75em 0;
}
.markdown-body :deep(pre) {
  background: #f3f4f6;
  border: 2px solid #000;
  padding: 1rem;
  overflow-x: auto;
  margin: 1rem 0;
}
.markdown-body :deep(code) {
  font-family: monospace;
  font-size: 0.875em;
}
.markdown-body :deep(pre code) {
  font-size: 0.875em;
}
.markdown-body :deep(blockquote) {
  border-left: 3px solid #000;
  padding-left: 1rem;
  margin: 1rem 0;
  color: #6b7280;
}
.markdown-body :deep(a) {
  color: #000;
  text-decoration: underline;
}
.markdown-body :deep(img) {
  max-width: 100%;
  filter: grayscale(100%);
}
.markdown-body :deep(table) {
  border-collapse: collapse;
  width: 100%;
  margin: 1rem 0;
}
.markdown-body :deep(th),
.markdown-body :deep(td) {
  border: 2px solid #000;
  padding: 0.5rem 0.75rem;
}
.markdown-body :deep(th) {
  font-weight: 900;
  background: #f3f4f6;
}

/* ── Back to top ──────────────────────────────────────────────────────────── */
.back-to-top {
  position: fixed;
  bottom: 2rem;
  right: 2rem;
  z-index: 50;
  font-size: 0.7rem;
  font-weight: 900;
  text-transform: uppercase;
  letter-spacing: 0.1em;
  padding: 0.5rem 0.875rem;
  border: 2px solid #000;
  background: #fff;
  cursor: pointer;
  transition: all 0.15s;
  box-shadow: 4px 4px 0px 0px rgba(0,0,0,1);
}

.back-to-top:hover {
  background: #000;
  color: #fff;
}
</style>
