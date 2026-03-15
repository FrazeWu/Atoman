<template>
  <div class="a-page-xl" style="padding-bottom:12rem">
    <APageHeader title="订阅" accent sub="聚合你感兴趣的 RSS 订阅源" style="margin-bottom:2.5rem">
      <template #action>
        <ABtn v-if="authStore.isAuthenticated" @click="showAddModal = true">+ 添加订阅</ABtn>
      </template>
    </APageHeader>

    <div v-if="!authStore.isAuthenticated" style="min-height:50vh;display:flex;flex-direction:column;align-items:center;justify-content:center;text-align:center">
      <p class="a-title-xl a-muted" style="margin-bottom:1.5rem">订阅</p>
      <p class="a-muted" style="max-width:28rem;margin-bottom:2rem">登录后即可添加 RSS 源，构建你的个性化信息流。</p>
      <ABtn to="/login" size="lg">登录</ABtn>
    </div>

    <template v-else>
      <div style="display:flex;gap:2rem">
        <!-- Left sidebar -->
        <div style="width:18rem;flex-shrink:0">
          <div style="border:2px solid #000">
            <!-- All subscriptions -->
            <button
              @click="selectAll"
              style="width:100%;text-align:left;padding:1rem 1.25rem;font-weight:900;font-size:.75rem;text-transform:uppercase;letter-spacing:.1em;border-bottom:2px solid #000;cursor:pointer;transition:all .2s;background:none;border-right:none;border-left:none;border-top:none"
              :style="isAllActive ? 'background:#000;color:#fff' : 'background:#fff;color:#000'"
            >
              全部订阅
            </button>

            <div v-if="loadingSubscriptions" style="padding:1rem">
              <div v-for="i in 4" :key="i" class="a-skeleton" style="height:3rem;margin-bottom:.5rem" />
            </div>

            <template v-else>
              <!-- Groups -->
              <template v-for="group in groups" :key="group.id">
                <div
                  style="display:flex;align-items:center;border-bottom:1px solid #e5e7eb;cursor:pointer;transition:all .2s"
                  :style="activeGroupId === group.id ? 'background:#000;color:#fff' : 'background:#f9fafb;color:#000'"
                >
                  <button
                    @click="toggleGroup(group)"
                    style="flex:1;text-align:left;padding:.75rem 1.25rem;font-weight:900;font-size:.7rem;text-transform:uppercase;letter-spacing:.08em;background:none;border:none;cursor:pointer;color:inherit"
                  >
                    {{ expandedGroups.has(group.id) ? '▾' : '▸' }} {{ group.name }}
                  </button>
                  <div style="display:flex;gap:.25rem;padding-right:.75rem">
                    <button
                      @click.stop="startRenameGroup(group)"
                      style="font-size:.65rem;font-weight:900;padding:.25rem .4rem;border:1px solid;cursor:pointer;transition:all .2s;background:none"
                      :style="activeGroupId === group.id ? 'border-color:#fff;color:#fff' : 'border-color:#9ca3af;color:#6b7280'"
                    >改名</button>
                    <button
                      @click.stop="requestRemoveGroup(group.id)"
                      style="font-size:.65rem;font-weight:900;padding:.25rem .4rem;border:1px solid;cursor:pointer;transition:all .2s;background:none"
                      :style="activeGroupId === group.id ? 'border-color:#fff;color:#fff' : 'border-color:#fca5a5;color:#ef4444'"
                    >删除</button>
                  </div>
                </div>
                <!-- Rename inline -->
                <div v-if="renamingGroupId === group.id" style="padding:.5rem .75rem;border-bottom:1px solid #e5e7eb;background:#fff;display:flex;gap:.5rem">
                  <input v-model="renameGroupName" class="a-input" style="flex:1;padding:.35rem .5rem;font-size:.8rem" @keyup.enter="confirmRenameGroup" @keyup.esc="renamingGroupId = null" />
                  <button @click="confirmRenameGroup" style="font-weight:900;font-size:.7rem;padding:.35rem .5rem;background:#000;color:#fff;border:none;cursor:pointer">确认</button>
                </div>
                <!-- Group subscriptions -->
                <template v-if="expandedGroups.has(group.id)">
                  <div
                    v-for="sub in subscriptionsInGroup(group.id)"
                    :key="sub.id"
                    style="display:flex;align-items:flex-start;justify-content:space-between;padding:.75rem 1.25rem .75rem 2rem;border-bottom:1px solid #f3f4f6;cursor:pointer;transition:all .2s"
                    :style="activeSourceId === sub.id ? 'background:#000;color:#fff' : 'background:#fff;color:#000'"
                    @click="selectSource(sub.id)"
                  >
                    <div style="flex:1;min-width:0">
                      <span class="a-label" style="display:block;margin-bottom:.15rem;color:#9ca3af;font-size:.65rem">
                        {{ sourceTypeLabel(sub.feed_source?.source_type || '') }}
                      </span>
                      <span style="font-weight:700;font-size:.8rem;display:block;overflow:hidden;text-overflow:ellipsis;white-space:nowrap">
                        {{ sub.title || sub.feed_source?.title || '未命名' }}
                      </span>
                    </div>
                    <div style="display:flex;gap:.25rem;align-items:center;margin-left:.5rem;flex-shrink:0">
                      <div style="position:relative">
                        <button
                          @click.stop="toggleGroupPopover(sub.id)"
                          style="font-size:.65rem;padding:.2rem .45rem;border:1px solid #d1d5db;background:#fff;color:#000;cursor:pointer;font-weight:700"
                          :title="'移至分组'"
                        >
                          {{ currentGroupName(sub) }} ▾
                        </button>
                        <div
                          v-if="groupPopoverSubId === sub.id"
                          class="group-popover"
                          @click.stop
                        >
                          <button
                            class="group-popover-item"
                            @click="setSubscriptionGroupAndClose(sub.id, defaultGroupId || '')"
                          >
                            默认分组
                          </button>
                          <button
                            v-for="g in nonDefaultGroups"
                            :key="g.id"
                            class="group-popover-item"
                            @click="setSubscriptionGroupAndClose(sub.id, g.id)"
                          >
                            {{ g.name }}
                          </button>
                        </div>
                      </div>
                      <span
                        @click.stop="requestUnsubscribeSource(sub.id)"
                        style="font-size:.75rem;font-weight:900;background:none;border:none;cursor:pointer;opacity:0.4;transition:opacity .2s;color:#ef4444"
                      >✕</span>
                    </div>
                  </div>
                </template>
              </template>

              <!-- Ungrouped subscriptions -->
              <div
                v-if="ungroupedSubscriptions.length && !defaultGroupId"
                style="padding:.6rem 1.25rem;font-weight:900;font-size:.65rem;text-transform:uppercase;letter-spacing:.08em;color:#9ca3af;border-bottom:1px solid #e5e7eb;background:#f9fafb"
              >
                默认分组
              </div>
              <div
                v-for="sub in ungroupedSubscriptions"
                :key="sub.id"
                style="display:flex;align-items:flex-start;justify-content:space-between;padding:.75rem 1.25rem;border-bottom:1px solid #f3f4f6;cursor:pointer;transition:all .2s"
                :style="activeSourceId === sub.id ? 'background:#000;color:#fff' : 'background:#fff;color:#000'"
                @click="selectSource(sub.id)"
              >
                <div style="flex:1;min-width:0">
                  <span class="a-label" style="display:block;margin-bottom:.15rem;color:#9ca3af;font-size:.65rem">
                    {{ sourceTypeLabel(sub.feed_source?.source_type || '') }}
                  </span>
                  <span style="font-weight:700;font-size:.8rem;display:block;overflow:hidden;text-overflow:ellipsis;white-space:nowrap">
                    {{ sub.title || sub.feed_source?.title || '未命名' }}
                  </span>
                </div>
                <div style="display:flex;gap:.25rem;align-items:center;margin-left:.5rem;flex-shrink:0">
                  <div style="position:relative">
                    <button
                      @click.stop="toggleGroupPopover(sub.id)"
                      style="font-size:.65rem;padding:.2rem .45rem;border:1px solid #d1d5db;background:#fff;color:#000;cursor:pointer;font-weight:700"
                      :title="'移至分组'"
                    >
                      {{ currentGroupName(sub) }} ▾
                    </button>
                    <div
                      v-if="groupPopoverSubId === sub.id"
                      class="group-popover"
                      @click.stop
                    >
                      <button
                        class="group-popover-item"
                        @click="setSubscriptionGroupAndClose(sub.id, defaultGroupId || '')"
                      >
                        默认分组
                      </button>
                      <button
                        v-for="g in nonDefaultGroups"
                        :key="g.id"
                        class="group-popover-item"
                        @click="setSubscriptionGroupAndClose(sub.id, g.id)"
                      >
                        {{ g.name }}
                      </button>
                    </div>
                  </div>
                  <span
                    @click.stop="requestUnsubscribeSource(sub.id)"
                    style="font-size:.75rem;font-weight:900;background:none;border:none;cursor:pointer;opacity:0.4;transition:opacity .2s;color:#ef4444"
                  >✕</span>
                </div>
              </div>

              <!-- New group button -->
              <div style="padding:.75rem 1.25rem;border-top:2px solid #000">
                <div v-if="addingGroup" style="display:flex;gap:.5rem">
                  <input
                    v-model="newGroupName"
                    class="a-input"
                    placeholder="分组名称"
                    style="flex:1;padding:.35rem .5rem;font-size:.8rem"
                    @keyup.enter="confirmAddGroup"
                    @keyup.esc="cancelAddGroup"
                    ref="newGroupInput"
                  />
                  <button @click="confirmAddGroup" style="font-weight:900;font-size:.7rem;padding:.35rem .5rem;background:#000;color:#fff;border:none;cursor:pointer">确认</button>
                  <button @click="cancelAddGroup" style="font-weight:900;font-size:.7rem;padding:.35rem .5rem;background:#fff;color:#000;border:1px solid #000;cursor:pointer">取消</button>
                </div>
                <button
                  v-else
                  @click="startAddGroup"
                  style="font-weight:900;font-size:.7rem;text-transform:uppercase;letter-spacing:.08em;background:none;border:none;cursor:pointer;color:#6b7280;transition:color .2s"
                >+ 新建分组</button>
              </div>
            </template>
          </div>
        </div>

        <!-- Right: Timeline -->
        <div style="flex:1;min-width:0">
          <div style="display:flex;justify-content:flex-end;margin-bottom:1rem">
            <button
              @click="markAllReadAndRefresh"
              style="font-weight:900;font-size:.7rem;text-transform:uppercase;letter-spacing:.08em;padding:.5rem 1rem;border:2px solid #000;background:#fff;cursor:pointer;transition:all .2s"
              :style="markingAllRead ? 'background:#000;color:#fff' : ''"
            >
              {{ markingAllRead ? '标记中...' : '全部已读' }}
            </button>
          </div>

          <div v-if="loadingTimeline" style="display:flex;flex-direction:column;gap:1rem">
            <div v-for="i in 5" :key="i" class="a-skeleton" style="height:7rem" />
          </div>

          <AEmpty
            v-else-if="!timeline.length"
            :text="subscriptions.length ? '订阅源暂无更新' : '订阅后开始探索'"
          />

          <div v-else style="display:flex;flex-direction:column;gap:1rem">
            <template v-for="item in timeline" :key="itemKey(item)">
              <!-- Internal Post -->
              <RouterLink
                v-if="item.type === 'post' && item.post"
                :to="`/blog/posts/${item.post.id}`"
                class="a-card a-card-hover"
                style="display:block;text-decoration:none;color:#000;transition:all .3s;position:relative"
                @click="onItemClick(item)"
              >
                <div style="display:flex;gap:1rem;align-items:flex-start">
                  <span class="a-badge-fill" style="flex-shrink:0">博客</span>
                  <div style="flex:1;min-width:0">
                    <div style="display:flex;align-items:center;gap:.75rem;margin-bottom:.5rem;flex-wrap:wrap">
                      <span class="a-label a-muted">{{ item.post.user?.display_name || item.post.user?.username || '未知作者' }}</span>
                      <span style="font-size:.75rem;color:#d1d5db">{{ formatDate(item.published_at) }}</span>
                      <span class="a-label" style="margin-left:auto;color:#d1d5db">内部文章 →</span>
                    </div>
                    <h3 class="a-clamp-2" style="font-weight:900;font-size:1.125rem;letter-spacing:-0.025em;margin-bottom:.5rem">
                      {{ item.post.title }}
                    </h3>
                    <p v-if="item.post.summary" class="a-muted a-clamp-2" style="font-size:.875rem">
                      {{ item.post.summary }}
                    </p>
                  </div>
                </div>
                <span
                  v-if="item.is_read"
                  class="a-label"
                  style="position:absolute;right:1rem;bottom:1rem;color:#2bb24c;border:1px solid #2bb24c;padding:.1rem .4rem;background:#f0fff4"
                >已读</span>
              </RouterLink>

              <!-- External RSS Item -->
              <a
                v-else-if="item.type === 'feed_item' && item.feed_item"
                :href="item.feed_item.link"
                target="_blank"
                rel="noopener noreferrer"
                class="a-card a-card-hover"
                style="display:block;text-decoration:none;color:#000;transition:all .3s;position:relative"
                @click="onItemClick(item)"
              >
                <div style="display:flex;gap:1rem;align-items:flex-start">
                  <!-- Podcast thumbnail -->
                  <img
                    v-if="item.feed_item.image_url"
                    :src="item.feed_item.image_url"
                    style="width:4rem;height:4rem;object-fit:cover;border:2px solid #000;filter:grayscale(100%);flex-shrink:0"
                  />
                  <span v-else class="a-badge" style="flex-shrink:0">外部</span>

                  <div style="flex:1;min-width:0">
                    <div style="display:flex;align-items:center;gap:.75rem;margin-bottom:.5rem;flex-wrap:wrap">
                      <span class="a-label a-muted">{{ item.feed_item.author || item.feed_item.feed_source?.title || 'RSS' }}</span>
                      <span v-if="item.feed_item.duration" style="font-size:.7rem;color:#9ca3af;font-weight:700">
                        时长: {{ item.feed_item.duration }}
                      </span>
                      <span style="font-size:.75rem;color:#d1d5db">{{ formatDate(item.feed_item.published_at) }}</span>
                      <span class="a-label" style="margin-left:auto;color:#d1d5db">↗ 外部链接</span>
                    </div>
                    <h3 class="a-clamp-2" style="font-weight:900;font-size:1.125rem;letter-spacing:-0.025em;margin-bottom:.5rem">
                      {{ item.feed_item.title }}
                    </h3>
                    <p v-if="item.feed_item.summary" class="a-muted a-clamp-2" style="font-size:.875rem">
                      {{ stripHtml(item.feed_item.summary) }}
                    </p>
                    <!-- Podcast play button -->
                    <div
                      v-if="item.feed_item.enclosure_url && item.feed_item.enclosure_type?.startsWith('audio/')"
                      style="margin-top:.75rem"
                      @click.prevent.stop="playPodcast(item.feed_item, $event)"
                    >
                      <button
                        style="font-weight:900;font-size:.7rem;text-transform:uppercase;letter-spacing:.08em;padding:.4rem 1rem;border:2px solid #000;cursor:pointer;transition:all .2s"
                        :style="isPodcastPlaying(item.feed_item) ? 'background:#000;color:#fff' : 'background:#fff;color:#000'"
                      >
                        {{ isPodcastPlaying(item.feed_item) ? '■ 播放中' : '▶ 播放' }}
                      </button>
                    </div>
                  </div>
                </div>
                <span
                  v-if="item.is_read"
                  class="a-label"
                  style="position:absolute;right:1rem;bottom:1rem;color:#2bb24c;border:1px solid #2bb24c;padding:.1rem .4rem;background:#f0fff4"
                >已读</span>
              </a>
            </template>

            <!-- Pagination -->
            <div style="display:flex;flex-direction:column;align-items:center;gap:.75rem;padding:1.5rem 0">
              <p style="font-size:.75rem;color:#9ca3af;font-weight:700">
                已加载 {{ timeline.length }} / {{ totalItems }} 条
              </p>
              <button
                v-if="timeline.length < totalItems"
                @click="loadMore"
                :disabled="loadingMore"
                style="font-weight:900;font-size:.75rem;text-transform:uppercase;letter-spacing:.08em;padding:.75rem 2rem;border:2px solid #000;cursor:pointer;transition:all .2s;background:#fff"
                :style="loadingMore ? 'opacity:.5;cursor:not-allowed' : 'hover:background:#000'"
              >
                {{ loadingMore ? '加载中...' : '加载更多' }}
              </button>
            </div>
          </div>
        </div>
      </div>
    </template>

    <!-- Add Subscription Modal -->
    <AModal v-if="showAddModal" @close="showAddModal = false" size="md">
      <h3 class="a-subtitle" style="margin-bottom:1.5rem">添加订阅</h3>
      <div style="display:flex;flex-direction:column;gap:1rem;margin-bottom:1rem">
        <div class="a-field">
          <label class="a-field-label">RSS 地址 *</label>
          <input v-model="newRssUrl" placeholder="https://example.com/feed.xml" class="a-input" />
        </div>
        <div class="a-field">
          <label class="a-field-label">自定义名称（可选）</label>
          <input v-model="newRssTitle" placeholder="例如：GitHub Blog" class="a-input" />
        </div>
        <div v-if="groups.length" class="a-field">
          <label class="a-field-label">添加到分组（可选）</label>
          <select v-model="newRssGroupId" class="a-input" style="cursor:pointer">
            <option :value="defaultGroupId || ''">默认分组</option>
            <option v-for="g in nonDefaultGroups" :key="g.id" :value="g.id">{{ g.name }}</option>
          </select>
        </div>
      </div>
      <div v-if="addError" class="a-error" style="margin-bottom:1rem">{{ addError }}</div>
      <div style="display:flex;gap:.75rem">
        <ABtn style="flex:1" @click="addSubscription" :loading="adding" loadingText="添加中...">添加</ABtn>
        <ABtn outline @click="showAddModal = false">取消</ABtn>
      </div>
    </AModal>

    <AConfirm
      :show="showDeleteConfirm"
      :title="deleteConfirmTitle"
      :message="deleteConfirmMessage"
      confirm-text="删除"
      cancel-text="取消"
      danger
      @confirm="confirmDeleteAction"
      @cancel="cancelDeleteAction"
    />
  </div>
