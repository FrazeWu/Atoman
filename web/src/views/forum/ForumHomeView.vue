<template>
  <div class="forum-page">
    <!-- ── Left sidebar ──────────────────────────────────────────────────── -->
    <aside class="forum-sidebar">
      <!-- Nav section -->
      <div class="sidebar-section">
        <div
          class="sidebar-item sidebar-item-active"
          @click="selectCategory(null)"
          :class="{ 'sidebar-item-active': !activeCategoryId && activeTab === 'latest' }"
        >
          <span class="sidebar-item-icon">☰</span>
          <span>Topics</span>
        </div>
        <div class="sidebar-item sidebar-sub" @click="router.push('/forum/new')" v-if="authStore.isAuthenticated">
          <span class="sidebar-item-icon">✦</span>
          <span>Upcoming events</span>
        </div>
        <div class="sidebar-item sidebar-sub">
          <span class="sidebar-item-icon">◎</span>
          <span>More</span>
        </div>
      </div>

      <div class="sidebar-divider" />

      <!-- Develop section -->
      <div class="sidebar-section">
        <div class="sidebar-section-label">DEVELOP</div>
        <div
          v-for="cat in forumStore.categories"
          :key="cat.id"
          class="sidebar-item"
          :class="{ 'sidebar-item-active': activeCategoryId === cat.id }"
          @click="selectCategory(cat.id)"
        >
          <span
            class="sidebar-cat-dot"
            :style="{ background: cat.color || 'var(--a-color-fg)' }"
          />
          <span class="sidebar-item-text">{{ cat.name }}</span>
          <span class="sidebar-item-count">{{ cat.topic_count || 0 }}</span>
        </div>
      </div>

      <div class="sidebar-divider" />

      <!-- Tags section -->
      <div class="sidebar-section">
        <div class="sidebar-section-label">TAGS</div>
        <div
          v-for="tag in popularTags"
          :key="tag"
          class="sidebar-tag"
          :class="{ 'sidebar-tag-active': activeTag === tag }"
          @click="filterByTag(tag)"
        >
          {{ tag }}
        </div>
      </div>

      <!-- Keyboard shortcuts hint -->
      <div class="sidebar-shortcuts">
        <div><kbd>J</kbd><kbd>K</kbd> 上下选择</div>
        <div><kbd>Enter</kbd> 打开话题</div>
        <div><kbd>N</kbd> 发新话题</div>
        <div><kbd>/</kbd> 搜索</div>
      </div>
    </aside>

    <!-- ── Main content ──────────────────────────────────────────────────── -->
    <main class="forum-main">
      <!-- Top tab bar -->
      <div class="tab-bar">
        <div class="tab-bar-left">
          <ABtn
            v-for="(label, key) in tabOptions"
            :key="key"
            outline
            size="sm"
            class="forum-tab-btn"
            :class="{ 'forum-tab-btn-active': activeTab === key }"
            @click="setTab(key as TabKey)"
          >
            {{ label }}
          </ABtn>
        </div>
        <div class="tab-bar-right">
          <ABtn v-if="authStore.isAuthenticated" outline size="sm" @click="router.push('/forum/new')">
            发新话题
          </ABtn>
          <ABtn v-if="authStore.isAuthenticated" outline size="sm" @click="catReqModalOpen = true">
            申请分类
          </ABtn>
        </div>
      </div>

      <!-- Filter bar -->
      <div class="filter-bar">
        <div class="filter-left">
          <div class="forum-category-select">
            <ASelect
              v-model="selectedCategoryValue"
              :options="categoryOptions"
              placeholder="全部分类"
            />
          </div>

          <div v-if="activeTag" class="active-tag-chip a-badge a-badge-fill">
            # {{ activeTag }}
            <button @click="clearTag" class="tag-chip-close">×</button>
          </div>
        </div>

        <div class="filter-right">
          <div class="search-wrap">
            <AInput
              v-model="searchQuery"
              placeholder="搜索话题..."
              class="search-input"
              @keydown.enter="doSearch"
              @input="onSearchInput"
              ref="searchInputRef"
            />
            <button v-if="searchQuery" class="search-clear" @click="clearSearch">×</button>
          </div>
        </div>
      </div>

      <!-- Topic list header row -->
      <div class="topic-list-header">
        <span class="th-title a-label">话题</span>
        <span class="th-stats a-label">
          <span>回复</span>
          <span>浏览</span>
          <span>活跃</span>
        </span>
      </div>

      <!-- Loading state -->
      <div v-if="forumStore.loading" class="topic-list">
        <div v-for="i in 8" :key="i" class="topic-row-skeleton a-skeleton" />
      </div>

      <!-- Empty state -->
      <AEmpty v-else-if="forumStore.topics.length === 0" text="暂无话题，来发第一个吧" />

      <!-- Topic rows -->
      <div v-else ref="topicListRef" class="topic-list">
        <div
          v-for="(topic, index) in forumStore.topics"
          :key="topic.id"
          class="topic-row a-card"
          :class="{
            'topic-row-pinned': topic.pinned,
            'topic-row-focused': focusedIndex === index,
          }"
          @click="router.push(`/topic/${topic.id}`)"
        >
          <!-- Left: category dot + title column -->
          <div class="tr-left">
            <!-- Category badge -->
            <div class="tr-tags">
              <span v-if="topic.pinned" class="tr-badge a-badge tr-badge-pin">置顶</span>
              <span v-if="topic.closed" class="tr-badge a-badge tr-badge-closed">已关闭</span>
              <span
                v-if="topic.category"
                class="tr-badge a-badge tr-badge-cat"
                :style="{ borderColor: topic.category.color, color: topic.category.color }"
                @click.stop="selectCategory(topic.category!.id)"
              >{{ topic.category.name }}</span>
              <span
                v-for="tag in (topic.tags || []).slice(0, 2)"
                :key="tag"
                class="tr-badge a-badge tr-badge-tag"
                @click.stop="filterByTag(tag)"
              ># {{ tag }}</span>
            </div>

            <!-- Title -->
            <p class="tr-title">{{ topic.title }}</p>

            <!-- Author + time on mobile -->
            <p class="tr-meta">
              <span class="tr-author">{{ topic.user?.display_name || topic.user?.username || '匿名' }}</span>
              <span class="tr-sep">·</span>
              <span>{{ formatTime(topic.created_at) }}</span>
              <template v-if="topic.is_bookmarked">
                <span class="tr-sep">·</span>
                <span class="tr-bookmarked">收藏</span>
              </template>
            </p>
          </div>

          <!-- Right: participant avatars + stats -->
          <div class="tr-right">
            <!-- Participant avatars -->
            <div class="tr-avatars">
              <div class="tr-avatar tr-avatar-op" :title="topic.user?.display_name || topic.user?.username">
                {{ avatarInitial(topic.user?.display_name || topic.user?.username) }}
              </div>
            </div>

            <!-- Stats -->
            <div class="tr-stats">
              <span class="tr-stat">
                <span class="tr-stat-val">{{ topic.reply_count }}</span>
              </span>
              <span class="tr-stat">
                <span class="tr-stat-val">{{ formatCount(topic.view_count) }}</span>
              </span>
              <span class="tr-stat tr-stat-time">
                {{ formatTimeShort(topic.last_reply_at || topic.created_at) }}
              </span>
            </div>
          </div>
        </div>
      </div>

      <!-- Load more -->
      <div v-if="forumStore.topics.length < forumStore.topicsTotal && !forumStore.loading" class="load-more-wrap">
        <ABtn outline @click="loadMore" :loading="loadingMore">加载更多</ABtn>
      </div>
    </main>
  </div>

  <!-- Category Request Modal -->
  <AModal v-if="catReqModalOpen" @close="catReqModalOpen = false" size="md">
    <h3 class="a-subtitle" style="margin-bottom:1.25rem">申请新分类</h3>
    <div style="display:flex;flex-direction:column;gap:1rem">
      <AInput v-model="catReqForm.name" label="分类名称 *" placeholder="分类名称" />
      <ATextarea v-model="catReqForm.description" label="描述" :rows="3" placeholder="分类用途说明" />
      <ATextarea v-model="catReqForm.reason" label="申请理由 *" :rows="4" placeholder="说明为什么需要此分区" />
    </div>
    <div style="display:flex;gap:.5rem;justify-content:flex-end;margin-top:1.5rem">
      <ABtn outline @click="catReqModalOpen = false">取消</ABtn>
      <ABtn @click="submitCategoryRequest">提交申请</ABtn>
    </div>
  </AModal>
