<template>
  <div v-if="player.currentSong" class="player">
    <div class="player-inner">
      <!-- Song info -->
      <div class="player-info">
        <img
          v-if="player.currentSong.cover_url"
          :src="player.currentSong.cover_url"
          :alt="player.currentSong.title"
          class="player-cover"
        />
        <div class="player-meta">
          <p class="player-title">{{ player.currentSong.title }}</p>
          <p class="player-artist">{{ artistText }}</p>
        </div>
      </div>

      <!-- Controls -->
      <div class="player-controls">
        <!-- Shuffle -->
        <button class="ctrl-btn" :class="{ active: player.isShuffled }" @click="player.toggleShuffle()" title="随机">
          随机
        </button>
        <!-- Prev -->
        <button class="ctrl-btn" @click="player.playPrevious()" title="上一首">上一首</button>
        <!-- Play/Pause -->
        <button class="ctrl-btn play-btn" @click="player.togglePlay()">
          {{ player.isPlaying ? '暂停' : '播放' }}
        </button>
        <!-- Next -->
        <button class="ctrl-btn" @click="player.playNext()" title="下一首">下一首</button>
        <!-- Repeat -->
        <button
          class="ctrl-btn"
          :class="{ active: player.repeatMode !== 'none' }"
          @click="player.toggleRepeat()"
          title="循环"
        >
          {{ repeatText }}
        </button>
      </div>

      <!-- Progress + Volume -->
      <div class="player-right">
        <!-- Time + progress -->
        <div class="player-progress-wrap">
          <span class="player-time">{{ formatTime(player.currentTime) }}</span>
          <div class="progress-bar" @click="seek">
            <div class="progress-fill" :style="{ width: progressPct + '%' }" />
          </div>
          <span class="player-time">{{ formatTime(player.duration) }}</span>
        </div>
        <!-- Volume -->
        <div class="volume-wrap">
          <span class="vol-label">音量</span>
          <input
            type="range"
            min="0"
            max="1"
            step="0.01"
            :value="player.volume"
            @input="(e) => player.setVolume(parseFloat((e.target as HTMLInputElement).value))"
            class="vol-slider"
          />
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import { usePlayerStore } from '@/stores/player'

const player = usePlayerStore()

const progressPct = computed(() => {
  if (!player.duration) return 0
  return (player.currentTime / player.duration) * 100
})

const artistText = computed(() => {
  if (!player.currentSong) return '未知艺术家'
  if (player.currentSong.artists?.length) {
    return player.currentSong.artists.map((artist) => artist.name).join(', ')
  }
  return player.currentSong.artist || '未知艺术家'
})

const repeatText = computed(() => {
  if (player.repeatMode === 'one') return '单曲循环'
  if (player.repeatMode === 'all') return '列表循环'
  return '不循环'
})

const formatTime = (s: number) => {
  if (!s || isNaN(s)) return '0:00'
  const m = Math.floor(s / 60)
  const sec = Math.floor(s % 60)
  return `${m}:${sec.toString().padStart(2, '0')}`
}

const seek = (e: MouseEvent) => {
  const bar = e.currentTarget as HTMLElement
  const pct = e.offsetX / bar.offsetWidth
  player.seek(pct * player.duration)
}
</script>

<style scoped>
.player {
  position: fixed;
  bottom: 0;
  width: 100%;
  z-index: 50;
  background: #fff;
  border-top: 2px solid #000;
}
.player-inner {
  max-width: 1152px;
  margin: 0 auto;
  padding: 0 2rem;
  height: 72px;
  display: grid;
  grid-template-columns: 1fr auto 1fr;
  align-items: center;
  gap: 1.5rem;
}
.player-info {
  display: flex;
  align-items: center;
  gap: 0.75rem;
  min-width: 0;
}
.player-cover {
  width: 44px;
  height: 44px;
  border: 2px solid #000;
  object-fit: cover;
  filter: grayscale(1);
  flex-shrink: 0;
}
.player-meta { min-width: 0; }
.player-title {
  font-size: 0.875rem;
  font-weight: 900;
  margin: 0;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}
.player-artist {
  font-size: 0.75rem;
  color: #6b7280;
  font-weight: 500;
  margin: 0;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}
.player-controls {
  display: flex;
  align-items: center;
  gap: 0.5rem;
}
.ctrl-btn {
  background: none;
  border: 1px solid #000;
  cursor: pointer;
  padding: 0.35rem 0.6rem;
  font-size: 0.75rem;
  font-weight: 900;
  transition: all 0.15s;
  line-height: 1.5;
  white-space: nowrap;
}
.ctrl-btn:hover { background: #000; color: #fff; }
.ctrl-btn.active { background: #000; color: #fff; }
.play-btn {
  padding: 0.35rem 1rem;
  background: #000;
  color: #fff;
  border: 2px solid #000;
  font-size: 0.8rem;
}
.play-btn:hover { background: #fff; color: #000; }
.player-right {
  display: flex;
  flex-direction: column;
  justify-content: center;
  gap: 0.375rem;
}
.player-progress-wrap {
  display: flex;
  align-items: center;
  gap: 0.5rem;
}
.player-time { font-size: 0.7rem; color: #9ca3af; font-weight: 700; flex-shrink: 0; }
.progress-bar {
  flex: 1;
  height: 4px;
  background: #e5e7eb;
  cursor: pointer;
  position: relative;
}
.progress-fill {
  height: 100%;
  background: #000;
  transition: width 0.1s linear;
}
.volume-wrap {
  display: flex;
  align-items: center;
  gap: 0.375rem;
  justify-content: flex-end;
}
.vol-label {
  font-size: 0.7rem;
  color: #6b7280;
  font-weight: 800;
}
.vol-slider {
  width: 80px;
  accent-color: #000;
  cursor: pointer;
}
</style>