</template>

<script setup lang="ts">
import { ref, computed, watch, onMounted, onUnmounted, nextTick } from 'vue'
import { RouterLink } from 'vue-router'
import ABtn from '@/components/ui/ABtn.vue'
import AModal from '@/components/ui/AModal.vue'
import AEmpty from '@/components/ui/AEmpty.vue'
import APageHeader from '@/components/ui/APageHeader.vue'
import AConfirm from '@/components/ui/AConfirm.vue'
import { useAuthStore } from '@/stores/auth'
import { usePlayerStore } from '@/stores/player'
import type { Subscription, SubscriptionGroup, TimelineItem, FeedItem } from '@/types'

const authStore = useAuthStore()
const playerStore = usePlayerStore()
const API_URL = import.meta.env.VITE_API_URL || '/api'
const authHeaders = () => ({ Authorization: `Bearer ${authStore.token}` })

const playPodcast = (feedItem: FeedItem, event: Event) => {
  event.preventDefault()
  event.stopPropagation()
  // Mark as read
  const timelineItem = timeline.value.find(
    (t) => t.type === 'feed_item' && t.feed_item?.id === feedItem.id
  )
  if (timelineItem && !timelineItem.is_read) {
    timelineItem.is_read = true
    fetch(`${API_URL}/feed/timeline/mark-read`, {
      method: 'POST',
      headers: { 'Content-Type': 'application/json', ...authHeaders() },
      body: JSON.stringify({ feed_item_ids: [feedItem.id] }),
    }).catch(console.error)
  }
  // Build a temporary Song object compatible with playerStore
  const tempSong = {
    id: -1,
    title: feedItem.title || '未知播客',
    artist: feedItem.author || feedItem.feed_source?.title || 'Podcast',
    album: feedItem.feed_source?.title || 'Podcast',
    album_id: -1,
    year: new Date(feedItem.published_at || '').getFullYear() || 0,
    release_date: feedItem.published_at || '',
    lyrics: feedItem.summary || '',
    audio_url: feedItem.enclosure_url || '',
    cover_url: feedItem.image_url || '',
    status: 'approved' as const,
  }
  playerStore.playSong(tempSong)
}

