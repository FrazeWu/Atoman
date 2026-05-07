<template>
  <div class="person-map-page">
    <!-- Left panel -->
    <aside class="person-panel">
      <!-- Back + person info -->
      <div class="panel-header">
        <RouterLink to="/timeline/persons" class="back-link">← 所有人物</RouterLink>
        <div v-if="loading && !currentPerson" style="padding:2rem;text-align:center;font-weight:700">加载中...</div>
        <template v-else-if="currentPerson">
          <div class="person-name">{{ currentPerson.name }}</div>
          <div v-if="currentPerson.birth_date || currentPerson.death_date" class="person-years">
            {{ formatYear(currentPerson.birth_date) }}
            <span v-if="currentPerson.birth_date && currentPerson.death_date"> — </span>
            {{ formatYear(currentPerson.death_date) }}
          </div>
          <p v-if="currentPerson.bio" class="person-bio">{{ currentPerson.bio }}</p>
          <div v-if="currentPerson.tags?.length" class="person-tags">
            <span v-for="tag in currentPerson.tags" :key="tag" class="a-badge" style="margin-right:4px">{{ tag }}</span>
          </div>
          <div v-if="canEdit" class="person-actions">
            <ABtn size="sm" outline @click="openEditPerson">编辑人物</ABtn>
            <button class="danger-btn" @click="confirmDeletePerson">删除</button>
          </div>
        </template>
      </div>

      <!-- Locations list -->
      <div class="locations-section">
        <div class="locations-header">
          <span class="locations-title">地点轨迹</span>
          <ABtn v-if="canEdit" size="sm" @click="openAddLocation">+ 添加地点</ABtn>
        </div>

        <div v-if="!currentPerson?.locations?.length" class="locations-empty">
          暂无地点记录
        </div>

        <div
          v-for="(loc, idx) in sortedLocations"
          :key="loc.id"
          class="location-item"
          :class="{ active: selectedLocationId === loc.id }"
          @click="focusLocation(loc)"
        >
          <div class="loc-index">{{ idx + 1 }}</div>
          <div class="loc-info">
            <div class="loc-name">{{ loc.place_name }}</div>
            <div class="loc-date">{{ formatDate(loc.date) }}<span v-if="loc.end_date"> — {{ formatDate(loc.end_date) }}</span></div>
            <div v-if="loc.source" class="loc-source">来源: {{ loc.source }}</div>
            <div v-if="loc.note" class="loc-note">{{ loc.note }}</div>
          </div>
          <div v-if="canEdit" class="loc-actions">
            <button class="icon-btn" @click.stop="openEditLocation(loc)" title="编辑">✎</button>
            <button class="icon-btn danger" @click.stop="confirmDeleteLocation(loc)" title="删除">✕</button>
          </div>
        </div>
      </div>
    </aside>

    <!-- Map -->
    <main class="map-container">
      <!-- Pick mode banner -->
      <div v-if="pickingCoords" class="pick-banner">
        <span>🎯 点击地图选择坐标</span>
        <button class="pick-cancel-btn" @click="cancelPickCoords">取消</button>
      </div>
      <div ref="mapEl" class="ol-map" :class="{ 'picking-cursor': pickingCoords }" />
      <!-- Popup overlay -->
      <div ref="popupEl" class="map-popup" :class="{ visible: popupVisible }">
        <div class="popup-content">
          <button class="popup-close" @click="popupVisible = false">✕</button>
          <div class="popup-name">{{ popupData.name }}</div>
          <div class="popup-date">{{ popupData.date }}</div>
          <div v-if="popupData.source" class="popup-source">{{ popupData.source }}</div>
          <div v-if="popupData.note" class="popup-note">{{ popupData.note }}</div>
        </div>
      </div>
    </main>

    <!-- Add/Edit Location Modal -->
    <AModal v-if="showLocationForm" size="md" @close="closeLocationForm">
      <div class="a-modal-header">
        <h2 class="a-modal-title">{{ editingLocation ? '编辑地点' : '添加地点' }}</h2>
        <button class="a-modal-close" @click="closeLocationForm">✕</button>
      </div>
      <div class="a-modal-body">
        <div class="form-group">
          <label class="form-label">时间 *</label>
          <DatetimePicker v-model="locForm.date" placeholder="选择时间" />
        </div>
        <div class="form-group">
          <label class="form-label">结束时间（可选）</label>
          <DatetimePicker v-model="locForm.end_date" placeholder="离开时间（可选）" />
        </div>
        <div class="form-group">
          <label class="form-label">地名 *</label>
          <AInput v-model="locForm.place_name" placeholder="如 巴黎, 法国" />
        </div>
        <div class="form-group">
          <label class="form-label">来源 *</label>
          <AInput v-model="locForm.source" placeholder="史料来源，如《史记》卷一" />
        </div>

        <!-- Coordinates -->
        <div class="coords-section">
          <div class="coords-header">
            <span class="form-label" style="margin:0">坐标 *</span>
            <button
              type="button"
              class="pick-map-btn"
              :class="{ active: pickingCoords }"
              @click="startPickCoords"
            >
              {{ pickingCoords ? '⏳ 等待点击地图...' : '📍 在地图上选取' }}
            </button>
          </div>
          <div class="form-row" style="margin-top:0.5rem">
            <div class="form-group" style="flex:1;margin-bottom:0">
              <label class="form-label">纬度</label>
              <input
                class="a-input"
                type="number"
                step="any"
                :value="locForm.latitude"
                @input="locForm.latitude = parseFloat(($event.target as HTMLInputElement).value) || 0"
                placeholder="48.8566"
              />
            </div>
            <div class="form-group" style="flex:1;margin-bottom:0">
              <label class="form-label">经度</label>
              <input
                class="a-input"
                type="number"
                step="any"
                :value="locForm.longitude"
                @input="locForm.longitude = parseFloat(($event.target as HTMLInputElement).value) || 0"
                placeholder="2.3522"
              />
            </div>
          </div>
          <p v-if="lastPickedCoords" class="picked-coords-hint">
            ✓ 已选取: {{ lastPickedCoords.lat.toFixed(5) }}, {{ lastPickedCoords.lng.toFixed(5) }}
          </p>
        </div>

        <div class="form-group" style="margin-top:1rem">
          <label class="form-label">备注</label>
          <ATextarea v-model="locForm.note" :rows="3" placeholder="在此地的活动说明" />
        </div>
      </div>
      <template #footer>
        <div class="a-modal-footer">
          <ABtn outline @click="closeLocationForm">取消</ABtn>
          <ABtn :disabled="locSubmitting" @click="submitLocation">
            {{ locSubmitting ? '保存中...' : (editingLocation ? '保存' : '添加') }}
          </ABtn>
        </div>
      </template>
    </AModal>

    <!-- Edit Person Modal -->
    <AModal v-if="showPersonForm" size="md" @close="showPersonForm = false">
      <div class="a-modal-header">
        <h2 class="a-modal-title">编辑人物</h2>
        <button class="a-modal-close" @click="showPersonForm = false">✕</button>
      </div>
      <div class="a-modal-body">
        <div class="form-group">
          <label class="form-label">姓名 *</label>
          <AInput v-model="personForm.name" placeholder="历史人物姓名" />
        </div>
        <div class="form-row">
          <div class="form-group" style="flex:1">
            <label class="form-label">出生日期</label>
            <DatetimePicker v-model="personForm.birth_date" placeholder="出生日期" :show-time="false" />
          </div>
          <div class="form-group" style="flex:1">
            <label class="form-label">去世日期</label>
            <DatetimePicker v-model="personForm.death_date" placeholder="去世日期" :show-time="false" />
          </div>
        </div>
        <div class="form-group">
          <label class="form-label">简介</label>
          <ATextarea v-model="personForm.bio" :rows="4" placeholder="人物生平简介" />
        </div>
        <div class="form-group">
          <label class="form-label">标签（逗号分隔）</label>
          <AInput v-model="personTagsInput" placeholder="政治家, 军事家" />
        </div>
      </div>
      <template #footer>
        <div class="a-modal-footer">
          <ABtn outline @click="showPersonForm = false">取消</ABtn>
          <ABtn :disabled="personSubmitting" @click="submitEditPerson">
            {{ personSubmitting ? '保存中...' : '保存' }}
          </ABtn>
        </div>
      </template>
    </AModal>

    <!-- Confirm Delete Location -->
    <AConfirm
      :show="!!deletingLocation"
      title="删除地点记录"
      :message="deletingLocation ? `确定要删除「${deletingLocation.place_name}」的地点记录吗？` : ''"
      @confirm="doDeleteLocation"
      @cancel="deletingLocation = null"
    />

    <!-- Confirm Delete Person -->
    <AConfirm
      :show="confirmPersonDelete"
      title="删除人物"
      :message="`确定要删除「${currentPerson?.name}」及其所有地点记录吗？此操作不可撤销。`"
      @confirm="doDeletePerson"
      @cancel="confirmPersonDelete = false"
    />
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted, onUnmounted, watch } from 'vue'
import { useRoute, useRouter, RouterLink } from 'vue-router'
import Map from 'ol/Map'
import View from 'ol/View'
import TileLayer from 'ol/layer/Tile'
import OSM from 'ol/source/OSM'
import VectorLayer from 'ol/layer/Vector'
import VectorSource from 'ol/source/Vector'
import Feature from 'ol/Feature'
import Point from 'ol/geom/Point'
import LineString from 'ol/geom/LineString'
import { fromLonLat, toLonLat } from 'ol/proj'
import { Style, Circle as CircleStyle, Fill, Stroke } from 'ol/style'
import Overlay from 'ol/Overlay'
import { useTimelineStore } from '@/stores/timeline'
import { useAuthStore } from '@/stores/auth'
import type { PersonLocation } from '@/types'
import ABtn from '@/components/ui/ABtn.vue'
import AModal from '@/components/ui/AModal.vue'
import AInput from '@/components/ui/AInput.vue'
import ATextarea from '@/components/ui/ATextarea.vue'
import AConfirm from '@/components/ui/AConfirm.vue'
import DatetimePicker from '@/components/ui/DatetimePicker.vue'

