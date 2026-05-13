<template>
  <div class="a-page-xl tl-page">
    <APageHeader title="时间线" accent sub="泳道比较与地图查看">
      <template #action>
        <div class="tl-page-actions">
          <ABtn v-if="authStore.isAuthenticated" outline @click="showPersonForm = true">新建人物</ABtn>
          <ABtn v-if="authStore.isAuthenticated" @click="openCreate">新建事件</ABtn>
        </div>
      </template>
    </APageHeader>

    <div class="tl-toolbar">
      <div class="tl-filter-group">
        <label class="filter-label">起始年份</label>
        <input v-model.number="yearStart" type="number" placeholder="如 1800" class="a-input" style="width:120px" />
      </div>
      <div class="tl-filter-group">
        <label class="filter-label">结束年份</label>
        <input v-model.number="yearEnd" type="number" placeholder="如 2000" class="a-input" style="width:120px" />
      </div>
      <div class="tl-filter-group">
        <label class="filter-label">分类</label>
        <input v-model="filterCategory" type="text" placeholder="政治 / 文化 / 科技…" class="a-input" style="width:160px" />
      </div>
      <ABtn outline @click="applyFilter">筛选</ABtn>
      <ABtn outline @click="resetFilter">重置</ABtn>

      <div v-if="batchSelectedIds.length" class="tl-toolbar-batch">
        <span class="tl-toolbar-batch-count">已勾选 {{ batchSelectedIds.length }} 条</span>
        <button class="tl-action-btn" @click="addBatchToCompare">加入对比池</button>
        <button class="tl-action-btn tl-action-btn-secondary" @click="clearBatchSelection">清空勾选</button>
      </div>

      <div class="tl-mode-switch" style="margin-left:auto">
        <button
          class="tl-mode-btn"
          :class="{ 'tl-mode-btn-active': viewMode === 'lanes' }"
          @click="viewMode = 'lanes'"
        >泳道浏览</button>
        <button
          class="tl-mode-btn"
          :class="{ 'tl-mode-btn-active': viewMode === 'map' }"
          @click="viewMode = 'map'"
        >地图在线查看</button>
      </div>
    </div>

    <div v-if="loading && events.length === 0 && compareIds.length === 0" class="tl-state-block">
      <p class="font-bold">加载中...</p>
    </div>

    <AEmpty v-else-if="!loading && events.length === 0 && compareIds.length === 0" text="暂无历史事件" />

    <div v-else class="tl-shell">
      <aside class="tl-source-panel">
        <section class="a-card tl-panel-card">
          <div class="tl-panel-head">
            <div>
              <div class="tl-panel-kicker">COMPARE POOL</div>
              <h2 class="tl-panel-title">对比池</h2>
            </div>
            <span class="tl-panel-count">{{ compareEvents.length }} 条</span>
          </div>

          <p class="tl-panel-note">自由组合任意事件。泳道会按开始时间和持续时间并排展示。</p>

          <div v-if="compareIds.length && hydratingCompare && !compareEvents.length" class="tl-panel-empty">
            正在从 URL 恢复对比事件...
          </div>
          <div v-else-if="compareEvents.length" class="tl-compare-list">
            <article
              v-for="(event, index) in compareEvents"
              :key="event.id"
              class="tl-compare-item"
              :class="{ 'tl-compare-item-active': activeCompareId === event.id }"
              @click="setActiveCompare(event.id)"
            >
              <div class="tl-compare-order">{{ index + 1 }}</div>
              <div class="tl-compare-body">
                <div class="tl-compare-title-row">
                  <h3 class="tl-compare-title">{{ event.title }}</h3>
                  <span class="tl-inline-chip">{{ getDurationLabel(event) }}</span>
                </div>
                <div class="tl-compare-range">{{ formatEventRange(event) }}</div>
              </div>
              <div class="tl-compare-actions">
                <button class="tl-mini-btn" @click.stop="openDetail(event)">详情</button>
                <button class="tl-mini-btn danger" @click.stop="removeCompareId(event.id)">移出</button>
              </div>
            </article>
          </div>
          <div v-else class="tl-panel-empty">
            还没有对比事件。你可以从下方来源列表或地图标记中添加。
          </div>

          <div v-if="compareIds.length" class="tl-panel-actions">
            <button class="tl-action-btn tl-action-btn-secondary" @click="clearComparePool">清空对比池</button>
          </div>
        </section>

        <section class="a-card tl-panel-card">
          <div class="tl-panel-head">
            <div>
              <div class="tl-panel-kicker">EVENT SOURCE</div>
              <h2 class="tl-panel-title">事件来源</h2>
            </div>
            <span class="tl-panel-count">{{ sortedEvents.length }} 条</span>
          </div>

          <p class="tl-panel-note">可逐条加入，也可先勾选多条再一次性加入对比池。</p>

          <div v-if="!sortedEvents.length" class="tl-panel-empty">
            当前筛选条件下没有可用事件。
          </div>
          <div v-else class="tl-source-list">
            <article
              v-for="event in sortedEvents"
              :key="event.id"
              class="tl-source-item"
              :class="{
                'tl-source-item-compared': isCompared(event.id),
                'tl-source-item-active': activeCompareId === event.id,
              }"
            >
              <label class="tl-source-select">
                <input type="checkbox" :checked="isBatchSelected(event.id)" @change="toggleBatchSelection(event.id)" />
                <span>待比较</span>
              </label>

              <div class="tl-source-content">
                <div class="tl-source-top">
                  <div class="tl-card-date">{{ formatDatetime(event.event_date) }}</div>
                  <div class="tl-source-badges">
                    <span v-if="isCompared(event.id)" class="tl-inline-chip">已加入</span>
                    <span v-if="canMapEvent(event)" class="tl-inline-chip">有坐标</span>
                  </div>
                </div>

                <h3 class="tl-source-title">{{ event.title }}</h3>
                <p v-if="event.description" class="tl-source-desc">{{ event.description }}</p>

                <div class="tl-source-meta">
                  <span v-if="event.category" class="tl-card-category">{{ event.category }}</span>
                  <span v-if="event.location" class="tl-card-loc">📍 {{ event.location }}</span>
                  <span class="tl-source-duration">{{ getDurationLabel(event) }}</span>
                </div>

                <div class="tl-source-actions">
                  <button class="tl-action-btn" @click="toggleCompareEvent(event)">
                    {{ isCompared(event.id) ? '移出对比' : '加入对比' }}
                  </button>
                  <button class="tl-action-btn tl-action-btn-secondary" @click="openDetail(event)">展开详情</button>
                  <button v-if="canEdit(event)" class="tl-action-btn tl-action-btn-secondary" @click="openEdit(event)">编辑</button>
                </div>
              </div>
            </article>
          </div>
        </section>
      </aside>

      <main class="tl-stage">
        <section v-show="viewMode === 'lanes'" class="tl-stage-pane">
          <div class="tl-stage-head">
            <div>
              <div class="tl-section-kicker">LANE COMPARE</div>
              <h2 class="tl-section-title">泳道浏览</h2>
            </div>
            <div class="tl-stage-meta">
              {{ compareEvents.length ? `已对比 ${compareEvents.length} 条` : '等待加入事件' }}
            </div>
          </div>

          <div v-if="compareIds.length && hydratingCompare && !compareEvents.length" class="a-card tl-stage-empty">
            <div class="tl-empty-title">正在恢复对比池</div>
            <p>URL 中的事件正在加载，稍后就会生成泳道。</p>
          </div>
          <div v-else-if="!compareEvents.length" class="a-card tl-stage-empty">
            <div class="tl-empty-title">对比池还是空的</div>
            <p>从左侧事件来源中自由加入你想比较的事件。带结束时间的事件会显示持续条，单点事件只显示发生时刻。</p>
          </div>
          <div v-else class="a-card tl-lane-board">
            <div class="tl-lane-ruler">
              <div
                v-for="tick in laneTicks"
                :key="`${tick.label}-${tick.pct}`"
                class="tl-lane-ruler-tick"
                :style="{ left: tick.pct + '%' }"
              >
                <div class="tl-lane-ruler-line" />
                <span class="tl-lane-ruler-label">{{ tick.label }}</span>
              </div>
            </div>

            <article
              v-for="event in compareEvents"
              :key="event.id"
              class="tl-lane-row"
              :class="{ 'tl-lane-row-active': activeCompareId === event.id }"
            >
              <div class="tl-lane-info">
                <div class="tl-lane-info-top">
                  <h3 class="tl-lane-title">{{ event.title }}</h3>
                  <span class="tl-inline-chip">{{ getDurationLabel(event) }}</span>
                </div>

                <div class="tl-lane-range">{{ formatEventRange(event) }}</div>

                <div class="tl-lane-meta">
                  <span v-if="event.category" class="tl-card-category">{{ event.category }}</span>
                  <span v-if="event.location" class="tl-card-loc">📍 {{ event.location }}</span>
                </div>

                <div class="tl-lane-actions">
                  <button class="tl-inline-link" @click="setActiveCompare(event.id)">聚焦</button>
                  <button class="tl-inline-link" @click="openDetail(event)">详情</button>
                  <button v-if="canEdit(event)" class="tl-inline-link" @click="openEdit(event)">编辑</button>
                  <button class="tl-inline-link danger" @click="removeCompareId(event.id)">移出</button>
                </div>
              </div>

              <div class="tl-lane-track" @click="setActiveCompare(event.id)">
                <div class="tl-lane-track-base" />
                <div
                  class="tl-lane-bar-shell"
                  :class="{ 'tl-lane-bar-shell-point': isInstantEvent(event) }"
                  :style="getLaneStyle(event)"
                >
                  <div class="tl-lane-bar" :class="{ 'tl-lane-bar-active': activeCompareId === event.id }" />
                  <div class="tl-lane-bar-caption">{{ getDurationLabel(event) }}</div>
                </div>
              </div>
            </article>
          </div>
        </section>

        <section v-show="viewMode === 'map'" class="tl-stage-pane">
          <div class="tl-stage-head">
            <div>
              <div class="tl-section-kicker">MAP VIEW</div>
              <h2 class="tl-section-title">地图在线查看</h2>
            </div>
            <div class="tl-stage-meta">当前可定位 {{ mapRenderableEvents.length }} 条</div>
          </div>

          <div class="a-card tl-map-stage-card">
            <div class="tl-map-stage-note">
              <span>点击地图标记即可展开详情，并把事件加入同一对比池。</span>
              <span v-if="compareEvents.length">当前对比池 {{ compareEvents.length }} 条。</span>
            </div>

            <div class="tl-map-canvas">
              <div ref="eventMapEl" class="tl-event-map" />

              <div v-if="mapOverlayMessage" class="tl-map-overlay">
                {{ mapOverlayMessage }}
              </div>

              <div
                ref="eventPopupEl"
                class="tl-map-popup"
                :class="{ visible: mapPopupVisible }"
                @click.stop
              >
                <template v-if="mapPopupEvent">
                  <button type="button" class="tl-popup-close" aria-label="关闭地图弹窗" @click.stop="hideMapPopup">✕</button>
                  <div class="tl-popup-title">{{ mapPopupEvent.title }}</div>
                  <div class="tl-popup-date">{{ formatEventRange(mapPopupEvent) }}</div>
                  <div v-if="mapPopupEvent.location" class="tl-popup-location">{{ mapPopupEvent.location }}</div>
                  <div v-if="mapPopupEvent.category" class="tl-popup-category">{{ mapPopupEvent.category }}</div>
                  <div class="tl-popup-actions">
                    <button class="tl-action-btn" @click.stop="toggleCompareEvent(mapPopupEvent)">
                      {{ isCompared(mapPopupEvent.id) ? '移出对比' : '加入对比' }}
                    </button>
                    <button class="tl-action-btn tl-action-btn-secondary" @click.stop="openDetail(mapPopupEvent)">展开详情</button>
                  </div>
                </template>
              </div>
            </div>
          </div>
        </section>
      </main>
    </div>

    <AModal v-if="detailEvent" size="lg" @close="detailEvent = null">
      <div class="a-modal-header">
        <h2 class="a-modal-title">{{ detailEvent.title }}</h2>
        <button class="a-modal-close" @click="detailEvent = null">✕</button>
      </div>
      <div class="a-modal-body">
        <div class="tl-detail-meta">
          <span>{{ formatDatetime(detailEvent.event_date) }}</span>
          <span v-if="detailEvent.end_date"> — {{ formatDatetime(detailEvent.end_date) }}</span>
          <span v-if="detailEvent.category" class="tl-badge">{{ detailEvent.category }}</span>
        </div>
        <div v-if="detailEvent.location" class="tl-detail-field">
          <span class="tl-field-label">所在位置</span>
          <span>{{ detailEvent.location }}</span>
        </div>
        <div v-if="detailEvent.source" class="tl-detail-field">
          <span class="tl-field-label">来源</span>
          <span>{{ detailEvent.source }}</span>
        </div>
        <p v-if="detailEvent.description" class="tl-detail-desc">{{ detailEvent.description }}</p>
        <div v-if="detailEvent.content" class="tl-detail-content" v-html="renderContent(detailEvent.content)" />
        <div v-if="detailEvent.tags?.length" class="tl-tags">
          <span v-for="tag in detailEvent.tags" :key="tag" class="a-badge">{{ tag }}</span>
        </div>
      </div>
      <template #footer>
        <div class="a-modal-footer" v-if="canEdit(detailEvent)">
          <ABtn outline @click="openEdit(detailEvent)">编辑</ABtn>
          <ABtn variant="danger" @click="confirmDelete(detailEvent)">删除</ABtn>
        </div>
      </template>
    </AModal>

    <AModal v-if="showForm" size="lg" @close="closeForm">
      <div class="a-modal-header">
        <h2 class="a-modal-title">{{ editingEvent ? '编辑事件' : '新建事件' }}</h2>
        <button class="a-modal-close" @click="closeForm">✕</button>
      </div>
      <div class="a-modal-body">
        <div class="form-group">
          <label class="form-label">标题 *</label>
          <AInput v-model="form.title" placeholder="事件标题" />
        </div>
        <div class="form-row">
          <div class="form-group" style="flex:1">
            <label class="form-label">事件时间 *</label>
            <DatetimePicker v-model="form.event_date" placeholder="选择事件时间" />
          </div>
          <div class="form-group" style="flex:1">
            <label class="form-label">结束时间 (可选)</label>
            <DatetimePicker v-model="form.end_date" placeholder="选择结束时间" />
          </div>
        </div>
        <div class="form-row">
          <div class="form-group" style="flex:1">
            <label class="form-label">所在位置 *</label>
            <AInput v-model="form.location" placeholder="如 巴黎, 法国" />
          </div>
          <div class="form-group" style="flex:1">
            <label class="form-label">分类</label>
            <AInput v-model="form.category" placeholder="如 政治、科技、文化" />
          </div>
        </div>
        <div class="form-row">
          <div class="form-group" style="flex:1">
            <label class="form-label">纬度</label>
            <input
              v-model.number="form.latitude"
              type="number"
              step="any"
              min="-90"
              max="90"
              class="a-input"
              placeholder="48.8566"
            />
          </div>
          <div class="form-group" style="flex:1">
            <label class="form-label">经度</label>
            <input
              v-model.number="form.longitude"
              type="number"
              step="any"
              min="-180"
              max="180"
              class="a-input"
              placeholder="2.3522"
            />
          </div>
        </div>
        <p class="tl-field-help">纬度范围 -90 到 90，经度范围 -180 到 180；不确定时可留空。</p>
        <p v-if="formError" class="tl-form-error">{{ formError }}</p>
        <div class="form-group">
          <label class="form-label">来源 *</label>
          <AInput v-model="form.source" placeholder="如 《法国大革命史》第3章 / https://..." />
        </div>
        <div class="form-group">
          <label class="form-label">简介</label>
          <ATextarea v-model="form.description" :rows="2" placeholder="一句话简介" />
        </div>
        <div class="form-group">
          <label class="form-label">详细内容</label>
          <ATextarea v-model="form.content" :rows="5" placeholder="详细描述（支持 Markdown）" />
        </div>
        <div class="form-group">
          <label class="form-label">标签 (逗号分隔)</label>
          <AInput v-model="tagsInput" placeholder="战争, 革命, 法国" />
        </div>
      </div>
      <template #footer>
        <div class="a-modal-footer">
          <ABtn outline @click="closeForm">取消</ABtn>
          <ABtn :disabled="submitting" @click="submitForm">
            {{ submitting ? '保存中...' : (editingEvent ? '保存' : '创建') }}
          </ABtn>
        </div>
      </template>
    </AModal>

    <AModal v-if="showPersonForm" size="md" @close="showPersonForm = false">
      <div class="a-modal-header">
        <h2 class="a-modal-title">新建人物</h2>
        <button class="a-modal-close" @click="showPersonForm = false">✕</button>
      </div>
      <div class="a-modal-body">
        <div class="form-group">
          <label class="form-label">姓名 *</label>
          <AInput v-model="personForm.name" placeholder="历史人物姓名" />
        </div>
        <div class="form-row">
          <div class="form-group" style="flex:1">
            <label class="form-label">出生年份</label>
            <AInput v-model="personForm.birth_date" placeholder="YYYY-MM-DD" />
          </div>
          <div class="form-group" style="flex:1">
            <label class="form-label">去世年份</label>
            <AInput v-model="personForm.death_date" placeholder="YYYY-MM-DD" />
          </div>
        </div>
        <div class="form-group">
          <label class="form-label">简介</label>
          <ATextarea v-model="personForm.bio" :rows="3" placeholder="人物生平简介" />
        </div>
        <div class="form-group">
          <label class="form-label">标签 (逗号分隔)</label>
          <AInput v-model="personTagsInput" placeholder="政治家, 军事家" />
        </div>
      </div>
      <template #footer>
        <div class="a-modal-footer">
          <ABtn outline @click="showPersonForm = false">取消</ABtn>
          <ABtn :disabled="personSubmitting" @click="submitPerson">
            {{ personSubmitting ? '创建中...' : '创建并添加轨迹' }}
          </ABtn>
        </div>
      </template>
    </AModal>

    <AConfirm
      :show="!!deletingEvent"
      title="删除事件"
      :message="deletingEvent ? `确定要删除「${deletingEvent.title}」吗？此操作不可撤销。` : ''"
      @confirm="doDelete"
      @cancel="deletingEvent = null"
    />
  </div>