const isPodcastPlaying = (feedItem: FeedItem) =>
  playerStore.currentSong?.audio_url === feedItem.enclosure_url && playerStore.isPlaying

// State
const subscriptions = ref<Subscription[]>([])
const groups = ref<SubscriptionGroup[]>([])
const timeline = ref<TimelineItem[]>([])
const totalItems = ref(0)
const currentPage = ref(1)
const pageLimit = 20

const activeSourceId = ref<number | null>(null)
const activeGroupId = ref<string | null>(null)
const expandedGroups = ref<Set<string>>(new Set())
const groupPopoverSubId = ref<number | null>(null)

const loadingSubscriptions = ref(true)
const loadingTimeline = ref(false)
const loadingMore = ref(false)
const markingAllRead = ref(false)

// Group management
const addingGroup = ref(false)
const newGroupName = ref('')
const newGroupInput = ref<HTMLInputElement | null>(null)
const renamingGroupId = ref<string | null>(null)
const renameGroupName = ref('')

// Modal state
const showAddModal = ref(false)
const newRssUrl = ref('')
const newRssTitle = ref('')
const newRssGroupId = ref('')
const addError = ref('')
const adding = ref(false)

const showDeleteConfirm = ref(false)
const deleteConfirmTitle = ref('请确认删除')
const deleteConfirmMessage = ref('该操作不可撤销，是否继续？')
let pendingDeleteAction: (() => Promise<void> | void) | null = null