const route = useRoute()
const router = useRouter()
const store = useTimelineStore()
const authStore = useAuthStore()

const { currentPerson, loading } = store

// Map refs
const mapEl = ref<HTMLElement | null>(null)
const popupEl = ref<HTMLElement | null>(null)
let olMap: Map | null = null
let vectorSource: VectorSource | null = null
let popupOverlay: Overlay | null = null

const popupVisible = ref(false)
const popupData = ref({ name: '', date: '', source: '', note: '' })
const selectedLocationId = ref<string | null>(null)

// Coordinate picking mode
const pickingCoords = ref(false)
const lastPickedCoords = ref<{ lat: number; lng: number } | null>(null)

const canEdit = computed(() => {
  return authStore.isAuthenticated &&
    (currentPerson?.user_id === authStore.user?.uuid || authStore.user?.role === 'admin')
})

const sortedLocations = computed(() => {
  return [...(currentPerson?.locations || [])].sort(
    (a, b) => new Date(a.date).getTime() - new Date(b.date).getTime()
  )
})

const formatYear = (d?: string) => (d ? d.slice(0, 4) : '')
const formatDate = (d: string) => (d ? d.slice(0, 10) : '')

// ====== Map initialization ======

const buildPointStyle = (active = false) =>
  new Style({
    image: new CircleStyle({
      radius: active ? 10 : 7,
      fill: new Fill({ color: active ? '#000' : '#fff' }),
      stroke: new Stroke({ color: '#000', width: 2 }),
    }),
  })