</template>

<script setup lang="ts">
import { ref, computed, onMounted, onBeforeUnmount } from 'vue'
import { useRouter } from 'vue-router'
import { useForumStore } from '@/stores/forum'
import { useAuthStore } from '@/stores/auth'
import ABtn from '@/components/ui/ABtn.vue'
import AEmpty from '@/components/ui/AEmpty.vue'
import AInput from '@/components/ui/AInput.vue'
import ASelect from '@/components/ui/ASelect.vue'
import ATextarea from '@/components/ui/ATextarea.vue'
import AModal from '@/components/ui/AModal.vue'

type TabKey = 'latest' | 'top' | 'active' | 'new' | 'bookmarked' | 'featured'

const router = useRouter()
const forumStore = useForumStore()
const authStore = useAuthStore()

const activeCategoryId = ref<string | null>(null)
const activeTab = ref<TabKey>('latest')
const activeTag = ref('')
const page = ref(1)
const loadingMore = ref(false)
const focusedIndex = ref(-1)
const searchQuery = ref('')
const topicListRef = ref<HTMLElement | null>(null)
const searchInputRef = ref<HTMLInputElement | null>(null)
const ALL_CATEGORY_VALUE = '__all__'

const tabOptions: Record<TabKey, string> = {
  latest: '最新',
  active: '最活跃',
  top: '最热',
  new: '未读',
  bookmarked: '已收藏',
  featured: '精选',
}

