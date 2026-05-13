<script setup lang="ts">
import { computed, ref, onMounted } from 'vue';
import { useRoute, RouterLink } from 'vue-router';
import { usePlayerStore } from '@/stores/player';
import { useAuthStore } from '@/stores/auth';

const route = useRoute();
const player = usePlayerStore();
const authStore = useAuthStore();
const API_URL = import.meta.env.VITE_API_URL || '/api';

// Get album id (UUID or name) and optional artist name from route params
const albumParam = decodeURIComponent(route.params.albumId as string);
const artistName = route.params.artistName ? decodeURIComponent(route.params.artistName as string) : null;

// UUID pattern check
const isUuid = /^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$/i.test(albumParam);

// Find album UUID by matching songs — support both UUID (album_id) and name (album title)
const albumUuid = computed(() => {
  const matchingSong = player.songs.find(song => {
    if (isUuid) {
      return String(song.album_id) === albumParam;
    }
    const albumMatch = song.album.toLowerCase() === albumParam.toLowerCase();
    if (artistName) {
      return albumMatch && song.artist.toLowerCase() === artistName.toLowerCase();
    }
    return albumMatch;
  });
  return matchingSong ? String(matchingSong.album_id) : null;
});

// Filter songs by album UUID
const albumSongs = computed(() => {
  const uuid = albumUuid.value;
  if (!uuid) return [];
  
  let songs = player.songs.filter(song => String(song.album_id) === uuid);
  
  // Sort by track number and release date
  return songs.sort((a, b) => {
    if (a.year !== b.year) return a.year - b.year;
    return (a.release_date || '').localeCompare(b.release_date || '');
  });
});

const albumInfo = computed(() => {
  if (albumSongs.value.length === 0) return null;
  const firstSong = albumSongs.value[0];
  return {
    title: firstSong.album,
    artist: firstSong.artist,
    year: firstSong.year,
    releaseDate: firstSong.release_date,
    cover: firstSong.cover_url,
    trackCount: albumSongs.value.length,
    status: firstSong.status,
    uuid: albumUuid.value
  };
});

// Fetch album protection status
const protection = ref<any>(null)
const loadingProtection = ref(false)

const fetchProtection = async () => {
  if (!albumUuid.value) return
  loadingProtection.value = true
  try {
    const res = await fetch(`${API_URL}/albums/${albumUuid.value}/protection`, {
      headers: authStore.token ? { Authorization: `Bearer ${authStore.token}` } : {}
    })
    const data = await res.json()
    protection.value = data.data
  } catch (e) {
    protection.value = { protection_level: 'none' }
  } finally {
    loadingProtection.value = false
  }
}

const protectionLabel = computed(() => {
  if (!protection.value) return ''
  const level = protection.value.protection_level
  if (level === 'full') return '完全保护'
  if (level === 'semi') return '半保护'
  return ''
})

const canEdit = computed(() => {
  if (!authStore.isAuthenticated) return false
  if (authStore.user?.role === 'admin') return true
  if (protection.value?.protection_level === 'full') return false
  return true
})

// Fetch album discussion count
const discussionCount = ref<number>(0)

const fetchDiscussionCount = async () => {
  if (!albumUuid.value) return
  try {
    const res = await fetch(`${API_URL}/albums/${albumUuid.value}/discussions/unread-count`, {
      headers: authStore.token ? { Authorization: `Bearer ${authStore.token}` } : {}
    })
    const data = await res.json()
    discussionCount.value = data.data?.unread_count || 0
  } catch (e) {
    discussionCount.value = 0
  }
}

const markDiscussionAsRead = async () => {
  // This is called when navigating to discussion view
  discussionCount.value = 0
}

onMounted(async () => {
  await player.fetchSongs();
  fetchProtection();
  fetchDiscussionCount();
  fetchAlbumDetails();
});

// Fetch full album details (entry_status, album_type)
const albumDetails = ref<any>(null)
const fetchAlbumDetails = async () => {
  if (!albumUuid.value) return
  try {
    const res = await fetch(`${API_URL}/albums/${albumUuid.value}`, {
      headers: authStore.token ? { Authorization: `Bearer ${authStore.token}` } : {}
    })
    if (res.ok) {
      const data = await res.json()
      albumDetails.value = data
    }
  } catch (e) { /* ignore */ }
}

const changeEntryStatus = async (status: string) => {
  if (!albumUuid.value) return
  try {
    await fetch(`${API_URL}/albums/${albumUuid.value}/status`, {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
        Authorization: `Bearer ${authStore.token}`,
      },
      body: JSON.stringify({ status }),
    })
    if (albumDetails.value) albumDetails.value.entry_status = status
  } catch (e) { console.error('Failed to change status:', e) }
}
</script>