</template>

<script setup lang="ts">
import { computed, nextTick, onBeforeUnmount, onMounted, ref, watch } from 'vue'
import { storeToRefs } from 'pinia'
import { useRoute, useRouter } from 'vue-router'
import OlMap from 'ol/Map'
import View from 'ol/View'
import TileLayer from 'ol/layer/Tile'
import OSM from 'ol/source/OSM'
import VectorLayer from 'ol/layer/Vector'
import VectorSource from 'ol/source/Vector'
import Feature from 'ol/Feature'
import Point from 'ol/geom/Point'
import { fromLonLat } from 'ol/proj'
import { Circle as CircleStyle, Fill, Stroke, Style } from 'ol/style'
import Overlay from 'ol/Overlay'
import { useTimelineStore } from '@/stores/timeline'
import { useAuthStore } from '@/stores/auth'
import type { TimelineEvent } from '@/types'
import ABtn from '@/components/ui/ABtn.vue'
import AModal from '@/components/ui/AModal.vue'
import AEmpty from '@/components/ui/AEmpty.vue'
import APageHeader from '@/components/ui/APageHeader.vue'
import AInput from '@/components/ui/AInput.vue'
import ATextarea from '@/components/ui/ATextarea.vue'
import AConfirm from '@/components/ui/AConfirm.vue'
import DatetimePicker from '@/components/ui/DatetimePicker.vue'

