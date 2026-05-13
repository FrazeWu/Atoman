import { defineStore } from 'pinia';
import { ref, watch } from 'vue';
import type { Song, RepeatMode } from '@/types';

const API_URL = import.meta.env.VITE_API_URL || '/api';

export const usePlayerStore = defineStore('player', () => {
  const songs = ref<Song[]>([]);
  const currentSong = ref<Song | null>(null);
  const isPlaying = ref(false);
  const isShuffled = ref(false);
  const repeatMode = ref<RepeatMode>('none');
  const volume = ref(1);
  const currentTime = ref(0);
  const duration = ref(0);

  // Album-based queue
  const queue = ref<Song[]>([]);
  const currentAlbum = ref<Song[] | null>(null);

  const audio = new Audio();
  const hasRestoredPlayback = ref(false);

  audio.addEventListener('timeupdate', () => {
    currentTime.value = audio.currentTime;
  });
  audio.addEventListener('durationchange', () => {
    duration.value = audio.duration;
  });
  audio.addEventListener('ended', () => {
    playNext();
  });

  const savePlaybackState = () => {
    if (currentSong.value) {
      const state = {
        songId: currentSong.value.id,
        currentTime: audio.currentTime,
        volume: volume.value,
        isPlaying: isPlaying.value,
        isShuffled: isShuffled.value,
        repeatMode: repeatMode.value
      };
      localStorage.setItem('playbackState', JSON.stringify(state));
    }
  };

  const loadPlaybackState = async () => {
    if (hasRestoredPlayback.value || currentSong.value || isPlaying.value) return;

    const savedState = localStorage.getItem('playbackState');
    if (savedState) {
      try {
        const state = JSON.parse(savedState);
        volume.value = state.volume || 1;
        isShuffled.value = state.isShuffled || false;
        repeatMode.value = state.repeatMode || 'none';

        if (state.songId) {
          const song = songs.value.find(s => s.id === state.songId);
          if (song) {
            audio.src = song.audio_url;
            audio.volume = volume.value;
            audio.currentTime = state.currentTime || 0;
            currentSong.value = song;
            isPlaying.value = false;
          }
        }
      } catch (error) {
        console.error('Failed to restore playback state:', error);
      }
    }

    hasRestoredPlayback.value = true;
  };

  watch([currentSong, currentTime, volume, isPlaying, isShuffled, repeatMode], () => {
    savePlaybackState();
  }, { deep: true });

  const fetchSongs = async () => {
    try {
      const response = await fetch(`${API_URL}/songs`);
      if (response.ok) {
        songs.value = await response.json();
        await loadPlaybackState();
      }
    } catch (error) {
      console.error('Failed to fetch songs:', error);
    }
  };

  // Play a single song (uses queue if set, else falls back to all songs)
  const playSong = (song: Song) => {
    if (currentSong.value?.id === song.id) {
      togglePlay();
    } else {
      audio.src = song.audio_url;
      audio.volume = volume.value;
      audio.play();
      currentSong.value = song;
      isPlaying.value = true;
    }
  };

  // Play an album starting from a specific index
  const playAlbum = (albumSongs: Song[], startIndex = 0) => {
    if (albumSongs.length === 0) return;
    currentAlbum.value = albumSongs;
    queue.value = [...albumSongs];
    const song = albumSongs[startIndex];
    audio.src = song.audio_url;
    audio.volume = volume.value;
    audio.play();
    currentSong.value = song;
    isPlaying.value = true;
  };

  const togglePlay = () => {
    if (currentSong.value) {
      if (isPlaying.value) {
        audio.pause();
      } else {
        audio.play();
      }
      isPlaying.value = !isPlaying.value;
    }
  };

  const _getActiveList = () => queue.value.length > 0 ? queue.value : songs.value;

  const playNext = () => {
    const list = _getActiveList();
    if (!currentSong.value || list.length === 0) return;
    const currentIndex = list.findIndex(s => s.id === currentSong.value?.id);

    let nextIndex;
    if (isShuffled.value) {
      nextIndex = Math.floor(Math.random() * list.length);
    } else {
      if (repeatMode.value === 'one') {
        audio.currentTime = 0;
        audio.play();
        isPlaying.value = true;
        return;
      } else if (repeatMode.value === 'all' || currentIndex < list.length - 1) {
        nextIndex = (currentIndex + 1) % list.length;
      } else {
        isPlaying.value = false;
        return;
      }
    }
    playSong(list[nextIndex]);
  };

  const playPrevious = () => {
    const list = _getActiveList();
    if (!currentSong.value || list.length === 0) return;
    const currentIndex = list.findIndex(s => s.id === currentSong.value?.id);
    let prevIndex = (currentIndex - 1 + list.length) % list.length;
    playSong(list[prevIndex]);
  };

  const toggleShuffle = () => {
    isShuffled.value = !isShuffled.value;
  };

  const toggleRepeat = () => {
    const modes: RepeatMode[] = ['none', 'all', 'one'];
    const nextMode = modes[(modes.indexOf(repeatMode.value) + 1) % modes.length];
    repeatMode.value = nextMode;
  };

  const setVolume = (v: number) => {
    audio.volume = v;
    volume.value = v;
  };

  const seek = (time: number) => {
    audio.currentTime = time;
    currentTime.value = time;
  };

  return {
    songs,
    currentSong,
    isPlaying,
    isShuffled,
    repeatMode,
    volume,
    currentTime,
    duration,
    queue,
    currentAlbum,
    fetchSongs,
    playSong,
    playAlbum,
    togglePlay,
    playNext,
    playPrevious,
    toggleShuffle,
    toggleRepeat,
    setVolume,
    seek
  };
});