const lineStyle = new Style({
  stroke: new Stroke({ color: '#000', width: 2, lineDash: [4, 4] }),
})

const renderMap = () => {
  if (!vectorSource) return
  vectorSource.clear()

  const locs = sortedLocations.value
  if (!locs.length) return

  // Draw line connecting all points in order
  if (locs.length >= 2) {
    const coords = locs.map((l) => fromLonLat([l.longitude, l.latitude]))
    const line = new Feature({ geometry: new LineString(coords) })
    line.setStyle(lineStyle)
    vectorSource.addFeature(line)
  }

  // Draw points
  for (const loc of locs) {
    const point = new Feature({
      geometry: new Point(fromLonLat([loc.longitude, loc.latitude])),
      locationId: loc.id,
      placeName: loc.place_name,
      date: formatDate(loc.date) + (loc.end_date ? ' — ' + formatDate(loc.end_date) : ''),
      source: loc.source || '',
      note: loc.note,
    })
    point.setStyle(buildPointStyle(selectedLocationId.value === loc.id))
    vectorSource.addFeature(point)
  }

  // Fit view to all points
  if (olMap && locs.length) {
    const extent = vectorSource.getExtent()
    if (locs.length === 1) {
      olMap.getView().setCenter(fromLonLat([locs[0].longitude, locs[0].latitude]))
      olMap.getView().setZoom(6)
    } else {
      olMap.getView().fit(extent, { padding: [60, 60, 60, 60], maxZoom: 8 })
    }
  }
}