const sortMap: Record<TabKey, 'latest' | 'top' | 'active' | 'featured'> = {
  latest: 'latest',
  active: 'active',
  top: 'top',
  new: 'latest',
  bookmarked: 'latest',
  featured: 'featured',
}

// Popular tags — derived from loaded topics
const popularTags = computed(() => {
  const tagCount: Record<string, number> = {}
  forumStore.topics.forEach((t) => {
    ;(t.tags || []).forEach((tag) => {
      tagCount[tag] = (tagCount[tag] || 0) + 1
    })
  })
  return Object.entries(tagCount)
    .sort((a, b) => b[1] - a[1])
    .slice(0, 12)
    .map(([tag]) => tag)
})

const categoryOptions = computed(() => [
  { label: '全部分类', value: ALL_CATEGORY_VALUE },
  ...forumStore.categories.map((cat) => ({ label: cat.name, value: cat.id })),
])

const selectedCategoryValue = computed({
  get: () => activeCategoryId.value ?? ALL_CATEGORY_VALUE,
  set: (value: string | number | null) => {
    void selectCategory(value === ALL_CATEGORY_VALUE ? null : String(value))
  },
})

// ─── Formatters ──────────────────────────────────────────────────────────────

const formatTime = (iso: string) => {
  if (!iso) return ''
  const d = new Date(iso)
  const now = Date.now()
  const diff = now - d.getTime()
  const mins = Math.floor(diff / 60000)
  if (mins < 60) return `${mins} 分钟前`
  const hours = Math.floor(mins / 60)
  if (hours < 24) return `${hours} 小时前`
  const days = Math.floor(hours / 24)
  if (days < 30) return `${days} 天前`
  return d.toLocaleDateString('zh-CN')
}

const formatTimeShort = (iso: string) => {
  if (!iso) return ''
  const d = new Date(iso)
  const now = Date.now()
  const diff = now - d.getTime()
  const mins = Math.floor(diff / 60000)
  if (mins < 60) return `${mins}分`
  const hours = Math.floor(mins / 60)
  if (hours < 24) return `${hours}时`
  const days = Math.floor(hours / 24)
  if (days < 30) return `${days}天`
  return d.toLocaleDateString('zh-CN', { month: 'numeric', day: 'numeric' })
}