// Computed
const isAllActive = computed(() => activeSourceId.value === null && activeGroupId.value === null)

const defaultGroupId = computed(() => groups.value.find((g) => g.name === '默认分组')?.id || '')

const nonDefaultGroups = computed(() =>
  groups.value.filter((g) => g.name !== '默认分组')
)

const subscriptionsInGroup = (groupId: string) => {
  if (groupId === defaultGroupId.value) {
    return subscriptions.value.filter((s) => !s.subscription_group_id || s.subscription_group_id === groupId)
  }
  return subscriptions.value.filter((s) => s.subscription_group_id === groupId)
}

const ungroupedSubscriptions = computed(() =>
  subscriptions.value.filter((s) => !s.subscription_group_id)
)

const currentGroupName = (sub: Subscription) => {
  const groupId = sub.subscription_group_id || defaultGroupId.value
  if (!groupId) return '默认分组'
  return groups.value.find((g) => g.id === groupId)?.name || '默认分组'
}

// Helpers
const itemKey = (item: TimelineItem) => {
  if (item.type === 'post' && item.post) return `post-${item.post.id}`
  if (item.type === 'feed_item' && item.feed_item) return `feed-${item.feed_item.id}`
  return Math.random().toString()
}

const sourceTypeLabel = (type: string) => {
  if (type === 'external_rss') return 'RSS'
  if (type === 'internal_user') return '用户'
  if (type === 'internal_channel') return '频道'
  if (type === 'internal_collection') return '合集'
  return type
}

