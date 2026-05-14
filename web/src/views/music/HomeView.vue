<script setup lang="ts">
import { ref, computed, watch, onUnmounted } from 'vue'
import { RouterLink } from 'vue-router'
import { usePlayerStore } from '@/stores/player'
import { useAuthStore } from '@/stores/auth'
import type { Album } from '@/types'

const API_URL = import.meta.env.VITE_API_URL || '/api'
const player = usePlayerStore()
const authStore = useAuthStore()
player.fetchSongs()

interface ArtistOption { id: string; name: string }

const artists = ref<ArtistOption[]>([])
const albums = ref<Album[]>([])
const selectedArtistName = ref('')
const selectedArtistId = ref<string | null>(null)
const searchQuery = ref('')
const showDropdown = ref(false)
const protectionStatuses = ref<Map<string, any>>(new Map())
const discussionCounts = ref<Map<string, number>>(new Map())

const dropdownArtists = computed(() => {
  const q = searchQuery.value.toLowerCase()
  return q ? artists.value.filter((a) => a.name.toLowerCase().includes(q)) : artists.value
})

let searchTimer: ReturnType<typeof setTimeout>
watch(searchQuery, (q) => {
  clearTimeout(searchTimer)
  searchTimer = setTimeout(() => fetchArtists(q), 300)
})

async function fetchArtists(q = '') {
  try {
    const params = q ? `?q=${encodeURIComponent(q)}` : ''
    const res = await fetch(`${API_URL}/artists${params}`)
    if (res.ok) {
      artists.value = await res.json()
      // 首次加载时随机选一个艺术家
      if (!q && !selectedArtistName.value && artists.value.length > 0) {
        const pick = artists.value[Math.floor(Math.random() * artists.value.length)]
        selectedArtistName.value = pick.name
        selectedArtistId.value = String(pick.id)
      }
    }
  } catch (e) {
    console.error('Failed to fetch artists:', e)
  }
}

async function fetchAlbums() {
  try {
    const res = await fetch(`${API_URL}/albums`)
    if (res.ok) {
      albums.value = await res.json()
    }
  } catch (e) {
    console.error('Failed to fetch albums:', e)
  }
}

function selectArtist(name: string, id?: string) {
  selectedArtistName.value = name
  selectedArtistId.value = id || null
  searchQuery.value = ''
  showDropdown.value = false
}

function pickRandom() {
  if (artists.value.length > 0) {
    const pick = artists.value[Math.floor(Math.random() * artists.value.length)]
    selectedArtistName.value = pick.name
    selectedArtistId.value = String(pick.id)
    searchQuery.value = ''
  }
}

async function fetchProtectionStatus(albumId: string | number) {
  if (protectionStatuses.value.has(String(albumId))) {
    return protectionStatuses.value.get(String(albumId))
  }
  try {
    const res = await fetch(`${API_URL}/albums/${albumId}/protection`)
    const data = await res.json()
    const status = data.data || { protection_level: 'none' }
    protectionStatuses.value.set(String(albumId), status)
    return status
  } catch (e) {
    console.error('Failed to fetch protection status:', e)
    return { protection_level: 'none' }
  }
}

function getProtectionLabel(protectionLevel: string) {
  if (protectionLevel === 'full') return '完全保护'
  if (protectionLevel === 'semi') return '半保护'
  return ''
}

async function fetchDiscussionCount(albumId: string) {
  if (discussionCounts.value.has(albumId)) {
    return discussionCounts.value.get(albumId)
  }
  try {
    const res = await fetch(`${API_URL}/albums/${albumId}/discussions/unread-count`)
    const data = await res.json()
    const count = data.data?.unread_count || 0
    discussionCounts.value.set(albumId, count)
    return count
  } catch (e) {
    console.error('Failed to fetch discussion count:', e)
    return 0
  }
}

fetchArtists()
fetchAlbums()
onUnmounted(() => clearTimeout(searchTimer))

