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
          style="font-weight:900;font-size:.75rem;text-transform:uppercase;letter-spacing:.1em;text-decoration:none;color:var(--a-color-muted);border-bottom:1px solid transparent;transition:border-color .2s"
          @mouseenter="($event.currentTarget as HTMLElement).style.borderBottomColor='var(--a-color-muted)'"
          @mouseleave="($event.currentTarget as HTMLElement).style.borderBottomColor='transparent'"
        >论坛</RouterLink>
        <span style="color:var(--a-color-disabled-border)">/</span>
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
          <span v-if="forumStore.currentTopic.featured" class="badge-featured">精华</span>
          <span v-if="forumStore.currentTopic.closed" class="badge-closed">已关闭</span>
          {{ forumStore.currentTopic.title }}
        </h1>

        <!-- Tags -->
        <div v-if="(forumStore.currentTopic.tags || []).length > 0" style="display:flex;flex-wrap:wrap;gap:.4rem;margin-bottom:.75rem">
          <span
            v-for="tag in forumStore.currentTopic.tags"
            :key="tag"
            style="font-size:.65rem;font-weight:700;padding:.1rem .5rem;border:1.5px solid var(--a-color-disabled-border);color:var(--a-color-muted)"
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

          <!-- Report button (non-owner, authenticated) -->
          <button
            v-if="authStore.isAuthenticated && authStore.user?.uuid !== forumStore.currentTopic.user_id"
            @click="reportModalOpen = true"
            class="action-btn"
          >举报</button>

          <!-- Admin: feature/unfeature -->
          <button
            v-if="authStore.user?.role === 'admin'"
            @click="toggleFeatured"
            class="action-btn"
            :class="{ 'action-btn-active': forumStore.currentTopic.featured }"
          >{{ forumStore.currentTopic.featured ? '取消精华' : '设为精华' }}</button>
        </div>
      </div>

      <!-- Layout: content + optional ToC -->
      <div class="topic-layout">
        <div class="topic-main">
          <!-- Topic content (Markdown rendered) -->
          <div
            class="markdown-body"
            style="border:var(--a-border);padding:2rem;margin-bottom:2.5rem;background:var(--a-color-bg)"
            v-html="renderMarkdown(forumStore.currentTopic.content)"
            ref="contentRef"
          />

          <!-- Reply sort -->
          <div style="display:flex;align-items:center;justify-content:space-between;margin-bottom:1.25rem">
            <h2 style="font-weight:900;font-size:.75rem;text-transform:uppercase;letter-spacing:.15em;margin:0;padding-bottom:.75rem;border-bottom:var(--a-border);flex:1">
              {{ forumStore.currentTopic.reply_count }} 条回复
            </h2>
            <div style="display:flex;gap:0;border:var(--a-border);margin-left:1rem">
              <button
                v-for="(label, s) in replySortOptions"
                :key="s"
                class="a-tab-btn"
                :class="{ 'a-tab-btn-active': replySort === s }"
                style="padding:.35rem .875rem;font-size:.7rem;border-right:var(--a-border)"
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
                style="display:flex;align-items:flex-start;justify-content:space-between;gap:1rem;padding:.8rem 1rem;background:var(--a-color-disabled-bg);border-left:3px solid var(--a-color-fg);margin-bottom:1rem;font-size:.8rem;font-weight:700"
              >
                <div>
                  <div>引用 #{{ quotedReply.floor_number }} @{{ getReplyAuthor(quotedReply) }} 的回复</div>
                  <div style="margin-top:.35rem;font-size:.75rem;font-weight:500;color:var(--a-color-muted);line-height:1.6">
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

              <!-- AEditor for reply -->
              <div class="reply-editor-wrap">
                <AEditor
                  v-model="replyContent"
                  mode="plain"
                  :enable-mentions="true"
                  placeholder="写下你的回复…（支持 Markdown，@ 提及用户）"
                />
              </div>

              <div style="display:flex;justify-content:flex-end;margin-top:.75rem;gap:.75rem">
                <button
                  v-if="replyContent"
                  type="button"
                  style="font-size:.7rem;font-weight:900;text-transform:uppercase;letter-spacing:.08em;background:none;border:none;cursor:pointer;color:var(--a-color-muted-soft)"
                  @click="clearReplyDraft"
                >清除草稿</button>
                <ABtn @click="submitReply" :loading="submitting" :disabled="!replyContent.trim()">提交回复</ABtn>
              </div>
            </div>
          </div>
        </div>

        <!-- Table of Contents sidebar (desktop only) -->
        <aside v-if="tocItems.length > 0" class="topic-toc">
          <div style="font-size:.65rem;font-weight:900;text-transform:uppercase;letter-spacing:.12em;margin-bottom:.75rem;color:var(--a-color-muted-soft)">目录</div>
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

    <div v-else-if="!forumStore.loading" style="padding:4rem 0;text-align:center;font-weight:900;font-size:.8rem;text-transform:uppercase;letter-spacing:.1em;color:var(--a-color-muted-soft)">
      话题不存在
    </div>

    <!-- Back to top button -->
    <button
      v-if="showBackTop"
      class="back-to-top"
      @click="scrollToTop"
    >↑ 顶部</button>
  </div>

  <!-- Report Modal -->
  <div
    v-if="reportModalOpen"
    style="position:fixed;inset:0;background:rgba(0,0,0,.45);z-index:1000;display:flex;align-items:center;justify-content:center"
    @click.self="reportModalOpen = false"
  >
    <div style="background:var(--a-color-bg);border:var(--a-border);padding:1.5rem;width:min(420px,90vw);display:flex;flex-direction:column;gap:1rem">
      <h3 style="margin:0;font-size:.9rem;font-weight:900;text-transform:uppercase">举报内容</h3>
      <div style="display:flex;flex-direction:column;gap:.5rem">
        <label style="font-size:.75rem;font-weight:700">举报原因 *</label>
        <select v-model="reportForm.reason" style="border:var(--a-border);padding:.5rem;background:var(--a-color-bg);font-size:.85rem">
          <option value="">请选择原因</option>
          <option value="spam">垃圾广告</option>
          <option value="harassment">骚扰攻击</option>
          <option value="misinformation">虚假信息</option>
          <option value="off-topic">偏离主题</option>
          <option value="other">其他</option>
        </select>
      </div>
      <div style="display:flex;flex-direction:column;gap:.5rem">
        <label style="font-size:.75rem;font-weight:700">补充说明</label>
        <textarea v-model="reportForm.note" style="border:var(--a-border);padding:.5rem;background:var(--a-color-bg);font-size:.85rem;resize:vertical;min-height:80px" placeholder="可选：详细说明" />
      </div>
      <div style="display:flex;gap:.5rem;justify-content:flex-end">
        <button @click="reportModalOpen = false" style="padding:.5rem 1rem;border:var(--a-border);background:none;cursor:pointer;font-size:.8rem">取消</button>
        <button @click="submitReport" style="padding:.5rem 1rem;border:none;background:var(--a-color-fg);color:var(--a-color-bg);cursor:pointer;font-size:.8rem;font-weight:700">提交举报</button>
      </div>
    </div>
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
import AEditor from '@/components/shared/AEditor.vue'
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