const formatDate = (d?: string) => {
  if (!d) return ''
  return new Date(d).toLocaleDateString('zh-CN', { month: 'short', day: 'numeric' })
}

const stripHtml = (html: string) =>
  html
    .replace(/<[^>]*>/g, '')
    .replace(/&amp;/g, '&')
    .replace(/&lt;/g, '<')
    .replace(/&gt;/g, '>')
    .replace(/&quot;/g, '"')
    .trim()

// Fetch
const fetchSubscriptions = async () => {
  if (!authStore.isAuthenticated) return
  loadingSubscriptions.value = true
  try {
    const res = await fetch(`${API_URL}/feed/subscriptions`, { headers: authHeaders() })
    if (res.ok) {
      const d = await res.json()
      subscriptions.value = d.data || []
    }
  } finally {
    loadingSubscriptions.value = false
  }
}

const fetchGroups = async () => {
  if (!authStore.isAuthenticated) return
  try {
    const res = await fetch(`${API_URL}/feed/groups`, { headers: authHeaders() })
    if (res.ok) {
      const d = await res.json()
      groups.value = d.data || []
      if (!newRssGroupId.value && defaultGroupId.value) {
        newRssGroupId.value = defaultGroupId.value
      }
    }
  } catch (e) {
    console.error(e)
  }
}