<template>
  <div class="page-container">
    <RouterLink to="/music" class="back-link">← 返回时间线</RouterLink>

    <div v-if="!albumInfo" class="error-message">专辑未找到</div>
    <div v-else-if="albumInfo" class="album-content">
      <!-- Album header -->
      <div class="album-header">
        <img
          :src="albumInfo.cover || 'data:image/svg+xml,%3Csvg xmlns=%22http://www.w3.org/2000/svg%22 width=%22300%22 height=%22300%22 fill=%22%23000%22/%3E%3C/svg%3E'"
          class="album-cover"
          :alt="albumInfo.title"
        />
        <div class="album-info">
          <h1 class="album-title">
            {{ albumInfo.title }}
            <span class="pending-badge" :class="albumInfo.status === 'closed' ? 'closed-badge' : 'open-badge'">
              {{ albumInfo.status === 'closed' ? '关闭' : '开放' }}
            </span>
            <span v-if="albumDetails?.album_type" class="type-badge">{{ albumDetails.album_type.toUpperCase() }}</span>
            <span v-if="albumDetails?.entry_status === 'confirmed'" class="entry-badge entry-confirmed">已确认</span>
            <span v-else-if="albumDetails?.entry_status === 'disputed'" class="entry-badge entry-disputed">争议</span>
          </h1>
          <p class="album-artist">{{ albumInfo.artist }}</p>
          <p class="album-tracks">{{ albumInfo.trackCount }} {{ albumInfo.trackCount === 1 ? 'track' : 'tracks' }}</p>

          <!-- Admin status actions -->
          <div v-if="authStore.user?.role === 'admin' && albumDetails" class="admin-entry-actions">
            <button v-if="albumDetails.entry_status !== 'confirmed'" @click="changeEntryStatus('confirmed')" class="entry-action-btn entry-confirm">确认条目</button>
            <button v-if="albumDetails.entry_status === 'confirmed'" @click="changeEntryStatus('disputed')" class="entry-action-btn entry-dispute">设为争议</button>
            <button v-if="albumDetails.entry_status === 'disputed'" @click="changeEntryStatus('open')" class="entry-action-btn entry-open">重新开放</button>
          </div>

          <div class="album-actions">
            <button @click="player.playSong(albumSongs[0])" class="btn-play-album">
              ▶ 播放专辑
            </button>
            <RouterLink
              v-if="authStore.isAuthenticated && albumInfo?.artist"
              :to="`/music/artists/${encodeURIComponent(albumInfo.artist)}/albums/${encodeURIComponent(albumInfo.title)}/edit`"
              class="btn-edit-album"
            >
              编辑专辑
            </RouterLink>
          </div>

          <div class="wiki-links">
            <RouterLink
              :to="`/music/albums/${albumUuid}/history`"
              class="wiki-link"
            >
              📖 修订历史
            </RouterLink>
            <RouterLink
              :to="`/music/albums/${albumUuid}/discussion`"
              class="wiki-link"
              @click="markDiscussionAsRead"
            >
              💬 讨论
              <span v-if="discussionCount > 0" class="discussion-count-badge">
                {{ discussionCount }}
              </span>
            </RouterLink>
          </div>
          <div v-if="protectionLabel" class="wiki-meta">
            <span class="protection-badge" :class="[`protection-${protection?.protection_level}`]">
              🔒 {{ protectionLabel }}
            </span>
            <span v-if="!canEdit" class="status-badge status-draft">
              仅管理员可编辑
            </span>
          </div>
        </div>
      </div>

      <!-- Track list -->
      <div class="tracklist-container">
        <div class="tracklist-header">
          <h2 class="tracklist-heading">歌曲列表</h2>
        </div>
        <div class="tracklist-body">
          <div
            v-for="(song, index) in albumSongs"
            :key="song.id"
            class="track-row"
          >
            <span class="track-num">{{ String(index + 1).padStart(2, '0') }}</span>
            <div class="track-title">
              <h3>{{ song.title }}</h3>
              <div class="track-wiki-links">
                <RouterLink :to="`/music/songs/${song.id}/history`">历史</RouterLink>
                <RouterLink :to="`/music/songs/${song.id}/discussion`">讨论</RouterLink>
              </div>
            </div>
            <button
              @click="player.playSong(song)"
              class="btn-track-play"
              :class="{ 'btn-track-active': player.currentSong?.id === song.id && player.isPlaying }"
            >
              {{ (player.currentSong?.id === song.id && player.isPlaying) ? '⏸ 暂停' : '▶ 播放' }}
            </button>
          </div>
        </div>
      </div>
    </div>

    <div v-else class="not-found">
      <p>专辑未找到</p>
    </div>
  </div>
