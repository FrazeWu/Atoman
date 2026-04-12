<script setup lang="ts">
import { ref, computed, watch, onUnmounted } from 'vue'
import { RouterLink } from 'vue-router'
import { usePlayerStore } from '@/stores/player'

const API_URL = import.meta.env.VITE_API_URL || '/api'
const player = usePlayerStore()
player.fetchSongs()

interface ArtistOption { id: number; name: string }

const artists = ref<ArtistOption[]>([])
const selectedArtistName = ref('')
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
    if (res.ok) artists.value = await res.json()
  } catch (e) {
    console.error('Failed to fetch artists:', e)
  }
}

function selectArtist(name: string) {
  selectedArtistName.value = name
  searchQuery.value = ''
  showDropdown.value = false
}

function pickRandom() {
  if (artists.value.length > 0) {
    const pick = artists.value[Math.floor(Math.random() * artists.value.length)]
    selectedArtistName.value = pick.name
    searchQuery.value = ''
  }
}

async function fetchProtectionStatus(albumId: number) {
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

const stopOnce = watch(
  () => player.songs.length,
  (len) => {
    if (len > 0 && !selectedArtistName.value) {
      const names = [...new Set(player.songs.map((s) => s.artist))]
      selectedArtistName.value = names[Math.floor(Math.random() * names.length)]
      stopOnce()
    }
  },
)

fetchArtists()
onUnmounted(() => clearTimeout(searchTimer))

interface AlbumGroup {
  id: number
  album: string
  year: number
  release_date: string
  cover_url: string
  artist: string
  songs: typeof player.songs
}

const albumGroups = computed(() => {
  const songs = selectedArtistName.value
    ? player.songs.filter((s) => s.artist === selectedArtistName.value)
    : [...player.songs]

  const groups = new Map<string, AlbumGroup>()
  songs.forEach((song) => {
    const key = `${song.album}-${song.year}`
    if (!groups.has(key)) {
      groups.set(key, {
        id: song.album_id, // Use album_id (UUID) from song
        album: song.album,
        year: song.year,
        release_date: song.release_date,
        cover_url: song.cover_url,
        artist: song.artist,
        songs: [],
      })
    }
    groups.get(key)!.songs.push(song)
  })
  
  const result = Array.from(groups.values()).sort((a, b) => b.year - a.year)
  
  // Fetch protection status for all albums
  result.forEach((album) => {
    if (!protectionStatuses.value.has(String(album.id))) {
      fetchProtectionStatus(album.id)
    }
  })

  // Fetch discussion counts for all albums
  result.forEach((album) => {
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
        {{ selectedArtistName ? selectedArtistName.toUpperCase() : 'ATOMAN' }}<br />TIMELINE
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
          <div v-if="showDropdown && dropdownArtists.length > 0" class="search-dropdown">
            <button
              v-for="artist in dropdownArtists"
              :key="artist.id"
              @mousedown.prevent="selectArtist(artist.name)"
              class="search-item"
              :class="{ active: artist.name === selectedArtistName }"
            >
              {{ artist.name }}
            </button>
          </div>
        </div>

        <button @click="pickRandom" class="btn-random">随机</button>

        <button v-if="selectedArtistName" @click="selectedArtistName = ''" class="btn-all">
          全部
        </button>
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

                  <div class="album-actions">
                    <button @click="player.playSong(albumGroup.songs[0])" class="btn-play">
                      ▶ 播放
                    </button>
                    <RouterLink
                      :to="`/music/artists/${encodeURIComponent(albumGroup.artist.replace(/ /g, '_'))}/albums/${encodeURIComponent(albumGroup.album.replace(/ /g, '_'))}`"
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
  border-left: 8px solid #000;
  padding-left: 1.5rem;
  line-height: 1.1;
  margin: 0 0 0.75rem;
}
.home-subtitle { color: #6b7280; max-width: 32rem; font-size: 0.875rem; margin: 0 0 1.5rem; }
.search-row { display: flex; align-items: center; gap: 0.75rem; flex-wrap: wrap; }
.search-wrap { position: relative; }
.search-input {
  border: 2px solid #000;
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
  background: #fff;
  border: 2px solid #000;
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
  background: #fff;
  border: none;
  cursor: pointer;
  transition: all 0.1s;
}
.search-item:hover, .search-item.active { background: #000; color: #fff; }
.btn-random, .btn-all {
  border: 2px solid #000;
  padding: 0.5rem 1rem;
  font-size: 0.75rem;
  font-weight: 900;
  text-transform: uppercase;
  letter-spacing: 0.1em;
  background: #fff;
  cursor: pointer;
  transition: all 0.2s;
}
.btn-random:hover, .btn-all:hover { background: #000; color: #fff; }
.btn-all { border-width: 1px; color: #6b7280; }

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
  background: #000;
  z-index: 0;
}
.timeline-empty {
  position: relative;
  z-index: 10;
  margin-left: calc(33.333% + 2rem);
  padding-top: 4rem;
  color: #9ca3af;
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
  background: #000;
  color: #fff;
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
  border: 4px solid #fff;
  background: #000;
  z-index: 20;
  transition: transform 0.2s;
}
.timeline-node.playing { transform: translateX(-50%) scale(1.5); }
.album-card-wrap {
  margin-left: calc(33.333% + 2rem);
  width: calc(66.666% - 2rem);
}
.album-card {
  background: #fff;
  border: 2px solid #000;
  padding: 1.5rem;
  transition: box-shadow 0.3s;
}
.album-card:hover { box-shadow: 10px 10px 0px 0px rgba(0,0,0,1); }
.album-card-inner { display: flex; gap: 1.5rem; }
.album-cover {
  width: 128px;
  height: 128px;
  border: 2px solid #000;
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
.album-date { font-size: 0.75rem; color: #6b7280; margin: 0 0 0.25rem; }
.album-tracks { font-size: 0.75rem; color: #9ca3af; margin: 0; }
.album-actions { display: flex; gap: 0.75rem; margin-top: 1rem; }
.btn-play, .btn-detail {
  border: 2px solid #000;
  padding: 0.5rem 1rem;
  font-size: 0.75rem;
  font-weight: 900;
  text-transform: uppercase;
  letter-spacing: 0.1em;
  background: #fff;
  cursor: pointer;
  transition: all 0.2s;
  text-decoration: none;
  color: #000;
  display: inline-block;
}
.btn-play:hover, .btn-detail:hover { background: #000; color: #fff; }

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
  color: #fff;
}

.protection-semi {
  background: #facc15;
  color: #000;
}
</style>