const initMap = () => {
  if (!mapEl.value || olMap) return

  vectorSource = new VectorSource()
  const vectorLayer = new VectorLayer({ source: vectorSource })

  popupOverlay = new Overlay({
    element: popupEl.value!,
    positioning: 'bottom-center',
    stopEvent: false,
    offset: [0, -12],
  })

  olMap = new Map({
    target: mapEl.value,
    layers: [
      new TileLayer({ source: new OSM() }),
      vectorLayer,
    ],
    overlays: [popupOverlay],
    view: new View({
      center: fromLonLat([20, 30]),
      zoom: 2,
    }),
  })

  olMap.on('click', (evt: any) => {
    // If in picking mode, capture coordinates
    if (pickingCoords.value) {
      const lonLat = toLonLat(evt.coordinate)
      const lng = lonLat[0]
      const lat = lonLat[1]
      locForm.value.latitude = parseFloat(lat.toFixed(6))
      locForm.value.longitude = parseFloat(lng.toFixed(6))
      lastPickedCoords.value = { lat, lng }
      pickingCoords.value = false
      return
    }

    // Normal click: show popup for features
    const feature = olMap!.forEachFeatureAtPixel(evt.pixel, (f) => f)
    if (feature) {
      const locId = feature.get('locationId')
      if (locId) {
        selectedLocationId.value = locId
        popupData.value = {
          name: feature.get('placeName') || '',
          date: feature.get('date') || '',
          source: feature.get('source') || '',
          note: feature.get('note') || '',
        }
        popupOverlay!.setPosition(
          (feature.getGeometry() as Point).getCoordinates()
        )
        popupVisible.value = true
        renderMap()
        return
      }
    }
    popupVisible.value = false
    selectedLocationId.value = null
    renderMap()
  })

  olMap.on('pointermove', (evt: any) => {
    if (pickingCoords.value) {
      olMap!.getTargetElement().style.cursor = 'crosshair'
      return
    }
    const hit = olMap!.hasFeatureAtPixel(evt.pixel)
    olMap!.getTargetElement().style.cursor = hit ? 'pointer' : ''
  })

  renderMap()
}

const focusLocation = (loc: PersonLocation) => {
  selectedLocationId.value = loc.id
  popupData.value = {
    name: loc.place_name,
    date: formatDate(loc.date) + (loc.end_date ? ' — ' + formatDate(loc.end_date) : ''),
    source: (loc as any).source || '',
    note: loc.note,
  }
  if (olMap && popupOverlay) {
    const coord = fromLonLat([loc.longitude, loc.latitude])
    popupOverlay.setPosition(coord)
    olMap.getView().animate({ center: coord, zoom: Math.max(olMap.getView().getZoom() || 4, 5), duration: 400 })
    popupVisible.value = true
  }
  renderMap()
}

// ====== Coordinate picking ======
const startPickCoords = () => {
  if (!showLocationForm.value) return
  pickingCoords.value = true
  // Briefly close modal focus so map is clickable — modal stays open but map click is captured first
}

const cancelPickCoords = () => {
  pickingCoords.value = false
}

// Watch for location changes to re-render map
watch(
  () => currentPerson?.locations,
  () => renderMap(),
  { deep: true }
)

