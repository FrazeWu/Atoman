<script setup lang="ts">
import { computed } from 'vue';
import { useRoute, RouterLink } from 'vue-router';
import { usePlayerStore } from '@/stores/player';
import { useAuthStore } from '@/stores/auth';

const route = useRoute();
const player = usePlayerStore();
const authStore = useAuthStore();

const singerName = decodeURIComponent(route.params.artist as string).replace(/_/g, ' ');
const albumName = decodeURIComponent(route.params.album as string).replace(/_/g, ' ');

const albumSongs = computed(() => {
  return player.songs.filter(song => song.album === albumName && song.artist === singerName);
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
    status: firstSong.status
  };
});

// Since the correction modal is moved to EditAlbumView, this component
// no longer needs the correction-related functions or refs.

// Initial song fetching, ensuring playerStore has song data for computed properties
player.fetchSongs();
</script>

<template>
  <div class="page-container">
    <RouterLink to="/" class="back-link">← 返回时间线</RouterLink>

    <div v-if="albumInfo" class="album-content">
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
            <span v-if="albumInfo.status === 'pending'" class="pending-badge">待审核</span>
          </h1>
          <p class="album-artist">{{ albumInfo.artist }}</p>
          <p class="album-tracks">{{ albumInfo.trackCount }} {{ albumInfo.trackCount === 1 ? 'track' : 'tracks' }}</p>

          <div class="album-actions">
            <button @click="player.playSong(albumSongs[0])" class="btn-play-album">
              ▶ 播放专辑
            </button>
            <RouterLink
              v-if="authStore.isAuthenticated"
              :to="`/artist=${singerName.replace(/ /g, '_')}/album=${albumName.replace(/ /g, '_')}/edit`"
              class="btn-edit-album"
            >
              编辑专辑
            </RouterLink>
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
  background: #facc15;
  color: #713f12;
  font-size: 0.875rem;
  font-weight: 700;
  padding: 0.25rem 0.75rem;
  border-radius: 9999px;
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
</style>