type TimelineViewMode = 'lanes' | 'map'

const API_URL = import.meta.env.VITE_API_URL || '/api'
const DAY_MS = 24 * 60 * 60 * 1000

const store = useTimelineStore()
const authStore = useAuthStore()
const route = useRoute()
const router = useRouter()

const { events, loading } = storeToRefs(store)

const viewMode = ref<TimelineViewMode>('lanes')

const yearStart = ref<number | null>(null)
const yearEnd = ref<number | null>(null)
const filterCategory = ref('')

const compareIds = ref<string[]>([])
const activeCompareId = ref<string | null>(null)
const batchSelectedIds = ref<string[]>([])
const hydratedCompareEvents = ref<TimelineEvent[]>([])
const hydratingCompare = ref(false)
const routeSyncing = ref(false)

const sortedEvents = computed(() =>
  [...events.value].sort((a, b) => new Date(a.event_date).getTime() - new Date(b.event_date).getTime())
)

const knownEvents = computed(() => {
  const map = new Map<string, TimelineEvent>()
  for (const event of hydratedCompareEvents.value) {
    map.set(event.id, event)
  }
  for (const event of sortedEvents.value) {
    map.set(event.id, event)
  }
  return map
})

const compareEvents = computed(() =>
  compareIds.value
    .map((id) => knownEvents.value.get(id))
    .filter((event): event is TimelineEvent => Boolean(event))
)

const activeCompareEvent = computed(() =>
  compareEvents.value.find((event) => event.id === activeCompareId.value) ?? compareEvents.value[0] ?? null
)

const compareSet = computed(() => new Set(compareIds.value))

const normalizeSingleQueryValue = (value: unknown): string | null => {
  if (typeof value === 'string') return value
  if (Array.isArray(value) && typeof value[0] === 'string') return value[0]
  return null
}

const uniqueIds = (ids: string[]) => Array.from(new Set(ids.filter(Boolean)))

const sameIds = (left: string[], right: string[]) =>
  left.length === right.length && left.every((value, index) => value === right[index])

const parseCompareQuery = (value: unknown): string[] => {
  if (typeof value === 'string') {
    return value.split(',').map((item) => item.trim()).filter(Boolean)
  }
  if (Array.isArray(value)) {
    return value.flatMap((item) => parseCompareQuery(item))
  }
  return []
}

const parseModeQuery = (value: unknown): TimelineViewMode =>
  normalizeSingleQueryValue(value) === 'map' ? 'map' : 'lanes'

const getEventStartMs = (event: TimelineEvent) => new Date(event.event_date).getTime()

const getEventEndMs = (event: TimelineEvent) => {
  if (!event.end_date) return getEventStartMs(event)
  const endMs = new Date(event.end_date).getTime()
  return Number.isFinite(endMs) ? endMs : getEventStartMs(event)
}

const isInstantEvent = (event: TimelineEvent) => getEventEndMs(event) <= getEventStartMs(event)

const canMapEvent = (event: TimelineEvent) =>
  typeof event.latitude === 'number' &&
  Number.isFinite(event.latitude) &&
  typeof event.longitude === 'number' &&
  Number.isFinite(event.longitude)

const mapRenderableEvents = computed(() => {
  const map = new Map<string, TimelineEvent>()
  for (const event of sortedEvents.value) {
    if (canMapEvent(event)) map.set(event.id, event)
  }
  for (const event of compareEvents.value) {
    if (canMapEvent(event)) map.set(event.id, event)
  }
  return Array.from(map.values()).sort((a, b) => getEventStartMs(a) - getEventStartMs(b))
})

const compareBounds = computed(() => {
  if (!compareEvents.value.length) return null

  let min = Number.POSITIVE_INFINITY
  let max = Number.NEGATIVE_INFINITY

  for (const event of compareEvents.value) {
    min = Math.min(min, getEventStartMs(event))
    max = Math.max(max, getEventEndMs(event))
  }

  const safeMax = max === min ? max + DAY_MS : max

  return {
    min,
    max: safeMax,
    span: safeMax - min,
  }
})

const formatDatetime = (value: string) => {
  if (!value) return ''

  const date = new Date(value)
  if (Number.isNaN(date.getTime())) return value.slice(0, 16)

  const hours = date.getHours()
  const minutes = date.getMinutes()
  const dateLabel = value.slice(0, 10)

  if (hours === 0 && minutes === 0) return dateLabel

  return `${dateLabel} ${String(hours).padStart(2, '0')}:${String(minutes).padStart(2, '0')}`
}