// Report & Featured
const API_URL = import.meta.env.VITE_API_URL || '/api'
const reportModalOpen = ref(false)
const reportForm = ref({ reason: '', note: '' })

const submitReport = async () => {
  if (!reportForm.value.reason.trim()) { alert('请选择举报原因'); return }
  const res = await fetch(`${API_URL}/forum/report`, {
    method: 'POST',
    headers: {
      'Content-Type': 'application/json',
      Authorization: `Bearer ${authStore.token}`,
    },
    body: JSON.stringify({
      target_type: 'topic',
      target_id: forumStore.currentTopic!.id,
      reason: reportForm.value.reason,
      note: reportForm.value.note,
    }),
  })
  if (res.ok) {
    reportModalOpen.value = false
    reportForm.value = { reason: '', note: '' }
    alert('举报已提交')
  } else {
    const d = await res.json()
    alert(`举报失败: ${d.error || '未知错误'}`)
  }
}

const toggleFeatured = async () => {
  const topic = forumStore.currentTopic!
  const method = topic.featured ? 'DELETE' : 'POST'
  const res = await fetch(`${API_URL}/forum/topics/${topic.id}/feature`, {
    method,
    headers: { Authorization: `Bearer ${authStore.token}` },
  })
  if (res.ok) {
    topic.featured = !topic.featured
  }
}
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
  border: 1.5px solid var(--a-color-fg);
  margin-right: 0.6rem;
  vertical-align: middle;
}

.badge-closed {
  font-size: 0.7rem;
  font-weight: 900;
  text-transform: uppercase;
  letter-spacing: 0.1em;
  padding: 0.15rem 0.4rem;
  border: 1.5px solid var(--a-color-muted-soft);
  color: var(--a-color-muted-soft);
  margin-right: 0.6rem;
  vertical-align: middle;
}