interface AlbumGroup {
  id: string | number
  album: string
  year: number
  release_date: string
  cover_url: string
  artist: string
  status?: string
  album_type?: string
  entry_status?: string
  songs: typeof player.songs
}

const albumGroups = computed(() => {
  const selectedAlbums = selectedArtistName.value
    ? albums.value.filter((album) => album.artists?.some((artist) => artist.name === selectedArtistName.value))
    : albums.value

  const groups = new Map<string, AlbumGroup>()
  selectedAlbums.forEach((album) => {
    const artistName = album.artists?.[0]?.name || 'Unknown Artist'
    const releaseDate = album.release_date ? String(album.release_date).slice(0, 10) : ''
    const year = album.year || (releaseDate ? Number(releaseDate.slice(0, 4)) : 0)
    groups.set(String(album.id), {
      id: album.id,
      album: album.title,
      year,
      release_date: releaseDate,
      cover_url: album.cover_url || '',
      artist: artistName,
      status: album.status,
      album_type: album.album_type,
      entry_status: album.entry_status,
      songs: [],
    })
  })

  const songs = selectedArtistName.value
    ? player.songs.filter((s) => s.artist === selectedArtistName.value)
    : [...player.songs]

  songs.forEach((song) => {
    const key = String(song.album_id || `${song.album}-${song.year}`)
    if (!groups.has(key)) {
      groups.set(key, {
        id: song.album_id,
        album: song.album,
        year: song.year,
        release_date: song.release_date,
        cover_url: song.cover_url,
        artist: song.artist,
        status: song.status,
        songs: [],
      })
    }
    const group = groups.get(key)!
    group.songs.push(song)
    if (!group.cover_url && song.cover_url) {
      group.cover_url = song.cover_url
    }
  })

  const result = Array.from(groups.values()).sort((a, b) => b.year - a.year)

  result.forEach((album) => {
    if (!protectionStatuses.value.has(String(album.id))) {
      fetchProtectionStatus(album.id)
    }
    if (!discussionCounts.value.has(String(album.id))) {
      fetchDiscussionCount(String(album.id))
    }
  })

  return result
})

const shouldShowYear = (index: number) =>
  index === 0 || albumGroups.value[index - 1].year !== albumGroups.value[index].year
</script>