const formatTickLabel = (timestamp: number) => {
  const bounds = compareBounds.value
  if (!bounds) return ''

  const totalDays = bounds.span / DAY_MS
  const date = new Date(timestamp)
  const year = date.getFullYear()
  const month = String(date.getMonth() + 1).padStart(2, '0')
  const day = String(date.getDate()).padStart(2, '0')

  if (totalDays > 365 * 2) return `${year}`
  if (totalDays > 60) return `${year}-${month}`
  return `${year}-${month}-${day}`
}

const formatEventRange = (event: TimelineEvent) =>
  isInstantEvent(event)
    ? formatDatetime(event.event_date)
    : `${formatDatetime(event.event_date)} — ${formatDatetime(event.end_date || event.event_date)}`

const getDurationLabel = (event: TimelineEvent) => {
  if (isInstantEvent(event)) return '单点事件'

  const diffMs = Math.max(getEventEndMs(event) - getEventStartMs(event), 0)
  const totalMinutes = diffMs / (60 * 1000)
  const totalHours = diffMs / (60 * 60 * 1000)
  const totalDays = diffMs / DAY_MS

  if (totalMinutes < 60) return `${Math.max(Math.round(totalMinutes), 1)} 分钟`
  if (totalHours < 24) return `${Math.max(Math.round(totalHours), 1)} 小时`
  if (totalDays < 60) return `${Math.max(Math.round(totalDays), 1)} 天`

  const totalMonths = totalDays / 30
  if (totalMonths < 24) {
    return `${totalMonths < 6 ? totalMonths.toFixed(1) : Math.round(totalMonths)} 个月`
  }

  const totalYears = totalDays / 365
  return `${totalYears < 10 ? totalYears.toFixed(1) : Math.round(totalYears)} 年`
}

const laneTicks = computed(() => {
  const bounds = compareBounds.value
  if (!bounds) return []

  const tickCount = Math.min(Math.max(compareEvents.value.length + 1, 4), 6)

  return Array.from({ length: tickCount }, (_, index) => {
    const ratio = tickCount === 1 ? 0 : index / (tickCount - 1)
    const timestamp = bounds.min + bounds.span * ratio

    return {
      pct: ratio * 100,
      label: formatTickLabel(timestamp),
    }
  })
})

const clampPercent = (value: number) => Math.max(0, Math.min(100, value))

const getLaneStyle = (event: TimelineEvent) => {
  const bounds = compareBounds.value
  if (!bounds) return {}

  const startRatio = (getEventStartMs(event) - bounds.min) / bounds.span
  const left = clampPercent(startRatio * 100)

  if (isInstantEvent(event)) {
    return { left: `${left}%` }
  }

  const endRatio = (getEventEndMs(event) - bounds.min) / bounds.span
  const width = Math.max((endRatio - startRatio) * 100, 1.75)

  return {
    left: `${left}%`,
    width: `${width}%`,
  }
}

const renderContent = (content: string) => content.replace(/\n/g, '<br>')

const isCompared = (id: string) => compareSet.value.has(id)

const isBatchSelected = (id: string) => batchSelectedIds.value.includes(id)

const toggleBatchSelection = (id: string) => {
  batchSelectedIds.value = isBatchSelected(id)
    ? batchSelectedIds.value.filter((value) => value !== id)
    : [...batchSelectedIds.value, id]
}

const clearBatchSelection = () => {
  batchSelectedIds.value = []
}

const setActiveCompare = (id: string) => {
  if (!compareSet.value.has(id)) return
  activeCompareId.value = id
}

const upsertHydratedEvent = (event: TimelineEvent) => {
  hydratedCompareEvents.value = [...hydratedCompareEvents.value.filter((item) => item.id !== event.id), event]
}

const removeHydratedEvent = (id: string) => {
  hydratedCompareEvents.value = hydratedCompareEvents.value.filter((item) => item.id !== id)
}

const addCompareEvent = (event: TimelineEvent) => {
  upsertHydratedEvent(event)
  compareIds.value = uniqueIds([...compareIds.value, event.id])
  activeCompareId.value = event.id
}

const removeCompareId = (id: string) => {
  compareIds.value = compareIds.value.filter((value) => value !== id)
  if (activeCompareId.value === id) {
    activeCompareId.value = compareIds.value[compareIds.value.length - 1] ?? null
  }
}

const toggleCompareEvent = (event: TimelineEvent) => {
  if (isCompared(event.id)) {
    removeCompareId(event.id)
    return
  }

  addCompareEvent(event)
}

const addBatchToCompare = () => {
  const additions = sortedEvents.value
    .filter((event) => batchSelectedIds.value.includes(event.id))
    .map((event) => event.id)

  if (!additions.length) return

  compareIds.value = uniqueIds([...compareIds.value, ...additions])
  activeCompareId.value = additions[additions.length - 1]
}

const clearComparePool = () => {
  compareIds.value = []
  activeCompareId.value = null
  hideMapPopup()
}

const fetchEventById = async (id: string) => {
  try {
    const response = await fetch(`${API_URL}/timeline/events/${id}`)
    if (!response.ok) return null
    const data = await response.json()
    return data.data as TimelineEvent
  } catch (error) {
    console.error(error)
    return null
  }
}

const hydrateComparePool = async (ids: string[]) => {
  const missingIds = uniqueIds(ids.filter((id) => !knownEvents.value.has(id)))
  if (!missingIds.length) return

  hydratingCompare.value = true
  try {
    const fetchedEvents = await Promise.all(missingIds.map((id) => fetchEventById(id)))
    const resolvedEvents = fetchedEvents.filter((event): event is TimelineEvent => Boolean(event))
    const resolvedIds = new Set(resolvedEvents.map((event) => event.id))

    for (const event of resolvedEvents) {
      upsertHydratedEvent(event)
    }

    const invalidIds = missingIds.filter((id) => !resolvedIds.has(id))
    if (invalidIds.length) {
      compareIds.value = compareIds.value.filter((id) => !invalidIds.includes(id))
    }
  } finally {
    hydratingCompare.value = false
  }
}

const canEdit = (event: TimelineEvent) =>
  authStore.isAuthenticated &&
  (event.user_id === authStore.user?.uuid || authStore.user?.role === 'admin')

const detailEvent = ref<TimelineEvent | null>(null)
const showForm = ref(false)
const editingEvent = ref<TimelineEvent | null>(null)
const deletingEvent = ref<TimelineEvent | null>(null)
const submitting = ref(false)
const formError = ref('')

const emptyForm = () => ({
  title: '',
  event_date: '',
  end_date: '',
  location: '',
  latitude: null as number | null,
  longitude: null as number | null,
  source: '',
  category: '',
  description: '',
  content: '',
})

const form = ref(emptyForm())
const tagsInput = ref('')

const isFiniteCoordinate = (value: number | null) => typeof value === 'number' && Number.isFinite(value)

const getCoordinateValidationError = () => {
  const { latitude, longitude } = form.value
  const hasLatitude = isFiniteCoordinate(latitude)
  const hasLongitude = isFiniteCoordinate(longitude)

  if (hasLatitude !== hasLongitude) {
    return '经纬度需要同时填写，或同时留空。'
  }

  if (hasLatitude && latitude !== null && (latitude < -90 || latitude > 90)) {
    return '纬度必须在 -90 到 90 之间。'
  }

  if (hasLongitude && longitude !== null && (longitude < -180 || longitude > 180)) {
    return '经度必须在 -180 到 180 之间。'
  }

  return ''
}

const openDetail = (event: TimelineEvent) => {
  if (isCompared(event.id)) {
    activeCompareId.value = event.id
  }
  detailEvent.value = event
}

const openCreate = () => {
  editingEvent.value = null
  form.value = emptyForm()
  tagsInput.value = ''
  formError.value = ''
  showForm.value = true
}