// ====== Location form ======
const showLocationForm = ref(false)
const editingLocation = ref<PersonLocation | null>(null)
const locSubmitting = ref(false)
const locForm = ref({
  date: '',
  end_date: '',
  place_name: '',
  source: '',
  latitude: 0,
  longitude: 0,
  note: '',
})
const deletingLocation = ref<PersonLocation | null>(null)

const openAddLocation = () => {
  editingLocation.value = null
  lastPickedCoords.value = null
  locForm.value = { date: '', end_date: '', place_name: '', source: '', latitude: 0, longitude: 0, note: '' }
  showLocationForm.value = true
}

const openEditLocation = (loc: PersonLocation) => {
  editingLocation.value = loc
  lastPickedCoords.value = null
  locForm.value = {
    date: loc.date ? loc.date.slice(0, 16) : '',
    end_date: loc.end_date ? loc.end_date.slice(0, 16) : '',
    place_name: loc.place_name,
    source: (loc as any).source || '',
    latitude: loc.latitude,
    longitude: loc.longitude,
    note: loc.note || '',
  }
  showLocationForm.value = true
}

const closeLocationForm = () => {
  showLocationForm.value = false
  editingLocation.value = null
  pickingCoords.value = false
  lastPickedCoords.value = null
}

const submitLocation = async () => {
  if (!locForm.value.date || !locForm.value.place_name || !locForm.value.source || !currentPerson?.id) return
  locSubmitting.value = true
  try {
    if (editingLocation.value) {
      await store.updateLocation(editingLocation.value.id, locForm.value)
    } else {
      await store.addLocation(currentPerson.id, locForm.value)
    }
    closeLocationForm()
  } catch (e) {
    console.error(e)
  } finally {
    locSubmitting.value = false
  }
}

const confirmDeleteLocation = (loc: PersonLocation) => {
  deletingLocation.value = loc
}

const doDeleteLocation = async () => {
  if (!deletingLocation.value) return
  await store.deleteLocation(deletingLocation.value.id)
  deletingLocation.value = null
  popupVisible.value = false
  selectedLocationId.value = null
  renderMap()
}

// ====== Person edit/delete ======
const showPersonForm = ref(false)
const confirmPersonDelete = ref(false)
const personSubmitting = ref(false)
const personForm = ref({ name: '', bio: '', birth_date: '', death_date: '' })
const personTagsInput = ref('')

const openEditPerson = () => {
  if (!currentPerson) return
  personForm.value = {
    name: currentPerson.name,
    bio: currentPerson.bio || '',
    birth_date: currentPerson.birth_date ? currentPerson.birth_date.slice(0, 10) : '',
    death_date: currentPerson.death_date ? currentPerson.death_date.slice(0, 10) : '',
  }
  personTagsInput.value = (currentPerson.tags || []).join(', ')
  showPersonForm.value = true
}

const submitEditPerson = async () => {
  if (!currentPerson || !personForm.value.name) return
  personSubmitting.value = true
  try {
    const tags = personTagsInput.value.split(',').map((t) => t.trim()).filter(Boolean)
    await store.updatePerson(currentPerson.id, { ...personForm.value, tags })
    showPersonForm.value = false
  } catch (e) {
    console.error(e)
  } finally {
    personSubmitting.value = false
  }
}

const confirmDeletePerson = () => {
  confirmPersonDelete.value = true
}

const doDeletePerson = async () => {
  if (!currentPerson) return
  await store.deletePerson(currentPerson.id)
  router.push('/timeline/persons')
}

onMounted(async () => {
  const id = route.params.id as string
  await store.fetchPerson(id)
  initMap()
})

onUnmounted(() => {
  olMap?.setTarget(undefined)
  olMap = null
})
</script>

<style scoped>
.person-map-page {
  display: flex;
  height: calc(100vh - 64px);
  overflow: hidden;
}

/* Left panel */
.person-panel {
  width: 320px;
  flex-shrink: 0;
  border-right: 2px solid #000;
  display: flex;
  flex-direction: column;
  overflow: hidden;
}

.panel-header {
  padding: 1.25rem;
  border-bottom: 2px solid #000;
  flex-shrink: 0;
}

