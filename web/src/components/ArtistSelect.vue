<template>
  <div class="artist-select" ref="wrapRef">
    <label class="field-label">{{ label || '艺术家' }}</label>

    <!-- Selected tags -->
    <div v-if="selected.length" class="tags">
      <span v-for="a in selected" :key="a.id" class="tag">
        {{ a.name }}
        <button class="tag-remove" @click="remove(a.id)">✕</button>
      </span>
    </div>

    <!-- Input -->
    <div class="input-wrap">
      <input
        ref="inputRef"
        v-model="query"
        :placeholder="placeholder || '搜索或新增艺术家'"
        class="field-input"
        @focus="open = true"
        @keydown.esc="open = false"
      />
    </div>

    <!-- Dropdown -->
    <div v-if="open" class="dropdown">
      <!-- Filtered results -->
      <div
        v-for="a in filtered"
        :key="a.id"
        class="dropdown-item"
        @mousedown.prevent="select(a)"
      >
        {{ a.name }}
      </div>

      <!-- Create new -->
      <div v-if="query.trim() && !exactMatch" class="dropdown-section">
        <div class="dropdown-divider" />
        <p class="dropdown-hint">新增艺术家</p>
        <div class="new-form">
          <input v-model="newName" placeholder="艺术家姓名" class="field-input field-input-sm" />
          <button class="btn-create" @mousedown.prevent="createNew">添加</button>
        </div>
      </div>

      <div v-if="!filtered.length && !query.trim()" class="dropdown-empty">输入搜索或新增</div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, watch, onMounted, onUnmounted } from 'vue'
import { useApi } from '@/composables/useApi'
import { useAuthStore } from '@/stores/auth'
import type { Artist } from '@/types'

const props = defineProps<{
  modelValue: Artist[]
  label?: string
  placeholder?: string
}>()

const emit = defineEmits<{
  (e: 'update:modelValue', v: Artist[]): void
}>()

const api = useApi()
const authStore = useAuthStore()

const wrapRef = ref<HTMLElement | null>(null)
const inputRef = ref<HTMLInputElement | null>(null)
const query = ref('')
const newName = ref('')
const open = ref(false)
const allArtists = ref<Artist[]>([])

const selected = computed(() => props.modelValue)

const filtered = computed(() => {
  const q = query.value.toLowerCase()
  const selectedIds = new Set(selected.value.map(a => a.id))
  return allArtists.value
    .filter(a => !selectedIds.has(a.id) && a.name.toLowerCase().includes(q))
    .slice(0, 8)
})

const exactMatch = computed(() =>
  allArtists.value.some(a => a.name.toLowerCase() === query.value.trim().toLowerCase())
)

const select = (a: Artist) => {
  emit('update:modelValue', [...selected.value, a])
  query.value = ''
}

const remove = (id: number) => {
  emit('update:modelValue', selected.value.filter(a => a.id !== id))
}

const createNew = async () => {
  const name = (newName.value || query.value).trim()
  if (!name) return
  try {
    const res = await fetch('/api/artists', {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
        Authorization: `Bearer ${authStore.token}`
      },
      body: JSON.stringify({ name })
    })
    if (res.ok) {
      const d = await res.json()
      const artist: Artist = d.data || d
      allArtists.value.push(artist)
      select(artist)
      newName.value = ''
    }
  } catch (e) { console.error(e) }
}

const fetchArtists = async () => {
  try {
    const res = await fetch(api.artists || '/api/artists')
    if (res.ok) {
      const d = await res.json()
      allArtists.value = d.data || d || []
    }
  } catch (e) { console.error(e) }
}

const clickOutside = (e: MouseEvent) => {
  if (wrapRef.value && !wrapRef.value.contains(e.target as Node)) open.value = false
}

onMounted(() => {
  fetchArtists()
  document.addEventListener('click', clickOutside)
})
onUnmounted(() => document.removeEventListener('click', clickOutside))
watch(query, (v) => { if (v) open.value = true })
</script>

<style scoped>
.artist-select { position: relative; }
.field-label {
  display: block;
  font-size: 0.75rem;
  font-weight: 900;
  text-transform: uppercase;
  letter-spacing: 0.1em;
  color: #6b7280;
  margin-bottom: 0.5rem;
}
.tags {
  display: flex;
  flex-wrap: wrap;
  gap: 0.5rem;
  margin-bottom: 0.5rem;
}
.tag {
  display: flex;
  align-items: center;
  gap: 0.25rem;
  background: #000;
  color: #fff;
  font-size: 0.75rem;
  font-weight: 700;
  padding: 0.25rem 0.5rem;
}
.tag-remove {
  background: none;
  border: none;
  color: #fff;
  cursor: pointer;
  font-size: 0.7rem;
  padding: 0;
  line-height: 1;
}
.input-wrap { position: relative; }
.field-input {
  width: 100%;
  background: #fff;
  border: 2px solid #000;
  padding: 0.75rem 1rem;
  font-size: 0.875rem;
  font-weight: 500;
  outline: none;
  transition: box-shadow 0.2s;
  box-sizing: border-box;
}
.field-input:focus { box-shadow: 5px 5px 0px 0px rgba(0,0,0,1); }
.field-input-sm { padding: 0.5rem 0.75rem; }
.dropdown {
  position: absolute;
  left: 0;
  right: 0;
  top: calc(100% + 2px);
  background: #fff;
  border: 2px solid #000;
  box-shadow: 4px 4px 0px 0px rgba(0,0,0,1);
  z-index: 50;
  max-height: 240px;
  overflow-y: auto;
}
.dropdown-item {
  padding: 0.625rem 1rem;
  font-size: 0.875rem;
  font-weight: 500;
  cursor: pointer;
  transition: background 0.1s;
}
.dropdown-item:hover { background: #f3f4f6; }
.dropdown-empty { padding: 1rem; color: #9ca3af; font-size: 0.875rem; text-align: center; }
.dropdown-section { padding: 0.5rem; }
.dropdown-divider { height: 1px; background: #e5e7eb; margin: 0.25rem 0; }
.dropdown-hint { font-size: 0.75rem; font-weight: 900; text-transform: uppercase; letter-spacing: 0.1em; color: #9ca3af; margin: 0 0 0.5rem; }
.new-form { display: flex; gap: 0.5rem; }
.btn-create {
  background: #000;
  color: #fff;
  border: 2px solid #000;
  padding: 0.5rem 0.75rem;
  font-size: 0.75rem;
  font-weight: 900;
  cursor: pointer;
  white-space: nowrap;
  transition: all 0.15s;
}
.btn-create:hover { background: #fff; color: #000; }
</style>