const fetchTimeline = async (append = false) => {
  if (!authStore.isAuthenticated) return
  if (append) loadingMore.value = true
  else loadingTimeline.value = true

  try {
    const params = new URLSearchParams({ page: String(currentPage.value), limit: String(pageLimit) })
    if (activeSourceId.value !== null) params.set('source_id', String(activeSourceId.value))
    if (activeGroupId.value !== null) params.set('group_id', activeGroupId.value)

    const res = await fetch(`${API_URL}/feed/timeline?${params}`, { headers: authHeaders() })
    if (res.ok) {
      const d = await res.json()
      const items: TimelineItem[] = d.data || []
      if (append) {
        timeline.value = [...timeline.value, ...items]
      } else {
        timeline.value = items
      }
      totalItems.value = d.total || 0
    }
  } catch (e) {
    console.error(e)
  } finally {
    loadingTimeline.value = false
    loadingMore.value = false
  }
}

const loadMore = async () => {
  currentPage.value++
  await fetchTimeline(true)
}

// Selection
const selectAll = () => {
  activeSourceId.value = null
  activeGroupId.value = null
  currentPage.value = 1
  fetchTimeline()
}

const selectSource = (id: number) => {
  activeSourceId.value = id
  activeGroupId.value = null
  currentPage.value = 1
  fetchTimeline()
}