</template>

<style scoped>
.page-container {
  max-width: 1024px;
  margin: 0 auto;
  padding: 5rem 2rem;
}

.back-link {
  display: inline-flex;
  align-items: center;
  gap: 0.5rem;
  margin-bottom: 2rem;
  font-weight: 700;
  text-decoration: none;
  color: #000;
  transition: opacity 0.2s;
}
.back-link:hover { text-decoration: underline; }

.album-content {
  display: flex;
  flex-direction: column;
  gap: 2rem;
}

.album-header {
  display: flex;
  gap: 2rem;
  align-items: flex-start;
}

.album-cover {
  width: 16rem;
  height: 16rem;
  border: 4px solid #000;
  object-fit: cover;
  flex-shrink: 0;
  box-shadow: 15px 15px 0px 0px rgba(0,0,0,1);
  transition: box-shadow 0.2s;
}
.album-cover:hover { box-shadow: none; }

.album-info { flex: 1; }

.album-title {
  font-size: 3rem;
  font-weight: 900;
  letter-spacing: -0.05em;
  margin: 0 0 0.5rem 0;
  display: flex;
  align-items: center;
  flex-wrap: wrap;
  gap: 0.75rem;
}

.pending-badge {
  display: inline-block;
  font-size: 0.875rem;
  font-weight: 700;
  padding: 0.25rem 0.75rem;
  border-radius: 9999px;
}

.open-badge {
  background: #dcfce7;
  color: #166534;
}

.closed-badge {
  background: #fee2e2;
  color: #991b1b;
}

.album-artist {
  font-size: 1.25rem;
  font-weight: 700;
  color: #4b5563;
  margin: 0 0 0.25rem 0;
}

.album-tracks {
  font-size: 1.125rem;
  color: #6b7280;
  margin: 0 0 1.5rem 0;
}

.album-actions {
  display: flex;
  gap: 1rem;
  flex-wrap: wrap;
}

.wiki-links {
  display: flex;
  gap: 0.75rem;
  margin-top: 0.5rem;
}

.wiki-link {
  display: inline-flex;
  align-items: center;
  gap: 0.25rem;
  padding: 0.5rem 1rem;
  border: 2px solid #000;
  background: #fff;
  color: #000;
  font-weight: 700;
  font-size: 0.75rem;
  text-decoration: none;
  text-transform: uppercase;
  letter-spacing: 0.05em;
}

.wiki-link:hover {
  background: #000;
  color: #fff;
}

.wiki-meta {
  display: flex;
  gap: 0.75rem;
  flex-wrap: wrap;
  padding: 1rem;
  background: #f9fafb;
  border: 2px solid #000;
}

.protection-badge,
.status-badge {
  display: inline-flex;
  align-items: center;
  padding: 0.375rem 0.75rem;
  font-weight: 700;
  font-size: 0.75rem;
  text-transform: uppercase;
  letter-spacing: 0.05em;
}

.protection-full {
  background: #dc2626;
  color: #fff;
}

.protection-semi {
  background: #facc15;
  color: #000;
}

.status-verified {
  background: #16a34a;
  color: #fff;
}

.status-pending {
  background: #facc15;
  color: #000;
}

.status-draft {
  background: #6b7280;
  color: #fff;
}

.btn-play-album {
  background: #fff;
  color: #000;
  padding: 1rem 2rem;
  font-weight: 900;
  font-size: 0.875rem;
  text-transform: uppercase;
  letter-spacing: 0.1em;
  border: 4px solid #000;
  cursor: pointer;
  box-shadow: 8px 8px 0px 0px rgba(0,0,0,1);
  transition: all 0.2s;
}
.btn-play-album:hover {
  background: #000;
  color: #fff;
  box-shadow: none;
}

.btn-edit-album {
  display: inline-block;
  text-align: center;
  text-decoration: none;
  padding: 1rem 2rem;
  font-weight: 900;
  font-size: 0.875rem;
  text-transform: uppercase;
  letter-spacing: 0.1em;
  border: 4px solid #000;
  color: #000;
  box-shadow: 8px 8px 0px 0px rgba(0,0,0,1);
  transition: all 0.2s;
}
.btn-edit-album:hover {
  background: #000;
  color: #fff;
  box-shadow: none;
}