const openEdit = (event: TimelineEvent) => {
  if (isCompared(event.id)) {
    activeCompareId.value = event.id
  }

  editingEvent.value = event
  form.value = {
    title: event.title,
    event_date: event.event_date.slice(0, 16).replace(' ', 'T'),
    end_date: event.end_date ? event.end_date.slice(0, 16).replace(' ', 'T') : '',
    location: event.location || '',
    latitude: event.latitude ?? null,
    longitude: event.longitude ?? null,
    source: event.source || '',
    category: event.category || '',
    description: event.description || '',
    content: event.content || '',
  }
  tagsInput.value = (event.tags || []).join(', ')
  detailEvent.value = null
  formError.value = ''
  showForm.value = true
}

const closeForm = () => {
  showForm.value = false
  editingEvent.value = null
  formError.value = ''
}

const submitForm = async () => {
  if (!form.value.title || !form.value.event_date || !form.value.location || !form.value.source) return

  const coordinateError = getCoordinateValidationError()
  if (coordinateError) {
    formError.value = coordinateError
    return
  }

  formError.value = ''
  submitting.value = true

  try {
    const tags = tagsInput.value.split(',').map((tag) => tag.trim()).filter(Boolean)
    const payload = {
      ...form.value,
      latitude: typeof form.value.latitude === 'number' && Number.isFinite(form.value.latitude)
        ? form.value.latitude
        : null,
      longitude: typeof form.value.longitude === 'number' && Number.isFinite(form.value.longitude)
        ? form.value.longitude
        : null,
      tags,
    }

    const savedEvent = editingEvent.value
      ? await store.updateEvent(editingEvent.value.id, payload)
      : await store.createEvent(payload)

    if (savedEvent) {
      upsertHydratedEvent(savedEvent)
      if (detailEvent.value?.id === savedEvent.id) {
        detailEvent.value = savedEvent
      }
      if (mapPopupEvent.value?.id === savedEvent.id) {
        mapPopupEvent.value = savedEvent
      }
    }

    closeForm()
  } catch (error) {
    console.error(error)
    formError.value = error instanceof Error ? error.message : '保存失败，请稍后重试。'
  } finally {
    submitting.value = false
  }
}

const confirmDelete = (event: TimelineEvent) => {
  detailEvent.value = null
  deletingEvent.value = event
}

const doDelete = async () => {
  if (!deletingEvent.value) return

  const deletingId = deletingEvent.value.id

  await store.deleteEvent(deletingId)
  removeCompareId(deletingId)
  removeHydratedEvent(deletingId)

  if (mapPopupEvent.value?.id === deletingId) {
    hideMapPopup()
  }

  if (detailEvent.value?.id === deletingId) {
    detailEvent.value = null
  }

  deletingEvent.value = null
}

const showPersonForm = ref(false)
const personSubmitting = ref(false)
const personForm = ref({ name: '', bio: '', birth_date: '', death_date: '' })
const personTagsInput = ref('')

const submitPerson = async () => {
  if (!personForm.value.name) return

  personSubmitting.value = true
  try {
    const tags = personTagsInput.value.split(',').map((tag) => tag.trim()).filter(Boolean)
    const created = await store.createPerson({ ...personForm.value, tags })
    showPersonForm.value = false
    personForm.value = { name: '', bio: '', birth_date: '', death_date: '' }
    personTagsInput.value = ''
    router.push(`/timeline/persons/${created.id}`)
  } catch (error) {
    console.error(error)
  } finally {
    personSubmitting.value = false
  }
}

const applyFilter = () => {
  store.fetchEvents({
    category: filterCategory.value || undefined,
    yearStart: yearStart.value || undefined,
    yearEnd: yearEnd.value || undefined,
    limit: 200,
  })
}

const resetFilter = () => {
  yearStart.value = null
  yearEnd.value = null
  filterCategory.value = ''
  store.fetchEvents({ limit: 200 })
}

const mapPopupVisible = ref(false)
const mapPopupEvent = ref<TimelineEvent | null>(null)
const eventMapEl = ref<HTMLElement | null>(null)
const eventPopupEl = ref<HTMLElement | null>(null)

let eventMap: OlMap | null = null
let eventVectorSource: VectorSource | null = null
let eventPopupOverlay: Overlay | null = null

const mapOverlayMessage = computed(() => {
  if (!mapRenderableEvents.value.length) {
    return '当前筛选结果与对比池中暂无可定位事件。'
  }

  if (!compareEvents.value.length) {
    return '点击任意地图标记查看事件，并把它加入对比池。'
  }

  if (!mapPopupEvent.value) {
    return '地图会高亮对比池中的事件。点击标记可以继续添加或查看详情。'
  }

  return ''
})

const mapFocusEvent = computed(() => {
  if (mapPopupEvent.value && canMapEvent(mapPopupEvent.value)) {
    return mapPopupEvent.value
  }

  if (activeCompareEvent.value && canMapEvent(activeCompareEvent.value)) {
    return activeCompareEvent.value
  }

  return null
})

const buildEventPointStyle = (event: TimelineEvent) => {
  const isActive = mapPopupEvent.value?.id === event.id || activeCompareId.value === event.id
  const inCompare = compareSet.value.has(event.id)

  return new Style({
    image: new CircleStyle({
      radius: isActive ? 10 : inCompare ? 8 : 6,
      fill: new Fill({ color: inCompare || isActive ? 'var(--a-color-fg)' : 'var(--a-color-bg)' }),
      stroke: new Stroke({ color: inCompare || isActive ? 'var(--a-color-bg)' : 'var(--a-color-fg)', width: isActive ? 3 : 2 }),
    }),
  })
}

const hideMapPopup = () => {
  mapPopupVisible.value = false
  mapPopupEvent.value = null
  eventPopupOverlay?.setPosition(undefined)
}

const setMapPopup = (event: TimelineEvent, coordinate?: number[]) => {
  mapPopupEvent.value = event

  if (!coordinate || !eventPopupOverlay) {
    hideMapPopup()
    return
  }

  eventPopupOverlay.setPosition(coordinate)
  mapPopupVisible.value = true
}

const fitMapToEvents = () => {
  if (!eventMap || !eventVectorSource || !mapRenderableEvents.value.length) return

  const extent = eventVectorSource.getExtent()

  if (mapRenderableEvents.value.length === 1) {
    const event = mapRenderableEvents.value[0]
    eventMap.getView().setCenter(fromLonLat([event.longitude!, event.latitude!]))
    eventMap.getView().setZoom(4)
    return
  }

  eventMap.getView().fit(extent, { padding: [40, 40, 40, 40], maxZoom: 7 })
}

const focusMapEventOnMap = (animate = true) => {
  if (!eventMap || !mapFocusEvent.value || !canMapEvent(mapFocusEvent.value)) {
    hideMapPopup()
    if (mapRenderableEvents.value.length) {
      fitMapToEvents()
    }
    return
  }

  const coordinate = fromLonLat([mapFocusEvent.value.longitude!, mapFocusEvent.value.latitude!])

  if (mapPopupEvent.value?.id === mapFocusEvent.value.id) {
    setMapPopup(mapFocusEvent.value, coordinate)
  }

  const view = eventMap.getView()

  if (animate) {
    view.animate({
      center: coordinate,
      zoom: Math.max(view.getZoom() || 3, 4),
      duration: 350,
    })
    return
  }

  view.setCenter(coordinate)
  if ((view.getZoom() || 0) < 4) {
    view.setZoom(4)
  }
}

const renderEventMap = () => {
  if (!eventVectorSource) return

  eventVectorSource.clear()

  for (const event of mapRenderableEvents.value) {
    const feature = new Feature({
      geometry: new Point(fromLonLat([event.longitude!, event.latitude!])),
      eventId: event.id,
    })
    feature.setStyle(buildEventPointStyle(event))
    eventVectorSource.addFeature(feature)
  }

  if (!eventMap) return

  if (mapPopupEvent.value && !mapRenderableEvents.value.some((event) => event.id === mapPopupEvent.value?.id)) {
    hideMapPopup()
  }

  if (!mapRenderableEvents.value.length) {
    hideMapPopup()
    eventMap.getView().setCenter(fromLonLat([20, 30]))
    eventMap.getView().setZoom(2)
    return
  }

  if (mapFocusEvent.value && canMapEvent(mapFocusEvent.value)) {
    focusMapEventOnMap(false)
    return
  }

  fitMapToEvents()
}