<template>
  <div class="home-view">
    <!-- Header -->
    <div class="home-header">
      <h1 class="home-title">
        <template v-if="selectedArtistId">
          <RouterLink :to="`/music/artists/${selectedArtistId}`" class="artist-title-link">
            {{ selectedArtistName ? selectedArtistName.toUpperCase() : 'ATOMAN' }}
          </RouterLink>
        </template>
        <template v-else>
          {{ selectedArtistName ? selectedArtistName.toUpperCase() : 'ATOMAN' }}
        </template>
        <br />TIMELINE
      </h1>
      <p class="home-subtitle">
        An interactive archival project. Browse the complete discography of any artist.
      </p>

      <!-- Artist search -->
      <div class="search-row">
        <div class="search-wrap">
          <input
            v-model="searchQuery"
            @focus="showDropdown = true"
            @blur="showDropdown = false"
            placeholder="搜索艺术家..."
            class="search-input"
          />
          <div v-if="showDropdown" class="search-dropdown">
            <button
              v-for="artist in dropdownArtists"
              :key="artist.id"
              @mousedown.prevent="selectArtist(artist.name, String(artist.id))"
              class="search-item"
              :class="{ active: artist.name === selectedArtistName }"
            >
              {{ artist.name }}
            </button>
            <RouterLink
              to="/music/artists/new"
              @mousedown.prevent
              class="search-item search-item-new"
            >
              + 新建艺术家
            </RouterLink>
          </div>
        </div>

        <button @click="pickRandom" class="btn-random">随机</button>

        <button v-if="selectedArtistName" @click="selectedArtistName = ''; selectedArtistId = null" class="btn-all">
          全部
        </button>

        <RouterLink
          v-if="authStore.isAuthenticated"
          to="/music/contribute"
          class="btn-contribute"
        >
          贡献专辑
        </RouterLink>
      </div>
    </div>

    <!-- Timeline -->
    <div class="timeline-wrap" :style="{ minHeight: Math.max(albumGroups.length * 160, 400) + 'px' }">
      <div class="timeline-line" />

      <!-- Empty state -->
      <div v-if="albumGroups.length === 0" class="timeline-empty">
        <p>{{ player.songs.length === 0 ? '加载中...' : '该艺术家暂无专辑' }}</p>
      </div>

      <div class="albums-list">
        <div
          v-for="(albumGroup, index) in albumGroups"
          :key="`${albumGroup.album}-${albumGroup.year}`"
          class="album-row"
        >
          <!-- Year label -->
          <div v-if="shouldShowYear(index)" class="year-label">
            <span class="year-badge">{{ albumGroup.year }}</span>
          </div>

          <!-- Timeline node -->
          <div
            class="timeline-node"
            :class="{ playing: player.currentSong?.album === albumGroup.album }"
          />

          <!-- Album card -->
          <div class="album-card-wrap">
            <div class="album-card">
              <div class="album-card-inner">
                <img
                  :src="albumGroup.cover_url || 'data:image/svg+xml,%3Csvg xmlns=%22http://www.w3.org/2000/svg%22 width=%22128%22 height=%22128%22%3E%3Crect width=%22128%22 height=%22128%22 fill=%22%23111%22/%3E%3C/svg%3E'"
                  class="album-cover"
                  :alt="albumGroup.album"
                />
                <div class="album-info">
                  <h3 class="album-title">{{ albumGroup.album }}</h3>
                  <p class="album-artist">{{ albumGroup.artist }}</p>
                  <p class="album-date">{{ albumGroup.release_date }}</p>
                  <p class="album-tracks">
                    {{ albumGroup.songs.length }}
                    {{ albumGroup.songs.length === 1 ? 'track' : 'tracks' }}
                  </p>
                  <div class="album-badges-row">
                    <span v-if="albumGroup.album_type" class="album-type-badge">{{ albumGroup.album_type.toUpperCase() }}</span>
                    <span v-if="albumGroup.entry_status === 'confirmed'" class="entry-badge entry-confirmed">已确认</span>
                    <span v-else-if="albumGroup.entry_status === 'disputed'" class="entry-badge entry-disputed">争议</span>
                  </div>
                  <p class="album-status" :class="albumGroup.status === 'closed' ? 'album-status-closed' : 'album-status-open'">
                    状态：{{ albumGroup.status === 'closed' ? '关闭' : '开放' }}
                  </p>

                  <div class="album-actions">
                    <button
                      v-if="albumGroup.songs.length"
                      @click="player.playSong(albumGroup.songs[0])"
                      class="btn-play"
                    >
                      ▶ 播放
                    </button>
                    <RouterLink
                      :to="`/music/albums/${albumGroup.id}`"
                      class="btn-detail"
                    >
                      详情
                    </RouterLink>
                  </div>

                  <!-- Protection badge -->
                  <div v-if="protectionStatuses.get(String(albumGroup.id)) && getProtectionLabel(protectionStatuses.get(String(albumGroup.id))?.protection_level)" class="album-protection">
                    <span
                      class="protection-badge"
                      :class="`protection-${protectionStatuses.get(String(albumGroup.id))?.protection_level}`"
                    >
                      🔒 {{ getProtectionLabel(protectionStatuses.get(String(albumGroup.id))?.protection_level) }}
                    </span>
                  </div>
                </div>
                </div>
              
            </div>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<style scoped>