const toggleGroup = (group: SubscriptionGroup) => {
  if (expandedGroups.value.has(group.id)) {
    expandedGroups.value.delete(group.id)
    const newSet = new Set(expandedGroups.value)
    expandedGroups.value = newSet
  } else {
    const newSet = new Set(expandedGroups.value)
    newSet.add(group.id)
    expandedGroups.value = newSet
    activeGroupId.value = group.id
    activeSourceId.value = null
    currentPage.value = 1
    fetchTimeline()
  }
}

// Read state
const onItemClick = (item: TimelineItem) => {
  if (item.type === 'feed_item' && item.feed_item && !item.is_read) {
    item.is_read = true
    fetch(`${API_URL}/feed/timeline/mark-read`, {
      method: 'POST',
      headers: { 'Content-Type': 'application/json', ...authHeaders() },
      body: JSON.stringify({ feed_item_ids: [item.feed_item.id] }),
    }).catch(console.error)
  }
}

const markAllReadAndRefresh = async () => {
  markingAllRead.value = true
  try {
    await fetch(`${API_URL}/feed/timeline/mark-all-read`, {
      method: 'POST',
      headers: authHeaders(),
    })
    timeline.value.forEach((item) => {
      if (item.type === 'feed_item') item.is_read = true
    })
  } finally {
    markingAllRead.value = false
  }
}

// Unsubscribe
const unsubscribeSource = async (id: number) => {
  try {
    await fetch(`${API_URL}/feed/subscriptions/${id}`, { method: 'DELETE', headers: authHeaders() })
    if (activeSourceId.value === id) activeSourceId.value = null
    await fetchSubscriptions()
    await fetchTimeline()
  } catch (e) {
    console.error(e)
  }
}

const requestDeleteAction = (title: string, message: string, action: () => Promise<void> | void) => {
  deleteConfirmTitle.value = title
  deleteConfirmMessage.value = message
  pendingDeleteAction = action
  showDeleteConfirm.value = true
}

const cancelDeleteAction = () => {
  showDeleteConfirm.value = false
  pendingDeleteAction = null
}

const confirmDeleteAction = async () => {
  const action = pendingDeleteAction
  cancelDeleteAction()
  if (action) {
    await action()
  }
}

const requestUnsubscribeSource = (id: number) => {
  requestDeleteAction('取消订阅', '确定要取消这个订阅吗？', () => unsubscribeSource(id))
}

// Group management
const startAddGroup = async () => {
  addingGroup.value = true
  await nextTick()
  newGroupInput.value?.focus()
}

const cancelAddGroup = () => {
  addingGroup.value = false
  newGroupName.value = ''
}

const confirmAddGroup = async () => {
  const name = newGroupName.value.trim()
  if (!name) return
  try {
    const res = await fetch(`${API_URL}/feed/groups`, {
      method: 'POST',
      headers: { 'Content-Type': 'application/json', ...authHeaders() },
      body: JSON.stringify({ name }),
    })
    if (res.ok) {
      await fetchGroups()
      addingGroup.value = false
      newGroupName.value = ''
    }
  } catch (e) {
    console.error(e)
  }
}

const startRenameGroup = (group: SubscriptionGroup) => {
  renamingGroupId.value = group.id
  renameGroupName.value = group.name
}

const confirmRenameGroup = async () => {
  if (!renamingGroupId.value) return
  const name = renameGroupName.value.trim()
  if (!name) return
  try {
    await fetch(`${API_URL}/feed/groups/${renamingGroupId.value}`, {
      method: 'PUT',
      headers: { 'Content-Type': 'application/json', ...authHeaders() },
      body: JSON.stringify({ name }),
    })
    await fetchGroups()
    renamingGroupId.value = null
  } catch (e) {
    console.error(e)
  }
}