const initEventMap = () => {
  if (!eventMapEl.value || !eventPopupEl.value) return

  if (eventMap) {
    eventMap.setTarget(eventMapEl.value)
    eventPopupOverlay?.setElement(eventPopupEl.value)
    eventMap.updateSize()
    return
  }

  eventVectorSource = new VectorSource()
  const vectorLayer = new VectorLayer({ source: eventVectorSource })

  eventPopupOverlay = new Overlay({
    element: eventPopupEl.value,
    positioning: 'bottom-center',
    stopEvent: true,
    offset: [0, -18],
  })

  eventMap = new OlMap({
    target: eventMapEl.value,
    layers: [new TileLayer({ source: new OSM() }), vectorLayer],
    overlays: [eventPopupOverlay],
    view: new View({
      center: fromLonLat([20, 30]),
      zoom: 2,
    }),
  })

  eventMap.on('click', (evt: any) => {
    const feature = eventMap?.forEachFeatureAtPixel(evt.pixel, (item) => item)
    if (!feature) {
      hideMapPopup()
      return
    }

    const eventId = feature.get('eventId') as string | undefined
    const event = knownEvents.value.get(eventId || '')

    if (!event) {
      hideMapPopup()
      return
    }

    const coordinate = (feature.getGeometry() as Point).getCoordinates()
    setMapPopup(event, coordinate)
    if (isCompared(event.id)) {
      activeCompareId.value = event.id
    }
    renderEventMap()
  })

  eventMap.on('pointermove', (evt: any) => {
    const hit = eventMap?.hasFeatureAtPixel(evt.pixel)
    if (eventMap) {
      eventMap.getTargetElement().style.cursor = hit ? 'pointer' : ''
    }
  })

  renderEventMap()
}

watch(
  [() => route.query.mode, () => route.query.compare],
  () => {
    if (routeSyncing.value) return

    const nextMode = parseModeQuery(route.query.mode)
    const nextCompareIds = uniqueIds(parseCompareQuery(route.query.compare))

    if (viewMode.value !== nextMode) {
      viewMode.value = nextMode
    }

    if (!sameIds(compareIds.value, nextCompareIds)) {
      compareIds.value = nextCompareIds
    }
  },
  { immediate: true }
)

watch(
  [viewMode, () => compareIds.value.join(',')],
  async () => {
    const currentMode = parseModeQuery(route.query.mode)
    const currentCompare = uniqueIds(parseCompareQuery(route.query.compare))
    const targetCompare = compareIds.value

    if (currentMode === viewMode.value && sameIds(currentCompare, targetCompare)) {
      return
    }

    routeSyncing.value = true
    try {
      await router.replace({
        query: {
          ...route.query,
          mode: viewMode.value === 'lanes' ? undefined : viewMode.value,
          compare: compareIds.value.length ? compareIds.value.join(',') : undefined,
        },
      })
    } finally {
      routeSyncing.value = false
    }
  }
)

watch(
  () => compareIds.value.join(','),
  () => {
    const ids = compareIds.value

    if (!ids.length) {
      activeCompareId.value = null
      return
    }

    if (!activeCompareId.value || !ids.includes(activeCompareId.value)) {
      activeCompareId.value = ids[ids.length - 1]
    }

    hydrateComparePool(ids)
  },
  { immediate: true }
)

watch(
  sortedEvents,
  (nextEvents) => {
    const availableIds = new Set(nextEvents.map((event) => event.id))
    batchSelectedIds.value = batchSelectedIds.value.filter((id) => availableIds.has(id))
  },
  { immediate: true }
)

watch(
  [
    loading,
    viewMode,
    () => mapRenderableEvents.value.map((event) => `${event.id}:${event.updated_at}`).join('|'),
    () => compareIds.value.join(','),
    activeCompareId,
    () => mapPopupEvent.value?.id,
  ],
  async ([isLoading]) => {
    if (isLoading) return

    await nextTick()
    initEventMap()

    if (viewMode.value === 'map') {
      eventMap?.updateSize()
    }

    renderEventMap()
  },
  { immediate: true }
)

onMounted(() => {
  store.fetchEvents({ limit: 200 })

  if (route.query.create === 'event' && authStore.isAuthenticated) {
    openCreate()
  }
})

onBeforeUnmount(() => {
  eventMap?.setTarget(undefined)
  eventMap = null
  eventVectorSource = null
  eventPopupOverlay = null
})
</script>

<style scoped>
.tl-page {
  padding-bottom: 6rem;
}

.tl-page-actions {
  display: flex;
  gap: 0.75rem;
}

.filter-label {
  font-size: 0.7rem;
  font-weight: 900;
  text-transform: uppercase;
  letter-spacing: 0.08em;
  color: var(--a-color-muted);
}

.tl-filter-group {
  display: flex;
  flex-direction: column;
  gap: 4px;
}

.tl-toolbar {
  display: flex;
  gap: 1rem;
  margin-bottom: 1.5rem;
  flex-wrap: wrap;
  align-items: flex-end;
  padding: 1.1rem 1.15rem 1.2rem;
  border: 2px solid var(--a-color-fg);
  background:
    linear-gradient(180deg, rgba(0, 0, 0, 0.04), transparent 32%),
    repeating-linear-gradient(90deg, rgba(0, 0, 0, 0.035) 0, rgba(0, 0, 0, 0.035) 1px, transparent 1px, transparent 18px),
    var(--a-color-bg);
  box-shadow: 10px 10px 0 0 rgba(0, 0, 0, 1);
}

.tl-toolbar-batch {
  display: flex;
  align-items: center;
  gap: 0.75rem;
  flex-wrap: wrap;
}

.tl-toolbar-batch-count {
  font-size: 0.72rem;
  font-weight: 900;
  text-transform: uppercase;
  letter-spacing: 0.08em;
  color: var(--a-color-muted);
}

.tl-mode-switch {
  display: inline-flex;
  border: 2px solid var(--a-color-fg);
  background: var(--a-color-bg);
  box-shadow: 6px 6px 0 0 rgba(0, 0, 0, 1);
}

.tl-mode-btn {
  border: none;
  border-right: 2px solid var(--a-color-fg);
  background: var(--a-color-bg);
  color: var(--a-color-fg);
  cursor: pointer;
  padding: 0.6rem 1.05rem;
  font-size: 0.7rem;
  font-weight: 900;
  text-transform: uppercase;
  letter-spacing: 0.08em;
}

.tl-mode-btn:last-child {
  border-right: none;
}

.tl-mode-btn-active,
.tl-mode-btn:hover {
  background: var(--a-color-fg);
  color: var(--a-color-bg);
}

.tl-state-block {
  padding: 4rem;
  text-align: center;
}

.tl-shell {
  display: grid;
  grid-template-columns: minmax(320px, 380px) minmax(0, 1fr);
  gap: 1.25rem;
  align-items: start;
}

.tl-source-panel,
.tl-stage {
  display: flex;
  flex-direction: column;
  gap: 1.15rem;
}

.tl-panel-card,
.tl-lane-board,
.tl-map-stage-card,
.tl-stage-empty {
  padding: 1.35rem;
  background:
    linear-gradient(180deg, rgba(0, 0, 0, 0.04), transparent 28%),
    var(--a-color-bg);
  box-shadow: 10px 10px 0 0 rgba(0, 0, 0, 1);
}

.tl-panel-head,
.tl-stage-head {
  display: flex;
  align-items: flex-start;
  justify-content: space-between;
  gap: 0.75rem;
  margin-bottom: 1rem;
}

.tl-panel-kicker,
.tl-section-kicker {
  margin-bottom: 0.35rem;
  font-size: 0.68rem;
  font-weight: 900;
  text-transform: uppercase;
  letter-spacing: 0.14em;
  color: var(--a-color-muted);
}

.tl-panel-title,
.tl-section-title {
  margin: 0;
  font-size: 1.4rem;
  font-weight: 900;
  letter-spacing: -0.05em;
  line-height: 1;
}

.tl-panel-count,
.tl-stage-meta {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  min-height: 2rem;
  padding: 0.2rem 0.7rem;
  border: 2px solid var(--a-color-fg);
  background: var(--a-color-bg);
  font-size: 0.7rem;
  font-weight: 900;
  text-transform: uppercase;
  letter-spacing: 0.08em;
  white-space: nowrap;
}