const formatCount = (n: number) => {
  if (!n) return '0'
  if (n >= 1000) return `${(n / 1000).toFixed(1)}k`
  return String(n)
}

const avatarInitial = (name?: string) => {
  if (!name) return '?'
  return name.charAt(0).toUpperCase()
}

// ─── Data loading ─────────────────────────────────────────────────────────────

const loadTopics = async (resetPage = true) => {
  if (resetPage) {
    page.value = 1
    focusedIndex.value = -1
  }
  await forumStore.fetchTopics({
    categoryId: activeCategoryId.value ?? undefined,
    sort: sortMap[activeTab.value],
    tag: activeTag.value || undefined,
    page: page.value,
  })
}

const selectCategory = async (id: string | null) => {
  activeCategoryId.value = id
  await loadTopics()
}

const setTab = async (tab: TabKey) => {
  activeTab.value = tab
  await loadTopics()
}

const filterByTag = async (tag: string) => {
  activeTag.value = tag
  await loadTopics()
}

const clearTag = async () => {
  activeTag.value = ''
  await loadTopics()
}

const loadMore = async () => {
  loadingMore.value = true
  page.value++
  const query = new URLSearchParams({
    page: String(page.value),
    limit: '20',
    sort: sortMap[activeTab.value],
  })
  if (activeCategoryId.value) query.set('category_id', activeCategoryId.value)
  if (activeTag.value) query.set('tag', activeTag.value)
  const res = await fetch(`/api/forum/topics?${query}`, {
    headers: authStore.isAuthenticated ? { Authorization: `Bearer ${authStore.token}` } : {},
  })
  if (res.ok) {
    const data = await res.json()
    forumStore.topics.push(...(data.data || []))
  }
  loadingMore.value = false
}

// ─── Search ──────────────────────────────────────────────────────────────────

let searchTimer: ReturnType<typeof setTimeout> | null = null

const doSearch = () => {
  if (searchQuery.value.trim()) {
    router.push(`/forum/search?q=${encodeURIComponent(searchQuery.value.trim())}`)
  }
}

const onSearchInput = () => {
  if (searchTimer) clearTimeout(searchTimer)
  searchTimer = setTimeout(() => {
    if (searchQuery.value.trim().length >= 2) doSearch()
  }, 500)
}

const clearSearch = () => {
  searchQuery.value = ''
  if (searchTimer) clearTimeout(searchTimer)
}

// ─── Keyboard navigation ─────────────────────────────────────────────────────

const handleKeydown = (e: KeyboardEvent) => {
  const target = e.target as HTMLElement
  if (target.tagName === 'INPUT' || target.tagName === 'TEXTAREA') return

  const topics = forumStore.topics
  switch (e.key) {
    case 'j':
      e.preventDefault()
      focusedIndex.value = Math.min(focusedIndex.value + 1, topics.length - 1)
      scrollToFocused()
      break
    case 'k':
      e.preventDefault()
      focusedIndex.value = Math.max(focusedIndex.value - 1, 0)
      scrollToFocused()
      break
    case 'Enter':
      if (focusedIndex.value >= 0 && topics[focusedIndex.value]) {
        router.push(`/topic/${topics[focusedIndex.value].id}`)
      }
      break
    case 'n':
      if (authStore.isAuthenticated) router.push('/forum/new')
      break
    case '/':
      e.preventDefault()
      searchInputRef.value?.focus()
      break
  }
}

const scrollToFocused = () => {
  if (!topicListRef.value) return
  const rows = topicListRef.value.querySelectorAll('.topic-row')
  const el = rows[focusedIndex.value] as HTMLElement
  el?.scrollIntoView({ behavior: 'smooth', block: 'nearest' })
}

onMounted(async () => {
  await forumStore.fetchCategories()
  await forumStore.fetchTopics({ sort: 'latest', page: 1 })
  window.addEventListener('keydown', handleKeydown)
})

onBeforeUnmount(() => {
  window.removeEventListener('keydown', handleKeydown)
  if (searchTimer) clearTimeout(searchTimer)
})

// Category Request
const API_URL = import.meta.env.VITE_API_URL || '/api'
const catReqModalOpen = ref(false)
const catReqForm = ref({ name: '', description: '', reason: '' })