const removeGroup = async (id: string) => {
  try {
    await fetch(`${API_URL}/feed/groups/${id}`, { method: 'DELETE', headers: authHeaders() })
    if (activeGroupId.value === id) activeGroupId.value = null
    expandedGroups.value.delete(id)
    expandedGroups.value = new Set(expandedGroups.value)
    await fetchGroups()
    await fetchSubscriptions()
    await fetchTimeline()
  } catch (e) {
    console.error(e)
  }
}

const requestRemoveGroup = (id: string) => {
  requestDeleteAction(
    '删除分组',
    '删除分组后，该分组下订阅会移回默认分组。确定删除吗？',
    () => removeGroup(id)
  )
}

const moveToGroup = async (subId: number, groupId: string) => {
  try {
    await fetch(`${API_URL}/feed/subscriptions/${subId}/group`, {
      method: 'PUT',
      headers: { 'Content-Type': 'application/json', ...authHeaders() },
      body: JSON.stringify({ group_id: groupId || null }),
    })
    await fetchSubscriptions()
  } catch (e) {
    console.error(e)
  }
}

const setSubscriptionGroupAndClose = async (subId: number, groupId: string) => {
  await moveToGroup(subId, groupId)
  groupPopoverSubId.value = null
}

const toggleGroupPopover = (subId: number) => {
  groupPopoverSubId.value = groupPopoverSubId.value === subId ? null : subId
}

const closeGroupPopover = () => {
  groupPopoverSubId.value = null
}

const onDocumentMousedown = () => {
  closeGroupPopover()
}

// Add subscription
const addSubscription = async () => {
  addError.value = ''
  adding.value = true
  try {
    if (!newRssUrl.value.trim()) {
      addError.value = 'RSS 地址不能为空'
      return
    }
    const res = await fetch(`${API_URL}/feed/subscriptions`, {
      method: 'POST',
      headers: { 'Content-Type': 'application/json', ...authHeaders() },
      body: JSON.stringify({
        target_type: 'external_rss',
        rss_url: newRssUrl.value.trim(),
        title: newRssTitle.value.trim(),
      }),
    })
    if (res.ok) {
      const d = await res.json()
      const newSub = d.data
      if (newRssGroupId.value && newSub?.id) {
        await fetch(`${API_URL}/feed/subscriptions/${newSub.id}/group`, {
          method: 'PUT',
          headers: { 'Content-Type': 'application/json', ...authHeaders() },
          body: JSON.stringify({ group_id: newRssGroupId.value }),
        })
      }
      showAddModal.value = false
      newRssUrl.value = ''
      newRssTitle.value = ''
      newRssGroupId.value = ''
      await fetchSubscriptions()
      await fetchTimeline()
    } else {
      const err = await res.json()
      addError.value = err.error || '添加失败'
    }
  } catch (e) {
    addError.value = '网络错误，请重试'
  } finally {
    adding.value = false
  }
}

watch(showAddModal, (v) => { if (!v) addError.value = '' })

onMounted(async () => {
  document.addEventListener('mousedown', onDocumentMousedown)
  if (authStore.isAuthenticated) {
    await Promise.all([fetchSubscriptions(), fetchGroups()])
    await fetchTimeline()
  }
})

onUnmounted(() => {
  document.removeEventListener('mousedown', onDocumentMousedown)
})
</script>

<style scoped>
.group-popover {
  position: absolute;
  right: 0;
  top: calc(100% + 0.25rem);
  min-width: 9rem;
  max-height: 14rem;
  overflow-y: auto;
  background: #fff;
  border: 2px solid #000;
  box-shadow: 8px 8px 0 0 rgba(0, 0, 0, 1);
  z-index: 40;
}

.group-popover-item {
  width: 100%;
  border: none;
  border-bottom: 1px solid #e5e7eb;
  background: #fff;
  color: #000;
  text-align: left;
  font-size: 0.7rem;
  font-weight: 900;
  text-transform: uppercase;
  letter-spacing: 0.06em;
  padding: 0.45rem 0.6rem;
  cursor: pointer;
}

.group-popover-item:last-child {
  border-bottom: none;
}

.group-popover-item:hover {
  background: #000;
  color: #fff;
}
</style>