.home-view { position: relative; padding-top: 3rem; padding-bottom: 12rem; }
.home-header { max-width: 72rem; margin: 0 auto; padding: 0 2rem; margin-bottom: 3rem; }
.home-title {
  font-size: 3rem;
  font-weight: 900;
  font-style: italic;
  letter-spacing: -0.05em;
  border-left: 8px solid var(--a-color-fg);
  padding-left: 1.5rem;
  line-height: 1.1;
  margin: 0 0 0.75rem;
}
.home-subtitle { color: var(--a-color-muted); max-width: 32rem; font-size: 0.875rem; margin: 0 0 1.5rem; }
.search-row { display: flex; align-items: center; gap: 0.75rem; flex-wrap: wrap; }
.search-wrap { position: relative; }
.search-input {
  border: 2px solid var(--a-color-fg);
  padding: 0.5rem 1rem;
  font-size: 0.875rem;
  font-weight: 500;
  outline: none;
  transition: box-shadow 0.2s;
  width: 240px;
}
.search-input:focus { box-shadow: 5px 5px 0px 0px rgba(0,0,0,1); }
.search-dropdown {
  position: absolute;
  top: 100%;
  left: 0;
  z-index: 50;
  width: 240px;
  background: var(--a-color-bg);
  border: 2px solid var(--a-color-fg);
  border-top: none;
  max-height: 208px;
  overflow-y: auto;
}
.search-item {
  display: block;
  width: 100%;
  text-align: left;
  padding: 0.5rem 1rem;
  font-size: 0.875rem;
  font-weight: 500;
  background: var(--a-color-bg);
  border: none;
  cursor: pointer;
  transition: all 0.1s;
}
.search-item:hover, .search-item.active { background: var(--a-color-fg); color: var(--a-color-bg); }
.btn-random, .btn-all {
  border: 2px solid var(--a-color-fg);
  padding: 0.5rem 1rem;
  font-size: 0.75rem;
  font-weight: 900;
  text-transform: uppercase;
  letter-spacing: 0.1em;
  background: var(--a-color-bg);
  cursor: pointer;
  transition: all 0.2s;
}
.btn-random:hover, .btn-all:hover { background: var(--a-color-fg); color: var(--a-color-bg); }
.btn-all { border-width: 1px; color: var(--a-color-muted); }
.btn-contribute {
  border: 2px solid var(--a-color-fg);
  padding: 0.5rem 1rem;
  font-size: 0.75rem;
  font-weight: 900;
  text-transform: uppercase;
  letter-spacing: 0.1em;
  background: var(--a-color-fg);
  color: var(--a-color-bg);
  cursor: pointer;
  transition: all 0.2s;
  text-decoration: none;
  display: inline-block;
}
.btn-contribute:hover { background: var(--a-color-bg); color: var(--a-color-fg); }
.album-status {
  font-size: 0.75rem;
  font-weight: 800;
  letter-spacing: 0.06em;
  text-transform: uppercase;
  margin: 0.25rem 0 0.75rem;
}
.album-status-open { color: var(--a-color-success); }
.album-status-closed { color: var(--a-color-danger); }
.search-item-new {
  display: block;
  width: 100%;
  text-align: left;
  padding: 0.5rem 1rem;
  font-size: 0.875rem;
  font-weight: 900;
  background: var(--a-color-surface);
  border-top: 1px solid var(--a-color-disabled-border);
  color: var(--a-color-fg);
  text-decoration: none;
  cursor: pointer;
  transition: all 0.1s;
}
.search-item-new:hover { background: var(--a-color-fg); color: var(--a-color-bg); }