const submitCategoryRequest = async () => {
  if (!catReqForm.value.name.trim() || !catReqForm.value.reason.trim()) {
    alert('请填写分区名称和申请理由')
    return
  }
  const res = await fetch(`${API_URL}/forum/category-requests`, {
    method: 'POST',
    headers: {
      'Content-Type': 'application/json',
      Authorization: `Bearer ${authStore.token}`,
    },
    body: JSON.stringify(catReqForm.value),
  })
  if (res.ok) {
    catReqModalOpen.value = false
    catReqForm.value = { name: '', description: '', reason: '' }
    alert('申请已提交，请等待管理员审核')
  } else {
    const d = await res.json()
    alert(`提交失败: ${d.error || '未知错误'}`)
  }
}
</script>

<style scoped>
/* ── Page layout ─────────────────────────────────────────────────────────── */
.forum-page {
  display: flex;
  min-height: calc(100vh - 4rem);
  max-width: 1200px;
  margin: 0 auto;
  padding: 0 1rem;
}

/* ── Sidebar ─────────────────────────────────────────────────────────────── */
.forum-sidebar {
  width: 220px;
  flex-shrink: 0;
  padding: 1.5rem 0 2rem;
  border-right: var(--a-border);
  position: sticky;
  top: 0;
  height: 100vh;
  overflow-y: auto;
}

.sidebar-section {
  padding: 0.5rem 0;
}

.sidebar-section-label {
  font-size: 0.6rem;
  font-weight: 900;
  letter-spacing: 0.15em;
  text-transform: uppercase;
  color: var(--a-color-muted-soft);
  padding: 0.25rem 1.25rem 0.5rem;
}

.sidebar-divider {
  border-bottom: 1px solid var(--a-color-disabled-border);
  margin: 0.5rem 0;
}

.sidebar-item {
  display: flex;
  align-items: center;
  gap: 0.6rem;
  padding: 0.55rem 1.25rem;
  font-size: 0.85rem;
  font-weight: 700;
  cursor: pointer;
  color: var(--a-color-muted);
  transition: background 0.12s;
  position: relative;
}

.sidebar-item:hover {
  background: var(--a-color-disabled-bg);
}

.sidebar-item-active {
  background: var(--a-color-fg);
  color: var(--a-color-bg);
}

.sidebar-item-active:hover {
  background: var(--a-color-fg);
}

.sidebar-sub {
  font-size: 0.8rem;
  font-weight: 600;
  color: var(--a-color-muted);
  padding-left: 1.75rem;
}

.sidebar-item-icon {
  font-size: 0.75rem;
  width: 1rem;
  text-align: center;
  flex-shrink: 0;
}