.back-link {
  font-size: 0.75rem;
  font-weight: 900;
  text-transform: uppercase;
  letter-spacing: 0.08em;
  color: #6b7280;
  text-decoration: none;
  display: block;
  margin-bottom: 0.75rem;
}
.back-link:hover { color: #000; text-decoration: underline; }

.person-name {
  font-size: 1.5rem;
  font-weight: 900;
  letter-spacing: -0.04em;
  line-height: 1.2;
  margin-bottom: 0.25rem;
}

.person-years {
  font-size: 0.75rem;
  font-weight: 700;
  color: #6b7280;
  margin-bottom: 0.5rem;
}

.person-bio {
  font-size: 0.8rem;
  color: #4b5563;
  line-height: 1.5;
  margin-bottom: 0.75rem;
  overflow: hidden;
  display: -webkit-box;
  -webkit-line-clamp: 4;
  -webkit-box-orient: vertical;
}

.person-tags {
  margin-bottom: 0.75rem;
}

.person-actions {
  display: flex;
  align-items: center;
  gap: 0.5rem;
}

.danger-btn {
  font-size: 0.75rem;
  font-weight: 900;
  text-transform: uppercase;
  letter-spacing: 0.06em;
  color: #ef4444;
  background: none;
  border: none;
  cursor: pointer;
  padding: 0;
}
.danger-btn:hover { text-decoration: underline; }

/* Locations list */
.locations-section {
  flex: 1;
  overflow-y: auto;
  display: flex;
  flex-direction: column;
}

.locations-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 0.75rem 1.25rem;
  border-bottom: 1px solid #e5e7eb;
  flex-shrink: 0;
}

.locations-title {
  font-size: 0.7rem;
  font-weight: 900;
  text-transform: uppercase;
  letter-spacing: 0.1em;
  color: #6b7280;
}

.locations-empty {
  padding: 2rem;
  text-align: center;
  font-size: 0.8rem;
  color: #9ca3af;
}