/* Timeline */
.timeline-wrap {
  position: relative;
  max-width: 72rem;
  margin: 0 auto;
  padding: 0 2rem;
}
.timeline-line {
  position: absolute;
  left: calc(33.333% + 2rem);
  top: 0;
  bottom: 0;
  width: 4px;
  background: var(--a-color-fg);
  z-index: 0;
}
.timeline-empty {
  position: relative;
  z-index: 10;
  margin-left: calc(33.333% + 2rem);
  padding-top: 4rem;
  color: var(--a-color-muted-soft);
  font-weight: 500;
}
.albums-list { display: flex; flex-direction: column; gap: 6rem; position: relative; z-index: 10; }
.album-row { position: relative; display: flex; align-items: center; }
.year-label {
  position: absolute;
  left: 33.333%;
  transform: translateX(-50%);
  top: -3rem;
  z-index: 20;
}
.year-badge {
  background: var(--a-color-fg);
  color: var(--a-color-bg);
  padding: 0.25rem 1rem;
  font-size: 0.875rem;
  font-weight: 700;
  letter-spacing: 0.1em;
}
.timeline-node {
  position: absolute;
  left: 33.333%;
  transform: translateX(-50%);
  width: 24px;
  height: 24px;
  border-radius: 9999px;
  border: 4px solid var(--a-color-bg);
  background: var(--a-color-fg);
  z-index: 20;
  transition: transform 0.2s;
}
.timeline-node.playing { transform: translateX(-50%) scale(1.5); }
.album-card-wrap {
  margin-left: calc(33.333% + 2rem);
  width: calc(66.666% - 2rem);
}
.album-card {
  background: var(--a-color-bg);
  border: 2px solid var(--a-color-fg);
  padding: 1.5rem;
  transition: box-shadow 0.3s;
}
.album-card:hover { box-shadow: 10px 10px 0px 0px rgba(0,0,0,1); }
.album-card-inner { display: flex; gap: 1.5rem; }
.album-cover {
  width: 128px;
  height: 128px;
  border: 2px solid var(--a-color-fg);
  object-fit: cover;
  flex-shrink: 0;
}
.album-info { display: flex; flex-direction: column; justify-content: center; flex: 1; }
.album-title {
  font-size: 1.5rem;
  font-weight: 900;
  letter-spacing: -0.03em;
  line-height: 1.2;
  margin: 0 0 0.25rem;
}
.album-artist { font-size: 0.875rem; font-weight: 700; color: #4b5563; margin: 0 0 0.25rem; }
.album-date { font-size: 0.75rem; color: var(--a-color-muted); margin: 0 0 0.25rem; }
.album-tracks { font-size: 0.75rem; color: var(--a-color-muted-soft); margin: 0; }
.album-actions { display: flex; gap: 0.75rem; margin-top: 1rem; }
.btn-play, .btn-detail {
  border: 2px solid var(--a-color-fg);
  padding: 0.5rem 1rem;
  font-size: 0.75rem;
  font-weight: 900;
  text-transform: uppercase;
  letter-spacing: 0.1em;
  background: var(--a-color-bg);
  cursor: pointer;
  transition: all 0.2s;
  text-decoration: none;
  color: var(--a-color-fg);
  display: inline-block;
}
.btn-play:hover, .btn-detail:hover { background: var(--a-color-fg); color: var(--a-color-bg); }

/* Protection badge */
.album-protection {
  margin-top: 0.75rem;
  display: flex;
  gap: 0.5rem;
}

.protection-badge {
  display: inline-flex;
  align-items: center;
  padding: 0.25rem 0.75rem;
  font-size: 0.625rem;
  font-weight: 700;
  text-transform: uppercase;
  letter-spacing: 0.05em;
  border-radius: 2px;
}

.protection-full {
  background: #dc2626;
  color: var(--a-color-bg);
}

.protection-semi {
  background: #facc15;
  color: var(--a-color-fg);
}
.artist-title-link {
  color: inherit;
  text-decoration: none;
}
.artist-title-link:hover {
  text-decoration: underline;
}
.album-badges-row {
  display: flex;
  gap: 0.375rem;
  margin: 0.375rem 0 0.375rem;
  flex-wrap: wrap;
}
.album-type-badge {
  font-size: 0.5rem;
  font-weight: 900;
  letter-spacing: 0.1em;
  border: 1px solid var(--a-color-fg);
  padding: 0.125rem 0.375rem;
}
.entry-badge {
  font-size: 0.5rem;
  font-weight: 900;
  letter-spacing: 0.1em;
  padding: 0.125rem 0.375rem;
  border: 1px solid;
}
.entry-confirmed { border-color: var(--a-color-success); color: var(--a-color-success); }
.entry-disputed { border-color: var(--a-color-danger); color: var(--a-color-danger); }
</style>