/* Track list */
.tracklist-container {
  border: 2px solid #000;
  background: #fff;
}

.tracklist-header {
  border-bottom: 2px solid #000;
  background: #f9fafb;
  padding: 0.75rem 1.5rem;
}

.tracklist-heading {
  font-weight: 900;
  text-transform: uppercase;
  letter-spacing: 0.1em;
  font-size: 0.875rem;
  margin: 0;
}

.tracklist-body { }

.track-row {
  display: flex;
  align-items: center;
  gap: 1.5rem;
  padding: 1rem 1.5rem;
  border-bottom: 2px solid #f3f4f6;
  transition: background 0.2s;
}
.track-row:last-child { border-bottom: none; }
.track-row:hover { background: #f9fafb; }

.track-num {
  font-size: 1.25rem;
  font-weight: 900;
  color: #9ca3af;
  width: 3rem;
  text-align: right;
  flex-shrink: 0;
}

.track-title {
  flex: 1;
}
.track-title h3 {
  font-weight: 700;
  font-size: 1.125rem;
  margin: 0;
}

.track-wiki-links {
  display: flex;
  gap: 0.75rem;
  margin-top: 0.25rem;
  font-size: 0.8rem;
  font-weight: 800;
}
.track-wiki-links a {
  color: #555;
  text-decoration: none;
}
.track-wiki-links a:hover {
  color: #000;
  text-decoration: underline;
}

.btn-track-play {
  border: 2px solid #000;
  padding: 0.5rem 1.25rem;
  font-weight: 900;
  font-size: 0.875rem;
  text-transform: uppercase;
  letter-spacing: 0.1em;
  background: #fff;
  color: #000;
  cursor: pointer;
  transition: all 0.2s;
  flex-shrink: 0;
}
.btn-track-play:hover,
.btn-track-play.btn-track-active {
  background: #000;
  color: #fff;
}

.not-found {
  text-align: center;
  padding: 5rem 0;
  font-size: 1.5rem;
  font-weight: 900;
  color: #9ca3af;
}

.discussion-count-badge {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  background: #3b82f6;
  color: #fff;
  font-size: 0.625rem;
  font-weight: 700;
  min-width: 1.25rem;
  height: 1.25rem;
  border-radius: 9999px;
  margin-left: 0.25rem;
}

@media (max-width: 639px) {
  .page-container { padding: 2rem 1rem; }
  .album-header { flex-direction: column; gap: 1.25rem; }
  .album-cover { width: 100%; height: auto; max-width: 14rem; box-shadow: 8px 8px 0px 0px rgba(0,0,0,1); }
  .album-title { font-size: 1.875rem; }
  .album-actions { flex-direction: column; }
  .btn-play-album, .btn-edit-album { width: 100%; text-align: center; box-sizing: border-box; }
  .track-row { gap: 0.75rem; padding: 0.75rem 1rem; }
  .track-num { width: 2rem; font-size: 1rem; }
  .track-title h3 { font-size: 1rem; }
}

.type-badge {
  font-size: 0.625rem;
  font-weight: 900;
  letter-spacing: 0.1em;
  border: 1px solid #000;
  padding: 0.125rem 0.5rem;
  vertical-align: middle;
  margin-left: 0.5rem;
}
.entry-badge {
  font-size: 0.625rem;
  font-weight: 900;
  letter-spacing: 0.1em;
  padding: 0.125rem 0.5rem;
  border: 1px solid;
  vertical-align: middle;
  margin-left: 0.375rem;
}
.entry-confirmed { border-color: #166534; color: #166534; }
.entry-disputed { border-color: #991b1b; color: #991b1b; }
.admin-entry-actions { display: flex; gap: 0.5rem; margin-top: 0.75rem; flex-wrap: wrap; }
.entry-action-btn {
  border: 2px solid;
  padding: 0.25rem 0.75rem;
  font-size: 0.625rem;
  font-weight: 900;
  text-transform: uppercase;
  letter-spacing: 0.1em;
  cursor: pointer;
  background: transparent;
  transition: all 0.15s;
}
.entry-confirm { border-color: #166534; color: #166534; }
.entry-confirm:hover { background: #166534; color: #fff; }
.entry-dispute { border-color: #991b1b; color: #991b1b; }
.entry-dispute:hover { background: #991b1b; color: #fff; }
.entry-open { border-color: #000; color: #000; }
.entry-open:hover { background: #000; color: #fff; }
</style>