.badge-featured {
  font-size: 0.7rem;
  font-weight: 900;
  text-transform: uppercase;
  letter-spacing: 0.1em;
  padding: 0.15rem 0.4rem;
  border: 1.5px solid var(--a-color-accent, #f59e0b);
  color: var(--a-color-accent, #f59e0b);
  margin-right: 0.6rem;
  vertical-align: middle;
}

.topic-meta {
  display: flex;
  align-items: center;
  gap: 1.5rem;
  font-size: 0.8rem;
  font-weight: 700;
  color: var(--a-color-muted);
  flex-wrap: wrap;
}

.action-btn {
  font-size: 0.75rem;
  font-weight: 900;
  text-transform: uppercase;
  letter-spacing: 0.08em;
  padding: 0.35rem 0.75rem;
  border: 1.5px solid var(--a-color-fg);
  cursor: pointer;
  transition: all 0.15s;
  background: var(--a-color-bg);
  color: var(--a-color-fg);
}

.action-btn:hover {
  background: var(--a-color-fg);
  color: var(--a-color-bg);
}

.action-btn-active {
  background: var(--a-color-fg);
  color: var(--a-color-bg);
}

.action-btn-active:hover {
  background: var(--a-color-bg);
  color: var(--a-color-fg);
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
  border: var(--a-border);
  padding: 1rem 1.25rem;
  background: var(--a-color-bg);
  max-height: 70vh;
  overflow-y: auto;
}

.toc-item {
  display: block;
  font-size: 0.75rem;
  font-weight: 700;
  color: var(--a-color-muted);
  text-decoration: none;
  padding: 0.25rem 0;
  border-bottom: 1px solid var(--a-color-disabled-bg);
  transition: color 0.15s;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.toc-item:hover {
  color: var(--a-color-fg);
}

@media (max-width: 1024px) {
  .topic-toc {
    display: none;
  }
}

/* ── Reply form ───────────────────────────────────────────────────────────── */
.reply-closed-notice {
  border: 2px solid var(--a-color-muted-soft);
  padding: 1.25rem 1.5rem;
  text-align: center;
  font-weight: 900;
  font-size: 0.8rem;
  text-transform: uppercase;
  letter-spacing: 0.1em;
  color: var(--a-color-muted);
}

.reply-login-notice {
  border: var(--a-border);
  padding: 1.25rem 1.5rem;
  text-align: center;
}

.reply-form-wrap {
  border: var(--a-border);
  padding: 1.5rem;
}

.reply-login-text {
  font-weight: 700;
  font-size: 0.9rem;
  margin: 0 0 1rem;
}

.reply-form-title {
  font-weight: 900;
  font-size: 0.7rem;
  text-transform: uppercase;
  letter-spacing: 0.15em;
  margin: 0 0 1rem;
}

.reply-actions {
  display: flex;
  justify-content: flex-end;
  margin-top: 0.75rem;
  gap: 0.75rem;
}

.reply-draft-clear {
  font-size: 0.7rem;
  font-weight: 900;
  text-transform: uppercase;
  letter-spacing: 0.08em;
  background: none;
  border: none;
  cursor: pointer;
  color: var(--a-color-muted-soft);
}

.reply-draft-clear:hover {
  text-decoration: underline;
}

.topic-not-found {
  padding: 4rem 0;
  text-align: center;
  font-weight: 900;
  font-size: 0.8rem;
  text-transform: uppercase;
  letter-spacing: 0.1em;
  color: var(--a-color-muted-soft);
}

.topic-divider {
  color: var(--a-color-disabled-border);
}

.topic-content-card {
  border: var(--a-border);
  padding: 2rem;
  margin-bottom: 2.5rem;
  background: var(--a-color-bg);
}

.reply-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  margin-bottom: 1.25rem;
}

.reply-count-title {
  font-weight: 900;
  font-size: 0.75rem;
  text-transform: uppercase;
  letter-spacing: 0.15em;
  margin: 0;
  padding-bottom: 0.75rem;
  border-bottom: var(--a-border);
  flex: 1;
}

.reply-sort-tabs {
  display: flex;
  gap: 0;
  border: var(--a-border);
  margin-left: 1rem;
}

.reply-sort-tab {
  padding: 0.35rem 0.875rem;
  font-size: 0.7rem;
  border-right: var(--a-border);
}

.reply-sort-tab:last-child {
  border-right: none;
}

.reply-list {
  display: flex;
  flex-direction: column;
  gap: 0.75rem;
}

.quote-box {
  display: flex;
  align-items: flex-start;
  justify-content: space-between;
  gap: 1rem;
  padding: 0.8rem 1rem;
  background: var(--a-color-disabled-bg);
  border-left: 3px solid var(--a-color-fg);
  margin-bottom: 1rem;
  font-size: 0.8rem;
  font-weight: 700;
}

.quote-preview {
  margin-top: 0.35rem;
  font-size: 0.75rem;
  font-weight: 500;
  color: var(--a-color-muted);
  line-height: 1.6;
}

.clear-quote-btn {
  font-size: 0.7rem;
  font-weight: 900;
  text-transform: uppercase;
  letter-spacing: 0.08em;
  background: none;
  border: none;
  cursor: pointer;
}

.draft-restored {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 0.6rem 1rem;
  background: #fef3c7;
  border: 1.5px solid #f59e0b;
  margin-bottom: 1rem;
  font-size: 0.8rem;
  font-weight: 700;
}

.clear-draft-btn {
  background: none;
  border: none;
  font-size: 0.7rem;
  font-weight: 900;
  text-transform: uppercase;
  letter-spacing: 0.08em;
  cursor: pointer;
  color: #92400e;
}

.toc-title {
  font-size: 0.65rem;
  font-weight: 900;
  text-transform: uppercase;
  letter-spacing: 0.12em;
  margin-bottom: 0.75rem;
  color: var(--a-color-muted-soft);
}

.loading-state {
  padding: 4rem 0;
  text-align: center;
  font-weight: 900;
  letter-spacing: 0.1em;
  font-size: 0.75rem;
  text-transform: uppercase;
}

.page-root {
  padding-bottom: 8rem;
}

.topic-breadcrumb {
  display: flex;
  align-items: center;
  gap: 0.75rem;
  margin-bottom: 1.5rem;
  flex-wrap: wrap;
}

.topic-tag-row {
  display: flex;
  flex-wrap: wrap;
  gap: 0.4rem;
  margin-bottom: 0.75rem;
}

.reply-form-root {
  margin-top: 2.5rem;
}

.back-link-muted {
  font-weight: 900;
  font-size: 0.75rem;
  text-transform: uppercase;
  letter-spacing: 0.1em;
  text-decoration: none;
  color: var(--a-color-muted);
  border-bottom: 1px solid transparent;
  transition: border-color 0.2s;
}

.back-link-muted:hover {
  border-bottom-color: var(--a-color-muted);
}

.category-pill {
  font-size: 0.65rem;
  font-weight: 900;
  text-transform: uppercase;
  letter-spacing: 0.08em;
  padding: 0.15rem 0.5rem;
  border: 1.5px solid;
  cursor: pointer;
}

.tag-pill {
  font-size: 0.65rem;
  font-weight: 700;
  padding: 0.1rem 0.5rem;
  border: 1.5px solid var(--a-color-disabled-border);
  color: var(--a-color-muted);
}

.reply-login-cta {
  margin-top: 1rem;
}

.reply-form-grid {
  display: flex;
  flex-direction: column;
  gap: 0.75rem;
}

.reply-sort-wrap {
  display: flex;
  align-items: center;
  justify-content: space-between;
  margin-bottom: 1.25rem;
}

.reply-scroll-list {
  display: flex;
  flex-direction: column;
  gap: 0.75rem;
}

.reply-content-note {
  color: var(--a-color-muted);
}

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
  border: var(--a-border);
  background: var(--a-color-bg);
  cursor: pointer;
  transition: all 0.15s;
  box-shadow: var(--a-shadow-button);
}

.back-to-top:hover {
  background: var(--a-color-fg);
  color: var(--a-color-bg);
}

.markdown-body :deep(pre) {
  background: var(--a-color-disabled-bg);
  border: var(--a-border);
  padding: 1rem;
  overflow-x: auto;
  margin: 1rem 0;
}

.markdown-body :deep(blockquote) {
  border-left: 3px solid var(--a-color-fg);
  padding-left: 1rem;
  margin: 1rem 0;
  color: var(--a-color-muted);
}

.markdown-body :deep(a) {
  color: var(--a-color-fg);
  text-decoration: underline;
}

.markdown-body :deep(th),
.markdown-body :deep(td) {
  border: var(--a-border);
  padding: 0.5rem 0.75rem;
}

.markdown-body :deep(th) {
  font-weight: 900;
  background: var(--a-color-disabled-bg);
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

.markdown-body :deep(pre code) {
  font-size: 0.875em;
}

.markdown-body :deep(code) {
  font-family: monospace;
  font-size: 0.875em;
}

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

.markdown-body :deep(toc-item:hover) {
  color: var(--a-color-fg);
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
  background: var(--a-color-disabled-bg);
  border: var(--a-border);
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
  border-left: 3px solid var(--a-color-fg);
  padding-left: 1rem;
  margin: 1rem 0;
  color: var(--a-color-muted);
}
.markdown-body :deep(a) {
  color: var(--a-color-fg);
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
  border: var(--a-border);
  padding: 0.5rem 0.75rem;
}
.markdown-body :deep(th) {
  font-weight: 900;
  background: var(--a-color-disabled-bg);
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
  border: var(--a-border);
  background: var(--a-color-bg);
  cursor: pointer;
  transition: all 0.15s;
  box-shadow: var(--a-shadow-button);
}

.back-to-top:hover {
  background: var(--a-color-fg);
  color: var(--a-color-bg);
}
</style>