.tl-panel-note,
.tl-map-stage-note,
.tl-stage-empty p {
  margin: 0;
  font-size: 0.8rem;
  line-height: 1.65;
  color: var(--a-color-muted);
}

.tl-panel-empty {
  padding: 1rem;
  border: 2px dashed var(--a-color-muted-soft);
  font-size: 0.82rem;
  line-height: 1.6;
  color: var(--a-color-muted);
  background: rgba(255, 255, 255, 0.9);
}

.tl-panel-actions {
  display: flex;
  flex-wrap: wrap;
  gap: 0.75rem;
  margin-top: 1rem;
}

.tl-action-btn,
.tl-mini-btn {
  border: 2px solid var(--a-color-fg);
  background: var(--a-color-fg);
  color: var(--a-color-bg);
  cursor: pointer;
  font-size: 0.72rem;
  font-weight: 900;
  text-transform: uppercase;
  letter-spacing: 0.08em;
  padding: 0.55rem 0.8rem;
}

.tl-mini-btn {
  padding: 0.35rem 0.55rem;
}

.tl-action-btn-secondary,
.tl-mini-btn:not(.danger) {
  background: var(--a-color-bg);
  color: var(--a-color-fg);
}

.tl-action-btn.danger,
.tl-mini-btn.danger,
.tl-inline-link.danger {
  color: #991b1b;
}

.tl-compare-list,
.tl-source-list {
  display: flex;
  flex-direction: column;
  gap: 0.85rem;
}

.tl-source-list {
  max-height: calc(100vh - 18rem);
  overflow: auto;
  padding-right: 0.1rem;
}

.tl-compare-item,
.tl-source-item {
  display: flex;
  gap: 0.9rem;
  padding: 1rem;
  border: 2px solid var(--a-color-fg);
  background: var(--a-color-bg);
  box-shadow: 8px 8px 0 0 rgba(0, 0, 0, 1);
  transition: transform 0.18s ease, box-shadow 0.18s ease;
}

.tl-compare-item:hover,
.tl-source-item:hover {
  transform: translate(-3px, -3px);
  box-shadow: 11px 11px 0 0 rgba(0, 0, 0, 1);
}

.tl-compare-item-active,
.tl-source-item-active {
  transform: translate(-3px, -3px);
  box-shadow: 11px 11px 0 0 rgba(0, 0, 0, 1);
}

.tl-source-item-compared,
.tl-compare-item-active {
  background: var(--a-color-fg);
  color: var(--a-color-bg);
}

.tl-source-item-compared .tl-card-date,
.tl-source-item-compared .tl-source-desc,
.tl-source-item-compared .tl-card-loc,
.tl-source-item-compared .tl-source-duration,
.tl-compare-item-active .tl-compare-range {
  color: var(--a-color-bg);
}

.tl-source-item-compared .tl-card-category,
.tl-source-item-compared .tl-inline-chip,
.tl-compare-item-active .tl-inline-chip {
  border-color: var(--a-color-bg);
  color: var(--a-color-bg);
}

.tl-source-item-compared .tl-action-btn {
  background: var(--a-color-bg);
  color: var(--a-color-fg);
}

.tl-source-item-compared .tl-action-btn-secondary {
  background: transparent;
  color: var(--a-color-bg);
  border-color: var(--a-color-bg);
}

.tl-compare-order {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  width: 2rem;
  height: 2rem;
  border: 2px solid var(--a-color-fg);
  font-size: 0.8rem;
  font-weight: 900;
  flex-shrink: 0;
}

.tl-compare-body,
.tl-source-content {
  min-width: 0;
  flex: 1;
}

.tl-compare-title-row,
.tl-lane-info-top,
.tl-source-top {
  display: flex;
  align-items: flex-start;
  justify-content: space-between;
  gap: 0.75rem;
}

.tl-compare-title,
.tl-source-title,
.tl-lane-title {
  margin: 0;
  font-size: 1rem;
  font-weight: 900;
  letter-spacing: -0.03em;
}

.tl-compare-range,
.tl-source-desc,
.tl-source-duration,
.tl-lane-range {
  font-size: 0.78rem;
  line-height: 1.6;
  color: var(--a-color-muted);
}

.tl-compare-range,
.tl-lane-range {
  margin-top: 0.35rem;
}

.tl-compare-actions {
  display: flex;
  flex-direction: column;
  gap: 0.5rem;
  justify-content: center;
}

.tl-source-select {
  display: flex;
  flex-direction: column;
  align-items: flex-start;
  gap: 0.4rem;
  width: 4.5rem;
  flex-shrink: 0;
  font-size: 0.68rem;
  font-weight: 900;
  text-transform: uppercase;
  letter-spacing: 0.08em;
  color: var(--a-color-muted);
}

.tl-source-select input {
  accent-color: var(--a-color-fg);
}

.tl-card-date {
  font-size: 0.7rem;
  font-weight: 900;
  text-transform: uppercase;
  letter-spacing: 0.08em;
  color: var(--a-color-muted);
}

.tl-source-badges,
.tl-source-meta,
.tl-lane-meta,
.tl-source-actions,
.tl-lane-actions,
.tl-popup-actions {
  display: flex;
  flex-wrap: wrap;
  gap: 0.55rem;
}

.tl-source-badges {
  justify-content: flex-end;
}

.tl-source-desc {
  margin: 0.55rem 0 0.75rem;
  overflow: hidden;
  display: -webkit-box;
  -webkit-line-clamp: 2;
  -webkit-box-orient: vertical;
}

.tl-source-meta,
.tl-lane-meta {
  align-items: center;
  margin-bottom: 0.85rem;
}

.tl-source-actions,
.tl-lane-actions {
  margin-top: 0.75rem;
}

.tl-card-category {
  display: inline-block;
  font-size: 0.65rem;
  font-weight: 900;
  text-transform: uppercase;
  letter-spacing: 0.08em;
  border: 2px solid var(--a-color-fg);
  padding: 2px 6px;
}

.tl-card-loc {
  font-size: 0.74rem;
  font-weight: 700;
  color: var(--a-color-muted);
}

.tl-inline-chip {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  padding: 0.15rem 0.45rem;
  border: 1px solid var(--a-color-fg);
  font-size: 0.62rem;
  font-weight: 900;
  text-transform: uppercase;
  letter-spacing: 0.06em;
  color: var(--a-color-fg);
}

.tl-stage-pane {
  display: flex;
  flex-direction: column;
  gap: 1rem;
}

.tl-empty-title {
  margin-bottom: 0.5rem;
  font-size: 1.1rem;
  font-weight: 900;
  letter-spacing: -0.03em;
}

.tl-lane-board {
  display: flex;
  flex-direction: column;
  gap: 0.75rem;
}

.tl-lane-ruler {
  position: relative;
  height: 3.5rem;
  border-bottom: 2px solid var(--a-color-fg);
  margin-bottom: 0.25rem;
}

.tl-lane-ruler-tick {
  position: absolute;
  bottom: 0;
  transform: translateX(-50%);
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 0.3rem;
}

.tl-lane-ruler-line {
  width: 2px;
  height: 1rem;
  background: var(--a-color-muted-soft);
}

.tl-lane-ruler-label {
  font-size: 0.68rem;
  font-weight: 900;
  letter-spacing: 0.06em;
  white-space: nowrap;
}

.tl-lane-row {
  display: grid;
  grid-template-columns: minmax(250px, 0.9fr) minmax(0, 1.4fr);
  gap: 1rem;
  padding: 1rem 0;
  border-top: 2px solid rgba(0, 0, 0, 0.08);
}

.tl-lane-row:first-of-type {
  border-top: none;
}

.tl-lane-row-active {
  box-shadow: inset 6px 0 0 var(--a-color-fg);
  padding-left: 1rem;
}

.tl-lane-track {
  position: relative;
  min-height: 4.5rem;
  cursor: pointer;
}