.sidebar-item-text {
  flex: 1;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.sidebar-item-count {
  font-size: 0.65rem;
  font-weight: 900;
  color: var(--a-color-muted-soft);
  background: var(--a-color-disabled-bg);
  padding: 0.1rem 0.4rem;
  border-radius: 2px;
}

.sidebar-item-active .sidebar-item-count {
  background: rgba(255, 255, 255, 0.2);
  color: var(--a-color-bg);
}

.sidebar-cat-dot {
  width: 8px;
  height: 8px;
  border-radius: 50%;
  flex-shrink: 0;
}

/* Tags in sidebar */
.sidebar-tag {
  display: inline-block;
  font-size: 0.7rem;
  font-weight: var(--a-font-weight-strong);
  padding: 0.2rem 0.6rem;
  margin: 0.15rem 0.25rem;
  margin-left: 1.25rem;
  border: var(--a-border);
  border-color: var(--a-color-disabled-border);
  color: var(--a-color-muted);
  cursor: pointer;
  transition: all 0.12s;
}

.sidebar-tag:hover,
.sidebar-tag-active {
  border-color: var(--a-color-fg);
  color: var(--a-color-bg);
  background: var(--a-color-fg);
}

/* Shortcuts */
.sidebar-shortcuts {
  margin-top: 1.5rem;
  padding: 0.75rem 1.25rem;
  border-top: 1px solid var(--a-color-disabled-border);
  font-size: 0.65rem;
  font-weight: 700;
  text-transform: uppercase;
  letter-spacing: 0.06em;
  color: var(--a-color-muted-soft);
  line-height: 2.2;
}

kbd {
  display: inline-block;
  padding: 0.05em 0.3em;
  border: 1.5px solid var(--a-color-disabled-border);
  font-size: 0.6rem;
  font-weight: 900;
  font-family: monospace;
  background: var(--a-color-surface);
  color: var(--a-color-muted);
  margin-right: 0.2em;
}

/* ── Main content ────────────────────────────────────────────────────────── */
.forum-main {
  flex: 1;
  min-width: 0;
  padding: 0 0 4rem 0;
}

/* ── Tab bar ─────────────────────────────────────────────────────────────── */
.tab-bar {
  display: flex;
  align-items: center;
  justify-content: space-between;
  border-bottom: var(--a-border);
  padding: 0 1.5rem;
  position: sticky;
  top: 0;
  background: var(--a-color-bg);
  z-index: 10;
}

.tab-bar-left {
  display: flex;
  align-items: center;
  gap: 0.5rem;
  flex-wrap: wrap;
}

.tab-bar-right {
  display: flex;
  align-items: center;
  gap: 0.75rem;
}

.forum-tab-btn {
  box-shadow: none;
}

.forum-tab-btn-active {
  background: var(--a-color-fg);
  color: var(--a-color-bg);
  border-color: var(--a-color-fg);
}

/* ── Filter bar ──────────────────────────────────────────────────────────── */
.filter-bar {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 0.75rem 1.5rem;
  border-bottom: 1px solid var(--a-color-disabled-border);
  gap: 1rem;
}

.filter-left {
  display: flex;
  align-items: center;
  gap: 0.75rem;
}

.filter-right {
  display: flex;
  align-items: center;
}

.forum-category-select {
  min-width: 12rem;
}

.forum-category-select :deep(.a-field) {
  gap: 0;
}

.forum-category-select :deep(.a-select-trigger) {
  min-height: 2.5rem;
  font-size: 0.8rem;
  font-weight: var(--a-font-weight-black);
  letter-spacing: var(--a-letter-spacing-wide);
}

/* Active tag chip */
.active-tag-chip {
  display: inline-flex;
  align-items: center;
  gap: 0.3rem;
}

.tag-chip-close {
  background: none;
  border: none;
  color: var(--a-color-bg);
  cursor: pointer;
  font-size: 1rem;
  line-height: 1;
  padding: 0;
  opacity: 0.7;
}

.tag-chip-close:hover {
  opacity: 1;
}

/* Search */
.search-wrap {
  position: relative;
}

.search-wrap :deep(.a-field) {
  gap: 0;
}

.search-input {
  width: 200px;
}

.search-input :deep(.a-input) {
  padding-right: 2rem;
  font-size: 0.8rem;
}

.search-clear {
  position: absolute;
  right: 0.6rem;
  top: 50%;
  transform: translateY(-50%);
  background: none;
  border: none;
  font-size: 1.1rem;
  cursor: pointer;
  color: var(--a-color-muted-soft);
  padding: 0;
  line-height: 1;
}

/* ── Topic list header ───────────────────────────────────────────────────── */
.topic-list-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 0.5rem 1.5rem;
  border-bottom: var(--a-border);
  background: var(--a-color-surface);
}

.th-title {
  color: var(--a-color-muted-soft);
  flex: 1;
}

.th-stats {
  display: flex;
  gap: 0;
  color: var(--a-color-muted-soft);
}

.th-stats span {
  width: 60px;
  text-align: right;
}

/* ── Topic rows ──────────────────────────────────────────────────────────── */
.topic-list {
  display: flex;
  flex-direction: column;
  gap: 0.75rem;
}

.topic-row-skeleton {
  height: 68px;
}

.topic-row {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 0.875rem 1.25rem;
  cursor: pointer;
  transition: background 0.1s;
  gap: 1rem;
  margin: 0;
}

.topic-row:hover,
.topic-row-focused {
  background: var(--a-color-surface);
}

.topic-row-pinned {
  background: var(--a-color-surface);
}

