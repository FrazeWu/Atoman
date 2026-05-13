<template>
  <div class="artist-select" ref="wrapRef">
    <label v-if="label" class="field-label">{{ label }}</label>

    <!-- Selected tags -->
    <div v-if="selected.length" class="tags">
      <span v-for="a in selected" :key="a.id" class="tag">
        {{ a.name }}
        <button class="tag-remove" type="button" :disabled="disabled" @click="remove(a.id)">✕</button>
      </span>
    </div>

    <!-- Input -->
    <div class="input-wrap">
      <input
        ref="inputRef"
        v-model="query"
        :placeholder="placeholder || '搜索艺术家'"
        class="field-input"
        :disabled="disabled"
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

      <div class="dropdown-section">
        <div class="dropdown-divider" />
        <RouterLink :to="`/music/artists/add${query.trim() ? `?name=${encodeURIComponent(query.trim())}` : ''}`" class="add-artist-link" @mousedown.prevent>
          没找到艺术家？前往新增页面
        </RouterLink>
      </div>

      <div v-if="!filtered.length && !query.trim()" class="dropdown-empty">输入关键字搜索</div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, watch, onMounted, onUnmounted } from 'vue'
import { RouterLink } from 'vue-router'
import { useApi } from '@/composables/useApi'
import type { Artist } from '@/types'

const props = defineProps<{
  modelValue: Artist[]
  label?: string
  placeholder?: string
  disabled?: boolean
}>()

const emit = defineEmits<{
  (e: 'update:modelValue', v: Artist[]): void
}>()

const api = useApi()

const wrapRef = ref<HTMLElement | null>(null)
const inputRef = ref<HTMLInputElement | null>(null)
const query = ref('')
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

const select = (a: Artist) => {
  if (props.disabled) return
  emit('update:modelValue', [...selected.value, a])
  query.value = ''
}

const remove = (id: number) => {
  if (props.disabled) return
  emit('update:modelValue', selected.value.filter(a => a.id !== id))
}

const dedupeArtistsByName = (artists: Artist[]) => {
  const seen = new Set<string>()
  return artists.filter((artist) => {
    const key = artist.name.trim().toLowerCase()
    if (seen.has(key)) return false
    seen.add(key)
    return true
  })
}

const fetchArtists = async () => {
  try {
    const res = await fetch(api.artists || '/api/artists')
    if (res.ok) {
      const d = await res.json()
      allArtists.value = dedupeArtistsByName(d.data || d || [])
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
.add-artist-link {
  display: inline-block;
  background: #000;
  color: #fff;
  text-decoration: none;
  border: 2px solid #000;
  padding: 0.45rem 0.75rem;
  font-size: 0.75rem;
  font-weight: 900;
  text-transform: uppercase;
  letter-spacing: 0.08em;
  transition: all 0.15s;
}
.add-artist-link:hover { background: #fff; color: #000; }
</style>