.location-item {
  display: flex;
  align-items: flex-start;
  gap: 0.75rem;
  padding: 0.875rem 1.25rem;
  border-bottom: 1px solid #f3f4f6;
  cursor: pointer;
  transition: background 0.1s;
}
.location-item:hover { background: #f9fafb; }
.location-item.active { background: #f0f0f0; border-left: 3px solid #000; }

.loc-index {
  flex-shrink: 0;
  width: 20px;
  height: 20px;
  border: 2px solid #000;
  border-radius: 50%;
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 0.65rem;
  font-weight: 900;
  margin-top: 2px;
}

.location-item.active .loc-index {
  background: #000;
  color: #fff;
}

.loc-info { flex: 1; min-width: 0; }
.loc-name { font-size: 0.875rem; font-weight: 700; margin-bottom: 2px; }
.loc-date { font-size: 0.7rem; color: #6b7280; font-weight: 600; }
.loc-source {
  font-size: 0.65rem;
  color: #9ca3af;
  margin-top: 1px;
  font-style: italic;
  overflow: hidden;
  white-space: nowrap;
  text-overflow: ellipsis;
}
.loc-note {
  font-size: 0.75rem;
  color: #9ca3af;
  margin-top: 2px;
  overflow: hidden;
  white-space: nowrap;
  text-overflow: ellipsis;
}

.loc-actions {
  display: flex;
  gap: 4px;
  flex-shrink: 0;
}

.icon-btn {
  background: none;
  border: none;
  cursor: pointer;
  font-size: 0.75rem;
  color: #6b7280;
  padding: 2px 4px;
  transition: color 0.15s;
}
.icon-btn:hover { color: #000; }
.icon-btn.danger:hover { color: #ef4444; }

/* Map */
.map-container {
  flex: 1;
  position: relative;
  overflow: hidden;
}

.ol-map {
  width: 100%;
  height: 100%;
}

.ol-map.picking-cursor {
  cursor: crosshair !important;
}

/* Pick mode banner */
.pick-banner {
  position: absolute;
  top: 12px;
  left: 50%;
  transform: translateX(-50%);
  z-index: 20;
  background: #000;
  color: #fff;
  padding: 8px 16px;
  font-size: 0.8rem;
  font-weight: 900;
  letter-spacing: 0.04em;
  display: flex;
  align-items: center;
  gap: 1rem;
  box-shadow: 4px 4px 0 0 rgba(0,0,0,0.3);
  pointer-events: auto;
  white-space: nowrap;
}

.pick-cancel-btn {
  background: none;
  border: 1px solid rgba(255,255,255,0.5);
  color: #fff;
  font-size: 0.7rem;
  font-weight: 900;
  text-transform: uppercase;
  letter-spacing: 0.06em;
  padding: 2px 8px;
  cursor: pointer;
  transition: background 0.1s;
}
.pick-cancel-btn:hover { background: rgba(255,255,255,0.15); }

/* Popup */
.map-popup {
  position: absolute;
  pointer-events: none;
  z-index: 10;
  opacity: 0;
  transform: translateY(4px);
  transition: opacity 0.15s, transform 0.15s;
}
.map-popup.visible {
  opacity: 1;
  transform: translateY(0);
  pointer-events: auto;
}

.popup-content {
  background: #fff;
  border: 2px solid #000;
  box-shadow: 4px 4px 0px 0px rgba(0,0,0,1);
  padding: 0.75rem 1rem;
  min-width: 160px;
  max-width: 240px;
  position: relative;
}

.popup-close {
  position: absolute;
  top: 4px;
  right: 6px;
  background: none;
  border: none;
  font-size: 0.75rem;
  cursor: pointer;
  color: #9ca3af;
  line-height: 1;
}
.popup-close:hover { color: #000; }

.popup-name { font-size: 0.875rem; font-weight: 900; margin-bottom: 3px; padding-right: 16px; }
.popup-date { font-size: 0.7rem; color: #6b7280; font-weight: 600; }
.popup-source { font-size: 0.65rem; color: #9ca3af; margin-top: 2px; font-style: italic; }
.popup-note { font-size: 0.75rem; color: #4b5563; margin-top: 4px; }

/* Location form */
.form-group { margin-bottom: 1rem; }
.form-row { display: flex; gap: 1rem; margin-bottom: 1rem; }
.form-label {
  display: block;
  font-size: 0.75rem;
  font-weight: 900;
  text-transform: uppercase;
  letter-spacing: 0.08em;
  margin-bottom: 0.4rem;
}

/* Coordinates section */
.coords-section {
  border: 2px solid #e5e7eb;
  padding: 0.75rem;
  margin-bottom: 0;
}

.coords-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  margin-bottom: 0.25rem;
}

.pick-map-btn {
  font-size: 0.7rem;
  font-weight: 900;
  text-transform: uppercase;
  letter-spacing: 0.06em;
  background: none;
  border: 2px solid #000;
  cursor: pointer;
  padding: 3px 10px;
  transition: background 0.1s, color 0.1s;
  white-space: nowrap;
}
.pick-map-btn:hover,
.pick-map-btn.active {
  background: #000;
  color: #fff;
}

.picked-coords-hint {
  font-size: 0.7rem;
  color: #059669;
  font-weight: 700;
  margin-top: 0.4rem;
  margin-bottom: 0;
}

.a-modal-header {
  display: flex; align-items: center; justify-content: space-between;
  padding: 1.25rem 1.5rem; border-bottom: 2px solid #000;
}
.a-modal-title { font-size: 1.25rem; font-weight: 900; letter-spacing: -0.03em; }
.a-modal-close {
  background: none; border: none; font-size: 1rem; font-weight: 900;
  cursor: pointer; padding: 0.25rem 0.5rem; color: #000;
}
.a-modal-body { padding: 1.5rem; overflow-y: auto; max-height: 65vh; }
.a-modal-footer {
  display: flex; justify-content: flex-end; gap: 0.75rem;
  padding: 1rem 1.5rem; border-top: 2px solid #000;
}

/* Mobile: stack layout */
@media (max-width: 767px) {
  .person-map-page { flex-direction: column; height: auto; overflow: auto; }
  .person-panel { width: 100%; border-right: none; border-bottom: 2px solid #000; max-height: 50vh; overflow-y: auto; }
  .map-container { height: 60vw; min-height: 300px; }
}
</style>