/* Left side */
.tr-left {
  flex: 1;
  min-width: 0;
}

.tr-tags {
  display: flex;
  flex-wrap: wrap;
  align-items: center;
  gap: 0.3rem;
  margin-bottom: 0.3rem;
}

.tr-badge {
  font-size: 0.6rem;
  font-weight: var(--a-font-weight-black);
}

.tr-badge-pin {
  color: var(--a-color-fg);
  text-transform: uppercase;
}

.tr-badge-closed {
  border-color: var(--a-color-disabled-border);
  color: var(--a-color-muted-soft);
  text-transform: uppercase;
}

.tr-badge-cat {
  cursor: pointer;
  text-transform: uppercase;
  transition: all 0.1s;
}

.tr-badge-cat:hover {
  background: currentColor;
  color: var(--a-color-bg);
}

.tr-badge-tag {
  border-color: var(--a-color-disabled-border);
  color: var(--a-color-muted);
  cursor: pointer;
  transition: all 0.1s;
}

.tr-badge-tag:hover {
  border-color: var(--a-color-fg);
  color: var(--a-color-fg);
}

.tr-title {
  font-size: 0.92rem;
  font-weight: 700;
  margin: 0 0 0.25rem;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
  color: var(--a-color-fg);
  transition: color 0.1s;
}

.topic-row:hover .tr-title,
.topic-row-focused .tr-title {
  text-decoration: underline;
}

.tr-meta {
  font-size: 0.72rem;
  font-weight: 500;
  color: var(--a-color-muted-soft);
  margin: 0;
  display: flex;
  align-items: center;
  gap: 0.3rem;
}

.tr-author {
  font-weight: 700;
  color: var(--a-color-muted);
}

.tr-sep {
  color: var(--a-color-disabled-border);
}

.tr-bookmarked {
  font-weight: 900;
  color: var(--a-color-fg);
  font-size: 0.6rem;
  text-transform: uppercase;
  letter-spacing: 0.06em;
}

/* Right side */
.tr-right {
  display: flex;
  align-items: center;
  gap: 1.25rem;
  flex-shrink: 0;
}

/* Avatars */
.tr-avatars {
  display: flex;
  align-items: center;
}

.tr-avatar {
  width: 28px;
  height: 28px;
  border: var(--a-border);
  background: var(--a-color-disabled-bg);
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 0.65rem;
  font-weight: 900;
  color: var(--a-color-muted);
}

.tr-avatar-op {
  background: var(--a-color-fg);
  color: var(--a-color-bg);
}

/* Stats */
.tr-stats {
  display: flex;
  align-items: center;
}

.tr-stat {
  width: 60px;
  text-align: right;
  font-size: 0.75rem;
  font-weight: 700;
  color: var(--a-color-muted);
}

.tr-stat-val {
  font-weight: 900;
  color: var(--a-color-muted);
}

.tr-stat-time {
  color: var(--a-color-muted-soft);
  font-weight: 600;
  font-size: 0.7rem;
}

/* Load more */
.load-more-wrap {
  padding: 1.5rem;
  text-align: center;
}

/* ── Responsive ──────────────────────────────────────────────────────────── */
@media (max-width: 1024px) {
  .forum-sidebar {
    width: 180px;
  }
}

@media (max-width: 768px) {
  .forum-page {
    display: block;
    padding: 0;
  }

  .forum-sidebar {
    display: none;
  }

  .forum-main {
    padding-bottom: 4rem;
  }

  .tab-bar {
    padding: 0 0.75rem;
    overflow-x: auto;
  }

  .tab-bar-left {
    flex-wrap: nowrap;
  }

  .filter-bar {
    padding: 0.625rem 0.75rem;
    flex-wrap: wrap;
    gap: 0.5rem;
  }

  .search-input {
    width: 140px;
  }

  .topic-row {
    padding: 0.75rem;
  }

  .topic-list-header {
    padding: 0.5rem 0.75rem;
  }

  .tr-right {
    gap: 0.75rem;
  }

  .tr-stat {
    width: 46px;
    font-size: 0.7rem;
  }

  .th-stats span {
    width: 46px;
  }
}
</style>