.tl-lane-track-base {
  position: absolute;
  left: 0;
  right: 0;
  top: 50%;
  height: 4px;
  transform: translateY(-50%);
  background: linear-gradient(90deg, #d1d5db 0%, var(--a-color-fg) 100%);
}

.tl-lane-bar-shell {
  position: absolute;
  top: 50%;
  transform: translateY(-50%);
  display: flex;
  flex-direction: column;
  align-items: flex-start;
}

.tl-lane-bar-shell-point {
  width: 0 !important;
  align-items: center;
}

.tl-lane-bar {
  width: 100%;
  min-width: 1.25rem;
  height: 1rem;
  border-radius: 999px;
  border: 2px solid var(--a-color-fg);
  background: var(--a-color-bg);
  box-shadow: 6px 6px 0 0 rgba(0, 0, 0, 1);
}

.tl-lane-bar-active {
  background: var(--a-color-fg);
}

.tl-lane-bar-shell-point .tl-lane-bar {
  width: 1rem;
  height: 1rem;
  min-width: 1rem;
  border-radius: 999px;
  background: var(--a-color-fg);
  box-shadow: 0 0 0 4px var(--a-color-bg), 0 0 0 6px var(--a-color-fg);
  transform: translateX(-50%);
}

.tl-lane-bar-caption {
  margin-top: 0.6rem;
  font-size: 0.68rem;
  font-weight: 900;
  letter-spacing: 0.05em;
  white-space: nowrap;
}

.tl-lane-bar-shell-point .tl-lane-bar-caption {
  transform: translateX(-50%);
}

.tl-inline-link {
  border: none;
  background: none;
  padding: 0;
  cursor: pointer;
  font-size: 0.72rem;
  font-weight: 900;
  text-transform: uppercase;
  letter-spacing: 0.08em;
  color: var(--a-color-fg);
}

.tl-map-stage-note {
  display: flex;
  justify-content: space-between;
  gap: 1rem;
  flex-wrap: wrap;
  margin-bottom: 1rem;
}

.tl-map-canvas {
  position: relative;
  min-height: 36rem;
  border: 2px solid var(--a-color-fg);
  overflow: hidden;
  background:
    repeating-linear-gradient(0deg, var(--a-color-bg) 0, var(--a-color-bg) 29px, var(--a-color-disabled-bg) 29px, var(--a-color-disabled-bg) 30px),
    var(--a-color-bg);
}

.tl-event-map {
  width: 100%;
  height: 36rem;
}

.tl-map-overlay {
  position: absolute;
  left: 1rem;
  right: 1rem;
  bottom: 1rem;
  z-index: 10;
  padding: 0.75rem 0.9rem;
  background: rgba(255, 255, 255, 0.96);
  border: 2px solid var(--a-color-fg);
  font-size: 0.78rem;
  font-weight: 700;
  line-height: 1.55;
}

.tl-map-popup {
  min-width: 14rem;
  max-width: 18rem;
  padding: 0.75rem 0.85rem;
  background: rgba(255, 255, 255, 0.98);
  border: 2px solid var(--a-color-fg);
  box-shadow: 8px 8px 0 0 rgba(0, 0, 0, 1);
  opacity: 0;
  pointer-events: none;
  transition: opacity 0.15s;
}

.tl-map-popup.visible {
  opacity: 1;
  pointer-events: auto;
}

.tl-popup-close {
  position: absolute;
  top: 0.25rem;
  right: 0.35rem;
  border: none;
  background: none;
  cursor: pointer;
  font-size: 0.8rem;
  font-weight: 900;
}

.tl-popup-title {
  font-size: 0.95rem;
  font-weight: 900;
  letter-spacing: -0.02em;
  margin-bottom: 0.25rem;
  padding-right: 1rem;
}

.tl-popup-date,
.tl-popup-location,
.tl-popup-category {
  font-size: 0.72rem;
  line-height: 1.55;
  color: var(--a-color-muted);
}

.tl-popup-actions {
  margin-top: 0.85rem;
}

.tl-detail-meta {
  font-size: 0.8rem;
  font-weight: 700;
  color: var(--a-color-muted);
  margin-bottom: 0.75rem;
  display: flex;
  align-items: center;
  gap: 0.5rem;
  flex-wrap: wrap;
}

.tl-badge {
  font-size: 0.65rem;
  font-weight: 900;
  text-transform: uppercase;
  letter-spacing: 0.08em;
  border: 2px solid var(--a-color-fg);
  padding: 2px 6px;
  color: var(--a-color-fg);
}

.tl-detail-field {
  display: flex;
  align-items: baseline;
  gap: 0.5rem;
  font-size: 0.85rem;
  margin-bottom: 0.5rem;
}

.tl-field-label {
  font-size: 0.65rem;
  font-weight: 900;
  text-transform: uppercase;
  letter-spacing: 0.08em;
  color: var(--a-color-muted);
  flex-shrink: 0;
}

.tl-detail-desc {
  font-size: 0.9rem;
  color: var(--a-color-muted);
  margin-bottom: 1rem;
  line-height: 1.6;
}

.tl-detail-content {
  font-size: 0.875rem;
  line-height: 1.7;
  margin-bottom: 1rem;
}

.tl-tags {
  display: flex;
  flex-wrap: wrap;
  gap: 0.5rem;
}

.form-group {
  margin-bottom: 1rem;
}

.form-row {
  display: flex;
  gap: 1rem;
  margin-bottom: 1rem;
}

.form-label {
  display: block;
  font-size: 0.75rem;
  font-weight: 900;
  text-transform: uppercase;
  letter-spacing: 0.08em;
  margin-bottom: 0.4rem;
}

.tl-field-help {
  margin: -0.35rem 0 1rem;
  font-size: 0.72rem;
  line-height: 1.55;
  color: var(--a-color-muted);
}

.tl-form-error {
  margin: -0.25rem 0 1rem;
  padding: 0.7rem 0.8rem;
  border: 2px solid var(--a-color-fg);
  background: #fef2f2;
  color: #991b1b;
  font-size: 0.75rem;
  font-weight: 800;
  line-height: 1.5;
}

.a-modal-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 1.25rem 1.5rem;
  border-bottom: 2px solid var(--a-color-fg);
}

.a-modal-title {
  font-size: 1.25rem;
  font-weight: 900;
  letter-spacing: -0.03em;
}

.a-modal-close {
  background: none;
  border: none;
  font-size: 1rem;
  font-weight: 900;
  cursor: pointer;
  padding: 0.25rem 0.5rem;
  color: var(--a-color-fg);
}

.a-modal-body {
  padding: 1.5rem;
  overflow-y: auto;
  max-height: 60vh;
}

.a-modal-footer {
  display: flex;
  justify-content: flex-end;
  gap: 0.75rem;
  padding: 1rem 1.5rem;
  border-top: 2px solid var(--a-color-fg);
}

@media (max-width: 1180px) {
  .tl-shell {
    grid-template-columns: 1fr;
  }

  .tl-source-list {
    max-height: none;
  }
}

@media (max-width: 768px) {
  .tl-toolbar,
  .tl-panel-card,
  .tl-lane-board,
  .tl-map-stage-card,
  .tl-stage-empty,
  .tl-mode-switch,
  .tl-source-item,
  .tl-compare-item {
    box-shadow: 6px 6px 0 0 rgba(0, 0, 0, 1);
  }

  .tl-mode-switch {
    width: 100%;
    margin-left: 0 !important;
  }

  .tl-mode-btn {
    flex: 1;
  }

  .tl-panel-head,
  .tl-stage-head,
  .tl-map-stage-note,
  .tl-compare-title-row,
  .tl-source-top,
  .tl-lane-info-top {
    flex-direction: column;
    align-items: flex-start;
  }

  .form-row,
  .tl-lane-row,
  .tl-source-item,
  .tl-compare-item {
    grid-template-columns: 1fr;
    flex-direction: column;
  }

  .tl-source-select {
    width: auto;
    flex-direction: row;
    align-items: center;
  }

  .tl-compare-actions {
    flex-direction: row;
    justify-content: flex-start;
  }

  .tl-map-canvas,
  .tl-event-map {
    min-height: 24rem;
    height: 24rem;
  }

  .tl-lane-row-active {
    padding-left: 0.5rem;
  }
}
</style>